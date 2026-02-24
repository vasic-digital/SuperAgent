package selfimprove

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// ---------------------------------------------------------------------------
// FeedbackType constants
// ---------------------------------------------------------------------------

func TestFeedbackType_Values(t *testing.T) {
	tests := []struct {
		name     string
		ft       FeedbackType
		expected string
	}{
		{"Positive", FeedbackTypePositive, "positive"},
		{"Negative", FeedbackTypeNegative, "negative"},
		{"Neutral", FeedbackTypeNeutral, "neutral"},
		{"Suggestion", FeedbackTypeSuggestion, "suggestion"},
		{"Correction", FeedbackTypeCorrection, "correction"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.ft))
		})
	}
}

func TestFeedbackType_Uniqueness(t *testing.T) {
	types := []FeedbackType{
		FeedbackTypePositive,
		FeedbackTypeNegative,
		FeedbackTypeNeutral,
		FeedbackTypeSuggestion,
		FeedbackTypeCorrection,
	}
	seen := make(map[FeedbackType]bool)
	for _, ft := range types {
		assert.False(t, seen[ft], "duplicate FeedbackType: %s", ft)
		seen[ft] = true
	}
	assert.Equal(t, 5, len(seen))
}

// ---------------------------------------------------------------------------
// FeedbackSource constants
// ---------------------------------------------------------------------------

func TestFeedbackSource_Values(t *testing.T) {
	tests := []struct {
		name     string
		fs       FeedbackSource
		expected string
	}{
		{"Human", FeedbackSourceHuman, "human"},
		{"AI", FeedbackSourceAI, "ai"},
		{"Debate", FeedbackSourceDebate, "debate"},
		{"Verifier", FeedbackSourceVerifier, "verifier"},
		{"Metric", FeedbackSourceMetric, "metric"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.fs))
		})
	}
}

func TestFeedbackSource_Uniqueness(t *testing.T) {
	sources := []FeedbackSource{
		FeedbackSourceHuman,
		FeedbackSourceAI,
		FeedbackSourceDebate,
		FeedbackSourceVerifier,
		FeedbackSourceMetric,
	}
	seen := make(map[FeedbackSource]bool)
	for _, fs := range sources {
		assert.False(t, seen[fs], "duplicate FeedbackSource: %s", fs)
		seen[fs] = true
	}
	assert.Equal(t, 5, len(seen))
}

// ---------------------------------------------------------------------------
// DimensionType constants
// ---------------------------------------------------------------------------

func TestDimensionType_Values(t *testing.T) {
	tests := []struct {
		name     string
		dt       DimensionType
		expected string
	}{
		{"Accuracy", DimensionAccuracy, "accuracy"},
		{"Relevance", DimensionRelevance, "relevance"},
		{"Helpfulness", DimensionHelpfulness, "helpfulness"},
		{"Harmless", DimensionHarmless, "harmlessness"},
		{"Honest", DimensionHonest, "honesty"},
		{"Coherence", DimensionCoherence, "coherence"},
		{"Creativity", DimensionCreativity, "creativity"},
		{"Formatting", DimensionFormatting, "formatting"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.dt))
		})
	}
}

func TestDimensionType_Uniqueness(t *testing.T) {
	dims := []DimensionType{
		DimensionAccuracy,
		DimensionRelevance,
		DimensionHelpfulness,
		DimensionHarmless,
		DimensionHonest,
		DimensionCoherence,
		DimensionCreativity,
		DimensionFormatting,
	}
	seen := make(map[DimensionType]bool)
	for _, d := range dims {
		assert.False(t, seen[d], "duplicate DimensionType: %s", d)
		seen[d] = true
	}
	assert.Equal(t, 8, len(seen))
}

// ---------------------------------------------------------------------------
// PolicyUpdateType constants
// ---------------------------------------------------------------------------

func TestPolicyUpdateType_Values(t *testing.T) {
	tests := []struct {
		name     string
		put      PolicyUpdateType
		expected string
	}{
		{"PromptRefinement", PolicyUpdatePromptRefinement, "prompt_refinement"},
		{"GuidelineAddition", PolicyUpdateGuidelineAddition, "guideline_addition"},
		{"ExampleAddition", PolicyUpdateExampleAddition, "example_addition"},
		{"ConstraintUpdate", PolicyUpdateConstraintUpdate, "constraint_update"},
		{"ToneAdjustment", PolicyUpdateToneAdjustment, "tone_adjustment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.put))
		})
	}
}

