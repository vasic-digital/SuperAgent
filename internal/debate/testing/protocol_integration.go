// Package testing integrates test-case-driven debate into the protocol.
package testing

import (
	"context"
	"fmt"
	"sync"
)

// DebateTestIntegration integrates test-driven debate into protocol phases.
type DebateTestIntegration struct {
	generator TestCaseGenerator
	executor  TestExecutor
	analyzer  ContrastiveAnalyzer
	mu        sync.RWMutex
	testCache map[string][]*TestCase // solution_id -> tests
}

// NewDebateTestIntegration creates a test-driven debate integration.
func NewDebateTestIntegration(
	generator TestCaseGenerator,
	executor TestExecutor,
	analyzer ContrastiveAnalyzer,
) *DebateTestIntegration {
	return &DebateTestIntegration{
		generator: generator,
		executor:  executor,
		analyzer:  analyzer,
		testCache: make(map[string][]*TestCase),
	}
}

// TestDrivenDebateRound executes a full test-driven debate round.
// This implements Phase 3 from the documentation: Test-Case-Driven Debate Rounds.
func (d *DebateTestIntegration) TestDrivenDebateRound(
	ctx context.Context,
	solutions []*Solution,
	roundNum int,
) (*RoundResult, error) {
	result := &RoundResult{
		Round:     roundNum,
		Solutions: solutions,
		TestCases: make([]*TestCase, 0),
		Results:   make(map[string]*TestExecutionResult),
		Analyses:  make([]*ContrastiveAnalysis, 0),
	}

	// Step 1: Each agent generates adversarial test cases for opponent solutions
	testCases, err := d.generateAdversarialTests(ctx, solutions)
	if err != nil {
		return nil, fmt.Errorf("test generation failed: %w", err)
	}
	result.TestCases = testCases

	// Step 2: Validate test cases (ensure they're executable)
	validTests, err := d.validateTests(ctx, testCases)
	if err != nil {
		return nil, fmt.Errorf("test validation failed: %w", err)
	}

	// Step 3: Execute validated tests against all solutions
	executionResults, err := d.executeTests(ctx, validTests, solutions)
	if err != nil {
		return nil, fmt.Errorf("test execution failed: %w", err)
	}
	result.Results = executionResults

	// Step 4: Contrastive analysis of execution results
	analyses, err := d.analyzeResults(ctx, validTests, executionResults, solutions)
	if err != nil {
		return nil, fmt.Errorf("contrastive analysis failed: %w", err)
	}
	result.Analyses = analyses

	// Step 5: Determine winners and refine solutions
	winner, recommendations := d.determineWinner(analyses)
	result.Winner = winner
	result.Recommendations = recommendations

	return result, nil
}

// generateAdversarialTests generates test cases for each solution.
func (d *DebateTestIntegration) generateAdversarialTests(
	ctx context.Context,
	solutions []*Solution,
) ([]*TestCase, error) {
	allTests := make([]*TestCase, 0)

	// Each agent generates tests for opponent solutions
	for i, solution := range solutions {
		for j, target := range solutions {
			if i == j {
				continue // Don't generate tests for own solution
			}

			req := &GenerateRequest{
				AgentID:        solution.AgentID,
				TargetSolution: target.Code,
				Language:       target.Language,
				Context:        "Adversarial test generation for debate",
				Difficulty:     DifficultyModerate,
			}

			testCase, err := d.generator.GenerateTestCase(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to generate test from %s to %s: %w",
					solution.AgentID, target.AgentID, err)
			}

			allTests = append(allTests, testCase)
		}
	}

	return allTests, nil
}

// validateTests validates all generated test cases.
func (d *DebateTestIntegration) validateTests(
	ctx context.Context,
	tests []*TestCase,
) ([]*TestCase, error) {
	validTests := make([]*TestCase, 0)

	for _, test := range tests {
		validation, err := d.generator.ValidateTestCase(ctx, test)
		if err != nil {
			// Log but continue with other tests
			continue
		}

		if validation.Valid && validation.Executable {
			validTests = append(validTests, test)
		}
	}

	if len(validTests) == 0 {
		return nil, fmt.Errorf("no valid tests generated")
	}

	return validTests, nil
}

// executeTests executes all tests against all solutions.
func (d *DebateTestIntegration) executeTests(
	ctx context.Context,
	tests []*TestCase,
	solutions []*Solution,
) (map[string]*TestExecutionResult, error) {
	results := make(map[string]*TestExecutionResult)

	for _, test := range tests {
		for _, solution := range solutions {
			result, err := d.executor.Execute(ctx, test, solution)
			if err != nil {
				// Log error but continue
				continue
			}

			key := fmt.Sprintf("%s:%s", test.ID, solution.ID)
			results[key] = result
		}
	}

	return results, nil
}

// analyzeResults performs contrastive analysis on execution results.
func (d *DebateTestIntegration) analyzeResults(
	ctx context.Context,
	tests []*TestCase,
	results map[string]*TestExecutionResult,
	solutions []*Solution,
) ([]*ContrastiveAnalysis, error) {
	analyses := make([]*ContrastiveAnalysis, 0)

	// Group results by test case
	resultsByTest := make(map[string]map[string]*TestExecutionResult)
	for _, result := range results {
		// Parse key: testID:solutionID
		testID := result.TestID
		if resultsByTest[testID] == nil {
			resultsByTest[testID] = make(map[string]*TestExecutionResult)
		}
		resultsByTest[testID][result.SolutionID] = result
	}

	// Analyze each test case
	for _, test := range tests {
		testResults := resultsByTest[test.ID]
		if len(testResults) == 0 {
			continue
		}

		analysis, err := d.analyzer.Analyze(ctx, test, testResults)
		if err != nil {
			// Log but continue
			continue
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// determineWinner determines the winning solution based on analyses.
func (d *DebateTestIntegration) determineWinner(
	analyses []*ContrastiveAnalysis,
) (string, []*Recommendation) {
	// Count wins for each solution
	wins := make(map[string]int)
	allRecommendations := make([]*Recommendation, 0)

	for _, analysis := range analyses {
		if analysis.Winner != "" {
			wins[analysis.Winner]++
		}
		allRecommendations = append(allRecommendations, analysis.Recommendations...)
	}

	// Find solution with most wins
	var winner string
	var maxWins int
	for solutionID, winCount := range wins {
		if winCount > maxWins {
			maxWins = winCount
			winner = solutionID
		}
	}

	return winner, allRecommendations
}

// RoundResult contains results of a test-driven debate round.
type RoundResult struct {
	Round           int                            `json:"round"`
	Solutions       []*Solution                    `json:"solutions"`
	TestCases       []*TestCase                    `json:"test_cases"`
	Results         map[string]*TestExecutionResult `json:"results"`
	Analyses        []*ContrastiveAnalysis         `json:"analyses"`
	Winner          string                         `json:"winner"`
	Recommendations []*Recommendation              `json:"recommendations"`
}
