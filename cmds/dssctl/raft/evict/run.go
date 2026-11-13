package evict

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/interuss/dss/cmds/dssctl/raft/raftgroup"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run a cleanup job now",
		Long: `Run a cleanup job now on the Raft groups selected with --groups or both rid and scd if unset.

The TTL and limit of a group are those of its eviction settings, unless both given with --<group>_ttl and
--<group>_limit.`,
		RunE: run,
	}

	runFlagValues runFlags
)

func init() {
	runFlagValues.register(runCmd.Flags())
}

type runFlags struct {
	ttls   map[string]*time.Duration
	limits map[string]*int64
}

func (f *runFlags) register(fs *pflag.FlagSet) {
	f.ttls = map[string]*time.Duration{}
	f.limits = map[string]*int64{}
	for _, group := range raftgroup.EvictionGroupNames {
		f.ttls[group] = fs.Duration(group+"_ttl", 0, fmt.Sprintf("time-to-live duration of the entities of the %s group, instead of the one of its eviction settings", group))
		f.limits[group] = fs.Int64(group+"_limit", 0, fmt.Sprintf("maximum number of entities of the %s group deleted per entity type, 0 for unlimited, instead of the one of its eviction settings", group))
	}
}

func (f *runFlags) parameters(fs *pflag.FlagSet, names []string) (map[string]*restapi.EvictionRunParameters, error) {
	result := map[string]*restapi.EvictionRunParameters{}
	for _, group := range raftgroup.EvictionGroupNames {
		ttlChanged, limitChanged := fs.Changed(group+"_ttl"), fs.Changed(group+"_limit")
		if !slices.Contains(names, group) {
			if ttlChanged || limitChanged {
				return nil, stacktrace.NewError("--%s_ttl or --%s_limit given while the %s group is not selected with --groups", group, group, group)
			}
			continue
		}
		if ttlChanged != limitChanged {
			return nil, stacktrace.NewError("--%s_ttl and --%s_limit must be given together", group, group)
		}
		params := &restapi.EvictionRunParameters{}
		if ttlChanged {
			params.Parameters = &restapi.EvictionParameters{
				Ttl:   f.ttls[group].String(),
				Limit: new(*f.limits[group]),
			}
		}
		result[group] = params
	}
	return result, nil
}

func run(cmd *cobra.Command, _ []string) error {
	groups, err := raftgroup.SelectedEviction(cmd.Flags())
	if err != nil {
		return err
	}
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}
	params, err := runFlagValues.parameters(cmd.Flags(), names)
	if err != nil {
		return err
	}

	return raftgroup.ForEach(groups, func(g raftgroup.Group) (string, error) {
		result, err := g.RunEviction(cmd.Context(), params[g.Name])
		if err != nil {
			return "", err
		}
		counts := make([]string, 0, len(result.Counts))
		for _, c := range result.Counts {
			counts = append(counts, fmt.Sprintf("%d %s", c.Count, c.Type))
		}
		return "deleted " + strings.Join(counts, ", "), nil
	})
}
