// Package servers provides MCP server adapters for various services.
package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// FigmaAdapter provides MCP-compatible interface to Figma API.
type FigmaAdapter struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
	mu         sync.RWMutex
	connected  bool
}

// FigmaAdapterConfig holds configuration for FigmaAdapter.
type FigmaAdapterConfig struct {
	APIToken string
	Timeout  time.Duration
}

// FigmaFile represents a Figma file.
type FigmaFile struct {
	Name          string          `json:"name"`
	LastModified  string          `json:"lastModified"`
	ThumbnailURL  string          `json:"thumbnailUrl"`
	Version       string          `json:"version"`
	Document      *FigmaDocument  `json:"document,omitempty"`
	Components    map[string]FigmaComponent `json:"components,omitempty"`
	SchemaVersion int             `json:"schemaVersion"`
}

// FigmaDocument represents the document structure.
type FigmaDocument struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	Children []FigmaNode  `json:"children,omitempty"`
}

// FigmaNode represents a node in the Figma document tree.
type FigmaNode struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Children         []FigmaNode            `json:"children,omitempty"`
	AbsoluteBoundingBox *FigmaRect          `json:"absoluteBoundingBox,omitempty"`
	Fills            []FigmaPaint           `json:"fills,omitempty"`
	Strokes          []FigmaPaint           `json:"strokes,omitempty"`
	StrokeWeight     float64                `json:"strokeWeight,omitempty"`
	CornerRadius     float64                `json:"cornerRadius,omitempty"`
	Characters       string                 `json:"characters,omitempty"`
	Style            map[string]interface{} `json:"style,omitempty"`
}

// FigmaRect represents a bounding box.
type FigmaRect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// FigmaPaint represents a fill or stroke.
type FigmaPaint struct {
	Type      string     `json:"type"`
	Visible   bool       `json:"visible"`
	Opacity   float64    `json:"opacity"`
	Color     *FigmaColor `json:"color,omitempty"`
	BlendMode string     `json:"blendMode,omitempty"`
}

// FigmaColor represents a color.
type FigmaColor struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

// FigmaComponent represents a component in Figma.
type FigmaComponent struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FigmaComment represents a comment on a file.
type FigmaComment struct {
	ID           string    `json:"id"`
	Message      string    `json:"message"`
	FileKey      string    `json:"file_key"`
	ClientMeta   interface{} `json:"client_meta,omitempty"`
	CreatedAt    string    `json:"created_at"`
	ResolvedAt   string    `json:"resolved_at,omitempty"`
	User         FigmaUser `json:"user"`
}

// FigmaUser represents a Figma user.
type FigmaUser struct {
	Handle string `json:"handle"`
	ImgURL string `json:"img_url"`
	ID     string `json:"id"`
}

// FigmaProject represents a Figma project.
type FigmaProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FigmaTeamProject represents a team's projects.
type FigmaTeamProject struct {
	Name     string         `json:"name"`
	Projects []FigmaProject `json:"projects"`
}

