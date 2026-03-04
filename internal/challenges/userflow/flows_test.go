package userflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"digital.vasic.challenges/pkg/challenge"
	uf "digital.vasic.challenges/pkg/userflow"
)

func TestHealthCheckFlow(t *testing.T) {
	flow := HealthCheckFlow()
	assert.NotEmpty(t, flow.Steps)
	assert.GreaterOrEqual(t, len(flow.Steps), 3,
		"should have at least 3 health check steps")

	for _, step := range flow.Steps {
		assert.NotEmpty(t, step.Name)
		assert.Equal(t, "GET", step.Method)
		assert.NotEmpty(t, step.Path)
	}
}

func TestProviderDiscoveryFlow(t *testing.T) {
	flow := ProviderDiscoveryFlow("test-token")
	assert.NotEmpty(t, flow.Steps)
	assert.GreaterOrEqual(t, len(flow.Steps), 4)

	// First step should be list models
	assert.Equal(t, "list_models", flow.Steps[0].Name)
	assert.Equal(t, "/v1/models", flow.Steps[0].Path)
}

func TestChatCompletionFlow(t *testing.T) {
	flow := ChatCompletionFlow()
	require.Len(t, flow.Steps, 1)

	step := flow.Steps[0]
	assert.Equal(t, "POST", step.Method)
	assert.Equal(t, "/v1/chat/completions", step.Path)
	assert.NotEmpty(t, step.Body)
	assert.Contains(t, step.Body, "messages")
}

func TestStreamingCompletionFlow(t *testing.T) {
	flow := StreamingCompletionFlow()
	require.Len(t, flow.Steps, 1)

	step := flow.Steps[0]
	assert.Contains(t, step.Body, `"stream": true`)
}

func TestEmbeddingsFlow(t *testing.T) {
	flow := EmbeddingsFlow()
	require.Len(t, flow.Steps, 1)
	assert.Equal(t, "/v1/embeddings/generate", flow.Steps[0].Path)
}

func TestFormattersFlow(t *testing.T) {
	flow := FormattersFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 2)

	assert.Equal(t, "/v1/formatters", flow.Steps[0].Path)
	assert.Equal(t, "GET", flow.Steps[0].Method)

	assert.Equal(t, "/v1/format", flow.Steps[1].Path)
	assert.Equal(t, "POST", flow.Steps[1].Method)
}

func TestDebateFlow(t *testing.T) {
	flow := DebateFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 2)
	assert.Contains(t, flow.Steps[0].Body, "topic")
}

func TestMonitoringFlow(t *testing.T) {
	flow := MonitoringFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 5)

	paths := make([]string, len(flow.Steps))
	for i, s := range flow.Steps {
		paths[i] = s.Path
	}
	assert.Contains(t, paths,
		"/v1/monitoring/status")
	assert.Contains(t, paths,
		"/v1/monitoring/circuit-breakers")
	assert.Contains(t, paths,
		"/v1/monitoring/provider-health")
}

func TestMCPProtocolFlow(t *testing.T) {
	flow := MCPProtocolFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 2)
}

func TestRAGFlow(t *testing.T) {
	flow := RAGFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 2)
}

func TestFeatureFlagsFlow(t *testing.T) {
	flow := FeatureFlagsFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 1)
	assert.Equal(t, "/v1/features", flow.Steps[0].Path)
}

func TestAuthenticationFlow(t *testing.T) {
	flow := AuthenticationFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 5,
		"should have at least 5 auth steps")

	assert.Equal(t, "login_valid", flow.Steps[0].Name)
	assert.Equal(t, "POST", flow.Steps[0].Method)
	assert.Equal(t, "/v1/auth/login", flow.Steps[0].Path)

	// Bad credentials step should be last.
	last := flow.Steps[len(flow.Steps)-1]
	assert.Equal(t, "login_bad_credentials", last.Name)
	assert.Contains(t, last.Body, "invalid")
}

