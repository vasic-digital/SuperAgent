package comprehensive

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentPool(t *testing.T) {
	logger := logrus.New()
	pool := NewAgentPool(logger)

	assert.NotNil(t, pool)
	assert.Equal(t, 0, pool.Size())
}

func TestAgentPool_Add(t *testing.T) {
	pool := NewAgentPool(nil)
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)

	pool.Add(agent)

	assert.Equal(t, 1, pool.Size())

	// Verify we can retrieve it
	retrieved, ok := pool.Get(agent.ID)
	assert.True(t, ok)
	assert.Equal(t, agent.ID, retrieved.ID)
}

func TestAgentPool_GetByRole(t *testing.T) {
	pool := NewAgentPool(nil)

	gen1 := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	gen2 := NewAgent(RoleGenerator, "anthropic", "claude", 8.0)
	critic := NewAgent(RoleCritic, "openai", "gpt-4", 8.5)

	pool.Add(gen1)
	pool.Add(gen2)
	pool.Add(critic)

	generators := pool.GetByRole(RoleGenerator)
	assert.Len(t, generators, 2)

	critics := pool.GetByRole(RoleCritic)
	assert.Len(t, critics, 1)
}

func TestAgentPool_SelectBestForRole(t *testing.T) {
	pool := NewAgentPool(nil)

	agent1 := NewAgent(RoleGenerator, "openai", "gpt-4", 7.5)
	agent2 := NewAgent(RoleGenerator, "anthropic", "claude", 8.5)
	agent3 := NewAgent(RoleGenerator, "google", "gemini", 8.0)

	pool.Add(agent1)
	pool.Add(agent2)
	pool.Add(agent3)

	best := pool.SelectBestForRole(RoleGenerator)
	assert.NotNil(t, best)
	assert.Equal(t, 8.5, best.Score)
}

func TestAgentPool_SelectTopNForRole(t *testing.T) {
	pool := NewAgentPool(nil)

	for i := 0; i < 5; i++ {
		agent := NewAgent(RoleGenerator, "openai", "gpt-4", float64(i)+5.0)
		pool.Add(agent)
	}

	top3 := pool.SelectTopNForRole(RoleGenerator, 3)
	assert.Len(t, top3, 3)

	// Should be sorted by score descending
	for i := 0; i < len(top3)-1; i++ {
		assert.GreaterOrEqual(t, top3[i].Score, top3[i+1].Score)
	}
}

func TestAgentPool_Remove(t *testing.T) {
	pool := NewAgentPool(nil)
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)

	pool.Add(agent)
	assert.Equal(t, 1, pool.Size())

	// Remove it
	removed := pool.Remove(agent.ID)
	assert.True(t, removed)
	assert.Equal(t, 0, pool.Size())

	// Try to remove again
	removed = pool.Remove(agent.ID)
	assert.False(t, removed)
}

func TestAgentPool_HasRole(t *testing.T) {
	pool := NewAgentPool(nil)

	assert.False(t, pool.HasRole(RoleGenerator))

	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	pool.Add(agent)

	assert.True(t, pool.HasRole(RoleGenerator))
	assert.False(t, pool.HasRole(RoleCritic))
}

func TestAgentPool_Clear(t *testing.T) {
	pool := NewAgentPool(nil)

	pool.Add(NewAgent(RoleGenerator, "openai", "gpt-4", 8.5))
	pool.Add(NewAgent(RoleCritic, "anthropic", "claude", 8.5))

	assert.Equal(t, 2, pool.Size())

	pool.Clear()

	assert.Equal(t, 0, pool.Size())
}

func TestAgentFactory_CreateAgent(t *testing.T) {
	pool := NewAgentPool(nil)
	factory := NewAgentFactory(pool, nil)

	agent, err := factory.CreateAgent(RoleGenerator, "openai", "gpt-4", 8.5)

	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, RoleGenerator, agent.Role)
	assert.Equal(t, "openai", agent.Provider)
	assert.Equal(t, "gpt-4", agent.Model)

	// Should be added to pool
	assert.Equal(t, 1, pool.Size())
}

func TestAgentFactory_CreateAgent_InvalidRole(t *testing.T) {
	pool := NewAgentPool(nil)
	factory := NewAgentFactory(pool, nil)

	_, err := factory.CreateAgent(Role("invalid"), "openai", "gpt-4", 8.5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestAgentFactory_CreateAgent_EmptyProvider(t *testing.T) {
	pool := NewAgentPool(nil)
	factory := NewAgentFactory(pool, nil)

	_, err := factory.CreateAgent(RoleGenerator, "", "gpt-4", 8.5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider cannot be empty")
}

func TestAgentFactory_CreateTeam(t *testing.T) {
	pool := NewAgentPool(nil)
	factory := NewAgentFactory(pool, nil)

	config := map[Role]AgentConfig{
		RoleGenerator: {Provider: "openai", Model: "gpt-4", Score: 8.5},
		RoleCritic:    {Provider: "anthropic", Model: "claude", Score: 8.0},
	}

	agents, err := factory.CreateTeam(config)

	assert.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Equal(t, 2, pool.Size())
}

func TestBaseAgent_GetID(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	base := NewBaseAgent(agent, nil, nil)

	assert.Equal(t, agent.ID, base.GetID())
}

func TestBaseAgent_GetRole(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	base := NewBaseAgent(agent, nil, nil)

	assert.Equal(t, RoleGenerator, base.GetRole())
}

func TestBaseAgent_GetCapabilities(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	base := NewBaseAgent(agent, nil, nil)

	caps := base.GetCapabilities()
	assert.NotEmpty(t, caps)
}

func TestBaseAgent_CanHandle(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	base := NewBaseAgent(agent, nil, nil)

	// Base implementation always returns true
	assert.True(t, base.CanHandle("anything"))
}

func TestBaseAgent_UpdateScore(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	base := NewBaseAgent(agent, nil, nil)

	base.UpdateScore(9.0)
	assert.Equal(t, 9.0, agent.Score)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkAgentPool_Add(b *testing.B) {
	pool := NewAgentPool(nil)
	agents := make([]*Agent, b.N)
	for i := 0; i < b.N; i++ {
		agents[i] = NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Add(agents[i])
	}
}

func BenchmarkAgentPool_Get(b *testing.B) {
	pool := NewAgentPool(nil)
	agents := make([]*Agent, 100)
	for i := 0; i < 100; i++ {
		agents[i] = NewAgent(RoleGenerator, "openai", "gpt-4", float64(i))
		pool.Add(agents[i])
	}
	targetID := agents[50].ID
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Get(targetID)
	}
}

func BenchmarkAgentPool_SelectBestForRole(b *testing.B) {
	pool := NewAgentPool(nil)
	roles := AllRoles()
	for i := 0; i < 50; i++ {
		role := roles[i%len(roles)]
		agent := NewAgent(role, "provider", "model", float64(i)+5.0)
		pool.Add(agent)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.SelectBestForRole(RoleGenerator)
	}
}

func BenchmarkAgentPool_SelectTopNForRole(b *testing.B) {
	pool := NewAgentPool(nil)
	for i := 0; i < 20; i++ {
		agent := NewAgent(RoleGenerator, "openai", "gpt-4", float64(i)+5.0)
		pool.Add(agent)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.SelectTopNForRole(RoleGenerator, 5)
	}
}

func BenchmarkNewAgent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	}
}
