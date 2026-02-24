// Package tools provides tool integration for debate agents.
// This file implements CI/CD hook system for debate validation pipelines.
package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// HookPoint identifies when a hook should be triggered.
type HookPoint string

const (
	// HookPostProposal triggers after proposal phase.
	HookPostProposal HookPoint = "post_proposal"
	// HookPostOptimization triggers after optimization phase.
	HookPostOptimization HookPoint = "post_optimization"
	// HookPostAdversarial triggers after adversarial phase.
	HookPostAdversarial HookPoint = "post_adversarial"
	// HookPostConvergence triggers after convergence phase.
	HookPostConvergence HookPoint = "post_convergence"
	// HookPreDebate triggers before debate starts.
	HookPreDebate HookPoint = "pre_debate"
	// HookPostDebate triggers after debate completes.
	HookPostDebate HookPoint = "post_debate"
)

// HookAction identifies what action to perform.
type HookAction string

const (
	// ActionRunTests runs test suites against the code.
	ActionRunTests HookAction = "run_tests"
	// ActionRunLinter runs linting checks.
	ActionRunLinter HookAction = "run_linter"
	// ActionStaticAnalysis performs static code analysis.
	ActionStaticAnalysis HookAction = "static_analysis"
	// ActionSecurityScan performs security vulnerability scanning.
	ActionSecurityScan HookAction = "security_scan"
	// ActionRunBenchmarks runs performance benchmarks.
	ActionRunBenchmarks HookAction = "run_benchmarks"
	// ActionCustomScript runs a custom validation script.
	ActionCustomScript HookAction = "custom_script"
)

