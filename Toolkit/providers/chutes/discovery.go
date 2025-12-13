// Package chutes provides model discovery for Chutes API.
package chutes

import (
	"context"
	"strings"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Discovery implements the ModelDiscovery interface for Chutes.
type Discovery struct {
	client *Client
}

// NewDiscovery creates a new Chutes model discovery instance.
func NewDiscovery(apiKey string) *Discovery {
	return &Discovery{
		client: NewClient(apiKey),
	}
}

// Discover discovers available models from Chutes.
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

// convertToModelInfo converts Chutes model info to toolkit ModelInfo.
func (d *Discovery) convertToModelInfo(model ModelInfo) toolkit.ModelInfo {
	capabilities := d.inferCapabilities(model.ID, model.Type)

	return toolkit.ModelInfo{
		ID:           model.ID,
		Name:         d.formatModelName(model.ID),
		Category:     d.inferCategory(model.ID, model.Type),
		Capabilities: capabilities,
		Provider:     "chutes",
		Description:  d.getModelDescription(model.ID),
	}
}

// inferCapabilities infers model capabilities from ID and type.
func (d *Discovery) inferCapabilities(modelID, modelType string) toolkit.ModelCapabilities {
	capabilities := toolkit.ModelCapabilities{}

	modelLower := strings.ToLower(modelID)
	typeLower := strings.ToLower(modelType)

	// Embedding capabilities
	capabilities.SupportsEmbedding = strings.Contains(typeLower, "embedding")

	// Rerank capabilities
	capabilities.SupportsRerank = strings.Contains(typeLower, "rerank")

	// Audio capabilities
	audioKeywords := []string{"tts", "audio", "speech", "voice"}
	capabilities.SupportsAudio = d.containsAny(modelLower, audioKeywords) || d.containsAny(typeLower, audioKeywords)

	// Video capabilities
	videoKeywords := []string{"t2v", "video", "i2v", "flux"}
	capabilities.SupportsVideo = d.containsAny(modelLower, videoKeywords) || d.containsAny(typeLower, videoKeywords)

	// Vision capabilities
	visionKeywords := []string{"vl", "vision", "visual", "multimodal"}
	capabilities.SupportsVision = d.containsAny(modelLower, visionKeywords) || d.containsAny(typeLower, visionKeywords)

	// Chat capabilities - Chutes specific models
	specializedKeywords := []string{
		"embedding", "rerank", "tts", "speech", "audio", "video",
		"t2v", "i2v", "flux", "vl", "vision", "visual", "multimodal", "image",
	}

	isSpecialized := d.containsAny(modelLower, specializedKeywords)
	chatTypeIndicators := []string{"chat", "text", "completion", "instruction", "instruct"}
	hasChatType := d.containsAny(typeLower, chatTypeIndicators)
	chatIDIndicators := []string{
		"instruct", "chat", "qwen", "deepseek", "glm", "kimi",
		"llama", "mistral", "mixtral", "gemma", "phi", "yi",
	}
	hasChatID := d.containsAny(modelLower, chatIDIndicators)

	if hasChatType && !isSpecialized {
		capabilities.SupportsChat = true
	} else if hasChatID && !isSpecialized {
		capabilities.SupportsChat = true
	} else if typeLower == "" || typeLower == "unknown" {
		capabilities.SupportsChat = false
	} else {
		capabilities.SupportsChat = false
	}

	// Function calling support
	capabilities.FunctionCalling = d.supportsFunctionCalling(modelID)

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
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		name := strings.ReplaceAll(parts[1], "-", " ")
		name = strings.ReplaceAll(name, "_", " ")
		return parts[0] + " " + name
	}
	return strings.ReplaceAll(modelID, "-", " ")
}

// inferCategory infers the model category.
func (d *Discovery) inferCategory(modelID, modelType string) toolkit.ModelCategory {
	typeLower := strings.ToLower(modelType)
	modelLower := strings.ToLower(modelID)

	if strings.Contains(typeLower, "embedding") {
		return toolkit.CategoryEmbedding
	}
	if strings.Contains(typeLower, "rerank") {
		return toolkit.CategoryRerank
	}
	if strings.Contains(modelLower, "vl") || strings.Contains(modelLower, "vision") ||
		strings.Contains(modelLower, "multimodal") {
		return toolkit.CategoryMultimodal
	}
	if strings.Contains(modelLower, "image") || strings.Contains(modelLower, "flux") {
		return toolkit.CategoryImage
	}
	if strings.Contains(modelLower, "video") || strings.Contains(modelLower, "t2v") {
		return toolkit.CategoryMultimodal
	}

	return toolkit.CategoryChat
}

// getModelDescription returns a description for the model.
func (d *Discovery) getModelDescription(modelID string) string {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "qwen") {
		return "Qwen series models hosted on Chutes"
	}
	if strings.Contains(modelLower, "deepseek") {
		return "DeepSeek models hosted on Chutes"
	}
	if strings.Contains(modelLower, "glm") {
		return "GLM models hosted on Chutes"
	}
	if strings.Contains(modelLower, "kimi") {
		return "Kimi models hosted on Chutes"
	}

	return "Chutes hosted model"
}

// supportsFunctionCalling checks if model supports function calling.
func (d *Discovery) supportsFunctionCalling(modelID string) bool {
	modelLower := strings.ToLower(modelID)

	supportedModels := []string{"qwen", "deepseek", "glm", "kimi"}
	for _, model := range supportedModels {
		if strings.Contains(modelLower, model) {
			return true
		}
	}

	return false
}

// inferContextWindow infers context window size.
func (d *Discovery) inferContextWindow(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "deepseek") {
		if strings.Contains(modelLower, "r1") {
			return 131072
		}
		return 131072 // V3 series
	}

	if strings.Contains(modelLower, "qwen") {
		if strings.Contains(modelLower, "qwen3") {
			return 32768
		}
		if strings.Contains(modelLower, "qwen2.5") {
			if strings.Contains(modelLower, "72b") {
				return 131072
			}
			return 32768
		}
		if strings.Contains(modelLower, "qwen2-vl") {
			return 32768
		}
	}

	if strings.Contains(modelLower, "glm") {
		if strings.Contains(modelLower, "4.6") {
			return 131072
		}
		return 32768
	}

	if strings.Contains(modelLower, "kimi") {
		return 131072
	}

	return 4096
}

// inferMaxTokens infers maximum tokens for output.
func (d *Discovery) inferMaxTokens(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "deepseek-r1") {
		return 8192
	}
	if strings.Contains(modelLower, "qwen3") {
		return 8192
	}
	if strings.Contains(modelLower, "qwen2.5-72b") {
		return 8192
	}

	return 4096
}
