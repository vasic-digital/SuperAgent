// Package clis provides CLI agent integration types and interfaces for HelixAgent.
// This package defines the common types used across all CLI agent implementations.
package clis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AgentType represents the type of CLI agent.
type AgentType string

// Supported agent types.
const (
	TypeAider          AgentType = "aider"
	TypeClaudeCode     AgentType = "claude_code"
	TypeCodex          AgentType = "codex"
	TypeCline          AgentType = "cline"
	TypeOpenHands      AgentType = "openhands"
	TypeKiro           AgentType = "kiro"
	TypeContinue       AgentType = "continue"
	TypeSupermaven     AgentType = "supermaven"
	TypeCursor         AgentType = "cursor"
	TypeWindsurf       AgentType = "windsurf"
	TypeAugment        AgentType = "augment"
	TypeSourcegraph    AgentType = "sourcegraph"
	TypeCodeium        AgentType = "codeium"
	TypeTabnine        AgentType = "tabnine"
	TypeCodeGPT        AgentType = "codegpt"
	TypeTwin           AgentType = "twin"
	TypeDevin          AgentType = "devin"
	TypeDevika         AgentType = "devika"
	TypeSWEAgent       AgentType = "swe_agent"
	TypeGPTPilot       AgentType = "gpt_pilot"
	TypeMetamorph      AgentType = "metamorph"
	TypeJunie          AgentType = "junie"
	TypeAmazonQ        AgentType = "amazon_q"
	TypeGitHubCopilot  AgentType = "github_copilot"
	TypeJetBrainsAI    AgentType = "jetbrains_ai"
	TypeCodeGemma      AgentType = "codegemma"
	TypeStarCoder      AgentType = "starcoder"
	TypeQwenCoder      AgentType = "qwen_coder"
	TypeMistralCode    AgentType = "mistral_code"
	TypeGeminiAssist   AgentType = "gemini_assist"
	TypeCodey          AgentType = "codey"
	TypeLlamaCode      AgentType = "llama_code"
	TypeDeepSeekCoder  AgentType = "deepseek_coder"
	TypeWizardCoder    AgentType = "wizardcoder"
	TypePhind          AgentType = "phind"
	TypeCody           AgentType = "cody"
	TypeCursorSh       AgentType = "cursor_sh"
	TypeTrae           AgentType = "trae"
	TypeBlackbox       AgentType = "blackbox"
	TypeLovable        AgentType = "lovable"
	TypeV0             AgentType = "v0"
	TypeTempo          AgentType = "tempo"
	TypeBolt           AgentType = "bolt"
	TypeReplitAgent    AgentType = "replit_agent"
	TypeIDX            AgentType = "idx"
	TypeFirebaseStudio AgentType = "firebase_studio"
	TypeCascade        AgentType = "cascade"
	TypeHelixAgent     AgentType = "helixagent"
)

// InstanceStatus represents the lifecycle state of an agent instance.
type InstanceStatus string

// Instance lifecycle states.
const (
	StatusCreating    InstanceStatus = "creating"
	StatusIdle        InstanceStatus = "idle"
	StatusActive      InstanceStatus = "active"
	StatusBackground  InstanceStatus = "background"
	StatusDegraded    InstanceStatus = "degraded"
	StatusRecovering  InstanceStatus = "recovering"
	StatusTerminating InstanceStatus = "terminating"
	StatusTerminated  InstanceStatus = "terminated"
	StatusFailed      InstanceStatus = "failed"
)

// HealthStatus represents the health state of an instance.
type HealthStatus string

// Health states.
const (
	HealthHealthy   HealthStatus = "healthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthUnknown   HealthStatus = "unknown"
)

// RequestType represents the type of request sent to an instance.
type RequestType string

// Request types.
const (
	RequestTypeExecute  RequestType = "execute"
	RequestTypeQuery    RequestType = "query"
	RequestTypeToolCall RequestType = "tool_call"
	RequestTypeStream   RequestType = "stream"
	RequestTypeHealth   RequestType = "health"
	RequestTypeCancel   RequestType = "cancel"
)

// EventType represents the type of event emitted by an instance.
type EventType string

// Event types.
const (
	EventTypeStatus   EventType = "status"
	EventTypeProgress EventType = "progress"
	EventTypeError    EventType = "error"
	EventTypeComplete EventType = "complete"
	EventTypeLog      EventType = "log"
	EventTypeMetrics  EventType = "metrics"
)

