package integration

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/services"
)

// ============================================================================
// Multi-Pass Validation Integration Tests
// ============================================================================
// These tests validate the complete multi-pass validation flow including:
// - Phase transitions
// - Streaming callbacks
// - Quality improvement tracking
// - Final synthesis generation
// ============================================================================

func TestMultiPassValidation_FullFlow(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("complete_validation_flow_with_mock_responses", func(t *testing.T) {
		// Create debate service
		debateService := services.NewDebateService(logger)
		require.NotNil(t, debateService)

		// Create validator
		validator := services.NewMultiPassValidator(debateService, logger)
		require.NotNil(t, validator)

		// Create mock debate result
		debateResult := &services.DebateResult{
			DebateID: "integration-test-1",
			Topic:    "What is the best approach to software testing?",
			AllResponses: []services.ParticipantResponse{
				{
					ParticipantID:   "analyst-1",
					ParticipantName: "THE ANALYST",
					Role:            "analyst",
					Content:         "Software testing should follow a pyramid approach with unit tests at the base, integration tests in the middle, and E2E tests at the top. This ensures comprehensive coverage while maintaining fast feedback cycles.",
					QualityScore:    0.85,
					Confidence:      0.8,
					LLMProvider:     "claude",
					LLMModel:        "claude-sonnet-4.5",
				},
				{
					ParticipantID:   "proposer-1",
					ParticipantName: "THE PROPOSER",
					Role:            "proposer",
					Content:         "I propose we focus on behavior-driven development (BDD) as it bridges the gap between technical and business teams. Tests become living documentation.",
					QualityScore:    0.82,
					Confidence:      0.75,
					LLMProvider:     "deepseek",
					LLMModel:        "deepseek-v3",
				},
				{
					ParticipantID:   "critic-1",
					ParticipantName: "THE CRITIC",
					Role:            "critic",
					Content:         "While these approaches have merit, we must consider the cost-benefit ratio. Over-testing can slow down development velocity without proportional quality gains.",
					QualityScore:    0.78,
					Confidence:      0.7,
					LLMProvider:     "gemini",
					LLMModel:        "gemini-2.0-flash",
				},
			},
			StartTime:    time.Now().Add(-5 * time.Minute),
			EndTime:      time.Now(),
			Duration:     5 * time.Minute,
			QualityScore: 0.82,
			Success:      true,
		}

		// Run validation
		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify result structure
		assert.Equal(t, "integration-test-1", result.DebateID)
		assert.Equal(t, debateResult.Topic, result.Topic)
		assert.GreaterOrEqual(t, len(result.Phases), 1)
		assert.NotEmpty(t, result.FinalResponse)
		assert.Greater(t, result.TotalDuration, time.Duration(0))
		assert.GreaterOrEqual(t, result.OverallConfidence, 0.0)
		assert.LessOrEqual(t, result.OverallConfidence, 1.0)
	})

	t.Run("validation_with_streaming_callbacks", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		// Track phase callbacks
		var mu sync.Mutex
		phaseCallbacks := make(map[services.ValidationPhase]bool)

		// Set up callbacks for each phase
		phases := []services.ValidationPhase{
			services.PhaseValidation,
			services.PhasePolishImprove,
			services.PhaseFinalConclusion,
		}
		for _, phase := range phases {
			p := phase // Capture for closure
			validator.SetPhaseCallback(p, func(result *services.PhaseResult) {
				mu.Lock()
				defer mu.Unlock()
				phaseCallbacks[p] = true
			})
		}

		// Create debate result
		debateResult := &services.DebateResult{
			DebateID: "callback-test",
			Topic:    "Test topic",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "p1", Content: "Response 1", QualityScore: 0.7, Confidence: 0.6},
				{ParticipantID: "p2", Content: "Response 2", QualityScore: 0.8, Confidence: 0.7},
			},
			StartTime: time.Now(),
		}

		// Run validation
		ctx := context.Background()
		_, err := validator.ValidateAndImprove(ctx, debateResult)
		require.NoError(t, err)

		// Verify callbacks were invoked
		mu.Lock()
		defer mu.Unlock()
		assert.True(t, phaseCallbacks[services.PhaseFinalConclusion], "Final conclusion callback should be invoked")
	})

	t.Run("validation_skips_polish_on_high_confidence", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		// Set config to skip polish on high confidence
		config := services.DefaultValidationConfig()
		config.MinConfidenceToSkip = 0.5 // Low threshold - should skip polish
		validator.SetConfig(config)

		// Create high-confidence responses
		debateResult := &services.DebateResult{
			DebateID: "high-confidence-test",
			Topic:    "Simple topic",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "p1", Content: "High confidence response", QualityScore: 0.95, Confidence: 0.95},
				{ParticipantID: "p2", Content: "Another confident response", QualityScore: 0.92, Confidence: 0.92},
			},
			StartTime: time.Now(),
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)
		require.NoError(t, err)

		// Check that polish was skipped (should have fewer phases)
		hasPolishPhase := false
		for _, phase := range result.Phases {
			if phase.Phase == services.PhasePolishImprove {
				hasPolishPhase = true
			}
		}
		assert.False(t, hasPolishPhase, "Polish phase should be skipped for high confidence responses")
	})
}