func TestPolicyUpdateType_Uniqueness(t *testing.T) {
	types := []PolicyUpdateType{
		PolicyUpdatePromptRefinement,
		PolicyUpdateGuidelineAddition,
		PolicyUpdateExampleAddition,
		PolicyUpdateConstraintUpdate,
		PolicyUpdateToneAdjustment,
	}
	seen := make(map[PolicyUpdateType]bool)
	for _, pt := range types {
		assert.False(t, seen[pt], "duplicate PolicyUpdateType: %s", pt)
		seen[pt] = true
	}
	assert.Equal(t, 5, len(seen))
}

// ---------------------------------------------------------------------------
// Feedback struct
// ---------------------------------------------------------------------------

func TestFeedback_ZeroValue(t *testing.T) {
	var f Feedback
	assert.Empty(t, f.ID)
	assert.Empty(t, f.SessionID)
	assert.Empty(t, f.PromptID)
	assert.Empty(t, f.ResponseID)
	assert.Equal(t, FeedbackType(""), f.Type)
	assert.Equal(t, FeedbackSource(""), f.Source)
	assert.Equal(t, float64(0), f.Score)
	assert.Nil(t, f.Dimensions)
	assert.Empty(t, f.Comment)
	assert.Empty(t, f.Correction)
	assert.Empty(t, f.ProviderName)
	assert.Empty(t, f.Model)
	assert.Nil(t, f.Metadata)
	assert.True(t, f.CreatedAt.IsZero())
}

func TestFeedback_FullPopulation(t *testing.T) {
	now := time.Now()
	f := Feedback{
		ID:         "fb-001",
		SessionID:  "sess-001",
		PromptID:   "prompt-001",
		ResponseID: "resp-001",
		Type:       FeedbackTypePositive,
		Source:     FeedbackSourceHuman,
		Score:      0.85,
		Dimensions: map[DimensionType]float64{
			DimensionAccuracy:  0.9,
			DimensionRelevance: 0.8,
		},
		Comment:      "Great response",
		Correction:   "Minor fix here",
		ProviderName: "claude",
		Model:        "claude-3-sonnet",
		Metadata:     map[string]interface{}{"key": "value"},
		CreatedAt:    now,
	}

	assert.Equal(t, "fb-001", f.ID)
	assert.Equal(t, "sess-001", f.SessionID)
	assert.Equal(t, FeedbackTypePositive, f.Type)
	assert.Equal(t, FeedbackSourceHuman, f.Source)
	assert.Equal(t, 0.85, f.Score)
	assert.Len(t, f.Dimensions, 2)
	assert.Equal(t, 0.9, f.Dimensions[DimensionAccuracy])
	assert.Equal(t, "Great response", f.Comment)
	assert.Equal(t, "Minor fix here", f.Correction)
	assert.Equal(t, "claude", f.ProviderName)
	assert.Equal(t, "claude-3-sonnet", f.Model)
	assert.Equal(t, "value", f.Metadata["key"])
	assert.Equal(t, now, f.CreatedAt)
}

func TestFeedback_ScoreBounds(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"MinBound", -1.0},
		{"Zero", 0.0},
		{"MaxBound", 1.0},
		{"Negative", -0.5},
		{"Positive", 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Feedback{Score: tt.score}
			assert.Equal(t, tt.score, f.Score)
		})
	}
}

// ---------------------------------------------------------------------------
// TrainingExample struct
// ---------------------------------------------------------------------------

func TestTrainingExample_ZeroValue(t *testing.T) {
	var te TrainingExample
	assert.Empty(t, te.ID)
	assert.Empty(t, te.Prompt)
	assert.Empty(t, te.Response)
	assert.Empty(t, te.PreferredResponse)
	assert.Empty(t, te.RejectedResponse)
	assert.Nil(t, te.Feedback)
	assert.Equal(t, float64(0), te.RewardScore)
	assert.Nil(t, te.Dimensions)
	assert.Empty(t, te.SystemPrompt)
	assert.Empty(t, te.ProviderName)
	assert.Empty(t, te.Model)
	assert.Nil(t, te.Metadata)
	assert.True(t, te.CreatedAt.IsZero())
}

