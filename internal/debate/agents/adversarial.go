// Package agents provides specialized agent implementations for AI debates.
// This file implements the Red/Blue Team adversarial dynamics protocol
// for attack-defend cycles that harden code through iterative vulnerability
// discovery and patching.
package agents

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Vulnerability represents a security or correctness issue found by the red team.
type Vulnerability struct {
	ID          string `json:"id"`
	Category    string `json:"category"`    // injection, overflow, race_condition, logic_error, etc.
	Severity    string `json:"severity"`    // critical, high, medium, low
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
	Exploit     string `json:"exploit"` // how to exploit
}

// EdgeCase represents a boundary condition that may cause unexpected behavior.
type EdgeCase struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Input       string `json:"input"`
	Expected    string `json:"expected"`
}

// StressScenario represents a high-load or resource-exhaustion scenario.
type StressScenario struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Load        string `json:"load"`
	Expected    string `json:"expected_behavior"`
}

// AttackReport contains the red team's findings for a single round.
type AttackReport struct {
	Vulnerabilities []Vulnerability  `json:"vulnerabilities"`
	EdgeCases       []EdgeCase       `json:"edge_cases"`
	StressScenarios []StressScenario `json:"stress_scenarios"`
	OverallRisk     float64          `json:"overall_risk"` // 0-1
	Round           int              `json:"round"`
}

// DefenseReport contains the blue team's response and patches for a single round.
type DefenseReport struct {
	PatchedVulnerabilities []string          `json:"patched_vulnerabilities"` // IDs
	Patches                map[string]string `json:"patches"`                 // vuln ID -> fix description
	RemainingRisks         []string          `json:"remaining_risks"`
	ConfidenceInDefense    float64           `json:"confidence_in_defense"`
	PatchedCode            string            `json:"patched_code"`
	Round                  int               `json:"round"`
}

// AdversarialResult holds the outcome of a complete adversarial protocol run.
type AdversarialResult struct {
	Rounds         int              `json:"rounds"`
	FinalCode      string           `json:"final_code"`
	AttackReports  []*AttackReport  `json:"attack_reports"`
	DefenseReports []*DefenseReport `json:"defense_reports"`
	AllResolved    bool             `json:"all_resolved"`
	RemainingRisks []string         `json:"remaining_risks"`
	Duration       time.Duration    `json:"duration"`
}

// AdversarialLLMClient is the interface for LLM interactions used by the
// adversarial protocol. Callers provide their own implementation backed
// by any provider.
type AdversarialLLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// AdversarialConfig tunes the adversarial protocol behavior.
type AdversarialConfig struct {
	MaxRounds          int           `json:"max_rounds"`
	MinVulnerabilities int           `json:"min_vulnerabilities"` // min vulns to continue
	RiskThreshold      float64       `json:"risk_threshold"`      // stop if risk below
	Timeout            time.Duration `json:"timeout"`
}

// AdversarialProtocol orchestrates Red/Blue Team attack-defend cycles.
type AdversarialProtocol struct {
	config    AdversarialConfig
	redTeam   *RedTeamAgent
	blueTeam  *BlueTeamAgent
	llmClient AdversarialLLMClient
}

// RedTeamAgent represents the offensive agent that discovers vulnerabilities.
type RedTeamAgent struct {
	ID   string
	Role string
}

// BlueTeamAgent represents the defensive agent that patches vulnerabilities.
type BlueTeamAgent struct {
	ID   string
	Role string
}

// DefaultAdversarialConfig returns sensible defaults for the adversarial protocol.
func DefaultAdversarialConfig() AdversarialConfig {
	return AdversarialConfig{
		MaxRounds:          3,
		MinVulnerabilities: 1,
		RiskThreshold:      0.2,
		Timeout:            5 * time.Minute,
	}
}

// NewAdversarialProtocol creates a new adversarial protocol with internal
// red and blue team agents.
func NewAdversarialProtocol(
	config AdversarialConfig,
	llmClient AdversarialLLMClient,
) *AdversarialProtocol {
	return &AdversarialProtocol{
		config: config,
		redTeam: &RedTeamAgent{
			ID:   "red-team-agent",
			Role: "attacker",
		},
		blueTeam: &BlueTeamAgent{
			ID:   "blue-team-agent",
			Role: "defender",
		},
		llmClient: llmClient,
	}
}

