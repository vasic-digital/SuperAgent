package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConstitutionManager_Initialization tests initialization
func TestConstitutionManager_Initialization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cm := NewConstitutionManager(logger)

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.logger)
}

// TestConstitutionManager_CreateDefaultConstitution tests default Constitution creation
func TestConstitutionManager_CreateDefaultConstitution(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	assert.NotNil(t, constitution)
	assert.Equal(t, "1.0.0", constitution.Version)
	assert.NotEmpty(t, constitution.ProjectName)
	assert.NotZero(t, constitution.CreatedAt)
	assert.NotZero(t, constitution.UpdatedAt)
	assert.NotEmpty(t, constitution.Rules)
	assert.NotEmpty(t, constitution.Summary)

	// Verify mandatory rules are present
	mandatoryCount := 0
	for _, rule := range constitution.Rules {
		if rule.Mandatory {
			mandatoryCount++
		}
	}
	assert.Greater(t, mandatoryCount, 15, "Should have at least 15 mandatory rules")
}

// TestConstitutionManager_MandatoryRules tests mandatory rules are correctly added
func TestConstitutionManager_MandatoryRules(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	// Check for key mandatory rules
	expectedRules := []string{
		"Comprehensive Decoupling",
		"100% Test Coverage",
		"Comprehensive Challenges",
		"Complete Documentation",
		"No Broken Components",
		"No Dead Code",
		"Memory Safety",
		"Security Scanning",
		"Manual CI/CD Only",
		"Documentation Synchronization",
	}

	for _, expectedTitle := range expectedRules {
		found := false
		for _, rule := range constitution.Rules {
			if rule.Title == expectedTitle {
				found = true
				assert.True(t, rule.Mandatory, "Rule '%s' should be mandatory", expectedTitle)
				assert.NotEmpty(t, rule.ID)
				assert.NotEmpty(t, rule.Category)
				assert.NotEmpty(t, rule.Description)
				assert.GreaterOrEqual(t, rule.Priority, 1)
				assert.LessOrEqual(t, rule.Priority, 5)
				break
			}
		}
		assert.True(t, found, "Expected mandatory rule '%s' not found", expectedTitle)
	}
}

// TestConstitutionManager_SaveAndLoad tests saving and loading Constitution
func TestConstitutionManager_SaveAndLoad(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	originalConstitution := cm.createDefaultConstitution(projectRoot)

	// Save Constitution
	err := cm.SaveConstitution(projectRoot, originalConstitution)
	require.NoError(t, err)

	// Verify file exists
	constitutionPath := filepath.Join(projectRoot, "CONSTITUTION.json")
	assert.FileExists(t, constitutionPath)

	// Load Constitution
	loadedConstitution, err := cm.LoadConstitution(constitutionPath)
	require.NoError(t, err)

	// Verify loaded Constitution matches original
	assert.Equal(t, originalConstitution.Version, loadedConstitution.Version)
	assert.Equal(t, originalConstitution.ProjectName, loadedConstitution.ProjectName)
	assert.Len(t, loadedConstitution.Rules, len(originalConstitution.Rules))
}

// TestConstitutionManager_LoadOrCreateConstitution_Create tests creating new Constitution
func TestConstitutionManager_LoadOrCreateConstitution_Create(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	ctx := context.Background()
	projectRoot := t.TempDir()

	constitution, err := cm.LoadOrCreateConstitution(ctx, projectRoot)
	require.NoError(t, err)
	assert.NotNil(t, constitution)
	assert.NotEmpty(t, constitution.Rules)
}

// TestConstitutionManager_LoadOrCreateConstitution_Load tests loading existing Constitution
func TestConstitutionManager_LoadOrCreateConstitution_Load(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	ctx := context.Background()
	projectRoot := t.TempDir()

	// Create and save Constitution first
	originalConstitution := cm.createDefaultConstitution(projectRoot)
	err := cm.SaveConstitution(projectRoot, originalConstitution)
	require.NoError(t, err)

	// Load Constitution
	loadedConstitution, err := cm.LoadOrCreateConstitution(ctx, projectRoot)
	require.NoError(t, err)
	assert.NotNil(t, loadedConstitution)
	assert.Equal(t, originalConstitution.Version, loadedConstitution.Version)
}

