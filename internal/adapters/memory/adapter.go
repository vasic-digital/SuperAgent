// Package memory provides adapters between HelixAgent's internal/memory types
// and the extracted digital.vasic.memory module.
package memory

import (
	"context"
	"errors"
	"time"

	helixmem "dev.helix.agent/internal/memory"
	modentity "digital.vasic.memory/pkg/entity"
	modmem "digital.vasic.memory/pkg/store"
)

// StoreAdapter adapts the module's MemoryStore interface to HelixAgent's MemoryStore.
type StoreAdapter struct {
	store modmem.MemoryStore
}

// NewStoreAdapter wraps a module store with the HelixAgent interface.
func NewStoreAdapter(store modmem.MemoryStore) *StoreAdapter {
	return &StoreAdapter{store: store}
}

// Add stores a new memory.
func (a *StoreAdapter) Add(ctx context.Context, memory *helixmem.Memory) error {
	modMemory := ToModuleMemory(memory)
	return a.store.Add(ctx, modMemory)
}

// Get retrieves a memory by ID.
func (a *StoreAdapter) Get(ctx context.Context, id string) (*helixmem.Memory, error) {
	modMemory, err := a.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToHelixMemory(modMemory), nil
}

// Update modifies an existing memory.
func (a *StoreAdapter) Update(ctx context.Context, memory *helixmem.Memory) error {
	modMemory := ToModuleMemory(memory)
	return a.store.Update(ctx, modMemory)
}

// Delete removes a memory by ID.
func (a *StoreAdapter) Delete(ctx context.Context, id string) error {
	return a.store.Delete(ctx, id)
}

// Search returns memories matching the query and options.
func (a *StoreAdapter) Search(ctx context.Context, query string, opts *helixmem.SearchOptions) ([]*helixmem.Memory, error) {
	modOpts := ToModuleSearchOptions(opts)
	modMemories, err := a.store.Search(ctx, query, modOpts)
	if err != nil {
		return nil, err
	}
	return ToHelixMemories(modMemories), nil
}

// GetByUser retrieves memories for a user.
func (a *StoreAdapter) GetByUser(ctx context.Context, userID string, opts *helixmem.ListOptions) ([]*helixmem.Memory, error) {
	// Module store uses Scope-based listing; filter by global for now
	modOpts := ToModuleListOptions(opts)
	modMemories, err := a.store.List(ctx, modmem.ScopeUser, modOpts)
	if err != nil {
		return nil, err
	}
	// Filter by user ID in metadata
	var result []*helixmem.Memory
	for _, m := range modMemories {
		if uid, ok := m.Metadata["user_id"].(string); ok && uid == userID {
			result = append(result, ToHelixMemory(m))
		}
	}
	return result, nil
}

// GetBySession retrieves memories for a session.
func (a *StoreAdapter) GetBySession(ctx context.Context, sessionID string) ([]*helixmem.Memory, error) {
	modOpts := modmem.DefaultListOptions()
	modMemories, err := a.store.List(ctx, modmem.ScopeSession, modOpts)
	if err != nil {
		return nil, err
	}
	var result []*helixmem.Memory
	for _, m := range modMemories {
		if sid, ok := m.Metadata["session_id"].(string); ok && sid == sessionID {
			result = append(result, ToHelixMemory(m))
		}
	}
	return result, nil
}

// Entity and Relationship methods
// Note: These methods require a graph backend (digital.vasic.memory/pkg/graph)
// which is not available in the current vendor. They return errors indicating
// that entity/relationship storage requires graph integration.

// AddEntity adds an entity to the knowledge graph.
// This feature requires graph backend integration (not available in vendor).
func (a *StoreAdapter) AddEntity(ctx context.Context, entity *helixmem.Entity) error {
	return errors.New("entity storage requires graph backend (digital.vasic.memory/pkg/graph) - not available in vendor")
}

// GetEntity retrieves an entity from the knowledge graph.
// This feature requires graph backend integration (not available in vendor).
func (a *StoreAdapter) GetEntity(ctx context.Context, id string) (*helixmem.Entity, error) {
	return nil, errors.New("entity retrieval requires graph backend (digital.vasic.memory/pkg/graph) - not available in vendor")
}

// SearchEntities searches for entities in the knowledge graph.
// This feature requires graph backend integration (not available in vendor).
func (a *StoreAdapter) SearchEntities(ctx context.Context, query string, limit int) ([]*helixmem.Entity, error) {
	return nil, errors.New("entity search requires graph backend (digital.vasic.memory/pkg/graph) - not available in vendor")
}

// AddRelationship adds a relationship between entities.
// This feature requires graph backend integration (not available in vendor).
func (a *StoreAdapter) AddRelationship(ctx context.Context, rel *helixmem.Relationship) error {
	return errors.New("relationship storage requires graph backend (digital.vasic.memory/pkg/graph) - not available in vendor")
}