// Execute runs the full adversarial attack-defend loop on the given solution.
// It returns when no new vulnerabilities are found, risk drops below the
// configured threshold, or the maximum number of rounds is reached.
func (ap *AdversarialProtocol) Execute(
	ctx context.Context,
	solution string,
	language string,
) (*AdversarialResult, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, ap.config.Timeout)
	defer cancel()

	result := &AdversarialResult{
		AttackReports:  make([]*AttackReport, 0, ap.config.MaxRounds),
		DefenseReports: make([]*DefenseReport, 0, ap.config.MaxRounds),
	}

	currentCode := solution
	var lastDefense *DefenseReport

	for round := 1; round <= ap.config.MaxRounds; round++ {
		select {
		case <-ctx.Done():
			result.Rounds = round - 1
			result.FinalCode = currentCode
			result.Duration = time.Since(start)
			result.RemainingRisks = collectRemainingRisks(result)
			result.AllResolved = len(result.RemainingRisks) == 0
			return result, fmt.Errorf("adversarial protocol timed out: %w", ctx.Err())
		default:
		}

		// Red team attacks
		attackReport, err := ap.attack(ctx, currentCode, language, round, lastDefense)
		if err != nil {
			return nil, fmt.Errorf("red team attack failed in round %d: %w", round, err)
		}
		attackReport.Round = round
		result.AttackReports = append(result.AttackReports, attackReport)

		// Check termination: no vulnerabilities found
		if len(attackReport.Vulnerabilities) < ap.config.MinVulnerabilities {
			result.Rounds = round
			result.FinalCode = currentCode
			result.Duration = time.Since(start)
			result.RemainingRisks = collectRemainingRisks(result)
			result.AllResolved = len(result.RemainingRisks) == 0
			return result, nil
		}

		// Check termination: risk below threshold
		if attackReport.OverallRisk < ap.config.RiskThreshold {
			result.Rounds = round
			result.FinalCode = currentCode
			result.Duration = time.Since(start)
			result.RemainingRisks = collectRemainingRisks(result)
			result.AllResolved = len(result.RemainingRisks) == 0
			return result, nil
		}

		// Blue team defends
		defenseReport, err := ap.defend(ctx, currentCode, language, attackReport)
		if err != nil {
			return nil, fmt.Errorf("blue team defense failed in round %d: %w", round, err)
		}
		defenseReport.Round = round
		result.DefenseReports = append(result.DefenseReports, defenseReport)

		// Update current code with patched version
		if defenseReport.PatchedCode != "" {
			currentCode = defenseReport.PatchedCode
		}
		lastDefense = defenseReport
	}

	result.Rounds = ap.config.MaxRounds
	result.FinalCode = currentCode
	result.Duration = time.Since(start)
	result.RemainingRisks = collectRemainingRisks(result)
	result.AllResolved = len(result.RemainingRisks) == 0

	return result, nil
}

// attack instructs the red team to find vulnerabilities in the current solution.
// On LLM failure it returns a deterministic fallback report.
func (ap *AdversarialProtocol) attack(
	ctx context.Context,
	solution string,
	language string,
	round int,
	priorDefense *DefenseReport,
) (*AttackReport, error) {
	prompt := ap.buildAttackPrompt(solution, language, round, priorDefense)

	response, err := ap.llmClient.Complete(ctx, prompt)
	if err != nil {
		// LLM failed — use deterministic fallback
		return ap.generateFallbackAttack(solution, language, round), nil
	}

	report, err := ap.parseAttackResponse(response)
	if err != nil {
		// Parse failed — use deterministic fallback
		return ap.generateFallbackAttack(solution, language, round), nil
	}

	return report, nil
}

// defend instructs the blue team to patch identified vulnerabilities.
// On LLM failure it returns a deterministic fallback defense.
func (ap *AdversarialProtocol) defend(
	ctx context.Context,
	solution string,
	language string,
	attackReport *AttackReport,
) (*DefenseReport, error) {
	prompt := ap.buildDefensePrompt(solution, language, attackReport)

	response, err := ap.llmClient.Complete(ctx, prompt)
	if err != nil {
		// LLM failed — use deterministic fallback
		return ap.generateFallbackDefense(solution, attackReport), nil
	}

	report, err := ap.parseDefenseResponse(response)
	if err != nil {
		// Parse failed — use deterministic fallback
		return ap.generateFallbackDefense(solution, attackReport), nil
	}

	return report, nil
}

