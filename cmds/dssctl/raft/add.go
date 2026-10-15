package raft

import (
	"fmt"

	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add new a node to the Raft groups of the DSS node",
		Long: `Add new a node to the Raft groups of the DSS node, all of them unless selected with --groups.

The node must be running before it is added, started with --raft_join and with its own entry in the
--<group>_raft_peers flags: as soon as it is added, it counts in the quorum of the group.`,
		RunE: add,
	}

	addNodeID uint64
	peerURLs  = map[string]*string{}
)

func init() {
	addCmd.Flags().Uint64Var(&addNodeID, "node_id", 0, "ID of the node to add, its --raft_node_id (required)")
	for _, name := range groupNames {
		peerURLs[name] = addCmd.Flags().String(name+"_url", "", fmt.Sprintf("URL at which the other members of the %s Raft group reach the node, as in its --%s_raft_peers flag (required when adding the node to %s)", name, name, name))
	}
	cobra.CheckErr(addCmd.MarkFlagRequired("node_id"))
}

func add(cmd *cobra.Command, _ []string) error {
	groups, err := selectedGroups()
	if err != nil {
		return err
	}

	for _, g := range groups {
		if *peerURLs[g.name] == "" {
			return stacktrace.NewError("--%s_url is required to add the node to the %s Raft group", g.name, g.name)
		}
	}

	return forEachGroup(groups, func(g namedGroup) (string, error) {
		peerURL := *peerURLs[g.name]
		_, err := g.AddNode(cmd.Context(), addNodeID, peerURL)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("added node %d (%s)", addNodeID, peerURL), nil
	})
}
