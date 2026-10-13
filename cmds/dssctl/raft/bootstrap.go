package raft

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap new Raft groups with the DSS node as their only member",
	Long: `Bootstrap new Raft groups with the DSS node as their only member, all of them unless selected with --groups.

The DSS node must be waiting, i.e. started with an empty data directory. The other nodes are then added to the groups
with dssctl raft add.`,
	RunE: bootstrap,
}

func bootstrap(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		membership, err := g.Bootstrap(cmd.Context())
		if err != nil {
			return "", err
		}
		// Bootstrap returns once the DSS node is the leader of the new group.
		return fmt.Sprintf("bootstrapped with node %s as leader", membership.LeaderId), nil
	})
}