func TestTrainingExample_FullPopulation(t *testing.T) {
	now := time.Now()
	fb := &Feedback{ID: "fb-1", Score: 0.9}
	te := TrainingExample{
		ID:                "te-001",
		Prompt:            "What is Go?",
		Response:          "Go is a programming language.",
		PreferredResponse: "Go is a statically typed, compiled language.",
		RejectedResponse:  "Go is a game.",
		Feedback:          []*Feedback{fb},
		RewardScore:       0.88,
		Dimensions: map[DimensionType]float64{
			DimensionAccuracy:    0.95,
			DimensionHelpfulness: 0.80,
		},
		SystemPrompt: "Be helpful",
		ProviderName: "openai",
		Model:        "gpt-4",
		Metadata:     map[string]interface{}{"version": 2},
		CreatedAt:    now,
	}

	assert.Equal(t, "te-001", te.ID)
	assert.Equal(t, "What is Go?", te.Prompt)
	assert.Equal(t, 0.88, te.RewardScore)
	assert.Len(t, te.Feedback, 1)
	assert.Equal(t, "fb-1", te.Feedback[0].ID)
	assert.Equal(t, 0.95, te.Dimensions[DimensionAccuracy])
	assert.Equal(t, "openai", te.ProviderName)
	assert.Equal(t, now, te.CreatedAt)
}

func TestTrainingExample_WithNilFeedback(t *testing.T) {
	te := TrainingExample{
		ID:       "te-nil",
		Feedback: nil,
	}
	assert.Nil(t, te.Feedback)
}

func TestTrainingExample_WithEmptyFeedback(t *testing.T) {
	te := TrainingExample{
		ID:       "te-empty",
		Feedback: []*Feedback{},
	}
	assert.NotNil(t, te.Feedback)
	assert.Len(t, te.Feedback, 0)
}

// ---------------------------------------------------------------------------
// PreferencePair struct
// ---------------------------------------------------------------------------

func TestPreferencePair_ZeroValue(t *testing.T) {
	var pp PreferencePair
	assert.Empty(t, pp.ID)
	assert.Empty(t, pp.Prompt)
	assert.Empty(t, pp.Chosen)
	assert.Empty(t, pp.Rejected)
	assert.Equal(t, float64(0), pp.ChosenScore)
	assert.Equal(t, float64(0), pp.RejectedScore)
	assert.Equal(t, float64(0), pp.Margin)
	assert.Equal(t, FeedbackSource(""), pp.Source)
	assert.Nil(t, pp.Metadata)
	assert.True(t, pp.CreatedAt.IsZero())
}

func TestPreferencePair_FullPopulation(t *testing.T) {
	now := time.Now()
	pp := PreferencePair{
		ID:            "pp-001",
		Prompt:        "Explain AI",
		Chosen:        "AI is artificial intelligence.",
		Rejected:      "AI is a movie.",
		ChosenScore:   0.9,
		RejectedScore: 0.3,
		Margin:        0.6,
		Source:        FeedbackSourceDebate,
		Metadata:      map[string]interface{}{"debate_id": "d-1"},
		CreatedAt:     now,
	}

	assert.Equal(t, "pp-001", pp.ID)
	assert.Equal(t, "Explain AI", pp.Prompt)
	assert.Equal(t, "AI is artificial intelligence.", pp.Chosen)
	assert.Equal(t, "AI is a movie.", pp.Rejected)
	assert.Equal(t, 0.9, pp.ChosenScore)
	assert.Equal(t, 0.3, pp.RejectedScore)
	assert.Equal(t, 0.6, pp.Margin)
	assert.Equal(t, FeedbackSourceDebate, pp.Source)
	assert.Equal(t, "d-1", pp.Metadata["debate_id"])
	assert.Equal(t, now, pp.CreatedAt)
}

func TestPreferencePair_ChosenScoreGreaterThanRejected(t *testing.T) {
	pp := PreferencePair{
		ChosenScore:   0.9,
		RejectedScore: 0.3,
		Margin:        0.6,
	}
	assert.Greater(t, pp.ChosenScore, pp.RejectedScore)
	assert.InDelta(t, pp.ChosenScore-pp.RejectedScore, pp.Margin, 0.001)
}

// ---------------------------------------------------------------------------
// FeedbackFilter struct
// ---------------------------------------------------------------------------

func TestFeedbackFilter_ZeroValue(t *testing.T) {
	var ff FeedbackFilter
	assert.Nil(t, ff.SessionIDs)
	assert.Nil(t, ff.PromptIDs)
	assert.Nil(t, ff.Types)
	assert.Nil(t, ff.Sources)
	assert.Nil(t, ff.MinScore)
	assert.Nil(t, ff.MaxScore)
	assert.Nil(t, ff.ProviderNames)
	assert.Nil(t, ff.Models)
	assert.Nil(t, ff.StartTime)
	assert.Nil(t, ff.EndTime)
	assert.Equal(t, 0, ff.Limit)
	assert.Equal(t, 0, ff.Offset)
}

