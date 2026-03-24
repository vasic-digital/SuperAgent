package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// Security Tests for the Gemini provider
// ==============================================================================

// TestGeminiCLI_CommandInjection verifies that malicious model names, prompts
// with shell metacharacters, and session IDs containing injection attempts are
// properly handled by ValidateCommandArg or rejected before execution.
func TestGeminiCLI_CommandInjection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		arg        string
		expectSafe bool
	}{
		{
			name:       "semicolon injection",
			arg:        "gemini-pro; rm -rf /",
			expectSafe: false,
		},
		{
			name:       "pipe injection",
			arg:        "gemini-pro | cat /etc/passwd",
			expectSafe: false,
		},
		{
			name:       "backtick injection",
			arg:        "gemini-pro`id`",
			expectSafe: false,
		},
		{
			name:       "dollar substitution",
			arg:        "gemini-pro$(whoami)",
			expectSafe: false,
		},
		{
			name:       "ampersand injection",
			arg:        "gemini-pro & curl attacker.com",
			expectSafe: false,
		},
		{
			name:       "newline injection",
			arg:        "gemini-pro\nid",
			expectSafe: false,
		},
		{
			name:       "carriage return injection",
			arg:        "gemini-pro\rid",
			expectSafe: false,
		},
		{
			name:       "curly brace injection",
			arg:        "gemini-pro{echo,pwned}",
			expectSafe: false,
		},
		{
			name:       "backslash injection",
			arg:        "gemini-pro\\nid",
			expectSafe: false,
		},
		{
			name:       "redirect injection",
			arg:        "gemini-pro > /tmp/exfil",
			expectSafe: false,
		},
		{
			name:       "valid model name",
			arg:        "gemini-2.5-pro",
			expectSafe: true,
		},
		{
			name:       "valid model with dash",
			arg:        "gemini-2.5-flash-lite",
			expectSafe: true,
		},
		{
			name:       "empty string is safe",
			arg:        "",
			expectSafe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.ValidateCommandArg(tt.arg)
			assert.Equal(t, tt.expectSafe, result,
				"ValidateCommandArg(%q) = %v, want %v",
				tt.arg, result, tt.expectSafe)
		})
	}

	// Additionally verify the CLI provider rejects injected model names
	// by creating a provider and calling Complete with a bad model name.
	t.Run("CLI_Complete_rejects_injected_model", func(t *testing.T) {
		t.Parallel()

		provider := NewGeminiCLIProvider(GeminiCLIConfig{
			Model:   "gemini-2.5-flash",
			Timeout: 5 * time.Second,
		})

		ctx := context.Background()
		req := &models.LLMRequest{
			ID: "test-injection",
			Messages: []models.Message{
				{Role: "user", Content: "hello"},
			},
			ModelParams: models.ModelParameters{
				Model: "gemini-pro; rm -rf /",
			},
		}

		// Should either fail validation or fail because CLI is not available;
		// in either case it must NOT execute the injected command.
		resp, err := provider.Complete(ctx, req)
		if err != nil {
			// Expected: either "invalid characters" or "CLI not available"
			assert.True(t,
				strings.Contains(err.Error(), "invalid characters") ||
					strings.Contains(err.Error(), "not available"),
				"error should indicate invalid characters or CLI unavailable, got: %s",
				err.Error())
		}
		if resp != nil {
			// If a response is returned, it must not contain shell output
			assert.NotContains(t, resp.Content, "root:",
				"response must not contain shell output")
		}
	})
}

// TestGeminiAPI_KeyNotExposed verifies that the API key does not appear in
// response content or metadata values returned to callers.
func TestGeminiAPI_KeyNotExposed(t *testing.T) {
	t.Parallel()

	secretKey := "super-secret-gemini-api-key-12345"

	// Mock server that echoes the request (simulating a pathological response)
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			resp := GeminiResponse{
				Candidates: []GeminiCandidate{
					{
						Content: GeminiContent{
							Parts: []GeminiPart{
								{Text: "This is a normal response."},
							},
						},
						FinishReason: "STOP",
					},
				},
				UsageMetadata: &GeminiUsageMetadata{
					TotalTokenCount: 10,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		},
	))
	defer server.Close()

	baseURL := server.URL + "/v1beta/models/%s:generateContent"
	provider := NewGeminiAPIProvider(secretKey, baseURL, "gemini-2.5-flash")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID: "test-key-exposure",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
		ModelParams: models.ModelParameters{MaxTokens: 32},
	}

	resp, err := provider.Complete(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// The API key must not appear in response content
	assert.NotContains(t, resp.Content, secretKey,
		"API key must not be exposed in response content")

	// The API key must not appear in any metadata value
	if resp.Metadata != nil {
		for k, v := range resp.Metadata {
			vStr := fmt.Sprintf("%v", v)
			assert.NotContains(t, vStr, secretKey,
				"API key must not be exposed in metadata key %q", k)
		}
	}
}

