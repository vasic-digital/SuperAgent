// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// SVGMakerAdapterConfig holds configuration for SVGMaker MCP adapter
type SVGMakerAdapterConfig struct {
	// DefaultWidth is the default SVG width
	DefaultWidth int `json:"default_width,omitempty"`
	// DefaultHeight is the default SVG height
	DefaultHeight int `json:"default_height,omitempty"`
	// MaxWidth is the maximum allowed width
	MaxWidth int `json:"max_width,omitempty"`
	// MaxHeight is the maximum allowed height
	MaxHeight int `json:"max_height,omitempty"`
	// DefaultStrokeWidth is the default stroke width
	DefaultStrokeWidth float64 `json:"default_stroke_width,omitempty"`
	// DefaultFontFamily is the default font family
	DefaultFontFamily string `json:"default_font_family,omitempty"`
	// DefaultFontSize is the default font size
	DefaultFontSize int `json:"default_font_size,omitempty"`
}

// DefaultSVGMakerAdapterConfig returns default configuration
func DefaultSVGMakerAdapterConfig() SVGMakerAdapterConfig {
	return SVGMakerAdapterConfig{
		DefaultWidth:       800,
		DefaultHeight:      600,
		MaxWidth:           4000,
		MaxHeight:          4000,
		DefaultStrokeWidth: 1.0,
		DefaultFontFamily:  "Arial, sans-serif",
		DefaultFontSize:    12,
	}
}

// SVGMakerAdapter implements MCP adapter for SVG generation
type SVGMakerAdapter struct {
	config      SVGMakerAdapterConfig
	initialized bool
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// NewSVGMakerAdapter creates a new SVGMaker MCP adapter
func NewSVGMakerAdapter(config SVGMakerAdapterConfig, logger *logrus.Logger) *SVGMakerAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.DefaultWidth <= 0 {
		config.DefaultWidth = 800
	}
	if config.DefaultHeight <= 0 {
		config.DefaultHeight = 600
	}
	if config.MaxWidth <= 0 {
		config.MaxWidth = 4000
	}
	if config.MaxHeight <= 0 {
		config.MaxHeight = 4000
	}
	if config.DefaultStrokeWidth <= 0 {
		config.DefaultStrokeWidth = 1.0
	}
	if config.DefaultFontFamily == "" {
		config.DefaultFontFamily = "Arial, sans-serif"
	}
	if config.DefaultFontSize <= 0 {
		config.DefaultFontSize = 12
	}

	return &SVGMakerAdapter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the adapter
func (s *SVGMakerAdapter) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.initialized = true
	s.logger.Info("SVGMaker adapter initialized")
	return nil
}

// Health returns health status
func (s *SVGMakerAdapter) Health(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return fmt.Errorf("SVGMaker adapter not initialized")
	}
	return nil
}

// Close closes the adapter
func (s *SVGMakerAdapter) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = false
	return nil
}

// SVGDocument represents an SVG document
type SVGDocument struct {
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	ViewBox    string       `json:"viewbox,omitempty"`
	Background string       `json:"background,omitempty"`
	Elements   []SVGElement `json:"elements"`
	Defs       []SVGDef     `json:"defs,omitempty"`
}

// SVGElement represents an SVG element
type SVGElement struct {
	Type       string                 `json:"type"` // rect, circle, ellipse, line, polyline, polygon, path, text, group, image
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Content    string                 `json:"content,omitempty"`  // For text elements
	Children   []SVGElement           `json:"children,omitempty"` // For groups
}

