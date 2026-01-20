package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockLLMGenerator struct {
	generateFunc            func(ctx context.Context, input string) (string, error)
	generateWithContextFunc func(ctx context.Context, input string, context []string) (string, error)
}

func (m *mockLLMGenerator) Generate(ctx context.Context, input string) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, input)
	}
	return "Mock response for: " + input, nil
}

func (m *mockLLMGenerator) GenerateWithContext(ctx context.Context, input string, context []string) (string, error) {
	if m.generateWithContextFunc != nil {
		return m.generateWithContextFunc(ctx, input, context)
	}
	return "Mock response with context for: " + input, nil
}

type mockLLMEvaluator struct {
	evaluateFunc          func(ctx context.Context, prompt string) (string, error)
	evaluateWithScoreFunc func(ctx context.Context, prompt string) (float64, string, error)
}

func (m *mockLLMEvaluator) Evaluate(ctx context.Context, prompt string) (string, error) {
	if m.evaluateFunc != nil {
		return m.evaluateFunc(ctx, prompt)
	}
	return `{"relevant": true}`, nil
}

func (m *mockLLMEvaluator) EvaluateWithScore(ctx context.Context, prompt string) (float64, string, error) {
	if m.evaluateWithScoreFunc != nil {
		return m.evaluateWithScoreFunc(ctx, prompt)
	}
	return 0.8, "Good response", nil
}

type mockDebateService struct {
	runDebateFunc func(ctx context.Context, topic string) (string, float64, error)
}

func (m *mockDebateService) RunDebate(ctx context.Context, topic string) (string, float64, error) {
	if m.runDebateFunc != nil {
		return m.runDebateFunc(ctx, topic)
	}
	return `{"score": 0.85}`, 0.85, nil
}

// Tests for types.go

func TestTestCase(t *testing.T) {
	tc := &TestCase{
		ID:             "test-1",
		Name:           "Test Case 1",
		Description:    "A test case description",
		Input:          "What is Go?",
		ExpectedOutput: "Go is a programming language.",
		Context:        []string{"Go is a programming language developed by Google."},
		GroundTruth:    "Go is a programming language.",
		Metadata:       map[string]interface{}{"difficulty": "easy"},
		Tags:           []string{"programming", "go"},
	}

	assert.Equal(t, "test-1", tc.ID)
	assert.Equal(t, "Test Case 1", tc.Name)
	assert.Equal(t, "A test case description", tc.Description)
	assert.Equal(t, "What is Go?", tc.Input)
	assert.Equal(t, "Go is a programming language.", tc.ExpectedOutput)
	assert.Len(t, tc.Context, 1)
	assert.Equal(t, "easy", tc.Metadata["difficulty"])
	assert.Contains(t, tc.Tags, "programming")
}

func TestTestResult(t *testing.T) {
	tc := &TestCase{ID: "test-1", Name: "Test 1"}
	now := time.Now()

	result := &TestResult{
		TestCase:     tc,
		ActualOutput: "Response",
		Passed:       true,
		Score:        0.85,
		MetricScores: map[string]float64{"relevancy": 0.9, "correctness": 0.8},
		Errors:       nil,
		Duration:     100 * time.Millisecond,
		Timestamp:    now,
		Metadata:     map[string]interface{}{"retries": 0},
	}

	assert.Equal(t, tc, result.TestCase)
	assert.Equal(t, "Response", result.ActualOutput)
	assert.True(t, result.Passed)
	assert.Equal(t, 0.85, result.Score)
	assert.Equal(t, 0.9, result.MetricScores["relevancy"])
	assert.Equal(t, 100*time.Millisecond, result.Duration)
}

func TestTestSuite(t *testing.T) {
	suite := &TestSuite{
		ID:          "suite-1",
		Name:        "Test Suite 1",
		Description: "A test suite",
		Cases: []*TestCase{
			{ID: "test-1", Name: "Test 1"},
			{ID: "test-2", Name: "Test 2"},
		},
		Metrics: []string{"answer_relevancy", "faithfulness"},
		Config:  DefaultTestConfig(),
	}

	assert.Equal(t, "suite-1", suite.ID)
	assert.Len(t, suite.Cases, 2)
	assert.Len(t, suite.Metrics, 2)
	assert.NotNil(t, suite.Config)
}

func TestDefaultTestConfig(t *testing.T) {
	config := DefaultTestConfig()

	assert.Equal(t, 0.7, config.PassThreshold)
	assert.Equal(t, 5, config.MaxConcurrent)
	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.False(t, config.Verbose)
	assert.Equal(t, 0, config.RetryCount)
}

