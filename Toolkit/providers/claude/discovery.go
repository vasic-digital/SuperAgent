// Package claude provides model discovery for Anthropic Claude API.
package claude

import (
	"context"
	"strings"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Discovery implements the ModelDiscovery interface for Claude.
type Discovery struct {
	client *Client
}

// NewDiscovery creates a new Claude model discovery instance.
func NewDiscovery(apiKey, version string) *Discovery {
	return &Discovery{
		client: NewClient(apiKey, version),
	}
}

// Discover discovers available models from Claude.
func (d *Discovery) Discover(ctx context.Context) ([]toolkit.ModelInfo, error) {
	models, err := d.client.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	var modelInfos []toolkit.ModelInfo
	for _, model := range models {
		modelInfo := d.convertToModelInfo(model)
		modelInfos = append(modelInfos, modelInfo)
	}

	return modelInfos, nil
}

// convertToModelInfo converts Claude model info to toolkit ModelInfo.
func (d *Discovery) convertToModelInfo(model ModelInfo) toolkit.ModelInfo {
	capabilities := d.inferCapabilities(model.ID, model.Type)

	return toolkit.ModelInfo{
		ID:           model.ID,
		Name:         d.formatModelName(model.ID),
		Category:     d.inferCategory(model.ID, model.Type),
		Capabilities: capabilities,
		Provider:     "claude",
		Description:  d.getModelDescription(model.ID),
	}
}

// inferCapabilities infers model capabilities from ID and type.
func (d *Discovery) inferCapabilities(modelID, modelType string) toolkit.ModelCapabilities {
	capabilities := toolkit.ModelCapabilities{}

	modelLower := strings.ToLower(modelID)

	// Claude models primarily support chat
	capabilities.SupportsChat = true

	// Vision capabilities for certain models
	visionKeywords := []string{"claude-3", "vision"}
	capabilities.SupportsVision = d.containsAny(modelLower, visionKeywords)

	// Function calling support (available in Claude 3+)
	capabilities.FunctionCalling = strings.Contains(modelLower, "claude-3")

	// Context window and max tokens
	capabilities.ContextWindow = d.inferContextWindow(modelID)
	capabilities.MaxTokens = d.inferMaxTokens(modelID)

	return capabilities
}

// containsAny checks if the string contains any of the keywords.
func (d *Discovery) containsAny(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

// formatModelName formats model ID into human-readable name.
func (d *Discovery) formatModelName(modelID string) string {
	return strings.ReplaceAll(modelID, "-", " ")
}

// inferCategory infers the model category.
func (d *Discovery) inferCategory(modelID, modelType string) toolkit.ModelCategory {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "vision") || strings.Contains(modelLower, "claude-3") {
		return toolkit.CategoryMultimodal
	}

	return toolkit.CategoryChat
}

// getModelDescription returns a description for the model.
func (d *Discovery) getModelDescription(modelID string) string {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "claude-3-opus") {
		return "Claude 3 Opus - Most capable model for complex tasks"
	}
	if strings.Contains(modelLower, "claude-3-sonnet") {
		return "Claude 3 Sonnet - Balanced model for various tasks"
	}
	if strings.Contains(modelLower, "claude-3-haiku") {
		return "Claude 3 Haiku - Fast and efficient model"
	}
	if strings.Contains(modelLower, "claude-3-5-sonnet") {
		return "Claude 3.5 Sonnet - Latest Sonnet model with improved capabilities"
	}

	return "Anthropic Claude model"
}

// inferContextWindow infers context window size.
func (d *Discovery) inferContextWindow(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "claude-3") {
		return 200000 // Claude 3 models have 200k context
	}

	return 100000 // Default for Claude models
}

// inferMaxTokens infers maximum tokens for output.
func (d *Discovery) inferMaxTokens(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "claude-3") {
		return 4096 // Claude 3 models have 4k max output
	}

	return 4096
}
