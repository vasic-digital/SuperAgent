//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"testing"

	"dev.helix.agent/internal/models"
)

// FuzzLLMRequestParsing tests that parsing LLMRequest from arbitrary JSON bytes
// never panics. This ensures robustness against malformed API payloads.
func FuzzLLMRequestParsing(f *testing.F) {
	// Seed corpus: valid LLMRequest payloads
	f.Add([]byte(`{"prompt":"hello","messages":[{"role":"user","content":"hi"}]}`))
	f.Add([]byte(`{"model_params":{"temperature":0.7,"max_tokens":100}}`))
	f.Add([]byte(`{"memory_enhanced":true,"memory":{"key":"value"}}`))
	f.Add([]byte(`{"tools":[{"type":"function","function":{"name":"bash","description":"run command"}}]}`))
	f.Add([]byte(`{"tool_choice":"auto"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte("\x00\x01\x02\xff\xfe\xfd"))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzLLMRequestParsing panicked with input %q: %v", data, r)
			}
		}()

		// Attempt to unmarshal into LLMRequest — must not panic
		var req models.LLMRequest
		_ = json.Unmarshal(data, &req)

		// If parsing succeeded, attempt to re-marshal for round-trip safety
		if req.Prompt != "" || len(req.Messages) > 0 || len(req.Memory) > 0 {
			_, _ = json.Marshal(&req)
		}

		// Also try parsing as LLMResponse
		var resp models.LLMResponse
		_ = json.Unmarshal(data, &resp)

		// Also try parsing as a generic OpenAI-compatible chat request
		var chatReq struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Temperature *float64 `json:"temperature"`
			MaxTokens   *int     `json:"max_tokens"`
			Stream      bool     `json:"stream"`
			Tools       []struct {
				Type string `json:"type"`
			} `json:"tools"`
		}
		_ = json.Unmarshal(data, &chatReq)

		// Safe field access — must not panic
		for _, msg := range chatReq.Messages {
			_ = len(msg.Role)
			_ = len(msg.Content)
		}
	})
}

// FuzzMessageParsing tests that parsing Message structs from arbitrary JSON
// never panics. Messages are the core unit of LLM communication.
func FuzzMessageParsing(f *testing.F) {
	// Seed corpus: valid Message payloads
	f.Add(`{"role":"user","content":"hello world"}`)
	f.Add(`{"role":"assistant","content":"I can help with that"}`)
	f.Add(`{"role":"system","content":"You are a helpful AI assistant"}`)
	f.Add(`{"role":"","content":""}`)
	f.Add(`{"role":"tool","content":null}`)
	f.Add(`{}`)
	f.Add(`invalid json`)
	f.Add(`{"role":"` + string(make([]byte, 10000)) + `"}`)

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzMessageParsing panicked with input %q: %v", input, r)
			}
		}()

		var msg models.Message
		if err := json.Unmarshal([]byte(input), &msg); err != nil {
			return // invalid JSON is acceptable; panics are not
		}

		// Safe field access
		_ = len(msg.Role)
		_ = len(msg.Content)

		// Re-marshal for round-trip safety
		_, _ = json.Marshal(&msg)

		// Also parse as slice of messages
		var msgs []models.Message
		_ = json.Unmarshal([]byte(input), &msgs)
		for _, m := range msgs {
			_ = len(m.Role)
		}
	})
}