func TestTestReport(t *testing.T) {
	suite := &TestSuite{ID: "suite-1", Name: "Suite 1"}
	now := time.Now()

	report := &TestReport{
		Suite:          suite,
		Results:        []*TestResult{{Passed: true}, {Passed: false}},
		TotalTests:     2,
		PassedTests:    1,
		FailedTests:    1,
		AverageScore:   0.75,
		MetricAverages: map[string]float64{"relevancy": 0.8},
		Duration:       5 * time.Second,
		Timestamp:      now,
	}

	assert.Equal(t, suite, report.Suite)
	assert.Len(t, report.Results, 2)
	assert.Equal(t, 2, report.TotalTests)
	assert.Equal(t, 1, report.PassedTests)
	assert.Equal(t, 1, report.FailedTests)
	assert.Equal(t, 0.75, report.AverageScore)
}

func TestMetricType(t *testing.T) {
	assert.Equal(t, MetricType("answer_relevancy"), MetricAnswerRelevancy)
	assert.Equal(t, MetricType("answer_correctness"), MetricAnswerCorrectness)
	assert.Equal(t, MetricType("context_precision"), MetricContextPrecision)
	assert.Equal(t, MetricType("faithfulness"), MetricFaithfulness)
	assert.Equal(t, MetricType("toxicity"), MetricToxicity)
	assert.Equal(t, MetricType("hallucination"), MetricHallucination)
	assert.Equal(t, MetricType("custom"), MetricCustom)
}

func TestMetricInput(t *testing.T) {
	input := &MetricInput{
		Input:          "What is Go?",
		Output:         "Go is a programming language.",
		ExpectedOutput: "Go is a programming language developed by Google.",
		Context:        []string{"Context 1", "Context 2"},
		GroundTruth:    "Go is a programming language.",
	}

	assert.Equal(t, "What is Go?", input.Input)
	assert.Equal(t, "Go is a programming language.", input.Output)
	assert.Len(t, input.Context, 2)
}

func TestMetricOutput(t *testing.T) {
	output := &MetricOutput{
		Score:  0.85,
		Reason: "Good match",
		Details: map[string]interface{}{
			"matches": 5,
		},
	}

	assert.Equal(t, 0.85, output.Score)
	assert.Equal(t, "Good match", output.Reason)
	assert.Equal(t, 5, output.Details["matches"])
}

func TestDefaultSynthesizerOptions(t *testing.T) {
	opts := DefaultSynthesizerOptions()

	assert.Equal(t, 10, opts.Count)
	assert.Contains(t, opts.QuestionTypes, "factual")
	assert.Contains(t, opts.QuestionTypes, "reasoning")
	assert.Contains(t, opts.Difficulties, "easy")
	assert.Contains(t, opts.Difficulties, "hard")
	assert.True(t, opts.MultiHop)
	assert.True(t, opts.EdgeCases)
}

func TestBenchmarkConfig(t *testing.T) {
	config := &BenchmarkConfig{
		Iterations:       100,
		WarmupIterations: 10,
		MeasureLatency:   true,
		MeasureTokens:    true,
		MeasureCost:      false,
	}

	assert.Equal(t, 100, config.Iterations)
	assert.Equal(t, 10, config.WarmupIterations)
	assert.True(t, config.MeasureLatency)
	assert.True(t, config.MeasureTokens)
	assert.False(t, config.MeasureCost)
}

func TestBenchmarkResult(t *testing.T) {
	result := &BenchmarkResult{
		Name:            "Benchmark 1",
		Iterations:      100,
		MeanLatency:     50 * time.Millisecond,
		P50Latency:      45 * time.Millisecond,
		P90Latency:      80 * time.Millisecond,
		P99Latency:      120 * time.Millisecond,
		TotalTokens:     10000,
		TokensPerSecond: 500.0,
		TotalCost:       0.05,
		Errors:          2,
		Metadata:        map[string]interface{}{"model": "gpt-4"},
	}

	assert.Equal(t, "Benchmark 1", result.Name)
	assert.Equal(t, 100, result.Iterations)
	assert.Equal(t, 50*time.Millisecond, result.MeanLatency)
	assert.Equal(t, 500.0, result.TokensPerSecond)
	assert.Equal(t, 2, result.Errors)
}

