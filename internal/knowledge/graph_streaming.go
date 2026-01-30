package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dev.helix.agent/internal/messaging"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sirupsen/logrus"
)

// StreamingKnowledgeGraph manages real-time knowledge graph updates via Kafka -> Neo4j
type StreamingKnowledgeGraph struct {
	driver       neo4j.DriverWithContext
	database     string
	broker       messaging.MessageBroker
	logger       *logrus.Logger
	stopCh       chan struct{}
	entityTopic  string
	memoryTopic  string
	debateTopic  string
}

// EntityUpdateType defines the type of entity update
type EntityUpdateType string

const (
	// EntityCreated indicates a new entity was created
	EntityCreated EntityUpdateType = "entity.created"

	// EntityUpdated indicates an existing entity was updated
	EntityUpdated EntityUpdateType = "entity.updated"

	// EntityDeleted indicates an entity was deleted
	EntityDeleted EntityUpdateType = "entity.deleted"

	// EntityMerged indicates entities were merged
	EntityMerged EntityUpdateType = "entity.merged"

	// RelationshipCreated indicates a new relationship was created
	RelationshipCreated EntityUpdateType = "relationship.created"

	// RelationshipUpdated indicates a relationship was updated
	RelationshipUpdated EntityUpdateType = "relationship.updated"

	// RelationshipDeleted indicates a relationship was deleted
	RelationshipDeleted EntityUpdateType = "relationship.deleted"
)

