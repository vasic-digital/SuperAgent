package comprehensive

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// SecurityTool scans for security vulnerabilities
type SecurityTool struct {
	logger *logrus.Logger
}

// NewSecurityTool creates a new security tool
func NewSecurityTool(logger *logrus.Logger) *SecurityTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &SecurityTool{
		logger: logger,
	}
}

// GetName returns the tool name
func (t *SecurityTool) GetName() string {
	return "security_scan"
}

// GetType returns the tool type
func (t *SecurityTool) GetType() ToolType {
	return ToolTypeSecurity
}

// GetDescription returns the description
func (t *SecurityTool) GetDescription() string {
	return "Scan code for security vulnerabilities"
}

// GetInputSchema returns the input schema
func (t *SecurityTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Source code to scan",
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Programming language",
				"default":     "go",
			},
		},
		"required": []string{"code"},
	}
}

// Validate validates inputs
func (t *SecurityTool) Validate(inputs map[string]interface{}) error {
	if _, ok := inputs["code"].(string); !ok {
		return fmt.Errorf("code is required")
	}
	return nil
}

// Execute executes the tool
func (t *SecurityTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	code := inputs["code"].(string)

	var vulnerabilities []SecurityVulnerability

	// Check for common security issues
	vulnerabilities = append(vulnerabilities, t.checkHardcodedSecrets(code)...)
	vulnerabilities = append(vulnerabilities, t.checkSQLInjection(code)...)
	vulnerabilities = append(vulnerabilities, t.checkCommandInjection(code)...)
	vulnerabilities = append(vulnerabilities, t.checkPathTraversal(code)...)
	vulnerabilities = append(vulnerabilities, t.checkInsecureRandom(code)...)

	result := NewToolResult(fmt.Sprintf("Found %d vulnerabilities", len(vulnerabilities)))
	result.Data["vulnerabilities"] = vulnerabilities
	result.Data["count"] = len(vulnerabilities)
	result.Data["severity_counts"] = t.countBySeverity(vulnerabilities)

	return result, nil
}

