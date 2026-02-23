// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MemoryAdapterConfig holds configuration for Memory MCP adapter
type MemoryAdapterConfig struct {
	// StoragePath is where memory data is persisted
	StoragePath string `json:"storage_path,omitempty"`
	// MaxEntities is the maximum number of entities to store
	MaxEntities int `json:"max_entities,omitempty"`
	// MaxRelations is the maximum number of relations to store
	MaxRelations int `json:"max_relations,omitempty"`
	// EnablePersistence enables persistent storage
	EnablePersistence bool `json:"enable_persistence"`
	// AutoSaveInterval is how often to auto-save (0 disables)
	AutoSaveInterval time.Duration `json:"auto_save_interval,omitempty"`
}

// DefaultMemoryAdapterConfig returns default configuration
func DefaultMemoryAdapterConfig() MemoryAdapterConfig {
	homeDir, _ := os.UserHomeDir()
	return MemoryAdapterConfig{
		StoragePath:       filepath.Join(homeDir, ".helix", "memory"),
		MaxEntities:       10000,
		MaxRelations:      50000,
		EnablePersistence: true,
		AutoSaveInterval:  5 * time.Minute,
	}
}

// Entity represents an entity in the knowledge graph
type Entity struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	EntityType   string                 `json:"entity_type"`
	Observations []string               `json:"observations,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Relation represents a relation between entities
type Relation struct {
	ID           string                 `json:"id"`
	FromEntity   string                 `json:"from_entity"`
	ToEntity     string                 `json:"to_entity"`
	RelationType string                 `json:"relation_type"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Strength     float32                `json:"strength"`
	CreatedAt    time.Time              `json:"created_at"`
}

