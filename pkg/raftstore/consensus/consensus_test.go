package consensus

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/pb/auxpb"
	auxsnapshotpb "github.com/interuss/dss/pkg/pb/auxpb/snapshotpb"
	proposalpb "github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
)

// testNode bundles a Consensus instance with an in-memory key-value state machine and its dedicated Raft listener.
type testNode struct {
	consensus *Consensus
	url       string
	dataDir   string

	mu           sync.RWMutex
	kv           map[string]string
	stop         context.CancelFunc
	done         chan struct{}
	shutdownOnce sync.Once
	shutdownErr  error
}

func startTestListener(t *testing.T) (*listener, string) {
	t.Helper()

	l, addr, err := startServerListener(t.Context(), zap.NewNop(), "127.0.0.1:0", etcdtransport.TLSInfo{}, func(err error) {
		t.Errorf("unexpected listener error: %v", err)
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = l.server.Shutdown(shutdownCtx)
	})

	return l, fmt.Sprintf("http://%s", addr)
}

func startTestNode(t *testing.T, dataDir string, snapshotInterval uint64) *testNode {
	t.Helper()

	l, peerURL := startTestListener(t)
	return openTestNodeOnListener(t, l, peerURL, dataDir, snapshotInterval)
}

func openTestNodeOnListener(t *testing.T, l *listener, peerURL string, dataDir string, snapshotInterval uint64) *testNode {
	t.Helper()

	// Use context.Background() rather than t.Context(): t.Context() is canceled before t.Cleanup runs,
	// whereas consumeCommits must keep draining unbuffered commitC until consensus.Stop completes.
	appCtx, cancel := context.WithCancel(context.Background())
	commitC := make(chan EntryCommit)
	n := &testNode{
		url:     peerURL,
		dataDir: dataDir,
		kv:      map[string]string{},
		stop:    cancel,
		done:    make(chan struct{}),
	}

	go func() {
		defer close(n.done)
		n.consumeCommits(appCtx, commitC)
	}()

	connParams := params.ConnectParameters{
		URL:                     peerURL,
		DataDir:                 dataDir,
		ClusterID:               1,
		Group:                   "aux",
		GroupClusterID:          42,
		SnapshotCatchupEntries:  2,
		SnapshotIntervalEntries: snapshotInterval,
		TickInterval:            10 * time.Millisecond,
		ElectionTick:            10,
		HeartbeatTick:           1,
		MaxSizePerMsg:           1024 * 1024,
		MaxInflightMsgs:         64,
		Insecure:                true,
	}

	c, err := newConsensus(t.Context(), zap.NewNop(), connParams, n.snapshot, commitC, l)
	if err != nil {
		cancel()
		<-n.done
		require.NoError(t, err)
	}
	n.consensus = c
	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = n.shutdown(shutdownCtx)
	})
	return n
}

func (n *testNode) shutdown(ctx context.Context) error {
	n.shutdownOnce.Do(func() {
		n.shutdownErr = n.consensus.Stop(ctx)
		n.stop()
		<-n.done
	})
	return n.shutdownErr
}

func (n *testNode) snapshot() (*proposalpb.Snapshot, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	participants := make([]*auxsnapshotpb.Participant, 0, len(n.kv))
	for k, v := range n.kv {
		participants = append(participants, &auxsnapshotpb.Participant{
			Locality:       k,
			PublicEndpoint: v,
		})
	}
	return &proposalpb.Snapshot{
		AppSnapshot: &proposalpb.Snapshot_AuxSnapshot{
			AuxSnapshot: &auxsnapshotpb.Snapshot{Participants: participants},
		},
	}, nil
}

func (n *testNode) consumeCommits(ctx context.Context, commitC <-chan EntryCommit) {
	for {
		select {
		case <-ctx.Done():
			return
		case commit := <-commitC:
			if commit.Snapshot != nil {
				snap := commit.Snapshot.GetAuxSnapshot()
				restored := make(map[string]string, len(snap.GetParticipants()))
				for _, p := range snap.GetParticipants() {
					restored[p.GetLocality()] = p.GetPublicEndpoint()
				}
				n.mu.Lock()
				n.kv = restored
				n.mu.Unlock()
				continue
			}

			switch req := commit.Prop.Payload.GetAuxPayload().GetPayload().(type) {
			case *auxpb.Payload_SaveOwnMetadata:
				n.mu.Lock()
				n.kv[req.SaveOwnMetadata.GetLocality()] = req.SaveOwnMetadata.GetPublicEndpoint()
				n.mu.Unlock()
				commit.Done <- ProposalResult{}
			case *auxpb.Payload_GetDssMetadata:
				n.mu.RLock()
				snapshot := maps.Clone(n.kv)
				n.mu.RUnlock()
				commit.Done <- ProposalResult{Result: snapshot}
			default:
				commit.Done <- ProposalResult{Error: fmt.Errorf("unexpected payload %T", req)}
			}
		}
	}
}

