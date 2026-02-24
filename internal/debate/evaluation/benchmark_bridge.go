// Package evaluation provides benchmark evaluation for debate outcomes.
// It bridges debate results to benchmark scoring systems (SWE-bench, HumanEval,
// MMLU, custom) and performs static code analysis for quality metrics.
package evaluation

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// BenchmarkType identifies the type of benchmark.
type BenchmarkType string

const (
	// BenchmarkSWEBench represents the SWE-bench benchmark for software engineering tasks.
	BenchmarkSWEBench BenchmarkType = "swe_bench"
	// BenchmarkHumanEval represents the HumanEval benchmark for code generation.
	BenchmarkHumanEval BenchmarkType = "human_eval"
	// BenchmarkMMLU represents the MMLU benchmark for multitask language understanding.
	BenchmarkMMLU BenchmarkType = "mmlu"
	// BenchmarkCustom represents a custom benchmark with static analysis scoring.
	BenchmarkCustom BenchmarkType = "custom"
)

// EvaluationScore captures the result of evaluating a debate outcome.
type EvaluationScore struct {
	BenchmarkType BenchmarkType          `json:"benchmark_type"`
	Score         float64                `json:"score"`              // 0-1
	Details       map[string]float64     `json:"details"`            // per-metric scores
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BenchmarkProblem represents a single benchmark problem.
type BenchmarkProblem struct {
	ID          string        `json:"id"`
	Type        BenchmarkType `json:"type"`
	Description string        `json:"description"`
	TestCases   []string      `json:"test_cases"`
	Expected    string        `json:"expected"`
	Language    string        `json:"language"`
	Difficulty  string        `json:"difficulty"` // easy, medium, hard
}

// Evaluator is the interface for benchmark-specific evaluation.
type Evaluator interface {
	Evaluate(ctx context.Context, solution string, problem *BenchmarkProblem) (*EvaluationScore, error)
}

// DebateResultForEval is a simplified view of a debate result for evaluation.
type DebateResultForEval struct {
	ID            string                 `json:"id"`
	Topic         string                 `json:"topic"`
	FinalSolution string                 `json:"final_solution"`
	Consensus     float64                `json:"consensus"`
	Language      string                 `json:"language"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BenchmarkBridge connects debate results to benchmark evaluation.
type BenchmarkBridge struct {
	evaluators map[BenchmarkType]Evaluator
	mu         sync.RWMutex
}

// NewBenchmarkBridge creates a new BenchmarkBridge with a default custom evaluator
// registered for the BenchmarkCustom type.
func NewBenchmarkBridge() *BenchmarkBridge {
	b := &BenchmarkBridge{
		evaluators: make(map[BenchmarkType]Evaluator),
	}
	b.evaluators[BenchmarkCustom] = &customEvaluator{bridge: b}
	return b
}

// RegisterEvaluator registers an evaluator for a specific benchmark type.
func (b *BenchmarkBridge) RegisterEvaluator(benchmarkType BenchmarkType, evaluator Evaluator) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.evaluators[benchmarkType] = evaluator
}

// EvaluateDebateResult evaluates a debate result against a benchmark.
// It looks up the evaluator for the given benchmark type, constructs a temporary
// BenchmarkProblem from the debate result, and delegates to the evaluator.
// If no evaluator is registered for the requested type, the custom evaluator is used.
func (b *BenchmarkBridge) EvaluateDebateResult(
	ctx context.Context,
	result *DebateResultForEval,
	benchmarkType BenchmarkType,
) (*EvaluationScore, error) {
	if result == nil {
		return nil, fmt.Errorf("debate result cannot be nil")
	}

	b.mu.RLock()
	evaluator, ok := b.evaluators[benchmarkType]
	if !ok {
		evaluator = b.evaluators[BenchmarkCustom]
	}
	b.mu.RUnlock()

	if evaluator == nil {
		return nil, fmt.Errorf("no evaluator available for benchmark type: %s", benchmarkType)
	}

	problem := &BenchmarkProblem{
		ID:          result.ID,
		Type:        benchmarkType,
		Description: result.Topic,
		Language:    result.Language,
		Difficulty:  "medium",
	}

	score, err := evaluator.Evaluate(ctx, result.FinalSolution, problem)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed for benchmark %s: %w", benchmarkType, err)
	}

	// Inject consensus into metadata if available.
	if score.Metadata == nil {
		score.Metadata = make(map[string]interface{})
	}
	score.Metadata["debate_id"] = result.ID
	score.Metadata["consensus"] = result.Consensus

	return score, nil
}

// CalculateCustomMetrics performs static analysis on a code solution and returns
// metric scores in the range [0, 1] for five dimensions:
//   - correctness: base 0.7, adjusted by consensus when available via context
//   - maintainability: comments ratio, function length heuristics, naming conventions
//   - performance: detection of common performance anti-patterns
//   - security: detection of common security patterns
//   - test_coverage: presence of test code indicators
func (b *BenchmarkBridge) CalculateCustomMetrics(
	solution string,
	language string,
) (map[string]float64, error) {
	if solution == "" {
		return map[string]float64{
			"correctness":     0.0,
			"maintainability": 0.0,
			"performance":     0.0,
			"security":        0.0,
			"test_coverage":   0.0,
		}, nil
	}

	metrics := make(map[string]float64, 5)

	metrics["correctness"] = calculateCorrectness(solution, language)
	metrics["maintainability"] = calculateMaintainability(solution, language)
	metrics["performance"] = calculatePerformance(solution, language)
	metrics["security"] = calculateSecurity(solution, language)
	metrics["test_coverage"] = calculateTestCoverage(solution, language)

	return metrics, nil
}

// calculateCorrectness returns a correctness score.
// Base is 0.7 since we cannot truly verify correctness through static analysis.
// Adjustments are made based on structural indicators.
func calculateCorrectness(solution, language string) float64 {
	score := 0.7
	lower := strings.ToLower(solution)
	lines := strings.Split(solution, "\n")

	// Bonus for error handling presence.
	if containsErrorHandling(lower, language) {
		score += 0.05
	}

	// Bonus for having return statements (indicates complete logic).
	if strings.Contains(lower, "return") {
		score += 0.03
	}

	// Bonus for having imports (indicates real code).
	if containsImports(lower, language) {
		score += 0.02
	}

	// Penalty for very short solutions (likely incomplete).
	if len(lines) < 5 {
		score -= 0.1
	}

	// Penalty for TODO/FIXME markers (indicates incomplete work).
	todoCount := strings.Count(lower, "todo") + strings.Count(lower, "fixme")
	score -= float64(todoCount) * 0.05

	return clampScore(score)
}

// calculateMaintainability evaluates code maintainability through:
// - comment ratio (target: 10-30% of lines)
// - function length (prefer shorter functions)
// - naming conventions (camelCase/PascalCase vs unclear names)
func calculateMaintainability(solution, language string) float64 {
	score := 0.5
	lines := strings.Split(solution, "\n")
	totalLines := len(lines)
	if totalLines == 0 {
		return 0.0
	}

	// Comment ratio analysis.
	commentLines := countCommentLines(lines, language)
	commentRatio := float64(commentLines) / float64(totalLines)
	switch {
	case commentRatio >= 0.10 && commentRatio <= 0.30:
		score += 0.2 // Ideal comment ratio.
	case commentRatio > 0.05 && commentRatio < 0.10:
		score += 0.1 // Acceptable.
	case commentRatio > 0.30:
		score += 0.05 // Over-commented.
	default:
		score -= 0.1 // Under-commented.
	}

	// Function length heuristics: count functions and check average length.
	funcCount, avgFuncLen := analyzeFunctions(lines, language)
	if funcCount > 0 {
		switch {
		case avgFuncLen <= 20:
			score += 0.15 // Short, focused functions.
		case avgFuncLen <= 40:
			score += 0.1 // Reasonable length.
		case avgFuncLen <= 60:
			score += 0.05 // Getting long.
		default:
			score -= 0.1 // Too long.
		}
	}

	// Naming convention checks.
	if hasGoodNaming(solution, language) {
		score += 0.1
	}

	// Bonus for consistent indentation.
	if hasConsistentIndentation(lines) {
		score += 0.05
	}

	return clampScore(score)
}

// calculatePerformance checks for common performance anti-patterns.
func calculatePerformance(solution, language string) float64 {
	score := 0.8
	lower := strings.ToLower(solution)
	lines := strings.Split(solution, "\n")

	// Check for nested loops (potential O(n^2) or worse).
	nestedLoopCount := countNestedLoops(lines, language)
	score -= float64(nestedLoopCount) * 0.1

	// Check for unnecessary string concatenation in loops.
	if hasStringConcatInLoop(lines, language) {
		score -= 0.1
	}

	// Check for common allocation anti-patterns.
	if hasUnnecessaryAllocations(lower, language) {
		score -= 0.1
	}

	// Bonus for using efficient data structures.
	if usesEfficientStructures(lower, language) {
		score += 0.05
	}

	// Bonus for buffered I/O usage.
	if usesBufferedIO(lower, language) {
		score += 0.05
	}

	return clampScore(score)
}

// calculateSecurity checks for common security patterns and anti-patterns.
func calculateSecurity(solution, language string) float64 {
	score := 0.7
	lower := strings.ToLower(solution)

	// Bonus for input validation patterns.
	if hasInputValidation(lower, language) {
		score += 0.1
	}

	// Bonus for error handling (prevents information leakage).
	if containsErrorHandling(lower, language) {
		score += 0.1
	}

	// Penalty for hardcoded secrets.
	if hasHardcodedSecrets(lower) {
		score -= 0.3
	}

	// Penalty for SQL injection patterns.
	if hasSQLInjectionRisk(lower, language) {
		score -= 0.2
	}

	// Penalty for command injection patterns.
	if hasCommandInjectionRisk(lower, language) {
		score -= 0.15
	}

	// Bonus for use of parameterized queries.
	if usesParameterizedQueries(lower, language) {
		score += 0.1
	}

	// Bonus for proper TLS/crypto usage.
	if usesSecureCrypto(lower, language) {
		score += 0.05
	}

	return clampScore(score)
}

// calculateTestCoverage checks if test code is present in the solution.
func calculateTestCoverage(solution, language string) float64 {
	score := 0.0
	lower := strings.ToLower(solution)

	// Check for test function definitions.
	testIndicators := getTestIndicators(language)
	foundIndicators := 0
	for _, indicator := range testIndicators {
		if strings.Contains(lower, indicator) {
			foundIndicators++
		}
	}

	if len(testIndicators) > 0 {
		score = float64(foundIndicators) / float64(len(testIndicators))
	}

	// Bonus for assertion usage.
	if hasAssertions(lower, language) {
		score += 0.2
	}

	// Bonus for table-driven tests (Go pattern).
	if language == "go" && (strings.Contains(lower, "testcases") ||
		strings.Contains(lower, "tests := []") ||
		strings.Contains(lower, "tt.name")) {
		score += 0.1
	}

	// Bonus for test setup/teardown.
	if hasTestSetupTeardown(lower, language) {
		score += 0.1
	}

	return clampScore(score)
}

// --- Helper functions ---

func containsErrorHandling(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "if err != nil") ||
			(strings.Contains(lower, "error") && strings.Contains(lower, "return"))
	case "python":
		return strings.Contains(lower, "try:") && strings.Contains(lower, "except")
	case "javascript", "typescript":
		return strings.Contains(lower, "try {") || strings.Contains(lower, "catch")
	case "java":
		return strings.Contains(lower, "try {") && strings.Contains(lower, "catch")
	case "rust":
		return strings.Contains(lower, "result<") || strings.Contains(lower, "?;")
	default:
		return strings.Contains(lower, "try") || strings.Contains(lower, "catch") ||
			strings.Contains(lower, "error")
	}
}

func containsImports(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "import")
	case "python":
		return strings.Contains(lower, "import") || strings.Contains(lower, "from ")
	case "javascript", "typescript":
		return strings.Contains(lower, "import ") || strings.Contains(lower, "require(")
	case "java":
		return strings.Contains(lower, "import ")
	case "rust":
		return strings.Contains(lower, "use ")
	default:
		return strings.Contains(lower, "import")
	}
}

func countCommentLines(lines []string, language string) int {
	count := 0
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch language {
		case "go", "java", "javascript", "typescript", "rust", "c", "cpp":
			if inBlock {
				count++
				if strings.Contains(trimmed, "*/") {
					inBlock = false
				}
				continue
			}
			if strings.HasPrefix(trimmed, "//") {
				count++
			} else if strings.HasPrefix(trimmed, "/*") {
				count++
				inBlock = true
				if strings.Contains(trimmed, "*/") {
					inBlock = false
				}
			}
		case "python":
			if strings.HasPrefix(trimmed, "#") {
				count++
			}
			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				count++
				inBlock = !inBlock
			} else if inBlock {
				count++
			}
		default:
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				count++
			}
		}
	}
	return count
}

func analyzeFunctions(lines []string, language string) (int, int) {
	funcCount := 0
	totalFuncLines := 0
	inFunc := false
	currentFuncLines := 0
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch language {
		case "go":
			if strings.HasPrefix(trimmed, "func ") {
				if inFunc {
					totalFuncLines += currentFuncLines
				}
				funcCount++
				inFunc = true
				currentFuncLines = 0
				braceDepth = 0
			}
		case "python":
			if strings.HasPrefix(trimmed, "def ") {
				if inFunc {
					totalFuncLines += currentFuncLines
				}
				funcCount++
				inFunc = true
				currentFuncLines = 0
			}
		case "javascript", "typescript":
			if strings.Contains(trimmed, "function ") ||
				strings.Contains(trimmed, "=>") {
				if inFunc && braceDepth == 0 {
					totalFuncLines += currentFuncLines
				}
				funcCount++
				inFunc = true
				currentFuncLines = 0
				braceDepth = 0
			}
		case "java":
			if (strings.Contains(trimmed, "public ") ||
				strings.Contains(trimmed, "private ") ||
				strings.Contains(trimmed, "protected ")) &&
				strings.Contains(trimmed, "(") &&
				!strings.Contains(trimmed, "class ") {
				if inFunc {
					totalFuncLines += currentFuncLines
				}
				funcCount++
				inFunc = true
				currentFuncLines = 0
				braceDepth = 0
			}
		case "rust":
			if strings.HasPrefix(trimmed, "fn ") ||
				strings.HasPrefix(trimmed, "pub fn ") {
				if inFunc {
					totalFuncLines += currentFuncLines
				}
				funcCount++
				inFunc = true
				currentFuncLines = 0
				braceDepth = 0
			}
		}

		if inFunc {
			currentFuncLines++
			braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if language != "python" && braceDepth <= 0 && currentFuncLines > 1 {
				totalFuncLines += currentFuncLines
				inFunc = false
				currentFuncLines = 0
			}
		}
	}

	// Flush last function if still in one.
	if inFunc && currentFuncLines > 0 {
		totalFuncLines += currentFuncLines
	}

	if funcCount == 0 {
		return 0, 0
	}
	return funcCount, totalFuncLines / funcCount
}

func hasGoodNaming(solution, language string) bool {
	lines := strings.Split(solution, "\n")
	goodNames := 0
	totalNames := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch language {
		case "go":
			// Check variable declarations.
			if strings.Contains(trimmed, ":=") || strings.Contains(trimmed, "var ") {
				totalNames++
				parts := strings.Fields(trimmed)
				if len(parts) > 0 {
					name := parts[0]
					if len(name) > 1 && len(name) < 30 {
						goodNames++
					}
				}
			}
		case "python":
			if strings.Contains(trimmed, "=") && !strings.HasPrefix(trimmed, "#") &&
				!strings.Contains(trimmed, "==") {
				totalNames++
				parts := strings.Split(trimmed, "=")
				if len(parts) > 0 {
					name := strings.TrimSpace(parts[0])
					if len(name) > 1 && len(name) < 30 &&
						!strings.Contains(name, " ") {
						goodNames++
					}
				}
			}
		default:
			if strings.Contains(trimmed, "=") && !strings.HasPrefix(trimmed, "//") {
				totalNames++
				goodNames++ // Generous default.
			}
		}
	}

	if totalNames == 0 {
		return true // No names to judge.
	}
	return float64(goodNames)/float64(totalNames) >= 0.7
}

func hasConsistentIndentation(lines []string) bool {
	if len(lines) < 3 {
		return true
	}

	tabCount := 0
	spaceCount := 0
	for _, line := range lines {
		if len(line) == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "\t") {
			tabCount++
		} else if strings.HasPrefix(line, "  ") {
			spaceCount++
		}
	}

	total := tabCount + spaceCount
	if total == 0 {
		return true
	}
	// Consistent if mostly one style (>80%).
	dominant := tabCount
	if spaceCount > tabCount {
		dominant = spaceCount
	}
	return float64(dominant)/float64(total) >= 0.8
}

func countNestedLoops(lines []string, language string) int {
	nestedCount := 0
	loopDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		isLoop := false
		switch language {
		case "go":
			isLoop = strings.HasPrefix(lower, "for ")
		case "python":
			isLoop = strings.HasPrefix(lower, "for ") ||
				strings.HasPrefix(lower, "while ")
		case "javascript", "typescript":
			isLoop = strings.HasPrefix(lower, "for ") ||
				strings.HasPrefix(lower, "for(") ||
				strings.HasPrefix(lower, "while ")
		case "java":
			isLoop = strings.HasPrefix(lower, "for ") ||
				strings.HasPrefix(lower, "for(") ||
				strings.HasPrefix(lower, "while ")
		default:
			isLoop = strings.HasPrefix(lower, "for ") ||
				strings.HasPrefix(lower, "while ")
		}

		if isLoop {
			loopDepth++
			if loopDepth >= 2 {
				nestedCount++
			}
		}

		// Track brace depth for loop exit (non-Python).
		if language != "python" {
			closeBraces := strings.Count(trimmed, "}")
			for i := 0; i < closeBraces && loopDepth > 0; i++ {
				loopDepth--
			}
		}
	}

	return nestedCount
}

func hasStringConcatInLoop(lines []string, language string) bool {
	inLoop := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(trimmed, "for ") ||
			strings.HasPrefix(trimmed, "while ") {
			inLoop = true
			continue
		}
		if inLoop {
			switch language {
			case "go":
				if strings.Contains(trimmed, "+= ") &&
					strings.Contains(trimmed, "\"") {
					return true
				}
			case "python":
				if strings.Contains(trimmed, "+= ") &&
					(strings.Contains(trimmed, "\"") ||
						strings.Contains(trimmed, "'")) {
					return true
				}
			case "javascript", "typescript":
				if strings.Contains(trimmed, "+= ") &&
					(strings.Contains(trimmed, "\"") ||
						strings.Contains(trimmed, "'") ||
						strings.Contains(trimmed, "`")) {
					return true
				}
			case "java":
				if strings.Contains(trimmed, "+= ") &&
					strings.Contains(trimmed, "\"") {
					return true
				}
			}
		}
		if strings.Contains(trimmed, "}") {
			inLoop = false
		}
	}
	return false
}

