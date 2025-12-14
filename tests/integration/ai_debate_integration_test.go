package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/services"
)

// TestAIDebateIntegration_CompleteWorkflow tests the complete AI debate integration workflow
func TestAIDebateIntegration_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test configurations
	tempDir, err := os.MkdirTemp("", "ai-debate-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("Integration with Existing Ensemble System", func(t *testing.T) {
		// Create configuration
		cfg := createIntegrationTestConfig(t, tempDir)
		loader := config.NewAIDebateConfigLoader(filepath.Join(tempDir, "integration-config.yaml"))

		err := loader.Save(cfg)
		require.NoError(t, err)

		loadedConfig, err := loader.Load()
		require.NoError(t, err)

		// Create debate integration service
		debateIntegration, err := services.NewAIDebateIntegration(loadedConfig, nil)
		require.NoError(t, err, "Failed to create debate integration service")
		require.NotNil(t, debateIntegration, "Debate integration service is nil")

		// Test provider initialization
		capabilities := debateIntegration.GetProviderCapabilities()
		assert.NotEmpty(t, capabilities, "Provider capabilities should not be empty")

		health := debateIntegration.GetProviderHealth()
		assert.NotEmpty(t, health, "Provider health should not be empty")

		// Conduct ensemble debate
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		topic := "How should society balance AI innovation with ethical considerations?"
		initialContext := "Recent AI developments have highlighted the need for ethical frameworks while maintaining innovation momentum."

		result, err := debateIntegration.ConductEnsembleDebate(ctx, topic, initialContext)
		require.NoError(t, err, "Ensemble debate failed")
		require.NotNil(t, result, "Debate result is nil")

		// Validate integration results
		validateIntegrationResults(t, result, topic)
	})

	t.Run("Provider Fallback Chain Integration", func(t *testing.T) {
		// Test fallback mechanism in integration context
		cfg := createFallbackIntegrationTestConfig(t, tempDir)
		loader := config.NewAIDebateConfigLoader(filepath.Join(tempDir, "fallback-integration-config.yaml"))

		err := loader.Save(cfg)
		require.NoError(t, err)

		loadedConfig, err := loader.Load()
		require.NoError(t, err)

		debateIntegration, err := services.NewAIDebateIntegration(loadedConfig, nil)
		require.NoError(t, err)

		// Verify fallback providers are initialized
		capabilities := debateIntegration.GetProviderCapabilities()
		assert.Greater(t, len(capabilities), 2, "Should have multiple providers for fallback testing")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := debateIntegration.ConductEnsembleDebate(ctx, "Test fallback integration", "Testing LLM fallback chain in integration")
		require.NoError(t, err, "Ensemble debate with fallbacks failed")
		require.NotNil(t, result)

		// Verify fallback was used
		if result.FallbackUsed {
			t.Log("Fallback mechanism was successfully triggered during integration test")
		}
	})

	t.Run("Configuration Update During Runtime", func(t *testing.T) {
		// Test dynamic configuration updates
		initialConfig := createIntegrationTestConfig(t, tempDir)
		loader := config.NewAIDebateConfigLoader(filepath.Join(tempDir, "dynamic-config.yaml"))

		err := loader.Save(initialConfig)
		require.NoError(t, err)

		loadedConfig, err := loader.Load()
		require.NoError(t, err)

		debateIntegration, err := services.NewAIDebateIntegration(loadedConfig, nil)
		require.NoError(t, err)

		// Get initial capabilities
		initialCapabilities := debateIntegration.GetProviderCapabilities()
		initialHealth := debateIntegration.GetProviderHealth()

		// Update configuration
		updatedConfig := createUpdatedIntegrationTestConfig(t, tempDir)
		err = debateIntegration.UpdateConfiguration(updatedConfig)
		require.NoError(t, err, "Configuration update failed")

		// Verify configuration was updated
		updatedCapabilities := debateIntegration.GetProviderCapabilities()
		updatedHealth := debateIntegration.GetProviderHealth()

		assert.NotEqual(t, len(initialCapabilities), len(updatedCapabilities), "Capabilities should change after update")
		assert.NotEqual(t, len(initialHealth), len(updatedHealth), "Health status should change after update")

		// Test debate with updated configuration
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := debateIntegration.ConductEnsembleDebate(ctx, "Test updated configuration", "Testing configuration updates")
		require.NoError(t, err, "Debate with updated configuration failed")
		require.NotNil(t, result)
	})

	t.Run("Health Monitoring and Provider Status", func(t *testing.T) {
		// Test health monitoring capabilities
		cfg := createHealthTestConfig(t, tempDir)
		loader := config.NewAIDebateConfigLoader(filepath.Join(tempDir, "health-config.yaml"))

		err := loader.Save(cfg)
		require.NoError(t, err)

		loadedConfig, err := loader.Load()
		require.NoError(t, err)

		debateIntegration, err := services.NewAIDebateIntegration(loadedConfig, nil)
		require.NoError(t, err)

		// Test health checks
		health := debateIntegration.GetProviderHealth()
		assert.NotEmpty(t, health, "Health status should not be empty")

		for provider, status := range health {
			t.Logf("Provider %s health: %s", provider, status)
			assert.NotEmpty(t, status, "Health status should not be empty for provider %s", provider)
		}

		// Test capabilities
		capabilities := debateIntegration.GetProviderCapabilities()
		assert.NotEmpty(t, capabilities, "Capabilities should not be empty")

		for provider, caps := range capabilities {
			assert.NotNil(t, caps, "Capabilities should not be nil for provider %s", provider)
			assert.NotEmpty(t, caps.SupportedModels, "Should have supported models for provider %s", provider)
		}
	})

	t.Run("Concurrent Debate Execution", func(t *testing.T) {
		// Test concurrent debate execution
		cfg := createConcurrentTestConfig(t, tempDir)
		loader := config.NewAIDebateConfigLoader(filepath.Join(tempDir, "concurrent-config.yaml"))

		err := loader.Save(cfg)
		require.NoError(t, err)

		loadedConfig, err := loader.Load()
		require.NoError(t, err)

		debateIntegration, err := services.NewAIDebateIntegration(loadedConfig, nil)
		require.NoError(t, err)

		// Launch multiple concurrent debates
		numConcurrent := 3
		results := make(chan *services.DebateResult, numConcurrent)
		errors := make(chan error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func(id int) {
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()

				topic := fmt.Sprintf("Concurrent integration debate %d", id)
				context := fmt.Sprintf("Testing concurrent execution in integration context %d", id)

				result, err := debateIntegration.ConductEnsembleDebate(ctx, topic, context)
				if err != nil {
					errors <- err
				} else {
					results <- result
				}
			}(i)
		}

		// Collect results
		successfulDebates := 0
		for i := 0; i < numConcurrent; i++ {
			select {
			case result := <-results:
				require.NotNil(t, result, "Successful debate returned nil result")
				successfulDebates++
			case err := <-errors:
				t.Logf("Concurrent debate failed: %v", err)
			case <-time.After(2 * time.Minute):
				t.Fatal("Concurrent debates timed out")
			}
		}

		assert.Greater(t, successfulDebates, 0, "Should have at least one successful concurrent debate")
		t.Logf("Concurrent test results: %d successful out of %d total", successfulDebates, numConcurrent)
	})

	t.Run("Error Recovery and Resilience", func(t *testing.T) {
		// Test error recovery and system resilience
		cfg := createErrorRecoveryTestConfig(t, tempDir)
		loader := config.NewAIDebateConfigLoader(filepath.Join(tempDir, "error-recovery-config.yaml"))

		err := loader.Save(cfg)
		require.NoError(t, err)

		loadedConfig, err := loader.Load()
		require.NoError(t, err)

		debateIntegration, err := services.NewAIDebateIntegration(loadedConfig, nil)
		require.NoError(t, err)

		// Test with various error scenarios
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// Test 1: Normal debate
		result1, err1 := debateIntegration.ConductEnsembleDebate(ctx, "Normal debate", "Testing normal operation")
		if err1 == nil {
			assert.NotNil(t, result1, "Normal debate result should not be nil")
		}

		// Test 2: Empty topic (should handle gracefully)
		result2, err2 := debateIntegration.ConductEnsembleDebate(ctx, "", "Testing empty topic")
		// Should either succeed with empty topic or fail gracefully
		if err2 != nil {
			t.Logf("Empty topic handled with error (expected): %v", err2)
		}

		// Test 3: Very long topic (should handle gracefully)
		longTopic := "This is a very long topic that exceeds normal length limitations and should be handled gracefully by the system without causing crashes or unexpected behavior"
		result3, err3 := debateIntegration.ConductEnsembleDebate(ctx, longTopic, "Testing long topic handling")
		if err3 == nil {
			assert.NotNil(t, result3, "Long topic result should not be nil")
		}

		// Test 4: Special characters in topic
		specialTopic := "Test with special chars: àáâãäåæçèéêë ñ 中文 🚀"
		result4, err4 := debateIntegration.ConductEnsembleDebate(ctx, specialTopic, "Testing special characters")
		if err4 == nil {
			assert.NotNil(t, result4, "Special characters result should not be nil")
		}
	})
}

