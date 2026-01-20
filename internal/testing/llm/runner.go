package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StandardTestRunner runs LLM test suites
// Integrates with HelixAgent's AI debate system for evaluation
type StandardTestRunner struct {
	generator  LLMGenerator
	evaluator  LLMEvaluator
	metrics    map[MetricType]Metric
	config     *TestConfig
	logger     *logrus.Logger
}

// LLMGenerator generates responses from an LLM
type LLMGenerator interface {
	// Generate generates a response for the given input
	Generate(ctx context.Context, input string) (string, error)
	// GenerateWithContext generates a response with additional context
	GenerateWithContext(ctx context.Context, input string, context []string) (string, error)
}

// NewStandardTestRunner creates a new test runner
func NewStandardTestRunner(generator LLMGenerator, evaluator LLMEvaluator, config *TestConfig, logger *logrus.Logger) *StandardTestRunner {
	if config == nil {
		config = DefaultTestConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	runner := &StandardTestRunner{
		generator: generator,
		evaluator: evaluator,
		metrics:   make(map[MetricType]Metric),
		config:    config,
		logger:    logger,
	}

	// Register default metrics
	runner.RegisterMetric(NewAnswerRelevancyMetric(evaluator, logger))
	runner.RegisterMetric(NewAnswerCorrectnessMetric(evaluator, logger))
	runner.RegisterMetric(NewContextPrecisionMetric(evaluator, logger))
	runner.RegisterMetric(NewFaithfulnessMetric(evaluator, logger))
	runner.RegisterMetric(NewToxicityMetric(evaluator, logger))
	runner.RegisterMetric(NewHallucinationMetric(evaluator, logger))

	return runner
}

// RegisterMetric registers a metric for use in tests
func (r *StandardTestRunner) RegisterMetric(metric Metric) {
	r.metrics[metric.Type()] = metric
}

// Run runs a test suite and returns a report
func (r *StandardTestRunner) Run(ctx context.Context, suite *TestSuite) (*TestReport, error) {
	startTime := time.Now()

	report := &TestReport{
		Suite:          suite,
		Results:        make([]*TestResult, 0, len(suite.Cases)),
		TotalTests:     len(suite.Cases),
		MetricAverages: make(map[string]float64),
		Timestamp:      startTime,
	}

	// Get metrics to use
	metrics := r.getMetrics(suite.Metrics)
	if len(metrics) == 0 {
		return nil, fmt.Errorf("no valid metrics specified")
	}

	// Run tests with concurrency control
	results := make(chan *TestResult, len(suite.Cases))
	semaphore := make(chan struct{}, r.config.MaxConcurrent)
	var wg sync.WaitGroup

	for _, tc := range suite.Cases {
		wg.Add(1)
		go func(testCase *TestCase) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := r.runSingleCase(ctx, testCase, metrics)
			results <- result
		}(tc)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	metricSums := make(map[string]float64)
	metricCounts := make(map[string]int)

	for result := range results {
		report.Results = append(report.Results, result)

		if result.Passed {
			report.PassedTests++
		} else {
			report.FailedTests++
		}

		// Aggregate metric scores
		for name, score := range result.MetricScores {
			metricSums[name] += score
			metricCounts[name]++
		}
	}

	// Calculate averages
	var totalScore float64
	for name, sum := range metricSums {
		avg := sum / float64(metricCounts[name])
		report.MetricAverages[name] = avg
		totalScore += avg
	}

	if len(metricSums) > 0 {
		report.AverageScore = totalScore / float64(len(metricSums))
	}

	report.Duration = time.Since(startTime)

	r.logger.WithFields(logrus.Fields{
		"suite":    suite.Name,
		"total":    report.TotalTests,
		"passed":   report.PassedTests,
		"failed":   report.FailedTests,
		"avg_score": report.AverageScore,
		"duration": report.Duration,
	}).Info("Test suite completed")

	return report, nil
}

// RunCase runs a single test case
func (r *StandardTestRunner) RunCase(ctx context.Context, testCase *TestCase, metricsList []Metric) (*TestResult, error) {
	return r.runSingleCase(ctx, testCase, metricsList), nil
}

func (r *StandardTestRunner) runSingleCase(ctx context.Context, testCase *TestCase, metrics []Metric) *TestResult {
	startTime := time.Now()

	result := &TestResult{
		TestCase:     testCase,
		MetricScores: make(map[string]float64),
		Timestamp:    startTime,
	}

	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	// Generate response
	var output string
	var err error

	if len(testCase.Context) > 0 {
		output, err = r.generator.GenerateWithContext(testCtx, testCase.Input, testCase.Context)
	} else {
		output, err = r.generator.Generate(testCtx, testCase.Input)
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Generation failed: %v", err))
		result.Duration = time.Since(startTime)
		return result
	}

	result.ActualOutput = output

	// Evaluate with each metric
	metricInput := &MetricInput{
		Input:          testCase.Input,
		Output:         output,
		ExpectedOutput: testCase.ExpectedOutput,
		Context:        testCase.Context,
		GroundTruth:    testCase.GroundTruth,
	}

	var totalScore float64
	for _, metric := range metrics {
		metricOutput, err := metric.Evaluate(testCtx, metricInput)
		if err != nil {
			r.logger.WithError(err).WithField("metric", metric.Name()).Warn("Metric evaluation failed")
			continue
		}

		result.MetricScores[metric.Name()] = metricOutput.Score
		totalScore += metricOutput.Score
	}

	// Calculate overall score
	if len(result.MetricScores) > 0 {
		result.Score = totalScore / float64(len(result.MetricScores))
	}

	// Determine pass/fail
	config := r.config
	if testCase.Metadata != nil {
		if threshold, ok := testCase.Metadata["pass_threshold"].(float64); ok {
			config = &TestConfig{PassThreshold: threshold}
		}
	}
	result.Passed = result.Score >= config.PassThreshold

	result.Duration = time.Since(startTime)

	if r.config.Verbose {
		r.logger.WithFields(logrus.Fields{
			"test":   testCase.Name,
			"passed": result.Passed,
			"score":  result.Score,
		}).Debug("Test case completed")
	}

	return result
}

func (r *StandardTestRunner) getMetrics(metricNames []string) []Metric {
	var metrics []Metric

	for _, name := range metricNames {
		metricType := MetricType(name)
		if metric, exists := r.metrics[metricType]; exists {
			metrics = append(metrics, metric)
		}
	}

	// If no metrics specified, use defaults
	if len(metrics) == 0 {
		for _, m := range r.metrics {
			metrics = append(metrics, m)
		}
	}

	return metrics
}

// DebateLLMEvaluator uses the AI debate system for evaluation
type DebateLLMEvaluator struct {
	debateService DebateEvaluatorService
	logger        *logrus.Logger
}

// DebateEvaluatorService interface for the debate service
type DebateEvaluatorService interface {
	// RunDebate runs a debate on a topic and returns consensus
	RunDebate(ctx context.Context, topic string) (string, float64, error)
}

// NewDebateLLMEvaluator creates a new debate-based evaluator
func NewDebateLLMEvaluator(debateService DebateEvaluatorService, logger *logrus.Logger) *DebateLLMEvaluator {
	return &DebateLLMEvaluator{
		debateService: debateService,
		logger:        logger,
	}
}

// Evaluate evaluates using AI debate
func (e *DebateLLMEvaluator) Evaluate(ctx context.Context, prompt string) (string, error) {
	response, _, err := e.debateService.RunDebate(ctx, prompt)
	return response, err
}

// EvaluateWithScore evaluates and extracts a numeric score
func (e *DebateLLMEvaluator) EvaluateWithScore(ctx context.Context, prompt string) (float64, string, error) {
	response, confidence, err := e.debateService.RunDebate(ctx, prompt)
	if err != nil {
		return 0, "", err
	}

	// Try to extract score from response
	score := extractScoreFromResponse(response)
	if score < 0 {
		// Use confidence as fallback
		score = confidence
	}

	return score, response, nil
}

func extractScoreFromResponse(response string) float64 {
	// Try to parse JSON response
	var result struct {
		Score float64 `json:"score"`
	}

	if err := json.Unmarshal([]byte(response), &result); err == nil {
		return result.Score
	}

	// Try to find score in text
	// Look for patterns like "score: 0.8" or "0.8/1"
	// This is a simplified extraction
	return -1
}

// TestCaseSynthesizer generates test cases
type TestCaseSynthesizer struct {
	generator LLMGenerator
	logger    *logrus.Logger
}

// NewTestCaseSynthesizer creates a new synthesizer
func NewTestCaseSynthesizer(generator LLMGenerator, logger *logrus.Logger) *TestCaseSynthesizer {
	return &TestCaseSynthesizer{
		generator: generator,
		logger:    logger,
	}
}

// GenerateFromDocuments generates test cases from documents
func (s *TestCaseSynthesizer) GenerateFromDocuments(ctx context.Context, documents []string, opts *SynthesizerOptions) ([]*TestCase, error) {
	if opts == nil {
		opts = DefaultSynthesizerOptions()
	}

	var cases []*TestCase

	prompt := fmt.Sprintf(`Generate %d diverse question-answer pairs based on the following documents.

Documents:
%s

For each test case, provide:
1. A question that can be answered from the documents
2. The expected answer
3. Difficulty level (easy/medium/hard)

Format your response as JSON array:
[{"question": "...", "answer": "...", "difficulty": "..."}]`, opts.Count, limitText(mergeDocuments(documents), 4000))

	response, err := s.generator.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate test cases: %w", err)
	}

	// Parse response
	var generated []struct {
		Question   string `json:"question"`
		Answer     string `json:"answer"`
		Difficulty string `json:"difficulty"`
	}

	if err := json.Unmarshal([]byte(extractJSON(response)), &generated); err != nil {
		s.logger.WithError(err).Warn("Failed to parse generated test cases")
		return cases, nil
	}

	for i, g := range generated {
		cases = append(cases, &TestCase{
			ID:             uuid.New().String(),
			Name:           fmt.Sprintf("Generated Test %d", i+1),
			Input:          g.Question,
			ExpectedOutput: g.Answer,
			Context:        documents,
			Tags:           []string{g.Difficulty, "generated"},
		})
	}

	return cases, nil
}

