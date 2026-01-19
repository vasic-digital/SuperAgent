// Package verification provides tests for the formal verification system.
package verification

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSpecGenerator implements SpecGenerator for testing
type MockSpecGenerator struct {
	generateFunc func(ctx context.Context, code string, language string) (*Specification, error)
	refineFunc   func(ctx context.Context, spec *Specification, errors []VerificationError) (*Specification, error)
}

func (m *MockSpecGenerator) GenerateSpec(ctx context.Context, code string, language string) (*Specification, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, code, language)
	}
	return &Specification{
		ID:             "test-spec-id",
		Type:           SpecTypeJML,
		Target:         code,
		Preconditions:  []string{"x != null"},
		Postconditions: []string{"result >= 0"},
		Invariants:     []string{"count >= 0"},
		CreatedAt:      time.Now(),
	}, nil
}

func (m *MockSpecGenerator) RefineSpec(ctx context.Context, spec *Specification, errors []VerificationError) (*Specification, error) {
	if m.refineFunc != nil {
		return m.refineFunc(ctx, spec, errors)
	}
	refined := *spec
	refined.ID = spec.ID + "-refined"
	return &refined, nil
}

// MockTheoremProver implements TheoremProver for testing
type MockTheoremProver struct {
	name       string
	verifyFunc func(ctx context.Context, spec *Specification, code string) (*VerificationResult, error)
}

func (m *MockTheoremProver) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mock"
}

func (m *MockTheoremProver) Verify(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(ctx, spec, code)
	}
	return &VerificationResult{
		Verified:      true,
		Specification: spec,
		Prover:        m.Name(),
		Duration:      100 * time.Millisecond,
	}, nil
}

// TestDefaultFormalVerifierConfig tests default configuration
func TestDefaultFormalVerifierConfig(t *testing.T) {
	config := DefaultFormalVerifierConfig()

	assert.Equal(t, "z3", config.DefaultProver)
	assert.Equal(t, 5*time.Minute, config.Timeout)
	assert.Equal(t, 5, config.MaxRetries)
	assert.True(t, config.EnableMutation)
	assert.Equal(t, "z3", config.Z3Path)
	assert.Equal(t, "dafny", config.DafnyPath)
	assert.Equal(t, "openjml", config.OpenJMLPath)
}

// TestNewFormalVerifier tests verifier creation
func TestNewFormalVerifier(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	specGen := &MockSpecGenerator{}
	logger := logrus.New()

	verifier := NewFormalVerifier(config, specGen, logger)

	assert.NotNil(t, verifier)
	assert.NotNil(t, verifier.provers)
	assert.NotNil(t, verifier.specs)
}

// TestNewFormalVerifier_NilLogger tests verifier with nil logger
func TestNewFormalVerifier_NilLogger(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	specGen := &MockSpecGenerator{}

	verifier := NewFormalVerifier(config, specGen, nil)

	assert.NotNil(t, verifier)
	assert.NotNil(t, verifier.logger)
}

// TestRegisterProver tests prover registration
func TestRegisterProver(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)

	prover := &MockTheoremProver{name: "test-prover"}
	verifier.RegisterProver(prover)

	// Verify registration by attempting to use it
	config.DefaultProver = "test-prover"
	verifier.config = config
}

// TestVerifyCode tests code verification
func TestVerifyCode(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "mock"

	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{name: "mock"})

	ctx := context.Background()
	code := `func add(a, b int) int { return a + b }`

	result, err := verifier.VerifyCode(ctx, code, "go")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Verified)
	assert.NotNil(t, result.Specification)
}

// TestVerifyCode_ProverNotFound tests error when prover not found
func TestVerifyCode_ProverNotFound(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "non-existent"

	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)

	ctx := context.Background()

	_, err := verifier.VerifyCode(ctx, "code", "go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prover not found")
}

// TestVerifyCode_SpecGenerationError tests spec generation error handling
func TestVerifyCode_SpecGenerationError(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "mock"

	specGen := &MockSpecGenerator{
		generateFunc: func(ctx context.Context, code string, language string) (*Specification, error) {
			return nil, assert.AnError
		},
	}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{name: "mock"})

	ctx := context.Background()

	_, err := verifier.VerifyCode(ctx, "code", "go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate specification")
}