func hasUnnecessaryAllocations(lower, language string) bool {
	switch language {
	case "go":
		// Repeated make() calls or append in tight loops without pre-allocation.
		return strings.Count(lower, "make([]") > 3 ||
			(strings.Contains(lower, "for ") &&
				strings.Count(lower, "append(") > 3)
	case "java":
		// Creating objects in loops.
		return strings.Contains(lower, "new ") &&
			strings.Count(lower, "new ") > 5
	default:
		return false
	}
}

func usesEfficientStructures(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "map[") ||
			strings.Contains(lower, "sync.pool") ||
			strings.Contains(lower, "sync.map")
	case "python":
		return strings.Contains(lower, "set(") ||
			strings.Contains(lower, "defaultdict") ||
			strings.Contains(lower, "collections.")
	case "java":
		return strings.Contains(lower, "hashmap") ||
			strings.Contains(lower, "hashset") ||
			strings.Contains(lower, "concurrenthashmap")
	default:
		return strings.Contains(lower, "map") || strings.Contains(lower, "set")
	}
}

func usesBufferedIO(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "bufio.") ||
			strings.Contains(lower, "bytes.buffer")
	case "java":
		return strings.Contains(lower, "bufferedreader") ||
			strings.Contains(lower, "bufferedwriter")
	case "python":
		return strings.Contains(lower, "bufferedreader") ||
			strings.Contains(lower, "io.bytesio")
	default:
		return false
	}
}

