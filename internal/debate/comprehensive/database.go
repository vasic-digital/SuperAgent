package comprehensive

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DatabaseTool executes SQL queries
// Note: This is a stub implementation. In production, import the appropriate driver:
// _ "github.com/lib/pq" for PostgreSQL
// _ "github.com/go-sql-driver/mysql" for MySQL
// _ "github.com/mattn/go-sqlite3" for SQLite

type DatabaseTool struct {
	db       *sql.DB
	logger   *logrus.Logger
	readonly bool
}

// NewDatabaseTool creates a new database tool
func NewDatabaseTool(db *sql.DB, readonly bool, logger *logrus.Logger) *DatabaseTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &DatabaseTool{
		db:       db,
		logger:   logger,
		readonly: readonly,
	}
}

// GetName returns the tool name
func (t *DatabaseTool) GetName() string {
	return "query_database"
}

// GetType returns the tool type
func (t *DatabaseTool) GetType() ToolType {
	return ToolTypeDatabase
}

// GetDescription returns the description
func (t *DatabaseTool) GetDescription() string {
	return "Execute SQL queries against database"
}

// GetInputSchema returns the input schema
func (t *DatabaseTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "SQL query to execute",
			},
			"params": map[string]interface{}{
				"type":        "array",
				"description": "Query parameters",
			},
		},
		"required": []string{"query"},
	}
}

// Validate validates inputs
func (t *DatabaseTool) Validate(inputs map[string]interface{}) error {
	query, ok := inputs["query"].(string)
	if !ok || query == "" {
		return fmt.Errorf("query is required")
	}

	// Check for dangerous operations in readonly mode
	if t.readonly {
		dangerous := []string{"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE"}
		upperQuery := strings.ToUpper(strings.TrimSpace(query))
		for _, d := range dangerous {
			if strings.HasPrefix(upperQuery, d) {
				return fmt.Errorf("write operations not allowed in readonly mode: %s", d)
			}
		}
	}

	return nil
}

// Execute executes the tool
func (t *DatabaseTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	if t.db == nil {
		return NewToolError("database connection not available"), nil
	}

	//nolint:errcheck // schema validation ensures correct type
	query := inputs["query"].(string)

	var params []interface{}
	if p, ok := inputs["params"].([]interface{}); ok {
		params = p
	}

	t.logger.WithField("query", query).Debug("Executing database query")

	start := time.Now()

	// Check if it's a SELECT query
	isSelect := strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT")

	if isSelect {
		return t.executeQuery(ctx, query, params, start)
	}

	return t.executeExec(ctx, query, params, start)
}

// executeQuery executes a SELECT query
func (t *DatabaseTool) executeQuery(ctx context.Context, query string, params []interface{}, start time.Time) (*ToolResult, error) {
	rows, err := t.db.QueryContext(ctx, query, params...)
	if err != nil {
		return NewToolError(fmt.Sprintf("query failed: %v", err)), nil
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return NewToolError(fmt.Sprintf("failed to get columns: %v", err)), nil
	}

	// Fetch all rows
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return NewToolError(fmt.Sprintf("failed to scan row: %v", err)), nil
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return NewToolError(fmt.Sprintf("row iteration error: %v", err)), nil
	}

	result := NewToolResult(fmt.Sprintf("Query returned %d rows", len(results)))
	result.Duration = time.Since(start)
	result.Data["columns"] = columns
	result.Data["rows"] = results
	result.Data["count"] = len(results)

	return result, nil
}

// executeExec executes INSERT, UPDATE, DELETE, etc.
func (t *DatabaseTool) executeExec(ctx context.Context, query string, params []interface{}, start time.Time) (*ToolResult, error) {
	res, err := t.db.ExecContext(ctx, query, params...)
	if err != nil {
		return NewToolError(fmt.Sprintf("execution failed: %v", err)), nil
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		rowsAffected = -1
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		lastID = -1
	}

	result := NewToolResult(fmt.Sprintf("Query executed successfully, %d rows affected", rowsAffected))
	result.Duration = time.Since(start)
	result.Data["rows_affected"] = rowsAffected
	result.Data["last_insert_id"] = lastID

	return result, nil
}

// Close closes the database connection
func (t *DatabaseTool) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}
