package config

import (
	"fmt"
	"strconv"
	"strings"

	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the eviction settings of the Raft groups",
}

func init() {
	Cmd.AddCommand(getCmd, setCmd, disableCmd, enableCmd)
}

func formatState(state *restapi.EvictionState) string {
	lastRun := "never"
	if state.LastRun != nil {
		lastRun = *state.LastRun
	}
	return formatConfig(&state.Config, state.Disabled != nil && *state.Disabled) + "\n  last run: " + lastRun
}

func formatConfig(config *restapi.EvictionConfig, disabled bool) string {
	interval := config.Interval
	if disabled {
		interval += " (scheduled cleanup jobs disabled)"
	}
	limit := "unlimited"
	if l := config.Parameters.Limit; l != nil && *l != 0 {
		limit = strconv.Itoa(int(*l))
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "\n  interval: %s", interval)
	fmt.Fprintf(&builder, "\n  ttl: %s", config.Parameters.Ttl)
	fmt.Fprintf(&builder, "\n  limit: %s", limit)
	return builder.String()
}