func TestConversationTest(t *testing.T) {
	conv := &ConversationTest{
		ID:   "conv-1",
		Name: "Conversation 1",
		Turns: []*ConversationTurn{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!", Expected: "Hi"},
		},
		Assertions: []*ConversationAssertion{
			{Type: "contains", TurnIndex: 1, Value: "Hi", Message: "Should greet back"},
		},
	}

	assert.Equal(t, "conv-1", conv.ID)
	assert.Len(t, conv.Turns, 2)
	assert.Len(t, conv.Assertions, 1)
	assert.Equal(t, "user", conv.Turns[0].Role)
}

func TestABTestConfig(t *testing.T) {
	config := &ABTestConfig{
		Name: "A/B Test 1",
		Variants: []*ABVariant{
			{Name: "control", Config: map[string]interface{}{"temperature": 0.7}},
			{Name: "variant_a", Config: map[string]interface{}{"temperature": 0.9}},
		},
		Metrics:               []MetricType{MetricAnswerRelevancy, MetricFaithfulness},
		SampleSize:            100,
		SignificanceThreshold: 0.05,
	}

	assert.Equal(t, "A/B Test 1", config.Name)
	assert.Len(t, config.Variants, 2)
	assert.Equal(t, 100, config.SampleSize)
	assert.Equal(t, 0.05, config.SignificanceThreshold)
}

func TestABTestResult(t *testing.T) {
	config := &ABTestConfig{Name: "A/B Test 1"}
	result := &ABTestResult{
		Config: config,
		VariantResults: map[string]*VariantResult{
			"control": {
				Variant:      "control",
				SampleSize:   100,
				MeanScores:   map[string]float64{"relevancy": 0.8},
				StdDevScores: map[string]float64{"relevancy": 0.1},
			},
		},
		Winner:      "control",
		Significant: true,
		PValue:      0.03,
	}

	assert.Equal(t, config, result.Config)
	assert.Equal(t, "control", result.Winner)
	assert.True(t, result.Significant)
	assert.Equal(t, 0.03, result.PValue)
}

// Tests for runner.go

func TestNewStandardTestRunner(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		generator := &mockLLMGenerator{}
		evaluator := &mockLLMEvaluator{}

		runner := NewStandardTestRunner(generator, evaluator, nil, nil)

		assert.NotNil(t, runner)
		assert.NotNil(t, runner.config)
		assert.NotNil(t, runner.logger)
		assert.NotNil(t, runner.metrics)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		generator := &mockLLMGenerator{}
		evaluator := &mockLLMEvaluator{}
		config := &TestConfig{PassThreshold: 0.8, MaxConcurrent: 3}
		logger := logrus.New()

		runner := NewStandardTestRunner(generator, evaluator, config, logger)

		assert.Equal(t, 0.8, runner.config.PassThreshold)
		assert.Equal(t, logger, runner.logger)
	})

	t.Run("RegistersDefaultMetrics", func(t *testing.T) {
		generator := &mockLLMGenerator{}
		evaluator := &mockLLMEvaluator{}

		runner := NewStandardTestRunner(generator, evaluator, nil, nil)

		assert.Contains(t, runner.metrics, MetricAnswerRelevancy)
		assert.Contains(t, runner.metrics, MetricAnswerCorrectness)
		assert.Contains(t, runner.metrics, MetricContextPrecision)
		assert.Contains(t, runner.metrics, MetricFaithfulness)
		assert.Contains(t, runner.metrics, MetricToxicity)
		assert.Contains(t, runner.metrics, MetricHallucination)
	})
}

func TestStandardTestRunner_RegisterMetric(t *testing.T) {
	generator := &mockLLMGenerator{}
	runner := NewStandardTestRunner(generator, nil, nil, nil)

	// Create a custom metric
	customMetric := NewAnswerRelevancyMetric(nil, nil)

	runner.RegisterMetric(customMetric)

	assert.Contains(t, runner.metrics, MetricAnswerRelevancy)
}

