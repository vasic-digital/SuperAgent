// Package adapters provides MCP server adapters.
// This file implements the Google Drive MCP server adapter.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// GoogleDriveConfig configures the Google Drive adapter.
type GoogleDriveConfig struct {
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	RefreshToken string        `json:"refresh_token"`
	Timeout      time.Duration `json:"timeout"`
}

// DefaultGoogleDriveConfig returns default configuration.
func DefaultGoogleDriveConfig() GoogleDriveConfig {
	return GoogleDriveConfig{
		Timeout: 60 * time.Second,
	}
}

// GoogleDriveAdapter implements the Google Drive MCP server.
type GoogleDriveAdapter struct {
	config GoogleDriveConfig
	client GoogleDriveClient
}

// GoogleDriveClient interface for Google Drive operations.
type GoogleDriveClient interface {
	ListFiles(ctx context.Context, query string, pageSize int) ([]DriveFile, error)
	GetFile(ctx context.Context, fileID string) (*DriveFile, error)
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)
	CreateFile(ctx context.Context, name, mimeType, parentID string, content io.Reader) (*DriveFile, error)
	UpdateFile(ctx context.Context, fileID string, content io.Reader) (*DriveFile, error)
	DeleteFile(ctx context.Context, fileID string) error
	CopyFile(ctx context.Context, fileID, name string) (*DriveFile, error)
	MoveFile(ctx context.Context, fileID, newParentID string) (*DriveFile, error)
	CreateFolder(ctx context.Context, name, parentID string) (*DriveFile, error)
	SearchFiles(ctx context.Context, query string) ([]DriveFile, error)
	ShareFile(ctx context.Context, fileID, email, role string) error
}

// DriveFile represents a Google Drive file.
type DriveFile struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	MimeType     string    `json:"mimeType"`
	Size         int64     `json:"size"`
	CreatedTime  time.Time `json:"createdTime"`
	ModifiedTime time.Time `json:"modifiedTime"`
	Parents      []string  `json:"parents,omitempty"`
	WebViewLink  string    `json:"webViewLink,omitempty"`
	IconLink     string    `json:"iconLink,omitempty"`
	Owners       []Owner   `json:"owners,omitempty"`
	Shared       bool      `json:"shared"`
}

// Owner represents a file owner.
type Owner struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// NewGoogleDriveAdapter creates a new Google Drive adapter.
func NewGoogleDriveAdapter(config GoogleDriveConfig, client GoogleDriveClient) *GoogleDriveAdapter {
	return &GoogleDriveAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *GoogleDriveAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "google-drive",
		Version:     "1.0.0",
		Description: "Google Drive file management including list, upload, download, and sharing",
		Capabilities: []string{
			"file_management",
			"folder_management",
			"sharing",
			"search",
		},
	}
}

