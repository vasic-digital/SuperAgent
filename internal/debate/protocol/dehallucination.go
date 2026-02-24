// Package protocol provides the 5-phase debate protocol implementation.
// This file implements the Communicative Dehallucination pre-debate phase,
// inspired by ChatDev's clarification protocol. It uses a two-role LLM
// interaction (assistant generating questions, domain expert answering)
// to resolve ambiguities before the main debate begins.
package protocol

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ClarificationRequest represents a question to clarify task requirements.
type ClarificationRequest struct {
	Question string `json:"question"`
	Category string `json:"category"` // requirements, constraints, edge_cases, performance, integration
	Priority int    `json:"priority"` // 1-5 (5 = highest)
}

// ClarificationResponse represents an answer to a clarification question.
type ClarificationResponse struct {
	Answer               string   `json:"answer"`
	Confidence           float64  `json:"confidence"`
	RemainingAmbiguities []string `json:"remaining_ambiguities"`
}

// DehallucationConfig configures the dehallucination phase.
type DehallucationConfig struct {
	Enabled                bool    `json:"enabled"`
	MaxClarificationRounds int     `json:"max_clarification_rounds"` // default 3
	ConfidenceThreshold    float64 `json:"confidence_threshold"`     // default 0.9
}

// DefaultDehallucationConfig returns sensible defaults.
func DefaultDehallucationConfig() DehallucationConfig {
	return DehallucationConfig{
		Enabled:                true,
		MaxClarificationRounds: 3,
		ConfidenceThreshold:    0.9,
	}
}

// DehallucationResult captures the outcome of the dehallucination phase.
type DehallucationResult struct {
	ClarifiedTask       string                  `json:"clarified_task"`
	ClarificationRounds int                     `json:"clarification_rounds"`
	FinalConfidence     float64                 `json:"final_confidence"`
	Clarifications      []ClarificationRequest  `json:"clarifications"`
	Responses           []ClarificationResponse `json:"responses"`
	Skipped             bool                    `json:"skipped"` // true if confidence was already high
	Duration            time.Duration           `json:"duration"`
}

// DehallucationLLMClient interface for LLM calls in dehallucination.
type DehallucationLLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// DehallucationPhase implements the ChatDev-inspired clarification protocol.
type DehallucationPhase struct {
	config    DehallucationConfig
	llmClient DehallucationLLMClient
}

// NewDehallucationPhase creates a new dehallucination phase instance.
func NewDehallucationPhase(
	config DehallucationConfig,
	llmClient DehallucationLLMClient,
) *DehallucationPhase {
	return &DehallucationPhase{
		config:    config,
		llmClient: llmClient,
	}
}

// Execute runs the communicative dehallucination phase. It generates
// clarification questions and answers them iteratively until the
// confidence threshold is reached or the maximum number of rounds
// is exhausted.
func (d *DehallucationPhase) Execute(
	ctx context.Context,
	task string,
	taskContext map[string]interface{},
) (*DehallucationResult, error) {
	startTime := time.Now()

	// If not enabled, return immediately with Skipped=true.
	if !d.config.Enabled {
		return &DehallucationResult{
			ClarifiedTask: task,
			Skipped:       true,
			Duration:      time.Since(startTime),
		}, nil
	}

	// Initial assessment: generate clarifications with no prior answers.
	questions, confidence, err := d.generateClarifications(ctx, task, nil)
	if err != nil {
		return nil, fmt.Errorf("initial clarification generation failed: %w", err)
	}

	// If confidence is already above threshold, skip the phase.
	if confidence >= d.config.ConfidenceThreshold {
		return &DehallucationResult{
			ClarifiedTask:   task,
			FinalConfidence: confidence,
			Skipped:         true,
			Duration:        time.Since(startTime),
		}, nil
	}

	allClarifications := make([]ClarificationRequest, 0)
	allResponses := make([]ClarificationResponse, 0)
	rounds := 0

	maxRounds := d.config.MaxClarificationRounds
	if maxRounds <= 0 {
		maxRounds = 3
	}

	for round := 0; round < maxRounds; round++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		rounds++

		// Step (a): Generate clarification questions (LLM as assistant).
		if round > 0 {
			questions, confidence, err = d.generateClarifications(
				ctx, task, allResponses,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"clarification generation failed (round %d): %w",
					round+1, err,
				)
			}

			if confidence >= d.config.ConfidenceThreshold {
				break
			}
		}

		if len(questions) == 0 {
			break
		}

		allClarifications = append(allClarifications, questions...)

		// Step (b): Answer clarifications (LLM as domain expert).
		response, err := d.answerClarifications(ctx, task, questions)
		if err != nil {
			return nil, fmt.Errorf(
				"clarification answering failed (round %d): %w",
				round+1, err,
			)
		}

		allResponses = append(allResponses, *response)

		// Step (d): Check if confidence threshold is met.
		if response.Confidence >= d.config.ConfidenceThreshold {
			confidence = response.Confidence
			break
		}

		confidence = response.Confidence
	}

	// Build the clarified task from original + all responses.
	clarifiedTask := d.buildClarifiedTask(task, allResponses)

	return &DehallucationResult{
		ClarifiedTask:       clarifiedTask,
		ClarificationRounds: rounds,
		FinalConfidence:     confidence,
		Clarifications:      allClarifications,
		Responses:           allResponses,
		Skipped:             false,
		Duration:            time.Since(startTime),
	}, nil
}

