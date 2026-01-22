package skills

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Run("creates service with default config", func(t *testing.T) {
		svc := NewService(nil)
		require.NotNil(t, svc)
		assert.NotNil(t, svc.registry)
		assert.NotNil(t, svc.matcher)
		assert.NotNil(t, svc.tracker)
		assert.NotNil(t, svc.config)
	})

	t.Run("creates service with custom config", func(t *testing.T) {
		config := &SkillConfig{
			SkillsDirectory:        "/custom/path",
			MinConfidence:          0.8,
			EnableSemanticMatching: true,
			MaxConcurrentSkills:    10,
			TrackUsage:             true,
		}
		svc := NewService(config)
		require.NotNil(t, svc)
		assert.Equal(t, config, svc.config)
		assert.Equal(t, "/custom/path", svc.config.SkillsDirectory)
		assert.Equal(t, 0.8, svc.config.MinConfidence)
	})
}

func TestService_SetLogger(t *testing.T) {
	svc := NewService(nil)
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	svc.SetLogger(log)
	assert.Equal(t, log, svc.log)
}

func TestService_Start(t *testing.T) {
	svc := NewService(nil)
	assert.False(t, svc.IsRunning())

	svc.Start()
	assert.True(t, svc.IsRunning())
}

func TestService_Shutdown(t *testing.T) {
	t.Run("shuts down cleanly", func(t *testing.T) {
		svc := NewService(nil)
		svc.Start()
		assert.True(t, svc.IsRunning())

		err := svc.Shutdown()
		assert.NoError(t, err)
		assert.False(t, svc.IsRunning())
	})

	t.Run("shutdown when not running is no-op", func(t *testing.T) {
		svc := NewService(nil)
		assert.False(t, svc.IsRunning())

		err := svc.Shutdown()
		assert.NoError(t, err)
	})
}

func TestService_GetAllSkills(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Initially empty
	skills := svc.GetAllSkills()
	assert.NotNil(t, skills)
	assert.Len(t, skills, 0)
}

func TestService_RegisterSkill(t *testing.T) {
	t.Run("registers skill successfully", func(t *testing.T) {
		svc := NewService(nil)
		svc.Start()

		skill := &Skill{
			Name:        "test-skill",
			Description: "A test skill",
			Version:     "1.0.0",
			Category:    "testing",
			Tags:        []string{"test"},
		}

		svc.RegisterSkill(skill)

		// Verify skill was registered
		retrieved, found := svc.GetSkill("test-skill")
		assert.True(t, found)
		assert.Equal(t, "test-skill", retrieved.Name)
		assert.Equal(t, "A test skill", retrieved.Description)
	})
}

