// Package adapters provides MCP server adapters.
// This file implements the MongoDB MCP server adapter.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MongoDBConfig configures the MongoDB adapter.
type MongoDBConfig struct {
	URI         string        `json:"uri"`
	Database    string        `json:"database"`
	Timeout     time.Duration `json:"timeout"`
	MaxPoolSize int           `json:"max_pool_size"`
	EnableTLS   bool          `json:"enable_tls"`
}

// DefaultMongoDBConfig returns default configuration.
func DefaultMongoDBConfig() MongoDBConfig {
	return MongoDBConfig{
		URI:         "mongodb://localhost:27017",
		Timeout:     30 * time.Second,
		MaxPoolSize: 10,
		EnableTLS:   false,
	}
}

// MongoDBAdapter implements the MongoDB MCP server.
type MongoDBAdapter struct {
	config MongoDBConfig
	client MongoDBClient
}

// MongoDBClient interface for MongoDB operations.
type MongoDBClient interface {
	ListDatabases(ctx context.Context) ([]string, error)
	ListCollections(ctx context.Context, database string) ([]string, error)
	Find(ctx context.Context, database, collection string, filter map[string]interface{}, options FindOptions) ([]map[string]interface{}, error)
	FindOne(ctx context.Context, database, collection string, filter map[string]interface{}) (map[string]interface{}, error)
	InsertOne(ctx context.Context, database, collection string, document map[string]interface{}) (string, error)
	InsertMany(ctx context.Context, database, collection string, documents []map[string]interface{}) ([]string, error)
	UpdateOne(ctx context.Context, database, collection string, filter, update map[string]interface{}) (int64, error)
	UpdateMany(ctx context.Context, database, collection string, filter, update map[string]interface{}) (int64, error)
	DeleteOne(ctx context.Context, database, collection string, filter map[string]interface{}) (int64, error)
	DeleteMany(ctx context.Context, database, collection string, filter map[string]interface{}) (int64, error)
	Aggregate(ctx context.Context, database, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error)
	CreateIndex(ctx context.Context, database, collection string, keys map[string]interface{}, options IndexOptions) (string, error)
	DropIndex(ctx context.Context, database, collection, indexName string) error
	Count(ctx context.Context, database, collection string, filter map[string]interface{}) (int64, error)
}

// FindOptions represents options for find operations.
type FindOptions struct {
	Limit      int64                  `json:"limit,omitempty"`
	Skip       int64                  `json:"skip,omitempty"`
	Sort       map[string]interface{} `json:"sort,omitempty"`
	Projection map[string]interface{} `json:"projection,omitempty"`
}

// IndexOptions represents options for index creation.
type IndexOptions struct {
	Unique     bool   `json:"unique,omitempty"`
	Background bool   `json:"background,omitempty"`
	Name       string `json:"name,omitempty"`
}

// NewMongoDBAdapter creates a new MongoDB adapter.
func NewMongoDBAdapter(config MongoDBConfig, client MongoDBClient) *MongoDBAdapter {
	return &MongoDBAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *MongoDBAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "mongodb",
		Version:     "1.0.0",
		Description: "MongoDB database operations including CRUD, aggregation, and index management",
		Capabilities: []string{
			"crud",
			"aggregation",
			"indexing",
			"transactions",
		},
	}
}

