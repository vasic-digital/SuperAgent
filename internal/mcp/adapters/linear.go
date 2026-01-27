// Package adapters provides MCP server adapters.
// This file implements the Linear MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// LinearConfig configures the Linear adapter.
type LinearConfig struct {
	APIKey  string        `json:"api_key"`
	Timeout time.Duration `json:"timeout"`
}

// DefaultLinearConfig returns default configuration.
func DefaultLinearConfig() LinearConfig {
	return LinearConfig{
		Timeout: 30 * time.Second,
	}
}

// LinearAdapter implements the Linear MCP server.
type LinearAdapter struct {
	config LinearConfig
	client LinearClient
}

// LinearClient interface for Linear operations.
type LinearClient interface {
	// Issues
	GetIssue(ctx context.Context, id string) (*LinearIssue, error)
	ListIssues(ctx context.Context, filter IssueFilter) ([]LinearIssue, error)
	CreateIssue(ctx context.Context, input CreateIssueInput) (*LinearIssue, error)
	UpdateIssue(ctx context.Context, id string, input UpdateIssueInput) (*LinearIssue, error)
	DeleteIssue(ctx context.Context, id string) error

	// Teams
	ListTeams(ctx context.Context) ([]LinearTeam, error)
	GetTeam(ctx context.Context, id string) (*LinearTeam, error)

	// Projects
	ListProjects(ctx context.Context, teamID string) ([]LinearProject, error)
	GetProject(ctx context.Context, id string) (*LinearProject, error)
	CreateProject(ctx context.Context, input CreateProjectInput) (*LinearProject, error)

	// Cycles
	ListCycles(ctx context.Context, teamID string) ([]LinearCycle, error)
	GetActiveCycle(ctx context.Context, teamID string) (*LinearCycle, error)

	// Labels
	ListLabels(ctx context.Context, teamID string) ([]LinearLabel, error)

	// Users
	ListUsers(ctx context.Context) ([]LinearUser, error)
	GetMe(ctx context.Context) (*LinearUser, error)

	// Comments
	ListComments(ctx context.Context, issueID string) ([]LinearComment, error)
	CreateComment(ctx context.Context, issueID, body string) (*LinearComment, error)

	// Workflow States
	ListWorkflowStates(ctx context.Context, teamID string) ([]LinearWorkflowState, error)

	// Search
	SearchIssues(ctx context.Context, query string, limit int) ([]LinearIssue, error)
}

// LinearIssue represents a Linear issue.
type LinearIssue struct {
	ID          string               `json:"id"`
	Identifier  string               `json:"identifier"`
	Title       string               `json:"title"`
	Description string               `json:"description,omitempty"`
	Priority    int                  `json:"priority"`
	State       *LinearWorkflowState `json:"state,omitempty"`
	Assignee    *LinearUser          `json:"assignee,omitempty"`
	Team        *LinearTeam          `json:"team,omitempty"`
	Project     *LinearProject       `json:"project,omitempty"`
	Cycle       *LinearCycle         `json:"cycle,omitempty"`
	Labels      []LinearLabel        `json:"labels,omitempty"`
	DueDate     string               `json:"dueDate,omitempty"`
	Estimate    *float64             `json:"estimate,omitempty"`
	URL         string               `json:"url"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

// LinearTeam represents a Linear team.
type LinearTeam struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description,omitempty"`
	IssueCount  int    `json:"issueCount"`
}

// LinearProject represents a Linear project.
type LinearProject struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	State       string  `json:"state"`
	Progress    float64 `json:"progress"`
	StartDate   string  `json:"startDate,omitempty"`
	TargetDate  string  `json:"targetDate,omitempty"`
	URL         string  `json:"url"`
}

// LinearCycle represents a Linear cycle (sprint).
type LinearCycle struct {
	ID         string  `json:"id"`
	Number     int     `json:"number"`
	Name       string  `json:"name,omitempty"`
	StartsAt   string  `json:"startsAt"`
	EndsAt     string  `json:"endsAt"`
	Progress   float64 `json:"progress"`
	IssueCount int     `json:"issueCountHistory,omitempty"`
}

// LinearLabel represents a Linear label.
type LinearLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// LinearUser represents a Linear user.
type LinearUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	Active      bool   `json:"active"`
}

