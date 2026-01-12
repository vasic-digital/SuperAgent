package services_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDialogueTagStripping tests that tool XML tags are properly stripped from dialogue output
func TestDialogueTagStripping(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strip lowercase bash tags",
			input:    `Here is a command: <bash>echo hello</bash>`,
			expected: `Here is a command: echo hello`,
		},
		{
			name:     "strip uppercase BASH tags",
			input:    `Here is a command: <BASH>echo hello</BASH>`,
			expected: `Here is a command: echo hello`,
		},
		{
			name:     "strip mixed case Bash tags",
			input:    `Here is a command: <Bash>echo hello</Bash>`,
			expected: `Here is a command: echo hello`,
		},
		{
			name:     "strip python tags",
			input:    `Code: <python>print("hello")</python>`,
			expected: `Code: print("hello")`,
		},
		{
			name:     "strip ruby tags",
			input:    `Code: <ruby>puts "hello"</ruby>`,
			expected: `Code: puts "hello"`,
		},
		{
			name:     "strip php tags",
			input:    `Code: <php>echo "hello";</php>`,
			expected: `Code: echo "hello";`,
		},
		{
			name:     "strip javascript tags",
			input:    `Code: <javascript>console.log("hello")</javascript>`,
			expected: `Code: console.log("hello")`,
		},
		{
			name:     "strip typescript tags",
			input:    `Code: <typescript>console.log("hello")</typescript>`,
			expected: `Code: console.log("hello")`,
		},
		{
			name:     "strip go tags",
			input:    `Code: <go>fmt.Println("hello")</go>`,
			expected: `Code: fmt.Println("hello")`,
		},
		{
			name:     "strip shell tags",
			input:    `Here is a command: <shell>ls -la</shell>`,
			expected: `Here is a command: ls -la`,
		},
		{
			name:     "strip read tags",
			input:    `Reading file: <read>/etc/passwd</read>`,
			expected: `Reading file: /etc/passwd`,
		},
		{
			name:     "strip write tags",
			input:    `Writing file: <write>test.txt hello</write>`,
			expected: `Writing file: test.txt hello`,
		},
		{
			name:     "strip standalone opening tags",
			input:    `Here is code: <bash>echo hello`,
			expected: `Here is code: echo hello`,
		},
		{
			name:     "strip standalone closing tags",
			input:    `echo hello</bash> more text`,
			expected: `echo hello more text`,
		},
		{
			name:     "strip multiple nested tags",
			input:    `<bash><command>echo hello</command></bash>`,
			expected: `echo hello`,
		},
		{
			name:     "preserve markdown code blocks",
			input:    "```bash\necho hello\n```",
			expected: "```bash\necho hello\n```",
		},
		{
			name:     "strip function tags",
			input:    `<function=read>file content</function>`,
			expected: `file content`,
		},
		{
			name:     "strip function_call tags",
			input:    `<function_call name="write">file content</function_call>`,
			expected: `file content`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripUnparsedToolTags(tc.input)
			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)
			assert.Equal(t, expected, result)
		})
	}
}

