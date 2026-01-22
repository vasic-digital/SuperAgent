// Package skills provides a skill loader for HelixAgent.
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// SkillLoader loads skills from the filesystem.
type SkillLoader struct {
	parser   *Parser
	registry *Registry
	log      *logrus.Logger
	loaded   map[string]*Skill
}

// LoaderConfig configures the skill loader.
type LoaderConfig struct {
	SkillsDir     string   // Base directory for skills
	Categories    []string // Categories to load (empty = all)
	EnabledSkills []string // Specific skills to enable (empty = all)
}

// NewSkillLoader creates a new skill loader.
func NewSkillLoader(registry *Registry) *SkillLoader {
	return &SkillLoader{
		parser:   NewParser(),
		registry: registry,
		log:      logrus.New(),
		loaded:   make(map[string]*Skill),
	}
}

// SetLogger sets the logger.
func (l *SkillLoader) SetLogger(log *logrus.Logger) {
	l.log = log
}

// LoadFromDirectory loads all skills from a directory.
func (l *SkillLoader) LoadFromDirectory(dir string) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, fmt.Errorf("skills directory not found: %s", dir)
	}

	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for SKILL.md files
		if info.IsDir() || !strings.EqualFold(info.Name(), "SKILL.md") {
			return nil
		}

		skill, err := l.parser.ParseFile(path)
		if err != nil {
			l.log.WithError(err).WithField("path", path).Warn("Failed to parse skill file")
			return nil // Continue loading other skills
		}

		// Extract category from path
		relPath, _ := filepath.Rel(dir, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) > 1 {
			// Use parent directory as category if not set
			if skill.Category == "" {
				skill.Category = parts[0]
			}
		}

		// Register the skill
		l.registry.RegisterSkill(skill)

		l.loaded[skill.Name] = skill
		count++
		l.log.WithFields(logrus.Fields{
			"name":     skill.Name,
			"category": skill.Category,
			"path":     path,
		}).Debug("Loaded skill")

		return nil
	})

	if err != nil {
		return count, fmt.Errorf("error walking skills directory: %w", err)
	}

	l.log.WithField("count", count).Info("Loaded skills from directory")
	return count, nil
}

// LoadFromConfig loads skills based on configuration.
func (l *SkillLoader) LoadFromConfig(cfg *LoaderConfig) (int, error) {
	if cfg.SkillsDir == "" {
		return 0, fmt.Errorf("skills directory not configured")
	}

	total := 0

	// If no categories specified, load all
	if len(cfg.Categories) == 0 {
		return l.LoadFromDirectory(cfg.SkillsDir)
	}

	// Load specific categories
	for _, category := range cfg.Categories {
		catDir := filepath.Join(cfg.SkillsDir, category)
		if _, err := os.Stat(catDir); os.IsNotExist(err) {
			l.log.WithField("category", category).Warn("Category directory not found")
			continue
		}

		count, err := l.LoadFromDirectory(catDir)
		if err != nil {
			l.log.WithError(err).WithField("category", category).Warn("Failed to load category")
			continue
		}
		total += count
	}

	// Filter by enabled skills if specified
	if len(cfg.EnabledSkills) > 0 {
		enabled := make(map[string]bool)
		for _, name := range cfg.EnabledSkills {
			enabled[name] = true
		}

		// Remove skills not in enabled list
		for name, skill := range l.loaded {
			if !enabled[name] {
				l.registry.Remove(name)
				delete(l.loaded, name)
				l.log.WithField("skill", skill.Name).Debug("Disabled skill not in enabled list")
			}
		}
	}

	return total, nil
}

// LoadBuiltinSkills loads the built-in skills from the default location.
func (l *SkillLoader) LoadBuiltinSkills() (int, error) {
	// Find the project root by looking for go.mod
	cwd, err := os.Getwd()
	if err != nil {
		return 0, err
	}

	// Walk up to find skills directory
	dir := cwd
	for {
		skillsDir := filepath.Join(dir, "skills")
		if _, err := os.Stat(skillsDir); err == nil {
			return l.LoadFromDirectory(skillsDir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return 0, fmt.Errorf("could not find skills directory")
}

// GetLoaded returns all loaded skills.
func (l *SkillLoader) GetLoaded() map[string]*Skill {
	return l.loaded
}

// GetLoadedCount returns the number of loaded skills.
func (l *SkillLoader) GetLoadedCount() int {
	return len(l.loaded)
}

// GetLoadedByCategory returns loaded skills grouped by category.
func (l *SkillLoader) GetLoadedByCategory() map[string][]*Skill {
	result := make(map[string][]*Skill)
	for _, skill := range l.loaded {
		result[skill.Category] = append(result[skill.Category], skill)
	}
	return result
}

// ReloadSkill reloads a specific skill from its source file.
func (l *SkillLoader) ReloadSkill(name string) error {
	skill, ok := l.loaded[name]
	if !ok {
		return fmt.Errorf("skill not found: %s", name)
	}

	if skill.FilePath == "" {
		return fmt.Errorf("skill has no source path: %s", name)
	}

	newSkill, err := l.parser.ParseFile(skill.FilePath)
	if err != nil {
		return fmt.Errorf("failed to reload skill: %w", err)
	}

	// Preserve the category
	if newSkill.Category == "" {
		newSkill.Category = skill.Category
	}

	// Update registry
	l.registry.Remove(name)
	l.registry.RegisterSkill(newSkill)

	l.loaded[newSkill.Name] = newSkill
	l.log.WithField("skill", name).Info("Reloaded skill")

	return nil
}

// SkillInventory provides an inventory of all available skills.
type SkillInventory struct {
	TotalSkills      int                    `json:"total_skills"`
	Categories       []string               `json:"categories"`
	SkillsByCategory map[string][]SkillInfo `json:"skills_by_category"`
}

// SkillInfo provides basic info about a skill.
type SkillInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Version     string   `json:"version"`
	Triggers    []string `json:"triggers"`
	ToolsUsed   string   `json:"tools_used"`
}

// GetInventory returns an inventory of all loaded skills.
func (l *SkillLoader) GetInventory() *SkillInventory {
	inventory := &SkillInventory{
		TotalSkills:      len(l.loaded),
		Categories:       make([]string, 0),
		SkillsByCategory: make(map[string][]SkillInfo),
	}

	categorySet := make(map[string]bool)
	for _, skill := range l.loaded {
		if !categorySet[skill.Category] {
			categorySet[skill.Category] = true
			inventory.Categories = append(inventory.Categories, skill.Category)
		}

		info := SkillInfo{
			Name:        skill.Name,
			Description: skill.Description,
			Category:    skill.Category,
			Version:     skill.Version,
			Triggers:    skill.TriggerPhrases,
			ToolsUsed:   skill.AllowedTools,
		}
		inventory.SkillsByCategory[skill.Category] = append(
			inventory.SkillsByCategory[skill.Category], info)
	}

	return inventory
}
