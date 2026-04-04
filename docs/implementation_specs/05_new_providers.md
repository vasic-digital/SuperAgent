# Implementation Specification: New Provider Adapters

**Document ID:** IMPL-005  
**Feature:** Provider Expansion  
**Priority:** CRITICAL  
**Phase:** 1  
**Estimated Effort:** 4 weeks  
**Source:** Cline, GPTMe, Multiple Agents

---

## Overview

Implement 10+ new LLM provider adapters to expand model coverage and support IDE-integrated models, local models, and cloud APIs.

## Provider Matrix

| Provider | Type | Auth | Models | Effort | Priority |
|----------|------|------|--------|--------|----------|
| VS Code LM API | IDE | OAuth | Copilot | 1 week | Critical |
| LM Studio | Local | None | Local | 3 days | Critical |
| Anthropic Computer Use | Cloud | API Key | claude-3-5-sonnet | 1 week | High |
| Azure OpenAI | Cloud | Service Principal | GPT-4 | 1 week | High |
| Google Vertex AI | Cloud | Service Account | Gemini | 1 week | Medium |
| Together AI | Cloud | API Key | OSS Models | 3 days | Medium |
| Replicate | Cloud | API Key | Various | 3 days | Medium |
| Cohere | Cloud | API Key | Command | 3 days | Low |
| AI21 Labs | Cloud | API Key | Jurassic | 3 days | Low |
| Baseten | Cloud | API Key | Deployed | 3 days | Low |

## Implementation

### 1. VS Code LM API Provider

```go
// internal/llm/providers/vscode/vscode.go

package vscode

import (
    "context"
    "dev.helix.agent/internal/llm"
    "dev.helix.agent/internal/models"
)

// VSCodeProvider uses VS Code's Language Model API
type VSCodeProvider struct {
    client    *lm.LanguageModel
    config    Config
}

type Config struct {
    // VS Code LM API uses OAuth, no explicit API key needed
    // when running inside VS Code extension
    Model     string // "copilot-gpt-4", "copilot-gpt-3.5"
}

func NewVSCodeProvider(config Config) *VSCodeProvider {
    return &VSCodeProvider{config: config}
}

func (p *VSCodeProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    // Use VS Code LanguageModel API
    // This requires running inside VS Code extension context
    
    messages := convertMessages(req.Messages)
    
    response, err := p.client.sendChatRequest(p.config.Model, messages, lm.ChatRequestOptions{
        Justification: req.Metadata["justification"], // Required by VS Code LM API
    })
    
    if err != nil {
        return nil, err
    }
    
    return &models.LLMResponse{
        Content: response.Content,
        ProviderID: "vscode",
        ProviderName: "VS Code LM",
    }, nil
}

func (p *VSCodeProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
    // Stream using VS Code LM API
}
```

### 2. LM Studio Provider

```go
// internal/llm/providers/lmstudio/lmstudio.go

package lmstudio

// LMStudioProvider connects to local LM Studio instance
type LMStudioProvider struct {
    baseURL string
    client  *http.Client
    model   string
}

func NewLMStudioProvider(baseURL, model string) *LMStudioProvider {
    return &LMStudioProvider{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 120 * time.Second},
        model:   model,
    }
}

func (p *LMStudioProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    // LM Studio uses OpenAI-compatible API
    body := map[string]interface{}{
        "model":       p.model,
        "messages":    convertMessages(req.Messages),
        "temperature": req.ModelParams.Temperature,
        "max_tokens":  req.ModelParams.MaxTokens,
    }
    
    resp, err := p.client.Post(
        p.baseURL+"/v1/chat/completions",
        "application/json",
        jsonBody(body),
    )
    
    // Parse OpenAI-compatible response
}
```

### 3. Anthropic Computer Use Provider