// buildAttackPrompt constructs the prompt sent to the LLM for the red team
// attack phase. If a prior defense exists, the prompt incorporates it so the
// red team can attempt to bypass the applied patches.
func (ap *AdversarialProtocol) buildAttackPrompt(
	solution, language string,
	round int,
	priorDefense *DefenseReport,
) string {
	var b strings.Builder

	b.WriteString("You are a Red Team security analyst. ")
	b.WriteString("Analyze the following ")
	b.WriteString(language)
	b.WriteString(" code for vulnerabilities, edge cases, and stress scenarios.\n\n")

	b.WriteString("Round: ")
	b.WriteString(fmt.Sprintf("%d", round))
	b.WriteString("\n\n")

	if priorDefense != nil && len(priorDefense.PatchedVulnerabilities) > 0 {
		b.WriteString("The blue team has applied the following patches:\n")
		for _, id := range priorDefense.PatchedVulnerabilities {
			if desc, ok := priorDefense.Patches[id]; ok {
				b.WriteString("- ")
				b.WriteString(id)
				b.WriteString(": ")
				b.WriteString(desc)
				b.WriteString("\n")
			}
		}
		b.WriteString("\nAttempt to find new vulnerabilities or bypass existing patches.\n\n")
	}

	b.WriteString("Code to analyze:\n```")
	b.WriteString(language)
	b.WriteString("\n")
	b.WriteString(solution)
	b.WriteString("\n```\n\n")

	b.WriteString("Respond with structured output using the exact format below.\n")
	b.WriteString("Use --- to separate sections.\n\n")

	b.WriteString("VULNERABILITIES\n")
	b.WriteString("ID: VULN-001\n")
	b.WriteString("Category: <injection|overflow|race_condition|logic_error|auth|xss|other>\n")
	b.WriteString("Severity: <critical|high|medium|low>\n")
	b.WriteString("Description: <description>\n")
	b.WriteString("Evidence: <evidence from code>\n")
	b.WriteString("Exploit: <how to exploit>\n")
	b.WriteString("---\n\n")

	b.WriteString("EDGE_CASES\n")
	b.WriteString("ID: EDGE-001\n")
	b.WriteString("Description: <description>\n")
	b.WriteString("Input: <input that triggers>\n")
	b.WriteString("Expected: <expected safe behavior>\n")
	b.WriteString("---\n\n")

	b.WriteString("STRESS_SCENARIOS\n")
	b.WriteString("ID: STRESS-001\n")
	b.WriteString("Description: <description>\n")
	b.WriteString("Load: <load description>\n")
	b.WriteString("Expected: <expected behavior under load>\n")
	b.WriteString("---\n\n")

	b.WriteString("OVERALL_RISK: <0.0-1.0>\n")

	return b.String()
}

// buildDefensePrompt constructs the prompt sent to the LLM for the blue team
// defense phase. It includes the full attack report so the defender knows
// exactly which vulnerabilities to patch.
func (ap *AdversarialProtocol) buildDefensePrompt(
	solution, language string,
	attackReport *AttackReport,
) string {
	var b strings.Builder

	b.WriteString("You are a Blue Team defensive security specialist. ")
	b.WriteString("Patch the following vulnerabilities in the ")
	b.WriteString(language)
	b.WriteString(" code.\n\n")

	b.WriteString("Vulnerabilities found:\n")
	for _, vuln := range attackReport.Vulnerabilities {
		b.WriteString("- ")
		b.WriteString(vuln.ID)
		b.WriteString(" [")
		b.WriteString(vuln.Severity)
		b.WriteString("] (")
		b.WriteString(vuln.Category)
		b.WriteString("): ")
		b.WriteString(vuln.Description)
		b.WriteString("\n")
		b.WriteString("  Exploit: ")
		b.WriteString(vuln.Exploit)
		b.WriteString("\n")
	}

	if len(attackReport.EdgeCases) > 0 {
		b.WriteString("\nEdge cases:\n")
		for _, ec := range attackReport.EdgeCases {
			b.WriteString("- ")
			b.WriteString(ec.ID)
			b.WriteString(": ")
			b.WriteString(ec.Description)
			b.WriteString(" (input: ")
			b.WriteString(ec.Input)
			b.WriteString(")\n")
		}
	}

	b.WriteString("\nOriginal code:\n```")
	b.WriteString(language)
	b.WriteString("\n")
	b.WriteString(solution)
	b.WriteString("\n```\n\n")

	b.WriteString("Respond with structured output using the exact format below.\n")
	b.WriteString("Use --- to separate sections.\n\n")

	b.WriteString("PATCHED_VULNERABILITIES: <comma-separated IDs>\n")
	b.WriteString("PATCHES\n")
	b.WriteString("<VULN-ID>: <fix description>\n")
	b.WriteString("---\n")
	b.WriteString("REMAINING_RISKS: <comma-separated descriptions or NONE>\n")
	b.WriteString("CONFIDENCE: <0.0-1.0>\n")
	b.WriteString("PATCHED_CODE\n```")
	b.WriteString(language)
	b.WriteString("\n<patched code here>\n```\n")

	return b.String()
}