// LinearComment represents a comment on an issue.
type LinearComment struct {
	ID        string      `json:"id"`
	Body      string      `json:"body"`
	User      *LinearUser `json:"user,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
}

// LinearWorkflowState represents a workflow state.
type LinearWorkflowState struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Type     string `json:"type"` // backlog, unstarted, started, completed, canceled
	Position int    `json:"position"`
}

// IssueFilter filters issues.
type IssueFilter struct {
	TeamID     string   `json:"teamId,omitempty"`
	ProjectID  string   `json:"projectId,omitempty"`
	CycleID    string   `json:"cycleId,omitempty"`
	AssigneeID string   `json:"assigneeId,omitempty"`
	StateIDs   []string `json:"stateIds,omitempty"`
	LabelIDs   []string `json:"labelIds,omitempty"`
	Priority   *int     `json:"priority,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// CreateIssueInput represents input for creating an issue.
type CreateIssueInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	TeamID      string   `json:"teamId"`
	StateID     string   `json:"stateId,omitempty"`
	AssigneeID  string   `json:"assigneeId,omitempty"`
	ProjectID   string   `json:"projectId,omitempty"`
	CycleID     string   `json:"cycleId,omitempty"`
	LabelIDs    []string `json:"labelIds,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Estimate    *float64 `json:"estimate,omitempty"`
	DueDate     string   `json:"dueDate,omitempty"`
}

// UpdateIssueInput represents input for updating an issue.
type UpdateIssueInput struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	StateID     *string  `json:"stateId,omitempty"`
	AssigneeID  *string  `json:"assigneeId,omitempty"`
	ProjectID   *string  `json:"projectId,omitempty"`
	CycleID     *string  `json:"cycleId,omitempty"`
	LabelIDs    []string `json:"labelIds,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Estimate    *float64 `json:"estimate,omitempty"`
	DueDate     *string  `json:"dueDate,omitempty"`
}

// CreateProjectInput represents input for creating a project.
type CreateProjectInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	TeamIDs     []string `json:"teamIds"`
	StartDate   string   `json:"startDate,omitempty"`
	TargetDate  string   `json:"targetDate,omitempty"`
}

// NewLinearAdapter creates a new Linear adapter.
func NewLinearAdapter(config LinearConfig, client LinearClient) *LinearAdapter {
	return &LinearAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *LinearAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "linear",
		Version:     "1.0.0",
		Description: "Linear issue tracking and project management",
		Capabilities: []string{
			"issues",
			"projects",
			"teams",
			"cycles",
			"labels",
			"comments",
			"search",
		},
	}
}

// ListTools returns available tools.
func (a *LinearAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "linear_get_issue",
			Description: "Get a Linear issue by ID or identifier",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Issue ID or identifier (e.g., 'ENG-123')",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "linear_list_issues",
			Description: "List issues with optional filters",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by team ID",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by project ID",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by assignee ID",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of issues to return",
						"default":     25,
					},
				},
			},
		},
		{
			Name:        "linear_create_issue",
			Description: "Create a new Linear issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Issue title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Issue description (markdown supported)",
					},
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Assignee user ID",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID",
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "Priority (0=None, 1=Urgent, 2=High, 3=Medium, 4=Low)",
					},
				},
				"required": []string{"title", "team_id"},
			},
		},
		{
			Name:        "linear_update_issue",
			Description: "Update an existing issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Issue ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description",
					},
					"state_id": map[string]interface{}{
						"type":        "string",
						"description": "New state ID",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "New assignee ID",
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "New priority",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "linear_list_teams",
			Description: "List all teams",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "linear_list_projects",
			Description: "List projects for a team",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID",
					},
				},
			},
		},
		{
			Name:        "linear_create_project",
			Description: "Create a new project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Project name",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Project description",
					},
					"team_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Team IDs to associate",
					},
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Start date (YYYY-MM-DD)",
					},
					"target_date": map[string]interface{}{
						"type":        "string",
						"description": "Target date (YYYY-MM-DD)",
					},
				},
				"required": []string{"name", "team_ids"},
			},
		},
		{
			Name:        "linear_list_cycles",
			Description: "List cycles (sprints) for a team",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID",
					},
				},
				"required": []string{"team_id"},
			},
		},
		{
			Name:        "linear_list_labels",
			Description: "List labels for a team",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID",
					},
				},
			},
		},
		{
			Name:        "linear_list_users",
			Description: "List workspace users",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "linear_get_me",
			Description: "Get current user info",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "linear_add_comment",
			Description: "Add a comment to an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_id": map[string]interface{}{
						"type":        "string",
						"description": "Issue ID",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Comment body (markdown supported)",
					},
				},
				"required": []string{"issue_id", "body"},
			},
		},
		{
			Name:        "linear_list_workflow_states",
			Description: "List workflow states for a team",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID",
					},
				},
				"required": []string{"team_id"},
			},
		},
		{
			Name:        "linear_search_issues",
			Description: "Search for issues",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results",
						"default":     20,
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *LinearAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "linear_get_issue":
		return a.getIssue(ctx, args)
	case "linear_list_issues":
		return a.listIssues(ctx, args)
	case "linear_create_issue":
		return a.createIssue(ctx, args)
	case "linear_update_issue":
		return a.updateIssue(ctx, args)
	case "linear_list_teams":
		return a.listTeams(ctx)
	case "linear_list_projects":
		return a.listProjects(ctx, args)
	case "linear_create_project":
		return a.createProject(ctx, args)
	case "linear_list_cycles":
		return a.listCycles(ctx, args)
	case "linear_list_labels":
		return a.listLabels(ctx, args)
	case "linear_list_users":
		return a.listUsers(ctx)
	case "linear_get_me":
		return a.getMe(ctx)
	case "linear_add_comment":
		return a.addComment(ctx, args)
	case "linear_list_workflow_states":
		return a.listWorkflowStates(ctx, args)
	case "linear_search_issues":
		return a.searchIssues(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *LinearAdapter) getIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)

	issue, err := a.client.GetIssue(ctx, id)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: formatLinearIssue(issue)}},
	}, nil
}

