package raft

import (
	"context"
	"fmt"

	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
)

var (
	removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a node from the Raft groups of the DSS node",
		Long: `Remove a node from the Raft groups of the DSS node, all of them unless selected with --groups.

A node has a different ID in each group: it is removed by --url from all the selected groups, or by --node_id from a
single group selected with --groups.

The ID of a removed node is retired permanently: it can never be added again to the same group.
A voter cannot be removed if the group would be left with fewer than 3 voters.`,
		RunE: remove,
	}

	removeURL    string
	removeNodeID string
)

func init() {
	removeCmd.Flags().StringVar(&removeURL, "url", "", "URL of the node to remove, i.e. its --raft_peer_url")
	removeCmd.Flags().StringVar(&removeNodeID, "node_id", "", "ID of the node to remove, only when a single group is selected with --groups")
	removeCmd.MarkFlagsMutuallyExclusive("url", "node_id")
	removeCmd.MarkFlagsOneRequired("url", "node_id")
}

func remove(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}
	if removeNodeID != "" && len(groups) != 1 {
		return stacktrace.NewError("--node_id identifies a node in a single group, select it with --groups")
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		nodeID := removeNodeID
		if removeURL != "" {
			var err error
			nodeID, err = memberID(cmd.Context(), g, removeURL)
			if err != nil {
				return "", err
			}
		}

		if err := g.RemoveNode(cmd.Context(), nodeID); err != nil {
			return "", err
		}
		return fmt.Sprintf("removed node %s", nodeID), nil
	})
}

// memberID returns the ID of the member of g at url.
func memberID(ctx context.Context, g namedGroup, url string) (string, error) {
	membership, err := g.Members(ctx)
	if err != nil {
		return "", err
	}

	for _, node := range membership.Nodes {
		if node.State == stateMember && node.Url == url {
			return node.Id, nil
		}
	}
	return "", stacktrace.NewError("no member has the URL %q", url)
}