// MessageType represents the type of inter-agent message.
type MessageType string

// Message types.
const (
	MessageTypeRequest    MessageType = "request"
	MessageTypeResponse   MessageType = "response"
	MessageTypeEvent      MessageType = "event"
	MessageTypeHeartbeat  MessageType = "heartbeat"
	MessageTypeCommand    MessageType = "command"
	MessageTypeResult     MessageType = "result"
	MessageTypeError      MessageType = "error"
)

// AgentInstance represents a running CLI agent instance.
type AgentInstance struct {
	// Identification
	ID   string    `json:"id"`
	Type AgentType `json:"type"`
	Name string    `json:"name"`

	// Lifecycle state
	Status        InstanceStatus         `json:"status"`
	Health        HealthStatus           `json:"health"`
	HealthDetails map[string]interface{} `json:"health_details,omitempty"`

	// Configuration
	Config   InstanceConfig `json:"config"`
	Provider ProviderConfig `json:"provider"`

	// Resource limits
	Resources ResourceLimits `json:"resources"`

	// Current execution context
	SessionID string `json:"session_id,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	Workspace string `json:"workspace,omitempty"`

	// Runtime state (not persisted)
	State map[string]interface{} `json:"state,omitempty"`

	// Metrics
	RequestsProcessed int   `json:"requests_processed"`
	ErrorsCount       int   `json:"errors_count"`
	TotalExecTimeMs   int64 `json:"total_execution_time_ms"`

	// Communication channels (not serialized)
	RequestCh  chan *Request  `json:"-"`
	ResponseCh chan *Response `json:"-"`
	EventCh    chan *Event    `json:"-"`

	// Lifecycle timestamps
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	TerminatedAt *time.Time `json:"terminated_at,omitempty"`

	// Last health check
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
}

// InstanceConfig contains configuration for an agent instance.
type InstanceConfig struct {
	// Resource limits
	MaxMemoryMB   int `json:"max_memory_mb"`
	MaxCPUPercent int `json:"max_cpu_percent"`
	MaxDiskMB     int `json:"max_disk_mb"`

	// Timeouts
	RequestTimeout      time.Duration `json:"request_timeout"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	MaxLifetime         time.Duration `json:"max_lifetime"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`

	// Feature flags
	EnableStreaming bool `json:"enable_streaming"`
	EnableCaching   bool `json:"enable_caching"`
	EnableMetrics   bool `json:"enable_metrics"`

	// Type-specific configuration
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// DefaultInstanceConfig returns default configuration.
func DefaultInstanceConfig() InstanceConfig {
	return InstanceConfig{
		MaxMemoryMB:         2048,
		MaxCPUPercent:       100,
		MaxDiskMB:           10240,
		RequestTimeout:      5 * time.Minute,
		IdleTimeout:         30 * time.Minute,
		MaxLifetime:         24 * time.Hour,
		HealthCheckInterval: 30 * time.Second,
		EnableStreaming:     true,
		EnableCaching:       true,
		EnableMetrics:       true,
	}
}

// ProviderConfig contains LLM provider configuration.
type ProviderConfig struct {
	Name       string                 `json:"name"`
	APIKey     string                 `json:"api_key,omitempty"`
	BaseURL    string                 `json:"base_url,omitempty"`
	Model      string                 `json:"model"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ResourceLimits defines resource constraints for an instance.
type ResourceLimits struct {
	MemoryMB   int64   `json:"memory_mb"`
	CPUPercent float64 `json:"cpu_percent"`
	DiskMB     int64   `json:"disk_mb"`
}

// Request represents a request sent to an agent instance.
type Request struct {
	ID      string      `json:"id"`
	Type    RequestType `json:"type"`
	Payload interface{} `json:"payload"`
	Timeout time.Duration `json:"timeout"`

	// Context information
	SessionID string `json:"session_id,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}

// Response represents a response from an agent instance.
type Response struct {
	RequestID string                 `json:"request_id"`
	Success   bool                   `json:"success"`
	Result    interface{}            `json:"result,omitempty"`
	Error     *ErrorDetail           `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *ErrorDetail) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Event represents an event emitted by an agent instance.
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Payload   interface{}            `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents an inter-agent communication message.
type Message struct {
	ID        string    `json:"id"`
	Type      MessageType `json:"type"`
	SessionID string    `json:"session_id"`

	// Routing
	SourceID   string    `json:"source_id"`
	SourceType AgentType `json:"source_type"`
	TargetID   string    `json:"target_id,omitempty"`
	TargetType AgentType `json:"target_type,omitempty"`
	Broadcast  bool      `json:"broadcast"`

	// Content
	Payload   interface{} `json:"payload"`
	SizeBytes int         `json:"size_bytes,omitempty"`

	// Status
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	ReceivedAt  *time.Time `json:"received_at,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// Task represents a background task.
type Task struct {
	ID       string      `json:"id"`
	Type     TaskType    `json:"type"`
	Name     string      `json:"name"`
	Payload  interface{} `json:"payload"`
	Priority int         `json:"priority"`

	// Assignment
	AssignedInstanceID string `json:"assigned_instance_id,omitempty"`

	// Progress
	Status          TaskStatus  `json:"status"`
	ProgressPercent int         `json:"progress_percent"`
	Result          interface{} `json:"result,omitempty"`
	ErrorMessage    string      `json:"error_message,omitempty"`

	// Retry logic
	RetryCount int `json:"retry_count"`
	MaxRetries int `json:"max_retries"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// TaskType represents the type of background task.
type TaskType string

// Task types.
const (
	TaskTypeGitOperation  TaskType = "git_operation"
	TaskTypeCodeAnalysis  TaskType = "code_analysis"
	TaskTypeDocumentation TaskType = "documentation"
	TaskTypeTesting       TaskType = "testing"
	TaskTypeLinting       TaskType = "linting"
	TaskTypeBuild         TaskType = "build"
	TaskTypeDeploy        TaskType = "deploy"
	TaskTypeCodeReview    TaskType = "code_review"
	TaskTypeRefactoring   TaskType = "refactoring"
	TaskTypeOptimization  TaskType = "optimization"
)

// TaskStatus represents the status of a task.
type TaskStatus string

// Task statuses.
const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusExpired   TaskStatus = "expired"
)

// EnsembleStrategy represents the coordination strategy for ensemble execution.
type EnsembleStrategy string

// Ensemble strategies.
const (
	StrategyVoting     EnsembleStrategy = "voting"
	StrategyDebate     EnsembleStrategy = "debate"
	StrategyConsensus  EnsembleStrategy = "consensus"
	StrategyPipeline   EnsembleStrategy = "pipeline"
	StrategyParallel   EnsembleStrategy = "parallel"
	StrategySequential EnsembleStrategy = "sequential"
	StrategyExpertPanel EnsembleStrategy = "expert_panel"
)

// EnsembleSession represents a multi-instance ensemble execution session.
type EnsembleSession struct {
	ID string `json:"id"`

	// Strategy configuration
	Strategy EnsembleStrategy `json:"strategy"`
	Config   EnsembleConfig   `json:"config"`

	// Participants
	ParticipantTypes        []AgentType `json:"participant_types"`
	PrimaryInstanceID       string      `json:"primary_instance_id,omitempty"`
	CritiqueInstanceIDs     []string    `json:"critique_instance_ids,omitempty"`
	VerificationInstanceIDs []string    `json:"verification_instance_ids,omitempty"`
	FallbackInstanceIDs     []string    `json:"fallback_instance_ids,omitempty"`

	// Status
	Status string `json:"status"`

	// Context and results
	Context             map[string]interface{} `json:"context,omitempty"`
	TaskDefinition      map[string]interface{} `json:"task_definition"`
	IntermediateResults map[string]interface{} `json:"intermediate_results,omitempty"`
	FinalResult         interface{}            `json:"final_result,omitempty"`

	// Consensus tracking
	ConsensusReached bool                   `json:"consensus_reached,omitempty"`
	ConfidenceScore  float64                `json:"confidence_score,omitempty"`
	VotingResults    map[string]interface{} `json:"voting_results,omitempty"`

	// Metrics
	TotalDurationMs int64 `json:"total_duration_ms,omitempty"`
	TokensConsumed  int   `json:"tokens_consumed,omitempty"`
	APICalls        int   `json:"api_calls,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// EnsembleConfig contains ensemble execution configuration.
type EnsembleConfig struct {
	// Voting/consensus settings
	MinParticipants      int     `json:"min_participants"`
	ConsensusThreshold   float64 `json:"consensus_threshold"`
	MaxRounds            int     `json:"max_rounds"`

	// Timeout settings
	TimeoutPerRound time.Duration `json:"timeout_per_round"`
	TotalTimeout    time.Duration `json:"total_timeout"`

	// Feature flags
	EnableStreaming  bool `json:"enable_streaming"`
	EnableFallbacks  bool `json:"enable_fallbacks"`
	RequireConsensus bool `json:"require_consensus"`
}

// DefaultEnsembleConfig returns default ensemble configuration.
func DefaultEnsembleConfig() EnsembleConfig {
	return EnsembleConfig{
		MinParticipants:    2,
		ConsensusThreshold: 0.6,
		MaxRounds:          3,
		TimeoutPerRound:    5 * time.Minute,
		TotalTimeout:       15 * time.Minute,
		EnableStreaming:    true,
		EnableFallbacks:    true,
		RequireConsensus:   false,
	}
}

// Feature represents a CLI agent feature in the registry.
type Feature struct {
	ID string `json:"id"`

	// Identification
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`

	// Source information
	SourceAgent string `json:"source_agent"`
	SourceFile  string `json:"source_file,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`

	// Implementation details
	ImplementationType  string                 `json:"implementation_type"`
	InternalPath        string                 `json:"internal_path,omitempty"`
	InterfaceDefinition map[string]interface{} `json:"interface_definition,omitempty"`

	// Dependencies
	Dependencies      []string `json:"dependencies,omitempty"`
	ExternalDeps      []string `json:"external_deps,omitempty"`
	RequiredProviders []string `json:"required_providers,omitempty"`

	// Status
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	Complexity  string `json:"complexity,omitempty"`
	EffortHours int    `json:"estimated_effort_hours,omitempty"`

	// Notes
	PortingNotes      string `json:"porting_notes,omitempty"`
	PortingChallenges string `json:"porting_challenges,omitempty"`

	// Metrics
	TestCoverage float64 `json:"test_coverage,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CacheEntry represents a cached response.
type CacheEntry struct {
	Key       string        `json:"key"`
	Value     interface{}   `json:"value"`
	Embedding []float32     `json:"embedding,omitempty"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time     `json:"created_at"`
}

// Metric represents a performance metric.
type Metric struct {
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	RecordedAt time.Time         `json:"recorded_at"`
}

// HealthCheckResult represents the result of a health check.
type HealthCheckResult struct {
	Healthy   bool                   `json:"healthy"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CheckedAt time.Time              `json:"checked_at"`
}

// ExecutionResult represents the result of executing a task.
type ExecutionResult struct {
	Success  bool                   `json:"success"`
	Result   interface{}            `json:"result,omitempty"`
	Error    *ErrorDetail           `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToJSON serializes the instance to JSON.
func (i *AgentInstance) ToJSON() ([]byte, error) {
	return json.Marshal(i)
}

// FromJSON deserializes an instance from JSON.
func (i *AgentInstance) FromJSON(data []byte) error {
	return json.Unmarshal(data, i)
}

// IsActive returns true if the instance is in an active state.
func (i *AgentInstance) IsActive() bool {
	return i.Status == StatusActive || i.Status == StatusIdle || i.Status == StatusBackground
}

// IsHealthy returns true if the instance is healthy.
func (i *AgentInstance) IsHealthy() bool {
	return i.Health == HealthHealthy
}

// CanAcceptWork returns true if the instance can accept new work.
func (i *AgentInstance) CanAcceptWork() bool {
	return i.IsActive() && i.IsHealthy() && i.Status != StatusDegraded
}

// Age returns how long the instance has been running.
func (i *AgentInstance) Age() time.Duration {
	return time.Since(i.CreatedAt)
}

// Uptime returns how long the instance has been in an active state.
func (i *AgentInstance) Uptime() time.Duration {
	if i.StartedAt == nil {
		return 0
	}
	return time.Since(*i.StartedAt)
}

// String returns a string representation of the instance.
func (i *AgentInstance) String() string {
	return fmt.Sprintf("AgentInstance[%s|%s|%s|%s]", i.ID, i.Type, i.Name, i.Status)
}

// String returns a string representation of the request.
func (r *Request) String() string {
	return fmt.Sprintf("Request[%s|%s]", r.ID, r.Type)
}

// String returns a string representation of the response.
func (r *Response) String() string {
	status := "success"
	if !r.Success {
		status = "failed"
	}
	return fmt.Sprintf("Response[%s|%s|%v]", r.RequestID, status, r.Duration)
}

// String returns a string representation of the event.
func (e *Event) String() string {
	return fmt.Sprintf("Event[%s|%s|%s]", e.ID, e.Type, e.Source)
}
