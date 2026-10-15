package raft

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/interuss/dss/pkg/admin"
	"github.com/interuss/dss/pkg/admin/client"
	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
)

// groupNames are the Raft groups of a DSS node, in the order in which they are processed.
var groupNames = []string{"aux", "rid", "scd"}

var (
	RaftCmd = &cobra.Command{
		Use:   "raft",
		Short: "Manage the membership of the Raft groups (aux, rid, scd) of a DSS node",
		Long: `Manage the membership of the Raft groups (aux, rid, scd) of a DSS node, through its admin API.

A change is requested to each selected group independently: it may succeed for some groups and fail for others,
in which case it can be retried for the failed groups only with --groups.`,
	}

	groups []string
)

func init() {
	RaftCmd.PersistentFlags().StringSliceVar(&groups, "groups", groupNames, "Comma-separated Raft groups to act on, all of them if not set")
	RaftCmd.AddCommand(addCmd, removeCmd, listCmd)
}

// namedGroup is a Raft group of the DSS node, managed through its admin API.
type namedGroup struct {
	name string
	admin.RaftGroup
}

// selectedGroups returns the Raft groups selected with --groups, in processing order.
func selectedGroups() ([]namedGroup, error) {
	for _, name := range groups {
		if !slices.Contains(groupNames, name) {
			return nil, stacktrace.NewError("unknown Raft group %q in --groups, expected some of %s", name, strings.Join(groupNames, ","))
		}
	}

	var result []namedGroup
	for _, name := range groupNames {
		if !slices.Contains(groups, name) {
			continue
		}
		group, err := client.InitRaftGroup(name)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to initialize the %s Raft group", name)
		}
		result = append(result, namedGroup{name: name, RaftGroup: group})
	}
	if len(result) == 0 {
		return nil, stacktrace.NewError("no Raft group selected with --groups")
	}
	return result, nil
}

// forEachGroup calls fn for every group concurrently
// then prints the results of the groups that succeeded
// and returns the errors of the failed ones.
func forEachGroup(groups []namedGroup, fn func(g namedGroup) (string, error)) error {
	type result struct {
		msg string
		err error
	}

	results := make([]result, len(groups))

	var wg sync.WaitGroup
	for i, g := range groups {
		wg.Go(func() {
			msg, err := fn(g)
			results[i] = result{msg: msg, err: err}
		})
	}
	wg.Wait()

	var failures []string
	for i, g := range groups {
		if err := results[i].err; err != nil {
			// %#s formats stacktrace errors on a single line, without the stack.
			failures = append(failures, fmt.Sprintf("%s: %#s", g.name, err))
			continue
		}
		fmt.Printf("%s: %s\n", g.name, results[i].msg)
	}

	switch len(failures) {
	case 0:
		return nil
	case 1:
		return errors.New(failures[0])
	default:
		return fmt.Errorf("%d Raft groups failed:\n  %s", len(failures), strings.Join(failures, "\n  "))
	}
}
