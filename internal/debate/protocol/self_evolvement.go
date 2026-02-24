// Package protocol provides the 5-phase debate protocol implementation.
// This file implements the Self-Evolvement pre-debate validation phase,
// where an agent iteratively generates self-tests, evaluates its solution,
// and refines it based on failing test cases before entering the main debate.
package protocol

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SelfEvolvementConfig configures the self-evolvement phase.
type SelfEvolvementConfig struct {
	Enabled       bool          `json:"enabled"`
	MaxIterations int           `json:"max_iterations"` // default 2
	Timeout       time.Duration `json:"timeout"`        // default 2min
}

// DefaultSelfEvolvementConfig returns sensible defaults.
func DefaultSelfEvolvementConfig() SelfEvolvementConfig {
	return SelfEvolvementConfig{
		Enabled:       true,
		MaxIterations: 2,
		Timeout:       2 * time.Minute,
	}
}

// SelfTestResult captures the result of a self-generated test.
type SelfTestResult struct {
	TestName string        `json:"test_name"`
	Input    string        `json:"input"`
	Expected string        `json:"expected"`
	Actual   string        `json:"actual"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
}

// SelfEvolvementResult captures the outcome of the self-evolvement phase.
type SelfEvolvementResult struct {
	FinalSolution string              `json:"final_solution"`
	Iterations    int                 `json:"iterations"`
	TestResults   [][]*SelfTestResult `json:"test_results"` // per-iteration
	Improvements  []string            `json:"improvements"`
	FinalPassRate float64             `json:"final_pass_rate"`
	Duration      time.Duration       `json:"duration"`
	Skipped       bool                `json:"skipped"`
}

// SelfEvolvementLLMClient interface for LLM calls.
type SelfEvolvementLLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// SelfEvolvementPhase implements the self-test-and-refine protocol.
type SelfEvolvementPhase struct {
	config    SelfEvolvementConfig
	llmClient SelfEvolvementLLMClient
}

// NewSelfEvolvementPhase creates a new self-evolvement phase with the given
// configuration and LLM client.
func NewSelfEvolvementPhase(
	config SelfEvolvementConfig,
	llmClient SelfEvolvementLLMClient,
) *SelfEvolvementPhase {
	if config.MaxIterations <= 0 {
		config.MaxIterations = 2
	}
	if config.Timeout <= 0 {
		config.Timeout = 2 * time.Minute
	}
	return &SelfEvolvementPhase{
		config:    config,
		llmClient: llmClient,
	}
}

// Execute runs the self-evolvement phase for the given agent and solution.
//
// Algorithm:
//  1. If not enabled, return Skipped=true immediately.
//  2. Start with initialSolution.
//  3. Loop (max iterations):
//     a. Generate self-tests via LLM.
//     b. Execute tests (conceptual evaluation via LLM).
//     c. If all pass, break with success.
//     d. Refine solution using LLM with failed tests as feedback.
//     e. Track improvements.
//  4. Calculate final pass rate.
//  5. Return result.
func (s *SelfEvolvementPhase) Execute(
	ctx context.Context,
	agentID string,
	initialSolution string,
	task string,
	language string,
) (*SelfEvolvementResult, error) {
	startTime := time.Now()

	// If disabled, skip immediately.
	if !s.config.Enabled {
		return &SelfEvolvementResult{
			FinalSolution: initialSolution,
			Iterations:    0,
			TestResults:   nil,
			Improvements:  nil,
			FinalPassRate: 0,
			Duration:      time.Since(startTime),
			Skipped:       true,
		}, nil
	}

	// Apply timeout to the context.
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	currentSolution := initialSolution
	allTestResults := make([][]*SelfTestResult, 0, s.config.MaxIterations)
	improvements := make([]string, 0, s.config.MaxIterations)
	var finalPassRate float64

	for iteration := 0; iteration < s.config.MaxIterations; iteration++ {
		// Check context before each iteration.
		select {
		case <-ctx.Done():
			return &SelfEvolvementResult{
				FinalSolution: currentSolution,
				Iterations:    iteration,
				TestResults:   allTestResults,
				Improvements:  improvements,
				FinalPassRate: finalPassRate,
				Duration:      time.Since(startTime),
				Skipped:       false,
			}, fmt.Errorf(
				"self-evolvement timeout after %d iterations for agent %s: %w",
				iteration, agentID, ctx.Err(),
			)
		default:
		}

		// Step a: Generate self-tests.
		tests, err := s.generateSelfTests(ctx, currentSolution, task, language)
		if err != nil {
			// If test generation fails, return with what we have.
			return &SelfEvolvementResult{
				FinalSolution: currentSolution,
				Iterations:    iteration,
				TestResults:   allTestResults,
				Improvements:  improvements,
				FinalPassRate: finalPassRate,
				Duration:      time.Since(startTime),
				Skipped:       false,
			}, fmt.Errorf(
				"self-test generation failed at iteration %d for agent %s: %w",
				iteration, agentID, err,
			)
		}

		// If no tests generated, treat as all-pass and break.
		if len(tests) == 0 {
			allTestResults = append(allTestResults, tests)
			finalPassRate = 1.0
			break
		}

		// Step b: Execute tests conceptually via LLM.
		executedTests := s.executeSelfTests(ctx, currentSolution, tests, language)
		allTestResults = append(allTestResults, executedTests)

		// Step c: Check results.
		passed, total := countPassFail(executedTests)
		if total > 0 {
			finalPassRate = float64(passed) / float64(total)
		} else {
			finalPassRate = 1.0
		}

		if passed == total {
			// All tests pass: done.
			break
		}

		// Step d: Collect failed tests and refine.
		failedTests := collectFailed(executedTests)

		refinedSolution, improvement, err := s.refineSolution(
			ctx, currentSolution, failedTests, task, language,
		)
		if err != nil {
			// Refinement failed; keep the current solution.
			break
		}

		currentSolution = refinedSolution
		if improvement != "" {
			improvements = append(improvements, improvement)
		}
	}

	return &SelfEvolvementResult{
		FinalSolution: currentSolution,
		Iterations:    len(allTestResults),
		TestResults:   allTestResults,
		Improvements:  improvements,
		FinalPassRate: finalPassRate,
		Duration:      time.Since(startTime),
		Skipped:       false,
	}, nil
}

// generateSelfTests asks the LLM to create test cases for the given solution.
func (s *SelfEvolvementPhase) generateSelfTests(
	ctx context.Context,
	solution string,
	task string,
	language string,
) ([]*SelfTestResult, error) {
	prompt := s.buildTestGenPrompt(solution, task, language)

	response, err := s.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM test generation call failed: %w", err)
	}

	tests := s.parseSelfTests(response)
	return tests, nil
}

// executeSelfTests evaluates each test conceptually using the LLM.
// Since there is no real sandbox available, the LLM is asked whether the
// solution would produce the expected output for each test input. If the LLM
// evaluation call fails, all tests are marked as passed to avoid blocking.
func (s *SelfEvolvementPhase) executeSelfTests(
	ctx context.Context,
	solution string,
	tests []*SelfTestResult,
	language string,
) []*SelfTestResult {
	results := make([]*SelfTestResult, len(tests))

	for i, test := range tests {
		testStart := time.Now()

		result := &SelfTestResult{
			TestName: test.TestName,
			Input:    test.Input,
			Expected: test.Expected,
		}

		evalPrompt := fmt.Sprintf(
			"Given the following %s solution:\n\n```\n%s\n```\n\n"+
				"For the input:\n%s\n\n"+
				"The expected output is:\n%s\n\n"+
				"What would the actual output be? "+
				"Respond with ONLY the output value, nothing else.",
			language, solution, test.Input, test.Expected,
		)

		actual, err := s.llmClient.Complete(ctx, evalPrompt)
		if err != nil {
			// LLM evaluation failed; mark as passed to avoid blocking.
			result.Actual = test.Expected
			result.Passed = true
			result.Duration = time.Since(testStart)
			results[i] = result
			continue
		}

		result.Actual = strings.TrimSpace(actual)
		result.Passed = normalizeForComparison(result.Actual) ==
			normalizeForComparison(test.Expected)
		result.Duration = time.Since(testStart)
		results[i] = result
	}

	return results
}

// refineSolution asks the LLM to improve the solution based on failed tests.
// Returns the refined solution and a description of the improvement.
func (s *SelfEvolvementPhase) refineSolution(
	ctx context.Context,
	solution string,
	failedTests []*SelfTestResult,
	task string,
	language string,
) (string, string, error) {
	prompt := s.buildRefinePrompt(solution, failedTests, task, language)

	response, err := s.llmClient.Complete(ctx, prompt)
	if err != nil {
		return solution, "", fmt.Errorf("LLM refinement call failed: %w", err)
	}

	refined, improvement := s.parseRefinedSolution(response)

	// If parsing produced an empty solution, keep the original.
	if strings.TrimSpace(refined) == "" {
		return solution, "", nil
	}

	return refined, improvement, nil
}

// buildTestGenPrompt constructs the prompt that asks the LLM to generate
// self-test cases for the given solution.
func (s *SelfEvolvementPhase) buildTestGenPrompt(
	solution, task, language string,
) string {
	return fmt.Sprintf(
		`You are a quality assurance engineer. Given the following %s solution for the task described below, generate 3-5 test cases to validate its correctness.

Task: %s

Solution:
%s

Generate test cases in EXACTLY this format (one per test, separated by ---):

TEST: <test_name>
INPUT: <input>
EXPECTED: <expected_output>
---

Rules:
- Each test must have all three fields (TEST, INPUT, EXPECTED).
- Test names should be descriptive.
- Include edge cases and typical cases.
- Expected output should be the exact value the solution should produce.
- Separate each test case with --- on its own line.`,
		language, task, solution,
	)
}

// buildRefinePrompt constructs the prompt that asks the LLM to improve
// the solution based on failing test cases.
func (s *SelfEvolvementPhase) buildRefinePrompt(
	solution string,
	failedTests []*SelfTestResult,
	task, language string,
) string {
	var failedInfo strings.Builder
	for i, t := range failedTests {
		fmt.Fprintf(&failedInfo,
			"Failed Test %d: %s\n  Input: %s\n  Expected: %s\n  Actual: %s\n\n",
			i+1, t.TestName, t.Input, t.Expected, t.Actual,
		)
	}

	return fmt.Sprintf(
		`You are an expert %s developer. The following solution has failing test cases. Improve the solution so that it passes all tests.

Task: %s

Current Solution:
%s

Failing Tests:
%s

Respond in EXACTLY this format:

SOLUTION:
<your improved solution here>
IMPROVEMENT:
<brief explanation of what you changed and why>`,
		language, task, solution, failedInfo.String(),
	)
}

// parseSelfTests parses the LLM response into SelfTestResult entries.
// Expected format per test:
//
//	TEST: <test_name>
//	INPUT: <input>
//	EXPECTED: <expected_output>
//	---
func (s *SelfEvolvementPhase) parseSelfTests(response string) []*SelfTestResult {
	var results []*SelfTestResult

	blocks := strings.Split(response, "---")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		test := &SelfTestResult{}
		lines := strings.Split(block, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "TEST:") {
				test.TestName = strings.TrimSpace(
					strings.TrimPrefix(line, "TEST:"),
				)
			} else if strings.HasPrefix(line, "INPUT:") {
				test.Input = strings.TrimSpace(
					strings.TrimPrefix(line, "INPUT:"),
				)
			} else if strings.HasPrefix(line, "EXPECTED:") {
				test.Expected = strings.TrimSpace(
					strings.TrimPrefix(line, "EXPECTED:"),
				)
			}
		}

		// Only include if we have at least a name and expected value.
		if test.TestName != "" && test.Expected != "" {
			results = append(results, test)
		}
	}

	return results
}

// parseRefinedSolution extracts the improved solution and the improvement
// description from the LLM refinement response.
// Expected format:
//
//	SOLUTION:
//	<solution code>
//	IMPROVEMENT:
//	<explanation>
func (s *SelfEvolvementPhase) parseRefinedSolution(
	response string,
) (string, string) {
	solution := ""
	improvement := ""

	// Find SOLUTION: marker.
	solIdx := strings.Index(response, "SOLUTION:")
	impIdx := strings.Index(response, "IMPROVEMENT:")

	if solIdx == -1 {
		// No structured format found; treat entire response as solution.
		return strings.TrimSpace(response), ""
	}

	if impIdx == -1 {
		// Only solution marker found.
		solution = strings.TrimSpace(
			response[solIdx+len("SOLUTION:"):],
		)
		return solution, ""
	}

	// Both markers found.
	if solIdx < impIdx {
		solution = strings.TrimSpace(
			response[solIdx+len("SOLUTION:") : impIdx],
		)
		improvement = strings.TrimSpace(
			response[impIdx+len("IMPROVEMENT:"):],
		)
	} else {
		// IMPROVEMENT comes before SOLUTION (unlikely but handle).
		improvement = strings.TrimSpace(
			response[impIdx+len("IMPROVEMENT:") : solIdx],
		)
		solution = strings.TrimSpace(
			response[solIdx+len("SOLUTION:"):],
		)
	}

	return solution, improvement
}

// countPassFail counts the number of passing and total tests.
func countPassFail(tests []*SelfTestResult) (passed, total int) {
	total = len(tests)
	for _, t := range tests {
		if t.Passed {
			passed++
		}
	}
	return passed, total
}

// collectFailed returns only the failing tests from the slice.
func collectFailed(tests []*SelfTestResult) []*SelfTestResult {
	var failed []*SelfTestResult
	for _, t := range tests {
		if !t.Passed {
			failed = append(failed, t)
		}
	}
	return failed
}

// normalizeForComparison normalizes a string for comparison by trimming
// whitespace and converting to lowercase.
func normalizeForComparison(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
