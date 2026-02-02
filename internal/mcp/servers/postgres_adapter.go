// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

// PostgresAdapterConfig holds configuration for PostgreSQL MCP adapter
type PostgresAdapterConfig struct {
	// Host is the database host
	Host string `json:"host,omitempty"`
	// Port is the database port
	Port int `json:"port,omitempty"`
	// User is the database user
	User string `json:"user,omitempty"`
	// Password is the database password
	Password string `json:"password,omitempty"`
	// Database is the database name
	Database string `json:"database,omitempty"`
	// SSLMode is the SSL mode (disable, require, verify-ca, verify-full)
	SSLMode string `json:"ssl_mode,omitempty"`
	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int `json:"max_open_conns,omitempty"`
	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int `json:"max_idle_conns,omitempty"`
	// ConnMaxLifetime is the maximum connection lifetime
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime,omitempty"`
	// QueryTimeout is the default query timeout
	QueryTimeout time.Duration `json:"query_timeout,omitempty"`
	// ReadOnly restricts operations to read-only queries
	ReadOnly bool `json:"read_only"`
	// AllowedSchemas restricts operations to specific schemas
	AllowedSchemas []string `json:"allowed_schemas,omitempty"`
}

// DefaultPostgresAdapterConfig returns default configuration
func DefaultPostgresAdapterConfig() PostgresAdapterConfig {
	return PostgresAdapterConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Database:        "postgres",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		QueryTimeout:    30 * time.Second,
		ReadOnly:        false,
		AllowedSchemas:  []string{"public"},
	}
}

// PostgresAdapter implements MCP adapter for PostgreSQL operations
type PostgresAdapter struct {
	config      PostgresAdapterConfig
	db          *sql.DB
	initialized bool
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// NewPostgresAdapter creates a new PostgreSQL MCP adapter
func NewPostgresAdapter(config PostgresAdapterConfig, logger *logrus.Logger) *PostgresAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = 10
	}
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = 30 * time.Minute
	}
	if config.QueryTimeout <= 0 {
		config.QueryTimeout = 30 * time.Second
	}

	return &PostgresAdapter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the PostgreSQL adapter
func (p *PostgresAdapter) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.Database,
		p.config.SSLMode,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(p.config.MaxOpenConns)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	p.initialized = true
	p.logger.WithFields(logrus.Fields{
		"host":     p.config.Host,
		"port":     p.config.Port,
		"database": p.config.Database,
	}).Info("PostgreSQL adapter initialized")

	return nil
}

// Health returns health status
func (p *PostgresAdapter) Health(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized || p.db == nil {
		return fmt.Errorf("PostgreSQL adapter not initialized")
	}

	return p.db.PingContext(ctx)
}

// Close closes the adapter and database connection
func (p *PostgresAdapter) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.db != nil {
		if err := p.db.Close(); err != nil {
			return err
		}
		p.db = nil
	}

	p.initialized = false
	return nil
}

// isSchemaAllowed checks if a schema is allowed
func (p *PostgresAdapter) isSchemaAllowed(schema string) bool {
	if len(p.config.AllowedSchemas) == 0 {
		return true
	}
	for _, allowed := range p.config.AllowedSchemas {
		if strings.EqualFold(allowed, schema) {
			return true
		}
	}
	return false
}

// isReadOnlyQuery checks if a query is read-only
func (p *PostgresAdapter) isReadOnlyQuery(query string) bool {
	query = strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(query, "SELECT") ||
		strings.HasPrefix(query, "WITH") ||
		strings.HasPrefix(query, "EXPLAIN") ||
		strings.HasPrefix(query, "SHOW")
}

