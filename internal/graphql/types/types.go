// Package types defines GraphQL types for the HelixAgent API.
package types

import (
	"time"
)

// Provider represents an LLM provider.
type Provider struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"` // api_key, oauth, free
	Status       string         `json:"status"` // active, inactive, degraded
	Score        float64        `json:"score"`
	Models       []Model        `json:"models"`
	HealthStatus *HealthStatus  `json:"health_status"`
	Capabilities *Capabilities  `json:"capabilities"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Model represents an LLM model.
type Model struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	ProviderID      string     `json:"provider_id"`
	Version         string     `json:"version"`
	ContextWindow   int        `json:"context_window"`
	MaxTokens       int        `json:"max_tokens"`
	SupportsTools   bool       `json:"supports_tools"`
	SupportsVision  bool       `json:"supports_vision"`
	SupportsStreaming bool     `json:"supports_streaming"`
	Score           float64    `json:"score"`
	Rank            int        `json:"rank"`
	CreatedAt       time.Time  `json:"created_at"`
}

// HealthStatus represents the health status of a provider.
type HealthStatus struct {
	Status       string    `json:"status"` // healthy, degraded, unhealthy
	Latency      int64     `json:"latency_ms"`
	LastCheck    time.Time `json:"last_check"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Capabilities represents provider capabilities.
type Capabilities struct {
	Chat            bool `json:"chat"`
	Completions     bool `json:"completions"`
	Embeddings      bool `json:"embeddings"`
	Vision          bool `json:"vision"`
	ToolUse         bool `json:"tool_use"`
	Streaming       bool `json:"streaming"`
	FunctionCalling bool `json:"function_calling"`
}

// Debate represents an AI debate session.
type Debate struct {
	ID           string        `json:"id"`
	Topic        string        `json:"topic"`
	Status       string        `json:"status"` // pending, running, completed, failed
	Participants []Participant `json:"participants"`
	Rounds       []DebateRound `json:"rounds"`
	Conclusion   string        `json:"conclusion,omitempty"`
	Confidence   float64       `json:"confidence"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
}

// Participant represents a debate participant.
type Participant struct {
	ID         string  `json:"id"`
	ProviderID string  `json:"provider_id"`
	ModelID    string  `json:"model_id"`
	Position   string  `json:"position"`
	Role       string  `json:"role"` // primary, fallback
	Score      float64 `json:"score"`
}

// DebateRound represents a round in a debate.
type DebateRound struct {
	ID          string     `json:"id"`
	DebateID    string     `json:"debate_id"`
	RoundNumber int        `json:"round_number"`
	Responses   []Response `json:"responses"`
	Summary     string     `json:"summary,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Response represents a participant's response in a round.
type Response struct {
	ParticipantID string    `json:"participant_id"`
	Content       string    `json:"content"`
	Confidence    float64   `json:"confidence"`
	TokenCount    int       `json:"token_count"`
	Latency       int64     `json:"latency_ms"`
	CreatedAt     time.Time `json:"created_at"`
}

// Task represents a background task.
type Task struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Status      string     `json:"status"` // pending, queued, running, completed, failed
	Priority    int        `json:"priority"`
	Progress    int        `json:"progress"` // 0-100
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// VerificationResults represents the verification results.
type VerificationResults struct {
	TotalProviders   int       `json:"total_providers"`
	VerifiedProviders int      `json:"verified_providers"`
	TotalModels      int       `json:"total_models"`
	VerifiedModels   int       `json:"verified_models"`
	OverallScore     float64   `json:"overall_score"`
	LastVerified     time.Time `json:"last_verified"`
}

// ProviderScore represents a provider's score breakdown.
type ProviderScore struct {
	ProviderID        string  `json:"provider_id"`
	ProviderName      string  `json:"provider_name"`
	OverallScore      float64 `json:"overall_score"`
	ResponseSpeed     float64 `json:"response_speed"`
	ModelEfficiency   float64 `json:"model_efficiency"`
	CostEffectiveness float64 `json:"cost_effectiveness"`
	Capability        float64 `json:"capability"`
	Recency           float64 `json:"recency"`
}

// Connection represents a paginated connection (GraphQL Relay pattern).
type Connection[T any] struct {
	Edges      []Edge[T] `json:"edges"`
	PageInfo   PageInfo  `json:"page_info"`
	TotalCount int       `json:"total_count"`
}

// Edge represents an edge in a connection.
type Edge[T any] struct {
	Node   T      `json:"node"`
	Cursor string `json:"cursor"`
}

// PageInfo represents pagination information.
type PageInfo struct {
	HasNextPage     bool   `json:"has_next_page"`
	HasPreviousPage bool   `json:"has_previous_page"`
	StartCursor     string `json:"start_cursor,omitempty"`
	EndCursor       string `json:"end_cursor,omitempty"`
}

// Input Types

// ProviderFilter represents filter options for providers.
type ProviderFilter struct {
	Status    *string  `json:"status,omitempty"`
	Type      *string  `json:"type,omitempty"`
	MinScore  *float64 `json:"min_score,omitempty"`
	MaxScore  *float64 `json:"max_score,omitempty"`
}

// DebateFilter represents filter options for debates.
type DebateFilter struct {
	Status *string `json:"status,omitempty"`
}

// TaskFilter represents filter options for tasks.
type TaskFilter struct {
	Status *string `json:"status,omitempty"`
	Type   *string `json:"type,omitempty"`
}

// CreateDebateInput represents input for creating a debate.
type CreateDebateInput struct {
	Topic        string   `json:"topic"`
	Participants []string `json:"participants,omitempty"` // Provider IDs
	RoundCount   int      `json:"round_count,omitempty"`
}

// DebateResponseInput represents input for submitting a debate response.
type DebateResponseInput struct {
	DebateID      string `json:"debate_id"`
	ParticipantID string `json:"participant_id"`
	Content       string `json:"content"`
}

// CreateTaskInput represents input for creating a task.
type CreateTaskInput struct {
	Type     string                 `json:"type"`
	Priority int                    `json:"priority,omitempty"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
}

// Subscription Types

// DebateUpdate represents a debate update event.
type DebateUpdate struct {
	DebateID   string       `json:"debate_id"`
	Type       string       `json:"type"` // round_started, response_received, completed
	Round      *DebateRound `json:"round,omitempty"`
	Conclusion string       `json:"conclusion,omitempty"`
	Timestamp  time.Time    `json:"timestamp"`
}

// TaskProgress represents a task progress update.
type TaskProgress struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ProviderHealthUpdate represents a provider health update.
type ProviderHealthUpdate struct {
	ProviderID string       `json:"provider_id"`
	Status     HealthStatus `json:"status"`
	Timestamp  time.Time    `json:"timestamp"`
}

// TokenStreamEvent represents a token streaming event.
type TokenStreamEvent struct {
	RequestID   string    `json:"request_id"`
	Token       string    `json:"token"`
	IsComplete  bool      `json:"is_complete"`
	TokenCount  int       `json:"token_count"`
	Timestamp   time.Time `json:"timestamp"`
}
