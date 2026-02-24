package reflexion

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ReflexionConfig configures the reflexion loop behavior.
type ReflexionConfig struct {
	MaxAttempts         int           `json:"max_attempts"`
	ConfidenceThreshold float64       `json:"confidence_threshold"`
	Timeout             time.Duration `json:"timeout"`
}

// DefaultReflexionConfig returns sensible defaults.
func DefaultReflexionConfig() ReflexionConfig {
	return ReflexionConfig{
		MaxAttempts:         3,
		ConfidenceThreshold: 0.95,
		Timeout:             5 * time.Minute,
	}
}

// TestResult represents the outcome of a test execution.
type TestResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// TestExecutor executes tests against code.
type TestExecutor interface {
	Execute(ctx context.Context, code string, language string) (
		[]*TestResult, error,
	)
}

// CodeGenerator generates code based on task and prior reflections.
type CodeGenerator func(
	ctx context.Context,
	task string,
	priorReflections []*Reflection,
) (string, error)

// ReflexionTask describes what the reflexion loop should work on.
type ReflexionTask struct {
	Description   string        `json:"description"`
	InitialCode   string        `json:"initial_code"`
	Language      string        `json:"language"`
	AgentID       string        `json:"agent_id"`
	SessionID     string        `json:"session_id"`
	CodeGenerator CodeGenerator `json:"-"`
}

// ReflexionResult captures the complete outcome of a reflexion loop
// execution.
type ReflexionResult struct {
	FinalCode       string        `json:"final_code"`
	Attempts        int           `json:"attempts"`
	AllPassed       bool          `json:"all_passed"`
	Reflections     []*Reflection `json:"reflections"`
	Episodes        []*Episode    `json:"episodes"`
	FinalConfidence float64       `json:"final_confidence"`
	Duration        time.Duration `json:"duration"`
	TestResults     []*TestResult `json:"test_results"`
}

// ReflexionLoop implements the Reflexion verbal reinforcement learning
// cycle. It iteratively generates code, runs tests, reflects on failures,
// and uses accumulated reflections to guide subsequent attempts.
type ReflexionLoop struct {
	config    ReflexionConfig
	generator *ReflectionGenerator
	executor  TestExecutor
	memory    *EpisodicMemoryBuffer
}

// NewReflexionLoop creates a new ReflexionLoop with the given
// configuration, reflection generator, test executor, and episodic
// memory buffer.
func NewReflexionLoop(
	config ReflexionConfig,
	generator *ReflectionGenerator,
	executor TestExecutor,
	memory *EpisodicMemoryBuffer,
) *ReflexionLoop {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultReflexionConfig().MaxAttempts
	}
	if config.ConfidenceThreshold <= 0 {
		config.ConfidenceThreshold = DefaultReflexionConfig().ConfidenceThreshold
	}
	if config.Timeout <= 0 {
		config.Timeout = DefaultReflexionConfig().Timeout
	}
	return &ReflexionLoop{
		config:    config,
		generator: generator,
		executor:  executor,
		memory:    memory,
	}
}