func TestStandardTestRunner_Run(t *testing.T) {
	t.Run("SuccessfulRun", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return "Go is a programming language.", nil
			},
		}
		evaluator := &mockLLMEvaluator{
			evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
				return 0.9, "Good", nil
			},
		}

		runner := NewStandardTestRunner(generator, evaluator, &TestConfig{
			PassThreshold: 0.5,
			MaxConcurrent: 2,
			Timeout:       5 * time.Second,
		}, nil)

		suite := &TestSuite{
			ID:   "suite-1",
			Name: "Test Suite",
			Cases: []*TestCase{
				{ID: "test-1", Name: "Test 1", Input: "What is Go?", ExpectedOutput: "Go is a programming language."},
				{ID: "test-2", Name: "Test 2", Input: "What is Python?", ExpectedOutput: "Python is a programming language."},
			},
			Metrics: []string{"answer_relevancy"},
		}

		report, err := runner.Run(context.Background(), suite)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 2, report.TotalTests)
		assert.Greater(t, report.PassedTests, 0)
	})

	t.Run("NoMetrics", func(t *testing.T) {
		generator := &mockLLMGenerator{}
		runner := NewStandardTestRunner(generator, nil, nil, nil)
		runner.metrics = make(map[MetricType]Metric) // Clear metrics

		suite := &TestSuite{
			ID:      "suite-1",
			Cases:   []*TestCase{{ID: "test-1", Input: "Test"}},
			Metrics: []string{"nonexistent_metric"},
		}

		_, err := runner.Run(context.Background(), suite)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no valid metrics")
	})

	t.Run("GenerationError", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return "", errors.New("generation failed")
			},
		}

		runner := NewStandardTestRunner(generator, nil, &TestConfig{
			PassThreshold: 0.5,
			MaxConcurrent: 1,
			Timeout:       time.Second,
		}, nil)

		suite := &TestSuite{
			ID:    "suite-1",
			Cases: []*TestCase{{ID: "test-1", Input: "Test"}},
		}

		report, err := runner.Run(context.Background(), suite)

		require.NoError(t, err)
		assert.NotEmpty(t, report.Results[0].Errors)
	})

	t.Run("WithContext", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateWithContextFunc: func(ctx context.Context, input string, context []string) (string, error) {
				return "Response with context", nil
			},
		}

		runner := NewStandardTestRunner(generator, nil, &TestConfig{
			PassThreshold: 0.5,
			MaxConcurrent: 1,
			Timeout:       time.Second,
		}, nil)

		suite := &TestSuite{
			ID: "suite-1",
			Cases: []*TestCase{
				{ID: "test-1", Input: "Test", Context: []string{"Some context"}},
			},
		}

		report, err := runner.Run(context.Background(), suite)

		require.NoError(t, err)
		assert.Equal(t, "Response with context", report.Results[0].ActualOutput)
	})
}

func TestStandardTestRunner_RunCase(t *testing.T) {
	generator := &mockLLMGenerator{
		generateFunc: func(ctx context.Context, input string) (string, error) {
			return "Test response", nil
		},
	}
	evaluator := &mockLLMEvaluator{
		evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
			return 0.8, "Good", nil
		},
	}

	runner := NewStandardTestRunner(generator, evaluator, &TestConfig{
		PassThreshold: 0.7,
		Timeout:       time.Second,
	}, nil)

	testCase := &TestCase{
		ID:             "test-1",
		Name:           "Test 1",
		Input:          "What is Go?",
		ExpectedOutput: "Go is a programming language.",
	}

	metrics := []Metric{NewAnswerRelevancyMetric(evaluator, nil)}

	result, err := runner.RunCase(context.Background(), testCase, metrics)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testCase, result.TestCase)
	assert.Equal(t, "Test response", result.ActualOutput)
}

func TestStandardTestRunner_CustomThreshold(t *testing.T) {
	generator := &mockLLMGenerator{
		generateFunc: func(ctx context.Context, input string) (string, error) {
			return "Test response", nil
		},
	}

	runner := NewStandardTestRunner(generator, nil, &TestConfig{
		PassThreshold: 0.5,
		MaxConcurrent: 1,
		Timeout:       time.Second,
	}, nil)

	testCase := &TestCase{
		ID:       "test-1",
		Input:    "Test",
		Metadata: map[string]interface{}{"pass_threshold": 0.9},
	}

	suite := &TestSuite{
		ID:    "suite-1",
		Cases: []*TestCase{testCase},
	}

	report, err := runner.Run(context.Background(), suite)

	require.NoError(t, err)
	assert.NotNil(t, report)
}

// Tests for DebateLLMEvaluator

func TestNewDebateLLMEvaluator(t *testing.T) {
	debateService := &mockDebateService{}
	logger := logrus.New()

	evaluator := NewDebateLLMEvaluator(debateService, logger)

	assert.NotNil(t, evaluator)
	assert.Equal(t, debateService, evaluator.debateService)
	assert.Equal(t, logger, evaluator.logger)
}