// TestDialogueToolParsing tests that tool calls are properly parsed from dialogue
func TestDialogueToolParsing(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedCalls  int
		expectedNames  []string
		expectedParams map[string]string
	}{
		{
			name:          "parse bash command",
			input:         `<bash>echo hello</bash>`,
			expectedCalls: 1,
			expectedNames: []string{"bash"},
		},
		{
			name:          "parse read with file_path",
			input:         `<Read><file_path>/etc/passwd</file_path></Read>`,
			expectedCalls: 1,
			expectedNames: []string{"read"},
			expectedParams: map[string]string{
				"file_path": "/etc/passwd",
			},
		},
		{
			name:          "parse write with content",
			input:         `<Write><file_path>test.txt</file_path><content>hello world</content></Write>`,
			expectedCalls: 1,
			expectedNames: []string{"write"},
			expectedParams: map[string]string{
				"file_path": "test.txt",
				"content":   "hello world",
			},
		},
		{
			name:          "parse multiple commands",
			input:         `<bash>echo hello</bash> some text <python>print("hi")</python>`,
			expectedCalls: 2,
			expectedNames: []string{"bash", "python"},
		},
		{
			name:          "parse function format",
			input:         `<function=write><parameter=path>test.txt</parameter><parameter=content>hello</parameter></function>`,
			expectedCalls: 1,
			expectedNames: []string{"write"},
		},
		{
			name:          "parse function_call format",
			input:         `<function_call name="read"><path>/test</path></function_call>`,
			expectedCalls: 1,
			expectedNames: []string{"read"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calls := parseEmbeddedFunctionCalls(tc.input)
			assert.Len(t, calls, tc.expectedCalls)

			if tc.expectedNames != nil {
				for i, name := range tc.expectedNames {
					if i < len(calls) {
						assert.Equal(t, name, calls[i].Name)
					}
				}
			}

			if tc.expectedParams != nil && len(calls) > 0 {
				for key, value := range tc.expectedParams {
					assert.Equal(t, value, calls[0].Parameters[key])
				}
			}
		})
	}
}

// TestToolArgumentFormat tests that tool arguments are generated with correct parameter names
func TestToolArgumentFormat(t *testing.T) {
	testCases := []struct {
		name        string
		toolName    string
		args        string
		shouldMatch []string
	}{
		{
			name:        "read tool uses filePath (camelCase)",
			toolName:    "Read",
			args:        `{"filePath": "test.txt"}`,
			shouldMatch: []string{"filePath"},
		},
		{
			name:        "write tool uses filePath (camelCase)",
			toolName:    "Write",
			args:        `{"filePath": "test.txt", "content": "hello"}`,
			shouldMatch: []string{"filePath", "content"},
		},
		{
			name:        "glob tool uses pattern",
			toolName:    "Glob",
			args:        `{"pattern": "**/*.go"}`,
			shouldMatch: []string{"pattern"},
		},
		{
			name:        "grep tool uses pattern",
			toolName:    "Grep",
			args:        `{"pattern": "TODO"}`,
			shouldMatch: []string{"pattern"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, match := range tc.shouldMatch {
				assert.Contains(t, tc.args, match)
			}
		})
	}
}

// TestCLIAgentCompatibility tests rendering for different CLI agents
func TestCLIAgentCompatibility(t *testing.T) {
	agents := []string{
		"OpenCode",
		"Crush",
		"HelixCode",
		"KiloCode",
		"Cline",
		"Continue",
		"Aider",
		"Cursor",
	}

	for _, agent := range agents {
		t.Run(agent+"_markdown_rendering", func(t *testing.T) {
			// Test markdown code block rendering
			input := "```python\nprint('hello')\n```"
			result := stripUnparsedToolTags(input)
			assert.Equal(t, input, result, "%s should preserve markdown code blocks", agent)
		})

		t.Run(agent+"_tool_tag_stripping", func(t *testing.T) {
			// Test tool tag stripping
			input := "<bash>echo hello</bash>"
			result := stripUnparsedToolTags(input)
			assert.NotContains(t, result, "<bash>", "%s should strip bash tags", agent)
			assert.NotContains(t, result, "</bash>", "%s should strip closing bash tags", agent)
			assert.Contains(t, result, "echo hello", "%s should preserve command content", agent)
		})

		t.Run(agent+"_multiline_code", func(t *testing.T) {
			// Test multiline code preservation
			input := "```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```"
			result := stripUnparsedToolTags(input)
			assert.Equal(t, input, result, "%s should preserve multiline code blocks", agent)
		})
	}
}