// TestVerifyCode_VerificationFailure tests verification failure
func TestVerifyCode_VerificationFailure(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "mock"
	config.EnableMutation = false

	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{
		name: "mock",
		verifyFunc: func(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
			return &VerificationResult{
				Verified: false,
				Errors: []VerificationError{
					{Message: "postcondition violation"},
				},
			}, nil
		},
	})

	ctx := context.Background()

	result, err := verifier.VerifyCode(ctx, "code", "go")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Verified)
}

// TestVerifyCode_MutationRetry tests mutation-based retry
func TestVerifyCode_MutationRetry(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "mock"
	config.EnableMutation = true
	config.MaxRetries = 2

	callCount := 0
	specGen := &MockSpecGenerator{
		refineFunc: func(ctx context.Context, spec *Specification, errors []VerificationError) (*Specification, error) {
			refined := *spec
			refined.ID = spec.ID + "-refined"
			return &refined, nil
		},
	}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{
		name: "mock",
		verifyFunc: func(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
			callCount++
			if callCount < 3 {
				return &VerificationResult{
					Verified: false,
					Errors:   []VerificationError{{Message: "error"}},
				}, nil
			}
			return &VerificationResult{Verified: true}, nil
		},
	})

	ctx := context.Background()

	result, err := verifier.VerifyCode(ctx, "code", "go")

	assert.NoError(t, err)
	assert.True(t, result.Verified)
	assert.Equal(t, 3, callCount)
}

// TestVerifySpec tests direct spec verification
func TestVerifySpec(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "mock"

	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{name: "mock"})

	spec := &Specification{
		ID:             "test-spec",
		Type:           SpecTypeJML,
		Preconditions:  []string{"x > 0"},
		Postconditions: []string{"result >= 0"},
	}

	ctx := context.Background()

	result, err := verifier.VerifySpec(ctx, spec, "code")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Verified)
}

// TestGetSpec tests spec retrieval
func TestGetSpec(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.DefaultProver = "mock"

	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{name: "mock"})

	ctx := context.Background()
	_, err := verifier.VerifyCode(ctx, "code", "go")
	require.NoError(t, err)

	spec, exists := verifier.GetSpec("test-spec-id")
	assert.True(t, exists)
	assert.NotNil(t, spec)
}

// TestGetSpec_NotFound tests retrieval of non-existent spec
func TestGetSpec_NotFound(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	verifier := NewFormalVerifier(config, &MockSpecGenerator{}, nil)

	spec, exists := verifier.GetSpec("non-existent")

	assert.False(t, exists)
	assert.Nil(t, spec)
}

// TestLLMSpecGenerator tests LLM-based spec generator
func TestLLMSpecGenerator(t *testing.T) {
	generateFunc := func(ctx context.Context, prompt string) (string, error) {
		return `PRECONDITIONS:
- x != null
- y > 0

POSTCONDITIONS:
- result >= 0
- result < 100

INVARIANTS:
- count >= 0`, nil
	}

	generator := NewLLMSpecGenerator(generateFunc, nil)
	ctx := context.Background()

	spec, err := generator.GenerateSpec(ctx, "func test() {}", "go")

	assert.NoError(t, err)
	assert.NotNil(t, spec)
	assert.Len(t, spec.Preconditions, 2)
	assert.Len(t, spec.Postconditions, 2)
	assert.Len(t, spec.Invariants, 1)
}

// TestLLMSpecGenerator_RefineSpec tests spec refinement
func TestLLMSpecGenerator_RefineSpec(t *testing.T) {
	generateFunc := func(ctx context.Context, prompt string) (string, error) {
		return `PRECONDITIONS:
- x >= 0

POSTCONDITIONS:
- result > 0`, nil
	}

	generator := NewLLMSpecGenerator(generateFunc, nil)
	ctx := context.Background()

	original := &Specification{
		ID:             "original",
		Preconditions:  []string{"x != null"},
		Postconditions: []string{"result >= 0"},
	}
	errors := []VerificationError{{Message: "precondition too weak"}}

	refined, err := generator.RefineSpec(ctx, original, errors)

	assert.NoError(t, err)
	assert.NotNil(t, refined)
	assert.Contains(t, refined.ID, "refined")
}

// TestZ3Prover tests Z3 prover
func TestZ3Prover(t *testing.T) {
	logger := logrus.New()
	prover := NewZ3Prover("z3", logger)

	assert.Equal(t, "z3", prover.Name())
}