// GetRelationships gets all relationships for an entity.
// This feature requires graph backend integration (not available in vendor).
func (a *StoreAdapter) GetRelationships(ctx context.Context, entityID string) ([]*helixmem.Relationship, error) {
	return nil, errors.New("relationship retrieval requires graph backend (digital.vasic.memory/pkg/graph) - not available in vendor")
}

// ToModuleMemory converts a HelixAgent Memory to a module Memory.
func ToModuleMemory(h *helixmem.Memory) *modmem.Memory {
	if h == nil {
		return nil
	}
	metadata := make(map[string]any)
	for k, v := range h.Metadata {
		metadata[k] = v
	}
	metadata["user_id"] = h.UserID
	metadata["session_id"] = h.SessionID
	metadata["type"] = string(h.Type)
	metadata["category"] = h.Category
	metadata["importance"] = h.Importance
	metadata["access_count"] = h.AccessCount

	return &modmem.Memory{
		ID:        h.ID,
		Content:   h.Content,
		Metadata:  metadata,
		Embedding: h.Embedding,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

// ToHelixMemory converts a module Memory to a HelixAgent Memory.
func ToHelixMemory(m *modmem.Memory) *helixmem.Memory {
	if m == nil {
		return nil
	}
	h := &helixmem.Memory{
		ID:        m.ID,
		Content:   m.Content,
		Metadata:  make(map[string]interface{}),
		Embedding: m.Embedding,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	// Extract HelixAgent-specific fields from metadata
	for k, v := range m.Metadata {
		switch k {
		case "user_id":
			h.UserID, _ = v.(string)
		case "session_id":
			h.SessionID, _ = v.(string)
		case "type":
			h.Type = helixmem.MemoryType(v.(string))
		case "category":
			h.Category, _ = v.(string)
		case "importance":
			if imp, ok := v.(float64); ok {
				h.Importance = imp
			}
		case "access_count":
			if count, ok := v.(int); ok {
				h.AccessCount = count
			}
		default:
			h.Metadata[k] = v
		}
	}

	return h
}

// ToHelixMemories converts a slice of module Memories to HelixAgent Memories.
func ToHelixMemories(modMemories []*modmem.Memory) []*helixmem.Memory {
	result := make([]*helixmem.Memory, len(modMemories))
	for i, m := range modMemories {
		result[i] = ToHelixMemory(m)
	}
	return result
}

// ToModuleSearchOptions converts HelixAgent SearchOptions to module SearchOptions.
func ToModuleSearchOptions(h *helixmem.SearchOptions) *modmem.SearchOptions {
	if h == nil {
		return modmem.DefaultSearchOptions()
	}
	opts := &modmem.SearchOptions{
		TopK:     h.TopK,
		MinScore: h.MinScore,
	}
	if h.TimeRange != nil {
		opts.TimeRange = &modmem.TimeRange{
			Start: h.TimeRange.Start,
			End:   h.TimeRange.End,
		}
	}
	return opts
}

// ToModuleListOptions converts HelixAgent ListOptions to module ListOptions.
func ToModuleListOptions(h *helixmem.ListOptions) *modmem.ListOptions {
	if h == nil {
		return modmem.DefaultListOptions()
	}
	return &modmem.ListOptions{
		Offset:  h.Offset,
		Limit:   h.Limit,
		OrderBy: h.SortBy,
	}
}

// EntityExtractorAdapter adapts the module's PatternExtractor to HelixAgent's interface.
type EntityExtractorAdapter struct {
	extractor *modentity.PatternExtractor
}

// NewEntityExtractorAdapter creates a new entity extractor adapter.
func NewEntityExtractorAdapter() *EntityExtractorAdapter {
	return &EntityExtractorAdapter{
		extractor: modentity.NewPatternExtractor(),
	}
}

// ExtractEntities extracts entities from text.
func (a *EntityExtractorAdapter) ExtractEntities(text string) ([]*helixmem.Entity, error) {
	entities, _, err := a.extractor.Extract(text)
	if err != nil {
		return nil, err
	}
	result := make([]*helixmem.Entity, len(entities))
	for i, e := range entities {
		result[i] = &helixmem.Entity{
			Name: e.Name,
			Type: e.Type,
			Properties: func() map[string]interface{} {
				props := make(map[string]interface{})
				for k, v := range e.Attributes {
					props[k] = v
				}
				return props
			}(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return result, nil
}

// ExtractRelations extracts relations from text.
func (a *EntityExtractorAdapter) ExtractRelations(text string) ([]*helixmem.Relationship, error) {
	_, relations, err := a.extractor.Extract(text)
	if err != nil {
		return nil, err
	}
	result := make([]*helixmem.Relationship, len(relations))
	for i, r := range relations {
		result[i] = &helixmem.Relationship{
			SourceID:  r.Subject,
			TargetID:  r.Object,
			Type:      r.Predicate,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return result, nil
}
