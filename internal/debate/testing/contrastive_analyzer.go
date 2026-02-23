// Package testing provides contrastive analysis of test execution results.
package testing

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// ContrastiveAnalysis contains the comparison of test results across solutions.
type ContrastiveAnalysis struct {
	TestID          string                          `json:"test_id"`
	SolutionResults map[string]*TestExecutionResult `json:"solution_results"`
	Winner          string                          `json:"winner"`           // Solution ID with best results
	Differences     []*ResultDifference             `json:"differences"`      // Key differences found
	RootCauses      []*RootCause                    `json:"root_causes"`      // Identified root causes
	Recommendations []*Recommendation               `json:"recommendations"`  // Improvement suggestions
	ConfidenceScore float64                         `json:"confidence_score"` // Analysis confidence (0-1)
	AnalysisMethod  AnalysisMethod                  `json:"analysis_method"`  // Method used
	Timestamp       int64                           `json:"timestamp"`
}

// ResultDifference describes a difference between solution results.
type ResultDifference struct {
	Aspect       string  `json:"aspect"`       // What differs (correctness, performance, etc.)
	Solution1    string  `json:"solution_1"`   // First solution ID
	Solution2    string  `json:"solution_2"`   // Second solution ID
	Value1       string  `json:"value_1"`      // First value
	Value2       string  `json:"value_2"`      // Second value
	Significance float64 `json:"significance"` // How significant (0-1)
	Explanation  string  `json:"explanation"`  // Human-readable explanation
}

// RootCause identifies why a solution failed or performed poorly.
type RootCause struct {
	Type        RootCauseType `json:"type"`
	SolutionID  string        `json:"solution_id"`
	Description string        `json:"description"`
	Evidence    []string      `json:"evidence"`   // Supporting evidence
	Confidence  float64       `json:"confidence"` // Confidence in this root cause (0-1)
	Severity    Severity      `json:"severity"`   // How severe
}

// RootCauseType categorizes root causes.
type RootCauseType string

const (
	RootCauseLogicError        RootCauseType = "logic_error"
	RootCauseBoundaryCondition RootCauseType = "boundary_condition"
	RootCauseRaceCondition     RootCauseType = "race_condition"
	RootCauseMemoryLeak        RootCauseType = "memory_leak"
	RootCausePerformance       RootCauseType = "performance"
	RootCauseSecurity          RootCauseType = "security"
	RootCauseAPI               RootCauseType = "api_misuse"
	RootCauseTypo              RootCauseType = "typo"
)

// Severity levels for root causes.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Recommendation suggests improvements.
type Recommendation struct {
	SolutionID  string `json:"solution_id"`
	Type        string `json:"type"`     // "fix", "optimize", "refactor"
	Priority    int    `json:"priority"` // 1 (highest) - 5 (lowest)
	Description string `json:"description"`
	CodeChange  string `json:"code_change"` // Suggested code change
	Rationale   string `json:"rationale"`   // Why this recommendation
}

// AnalysisMethod defines how contrastive analysis is performed.
type AnalysisMethod string

const (
	AnalysisMethodDifferential AnalysisMethod = "differential" // Compare outputs directly
	AnalysisMethodStatistical  AnalysisMethod = "statistical"  // Statistical comparison
	AnalysisMethodLLMBased     AnalysisMethod = "llm_based"    // Use LLM for analysis
	AnalysisMethodHybrid       AnalysisMethod = "hybrid"       // Combine multiple methods
)

// ContrastiveAnalyzer performs contrastive analysis of test results.
type ContrastiveAnalyzer interface {
	// Analyze performs contrastive analysis of test results.
	Analyze(ctx context.Context, testCase *TestCase, results map[string]*TestExecutionResult) (*ContrastiveAnalysis, error)

	// CompareTwo compares two specific solutions.
	CompareTwo(ctx context.Context, testCase *TestCase, result1, result2 *TestExecutionResult) (*ContrastiveAnalysis, error)

	// IdentifyRootCauses identifies root causes of failures.
	IdentifyRootCauses(ctx context.Context, testCase *TestCase, result *TestExecutionResult, solution *Solution) ([]*RootCause, error)
}

// DifferentialContrastiveAnalyzer uses differential analysis.
type DifferentialContrastiveAnalyzer struct {
	llmClient interface{} // Optional LLM for deep analysis
}

// NewDifferentialContrastiveAnalyzer creates a differential analyzer.
func NewDifferentialContrastiveAnalyzer(llmClient interface{}) *DifferentialContrastiveAnalyzer {
	return &DifferentialContrastiveAnalyzer{
		llmClient: llmClient,
	}
}

