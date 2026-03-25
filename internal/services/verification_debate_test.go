package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func newTestVerificationDebate() *VerificationDebate {
	return NewVerificationDebate(newTestLogger())
}

func mockCompleteFunc(
	content string, err error,
) func(context.Context, []models.Message) (*models.LLMResponse, error) {
	return func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		if err != nil {
			return nil, err
		}
		return &models.LLMResponse{Content: content}, nil
	}
}

func TestVerificationDebate_AllResultsPass(t *testing.T) {
	vd := newTestVerificationDebate()
	results := []AgenticResult{
		{TaskID: "task-1", AgentID: "agent-1", Content: "Result for task 1"},
		{TaskID: "task-2", AgentID: "agent-2", Content: "Result for task 2"},
	}

	vr, err := vd.Verify(
		context.Background(),
		"Summarize the document",
		results,
		mockCompleteFunc("APPROVED:0.95", nil),
	)

	require.NoError(t, err)
	assert.True(t, vr.Approved)
	assert.InDelta(t, 0.95, vr.Confidence, 0.001)
	assert.Empty(t, vr.FailedTasks)
	assert.Equal(t, "Verification passed", vr.Summary)
}

func TestVerificationDebate_PartialFailure(t *testing.T) {
	vd := newTestVerificationDebate()
	results := []AgenticResult{
		{TaskID: "task-1", AgentID: "agent-1", Content: "Good result"},
		{
			TaskID:  "task-2",
			AgentID: "agent-2",
			Error:   errors.New("provider timeout"),
		},
	}

	vr, err := vd.Verify(
		context.Background(),
		"Analyze code",
		results,
		mockCompleteFunc("APPROVED:0.80", nil),
	)

	require.NoError(t, err)
	assert.True(t, vr.Approved)
	assert.InDelta(t, 0.80, vr.Confidence, 0.001)
	assert.Equal(t, []string{"task-2"}, vr.FailedTasks)
}

func TestVerificationDebate_AllFailed(t *testing.T) {
	vd := newTestVerificationDebate()
	results := []AgenticResult{
		{
			TaskID:  "task-1",
			AgentID: "agent-1",
			Error:   errors.New("timeout"),
		},
		{
			TaskID:  "task-2",
			AgentID: "agent-2",
			Error:   errors.New("rate limited"),
		},
	}

	// completeFunc should NOT be called when all tasks failed.
	called := false
	completeFn := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		called = true
		return &models.LLMResponse{Content: "APPROVED:1.0"}, nil
	}

	vr, err := vd.Verify(
		context.Background(),
		"Do something",
		results,
		completeFn,
	)

	require.NoError(t, err)
	assert.False(t, called, "LLM should not be called when all tasks failed")
	assert.False(t, vr.Approved)
	assert.Equal(t, 0.0, vr.Confidence)
	assert.Equal(t, []string{"task-1", "task-2"}, vr.FailedTasks)
	assert.Contains(t, vr.Issues, "all tasks failed")
	assert.Equal(t, "All agent tasks failed during execution", vr.Summary)
}

func TestVerificationDebate_IssuesDetected(t *testing.T) {
	vd := newTestVerificationDebate()
	results := []AgenticResult{
		{TaskID: "task-1", AgentID: "agent-1", Content: "Partial result"},
	}

	vr, err := vd.Verify(
		context.Background(),
		"Full analysis required",
		results,
		mockCompleteFunc("ISSUES:incomplete coverage|missing security review", nil),
	)

	require.NoError(t, err)
	assert.False(t, vr.Approved)
	assert.InDelta(t, 0.3, vr.Confidence, 0.001)
	assert.Len(t, vr.Issues, 2)
	assert.Equal(t, "incomplete coverage", vr.Issues[0])
	assert.Equal(t, "missing security review", vr.Issues[1])
	assert.Equal(t, "Verification found 2 issues", vr.Summary)
}

func TestVerificationDebate_LLMError(t *testing.T) {
	vd := newTestVerificationDebate()
	results := []AgenticResult{
		{TaskID: "task-1", AgentID: "agent-1", Content: "Some result"},
	}

	vr, err := vd.Verify(
		context.Background(),
		"Check this",
		results,
		mockCompleteFunc("", errors.New("connection refused")),
	)

	require.NoError(t, err)
	assert.True(t, vr.Approved)
	assert.InDelta(t, 0.5, vr.Confidence, 0.001)
	assert.Contains(t, vr.Issues, "verification LLM unavailable")
	assert.Equal(t, "Verification skipped due to LLM error", vr.Summary)
}

func TestVerificationDebate_ParseApproved(t *testing.T) {
	vd := newTestVerificationDebate()

	tests := []struct {
		name       string
		input      string
		confidence float64
	}{
		{
			name:       "with explicit confidence",
			input:      "APPROVED:0.95",
			confidence: 0.95,
		},
		{
			name:       "without confidence value",
			input:      "APPROVED",
			confidence: 0.8,
		},
		{
			name:       "with low confidence",
			input:      "APPROVED:0.60",
			confidence: 0.60,
		},
		{
			name:       "with perfect confidence",
			input:      "APPROVED:1.0",
			confidence: 1.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := vd.parseVerificationResponse(tc.input, nil)
			assert.True(t, result.Approved)
			assert.InDelta(t, tc.confidence, result.Confidence, 0.001)
			assert.Equal(t, "Verification passed", result.Summary)
		})
	}
}

func TestVerificationDebate_ParseIssues(t *testing.T) {
	vd := newTestVerificationDebate()

	tests := []struct {
		name           string
		input          string
		expectedIssues []string
	}{
		{
			name:           "multiple issues",
			input:          "ISSUES:a|b|c",
			expectedIssues: []string{"a", "b", "c"},
		},
		{
			name:           "single issue",
			input:          "ISSUES:something wrong",
			expectedIssues: []string{"something wrong"},
		},
		{
			name:           "issues with whitespace",
			input:          "ISSUES: leading space | trailing space ",
			expectedIssues: []string{"leading space", "trailing space"},
		},
		{
			name:           "unrecognized format treated as single issue",
			input:          "Something unexpected from the LLM",
			expectedIssues: []string{"Something unexpected from the LLM"},
		},
	}

	failedTasks := []string{"task-x"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := vd.parseVerificationResponse(tc.input, failedTasks)
			assert.False(t, result.Approved)
			assert.InDelta(t, 0.3, result.Confidence, 0.001)
			assert.Equal(t, tc.expectedIssues, result.Issues)
			assert.Equal(t, failedTasks, result.FailedTasks)
		})
	}
}

func TestVerificationDebate_TruncateVerification(t *testing.T) {
	assert.Equal(t, "short", truncateVerification("short", 100))
	assert.Equal(t, "ab...", truncateVerification("abcdef", 2))
	assert.Equal(t, "", truncateVerification("", 10))
	assert.Equal(t, "exact", truncateVerification("exact", 5))
}
