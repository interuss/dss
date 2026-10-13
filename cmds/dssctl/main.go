package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/interuss/dss/cmds/dssctl/raft"
	"github.com/spf13/cobra"
)

var (
	dssctlCmd = &cobra.Command{
		Use:   "dssctl",
		Short: "DSS node admin command line tool",
		// The error is printed by main.
		SilenceErrors: true,
		// We expect most issues to be related to membership validation and, therefore, unrelated to the usage of the command.
		SilenceUsage: true,
	}
)

func init() {
	dssctlCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	dssctlCmd.AddCommand(raft.RaftCmd)
}

func main() {
	if err := dssctlCmd.Execute(); err != nil {
		// %#s formats stacktrace errors without the stack.
		fmt.Fprintf(os.Stderr, "Error: %#s\n", err)
		os.Exit(1)
	}
}
