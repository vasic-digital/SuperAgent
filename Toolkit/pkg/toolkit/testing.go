// Package toolkit provides testing utilities for provider/agent combinations.
package toolkit

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestProvider defines an interface for testable providers.
type TestProvider interface {
	Provider
	// MockResponse allows setting mock responses for testing
	MockResponse(method string, response interface{}) error
	// ResetMocks clears all mock responses
	ResetMocks()
}

// TestAgent defines an interface for testable agents.
type TestAgent interface {
	Agent
	// MockExecute allows setting mock execution results for testing
	MockExecute(task string, result string) error
	// ResetMocks clears all mock results
	ResetMocks()
}

// TestSuite represents a collection of tests for a provider/agent combination.
type TestSuite struct {
	Name     string
	Provider TestProvider
	Agent    TestAgent
	Tests    []TestCase
}

// TestCase represents a single test case.
type TestCase struct {
	Name        string
	Description string
	Setup       func() error
	Execute     func(ctx context.Context) error
	Assert      func() error
	Cleanup     func() error
	Timeout     time.Duration
}

// TestRunner runs test suites.
type TestRunner struct {
	suites []TestSuite
}

// NewTestRunner creates a new TestRunner.
func NewTestRunner() *TestRunner {
	return &TestRunner{
		suites: make([]TestSuite, 0),
	}
}

// AddSuite adds a test suite to the runner.
func (r *TestRunner) AddSuite(suite TestSuite) {
	r.suites = append(r.suites, suite)
}

// Run executes all test suites.
func (r *TestRunner) Run(t *testing.T) {
	for _, suite := range r.suites {
		t.Run(suite.Name, func(t *testing.T) {
			r.runSuite(t, suite)
		})
	}
}

func (r *TestRunner) runSuite(t *testing.T, suite TestSuite) {
	// Reset mocks before each suite
	if suite.Provider != nil {
		suite.Provider.ResetMocks()
	}
	if suite.Agent != nil {
		suite.Agent.ResetMocks()
	}

	for _, testCase := range suite.Tests {
		t.Run(testCase.Name, func(t *testing.T) {
			r.runTestCase(t, testCase)
		})
	}
}

func (r *TestRunner) runTestCase(t *testing.T, testCase TestCase) {
	ctx := context.Background()
	if testCase.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, testCase.Timeout)
		defer cancel()
	}

	// Setup
	if testCase.Setup != nil {
		if err := testCase.Setup(); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// Execute
	if testCase.Execute != nil {
		if err := testCase.Execute(ctx); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	}

	// Assert
	if testCase.Assert != nil {
		if err := testCase.Assert(); err != nil {
			t.Fatalf("Assert failed: %v", err)
		}
	}

	// Cleanup
	if testCase.Cleanup != nil {
		if err := testCase.Cleanup(); err != nil {
			t.Errorf("Cleanup failed: %v", err)
		}
	}
}

// IntegrationTestHelper provides utilities for integration testing.
type IntegrationTestHelper struct {
	ProviderRegistry *ProviderRegistry
	AgentRegistry    *AgentRegistry
	ConfigBuilder    *GenericConfigBuilder
}

// NewIntegrationTestHelper creates a new IntegrationTestHelper.
func NewIntegrationTestHelper() *IntegrationTestHelper {
	return &IntegrationTestHelper{
		ProviderRegistry: NewProviderRegistry(),
		AgentRegistry:    NewAgentRegistry(),
		ConfigBuilder:    NewGenericConfigBuilder(),
	}
}

// SetupProvider registers a provider for testing.
func (h *IntegrationTestHelper) SetupProvider(name string, provider Provider) error {
	return h.ProviderRegistry.Register(name, provider)
}

// SetupAgent registers an agent for testing.
func (h *IntegrationTestHelper) SetupAgent(name string, agent Agent) error {
	return h.AgentRegistry.Register(name, agent)
}

// SetupConfigBuilder registers a config builder.
func (h *IntegrationTestHelper) SetupConfigBuilder(agentType string, builder ConfigBuilderFunc) {
	h.ConfigBuilder.Register(agentType, builder)
}

