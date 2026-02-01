// Package adapters provides MCP server adapters.
// This file implements the SVGMaker MCP server adapter for AI-powered SVG generation.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SVGMakerConfig configures the SVGMaker adapter.
type SVGMakerConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

// DefaultSVGMakerConfig returns default configuration.
func DefaultSVGMakerConfig() SVGMakerConfig {
	return SVGMakerConfig{
		BaseURL: "https://api.svgmaker.io/v1",
		Timeout: 60 * time.Second,
	}
}

// SVGMakerAdapter implements the SVGMaker MCP server.
type SVGMakerAdapter struct {
	config     SVGMakerConfig
	httpClient *http.Client
}

// NewSVGMakerAdapter creates a new SVGMaker adapter.
func NewSVGMakerAdapter(config SVGMakerConfig) *SVGMakerAdapter {
	if config.BaseURL == "" {
		config.BaseURL = DefaultSVGMakerConfig().BaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultSVGMakerConfig().Timeout
	}
	return &SVGMakerAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *SVGMakerAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "svgmaker",
		Version:     "1.0.0",
		Description: "AI-powered SVG generation, editing, and optimization",
		Capabilities: []string{
			"generate_svg",
			"edit_svg",
			"optimize_svg",
			"image_to_svg",
			"text_to_icon",
			"svg_to_png",
		},
	}
}

// ListTools returns available tools.
func (a *SVGMakerAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "svg_generate",
			Description: "Generate an SVG from a text description using AI",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "Description of the SVG to generate",
					},
					"style": map[string]interface{}{
						"type":        "string",
						"description": "Visual style (e.g., minimal, detailed, flat, gradient)",
						"enum":        []string{"minimal", "detailed", "flat", "gradient", "outline", "filled"},
						"default":     "minimal",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "SVG width",
						"default":     256,
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "SVG height",
						"default":     256,
					},
					"color_scheme": map[string]interface{}{
						"type":        "string",
						"description": "Color scheme (e.g., monochrome, colorful, pastel)",
						"default":     "colorful",
					},
				},
				"required": []string{"prompt"},
			},
		},
		{
			Name:        "svg_edit",
			Description: "Edit an existing SVG based on instructions",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"svg": map[string]interface{}{
						"type":        "string",
						"description": "The SVG content to edit",
					},
					"instructions": map[string]interface{}{
						"type":        "string",
						"description": "Natural language instructions for editing",
					},
				},
				"required": []string{"svg", "instructions"},
			},
		},
		{
			Name:        "svg_optimize",
			Description: "Optimize an SVG for smaller file size",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"svg": map[string]interface{}{
						"type":        "string",
						"description": "The SVG content to optimize",
					},
					"precision": map[string]interface{}{
						"type":        "integer",
						"description": "Decimal precision for coordinates (0-8)",
						"default":     2,
					},
					"remove_metadata": map[string]interface{}{
						"type":        "boolean",
						"description": "Remove metadata and comments",
						"default":     true,
					},
					"minify": map[string]interface{}{
						"type":        "boolean",
						"description": "Minify the output",
						"default":     true,
					},
				},
				"required": []string{"svg"},
			},
		},
		{
			Name:        "svg_from_image",
			Description: "Convert a raster image to SVG (vectorize)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Base64-encoded image to convert",
					},
					"mode": map[string]interface{}{
						"type":        "string",
						"description": "Conversion mode",
						"enum":        []string{"color", "grayscale", "monochrome", "silhouette"},
						"default":     "color",
					},
					"detail_level": map[string]interface{}{
						"type":        "string",
						"description": "Level of detail in vectorization",
						"enum":        []string{"low", "medium", "high"},
						"default":     "medium",
					},
					"smooth_curves": map[string]interface{}{
						"type":        "boolean",
						"description": "Apply curve smoothing",
						"default":     true,
					},
				},
				"required": []string{"image"},
			},
		},
		{
			Name:        "svg_icon",
			Description: "Generate an icon SVG from a concept",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"concept": map[string]interface{}{
						"type":        "string",
						"description": "Icon concept (e.g., 'settings', 'user', 'home')",
					},
					"style": map[string]interface{}{
						"type":        "string",
						"description": "Icon style",
						"enum":        []string{"outline", "filled", "duotone", "rounded"},
						"default":     "outline",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Icon size in pixels",
						"default":     24,
					},
					"stroke_width": map[string]interface{}{
						"type":        "number",
						"description": "Stroke width for outline icons",
						"default":     2.0,
					},
				},
				"required": []string{"concept"},
			},
		},
		{
			Name:        "svg_to_png",
			Description: "Convert SVG to PNG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"svg": map[string]interface{}{
						"type":        "string",
						"description": "The SVG content to convert",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Output width in pixels",
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Output height in pixels",
					},
					"background": map[string]interface{}{
						"type":        "string",
						"description": "Background color (hex or 'transparent')",
						"default":     "transparent",
					},
				},
				"required": []string{"svg"},
			},
		},
		{
			Name:        "svg_analyze",
			Description: "Analyze an SVG and return its structure",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"svg": map[string]interface{}{
						"type":        "string",
						"description": "The SVG content to analyze",
					},
				},
				"required": []string{"svg"},
			},
		},
		{
			Name:        "svg_combine",
			Description: "Combine multiple SVGs into one",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"svgs": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Array of SVG contents to combine",
					},
					"layout": map[string]interface{}{
						"type":        "string",
						"description": "Layout arrangement",
						"enum":        []string{"horizontal", "vertical", "grid", "overlay"},
						"default":     "horizontal",
					},
					"spacing": map[string]interface{}{
						"type":        "integer",
						"description": "Spacing between elements",
						"default":     10,
					},
				},
				"required": []string{"svgs"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *SVGMakerAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "svg_generate":
		return a.generateSVG(ctx, args)
	case "svg_edit":
		return a.editSVG(ctx, args)
	case "svg_optimize":
		return a.optimizeSVG(ctx, args)
	case "svg_from_image":
		return a.imageToSVG(ctx, args)
	case "svg_icon":
		return a.generateIcon(ctx, args)
	case "svg_to_png":
		return a.svgToPNG(ctx, args)
	case "svg_analyze":
		return a.analyzeSVG(ctx, args)
	case "svg_combine":
		return a.combineSVGs(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *SVGMakerAdapter) generateSVG(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	prompt, _ := args["prompt"].(string)
	style, _ := args["style"].(string)
	if style == "" {
		style = "minimal"
	}
	width := getIntArg(args, "width", 256)
	height := getIntArg(args, "height", 256)
	colorScheme, _ := args["color_scheme"].(string)
	if colorScheme == "" {
		colorScheme = "colorful"
	}

	payload := map[string]interface{}{
		"prompt":       prompt,
		"style":        style,
		"width":        width,
		"height":       height,
		"color_scheme": colorScheme,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/generate", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGGenerateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Generated SVG for: %s (style: %s, %dx%d)", prompt, style, width, height)},
			{Type: "text", MimeType: "image/svg+xml", Text: result.SVG},
		},
	}, nil
}