func hasInputValidation(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "validate") ||
			strings.Contains(lower, "if len(") ||
			strings.Contains(lower, "strings.trimspace")
	case "python":
		return strings.Contains(lower, "validate") ||
			strings.Contains(lower, "isinstance(") ||
			strings.Contains(lower, "if not ")
	case "javascript", "typescript":
		return strings.Contains(lower, "validate") ||
			strings.Contains(lower, "typeof ") ||
			strings.Contains(lower, "instanceof ")
	case "java":
		return strings.Contains(lower, "validate") ||
			strings.Contains(lower, "objects.requirenonnull") ||
			strings.Contains(lower, "instanceof ")
	default:
		return strings.Contains(lower, "validate") ||
			strings.Contains(lower, "check")
	}
}

func hasHardcodedSecrets(lower string) bool {
	secretPatterns := []string{
		"password = \"",
		"password =\"",
		"api_key = \"",
		"api_key =\"",
		"apikey = \"",
		"apikey =\"",
		"secret = \"",
		"secret =\"",
		"token = \"",
		"token =\"",
		"private_key = \"",
	}
	for _, pattern := range secretPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func hasSQLInjectionRisk(lower, language string) bool {
	// Check for string formatting/concatenation in SQL queries.
	sqlIndicators := []string{"select ", "insert ", "update ", "delete "}
	hasSQLKeyword := false
	for _, kw := range sqlIndicators {
		if strings.Contains(lower, kw) {
			hasSQLKeyword = true
			break
		}
	}
	if !hasSQLKeyword {
		return false
	}

	switch language {
	case "go":
		return strings.Contains(lower, "fmt.sprintf") &&
			strings.Contains(lower, "select")
	case "python":
		return (strings.Contains(lower, "f\"") ||
			strings.Contains(lower, ".format(") ||
			strings.Contains(lower, "% ")) &&
			(strings.Contains(lower, "select") ||
				strings.Contains(lower, "insert"))
	case "java":
		return strings.Contains(lower, "string.format") &&
			(strings.Contains(lower, "select") ||
				strings.Contains(lower, "insert"))
	default:
		return false
	}
}

func hasCommandInjectionRisk(lower, language string) bool {
	switch language {
	case "python":
		return strings.Contains(lower, "os.system(") ||
			(strings.Contains(lower, "subprocess.call(") &&
				strings.Contains(lower, "shell=true"))
	case "javascript", "typescript":
		return strings.Contains(lower, "eval(")
	default:
		return false
	}
}

func usesParameterizedQueries(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "db.query(") ||
			strings.Contains(lower, "db.exec(") ||
			strings.Contains(lower, "db.queryrow(")
	case "python":
		return (strings.Contains(lower, "execute(") &&
			strings.Contains(lower, "%s")) ||
			(strings.Contains(lower, "cursor.execute(") &&
				strings.Contains(lower, "?"))
	case "java":
		return strings.Contains(lower, "preparedstatement") ||
			strings.Contains(lower, "preparestatement")
	default:
		return false
	}
}