func TestFeedbackFilter_FullPopulation(t *testing.T) {
	minScore := -0.5
	maxScore := 0.9
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	ff := FeedbackFilter{
		SessionIDs:    []string{"sess-1", "sess-2"},
		PromptIDs:     []string{"p-1"},
		Types:         []FeedbackType{FeedbackTypePositive, FeedbackTypeNegative},
		Sources:       []FeedbackSource{FeedbackSourceHuman},
		MinScore:      &minScore,
		MaxScore:      &maxScore,
		ProviderNames: []string{"claude"},
		Models:        []string{"claude-3-sonnet"},
		StartTime:     &start,
		EndTime:       &end,
		Limit:         50,
		Offset:        10,
	}

	require.NotNil(t, ff.MinScore)
	assert.Equal(t, -0.5, *ff.MinScore)
	require.NotNil(t, ff.MaxScore)
	assert.Equal(t, 0.9, *ff.MaxScore)
	assert.Len(t, ff.SessionIDs, 2)
	assert.Len(t, ff.Types, 2)
	assert.Equal(t, 50, ff.Limit)
	assert.Equal(t, 10, ff.Offset)
}

func TestFeedbackFilter_NilOptionalPointers(t *testing.T) {
	ff := FeedbackFilter{
		Limit: 100,
	}
	assert.Nil(t, ff.MinScore)
	assert.Nil(t, ff.MaxScore)
	assert.Nil(t, ff.StartTime)
	assert.Nil(t, ff.EndTime)
}

// ---------------------------------------------------------------------------
// AggregatedFeedback struct
// ---------------------------------------------------------------------------

func TestAggregatedFeedback_ZeroValue(t *testing.T) {
	var af AggregatedFeedback
	assert.Equal(t, 0, af.TotalCount)
	assert.Equal(t, float64(0), af.AverageScore)
	assert.Nil(t, af.ScoreDistribution)
	assert.Nil(t, af.TypeDistribution)
	assert.Nil(t, af.SourceDistribution)
	assert.Nil(t, af.DimensionAverages)
	assert.Nil(t, af.ProviderStats)
	assert.Nil(t, af.TrendData)
}

func TestAggregatedFeedback_FullPopulation(t *testing.T) {
	af := AggregatedFeedback{
		TotalCount:  150,
		AverageScore: 0.72,
		ScoreDistribution: map[string]int{
			"0.0-0.2": 5,
			"0.2-0.4": 10,
			"0.4-0.6": 30,
			"0.6-0.8": 60,
			"0.8-1.0": 45,
		},
		TypeDistribution: map[FeedbackType]int{
			FeedbackTypePositive: 100,
			FeedbackTypeNegative: 30,
			FeedbackTypeNeutral:  20,
		},
		SourceDistribution: map[FeedbackSource]int{
			FeedbackSourceHuman: 80,
			FeedbackSourceAI:    70,
		},
		DimensionAverages: map[DimensionType]float64{
			DimensionAccuracy:    0.85,
			DimensionRelevance:   0.78,
			DimensionHelpfulness: 0.80,
		},
		ProviderStats: map[string]*ProviderFeedbackStats{
			"claude": {
				ProviderName: "claude",
				TotalCount:   50,
				AverageScore: 0.82,
				Dimensions: map[DimensionType]float64{
					DimensionAccuracy: 0.88,
				},
			},
		},
		TrendData: []*TrendPoint{
			{Timestamp: time.Now(), AverageScore: 0.70, Count: 10},
		},
	}

	assert.Equal(t, 150, af.TotalCount)
	assert.Equal(t, 0.72, af.AverageScore)
	assert.Len(t, af.ScoreDistribution, 5)
	assert.Len(t, af.TypeDistribution, 3)
	assert.Len(t, af.SourceDistribution, 2)
	assert.Equal(t, 0.85, af.DimensionAverages[DimensionAccuracy])
	require.NotNil(t, af.ProviderStats["claude"])
	assert.Equal(t, 50, af.ProviderStats["claude"].TotalCount)
	assert.Len(t, af.TrendData, 1)
}

// ---------------------------------------------------------------------------
// ProviderFeedbackStats struct
// ---------------------------------------------------------------------------

