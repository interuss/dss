package config

import (
	"github.com/interuss/dss/cmds/dssctl/raft/raftgroup"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Print the eviction settings of the Raft groups",
	Long:  "Print the eviction settings of the Raft groups selected with --groups or both rid and scd if unset.",
	RunE:  get,
}

func get(cmd *cobra.Command, _ []string) error {
	groups, err := raftgroup.SelectedEviction(cmd.Flags())
	if err != nil {
		return err
	}

	return raftgroup.ForEach(groups, func(g raftgroup.Group) (string, error) {
		state, err := g.EvictionState(cmd.Context())
		if err != nil {
			return "", err
		}
		return formatState(state), nil
	})
}
