package params

import (
	"flag"
)

type (
	// ConnectParameters bundles up parameters used for connecting to a raft cluster.
	ConnectParameters struct {
		Peers string
	}
)

var (
	connectParameters ConnectParameters
)

func init() {
	flag.StringVar(&connectParameters.Peers, "raft_peers", "", "comma-separated list of raft cluster peers, e.g. node-1=10.0.0.1:7000,node-2=10.0.0.2:7000")
}

// ConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetConnectParameters() ConnectParameters {
	return connectParameters
}
