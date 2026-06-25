package node

import (
	"github.com/spf13/cobra"
	"go.etcd.io/raft/v3/raftpb"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a node from every raftstore cluster known to the target DSS instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return changeMembership(addr, raftpb.ConfChangeRemoveNode, nodeID, nil)
	},
}