// Execute runs the reflexion loop for the given task. It iteratively
// generates code, runs tests, generates reflections on failures, stores
// episodes, and retries until all tests pass, the confidence threshold
// is met, or the maximum number of attempts is exhausted.
func (l *ReflexionLoop) Execute(
	ctx context.Context,
	task *ReflexionTask,
) (*ReflexionResult, error) {
	if task == nil {
		return nil, fmt.Errorf("reflexion task must not be nil")
	}
	if l.executor == nil {
		return nil, fmt.Errorf("test executor must not be nil")
	}

	start := time.Now()

	// Apply timeout to the context.
	ctx, cancel := context.WithTimeout(ctx, l.config.Timeout)
	defer cancel()

	var (
		reflections []*Reflection
		episodes    []*Episode
		currentCode = task.InitialCode
	)

	// If no initial code is provided, generate it.
	if currentCode == "" {
		if task.CodeGenerator == nil {
			return nil, fmt.Errorf(
				"neither initial code nor code generator provided",
			)
		}
		generated, err := task.CodeGenerator(
			ctx, task.Description, nil,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"initial code generation failed: %w", err,
			)
		}
		currentCode = generated
	}

	var lastTestResults []*TestResult

	for attempt := 1; attempt <= l.config.MaxAttempts; attempt++ {
		// Check for context cancellation / timeout.
		if err := ctx.Err(); err != nil {
			return &ReflexionResult{
				FinalCode:       currentCode,
				Attempts:        attempt - 1,
				AllPassed:       false,
				Reflections:     reflections,
				Episodes:        episodes,
				FinalConfidence: lastReflectionConfidence(reflections),
				Duration:        time.Since(start),
				TestResults:     lastTestResults,
			}, fmt.Errorf("reflexion loop timed out: %w", err)
		}

		// Run tests against the current code.
		testResults, err := l.executor.Execute(
			ctx, currentCode, task.Language,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"test execution failed on attempt %d: %w", attempt, err,
			)
		}
		lastTestResults = testResults

		allPassed, confidence := evaluateTestResults(testResults)

		// Success: all tests passed.
		if allPassed {
			// If there is no code generator (static code), or
			// confidence meets threshold, return success.
			if task.CodeGenerator == nil ||
				confidence >= l.config.ConfidenceThreshold {
				return &ReflexionResult{
					FinalCode:       currentCode,
					Attempts:        attempt,
					AllPassed:       true,
					Reflections:     reflections,
					Episodes:        episodes,
					FinalConfidence: confidence,
					Duration:        time.Since(start),
					TestResults:     testResults,
				}, nil
			}
		}

		// Tests failed (or confidence below threshold with generator).
		// If no code generator, we cannot improve -- return failure.
		if task.CodeGenerator == nil {
			return &ReflexionResult{
				FinalCode:       currentCode,
				Attempts:        attempt,
				AllPassed:       allPassed,
				Reflections:     reflections,
				Episodes:        episodes,
				FinalConfidence: confidence,
				Duration:        time.Since(start),
				TestResults:     testResults,
			}, nil
		}

		// Collect error messages from failed tests.
		errorMessages := collectErrors(testResults)

		// Build the test results map for episode storage.
		testResultsMap := buildTestResultsMap(testResults)

		// Generate a reflection.
		reflectionReq := &ReflectionRequest{
			Code:             currentCode,
			TestResults:      testResultsMap,
			ErrorMessages:    errorMessages,
			PriorReflections: reflections,
			TaskDescription:  task.Description,
			AttemptNumber:    attempt,
		}

		reflection, err := l.generateReflection(ctx, reflectionReq)
		if err != nil {
			return nil, fmt.Errorf(
				"reflection generation failed on attempt %d: %w",
				attempt, err,
			)
		}

		// Build and store the episode.
		episode := &Episode{
			SessionID:       task.SessionID,
			AgentID:         task.AgentID,
			TaskDescription: task.Description,
			AttemptNumber:   attempt,
			Code:            currentCode,
			TestResults:     testResultsMap,
			FailureAnalysis: buildFailureAnalysis(errorMessages),
			Reflection:      reflection,
			Improvement:     reflection.WhatToChangeNext,
			Confidence:      reflection.ConfidenceInFix,
			Timestamp:       time.Now(),
		}

		if l.memory != nil {
			if storeErr := l.memory.Store(episode); storeErr != nil {
				// Non-fatal: log but continue.
				_ = storeErr
			}
		}

		episodes = append(episodes, episode)
		reflections = append(reflections, reflection)

		// If max attempts reached, return with accumulated data.
		if attempt >= l.config.MaxAttempts {
			return &ReflexionResult{
				FinalCode:       currentCode,
				Attempts:        attempt,
				AllPassed:       false,
				Reflections:     reflections,
				Episodes:        episodes,
				FinalConfidence: reflection.ConfidenceInFix,
				Duration:        time.Since(start),
				TestResults:     testResults,
			}, nil
		}

		// Generate improved code using accumulated reflections.
		newCode, genErr := task.CodeGenerator(
			ctx, task.Description, reflections,
		)
		if genErr != nil {
			return nil, fmt.Errorf(
				"code generation failed on attempt %d: %w",
				attempt+1, genErr,
			)
		}
		currentCode = newCode
	}

	// This is unreachable given the loop logic, but satisfies the
	// compiler.
	return &ReflexionResult{
		FinalCode:       currentCode,
		Attempts:        l.config.MaxAttempts,
		AllPassed:       false,
		Reflections:     reflections,
		Episodes:        episodes,
		FinalConfidence: lastReflectionConfidence(reflections),
		Duration:        time.Since(start),
		TestResults:     lastTestResults,
	}, nil
}