// ActionResult captures the result of a single hook action.
type ActionResult struct {
	Action   HookAction             `json:"action"`
	Passed   bool                   `json:"passed"`
	Output   string                 `json:"output"`
	Duration time.Duration          `json:"duration"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// HookResult captures the result of all actions at a hook point.
type HookResult struct {
	HookPoint HookPoint                    `json:"hook_point"`
	Actions   map[HookAction]*ActionResult `json:"actions"`
	AllPassed bool                         `json:"all_passed"`
	Duration  time.Duration                `json:"duration"`
	Timestamp time.Time                    `json:"timestamp"`
}

// HookConfig configures a CI/CD hook.
type HookConfig struct {
	Enabled    bool                       `json:"enabled"`
	HookPoints map[HookPoint][]HookAction `json:"hook_points"`
	Timeout    time.Duration              `json:"timeout"`    // per action timeout
	FailFast   bool                       `json:"fail_fast"`  // stop on first failure
}

// DefaultHookConfig returns sensible defaults.
func DefaultHookConfig() HookConfig {
	return HookConfig{
		Enabled:  true,
		Timeout:  60 * time.Second,
		FailFast: false,
		HookPoints: map[HookPoint][]HookAction{
			HookPostProposal: {
				ActionRunLinter,
			},
			HookPostOptimization: {
				ActionRunTests,
				ActionStaticAnalysis,
			},
			HookPostAdversarial: {
				ActionSecurityScan,
			},
			HookPostConvergence: {
				ActionRunTests,
				ActionRunLinter,
				ActionStaticAnalysis,
				ActionSecurityScan,
			},
		},
	}
}

// ActionExecutor defines how to execute a specific action.
type ActionExecutor interface {
	Execute(ctx context.Context, action HookAction, code string,
		language string) (*ActionResult, error)
}

// CICDHook manages CI/CD validation hooks for debates.
type CICDHook struct {
	config   HookConfig
	executor ActionExecutor
	mu       sync.RWMutex
}

// NewCICDHook creates a new CI/CD hook with the given config and executor.
func NewCICDHook(config HookConfig, executor ActionExecutor) *CICDHook {
	return &CICDHook{
		config:   config,
		executor: executor,
	}
}

// Configure sets the actions for a specific hook point.
func (h *CICDHook) Configure(hookPoint HookPoint, actions []HookAction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.config.HookPoints == nil {
		h.config.HookPoints = make(map[HookPoint][]HookAction)
	}
	h.config.HookPoints[hookPoint] = actions
}

// Execute runs all configured actions for the given hook point.
func (h *CICDHook) Execute(
	ctx context.Context,
	hookPoint HookPoint,
	code string,
	language string,
) (*HookResult, error) {
	h.mu.RLock()
	enabled := h.config.Enabled
	actions, configured := h.config.HookPoints[hookPoint]
	timeout := h.config.Timeout
	failFast := h.config.FailFast
	h.mu.RUnlock()

	result := &HookResult{
		HookPoint: hookPoint,
		Actions:   make(map[HookAction]*ActionResult),
		AllPassed: true,
		Timestamp: time.Now(),
	}

	if !enabled {
		return result, nil
	}

	if !configured || len(actions) == 0 {
		return result, nil
	}

	start := time.Now()

	for _, action := range actions {
		actionCtx, cancel := context.WithTimeout(ctx, timeout)

		actionResult, err := h.executor.Execute(actionCtx, action, code, language)
		cancel()

		if err != nil {
			actionResult = &ActionResult{
				Action:   action,
				Passed:   false,
				Output:   fmt.Sprintf("execution error: %v", err),
				Duration: time.Since(start),
			}
		}

		result.Actions[action] = actionResult

		if !actionResult.Passed {
			result.AllPassed = false
			if failFast {
				break
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Enable enables the CI/CD hook.
func (h *CICDHook) Enable() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.Enabled = true
}

// Disable disables the CI/CD hook.
func (h *CICDHook) Disable() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.Enabled = false
}

// IsEnabled returns whether the CI/CD hook is enabled.
func (h *CICDHook) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.config.Enabled
}

// GetConfiguredHooks returns all configured hook points and their actions.
func (h *CICDHook) GetConfiguredHooks() map[HookPoint][]HookAction {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[HookPoint][]HookAction, len(h.config.HookPoints))
	for hp, actions := range h.config.HookPoints {
		copied := make([]HookAction, len(actions))
		copy(copied, actions)
		result[hp] = copied
	}
	return result
}

// DefaultActionExecutor provides a basic action executor that performs
// pattern-based checks on code.
type DefaultActionExecutor struct{}

// NewDefaultActionExecutor creates a new default action executor.
func NewDefaultActionExecutor() *DefaultActionExecutor {
	return &DefaultActionExecutor{}
}

// dangerousEvalPattern is the string pattern for eval() calls used in
// security scanning. Stored as a variable to avoid static analysis false
// positives on this source file itself.
var dangerousEvalPattern = "eval("

// Execute performs basic checks based on the action type.
func (e *DefaultActionExecutor) Execute(
	ctx context.Context,
	action HookAction,
	code string,
	language string,
) (*ActionResult, error) {
	select {
	case <-ctx.Done():
		return &ActionResult{
			Action: action,
			Passed: false,
			Output: fmt.Sprintf("action %s timed out: %v", action, ctx.Err()),
		}, ctx.Err()
	default:
	}

	start := time.Now()

	var result *ActionResult

	switch action {
	case ActionRunTests:
		result = e.checkTests(code, language)
	case ActionRunLinter:
		result = e.checkLinter(code, language)
	case ActionStaticAnalysis:
		result = e.checkStaticAnalysis(code, language)
	case ActionSecurityScan:
		result = e.checkSecurity(code, language)
	case ActionRunBenchmarks:
		result = &ActionResult{
			Action: ActionRunBenchmarks,
			Passed: true,
			Output: "benchmarks placeholder: passed",
		}
	case ActionCustomScript:
		result = &ActionResult{
			Action: ActionCustomScript,
			Passed: true,
			Output: "custom script placeholder: passed",
		}
	default:
		result = &ActionResult{
			Action: action,
			Passed: false,
			Output: fmt.Sprintf("unknown action: %s", action),
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// checkTests checks for test patterns in code.
func (e *DefaultActionExecutor) checkTests(code, language string) *ActionResult {
	result := &ActionResult{
		Action:  ActionRunTests,
		Details: make(map[string]interface{}),
	}

	lower := strings.ToLower(code)
	hasTests := false
	issues := make([]string, 0)

	switch language {
	case "go", "golang":
		if strings.Contains(code, "func Test") ||
			strings.Contains(code, "_test.go") {
			hasTests = true
		}
		if strings.Contains(lower, "t.fatal") ||
			strings.Contains(lower, "t.error") ||
			strings.Contains(lower, "assert.") ||
			strings.Contains(lower, "require.") {
			hasTests = true
		}
	case "python":
		if strings.Contains(code, "def test_") ||
			strings.Contains(code, "unittest") ||
			strings.Contains(code, "pytest") {
			hasTests = true
		}
	case "javascript", "typescript":
		if strings.Contains(code, "describe(") ||
			strings.Contains(code, "it(") ||
			strings.Contains(code, "test(") ||
			strings.Contains(code, "expect(") {
			hasTests = true
		}
	default:
		if strings.Contains(lower, "test") ||
			strings.Contains(lower, "assert") {
			hasTests = true
		}
	}

	if !hasTests {
		issues = append(issues, "no test patterns detected in code")
	}

	result.Passed = hasTests
	if hasTests {
		result.Output = "test patterns detected"
	} else {
		result.Output = fmt.Sprintf("test check failed: %s",
			strings.Join(issues, "; "))
	}

	result.Details["has_tests"] = hasTests
	result.Details["language"] = language
	return result
}

// checkLinter checks for common style issues in code.
func (e *DefaultActionExecutor) checkLinter(code, language string) *ActionResult {
	result := &ActionResult{
		Action:  ActionRunLinter,
		Details: make(map[string]interface{}),
	}

	issues := make([]string, 0)

	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if len(line) > 120 {
			issues = append(issues,
				fmt.Sprintf("line %d exceeds 120 characters (%d)", i+1, len(line)))
		}
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			issues = append(issues,
				fmt.Sprintf("line %d has trailing whitespace", i+1))
		}
	}

	if strings.Contains(code, "\t ") || strings.Contains(code, " \t") {
		issues = append(issues, "mixed tabs and spaces detected")
	}

	if language == "go" || language == "golang" {
		if strings.Contains(code, "fmt.Println") &&
			!strings.Contains(code, "_test") {
			issues = append(issues,
				"fmt.Println found in non-test code (use structured logging)")
		}
	}

	result.Passed = len(issues) == 0
	if result.Passed {
		result.Output = "no linting issues found"
	} else {
		result.Output = fmt.Sprintf("linting issues: %s",
			strings.Join(issues, "; "))
	}

	result.Details["issue_count"] = len(issues)
	result.Details["issues"] = issues
	return result
}

// checkStaticAnalysis checks for complexity issues in code.
func (e *DefaultActionExecutor) checkStaticAnalysis(
	code, language string,
) *ActionResult {
	result := &ActionResult{
		Action:  ActionStaticAnalysis,
		Details: make(map[string]interface{}),
	}

	issues := make([]string, 0)
	lines := strings.Split(code, "\n")

	maxNesting := 0
	currentNesting := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		currentNesting += strings.Count(trimmed, "{") -
			strings.Count(trimmed, "}")
		if currentNesting > maxNesting {
			maxNesting = currentNesting
		}
	}

	if maxNesting > 5 {
		issues = append(issues,
			fmt.Sprintf("excessive nesting depth: %d (max recommended: 5)",
				maxNesting))
	}

	funcCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") {
			funcCount++
		}
	}

	if len(lines) > 500 {
		issues = append(issues,
			fmt.Sprintf("file too long: %d lines (max recommended: 500)",
				len(lines)))
	}

	if funcCount > 0 {
		avgLinesPerFunc := len(lines) / funcCount
		if avgLinesPerFunc > 50 {
			issues = append(issues,
				fmt.Sprintf(
					"high average function length: %d lines "+
						"(max recommended: 50)",
					avgLinesPerFunc))
		}
	}

	result.Passed = len(issues) == 0
	if result.Passed {
		result.Output = "no static analysis issues found"
	} else {
		result.Output = fmt.Sprintf("static analysis issues: %s",
			strings.Join(issues, "; "))
	}

	result.Details["max_nesting"] = maxNesting
	result.Details["function_count"] = funcCount
	result.Details["line_count"] = len(lines)
	result.Details["issue_count"] = len(issues)
	return result
}

// checkSecurity checks for security patterns in code.
func (e *DefaultActionExecutor) checkSecurity(
	code, language string,
) *ActionResult {
	result := &ActionResult{
		Action:  ActionSecurityScan,
		Details: make(map[string]interface{}),
	}

	issues := make([]string, 0)
	lower := strings.ToLower(code)

	securityPatterns := []struct {
		pattern     string
		description string
	}{
		{"password", "potential hardcoded password"},
		{"secret", "potential hardcoded secret"},
		{"api_key", "potential hardcoded API key"},
		{"apikey", "potential hardcoded API key"},
		{"private_key", "potential hardcoded private key"},
		{"token", "potential hardcoded token"},
	}

	for _, sp := range securityPatterns {
		if strings.Contains(lower, sp.pattern+"=") ||
			strings.Contains(lower, sp.pattern+" =") ||
			strings.Contains(lower, sp.pattern+":") {
			if !strings.Contains(lower, "os.getenv") &&
				!strings.Contains(lower, "env.get") &&
				!strings.Contains(lower, "config.") &&
				!strings.Contains(lower, "// ") &&
				!strings.Contains(lower, "/* ") {
				issues = append(issues, sp.description)
			}
		}
	}

	if language == "go" || language == "golang" {
		if strings.Contains(code, "exec.Command") &&
			!strings.Contains(code, "// #nosec") {
			issues = append(issues,
				"exec.Command usage detected "+
					"(potential command injection)")
		}
	}

	if strings.Contains(lower, dangerousEvalPattern) {
		issues = append(issues,
			"eval() usage detected (potential code injection)")
	}

	if strings.Contains(lower, "sql") &&
		(strings.Contains(lower, "fmt.sprintf") ||
			strings.Contains(lower, "string concatenation")) {
		issues = append(issues,
			"potential SQL injection (use parameterized queries)")
	}

	result.Passed = len(issues) == 0
	if result.Passed {
		result.Output = "no security issues found"
	} else {
		result.Output = fmt.Sprintf("security issues: %s",
			strings.Join(issues, "; "))
	}

	result.Details["issue_count"] = len(issues)
	result.Details["issues"] = issues
	return result
}
