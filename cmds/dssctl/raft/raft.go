package raft

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/interuss/dss/pkg/admin/client"
	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
)

// groupNames are the Raft groups of a DSS node, in the order in which they are processed.
var groupNames = []string{"aux", "rid", "scd"}

// The states of a node in a Raft group, as returned by the admin API.
const (
	stateMember  = "member"
	stateRemoved = "removed"
)

var (
	RaftCmd = &cobra.Command{
		Use:   "raft",
		Short: "Manage the membership of the Raft groups (aux, rid, scd) of a DSS node",
		Long: `Manage the membership of the Raft groups (aux, rid, scd) of a DSS node, through its admin API.

A change is requested to each selected group independently: it may succeed for some groups and fail for others,
in which case it can be retried for the failed groups only with --groups.`,
	}

	groups       []string
	adminSocket  string
	adminTimeout time.Duration
)

func init() {
	RaftCmd.PersistentFlags().StringSliceVar(&groups, "groups", groupNames, "Comma-separated Raft groups to act on, all of them if not set")
	RaftCmd.PersistentFlags().StringVar(&adminSocket, "admin_socket", "/run/dss/admin.sock", "Path of the Unix socket of the admin API of the DSS node, matching its --admin_socket")
	RaftCmd.PersistentFlags().DurationVar(&adminTimeout, "admin_timeout", time.Minute, "Timeout of each request to the admin API")
	RaftCmd.AddCommand(bootstrapCmd, addCmd, removeCmd, listCmd, nodeCmd)
}

// namedGroup is a Raft group of the DSS node, managed through its admin API.
type namedGroup struct {
	name string
	*client.RaftGroup
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
		group, err := client.NewRaftGroup(adminSocket, name, adminTimeout)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid --admin_socket")
		}
		result = append(result, namedGroup{name: name, RaftGroup: group})
	}
	if len(result) == 0 {
		return nil, stacktrace.NewError("no Raft group selected with --groups")
	}
	return result, nil
}

// forEachGroup calls fn for every group concurrently,
// prints the result of each group as soon as it completes
// and returns an error naming the failed groups.
func forEachGroup(groups []namedGroup, fn func(g namedGroup) (string, error)) error {
	var (
		mu     sync.Mutex
		failed []string
		wg     sync.WaitGroup
	)
	for _, g := range groups {
		wg.Go(func() {
			msg, err := fn(g)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				// %#s formats stacktrace errors on a single line, without the stack.
				fmt.Fprintf(os.Stderr, "%s: %#s\n", g.name, err)
				failed = append(failed, g.name)
				return
			}
			fmt.Printf("%s: %s\n", g.name, msg)
		})
	}
	wg.Wait()

	if len(failed) == 0 {
		return nil
	}
	slices.SortFunc(failed, func(a, b string) int { return slices.Index(groupNames, a) - slices.Index(groupNames, b) })
	return fmt.Errorf("failed for Raft groups %s", strings.Join(failed, ","))
}
