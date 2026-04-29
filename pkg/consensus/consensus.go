package consensus

import (
	"go.etcd.io/raft/v3"
)

type Consensus struct {
	node raft.Node
}