func (n *testNode) put(t *testing.T, key, val string) {
	t.Helper()
	ctx := timestamp.NewContext(t.Context(), time.Now())
	payload := &proposalpb.Payload{
		Payload: &proposalpb.Payload_AuxPayload{
			AuxPayload: &auxpb.Payload{
				Payload: &auxpb.Payload_SaveOwnMetadata{
					SaveOwnMetadata: &auxpb.SaveOwnMetadata{
						Locality:       key,
						PublicEndpoint: val,
					},
				},
			},
		},
	}
	_, err := n.consensus.HandleWritePayload[any](ctx, payload)
	require.NoError(t, err)
}

func (n *testNode) readAll(t *testing.T) map[string]string {
	t.Helper()
	ctx := timestamp.NewContext(t.Context(), time.Now())
	payload := &proposalpb.Payload{
		Payload: &proposalpb.Payload_AuxPayload{
			AuxPayload: &auxpb.Payload{
				Payload: &auxpb.Payload_GetDssMetadata{
					GetDssMetadata: &auxpb.GetDSSMetadata{},
				},
			},
		},
	}
	got, err := n.consensus.HandleReadPayload[map[string]string](ctx, payload)
	require.NoError(t, err)
	return got
}

func TestConsensusSingleNodeLifecycleAndRestart(t *testing.T) {
	dataDir := t.TempDir()
	l, peerURL := startTestListener(t)

	n1 := openTestNodeOnListener(t, l, peerURL, dataDir, 3)

	// Before bootstrap: node is waiting and rejects membership/data requests.
	info, err := n1.consensus.Node()
	require.NoError(t, err)
	assert.Equal(t, node.Waiting, info.State)
	assert.False(t, n1.consensus.Ready())

	_, err = n1.consensus.Members()
	assert.Equal(t, dsserr.NotFound, stacktrace.GetCode(err))

	var leaderRuns atomic.Int32
	leaderStopped := make(chan struct{})
	n1.consensus.RunAsLeader(t.Context(), func(ctx context.Context) {
		leaderRuns.Add(1)
		<-ctx.Done()
		close(leaderStopped)
	})

	membership, err := n1.consensus.Bootstrap(t.Context())
	require.NoError(t, err)
	assert.Equal(t, info.ID, membership.LeaderID)
	require.Eventually(t, n1.consensus.Ready, 5*time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool { return leaderRuns.Load() == 1 }, 5*time.Second, 5*time.Millisecond)

	// Registering RunAsLeader while already leader starts the callback immediately.
	postElectionStopped := make(chan struct{})
	var postElectionRuns atomic.Int32
	n1.consensus.RunAsLeader(t.Context(), func(ctx context.Context) {
		postElectionRuns.Add(1)
		<-ctx.Done()
		close(postElectionStopped)
	})
	require.Eventually(t, func() bool { return postElectionRuns.Load() == 1 }, 5*time.Second, 5*time.Millisecond)

	// Bootstrapping a second time is rejected.
	_, err = n1.consensus.Bootstrap(t.Context())
	assert.Equal(t, dsserr.AlreadyExists, stacktrace.GetCode(err))

	// Write enough entries to cross SnapshotIntervalEntries (3) and verify linearizable read.
	for i := range 5 {
		n1.put(t, fmt.Sprintf("uss-%d", i), fmt.Sprintf("https://uss-%d.example", i))
	}
	got := n1.readAll(t)
	assert.Len(t, got, 5)
	assert.Equal(t, "https://uss-4.example", got["uss-4"])

	// Cleanly shut down and restart on the same DataDir; leader routines must stop and state + identity must recover from snapshot + WAL.
	require.NoError(t, n1.shutdown(t.Context()))
	<-leaderStopped
	<-postElectionStopped

	n2 := openTestNodeOnListener(t, l, peerURL, dataDir, 100)

	restartedInfo, err := n2.consensus.Node()
	require.NoError(t, err)
	assert.Equal(t, info.ID, restartedInfo.ID)
	assert.Equal(t, node.Member, restartedInfo.State)

	require.Eventually(t, n2.consensus.Ready, 5*time.Second, 5*time.Millisecond)
	assert.Equal(t, got, n2.readAll(t))
}

