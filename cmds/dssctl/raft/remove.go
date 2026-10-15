package raft

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a node from the Raft groups of the DSS node",
		Long: `Remove a node from the Raft groups of the DSS node, all of them unless selected with --groups.

The ID of a removed node is retired permanently: it can never be added again to the same group.
A voter cannot be removed if the group would be left with fewer than 3 voters.`,
		RunE: remove,
	}

	removeNodeID uint64
)

func init() {
	removeCmd.Flags().Uint64Var(&removeNodeID, "node_id", 0, "ID of the node to remove (required)")
	cobra.CheckErr(removeCmd.MarkFlagRequired("node_id"))
}

func remove(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		if _, err := g.RemoveNode(cmd.Context(), removeNodeID); err != nil {
			return "", err
		}
		return fmt.Sprintf("removed node %d", removeNodeID), nil
	})
}
