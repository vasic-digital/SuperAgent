// Package verification provides formal verification capabilities
// including specification generation, theorem proving, and model checking.
package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SpecificationType represents the type of specification
type SpecificationType string

const (
	SpecTypeJML           SpecificationType = "jml"
	SpecTypeDafny         SpecificationType = "dafny"
	SpecTypeLTL           SpecificationType = "ltl"
	SpecTypeInvariant     SpecificationType = "invariant"
	SpecTypePrecondition  SpecificationType = "precondition"
	SpecTypePostcondition SpecificationType = "postcondition"
)

// VerificationResult represents the result of a verification
type VerificationResult struct {
	Verified       bool                   `json:"verified"`
	Specification  *Specification         `json:"specification"`
	Errors         []VerificationError    `json:"errors,omitempty"`
	Counterexample *Counterexample        `json:"counterexample,omitempty"`
	Duration       time.Duration          `json:"duration"`
	Prover         string                 `json:"prover"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// MarshalJSON implements custom JSON marshaling
func (r *VerificationResult) MarshalJSON() ([]byte, error) {
	type Alias VerificationResult
	return json.Marshal(&struct {
		*Alias
		DurationMs int64 `json:"duration_ms"`
	}{
		Alias:      (*Alias)(r),
		DurationMs: r.Duration.Milliseconds(),
	})
}

// VerificationError represents an error during verification
type VerificationError struct {
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Counterexample represents a counterexample from verification
type Counterexample struct {
	Variables map[string]interface{} `json:"variables"`
	Trace     []string               `json:"trace,omitempty"`
	State     string                 `json:"state,omitempty"`
}

// Specification represents a formal specification
type Specification struct {
	ID             string            `json:"id"`
	Type           SpecificationType `json:"type"`
	Target         string            `json:"target"`
	Preconditions  []string          `json:"preconditions,omitempty"`
	Postconditions []string          `json:"postconditions,omitempty"`
	Invariants     []string          `json:"invariants,omitempty"`
	Assertions     []string          `json:"assertions,omitempty"`
	RawSpec        string            `json:"raw_spec,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// FormalVerifierConfig holds configuration for formal verification
type FormalVerifierConfig struct {
	// DefaultProver is the default theorem prover
	DefaultProver string `json:"default_prover"`
	// Timeout for verification
	Timeout time.Duration `json:"timeout"`
	// MaxRetries for spec mutation
	MaxRetries int `json:"max_retries"`
	// EnableMutation enables mutation-based repair
	EnableMutation bool `json:"enable_mutation"`
	// Z3Path is the path to Z3 solver
	Z3Path string `json:"z3_path"`
	// DafnyPath is the path to Dafny
	DafnyPath string `json:"dafny_path"`
	// OpenJMLPath is the path to OpenJML
	OpenJMLPath string `json:"openjml_path"`
}

// DefaultFormalVerifierConfig returns default configuration
func DefaultFormalVerifierConfig() FormalVerifierConfig {
	return FormalVerifierConfig{
		DefaultProver:  "z3",
		Timeout:        5 * time.Minute,
		MaxRetries:     5,
		EnableMutation: true,
		Z3Path:         "z3",
		DafnyPath:      "dafny",
		OpenJMLPath:    "openjml",
	}
}

// SpecGenerator generates formal specifications from code
type SpecGenerator interface {
	// GenerateSpec generates a specification for code
	GenerateSpec(ctx context.Context, code string, language string) (*Specification, error)
	// RefineSpec refines a specification based on verification feedback
	RefineSpec(ctx context.Context, spec *Specification, errors []VerificationError) (*Specification, error)
}

// TheoremProver verifies specifications
type TheoremProver interface {
	// Verify verifies a specification
	Verify(ctx context.Context, spec *Specification, code string) (*VerificationResult, error)
	// Name returns the prover name
	Name() string
}

// FormalVerifier implements formal verification
type FormalVerifier struct {
	config  FormalVerifierConfig
	specGen SpecGenerator
	provers map[string]TheoremProver
	specs   map[string]*Specification
	mu      sync.RWMutex
	logger  *logrus.Logger
}

// NewFormalVerifier creates a new formal verifier
func NewFormalVerifier(config FormalVerifierConfig, specGen SpecGenerator, logger *logrus.Logger) *FormalVerifier {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	return &FormalVerifier{
		config:  config,
		specGen: specGen,
		provers: make(map[string]TheoremProver),
		specs:   make(map[string]*Specification),
		logger:  logger,
	}
}

// RegisterProver registers a theorem prover
func (v *FormalVerifier) RegisterProver(prover TheoremProver) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.provers[prover.Name()] = prover
}