// parseAttackResponse parses the LLM response into an AttackReport using
// simple line-based parsing keyed on section headers.
func (ap *AdversarialProtocol) parseAttackResponse(
	response string,
) (*AttackReport, error) {
	report := &AttackReport{
		Vulnerabilities: make([]Vulnerability, 0),
		EdgeCases:       make([]EdgeCase, 0),
		StressScenarios: make([]StressScenario, 0),
	}

	lines := strings.Split(response, "\n")
	section := ""
	var currentVuln *Vulnerability
	var currentEdge *EdgeCase
	var currentStress *StressScenario

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// Section headers
		if line == "VULNERABILITIES" {
			section = "vuln"
			continue
		}
		if line == "EDGE_CASES" {
			section = "edge"
			continue
		}
		if line == "STRESS_SCENARIOS" {
			section = "stress"
			continue
		}

		// Separator between items within a section
		if line == "---" {
			if currentVuln != nil && currentVuln.ID != "" {
				report.Vulnerabilities = append(
					report.Vulnerabilities, *currentVuln,
				)
			}
			if currentEdge != nil && currentEdge.ID != "" {
				report.EdgeCases = append(report.EdgeCases, *currentEdge)
			}
			if currentStress != nil && currentStress.ID != "" {
				report.StressScenarios = append(
					report.StressScenarios, *currentStress,
				)
			}
			currentVuln = nil
			currentEdge = nil
			currentStress = nil
			continue
		}

		// Overall risk (can appear anywhere)
		if strings.HasPrefix(line, "OVERALL_RISK:") {
			riskStr := strings.TrimSpace(
				strings.TrimPrefix(line, "OVERALL_RISK:"),
			)
			var risk float64
			if _, err := fmt.Sscanf(riskStr, "%f", &risk); err == nil {
				report.OverallRisk = risk
			}
			continue
		}

		// Parse key-value pairs within sections
		switch section {
		case "vuln":
			currentVuln = parseVulnLine(line, currentVuln)
		case "edge":
			currentEdge = parseEdgeLine(line, currentEdge)
		case "stress":
			currentStress = parseStressLine(line, currentStress)
		}
	}

	// Flush any trailing items that were not followed by ---
	if currentVuln != nil && currentVuln.ID != "" {
		report.Vulnerabilities = append(report.Vulnerabilities, *currentVuln)
	}
	if currentEdge != nil && currentEdge.ID != "" {
		report.EdgeCases = append(report.EdgeCases, *currentEdge)
	}
	if currentStress != nil && currentStress.ID != "" {
		report.StressScenarios = append(
			report.StressScenarios, *currentStress,
		)
	}

	return report, nil
}

// parseVulnLine parses a single line within the VULNERABILITIES section.
func parseVulnLine(line string, current *Vulnerability) *Vulnerability {
	if current == nil {
		current = &Vulnerability{}
	}

	if val, ok := extractField(line, "ID:"); ok {
		// New vulnerability starts with an ID line
		if current.ID != "" {
			newV := &Vulnerability{}
			newV.ID = val
			return newV
		}
		current.ID = val
	} else if val, ok := extractField(line, "Category:"); ok {
		current.Category = val
	} else if val, ok := extractField(line, "Severity:"); ok {
		current.Severity = val
	} else if val, ok := extractField(line, "Description:"); ok {
		current.Description = val
	} else if val, ok := extractField(line, "Evidence:"); ok {
		current.Evidence = val
	} else if val, ok := extractField(line, "Exploit:"); ok {
		current.Exploit = val
	}

	return current
}

