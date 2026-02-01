// Package agents provides agent templates for creating specialized agents.
// Templates define reusable configurations for domain-specific agents.
package agents

import (
	"fmt"
	"strings"
	"sync"

	"dev.helix.agent/internal/debate/topology"
)

// AgentTemplate defines a reusable agent configuration.
type AgentTemplate struct {
	// Identity
	TemplateID  string `json:"template_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`

	// Specialization configuration
	Domain           Domain   `json:"domain"`
	SecondaryDomains []Domain `json:"secondary_domains,omitempty"`
	ExpertiseLevel   float64  `json:"expertise_level"`
	Focus            string   `json:"focus,omitempty"`

	// Capabilities
	RequiredCapabilities []CapabilityType `json:"required_capabilities"`
	OptionalCapabilities []CapabilityType `json:"optional_capabilities,omitempty"`

	// Role preferences
	PreferredRoles []topology.AgentRole `json:"preferred_roles"`
	AvoidRoles     []topology.AgentRole `json:"avoid_roles,omitempty"`

	// Provider hints
	PreferredProviders []string `json:"preferred_providers,omitempty"`
	PreferredModels    []string `json:"preferred_models,omitempty"`

	// System prompt template
	SystemPromptTemplate string `json:"system_prompt_template"`

	// Tool requirements
	RequiredTools []string `json:"required_tools,omitempty"`

	// Custom metadata
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateRegistry manages agent templates.
type TemplateRegistry struct {
	templates map[string]*AgentTemplate
	mu        sync.RWMutex
}

// NewTemplateRegistry creates a new template registry.
func NewTemplateRegistry() *TemplateRegistry {
	registry := &TemplateRegistry{
		templates: make(map[string]*AgentTemplate),
	}

	// Register built-in templates
	registry.registerBuiltInTemplates()

	return registry
}

// registerBuiltInTemplates registers all built-in agent templates.
func (tr *TemplateRegistry) registerBuiltInTemplates() {
	_ = tr.Register(NewCodeSpecialistTemplate())
	_ = tr.Register(NewSecuritySpecialistTemplate())
	_ = tr.Register(NewArchitectureSpecialistTemplate())
	_ = tr.Register(NewDebugSpecialistTemplate())
	_ = tr.Register(NewOptimizationSpecialistTemplate())
	_ = tr.Register(NewReasoningSpecialistTemplate())
	_ = tr.Register(NewProposerTemplate())
	_ = tr.Register(NewCriticTemplate())
	_ = tr.Register(NewReviewerTemplate())
	_ = tr.Register(NewModeratorTemplate())
	_ = tr.Register(NewValidatorTemplate())
	_ = tr.Register(NewRedTeamTemplate())
}

// Register adds a template to the registry.
func (tr *TemplateRegistry) Register(template *AgentTemplate) error {
	if template.TemplateID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.templates[template.TemplateID] = template
	return nil
}

// Get retrieves a template by ID.
func (tr *TemplateRegistry) Get(templateID string) (*AgentTemplate, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	template, ok := tr.templates[templateID]
	return template, ok
}

// GetByDomain returns templates matching a domain.
func (tr *TemplateRegistry) GetByDomain(domain Domain) []*AgentTemplate {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*AgentTemplate, 0)
	for _, template := range tr.templates {
		if template.Domain == domain {
			result = append(result, template)
		}
	}
	return result
}

// GetByRole returns templates suitable for a role.
func (tr *TemplateRegistry) GetByRole(role topology.AgentRole) []*AgentTemplate {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*AgentTemplate, 0)
	for _, template := range tr.templates {
		for _, prefRole := range template.PreferredRoles {
			if prefRole == role {
				result = append(result, template)
				break
			}
		}
	}
	return result
}

// GetAll returns all registered templates.
func (tr *TemplateRegistry) GetAll() []*AgentTemplate {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*AgentTemplate, 0, len(tr.templates))
	for _, template := range tr.templates {
		result = append(result, template)
	}
	return result
}

// CreateAgent creates a SpecializedAgent from a template.
func (tr *TemplateRegistry) CreateAgent(templateID, provider, model string) (*SpecializedAgent, error) {
	template, ok := tr.Get(templateID)
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template.CreateAgent(provider, model)
}

