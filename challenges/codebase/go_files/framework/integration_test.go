package framework

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockChallenge is a test challenge for integration testing.
type MockChallenge struct {
	BaseChallenge
	shouldFail   bool
	sleepTime    time.Duration
	customResult *ChallengeResult
}

func NewMockChallenge(id ChallengeID, name string, deps []ChallengeID) *MockChallenge {
	return &MockChallenge{
		BaseChallenge: BaseChallenge{
			id:           id,
			name:         name,
			description:  "Mock challenge for testing",
			dependencies: deps,
		},
	}
}

func (m *MockChallenge) Execute(ctx context.Context) (*ChallengeResult, error) {
	if m.sleepTime > 0 {
		time.Sleep(m.sleepTime)
	}

	if m.customResult != nil {
		return m.customResult, nil
	}

	result := &ChallengeResult{
		ChallengeID:   m.id,
		ChallengeName: m.name,
		StartTime:     time.Now(),
		EndTime:       time.Now(),
		Duration:      m.sleepTime,
	}

	if m.shouldFail {
		result.Status = StatusFailed
		result.Error = "Mock failure"
	} else {
		result.Status = StatusPassed
		result.Assertions = []AssertionResult{
			{Type: "not_empty", Target: "response", Passed: true, Message: "Response is not empty"},
			{Type: "quality_score", Target: "response", Passed: true, Message: "Quality met"},
		}
		result.Metrics = map[string]MetricValue{
			"response_time": {Name: "Response Time", Value: 100.0, Unit: "ms"},
		}
	}

	return result, nil
}

func TestIntegration_FullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create registry with challenges
	registry := NewRegistry()

	ch1 := NewMockChallenge("provider_verification", "Provider Verification", nil)
	ch2 := NewMockChallenge("ai_debate_formation", "AI Debate Formation", []ChallengeID{"provider_verification"})
	ch3 := NewMockChallenge("api_quality_test", "API Quality Test", []ChallengeID{"ai_debate_formation"})

	registry.Register(ch1)
	registry.Register(ch2)
	registry.Register(ch3)

	// 2. Get dependency order
	order, err := registry.GetDependencyOrder()
	if err != nil {
		t.Fatalf("GetDependencyOrder failed: %v", err)
	}

	if len(order) != 3 {
		t.Errorf("Expected 3 challenges in order, got %d", len(order))
	}

	// Verify order by checking IDs
	if order[0].ID() != "provider_verification" {
		t.Errorf("First challenge should be provider_verification, got %s", order[0].ID())
	}
	if order[1].ID() != "ai_debate_formation" {
		t.Errorf("Second challenge should be ai_debate_formation, got %s", order[1].ID())
	}
	if order[2].ID() != "api_quality_test" {
		t.Errorf("Third challenge should be api_quality_test, got %s", order[2].ID())
	}

	// 3. Execute challenges and collect results
	var results []*ChallengeResult
	for _, ch := range order {
		result, err := ch.Execute(context.Background())
		if err != nil {
			t.Fatalf("Challenge %s execution failed: %v", ch.ID(), err)
		}
		results = append(results, result)
	}

	// 4. Setup logging
	logger, err := SetupLogging(tmpDir, true)
	if err != nil {
		t.Fatalf("SetupLogging failed: %v", err)
	}

	for _, result := range results {
		logger.Info("Challenge completed",
			Field{Key: "challenge_id", Value: result.ChallengeID},
			Field{Key: "status", Value: result.Status},
		)
	}
	logger.Close()

	// 5. Generate reports
	mdReporter := NewMarkdownReporter(tmpDir)
	jsonReporter := NewJSONReporter(tmpDir, true)

	for i, result := range results {
		mdFilename := string(result.ChallengeID) + "_report.md"
		jsonFilename := string(result.ChallengeID) + "_report.json"

		if err := mdReporter.SaveReport(result, mdFilename); err != nil {
			t.Errorf("SaveReport (md) for challenge %d failed: %v", i, err)
		}

		if err := jsonReporter.SaveReport(result, jsonFilename); err != nil {
			t.Errorf("SaveReport (json) for challenge %d failed: %v", i, err)
		}
	}

	// 6. Generate master summary
	summary := BuildMasterSummary(results)
	if err := SaveMasterSummary(summary, tmpDir); err != nil {
		t.Fatalf("SaveMasterSummary failed: %v", err)
	}

	// 7. Append to history
	historyPath := filepath.Join(tmpDir, "history.jsonl")
	for i, result := range results {
		resultsPath := filepath.Join(tmpDir, "run_1", string(result.ChallengeID))
		if err := AppendToHistory(historyPath, result, resultsPath); err != nil {
			t.Errorf("AppendToHistory for result %d failed: %v", i, err)
		}
	}

	// 8. Verify all files were created
	expectedFiles := []string{
		"challenge.log",
		"provider_verification_report.md",
		"provider_verification_report.json",
		"ai_debate_formation_report.md",
		"ai_debate_formation_report.json",
		"api_quality_test_report.md",
		"api_quality_test_report.json",
		"history.jsonl",
		"latest_summary.json",
		"latest_summary.md",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not found: %s", f)
		}
	}

	// 9. Verify summary content
	if summary.TotalChallenges != 3 {
		t.Errorf("Summary TotalChallenges = %d, want 3", summary.TotalChallenges)
	}
	if summary.PassedChallenges != 3 {
		t.Errorf("Summary PassedChallenges = %d, want 3", summary.PassedChallenges)
	}
	if summary.AveragePassRate != 1.0 {
		t.Errorf("Summary AveragePassRate = %f, want 1.0", summary.AveragePassRate)
	}
}