```go
// internal/llm/providers/anthropic_cu/anthropic_cu.go

package anthropic_cu

// AnthropicComputerUseProvider enables computer use capabilities
type AnthropicComputerUseProvider struct {
    client *anthropic.Client
    config Config
}

type Config struct {
    APIKey      string
    Model       string // "claude-3-5-sonnet-20241022"
    MaxTokens   int
    Tools       []ComputerTool
}

type ComputerTool struct {
    Type        string `json:"type"` // "computer_20241022"
    DisplayWidth  int  `json:"display_width_px"
    DisplayHeight int  `json:"display_height_px"
    DisplayNumber int  `json:"display_number"
}

func (p *AnthropicComputerUseProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    // Enable computer use tool
    tools := []anthropic.ToolDefinition{
        {
            Type: "computer_20241022",
            DisplayWidth:  1024,
            DisplayHeight: 768,
            DisplayNumber: 1,
        },
        {
            Type: "text_editor_20241022",
        },
        {
            Type: "bash_20241022",
        },
    }
    
    response, err := p.client.CreateMessage(ctx, anthropic.MessageRequest{
        Model:     p.config.Model,
        Messages:  convertMessages(req.Messages),
        MaxTokens: p.config.MaxTokens,
        Tools:     tools,
    })
    
    // Handle tool use blocks for computer interaction
    return p.handleResponse(response)
}

func (p *AnthropicComputerUseProvider) handleResponse(resp *anthropic.Message) (*models.LLMResponse, error) {
    for _, content := range resp.Content {
        switch content.Type {
        case "tool_use":
            // Handle computer use, text editor, bash tools
            // Return as tool calls for execution
        case "text":
            // Regular text response
        }
    }
}
```

### 4. Azure OpenAI Provider

```go
// internal/llm/providers/azure/azure.go

package azure

// AzureOpenAIProvider uses Azure OpenAI Service
type AzureOpenAIProvider struct {
    endpoint   string // https://{resource}.openai.azure.com
    deployment string
    apiKey     string
    apiVersion string
    client     *http.Client
}

func NewAzureOpenAIProvider(endpoint, deployment, apiKey string) *AzureOpenAIProvider {
    return &AzureOpenAIProvider{
        endpoint:   endpoint,
        deployment: deployment,
        apiKey:     apiKey,
        apiVersion: "2024-02-01",
        client:     &http.Client{Timeout: 120 * time.Second},
    }
}

func (p *AzureOpenAIProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
        p.endpoint, p.deployment, p.apiVersion)
    
    body := map[string]interface{}{
        "messages":    convertMessages(req.Messages),
        "temperature": req.ModelParams.Temperature,
        "max_tokens":  req.ModelParams.MaxTokens,
    }
    
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, jsonBody(body))
    httpReq.Header.Set("api-key", p.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := p.client.Do(httpReq)
    // Parse response
}
```

### 5. Together AI Provider

```go
// internal/llm/providers/together/together.go

package together

// TogetherProvider for Together AI API
type TogetherProvider struct {
    apiKey string
    model  string
    client *http.Client
}

func NewTogetherProvider(apiKey, model string) *TogetherProvider {
    return &TogetherProvider{
        apiKey: apiKey,
        model:  model,
        client: &http.Client{Timeout: 120 * time.Second},
    }
}

func (p *TogetherProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    body := map[string]interface{}{
        "model":             p.model,
        "messages":          convertMessages(req.Messages),
        "temperature":       req.ModelParams.Temperature,
        "max_tokens":        req.ModelParams.MaxTokens,
        "top_p":             req.ModelParams.TopP,
        "repetition_penalty": req.ModelParams.RepetitionPenalty,
    }
    
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", 
        "https://api.together.xyz/v1/chat/completions", jsonBody(body))
    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
    
    resp, err := p.client.Do(httpReq)
    // Parse OpenAI-compatible response
}
```

## Provider Registration

```go
// internal/services/provider_registry.go additions