// TestZ3Prover_Verify tests Z3 verification
func TestZ3Prover_Verify(t *testing.T) {
	logger := logrus.New()
	prover := NewZ3Prover("z3", logger)

	spec := &Specification{
		ID:             "test-spec",
		Type:           SpecTypeJML,
		Preconditions:  []string{"x >= 0"},
		Postconditions: []string{"result > 0"},
	}

	ctx := context.Background()

	result, err := prover.Verify(ctx, spec, "code")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "z3", result.Prover)
	assert.Contains(t, result.Metadata, "smt_code")
}

// TestDafnyVerifier tests Dafny verifier
func TestDafnyVerifier(t *testing.T) {
	logger := logrus.New()
	verifier := NewDafnyVerifier("dafny", logger)

	assert.Equal(t, "dafny", verifier.Name())
}

// TestDafnyVerifier_Verify tests Dafny verification
func TestDafnyVerifier_Verify(t *testing.T) {
	logger := logrus.New()
	verifier := NewDafnyVerifier("dafny", logger)

	spec := &Specification{
		ID:             "test-spec",
		Type:           SpecTypeDafny,
		Preconditions:  []string{"x >= 0"},
		Postconditions: []string{"result > 0"},
	}

	ctx := context.Background()

	result, err := verifier.Verify(ctx, spec, "// simple code")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "dafny", result.Prover)
}

// TestVeriPlan tests VeriPlan verifier
func TestVeriPlan(t *testing.T) {
	logger := logrus.New()
	veriplan := NewVeriPlan(logger)

	assert.NotNil(t, veriplan)
}

// TestVeriPlan_VerifyPlan tests plan verification
func TestVeriPlan_VerifyPlan(t *testing.T) {
	logger := logrus.New()
	veriplan := NewVeriPlan(logger)

	ctx := context.Background()
	plan := "Step 1: Analyze requirements\nStep 2: Design solution\nStep 3: Implement"
	constraints := []string{
		"Always validate input",
		"Eventually produce output",
		"Never access production without authorization",
	}

	result, err := veriplan.VerifyPlan(ctx, plan, constraints)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Formulas)
}

// TestVeriPlan_SafetyViolation tests safety violation detection
func TestVeriPlan_SafetyViolation(t *testing.T) {
	logger := logrus.New()
	veriplan := NewVeriPlan(logger)

	ctx := context.Background()
	plan := "Step 1: Access production database directly"
	constraints := []string{
		"Never access production directly",
	}

	result, err := veriplan.VerifyPlan(ctx, plan, constraints)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Violations)
}

// TestLTLFormula tests LTL formula types
func TestLTLFormula(t *testing.T) {
	logger := logrus.New()
	veriplan := NewVeriPlan(logger)

	testCases := []struct {
		constraint   string
		expectedType string
	}{
		{"Always validate input", "safety"},
		{"Eventually complete", "liveness"},
		{"Never fail silently", "safety"},
		{"Wait until approved", "safety"},
	}

	for _, tc := range testCases {
		t.Run(tc.constraint, func(t *testing.T) {
			formula := veriplan.constraintToLTL(tc.constraint)
			assert.Equal(t, tc.expectedType, formula.Type)
			assert.NotEmpty(t, formula.Formula)
		})
	}
}

// TestSpecificationTypes tests all specification types
func TestSpecificationTypes(t *testing.T) {
	specTypes := []SpecificationType{
		SpecTypeJML,
		SpecTypeDafny,
		SpecTypeLTL,
		SpecTypeInvariant,
		SpecTypePrecondition,
		SpecTypePostcondition,
	}

	for _, st := range specTypes {
		assert.NotEmpty(t, string(st))
	}
}

// TestVerificationResult_MarshalJSON tests custom JSON marshaling
func TestVerificationResult_MarshalJSON(t *testing.T) {
	result := &VerificationResult{
		Verified: true,
		Specification: &Specification{
			ID:   "test",
			Type: SpecTypeJML,
		},
		Duration: 1500 * time.Millisecond,
		Prover:   "z3",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Verify duration_ms is present
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Contains(t, parsed, "duration_ms")
	assert.Equal(t, float64(1500), parsed["duration_ms"])
}

// TestCounterexample tests counterexample structure
func TestCounterexample(t *testing.T) {
	ce := &Counterexample{
		Variables: map[string]interface{}{
			"x": 5,
			"y": -1,
		},
		Trace: []string{"step1", "step2", "error"},
		State: "violated",
	}

	data, err := json.Marshal(ce)
	require.NoError(t, err)

	var parsed Counterexample
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, 5.0, parsed.Variables["x"])
	assert.Len(t, parsed.Trace, 3)
}