// GenerateFromSchema generates test cases from a schema
func (s *TestCaseSynthesizer) GenerateFromSchema(ctx context.Context, schema interface{}, opts *SynthesizerOptions) ([]*TestCase, error) {
	if opts == nil {
		opts = DefaultSynthesizerOptions()
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`Generate %d test cases for an API/function with the following schema:

%s

Include edge cases and boundary conditions. Format as JSON array:
[{"input": {...}, "expected_output": {...}, "description": "..."}]`, opts.Count, string(schemaJSON))

	response, err := s.generator.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse and convert to TestCase format
	var generated []struct {
		Input          interface{} `json:"input"`
		ExpectedOutput interface{} `json:"expected_output"`
		Description    string      `json:"description"`
	}

	if err := json.Unmarshal([]byte(extractJSON(response)), &generated); err != nil {
		return nil, err
	}

	var cases []*TestCase
	for i, g := range generated {
		inputJSON, _ := json.Marshal(g.Input)
		expectedJSON, _ := json.Marshal(g.ExpectedOutput)

		cases = append(cases, &TestCase{
			ID:             uuid.New().String(),
			Name:           fmt.Sprintf("Schema Test %d", i+1),
			Description:    g.Description,
			Input:          string(inputJSON),
			ExpectedOutput: string(expectedJSON),
			Tags:           []string{"schema-generated"},
		})
	}

	return cases, nil
}

