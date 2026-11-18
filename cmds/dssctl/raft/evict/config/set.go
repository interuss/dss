package config

import (
	"fmt"
	"time"

	"github.com/interuss/dss/cmds/dssctl/raft/raftgroup"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	setCmd = &cobra.Command{
		Use:   "set",
		Short: "Change the eviction settings of the Raft groups",
		Long: `Change the eviction settings of the Raft groups. The --<group>_interval, --<group>_ttl and --<group>_limit of a
group are given together, and the groups without them are left as they are.`,
		RunE: set,
	}

	setFlagValues setFlags
)

func init() {
	setFlagValues.register(setCmd.Flags())
}

// setFlags are the flags of dssctl raft evict config set.
type setFlags struct {
	intervals map[string]*time.Duration
	ttls      map[string]*time.Duration
	limits    map[string]*int64
}

func (f *setFlags) register(fs *pflag.FlagSet) {
	f.intervals = map[string]*time.Duration{}
	f.ttls = map[string]*time.Duration{}
	f.limits = map[string]*int64{}
	for _, group := range raftgroup.EvictionGroupNames {
		f.intervals[group] = fs.Duration(group+"_interval", 0, fmt.Sprintf("time between two scheduled cleanup jobs of the %s group", group))
		f.ttls[group] = fs.Duration(group+"_ttl", 0, fmt.Sprintf("time-to-live duration of the entities of the %s group", group))
		f.limits[group] = fs.Int64(group+"_limit", 0, fmt.Sprintf("maximum number of entities of the %s group deleted per entity type by a cleanup job, 0 for unlimited", group))
	}
}

func (f *setFlags) changedGroups(fs *pflag.FlagSet) ([]string, error) {
	var result []string
	for _, group := range raftgroup.EvictionGroupNames {
		interval, ttl, limit := fs.Changed(group+"_interval"), fs.Changed(group+"_ttl"), fs.Changed(group+"_limit")
		if interval != ttl || ttl != limit {
			return nil, stacktrace.NewError("--%s_interval, --%s_ttl and --%s_limit must be given together", group, group, group)
		}
		if interval {
			result = append(result, group)
		}
	}
	if len(result) == 0 {
		return nil, stacktrace.NewError("no eviction setting given")
	}
	return result, nil
}

func (f *setFlags) evictionConfig(group string) *restapi.EvictionConfig {
	return &restapi.EvictionConfig{
		Interval: f.intervals[group].String(),
		Parameters: restapi.EvictionParameters{
			Ttl:   f.ttls[group].String(),
			Limit: new(*f.limits[group]),
		},
	}
}

func set(cmd *cobra.Command, _ []string) error {
	names, err := setFlagValues.changedGroups(cmd.Flags())
	if err != nil {
		return err
	}
	groups, err := raftgroup.New(names)
	if err != nil {
		return err
	}

	return raftgroup.ForEach(groups, func(g raftgroup.Group) (string, error) {
		stored, err := g.SetEvictionConfig(cmd.Context(), setFlagValues.evictionConfig(g.Name))
		if err != nil {
			return "", err
		}
		return formatConfig(stored, false), nil
	})
}