// EntityUpdate represents an entity update event from Kafka
type EntityUpdate struct {
	UpdateID       string                 `json:"update_id"`
	UpdateType     EntityUpdateType       `json:"update_type"`
	Timestamp      time.Time              `json:"timestamp"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	UserID         string                 `json:"user_id"`
	Entity         *GraphEntity           `json:"entity,omitempty"`
	Relationship   *GraphRelationship     `json:"relationship,omitempty"`
	SourceID       string                 `json:"source_id,omitempty"`  // For entity merges
	TargetID       string                 `json:"target_id,omitempty"`  // For entity merges
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// GraphEntity represents an entity in the knowledge graph
type GraphEntity struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Value      string                 `json:"value"`
	Properties map[string]interface{} `json:"properties"`
	Confidence float64                `json:"confidence"`
	Importance float64                `json:"importance"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// GraphRelationship represents a relationship between entities
type GraphRelationship struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"` // RELATED_TO, MENTIONED_IN, etc.
	SourceID        string                 `json:"source_id"`
	TargetID        string                 `json:"target_id"`
	Strength        float64                `json:"strength"`
	CooccurrenceCount int                  `json:"cooccurrence_count"`
	Contexts        []string               `json:"contexts,omitempty"`
	Properties      map[string]interface{} `json:"properties,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// GraphStreamingConfig defines configuration for knowledge graph streaming
type GraphStreamingConfig struct {
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	Neo4jDatabase string
	EntityTopic   string
	MemoryTopic   string
	DebateTopic   string
}

// NewStreamingKnowledgeGraph creates a new streaming knowledge graph
func NewStreamingKnowledgeGraph(
	config GraphStreamingConfig,
	broker messaging.MessageBroker,
	logger *logrus.Logger,
) (*StreamingKnowledgeGraph, error) {
	if logger == nil {
		logger = logrus.New()
	}

	// Create Neo4j driver
	auth := neo4j.BasicAuth(config.Neo4jUser, config.Neo4jPassword, "")
	driver, err := neo4j.NewDriverWithContext(config.Neo4jURI, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	skg := &StreamingKnowledgeGraph{
		driver:       driver,
		database:     config.Neo4jDatabase,
		broker:       broker,
		logger:       logger,
		stopCh:       make(chan struct{}),
		entityTopic:  config.EntityTopic,
		memoryTopic:  config.MemoryTopic,
		debateTopic:  config.DebateTopic,
	}

	// Initialize graph schema
	if err := skg.initializeSchema(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"uri":      config.Neo4jURI,
		"database": config.Neo4jDatabase,
	}).Info("Streaming knowledge graph initialized")

	return skg, nil
}

// initializeSchema creates indexes and constraints in Neo4j
func (skg *StreamingKnowledgeGraph) initializeSchema(ctx context.Context) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	// Create constraints and indexes
	queries := []string{
		// Entity constraints
		"CREATE CONSTRAINT entity_id_unique IF NOT EXISTS FOR (e:Entity) REQUIRE e.id IS UNIQUE",

		// Entity indexes
		"CREATE INDEX entity_type_idx IF NOT EXISTS FOR (e:Entity) ON (e.type)",
		"CREATE INDEX entity_name_idx IF NOT EXISTS FOR (e:Entity) ON (e.name)",
		"CREATE INDEX entity_importance_idx IF NOT EXISTS FOR (e:Entity) ON (e.importance)",

		// Conversation constraints
		"CREATE CONSTRAINT conversation_id_unique IF NOT EXISTS FOR (c:Conversation) REQUIRE c.id IS UNIQUE",

		// Relationship indexes
		"CREATE INDEX rel_strength_idx IF NOT EXISTS FOR ()-[r:RELATED_TO]-() ON (r.strength)",
		"CREATE INDEX rel_cooccurrence_idx IF NOT EXISTS FOR ()-[r:RELATED_TO]-() ON (r.cooccurrence_count)",
	}

	for _, query := range queries {
		_, err := session.Run(ctx, query, nil)
		if err != nil {
			skg.logger.WithError(err).WithField("query", query).Warn("Failed to create constraint/index")
			// Continue with other queries even if one fails
		}
	}

	skg.logger.Info("Neo4j schema initialized")
	return nil
}

// StartStreaming begins consuming updates from Kafka and applying them to Neo4j
func (skg *StreamingKnowledgeGraph) StartStreaming(ctx context.Context) error {
	skg.logger.Info("Starting knowledge graph streaming")

	// Create message handler
	handler := func(ctx context.Context, msg *messaging.Message) error {
		return skg.processUpdate(ctx, msg)
	}

	// Subscribe to entity updates topic
	_, err := skg.broker.Subscribe(ctx, skg.entityTopic, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to entity topic: %w", err)
	}

	skg.logger.WithField("topic", skg.entityTopic).Info("Subscribed to entity updates")

	return nil
}

// processUpdate processes a single entity update
func (skg *StreamingKnowledgeGraph) processUpdate(ctx context.Context, msg *messaging.Message) error {
	// Parse update
	var update EntityUpdate
	if err := json.Unmarshal(msg.Payload, &update); err != nil {
		return fmt.Errorf("failed to unmarshal update: %w", err)
	}

	skg.logger.WithFields(logrus.Fields{
		"update_id":   update.UpdateID,
		"update_type": update.UpdateType,
		"entity_id":   getEntityID(&update),
	}).Debug("Processing entity update")

	// Apply update based on type
	switch update.UpdateType {
	case EntityCreated:
		return skg.createEntity(ctx, update.Entity)

	case EntityUpdated:
		return skg.updateEntity(ctx, update.Entity)

	case EntityDeleted:
		return skg.deleteEntity(ctx, update.Entity.ID)

	case EntityMerged:
		return skg.mergeEntities(ctx, update.SourceID, update.TargetID)

	case RelationshipCreated:
		return skg.createRelationship(ctx, update.Relationship)

	case RelationshipUpdated:
		return skg.updateRelationship(ctx, update.Relationship)

	case RelationshipDeleted:
		return skg.deleteRelationship(ctx, update.Relationship.ID)

	default:
		return fmt.Errorf("unknown update type: %s", update.UpdateType)
	}
}

// createEntity creates a new entity node in Neo4j
func (skg *StreamingKnowledgeGraph) createEntity(ctx context.Context, entity *GraphEntity) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MERGE (e:Entity {id: $id})
		SET e.name = $name,
		    e.type = $type,
		    e.value = $value,
		    e.properties = $properties,
		    e.confidence = $confidence,
		    e.importance = $importance,
		    e.created_at = datetime($created_at),
		    e.updated_at = datetime($updated_at)
		RETURN e.id
	`

	params := map[string]interface{}{
		"id":         entity.ID,
		"name":       entity.Name,
		"type":       entity.Type,
		"value":      entity.Value,
		"properties": entity.Properties,
		"confidence": entity.Confidence,
		"importance": entity.Importance,
		"created_at": entity.CreatedAt.Format(time.RFC3339),
		"updated_at": entity.UpdatedAt.Format(time.RFC3339),
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	skg.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"entity_type": entity.Type,
		"entity_name": entity.Name,
	}).Debug("Entity created in graph")

	return nil
}

