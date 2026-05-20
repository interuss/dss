package consensus

import (
	"go.etcd.io/raft/v3"
)

const (
	defaultClusterID uint64 = 1

	defaultElectionTick  = 20
	defaultHeartbeatTick = 1

	defaultMaxSizePerMsg   = 1024 * 1024
	defaultMaxInflightMsgs = 4096 / 8
)

func defaultConfig(storage raft.Storage) *raft.Config {
	return &raft.Config{
		ID:              defaultClusterID,
		ElectionTick:    defaultElectionTick,
		HeartbeatTick:   defaultHeartbeatTick,
		MaxSizePerMsg:   defaultMaxSizePerMsg,
		MaxInflightMsgs: defaultMaxInflightMsgs,
		Storage:         storage,
	}
}