// QueryResult represents the result of a query
type QueryResult struct {
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"row_count"`
	AffectedRows int64                    `json:"affected_rows,omitempty"`
	Duration     time.Duration            `json:"duration"`
}

// Query executes a SQL query and returns results
func (p *PostgresAdapter) Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	// Check read-only restriction
	if p.config.ReadOnly && !p.isReadOnlyQuery(query) {
		return nil, fmt.Errorf("write operations not allowed in read-only mode")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.QueryTimeout)
	defer cancel()

	start := time.Now()

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    []map[string]interface{}{},
	}

	// Create scan destinations
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert byte slices to strings for JSON compatibility
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	result.RowCount = len(result.Rows)
	result.Duration = time.Since(start)

	return result, nil
}

// Execute executes a SQL statement (INSERT, UPDATE, DELETE)
func (p *PostgresAdapter) Execute(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if p.config.ReadOnly {
		return nil, fmt.Errorf("write operations not allowed in read-only mode")
	}

	ctx, cancel := context.WithTimeout(ctx, p.config.QueryTimeout)
	defer cancel()

	start := time.Now()

	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute failed: %w", err)
	}

	affected, _ := result.RowsAffected()

	return &QueryResult{
		AffectedRows: affected,
		Duration:     time.Since(start),
	}, nil
}

// TableInfo represents information about a table
type TableInfo struct {
	Schema      string       `json:"schema"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Owner       string       `json:"owner,omitempty"`
	RowEstimate int64        `json:"row_estimate,omitempty"`
	Columns     []ColumnInfo `json:"columns,omitempty"`
}

// ColumnInfo represents information about a column
type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	IsNullable   bool   `json:"is_nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key"`
}

// ListTables returns a list of tables in the database
func (p *PostgresAdapter) ListTables(ctx context.Context, schema string) ([]TableInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if schema == "" {
		schema = "public"
	}

	// Check schema permissions first (fail fast for security)
	if !p.isSchemaAllowed(schema) {
		return nil, fmt.Errorf("schema not allowed: %s", schema)
	}

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	query := `
		SELECT
			schemaname,
			tablename,
			tableowner,
			n_live_tup
		FROM pg_stat_user_tables
		WHERE schemaname = $1
		ORDER BY tablename
	`

	rows, err := p.db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		t.Type = "table"
		if err := rows.Scan(&t.Schema, &t.Name, &t.Owner, &t.RowEstimate); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, t)
	}

	return tables, nil
}

// DescribeTable returns detailed information about a table
func (p *PostgresAdapter) DescribeTable(ctx context.Context, schema, table string) (*TableInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if schema == "" {
		schema = "public"
	}

	// Check schema permissions first (fail fast for security)
	if !p.isSchemaAllowed(schema) {
		return nil, fmt.Errorf("schema not allowed: %s", schema)
	}

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	info := &TableInfo{
		Schema: schema,
		Name:   table,
		Type:   "table",
	}

	// Get columns
	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES',
			COALESCE(c.column_default, ''),
			COALESCE(kcu.column_name IS NOT NULL, false) as is_pk
		FROM information_schema.columns c
		LEFT JOIN information_schema.table_constraints tc
			ON tc.table_schema = c.table_schema
			AND tc.table_name = c.table_name
			AND tc.constraint_type = 'PRIMARY KEY'
		LEFT JOIN information_schema.key_column_usage kcu
			ON kcu.table_schema = c.table_schema
			AND kcu.table_name = c.table_name
			AND kcu.column_name = c.column_name
			AND kcu.constraint_name = tc.constraint_name
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := p.db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable, &col.DefaultValue, &col.IsPrimaryKey); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		info.Columns = append(info.Columns, col)
	}

	if len(info.Columns) == 0 {
		return nil, fmt.Errorf("table not found: %s.%s", schema, table)
	}

	return info, nil
}

// ListSchemas returns a list of schemas in the database
func (p *PostgresAdapter) ListSchemas(ctx context.Context) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT LIKE 'pg_%'
		AND schema_name != 'information_schema'
		ORDER BY schema_name
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// IndexInfo represents information about an index
type IndexInfo struct {
	Name       string   `json:"name"`
	Table      string   `json:"table"`
	Columns    []string `json:"columns"`
	IsUnique   bool     `json:"is_unique"`
	IsPrimary  bool     `json:"is_primary"`
	Definition string   `json:"definition,omitempty"`
}

