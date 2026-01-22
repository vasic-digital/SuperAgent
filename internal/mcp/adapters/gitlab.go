// Package adapters provides MCP server adapters.
// This file implements the GitLab MCP server adapter.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GitLabConfig configures the GitLab adapter.
type GitLabConfig struct {
	BaseURL string        `json:"base_url"`
	Token   string        `json:"token"`
	Timeout time.Duration `json:"timeout"`
}

// DefaultGitLabConfig returns default configuration.
func DefaultGitLabConfig() GitLabConfig {
	return GitLabConfig{
		BaseURL: "https://gitlab.com",
		Timeout: 60 * time.Second,
	}
}

// GitLabAdapter implements the GitLab MCP server.
type GitLabAdapter struct {
	config     GitLabConfig
	httpClient *http.Client
}

// NewGitLabAdapter creates a new GitLab adapter.
func NewGitLabAdapter(config GitLabConfig) *GitLabAdapter {
	return &GitLabAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *GitLabAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "gitlab",
		Version:     "1.0.0",
		Description: "GitLab API integration for repositories, issues, merge requests, and CI/CD",
		Capabilities: []string{
			"repository_management",
			"issue_tracking",
			"merge_requests",
			"ci_cd",
			"user_management",
		},
	}
}

// ListTools returns available tools.
func (a *GitLabAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "gitlab_list_projects",
			Description: "List GitLab projects",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owned": map[string]interface{}{
						"type":        "boolean",
						"description": "Only list owned projects",
						"default":     false,
					},
					"search": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"per_page": map[string]interface{}{
						"type":        "integer",
						"description": "Results per page",
						"default":     20,
					},
				},
			},
		},
		{
			Name:        "gitlab_get_project",
			Description: "Get project details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path (e.g., 'group/project')",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "gitlab_list_issues",
			Description: "List issues in a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Issue state",
						"enum":        []string{"opened", "closed", "all"},
						"default":     "opened",
					},
					"labels": map[string]interface{}{
						"type":        "string",
						"description": "Comma-separated list of labels",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "gitlab_create_issue",
			Description: "Create a new issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Issue title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Issue description",
					},
					"labels": map[string]interface{}{
						"type":        "string",
						"description": "Comma-separated list of labels",
					},
					"assignee_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "integer"},
						"description": "Assignee user IDs",
					},
				},
				"required": []string{"project_id", "title"},
			},
		},
		{
			Name:        "gitlab_list_merge_requests",
			Description: "List merge requests in a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "MR state",
						"enum":        []string{"opened", "closed", "merged", "all"},
						"default":     "opened",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "gitlab_create_merge_request",
			Description: "Create a new merge request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"source_branch": map[string]interface{}{
						"type":        "string",
						"description": "Source branch",
					},
					"target_branch": map[string]interface{}{
						"type":        "string",
						"description": "Target branch",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "MR title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "MR description",
					},
				},
				"required": []string{"project_id", "source_branch", "target_branch", "title"},
			},
		},
		{
			Name:        "gitlab_get_file",
			Description: "Get file content from repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "File path in repository",
					},
					"ref": map[string]interface{}{
						"type":        "string",
						"description": "Branch or commit",
						"default":     "main",
					},
				},
				"required": []string{"project_id", "file_path"},
			},
		},
		{
			Name:        "gitlab_list_pipelines",
			Description: "List CI/CD pipelines",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Pipeline status",
						"enum":        []string{"running", "pending", "success", "failed", "canceled", "skipped"},
					},
					"ref": map[string]interface{}{
						"type":        "string",
						"description": "Branch name",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "gitlab_trigger_pipeline",
			Description: "Trigger a new pipeline",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"ref": map[string]interface{}{
						"type":        "string",
						"description": "Branch to run pipeline on",
					},
					"variables": map[string]interface{}{
						"type":        "object",
						"description": "Pipeline variables",
					},
				},
				"required": []string{"project_id", "ref"},
			},
		},
		{
			Name:        "gitlab_list_branches",
			Description: "List repository branches",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or path",
					},
					"search": map[string]interface{}{
						"type":        "string",
						"description": "Search pattern",
					},
				},
				"required": []string{"project_id"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *GitLabAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "gitlab_list_projects":
		return a.listProjects(ctx, args)
	case "gitlab_get_project":
		return a.getProject(ctx, args)
	case "gitlab_list_issues":
		return a.listIssues(ctx, args)
	case "gitlab_create_issue":
		return a.createIssue(ctx, args)
	case "gitlab_list_merge_requests":
		return a.listMergeRequests(ctx, args)
	case "gitlab_create_merge_request":
		return a.createMergeRequest(ctx, args)
	case "gitlab_get_file":
		return a.getFile(ctx, args)
	case "gitlab_list_pipelines":
		return a.listPipelines(ctx, args)
	case "gitlab_trigger_pipeline":
		return a.triggerPipeline(ctx, args)
	case "gitlab_list_branches":
		return a.listBranches(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *GitLabAdapter) listProjects(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Implementation would call GitLab API
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "GitLab list projects - implementation requires GitLab API client"}},
	}, nil
}

