package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Manager provides memory management with Mem0-style capabilities
type Manager struct {
	store      MemoryStore
	extractor  MemoryExtractor
	summarizer MemorySummarizer
	embedder   Embedder
	config     *MemoryConfig
	logger     *logrus.Logger
	mu         sync.RWMutex
}

// Embedder generates embeddings for memories
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, query string) ([]float32, error)
}

// NewManager creates a new memory manager
func NewManager(
	store MemoryStore,
	extractor MemoryExtractor,
	summarizer MemorySummarizer,
	embedder Embedder,
	config *MemoryConfig,
	logger *logrus.Logger,
) *Manager {
	if config == nil {
		config = DefaultMemoryConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &Manager{
		store:      store,
		extractor:  extractor,
		summarizer: summarizer,
		embedder:   embedder,
		config:     config,
		logger:     logger,
	}
}

// AddMemory adds a new memory
func (m *Manager) AddMemory(ctx context.Context, memory *Memory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}

	now := time.Now()
	memory.CreatedAt = now
	memory.UpdatedAt = now
	memory.LastAccess = now

	// Generate embedding if not provided
	if len(memory.Embedding) == 0 && m.embedder != nil {
		embeddings, err := m.embedder.Embed(ctx, []string{memory.Content})
		if err != nil {
			m.logger.WithError(err).Warn("Failed to generate embedding")
		} else if len(embeddings) > 0 {
			memory.Embedding = embeddings[0]
		}
	}

	// Calculate importance if not set
	if memory.Importance == 0 {
		memory.Importance = m.calculateImportance(memory)
	}

	return m.store.Add(ctx, memory)
}

// AddFromMessages extracts and adds memories from conversation messages
func (m *Manager) AddFromMessages(ctx context.Context, messages []Message, userID, sessionID string) ([]*Memory, error) {
	if m.extractor == nil {
		return nil, fmt.Errorf("no extractor configured")
	}

	// Extract memories
	memories, err := m.extractor.Extract(ctx, messages, userID)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	// Add session ID to all memories
	for _, mem := range memories {
		mem.SessionID = sessionID
	}

	// Extract entities and relationships if graph is enabled
	if m.config.EnableGraph {
		for _, mem := range memories {
			entities, err := m.extractor.ExtractEntities(ctx, mem.Content)
			if err != nil {
				m.logger.WithError(err).Debug("Entity extraction failed")
				continue
			}

			for _, entity := range entities {
				if err := m.store.AddEntity(ctx, entity); err != nil {
					m.logger.WithError(err).Debug("Failed to add entity")
				}
			}

			relationships, err := m.extractor.ExtractRelationships(ctx, mem.Content, entities)
			if err != nil {
				m.logger.WithError(err).Debug("Relationship extraction failed")
				continue
			}

			for _, rel := range relationships {
				if err := m.store.AddRelationship(ctx, rel); err != nil {
					m.logger.WithError(err).Debug("Failed to add relationship")
				}
			}
		}
	}

	// Store memories
	for _, mem := range memories {
		if err := m.AddMemory(ctx, mem); err != nil {
			m.logger.WithError(err).Warn("Failed to add memory")
		}
	}

	m.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"session_id":  sessionID,
		"memory_count": len(memories),
	}).Debug("Memories added from messages")

	return memories, nil
}

// Search searches for relevant memories
func (m *Manager) Search(ctx context.Context, query string, opts *SearchOptions) ([]*Memory, error) {
	if opts == nil {
		opts = DefaultSearchOptions()
	}

	// Generate query embedding
	if m.embedder != nil {
		// Let the store handle the embedding-based search
	}

	return m.store.Search(ctx, query, opts)
}

// GetContext retrieves relevant context for a query
func (m *Manager) GetContext(ctx context.Context, query string, userID string, maxTokens int) (string, error) {
	opts := &SearchOptions{
		UserID:       userID,
		TopK:         20,
		MinScore:     0.5,
		IncludeGraph: m.config.EnableGraph,
	}

	memories, err := m.Search(ctx, query, opts)
	if err != nil {
		return "", err
	}

	// Build context string
	var context string
	totalChars := 0
	approxCharsPerToken := 4

	for _, mem := range memories {
		if totalChars+len(mem.Content) > maxTokens*approxCharsPerToken {
			break
		}
		context += fmt.Sprintf("- %s\n", mem.Content)
		totalChars += len(mem.Content) + 3
	}

	return context, nil
}