func (a *LinearAdapter) listIssues(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filter := IssueFilter{
		TeamID:     getStringArg(args, "team_id", ""),
		ProjectID:  getStringArg(args, "project_id", ""),
		AssigneeID: getStringArg(args, "assignee_id", ""),
		Limit:      getIntArg(args, "limit", 25),
	}

	issues, err := a.client.ListIssues(ctx, filter)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issues:\n\n", len(issues)))

	for _, issue := range issues {
		sb.WriteString(formatLinearIssueLine(issue))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) createIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	input := CreateIssueInput{
		Title:       getStringArg(args, "title", ""),
		Description: getStringArg(args, "description", ""),
		TeamID:      getStringArg(args, "team_id", ""),
		AssigneeID:  getStringArg(args, "assignee_id", ""),
		ProjectID:   getStringArg(args, "project_id", ""),
	}

	if p, ok := args["priority"].(float64); ok {
		priority := int(p)
		input.Priority = &priority
	}

	issue, err := a.client.CreateIssue(ctx, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created issue %s: %s\n%s", issue.Identifier, issue.Title, issue.URL)}},
	}, nil
}

func (a *LinearAdapter) updateIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)
	input := UpdateIssueInput{}

	if title, ok := args["title"].(string); ok {
		input.Title = &title
	}
	if desc, ok := args["description"].(string); ok {
		input.Description = &desc
	}
	if stateID, ok := args["state_id"].(string); ok {
		input.StateID = &stateID
	}
	if assigneeID, ok := args["assignee_id"].(string); ok {
		input.AssigneeID = &assigneeID
	}
	if p, ok := args["priority"].(float64); ok {
		priority := int(p)
		input.Priority = &priority
	}

	issue, err := a.client.UpdateIssue(ctx, id, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Updated issue %s", issue.Identifier)}},
	}, nil
}

