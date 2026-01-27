// Package pgvector provides a client for PostgreSQL with pgvector extension.
package pgvector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Client provides an interface to interact with PostgreSQL pgvector.
type Client struct {
	config    *Config
	pool      *pgxpool.Pool
	logger    *logrus.Logger
	mu        sync.RWMutex
	connected bool
}

// Config holds pgvector configuration.
type Config struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	SSLMode         string        `json:"ssl_mode"`
	MaxConns        int32         `json:"max_conns"`
	MinConns        int32         `json:"min_conns"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
}

// DefaultConfig returns default pgvector configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Database:        "postgres",
		SSLMode:         "disable",
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
		ConnectTimeout:  30 * time.Second,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("invalid port")
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}
	return nil
}

// ConnectionString returns the PostgreSQL connection string.
func (c *Config) ConnectionString() string {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s",
		c.Host, c.Port, c.User, c.Database)
	if c.Password != "" {
		connStr += fmt.Sprintf(" password=%s", c.Password)
	}
	if c.SSLMode != "" {
		connStr += fmt.Sprintf(" sslmode=%s", c.SSLMode)
	}
	if c.ConnectTimeout > 0 {
		connStr += fmt.Sprintf(" connect_timeout=%d", int(c.ConnectTimeout.Seconds()))
	}
	return connStr
}

// NewClient creates a new pgvector client.
func NewClient(config *Config, logger *logrus.Logger) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &Client{
		config:    config,
		logger:    logger,
		connected: false,
	}, nil
}

// Connect establishes connection to PostgreSQL and ensures pgvector extension exists.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	poolConfig, err := pgxpool.ParseConfig(c.config.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = c.config.MaxConns
	poolConfig.MinConns = c.config.MinConns
	poolConfig.MaxConnLifetime = c.config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = c.config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Ensure pgvector extension exists
	_, err = pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		pool.Close()
		return fmt.Errorf("failed to enable vector extension: %w", err)
	}

	c.pool = pool
	c.connected = true
	c.logger.Info("Connected to PostgreSQL with pgvector")
	return nil
}

// Close closes the database connection pool.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pool != nil {
		c.pool.Close()
		c.pool = nil
	}
	c.connected = false
	return nil
}

// IsConnected returns whether the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// HealthCheck checks the health of the database connection.
func (c *Client) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.pool == nil {
		return fmt.Errorf("not connected")
	}

	return c.pool.Ping(ctx)
}

// DistanceMetric represents the distance metric for vector search.
type DistanceMetric string

const (
	// DistanceL2 uses Euclidean distance (default).
	DistanceL2 DistanceMetric = "l2"
	// DistanceIP uses inner product (cosine for normalized vectors).
	DistanceIP DistanceMetric = "ip"
	// DistanceCosine uses cosine distance.
	DistanceCosine DistanceMetric = "cosine"
)

// IndexType represents the vector index type.
type IndexType string

const (
	// IndexTypeIVFFlat uses IVF with flat storage.
	IndexTypeIVFFlat IndexType = "ivfflat"
	// IndexTypeHNSW uses hierarchical navigable small world.
	IndexTypeHNSW IndexType = "hnsw"
)

// TableSchema defines a pgvector table schema.
type TableSchema struct {
	TableName       string
	VectorColumn    string
	Dimension       int
	IDColumn        string
	MetadataColumns []ColumnDef
}

// ColumnDef defines a column definition.
type ColumnDef struct {
	Name     string
	Type     string
	Nullable bool
}

// CreateTable creates a vector table with the specified schema.
func (c *Client) CreateTable(ctx context.Context, schema *TableSchema) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Build column definitions
	columns := []string{
		fmt.Sprintf("%s TEXT PRIMARY KEY", schema.IDColumn),
		fmt.Sprintf("%s vector(%d)", schema.VectorColumn, schema.Dimension),
	}

	for _, col := range schema.MetadataColumns {
		nullable := ""
		if !col.Nullable {
			nullable = " NOT NULL"
		}
		columns = append(columns, fmt.Sprintf("%s %s%s", col.Name, col.Type, nullable))
	}

	// Add timestamps
	columns = append(columns, "created_at TIMESTAMPTZ DEFAULT NOW()")
	columns = append(columns, "updated_at TIMESTAMPTZ DEFAULT NOW()")

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
		schema.TableName, strings.Join(columns, ", "))

	_, err := c.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	c.logger.WithField("table", schema.TableName).Info("Table created")
	return nil
}

// DropTable drops a vector table.
func (c *Client) DropTable(ctx context.Context, tableName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := c.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	c.logger.WithField("table", tableName).Info("Table dropped")
	return nil
}

// TableExists checks if a table exists.
func (c *Client) TableExists(ctx context.Context, tableName string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return false, fmt.Errorf("not connected")
	}

	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	)`

	var exists bool
	err := c.pool.QueryRow(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return exists, nil
}

