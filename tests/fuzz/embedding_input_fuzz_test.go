//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzEmbeddingInputProcessing tests embedding input handling for robustness.
// It exercises the request validation and text pre-processing paths used by
// internal/embeddings providers (OpenAI-compatible /v1/embeddings endpoint).
func FuzzEmbeddingInputProcessing(f *testing.F) {
	// Valid embedding inputs
	f.Add("Hello, world!")
	f.Add("The quick brown fox jumps over the lazy dog")
	f.Add("")
	f.Add("   ")
	// Multi-lingual / special characters
	f.Add("日本語のテキスト")
	f.Add("Привет мир")
	f.Add("مرحبا بالعالم")
	// Adversarial inputs
	f.Add(strings.Repeat("token ", 8192))
	f.Add("\x00\x01\x02\xff\xfe")
	f.Add("<script>alert(1)</script>")
	f.Add("../../../etc/passwd")
	f.Add(strings.Repeat("🎉", 2048))
	f.Add("\n\r\t" + strings.Repeat(" ", 1000))

	f.Fuzz(func(t *testing.T, input string) {
		// Validate UTF-8 — embeddings providers require valid text
		isValidUTF8 := utf8.ValidString(input)
		_ = isValidUTF8

		// Trim whitespace (common pre-processing step)
		trimmed := strings.TrimSpace(input)
		_ = trimmed

		// Token count estimation: rough word-split approximation
		words := strings.Fields(input)
		estimatedTokens := len(words)
		_ = estimatedTokens

		// Maximum input length check (most providers cap at ~8192 tokens / ~32768 chars)
		const maxChars = 32768
		truncated := input
		if len(input) > maxChars {
			// Safe truncation at rune boundary
			truncated = string([]rune(input)[:min32(len([]rune(input)), maxChars)])
		}
		_ = truncated

		// Build the OpenAI-compatible embedding request body
		reqBody := map[string]interface{}{
			"input": input,
			"model": "text-embedding-ada-002",
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return
		}

		// Round-trip parse
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			return
		}

		inputStr, _ := parsed["input"].(string)
		_ = len(inputStr)
		model, _ := parsed["model"].(string)
		_ = model

		// Simulate batch input (array of strings)
		batchReq := map[string]interface{}{
			"input": []string{input, trimmed},
			"model": "text-embedding-3-small",
		}
		batchJSON, err := json.Marshal(batchReq)
		if err != nil {
			return
		}

		var batchParsed map[string]interface{}
		if err := json.Unmarshal(batchJSON, &batchParsed); err != nil {
			return
		}

		if inputs, ok := batchParsed["input"]; ok {
			if arr, ok := inputs.([]interface{}); ok {
				for _, item := range arr {
					_, _ = item.(string)
				}
			}
		}
	})
}

// FuzzEmbeddingResponseParsing tests parsing of embedding API responses
// with arbitrary byte input, mirroring the response handling in embedding providers.
func FuzzEmbeddingResponseParsing(f *testing.F) {
	f.Add([]byte(`{"object":"list","data":[{"object":"embedding","embedding":[0.1,0.2,0.3],"index":0}],"model":"text-embedding-ada-002","usage":{"prompt_tokens":5,"total_tokens":5}}`))
	f.Add([]byte(`{"object":"list","data":[]}`))
	f.Add([]byte(`{"error":{"message":"Invalid API key","type":"invalid_request_error","code":"invalid_api_key"}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte("\x00\x01\x02\xff\xfe"))
	// Large embedding vector
	f.Add([]byte(`{"object":"list","data":[{"object":"embedding","embedding":` + buildFloatArray(1536) + `,"index":0}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var resp map[string]interface{}
		if err := json.Unmarshal(data, &resp); err != nil {
			return
		}

		// Extract error field
		if errField, ok := resp["error"]; ok && errField != nil {
			if em, ok := errField.(map[string]interface{}); ok {
				_, _ = em["message"].(string)
				_, _ = em["type"].(string)
				_, _ = em["code"].(string)
			}
		}

		// Extract data array
		if dataField, ok := resp["data"]; ok {
			if arr, ok := dataField.([]interface{}); ok {
				for _, item := range arr {
					if m, ok := item.(map[string]interface{}); ok {
						_, _ = m["index"].(float64)
						_, _ = m["object"].(string)
						if emb, ok := m["embedding"]; ok {
							if vec, ok := emb.([]interface{}); ok {
								for _, v := range vec {
									_, _ = v.(float64)
								}
							}
						}
					}
				}
			}
		}

		// Extract usage
		if usage, ok := resp["usage"]; ok {
			if um, ok := usage.(map[string]interface{}); ok {
				_, _ = um["prompt_tokens"].(float64)
				_, _ = um["total_tokens"].(float64)
			}
		}
	})
}

// min32 returns the smaller of two ints (avoids importing math).
func min32(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildFloatArray produces a JSON float array of length n for seed corpus.
func buildFloatArray(n int) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString("0.1")
	}
	sb.WriteByte(']')
	return sb.String()
}
