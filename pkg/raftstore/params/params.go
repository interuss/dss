package params

import (
	"flag"
	"net/url"
	"strconv"
	"strings"

	"github.com/interuss/stacktrace"
)

const (
	defaultDataDir = "raft_data"

	// the default Raft related parameters are the same as the default values used by etcd for the moment.
	// TODO - review and adjust these parameters as needed based on testing and performance tuning.
	defaultSnapshotCatchupEntries = 10000
)

type (
	// ConnectParameters bundles up parameters used for connecting nodes in a raftstore cluster.
	ConnectParameters struct {
		ID    uint64
		Peers string

		// DataDir is the directory where the node persists its Raft state (WAL segments and snapshots).
		// This data is required for the node to restart and rejoin the cluster without having to receive
		// a full snapshot from the leader. It must not be deleted while the node is running or
		// across restarts unless the node is being permanently shut down.
		// If the directory is lost, the node will recover by receiving a snapshot from the leader.
		DataDir string

		// SnapshotCatchupEntries is the number of entries for a slow follower to catch-up after compacting.
		// This gives the follower a buffer of entries while avoiding the need to send a full snapshot.
		SnapshotCatchupEntries uint64
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

var (
	connectParameters ConnectParameters
)

func init() {
	flag.Uint64Var(&connectParameters.ID, "raft_node_id", 0, "raft node ID for this instance (must be non-zero and unique within the cluster)")
	flag.StringVar(&connectParameters.Peers, "raft_peers", "", `comma-separated "nodeID=peerURL" pairs for all cluster members, including the current node, e.g. "1=http://node1:9021,2=http://node2:9021,3=http://node3:9021"`)
	flag.StringVar(&connectParameters.DataDir, "raft_datadir", defaultDataDir, "directory for raft data (WAL segments and snapshots), required for restarts. These should not be deleted while the node is running or across restarts unless the node is being permanently shut down.")

	flag.Uint64Var(&connectParameters.SnapshotCatchupEntries, "raft_snapshot_catchup_entries", defaultSnapshotCatchupEntries, "number of entries for a slow follower to catch-up after compacting")
}

// GetConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetConnectParameters() ConnectParameters {
	return connectParameters
}