// SVGDef represents an SVG definition (gradients, patterns, etc.)
type SVGDef struct {
	Type       string                 `json:"type"` // linearGradient, radialGradient, pattern, clipPath, mask
	ID         string                 `json:"id"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Stops      []GradientStop         `json:"stops,omitempty"` // For gradients
}

// GradientStop represents a gradient stop
type GradientStop struct {
	Offset  string  `json:"offset"`
	Color   string  `json:"color"`
	Opacity float64 `json:"opacity,omitempty"`
}

// CreateSVG creates an SVG document from a specification
func (s *SVGMakerAdapter) CreateSVG(ctx context.Context, doc *SVGDocument) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return "", fmt.Errorf("adapter not initialized")
	}

	// Apply defaults
	if doc.Width <= 0 {
		doc.Width = s.config.DefaultWidth
	}
	if doc.Height <= 0 {
		doc.Height = s.config.DefaultHeight
	}

	// Validate dimensions
	if doc.Width > s.config.MaxWidth {
		return "", fmt.Errorf("width exceeds maximum (%d > %d)", doc.Width, s.config.MaxWidth)
	}
	if doc.Height > s.config.MaxHeight {
		return "", fmt.Errorf("height exceeds maximum (%d > %d)", doc.Height, s.config.MaxHeight)
	}

	var sb strings.Builder

	// Write XML declaration and SVG opening tag
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d"`, doc.Width, doc.Height))

	if doc.ViewBox != "" {
		sb.WriteString(fmt.Sprintf(` viewBox="%s"`, escapeXML(doc.ViewBox)))
	} else {
		sb.WriteString(fmt.Sprintf(` viewBox="0 0 %d %d"`, doc.Width, doc.Height))
	}
	sb.WriteString(">\n")

	// Write defs if any
	if len(doc.Defs) > 0 {
		sb.WriteString("  <defs>\n")
		for _, def := range doc.Defs {
			sb.WriteString(s.renderDef(def, 4))
		}
		sb.WriteString("  </defs>\n")
	}

	// Write background if specified
	if doc.Background != "" {
		sb.WriteString(fmt.Sprintf(`  <rect width="100%%" height="100%%" fill="%s"/>`, escapeXML(doc.Background)))
		sb.WriteString("\n")
	}

	// Write elements
	for _, elem := range doc.Elements {
		sb.WriteString(s.renderElement(elem, 2))
	}

	sb.WriteString("</svg>")

	return sb.String(), nil
}

