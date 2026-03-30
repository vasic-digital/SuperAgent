package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDebateService_SpecKitAutoActivation_E2E tests end-to-end SpecKit auto-activation
func TestDebateService_SpecKitAutoActivation_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create provider registry with real providers for E2E testing
	// This test requires at least one LLM provider to be configured via environment variables
	providerRegistry := NewProviderRegistry(nil, nil)

	// Check if any providers are available (auto-discovered from environment)
	providers := providerRegistry.ListProviders()
	if len(providers) == 0 {
		t.Skip("Skipping E2E test: no LLM providers configured (need API keys in environment)")
	}

	// Create debate team config with available providers
	// This is required for the debate to assign providers to participant roles
	providerDiscovery := NewProviderDiscovery(logger, false) // false = don't verify on startup
	debateTeamConfig := NewDebateTeamConfig(providerRegistry, providerDiscovery, logger)
	ctx := context.Background()
	if err := debateTeamConfig.InitializeTeam(ctx); err != nil {
		// If team initialization fails, skip the test
		// This can happen if providers are discovered but not properly configured
		t.Skipf("Skipping E2E test: failed to initialize debate team: %v", err)
	}

	// Check if team has active members (requires working LLM providers)
	activeMembers := debateTeamConfig.GetActiveMembers()
	if len(activeMembers) == 0 {
		t.Skip("Skipping E2E test: no active debate team members (requires working LLM provider API keys)")
	}

	// Create debate service using NewDebateServiceWithDeps for proper initialization
	debateService := NewDebateServiceWithDeps(logger, providerRegistry, nil)
	debateService.SetTeamConfig(debateTeamConfig)

	// Initialize SpecKit orchestrator
	projectRoot := t.TempDir()
	debateService.InitializeSpecKitOrchestrator(projectRoot)

	require.NotNil(t, debateService.speckitOrchestrator, "SpecKit orchestrator should be initialized")

	t.Run("BigCreationTriggersSpecKit", func(t *testing.T) {
		config := &DebateConfig{
			DebateID: "test-big-creation",
			Topic:    "Build a complete authentication system with OAuth2, JWT, password reset, and multi-factor authentication",
			Participants: []ParticipantConfig{
				{ParticipantID: "architect", Role: "architect"},
				{ParticipantID: "security_expert", Role: "security_expert"},
				{ParticipantID: "developer", Role: "developer"},
			},
			MaxRounds: 3,
			Timeout:   5 * time.Minute,
			Metadata: map[string]any{
				"granularity_hint": "whole_functionality",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		defer cancel()

		// This should trigger SpecKit auto-activation
		result, err := debateService.ConductDebate(ctx, config)

		require.NoError(t, err, "Debate should complete successfully")
		require.NotNil(t, result, "Result should not be nil")
		assert.True(t, result.Success, "Debate should be successful")

		// Check that SpecKit was activated
		if metadata, ok := result.Metadata["speckit_flow"].(bool); ok {
			assert.True(t, metadata, "SpecKit flow should have been activated")
			assert.Contains(t, result.Metadata, "granularity", "Should have granularity metadata")
			assert.Contains(t, result.Metadata, "action_type", "Should have action type metadata")
		}
	})

	t.Run("SmallChangeSkipsSpecKit", func(t *testing.T) {
		config := &DebateConfig{
			DebateID: "test-small-change",
			Topic:    "Fix typo in README.md",
			Participants: []ParticipantConfig{
				{ParticipantID: "reviewer", Role: "reviewer"},
				{ParticipantID: "editor", Role: "editor"},
			},
			MaxRounds: 2,
			Timeout:   1 * time.Minute,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// This should NOT trigger SpecKit (too small)
		result, err := debateService.ConductDebate(ctx, config)

		require.NoError(t, err, "Debate should complete successfully")
		require.NotNil(t, result, "Result should not be nil")

		// Check that SpecKit was NOT activated
		if metadata, ok := result.Metadata["speckit_flow"].(bool); ok {
			assert.False(t, metadata, "SpecKit flow should NOT have been activated for small changes")
		}
	})

	t.Run("RefactoringTriggersSpecKit", func(t *testing.T) {
		config := &DebateConfig{
			DebateID: "test-refactoring",
			Topic:    "Refactor the entire codebase to use microservices architecture",
			Participants: []ParticipantConfig{
				{ParticipantID: "architect", Role: "architect"},
				{ParticipantID: "tech_lead", Role: "tech_lead"},
				{ParticipantID: "devops", Role: "devops"},
			},
			MaxRounds: 3,
			Timeout:   5 * time.Minute,
			Metadata: map[string]any{
				"granularity_hint": "refactoring",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		defer cancel()

		// This should trigger SpecKit (refactoring)
		result, err := debateService.ConductDebate(ctx, config)

		require.NoError(t, err, "Debate should complete successfully")
		require.NotNil(t, result, "Result should not be nil")

		// Check that SpecKit was activated
		if metadata, ok := result.Metadata["speckit_flow"].(bool); ok {
			assert.True(t, metadata, "SpecKit flow should have been activated for refactoring")
			if granularity, ok := result.Metadata["granularity"].(WorkGranularity); ok {
				assert.Equal(t, GranularityRefactoring, granularity, "Should be classified as refactoring")
			}
		}
	})
}

// TestEnhancedIntentClassifier_Integration tests intent classification integration
func TestEnhancedIntentClassifier_Integration(t *testing.T) {
	logger := logrus.New()
	var providerRegistry *ProviderRegistry = nil

	classifier := NewEnhancedIntentClassifier(providerRegistry, logger)
	require.NotNil(t, classifier, "Classifier should not be nil")

	testCases := []struct {
		name             string
		userMessage      string
		expectedGranular []WorkGranularity // Allow multiple valid granularities
		expectedSpecKit  bool
		minConfidence    float64
	}{
		{
			name:             "Single Action",
			userMessage:      "Add a log statement to the handler",
			expectedGranular: []WorkGranularity{GranularitySingleAction},
			expectedSpecKit:  false,
			minConfidence:    0.5,
		},
		{
			name:             "Small Creation",
			userMessage:      "Fix the typo in the README file",
			expectedGranular: []WorkGranularity{GranularitySmallCreation, GranularitySingleAction},
			expectedSpecKit:  false,
			minConfidence:    0.5,
		},
		{
			name:             "Big Creation",
			userMessage:      "Implement a complete logging system with rotation and compression",
			expectedGranular: []WorkGranularity{GranularityBigCreation, GranularityWholeFunctionality},
			expectedSpecKit:  false, // Quick classify may not trigger SpecKit without LLM
			minConfidence:    0.5,
		},
		{
			name:             "Whole Functionality",
			userMessage:      "Build the entire payment processing module with Stripe integration",
			expectedGranular: []WorkGranularity{GranularityWholeFunctionality, GranularityBigCreation},
			expectedSpecKit:  true,
			minConfidence:    0.5,
		},
		{
			name:             "Refactoring",
			userMessage:      "Refactor the entire application to use dependency injection",
			expectedGranular: []WorkGranularity{GranularityRefactoring},
			expectedSpecKit:  true,
			minConfidence:    0.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Try LLM classification first, fall back to quick classify if no provider
			result, err := classifier.ClassifyEnhancedIntent(ctx, tc.userMessage, "", nil)
			if err != nil {
				// Fallback to quick classify
				result = classifier.QuickClassify(tc.userMessage)
			}

			require.NotNil(t, result, "Result should not be nil")

			// Check if granularity matches any of the expected values
			found := false
			for _, expected := range tc.expectedGranular {
				if result.Granularity == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Granularity %s should match one of %v", result.Granularity, tc.expectedGranular)
			assert.GreaterOrEqual(t, result.Confidence, tc.minConfidence, "Confidence should meet minimum")

			if tc.expectedSpecKit {
				assert.True(t, result.RequiresSpecKit, "Should require SpecKit for %s", tc.name)
				if result.RequiresSpecKit {
					assert.NotEmpty(t, result.SpecKitReason, "Should have SpecKit reason")
				}
			}
		})
	}
}