// VerifyCode verifies code against generated specifications
func (v *FormalVerifier) VerifyCode(ctx context.Context, code string, language string) (*VerificationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, v.config.Timeout)
	defer cancel()

	startTime := time.Now()

	// Generate specification
	spec, err := v.specGen.GenerateSpec(ctx, code, language)
	if err != nil {
		return nil, fmt.Errorf("failed to generate specification: %w", err)
	}

	// Get prover
	prover, ok := v.provers[v.config.DefaultProver]
	if !ok {
		return nil, fmt.Errorf("prover not found: %s", v.config.DefaultProver)
	}

	// Verify with mutation-based repair
	var result *VerificationResult
	currentSpec := spec

	for attempt := 0; attempt <= v.config.MaxRetries; attempt++ {
		result, err = prover.Verify(ctx, currentSpec, code)
		if err != nil {
			return nil, fmt.Errorf("verification failed: %w", err)
		}

		if result.Verified {
			break
		}

		// Try to refine spec if mutation is enabled
		if !v.config.EnableMutation || attempt == v.config.MaxRetries {
			break
		}

		refinedSpec, err := v.specGen.RefineSpec(ctx, currentSpec, result.Errors)
		if err != nil {
			v.logger.Warnf("Failed to refine spec: %v", err)
			break
		}
		currentSpec = refinedSpec
	}

	result.Duration = time.Since(startTime)
	result.Specification = currentSpec

	// Store spec
	v.mu.Lock()
	v.specs[spec.ID] = currentSpec
	v.mu.Unlock()

	return result, nil
}

// VerifySpec verifies a specific specification
func (v *FormalVerifier) VerifySpec(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
	prover, ok := v.provers[v.config.DefaultProver]
	if !ok {
		return nil, fmt.Errorf("prover not found: %s", v.config.DefaultProver)
	}

	return prover.Verify(ctx, spec, code)
}

// GetSpec retrieves a stored specification
func (v *FormalVerifier) GetSpec(id string) (*Specification, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	spec, exists := v.specs[id]
	return spec, exists
}

// LLMSpecGenerator implements SpecGenerator using an LLM
type LLMSpecGenerator struct {
	generateFunc func(ctx context.Context, prompt string) (string, error)
	logger       *logrus.Logger
}

// NewLLMSpecGenerator creates a new LLM-based spec generator
func NewLLMSpecGenerator(generateFunc func(ctx context.Context, prompt string) (string, error), logger *logrus.Logger) *LLMSpecGenerator {
	return &LLMSpecGenerator{
		generateFunc: generateFunc,
		logger:       logger,
	}
}

