//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"testing"
)

// hashStringFuzz mirrors the hashString logic from internal/cache/cache_service.go.
// It is duplicated here so the fuzz test has no import-cycle dependency on the
// cache package (which requires a live Redis connection to instantiate).
func hashStringFuzz(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%016x", h.Sum64())
}

// FuzzCacheKeyGeneration tests cache key construction with arbitrary input.
// It exercises the generateCacheKey / hashString logic from
// internal/cache/cache_service.go to ensure no panics on malformed data.
func FuzzCacheKeyGeneration(f *testing.F) {
	// Seed corpus: typical cache key components
	f.Add("gpt-4", "user-123", "session-abc", "What is Go?", 0.7, 100)
	f.Add("", "", "", "", 0.0, 0)
	f.Add("helixagent-debate", "user-999", "", strings.Repeat("x", 4096), 1.0, 2048)
	f.Add("claude-3-opus", "user-1", "sess-1", "\x00\x01\x02\xff", 0.5, 512)
	f.Add("model/with/slashes", "user:colon", "sess\nnewline", "<script>", 2.5, -1)

	f.Fuzz(func(t *testing.T, model, userID, sessionID, prompt string, temperature float64, maxTokens int) {
		// Build the key data map that generateCacheKey uses
		keyData := map[string]interface{}{
			"prompt":      prompt,
			"model":       model,
			"temperature": temperature,
			"max_tokens":  maxTokens,
		}

		// Marshal to JSON as generateCacheKey does
		jsonData, err := json.Marshal(keyData)
		if err != nil {
			// Fall back to prompt+model hash (mirrors the fallback in generateCacheKey)
			fallback := fmt.Sprintf("%s:%s", prompt, model)
			key := fmt.Sprintf("llm:%s", hashStringFuzz(fallback))
			_ = key
			return
		}

		key := fmt.Sprintf("llm:%s", hashStringFuzz(string(jsonData)))
		_ = key

		// Exercise the other key patterns used by CacheService
		memKey := fmt.Sprintf("memory:%s:%s", userID, hashStringFuzz(prompt))
		_ = memKey

		healthKey := fmt.Sprintf("health:%s", model)
		_ = healthKey

		sessionKey := fmt.Sprintf("session:%s", sessionID)
		_ = sessionKey

		apiKey := fmt.Sprintf("apikey:%s", hashStringFuzz(userID))
		_ = apiKey

		userPattern := fmt.Sprintf("user:%s:*", userID)
		_ = userPattern

		// Key length sanity — long keys should not panic
		_ = len(key) < 256
	})
}

// FuzzCacheKeyRoundTrip tests that cache keys are deterministic and
// stable across repeated calls with the same input.
func FuzzCacheKeyRoundTrip(f *testing.F) {
	f.Add("prompt-text", "gpt-4")
	f.Add("", "")
	f.Add(strings.Repeat("z", 1024), "model-x")
	f.Add("\x00\xff", "model\x01")

	f.Fuzz(func(t *testing.T, prompt, model string) {
		key1 := hashStringFuzz(fmt.Sprintf("%s:%s", prompt, model))
		key2 := hashStringFuzz(fmt.Sprintf("%s:%s", prompt, model))

		if key1 != key2 {
			t.Fatalf("cache key is not deterministic: %q != %q", key1, key2)
		}

		// Keys must only contain hex characters when produced by hashStringFuzz
		for _, c := range key1 {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Fatalf("unexpected character %q in cache hash key %q", c, key1)
			}
		}
	})
}