// TestProviderAgentCombination tests a provider/agent combination.
func (h *IntegrationTestHelper) TestProviderAgentCombination(t *testing.T, providerName, agentName, agentType string, config map[string]interface{}) {
	provider, ok := h.ProviderRegistry.Get(providerName)
	if !ok {
		t.Fatalf("Provider %s not found", providerName)
	}

	agent, ok := h.AgentRegistry.Get(agentName)
	if !ok {
		t.Fatalf("Agent %s not found", agentName)
	}

	// Build config
	builtConfig, err := h.ConfigBuilder.Build(agentType, config)
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}

	// Test basic functionality
	ctx := context.Background()

	// Test model discovery
	models, err := provider.DiscoverModels(ctx)
	if err != nil {
		t.Errorf("Model discovery failed: %v", err)
	} else if len(models) == 0 {
		t.Log("Warning: No models discovered")
	}

	// Test agent execution
	result, err := agent.Execute(ctx, "test task", builtConfig)
	if err != nil {
		t.Errorf("Agent execution failed: %v", err)
	} else {
		t.Logf("Agent result: %s", result)
	}
}

// MockProvider provides a mock implementation of Provider for testing.
type MockProvider struct {
	name        string
	chatResp    *ChatResponse
	embedResp   *EmbeddingResponse
	rerankResp  *RerankResponse
	models      []ModelInfo
	chatErr     error
	embedErr    error
	rerankErr   error
	discoverErr error
	validateErr error
}

// NewMockProvider creates a new MockProvider.
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name: name,
		models: []ModelInfo{
			{
				ID:       "test-model",
				Name:     "Test Model",
				Category: CategoryChat,
				Capabilities: ModelCapabilities{
					SupportsChat: true,
					MaxTokens:    4096,
				},
				Provider: name,
			},
		},
	}
}

// Name returns the provider name.
func (m *MockProvider) Name() string {
	return m.name
}

// Chat returns a mock chat response.
func (m *MockProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if m.chatErr != nil {
		return ChatResponse{}, m.chatErr
	}
	if m.chatResp != nil {
		return *m.chatResp, nil
	}
	return ChatResponse{
		Choices: []ChatChoice{
			{
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Mock response",
				},
			},
		},
	}, nil
}

// Embed returns a mock embedding response.
func (m *MockProvider) Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResponse, error) {
	if m.embedErr != nil {
		return EmbeddingResponse{}, m.embedErr
	}
	if m.embedResp != nil {
		return *m.embedResp, nil
	}
	return EmbeddingResponse{
		Data: []EmbeddingData{
			{
				Embedding: []float64{0.1, 0.2, 0.3},
			},
		},
	}, nil
}

// Rerank returns a mock rerank response.
func (m *MockProvider) Rerank(ctx context.Context, req RerankRequest) (RerankResponse, error) {
	if m.rerankErr != nil {
		return RerankResponse{}, m.rerankErr
	}
	if m.rerankResp != nil {
		return *m.rerankResp, nil
	}
	return RerankResponse{
		Results: []RerankResult{
			{Index: 0, Score: 0.9},
		},
	}, nil
}

// DiscoverModels returns mock models.
func (m *MockProvider) DiscoverModels(ctx context.Context) ([]ModelInfo, error) {
	if m.discoverErr != nil {
		return nil, m.discoverErr
	}
	return m.models, nil
}

// ValidateConfig validates the config.
func (m *MockProvider) ValidateConfig(config map[string]interface{}) error {
	return m.validateErr
}

// SetChatResponse sets the mock chat response.
func (m *MockProvider) SetChatResponse(resp ChatResponse) {
	m.chatResp = &resp
}

// SetChatError sets the mock chat error.
func (m *MockProvider) SetChatError(err error) {
	m.chatErr = err
}

// MockAgent provides a mock implementation of Agent for testing.
type MockAgent struct {
	name         string
	result       string
	executeErr   error
	validateErr  error
	capabilities []string
}

// NewMockAgent creates a new MockAgent.
func NewMockAgent(name string) *MockAgent {
	return &MockAgent{
		name:         name,
		result:       "Mock agent result",
		capabilities: []string{"chat", "code"},
	}
}

// Name returns the agent name.
func (m *MockAgent) Name() string {
	return m.name
}

// Execute returns a mock execution result.
func (m *MockAgent) Execute(ctx context.Context, task string, config interface{}) (string, error) {
	if m.executeErr != nil {
		return "", m.executeErr
	}
	return fmt.Sprintf("%s: %s", m.result, task), nil
}

// ValidateConfig validates the config.
func (m *MockAgent) ValidateConfig(config interface{}) error {
	return m.validateErr
}

// Capabilities returns the agent capabilities.
func (m *MockAgent) Capabilities() []string {
	return m.capabilities
}

// SetExecuteResult sets the mock execution result.
func (m *MockAgent) SetExecuteResult(result string) {
	m.result = result
}

// SetExecuteError sets the mock execution error.
func (m *MockAgent) SetExecuteError(err error) {
	m.executeErr = err
}
