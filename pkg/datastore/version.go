package datastore

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
	"go.uber.org/multierr"
	"regexp"
)

type VersionRegex struct {
	Name         string
	VersionRegex *regexp.Regexp
}

type Version struct {
	version *semver.Version
	dsType  string
}

var cockroachDB = &VersionRegex{"cockroach", regexp.MustCompile(`v((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)}
var yugabyte = &VersionRegex{"yugabyte", regexp.MustCompile(`-YB-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)}

func (dt *VersionRegex) parseVersion(fullVersion string) (*semver.Version, error) {
	match := dt.VersionRegex.FindStringSubmatch(fullVersion)
	if len(match) < 2 {
		return nil, stacktrace.NewError("Unable to extract version for %s from %s using %s", dt.Name, fullVersion, dt.VersionRegex)
	}
	return semver.NewVersion(match[1])
}

func versionFromString(fullVersion string) (*Version, error) {
	version, err := cockroachDB.parseVersion(fullVersion)
	if err == nil {
		return &Version{version, cockroachDB.Name}, nil
	}

	version, err2 := yugabyte.parseVersion(fullVersion)
	if err2 == nil {
		return &Version{version, yugabyte.Name}, nil
	}

	return nil, stacktrace.Propagate(multierr.Combine(err, err2), "Unable to extract datastore type and version")
}

func (m *Version) String() string {
	return fmt.Sprintf("%s@%s", m.dsType, m.version.String())
}

func (m *Version) Version() *semver.Version {
	return m.version
}

func (m *Version) DatastoreType() string {
	return m.dsType
}

func (m *Version) IsCockroachDB() bool {
	return m.dsType == cockroachDB.Name
}

func (m *Version) IsYugabyte() bool {
	return m.dsType == yugabyte.Name
}