func (a *SVGMakerAdapter) editSVG(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	svg, _ := args["svg"].(string)
	instructions, _ := args["instructions"].(string)

	payload := map[string]interface{}{
		"svg":          svg,
		"instructions": instructions,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/edit", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGEditResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Edited SVG: %s", instructions)},
			{Type: "text", MimeType: "image/svg+xml", Text: result.SVG},
		},
	}, nil
}

func (a *SVGMakerAdapter) optimizeSVG(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	svg, _ := args["svg"].(string)
	precision := getIntArg(args, "precision", 2)
	removeMetadata := getBoolArg(args, "remove_metadata", true)
	minify := getBoolArg(args, "minify", true)

	payload := map[string]interface{}{
		"svg":             svg,
		"precision":       precision,
		"remove_metadata": removeMetadata,
		"minify":          minify,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/optimize", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGOptimizeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Optimized SVG: %d bytes -> %d bytes (%.1f%% reduction)", result.OriginalSize, result.OptimizedSize, result.Reduction)},
			{Type: "text", MimeType: "image/svg+xml", Text: result.SVG},
		},
	}, nil
}

func (a *SVGMakerAdapter) imageToSVG(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	image, _ := args["image"].(string)
	mode, _ := args["mode"].(string)
	if mode == "" {
		mode = "color"
	}
	detailLevel, _ := args["detail_level"].(string)
	if detailLevel == "" {
		detailLevel = "medium"
	}
	smoothCurves := getBoolArg(args, "smooth_curves", true)

	payload := map[string]interface{}{
		"image":         image,
		"mode":          mode,
		"detail_level":  detailLevel,
		"smooth_curves": smoothCurves,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/vectorize", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGVectorizeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Vectorized image (mode: %s, detail: %s, paths: %d)", mode, detailLevel, result.PathCount)},
			{Type: "text", MimeType: "image/svg+xml", Text: result.SVG},
		},
	}, nil
}