// GetUserMemories retrieves all memories for a user
func (m *Manager) GetUserMemories(ctx context.Context, userID string, opts *ListOptions) ([]*Memory, error) {
	return m.store.GetByUser(ctx, userID, opts)
}

// GetSessionMemories retrieves memories from a specific session
func (m *Manager) GetSessionMemories(ctx context.Context, sessionID string) ([]*Memory, error) {
	return m.store.GetBySession(ctx, sessionID)
}

// DeleteMemory deletes a memory by ID
func (m *Manager) DeleteMemory(ctx context.Context, id string) error {
	return m.store.Delete(ctx, id)
}

// DeleteUserMemories deletes all memories for a user
func (m *Manager) DeleteUserMemories(ctx context.Context, userID string) error {
	memories, err := m.store.GetByUser(ctx, userID, &ListOptions{Limit: 10000})
	if err != nil {
		return err
	}

	for _, mem := range memories {
		if err := m.store.Delete(ctx, mem.ID); err != nil {
			m.logger.WithError(err).Warn("Failed to delete memory")
		}
	}

	return nil
}

// SummarizeHistory summarizes conversation history
func (m *Manager) SummarizeHistory(ctx context.Context, messages []Message) (string, error) {
	if m.summarizer == nil {
		return "", fmt.Errorf("no summarizer configured")
	}

	return m.summarizer.Summarize(ctx, messages)
}

// SummarizeProgressively progressively summarizes long history
func (m *Manager) SummarizeProgressively(ctx context.Context, messages []Message, existingSummary string) (string, error) {
	if m.summarizer == nil {
		return "", fmt.Errorf("no summarizer configured")
	}

	return m.summarizer.SummarizeProgressive(ctx, messages, existingSummary)
}

// GetRelatedEntities finds entities related to a query
func (m *Manager) GetRelatedEntities(ctx context.Context, query string, limit int) ([]*Entity, error) {
	return m.store.SearchEntities(ctx, query, limit)
}

// GetEntityRelationships gets relationships for an entity
func (m *Manager) GetEntityRelationships(ctx context.Context, entityID string) ([]*Relationship, error) {
	return m.store.GetRelationships(ctx, entityID)
}

// calculateImportance calculates memory importance score
func (m *Manager) calculateImportance(memory *Memory) float64 {
	importance := 0.5 // Base importance

	// Boost for semantic memories (facts)
	if memory.Type == MemoryTypeSemantic {
		importance += 0.2
	}

	// Boost for longer, more detailed content
	if len(memory.Content) > 100 {
		importance += 0.1
	}

	// Boost if it has entities
	if memory.Metadata != nil {
		if entities, ok := memory.Metadata["entities"].([]interface{}); ok && len(entities) > 0 {
			importance += 0.1
		}
	}

	// Cap at 1.0
	if importance > 1.0 {
		importance = 1.0
	}

	return importance
}

// Cleanup removes expired memories
func (m *Manager) Cleanup(ctx context.Context) error {
	// This would be called periodically to remove expired memories
	// Implementation depends on the store's capabilities
	m.logger.Debug("Memory cleanup started")
	return nil
}

// Stats returns memory statistics
type Stats struct {
	TotalMemories  int            `json:"total_memories"`
	MemoriesByType map[string]int `json:"memories_by_type"`
	TotalEntities  int            `json:"total_entities"`
	TotalRelations int            `json:"total_relations"`
	OldestMemory   *time.Time     `json:"oldest_memory,omitempty"`
	NewestMemory   *time.Time     `json:"newest_memory,omitempty"`
}

// GetStats returns memory statistics for a user
func (m *Manager) GetStats(ctx context.Context, userID string) (*Stats, error) {
	memories, err := m.store.GetByUser(ctx, userID, &ListOptions{Limit: 100000})
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		TotalMemories:  len(memories),
		MemoriesByType: make(map[string]int),
	}

	for _, mem := range memories {
		stats.MemoriesByType[string(mem.Type)]++

		if stats.OldestMemory == nil || mem.CreatedAt.Before(*stats.OldestMemory) {
			stats.OldestMemory = &mem.CreatedAt
		}
		if stats.NewestMemory == nil || mem.CreatedAt.After(*stats.NewestMemory) {
			stats.NewestMemory = &mem.CreatedAt
		}
	}

	return stats, nil
}
