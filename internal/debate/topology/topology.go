// Package topology provides debate coordination topologies for AI ensemble orchestration.
// Implements Graph-Mesh, Star, and Chain topologies per ACL 2025 research findings.
package topology

import (
	"context"
	"sync"
	"time"
)

// TopologyType identifies the coordination topology.
type TopologyType string

const (
	// TopologyGraphMesh - All agents can communicate with all others (best performance)
	TopologyGraphMesh TopologyType = "graph_mesh"
	// TopologyStar - Moderator-centric communication (simple but bottleneck)
	TopologyStar TopologyType = "star"
	// TopologyChain - Sequential communication (deterministic but slow)
	TopologyChain TopologyType = "chain"
	// TopologyTree - Hierarchical parent-child communication (balanced delegation)
	TopologyTree TopologyType = "tree"
)

// AgentRole defines the role of an agent in the debate.
type AgentRole string

const (
	// Core proposal & generation roles
	RoleProposer  AgentRole = "proposer"  // Generates initial solutions (generic)
	RoleGenerator AgentRole = "generator" // Code generation specialist (documentation requirement)

	// Analysis & critique roles
	RoleCritic   AgentRole = "critic"   // Identifies weaknesses
	RoleReviewer AgentRole = "reviewer" // Evaluates quality

	// Improvement & optimization roles
	RoleOptimizer           AgentRole = "optimizer"            // General improvements
	RoleRefactorer          AgentRole = "refactorer"           // Refactoring specialist (documentation requirement)
	RolePerformanceAnalyzer AgentRole = "performance_analyzer" // Performance optimization specialist (documentation requirement)

	// Coordination roles
	RoleModerator AgentRole = "moderator" // Facilitates discussion

	// Specialized roles
	RoleArchitect AgentRole = "architect"  // Designs structure
	RoleSecurity  AgentRole = "security"   // Security analysis
	RoleTestAgent AgentRole = "test_agent" // Test generation
	RoleValidator AgentRole = "validator"  // Final validation

	// Adversarial roles
	RoleRedTeam  AgentRole = "red_team"  // Adversarial testing
	RoleBlueTeam AgentRole = "blue_team" // Defensive validation

	// Knowledge roles
	RoleTeacher AgentRole = "teacher" // Knowledge transfer

	// Extended specialized roles (from debate spec documents)
	RoleCompiler    AgentRole = "compiler"    // Validates syntax, type safety, build correctness
	RoleExecutor    AgentRole = "executor"    // Runs code in sandbox, collects runtime feedback
	RoleJudge       AgentRole = "judge"       // Scores solutions objectively against rubric
	RoleImplementer AgentRole = "implementer" // Turns specs into concrete code
	RoleDesigner    AgentRole = "designer"    // High-level system design and decomposition
)

// Agent represents a participant in the debate topology.
type Agent struct {
	ID             string                 `json:"id"`
	Role           AgentRole              `json:"role"`
	Provider       string                 `json:"provider"`
	Model          string                 `json:"model"`
	Score          float64                `json:"score"`          // LLMsVerifier score
	Confidence     float64                `json:"confidence"`     // Current confidence level
	Specialization string                 `json:"specialization"` // code, reasoning, vision, etc.
	Capabilities   []string               `json:"capabilities"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`

	// Dynamic state
	lastActive      time.Time
	messageCount    int
	avgResponseTime time.Duration
	mu              sync.RWMutex
}

// UpdateActivity updates the agent's activity metrics.
func (a *Agent) UpdateActivity(responseTime time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastActive = time.Now()
	a.messageCount++

	// Rolling average response time
	if a.avgResponseTime == 0 {
		a.avgResponseTime = responseTime
	} else {
		a.avgResponseTime = (a.avgResponseTime + responseTime) / 2
	}
}

// GetMetrics returns the agent's current metrics.
func (a *Agent) GetMetrics() AgentMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return AgentMetrics{
		MessageCount:    a.messageCount,
		AvgResponseTime: a.avgResponseTime,
		LastActive:      a.lastActive,
	}
}

