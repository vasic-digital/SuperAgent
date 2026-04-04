// Package templates provides template management
package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// TemplateManager handles template CRUD operations
type TemplateManager struct {
	templatesDir string
	templates    map[string]*ContextTemplate
	mu           sync.RWMutex
}

// Manager is an alias for TemplateManager for backward compatibility
type Manager = TemplateManager

// ManagerConfig holds manager configuration
type ManagerConfig struct {
	TemplatesDir string
	MaxTemplates int
}

// DefaultManagerConfig returns default configuration
func DefaultManagerConfig() ManagerConfig {
	homeDir, _ := os.UserHomeDir()
	return ManagerConfig{
		TemplatesDir: filepath.Join(homeDir, ".helixagent", "templates"),
		MaxTemplates: 100,
	}
}

// NewManager creates a new template manager
func NewManager(config ManagerConfig) (*TemplateManager, error) {
	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(config.TemplatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	m := &TemplateManager{
		templatesDir: config.TemplatesDir,
		templates:    make(map[string]*ContextTemplate),
	}

	// Load built-in templates
	if err := m.loadBuiltInTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load built-in templates: %w", err)
	}

	// Load user templates
	if err := m.loadUserTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load user templates: %w", err)
	}

	return m, nil
}

// Create creates a new template
func (m *TemplateManager) Create(template *ContextTemplate) error {
	if err := template.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.templates[template.Metadata.ID]; exists {
		return fmt.Errorf("template with ID %s already exists", template.Metadata.ID)
	}

	template.Metadata.CreatedAt = time.Now()
	template.Metadata.UpdatedAt = time.Now()

	// Save to file
	if err := m.saveTemplate(template); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	m.templates[template.Metadata.ID] = template
	return nil
}

// Get retrieves a template by ID
func (m *TemplateManager) Get(id string) (*ContextTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	template, exists := m.templates[id]
	if !exists {
		return nil, fmt.Errorf("template %s not found", id)
	}

	return template, nil
}

// Update updates an existing template
func (m *TemplateManager) Update(template *ContextTemplate) error {
	if err := template.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.templates[template.Metadata.ID]; !exists {
		return fmt.Errorf("template %s not found", template.Metadata.ID)
	}

	template.Metadata.UpdatedAt = time.Now()

	if err := m.saveTemplate(template); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	m.templates[template.Metadata.ID] = template
	return nil
}

// Delete deletes a template
func (m *TemplateManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.templates[id]; !exists {
		return fmt.Errorf("template %s not found", id)
	}

	// Delete file
	path := m.getTemplatePath(id)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	delete(m.templates, id)
	return nil
}

// List returns all templates
func (m *TemplateManager) List() []*ContextTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*ContextTemplate, 0, len(m.templates))
	for _, template := range m.templates {
		list = append(list, template)
	}

	return list
}

// ListByTag returns templates with a specific tag
func (m *TemplateManager) ListByTag(tag string) []*ContextTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*ContextTemplate
	for _, template := range m.templates {
		for _, t := range template.Metadata.Tags {
			if strings.EqualFold(t, tag) {
				list = append(list, template)
				break
			}
		}
	}

	return list
}

// ApplyTemplate applies a template with variables and returns the resolved context
func (m *TemplateManager) ApplyTemplate(templateID string, variables map[string]string) (*ResolvedContext, error) {
	m.mu.RLock()
	template, exists := m.templates[templateID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template %s not found", templateID)
	}

	resolver := NewResolver(".")
	return resolver.Resolve(template, variables)
}

// GetTemplate retrieves a template by ID (alias for Get)
func (m *TemplateManager) GetTemplate(id string) (*ContextTemplate, error) {
	return m.Get(id)
}

// saveTemplate saves a template to file
func (m *TemplateManager) saveTemplate(template *ContextTemplate) error {
	path := m.getTemplatePath(template.Metadata.ID)

	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// getTemplatePath returns the file path for a template
func (m *TemplateManager) getTemplatePath(id string) string {
	return filepath.Join(m.templatesDir, id+".yaml")
}

// loadBuiltInTemplates loads built-in templates
func (m *TemplateManager) loadBuiltInTemplates() error {
	// Create built-in templates
	builtIns := []*ContextTemplate{
		m.createOnboardingTemplate(),
		m.createBugFixTemplate(),
		m.createCodeReviewTemplate(),
		m.createFeatureDevTemplate(),
	}

	for _, template := range builtIns {
		// Only create if doesn't exist
		path := m.getTemplatePath(template.Metadata.ID)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := m.saveTemplate(template); err != nil {
				return err
			}
		}
		m.templates[template.Metadata.ID] = template
	}

	return nil
}

