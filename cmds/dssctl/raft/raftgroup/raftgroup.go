package raftgroup

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/interuss/dss/cmds/dssctl/admin"
	"github.com/interuss/dss/pkg/admin/client"
	"github.com/interuss/stacktrace"
	"github.com/spf13/pflag"
)

var Names = []string{"aux", "rid", "scd"}

var EvictionGroupNames = []string{"rid", "scd"}

var selected []string

// RegisterFlags registers the flags selecting the Raft groups.
func RegisterFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&selected, "groups", Names, "Comma-separated Raft groups to act on, all of them if not set")
}

// Group is a Raft group of the DSS node, managed through its admin API.
type Group struct {
	Name string
	*client.RaftGroup
}

// Selected returns the Raft groups selected with --groups, in processing order.
func Selected() ([]Group, error) {
	for _, name := range selected {
		if !slices.Contains(Names, name) {
			return nil, stacktrace.NewError("unknown Raft group %q in --groups, expected some of %s", name, strings.Join(Names, ","))
		}
	}

	var names []string
	for _, name := range Names {
		if slices.Contains(selected, name) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, stacktrace.NewError("no Raft group selected with --groups")
	}
	return New(names)
}

// SelectedEviction returns the Raft groups with eviction selected with --groups, all of them if not set.
func SelectedEviction(fs *pflag.FlagSet) ([]Group, error) {
	if !fs.Changed("groups") {
		return New(EvictionGroupNames)
	}
	groups, err := Selected()
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		if !slices.Contains(EvictionGroupNames, g.Name) {
			return nil, stacktrace.NewError("the %s Raft group has no eviction, expected some of %s in --groups", g.Name, strings.Join(EvictionGroupNames, ","))
		}
	}
	return groups, nil
}

// New returns the Raft groups of names, in the order of names.
func New(names []string) ([]Group, error) {
	result := make([]Group, 0, len(names))
	for _, name := range names {
		group, err := client.NewRaftGroup(admin.Socket, name, admin.Timeout)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid --admin_socket")
		}
		result = append(result, Group{Name: name, RaftGroup: group})
	}
	return result, nil
}

// ForEach calls fn for every group concurrently,
// prints the result of each group as soon as it completes
// and returns an error naming the failed groups.
func ForEach(groups []Group, fn func(g Group) (string, error)) error {
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
				fmt.Fprintf(os.Stderr, "%s: %#s\n", g.Name, err)
				failed = append(failed, g.Name)
				return
			}
			fmt.Printf("%s: %s\n", g.Name, msg)
		})
	}
	wg.Wait()

	if len(failed) == 0 {
		return nil
	}
	slices.SortFunc(failed, func(a, b string) int { return slices.Index(Names, a) - slices.Index(Names, b) })
	return fmt.Errorf("failed for Raft groups %s", strings.Join(failed, ","))
}
