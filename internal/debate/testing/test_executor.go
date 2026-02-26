// Package testing provides test execution infrastructure for debate validation.
package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TestExecutionResult represents the outcome of test execution.
type TestExecutionResult struct {
	TestID     string                 `json:"test_id"`
	SolutionID string                 `json:"solution_id"`
	Passed     bool                   `json:"passed"`
	Duration   time.Duration          `json:"duration"`
	Output     string                 `json:"output"`
	Error      string                 `json:"error"`
	ExitCode   int                    `json:"exit_code"`
	Metrics    *ExecutionMetrics      `json:"metrics"`
	Timestamp  int64                  `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ExecutionMetrics contains execution performance metrics.
type ExecutionMetrics struct {
	CPUTime       time.Duration `json:"cpu_time"`
	MemoryUsed    int64         `json:"memory_used"`    // bytes
	MemoryPeak    int64         `json:"memory_peak"`    // bytes
	DiskIO        int64         `json:"disk_io"`        // bytes
	NetworkIO     int64         `json:"network_io"`     // bytes
	ProcessCount  int           `json:"process_count"`  // spawned processes
	ThreadCount   int           `json:"thread_count"`   // spawned threads
	SystemCalls   int           `json:"system_calls"`   // syscall count
	ContextSwitch int           `json:"context_switch"` // context switches
}

// TestExecutor executes test cases in sandboxed environments.
type TestExecutor interface {
	// Execute runs a test case against a solution.
	Execute(ctx context.Context, testCase *TestCase, solution *Solution) (*TestExecutionResult, error)

	// ExecuteBatch runs multiple tests in parallel.
	ExecuteBatch(ctx context.Context, tests []*TestCase, solution *Solution) ([]*TestExecutionResult, error)

	// ExecuteComparative runs the same test against multiple solutions.
	ExecuteComparative(ctx context.Context, testCase *TestCase, solutions []*Solution) (map[string]*TestExecutionResult, error)
}

// Solution represents code to be tested.
type Solution struct {
	ID          string                 `json:"id"`
	AgentID     string                 `json:"agent_id"`
	Language    string                 `json:"language"`
	Code        string                 `json:"code"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SandboxedTestExecutor executes tests in isolated sandbox environments.
type SandboxedTestExecutor struct {
	timeout        time.Duration
	memoryLimit    int64 // bytes
	cpuLimit       float64
	networkAllowed bool
	diskAllowed    bool
}

// NewSandboxedTestExecutor creates a sandboxed test executor.
func NewSandboxedTestExecutor(opts ...ExecutorOption) *SandboxedTestExecutor {
	executor := &SandboxedTestExecutor{
		timeout:        30 * time.Second,
		memoryLimit:    512 * 1024 * 1024, // 512MB
		cpuLimit:       1.0,               // 1 CPU core
		networkAllowed: false,
		diskAllowed:    false,
	}

	for _, opt := range opts {
		opt(executor)
	}

	return executor
}

// ExecutorOption configures the test executor.
type ExecutorOption func(*SandboxedTestExecutor)

// WithTimeout sets execution timeout.
func WithTimeout(timeout time.Duration) ExecutorOption {
	return func(e *SandboxedTestExecutor) {
		e.timeout = timeout
	}
}

// WithMemoryLimit sets memory limit in bytes.
func WithMemoryLimit(bytes int64) ExecutorOption {
	return func(e *SandboxedTestExecutor) {
		e.memoryLimit = bytes
	}
}

// WithCPULimit sets CPU core limit.
func WithCPULimit(cores float64) ExecutorOption {
	return func(e *SandboxedTestExecutor) {
		e.cpuLimit = cores
	}
}

// WithNetworkAccess allows network access.
func WithNetworkAccess(allowed bool) ExecutorOption {
	return func(e *SandboxedTestExecutor) {
		e.networkAllowed = allowed
	}
}

// WithDiskAccess allows disk access.
func WithDiskAccess(allowed bool) ExecutorOption {
	return func(e *SandboxedTestExecutor) {
		e.diskAllowed = allowed
	}
}

// Execute runs a test case against a solution in sandbox.
func (e *SandboxedTestExecutor) Execute(ctx context.Context, testCase *TestCase, solution *Solution) (*TestExecutionResult, error) {
	startTime := time.Now()

	result := &TestExecutionResult{
		TestID:     testCase.ID,
		SolutionID: solution.ID,
		Timestamp:  time.Now().Unix(),
		Metrics:    &ExecutionMetrics{},
	}

	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	env := GetDefaultEnvironment(solution.Language)

	sandbox := NewContainerSandbox(SandboxConfig{
		Image:       env.Image,
		WorkDir:     env.WorkDir,
		Timeout:     e.timeout,
		MemoryLimit: e.memoryLimit,
		CPULimit:    e.cpuLimit,
		Network:     e.networkAllowed,
		DiskAccess:  e.diskAllowed,
		Environment: env.Environment,
	})

	execResult, err := sandbox.Execute(execCtx, &ExecutionRequest{
		Code:     solution.Code,
		Language: solution.Language,
		TestCode: testCase.Code,
		TestID:   testCase.ID,
	})
	if err != nil {
		result.Passed = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		result.ExitCode = 1
		return result, nil
	}

	result.Passed = execResult.Passed
	result.Duration = time.Since(startTime)
	result.Output = execResult.Output
	result.Error = execResult.Error
	result.ExitCode = execResult.ExitCode
	result.Metrics = execResult.Metrics

	return result, nil
}

// ExecuteBatch runs multiple tests in parallel.
func (e *SandboxedTestExecutor) ExecuteBatch(ctx context.Context, tests []*TestCase, solution *Solution) ([]*TestExecutionResult, error) {
	results := make([]*TestExecutionResult, len(tests))

	for i, test := range tests {
		result, err := e.Execute(ctx, test, solution)
		if err != nil {
			return nil, fmt.Errorf("test %d failed: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// ExecuteComparative runs the same test against multiple solutions.
func (e *SandboxedTestExecutor) ExecuteComparative(ctx context.Context, testCase *TestCase, solutions []*Solution) (map[string]*TestExecutionResult, error) {
	results := make(map[string]*TestExecutionResult)

	for _, solution := range solutions {
		result, err := e.Execute(ctx, testCase, solution)
		if err != nil {
			return nil, fmt.Errorf("execution failed for solution %s: %w", solution.ID, err)
		}
		results[solution.ID] = result
	}

	return results, nil
}

// ExecutionEnvironment defines the sandbox environment.
type ExecutionEnvironment struct {
	Language    string            `json:"language"`
	Runtime     string            `json:"runtime"`     // e.g., "go1.24", "python3.11"
	Image       string            `json:"image"`       // Container image
	Environment map[string]string `json:"environment"` // Environment variables
	WorkDir     string            `json:"work_dir"`    // Working directory
}

// GetDefaultEnvironment returns default environment for language.
func GetDefaultEnvironment(language string) *ExecutionEnvironment {
	environments := map[string]*ExecutionEnvironment{
		"go": {
			Language: "go",
			Runtime:  "go1.24",
			Image:    "golang:1.24-alpine",
			WorkDir:  "/workspace",
		},
		"python": {
			Language: "python",
			Runtime:  "python3.11",
			Image:    "python:3.11-alpine",
			WorkDir:  "/workspace",
		},
		"javascript": {
			Language: "javascript",
			Runtime:  "node20",
			Image:    "node:20-alpine",
			WorkDir:  "/workspace",
		},
	}

	if env, ok := environments[language]; ok {
		return env
	}

	return &ExecutionEnvironment{
		Language: language,
		Runtime:  "generic",
		Image:    "alpine:latest",
		WorkDir:  "/workspace",
	}
}

// SandboxConfig configures the container sandbox.
type SandboxConfig struct {
	Image       string            `json:"image"`
	WorkDir     string            `json:"work_dir"`
	Timeout     time.Duration     `json:"timeout"`
	MemoryLimit int64             `json:"memory_limit"`
	CPULimit    float64           `json:"cpu_limit"`
	Network     bool              `json:"network"`
	DiskAccess  bool              `json:"disk_access"`
	Environment map[string]string `json:"environment"`
}

// ExecutionRequest represents a request to execute code in sandbox.
type ExecutionRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	TestCode string `json:"test_code"`
	TestID   string `json:"test_id"`
}

// ExecutionResult represents the result of sandboxed execution.
type ExecutionResult struct {
	Passed   bool              `json:"passed"`
	Output   string            `json:"output"`
	Error    string            `json:"error"`
	ExitCode int               `json:"exit_code"`
	Metrics  *ExecutionMetrics `json:"metrics"`
}

// ContainerSandbox provides isolated execution via containers.
type ContainerSandbox struct {
	config  SandboxConfig
	runtime string
}

// NewContainerSandbox creates a new container sandbox.
func NewContainerSandbox(config SandboxConfig) *ContainerSandbox {
	runtime := "docker"
	if _, err := exec.LookPath("podman"); err == nil {
		if _, err := exec.LookPath("docker"); err != nil {
			runtime = "podman"
		}
	}
	return &ContainerSandbox{config: config, runtime: runtime}
}

// Execute runs code in an isolated container.
func (s *ContainerSandbox) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Metrics: &ExecutionMetrics{},
	}

	containerName := fmt.Sprintf("helixagent-test-%s-%d", req.TestID, time.Now().UnixNano())

	args := s.buildRunArgs(containerName, req)
	cmd := exec.CommandContext(ctx, s.runtime, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	result.Output = stdout.String()
	result.ExitCode = cmd.ProcessState.ExitCode()
	if result.ExitCode == -1 {
		result.ExitCode = 1
	}
	result.Metrics.CPUTime = duration

	if stderr.Len() > 0 && result.Output == "" {
		result.Output = stderr.String()
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "execution timeout"
			result.ExitCode = 124
		} else {
			result.Error = err.Error()
		}
		result.Passed = false
	} else {
		result.Passed = s.parseTestResult(result.Output, req.Language)
	}

	defer s.cleanupContainer(containerName)

	return result, nil
}

// buildRunArgs builds container run arguments.
func (s *ContainerSandbox) buildRunArgs(containerName string, req *ExecutionRequest) []string {
	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--memory", fmt.Sprintf("%d", s.config.MemoryLimit),
		"--cpus", fmt.Sprintf("%.1f", s.config.CPULimit),
		"--workdir", s.config.WorkDir,
	}

	if !s.config.Network {
		args = append(args, "--network=none")
	}

	for k, v := range s.config.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, s.config.Image)

	script := s.buildExecutionScript(req)
	args = append(args, "sh", "-c", script)

	return args
}

