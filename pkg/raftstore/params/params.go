package params

import (
	"flag"
	"fmt"
	"hash/fnv"
	"net/url"
	"path/filepath"
	"time"

	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
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
		// URL is the URL at which the other nodes reach this node.
		URL string

		// DataDir is the join between the base data directory provided by the user and a subfolder for the specific group (aux, rid, scd).
		DataDir   string
		ClusterID uint64
		// Group is the name of the group (aux, rid, scd).
		Group string
		// GroupClusterID identifies the group (aux, rid, scd) within the cluster.
		// It is derived from ClusterID and is used by the transport since the groups share a listener (see consensus/listener.go).
		GroupClusterID uint64

		SnapshotCatchupEntries  uint64
		SnapshotIntervalEntries uint64
		TickInterval            time.Duration

		ElectionTick  int
		HeartbeatTick int

		MaxSizePerMsg   uint64
		MaxInflightMsgs int

		// TLSCertFile, TLSKeyFile and TLSCAFile configure mutual TLS between Raft nodes. They are required unless Insecure is set.
		TLSCertFile string
		TLSKeyFile  string
		TLSCAFile   string
		// Insecure runs Raft between nodes over plain HTTP, without authentication, instead of mutual TLS.
		Insecure bool
	}
)

// TLSEnabled reports whether Raft nodes authenticate each other with mutual TLS, which they do unless Insecure is set.
func (c ConnectParameters) TLSEnabled() bool {
	return !c.Insecure
}

// URLScheme is the scheme of the URLs at which the Raft nodes reach each other: https with mutual TLS, http otherwise.
func (c ConnectParameters) URLScheme() string {
	if c.TLSEnabled() {
		return "https"
	}
	return "http"
}

// RaftPeerUrl returns --raft_peer_url, the URL at which the other nodes reach this node.
func (c ConnectParameters) RaftPeerUrl() (*url.URL, error) {
	nodeURL, err := url.Parse(c.URL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid --raft_peer_url %s", c.URL)
	}
	return nodeURL, nil
}

func (c ConnectParameters) ElectionInterval() time.Duration {
	return time.Duration(c.ElectionTick) * c.TickInterval
}

func (c ConnectParameters) RaftConfig(nodeID uint64, storage raft.Storage) *raft.Config {
	return &raft.Config{
		ID:              nodeID,
		ElectionTick:    c.ElectionTick,
		HeartbeatTick:   c.HeartbeatTick,
		MaxSizePerMsg:   c.MaxSizePerMsg,
		MaxInflightMsgs: c.MaxInflightMsgs,
		Storage:         storage,
		// a removed leader steps down so that the remaining nodes elect a new leader
		StepDownOnRemoval: true,
	}
}

var (
	connectParameters ConnectParameters
)

func init() {
	flag.StringVar(&connectParameters.URL, "raft_peer_url", "", "URL at which the other DSS nodes of the cluster reach this node, unique within the cluster, e.g. http://node1:9011. The Raft groups (aux, rid, scd) of the node share a single listener on its port.")
	flag.Uint64Var(&connectParameters.ClusterID, "raft_cluster_id", 1, "ID of the cluster, used to isolate different Raft clusters running in the same network (must be the same for all nodes in the cluster).")
	flag.StringVar(&connectParameters.DataDir, "raft_datadir", defaultDataDir, "Directory for raft data (node identity, WAL segments and snapshots), required for restarts. These should not be deleted while the node is running or across restarts unless the node is being permanently shut down. A node starting with an empty directory gets a new ID and waits to be added to the cluster: if the directory is lost, the node must be added again and its old ID removed.")

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

	flag.StringVar(&connectParameters.TLSCertFile, "raft_tls_cert_file", "", "PEM certificate of this node, presented to the other Raft nodes. Required with --raft_tls_key_file and --raft_tls_ca_file for mutual TLS between Raft nodes, unless --raft_insecure is set. Peer URLs must then use https.")
	flag.StringVar(&connectParameters.TLSKeyFile, "raft_tls_key_file", "", "PEM private key of --raft_tls_cert_file.")
	flag.StringVar(&connectParameters.TLSCAFile, "raft_tls_ca_file", "", "PEM bundle of the CA certificates that sign the certificates of the other Raft nodes, usually one CA per participant.")
	flag.BoolVar(&connectParameters.Insecure, "raft_insecure", false, "Run the Raft traffic between nodes over plain HTTP, without authentication, instead of mutual TLS. Peer URLs must then use http. For local development and tests only.")
}

// GetConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetConnectParameters(subfolder string) (ConnectParameters, error) {
	if _, err := types.NewURLs([]string{connectParameters.URL}); err != nil {
		return ConnectParameters{}, stacktrace.Propagate(err, "--raft_peer_url is required and must be a valid URL with a port and without a path")
	}
	p := connectParameters
	if p.Insecure && (p.TLSCertFile != "" || p.TLSKeyFile != "" || p.TLSCAFile != "") {
		return ConnectParameters{}, stacktrace.NewError("--raft_insecure cannot be used with --raft_tls_cert_file, --raft_tls_key_file or --raft_tls_ca_file")
	}
	if !p.Insecure && (p.TLSCertFile == "" || p.TLSKeyFile == "" || p.TLSCAFile == "") {
		return ConnectParameters{}, stacktrace.NewError("--raft_tls_cert_file, --raft_tls_key_file and --raft_tls_ca_file are required unless --raft_insecure is set")
	}
	if nodeURL, err := url.Parse(p.URL); err != nil || nodeURL.Scheme != p.URLScheme() {
		return ConnectParameters{}, stacktrace.NewError("invalid --raft_peer_url %s: scheme must be %s (https when mutual TLS is enabled, http otherwise)", p.URL, p.URLScheme())
	}
	p.DataDir = filepath.Join(connectParameters.DataDir, subfolder)
	p.Group = subfolder
	groupID, err := groupClusterID(connectParameters.ClusterID, subfolder)
	if err != nil {
		return ConnectParameters{}, stacktrace.Propagate(err, "failed to derive the cluster ID of the %s group", subfolder)
	}
	p.GroupClusterID = groupID
	return p, nil
}

// groupClusterID derives an ID for the group from the cluster ID.
func groupClusterID(clusterID uint64, group string) (uint64, error) {
	h := fnv.New64a()
	_, err := fmt.Fprintf(h, "%d/%s", clusterID, group)
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to hash the cluster ID and the group")
	}
	return h.Sum64(), nil
}
