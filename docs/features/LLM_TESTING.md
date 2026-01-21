# LLM Testing Framework

## Overview

The LLM Testing Framework in HelixAgent provides DeepEval-style testing capabilities for evaluating LLM outputs. It includes RAGAS metrics for RAG evaluation, custom assertion support, and comprehensive test suites for ensuring AI response quality.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    LLM Testing Framework                         │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Test Runner                            │  │
│  │  ├─ Test case execution                                  │  │
│  │  ├─ Parallel evaluation                                  │  │
│  │  └─ Result aggregation                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Metrics Suite                          │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐│  │
│  │  │ RAGAS       │ │ Custom      │ │ Statistical         ││  │
│  │  │ Metrics     │ │ Assertions  │ │ Measures            ││  │
│  │  └─────────────┘ └─────────────┘ └─────────────────────┘│  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Evaluation Models                      │  │
│  │  ├─ Judge LLM (for subjective evaluation)                │  │
│  │  ├─ Embedding model (for similarity)                     │  │
│  │  └─ Custom evaluators                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Test Runner (`internal/testing/llm/runner.go`)

Executes LLM test suites and collects results.

```go
import "dev.helix.agent/internal/testing/llm"

// Create test runner
runner := llm.NewTestRunner(&llm.RunnerConfig{
    JudgeLLM:        judgeLLMProvider,
    EmbeddingModel:  embeddingProvider,
    Parallelism:     4,
    Timeout:         30 * time.Second,
})

// Run test suite
results, err := runner.Run(ctx, testSuite)
```

### 2. Test Cases

Define test cases with inputs, expected outputs, and evaluation criteria.

```go
testSuite := &llm.TestSuite{
    Name: "Code Generation Quality",
    Cases: []*llm.TestCase{
        {
            Name:  "Simple function generation",
            Input: "Write a function to calculate fibonacci numbers",
            ExpectedOutput: &llm.ExpectedOutput{
                Contains:     []string{"def fibonacci", "return"},
                NotContains:  []string{"TODO", "placeholder"},
                MinLength:    50,
                MaxLength:    500,
            },
            Metrics: []llm.Metric{
                llm.MetricFaithfulness,
                llm.MetricAnswerRelevancy,
                llm.MetricCodeCorrectness,
            },
        },
    },
}
```

### 3. RAGAS Metrics (`internal/testing/llm/ragas.go`)

Implementation of RAGAS evaluation metrics for RAG systems.

```go
ragas := llm.NewRAGASEvaluator(&llm.RAGASConfig{
    JudgeLLM:       judgeLLMProvider,
    EmbeddingModel: embeddingProvider,
})

// Evaluate RAG response
result, err := ragas.Evaluate(ctx, &llm.RAGASInput{
    Question:      "What is the capital of France?",
    Answer:        "The capital of France is Paris.",
    Context:       []string{"Paris is the capital and largest city of France."},
    GroundTruth:   "Paris",
})

fmt.Printf("Faithfulness: %.2f\n", result.Faithfulness)
fmt.Printf("Answer Relevancy: %.2f\n", result.AnswerRelevancy)
fmt.Printf("Context Precision: %.2f\n", result.ContextPrecision)
fmt.Printf("Context Recall: %.2f\n", result.ContextRecall)
```

## Available Metrics

### RAGAS Metrics

| Metric | Description | Range |
|--------|-------------|-------|
| `Faithfulness` | Answer grounded in provided context | 0.0 - 1.0 |
| `AnswerRelevancy` | Answer relevant to the question | 0.0 - 1.0 |
| `ContextPrecision` | Retrieved context contains relevant info | 0.0 - 1.0 |
| `ContextRecall` | Retrieved context covers ground truth | 0.0 - 1.0 |

### Quality Metrics

| Metric | Description | Range |
|--------|-------------|-------|
| `Coherence` | Logical flow and structure | 0.0 - 1.0 |
| `Fluency` | Grammar and readability | 0.0 - 1.0 |
| `Completeness` | Covers all aspects of question | 0.0 - 1.0 |
| `Conciseness` | No unnecessary content | 0.0 - 1.0 |

### Code-Specific Metrics

| Metric | Description | Range |
|--------|-------------|-------|
| `CodeCorrectness` | Code executes correctly | 0.0 - 1.0 |
| `CodeStyle` | Follows style guidelines | 0.0 - 1.0 |
| `CodeEfficiency` | Algorithmic efficiency | 0.0 - 1.0 |
| `CodeSecurity` | No security vulnerabilities | 0.0 - 1.0 |

## Custom Assertions

### Built-in Assertions

```go
assertions := []llm.Assertion{
    // Content assertions
    llm.Contains("expected substring"),
    llm.NotContains("forbidden content"),
    llm.MatchesRegex(`\d{4}-\d{2}-\d{2}`), // Date format

    // Length assertions
    llm.MinLength(100),
    llm.MaxLength(1000),
    llm.TokenCount(50, 200),

    // Semantic assertions
    llm.SemanticSimilarity("expected meaning", 0.8),
    llm.TopicMatch([]string{"programming", "software"}, 0.7),

    // Format assertions
    llm.IsValidJSON(),
    llm.IsValidCode("go"),
    llm.HasStructure(expectedSchema),
}
```

