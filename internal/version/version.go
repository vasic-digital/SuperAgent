package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// Variables set via -ldflags -X at build time.
var (
	Version     = "dev"
	VersionCode = "0"
	GitCommit   = "unknown"
	GitBranch   = "unknown"
	BuildDate   = "unknown"
	SourceHash  = "unknown"
	Builder     = "local"
)

// Info holds complete build version information.
type Info struct {
	Version     string `json:"version"`
	VersionCode string `json:"version_code"`
	GitCommit   string `json:"git_commit"`
	GitBranch   string `json:"git_branch"`
	BuildDate   string `json:"build_date"`
	SourceHash  string `json:"source_hash"`
	Builder     string `json:"builder"`
	GoVersion   string `json:"go_version"`
	Platform    string `json:"platform"`
}

// Get returns the current build version information.
func Get() Info {
	return Info{
		Version:     Version,
		VersionCode: VersionCode,
		GitCommit:   GitCommit,
		GitBranch:   GitBranch,
		BuildDate:   BuildDate,
		SourceHash:  SourceHash,
		Builder:     Builder,
		GoVersion:   runtime.Version(),
		Platform:    runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// Short returns a one-line version summary.
func Short() string {
	return fmt.Sprintf(
		"HelixAgent v%s (build %s, commit %s, %s)",
		Version, VersionCode, GitCommit, BuildDate,
	)
}

// String returns a multi-line version description.
func (i Info) String() string {
	return fmt.Sprintf(
		"HelixAgent v%s\n"+
			"Version Code: %s\n"+
			"Git Commit:   %s\n"+
			"Git Branch:   %s\n"+
			"Build Date:   %s\n"+
			"Source Hash:  %s\n"+
			"Builder:      %s\n"+
			"Go Version:   %s\n"+
			"Platform:     %s",
		i.Version, i.VersionCode, i.GitCommit, i.GitBranch,
		i.BuildDate, i.SourceHash, i.Builder, i.GoVersion, i.Platform,
	)
}

// JSON returns the version info as a JSON string.
func (i Info) JSON() string {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data)
}