// ListTools returns available tools.
func (a *MongoDBAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "mongodb_list_databases",
			Description: "List all databases",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "mongodb_list_collections",
			Description: "List collections in a database",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
				},
				"required": []string{"database"},
			},
		},
		{
			Name:        "mongodb_find",
			Description: "Find documents in a collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Query filter",
						"default":     map[string]interface{}{},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of documents",
						"default":     100,
					},
					"skip": map[string]interface{}{
						"type":        "integer",
						"description": "Number of documents to skip",
						"default":     0,
					},
					"sort": map[string]interface{}{
						"type":        "object",
						"description": "Sort specification",
					},
				},
				"required": []string{"database", "collection"},
			},
		},
		{
			Name:        "mongodb_find_one",
			Description: "Find a single document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Query filter",
					},
				},
				"required": []string{"database", "collection", "filter"},
			},
		},
		{
			Name:        "mongodb_insert_one",
			Description: "Insert a single document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"document": map[string]interface{}{
						"type":        "object",
						"description": "Document to insert",
					},
				},
				"required": []string{"database", "collection", "document"},
			},
		},
		{
			Name:        "mongodb_insert_many",
			Description: "Insert multiple documents",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"documents": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "object"},
						"description": "Documents to insert",
					},
				},
				"required": []string{"database", "collection", "documents"},
			},
		},
		{
			Name:        "mongodb_update_one",
			Description: "Update a single document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Query filter",
					},
					"update": map[string]interface{}{
						"type":        "object",
						"description": "Update operations",
					},
				},
				"required": []string{"database", "collection", "filter", "update"},
			},
		},
		{
			Name:        "mongodb_delete_one",
			Description: "Delete a single document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Query filter",
					},
				},
				"required": []string{"database", "collection", "filter"},
			},
		},
		{
			Name:        "mongodb_aggregate",
			Description: "Run an aggregation pipeline",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"pipeline": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "object"},
						"description": "Aggregation pipeline stages",
					},
				},
				"required": []string{"database", "collection", "pipeline"},
			},
		},
		{
			Name:        "mongodb_count",
			Description: "Count documents matching a filter",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Query filter",
						"default":     map[string]interface{}{},
					},
				},
				"required": []string{"database", "collection"},
			},
		},
		{
			Name:        "mongodb_create_index",
			Description: "Create an index on a collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"keys": map[string]interface{}{
						"type":        "object",
						"description": "Index keys (field: 1 for asc, -1 for desc)",
					},
					"unique": map[string]interface{}{
						"type":        "boolean",
						"description": "Create unique index",
						"default":     false,
					},
				},
				"required": []string{"database", "collection", "keys"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *MongoDBAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "mongodb_list_databases":
		return a.listDatabases(ctx)
	case "mongodb_list_collections":
		return a.listCollections(ctx, args)
	case "mongodb_find":
		return a.find(ctx, args)
	case "mongodb_find_one":
		return a.findOne(ctx, args)
	case "mongodb_insert_one":
		return a.insertOne(ctx, args)
	case "mongodb_insert_many":
		return a.insertMany(ctx, args)
	case "mongodb_update_one":
		return a.updateOne(ctx, args)
	case "mongodb_delete_one":
		return a.deleteOne(ctx, args)
	case "mongodb_aggregate":
		return a.aggregate(ctx, args)
	case "mongodb_count":
		return a.count(ctx, args)
	case "mongodb_create_index":
		return a.createIndex(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *MongoDBAdapter) listDatabases(ctx context.Context) (*ToolResult, error) {
	databases, err := a.client.ListDatabases(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d databases:\n\n", len(databases)))
	for _, db := range databases {
		sb.WriteString(fmt.Sprintf("- %s\n", db))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *MongoDBAdapter) listCollections(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)

	collections, err := a.client.ListCollections(ctx, database)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Collections in '%s' (%d):\n\n", database, len(collections)))
	for _, coll := range collections {
		sb.WriteString(fmt.Sprintf("- %s\n", coll))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *MongoDBAdapter) find(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	filter, _ := args["filter"].(map[string]interface{})
	if filter == nil {
		filter = map[string]interface{}{}
	}

	options := FindOptions{
		Limit: int64(getIntArg(args, "limit", 100)),
		Skip:  int64(getIntArg(args, "skip", 0)),
	}
	if sort, ok := args["sort"].(map[string]interface{}); ok {
		options.Sort = sort
	}

	docs, err := a.client.Find(ctx, database, collection, filter, options)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(docs, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Found %d documents:\n\n%s", len(docs), string(data))}},
	}, nil
}

func (a *MongoDBAdapter) findOne(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	filter, _ := args["filter"].(map[string]interface{})

	doc, err := a.client.FindOne(ctx, database, collection, filter)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(doc, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(data)}},
	}, nil
}

func (a *MongoDBAdapter) insertOne(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	document, _ := args["document"].(map[string]interface{})

	id, err := a.client.InsertOne(ctx, database, collection, document)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Inserted document with ID: %s", id)}},
	}, nil
}

func (a *MongoDBAdapter) insertMany(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	docsRaw, _ := args["documents"].([]interface{})

	var documents []map[string]interface{}
	for _, d := range docsRaw {
		if doc, ok := d.(map[string]interface{}); ok {
			documents = append(documents, doc)
		}
	}

	ids, err := a.client.InsertMany(ctx, database, collection, documents)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Inserted %d documents with IDs: %v", len(ids), ids)}},
	}, nil
}

func (a *MongoDBAdapter) updateOne(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	filter, _ := args["filter"].(map[string]interface{})
	update, _ := args["update"].(map[string]interface{})

	matched, err := a.client.UpdateOne(ctx, database, collection, filter, update)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Updated %d document(s)", matched)}},
	}, nil
}

func (a *MongoDBAdapter) deleteOne(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	filter, _ := args["filter"].(map[string]interface{})

	deleted, err := a.client.DeleteOne(ctx, database, collection, filter)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Deleted %d document(s)", deleted)}},
	}, nil
}

func (a *MongoDBAdapter) aggregate(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	pipelineRaw, _ := args["pipeline"].([]interface{})

	var pipeline []map[string]interface{}
	for _, p := range pipelineRaw {
		if stage, ok := p.(map[string]interface{}); ok {
			pipeline = append(pipeline, stage)
		}
	}

	results, err := a.client.Aggregate(ctx, database, collection, pipeline)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Aggregation results (%d):\n\n%s", len(results), string(data))}},
	}, nil
}

func (a *MongoDBAdapter) count(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	filter, _ := args["filter"].(map[string]interface{})
	if filter == nil {
		filter = map[string]interface{}{}
	}

	count, err := a.client.Count(ctx, database, collection, filter)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Count: %d documents", count)}},
	}, nil
}

func (a *MongoDBAdapter) createIndex(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	database, _ := args["database"].(string)
	collection, _ := args["collection"].(string)
	keys, _ := args["keys"].(map[string]interface{})
	unique, _ := args["unique"].(bool)

	options := IndexOptions{
		Unique: unique,
	}

	indexName, err := a.client.CreateIndex(ctx, database, collection, keys, options)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created index: %s", indexName)}},
	}, nil
}