// KnowledgeGraph represents the in-memory knowledge graph
type KnowledgeGraph struct {
	Entities  map[string]*Entity   `json:"entities"`
	Relations map[string]*Relation `json:"relations"`
	Version   int                  `json:"version"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// MemoryAdapter implements MCP adapter for knowledge graph memory
type MemoryAdapter struct {
	config      MemoryAdapterConfig
	graph       *KnowledgeGraph
	initialized bool
	dirty       bool // Indicates unsaved changes
	mu          sync.RWMutex
	logger      *logrus.Logger
	stopChan    chan struct{}
}

// NewMemoryAdapter creates a new Memory MCP adapter
func NewMemoryAdapter(config MemoryAdapterConfig, logger *logrus.Logger) *MemoryAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.MaxEntities <= 0 {
		config.MaxEntities = 10000
	}
	if config.MaxRelations <= 0 {
		config.MaxRelations = 50000
	}

	return &MemoryAdapter{
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Initialize initializes the Memory adapter
func (m *MemoryAdapter) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize empty graph
	m.graph = &KnowledgeGraph{
		Entities:  make(map[string]*Entity),
		Relations: make(map[string]*Relation),
		Version:   1,
		UpdatedAt: time.Now(),
	}

	// Load from storage if persistence is enabled
	if m.config.EnablePersistence {
		// #nosec G301 -- memory storage directory uses standard 0750 permissions
		if err := os.MkdirAll(m.config.StoragePath, 0750); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}

		if err := m.loadFromDisk(); err != nil {
			m.logger.WithError(err).Warn("Failed to load existing memory, starting fresh")
		}
	}

	m.initialized = true

	// Start auto-save if configured
	if m.config.EnablePersistence && m.config.AutoSaveInterval > 0 {
		go m.autoSaveLoop()
	}

	m.logger.Info("Memory adapter initialized")
	return nil
}

// loadFromDisk loads the knowledge graph from disk
func (m *MemoryAdapter) loadFromDisk() error {
	graphFile := filepath.Join(m.config.StoragePath, "knowledge_graph.json")

	data, err := os.ReadFile(graphFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing data
		}
		return err
	}

	var graph KnowledgeGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return fmt.Errorf("failed to unmarshal graph: %w", err)
	}

	m.graph = &graph
	return nil
}

// saveToDisk saves the knowledge graph to disk
func (m *MemoryAdapter) saveToDisk() error {
	if !m.config.EnablePersistence {
		return nil
	}

	graphFile := filepath.Join(m.config.StoragePath, "knowledge_graph.json")

	data, err := json.MarshalIndent(m.graph, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	// Use 0600 permissions for user's private knowledge graph data
	tempFile := graphFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempFile, graphFile); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	m.dirty = false
	return nil
}

// autoSaveLoop periodically saves the graph
func (m *MemoryAdapter) autoSaveLoop() {
	ticker := time.NewTicker(m.config.AutoSaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.mu.Lock()
			if m.dirty {
				if err := m.saveToDisk(); err != nil {
					m.logger.WithError(err).Error("Auto-save failed")
				}
			}
			m.mu.Unlock()
		}
	}
}

// Health returns health status
func (m *MemoryAdapter) Health(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return fmt.Errorf("memory adapter not initialized")
	}

	return nil
}

// Close closes the adapter and saves data
func (m *MemoryAdapter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop auto-save
	close(m.stopChan)

	// Save before closing
	if m.dirty && m.config.EnablePersistence {
		if err := m.saveToDisk(); err != nil {
			m.logger.WithError(err).Error("Failed to save on close")
		}
	}

	m.initialized = false
	return nil
}

// CreateEntity creates a new entity
func (m *MemoryAdapter) CreateEntity(ctx context.Context, name, entityType string, observations []string, properties map[string]interface{}) (*Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if len(m.graph.Entities) >= m.config.MaxEntities {
		return nil, fmt.Errorf("maximum entity limit reached (%d)", m.config.MaxEntities)
	}

	entity := &Entity{
		ID:           uuid.New().String(),
		Name:         name,
		EntityType:   entityType,
		Observations: observations,
		Properties:   properties,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	m.graph.Entities[entity.ID] = entity
	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return entity, nil
}

// GetEntity retrieves an entity by ID
func (m *MemoryAdapter) GetEntity(ctx context.Context, id string) (*Entity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	entity, ok := m.graph.Entities[id]
	if !ok {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	return entity, nil
}

// GetEntityByName retrieves an entity by name
func (m *MemoryAdapter) GetEntityByName(ctx context.Context, name string) (*Entity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	for _, entity := range m.graph.Entities {
		if strings.EqualFold(entity.Name, name) {
			return entity, nil
		}
	}

	return nil, fmt.Errorf("entity not found: %s", name)
}

// UpdateEntity updates an existing entity
func (m *MemoryAdapter) UpdateEntity(ctx context.Context, id string, observations []string, properties map[string]interface{}) (*Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	entity, ok := m.graph.Entities[id]
	if !ok {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	if observations != nil {
		entity.Observations = append(entity.Observations, observations...)
	}
	if properties != nil {
		if entity.Properties == nil {
			entity.Properties = make(map[string]interface{})
		}
		for k, v := range properties {
			entity.Properties[k] = v
		}
	}
	entity.UpdatedAt = time.Now()

	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return entity, nil
}

// DeleteEntity deletes an entity and its relations
func (m *MemoryAdapter) DeleteEntity(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if _, ok := m.graph.Entities[id]; !ok {
		return fmt.Errorf("entity not found: %s", id)
	}

	// Delete related relations
	for relID, rel := range m.graph.Relations {
		if rel.FromEntity == id || rel.ToEntity == id {
			delete(m.graph.Relations, relID)
		}
	}

	delete(m.graph.Entities, id)
	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return nil
}

// SearchEntities searches for entities by name or type
func (m *MemoryAdapter) SearchEntities(ctx context.Context, query string, entityType string, limit int) ([]*Entity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if limit <= 0 {
		limit = 100
	}

	var results []*Entity
	queryLower := strings.ToLower(query)

	for _, entity := range m.graph.Entities {
		if entityType != "" && entity.EntityType != entityType {
			continue
		}

		if query != "" {
			nameLower := strings.ToLower(entity.Name)
			if !strings.Contains(nameLower, queryLower) {
				// Check observations
				found := false
				for _, obs := range entity.Observations {
					if strings.Contains(strings.ToLower(obs), queryLower) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		results = append(results, entity)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// CreateRelation creates a relation between two entities
func (m *MemoryAdapter) CreateRelation(ctx context.Context, fromEntity, toEntity, relationType string, strength float32, properties map[string]interface{}) (*Relation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if len(m.graph.Relations) >= m.config.MaxRelations {
		return nil, fmt.Errorf("maximum relation limit reached (%d)", m.config.MaxRelations)
	}

	// Verify both entities exist
	if _, ok := m.graph.Entities[fromEntity]; !ok {
		return nil, fmt.Errorf("from entity not found: %s", fromEntity)
	}
	if _, ok := m.graph.Entities[toEntity]; !ok {
		return nil, fmt.Errorf("to entity not found: %s", toEntity)
	}

	if strength <= 0 {
		strength = 1.0
	}

	relation := &Relation{
		ID:           uuid.New().String(),
		FromEntity:   fromEntity,
		ToEntity:     toEntity,
		RelationType: relationType,
		Properties:   properties,
		Strength:     strength,
		CreatedAt:    time.Now(),
	}

	m.graph.Relations[relation.ID] = relation
	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return relation, nil
}

// GetRelation retrieves a relation by ID
func (m *MemoryAdapter) GetRelation(ctx context.Context, id string) (*Relation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	relation, ok := m.graph.Relations[id]
	if !ok {
		return nil, fmt.Errorf("relation not found: %s", id)
	}

	return relation, nil
}

// GetEntityRelations retrieves all relations for an entity
func (m *MemoryAdapter) GetEntityRelations(ctx context.Context, entityID string, direction string) ([]*Relation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	var results []*Relation

	for _, rel := range m.graph.Relations {
		switch direction {
		case "outgoing":
			if rel.FromEntity == entityID {
				results = append(results, rel)
			}
		case "incoming":
			if rel.ToEntity == entityID {
				results = append(results, rel)
			}
		default:
			if rel.FromEntity == entityID || rel.ToEntity == entityID {
				results = append(results, rel)
			}
		}
	}

	return results, nil
}

// DeleteRelation deletes a relation
func (m *MemoryAdapter) DeleteRelation(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if _, ok := m.graph.Relations[id]; !ok {
		return fmt.Errorf("relation not found: %s", id)
	}

	delete(m.graph.Relations, id)
	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return nil
}

// GraphStatistics contains statistics about the knowledge graph
type GraphStatistics struct {
	TotalEntities  int            `json:"total_entities"`
	TotalRelations int            `json:"total_relations"`
	EntityTypes    map[string]int `json:"entity_types"`
	RelationTypes  map[string]int `json:"relation_types"`
	Version        int            `json:"version"`
	LastUpdated    time.Time      `json:"last_updated"`
	StorageSize    int64          `json:"storage_size_bytes,omitempty"`
}

// GetStatistics returns statistics about the knowledge graph
func (m *MemoryAdapter) GetStatistics(ctx context.Context) (*GraphStatistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	stats := &GraphStatistics{
		TotalEntities:  len(m.graph.Entities),
		TotalRelations: len(m.graph.Relations),
		EntityTypes:    make(map[string]int),
		RelationTypes:  make(map[string]int),
		Version:        m.graph.Version,
		LastUpdated:    m.graph.UpdatedAt,
	}

	for _, entity := range m.graph.Entities {
		stats.EntityTypes[entity.EntityType]++
	}

	for _, relation := range m.graph.Relations {
		stats.RelationTypes[relation.RelationType]++
	}

	// Get storage size if persistence is enabled
	if m.config.EnablePersistence {
		graphFile := filepath.Join(m.config.StoragePath, "knowledge_graph.json")
		if info, err := os.Stat(graphFile); err == nil {
			stats.StorageSize = info.Size()
		}
	}

	return stats, nil
}

// AddObservation adds an observation to an entity
func (m *MemoryAdapter) AddObservation(ctx context.Context, entityID, observation string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	entity, ok := m.graph.Entities[entityID]
	if !ok {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	entity.Observations = append(entity.Observations, observation)
	entity.UpdatedAt = time.Now()

	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return nil
}

// ReadGraph reads specific nodes from the graph
func (m *MemoryAdapter) ReadGraph(ctx context.Context, entityNames []string) (map[string]*EntityWithRelations, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	result := make(map[string]*EntityWithRelations)

	// Find entities by name
	for _, name := range entityNames {
		for _, entity := range m.graph.Entities {
			if strings.EqualFold(entity.Name, name) {
				ewr := &EntityWithRelations{
					Entity:    entity,
					Relations: []*Relation{},
				}

				// Get relations
				for _, rel := range m.graph.Relations {
					if rel.FromEntity == entity.ID || rel.ToEntity == entity.ID {
						ewr.Relations = append(ewr.Relations, rel)
					}
				}

				result[name] = ewr
				break
			}
		}
	}

	return result, nil
}

// EntityWithRelations contains an entity and its relations
type EntityWithRelations struct {
	Entity    *Entity     `json:"entity"`
	Relations []*Relation `json:"relations"`
}

// OpenNodes retrieves entities and their immediate connections
func (m *MemoryAdapter) OpenNodes(ctx context.Context, entityIDs []string) ([]*EntityWithRelations, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	var results []*EntityWithRelations

	for _, id := range entityIDs {
		entity, ok := m.graph.Entities[id]
		if !ok {
			continue
		}

		ewr := &EntityWithRelations{
			Entity:    entity,
			Relations: []*Relation{},
		}

		// Get relations and connected entities
		for _, rel := range m.graph.Relations {
			if rel.FromEntity == id || rel.ToEntity == id {
				ewr.Relations = append(ewr.Relations, rel)
			}
		}

		results = append(results, ewr)
	}

	return results, nil
}

// Save forces a save to disk
func (m *MemoryAdapter) Save(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	return m.saveToDisk()
}

// Clear clears all data from the graph
func (m *MemoryAdapter) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	m.graph.Entities = make(map[string]*Entity)
	m.graph.Relations = make(map[string]*Relation)
	m.graph.Version++
	m.graph.UpdatedAt = time.Now()
	m.dirty = true

	return nil
}

// SearchResult represents a search result with relevance score
type MemorySearchResult struct {
	Entity    *Entity  `json:"entity"`
	Score     float32  `json:"score"`
	MatchedIn []string `json:"matched_in"`
}

// SearchWithRelevance searches with relevance scoring
func (m *MemoryAdapter) SearchWithRelevance(ctx context.Context, query string, limit int) ([]*MemorySearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if limit <= 0 {
		limit = 20
	}

	var results []*MemorySearchResult
	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	for _, entity := range m.graph.Entities {
		var score float32
		var matchedIn []string

		nameLower := strings.ToLower(entity.Name)

		// Exact name match gets highest score
		if nameLower == queryLower {
			score += 10.0
			matchedIn = append(matchedIn, "name (exact)")
		} else if strings.Contains(nameLower, queryLower) {
			score += 5.0
			matchedIn = append(matchedIn, "name (partial)")
		}

		// Term matching
		for _, term := range queryTerms {
			if strings.Contains(nameLower, term) {
				score += 1.0
			}
		}

		// Check observations
		for _, obs := range entity.Observations {
			obsLower := strings.ToLower(obs)
			if strings.Contains(obsLower, queryLower) {
				score += 3.0
				matchedIn = append(matchedIn, "observation")
			}
			for _, term := range queryTerms {
				if strings.Contains(obsLower, term) {
					score += 0.5
				}
			}
		}

		// Check type
		if strings.Contains(strings.ToLower(entity.EntityType), queryLower) {
			score += 2.0
			matchedIn = append(matchedIn, "type")
		}

		if score > 0 {
			results = append(results, &MemorySearchResult{
				Entity:    entity,
				Score:     score,
				MatchedIn: matchedIn,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (m *MemoryAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "memory_create_entity",
			Description: "Create a new entity in the knowledge graph",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Entity name",
					},
					"entity_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of entity (e.g., person, concept, project)",
					},
					"observations": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Initial observations about the entity",
					},
					"properties": map[string]interface{}{
						"type":        "object",
						"description": "Additional properties",
					},
				},
				"required": []string{"name", "entity_type"},
			},
		},
		{
			Name:        "memory_get_entity",
			Description: "Get an entity by ID or name",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Entity ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Entity name",
					},
				},
			},
		},
		{
			Name:        "memory_update_entity",
			Description: "Update an existing entity",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Entity ID",
					},
					"observations": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "New observations to add",
					},
					"properties": map[string]interface{}{
						"type":        "object",
						"description": "Properties to update",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "memory_delete_entity",
			Description: "Delete an entity and its relations",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Entity ID",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "memory_search",
			Description: "Search for entities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"entity_type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by entity type",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum results",
						"default":     20,
					},
				},
			},
		},
		{
			Name:        "memory_create_relation",
			Description: "Create a relation between two entities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"from_entity": map[string]interface{}{
						"type":        "string",
						"description": "Source entity ID",
					},
					"to_entity": map[string]interface{}{
						"type":        "string",
						"description": "Target entity ID",
					},
					"relation_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of relation (e.g., related_to, part_of, knows)",
					},
					"strength": map[string]interface{}{
						"type":        "number",
						"description": "Relation strength (0-1)",
						"default":     1.0,
					},
					"properties": map[string]interface{}{
						"type":        "object",
						"description": "Additional relation properties",
					},
				},
				"required": []string{"from_entity", "to_entity", "relation_type"},
			},
		},
		{
			Name:        "memory_get_relations",
			Description: "Get relations for an entity",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"entity_id": map[string]interface{}{
						"type":        "string",
						"description": "Entity ID",
					},
					"direction": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"incoming", "outgoing", "all"},
						"description": "Direction filter",
						"default":     "all",
					},
				},
				"required": []string{"entity_id"},
			},
		},
		{
			Name:        "memory_delete_relation",
			Description: "Delete a relation",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Relation ID",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "memory_read_graph",
			Description: "Read specific nodes from the knowledge graph by name",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"entity_names": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Entity names to read",
					},
				},
				"required": []string{"entity_names"},
			},
		},
		{
			Name:        "memory_add_observation",
			Description: "Add an observation to an entity",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"entity_id": map[string]interface{}{
						"type":        "string",
						"description": "Entity ID",
					},
					"observation": map[string]interface{}{
						"type":        "string",
						"description": "Observation to add",
					},
				},
				"required": []string{"entity_id", "observation"},
			},
		},
		{
			Name:        "memory_statistics",
			Description: "Get knowledge graph statistics",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "memory_save",
			Description: "Force save the knowledge graph to disk",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "memory_clear",
			Description: "Clear all data from the knowledge graph",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (m *MemoryAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	initialized := m.initialized
	m.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	switch toolName {
	case "memory_create_entity":
		name, _ := params["name"].(string)
		entityType, _ := params["entity_type"].(string)
		var observations []string
		if obs, ok := params["observations"].([]interface{}); ok {
			for _, o := range obs {
				if s, ok := o.(string); ok {
					observations = append(observations, s)
				}
			}
		}
		properties, _ := params["properties"].(map[string]interface{})
		return m.CreateEntity(ctx, name, entityType, observations, properties)

	case "memory_get_entity":
		if id, ok := params["id"].(string); ok && id != "" {
			return m.GetEntity(ctx, id)
		}
		if name, ok := params["name"].(string); ok && name != "" {
			return m.GetEntityByName(ctx, name)
		}
		return nil, fmt.Errorf("either id or name required")

	case "memory_update_entity":
		id, _ := params["id"].(string)
		var observations []string
		if obs, ok := params["observations"].([]interface{}); ok {
			for _, o := range obs {
				if s, ok := o.(string); ok {
					observations = append(observations, s)
				}
			}
		}
		properties, _ := params["properties"].(map[string]interface{})
		return m.UpdateEntity(ctx, id, observations, properties)

	case "memory_delete_entity":
		id, _ := params["id"].(string)
		return map[string]interface{}{"success": true}, m.DeleteEntity(ctx, id)

	case "memory_search":
		query, _ := params["query"].(string)
		entityType, _ := params["entity_type"].(string)
		limit := 20
		if l, ok := params["limit"].(float64); ok {
			limit = int(l)
		}
		return m.SearchEntities(ctx, query, entityType, limit)

	case "memory_create_relation":
		fromEntity, _ := params["from_entity"].(string)
		toEntity, _ := params["to_entity"].(string)
		relationType, _ := params["relation_type"].(string)
		strength := float32(1.0)
		if s, ok := params["strength"].(float64); ok {
			strength = float32(s)
		}
		properties, _ := params["properties"].(map[string]interface{})
		return m.CreateRelation(ctx, fromEntity, toEntity, relationType, strength, properties)

	case "memory_get_relations":
		entityID, _ := params["entity_id"].(string)
		direction, _ := params["direction"].(string)
		if direction == "" {
			direction = "all"
		}
		return m.GetEntityRelations(ctx, entityID, direction)

	case "memory_delete_relation":
		id, _ := params["id"].(string)
		return map[string]interface{}{"success": true}, m.DeleteRelation(ctx, id)

	case "memory_read_graph":
		var names []string
		if n, ok := params["entity_names"].([]interface{}); ok {
			for _, name := range n {
				if s, ok := name.(string); ok {
					names = append(names, s)
				}
			}
		}
		return m.ReadGraph(ctx, names)

	case "memory_add_observation":
		entityID, _ := params["entity_id"].(string)
		observation, _ := params["observation"].(string)
		return map[string]interface{}{"success": true}, m.AddObservation(ctx, entityID, observation)

	case "memory_statistics":
		return m.GetStatistics(ctx)

	case "memory_save":
		return map[string]interface{}{"success": true}, m.Save(ctx)

	case "memory_clear":
		return map[string]interface{}{"success": true}, m.Clear(ctx)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (m *MemoryAdapter) GetCapabilities() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	caps := map[string]interface{}{
		"name":               "memory",
		"max_entities":       m.config.MaxEntities,
		"max_relations":      m.config.MaxRelations,
		"persistence":        m.config.EnablePersistence,
		"auto_save_interval": m.config.AutoSaveInterval.String(),
		"tools":              len(m.GetMCPTools()),
	}

	if m.initialized {
		caps["entity_count"] = len(m.graph.Entities)
		caps["relation_count"] = len(m.graph.Relations)
		caps["version"] = m.graph.Version
	}

	return caps
}

// MarshalJSON implements custom JSON marshaling
func (m *MemoryAdapter) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"initialized":  m.initialized,
		"capabilities": m.GetCapabilities(),
	})
}
