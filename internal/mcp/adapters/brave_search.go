// Package adapters provides MCP server adapters.
// This file implements the Brave Search MCP server adapter.
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

// BraveSearchConfig configures the Brave Search adapter.
type BraveSearchConfig struct {
	APIKey         string        `json:"api_key"`
	BaseURL        string        `json:"base_url"`
	Timeout        time.Duration `json:"timeout"`
	MaxResults     int           `json:"max_results"`
	SafeSearch     string        `json:"safe_search"` // off, moderate, strict
	CountryCode    string        `json:"country_code"`
	SearchLanguage string        `json:"search_language"`
}

// DefaultBraveSearchConfig returns default configuration.
func DefaultBraveSearchConfig() BraveSearchConfig {
	return BraveSearchConfig{
		BaseURL:        "https://api.search.brave.com/res/v1",
		Timeout:        30 * time.Second,
		MaxResults:     10,
		SafeSearch:     "moderate",
		CountryCode:    "us",
		SearchLanguage: "en",
	}
}

// BraveSearchAdapter implements the Brave Search MCP server.
type BraveSearchAdapter struct {
	config     BraveSearchConfig
	httpClient *http.Client
}

// NewBraveSearchAdapter creates a new Brave Search adapter.
func NewBraveSearchAdapter(config BraveSearchConfig) *BraveSearchAdapter {
	return &BraveSearchAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *BraveSearchAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "brave-search",
		Version:     "1.0.0",
		Description: "Brave Search API integration for web, image, video, and news search",
		Capabilities: []string{
			"web_search",
			"image_search",
			"video_search",
			"news_search",
			"local_search",
		},
	}
}

// ListTools returns available tools.
func (a *BraveSearchAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "brave_web_search",
			Description: "Search the web using Brave Search",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results (max 20)",
						"default":     10,
					},
					"freshness": map[string]interface{}{
						"type":        "string",
						"description": "Filter by time: pd (day), pw (week), pm (month), py (year)",
						"enum":        []string{"", "pd", "pw", "pm", "py"},
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "brave_local_search",
			Description: "Search for local businesses and places",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (e.g., 'coffee shops near me')",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "brave_image_search",
			Description: "Search for images using Brave Search",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Image search query",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results",
						"default":     10,
					},
					"size": map[string]interface{}{
						"type":        "string",
						"description": "Image size filter",
						"enum":        []string{"", "small", "medium", "large", "wallpaper"},
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "brave_news_search",
			Description: "Search for news articles",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "News search query",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results",
						"default":     10,
					},
					"freshness": map[string]interface{}{
						"type":        "string",
						"description": "Filter by time",
						"enum":        []string{"", "pd", "pw", "pm"},
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *BraveSearchAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "brave_web_search":
		return a.webSearch(ctx, args)
	case "brave_local_search":
		return a.localSearch(ctx, args)
	case "brave_image_search":
		return a.imageSearch(ctx, args)
	case "brave_news_search":
		return a.newsSearch(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *BraveSearchAdapter) webSearch(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	count := getIntArg(args, "count", a.config.MaxResults)
	freshness, _ := args["freshness"].(string)

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("safesearch", a.config.SafeSearch)
	params.Set("country", a.config.CountryCode)
	params.Set("search_lang", a.config.SearchLanguage)
	if freshness != "" {
		params.Set("freshness", freshness)
	}

	resp, err := a.makeRequest(ctx, "/web/search", params)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var results BraveWebSearchResponse
	if err := json.Unmarshal(resp, &results); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	// Format results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results for '%s':\n\n", len(results.Web.Results), query))

	for i, result := range results.Web.Results {
		sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		sb.WriteString(fmt.Sprintf("   %s\n\n", result.Description))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *BraveSearchAdapter) localSearch(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	count := getIntArg(args, "count", 5)

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("result_filter", "locations")

	resp, err := a.makeRequest(ctx, "/web/search", params)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var results BraveLocalSearchResponse
	if err := json.Unmarshal(resp, &results); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Local results for '%s':\n\n", query))

	if results.Locations != nil {
		for i, loc := range results.Locations.Results {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, loc.Title))
			if loc.Address != "" {
				sb.WriteString(fmt.Sprintf("   Address: %s\n", loc.Address))
			}
			if loc.Rating != nil {
				sb.WriteString(fmt.Sprintf("   Rating: %.1f (%d reviews)\n", loc.Rating.Value, loc.Rating.Count))
			}
			if loc.Phone != "" {
				sb.WriteString(fmt.Sprintf("   Phone: %s\n", loc.Phone))
			}
			sb.WriteString("\n")
		}
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *BraveSearchAdapter) imageSearch(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	count := getIntArg(args, "count", 10)
	size, _ := args["size"].(string)

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("safesearch", a.config.SafeSearch)
	if size != "" {
		params.Set("size", size)
	}

	resp, err := a.makeRequest(ctx, "/images/search", params)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var results BraveImageSearchResponse
	if err := json.Unmarshal(resp, &results); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d images for '%s':\n\n", len(results.Results), query))

	for i, img := range results.Results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, img.Title))
		sb.WriteString(fmt.Sprintf("   Source: %s\n", img.Source))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", img.URL))
		sb.WriteString(fmt.Sprintf("   Size: %dx%d\n\n", img.Properties.Width, img.Properties.Height))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *BraveSearchAdapter) newsSearch(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	count := getIntArg(args, "count", 10)
	freshness, _ := args["freshness"].(string)

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	if freshness != "" {
		params.Set("freshness", freshness)
	}

	resp, err := a.makeRequest(ctx, "/news/search", params)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var results BraveNewsSearchResponse
	if err := json.Unmarshal(resp, &results); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d news articles for '%s':\n\n", len(results.Results), query))

	for i, article := range results.Results {
		sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, article.Title))
		sb.WriteString(fmt.Sprintf("   Source: %s\n", article.Source))
		sb.WriteString(fmt.Sprintf("   Published: %s\n", article.Age))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", article.URL))
		sb.WriteString(fmt.Sprintf("   %s\n\n", article.Description))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *BraveSearchAdapter) makeRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s?%s", a.config.BaseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", a.config.APIKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return body, nil
}

// Response types

type BraveWebSearchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

type BraveLocalSearchResponse struct {
	Locations *struct {
		Results []struct {
			Title   string `json:"title"`
			Address string `json:"address"`
			Phone   string `json:"phone"`
			Rating  *struct {
				Value float64 `json:"ratingValue"`
				Count int     `json:"ratingCount"`
			} `json:"rating"`
		} `json:"results"`
	} `json:"locations"`
}

type BraveImageSearchResponse struct {
	Results []struct {
		Title      string `json:"title"`
		URL        string `json:"url"`
		Source     string `json:"source"`
		Properties struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"properties"`
	} `json:"results"`
}

type BraveNewsSearchResponse struct {
	Results []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Age         string `json:"age"`
	} `json:"results"`
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	if v, ok := args[key].(int); ok {
		return v
	}
	return defaultVal
}
