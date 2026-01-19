// Package servers provides MCP server adapters for various services.
package servers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// StableDiffusionConfig contains configuration for Stable Diffusion WebUI adapter.
type StableDiffusionConfig struct {
	BaseURL  string        `json:"base_url"`
	Username string        `json:"username,omitempty"`
	Password string        `json:"password,omitempty"`
	Timeout  time.Duration `json:"timeout"`
}

// SDTxt2ImgRequest represents a text-to-image request.
type SDTxt2ImgRequest struct {
	Prompt            string   `json:"prompt"`
	NegativePrompt    string   `json:"negative_prompt,omitempty"`
	Styles            []string `json:"styles,omitempty"`
	Seed              int64    `json:"seed,omitempty"`
	Subseed           int64    `json:"subseed,omitempty"`
	SubseedStrength   float64  `json:"subseed_strength,omitempty"`
	SeedResizeFromH   int      `json:"seed_resize_from_h,omitempty"`
	SeedResizeFromW   int      `json:"seed_resize_from_w,omitempty"`
	SamplerName       string   `json:"sampler_name,omitempty"`
	BatchSize         int      `json:"batch_size,omitempty"`
	NIter             int      `json:"n_iter,omitempty"`
	Steps             int      `json:"steps,omitempty"`
	CFGScale          float64  `json:"cfg_scale,omitempty"`
	Width             int      `json:"width,omitempty"`
	Height            int      `json:"height,omitempty"`
	RestoreFaces      bool     `json:"restore_faces,omitempty"`
	Tiling            bool     `json:"tiling,omitempty"`
	DoNotSaveSamples  bool     `json:"do_not_save_samples,omitempty"`
	DoNotSaveGrid     bool     `json:"do_not_save_grid,omitempty"`
	DenoisingStrength float64  `json:"denoising_strength,omitempty"`
	EnableHR          bool     `json:"enable_hr,omitempty"`
	HRScale           float64  `json:"hr_scale,omitempty"`
	HRUpscaler        string   `json:"hr_upscaler,omitempty"`
	HRSecondPassSteps int      `json:"hr_second_pass_steps,omitempty"`
	HRResizeX         int      `json:"hr_resize_x,omitempty"`
	HRResizeY         int      `json:"hr_resize_y,omitempty"`
	OverrideSettings  map[string]interface{} `json:"override_settings,omitempty"`
	AlwaysonScripts   map[string]interface{} `json:"alwayson_scripts,omitempty"`
}

// SDImg2ImgRequest represents an image-to-image request.
type SDImg2ImgRequest struct {
	InitImages        []string `json:"init_images"`
	Prompt            string   `json:"prompt"`
	NegativePrompt    string   `json:"negative_prompt,omitempty"`
	Styles            []string `json:"styles,omitempty"`
	Seed              int64    `json:"seed,omitempty"`
	SamplerName       string   `json:"sampler_name,omitempty"`
	BatchSize         int      `json:"batch_size,omitempty"`
	NIter             int      `json:"n_iter,omitempty"`
	Steps             int      `json:"steps,omitempty"`
	CFGScale          float64  `json:"cfg_scale,omitempty"`
	Width             int      `json:"width,omitempty"`
	Height            int      `json:"height,omitempty"`
	RestoreFaces      bool     `json:"restore_faces,omitempty"`
	Tiling            bool     `json:"tiling,omitempty"`
	DenoisingStrength float64  `json:"denoising_strength,omitempty"`
	Mask              string   `json:"mask,omitempty"`
	MaskBlur          int      `json:"mask_blur,omitempty"`
	InpaintingFill    int      `json:"inpainting_fill,omitempty"`
	InpaintFullRes    bool     `json:"inpaint_full_res,omitempty"`
	InpaintFullResPadding int  `json:"inpaint_full_res_padding,omitempty"`
	InpaintingMaskInvert int   `json:"inpainting_mask_invert,omitempty"`
	ResizeMode        int      `json:"resize_mode,omitempty"`
	OverrideSettings  map[string]interface{} `json:"override_settings,omitempty"`
	AlwaysonScripts   map[string]interface{} `json:"alwayson_scripts,omitempty"`
}

// SDGenerationResponse represents an image generation response.
type SDGenerationResponse struct {
	Images     []string               `json:"images"`
	Parameters map[string]interface{} `json:"parameters"`
	Info       string                 `json:"info"`
}