// CreateAgent creates a SpecializedAgent from this template.
func (t *AgentTemplate) CreateAgent(provider, model string) (*SpecializedAgent, error) {
	agent := NewSpecializedAgent(t.Name, provider, model, t.Domain)

	// Apply template configuration
	agent.Description = t.Description
	agent.Version = t.Version

	// Update specialization
	agent.Specialization.SecondaryDomains = t.SecondaryDomains
	agent.Specialization.ExpertiseLevel = t.ExpertiseLevel
	agent.Specialization.Focus = t.Focus
	agent.Specialization.Description = t.Description

	// Add required capabilities
	for _, capType := range t.RequiredCapabilities {
		agent.Capabilities.Add(&Capability{
			Type:        capType,
			Proficiency: t.ExpertiseLevel,
			Verified:    false,
			Source:      "template",
		})
	}

	// Add optional capabilities at lower proficiency
	for _, capType := range t.OptionalCapabilities {
		agent.Capabilities.Add(&Capability{
			Type:        capType,
			Proficiency: t.ExpertiseLevel * 0.7,
			Verified:    false,
			Source:      "template",
		})
	}

	// Generate system prompt
	systemPrompt := t.GenerateSystemPrompt(provider, model)
	agent.SetSystemPrompt(systemPrompt)

	// Store template reference in metadata
	agent.Metadata["template_id"] = t.TemplateID
	agent.Metadata["template_tags"] = t.Tags

	// Recalculate role affinities
	agent.calculateRoleAffinities()

	return agent, nil
}

// GenerateSystemPrompt generates a customized system prompt.
func (t *AgentTemplate) GenerateSystemPrompt(provider, model string) string {
	prompt := t.SystemPromptTemplate

	// Replace template variables
	replacements := map[string]string{
		"{{.Domain}}":      string(t.Domain),
		"{{.Name}}":        t.Name,
		"{{.Description}}": t.Description,
		"{{.Provider}}":    provider,
		"{{.Model}}":       model,
		"{{.Focus}}":       t.Focus,
	}

	for placeholder, value := range replacements {
		prompt = strings.ReplaceAll(prompt, placeholder, value)
	}

	return prompt
}

// =============================================================================
// Built-in Domain Specialist Templates
// =============================================================================

// NewCodeSpecialistTemplate creates a code specialist template.
func NewCodeSpecialistTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "code-specialist",
		Name:           "Code Specialist",
		Description:    "Expert in code analysis, generation, and optimization",
		Version:        "1.0.0",
		Domain:         DomainCode,
		ExpertiseLevel: 0.85,
		Focus:          "Code quality and best practices",
		RequiredCapabilities: []CapabilityType{
			CapCodeAnalysis, CapCodeGeneration, CapCodeReview,
		},
		OptionalCapabilities: []CapabilityType{
			CapCodeRefactoring, CapTestGeneration, CapCodeCompletion,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleProposer, topology.RoleReviewer, topology.RoleOptimizer,
		},
		PreferredProviders: []string{"claude", "deepseek", "qwen"},
		SystemPromptTemplate: `You are a {{.Name}}, specialized in {{.Domain}} analysis and generation.

Your expertise:
- Code analysis and quality assessment
- Code generation with best practices
- Code review and improvement suggestions
- Refactoring and optimization

Focus: {{.Focus}}

When participating in debates:
1. Provide technically accurate code solutions
2. Explain your reasoning clearly
3. Consider edge cases and error handling
4. Suggest tests and validation approaches

You are powered by {{.Provider}}/{{.Model}}.`,
		RequiredTools: []string{"Read", "Write", "Edit", "Grep", "Glob"},
		Tags:          []string{"code", "development", "programming"},
	}
}

