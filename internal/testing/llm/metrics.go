package llm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// BaseMetric provides common functionality for metrics
type BaseMetric struct {
	name        string
	metricType  MetricType
	description string
	evaluator   LLMEvaluator
	logger      *logrus.Logger
}

// Name returns the metric name
func (m *BaseMetric) Name() string {
	return m.name
}

// Type returns the metric type
func (m *BaseMetric) Type() MetricType {
	return m.metricType
}

// Description returns the metric description
func (m *BaseMetric) Description() string {
	return m.description
}

// AnswerRelevancyMetric measures how relevant the answer is to the question
type AnswerRelevancyMetric struct {
	BaseMetric
}

// NewAnswerRelevancyMetric creates a new answer relevancy metric
func NewAnswerRelevancyMetric(evaluator LLMEvaluator, logger *logrus.Logger) *AnswerRelevancyMetric {
	return &AnswerRelevancyMetric{
		BaseMetric: BaseMetric{
			name:        "answer_relevancy",
			metricType:  MetricAnswerRelevancy,
			description: "Measures how relevant the generated answer is to the input question",
			evaluator:   evaluator,
			logger:      logger,
		},
	}
}

// Evaluate evaluates answer relevancy
func (m *AnswerRelevancyMetric) Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error) {
	if m.evaluator == nil {
		// Fallback to heuristic evaluation
		return m.heuristicEvaluate(input)
	}

	prompt := fmt.Sprintf(`Evaluate how relevant the following answer is to the question.

Question: %s

Answer: %s

Rate the relevancy on a scale of 0 to 1, where:
- 0: Completely irrelevant
- 0.5: Partially relevant
- 1: Highly relevant

Respond with a JSON object:
{"score": <number>, "reason": "<explanation>"}`, input.Input, input.Output)

	score, reason, err := m.evaluator.EvaluateWithScore(ctx, prompt)
	if err != nil {
		return m.heuristicEvaluate(input)
	}

	return &MetricOutput{
		Score:  score,
		Reason: reason,
	}, nil
}

func (m *AnswerRelevancyMetric) heuristicEvaluate(input *MetricInput) (*MetricOutput, error) {
	// Simple word overlap heuristic
	inputWords := strings.Fields(strings.ToLower(input.Input))
	outputWords := strings.Fields(strings.ToLower(input.Output))

	if len(outputWords) == 0 {
		return &MetricOutput{Score: 0, Reason: "Empty output"}, nil
	}

	matches := 0
	for _, iw := range inputWords {
		for _, ow := range outputWords {
			if iw == ow {
				matches++
				break
			}
		}
	}

	score := float64(matches) / float64(len(inputWords))
	if score > 1 {
		score = 1
	}

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Word overlap: %d/%d", matches, len(inputWords)),
	}, nil
}

// ContextPrecisionMetric measures context precision for RAG
type ContextPrecisionMetric struct {
	BaseMetric
}

// NewContextPrecisionMetric creates a new context precision metric
func NewContextPrecisionMetric(evaluator LLMEvaluator, logger *logrus.Logger) *ContextPrecisionMetric {
	return &ContextPrecisionMetric{
		BaseMetric: BaseMetric{
			name:        "context_precision",
			metricType:  MetricContextPrecision,
			description: "Measures the precision of retrieved context (relevant items / total retrieved)",
			evaluator:   evaluator,
			logger:      logger,
		},
	}
}

// Evaluate evaluates context precision
func (m *ContextPrecisionMetric) Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error) {
	if len(input.Context) == 0 {
		return &MetricOutput{Score: 0, Reason: "No context provided"}, nil
	}

	if m.evaluator == nil {
		return m.heuristicEvaluate(input)
	}

	relevantCount := 0
	for _, ctx := range input.Context {
		prompt := fmt.Sprintf(`Determine if the following context is relevant to answering the question.

Question: %s

Context: %s

Respond with: {"relevant": true/false}`, input.Input, ctx)

		response, err := m.evaluator.Evaluate(ctx, prompt)
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(response), `"relevant": true`) ||
			strings.Contains(strings.ToLower(response), `"relevant":true`) {
			relevantCount++
		}
	}

	score := float64(relevantCount) / float64(len(input.Context))

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Relevant contexts: %d/%d", relevantCount, len(input.Context)),
	}, nil
}

func (m *ContextPrecisionMetric) heuristicEvaluate(input *MetricInput) (*MetricOutput, error) {
	inputLower := strings.ToLower(input.Input)
	relevantCount := 0

	for _, ctx := range input.Context {
		ctxLower := strings.ToLower(ctx)
		// Simple relevance check: any word from input appears in context
		for _, word := range strings.Fields(inputLower) {
			if len(word) > 3 && strings.Contains(ctxLower, word) {
				relevantCount++
				break
			}
		}
	}

	score := float64(relevantCount) / float64(len(input.Context))

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Relevant contexts (heuristic): %d/%d", relevantCount, len(input.Context)),
	}, nil
}

