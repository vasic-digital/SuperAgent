// Package adapters provides MCP server adapters.
// This file implements the Notion MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// NotionConfig configures the Notion adapter.
type NotionConfig struct {
	APIKey  string        `json:"api_key"`
	Timeout time.Duration `json:"timeout"`
}

// NotionAdapter implements the Notion MCP server.
type NotionAdapter struct {
	config NotionConfig
	client NotionClient
}

// NotionClient interface for Notion operations.
type NotionClient interface {
	SearchPages(ctx context.Context, query string, filter NotionFilter) ([]NotionPage, error)
	GetPage(ctx context.Context, pageID string) (*NotionPage, error)
	CreatePage(ctx context.Context, parentID string, properties map[string]interface{}, content []NotionBlock) (*NotionPage, error)
	UpdatePage(ctx context.Context, pageID string, properties map[string]interface{}) (*NotionPage, error)
	DeletePage(ctx context.Context, pageID string) error
	GetDatabase(ctx context.Context, databaseID string) (*NotionDatabase, error)
	QueryDatabase(ctx context.Context, databaseID string, filter, sorts interface{}) ([]NotionPage, error)
	AppendBlocks(ctx context.Context, pageID string, blocks []NotionBlock) error
	GetBlocks(ctx context.Context, pageID string) ([]NotionBlock, error)
	CreateComment(ctx context.Context, pageID, content string) error
	GetComments(ctx context.Context, pageID string) ([]NotionComment, error)
}

