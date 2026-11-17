package raft

import (
	"github.com/interuss/dss/cmds/dssctl/raft/evict"
	"github.com/interuss/dss/cmds/dssctl/raft/raftgroup"
	"github.com/spf13/cobra"
)

// The states of a node in a Raft group, as returned by the admin API.
const (
	stateMember  = "member"
	stateRemoved = "removed"
)

var RaftCmd = &cobra.Command{
	Use:   "raft",
	Short: "Manage the membership of the Raft groups (aux, rid, scd) of a DSS node",
	Long: `Manage the membership of the Raft groups (aux, rid, scd) of a DSS node, through its admin API.

A change is requested to each selected group independently: it may succeed for some groups and fail for others,
in which case it can be retried for the failed groups only with --groups.`,
}

func init() {
	raftgroup.RegisterFlags(RaftCmd.PersistentFlags())
	RaftCmd.AddCommand(bootstrapCmd, addCmd, removeCmd, listCmd, nodeCmd, evict.Cmd)
}
