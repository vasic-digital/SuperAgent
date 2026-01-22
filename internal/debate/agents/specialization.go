// Package agents provides specialized agent implementations for AI debates.
// Implements domain specialization, capability discovery, and role-based optimization
// based on research findings from ACL 2025 and MiniMax frameworks.
package agents

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"dev.helix.agent/internal/debate/topology"
	"github.com/google/uuid"
)

// Domain represents an agent's primary area of expertise.
type Domain string

const (
	DomainCode         Domain = "code"         // Code analysis, generation, completion
	DomainSecurity     Domain = "security"     // Vulnerability detection, threat modeling
	DomainArchitecture Domain = "architecture" // System design, scalability analysis
	DomainDebug        Domain = "debug"        // Error diagnosis, trace analysis
	DomainOptimization Domain = "optimization" // Performance analysis, benchmarking
	DomainReasoning    Domain = "reasoning"    // Logical reasoning, problem-solving
	DomainGeneral      Domain = "general"      // General-purpose, no specific domain
)

// CapabilityType identifies specific agent capabilities.
type CapabilityType string

const (
	// Code capabilities
	CapCodeAnalysis    CapabilityType = "code_analysis"
	CapCodeGeneration  CapabilityType = "code_generation"
	CapCodeCompletion  CapabilityType = "code_completion"
	CapCodeRefactoring CapabilityType = "code_refactoring"
	CapTestGeneration  CapabilityType = "test_generation"
	CapCodeReview      CapabilityType = "code_review"

	// Security capabilities
	CapVulnerabilityDetection CapabilityType = "vulnerability_detection"
	CapThreatModeling         CapabilityType = "threat_modeling"
	CapSecurityAudit          CapabilityType = "security_audit"
	CapPenetrationTesting     CapabilityType = "penetration_testing"

	// Architecture capabilities
	CapSystemDesign       CapabilityType = "system_design"
	CapScalabilityDesign  CapabilityType = "scalability_design"
	CapPatternRecognition CapabilityType = "pattern_recognition"
	CapAPIDesign          CapabilityType = "api_design"
	CapDatabaseDesign     CapabilityType = "database_design"

	// Debug capabilities
	CapErrorDiagnosis     CapabilityType = "error_diagnosis"
	CapStackTraceAnalysis CapabilityType = "stack_trace_analysis"
	CapLogAnalysis        CapabilityType = "log_analysis"
	CapRootCauseAnalysis  CapabilityType = "root_cause_analysis"

	// Optimization capabilities
	CapPerformanceAnalysis  CapabilityType = "performance_analysis"
	CapBenchmarking         CapabilityType = "benchmarking"
	CapResourceOptimization CapabilityType = "resource_optimization"
	CapMemoryOptimization   CapabilityType = "memory_optimization"

	// Reasoning capabilities
	CapLogicalReasoning     CapabilityType = "logical_reasoning"
	CapMathematicalProof    CapabilityType = "mathematical_proof"
	CapProblemDecomposition CapabilityType = "problem_decomposition"
	CapCreativeThinking     CapabilityType = "creative_thinking"

	// General capabilities
	CapTextGeneration CapabilityType = "text_generation"
	CapSummarization  CapabilityType = "summarization"
	CapTranslation    CapabilityType = "translation"
	CapConversation   CapabilityType = "conversation"
)

// Capability represents a single capability with its proficiency level.
type Capability struct {
	Type        CapabilityType `json:"type"`
	Proficiency float64        `json:"proficiency"` // 0-1 proficiency level
	Verified    bool           `json:"verified"`    // Discovered at runtime
	Source      string         `json:"source"`      // How discovered (template, provider, runtime)
}

// CapabilitySet holds all capabilities for an agent.
type CapabilitySet struct {
	Capabilities map[CapabilityType]*Capability `json:"capabilities"`
	mu           sync.RWMutex
}

// NewCapabilitySet creates a new capability set.
func NewCapabilitySet() *CapabilitySet {
	return &CapabilitySet{
		Capabilities: make(map[CapabilityType]*Capability),
	}
}

// Add adds a capability to the set.
func (cs *CapabilitySet) Add(cap *Capability) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Capabilities[cap.Type] = cap
}

// Get retrieves a capability by type.
func (cs *CapabilitySet) Get(capType CapabilityType) (*Capability, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	cap, ok := cs.Capabilities[capType]
	return cap, ok
}

// HasCapability checks if the agent has a capability with minimum proficiency.
func (cs *CapabilitySet) HasCapability(capType CapabilityType, minProficiency float64) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	cap, ok := cs.Capabilities[capType]
	return ok && cap.Proficiency >= minProficiency
}

