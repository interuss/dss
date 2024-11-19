package datastore

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
	"regexp"
)

type VersionRegex struct {
	Name         string
	VersionRegex string
}

type Version struct {
	version *semver.Version
	dsName  string
}

var COCKROACHDB = &VersionRegex{"cockroach", `v((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`}
var YUGABYTE = &VersionRegex{"yugabyte", `-YB-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`}

func (dt *VersionRegex) ParseVersion(fullVersion string) (*semver.Version, error) {
	re := regexp.MustCompile(dt.VersionRegex)
	match := re.FindStringSubmatch(fullVersion)
	if len(match) < 2 {
		return nil, stacktrace.NewError("Unable to extract version for %s from %s using %s", dt.Name, fullVersion, dt.VersionRegex)
	}
	return semver.NewVersion(match[1])
}

func (m *Version) Version() *semver.Version {
	return m.version
}

func (m *Version) DbType() string {
	return m.dsName
}

func VersionFromString(fullVersion string) (*Version, error) {
	version, err := COCKROACHDB.ParseVersion(fullVersion)
	if err == nil {
		return &Version{version, COCKROACHDB.Name}, nil
	}

	version, err2 := YUGABYTE.ParseVersion(fullVersion)
	if err2 == nil {
		return &Version{version, YUGABYTE.Name}, nil
	}

	return nil, stacktrace.Propagate(err2, "Unable to extract datastore type and version")
}

func (m *Version) IsCockroachDB() bool {
	return m.dsName == COCKROACHDB.Name
}

func (m *Version) IsYugabyte() bool {
	return m.dsName == YUGABYTE.Name
}

func (m *Version) String() string {
	return fmt.Sprintf("%s@%s", m.dsName, m.version.String())
}