func TestErrorHandlingFlow(t *testing.T) {
	flow := ErrorHandlingFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 5,
		"should have at least 5 error handling steps")

	// First step tests nonexistent model.
	assert.Equal(t,
		"nonexistent_model", flow.Steps[0].Name)
	assert.Contains(t, flow.Steps[0].Body,
		"no-such-model-xyz")

	// Second step tests invalid JSON.
	assert.Equal(t, "invalid_json", flow.Steps[1].Name)

	// Should include a 404 endpoint test.
	paths := make([]string, len(flow.Steps))
	for i, s := range flow.Steps {
		paths[i] = s.Path
	}
	assert.Contains(t, paths,
		"/v1/nonexistent-endpoint")
}

func TestConcurrentUsersFlow(t *testing.T) {
	flow := ConcurrentUsersFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 5,
		"should have at least 5 concurrent steps")

	// Starts with baseline health.
	assert.Equal(t, "baseline_health",
		flow.Steps[0].Name)
	assert.Equal(t, "/health", flow.Steps[0].Path)

	// Ends with post-load health check.
	last := flow.Steps[len(flow.Steps)-1]
	assert.Equal(t, "post_load_health", last.Name)
	assert.Equal(t, "/v1/health", last.Path)
}

func TestMultiTurnConversationFlow(t *testing.T) {
	flow := MultiTurnConversationFlow()
	require.Len(t, flow.Steps, 3,
		"should have 3 multi-turn steps")

	assert.Equal(t, "initial_message",
		flow.Steps[0].Name)
	assert.Equal(t, "follow_up",
		flow.Steps[1].Name)
	assert.Equal(t, "summarize",
		flow.Steps[2].Name)

	for _, step := range flow.Steps {
		assert.Equal(t, "POST", step.Method)
		assert.Equal(t,
			"/v1/chat/completions", step.Path)
		assert.NotEmpty(t, step.Body)
		assert.Contains(t, step.AcceptedStatuses, 200)
		assert.Contains(t, step.AcceptedStatuses, 501)
	}
}

func TestToolCallingFlow(t *testing.T) {
	flow := ToolCallingFlow()
	require.Len(t, flow.Steps, 2,
		"should have 2 tool calling steps")

	assert.Equal(t, "tool_choice_call",
		flow.Steps[0].Name)
	assert.Contains(t, flow.Steps[0].Body,
		"tool_choice")
	assert.Contains(t, flow.Steps[0].Body,
		"get_weather")

	assert.Equal(t, "legacy_function_call",
		flow.Steps[1].Name)
	assert.Contains(t, flow.Steps[1].Body,
		"function_call")
	assert.Contains(t, flow.Steps[1].Body,
		"calculate")

	for _, step := range flow.Steps {
		assert.Equal(t, "POST", step.Method)
		assert.Contains(t,
			step.AcceptedStatuses, 200)
		assert.Contains(t,
			step.AcceptedStatuses, 400)
		assert.Contains(t,
			step.AcceptedStatuses, 501)
	}
}

func TestProviderFailoverFlow(t *testing.T) {
	flow := ProviderFailoverFlow()
	require.Len(t, flow.Steps, 4,
		"should have 4 failover steps")

	assert.Equal(t, "fallback_chain",
		flow.Steps[0].Name)
	assert.Equal(t, "GET", flow.Steps[0].Method)
	assert.Equal(t,
		"/v1/monitoring/fallback-chain",
		flow.Steps[0].Path)

	assert.Equal(t, "circuit_breakers",
		flow.Steps[1].Name)
	assert.Equal(t,
		"/v1/monitoring/circuit-breakers",
		flow.Steps[1].Path)

	assert.Equal(t, "nonexistent_model_failover",
		flow.Steps[2].Name)
	assert.Equal(t, "POST", flow.Steps[2].Method)
	assert.Contains(t, flow.Steps[2].Body,
		"nonexistent-provider/fake-model")

	assert.Equal(t, "post_failover_health",
		flow.Steps[3].Name)
	assert.Equal(t,
		"/v1/monitoring/provider-health",
		flow.Steps[3].Path)
}

