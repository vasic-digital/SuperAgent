// Package topology provides factory functions for creating debate topologies.
package topology

import (
	"fmt"
)

// NewTopology creates a topology based on the specified type.
func NewTopology(topologyType TopologyType, config TopologyConfig) (Topology, error) {
	config.Type = topologyType

	switch topologyType {
	case TopologyGraphMesh:
		return NewGraphMeshTopology(config), nil
	case TopologyStar:
		return NewStarTopology(config), nil
	case TopologyChain:
		return NewChainTopology(config), nil
	default:
		return nil, fmt.Errorf("unknown topology type: %s", topologyType)
	}
}

// NewDefaultTopology creates a Graph-Mesh topology with default configuration.
// Graph-Mesh is the recommended topology per ACL 2025 research findings.
func NewDefaultTopology() Topology {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	return NewGraphMeshTopology(config)
}

// CreateOptimalTopology selects the best topology based on agent count and requirements.
func CreateOptimalTopology(agentCount int, requirements TopologyRequirements) (Topology, error) {
	topologyType := SelectTopologyType(agentCount, requirements)
	config := CreateTopologyConfig(topologyType, requirements)
	return NewTopology(topologyType, config)
}

// TopologyRequirements specifies requirements for topology selection.
type TopologyRequirements struct {
	// MaxLatency is the maximum acceptable message latency
	MaxLatency int `json:"max_latency_ms"`
	// RequireOrdering requires strict message ordering
	RequireOrdering bool `json:"require_ordering"`
	// MaxParallelism limits parallel operations
	MaxParallelism int `json:"max_parallelism"`
	// EnableDynamicRoles allows runtime role changes
	EnableDynamicRoles bool `json:"enable_dynamic_roles"`
	// CentralizedControl requires a central coordinator
	CentralizedControl bool `json:"centralized_control"`
	// Deterministic requires deterministic behavior
	Deterministic bool `json:"deterministic"`
}

// SelectTopologyType selects the best topology type based on requirements.
func SelectTopologyType(agentCount int, req TopologyRequirements) TopologyType {
	// If strict ordering or deterministic behavior required, use Chain
	if req.RequireOrdering || req.Deterministic {
		return TopologyChain
	}

	// If centralized control required, use Star
	if req.CentralizedControl {
		return TopologyStar
	}

	// For small agent counts, any topology works well
	if agentCount <= 3 {
		return TopologyChain // Simple and efficient for small groups
	}

	// For medium to large groups, prefer Graph-Mesh for parallelism
	if agentCount <= 6 {
		// Star is simpler for medium groups
		if req.MaxParallelism <= 2 {
			return TopologyStar
		}
		return TopologyGraphMesh
	}

	// Large groups benefit most from Graph-Mesh parallelism
	return TopologyGraphMesh
}

// CreateTopologyConfig creates a configuration based on requirements.
func CreateTopologyConfig(topologyType TopologyType, req TopologyRequirements) TopologyConfig {
	config := DefaultTopologyConfig(topologyType)

	if req.MaxParallelism > 0 {
		config.MaxParallelism = req.MaxParallelism
	}

	config.EnableDynamicRoles = req.EnableDynamicRoles

	if req.MaxLatency > 0 {
		// Derive message timeout from latency requirement
		// Give 3x latency for message timeout to account for processing
		config.MessageTimeout = config.MessageTimeout // Keep default for now
	}

	return config
}

// TopologyComparison compares topology options.
type TopologyComparison struct {
	GraphMesh TopologyCharacteristics `json:"graph_mesh"`
	Star      TopologyCharacteristics `json:"star"`
	Chain     TopologyCharacteristics `json:"chain"`
}

// TopologyCharacteristics describes a topology's characteristics.
type TopologyCharacteristics struct {
	Type               TopologyType `json:"type"`
	MaxParallelism     string       `json:"max_parallelism"`     // "full", "limited", "none"
	MessageOrdering    string       `json:"message_ordering"`    // "none", "partial", "strict"
	Bottleneck         string       `json:"bottleneck"`          // "none", "moderator", "sequential"
	Complexity         string       `json:"complexity"`          // "low", "medium", "high"
	BestFor            []string     `json:"best_for"`
	ResearchPerformance string      `json:"research_performance"` // ACL 2025 ranking
}