// Analyze performs differential contrastive analysis.
func (a *DifferentialContrastiveAnalyzer) Analyze(ctx context.Context, testCase *TestCase, results map[string]*TestExecutionResult) (*ContrastiveAnalysis, error) {
	analysis := &ContrastiveAnalysis{
		TestID:          testCase.ID,
		SolutionResults: results,
		Differences:     make([]*ResultDifference, 0),
		RootCauses:      make([]*RootCause, 0),
		Recommendations: make([]*Recommendation, 0),
		AnalysisMethod:  AnalysisMethodDifferential,
		Timestamp:       time.Now().Unix(),
	}

	var bestSolution string
	var bestDuration float64 = math.MaxFloat64

	for solutionID, result := range results {
		if result.Passed {
			duration := result.Duration.Seconds()
			if duration < bestDuration {
				bestDuration = duration
				bestSolution = solutionID
			}
		}
	}

	analysis.Winner = bestSolution

	solutionIDs := make([]string, 0, len(results))
	for id := range results {
		solutionIDs = append(solutionIDs, id)
	}

	for i := 0; i < len(solutionIDs); i++ {
		for j := i + 1; j < len(solutionIDs); j++ {
			id1 := solutionIDs[i]
			id2 := solutionIDs[j]
			result1 := results[id1]
			result2 := results[id2]

			diffs := a.comparePair(id1, result1, id2, result2)
			analysis.Differences = append(analysis.Differences, diffs...)
		}
	}

	for solutionID, result := range results {
		if !result.Passed && result.Error != "" {
			causes := a.analyzeErrorDeep(testCase, result, &Solution{ID: solutionID})
			analysis.RootCauses = append(analysis.RootCauses, causes...)
		}
	}

	analysis.ConfidenceScore = a.calculateConfidence(analysis)

	return analysis, nil
}

// CompareTwo compares two specific solutions.
func (a *DifferentialContrastiveAnalyzer) CompareTwo(ctx context.Context, testCase *TestCase, result1, result2 *TestExecutionResult) (*ContrastiveAnalysis, error) {
	results := map[string]*TestExecutionResult{
		result1.SolutionID: result1,
		result2.SolutionID: result2,
	}

	return a.Analyze(ctx, testCase, results)
}

// IdentifyRootCauses identifies root causes of failures.
func (a *DifferentialContrastiveAnalyzer) IdentifyRootCauses(ctx context.Context, testCase *TestCase, result *TestExecutionResult, solution *Solution) ([]*RootCause, error) {
	rootCauses := make([]*RootCause, 0)

	// Analyze based on test category and result
	if !result.Passed {
		// Failed test - identify why
		if result.Error != "" {
			cause := a.analyzeError(testCase, result, solution)
			if cause != nil {
				rootCauses = append(rootCauses, cause)
			}
		}

		// Check for common failure patterns
		patterns := a.detectFailurePatterns(testCase, result, solution)
		rootCauses = append(rootCauses, patterns...)
	}

	// Check performance issues even if test passed
	if result.Passed {
		perfIssues := a.detectPerformanceIssues(result)
		rootCauses = append(rootCauses, perfIssues...)
	}

	return rootCauses, nil
}

// comparePair compares two solution results.
func (a *DifferentialContrastiveAnalyzer) comparePair(id1 string, result1 *TestExecutionResult, id2 string, result2 *TestExecutionResult) []*ResultDifference {
	diffs := make([]*ResultDifference, 0)

	// Compare correctness
	if result1.Passed != result2.Passed {
		diffs = append(diffs, &ResultDifference{
			Aspect:       "correctness",
			Solution1:    id1,
			Solution2:    id2,
			Value1:       fmt.Sprintf("%t", result1.Passed),
			Value2:       fmt.Sprintf("%t", result2.Passed),
			Significance: 1.0, // Correctness is most significant
			Explanation:  "One solution passed while the other failed",
		})
	}

	// Compare performance
	if result1.Duration != result2.Duration {
		ratio := float64(result1.Duration) / float64(result2.Duration)
		significance := math.Abs(1.0 - ratio)
		if significance > 0.1 { // >10% difference
			diffs = append(diffs, &ResultDifference{
				Aspect:       "performance",
				Solution1:    id1,
				Solution2:    id2,
				Value1:       result1.Duration.String(),
				Value2:       result2.Duration.String(),
				Significance: math.Min(significance, 1.0),
				Explanation:  fmt.Sprintf("Performance differs by %.1f%%", significance*100),
			})
		}
	}

	// Compare memory usage
	if result1.Metrics != nil && result2.Metrics != nil {
		if result1.Metrics.MemoryUsed != result2.Metrics.MemoryUsed {
			ratio := float64(result1.Metrics.MemoryUsed) / float64(result2.Metrics.MemoryUsed)
			significance := math.Abs(1.0 - ratio)
			if significance > 0.2 { // >20% difference
				diffs = append(diffs, &ResultDifference{
					Aspect:       "memory",
					Solution1:    id1,
					Solution2:    id2,
					Value1:       fmt.Sprintf("%d bytes", result1.Metrics.MemoryUsed),
					Value2:       fmt.Sprintf("%d bytes", result2.Metrics.MemoryUsed),
					Significance: math.Min(significance, 1.0),
					Explanation:  "Significant memory usage difference",
				})
			}
		}
	}

	return diffs
}

