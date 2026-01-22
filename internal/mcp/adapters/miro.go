// Package adapters provides MCP server adapters.
// This file implements the Miro MCP server adapter for whiteboard collaboration.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MiroConfig configures the Miro adapter.
type MiroConfig struct {
	AccessToken string        `json:"access_token"`
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
}

// DefaultMiroConfig returns default configuration.
func DefaultMiroConfig() MiroConfig {
	return MiroConfig{
		BaseURL: "https://api.miro.com/v2",
		Timeout: 30 * time.Second,
	}
}

// MiroAdapter implements the Miro MCP server.
type MiroAdapter struct {
	config     MiroConfig
	httpClient *http.Client
}

// NewMiroAdapter creates a new Miro adapter.
func NewMiroAdapter(config MiroConfig) *MiroAdapter {
	if config.BaseURL == "" {
		config.BaseURL = DefaultMiroConfig().BaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultMiroConfig().Timeout
	}
	return &MiroAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *MiroAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "miro",
		Version:     "1.0.0",
		Description: "Miro whiteboard collaboration integration",
		Capabilities: []string{
			"list_boards",
			"create_board",
			"get_board",
			"create_sticky_note",
			"create_shape",
			"create_text",
			"create_connector",
			"list_items",
			"get_item",
			"delete_item",
			"export_board",
		},
	}
}