func TestFullSystemFlow(t *testing.T) {
	flow := FullSystemFlow()
	assert.GreaterOrEqual(t, len(flow.Steps), 7,
		"full system flow should cover all phases")

	// Verify it starts with health and ends with completion.
	assert.Equal(t, "/health", flow.Steps[0].Path)
	lastStep := flow.Steps[len(flow.Steps)-1]
	assert.Equal(t, "/v1/chat/completions", lastStep.Path)
}

func TestChallengeConstructors_NotNil(t *testing.T) {
	var adapter mockAPIAdapter
	healthDep := []challenge.ID{"helix-health-check"}
	providerDep := []challenge.ID{
		"helix-provider-discovery",
	}
	completionDep := []challenge.ID{
		"helix-chat-completion",
	}
	embeddingsDep := []challenge.ID{
		"helix-embeddings",
	}

	tests := []struct {
		name       string
		expectedID string
		fn         func() interface{}
	}{
		{"HealthCheck", "helix-health-check",
			func() interface{} {
				return NewHealthCheckChallenge(&adapter)
			}},
		{"FeatureFlags", "helix-feature-flags",
			func() interface{} {
				return NewFeatureFlagsChallenge(&adapter)
			}},
		{"ProviderDiscovery",
			"helix-provider-discovery",
			func() interface{} {
				return NewProviderDiscoveryChallenge(
					&adapter, healthDep,
				)
			}},
		{"Monitoring", "helix-monitoring",
			func() interface{} {
				return NewMonitoringChallenge(
					&adapter, healthDep,
				)
			}},
		{"Formatters", "helix-formatters",
			func() interface{} {
				return NewFormattersChallenge(
					&adapter, healthDep,
				)
			}},
		{"ChatCompletion", "helix-chat-completion",
			func() interface{} {
				return NewChatCompletionChallenge(
					&adapter, providerDep,
				)
			}},
		{"StreamingCompletion",
			"helix-streaming-completion",
			func() interface{} {
				return NewStreamingCompletionChallenge(
					&adapter, completionDep,
				)
			}},
		{"Embeddings", "helix-embeddings",
			func() interface{} {
				return NewEmbeddingsChallenge(
					&adapter, providerDep,
				)
			}},
		{"Debate", "helix-debate",
			func() interface{} {
				return NewDebateChallenge(
					&adapter, completionDep,
				)
			}},
		{"MCPProtocol", "helix-mcp-protocol",
			func() interface{} {
				return NewMCPChallenge(
					&adapter, healthDep,
				)
			}},
		{"RAG", "helix-rag",
			func() interface{} {
				return NewRAGChallenge(
					&adapter, embeddingsDep,
				)
			}},
		{"Authentication", "helix-authentication",
			func() interface{} {
				return NewAuthenticationChallenge(
					&adapter, healthDep,
				)
			}},
		{"ErrorHandling", "helix-error-handling",
			func() interface{} {
				return NewErrorHandlingChallenge(
					&adapter, healthDep,
				)
			}},
		{"ConcurrentUsers",
			"helix-concurrent-users",
			func() interface{} {
				return NewConcurrentUsersChallenge(
					&adapter, healthDep,
				)
			}},
		{"FullSystem", "helix-full-system",
			func() interface{} {
				return NewFullSystemChallenge(&adapter)
			}},
		{"MultiTurnConversation",
			"helix-multi-turn",
			func() interface{} {
				return NewMultiTurnConversationChallenge(
					&adapter, completionDep,
				)
			}},
		{"ToolCalling", "helix-tool-calling",
			func() interface{} {
				return NewToolCallingChallenge(
					&adapter, completionDep,
				)
			}},
		{"ProviderFailover",
			"helix-provider-failover",
			func() interface{} {
				return NewProviderFailoverChallenge(
					&adapter, providerDep,
				)
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			require.NotNil(t, result)
			c, ok := result.(challenge.Challenge)
			require.True(t, ok,
				"must implement challenge.Challenge")
			assert.Equal(t,
				challenge.ID(tt.expectedID),
				c.ID(),
			)
		})
	}
}