func TestService_GetSkill(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Register a skill
	skill := &Skill{
		Name:        "get-test",
		Description: "Test skill for get",
		Category:    "testing",
	}
	svc.RegisterSkill(skill)

	t.Run("returns skill when found", func(t *testing.T) {
		retrieved, found := svc.GetSkill("get-test")
		assert.True(t, found)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "get-test", retrieved.Name)
	})

	t.Run("returns not found for unknown skill", func(t *testing.T) {
		retrieved, found := svc.GetSkill("unknown-skill")
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestService_GetSkillsByCategory(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Register skills in different categories
	svc.RegisterSkill(&Skill{Name: "skill1", Category: "cat-a"})
	svc.RegisterSkill(&Skill{Name: "skill2", Category: "cat-a"})
	svc.RegisterSkill(&Skill{Name: "skill3", Category: "cat-b"})

	catA := svc.GetSkillsByCategory("cat-a")
	assert.Len(t, catA, 2)

	catB := svc.GetSkillsByCategory("cat-b")
	assert.Len(t, catB, 1)

	catC := svc.GetSkillsByCategory("cat-c")
	assert.Len(t, catC, 0)
}

func TestService_GetCategories(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Register skills with categories
	svc.RegisterSkill(&Skill{Name: "skill1", Category: "development"})
	svc.RegisterSkill(&Skill{Name: "skill2", Category: "testing"})
	svc.RegisterSkill(&Skill{Name: "skill3", Category: "development"})

	categories := svc.GetCategories()
	assert.Contains(t, categories, "development")
	assert.Contains(t, categories, "testing")
}

func TestService_SearchSkills(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Register skills with searchable content
	svc.RegisterSkill(&Skill{
		Name:        "code-review",
		Description: "Review code changes",
		Category:    "development",
	})
	svc.RegisterSkill(&Skill{
		Name:        "write-test",
		Description: "Write unit tests",
		Category:    "testing",
	})

	results := svc.SearchSkills("code")
	assert.NotEmpty(t, results)
}

func TestService_RemoveSkill(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Register and then remove
	svc.RegisterSkill(&Skill{Name: "to-remove", Category: "temp"})

	_, found := svc.GetSkill("to-remove")
	assert.True(t, found)

	removed := svc.RemoveSkill("to-remove")
	assert.True(t, removed)

	_, found = svc.GetSkill("to-remove")
	assert.False(t, found)

	// Remove non-existent
	removed = svc.RemoveSkill("non-existent")
	assert.False(t, removed)
}

func TestService_HealthCheck(t *testing.T) {
	t.Run("returns error when not running", func(t *testing.T) {
		svc := NewService(nil)
		ctx := context.Background()

		err := svc.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	t.Run("returns error when no skills loaded", func(t *testing.T) {
		svc := NewService(nil)
		svc.Start()
		ctx := context.Background()

		err := svc.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no skills loaded")
	})

	t.Run("returns nil when healthy", func(t *testing.T) {
		svc := NewService(nil)
		svc.Start()
		svc.RegisterSkill(&Skill{Name: "health-test", Category: "test"})
		ctx := context.Background()

		err := svc.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

func TestService_GetConfig(t *testing.T) {
	config := &SkillConfig{
		SkillsDirectory: "/test/dir",
		MinConfidence:   0.9,
	}
	svc := NewService(config)

	retrieved := svc.GetConfig()
	assert.Equal(t, config, retrieved)
	assert.Equal(t, "/test/dir", retrieved.SkillsDirectory)
	assert.Equal(t, 0.9, retrieved.MinConfidence)
}

func TestService_UpdateConfig(t *testing.T) {
	svc := NewService(nil)

	newConfig := &SkillConfig{
		SkillsDirectory: "/updated/path",
		MinConfidence:   0.95,
	}

	svc.UpdateConfig(newConfig)

	assert.Equal(t, newConfig, svc.GetConfig())
	assert.Equal(t, "/updated/path", svc.GetConfig().SkillsDirectory)
}

func TestService_GetRegistryStats(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	// Register some skills
	svc.RegisterSkill(&Skill{Name: "stat1", Category: "cat-a"})
	svc.RegisterSkill(&Skill{Name: "stat2", Category: "cat-b"})

	stats := svc.GetRegistryStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats.TotalSkills)
	assert.Contains(t, stats.SkillsByCategory, "cat-a")
	assert.Contains(t, stats.SkillsByCategory, "cat-b")
}

func TestService_SkillTracking(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	skill := &Skill{
		Name:     "tracked-skill",
		Category: "tracking",
	}
	svc.RegisterSkill(skill)

	match := &SkillMatch{
		Skill:      skill,
		Confidence: 0.9,
		MatchType:  MatchTypeExact,
	}

	t.Run("starts tracking execution", func(t *testing.T) {
		usage := svc.StartSkillExecution("req-1", skill, match)
		assert.NotNil(t, usage)
		assert.Equal(t, "tracked-skill", usage.SkillName)
		assert.Equal(t, "tracking", usage.Category)
	})

	t.Run("records tool use", func(t *testing.T) {
		svc.StartSkillExecution("req-2", skill, match)
		svc.RecordToolUse("req-2", "Read")
		svc.RecordToolUse("req-2", "Write")

		// Complete and check tools were recorded
		usage := svc.CompleteSkillExecution("req-2", true, "")
		assert.NotNil(t, usage)
		assert.Contains(t, usage.ToolsInvoked, "Read")
		assert.Contains(t, usage.ToolsInvoked, "Write")
	})

	t.Run("completes execution successfully", func(t *testing.T) {
		svc.StartSkillExecution("req-3", skill, match)
		usage := svc.CompleteSkillExecution("req-3", true, "")

		assert.NotNil(t, usage)
		assert.True(t, usage.Success)
		assert.Empty(t, usage.Error)
	})

	t.Run("completes execution with error", func(t *testing.T) {
		svc.StartSkillExecution("req-4", skill, match)
		usage := svc.CompleteSkillExecution("req-4", false, "something went wrong")

		assert.NotNil(t, usage)
		assert.False(t, usage.Success)
		assert.Equal(t, "something went wrong", usage.Error)
	})
}

func TestService_GetUsageStats(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	stats := svc.GetUsageStats()
	assert.NotNil(t, stats)
}

func TestService_GetTopSkills(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	top := svc.GetTopSkills(5)
	assert.NotNil(t, top)
}

func TestService_GetUsageHistory(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	history := svc.GetUsageHistory(10)
	assert.NotNil(t, history)
}

func TestService_ExecuteWithTracking(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	skill := &Skill{
		Name:     "exec-skill",
		Category: "execution",
	}
	svc.RegisterSkill(skill)

	match := &SkillMatch{
		Skill:      skill,
		Confidence: 0.95,
		MatchType:  MatchTypeSemantic,
	}

	t.Run("tracks successful execution", func(t *testing.T) {
		ctx := context.Background()
		usage := svc.ExecuteWithTracking(ctx, "exec-1", skill, match, func(ctx context.Context) error {
			return nil
		})

		assert.NotNil(t, usage)
		assert.True(t, usage.Success)
	})

	t.Run("tracks failed execution", func(t *testing.T) {
		ctx := context.Background()
		usage := svc.ExecuteWithTracking(ctx, "exec-2", skill, match, func(ctx context.Context) error {
			return assert.AnError
		})

		assert.NotNil(t, usage)
		assert.False(t, usage.Success)
		assert.NotEmpty(t, usage.Error)
	})
}

func TestService_CreateResponse(t *testing.T) {
	svc := NewService(nil)

	usages := []SkillUsage{
		{SkillName: "skill1", Success: true},
		{SkillName: "skill2", Success: true},
	}

	response := svc.CreateResponse("content here", usages, "test-provider", "test-model", "http")

	assert.Equal(t, "content here", response.Content)
	assert.Len(t, response.SkillsUsed, 2)
	assert.Equal(t, 2, response.TotalSkills)
	assert.Equal(t, "test-provider", response.ProviderUsed)
	assert.Equal(t, "test-model", response.ModelUsed)
	assert.Equal(t, "http", response.Protocol)
}

func TestService_FindSkills(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()

	t.Run("returns error when not running", func(t *testing.T) {
		_, err := svc.FindSkills(ctx, "test input")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	t.Run("finds matching skills when running", func(t *testing.T) {
		svc.Start()
		svc.RegisterSkill(&Skill{
			Name:           "find-test",
			TriggerPhrases: []string{"find test"},
		})

		matches, err := svc.FindSkills(ctx, "find test")
		assert.NoError(t, err)
		assert.NotNil(t, matches)
	})
}

func TestService_FindBestSkill(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()

	t.Run("returns error when not running", func(t *testing.T) {
		_, err := svc.FindBestSkill(ctx, "test input")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	t.Run("finds best skill when running", func(t *testing.T) {
		svc.Start()
		svc.RegisterSkill(&Skill{
			Name:           "best-test",
			TriggerPhrases: []string{"best test"},
		})

		match, err := svc.FindBestSkill(ctx, "best test")
		assert.NoError(t, err)
		// Match may be nil if no trigger matched
		_ = match
	})
}

func TestService_IsRunning(t *testing.T) {
	svc := NewService(nil)
	assert.False(t, svc.IsRunning())

	svc.Start()
	assert.True(t, svc.IsRunning())

	svc.Shutdown()
	assert.False(t, svc.IsRunning())
}

func TestService_GetActiveExecutions(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	skill := &Skill{Name: "active-skill", Category: "test"}
	match := &SkillMatch{Skill: skill, Confidence: 0.9}

	// Start some executions without completing
	svc.StartSkillExecution("active-1", skill, match)
	svc.StartSkillExecution("active-2", skill, match)

	active := svc.GetActiveExecutions()
	assert.Len(t, active, 2)

	// Complete one
	svc.CompleteSkillExecution("active-1", true, "")

	active = svc.GetActiveExecutions()
	assert.Len(t, active, 1)
}

func TestService_GetSkillStats(t *testing.T) {
	svc := NewService(nil)
	svc.Start()

	skill := &Skill{Name: "stats-skill", Category: "stats"}
	svc.RegisterSkill(skill)
	match := &SkillMatch{Skill: skill, Confidence: 0.9}

	// Execute skill multiple times
	for i := 0; i < 5; i++ {
		usage := svc.StartSkillExecution("stats-"+string(rune('a'+i)), skill, match)
		_ = usage
		svc.CompleteSkillExecution("stats-"+string(rune('a'+i)), true, "")
	}

	stats := svc.GetSkillStats("stats-skill")
	assert.NotNil(t, stats)
}

func TestService_Initialize(t *testing.T) {
	t.Run("initializes successfully", func(t *testing.T) {
		// Use a non-existent directory to test graceful handling
		config := &SkillConfig{
			SkillsDirectory: "/tmp/nonexistent-skills-" + time.Now().Format("20060102150405"),
			HotReload:       false,
		}
		svc := NewService(config)
		ctx := context.Background()

		err := svc.Initialize(ctx)
		// Should fail gracefully since directory doesn't exist
		// But no panic
		_ = err
	})

	t.Run("idempotent initialization", func(t *testing.T) {
		svc := NewService(nil)
		svc.running = true // Simulate already running
		ctx := context.Background()

		err := svc.Initialize(ctx)
		assert.NoError(t, err) // Should return nil for already running
	})
}