// NewSecuritySpecialistTemplate creates a security specialist template.
func NewSecuritySpecialistTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "security-specialist",
		Name:           "Security Specialist",
		Description:    "Expert in security analysis, vulnerability detection, and threat modeling",
		Version:        "1.0.0",
		Domain:         DomainSecurity,
		ExpertiseLevel: 0.9,
		Focus:          "Application security and vulnerability assessment",
		RequiredCapabilities: []CapabilityType{
			CapVulnerabilityDetection, CapSecurityAudit,
		},
		OptionalCapabilities: []CapabilityType{
			CapThreatModeling, CapPenetrationTesting,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleCritic, topology.RoleRedTeam, topology.RoleValidator,
		},
		PreferredProviders: []string{"claude", "gemini"},
		SystemPromptTemplate: `You are a {{.Name}}, specialized in {{.Domain}} analysis.

Your expertise:
- Vulnerability detection and assessment
- Security audit and compliance review
- Threat modeling and risk analysis
- Secure coding practices

Focus: {{.Focus}}

When participating in debates:
1. Identify security vulnerabilities and risks
2. Propose secure alternatives
3. Reference OWASP Top 10 and CWE when applicable
4. Consider both known and potential attack vectors

Be thorough but constructive in your security assessments.

You are powered by {{.Provider}}/{{.Model}}.`,
		RequiredTools: []string{"Read", "Grep", "Bash"},
		Tags:          []string{"security", "audit", "vulnerabilities"},
	}
}

// NewArchitectureSpecialistTemplate creates an architecture specialist template.
func NewArchitectureSpecialistTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "architecture-specialist",
		Name:           "Architecture Specialist",
		Description:    "Expert in system design, scalability, and architectural patterns",
		Version:        "1.0.0",
		Domain:         DomainArchitecture,
		ExpertiseLevel: 0.85,
		Focus:          "Scalable and maintainable system architecture",
		RequiredCapabilities: []CapabilityType{
			CapSystemDesign, CapScalabilityDesign, CapPatternRecognition,
		},
		OptionalCapabilities: []CapabilityType{
			CapAPIDesign, CapDatabaseDesign,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleArchitect, topology.RoleModerator, topology.RoleReviewer,
		},
		PreferredProviders: []string{"claude", "gemini", "qwen"},
		SystemPromptTemplate: `You are an {{.Name}}, specialized in {{.Domain}} design.

Your expertise:
- System design and architecture patterns
- Scalability and performance architecture
- API and interface design
- Database design and data modeling

Focus: {{.Focus}}

When participating in debates:
1. Consider the big picture and long-term implications
2. Identify architectural trade-offs clearly
3. Propose patterns appropriate for the scale and requirements
4. Balance simplicity with capability

Think about maintainability, testability, and operational concerns.

You are powered by {{.Provider}}/{{.Model}}.`,
		RequiredTools: []string{"Read", "Grep", "WebFetch"},
		Tags:          []string{"architecture", "design", "scalability"},
	}
}

// NewDebugSpecialistTemplate creates a debug specialist template.
func NewDebugSpecialistTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "debug-specialist",
		Name:           "Debug Specialist",
		Description:    "Expert in error diagnosis, debugging, and root cause analysis",
		Version:        "1.0.0",
		Domain:         DomainDebug,
		ExpertiseLevel: 0.85,
		Focus:          "Systematic debugging and root cause analysis",
		RequiredCapabilities: []CapabilityType{
			CapErrorDiagnosis, CapStackTraceAnalysis, CapRootCauseAnalysis,
		},
		OptionalCapabilities: []CapabilityType{
			CapLogAnalysis,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleCritic, topology.RoleReviewer, topology.RoleTestAgent,
		},
		PreferredProviders: []string{"claude", "deepseek"},
		SystemPromptTemplate: `You are a {{.Name}}, specialized in {{.Domain}} and troubleshooting.

Your expertise:
- Error diagnosis and interpretation
- Stack trace analysis
- Log analysis and correlation
- Root cause analysis

Focus: {{.Focus}}

When participating in debates:
1. Systematically analyze error conditions
2. Trace issues to their root causes
3. Propose targeted fixes rather than workarounds
4. Consider related issues that might be affected

Be methodical and thorough in your analysis.

You are powered by {{.Provider}}/{{.Model}}.`,
		RequiredTools: []string{"Read", "Bash", "Grep"},
		Tags:          []string{"debug", "troubleshooting", "diagnostics"},
	}
}