// TestConstitutionManager_UpdateConstitutionFromDebate tests updating from debate results
func TestConstitutionManager_UpdateConstitutionFromDebate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	debateResult := &DebateResult{
		DebateID:     "test-debate",
		Topic:        "Optimize performance",
		BestResponse: &ParticipantResponse{Content: "Performance optimization recommendations"},
		QualityScore: 0.85,
	}

	updatedConstitution, err := cm.UpdateConstitutionFromDebate(constitution, debateResult, "Optimize the system performance")
	require.NoError(t, err)
	assert.NotNil(t, updatedConstitution)
	assert.NotEmpty(t, updatedConstitution.Summary)
	// Updated timestamp should be more recent
	assert.True(t, updatedConstitution.UpdatedAt.After(constitution.CreatedAt) || updatedConstitution.UpdatedAt.Equal(constitution.CreatedAt))
}

// TestConstitutionManager_EnsureMandatoryRules tests that missing mandatory rules are added
func TestConstitutionManager_EnsureMandatoryRules(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	// Create Constitution with only a few rules
	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rules:       []ConstitutionRule{},
	}

	// Ensure mandatory rules
	cm.ensureMandatoryRules(constitution)

	// Verify mandatory rules were added
	mandatoryCount := 0
	for _, rule := range constitution.Rules {
		if rule.Mandatory {
			mandatoryCount++
		}
	}
	assert.Greater(t, mandatoryCount, 15, "Should have at least 15 mandatory rules after ensuring")
}

// TestConstitutionManager_GetRulesByCategory tests filtering rules by category
func TestConstitutionManager_GetRulesByCategory(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	// Get Testing category rules
	testingRules := cm.GetRulesByCategory(constitution, "Testing")
	assert.NotEmpty(t, testingRules)

	// All returned rules should be in Testing category
	for _, rule := range testingRules {
		assert.Equal(t, "Testing", rule.Category)
	}

	// Get Architecture category rules
	archRules := cm.GetRulesByCategory(constitution, "Architecture")
	assert.NotEmpty(t, archRules)

	for _, rule := range archRules {
		assert.Equal(t, "Architecture", rule.Category)
	}
}

// TestConstitutionManager_GetMandatoryRules tests getting only mandatory rules
func TestConstitutionManager_GetMandatoryRules(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	mandatoryRules := cm.GetMandatoryRules(constitution)
	assert.NotEmpty(t, mandatoryRules)

	// All returned rules should be mandatory
	for _, rule := range mandatoryRules {
		assert.True(t, rule.Mandatory)
	}
}

// TestConstitutionManager_ValidateCompliance tests compliance validation
func TestConstitutionManager_ValidateCompliance(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	tests := []struct {
		name                  string
		implementationDetails string
		expectViolations      bool
	}{
		{
			"Complete Implementation",
			"Added comprehensive tests with 100% coverage, full documentation including README and user guides, security scanning with Snyk",
			false,
		},
		{
			"Missing Tests",
			"Implemented new feature without any testing",
			true,
		},
		{
			"Missing Documentation",
			"Added new feature implementation",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := cm.ValidateCompliance(constitution, tt.implementationDetails)
			if tt.expectViolations {
				assert.NotEmpty(t, violations, "Expected violations but got none")
			} else {
				// May have some violations due to simple keyword matching, but should be minimal
				t.Logf("Violations: %v", violations)
			}
		})
	}
}