func (a *LinearAdapter) listTeams(ctx context.Context) (*ToolResult, error) {
	teams, err := a.client.ListTeams(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d teams:\n\n", len(teams)))

	for _, team := range teams {
		sb.WriteString(fmt.Sprintf("- %s (%s) - %s [%d issues]\n", team.Name, team.Key, team.ID, team.IssueCount))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) listProjects(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	teamID := getStringArg(args, "team_id", "")

	projects, err := a.client.ListProjects(ctx, teamID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d projects:\n\n", len(projects)))

	for _, project := range projects {
		sb.WriteString(fmt.Sprintf("- %s [%s] - %.0f%% complete\n", project.Name, project.State, project.Progress*100))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) createProject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	input := CreateProjectInput{
		Name:        getStringArg(args, "name", ""),
		Description: getStringArg(args, "description", ""),
		StartDate:   getStringArg(args, "start_date", ""),
		TargetDate:  getStringArg(args, "target_date", ""),
	}

	if teamIDsRaw, ok := args["team_ids"].([]interface{}); ok {
		for _, id := range teamIDsRaw {
			if s, ok := id.(string); ok {
				input.TeamIDs = append(input.TeamIDs, s)
			}
		}
	}

	project, err := a.client.CreateProject(ctx, input)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created project: %s\n%s", project.Name, project.URL)}},
	}, nil
}

func (a *LinearAdapter) listCycles(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	teamID := getStringArg(args, "team_id", "")

	cycles, err := a.client.ListCycles(ctx, teamID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d cycles:\n\n", len(cycles)))

	for _, cycle := range cycles {
		name := cycle.Name
		if name == "" {
			name = fmt.Sprintf("Cycle %d", cycle.Number)
		}
		sb.WriteString(fmt.Sprintf("- %s (%s - %s) - %.0f%% complete\n",
			name, cycle.StartsAt[:10], cycle.EndsAt[:10], cycle.Progress*100))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) listLabels(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	teamID := getStringArg(args, "team_id", "")

	labels, err := a.client.ListLabels(ctx, teamID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d labels:\n\n", len(labels)))

	for _, label := range labels {
		sb.WriteString(fmt.Sprintf("- %s (%s) - %s\n", label.Name, label.ID, label.Color))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) listUsers(ctx context.Context) (*ToolResult, error) {
	users, err := a.client.ListUsers(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d users:\n\n", len(users)))

	for _, user := range users {
		status := "active"
		if !user.Active {
			status = "inactive"
		}
		sb.WriteString(fmt.Sprintf("- %s (%s) - %s\n", user.DisplayName, user.ID, status))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) getMe(ctx context.Context) (*ToolResult, error) {
	user, err := a.client.GetMe(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Current user: %s (%s)\nEmail: %s",
			user.DisplayName, user.ID, user.Email)}},
	}, nil
}

func (a *LinearAdapter) addComment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueID := getStringArg(args, "issue_id", "")
	body := getStringArg(args, "body", "")

	comment, err := a.client.CreateComment(ctx, issueID, body)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Added comment (ID: %s)", comment.ID)}},
	}, nil
}

func (a *LinearAdapter) listWorkflowStates(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	teamID := getStringArg(args, "team_id", "")

	states, err := a.client.ListWorkflowStates(ctx, teamID)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d workflow states:\n\n", len(states)))

	for _, state := range states {
		sb.WriteString(fmt.Sprintf("- %s (%s) - Type: %s\n", state.Name, state.ID, state.Type))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *LinearAdapter) searchIssues(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := getStringArg(args, "query", "")
	limit := getIntArg(args, "limit", 20)

	issues, err := a.client.SearchIssues(ctx, query, limit)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issues matching '%s':\n\n", len(issues), query))

	for _, issue := range issues {
		sb.WriteString(formatLinearIssueLine(issue))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

// Helper functions

func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

func formatLinearIssue(issue *LinearIssue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s: %s**\n", issue.Identifier, issue.Title))
	sb.WriteString(fmt.Sprintf("URL: %s\n", issue.URL))

	if issue.State != nil {
		sb.WriteString(fmt.Sprintf("State: %s\n", issue.State.Name))
	}
	if issue.Assignee != nil {
		sb.WriteString(fmt.Sprintf("Assignee: %s\n", issue.Assignee.DisplayName))
	}
	if issue.Team != nil {
		sb.WriteString(fmt.Sprintf("Team: %s\n", issue.Team.Name))
	}
	if issue.Project != nil {
		sb.WriteString(fmt.Sprintf("Project: %s\n", issue.Project.Name))
	}

	priorityNames := []string{"None", "Urgent", "High", "Medium", "Low"}
	if issue.Priority >= 0 && issue.Priority < len(priorityNames) {
		sb.WriteString(fmt.Sprintf("Priority: %s\n", priorityNames[issue.Priority]))
	}

	if issue.Description != "" {
		sb.WriteString(fmt.Sprintf("\n---\n%s\n", issue.Description))
	}

	return sb.String()
}

func formatLinearIssueLine(issue LinearIssue) string {
	state := "?"
	if issue.State != nil {
		state = issue.State.Name
	}
	assignee := "Unassigned"
	if issue.Assignee != nil {
		assignee = issue.Assignee.DisplayName
	}
	return fmt.Sprintf("- [%s] %s: %s (%s, %s)\n", state, issue.Identifier, issue.Title, assignee, issue.URL)
}