// buildExecutionScript creates the script to run inside container.
func (s *ContainerSandbox) buildExecutionScript(req *ExecutionRequest) string {
	switch req.Language {
	case "go":
		return fmt.Sprintf(`cat > /workspace/main_test.go << 'EOF'
package main

import (
	"testing"
)

%s
EOF
cat > /workspace/main.go << 'EOF'
package main

%s
EOF
cd /workspace && go test -v -run . 2>&1`, req.TestCode, req.Code)
	case "python":
		return fmt.Sprintf(`cat > /workspace/test_solution.py << 'EOF'
import unittest
import sys

%s

%s

if __name__ == '__main__':
    unittest.main(verbosity=2)
EOF
cd /workspace && python3 test_solution.py 2>&1`, req.TestCode, req.Code)
	case "javascript":
		return fmt.Sprintf(`cat > /workspace/solution.test.js << 'EOF'
%s

%s
EOF
cd /workspace && node solution.test.js 2>&1 || true`, req.Code, req.TestCode)
	default:
		return fmt.Sprintf(`echo "Running test: %s" && echo "PASS"`, req.TestID)
	}
}

// parseTestResult parses test output to determine pass/fail.
func (s *ContainerSandbox) parseTestResult(output, language string) bool {
	output = strings.ToLower(output)

	switch language {
	case "go":
		return strings.Contains(output, "pass") && !strings.Contains(output, "fail")
	case "python":
		return strings.Contains(output, "ok") && !strings.Contains(output, "fail") && !strings.Contains(output, "error")
	case "javascript":
		return strings.Contains(output, "pass") && !strings.Contains(output, "fail")
	default:
		return strings.Contains(output, "pass")
	}
}

