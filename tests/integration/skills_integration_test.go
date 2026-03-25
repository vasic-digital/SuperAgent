//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/agents"
	"dev.helix.agent/internal/skills"
)

// TestSkillRegistry_RegisterAndRetrieve verifies that skills can be registered
// and retrieved by name and category.
func TestSkillRegistry_RegisterAndRetrieve(t *testing.T) {
	config := skills.DefaultSkillConfig()
	config.MinConfidence = 0.3
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	skill := &skills.Skill{
		Name:           "go-testing",
		Description:    "Run Go tests with coverage",
		Category:       "development",
		TriggerPhrases: []string{"run tests", "go test", "test coverage"},
		Instructions:   "Use `go test ./... -cover` to run all tests with coverage.",
	}
	service.RegisterSkill(skill)

	got, ok := service.GetSkill("go-testing")
	require.True(t, ok, "skill should be retrievable by name")
	assert.Equal(t, "go-testing", got.Name)
	assert.Equal(t, "development", got.Category)
}

// TestSkillRegistry_GetByCategory verifies filtering by category returns only
// skills belonging to that category.
func TestSkillRegistry_GetByCategory(t *testing.T) {
	config := skills.DefaultSkillConfig()
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	devSkill := &skills.Skill{
		Name:           "lint-code",
		Description:    "Run linter on codebase",
		Category:       "development",
		TriggerPhrases: []string{"lint", "run linter"},
	}
	opsSkill := &skills.Skill{
		Name:           "restart-service",
		Description:    "Restart a systemd service",
		Category:       "operations",
		TriggerPhrases: []string{"restart service", "systemctl restart"},
	}
	service.RegisterSkill(devSkill)
	service.RegisterSkill(opsSkill)

	devSkills := service.GetSkillsByCategory("development")
	assert.NotEmpty(t, devSkills)
	for _, s := range devSkills {
		assert.Equal(t, "development", s.Category)
	}

	opsSkills := service.GetSkillsByCategory("operations")
	assert.NotEmpty(t, opsSkills)
	for _, s := range opsSkills {
		assert.Equal(t, "operations", s.Category)
	}
}

// TestSkillMatcher_ExactMatch verifies that exact trigger phrases produce
// high-confidence matches.
func TestSkillMatcher_ExactMatch(t *testing.T) {
	config := skills.DefaultSkillConfig()
	config.MinConfidence = 0.3
	config.EnableSemanticMatching = false
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	service.RegisterSkill(&skills.Skill{
		Name:           "docker-build",
		Description:    "Build a Docker image",
		Category:       "devops",
		TriggerPhrases: []string{"docker build", "build docker image", "build container"},
	})

	ctx := context.Background()
	matches, err := service.FindSkills(ctx, "docker build my-app")
	require.NoError(t, err)
	assert.NotEmpty(t, matches, "exact trigger should produce at least one match")

	found := false
	for _, m := range matches {
		if m.Skill.Name == "docker-build" {
			found = true
			assert.GreaterOrEqual(t, m.Confidence, 0.3)
		}
	}
	assert.True(t, found, "docker-build skill should be in matches")
}

// TestSkillMatcher_NoMatchForUnrelatedQuery verifies that unrelated queries
// do not produce spurious skill matches.
func TestSkillMatcher_NoMatchForUnrelatedQuery(t *testing.T) {
	config := skills.DefaultSkillConfig()
	config.MinConfidence = 0.8
	config.EnableSemanticMatching = false
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	service.RegisterSkill(&skills.Skill{
		Name:           "specific-skill",
		Description:    "A very specific skill",
		Category:       "test",
		TriggerPhrases: []string{"zephyr turbine calibration"},
	})

	ctx := context.Background()
	matches, err := service.FindSkills(ctx, "hello world")
	require.NoError(t, err)
	assert.Empty(t, matches, "unrelated query should not match")
}

// TestSkillMatcher_BestSkill verifies FindBestSkill returns the single highest
// confidence match.
func TestSkillMatcher_BestSkill(t *testing.T) {
	config := skills.DefaultSkillConfig()
	config.MinConfidence = 0.3
	config.EnableSemanticMatching = false
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	service.RegisterSkill(&skills.Skill{
		Name:           "git-commit",
		Description:    "Commit changes with git",
		Category:       "vcs",
		TriggerPhrases: []string{"git commit", "commit changes", "make commit"},
	})
	service.RegisterSkill(&skills.Skill{
		Name:           "git-push",
		Description:    "Push commits to remote",
		Category:       "vcs",
		TriggerPhrases: []string{"git push", "push to remote"},
	})

	ctx := context.Background()
	best, err := service.FindBestSkill(ctx, "git commit my work")
	require.NoError(t, err)
	require.NotNil(t, best, "should find best skill for clear trigger")
	assert.Equal(t, "git-commit", best.Skill.Name)
}