// parseEdgeLine parses a single line within the EDGE_CASES section.
func parseEdgeLine(line string, current *EdgeCase) *EdgeCase {
	if current == nil {
		current = &EdgeCase{}
	}

	if val, ok := extractField(line, "ID:"); ok {
		if current.ID != "" {
			newE := &EdgeCase{}
			newE.ID = val
			return newE
		}
		current.ID = val
	} else if val, ok := extractField(line, "Description:"); ok {
		current.Description = val
	} else if val, ok := extractField(line, "Input:"); ok {
		current.Input = val
	} else if val, ok := extractField(line, "Expected:"); ok {
		current.Expected = val
	}

	return current
}

// parseStressLine parses a single line within the STRESS_SCENARIOS section.
func parseStressLine(line string, current *StressScenario) *StressScenario {
	if current == nil {
		current = &StressScenario{}
	}

	if val, ok := extractField(line, "ID:"); ok {
		if current.ID != "" {
			newS := &StressScenario{}
			newS.ID = val
			return newS
		}
		current.ID = val
	} else if val, ok := extractField(line, "Description:"); ok {
		current.Description = val
	} else if val, ok := extractField(line, "Load:"); ok {
		current.Load = val
	} else if val, ok := extractField(line, "Expected:"); ok {
		current.Expected = val
	}

	return current
}

// extractField checks whether a line starts with the given key prefix and
// returns the trimmed value after it.
func extractField(line, key string) (string, bool) {
	if strings.HasPrefix(line, key) {
		return strings.TrimSpace(strings.TrimPrefix(line, key)), true
	}
	return "", false
}

// parseDefenseResponse parses the LLM response into a DefenseReport using
// simple line-based parsing keyed on section headers.
func (ap *AdversarialProtocol) parseDefenseResponse(
	response string,
) (*DefenseReport, error) {
	report := &DefenseReport{
		PatchedVulnerabilities: make([]string, 0),
		Patches:                make(map[string]string),
		RemainingRisks:         make([]string, 0),
	}

	lines := strings.Split(response, "\n")
	inPatchSection := false
	inCodeBlock := false
	codeBlockSeen := false
	var codeLines []string

	for i, rawLine := range lines {
		line := strings.TrimSpace(rawLine)

		// Detect the start of the PATCHED_CODE section
		if strings.HasPrefix(line, "PATCHED_CODE") {
			codeBlockSeen = true
			continue
		}

		// Opening code fence after PATCHED_CODE header
		if codeBlockSeen && !inCodeBlock && strings.HasPrefix(line, "```") {
			inCodeBlock = true
			codeLines = make([]string, 0)
			continue
		}

		// Closing code fence
		if inCodeBlock {
			if strings.HasPrefix(line, "```") {
				inCodeBlock = false
				report.PatchedCode = strings.Join(codeLines, "\n")
				continue
			}
			codeLines = append(codeLines, rawLine)
			continue
		}

		_ = i // suppress unused warning if needed

		if line == "" {
			continue
		}

		// PATCHED_VULNERABILITIES line
		if val, ok := extractField(line, "PATCHED_VULNERABILITIES:"); ok {
			ids := strings.Split(val, ",")
			for _, id := range ids {
				trimmed := strings.TrimSpace(id)
				if trimmed != "" {
					report.PatchedVulnerabilities = append(
						report.PatchedVulnerabilities, trimmed,
					)
				}
			}
			continue
		}

		// PATCHES section header
		if line == "PATCHES" {
			inPatchSection = true
			continue
		}
		if line == "---" {
			inPatchSection = false
			continue
		}

		// Patch entries: VULN-ID: description
		if inPatchSection {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if key != "" && val != "" {
					report.Patches[key] = val
				}
			}
			continue
		}

		// REMAINING_RISKS line
		if val, ok := extractField(line, "REMAINING_RISKS:"); ok {
			if strings.ToUpper(strings.TrimSpace(val)) != "NONE" {
				risks := strings.Split(val, ",")
				for _, risk := range risks {
					trimmed := strings.TrimSpace(risk)
					if trimmed != "" {
						report.RemainingRisks = append(
							report.RemainingRisks, trimmed,
						)
					}
				}
			}
			continue
		}

		// CONFIDENCE line
		if val, ok := extractField(line, "CONFIDENCE:"); ok {
			var conf float64
			if _, err := fmt.Sscanf(val, "%f", &conf); err == nil {
				report.ConfidenceInDefense = conf
			}
			continue
		}
	}

	return report, nil
}

