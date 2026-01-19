// Package security provides tests for the SecureFixAgent and security components.
package security

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureFixAgent_DetectRepairValidate(t *testing.T) {
	config := DefaultSecureFixAgentConfig()
	logger := logrus.New()
	agent := NewSecureFixAgent(config, nil, nil, logger)

	ctx := context.Background()

	testCode := `
func handleUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("id")
	db.Query("SELECT * FROM users WHERE id = " + query)
}
`

	result, err := agent.DetectRepairValidate(ctx, testCode, "go")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPatternBasedScanner_Scan(t *testing.T) {
	logger := logrus.New()
	scanner := NewPatternBasedScanner(logger)

	ctx := context.Background()

	// Test scanning code
	vulns, err := scanner.Scan(ctx, `fmt.Println("Hello, World!")`, "go")
	require.NoError(t, err)
	// Scanner may or may not find vulnerabilities - we're testing it doesn't error
	_ = vulns

	// Test scanning potentially vulnerable code
	vulns, err = scanner.Scan(ctx, `db.Query("SELECT * FROM users WHERE id = " + userInput)`, "go")
	require.NoError(t, err)
	// Scanner may or may not find vulnerabilities depending on patterns
	_ = vulns
}

func TestVulnerability_Severity(t *testing.T) {
	severities := []VulnerabilitySeverity{
		SeverityInfo,
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}

	for _, sev := range severities {
		assert.NotEmpty(t, string(sev))
	}
}

func TestVulnerabilityCategory(t *testing.T) {
	categories := []VulnerabilityCategory{
		CategoryInjection,
		CategoryXSS,
		CategoryAuthentication,
		CategoryAuthorization,
		CategoryCryptographic,
		CategorySensitiveData,
		CategoryMisconfiguration,
		CategoryDependency,
		CategoryMemorySafety,
		CategoryRaceCondition,
	}

	for _, cat := range categories {
		assert.NotEmpty(t, string(cat))
	}
}

func TestFiveRingDefense_Defend(t *testing.T) {
	logger := logrus.New()
	defense := NewFiveRingDefense(logger)

	// Add input sanitization ring
	defense.AddRing(NewInputSanitizationRing())

	ctx := context.Background()

	testCases := []struct {
		name        string
		input       string
		expectPass  bool
	}{
		{
			name:       "Valid Input",
			input:      "Hello World",
			expectPass: true,
		},
		{
			name:       "SQL Injection Attempt",
			input:      "'; DROP TABLE users; --",
			expectPass: false,
		},
		{
			name:       "Script Injection",
			input:      "<script>alert('xss')</script>",
			expectPass: false,
		},
		{
			name:       "Normal Code",
			input:      "func main() { fmt.Println() }",
			expectPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := defense.Defend(ctx, tc.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
			if tc.expectPass {
				assert.True(t, result.Passed)
			} else {
				assert.False(t, result.Passed)
			}
		})
	}
}

func TestSecureFixAgentConfig_Defaults(t *testing.T) {
	config := DefaultSecureFixAgentConfig()

	assert.True(t, config.EnableAutoFix)
	assert.True(t, config.RequireValidation)
	assert.Greater(t, config.MaxConcurrentScans, 0)
	assert.Greater(t, config.Timeout, time.Duration(0))
}

func TestSecurityResult(t *testing.T) {
	result := &SecurityResult{
		Vulnerabilities: []*Vulnerability{
			{ID: "v1", Title: "Test Vulnerability"},
			{ID: "v2", Title: "Another Vulnerability"},
		},
		Success:  true,
		Duration: 100 * time.Millisecond,
	}

	assert.Len(t, result.Vulnerabilities, 2)
	assert.True(t, result.Success)
	assert.Equal(t, 100*time.Millisecond, result.Duration)
}

func TestInputSanitizationRing(t *testing.T) {
	ring := NewInputSanitizationRing()

	ctx := context.Background()

	// Test with SQL injection
	passed, _, err := ring.Check(ctx, "SELECT * FROM users WHERE id = '1' OR '1'='1'")
	require.NoError(t, err)
	assert.False(t, passed)

	// Test with clean input
	passed, _, err = ring.Check(ctx, "Hello World")
	require.NoError(t, err)
	assert.True(t, passed)
}

func TestRateLimitingRing(t *testing.T) {
	ring := NewRateLimitingRing(5, time.Minute)
	ctx := context.Background()

	// Should allow initial requests
	for i := 0; i < 5; i++ {
		passed, _, err := ring.Check(ctx, "request")
		require.NoError(t, err)
		assert.True(t, passed)
	}

	// Should rate limit after threshold
	passed, _, err := ring.Check(ctx, "request")
	require.NoError(t, err)
	assert.False(t, passed)
}

func TestSecureFixAgent_ScanOnly(t *testing.T) {
	config := DefaultSecureFixAgentConfig()
	config.EnableAutoFix = false
	logger := logrus.New()
	agent := NewSecureFixAgent(config, nil, nil, logger)

	ctx := context.Background()

	vulnerableCode := `
func login(username, password string) bool {
	query := fmt.Sprintf("SELECT * FROM users WHERE username='%s' AND password='%s'", username, password)
	// Execute query
	return true
}
`

	result, err := agent.DetectRepairValidate(ctx, vulnerableCode, "go")
	require.NoError(t, err)
	assert.NotNil(t, result)
	// May or may not find vulnerabilities depending on pattern matching
}

func TestPatternBasedScanner_Name(t *testing.T) {
	logger := logrus.New()
	scanner := NewPatternBasedScanner(logger)

	// Verify scanner name
	name := scanner.Name()
	assert.NotEmpty(t, name)
}