// CreateIndexRequest represents an index creation request.
type CreateIndexRequest struct {
	TableName      string
	IndexName      string
	VectorColumn   string
	IndexType      IndexType
	Metric         DistanceMetric
	Lists          int // For IVFFlat: number of inverted lists
	M              int // For HNSW: max number of connections
	EfConstruction int // For HNSW: size of dynamic candidate list during construction
}

// CreateIndex creates a vector index.
func (c *Client) CreateIndex(ctx context.Context, req *CreateIndexRequest) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Determine operator class based on metric
	var opClass string
	switch req.Metric {
	case DistanceIP:
		opClass = "vector_ip_ops"
	case DistanceCosine:
		opClass = "vector_cosine_ops"
	default:
		opClass = "vector_l2_ops"
	}

	var query string
	switch req.IndexType {
	case IndexTypeHNSW:
		m := req.M
		if m == 0 {
			m = 16
		}
		efConstruction := req.EfConstruction
		if efConstruction == 0 {
			efConstruction = 64
		}
		query = fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS %s ON %s USING hnsw (%s %s) WITH (m = %d, ef_construction = %d)",
			req.IndexName, req.TableName, req.VectorColumn, opClass, m, efConstruction)
	default: // IVFFlat
		lists := req.Lists
		if lists == 0 {
			lists = 100
		}
		query = fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS %s ON %s USING ivfflat (%s %s) WITH (lists = %d)",
			req.IndexName, req.TableName, req.VectorColumn, opClass, lists)
	}

	_, err := c.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"table":     req.TableName,
		"index":     req.IndexName,
		"indexType": req.IndexType,
	}).Info("Index created")

	return nil
}

// Vector represents a vector record.
type Vector struct {
	ID       string
	Vector   []float32
	Metadata map[string]interface{}
}

// UpsertRequest represents an upsert request.
type UpsertRequest struct {
	TableName    string
	VectorColumn string
	IDColumn     string
	Vectors      []Vector
}

// Upsert inserts or updates vectors in the table.
func (c *Client) Upsert(ctx context.Context, req *UpsertRequest) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected")
	}

	if len(req.Vectors) == 0 {
		return 0, nil
	}

	// Ensure all vectors have IDs
	for i := range req.Vectors {
		if req.Vectors[i].ID == "" {
			req.Vectors[i].ID = uuid.New().String()
		}
	}

	// Build column list from first vector's metadata
	columns := []string{req.IDColumn, req.VectorColumn}
	metadataKeys := make([]string, 0)
	if len(req.Vectors) > 0 && len(req.Vectors[0].Metadata) > 0 {
		for k := range req.Vectors[0].Metadata {
			metadataKeys = append(metadataKeys, k)
			columns = append(columns, k)
		}
	}
	columns = append(columns, "updated_at")

	// Build placeholders and values
	batch := &pgx.Batch{}
	for _, v := range req.Vectors {
		placeholders := make([]string, len(columns))
		values := make([]interface{}, len(columns))

		values[0] = v.ID
		placeholders[0] = "$1"

		values[1] = vectorToString(v.Vector)
		placeholders[1] = "$2"

		for i, k := range metadataKeys {
			values[i+2] = v.Metadata[k]
			placeholders[i+2] = fmt.Sprintf("$%d", i+3)
		}

		values[len(values)-1] = time.Now()
		placeholders[len(placeholders)-1] = fmt.Sprintf("$%d", len(placeholders))

		// Build upsert query
		updateCols := make([]string, 0)
		for i, col := range columns {
			if col != req.IDColumn {
				updateCols = append(updateCols, fmt.Sprintf("%s = %s", col, placeholders[i]))
			}
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
			req.TableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			req.IDColumn,
			strings.Join(updateCols, ", "))

		batch.Queue(query, values...)
	}

	results := c.pool.SendBatch(ctx, batch)
	defer results.Close()

	count := 0
	for range req.Vectors {
		_, err := results.Exec()
		if err != nil {
			return count, fmt.Errorf("failed to upsert vector: %w", err)
		}
		count++
	}

	c.logger.WithFields(logrus.Fields{
		"table": req.TableName,
		"count": count,
	}).Debug("Vectors upserted")

	return count, nil
}

