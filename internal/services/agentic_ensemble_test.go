package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestLogger creates a silent logger for tests.
func newTestEnsembleLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.PanicLevel)
	return l
}

// newMinimalConfig returns an AgenticEnsembleConfig suitable for fast
// unit tests.
func newMinimalConfig() AgenticEnsembleConfig {
	return AgenticEnsembleConfig{
		MaxConcurrentAgents:       2,
		MaxIterationsPerAgent:     3,
		MaxToolIterationsPerPhase: 2,
		AgentTimeout:              5 * time.Second,
		GlobalTimeout:             10 * time.Second,
		ToolIterationTimeout:      2 * time.Second,
		EnableVision:              false,
		EnableMemory:              false,
		EnableExecution:           true,
	}
}

// newTestRequest creates an LLMRequest with the given user message.
func newTestRequest(msg string) *models.LLMRequest {
	return &models.LLMRequest{
		ID: "test-req-1",
		Messages: []models.Message{
			{Role: "user", Content: msg},
		},
		Prompt: msg,
	}
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ReasonMode
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ReasonMode(t *testing.T) {
	logger := newTestEnsembleLogger()

	// Without an intent classifier the ensemble always selects reason mode.
	// Without a debate service the reason mode returns an error.
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, // no debate service
		nil, // no intent classifier
		nil, // no tool executor
		nil, // no planner
		nil, // no verifier
		nil, // no provider registry
		cfg,
		logger,
	)

	req := newTestRequest("What is the capital of France?")
	_, err := ensemble.Process(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "debate service not configured")
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ExecuteMode
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ExecuteMode(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	cfg.EnableExecution = true

	// Without a debate service even execute mode cannot proceed (understand
	// stage calls debate).
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	req := newTestRequest("Refactor the authentication module")
	_, err := ensemble.Process(context.Background(), req)
	require.Error(t, err)
	// The error should reference the debate service being missing because
	// even execute mode starts with the understand stage.
	assert.Contains(t, err.Error(), "debate service not configured")
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ExecuteDisabled
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ExecuteDisabled(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	cfg.EnableExecution = false

	// Even with an actionable intent, when execution is disabled the
	// ensemble should fall back to reason mode.
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	req := newTestRequest("Deploy the application to production")
	_, err := ensemble.Process(context.Background(), req)
	require.Error(t, err)
	// Falls back to reason mode which needs the debate service.
	assert.Contains(t, err.Error(), "debate service not configured")
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_GlobalTimeout
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_GlobalTimeout(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	cfg.GlobalTimeout = 1 * time.Millisecond // Extremely short

	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	req := newTestRequest("Do something")
	_, err := ensemble.Process(context.Background(), req)
	require.Error(t, err)
	// Should fail because of missing debate service (the timeout context
	// is still set even though the service is nil).
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_NilDependencies
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_NilDependencies(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()

	// All nil dependencies — should not panic.
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)
	require.NotNil(t, ensemble)

	// Nil request should return an error without panicking.
	_, err := ensemble.Process(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil request")
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_NilLogger
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_NilLogger(t *testing.T) {
	cfg := newMinimalConfig()

	// Nil logger should be replaced with a default.
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, nil,
	)
	require.NotNil(t, ensemble)
	assert.NotNil(t, ensemble.logger)
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ClassifyMode_NoClassifier
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ClassifyMode_NoClassifier(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()

	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	req := newTestRequest("Tell me a joke")
	mode := ensemble.classifyMode(context.Background(), req)
	assert.Equal(t, AgenticModeReason, mode)
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ClassifyMode_EmptyMessage
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ClassifyMode_EmptyMessage(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()

	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	req := &models.LLMRequest{ID: "empty"}
	mode := ensemble.classifyMode(context.Background(), req)
	assert.Equal(t, AgenticModeReason, mode)
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ExtractUserMessage
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ExtractUserMessage(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	t.Run("from messages", func(t *testing.T) {
		req := &models.LLMRequest{
			Messages: []models.Message{
				{Role: "system", Content: "You are helpful."},
				{Role: "user", Content: "Hello!"},
				{Role: "assistant", Content: "Hi there."},
				{Role: "user", Content: "What time is it?"},
			},
			Prompt: "fallback prompt",
		}
		msg := ensemble.extractUserMessage(req)
		assert.Equal(t, "What time is it?", msg)
	})

	t.Run("from prompt fallback", func(t *testing.T) {
		req := &models.LLMRequest{
			Messages: []models.Message{
				{Role: "system", Content: "System only."},
			},
			Prompt: "prompt text",
		}
		msg := ensemble.extractUserMessage(req)
		assert.Equal(t, "prompt text", msg)
	})

	t.Run("nil request", func(t *testing.T) {
		msg := ensemble.extractUserMessage(nil)
		assert.Equal(t, "", msg)
	})
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_ExtractContent
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_ExtractContent(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	t.Run("nil result", func(t *testing.T) {
		assert.Equal(t, "", ensemble.extractContent(nil))
	})

	t.Run("selected response", func(t *testing.T) {
		result := &EnsembleResult{
			Selected: &models.LLMResponse{Content: "selected content"},
		}
		assert.Equal(t, "selected content", ensemble.extractContent(result))
	})

	t.Run("first response fallback", func(t *testing.T) {
		result := &EnsembleResult{
			Responses: []*models.LLMResponse{
				{Content: "first response"},
			},
		}
		assert.Equal(t, "first response", ensemble.extractContent(result))
	})
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_DebateResultToEnsemble
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_DebateResultToEnsemble(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	t.Run("nil debate result", func(t *testing.T) {
		result := ensemble.debateResultToEnsemble(nil, AgenticModeReason)
		require.NotNil(t, result)
		assert.Nil(t, result.Selected)
	})

	t.Run("with best response", func(t *testing.T) {
		dr := &DebateResult{
			DebateID:     "test-debate",
			SessionID:    "test-session",
			Duration:     2 * time.Second,
			TotalRounds:  3,
			QualityScore: 0.85,
			Success:      true,
			BestResponse: &ParticipantResponse{
				Content:      "Best answer here",
				LLMProvider:  "deepseek",
				QualityScore: 0.9,
			},
			Participants: []ParticipantResponse{
				{
					ParticipantID: "p1",
					Content:       "Response 1",
					LLMProvider:   "deepseek",
					QualityScore:  0.9,
				},
			},
		}

		result := ensemble.debateResultToEnsemble(dr, AgenticModeReason)
		require.NotNil(t, result)
		require.NotNil(t, result.Selected)
		assert.Equal(t, "Best answer here", result.Selected.Content)
		assert.Equal(t, "deepseek", result.Selected.ProviderName)
		assert.Equal(t, "debate", result.VotingMethod)
		assert.Len(t, result.Responses, 1)

		agenticMeta, ok := result.Metadata["agentic"].(*AgenticMetadata)
		require.True(t, ok)
		assert.Equal(t, "reason", agenticMeta.Mode)
	})

	t.Run("with consensus fallback", func(t *testing.T) {
		dr := &DebateResult{
			DebateID: "test-debate-2",
			Duration: 1 * time.Second,
			Consensus: &ConsensusResult{
				FinalPosition: "Consensus position",
				Confidence:    0.88,
			},
		}

		result := ensemble.debateResultToEnsemble(dr, AgenticModeExecute)
		require.NotNil(t, result)
		require.NotNil(t, result.Selected)
		assert.Equal(t, "Consensus position", result.Selected.Content)
		assert.Equal(t, "consensus", result.Selected.ProviderName)
	})
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_SynthesiseResults
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_SynthesiseResults(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	t.Run("empty results", func(t *testing.T) {
		s := ensemble.synthesiseResults(nil, nil)
		assert.Equal(t, "No results were produced.", s)
	})

	t.Run("all failed", func(t *testing.T) {
		results := []AgenticResult{
			{TaskID: "t1", Error: assert.AnError},
		}
		s := ensemble.synthesiseResults(results, nil)
		assert.Equal(t, "All agent tasks failed during execution.", s)
	})

	t.Run("mixed results", func(t *testing.T) {
		results := []AgenticResult{
			{TaskID: "t1", Content: "Result one"},
			{TaskID: "t2", Error: assert.AnError},
			{TaskID: "t3", Content: "Result three"},
		}
		s := ensemble.synthesiseResults(results, nil)
		assert.Contains(t, s, "Result one")
		assert.Contains(t, s, "Result three")
	})

	t.Run("with verification issues", func(t *testing.T) {
		results := []AgenticResult{
			{TaskID: "t1", Content: "Content"},
		}
		v := &AgenticVerificationResult{
			Approved: false,
			Issues:   []string{"missing coverage", "incomplete"},
		}
		s := ensemble.synthesiseResults(results, v)
		assert.Contains(t, s, "Content")
		assert.Contains(t, s, "missing coverage")
		assert.Contains(t, s, "incomplete")
	})
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_BuildToolSummary
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_BuildToolSummary(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	t.Run("empty", func(t *testing.T) {
		assert.Nil(t, ensemble.buildToolSummary(nil))
	})

	t.Run("aggregates by protocol", func(t *testing.T) {
		execs := []AgenticToolExecution{
			{Protocol: "mcp", Operation: "op1"},
			{Protocol: "mcp", Operation: "op2"},
			{Protocol: "rag", Operation: "search"},
		}
		summaries := ensemble.buildToolSummary(execs)
		require.Len(t, summaries, 2)

		counts := make(map[string]int)
		for _, s := range summaries {
			counts[s.Protocol] = s.Count
		}
		assert.Equal(t, 2, counts["mcp"])
		assert.Equal(t, 1, counts["rag"])
	})
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_CalculateConfidence
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_CalculateConfidence(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	assert.Equal(t, 0.5, ensemble.calculateConfidence(nil))
	assert.Equal(t, 0.95, ensemble.calculateConfidence(
		&AgenticVerificationResult{Confidence: 0.95},
	))
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_SelectToolCapableProvider_NilRegistry
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_SelectToolCapableProvider_NilRegistry(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)
	assert.Nil(t, ensemble.selectToolCapableProvider())
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_GetCompleteFunc_NilProvider
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_GetCompleteFunc_NilProvider(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)
	assert.Nil(t, ensemble.getCompleteFunc())
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_VerifyResults_NilVerifier
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_VerifyResults_NilVerifier(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	req := newTestRequest("test")
	results := []AgenticResult{{TaskID: "t1", Content: "ok"}}
	v := ensemble.verifyResults(context.Background(), req, results)
	require.NotNil(t, v)
	assert.True(t, v.Approved)
	assert.Equal(t, 0.5, v.Confidence)
}

// ---------------------------------------------------------------------------
// TestAgenticEnsemble_PlanTasks_NilPlanner
// ---------------------------------------------------------------------------

func TestAgenticEnsemble_PlanTasks_NilPlanner(t *testing.T) {
	logger := newTestEnsembleLogger()
	cfg := newMinimalConfig()
	ensemble := NewAgenticEnsemble(
		nil, nil, nil, nil, nil, nil, cfg, logger,
	)

	_, err := ensemble.planTasks(context.Background(), "some decision")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execution planner not configured")
}