func usesSecureCrypto(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "crypto/") ||
			strings.Contains(lower, "crypto.") ||
			strings.Contains(lower, "tls.")
	case "python":
		return strings.Contains(lower, "hashlib") ||
			strings.Contains(lower, "cryptography") ||
			strings.Contains(lower, "ssl.")
	case "java":
		return strings.Contains(lower, "javax.crypto") ||
			strings.Contains(lower, "messagedigest")
	default:
		return strings.Contains(lower, "crypto") || strings.Contains(lower, "tls")
	}
}

func getTestIndicators(language string) []string {
	switch language {
	case "go":
		return []string{"func test", "_test.go", "testing.t", "t.run("}
	case "python":
		return []string{"def test_", "unittest", "pytest", "assert"}
	case "javascript", "typescript":
		return []string{"describe(", "it(", "test(", "expect("}
	case "java":
		return []string{"@test", "assertequals", "asserttrue", "junit"}
	case "rust":
		return []string{"#[test]", "#[cfg(test)]", "assert!", "assert_eq!"}
	default:
		return []string{"test", "assert"}
	}
}

func hasAssertions(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "assert.") ||
			strings.Contains(lower, "require.") ||
			strings.Contains(lower, "t.fatal") ||
			strings.Contains(lower, "t.error")
	case "python":
		return strings.Contains(lower, "assert ") ||
			strings.Contains(lower, "assertequal") ||
			strings.Contains(lower, "asserttrue")
	case "javascript", "typescript":
		return strings.Contains(lower, "expect(") ||
			strings.Contains(lower, "assert(") ||
			strings.Contains(lower, "assert.")
	case "java":
		return strings.Contains(lower, "assertequals") ||
			strings.Contains(lower, "asserttrue") ||
			strings.Contains(lower, "assertthat")
	case "rust":
		return strings.Contains(lower, "assert!") ||
			strings.Contains(lower, "assert_eq!")
	default:
		return strings.Contains(lower, "assert")
	}
}