// NotionFilter represents search filters.
type NotionFilter struct {
	Property string      `json:"property,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// NotionPage represents a Notion page.
type NotionPage struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	URL            string                 `json:"url"`
	ParentID       string                 `json:"parent_id"`
	ParentType     string                 `json:"parent_type"`
	Properties     map[string]interface{} `json:"properties"`
	CreatedTime    time.Time              `json:"created_time"`
	LastEditedTime time.Time              `json:"last_edited_time"`
	Archived       bool                   `json:"archived"`
}

// NotionDatabase represents a Notion database.
type NotionDatabase struct {
	ID          string                       `json:"id"`
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	Properties  map[string]NotionPropertyDef `json:"properties"`
	URL         string                       `json:"url"`
}

// NotionPropertyDef represents a database property definition.
type NotionPropertyDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// NotionBlock represents a content block.
type NotionBlock struct {
	ID       string      `json:"id,omitempty"`
	Type     string      `json:"type"`
	Content  interface{} `json:"content"`
	Children []NotionBlock `json:"children,omitempty"`
}

// NotionComment represents a comment.
type NotionComment struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	CreatedBy   string    `json:"created_by"`
	CreatedTime time.Time `json:"created_time"`
}

// NewNotionAdapter creates a new Notion adapter.
func NewNotionAdapter(config NotionConfig, client NotionClient) *NotionAdapter {
	return &NotionAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *NotionAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "notion",
		Version:     "1.0.0",
		Description: "Notion workspace integration for pages, databases, and content blocks",
		Capabilities: []string{
			"pages",
			"databases",
			"blocks",
			"search",
			"comments",
		},
	}
}

// ListTools returns available tools.
func (a *NotionAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "notion_search",
			Description: "Search for pages in Notion",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "notion_get_page",
			Description: "Get a page by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "Page ID",
					},
				},
				"required": []string{"page_id"},
			},
		},
		{
			Name:        "notion_create_page",
			Description: "Create a new page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "Parent page or database ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Page title",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Page content (markdown)",
					},
				},
				"required": []string{"parent_id", "title"},
			},
		},
		{
			Name:        "notion_update_page",
			Description: "Update page properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "Page ID",
					},
					"properties": map[string]interface{}{
						"type":        "object",
						"description": "Properties to update",
					},
				},
				"required": []string{"page_id", "properties"},
			},
		},
		{
			Name:        "notion_delete_page",
			Description: "Archive/delete a page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "Page ID",
					},
				},
				"required": []string{"page_id"},
			},
		},
		{
			Name:        "notion_get_database",
			Description: "Get database schema",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database_id": map[string]interface{}{
						"type":        "string",
						"description": "Database ID",
					},
				},
				"required": []string{"database_id"},
			},
		},
		{
			Name:        "notion_query_database",
			Description: "Query a database",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database_id": map[string]interface{}{
						"type":        "string",
						"description": "Database ID",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Filter conditions",
					},
					"sorts": map[string]interface{}{
						"type":        "array",
						"description": "Sort conditions",
					},
				},
				"required": []string{"database_id"},
			},
		},
		{
			Name:        "notion_get_blocks",
			Description: "Get page content blocks",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "Page ID",
					},
				},
				"required": []string{"page_id"},
			},
		},
		{
			Name:        "notion_append_blocks",
			Description: "Append content to a page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "Page ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to append (markdown)",
					},
				},
				"required": []string{"page_id", "content"},
			},
		},
		{
			Name:        "notion_add_comment",
			Description: "Add a comment to a page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "Page ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Comment content",
					},
				},
				"required": []string{"page_id", "content"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *NotionAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "notion_search":
		return a.search(ctx, args)
	case "notion_get_page":
		return a.getPage(ctx, args)
	case "notion_create_page":
		return a.createPage(ctx, args)
	case "notion_update_page":
		return a.updatePage(ctx, args)
	case "notion_delete_page":
		return a.deletePage(ctx, args)
	case "notion_get_database":
		return a.getDatabase(ctx, args)
	case "notion_query_database":
		return a.queryDatabase(ctx, args)
	case "notion_get_blocks":
		return a.getBlocks(ctx, args)
	case "notion_append_blocks":
		return a.appendBlocks(ctx, args)
	case "notion_add_comment":
		return a.addComment(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *NotionAdapter) search(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)

	pages, err := a.client.SearchPages(ctx, query, NotionFilter{})
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results for '%s':\n\n", len(pages), query))

	for _, p := range pages {
		sb.WriteString(fmt.Sprintf("- %s\n", p.Title))
		sb.WriteString(fmt.Sprintf("  ID: %s\n", p.ID))
		sb.WriteString(fmt.Sprintf("  URL: %s\n", p.URL))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *NotionAdapter) getPage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pageID, _ := args["page_id"].(string)

	page, err := a.client.GetPage(ctx, pageID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Page: %s\n", page.Title))
	sb.WriteString(fmt.Sprintf("ID: %s\n", page.ID))
	sb.WriteString(fmt.Sprintf("URL: %s\n", page.URL))
	sb.WriteString(fmt.Sprintf("Created: %s\n", page.CreatedTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Last edited: %s\n", page.LastEditedTime.Format(time.RFC3339)))

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *NotionAdapter) createPage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	parentID, _ := args["parent_id"].(string)
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)

	properties := map[string]interface{}{
		"title": title,
	}

	var blocks []NotionBlock
	if content != "" {
		blocks = []NotionBlock{{
			Type:    "paragraph",
			Content: content,
		}}
	}

	page, err := a.client.CreatePage(ctx, parentID, properties, blocks)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created page: %s (ID: %s)", page.Title, page.ID)}},
	}, nil
}

func (a *NotionAdapter) updatePage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pageID, _ := args["page_id"].(string)
	properties, _ := args["properties"].(map[string]interface{})

	page, err := a.client.UpdatePage(ctx, pageID, properties)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Updated page: %s", page.Title)}},
	}, nil
}

func (a *NotionAdapter) deletePage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pageID, _ := args["page_id"].(string)

	err := a.client.DeletePage(ctx, pageID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Archived page: %s", pageID)}},
	}, nil
}

func (a *NotionAdapter) getDatabase(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	databaseID, _ := args["database_id"].(string)

	db, err := a.client.GetDatabase(ctx, databaseID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Database: %s\n", db.Title))
	sb.WriteString(fmt.Sprintf("Description: %s\n", db.Description))
	sb.WriteString("\nProperties:\n")
	for name, prop := range db.Properties {
		sb.WriteString(fmt.Sprintf("  - %s (%s)\n", name, prop.Type))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *NotionAdapter) queryDatabase(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	databaseID, _ := args["database_id"].(string)
	filter := args["filter"]
	sorts := args["sorts"]

	pages, err := a.client.QueryDatabase(ctx, databaseID, filter, sorts)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Query results (%d):\n\n", len(pages)))

	for _, p := range pages {
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)\n", p.Title, p.ID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *NotionAdapter) getBlocks(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pageID, _ := args["page_id"].(string)

	blocks, err := a.client.GetBlocks(ctx, pageID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Page content (%d blocks):\n\n", len(blocks)))

	for _, b := range blocks {
		sb.WriteString(fmt.Sprintf("[%s] %v\n", b.Type, b.Content))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *NotionAdapter) appendBlocks(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pageID, _ := args["page_id"].(string)
	content, _ := args["content"].(string)

	blocks := []NotionBlock{{
		Type:    "paragraph",
		Content: content,
	}}

	err := a.client.AppendBlocks(ctx, pageID, blocks)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Content appended successfully"}},
	}, nil
}

func (a *NotionAdapter) addComment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pageID, _ := args["page_id"].(string)
	content, _ := args["content"].(string)

	err := a.client.CreateComment(ctx, pageID, content)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Comment added successfully"}},
	}, nil
}
