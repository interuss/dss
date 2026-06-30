package params

import (
	"flag"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/interuss/stacktrace"
	"go.etcd.io/raft/v3"
)

const (
	defaultDataDir = "raft_data"

	// the default Raft related parameters are the same as the default values used by etcd for the moment.
	// TODO - review and adjust these parameters as needed based on testing and performance tuning.

	defaultSnapshotCatchupEntries  = 5000
	defaultSnapshotIntervalEntries = 10000
	defaultTickInterval            = 100 * time.Millisecond

	// follower waits 10 x defaultTickInterval without a heartbeat before starting an election
	defaultElectionTick = 10
	// leader sends a heartbeat every tick, must be < defaultElectionTick
	defaultHeartbeatTick = 1

	defaultMaxSizePerMsg   = 1024 * 1024
	defaultMaxInflightMsgs = 4096 / 8
)

type (
	// ConnectParameters bundles up parameters used for connecting nodes in a raftstore cluster.
	ConnectParameters struct {
		// unique node identifier within the cluster, 0 is invalid
		NodeID uint64
		// comma-separated "nodeID=peerURL" pairs defining all cluster members including this node
		Peers string

		// DataDir is the directory where the node persists its Raft state (WAL segments and snapshots).
		// This data is required for the node to restart and rejoin the cluster without having to receive
		// a full snapshot from the leader. It must not be deleted while the node is running or
		// across restarts unless the node is being permanently shut down.
		// If the directory is lost, the node will recover by receiving a snapshot from the leader.
		DataDir string
		// discriminates this cluster from others sharing the same network, must be identical on all nodes
		ClusterID uint64

		// number of entries for a slow follower to catch-up after compacting.
		// This gives the follower a buffer of entries while avoiding the need to send a full snapshot.
		SnapshotCatchupEntries uint64
		// number of entries applied before triggering a snapshot
		SnapshotIntervalEntries uint64
		// base time unit for Raft's logical clock, scales both election and heartbeat timers
		TickInterval time.Duration

		// ticks without a heartbeat before a follower promotes to candidate, effective timeout = ElectionTick × TickInterval
		ElectionTick int
		// ticks between leader heartbeats, must be < ElectionTick
		HeartbeatTick int

		// max byte size of a message sent to a peer
		MaxSizePerMsg uint64
		// max number of in-flight messages during optimistic replication phase
		MaxInflightMsgs int
	}
)

// PeerMap parses the Peers string into a map of node ID to peer URL.
func (c ConnectParameters) PeerMap() (map[uint64]*url.URL, error) {
	peers := make(map[uint64]*url.URL)

	for entry := range strings.SplitSeq(c.Peers, ",") {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, stacktrace.NewError("invalid peer entry %s: must be in format nodeID=peerURL", entry)
		}

		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid peer ID %s", parts[0])
		}

		if id == 0 {
			return nil, stacktrace.NewError("invalid peer ID 0: peer IDs must be greater than 0")
		}

		if _, exists := peers[id]; exists {
			return nil, stacktrace.NewError("duplicate peer ID %d", id)
		}

		peerURL, err := url.Parse(parts[1])
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid peer URL %s", parts[1])
		}

		peers[id] = peerURL
	}

	return peers, nil
}

func (c ConnectParameters) ElectionInterval() time.Duration {
	return time.Duration(c.ElectionTick) * c.TickInterval
}

func (c ConnectParameters) RaftConfig(storage raft.Storage) *raft.Config {
	return &raft.Config{
		ID:              c.NodeID,
		ElectionTick:    c.ElectionTick,
		HeartbeatTick:   c.HeartbeatTick,
		MaxSizePerMsg:   c.MaxSizePerMsg,
		MaxInflightMsgs: c.MaxInflightMsgs,
		Storage:         storage,
	}
}

var (
	connectParameters ConnectParameters
)

func init() {
	flag.Uint64Var(&connectParameters.NodeID, "raft_node_id", 0, "Raft node ID for this instance (must be non-zero and unique within the cluster).")
	flag.Uint64Var(&connectParameters.ClusterID, "raft_cluster_id", 1, "ID of the cluster, used to isolate different Raft clusters running in the same network (must be the same for all nodes in the cluster).")
	flag.StringVar(&connectParameters.DataDir, "raft_datadir", defaultDataDir, "Directory for raft data (WAL segments and snapshots), required for restarts. These should not be deleted while the node is running or across restarts unless the node is being permanently shut down.")

	flag.Uint64Var(&connectParameters.SnapshotCatchupEntries, "raft_snapshot_catchup_entries", defaultSnapshotCatchupEntries,
		"Log entries retained after compaction so a slow follower can catch up via replication rather than a full snapshot. Higher values tolerate slower followers but increase disk usage.")
	flag.Uint64Var(&connectParameters.SnapshotIntervalEntries, "raft_snapshot_interval_entries", defaultSnapshotIntervalEntries,
		"Applied Raft log entries to accumulate before triggering a snapshot. Lower values reduce recovery time but increase I/O frequency.")
	flag.DurationVar(&connectParameters.TickInterval, "raft_tick_interval", defaultTickInterval,
		"Base time unit for Raft's logical clock. Election timeout = raft_election_tick x this value; heartbeat interval = raft_heartbeat_tick x this value. Smaller values improve responsiveness but increase network traffic.")

	flag.IntVar(&connectParameters.ElectionTick, "raft_election_tick", defaultElectionTick,
		"Ticks a follower waits without a leader heartbeat before starting an election. Effective timeout = raft_election_tick x raft_tick_interval. Must be greater than raft_heartbeat_tick. Higher values tolerate slow leaders but delay failover.")
	flag.IntVar(&connectParameters.HeartbeatTick, "raft_heartbeat_tick", defaultHeartbeatTick,
		"Ticks between leader heartbeats. Effective interval = raft_heartbeat_tick x raft_tick_interval. Must be less than raft_election_tick. Lower values detect follower loss faster but increase network traffic.")
	flag.Uint64Var(&connectParameters.MaxSizePerMsg, "raft_max_size_per_msg", defaultMaxSizePerMsg, "Maximum bytes in a single Raft message sent to a peer. Smaller values lower the recovery cost but increase the number of messages sent, affecting throughput during replication.")
	flag.IntVar(&connectParameters.MaxInflightMsgs, "raft_max_inflight_msgs", defaultMaxInflightMsgs, "Maximum number of in-flight Raft messages during optimistic replication phase. This should be set to avoid overflowing the transport layer sending buffer.")
}

// GetConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetConnectParameters(subfolder string) (ConnectParameters, error) {
	if connectParameters.NodeID == 0 {
		return ConnectParameters{}, stacktrace.NewError("--raft_node_id is required and must be non-zero")
	}
	p := connectParameters
	p.DataDir = filepath.Join(connectParameters.DataDir, subfolder)
	return p, nil
}

func GetClusterID() uint64 {
	return connectParameters.ClusterID
}
