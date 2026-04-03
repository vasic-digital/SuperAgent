package subagent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ProviderType: "openai",
				APIKey:       "test-key",
				Logger:       logger,
			},
			wantErr: false,
		},
		{
			name:    "nil config fails",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing logger uses default",
			config: &Config{
				ProviderType: "openai",
				APIKey:       "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestManager_CreateAgent(t *testing.T) {
	logger := zap.NewNop()
	manager, err := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		agentType string
		config    ProfileConfig
		wantErr   bool
	}{
		{
			name:      "create explore agent",
			agentType: "explore",
			config: ProfileConfig{
				Name:        "test-explorer",
				Model:       "gpt-4o-mini",
				MaxTokens:   2000,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name:      "create plan agent",
			agentType: "plan",
			config: ProfileConfig{
				Name:      "test-planner",
				Model:     "gpt-4o",
				MaxTokens: 3000,
			},
			wantErr: false,
		},
		{
			name:      "create general agent",
			agentType: "general",
			config: ProfileConfig{
				Name:      "test-general",
				Model:     "gpt-4o",
				MaxTokens: 4000,
				Tools:     []string{"file_read", "file_write"},
			},
			wantErr: false,
		},
		{
			name:      "invalid agent type",
			agentType: "invalid",
			config:    ProfileConfig{Name: "test"},
			wantErr:   true,
		},
		{
			name:      "missing name fails",
			agentType: "explore",
			config:    ProfileConfig{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := manager.CreateAgent(ctx, tt.agentType, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
				assert.Equal(t, tt.config.Name, agent.Name())
				assert.Equal(t, tt.agentType, agent.Type())
			}
		})
	}
}

func TestManager_ListAgents(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx := context.Background()

	// Initially empty
	agents := manager.ListAgents()
	assert.Empty(t, agents)

	// Create some agents
	manager.CreateAgent(ctx, "explore", ProfileConfig{Name: "agent1"})
	manager.CreateAgent(ctx, "plan", ProfileConfig{Name: "agent2"})
	manager.CreateAgent(ctx, "general", ProfileConfig{Name: "agent3"})

	agents = manager.ListAgents()
	assert.Len(t, agents, 3)

	// Check names
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name()
	}
	assert.Contains(t, names, "agent1")
	assert.Contains(t, names, "agent2")
	assert.Contains(t, names, "agent3")
}

func TestManager_GetAgent(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx := context.Background()
	manager.CreateAgent(ctx, "explore", ProfileConfig{Name: "test-agent"})

	// Get existing agent
	agent, err := manager.GetAgent("test-agent")
	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "test-agent", agent.Name())

	// Get non-existent agent
	agent, err = manager.GetAgent("non-existent")
	assert.Error(t, err)
	assert.Nil(t, agent)
}

func TestManager_Shutdown(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create agents
	manager.CreateAgent(ctx, "explore", ProfileConfig{Name: "agent1"})
	manager.CreateAgent(ctx, "plan", ProfileConfig{Name: "agent2"})

	// Shutdown
	err := manager.Shutdown(ctx)
	assert.NoError(t, err)

	// List should be empty after shutdown
	agents := manager.ListAgents()
	assert.Empty(t, agents)
}

func TestAgent_Execute(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx := context.Background()
	agent, _ := manager.CreateAgent(ctx, "explore", ProfileConfig{
		Name:      "test-agent",
		Model:     "gpt-4o-mini",
		MaxTokens: 1000,
	})

	task := Task{
		Description: "Test task",
		MaxSteps:    5,
		Timeout:     30 * time.Second,
	}

	result, err := agent.Execute(ctx, task)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotZero(t, result.Duration)
}

func TestAgent_CreatePlan(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx := context.Background()
	agent, _ := manager.CreateAgent(ctx, "plan", ProfileConfig{
		Name:      "planner",
		Model:     "gpt-4o",
		MaxTokens: 2000,
	})

	input := PlanInput{
		Objective: "Implement a feature",
		Discoveries: []string{
			"User needs authentication",
			"Must support OAuth",
		},
		Constraints: []string{
			"Use existing database",
			"JWT tokens only",
		},
	}

	plan, err := agent.CreatePlan(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.NotEmpty(t, plan.Steps)
}

func TestAgent_ExecutePlan(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx := context.Background()
	agent, _ := manager.CreateAgent(ctx, "general", ProfileConfig{
		Name:      "impl",
		Model:     "gpt-4o",
		MaxTokens: 3000,
	})

	plan := PlanResult{
		Steps: []PlanStep{
			{Description: "Create file", Priority: "high"},
			{Description: "Add code", Priority: "high"},
		},
		FilesToCreate: []string{"test.go"},
	}

	result, err := agent.ExecutePlan(ctx, plan)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgent_Cancel(t *testing.T) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})

	ctx := context.Background()
	agent, _ := manager.CreateAgent(ctx, "explore", ProfileConfig{
		Name: "cancellable",
	})

	// Start long-running task
	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		agent.Execute(taskCtx, Task{
			Description: "Long task",
			MaxSteps:    100,
		})
	}()

	// Cancel immediately
	time.Sleep(10 * time.Millisecond)
	err := agent.Cancel()
	assert.NoError(t, err)
}

func BenchmarkManager_CreateAgent(b *testing.B) {
	logger := zap.NewNop()
	manager, _ := NewManager(&Config{
		ProviderType: "openai",
		APIKey:       "test-key",
		Logger:       logger,
	})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateAgent(ctx, "explore", ProfileConfig{
			Name: fmt.Sprintf("agent-%d", i),
		})
	}
}
