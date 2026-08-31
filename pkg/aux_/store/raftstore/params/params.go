package params

import (
	"flag"

	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
)

const peersFlag = "aux_raft_peers"

var peers string

func init() {
	flag.StringVar(&peers, peersFlag, "", `Comma-separated "nodeID=peerURL" pairs for the aux store, e.g. "1=http://node1:9031,2=http://node2:9031,3=http://node3:9031"`)
}

func GetConnectParameters() (raftparams.ConnectParameters, error) {
	if peers == "" {
		return raftparams.ConnectParameters{}, stacktrace.NewError("--%s is required", peersFlag)
	}

	p, err := raftparams.GetConnectParameters("aux")
	if err != nil {
		return raftparams.ConnectParameters{}, err
	}
	p.Peers = peers
	return p, nil
}