// renderDef renders an SVG definition
func (s *SVGMakerAdapter) renderDef(def SVGDef, indent int) string {
	prefix := strings.Repeat(" ", indent)
	var sb strings.Builder

	switch def.Type {
	case "linearGradient", "radialGradient":
		sb.WriteString(fmt.Sprintf(`%s<%s id="%s"`, prefix, def.Type, escapeXML(def.ID)))
		for k, v := range def.Attributes {
			sb.WriteString(fmt.Sprintf(` %s="%v"`, k, escapeXML(fmt.Sprintf("%v", v))))
		}
		sb.WriteString(">\n")
		for _, stop := range def.Stops {
			opacity := ""
			if stop.Opacity > 0 && stop.Opacity < 1 {
				opacity = fmt.Sprintf(` stop-opacity="%.2f"`, stop.Opacity)
			}
			sb.WriteString(fmt.Sprintf(`%s  <stop offset="%s" stop-color="%s"%s/>`,
				prefix, escapeXML(stop.Offset), escapeXML(stop.Color), opacity))
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s</%s>\n", prefix, def.Type))

	case "pattern":
		sb.WriteString(fmt.Sprintf(`%s<pattern id="%s"`, prefix, escapeXML(def.ID)))
		for k, v := range def.Attributes {
			sb.WriteString(fmt.Sprintf(` %s="%v"`, k, escapeXML(fmt.Sprintf("%v", v))))
		}
		sb.WriteString("/>\n")

	case "clipPath":
		sb.WriteString(fmt.Sprintf(`%s<clipPath id="%s"`, prefix, escapeXML(def.ID)))
		for k, v := range def.Attributes {
			sb.WriteString(fmt.Sprintf(` %s="%v"`, k, escapeXML(fmt.Sprintf("%v", v))))
		}
		sb.WriteString("/>\n")
	}

	return sb.String()
}

// renderElement renders an SVG element
func (s *SVGMakerAdapter) renderElement(elem SVGElement, indent int) string {
	prefix := strings.Repeat(" ", indent)
	var sb strings.Builder

	switch elem.Type {
	case "rect":
		sb.WriteString(fmt.Sprintf("%s<rect", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"x", "y", "width", "height", "rx", "ry", "fill", "stroke", "stroke-width", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "circle":
		sb.WriteString(fmt.Sprintf("%s<circle", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"cx", "cy", "r", "fill", "stroke", "stroke-width", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "ellipse":
		sb.WriteString(fmt.Sprintf("%s<ellipse", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"cx", "cy", "rx", "ry", "fill", "stroke", "stroke-width", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "line":
		sb.WriteString(fmt.Sprintf("%s<line", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"x1", "y1", "x2", "y2", "stroke", "stroke-width", "stroke-linecap", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "polyline":
		sb.WriteString(fmt.Sprintf("%s<polyline", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"points", "fill", "stroke", "stroke-width", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "polygon":
		sb.WriteString(fmt.Sprintf("%s<polygon", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"points", "fill", "stroke", "stroke-width", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "path":
		sb.WriteString(fmt.Sprintf("%s<path", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"d", "fill", "stroke", "stroke-width", "stroke-linecap", "stroke-linejoin", "fill-rule", "opacity", "transform"})
		sb.WriteString("/>\n")

	case "text":
		sb.WriteString(fmt.Sprintf("%s<text", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"x", "y", "font-family", "font-size", "font-weight", "fill", "text-anchor", "dominant-baseline", "opacity", "transform"})
		sb.WriteString(">")
		sb.WriteString(escapeXML(elem.Content))
		sb.WriteString("</text>\n")

	case "group", "g":
		sb.WriteString(fmt.Sprintf("%s<g", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"id", "class", "fill", "stroke", "stroke-width", "opacity", "transform", "clip-path"})
		sb.WriteString(">\n")
		for _, child := range elem.Children {
			sb.WriteString(s.renderElement(child, indent+2))
		}
		sb.WriteString(fmt.Sprintf("%s</g>\n", prefix))

	case "image":
		sb.WriteString(fmt.Sprintf("%s<image", prefix))
		s.writeAttributes(&sb, elem.Attributes, []string{"x", "y", "width", "height", "href", "xlink:href", "preserveAspectRatio", "opacity", "transform"})
		sb.WriteString("/>\n")

	default:
		// Generic element
		sb.WriteString(fmt.Sprintf("%s<%s", prefix, elem.Type))
		for k, v := range elem.Attributes {
			sb.WriteString(fmt.Sprintf(` %s="%v"`, k, escapeXML(fmt.Sprintf("%v", v))))
		}
		if elem.Content != "" || len(elem.Children) > 0 {
			sb.WriteString(">")
			if elem.Content != "" {
				sb.WriteString(escapeXML(elem.Content))
			}
			if len(elem.Children) > 0 {
				sb.WriteString("\n")
				for _, child := range elem.Children {
					sb.WriteString(s.renderElement(child, indent+2))
				}
				sb.WriteString(prefix)
			}
			sb.WriteString(fmt.Sprintf("</%s>\n", elem.Type))
		} else {
			sb.WriteString("/>\n")
		}
	}

	return sb.String()
}

// writeAttributes writes attributes to the string builder
func (s *SVGMakerAdapter) writeAttributes(sb *strings.Builder, attrs map[string]interface{}, allowedAttrs []string) {
	// Write allowed attributes in order
	for _, key := range allowedAttrs {
		if val, ok := attrs[key]; ok {
			sb.WriteString(fmt.Sprintf(` %s="%v"`, key, escapeXML(fmt.Sprintf("%v", val))))
		}
	}
	// Write any remaining attributes
	for key, val := range attrs {
		found := false
		for _, allowed := range allowedAttrs {
			if key == allowed {
				found = true
				break
			}
		}
		if !found {
			sb.WriteString(fmt.Sprintf(` %s="%v"`, key, escapeXML(fmt.Sprintf("%v", val))))
		}
	}
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// CreateRectangle creates a rectangle SVG
func (s *SVGMakerAdapter) CreateRectangle(ctx context.Context, x, y, width, height float64, fill, stroke string, strokeWidth float64) (string, error) {
	doc := &SVGDocument{
		Width:  int(x + width + 10),
		Height: int(y + height + 10),
		Elements: []SVGElement{
			{
				Type: "rect",
				Attributes: map[string]interface{}{
					"x":            x,
					"y":            y,
					"width":        width,
					"height":       height,
					"fill":         fill,
					"stroke":       stroke,
					"stroke-width": strokeWidth,
				},
			},
		},
	}
	return s.CreateSVG(ctx, doc)
}

// CreateCircle creates a circle SVG
func (s *SVGMakerAdapter) CreateCircle(ctx context.Context, cx, cy, r float64, fill, stroke string, strokeWidth float64) (string, error) {
	doc := &SVGDocument{
		Width:  int(cx + r + 10),
		Height: int(cy + r + 10),
		Elements: []SVGElement{
			{
				Type: "circle",
				Attributes: map[string]interface{}{
					"cx":           cx,
					"cy":           cy,
					"r":            r,
					"fill":         fill,
					"stroke":       stroke,
					"stroke-width": strokeWidth,
				},
			},
		},
	}
	return s.CreateSVG(ctx, doc)
}

// CreateLine creates a line SVG
func (s *SVGMakerAdapter) CreateLine(ctx context.Context, x1, y1, x2, y2 float64, stroke string, strokeWidth float64) (string, error) {
	maxX := math.Max(x1, x2)
	maxY := math.Max(y1, y2)
	doc := &SVGDocument{
		Width:  int(maxX + 10),
		Height: int(maxY + 10),
		Elements: []SVGElement{
			{
				Type: "line",
				Attributes: map[string]interface{}{
					"x1":           x1,
					"y1":           y1,
					"x2":           x2,
					"y2":           y2,
					"stroke":       stroke,
					"stroke-width": strokeWidth,
				},
			},
		},
	}
	return s.CreateSVG(ctx, doc)
}

// CreateText creates a text SVG
func (s *SVGMakerAdapter) CreateText(ctx context.Context, x, y float64, text, fontFamily string, fontSize int, fill string) (string, error) {
	if fontFamily == "" {
		fontFamily = s.config.DefaultFontFamily
	}
	if fontSize <= 0 {
		fontSize = s.config.DefaultFontSize
	}

	doc := &SVGDocument{
		Width:  int(x) + len(text)*fontSize/2 + 20,
		Height: int(y) + fontSize + 10,
		Elements: []SVGElement{
			{
				Type:    "text",
				Content: text,
				Attributes: map[string]interface{}{
					"x":           x,
					"y":           y,
					"font-family": fontFamily,
					"font-size":   fontSize,
					"fill":        fill,
				},
			},
		},
	}
	return s.CreateSVG(ctx, doc)
}

// CreatePath creates a path SVG
func (s *SVGMakerAdapter) CreatePath(ctx context.Context, d, fill, stroke string, strokeWidth float64) (string, error) {
	doc := &SVGDocument{
		Width:  s.config.DefaultWidth,
		Height: s.config.DefaultHeight,
		Elements: []SVGElement{
			{
				Type: "path",
				Attributes: map[string]interface{}{
					"d":            d,
					"fill":         fill,
					"stroke":       stroke,
					"stroke-width": strokeWidth,
				},
			},
		},
	}
	return s.CreateSVG(ctx, doc)
}

// ChartData represents data for chart generation
type ChartData struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
	Colors []string  `json:"colors,omitempty"`
	Title  string    `json:"title,omitempty"`
}

// CreateBarChart creates a bar chart SVG
func (s *SVGMakerAdapter) CreateBarChart(ctx context.Context, data *ChartData, width, height int) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return "", fmt.Errorf("adapter not initialized")
	}

	if len(data.Labels) == 0 || len(data.Values) == 0 {
		return "", fmt.Errorf("chart data cannot be empty")
	}

	if len(data.Labels) != len(data.Values) {
		return "", fmt.Errorf("labels and values must have same length")
	}

	if width <= 0 {
		width = s.config.DefaultWidth
	}
	if height <= 0 {
		height = s.config.DefaultHeight
	}

	// Find max value for scaling
	maxVal := data.Values[0]
	for _, v := range data.Values {
		if v > maxVal {
			maxVal = v
		}
	}

	padding := 60
	chartWidth := width - padding*2
	chartHeight := height - padding*2
	barWidth := float64(chartWidth) / float64(len(data.Values)) * 0.8
	barGap := float64(chartWidth) / float64(len(data.Values)) * 0.2

	elements := []SVGElement{}

	// Default colors if not provided
	colors := data.Colors
	if len(colors) == 0 {
		colors = []string{"#4285f4", "#ea4335", "#fbbc04", "#34a853", "#673ab7", "#ff6d00", "#795548"}
	}

	// Add bars
	for i, val := range data.Values {
		barHeight := (val / maxVal) * float64(chartHeight)
		x := float64(padding) + float64(i)*(barWidth+barGap)
		y := float64(height-padding) - barHeight
		color := colors[i%len(colors)]

		elements = append(elements, SVGElement{
			Type: "rect",
			Attributes: map[string]interface{}{
				"x":      x,
				"y":      y,
				"width":  barWidth,
				"height": barHeight,
				"fill":   color,
			},
		})

		// Add label
		labelX := x + barWidth/2
		labelY := float64(height - padding + 20)
		elements = append(elements, SVGElement{
			Type:    "text",
			Content: data.Labels[i],
			Attributes: map[string]interface{}{
				"x":           labelX,
				"y":           labelY,
				"font-family": s.config.DefaultFontFamily,
				"font-size":   10,
				"text-anchor": "middle",
				"fill":        "#333",
			},
		})

		// Add value
		elements = append(elements, SVGElement{
			Type:    "text",
			Content: fmt.Sprintf("%.1f", val),
			Attributes: map[string]interface{}{
				"x":           labelX,
				"y":           y - 5,
				"font-family": s.config.DefaultFontFamily,
				"font-size":   10,
				"text-anchor": "middle",
				"fill":        "#333",
			},
		})
	}

	// Add title if provided
	if data.Title != "" {
		elements = append(elements, SVGElement{
			Type:    "text",
			Content: data.Title,
			Attributes: map[string]interface{}{
				"x":           float64(width) / 2,
				"y":           30,
				"font-family": s.config.DefaultFontFamily,
				"font-size":   16,
				"font-weight": "bold",
				"text-anchor": "middle",
				"fill":        "#333",
			},
		})
	}

	doc := &SVGDocument{
		Width:      width,
		Height:     height,
		Background: "#fff",
		Elements:   elements,
	}

	return s.CreateSVG(ctx, doc)
}

// CreatePieChart creates a pie chart SVG
func (s *SVGMakerAdapter) CreatePieChart(ctx context.Context, data *ChartData, width, height int) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return "", fmt.Errorf("adapter not initialized")
	}

	if len(data.Labels) == 0 || len(data.Values) == 0 {
		return "", fmt.Errorf("chart data cannot be empty")
	}

	if width <= 0 {
		width = s.config.DefaultWidth
	}
	if height <= 0 {
		height = s.config.DefaultHeight
	}

	// Calculate total
	total := 0.0
	for _, v := range data.Values {
		total += v
	}

	cx := float64(width) / 2
	cy := float64(height) / 2
	radius := math.Min(float64(width), float64(height))/2 - 40

	// Default colors
	colors := data.Colors
	if len(colors) == 0 {
		colors = []string{"#4285f4", "#ea4335", "#fbbc04", "#34a853", "#673ab7", "#ff6d00", "#795548"}
	}

	elements := []SVGElement{}
	startAngle := -math.Pi / 2

	for i, val := range data.Values {
		percentage := val / total
		angle := percentage * 2 * math.Pi

		// Calculate arc endpoints
		x1 := cx + radius*math.Cos(startAngle)
		y1 := cy + radius*math.Sin(startAngle)
		x2 := cx + radius*math.Cos(startAngle+angle)
		y2 := cy + radius*math.Sin(startAngle+angle)

		// Large arc flag
		largeArc := 0
		if angle > math.Pi {
			largeArc = 1
		}

		color := colors[i%len(colors)]

		// Create pie slice path
		d := fmt.Sprintf("M %.2f %.2f L %.2f %.2f A %.2f %.2f 0 %d 1 %.2f %.2f Z",
			cx, cy, x1, y1, radius, radius, largeArc, x2, y2)

		elements = append(elements, SVGElement{
			Type: "path",
			Attributes: map[string]interface{}{
				"d":            d,
				"fill":         color,
				"stroke":       "#fff",
				"stroke-width": 2,
			},
		})

		// Add label line and text
		midAngle := startAngle + angle/2
		labelRadius := radius * 1.2
		labelX := cx + labelRadius*math.Cos(midAngle)
		labelY := cy + labelRadius*math.Sin(midAngle)

		anchor := "start"
		if labelX < cx {
			anchor = "end"
		}

		elements = append(elements, SVGElement{
			Type:    "text",
			Content: fmt.Sprintf("%s (%.1f%%)", data.Labels[i], percentage*100),
			Attributes: map[string]interface{}{
				"x":           labelX,
				"y":           labelY,
				"font-family": s.config.DefaultFontFamily,
				"font-size":   10,
				"text-anchor": anchor,
				"fill":        "#333",
			},
		})

		startAngle += angle
	}

	// Add title
	if data.Title != "" {
		elements = append(elements, SVGElement{
			Type:    "text",
			Content: data.Title,
			Attributes: map[string]interface{}{
				"x":           float64(width) / 2,
				"y":           20,
				"font-family": s.config.DefaultFontFamily,
				"font-size":   16,
				"font-weight": "bold",
				"text-anchor": "middle",
				"fill":        "#333",
			},
		})
	}

	doc := &SVGDocument{
		Width:      width,
		Height:     height,
		Background: "#fff",
		Elements:   elements,
	}

	return s.CreateSVG(ctx, doc)
}

// ValidateSVG validates an SVG string
func (s *SVGMakerAdapter) ValidateSVG(ctx context.Context, svg string) (bool, []string) {
	var errors []string

	// Check for basic SVG structure
	if !strings.Contains(svg, "<svg") {
		errors = append(errors, "missing <svg> tag")
	}
	if !strings.Contains(svg, "</svg>") {
		errors = append(errors, "missing </svg> closing tag")
	}

	// Check for namespace
	if !strings.Contains(svg, `xmlns="http://www.w3.org/2000/svg"`) && !strings.Contains(svg, `xmlns='http://www.w3.org/2000/svg'`) {
		errors = append(errors, "missing xmlns namespace declaration")
	}

	// Check for balanced tags
	openTags := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)[^/>]*>`).FindAllStringSubmatch(svg, -1)
	closeTags := regexp.MustCompile(`</([a-zA-Z][a-zA-Z0-9]*)>`).FindAllStringSubmatch(svg, -1)

	openCount := make(map[string]int)
	closeCount := make(map[string]int)

	for _, match := range openTags {
		openCount[match[1]]++
	}
	for _, match := range closeTags {
		closeCount[match[1]]++
	}

	// Check if counts match (accounting for self-closing tags)
	for tag, count := range openCount {
		if closeCount[tag] != count {
			// Check if self-closing
			selfClosing := regexp.MustCompile(`<`+tag+`[^>]*/>`).FindAllString(svg, -1)
			if closeCount[tag]+len(selfClosing) != count {
				errors = append(errors, fmt.Sprintf("unbalanced <%s> tags", tag))
			}
		}
	}

	return len(errors) == 0, errors
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (s *SVGMakerAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "svg_create",
			Description: "Create a custom SVG from a document specification",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "SVG width in pixels",
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "SVG height in pixels",
					},
					"background": map[string]interface{}{
						"type":        "string",
						"description": "Background color",
					},
					"elements": map[string]interface{}{
						"type":        "array",
						"description": "SVG elements to render",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"type":       map[string]interface{}{"type": "string"},
								"attributes": map[string]interface{}{"type": "object"},
								"content":    map[string]interface{}{"type": "string"},
							},
						},
					},
				},
			},
		},
		{
			Name:        "svg_rectangle",
			Description: "Create a simple rectangle SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"x":            map[string]interface{}{"type": "number", "description": "X position"},
					"y":            map[string]interface{}{"type": "number", "description": "Y position"},
					"width":        map[string]interface{}{"type": "number", "description": "Width"},
					"height":       map[string]interface{}{"type": "number", "description": "Height"},
					"fill":         map[string]interface{}{"type": "string", "description": "Fill color"},
					"stroke":       map[string]interface{}{"type": "string", "description": "Stroke color"},
					"stroke_width": map[string]interface{}{"type": "number", "description": "Stroke width"},
				},
				"required": []string{"width", "height"},
			},
		},
		{
			Name:        "svg_circle",
			Description: "Create a simple circle SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cx":           map[string]interface{}{"type": "number", "description": "Center X"},
					"cy":           map[string]interface{}{"type": "number", "description": "Center Y"},
					"r":            map[string]interface{}{"type": "number", "description": "Radius"},
					"fill":         map[string]interface{}{"type": "string", "description": "Fill color"},
					"stroke":       map[string]interface{}{"type": "string", "description": "Stroke color"},
					"stroke_width": map[string]interface{}{"type": "number", "description": "Stroke width"},
				},
				"required": []string{"r"},
			},
		},
		{
			Name:        "svg_line",
			Description: "Create a simple line SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"x1":           map[string]interface{}{"type": "number", "description": "Start X"},
					"y1":           map[string]interface{}{"type": "number", "description": "Start Y"},
					"x2":           map[string]interface{}{"type": "number", "description": "End X"},
					"y2":           map[string]interface{}{"type": "number", "description": "End Y"},
					"stroke":       map[string]interface{}{"type": "string", "description": "Stroke color"},
					"stroke_width": map[string]interface{}{"type": "number", "description": "Stroke width"},
				},
				"required": []string{"x1", "y1", "x2", "y2"},
			},
		},
		{
			Name:        "svg_text",
			Description: "Create text SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"x":           map[string]interface{}{"type": "number", "description": "X position"},
					"y":           map[string]interface{}{"type": "number", "description": "Y position"},
					"text":        map[string]interface{}{"type": "string", "description": "Text content"},
					"font_family": map[string]interface{}{"type": "string", "description": "Font family"},
					"font_size":   map[string]interface{}{"type": "integer", "description": "Font size"},
					"fill":        map[string]interface{}{"type": "string", "description": "Text color"},
				},
				"required": []string{"text"},
			},
		},
		{
			Name:        "svg_path",
			Description: "Create a path SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"d":            map[string]interface{}{"type": "string", "description": "Path data (d attribute)"},
					"fill":         map[string]interface{}{"type": "string", "description": "Fill color"},
					"stroke":       map[string]interface{}{"type": "string", "description": "Stroke color"},
					"stroke_width": map[string]interface{}{"type": "number", "description": "Stroke width"},
				},
				"required": []string{"d"},
			},
		},
		{
			Name:        "svg_bar_chart",
			Description: "Create a bar chart SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Bar labels",
					},
					"values": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "number"},
						"description": "Bar values",
					},
					"colors": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Bar colors",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Chart title",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Chart width",
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Chart height",
					},
				},
				"required": []string{"labels", "values"},
			},
		},
		{
			Name:        "svg_pie_chart",
			Description: "Create a pie chart SVG",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Slice labels",
					},
					"values": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "number"},
						"description": "Slice values",
					},
					"colors": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Slice colors",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Chart title",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Chart width",
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Chart height",
					},
				},
				"required": []string{"labels", "values"},
			},
		},
		{
			Name:        "svg_validate",
			Description: "Validate an SVG string",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"svg": map[string]interface{}{
						"type":        "string",
						"description": "SVG content to validate",
					},
				},
				"required": []string{"svg"},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (s *SVGMakerAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	s.mu.RLock()
	initialized := s.initialized
	s.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	switch toolName {
	case "svg_create":
		doc := &SVGDocument{}
		if w, ok := params["width"].(float64); ok {
			doc.Width = int(w)
		}
		if h, ok := params["height"].(float64); ok {
			doc.Height = int(h)
		}
		if bg, ok := params["background"].(string); ok {
			doc.Background = bg
		}
		if elems, ok := params["elements"].([]interface{}); ok {
			for _, e := range elems {
				if elem, ok := e.(map[string]interface{}); ok {
					svgElem := SVGElement{}
					if t, ok := elem["type"].(string); ok {
						svgElem.Type = t
					}
					if attrs, ok := elem["attributes"].(map[string]interface{}); ok {
						svgElem.Attributes = attrs
					}
					if content, ok := elem["content"].(string); ok {
						svgElem.Content = content
					}
					doc.Elements = append(doc.Elements, svgElem)
				}
			}
		}
		svg, err := s.CreateSVG(ctx, doc)
		return map[string]interface{}{"svg": svg}, err

	case "svg_rectangle":
		x, _ := params["x"].(float64)
		y, _ := params["y"].(float64)
		width, _ := params["width"].(float64)
		height, _ := params["height"].(float64)
		fill, _ := params["fill"].(string)
		stroke, _ := params["stroke"].(string)
		strokeWidth, _ := params["stroke_width"].(float64)
		if fill == "" {
			fill = "none"
		}
		if stroke == "" {
			stroke = "black"
		}
		if strokeWidth == 0 {
			strokeWidth = s.config.DefaultStrokeWidth
		}
		svg, err := s.CreateRectangle(ctx, x, y, width, height, fill, stroke, strokeWidth)
		return map[string]interface{}{"svg": svg}, err

	case "svg_circle":
		cx, _ := params["cx"].(float64)
		cy, _ := params["cy"].(float64)
		r, _ := params["r"].(float64)
		fill, _ := params["fill"].(string)
		stroke, _ := params["stroke"].(string)
		strokeWidth, _ := params["stroke_width"].(float64)
		if fill == "" {
			fill = "none"
		}
		if stroke == "" {
			stroke = "black"
		}
		if strokeWidth == 0 {
			strokeWidth = s.config.DefaultStrokeWidth
		}
		if cx == 0 {
			cx = r + 10
		}
		if cy == 0 {
			cy = r + 10
		}
		svg, err := s.CreateCircle(ctx, cx, cy, r, fill, stroke, strokeWidth)
		return map[string]interface{}{"svg": svg}, err

	case "svg_line":
		x1, _ := params["x1"].(float64)
		y1, _ := params["y1"].(float64)
		x2, _ := params["x2"].(float64)
		y2, _ := params["y2"].(float64)
		stroke, _ := params["stroke"].(string)
		strokeWidth, _ := params["stroke_width"].(float64)
		if stroke == "" {
			stroke = "black"
		}
		if strokeWidth == 0 {
			strokeWidth = s.config.DefaultStrokeWidth
		}
		svg, err := s.CreateLine(ctx, x1, y1, x2, y2, stroke, strokeWidth)
		return map[string]interface{}{"svg": svg}, err

	case "svg_text":
		x, _ := params["x"].(float64)
		y, _ := params["y"].(float64)
		text, _ := params["text"].(string)
		fontFamily, _ := params["font_family"].(string)
		fontSize := s.config.DefaultFontSize
		if fs, ok := params["font_size"].(float64); ok {
			fontSize = int(fs)
		}
		fill, _ := params["fill"].(string)
		if fill == "" {
			fill = "black"
		}
		if y == 0 {
			y = float64(fontSize) + 10
		}
		svg, err := s.CreateText(ctx, x, y, text, fontFamily, fontSize, fill)
		return map[string]interface{}{"svg": svg}, err

	case "svg_path":
		d, _ := params["d"].(string)
		fill, _ := params["fill"].(string)
		stroke, _ := params["stroke"].(string)
		strokeWidth, _ := params["stroke_width"].(float64)
		if fill == "" {
			fill = "none"
		}
		if stroke == "" {
			stroke = "black"
		}
		if strokeWidth == 0 {
			strokeWidth = s.config.DefaultStrokeWidth
		}
		svg, err := s.CreatePath(ctx, d, fill, stroke, strokeWidth)
		return map[string]interface{}{"svg": svg}, err

	case "svg_bar_chart":
		data := &ChartData{}
		if labels, ok := params["labels"].([]interface{}); ok {
			for _, l := range labels {
				if s, ok := l.(string); ok {
					data.Labels = append(data.Labels, s)
				}
			}
		}
		if values, ok := params["values"].([]interface{}); ok {
			for _, v := range values {
				if f, ok := v.(float64); ok {
					data.Values = append(data.Values, f)
				}
			}
		}
		if colors, ok := params["colors"].([]interface{}); ok {
			for _, c := range colors {
				if s, ok := c.(string); ok {
					data.Colors = append(data.Colors, s)
				}
			}
		}
		if title, ok := params["title"].(string); ok {
			data.Title = title
		}
		width := 0
		height := 0
		if w, ok := params["width"].(float64); ok {
			width = int(w)
		}
		if h, ok := params["height"].(float64); ok {
			height = int(h)
		}
		svg, err := s.CreateBarChart(ctx, data, width, height)
		return map[string]interface{}{"svg": svg}, err

	case "svg_pie_chart":
		data := &ChartData{}
		if labels, ok := params["labels"].([]interface{}); ok {
			for _, l := range labels {
				if s, ok := l.(string); ok {
					data.Labels = append(data.Labels, s)
				}
			}
		}
		if values, ok := params["values"].([]interface{}); ok {
			for _, v := range values {
				if f, ok := v.(float64); ok {
					data.Values = append(data.Values, f)
				}
			}
		}
		if colors, ok := params["colors"].([]interface{}); ok {
			for _, c := range colors {
				if s, ok := c.(string); ok {
					data.Colors = append(data.Colors, s)
				}
			}
		}
		if title, ok := params["title"].(string); ok {
			data.Title = title
		}
		width := 0
		height := 0
		if w, ok := params["width"].(float64); ok {
			width = int(w)
		}
		if h, ok := params["height"].(float64); ok {
			height = int(h)
		}
		svg, err := s.CreatePieChart(ctx, data, width, height)
		return map[string]interface{}{"svg": svg}, err

	case "svg_validate":
		svg, _ := params["svg"].(string)
		valid, errors := s.ValidateSVG(ctx, svg)
		return map[string]interface{}{"valid": valid, "errors": errors}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (s *SVGMakerAdapter) GetCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"name":           "svgmaker",
		"default_width":  s.config.DefaultWidth,
		"default_height": s.config.DefaultHeight,
		"max_width":      s.config.MaxWidth,
		"max_height":     s.config.MaxHeight,
		"tools":          len(s.GetMCPTools()),
	}
}

// MarshalJSON implements custom JSON marshaling
func (s *SVGMakerAdapter) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"initialized":  s.initialized,
		"capabilities": s.GetCapabilities(),
	})
}