// ListIndexes returns indexes for a table
func (p *PostgresAdapter) ListIndexes(ctx context.Context, schema, table string) ([]IndexInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT
			i.relname as index_name,
			t.relname as table_name,
			array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)) as columns,
			ix.indisunique,
			ix.indisprimary,
			pg_get_indexdef(ix.indexrelid) as definition
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE n.nspname = $1 AND t.relname = $2
		GROUP BY i.relname, t.relname, ix.indisunique, ix.indisprimary, ix.indexrelid
		ORDER BY i.relname
	`

	rows, err := p.db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var columns string
		if err := rows.Scan(&idx.Name, &idx.Table, &columns, &idx.IsUnique, &idx.IsPrimary, &idx.Definition); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		// Parse array format {col1,col2}
		columns = strings.Trim(columns, "{}")
		if columns != "" {
			idx.Columns = strings.Split(columns, ",")
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	DatabaseName   string    `json:"database_name"`
	Size           string    `json:"size"`
	Connections    int       `json:"connections"`
	ActiveQueries  int       `json:"active_queries"`
	TransactionID  int64     `json:"transaction_id"`
	LastVacuum     time.Time `json:"last_vacuum,omitempty"`
	LastAutoVacuum time.Time `json:"last_auto_vacuum,omitempty"`
}

// GetStats returns database statistics
func (p *PostgresAdapter) GetStats(ctx context.Context) (*DatabaseStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized || p.db == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	stats := &DatabaseStats{
		DatabaseName: p.config.Database,
	}

	// Get database size
	row := p.db.QueryRowContext(ctx, "SELECT pg_size_pretty(pg_database_size(current_database()))")
	_ = row.Scan(&stats.Size)

	// Get connection count
	_ = p.db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database()").Scan(&stats.Connections)

	// Get active queries
	_ = p.db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND state = 'active'").Scan(&stats.ActiveQueries)

	return stats, nil
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (p *PostgresAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "postgres_query",
			Description: "Execute a SQL query and return results",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL query to execute",
					},
					"params": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Query parameters",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "postgres_execute",
			Description: "Execute a SQL statement (INSERT, UPDATE, DELETE)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL statement to execute",
					},
					"params": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Query parameters",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "postgres_list_tables",
			Description: "List all tables in a schema",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"schema": map[string]interface{}{
						"type":        "string",
						"description": "Schema name (default: public)",
						"default":     "public",
					},
				},
			},
		},
		{
			Name:        "postgres_describe_table",
			Description: "Get detailed information about a table",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"table": map[string]interface{}{
						"type":        "string",
						"description": "Table name",
					},
					"schema": map[string]interface{}{
						"type":        "string",
						"description": "Schema name (default: public)",
						"default":     "public",
					},
				},
				"required": []string{"table"},
			},
		},
		{
			Name:        "postgres_list_schemas",
			Description: "List all schemas in the database",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "postgres_list_indexes",
			Description: "List indexes for a table",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"table": map[string]interface{}{
						"type":        "string",
						"description": "Table name",
					},
					"schema": map[string]interface{}{
						"type":        "string",
						"description": "Schema name (default: public)",
						"default":     "public",
					},
				},
				"required": []string{"table"},
			},
		},
		{
			Name:        "postgres_stats",
			Description: "Get database statistics",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (p *PostgresAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	p.mu.RLock()
	initialized := p.initialized
	p.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	switch toolName {
	case "postgres_query":
		query, _ := params["query"].(string)
		var args []interface{}
		if p, ok := params["params"].([]interface{}); ok {
			args = p
		}
		return p.Query(ctx, query, args...)

	case "postgres_execute":
		query, _ := params["query"].(string)
		var args []interface{}
		if p, ok := params["params"].([]interface{}); ok {
			args = p
		}
		return p.Execute(ctx, query, args...)

	case "postgres_list_tables":
		schema, _ := params["schema"].(string)
		return p.ListTables(ctx, schema)

	case "postgres_describe_table":
		table, _ := params["table"].(string)
		schema, _ := params["schema"].(string)
		return p.DescribeTable(ctx, schema, table)

	case "postgres_list_schemas":
		return p.ListSchemas(ctx)

	case "postgres_list_indexes":
		table, _ := params["table"].(string)
		schema, _ := params["schema"].(string)
		return p.ListIndexes(ctx, schema, table)

	case "postgres_stats":
		return p.GetStats(ctx)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (p *PostgresAdapter) GetCapabilities() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"name":            "postgres",
		"host":            p.config.Host,
		"port":            p.config.Port,
		"database":        p.config.Database,
		"read_only":       p.config.ReadOnly,
		"allowed_schemas": p.config.AllowedSchemas,
		"tools":           len(p.GetMCPTools()),
		"initialized":     p.initialized,
	}
}

// MarshalJSON implements custom JSON marshaling
func (p *PostgresAdapter) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"initialized":  p.initialized,
		"capabilities": p.GetCapabilities(),
	})
}
