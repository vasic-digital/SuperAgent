// Package adapters provides MCP server adapters.
// This file implements the Puppeteer MCP server adapter for browser automation.
package adapters

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// PuppeteerConfig configures the Puppeteer adapter.
type PuppeteerConfig struct {
	Headless        bool          `json:"headless"`
	DefaultTimeout  time.Duration `json:"default_timeout"`
	ViewportWidth   int           `json:"viewport_width"`
	ViewportHeight  int           `json:"viewport_height"`
	UserAgent       string        `json:"user_agent,omitempty"`
	ProxyServer     string        `json:"proxy_server,omitempty"`
}

// DefaultPuppeteerConfig returns default configuration.
func DefaultPuppeteerConfig() PuppeteerConfig {
	return PuppeteerConfig{
		Headless:       true,
		DefaultTimeout: 30 * time.Second,
		ViewportWidth:  1920,
		ViewportHeight: 1080,
	}
}

// PuppeteerAdapter implements the Puppeteer MCP server.
type PuppeteerAdapter struct {
	config  PuppeteerConfig
	browser BrowserClient
}

// BrowserClient interface for browser automation.
type BrowserClient interface {
	Navigate(ctx context.Context, url string) error
	Screenshot(ctx context.Context, options ScreenshotOptions) ([]byte, error)
	PDF(ctx context.Context, options PDFOptions) ([]byte, error)
	Click(ctx context.Context, selector string) error
	Type(ctx context.Context, selector, text string) error
	Select(ctx context.Context, selector string, values []string) error
	Evaluate(ctx context.Context, script string) (interface{}, error)
	WaitForSelector(ctx context.Context, selector string, timeout time.Duration) error
	WaitForNavigation(ctx context.Context, timeout time.Duration) error
	GetContent(ctx context.Context) (string, error)
	GetTitle(ctx context.Context) (string, error)
	GetURL(ctx context.Context) (string, error)
	SetViewport(ctx context.Context, width, height int) error
	SetCookie(ctx context.Context, cookie Cookie) error
	GetCookies(ctx context.Context) ([]Cookie, error)
	ClearCookies(ctx context.Context) error
	ScrollTo(ctx context.Context, x, y int) error
	Hover(ctx context.Context, selector string) error
	Focus(ctx context.Context, selector string) error
	Press(ctx context.Context, key string) error
	Close(ctx context.Context) error
}

// ScreenshotOptions represents screenshot options.
type ScreenshotOptions struct {
	FullPage bool   `json:"full_page"`
	Type     string `json:"type"` // "png" or "jpeg"
	Quality  int    `json:"quality,omitempty"`
	Clip     *Clip  `json:"clip,omitempty"`
}

// PDFOptions represents PDF generation options.
type PDFOptions struct {
	Format          string `json:"format,omitempty"` // "A4", "Letter", etc.
	PrintBackground bool   `json:"print_background"`
	Landscape       bool   `json:"landscape"`
	Scale           float64 `json:"scale,omitempty"`
	MarginTop       string `json:"margin_top,omitempty"`
	MarginBottom    string `json:"margin_bottom,omitempty"`
	MarginLeft      string `json:"margin_left,omitempty"`
	MarginRight     string `json:"margin_right,omitempty"`
}

