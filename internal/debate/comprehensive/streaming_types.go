package comprehensive

import (
	"time"
)

// StreamEventType represents the type of streaming event
type StreamEventType string

const (
	// Debate lifecycle events
	StreamEventDebateStart    StreamEventType = "debate_start"
	StreamEventDebateComplete StreamEventType = "debate_complete"
	StreamEventDebateError    StreamEventType = "debate_error"

	// Phase events
	StreamEventPhaseStart    StreamEventType = "phase_start"
	StreamEventPhaseComplete StreamEventType = "phase_complete"
	StreamEventPhaseError    StreamEventType = "phase_error"

	// Team events
	StreamEventTeamStart    StreamEventType = "team_start"
	StreamEventTeamComplete StreamEventType = "team_complete"

	// Agent events
	StreamEventAgentStart    StreamEventType = "agent_start"
	StreamEventAgentResponse StreamEventType = "agent_response"
	StreamEventAgentError    StreamEventType = "agent_error"
	StreamEventAgentComplete StreamEventType = "agent_complete"
	StreamEventAgentFallback StreamEventType = "agent_fallback"

	// Tool events
	StreamEventToolCall   StreamEventType = "tool_call"
	StreamEventToolResult StreamEventType = "tool_result"

	// Consensus events
	StreamEventConsensusUpdate  StreamEventType = "consensus_update"
	StreamEventConsensusReached StreamEventType = "consensus_reached"
)

