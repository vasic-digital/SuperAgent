package comprehensive

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToolResult_NewToolResult(t *testing.T) {
	result := NewToolResult("Success output")

	assert.True(t, result.Success)
	assert.Equal(t, "Success output", result.Output)
	assert.Empty(t, result.Error)
	assert.NotNil(t, result.Data)
	assert.NotZero(t, result.Timestamp)
}

func TestToolResult_NewToolError(t *testing.T) {
	result := NewToolError("Something went wrong")

	assert.False(t, result.Success)
	assert.Empty(t, result.Output)
	assert.Equal(t, "Something went wrong", result.Error)
	assert.NotNil(t, result.Data)
}

func TestToolRegistry_Register(t *testing.T) {
	registry := NewToolRegistry(nil)

	// Create a mock tool
	tool := NewCodeTool(".", nil)

	registry.Register(tool)

	// Verify tool is registered
	retrieved, ok := registry.Get(tool.GetName())
	assert.True(t, ok)
	assert.Equal(t, tool.GetName(), retrieved.GetName())
}

func TestToolRegistry_GetAll(t *testing.T) {
	registry := NewToolRegistry(nil)

	// Register multiple tools
	registry.Register(NewCodeTool(".", nil))
	registry.Register(NewCommandTool(".", 30*time.Second, nil))
	registry.Register(NewTestTool(".", nil))

	all := registry.GetAll()
	assert.Len(t, all, 3)
}

func TestToolRegistry_GetByType(t *testing.T) {
	registry := NewToolRegistry(nil)

	// Register tools of different types
	registry.Register(NewCodeTool(".", nil))
	registry.Register(NewSearchTool(".", nil))
	registry.Register(NewCommandTool(".", 30*time.Second, nil))

	// Get code tools
	codeTools := registry.GetByType(ToolTypeCode)
	assert.Len(t, codeTools, 2) // Code and Search

	// Get command tools
	cmdTools := registry.GetByType(ToolTypeCommand)
	assert.Len(t, cmdTools, 1) // Only Command
}

func TestCodeTool_GetName(t *testing.T) {
	tool := NewCodeTool(".", nil)
	assert.Equal(t, "code", tool.GetName())
}

func TestCodeTool_GetType(t *testing.T) {
	tool := NewCodeTool(".", nil)
	assert.Equal(t, ToolTypeCode, tool.GetType())
}