// GetByDomain returns all capabilities for a domain.
func (cs *CapabilitySet) GetByDomain(domain Domain) []*Capability {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	domainCaps := getDomainCapabilities(domain)
	result := make([]*Capability, 0)

	for _, capType := range domainCaps {
		if cap, ok := cs.Capabilities[capType]; ok {
			result = append(result, cap)
		}
	}

	return result
}

// CalculateDomainScore calculates the average proficiency for a domain.
func (cs *CapabilitySet) CalculateDomainScore(domain Domain) float64 {
	caps := cs.GetByDomain(domain)
	if len(caps) == 0 {
		return 0
	}

	total := 0.0
	for _, cap := range caps {
		total += cap.Proficiency
	}
	return total / float64(len(caps))
}

// getDomainCapabilities returns capabilities associated with a domain.
func getDomainCapabilities(domain Domain) []CapabilityType {
	switch domain {
	case DomainCode:
		return []CapabilityType{
			CapCodeAnalysis, CapCodeGeneration, CapCodeCompletion,
			CapCodeRefactoring, CapTestGeneration, CapCodeReview,
		}
	case DomainSecurity:
		return []CapabilityType{
			CapVulnerabilityDetection, CapThreatModeling,
			CapSecurityAudit, CapPenetrationTesting,
		}
	case DomainArchitecture:
		return []CapabilityType{
			CapSystemDesign, CapScalabilityDesign, CapPatternRecognition,
			CapAPIDesign, CapDatabaseDesign,
		}
	case DomainDebug:
		return []CapabilityType{
			CapErrorDiagnosis, CapStackTraceAnalysis,
			CapLogAnalysis, CapRootCauseAnalysis,
		}
	case DomainOptimization:
		return []CapabilityType{
			CapPerformanceAnalysis, CapBenchmarking,
			CapResourceOptimization, CapMemoryOptimization,
		}
	case DomainReasoning:
		return []CapabilityType{
			CapLogicalReasoning, CapMathematicalProof,
			CapProblemDecomposition, CapCreativeThinking,
		}
	default:
		return []CapabilityType{
			CapTextGeneration, CapSummarization,
			CapTranslation, CapConversation,
		}
	}
}

// Specialization defines an agent's area of expertise.
type Specialization struct {
	PrimaryDomain    Domain   `json:"primary_domain"`
	SecondaryDomains []Domain `json:"secondary_domains,omitempty"`
	ExpertiseLevel   float64  `json:"expertise_level"` // 0-1 overall expertise
	Focus            string   `json:"focus,omitempty"` // Specific focus area
	Description      string   `json:"description,omitempty"`
}

// RoleAffinity defines how well-suited an agent is for a debate role.
type RoleAffinity struct {
	Role      topology.AgentRole `json:"role"`
	Affinity  float64            `json:"affinity"` // 0-1 how well suited
	Rationale string             `json:"rationale,omitempty"`
}

