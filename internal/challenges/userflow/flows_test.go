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
		{"FullSystem", "helix-full-system",
			func() interface{} {
				return NewFullSystemChallenge(&adapter)
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
	assert.Equal(t, 12, o.ChallengeCount(),
		"should register 12 challenges")
}

func TestOrchestratorListChallenges(t *testing.T) {
	o := NewOrchestrator("http://localhost:7061")
	ids := o.ListChallenges()
	assert.Len(t, ids, 12)

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
	assert.Contains(t, summary, "12 challenges")
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
