// Package adapters provides MCP server adapters.
// This file implements the Asana MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AsanaConfig configures the Asana adapter.
type AsanaConfig struct {
	AccessToken string        `json:"access_token"`
	Timeout     time.Duration `json:"timeout"`
	WorkspaceID string        `json:"workspace_id,omitempty"`
}

// DefaultAsanaConfig returns default configuration.
func DefaultAsanaConfig() AsanaConfig {
	return AsanaConfig{
		Timeout: 30 * time.Second,
	}
}

// AsanaAdapter implements the Asana MCP server.
type AsanaAdapter struct {
	config AsanaConfig
	client AsanaClient
}

// AsanaClient interface for Asana operations.
type AsanaClient interface {
	// Tasks
	GetTask(ctx context.Context, taskID string) (*AsanaTask, error)
	ListTasks(ctx context.Context, filter AsanaTaskFilter) ([]AsanaTask, error)
	CreateTask(ctx context.Context, input CreateAsanaTaskInput) (*AsanaTask, error)
	UpdateTask(ctx context.Context, taskID string, input UpdateAsanaTaskInput) (*AsanaTask, error)
	DeleteTask(ctx context.Context, taskID string) error
	CompleteTask(ctx context.Context, taskID string) error

	// Projects
	ListProjects(ctx context.Context, workspaceID string) ([]AsanaProject, error)
	GetProject(ctx context.Context, projectID string) (*AsanaProject, error)
	CreateProject(ctx context.Context, input CreateAsanaProjectInput) (*AsanaProject, error)
	ListProjectTasks(ctx context.Context, projectID string) ([]AsanaTask, error)

	// Sections
	ListSections(ctx context.Context, projectID string) ([]AsanaSection, error)
	CreateSection(ctx context.Context, projectID, name string) (*AsanaSection, error)
	AddTaskToSection(ctx context.Context, sectionID, taskID string) error

	// Workspaces
	ListWorkspaces(ctx context.Context) ([]AsanaWorkspace, error)

	// Users
	ListUsers(ctx context.Context, workspaceID string) ([]AsanaUser, error)
	GetMe(ctx context.Context) (*AsanaUser, error)

	// Tags
	ListTags(ctx context.Context, workspaceID string) ([]AsanaTag, error)
	AddTagToTask(ctx context.Context, taskID, tagID string) error
	RemoveTagFromTask(ctx context.Context, taskID, tagID string) error

	// Subtasks
	ListSubtasks(ctx context.Context, taskID string) ([]AsanaTask, error)
	CreateSubtask(ctx context.Context, parentTaskID string, input CreateAsanaTaskInput) (*AsanaTask, error)

	// Stories (Comments)
	ListStories(ctx context.Context, taskID string) ([]AsanaStory, error)
	AddComment(ctx context.Context, taskID, text string) (*AsanaStory, error)

	// Search
	SearchTasks(ctx context.Context, workspaceID, query string) ([]AsanaTask, error)
}

// AsanaTask represents an Asana task.
type AsanaTask struct {
	GID          string            `json:"gid"`
	Name         string            `json:"name"`
	Notes        string            `json:"notes,omitempty"`
	Completed    bool              `json:"completed"`
	DueOn        string            `json:"due_on,omitempty"`
	DueAt        string            `json:"due_at,omitempty"`
	StartOn      string            `json:"start_on,omitempty"`
	Assignee     *AsanaUser        `json:"assignee,omitempty"`
	Projects     []AsanaProject    `json:"projects,omitempty"`
	Tags         []AsanaTag        `json:"tags,omitempty"`
	Parent       *AsanaTask        `json:"parent,omitempty"`
	Workspace    *AsanaWorkspace   `json:"workspace,omitempty"`
	Memberships  []AsanaMembership `json:"memberships,omitempty"`
	PermalinkURL string            `json:"permalink_url"`
	CreatedAt    time.Time         `json:"created_at"`
	ModifiedAt   time.Time         `json:"modified_at"`
}

// AsanaProject represents an Asana project.
type AsanaProject struct {
	GID          string          `json:"gid"`
	Name         string          `json:"name"`
	Notes        string          `json:"notes,omitempty"`
	Color        string          `json:"color,omitempty"`
	Archived     bool            `json:"archived"`
	Public       bool            `json:"public"`
	DueOn        string          `json:"due_on,omitempty"`
	StartOn      string          `json:"start_on,omitempty"`
	Workspace    *AsanaWorkspace `json:"workspace,omitempty"`
	Team         *AsanaTeam      `json:"team,omitempty"`
	PermalinkURL string          `json:"permalink_url"`
	CreatedAt    time.Time       `json:"created_at"`
}

