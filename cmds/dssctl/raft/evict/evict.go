package evict

import (
	"github.com/interuss/dss/cmds/dssctl/raft/evict/config"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "evict",
	Short: "Evict the expired entities",
	Long: `Evict the expired entities through its admin API.
	The leader of each group runs a cleanup job at the interval and using the parameters of its eviction settings.`,
}

func init() {
	Cmd.AddCommand(runCmd, config.Cmd)
}