func TestProviderFeedbackStats_ZeroValue(t *testing.T) {
	var pfs ProviderFeedbackStats
	assert.Empty(t, pfs.ProviderName)
	assert.Equal(t, 0, pfs.TotalCount)
	assert.Equal(t, float64(0), pfs.AverageScore)
	assert.Nil(t, pfs.Dimensions)
}

func TestProviderFeedbackStats_FullPopulation(t *testing.T) {
	pfs := ProviderFeedbackStats{
		ProviderName: "openai",
		TotalCount:   200,
		AverageScore: 0.76,
		Dimensions: map[DimensionType]float64{
			DimensionAccuracy:    0.80,
			DimensionCoherence:   0.85,
			DimensionHelpfulness: 0.72,
		},
	}

	assert.Equal(t, "openai", pfs.ProviderName)
	assert.Equal(t, 200, pfs.TotalCount)
	assert.Equal(t, 0.76, pfs.AverageScore)
	assert.Len(t, pfs.Dimensions, 3)
}

// ---------------------------------------------------------------------------
// TrendPoint struct
// ---------------------------------------------------------------------------

func TestTrendPoint_ZeroValue(t *testing.T) {
	var tp TrendPoint
	assert.True(t, tp.Timestamp.IsZero())
	assert.Equal(t, float64(0), tp.AverageScore)
	assert.Equal(t, 0, tp.Count)
}

func TestTrendPoint_FullPopulation(t *testing.T) {
	now := time.Now()
	tp := TrendPoint{
		Timestamp:    now,
		AverageScore: 0.81,
		Count:        42,
	}
	assert.Equal(t, now, tp.Timestamp)
	assert.Equal(t, 0.81, tp.AverageScore)
	assert.Equal(t, 42, tp.Count)
}

// ---------------------------------------------------------------------------
// PolicyUpdate struct
// ---------------------------------------------------------------------------

func TestPolicyUpdate_ZeroValue(t *testing.T) {
	var pu PolicyUpdate
	assert.Empty(t, pu.ID)
	assert.Empty(t, pu.OldPolicy)
	assert.Empty(t, pu.NewPolicy)
	assert.Equal(t, PolicyUpdateType(""), pu.UpdateType)
	assert.Empty(t, pu.Change)
	assert.Empty(t, pu.Reason)
	assert.Equal(t, float64(0), pu.ImprovementScore)
	assert.Nil(t, pu.Examples)
	assert.Nil(t, pu.AppliedAt)
	assert.Nil(t, pu.Metadata)
	assert.True(t, pu.CreatedAt.IsZero())
}

func TestPolicyUpdate_FullPopulation(t *testing.T) {
	now := time.Now()
	applied := now.Add(-1 * time.Hour)
	pu := PolicyUpdate{
		ID:               "pu-001",
		OldPolicy:        "Be helpful.",
		NewPolicy:        "Be helpful, accurate, and concise.",
		UpdateType:       PolicyUpdatePromptRefinement,
		Change:           "Added accuracy and conciseness",
		Reason:           "Feedback indicated need for more precise responses",
		ImprovementScore: 0.15,
		Examples: []*TrainingExample{
			{ID: "te-1", Prompt: "Test", Response: "Result"},
		},
		AppliedAt: &applied,
		Metadata:  map[string]interface{}{"iteration": 3},
		CreatedAt: now,
	}

	assert.Equal(t, "pu-001", pu.ID)
	assert.Equal(t, PolicyUpdatePromptRefinement, pu.UpdateType)
	assert.Equal(t, 0.15, pu.ImprovementScore)
	assert.Len(t, pu.Examples, 1)
	require.NotNil(t, pu.AppliedAt)
	assert.Equal(t, applied, *pu.AppliedAt)
	assert.Equal(t, 3, pu.Metadata["iteration"])
}

func TestPolicyUpdate_NilAppliedAt(t *testing.T) {
	pu := PolicyUpdate{
		ID:         "pu-nil",
		UpdateType: PolicyUpdateGuidelineAddition,
	}
	assert.Nil(t, pu.AppliedAt)
}

// ---------------------------------------------------------------------------
// SelfImprovementConfig struct
// ---------------------------------------------------------------------------

