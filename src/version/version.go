package version

import "fmt"

// These variables are set at build time via ldflags.
// Example: go build -ldflags "-X github.com/user/janitarr/src/version.Version=v1.0.0"
var (
	// Version is the semantic version (e.g., "v1.0.0" or "v0.4.0-137-g4046f35")
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// Info returns a formatted string with all version information
func Info() string {
	return fmt.Sprintf("janitarr %s (commit: %s, built: %s)", Version, Commit, BuildDate)
}

// Short returns just the version string
func Short() string {
	return Version
}
