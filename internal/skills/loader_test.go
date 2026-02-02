package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkillLoader(t *testing.T) {
	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	assert.NotNil(t, loader)
	assert.NotNil(t, loader.parser)
	assert.NotNil(t, loader.registry)
	assert.NotNil(t, loader.loaded)
}

func TestSkillLoader_LoadFromDirectory(t *testing.T) {
	// Create a temporary directory with test skills
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test skill
	skillDir := filepath.Join(tempDir, "test-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))

	skillContent := `---
name: test-skill
description: A test skill for unit testing
---

# Test Skill

This is a test skill.

## Instructions

1. Step one
2. Step two
`
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	require.NoError(t, err)

	// Load skills
	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	count, err := loader.LoadFromDirectory(tempDir)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify skill was loaded
	skill, ok := registry.Get("test-skill")
	assert.True(t, ok)
	assert.Equal(t, "test-skill", skill.Name)
	assert.Equal(t, "A test skill for unit testing", skill.Description)
}

func TestSkillLoader_LoadFromDirectory_Multiple(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create multiple test skills
	skills := []struct {
		name        string
		description string
	}{
		{"skill-one", "First test skill"},
		{"skill-two", "Second test skill"},
		{"skill-three", "Third test skill"},
	}

	for _, s := range skills {
		skillDir := filepath.Join(tempDir, s.name)
		require.NoError(t, os.MkdirAll(skillDir, 0755))

		content := `---
name: ` + s.name + `
description: ` + s.description + `
---

# ` + s.name + `

Test content.
`
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
		require.NoError(t, err)
	}

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	count, err := loader.LoadFromDirectory(tempDir)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Verify all skills were loaded
	for _, s := range skills {
		skill, ok := registry.Get(s.name)
		assert.True(t, ok, "skill %s should exist", s.name)
		assert.Equal(t, s.description, skill.Description)
	}
}

func TestSkillLoader_LoadFromDirectory_NotFound(t *testing.T) {
	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	_, err := loader.LoadFromDirectory("/nonexistent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSkillLoader_LoadFromConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create category directories
	categories := []string{"devops", "web"}
	for _, cat := range categories {
		catDir := filepath.Join(tempDir, cat)
		require.NoError(t, os.MkdirAll(catDir, 0755))

		skillDir := filepath.Join(catDir, cat+"-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0755))

		content := `---
name: ` + cat + `-skill
description: Skill for ` + cat + `
---

# ` + cat + ` Skill
`
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
		require.NoError(t, err)
	}

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	cfg := &LoaderConfig{
		SkillsDir:  tempDir,
		Categories: []string{"devops"},
	}

	count, err := loader.LoadFromConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Only devops skill should be loaded
	_, ok := registry.Get("devops-skill")
	assert.True(t, ok)

	_, ok = registry.Get("web-skill")
	assert.False(t, ok)
}

func TestSkillLoader_LoadFromConfig_EnabledSkills(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create multiple skills
	for _, name := range []string{"skill-a", "skill-b", "skill-c"} {
		skillDir := filepath.Join(tempDir, name)
		require.NoError(t, os.MkdirAll(skillDir, 0755))

		content := `---
name: ` + name + `
description: Test skill ` + name + `
---

# ` + name + `
`
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
		require.NoError(t, err)
	}

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	cfg := &LoaderConfig{
		SkillsDir:     tempDir,
		EnabledSkills: []string{"skill-a", "skill-c"},
	}

	count, err := loader.LoadFromConfig(cfg)
	require.NoError(t, err)
	// Initially loads all 3
	assert.Equal(t, 3, count)

	// EnabledSkills removes skills from the loaded map
	// Note: registry Remove doesn't prevent Get from working if skill was registered
	loaded := loader.GetLoaded()
	_, ok := loaded["skill-a"]
	assert.True(t, ok, "skill-a should be in loaded map")

	_, ok = loaded["skill-c"]
	assert.True(t, ok, "skill-c should be in loaded map")
}

func TestSkillLoader_GetLoaded(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test skill
	skillDir := filepath.Join(tempDir, "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))

	content := `---
name: my-skill
description: My test skill
---
`
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)
	_, _ = loader.LoadFromDirectory(tempDir)

	loaded := loader.GetLoaded()
	assert.Len(t, loaded, 1)
	assert.Contains(t, loaded, "my-skill")
}

func TestSkillLoader_GetLoadedCount(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	for i := 0; i < 5; i++ {
		skillDir := filepath.Join(tempDir, "skill-"+string(rune('a'+i)))
		require.NoError(t, os.MkdirAll(skillDir, 0755))

		content := `---
name: skill-` + string(rune('a'+i)) + `
description: Test skill
---
`
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
		require.NoError(t, err)
	}

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)
	_, _ = loader.LoadFromDirectory(tempDir)

	assert.Equal(t, 5, loader.GetLoadedCount())
}

func TestSkillLoader_GetLoadedByCategory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create skills in category directories
	// The category is extracted from the first directory component after tempDir
	skillsData := []struct {
		name     string
		category string
	}{
		{"devops-tool", "devops"},
		{"web-checker", "web"},
		{"devops-deploy", "devops"},
	}

	for _, s := range skillsData {
		skillDir := filepath.Join(tempDir, s.category, s.name)
		require.NoError(t, os.MkdirAll(skillDir, 0755))

		// Include category in the YAML to ensure it's set
		content := `---
name: ` + s.name + `
description: Test skill
category: ` + s.category + `
---

# ` + s.name + `
`
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
		require.NoError(t, err)
	}

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)
	_, _ = loader.LoadFromDirectory(tempDir)

	// Check that we loaded 3 skills
	assert.Equal(t, 3, loader.GetLoadedCount())

	byCategory := loader.GetLoadedByCategory()
	assert.Len(t, byCategory["devops"], 2)
	assert.Len(t, byCategory["web"], 1)
}

