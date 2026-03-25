//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"dev.helix.agent/internal/tools"
)

// FuzzToolValidation tests that tool validation functions never panic when
// given arbitrary path, symbol, git-ref, or command argument inputs.
func FuzzToolValidation(f *testing.F) {
	// Seed corpus: valid and boundary-breaking inputs
	f.Add("/home/user/project/main.go", "MyFunction", "main", "build")
	f.Add("../../../etc/passwd", "../../../../root", "HEAD~1", "rm -rf /")
	f.Add("", "", "", "")
	f.Add("/valid/path.go", "validSymbol_123", "feature/my-branch", "go test ./...")
	f.Add(strings.Repeat("a", 10000), strings.Repeat("b", 5000), "refs/heads/main", "echo hello")
	f.Add("\x00\x01\x02", "\xff\xfe", "tag/v1.0.0", "cmd && evil")
	f.Add("path with spaces/file.go", "sym$bol", "branch;injection", "arg|pipe")

	f.Fuzz(func(t *testing.T, path, symbol, gitRef, cmdArg string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzToolValidation panicked: path=%q symbol=%q gitRef=%q cmdArg=%q panic=%v",
					path, symbol, gitRef, cmdArg, r)
			}
		}()

		// ValidatePath must not panic
		_ = tools.ValidatePath(path)

		// ValidateSymbol must not panic
		_ = tools.ValidateSymbol(symbol)

		// ValidateGitRef must not panic
		_ = tools.ValidateGitRef(gitRef)

		// ValidateCommandArg must not panic
		_ = tools.ValidateCommandArg(cmdArg)

		// SanitizePath must not panic
		_, _ = tools.SanitizePath(path)

		// filepath.Clean is used internally — verify it doesn't panic either
		if path != "" {
			_ = filepath.Clean(path)
		}
	})
}

// FuzzToolSchemaJSON tests that deserializing arbitrary JSON into ToolSchema
// and ToolFunction types never causes panics, even with deeply nested or
// extremely large inputs.
func FuzzToolSchemaJSON(f *testing.F) {
	// Seed corpus: realistic tool schema definitions
	f.Add(`{"name":"Bash","description":"Execute shell commands","required_fields":["command"],"parameters":{"command":{"type":"string","description":"The command","required":true}}}`)
	f.Add(`{"name":"Read","description":"Read a file","parameters":{"file_path":{"type":"string","required":true},"limit":{"type":"integer","required":false}}}`)
	f.Add(`{"type":"function","function":{"name":"grep","description":"Search files","parameters":{"type":"object","properties":{"pattern":{"type":"string"},"path":{"type":"string"}}}}}`)
	f.Add(`{}`)
	f.Add(`{"name":"","parameters":null}`)
	f.Add(`invalid json`)
	f.Add(`{"parameters":{"` + strings.Repeat("x", 1000) + `":{"type":"string"}}}`)
	f.Add(`{"examples":[{"description":"test","arguments":"{}"},{"description":"another","arguments":"null"}]}`)

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzToolSchemaJSON panicked with input %q: %v", input, r)
			}
		}()

		// Attempt to unmarshal into ToolSchema
		var schema tools.ToolSchema
		if err := json.Unmarshal([]byte(input), &schema); err != nil {
			return // invalid JSON is fine; panics are not
		}

		// Safe field access
		_ = len(schema.Name)
		_ = len(schema.Description)
		_ = len(schema.RequiredFields)
		_ = len(schema.Parameters)

		for name, param := range schema.Parameters {
			_ = len(name)
			_ = len(param.Type)
			_ = len(param.Description)
			_ = param.Required
		}

		// Re-marshal for round-trip safety
		_, _ = json.Marshal(&schema)

		// Also try to look up in the registry — must not panic
		if schema.Name != "" {
			_, _ = tools.ToolSchemaRegistry[schema.Name]
		}
	})
}

// FuzzToolCallArgumentParsing tests that parsing tool call arguments
// (which are JSON strings embedded in LLM responses) never panics.
func FuzzToolCallArgumentParsing(f *testing.F) {
	// Seed corpus: realistic tool call argument payloads
	f.Add(`{"command":"go build ./...","description":"Build the project"}`)
	f.Add(`{"file_path":"/home/user/main.go","limit":100}`)
	f.Add(`{"pattern":"func.*Error","path":"./internal/"}`)
	f.Add(`{}`)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`"string value"`)
	f.Add(`42`)
	f.Add(`true`)
	f.Add(`{"nested":{"deep":{"deeper":{"value":` + strings.Repeat(`{"x":`, 50) + `null` + strings.Repeat(`}`, 50) + `}}}}`)

	f.Fuzz(func(t *testing.T, argsJSON string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzToolCallArgumentParsing panicked with input %q: %v", argsJSON, r)
			}
		}()

		// Parse as generic map (common operation in tool dispatching)
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return
		}

		// Safe extraction of common tool parameters
		if cmd, ok := args["command"].(string); ok {
			_ = len(cmd)
			_ = tools.ValidateCommandArg(cmd)
		}
		if path, ok := args["file_path"].(string); ok {
			_ = tools.ValidatePath(path)
			_, _ = tools.SanitizePath(path)
		}
		if sym, ok := args["symbol"].(string); ok {
			_ = tools.ValidateSymbol(sym)
		}

		// Re-marshal the extracted args
		_, _ = json.Marshal(args)
	})
}