// ListTools returns available tools.
func (a *MiroAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "miro_list_boards",
			Description: "List all accessible boards",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for board names",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of boards to return",
						"default":     50,
					},
				},
			},
		},
		{
			Name:        "miro_create_board",
			Description: "Create a new Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Board name",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Board description",
					},
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID to create board in",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "miro_get_board",
			Description: "Get board details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "miro_create_sticky_note",
			Description: "Create a sticky note on a board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text content of the sticky note",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position on the board",
						"default":     0,
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position on the board",
						"default":     0,
					},
					"color": map[string]interface{}{
						"type":        "string",
						"description": "Sticky note color",
						"enum":        []string{"gray", "light_yellow", "yellow", "orange", "light_green", "green", "dark_green", "cyan", "light_pink", "pink", "violet", "red", "light_blue", "blue", "dark_blue", "black"},
						"default":     "light_yellow",
					},
					"width": map[string]interface{}{
						"type":        "number",
						"description": "Width of the sticky note",
						"default":     228,
					},
				},
				"required": []string{"board_id", "content"},
			},
		},
		{
			Name:        "miro_create_shape",
			Description: "Create a shape on a board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"shape": map[string]interface{}{
						"type":        "string",
						"description": "Shape type",
						"enum":        []string{"rectangle", "round_rectangle", "circle", "triangle", "rhombus", "parallelogram", "trapezoid", "pentagon", "hexagon", "octagon", "wedge_round_rectangle_callout", "star", "flow_chart_predefined_process", "cloud", "cross", "can", "right_arrow", "left_arrow", "left_right_arrow", "left_brace", "right_brace"},
						"default":     "rectangle",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text content inside the shape",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
						"default":     0,
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
						"default":     0,
					},
					"width": map[string]interface{}{
						"type":        "number",
						"description": "Width",
						"default":     200,
					},
					"height": map[string]interface{}{
						"type":        "number",
						"description": "Height",
						"default":     200,
					},
					"fill_color": map[string]interface{}{
						"type":        "string",
						"description": "Fill color (hex)",
						"default":     "#ffffff",
					},
					"border_color": map[string]interface{}{
						"type":        "string",
						"description": "Border color (hex)",
						"default":     "#000000",
					},
				},
				"required": []string{"board_id", "shape"},
			},
		},
		{
			Name:        "miro_create_text",
			Description: "Create a text item on a board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text content (supports basic HTML)",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
						"default":     0,
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
						"default":     0,
					},
					"width": map[string]interface{}{
						"type":        "number",
						"description": "Width of text box",
						"default":     200,
					},
					"font_size": map[string]interface{}{
						"type":        "integer",
						"description": "Font size",
						"default":     14,
					},
				},
				"required": []string{"board_id", "content"},
			},
		},
		{
			Name:        "miro_create_connector",
			Description: "Create a connector between two items",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"start_item_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the starting item",
					},
					"end_item_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the ending item",
					},
					"style": map[string]interface{}{
						"type":        "string",
						"description": "Connector line style",
						"enum":        []string{"straight", "elbowed", "curved"},
						"default":     "elbowed",
					},
					"start_cap": map[string]interface{}{
						"type":        "string",
						"description": "Start cap style",
						"enum":        []string{"none", "stealth", "arrow", "filled_triangle", "triangle", "filled_diamond", "diamond", "filled_oval", "oval", "erd_one", "erd_many", "erd_one_or_many", "erd_only_one", "erd_zero_or_many", "erd_zero_or_one"},
						"default":     "none",
					},
					"end_cap": map[string]interface{}{
						"type":        "string",
						"description": "End cap style",
						"enum":        []string{"none", "stealth", "arrow", "filled_triangle", "triangle", "filled_diamond", "diamond", "filled_oval", "oval", "erd_one", "erd_many", "erd_one_or_many", "erd_only_one", "erd_zero_or_many", "erd_zero_or_one"},
						"default":     "stealth",
					},
				},
				"required": []string{"board_id", "start_item_id", "end_item_id"},
			},
		},
		{
			Name:        "miro_list_items",
			Description: "List items on a board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by item type",
						"enum":        []string{"sticky_note", "shape", "text", "connector", "card", "frame", "image"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of items to return",
						"default":     50,
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "miro_get_item",
			Description: "Get details of a specific item",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"item_id": map[string]interface{}{
						"type":        "string",
						"description": "Item ID",
					},
				},
				"required": []string{"board_id", "item_id"},
			},
		},
		{
			Name:        "miro_delete_item",
			Description: "Delete an item from a board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"item_id": map[string]interface{}{
						"type":        "string",
						"description": "Item ID",
					},
				},
				"required": []string{"board_id", "item_id"},
			},
		},
		{
			Name:        "miro_create_frame",
			Description: "Create a frame to group items",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Frame title",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
						"default":     0,
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
						"default":     0,
					},
					"width": map[string]interface{}{
						"type":        "number",
						"description": "Width",
						"default":     800,
					},
					"height": map[string]interface{}{
						"type":        "number",
						"description": "Height",
						"default":     600,
					},
				},
				"required": []string{"board_id", "title"},
			},
		},
		{
			Name:        "miro_export_board",
			Description: "Export board as image or PDF",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Board ID",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Export format",
						"enum":        []string{"png", "pdf"},
						"default":     "png",
					},
				},
				"required": []string{"board_id"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *MiroAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "miro_list_boards":
		return a.listBoards(ctx, args)
	case "miro_create_board":
		return a.createBoard(ctx, args)
	case "miro_get_board":
		return a.getBoard(ctx, args)
	case "miro_create_sticky_note":
		return a.createStickyNote(ctx, args)
	case "miro_create_shape":
		return a.createShape(ctx, args)
	case "miro_create_text":
		return a.createText(ctx, args)
	case "miro_create_connector":
		return a.createConnector(ctx, args)
	case "miro_list_items":
		return a.listItems(ctx, args)
	case "miro_get_item":
		return a.getItem(ctx, args)
	case "miro_delete_item":
		return a.deleteItem(ctx, args)
	case "miro_create_frame":
		return a.createFrame(ctx, args)
	case "miro_export_board":
		return a.exportBoard(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *MiroAdapter) listBoards(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	params := url.Values{}
	if query, ok := args["query"].(string); ok && query != "" {
		params.Set("query", query)
	}
	limit := getIntArg(args, "limit", 50)
	params.Set("limit", fmt.Sprintf("%d", limit))

	resp, err := a.makeRequest(ctx, http.MethodGet, "/boards", params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result MiroBoardsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Boards (%d):\n\n", len(result.Data)))

	for _, board := range result.Data {
		sb.WriteString(fmt.Sprintf("- %s\n", board.Name))
		sb.WriteString(fmt.Sprintf("  ID: %s\n", board.ID))
		sb.WriteString(fmt.Sprintf("  Created: %s\n", board.CreatedAt))
		sb.WriteString(fmt.Sprintf("  URL: %s\n\n", board.ViewLink))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *MiroAdapter) createBoard(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)

	payload := map[string]interface{}{
		"name": name,
	}
	if description != "" {
		payload["description"] = description
	}
	if teamID, ok := args["team_id"].(string); ok && teamID != "" {
		payload["teamId"] = teamID
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/boards", nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var board MiroBoard
	if err := json.Unmarshal(resp, &board); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Board created: %s\nID: %s\nURL: %s", board.Name, board.ID, board.ViewLink)}},
	}, nil
}

func (a *MiroAdapter) getBoard(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)

	endpoint := fmt.Sprintf("/boards/%s", boardID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var board MiroBoard
	if err := json.Unmarshal(resp, &board); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Board: %s\n", board.Name))
	sb.WriteString(fmt.Sprintf("ID: %s\n", board.ID))
	sb.WriteString(fmt.Sprintf("Description: %s\n", board.Description))
	sb.WriteString(fmt.Sprintf("Created: %s\n", board.CreatedAt))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", board.ModifiedAt))
	sb.WriteString(fmt.Sprintf("URL: %s\n", board.ViewLink))
	if board.Owner != nil {
		sb.WriteString(fmt.Sprintf("Owner: %s\n", board.Owner.Name))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *MiroAdapter) createStickyNote(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	content, _ := args["content"].(string)
	x := getFloatArg(args, "x", 0)
	y := getFloatArg(args, "y", 0)
	color, _ := args["color"].(string)
	if color == "" {
		color = "light_yellow"
	}
	width := getFloatArg(args, "width", 228)

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"content": content,
			"shape":   "square",
		},
		"style": map[string]interface{}{
			"fillColor": color,
		},
		"position": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"geometry": map[string]interface{}{
			"width": width,
		},
	}

	endpoint := fmt.Sprintf("/boards/%s/sticky_notes", boardID)
	resp, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var item MiroItem
	if err := json.Unmarshal(resp, &item); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Sticky note created (ID: %s)", item.ID)}},
	}, nil
}

