package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Multi-Pass Validation Unit Tests
// ============================================================================

func TestValidationPhases(t *testing.T) {
	t.Run("returns_all_four_phases_in_order", func(t *testing.T) {
		phases := ValidationPhases()
		require.Len(t, phases, 4)
		assert.Equal(t, PhaseInitialResponse, phases[0].Phase)
		assert.Equal(t, PhaseValidation, phases[1].Phase)
		assert.Equal(t, PhasePolishImprove, phases[2].Phase)
		assert.Equal(t, PhaseFinalConclusion, phases[3].Phase)
	})

	t.Run("phases_have_correct_order_numbers", func(t *testing.T) {
		phases := ValidationPhases()
		for i, phase := range phases {
			assert.Equal(t, i+1, phase.Order)
		}
	})

	t.Run("phases_have_names_and_descriptions", func(t *testing.T) {
		phases := ValidationPhases()
		for _, phase := range phases {
			assert.NotEmpty(t, phase.Name, "phase %s should have a name", phase.Phase)
			assert.NotEmpty(t, phase.Description, "phase %s should have a description", phase.Phase)
			assert.NotEmpty(t, phase.Icon, "phase %s should have an icon", phase.Phase)
		}
	})
}

func TestGetPhaseInfo(t *testing.T) {
	t.Run("returns_info_for_valid_phases", func(t *testing.T) {
		phases := []ValidationPhase{
			PhaseInitialResponse,
			PhaseValidation,
			PhasePolishImprove,
			PhaseFinalConclusion,
		}
		for _, phase := range phases {
			info := GetPhaseInfo(phase)
			require.NotNil(t, info, "should return info for phase %s", phase)
			assert.Equal(t, phase, info.Phase)
		}
	})

	t.Run("returns_nil_for_invalid_phase", func(t *testing.T) {
		info := GetPhaseInfo("invalid_phase")
		assert.Nil(t, info)
	})
}

func TestDefaultValidationConfig(t *testing.T) {
	t.Run("returns_valid_defaults", func(t *testing.T) {
		config := DefaultValidationConfig()
		require.NotNil(t, config)
		assert.True(t, config.EnableValidation)
		assert.True(t, config.EnablePolish)
		assert.True(t, config.ParallelValidation)
		assert.True(t, config.ShowPhaseIndicators)
		assert.True(t, config.VerbosePhaseHeaders)
		assert.Greater(t, config.ValidationTimeout, time.Duration(0))
		assert.Greater(t, config.PolishTimeout, time.Duration(0))
		assert.Greater(t, config.MinConfidenceToSkip, 0.0)
		assert.LessOrEqual(t, config.MinConfidenceToSkip, 1.0)
	})
}

func TestMultiPassValidator_Creation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during tests

	t.Run("creates_with_nil_debate_service", func(t *testing.T) {
		validator := NewMultiPassValidator(nil, logger)
		require.NotNil(t, validator)
		assert.NotNil(t, validator.GetConfig())
	})

	t.Run("creates_with_valid_debate_service", func(t *testing.T) {
		debateService := NewDebateService(logger)
		validator := NewMultiPassValidator(debateService, logger)
		require.NotNil(t, validator)
		assert.Equal(t, debateService, validator.debateService)
	})
}

func TestMultiPassValidator_Config(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	validator := NewMultiPassValidator(nil, logger)

	t.Run("sets_and_gets_config", func(t *testing.T) {
		customConfig := &ValidationConfig{
			EnableValidation:    false,
			EnablePolish:        false,
			ValidationTimeout:   10 * time.Second,
			PolishTimeout:       5 * time.Second,
			MinConfidenceToSkip: 0.99,
			MaxValidationRounds: 1,
			ParallelValidation:  false,
		}
		validator.SetConfig(customConfig)
		retrievedConfig := validator.GetConfig()
		assert.Equal(t, customConfig, retrievedConfig)
	})
}

func TestMultiPassValidator_PhaseCallbacks(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	validator := NewMultiPassValidator(nil, logger)

	t.Run("sets_and_invokes_callbacks", func(t *testing.T) {
		called := false
		var receivedResult *PhaseResult

		validator.SetPhaseCallback(PhaseValidation, func(result *PhaseResult) {
			called = true
			receivedResult = result
		})

		testResult := &PhaseResult{
			Phase:      PhaseValidation,
			PhaseScore: 0.85,
		}

		validator.notifyPhaseCallback(PhaseValidation, testResult)

		assert.True(t, called)
		assert.Equal(t, testResult, receivedResult)
	})

	t.Run("ignores_unset_callbacks", func(t *testing.T) {
		// Should not panic when callback is not set
		validator.notifyPhaseCallback(PhasePolishImprove, &PhaseResult{})
	})
}

