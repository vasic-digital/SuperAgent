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

func TestConstitutionWatcher_Initialization(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	assert.NotNil(t, watcher, "Watcher should not be nil")
	assert.Equal(t, projectRoot, watcher.projectRoot, "Project root should match")
	assert.Equal(t, 5*time.Minute, watcher.checkInterval, "Check interval should be 5 minutes")
	assert.False(t, watcher.enabled, "Should be disabled by default")
}

func TestConstitutionWatcher_EnableDisable(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	assert.False(t, watcher.enabled, "Should start disabled")

	watcher.Enable()
	assert.True(t, watcher.enabled, "Should be enabled after Enable()")

	watcher.Disable()
	assert.False(t, watcher.enabled, "Should be disabled after Disable()")
}

func TestConstitutionWatcher_DetectNewModules(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	// Create a new module directory with go.mod
	moduleName := "TestModule"
	moduleDir := filepath.Join(projectRoot, moduleName)
	err := os.Mkdir(moduleDir, 0755)
	require.NoError(t, err, "Should create module directory")

	goModPath := filepath.Join(moduleDir, "go.mod")
	err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
	require.NoError(t, err, "Should create go.mod file")

	// Detect new modules
	newModules := watcher.detectNewModules()

	assert.Greater(t, len(newModules), 0, "Should detect new module")
	assert.Contains(t, newModules, moduleName, "Should include the test module")
}

func TestConstitutionWatcher_DetectDocumentationChanges(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	// Create AGENTS.md
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	err := os.WriteFile(agentsPath, []byte("# Agents\n"), 0644)
	require.NoError(t, err, "Should create AGENTS.md")

	// Set last mod time to past
	watcher.lastModTime = time.Now().Add(-1 * time.Hour)

	// Check for changes
	changed := watcher.detectDocumentationChanges()

	assert.True(t, changed, "Should detect documentation change")
}

func TestConstitutionWatcher_DetectStructureChanges(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	// Set last mod time to past
	watcher.lastModTime = time.Now().Add(-1 * time.Hour)

	// Create a new directory
	newDir := filepath.Join(projectRoot, "NewFeature")
	err := os.Mkdir(newDir, 0755)
	require.NoError(t, err, "Should create new directory")

	// Check for structure changes
	changed := watcher.detectStructureChanges()

	assert.True(t, changed, "Should detect structure change")
}

func TestConstitutionWatcher_UpdateForNewModule(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	// Create a default Constitution first
	constitution := constitutionManager.createDefaultConstitution("TestProject")
	err := constitutionManager.SaveConstitution(projectRoot, constitution)
	require.NoError(t, err, "Should save initial Constitution")

	// Create an update event
	event := &ConstitutionUpdateEvent{
		Trigger:     TriggerNewModule,
		Description: "TestModule extracted",
		Timestamp:   time.Now(),
	}

	// Apply update
	ctx := context.Background()
	err = watcher.updateForNewModule(ctx, event)
	require.NoError(t, err, "Should update Constitution for new module")

	// Load and verify Constitution was updated
	constitutionPath := filepath.Join(projectRoot, "CONSTITUTION.json")
	updatedConstitution, err := constitutionManager.LoadConstitution(constitutionPath)
	require.NoError(t, err, "Should load updated Constitution")

	// Check that a new rule was added
	assert.Greater(t, len(updatedConstitution.Rules), len(constitution.Rules), "Should have added a new rule")

	// Find the new module rule
	found := false
	for _, rule := range updatedConstitution.Rules {
		if rule.Title == "New Module Decoupling" {
			found = true
			assert.Contains(t, rule.Description, "TestModule extracted", "Rule should reference the module")
			break
		}
	}
	assert.True(t, found, "Should have added New Module Decoupling rule")
}

func TestConstitutionWatcher_SyncDocumentation(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	// Create Constitution and documentation files
	constitution := constitutionManager.createDefaultConstitution("TestProject")
	err := constitutionManager.SaveConstitution(projectRoot, constitution)
	require.NoError(t, err, "Should save Constitution")

	// Create placeholder documentation files
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	err = os.WriteFile(agentsPath, []byte("# Agents\n\n"), 0644)
	require.NoError(t, err, "Should create AGENTS.md")

	claudePath := filepath.Join(projectRoot, "CLAUDE.md")
	err = os.WriteFile(claudePath, []byte("# Claude\n\n"), 0644)
	require.NoError(t, err, "Should create CLAUDE.md")

	// Sync documentation
	ctx := context.Background()
	err = watcher.syncDocumentation(ctx)
	require.NoError(t, err, "Should sync documentation")

	// Verify Constitution section was added
	agentsContent, err := os.ReadFile(agentsPath)
	require.NoError(t, err, "Should read AGENTS.md")
	assert.Contains(t, string(agentsContent), "BEGIN_CONSTITUTION", "Should have Constitution section")

	claudeContent, err := os.ReadFile(claudePath)
	require.NoError(t, err, "Should read CLAUDE.md")
	assert.Contains(t, string(claudeContent), "BEGIN_CONSTITUTION", "Should have Constitution section")
}

func TestConstitutionWatcher_DetectChanges(t *testing.T) {
	logger := logrus.New()
	projectRoot := t.TempDir()

	constitutionManager := NewConstitutionManager(logger)
	documentationSync := NewDocumentationSync(logger)

	watcher := NewConstitutionWatcher(constitutionManager, documentationSync, logger, projectRoot)

	// Set last mod time to past
	watcher.lastModTime = time.Now().Add(-1 * time.Hour)

	// Create some changes
	// 1. New module
	moduleName := "NewModule"
	moduleDir := filepath.Join(projectRoot, moduleName)
	err := os.Mkdir(moduleDir, 0755)
	require.NoError(t, err, "Should create module directory")

	goModPath := filepath.Join(moduleDir, "go.mod")
	err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
	require.NoError(t, err, "Should create go.mod")

	// 2. Documentation change
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	err = os.WriteFile(agentsPath, []byte("# Agents\n"), 0644)
	require.NoError(t, err, "Should create AGENTS.md")

	// Detect changes
	events := watcher.detectChanges()

	assert.Greater(t, len(events), 0, "Should detect at least one change")

	// Check that we detected the new module
	foundModuleTrigger := false
	for _, event := range events {
		if event.Trigger == TriggerNewModule {
			foundModuleTrigger = true
			assert.Contains(t, event.Description, moduleName, "Should mention the new module")
		}
	}
	assert.True(t, foundModuleTrigger, "Should have detected new module trigger")
}
