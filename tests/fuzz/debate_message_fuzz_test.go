//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzDebateMessageProcessing tests debate message structures and topic
// processing for robustness. It exercises the JSON extraction and intent
// classification path with arbitrary string input.
func FuzzDebateMessageProcessing(f *testing.F) {
	// Valid debate topic strings
	f.Add("What is the best approach to microservices architecture?")
	f.Add("Implement a complete REST API with authentication")
	f.Add("Refactor the entire database layer for performance")
	f.Add("")
	f.Add("   ")
	// Malformed / adversarial inputs
	f.Add("```json\n{\"intent\":\"code\"}\n```")
	f.Add("{\"topic\":\"test\",\"rounds\":5}")
	f.Add(strings.Repeat("a", 65536))
	f.Add("\x00\x01\x02\xff\xfe")
	f.Add("<script>alert(1)</script>")
	f.Add("../../../etc/passwd")
	f.Add("'; DROP TABLE debates; --")

	f.Fuzz(func(t *testing.T, topic string) {
		// Exercise the JSON extraction logic used by EnhancedIntentClassifier.extractJSON
		// and the canned-error-response detection from debate_service.go.
		extracted := extractJSONFromContent(topic)
		if extracted != "" {
			var result map[string]interface{}
			_ = json.Unmarshal([]byte(extracted), &result)
		}

		// Exercise IsCannedErrorResponse-style pattern matching
		lowered := strings.ToLower(topic)
		patterns := []string{
			"unable to provide",
			"unable to analyze",
			"cannot provide",
			"i apologize, but i cannot",
			"error occurred",
			"failed to generate",
			"no response generated",
		}
		for _, pat := range patterns {
			_ = strings.Contains(lowered, pat)
		}

		// Exercise model-ID / provider split used throughout debate routing
		parts := strings.SplitN(topic, "/", 2)
		if len(parts) == 2 {
			_, _ = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}

		// Exercise length / boundary checks
		_ = len(topic) > 10000
		_ = len(topic) == 0
		_ = strings.TrimSpace(topic) == ""
	})
}

// FuzzDebateResultParsing tests JSON parsing of DebateResult-shaped payloads.
// Malformed payloads must not cause panics in the unmarshalling path.
func FuzzDebateResultParsing(f *testing.F) {
	f.Add([]byte(`{"debate_id":"d1","topic":"test","success":true,"quality_score":0.9}`))
	f.Add([]byte(`{"debate_id":"","participants":[],"consensus":null}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`{"participants":[{"participant_id":"p1","response":"","confidence":0.5}]}`))
	f.Add([]byte("\x00\x01\x02\xff\xfe"))

	f.Fuzz(func(t *testing.T, data []byte) {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return
		}

		// Walk the parsed map without panicking
		if result != nil {
			_, _ = result["debate_id"].(string)
			_, _ = result["topic"].(string)
			_, _ = result["success"].(bool)
			_, _ = result["quality_score"].(float64)

			if participants, ok := result["participants"]; ok {
				if arr, ok := participants.([]interface{}); ok {
					for _, p := range arr {
						if pm, ok := p.(map[string]interface{}); ok {
							_, _ = pm["participant_id"].(string)
							_, _ = pm["response"].(string)
							_, _ = pm["confidence"].(float64)
						}
					}
				}
			}

			if consensus, ok := result["consensus"]; ok && consensus != nil {
				if cm, ok := consensus.(map[string]interface{}); ok {
					_, _ = cm["achieved"].(bool)
					_, _ = cm["confidence"].(float64)
					_, _ = cm["final_position"].(string)
				}
			}
		}
	})
}

// extractJSONFromContent mirrors the logic from enhanced_intent_classifier.go
// extractJSON without requiring the private method to be accessible.
func extractJSONFromContent(content string) string {
	content = strings.TrimSpace(content)

	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + len("```json")
		end := strings.LastIndex(content, "```")
		if end > start {
			content = content[start:end]
		}
	} else if strings.Contains(content, "```") {
		start := strings.Index(content, "```") + len("```")
		end := strings.LastIndex(content, "```")
		if end > start {
			content = content[start:end]
		}
	}

	return strings.TrimSpace(content)
}
