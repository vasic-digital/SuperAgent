// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// FetchAdapterConfig holds configuration for Fetch MCP adapter
type FetchAdapterConfig struct {
	// UserAgent is the User-Agent header to use
	UserAgent string `json:"user_agent,omitempty"`
	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout,omitempty"`
	// MaxRedirects is the maximum number of redirects to follow
	MaxRedirects int `json:"max_redirects,omitempty"`
	// MaxResponseSize is the maximum response size in bytes
	MaxResponseSize int64 `json:"max_response_size,omitempty"`
	// AllowedDomains restricts fetching to specific domains (empty means all allowed)
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	// BlockedDomains blocks fetching from specific domains
	BlockedDomains []string `json:"blocked_domains,omitempty"`
	// IgnoreRobotsTxt ignores robots.txt restrictions
	IgnoreRobotsTxt bool `json:"ignore_robots_txt"`
	// DefaultHeaders are headers added to all requests
	DefaultHeaders map[string]string `json:"default_headers,omitempty"`
}

// DefaultFetchAdapterConfig returns default configuration
func DefaultFetchAdapterConfig() FetchAdapterConfig {
	return FetchAdapterConfig{
		UserAgent:       "HelixAgent-Fetch/1.0",
		Timeout:         30 * time.Second,
		MaxRedirects:    10,
		MaxResponseSize: 10 * 1024 * 1024, // 10MB
		IgnoreRobotsTxt: false,
		DefaultHeaders:  make(map[string]string),
	}
}

// FetchAdapter implements MCP adapter for web content fetching
type FetchAdapter struct {
	config      FetchAdapterConfig
	client      *http.Client
	initialized bool
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// NewFetchAdapter creates a new Fetch MCP adapter
func NewFetchAdapter(config FetchAdapterConfig, logger *logrus.Logger) *FetchAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.UserAgent == "" {
		config.UserAgent = "HelixAgent-Fetch/1.0"
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRedirects <= 0 {
		config.MaxRedirects = 10
	}
	if config.MaxResponseSize <= 0 {
		config.MaxResponseSize = 10 * 1024 * 1024
	}
	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string)
	}

	return &FetchAdapter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the Fetch adapter
func (f *FetchAdapter) Initialize(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	redirectFunc := func(req *http.Request, via []*http.Request) error {
		if len(via) >= f.config.MaxRedirects {
			return fmt.Errorf("stopped after %d redirects", f.config.MaxRedirects)
		}
		return nil
	}

	f.client = &http.Client{
		Timeout:       f.config.Timeout,
		CheckRedirect: redirectFunc,
	}

	f.initialized = true
	f.logger.Info("Fetch adapter initialized")
	return nil
}

// Health returns health status
func (f *FetchAdapter) Health(ctx context.Context) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.initialized {
		return fmt.Errorf("Fetch adapter not initialized")
	}

	return nil
}

// Close closes the adapter
func (f *FetchAdapter) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.initialized = false
	return nil
}

// isDomainAllowed checks if a domain is allowed
func (f *FetchAdapter) isDomainAllowed(domain string) bool {
	domain = strings.ToLower(domain)

	// Check blocked domains first
	for _, blocked := range f.config.BlockedDomains {
		if strings.EqualFold(blocked, domain) || strings.HasSuffix(domain, "."+strings.ToLower(blocked)) {
			return false
		}
	}

	// If no allowed domains specified, allow all (except blocked)
	if len(f.config.AllowedDomains) == 0 {
		return true
	}

	// Check allowed domains
	for _, allowed := range f.config.AllowedDomains {
		if strings.EqualFold(allowed, domain) || strings.HasSuffix(domain, "."+strings.ToLower(allowed)) {
			return true
		}
	}

	return false
}

// FetchResult represents the result of a fetch operation
type FetchResult struct {
	URL            string              `json:"url"`
	FinalURL       string              `json:"final_url,omitempty"`
	StatusCode     int                 `json:"status_code"`
	Status         string              `json:"status"`
	Headers        map[string][]string `json:"headers"`
	ContentType    string              `json:"content_type"`
	ContentLength  int64               `json:"content_length"`
	Content        string              `json:"content,omitempty"`
	Duration       time.Duration       `json:"duration"`
	Redirects      int                 `json:"redirects,omitempty"`
	Error          string              `json:"error,omitempty"`
}

