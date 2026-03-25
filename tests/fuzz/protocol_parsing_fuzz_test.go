//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"strings"
	"testing"

	"dev.helix.agent/internal/mcp/bridge"
)

// FuzzJSONRPCRequestParsing tests that parsing MCP JSON-RPC 2.0 request messages
// from arbitrary bytes never panics. JSON-RPC is the wire protocol for all MCP
// communication between HelixAgent and tool servers.
func FuzzJSONRPCRequestParsing(f *testing.F) {
	// Seed corpus: valid and boundary JSON-RPC request payloads
	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":"req-abc","method":"tools/call","params":{"name":"bash","arguments":{"command":"ls"}}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":null,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}`))
	f.Add([]byte(`{"jsonrpc":"1.0","id":0,"method":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte("\x00\x01\x02\xff\xfe"))
	f.Add([]byte(`{"jsonrpc":"2.0","id":` + strings.Repeat("9", 100) + `,"method":"test"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzJSONRPCRequestParsing panicked with input %q: %v", data, r)
			}
		}()

		// Unmarshal into JSONRPCRequest — must not panic
		var req bridge.JSONRPCRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}

		// Safe field access
		_ = len(req.JSONRPC)
		_ = len(req.Method)
		_ = len(req.Params)

		// Re-marshal for round-trip safety
		_, _ = json.Marshal(&req)
	})
}

// FuzzJSONRPCResponseParsing tests that parsing MCP JSON-RPC 2.0 response messages
// from arbitrary bytes never panics. Responses may contain nested result or error
// objects from tool servers.
func FuzzJSONRPCResponseParsing(f *testing.F) {
	// Seed corpus: valid and boundary JSON-RPC response payloads
	f.Add(`{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"bash","description":"Execute shell"}]}}`)
	f.Add(`{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"Invalid Request"}}`)
	f.Add(`{"jsonrpc":"2.0","id":2,"error":{"code":-32601,"message":"Method not found","data":{"method":"unknown"}}}`)
	f.Add(`{"jsonrpc":"2.0","id":null,"result":null}`)
	f.Add(`{"jsonrpc":"2.0","result":{}}`)
	f.Add(`{}`)
	f.Add(`{"error":{"code":0}}`)
	f.Add(`invalid json`)
	f.Add(`{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"` + strings.Repeat("hello ", 500) + `"}]}}`)

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzJSONRPCResponseParsing panicked with input %q: %v", input, r)
			}
		}()

		var resp bridge.JSONRPCResponse
		if err := json.Unmarshal([]byte(input), &resp); err != nil {
			return
		}

		// Safe field access
		_ = len(resp.JSONRPC)
		_ = len(resp.Result)

		if resp.Error != nil {
			_ = resp.Error.Code
			_ = len(resp.Error.Message)
		}

		// Re-marshal for round-trip safety
		_, _ = json.Marshal(&resp)
	})
}

// FuzzMCPProtocolBatch tests that parsing batched JSON-RPC messages (arrays of
// requests or responses) never panics, including deeply nested or oversized batches.
func FuzzMCPProtocolBatch(f *testing.F) {
	// Seed corpus: batch request/response payloads
	f.Add(`[{"jsonrpc":"2.0","id":1,"method":"tools/list"},{"jsonrpc":"2.0","id":2,"method":"resources/list"}]`)
	f.Add(`[{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}]`)
	f.Add(`[]`)
	f.Add(`[null]`)
	f.Add(`[{}]`)
	f.Add(`[` + strings.Repeat(`{"jsonrpc":"2.0","id":1,"method":"ping"},`, 100) + `{"jsonrpc":"2.0","id":101,"method":"ping"}]`)

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzMCPProtocolBatch panicked with input %q: %v", input, r)
			}
		}()

		// Try as batch of requests
		var reqs []bridge.JSONRPCRequest
		if err := json.Unmarshal([]byte(input), &reqs); err == nil {
			for _, req := range reqs {
				_ = len(req.Method)
				_ = len(req.JSONRPC)
			}
			_, _ = json.Marshal(reqs)
		}

		// Try as batch of responses
		var resps []bridge.JSONRPCResponse
		if err := json.Unmarshal([]byte(input), &resps); err == nil {
			for _, resp := range resps {
				_ = len(resp.JSONRPC)
				if resp.Error != nil {
					_ = resp.Error.Code
				}
			}
		}

		// Try as generic JSON array
		var generic []map[string]interface{}
		if err := json.Unmarshal([]byte(input), &generic); err == nil {
			for _, item := range generic {
				if id, ok := item["id"]; ok {
					_ = id
				}
				if method, ok := item["method"].(string); ok {
					_ = len(method)
				}
			}
		}
	})
}