// SecurityVulnerability represents a found vulnerability
type SecurityVulnerability struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"` // critical, high, medium, low
	Message     string `json:"message"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	CWE         string `json:"cwe,omitempty"`
	Remediation string `json:"remediation"`
}

// checkHardcodedSecrets checks for hardcoded secrets
func (t *SecurityTool) checkHardcodedSecrets(code string) []SecurityVulnerability {
	var vulns []SecurityVulnerability
	lines := strings.Split(code, "\n")

	// Patterns to check
	patterns := map[string]string{
		"password":    "Potential hardcoded password",
		"secret":      "Potential hardcoded secret",
		"api_key":     "Potential hardcoded API key",
		"apikey":      "Potential hardcoded API key",
		"token":       "Potential hardcoded token",
		"private_key": "Potential hardcoded private key",
	}

	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		for pattern, message := range patterns {
			if strings.Contains(lowerLine, pattern) {
				// Check if it looks like an assignment
				if strings.Contains(line, "=") || strings.Contains(line, ":") {
					vulns = append(vulns, SecurityVulnerability{
						Type:        "hardcoded_secret",
						Severity:    "critical",
						Message:     message,
						Line:        i + 1,
						Column:      1,
						CWE:         "CWE-798",
						Remediation: "Use environment variables or secure secret management",
					})
				}
			}
		}
	}

	return vulns
}

// checkSQLInjection checks for SQL injection vulnerabilities
func (t *SecurityTool) checkSQLInjection(code string) []SecurityVulnerability {
	var vulns []SecurityVulnerability
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		// Check for string concatenation in SQL
		if strings.Contains(line, "SELECT") || strings.Contains(line, "INSERT") ||
			strings.Contains(line, "UPDATE") || strings.Contains(line, "DELETE") {
			if strings.Contains(line, "+") || strings.Contains(line, "fmt.Sprintf") {
				vulns = append(vulns, SecurityVulnerability{
					Type:        "sql_injection",
					Severity:    "critical",
					Message:     "Potential SQL injection via string concatenation",
					Line:        i + 1,
					Column:      1,
					CWE:         "CWE-89",
					Remediation: "Use parameterized queries or prepared statements",
				})
			}
		}
	}

	return vulns
}

// checkCommandInjection checks for command injection
func (t *SecurityTool) checkCommandInjection(code string) []SecurityVulnerability {
	var vulns []SecurityVulnerability
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		// Check for exec.Command with user input
		if strings.Contains(line, "exec.Command") || strings.Contains(line, "os/exec") {
			if strings.Contains(line, "+") || strings.Contains(line, "fmt.Sprintf") {
				vulns = append(vulns, SecurityVulnerability{
					Type:        "command_injection",
					Severity:    "critical",
					Message:     "Potential command injection via string concatenation",
					Line:        i + 1,
					Column:      1,
					CWE:         "CWE-78",
					Remediation: "Use parameterized commands or validate inputs strictly",
				})
			}
		}
	}

	return vulns
}

// checkPathTraversal checks for path traversal
func (t *SecurityTool) checkPathTraversal(code string) []SecurityVulnerability {
	var vulns []SecurityVulnerability
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		// Check for file operations with user input
		if strings.Contains(line, "os.Open") || strings.Contains(line, "os.Create") ||
			strings.Contains(line, "os.ReadFile") {
			if strings.Contains(line, "+") || strings.Contains(line, "fmt.Sprintf") {
				vulns = append(vulns, SecurityVulnerability{
					Type:        "path_traversal",
					Severity:    "high",
					Message:     "Potential path traversal via string concatenation",
					Line:        i + 1,
					Column:      1,
					CWE:         "CWE-22",
					Remediation: "Validate and sanitize file paths",
				})
			}
		}
	}

	return vulns
}

// checkInsecureRandom checks for insecure random number generation
func (t *SecurityTool) checkInsecureRandom(code string) []SecurityVulnerability {
	var vulns []SecurityVulnerability
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		// Check for math/rand usage (should use crypto/rand for security)
		if strings.Contains(line, "math/rand") {
			vulns = append(vulns, SecurityVulnerability{
				Type:        "insecure_random",
				Severity:    "medium",
				Message:     "Using math/rand for security-sensitive operations",
				Line:        i + 1,
				Column:      1,
				CWE:         "CWE-338",
				Remediation: "Use crypto/rand for security-sensitive random number generation",
			})
		}
	}

	return vulns
}

// countBySeverity counts vulnerabilities by severity
func (t *SecurityTool) countBySeverity(vulns []SecurityVulnerability) map[string]int {
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, v := range vulns {
		counts[v.Severity]++
	}

	return counts
}

// PerformanceTool analyzes performance characteristics
type PerformanceTool struct {
	logger *logrus.Logger
}

// NewPerformanceTool creates a new performance tool
func NewPerformanceTool(logger *logrus.Logger) *PerformanceTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &PerformanceTool{
		logger: logger,
	}
}

// GetName returns the tool name
func (t *PerformanceTool) GetName() string {
	return "performance_profile"
}

// GetType returns the tool type
func (t *PerformanceTool) GetType() ToolType {
	return ToolTypePerformance
}

// GetDescription returns the description
func (t *PerformanceTool) GetDescription() string {
	return "Analyze code for performance issues"
}

// GetInputSchema returns the input schema
func (t *PerformanceTool) GetInputSchema() map[string]interface{} {
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
func (t *PerformanceTool) Validate(inputs map[string]interface{}) error {
	if _, ok := inputs["code"].(string); !ok {
		return fmt.Errorf("code is required")
	}
	return nil
}

// Execute executes the tool
func (t *PerformanceTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	code := inputs["code"].(string)

	var issues []PerformanceIssue

	// Check for performance issues
	issues = append(issues, t.checkInefficientLoops(code)...)
	issues = append(issues, t.checkMemoryAllocations(code)...)
	issues = append(issues, t.checkStringConcatenation(code)...)
	issues = append(issues, t.checkRecursion(code)...)

	result := NewToolResult(fmt.Sprintf("Found %d performance issues", len(issues)))
	result.Data["issues"] = issues
	result.Data["count"] = len(issues)

	return result, nil
}

// PerformanceIssue represents a performance issue
type PerformanceIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Suggestion string `json:"suggestion"`
}

// checkInefficientLoops checks for inefficient loop patterns
func (t *PerformanceTool) checkInefficientLoops(code string) []PerformanceIssue {
	var issues []PerformanceIssue
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		// Check for range over large arrays
		if strings.Contains(line, "range") && strings.Contains(line, "[") {
			issues = append(issues, PerformanceIssue{
				Type:       "inefficient_loop",
				Severity:   "medium",
				Message:    "Range over array may cause unnecessary copying",
				Line:       i + 1,
				Suggestion: "Use range over slice or pointer",
			})
		}
	}

	return issues
}

// checkMemoryAllocations checks for excessive memory allocations
func (t *PerformanceTool) checkMemoryAllocations(code string) []PerformanceIssue {
	var issues []PerformanceIssue
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		// Check for make() in loops
		if strings.Contains(line, "for") || strings.Contains(line, "range") {
			// Check next few lines for make
			for j := i + 1; j < len(lines) && j < i+5; j++ {
				if strings.Contains(lines[j], "make(") {
					issues = append(issues, PerformanceIssue{
						Type:       "allocation_in_loop",
						Severity:   "high",
						Message:    "Memory allocation inside loop",
						Line:       j + 1,
						Suggestion: "Pre-allocate memory outside the loop",
					})
				}
			}
		}
	}

	return issues
}

// checkStringConcatenation checks for inefficient string concatenation
func (t *PerformanceTool) checkStringConcatenation(code string) []PerformanceIssue {
	var issues []PerformanceIssue
	lines := strings.Split(code, "\n")

	concatCount := 0
	concatStart := 0

	for i, line := range lines {
		if strings.Contains(line, "+=") && strings.Contains(line, "\"") {
			if concatCount == 0 {
				concatStart = i
			}
			concatCount++

			if concatCount > 3 {
				issues = append(issues, PerformanceIssue{
					Type:       "string_concatenation",
					Severity:   "medium",
					Message:    "Multiple string concatenations detected",
					Line:       concatStart + 1,
					Suggestion: "Use strings.Builder for efficient string concatenation",
				})
				concatCount = 0
			}
		} else {
			concatCount = 0
		}
	}

	return issues
}

// checkRecursion checks for potentially expensive recursion
func (t *PerformanceTool) checkRecursion(code string) []PerformanceIssue {
	var issues []PerformanceIssue
	lines := strings.Split(code, "\n")

	// Find function definitions
	for i, line := range lines {
		if strings.Contains(line, "func ") {
			// Extract function name (simplified)
			parts := strings.Fields(line)
			for j, part := range parts {
				if part == "func" && j+1 < len(parts) {
					funcName := strings.TrimSuffix(parts[j+1], "(")

					// Check if function calls itself
					for k := i + 1; k < len(lines); k++ {
						if strings.Contains(lines[k], funcName+"(") {
							issues = append(issues, PerformanceIssue{
								Type:       "recursion",
								Severity:   "info",
								Message:    fmt.Sprintf("Function %s uses recursion", funcName),
								Line:       i + 1,
								Suggestion: "Consider iterative approach for large inputs",
							})
							break
						}
						if strings.Contains(lines[k], "func ") {
							break // End of function
						}
					}
				}
			}
		}
	}

	return issues
}