func (a *MiroAdapter) createShape(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	shape, _ := args["shape"].(string)
	content, _ := args["content"].(string)
	x := getFloatArg(args, "x", 0)
	y := getFloatArg(args, "y", 0)
	width := getFloatArg(args, "width", 200)
	height := getFloatArg(args, "height", 200)
	fillColor, _ := args["fill_color"].(string)
	if fillColor == "" {
		fillColor = "#ffffff"
	}
	borderColor, _ := args["border_color"].(string)
	if borderColor == "" {
		borderColor = "#000000"
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"shape":   shape,
			"content": content,
		},
		"style": map[string]interface{}{
			"fillColor":   fillColor,
			"borderColor": borderColor,
		},
		"position": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"geometry": map[string]interface{}{
			"width":  width,
			"height": height,
		},
	}

	endpoint := fmt.Sprintf("/boards/%s/shapes", boardID)
	resp, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var item MiroItem
	if err := json.Unmarshal(resp, &item); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Shape created (ID: %s)", item.ID)}},
	}, nil
}

func (a *MiroAdapter) createText(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	content, _ := args["content"].(string)
	x := getFloatArg(args, "x", 0)
	y := getFloatArg(args, "y", 0)
	width := getFloatArg(args, "width", 200)
	fontSize := getIntArg(args, "font_size", 14)

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"content": content,
		},
		"style": map[string]interface{}{
			"fontSize": fmt.Sprintf("%d", fontSize),
		},
		"position": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"geometry": map[string]interface{}{
			"width": width,
		},
	}

	endpoint := fmt.Sprintf("/boards/%s/texts", boardID)
	resp, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var item MiroItem
	if err := json.Unmarshal(resp, &item); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Text created (ID: %s)", item.ID)}},
	}, nil
}

func (a *MiroAdapter) createConnector(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	startItemID, _ := args["start_item_id"].(string)
	endItemID, _ := args["end_item_id"].(string)
	style, _ := args["style"].(string)
	if style == "" {
		style = "elbowed"
	}
	startCap, _ := args["start_cap"].(string)
	if startCap == "" {
		startCap = "none"
	}
	endCap, _ := args["end_cap"].(string)
	if endCap == "" {
		endCap = "stealth"
	}

	payload := map[string]interface{}{
		"startItem": map[string]interface{}{
			"id": startItemID,
		},
		"endItem": map[string]interface{}{
			"id": endItemID,
		},
		"style": map[string]interface{}{
			"strokeStyle":    style,
			"startStrokeCap": startCap,
			"endStrokeCap":   endCap,
		},
	}

	endpoint := fmt.Sprintf("/boards/%s/connectors", boardID)
	resp, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var item MiroItem
	if err := json.Unmarshal(resp, &item); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Connector created (ID: %s)", item.ID)}},
	}, nil
}

