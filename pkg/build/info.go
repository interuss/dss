package build

// Constants describing the Build
var (
	time   = "undefined"
	commit = "undefined"
	host   = "undefined"
)

// Description bundles up information about a build.
type Description struct {
	Time   string // The timestamp of the build.
	Commit string // The commit hash of the build.
	Host   string // The host where the build happened.
}

// Describe returns a Description of a build.
func Describe() Description {
	return Description{
		Time:   time,
		Commit: commit,
		Host:   host,
	}
}
