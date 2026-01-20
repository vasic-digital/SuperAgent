package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryStore provides an in-memory implementation of MemoryStore
type InMemoryStore struct {
	memories      map[string]*Memory
	entities      map[string]*Entity
	relationships map[string]*Relationship

	// Indexes for faster lookups
	userIndex    map[string][]string // userID -> memoryIDs
	sessionIndex map[string][]string // sessionID -> memoryIDs
	entityIndex  map[string][]string // entityID -> relationshipIDs

	mu sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories:      make(map[string]*Memory),
		entities:      make(map[string]*Entity),
		relationships: make(map[string]*Relationship),
		userIndex:     make(map[string][]string),
		sessionIndex:  make(map[string][]string),
		entityIndex:   make(map[string][]string),
	}
}

// Add adds a new memory
func (s *InMemoryStore) Add(ctx context.Context, memory *Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}

	s.memories[memory.ID] = memory

	// Update indexes
	if memory.UserID != "" {
		s.userIndex[memory.UserID] = append(s.userIndex[memory.UserID], memory.ID)
	}
	if memory.SessionID != "" {
		s.sessionIndex[memory.SessionID] = append(s.sessionIndex[memory.SessionID], memory.ID)
	}

	return nil
}

// Get retrieves a memory by ID
func (s *InMemoryStore) Get(ctx context.Context, id string) (*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memory, exists := s.memories[id]
	if !exists {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	// Update access stats
	memory.AccessCount++
	memory.LastAccess = time.Now()

	return memory, nil
}

// Update updates an existing memory
func (s *InMemoryStore) Update(ctx context.Context, memory *Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.memories[memory.ID]; !exists {
		return fmt.Errorf("memory not found: %s", memory.ID)
	}

	memory.UpdatedAt = time.Now()
	s.memories[memory.ID] = memory

	return nil
}

// Delete removes a memory
func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memory, exists := s.memories[id]
	if !exists {
		return fmt.Errorf("memory not found: %s", id)
	}

	// Remove from indexes
	if memory.UserID != "" {
		s.userIndex[memory.UserID] = removeFromSlice(s.userIndex[memory.UserID], id)
	}
	if memory.SessionID != "" {
		s.sessionIndex[memory.SessionID] = removeFromSlice(s.sessionIndex[memory.SessionID], id)
	}

	delete(s.memories, id)
	return nil
}

// Search searches for relevant memories
func (s *InMemoryStore) Search(ctx context.Context, query string, opts *SearchOptions) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts == nil {
		opts = DefaultSearchOptions()
	}

	var results []*Memory
	queryLower := strings.ToLower(query)

	for _, memory := range s.memories {
		// Filter by user
		if opts.UserID != "" && memory.UserID != opts.UserID {
			continue
		}

		// Filter by session
		if opts.SessionID != "" && memory.SessionID != opts.SessionID {
			continue
		}

		// Filter by type
		if opts.Type != "" && memory.Type != opts.Type {
			continue
		}

		// Filter by category
		if opts.Category != "" && memory.Category != opts.Category {
			continue
		}

		// Filter by time range
		if opts.TimeRange != nil {
			if memory.CreatedAt.Before(opts.TimeRange.Start) || memory.CreatedAt.After(opts.TimeRange.End) {
				continue
			}
		}

		// Simple text matching (in production, use vector similarity)
		score := s.calculateMatchScore(queryLower, memory)
		if score >= opts.MinScore {
			memoryCopy := *memory
			memoryCopy.Importance = score // Use importance as score for sorting
			results = append(results, &memoryCopy)
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Importance > results[j].Importance
	})

	// Limit results
	if opts.TopK > 0 && len(results) > opts.TopK {
		results = results[:opts.TopK]
	}

	return results, nil
}