func TestMultiPassValidation_PhaseFormatting(t *testing.T) {
	t.Run("phase_headers_are_formatted_correctly", func(t *testing.T) {
		phases := []services.ValidationPhase{
			services.PhaseInitialResponse,
			services.PhaseValidation,
			services.PhasePolishImprove,
			services.PhaseFinalConclusion,
		}

		for _, phase := range phases {
			header := services.FormatPhaseHeader(phase, true)
			assert.NotEmpty(t, header, "Header for %s should not be empty", phase)
			assert.Contains(t, header, "PHASE", "Header should contain PHASE")
			assert.Contains(t, header, "═", "Header should have box drawing characters")
		}
	})

	t.Run("phase_footers_include_metrics", func(t *testing.T) {
		result := &services.PhaseResult{
			Phase:        services.PhaseValidation,
			Duration:     5 * time.Second,
			PhaseScore:   0.85,
			PhaseSummary: "Test summary",
		}

		footer := services.FormatPhaseFooter(services.PhaseValidation, result, true)
		assert.Contains(t, footer, "5s")
		assert.Contains(t, footer, "0.85")
	})

	t.Run("multipass_output_includes_all_sections", func(t *testing.T) {
		result := &services.MultiPassResult{
			DebateID: "format-test",
			Topic:    "Format testing topic",
			Phases: []*services.PhaseResult{
				{Phase: services.PhaseInitialResponse, PhaseScore: 0.8},
				{Phase: services.PhaseValidation, PhaseScore: 0.85},
				{Phase: services.PhaseFinalConclusion, PhaseScore: 0.9, Responses: []services.ParticipantResponse{{Content: "Final"}}},
			},
			TotalDuration:      time.Minute,
			OverallConfidence:  0.88,
			QualityImprovement: 10.0,
		}

		output := services.FormatMultiPassOutput(result)

		// Verify all major sections are present
		assert.Contains(t, output, "HELIXAGENT")
		assert.Contains(t, output, "MULTI-PASS VALIDATION")
		assert.Contains(t, output, "Format testing topic")
		assert.Contains(t, output, "FINAL SUMMARY")
		assert.Contains(t, output, "88%") // Confidence
	})
}

func TestMultiPassValidation_EdgeCases(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	t.Run("handles_empty_responses", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		debateResult := &services.DebateResult{
			DebateID:     "empty-test",
			Topic:        "Empty topic",
			AllResponses: []services.ParticipantResponse{},
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "empty-test", result.DebateID)
	})

	t.Run("handles_single_response", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		debateResult := &services.DebateResult{
			DebateID: "single-test",
			Topic:    "Single response topic",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "solo", Content: "Only response", QualityScore: 0.8},
			},
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.FinalResponse)
	})

	t.Run("handles_very_long_content", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		longContent := strings.Repeat("This is a very long response that goes on and on. ", 100)

		debateResult := &services.DebateResult{
			DebateID: "long-content-test",
			Topic:    "Long content topic",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "verbose", Content: longContent, QualityScore: 0.75},
				{ParticipantID: "concise", Content: "Short response", QualityScore: 0.85},
			},
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("handles_special_characters_in_content", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		debateResult := &services.DebateResult{
			DebateID: "special-char-test",
			Topic:    "Topic with <special> & \"characters\"",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "p1", Content: "Response with <html>, \"quotes\", and & symbols", QualityScore: 0.8},
			},
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("handles_context_cancellation_gracefully", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		debateResult := &services.DebateResult{
			DebateID: "cancel-test",
			Topic:    "Cancellation test",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "p1", Content: "Response", QualityScore: 0.8},
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result, err := validator.ValidateAndImprove(ctx, debateResult)

		// Should handle gracefully
		assert.NotNil(t, result)
		assert.NoError(t, err)
	})
}