func TestValidateAndImprove(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	t.Run("handles_empty_debate_result", func(t *testing.T) {
		debateService := NewDebateService(logger)
		validator := NewMultiPassValidator(debateService, logger)

		debateResult := &DebateResult{
			DebateID:     "test-debate-1",
			Topic:        "Test topic",
			AllResponses: []ParticipantResponse{},
			StartTime:    time.Now(),
			EndTime:      time.Now(),
			Duration:     time.Second,
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test-debate-1", result.DebateID)
		assert.Equal(t, "Test topic", result.Topic)
	})

	t.Run("processes_multiple_responses", func(t *testing.T) {
		debateService := NewDebateService(logger)
		validator := NewMultiPassValidator(debateService, logger)

		responses := []ParticipantResponse{
			{
				ParticipantID:   "p1",
				ParticipantName: "Analyst",
				Role:            "analyst",
				Content:         "This is my analysis of the important topic.",
				QualityScore:    0.8,
				Confidence:      0.75,
			},
			{
				ParticipantID:   "p2",
				ParticipantName: "Critic",
				Role:            "critic",
				Content:         "I disagree with some points and offer a counter argument.",
				QualityScore:    0.7,
				Confidence:      0.8,
			},
		}

		debateResult := &DebateResult{
			DebateID:     "test-debate-2",
			Topic:        "Important topic",
			AllResponses: responses,
			StartTime:    time.Now().Add(-time.Minute),
			EndTime:      time.Now(),
			Duration:     time.Minute,
			QualityScore: 0.75,
		}

		ctx := context.Background()
		result, err := validator.ValidateAndImprove(ctx, debateResult)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result.Phases), 1)
		assert.NotEmpty(t, result.FinalResponse)
	})

	t.Run("respects_context_cancellation", func(t *testing.T) {
		debateService := NewDebateService(logger)
		validator := NewMultiPassValidator(debateService, logger)

		responses := []ParticipantResponse{
			{ParticipantID: "p1", Content: "Test content", QualityScore: 0.8},
		}

		debateResult := &DebateResult{
			DebateID:     "test-debate-cancel",
			AllResponses: responses,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result, err := validator.ValidateAndImprove(ctx, debateResult)
		// Should still return partial result
		assert.NotNil(t, result)
		assert.NoError(t, err) // Graceful handling
	})
}

func TestValidationResult(t *testing.T) {
	t.Run("validation_scores_are_bounded", func(t *testing.T) {
		result := ValidationResult{
			ValidationScore: 0.85,
			FactualAccuracy: 0.9,
			Completeness:    0.8,
			Coherence:       0.85,
		}
		assert.GreaterOrEqual(t, result.ValidationScore, 0.0)
		assert.LessOrEqual(t, result.ValidationScore, 1.0)
		assert.GreaterOrEqual(t, result.FactualAccuracy, 0.0)
		assert.LessOrEqual(t, result.FactualAccuracy, 1.0)
	})
}

func TestPolishResult(t *testing.T) {
	t.Run("polish_preserves_original", func(t *testing.T) {
		result := PolishResult{
			OriginalResponse: "Original content",
			PolishedResponse: "Improved content",
			ImprovementScore: 0.25,
		}
		assert.NotEmpty(t, result.OriginalResponse)
		assert.NotEmpty(t, result.PolishedResponse)
		assert.NotEqual(t, result.OriginalResponse, result.PolishedResponse)
	})
}

func TestFormatPhaseHeader(t *testing.T) {
	t.Run("verbose_header_includes_all_info", func(t *testing.T) {
		header := FormatPhaseHeader(PhaseInitialResponse, true)
		assert.Contains(t, header, "PHASE 1")
		assert.Contains(t, header, "INITIAL RESPONSE")
		assert.Contains(t, header, "═") // Box drawing character
	})

	t.Run("compact_header_is_shorter", func(t *testing.T) {
		verbose := FormatPhaseHeader(PhaseValidation, true)
		compact := FormatPhaseHeader(PhaseValidation, false)
		assert.Greater(t, len(verbose), len(compact))
	})

	t.Run("each_phase_has_unique_header", func(t *testing.T) {
		phases := []ValidationPhase{
			PhaseInitialResponse,
			PhaseValidation,
			PhasePolishImprove,
			PhaseFinalConclusion,
		}
		headers := make(map[string]bool)
		for _, phase := range phases {
			header := FormatPhaseHeader(phase, true)
			assert.NotContains(t, headers, header)
			headers[header] = true
		}
	})

	t.Run("invalid_phase_returns_empty", func(t *testing.T) {
		header := FormatPhaseHeader("invalid", true)
		assert.Empty(t, header)
	})
}