// generateFallbackAttack produces a deterministic attack report by scanning
// the solution text for common vulnerability patterns. This is used when the
// LLM is unavailable or returns an unparseable response.
func (ap *AdversarialProtocol) generateFallbackAttack(
	solution, language string,
	round int,
) *AttackReport {
	report := &AttackReport{
		Vulnerabilities: make([]Vulnerability, 0),
		EdgeCases:       make([]EdgeCase, 0),
		StressScenarios: make([]StressScenario, 0),
		Round:           round,
	}

	vulnID := 1
	edgeID := 1

	lower := strings.ToLower(solution)

	// Check for SQL injection patterns
	if containsAny(lower, "query(", "sprintf", "format(") &&
		!containsAny(lower, "prepare", "parameterize", "placeholder") {
		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:       fmt.Sprintf("VULN-%03d", vulnID),
			Category: "injection",
			Severity: "critical",
			Description: "Potential SQL/command injection via " +
				"string concatenation",
			Evidence: "Code uses string formatting for query construction",
			Exploit: "Supply malicious input containing SQL " +
				"metacharacters",
		})
		vulnID++
	}

	// Check for missing input validation
	if !containsAny(lower, "validate", "sanitize", "escape", "clean") {
		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:          fmt.Sprintf("VULN-%03d", vulnID),
			Category:    "logic_error",
			Severity:    "medium",
			Description: "No input validation or sanitization detected",
			Evidence:    "Missing validation/sanitization functions",
			Exploit: "Supply unexpected input types or " +
				"boundary values",
		})
		vulnID++
	}

	// Check for race conditions
	if containsAny(lower, "goroutine", "go func", "thread", "async",
		"concurrent") &&
		!containsAny(lower, "mutex", "lock", "sync.", "atomic",
			"channel") {
		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:       fmt.Sprintf("VULN-%03d", vulnID),
			Category: "race_condition",
			Severity: "high",
			Description: "Concurrent access without " +
				"synchronization primitives",
			Evidence: "Goroutines/threads without visible mutex or " +
				"channel protection",
			Exploit: "Trigger concurrent access to shared state",
		})
		vulnID++
	}

	// Check for missing error handling
	if containsAny(lower, "err !=", "error", "err :=") &&
		containsAny(lower, "_ =", "_ ,") {
		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:          fmt.Sprintf("VULN-%03d", vulnID),
			Category:    "logic_error",
			Severity:    "medium",
			Description: "Error values discarded without handling",
			Evidence: "Blank identifier used to suppress " +
				"error returns",
			Exploit: "Trigger error conditions that are " +
				"silently ignored",
		})
		vulnID++
	}

	// Check for buffer/overflow risk patterns
	if containsAny(lower, "unsafe", "pointer", "buffer", "alloc", "[0]") {
		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:       fmt.Sprintf("VULN-%03d", vulnID),
			Category: "overflow",
			Severity: "high",
			Description: "Potential buffer overflow or " +
				"unsafe memory access",
			Evidence: "Code uses unsafe operations or " +
				"raw buffer access",
			Exploit: "Supply oversized input to overflow " +
				"buffer boundaries",
		})
		vulnID++
	}

	// Language-specific edge cases
	switch strings.ToLower(language) {
	case "go", "golang":
		report.EdgeCases = append(report.EdgeCases, EdgeCase{
			ID:          fmt.Sprintf("EDGE-%03d", edgeID),
			Description: "Nil pointer dereference on zero-value struct",
			Input:       "nil or zero-value argument",
			Expected:    "Graceful error return instead of panic",
		})
		edgeID++
	case "python":
		report.EdgeCases = append(report.EdgeCases, EdgeCase{
			ID:          fmt.Sprintf("EDGE-%03d", edgeID),
			Description: "None passed where object expected",
			Input:       "None",
			Expected:    "TypeError handled gracefully",
		})
		edgeID++
	case "javascript", "typescript", "js", "ts":
		report.EdgeCases = append(report.EdgeCases, EdgeCase{
			ID:          fmt.Sprintf("EDGE-%03d", edgeID),
			Description: "Undefined/null passed where value expected",
			Input:       "undefined",
			Expected:    "Null check prevents runtime error",
		})
		edgeID++
	default:
		report.EdgeCases = append(report.EdgeCases, EdgeCase{
			ID:          fmt.Sprintf("EDGE-%03d", edgeID),
			Description: "Empty input provided",
			Input:       "empty string or empty collection",
			Expected:    "Graceful handling without crash",
		})
		edgeID++
	}

	_ = edgeID // consumed above

	// Standard stress scenario
	report.StressScenarios = append(report.StressScenarios, StressScenario{
		ID:          "STRESS-001",
		Description: "High concurrency load",
		Load:        "1000 concurrent requests",
		Expected:    "Responses within acceptable latency, no resource leaks",
	})

	// Calculate overall risk based on findings
	report.OverallRisk = calculateFallbackRisk(report)

	return report
}