// updateEntity updates an existing entity node
func (skg *StreamingKnowledgeGraph) updateEntity(ctx context.Context, entity *GraphEntity) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (e:Entity {id: $id})
		SET e.name = $name,
		    e.value = $value,
		    e.properties = $properties,
		    e.confidence = $confidence,
		    e.importance = $importance,
		    e.updated_at = datetime($updated_at)
		RETURN e.id
	`

	params := map[string]interface{}{
		"id":         entity.ID,
		"name":       entity.Name,
		"value":      entity.Value,
		"properties": entity.Properties,
		"confidence": entity.Confidence,
		"importance": entity.Importance,
		"updated_at": entity.UpdatedAt.Format(time.RFC3339),
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	skg.logger.WithField("entity_id", entity.ID).Debug("Entity updated in graph")
	return nil
}

// deleteEntity deletes an entity node and its relationships
func (skg *StreamingKnowledgeGraph) deleteEntity(ctx context.Context, entityID string) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (e:Entity {id: $id})
		DETACH DELETE e
	`

	params := map[string]interface{}{
		"id": entityID,
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	skg.logger.WithField("entity_id", entityID).Debug("Entity deleted from graph")
	return nil
}

// mergeEntities merges two entity nodes
func (skg *StreamingKnowledgeGraph) mergeEntities(ctx context.Context, sourceID, targetID string) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	// Transfer relationships and delete source
	cypher := `
		MATCH (source:Entity {id: $source_id})
		MATCH (target:Entity {id: $target_id})

		// Transfer relationships
		MATCH (source)-[r]->(other)
		WHERE other <> target
		MERGE (target)-[r2:RELATED_TO]->(other)
		ON CREATE SET r2 = properties(r)
		ON MATCH SET r2.strength = r2.strength + r.strength,
		             r2.cooccurrence_count = r2.cooccurrence_count + r.cooccurrence_count

		// Delete source
		DETACH DELETE source

		// Update target importance
		SET target.importance = target.importance + source.importance
		RETURN target.id
	`

	params := map[string]interface{}{
		"source_id": sourceID,
		"target_id": targetID,
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to merge entities: %w", err)
	}

	skg.logger.WithFields(logrus.Fields{
		"source_id": sourceID,
		"target_id": targetID,
	}).Debug("Entities merged in graph")

	return nil
}

