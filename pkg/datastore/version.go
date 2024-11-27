package datastore

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
	"go.uber.org/multierr"
	"regexp"
)

type versionRegex struct {
	Name         string
	VersionRegex *regexp.Regexp
}

type version struct {
	version *semver.Version
	dsType  string
}

var cockroachDB = &versionRegex{"cockroach", regexp.MustCompile(`v((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)}
var yugabyte = &versionRegex{"yugabyte", regexp.MustCompile(`-YB-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)}

func (dt *versionRegex) parseVersion(fullVersion string) (*semver.Version, error) {
	match := dt.VersionRegex.FindStringSubmatch(fullVersion)
	if len(match) < 2 {
		return nil, stacktrace.NewError("Unable to extract version for %s from %s using %s", dt.Name, fullVersion, dt.VersionRegex)
	}
	return semver.NewVersion(match[1])
}

func versionFromString(fullVersion string) (*version, error) {
	v, err := cockroachDB.parseVersion(fullVersion)
	if err == nil {
		return &version{v, cockroachDB.Name}, nil
	}

	v, err2 := yugabyte.parseVersion(fullVersion)
	if err2 == nil {
		return &version{v, yugabyte.Name}, nil
	}

	return nil, stacktrace.Propagate(multierr.Combine(err, err2), "Unable to extract datastore type and version")
}

func (m *version) String() string {
	return fmt.Sprintf("%s@%s", m.dsType, m.version.String())
}

func (m *version) Version() *semver.Version {
	return m.version
}

func (m *version) IsCockroachDB() bool {
	return m.dsType == cockroachDB.Name
}

func (m *version) IsYugabyte() bool {
	return m.dsType == yugabyte.Name
}
