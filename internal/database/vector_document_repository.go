package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VectorDocument represents a document with vector embeddings for semantic search
type VectorDocument struct {
	ID                string          `db:"id" json:"id"`
	Title             string          `db:"title" json:"title"`
	Content           string          `db:"content" json:"content"`
	Metadata          json.RawMessage `db:"metadata" json:"metadata"`
	EmbeddingID       *string         `db:"embedding_id" json:"embedding_id,omitempty"`
	EmbeddingProvider string          `db:"embedding_provider" json:"embedding_provider"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
}

// VectorSearchResult represents a document matched by vector similarity search
type VectorSearchResult struct {
	Document   VectorDocument `json:"document"`
	Similarity float64        `json:"similarity"`
}

// VectorDocumentFilter contains filter options for listing documents
type VectorDocumentFilter struct {
	Provider  string
	TitleLike string
	Limit     int
	Offset    int
}

// VectorDocumentRepository handles vector document database operations
type VectorDocumentRepository struct {
	pool *pgxpool.Pool
}

// NewVectorDocumentRepository creates a new vector document repository
func NewVectorDocumentRepository(pool *pgxpool.Pool) *VectorDocumentRepository {
	return &VectorDocumentRepository{pool: pool}
}

// Create creates a new vector document
func (r *VectorDocumentRepository) Create(ctx context.Context, doc *VectorDocument) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.Metadata == nil {
		doc.Metadata = json.RawMessage("{}")
	}
	if doc.EmbeddingProvider == "" {
		doc.EmbeddingProvider = "pgvector"
	}

	query := `
		INSERT INTO vector_documents (id, title, content, metadata, embedding_id, embedding_provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		doc.ID, doc.Title, doc.Content, doc.Metadata,
		doc.EmbeddingID, doc.EmbeddingProvider, now, now)
	if err != nil {
		return fmt.Errorf("failed to create vector document: %w", err)
	}

	doc.CreatedAt = now
	doc.UpdatedAt = now
	return nil
}

// GetByID retrieves a vector document by ID
func (r *VectorDocumentRepository) GetByID(ctx context.Context, id string) (*VectorDocument, error) {
	query := `
		SELECT id, title, content, metadata, embedding_id, embedding_provider, created_at, updated_at
		FROM vector_documents WHERE id = $1
	`
	var doc VectorDocument
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.Title, &doc.Content, &doc.Metadata,
		&doc.EmbeddingID, &doc.EmbeddingProvider, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get vector document: %w", err)
	}
	return &doc, nil
}

// Update updates an existing vector document
func (r *VectorDocumentRepository) Update(ctx context.Context, doc *VectorDocument) error {
	query := `
		UPDATE vector_documents
		SET title = $2, content = $3, metadata = $4, embedding_id = $5, embedding_provider = $6, updated_at = $7
		WHERE id = $1
	`
	now := time.Now()
	result, err := r.pool.Exec(ctx, query,
		doc.ID, doc.Title, doc.Content, doc.Metadata,
		doc.EmbeddingID, doc.EmbeddingProvider, now)
	if err != nil {
		return fmt.Errorf("failed to update vector document: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vector document not found: %s", doc.ID)
	}
	doc.UpdatedAt = now
	return nil
}

// Delete removes a vector document by ID
func (r *VectorDocumentRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM vector_documents WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete vector document: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vector document not found: %s", id)
	}
	return nil
}