// SDUpscaleRequest represents an upscaling request.
type SDUpscaleRequest struct {
	ResizeMode         int     `json:"resize_mode"`
	ShowExtrasResults  bool    `json:"show_extras_results"`
	GFPGANVisibility   float64 `json:"gfpgan_visibility"`
	CodeformerVisibility float64 `json:"codeformer_visibility"`
	CodeformerWeight   float64 `json:"codeformer_weight"`
	UpscalingResize    float64 `json:"upscaling_resize"`
	UpscalingResizeW   int     `json:"upscaling_resize_w"`
	UpscalingResizeH   int     `json:"upscaling_resize_h"`
	UpscalingCrop      bool    `json:"upscaling_crop"`
	Upscaler1          string  `json:"upscaler_1"`
	Upscaler2          string  `json:"upscaler_2"`
	Upscaler2Visibility float64 `json:"upscaler_2_visibility"`
	Image              string  `json:"image"`
}

// SDUpscaleResponse represents an upscaling response.
type SDUpscaleResponse struct {
	HTMLInfo string `json:"html_info"`
	Image    string `json:"image"`
}

// SDModel represents a loaded model.
type SDModel struct {
	Title     string `json:"title"`
	ModelName string `json:"model_name"`
	Hash      string `json:"hash,omitempty"`
	SHA256    string `json:"sha256,omitempty"`
	Filename  string `json:"filename"`
	Config    string `json:"config,omitempty"`
}

// SDSampler represents a sampler.
type SDSampler struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Options map[string]interface{} `json:"options"`
}

// SDUpscaler represents an upscaler.
type SDUpscaler struct {
	Name       string  `json:"name"`
	ModelName  string  `json:"model_name,omitempty"`
	ModelPath  string  `json:"model_path,omitempty"`
	ModelURL   string  `json:"model_url,omitempty"`
	Scale      float64 `json:"scale,omitempty"`
}

