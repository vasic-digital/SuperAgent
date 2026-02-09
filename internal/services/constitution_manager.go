package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ConstitutionRule represents a single rule in the Constitution
type ConstitutionRule struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Mandatory   bool      `json:"mandatory"`
	Priority    int       `json:"priority"` // 1=highest, 5=lowest
	AddedAt     time.Time `json:"added_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Constitution represents the project Constitution
type Constitution struct {
	Version     string             `json:"version"`
	ProjectName string             `json:"project_name"`
	Summary     string             `json:"summary"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Rules       []ConstitutionRule `json:"rules"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

// ConstitutionManager manages Constitution files
type ConstitutionManager struct {
	logger *logrus.Logger
}

// NewConstitutionManager creates a new Constitution manager
func NewConstitutionManager(logger *logrus.Logger) *ConstitutionManager {
	return &ConstitutionManager{
		logger: logger,
	}
}

// LoadOrCreateConstitution loads existing Constitution or creates a new one with mandatory rules
func (cm *ConstitutionManager) LoadOrCreateConstitution(ctx context.Context, projectRoot string) (*Constitution, error) {
	constitutionPath := filepath.Join(projectRoot, "CONSTITUTION.json")

	// Try to load existing Constitution
	if _, err := os.Stat(constitutionPath); err == nil {
		constitution, err := cm.LoadConstitution(constitutionPath)
		if err != nil {
			cm.logger.Warnf("Failed to load existing Constitution: %v, creating new one", err)
		} else {
			cm.logger.Info("Loaded existing Constitution")
			// Ensure mandatory rules are present
			cm.ensureMandatoryRules(constitution)
			return constitution, nil
		}
	}

	// Create new Constitution with mandatory rules
	cm.logger.Info("Creating new Constitution with mandatory rules")
	constitution := cm.createDefaultConstitution(projectRoot)
	return constitution, nil
}

// LoadConstitution loads Constitution from file
func (cm *ConstitutionManager) LoadConstitution(path string) (*Constitution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Constitution file: %w", err)
	}

	var constitution Constitution
	if err := json.Unmarshal(data, &constitution); err != nil {
		return nil, fmt.Errorf("failed to parse Constitution JSON: %w", err)
	}

	return &constitution, nil
}

// SaveConstitution saves Constitution to file
func (cm *ConstitutionManager) SaveConstitution(projectRoot string, constitution *Constitution) error {
	constitutionPath := filepath.Join(projectRoot, "CONSTITUTION.json")

	constitution.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(constitution, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Constitution: %w", err)
	}

	if err := os.WriteFile(constitutionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Constitution file: %w", err)
	}

	cm.logger.WithField("path", constitutionPath).Info("Saved Constitution")
	return nil
}

// UpdateConstitutionFromDebate updates Constitution based on debate results
func (cm *ConstitutionManager) UpdateConstitutionFromDebate(constitution *Constitution, debateResult *DebateResult, userRequest string) (*Constitution, error) {
	// Parse debate output for new/updated rules
	// For now, we'll add a rule based on the user request context
	// In a full implementation, this would parse the debate output for structured rules

	cm.logger.Info("Updating Constitution from debate results")

	// Ensure mandatory rules are present
	cm.ensureMandatoryRules(constitution)

	// Add context-specific rule based on user request
	contextRule := cm.deriveContextRule(userRequest, debateResult)
	if contextRule != nil {
		cm.addOrUpdateRule(constitution, contextRule)
	}

	constitution.Summary = cm.generateSummary(constitution)
	constitution.UpdatedAt = time.Now()

	return constitution, nil
}

// createDefaultConstitution creates a new Constitution with mandatory rules
func (cm *ConstitutionManager) createDefaultConstitution(projectRoot string) *Constitution {
	now := time.Now()

	// Extract project name from path
	projectName := filepath.Base(projectRoot)

	constitution := &Constitution{
		Version:     "1.0.0",
		ProjectName: projectName,
		CreatedAt:   now,
		UpdatedAt:   now,
		Rules:       []ConstitutionRule{},
		Metadata:    make(map[string]any),
	}

	// Add all mandatory rules
	cm.addMandatoryRules(constitution)

	constitution.Summary = cm.generateSummary(constitution)

	return constitution
}

// addMandatoryRules adds all mandatory Constitution rules
func (cm *ConstitutionManager) addMandatoryRules(constitution *Constitution) {
	now := time.Now()
	ruleID := 1

	mandatoryRules := []struct {
		category    string
		title       string
		description string
		priority    int
	}{
		{
			"Architecture",
			"Comprehensive Decoupling",
			"Identify all parts and functionalities that can be extracted as separate modules (libraries) and reused in various projects. Perform additional work to make each module fully decoupled and independent. Each module must be a separate project with its own CLAUDE.md, AGENTS.md, README.md, docs/, tests, and challenges.",
			1,
		},
		{
			"Testing",
			"100% Test Coverage",
			"Every component MUST have 100% test coverage across ALL test types: unit, integration, E2E, security, stress, chaos, automation, and benchmark tests. No false positives. Use real data and live services (mocks only in unit tests).",
			1,
		},
		{
			"Testing",
			"Comprehensive Challenges",
			"Every component MUST have Challenge scripts validating real-life use cases. No false success - validate actual behavior, not return codes.",
			1,
		},
		{
			"Documentation",
			"Complete Documentation",
			"Every module and feature MUST have complete documentation: README.md, CLAUDE.md, AGENTS.md, user guides, step-by-step manuals, video courses, diagrams, SQL definitions, and website content. No component can remain undocumented.",
			1,
		},
		{
			"Quality",
			"No Broken Components",
			"No module, application, library, or test can remain broken, disabled, or incomplete. Everything must be fully functional and operational.",
			1,
		},
		{
			"Quality",
			"No Dead Code",
			"Identify and remove all 'dead code' - features or functionalities left unconnected with the system. Perform comprehensive research and cleanup.",
			1,
		},
		{
			"Safety",
			"Memory Safety",
			"Perform comprehensive research for memory leaks, deadlocks, and race conditions. Apply safety fixes and improvements to prevent these issues.",
			1,
		},
		{
			"Security",
			"Security Scanning",
			"Execute Snyk and SonarQube scanning. Analyze findings in depth and resolve everything. Ensure scanning infrastructure is accessible via containerization (Docker/Podman).",
			1,
		},
		{
			"Performance",
			"Monitoring and Metrics",
			"Create tests that run and perform monitoring and metrics collection. Use collected data for proper optimizations.",
			2,
		},
		{
			"Performance",
			"Lazy Loading and Non-Blocking",
			"Implement lazy loading and lazy initialization wherever possible. Introduce semaphore mechanisms and non-blocking mechanisms to ensure flawless responsiveness.",
			2,
		},
		{
			"Principles",
			"Software Principles",
			"Apply all software principles: KISS, DRY, SOLID, YAGNI, etc. Ensure code is clean, maintainable, and follows best practices.",
			2,
		},
		{
			"Principles",
			"Design Patterns",
			"Use appropriate design patterns: Proxy, Facade, Factory, Abstract Factory, Observer, Mediator, Strategy, etc. Apply patterns where they add value.",
			2,
		},
		{
			"Stability",
			"Rock-Solid Changes",
			"All changes must be safe, non-error-prone, and MUST NOT BREAK any existing working functionality. Ensure backward compatibility unless explicitly breaking.",
			1,
		},
		{
			"Testing",
			"Stress and Integration Tests",
			"Introduce comprehensive stress and integration tests validating that the system is responsive and not possible to overload or break.",
			2,
		},
		{
			"Containerization",
			"Full Containerization",
			"All services MUST run in containers (Docker/Podman/K8s). Support local default execution AND remote configuration. Services must auto-boot before HelixAgent is ready.",
			2,
		},
		{
			"Configuration",
			"Unified Configuration",
			"CLI agent config export uses only HelixAgent + LLMsVerifier's unified generator. No third-party scripts.",
			2,
		},
		{
			"Observability",
			"Health and Monitoring",
			"Every service MUST expose health endpoints. Circuit breakers for all external dependencies. Prometheus/OpenTelemetry integration.",
			2,
		},
		{
			"GitOps",
			"GitSpec Compliance",
			"Follow GitSpec constitution and all constraints from AGENTS.md and CLAUDE.md.",
			2,
		},
		{
			"CI/CD",
			"Manual CI/CD Only",
			"NO GitHub Actions enabled. All CI/CD workflows and pipelines must be executed manually only.",
			1,
		},
		{
			"Documentation",
			"Documentation Synchronization",
			"Anything added to Constitution MUST be present in AGENTS.md and CLAUDE.md, and vice versa. Keep all three synchronized.",
			1,
		},
	}

	for _, rule := range mandatoryRules {
		constitution.Rules = append(constitution.Rules, ConstitutionRule{
			ID:          fmt.Sprintf("CONST-%03d", ruleID),
			Category:    rule.category,
			Title:       rule.title,
			Description: rule.description,
			Mandatory:   true,
			Priority:    rule.priority,
			AddedAt:     now,
			UpdatedAt:   now,
		})
		ruleID++
	}
}

// ensureMandatoryRules ensures all mandatory rules are present in the Constitution
func (cm *ConstitutionManager) ensureMandatoryRules(constitution *Constitution) {
	// Create a temporary Constitution with mandatory rules
	tempConstitution := &Constitution{
		Rules: []ConstitutionRule{},
	}
	cm.addMandatoryRules(tempConstitution)

	// Check which mandatory rules are missing
	existingTitles := make(map[string]bool)
	for _, rule := range constitution.Rules {
		existingTitles[rule.Title] = true
	}

	// Add missing mandatory rules
	for _, mandatoryRule := range tempConstitution.Rules {
		if !existingTitles[mandatoryRule.Title] {
			cm.logger.WithField("title", mandatoryRule.Title).Info("Adding missing mandatory rule")
			constitution.Rules = append(constitution.Rules, mandatoryRule)
		}
	}
}

// deriveContextRule derives a context-specific rule from user request and debate
func (cm *ConstitutionManager) deriveContextRule(userRequest string, debateResult *DebateResult) *ConstitutionRule {
	// Analyze user request to determine if a new rule should be added
	lowerRequest := strings.ToLower(userRequest)

	// Example: if request involves new technology or pattern, add a rule about it
	// This is a simplified version - full implementation would use LLM analysis

	var rule *ConstitutionRule

	if strings.Contains(lowerRequest, "performance") || strings.Contains(lowerRequest, "optimize") {
		rule = &ConstitutionRule{
			ID:          fmt.Sprintf("CONTEXT-%d", time.Now().Unix()),
			Category:    "Performance",
			Title:       "Performance Optimization Context",
			Description: fmt.Sprintf("Context from request: %s. Ensure all performance optimizations maintain system stability and don't sacrifice code clarity.", truncateRequest(userRequest, 100)),
			Mandatory:   false,
			Priority:    3,
			AddedAt:     time.Now(),
			UpdatedAt:   time.Now(),
		}
	} else if strings.Contains(lowerRequest, "security") || strings.Contains(lowerRequest, "auth") {
		rule = &ConstitutionRule{
			ID:          fmt.Sprintf("CONTEXT-%d", time.Now().Unix()),
			Category:    "Security",
			Title:       "Security Enhancement Context",
			Description: fmt.Sprintf("Context from request: %s. Ensure all security measures follow industry best practices and don't introduce vulnerabilities.", truncateRequest(userRequest, 100)),
			Mandatory:   false,
			Priority:    2,
			AddedAt:     time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	return rule
}

// addOrUpdateRule adds a new rule or updates existing one
func (cm *ConstitutionManager) addOrUpdateRule(constitution *Constitution, newRule *ConstitutionRule) {
	// Check if rule with same title exists
	for i, rule := range constitution.Rules {
		if rule.Title == newRule.Title {
			// Update existing rule
			constitution.Rules[i].Description = newRule.Description
			constitution.Rules[i].UpdatedAt = time.Now()
			cm.logger.WithField("title", newRule.Title).Info("Updated existing Constitution rule")
			return
		}
	}

	// Add new rule
	constitution.Rules = append(constitution.Rules, *newRule)
	cm.logger.WithField("title", newRule.Title).Info("Added new Constitution rule")
}

// generateSummary generates a summary of the Constitution
func (cm *ConstitutionManager) generateSummary(constitution *Constitution) string {
	mandatoryCount := 0
	categories := make(map[string]int)

	for _, rule := range constitution.Rules {
		if rule.Mandatory {
			mandatoryCount++
		}
		categories[rule.Category]++
	}

	var categorySummary strings.Builder
	for category, count := range categories {
		if categorySummary.Len() > 0 {
			categorySummary.WriteString(", ")
		}
		categorySummary.WriteString(fmt.Sprintf("%s: %d", category, count))
	}

	return fmt.Sprintf("Constitution with %d rules (%d mandatory) across categories: %s",
		len(constitution.Rules),
		mandatoryCount,
		categorySummary.String(),
	)
}

// GetRulesByCategory returns all rules in a specific category
func (cm *ConstitutionManager) GetRulesByCategory(constitution *Constitution, category string) []ConstitutionRule {
	var rules []ConstitutionRule
	for _, rule := range constitution.Rules {
		if rule.Category == category {
			rules = append(rules, rule)
		}
	}
	return rules
}

// GetMandatoryRules returns all mandatory rules
func (cm *ConstitutionManager) GetMandatoryRules(constitution *Constitution) []ConstitutionRule {
	var rules []ConstitutionRule
	for _, rule := range constitution.Rules {
		if rule.Mandatory {
			rules = append(rules, rule)
		}
	}
	return rules
}

// ValidateCompliance checks if implementation complies with Constitution
func (cm *ConstitutionManager) ValidateCompliance(constitution *Constitution, implementationDetails string) []string {
	var violations []string

	// Simple keyword-based validation for mandatory rules
	// Full implementation would use LLM analysis

	lowerDetails := strings.ToLower(implementationDetails)

	for _, rule := range constitution.Rules {
		if !rule.Mandatory {
			continue
		}

		// Check for violations based on rule category
		switch rule.Category {
		case "Testing":
			if strings.Contains(rule.Title, "100%") && !strings.Contains(lowerDetails, "test") {
				violations = append(violations, fmt.Sprintf("Rule '%s' may be violated: no testing mentioned", rule.Title))
			}
		case "Documentation":
			if !strings.Contains(lowerDetails, "document") && !strings.Contains(lowerDetails, "readme") {
				violations = append(violations, fmt.Sprintf("Rule '%s' may be violated: no documentation mentioned", rule.Title))
			}
		case "Security":
			if strings.Contains(rule.Title, "Scanning") && !strings.Contains(lowerDetails, "scan") && !strings.Contains(lowerDetails, "security") {
				violations = append(violations, fmt.Sprintf("Rule '%s' may be violated: no security scanning mentioned", rule.Title))
			}
		}
	}

	return violations
}

// ExportConstitutionMarkdown exports Constitution as Markdown
func (cm *ConstitutionManager) ExportConstitutionMarkdown(constitution *Constitution) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s Constitution\n\n", constitution.ProjectName))
	md.WriteString(fmt.Sprintf("**Version:** %s\n", constitution.Version))
	md.WriteString(fmt.Sprintf("**Created:** %s\n", constitution.CreatedAt.Format("2006-01-02")))
	md.WriteString(fmt.Sprintf("**Updated:** %s\n\n", constitution.UpdatedAt.Format("2006-01-02")))
	md.WriteString(fmt.Sprintf("%s\n\n", constitution.Summary))

	// Group rules by category
	categories := make(map[string][]ConstitutionRule)
	for _, rule := range constitution.Rules {
		categories[rule.Category] = append(categories[rule.Category], rule)
	}

	// Output rules by category
	for category, rules := range categories {
		md.WriteString(fmt.Sprintf("## %s\n\n", category))
		for _, rule := range rules {
			mandatoryTag := ""
			if rule.Mandatory {
				mandatoryTag = " **[MANDATORY]**"
			}
			priorityTag := fmt.Sprintf(" (Priority: %d)", rule.Priority)
			md.WriteString(fmt.Sprintf("### %s%s%s\n\n", rule.Title, mandatoryTag, priorityTag))
			md.WriteString(fmt.Sprintf("**ID:** %s\n\n", rule.ID))
			md.WriteString(fmt.Sprintf("%s\n\n", rule.Description))
		}
	}

	return md.String()
}

// truncateRequest truncates a request string to max length
func truncateRequest(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
