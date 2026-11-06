package raft

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Describe the DSS node in its Raft groups",
	Long: `Describe the DSS node in its Raft groups, all of them unless selected with --groups: its state (waiting, member or
removed), ID and URL in each group.`,
	RunE: node,
}

func node(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		n, err := g.Node(cmd.Context())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s (ID %s, %s)", n.State, n.Id, n.Url), nil
	})
}