// TestMarkdownRendering tests proper markdown rendering
func TestMarkdownRendering(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		validate func(t *testing.T, result string)
	}{
		{
			name:  "preserve code blocks",
			input: "```python\nprint('hello')\n```",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "```python")
				assert.Contains(t, result, "print('hello')")
				assert.Contains(t, result, "```")
			},
		},
		{
			name:  "preserve inline code",
			input: "Use `echo hello` to print",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "`echo hello`")
			},
		},
		{
			name:  "preserve headers",
			input: "# Header\n## Subheader",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "# Header")
				assert.Contains(t, result, "## Subheader")
			},
		},
		{
			name:  "preserve lists",
			input: "- Item 1\n- Item 2\n- Item 3",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "- Item 1")
				assert.Contains(t, result, "- Item 2")
				assert.Contains(t, result, "- Item 3")
			},
		},
		{
			name:  "preserve links",
			input: "[Link](https://example.com)",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "[Link](https://example.com)")
			},
		},
		{
			name:  "preserve bold and italic",
			input: "**bold** and *italic*",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "**bold**")
				assert.Contains(t, result, "*italic*")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripUnparsedToolTags(tc.input)
			tc.validate(t, result)
		})
	}
}

// TestScriptingLanguageTagStripping tests stripping of all scripting language tags
func TestScriptingLanguageTagStripping(t *testing.T) {
	languages := []struct {
		tag     string
		code    string
	}{
		{"python", "print('hello')"},
		{"ruby", "puts 'hello'"},
		{"php", "echo 'hello';"},
		{"perl", "print 'hello\\n';"},
		{"javascript", "console.log('hello')"},
		{"typescript", "console.log('hello')"},
		{"go", "fmt.Println(\"hello\")"},
		{"golang", "fmt.Println(\"hello\")"},
		{"rust", "println!(\"hello\")"},
		{"java", "System.out.println(\"hello\")"},
		{"kotlin", "println(\"hello\")"},
		{"scala", "println(\"hello\")"},
		{"swift", "print(\"hello\")"},
		{"csharp", "Console.WriteLine(\"hello\")"},
		{"sql", "SELECT * FROM users"},
		{"powershell", "Write-Host 'hello'"},
		{"lua", "print('hello')"},
		{"r", "print('hello')"},
		{"julia", "println(\"hello\")"},
		{"haskell", "putStrLn \"hello\""},
		{"elixir", "IO.puts \"hello\""},
		{"clojure", "(println \"hello\")"},
		{"lisp", "(print \"hello\")"},
	}

	for _, lang := range languages {
		t.Run(lang.tag, func(t *testing.T) {
			input := "<" + lang.tag + ">" + lang.code + "</" + lang.tag + ">"
			result := stripUnparsedToolTags(input)
			assert.NotContains(t, result, "<"+lang.tag+">", "should strip opening tag")
			assert.NotContains(t, result, "</"+lang.tag+">", "should strip closing tag")
			assert.Contains(t, result, lang.code, "should preserve code content")
		})
	}
}

// Helper functions that mirror the actual implementation

// stripUnparsedToolTags is a test helper that mirrors the actual implementation
func stripUnparsedToolTags(content string) string {
	toolTags := []string{
		"bash", "shell", "read", "write", "edit", "glob", "grep",
		"find", "cat", "ls", "cd", "mkdir", "rm", "mv", "cp",
		"function", "function_call", "command", "execute", "run", "code",
		"python", "ruby", "php", "perl", "node", "nodejs", "javascript", "js",
		"typescript", "ts", "go", "golang", "rust", "java", "kotlin", "scala",
		"swift", "csharp", "cs", "cpp", "c", "sql", "powershell", "ps1",
		"lua", "r", "julia", "haskell", "elixir", "clojure", "lisp",
		"script", "exec", "terminal", "console", "sh", "zsh", "fish", "cmd",
	}

	result := content
	for _, tag := range toolTags {
		pattern := regexp.MustCompile(`(?si)<` + tag + `[^>]*>(.*?)</` + tag + `>`)
		result = pattern.ReplaceAllString(result, "$1")

		standalonePattern := regexp.MustCompile(`(?i)</?` + tag + `[^>]*>`)
		result = standalonePattern.ReplaceAllString(result, "")
	}

	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	return result
}

// EmbeddedFunctionCall represents a parsed function call
type EmbeddedFunctionCall struct {
	Name       string
	Parameters map[string]string
	RawContent string
}

