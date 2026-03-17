# debate/testing - Test-Driven Debate

Provides test case generation, contrastive analysis, protocol integration testing, and test execution for validating debate outcomes.

## Purpose

The testing package enables test-driven debate by automatically generating test cases for debate proposals, executing them against debate outputs, and performing contrastive analysis to identify differences between competing proposals.

## Key Components

### TestCaseGenerator

Automatically generates test cases from debate proposals and expected outcomes.

```go
generator := testing.NewTestCaseGenerator(llmProvider, logger)
testCases, err := generator.Generate(ctx, proposal, constraints)
```

### ContrastiveAnalyzer

Compares two or more debate proposals to identify key differences, strengths, and weaknesses.

```go
analyzer := testing.NewContrastiveAnalyzer(logger)
analysis, err := analyzer.Analyze(ctx, proposals)
// Returns: differences, strengths per proposal, weaknesses per proposal
```

### TestExecutor

Executes generated test cases against debate outputs and reports results.

```go
executor := testing.NewTestExecutor(codeRunner, logger)
results, err := executor.Execute(ctx, testCases, debateOutput)
```

### ProtocolIntegration

Integrates test execution into the debate protocol, running tests between phases.

```go
integration := testing.NewProtocolIntegration(generator, executor, logger)
integration.ValidatePhaseOutput(ctx, phase, output)
```

## Key Types

- **TestCase** -- Generated test with input, expected output, and validation criteria
- **TestResult** -- Execution result with pass/fail, actual output, and diagnostics
- **ContrastiveReport** -- Detailed comparison of multiple proposals
- **ValidationResult** -- Phase-level validation outcome

## Usage within Debate System

The testing package is invoked during the Review and Optimization phases. The test case generator creates validation criteria from the debate topic, the executor runs these against proposals, and the contrastive analyzer helps the voting system make informed decisions.

## Files

- `test_case_generator.go` -- LLM-based test case generation
- `test_executor.go` -- Test execution engine
- `contrastive_analyzer.go` -- Multi-proposal comparison
- `protocol_integration.go` -- Debate protocol hooks
- `test_driven_debate_test.go` -- Unit tests