// createRelationship creates a relationship between entities
func (skg *StreamingKnowledgeGraph) createRelationship(ctx context.Context, rel *GraphRelationship) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (source:Entity {id: $source_id})
		MATCH (target:Entity {id: $target_id})
		MERGE (source)-[r:RELATED_TO]->(target)
		ON CREATE SET r.id = $id,
		              r.type = $type,
		              r.strength = $strength,
		              r.cooccurrence_count = $cooccurrence_count,
		              r.contexts = $contexts,
		              r.properties = $properties,
		              r.created_at = datetime($created_at),
		              r.updated_at = datetime($updated_at)
		ON MATCH SET r.strength = r.strength + $strength,
		             r.cooccurrence_count = r.cooccurrence_count + $cooccurrence_count,
		             r.updated_at = datetime($updated_at)
		RETURN r.id
	`

	params := map[string]interface{}{
		"id":                 rel.ID,
		"source_id":          rel.SourceID,
		"target_id":          rel.TargetID,
		"type":               rel.Type,
		"strength":           rel.Strength,
		"cooccurrence_count": rel.CooccurrenceCount,
		"contexts":           rel.Contexts,
		"properties":         rel.Properties,
		"created_at":         rel.CreatedAt.Format(time.RFC3339),
		"updated_at":         rel.UpdatedAt.Format(time.RFC3339),
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	skg.logger.WithFields(logrus.Fields{
		"relationship_id": rel.ID,
		"source_id":       rel.SourceID,
		"target_id":       rel.TargetID,
	}).Debug("Relationship created in graph")

	return nil
}

// updateRelationship updates an existing relationship
func (skg *StreamingKnowledgeGraph) updateRelationship(ctx context.Context, rel *GraphRelationship) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MATCH ()-[r:RELATED_TO {id: $id}]->()
		SET r.strength = $strength,
		    r.cooccurrence_count = $cooccurrence_count,
		    r.contexts = $contexts,
		    r.properties = $properties,
		    r.updated_at = datetime($updated_at)
		RETURN r.id
	`

	params := map[string]interface{}{
		"id":                 rel.ID,
		"strength":           rel.Strength,
		"cooccurrence_count": rel.CooccurrenceCount,
		"contexts":           rel.Contexts,
		"properties":         rel.Properties,
		"updated_at":         rel.UpdatedAt.Format(time.RFC3339),
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to update relationship: %w", err)
	}

	skg.logger.WithField("relationship_id", rel.ID).Debug("Relationship updated in graph")
	return nil
}

// deleteRelationship deletes a relationship
func (skg *StreamingKnowledgeGraph) deleteRelationship(ctx context.Context, relationshipID string) error {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MATCH ()-[r:RELATED_TO {id: $id}]->()
		DELETE r
	`

	params := map[string]interface{}{
		"id": relationshipID,
	}

	_, err := session.Run(ctx, cypher, params)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	skg.logger.WithField("relationship_id", relationshipID).Debug("Relationship deleted from graph")
	return nil
}

// GetEntity retrieves an entity from the graph
func (skg *StreamingKnowledgeGraph) GetEntity(ctx context.Context, entityID string) (*GraphEntity, error) {
	session := skg.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: skg.database,
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (e:Entity {id: $id})
		RETURN e
	`

	params := map[string]interface{}{
		"id": entityID,
	}

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		node, _ := record.Get("e")
		return nodeToEntity(node.(neo4j.Node)), nil
	}

	return nil, fmt.Errorf("entity not found: %s", entityID)
}

// Stop stops the streaming consumer
func (skg *StreamingKnowledgeGraph) Stop(ctx context.Context) error {
	skg.logger.Info("Stopping knowledge graph streaming")

	close(skg.stopCh)

	if err := skg.driver.Close(ctx); err != nil {
		return fmt.Errorf("failed to close Neo4j driver: %w", err)
	}

	return nil
}

// Helper functions

func getEntityID(update *EntityUpdate) string {
	if update.Entity != nil {
		return update.Entity.ID
	}
	if update.Relationship != nil {
		return update.Relationship.ID
	}
	return ""
}

func nodeToEntity(node neo4j.Node) *GraphEntity {
	props := node.Props

	entity := &GraphEntity{
		ID:         getString(props, "id"),
		Type:       getString(props, "type"),
		Name:       getString(props, "name"),
		Value:      getString(props, "value"),
		Confidence: getFloat64(props, "confidence"),
		Importance: getFloat64(props, "importance"),
	}

	// Parse timestamps
	if createdAt, ok := props["created_at"].(time.Time); ok {
		entity.CreatedAt = createdAt
	}
	if updatedAt, ok := props["updated_at"].(time.Time); ok {
		entity.UpdatedAt = updatedAt
	}

	// Extract properties map
	if properties, ok := props["properties"].(map[string]interface{}); ok {
		entity.Properties = properties
	}

	return entity
}

func getString(props map[string]interface{}, key string) string {
	if val, ok := props[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(props map[string]interface{}, key string) float64 {
	if val, ok := props[key].(float64); ok {
		return val
	}
	return 0.0
}