// analyzeError analyzes an error to identify root cause.
func (a *DifferentialContrastiveAnalyzer) analyzeError(testCase *TestCase, result *TestExecutionResult, solution *Solution) *RootCause {
	return a.analyzeErrorPatterns(result.Error, solution.ID)
}

// analyzeErrorDeep performs sophisticated error analysis.
func (a *DifferentialContrastiveAnalyzer) analyzeErrorDeep(testCase *TestCase, result *TestExecutionResult, solution *Solution) []*RootCause {
	causes := make([]*RootCause, 0)

	errorMsg := strings.ToLower(result.Error)

	nilPointerPatterns := []string{"nil pointer", "null pointer", "nil reference", "npe"}
	for _, pattern := range nilPointerPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseLogicError,
				SolutionID:  solution.ID,
				Description: "Null pointer dereference detected",
				Evidence:    []string{result.Error},
				Confidence:  0.95,
				Severity:    SeverityCritical,
			})
		}
	}

	indexPatterns := []string{"index out of range", "array index", "slice bounds", "out of bounds"}
	for _, pattern := range indexPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseBoundaryCondition,
				SolutionID:  solution.ID,
				Description: "Array/slice index out of bounds",
				Evidence:    []string{result.Error},
				Confidence:  0.9,
				Severity:    SeverityHigh,
			})
		}
	}

	typePatterns := []string{"type assertion", "interface conversion", "cannot convert", "type mismatch"}
	for _, pattern := range typePatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseLogicError,
				SolutionID:  solution.ID,
				Description: "Type assertion or conversion error",
				Evidence:    []string{result.Error},
				Confidence:  0.85,
				Severity:    SeverityHigh,
			})
		}
	}

	overflowPatterns := []string{"overflow", "underflow", "integer overflow"}
	for _, pattern := range overflowPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseBoundaryCondition,
				SolutionID:  solution.ID,
				Description: "Integer overflow/underflow detected",
				Evidence:    []string{result.Error},
				Confidence:  0.9,
				Severity:    SeverityHigh,
			})
		}
	}

	securityPatterns := []string{"sql injection", "xss", "injection", "escape", "sanitize"}
	for _, pattern := range securityPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseSecurity,
				SolutionID:  solution.ID,
				Description: "Potential security vulnerability",
				Evidence:    []string{result.Error},
				Confidence:  0.8,
				Severity:    SeverityCritical,
			})
		}
	}

	deadlinePattern := regexp.MustCompile(`deadline exceeded|timeout|timed out`)
	if deadlinePattern.MatchString(errorMsg) {
		causes = append(causes, &RootCause{
			Type:        RootCausePerformance,
			SolutionID:  solution.ID,
			Description: "Execution timeout - possible infinite loop or slow operation",
			Evidence:    []string{result.Error},
			Confidence:  0.85,
			Severity:    SeverityHigh,
		})
	}

	memoryPatterns := []string{"out of memory", "oom", "memory allocation", "cannot allocate"}
	for _, pattern := range memoryPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseMemoryLeak,
				SolutionID:  solution.ID,
				Description: "Memory allocation failure - possible leak",
				Evidence:    []string{result.Error},
				Confidence:  0.9,
				Severity:    SeverityCritical,
			})
		}
	}

	deadlockPatterns := []string{"deadlock", "all goroutines", "locked mutex"}
	for _, pattern := range deadlockPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseRaceCondition,
				SolutionID:  solution.ID,
				Description: "Deadlock detected - concurrent access issue",
				Evidence:    []string{result.Error},
				Confidence:  0.95,
				Severity:    SeverityCritical,
			})
		}
	}

	apiPatterns := []string{"undefined", "not found", "no such", "does not exist", "unknown field", "unknown method"}
	for _, pattern := range apiPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseAPI,
				SolutionID:  solution.ID,
				Description: "API misuse - undefined or missing reference",
				Evidence:    []string{result.Error},
				Confidence:  0.8,
				Severity:    SeverityMedium,
			})
		}
	}

	syntaxPatterns := []string{"syntax error", "unexpected token", "parse error", "invalid syntax"}
	for _, pattern := range syntaxPatterns {
		if strings.Contains(errorMsg, pattern) {
			causes = append(causes, &RootCause{
				Type:        RootCauseTypo,
				SolutionID:  solution.ID,
				Description: "Syntax error in solution code",
				Evidence:    []string{result.Error},
				Confidence:  0.95,
				Severity:    SeverityMedium,
			})
		}
	}

	if len(causes) == 0 && result.Error != "" {
		causes = append(causes, &RootCause{
			Type:        RootCauseLogicError,
			SolutionID:  solution.ID,
			Description: "Unidentified error - manual review recommended",
			Evidence:    []string{result.Error},
			Confidence:  0.5,
			Severity:    SeverityMedium,
		})
	}

	return causes
}