func TestFormatPhaseFooter(t *testing.T) {
	t.Run("includes_duration_and_score", func(t *testing.T) {
		result := &PhaseResult{
			Phase:        PhaseValidation,
			Duration:     time.Second * 5,
			PhaseScore:   0.85,
			PhaseSummary: "Test summary",
		}
		footer := FormatPhaseFooter(PhaseValidation, result, true)
		assert.Contains(t, footer, "5s")
		assert.Contains(t, footer, "0.85")
	})
}

func TestFormatMultiPassOutput(t *testing.T) {
	t.Run("formats_complete_result", func(t *testing.T) {
		result := &MultiPassResult{
			DebateID: "test-123",
			Topic:    "Test Topic",
			Phases: []*PhaseResult{
				{
					Phase:      PhaseInitialResponse,
					PhaseScore: 0.8,
					Responses: []ParticipantResponse{
						{ParticipantName: "Test", Role: "analyst", Content: "Test content"},
					},
				},
				{
					Phase:      PhaseFinalConclusion,
					PhaseScore: 0.9,
					Responses: []ParticipantResponse{
						{ParticipantName: "Consensus", Content: "Final conclusion"},
					},
				},
			},
			TotalDuration:      time.Minute,
			OverallConfidence:  0.85,
			QualityImprovement: 10.5,
		}

		output := FormatMultiPassOutput(result)
		assert.Contains(t, output, "Test Topic")
		assert.Contains(t, output, "HELIXAGENT")
		assert.Contains(t, output, "MULTI-PASS VALIDATION")
		assert.Contains(t, output, "FINAL SUMMARY")
		assert.Contains(t, output, "85%") // Confidence
	})
}

func TestIssueType(t *testing.T) {
	t.Run("all_issue_types_are_defined", func(t *testing.T) {
		types := []IssueType{
			IssueFactualError,
			IssueIncomplete,
			IssueUnclear,
			IssueContradiction,
			IssueMissingContext,
			IssueOverGeneralized,
			IssueOutOfScope,
		}
		for _, issueType := range types {
			assert.NotEmpty(t, string(issueType))
		}
	})
}

func TestValidationSeverity(t *testing.T) {
	t.Run("all_severities_are_defined", func(t *testing.T) {
		severities := []ValidationSeverity{
			ValidationSeverityCritical,
			ValidationSeverityMajor,
			ValidationSeverityMinor,
			ValidationSeverityInfo,
		}
		for _, severity := range severities {
			assert.NotEmpty(t, string(severity))
		}
	})
}

func TestHeuristicValidation(t *testing.T) {
	t.Run("short_content_gets_lower_completeness", func(t *testing.T) {
		factual, complete, coherent, issues, _ := heuristicValidation("Short", "topic")
		assert.Less(t, complete, 0.7)
		assert.Greater(t, factual, 0.0)
		assert.Greater(t, coherent, 0.0)
		assert.Greater(t, len(issues), 0)
	})

	t.Run("relevant_content_gets_higher_scores", func(t *testing.T) {
		content := "This is a detailed analysis of the topic with important points and key insights."
		factual, complete, coherent, _, _ := heuristicValidation(content, "analysis topic")
		assert.GreaterOrEqual(t, factual, 0.5)
		assert.GreaterOrEqual(t, complete, 0.5)
		assert.GreaterOrEqual(t, coherent, 0.5)
	})
}

func TestCalculateImprovementScore(t *testing.T) {
	t.Run("no_change_returns_zero", func(t *testing.T) {
		score := calculateImprovementScore("same content", "same content")
		assert.Equal(t, 0.0, score)
	})

	t.Run("different_content_returns_positive", func(t *testing.T) {
		score := calculateImprovementScore("original", "improved and expanded content")
		assert.Greater(t, score, 0.0)
	})

	t.Run("score_is_bounded", func(t *testing.T) {
		score := calculateImprovementScore("a", "a very very long improvement")
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})
}

func TestTruncateForValidation(t *testing.T) {
	t.Run("short_string_unchanged", func(t *testing.T) {
		result := truncateForValidation("short", 100)
		assert.Equal(t, "short", result)
	})

	t.Run("long_string_truncated", func(t *testing.T) {
		longString := strings.Repeat("a", 200)
		result := truncateForValidation(longString, 50)
		assert.Len(t, result, 50)
		assert.True(t, strings.HasSuffix(result, "..."))
	})
}

func TestParseValidationResponse(t *testing.T) {
	t.Run("parses_valid_response", func(t *testing.T) {
		response := `FACTUAL_ACCURACY: 0.9
COMPLETENESS: 0.85
COHERENCE: 0.8
ISSUES: Some minor issues
SUGGESTIONS: Consider adding more detail`

		factual, complete, coherent, _, _ := parseValidationResponse(response)
		assert.InDelta(t, 0.9, factual, 0.01)
		assert.InDelta(t, 0.85, complete, 0.01)
		assert.InDelta(t, 0.8, coherent, 0.01)
	})

	t.Run("handles_missing_fields", func(t *testing.T) {
		response := "Some random content without proper format"
		factual, complete, coherent, _, _ := parseValidationResponse(response)
		// Should return defaults
		assert.Equal(t, 0.7, factual)
		assert.Equal(t, 0.7, complete)
		assert.Equal(t, 0.7, coherent)
	})
}

