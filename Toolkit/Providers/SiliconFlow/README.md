# SiliconFlow Go Provider

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/)

A Go implementation of the SiliconFlow AI provider for the SuperAgent Toolkit. Provides comprehensive integration with SiliconFlow's API endpoints including chat completions, embeddings, image generation, and model discovery.

## ✨ Features

- **🔄 Dynamic Model Discovery**: Automatically fetches and categorizes all available SiliconFlow models
- **🎯 Full API Coverage**: Supports all SiliconFlow endpoints (chat, embeddings, rerank, images, audio, video)
- **✅ Comprehensive Testing**: Real API integration tests with extensive model validation
- **📋 Schema Compatibility**: Fully compatible with SuperAgent Toolkit interfaces
- **💾 Intelligent Caching**: Reduces API calls with smart response caching
- **🛡️ Enterprise Security**: Secure API key handling with proper error management
- **📊 Rich Logging**: Structured logging with context for monitoring

## 🚀 Quick Start

### Installation

Add to your Go project:

```bash
go get github.com/superagent/toolkit/SiliconFlow/providers/siliconflow
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/superagent/toolkit/SiliconFlow/providers/siliconflow"
    "github.com/superagent/toolkit/pkg/toolkit"
)

func main() {
    // Create provider configuration
    config := map[string]interface{}{
        "api_key": "your-siliconflow-api-key",
        "base_url": "https://api.siliconflow.com/v1",
        "timeout": 30000,
    }

    // Create provider instance
    provider, err := siliconflow.NewProvider(config)
    if err != nil {
        log.Fatal(err)
    }

    // Discover available models
    ctx := context.Background()
    models, err := provider.DiscoverModels(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Discovered %d models", len(models))

    // Perform chat completion
    req := toolkit.ChatRequest{
        Model: "Qwen/Qwen2.5-7B-Instruct",
        Messages: []toolkit.ChatMessage{
            {Role: "user", Content: "Hello, how are you?"},
        },
        MaxTokens:   1000,
        Temperature: 0.7,
    }

    resp, err := provider.Chat(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Response: %s", resp.Content)
}
```

## 📚 Architecture

### Core Components

```
SiliconFlow/
├── providers/
│   └── siliconflow/
│       ├── builder.go    # Provider factory and registration
│       ├── client.go     # HTTP client and API interactions
│       └── discovery.go  # Model discovery and categorization
└── siliconflow.go        # Main provider implementation
```

### API Client Features

The `SiliconFlowProvider` implements the full SuperAgent Toolkit `Provider` interface:

#### Chat Completions
```go
req := toolkit.ChatRequest{
    Model: "Qwen/Qwen2.5-7B-Instruct",
    Messages: []toolkit.ChatMessage{
        {Role: "user", Content: "Explain quantum computing"},
    },
    MaxTokens:   2000,
    Temperature: 0.7,
}

resp, err := provider.Chat(ctx, req)
```

#### Embeddings
```go
req := toolkit.EmbeddingRequest{
    Model: "text-embedding-ada-002",
    Input: []string{"Hello world", "How are you?"},
}

resp, err := provider.Embed(ctx, req)
```

#### Image Generation
```go
// Note: Image generation support depends on available models
req := toolkit.ImageGenerationRequest{
    Model: "black-forest-labs/FLUX.1-dev",
    Prompt: "A beautiful sunset over mountains",
    Size: "1024x1024",
}

resp, err := provider.CreateImage(ctx, req)
```

### Dynamic Model Discovery

The provider automatically discovers and categorizes models:

```go
models, err := provider.DiscoverModels(ctx)
if err != nil {
    log.Fatal(err)
}

for _, model := range models {
    log.Printf("Model: %s (%s) - %s", model.Name, model.ID, model.Category)
    log.Printf("  Context Window: %d", model.Capabilities.ContextWindow)
    log.Printf("  Supports Vision: %v", model.Capabilities.SupportsVision)
}
```

## 🔧 Configuration

### Provider Configuration

