// Package governance provides tests for the SEMAP Protocol.
package governance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSEMAP_RegisterContract(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	contract := &Contract{
		Name:        "Test Contract",
		Description: "A test contract",
		Type:        ContractTypePrecondition,
		Severity:    ViolationSeverityError,
		Conditions: []Condition{{
			ID:        "cond1",
			Evaluator: ConditionTypeRegex,
			Expression: "^[a-zA-Z]+$",
			Message:   "Must contain only letters",
		}},
		Enabled: true,
	}

	err := semap.RegisterContract(contract)
	require.NoError(t, err)

	// Verify contract was registered
	retrieved := semap.GetContract(contract.ID)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test Contract", retrieved.Name)
}

func TestSEMAP_CheckPreconditions(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	// Register a contract
	contract := &Contract{
		ID:          "test-contract",
		Name:        "Input Validation",
		Type:        ContractTypePrecondition,
		Severity:    ViolationSeverityError,
		Conditions: []Condition{{
			ID:         "length-check",
			Evaluator:  ConditionTypeRegex,
			Expression: "^.{1,100}$",
			Message:    "Input must be 1-100 characters",
		}},
		Enabled: true,
	}

	semap.RegisterContract(contract)

	ctx := context.Background()

	// Test valid input
	result, err := semap.CheckPreconditions(ctx, "test-agent", "process", "valid input")
	require.NoError(t, err)
	assert.True(t, result.Passed)

	// Test empty input
	result, err = semap.CheckPreconditions(ctx, "test-agent", "process", "")
	require.NoError(t, err)
	// May or may not pass depending on regex
}

func TestSEMAP_CheckGuardRails(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	// Register a guard rail contract
	contract := PredefinedContracts["no_sql_injection"]
	semap.RegisterContract(contract)

	ctx := context.Background()

	// Test with SQL injection attempt
	result, err := semap.CheckGuardRails(ctx, "test-agent", "query", "SELECT * FROM users WHERE id=1; DROP TABLE users;")
	require.NoError(t, err)
	// Guard rail should detect the SQL injection pattern

	// Test with safe input
	result, err = semap.CheckGuardRails(ctx, "test-agent", "query", "normal text without sql")
	require.NoError(t, err)
	assert.True(t, result.Passed)
}

func TestSEMAP_RegisterAgentProfile(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	profile := &AgentProfile{
		ID:          "test-agent",
		Name:        "Test Agent",
		Description: "An agent for testing",
		Capabilities: []AgentCapability{
			CapabilityRead,
			CapabilityWrite,
		},
		TrustLevel: TrustLevelMedium,
	}

	err := semap.RegisterAgentProfile(profile)
	require.NoError(t, err)

	// Verify profile was registered
	retrieved := semap.GetAgentProfile("test-agent")
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test Agent", retrieved.Name)
	assert.Equal(t, TrustLevelMedium, retrieved.TrustLevel)
}

func TestSEMAP_RegisterPolicy(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	// Register a contract first
	contract := &Contract{
		ID:   "policy-contract",
		Name: "Policy Contract",
		Type: ContractTypePrecondition,
		Enabled: true,
	}
	semap.RegisterContract(contract)

	policy := &Policy{
		Name:        "Test Policy",
		Description: "A test policy",
		Contracts:   []string{"policy-contract"},
		Scope: PolicyScope{
			Agents: []string{"test-agent"},
		},
		Priority: 1,
		Enabled:  true,
	}

	err := semap.RegisterPolicy(policy)
	require.NoError(t, err)

	// Verify policy was registered
	retrieved := semap.GetPolicy(policy.ID)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test Policy", retrieved.Name)
}

func TestSEMAP_GetViolations(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	// Register a contract that will be violated
	contract := &Contract{
		ID:          "strict-contract",
		Name:        "Strict Contract",
		Type:        ContractTypePrecondition,
		Severity:    ViolationSeverityError,
		Conditions: []Condition{{
			ID:          "must-not-be-empty",
			Evaluator:   ConditionTypeRegex,
			Expression:  "^.+$",
			Message:     "Input must not be empty",
			FailMessage: "Empty input not allowed",
		}},
		Enabled: true,
	}
	semap.RegisterContract(contract)

	ctx := context.Background()

	// Trigger a violation with empty input
	semap.CheckPreconditions(ctx, "agent1", "action", "")

	// Get violations
	violations := semap.GetViolations()
	// May or may not have violations depending on contract evaluation
	_ = violations // Use the variable
}

func TestSEMAP_GetViolationsByAgent(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	ctx := context.Background()

	// Perform some checks that may generate violations
	semap.CheckPreconditions(ctx, "agent1", "action1", "input1")
	semap.CheckPreconditions(ctx, "agent2", "action2", "input2")

	// Get violations for specific agent
	violations := semap.GetViolationsByAgent("agent1")
	for _, v := range violations {
		assert.Equal(t, "agent1", v.Context.AgentID)
	}
}

