package consensus

import (
	"context"
	"slices"
	"sync/atomic"
	"time"

	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	v2stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/raft/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var durationBuckets = []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// metrics manages OpenTelemetry instruments and status observations for a Raft group node.
type metrics struct {
	baseAttrs        []attribute.KeyValue
	registration     metric.Registration
	proposalDuration metric.Float64Histogram
	snapshotDuration metric.Float64Histogram
}

// newMetrics exposes the Raft status of this node as OpenTelemetry instruments.
// Peer metrics are only reported by the leader.
func newMetrics(
	group string,
	nodeID uint64,
	nodeState *node.NodeState,
	storage *storage,
	snapshotIndex *atomic.Uint64,
	tracker *proposalsTracker,
	readTracker *readIndexTracker,
	leaderStats *v2stats.LeaderStats,
) (*metrics, error) {
	meter := otel.Meter("raftstore")

	proposalDuration, err := meter.Float64Histogram("raft_proposal_duration_seconds", metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create histogram")
	}
	snapshotDuration, err := meter.Float64Histogram("raft_snapshot_duration_seconds", metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create histogram")
	}

	gauges := map[string]metric.Int64ObservableGauge{}
	for _, name := range []string{
		"raft_term", "raft_commit_index", "raft_applied_index", "raft_leader_id", "raft_is_leader", "raft_state",
		"raft_snapshot_index", "raft_log_entries", "raft_proposals_pending", "raft_follower_lag_entries",
	} {
		gauges[name], err = meter.Int64ObservableGauge(name)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create gauge %s", name)
		}
	}
	latencies := map[string]metric.Float64ObservableGauge{}
	for _, name := range []string{
		"raft_peer_latency_seconds", "raft_peer_latency_average_seconds", "raft_peer_latency_stddev_seconds",
		"raft_peer_latency_min_seconds", "raft_peer_latency_max_seconds",
	} {
		latencies[name], err = meter.Float64ObservableGauge(name)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create gauge %s", name)
		}
	}
	sendSuccess, err := meter.Int64ObservableCounter("raft_peer_send_success_total")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create counter")
	}
	sendFail, err := meter.Int64ObservableCounter("raft_peer_send_fail_total")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create counter")
	}

	instruments := []metric.Observable{sendSuccess, sendFail}
	for _, g := range gauges {
		instruments = append(instruments, g)
	}
	for _, g := range latencies {
		instruments = append(instruments, g)
	}

	m := &metrics{
		baseAttrs: []attribute.KeyValue{
			attribute.String("group", group),
			attribute.String("node_id", types.ID(nodeID).String()),
		},
		proposalDuration: proposalDuration,
		snapshotDuration: snapshotDuration,
	}

	m.registration, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		attrs := metric.WithAttributes(m.baseAttrs...)

		o.ObserveInt64(gauges["raft_snapshot_index"], int64(snapshotIndex.Load()), attrs)

		firstIndex, err := storage.FirstIndex()
		if err != nil {
			return stacktrace.Propagate(err, "failed to get first index")
		}
		lastIndex, err := storage.LastIndex()
		if err != nil {
			return stacktrace.Propagate(err, "failed to get last index")
		}
		o.ObserveInt64(gauges["raft_log_entries"], int64(lastIndex+1-firstIndex), attrs)

		o.ObserveInt64(gauges["raft_proposals_pending"], int64(tracker.pendingCount()), metric.WithAttributes(m.attributes(attribute.String("type", "write"))...))
		o.ObserveInt64(gauges["raft_proposals_pending"], int64(readTracker.pendingCount()), metric.WithAttributes(m.attributes(attribute.String("type", "read"))...))

		raftNode := nodeState.RaftNode()
		if raftNode == nil {
			// A node found removed at startup has no raft node to report the status of.
			return nil
		}
		status := raftNode.Status()

		o.ObserveInt64(gauges["raft_term"], int64(status.GetTerm()), attrs)
		o.ObserveInt64(gauges["raft_commit_index"], int64(status.GetCommit()), attrs)
		o.ObserveInt64(gauges["raft_applied_index"], int64(status.Applied), attrs)
		o.ObserveInt64(gauges["raft_leader_id"], 1, metric.WithAttributes(m.attributes(attribute.String("leader_id", types.ID(status.Lead).String()))...))

		isLeader := int64(0)
		if status.RaftState == raft.StateLeader {
			isLeader = 1
		}
		o.ObserveInt64(gauges["raft_is_leader"], isLeader, attrs)
		o.ObserveInt64(gauges["raft_state"], 1, metric.WithAttributes(m.attributes(attribute.String("state", status.RaftState.String()))...))

		if status.RaftState != raft.StateLeader {
			return nil
		}

		for id, pr := range status.Progress {
			if id == nodeID {
				continue
			}
			lag := uint64(0)
			if status.GetCommit() > pr.Match {
				lag = status.GetCommit() - pr.Match
			}
			o.ObserveInt64(gauges["raft_follower_lag_entries"], int64(lag), metric.WithAttributes(m.attributes(attribute.String("peer_id", types.ID(id).String()))...))
		}

		type peerStats struct {
			latency v2stats.LatencyStats
			counts  v2stats.CountsStats
		}
		leaderStats.Lock()
		followers := make(map[string]peerStats, len(leaderStats.Followers))
		for peerID, fs := range leaderStats.Followers {
			fs.Lock()
			followers[peerID] = peerStats{latency: fs.Latency, counts: fs.Counts}
			fs.Unlock()
		}
		leaderStats.Unlock()

		const msToSeconds = 1.0 / 1000.0
		for peerID, fs := range followers {
			peerAttrs := metric.WithAttributes(m.attributes(attribute.String("peer_id", peerID))...)
			o.ObserveFloat64(latencies["raft_peer_latency_seconds"], fs.latency.Current*msToSeconds, peerAttrs)
			o.ObserveFloat64(latencies["raft_peer_latency_average_seconds"], fs.latency.Average*msToSeconds, peerAttrs)
			o.ObserveFloat64(latencies["raft_peer_latency_stddev_seconds"], fs.latency.StandardDeviation*msToSeconds, peerAttrs)
			o.ObserveFloat64(latencies["raft_peer_latency_min_seconds"], fs.latency.Minimum*msToSeconds, peerAttrs)
			o.ObserveFloat64(latencies["raft_peer_latency_max_seconds"], fs.latency.Maximum*msToSeconds, peerAttrs)
			o.ObserveInt64(sendSuccess, int64(fs.counts.Success), peerAttrs)
			o.ObserveInt64(sendFail, int64(fs.counts.Fail), peerAttrs)
		}
		return nil
	}, instruments...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to register callback")
	}

	return m, nil
}

func (m *metrics) attributes(extra ...attribute.KeyValue) []attribute.KeyValue {
	return slices.Concat(m.baseAttrs, extra)
}

func (m *metrics) recordProposal(ctx context.Context, start time.Time, proposalType string) {
	m.proposalDuration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(m.attributes(attribute.String("type", proposalType))...))
}

func (m *metrics) recordSnapshot(ctx context.Context, start time.Time) {
	m.snapshotDuration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(m.baseAttrs...))
}

func (m *metrics) unregister() error {
	return m.registration.Unregister()
}