// SearchRequest represents a vector search request.
type SearchRequest struct {
	TableName     string
	VectorColumn  string
	IDColumn      string
	QueryVector   []float32
	Limit         int
	Metric        DistanceMetric
	Filter        string
	OutputColumns []string
}

// SearchResult represents a search result.
type SearchResult struct {
	ID       string
	Distance float32
	Metadata map[string]interface{}
}

// Search performs vector similarity search.
func (c *Client) Search(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Determine distance operator based on metric
	var distanceOp string
	switch req.Metric {
	case DistanceIP:
		distanceOp = "<#>" // Negative inner product
	case DistanceCosine:
		distanceOp = "<=>" // Cosine distance
	default:
		distanceOp = "<->" // L2 distance
	}

	// Build select columns
	selectCols := []string{req.IDColumn}
	selectCols = append(selectCols, fmt.Sprintf("%s %s $1::vector AS distance", req.VectorColumn, distanceOp))
	for _, col := range req.OutputColumns {
		if col != req.IDColumn && col != "distance" {
			selectCols = append(selectCols, col)
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectCols, ", "), req.TableName)

	args := []interface{}{vectorToString(req.QueryVector)}
	argIdx := 2

	if req.Filter != "" {
		query += fmt.Sprintf(" WHERE %s", req.Filter)
	}

	query += fmt.Sprintf(" ORDER BY %s %s $1::vector LIMIT $%d", req.VectorColumn, distanceOp, argIdx)
	args = append(args, req.Limit)

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	results := make([]SearchResult, 0)
	for rows.Next() {
		result := SearchResult{
			Metadata: make(map[string]interface{}),
		}

		// Prepare scan destinations
		scanDests := make([]interface{}, len(selectCols))
		scanDests[0] = &result.ID
		scanDests[1] = &result.Distance

		// Create placeholders for metadata columns
		metadataValues := make([]interface{}, len(req.OutputColumns))
		for i := range req.OutputColumns {
			metadataValues[i] = new(interface{})
			if i+2 < len(scanDests) {
				scanDests[i+2] = metadataValues[i]
			}
		}

		if err := rows.Scan(scanDests...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Populate metadata
		for i, col := range req.OutputColumns {
			if col != req.IDColumn && col != "distance" {
				result.Metadata[col] = *(metadataValues[i].(*interface{}))
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// DeleteRequest represents a delete request.
type DeleteRequest struct {
	TableName string
	IDColumn  string
	IDs       []string
	Filter    string
}

// Delete removes vectors from the table.
func (c *Client) Delete(ctx context.Context, req *DeleteRequest) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected")
	}

	var query string
	var args []interface{}

	if len(req.IDs) > 0 {
		placeholders := make([]string, len(req.IDs))
		for i, id := range req.IDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, id)
		}
		query = fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
			req.TableName, req.IDColumn, strings.Join(placeholders, ", "))
	} else if req.Filter != "" {
		query = fmt.Sprintf("DELETE FROM %s WHERE %s", req.TableName, req.Filter)
	} else {
		return 0, fmt.Errorf("either IDs or filter must be specified")
	}

	result, err := c.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete vectors: %w", err)
	}

	count := result.RowsAffected()
	c.logger.WithFields(logrus.Fields{
		"table":   req.TableName,
		"deleted": count,
	}).Debug("Vectors deleted")

	return count, nil
}

// Get retrieves vectors by IDs.
func (c *Client) Get(ctx context.Context, tableName, idColumn string, ids []string, outputColumns []string) ([]map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Build select columns
	selectCols := "*"
	if len(outputColumns) > 0 {
		selectCols = strings.Join(outputColumns, ", ")
	}

	// Build placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s IN (%s)",
		selectCols, tableName, idColumn, strings.Join(placeholders, ", "))

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get vectors: %w", err)
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	fieldDescriptions := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to get row values: %w", err)
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Count returns the number of vectors in a table.
func (c *Client) Count(ctx context.Context, tableName string, filter string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if filter != "" {
		query += fmt.Sprintf(" WHERE %s", filter)
	}

	var count int64
	err := c.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count vectors: %w", err)
	}

	return count, nil
}

// GetPool returns the underlying connection pool for advanced operations.
func (c *Client) GetPool() *pgxpool.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pool
}

// vectorToString converts a float32 slice to pgvector string format.
func vectorToString(v []float32) string {
	parts := make([]string, len(v))
	for i, val := range v {
		parts[i] = fmt.Sprintf("%f", val)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