### Custom Assertion Functions

```go
// Define custom assertion
customAssertion := llm.CustomAssertion(func(ctx context.Context, output string) (bool, string) {
    // Check for specific requirement
    if !strings.Contains(output, "error handling") {
        return false, "Output should mention error handling"
    }
    return true, ""
})

testCase := &llm.TestCase{
    Name:       "Error handling mention",
    Input:      "Write production-ready code",
    Assertions: []llm.Assertion{customAssertion},
}
```

## Test Suite Configuration

```go
suite := &llm.TestSuite{
    Name:        "API Response Quality",
    Description: "Tests for API endpoint response generation",

    // Global configuration
    Config: &llm.SuiteConfig{
        Parallelism:    4,
        Timeout:        60 * time.Second,
        RetryCount:     2,
        FailFast:       false,

        // Default thresholds
        Thresholds: map[llm.Metric]float64{
            llm.MetricFaithfulness:    0.8,
            llm.MetricAnswerRelevancy: 0.7,
        },
    },

    // Test setup
    Setup: func(ctx context.Context) error {
        // Initialize test resources
        return nil
    },

    // Test teardown
    Teardown: func(ctx context.Context) error {
        // Cleanup test resources
        return nil
    },

    Cases: testCases,
}
```

## Running Tests

### Command Line

```bash
# Run all LLM tests
go test -v ./internal/testing/llm/...

# Run specific test suite
go test -v -run TestRAGASMetrics ./internal/testing/llm/...

# Run with coverage
go test -cover ./internal/testing/llm/...

# Run benchmarks
go test -bench=. ./internal/testing/llm/...
```

### Programmatic Execution

```go
runner := llm.NewTestRunner(config)

// Run and get results
results, err := runner.Run(ctx, suite)
if err != nil {
    log.Fatal(err)
}

// Check results
for _, result := range results.Cases {
    fmt.Printf("Test: %s\n", result.Name)
    fmt.Printf("  Status: %s\n", result.Status)
    for metric, score := range result.Metrics {
        fmt.Printf("  %s: %.2f\n", metric, score)
    }
}

// Overall summary
fmt.Printf("\nPassed: %d/%d\n", results.Passed, results.Total)
fmt.Printf("Average Faithfulness: %.2f\n", results.AverageMetric(llm.MetricFaithfulness))
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: LLM Tests

on: [push, pull_request]

jobs:
  llm-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run LLM Tests
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          go test -v -json ./internal/testing/llm/... > test-results.json

      - name: Upload Results
        uses: actions/upload-artifact@v4
        with:
          name: llm-test-results
          path: test-results.json
```

## Reporting

### JSON Report

```go
report, err := runner.GenerateReport(results, llm.ReportFormatJSON)
os.WriteFile("llm-test-report.json", report, 0644)
```

### HTML Report

```go
report, err := runner.GenerateReport(results, llm.ReportFormatHTML)
os.WriteFile("llm-test-report.html", report, 0644)
```

### Console Report

```go
runner.PrintReport(os.Stdout, results)
```

Output:
```
LLM Test Results: API Response Quality
=====================================

Test Cases: 15/18 passed (83.3%)

Metrics Summary:
  Faithfulness:      0.87 (threshold: 0.80) ✓
  Answer Relevancy:  0.82 (threshold: 0.70) ✓
  Context Precision: 0.75 (threshold: 0.70) ✓
  Context Recall:    0.68 (threshold: 0.70) ✗

Failed Tests:
  - Complex reasoning test: Faithfulness below threshold (0.65)
  - Edge case handling: Timeout exceeded
  - Multi-step instruction: Missing required content
```

## Best Practices

1. **Use representative test cases**: Include diverse inputs that cover expected use cases
2. **Set appropriate thresholds**: Balance between strictness and realistic expectations
3. **Include negative tests**: Test handling of invalid or edge case inputs
4. **Run tests regularly**: Integrate into CI/CD pipeline
5. **Monitor trends**: Track metric scores over time to detect regressions

## Key Files

| File | Description |
|------|-------------|
| `internal/testing/llm/runner.go` | Test runner implementation |
| `internal/testing/llm/ragas.go` | RAGAS metrics implementation |
| `internal/testing/llm/assertions.go` | Built-in assertions |
| `internal/testing/llm/metrics.go` | Metric definitions |
| `internal/testing/llm/report.go` | Report generation |
| `internal/testing/llm/llm_test.go` | Framework tests |

## See Also

- [RAG System](./RAG_SYSTEM.md)
- [Quality Assurance](../guides/qa-testing.md)
- [CI/CD Integration](../guides/cicd.md)
