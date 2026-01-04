package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// CogneeMemory represents a Cognee memory entry in the database
type CogneeMemory struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	DatasetName string                 `json:"dataset_name"`
	ContentType string                 `json:"content_type"`
	Content     string                 `json:"content"`
	VectorID    *string                `json:"vector_id"`
	GraphNodes  map[string]interface{} `json:"graph_nodes"`
	SearchKey   *string                `json:"search_key"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CogneeMemoryRepository handles Cognee memory database operations
type CogneeMemoryRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewCogneeMemoryRepository creates a new CogneeMemoryRepository
func NewCogneeMemoryRepository(pool *pgxpool.Pool, log *logrus.Logger) *CogneeMemoryRepository {
	return &CogneeMemoryRepository{
		pool: pool,
		log:  log,
	}
}

// Create creates a new Cognee memory in the database
func (r *CogneeMemoryRepository) Create(ctx context.Context, memory *CogneeMemory) error {
	query := `
		INSERT INTO cognee_memories (session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	graphNodesJSON, _ := json.Marshal(memory.GraphNodes)

	err := r.pool.QueryRow(ctx, query,
		memory.SessionID, memory.DatasetName, memory.ContentType, memory.Content,
		memory.VectorID, graphNodesJSON, memory.SearchKey,
	).Scan(&memory.ID, &memory.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create cognee memory: %w", err)
	}

	return nil
}

// GetByID retrieves a memory by its ID
func (r *CogneeMemoryRepository) GetByID(ctx context.Context, id string) (*CogneeMemory, error) {
	query := `
		SELECT id, session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key, created_at
		FROM cognee_memories
		WHERE id = $1
	`

	memory := &CogneeMemory{}
	var graphNodesJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&memory.ID, &memory.SessionID, &memory.DatasetName, &memory.ContentType,
		&memory.Content, &memory.VectorID, &graphNodesJSON, &memory.SearchKey, &memory.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("cognee memory not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cognee memory: %w", err)
	}

	if len(graphNodesJSON) > 0 {
		json.Unmarshal(graphNodesJSON, &memory.GraphNodes)
	}

	return memory, nil
}