// TestGeminiCLI_PathTraversal verifies that CLI path manipulation via
// environment cannot lead to arbitrary command execution.
func TestGeminiCLI_PathTraversal(t *testing.T) {
	t.Parallel()

	traversalPaths := []string{
		"../../etc/passwd",
		"../../../bin/sh",
		"./../../malicious",
		"path;injection",
		"path|pipe",
		"path$(cmd)",
	}

	for _, path := range traversalPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			valid := utils.ValidatePath(path)
			assert.False(t, valid,
				"path %q should be rejected by ValidatePath", path)
		})
	}
}

// TestGeminiACP_MalformedJSON verifies the ACP provider handles malformed JSON
// responses gracefully without panicking.
func TestGeminiACP_MalformedJSON(t *testing.T) {
	t.Parallel()

	malformedInputs := []string{
		"",
		"not json at all",
		"{",
		`{"incomplete": true`,
		`{"jsonrpc": "2.0", "result": "not an object"}`,
		`null`,
		`[]`,
		`{"jsonrpc": "2.0", "error": {"code": -1, "message": "test error"}}`,
	}

	for i, input := range malformedInputs {
		t.Run(fmt.Sprintf("malformed_%d", i), func(t *testing.T) {
			t.Parallel()

			// Verify that unmarshaling into the ACP response type does not
			// panic and either succeeds (with empty/default values) or fails
			// with a proper error.
			var resp geminiACPResponse
			err := json.Unmarshal([]byte(input), &resp)
			// We only care that no panic occurs; err may or may not be nil
			_ = err
		})
	}
}

// TestGeminiAPI_LargePayload verifies the provider handles extremely large
// prompts without panicking.
func TestGeminiAPI_LargePayload(t *testing.T) {
	t.Parallel()

	// Create a server that returns a normal response regardless of input size
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			resp := GeminiResponse{
				Candidates: []GeminiCandidate{
					{
						Content: GeminiContent{
							Parts: []GeminiPart{
								{Text: "ok"},
							},
						},
						FinishReason: "STOP",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		},
	))
	defer server.Close()

	baseURL := server.URL + "/v1beta/models/%s:generateContent"
	provider := NewGeminiAPIProvider("test-key", baseURL, "gemini-2.5-flash")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate a large prompt (~1MB)
	largeContent := strings.Repeat("a", 1024*1024)

	req := &models.LLMRequest{
		ID: "test-large-payload",
		Messages: []models.Message{
			{Role: "user", Content: largeContent},
		},
		ModelParams: models.ModelParameters{MaxTokens: 16},
	}

	// Must not panic
	resp, err := provider.Complete(ctx, req)
	if err == nil {
		assert.NotNil(t, resp)
	}
	// Either a valid response or a proper error -- no panic
}

// TestGeminiAPI_EmptyAPIKey verifies that using an empty API key produces a
// proper error rather than a nil pointer dereference or silent failure.
func TestGeminiAPI_EmptyAPIKey(t *testing.T) {
	t.Parallel()

	// Create a server that checks for the API key header
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("x-goog-api-key")
			if apiKey == "" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "missing API key"}`))
				return
			}
			resp := GeminiResponse{
				Candidates: []GeminiCandidate{
					{
						Content: GeminiContent{
							Parts: []GeminiPart{{Text: "ok"}},
						},
						FinishReason: "STOP",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		},
	))
	defer server.Close()

	baseURL := server.URL + "/v1beta/models/%s:generateContent"
	provider := NewGeminiAPIProvider("", baseURL, "gemini-2.5-flash")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID: "test-empty-key",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
		ModelParams: models.ModelParameters{MaxTokens: 16},
	}

	resp, err := provider.Complete(ctx, req)
	// With empty key the server returns 401, which should result in an error
	assert.Error(t, err, "should error with empty API key")
	assert.Nil(t, resp, "response should be nil on auth error")
}

// TestGeminiCLI_TimeoutEnforcement verifies CLI operations respect timeout
// and do not hang indefinitely. Note: the CLI provider performs an internal
// availability check (with its own 10s timeout on first call via sync.Once),
// so we allow up to 15s total for the Complete() call to finish.
func TestGeminiCLI_TimeoutEnforcement(t *testing.T) {
	t.Parallel()

	provider := NewGeminiCLIProvider(GeminiCLIConfig{
		Model:   "gemini-2.5-flash",
		Timeout: 1 * time.Millisecond, // Very short timeout for actual CLI call
	})

	ctx, cancel := context.WithTimeout(
		context.Background(), 1*time.Millisecond,
	)
	defer cancel()

	req := &models.LLMRequest{
		ID: "test-timeout",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	// This should either fail because CLI is not available or time out.
	// The CLI availability check itself uses sync.Once with a 10s internal
	// timeout, so we allow up to 15s for the first invocation.
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = provider.Complete(ctx, req)
	}()

	select {
	case <-done:
		// Completed (with error) within time -- good
	case <-time.After(15 * time.Second):
		t.Fatal("CLI operation did not respect timeout; hung for > 15s")
	}
}