// SpecializedAgent represents an agent with domain specialization.
type SpecializedAgent struct {
	// Core identity
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`

	// Provider information
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	Score    float64 `json:"score"` // LLMsVerifier score

	// Specialization
	Specialization *Specialization `json:"specialization"`
	Capabilities   *CapabilitySet  `json:"capabilities"`

	// Role mapping
	RoleAffinities []RoleAffinity     `json:"role_affinities"`
	PrimaryRole    topology.AgentRole `json:"primary_role"`

	// Runtime state
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActive   time.Time              `json:"last_active"`

	mu sync.RWMutex
}

// NewSpecializedAgent creates a new specialized agent.
func NewSpecializedAgent(name, provider, model string, domain Domain) *SpecializedAgent {
	agent := &SpecializedAgent{
		ID:       uuid.New().String(),
		Name:     name,
		Version:  "1.0.0",
		Provider: provider,
		Model:    model,
		Score:    7.0,
		Specialization: &Specialization{
			PrimaryDomain:  domain,
			ExpertiseLevel: 0.7,
		},
		Capabilities:   NewCapabilitySet(),
		RoleAffinities: make([]RoleAffinity, 0),
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
	}

	// Set default capabilities based on domain
	agent.initDefaultCapabilities()

	// Calculate role affinities
	agent.calculateRoleAffinities()

	return agent
}

// initDefaultCapabilities sets up default capabilities based on domain.
func (sa *SpecializedAgent) initDefaultCapabilities() {
	domain := sa.Specialization.PrimaryDomain
	capTypes := getDomainCapabilities(domain)

	for _, capType := range capTypes {
		sa.Capabilities.Add(&Capability{
			Type:        capType,
			Proficiency: sa.Specialization.ExpertiseLevel,
			Verified:    false,
			Source:      "template",
		})
	}

	// Add general capabilities at lower proficiency
	generalCaps := getDomainCapabilities(DomainGeneral)
	for _, capType := range generalCaps {
		sa.Capabilities.Add(&Capability{
			Type:        capType,
			Proficiency: 0.5,
			Verified:    false,
			Source:      "default",
		})
	}
}

// calculateRoleAffinities determines how well the agent fits each debate role.
func (sa *SpecializedAgent) calculateRoleAffinities() {
	domain := sa.Specialization.PrimaryDomain

	// Define role affinity mappings based on domain
	affinityMap := getDomainRoleAffinities(domain)

	for role, affinity := range affinityMap {
		sa.RoleAffinities = append(sa.RoleAffinities, RoleAffinity{
			Role:      role,
			Affinity:  affinity * sa.Specialization.ExpertiseLevel,
			Rationale: fmt.Sprintf("%s specialist suited for %s", domain, role),
		})
	}

	// Sort by affinity (highest first)
	sort.Slice(sa.RoleAffinities, func(i, j int) bool {
		return sa.RoleAffinities[i].Affinity > sa.RoleAffinities[j].Affinity
	})

	// Set primary role
	if len(sa.RoleAffinities) > 0 {
		sa.PrimaryRole = sa.RoleAffinities[0].Role
	}
}

// getDomainRoleAffinities returns the role affinity scores for a domain.
func getDomainRoleAffinities(domain Domain) map[topology.AgentRole]float64 {
	switch domain {
	case DomainCode:
		return map[topology.AgentRole]float64{
			topology.RoleProposer:  0.9, // Generate code solutions
			topology.RoleReviewer:  0.8, // Code review
			topology.RoleOptimizer: 0.7, // Code optimization
			topology.RoleCritic:    0.6, // Find code issues
			topology.RoleModerator: 0.4,
			topology.RoleValidator: 0.5,
		}
	case DomainSecurity:
		return map[topology.AgentRole]float64{
			topology.RoleCritic:    0.95, // Find vulnerabilities
			topology.RoleRedTeam:   0.9,  // Adversarial testing
			topology.RoleValidator: 0.85, // Security validation
			topology.RoleReviewer:  0.7,  // Security review
			topology.RoleBlueTeam:  0.8,  // Defensive analysis
			topology.RoleModerator: 0.3,
		}
	case DomainArchitecture:
		return map[topology.AgentRole]float64{
			topology.RoleArchitect: 0.95, // System design
			topology.RoleModerator: 0.8,  // Guide discussions
			topology.RoleReviewer:  0.75, // Architecture review
			topology.RoleProposer:  0.7,  // Propose designs
			topology.RoleOptimizer: 0.6,  // Optimize architecture
			topology.RoleValidator: 0.5,
		}
	case DomainDebug:
		return map[topology.AgentRole]float64{
			topology.RoleCritic:    0.9,  // Find issues
			topology.RoleReviewer:  0.85, // Analyze code
			topology.RoleTestAgent: 0.8,  // Test scenarios
			topology.RoleValidator: 0.7,  // Validate fixes
			topology.RoleOptimizer: 0.5,
			topology.RoleModerator: 0.3,
		}
	case DomainOptimization:
		return map[topology.AgentRole]float64{
			topology.RoleOptimizer: 0.95, // Optimization focus
			topology.RoleCritic:    0.8,  // Performance critique
			topology.RoleReviewer:  0.7,  // Performance review
			topology.RoleProposer:  0.6,  // Optimization ideas
			topology.RoleValidator: 0.5,
			topology.RoleModerator: 0.4,
		}
	case DomainReasoning:
		return map[topology.AgentRole]float64{
			topology.RoleModerator: 0.9,  // Guide logical flow
			topology.RoleReviewer:  0.85, // Evaluate reasoning
			topology.RoleValidator: 0.8,  // Validate conclusions
			topology.RoleCritic:    0.75, // Find logical flaws
			topology.RoleProposer:  0.7,  // Propose solutions
			topology.RoleTeacher:   0.85, // Explain reasoning
		}
	default:
		// General domain - balanced affinities
		return map[topology.AgentRole]float64{
			topology.RoleProposer:  0.6,
			topology.RoleCritic:    0.6,
			topology.RoleReviewer:  0.6,
			topology.RoleOptimizer: 0.6,
			topology.RoleModerator: 0.6,
			topology.RoleValidator: 0.6,
		}
	}
}

// GetAffinityForRole returns the affinity for a specific role.
func (sa *SpecializedAgent) GetAffinityForRole(role topology.AgentRole) float64 {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	for _, affinity := range sa.RoleAffinities {
		if affinity.Role == role {
			return affinity.Affinity
		}
	}
	return 0.3 // Default low affinity for unknown roles
}

// GetBestRole returns the role with highest affinity.
func (sa *SpecializedAgent) GetBestRole() topology.AgentRole {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	if len(sa.RoleAffinities) > 0 {
		return sa.RoleAffinities[0].Role
	}
	return topology.RoleProposer
}

// GetRolesAboveThreshold returns roles with affinity above threshold.
func (sa *SpecializedAgent) GetRolesAboveThreshold(threshold float64) []topology.AgentRole {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	roles := make([]topology.AgentRole, 0)
	for _, affinity := range sa.RoleAffinities {
		if affinity.Affinity >= threshold {
			roles = append(roles, affinity.Role)
		}
	}
	return roles
}

// SetScore updates the agent's LLMsVerifier score.
func (sa *SpecializedAgent) SetScore(score float64) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.Score = score
}

// SetSystemPrompt sets the agent's system prompt.
func (sa *SpecializedAgent) SetSystemPrompt(prompt string) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.SystemPrompt = prompt
}

// ToTopologyAgent converts to topology.Agent for use in debates.
func (sa *SpecializedAgent) ToTopologyAgent() *topology.Agent {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	capabilities := make([]string, 0)
	for capType := range sa.Capabilities.Capabilities {
		capabilities = append(capabilities, string(capType))
	}

	return &topology.Agent{
		ID:             sa.ID,
		Role:           sa.PrimaryRole,
		Provider:       sa.Provider,
		Model:          sa.Model,
		Score:          sa.Score,
		Confidence:     sa.Specialization.ExpertiseLevel,
		Specialization: string(sa.Specialization.PrimaryDomain),
		Capabilities:   capabilities,
		Metadata: map[string]interface{}{
			"specialization":  sa.Specialization,
			"role_affinities": sa.RoleAffinities,
			"system_prompt":   sa.SystemPrompt,
		},
	}
}

// UpdateActivity marks the agent as active.
func (sa *SpecializedAgent) UpdateActivity() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.LastActive = time.Now()
}

// DiscoverCapabilities performs runtime capability discovery.
func (sa *SpecializedAgent) DiscoverCapabilities(ctx context.Context, discoverer CapabilityDiscoverer) error {
	if discoverer == nil {
		return nil
	}

	discovered, err := discoverer.DiscoverCapabilities(ctx, sa.Provider, sa.Model)
	if err != nil {
		return fmt.Errorf("capability discovery failed: %w", err)
	}

	sa.mu.Lock()
	defer sa.mu.Unlock()

	for _, cap := range discovered {
		cap.Verified = true
		cap.Source = "runtime"
		sa.Capabilities.Add(cap)
	}

	// Recalculate role affinities with new capabilities
	sa.calculateRoleAffinities()

	return nil
}

// CapabilityDiscoverer discovers capabilities for a provider/model.
type CapabilityDiscoverer interface {
	DiscoverCapabilities(ctx context.Context, provider, model string) ([]*Capability, error)
}

// AgentScore represents the composite score for agent selection.
type AgentScore struct {
	AgentID        string  `json:"agent_id"`
	VerifierScore  float64 `json:"verifier_score"`  // LLMsVerifier score
	DomainScore    float64 `json:"domain_score"`    // Domain expertise
	RoleAffinity   float64 `json:"role_affinity"`   // Role fit
	CompositeScore float64 `json:"composite_score"` // Weighted total
}

// CalculateCompositeScore calculates the weighted composite score for selection.
func CalculateCompositeScore(verifierScore, domainScore, roleAffinity float64) float64 {
	// Weights: 40% verifier, 35% domain, 25% role affinity
	return verifierScore*0.4 + domainScore*0.35 + roleAffinity*0.25
}

// ScoreAgent calculates the complete score for an agent in a role context.
func (sa *SpecializedAgent) ScoreAgent(role topology.AgentRole, domain Domain) *AgentScore {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	domainScore := sa.Capabilities.CalculateDomainScore(domain)
	roleAffinity := sa.GetAffinityForRole(role)

	return &AgentScore{
		AgentID:        sa.ID,
		VerifierScore:  sa.Score,
		DomainScore:    domainScore,
		RoleAffinity:   roleAffinity,
		CompositeScore: CalculateCompositeScore(sa.Score/10.0, domainScore, roleAffinity),
	}
}
