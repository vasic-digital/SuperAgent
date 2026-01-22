// Package adapters provides MCP server adapters.
// This file implements the Figma MCP server adapter for design integration.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FigmaConfig configures the Figma adapter.
type FigmaConfig struct {
	AccessToken string        `json:"access_token"`
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
}

// DefaultFigmaConfig returns default configuration.
func DefaultFigmaConfig() FigmaConfig {
	return FigmaConfig{
		BaseURL: "https://api.figma.com/v1",
		Timeout: 60 * time.Second,
	}
}

// FigmaAdapter implements the Figma MCP server.
type FigmaAdapter struct {
	config     FigmaConfig
	httpClient *http.Client
}

// NewFigmaAdapter creates a new Figma adapter.
func NewFigmaAdapter(config FigmaConfig) *FigmaAdapter {
	if config.BaseURL == "" {
		config.BaseURL = DefaultFigmaConfig().BaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultFigmaConfig().Timeout
	}
	return &FigmaAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *FigmaAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "figma",
		Version:     "1.0.0",
		Description: "Figma design integration for reading and modifying design files",
		Capabilities: []string{
			"get_file",
			"get_file_nodes",
			"get_images",
			"get_components",
			"get_styles",
			"get_comments",
			"post_comment",
			"get_team_projects",
			"get_project_files",
		},
	}
}

// ListTools returns available tools.
func (a *FigmaAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "figma_get_file",
			Description: "Get a Figma file by key with all design data",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key (from URL)",
					},
					"depth": map[string]interface{}{
						"type":        "integer",
						"description": "Depth of node tree to return (default: 2)",
						"default":     2,
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_get_file_nodes",
			Description: "Get specific nodes from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key",
					},
					"node_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Array of node IDs to retrieve",
					},
				},
				"required": []string{"file_key", "node_ids"},
			},
		},
		{
			Name:        "figma_get_images",
			Description: "Export images from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key",
					},
					"node_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Array of node IDs to export",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Image format: jpg, png, svg, pdf",
						"enum":        []string{"jpg", "png", "svg", "pdf"},
						"default":     "png",
					},
					"scale": map[string]interface{}{
						"type":        "number",
						"description": "Image scale (0.01 to 4)",
						"default":     1,
					},
				},
				"required": []string{"file_key", "node_ids"},
			},
		},
		{
			Name:        "figma_get_components",
			Description: "Get all components from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key",
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_get_styles",
			Description: "Get all styles from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key",
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_get_comments",
			Description: "Get comments on a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key",
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_post_comment",
			Description: "Post a comment on a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The Figma file key",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Comment message",
					},
					"node_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional node ID to attach comment to",
					},
				},
				"required": []string{"file_key", "message"},
			},
		},
		{
			Name:        "figma_get_team_projects",
			Description: "Get all projects for a team",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "The team ID",
					},
				},
				"required": []string{"team_id"},
			},
		},
		{
			Name:        "figma_get_project_files",
			Description: "Get all files in a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "The project ID",
					},
				},
				"required": []string{"project_id"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *FigmaAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "figma_get_file":
		return a.getFile(ctx, args)
	case "figma_get_file_nodes":
		return a.getFileNodes(ctx, args)
	case "figma_get_images":
		return a.getImages(ctx, args)
	case "figma_get_components":
		return a.getComponents(ctx, args)
	case "figma_get_styles":
		return a.getStyles(ctx, args)
	case "figma_get_comments":
		return a.getComments(ctx, args)
	case "figma_post_comment":
		return a.postComment(ctx, args)
	case "figma_get_team_projects":
		return a.getTeamProjects(ctx, args)
	case "figma_get_project_files":
		return a.getProjectFiles(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *FigmaAdapter) getFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)
	depth := getIntArg(args, "depth", 2)

	endpoint := fmt.Sprintf("/files/%s?depth=%d", fileKey, depth)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var file FigmaFileResponse
	if err := json.Unmarshal(resp, &file); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %s\n", file.Name))
	sb.WriteString(fmt.Sprintf("Last Modified: %s\n", file.LastModified))
	sb.WriteString(fmt.Sprintf("Version: %s\n", file.Version))
	sb.WriteString(fmt.Sprintf("Thumbnail: %s\n\n", file.ThumbnailURL))

	sb.WriteString("Document Structure:\n")
	a.formatNode(&sb, &file.Document, 0)

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) formatNode(sb *strings.Builder, node *FigmaNode, indent int) {
	prefix := strings.Repeat("  ", indent)
	sb.WriteString(fmt.Sprintf("%s- %s (%s) [ID: %s]\n", prefix, node.Name, node.Type, node.ID))

	if len(node.Children) > 0 && indent < 3 {
		for _, child := range node.Children {
			a.formatNode(sb, &child, indent+1)
		}
	}
}

