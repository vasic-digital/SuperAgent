package comprehensive

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// CommandTool executes shell commands
type CommandTool struct {
	allowedCommands []string
	workDir         string
	timeout         time.Duration
	logger          *logrus.Logger
}

// NewCommandTool creates a new command tool
func NewCommandTool(workDir string, timeout time.Duration, logger *logrus.Logger) *CommandTool {
	if logger == nil {
		logger = logrus.New()
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &CommandTool{
		allowedCommands: []string{
			"go", "git", "make", "docker", "ls", "cat", "grep",
			"find", "wc", "head", "tail", "mkdir", "rm", "cp", "mv",
			"pwd", "echo", "test", "bash", "sh",
		},
		workDir: workDir,
		timeout: timeout,
		logger:  logger,
	}
}

// GetName returns the tool name
func (t *CommandTool) GetName() string {
	return "execute_command"
}

// GetType returns the tool type
func (t *CommandTool) GetType() ToolType {
	return ToolTypeCommand
}

// GetDescription returns the description
func (t *CommandTool) GetDescription() string {
	return "Execute shell commands in a sandboxed environment"
}

// GetInputSchema returns the input schema
func (t *CommandTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Command to execute",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for command execution",
			},
			"timeout": map[string]interface{}{
				"type":        "number",
				"description": "Timeout in seconds (default: 30)",
			},
		},
		"required": []string{"command"},
	}
}

// Validate validates inputs
func (t *CommandTool) Validate(inputs map[string]interface{}) error {
	command, ok := inputs["command"].(string)
	if !ok || command == "" {
		return fmt.Errorf("command is required")
	}

	// Check for dangerous commands
	dangerous := []string{"rm -rf /", "> /dev/sda", "mkfs", "dd if=/dev/zero"}
	for _, d := range dangerous {
		if strings.Contains(command, d) {
			return fmt.Errorf("dangerous command detected: %s", d)
		}
	}

	// Check if command is allowed
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := parts[0]
	allowed := false
	for _, allowedCmd := range t.allowedCommands {
		if cmd == allowedCmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("command not allowed: %s", cmd)
	}

	return nil
}

// Execute executes the tool
func (t *CommandTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	command := inputs["command"].(string)

	workDir := t.workDir
	if wd, ok := inputs["working_dir"].(string); ok && wd != "" {
		workDir = wd
	}

	timeout := t.timeout
	if to, ok := inputs["timeout"].(float64); ok && to > 0 {
		timeout = time.Duration(to) * time.Second
	}

	t.logger.WithFields(logrus.Fields{
		"command": command,
		"workDir": workDir,
	}).Debug("Executing command")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &ToolResult{
		Timestamp: time.Now(),
		Duration:  duration,
		Data:      make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("command failed: %v", err)
		result.Output = stderr.String()
		result.Data["exit_code"] = getExitCode(err)
	} else {
		result.Success = true
		result.Output = stdout.String()
		result.Data["exit_code"] = 0
	}

	result.Data["stdout"] = stdout.String()
	result.Data["stderr"] = stderr.String()
	result.Data["command"] = command

	return result, nil
}

// getExitCode extracts exit code from error
func getExitCode(err error) int {
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1
}

// TestTool runs tests
type TestTool struct {
	workDir string
	logger  *logrus.Logger
}

// NewTestTool creates a new test tool
func NewTestTool(workDir string, logger *logrus.Logger) *TestTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &TestTool{
		workDir: workDir,
		logger:  logger,
	}
}

// GetName returns the tool name
func (t *TestTool) GetName() string {
	return "run_tests"
}

// GetType returns the tool type
func (t *TestTool) GetType() ToolType {
	return ToolTypeCommand
}

// GetDescription returns the description
func (t *TestTool) GetDescription() string {
	return "Run Go tests and analyze results"
}

// GetInputSchema returns the input schema
func (t *TestTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"package": map[string]interface{}{
				"type":        "string",
				"description": "Package to test (default: ./...)",
			},
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable verbose output",
			},
			"race": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable race detector",
			},
			"coverage": map[string]interface{}{
				"type":        "boolean",
				"description": "Generate coverage report",
			},
		},
	}
}

// Validate validates inputs
func (t *TestTool) Validate(inputs map[string]interface{}) error {
	return nil
}

// Execute executes the tool
func (t *TestTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	pkg := "./..."
	if p, ok := inputs["package"].(string); ok && p != "" {
		pkg = p
	}

	verbose := false
	if v, ok := inputs["verbose"].(bool); ok {
		verbose = v
	}

	race := false
	if r, ok := inputs["race"].(bool); ok {
		race = r
	}

	coverage := false
	if c, ok := inputs["coverage"].(bool); ok {
		coverage = c
	}

	// Build command
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if race {
		args = append(args, "-race")
	}
	if coverage {
		args = append(args, "-cover")
	}
	args = append(args, pkg)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = t.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &ToolResult{
		Timestamp: time.Now(),
		Duration:  duration,
		Data:      make(map[string]interface{}),
	}

	output := stdout.String() + stderr.String()

	if err != nil {
		result.Success = false
		result.Error = "tests failed"
		result.Output = output
		result.Data["exit_code"] = getExitCode(err)
	} else {
		result.Success = true
		result.Output = output
		result.Data["exit_code"] = 0
	}

	// Parse test results
	result.Data["passed"] = strings.Count(output, "PASS")
	result.Data["failed"] = strings.Count(output, "FAIL")

	// Parse coverage if enabled
	if coverage {
		// Extract coverage percentage
		if idx := strings.Index(output, "coverage:"); idx != -1 {
			parts := strings.Fields(output[idx:])
			if len(parts) > 1 {
				result.Data["coverage"] = parts[1]
			}
		}
	}

	return result, nil
}

// BuildTool builds Go code
type BuildTool struct {
	workDir string
	logger  *logrus.Logger
}

// NewBuildTool creates a new build tool
func NewBuildTool(workDir string, logger *logrus.Logger) *BuildTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &BuildTool{
		workDir: workDir,
		logger:  logger,
	}
}

// GetName returns the tool name
func (t *BuildTool) GetName() string {
	return "build"
}

// GetType returns the tool type
func (t *BuildTool) GetType() ToolType {
	return ToolTypeCommand
}

// GetDescription returns the description
func (t *BuildTool) GetDescription() string {
	return "Build Go code and check for compilation errors"
}

// GetInputSchema returns the input schema
func (t *BuildTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"package": map[string]interface{}{
				"type":        "string",
				"description": "Package to build",
			},
			"output": map[string]interface{}{
				"type":        "string",
				"description": "Output binary name",
			},
		},
	}
}

// Validate validates inputs
func (t *BuildTool) Validate(inputs map[string]interface{}) error {
	return nil
}

// Execute executes the tool
func (t *BuildTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	pkg := "."
	if p, ok := inputs["package"].(string); ok && p != "" {
		pkg = p
	}

	args := []string{"build"}
	if output, ok := inputs["output"].(string); ok && output != "" {
		args = append(args, "-o", output)
	}
	args = append(args, pkg)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = t.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &ToolResult{
		Timestamp: time.Now(),
		Duration:  duration,
		Data:      make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = "build failed"
		result.Output = stderr.String()
		result.Data["exit_code"] = getExitCode(err)
	} else {
		result.Success = true
		result.Output = "Build successful"
		if stdout.Len() > 0 {
			result.Output = stdout.String()
		}
		result.Data["exit_code"] = 0
	}

	return result, nil
}