func TestDebateLLMEvaluator_Evaluate(t *testing.T) {
	debateService := &mockDebateService{
		runDebateFunc: func(ctx context.Context, topic string) (string, float64, error) {
			return "Debate response for: " + topic, 0.85, nil
		},
	}

	evaluator := NewDebateLLMEvaluator(debateService, nil)

	response, err := evaluator.Evaluate(context.Background(), "Test prompt")

	require.NoError(t, err)
	assert.Contains(t, response, "Debate response")
}

func TestDebateLLMEvaluator_EvaluateWithScore(t *testing.T) {
	t.Run("WithScoreInResponse", func(t *testing.T) {
		debateService := &mockDebateService{
			runDebateFunc: func(ctx context.Context, topic string) (string, float64, error) {
				return `{"score": 0.9, "reason": "Good"}`, 0.85, nil
			},
		}

		evaluator := NewDebateLLMEvaluator(debateService, nil)

		score, reason, err := evaluator.EvaluateWithScore(context.Background(), "Test")

		require.NoError(t, err)
		assert.Equal(t, 0.9, score)
		assert.Contains(t, reason, "score")
	})

	t.Run("WithConfidenceFallback", func(t *testing.T) {
		debateService := &mockDebateService{
			runDebateFunc: func(ctx context.Context, topic string) (string, float64, error) {
				return "No JSON here", 0.75, nil
			},
		}

		evaluator := NewDebateLLMEvaluator(debateService, nil)

		score, _, err := evaluator.EvaluateWithScore(context.Background(), "Test")

		require.NoError(t, err)
		assert.Equal(t, 0.75, score) // Uses confidence as fallback
	})

	t.Run("WithError", func(t *testing.T) {
		debateService := &mockDebateService{
			runDebateFunc: func(ctx context.Context, topic string) (string, float64, error) {
				return "", 0, errors.New("debate failed")
			},
		}

		evaluator := NewDebateLLMEvaluator(debateService, nil)

		_, _, err := evaluator.EvaluateWithScore(context.Background(), "Test")

		require.Error(t, err)
	})
}

func TestExtractScoreFromResponse(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		score := extractScoreFromResponse(`{"score": 0.85}`)
		assert.Equal(t, 0.85, score)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		score := extractScoreFromResponse("No score here")
		assert.Equal(t, -1.0, score)
	})
}

// Tests for TestCaseSynthesizer

func TestNewTestCaseSynthesizer(t *testing.T) {
	generator := &mockLLMGenerator{}
	logger := logrus.New()

	synthesizer := NewTestCaseSynthesizer(generator, logger)

	assert.NotNil(t, synthesizer)
	assert.Equal(t, generator, synthesizer.generator)
	assert.Equal(t, logger, synthesizer.logger)
}

func TestTestCaseSynthesizer_GenerateFromDocuments(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return `[{"question": "What is Go?", "answer": "A programming language", "difficulty": "easy"}]`, nil
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		cases, err := synthesizer.GenerateFromDocuments(
			context.Background(),
			[]string{"Go is a programming language developed by Google."},
			nil,
		)

		require.NoError(t, err)
		assert.Len(t, cases, 1)
		assert.Equal(t, "What is Go?", cases[0].Input)
		assert.Contains(t, cases[0].Tags, "easy")
	})

	t.Run("GenerationError", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return "", errors.New("generation failed")
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		_, err := synthesizer.GenerateFromDocuments(context.Background(), []string{"Test"}, nil)

		require.Error(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return "Invalid JSON", nil
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		cases, err := synthesizer.GenerateFromDocuments(context.Background(), []string{"Test"}, nil)

		require.NoError(t, err)
		assert.Empty(t, cases)
	})
}

func TestTestCaseSynthesizer_GenerateFromSchema(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return `[{"input": {"name": "test"}, "expected_output": {"status": "ok"}, "description": "Test case"}]`, nil
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		schema := map[string]interface{}{"type": "object"}
		cases, err := synthesizer.GenerateFromSchema(context.Background(), schema, nil)

		require.NoError(t, err)
		assert.Len(t, cases, 1)
		assert.Contains(t, cases[0].Tags, "schema-generated")
	})

	t.Run("GenerationError", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return "", errors.New("generation failed")
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		_, err := synthesizer.GenerateFromSchema(context.Background(), map[string]interface{}{}, nil)

		require.Error(t, err)
	})
}

