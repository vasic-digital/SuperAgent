package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ConstitutionWatcher monitors project changes and triggers Constitution updates
type ConstitutionWatcher struct {
	constitutionManager *ConstitutionManager
	documentationSync   *DocumentationSync
	logger              *logrus.Logger
	projectRoot         string
	checkInterval       time.Duration
	lastModTime         time.Time
	enabled             bool
}

// ConstitutionUpdateTrigger represents reasons to update Constitution
type ConstitutionUpdateTrigger string

const (
	TriggerNewModule         ConstitutionUpdateTrigger = "new_module_extracted"
	TriggerTestCoveragesDrop ConstitutionUpdateTrigger = "test_coverage_drop"
	TriggerDocumentationSync ConstitutionUpdateTrigger = "documentation_changed"
	TriggerProjectStructure  ConstitutionUpdateTrigger = "project_structure_changed"
)

// ConstitutionUpdateEvent represents a Constitution update event
type ConstitutionUpdateEvent struct {
	Trigger     ConstitutionUpdateTrigger `json:"trigger"`
	Description string                    `json:"description"`
	Timestamp   time.Time                 `json:"timestamp"`
	Applied     bool                      `json:"applied"`
	Error       string                    `json:"error,omitempty"`
}

// NewConstitutionWatcher creates a new Constitution watcher
func NewConstitutionWatcher(
	constitutionManager *ConstitutionManager,
	documentationSync *DocumentationSync,
	logger *logrus.Logger,
	projectRoot string,
) *ConstitutionWatcher {
	return &ConstitutionWatcher{
		constitutionManager: constitutionManager,
		documentationSync:   documentationSync,
		logger:              logger,
		projectRoot:         projectRoot,
		checkInterval:       5 * time.Minute, // Check every 5 minutes
		enabled:             false,            // Disabled by default
	}
}

// Enable enables the Constitution watcher
func (cw *ConstitutionWatcher) Enable() {
	cw.enabled = true
	cw.logger.Info("[Constitution Watcher] Enabled")
}

// Disable disables the Constitution watcher
func (cw *ConstitutionWatcher) Disable() {
	cw.enabled = false
	cw.logger.Info("[Constitution Watcher] Disabled")
}

// Start starts the Constitution watcher in the background
func (cw *ConstitutionWatcher) Start(ctx context.Context) {
	if !cw.enabled {
		cw.logger.Info("[Constitution Watcher] Not enabled, skipping")
		return
	}

	cw.logger.Info("[Constitution Watcher] Starting...")

	ticker := time.NewTicker(cw.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cw.logger.Info("[Constitution Watcher] Stopped")
			return
		case <-ticker.C:
			cw.checkForUpdates(ctx)
		}
	}
}

// checkForUpdates checks for project changes that should trigger Constitution updates
func (cw *ConstitutionWatcher) checkForUpdates(ctx context.Context) {
	events := cw.detectChanges()

	if len(events) == 0 {
		return
	}

	cw.logger.WithField("event_count", len(events)).Info("[Constitution Watcher] Detected changes requiring Constitution update")

	for _, event := range events {
		if err := cw.applyUpdate(ctx, event); err != nil {
			cw.logger.WithError(err).WithField("trigger", event.Trigger).Error("[Constitution Watcher] Failed to apply update")
			event.Applied = false
			event.Error = err.Error()
		} else {
			event.Applied = true
		}
	}
}

// detectChanges detects project changes that require Constitution updates
func (cw *ConstitutionWatcher) detectChanges() []*ConstitutionUpdateEvent {
	events := make([]*ConstitutionUpdateEvent, 0)

	// Check 1: New modules extracted (look for new go.mod files)
	if newModules := cw.detectNewModules(); len(newModules) > 0 {
		events = append(events, &ConstitutionUpdateEvent{
			Trigger:     TriggerNewModule,
			Description: fmt.Sprintf("Detected %d new modules: %v", len(newModules), newModules),
			Timestamp:   time.Now(),
		})
	}

	// Check 2: Documentation changes (AGENTS.md or CLAUDE.md modified)
	if docsChanged := cw.detectDocumentationChanges(); docsChanged {
		events = append(events, &ConstitutionUpdateEvent{
			Trigger:     TriggerDocumentationSync,
			Description: "Documentation files modified, sync may be needed",
			Timestamp:   time.Now(),
		})
	}

	// Check 3: Project structure changes (new directories in root)
	if structureChanged := cw.detectStructureChanges(); structureChanged {
		events = append(events, &ConstitutionUpdateEvent{
			Trigger:     TriggerProjectStructure,
			Description: "Project structure changed",
			Timestamp:   time.Now(),
		})
	}

	return events
}

