// Package security contains security scanning infrastructure validation tests.
// These tests verify that Snyk and SonarQube scanning infrastructure is
// properly configured and that configuration files are synchronized.
package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot walks up from the current working directory to find go.mod,
// which marks the project root.
func getProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// =============================================================================
// SNYK CONFIGURATION TESTS
// =============================================================================

func TestSecurity_SnykConfigExists(t *testing.T) {
	root := getProjectRoot(t)
	_, err := os.Stat(filepath.Join(root, ".snyk"))
	assert.NoError(t, err, ".snyk configuration file must exist at project root")
}

func TestSecurity_SnykConfigHasLanguageSettings(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".snyk"))
	require.NoError(t, err, "must be able to read .snyk")
	assert.Contains(t, string(content), "language-settings",
		".snyk must contain language-settings section")
}

func TestSecurity_SnykConfigHasGoLanguage(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".snyk"))
	require.NoError(t, err, "must be able to read .snyk")
	assert.Contains(t, string(content), "go:",
		".snyk must include Go language configuration")
}

func TestSecurity_SnykConfigHasVersion(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".snyk"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "version:",
		".snyk must contain a version field")
}

func TestSecurity_SnykConfigHasIgnoreSection(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".snyk"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "ignore:",
		".snyk must have an ignore section for vulnerability exceptions")
}

// =============================================================================
// SNYK COMPOSE TESTS
// =============================================================================

func TestSecurity_SnykComposeExists(t *testing.T) {
	root := getProjectRoot(t)
	_, err := os.Stat(filepath.Join(root, "docker", "security", "snyk",
		"docker-compose.yml"))
	assert.NoError(t, err, "Snyk compose file must exist")
}

func TestSecurity_SnykDockerfileExists(t *testing.T) {
	root := getProjectRoot(t)
	_, err := os.Stat(filepath.Join(root, "docker", "security", "snyk",
		"Dockerfile"))
	assert.NoError(t, err, "Snyk Dockerfile must exist")
}

func TestSecurity_SnykComposeHasRequiredServices(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "snyk",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err, "must be able to read Snyk compose file")

	composeStr := string(content)
	requiredServices := []string{
		"snyk-deps", "snyk-code", "snyk-iac", "snyk-full",
	}
	for _, svc := range requiredServices {
		assert.Contains(t, composeStr, svc,
			"Snyk compose must define service: %s", svc)
	}
}

func TestSecurity_SnykComposeHasReportsVolume(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "snyk",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "snyk-reports",
		"Snyk compose must define a reports volume")
}

func TestSecurity_SnykComposeHasSecurityNetwork(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "snyk",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "security-network",
		"Snyk compose must reference the security network")
}

func TestSecurity_SnykComposeUsesReadOnlyMount(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "snyk",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), ":ro",
		"Snyk compose must mount app source as read-only")
}

func TestSecurity_SnykDockerfileHasScanScripts(t *testing.T) {
	root := getProjectRoot(t)
	dockerfilePath := filepath.Join(root, "docker", "security", "snyk",
		"Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	require.NoError(t, err)

	dockerfileStr := string(content)
	scripts := []string{
		"scan-dependencies.sh", "scan-code.sh",
		"scan-iac.sh", "scan-all.sh",
	}
	for _, script := range scripts {
		assert.Contains(t, dockerfileStr, script,
			"Snyk Dockerfile must create script: %s", script)
	}
}

func TestSecurity_SnykDockerfileUsesOfficialImage(t *testing.T) {
	root := getProjectRoot(t)
	dockerfilePath := filepath.Join(root, "docker", "security", "snyk",
		"Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "snyk/snyk-cli",
		"Snyk Dockerfile must use official snyk/snyk-cli base image")
}

// =============================================================================
// SONARQUBE CONFIGURATION TESTS
// =============================================================================

func TestSecurity_SonarQubeConfigExists(t *testing.T) {
	root := getProjectRoot(t)
	_, err := os.Stat(filepath.Join(root, "sonar-project.properties"))
	assert.NoError(t, err,
		"sonar-project.properties must exist at project root")
}

func TestSecurity_SonarQubeComposeExists(t *testing.T) {
	root := getProjectRoot(t)
	_, err := os.Stat(filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml"))
	assert.NoError(t, err, "SonarQube compose file must exist")
}

func TestSecurity_SonarQubeConfigHasProjectKey(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.projectKey=helixagent",
		"SonarQube config must have project key 'helixagent'")
}

func TestSecurity_SonarQubeConfigHasProjectName(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.projectName=HelixAgent",
		"SonarQube config must have project name 'HelixAgent'")
}