func hasTestSetupTeardown(lower, language string) bool {
	switch language {
	case "go":
		return strings.Contains(lower, "testmain") ||
			strings.Contains(lower, "t.cleanup")
	case "python":
		return strings.Contains(lower, "setup") ||
			strings.Contains(lower, "teardown") ||
			strings.Contains(lower, "@pytest.fixture")
	case "javascript", "typescript":
		return strings.Contains(lower, "beforeeach") ||
			strings.Contains(lower, "aftereach") ||
			strings.Contains(lower, "beforeall") ||
			strings.Contains(lower, "afterall")
	case "java":
		return strings.Contains(lower, "@before") ||
			strings.Contains(lower, "@after") ||
			strings.Contains(lower, "@beforeeach")
	default:
		return strings.Contains(lower, "setup") ||
			strings.Contains(lower, "teardown")
	}
}

// clampScore clamps a score to the [0, 1] range.
func clampScore(score float64) float64 {
	return math.Max(0.0, math.Min(1.0, score))
}

// --- customEvaluator ---

// customEvaluator implements Evaluator using static analysis via CalculateCustomMetrics.
type customEvaluator struct {
	bridge *BenchmarkBridge
}

// Evaluate performs custom evaluation on a solution by delegating to
// CalculateCustomMetrics and computing an aggregate score from the
// individual metrics.
func (e *customEvaluator) Evaluate(
	ctx context.Context,
	solution string,
	problem *BenchmarkProblem,
) (*EvaluationScore, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	language := problem.Language
	if language == "" {
		language = "go" // Default to Go for HelixAgent context.
	}

	metrics, err := e.bridge.CalculateCustomMetrics(solution, language)
	if err != nil {
		return nil, fmt.Errorf("custom metric calculation failed: %w", err)
	}

	// Weighted aggregate score.
	weights := map[string]float64{
		"correctness":     0.30,
		"maintainability": 0.20,
		"performance":     0.20,
		"security":        0.20,
		"test_coverage":   0.10,
	}

	var aggregate float64
	for metric, weight := range weights {
		if val, ok := metrics[metric]; ok {
			aggregate += val * weight
		}
	}
	aggregate = clampScore(aggregate)

	return &EvaluationScore{
		BenchmarkType: BenchmarkCustom,
		Score:         aggregate,
		Details:       metrics,
		Timestamp:     time.Now(),
		Metadata: map[string]interface{}{
			"language":   language,
			"difficulty": problem.Difficulty,
			"problem_id": problem.ID,
		},
	}, nil
}

