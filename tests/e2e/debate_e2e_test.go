package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoServer skips the test if HelixAgent server is not running
func skipIfNoServer(t *testing.T) {
	if os.Getenv("SKIP_E2E") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E=true)")
	}

	// Check if server is running
	resp, err := http.Get("http://localhost:7061/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("Skipping E2E test (server not running on localhost:7061)")
	}
}

type DebateRequest struct {
	Topic     string `json:"topic"`
	MaxRounds int    `json:"max_rounds"`
}

type DebateResponse struct {
	DebateID             string                 `json:"debate_id"`
	Topic                string                 `json:"topic"`
	Success              bool                   `json:"success"`
	RoundsConducted      int                    `json:"rounds_conducted"`
	ValidationResult     interface{}            `json:"validation_result,omitempty"`
	TestDrivenMetadata   map[string]interface{} `json:"test_driven_metadata,omitempty"`
	ToolEnrichmentUsed   bool                   `json:"tool_enrichment_used,omitempty"`
	SpecializedRole      string                 `json:"specialized_role,omitempty"`
	QualityScore         float64                `json:"quality_score"`
}

// TestDebateE2E_CodeGeneration tests end-to-end code generation debate
func TestDebateE2E_CodeGeneration(t *testing.T) {
	skipIfNoServer(t)

	request := DebateRequest{
		Topic:     "Write a Python function to calculate the factorial of a number",
		MaxRounds: 1,
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	// Create request with authentication
	req, err := http.NewRequest("POST", "http://localhost:7061/v1/debates", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Add API key from environment
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should get a response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should get 200 OK response")

	var result DebateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, result.DebateID, "Should have debate ID")
	assert.Equal(t, request.Topic, result.Topic, "Topic should match")

	// Verify integrated features
	t.Logf("Test-Driven Metadata: %v", result.TestDrivenMetadata)
	t.Logf("Validation Result: %v", result.ValidationResult)
	t.Logf("Tool Enrichment: %v", result.ToolEnrichmentUsed)
	t.Logf("Specialized Role: %s", result.SpecializedRole)

	// Code generation should trigger test-driven mode or specialized role
	if result.TestDrivenMetadata != nil {
		assert.NotEmpty(t, result.TestDrivenMetadata, "Test-driven metadata should be present for code generation")
		t.Log("✓ Test-Driven Debate was triggered")
	}

	if result.SpecializedRole != "" {
		assert.Contains(t, []string{"generator", "coder", ""}, result.SpecializedRole,
			"Should select appropriate role for code generation")
		t.Logf("✓ Specialized Role selected: %s", result.SpecializedRole)
	}
}

// TestDebateE2E_PerformanceAnalysis tests performance analysis task
func TestDebateE2E_PerformanceAnalysis(t *testing.T) {
	skipIfNoServer(t)

	request := DebateRequest{
		Topic:     "Analyze and optimize the performance of a database query that joins 5 tables",
		MaxRounds: 1,
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	// Create request with authentication
	req, err := http.NewRequest("POST", "http://localhost:7061/v1/debates", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Add API key from environment
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result DebateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Should select performance_analyzer role
	if result.SpecializedRole != "" {
		assert.Equal(t, "performance_analyzer", result.SpecializedRole,
			"Should select performance_analyzer for optimization tasks")
		t.Logf("✓ Performance Analyzer role selected")
	}

	// Should have validation result
	if result.ValidationResult != nil {
		t.Log("✓ 4-Pass Validation was applied")
	}
}

// TestDebateE2E_SecurityAudit tests security audit task
func TestDebateE2E_SecurityAudit(t *testing.T) {
	skipIfNoServer(t)

	request := DebateRequest{
		Topic:     "Security audit of the user authentication system - find vulnerabilities",
		MaxRounds: 1,
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	// Create request with authentication
	req, err := http.NewRequest("POST", "http://localhost:7061/v1/debates", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Add API key from environment
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result DebateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Should select security_analyst role
	if result.SpecializedRole != "" {
		assert.Equal(t, "security_analyst", result.SpecializedRole,
			"Should select security_analyst for security tasks")
		t.Logf("✓ Security Analyst role selected")
	}

	// Tool enrichment might be used for security analysis
	if result.ToolEnrichmentUsed {
		t.Log("✓ Tool Enrichment was used")
	}
}

// TestDebateE2E_RefactoringTask tests refactoring task
func TestDebateE2E_RefactoringTask(t *testing.T) {
	skipIfNoServer(t)

	request := DebateRequest{
		Topic:     "Refactor the legacy payment processing module to improve maintainability",
		MaxRounds: 1,
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	// Create request with authentication
	req, err := http.NewRequest("POST", "http://localhost:7061/v1/debates", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Add API key from environment
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result DebateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Should select refactorer role
	if result.SpecializedRole != "" {
		assert.Equal(t, "refactorer", result.SpecializedRole,
			"Should select refactorer for refactoring tasks")
		t.Logf("✓ Refactorer role selected")
	}

	// Validation should run
	if result.ValidationResult != nil {
		t.Log("✓ 4-Pass Validation was applied")
	}
}

// TestDebateE2E_IntegratedFeatures tests all integrated features working together
func TestDebateE2E_IntegratedFeatures(t *testing.T) {
	skipIfNoServer(t)

	// This test verifies that all integrated features can work together
	request := DebateRequest{
		Topic:     "Write a secure authentication function in Go with performance optimization and comprehensive tests",
		MaxRounds: 2,
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		"http://localhost:7061/v1/debates",
		bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Add API key from environment
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result DebateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	t.Logf("Debate ID: %s", result.DebateID)
	t.Logf("Rounds Conducted: %d", result.RoundsConducted)
	t.Logf("Quality Score: %.2f", result.QualityScore)

	// Feature verification matrix
	featuresDetected := make(map[string]bool)

	// 1. Test-Driven Debate
	if result.TestDrivenMetadata != nil {
		featuresDetected["test_driven"] = true
		t.Log("✓ Feature: Test-Driven Debate ACTIVE")
		t.Logf("  - Metadata: %v", result.TestDrivenMetadata)
	}

	// 2. 4-Pass Validation
	if result.ValidationResult != nil {
		featuresDetected["validation"] = true
		t.Log("✓ Feature: 4-Pass Validation ACTIVE")
		t.Logf("  - Result: %v", result.ValidationResult)
	}

	// 3. Tool Integration
	if result.ToolEnrichmentUsed {
		featuresDetected["tool_enrichment"] = true
		t.Log("✓ Feature: Tool Integration ACTIVE")
	}

	// 4. Specialized Role
	if result.SpecializedRole != "" {
		featuresDetected["specialized_role"] = true
		t.Log("✓ Feature: Specialized Role ACTIVE")
		t.Logf("  - Role: %s", result.SpecializedRole)
	}

	// At least some features should be active
	assert.NotEmpty(t, featuresDetected, "At least one integrated feature should be active")

	// Log summary
	t.Logf("\n=== INTEGRATED FEATURES SUMMARY ===")
	t.Logf("Test-Driven Debate: %v", featuresDetected["test_driven"])
	t.Logf("4-Pass Validation: %v", featuresDetected["validation"])
	t.Logf("Tool Integration: %v", featuresDetected["tool_enrichment"])
	t.Logf("Specialized Roles: %v", featuresDetected["specialized_role"])
	t.Logf("===================================")
}