func TestOrchestratorConstruction(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	assert.NotNil(t, o)
	assert.Equal(t, 18, o.ChallengeCount(),
		"should register 18 challenges")
}

func TestOrchestratorListChallenges(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	ids := o.ListChallenges()
	assert.Len(t, ids, 18)

	// Verify key challenges are present.
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	assert.True(t, idSet["helix-health-check"])
	assert.True(t, idSet["helix-chat-completion"])
	assert.True(t, idSet["helix-full-system"])
}

func TestOrchestratorSummary(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	summary := o.Summary()
	assert.Contains(t, summary, "18 challenges")
	assert.Contains(t, summary, "localhost:7061")
}

// mockAPIAdapter is a minimal mock for constructor testing.
type mockAPIAdapter struct{}

func (m *mockAPIAdapter) Login(
	_ context.Context, _ uf.Credentials,
) (string, error) {
	return "", nil
}
func (m *mockAPIAdapter) LoginWithRetry(
	_ context.Context, _ uf.Credentials, _ int,
) (string, error) {
	return "", nil
}
func (m *mockAPIAdapter) Get(
	_ context.Context, _ string,
) (int, map[string]interface{}, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) GetRaw(
	_ context.Context, _ string,
) (int, []byte, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) GetArray(
	_ context.Context, _ string,
) (int, []interface{}, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) PostJSON(
	_ context.Context, _, _ string,
) (int, []byte, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) PutJSON(
	_ context.Context, _, _ string,
) (int, []byte, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) Delete(
	_ context.Context, _ string,
) (int, []byte, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) DeleteWithBody(
	_ context.Context, _, _ string,
) (int, []byte, error) {
	return 200, nil, nil
}
func (m *mockAPIAdapter) WebSocketConnect(
	_ context.Context, _ string,
) (uf.WebSocketConn, error) {
	return nil, nil
}
func (m *mockAPIAdapter) SetToken(_ string) {}
func (m *mockAPIAdapter) Available(
	_ context.Context,
) bool {
	return false
}

func TestOrchestrator_Challenges(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	challenges := o.Challenges()
	require.Len(t, challenges, 18,
		"Challenges() must return all 18 challenges")

	for _, c := range challenges {
		_, ok := c.(challenge.Challenge)
		assert.True(t, ok,
			"challenge %s must implement Challenge",
			c.ID(),
		)
	}
}

func TestOrchestrator_ChallengeCategories(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	challenges := o.Challenges()

	for _, c := range challenges {
		assert.Equal(t, "api", c.Category(),
			"challenge %s category should be 'api'",
			c.ID(),
		)
	}
}

func TestSetCategory_Override(t *testing.T) {
	var adapter mockAPIAdapter
	c := NewHealthCheckChallenge(&adapter)
	assert.Equal(t, "api", c.Category(),
		"initial category should be 'api'")

	c.SetCategory("userflow")
	assert.Equal(t, "userflow", c.Category(),
		"category should be overridden to 'userflow'")
}

func TestOrchestrator_RunByID_NotFound(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	ctx := context.Background()

	_, err := o.RunByID(ctx, "nonexistent-id")
	require.Error(t, err,
		"RunByID with unknown ID must return error")
	assert.Contains(t, err.Error(),
		"failed to get challenge",
		"error should mention lookup failure")
}

func TestOrchestrator_RunAll_NoServer(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	// Must not panic with a cancelled context.
	results, err := o.RunAll(ctx)
	_ = results
	_ = err
}

func TestOrchestrator_RunByID_CancelledContext(
	t *testing.T,
) {
	o := NewOrchestrator("http://localhost:7061")

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	// Must not panic with a cancelled context.
	result, err := o.RunByID(
		ctx, "helix-health-check",
	)
	if err != nil {
		return
	}
	require.NotNil(t, result,
		"result must not be nil when err is nil")
}
