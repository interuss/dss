package node

import (
	"github.com/spf13/cobra"
	"go.etcd.io/raft/v3/raftpb"
)

var addAddresses map[string]string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a node to the raftstore",
	RunE: func(cmd *cobra.Command, args []string) error {
		return changeMembership(addr, raftpb.ConfChangeAddNode, nodeID, addAddresses)
	},
}

func init() {
	addCmd.Flags().StringToStringVar(&addAddresses, "addresses", nil, "host:port the new node's Raft transport will listen on for each store, e.g. \"rid=node4:8081,scd=node4:8082,aux=node4:8083\" (required)")
	err := addCmd.MarkFlagRequired("addresses")
	cobra.CheckErr(err)
}
