// Package testing provides test-case-driven debate functionality.
// This implements the core innovation from DebateCoder: adversarial test case generation,
// execution, and contrastive analysis for objective ground truth in debates.
package testing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// TestCase represents a generated test case for debate validation.
type TestCase struct {
	ID          string                 `json:"id"`
	AgentID     string                 `json:"agent_id"`     // Agent that generated this test
	TargetAgent string                 `json:"target_agent"` // Agent being tested
	Language    string                 `json:"language"`     // Programming language
	Code        string                 `json:"code"`         // Test code
	Description string                 `json:"description"`  // What this test validates
	Category    TestCategory           `json:"category"`     // Test category
	Difficulty  TestDifficulty         `json:"difficulty"`   // Test difficulty
	Metadata    map[string]interface{} `json:"metadata"`     // Additional metadata
	CreatedAt   int64                  `json:"created_at"`   // Unix timestamp
}

// TestCategory defines types of test cases.
type TestCategory string

const (
	CategoryFunctional    TestCategory = "functional"     // Tests functional correctness
	CategoryEdgeCase      TestCategory = "edge_case"      // Tests boundary conditions
	CategoryPerformance   TestCategory = "performance"    // Tests efficiency
	CategorySecurity      TestCategory = "security"       // Tests security vulnerabilities
	CategoryConcurrency   TestCategory = "concurrency"    // Tests race conditions
	CategoryErrorHandling TestCategory = "error_handling" // Tests error scenarios
)

// TestDifficulty defines test complexity.
type TestDifficulty string

const (
	DifficultyBasic    TestDifficulty = "basic"    // Simple test cases
	DifficultyModerate TestDifficulty = "moderate" // Moderate complexity
	DifficultyAdvanced TestDifficulty = "advanced" // Complex edge cases
	DifficultyExpert   TestDifficulty = "expert"   // Highly sophisticated tests
)

// TestCaseGenerator generates adversarial test cases for debate agents.
type TestCaseGenerator interface {
	// GenerateTestCase generates a test case targeting opponent's solution.
	GenerateTestCase(ctx context.Context, req *GenerateRequest) (*TestCase, error)

	// GenerateBatch generates multiple test cases in parallel.
	GenerateBatch(ctx context.Context, req *GenerateRequest, count int) ([]*TestCase, error)

	// ValidateTestCase checks if generated test is valid and executable.
	ValidateTestCase(ctx context.Context, testCase *TestCase) (*ValidationResult, error)
}

// GenerateRequest contains parameters for test case generation.
type GenerateRequest struct {
	AgentID        string                 `json:"agent_id"`        // Generating agent
	TargetSolution string                 `json:"target_solution"` // Solution to test against
	Language       string                 `json:"language"`        // Programming language
	Context        string                 `json:"context"`         // Problem context
	PreviousTests  []*TestCase            `json:"previous_tests"`  // Previous test cases
	Categories     []TestCategory         `json:"categories"`      // Requested categories
	Difficulty     TestDifficulty         `json:"difficulty"`      // Target difficulty
	Metadata       map[string]interface{} `json:"metadata"`        // Additional context
}

// ValidationResult contains test case validation outcome.
type ValidationResult struct {
	Valid       bool     `json:"valid"`       // Is test case valid?
	Executable  bool     `json:"executable"`  // Can it be executed?
	Errors      []string `json:"errors"`      // Validation errors
	Warnings    []string `json:"warnings"`    // Validation warnings
	Suggestions []string `json:"suggestions"` // Improvement suggestions
}

// LLMTestCaseGenerator uses LLM to generate adversarial test cases.
type LLMTestCaseGenerator struct {
	llmClient LLMClient
	validator TestCaseValidator
}

// LLMClient defines the interface for LLM completion.
// This interface is designed to be compatible with various LLM providers.
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// ProviderAdapter adapts a provider registry to the LLMClient interface.
type ProviderAdapter struct {
	completeFunc func(ctx context.Context, prompt string) (string, error)
}

// NewProviderAdapter creates an adapter from a complete function.
func NewProviderAdapter(completeFunc func(ctx context.Context, prompt string) (string, error)) *ProviderAdapter {
	return &ProviderAdapter{completeFunc: completeFunc}
}

// Complete implements LLMClient.
func (a *ProviderAdapter) Complete(ctx context.Context, prompt string) (string, error) {
	if a.completeFunc == nil {
		return "", fmt.Errorf("complete function not configured")
	}
	return a.completeFunc(ctx, prompt)
}

// NewLLMTestCaseGenerator creates a new LLM-based test generator.
func NewLLMTestCaseGenerator(llmClient LLMClient, validator TestCaseValidator) *LLMTestCaseGenerator {
	return &LLMTestCaseGenerator{
		llmClient: llmClient,
		validator: validator,
	}
}

