package node

import (
	"github.com/spf13/cobra"
	"go.etcd.io/raft/v3/raftpb"
)

var updateAddresses map[string]string

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a node's address across every raftstore cluster known to the target DSS instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return changeMembership(addr, raftpb.ConfChangeUpdateNode, nodeID, updateAddresses)
	},
}

func init() {
	updateCmd.Flags().StringToStringVar(&updateAddresses, "addresses", nil, "new host:port the node's Raft transport will listen on for each store, e.g. \"rid=node4-new:8081,scd=node4-new:8082,aux=node4-new:8083\" (required)")
	err := updateCmd.MarkFlagRequired("addresses")
	cobra.CheckErr(err)
}
