package params

import (
	"flag"

	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
)

const peersFlag = "rid_raft_peers"

var peers string

func init() {
	flag.StringVar(&peers, peersFlag, "", `Comma-separated "nodeID=peerURL" pairs for the rid store, e.g. "1=http://node1:9011,2=http://node2:9011,3=http://node3:9011"`)
}

func GetConnectParameters() (raftparams.ConnectParameters, error) {
	if peers == "" {
		return raftparams.ConnectParameters{}, stacktrace.NewError("--%s is required", peersFlag)
	}

	p, err := raftparams.GetConnectParameters("rid")
	if err != nil {
		return raftparams.ConnectParameters{}, err
	}
	p.Peers = peers
	return p, nil
}