// List retrieves vector documents with optional filtering
func (r *VectorDocumentRepository) List(ctx context.Context, filter VectorDocumentFilter) ([]*VectorDocument, error) {
	query := `
		SELECT id, title, content, metadata, embedding_id, embedding_provider, created_at, updated_at
		FROM vector_documents WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter.Provider != "" {
		query += fmt.Sprintf(" AND embedding_provider = $%d", argNum)
		args = append(args, filter.Provider)
		argNum++
	}

	if filter.TitleLike != "" {
		query += fmt.Sprintf(" AND title ILIKE $%d", argNum)
		args = append(args, "%"+filter.TitleLike+"%")
		argNum++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list vector documents: %w", err)
	}
	defer rows.Close()

	var docs []*VectorDocument
	for rows.Next() {
		var doc VectorDocument
		if err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Content, &doc.Metadata,
			&doc.EmbeddingID, &doc.EmbeddingProvider, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan vector document: %w", err)
		}
		docs = append(docs, &doc)
	}

	return docs, rows.Err()
}

// Count returns the total number of vector documents
func (r *VectorDocumentRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM vector_documents").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count vector documents: %w", err)
	}
	return count, nil
}

// CountByProvider returns the count of documents per embedding provider
func (r *VectorDocumentRepository) CountByProvider(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT embedding_provider, COUNT(*) as count
		FROM vector_documents
		GROUP BY embedding_provider
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count by provider: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var provider string
		var count int64
		if err := rows.Scan(&provider, &count); err != nil {
			return nil, fmt.Errorf("failed to scan provider count: %w", err)
		}
		counts[provider] = count
	}

	return counts, rows.Err()
}

// SearchByTitle searches documents by title (case-insensitive)
func (r *VectorDocumentRepository) SearchByTitle(ctx context.Context, titleQuery string, limit int) ([]*VectorDocument, error) {
	query := `
		SELECT id, title, content, metadata, embedding_id, embedding_provider, created_at, updated_at
		FROM vector_documents
		WHERE title ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, "%"+titleQuery+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents by title: %w", err)
	}
	defer rows.Close()

	var docs []*VectorDocument
	for rows.Next() {
		var doc VectorDocument
		if err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Content, &doc.Metadata,
			&doc.EmbeddingID, &doc.EmbeddingProvider, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}

	return docs, rows.Err()
}

// GetDocumentsWithoutEmbeddings retrieves documents that don't have embeddings yet
func (r *VectorDocumentRepository) GetDocumentsWithoutEmbeddings(ctx context.Context, limit int) ([]*VectorDocument, error) {
	query := `
		SELECT id, title, content, metadata, embedding_id, embedding_provider, created_at, updated_at
		FROM vector_documents
		WHERE embedding IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents without embeddings: %w", err)
	}
	defer rows.Close()

	var docs []*VectorDocument
	for rows.Next() {
		var doc VectorDocument
		if err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Content, &doc.Metadata,
			&doc.EmbeddingID, &doc.EmbeddingProvider, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}

	return docs, rows.Err()
}

// UpdateMetadata updates only the metadata field of a document
func (r *VectorDocumentRepository) UpdateMetadata(ctx context.Context, id string, metadata json.RawMessage) error {
	query := `
		UPDATE vector_documents
		SET metadata = $2, updated_at = $3
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, id, metadata, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vector document not found: %s", id)
	}
	return nil
}

// DeleteByProvider deletes all documents with a specific embedding provider
func (r *VectorDocumentRepository) DeleteByProvider(ctx context.Context, provider string) (int64, error) {
	result, err := r.pool.Exec(ctx, "DELETE FROM vector_documents WHERE embedding_provider = $1", provider)
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents by provider: %w", err)
	}
	return result.RowsAffected(), nil
}

// DeleteOlderThan deletes documents older than the specified time
func (r *VectorDocumentRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.pool.Exec(ctx, "DELETE FROM vector_documents WHERE created_at < $1", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old documents: %w", err)
	}
	return result.RowsAffected(), nil
}

// BulkCreate creates multiple documents in a single transaction
func (r *VectorDocumentRepository) BulkCreate(ctx context.Context, docs []*VectorDocument) error {
	if len(docs) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() //nolint:errcheck

	query := `
		INSERT INTO vector_documents (id, title, content, metadata, embedding_id, embedding_provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	for _, doc := range docs {
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}
		if doc.Metadata == nil {
			doc.Metadata = json.RawMessage("{}")
		}
		if doc.EmbeddingProvider == "" {
			doc.EmbeddingProvider = "pgvector"
		}

		_, err := tx.Exec(ctx, query,
			doc.ID, doc.Title, doc.Content, doc.Metadata,
			doc.EmbeddingID, doc.EmbeddingProvider, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert document %s: %w", doc.ID, err)
		}
		doc.CreatedAt = now
		doc.UpdatedAt = now
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
