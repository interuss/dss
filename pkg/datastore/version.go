package datastore

import (
	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
	"go.uber.org/multierr"
	"regexp"
)

type Type string

const (
	CockroachDB Type = "cockroachdb"
	Yugabyte    Type = "yugabyte"
)

type Version struct {
	SemVer *semver.Version
	Type   Type
}

var cockroachDBRegex = regexp.MustCompile(`v((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)
var yugabyteRegex = regexp.MustCompile(`-YB-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)

func parseVersion(fullVersion string, regex *regexp.Regexp) (*semver.Version, error) {
	match := regex.FindStringSubmatch(fullVersion)
	if len(match) < 2 {
		return nil, stacktrace.NewError("Unable to extract version from %s using %s", fullVersion, regex.String())
	}
	return semver.NewVersion(match[1])
}

func NewVersion(fullVersion string) (*Version, error) {
	v, err := parseVersion(fullVersion, cockroachDBRegex)
	if err == nil {
		return &Version{v, CockroachDB}, nil
	}

	v, err2 := parseVersion(fullVersion, yugabyteRegex)
	if err2 == nil {
		return &Version{v, Yugabyte}, nil
	}

	return nil, stacktrace.Propagate(multierr.Combine(err, err2), "Unable to extract datastore type and version")
}