```go
config := map[string]interface{}{
    "api_key":    "your-siliconflow-api-key",     // Required
    "base_url":   "https://api.siliconflow.com/v1", // Optional, defaults to SiliconFlow API
    "timeout":    30000,                          // Optional, milliseconds
    "retries":    3,                              // Optional
    "rate_limit": 60,                             // Optional, requests per minute
}
```

### Integration with SuperAgent Toolkit

Register the provider with the toolkit:

```go
tk := toolkit.NewToolkit()

// Register SiliconFlow provider factory
if err := siliconflow.Register(tk.GetProviderFactoryRegistry()); err != nil {
    log.Fatal(err)
}

// Create provider instance
provider, err := tk.CreateProvider("siliconflow", config)
if err != nil {
    log.Fatal(err)
}
```

## 🧪 Testing

### Running Tests

```bash
# Run unit tests
go test ./providers/siliconflow/...

# Run with verbose output
go test -v ./providers/siliconflow/...

# Run integration tests (requires API key)
SILICONFLOW_API_KEY=your-key go test -tags=integration ./providers/siliconflow/...
```

### Test Coverage

- ✅ **API Client Tests**: All endpoints tested with mocked responses
- ✅ **Model Discovery Tests**: Categorization and capability inference
- ✅ **Configuration Tests**: Provider configuration validation
- ✅ **Integration Tests**: Real API calls (when API key provided)
- ✅ **Error Handling Tests**: Comprehensive error scenarios

## 📊 Model Categories

| Category | Count | Default Model | Capabilities |
|----------|-------|---------------|--------------|
| Chat | 77+ | Qwen/Qwen2.5-14B-Instruct | Text generation, reasoning, function calling |
| Vision | 12+ | Qwen/Qwen2.5-VL-7B-Instruct | Image understanding, visual chat |
| Audio | 3+ | FunAudioLLM/CosyVoice2-0.5B | Speech synthesis, voice cloning |
| Video | 2+ | Wan-AI/Wan2.2-T2V-A14B | Video generation from text |
| Embedding | 0 | - | Text embeddings (coming soon) |
| Rerank | 0 | - | Document reranking (coming soon) |

## 🔒 Security

- **API Key Protection**: Keys handled securely, never logged
- **HTTPS Only**: All API calls use HTTPS
- **Input Validation**: Comprehensive validation of all inputs
- **Error Sanitization**: Sensitive information removed from errors

## 📈 Performance

- **Smart Caching**: Model discovery results cached to reduce API calls
- **Connection Pooling**: HTTP client reuses connections
- **Concurrent Safety**: Thread-safe implementation
- **Memory Optimized**: Efficient JSON processing and streaming support

## 🐛 Troubleshooting

### Common Issues

1. **API Key Not Found**
    ```bash
    export SILICONFLOW_API_KEY=your-key
    ```

2. **Timeout Errors**
    - Increase timeout in configuration
    - Check network connectivity

3. **Rate Limiting**
    - Implement retry logic with backoff
    - Check rate limit configuration

4. **Model Not Found**
    - Verify model ID is correct
    - Run model discovery to see available models

### Debug Mode

Enable detailed logging:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
// Provider will log detailed information
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Run linting: `golangci-lint run`
6. Submit a pull request

### Development Setup

```bash
# Clone the repository
git clone <repository-url>
cd SiliconFlow

# Initialize Go module (if not already done)
go mod init github.com/superagent/toolkit/SiliconFlow

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Format code
gofmt -w .

# Run linter
golangci-lint run
```

### File Structure

```
SiliconFlow/
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── providers/
│   └── siliconflow/
│       ├── builder.go     # Provider factory and registration
│       ├── client.go      # HTTP client implementation
│       ├── discovery.go   # Model discovery logic
│       └── siliconflow.go # Main provider implementation
├── .gitignore             # Go-specific ignore patterns
├── README.md              # This documentation
├── AGENTS.md              # Agent development guidelines
├── LICENSE                # MIT license
└── API_REFERENCE.md       # Technical API documentation
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- **SiliconFlow**: For providing comprehensive AI infrastructure
- **SuperAgent Team**: For the toolkit framework
- **Go Community**: For excellent HTTP and JSON libraries
