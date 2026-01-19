// Package adapters provides MCP server adapters.
// This file implements the Stable Diffusion WebUI MCP server adapter for image generation.
package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StableDiffusionConfig configures the Stable Diffusion adapter.
type StableDiffusionConfig struct {
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
	Auth    *SDAuth       `json:"auth,omitempty"`
}

// SDAuth represents authentication for Stable Diffusion WebUI.
type SDAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DefaultStableDiffusionConfig returns default configuration.
func DefaultStableDiffusionConfig() StableDiffusionConfig {
	return StableDiffusionConfig{
		BaseURL: "http://127.0.0.1:7860",
		Timeout: 300 * time.Second, // Image generation can take time
	}
}

// StableDiffusionAdapter implements the Stable Diffusion WebUI MCP server.
type StableDiffusionAdapter struct {
	config     StableDiffusionConfig
	httpClient *http.Client
}

// NewStableDiffusionAdapter creates a new Stable Diffusion adapter.
func NewStableDiffusionAdapter(config StableDiffusionConfig) *StableDiffusionAdapter {
	if config.BaseURL == "" {
		config.BaseURL = DefaultStableDiffusionConfig().BaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultStableDiffusionConfig().Timeout
	}
	return &StableDiffusionAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetServerInfo returns server information.
func (a *StableDiffusionAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "stable-diffusion",
		Version:     "1.0.0",
		Description: "Stable Diffusion WebUI integration for AI image generation",
		Capabilities: []string{
			"txt2img",
			"img2img",
			"upscale",
			"controlnet",
			"models",
			"samplers",
			"loras",
		},
	}
}