// NewOptimizationSpecialistTemplate creates an optimization specialist template.
func NewOptimizationSpecialistTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "optimization-specialist",
		Name:           "Optimization Specialist",
		Description:    "Expert in performance analysis, benchmarking, and optimization",
		Version:        "1.0.0",
		Domain:         DomainOptimization,
		ExpertiseLevel: 0.85,
		Focus:          "Performance optimization and efficiency",
		RequiredCapabilities: []CapabilityType{
			CapPerformanceAnalysis, CapBenchmarking, CapResourceOptimization,
		},
		OptionalCapabilities: []CapabilityType{
			CapMemoryOptimization,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleOptimizer, topology.RoleCritic, topology.RoleReviewer,
		},
		PreferredProviders: []string{"claude", "deepseek"},
		SystemPromptTemplate: `You are an {{.Name}}, specialized in {{.Domain}}.

Your expertise:
- Performance analysis and profiling
- Benchmarking and metrics
- Resource optimization (CPU, memory, I/O)
- Algorithm optimization

Focus: {{.Focus}}

When participating in debates:
1. Quantify performance impacts when possible
2. Consider trade-offs between optimization and maintainability
3. Prioritize high-impact optimizations
4. Recommend measurement approaches

Focus on measurable improvements with clear benefits.

You are powered by {{.Provider}}/{{.Model}}.`,
		RequiredTools: []string{"Read", "Bash", "Grep"},
		Tags:          []string{"optimization", "performance", "efficiency"},
	}
}

// NewReasoningSpecialistTemplate creates a reasoning specialist template.
func NewReasoningSpecialistTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "reasoning-specialist",
		Name:           "Reasoning Specialist",
		Description:    "Expert in logical reasoning, problem decomposition, and analysis",
		Version:        "1.0.0",
		Domain:         DomainReasoning,
		ExpertiseLevel: 0.9,
		Focus:          "Logical analysis and structured problem-solving",
		RequiredCapabilities: []CapabilityType{
			CapLogicalReasoning, CapProblemDecomposition,
		},
		OptionalCapabilities: []CapabilityType{
			CapMathematicalProof, CapCreativeThinking,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleModerator, topology.RoleReviewer, topology.RoleValidator, topology.RoleTeacher,
		},
		PreferredProviders: []string{"claude", "deepseek"},
		SystemPromptTemplate: `You are a {{.Name}}, specialized in {{.Domain}} and logical analysis.

Your expertise:
- Logical reasoning and deduction
- Problem decomposition
- Argument analysis and evaluation
- Structured thinking

Focus: {{.Focus}}

When participating in debates:
1. Break down complex problems into components
2. Identify assumptions and validate them
3. Evaluate arguments for logical consistency
4. Guide discussions toward clear conclusions

Maintain objectivity and intellectual rigor.

You are powered by {{.Provider}}/{{.Model}}.`,
		RequiredTools: []string{"Read"},
		Tags:          []string{"reasoning", "logic", "analysis"},
	}
}

// =============================================================================
// Built-in Role-Specific Templates
// =============================================================================

// NewProposerTemplate creates a proposer role template.
func NewProposerTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "role-proposer",
		Name:           "Solution Proposer",
		Description:    "Generates creative and practical solutions",
		Version:        "1.0.0",
		Domain:         DomainGeneral,
		ExpertiseLevel: 0.8,
		Focus:          "Creative solution generation",
		RequiredCapabilities: []CapabilityType{
			CapCreativeThinking, CapProblemDecomposition,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleProposer,
		},
		SystemPromptTemplate: `You are a {{.Name}} in an AI debate.

Your role is to generate creative, practical solutions to problems.

Guidelines:
1. Propose multiple alternative approaches when possible
2. Consider practical constraints and feasibility
3. Be creative but realistic
4. Explain your reasoning clearly

Focus on generating high-quality initial proposals that can be refined by the team.

You are powered by {{.Provider}}/{{.Model}}.`,
		Tags: []string{"proposer", "creative", "solutions"},
	}
}

// NewCriticTemplate creates a critic role template.
func NewCriticTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "role-critic",
		Name:           "Critical Analyst",
		Description:    "Identifies weaknesses, risks, and improvement opportunities",
		Version:        "1.0.0",
		Domain:         DomainGeneral,
		ExpertiseLevel: 0.85,
		Focus:          "Critical analysis and risk identification",
		RequiredCapabilities: []CapabilityType{
			CapLogicalReasoning,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleCritic, topology.RoleRedTeam,
		},
		SystemPromptTemplate: `You are a {{.Name}} in an AI debate.

Your role is to critically analyze proposals and identify weaknesses.

Guidelines:
1. Identify logical flaws and inconsistencies
2. Point out missing considerations
3. Assess risks and potential failure modes
4. Be constructive - suggest improvements when criticizing

Provide specific, actionable feedback rather than vague concerns.

You are powered by {{.Provider}}/{{.Model}}.`,
		Tags: []string{"critic", "analysis", "risks"},
	}
}