// FaithfulnessMetric measures if the answer is faithful to the context
type FaithfulnessMetric struct {
	BaseMetric
}

// NewFaithfulnessMetric creates a new faithfulness metric
func NewFaithfulnessMetric(evaluator LLMEvaluator, logger *logrus.Logger) *FaithfulnessMetric {
	return &FaithfulnessMetric{
		BaseMetric: BaseMetric{
			name:        "faithfulness",
			metricType:  MetricFaithfulness,
			description: "Measures if the answer is faithful to the provided context (no hallucinations)",
			evaluator:   evaluator,
			logger:      logger,
		},
	}
}

// Evaluate evaluates faithfulness
func (m *FaithfulnessMetric) Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error) {
	if len(input.Context) == 0 {
		return &MetricOutput{Score: 0.5, Reason: "No context to verify faithfulness"}, nil
	}

	if m.evaluator == nil {
		return m.heuristicEvaluate(input)
	}

	contextStr := strings.Join(input.Context, "\n\n")
	prompt := fmt.Sprintf(`Evaluate if the following answer is faithful to the given context (i.e., all claims in the answer are supported by the context).

Context:
%s

Answer: %s

Rate faithfulness on a scale of 0 to 1:
- 0: Contains significant unsupported claims
- 0.5: Some claims may not be fully supported
- 1: All claims are supported by context

Respond with: {"score": <number>, "reason": "<explanation>"}`, contextStr, input.Output)

	score, reason, err := m.evaluator.EvaluateWithScore(ctx, prompt)
	if err != nil {
		return m.heuristicEvaluate(input)
	}

	return &MetricOutput{
		Score:  score,
		Reason: reason,
	}, nil
}

func (m *FaithfulnessMetric) heuristicEvaluate(input *MetricInput) (*MetricOutput, error) {
	// Simple heuristic: check if output words appear in context
	contextStr := strings.ToLower(strings.Join(input.Context, " "))
	outputWords := strings.Fields(strings.ToLower(input.Output))

	if len(outputWords) == 0 {
		return &MetricOutput{Score: 0, Reason: "Empty output"}, nil
	}

	supportedWords := 0
	for _, word := range outputWords {
		if len(word) > 4 && strings.Contains(contextStr, word) {
			supportedWords++
		}
	}

	score := float64(supportedWords) / float64(len(outputWords))

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Supported words (heuristic): %d/%d", supportedWords, len(outputWords)),
	}, nil
}

// ToxicityMetric measures toxicity in output
type ToxicityMetric struct {
	BaseMetric
	toxicPatterns []*regexp.Regexp
}

