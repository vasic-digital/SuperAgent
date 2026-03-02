package comprehensive

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// PromptBuilder helps construct prompts for agents
type PromptBuilder struct {
	systemPrompt string
	context      []string
	instructions []string
	examples     []string
	constraints  []string
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder(systemPrompt string) *PromptBuilder {
	return &PromptBuilder{
		systemPrompt: systemPrompt,
		context:      make([]string, 0),
		instructions: make([]string, 0),
		examples:     make([]string, 0),
		constraints:  make([]string, 0),
	}
}

// AddContext adds context to the prompt
func (pb *PromptBuilder) AddContext(context string) *PromptBuilder {
	pb.context = append(pb.context, context)
	return pb
}

// AddInstruction adds an instruction
func (pb *PromptBuilder) AddInstruction(instruction string) *PromptBuilder {
	pb.instructions = append(pb.instructions, instruction)
	return pb
}

// AddExample adds an example
func (pb *PromptBuilder) AddExample(input, output string) *PromptBuilder {
	example := fmt.Sprintf("Input: %s\nOutput: %s", input, output)
	pb.examples = append(pb.examples, example)
	return pb
}

// AddConstraint adds a constraint
func (pb *PromptBuilder) AddConstraint(constraint string) *PromptBuilder {
	pb.constraints = append(pb.constraints, constraint)
	return pb
}

// Build builds the complete prompt
func (pb *PromptBuilder) Build() string {
	var parts []string

	// System prompt
	if pb.systemPrompt != "" {
		parts = append(parts, fmt.Sprintf("# System\n%s", pb.systemPrompt))
	}

	// Context
	if len(pb.context) > 0 {
		parts = append(parts, fmt.Sprintf("# Context\n%s", strings.Join(pb.context, "\n\n")))
	}

	// Instructions
	if len(pb.instructions) > 0 {
		parts = append(parts, fmt.Sprintf("# Instructions\n%s", strings.Join(pb.instructions, "\n")))
	}

	// Examples
	if len(pb.examples) > 0 {
		parts = append(parts, fmt.Sprintf("# Examples\n%s", strings.Join(pb.examples, "\n\n")))
	}

	// Constraints
	if len(pb.constraints) > 0 {
		parts = append(parts, fmt.Sprintf("# Constraints\n%s", strings.Join(pb.constraints, "\n")))
	}

	return strings.Join(parts, "\n\n")
}

// RolePrompts provides role-specific prompts
type RolePrompts struct{}

// Architect returns the prompt for architect agent
func (rp RolePrompts) Architect() string {
	return NewPromptBuilder("You are an expert software architect specializing in system design.").
		AddInstruction("Design scalable, maintainable systems with clear component boundaries").
		AddInstruction("Consider performance, security, and reliability in all designs").
		AddInstruction("Use established design patterns and best practices").
		AddConstraint("Provide clear rationale for all architectural decisions").
		AddConstraint("Consider trade-offs between different approaches").
		Build()
}

// Generator returns the prompt for generator agent
func (rp RolePrompts) Generator() string {
	return NewPromptBuilder("You are an expert software developer specializing in code generation").
		AddInstruction("Generate clean, idiomatic code following best practices").
		AddInstruction("Include comprehensive error handling").
		AddInstruction("Write self-documenting code with clear naming").
		AddInstruction("Follow the project's coding standards and conventions").
		AddConstraint("Ensure code compiles without errors").
		AddConstraint("Handle edge cases appropriately").
		AddConstraint("Add comments for complex logic").
		Build()
}

// Critic returns the prompt for critic agent
func (rp RolePrompts) Critic() string {
	return NewPromptBuilder("You are a rigorous code reviewer focused on finding issues").
		AddInstruction("Identify bugs, security vulnerabilities, and logical errors").
		AddInstruction("Check for proper error handling").
		AddInstruction("Verify edge case coverage").
		AddInstruction("Assess code quality and maintainability").
		AddInstruction("Provide specific, actionable feedback").
		AddConstraint("Be thorough but constructive").
		AddConstraint("Prioritize critical issues over style preferences").
		Build()
}

// Refactoring returns the prompt for refactoring agent
func (rp RolePrompts) Refactoring() string {
	return NewPromptBuilder("You are a code refactoring specialist").
		AddInstruction("Improve code without changing behavior").
		AddInstruction("Reduce complexity and improve readability").
		AddInstruction("Apply appropriate design patterns").
		AddInstruction("Optimize for maintainability").
		AddConstraint("Preserve all existing functionality").
		AddConstraint("Maintain or improve test coverage").
		AddConstraint("Follow the DRY and SOLID principles").
		Build()
}

// Tester returns the prompt for tester agent
func (rp RolePrompts) Tester() string {
	return NewPromptBuilder("You are a testing expert specializing in comprehensive test coverage").
		AddInstruction("Generate unit, integration, and edge case tests").
		AddInstruction("Ensure high code coverage").
		AddInstruction("Test both happy paths and error conditions").
		AddInstruction("Use appropriate testing frameworks and patterns").
		AddConstraint("Tests must be deterministic").
		AddConstraint("Include clear test descriptions").
		Build()
}

// Validator returns the prompt for validator agent
func (rp RolePrompts) Validator() string {
	return NewPromptBuilder("You are a correctness validation expert").
		AddInstruction("Verify code correctness through analysis").
		AddInstruction("Check for logical consistency").
		AddInstruction("Validate against requirements").
		AddInstruction("Identify potential race conditions").
		AddConstraint("Be precise in correctness assessments").
		AddConstraint("Distinguish between certain and probable issues").
		Build()
}

// Security returns the prompt for security agent
func (rp RolePrompts) Security() string {
	return NewPromptBuilder("You are a security expert and penetration tester").
		AddInstruction("Identify security vulnerabilities and attack vectors").
		AddInstruction("Check for OWASP top 10 vulnerabilities").
		AddInstruction("Analyze input validation and sanitization").
		AddInstruction("Assess authentication and authorization").
		AddConstraint("Think like an attacker").
		AddConstraint("Prioritize high-impact vulnerabilities").
		Build()
}

// Performance returns the prompt for performance agent
func (rp RolePrompts) Performance() string {
	return NewPromptBuilder("You are a performance optimization expert").
		AddInstruction("Identify performance bottlenecks").
		AddInstruction("Suggest algorithmic improvements").
		AddInstruction("Optimize resource usage").
		AddInstruction("Consider memory allocation and GC impact").
		AddConstraint("Benchmark before and after changes").
		AddConstraint("Consider maintainability vs performance trade-offs").
		Build()
}

// RedTeam returns the prompt for red team agent
func (rp RolePrompts) RedTeam() string {
	return NewPromptBuilder("You are an adversarial testing expert (Red Team) specializing in breaking systems and finding vulnerabilities").
		AddInstruction("Test adversarial scenarios and attack vectors").
		AddInstruction("Find edge cases that cause failures").
		AddInstruction("Think like an attacker trying to exploit the system").
		AddInstruction("Identify security vulnerabilities through offensive testing").
		AddConstraint("Focus on practical exploitation scenarios").
		AddConstraint("Prioritize high-impact vulnerabilities").
		Build()
}

// BlueTeam returns the prompt for blue team agent
func (rp RolePrompts) BlueTeam() string {
	return NewPromptBuilder("You are a defensive implementation expert (Blue Team) specializing in robust, secure code").
		AddInstruction("Implement defensive solutions with robust error handling").
		AddInstruction("Add validation and security safeguards").
		AddInstruction("Ensure code handles edge cases gracefully").
		AddInstruction("Focus on reliability and security").
		AddConstraint("Prioritize security over convenience").
		AddConstraint("Ensure all failure modes are handled").
		Build()
}

// Moderator returns the prompt for moderator agent
func (rp RolePrompts) Moderator() string {
	return NewPromptBuilder("You are a debate moderator ensuring productive discussion").
		AddInstruction("Facilitate structured debate between agents").
		AddInstruction("Ensure all perspectives are heard").
		AddInstruction("Identify areas of agreement and disagreement").
		AddInstruction("Guide toward consensus").
		AddConstraint("Remain neutral and objective").
		AddConstraint("Prevent circular arguments").
		Build()
}

// Parser provides parsing utilities
type Parser struct{}

// ParseCodeBlocks extracts code blocks from markdown
func (p Parser) ParseCodeBlocks(content string) []CodeBlock {
	var blocks []CodeBlock

	// Match code blocks with optional language specifier
	re := regexp.MustCompile("(?s)```(\\w+)?\\n(.*?)```")
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		language := ""
		if len(match) > 1 {
			language = match[1]
		}
		code := ""
		if len(match) > 2 {
			code = match[2]
		}
		blocks = append(blocks, CodeBlock{
			Language: language,
			Code:     code,
		})
	}

	return blocks
}