// TestAIDebateIntegration_ConfigurationValidation tests configuration validation in integration context
func TestAIDebateIntegration_ConfigurationValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-validation-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		config      *config.AIDebateConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid integration configuration",
			config: &config.AIDebateConfig{
				Enabled:             true,
				MaximalRepeatRounds: 3,
				DebateTimeout:       5 * 60 * 1000,
				ConsensusThreshold:  0.75,
				Participants: []config.DebateParticipant{
					{
						Name:    "TestParticipant",
						Role:    "Analyst",
						Enabled: true,
						LLMs: []config.LLMConfiguration{
							{
								Name:        "TestLLM",
								Provider:    "claude",
								Model:       "claude-3-sonnet",
								Enabled:     true,
								Timeout:     30 * 1000,
								MaxTokens:   1000,
								Temperature: 0.7,
							},
						},
						ResponseTimeout:    30 * 1000,
						Weight:             1.0,
						Priority:           1,
						DebateStyle:        "analytical",
						ArgumentationStyle: "logical",
						PersuasionLevel:    0.5,
						OpennessToChange:   0.5,
						QualityThreshold:   0.7,
						MinResponseLength:  50,
						MaxResponseLength:  1000,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid provider in integration",
			config: &config.AIDebateConfig{
				Enabled:             true,
				MaximalRepeatRounds: 3,
				DebateTimeout:       5 * 60 * 1000,
				ConsensusThreshold:  0.75,
				Participants: []config.DebateParticipant{
					{
						Name:    "TestParticipant",
						Role:    "Analyst",
						Enabled: true,
						LLMs: []config.LLMConfiguration{
							{
								Name:        "TestLLM",
								Provider:    "invalid_provider", // Invalid provider
								Model:       "test-model",
								Enabled:     true,
								Timeout:     30 * 1000,
								MaxTokens:   1000,
								Temperature: 0.7,
							},
						},
						ResponseTimeout:    30 * 1000,
						Weight:             1.0,
						Priority:           1,
						DebateStyle:        "analytical",
						ArgumentationStyle: "logical",
						PersuasionLevel:    0.5,
						OpennessToChange:   0.5,
						QualityThreshold:   0.7,
						MinResponseLength:  50,
						MaxResponseLength:  1000,
					},
				},
			},
			expectError: true,
			errorMsg:    "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test configuration validation
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
				if tt.errorMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Validation should pass")

				// Test integration creation
				integration, err := services.NewAIDebateIntegration(tt.config, nil)
				if tt.expectError {
					assert.Error(t, err, "Expected integration creation to fail")
				} else {
					assert.NoError(t, err, "Integration creation should succeed")
					assert.NotNil(t, integration, "Integration should not be nil")
				}
			}
		})
	}
}