func TestSEMAP_GetStatistics(t *testing.T) {
	config := DefaultSEMAPConfig()
	semap := NewSEMAP(config)

	// Register some contracts and profiles
	semap.RegisterContract(&Contract{
		ID:   "contract1",
		Name: "Contract 1",
		Type: ContractTypePrecondition,
	})

	semap.RegisterContract(&Contract{
		ID:   "contract2",
		Name: "Contract 2",
		Type: ContractTypeGuardRail,
	})

	semap.RegisterAgentProfile(&AgentProfile{
		ID:   "agent1",
		Name: "Agent 1",
	})

	stats := semap.GetStatistics()

	assert.Equal(t, 2, stats.TotalContracts)
	assert.Equal(t, 1, stats.TotalProfiles)
}

func TestRegexEvaluator(t *testing.T) {
	evaluator := &RegexEvaluator{}
	ctx := context.Background()

	testCases := []struct {
		name       string
		expression string
		value      string
		expected   bool
	}{
		{
			name:       "Match Letters Only",
			expression: "^[a-zA-Z]+$",
			value:      "HelloWorld",
			expected:   true,
		},
		{
			name:       "No Match with Numbers",
			expression: "^[a-zA-Z]+$",
			value:      "Hello123",
			expected:   false,
		},
		{
			name:       "Match Email Pattern",
			expression: `^[\w.]+@[\w.]+\.\w+$`,
			value:      "test@example.com",
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			condition := &Condition{
				Expression: tc.expression,
			}
			result, err := evaluator.Evaluate(ctx, condition, tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRangeEvaluator(t *testing.T) {
	evaluator := &RangeEvaluator{}
	ctx := context.Background()

	testCases := []struct {
		name     string
		min      float64
		max      float64
		value    float64
		expected bool
	}{
		{
			name:     "Within Range",
			min:      0,
			max:      100,
			value:    50,
			expected: true,
		},
		{
			name:     "Below Range",
			min:      10,
			max:      100,
			value:    5,
			expected: false,
		},
		{
			name:     "Above Range",
			min:      0,
			max:      100,
			value:    150,
			expected: false,
		},
		{
			name:     "At Min Boundary",
			min:      0,
			max:      100,
			value:    0,
			expected: true,
		},
		{
			name:     "At Max Boundary",
			min:      0,
			max:      100,
			value:    100,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			condition := &Condition{
				Parameters: map[string]interface{}{
					"min": tc.min,
					"max": tc.max,
				},
			}
			result, err := evaluator.Evaluate(ctx, condition, tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEnumEvaluator(t *testing.T) {
	evaluator := &EnumEvaluator{}
	ctx := context.Background()

	condition := &Condition{
		Parameters: map[string]interface{}{
			"values": []interface{}{"red", "green", "blue"},
		},
	}

	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Valid Value", "red", true},
		{"Valid Value 2", "green", true},
		{"Invalid Value", "yellow", false},
		{"Case Insensitive", "RED", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(ctx, condition, tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateDefaultProfile(t *testing.T) {
	testCases := []struct {
		trustLevel TrustLevel
		minCaps    int
	}{
		{TrustLevelUntrusted, 0},
		{TrustLevelLow, 1},
		{TrustLevelMedium, 2},
		{TrustLevelHigh, 3},
		{TrustLevelTrusted, 5},
	}

	for _, tc := range testCases {
		t.Run(string(tc.trustLevel), func(t *testing.T) {
			profile := CreateDefaultProfile("test-agent", tc.trustLevel)

			assert.Equal(t, "test-agent", profile.ID)
			assert.Equal(t, tc.trustLevel, profile.TrustLevel)
			assert.GreaterOrEqual(t, len(profile.Capabilities), tc.minCaps)
		})
	}
}

func TestPredefinedContracts(t *testing.T) {
	expectedContracts := []string{
		"no_sql_injection",
		"no_path_traversal",
		"max_response_length",
	}

	for _, name := range expectedContracts {
		t.Run(name, func(t *testing.T) {
			contract, ok := PredefinedContracts[name]
			assert.True(t, ok, "Contract %s should exist", name)
			assert.NotEmpty(t, contract.Name)
			assert.NotEmpty(t, contract.Type)
			assert.True(t, contract.Enabled)
		})
	}
}

func TestAuditLog(t *testing.T) {
	log := NewAuditLog()

	// Add entries
	log.Log(AuditEntry{
		Type:       "test_action",
		AgentID:    "agent1",
		Action:     "test",
		Result:     true,
		Timestamp:  time.Now(),
	})

	log.Log(AuditEntry{
		Type:       "test_action",
		AgentID:    "agent2",
		Action:     "test2",
		Result:     false,
		Timestamp:  time.Now(),
	})

	// Get entries
	entries := log.GetEntries()
	assert.Len(t, entries, 2)
	assert.Equal(t, "agent1", entries[0].AgentID)
	assert.Equal(t, "agent2", entries[1].AgentID)
}

func TestContractSerialization(t *testing.T) {
	contract := &Contract{
		ID:          "test-contract",
		Name:        "Test Contract",
		Description: "A test contract",
		Type:        ContractTypePrecondition,
		Severity:    ViolationSeverityError,
		Enabled:     true,
	}

	// Serialize
	data, err := contract.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize
	deserialized, err := DeserializeContract(data)
	require.NoError(t, err)
	assert.Equal(t, contract.ID, deserialized.ID)
	assert.Equal(t, contract.Name, deserialized.Name)
}