// ListTools returns available tools.
func (a *GoogleDriveAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "drive_list_files",
			Description: "List files in Google Drive",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (Google Drive query syntax)",
						"default":     "",
					},
					"page_size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of files to return",
						"default":     50,
					},
				},
			},
		},
		{
			Name:        "drive_get_file",
			Description: "Get file metadata",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID",
					},
				},
				"required": []string{"file_id"},
			},
		},
		{
			Name:        "drive_download_file",
			Description: "Download file content",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID",
					},
				},
				"required": []string{"file_id"},
			},
		},
		{
			Name:        "drive_create_file",
			Description: "Create a new file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "File name",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "File content",
					},
					"mime_type": map[string]interface{}{
						"type":        "string",
						"description": "MIME type",
						"default":     "text/plain",
					},
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "Parent folder ID",
						"default":     "",
					},
				},
				"required": []string{"name", "content"},
			},
		},
		{
			Name:        "drive_update_file",
			Description: "Update file content",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "New file content",
					},
				},
				"required": []string{"file_id", "content"},
			},
		},
		{
			Name:        "drive_delete_file",
			Description: "Delete a file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID",
					},
				},
				"required": []string{"file_id"},
			},
		},
		{
			Name:        "drive_create_folder",
			Description: "Create a new folder",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Folder name",
					},
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "Parent folder ID",
						"default":     "",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "drive_search",
			Description: "Search for files",
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
			Name:        "drive_share_file",
			Description: "Share a file with another user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Email address to share with",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"description": "Permission role",
						"enum":        []string{"reader", "writer", "commenter", "owner"},
						"default":     "reader",
					},
				},
				"required": []string{"file_id", "email"},
			},
		},
		{
			Name:        "drive_copy_file",
			Description: "Copy a file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID to copy",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name for the copy",
					},
				},
				"required": []string{"file_id", "name"},
			},
		},
		{
			Name:        "drive_move_file",
			Description: "Move a file to a different folder",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID",
					},
					"new_parent_id": map[string]interface{}{
						"type":        "string",
						"description": "Destination folder ID",
					},
				},
				"required": []string{"file_id", "new_parent_id"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *GoogleDriveAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "drive_list_files":
		return a.listFiles(ctx, args)
	case "drive_get_file":
		return a.getFile(ctx, args)
	case "drive_download_file":
		return a.downloadFile(ctx, args)
	case "drive_create_file":
		return a.createFile(ctx, args)
	case "drive_update_file":
		return a.updateFile(ctx, args)
	case "drive_delete_file":
		return a.deleteFile(ctx, args)
	case "drive_create_folder":
		return a.createFolder(ctx, args)
	case "drive_search":
		return a.search(ctx, args)
	case "drive_share_file":
		return a.shareFile(ctx, args)
	case "drive_copy_file":
		return a.copyFile(ctx, args)
	case "drive_move_file":
		return a.moveFile(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *GoogleDriveAdapter) listFiles(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	pageSize := getIntArg(args, "page_size", 50)

	files, err := a.client.ListFiles(ctx, query, pageSize)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d files:\n\n", len(files)))

	for _, file := range files {
		icon := "📄"
		if strings.Contains(file.MimeType, "folder") {
			icon = "📁"
		}
		sb.WriteString(fmt.Sprintf("%s **%s**\n", icon, file.Name))
		sb.WriteString(fmt.Sprintf("   ID: %s\n", file.ID))
		sb.WriteString(fmt.Sprintf("   Type: %s, Size: %d bytes\n", file.MimeType, file.Size))
		sb.WriteString(fmt.Sprintf("   Modified: %s\n\n", file.ModifiedTime.Format(time.RFC3339)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *GoogleDriveAdapter) getFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)

	file, err := a.client.GetFile(ctx, fileID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(file, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(data)}},
	}, nil
}

func (a *GoogleDriveAdapter) downloadFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)

	reader, err := a.client.DownloadFile(ctx, fileID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}
	defer func() { _ = reader.Close() }()

	content, err := io.ReadAll(reader)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(content)}},
	}, nil
}

func (a *GoogleDriveAdapter) createFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	content, _ := args["content"].(string)
	mimeType, _ := args["mime_type"].(string)
	if mimeType == "" {
		mimeType = "text/plain"
	}
	parentID, _ := args["parent_id"].(string)

	file, err := a.client.CreateFile(ctx, name, mimeType, parentID, strings.NewReader(content))
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created file '%s' with ID: %s", file.Name, file.ID)}},
	}, nil
}

func (a *GoogleDriveAdapter) updateFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)
	content, _ := args["content"].(string)

	file, err := a.client.UpdateFile(ctx, fileID, strings.NewReader(content))
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Updated file '%s'", file.Name)}},
	}, nil
}

func (a *GoogleDriveAdapter) deleteFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)

	err := a.client.DeleteFile(ctx, fileID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Deleted file with ID: %s", fileID)}},
	}, nil
}

func (a *GoogleDriveAdapter) createFolder(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	parentID, _ := args["parent_id"].(string)

	folder, err := a.client.CreateFolder(ctx, name, parentID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created folder '%s' with ID: %s", folder.Name, folder.ID)}},
	}, nil
}

func (a *GoogleDriveAdapter) search(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)

	files, err := a.client.SearchFiles(ctx, query)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for '%s' (%d files):\n\n", query, len(files)))

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)\n", file.Name, file.ID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *GoogleDriveAdapter) shareFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)
	email, _ := args["email"].(string)
	role, _ := args["role"].(string)
	if role == "" {
		role = "reader"
	}

	err := a.client.ShareFile(ctx, fileID, email, role)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Shared file %s with %s as %s", fileID, email, role)}},
	}, nil
}

func (a *GoogleDriveAdapter) copyFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)
	name, _ := args["name"].(string)

	file, err := a.client.CopyFile(ctx, fileID, name)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Copied file to '%s' (ID: %s)", file.Name, file.ID)}},
	}, nil
}

func (a *GoogleDriveAdapter) moveFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fileID, _ := args["file_id"].(string)
	newParentID, _ := args["new_parent_id"].(string)

	file, err := a.client.MoveFile(ctx, fileID, newParentID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Moved file '%s' to folder %s", file.Name, newParentID)}},
	}, nil
}