// AgentMetrics holds agent performance metrics.
type AgentMetrics struct {
	MessageCount    int           `json:"message_count"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	LastActive      time.Time     `json:"last_active"`
}

// Message represents a communication between agents.
type Message struct {
	ID          string                 `json:"id"`
	FromAgent   string                 `json:"from_agent"`
	ToAgents    []string               `json:"to_agents"` // Empty = broadcast
	Content     string                 `json:"content"`
	MessageType MessageType            `json:"message_type"`
	Phase       DebatePhase            `json:"phase"`
	Round       int                    `json:"round"`
	Timestamp   time.Time              `json:"timestamp"`
	ReplyTo     string                 `json:"reply_to,omitempty"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageType identifies the type of message.
type MessageType string

const (
	MessageTypeProposal     MessageType = "proposal"
	MessageTypeCritique     MessageType = "critique"
	MessageTypeReview       MessageType = "review"
	MessageTypeOptimization MessageType = "optimization"
	MessageTypeConvergence  MessageType = "convergence"
	MessageTypeQuestion     MessageType = "question"
	MessageTypeAnswer       MessageType = "answer"
	MessageTypeAcknowledge  MessageType = "acknowledge"
	MessageTypeValidation   MessageType = "validation"
	MessageTypeRefinement   MessageType = "refinement"
)

// DebatePhase represents the current phase of the debate.
type DebatePhase string

const (
	// Pre-debate phases
	PhaseDehallucination DebatePhase = "dehallucination"  // Proactive clarification to reduce hallucination
	PhaseSelfEvolvement  DebatePhase = "self_evolvement"  // Agents self-test and refine before debate

	// Core debate phases
	PhaseProposal     DebatePhase = "proposal"     // Initial solution generation
	PhaseCritique     DebatePhase = "critique"      // Weakness identification
	PhaseReview       DebatePhase = "review"        // Quality evaluation
	PhaseOptimization DebatePhase = "optimization"  // Solution improvement
	PhaseAdversarial  DebatePhase = "adversarial"   // Red/Blue Team attack-defend cycle
	PhaseConvergence  DebatePhase = "convergence"   // Final consensus
)

// CommunicationChannel represents a communication pathway between agents.
type CommunicationChannel struct {
	FromAgent     string
	ToAgent       string
	Bidirectional bool
	Weight        float64 // Channel priority/bandwidth
}

// Topology defines the interface for debate coordination topologies.
type Topology interface {
	// GetType returns the topology type.
	GetType() TopologyType

	// Initialize sets up the topology with the given agents.
	Initialize(ctx context.Context, agents []*Agent) error

	// GetAgents returns all agents in the topology.
	GetAgents() []*Agent

	// GetAgent returns an agent by ID.
	GetAgent(id string) (*Agent, bool)

	// GetAgentsByRole returns agents with the specified role.
	GetAgentsByRole(role AgentRole) []*Agent

	// GetChannels returns all communication channels.
	GetChannels() []CommunicationChannel

	// CanCommunicate checks if two agents can communicate directly.
	CanCommunicate(fromID, toID string) bool

	// GetCommunicationTargets returns agents that the given agent can message.
	GetCommunicationTargets(agentID string) []*Agent

	// RouteMessage determines the routing for a message.
	RouteMessage(msg *Message) ([]*Agent, error)

	// BroadcastMessage sends a message to all reachable agents.
	BroadcastMessage(ctx context.Context, msg *Message) error

	// SendMessage sends a message to specific agents.
	SendMessage(ctx context.Context, msg *Message) error

	// GetNextPhase determines the next debate phase.
	GetNextPhase(currentPhase DebatePhase) DebatePhase

	// SelectLeader dynamically selects a leader for the current phase.
	SelectLeader(phase DebatePhase) (*Agent, error)

	// AssignRole dynamically assigns a role to an agent.
	AssignRole(agentID string, role AgentRole) error

	// GetParallelGroups returns groups of agents that can work in parallel.
	GetParallelGroups(phase DebatePhase) [][]*Agent

	// GetMetrics returns topology performance metrics.
	GetMetrics() TopologyMetrics

	// Close cleans up resources.
	Close() error
}

// TopologyMetrics holds topology performance metrics.
type TopologyMetrics struct {
	TotalMessages      int64                   `json:"total_messages"`
	BroadcastCount     int64                   `json:"broadcast_count"`
	DirectMessageCount int64                   `json:"direct_message_count"`
	AvgMessageLatency  time.Duration           `json:"avg_message_latency"`
	AgentMetrics       map[string]AgentMetrics `json:"agent_metrics"`
	PhaseTransitions   int                     `json:"phase_transitions"`
	ChannelUtilization map[string]float64      `json:"channel_utilization"`
}

// TopologyConfig configures a topology instance.
type TopologyConfig struct {
	Type                TopologyType           `json:"type"`
	MaxParallelism      int                    `json:"max_parallelism"`       // Max concurrent operations
	MessageTimeout      time.Duration          `json:"message_timeout"`       // Per-message timeout
	EnableDynamicRoles  bool                   `json:"enable_dynamic_roles"`  // Allow runtime role changes
	EnableLoadBalancing bool                   `json:"enable_load_balancing"` // Distribute load across agents
	PriorityChannels    []string               `json:"priority_channels"`     // High-priority agent pairs
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultTopologyConfig returns a default configuration.
func DefaultTopologyConfig(topologyType TopologyType) TopologyConfig {
	return TopologyConfig{
		Type:                topologyType,
		MaxParallelism:      8,
		MessageTimeout:      30 * time.Second,
		EnableDynamicRoles:  true,
		EnableLoadBalancing: true,
		PriorityChannels:    nil,
		Metadata:            make(map[string]interface{}),
	}
}

// BaseTopology provides common functionality for all topologies.
type BaseTopology struct {
	config       TopologyConfig
	agents       map[string]*Agent
	agentsByRole map[AgentRole][]*Agent
	channels     []CommunicationChannel
	metrics      TopologyMetrics
	msgChan      chan *Message

	mu        sync.RWMutex
	metricsMu sync.RWMutex
	closed    bool
}

// NewBaseTopology creates a new base topology.
func NewBaseTopology(config TopologyConfig) *BaseTopology {
	return &BaseTopology{
		config:       config,
		agents:       make(map[string]*Agent),
		agentsByRole: make(map[AgentRole][]*Agent),
		channels:     make([]CommunicationChannel, 0),
		metrics: TopologyMetrics{
			AgentMetrics:       make(map[string]AgentMetrics),
			ChannelUtilization: make(map[string]float64),
		},
		msgChan: make(chan *Message, 1000),
	}
}

// AddAgent adds an agent to the topology.
func (bt *BaseTopology) AddAgent(agent *Agent) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.agents[agent.ID] = agent
	bt.agentsByRole[agent.Role] = append(bt.agentsByRole[agent.Role], agent)
}

// GetAgents returns all agents.
func (bt *BaseTopology) GetAgents() []*Agent {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	agents := make([]*Agent, 0, len(bt.agents))
	for _, agent := range bt.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetAgent returns an agent by ID.
func (bt *BaseTopology) GetAgent(id string) (*Agent, bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	agent, ok := bt.agents[id]
	return agent, ok
}

// GetAgentsByRole returns agents with the specified role.
func (bt *BaseTopology) GetAgentsByRole(role AgentRole) []*Agent {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	agents := bt.agentsByRole[role]
	result := make([]*Agent, len(agents))
	copy(result, agents)
	return result
}

// GetChannels returns all channels.
func (bt *BaseTopology) GetChannels() []CommunicationChannel {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	channels := make([]CommunicationChannel, len(bt.channels))
	copy(channels, bt.channels)
	return channels
}

// AddChannel adds a communication channel.
func (bt *BaseTopology) AddChannel(channel CommunicationChannel) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.channels = append(bt.channels, channel)
}

// AssignRole assigns a role to an agent.
func (bt *BaseTopology) AssignRole(agentID string, role AgentRole) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	agent, ok := bt.agents[agentID]
	if !ok {
		return ErrAgentNotFound
	}

	// Remove from old role list
	oldRoleAgents := bt.agentsByRole[agent.Role]
	for i, a := range oldRoleAgents {
		if a.ID == agentID {
			bt.agentsByRole[agent.Role] = append(oldRoleAgents[:i], oldRoleAgents[i+1:]...)
			break
		}
	}

	// Assign new role
	agent.Role = role
	bt.agentsByRole[role] = append(bt.agentsByRole[role], agent)

	return nil
}

