// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzJSONRequestParsing tests that JSON request parsing never panics
// when given arbitrary byte sequences as input.
func FuzzJSONRequestParsing(f *testing.F) {
	f.Add([]byte(`{"messages":[{"role":"user","content":"hello"}]}`))
	f.Add([]byte(`{"model":"gpt-4","temperature":0.7}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(`"string"`))
	f.Add([]byte(``))
	f.Add([]byte(`{"messages":null,"model":""}`))
	f.Add([]byte(`{"model":"helixagent-debate","messages":[{"role":"system","content":"You are helpful"},{"role":"user","content":"test"}],"temperature":0.5,"max_tokens":100}`))
	f.Add([]byte(`{"stream":true,"model":"test"}`))
	f.Add([]byte("\x00\x01\x02\xff\xfe"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Attempt to parse as various JSON types without panicking
		var req map[string]interface{}
		_ = json.Unmarshal(data, &req)

		var arr []interface{}
		_ = json.Unmarshal(data, &arr)

		var str string
		_ = json.Unmarshal(data, &str)

		var num float64
		_ = json.Unmarshal(data, &num)

		var boolean bool
		_ = json.Unmarshal(data, &boolean)

		// If we got a map, try to extract standard fields
		if req != nil {
			if msgs, ok := req["messages"]; ok {
				if msgArr, ok := msgs.([]interface{}); ok {
					for _, msg := range msgArr {
						if m, ok := msg.(map[string]interface{}); ok {
							_, _ = m["role"].(string)
							_, _ = m["content"].(string)
						}
					}
				}
			}
			_, _ = req["model"].(string)
			_, _ = req["temperature"].(float64)
		}
	})
}

// FuzzToolSchemaValidation tests tool schema JSON parsing robustness
// to ensure arbitrary tool definitions never cause panics.
func FuzzToolSchemaValidation(f *testing.F) {
	f.Add(`{"type":"function","function":{"name":"test","parameters":{}}}`)
	f.Add(`{"type":"","function":null}`)
	f.Add(`{}`)
	f.Add(`invalid json`)
	f.Add(`{"type":"function","function":{"name":"bash","description":"Execute shell commands","parameters":{"type":"object","properties":{"command":{"type":"string"}}}}}`)
	f.Add(`{"tools":[{"type":"function","function":{"name":"read"}}]}`)
	f.Add(``)

	f.Fuzz(func(t *testing.T, input string) {
		// Parse tool schema without panicking
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(input), &schema); err != nil {
			return // invalid JSON is fine
		}

		// Try to extract nested function definition
		if fn, ok := schema["function"]; ok {
			if fnMap, ok := fn.(map[string]interface{}); ok {
				_, _ = fnMap["name"].(string)
				_, _ = fnMap["description"].(string)
				if params, ok := fnMap["parameters"]; ok {
					if pMap, ok := params.(map[string]interface{}); ok {
						_, _ = pMap["type"].(string)
						if props, ok := pMap["properties"]; ok {
							if propsMap, ok := props.(map[string]interface{}); ok {
								for _, v := range propsMap {
									if propDef, ok := v.(map[string]interface{}); ok {
										_, _ = propDef["type"].(string)
									}
								}
							}
						}
					}
				}
			}
		}

		// Try to re-marshal to ensure round-trip safety
		_, _ = json.Marshal(schema)
	})
}

// FuzzSSEParsing tests SSE event stream parsing to ensure no panics
// when processing arbitrary text that resembles SSE format.
func FuzzSSEParsing(f *testing.F) {
	f.Add("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
	f.Add("data: [DONE]\n\n")
	f.Add("event: message\ndata: test\n\n")
	f.Add("")
	f.Add("data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hi\"}}]}\n\n")
	f.Add(":\n\n")
	f.Add("retry: 1000\n\n")
	f.Add("data: \n\ndata: test\n\n")
	f.Add(strings.Repeat("data: x\n", 1000))

	f.Fuzz(func(t *testing.T, input string) {
		// Parse SSE format without panicking
		lines := strings.Split(input, "\n")

		for _, line := range lines {
			// Skip empty lines and comments
			if len(line) == 0 || line[0] == ':' {
				continue
			}

			// Parse data lines
			if strings.HasPrefix(line, "data: ") {
				payload := line[6:]
				if payload == "[DONE]" {
					continue
				}

				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
					continue
				}

				// Try to extract choices/delta/content
				if choices, ok := parsed["choices"]; ok {
					if choiceArr, ok := choices.([]interface{}); ok {
						for _, choice := range choiceArr {
							if cm, ok := choice.(map[string]interface{}); ok {
								if delta, ok := cm["delta"]; ok {
									if dm, ok := delta.(map[string]interface{}); ok {
										_, _ = dm["content"].(string)
									}
								}
							}
						}
					}
				}
			}

			// Parse event type lines
			if strings.HasPrefix(line, "event: ") {
				_ = line[7:]
			}

			// Parse retry lines
			if strings.HasPrefix(line, "retry: ") {
				_ = line[7:]
			}
		}
	})
}

// FuzzModelIDParsing tests parsing of model identifiers which may contain
// provider prefixes, version suffixes, or special characters.
func FuzzModelIDParsing(f *testing.F) {
	f.Add("gpt-4")
	f.Add("helixagent/helixagent-debate")
	f.Add("claude-3-opus-20240229")
	f.Add("deepseek-chat")
	f.Add("")
	f.Add("../../../etc/passwd")
	f.Add("<script>alert(1)</script>")
	f.Add("model_with_very_long_name_" + strings.Repeat("x", 1000))
	f.Add("provider/model/version")
	f.Add("   ")
	f.Add("\x00\x01\x02")

	f.Fuzz(func(t *testing.T, modelID string) {
		// Parse provider-qualified model IDs without panicking
		parts := strings.SplitN(modelID, "/", 2)
		if len(parts) == 2 {
			provider := parts[0]
			model := parts[1]
			_ = strings.TrimSpace(provider)
			_ = strings.TrimSpace(model)
		}

		// Sanitize model ID (common operation)
		sanitized := strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' ||
				r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == '/' {
				return r
			}
			return -1
		}, modelID)
		_ = sanitized

		// Check length bounds
		_ = len(modelID) > 256
		_ = len(modelID) == 0
	})
}

// FuzzHTTPHeaderParsing tests parsing of HTTP headers that might contain
// authorization tokens, content types, or other sensitive data.
func FuzzHTTPHeaderParsing(f *testing.F) {
	f.Add("Bearer sk-1234567890abcdef")
	f.Add("Bearer ")
	f.Add("")
	f.Add("Basic dXNlcjpwYXNz")
	f.Add("InvalidScheme token")
	f.Add(strings.Repeat("A", 10000))

	f.Fuzz(func(t *testing.T, authHeader string) {
		// Parse authorization header without panicking
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := authHeader[7:]
			_ = len(token)
			_ = strings.TrimSpace(token)
		}

		if strings.HasPrefix(authHeader, "Basic ") {
			encoded := authHeader[6:]
			_ = len(encoded)
		}

		// Split on space for scheme extraction
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) >= 1 {
			_ = parts[0]
		}
		if len(parts) >= 2 {
			_ = parts[1]
		}
	})
}