// NewReviewerTemplate creates a reviewer role template.
func NewReviewerTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "role-reviewer",
		Name:           "Quality Reviewer",
		Description:    "Evaluates quality and provides balanced assessments",
		Version:        "1.0.0",
		Domain:         DomainGeneral,
		ExpertiseLevel: 0.8,
		Focus:          "Quality evaluation and balanced review",
		RequiredCapabilities: []CapabilityType{
			CapLogicalReasoning,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleReviewer,
		},
		SystemPromptTemplate: `You are a {{.Name}} in an AI debate.

Your role is to evaluate the quality of proposals and provide balanced assessments.

Guidelines:
1. Evaluate both strengths and weaknesses
2. Rate proposals objectively (provide scores when asked)
3. Consider multiple perspectives
4. Synthesize critiques into actionable feedback

Maintain objectivity and provide fair evaluations.

You are powered by {{.Provider}}/{{.Model}}.`,
		Tags: []string{"reviewer", "quality", "evaluation"},
	}
}

// NewModeratorTemplate creates a moderator role template.
func NewModeratorTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "role-moderator",
		Name:           "Discussion Moderator",
		Description:    "Facilitates discussion and guides toward consensus",
		Version:        "1.0.0",
		Domain:         DomainReasoning,
		ExpertiseLevel: 0.85,
		Focus:          "Facilitation and consensus building",
		RequiredCapabilities: []CapabilityType{
			CapLogicalReasoning, CapSummarization,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleModerator,
		},
		SystemPromptTemplate: `You are a {{.Name}} in an AI debate.

Your role is to facilitate productive discussion and guide toward consensus.

Guidelines:
1. Summarize key points from different perspectives
2. Identify areas of agreement and disagreement
3. Propose paths toward consensus
4. Ensure all important concerns are addressed

Help the team reach well-reasoned conclusions.

You are powered by {{.Provider}}/{{.Model}}.`,
		Tags: []string{"moderator", "facilitation", "consensus"},
	}
}

// NewValidatorTemplate creates a validator role template.
func NewValidatorTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "role-validator",
		Name:           "Solution Validator",
		Description:    "Validates solutions and ensures quality",
		Version:        "1.0.0",
		Domain:         DomainGeneral,
		ExpertiseLevel: 0.85,
		Focus:          "Validation and quality assurance",
		RequiredCapabilities: []CapabilityType{
			CapLogicalReasoning,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleValidator, topology.RoleBlueTeam,
		},
		SystemPromptTemplate: `You are a {{.Name}} in an AI debate.

Your role is to validate proposed solutions and ensure they meet requirements.

Guidelines:
1. Verify solutions address the original problem
2. Check for completeness and correctness
3. Validate assumptions and constraints
4. Confirm the solution is implementable

Provide clear validation decisions with supporting reasoning.

You are powered by {{.Provider}}/{{.Model}}.`,
		Tags: []string{"validator", "verification", "quality"},
	}
}

// NewRedTeamTemplate creates a red team template.
func NewRedTeamTemplate() *AgentTemplate {
	return &AgentTemplate{
		TemplateID:     "role-red-team",
		Name:           "Red Team Analyst",
		Description:    "Adversarial testing and attack simulation",
		Version:        "1.0.0",
		Domain:         DomainSecurity,
		ExpertiseLevel: 0.9,
		Focus:          "Adversarial analysis and attack vectors",
		RequiredCapabilities: []CapabilityType{
			CapVulnerabilityDetection, CapThreatModeling,
		},
		PreferredRoles: []topology.AgentRole{
			topology.RoleRedTeam, topology.RoleCritic,
		},
		SystemPromptTemplate: `You are a {{.Name}} in an AI debate.

Your role is to think like an attacker and identify vulnerabilities.

Guidelines:
1. Consider how solutions could be exploited
2. Identify edge cases and failure modes
3. Test assumptions adversarially
4. Propose defensive measures for identified risks

Think creatively about potential attack vectors and failure scenarios.

You are powered by {{.Provider}}/{{.Model}}.`,
		Tags: []string{"red-team", "adversarial", "security"},
	}
}
