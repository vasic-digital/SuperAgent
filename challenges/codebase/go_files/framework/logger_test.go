package framework

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONLogger_Info(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewJSONLogger(LoggerConfig{
		OutputPath: logPath,
		Level:      LevelInfo,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message", Field{Key: "key", Value: "value"})
	logger.Close()

	// Read log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}
	if entry.Fields["key"] != "value" {
		t.Errorf("Expected field key=value, got %v", entry.Fields["key"])
	}
}

func TestJSONLogger_Levels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger at WARN level
	logger, err := NewJSONLogger(LoggerConfig{
		OutputPath: logPath,
		Level:      LevelWarn,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Info should not be logged (below WARN level)
	logger.Info("info message")
	// Warn should be logged
	logger.Warn("warn message")
	// Error should be logged
	logger.Error("error message")
	logger.Close()

	// Read log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}
}

func TestJSONLogger_Debug_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create verbose logger
	logger, err := NewJSONLogger(LoggerConfig{
		OutputPath: logPath,
		Level:      LevelDebug,
		Verbose:    true,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Debug("debug message")
	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(data), "DEBUG") {
		t.Error("Expected debug message to be logged with verbose=true")
	}
}

func TestJSONLogger_WithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewJSONLogger(LoggerConfig{
		OutputPath: logPath,
		Level:      LevelInfo,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	childLogger := logger.WithFields(Field{Key: "request_id", Value: "123"})
	childLogger.Info("child message", Field{Key: "extra", Value: "data"})
	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Fields["request_id"] != "123" {
		t.Error("Expected request_id field from parent")
	}
	if entry.Fields["extra"] != "data" {
		t.Error("Expected extra field from call")
	}
}

func TestJSONLogger_APILogs(t *testing.T) {
	tmpDir := t.TempDir()
	requestLog := filepath.Join(tmpDir, "requests.log")
	responseLog := filepath.Join(tmpDir, "responses.log")

	logger, err := NewJSONLogger(LoggerConfig{
		OutputPath:     filepath.Join(tmpDir, "main.log"),
		APIRequestLog:  requestLog,
		APIResponseLog: responseLog,
		Level:          LevelInfo,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.LogAPIRequest(APIRequestLog{
		Timestamp: "2025-01-01T00:00:00Z",
		RequestID: "req-123",
		Method:    "POST",
		URL:       "https://api.example.com/v1/chat",
		Headers:   map[string]string{"Content-Type": "application/json"},
	})

	logger.LogAPIResponse(APIResponseLog{
		Timestamp:      "2025-01-01T00:00:01Z",
		RequestID:      "req-123",
		StatusCode:     200,
		ResponseTimeMs: 1500,
	})

	logger.Close()

	// Check request log
	reqData, err := os.ReadFile(requestLog)
	if err != nil {
		t.Fatalf("Failed to read request log: %v", err)
	}
	if !strings.Contains(string(reqData), "req-123") {
		t.Error("Request log should contain request ID")
	}

	// Check response log
	respData, err := os.ReadFile(responseLog)
	if err != nil {
		t.Fatalf("Failed to read response log: %v", err)
	}
	if !strings.Contains(string(respData), "1500") {
		t.Error("Response log should contain response time")
	}
}

func TestConsoleLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := &ConsoleLogger{
		output:  &buf,
		verbose: true,
	}

	logger.Info("test info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.Debug("test debug")

	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Error("Console output should contain INFO")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Console output should contain WARN")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Console output should contain ERROR")
	}
	if !strings.Contains(output, "DEBUG") {
		t.Error("Console output should contain DEBUG when verbose")
	}
}

func TestMultiLogger(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	logger1 := &ConsoleLogger{output: &buf1, verbose: false}
	logger2 := &ConsoleLogger{output: &buf2, verbose: false}

	multi := NewMultiLogger(logger1, logger2)
	multi.Info("multi message")

	if !strings.Contains(buf1.String(), "multi message") {
		t.Error("Logger 1 should have message")
	}
	if !strings.Contains(buf2.String(), "multi message") {
		t.Error("Logger 2 should have message")
	}
}

func TestNullLogger(t *testing.T) {
	logger := NullLogger{}

	// Should not panic
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Debug("test")
	logger.LogAPIRequest(APIRequestLog{})
	logger.LogAPIResponse(APIResponseLog{})
	_ = logger.WithFields(Field{Key: "test", Value: "value"})
	_ = logger.Close()
}

func TestRedactingLogger(t *testing.T) {
	var buf bytes.Buffer
	inner := &ConsoleLogger{output: &buf, verbose: false}

	secrets := []string{"sk-ant-secret-12345", "password123"}
	logger := NewRedactingLogger(inner, secrets...)

	logger.Info("API key is sk-ant-secret-12345 and password is password123")

	output := buf.String()

	if strings.Contains(output, "sk-ant-secret-12345") {
		t.Error("Secret should be redacted")
	}
	if strings.Contains(output, "password123") {
		t.Error("Password should be redacted")
	}
	if !strings.Contains(output, "sk-a") {
		t.Error("Redacted key should show first 4 chars")
	}
}

func TestRedactingLogger_APIRequest(t *testing.T) {
	var buf bytes.Buffer
	inner := &ConsoleLogger{output: &buf, verbose: false}

	secrets := []string{"secret-key"}
	logger := NewRedactingLogger(inner, secrets...)

	logger.LogAPIRequest(APIRequestLog{
		RequestID: "req-123",
		Method:    "POST",
		URL:       "https://api.example.com",
		Headers: map[string]string{
			"Authorization": "Bearer secret-key",
			"Content-Type":  "application/json",
		},
	})

	output := buf.String()

	// Authorization header should be redacted
	if strings.Contains(output, "secret-key") {
		t.Error("Authorization header value should be redacted")
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

func TestSetupLogging(t *testing.T) {
	tmpDir := t.TempDir()

	logger, err := SetupLogging(tmpDir, true)
	if err != nil {
		t.Fatalf("SetupLogging failed: %v", err)
	}
	defer logger.Close()

	logger.Info("test message")
	logger.Debug("debug message") // Should be logged with verbose=true

	// Check files were created
	files := []string{"challenge.log", "api_requests.log", "api_responses.log"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file to be created: %s", f)
		}
	}
}

func TestLogField(t *testing.T) {
	field := LogField("key", "value")

	if field.Key != "key" {
		t.Errorf("LogField key = %s, want 'key'", field.Key)
	}
	if field.Value != "value" {
		t.Errorf("LogField value = %v, want 'value'", field.Value)
	}
}
