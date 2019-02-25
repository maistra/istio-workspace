package version

var (
	// Version hold a semantic version of the running binary
	Version = "0.0.1"
	// Commit holds the commit hash against which the binary build was ran
	Commit string
	// BuildTime holds timestamp when the binary build was ran
	BuildTime string
)
