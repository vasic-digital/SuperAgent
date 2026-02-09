package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentationSync_Initialization tests initialization
func TestDocumentationSync_Initialization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	ds := NewDocumentationSync(logger)

	assert.NotNil(t, ds)
	assert.NotNil(t, ds.logger)
}

// TestDocumentationSync_GenerateConstitutionSection tests section generation
func TestDocumentationSync_GenerateConstitutionSection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Summary:     "Test summary",
		Rules: []ConstitutionRule{
			{
				ID:          "TEST-001",
				Category:    "Testing",
				Title:       "Test Coverage",
				Description: "100% test coverage required",
				Mandatory:   true,
				Priority:    1,
				AddedAt:     time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "TEST-002",
				Category:    "Performance",
				Title:       "Optimization",
				Description: "Optimize for speed",
				Mandatory:   false,
				Priority:    3,
				AddedAt:     time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	section := ds.generateConstitutionSection(constitution)

	assert.NotEmpty(t, section)
	assert.Contains(t, section, "# Project Constitution")
	assert.Contains(t, section, "Version:")
	assert.Contains(t, section, "1.0.0")
	assert.Contains(t, section, "## Mandatory Principles")
	assert.Contains(t, section, "Test Coverage")
	assert.Contains(t, section, "## Context-Specific Guidelines")
	assert.Contains(t, section, "Optimization")
}

// TestDocumentationSync_WrapSection tests section wrapping
func TestDocumentationSync_WrapSection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	content := "Test content"
	wrapped := ds.wrapSection("TestSection", content)

	assert.Contains(t, wrapped, "<!-- BEGIN_TESTSECTION -->")
	assert.Contains(t, wrapped, "Test content")
	assert.Contains(t, wrapped, "<!-- END_TESTSECTION -->")
}

// TestDocumentationSync_SyncToFile_CreateNew tests creating new file with section
func TestDocumentationSync_SyncToFile_CreateNew(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "TEST.md")

	content := "Test Constitution content"
	err := ds.syncToFile(filePath, "Constitution", content)
	require.NoError(t, err)

	// Verify file created
	assert.FileExists(t, filePath)

	// Verify content
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(fileContent), "<!-- BEGIN_CONSTITUTION -->")
	assert.Contains(t, string(fileContent), content)
	assert.Contains(t, string(fileContent), "<!-- END_CONSTITUTION -->")
}

// TestDocumentationSync_SyncToFile_UpdateExisting tests updating existing section
func TestDocumentationSync_SyncToFile_UpdateExisting(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "TEST.md")

	// Create initial file with section
	initialContent := `# Test File

<!-- BEGIN_CONSTITUTION -->
Old constitution content
<!-- END_CONSTITUTION -->

Some other content`

	err := os.WriteFile(filePath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Update section
	newContent := "New constitution content"
	err = ds.syncToFile(filePath, "Constitution", newContent)
	require.NoError(t, err)

	// Verify updated content
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	contentStr := string(fileContent)

	assert.Contains(t, contentStr, "<!-- BEGIN_CONSTITUTION -->")
	assert.Contains(t, contentStr, newContent)
	assert.Contains(t, contentStr, "<!-- END_CONSTITUTION -->")
	assert.NotContains(t, contentStr, "Old constitution content")
	assert.Contains(t, contentStr, "Some other content")
}

// TestDocumentationSync_SyncToFile_AppendToExisting tests appending to file without section
func TestDocumentationSync_SyncToFile_AppendToExisting(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "TEST.md")

	// Create file without Constitution section
	initialContent := `# Test File

Existing content without Constitution section`

	err := os.WriteFile(filePath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Add Constitution section
	content := "Constitution content"
	err = ds.syncToFile(filePath, "Constitution", content)
	require.NoError(t, err)

	// Verify appended content
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	contentStr := string(fileContent)

	assert.Contains(t, contentStr, "Existing content")
	assert.Contains(t, contentStr, "<!-- BEGIN_CONSTITUTION -->")
	assert.Contains(t, contentStr, content)
	assert.Contains(t, contentStr, "<!-- END_CONSTITUTION -->")
}

