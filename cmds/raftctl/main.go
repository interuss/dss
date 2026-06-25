package main

import (
	"github.com/interuss/dss/cmds/raftctl/node"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "raftctl",
	Short: "Manage a DSS instance's raftstore clusters",
}

func init() {
	rootCmd.AddCommand(node.NodeCmd)
}

func main() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