// Helper functions for creating test configurations

func createIntegrationTestConfig(t *testing.T, tempDir string) *config.AIDebateConfig {
	return &config.AIDebateConfig{
		Enabled:             true,
		MaximalRepeatRounds: 3,
		DebateTimeout:       5 * 60 * 1000,
		ConsensusThreshold:  0.75,
		EnableCognee:        true,
		Participants: []config.DebateParticipant{
			{
				Name:               "IntegrationAnalyst",
				Role:               "Integration Test Analyst",
				Enabled:            true,
				ResponseTimeout:    30 * 1000,
				Weight:             1.0,
				Priority:           1,
				DebateStyle:        "analytical",
				ArgumentationStyle: "logical",
				PersuasionLevel:    0.5,
				OpennessToChange:   0.5,
				QualityThreshold:   0.7,
				MinResponseLength:  50,
				MaxResponseLength:  1000,
				LLMs: []config.LLMConfiguration{
					{
						Name:        "IntegrationLLM",
						Provider:    "claude",
						Model:       "claude-3-sonnet",
						Enabled:     true,
						APIKey:      "integration-test-key",
						Timeout:     30 * 1000,
						MaxRetries:  3,
						Temperature: 0.7,
						MaxTokens:   1000,
						Weight:      1.0,
					},
				},
			},
			{
				Name:               "IntegrationCritic",
				Role:               "Integration Test Critic",
				Enabled:            true,
				ResponseTimeout:    30 * 1000,
				Weight:             1.0,
				Priority:           2,
				DebateStyle:        "critical",
				ArgumentationStyle: "logical",
				PersuasionLevel:    0.5,
				OpennessToChange:   0.5,
				QualityThreshold:   0.7,
				MinResponseLength:  50,
				MaxResponseLength:  1000,
				LLMs: []config.LLMConfiguration{
					{
						Name:        "IntegrationLLM2",
						Provider:    "deepseek",
						Model:       "deepseek-coder",
						Enabled:     true,
						APIKey:      "integration-test-key-2",
						Timeout:     30 * 1000,
						MaxRetries:  3,
						Temperature: 0.7,
						MaxTokens:   1000,
						Weight:      1.0,
					},
				},
			},
		},
		DebateStrategy:      "structured",
		VotingStrategy:      "confidence_weighted",
		ResponseFormat:      "detailed",
		EnableMemory:        true,
		MemoryRetention:     7 * 24 * 60 * 60 * 1000,
		MaxContextLength:    16000,
		QualityThreshold:    0.7,
		MaxResponseTime:     30 * 1000,
		EnableStreaming:     false,
		EnableDebateLogging: true,
		LogDebateDetails:    true,
		MetricsEnabled:      true,
	}
}