// StreamEvent represents a single event in the debate stream
type StreamEvent struct {
	Type      StreamEventType        `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	DebateID  string                 `json:"debate_id"`
	Phase     string                 `json:"phase,omitempty"`
	Team      Team                   `json:"team,omitempty"`
	Agent     *AgentInfo             `json:"agent,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Progress  float64                `json:"progress"`
	Metadata  map[string]interface{} `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
}

// AgentInfo represents agent information for streaming
type AgentInfo struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Role      Role           `json:"role"`
	Provider  string         `json:"provider"`
	Model     string         `json:"model"`
	Fallbacks []FallbackInfo `json:"fallbacks,omitempty"`
}

// FallbackInfo represents fallback provider information
type FallbackInfo struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Reason   string `json:"reason,omitempty"`
}

// StreamHandler is called for each streaming event
type StreamHandler func(event *StreamEvent) error

// DebateStreamRequest extends DebateRequest with streaming support
type DebateStreamRequest struct {
	*DebateRequest
	Stream        bool          `json:"stream"`
	StreamHandler StreamHandler `json:"-"`
}

// Team represents a debate team
type Team string

const (
	TeamDesign         Team = "design"         // Design team: Architect, Moderator
	TeamImplementation Team = "implementation" // Implementation team: Generator, Blue Team
	TeamQuality        Team = "quality"        // Quality team: Critic, Tester, Validator, Security, Performance
	TeamRedTeam        Team = "red_team"       // Red team: Red Team (adversarial)
	TeamRefactoring    Team = "refactoring"    // Refactoring team: Refactoring
)

// AllTeams returns all debate teams
func AllTeams() []Team {
	return []Team{
		TeamDesign,
		TeamImplementation,
		TeamQuality,
		TeamRedTeam,
		TeamRefactoring,
	}
}

// GetTeamForRole returns which team an agent role belongs to
func GetTeamForRole(role Role) Team {
	switch role {
	case RoleArchitect, RoleModerator:
		return TeamDesign
	case RoleGenerator, RoleBlueTeam:
		return TeamImplementation
	case RoleCritic, RoleTester, RoleValidator, RoleSecurity, RolePerformance:
		return TeamQuality
	case RoleRedTeam:
		return TeamRedTeam
	case RoleRefactoring:
		return TeamRefactoring
	default:
		return TeamQuality
	}
}

// GetRolesForTeam returns all roles in a team
func GetRolesForTeam(team Team) []Role {
	switch team {
	case TeamDesign:
		return []Role{RoleArchitect, RoleModerator}
	case TeamImplementation:
		return []Role{RoleGenerator, RoleBlueTeam}
	case TeamQuality:
		return []Role{RoleCritic, RoleTester, RoleValidator, RoleSecurity, RolePerformance}
	case TeamRedTeam:
		return []Role{RoleRedTeam}
	case TeamRefactoring:
		return []Role{RoleRefactoring}
	default:
		return []Role{}
	}
}

// StreamConfig configures streaming behavior
type StreamConfig struct {
	BufferSize     int           `json:"buffer_size"`
	FlushInterval  time.Duration `json:"flush_interval"`
	EnableProgress bool          `json:"enable_progress"`
	EnableTeams    bool          `json:"enable_teams"`
	EnablePhases   bool          `json:"enable_phases"`
}

// DefaultStreamConfig returns default streaming configuration
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		BufferSize:     100,
		FlushInterval:  100 * time.Millisecond,
		EnableProgress: true,
		EnableTeams:    true,
		EnablePhases:   true,
	}
}

// StreamOrchestrator orchestrates debate streaming
type StreamOrchestrator struct {
	config          *StreamConfig
	handler         StreamHandler
	debateID        string
	startTime       time.Time
	totalAgents     int
	completedAgents int
	currentPhase    string
	currentTeam     Team
}

// NewStreamOrchestrator creates a new streaming orchestrator
func NewStreamOrchestrator(config *StreamConfig, handler StreamHandler, debateID string, totalAgents int) *StreamOrchestrator {
	if config == nil {
		config = DefaultStreamConfig()
	}
	return &StreamOrchestrator{
		config:      config,
		handler:     handler,
		debateID:    debateID,
		startTime:   time.Now(),
		totalAgents: totalAgents,
	}
}

// SetPhase sets the current phase
func (so *StreamOrchestrator) SetPhase(phase string) {
	so.currentPhase = phase
}

// SetTeam sets the current team
func (so *StreamOrchestrator) SetTeam(team Team) {
	so.currentTeam = team
}

// Emit sends a streaming event
func (so *StreamOrchestrator) Emit(eventType StreamEventType, agent *Agent, content string, metadata map[string]interface{}) error {
	event := &StreamEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		DebateID:  so.debateID,
		Phase:     so.currentPhase,
		Team:      so.currentTeam,
		Content:   content,
		Progress:  so.calculateProgress(),
		Metadata:  metadata,
	}

	if agent != nil {
		event.Agent = &AgentInfo{
			ID:       agent.ID,
			Name:     agent.Name,
			Role:     agent.Role,
			Provider: agent.Provider,
			Model:    agent.Model,
		}
	}

	if so.handler != nil {
		return so.handler(event)
	}
	return nil
}

// EmitError emits an error event
func (so *StreamOrchestrator) EmitError(agent *Agent, err error) error {
	event := &StreamEvent{
		Type:      StreamEventDebateError,
		Timestamp: time.Now(),
		DebateID:  so.debateID,
		Phase:     so.currentPhase,
		Team:      so.currentTeam,
		Progress:  so.calculateProgress(),
		Error:     err.Error(),
	}

	if agent != nil {
		event.Agent = &AgentInfo{
			ID:       agent.ID,
			Name:     agent.Name,
			Role:     agent.Role,
			Provider: agent.Provider,
			Model:    agent.Model,
		}
	}

	if so.handler != nil {
		return so.handler(event)
	}
	return nil
}

// calculateProgress calculates current progress (0-100)
func (so *StreamOrchestrator) calculateProgress() float64 {
	if so.totalAgents == 0 {
		return 0
	}
	return float64(so.completedAgents) / float64(so.totalAgents) * 100
}

// IncrementCompleted increments completed agent count
func (so *StreamOrchestrator) IncrementCompleted() {
	so.completedAgents++
}

// GetDuration returns elapsed duration
func (so *StreamOrchestrator) GetDuration() time.Duration {
	return time.Since(so.startTime)
}

// AgentTeamAssignment tracks which agents are in which teams
type AgentTeamAssignment struct {
	Agent *Agent
	Team  Team
}

// AssignTeamsToAgents assigns team affiliations to a list of agents
func AssignTeamsToAgents(agents []*Agent) []*AgentTeamAssignment {
	assignments := make([]*AgentTeamAssignment, 0, len(agents))
	for _, agent := range agents {
		assignments = append(assignments, &AgentTeamAssignment{
			Agent: agent,
			Team:  GetTeamForRole(agent.Role),
		})
	}
	return assignments
}

// GetAgentsByTeam returns all agents assigned to a specific team
func GetAgentsByTeam(assignments []*AgentTeamAssignment, team Team) []*Agent {
	var agents []*Agent
	for _, assignment := range assignments {
		if assignment.Team == team {
			agents = append(agents, assignment.Agent)
		}
	}
	return agents
}

// GetTeamSummary returns a summary of all teams and their agents
func GetTeamSummary(assignments []*AgentTeamAssignment) map[Team][]*Agent {
	summary := make(map[Team][]*Agent)
	for _, team := range AllTeams() {
		summary[team] = GetAgentsByTeam(assignments, team)
	}
	return summary
}