// Clip represents a rectangular area.
type Clip struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Cookie represents a browser cookie.
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path,omitempty"`
	Expires  time.Time `json:"expires,omitempty"`
	HTTPOnly bool      `json:"httpOnly,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	SameSite string    `json:"sameSite,omitempty"`
}

// NewPuppeteerAdapter creates a new Puppeteer adapter.
func NewPuppeteerAdapter(config PuppeteerConfig, browser BrowserClient) *PuppeteerAdapter {
	return &PuppeteerAdapter{
		config:  config,
		browser: browser,
	}
}

// GetServerInfo returns server information.
func (a *PuppeteerAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "puppeteer",
		Version:     "1.0.0",
		Description: "Browser automation for web scraping, testing, and interaction",
		Capabilities: []string{
			"navigation",
			"screenshots",
			"pdf_generation",
			"form_interaction",
			"javascript_execution",
		},
	}
}

// ListTools returns available tools.
func (a *PuppeteerAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "puppeteer_navigate",
			Description: "Navigate to a URL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to navigate to",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "puppeteer_screenshot",
			Description: "Take a screenshot of the page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"full_page": map[string]interface{}{
						"type":        "boolean",
						"description": "Capture full scrollable page",
						"default":     false,
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Image type",
						"enum":        []string{"png", "jpeg"},
						"default":     "png",
					},
					"quality": map[string]interface{}{
						"type":        "integer",
						"description": "JPEG quality (0-100)",
						"default":     80,
					},
				},
			},
		},
		{
			Name:        "puppeteer_pdf",
			Description: "Generate PDF of the page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Paper format",
						"enum":        []string{"A4", "Letter", "Legal", "Tabloid"},
						"default":     "A4",
					},
					"landscape": map[string]interface{}{
						"type":        "boolean",
						"description": "Landscape orientation",
						"default":     false,
					},
					"print_background": map[string]interface{}{
						"type":        "boolean",
						"description": "Print background graphics",
						"default":     true,
					},
				},
			},
		},
		{
			Name:        "puppeteer_click",
			Description: "Click on an element",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"selector": map[string]interface{}{
						"type":        "string",
						"description": "CSS selector for the element",
					},
				},
				"required": []string{"selector"},
			},
		},
		{
			Name:        "puppeteer_type",
			Description: "Type text into an input field",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"selector": map[string]interface{}{
						"type":        "string",
						"description": "CSS selector for the input",
					},
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Text to type",
					},
				},
				"required": []string{"selector", "text"},
			},
		},
		{
			Name:        "puppeteer_select",
			Description: "Select option(s) in a dropdown",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"selector": map[string]interface{}{
						"type":        "string",
						"description": "CSS selector for the select element",
					},
					"values": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Option value(s) to select",
					},
				},
				"required": []string{"selector", "values"},
			},
		},
		{
			Name:        "puppeteer_evaluate",
			Description: "Execute JavaScript in the page context",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"script": map[string]interface{}{
						"type":        "string",
						"description": "JavaScript code to execute",
					},
				},
				"required": []string{"script"},
			},
		},
		{
			Name:        "puppeteer_wait_for_selector",
			Description: "Wait for an element to appear",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"selector": map[string]interface{}{
						"type":        "string",
						"description": "CSS selector to wait for",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in milliseconds",
						"default":     30000,
					},
				},
				"required": []string{"selector"},
			},
		},
		{
			Name:        "puppeteer_get_content",
			Description: "Get the page HTML content",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "puppeteer_get_title",
			Description: "Get the page title",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "puppeteer_scroll",
			Description: "Scroll to a position on the page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"x": map[string]interface{}{
						"type":        "integer",
						"description": "X coordinate",
						"default":     0,
					},
					"y": map[string]interface{}{
						"type":        "integer",
						"description": "Y coordinate",
						"default":     0,
					},
				},
			},
		},
		{
			Name:        "puppeteer_hover",
			Description: "Hover over an element",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"selector": map[string]interface{}{
						"type":        "string",
						"description": "CSS selector for the element",
					},
				},
				"required": []string{"selector"},
			},
		},
		{
			Name:        "puppeteer_set_cookie",
			Description: "Set a cookie",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Cookie name",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "Cookie value",
					},
					"domain": map[string]interface{}{
						"type":        "string",
						"description": "Cookie domain",
					},
				},
				"required": []string{"name", "value", "domain"},
			},
		},
		{
			Name:        "puppeteer_clear_cookies",
			Description: "Clear all cookies",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// CallTool executes a tool.
func (a *PuppeteerAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "puppeteer_navigate":
		return a.navigate(ctx, args)
	case "puppeteer_screenshot":
		return a.screenshot(ctx, args)
	case "puppeteer_pdf":
		return a.pdf(ctx, args)
	case "puppeteer_click":
		return a.click(ctx, args)
	case "puppeteer_type":
		return a.typeText(ctx, args)
	case "puppeteer_select":
		return a.selectOption(ctx, args)
	case "puppeteer_evaluate":
		return a.evaluate(ctx, args)
	case "puppeteer_wait_for_selector":
		return a.waitForSelector(ctx, args)
	case "puppeteer_get_content":
		return a.getContent(ctx)
	case "puppeteer_get_title":
		return a.getTitle(ctx)
	case "puppeteer_scroll":
		return a.scroll(ctx, args)
	case "puppeteer_hover":
		return a.hover(ctx, args)
	case "puppeteer_set_cookie":
		return a.setCookie(ctx, args)
	case "puppeteer_clear_cookies":
		return a.clearCookies(ctx)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *PuppeteerAdapter) navigate(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	url, _ := args["url"].(string)

	err := a.browser.Navigate(ctx, url)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	title, _ := a.browser.GetTitle(ctx)
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Navigated to %s\nPage title: %s", url, title)}},
	}, nil
}

func (a *PuppeteerAdapter) screenshot(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	fullPage, _ := args["full_page"].(bool)
	imgType, _ := args["type"].(string)
	if imgType == "" {
		imgType = "png"
	}
	quality := getIntArg(args, "quality", 80)

	options := ScreenshotOptions{
		FullPage: fullPage,
		Type:     imgType,
		Quality:  quality,
	}

	data, err := a.browser.Screenshot(ctx, options)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	mimeType := "image/png"
	if imgType == "jpeg" {
		mimeType = "image/jpeg"
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "image", MimeType: mimeType, Data: encoded},
		},
	}, nil
}

func (a *PuppeteerAdapter) pdf(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	format, _ := args["format"].(string)
	if format == "" {
		format = "A4"
	}
	landscape, _ := args["landscape"].(bool)
	printBackground, _ := args["print_background"].(bool)
	if _, ok := args["print_background"]; !ok {
		printBackground = true
	}

	options := PDFOptions{
		Format:          format,
		Landscape:       landscape,
		PrintBackground: printBackground,
	}

	data, err := a.browser.PDF(ctx, options)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return &ToolResult{
		Content: []ContentBlock{
			{Type: "resource", MimeType: "application/pdf", Data: encoded},
			{Type: "text", Text: fmt.Sprintf("Generated PDF (%d bytes)", len(data))},
		},
	}, nil
}

func (a *PuppeteerAdapter) click(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	selector, _ := args["selector"].(string)

	err := a.browser.Click(ctx, selector)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Clicked on '%s'", selector)}},
	}, nil
}

func (a *PuppeteerAdapter) typeText(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	selector, _ := args["selector"].(string)
	text, _ := args["text"].(string)

	err := a.browser.Type(ctx, selector, text)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Typed '%s' into '%s'", text, selector)}},
	}, nil
}

func (a *PuppeteerAdapter) selectOption(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	selector, _ := args["selector"].(string)
	valuesRaw, _ := args["values"].([]interface{})

	var values []string
	for _, v := range valuesRaw {
		if s, ok := v.(string); ok {
			values = append(values, s)
		}
	}

	err := a.browser.Select(ctx, selector, values)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Selected %v in '%s'", values, selector)}},
	}, nil
}

func (a *PuppeteerAdapter) evaluate(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	script, _ := args["script"].(string)

	result, err := a.browser.Evaluate(ctx, script)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Result: %v", result)}},
	}, nil
}

func (a *PuppeteerAdapter) waitForSelector(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	selector, _ := args["selector"].(string)
	timeout := time.Duration(getIntArg(args, "timeout", 30000)) * time.Millisecond

	err := a.browser.WaitForSelector(ctx, selector, timeout)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Element '%s' found", selector)}},
	}, nil
}

func (a *PuppeteerAdapter) getContent(ctx context.Context) (*ToolResult, error) {
	content, err := a.browser.GetContent(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	// Truncate if too long
	if len(content) > 50000 {
		content = content[:50000] + "\n... (truncated)"
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: content}},
	}, nil
}

func (a *PuppeteerAdapter) getTitle(ctx context.Context) (*ToolResult, error) {
	title, err := a.browser.GetTitle(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Page title: %s", title)}},
	}, nil
}

func (a *PuppeteerAdapter) scroll(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	x := getIntArg(args, "x", 0)
	y := getIntArg(args, "y", 0)

	err := a.browser.ScrollTo(ctx, x, y)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Scrolled to (%d, %d)", x, y)}},
	}, nil
}

func (a *PuppeteerAdapter) hover(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	selector, _ := args["selector"].(string)

	err := a.browser.Hover(ctx, selector)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Hovering over '%s'", selector)}},
	}, nil
}

func (a *PuppeteerAdapter) setCookie(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	value, _ := args["value"].(string)
	domain, _ := args["domain"].(string)

	cookie := Cookie{
		Name:   name,
		Value:  value,
		Domain: domain,
	}

	err := a.browser.SetCookie(ctx, cookie)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Set cookie '%s' for domain '%s'", name, domain)}},
	}, nil
}

func (a *PuppeteerAdapter) clearCookies(ctx context.Context) (*ToolResult, error) {
	err := a.browser.ClearCookies(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Cleared all cookies"}},
	}, nil
}

// FormatCookies formats cookies for display
func FormatCookies(cookies []Cookie) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d cookies:\n\n", len(cookies)))

	for _, c := range cookies {
		sb.WriteString(fmt.Sprintf("- %s=%s (domain: %s)\n", c.Name, c.Value, c.Domain))
	}

	return sb.String()
}
