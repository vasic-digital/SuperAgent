// Package llm provides LLM evaluation and testing frameworks
// inspired by DeepEval, RAGAS, and other modern testing approaches.
package llm

import (
	"context"
	"time"
)

// TestCase represents an LLM test case
type TestCase struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Input         string                 `json:"input"`
	ExpectedOutput string                `json:"expected_output,omitempty"`
	Context       []string               `json:"context,omitempty"`
	GroundTruth   string                 `json:"ground_truth,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
}

// TestResult contains the result of running a test case
type TestResult struct {
	TestCase      *TestCase         `json:"test_case"`
	ActualOutput  string            `json:"actual_output"`
	Passed        bool              `json:"passed"`
	Score         float64           `json:"score"`
	MetricScores  map[string]float64 `json:"metric_scores"`
	Errors        []string          `json:"errors,omitempty"`
	Duration      time.Duration     `json:"duration"`
	Timestamp     time.Time         `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TestSuite contains a collection of test cases
type TestSuite struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Cases       []*TestCase `json:"cases"`
	Metrics     []string    `json:"metrics"`
	Config      *TestConfig `json:"config,omitempty"`
}

// TestConfig configures test execution
type TestConfig struct {
	// Minimum passing score (0-1)
	PassThreshold float64 `json:"pass_threshold"`
	// Maximum concurrent tests
	MaxConcurrent int `json:"max_concurrent"`
	// Timeout per test
	Timeout time.Duration `json:"timeout"`
	// Enable verbose logging
	Verbose bool `json:"verbose"`
	// Retry failed tests
	RetryCount int `json:"retry_count"`
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		PassThreshold: 0.7,
		MaxConcurrent: 5,
		Timeout:       60 * time.Second,
		Verbose:       false,
		RetryCount:    0,
	}
}

// TestReport contains the results of running a test suite
type TestReport struct {
	Suite         *TestSuite    `json:"suite"`
	Results       []*TestResult `json:"results"`
	TotalTests    int           `json:"total_tests"`
	PassedTests   int           `json:"passed_tests"`
	FailedTests   int           `json:"failed_tests"`
	AverageScore  float64       `json:"average_score"`
	MetricAverages map[string]float64 `json:"metric_averages"`
	Duration      time.Duration `json:"duration"`
	Timestamp     time.Time     `json:"timestamp"`
}

// MetricType represents the type of evaluation metric
type MetricType string

const (
	// Answer Quality Metrics
	MetricAnswerRelevancy    MetricType = "answer_relevancy"
	MetricAnswerCorrectness  MetricType = "answer_correctness"
	MetricAnswerCompleteness MetricType = "answer_completeness"
	MetricAnswerConsistency  MetricType = "answer_consistency"

	// RAG Metrics
	MetricContextPrecision   MetricType = "context_precision"
	MetricContextRecall      MetricType = "context_recall"
	MetricContextRelevancy   MetricType = "context_relevancy"
	MetricFaithfulness       MetricType = "faithfulness"

	// Safety Metrics
	MetricToxicity           MetricType = "toxicity"
	MetricBias               MetricType = "bias"
	MetricHarmfulness        MetricType = "harmfulness"

	// Quality Metrics
	MetricFluency            MetricType = "fluency"
	MetricCoherence          MetricType = "coherence"
	MetricConciseness        MetricType = "conciseness"

	// Task-Specific Metrics
	MetricSummarization      MetricType = "summarization"
	MetricHallucination      MetricType = "hallucination"
	MetricFactualAccuracy    MetricType = "factual_accuracy"

	// Custom
	MetricCustom             MetricType = "custom"
)

// Metric evaluates LLM output
type Metric interface {
	// Name returns the metric name
	Name() string
	// Type returns the metric type
	Type() MetricType
	// Evaluate evaluates the output and returns a score (0-1)
	Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error)
	// Description returns a description of what the metric measures
	Description() string
}

// MetricInput contains input for metric evaluation
type MetricInput struct {
	Input          string   `json:"input"`
	Output         string   `json:"output"`
	ExpectedOutput string   `json:"expected_output,omitempty"`
	Context        []string `json:"context,omitempty"`
	GroundTruth    string   `json:"ground_truth,omitempty"`
}

