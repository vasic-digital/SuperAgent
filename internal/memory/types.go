// Package memory provides a Mem0-style memory system for persistent AI memory
// with fact extraction, graph relationships, and cross-session recall.
package memory

import (
	"context"
	"time"
)

// Memory represents a stored memory unit
type Memory struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	Content     string                 `json:"content"`
	Summary     string                 `json:"summary,omitempty"`
	Type        MemoryType             `json:"type"`
	Category    string                 `json:"category,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Embedding   []float32              `json:"embedding,omitempty"`
	Importance  float64                `json:"importance"`
	AccessCount int                    `json:"access_count"`
	LastAccess  time.Time              `json:"last_access"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// MemoryType categorizes different types of memories
type MemoryType string

const (
	MemoryTypeEpisodic   MemoryType = "episodic"   // Conversation/event memories
	MemoryTypeSemantic   MemoryType = "semantic"   // Facts and knowledge
	MemoryTypeProcedural MemoryType = "procedural" // How-to knowledge
	MemoryTypeWorking    MemoryType = "working"    // Short-term context
)

// Entity represents an extracted entity from memories
type Entity struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // person, place, thing, concept
	Properties map[string]interface{} `json:"properties,omitempty"`
	Aliases    []string               `json:"aliases,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Relationship represents a connection between entities
type Relationship struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"source_id"`
	TargetID   string                 `json:"target_id"`
	Type       string                 `json:"type"` // knows, works_at, located_in, etc.
	Properties map[string]interface{} `json:"properties,omitempty"`
	Strength   float64                `json:"strength"` // 0-1 confidence
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Message represents a conversation message
type Message struct {
	Role      string                 `json:"role"` // user, assistant, system
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryStore defines the interface for memory storage
type MemoryStore interface {
	// Memory CRUD
	Add(ctx context.Context, memory *Memory) error
	Get(ctx context.Context, id string) (*Memory, error)
	Update(ctx context.Context, memory *Memory) error
	Delete(ctx context.Context, id string) error

	// Search
	Search(ctx context.Context, query string, opts *SearchOptions) ([]*Memory, error)
	GetByUser(ctx context.Context, userID string, opts *ListOptions) ([]*Memory, error)
	GetBySession(ctx context.Context, sessionID string) ([]*Memory, error)

	// Entity operations
	AddEntity(ctx context.Context, entity *Entity) error
	GetEntity(ctx context.Context, id string) (*Entity, error)
	SearchEntities(ctx context.Context, query string, limit int) ([]*Entity, error)

	// Relationship operations
	AddRelationship(ctx context.Context, rel *Relationship) error
	GetRelationships(ctx context.Context, entityID string) ([]*Relationship, error)
}

// SearchOptions configures memory search
type SearchOptions struct {
	UserID       string     `json:"user_id,omitempty"`
	SessionID    string     `json:"session_id,omitempty"`
	Type         MemoryType `json:"type,omitempty"`
	Category     string     `json:"category,omitempty"`
	TopK         int        `json:"top_k"`
	MinScore     float64    `json:"min_score"`
	IncludeGraph bool       `json:"include_graph"`
	TimeRange    *TimeRange `json:"time_range,omitempty"`
}

// ListOptions for listing memories
type ListOptions struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	SortBy string `json:"sort_by"`
	Order  string `json:"order"` // asc, desc
}

// TimeRange for filtering by time
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		TopK:         10,
		MinScore:     0.5,
		IncludeGraph: false,
	}
}

// MemoryExtractor extracts memories from conversations
type MemoryExtractor interface {
	// Extract extracts memories from messages
	Extract(ctx context.Context, messages []Message, userID string) ([]*Memory, error)
	// ExtractEntities extracts entities from text
	ExtractEntities(ctx context.Context, text string) ([]*Entity, error)
	// ExtractRelationships extracts relationships from text and entities
	ExtractRelationships(ctx context.Context, text string, entities []*Entity) ([]*Relationship, error)
}

// MemorySummarizer summarizes conversation history
type MemorySummarizer interface {
	// Summarize creates a summary of messages
	Summarize(ctx context.Context, messages []Message) (string, error)
	// SummarizeProgressive progressively summarizes long history
	SummarizeProgressive(ctx context.Context, messages []Message, existingSummary string) (string, error)
}

// MemoryConfig configures the memory manager
type MemoryConfig struct {
	// Storage backend
	StorageType        string `json:"storage_type"` // memory, postgres, redis, qdrant
	VectorDBEndpoint   string `json:"vectordb_endpoint"`
	VectorDBAPIKey     string `json:"vectordb_api_key"`
	VectorDBCollection string `json:"vectordb_collection"`

	// Embedding configuration
	EmbeddingModel     string `json:"embedding_model"`
	EmbeddingEndpoint  string `json:"embedding_endpoint"`
	EmbeddingDimension int    `json:"embedding_dimension"`

	// Memory settings
	MaxMemoriesPerUser int           `json:"max_memories_per_user"`
	MemoryTTL          time.Duration `json:"memory_ttl"`
	EnableGraph        bool          `json:"enable_graph"`
	EnableCompression  bool          `json:"enable_compression"`

	// LLM for extraction
	ExtractorModel    string `json:"extractor_model"`
	ExtractorEndpoint string `json:"extractor_endpoint"`
	ExtractorAPIKey   string `json:"extractor_api_key"`
}

// DefaultMemoryConfig returns default configuration
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		StorageType:        "memory",
		VectorDBCollection: "helixagent_memories",
		EmbeddingDimension: 1536,
		MaxMemoriesPerUser: 10000,
		MemoryTTL:          0, // No expiration by default
		EnableGraph:        true,
		EnableCompression:  true,
	}
}