// generateClarifications uses the LLM (acting as an assistant) to assess
// the current understanding of the task and produce clarification questions.
func (d *DehallucationPhase) generateClarifications(
	ctx context.Context,
	task string,
	priorAnswers []ClarificationResponse,
) ([]ClarificationRequest, float64, error) {
	prompt := d.buildClarificationPrompt(task, priorAnswers)

	response, err := d.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, 0, fmt.Errorf("LLM completion for clarifications failed: %w", err)
	}

	questions, confidence := d.parseClarifications(response)
	return questions, confidence, nil
}

// answerClarifications uses the LLM (acting as a domain expert / instructor)
// to answer the generated clarification questions.
func (d *DehallucationPhase) answerClarifications(
	ctx context.Context,
	task string,
	questions []ClarificationRequest,
) (*ClarificationResponse, error) {
	prompt := d.buildAnswerPrompt(task, questions)

	response, err := d.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM completion for answers failed: %w", err)
	}

	answer := d.parseAnswer(response)
	return answer, nil
}

// buildClarificationPrompt constructs the prompt for the assistant role
// that generates clarification questions about the task.
func (d *DehallucationPhase) buildClarificationPrompt(
	task string,
	priorAnswers []ClarificationResponse,
) string {
	var sb strings.Builder

	sb.WriteString("You are reviewing a task for potential ambiguities. ")
	sb.WriteString("Task: ")
	sb.WriteString(task)
	sb.WriteString("\n\n")

	if len(priorAnswers) > 0 {
		sb.WriteString("Prior clarifications:\n")
		for i, answer := range priorAnswers {
			sb.WriteString(fmt.Sprintf(
				"Round %d Answer: %s (Confidence: %.2f)\n",
				i+1, answer.Answer, answer.Confidence,
			))
			if len(answer.RemainingAmbiguities) > 0 {
				sb.WriteString("Remaining ambiguities: ")
				sb.WriteString(strings.Join(answer.RemainingAmbiguities, "; "))
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Generate questions in this format:\n")
	sb.WriteString("QUESTION: <question>\n")
	sb.WriteString("CATEGORY: <category>\n")
	sb.WriteString("PRIORITY: <1-5>\n")
	sb.WriteString("...\n")
	sb.WriteString("CONFIDENCE: <your overall understanding confidence 0-1>\n")

	return sb.String()
}

// buildAnswerPrompt constructs the prompt for the domain expert role
// that answers clarification questions.
func (d *DehallucationPhase) buildAnswerPrompt(
	task string,
	questions []ClarificationRequest,
) string {
	var sb strings.Builder

	sb.WriteString(
		"You are a domain expert answering questions about a task. ",
	)
	sb.WriteString("Task: ")
	sb.WriteString(task)
	sb.WriteString("\n\nAnswer each question:\n")

	for _, q := range questions {
		sb.WriteString(fmt.Sprintf(
			"QUESTION: %s\n", q.Question,
		))
	}

	sb.WriteString("\nFor each question provide:\n")
	sb.WriteString("ANSWER: <your answer>\n")
	sb.WriteString("REMAINING_AMBIGUITIES: <any remaining unknowns>\n")
	sb.WriteString("OVERALL_CONFIDENCE: <0-1>\n")

	return sb.String()
}

// parseClarifications parses the LLM response into clarification requests
// and an overall confidence value.
func (d *DehallucationPhase) parseClarifications(
	response string,
) ([]ClarificationRequest, float64) {
	questions := make([]ClarificationRequest, 0)
	confidence := 0.0

	lines := strings.Split(response, "\n")

	var currentQuestion string
	var currentCategory string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "QUESTION:") {
			// If we have a pending question, flush it before starting a new one.
			if currentQuestion != "" {
				questions = append(questions, ClarificationRequest{
					Question: currentQuestion,
					Category: normalizeCategory(currentCategory),
					Priority: 3, // default priority
				})
			}
			currentQuestion = strings.TrimSpace(
				strings.TrimPrefix(trimmed, "QUESTION:"),
			)
			currentCategory = ""
			continue
		}

		if strings.HasPrefix(trimmed, "CATEGORY:") {
			currentCategory = strings.TrimSpace(
				strings.TrimPrefix(trimmed, "CATEGORY:"),
			)
			continue
		}

		if strings.HasPrefix(trimmed, "PRIORITY:") {
			priorityStr := strings.TrimSpace(
				strings.TrimPrefix(trimmed, "PRIORITY:"),
			)
			priority, err := strconv.Atoi(priorityStr)
			if err != nil || priority < 1 || priority > 5 {
				priority = 3
			}

			if currentQuestion != "" {
				questions = append(questions, ClarificationRequest{
					Question: currentQuestion,
					Category: normalizeCategory(currentCategory),
					Priority: priority,
				})
				currentQuestion = ""
				currentCategory = ""
			}
			continue
		}

		if strings.HasPrefix(trimmed, "CONFIDENCE:") {
			confStr := strings.TrimSpace(
				strings.TrimPrefix(trimmed, "CONFIDENCE:"),
			)
			parsed, err := strconv.ParseFloat(confStr, 64)
			if err == nil && parsed >= 0 && parsed <= 1 {
				confidence = parsed
			}
			continue
		}
	}

	// Flush any trailing question that was not closed by a PRIORITY line.
	if currentQuestion != "" {
		questions = append(questions, ClarificationRequest{
			Question: currentQuestion,
			Category: normalizeCategory(currentCategory),
			Priority: 3,
		})
	}

	return questions, confidence
}

// parseAnswer parses the LLM response into a ClarificationResponse.
func (d *DehallucationPhase) parseAnswer(response string) *ClarificationResponse {
	result := &ClarificationResponse{
		Confidence:           0.5,
		RemainingAmbiguities: make([]string, 0),
	}

	lines := strings.Split(response, "\n")
	var answers []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "ANSWER:") {
			answer := strings.TrimSpace(
				strings.TrimPrefix(trimmed, "ANSWER:"),
			)
			if answer != "" {
				answers = append(answers, answer)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "REMAINING_AMBIGUITIES:") {
			ambiguities := strings.TrimSpace(
				strings.TrimPrefix(trimmed, "REMAINING_AMBIGUITIES:"),
			)
			if ambiguities != "" && ambiguities != "none" &&
				ambiguities != "None" && ambiguities != "N/A" {
				parts := strings.Split(ambiguities, ";")
				for _, part := range parts {
					p := strings.TrimSpace(part)
					if p != "" {
						result.RemainingAmbiguities = append(
							result.RemainingAmbiguities, p,
						)
					}
				}
			}
			continue
		}

		if strings.HasPrefix(trimmed, "OVERALL_CONFIDENCE:") {
			confStr := strings.TrimSpace(
				strings.TrimPrefix(trimmed, "OVERALL_CONFIDENCE:"),
			)
			parsed, err := strconv.ParseFloat(confStr, 64)
			if err == nil && parsed >= 0 && parsed <= 1 {
				result.Confidence = parsed
			}
			continue
		}
	}

	if len(answers) > 0 {
		result.Answer = strings.Join(answers, " | ")
	} else {
		// Fall back to the entire response if no structured answers found.
		result.Answer = strings.TrimSpace(response)
	}

	return result
}