// CodeBlock represents a code block
type CodeBlock struct {
	Language string
	Code     string
}

// ExtractThoughts extracts thinking/reasoning sections
func (p Parser) ExtractThoughts(content string) []string {
	var thoughts []string

	// Look for common patterns
	patterns := []string{
		`(?i)Thinking:\s*(.+?)(?:\n\n|\n*$)`,
		`(?i)Reasoning:\s*(.+?)(?:\n\n|\n*$)`,
		`(?i)Analysis:\s*(.+?)(?:\n\n|\n*$)`,
		`<think>([\s\S]*?)</think>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				thoughts = append(thoughts, strings.TrimSpace(match[1]))
			}
		}
	}

	return thoughts
}

// ParseConfidence extracts confidence score from content
func (p Parser) ParseConfidence(content string) float64 {
	// Look for confidence indicators
	patterns := []string{
		`(?i)confidence[:\s]+(\d+(?:\.\d+)?)\s*%?`,
		`(?i)(\d+(?:\.\d+)?)\s*%\s+confident`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			var score float64
			n, err := fmt.Sscanf(matches[1], "%f", &score)
			if n == 1 && err == nil {
				if score > 1.0 {
					score = score / 100.0
				}
				if score >= 0 && score <= 1 {
					return score
				}
			}
		}
	}

	return 0.5 // Default confidence
}

// ExtractKeyPoints extracts bullet points or numbered lists
func (p Parser) ExtractKeyPoints(content string) []string {
	var points []string

	// Match bullet points and numbered lists
	re := regexp.MustCompile(`(?m)^\s*(?:[-*•]|\d+\.)\s+(.+)$`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			points = append(points, strings.TrimSpace(match[1]))
		}
	}

	return points
}

// Validator provides validation utilities
type Validator struct{}

// ValidateCode checks if code is valid Go code
func (v Validator) ValidateCode(code, language string) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(code) == "" {
		errors = append(errors, ValidationError{
			Type:    "empty",
			Message: "Code block is empty",
		})
		return errors
	}

	if language == "go" || language == "" {
		// Check for common Go issues
		if !strings.Contains(code, "package") {
			errors = append(errors, ValidationError{
				Type:    "missing_package",
				Message: "Go code missing package declaration",
			})
		}

		if strings.Count(code, "{") != strings.Count(code, "}") {
			errors = append(errors, ValidationError{
				Type:    "unbalanced_braces",
				Message: "Unbalanced curly braces",
			})
		}
	}

	return errors
}

// ValidationError represents a validation error
type ValidationError struct {
	Type    string
	Message string
	Line    int
	Column  int
}

// ValidateAgentResponse validates an agent response
func (v Validator) ValidateAgentResponse(resp *AgentResponse) []ValidationError {
	var errors []ValidationError

	if resp.Content == "" {
		errors = append(errors, ValidationError{
			Type:    "empty_response",
			Message: "Response content is empty",
		})
	}

	if resp.Confidence < 0 || resp.Confidence > 1 {
		errors = append(errors, ValidationError{
			Type:    "invalid_confidence",
			Message: "Confidence must be between 0 and 1",
		})
	}

	if resp.Latency > 5*time.Minute {
		errors = append(errors, ValidationError{
			Type:    "slow_response",
			Message: "Response took too long",
		})
	}

	return errors
}

// DebateRequestValidator validates debate requests
type DebateRequestValidator struct{}

// Validate validates a debate request
func (v DebateRequestValidator) Validate(req *DebateRequest) []ValidationError {
	var errors []ValidationError

	if req.ID == "" {
		errors = append(errors, ValidationError{
			Type:    "missing_id",
			Message: "Debate ID is required",
		})
	}

	if req.Topic == "" {
		errors = append(errors, ValidationError{
			Type:    "missing_topic",
			Message: "Debate topic is required",
		})
	}

	if req.MaxRounds < 1 || req.MaxRounds > 100 {
		errors = append(errors, ValidationError{
			Type:    "invalid_rounds",
			Message: "Max rounds must be between 1 and 100",
		})
	}

	if req.Timeout > 0 && req.Timeout < 30*time.Second {
		errors = append(errors, ValidationError{
			Type:    "invalid_timeout",
			Message: "Timeout must be at least 30 seconds",
		})
	}

	return errors
}

// Helper functions

// TruncateString truncates a string to max length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// CleanWhitespace normalizes whitespace in a string
func CleanWhitespace(s string) string {
	// Replace multiple whitespace with single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// CountTokens estimates token count (rough approximation)
func CountTokens(s string) int {
	// Rough estimate: ~4 characters per token
	return len(s) / 4
}
