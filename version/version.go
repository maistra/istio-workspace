package version

import "strconv"

const Latest = "latest"

var (
	// Version hold a semantic version of the running binary.
	Version = "v0.0.0"
	// Commit holds the commit hash against which the binary build was ran.
	Commit string
	// BuildTime holds timestamp when the binary build was ran.
	BuildTime string
	// Release indicates if built binary is an official release.
	Release = "false"
)

func Released() bool {
	released, _ := strconv.ParseBool(Release)

	return released
}

func CurrentVersion() string {
	currentVersion := Latest
	if Released() {
		currentVersion = Version
	}

	return currentVersion
}
