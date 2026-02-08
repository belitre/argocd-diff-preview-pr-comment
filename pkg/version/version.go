package version

import "fmt"

var (
	// Version is the semantic version of the application
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// Description is the short description of the application
	Description = "A CLI tool to process ArgoCD diffs and post them as PR comments"
)

// GetVersion returns the version string
func GetVersion() string {
	return Version
}

// GetCommit returns the commit hash
func GetCommit() string {
	return Commit
}

// GetDescription returns the application description
func GetDescription() string {
	return Description
}

// GetFullVersion returns a formatted string with version and commit
func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s)", Version, Commit)
}

// GetInfo returns all version information
func GetInfo() string {
	return fmt.Sprintf("Version: %s\nCommit: %s\nDescription: %s", Version, Commit, Description)
}
