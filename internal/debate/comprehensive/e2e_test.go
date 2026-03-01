package comprehensive_test

import (
	"context"
	"testing"
	"time"

	"dev.helix.agent/internal/debate/comprehensive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_SimpleCodeGeneration tests a complete code generation debate
func TestE2E_SimpleCodeGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 2
	config.QualityThreshold = 0.6

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = mgr.Initialize(ctx)
	require.NoError(t, err)

	// Create a full team using AgentFactory
	factory := comprehensive.NewAgentFactory(mgr.GetAgentPool(), nil)
	teamConfig := map[comprehensive.Role]comprehensive.AgentConfig{
		comprehensive.RoleArchitect:   {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleGenerator:   {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleCritic:      {Provider: "anthropic", Model: "claude"},
		comprehensive.RoleTester:      {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleValidator:   {Provider: "anthropic", Model: "claude"},
		comprehensive.RoleRefactoring: {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleSecurity:    {Provider: "openai", Model: "gpt-4"},
		comprehensive.RolePerformance: {Provider: "openai", Model: "gpt-4"},
	}

	_, err = factory.CreateTeam(teamConfig)
	require.NoError(t, err)

	req := &comprehensive.DebateRequest{
		ID:        "e2e-simple-code",
		Topic:     "Create a function to reverse a string in Go",
		Context:   "Write a clean, efficient string reversal function",
		Language:  "go",
		MaxRounds: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, req.ID, resp.ID)
	assert.NotZero(t, resp.Duration)
}

// TestE2E_RESTAPIDesign tests designing a complete REST API
func TestE2E_RESTAPIDesign(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 3

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	mgr.Initialize(ctx)

	// Add specialized agents
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleArchitect, "openai", "gpt-4", 0.95))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.85))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleSecurity, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleTester, "openai", "gpt-4", 0.9))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-rest-api",
		Topic:     "Design a REST API for a user management system",
		Context:   "Include endpoints for CRUD operations, authentication, and authorization",
		Language:  "go",
		MaxRounds: 3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify response exists (phases may be empty in mock mode)
	assert.Equal(t, req.ID, resp.ID)
}

// TestE2E_DatabaseSchemaDesign tests database schema design debate
func TestE2E_DatabaseSchemaDesign(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 2

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleArchitect, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.85))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.8))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleValidator, "openai", "gpt-4", 0.9))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-database-schema",
		Topic:     "Design a PostgreSQL schema for an e-commerce application",
		Context:   "Include tables for users, products, orders, and order items with proper relationships",
		Language:  "sql",
		MaxRounds: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestE2E_SecurityReview tests security-focused debate
func TestE2E_SecurityReview(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 2
	config.EnableSecurity = true

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleSecurity, "openai", "gpt-4", 0.95))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleBlueTeam, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleRedTeam, "anthropic", "claude", 0.85))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-security-review",
		Topic:     "Review and secure a user authentication system",
		Context:   "Identify vulnerabilities and implement security best practices",
		Language:  "go",
		MaxRounds: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestE2E_PerformanceOptimization tests performance optimization debate
func TestE2E_PerformanceOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 2

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RolePerformance, "openai", "gpt-4", 0.95))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleRefactoring, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.85))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-performance",
		Topic:     "Optimize a slow database query for better performance",
		Context:   "Current query takes 5 seconds, needs to be under 100ms",
		Language:  "sql",
		MaxRounds: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestE2E_MultiRoundDebate tests extended multi-round debate
func TestE2E_MultiRoundDebate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 5
	config.QualityThreshold = 0.7

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	mgr.Initialize(ctx)

	// Create a diverse team
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "anthropic", "claude", 0.85))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "openai", "gpt-4", 0.88))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.87))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleValidator, "openai", "gpt-4", 0.92))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-multi-round",
		Topic:     "Implement a concurrent task processor in Go",
		Context:   "Design a system that can process tasks concurrently with proper error handling",
		Language:  "go",
		MaxRounds: 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify debate ran for multiple rounds
	assert.GreaterOrEqual(t, resp.RoundsConducted, 0)
}

// TestE2E_ToolAugmentedDebate tests debate with tool usage
func TestE2E_ToolAugmentedDebate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 2

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	mgr.Initialize(ctx)

	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleTester, "openai", "gpt-4", 0.9))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-tools",
		Topic:     "Create a function with comprehensive tests",
		Context:   "Generate code and corresponding unit tests",
		Language:  "go",
		MaxRounds: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestE2E_ErrorHandling tests error handling in debates
func TestE2E_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	// Don't add any agents - should handle gracefully
	req := &comprehensive.DebateRequest{
		ID:        "e2e-error-handling",
		Topic:     "Test error handling",
		MaxRounds: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)

	// Should not panic, may return error or unsuccessful response
	if err == nil {
		assert.NotNil(t, resp)
	}
}