// generateFallbackDefense produces a deterministic defense report that
// acknowledges each vulnerability and provides generic fix descriptions.
// This is used when the LLM is unavailable or returns an unparseable response.
func (ap *AdversarialProtocol) generateFallbackDefense(
	solution string,
	attack *AttackReport,
) *DefenseReport {
	report := &DefenseReport{
		PatchedVulnerabilities: make([]string, 0, len(attack.Vulnerabilities)),
		Patches:                make(map[string]string, len(attack.Vulnerabilities)),
		RemainingRisks:         make([]string, 0),
		PatchedCode:            solution,
	}

	for _, vuln := range attack.Vulnerabilities {
		report.PatchedVulnerabilities = append(
			report.PatchedVulnerabilities, vuln.ID,
		)

		switch vuln.Category {
		case "injection":
			report.Patches[vuln.ID] = "Use parameterized queries " +
				"and input sanitization"
		case "overflow":
			report.Patches[vuln.ID] = "Add bounds checking and " +
				"safe buffer operations"
		case "race_condition":
			report.Patches[vuln.ID] = "Add mutex/channel " +
				"synchronization for shared state"
		case "logic_error":
			report.Patches[vuln.ID] = "Add proper validation " +
				"and error handling"
		case "auth":
			report.Patches[vuln.ID] = "Enforce authentication " +
				"and authorization checks"
		case "xss":
			report.Patches[vuln.ID] = "Escape output and apply " +
				"Content-Security-Policy"
		default:
			report.Patches[vuln.ID] = "Apply defensive coding " +
				"practices for " + vuln.Category
		}
	}

	// Mark edge cases as remaining risks since fallback cannot truly
	// patch code
	for _, ec := range attack.EdgeCases {
		report.RemainingRisks = append(
			report.RemainingRisks,
			fmt.Sprintf("Edge case %s: %s", ec.ID, ec.Description),
		)
	}

	// Confidence is low for fallback-generated defenses
	report.ConfidenceInDefense = 0.4

	return report
}

// containsAny returns true if s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// calculateFallbackRisk computes a risk score based on the number and
// severity of discovered vulnerabilities.
func calculateFallbackRisk(report *AttackReport) float64 {
	if len(report.Vulnerabilities) == 0 {
		return 0.0
	}

	total := 0.0
	for _, vuln := range report.Vulnerabilities {
		switch vuln.Severity {
		case "critical":
			total += 1.0
		case "high":
			total += 0.7
		case "medium":
			total += 0.4
		case "low":
			total += 0.2
		default:
			total += 0.3
		}
	}

	risk := total / float64(len(report.Vulnerabilities))
	if risk > 1.0 {
		risk = 1.0
	}
	return risk
}

// collectRemainingRisks gathers all unresolved risks from the final round
// of attack and defense reports.
func collectRemainingRisks(result *AdversarialResult) []string {
	risks := make([]string, 0)

	if len(result.DefenseReports) > 0 {
		lastDefense := result.DefenseReports[len(result.DefenseReports)-1]
		risks = append(risks, lastDefense.RemainingRisks...)
	}

	// If the last attack found vulnerabilities that were never defended
	if len(result.AttackReports) > len(result.DefenseReports) {
		lastAttack := result.AttackReports[len(result.AttackReports)-1]
		for _, vuln := range lastAttack.Vulnerabilities {
			risks = append(risks, fmt.Sprintf(
				"%s [%s]: %s", vuln.ID, vuln.Severity, vuln.Description,
			))
		}
	}

	return risks
}