func (r *ProviderRegistry) registerNewProviders() {
    // VS Code LM API (only available in VS Code extension context)
    r.registerProviderFactory("vscode", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return vscode.NewVSCodeProvider(vscode.Config{
            Model: cfg.Model,
        }), nil
    })
    
    // LM Studio
    r.registerProviderFactory("lmstudio", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return lmstudio.NewLMStudioProvider(cfg.BaseURL, cfg.Model), nil
    })
    
    // Anthropic Computer Use
    r.registerProviderFactory("anthropic-cu", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return anthropic_cu.NewAnthropicComputerUseProvider(anthropic_cu.Config{
            APIKey:    cfg.APIKey,
            Model:     cfg.Model,
            MaxTokens: cfg.MaxTokens,
        }), nil
    })
    
    // Azure OpenAI
    r.registerProviderFactory("azure-openai", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return azure.NewAzureOpenAIProvider(cfg.BaseURL, cfg.Deployment, cfg.APIKey), nil
    })
    
    // Together AI
    r.registerProviderFactory("together", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return together.NewTogetherProvider(cfg.APIKey, cfg.Model), nil
    })
    
    // Replicate
    r.registerProviderFactory("replicate", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return replicate.NewReplicateProvider(cfg.APIKey, cfg.Model), nil
    })
    
    // Cohere
    r.registerProviderFactory("cohere", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return cohere.NewCohereProvider(cfg.APIKey, cfg.Model), nil
    })
    
    // AI21 Labs
    r.registerProviderFactory("ai21", func(cfg ProviderConfig) (llm.LLMProvider, error) {
        return ai21.NewAI21Provider(cfg.APIKey, cfg.Model), nil
    })
}
```

## Configuration

```yaml
# configs/providers.yaml
providers:
  # VS Code LM API (VS Code extension only)
  vscode:
    enabled: false  # Auto-detected in VS Code
    model: "copilot-gpt-4"
    
  # LM Studio (local)
  lmstudio:
    enabled: false
    base_url: "http://localhost:1234"
    model: "local-model"
    
  # Anthropic Computer Use
  anthropic_cu:
    enabled: false
    api_key: "${ANTHROPIC_API_KEY}"
    model: "claude-3-5-sonnet-20241022"
    max_tokens: 4096
    
  # Azure OpenAI
  azure_openai:
    enabled: false
    endpoint: "${AZURE_OPENAI_ENDPOINT}"
    deployment: "${AZURE_OPENAI_DEPLOYMENT}"
    api_key: "${AZURE_OPENAI_API_KEY}"
    
  # Together AI
  together:
    enabled: false
    api_key: "${TOGETHER_API_KEY}"
    model: "meta-llama/Llama-2-70b-chat-hf"
    
  # Replicate
  replicate:
    enabled: false
    api_key: "${REPLICATE_API_TOKEN}"
    model: "meta/meta-llama-3-70b-instruct"
    
  # Cohere
  cohere:
    enabled: false
    api_key: "${COHERE_API_KEY}"
    model: "command-r-plus"
    
  # AI21 Labs
  ai21:
    enabled: false
    api_key: "${AI21_API_KEY}"
    model: "jamba-1.5-large"
```

## Implementation Timeline

**Week 1: Critical Providers**
- [ ] VS Code LM API provider
- [ ] LM Studio provider
- [ ] Anthropic Computer Use provider

**Week 2: Cloud Providers**
- [ ] Azure OpenAI provider
- [ ] Together AI provider
- [ ] Replicate provider

**Week 3: Additional Providers**
- [ ] Cohere provider
- [ ] AI21 Labs provider
- [ ] Baseten provider

**Week 4: Integration & Testing**
- [ ] Registration integration
- [ ] Configuration loading
- [ ] Health checks
- [ ] Testing

## Testing

```go
func TestVSCodeProvider_Complete(t *testing.T) {}
func TestLMStudioProvider_Complete(t *testing.T) {}
func TestAnthropicCUProvider_Complete(t *testing.T) {}
func TestAzureOpenAIProvider_Complete(t *testing.T) {}
```
