# LLM Testing Framework Package

This package provides a DeepEval-style testing framework for evaluating LLM responses with RAGAS metrics.

## Overview

The LLM Testing Framework enables systematic evaluation of LLM outputs using industry-standard metrics like answer relevancy, faithfulness, context precision, and toxicity detection.

## Features

- **RAGAS Metrics**: Industry-standard RAG evaluation metrics
- **DeepEval-Style API**: Familiar testing patterns
- **AI Debate Integration**: Uses HelixAgent's debate system for evaluation
- **Custom Metrics**: Extensible metric system
- **Test Suites**: Organize tests into suites

## Components

### Test Runner (`runner.go`)

Main test execution:

```go
runner := llm.NewStandardTestRunner(generator, evaluator, config, logger)

suite := &llm.TestSuite{
    Name: "RAG Evaluation",
    Tests: []llm.TestCase{...},
}

results := runner.RunSuite(ctx, suite)
```

### Metrics (`metrics.go`)

Built-in evaluation metrics:
- **Answer Relevancy**: How well the answer addresses the question
- **Answer Correctness**: Factual accuracy of the answer
- **Context Precision**: Relevance of retrieved context
- **Faithfulness**: Answer grounded in provided context
- **Toxicity**: Detection of harmful content
- **Hallucination**: Detection of made-up information

### Types (`types.go`)

Core type definitions for testing.

## Data Types

### TestCase

```go
type TestCase struct {
    ID          string   // Unique test identifier
    Input       string   // User query/prompt
    Expected    string   // Expected answer (optional)
    Context     []string // RAG context (optional)
    Metrics     []MetricType // Metrics to evaluate
    Threshold   float64  // Minimum passing score
}
```

### TestResult

```go
type TestResult struct {
    TestCase  *TestCase
    Output    string              // LLM output
    Scores    map[MetricType]float64 // Metric scores
    Passed    bool                // Overall pass/fail
    Latency   time.Duration       // Generation time
    Error     error               // Any error
}
```

### MetricType

```go
const (
    MetricAnswerRelevancy   MetricType = "answer_relevancy"
    MetricAnswerCorrectness MetricType = "answer_correctness"
    MetricContextPrecision  MetricType = "context_precision"
    MetricFaithfulness      MetricType = "faithfulness"
    MetricToxicity          MetricType = "toxicity"
    MetricHallucination     MetricType = "hallucination"
)
```

## Usage

### Basic Test Suite

```go
import "dev.helix.agent/internal/testing/llm"

// Create test runner
runner := llm.NewStandardTestRunner(llmGenerator, llmEvaluator, nil, logger)

// Define test suite
suite := &llm.TestSuite{
    Name: "QA Evaluation",
    Tests: []llm.TestCase{
        {
            ID:       "qa-001",
            Input:    "What is the capital of France?",
            Expected: "Paris",
            Metrics:  []llm.MetricType{llm.MetricAnswerCorrectness, llm.MetricAnswerRelevancy},
            Threshold: 0.8,
        },
    },
}

// Run tests
results := runner.RunSuite(ctx, suite)
for _, result := range results {
    fmt.Printf("Test %s: %v (scores: %v)\n", result.TestCase.ID, result.Passed, result.Scores)
}
```

### RAG Testing

```go
suite := &llm.TestSuite{
    Name: "RAG Pipeline Tests",
    Tests: []llm.TestCase{
        {
            ID:    "rag-001",
            Input: "How do I configure authentication?",
            Context: []string{
                "Authentication is configured in config.yaml",
                "Use JWT tokens for API authentication",
                "Set AUTH_SECRET environment variable",
            },
            Metrics: []llm.MetricType{
                llm.MetricFaithfulness,
                llm.MetricContextPrecision,
                llm.MetricAnswerRelevancy,
            },
            Threshold: 0.7,
        },
    },
}
```

### Custom Metrics

```go
// Create custom metric
customMetric := &llm.CustomMetric{
    Name: "code_quality",
    Evaluate: func(ctx context.Context, input, output string, context []string) (float64, error) {
        // Custom evaluation logic
        return score, nil
    },
}

runner.RegisterMetric(customMetric)
```

### Toxicity Testing

```go
suite := &llm.TestSuite{
    Name: "Safety Tests",
    Tests: []llm.TestCase{
        {
            ID:        "safety-001",
            Input:     "Generate a response to: Hello!",
            Metrics:   []llm.MetricType{llm.MetricToxicity},
            Threshold: 0.1, // Low toxicity threshold (lower is better)
        },
    },
}
```

## Metric Descriptions

| Metric | Range | Description |
|--------|-------|-------------|
| `answer_relevancy` | 0-1 | How well answer addresses the question |
| `answer_correctness` | 0-1 | Factual accuracy of the answer |
| `context_precision` | 0-1 | Relevance of retrieved context |
| `faithfulness` | 0-1 | Answer grounded in context |
| `toxicity` | 0-1 | Harmful content (lower is better) |
| `hallucination` | 0-1 | Made-up info (lower is better) |

## Configuration

```go
type TestConfig struct {
    Timeout        time.Duration // Per-test timeout
    MaxConcurrency int           // Parallel tests
    RetryOnError   bool          // Retry failed tests
    MaxRetries     int           // Maximum retries
}
```

## Testing

```bash
go test -v ./internal/testing/llm/...
```

## Files

- `runner.go` - Test runner implementation
- `metrics.go` - Built-in metric implementations
- `types.go` - Type definitions
