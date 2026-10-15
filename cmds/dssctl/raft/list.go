package raft

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/interuss/dss/pkg/admin"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the members of the Raft groups of the DSS node",
	Long:  "List the members of the Raft groups of the DSS node, all of them unless selected with --groups, as known by the node serving each admin API.",
	RunE:  list,
}

func list(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		members, err := g.Members(cmd.Context())
		if err != nil {
			return "", err
		}

		return formatMembers(members)
	})
}

// formatMembers formats m as a table of the members, followed by the removed IDs.
func formatMembers(m admin.Membership) (string, error) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "as known by node %d\n", m.NodeID)
	if m.LeaderID == 0 {
		builder.WriteString("  no leader known\n")
	}

	table := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
	_, err := fmt.Fprintln(table, "  ID\tURL\tROLE")
	if err != nil {
		return "", err
	}
	for _, id := range slices.Sorted(maps.Keys(m.Members)) {
		_, err := fmt.Fprintf(table, "  %d\t%s\t%s\n", id, m.Members[id], memberRole(m, id))
		if err != nil {
			return "", err
		}
	}
	_ = table.Flush()

	if len(m.RemovedIDs) > 0 {
		fmt.Fprintf(&builder, "  removed IDs: %v\n", slices.Sorted(slices.Values(m.RemovedIDs)))
	}
	return strings.TrimSuffix(builder.String(), "\n"), nil
}

func memberRole(m admin.Membership, id uint64) string {
	if id == m.LeaderID {
		return "leader"
	}
	// all the members of a Raft group are voters
	return "voter"
}