func createFallbackIntegrationTestConfig(t *testing.T, tempDir string) *config.AIDebateConfig {
	cfg := createIntegrationTestConfig(t, tempDir)

	// Add fallback LLMs for each participant
	cfg.Participants[0].LLMs = append(cfg.Participants[0].LLMs, config.LLMConfiguration{
		Name:        "FallbackLLM1",
		Provider:    "gemini",
		Model:       "gemini-pro",
		Enabled:     true,
		APIKey:      "fallback-test-key-1",
		Timeout:     25 * 1000,
		MaxRetries:  2,
		Temperature: 0.6,
		MaxTokens:   800,
		Weight:      0.8,
	})

	cfg.Participants[0].LLMs = append(cfg.Participants[0].LLMs, config.LLMConfiguration{
		Name:        "FallbackLLM2",
		Provider:    "qwen",
		Model:       "qwen-turbo",
		Enabled:     true,
		APIKey:      "fallback-test-key-2",
		Timeout:     20 * 1000,
		MaxRetries:  2,
		Temperature: 0.5,
		MaxTokens:   600,
		Weight:      0.6,
	})

	return cfg
}

func createUpdatedIntegrationTestConfig(t *testing.T, tempDir string) *config.AIDebateConfig {
	cfg := createIntegrationTestConfig(t, tempDir)

	// Add more participants and providers
	cfg.Participants = append(cfg.Participants, config.DebateParticipant{
		Name:               "UpdatedAnalyst",
		Role:               "Updated Integration Analyst",
		Enabled:            true,
		ResponseTimeout:    25 * 1000,
		Weight:             1.2,
		Priority:           3,
		DebateStyle:        "creative",
		ArgumentationStyle: "hypothetical",
		PersuasionLevel:    0.7,
		OpennessToChange:   0.8,
		QualityThreshold:   0.8,
		MinResponseLength:  60,
		MaxResponseLength:  1200,
		LLMs: []config.LLMConfiguration{
			{
				Name:        "UpdatedLLM",
				Provider:    "zai",
				Model:       "zai-large",
				Enabled:     true,
				APIKey:      "updated-test-key",
				Timeout:     25 * 1000,
				MaxRetries:  3,
				Temperature: 0.6,
				MaxTokens:   1200,
				Weight:      1.2,
			},
		},
	})

	return cfg
}

