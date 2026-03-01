package comprehensive

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/sirupsen/logrus"
)

// StaticAnalysisTool performs static code analysis
type StaticAnalysisTool struct {
	logger *logrus.Logger
}

// NewStaticAnalysisTool creates a new static analysis tool
func NewStaticAnalysisTool(logger *logrus.Logger) *StaticAnalysisTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &StaticAnalysisTool{
		logger: logger,
	}
}

// GetName returns the tool name
func (t *StaticAnalysisTool) GetName() string {
	return "static_analysis"
}

// GetType returns the tool type
func (t *StaticAnalysisTool) GetType() ToolType {
	return ToolTypeAnalysis
}

// GetDescription returns the description
func (t *StaticAnalysisTool) GetDescription() string {
	return "Perform static code analysis to find issues"
}

// GetInputSchema returns the input schema
func (t *StaticAnalysisTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Source code to analyze",
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Programming language (go, python, etc.)",
				"default":     "go",
			},
		},
		"required": []string{"code"},
	}
}

// Validate validates inputs
func (t *StaticAnalysisTool) Validate(inputs map[string]interface{}) error {
	code, ok := inputs["code"].(string)
	if !ok || code == "" {
		return fmt.Errorf("code is required")
	}

	return nil
}

// Execute executes the tool
func (t *StaticAnalysisTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	code := inputs["code"].(string)
	language := "go"
	if l, ok := inputs["language"].(string); ok {
		language = l
	}

	var issues []AnalysisIssue

	switch language {
	case "go":
		issues = t.analyzeGo(code)
	default:
		// Basic analysis for other languages
		issues = t.basicAnalysis(code)
	}

	result := NewToolResult(fmt.Sprintf("Found %d issues", len(issues)))
	result.Data["issues"] = issues
	result.Data["count"] = len(issues)
	result.Data["severity_counts"] = t.countBySeverity(issues)

	return result, nil
}

// AnalysisIssue represents a found issue
type AnalysisIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"` // error, warning, info
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Suggestion string `json:"suggestion,omitempty"`
}

// analyzeGo analyzes Go code
func (t *StaticAnalysisTool) analyzeGo(code string) []AnalysisIssue {
	var issues []AnalysisIssue

	// Parse the code
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example.go", code, parser.AllErrors)
	if err != nil {
		issues = append(issues, AnalysisIssue{
			Type:     "parse_error",
			Severity: "error",
			Message:  fmt.Sprintf("Parse error: %v", err),
		})
		return issues
	}

	// Check for issues
	issues = append(issues, t.checkGoIssues(f, fset)...)
	issues = append(issues, t.checkGoStyle(code)...)

	return issues
}

// checkGoIssues checks for Go-specific issues
func (t *StaticAnalysisTool) checkGoIssues(f *ast.File, fset *token.FileSet) []AnalysisIssue {
	var issues []AnalysisIssue

	// Check for empty function bodies
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Body != nil && len(x.Body.List) == 0 {
				pos := fset.Position(x.Pos())
				issues = append(issues, AnalysisIssue{
					Type:       "empty_function",
					Severity:   "warning",
					Message:    fmt.Sprintf("Function %s has empty body", x.Name.Name),
					Line:       pos.Line,
					Column:     pos.Column,
					Suggestion: "Implement the function or remove it",
				})
			}

			// Check for missing error handling
			if x.Body != nil {
				for _, stmt := range x.Body.List {
					if assign, ok := stmt.(*ast.AssignStmt); ok {
						for _, expr := range assign.Lhs {
							if ident, ok := expr.(*ast.Ident); ok && ident.Name == "err" {
								// Check if error is used
								// This is a simplified check
								pos := fset.Position(assign.Pos())
								issues = append(issues, AnalysisIssue{
									Type:       "error_unchecked",
									Severity:   "warning",
									Message:    "Error value assigned but not checked",
									Line:       pos.Line,
									Column:     pos.Column,
									Suggestion: "Add error handling with if err != nil",
								})
							}
						}
					}
				}
			}
		}
		return true
	})

	return issues
}

// checkGoStyle checks Go style issues
func (t *StaticAnalysisTool) checkGoStyle(code string) []AnalysisIssue {
	var issues []AnalysisIssue
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// Check line length
		if len(line) > 120 {
			issues = append(issues, AnalysisIssue{
				Type:       "long_line",
				Severity:   "info",
				Message:    fmt.Sprintf("Line exceeds 120 characters (%d)", len(line)),
				Line:       lineNum,
				Column:     1,
				Suggestion: "Break line into multiple lines",
			})
		}

		// Check for TODO without context
		if strings.Contains(line, "TODO") && !strings.Contains(line, "TODO(") {
			issues = append(issues, AnalysisIssue{
				Type:       "todo_format",
				Severity:   "info",
				Message:    "TODO should include author or issue reference",
				Line:       lineNum,
				Column:     1,
				Suggestion: "Use TODO(username) or TODO(#issue)",
			})
		}
	}

	return issues
}

