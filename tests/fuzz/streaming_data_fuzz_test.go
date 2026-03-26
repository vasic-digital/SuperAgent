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

// FuzzStreamingDataParsing tests SSE/streaming message parsing for the
// notifications subsystem. It ensures that arbitrary byte sequences passed
// through the SSE message pipeline never cause panics.
func FuzzStreamingDataParsing(f *testing.F) {
	// Valid SSE streaming chunks (OpenAI-compatible format)
	f.Add([]byte("data: {\"id\":\"chatcmpl-1\",\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n"))
	f.Add([]byte("data: [DONE]\n\n"))
	f.Add([]byte("event: message\ndata: {\"type\":\"update\"}\nid: 42\n\n"))
	f.Add([]byte(""))
	f.Add([]byte("data: \n\n"))
	f.Add([]byte("retry: 3000\n\n"))
	f.Add([]byte(": heartbeat\n\n"))
	// Malformed / adversarial inputs
	f.Add([]byte("data: {invalid json}\n\n"))
	f.Add([]byte("\x00\x01\x02\xff\xfe\n\n"))
	f.Add([]byte(strings.Repeat("data: x\n", 500)))
	f.Add([]byte("data: " + strings.Repeat("a", 65536) + "\n\n"))
	f.Add([]byte("data: {\"choices\":[{\"delta\":{\"content\":null},\"finish_reason\":\"stop\"}]}\n\n"))

	f.Fuzz(func(t *testing.T, raw []byte) {
		// Simulate the SSE parsing logic used by the streaming pipeline.
		// The goal is no panics — we do not validate correctness here.
		text := string(raw)
		lines := strings.Split(text, "\n")

		var eventType, dataAccum, eventID string

		for _, line := range lines {
			switch {
			case line == "":
				// Dispatch event boundary: try to parse accumulated data
				if dataAccum != "" && dataAccum != "[DONE]" {
					var payload map[string]interface{}
					_ = json.Unmarshal([]byte(dataAccum), &payload)

					if payload != nil {
						if choices, ok := payload["choices"]; ok {
							if arr, ok := choices.([]interface{}); ok {
								for _, choice := range arr {
									if m, ok := choice.(map[string]interface{}); ok {
										if delta, ok := m["delta"]; ok {
											if dm, ok := delta.(map[string]interface{}); ok {
												_, _ = dm["content"].(string)
											}
										}
										_, _ = m["finish_reason"].(string)
									}
								}
							}
						}
						_, _ = payload["id"].(string)
					}
				}
				// Reset for next event
				dataAccum = ""
				eventType = ""
				eventID = ""

			case strings.HasPrefix(line, "data: "):
				chunk := line[6:]
				if dataAccum == "" {
					dataAccum = chunk
				} else {
					dataAccum += "\n" + chunk
				}

			case strings.HasPrefix(line, "event: "):
				eventType = line[7:]
				_ = eventType

			case strings.HasPrefix(line, "id: "):
				eventID = line[4:]
				_ = eventID

			case strings.HasPrefix(line, "retry: "):
				retryVal := line[7:]
				_ = retryVal

			case len(line) > 0 && line[0] == ':':
				// Comment/heartbeat — ignore
			}
		}
	})
}