// GetBySessionID retrieves all memories for a session
func (r *CogneeMemoryRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*CogneeMemory, error) {
	query := `
		SELECT id, session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key, created_at
		FROM cognee_memories
		WHERE session_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cognee memories: %w", err)
	}
	defer rows.Close()

	memories := []*CogneeMemory{}
	for rows.Next() {
		memory := &CogneeMemory{}
		var graphNodesJSON []byte

		err := rows.Scan(
			&memory.ID, &memory.SessionID, &memory.DatasetName, &memory.ContentType,
			&memory.Content, &memory.VectorID, &graphNodesJSON, &memory.SearchKey, &memory.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if len(graphNodesJSON) > 0 {
			json.Unmarshal(graphNodesJSON, &memory.GraphNodes)
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// GetByDatasetName retrieves all memories for a dataset
func (r *CogneeMemoryRepository) GetByDatasetName(ctx context.Context, datasetName string, limit, offset int) ([]*CogneeMemory, int, error) {
	countQuery := `SELECT COUNT(*) FROM cognee_memories WHERE dataset_name = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, datasetName).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	query := `
		SELECT id, session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key, created_at
		FROM cognee_memories
		WHERE dataset_name = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, datasetName, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	memories := []*CogneeMemory{}
	for rows.Next() {
		memory := &CogneeMemory{}
		var graphNodesJSON []byte

		err := rows.Scan(
			&memory.ID, &memory.SessionID, &memory.DatasetName, &memory.ContentType,
			&memory.Content, &memory.VectorID, &graphNodesJSON, &memory.SearchKey, &memory.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if len(graphNodesJSON) > 0 {
			json.Unmarshal(graphNodesJSON, &memory.GraphNodes)
		}

		memories = append(memories, memory)
	}

	return memories, total, nil
}

// SearchByKey searches memories by search key
func (r *CogneeMemoryRepository) SearchByKey(ctx context.Context, searchKey string, limit int) ([]*CogneeMemory, error) {
	query := `
		SELECT id, session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key, created_at
		FROM cognee_memories
		WHERE search_key ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, searchKey, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}
	defer rows.Close()

	memories := []*CogneeMemory{}
	for rows.Next() {
		memory := &CogneeMemory{}
		var graphNodesJSON []byte

		err := rows.Scan(
			&memory.ID, &memory.SessionID, &memory.DatasetName, &memory.ContentType,
			&memory.Content, &memory.VectorID, &graphNodesJSON, &memory.SearchKey, &memory.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if len(graphNodesJSON) > 0 {
			json.Unmarshal(graphNodesJSON, &memory.GraphNodes)
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// SearchByContent searches memories by content
func (r *CogneeMemoryRepository) SearchByContent(ctx context.Context, searchTerm string, limit int) ([]*CogneeMemory, error) {
	query := `
		SELECT id, session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key, created_at
		FROM cognee_memories
		WHERE content ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories by content: %w", err)
	}
	defer rows.Close()

	memories := []*CogneeMemory{}
	for rows.Next() {
		memory := &CogneeMemory{}
		var graphNodesJSON []byte

		err := rows.Scan(
			&memory.ID, &memory.SessionID, &memory.DatasetName, &memory.ContentType,
			&memory.Content, &memory.VectorID, &graphNodesJSON, &memory.SearchKey, &memory.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if len(graphNodesJSON) > 0 {
			json.Unmarshal(graphNodesJSON, &memory.GraphNodes)
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// Update updates an existing memory
func (r *CogneeMemoryRepository) Update(ctx context.Context, memory *CogneeMemory) error {
	query := `
		UPDATE cognee_memories
		SET content_type = $2, content = $3, vector_id = $4, graph_nodes = $5, search_key = $6
		WHERE id = $1
	`

	graphNodesJSON, _ := json.Marshal(memory.GraphNodes)

	result, err := r.pool.Exec(ctx, query,
		memory.ID, memory.ContentType, memory.Content, memory.VectorID, graphNodesJSON, memory.SearchKey,
	)
	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("memory not found: %s", memory.ID)
	}

	return nil
}

// UpdateVectorID updates the vector ID for a memory
func (r *CogneeMemoryRepository) UpdateVectorID(ctx context.Context, id, vectorID string) error {
	query := `UPDATE cognee_memories SET vector_id = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, vectorID)
	if err != nil {
		return fmt.Errorf("failed to update vector ID: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("memory not found: %s", id)
	}

	return nil
}

// UpdateGraphNodes updates the graph nodes for a memory
func (r *CogneeMemoryRepository) UpdateGraphNodes(ctx context.Context, id string, graphNodes map[string]interface{}) error {
	query := `UPDATE cognee_memories SET graph_nodes = $2 WHERE id = $1`

	graphNodesJSON, _ := json.Marshal(graphNodes)

	result, err := r.pool.Exec(ctx, query, id, graphNodesJSON)
	if err != nil {
		return fmt.Errorf("failed to update graph nodes: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("memory not found: %s", id)
	}

	return nil
}

// Delete deletes a memory by its ID
func (r *CogneeMemoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM cognee_memories WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("memory not found: %s", id)
	}

	return nil
}

// DeleteBySessionID deletes all memories for a session
func (r *CogneeMemoryRepository) DeleteBySessionID(ctx context.Context, sessionID string) (int64, error) {
	query := `DELETE FROM cognee_memories WHERE session_id = $1`

	result, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete session memories: %w", err)
	}

	return result.RowsAffected(), nil
}

// DeleteByDatasetName deletes all memories for a dataset
func (r *CogneeMemoryRepository) DeleteByDatasetName(ctx context.Context, datasetName string) (int64, error) {
	query := `DELETE FROM cognee_memories WHERE dataset_name = $1`

	result, err := r.pool.Exec(ctx, query, datasetName)
	if err != nil {
		return 0, fmt.Errorf("failed to delete dataset memories: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetByVectorID retrieves a memory by its vector ID
func (r *CogneeMemoryRepository) GetByVectorID(ctx context.Context, vectorID string) (*CogneeMemory, error) {
	query := `
		SELECT id, session_id, dataset_name, content_type, content, vector_id, graph_nodes, search_key, created_at
		FROM cognee_memories
		WHERE vector_id = $1
	`

	memory := &CogneeMemory{}
	var graphNodesJSON []byte

	err := r.pool.QueryRow(ctx, query, vectorID).Scan(
		&memory.ID, &memory.SessionID, &memory.DatasetName, &memory.ContentType,
		&memory.Content, &memory.VectorID, &graphNodesJSON, &memory.SearchKey, &memory.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("memory not found for vector ID: %s", vectorID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory by vector ID: %w", err)
	}

	if len(graphNodesJSON) > 0 {
		json.Unmarshal(graphNodesJSON, &memory.GraphNodes)
	}

	return memory, nil
}

// ListDatasets returns a list of all unique dataset names
func (r *CogneeMemoryRepository) ListDatasets(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT dataset_name FROM cognee_memories ORDER BY dataset_name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}
	defer rows.Close()

	datasets := []string{}
	for rows.Next() {
		var datasetName string
		if err := rows.Scan(&datasetName); err != nil {
			return nil, fmt.Errorf("failed to scan dataset name: %w", err)
		}
		datasets = append(datasets, datasetName)
	}

	return datasets, nil
}

// GetDatasetStats returns statistics for a dataset
func (r *CogneeMemoryRepository) GetDatasetStats(ctx context.Context, datasetName string) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_memories,
			COUNT(DISTINCT session_id) as unique_sessions,
			COUNT(vector_id) as indexed_memories,
			MIN(created_at) as oldest_memory,
			MAX(created_at) as newest_memory
		FROM cognee_memories
		WHERE dataset_name = $1
	`

	var totalMemories, uniqueSessions, indexedMemories int
	var oldestMemory, newestMemory *time.Time

	err := r.pool.QueryRow(ctx, query, datasetName).Scan(
		&totalMemories, &uniqueSessions, &indexedMemories, &oldestMemory, &newestMemory,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset stats: %w", err)
	}

	stats := map[string]interface{}{
		"dataset_name":     datasetName,
		"total_memories":   totalMemories,
		"unique_sessions":  uniqueSessions,
		"indexed_memories": indexedMemories,
		"oldest_memory":    oldestMemory,
		"newest_memory":    newestMemory,
	}

	return stats, nil
}