// buildClarifiedTask constructs the clarified task description by
// combining the original task with all gathered clarification responses.
func (d *DehallucationPhase) buildClarifiedTask(
	task string,
	responses []ClarificationResponse,
) string {
	if len(responses) == 0 {
		return task
	}

	var sb strings.Builder
	sb.WriteString(task)
	sb.WriteString("\n\n--- Clarifications ---\n")

	for i, resp := range responses {
		sb.WriteString(fmt.Sprintf(
			"\nRound %d Clarification:\n%s\n", i+1, resp.Answer,
		))

		if len(resp.RemainingAmbiguities) > 0 {
			sb.WriteString("Note — remaining ambiguities: ")
			sb.WriteString(strings.Join(resp.RemainingAmbiguities, "; "))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// normalizeCategory ensures the category is one of the valid values.
func normalizeCategory(category string) string {
	normalized := strings.ToLower(strings.TrimSpace(category))

	validCategories := map[string]bool{
		"requirements": true,
		"constraints":  true,
		"edge_cases":   true,
		"performance":  true,
		"integration":  true,
	}

	if validCategories[normalized] {
		return normalized
	}

	// Map common aliases.
	aliases := map[string]string{
		"edge cases":  "edge_cases",
		"edgecases":   "edge_cases",
		"requirement": "requirements",
		"constraint":  "constraints",
	}

	if mapped, ok := aliases[normalized]; ok {
		return mapped
	}

	// Default to requirements for unrecognized categories.
	return "requirements"
}