// generateReflection uses the configured generator if available, or
// falls back to a basic manual reflection built from error analysis.
func (l *ReflexionLoop) generateReflection(
	ctx context.Context,
	req *ReflectionRequest,
) (*Reflection, error) {
	if l.generator != nil {
		return l.generator.Generate(ctx, req)
	}
	return buildManualReflection(req), nil
}

// buildManualReflection creates a basic reflection by analyzing test
// failure patterns when no ReflectionGenerator is available.
func buildManualReflection(req *ReflectionRequest) *Reflection {
	rootCause := "Test failures detected in code"
	whatWentWrong := "One or more tests did not pass"
	whatToChange := "Review failing tests and fix the underlying issues"
	confidence := 0.3

	if len(req.ErrorMessages) > 0 {
		combined := strings.Join(req.ErrorMessages, " ")
		lower := strings.ToLower(combined)

		switch {
		case containsAny(lower, "compile", "syntax", "undefined"):
			rootCause = "Compilation or syntax error"
			whatWentWrong = "Code contains syntax errors"
			whatToChange = "Fix syntax errors and undefined references"
			confidence = 0.5
		case containsAny(lower, "assert", "expected", "got", "mismatch"):
			rootCause = "Logic error causing assertion failure"
			whatWentWrong = "Code output does not match expected values"
			whatToChange = "Correct the algorithm logic"
			confidence = 0.4
		case containsAny(lower, "timeout", "deadline"):
			rootCause = "Execution timeout"
			whatWentWrong = "Code runs too slowly or has infinite loop"
			whatToChange = "Optimize algorithm or add termination"
			confidence = 0.4
		case containsAny(lower, "nil", "null", "panic"):
			rootCause = "Nil pointer or runtime panic"
			whatWentWrong = "Code dereferences nil value"
			whatToChange = "Add nil guards before pointer access"
			confidence = 0.5
		}
	}

	if len(req.PriorReflections) > 0 {
		last := req.PriorReflections[len(req.PriorReflections)-1]
		whatToChange = fmt.Sprintf(
			"Previous suggestion (%s) was insufficient; %s",
			last.WhatToChangeNext, whatToChange,
		)
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

// evaluateTestResults computes whether all tests passed and a
// confidence score. Confidence is the ratio of passing tests.
func evaluateTestResults(results []*TestResult) (bool, float64) {
	if len(results) == 0 {
		return true, 1.0
	}

	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	allPassed := passed == len(results)
	confidence := float64(passed) / float64(len(results))
	return allPassed, confidence
}

// collectErrors extracts error messages from failed test results.
func collectErrors(results []*TestResult) []string {
	var errors []string
	for _, r := range results {
		if !r.Passed && r.Error != "" {
			errors = append(errors, fmt.Sprintf(
				"[%s] %s", r.Name, r.Error,
			))
		}
	}
	return errors
}

// buildTestResultsMap converts test results into a map suitable for
// episode storage.
func buildTestResultsMap(
	results []*TestResult,
) map[string]interface{} {
	m := make(map[string]interface{})
	for _, r := range results {
		m[r.Name] = map[string]interface{}{
			"passed":   r.Passed,
			"output":   r.Output,
			"error":    r.Error,
			"duration": r.Duration.String(),
		}
	}
	return m
}

// buildFailureAnalysis produces a human-readable summary of the errors.
func buildFailureAnalysis(errors []string) string {
	if len(errors) == 0 {
		return "No errors captured"
	}
	return fmt.Sprintf(
		"%d test failure(s): %s",
		len(errors), strings.Join(errors, "; "),
	)
}

// lastReflectionConfidence returns the confidence from the most recent
// reflection, or 0.0 if there are no reflections.
func lastReflectionConfidence(reflections []*Reflection) float64 {
	if len(reflections) == 0 {
		return 0.0
	}
	return reflections[len(reflections)-1].ConfidenceInFix
}
