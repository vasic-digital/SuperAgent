// Package reflexion implements the Reflexion framework for iterative
// self-improvement through verbal reinforcement. It provides episodic memory,
// reflection generation, and evaluation capabilities for code generation tasks.
package reflexion

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LLMClient defines the interface for language model interaction
// used by the reflection generator to analyze failed attempts.
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// ReflectionRequest contains all context needed to generate a reflection
// about a failed code attempt.
type ReflectionRequest struct {
	// Code is the source code of the failed attempt.
	Code string `json:"code"`
	// TestResults contains structured test outcome data.
	TestResults map[string]interface{} `json:"test_results"`
	// ErrorMessages contains error strings from the failed attempt.
	ErrorMessages []string `json:"error_messages"`
	// PriorReflections contains reflections from previous attempts.
	PriorReflections []*Reflection `json:"prior_reflections"`
	// TaskDescription describes the task the code is attempting.
	TaskDescription string `json:"task_description"`
	// AttemptNumber is the current attempt number (1-based).
	AttemptNumber int `json:"attempt_number"`
}

// ReflectionGenerator uses an LLM to analyze failed code attempts and
// produce structured reflections that guide subsequent attempts.
type ReflectionGenerator struct {
	llmClient LLMClient
}

// NewReflectionGenerator creates a new ReflectionGenerator with the
// provided LLM client.
func NewReflectionGenerator(llmClient LLMClient) *ReflectionGenerator {
	return &ReflectionGenerator{
		llmClient: llmClient,
	}
}

// Generate produces a Reflection by analyzing the failed attempt described
// in the request. It uses the LLM client to perform analysis and falls back
// to deterministic generation if the LLM call fails.
func (rg *ReflectionGenerator) Generate(
	ctx context.Context,
	req *ReflectionRequest,
) (*Reflection, error) {
	if req == nil {
		return nil, fmt.Errorf("reflection request must not be nil")
	}

	prompt := rg.buildReflectionPrompt(req)

	response, err := rg.llmClient.Complete(ctx, prompt)
	if err != nil {
		// Fall back to deterministic reflection when LLM is unavailable.
		fallback := rg.generateFallbackReflection(req)
		return fallback, nil
	}

	reflection, err := rg.parseReflectionResponse(response)
	if err != nil {
		// LLM returned unparseable output; use fallback.
		fallback := rg.generateFallbackReflection(req)
		return fallback, nil
	}

	return reflection, nil
}

// buildReflectionPrompt constructs the analysis prompt sent to the LLM.
func (rg *ReflectionGenerator) buildReflectionPrompt(
	req *ReflectionRequest,
) string {
	var sb strings.Builder

	sb.WriteString("You are analyzing a failed code attempt. " +
		"Provide your analysis in exactly this format:\n")
	sb.WriteString("ROOT_CAUSE: <one sentence>\n")
	sb.WriteString("WHAT_WENT_WRONG: <one sentence>\n")
	sb.WriteString("WHAT_TO_CHANGE: <one sentence>\n")
	sb.WriteString("CONFIDENCE: <0.0-1.0>\n")
	sb.WriteString("\nContext:\n")

	sb.WriteString(fmt.Sprintf("Task: %s\n", req.TaskDescription))
	sb.WriteString(fmt.Sprintf("Attempt: %d\n", req.AttemptNumber))
	sb.WriteString(fmt.Sprintf("Code: %s\n", req.Code))

	testResultsJSON, err := json.Marshal(req.TestResults)
	if err != nil {
		sb.WriteString(fmt.Sprintf("Test Results: %v\n", req.TestResults))
	} else {
		sb.WriteString(fmt.Sprintf("Test Results: %s\n", string(testResultsJSON)))
	}

	sb.WriteString(fmt.Sprintf(
		"Errors: %s\n",
		strings.Join(req.ErrorMessages, "\n"),
	))

	if len(req.PriorReflections) > 0 {
		sb.WriteString("Prior Reflections:\n")
		for i, r := range req.PriorReflections {
			sb.WriteString(fmt.Sprintf(
				"  Reflection %d: root_cause=%q, what_went_wrong=%q, "+
					"what_to_change=%q, confidence=%.2f\n",
				i+1, r.RootCause, r.WhatWentWrong,
				r.WhatToChangeNext, r.ConfidenceInFix,
			))
		}
	} else {
		sb.WriteString("Prior Reflections: None\n")
	}

	return sb.String()
}