func TestSelfImprovementConfig_ZeroValue(t *testing.T) {
	var cfg SelfImprovementConfig
	assert.Empty(t, cfg.RewardModelProvider)
	assert.Empty(t, cfg.RewardModelName)
	assert.Equal(t, float64(0), cfg.MinRewardThreshold)
	assert.False(t, cfg.AutoCollectFeedback)
	assert.Equal(t, 0, cfg.FeedbackBatchSize)
	assert.Equal(t, float64(0), cfg.MinConfidenceForAuto)
	assert.Equal(t, time.Duration(0), cfg.OptimizationInterval)
	assert.Equal(t, 0, cfg.MinExamplesForUpdate)
	assert.Equal(t, 0, cfg.MaxPolicyUpdatesPerDay)
	assert.Nil(t, cfg.ConstitutionalPrinciples)
	assert.False(t, cfg.EnableSelfCritique)
	assert.False(t, cfg.UseDebateForReward)
	assert.False(t, cfg.UseDebateForOptimize)
	assert.Equal(t, 0, cfg.MaxBufferSize)
}

// ---------------------------------------------------------------------------
// DefaultSelfImprovementConfig
// ---------------------------------------------------------------------------

func TestDefaultSelfImprovementConfig_ReturnsNonNil(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	require.NotNil(t, cfg)
}

func TestDefaultSelfImprovementConfig_RewardModelProvider(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, "claude", cfg.RewardModelProvider)
}

func TestDefaultSelfImprovementConfig_RewardModelName(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, "claude-3-sonnet", cfg.RewardModelName)
}

func TestDefaultSelfImprovementConfig_MinRewardThreshold(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 0.5, cfg.MinRewardThreshold)
}

func TestDefaultSelfImprovementConfig_AutoCollectFeedback(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.True(t, cfg.AutoCollectFeedback)
}

func TestDefaultSelfImprovementConfig_FeedbackBatchSize(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 100, cfg.FeedbackBatchSize)
}

func TestDefaultSelfImprovementConfig_MinConfidenceForAuto(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 0.8, cfg.MinConfidenceForAuto)
}

func TestDefaultSelfImprovementConfig_OptimizationInterval(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 24*time.Hour, cfg.OptimizationInterval)
}

func TestDefaultSelfImprovementConfig_MinExamplesForUpdate(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 50, cfg.MinExamplesForUpdate)
}

func TestDefaultSelfImprovementConfig_MaxPolicyUpdatesPerDay(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 3, cfg.MaxPolicyUpdatesPerDay)
}

func TestDefaultSelfImprovementConfig_EnableSelfCritique(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.True(t, cfg.EnableSelfCritique)
}

func TestDefaultSelfImprovementConfig_UseDebateForReward(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.True(t, cfg.UseDebateForReward)
}

func TestDefaultSelfImprovementConfig_UseDebateForOptimize(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.True(t, cfg.UseDebateForOptimize)
}

func TestDefaultSelfImprovementConfig_MaxBufferSize(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	assert.Equal(t, 10000, cfg.MaxBufferSize)
}

func TestDefaultSelfImprovementConfig_ConstitutionalPrinciples(t *testing.T) {
	cfg := DefaultSelfImprovementConfig()
	require.NotNil(t, cfg.ConstitutionalPrinciples)
	assert.Len(t, cfg.ConstitutionalPrinciples, 5)

	expectedPrinciples := []string{
		"Be helpful, harmless, and honest",
		"Avoid generating harmful or misleading content",
		"Respect user privacy and confidentiality",
		"Acknowledge uncertainty when appropriate",
		"Provide balanced perspectives on controversial topics",
	}
	for i, expected := range expectedPrinciples {
		assert.Equal(t, expected, cfg.ConstitutionalPrinciples[i])
	}
}

func TestDefaultSelfImprovementConfig_IndependentInstances(t *testing.T) {
	cfg1 := DefaultSelfImprovementConfig()
	cfg2 := DefaultSelfImprovementConfig()

	cfg1.MinRewardThreshold = 0.99
	cfg1.ConstitutionalPrinciples = append(cfg1.ConstitutionalPrinciples, "extra")

	assert.NotEqual(t, cfg1.MinRewardThreshold, cfg2.MinRewardThreshold)
	assert.NotEqual(t, len(cfg1.ConstitutionalPrinciples), len(cfg2.ConstitutionalPrinciples))
}

// ---------------------------------------------------------------------------
// DebateResult struct
// ---------------------------------------------------------------------------

func TestDebateResult_ZeroValue(t *testing.T) {
	var dr DebateResult
	assert.Empty(t, dr.ID)
	assert.Empty(t, dr.Consensus)
	assert.Equal(t, float64(0), dr.Confidence)
	assert.Nil(t, dr.Participants)
	assert.Nil(t, dr.Votes)
}

