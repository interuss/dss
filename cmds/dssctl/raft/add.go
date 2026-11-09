package raft

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a node to the Raft groups of the DSS node",
		Long: `Add a node to the Raft groups of the DSS node, all of them unless selected with --groups.

The node to add must be running and waiting, i.e. started with an empty data directory: as soon as it is added, it
counts in the quorum of the group. The DSS node reads its ID in each group from the node itself, at --url.`,
		RunE: add,
	}

	addURL string
)

func init() {
	addCmd.Flags().StringVar(&addURL, "url", "", "URL of the node to add, i.e. its --raft_peer_url (required)")
	cobra.CheckErr(addCmd.MarkFlagRequired("url"))
}

func add(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		node, err := g.AddNode(cmd.Context(), addURL)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("added node %s (%s)", node.Id, node.Url), nil
	})
}
