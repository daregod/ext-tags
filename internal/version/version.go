// Package version holds build metadata injected at compile time.
package version

import "fmt"

// These variables are overridden by the build:
//
//	go build -ldflags "-X external-tags/internal/version.Version=v1.2.3 ..."
var (
	Version = "0.0.0-dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a human-readable build stamp.
func String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}