func (a *SVGMakerAdapter) generateIcon(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	concept, _ := args["concept"].(string)
	style, _ := args["style"].(string)
	if style == "" {
		style = "outline"
	}
	size := getIntArg(args, "size", 24)
	strokeWidth := getFloatArg(args, "stroke_width", 2.0)

	payload := map[string]interface{}{
		"concept":      concept,
		"style":        style,
		"size":         size,
		"stroke_width": strokeWidth,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/icon", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGIconResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Generated icon: %s (style: %s, size: %dpx)", concept, style, size)},
			{Type: "text", MimeType: "image/svg+xml", Text: result.SVG},
		},
	}, nil
}

func (a *SVGMakerAdapter) svgToPNG(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	svg, _ := args["svg"].(string)
	width := getIntArg(args, "width", 0)
	height := getIntArg(args, "height", 0)
	background, _ := args["background"].(string)
	if background == "" {
		background = "transparent"
	}

	payload := map[string]interface{}{
		"svg":        svg,
		"background": background,
	}
	if width > 0 {
		payload["width"] = width
	}
	if height > 0 {
		payload["height"] = height
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/convert/png", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGConvertResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Converted to PNG (%dx%d)", result.Width, result.Height)},
			{Type: "image", MimeType: "image/png", Data: result.Image},
		},
	}, nil
}

func (a *SVGMakerAdapter) analyzeSVG(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	svg, _ := args["svg"].(string)

	payload := map[string]interface{}{
		"svg": svg,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/analyze", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGAnalyzeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString("SVG Analysis:\n\n")
	sb.WriteString(fmt.Sprintf("Dimensions: %dx%d\n", result.Width, result.Height))
	sb.WriteString(fmt.Sprintf("ViewBox: %s\n", result.ViewBox))
	sb.WriteString(fmt.Sprintf("File Size: %d bytes\n\n", result.FileSize))
	sb.WriteString("Elements:\n")
	for elemType, count := range result.Elements {
		sb.WriteString(fmt.Sprintf("  - %s: %d\n", elemType, count))
	}
	sb.WriteString(fmt.Sprintf("\nColors used: %d\n", len(result.Colors)))
	for _, color := range result.Colors {
		sb.WriteString(fmt.Sprintf("  - %s\n", color))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SVGMakerAdapter) combineSVGs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	svgsRaw, _ := args["svgs"].([]interface{})
	layout, _ := args["layout"].(string)
	if layout == "" {
		layout = "horizontal"
	}
	spacing := getIntArg(args, "spacing", 10)

	var svgs []string
	for _, s := range svgsRaw {
		if str, ok := s.(string); ok {
			svgs = append(svgs, str)
		}
	}

	payload := map[string]interface{}{
		"svgs":    svgs,
		"layout":  layout,
		"spacing": spacing,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/combine", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SVGCombineResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Combined %d SVGs (layout: %s)", len(svgs), layout)},
			{Type: "text", MimeType: "image/svg+xml", Text: result.SVG},
		},
	}, nil
}

func (a *SVGMakerAdapter) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) ([]byte, error) {
	reqURL := a.config.BaseURL + endpoint

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
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

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

func getBoolArg(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}

// SVGMaker API response types

// SVGGenerateResponse represents a generate response.
type SVGGenerateResponse struct {
	SVG    string `json:"svg"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SVGEditResponse represents an edit response.
type SVGEditResponse struct {
	SVG     string   `json:"svg"`
	Changes []string `json:"changes"`
}

// SVGOptimizeResponse represents an optimize response.
type SVGOptimizeResponse struct {
	SVG           string  `json:"svg"`
	OriginalSize  int     `json:"original_size"`
	OptimizedSize int     `json:"optimized_size"`
	Reduction     float64 `json:"reduction"`
}

// SVGVectorizeResponse represents a vectorize response.
type SVGVectorizeResponse struct {
	SVG       string `json:"svg"`
	PathCount int    `json:"path_count"`
}

// SVGIconResponse represents an icon response.
type SVGIconResponse struct {
	SVG  string `json:"svg"`
	Name string `json:"name"`
}

// SVGConvertResponse represents a convert response.
type SVGConvertResponse struct {
	Image  string `json:"image"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SVGAnalyzeResponse represents an analyze response.
type SVGAnalyzeResponse struct {
	Width    int            `json:"width"`
	Height   int            `json:"height"`
	ViewBox  string         `json:"viewBox"`
	FileSize int            `json:"file_size"`
	Elements map[string]int `json:"elements"`
	Colors   []string       `json:"colors"`
}

// SVGCombineResponse represents a combine response.
type SVGCombineResponse struct {
	SVG    string `json:"svg"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