func TestCodeTool_Validate_MissingAction(t *testing.T) {
	tool := NewCodeTool(".", nil)

	inputs := map[string]interface{}{
		"path": "/test/file.go",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is required")
}

func TestCodeTool_Validate_InvalidAction(t *testing.T) {
	tool := NewCodeTool(".", nil)

	inputs := map[string]interface{}{
		"action": "invalid",
		"path":   "/test/file.go",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestCodeTool_Validate_PathTraversal(t *testing.T) {
	tool := NewCodeTool(".", nil)

	inputs := map[string]interface{}{
		"action": "read",
		"path":   "../../../etc/passwd",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal")
}

func TestCodeTool_Validate_MissingContent(t *testing.T) {
	tool := NewCodeTool(".", nil)

	inputs := map[string]interface{}{
		"action": "write",
		"path":   "/test/file.go",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content is required")
}

func TestCommandTool_GetName(t *testing.T) {
	tool := NewCommandTool(".", 30*time.Second, nil)
	assert.Equal(t, "execute_command", tool.GetName())
}

func TestCommandTool_Validate_MissingCommand(t *testing.T) {
	tool := NewCommandTool(".", 30*time.Second, nil)

	inputs := map[string]interface{}{}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command is required")
}

func TestCommandTool_Validate_DangerousCommand(t *testing.T) {
	tool := NewCommandTool(".", 30*time.Second, nil)

	inputs := map[string]interface{}{
		"command": "rm -rf /",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dangerous command")
}

func TestCommandTool_Validate_NotAllowed(t *testing.T) {
	tool := NewCommandTool(".", 30*time.Second, nil)

	inputs := map[string]interface{}{
		"command": "malicious_command",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command not allowed")
}

func TestTestTool_GetName(t *testing.T) {
	tool := NewTestTool(".", nil)
	assert.Equal(t, "run_tests", tool.GetName())
}

func TestBuildTool_GetName(t *testing.T) {
	tool := NewBuildTool(".", nil)
	assert.Equal(t, "build", tool.GetName())
}

func TestStaticAnalysisTool_GetName(t *testing.T) {
	tool := NewStaticAnalysisTool(nil)
	assert.Equal(t, "static_analysis", tool.GetName())
}

func TestStaticAnalysisTool_Validate_MissingCode(t *testing.T) {
	tool := NewStaticAnalysisTool(nil)

	inputs := map[string]interface{}{}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

func TestComplexityTool_GetName(t *testing.T) {
	tool := NewComplexityTool(nil)
	assert.Equal(t, "complexity", tool.GetName())
}

func TestLintTool_GetName(t *testing.T) {
	tool := NewLintTool(nil)
	assert.Equal(t, "lint", tool.GetName())
}

func TestSecurityTool_GetName(t *testing.T) {
	tool := NewSecurityTool(nil)
	assert.Equal(t, "security_scan", tool.GetName())
}

func TestPerformanceTool_GetName(t *testing.T) {
	tool := NewPerformanceTool(nil)
	assert.Equal(t, "performance_profile", tool.GetName())
}

func TestSecurityTool_checkHardcodedSecrets(t *testing.T) {
	tool := NewSecurityTool(nil)

	code := `
		password := "secret123"
		api_key := "sk-abc123"
	`

	vulns := tool.checkHardcodedSecrets(code)
	assert.NotEmpty(t, vulns)

	// Should find password vulnerability
	foundPassword := false
	for _, v := range vulns {
		if v.Type == "hardcoded_secret" {
			foundPassword = true
			assert.Equal(t, "critical", v.Severity)
			assert.Equal(t, "CWE-798", v.CWE)
		}
	}
	assert.True(t, foundPassword)
}

func TestSecurityTool_checkSQLInjection(t *testing.T) {
	tool := NewSecurityTool(nil)

	code := `
		query := "SELECT * FROM users WHERE id = " + userId
		db.Query(query)
	`

	vulns := tool.checkSQLInjection(code)
	assert.NotEmpty(t, vulns)

	if len(vulns) > 0 {
		assert.Equal(t, "sql_injection", vulns[0].Type)
		assert.Equal(t, "critical", vulns[0].Severity)
		assert.Equal(t, "CWE-89", vulns[0].CWE)
	}
}

func TestSecurityTool_checkCommandInjection(t *testing.T) {
	tool := NewSecurityTool(nil)

	code := `
		cmd := exec.Command("sh", "-c", "ls " + userInput)
	`

	vulns := tool.checkCommandInjection(code)
	assert.NotEmpty(t, vulns)
}

func TestSecurityTool_countBySeverity(t *testing.T) {
	tool := NewSecurityTool(nil)

	vulns := []SecurityVulnerability{
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "high"},
		{Severity: "medium"},
	}

	counts := tool.countBySeverity(vulns)
	assert.Equal(t, 2, counts["critical"])
	assert.Equal(t, 1, counts["high"])
	assert.Equal(t, 1, counts["medium"])
	assert.Equal(t, 0, counts["low"])
}

func TestPerformanceTool_checkMemoryAllocations(t *testing.T) {
	tool := NewPerformanceTool(nil)

	code := `
		for i := 0; i < 100; i++ {
			slice := make([]int, 100)
			_ = slice
		}
	`

	issues := tool.checkMemoryAllocations(code)
	assert.NotEmpty(t, issues)
}

func TestPerformanceTool_checkStringConcatenation(t *testing.T) {
	tool := NewPerformanceTool(nil)

	code := `
		result := ""
		result += "line1"
		result += "line2"
		result += "line3"
		result += "line4"
	`

	issues := tool.checkStringConcatenation(code)
	assert.NotEmpty(t, issues)
}

func TestComplexityTool_calculateCyclomatic(t *testing.T) {
	tool := NewComplexityTool(nil)

	inputs := map[string]interface{}{
		"code": `
			if x > 0 {
				for i := 0; i < 10; i++ {
					if i % 2 == 0 {
						switch x {
						case 1: return 1
						case 2: return 2
						}
					}
				}
			}
		`,
	}

	ctx := context.Background()
	result, err := tool.Execute(ctx, inputs)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Data["metrics"])
}

func TestLintTool_checkTrailingWhitespace(t *testing.T) {
	tool := NewLintTool(nil)

	inputs := map[string]interface{}{
		"code": "line1 \nline2\nline3 \t",
	}

	ctx := context.Background()
	result, err := tool.Execute(ctx, inputs)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	violations, ok := result.Data["violations"].([]string)
	if ok {
		assert.NotEmpty(t, violations)
	}
}

func TestDatabaseTool_Validate_Readonly(t *testing.T) {
	// Create tool with readonly=true
	tool := NewDatabaseTool(nil, true, nil)

	inputs := map[string]interface{}{
		"query": "INSERT INTO users VALUES (1, 'test')",
	}

	err := tool.Validate(inputs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write operations not allowed")
}

func TestDebateRequestValidator_Validate_MissingID(t *testing.T) {
	validator := DebateRequestValidator{}

	req := &DebateRequest{
		Topic: "Test topic",
	}

	errors := validator.Validate(req)
	assert.NotEmpty(t, errors)

	found := false
	for _, e := range errors {
		if e.Type == "missing_id" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestDebateRequestValidator_Validate_MissingTopic(t *testing.T) {
	validator := DebateRequestValidator{}

	req := &DebateRequest{
		ID: "debate-1",
	}

	errors := validator.Validate(req)
	assert.NotEmpty(t, errors)
}

func TestDebateRequestValidator_Validate_InvalidRounds(t *testing.T) {
	validator := DebateRequestValidator{}

	req := &DebateRequest{
		ID:        "debate-1",
		Topic:     "Test",
		MaxRounds: 0,
	}

	errors := validator.Validate(req)

	found := false
	for _, e := range errors {
		if e.Type == "invalid_rounds" {
			found = true
		}
	}
	assert.True(t, found)
}