func TestIntegration_EnvAndAssertions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `# Test environment
ANTHROPIC_API_KEY=sk-ant-api-test123456789
OPENAI_API_KEY=sk-openai-test-key-12345
TEST_VALUE=hello world
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Load environment
	loader := NewEnvLoader()
	if err := loader.Load(envPath); err != nil {
		t.Fatalf("EnvLoader.Load failed: %v", err)
	}

	// Test GetAPIKey
	anthropicKey := loader.GetAPIKey("anthropic")
	if anthropicKey != "sk-ant-api-test123456789" {
		t.Errorf("Unexpected anthropic key")
	}

	// Test assertions with API response simulation
	engine := NewAssertionEngine()

	testResponse := `Here is a Go function that checks if a number is prime:

func isPrime(n int) bool {
    if n <= 1 {
        return false
    }
    for i := 2; i*i <= n; i++ {
        if n%i == 0 {
            return false
        }
    }
    return true
}

This works because we only need to check divisors up to the square root of n.
If n is divisible by any number larger than its square root, then it must also
be divisible by a smaller number, which we would have already checked.`

	assertions := []AssertionDefinition{
		{Type: "not_empty", Target: "response"},
		{Type: "not_mock", Target: "response"},
		{Type: "min_length", Target: "response", Value: 100},
		{Type: "contains", Target: "response", Value: "func"},
		{Type: "code_valid", Target: "response"},
		{Type: "reasoning_present", Target: "response"},
	}

	data := map[string]any{
		"response": testResponse,
	}

	results := engine.EvaluateAll(assertions, data)

	passedCount := 0
	for _, r := range results {
		if r.Passed {
			passedCount++
		}
	}

	if passedCount != len(assertions) {
		t.Errorf("Expected all %d assertions to pass, but only %d passed", len(assertions), passedCount)
		for _, r := range results {
			if !r.Passed {
				t.Logf("Failed assertion: %s - %s", r.Type, r.Message)
			}
		}
	}

	// Test redaction
	redacted := loader.GetAllRedacted()
	if !hasPrefix(redacted["ANTHROPIC_API_KEY"], "sk-a") {
		t.Errorf("Redacted key should start with sk-a")
	}
	if strings.Contains(redacted["ANTHROPIC_API_KEY"], "test123456789") {
		t.Error("Redacted key should not contain original value")
	}
}

func TestIntegration_LoggingAndReporting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create logger with all features
	config := LoggerConfig{
		OutputPath:     filepath.Join(tmpDir, "test.log"),
		APIRequestLog:  filepath.Join(tmpDir, "api_requests.log"),
		APIResponseLog: filepath.Join(tmpDir, "api_responses.log"),
		Level:          LevelDebug,
		Verbose:        true,
	}

	logger, err := NewJSONLogger(config)
	if err != nil {
		t.Fatalf("NewJSONLogger failed: %v", err)
	}

	// Wrap with redacting logger
	secrets := []string{"sk-ant-secret-key-12345"}
	redactingLogger := NewRedactingLogger(logger, secrets...)

	// Log messages
	redactingLogger.Info("Starting test with key sk-ant-secret-key-12345")
	redactingLogger.Debug("Debug info")

	// Log API request/response
	redactingLogger.LogAPIRequest(APIRequestLog{
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: "req-001",
		Method:    "POST",
		URL:       "https://api.anthropic.com/v1/messages",
		Headers: map[string]string{
			"Authorization": "Bearer sk-ant-secret-key-12345",
			"Content-Type":  "application/json",
		},
	})

	redactingLogger.LogAPIResponse(APIResponseLog{
		Timestamp:      time.Now().Format(time.RFC3339),
		RequestID:      "req-001",
		StatusCode:     200,
		ResponseTimeMs: 1500,
	})

	redactingLogger.Close()

	// Verify logs don't contain secrets
	logData, err := os.ReadFile(filepath.Join(tmpDir, "test.log"))
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if strings.Contains(string(logData), "sk-ant-secret-key-12345") {
		t.Error("Log file should not contain unredacted secret")
	}

	// Create challenge result for reporting
	result := &ChallengeResult{
		ChallengeID:   "test_challenge",
		ChallengeName: "Test Challenge",
		Status:        StatusPassed,
		StartTime:     time.Now().Add(-5 * time.Minute),
		EndTime:       time.Now(),
		Duration:      5 * time.Minute,
		Assertions: []AssertionResult{
			{Type: "not_empty", Passed: true, Message: "OK"},
			{Type: "not_mock", Passed: true, Message: "OK"},
		},
		Metrics: map[string]MetricValue{
			"quality": {Name: "Quality Score", Value: 0.95, Unit: ""},
		},
		Logs: LogPaths{
			ChallengeLog: filepath.Join(tmpDir, "test.log"),
			OutputLog:    filepath.Join(tmpDir, "output.log"),
		},
	}

	// Generate and verify markdown report
	mdReporter := NewMarkdownReporter(tmpDir)
	mdData, err := mdReporter.GenerateReport(result)
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	mdContent := string(mdData)
	if !strings.Contains(mdContent, "PASSED") {
		t.Error("Markdown report should contain status")
	}
	if !strings.Contains(mdContent, "Quality Score") {
		t.Error("Markdown report should contain metrics")
	}
	if !strings.Contains(mdContent, "2/2") {
		t.Error("Markdown report should show assertion pass rate")
	}

	// Generate and verify JSON report
	jsonReporter := NewJSONReporter(tmpDir, true)
	jsonData, err := jsonReporter.GenerateReport(result)
	if err != nil {
		t.Fatalf("GenerateReport (JSON) failed: %v", err)
	}

	var parsed ChallengeResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON report: %v", err)
	}

	if parsed.Status != StatusPassed {
		t.Errorf("JSON report status = %s, want passed", parsed.Status)
	}
}

func TestIntegration_RegistryWithFailures(t *testing.T) {
	registry := NewRegistry()

	// Create challenges with mixed results
	ch1 := NewMockChallenge("setup", "Setup", nil)
	ch2 := NewMockChallenge("main_test", "Main Test", []ChallengeID{"setup"})
	ch2.shouldFail = true
	ch3 := NewMockChallenge("cleanup", "Cleanup", []ChallengeID{"main_test"})

	registry.Register(ch1)
	registry.Register(ch2)
	registry.Register(ch3)

	order, err := registry.GetDependencyOrder()
	if err != nil {
		t.Fatalf("GetDependencyOrder failed: %v", err)
	}

	var results []*ChallengeResult
	for _, ch := range order {
		result, _ := ch.Execute(context.Background())
		results = append(results, result)
	}

	// Build summary
	summary := BuildMasterSummary(results)

	if summary.TotalChallenges != 3 {
		t.Errorf("TotalChallenges = %d, want 3", summary.TotalChallenges)
	}
	if summary.PassedChallenges != 2 {
		t.Errorf("PassedChallenges = %d, want 2", summary.PassedChallenges)
	}
	if summary.FailedChallenges != 1 {
		t.Errorf("FailedChallenges = %d, want 1", summary.FailedChallenges)
	}

	// Verify pass rate calculation
	expectedRate := 2.0 / 3.0
	if summary.AveragePassRate != expectedRate {
		t.Errorf("AveragePassRate = %f, want %f", summary.AveragePassRate, expectedRate)
	}
}

func TestIntegration_MultiLoggerWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file logger
	fileLogger, err := NewJSONLogger(LoggerConfig{
		OutputPath: filepath.Join(tmpDir, "file.log"),
		Level:      LevelDebug,
		Verbose:    true,
	})
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}

	// Create console-like logger (to buffer for testing)
	consoleLogger := NewConsoleLogger(true)

	// Create multi logger
	multi := NewMultiLogger(fileLogger, consoleLogger)

	// Use the multi logger
	multi.Info("Test message", Field{Key: "test", Value: "value"})
	multi.Warn("Warning message")
	multi.Error("Error message")
	multi.Debug("Debug message")

	multi.Close()

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(tmpDir, "file.log"))
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Verify JSON format
	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Log line is not valid JSON: %s", line)
		}
	}
}

func TestIntegration_HistoryTracking(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.jsonl")

	// Simulate multiple runs
	for run := 1; run <= 3; run++ {
		result := &ChallengeResult{
			ChallengeID:   ChallengeID("challenge_1"),
			ChallengeName: "Challenge One",
			Status:        StatusPassed,
			EndTime:       time.Now(),
			Duration:      time.Duration(run) * time.Minute,
			Assertions: []AssertionResult{
				{Passed: true},
				{Passed: run > 1}, // Second assertion fails on first run
			},
		}

		resultsPath := filepath.Join(tmpDir, "results", "run_"+string(rune('0'+run)))
		if err := AppendToHistory(historyPath, result, resultsPath); err != nil {
			t.Errorf("AppendToHistory run %d failed: %v", run, err)
		}
	}

	// Read and verify history
	data, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(lines))
	}

	// Parse first entry (should have 1 passed assertion)
	var entry1 HistoricalEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("Failed to parse entry 1: %v", err)
	}
	if entry1.AssertionsPassed != 1 {
		t.Errorf("Entry 1 AssertionsPassed = %d, want 1", entry1.AssertionsPassed)
	}

	// Parse second entry (should have 2 passed assertions)
	var entry2 HistoricalEntry
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("Failed to parse entry 2: %v", err)
	}
	if entry2.AssertionsPassed != 2 {
		t.Errorf("Entry 2 AssertionsPassed = %d, want 2", entry2.AssertionsPassed)
	}
}