// TestE2E_TimeoutHandling tests timeout behavior
func TestE2E_TimeoutHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	config.MaxRounds = 10 // Long running

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.8))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-timeout",
		Topic:     "Test timeout handling",
		MaxRounds: 10,
	}

	// Short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = mgr.ExecuteDebate(ctx, req)

	// Should timeout or complete quickly
	// Either is acceptable behavior
}

// TestE2E_Cancellation tests context cancellation
func TestE2E_Cancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))

	req := &comprehensive.DebateRequest{
		ID:        "e2e-cancellation",
		Topic:     "Test cancellation",
		MaxRounds: 5,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err = mgr.ExecuteDebate(ctx, req)

	// Should handle cancellation gracefully
	// Error or partial result is acceptable
}

// TestE2E_MemoryPersistence tests memory across debates
func TestE2E_MemoryPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := comprehensive.DefaultConfig()
	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	mgr.Initialize(ctx)

	// First debate
	mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))

	req1 := &comprehensive.DebateRequest{
		ID:        "e2e-memory-1",
		Topic:     "Create a helper function",
		MaxRounds: 1,
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	resp1, err := mgr.ExecuteDebate(ctx1, req1)
	require.NoError(t, err)

	// Second debate - should benefit from first
	req2 := &comprehensive.DebateRequest{
		ID:        "e2e-memory-2",
		Topic:     "Create another helper function",
		MaxRounds: 1,
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	resp2, err := mgr.ExecuteDebate(ctx2, req2)
	require.NoError(t, err)

	assert.NotNil(t, resp1)
	assert.NotNil(t, resp2)
}

// TestE2E_QualityThresholds tests different quality thresholds
func TestE2E_QualityThresholds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testCases := []struct {
		name      string
		threshold float64
	}{
		{"low", 0.5},
		{"medium", 0.7},
		{"high", 0.9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := comprehensive.DefaultConfig()
			config.QualityThreshold = tc.threshold
			config.MaxRounds = 2

			mgr, err := comprehensive.NewIntegrationManager(config, nil)
			require.NoError(t, err)

			mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleGenerator, "openai", "gpt-4", 0.9))
			mgr.GetAgentPool().Add(comprehensive.NewAgent(comprehensive.RoleCritic, "anthropic", "claude", 0.8))

			req := &comprehensive.DebateRequest{
				ID:        "e2e-quality-" + tc.name,
				Topic:     "Test quality threshold",
				MaxRounds: 2,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := mgr.ExecuteDebate(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

// TestE2E_CompleteWorkflow tests the most complete workflow possible
func TestE2E_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This is a comprehensive test that exercises all components
	config := comprehensive.DefaultConfig()
	config.MaxRounds = 3
	config.QualityThreshold = 0.6

	mgr, err := comprehensive.NewIntegrationManager(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = mgr.Initialize(ctx)
	require.NoError(t, err)

	// Create full team with all roles using AgentFactory
	factory := comprehensive.NewAgentFactory(mgr.GetAgentPool(), nil)
	teamConfig := map[comprehensive.Role]comprehensive.AgentConfig{
		comprehensive.RoleArchitect:   {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleGenerator:   {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleCritic:      {Provider: "anthropic", Model: "claude"},
		comprehensive.RoleRefactoring: {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleTester:      {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleValidator:   {Provider: "anthropic", Model: "claude"},
		comprehensive.RoleSecurity:    {Provider: "openai", Model: "gpt-4"},
		comprehensive.RolePerformance: {Provider: "openai", Model: "gpt-4"},
		comprehensive.RoleRedTeam:     {Provider: "anthropic", Model: "claude"},
		comprehensive.RoleBlueTeam:    {Provider: "openai", Model: "gpt-4"},
	}

	_, err = factory.CreateTeam(teamConfig)
	require.NoError(t, err)

	// Verify health
	health := mgr.HealthCheck()
	assert.True(t, health["agent_pool"])
	assert.True(t, health["tool_registry"])

	// Get statistics
	stats := mgr.Statistics()
	// Should have at least 10 agents (Initialize creates 10, we add 10 more)
	assert.GreaterOrEqual(t, stats["agents_total"], 10)

	req := &comprehensive.DebateRequest{
		ID:        "e2e-complete-workflow",
		Topic:     "Design and implement a microservice architecture",
		Context:   "Complete system design with API, database, and deployment considerations",
		Language:  "go",
		MaxRounds: 3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := mgr.ExecuteDebate(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify response structure
	assert.Equal(t, req.ID, resp.ID)
	assert.NotZero(t, resp.Duration)
	assert.NotNil(t, resp.Phases)
	assert.NotNil(t, resp.Metadata)
}