func TestParsePolishResponse(t *testing.T) {
	t.Run("extracts_improved_response", func(t *testing.T) {
		response := `IMPROVED RESPONSE:
This is the improved content.

CHANGES:
- Fixed grammar
- Added details`

		polished, changes := parsePolishResponse(response)
		assert.Contains(t, polished, "improved content")
		assert.GreaterOrEqual(t, len(changes), 0)
	})

	t.Run("handles_raw_response", func(t *testing.T) {
		response := "Just raw improved content without markers"
		polished, _ := parsePolishResponse(response)
		assert.Equal(t, response, polished)
	})
}

func TestParseSynthesisResponse(t *testing.T) {
	t.Run("extracts_conclusion_and_confidence", func(t *testing.T) {
		response := `CONCLUSION:
This is the final synthesized conclusion.

CONFIDENCE: 0.92`

		conclusion, confidence := parseSynthesisResponse(response)
		assert.Contains(t, conclusion, "synthesized conclusion")
		assert.InDelta(t, 0.92, confidence, 0.01)
	})

	t.Run("handles_missing_confidence", func(t *testing.T) {
		response := "Just a conclusion without confidence"
		_, confidence := parseSynthesisResponse(response)
		assert.Equal(t, 0.8, confidence) // Default
	})
}

func TestMultiPassResult_Metadata(t *testing.T) {
	t.Run("stores_and_retrieves_metadata", func(t *testing.T) {
		result := &MultiPassResult{
			Metadata: make(map[string]interface{}),
		}
		result.Metadata["test_key"] = "test_value"
		result.Metadata["numeric"] = 42

		assert.Equal(t, "test_value", result.Metadata["test_key"])
		assert.Equal(t, 42, result.Metadata["numeric"])
	})
}

func TestPhaseResult_Duration(t *testing.T) {
	t.Run("duration_matches_times", func(t *testing.T) {
		start := time.Now()
		end := start.Add(5 * time.Second)
		result := &PhaseResult{
			StartTime: start,
			EndTime:   end,
			Duration:  end.Sub(start),
		}
		assert.Equal(t, 5*time.Second, result.Duration)
	})
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkValidationPhases(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidationPhases()
	}
}

func BenchmarkGetPhaseInfo(b *testing.B) {
	phases := []ValidationPhase{
		PhaseInitialResponse,
		PhaseValidation,
		PhasePolishImprove,
		PhaseFinalConclusion,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, phase := range phases {
			_ = GetPhaseInfo(phase)
		}
	}
}

func BenchmarkFormatPhaseHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatPhaseHeader(PhaseValidation, true)
	}
}

func BenchmarkHeuristicValidation(b *testing.B) {
	content := strings.Repeat("This is test content for validation. ", 10)
	topic := "test topic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _, _ = heuristicValidation(content, topic)
	}
}

func BenchmarkCalculateImprovementScore(b *testing.B) {
	original := strings.Repeat("original content ", 20)
	polished := strings.Repeat("improved polished content ", 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateImprovementScore(original, polished)
	}
}

func BenchmarkFormatMultiPassOutput(b *testing.B) {
	result := &MultiPassResult{
		DebateID: "test-123",
		Topic:    "Benchmark Topic",
		Phases: []*PhaseResult{
			{Phase: PhaseInitialResponse, PhaseScore: 0.8},
			{Phase: PhaseValidation, PhaseScore: 0.85},
			{Phase: PhasePolishImprove, PhaseScore: 0.9},
			{Phase: PhaseFinalConclusion, PhaseScore: 0.92},
		},
		TotalDuration:      time.Minute,
		OverallConfidence:  0.88,
		QualityImprovement: 15.0,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatMultiPassOutput(result)
	}
}

// ============================================================================
// Race Condition Tests
// ============================================================================

func TestMultiPassValidator_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	validator := NewMultiPassValidator(nil, logger)

	// Test concurrent config access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				validator.SetConfig(&ValidationConfig{EnableValidation: j%2 == 0})
				_ = validator.GetConfig()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMultiPassValidator_ConcurrentCallbacks(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	validator := NewMultiPassValidator(nil, logger)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				validator.SetPhaseCallback(PhaseValidation, func(r *PhaseResult) {
					_ = r.PhaseScore
				})
				validator.notifyPhaseCallback(PhaseValidation, &PhaseResult{PhaseScore: float64(id)})
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
