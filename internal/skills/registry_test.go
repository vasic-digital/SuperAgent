package skills

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestSkill(name, category string, triggers []string) *Skill {
	return &Skill{
		Name:           name,
		Category:       category,
		Description:    "Test skill: " + name,
		TriggerPhrases: triggers,
		Version:        "1.0.0",
		LoadedAt:       time.Now(),
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry(nil)

	skill := createTestSkill("test-skill", "test-category", []string{"test trigger"})
	registry.RegisterSkill(skill)

	// Get by name
	retrieved, ok := registry.Get("test-skill")
	require.True(t, ok)
	assert.Equal(t, skill.Name, retrieved.Name)
	assert.Equal(t, skill.Category, retrieved.Category)

	// Get non-existent
	_, ok = registry.Get("non-existent")
	assert.False(t, ok)
}

func TestRegistry_GetByCategory(t *testing.T) {
	registry := NewRegistry(nil)

	skill1 := createTestSkill("skill-1", "devops", []string{"trigger1"})
	skill2 := createTestSkill("skill-2", "devops", []string{"trigger2"})
	skill3 := createTestSkill("skill-3", "security", []string{"trigger3"})

	registry.RegisterSkill(skill1)
	registry.RegisterSkill(skill2)
	registry.RegisterSkill(skill3)

	// Get devops skills
	devopsSkills := registry.GetByCategory("devops")
	assert.Len(t, devopsSkills, 2)

	// Get security skills
	securitySkills := registry.GetByCategory("security")
	assert.Len(t, securitySkills, 1)

	// Get non-existent category
	nonExistent := registry.GetByCategory("non-existent")
	assert.Nil(t, nonExistent)
}

func TestRegistry_GetByTrigger(t *testing.T) {
	registry := NewRegistry(nil)

	skill1 := createTestSkill("skill-1", "devops", []string{"docker", "container"})
	skill2 := createTestSkill("skill-2", "devops", []string{"kubernetes", "container"})

	registry.RegisterSkill(skill1)
	registry.RegisterSkill(skill2)

	// Single skill trigger
	dockerSkills := registry.GetByTrigger("docker")
	assert.Len(t, dockerSkills, 1)
	assert.Equal(t, "skill-1", dockerSkills[0].Name)

	// Shared trigger
	containerSkills := registry.GetByTrigger("container")
	assert.Len(t, containerSkills, 2)
}

func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry(nil)

	skill := createTestSkill("removable-skill", "test", []string{"remove trigger"})
	registry.RegisterSkill(skill)

	// Verify it exists
	_, ok := registry.Get("removable-skill")
	require.True(t, ok)

	// Remove
	removed := registry.Remove("removable-skill")
	assert.True(t, removed)

	// Verify it's gone
	_, ok = registry.Get("removable-skill")
	assert.False(t, ok)

	// Try to remove non-existent
	removed = registry.Remove("non-existent")
	assert.False(t, removed)
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry(nil)

	skills := []*Skill{
		createTestSkill("skill-1", "cat1", []string{"t1"}),
		createTestSkill("skill-2", "cat2", []string{"t2"}),
		createTestSkill("skill-3", "cat3", []string{"t3"}),
	}

	for _, s := range skills {
		registry.RegisterSkill(s)
	}

	all := registry.GetAll()
	assert.Len(t, all, 3)
}

func TestRegistry_GetCategories(t *testing.T) {
	registry := NewRegistry(nil)

	registry.RegisterSkill(createTestSkill("s1", "devops", []string{"t1"}))
	registry.RegisterSkill(createTestSkill("s2", "security", []string{"t2"}))
	registry.RegisterSkill(createTestSkill("s3", "devops", []string{"t3"}))

	categories := registry.GetCategories()
	assert.Len(t, categories, 2)
	assert.Contains(t, categories, "devops")
	assert.Contains(t, categories, "security")
}

func TestRegistry_GetTriggers(t *testing.T) {
	registry := NewRegistry(nil)

	registry.RegisterSkill(createTestSkill("s1", "cat", []string{"trigger-a", "trigger-b"}))
	registry.RegisterSkill(createTestSkill("s2", "cat", []string{"trigger-c"}))

	triggers := registry.GetTriggers()
	assert.Len(t, triggers, 3)
	assert.Contains(t, triggers, "trigger-a")
	assert.Contains(t, triggers, "trigger-b")
	assert.Contains(t, triggers, "trigger-c")
}

func TestRegistry_Stats(t *testing.T) {
	registry := NewRegistry(nil)

	registry.RegisterSkill(createTestSkill("s1", "devops", []string{"t1", "t2"}))
	registry.RegisterSkill(createTestSkill("s2", "security", []string{"t3"}))

	stats := registry.Stats()
	assert.Equal(t, 2, stats.TotalSkills)
	assert.Equal(t, 3, stats.TotalTriggers)
	assert.Equal(t, 1, stats.SkillsByCategory["devops"])
	assert.Equal(t, 1, stats.SkillsByCategory["security"])
}

func TestRegistry_Search(t *testing.T) {
	registry := NewRegistry(nil)

	registry.RegisterSkill(createTestSkill("docker-compose-creator", "devops", []string{"docker compose"}))
	registry.RegisterSkill(createTestSkill("kubernetes-deployment", "devops", []string{"kubernetes", "k8s"}))
	registry.RegisterSkill(createTestSkill("security-scanner", "security", []string{"security scan"}))

	// Search by trigger
	results := registry.Search("docker")
	assert.GreaterOrEqual(t, len(results), 1)

	// Search by name
	results = registry.Search("kubernetes")
	assert.GreaterOrEqual(t, len(results), 1)

	// Search with no results
	results = registry.Search("nonexistent-skill-xyz")
	assert.Empty(t, results)
}

func TestRegistry_Concurrent(t *testing.T) {
	registry := NewRegistry(nil)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 100; i++ {
		skill := createTestSkill(
			"skill-"+string(rune('a'+i%26)),
			"category-"+string(rune('a'+i%5)),
			[]string{"trigger-" + string(rune('a'+i%26))},
		)
		registry.RegisterSkill(skill)
	}

	// Run concurrent operations
	done := make(chan bool, 10)

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = registry.GetAll()
				_ = registry.GetCategories()
				_ = registry.Stats()
			}
			done <- true
		}()
	}

	// Concurrent writes (with context, though not used in simple register)
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					skill := createTestSkill(
						"concurrent-skill-"+string(rune('0'+id))+"-"+string(rune('0'+j)),
						"concurrent-category",
						[]string{"concurrent-trigger"},
					)
					registry.RegisterSkill(skill)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify state is consistent
	stats := registry.Stats()
	assert.Greater(t, stats.TotalSkills, 0)
}