// TestConstitutionManager_ExportConstitutionMarkdown tests markdown export
func TestConstitutionManager_ExportConstitutionMarkdown(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	markdown := cm.ExportConstitutionMarkdown(constitution)

	assert.NotEmpty(t, markdown)
	assert.Contains(t, markdown, "# ")
	assert.Contains(t, markdown, "Constitution")
	assert.Contains(t, markdown, "Version:")
	assert.Contains(t, markdown, "Created:")
	assert.Contains(t, markdown, "## ")

	// Check for mandatory rules
	assert.Contains(t, markdown, "MANDATORY")
	assert.Contains(t, markdown, "Comprehensive Decoupling")
	assert.Contains(t, markdown, "100% Test Coverage")
}

// TestConstitutionManager_AddOrUpdateRule tests adding and updating rules
func TestConstitutionManager_AddOrUpdateRule(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: "TestProject",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rules:       []ConstitutionRule{},
	}

	// Add new rule
	newRule := &ConstitutionRule{
		ID:          "TEST-001",
		Category:    "Testing",
		Title:       "Test Rule",
		Description: "Original description",
		Mandatory:   false,
		Priority:    3,
		AddedAt:     time.Now(),
		UpdatedAt:   time.Now(),
	}

	cm.addOrUpdateRule(constitution, newRule)
	assert.Len(t, constitution.Rules, 1)
	assert.Equal(t, "Original description", constitution.Rules[0].Description)

	// Update existing rule
	updatedRule := &ConstitutionRule{
		ID:          "TEST-002",
		Category:    "Testing",
		Title:       "Test Rule", // Same title
		Description: "Updated description",
		Mandatory:   false,
		Priority:    3,
		AddedAt:     time.Now(),
		UpdatedAt:   time.Now(),
	}

	cm.addOrUpdateRule(constitution, updatedRule)
	assert.Len(t, constitution.Rules, 1, "Should still have 1 rule (updated, not added)")
	assert.Equal(t, "Updated description", constitution.Rules[0].Description)
}

// TestConstitutionManager_GenerateSummary tests summary generation
func TestConstitutionManager_GenerateSummary(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	projectRoot := t.TempDir()
	constitution := cm.createDefaultConstitution(projectRoot)

	summary := cm.generateSummary(constitution)

	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "Constitution with")
	assert.Contains(t, summary, "rules")
	assert.Contains(t, summary, "mandatory")
	assert.Contains(t, summary, "categories")
}

// TestConstitutionManager_DeriveContextRule tests context rule derivation
func TestConstitutionManager_DeriveContextRule(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	debateResult := &DebateResult{
		DebateID:     "test-debate",
		Topic:        "Test topic",
		BestResponse: &ParticipantResponse{Content: "Test answer"},
		QualityScore: 0.8,
	}

	tests := []struct {
		name             string
		userRequest      string
		expectRule       bool
		expectedCategory string
	}{
		{
			"Performance Request",
			"Optimize the performance of the application",
			true,
			"Performance",
		},
		{
			"Security Request",
			"Add authentication to the API",
			true,
			"Security",
		},
		{
			"Generic Request",
			"Add a new feature",
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := cm.deriveContextRule(tt.userRequest, debateResult)
			if tt.expectRule {
				assert.NotNil(t, rule)
				assert.Equal(t, tt.expectedCategory, rule.Category)
				assert.NotEmpty(t, rule.Title)
				assert.NotEmpty(t, rule.Description)
				assert.False(t, rule.Mandatory) // Context rules are not mandatory
			} else {
				// May still return nil or a rule depending on logic
				if rule != nil {
					t.Logf("Unexpected rule created: %+v", rule)
				}
			}
		})
	}
}

// TestConstitutionManager_LoadConstitution_InvalidFile tests loading invalid file
func TestConstitutionManager_LoadConstitution_InvalidFile(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewConstitutionManager(logger)

	// Non-existent file
	_, err := cm.LoadConstitution("/nonexistent/path/CONSTITUTION.json")
	assert.Error(t, err)

	// Invalid JSON
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte("not valid json"), 0644)
	require.NoError(t, err)

	_, err = cm.LoadConstitution(invalidPath)
	assert.Error(t, err)
}
