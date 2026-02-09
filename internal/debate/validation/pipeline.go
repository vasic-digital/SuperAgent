// Package validation provides multi-pass validation pipeline for debate outputs.
// Implements Initial → Validation → Polish → Final quality gates per CLAUDE.md requirements.
package validation

import (
	"context"
	"fmt"
	"time"
)

// ValidationPass defines the validation pass type.
type ValidationPass string

const (
	PassInitial    ValidationPass = "initial"    // First correctness check
	PassValidation ValidationPass = "validation" // Deep validation
	PassPolish     ValidationPass = "polish"     // Quality refinement
	PassFinal      ValidationPass = "final"      // Final verification
)

// ValidationResult contains the outcome of a validation pass.
type ValidationResult struct {
	Pass        ValidationPass         `json:"pass"`
	Passed      bool                   `json:"passed"`
	Score       float64                `json:"score"`        // 0-1
	Issues      []*ValidationIssue     `json:"issues"`       // Found issues
	Suggestions []*Suggestion          `json:"suggestions"`  // Improvement suggestions
	Metadata    map[string]interface{} `json:"metadata"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   int64                  `json:"timestamp"`
}

// ValidationIssue represents a problem found during validation.
type ValidationIssue struct {
	Severity    IssueSeverity `json:"severity"`
	Category    IssueCategory `json:"category"`
	Description string        `json:"description"`
	Location    string        `json:"location"` // Line number, function, etc.
	SuggestedFix string        `json:"suggested_fix"`
}

// IssueSeverity defines issue severity levels.
type IssueSeverity string

const (
	SeverityBlocker  IssueSeverity = "blocker"  // Must fix to proceed
	SeverityCritical IssueSeverity = "critical" // Should fix
	SeverityMajor    IssueSeverity = "major"    // Important
	SeverityMinor    IssueSeverity = "minor"    // Nice to fix
	SeverityInfo     IssueSeverity = "info"     // Informational
)

// IssueCategory categorizes validation issues.
type IssueCategory string

const (
	CategoryCorrectness   IssueCategory = "correctness"
	CategoryPerformance   IssueCategory = "performance"
	CategorySecurity      IssueCategory = "security"
	CategoryStyle         IssueCategory = "style"
	CategoryMaintainability IssueCategory = "maintainability"
	CategoryDocumentation IssueCategory = "documentation"
)

// Suggestion provides improvement recommendations.
type Suggestion struct {
	Type        string  `json:"type"`        // "refactor", "optimize", "simplify"
	Priority    int     `json:"priority"`    // 1 (highest) - 5 (lowest)
	Description string  `json:"description"`
	Before      string  `json:"before"`      // Current code
	After       string  `json:"after"`       // Suggested code
	Rationale   string  `json:"rationale"`
	Impact      float64 `json:"impact"`      // Expected improvement (0-1)
}

// ValidationPipeline orchestrates multi-pass validation.
type ValidationPipeline struct {
	validators map[ValidationPass]Validator
	config     *PipelineConfig
}

// PipelineConfig configures the validation pipeline.
type PipelineConfig struct {
	// Pass requirements
	RequireAllPasses    bool              `json:"require_all_passes"`    // All passes must succeed
	MinScoreThreshold   float64           `json:"min_score_threshold"`   // Minimum score to pass (0-1)
	MaxBlockerIssues    int               `json:"max_blocker_issues"`    // Max blocker issues allowed
	StopOnFirstFailure  bool              `json:"stop_on_first_failure"` // Stop on first failed pass

	// Pass-specific configs
	PassConfigs map[ValidationPass]*PassConfig `json:"pass_configs"`
}

// PassConfig configures a single validation pass.
type PassConfig struct {
	Enabled         bool          `json:"enabled"`
	Timeout         time.Duration `json:"timeout"`
	MinScore        float64       `json:"min_score"`
	RequiredChecks  []string      `json:"required_checks"`
}

// NewValidationPipeline creates a validation pipeline.
func NewValidationPipeline(config *PipelineConfig) *ValidationPipeline {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	return &ValidationPipeline{
		validators: make(map[ValidationPass]Validator),
		config:     config,
	}
}

// DefaultPipelineConfig returns default pipeline configuration.
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		RequireAllPasses:   true,
		MinScoreThreshold:  0.8, // 80%
		MaxBlockerIssues:   0,   // No blockers allowed
		StopOnFirstFailure: false,
		PassConfigs: map[ValidationPass]*PassConfig{
			PassInitial: {
				Enabled:  true,
				Timeout:  30 * time.Second,
				MinScore: 0.6, // 60% for initial pass
				RequiredChecks: []string{"syntax", "basic_correctness"},
			},
			PassValidation: {
				Enabled:  true,
				Timeout:  60 * time.Second,
				MinScore: 0.75, // 75% for validation pass
				RequiredChecks: []string{"logic", "edge_cases", "error_handling"},
			},
			PassPolish: {
				Enabled:  true,
				Timeout:  45 * time.Second,
				MinScore: 0.85, // 85% for polish pass
				RequiredChecks: []string{"style", "performance", "maintainability"},
			},
			PassFinal: {
				Enabled:  true,
				Timeout:  60 * time.Second,
				MinScore: 0.9, // 90% for final pass
				RequiredChecks: []string{"security", "documentation", "best_practices"},
			},
		},
	}
}

// RegisterValidator registers a validator for a specific pass.
func (p *ValidationPipeline) RegisterValidator(pass ValidationPass, validator Validator) {
	p.validators[pass] = validator
}

// Validate executes the full multi-pass validation pipeline.
func (p *ValidationPipeline) Validate(ctx context.Context, artifact *Artifact) (*PipelineResult, error) {
	result := &PipelineResult{
		Artifact:      artifact,
		PassResults:   make(map[ValidationPass]*ValidationResult),
		OverallPassed: true,
		StartTime:     time.Now(),
	}

	// Execute passes in order: Initial → Validation → Polish → Final
	passes := []ValidationPass{PassInitial, PassValidation, PassPolish, PassFinal}

	for _, pass := range passes {
		passConfig := p.config.PassConfigs[pass]
		if passConfig == nil || !passConfig.Enabled {
			continue
		}

		validator := p.validators[pass]
		if validator == nil {
			return nil, fmt.Errorf("no validator registered for pass: %s", pass)
		}

		// Execute pass with timeout
		passCtx, cancel := context.WithTimeout(ctx, passConfig.Timeout)
		passResult, err := validator.Validate(passCtx, artifact)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("validation pass %s failed: %w", pass, err)
		}

		result.PassResults[pass] = passResult

		// Check if pass succeeded
		passPassed := passResult.Passed &&
			passResult.Score >= passConfig.MinScore &&
			p.hasNoBlockers(passResult)

		if !passPassed {
			result.OverallPassed = false
			result.FailedPass = pass

			if p.config.StopOnFirstFailure {
				break
			}
		}
	}

	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)

	// Overall score is average of pass scores
	result.OverallScore = p.calculateOverallScore(result.PassResults)

	return result, nil
}

// hasNoBlockers checks if there are any blocker issues.
func (p *ValidationPipeline) hasNoBlockers(result *ValidationResult) bool {
	for _, issue := range result.Issues {
		if issue.Severity == SeverityBlocker {
			return false
		}
	}
	return true
}

// calculateOverallScore calculates the overall validation score.
func (p *ValidationPipeline) calculateOverallScore(results map[ValidationPass]*ValidationResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	var total float64
	for _, result := range results {
		total += result.Score
	}

	return total / float64(len(results))
}

// PipelineResult contains the results of the full validation pipeline.
type PipelineResult struct {
	Artifact      *Artifact                            `json:"artifact"`
	PassResults   map[ValidationPass]*ValidationResult `json:"pass_results"`
	OverallPassed bool                                 `json:"overall_passed"`
	OverallScore  float64                              `json:"overall_score"`
	FailedPass    ValidationPass                       `json:"failed_pass,omitempty"`
	StartTime     time.Time                            `json:"start_time"`
	EndTime       time.Time                            `json:"end_time"`
	TotalDuration time.Duration                        `json:"total_duration"`
}

// Artifact represents the item being validated.
type Artifact struct {
	Type     ArtifactType           `json:"type"`
	Content  string                 `json:"content"`
	Language string                 `json:"language"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ArtifactType defines types of artifacts.
type ArtifactType string

const (
	ArtifactCode         ArtifactType = "code"
	ArtifactArchitecture ArtifactType = "architecture"
	ArtifactDocumentation ArtifactType = "documentation"
	ArtifactTest         ArtifactType = "test"
)

// Validator performs validation for a specific pass.
type Validator interface {
	Validate(ctx context.Context, artifact *Artifact) (*ValidationResult, error)
}
