// Package testing provides tests for test-driven debate functionality.
package testing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMTestCaseGenerator_GenerateTestCase(t *testing.T) {
	validator := NewBasicTestCaseValidator()
	generator := NewLLMTestCaseGenerator(nil, validator)

	req := &GenerateRequest{
		AgentID:        "agent1",
		TargetSolution: "func add(a, b int) int { return a + b }",
		Language:       "go",
		Context:        "Test generation for add function",
		Difficulty:     DifficultyModerate,
	}

	ctx := context.Background()
	testCase, err := generator.GenerateTestCase(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, testCase)
	assert.Equal(t, "agent1", testCase.AgentID)
	assert.Equal(t, "go", testCase.Language)
}

func TestLLMTestCaseGenerator_GenerateBatch(t *testing.T) {
	validator := NewBasicTestCaseValidator()
	generator := NewLLMTestCaseGenerator(nil, validator)

	req := &GenerateRequest{
		AgentID:        "agent1",
		TargetSolution: "func multiply(a, b int) int { return a * b }",
		Language:       "go",
		Context:        "Batch test generation",
		Difficulty:     DifficultyBasic,
	}

	ctx := context.Background()
	tests, err := generator.GenerateBatch(ctx, req, 3)

	require.NoError(t, err)
	assert.Len(t, tests, 3)
}

func TestBasicTestCaseValidator_Validate(t *testing.T) {
	validator := NewBasicTestCaseValidator()

	tests := []struct {
		name     string
		testCase *TestCase
		wantValid bool
	}{
		{
			name: "Valid test case",
			testCase: &TestCase{
				ID:          "test1",
				Code:        "assert(add(1, 2) == 3)",
				Language:    "go",
				Description: "Test add function",
			},
			wantValid: true,
		},
		{
			name: "Empty code",
			testCase: &TestCase{
				ID:          "test2",
				Code:        "",
				Language:    "go",
				Description: "Empty test",
			},
			wantValid: false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Validate(ctx, tt.testCase)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestSandboxedTestExecutor_Execute(t *testing.T) {
	executor := NewSandboxedTestExecutor(
		WithTimeout(5 * time.Second),
		WithMemoryLimit(256 * 1024 * 1024),
	)

	testCase := &TestCase{
		ID:       "test1",
		Code:     "assert(add(1, 2) == 3)",
		Language: "go",
	}

	solution := &Solution{
		ID:       "solution1",
		AgentID:  "agent1",
		Language: "go",
		Code:     "func add(a, b int) int { return a + b }",
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, testCase, solution)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test1", result.TestID)
	assert.Equal(t, "solution1", result.SolutionID)
}

func TestSandboxedTestExecutor_ExecuteBatch(t *testing.T) {
	executor := NewSandboxedTestExecutor()

	tests := []*TestCase{
		{ID: "test1", Code: "assert(add(1, 2) == 3)", Language: "go"},
		{ID: "test2", Code: "assert(add(0, 0) == 0)", Language: "go"},
	}

	solution := &Solution{
		ID:       "solution1",
		Language: "go",
		Code:     "func add(a, b int) int { return a + b }",
	}

	ctx := context.Background()
	results, err := executor.ExecuteBatch(ctx, tests, solution)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestDifferentialContrastiveAnalyzer_Analyze(t *testing.T) {
	analyzer := NewDifferentialContrastiveAnalyzer(nil)

	testCase := &TestCase{
		ID:       "test1",
		Language: "go",
	}

	results := map[string]*TestExecutionResult{
		"solution1": {
			TestID:     "test1",
			SolutionID: "solution1",
			Passed:     true,
			Duration:   100 * time.Millisecond,
		},
		"solution2": {
			TestID:     "test1",
			SolutionID: "solution2",
			Passed:     false,
			Duration:   200 * time.Millisecond,
		},
	}

	ctx := context.Background()
	analysis, err := analyzer.Analyze(ctx, testCase, results)

	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "solution1", analysis.Winner)
	assert.NotEmpty(t, analysis.Differences)
}

func TestDebateTestIntegration_TestDrivenDebateRound(t *testing.T) {
	generator := NewLLMTestCaseGenerator(nil, NewBasicTestCaseValidator())
	executor := NewSandboxedTestExecutor()
	analyzer := NewDifferentialContrastiveAnalyzer(nil)

	integration := NewDebateTestIntegration(generator, executor, analyzer)

	solutions := []*Solution{
		{ID: "sol1", AgentID: "agent1", Language: "go", Code: "func add(a, b int) int { return a + b }"},
		{ID: "sol2", AgentID: "agent2", Language: "go", Code: "func add(a, b int) int { return a - b }"},
	}

	ctx := context.Background()
	result, err := integration.TestDrivenDebateRound(ctx, solutions, 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Round)
	assert.NotEmpty(t, result.Solutions)
}