func TestDebateResult_FullPopulation(t *testing.T) {
	dr := DebateResult{
		ID:         "debate-001",
		Consensus:  `The response is good. {"score": 0.85, "reasoning": "accurate and helpful"}`,
		Confidence: 0.92,
		Participants: map[string]string{
			"judge-1": "Response A is better",
			"judge-2": "Response A is slightly better",
		},
		Votes: map[string]float64{
			"judge-1": 0.9,
			"judge-2": 0.85,
		},
	}

	assert.Equal(t, "debate-001", dr.ID)
	assert.Contains(t, dr.Consensus, "score")
	assert.Equal(t, 0.92, dr.Confidence)
	assert.Len(t, dr.Participants, 2)
	assert.Len(t, dr.Votes, 2)
}

// ---------------------------------------------------------------------------
// DebateEvaluation struct
// ---------------------------------------------------------------------------

func TestDebateEvaluation_ZeroValue(t *testing.T) {
	var de DebateEvaluation
	assert.Equal(t, float64(0), de.Score)
	assert.Nil(t, de.Dimensions)
	assert.Empty(t, de.Consensus)
	assert.Empty(t, de.DebateID)
	assert.Nil(t, de.ParticipantVotes)
	assert.Equal(t, float64(0), de.Confidence)
}

func TestDebateEvaluation_FullPopulation(t *testing.T) {
	de := DebateEvaluation{
		Score: 0.88,
		Dimensions: map[DimensionType]float64{
			DimensionAccuracy:    0.90,
			DimensionHelpfulness: 0.86,
		},
		Consensus:        "Overall high quality",
		DebateID:         "de-001",
		ParticipantVotes: map[string]float64{"p1": 0.9, "p2": 0.86},
		Confidence:       0.94,
	}

	assert.Equal(t, 0.88, de.Score)
	assert.Len(t, de.Dimensions, 2)
	assert.Equal(t, "de-001", de.DebateID)
	assert.Equal(t, 0.94, de.Confidence)
}

// ---------------------------------------------------------------------------
// DebateComparison struct
// ---------------------------------------------------------------------------

func TestDebateComparison_ZeroValue(t *testing.T) {
	var dc DebateComparison
	assert.Equal(t, 0, dc.PreferredIndex)
	assert.Equal(t, float64(0), dc.Margin)
	assert.Empty(t, dc.Reasoning)
	assert.Empty(t, dc.DebateID)
	assert.Nil(t, dc.ParticipantPrefs)
	assert.Equal(t, float64(0), dc.Confidence)
}

func TestDebateComparison_FullPopulation(t *testing.T) {
	dc := DebateComparison{
		PreferredIndex:   1,
		Margin:           0.65,
		Reasoning:        "Response B was more comprehensive",
		DebateID:         "dc-001",
		ParticipantPrefs: map[string]int{"p1": 1, "p2": 1, "p3": 0},
		Confidence:       0.78,
	}

	assert.Equal(t, 1, dc.PreferredIndex)
	assert.Equal(t, 0.65, dc.Margin)
	assert.Equal(t, "Response B was more comprehensive", dc.Reasoning)
	assert.Len(t, dc.ParticipantPrefs, 3)
}

func TestDebateComparison_PreferredIndexValues(t *testing.T) {
	tests := []struct {
		name  string
		index int
	}{
		{"First", 0},
		{"Second", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := DebateComparison{PreferredIndex: tt.index}
			assert.Equal(t, tt.index, dc.PreferredIndex)
		})
	}
}

// ---------------------------------------------------------------------------
// Interface compliance (compile-time checks)
// ---------------------------------------------------------------------------

// Verify that the interfaces are well-defined by ensuring they can be
// used as type constraints. These are compile-time assertions.
var (
	_ RewardModel          = (*rewardModelImpl)(nil)
	_ FeedbackCollector    = (*feedbackCollectorImpl)(nil)
	_ PolicyOptimizer      = (*policyOptimizerImpl)(nil)
	_ DebateService        = (*debateServiceImpl)(nil)
	_ DebateRewardEvaluator = (*debateRewardEvalImpl)(nil)
	_ LLMProvider          = (*llmProviderImpl)(nil)
	_ ProviderVerifier     = (*providerVerifierImpl)(nil)
)

// Minimal implementations for compile-time interface checks
type rewardModelImpl struct{}

