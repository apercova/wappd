package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using -ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// Info holds version information
type Info struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("wappd version %s (commit: %s, built: %s, go: %s)",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion)
}

// Short returns a short version string
func (i Info) Short() string {
	return fmt.Sprintf("wappd version %s", i.Version)
}
