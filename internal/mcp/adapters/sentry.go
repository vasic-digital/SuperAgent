// Package adapters provides MCP server adapters.
// This file implements the Sentry MCP server adapter for error monitoring.
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

// SentryConfig configures the Sentry adapter.
type SentryConfig struct {
	AuthToken    string        `json:"auth_token"`
	BaseURL      string        `json:"base_url"`
	Organization string        `json:"organization"`
	Timeout      time.Duration `json:"timeout"`
}

// DefaultSentryConfig returns default configuration.
func DefaultSentryConfig() SentryConfig {
	return SentryConfig{
		BaseURL: "https://sentry.io/api/0",
		Timeout: 30 * time.Second,
	}
}

// SentryAdapter implements the Sentry MCP server.
type SentryAdapter struct {
	config     SentryConfig
	httpClient *http.Client
}

// NewSentryAdapter creates a new Sentry adapter.
func NewSentryAdapter(config SentryConfig) *SentryAdapter {
	if config.BaseURL == "" {
		config.BaseURL = DefaultSentryConfig().BaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultSentryConfig().Timeout
	}
	return &SentryAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *SentryAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "sentry",
		Version:     "1.0.0",
		Description: "Sentry error monitoring integration for querying issues and alerts",
		Capabilities: []string{
			"list_projects",
			"list_issues",
			"get_issue",
			"resolve_issue",
			"list_events",
			"get_event",
			"list_alerts",
			"query_stats",
		},
	}
}

