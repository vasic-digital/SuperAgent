//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"testing"
)

// FuzzMCPJSONRPCParsing tests JSON-RPC 2.0 message parsing used by the MCP
// SSE bridge. Malformed payloads must never cause panics in the unmarshal
// and validation path that mirrors internal/mcp/bridge/sse_bridge.go.
func FuzzMCPJSONRPCParsing(f *testing.F) {
	// Valid JSON-RPC 2.0 requests
	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":"req-abc","method":"tools/call","params":{"name":"bash","arguments":{"command":"ls"}}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":null,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`))
	// Edge / boundary cases
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	// Malformed JSON-RPC
	f.Add([]byte(`{"jsonrpc":"1.0","id":1,"method":"test"}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":1}`))
	f.Add([]byte(`{"jsonrpc":"2.0","method":""}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":` + string(make([]byte, 1024)) + `}`))
	f.Add([]byte("\x00\x01\x02\xff\xfe"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Mirror the parsing and validation logic from sse_bridge.go handleMessages.
		type jsonRPCRequest struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      interface{}     `json:"id,omitempty"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params,omitempty"`
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return // invalid JSON is expected and fine
		}

		// Validate JSON-RPC version (mirrors bridge validation)
		if req.JSONRPC != "2.0" {
			return
		}

		// Method must be non-empty
		if req.Method == "" {
			return
		}

		// Parse params if present
		if len(req.Params) > 0 {
			var params map[string]interface{}
			if err := json.Unmarshal(req.Params, &params); err != nil {
				// Array params are also valid in JSON-RPC
				var arrayParams []interface{}
				_ = json.Unmarshal(req.Params, &arrayParams)
			} else {
				// Try to extract common MCP param fields without panicking
				_, _ = params["name"].(string)
				if args, ok := params["arguments"]; ok {
					if m, ok := args.(map[string]interface{}); ok {
						for k, v := range m {
							_ = k
							_ = v
						}
					}
				}
				_, _ = params["protocolVersion"].(string)
			}
		}

		// Normalize the ID field (mirrors normalizeID logic in sse_bridge.go)
		switch v := req.ID.(type) {
		case float64:
			_ = int64(v)
		case string:
			_ = v
		case nil:
			// notification — no ID
		}

		// Re-marshal to ensure round-trip safety
		_, _ = json.Marshal(req)
	})
}

// FuzzMCPResponseParsing tests JSON-RPC 2.0 response parsing for MCP clients.
func FuzzMCPResponseParsing(f *testing.F) {
	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"bash","description":"Execute shell commands"}]}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":null,"result":null}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"result":` + string(make([]byte, 2048)) + `}`))
	f.Add([]byte("\x00\x01\xff"))

	f.Fuzz(func(t *testing.T, data []byte) {
		type jsonRPCError struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		}
		type jsonRPCResponse struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      interface{}     `json:"id,omitempty"`
			Result  json.RawMessage `json:"result,omitempty"`
			Error   *jsonRPCError   `json:"error,omitempty"`
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return
		}

		if resp.Error != nil {
			_ = resp.Error.Code
			_ = resp.Error.Message
		}

		if len(resp.Result) > 0 {
			var result map[string]interface{}
			_ = json.Unmarshal(resp.Result, &result)
		}
	})
}
