// Package main provides version information handling
package main

import (
	"os/exec"
	"strings"
	"time"
)

// getVersionInfo returns version information, using git if ldflags weren't provided
func getVersionInfo() (string, string, string, string) {
	// If version was set via ldflags, use those values
	if version != "dev" {
		return version, commit, date, builtBy
	}

	// Otherwise, determine from git
	gitVersion := getGitVersion()
	gitCommit := getGitCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)
	buildBy := "lizz"

	return gitVersion, gitCommit, buildDate, buildBy
}

// getGitVersion gets version from git describe
func getGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(output))
}

// getGitCommit gets current commit hash
func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))[:8] // Short hash
} 