// detectNewModules detects newly extracted modules
func (cw *ConstitutionWatcher) detectNewModules() []string {
	newModules := make([]string, 0)

	// Look for go.mod files that indicate extracted modules
	modulePatterns := []string{
		"*/go.mod",         // Top-level modules
		"*/*/go.mod",       // Nested modules
		"modules/*/go.mod", // Modules directory
	}

	for _, pattern := range modulePatterns {
		matches, err := filepath.Glob(filepath.Join(cw.projectRoot, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Skip root go.mod and vendor
			if match == filepath.Join(cw.projectRoot, "go.mod") {
				continue
			}
			if strings.Contains(match, "vendor") {
				continue
			}

			// Check if this is a new module (created after last check)
			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			if cw.lastModTime.IsZero() || info.ModTime().After(cw.lastModTime) {
				moduleName := filepath.Base(filepath.Dir(match))
				newModules = append(newModules, moduleName)
			}
		}
	}

	return newModules
}

// detectDocumentationChanges detects changes to documentation files
func (cw *ConstitutionWatcher) detectDocumentationChanges() bool {
	files := []string{
		filepath.Join(cw.projectRoot, "AGENTS.md"),
		filepath.Join(cw.projectRoot, "CLAUDE.md"),
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if !cw.lastModTime.IsZero() && info.ModTime().After(cw.lastModTime) {
			return true
		}
	}

	return false
}

// detectStructureChanges detects project structure changes
func (cw *ConstitutionWatcher) detectStructureChanges() bool {
	// Check if new top-level directories were added
	entries, err := os.ReadDir(cw.projectRoot)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip common directories
		name := entry.Name()
		if name == "." || name == "vendor" || name == "node_modules" || name == "bin" || name == ".git" {
			continue
		}

		dirPath := filepath.Join(cw.projectRoot, entry.Name())
		info, err := os.Stat(dirPath)
		if err != nil {
			continue
		}

		if !cw.lastModTime.IsZero() && info.ModTime().After(cw.lastModTime) {
			return true
		}
	}

	return false
}

// applyUpdate applies a Constitution update based on the trigger
func (cw *ConstitutionWatcher) applyUpdate(ctx context.Context, event *ConstitutionUpdateEvent) error {
	switch event.Trigger {
	case TriggerNewModule:
		return cw.updateForNewModule(ctx, event)
	case TriggerDocumentationSync:
		return cw.syncDocumentation(ctx)
	case TriggerProjectStructure:
		return cw.updateForStructureChange(ctx, event)
	case TriggerTestCoveragesDrop:
		return cw.flagCoverageViolation(ctx)
	default:
		return fmt.Errorf("unknown trigger: %s", event.Trigger)
	}
}

// updateForNewModule updates Constitution when new modules are extracted
func (cw *ConstitutionWatcher) updateForNewModule(ctx context.Context, event *ConstitutionUpdateEvent) error {
	cw.logger.Info("[Constitution Watcher] Updating for new module extraction")

	// Load current Constitution
	constitutionPath := filepath.Join(cw.projectRoot, "CONSTITUTION.json")
	constitution, err := cw.constitutionManager.LoadConstitution(constitutionPath)
	if err != nil {
		return fmt.Errorf("failed to load Constitution: %w", err)
	}

	// Add a context-specific rule for the new module
	newRule := ConstitutionRule{
		ID:          fmt.Sprintf("CONST-MOD-%d", time.Now().Unix()),
		Category:    "Architecture",
		Title:       "New Module Decoupling",
		Description: fmt.Sprintf("New modules detected: %s. Ensure proper decoupling and documentation.", event.Description),
		Mandatory:   false,
		Priority:    3,
		AddedAt:     time.Now(),
		UpdatedAt:   time.Now(),
	}

	constitution.Rules = append(constitution.Rules, newRule)
	constitution.UpdatedAt = time.Now()

	// Save updated Constitution
	if err := cw.constitutionManager.SaveConstitution(cw.projectRoot, constitution); err != nil {
		return fmt.Errorf("failed to save Constitution: %w", err)
	}

	// Sync to documentation
	if err := cw.documentationSync.SyncConstitutionToDocumentation(cw.projectRoot, constitution); err != nil {
		return fmt.Errorf("failed to sync documentation: %w", err)
	}

	cw.logger.Info("[Constitution Watcher] Constitution updated for new module")
	return nil
}

// syncDocumentation syncs Constitution across all documentation files
func (cw *ConstitutionWatcher) syncDocumentation(ctx context.Context) error {
	cw.logger.Info("[Constitution Watcher] Syncing documentation")

	constitutionPath := filepath.Join(cw.projectRoot, "CONSTITUTION.json")
	constitution, err := cw.constitutionManager.LoadConstitution(constitutionPath)
	if err != nil {
		return fmt.Errorf("failed to load Constitution: %w", err)
	}

	if err := cw.documentationSync.SyncConstitutionToDocumentation(cw.projectRoot, constitution); err != nil {
		return fmt.Errorf("failed to sync documentation: %w", err)
	}

	cw.logger.Info("[Constitution Watcher] Documentation synchronized")
	return nil
}

// updateForStructureChange updates Constitution when project structure changes
func (cw *ConstitutionWatcher) updateForStructureChange(ctx context.Context, event *ConstitutionUpdateEvent) error {
	cw.logger.Info("[Constitution Watcher] Project structure changed, validating compliance")

	// For now, just log the change. In future, could add rules dynamically
	cw.logger.WithField("description", event.Description).Info("[Constitution Watcher] Structure change detected")

	return nil
}

// flagCoverageViolation flags a test coverage violation
func (cw *ConstitutionWatcher) flagCoverageViolation(ctx context.Context) error {
	cw.logger.Warn("[Constitution Watcher] Test coverage dropped below 100% - CONST-002 violation")

	// This would typically trigger an alert or create a tracking issue
	// For now, just log the violation

	return nil
}

// UpdateLastCheckTime updates the last modification time checked
func (cw *ConstitutionWatcher) UpdateLastCheckTime() {
	cw.lastModTime = time.Now()
}