// GetByUser retrieves memories for a user
func (s *InMemoryStore) GetByUser(ctx context.Context, userID string, opts *ListOptions) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memoryIDs := s.userIndex[userID]
	if len(memoryIDs) == 0 {
		return []*Memory{}, nil
	}

	var results []*Memory
	for _, id := range memoryIDs {
		if memory, exists := s.memories[id]; exists {
			results = append(results, memory)
		}
	}

	// Sort
	if opts != nil && opts.SortBy != "" {
		s.sortMemories(results, opts.SortBy, opts.Order)
	}

	// Pagination
	if opts != nil {
		start := opts.Offset
		if start > len(results) {
			return []*Memory{}, nil
		}

		end := start + opts.Limit
		if end > len(results) || opts.Limit == 0 {
			end = len(results)
		}

		results = results[start:end]
	}

	return results, nil
}

// GetBySession retrieves memories for a session
func (s *InMemoryStore) GetBySession(ctx context.Context, sessionID string) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memoryIDs := s.sessionIndex[sessionID]
	if len(memoryIDs) == 0 {
		return []*Memory{}, nil
	}

	var results []*Memory
	for _, id := range memoryIDs {
		if memory, exists := s.memories[id]; exists {
			results = append(results, memory)
		}
	}

	// Sort by creation time
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// AddEntity adds an entity
func (s *InMemoryStore) AddEntity(ctx context.Context, entity *Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}

	now := time.Now()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	s.entities[entity.ID] = entity
	return nil
}

// GetEntity retrieves an entity
func (s *InMemoryStore) GetEntity(ctx context.Context, id string) (*Entity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entity, exists := s.entities[id]
	if !exists {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	return entity, nil
}

// SearchEntities searches for entities
func (s *InMemoryStore) SearchEntities(ctx context.Context, query string, limit int) ([]*Entity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	queryLower := strings.ToLower(query)
	var results []*Entity

	for _, entity := range s.entities {
		if strings.Contains(strings.ToLower(entity.Name), queryLower) {
			results = append(results, entity)
		}
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// AddRelationship adds a relationship
func (s *InMemoryStore) AddRelationship(ctx context.Context, rel *Relationship) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rel.ID == "" {
		rel.ID = uuid.New().String()
	}

	now := time.Now()
	rel.CreatedAt = now
	rel.UpdatedAt = now

	s.relationships[rel.ID] = rel

	// Update indexes
	s.entityIndex[rel.SourceID] = append(s.entityIndex[rel.SourceID], rel.ID)
	s.entityIndex[rel.TargetID] = append(s.entityIndex[rel.TargetID], rel.ID)

	return nil
}

// GetRelationships gets relationships for an entity
func (s *InMemoryStore) GetRelationships(ctx context.Context, entityID string) ([]*Relationship, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	relIDs := s.entityIndex[entityID]
	var results []*Relationship

	for _, id := range relIDs {
		if rel, exists := s.relationships[id]; exists {
			results = append(results, rel)
		}
	}

	return results, nil
}

// Helper functions

func (s *InMemoryStore) calculateMatchScore(query string, memory *Memory) float64 {
	contentLower := strings.ToLower(memory.Content)

	// Simple word overlap score
	queryWords := strings.Fields(query)
	if len(queryWords) == 0 {
		return 0
	}

	matches := 0
	for _, word := range queryWords {
		if strings.Contains(contentLower, word) {
			matches++
		}
	}

	return float64(matches) / float64(len(queryWords))
}

func (s *InMemoryStore) sortMemories(memories []*Memory, sortBy, order string) {
	sort.Slice(memories, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "created_at":
			less = memories[i].CreatedAt.Before(memories[j].CreatedAt)
		case "updated_at":
			less = memories[i].UpdatedAt.Before(memories[j].UpdatedAt)
		case "importance":
			less = memories[i].Importance < memories[j].Importance
		case "access_count":
			less = memories[i].AccessCount < memories[j].AccessCount
		default:
			less = memories[i].CreatedAt.Before(memories[j].CreatedAt)
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
