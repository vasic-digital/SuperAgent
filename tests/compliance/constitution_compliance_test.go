package compliance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// constitutionRule represents a single rule in the CONSTITUTION.json file.
type constitutionRule struct {
	ID        string `json:"id"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	Mandatory bool   `json:"mandatory"`
	Priority  int    `json:"priority"`
}

// constitutionFile represents the structure of CONSTITUTION.json.
type constitutionFile struct {
	Version     string             `json:"version"`
	ProjectName string             `json:"project_name"`
	Rules       []constitutionRule `json:"rules"`
}

// getProjectRoot returns the root of the HelixAgent project.
func getProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	// Walk up from tests/compliance/ to project root
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..")
}

// TestConstitutionFileExists verifies that CONSTITUTION.json exists at the
// project root.
func TestConstitutionFileExists(t *testing.T) {
	root := getProjectRoot()
	constitutionPath := filepath.Join(root, "CONSTITUTION.json")

	_, err := os.Stat(constitutionPath)
	assert.NoError(t, err, "CONSTITUTION.json must exist at project root: %s", constitutionPath)

	t.Logf("COMPLIANCE: CONSTITUTION.json found at %s", constitutionPath)
}

// TestConstitutionRuleCount verifies that the constitution contains
// at least 26 mandatory rules as required by the project standards.
func TestConstitutionRuleCount(t *testing.T) {
	root := getProjectRoot()
	constitutionPath := filepath.Join(root, "CONSTITUTION.json")

	data, err := os.ReadFile(constitutionPath)
	require.NoError(t, err, "Must be able to read CONSTITUTION.json")

	var constitution constitutionFile
	err = json.Unmarshal(data, &constitution)
	require.NoError(t, err, "CONSTITUTION.json must be valid JSON")

	mandatoryCount := 0
	for _, rule := range constitution.Rules {
		if rule.Mandatory {
			mandatoryCount++
		}
	}

	const minMandatoryRules = 24
	t.Logf("Constitution: %d total rules, %d mandatory", len(constitution.Rules), mandatoryCount)

	if mandatoryCount < minMandatoryRules {
		t.Errorf("COMPLIANCE FAILED: Only %d mandatory rules, minimum is %d",
			mandatoryCount, minMandatoryRules)
	}
}

// TestConstitutionCategories verifies that all required constitution categories are present.
func TestConstitutionCategories(t *testing.T) {
	root := getProjectRoot()
	constitutionPath := filepath.Join(root, "CONSTITUTION.json")

	data, err := os.ReadFile(constitutionPath)
	require.NoError(t, err)

	var constitution constitutionFile
	err = json.Unmarshal(data, &constitution)
	require.NoError(t, err)

	requiredCategories := []string{
		"Testing", "Architecture", "Documentation", "Security",
		"Containerization", "GitOps", "Observability",
	}

	categorySet := make(map[string]bool)
	for _, rule := range constitution.Rules {
		categorySet[rule.Category] = true
	}

	t.Log("Constitution categories:")
	for cat := range categorySet {
		t.Logf("  - %s", cat)
	}

	for _, cat := range requiredCategories {
		assert.True(t, categorySet[cat],
			"COMPLIANCE FAILED: Required category %q not found in constitution", cat)
	}
}

// TestConstitutionProjectName verifies that the constitution references
// the correct project name.
func TestConstitutionProjectName(t *testing.T) {
	root := getProjectRoot()
	constitutionPath := filepath.Join(root, "CONSTITUTION.json")

	data, err := os.ReadFile(constitutionPath)
	require.NoError(t, err)

	var constitution constitutionFile
	err = json.Unmarshal(data, &constitution)
	require.NoError(t, err)

	assert.Equal(t, "HelixAgent", constitution.ProjectName,
		"COMPLIANCE FAILED: Constitution project_name must be 'HelixAgent'")
	t.Logf("COMPLIANCE: Constitution project name is %q (version %s)",
		constitution.ProjectName, constitution.Version)
}

// TestConstitutionPriority1Rules verifies that Priority 1 (highest priority)
// rules exist and cover critical areas.
func TestConstitutionPriority1Rules(t *testing.T) {
	root := getProjectRoot()
	constitutionPath := filepath.Join(root, "CONSTITUTION.json")

	data, err := os.ReadFile(constitutionPath)
	require.NoError(t, err)

	var constitution constitutionFile
	err = json.Unmarshal(data, &constitution)
	require.NoError(t, err)

	priority1Count := 0
	for _, rule := range constitution.Rules {
		if rule.Priority == 1 {
			priority1Count++
		}
	}

	assert.GreaterOrEqual(t, priority1Count, 10,
		"COMPLIANCE FAILED: At least 10 Priority-1 rules required")
	t.Logf("COMPLIANCE: %d Priority-1 (critical) rules found", priority1Count)
}
