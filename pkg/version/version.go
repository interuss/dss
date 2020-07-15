package version

import (
	"fmt"
	"strconv"

	"github.com/coreos/go-semver/semver"
)

var (
	current     semver.Version
	tag, commit string
)

func init() {
	var (
		major, minor, patch int64
	)

	if n, err := fmt.Sscanf("v%d.%d.%d", tag, &major, &minor, &patch); n == 3 && err == nil {
		current.Major = major
		current.Minor = minor
		current.Patch = patch
	}

	if commit != "" {
		current.Metadata = commit
	} else {
		current.Metadata = "undefined"
	}
}

func parseIntOrDefault(s string, defaultValue int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

// Current returns the current version.
func Current() semver.Version {
	return current
}
