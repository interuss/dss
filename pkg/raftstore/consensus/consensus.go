package consensus

import (
	"go.etcd.io/etcd/rafthttp"
	"go.etcd.io/raft/v3"
)

type Consensus struct {
	node        raft.Node
	raftStorage *raft.MemoryStorage

	transport *rafthttp.Transport

	storage *storage
}