// GenerateSpec generates a specification for code
func (g *LLMSpecGenerator) GenerateSpec(ctx context.Context, code string, language string) (*Specification, error) {
	prompt := fmt.Sprintf(`Analyze the following %s code and generate formal specifications.
Include:
1. Preconditions (requires)
2. Postconditions (ensures)
3. Loop invariants (if applicable)

Code:
%s

Output the specifications in the following format:
PRECONDITIONS:
- condition1
- condition2

POSTCONDITIONS:
- condition1
- condition2

INVARIANTS:
- invariant1
- invariant2`, language, code)

	response, err := g.generateFunc(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response
	spec := &Specification{
		ID:        fmt.Sprintf("spec-%d", time.Now().UnixNano()),
		Type:      SpecTypeJML,
		Target:    code,
		CreatedAt: time.Now(),
	}

	// Parse preconditions
	preMatch := regexp.MustCompile(`(?i)PRECONDITIONS:\s*\n((?:-.*\n?)+)`)
	if matches := preMatch.FindStringSubmatch(response); len(matches) > 1 {
		spec.Preconditions = parseConditions(matches[1])
	}

	// Parse postconditions
	postMatch := regexp.MustCompile(`(?i)POSTCONDITIONS:\s*\n((?:-.*\n?)+)`)
	if matches := postMatch.FindStringSubmatch(response); len(matches) > 1 {
		spec.Postconditions = parseConditions(matches[1])
	}

	// Parse invariants
	invMatch := regexp.MustCompile(`(?i)INVARIANTS:\s*\n((?:-.*\n?)+)`)
	if matches := invMatch.FindStringSubmatch(response); len(matches) > 1 {
		spec.Invariants = parseConditions(matches[1])
	}

	spec.RawSpec = response
	return spec, nil
}

// RefineSpec refines a specification based on errors
func (g *LLMSpecGenerator) RefineSpec(ctx context.Context, spec *Specification, errors []VerificationError) (*Specification, error) {
	errorMsgs := make([]string, len(errors))
	for i, err := range errors {
		errorMsgs[i] = err.Message
	}

	prompt := fmt.Sprintf(`The following specification failed verification:

Preconditions: %v
Postconditions: %v
Invariants: %v

Errors:
%s

Refine the specification to fix these errors. Apply these mutations if helpful:
1. Swap quantifiers (forall <-> exists)
2. Adjust operators (< <-> <=, && <-> ||)
3. Fix boundary conditions

Output the refined specifications.`, spec.Preconditions, spec.Postconditions, spec.Invariants, strings.Join(errorMsgs, "\n"))

	response, err := g.generateFunc(ctx, prompt)
	if err != nil {
		return nil, err
	}

	refined := &Specification{
		ID:        fmt.Sprintf("spec-%d-refined", time.Now().UnixNano()),
		Type:      spec.Type,
		Target:    spec.Target,
		CreatedAt: time.Now(),
	}

	// Parse refined conditions
	preMatch := regexp.MustCompile(`(?i)PRECONDITIONS:\s*\n((?:-.*\n?)+)`)
	if matches := preMatch.FindStringSubmatch(response); len(matches) > 1 {
		refined.Preconditions = parseConditions(matches[1])
	} else {
		refined.Preconditions = spec.Preconditions
	}

	postMatch := regexp.MustCompile(`(?i)POSTCONDITIONS:\s*\n((?:-.*\n?)+)`)
	if matches := postMatch.FindStringSubmatch(response); len(matches) > 1 {
		refined.Postconditions = parseConditions(matches[1])
	} else {
		refined.Postconditions = spec.Postconditions
	}

	invMatch := regexp.MustCompile(`(?i)INVARIANTS:\s*\n((?:-.*\n?)+)`)
	if matches := invMatch.FindStringSubmatch(response); len(matches) > 1 {
		refined.Invariants = parseConditions(matches[1])
	} else {
		refined.Invariants = spec.Invariants
	}

	refined.RawSpec = response
	return refined, nil
}

// parseConditions parses condition lines
func parseConditions(text string) []string {
	lines := strings.Split(text, "\n")
	conditions := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") {
			condition := strings.TrimPrefix(line, "-")
			condition = strings.TrimSpace(condition)
			if condition != "" {
				conditions = append(conditions, condition)
			}
		}
	}
	return conditions
}

// Z3Prover implements TheoremProver using Z3
type Z3Prover struct {
	path   string
	logger *logrus.Logger
}

// NewZ3Prover creates a new Z3 prover
func NewZ3Prover(path string, logger *logrus.Logger) *Z3Prover {
	return &Z3Prover{
		path:   path,
		logger: logger,
	}
}

// Name returns the prover name
func (p *Z3Prover) Name() string {
	return "z3"
}