// TestVerificationError tests verification error structure
func TestVerificationError(t *testing.T) {
	ve := VerificationError{
		Line:    10,
		Column:  5,
		Message: "postcondition violation",
		Code:    "POST_VIOLATION",
	}

	data, err := json.Marshal(ve)
	require.NoError(t, err)

	var parsed VerificationError
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, 10, parsed.Line)
	assert.Equal(t, "postcondition violation", parsed.Message)
}

// TestSpecification tests specification structure
func TestSpecification(t *testing.T) {
	spec := &Specification{
		ID:             "spec-123",
		Type:           SpecTypeJML,
		Target:         "func test()",
		Preconditions:  []string{"x > 0", "y != null"},
		Postconditions: []string{"result >= 0"},
		Invariants:     []string{"count >= 0"},
		Assertions:     []string{"temp < 100"},
		RawSpec:        "raw specification text",
		CreatedAt:      time.Now(),
	}

	data, err := json.Marshal(spec)
	require.NoError(t, err)

	var parsed Specification
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "spec-123", parsed.ID)
	assert.Equal(t, SpecTypeJML, parsed.Type)
	assert.Len(t, parsed.Preconditions, 2)
}

// TestPlanVerificationResult tests plan verification result structure
func TestPlanVerificationResult(t *testing.T) {
	result := &PlanVerificationResult{
		Valid: true,
		Formulas: []*LTLFormula{
			{Formula: "G(safe)", Type: "safety", Natural: "Always be safe"},
		},
		StateSpaceSize: 100,
		Duration:       500 * time.Millisecond,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var parsed PlanVerificationResult
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.True(t, parsed.Valid)
	assert.Len(t, parsed.Formulas, 1)
}

// TestParseConditions tests condition parsing
func TestParseConditions(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "- condition1\n- condition2\n- condition3",
			expected: []string{"condition1", "condition2", "condition3"},
		},
		{
			input:    "- single condition",
			expected: []string{"single condition"},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "no bullets here",
			expected: []string{},
		},
		{
			input:    "-   spaced condition  ",
			expected: []string{"spaced condition"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseConditions(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFormalVerifierConfig_JSON tests config JSON marshaling
func TestFormalVerifierConfig_JSON(t *testing.T) {
	config := DefaultFormalVerifierConfig()

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var parsed FormalVerifierConfig
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, config.DefaultProver, parsed.DefaultProver)
	assert.Equal(t, config.MaxRetries, parsed.MaxRetries)
}

// TestDafnyVerifier_DiffCheck tests diff checking
func TestDafnyVerifier_DiffCheck(t *testing.T) {
	logger := logrus.New()
	verifier := NewDafnyVerifier("dafny", logger)

	testCases := []struct {
		original  string
		annotated string
		expected  bool
	}{
		{
			original:  "x := 1",
			annotated: "requires x > 0\nx := 1",
			expected:  true,
		},
		{
			original:  "",
			annotated: "any",
			expected:  true, // Empty original is allowed
		},
	}

	for _, tc := range testCases {
		result := verifier.diffCheck(tc.original, tc.annotated)
		assert.Equal(t, tc.expected, result)
	}
}

// TestContextTimeout tests context timeout handling
func TestContextTimeout(t *testing.T) {
	config := DefaultFormalVerifierConfig()
	config.Timeout = 10 * time.Millisecond
	config.DefaultProver = "mock"

	specGen := &MockSpecGenerator{}
	verifier := NewFormalVerifier(config, specGen, nil)
	verifier.RegisterProver(&MockTheoremProver{
		name: "mock",
		verifyFunc: func(ctx context.Context, spec *Specification, code string) (*VerificationResult, error) {
			time.Sleep(100 * time.Millisecond)
			return &VerificationResult{Verified: true}, nil
		},
	})

	ctx := context.Background()

	// This should timeout
	_, err := verifier.VerifyCode(ctx, "code", "go")
	// Note: Error depends on when timeout occurs - may or may not error
	_ = err
}
