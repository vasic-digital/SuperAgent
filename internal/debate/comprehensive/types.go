// Package comprehensive provides the complete multi-agent debate system
// as specified in the debate documentation.
package comprehensive

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Role represents an agent's specialized role in the debate
type Role string

const (
	RoleArchitect   Role = "architect"   // System design and planning
	RoleGenerator   Role = "generator"   // Code generation
	RoleCritic      Role = "critic"      // Flaw identification
	RoleRefactoring Role = "refactoring" // Code improvement
	RoleTester      Role = "tester"      // Test generation
	RoleValidator   Role = "validator"   // Correctness verification
	RoleSecurity    Role = "security"    // Security analysis
	RolePerformance Role = "performance" // Performance optimization
	RoleModerator   Role = "moderator"   // Debate facilitation
	RoleRedTeam     Role = "red_team"    // Adversarial testing
	RoleBlueTeam    Role = "blue_team"   // Defensive implementation
)

// AllRoles returns all available agent roles
func AllRoles() []Role {
	return []Role{
		RoleArchitect,
		RoleGenerator,
		RoleCritic,
		RoleRefactoring,
		RoleTester,
		RoleValidator,
		RoleSecurity,
		RolePerformance,
		RoleModerator,
		RoleRedTeam,
		RoleBlueTeam,
	}
}

// IsValid checks if a role is valid
func (r Role) IsValid() bool {
	for _, role := range AllRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// String returns the string representation
func (r Role) String() string {
	return string(r)
}

// Agent represents a specialized LLM agent in the debate system
type Agent struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Role         Role                   `json:"role"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	Score        float64                `json:"score"`
	Capabilities []Capability           `json:"capabilities"`
	Config       map[string]interface{} `json:"config"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActiveAt time.Time              `json:"last_active_at"`
}

// Capability represents an agent capability
type Capability string

const (
	CapabilityCodeGeneration    Capability = "code_generation"
	CapabilityCodeReview        Capability = "code_review"
	CapabilityTesting           Capability = "testing"
	CapabilitySecurityAnalysis  Capability = "security_analysis"
	CapabilityPerformanceTuning Capability = "performance_tuning"
	CapabilityArchitecture      Capability = "architecture"
	CapabilityRefactoring       Capability = "refactoring"
	CapabilityValidation        Capability = "validation"
	CapabilityToolUse           Capability = "tool_use"
)

