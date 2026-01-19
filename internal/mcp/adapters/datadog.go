// Package adapters provides MCP server adapters.
// This file implements the Datadog MCP server adapter for monitoring and analytics.
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

// DatadogConfig configures the Datadog adapter.
type DatadogConfig struct {
	APIKey         string        `json:"api_key"`
	ApplicationKey string        `json:"application_key"`
	Site           string        `json:"site"`     // e.g., "datadoghq.com", "datadoghq.eu"
	BaseURL        string        `json:"base_url"` // Optional: override full API base URL (for testing)
	Timeout        time.Duration `json:"timeout"`
}

// DefaultDatadogConfig returns default configuration.
func DefaultDatadogConfig() DatadogConfig {
	return DatadogConfig{
		Site:    "datadoghq.com",
		Timeout: 30 * time.Second,
	}
}

// DatadogAdapter implements the Datadog MCP server.
type DatadogAdapter struct {
	config     DatadogConfig
	httpClient *http.Client
}

// NewDatadogAdapter creates a new Datadog adapter.
func NewDatadogAdapter(config DatadogConfig) *DatadogAdapter {
	if config.Site == "" {
		config.Site = DefaultDatadogConfig().Site
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultDatadogConfig().Timeout
	}
	return &DatadogAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *DatadogAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "datadog",
		Version:     "1.0.0",
		Description: "Datadog monitoring and analytics integration",
		Capabilities: []string{
			"query_metrics",
			"list_dashboards",
			"get_dashboard",
			"list_monitors",
			"get_monitor",
			"mute_monitor",
			"query_logs",
			"list_hosts",
			"list_events",
		},
	}
}

