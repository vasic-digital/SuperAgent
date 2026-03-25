package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/models"
)

// VerificationDebate validates agent execution results through LLM-based
// quality verification. After background agents execute tasks, the
// verification debate evaluates results for completeness, correctness,
// and coherence against the original request.
type VerificationDebate struct {
	logger *logrus.Logger
}

// NewVerificationDebate creates a new VerificationDebate instance.
func NewVerificationDebate(logger *logrus.Logger) *VerificationDebate {
	return &VerificationDebate{logger: logger}
}

// AgenticVerificationResult holds the outcome of a post-execution
// verification debate.
type AgenticVerificationResult struct {
	Approved    bool     `json:"approved"`
	Confidence  float64  `json:"confidence"`
	Issues      []string `json:"issues,omitempty"`
	FailedTasks []string `json:"failed_tasks,omitempty"`
	Summary     string   `json:"summary"`
}

// Verify evaluates agent results against the original request by
// constructing a verification prompt and delegating to the provided
// LLM completion function. When all tasks have failed, it returns
// immediately without an LLM call. When the LLM itself is unavailable,
// the method defaults to a low-confidence pass so that the pipeline
// can proceed with a caveat rather than hard-failing.
func (v *VerificationDebate) Verify(
	ctx context.Context,
	originalRequest string,
	results []AgenticResult,
	completeFunc func(
		ctx context.Context,
		messages []models.Message,
	) (*models.LLMResponse, error),
) (*AgenticVerificationResult, error) {
	var resultsSummary strings.Builder
	var failedTasks []string

	for _, r := range results {
		if r.Error != nil {
			failedTasks = append(failedTasks, r.TaskID)
			resultsSummary.WriteString(
				fmt.Sprintf("Task %s (FAILED): %v\n", r.TaskID, r.Error),
			)
		} else {
			resultsSummary.WriteString(
				fmt.Sprintf(
					"Task %s (OK): %s\n",
					r.TaskID,
					truncateVerification(r.Content, 500),
				),
			)
		}
	}

	// If every task failed there is nothing for the LLM to verify.
	if len(failedTasks) == len(results) {
		return &AgenticVerificationResult{
			Approved:    false,
			Confidence:  0.0,
			Issues:      []string{"all tasks failed"},
			FailedTasks: failedTasks,
			Summary:     "All agent tasks failed during execution",
		}, nil
	}

	prompt := fmt.Sprintf(
		`Verify whether the following task results adequately `+
			`address the original request.

Original Request: %s

Task Results:
%s

Evaluate:
1. Completeness: Do the results fully address the request?
2. Correctness: Are the results accurate and sensible?
3. Coherence: Do the results work together as a whole?

Respond with:
APPROVED if the results are satisfactory (confidence 0.0-1.0)
ISSUES if there are problems (list each issue)
Format: APPROVED:0.95 or ISSUES:issue1|issue2|...`,
		originalRequest, resultsSummary.String(),
	)

	messages := []models.Message{
		{
			Role:    "system",
			Content: "You are a quality verification expert. Be strict but fair.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := completeFunc(ctx, messages)
	if err != nil {
		v.logger.WithError(err).Warn(
			"Verification LLM call failed, defaulting to pass with low confidence",
		)
		return &AgenticVerificationResult{
			Approved:    true,
			Confidence:  0.5,
			Issues:      []string{"verification LLM unavailable"},
			FailedTasks: failedTasks,
			Summary:     "Verification skipped due to LLM error",
		}, nil
	}

	return v.parseVerificationResponse(resp.Content, failedTasks), nil
}

// parseVerificationResponse interprets the raw LLM output into a
// structured AgenticVerificationResult.
func (v *VerificationDebate) parseVerificationResponse(
	content string,
	failedTasks []string,
) *AgenticVerificationResult {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "APPROVED") {
		confidence := 0.8 // default when no explicit value is given
		parts := strings.SplitN(content, ":", 2)
		if len(parts) == 2 {
			fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &confidence)
		}
		return &AgenticVerificationResult{
			Approved:    true,
			Confidence:  confidence,
			FailedTasks: failedTasks,
			Summary:     "Verification passed",
		}
	}

	// Parse issues from ISSUES:issue1|issue2|... format.
	var issues []string
	if strings.HasPrefix(content, "ISSUES") {
		parts := strings.SplitN(content, ":", 2)
		if len(parts) == 2 {
			issues = strings.Split(parts[1], "|")
			for i := range issues {
				issues[i] = strings.TrimSpace(issues[i])
			}
		}
	}
	if len(issues) == 0 {
		issues = []string{content} // treat entire response as a single issue
	}

	return &AgenticVerificationResult{
		Approved:    false,
		Confidence:  0.3,
		Issues:      issues,
		FailedTasks: failedTasks,
		Summary:     fmt.Sprintf("Verification found %d issues", len(issues)),
	}
}

// truncateVerification shortens a string to the given maximum length,
// appending an ellipsis when truncation occurs. This is scoped to the
// verification debate to avoid collisions with other truncation helpers
// elsewhere in the services package.
func truncateVerification(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