func TestSecurity_SonarQubeConfigHasVersion(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)

	version := extractSonarVersion(string(content))
	assert.NotEmpty(t, version,
		"SonarQube config must have a project version")
}

func TestSecurity_SonarQubeConfigHasCoverageReport(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.go.coverage.reportPaths",
		"SonarQube config must define Go coverage report path")
}

func TestSecurity_SonarQubeConfigHasTestReport(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.go.tests.reportPaths",
		"SonarQube config must define Go test report path")
}

func TestSecurity_SonarQubeConfigHasExclusions(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.exclusions",
		"SonarQube config must define source exclusions")
	assert.Contains(t, string(content), "vendor/**",
		"SonarQube exclusions must include vendor directory")
}

func TestSecurity_SonarQubeConfigHasQualityGate(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.qualitygate.wait",
		"SonarQube config must have quality gate wait setting")
}

func TestSecurity_SonarQubeConfigHasUTF8Encoding(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root,
		"sonar-project.properties"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sonar.sourceEncoding=UTF-8",
		"SonarQube config must set UTF-8 source encoding")
}

// =============================================================================
// SONARQUBE COMPOSE TESTS
// =============================================================================

func TestSecurity_SonarQubeComposeHasRequiredServices(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)

	composeStr := string(content)
	assert.Contains(t, composeStr, "sonarqube:",
		"SonarQube compose must define sonarqube service")
	assert.Contains(t, composeStr, "postgres:",
		"SonarQube compose must define postgres service")
	assert.Contains(t, composeStr, "sonar-scanner:",
		"SonarQube compose must define sonar-scanner service")
}

func TestSecurity_SonarQubeComposeHasHealthChecks(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)

	composeStr := string(content)
	assert.Contains(t, composeStr, "healthcheck:",
		"SonarQube compose must define health checks")
	assert.Contains(t, composeStr, "api/system/status",
		"SonarQube health check must use /api/system/status endpoint")
	assert.Contains(t, composeStr, "pg_isready",
		"PostgreSQL health check must use pg_isready")
}

func TestSecurity_SonarQubeComposeHasResourceLimits(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)

	composeStr := string(content)
	assert.Contains(t, composeStr, "mem_limit",
		"SonarQube compose must set memory limits")
	assert.Contains(t, composeStr, "cpus:",
		"SonarQube compose must set CPU limits")
}

func TestSecurity_SonarQubeComposeHasNetworkConfig(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)

	composeStr := string(content)
	assert.Contains(t, composeStr, "security-network",
		"SonarQube compose must define security-network")
	assert.Contains(t, composeStr, "driver: bridge",
		"Security network must use bridge driver")
}

func TestSecurity_SonarQubeComposeHasVolumes(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)

	composeStr := string(content)
	requiredVolumes := []string{
		"sonarqube_data:", "sonarqube_extensions:",
		"sonarqube_logs:", "postgres_data:",
	}
	for _, vol := range requiredVolumes {
		assert.Contains(t, composeStr, vol,
			"SonarQube compose must define volume: %s", vol)
	}
}

func TestSecurity_SonarQubeComposeUsesCommunityEdition(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "-community",
		"SonarQube compose must use community edition image")
}

func TestSecurity_SonarQubeComposeHasRestartPolicy(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "sonarqube",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "restart:",
		"SonarQube compose must have restart policy")
}

// =============================================================================
// CROSS-TOOL INTEGRATION TESTS
// =============================================================================

func TestSecurity_SecurityDirectoryHasBothTools(t *testing.T) {
	root := getProjectRoot(t)
	_, errSnyk := os.Stat(filepath.Join(root, "docker", "security", "snyk"))
	_, errSonar := os.Stat(filepath.Join(root, "docker", "security",
		"sonarqube"))
	assert.NoError(t, errSnyk,
		"docker/security/snyk directory must exist")
	assert.NoError(t, errSonar,
		"docker/security/sonarqube directory must exist")
}

func TestSecurity_MakefileHasSecurityScanTarget(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Makefile"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "security-scan:",
		"Makefile must have security-scan target")
}

func TestSecurity_MakefileHasSBOMTarget(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Makefile"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sbom:",
		"Makefile must have sbom target")
}

func TestSecurity_SnykNetworkReferencesSharedNetwork(t *testing.T) {
	root := getProjectRoot(t)
	composePath := filepath.Join(root, "docker", "security", "snyk",
		"docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "helixagent-sonarqube",
		"Snyk compose must reference SonarQube's shared security network")
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// extractSonarVersion extracts the sonar.projectVersion value from config.
func extractSonarVersion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "sonar.projectVersion=") {
			return strings.TrimPrefix(trimmed, "sonar.projectVersion=")
		}
	}
	return ""
}