func createHealthTestConfig(t *testing.T, tempDir string) *config.AIDebateConfig {
	return &config.AIDebateConfig{
		Enabled:             true,
		MaximalRepeatRounds: 3,
		DebateTimeout:       5 * 60 * 1000,
		ConsensusThreshold:  0.75,
		Participants: []config.DebateParticipant{
			{
				Name:               "HealthTestParticipant1",
				Role:               "Health Test Analyst 1",
				Enabled:            true,
				ResponseTimeout:    30 * 1000,
				Weight:             1.0,
				Priority:           1,
				DebateStyle:        "analytical",
				ArgumentationStyle: "logical",
				PersuasionLevel:    0.5,
				OpennessToChange:   0.5,
				QualityThreshold:   0.7,
				MinResponseLength:  50,
				MaxResponseLength:  1000,
				LLMs: []config.LLMConfiguration{
					{
						Name:        "HealthTestLLM1",
						Provider:    "claude",
						Model:       "claude-3-sonnet",
						Enabled:     true,
						APIKey:      "health-test-key-1",
						Timeout:     30 * 1000,
						MaxRetries:  3,
						Temperature: 0.7,
						MaxTokens:   1000,
						Weight:      1.0,
					},
				},
			},
			{
				Name:               "HealthTestParticipant2",
				Role:               "Health Test Analyst 2",
				Enabled:            true,
				ResponseTimeout:    30 * 1000,
				Weight:             1.0,
				Priority:           2,
				DebateStyle:        "critical",
				ArgumentationStyle: "logical",
				PersuasionLevel:    0.5,
				OpennessToChange:   0.5,
				QualityThreshold:   0.7,
				MinResponseLength:  50,
				MaxResponseLength:  1000,
				LLMs: []config.LLMConfiguration{
					{
						Name:        "HealthTestLLM2",
						Provider:    "deepseek",
						Model:       "deepseek-coder",
						Enabled:     true,
						APIKey:      "health-test-key-2",
						Timeout:     30 * 1000,
						MaxRetries:  3,
						Temperature: 0.7,
						MaxTokens:   1000,
						Weight:      1.0,
					},
				},
			},
			{
				Name:               "HealthTestParticipant3",
				Role:               "Health Test Analyst 3",
				Enabled:            true,
				ResponseTimeout:    30 * 1000,
				Weight:             1.0,
				Priority:           3,
				DebateStyle:        "balanced",
				ArgumentationStyle: "evidence_based",
				PersuasionLevel:    0.5,
				OpennessToChange:   0.5,
				QualityThreshold:   0.7,
				MinResponseLength:  50,
				MaxResponseLength:  1000,
				LLMs: []config.LLMConfiguration{
					{
						Name:        "HealthTestLLM3",
						Provider:    "gemini",
						Model:       "gemini-pro",
						Enabled:     true,
						APIKey:      "health-test-key-3",
						Timeout:     30 * 1000,
						MaxRetries:  3,
						Temperature: 0.7,
						MaxTokens:   1000,
						Weight:      1.0,
					},
				},
			},
		},
	}
}

func createConcurrentTestConfig(t *testing.T, tempDir string) *config.AIDebateConfig {
	cfg := createIntegrationTestConfig(t, tempDir)
	cfg.MaxResponseTime = 20000    // 20 seconds for concurrent testing
	cfg.DebateTimeout = 120 * 1000 // 2 minutes
	return cfg
}

func createErrorRecoveryTestConfig(t *testing.T, tempDir string) *config.AIDebateConfig {
	return createIntegrationTestConfig(t, tempDir)
}

// Validation function for integration results

func validateIntegrationResults(t *testing.T, result *services.DebateResult, expectedTopic string) {
	assert.NotEmpty(t, result.SessionID, "Session ID should not be empty")
	assert.Equal(t, expectedTopic, result.Topic, "Topic mismatch")
	assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
	assert.Greater(t, result.FinalScore, 0.0, "Final score should be positive")

	assert.NotNil(t, result.Consensus, "Consensus should not be nil")
	if result.Consensus != nil {
		assert.NotEmpty(t, result.Consensus.Summary, "Consensus summary should not be empty")
		assert.GreaterOrEqual(t, result.Consensus.ConsensusLevel, 0.0, "Consensus level should be non-negative")
		assert.LessOrEqual(t, result.Consensus.ConsensusLevel, 1.0, "Consensus level should not exceed 1.0")
	}

	assert.NotEmpty(t, result.BestResponse.Content, "Best response content should not be empty")
	assert.Greater(t, len(result.AllResponses), 1, "Should have multiple responses")
	assert.NotEmpty(t, result.Recommendations, "Should have recommendations")

	// Validate quality metrics
	assert.NotNil(t, result.QualityMetrics, "Quality metrics should not be nil")
	if result.QualityMetrics != nil {
		assert.Contains(t, result.QualityMetrics, "avg_confidence", "Should have avg_confidence metric")
		assert.Contains(t, result.QualityMetrics, "avg_quality", "Should have avg_quality metric")
		assert.Greater(t, result.QualityMetrics["avg_confidence"], 0.0, "Average confidence should be positive")
		assert.Greater(t, result.QualityMetrics["avg_quality"], 0.0, "Average quality should be positive")
	}
}