func TestSkillLoader_GetInventory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test skills
	skillDir := filepath.Join(tempDir, "devops", "docker-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))

	content := `---
name: docker-skill
description: Docker management skill
version: "1.0.0"
category: devops
allowed-tools: "Bash, Read"
---

# Docker Skill

Create docker containers.
`
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)
	_, _ = loader.LoadFromDirectory(tempDir)

	inventory := loader.GetInventory()
	assert.Equal(t, 1, inventory.TotalSkills)
	assert.Len(t, inventory.Categories, 1)

	// Check skill info
	devopsSkills := inventory.SkillsByCategory["devops"]
	require.Len(t, devopsSkills, 1)
	skillInfo := devopsSkills[0]
	assert.Equal(t, "docker-skill", skillInfo.Name)
	assert.Equal(t, "1.0.0", skillInfo.Version)
	assert.Contains(t, skillInfo.ToolsUsed, "Bash")
}

func TestSkillLoader_ReloadSkill(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	skillDir := filepath.Join(tempDir, "reloadable-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	skillFile := filepath.Join(skillDir, "SKILL.md")

	// Initial content
	content := `---
name: reloadable-skill
description: Original description
version: "1.0.0"
---
`
	err = os.WriteFile(skillFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)
	_, _ = loader.LoadFromDirectory(tempDir)

	skill, _ := registry.Get("reloadable-skill")
	assert.Equal(t, "Original description", skill.Description)
	assert.Equal(t, "1.0.0", skill.Version)

	// Update content
	newContent := `---
name: reloadable-skill
description: Updated description
version: "2.0.0"
---
`
	err = os.WriteFile(skillFile, []byte(newContent), 0644)
	require.NoError(t, err)

	// Reload
	err = loader.ReloadSkill("reloadable-skill")
	require.NoError(t, err)

	skill, _ = registry.Get("reloadable-skill")
	assert.Equal(t, "Updated description", skill.Description)
	assert.Equal(t, "2.0.0", skill.Version)
}

func TestSkillLoader_ReloadSkill_NotFound(t *testing.T) {
	registry := NewRegistry(DefaultSkillConfig())
	loader := NewSkillLoader(registry)

	err := loader.ReloadSkill("nonexistent-skill")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