func (a *MiroAdapter) listItems(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	limit := getIntArg(args, "limit", 50)

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if itemType, ok := args["type"].(string); ok && itemType != "" {
		params.Set("type", itemType)
	}

	endpoint := fmt.Sprintf("/boards/%s/items", boardID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result MiroItemsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Items on board (%d):\n\n", len(result.Data)))

	for _, item := range result.Data {
		sb.WriteString(fmt.Sprintf("- [%s] ID: %s\n", item.Type, item.ID))
		if item.Position != nil {
			sb.WriteString(fmt.Sprintf("  Position: (%.0f, %.0f)\n", item.Position.X, item.Position.Y))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *MiroAdapter) getItem(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	itemID, _ := args["item_id"].(string)

	endpoint := fmt.Sprintf("/boards/%s/items/%s", boardID, itemID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var item MiroItem
	if err := json.Unmarshal(resp, &item); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Item: %s\n", item.ID))
	sb.WriteString(fmt.Sprintf("Type: %s\n", item.Type))
	sb.WriteString(fmt.Sprintf("Created: %s\n", item.CreatedAt))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", item.ModifiedAt))
	if item.Position != nil {
		sb.WriteString(fmt.Sprintf("Position: (%.0f, %.0f)\n", item.Position.X, item.Position.Y))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *MiroAdapter) deleteItem(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	itemID, _ := args["item_id"].(string)

	endpoint := fmt.Sprintf("/boards/%s/items/%s", boardID, itemID)
	_, err := a.makeRequest(ctx, http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Item %s deleted successfully", itemID)}},
	}, nil
}

func (a *MiroAdapter) createFrame(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	title, _ := args["title"].(string)
	x := getFloatArg(args, "x", 0)
	y := getFloatArg(args, "y", 0)
	width := getFloatArg(args, "width", 800)
	height := getFloatArg(args, "height", 600)

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"title":  title,
			"format": "custom",
		},
		"position": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"geometry": map[string]interface{}{
			"width":  width,
			"height": height,
		},
	}

	endpoint := fmt.Sprintf("/boards/%s/frames", boardID)
	resp, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var item MiroItem
	if err := json.Unmarshal(resp, &item); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Frame created: %s (ID: %s)", title, item.ID)}},
	}, nil
}

func (a *MiroAdapter) exportBoard(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	format, _ := args["format"].(string)
	if format == "" {
		format = "png"
	}

	// Note: Miro export requires async processing. This returns the board URL for manual export.
	endpoint := fmt.Sprintf("/boards/%s", boardID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var board MiroBoard
	if err := json.Unmarshal(resp, &board); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{
			Type: "text",
			Text: fmt.Sprintf("Board export prepared.\nFormat requested: %s\nBoard URL: %s\n\nNote: For direct export, use the Miro web interface or the async export API.", format, board.ViewLink),
		}},
	}, nil
}

func (a *MiroAdapter) makeRequest(ctx context.Context, method, endpoint string, params url.Values, payload interface{}) ([]byte, error) {
	reqURL := a.config.BaseURL + endpoint
	if params != nil {
		reqURL += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if payload != nil {
		bodyJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return body, nil
}

// Miro API response types

// MiroBoardsResponse represents a boards list response.
type MiroBoardsResponse struct {
	Data []MiroBoard `json:"data"`
}

// MiroBoard represents a Miro board.
type MiroBoard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"createdAt"`
	ModifiedAt  string    `json:"modifiedAt"`
	ViewLink    string    `json:"viewLink"`
	Owner       *MiroUser `json:"owner"`
}

// MiroUser represents a Miro user.
type MiroUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MiroItemsResponse represents an items list response.
type MiroItemsResponse struct {
	Data []MiroItem `json:"data"`
}

// MiroItem represents a Miro item.
type MiroItem struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	CreatedAt  string        `json:"createdAt"`
	ModifiedAt string        `json:"modifiedAt"`
	Position   *MiroPosition `json:"position"`
	Geometry   *MiroGeometry `json:"geometry"`
}

// MiroPosition represents item position.
type MiroPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MiroGeometry represents item geometry.
type MiroGeometry struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}