func TestTestCaseSynthesizer_Augment(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return `[{"question": "Variation 1", "answer": "Answer 1"}]`, nil
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		original := []*TestCase{
			{ID: "test-1", Name: "Original", Input: "What is Go?", ExpectedOutput: "A language"},
		}

		augmented, err := synthesizer.Augment(context.Background(), original, nil)

		require.NoError(t, err)
		assert.Greater(t, len(augmented), len(original))
		assert.Contains(t, augmented[1].Tags, "augmented")
	})

	t.Run("GenerationError", func(t *testing.T) {
		generator := &mockLLMGenerator{
			generateFunc: func(ctx context.Context, input string) (string, error) {
				return "", errors.New("generation failed")
			},
		}

		synthesizer := NewTestCaseSynthesizer(generator, logrus.New())

		original := []*TestCase{
			{ID: "test-1", Input: "Test", ExpectedOutput: "Answer"},
		}

		augmented, err := synthesizer.Augment(context.Background(), original, nil)

		require.NoError(t, err)
		assert.Len(t, augmented, 1) // Only original, no augmented
	})
}

// Tests for helper functions

func TestMergeDocuments(t *testing.T) {
	docs := []string{"Doc 1", "Doc 2", "Doc 3"}
	result := mergeDocuments(docs)

	assert.Contains(t, result, "Doc 1")
	assert.Contains(t, result, "Doc 2")
	assert.Contains(t, result, "Doc 3")
	assert.Contains(t, result, "---")
}

func TestJoin(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := join([]string{}, ", ")
		assert.Equal(t, "", result)
	})

	t.Run("Single", func(t *testing.T) {
		result := join([]string{"one"}, ", ")
		assert.Equal(t, "one", result)
	})

	t.Run("Multiple", func(t *testing.T) {
		result := join([]string{"one", "two", "three"}, ", ")
		assert.Equal(t, "one, two, three", result)
	})
}

func TestLimitText(t *testing.T) {
	t.Run("WithinLimit", func(t *testing.T) {
		result := limitText("short text", 100)
		assert.Equal(t, "short text", result)
	})

	t.Run("ExceedsLimit", func(t *testing.T) {
		result := limitText("this is a long text that should be truncated", 10)
		assert.Equal(t, "this is a ...", result)
	})
}

func TestExtractJSON(t *testing.T) {
	t.Run("ArrayJSON", func(t *testing.T) {
		text := "Some text before [1, 2, 3] some text after"
		result := extractJSON(text)
		assert.Equal(t, "[1, 2, 3]", result)
	})

	t.Run("ObjectJSON", func(t *testing.T) {
		text := `Some text {"key": "value"} more text`
		result := extractJSON(text)
		assert.Equal(t, `{"key": "value"}`, result)
	})

	t.Run("NestedJSON", func(t *testing.T) {
		text := `Result: {"outer": {"inner": "value"}}`
		result := extractJSON(text)
		assert.Equal(t, `{"outer": {"inner": "value"}}`, result)
	})

	t.Run("NoJSON", func(t *testing.T) {
		text := "No JSON here"
		result := extractJSON(text)
		assert.Equal(t, "No JSON here", result)
	})

	t.Run("UnclosedJSON", func(t *testing.T) {
		text := "Start [1, 2, 3"
		result := extractJSON(text)
		assert.Equal(t, "Start [1, 2, 3", result)
	})
}

// Tests for metrics.go

func TestBaseMetric(t *testing.T) {
	metric := NewAnswerRelevancyMetric(nil, nil)

	assert.Equal(t, "answer_relevancy", metric.Name())
	assert.Equal(t, MetricAnswerRelevancy, metric.Type())
	assert.NotEmpty(t, metric.Description())
}

func TestAnswerRelevancyMetric(t *testing.T) {
	t.Run("HeuristicEvaluation", func(t *testing.T) {
		metric := NewAnswerRelevancyMetric(nil, nil)

		input := &MetricInput{
			Input:  "What is Go programming language?",
			Output: "Go is a programming language developed by Google.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Greater(t, output.Score, 0.0)
		assert.Contains(t, output.Reason, "Word overlap")
	})

	t.Run("EmptyOutput", func(t *testing.T) {
		metric := NewAnswerRelevancyMetric(nil, nil)

		input := &MetricInput{
			Input:  "What is Go?",
			Output: "",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.0, output.Score)
		assert.Contains(t, output.Reason, "Empty output")
	})

	t.Run("WithEvaluator", func(t *testing.T) {
		evaluator := &mockLLMEvaluator{
			evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
				return 0.9, "Highly relevant", nil
			},
		}

		metric := NewAnswerRelevancyMetric(evaluator, nil)

		input := &MetricInput{
			Input:  "What is Go?",
			Output: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.9, output.Score)
	})

	t.Run("EvaluatorError", func(t *testing.T) {
		evaluator := &mockLLMEvaluator{
			evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
				return 0, "", errors.New("evaluation failed")
			},
		}

		metric := NewAnswerRelevancyMetric(evaluator, nil)

		input := &MetricInput{
			Input:  "What is Go?",
			Output: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err) // Falls back to heuristic
		assert.Greater(t, output.Score, 0.0)
	})
}

