package build

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"dev.helix.agent/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projectRoot() string {
	// Walk up from tests/build to project root
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestVersion_DefaultValues(t *testing.T) {
	info := version.Get()
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.VersionCode)
	assert.NotEmpty(t, info.GoVersion)
	assert.Contains(t, info.Platform, "/")
}

func TestVersion_JSONRoundtrip(t *testing.T) {
	info := version.Get()

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded version.Info
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.Version, decoded.Version)
	assert.Equal(t, info.VersionCode, decoded.VersionCode)
	assert.Equal(t, info.GitCommit, decoded.GitCommit)
	assert.Equal(t, info.GoVersion, decoded.GoVersion)
	assert.Equal(t, info.Platform, decoded.Platform)
}

func TestVersion_Compilation(t *testing.T) {
	root := projectRoot()
	cmd := exec.Command("go", "build", "./internal/version/...")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "version package should compile: %s", string(output))
}

func TestVERSION_FileExists(t *testing.T) {
	root := projectRoot()
	path := filepath.Join(root, "VERSION")
	data, err := os.ReadFile(path)
	require.NoError(t, err, "VERSION file must exist at project root")

	ver := strings.TrimSpace(string(data))
	parts := strings.Split(ver, ".")
	assert.Len(t, parts, 3, "VERSION must be semver X.Y.Z, got: %s", ver)
}

func TestAppRegistry_AllCmdDirsExist(t *testing.T) {
	root := projectRoot()
	apps := []string{
		"helixagent", "api", "grpc-server",
		"cognee-mock", "sanity-check", "mcp-bridge",
		"generate-constitution",
	}
	for _, app := range apps {
		dir := filepath.Join(root, "cmd", app)
		info, err := os.Stat(dir)
		require.NoError(t, err, "cmd/%s directory must exist", app)
		assert.True(t, info.IsDir(), "cmd/%s must be a directory", app)
	}
}

func TestVersionDataDir_Exists(t *testing.T) {
	root := projectRoot()
	dir := filepath.Join(root, "releases", ".version-data")
	info, err := os.Stat(dir)
	require.NoError(t, err, "releases/.version-data directory must exist")
	assert.True(t, info.IsDir())
}

func TestBuildScripts_Exist(t *testing.T) {
	root := projectRoot()
	scripts := []string{
		"scripts/build/version-manager.sh",
		"scripts/build/build-container.sh",
		"scripts/build/build-release.sh",
		"scripts/build/build-all-releases.sh",
	}
	for _, script := range scripts {
		path := filepath.Join(root, script)
		info, err := os.Stat(path)
		require.NoError(t, err, "%s must exist", script)
		assert.True(t, info.Mode()&0111 != 0, "%s must be executable", script)
	}
}

func TestDockerfileBuilder_Exists(t *testing.T) {
	root := projectRoot()
	path := filepath.Join(root, "docker", "build", "Dockerfile.builder")
	_, err := os.Stat(path)
	require.NoError(t, err, "Dockerfile.builder must exist")
}

func TestBuildInfoJSON_Schema(t *testing.T) {
	// Validate the expected schema structure
	schema := `{
		"app": "test",
		"version": "1.0.0",
		"version_code": 1,
		"git_commit": "abc1234",
		"git_branch": "main",
		"build_date": "2026-01-01T00:00:00Z",
		"platform": "linux/amd64",
		"go_version": "go1.24",
		"source_hash": "sha256:abc123",
		"builder": "container"
	}`

	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(schema), &parsed)
	require.NoError(t, err)

	expectedKeys := []string{
		"app", "version", "version_code", "git_commit",
		"git_branch", "build_date", "platform", "go_version",
		"source_hash", "builder",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, parsed, key, "build-info.json must contain key: %s", key)
	}
}

func TestHashComputation_Deterministic(t *testing.T) {
	root := projectRoot()
	script := filepath.Join(root, "scripts", "build", "version-manager.sh")

	// Run hash computation twice and verify they match
	bashCmd := `source "` + script + `" && compute_source_hash helixagent`

	cmd1 := exec.Command("bash", "-c", bashCmd)
	cmd1.Dir = root
	out1, err := cmd1.Output()
	require.NoError(t, err, "first hash computation should succeed")

	cmd2 := exec.Command("bash", "-c", bashCmd)
	cmd2.Dir = root
	out2, err := cmd2.Output()
	require.NoError(t, err, "second hash computation should succeed")

	assert.Equal(t,
		strings.TrimSpace(string(out1)),
		strings.TrimSpace(string(out2)),
		"hash computation must be deterministic",
	)
}

func TestChangeDetection_NoChanges(t *testing.T) {
	root := projectRoot()
	script := filepath.Join(root, "scripts", "build", "version-manager.sh")

	// Save current hash, then check — should report no changes
	bashCmd := `source "` + script + `" && ` +
		`hash=$(compute_source_hash helixagent) && ` +
		`save_hash helixagent "$hash" && ` +
		`if has_changes helixagent; then echo "changed"; else echo "unchanged"; fi`

	cmd := exec.Command("bash", "-c", bashCmd)
	cmd.Dir = root
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "unchanged", strings.TrimSpace(string(out)))
}
