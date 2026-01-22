// Package adapters provides MCP server adapters.
// This file implements the Jira MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// JiraConfig configures the Jira adapter.
type JiraConfig struct {
	BaseURL   string        `json:"base_url"` // e.g., "https://your-domain.atlassian.net"
	Email     string        `json:"email"`
	APIToken  string        `json:"api_token"`
	Timeout   time.Duration `json:"timeout"`
	ProjectKey string       `json:"project_key,omitempty"`
}

// DefaultJiraConfig returns default configuration.
func DefaultJiraConfig() JiraConfig {
	return JiraConfig{
		Timeout: 30 * time.Second,
	}
}

// JiraAdapter implements the Jira MCP server.
type JiraAdapter struct {
	config JiraConfig
	client JiraClient
}

// JiraClient interface for Jira operations.
type JiraClient interface {
	// Issues
	GetIssue(ctx context.Context, issueKey string) (*JiraIssue, error)
	SearchIssues(ctx context.Context, jql string, maxResults int) ([]JiraIssue, error)
	CreateIssue(ctx context.Context, input CreateJiraIssueInput) (*JiraIssue, error)
	UpdateIssue(ctx context.Context, issueKey string, input UpdateJiraIssueInput) (*JiraIssue, error)
	DeleteIssue(ctx context.Context, issueKey string) error
	AssignIssue(ctx context.Context, issueKey, accountID string) error

	// Transitions
	GetTransitions(ctx context.Context, issueKey string) ([]JiraTransition, error)
	TransitionIssue(ctx context.Context, issueKey, transitionID string) error

	// Comments
	ListComments(ctx context.Context, issueKey string) ([]JiraComment, error)
	AddComment(ctx context.Context, issueKey, body string) (*JiraComment, error)

	// Projects
	ListProjects(ctx context.Context) ([]JiraProject, error)
	GetProject(ctx context.Context, projectKey string) (*JiraProject, error)

	// Sprints
	ListSprints(ctx context.Context, boardID int) ([]JiraSprint, error)
	GetActiveSprint(ctx context.Context, boardID int) (*JiraSprint, error)

	// Boards
	ListBoards(ctx context.Context, projectKey string) ([]JiraBoard, error)

	// Users
	SearchUsers(ctx context.Context, query string) ([]JiraUser, error)
	GetCurrentUser(ctx context.Context) (*JiraUser, error)

	// Issue Types
	ListIssueTypes(ctx context.Context, projectKey string) ([]JiraIssueType, error)

	// Priorities
	ListPriorities(ctx context.Context) ([]JiraPriority, error)

	// Statuses
	ListStatuses(ctx context.Context, projectKey string) ([]JiraStatus, error)

	// Components
	ListComponents(ctx context.Context, projectKey string) ([]JiraComponent, error)

	// Versions
	ListVersions(ctx context.Context, projectKey string) ([]JiraVersion, error)

	// Worklogs
	ListWorklogs(ctx context.Context, issueKey string) ([]JiraWorklog, error)
	AddWorklog(ctx context.Context, issueKey string, input AddWorklogInput) (*JiraWorklog, error)

	// Attachments
	ListAttachments(ctx context.Context, issueKey string) ([]JiraAttachment, error)

	// Watchers
	AddWatcher(ctx context.Context, issueKey, accountID string) error
	RemoveWatcher(ctx context.Context, issueKey, accountID string) error
}

// JiraIssue represents a Jira issue.
type JiraIssue struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	Self        string                 `json:"self"`
	Fields      JiraIssueFields        `json:"fields"`
}