// GetTopologyComparison returns a comparison of topology options.
func GetTopologyComparison() TopologyComparison {
	return TopologyComparison{
		GraphMesh: TopologyCharacteristics{
			Type:               TopologyGraphMesh,
			MaxParallelism:     "full",
			MessageOrdering:    "none",
			Bottleneck:         "none",
			Complexity:         "high",
			BestFor:            []string{"large teams", "complex debates", "maximum throughput"},
			ResearchPerformance: "#1 - Best performance per ACL 2025",
		},
		Star: TopologyCharacteristics{
			Type:               TopologyStar,
			MaxParallelism:     "limited",
			MessageOrdering:    "partial",
			Bottleneck:         "moderator",
			Complexity:         "medium",
			BestFor:            []string{"centralized control", "simple coordination", "moderate teams"},
			ResearchPerformance: "#2 - Good for controlled scenarios",
		},
		Chain: TopologyCharacteristics{
			Type:               TopologyChain,
			MaxParallelism:     "none",
			MessageOrdering:    "strict",
			Bottleneck:         "sequential",
			Complexity:         "low",
			BestFor:            []string{"deterministic workflows", "audit trails", "small teams"},
			ResearchPerformance: "#3 - Slowest but most predictable",
		},
	}
}

// AgentRoleAssignments provides recommended role assignments for debate teams.
type AgentRoleAssignments struct {
	MinimumTeam    map[AgentRole]int `json:"minimum_team"`    // 5 agents
	StandardTeam   map[AgentRole]int `json:"standard_team"`   // 12 agents
	MaximumTeam    map[AgentRole]int `json:"maximum_team"`    // 20+ agents
}

// GetRecommendedRoleAssignments returns recommended role distributions.
func GetRecommendedRoleAssignments() AgentRoleAssignments {
	return AgentRoleAssignments{
		// Minimum team: 5 agents
		MinimumTeam: map[AgentRole]int{
			RoleProposer:  1,
			RoleCritic:    1,
			RoleReviewer:  1,
			RoleOptimizer: 1,
			RoleModerator: 1,
		},
		// Standard team: 12 agents (from research)
		StandardTeam: map[AgentRole]int{
			RoleProposer:  2,
			RoleCritic:    2,
			RoleReviewer:  2,
			RoleOptimizer: 1,
			RoleModerator: 1,
			RoleArchitect: 1,
			RoleSecurity:  1,
			RoleTestAgent: 1,
			RoleValidator: 1,
		},
		// Maximum team: 20 agents
		MaximumTeam: map[AgentRole]int{
			RoleProposer:  3,
			RoleCritic:    3,
			RoleReviewer:  3,
			RoleOptimizer: 2,
			RoleModerator: 1,
			RoleArchitect: 2,
			RoleSecurity:  2,
			RoleTestAgent: 1,
			RoleRedTeam:   1,
			RoleBlueTeam:  1,
			RoleValidator: 1,
		},
	}
}

// CreateAgentFromSpec creates an Agent from a specification.
func CreateAgentFromSpec(id string, role AgentRole, provider, model string, score float64, specialization string) *Agent {
	return &Agent{
		ID:             id,
		Role:           role,
		Provider:       provider,
		Model:          model,
		Score:          score,
		Confidence:     0.5, // Default starting confidence
		Specialization: specialization,
		Capabilities:   inferCapabilities(specialization),
		Metadata:       make(map[string]interface{}),
	}
}

// inferCapabilities infers capabilities from specialization.
func inferCapabilities(specialization string) []string {
	switch specialization {
	case "code":
		return []string{"code_generation", "code_review", "debugging", "refactoring"}
	case "reasoning":
		return []string{"logical_analysis", "problem_solving", "critique", "evaluation"}
	case "vision":
		return []string{"image_analysis", "diagram_understanding", "visual_reasoning"}
	case "search":
		return []string{"web_search", "information_retrieval", "fact_checking"}
	case "embedding":
		return []string{"semantic_search", "similarity_matching", "clustering"}
	default:
		return []string{"general_purpose", "conversation", "summarization"}
	}
}
