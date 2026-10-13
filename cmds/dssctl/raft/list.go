package raft

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	restapi "github.com/interuss/dss/pkg/api/adminv1"
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
func formatMembers(m *restapi.RaftGroup) (string, error) {
	var builder strings.Builder
	builder.WriteString("\n")
	if m.LeaderId == "" {
		builder.WriteString("  no leader known\n")
	}

	table := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
	_, err := fmt.Fprintln(table, "  ID\tURL\tROLE")
	if err != nil {
		return "", err
	}
	var members []restapi.RaftNode
	var removed []string
	for _, node := range m.Nodes {
		if node.State == stateRemoved {
			removed = append(removed, node.Id)
			continue
		}
		members = append(members, node)
	}
	// The URLs are sorted rather than the IDs, which differ between groups, so that the groups list the nodes in the same order.
	slices.SortFunc(members, func(a, b restapi.RaftNode) int { return cmp.Compare(a.Url, b.Url) })
	for _, node := range members {
		_, err := fmt.Fprintf(table, "  %s\t%s\t%s\n", node.Id, node.Url, memberRole(m, node.Id))
		if err != nil {
			return "", err
		}
	}
	_ = table.Flush()

	if len(removed) > 0 {
		fmt.Fprintf(&builder, "  removed IDs: %v\n", removed)
	}
	return strings.TrimSuffix(builder.String(), "\n"), nil
}

func memberRole(m *restapi.RaftGroup, id string) string {
	if id == m.LeaderId {
		return "leader"
	}
	// all the members of a Raft group are voters
	return "voter"
}