func (a *FigmaAdapter) getFileNodes(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)
	nodeIDsRaw, _ := args["node_ids"].([]interface{})

	var nodeIDs []string
	for _, id := range nodeIDsRaw {
		if s, ok := id.(string); ok {
			nodeIDs = append(nodeIDs, s)
		}
	}

	endpoint := fmt.Sprintf("/files/%s/nodes?ids=%s", fileKey, strings.Join(nodeIDs, ","))
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var nodes FigmaNodesResponse
	if err := json.Unmarshal(resp, &nodes); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Retrieved %d nodes:\n\n", len(nodes.Nodes)))

	for nodeID, nodeData := range nodes.Nodes {
		sb.WriteString(fmt.Sprintf("Node ID: %s\n", nodeID))
		if nodeData.Document != nil {
			sb.WriteString(fmt.Sprintf("  Name: %s\n", nodeData.Document.Name))
			sb.WriteString(fmt.Sprintf("  Type: %s\n", nodeData.Document.Type))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) getImages(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)
	nodeIDsRaw, _ := args["node_ids"].([]interface{})
	format, _ := args["format"].(string)
	if format == "" {
		format = "png"
	}
	scale := getFloatArg(args, "scale", 1.0)

	var nodeIDs []string
	for _, id := range nodeIDsRaw {
		if s, ok := id.(string); ok {
			nodeIDs = append(nodeIDs, s)
		}
	}

	endpoint := fmt.Sprintf("/images/%s?ids=%s&format=%s&scale=%.2f", fileKey, strings.Join(nodeIDs, ","), format, scale)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var images FigmaImagesResponse
	if err := json.Unmarshal(resp, &images); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Exported %d images (format: %s, scale: %.2f):\n\n", len(images.Images), format, scale))

	for nodeID, imageURL := range images.Images {
		sb.WriteString(fmt.Sprintf("Node %s:\n  %s\n\n", nodeID, imageURL))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) getComponents(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)

	endpoint := fmt.Sprintf("/files/%s?depth=1", fileKey)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var file FigmaFileResponse
	if err := json.Unmarshal(resp, &file); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Components in file '%s':\n\n", file.Name))

	for key, comp := range file.Components {
		sb.WriteString(fmt.Sprintf("- %s (key: %s)\n", comp.Name, key))
		if comp.Description != "" {
			sb.WriteString(fmt.Sprintf("  Description: %s\n", comp.Description))
		}
	}

	if len(file.Components) == 0 {
		sb.WriteString("No components found in this file.\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) getStyles(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)

	endpoint := fmt.Sprintf("/files/%s/styles", fileKey)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var styles FigmaStylesResponse
	if err := json.Unmarshal(resp, &styles); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString("Styles:\n\n")

	if styles.Meta != nil && styles.Meta.Styles != nil {
		for _, style := range styles.Meta.Styles {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", style.Name, style.StyleType))
			if style.Description != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", style.Description))
			}
		}
	} else {
		sb.WriteString("No styles found.\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) getComments(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)

	endpoint := fmt.Sprintf("/files/%s/comments", fileKey)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var comments FigmaCommentsResponse
	if err := json.Unmarshal(resp, &comments); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Comments (%d):\n\n", len(comments.Comments)))

	for _, comment := range comments.Comments {
		sb.WriteString(fmt.Sprintf("- [%s] %s:\n", comment.CreatedAt, comment.User.Handle))
		sb.WriteString(fmt.Sprintf("  %s\n\n", comment.Message))
	}

	if len(comments.Comments) == 0 {
		sb.WriteString("No comments found.\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) postComment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileKey, _ := args["file_key"].(string)
	message, _ := args["message"].(string)
	nodeID, _ := args["node_id"].(string)

	body := map[string]interface{}{
		"message": message,
	}
	if nodeID != "" {
		body["client_meta"] = map[string]interface{}{
			"node_id": nodeID,
		}
	}

	bodyJSON, _ := json.Marshal(body)
	endpoint := fmt.Sprintf("/files/%s/comments", fileKey)
	resp, err := a.makeRequest(ctx, http.MethodPost, endpoint, bodyJSON)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var comment FigmaComment
	if err := json.Unmarshal(resp, &comment); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Comment posted successfully (ID: %s)", comment.ID)}},
	}, nil
}

func (a *FigmaAdapter) getTeamProjects(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	teamID, _ := args["team_id"].(string)

	endpoint := fmt.Sprintf("/teams/%s/projects", teamID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var projects FigmaProjectsResponse
	if err := json.Unmarshal(resp, &projects); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Projects in team (%d):\n\n", len(projects.Projects)))

	for _, project := range projects.Projects {
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)\n", project.Name, project.ID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) getProjectFiles(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)

	endpoint := fmt.Sprintf("/projects/%s/files", projectID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var files FigmaFilesResponse
	if err := json.Unmarshal(resp, &files); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Files in project (%d):\n\n", len(files.Files)))

	for _, file := range files.Files {
		sb.WriteString(fmt.Sprintf("- %s (key: %s)\n", file.Name, file.Key))
		sb.WriteString(fmt.Sprintf("  Last modified: %s\n", file.LastModified))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *FigmaAdapter) makeRequest(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	reqURL := a.config.BaseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Figma-Token", a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

func getFloatArg(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return defaultVal
}

// Figma API response types

// FigmaFileResponse represents a Figma file response.
type FigmaFileResponse struct {
	Name         string                   `json:"name"`
	LastModified string                   `json:"lastModified"`
	Version      string                   `json:"version"`
	ThumbnailURL string                   `json:"thumbnailUrl"`
	Document     FigmaNode                `json:"document"`
	Components   map[string]FigmaCompDef  `json:"components"`
	Styles       map[string]FigmaStyleDef `json:"styles"`
}

// FigmaNode represents a node in a Figma document.
type FigmaNode struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Children []FigmaNode `json:"children,omitempty"`
}

// FigmaCompDef represents a component definition.
type FigmaCompDef struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FigmaStyleDef represents a style definition.
type FigmaStyleDef struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	StyleType   string `json:"styleType"`
	Description string `json:"description"`
}

// FigmaNodesResponse represents a nodes response.
type FigmaNodesResponse struct {
	Nodes map[string]*FigmaNodeData `json:"nodes"`
}

// FigmaNodeData represents node data.
type FigmaNodeData struct {
	Document *FigmaNode `json:"document"`
}

// FigmaImagesResponse represents an images response.
type FigmaImagesResponse struct {
	Images map[string]string `json:"images"`
	Err    string            `json:"err,omitempty"`
}

// FigmaStylesResponse represents a styles response.
type FigmaStylesResponse struct {
	Meta *struct {
		Styles []FigmaStyle `json:"styles"`
	} `json:"meta"`
}

// FigmaStyle represents a style.
type FigmaStyle struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	StyleType   string `json:"style_type"`
	Description string `json:"description"`
}

// FigmaCommentsResponse represents a comments response.
type FigmaCommentsResponse struct {
	Comments []FigmaComment `json:"comments"`
}

// FigmaComment represents a comment.
type FigmaComment struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	CreatedAt string    `json:"created_at"`
	User      FigmaUser `json:"user"`
}

// FigmaUser represents a Figma user.
type FigmaUser struct {
	Handle string `json:"handle"`
	ImgURL string `json:"img_url"`
}

// FigmaProjectsResponse represents a projects response.
type FigmaProjectsResponse struct {
	Projects []FigmaProject `json:"projects"`
}

// FigmaProject represents a Figma project.
type FigmaProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FigmaFilesResponse represents a files response.
type FigmaFilesResponse struct {
	Files []FigmaFileMeta `json:"files"`
}

// FigmaFileMeta represents file metadata.
type FigmaFileMeta struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	LastModified string `json:"last_modified"`
	ThumbnailURL string `json:"thumbnail_url"`
}
