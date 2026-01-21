package version

import (
	"strings"
	"testing"
)

func TestShort(t *testing.T) {
	// Save original values
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "default dev version",
			version: "dev",
			want:    "dev",
		},
		{
			name:    "semantic version",
			version: "v1.0.0",
			want:    "v1.0.0",
		},
		{
			name:    "git describe version",
			version: "v0.4.0-137-g4046f35",
			want:    "v0.4.0-137-g4046f35",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			got := Short()
			if got != tt.want {
				t.Errorf("Short() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	tests := []struct {
		name      string
		version   string
		commit    string
		buildDate string
	}{
		{
			name:      "default values",
			version:   "dev",
			commit:    "unknown",
			buildDate: "unknown",
		},
		{
			name:      "build values",
			version:   "v1.0.0",
			commit:    "abc123",
			buildDate: "2026-01-21T12:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			Commit = tt.commit
			BuildDate = tt.buildDate

			got := Info()

			// Verify the info string contains all expected components
			if !strings.Contains(got, "janitarr") {
				t.Errorf("Info() missing 'janitarr' prefix: %q", got)
			}
			if !strings.Contains(got, tt.version) {
				t.Errorf("Info() missing version %q: %q", tt.version, got)
			}
			if !strings.Contains(got, tt.commit) {
				t.Errorf("Info() missing commit %q: %q", tt.commit, got)
			}
			if !strings.Contains(got, tt.buildDate) {
				t.Errorf("Info() missing build date %q: %q", tt.buildDate, got)
			}
		})
	}
}