func TestMultiPassValidation_QualityMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	t.Run("tracks_quality_improvement", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		// Create responses with varying quality
		debateResult := &services.DebateResult{
			DebateID: "quality-test",
			Topic:    "Quality tracking test",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "p1", Content: "Good response with details", QualityScore: 0.7, Confidence: 0.65},
				{ParticipantID: "p2", Content: "Another decent response", QualityScore: 0.75, Confidence: 0.7},
			},
			QualityScore: 0.725, // Average
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		assert.NotNil(t, result.Metadata)

		// Check quality metrics are tracked
		if initialQuality, ok := result.Metadata["initial_quality"]; ok {
			assert.Greater(t, initialQuality.(float64), 0.0)
		}
		if finalQuality, ok := result.Metadata["final_quality"]; ok {
			assert.Greater(t, finalQuality.(float64), 0.0)
		}
	})

	t.Run("confidence_is_bounded", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		debateResult := &services.DebateResult{
			DebateID: "confidence-test",
			Topic:    "Confidence bounds test",
			AllResponses: []services.ParticipantResponse{
				{ParticipantID: "p1", Content: "Response", QualityScore: 0.8, Confidence: 1.0},
				{ParticipantID: "p2", Content: "Response", QualityScore: 0.8, Confidence: 0.0},
			},
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.OverallConfidence, 0.0)
		assert.LessOrEqual(t, result.OverallConfidence, 1.0)
	})
}

func TestMultiPassValidation_Concurrency(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	t.Run("handles_concurrent_validations", func(t *testing.T) {
		debateService := services.NewDebateService(logger)

		// Run multiple validations concurrently
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				validator := services.NewMultiPassValidator(debateService, logger)
				debateResult := &services.DebateResult{
					DebateID: "concurrent-" + string(rune('0'+idx)),
					Topic:    "Concurrent test",
					AllResponses: []services.ParticipantResponse{
						{ParticipantID: "p1", Content: "Response", QualityScore: 0.8},
					},
				}

				ctx := context.Background()
				_, err := validator.ValidateAndImprove(ctx, debateResult)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Concurrent validation error: %v", err)
		}
	})

	t.Run("config_changes_are_thread_safe", func(t *testing.T) {
		debateService := services.NewDebateService(logger)
		validator := services.NewMultiPassValidator(debateService, logger)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				validator.SetConfig(&services.ValidationConfig{
					EnableValidation: true,
					EnablePolish:     false,
				})
			}()
			go func() {
				defer wg.Done()
				_ = validator.GetConfig()
			}()
		}
		wg.Wait()
		// Test passes if no race condition panic
	})
}

func TestValidationPhases_Integration(t *testing.T) {
	t.Run("all_phases_return_valid_info", func(t *testing.T) {
		phases := services.ValidationPhases()
		require.Len(t, phases, 4)

		for _, phase := range phases {
			info := services.GetPhaseInfo(phase.Phase)
			require.NotNil(t, info)
			assert.Equal(t, phase.Phase, info.Phase)
			assert.NotEmpty(t, info.Name)
			assert.NotEmpty(t, info.Description)
			assert.NotEmpty(t, info.Icon)
			assert.Greater(t, info.Order, 0)
		}
	})

	t.Run("phase_order_is_sequential", func(t *testing.T) {
		phases := services.ValidationPhases()
		for i, phase := range phases {
			assert.Equal(t, i+1, phase.Order)
		}
	})
}