// GetMetrics returns topology metrics.
func (bt *BaseTopology) GetMetrics() TopologyMetrics {
	bt.metricsMu.RLock()
	defer bt.metricsMu.RUnlock()

	// Copy metrics to avoid race conditions
	metrics := TopologyMetrics{
		TotalMessages:      bt.metrics.TotalMessages,
		BroadcastCount:     bt.metrics.BroadcastCount,
		DirectMessageCount: bt.metrics.DirectMessageCount,
		AvgMessageLatency:  bt.metrics.AvgMessageLatency,
		PhaseTransitions:   bt.metrics.PhaseTransitions,
		AgentMetrics:       make(map[string]AgentMetrics),
		ChannelUtilization: make(map[string]float64),
	}

	for k, v := range bt.metrics.AgentMetrics {
		metrics.AgentMetrics[k] = v
	}
	for k, v := range bt.metrics.ChannelUtilization {
		metrics.ChannelUtilization[k] = v
	}

	return metrics
}

// IncrementMessageCount increments the message counter.
func (bt *BaseTopology) IncrementMessageCount(isBroadcast bool) {
	bt.metricsMu.Lock()
	defer bt.metricsMu.Unlock()

	bt.metrics.TotalMessages++
	if isBroadcast {
		bt.metrics.BroadcastCount++
	} else {
		bt.metrics.DirectMessageCount++
	}
}