// Fetch fetches content from a URL
func (f *FetchAdapter) Fetch(ctx context.Context, targetURL string, method string, headers map[string]string, body string) (*FetchResult, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	// Parse URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check domain restrictions
	if !f.isDomainAllowed(parsedURL.Hostname()) {
		return nil, fmt.Errorf("domain not allowed: %s", parsedURL.Hostname())
	}

	// Default to GET
	if method == "" {
		method = "GET"
	}

	start := time.Now()

	// Create request
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", f.config.UserAgent)

	// Set default headers
	for k, v := range f.config.DefaultHeaders {
		req.Header.Set(k, v)
	}

	// Set custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Perform request
	resp, err := f.client.Do(req)
	if err != nil {
		return &FetchResult{
			URL:      targetURL,
			Duration: time.Since(start),
			Error:    err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	result := &FetchResult{
		URL:           targetURL,
		FinalURL:      resp.Request.URL.String(),
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       resp.Header,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		Duration:      time.Since(start),
	}

	// Count redirects
	if result.FinalURL != targetURL {
		result.Redirects = 1 // Simplified count
	}

	// Read body with size limit
	reader := io.LimitReader(resp.Body, f.config.MaxResponseSize)
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read body: %v", err)
		return result, nil
	}

	result.Content = string(bodyBytes)
	result.ContentLength = int64(len(bodyBytes))

	return result, nil
}

// FetchJSON fetches and parses JSON from a URL
func (f *FetchAdapter) FetchJSON(ctx context.Context, targetURL string, headers map[string]string) (interface{}, error) {
	result, err := f.Fetch(ctx, targetURL, "GET", headers, "")
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("fetch error: %s", result.Error)
	}

	if result.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %s", result.Status)
	}

	var data interface{}
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return data, nil
}

// ExtractLinks extracts links from HTML content
func (f *FetchAdapter) ExtractLinks(content string, baseURL string) ([]string, error) {
	// Simple regex-based link extraction
	linkRegex := regexp.MustCompile(`href=["']([^"']+)["']`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	base, err := url.Parse(baseURL)
	if err != nil {
		base = nil
	}

	var links []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		link := match[1]

		// Skip javascript and anchor links
		if strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "#") {
			continue
		}

		// Resolve relative URLs
		if base != nil && !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			resolved, err := base.Parse(link)
			if err != nil {
				continue
			}
			link = resolved.String()
		}

		if !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}

	return links, nil
}

// ExtractText extracts text content from HTML
func (f *FetchAdapter) ExtractText(content string) string {
	// Remove script and style tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
	content = styleRegex.ReplaceAllString(content, "")

	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	content = tagRegex.ReplaceAllString(content, " ")

	// Decode common HTML entities
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")

	// Normalize whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	content = spaceRegex.ReplaceAllString(content, " ")

	return strings.TrimSpace(content)
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (f *FetchAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "fetch_url",
			Description: "Fetch content from a URL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to fetch",
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": "HTTP method (GET, POST, etc.)",
						"default":     "GET",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": "Custom headers to send",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Request body (for POST/PUT)",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "fetch_json",
			Description: "Fetch and parse JSON from a URL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to fetch JSON from",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": "Custom headers to send",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "fetch_extract_links",
			Description: "Fetch a URL and extract all links from the HTML",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to fetch and extract links from",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "fetch_extract_text",
			Description: "Fetch a URL and extract plain text from the HTML",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to fetch and extract text from",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (f *FetchAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	f.mu.RLock()
	initialized := f.initialized
	f.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	switch toolName {
	case "fetch_url":
		targetURL, _ := params["url"].(string)
		method, _ := params["method"].(string)
		body, _ := params["body"].(string)
		var headers map[string]string
		if h, ok := params["headers"].(map[string]interface{}); ok {
			headers = make(map[string]string)
			for k, v := range h {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}
		return f.Fetch(ctx, targetURL, method, headers, body)

	case "fetch_json":
		targetURL, _ := params["url"].(string)
		var headers map[string]string
		if h, ok := params["headers"].(map[string]interface{}); ok {
			headers = make(map[string]string)
			for k, v := range h {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}
		return f.FetchJSON(ctx, targetURL, headers)

	case "fetch_extract_links":
		targetURL, _ := params["url"].(string)
		result, err := f.Fetch(ctx, targetURL, "GET", nil, "")
		if err != nil {
			return nil, err
		}
		if result.Error != "" {
			return nil, fmt.Errorf("fetch error: %s", result.Error)
		}
		return f.ExtractLinks(result.Content, result.FinalURL)

	case "fetch_extract_text":
		targetURL, _ := params["url"].(string)
		result, err := f.Fetch(ctx, targetURL, "GET", nil, "")
		if err != nil {
			return nil, err
		}
		if result.Error != "" {
			return nil, fmt.Errorf("fetch error: %s", result.Error)
		}
		return f.ExtractText(result.Content), nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (f *FetchAdapter) GetCapabilities() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return map[string]interface{}{
		"name":              "fetch",
		"user_agent":        f.config.UserAgent,
		"timeout":           f.config.Timeout.String(),
		"max_redirects":     f.config.MaxRedirects,
		"max_response_size": f.config.MaxResponseSize,
		"tools":             len(f.GetMCPTools()),
		"initialized":       f.initialized,
	}
}

// MarshalJSON implements custom JSON marshaling
func (f *FetchAdapter) MarshalJSON() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"initialized":  f.initialized,
		"capabilities": f.GetCapabilities(),
	})
}