// SDLora represents a LoRA model.
type SDLora struct {
	Name    string            `json:"name"`
	Alias   string            `json:"alias,omitempty"`
	Path    string            `json:"path"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SDProgress represents generation progress.
type SDProgress struct {
	Progress    float64                `json:"progress"`
	ETARelative float64                `json:"eta_relative"`
	State       map[string]interface{} `json:"state"`
	CurrentImage string                `json:"current_image,omitempty"`
	TextInfo    string                 `json:"textinfo,omitempty"`
}

// SDOptions represents WebUI options.
type SDOptions struct {
	SDModelCheckpoint string `json:"sd_model_checkpoint,omitempty"`
	SDVae             string `json:"sd_vae,omitempty"`
	CLIPStopAtLastLayers int `json:"CLIP_stop_at_last_layers,omitempty"`
	// Add more options as needed
}

// StableDiffusionAdapter implements ServerAdapter for Stable Diffusion WebUI API.
type StableDiffusionAdapter struct {
	mu        sync.RWMutex
	config    StableDiffusionConfig
	client    *http.Client
	connected bool
	baseURL   string
}

// NewStableDiffusionAdapter creates a new Stable Diffusion WebUI adapter.
func NewStableDiffusionAdapter(config StableDiffusionConfig) *StableDiffusionAdapter {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://127.0.0.1:7860"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 300 * time.Second // SD generation can take a while
	}

	return &StableDiffusionAdapter{
		config:  config,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Connect establishes connection to Stable Diffusion WebUI.
func (a *StableDiffusionAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Try to get samplers to verify connection
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/samplers", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Stable Diffusion WebUI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to connect: %s", string(body))
	}

	a.connected = true
	return nil
}

// Close closes the adapter connection.
func (a *StableDiffusionAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// Health checks if the adapter is healthy.
func (a *StableDiffusionAdapter) Health(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/samplers", nil)
	if err != nil {
		return err
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// Txt2Img generates images from text prompt.
func (a *StableDiffusionAdapter) Txt2Img(ctx context.Context, req *SDTxt2ImgRequest) (*SDGenerationResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Set defaults
	if req.Steps == 0 {
		req.Steps = 20
	}
	if req.CFGScale == 0 {
		req.CFGScale = 7.0
	}
	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}
	if req.BatchSize == 0 {
		req.BatchSize = 1
	}
	if req.NIter == 0 {
		req.NIter = 1
	}
	if req.SamplerName == "" {
		req.SamplerName = "Euler"
	}
	if req.Seed == 0 {
		req.Seed = -1
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/txt2img", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if a.config.Username != "" {
		httpReq.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generation failed: %s", string(respBody))
	}

	var result SDGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Img2Img generates images from an existing image.
func (a *StableDiffusionAdapter) Img2Img(ctx context.Context, req *SDImg2ImgRequest) (*SDGenerationResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Set defaults
	if req.Steps == 0 {
		req.Steps = 20
	}
	if req.CFGScale == 0 {
		req.CFGScale = 7.0
	}
	if req.DenoisingStrength == 0 {
		req.DenoisingStrength = 0.75
	}
	if req.BatchSize == 0 {
		req.BatchSize = 1
	}
	if req.NIter == 0 {
		req.NIter = 1
	}
	if req.SamplerName == "" {
		req.SamplerName = "Euler"
	}
	if req.Seed == 0 {
		req.Seed = -1
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/img2img", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if a.config.Username != "" {
		httpReq.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generation failed: %s", string(respBody))
	}

	var result SDGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Upscale upscales an image.
func (a *StableDiffusionAdapter) Upscale(ctx context.Context, req *SDUpscaleRequest) (*SDUpscaleResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Set defaults
	if req.UpscalingResize == 0 {
		req.UpscalingResize = 2.0
	}
	if req.Upscaler1 == "" {
		req.Upscaler1 = "R-ESRGAN 4x+"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/extra-single-image", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if a.config.Username != "" {
		httpReq.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to upscale image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upscale failed: %s", string(respBody))
	}

	var result SDUpscaleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetProgress returns generation progress.
func (a *StableDiffusionAdapter) GetProgress(ctx context.Context, skipCurrentImage bool) (*SDProgress, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/sdapi/v1/progress?skip_current_image=%t", a.baseURL, skipCurrentImage)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get progress: %s", string(body))
	}

	var progress SDProgress
	if err := json.NewDecoder(resp.Body).Decode(&progress); err != nil {
		return nil, fmt.Errorf("failed to decode progress: %w", err)
	}

	return &progress, nil
}

// Interrupt interrupts current generation.
func (a *StableDiffusionAdapter) Interrupt(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/interrupt", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to interrupt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to interrupt: %s", string(body))
	}

	return nil
}

// Skip skips current image in batch.
func (a *StableDiffusionAdapter) Skip(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/skip", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to skip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to skip: %s", string(body))
	}

	return nil
}

// GetModels returns available models.
func (a *StableDiffusionAdapter) GetModels(ctx context.Context) ([]SDModel, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/sd-models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get models: %s", string(body))
	}

	var models []SDModel
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("failed to decode models: %w", err)
	}

	return models, nil
}

// GetSamplers returns available samplers.
func (a *StableDiffusionAdapter) GetSamplers(ctx context.Context) ([]SDSampler, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/samplers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get samplers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get samplers: %s", string(body))
	}

	var samplers []SDSampler
	if err := json.NewDecoder(resp.Body).Decode(&samplers); err != nil {
		return nil, fmt.Errorf("failed to decode samplers: %w", err)
	}

	return samplers, nil
}

// GetUpscalers returns available upscalers.
func (a *StableDiffusionAdapter) GetUpscalers(ctx context.Context) ([]SDUpscaler, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/upscalers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get upscalers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get upscalers: %s", string(body))
	}

	var upscalers []SDUpscaler
	if err := json.NewDecoder(resp.Body).Decode(&upscalers); err != nil {
		return nil, fmt.Errorf("failed to decode upscalers: %w", err)
	}

	return upscalers, nil
}

// GetLoras returns available LoRA models.
func (a *StableDiffusionAdapter) GetLoras(ctx context.Context) ([]SDLora, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/loras", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get loras: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get loras: %s", string(body))
	}

	var loras []SDLora
	if err := json.NewDecoder(resp.Body).Decode(&loras); err != nil {
		return nil, fmt.Errorf("failed to decode loras: %w", err)
	}

	return loras, nil
}

// GetOptions returns current options.
func (a *StableDiffusionAdapter) GetOptions(ctx context.Context) (*SDOptions, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/sdapi/v1/options", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get options: %s", string(body))
	}

	var options SDOptions
	if err := json.NewDecoder(resp.Body).Decode(&options); err != nil {
		return nil, fmt.Errorf("failed to decode options: %w", err)
	}

	return &options, nil
}

// SetOptions sets WebUI options.
func (a *StableDiffusionAdapter) SetOptions(ctx context.Context, options map[string]interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	body, err := json.Marshal(options)
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/options", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set options: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set options: %s", string(respBody))
	}

	return nil
}

// RefreshModels refreshes the list of models.
func (a *StableDiffusionAdapter) RefreshModels(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/refresh-checkpoints", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to refresh models: %s", string(body))
	}

	return nil
}

// PNGInfo extracts metadata from a PNG image.
func (a *StableDiffusionAdapter) PNGInfo(ctx context.Context, imageBase64 string) (map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]string{
		"image": imageBase64,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/sdapi/v1/png-info", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.config.Username != "" {
		req.SetBasicAuth(a.config.Username, a.config.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get PNG info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get PNG info: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// DecodeImages decodes base64 images from a generation response.
func (a *StableDiffusionAdapter) DecodeImages(response *SDGenerationResponse) ([][]byte, error) {
	images := make([][]byte, len(response.Images))
	for i, img := range response.Images {
		decoded, err := base64.StdEncoding.DecodeString(img)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image %d: %w", i, err)
		}
		images[i] = decoded
	}
	return images, nil
}

// GetMCPTools returns the MCP tool definitions for Stable Diffusion.
func (a *StableDiffusionAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "sd_txt2img",
			Description: "Generate images from text prompt using Stable Diffusion",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "Text prompt describing the image to generate",
					},
					"negative_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Negative prompt (what to avoid)",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Image width in pixels (default: 512)",
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Image height in pixels (default: 512)",
					},
					"steps": map[string]interface{}{
						"type":        "integer",
						"description": "Number of sampling steps (default: 20)",
					},
					"cfg_scale": map[string]interface{}{
						"type":        "number",
						"description": "CFG scale/guidance scale (default: 7.0)",
					},
					"sampler_name": map[string]interface{}{
						"type":        "string",
						"description": "Sampler to use (default: Euler)",
					},
					"seed": map[string]interface{}{
						"type":        "integer",
						"description": "Random seed (-1 for random)",
					},
					"batch_size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of images per batch",
					},
				},
				"required": []string{"prompt"},
			},
		},
		{
			Name:        "sd_img2img",
			Description: "Generate images from an existing image using Stable Diffusion",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Base64 encoded input image",
					},
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "Text prompt describing the desired transformation",
					},
					"negative_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Negative prompt (what to avoid)",
					},
					"denoising_strength": map[string]interface{}{
						"type":        "number",
						"description": "Denoising strength (0-1, default: 0.75)",
					},
					"steps": map[string]interface{}{
						"type":        "integer",
						"description": "Number of sampling steps",
					},
					"cfg_scale": map[string]interface{}{
						"type":        "number",
						"description": "CFG scale/guidance scale",
					},
				},
				"required": []string{"image", "prompt"},
			},
		},
		{
			Name:        "sd_upscale",
			Description: "Upscale an image using Stable Diffusion extras",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Base64 encoded input image",
					},
					"scale": map[string]interface{}{
						"type":        "number",
						"description": "Upscaling factor (default: 2.0)",
					},
					"upscaler": map[string]interface{}{
						"type":        "string",
						"description": "Upscaler to use",
					},
				},
				"required": []string{"image"},
			},
		},
		{
			Name:        "sd_get_models",
			Description: "Get available Stable Diffusion models",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_get_samplers",
			Description: "Get available samplers",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_get_upscalers",
			Description: "Get available upscalers",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_get_loras",
			Description: "Get available LoRA models",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sd_get_progress",
			Description: "Get current generation progress",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"skip_current_image": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to skip the current preview image",
					},
				},
			},
		},
		{
			Name:        "sd_interrupt",
			Description: "Interrupt current generation",
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
			Name:        "sd_png_info",
			Description: "Extract generation metadata from a PNG image",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Base64 encoded PNG image",
					},
				},
				"required": []string{"image"},
			},
		},
	}
}