func (r *rewardModelImpl) Score(_ context.Context, _, _ string) (float64, error) {
	return 0, nil
}
func (r *rewardModelImpl) ScoreWithDimensions(_ context.Context, _, _ string) (map[DimensionType]float64, error) {
	return nil, nil
}
func (r *rewardModelImpl) Compare(_ context.Context, _, _, _ string) (*PreferencePair, error) {
	return nil, nil
}
func (r *rewardModelImpl) Train(_ context.Context, _ []*TrainingExample) error { return nil }

type feedbackCollectorImpl struct{}

func (f *feedbackCollectorImpl) Collect(_ context.Context, _ *Feedback) error { return nil }
func (f *feedbackCollectorImpl) GetBySession(_ context.Context, _ string) ([]*Feedback, error) {
	return nil, nil
}
func (f *feedbackCollectorImpl) GetByPrompt(_ context.Context, _ string) ([]*Feedback, error) {
	return nil, nil
}
func (f *feedbackCollectorImpl) GetAggregated(_ context.Context, _ *FeedbackFilter) (*AggregatedFeedback, error) {
	return nil, nil
}
func (f *feedbackCollectorImpl) Export(_ context.Context, _ *FeedbackFilter) ([]*TrainingExample, error) {
	return nil, nil
}

type policyOptimizerImpl struct{ policy string }

func (p *policyOptimizerImpl) Optimize(_ context.Context, _ []*TrainingExample) ([]*PolicyUpdate, error) {
	return nil, nil
}
func (p *policyOptimizerImpl) Apply(_ context.Context, _ *PolicyUpdate) error    { return nil }
func (p *policyOptimizerImpl) Rollback(_ context.Context, _ string) error        { return nil }
func (p *policyOptimizerImpl) GetHistory(_ context.Context, _ int) ([]*PolicyUpdate, error) {
	return nil, nil
}
func (p *policyOptimizerImpl) GetCurrentPolicy() string       { return p.policy }
func (p *policyOptimizerImpl) SetCurrentPolicy(policy string) { p.policy = policy }

type debateServiceImpl struct{}

func (d *debateServiceImpl) RunDebate(_ context.Context, _ string, _ []string) (*DebateResult, error) {
	return nil, nil
}

type debateRewardEvalImpl struct{}

func (d *debateRewardEvalImpl) EvaluateWithDebate(_ context.Context, _, _ string) (*DebateEvaluation, error) {
	return nil, nil
}
func (d *debateRewardEvalImpl) CompareWithDebate(_ context.Context, _, _, _ string) (*DebateComparison, error) {
	return nil, nil
}

type llmProviderImpl struct{}

func (l *llmProviderImpl) Complete(_ context.Context, _, _ string) (string, error) { return "", nil }

type providerVerifierImpl struct{}

func (p *providerVerifierImpl) GetProviderScore(_ string) float64 { return 0 }
func (p *providerVerifierImpl) IsProviderHealthy(_ string) bool   { return false }

func TestInterfaceCompliance_RewardModel(t *testing.T) {
	var rm RewardModel = &rewardModelImpl{}
	assert.NotNil(t, rm)
}

func TestInterfaceCompliance_FeedbackCollector(t *testing.T) {
	var fc FeedbackCollector = &feedbackCollectorImpl{}
	assert.NotNil(t, fc)
}

func TestInterfaceCompliance_PolicyOptimizer(t *testing.T) {
	po := &policyOptimizerImpl{policy: "initial"}
	var opt PolicyOptimizer = po
	assert.Equal(t, "initial", opt.GetCurrentPolicy())
	opt.SetCurrentPolicy("updated")
	assert.Equal(t, "updated", opt.GetCurrentPolicy())
}

func TestInterfaceCompliance_DebateService(t *testing.T) {
	var ds DebateService = &debateServiceImpl{}
	assert.NotNil(t, ds)
}

func TestInterfaceCompliance_DebateRewardEvaluator(t *testing.T) {
	var dre DebateRewardEvaluator = &debateRewardEvalImpl{}
	assert.NotNil(t, dre)
}

func TestInterfaceCompliance_LLMProvider(t *testing.T) {
	var lp LLMProvider = &llmProviderImpl{}
	assert.NotNil(t, lp)
}

func TestInterfaceCompliance_ProviderVerifier(t *testing.T) {
	var pv ProviderVerifier = &providerVerifierImpl{}
	assert.NotNil(t, pv)
	assert.Equal(t, float64(0), pv.GetProviderScore("any"))
	assert.False(t, pv.IsProviderHealthy("any"))
}