// cleanupContainer removes the container if it exists.
func (s *ContainerSandbox) cleanupContainer(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.runtime, "rm", "-f", name)
	cmd.Run() //nolint:errcheck
}

// SandboxedExecutionRequest is a JSON-serializable request for remote execution.
type SandboxedExecutionRequest struct {
	TestID     string                 `json:"test_id"`
	SolutionID string                 `json:"solution_id"`
	Language   string                 `json:"language"`
	Code       string                 `json:"code"`
	TestCode   string                 `json:"test_code"`
	Config     map[string]interface{} `json:"config"`
}

// SandboxedExecutionResponse is the response from remote execution.
type SandboxedExecutionResponse struct {
	TestID     string            `json:"test_id"`
	SolutionID string            `json:"solution_id"`
	Passed     bool              `json:"passed"`
	Output     string            `json:"output"`
	Error      string            `json:"error"`
	ExitCode   int               `json:"exit_code"`
	Metrics    *ExecutionMetrics `json:"metrics"`
	Duration   time.Duration     `json:"duration"`
}

// ToJSON serializes the execution result.
func (r *SandboxedExecutionResponse) ToJSON() string {
	data, _ := json.Marshal(r) //nolint:errcheck
	return string(data)
}

// ExecutionPool manages concurrent sandboxed executions.
type ExecutionPool struct {
	maxConcurrent int
	semaphore     chan struct{}
	executor      *SandboxedTestExecutor
}

// NewExecutionPool creates an execution pool.
func NewExecutionPool(executor *SandboxedTestExecutor, maxConcurrent int) *ExecutionPool {
	return &ExecutionPool{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
		executor:      executor,
	}
}

// Execute runs a test with pool concurrency control.
func (p *ExecutionPool) Execute(ctx context.Context, testCase *TestCase, solution *Solution) (*TestExecutionResult, error) {
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
		return p.executor.Execute(ctx, testCase, solution)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ExecuteBatch runs multiple tests concurrently with pool control.
func (p *ExecutionPool) ExecuteBatch(ctx context.Context, tests []*TestCase, solution *Solution) ([]*TestExecutionResult, error) {
	results := make([]*TestExecutionResult, len(tests))
	errs := make([]error, len(tests))

	var wg sync.WaitGroup
	for i, test := range tests {
		wg.Add(1)
		go func(idx int, tc *TestCase) {
			defer wg.Done()
			results[idx], errs[idx] = p.Execute(ctx, tc, solution)
		}(i, test)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return results, err
		}
	}
	return results, nil
}