// NewAgent creates a new agent with the given role
func NewAgent(role Role, provider, model string, score float64) *Agent {
	return &Agent{
		ID:           uuid.New().String(),
		Name:         fmt.Sprintf("%s-%s", role, uuid.New().String()[:8]),
		Role:         role,
		Provider:     provider,
		Model:        model,
		Score:        score,
		Capabilities: DefaultCapabilitiesForRole(role),
		Config:       make(map[string]interface{}),
		IsActive:     true,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
}

// DefaultCapabilitiesForRole returns default capabilities for a role
func DefaultCapabilitiesForRole(role Role) []Capability {
	switch role {
	case RoleArchitect:
		return []Capability{CapabilityArchitecture, CapabilityValidation}
	case RoleGenerator:
		return []Capability{CapabilityCodeGeneration, CapabilityToolUse}
	case RoleCritic:
		return []Capability{CapabilityCodeReview, CapabilitySecurityAnalysis}
	case RoleRefactoring:
		return []Capability{CapabilityRefactoring, CapabilityCodeReview}
	case RoleTester:
		return []Capability{CapabilityTesting, CapabilityValidation}
	case RoleValidator:
		return []Capability{CapabilityValidation, CapabilityTesting}
	case RoleSecurity:
		return []Capability{CapabilitySecurityAnalysis}
	case RolePerformance:
		return []Capability{CapabilityPerformanceTuning}
	case RoleRedTeam:
		return []Capability{CapabilitySecurityAnalysis, CapabilityTesting}
	case RoleBlueTeam:
		return []Capability{CapabilityCodeGeneration, CapabilitySecurityAnalysis}
	default:
		return []Capability{}
	}
}

// HasCapability checks if agent has a specific capability
func (a *Agent) HasCapability(cap Capability) bool {
	for _, c := range a.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// UpdateActivity updates the last active timestamp
func (a *Agent) UpdateActivity() {
	a.LastActiveAt = time.Now()
}

// AgentResponse represents a response from an agent
type AgentResponse struct {
	AgentID    string                 `json:"agent_id"`
	AgentRole  Role                   `json:"agent_role"`
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	Content    string                 `json:"content"`
	Confidence float64                `json:"confidence"`
	Score      float64                `json:"score"`
	ToolsUsed  []string               `json:"tools_used"`
	Metadata   map[string]interface{} `json:"metadata"`
	Latency    time.Duration          `json:"latency"`
	Timestamp  time.Time              `json:"timestamp"`
}

// NewAgentResponse creates a new agent response
func NewAgentResponse(agent *Agent, content string, confidence float64) *AgentResponse {
	return &AgentResponse{
		AgentID:    agent.ID,
		AgentRole:  agent.Role,
		Provider:   agent.Provider,
		Model:      agent.Model,
		Content:    content,
		Confidence: confidence,
		ToolsUsed:  make([]string, 0),
		Metadata:   make(map[string]interface{}),
		Timestamp:  time.Now(),
	}
}

// Message represents a message between agents
type Message struct {
	ID          string                 `json:"id"`
	FromAgentID string                 `json:"from_agent_id"`
	ToAgentID   string                 `json:"to_agent_id,omitempty"` // Empty for broadcast
	Type        MessageType            `json:"type"`
	Content     string                 `json:"content"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeProposal   MessageType = "proposal"
	MessageTypeCritique   MessageType = "critique"
	MessageTypeDefense    MessageType = "defense"
	MessageTypeConsensus  MessageType = "consensus"
	MessageTypeQuestion   MessageType = "question"
	MessageTypeResponse   MessageType = "response"
	MessageTypeToolCall   MessageType = "tool_call"
	MessageTypeToolResult MessageType = "tool_result"
	MessageTypeSystem     MessageType = "system"
)

// NewMessage creates a new message
func NewMessage(fromAgentID string, msgType MessageType, content string) *Message {
	return &Message{
		ID:          uuid.New().String(),
		FromAgentID: fromAgentID,
		Type:        msgType,
		Content:     content,
		Context:     make(map[string]interface{}),
		Timestamp:   time.Now(),
	}
}

// ConsensusResult represents the result of a consensus evaluation
type ConsensusResult struct {
	Reached    bool                   `json:"reached"`
	Confidence float64                `json:"confidence"`
	Summary    string                 `json:"summary"`
	KeyPoints  []string               `json:"key_points"`
	Dissents   []string               `json:"dissents"`
	FinalCode  string                 `json:"final_code,omitempty"`
	Votes      map[string]float64     `json:"votes"`
	Metadata   map[string]interface{} `json:"metadata"`
	AchievedAt time.Time              `json:"achieved_at"`
}

// NewConsensusResult creates a new consensus result
func NewConsensusResult() *ConsensusResult {
	return &ConsensusResult{
		Reached:    false,
		Confidence: 0.0,
		KeyPoints:  make([]string, 0),
		Dissents:   make([]string, 0),
		Votes:      make(map[string]float64),
		Metadata:   make(map[string]interface{}),
	}
}

// AddVote adds a vote from an agent
func (c *ConsensusResult) AddVote(agentID string, confidence float64) {
	c.Votes[agentID] = confidence
}

// CalculateConfidence calculates overall confidence from votes
func (c *ConsensusResult) CalculateConfidence() float64 {
	if len(c.Votes) == 0 {
		return 0.0
	}

	total := 0.0
	for _, confidence := range c.Votes {
		total += confidence
	}
	return total / float64(len(c.Votes))
}

// Context represents the shared context for a debate
type Context struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Codebase  string                 `json:"codebase,omitempty"`
	Language  string                 `json:"language"`
	Messages  []*Message             `json:"messages"`
	Responses []*AgentResponse       `json:"responses"`
	Artifacts map[string]*Artifact   `json:"artifacts"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Artifact represents a produced artifact (code, test, etc.)
type Artifact struct {
	ID          string                 `json:"id"`
	Type        ArtifactType           `json:"type"`
	Name        string                 `json:"name"`
	Content     string                 `json:"content"`
	AgentID     string                 `json:"agent_id"`
	Version     int                    `json:"version"`
	IsValidated bool                   `json:"is_validated"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ArtifactType represents the type of artifact
type ArtifactType string

const (
	ArtifactTypeCode   ArtifactType = "code"
	ArtifactTypeTest   ArtifactType = "test"
	ArtifactTypeDesign ArtifactType = "design"
	ArtifactTypeReview ArtifactType = "review"
	ArtifactTypePlan   ArtifactType = "plan"
	ArtifactTypeReport ArtifactType = "report"
)

// NewContext creates a new debate context
func NewContext(topic, codebase, language string) *Context {
	return &Context{
		ID:        uuid.New().String(),
		Topic:     topic,
		Codebase:  codebase,
		Language:  language,
		Messages:  make([]*Message, 0),
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddMessage adds a message to the context
func (c *Context) AddMessage(msg *Message) {
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// AddResponse adds an agent response to the context
func (c *Context) AddResponse(resp *AgentResponse) {
	c.Responses = append(c.Responses, resp)
	c.UpdatedAt = time.Now()
}

// AddArtifact adds an artifact to the context
func (c *Context) AddArtifact(artifact *Artifact) {
	c.Artifacts[artifact.ID] = artifact
	c.UpdatedAt = time.Now()
}

// GetMessagesByType returns messages filtered by type
func (c *Context) GetMessagesByType(msgType MessageType) []*Message {
	var result []*Message
	for _, msg := range c.Messages {
		if msg.Type == msgType {
			result = append(result, msg)
		}
	}
	return result
}

// GetResponsesByRole returns responses filtered by agent role
func (c *Context) GetResponsesByRole(role Role) []*AgentResponse {
	var result []*AgentResponse
	for _, resp := range c.Responses {
		if resp.AgentRole == role {
			result = append(result, resp)
		}
	}
	return result
}

// Score represents a scored evaluation
type Score struct {
	Value       float64                `json:"value"`
	MaxValue    float64                `json:"max_value"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewScore creates a new score
func NewScore(value, maxValue float64, category, description string) *Score {
	return &Score{
		Value:       value,
		MaxValue:    maxValue,
		Category:    category,
		Description: description,
		Metadata:    make(map[string]interface{}),
	}
}

// Percentage returns the score as a percentage
func (s *Score) Percentage() float64 {
	if s.MaxValue == 0 {
		return 0
	}
	return (s.Value / s.MaxValue) * 100
}

// IsPassing checks if the score meets a threshold
func (s *Score) IsPassing(threshold float64) bool {
	return s.Percentage() >= threshold
}

// AgentInterface defines the interface for agent implementations
type AgentInterface interface {
	// GetID returns the agent's unique ID
	GetID() string

	// GetRole returns the agent's role
	GetRole() Role

	// Process processes a message and generates a response
	Process(ctx context.Context, msg *Message, context *Context) (*AgentResponse, error)

	// CanHandle checks if the agent can handle a specific task
	CanHandle(taskType string) bool

	// GetCapabilities returns the agent's capabilities
	GetCapabilities() []Capability

	// UpdateScore updates the agent's score
	UpdateScore(score float64)
}