// AsanaSection represents a project section.
type AsanaSection struct {
	GID       string        `json:"gid"`
	Name      string        `json:"name"`
	Project   *AsanaProject `json:"project,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// AsanaMembership represents a task's project membership.
type AsanaMembership struct {
	Project *AsanaProject `json:"project,omitempty"`
	Section *AsanaSection `json:"section,omitempty"`
}

// AsanaWorkspace represents an Asana workspace.
type AsanaWorkspace struct {
	GID            string `json:"gid"`
	Name           string `json:"name"`
	IsOrganization bool   `json:"is_organization"`
}

// AsanaTeam represents an Asana team.
type AsanaTeam struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

// AsanaUser represents an Asana user.
type AsanaUser struct {
	GID        string           `json:"gid"`
	Name       string           `json:"name"`
	Email      string           `json:"email,omitempty"`
	Photo      *AsanaPhoto      `json:"photo,omitempty"`
	Workspaces []AsanaWorkspace `json:"workspaces,omitempty"`
}

// AsanaPhoto represents a user's photo.
type AsanaPhoto struct {
	Image128x128 string `json:"image_128x128,omitempty"`
}

// AsanaTag represents an Asana tag.
type AsanaTag struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// AsanaStory represents a story (comment/activity) on a task.
type AsanaStory struct {
	GID       string     `json:"gid"`
	Type      string     `json:"type"` // comment, system
	Text      string     `json:"text,omitempty"`
	CreatedBy *AsanaUser `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// AsanaTaskFilter filters tasks.
type AsanaTaskFilter struct {
	WorkspaceID   string `json:"workspace,omitempty"`
	ProjectID     string `json:"project,omitempty"`
	SectionID     string `json:"section,omitempty"`
	AssigneeID    string `json:"assignee,omitempty"`
	Completed     *bool  `json:"completed,omitempty"`
	ModifiedSince string `json:"modified_since,omitempty"`
}

// CreateAsanaTaskInput represents input for creating a task.
type CreateAsanaTaskInput struct {
	Name        string   `json:"name"`
	Notes       string   `json:"notes,omitempty"`
	WorkspaceID string   `json:"workspace,omitempty"`
	Projects    []string `json:"projects,omitempty"`
	AssigneeID  string   `json:"assignee,omitempty"`
	DueOn       string   `json:"due_on,omitempty"`
	StartOn     string   `json:"start_on,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateAsanaTaskInput represents input for updating a task.
type UpdateAsanaTaskInput struct {
	Name       *string `json:"name,omitempty"`
	Notes      *string `json:"notes,omitempty"`
	Completed  *bool   `json:"completed,omitempty"`
	AssigneeID *string `json:"assignee,omitempty"`
	DueOn      *string `json:"due_on,omitempty"`
	StartOn    *string `json:"start_on,omitempty"`
}

// CreateAsanaProjectInput represents input for creating a project.
type CreateAsanaProjectInput struct {
	Name        string `json:"name"`
	Notes       string `json:"notes,omitempty"`
	WorkspaceID string `json:"workspace"`
	TeamID      string `json:"team,omitempty"`
	Color       string `json:"color,omitempty"`
	Public      bool   `json:"public"`
}

// NewAsanaAdapter creates a new Asana adapter.
func NewAsanaAdapter(config AsanaConfig, client AsanaClient) *AsanaAdapter {
	return &AsanaAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *AsanaAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "asana",
		Version:     "1.0.0",
		Description: "Asana project and task management",
		Capabilities: []string{
			"tasks",
			"projects",
			"sections",
			"tags",
			"subtasks",
			"comments",
			"search",
		},
	}
}

// ListTools returns available tools.
func (a *AsanaAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "asana_get_task",
			Description: "Get an Asana task by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "asana_list_tasks",
			Description: "List tasks with optional filters",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by project GID",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by assignee GID (use 'me' for current user)",
					},
					"completed": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter by completion status",
					},
				},
			},
		},
		{
			Name:        "asana_create_task",
			Description: "Create a new Asana task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Task name",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Task description",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project GID to add task to",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Assignee GID (use 'me' for current user)",
					},
					"due_on": map[string]interface{}{
						"type":        "string",
						"description": "Due date (YYYY-MM-DD)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "asana_update_task",
			Description: "Update an existing task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "New name",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "New notes/description",
					},
					"completed": map[string]interface{}{
						"type":        "boolean",
						"description": "Completion status",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "New assignee GID",
					},
					"due_on": map[string]interface{}{
						"type":        "string",
						"description": "New due date (YYYY-MM-DD)",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "asana_complete_task",
			Description: "Mark a task as complete",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "asana_delete_task",
			Description: "Delete a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "asana_list_projects",
			Description: "List projects in a workspace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspace_id": map[string]interface{}{
						"type":        "string",
						"description": "Workspace GID",
					},
				},
			},
		},
		{
			Name:        "asana_get_project",
			Description: "Get a project by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project GID",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "asana_create_project",
			Description: "Create a new project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Project name",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Project description",
					},
					"workspace_id": map[string]interface{}{
						"type":        "string",
						"description": "Workspace GID",
					},
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team GID",
					},
					"color": map[string]interface{}{
						"type":        "string",
						"description": "Project color",
					},
					"public": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether project is public",
						"default":     false,
					},
				},
				"required": []string{"name", "workspace_id"},
			},
		},
		{
			Name:        "asana_list_sections",
			Description: "List sections in a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project GID",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "asana_create_section",
			Description: "Create a section in a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project GID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Section name",
					},
				},
				"required": []string{"project_id", "name"},
			},
		},
		{
			Name:        "asana_move_task_to_section",
			Description: "Move a task to a section",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
					"section_id": map[string]interface{}{
						"type":        "string",
						"description": "Section GID",
					},
				},
				"required": []string{"task_id", "section_id"},
			},
		},
		{
			Name:        "asana_list_workspaces",
			Description: "List accessible workspaces",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "asana_list_users",
			Description: "List users in a workspace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspace_id": map[string]interface{}{
						"type":        "string",
						"description": "Workspace GID",
					},
				},
			},
		},
		{
			Name:        "asana_get_me",
			Description: "Get current user info",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "asana_add_comment",
			Description: "Add a comment to a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Comment text",
					},
				},
				"required": []string{"task_id", "text"},
			},
		},
		{
			Name:        "asana_list_subtasks",
			Description: "List subtasks of a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Parent task GID",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "asana_create_subtask",
			Description: "Create a subtask",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parent_task_id": map[string]interface{}{
						"type":        "string",
						"description": "Parent task GID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Subtask name",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Subtask notes",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Assignee GID",
					},
					"due_on": map[string]interface{}{
						"type":        "string",
						"description": "Due date (YYYY-MM-DD)",
					},
				},
				"required": []string{"parent_task_id", "name"},
			},
		},
		{
			Name:        "asana_search_tasks",
			Description: "Search for tasks",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"workspace_id": map[string]interface{}{
						"type":        "string",
						"description": "Workspace GID",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "asana_add_tag_to_task",
			Description: "Add a tag to a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task GID",
					},
					"tag_id": map[string]interface{}{
						"type":        "string",
						"description": "Tag GID",
					},
				},
				"required": []string{"task_id", "tag_id"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *AsanaAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "asana_get_task":
		return a.getTask(ctx, args)
	case "asana_list_tasks":
		return a.listTasks(ctx, args)
	case "asana_create_task":
		return a.createTask(ctx, args)
	case "asana_update_task":
		return a.updateTask(ctx, args)
	case "asana_complete_task":
		return a.completeTask(ctx, args)
	case "asana_delete_task":
		return a.deleteTask(ctx, args)
	case "asana_list_projects":
		return a.listProjects(ctx, args)
	case "asana_get_project":
		return a.getProject(ctx, args)
	case "asana_create_project":
		return a.createProject(ctx, args)
	case "asana_list_sections":
		return a.listSections(ctx, args)
	case "asana_create_section":
		return a.createSection(ctx, args)
	case "asana_move_task_to_section":
		return a.moveTaskToSection(ctx, args)
	case "asana_list_workspaces":
		return a.listWorkspaces(ctx)
	case "asana_list_users":
		return a.listUsers(ctx, args)
	case "asana_get_me":
		return a.getMe(ctx)
	case "asana_add_comment":
		return a.addComment(ctx, args)
	case "asana_list_subtasks":
		return a.listSubtasks(ctx, args)
	case "asana_create_subtask":
		return a.createSubtask(ctx, args)
	case "asana_search_tasks":
		return a.searchTasks(ctx, args)
	case "asana_add_tag_to_task":
		return a.addTagToTask(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *AsanaAdapter) getTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")

	task, err := a.client.GetTask(ctx, taskID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: formatAsanaTask(task)}},
	}, nil
}

func (a *AsanaAdapter) listTasks(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filter := AsanaTaskFilter{
		ProjectID:  getStringArg(args, "project_id", ""),
		AssigneeID: getStringArg(args, "assignee_id", ""),
	}

	if completed, ok := args["completed"].(bool); ok {
		filter.Completed = &completed
	}

	tasks, err := a.client.ListTasks(ctx, filter)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tasks:\n\n", len(tasks)))

	for _, task := range tasks {
		sb.WriteString(formatAsanaTaskLine(task))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) createTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	input := CreateAsanaTaskInput{
		Name:       getStringArg(args, "name", ""),
		Notes:      getStringArg(args, "notes", ""),
		AssigneeID: getStringArg(args, "assignee_id", ""),
		DueOn:      getStringArg(args, "due_on", ""),
	}

	if projectID := getStringArg(args, "project_id", ""); projectID != "" {
		input.Projects = []string{projectID}
	}

	task, err := a.client.CreateTask(ctx, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created task: %s\n%s", task.Name, task.PermalinkURL)}},
	}, nil
}

func (a *AsanaAdapter) updateTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")
	input := UpdateAsanaTaskInput{}

	if name, ok := args["name"].(string); ok {
		input.Name = &name
	}
	if notes, ok := args["notes"].(string); ok {
		input.Notes = &notes
	}
	if completed, ok := args["completed"].(bool); ok {
		input.Completed = &completed
	}
	if assigneeID, ok := args["assignee_id"].(string); ok {
		input.AssigneeID = &assigneeID
	}
	if dueOn, ok := args["due_on"].(string); ok {
		input.DueOn = &dueOn
	}

	task, err := a.client.UpdateTask(ctx, taskID, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Updated task: %s", task.Name)}},
	}, nil
}

func (a *AsanaAdapter) completeTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")

	err := a.client.CompleteTask(ctx, taskID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Task marked as complete"}},
	}, nil
}

func (a *AsanaAdapter) deleteTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")

	err := a.client.DeleteTask(ctx, taskID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Task deleted"}},
	}, nil
}

func (a *AsanaAdapter) listProjects(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	workspaceID := getStringArg(args, "workspace_id", a.config.WorkspaceID)

	projects, err := a.client.ListProjects(ctx, workspaceID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d projects:\n\n", len(projects)))

	for _, project := range projects {
		status := "active"
		if project.Archived {
			status = "archived"
		}
		sb.WriteString(fmt.Sprintf("- %s (%s) [%s]\n", project.Name, project.GID, status))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) getProject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID := getStringArg(args, "project_id", "")

	project, err := a.client.GetProject(ctx, projectID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: formatAsanaProject(project)}},
	}, nil
}

func (a *AsanaAdapter) createProject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	input := CreateAsanaProjectInput{
		Name:        getStringArg(args, "name", ""),
		Notes:       getStringArg(args, "notes", ""),
		WorkspaceID: getStringArg(args, "workspace_id", a.config.WorkspaceID),
		TeamID:      getStringArg(args, "team_id", ""),
		Color:       getStringArg(args, "color", ""),
	}

	if public, ok := args["public"].(bool); ok {
		input.Public = public
	}

	project, err := a.client.CreateProject(ctx, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created project: %s\n%s", project.Name, project.PermalinkURL)}},
	}, nil
}

func (a *AsanaAdapter) listSections(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID := getStringArg(args, "project_id", "")

	sections, err := a.client.ListSections(ctx, projectID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d sections:\n\n", len(sections)))

	for _, section := range sections {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", section.Name, section.GID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) createSection(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectID := getStringArg(args, "project_id", "")
	name := getStringArg(args, "name", "")

	section, err := a.client.CreateSection(ctx, projectID, name)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created section: %s (%s)", section.Name, section.GID)}},
	}, nil
}

func (a *AsanaAdapter) moveTaskToSection(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")
	sectionID := getStringArg(args, "section_id", "")

	err := a.client.AddTaskToSection(ctx, sectionID, taskID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Task moved to section"}},
	}, nil
}

func (a *AsanaAdapter) listWorkspaces(ctx context.Context) (*ToolResult, error) {
	workspaces, err := a.client.ListWorkspaces(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d workspaces:\n\n", len(workspaces)))

	for _, ws := range workspaces {
		wsType := "workspace"
		if ws.IsOrganization {
			wsType = "organization"
		}
		sb.WriteString(fmt.Sprintf("- %s (%s) [%s]\n", ws.Name, ws.GID, wsType))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) listUsers(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	workspaceID := getStringArg(args, "workspace_id", a.config.WorkspaceID)

	users, err := a.client.ListUsers(ctx, workspaceID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d users:\n\n", len(users)))

	for _, user := range users {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", user.Name, user.GID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) getMe(ctx context.Context) (*ToolResult, error) {
	user, err := a.client.GetMe(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Current user: %s (%s)\nEmail: %s",
			user.Name, user.GID, user.Email)}},
	}, nil
}

func (a *AsanaAdapter) addComment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")
	text := getStringArg(args, "text", "")

	story, err := a.client.AddComment(ctx, taskID, text)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Added comment (GID: %s)", story.GID)}},
	}, nil
}

func (a *AsanaAdapter) listSubtasks(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")

	subtasks, err := a.client.ListSubtasks(ctx, taskID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d subtasks:\n\n", len(subtasks)))

	for _, task := range subtasks {
		sb.WriteString(formatAsanaTaskLine(task))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) createSubtask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	parentTaskID := getStringArg(args, "parent_task_id", "")
	input := CreateAsanaTaskInput{
		Name:       getStringArg(args, "name", ""),
		Notes:      getStringArg(args, "notes", ""),
		AssigneeID: getStringArg(args, "assignee_id", ""),
		DueOn:      getStringArg(args, "due_on", ""),
	}

	task, err := a.client.CreateSubtask(ctx, parentTaskID, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created subtask: %s (%s)", task.Name, task.GID)}},
	}, nil
}

func (a *AsanaAdapter) searchTasks(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := getStringArg(args, "query", "")
	workspaceID := getStringArg(args, "workspace_id", a.config.WorkspaceID)

	tasks, err := a.client.SearchTasks(ctx, workspaceID, query)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tasks matching '%s':\n\n", len(tasks), query))

	for _, task := range tasks {
		sb.WriteString(formatAsanaTaskLine(task))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AsanaAdapter) addTagToTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID := getStringArg(args, "task_id", "")
	tagID := getStringArg(args, "tag_id", "")

	err := a.client.AddTagToTask(ctx, taskID, tagID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Tag added to task"}},
	}, nil
}

// Helper functions

func formatAsanaTask(task *AsanaTask) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n", task.Name))
	sb.WriteString(fmt.Sprintf("GID: %s\n", task.GID))
	sb.WriteString(fmt.Sprintf("URL: %s\n", task.PermalinkURL))

	status := "incomplete"
	if task.Completed {
		status = "complete"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	if task.Assignee != nil {
		sb.WriteString(fmt.Sprintf("Assignee: %s\n", task.Assignee.Name))
	}
	if task.DueOn != "" {
		sb.WriteString(fmt.Sprintf("Due: %s\n", task.DueOn))
	}
	if len(task.Projects) > 0 {
		projectNames := make([]string, len(task.Projects))
		for i, p := range task.Projects {
			projectNames[i] = p.Name
		}
		sb.WriteString(fmt.Sprintf("Projects: %s\n", strings.Join(projectNames, ", ")))
	}
	if len(task.Tags) > 0 {
		tagNames := make([]string, len(task.Tags))
		for i, t := range task.Tags {
			tagNames[i] = t.Name
		}
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(tagNames, ", ")))
	}

	if task.Notes != "" {
		sb.WriteString(fmt.Sprintf("\n---\n%s\n", task.Notes))
	}

	return sb.String()
}

func formatAsanaTaskLine(task AsanaTask) string {
	status := "[ ]"
	if task.Completed {
		status = "[x]"
	}
	assignee := "Unassigned"
	if task.Assignee != nil {
		assignee = task.Assignee.Name
	}
	due := ""
	if task.DueOn != "" {
		due = fmt.Sprintf(" (Due: %s)", task.DueOn)
	}
	return fmt.Sprintf("%s %s - %s%s\n", status, task.Name, assignee, due)
}

func formatAsanaProject(project *AsanaProject) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n", project.Name))
	sb.WriteString(fmt.Sprintf("GID: %s\n", project.GID))
	sb.WriteString(fmt.Sprintf("URL: %s\n", project.PermalinkURL))

	status := "active"
	if project.Archived {
		status = "archived"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	if project.DueOn != "" {
		sb.WriteString(fmt.Sprintf("Due: %s\n", project.DueOn))
	}
	if project.StartOn != "" {
		sb.WriteString(fmt.Sprintf("Start: %s\n", project.StartOn))
	}

	if project.Notes != "" {
		sb.WriteString(fmt.Sprintf("\n---\n%s\n", project.Notes))
	}

	return sb.String()
}