// NewFigmaAdapter creates a new Figma adapter.
func NewFigmaAdapter(config FigmaAdapterConfig) *FigmaAdapter {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &FigmaAdapter{
		baseURL:  "https://api.figma.com/v1",
		apiToken: config.APIToken,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Connect establishes connection to Figma API.
func (a *FigmaAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Test connection by getting current user
	resp, err := a.doRequest(ctx, "GET", "/me", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Figma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Figma authentication failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	a.connected = true
	return nil
}

// IsConnected returns whether the adapter is connected.
func (a *FigmaAdapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// Health checks the health of Figma connection.
func (a *FigmaAdapter) Health(ctx context.Context) error {
	resp, err := a.doRequest(ctx, "GET", "/me", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// GetFile retrieves a Figma file by key.
func (a *FigmaAdapter) GetFile(ctx context.Context, fileKey string) (*FigmaFile, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/files/%s", fileKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get file: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var file FigmaFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	return &file, nil
}

// GetFileNodes retrieves specific nodes from a Figma file.
func (a *FigmaAdapter) GetFileNodes(ctx context.Context, fileKey string, nodeIDs []string) (map[string]FigmaNode, error) {
	// Build query string
	path := fmt.Sprintf("/files/%s/nodes?ids=%s", fileKey, joinIDs(nodeIDs))

	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get nodes: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Nodes map[string]struct {
			Document FigmaNode `json:"document"`
		} `json:"nodes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode nodes: %w", err)
	}

	nodes := make(map[string]FigmaNode)
	for id, node := range response.Nodes {
		nodes[id] = node.Document
	}

	return nodes, nil
}

// GetImages exports images from a Figma file.
func (a *FigmaAdapter) GetImages(ctx context.Context, fileKey string, nodeIDs []string, format string, scale float64) (map[string]string, error) {
	if format == "" {
		format = "png"
	}
	if scale == 0 {
		scale = 1.0
	}

	path := fmt.Sprintf("/images/%s?ids=%s&format=%s&scale=%f", fileKey, joinIDs(nodeIDs), format, scale)

	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get images: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Images map[string]string `json:"images"`
		Err    string            `json:"err,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode images: %w", err)
	}

	if response.Err != "" {
		return nil, fmt.Errorf("Figma API error: %s", response.Err)
	}

	return response.Images, nil
}

// GetComments retrieves comments for a file.
func (a *FigmaAdapter) GetComments(ctx context.Context, fileKey string) ([]FigmaComment, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/files/%s/comments", fileKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get comments: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Comments []FigmaComment `json:"comments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	return response.Comments, nil
}

// PostComment adds a comment to a file.
func (a *FigmaAdapter) PostComment(ctx context.Context, fileKey, message string, position *FigmaRect) (*FigmaComment, error) {
	body := map[string]interface{}{
		"message": message,
	}
	if position != nil {
		body["client_meta"] = map[string]interface{}{
			"x": position.X,
			"y": position.Y,
		}
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/files/%s/comments", fileKey), body)
	if err != nil {
		return nil, fmt.Errorf("failed to post comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to post comment: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var comment FigmaComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, fmt.Errorf("failed to decode comment: %w", err)
	}

	return &comment, nil
}

// GetTeamProjects retrieves projects for a team.
func (a *FigmaAdapter) GetTeamProjects(ctx context.Context, teamID string) (*FigmaTeamProject, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/teams/%s/projects", teamID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get team projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get team projects: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var team FigmaTeamProject
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		return nil, fmt.Errorf("failed to decode team projects: %w", err)
	}

	return &team, nil
}

// GetProjectFiles retrieves files for a project.
func (a *FigmaAdapter) GetProjectFiles(ctx context.Context, projectID string) ([]FigmaFile, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/projects/%s/files", projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get project files: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Files []FigmaFile `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode project files: %w", err)
	}

	return response.Files, nil
}

// GetFileComponents retrieves components from a file.
func (a *FigmaAdapter) GetFileComponents(ctx context.Context, fileKey string) (map[string]FigmaComponent, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/files/%s/components", fileKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get components: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Meta struct {
			Components map[string]FigmaComponent `json:"components"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode components: %w", err)
	}

	return response.Meta.Components, nil
}

// GetFileStyles retrieves styles from a file.
func (a *FigmaAdapter) GetFileStyles(ctx context.Context, fileKey string) (map[string]interface{}, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/files/%s/styles", fileKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get styles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get styles: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Meta struct {
			Styles map[string]interface{} `json:"styles"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode styles: %w", err)
	}

	return response.Meta.Styles, nil
}

// Close closes the adapter connection.
func (a *FigmaAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// doRequest performs an HTTP request to Figma API.
func (a *FigmaAdapter) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Figma-Token", a.apiToken)

	return a.httpClient.Do(req)
}

// GetMCPTools returns the MCP tool definitions for Figma.
func (a *FigmaAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "figma_get_file",
			Description: "Get a Figma file by its key",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The unique key of the Figma file",
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
						"description": "The unique key of the Figma file",
					},
					"node_ids": map[string]interface{}{
						"type":        "array",
						"description": "Array of node IDs to retrieve",
						"items":       map[string]interface{}{"type": "string"},
					},
				},
				"required": []string{"file_key", "node_ids"},
			},
		},
		{
			Name:        "figma_export_images",
			Description: "Export images from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The unique key of the Figma file",
					},
					"node_ids": map[string]interface{}{
						"type":        "array",
						"description": "Array of node IDs to export",
						"items":       map[string]interface{}{"type": "string"},
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Export format (jpg, png, svg, pdf)",
						"enum":        []string{"jpg", "png", "svg", "pdf"},
						"default":     "png",
					},
					"scale": map[string]interface{}{
						"type":        "number",
						"description": "Export scale (0.01 to 4)",
						"default":     1,
					},
				},
				"required": []string{"file_key", "node_ids"},
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
						"description": "The unique key of the Figma file",
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_post_comment",
			Description: "Add a comment to a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The unique key of the Figma file",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The comment message",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "Optional X position for the comment",
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Optional Y position for the comment",
					},
				},
				"required": []string{"file_key", "message"},
			},
		},
		{
			Name:        "figma_get_components",
			Description: "Get components from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The unique key of the Figma file",
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_get_styles",
			Description: "Get styles from a Figma file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_key": map[string]interface{}{
						"type":        "string",
						"description": "The unique key of the Figma file",
					},
				},
				"required": []string{"file_key"},
			},
		},
		{
			Name:        "figma_get_team_projects",
			Description: "Get projects for a Figma team",
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
			Description: "Get files in a Figma project",
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

// Helper function to join IDs for URL
func joinIDs(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	result := ids[0]
	for i := 1; i < len(ids); i++ {
		result += "," + ids[i]
	}
	return result
}