// parseReflectionResponse parses the structured LLM response into a
// Reflection. It expects exactly four labelled lines.
func (rg *ReflectionGenerator) parseReflectionResponse(
	response string,
) (*Reflection, error) {
	var (
		rootCause      string
		whatWentWrong  string
		whatToChange   string
		confidence     float64
		foundRoot      bool
		foundWrong     bool
		foundChange    bool
		foundConfidence bool
	)

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "ROOT_CAUSE:"):
			rootCause = strings.TrimSpace(
				strings.TrimPrefix(trimmed, "ROOT_CAUSE:"),
			)
			foundRoot = true

		case strings.HasPrefix(trimmed, "WHAT_WENT_WRONG:"):
			whatWentWrong = strings.TrimSpace(
				strings.TrimPrefix(trimmed, "WHAT_WENT_WRONG:"),
			)
			foundWrong = true

		case strings.HasPrefix(trimmed, "WHAT_TO_CHANGE:"):
			whatToChange = strings.TrimSpace(
				strings.TrimPrefix(trimmed, "WHAT_TO_CHANGE:"),
			)
			foundChange = true

		case strings.HasPrefix(trimmed, "CONFIDENCE:"):
			raw := strings.TrimSpace(
				strings.TrimPrefix(trimmed, "CONFIDENCE:"),
			)
			parsed, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse confidence value %q: %w", raw, err,
				)
			}
			confidence = parsed
			foundConfidence = true
		}
	}

	if !foundRoot || !foundWrong || !foundChange || !foundConfidence {
		missing := make([]string, 0, 4)
		if !foundRoot {
			missing = append(missing, "ROOT_CAUSE")
		}
		if !foundWrong {
			missing = append(missing, "WHAT_WENT_WRONG")
		}
		if !foundChange {
			missing = append(missing, "WHAT_TO_CHANGE")
		}
		if !foundConfidence {
			missing = append(missing, "CONFIDENCE")
		}
		return nil, fmt.Errorf(
			"reflection response missing required fields: %s",
			strings.Join(missing, ", "),
		)
	}

	return &Reflection{
		RootCause:        rootCause,
		WhatWentWrong:    whatWentWrong,
		WhatToChangeNext: whatToChange,
		ConfidenceInFix:  confidence,
		GeneratedAt:      time.Now(),
	}, nil
}

// generateFallbackReflection produces a deterministic reflection by
// analysing error message patterns when the LLM is unavailable.
func (rg *ReflectionGenerator) generateFallbackReflection(
	req *ReflectionRequest,
) *Reflection {
	rootCause := "Unknown error in code attempt"
	whatWentWrong := "The code did not pass the required tests"
	whatToChange := "Review the code logic and fix identified issues"
	confidence := 0.3

	// Analyze error messages to identify common patterns.
	errors := strings.Join(req.ErrorMessages, " ")
	errorsLower := strings.ToLower(errors)

	switch {
	case containsAny(errorsLower, "compile", "syntax", "unexpected",
		"undefined", "cannot find", "undeclared"):
		rootCause = "Compilation or syntax error in the code"
		whatWentWrong = "The code contains syntax errors preventing compilation"
		whatToChange = "Fix syntax errors and ensure all identifiers are defined"
		confidence = 0.6

	case containsAny(errorsLower, "test fail", "assert", "expected",
		"got", "mismatch", "not equal"):
		rootCause = "Test assertion failure due to incorrect logic"
		whatWentWrong = "The code logic does not produce expected output"
		whatToChange = "Re-examine the algorithm and correct the logic errors"
		confidence = 0.5

	case containsAny(errorsLower, "timeout", "deadline", "timed out",
		"context deadline exceeded"):
		rootCause = "Code execution exceeded the time limit"
		whatWentWrong = "The algorithm is too slow or contains an infinite loop"
		whatToChange = "Optimize the algorithm or add termination conditions"
		confidence = 0.5

	case containsAny(errorsLower, "nil pointer", "null pointer",
		"segmentation fault", "nil dereference"):
		rootCause = "Null/nil pointer dereference at runtime"
		whatWentWrong = "The code accesses a nil reference without a guard"
		whatToChange = "Add nil checks before dereferencing pointers"
		confidence = 0.6

	case containsAny(errorsLower, "index out of range", "out of bounds",
		"array index", "slice bounds"):
		rootCause = "Array or slice index out of bounds"
		whatWentWrong = "The code accesses an index beyond the collection length"
		whatToChange = "Add bounds checking before array/slice access"
		confidence = 0.6

	case containsAny(errorsLower, "permission", "denied", "forbidden",
		"unauthorized"):
		rootCause = "Permission or authorization error"
		whatWentWrong = "The code lacks required permissions for an operation"
		whatToChange = "Ensure proper permissions and credentials are configured"
		confidence = 0.4

	case containsAny(errorsLower, "import", "module", "package",
		"dependency"):
		rootCause = "Missing or incorrect import/dependency"
		whatWentWrong = "A required package or module is not available"
		whatToChange = "Add missing imports and verify dependency availability"
		confidence = 0.6

	case containsAny(errorsLower, "type", "conversion", "cast",
		"incompatible"):
		rootCause = "Type mismatch or invalid type conversion"
		whatWentWrong = "The code uses incompatible types in an operation"
		whatToChange = "Fix type annotations and ensure compatible type usage"
		confidence = 0.5

	case containsAny(errorsLower, "deadlock", "goroutine", "race",
		"concurrent"):
		rootCause = "Concurrency issue such as deadlock or race condition"
		whatWentWrong = "The code has unsafe concurrent access patterns"
		whatToChange = "Add proper synchronization with mutexes or channels"
		confidence = 0.4

	case containsAny(errorsLower, "memory", "allocation", "oom",
		"out of memory"):
		rootCause = "Excessive memory usage or allocation failure"
		whatWentWrong = "The code allocates more memory than available"
		whatToChange = "Reduce memory usage with streaming or smaller buffers"
		confidence = 0.4
	}

	// Incorporate learning from prior reflections if available.
	if len(req.PriorReflections) > 0 {
		last := req.PriorReflections[len(req.PriorReflections)-1]
		whatToChange = fmt.Sprintf(
			"Previous fix (%s) was insufficient; %s",
			last.WhatToChangeNext, whatToChange,
		)
		// Reduce confidence when repeated failures occur.
		confidence *= 0.9
	}

	return &Reflection{
		RootCause:        rootCause,
		WhatWentWrong:    whatWentWrong,
		WhatToChangeNext: whatToChange,
		ConfidenceInFix:  confidence,
		GeneratedAt:      time.Now(),
	}
}

// containsAny returns true if s contains any of the given substrings.
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