func TestContextPrecisionMetric(t *testing.T) {
	t.Run("NoContext", func(t *testing.T) {
		metric := NewContextPrecisionMetric(nil, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "Go is a programming language.",
			Context: []string{},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.0, output.Score)
		assert.Contains(t, output.Reason, "No context")
	})

	t.Run("HeuristicEvaluation", func(t *testing.T) {
		metric := NewContextPrecisionMetric(nil, nil)

		input := &MetricInput{
			Input:   "What about programming languages",
			Output:  "Programming is writing code.",
			Context: []string{"Programming languages are used to write code.", "Unrelated context about nothing"},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Greater(t, output.Score, 0.0)
	})

	t.Run("WithEvaluator", func(t *testing.T) {
		evaluator := &mockLLMEvaluator{
			evaluateFunc: func(ctx context.Context, prompt string) (string, error) {
				return `{"relevant": true}`, nil
			},
		}

		metric := NewContextPrecisionMetric(evaluator, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "Go is a programming language.",
			Context: []string{"Go is developed by Google."},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 1.0, output.Score)
	})
}

func TestFaithfulnessMetric(t *testing.T) {
	t.Run("NoContext", func(t *testing.T) {
		metric := NewFaithfulnessMetric(nil, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "Go is a programming language.",
			Context: []string{},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.5, output.Score)
		assert.Contains(t, output.Reason, "No context")
	})

	t.Run("HeuristicEvaluation", func(t *testing.T) {
		metric := NewFaithfulnessMetric(nil, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "Go is a statically typed programming language.",
			Context: []string{"Go is a statically typed programming language developed by Google."},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Greater(t, output.Score, 0.0)
	})

	t.Run("EmptyOutput", func(t *testing.T) {
		metric := NewFaithfulnessMetric(nil, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "",
			Context: []string{"Some context"},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.0, output.Score)
	})

	t.Run("WithEvaluator", func(t *testing.T) {
		evaluator := &mockLLMEvaluator{
			evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
				return 0.95, "Very faithful", nil
			},
		}

		metric := NewFaithfulnessMetric(evaluator, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "Go is a programming language.",
			Context: []string{"Go is a programming language."},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.95, output.Score)
	})
}

func TestToxicityMetric(t *testing.T) {
	t.Run("CleanOutput", func(t *testing.T) {
		metric := NewToxicityMetric(nil, nil)

		input := &MetricInput{
			Output: "Go is a great programming language for building applications.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 1.0, output.Score) // No toxicity = 1.0
	})

	t.Run("ToxicOutput", func(t *testing.T) {
		metric := NewToxicityMetric(nil, nil)

		input := &MetricInput{
			Output: "That stupid idiot should kill the process.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Less(t, output.Score, 1.0) // Has toxicity
		assert.Greater(t, output.Details["matches"].(int), 0)
	})
}

func TestHallucinationMetric(t *testing.T) {
	t.Run("NoGroundTruth", func(t *testing.T) {
		metric := NewHallucinationMetric(nil, nil)

		input := &MetricInput{
			Input:  "What is Go?",
			Output: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.5, output.Score)
		assert.Contains(t, output.Reason, "No ground truth")
	})

	t.Run("WithGroundTruth", func(t *testing.T) {
		metric := NewHallucinationMetric(nil, nil)

		input := &MetricInput{
			Input:       "What is Go?",
			Output:      "Go is a statically typed programming language.",
			GroundTruth: "Go is a statically typed programming language developed by Google.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Greater(t, output.Score, 0.0)
	})

	t.Run("WithContext", func(t *testing.T) {
		metric := NewHallucinationMetric(nil, nil)

		input := &MetricInput{
			Input:   "What is Go?",
			Output:  "Go is a programming language.",
			Context: []string{"Go is a programming language developed by Google."},
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Greater(t, output.Score, 0.0)
	})

	t.Run("EmptyOutput", func(t *testing.T) {
		metric := NewHallucinationMetric(nil, nil)

		input := &MetricInput{
			Input:       "What is Go?",
			Output:      "",
			GroundTruth: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 1.0, output.Score) // Empty output = no hallucination
	})

	t.Run("NoSignificantWords", func(t *testing.T) {
		metric := NewHallucinationMetric(nil, nil)

		input := &MetricInput{
			Input:       "Is it?",
			Output:      "Yes it is.",
			GroundTruth: "Yes it is indeed.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.7, output.Score) // No significant words
	})

	t.Run("WithEvaluator", func(t *testing.T) {
		evaluator := &mockLLMEvaluator{
			evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
				return 0.9, "No hallucinations detected", nil
			},
		}

		metric := NewHallucinationMetric(evaluator, nil)

		input := &MetricInput{
			Input:       "What is Go?",
			Output:      "Go is a programming language.",
			GroundTruth: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.9, output.Score)
	})
}

func TestAnswerCorrectnessMetric(t *testing.T) {
	t.Run("NoExpectedOutput", func(t *testing.T) {
		metric := NewAnswerCorrectnessMetric(nil, nil)

		input := &MetricInput{
			Input:  "What is Go?",
			Output: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.5, output.Score)
		assert.Contains(t, output.Reason, "No expected output")
	})

	t.Run("ExactMatch", func(t *testing.T) {
		metric := NewAnswerCorrectnessMetric(nil, nil)

		input := &MetricInput{
			Input:          "What is Go?",
			Output:         "Go is a programming language.",
			ExpectedOutput: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 1.0, output.Score)
		assert.Contains(t, output.Reason, "Exact match")
	})

	t.Run("PartialMatch", func(t *testing.T) {
		metric := NewAnswerCorrectnessMetric(nil, nil)

		input := &MetricInput{
			Input:          "What is Go?",
			Output:         "Go is a programming language.",
			ExpectedOutput: "Go is a statically typed compiled programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Greater(t, output.Score, 0.0)
		assert.Less(t, output.Score, 1.0)
	})

	t.Run("UsesGroundTruth", func(t *testing.T) {
		metric := NewAnswerCorrectnessMetric(nil, nil)

		input := &MetricInput{
			Input:       "What is Go?",
			Output:      "Go is a programming language.",
			GroundTruth: "Go is a programming language.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 1.0, output.Score)
	})

	t.Run("EmptyExpected", func(t *testing.T) {
		metric := NewAnswerCorrectnessMetric(nil, nil)

		input := &MetricInput{
			Input:          "What is Go?",
			Output:         "Go is a programming language.",
			ExpectedOutput: "",
			GroundTruth:    "",
		}

		// When heuristic is used with empty expected
		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.5, output.Score)
	})

	t.Run("WithEvaluator", func(t *testing.T) {
		evaluator := &mockLLMEvaluator{
			evaluateWithScoreFunc: func(ctx context.Context, prompt string) (float64, string, error) {
				return 0.95, "Very correct", nil
			},
		}

		metric := NewAnswerCorrectnessMetric(evaluator, nil)

		input := &MetricInput{
			Input:          "What is Go?",
			Output:         "Go is a programming language.",
			ExpectedOutput: "Go is a programming language developed by Google.",
		}

		output, err := metric.Evaluate(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, 0.95, output.Score)
	})
}

// Tests for helper functions in metrics.go

func TestMax(t *testing.T) {
	assert.Equal(t, 5, max(5, 3))
	assert.Equal(t, 5, max(3, 5))
	assert.Equal(t, 5, max(5, 5))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 3.0, min(5.0, 3.0))
	assert.Equal(t, 3.0, min(3.0, 5.0))
	assert.Equal(t, 5.0, min(5.0, 5.0))
}

// Concurrent test runner test
func TestStandardTestRunner_Concurrent(t *testing.T) {
	generator := &mockLLMGenerator{
		generateFunc: func(ctx context.Context, input string) (string, error) {
			time.Sleep(10 * time.Millisecond) // Simulate latency
			return "Response", nil
		},
	}

	runner := NewStandardTestRunner(generator, nil, &TestConfig{
		PassThreshold: 0.5,
		MaxConcurrent: 5,
		Timeout:       5 * time.Second,
	}, nil)

	cases := make([]*TestCase, 10)
	for i := 0; i < 10; i++ {
		cases[i] = &TestCase{ID: string(rune('A' + i)), Name: "Test", Input: "Test"}
	}

	suite := &TestSuite{
		ID:    "suite-1",
		Cases: cases,
	}

	report, err := runner.Run(context.Background(), suite)

	require.NoError(t, err)
	assert.Equal(t, 10, report.TotalTests)
}