// loadUserTemplates loads user-created templates
func (m *TemplateManager) loadUserTemplates() error {
	entries, err := os.ReadDir(m.templatesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(m.templatesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var template ContextTemplate
		if err := yaml.Unmarshal(data, &template); err != nil {
			continue
		}

		m.templates[template.Metadata.ID] = &template
	}

	return nil
}

// Built-in template creators
func (m *TemplateManager) createOnboardingTemplate() *ContextTemplate {
	return &ContextTemplate{
		APIVersion: "v1",
		Kind:       "ContextTemplate",
		Metadata: TemplateMetadata{
			ID:          "onboarding",
			Name:        "Project Onboarding",
			Description: "Get up to speed with a new project",
			Author:      "helixagent",
			Version:     "1.0.0",
			Tags:        []string{"onboarding", "documentation"},
		},
		Spec: TemplateSpec{
			Files: FileSpec{
				Include: []string{
					"README.md",
					"CONTRIBUTING.md",
					"docs/**/*.md",
					"Makefile",
					"package.json",
					"go.mod",
					"pyproject.toml",
				},
				Exclude: []string{
					"docs/node_modules/**",
				},
			},
			Instructions: `Help me understand this codebase. Focus on:
1. Project structure and architecture
2. Key technologies and dependencies
3. Development workflow
4. Testing approach
5. Important conventions`,
		},
	}
}

func (m *TemplateManager) createBugFixTemplate() *ContextTemplate {
	return &ContextTemplate{
		APIVersion: "v1",
		Kind:       "ContextTemplate",
		Metadata: TemplateMetadata{
			ID:          "bug-fix",
			Name:        "Bug Fix",
			Description: "Context for investigating and fixing bugs",
			Author:      "helixagent",
			Version:     "1.0.0",
			Tags:        []string{"debugging", "bug-fix"},
		},
		Spec: TemplateSpec{
			Files: FileSpec{
				Include: []string{
					"**/*test*.go",
					"**/*test*.py",
					"logs/**",
				},
			},
			GitContext: GitContextSpec{
				RecentCommits: RecentCommitsSpec{
					Enabled: true,
					Count:   10,
				},
			},
			Variables: []VariableDef{
				{
					Name:        "error_message",
					Description: "The error message or exception",
					Required:    true,
				},
				{
					Name:        "affected_component",
					Description: "Component where bug occurs",
					Required:    false,
				},
			},
			Instructions: `Help me fix this bug. Consider:
1. Root cause analysis
2. Impact assessment
3. Fix implementation
4. Test coverage
5. Regression prevention`,
		},
	}
}

func (m *TemplateManager) createCodeReviewTemplate() *ContextTemplate {
	return &ContextTemplate{
		APIVersion: "v1",
		Kind:       "ContextTemplate",
		Metadata: TemplateMetadata{
			ID:          "code-review",
			Name:        "Code Review",
			Description: "Context for reviewing code changes",
			Author:      "helixagent",
			Version:     "1.0.0",
			Tags:        []string{"review", "quality"},
		},
		Spec: TemplateSpec{
			GitContext: GitContextSpec{
				BranchDiff: BranchDiffSpec{
					Enabled:  true,
					Base:     "main",
					MaxFiles: 50,
				},
				RecentCommits: RecentCommitsSpec{
					Enabled: true,
					Count:   3,
				},
			},
			Files: FileSpec{
				Include: []string{
					"**/*test*",
					"CODEOWNERS",
				},
			},
			Instructions: `Review this code for:
1. Correctness and logic
2. Code quality and style
3. Test coverage
4. Documentation
5. Performance implications
6. Security considerations`,
		},
	}
}

func (m *TemplateManager) createFeatureDevTemplate() *ContextTemplate {
	return &ContextTemplate{
		APIVersion: "v1",
		Kind:       "ContextTemplate",
		Metadata: TemplateMetadata{
			ID:          "feature-dev",
			Name:        "Feature Development",
			Description: "Context for developing new features",
			Author:      "helixagent",
			Version:     "1.0.0",
			Tags:        []string{"development", "feature"},
		},
		Spec: TemplateSpec{
			Files: FileSpec{
				Include: []string{
					"README.md",
					"docs/ARCHITECTURE.md",
					"docs/API.md",
					"src/**/*",
				},
				Exclude: []string{
					"*_test.go",
					"vendor/**",
				},
			},
			Variables: []VariableDef{
				{
					Name:        "feature_name",
					Description: "Name of the feature being developed",
					Required:    true,
				},
				{
					Name:        "priority",
					Description: "Feature priority",
					Required:    false,
					Default:     "medium",
					Options:     []string{"low", "medium", "high", "critical"},
				},
			},
			Instructions: `You are helping develop a new feature. Consider:
1. API compatibility
2. Test coverage
3. Documentation updates
4. Performance implications`,
		},
	}
}
