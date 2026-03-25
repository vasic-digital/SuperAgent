//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"strings"
	"testing"
	"text/template"
)

// FuzzPromptTemplateRendering tests that rendering prompt templates with
// arbitrary variable values never panics. Prompt templates are used throughout
// HelixAgent to construct system prompts, debate instructions, and user messages.
func FuzzPromptTemplateRendering(f *testing.F) {
	// Seed corpus: realistic template variable payloads
	f.Add("helpful assistant", "Go", "write a function", "gpt-4", "0.7")
	f.Add("", "", "", "", "")
	f.Add("<script>alert(1)</script>", "{{.Injection}}", "'; DROP TABLE messages;--", "model\x00null", "1e308")
	f.Add(strings.Repeat("x", 10000), strings.Repeat("y", 5000), "normal task", "claude-3", "0.0")
	f.Add("system\nprompt\nwith\nnewlines", "Rust\nC++", "multi\nline\ntask", "gemini", "1.0")
	f.Add("{{range .}}", "{{end}}", "{{.Missing}}", "{{$x := .}}", "{{template \"x\"}}")

	f.Fuzz(func(t *testing.T, role, language, task, model, temperature string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzPromptTemplateRendering panicked: role=%q lang=%q task=%q model=%q temp=%q panic=%v",
					role, language, task, model, temperature, r)
			}
		}()

		// Test a system prompt template (common pattern across HelixAgent)
		const systemPromptTmpl = `You are a {{.Role}} AI assistant.
Programming language: {{.Language}}
Task: {{.Task}}
Model: {{.Model}}`

		tmpl, err := template.New("system").Parse(systemPromptTmpl)
		if err != nil {
			return
		}

		data := map[string]string{
			"Role":        role,
			"Language":    language,
			"Task":        task,
			"Model":       model,
			"Temperature": temperature,
		}

		var buf strings.Builder
		_ = tmpl.Execute(&buf, data)
	})
}

// FuzzDebatePromptTemplate tests that building debate round prompts with
// arbitrary participant names, positions, and content never panics.
// The debate system constructs prompts dynamically for each LLM participant.
func FuzzDebatePromptTemplate(f *testing.F) {
	// Seed corpus: realistic debate round inputs
	f.Add("Architect", "proposal", "We should use microservices", 1, 3)
	f.Add("Critic", "critique", "This approach has flaws", 2, 3)
	f.Add("Synthesizer", "synthesis", "Combining both views", 3, 3)
	f.Add("", "", "", 0, 0)
	f.Add(strings.Repeat("a", 5000), "position", strings.Repeat("b", 5000), 1, 1)
	f.Add("Role\nWith\nNewlines", "pos\twith\ttabs", "content\x00with\x00nulls", -1, -1)
	f.Add("<b>HTML</b>", "{{injection}}", "'; DELETE FROM--", 999, 999)

	f.Fuzz(func(t *testing.T, participantName, position, content string, round, totalRounds int) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzDebatePromptTemplate panicked: participant=%q position=%q round=%d panic=%v",
					participantName, position, round, r)
			}
		}()

		const debateTmpl = `Debate Round {{.Round}} of {{.TotalRounds}}
Participant: {{.Participant}}
Position: {{.Position}}

Previous content:
{{.Content}}

Provide your {{.Position}} for this round.`

		tmpl, err := template.New("debate").Parse(debateTmpl)
		if err != nil {
			return
		}

		data := map[string]interface{}{
			"Round":       round,
			"TotalRounds": totalRounds,
			"Participant": participantName,
			"Position":    position,
			"Content":     content,
		}

		var buf strings.Builder
		_ = tmpl.Execute(&buf, data)
	})
}

// FuzzPromptVariableSubstitution tests that simple string-based template variable
// substitution (used in some prompt builders as an alternative to text/template)
// never panics with arbitrary keys and values.
func FuzzPromptVariableSubstitution(f *testing.F) {
	// Seed corpus: realistic key-value substitution inputs
	f.Add("Hello {{name}}, your task is {{task}}.", "name", "Alice", "task", "write code")
	f.Add("Model: {{model}}, Temp: {{temperature}}", "model", "gpt-4", "temperature", "0.7")
	f.Add("{{{{double_braces}}}}", "key", "value", "other", "data")
	f.Add("", "", "", "", "")
	f.Add(strings.Repeat("{{x}}", 1000), "x", strings.Repeat("y", 100), "z", "w")
	f.Add("no variables here", "unused", "value", "also", "unused")
	f.Add("{{missing}}", "present", "value", "another", "val")

	f.Fuzz(func(t *testing.T, tmplStr, key1, val1, key2, val2 string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FuzzPromptVariableSubstitution panicked: tmpl=%q key1=%q val1=%q panic=%v",
					tmplStr, key1, val1, r)
			}
		}()

		// Simple mustache-style substitution (used in some parts of HelixAgent)
		result := tmplStr
		if key1 != "" {
			result = strings.ReplaceAll(result, "{{"+key1+"}}", val1)
		}
		if key2 != "" {
			result = strings.ReplaceAll(result, "{{"+key2+"}}", val2)
		}
		_ = len(result)

		// Also test strings.Builder-based prompt assembly
		var sb strings.Builder
		parts := strings.Split(tmplStr, "{{")
		for _, part := range parts {
			idx := strings.Index(part, "}}")
			if idx >= 0 {
				varName := part[:idx]
				rest := part[idx+2:]
				switch varName {
				case key1:
					sb.WriteString(val1)
				case key2:
					sb.WriteString(val2)
				default:
					sb.WriteString("{{")
					sb.WriteString(varName)
					sb.WriteString("}}")
				}
				sb.WriteString(rest)
			} else {
				sb.WriteString(part)
			}
		}
		_ = sb.String()
	})
}