// JiraIssueFields contains issue field data.
type JiraIssueFields struct {
	Summary     string          `json:"summary"`
	Description string          `json:"description,omitempty"`
	Status      *JiraStatus     `json:"status,omitempty"`
	Priority    *JiraPriority   `json:"priority,omitempty"`
	IssueType   *JiraIssueType  `json:"issuetype,omitempty"`
	Assignee    *JiraUser       `json:"assignee,omitempty"`
	Reporter    *JiraUser       `json:"reporter,omitempty"`
	Project     *JiraProject    `json:"project,omitempty"`
	Created     string          `json:"created,omitempty"`
	Updated     string          `json:"updated,omitempty"`
	DueDate     string          `json:"duedate,omitempty"`
	Labels      []string        `json:"labels,omitempty"`
	Components  []JiraComponent `json:"components,omitempty"`
	FixVersions []JiraVersion   `json:"fixVersions,omitempty"`
	Sprint      *JiraSprint     `json:"sprint,omitempty"`
	StoryPoints *float64        `json:"customfield_10016,omitempty"` // Story points (varies by instance)
	Parent      *JiraIssue      `json:"parent,omitempty"`
}

// JiraProject represents a Jira project.
type JiraProject struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Lead        *JiraUser `json:"lead,omitempty"`
	URL         string `json:"url,omitempty"`
	ProjectType string `json:"projectTypeKey,omitempty"`
}

// JiraUser represents a Jira user.
type JiraUser struct {
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	AvatarURL    string `json:"avatarUrls,omitempty"`
}

// JiraStatus represents an issue status.
type JiraStatus struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Category    *JiraStatusCategory `json:"statusCategory,omitempty"`
}

// JiraStatusCategory represents a status category.
type JiraStatusCategory struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// JiraPriority represents an issue priority.
type JiraPriority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraIssueType represents an issue type.
type JiraIssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask"`
}

// JiraComponent represents a project component.
type JiraComponent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// JiraVersion represents a project version.
type JiraVersion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// JiraSprint represents a sprint.
type JiraSprint struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"` // active, closed, future
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Goal      string `json:"goal,omitempty"`
}

// JiraBoard represents an agile board.
type JiraBoard struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // scrum, kanban
	Location *struct {
		ProjectKey string `json:"projectKey,omitempty"`
	} `json:"location,omitempty"`
}

// JiraTransition represents an issue transition.
type JiraTransition struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	To   *JiraStatus `json:"to,omitempty"`
}

// JiraComment represents a comment.
type JiraComment struct {
	ID      string    `json:"id"`
	Body    string    `json:"body"`
	Author  *JiraUser `json:"author,omitempty"`
	Created string    `json:"created,omitempty"`
	Updated string    `json:"updated,omitempty"`
}

// JiraWorklog represents a worklog entry.
type JiraWorklog struct {
	ID             string    `json:"id"`
	Author         *JiraUser `json:"author,omitempty"`
	TimeSpent      string    `json:"timeSpent"`
	TimeSpentSecs  int       `json:"timeSpentSeconds"`
	Comment        string    `json:"comment,omitempty"`
	Started        string    `json:"started"`
}

// JiraAttachment represents an attachment.
type JiraAttachment struct {
	ID       string    `json:"id"`
	Filename string    `json:"filename"`
	Author   *JiraUser `json:"author,omitempty"`
	Created  string    `json:"created"`
	Size     int       `json:"size"`
	MimeType string    `json:"mimeType"`
	Content  string    `json:"content"` // URL
}

// CreateJiraIssueInput represents input for creating an issue.
type CreateJiraIssueInput struct {
	ProjectKey  string   `json:"project_key"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	IssueType   string   `json:"issue_type"`
	Priority    string   `json:"priority,omitempty"`
	AssigneeID  string   `json:"assignee_id,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Components  []string `json:"components,omitempty"`
	ParentKey   string   `json:"parent_key,omitempty"` // For subtasks
	DueDate     string   `json:"due_date,omitempty"`
}