// MetricOutput contains the result of metric evaluation
type MetricOutput struct {
	Score       float64                `json:"score"`
	Reason      string                 `json:"reason,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// LLMEvaluator uses LLMs to evaluate responses
type LLMEvaluator interface {
	// Evaluate uses an LLM to evaluate the response
	Evaluate(ctx context.Context, prompt string) (string, error)
	// EvaluateWithScore evaluates and extracts a numeric score
	EvaluateWithScore(ctx context.Context, prompt string) (float64, string, error)
}

// DatasetLoader loads test datasets
type DatasetLoader interface {
	// Load loads a dataset by name
	Load(ctx context.Context, name string) (*TestSuite, error)
	// LoadFromFile loads a dataset from a file
	LoadFromFile(ctx context.Context, path string) (*TestSuite, error)
	// ListAvailable lists available datasets
	ListAvailable() []string
}

// TestRunner runs test suites
type TestRunner interface {
	// Run runs a test suite
	Run(ctx context.Context, suite *TestSuite) (*TestReport, error)
	// RunCase runs a single test case
	RunCase(ctx context.Context, testCase *TestCase, metrics []Metric) (*TestResult, error)
}

// Synthesizer generates test cases
type Synthesizer interface {
	// GenerateFromDocuments generates test cases from documents
	GenerateFromDocuments(ctx context.Context, documents []string, opts *SynthesizerOptions) ([]*TestCase, error)
	// GenerateFromSchema generates test cases from a schema
	GenerateFromSchema(ctx context.Context, schema interface{}, opts *SynthesizerOptions) ([]*TestCase, error)
	// Augment augments existing test cases
	Augment(ctx context.Context, cases []*TestCase, opts *SynthesizerOptions) ([]*TestCase, error)
}

// SynthesizerOptions configures test case synthesis
type SynthesizerOptions struct {
	// Number of test cases to generate
	Count int `json:"count"`
	// Types of questions to generate
	QuestionTypes []string `json:"question_types"`
	// Difficulty levels
	Difficulties []string `json:"difficulties"`
	// Include multi-hop questions
	MultiHop bool `json:"multi_hop"`
	// Include edge cases
	EdgeCases bool `json:"edge_cases"`
}

// DefaultSynthesizerOptions returns default synthesizer options
func DefaultSynthesizerOptions() *SynthesizerOptions {
	return &SynthesizerOptions{
		Count:         10,
		QuestionTypes: []string{"factual", "reasoning", "comparison"},
		Difficulties:  []string{"easy", "medium", "hard"},
		MultiHop:      true,
		EdgeCases:     true,
	}
}

// BenchmarkConfig configures benchmarking
type BenchmarkConfig struct {
	// Number of iterations
	Iterations int `json:"iterations"`
	// Warmup iterations
	WarmupIterations int `json:"warmup_iterations"`
	// Measure latency
	MeasureLatency bool `json:"measure_latency"`
	// Measure tokens
	MeasureTokens bool `json:"measure_tokens"`
	// Measure cost
	MeasureCost bool `json:"measure_cost"`
}

// BenchmarkResult contains benchmark results
type BenchmarkResult struct {
	Name           string                 `json:"name"`
	Iterations     int                    `json:"iterations"`
	MeanLatency    time.Duration          `json:"mean_latency"`
	P50Latency     time.Duration          `json:"p50_latency"`
	P90Latency     time.Duration          `json:"p90_latency"`
	P99Latency     time.Duration          `json:"p99_latency"`
	TotalTokens    int                    `json:"total_tokens"`
	TokensPerSecond float64               `json:"tokens_per_second"`
	TotalCost      float64                `json:"total_cost"`
	Errors         int                    `json:"errors"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationTest represents a multi-turn conversation test
type ConversationTest struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Turns       []*ConversationTurn     `json:"turns"`
	Assertions  []*ConversationAssertion `json:"assertions"`
}

// ConversationTurn represents a single turn in a conversation
type ConversationTurn struct {
	Role      string `json:"role"` // user, assistant
	Content   string `json:"content"`
	Expected  string `json:"expected,omitempty"`
}

// ConversationAssertion asserts something about the conversation
type ConversationAssertion struct {
	Type       string `json:"type"` // contains, not_contains, regex, custom
	TurnIndex  int    `json:"turn_index"`
	Value      string `json:"value"`
	Message    string `json:"message,omitempty"`
}

// A/B Test Types

// ABTestConfig configures A/B testing
type ABTestConfig struct {
	// Name of the A/B test
	Name string `json:"name"`
	// Variants to test
	Variants []*ABVariant `json:"variants"`
	// Metrics to measure
	Metrics []MetricType `json:"metrics"`
	// Sample size per variant
	SampleSize int `json:"sample_size"`
	// Statistical significance threshold
	SignificanceThreshold float64 `json:"significance_threshold"`
}

// ABVariant represents a variant in an A/B test
type ABVariant struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// ABTestResult contains A/B test results
type ABTestResult struct {
	Config       *ABTestConfig              `json:"config"`
	VariantResults map[string]*VariantResult `json:"variant_results"`
	Winner       string                     `json:"winner,omitempty"`
	Significant  bool                       `json:"significant"`
	PValue       float64                    `json:"p_value"`
}

// VariantResult contains results for a single variant
type VariantResult struct {
	Variant      string             `json:"variant"`
	SampleSize   int                `json:"sample_size"`
	MeanScores   map[string]float64 `json:"mean_scores"`
	StdDevScores map[string]float64 `json:"std_dev_scores"`
}