// TestDocumentationSync_ExtractSectionFromFile tests section extraction
func TestDocumentationSync_ExtractSectionFromFile(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "TEST.md")

	// Create file with section
	fileContent := `# Test File

<!-- BEGIN_CONSTITUTION -->
Extracted constitution content
<!-- END_CONSTITUTION -->

Other content`

	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err)

	// Extract section
	extracted, err := ds.extractSectionFromFile(filePath, "Constitution")
	require.NoError(t, err)
	assert.Contains(t, extracted, "Extracted constitution content")
	assert.NotContains(t, extracted, "BEGIN_CONSTITUTION")
	assert.NotContains(t, extracted, "END_CONSTITUTION")
}

// TestDocumentationSync_ExtractSectionFromFile_NoSection tests extraction when section doesn't exist
func TestDocumentationSync_ExtractSectionFromFile_NoSection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "TEST.md")

	// Create file without section
	fileContent := `# Test File

No Constitution section here`

	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err)

	// Try to extract section
	_, err = ds.extractSectionFromFile(filePath, "Constitution")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section not found")
}

// TestDocumentationSync_SyncConstitutionToDocumentation tests full sync workflow
func TestDocumentationSync_SyncConstitutionToDocumentation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	projectRoot := t.TempDir()

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Summary:     "Test Constitution",
		Rules: []ConstitutionRule{
			{
				ID:          "TEST-001",
				Category:    "Testing",
				Title:       "Test Rule",
				Description: "Test description",
				Mandatory:   true,
				Priority:    1,
				AddedAt:     time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	err := ds.SyncConstitutionToDocumentation(projectRoot, constitution)
	require.NoError(t, err)

	// Verify files created
	assert.FileExists(t, filepath.Join(projectRoot, "AGENTS.md"))
	assert.FileExists(t, filepath.Join(projectRoot, "CLAUDE.md"))
	assert.FileExists(t, filepath.Join(projectRoot, "CONSTITUTION.md"))

	// Verify content in AGENTS.md
	agentsContent, err := os.ReadFile(filepath.Join(projectRoot, "AGENTS.md"))
	require.NoError(t, err)
	assert.Contains(t, string(agentsContent), "<!-- BEGIN_CONSTITUTION -->")
	assert.Contains(t, string(agentsContent), "Test Rule")
	assert.Contains(t, string(agentsContent), "<!-- END_CONSTITUTION -->")

	// Verify content in CLAUDE.md
	claudeContent, err := os.ReadFile(filepath.Join(projectRoot, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Contains(t, string(claudeContent), "<!-- BEGIN_CONSTITUTION -->")
	assert.Contains(t, string(claudeContent), "Test Rule")
	assert.Contains(t, string(claudeContent), "<!-- END_CONSTITUTION -->")

	// Verify CONSTITUTION.md
	constitutionMD, err := os.ReadFile(filepath.Join(projectRoot, "CONSTITUTION.md"))
	require.NoError(t, err)
	assert.Contains(t, string(constitutionMD), "# TestProject Constitution")
	assert.Contains(t, string(constitutionMD), "Test Rule")
}

// TestDocumentationSync_ValidateSync tests validation of sync status
func TestDocumentationSync_ValidateSync(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	projectRoot := t.TempDir()

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Summary:     "Test Constitution",
		Rules: []ConstitutionRule{
			{
				ID:          "TEST-001",
				Category:    "Testing",
				Title:       "Test Rule",
				Description: "Test description",
				Mandatory:   true,
				Priority:    1,
				AddedAt:     time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	// Without syncing, should have issues
	issues := ds.ValidateSync(projectRoot, constitution)
	assert.NotEmpty(t, issues, "Should have issues before syncing")

	// After syncing, should have no issues
	err := ds.SyncConstitutionToDocumentation(projectRoot, constitution)
	require.NoError(t, err)

	issues = ds.ValidateSync(projectRoot, constitution)
	assert.Empty(t, issues, "Should have no issues after syncing")
}

// TestDocumentationSync_ExtractConstitutionFromDocumentation tests extraction from docs
func TestDocumentationSync_ExtractConstitutionFromDocumentation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	projectRoot := t.TempDir()

	// Create AGENTS.md with Constitution section
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	agentsContent := `# AGENTS

<!-- BEGIN_CONSTITUTION -->
Constitution content from AGENTS.md
<!-- END_CONSTITUTION -->
`

	err := os.WriteFile(agentsPath, []byte(agentsContent), 0644)
	require.NoError(t, err)

	// Extract Constitution
	extracted, err := ds.ExtractConstitutionFromDocumentation(projectRoot)
	require.NoError(t, err)
	assert.Contains(t, extracted, "Constitution content from AGENTS.md")
}

// TestDocumentationSync_ExtractConstitutionFromDocumentation_NoFiles tests extraction when files missing
func TestDocumentationSync_ExtractConstitutionFromDocumentation_NoFiles(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	projectRoot := t.TempDir()

	// No files exist
	_, err := ds.ExtractConstitutionFromDocumentation(projectRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Constitution content found")
}

// TestDocumentationSync_GenerateConstitutionReport tests report generation
func TestDocumentationSync_GenerateConstitutionReport(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	projectRoot := t.TempDir()

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Summary:     "Test Constitution",
		Rules: []ConstitutionRule{
			{
				ID:          "TEST-001",
				Category:    "Testing",
				Title:       "Test Rule 1",
				Description: "Test description 1",
				Mandatory:   true,
				Priority:    1,
				AddedAt:     time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "TEST-002",
				Category:    "Testing",
				Title:       "Test Rule 2",
				Description: "Test description 2",
				Mandatory:   false,
				Priority:    2,
				AddedAt:     time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	// Sync first
	err := ds.SyncConstitutionToDocumentation(projectRoot, constitution)
	require.NoError(t, err)

	// Generate report
	report := ds.GenerateConstitutionReport(projectRoot, constitution)

	assert.NotEmpty(t, report)
	assert.Contains(t, report, "# Constitution Synchronization Report")
	assert.Contains(t, report, "Generated:")
	assert.Contains(t, report, "## Constitution Summary")
	assert.Contains(t, report, "**Total Rules:** 2")
	assert.Contains(t, report, "**Mandatory Rules:** 1")
	assert.Contains(t, report, "**Context Rules:** 1")
	assert.Contains(t, report, "## Synchronization Status")
	assert.Contains(t, report, "## Documentation Files")
	assert.Contains(t, report, "CONSTITUTION.json")
	assert.Contains(t, report, "CONSTITUTION.md")
	assert.Contains(t, report, "AGENTS.md")
	assert.Contains(t, report, "CLAUDE.md")
	assert.Contains(t, report, "## Rules by Category")
	assert.Contains(t, report, "Testing")
}

// TestDocumentationSync_SyncFromDocumentation tests syncing from documentation
func TestDocumentationSync_SyncFromDocumentation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ds := NewDocumentationSync(logger)

	projectRoot := t.TempDir()

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Summary:     "Test Constitution",
		Rules:       []ConstitutionRule{},
	}

	// Create AGENTS.md with Constitution section
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	agentsContent := `# AGENTS

<!-- BEGIN_CONSTITUTION -->
Updated Constitution from documentation
<!-- END_CONSTITUTION -->
`

	err := os.WriteFile(agentsPath, []byte(agentsContent), 0644)
	require.NoError(t, err)

	// Sync from documentation
	err = ds.SyncFromDocumentation(projectRoot, constitution)
	require.NoError(t, err)
	// Note: Current implementation logs warning for manual review
	// Full implementation would parse and update constitution.Rules
}