// analyzeErrorPatterns analyzes error string for known patterns.
func (a *DifferentialContrastiveAnalyzer) analyzeErrorPatterns(errorMsg, solutionID string) *RootCause {
	errorMsg = strings.ToLower(errorMsg)

	switch {
	case strings.Contains(errorMsg, "nil") || strings.Contains(errorMsg, "null"):
		return &RootCause{
			Type:        RootCauseLogicError,
			SolutionID:  solutionID,
			Description: "Null reference encountered",
			Evidence:    []string{errorMsg},
			Confidence:  0.85,
			Severity:    SeverityHigh,
		}
	case strings.Contains(errorMsg, "index") || strings.Contains(errorMsg, "bound"):
		return &RootCause{
			Type:        RootCauseBoundaryCondition,
			SolutionID:  solutionID,
			Description: "Array boundary violation",
			Evidence:    []string{errorMsg},
			Confidence:  0.85,
			Severity:    SeverityHigh,
		}
	case strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "deadline"):
		return &RootCause{
			Type:        RootCausePerformance,
			SolutionID:  solutionID,
			Description: "Operation timed out",
			Evidence:    []string{errorMsg},
			Confidence:  0.9,
			Severity:    SeverityMedium,
		}
	default:
		return &RootCause{
			Type:        RootCauseLogicError,
			SolutionID:  solutionID,
			Description: "Test execution failed with error",
			Evidence:    []string{errorMsg},
			Confidence:  0.7,
			Severity:    SeverityHigh,
		}
	}
}

// detectFailurePatterns detects common failure patterns.
func (a *DifferentialContrastiveAnalyzer) detectFailurePatterns(testCase *TestCase, result *TestExecutionResult, solution *Solution) []*RootCause {
	causes := make([]*RootCause, 0)

	// Check for timeout
	if result.Error != "" && result.Duration >= 30*1000 { // 30s timeout placeholder
		causes = append(causes, &RootCause{
			Type:        RootCausePerformance,
			SolutionID:  solution.ID,
			Description: "Test execution timed out",
			Evidence:    []string{fmt.Sprintf("Duration: %v", result.Duration)},
			Confidence:  0.9,
			Severity:    SeverityHigh,
		})
	}

	return causes
}

// detectPerformanceIssues detects performance problems.
func (a *DifferentialContrastiveAnalyzer) detectPerformanceIssues(result *TestExecutionResult) []*RootCause {
	causes := make([]*RootCause, 0)

	if result.Metrics != nil {
		// Check memory usage
		if result.Metrics.MemoryUsed > 100*1024*1024 { // >100MB
			causes = append(causes, &RootCause{
				Type:        RootCausePerformance,
				SolutionID:  result.SolutionID,
				Description: "High memory usage detected",
				Evidence:    []string{fmt.Sprintf("Memory: %d MB", result.Metrics.MemoryUsed/(1024*1024))},
				Confidence:  0.7,
				Severity:    SeverityMedium,
			})
		}
	}

	return causes
}

// calculateConfidence calculates overall confidence in analysis.
func (a *DifferentialContrastiveAnalyzer) calculateConfidence(analysis *ContrastiveAnalysis) float64 {
	// Simple confidence based on number of data points
	baseConfidence := 0.5

	// More differences = higher confidence
	if len(analysis.Differences) > 0 {
		baseConfidence += 0.2
	}

	// Root causes identified = higher confidence
	if len(analysis.RootCauses) > 0 {
		baseConfidence += 0.2
	}

	// Clear winner = higher confidence
	if analysis.Winner != "" {
		baseConfidence += 0.1
	}

	return math.Min(baseConfidence, 1.0)
}