// Augment augments existing test cases with variations
func (s *TestCaseSynthesizer) Augment(ctx context.Context, cases []*TestCase, opts *SynthesizerOptions) ([]*TestCase, error) {
	if opts == nil {
		opts = DefaultSynthesizerOptions()
	}

	var augmented []*TestCase
	augmented = append(augmented, cases...) // Keep originals

	for _, tc := range cases {
		prompt := fmt.Sprintf(`Create 2-3 variations of this test case by:
1. Rephrasing the question
2. Asking for partial information
3. Adding complexity

Original:
Question: %s
Expected Answer: %s

Respond with JSON array:
[{"question": "...", "answer": "..."}]`, tc.Input, tc.ExpectedOutput)

		response, err := s.generator.Generate(ctx, prompt)
		if err != nil {
			continue
		}

		var variations []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		}

		if err := json.Unmarshal([]byte(extractJSON(response)), &variations); err != nil {
			continue
		}

		for i, v := range variations {
			augmented = append(augmented, &TestCase{
				ID:             uuid.New().String(),
				Name:           fmt.Sprintf("%s (Variation %d)", tc.Name, i+1),
				Input:          v.Question,
				ExpectedOutput: v.Answer,
				Context:        tc.Context,
				Tags:           append(tc.Tags, "augmented"),
			})
		}
	}

	return augmented, nil
}

// Helper functions

func mergeDocuments(docs []string) string {
	return fmt.Sprintf("---\n%s\n---", join(docs, "\n---\n"))
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

func limitText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func extractJSON(text string) string {
	// Try to extract JSON array or object from text
	start := -1
	end := -1

	for i, c := range text {
		if c == '[' || c == '{' {
			start = i
			break
		}
	}

	if start == -1 {
		return text
	}

	opener := text[start]
	closer := byte(']')
	if opener == '{' {
		closer = '}'
	}

	depth := 0
	for i := start; i < len(text); i++ {
		if text[i] == byte(opener) {
			depth++
		} else if text[i] == closer {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == -1 {
		return text
	}

	return text[start:end]
}