// ListTools returns available tools.
func (a *DatadogAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "datadog_query_metrics",
			Description: "Query time series metrics",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Metrics query (e.g., 'avg:system.cpu.user{*}')",
					},
					"from": map[string]interface{}{
						"type":        "integer",
						"description": "Start time (Unix timestamp)",
					},
					"to": map[string]interface{}{
						"type":        "integer",
						"description": "End time (Unix timestamp)",
					},
				},
				"required": []string{"query", "from", "to"},
			},
		},
		{
			Name:        "datadog_list_dashboards",
			Description: "List all dashboards",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "Filter dashboards by name",
					},
				},
			},
		},
		{
			Name:        "datadog_get_dashboard",
			Description: "Get dashboard details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dashboard_id": map[string]interface{}{
						"type":        "string",
						"description": "Dashboard ID",
					},
				},
				"required": []string{"dashboard_id"},
			},
		},
		{
			Name:        "datadog_list_monitors",
			Description: "List all monitors",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"group_states": map[string]interface{}{
						"type":        "string",
						"description": "Filter by state (alert, warn, no data, ok)",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Filter by name",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
				},
			},
		},
		{
			Name:        "datadog_get_monitor",
			Description: "Get monitor details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"monitor_id": map[string]interface{}{
						"type":        "integer",
						"description": "Monitor ID",
					},
				},
				"required": []string{"monitor_id"},
			},
		},
		{
			Name:        "datadog_mute_monitor",
			Description: "Mute or unmute a monitor",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"monitor_id": map[string]interface{}{
						"type":        "integer",
						"description": "Monitor ID",
					},
					"scope": map[string]interface{}{
						"type":        "string",
						"description": "Scope to mute (e.g., 'host:myhost')",
					},
					"end": map[string]interface{}{
						"type":        "integer",
						"description": "Unix timestamp when mute should end",
					},
				},
				"required": []string{"monitor_id"},
			},
		},
		{
			Name:        "datadog_unmute_monitor",
			Description: "Unmute a monitor",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"monitor_id": map[string]interface{}{
						"type":        "integer",
						"description": "Monitor ID",
					},
					"scope": map[string]interface{}{
						"type":        "string",
						"description": "Scope to unmute",
					},
				},
				"required": []string{"monitor_id"},
			},
		},
		{
			Name:        "datadog_query_logs",
			Description: "Query logs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Log search query",
					},
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Start time (ISO 8601 or relative like 'now-1h')",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "End time (ISO 8601 or relative like 'now')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of logs to return",
						"default":     50,
					},
					"sort": map[string]interface{}{
						"type":        "string",
						"description": "Sort order",
						"enum":        []string{"asc", "desc"},
						"default":     "desc",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "datadog_list_hosts",
			Description: "List infrastructure hosts",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "Filter string for hosts",
					},
					"sort_field": map[string]interface{}{
						"type":        "string",
						"description": "Sort field",
						"enum":        []string{"cpu", "memory", "apps", "name"},
					},
					"sort_dir": map[string]interface{}{
						"type":        "string",
						"description": "Sort direction",
						"enum":        []string{"asc", "desc"},
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of hosts to return",
						"default":     100,
					},
				},
			},
		},
		{
			Name:        "datadog_list_events",
			Description: "List events from the event stream",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start": map[string]interface{}{
						"type":        "integer",
						"description": "Start time (Unix timestamp)",
					},
					"end": map[string]interface{}{
						"type":        "integer",
						"description": "End time (Unix timestamp)",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Event priority",
						"enum":        []string{"normal", "low"},
					},
					"tags": map[string]interface{}{
						"type":        "string",
						"description": "Tags to filter by (comma-separated)",
					},
				},
				"required": []string{"start", "end"},
			},
		},
		{
			Name:        "datadog_create_event",
			Description: "Create a new event",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Event title",
					},
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Event body text",
					},
					"alert_type": map[string]interface{}{
						"type":        "string",
						"description": "Alert type",
						"enum":        []string{"error", "warning", "info", "success"},
						"default":     "info",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Event priority",
						"enum":        []string{"normal", "low"},
						"default":     "normal",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to attach to the event",
					},
				},
				"required": []string{"title", "text"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *DatadogAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "datadog_query_metrics":
		return a.queryMetrics(ctx, args)
	case "datadog_list_dashboards":
		return a.listDashboards(ctx, args)
	case "datadog_get_dashboard":
		return a.getDashboard(ctx, args)
	case "datadog_list_monitors":
		return a.listMonitors(ctx, args)
	case "datadog_get_monitor":
		return a.getMonitor(ctx, args)
	case "datadog_mute_monitor":
		return a.muteMonitor(ctx, args)
	case "datadog_unmute_monitor":
		return a.unmuteMonitor(ctx, args)
	case "datadog_query_logs":
		return a.queryLogs(ctx, args)
	case "datadog_list_hosts":
		return a.listHosts(ctx, args)
	case "datadog_list_events":
		return a.listEvents(ctx, args)
	case "datadog_create_event":
		return a.createEvent(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *DatadogAdapter) queryMetrics(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	from := getInt64Arg(args, "from", 0)
	to := getInt64Arg(args, "to", 0)

	params := url.Values{}
	params.Set("query", query)
	params.Set("from", fmt.Sprintf("%d", from))
	params.Set("to", fmt.Sprintf("%d", to))

	resp, err := a.makeRequest(ctx, http.MethodGet, "/api/v1/query", params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result DDMetricsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Metrics query: %s\n", query))
	sb.WriteString(fmt.Sprintf("Status: %s\n\n", result.Status))

	for _, series := range result.Series {
		sb.WriteString(fmt.Sprintf("Series: %s\n", series.DisplayName))
		sb.WriteString(fmt.Sprintf("  Scope: %s\n", series.Scope))
		sb.WriteString(fmt.Sprintf("  Points: %d\n", len(series.Pointlist)))
		if len(series.Pointlist) > 0 {
			sb.WriteString("  Recent values:\n")
			start := len(series.Pointlist) - 5
			if start < 0 {
				start = 0
			}
			for _, point := range series.Pointlist[start:] {
				if len(point) >= 2 {
					ts := int64(point[0].(float64) / 1000)
					val := point[1].(float64)
					sb.WriteString(fmt.Sprintf("    %s: %.2f\n", time.Unix(ts, 0).Format(time.RFC3339), val))
				}
			}
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) listDashboards(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filter, _ := args["filter"].(string)

	params := url.Values{}
	if filter != "" {
		params.Set("filter[shared]", "false")
	}

	resp, err := a.makeRequest(ctx, http.MethodGet, "/api/v1/dashboard", params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result DDDashboardsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Dashboards (%d):\n\n", len(result.Dashboards)))

	for _, dash := range result.Dashboards {
		if filter != "" && !strings.Contains(strings.ToLower(dash.Title), strings.ToLower(filter)) {
			continue
		}
		sb.WriteString(fmt.Sprintf("- %s\n", dash.Title))
		sb.WriteString(fmt.Sprintf("  ID: %s\n", dash.ID))
		sb.WriteString(fmt.Sprintf("  Layout: %s\n", dash.LayoutType))
		sb.WriteString(fmt.Sprintf("  URL: %s\n\n", dash.URL))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) getDashboard(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	dashboardID, _ := args["dashboard_id"].(string)

	endpoint := fmt.Sprintf("/api/v1/dashboard/%s", dashboardID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var dash DDDashboard
	if err := json.Unmarshal(resp, &dash); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Dashboard: %s\n", dash.Title))
	sb.WriteString(fmt.Sprintf("ID: %s\n", dash.ID))
	sb.WriteString(fmt.Sprintf("Description: %s\n", dash.Description))
	sb.WriteString(fmt.Sprintf("Layout: %s\n", dash.LayoutType))
	sb.WriteString(fmt.Sprintf("Created: %s\n", dash.CreatedAt))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", dash.ModifiedAt))
	sb.WriteString(fmt.Sprintf("Widgets: %d\n", len(dash.Widgets)))

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) listMonitors(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	params := url.Values{}

	if groupStates, ok := args["group_states"].(string); ok && groupStates != "" {
		params.Set("group_states", groupStates)
	}
	if name, ok := args["name"].(string); ok && name != "" {
		params.Set("name", name)
	}
	if tagsRaw, ok := args["tags"].([]interface{}); ok && len(tagsRaw) > 0 {
		var tags []string
		for _, t := range tagsRaw {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
		params.Set("tags", strings.Join(tags, ","))
	}

	resp, err := a.makeRequest(ctx, http.MethodGet, "/api/v1/monitor", params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var monitors []DDMonitor
	if err := json.Unmarshal(resp, &monitors); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Monitors (%d):\n\n", len(monitors)))

	for _, mon := range monitors {
		stateIcon := "?"
		switch mon.OverallState {
		case "OK":
			stateIcon = "[OK]"
		case "Alert":
			stateIcon = "[ALERT]"
		case "Warn":
			stateIcon = "[WARN]"
		case "No Data":
			stateIcon = "[NO DATA]"
		}
		sb.WriteString(fmt.Sprintf("- %s %s\n", stateIcon, mon.Name))
		sb.WriteString(fmt.Sprintf("  ID: %d | Type: %s\n", mon.ID, mon.Type))
		sb.WriteString(fmt.Sprintf("  Query: %s\n\n", truncate(mon.Query, 100)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) getMonitor(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	monitorID := getInt64Arg(args, "monitor_id", 0)

	endpoint := fmt.Sprintf("/api/v1/monitor/%d", monitorID)
	resp, err := a.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var mon DDMonitor
	if err := json.Unmarshal(resp, &mon); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Monitor: %s\n", mon.Name))
	sb.WriteString(fmt.Sprintf("ID: %d\n", mon.ID))
	sb.WriteString(fmt.Sprintf("Type: %s\n", mon.Type))
	sb.WriteString(fmt.Sprintf("State: %s\n", mon.OverallState))
	sb.WriteString(fmt.Sprintf("Query: %s\n", mon.Query))
	sb.WriteString(fmt.Sprintf("Message: %s\n", mon.Message))
	sb.WriteString(fmt.Sprintf("Created: %s\n", mon.Created))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", mon.Modified))

	if len(mon.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(mon.Tags, ", ")))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) muteMonitor(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	monitorID := getInt64Arg(args, "monitor_id", 0)
	scope, _ := args["scope"].(string)
	end := getInt64Arg(args, "end", 0)

	payload := map[string]interface{}{}
	if scope != "" {
		payload["scope"] = scope
	}
	if end > 0 {
		payload["end"] = end
	}

	endpoint := fmt.Sprintf("/api/v1/monitor/%d/mute", monitorID)
	_, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Monitor %d muted successfully", monitorID)}},
	}, nil
}

func (a *DatadogAdapter) unmuteMonitor(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	monitorID := getInt64Arg(args, "monitor_id", 0)
	scope, _ := args["scope"].(string)

	payload := map[string]interface{}{}
	if scope != "" {
		payload["scope"] = scope
	}

	endpoint := fmt.Sprintf("/api/v1/monitor/%d/unmute", monitorID)
	_, err := a.makeRequest(ctx, http.MethodPost, endpoint, nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Monitor %d unmuted successfully", monitorID)}},
	}, nil
}

func (a *DatadogAdapter) queryLogs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	from, _ := args["from"].(string)
	if from == "" {
		from = "now-1h"
	}
	to, _ := args["to"].(string)
	if to == "" {
		to = "now"
	}
	limit := getIntArg(args, "limit", 50)
	sort, _ := args["sort"].(string)
	if sort == "" {
		sort = "desc"
	}

	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"query": query,
			"from":  from,
			"to":    to,
		},
		"sort": sort,
		"page": map[string]interface{}{
			"limit": limit,
		},
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/api/v2/logs/events/search", nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result DDLogsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Logs for query: %s\n", query))
	sb.WriteString(fmt.Sprintf("Results: %d\n\n", len(result.Data)))

	for _, log := range result.Data {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", log.Attributes.Timestamp, log.Attributes.Service))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", log.Attributes.Status))
		sb.WriteString(fmt.Sprintf("  Message: %s\n\n", truncate(log.Attributes.Message, 200)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) listHosts(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	params := url.Values{}

	if filter, ok := args["filter"].(string); ok && filter != "" {
		params.Set("filter", filter)
	}
	if sortField, ok := args["sort_field"].(string); ok && sortField != "" {
		params.Set("sort_field", sortField)
	}
	if sortDir, ok := args["sort_dir"].(string); ok && sortDir != "" {
		params.Set("sort_dir", sortDir)
	}
	count := getIntArg(args, "count", 100)
	params.Set("count", fmt.Sprintf("%d", count))

	resp, err := a.makeRequest(ctx, http.MethodGet, "/api/v1/hosts", params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result DDHostsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hosts (total: %d, showing: %d):\n\n", result.TotalReturned, len(result.HostList)))

	for _, host := range result.HostList {
		sb.WriteString(fmt.Sprintf("- %s\n", host.Name))
		sb.WriteString(fmt.Sprintf("  ID: %d\n", host.ID))
		if host.Meta != nil {
			sb.WriteString(fmt.Sprintf("  Platform: %s\n", host.Meta.Platform))
		}
		sb.WriteString(fmt.Sprintf("  Up: %v\n", host.IsUp))
		if len(host.Apps) > 0 {
			sb.WriteString(fmt.Sprintf("  Apps: %s\n", strings.Join(host.Apps, ", ")))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) listEvents(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	start := getInt64Arg(args, "start", 0)
	end := getInt64Arg(args, "end", 0)

	params := url.Values{}
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("end", fmt.Sprintf("%d", end))

	if priority, ok := args["priority"].(string); ok && priority != "" {
		params.Set("priority", priority)
	}
	if tags, ok := args["tags"].(string); ok && tags != "" {
		params.Set("tags", tags)
	}

	resp, err := a.makeRequest(ctx, http.MethodGet, "/api/v1/events", params, nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result DDEventsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Events (%d):\n\n", len(result.Events)))

	for _, event := range result.Events {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", event.AlertType, event.Title))
		sb.WriteString(fmt.Sprintf("  ID: %d\n", event.ID))
		sb.WriteString(fmt.Sprintf("  Date: %s\n", time.Unix(event.DateHappened, 0).Format(time.RFC3339)))
		if event.Text != "" {
			sb.WriteString(fmt.Sprintf("  Text: %s\n", truncate(event.Text, 100)))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DatadogAdapter) createEvent(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	title, _ := args["title"].(string)
	text, _ := args["text"].(string)
	alertType, _ := args["alert_type"].(string)
	if alertType == "" {
		alertType = "info"
	}
	priority, _ := args["priority"].(string)
	if priority == "" {
		priority = "normal"
	}

	payload := map[string]interface{}{
		"title":      title,
		"text":       text,
		"alert_type": alertType,
		"priority":   priority,
	}

	if tagsRaw, ok := args["tags"].([]interface{}); ok && len(tagsRaw) > 0 {
		var tags []string
		for _, t := range tagsRaw {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
		payload["tags"] = tags
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/api/v1/events", nil, payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result DDEventCreateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Event created successfully (ID: %d)", result.Event.ID)}},
	}, nil
}

func (a *DatadogAdapter) makeRequest(ctx context.Context, method, endpoint string, params url.Values, payload interface{}) ([]byte, error) {
	var reqURL string
	if a.config.BaseURL != "" {
		reqURL = a.config.BaseURL + endpoint
	} else {
		reqURL = fmt.Sprintf("https://api.%s%s", a.config.Site, endpoint)
	}
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", a.config.APIKey)
	req.Header.Set("DD-APPLICATION-KEY", a.config.ApplicationKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return body, nil
}

func getInt64Arg(args map[string]interface{}, key string, defaultVal int64) int64 {
	if v, ok := args[key].(float64); ok {
		return int64(v)
	}
	if v, ok := args[key].(int64); ok {
		return v
	}
	if v, ok := args[key].(int); ok {
		return int64(v)
	}
	return defaultVal
}

// Datadog API response types

// DDMetricsResponse represents a metrics query response.
type DDMetricsResponse struct {
	Status string     `json:"status"`
	Series []DDSeries `json:"series"`
}

// DDSeries represents a metrics series.
type DDSeries struct {
	DisplayName string          `json:"display_name"`
	Scope       string          `json:"scope"`
	Pointlist   [][]interface{} `json:"pointlist"`
}

// DDDashboardsResponse represents a dashboards list response.
type DDDashboardsResponse struct {
	Dashboards []DDDashboardSummary `json:"dashboards"`
}

// DDDashboardSummary represents dashboard summary info.
type DDDashboardSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	LayoutType string `json:"layout_type"`
	URL        string `json:"url"`
}

// DDDashboard represents a full dashboard.
type DDDashboard struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	LayoutType  string        `json:"layout_type"`
	CreatedAt   string        `json:"created_at"`
	ModifiedAt  string        `json:"modified_at"`
	Widgets     []interface{} `json:"widgets"`
}

// DDMonitor represents a Datadog monitor.
type DDMonitor struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Query        string   `json:"query"`
	Message      string   `json:"message"`
	OverallState string   `json:"overall_state"`
	Tags         []string `json:"tags"`
	Created      string   `json:"created"`
	Modified     string   `json:"modified"`
}

// DDLogsResponse represents a logs query response.
type DDLogsResponse struct {
	Data []DDLogEntry `json:"data"`
}

// DDLogEntry represents a log entry.
type DDLogEntry struct {
	ID         string          `json:"id"`
	Attributes DDLogAttributes `json:"attributes"`
}

// DDLogAttributes represents log entry attributes.
type DDLogAttributes struct {
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// DDHostsResponse represents a hosts list response.
type DDHostsResponse struct {
	TotalReturned int      `json:"total_returned"`
	HostList      []DDHost `json:"host_list"`
}

// DDHost represents a host.
type DDHost struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Apps []string `json:"apps"`
	IsUp bool     `json:"is_up"`
	Meta *DDHostMeta `json:"meta"`
}

// DDHostMeta represents host metadata.
type DDHostMeta struct {
	Platform string `json:"platform"`
}

// DDEventsResponse represents an events list response.
type DDEventsResponse struct {
	Events []DDEvent `json:"events"`
}

// DDEvent represents an event.
type DDEvent struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Text         string `json:"text"`
	DateHappened int64  `json:"date_happened"`
	AlertType    string `json:"alert_type"`
	Priority     string `json:"priority"`
}

// DDEventCreateResponse represents an event creation response.
type DDEventCreateResponse struct {
	Event DDEvent `json:"event"`
}