// UpdateJiraIssueInput represents input for updating an issue.
type UpdateJiraIssueInput struct {
	Summary     *string  `json:"summary,omitempty"`
	Description *string  `json:"description,omitempty"`
	Priority    *string  `json:"priority,omitempty"`
	AssigneeID  *string  `json:"assignee_id,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	DueDate     *string  `json:"due_date,omitempty"`
}

// AddWorklogInput represents input for adding a worklog.
type AddWorklogInput struct {
	TimeSpent string `json:"time_spent"` // e.g., "2h 30m"
	Comment   string `json:"comment,omitempty"`
	Started   string `json:"started,omitempty"` // ISO datetime
}

// NewJiraAdapter creates a new Jira adapter.
func NewJiraAdapter(config JiraConfig, client JiraClient) *JiraAdapter {
	return &JiraAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *JiraAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "jira",
		Version:     "1.0.0",
		Description: "Jira issue tracking and project management",
		Capabilities: []string{
			"issues",
			"projects",
			"sprints",
			"boards",
			"comments",
			"transitions",
			"worklogs",
			"search",
		},
	}
}

// ListTools returns available tools.
func (a *JiraAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "jira_get_issue",
			Description: "Get a Jira issue by key",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key (e.g., 'PROJ-123')",
					},
				},
				"required": []string{"issue_key"},
			},
		},
		{
			Name:        "jira_search_issues",
			Description: "Search issues using JQL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"jql": map[string]interface{}{
						"type":        "string",
						"description": "JQL query (e.g., 'project = PROJ AND status = Open')",
					},
					"max_results": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum results to return",
						"default":     25,
					},
				},
				"required": []string{"jql"},
			},
		},
		{
			Name:        "jira_create_issue",
			Description: "Create a new Jira issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_key": map[string]interface{}{
						"type":        "string",
						"description": "Project key",
					},
					"summary": map[string]interface{}{
						"type":        "string",
						"description": "Issue summary",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Issue description",
					},
					"issue_type": map[string]interface{}{
						"type":        "string",
						"description": "Issue type (e.g., 'Bug', 'Story', 'Task')",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Priority (e.g., 'High', 'Medium', 'Low')",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Assignee account ID",
					},
					"labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Labels to add",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"description": "Due date (YYYY-MM-DD)",
					},
				},
				"required": []string{"project_key", "summary", "issue_type"},
			},
		},
		{
			Name:        "jira_update_issue",
			Description: "Update an existing issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
					"summary": map[string]interface{}{
						"type":        "string",
						"description": "New summary",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "New priority",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "New assignee account ID",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"description": "New due date (YYYY-MM-DD)",
					},
				},
				"required": []string{"issue_key"},
			},
		},
		{
			Name:        "jira_transition_issue",
			Description: "Transition an issue to a new status",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
					"transition_id": map[string]interface{}{
						"type":        "string",
						"description": "Transition ID (use jira_get_transitions to find available transitions)",
					},
				},
				"required": []string{"issue_key", "transition_id"},
			},
		},
		{
			Name:        "jira_get_transitions",
			Description: "Get available transitions for an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
				},
				"required": []string{"issue_key"},
			},
		},
		{
			Name:        "jira_assign_issue",
			Description: "Assign an issue to a user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "User account ID (use -1 to unassign)",
					},
				},
				"required": []string{"issue_key", "account_id"},
			},
		},
		{
			Name:        "jira_add_comment",
			Description: "Add a comment to an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Comment body",
					},
				},
				"required": []string{"issue_key", "body"},
			},
		},
		{
			Name:        "jira_list_comments",
			Description: "List comments on an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
				},
				"required": []string{"issue_key"},
			},
		},
		{
			Name:        "jira_list_projects",
			Description: "List accessible projects",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "jira_get_project",
			Description: "Get project details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_key": map[string]interface{}{
						"type":        "string",
						"description": "Project key",
					},
				},
				"required": []string{"project_key"},
			},
		},
		{
			Name:        "jira_list_sprints",
			Description: "List sprints for a board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "integer",
						"description": "Board ID",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "jira_list_boards",
			Description: "List agile boards",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_key": map[string]interface{}{
						"type":        "string",
						"description": "Filter by project key",
					},
				},
			},
		},
		{
			Name:        "jira_search_users",
			Description: "Search for users",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (name or email)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "jira_get_me",
			Description: "Get current user info",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "jira_list_issue_types",
			Description: "List issue types for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_key": map[string]interface{}{
						"type":        "string",
						"description": "Project key",
					},
				},
				"required": []string{"project_key"},
			},
		},
		{
			Name:        "jira_list_priorities",
			Description: "List available priorities",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "jira_list_statuses",
			Description: "List statuses for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_key": map[string]interface{}{
						"type":        "string",
						"description": "Project key",
					},
				},
				"required": []string{"project_key"},
			},
		},
		{
			Name:        "jira_add_worklog",
			Description: "Log work on an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
					"time_spent": map[string]interface{}{
						"type":        "string",
						"description": "Time spent (e.g., '2h 30m')",
					},
					"comment": map[string]interface{}{
						"type":        "string",
						"description": "Work description",
					},
				},
				"required": []string{"issue_key", "time_spent"},
			},
		},
		{
			Name:        "jira_add_watcher",
			Description: "Add a watcher to an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_key": map[string]interface{}{
						"type":        "string",
						"description": "Issue key",
					},
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "User account ID",
					},
				},
				"required": []string{"issue_key", "account_id"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *JiraAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "jira_get_issue":
		return a.getIssue(ctx, args)
	case "jira_search_issues":
		return a.searchIssues(ctx, args)
	case "jira_create_issue":
		return a.createIssue(ctx, args)
	case "jira_update_issue":
		return a.updateIssue(ctx, args)
	case "jira_transition_issue":
		return a.transitionIssue(ctx, args)
	case "jira_get_transitions":
		return a.getTransitions(ctx, args)
	case "jira_assign_issue":
		return a.assignIssue(ctx, args)
	case "jira_add_comment":
		return a.addComment(ctx, args)
	case "jira_list_comments":
		return a.listComments(ctx, args)
	case "jira_list_projects":
		return a.listProjects(ctx)
	case "jira_get_project":
		return a.getProject(ctx, args)
	case "jira_list_sprints":
		return a.listSprints(ctx, args)
	case "jira_list_boards":
		return a.listBoards(ctx, args)
	case "jira_search_users":
		return a.searchUsers(ctx, args)
	case "jira_get_me":
		return a.getMe(ctx)
	case "jira_list_issue_types":
		return a.listIssueTypes(ctx, args)
	case "jira_list_priorities":
		return a.listPriorities(ctx)
	case "jira_list_statuses":
		return a.listStatuses(ctx, args)
	case "jira_add_worklog":
		return a.addWorklog(ctx, args)
	case "jira_add_watcher":
		return a.addWatcher(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *JiraAdapter) getIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")

	issue, err := a.client.GetIssue(ctx, issueKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: formatJiraIssue(issue, a.config.BaseURL)}},
	}, nil
}

func (a *JiraAdapter) searchIssues(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	jql := getStringArg(args, "jql", "")
	maxResults := getIntArg(args, "max_results", 25)

	issues, err := a.client.SearchIssues(ctx, jql, maxResults)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issues:\n\n", len(issues)))

	for _, issue := range issues {
		sb.WriteString(formatJiraIssueLine(issue, a.config.BaseURL))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) createIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	input := CreateJiraIssueInput{
		ProjectKey:  getStringArg(args, "project_key", a.config.ProjectKey),
		Summary:     getStringArg(args, "summary", ""),
		Description: getStringArg(args, "description", ""),
		IssueType:   getStringArg(args, "issue_type", "Task"),
		Priority:    getStringArg(args, "priority", ""),
		AssigneeID:  getStringArg(args, "assignee_id", ""),
		DueDate:     getStringArg(args, "due_date", ""),
	}

	if labelsRaw, ok := args["labels"].([]interface{}); ok {
		for _, l := range labelsRaw {
			if s, ok := l.(string); ok {
				input.Labels = append(input.Labels, s)
			}
		}
	}

	issue, err := a.client.CreateIssue(ctx, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	url := fmt.Sprintf("%s/browse/%s", a.config.BaseURL, issue.Key)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created issue %s: %s\n%s", issue.Key, issue.Fields.Summary, url)}},
	}, nil
}

func (a *JiraAdapter) updateIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")
	input := UpdateJiraIssueInput{}

	if summary, ok := args["summary"].(string); ok {
		input.Summary = &summary
	}
	if description, ok := args["description"].(string); ok {
		input.Description = &description
	}
	if priority, ok := args["priority"].(string); ok {
		input.Priority = &priority
	}
	if assigneeID, ok := args["assignee_id"].(string); ok {
		input.AssigneeID = &assigneeID
	}
	if dueDate, ok := args["due_date"].(string); ok {
		input.DueDate = &dueDate
	}

	issue, err := a.client.UpdateIssue(ctx, issueKey, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Updated issue %s", issue.Key)}},
	}, nil
}

func (a *JiraAdapter) transitionIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")
	transitionID := getStringArg(args, "transition_id", "")

	err := a.client.TransitionIssue(ctx, issueKey, transitionID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Issue %s transitioned successfully", issueKey)}},
	}, nil
}

func (a *JiraAdapter) getTransitions(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")

	transitions, err := a.client.GetTransitions(ctx, issueKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available transitions for %s:\n\n", issueKey))

	for _, t := range transitions {
		toStatus := ""
		if t.To != nil {
			toStatus = fmt.Sprintf(" -> %s", t.To.Name)
		}
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)%s\n", t.Name, t.ID, toStatus))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) assignIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")
	accountID := getStringArg(args, "account_id", "")

	err := a.client.AssignIssue(ctx, issueKey, accountID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	msg := fmt.Sprintf("Issue %s assigned", issueKey)
	if accountID == "-1" {
		msg = fmt.Sprintf("Issue %s unassigned", issueKey)
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: msg}},
	}, nil
}

func (a *JiraAdapter) addComment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")
	body := getStringArg(args, "body", "")

	comment, err := a.client.AddComment(ctx, issueKey, body)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Added comment (ID: %s)", comment.ID)}},
	}, nil
}

func (a *JiraAdapter) listComments(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")

	comments, err := a.client.ListComments(ctx, issueKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d comments:\n\n", len(comments)))

	for _, c := range comments {
		author := "Unknown"
		if c.Author != nil {
			author = c.Author.DisplayName
		}
		sb.WriteString(fmt.Sprintf("---\n**%s** (%s):\n%s\n\n", author, c.Created, c.Body))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) listProjects(ctx context.Context) (*ToolResult, error) {
	projects, err := a.client.ListProjects(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d projects:\n\n", len(projects)))

	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("- %s (%s) - %s\n", p.Name, p.Key, p.ProjectType))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) getProject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectKey := getStringArg(args, "project_key", a.config.ProjectKey)

	project, err := a.client.GetProject(ctx, projectKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: formatJiraProject(project)}},
	}, nil
}

func (a *JiraAdapter) listSprints(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID := getIntArg(args, "board_id", 0)

	sprints, err := a.client.ListSprints(ctx, boardID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d sprints:\n\n", len(sprints)))

	for _, s := range sprints {
		dates := ""
		if s.StartDate != "" && s.EndDate != "" {
			dates = fmt.Sprintf(" (%s - %s)", s.StartDate[:10], s.EndDate[:10])
		}
		sb.WriteString(fmt.Sprintf("- %s [%s]%s\n", s.Name, s.State, dates))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) listBoards(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectKey := getStringArg(args, "project_key", a.config.ProjectKey)

	boards, err := a.client.ListBoards(ctx, projectKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d boards:\n\n", len(boards)))

	for _, b := range boards {
		sb.WriteString(fmt.Sprintf("- %s (ID: %d) [%s]\n", b.Name, b.ID, b.Type))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) searchUsers(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := getStringArg(args, "query", "")

	users, err := a.client.SearchUsers(ctx, query)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d users:\n\n", len(users)))

	for _, u := range users {
		status := "active"
		if !u.Active {
			status = "inactive"
		}
		sb.WriteString(fmt.Sprintf("- %s (%s) [%s]\n", u.DisplayName, u.AccountID, status))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) getMe(ctx context.Context) (*ToolResult, error) {
	user, err := a.client.GetCurrentUser(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Current user: %s (%s)\nEmail: %s",
			user.DisplayName, user.AccountID, user.EmailAddress)}},
	}, nil
}

func (a *JiraAdapter) listIssueTypes(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectKey := getStringArg(args, "project_key", a.config.ProjectKey)

	issueTypes, err := a.client.ListIssueTypes(ctx, projectKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issue types:\n\n", len(issueTypes)))

	for _, it := range issueTypes {
		subtask := ""
		if it.Subtask {
			subtask = " [subtask]"
		}
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)%s\n", it.Name, it.ID, subtask))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) listPriorities(ctx context.Context) (*ToolResult, error) {
	priorities, err := a.client.ListPriorities(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d priorities:\n\n", len(priorities)))

	for _, p := range priorities {
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)\n", p.Name, p.ID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) listStatuses(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectKey := getStringArg(args, "project_key", a.config.ProjectKey)

	statuses, err := a.client.ListStatuses(ctx, projectKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d statuses:\n\n", len(statuses)))

	for _, s := range statuses {
		category := ""
		if s.Category != nil {
			category = fmt.Sprintf(" [%s]", s.Category.Name)
		}
		sb.WriteString(fmt.Sprintf("- %s (ID: %s)%s\n", s.Name, s.ID, category))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *JiraAdapter) addWorklog(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")
	input := AddWorklogInput{
		TimeSpent: getStringArg(args, "time_spent", ""),
		Comment:   getStringArg(args, "comment", ""),
	}

	worklog, err := a.client.AddWorklog(ctx, issueKey, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Logged %s on %s (ID: %s)", worklog.TimeSpent, issueKey, worklog.ID)}},
	}, nil
}

func (a *JiraAdapter) addWatcher(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueKey := getStringArg(args, "issue_key", "")
	accountID := getStringArg(args, "account_id", "")

	err := a.client.AddWatcher(ctx, issueKey, accountID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Added watcher to %s", issueKey)}},
	}, nil
}

// Helper functions

func formatJiraIssue(issue *JiraIssue, baseURL string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s: %s**\n", issue.Key, issue.Fields.Summary))
	sb.WriteString(fmt.Sprintf("URL: %s/browse/%s\n", baseURL, issue.Key))

	if issue.Fields.Status != nil {
		sb.WriteString(fmt.Sprintf("Status: %s\n", issue.Fields.Status.Name))
	}
	if issue.Fields.Priority != nil {
		sb.WriteString(fmt.Sprintf("Priority: %s\n", issue.Fields.Priority.Name))
	}
	if issue.Fields.IssueType != nil {
		sb.WriteString(fmt.Sprintf("Type: %s\n", issue.Fields.IssueType.Name))
	}
	if issue.Fields.Assignee != nil {
		sb.WriteString(fmt.Sprintf("Assignee: %s\n", issue.Fields.Assignee.DisplayName))
	}
	if issue.Fields.Reporter != nil {
		sb.WriteString(fmt.Sprintf("Reporter: %s\n", issue.Fields.Reporter.DisplayName))
	}
	if issue.Fields.DueDate != "" {
		sb.WriteString(fmt.Sprintf("Due: %s\n", issue.Fields.DueDate))
	}
	if len(issue.Fields.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(issue.Fields.Labels, ", ")))
	}
	if issue.Fields.Sprint != nil {
		sb.WriteString(fmt.Sprintf("Sprint: %s\n", issue.Fields.Sprint.Name))
	}

	if issue.Fields.Description != "" {
		sb.WriteString(fmt.Sprintf("\n---\n%s\n", issue.Fields.Description))
	}

	return sb.String()
}

func formatJiraIssueLine(issue JiraIssue, baseURL string) string {
	status := "?"
	if issue.Fields.Status != nil {
		status = issue.Fields.Status.Name
	}
	assignee := "Unassigned"
	if issue.Fields.Assignee != nil {
		assignee = issue.Fields.Assignee.DisplayName
	}
	url := fmt.Sprintf("%s/browse/%s", baseURL, issue.Key)
	return fmt.Sprintf("- [%s] %s: %s (%s)\n  %s\n", status, issue.Key, issue.Fields.Summary, assignee, url)
}

func formatJiraProject(project *JiraProject) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s (%s)**\n", project.Name, project.Key))
	sb.WriteString(fmt.Sprintf("Type: %s\n", project.ProjectType))

	if project.Lead != nil {
		sb.WriteString(fmt.Sprintf("Lead: %s\n", project.Lead.DisplayName))
	}

	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("\n---\n%s\n", project.Description))
	}

	return sb.String()
}
