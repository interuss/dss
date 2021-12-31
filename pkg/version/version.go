package version

import (
	"fmt"

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

	if n, err := fmt.Sscanf(tag, "v%d.%d.%d", &major, &minor, &patch); n == 3 && err == nil {
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

// Current returns the current version.
func Current() semver.Version {
	return current
}