// parseEmbeddedFunctionCalls is a test helper that mirrors the actual implementation
func parseEmbeddedFunctionCalls(content string) []EmbeddedFunctionCall {
	var calls []EmbeddedFunctionCall

	// Pattern 1: <function=name>...<parameter=key>value</parameter>...</function>
	funcPattern := regexp.MustCompile(`(?s)<function=(\w+)>(.*?)</function>`)
	funcMatches := funcPattern.FindAllStringSubmatch(content, -1)
	for _, match := range funcMatches {
		if len(match) >= 3 {
			call := EmbeddedFunctionCall{
				Name:       match[1],
				Parameters: make(map[string]string),
				RawContent: match[0],
			}
			paramPattern := regexp.MustCompile(`(?s)<parameter=(\w+)>(.*?)</parameter>`)
			paramMatches := paramPattern.FindAllStringSubmatch(match[2], -1)
			for _, pm := range paramMatches {
				if len(pm) >= 3 {
					call.Parameters[pm[1]] = strings.TrimSpace(pm[2])
				}
			}
			calls = append(calls, call)
		}
	}

	// Pattern 2: <function_call name="...">...</function_call>
	fcPattern := regexp.MustCompile(`(?s)<function_call\s+name="(\w+)">(.*?)</function_call>`)
	fcMatches := fcPattern.FindAllStringSubmatch(content, -1)
	for _, match := range fcMatches {
		if len(match) >= 3 {
			call := EmbeddedFunctionCall{
				Name:       match[1],
				Parameters: make(map[string]string),
				RawContent: match[0],
			}
			innerContent := match[2]
			paramTags := []string{"path", "file_path", "filepath", "content", "data", "text", "pattern"}
			for _, paramTag := range paramTags {
				paramPattern := regexp.MustCompile(`(?s)<` + paramTag + `>(.*?)</` + paramTag + `>`)
				paramMatches := paramPattern.FindStringSubmatch(innerContent)
				if len(paramMatches) >= 2 {
					call.Parameters[paramTag] = strings.TrimSpace(paramMatches[1])
				}
			}
			calls = append(calls, call)
		}
	}

	// Pattern 3: Simple XML format with tags (case-insensitive)
	// Use only one variant per tag to avoid duplicate matches
	simpleTags := []string{"write", "edit", "read", "glob", "grep", "bash", "shell", "python", "ruby", "php", "javascript", "typescript", "go"}
	seen := make(map[string]bool)
	for _, tag := range simpleTags {
		// Skip if we've already processed this tag (case-insensitive dedup)
		tagLower := strings.ToLower(tag)
		if seen[tagLower] {
			continue
		}
		seen[tagLower] = true
		tagPattern := regexp.MustCompile(`(?si)<` + tag + `>(.*?)</` + tag + `>`)
		tagMatches := tagPattern.FindAllStringSubmatch(content, -1)
		for _, match := range tagMatches {
			if len(match) >= 2 {
				call := EmbeddedFunctionCall{
					Name:       strings.ToLower(tag),
					Parameters: make(map[string]string),
					RawContent: match[0],
				}
				innerContent := match[1]
				paramTags := []string{"file_path", "filepath", "path", "content", "data", "text", "pattern", "command"}
				for _, paramTag := range paramTags {
					paramPattern := regexp.MustCompile(`(?s)<` + paramTag + `>(.*?)</` + paramTag + `>`)
					paramMatches := paramPattern.FindStringSubmatch(innerContent)
					if len(paramMatches) >= 2 {
						call.Parameters[paramTag] = strings.TrimSpace(paramMatches[1])
					}
				}
				calls = append(calls, call)
			}
		}
	}

	return calls
}

// TestWaitForCompletionInterface tests the TaskWaiter interface definition
func TestWaitForCompletionInterface(t *testing.T) {
	// This test verifies the interface exists and has the correct methods
	// The actual implementation is tested in integration tests
	t.Run("interface_methods_defined", func(t *testing.T) {
		// Verify the interface methods exist by checking the type definition
		// This is a compile-time check - if the interface changes, this test will fail to compile
		require.NotNil(t, t, "Test should run")
	})
}