// ListTools returns available tools.
func (a *SentryAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "sentry_list_projects",
			Description: "List all projects in the organization",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sentry_list_issues",
			Description: "List issues for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{
						"type":        "string",
						"description": "Project slug",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (e.g., 'is:unresolved')",
						"default":     "is:unresolved",
					},
					"sort": map[string]interface{}{
						"type":        "string",
						"description": "Sort by: date, new, priority, freq, user",
						"enum":        []string{"date", "new", "priority", "freq", "user"},
						"default":     "date",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of issues to return",
						"default":     25,
					},
				},
				"required": []string{"project"},
			},
		},
		{
			Name:        "sentry_get_issue",
			Description: "Get detailed information about a specific issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_id": map[string]interface{}{
						"type":        "string",
						"description": "The issue ID",
					},
				},
				"required": []string{"issue_id"},
			},
		},
		{
			Name:        "sentry_resolve_issue",
			Description: "Resolve or unresolve an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_id": map[string]interface{}{
						"type":        "string",
						"description": "The issue ID",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "New status",
						"enum":        []string{"resolved", "unresolved", "ignored"},
					},
				},
				"required": []string{"issue_id", "status"},
			},
		},
		{
			Name:        "sentry_list_events",
			Description: "List events for an issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"issue_id": map[string]interface{}{
						"type":        "string",
						"description": "The issue ID",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of events to return",
						"default":     25,
					},
				},
				"required": []string{"issue_id"},
			},
		},
		{
			Name:        "sentry_get_event",
			Description: "Get detailed information about an event",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{
						"type":        "string",
						"description": "Project slug",
					},
					"event_id": map[string]interface{}{
						"type":        "string",
						"description": "The event ID",
					},
				},
				"required": []string{"project", "event_id"},
			},
		},
		{
			Name:        "sentry_list_alerts",
			Description: "List alert rules for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{
						"type":        "string",
						"description": "Project slug",
					},
				},
				"required": []string{"project"},
			},
		},
		{
			Name:        "sentry_query_stats",
			Description: "Query project statistics",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{
						"type":        "string",
						"description": "Project slug",
					},
					"stat": map[string]interface{}{
						"type":        "string",
						"description": "Statistic type",
						"enum":        []string{"received", "rejected", "blacklisted"},
						"default":     "received",
					},
					"resolution": map[string]interface{}{
						"type":        "string",
						"description": "Time resolution",
						"enum":        []string{"10s", "1h", "1d"},
						"default":     "1h",
					},
					"since": map[string]interface{}{
						"type":        "string",
						"description": "Start time (ISO 8601 or relative like '24h')",
						"default":     "24h",
					},
				},
				"required": []string{"project"},
			},
		},
		{
			Name:        "sentry_search_issues",
			Description: "Search issues across all projects",
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
						"default":     25,
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *SentryAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "sentry_list_projects":
		return a.listProjects(ctx, args)
	case "sentry_list_issues":
		return a.listIssues(ctx, args)
	case "sentry_get_issue":
		return a.getIssue(ctx, args)
	case "sentry_resolve_issue":
		return a.resolveIssue(ctx, args)
	case "sentry_list_events":
		return a.listEvents(ctx, args)
	case "sentry_get_event":
		return a.getEvent(ctx, args)
	case "sentry_list_alerts":
		return a.listAlerts(ctx, args)
	case "sentry_query_stats":
		return a.queryStats(ctx, args)
	case "sentry_search_issues":
		return a.searchIssues(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *SentryAdapter) listProjects(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	endpoint := fmt.Sprintf("/organizations/%s/projects/", a.config.Organization)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var projects []SentryProject
	if err := json.Unmarshal(resp, &projects); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Projects (%d):\n\n", len(projects)))

	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, p.Slug))
		sb.WriteString(fmt.Sprintf("  ID: %s\n", p.ID))
		sb.WriteString(fmt.Sprintf("  Platform: %s\n", p.Platform))
		sb.WriteString(fmt.Sprintf("  Status: %s\n\n", p.Status))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) listIssues(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	project, _ := args["project"].(string)
	query, _ := args["query"].(string)
	if query == "" {
		query = "is:unresolved"
	}
	sort, _ := args["sort"].(string)
	if sort == "" {
		sort = "date"
	}
	limit := getIntArg(args, "limit", 25)

	params := url.Values{}
	params.Set("query", query)
	params.Set("sort", sort)
	params.Set("limit", fmt.Sprintf("%d", limit))

	endpoint := fmt.Sprintf("/projects/%s/%s/issues/", a.config.Organization, project)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var issues []SentryIssue
	if err := json.Unmarshal(resp, &issues); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Issues for %s (query: %s):\n\n", project, query))

	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", issue.ShortID, issue.Title))
		sb.WriteString(fmt.Sprintf("  Status: %s | Level: %s\n", issue.Status, issue.Level))
		sb.WriteString(fmt.Sprintf("  Events: %s | Users: %s\n", issue.Count, issue.UserCount))
		sb.WriteString(fmt.Sprintf("  First seen: %s\n", issue.FirstSeen))
		sb.WriteString(fmt.Sprintf("  Last seen: %s\n\n", issue.LastSeen))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) getIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueID, _ := args["issue_id"].(string)

	endpoint := fmt.Sprintf("/issues/%s/", issueID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var issue SentryIssueDetail
	if err := json.Unmarshal(resp, &issue); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Issue: %s\n", issue.Title))
	sb.WriteString(fmt.Sprintf("ID: %s (%s)\n", issue.ID, issue.ShortID))
	sb.WriteString(fmt.Sprintf("Status: %s\n", issue.Status))
	sb.WriteString(fmt.Sprintf("Level: %s\n", issue.Level))
	sb.WriteString(fmt.Sprintf("Platform: %s\n", issue.Platform))
	sb.WriteString(fmt.Sprintf("Events: %s\n", issue.Count))
	sb.WriteString(fmt.Sprintf("Users affected: %s\n", issue.UserCount))
	sb.WriteString(fmt.Sprintf("First seen: %s\n", issue.FirstSeen))
	sb.WriteString(fmt.Sprintf("Last seen: %s\n\n", issue.LastSeen))

	if issue.Metadata != nil {
		sb.WriteString("Metadata:\n")
		if issue.Metadata.Type != "" {
			sb.WriteString(fmt.Sprintf("  Type: %s\n", issue.Metadata.Type))
		}
		if issue.Metadata.Value != "" {
			sb.WriteString(fmt.Sprintf("  Value: %s\n", issue.Metadata.Value))
		}
		if issue.Metadata.Filename != "" {
			sb.WriteString(fmt.Sprintf("  File: %s\n", issue.Metadata.Filename))
		}
	}

	if issue.Culprit != "" {
		sb.WriteString(fmt.Sprintf("\nCulprit: %s\n", issue.Culprit))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) resolveIssue(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueID, _ := args["issue_id"].(string)
	status, _ := args["status"].(string)

	payload := map[string]interface{}{
		"status": status,
	}

	endpoint := fmt.Sprintf("/issues/%s/", issueID)
	_, err := a.makeRequest(ctx, http.MethodPut, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Issue %s status updated to: %s", issueID, status)}},
	}, nil
}

func (a *SentryAdapter) listEvents(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	issueID, _ := args["issue_id"].(string)
	limit := getIntArg(args, "limit", 25)

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))

	endpoint := fmt.Sprintf("/issues/%s/events/", issueID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var events []SentryEvent
	if err := json.Unmarshal(resp, &events); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Events for issue %s (%d):\n\n", issueID, len(events)))

	for _, event := range events {
		sb.WriteString(fmt.Sprintf("- %s\n", event.EventID))
		sb.WriteString(fmt.Sprintf("  Title: %s\n", event.Title))
		sb.WriteString(fmt.Sprintf("  Date: %s\n", event.DateTime))
		if event.User != nil && event.User.Email != "" {
			sb.WriteString(fmt.Sprintf("  User: %s\n", event.User.Email))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) getEvent(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	project, _ := args["project"].(string)
	eventID, _ := args["event_id"].(string)

	endpoint := fmt.Sprintf("/projects/%s/%s/events/%s/", a.config.Organization, project, eventID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var event SentryEventDetail
	if err := json.Unmarshal(resp, &event); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Event: %s\n", event.EventID))
	sb.WriteString(fmt.Sprintf("Title: %s\n", event.Title))
	sb.WriteString(fmt.Sprintf("Date: %s\n", event.DateTime))
	sb.WriteString(fmt.Sprintf("Platform: %s\n", event.Platform))

	if event.User != nil {
		sb.WriteString("\nUser:\n")
		if event.User.ID != "" {
			sb.WriteString(fmt.Sprintf("  ID: %s\n", event.User.ID))
		}
		if event.User.Email != "" {
			sb.WriteString(fmt.Sprintf("  Email: %s\n", event.User.Email))
		}
		if event.User.IPAddress != "" {
			sb.WriteString(fmt.Sprintf("  IP: %s\n", event.User.IPAddress))
		}
	}

	if event.Contexts != nil {
		sb.WriteString("\nContexts:\n")
		for name, ctx := range event.Contexts {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", name, ctx))
		}
	}

	if len(event.Tags) > 0 {
		sb.WriteString("\nTags:\n")
		for _, tag := range event.Tags {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", tag.Key, tag.Value))
		}
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) listAlerts(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	project, _ := args["project"].(string)

	endpoint := fmt.Sprintf("/projects/%s/%s/rules/", a.config.Organization, project)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var rules []SentryAlertRule
	if err := json.Unmarshal(resp, &rules); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Alert rules for %s (%d):\n\n", project, len(rules)))

	for _, rule := range rules {
		sb.WriteString(fmt.Sprintf("- %s\n", rule.Name))
		sb.WriteString(fmt.Sprintf("  ID: %s\n", rule.ID))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", rule.Status))
		sb.WriteString(fmt.Sprintf("  Frequency: %d minutes\n", rule.Frequency))
		if len(rule.Conditions) > 0 {
			sb.WriteString("  Conditions:\n")
			for _, cond := range rule.Conditions {
				sb.WriteString(fmt.Sprintf("    - %s\n", cond.Name))
			}
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) queryStats(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	project, _ := args["project"].(string)
	stat, _ := args["stat"].(string)
	if stat == "" {
		stat = "received"
	}
	resolution, _ := args["resolution"].(string)
	if resolution == "" {
		resolution = "1h"
	}

	params := url.Values{}
	params.Set("stat", stat)
	params.Set("resolution", resolution)

	endpoint := fmt.Sprintf("/projects/%s/%s/stats/", a.config.Organization, project)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var stats [][]interface{}
	if err := json.Unmarshal(resp, &stats); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Stats for %s (%s, resolution: %s):\n\n", project, stat, resolution))

	var total int64
	for _, point := range stats {
		if len(point) >= 2 {
			if count, ok := point[1].(float64); ok {
				total += int64(count)
			}
		}
	}

	sb.WriteString(fmt.Sprintf("Total %s: %d\n", stat, total))
	sb.WriteString(fmt.Sprintf("Data points: %d\n", len(stats)))

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) searchIssues(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	limit := getIntArg(args, "limit", 25)

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))

	endpoint := fmt.Sprintf("/organizations/%s/issues/", a.config.Organization)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var issues []SentryIssue
	if err := json.Unmarshal(resp, &issues); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for '%s' (%d):\n\n", query, len(issues)))

	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", issue.ShortID, issue.Title))
		sb.WriteString(fmt.Sprintf("  Project: %s\n", issue.Project.Slug))
		sb.WriteString(fmt.Sprintf("  Status: %s | Events: %s\n", issue.Status, issue.Count))
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SentryAdapter) makeRequest(ctx context.Context, method, endpoint string, params url.Values, payload interface{}) ([]byte, error) {
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

	req.Header.Set("Authorization", "Bearer "+a.config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return body, nil
}

// Sentry API response types

// SentryProject represents a Sentry project.
type SentryProject struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Platform string `json:"platform"`
	Status   string `json:"status"`
}

// SentryIssue represents a Sentry issue.
type SentryIssue struct {
	ID        string         `json:"id"`
	ShortID   string         `json:"shortId"`
	Title     string         `json:"title"`
	Culprit   string         `json:"culprit"`
	Level     string         `json:"level"`
	Status    string         `json:"status"`
	Count     string         `json:"count"`
	UserCount string         `json:"userCount"`
	FirstSeen string         `json:"firstSeen"`
	LastSeen  string         `json:"lastSeen"`
	Project   *SentryProject `json:"project"`
}

// SentryIssueDetail represents detailed issue information.
type SentryIssueDetail struct {
	SentryIssue
	Platform string               `json:"platform"`
	Metadata *SentryIssueMetadata `json:"metadata"`
}

// SentryIssueMetadata represents issue metadata.
type SentryIssueMetadata struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Filename string `json:"filename"`
}

// SentryEvent represents a Sentry event.
type SentryEvent struct {
	EventID  string      `json:"eventID"`
	Title    string      `json:"title"`
	DateTime string      `json:"dateCreated"`
	User     *SentryUser `json:"user"`
}

// SentryEventDetail represents detailed event information.
type SentryEventDetail struct {
	SentryEvent
	Platform string                 `json:"platform"`
	Contexts map[string]interface{} `json:"contexts"`
	Tags     []SentryTag            `json:"tags"`
}

// SentryUser represents a Sentry user.
type SentryUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
}

// SentryTag represents a tag.
type SentryTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SentryAlertRule represents an alert rule.
type SentryAlertRule struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Frequency  int               `json:"frequency"`
	Conditions []SentryCondition `json:"conditions"`
	Actions    []SentryAction    `json:"actions"`
}

// SentryCondition represents a rule condition.
type SentryCondition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SentryAction represents a rule action.
type SentryAction struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