// ListTools returns available tools.
func (a *StableDiffusionAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "sd_txt2img",
			Description: "Generate an image from a text prompt",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "The text prompt describing the desired image",
					},
					"negative_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Negative prompt to avoid unwanted elements",
						"default":     "",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Image width in pixels",
						"default":     512,
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Image height in pixels",
						"default":     512,
					},
					"steps": map[string]interface{}{
						"type":        "integer",
						"description": "Number of sampling steps",
						"default":     20,
					},
					"cfg_scale": map[string]interface{}{
						"type":        "number",
						"description": "Classifier-free guidance scale",
						"default":     7.0,
					},
					"sampler_name": map[string]interface{}{
						"type":        "string",
						"description": "Sampler to use (e.g., Euler, DPM++ 2M)",
						"default":     "Euler",
					},
					"seed": map[string]interface{}{
						"type":        "integer",
						"description": "Random seed (-1 for random)",
						"default":     -1,
					},
					"batch_size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of images to generate",
						"default":     1,
					},
				},
				"required": []string{"prompt"},
			},
		},
		{
			Name:        "sd_img2img",
			Description: "Generate an image based on an input image and prompt",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "The text prompt",
					},
					"init_image": map[string]interface{}{
						"type":        "string",
						"description": "Base64-encoded input image",
					},
					"denoising_strength": map[string]interface{}{
						"type":        "number",
						"description": "How much to change the image (0-1)",
						"default":     0.75,
					},
					"negative_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Negative prompt",
						"default":     "",
					},
					"steps": map[string]interface{}{
						"type":        "integer",
						"description": "Number of sampling steps",
						"default":     20,
					},
					"cfg_scale": map[string]interface{}{
						"type":        "number",
						"description": "Classifier-free guidance scale",
						"default":     7.0,
					},
					"seed": map[string]interface{}{
						"type":        "integer",
						"description": "Random seed",
						"default":     -1,
					},
				},
				"required": []string{"prompt", "init_image"},
			},
		},
		{
			Name:        "sd_controlnet",
			Description: "Generate an image using ControlNet for precise control",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "The text prompt",
					},
					"control_image": map[string]interface{}{
						"type":        "string",
						"description": "Base64-encoded control image (e.g., edge map, pose)",
					},
					"controlnet_model": map[string]interface{}{
						"type":        "string",
						"description": "ControlNet model to use (e.g., canny, openpose, depth)",
					},
					"controlnet_weight": map[string]interface{}{
						"type":        "number",
						"description": "Control strength (0-2)",
						"default":     1.0,
					},
					"negative_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Negative prompt",
						"default":     "",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Image width",
						"default":     512,
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Image height",
						"default":     512,
					},
					"steps": map[string]interface{}{
						"type":        "integer",
						"description": "Number of steps",
						"default":     20,
					},
				},
				"required": []string{"prompt", "control_image", "controlnet_model"},
			},
		},
		{
			Name:        "sd_upscale",
			Description: "Upscale an image using AI upscaling",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Base64-encoded image to upscale",
					},
					"upscaler": map[string]interface{}{
						"type":        "string",
						"description": "Upscaler model (e.g., ESRGAN_4x, R-ESRGAN 4x+)",
						"default":     "ESRGAN_4x",
					},
					"scale": map[string]interface{}{
						"type":        "number",
						"description": "Upscale factor",
						"default":     2.0,
					},
				},
				"required": []string{"image"},
			},
		},
		{
			Name:        "sd_list_models",
			Description: "List available Stable Diffusion models",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_list_samplers",
			Description: "List available samplers",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_list_loras",
			Description: "List available LoRA models",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_set_model",
			Description: "Set the active Stable Diffusion model",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the model to load",
					},
				},
				"required": []string{"model_name"},
			},
		},
		{
			Name:        "sd_get_progress",
			Description: "Get current image generation progress",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// CallTool executes a tool.
func (a *StableDiffusionAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "sd_txt2img":
		return a.txt2img(ctx, args)
	case "sd_img2img":
		return a.img2img(ctx, args)
	case "sd_controlnet":
		return a.controlnet(ctx, args)
	case "sd_upscale":
		return a.upscale(ctx, args)
	case "sd_list_models":
		return a.listModels(ctx, args)
	case "sd_list_samplers":
		return a.listSamplers(ctx, args)
	case "sd_list_loras":
		return a.listLoras(ctx, args)
	case "sd_set_model":
		return a.setModel(ctx, args)
	case "sd_get_progress":
		return a.getProgress(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *StableDiffusionAdapter) txt2img(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	prompt, _ := args["prompt"].(string)
	negativePrompt, _ := args["negative_prompt"].(string)
	width := getIntArg(args, "width", 512)
	height := getIntArg(args, "height", 512)
	steps := getIntArg(args, "steps", 20)
	cfgScale := getFloatArg(args, "cfg_scale", 7.0)
	samplerName, _ := args["sampler_name"].(string)
	if samplerName == "" {
		samplerName = "Euler"
	}
	seed := getIntArg(args, "seed", -1)
	batchSize := getIntArg(args, "batch_size", 1)

	payload := map[string]interface{}{
		"prompt":          prompt,
		"negative_prompt": negativePrompt,
		"width":           width,
		"height":          height,
		"steps":           steps,
		"cfg_scale":       cfgScale,
		"sampler_name":    samplerName,
		"seed":            seed,
		"batch_size":      batchSize,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/sdapi/v1/txt2img", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SDImageResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var content []ContentBlock
	content = append(content, ContentBlock{
		Type: "text",
		Text: fmt.Sprintf("Generated %d image(s) with prompt: %s\nSeed: %d", len(result.Images), prompt, result.Info.Seed),
	})

	for i, img := range result.Images {
		content = append(content, ContentBlock{
			Type:     "image",
			MimeType: "image/png",
			Data:     img,
		})
		content = append(content, ContentBlock{
			Type: "text",
			Text: fmt.Sprintf("Image %d (base64 length: %d)", i+1, len(img)),
		})
	}

	return &ToolResult{Content: content}, nil
}

func (a *StableDiffusionAdapter) img2img(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	prompt, _ := args["prompt"].(string)
	initImage, _ := args["init_image"].(string)
	denoisingStrength := getFloatArg(args, "denoising_strength", 0.75)
	negativePrompt, _ := args["negative_prompt"].(string)
	steps := getIntArg(args, "steps", 20)
	cfgScale := getFloatArg(args, "cfg_scale", 7.0)
	seed := getIntArg(args, "seed", -1)

	payload := map[string]interface{}{
		"prompt":             prompt,
		"negative_prompt":    negativePrompt,
		"init_images":        []string{initImage},
		"denoising_strength": denoisingStrength,
		"steps":              steps,
		"cfg_scale":          cfgScale,
		"seed":               seed,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/sdapi/v1/img2img", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SDImageResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var content []ContentBlock
	content = append(content, ContentBlock{
		Type: "text",
		Text: fmt.Sprintf("Generated img2img with denoising: %.2f\nSeed: %d", denoisingStrength, result.Info.Seed),
	})

	for i, img := range result.Images {
		content = append(content, ContentBlock{
			Type:     "image",
			MimeType: "image/png",
			Data:     img,
		})
		content = append(content, ContentBlock{
			Type: "text",
			Text: fmt.Sprintf("Image %d", i+1),
		})
	}

	return &ToolResult{Content: content}, nil
}

func (a *StableDiffusionAdapter) controlnet(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	prompt, _ := args["prompt"].(string)
	controlImage, _ := args["control_image"].(string)
	controlnetModel, _ := args["controlnet_model"].(string)
	controlnetWeight := getFloatArg(args, "controlnet_weight", 1.0)
	negativePrompt, _ := args["negative_prompt"].(string)
	width := getIntArg(args, "width", 512)
	height := getIntArg(args, "height", 512)
	steps := getIntArg(args, "steps", 20)

	payload := map[string]interface{}{
		"prompt":          prompt,
		"negative_prompt": negativePrompt,
		"width":           width,
		"height":          height,
		"steps":           steps,
		"alwayson_scripts": map[string]interface{}{
			"controlnet": map[string]interface{}{
				"args": []map[string]interface{}{
					{
						"enabled":      true,
						"module":       controlnetModel,
						"model":        controlnetModel,
						"weight":       controlnetWeight,
						"input_image":  controlImage,
						"resize_mode":  1,
						"control_mode": 0,
					},
				},
			},
		},
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/sdapi/v1/txt2img", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SDImageResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var content []ContentBlock
	content = append(content, ContentBlock{
		Type: "text",
		Text: fmt.Sprintf("Generated with ControlNet (%s, weight: %.2f)", controlnetModel, controlnetWeight),
	})

	for _, img := range result.Images {
		content = append(content, ContentBlock{
			Type:     "image",
			MimeType: "image/png",
			Data:     img,
		})
	}

	return &ToolResult{Content: content}, nil
}

func (a *StableDiffusionAdapter) upscale(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	image, _ := args["image"].(string)
	upscaler, _ := args["upscaler"].(string)
	if upscaler == "" {
		upscaler = "ESRGAN_4x"
	}
	scale := getFloatArg(args, "scale", 2.0)

	payload := map[string]interface{}{
		"image":              image,
		"upscaler_1":         upscaler,
		"upscaling_resize":   scale,
	}

	resp, err := a.makeRequest(ctx, http.MethodPost, "/sdapi/v1/extra-single-image", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var result SDUpscaleResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Upscaled image with %s (%.1fx)", upscaler, scale)},
			{Type: "image", MimeType: "image/png", Data: result.Image},
		},
	}, nil
}

func (a *StableDiffusionAdapter) listModels(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resp, err := a.makeRequest(ctx, http.MethodGet, "/sdapi/v1/sd-models", nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var models []SDModel
	if err := json.Unmarshal(resp, &models); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available models (%d):\n\n", len(models)))

	for _, model := range models {
		sb.WriteString(fmt.Sprintf("- %s\n", model.Title))
		sb.WriteString(fmt.Sprintf("  Filename: %s\n", model.ModelName))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *StableDiffusionAdapter) listSamplers(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resp, err := a.makeRequest(ctx, http.MethodGet, "/sdapi/v1/samplers", nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var samplers []SDSampler
	if err := json.Unmarshal(resp, &samplers); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available samplers (%d):\n\n", len(samplers)))

	for _, sampler := range samplers {
		sb.WriteString(fmt.Sprintf("- %s\n", sampler.Name))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *StableDiffusionAdapter) listLoras(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resp, err := a.makeRequest(ctx, http.MethodGet, "/sdapi/v1/loras", nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var loras []SDLora
	if err := json.Unmarshal(resp, &loras); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available LoRAs (%d):\n\n", len(loras)))

	for _, lora := range loras {
		sb.WriteString(fmt.Sprintf("- %s\n", lora.Name))
		if lora.Alias != "" {
			sb.WriteString(fmt.Sprintf("  Alias: %s\n", lora.Alias))
		}
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *StableDiffusionAdapter) setModel(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	modelName, _ := args["model_name"].(string)

	payload := map[string]interface{}{
		"sd_model_checkpoint": modelName,
	}

	_, err := a.makeRequest(ctx, http.MethodPost, "/sdapi/v1/options", payload)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Model set to: %s", modelName)}},
	}, nil
}

func (a *StableDiffusionAdapter) getProgress(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resp, err := a.makeRequest(ctx, http.MethodGet, "/sdapi/v1/progress", nil)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var progress SDProgress
	if err := json.Unmarshal(resp, &progress); err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Progress: %.1f%%\n", progress.Progress*100))
	sb.WriteString(fmt.Sprintf("ETA: %.1f seconds\n", progress.ETARelative))
	if progress.State != nil {
		sb.WriteString(fmt.Sprintf("Step: %d/%d\n", progress.State.SamplingStep, progress.State.SamplingSteps))
		sb.WriteString(fmt.Sprintf("Job: %s\n", progress.State.Job))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *StableDiffusionAdapter) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) ([]byte, error) {
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

	if a.config.Auth != nil {
		auth := base64.StdEncoding.EncodeToString([]byte(a.config.Auth.Username + ":" + a.config.Auth.Password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return body, nil
}

// Stable Diffusion API response types

// SDImageResponse represents a txt2img/img2img response.
type SDImageResponse struct {
	Images []string `json:"images"`
	Info   SDInfo   `json:"info"`
}

// SDInfo represents generation info.
type SDInfo struct {
	Seed       int64   `json:"seed"`
	Prompt     string  `json:"prompt"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Steps      int     `json:"steps"`
	CFGScale   float64 `json:"cfg_scale"`
	SamplerName string `json:"sampler_name"`
}

// SDUpscaleResponse represents an upscale response.
type SDUpscaleResponse struct {
	Image string `json:"image"`
}

// SDModel represents a Stable Diffusion model.
type SDModel struct {
	Title     string `json:"title"`
	ModelName string `json:"model_name"`
	Hash      string `json:"hash"`
	Filename  string `json:"filename"`
}

// SDSampler represents a sampler.
type SDSampler struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

// SDLora represents a LoRA model.
type SDLora struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
	Path  string `json:"path"`
}

// SDProgress represents generation progress.
type SDProgress struct {
	Progress    float64  `json:"progress"`
	ETARelative float64  `json:"eta_relative"`
	State       *SDState `json:"state"`
}

// SDState represents the current state.
type SDState struct {
	Skipped       bool   `json:"skipped"`
	Interrupted   bool   `json:"interrupted"`
	Job           string `json:"job"`
	JobCount      int    `json:"job_count"`
	JobNo         int    `json:"job_no"`
	SamplingStep  int    `json:"sampling_step"`
	SamplingSteps int    `json:"sampling_steps"`
}