// IncrementPhaseTransitions increments the phase transition counter.
func (bt *BaseTopology) IncrementPhaseTransitions() {
	bt.metricsMu.Lock()
	defer bt.metricsMu.Unlock()

	bt.metrics.PhaseTransitions++
}

// UpdateAgentMetrics updates metrics for an agent.
func (bt *BaseTopology) UpdateAgentMetrics(agentID string, metrics AgentMetrics) {
	bt.metricsMu.Lock()
	defer bt.metricsMu.Unlock()

	bt.metrics.AgentMetrics[agentID] = metrics
}

// Close closes the topology.
func (bt *BaseTopology) Close() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.closed {
		return nil
	}

	bt.closed = true
	close(bt.msgChan)
	return nil
}

// GetNextPhase returns the next debate phase.
func (bt *BaseTopology) GetNextPhase(currentPhase DebatePhase) DebatePhase {
	bt.IncrementPhaseTransitions()

	switch currentPhase {
	case PhaseDehallucination:
		return PhaseSelfEvolvement
	case PhaseSelfEvolvement:
		return PhaseProposal
	case PhaseProposal:
		return PhaseCritique
	case PhaseCritique:
		return PhaseReview
	case PhaseReview:
		return PhaseOptimization
	case PhaseOptimization:
		return PhaseAdversarial
	case PhaseAdversarial:
		return PhaseConvergence
	case PhaseConvergence:
		return PhaseDehallucination // Cycle back if needed
	default:
		return PhaseProposal
	}
}

// Errors for topology operations.
var (
	ErrAgentNotFound   = TopologyError{Code: "AGENT_NOT_FOUND", Message: "agent not found in topology"}
	ErrChannelNotFound = TopologyError{Code: "CHANNEL_NOT_FOUND", Message: "communication channel not found"}
	ErrTopologyClosed  = TopologyError{Code: "TOPOLOGY_CLOSED", Message: "topology has been closed"}
	ErrInvalidMessage  = TopologyError{Code: "INVALID_MESSAGE", Message: "invalid message format"}
	ErrRoutingFailed   = TopologyError{Code: "ROUTING_FAILED", Message: "message routing failed"}
)

// TopologyError represents a topology-specific error.
type TopologyError struct {
	Code    string
	Message string
}

func (e TopologyError) Error() string {
	return e.Message
}