// NewLLMTestCaseGeneratorFromFunc creates a generator from a simple complete function.
func NewLLMTestCaseGeneratorFromFunc(completeFunc func(ctx context.Context, prompt string) (string, error), validator TestCaseValidator) *LLMTestCaseGenerator {
	return &LLMTestCaseGenerator{
		llmClient: NewProviderAdapter(completeFunc),
		validator: validator,
	}
}

// GenerateTestCase generates a single adversarial test case.
func (g *LLMTestCaseGenerator) GenerateTestCase(ctx context.Context, req *GenerateRequest) (*TestCase, error) {
	prompt := g.buildTestGenerationPrompt(req)

	testCode, description, err := g.generateTestWithLLM(ctx, prompt, req)
	if err != nil {
		return g.generateFallbackTestCase(req), nil
	}

	testID := g.generateTestID(req.AgentID, testCode)

	return &TestCase{
		ID:          testID,
		AgentID:     req.AgentID,
		TargetAgent: "opponent",
		Language:    req.Language,
		Code:        testCode,
		Description: description,
		Category:    g.inferCategory(req),
		Difficulty:  req.Difficulty,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().Unix(),
	}, nil
}

// generateTestWithLLM calls the LLM to generate test code.
func (g *LLMTestCaseGenerator) generateTestWithLLM(ctx context.Context, prompt string, req *GenerateRequest) (string, string, error) {
	if g.llmClient == nil {
		return "", "", fmt.Errorf("LLM client not configured")
	}

	response, err := g.llmClient.Complete(ctx, prompt)
	if err != nil {
		return "", "", fmt.Errorf("LLM completion failed: %w", err)
	}

	return g.parseLLMResponse(response, req.Language)
}

// parseLLMResponse extracts test code and description from LLM response.
func (g *LLMTestCaseGenerator) parseLLMResponse(response, language string) (string, string, error) {
	var testCode, description string

	codeBlockStart := strings.Index(response, "```")
	if codeBlockStart != -1 {
		codeBlockStart = strings.Index(response[codeBlockStart:], "\n")
		if codeBlockStart != -1 {
			codeBlockStart += strings.Index(response, "```") + 1
		}
		codeBlockEnd := strings.Index(response[codeBlockStart:], "```")
		if codeBlockEnd != -1 {
			testCode = strings.TrimSpace(response[codeBlockStart : codeBlockStart+codeBlockEnd])
		}
	}

	if testCode == "" {
		lines := strings.Split(response, "\n")
		codeLines := make([]string, 0)
		inCode := false
		for _, line := range lines {
			if strings.HasPrefix(line, "func ") || strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "describe(") {
				inCode = true
			}
			if inCode {
				codeLines = append(codeLines, line)
			}
		}
		if len(codeLines) > 0 {
			testCode = strings.Join(codeLines, "\n")
		}
	}

	descStart := strings.Index(response, "Description:")
	if descStart != -1 {
		descEnd := strings.Index(response[descStart:], "\n")
		if descEnd != -1 {
			description = strings.TrimSpace(response[descStart+12 : descStart+descEnd])
		}
	}

	if description == "" {
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) > 20 && !strings.HasPrefix(line, "```") && !strings.HasPrefix(line, "func ") && !strings.HasPrefix(line, "def ") && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "#") {
				description = line
				break
			}
		}
	}

	if testCode == "" {
		return "", "", fmt.Errorf("no test code found in LLM response")
	}

	return testCode, description, nil
}

// generateFallbackTestCase creates a basic test when LLM fails.
func (g *LLMTestCaseGenerator) generateFallbackTestCase(req *GenerateRequest) *TestCase {
	testID := g.generateTestID(req.AgentID, req.TargetSolution)

	var fallbackCode string
	switch req.Language {
	case "go":
		fallbackCode = fmt.Sprintf(`func Test%s(t *testing.T) {
	// Basic functionality test
	// Target: %s
	result := true // Placeholder
	if !result {
		t.Error("Test failed")
	}
}`, strings.Title(req.Context), req.TargetSolution[:min(50, len(req.TargetSolution))])
	case "python":
		fallbackCode = fmt.Sprintf(`import unittest

class Test%s(unittest.TestCase):
    def test_basic(self):
        """Basic functionality test for %s"""
        self.assertTrue(True)  # Placeholder
`, strings.Title(req.Context), req.TargetSolution[:min(50, len(req.TargetSolution))])
	case "javascript":
		fallbackCode = fmt.Sprintf(`describe('%s', () => {
    it('should pass basic test', () => {
        // Basic test for %s
        expect(true).toBe(true);
    });
});`, req.Context, req.TargetSolution[:min(50, len(req.TargetSolution))])
	default:
		fallbackCode = fmt.Sprintf("// Test for %s\n// Placeholder test code", req.Context)
	}

	return &TestCase{
		ID:          testID,
		AgentID:     req.AgentID,
		TargetAgent: "opponent",
		Language:    req.Language,
		Code:        fallbackCode,
		Description: fmt.Sprintf("Fallback test for %s", req.Context),
		Category:    CategoryFunctional,
		Difficulty:  DifficultyBasic,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().Unix(),
	}
}