// basicAnalysis performs basic analysis for any language
func (t *StaticAnalysisTool) basicAnalysis(code string) []AnalysisIssue {
	var issues []AnalysisIssue
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// Check for potential security issues
		securityPatterns := map[string]string{
			"password":     "Potential hardcoded password",
			"secret":       "Potential hardcoded secret",
			"api_key":      "Potential hardcoded API key",
			"TODO: remove": "TODO marker for removal",
			"FIXME":        "FIXME marker found",
			"HACK":         "HACK marker found",
			"XXX":          "XXX marker found",
		}

		lowerLine := strings.ToLower(line)
		for pattern, message := range securityPatterns {
			if strings.Contains(lowerLine, pattern) {
				issues = append(issues, AnalysisIssue{
					Type:     "marker",
					Severity: "warning",
					Message:  message,
					Line:     lineNum,
					Column:   1,
				})
			}
		}
	}

	return issues
}

// countBySeverity counts issues by severity
func (t *StaticAnalysisTool) countBySeverity(issues []AnalysisIssue) map[string]int {
	counts := map[string]int{
		"error":   0,
		"warning": 0,
		"info":    0,
	}

	for _, issue := range issues {
		counts[issue.Severity]++
	}

	return counts
}

// ComplexityTool analyzes code complexity
type ComplexityTool struct {
	logger *logrus.Logger
}

// NewComplexityTool creates a new complexity tool
func NewComplexityTool(logger *logrus.Logger) *ComplexityTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &ComplexityTool{
		logger: logger,
	}
}

// GetName returns the tool name
func (t *ComplexityTool) GetName() string {
	return "complexity"
}

// GetType returns the tool type
func (t *ComplexityTool) GetType() ToolType {
	return ToolTypeAnalysis
}

// GetDescription returns the description
func (t *ComplexityTool) GetDescription() string {
	return "Analyze code complexity metrics"
}

// GetInputSchema returns the input schema
func (t *ComplexityTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Source code to analyze",
			},
		},
		"required": []string{"code"},
	}
}

// Validate validates inputs
func (t *ComplexityTool) Validate(inputs map[string]interface{}) error {
	if _, ok := inputs["code"].(string); !ok {
		return fmt.Errorf("code is required")
	}
	return nil
}

// Execute executes the tool
func (t *ComplexityTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	code := inputs["code"].(string)

	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	// Count statements (simplified)
	statementCount := strings.Count(code, ";")
	if statementCount == 0 {
		statementCount = strings.Count(code, "\n")
	}

	// Count branches
	ifCount := strings.Count(code, "if ")
	forCount := strings.Count(code, "for ")
	switchCount := strings.Count(code, "switch ")

	// Calculate cyclomatic complexity (simplified)
	cyclomatic := 1 + ifCount + forCount + switchCount

	metrics := map[string]interface{}{
		"lines_of_code":         lineCount,
		"statement_count":       statementCount,
		"if_count":              ifCount,
		"for_count":             forCount,
		"switch_count":          switchCount,
		"cyclomatic_complexity": cyclomatic,
	}

	// Determine complexity level
	level := "low"
	if cyclomatic > 10 {
		level = "medium"
	}
	if cyclomatic > 20 {
		level = "high"
	}
	if cyclomatic > 50 {
		level = "very_high"
	}

	result := NewToolResult(fmt.Sprintf("Complexity: %s (cyclomatic: %d)", level, cyclomatic))
	result.Data["metrics"] = metrics
	result.Data["level"] = level

	return result, nil
}

// LintTool runs linters
type LintTool struct {
	logger *logrus.Logger
}

// NewLintTool creates a new lint tool
func NewLintTool(logger *logrus.Logger) *LintTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &LintTool{
		logger: logger,
	}
}

// GetName returns the tool name
func (t *LintTool) GetName() string {
	return "lint"
}

// GetType returns the tool type
func (t *LintTool) GetType() ToolType {
	return ToolTypeAnalysis
}

// GetDescription returns the description
func (t *LintTool) GetDescription() string {
	return "Run code linters"
}

// GetInputSchema returns the input schema
func (t *LintTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Source code to lint",
			},
		},
		"required": []string{"code"},
	}
}

// Validate validates inputs
func (t *LintTool) Validate(inputs map[string]interface{}) error {
	if _, ok := inputs["code"].(string); !ok {
		return fmt.Errorf("code is required")
	}
	return nil
}

// Execute executes the tool
func (t *LintTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	code := inputs["code"].(string)

	var violations []string

	// Check for common linting issues
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// Check trailing whitespace
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			violations = append(violations, fmt.Sprintf("Line %d: Trailing whitespace", lineNum))
		}

		// Check for tabs (should use spaces in many languages)
		if strings.Contains(line, "\t") {
			violations = append(violations, fmt.Sprintf("Line %d: Contains tabs", lineNum))
		}
	}

	result := NewToolResult(fmt.Sprintf("Found %d lint violations", len(violations)))
	result.Data["violations"] = violations
	result.Data["count"] = len(violations)

	return result, nil
}