func TestConsensusClusterMembershipReplicationAndSelfRemoval(t *testing.T) {
	n1 := startTestNode(t, t.TempDir(), 4)
	n2 := startTestNode(t, t.TempDir(), 4)
	n3 := startTestNode(t, t.TempDir(), 4)

	n1Info, err := n1.consensus.Node()
	require.NoError(t, err)

	_, err = n1.consensus.Bootstrap(t.Context())
	require.NoError(t, err)
	require.Eventually(t, n1.consensus.Ready, 5*time.Second, 5*time.Millisecond)

	added2, err := n1.consensus.AddNode(t.Context(), n2.url)
	require.NoError(t, err)
	added3, err := n1.consensus.AddNode(t.Context(), n3.url)
	require.NoError(t, err)

	require.Eventually(t, n2.consensus.Ready, 5*time.Second, 5*time.Millisecond)
	require.Eventually(t, n3.consensus.Ready, 5*time.Second, 5*time.Millisecond)

	// Write enough entries through leader and follower to trigger snapshots and compact the log.
	for i := range 6 {
		target := n1
		if i%2 == 1 {
			target = n2
		}
		target.put(t, fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}

	// Read via ReadIndex from a follower.
	gotFromFollower := n3.readAll(t)
	assert.Len(t, gotFromFollower, 6)
	assert.Equal(t, "v5", gotFromFollower["k5"])

	// Add a 4th node after compaction so it catches up via snapshot + log replication.
	n4 := startTestNode(t, t.TempDir(), 4)

	added4, err := n1.consensus.AddNode(t.Context(), n4.url)
	require.NoError(t, err)
	require.Eventually(t, n4.consensus.Ready, 5*time.Second, 5*time.Millisecond)
	assert.Equal(t, gotFromFollower, n4.readAll(t))

	// Remove n4 from the 4-node cluster by proposing self-removal directly on n4.
	n4Info, err := n4.consensus.Node()
	require.NoError(t, err)
	assert.Equal(t, added4.ID, n4Info.ID)

	require.NoError(t, n4.consensus.RemoveNode(t.Context(), added4.ID))

	require.Eventually(t, func() bool {
		info, err := n4.consensus.Node()
		return err == nil && info.State == node.Removed && !n4.consensus.Ready()
	}, 5*time.Second, 5*time.Millisecond)

	// Remaining 3-voter group reflects n4 as Removed and continues serving writes.
	require.Eventually(t, func() bool {
		members, err := n1.consensus.Members()
		return err == nil &&
			members.Nodes[added2.ID].State == node.Member &&
			members.Nodes[added3.ID].State == node.Member &&
			members.Nodes[added4.ID].State == node.Removed
	}, 5*time.Second, 5*time.Millisecond)

	n2.put(t, "after-removal", "ok")
	assert.Equal(t, "ok", n3.readAll(t)["after-removal"])

	// Shutting down n4 after self-removal succeeds cleanly, archives its WAL, and allows starting fresh on the same DataDir.
	require.NoError(t, n4.shutdown(t.Context()))
	assert.DirExists(t, filepath.Join(n4.dataDir, "archive", strconv.FormatUint(added4.ID, 10), "wal"))

	n4Reopened := startTestNode(t, n4.dataDir, 4)
	reopenedInfo, err := n4Reopened.consensus.Node()
	require.NoError(t, err)
	assert.NotEqual(t, added4.ID, reopenedInfo.ID)
	assert.Equal(t, node.Waiting, reopenedInfo.State)
	assert.False(t, n4Reopened.consensus.Ready())

	// Leader failover: stopping leader n1 cancels its leader routine, elects n2 or n3, starts the new leader's routine, and preserves quorum.
	n1LeaderStopped := make(chan struct{})
	n1.consensus.RunAsLeader(t.Context(), func(ctx context.Context) {
		<-ctx.Done()
		close(n1LeaderStopped)
	})
	var followerLeaderRuns atomic.Int32
	for _, follower := range []*testNode{n2, n3} {
		follower.consensus.RunAsLeader(t.Context(), func(ctx context.Context) {
			followerLeaderRuns.Add(1)
			<-ctx.Done()
		})
	}

	require.NoError(t, n1.shutdown(t.Context()))
	<-n1LeaderStopped

	require.Eventually(t, func() bool {
		if !n2.consensus.Ready() || !n3.consensus.Ready() {
			return false
		}
		members, err := n2.consensus.Members()
		return err == nil && (members.LeaderID == added2.ID || members.LeaderID == added3.ID) && members.LeaderID != n1Info.ID
	}, 5*time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool { return followerLeaderRuns.Load() >= 1 }, 5*time.Second, 5*time.Millisecond)

	n2.put(t, "after-failover", "ok")
	assert.Equal(t, "ok", n3.readAll(t)["after-failover"])
}

func TestConsensusMetricsCollection(t *testing.T) {
	prevProvider := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	t.Cleanup(func() {
		otel.SetMeterProvider(prevProvider)
		_ = provider.Shutdown(context.Background())
	})

	collectNames := func(t *testing.T) []string {
		t.Helper()
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(t.Context(), &rm))
		var names []string
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				names = append(names, m.Name)
			}
		}
		return names
	}

	n1 := startTestNode(t, t.TempDir(), 3)
	n2 := startTestNode(t, t.TempDir(), 3)
	n3 := startTestNode(t, t.TempDir(), 3)

	// Collecting while waiting (RaftNode() == nil) must succeed and report storage/pending gauges.
	preBootstrap := collectNames(t)
	assert.Contains(t, preBootstrap, "raft_snapshot_index")
	assert.Contains(t, preBootstrap, "raft_log_entries")
	assert.Contains(t, preBootstrap, "raft_proposals_pending")

	_, err := n1.consensus.Bootstrap(t.Context())
	require.NoError(t, err)
	require.Eventually(t, n1.consensus.Ready, 5*time.Second, 5*time.Millisecond)

	_, err = n1.consensus.AddNode(t.Context(), n2.url)
	require.NoError(t, err)
	_, err = n1.consensus.AddNode(t.Context(), n3.url)
	require.NoError(t, err)
	require.Eventually(t, n2.consensus.Ready, 5*time.Second, 5*time.Millisecond)
	require.Eventually(t, n3.consensus.Ready, 5*time.Second, 5*time.Millisecond)

	// Stop one follower and write enough entries to trigger a snapshot and non-zero follower lag.
	require.NoError(t, n3.shutdown(t.Context()))
	for i := range 3 {
		n1.put(t, fmt.Sprintf("mk-%d", i), "mv")
	}
	_ = n2.readAll(t)

	got := collectNames(t)
	for _, want := range []string{
		"raft_proposal_duration_seconds",
		"raft_snapshot_duration_seconds",
		"raft_term",
		"raft_commit_index",
		"raft_applied_index",
		"raft_leader_id",
		"raft_is_leader",
		"raft_state",
		"raft_snapshot_index",
		"raft_log_entries",
		"raft_proposals_pending",
		"raft_follower_lag_entries",
		"raft_peer_latency_seconds",
		"raft_peer_latency_average_seconds",
		"raft_peer_latency_stddev_seconds",
		"raft_peer_latency_min_seconds",
		"raft_peer_latency_max_seconds",
		"raft_peer_send_success_total",
		"raft_peer_send_fail_total",
	} {
		assert.Contains(t, got, want)
	}
}

func TestConsensusTransportErrorAndWALRecovery(t *testing.T) {
	dataDir := t.TempDir()
	l, peerURL := startTestListener(t)

	n1 := openTestNodeOnListener(t, l, peerURL, dataDir, 100)
	_, err := n1.consensus.Bootstrap(t.Context())
	require.NoError(t, err)
	require.Eventually(t, n1.consensus.Ready, 5*time.Second, 5*time.Millisecond)

	n1.put(t, "before-crash", "present")

	// Simulate a fatal transport error stopping the Ready consumer.
	n1.consensus.transport.ErrorC <- errors.New("simulated transport failure")
	require.Eventually(t, func() bool { return !n1.consensus.Ready() }, 5*time.Second, 5*time.Millisecond)

	// Stop must succeed cleanly after a transport error, releasing listener and WAL locks.
	require.NoError(t, n1.shutdown(t.Context()))

	// Restarting on the same DataDir recovers committed entries from WAL.
	n2 := openTestNodeOnListener(t, l, peerURL, dataDir, 100)

	require.Eventually(t, n2.consensus.Ready, 5*time.Second, 5*time.Millisecond)
	assert.Equal(t, map[string]string{"before-crash": "present"}, n2.readAll(t))
}
