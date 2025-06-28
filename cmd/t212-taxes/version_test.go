package main

import (
	"testing"
	"time"
)

func TestGetVersionInfo(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		commit         string
		date           string
		builtBy        string
		expectedBuiltBy string
	}{
		{
			name:           "ldflags provided",
			version:        "v1.0.0",
			commit:         "abc123",
			date:           "2024-01-01T00:00:00Z",
			builtBy:        "goreleaser",
			expectedBuiltBy: "goreleaser",
		},
		{
			name:           "dev version - should use git detection",
			version:        "dev",
			commit:         "unknown",
			date:           "unknown",
			builtBy:        "source",
			expectedBuiltBy: "lizz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := version
			origCommit := commit
			origDate := date
			origBuiltBy := builtBy

			// Set test values
			version = tt.version
			commit = tt.commit
			date = tt.date
			builtBy = tt.builtBy

			// Test version detection
			v, c, d, b := getVersionInfo()

			// Restore original values
			version = origVersion
			commit = origCommit
			date = origDate
			builtBy = origBuiltBy

			// Validate results
			if tt.version != "dev" {
				// When ldflags are provided, should return exact values
				if v != tt.version {
					t.Errorf("expected version %s, got %s", tt.version, v)
				}
				if c != tt.commit {
					t.Errorf("expected commit %s, got %s", tt.commit, c)
				}
				if d != tt.date {
					t.Errorf("expected date %s, got %s", tt.date, d)
				}
			} else {
				// When version is "dev", should use git detection
				if v == "dev" || v == "" {
					t.Errorf("expected git version detection, but got %s", v)
				}
				if c == "unknown" || c == "" {
					t.Errorf("expected git commit detection, but got %s", c)
				}
				// Date should be recent (within last minute)
				parsedDate, err := time.Parse(time.RFC3339, d)
				if err != nil {
					t.Errorf("expected valid RFC3339 date, got %s", d)
				}
				if time.Since(parsedDate) > time.Minute {
					t.Errorf("expected recent date, got %s", d)
				}
			}

			if b != tt.expectedBuiltBy {
				t.Errorf("expected builtBy %s, got %s", tt.expectedBuiltBy, b)
			}
		})
	}
}

func TestGetGitVersion(t *testing.T) {
	version := getGitVersion()
	
	// Should not be empty or "dev" (unless git is not available)
	if version == "" {
		t.Error("getGitVersion should not return empty string")
	}
	
	// In a git repository, should contain some version info
	// Could be a tag like "v1.0.0" or a commit hash
	if len(version) < 3 {
		t.Errorf("expected meaningful version, got %s", version)
	}
}

func TestGetGitCommit(t *testing.T) {
	commit := getGitCommit()
	
	// Should not be empty or "unknown" (unless git is not available)
	if commit == "" {
		t.Error("getGitCommit should not return empty string")
	}
	
	// Should be 8 characters (short hash)
	if commit != "unknown" && len(commit) != 8 {
		t.Errorf("expected 8-character commit hash, got %s (length %d)", commit, len(commit))
	}
	
	// Should only contain hex characters
	if commit != "unknown" {
		for _, char := range commit {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("commit hash should only contain hex characters, got %s", commit)
				break
			}
		}
	}
}

func TestVersionInfoIntegration(t *testing.T) {
	// Test that version info can be retrieved without panicking
	v, c, d, b := getVersionInfo()
	
	// Basic validation
	if v == "" {
		t.Error("version should not be empty")
	}
	if c == "" {
		t.Error("commit should not be empty")
	}
	if d == "" {
		t.Error("date should not be empty")
	}
	if b == "" {
		t.Error("builtBy should not be empty")
	}
	
	// Date should be valid RFC3339
	_, err := time.Parse(time.RFC3339, d)
	if err != nil {
		t.Errorf("date should be valid RFC3339 format, got %s: %v", d, err)
	}
	
	// Built by should be one of expected values
	expectedBuiltBy := []string{"lizz", "goreleaser", "github-actions", "source", "make"}
	found := false
	for _, expected := range expectedBuiltBy {
		if b == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("builtBy should be one of %v, got %s", expectedBuiltBy, b)
	}
} 