// --- DebateBenchmarkSuite ---

// DebateBenchmarkSuite runs a suite of benchmark problems against a debate system.
type DebateBenchmarkSuite struct {
	bridge   *BenchmarkBridge
	problems []*BenchmarkProblem
	mu       sync.RWMutex
}

// NewDebateBenchmarkSuite creates a new suite with the given benchmark bridge.
func NewDebateBenchmarkSuite(bridge *BenchmarkBridge) *DebateBenchmarkSuite {
	return &DebateBenchmarkSuite{
		bridge:   bridge,
		problems: make([]*BenchmarkProblem, 0),
	}
}

// AddProblem adds a benchmark problem to the suite.
func (s *DebateBenchmarkSuite) AddProblem(problem *BenchmarkProblem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.problems = append(s.problems, problem)
}

// GetProblems returns a copy of the problems in the suite.
func (s *DebateBenchmarkSuite) GetProblems() []*BenchmarkProblem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*BenchmarkProblem, len(s.problems))
	copy(result, s.problems)
	return result
}

// Run iterates over all problems in the suite, calls solveFn to obtain a debate
// result for each, and evaluates the result. It returns all evaluation scores.
func (s *DebateBenchmarkSuite) Run(
	ctx context.Context,
	solveFn func(ctx context.Context, problem *BenchmarkProblem) (*DebateResultForEval, error),
) ([]*EvaluationScore, error) {
	s.mu.RLock()
	problems := make([]*BenchmarkProblem, len(s.problems))
	copy(problems, s.problems)
	s.mu.RUnlock()

	if len(problems) == 0 {
		return []*EvaluationScore{}, nil
	}

	scores := make([]*EvaluationScore, 0, len(problems))

	for _, problem := range problems {
		select {
		case <-ctx.Done():
			return scores, ctx.Err()
		default:
		}

		result, err := solveFn(ctx, problem)
		if err != nil {
			return scores, fmt.Errorf(
				"solve failed for problem %s: %w", problem.ID, err,
			)
		}

		if result == nil {
			continue
		}

		score, err := s.bridge.EvaluateDebateResult(ctx, result, problem.Type)
		if err != nil {
			return scores, fmt.Errorf(
				"evaluation failed for problem %s: %w", problem.ID, err,
			)
		}

		scores = append(scores, score)
	}

	return scores, nil
}
