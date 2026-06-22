package params

import (
	"flag"

	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
)

const peersFlag = "scd_raft_peers"

var peers string

func init() {
	flag.StringVar(&peers, peersFlag, "", `Comma-separated "nodeID=peerURL" pairs for the scd store, e.g. "1=http://node1:9021,2=http://node2:9021,3=http://node3:9021"`)
}

func GetConnectParameters() (raftparams.ConnectParameters, error) {
	if peers == "" {
		return raftparams.ConnectParameters{}, stacktrace.NewError("--%s is required", peersFlag)
	}

	p, err := raftparams.GetConnectParameters("scd")
	if err != nil {
		return raftparams.ConnectParameters{}, err
	}
	p.Peers = peers
	return p, nil
}