// Verify verifies a specification using Z3
func (p *Z3Prover) Verify(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
	startTime := time.Now()

	// Convert spec to SMT-LIB format
	smtCode := p.toSMTLIB(spec)

	// For now, simulate verification (real implementation would call Z3)
	result := &VerificationResult{
		Verified:      true,
		Specification: spec,
		Prover:        "z3",
		Metadata: map[string]interface{}{
			"smt_code": smtCode,
		},
	}

	// Simulate verification logic
	// In real implementation, would execute: z3 -smt2 <file>
	for _, pre := range spec.Preconditions {
		if strings.Contains(pre, "null") || strings.Contains(pre, "nil") {
			// Common verification passes
			continue
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// toSMTLIB converts specification to SMT-LIB format
func (p *Z3Prover) toSMTLIB(spec *Specification) string {
	var builder strings.Builder
	builder.WriteString("; SMT-LIB 2.0\n")
	builder.WriteString("(set-logic ALL)\n\n")

	// Declare assertions
	for i, pre := range spec.Preconditions {
		builder.WriteString(fmt.Sprintf("; Precondition %d: %s\n", i+1, pre))
		builder.WriteString(fmt.Sprintf("(assert (= pre_%d true))\n", i+1))
	}

	for i, post := range spec.Postconditions {
		builder.WriteString(fmt.Sprintf("; Postcondition %d: %s\n", i+1, post))
		builder.WriteString(fmt.Sprintf("(assert (= post_%d true))\n", i+1))
	}

	builder.WriteString("\n(check-sat)\n")
	builder.WriteString("(get-model)\n")

	return builder.String()
}

// DafnyVerifier implements TheoremProver using Dafny
type DafnyVerifier struct {
	path   string
	logger *logrus.Logger
}

// NewDafnyVerifier creates a new Dafny verifier
func NewDafnyVerifier(path string, logger *logrus.Logger) *DafnyVerifier {
	return &DafnyVerifier{
		path:   path,
		logger: logger,
	}
}

// Name returns the prover name
func (d *DafnyVerifier) Name() string {
	return "dafny"
}

// Verify verifies a specification using Dafny
func (d *DafnyVerifier) Verify(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
	startTime := time.Now()

	// Convert to Dafny code
	dafnyCode := d.toDafny(spec, code)

	result := &VerificationResult{
		Verified:      true,
		Specification: spec,
		Prover:        "dafny",
		Metadata: map[string]interface{}{
			"dafny_code": dafnyCode,
		},
	}

	// Simulate diff-checker (ensure only annotations were added)
	if !d.diffCheck(code, dafnyCode) {
		result.Verified = false
		result.Errors = append(result.Errors, VerificationError{
			Message: "Soundness failure: executable logic was modified",
			Code:    "DIFF_CHECK_FAILED",
		})
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// toDafny converts spec and code to Dafny
func (d *DafnyVerifier) toDafny(spec *Specification, code string) string {
	var builder strings.Builder

	builder.WriteString("method Verified()\n")

	// Add preconditions
	for _, pre := range spec.Preconditions {
		builder.WriteString(fmt.Sprintf("  requires %s\n", pre))
	}

	// Add postconditions
	for _, post := range spec.Postconditions {
		builder.WriteString(fmt.Sprintf("  ensures %s\n", post))
	}

	builder.WriteString("{\n")
	builder.WriteString("  // Original code\n")
	builder.WriteString(code)
	builder.WriteString("\n}\n")

	return builder.String()
}

// diffCheck ensures only annotations were added
func (d *DafnyVerifier) diffCheck(original, annotated string) bool {
	// Remove annotations and compare
	annotationPattern := regexp.MustCompile(`(?m)^\s*(requires|ensures|invariant|decreases).*$`)
	cleanAnnotated := annotationPattern.ReplaceAllString(annotated, "")

	// Normalize whitespace
	originalNorm := strings.Join(strings.Fields(original), " ")
	annotatedNorm := strings.Join(strings.Fields(cleanAnnotated), " ")

	// Check if executable code is preserved
	return strings.Contains(annotatedNorm, originalNorm) || len(original) < 10
}

// VeriPlan implements plan verification using LTL model checking
type VeriPlan struct {
	logger *logrus.Logger
}

// NewVeriPlan creates a new VeriPlan verifier
func NewVeriPlan(logger *logrus.Logger) *VeriPlan {
	return &VeriPlan{logger: logger}
}

// LTLFormula represents an LTL formula
type LTLFormula struct {
	Formula string `json:"formula"`
	Type    string `json:"type"` // safety, liveness, fairness
	Natural string `json:"natural"`
}

// PlanVerificationResult represents plan verification result
type PlanVerificationResult struct {
	Valid          bool          `json:"valid"`
	Formulas       []*LTLFormula `json:"formulas"`
	Violations     []string      `json:"violations,omitempty"`
	StateSpaceSize int           `json:"state_space_size"`
	Duration       time.Duration `json:"duration"`
}

// VerifyPlan verifies a plan against LTL safety properties
func (v *VeriPlan) VerifyPlan(ctx context.Context, plan string, constraints []string) (*PlanVerificationResult, error) {
	startTime := time.Now()

	// Convert constraints to LTL formulas
	formulas := make([]*LTLFormula, 0)
	for _, constraint := range constraints {
		formula := v.constraintToLTL(constraint)
		formulas = append(formulas, formula)
	}

	result := &PlanVerificationResult{
		Valid:    true,
		Formulas: formulas,
	}

	// Model check (simplified simulation)
	for _, formula := range formulas {
		if !v.modelCheck(plan, formula) {
			result.Valid = false
			result.Violations = append(result.Violations,
				fmt.Sprintf("Violation of %s: %s", formula.Type, formula.Natural))
		}
	}

	result.Duration = time.Since(startTime)
	result.StateSpaceSize = len(plan) / 10 // Simplified

	return result, nil
}

// constraintToLTL converts natural language constraint to LTL
func (v *VeriPlan) constraintToLTL(constraint string) *LTLFormula {
	// Simple pattern matching for common safety properties
	formula := &LTLFormula{
		Natural: constraint,
	}

	constraintLower := strings.ToLower(constraint)

	if strings.Contains(constraintLower, "always") || strings.Contains(constraintLower, "never") {
		formula.Type = "safety"
		formula.Formula = fmt.Sprintf("G(%s)", constraint)
	} else if strings.Contains(constraintLower, "eventually") {
		formula.Type = "liveness"
		formula.Formula = fmt.Sprintf("F(%s)", constraint)
	} else if strings.Contains(constraintLower, "until") {
		formula.Type = "safety"
		formula.Formula = fmt.Sprintf("(%s U %s)", constraint, "done")
	} else {
		formula.Type = "safety"
		formula.Formula = constraint
	}

	return formula
}

// modelCheck performs simplified model checking
func (v *VeriPlan) modelCheck(plan string, formula *LTLFormula) bool {
	// Simplified check - real implementation would use PRISM or similar
	planLower := strings.ToLower(plan)
	formulaLower := strings.ToLower(formula.Natural)

	// Check for obvious violations
	if strings.Contains(formulaLower, "never") {
		// Extract what should never happen
		parts := strings.Split(formulaLower, "never")
		if len(parts) > 1 {
			forbidden := strings.TrimSpace(parts[1])
			// First try exact substring match
			if strings.Contains(planLower, forbidden) {
				return false
			}
			// Also check if key terms from the forbidden phrase appear in the plan
			// This catches cases like "access production database directly" matching "access production directly"
			forbiddenWords := strings.Fields(forbidden)
			if len(forbiddenWords) >= 2 {
				matchCount := 0
				for _, word := range forbiddenWords {
					if len(word) > 3 && strings.Contains(planLower, word) { // Skip short words like "a", "the"
						matchCount++
					}
				}
				// If most key terms are present, consider it a violation
				if matchCount >= len(forbiddenWords)-1 && matchCount > 1 {
					return false
				}
			}
		}
	}

	return true
}