// TestSkillService_SearchSkills verifies text search across skill names and
// descriptions.
func TestSkillService_SearchSkills(t *testing.T) {
	config := skills.DefaultSkillConfig()
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	service.RegisterSkill(&skills.Skill{
		Name:        "python-venv",
		Description: "Create and manage Python virtual environments",
		Category:    "python",
	})
	service.RegisterSkill(&skills.Skill{
		Name:        "python-pip",
		Description: "Install Python packages with pip",
		Category:    "python",
	})
	service.RegisterSkill(&skills.Skill{
		Name:        "node-npm",
		Description: "Manage Node.js packages with npm",
		Category:    "javascript",
	})

	results := service.SearchSkills("python")
	assert.GreaterOrEqual(t, len(results), 2, "search for 'python' should find at least 2 skills")

	for _, s := range results {
		nameOrDesc := s.Name + " " + s.Description
		assert.Contains(t, nameOrDesc, "python",
			"every result should mention 'python' in name or description")
	}
}

// TestSkillService_GetAllSkills verifies that GetAllSkills returns all registered
// skills in the registry.
func TestSkillService_GetAllSkills(t *testing.T) {
	config := skills.DefaultSkillConfig()
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	for i := 0; i < 5; i++ {
		service.RegisterSkill(&skills.Skill{
			Name:        "skill-" + string(rune('A'+i)),
			Description: "Test skill",
			Category:    "test",
		})
	}

	all := service.GetAllSkills()
	assert.GreaterOrEqual(t, len(all), 5)
}

// TestSkillService_GetCategories verifies that distinct categories are returned.
func TestSkillService_GetCategories(t *testing.T) {
	config := skills.DefaultSkillConfig()
	service := skills.NewService(config)
	service.Start()
	defer service.Shutdown() //nolint:errcheck

	service.RegisterSkill(&skills.Skill{Name: "s1", Category: "cat-alpha"})
	service.RegisterSkill(&skills.Skill{Name: "s2", Category: "cat-beta"})
	service.RegisterSkill(&skills.Skill{Name: "s3", Category: "cat-alpha"})

	categories := service.GetCategories()
	categorySet := make(map[string]bool)
	for _, c := range categories {
		categorySet[c] = true
	}

	assert.True(t, categorySet["cat-alpha"], "cat-alpha should be in categories")
	assert.True(t, categorySet["cat-beta"], "cat-beta should be in categories")
}

// TestCLIAgentRegistry_Contains48Agents verifies that all 48 CLI agents are
// present in the registry with required fields populated.
func TestCLIAgentRegistry_Contains48Agents(t *testing.T) {
	registry := agents.CLIAgentRegistry

	assert.GreaterOrEqual(t, len(registry), 48,
		"CLIAgentRegistry should contain at least 48 agents")

	for name, agent := range registry {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, agent.Name, "agent.Name must not be empty")
			assert.NotEmpty(t, agent.Description, "agent.Description must not be empty")
			assert.NotEmpty(t, agent.EntryPoint, "agent.EntryPoint must not be empty")
			assert.NotEmpty(t, agent.ConfigLocation, "agent.ConfigLocation must not be empty")
			assert.NotEmpty(t, agent.Protocols, "agent.Protocols must not be empty")
		})
	}
}

// TestCLIAgentRegistry_KeyAgentsPresent verifies that specific well-known agents
// are registered in the registry.
func TestCLIAgentRegistry_KeyAgentsPresent(t *testing.T) {
	registry := agents.CLIAgentRegistry

	requiredAgents := []string{
		"OpenCode",
		"Crush",
		"ClaudeCode",
		"KiloCode",
		"HelixCode",
	}

	for _, name := range requiredAgents {
		t.Run(name, func(t *testing.T) {
			agent, ok := registry[name]
			require.True(t, ok, "agent %s must be present in registry", name)
			assert.NotNil(t, agent)
			assert.Equal(t, name, agent.Name)
		})
	}
}

// TestSkillTypes_ParseAllowedTools verifies that the ParseAllowedTools function
// correctly parses various tool string formats.
func TestSkillTypes_ParseAllowedTools(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"single tool", "Read", 1},
		{"multiple tools", "Read, Write, Edit", 3},
		{"tool with constraints", "Bash(cmd:*)", 1},
		{"mixed", "Read, Bash(cmd:ls), Glob(pattern:*.go)", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := skills.ParseAllowedTools(tt.input)
			assert.Len(t, tools, tt.expected)
		})
	}
}

// TestSkillService_StartAndIsRunning verifies the service lifecycle transitions.
func TestSkillService_StartAndIsRunning(t *testing.T) {
	config := skills.DefaultSkillConfig()
	service := skills.NewService(config)

	assert.False(t, service.IsRunning(), "service should not be running before Start")

	service.Start()
	assert.True(t, service.IsRunning(), "service should be running after Start")

	err := service.Shutdown()
	require.NoError(t, err)
	assert.False(t, service.IsRunning(), "service should not be running after Shutdown")
}