func (a *GitLabAdapter) getProject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab get project %s - implementation requires GitLab API client", projectID)}},
	}, nil
}

func (a *GitLabAdapter) listIssues(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab list issues for %s - implementation requires GitLab API client", projectID)}},
	}, nil
}

func (a *GitLabAdapter) createIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	title, _ := args["title"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab create issue '%s' in %s - implementation requires GitLab API client", title, projectID)}},
	}, nil
}

func (a *GitLabAdapter) listMergeRequests(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab list MRs for %s - implementation requires GitLab API client", projectID)}},
	}, nil
}

func (a *GitLabAdapter) createMergeRequest(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	title, _ := args["title"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab create MR '%s' in %s - implementation requires GitLab API client", title, projectID)}},
	}, nil
}

func (a *GitLabAdapter) getFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	filePath, _ := args["file_path"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab get file %s from %s - implementation requires GitLab API client", filePath, projectID)}},
	}, nil
}

func (a *GitLabAdapter) listPipelines(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab list pipelines for %s - implementation requires GitLab API client", projectID)}},
	}, nil
}

func (a *GitLabAdapter) triggerPipeline(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	ref, _ := args["ref"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab trigger pipeline for %s on %s - implementation requires GitLab API client", projectID, ref)}},
	}, nil
}

func (a *GitLabAdapter) listBranches(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID, _ := args["project_id"].(string)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("GitLab list branches for %s - implementation requires GitLab API client", projectID)}},
	}, nil
}

// GitLab response types
type GitLabProject struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	PathWithNamespace string    `json:"path_with_namespace"`
	Description       string    `json:"description"`
	DefaultBranch     string    `json:"default_branch"`
	WebURL            string    `json:"web_url"`
	CreatedAt         time.Time `json:"created_at"`
	LastActivityAt    time.Time `json:"last_activity_at"`
}

type GitLabIssue struct {
	ID          int       `json:"id"`
	IID         int       `json:"iid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Labels      []string  `json:"labels"`
	WebURL      string    `json:"web_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type GitLabMergeRequest struct {
	ID           int       `json:"id"`
	IID          int       `json:"iid"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	SourceBranch string    `json:"source_branch"`
	TargetBranch string    `json:"target_branch"`
	WebURL       string    `json:"web_url"`
	CreatedAt    time.Time `json:"created_at"`
}

type GitLabPipeline struct {
	ID        int       `json:"id"`
	Status    string    `json:"status"`
	Ref       string    `json:"ref"`
	SHA       string    `json:"sha"`
	WebURL    string    `json:"web_url"`
	CreatedAt time.Time `json:"created_at"`
}

// Serialize serializes GitLab types to JSON
func (p *GitLabProject) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

func formatGitLabProjects(projects []GitLabProject) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d projects:\n\n", len(projects)))

	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", p.Name, p.PathWithNamespace))
		sb.WriteString(fmt.Sprintf("  Branch: %s, URL: %s\n", p.DefaultBranch, p.WebURL))
	}

	return sb.String()
}