// NewToxicityMetric creates a new toxicity metric
func NewToxicityMetric(evaluator LLMEvaluator, logger *logrus.Logger) *ToxicityMetric {
	return &ToxicityMetric{
		BaseMetric: BaseMetric{
			name:        "toxicity",
			metricType:  MetricToxicity,
			description: "Measures the level of toxic content in the output",
			evaluator:   evaluator,
			logger:      logger,
		},
		toxicPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(hate|kill|murder|attack|destroy)\b`),
			regexp.MustCompile(`(?i)\b(stupid|idiot|moron|dumb)\b`),
			regexp.MustCompile(`(?i)\b(racist|sexist|homophobic)\b`),
		},
	}
}

// Evaluate evaluates toxicity
func (m *ToxicityMetric) Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error) {
	// Count pattern matches
	totalMatches := 0
	for _, pattern := range m.toxicPatterns {
		matches := pattern.FindAllString(input.Output, -1)
		totalMatches += len(matches)
	}

	// Normalize score (0 = no toxicity, 1 = high toxicity)
	// We invert this so higher score = safer
	toxicityRatio := float64(totalMatches) / float64(max(len(strings.Fields(input.Output)), 1))
	score := 1.0 - min(toxicityRatio*10, 1.0) // Scale and invert

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Toxic pattern matches: %d", totalMatches),
		Details: map[string]interface{}{
			"matches": totalMatches,
		},
	}, nil
}

// HallucinationMetric detects hallucinations
type HallucinationMetric struct {
	BaseMetric
}

// NewHallucinationMetric creates a new hallucination metric
func NewHallucinationMetric(evaluator LLMEvaluator, logger *logrus.Logger) *HallucinationMetric {
	return &HallucinationMetric{
		BaseMetric: BaseMetric{
			name:        "hallucination",
			metricType:  MetricHallucination,
			description: "Detects hallucinated (unsupported) content in the output",
			evaluator:   evaluator,
			logger:      logger,
		},
	}
}

// Evaluate evaluates hallucination (1 = no hallucination, 0 = complete hallucination)
func (m *HallucinationMetric) Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error) {
	if input.GroundTruth == "" && len(input.Context) == 0 {
		return &MetricOutput{Score: 0.5, Reason: "No ground truth or context to verify"}, nil
	}

	if m.evaluator == nil {
		return m.heuristicEvaluate(input)
	}

	reference := input.GroundTruth
	if reference == "" {
		reference = strings.Join(input.Context, "\n\n")
	}

	prompt := fmt.Sprintf(`Analyze the following answer for hallucinations (claims not supported by the reference).

Reference/Ground Truth:
%s

Answer to check:
%s

Rate the hallucination level (0 = severe hallucination, 1 = no hallucination):
Respond with: {"score": <number>, "reason": "<explanation>", "hallucinated_claims": [<list of unsupported claims>]}`, reference, input.Output)

	score, reason, err := m.evaluator.EvaluateWithScore(ctx, prompt)
	if err != nil {
		return m.heuristicEvaluate(input)
	}

	return &MetricOutput{
		Score:  score,
		Reason: reason,
	}, nil
}

func (m *HallucinationMetric) heuristicEvaluate(input *MetricInput) (*MetricOutput, error) {
	reference := strings.ToLower(input.GroundTruth)
	if reference == "" {
		reference = strings.ToLower(strings.Join(input.Context, " "))
	}

	outputWords := strings.Fields(strings.ToLower(input.Output))
	if len(outputWords) == 0 {
		return &MetricOutput{Score: 1, Reason: "Empty output"}, nil
	}

	// Check what percentage of significant words are in reference
	supported := 0
	significant := 0
	for _, word := range outputWords {
		if len(word) > 5 { // Only check significant words
			significant++
			if strings.Contains(reference, word) {
				supported++
			}
		}
	}

	if significant == 0 {
		return &MetricOutput{Score: 0.7, Reason: "No significant words to verify"}, nil
	}

	score := float64(supported) / float64(significant)

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Supported significant words: %d/%d", supported, significant),
	}, nil
}

// AnswerCorrectnessMetric measures correctness against expected output
type AnswerCorrectnessMetric struct {
	BaseMetric
}

// NewAnswerCorrectnessMetric creates a new answer correctness metric
func NewAnswerCorrectnessMetric(evaluator LLMEvaluator, logger *logrus.Logger) *AnswerCorrectnessMetric {
	return &AnswerCorrectnessMetric{
		BaseMetric: BaseMetric{
			name:        "answer_correctness",
			metricType:  MetricAnswerCorrectness,
			description: "Measures the correctness of the answer compared to expected output",
			evaluator:   evaluator,
			logger:      logger,
		},
	}
}

// Evaluate evaluates answer correctness
func (m *AnswerCorrectnessMetric) Evaluate(ctx context.Context, input *MetricInput) (*MetricOutput, error) {
	if input.ExpectedOutput == "" && input.GroundTruth == "" {
		return &MetricOutput{Score: 0.5, Reason: "No expected output to compare"}, nil
	}

	expected := input.ExpectedOutput
	if expected == "" {
		expected = input.GroundTruth
	}

	if m.evaluator == nil {
		return m.heuristicEvaluate(input, expected)
	}

	prompt := fmt.Sprintf(`Compare the actual answer with the expected answer.

Expected Answer: %s

Actual Answer: %s

Rate the correctness (0 = completely wrong, 1 = perfectly correct):
Respond with: {"score": <number>, "reason": "<explanation>"}`, expected, input.Output)

	score, reason, err := m.evaluator.EvaluateWithScore(ctx, prompt)
	if err != nil {
		return m.heuristicEvaluate(input, expected)
	}

	return &MetricOutput{
		Score:  score,
		Reason: reason,
	}, nil
}

func (m *AnswerCorrectnessMetric) heuristicEvaluate(input *MetricInput, expected string) (*MetricOutput, error) {
	// Normalize strings
	actual := strings.ToLower(strings.TrimSpace(input.Output))
	expected = strings.ToLower(strings.TrimSpace(expected))

	// Exact match
	if actual == expected {
		return &MetricOutput{Score: 1.0, Reason: "Exact match"}, nil
	}

	// Calculate word overlap
	actualWords := strings.Fields(actual)
	expectedWords := strings.Fields(expected)

	if len(expectedWords) == 0 {
		return &MetricOutput{Score: 0.5, Reason: "Empty expected output"}, nil
	}

	matches := 0
	for _, ew := range expectedWords {
		for _, aw := range actualWords {
			if ew == aw {
				matches++
				break
			}
		}
	}

	score := float64(matches) / float64(len(expectedWords))

	return &MetricOutput{
		Score:  score,
		Reason: fmt.Sprintf("Word match: %d/%d", matches, len(expectedWords)),
	}, nil
}

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
