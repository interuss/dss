package consensus

import (
	"github.com/coreos/etcd/snap"
	"go.etcd.io/etcd/server/v3/storage/wal"
)

// storage persists the Raft log and snapshots.
type storage struct {
	wal         *wal.WAL
	snapshotter *snap.Snapshotter
}
