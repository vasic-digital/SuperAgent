// Package validation provides tests for validation pipeline.
package validation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockValidator is a mock validator for testing.
type MockValidator struct {
	passResult *ValidationResult
	failResult error
}

func (m *MockValidator) Validate(ctx context.Context, artifact *Artifact) (*ValidationResult, error) {
	if m.failResult != nil {
		return nil, m.failResult
	}
	return m.passResult, nil
}

func TestValidationPipeline_DefaultConfig(t *testing.T) {
	config := DefaultPipelineConfig()

	assert.True(t, config.RequireAllPasses)
	assert.Equal(t, 0.8, config.MinScoreThreshold)
	assert.Equal(t, 0, config.MaxBlockerIssues)
	assert.Len(t, config.PassConfigs, 4)
}

func TestValidationPipeline_RegisterValidator(t *testing.T) {
	pipeline := NewValidationPipeline(nil)
	validator := &MockValidator{}

	pipeline.RegisterValidator(PassInitial, validator)

	assert.NotNil(t, pipeline.validators[PassInitial])
}

func TestValidationPipeline_Validate_AllPass(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewValidationPipeline(config)

	// Register mock validators for all passes
	for _, pass := range []ValidationPass{PassInitial, PassValidation, PassPolish, PassFinal} {
		pipeline.RegisterValidator(pass, &MockValidator{
			passResult: &ValidationResult{
				Pass:   pass,
				Passed: true,
				Score:  0.9,
				Issues: []*ValidationIssue{},
			},
		})
	}

	artifact := &Artifact{
		Type:    ArtifactCode,
		Content: "func test() {}",
		Language: "go",
	}

	ctx := context.Background()
	result, err := pipeline.Validate(ctx, artifact)

	require.NoError(t, err)
	assert.True(t, result.OverallPassed)
	assert.Len(t, result.PassResults, 4)
	assert.GreaterOrEqual(t, result.OverallScore, 0.85)
}

func TestValidationPipeline_Validate_InitialFail(t *testing.T) {
	config := DefaultPipelineConfig()
	config.StopOnFirstFailure = true
	pipeline := NewValidationPipeline(config)

	// Initial pass fails
	pipeline.RegisterValidator(PassInitial, &MockValidator{
		passResult: &ValidationResult{
			Pass:   PassInitial,
			Passed: false,
			Score:  0.3,
			Issues: []*ValidationIssue{
				{Severity: SeverityBlocker, Description: "Syntax error"},
			},
		},
	})

	artifact := &Artifact{
		Type:    ArtifactCode,
		Content: "invalid code",
		Language: "go",
	}

	ctx := context.Background()
	result, err := pipeline.Validate(ctx, artifact)

	require.NoError(t, err)
	assert.False(t, result.OverallPassed)
	assert.Equal(t, PassInitial, result.FailedPass)
	assert.Len(t, result.PassResults, 1) // Only initial pass ran
}

func TestValidationPipeline_Validate_WithBlockers(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewValidationPipeline(config)

	pipeline.RegisterValidator(PassInitial, &MockValidator{
		passResult: &ValidationResult{
			Pass:   PassInitial,
			Passed: true,
			Score:  0.7,
			Issues: []*ValidationIssue{
				{Severity: SeverityBlocker, Description: "Critical bug"},
			},
		},
	})

	artifact := &Artifact{
		Type:    ArtifactCode,
		Content: "func test() { panic() }",
		Language: "go",
	}

	ctx := context.Background()
	result, err := pipeline.Validate(ctx, artifact)

	require.NoError(t, err)
	assert.False(t, result.OverallPassed)
}

func TestValidationPipeline_Validate_Timeout(t *testing.T) {
	config := DefaultPipelineConfig()
	config.PassConfigs[PassInitial].Timeout = 1 * time.Millisecond
	pipeline := NewValidationPipeline(config)

	// Validator that takes too long
	pipeline.RegisterValidator(PassInitial, &MockValidator{
		passResult: &ValidationResult{
			Pass:   PassInitial,
			Passed: true,
			Score:  0.9,
		},
	})

	artifact := &Artifact{
		Type:    ArtifactCode,
		Content: "func test() {}",
		Language: "go",
	}

	ctx := context.Background()
	_, err := pipeline.Validate(ctx, artifact)

	// Should handle timeout gracefully
	assert.Error(t, err)
}

func TestValidationResult_HasBlockers(t *testing.T) {
	tests := []struct {
		name     string
		issues   []*ValidationIssue
		expected bool
	}{
		{
			name: "No blockers",
			issues: []*ValidationIssue{
				{Severity: SeverityMinor},
				{Severity: SeverityInfo},
			},
			expected: false,
		},
		{
			name: "Has blocker",
			issues: []*ValidationIssue{
				{Severity: SeverityMinor},
				{Severity: SeverityBlocker},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Issues: tt.issues,
			}

			pipeline := NewValidationPipeline(nil)
			hasBlockers := !pipeline.hasNoBlockers(result)
			assert.Equal(t, tt.expected, hasBlockers)
		})
	}
}

func TestValidationPipeline_CalculateOverallScore(t *testing.T) {
	pipeline := NewValidationPipeline(nil)

	results := map[ValidationPass]*ValidationResult{
		PassInitial:    {Score: 0.8},
		PassValidation: {Score: 0.9},
		PassPolish:     {Score: 0.85},
		PassFinal:      {Score: 0.95},
	}

	score := pipeline.calculateOverallScore(results)

	expected := (0.8 + 0.9 + 0.85 + 0.95) / 4.0
	assert.InDelta(t, expected, score, 0.01)
}