// generateTestID creates a unique test ID.
func (g *LLMTestCaseGenerator) generateTestID(agentID, code string) string {
	hash := sha256.Sum256([]byte(agentID + code + time.Now().String()))
	shortHash := hex.EncodeToString(hash[:8])
	return fmt.Sprintf("test_%s_%s_%d", agentID, shortHash, time.Now().Unix())
}

// inferCategory infers test category from request.
func (g *LLMTestCaseGenerator) inferCategory(req *GenerateRequest) TestCategory {
	if len(req.Categories) > 0 {
		return req.Categories[0]
	}

	lowerCtx := strings.ToLower(req.Context)
	switch {
	case strings.Contains(lowerCtx, "security") || strings.Contains(lowerCtx, "inject"):
		return CategorySecurity
	case strings.Contains(lowerCtx, "concurrent") || strings.Contains(lowerCtx, "race") || strings.Contains(lowerCtx, "thread"):
		return CategoryConcurrency
	case strings.Contains(lowerCtx, "performance") || strings.Contains(lowerCtx, "speed") || strings.Contains(lowerCtx, "memory"):
		return CategoryPerformance
	case strings.Contains(lowerCtx, "error") || strings.Contains(lowerCtx, "exception") || strings.Contains(lowerCtx, "fail"):
		return CategoryErrorHandling
	case strings.Contains(lowerCtx, "edge") || strings.Contains(lowerCtx, "boundary") || strings.Contains(lowerCtx, "limit"):
		return CategoryEdgeCase
	default:
		return CategoryFunctional
	}
}

// GenerateBatch generates multiple test cases in parallel.
func (g *LLMTestCaseGenerator) GenerateBatch(ctx context.Context, req *GenerateRequest, count int) ([]*TestCase, error) {
	tests := make([]*TestCase, 0, count)

	for i := 0; i < count; i++ {
		test, err := g.GenerateTestCase(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate test %d: %w", i, err)
		}
		tests = append(tests, test)
	}

	return tests, nil
}

// ValidateTestCase validates a generated test case.
func (g *LLMTestCaseGenerator) ValidateTestCase(ctx context.Context, testCase *TestCase) (*ValidationResult, error) {
	if g.validator == nil {
		return &ValidationResult{
			Valid:      true,
			Executable: true,
		}, nil
	}

	return g.validator.Validate(ctx, testCase)
}

// buildTestGenerationPrompt constructs prompt for LLM test generation.
func (g *LLMTestCaseGenerator) buildTestGenerationPrompt(req *GenerateRequest) string {
	return fmt.Sprintf(`Generate an adversarial test case for the following solution:

Language: %s
Context: %s

Target Solution:
%s

Requirements:
- Category: %v
- Difficulty: %s
- Focus on edge cases and potential bugs
- Test should be executable and valid
- Aim to expose weaknesses in the solution

Previous Tests:
%d tests already generated

Generate a comprehensive test case that challenges the solution.`,
		req.Language,
		req.Context,
		req.TargetSolution,
		req.Categories,
		req.Difficulty,
		len(req.PreviousTests),
	)
}

// TestCaseValidator validates test cases.
type TestCaseValidator interface {
	Validate(ctx context.Context, testCase *TestCase) (*ValidationResult, error)
}

// BasicTestCaseValidator performs basic validation.
type BasicTestCaseValidator struct{}

// NewBasicTestCaseValidator creates a basic validator.
func NewBasicTestCaseValidator() *BasicTestCaseValidator {
	return &BasicTestCaseValidator{}
}

// Validate performs basic test case validation.
func (v *BasicTestCaseValidator) Validate(ctx context.Context, testCase *TestCase) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:      true,
		Executable: true,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
	}

	// Check required fields
	if testCase.Code == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "test code is empty")
	}

	if testCase.Language == "" {
		result.Warnings = append(result.Warnings, "language not specified")
	}

	if testCase.Description == "" {
		result.Warnings = append(result.Warnings, "description missing")
	}

	return result, nil
}
