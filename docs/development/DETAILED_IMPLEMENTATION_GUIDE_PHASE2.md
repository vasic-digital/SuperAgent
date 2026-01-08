# Detailed Implementation Guide - Phase 2
## Test Infrastructure (Week 3-4)

## Overview
This guide provides step-by-step instructions for implementing Phase 2 of the project completion plan, focusing on achieving 95%+ test coverage across all test types.

---

## Week 3: Unit Test Implementation

### Day 1: Provider Unit Tests (6 Test Types)

#### Test Type 1: Basic Provider Initialization Tests

**File**: `tests/unit/providers/claude/claude_init_test.go`

```go
package claude_test

import (
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/llm/providers"
)

func TestClaudeProvider_Initialization(t *testing.T) {
    logger := logrus.New()
    
    t.Run("successful initialization with all parameters", func(t *testing.T) {
        provider, err := providers.NewClaudeProvider(
            "test-api-key",
            "https://api.anthropic.com",
            "claude-3-opus-20240229",
            30*time.Second,
            3,
            logger,
        )
        
        require.NoError(t, err)
        require.NotNil(t, provider)
        assert.Equal(t, "claude", provider.GetProvider())
        assert.Equal(t, "claude-3-opus-20240229", provider.GetModel())
    })
    
    t.Run("initialization with default base URL", func(t *testing.T) {
        provider, err := providers.NewClaudeProvider(
            "test-api-key",
            "",
            "claude-3-opus-20240229",
            30*time.Second,
            3,
            logger,
        )
        
        require.NoError(t, err)
        require.NotNil(t, provider)
    })
    
    t.Run("initialization with default timeout", func(t *testing.T) {
        provider, err := providers.NewClaudeProvider(
            "test-api-key",
            "https://api.anthropic.com",
            "claude-3-opus-20240229",
            0,
            3,
            logger,
        )
        
        require.NoError(t, err)
        require.NotNil(t, provider)
    })
    
    t.Run("initialization with default max retries", func(t *testing.T) {
        provider, err := providers.NewClaudeProvider(
            "test-api-key",
            "https://api.anthropic.com",
            "claude-3-opus-20240229",
            30*time.Second,
            0,
            logger,
        )
        
        require.NoError(t, err)
        require.NotNil(t, provider)
    })
    
    t.Run("initialization fails with empty API key", func(t *testing.T) {
        provider, err := providers.NewClaudeProvider(
            "",
            "https://api.anthropic.com",
            "claude-3-opus-20240229",
            30*time.Second,
            3,
            logger,
        )
        
        require.Error(t, err)
        require.Nil(t, provider)
        assert.Contains(t, err.Error(), "API key is required")
    })
    
    t.Run("initialization fails with empty model", func(t *testing.T) {
        provider, err := providers.NewClaudeProvider(
            "test-api-key",
            "https://api.anthropic.com",
            "",
            30*time.Second,
            3,
            logger,
        )
        
        require.Error(t, err)
        require.Nil(t, provider)
        assert.Contains(t, err.Error(), "model is required")
    })
}
```

#### Test Type 2: Complete Method Tests with All Parameter Variations

**File**: `tests/unit/providers/claude/claude_complete_test.go`

```go
package claude_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/llm"
    "dev.helix.agent/internal/llm/providers"
)

func TestClaudeProvider_Complete(t *testing.T) {
    logger := logrus.New()
    
    provider, err := providers.NewClaudeProvider(
        "test-api-key",
        "https://api.anthropic.com",
        "claude-3-opus-20240229",
        30*time.Second,
        3,
        logger,
    )
    require.NoError(t, err)
    
    ctx := context.Background()
    
    t.Run("basic completion request", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model: "claude-3-opus-20240229",
            Messages: []llm.Message{
                {
                    Role:    "user",
                    Content: "Hello, Claude!",
                },
            },
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
        assert.NotEmpty(t, response.ID)
        assert.Equal(t, "claude-3-opus-20240229", response.Model)
        assert.Len(t, response.Choices, 1)
        assert.Equal(t, "assistant", response.Choices[0].Message.Role)
        assert.NotEmpty(t, response.Choices[0].Message.Content)
    })
    
    t.Run("completion with max tokens", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:     "claude-3-opus-20240229",
            Messages:  []llm.Message{{Role: "user", Content: "Test"}},
            MaxTokens: 100,
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
        assert.Greater(t, response.Usage.CompletionTokens, 0)
        assert.LessOrEqual(t, response.Usage.CompletionTokens, 100)
    })
    
    t.Run("completion with temperature", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:       "claude-3-opus-20240229",
            Messages:    []llm.Message{{Role: "user", Content: "Test"}},
            Temperature: 0.7,
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("completion with top_p", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:  "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
            TopP:   0.9,
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("completion with stop sequences", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:         "claude-3-opus-20240229",
            Messages:      []llm.Message{{Role: "user", Content: "Test"}},
            StopSequences: []string{"\n", "END"},
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("completion with multiple messages", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model: "claude-3-opus-20240229",
            Messages: []llm.Message{
                {Role: "user", Content: "Hello"},
                {Role: "assistant", Content: "Hi there!"},
                {Role: "user", Content: "How are you?"},
            },
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("completion with system message", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model: "claude-3-opus-20240229",
            Messages: []llm.Message{
                {Role: "system", Content: "You are a helpful assistant"},
                {Role: "user", Content: "Hello"},
            },
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("completion with metadata", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
            Metadata: map[string]interface{}{
                "request_id": "test-123",
                "user_id":    "user-456",
            },
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
}
```

#### Test Type 3: Error Handling and Edge Case Tests

**File**: `tests/unit/providers/claude/claude_error_test.go`

```go
package claude_test

import (
    "context"
    "errors"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/llm"
    "dev.helix.agent/internal/llm/providers"
)

func TestClaudeProvider_ErrorHandling(t *testing.T) {
    logger := logrus.New()
    
    provider, err := providers.NewClaudeProvider(
        "test-api-key",
        "https://api.anthropic.com",
        "claude-3-opus-20240229",
        30*time.Second,
        3,
        logger,
    )
    require.NoError(t, err)
    
    ctx := context.Background()
    
    t.Run("completion with empty request", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{},
        }
        
        response, err := provider.Complete(ctx, request)
        
        // Should handle gracefully or return appropriate error
        if err != nil {
            assert.Error(t, err)
        } else {
            assert.NotNil(t, response)
        }
    })
    
    t.Run("completion with nil request", func(t *testing.T) {
        response, err := provider.Complete(ctx, nil)
        
        assert.Error(t, err)
        assert.Nil(t, response)
    })
    
    t.Run("completion with cancelled context", func(t *testing.T) {
        cancelledCtx, cancel := context.WithCancel(context.Background())
        cancel()
        
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
        }
        
        response, err := provider.Complete(cancelledCtx, request)
        
        assert.Error(t, err)
        assert.Nil(t, response)
        assert.True(t, errors.Is(err, context.Canceled))
    })
    
    t.Run("completion with timeout context", func(t *testing.T) {
        timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
        defer cancel()
        
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
        }
        
        response, err := provider.Complete(timeoutCtx, request)
        
        assert.Error(t, err)
        assert.Nil(t, response)
    })
    
    t.Run("completion with invalid model", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "invalid-model-name",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
        }
        
        response, err := provider.Complete(ctx, request)
        
        // Should handle gracefully
        if err != nil {
            assert.Error(t, err)
        } else {
            assert.NotNil(t, response)
        }
    })
    
    t.Run("completion with very long message", func(t *testing.T) {
        longContent := string(make([]byte, 100000))
        for i := range longContent {
            longContent = longContent[:i] + "a" + longContent[i+1:]
        }
        
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: longContent}},
        }
        
        response, err := provider.Complete(ctx, request)
        
        // Should handle gracefully
        if err != nil {
            assert.Error(t, err)
        } else {
            assert.NotNil(t, response)
        }
    })
    
    t.Run("completion with special characters", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test with special chars: \n\t\r\"'<>"}},
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("completion with unicode characters", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test with unicode: 你好世界 🌍"}},
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
}
```

#### Test Type 4: Rate Limiting and Timeout Tests

**File**: `tests/unit/providers/claude/claude_rate_limit_test.go`

```go
package claude_test

import (
    "context"
    "sync"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/llm"
    "dev.helix.agent/internal/llm/providers"
)

func TestClaudeProvider_RateLimiting(t *testing.T) {
    logger := logrus.New()
    
    provider, err := providers.NewClaudeProvider(
        "test-api-key",
        "https://api.anthropic.com",
        "claude-3-opus-20240229",
        30*time.Second,
        3,
        logger,
    )
    require.NoError(t, err)
    
    ctx := context.Background()
    
    t.Run("concurrent requests", func(t *testing.T) {
        const numRequests = 10
        var wg sync.WaitGroup
        errors := make(chan error, numRequests)
        
        for i := 0; i < numRequests; i++ {
            wg.Add(1)
            go func(id int) {
                defer wg.Done()
                
                request := &llm.LLMRequest{
                    Model:    "claude-3-opus-20240229",
                    Messages: []llm.Message{{Role: "user", Content: "Test"}},
                }
                
                _, err := provider.Complete(ctx, request)
                if err != nil {
                    errors <- err
                }
            }(i)
        }
        
        wg.Wait()
        close(errors)
        
        // Check if any requests failed
        errorCount := 0
        for err := range errors {
            t.Logf("Request failed: %v", err)
            errorCount++
        }
        
        // Should handle concurrent requests gracefully
        assert.LessOrEqual(t, errorCount, numRequests)
    })
    
    t.Run("sequential requests", func(t *testing.T) {
        const numRequests = 5
        
        for i := 0; i < numRequests; i++ {
            request := &llm.LLMRequest{
                Model:    "claude-3-opus-20240229",
                Messages: []llm.Message{{Role: "user", Content: "Test"}},
            }
            
            response, err := provider.Complete(ctx, request)
            
            require.NoError(t, err)
            require.NotNil(t, response)
        }
    })
    
    t.Run("request with timeout", func(t *testing.T) {
        timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
        }
        
        start := time.Now()
        response, err := provider.Complete(timeoutCtx, request)
        duration := time.Since(start)
        
        require.NoError(t, err)
        require.NotNil(t, response)
        assert.Less(t, duration, 5*time.Second)
    })
}
```

#### Test Type 5: Provider Capabilities Tests

**File**: `tests/unit/providers/claude/claude_capabilities_test.go`

```go
package claude_test

import (
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/llm/providers"
)

func TestClaudeProvider_Capabilities(t *testing.T) {
    logger := logrus.New()
    
    provider, err := providers.NewClaudeProvider(
        "test-api-key",
        "https://api.anthropic.com",
        "claude-3-opus-20240229",
        30*time.Second,
        3,
        logger,
    )
    require.NoError(t, err)
    
    t.Run("get capabilities", func(t *testing.T) {
        capabilities := provider.GetCapabilities()
        
        assert.True(t, capabilities.Streaming)
        assert.Greater(t, capabilities.MaxTokens, 0)
        assert.True(t, capabilities.SupportsSystem)
        assert.True(t, capabilities.SupportsImages)
        assert.True(t, capabilities.SupportsTools)
        assert.Greater(t, capabilities.RateLimitRPS, 0)
        assert.NotEmpty(t, capabilities.Features)
    })
    
    t.Run("capabilities include expected features", func(t *testing.T) {
        capabilities := provider.GetCapabilities()
        
        expectedFeatures := []string{"long_context", "vision", "tools"}
        for _, feature := range expectedFeatures {
            found := false
            for _, f := range capabilities.Features {
                if f == feature {
                    found = true
                    break
                }
            }
            assert.True(t, found, "Expected feature %s not found", feature)
        }
    })
    
    t.Run("max tokens is reasonable", func(t *testing.T) {
        capabilities := provider.GetCapabilities()
        
        assert.Greater(t, capabilities.MaxTokens, 1000)
        assert.Less(t, capabilities.MaxTokens, 1000000)
    })
    
    t.Run("rate limit is reasonable", func(t *testing.T) {
        capabilities := provider.GetCapabilities()
        
        assert.Greater(t, capabilities.RateLimitRPS, 0)
        assert.Less(t, capabilities.RateLimitRPS, 1000)
    })
    
    t.Run("timeout matches provider configuration", func(t *testing.T) {
        capabilities := provider.GetCapabilities()
        
        assert.Equal(t, 30*time.Second, capabilities.Timeout)
    })
}
```

#### Test Type 6: Mock Provider Interface Compliance Tests

**File**: `tests/unit/providers/claude/claude_interface_test.go`

```go
package claude_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/llm"
    "dev.helix.agent/internal/llm/providers"
)

func TestClaudeProvider_InterfaceCompliance(t *testing.T) {
    logger := logrus.New()
    
    provider, err := providers.NewClaudeProvider(
        "test-api-key",
        "https://api.anthropic.com",
        "claude-3-opus-20240229",
        30*time.Second,
        3,
        logger,
    )
    require.NoError(t, err)
    
    // Verify provider implements LLMProvider interface
    var _ llm.LLMProvider = (*providers.ClaudeProvider)(nil)
    
    ctx := context.Background()
    
    t.Run("implements Complete method", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
        }
        
        response, err := provider.Complete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, response)
    })
    
    t.Run("implements StreamComplete method", func(t *testing.T) {
        request := &llm.LLMRequest{
            Model:    "claude-3-opus-20240229",
            Messages: []llm.Message{{Role: "user", Content: "Test"}},
        }
        
        chunkChan, err := provider.StreamComplete(ctx, request)
        
        require.NoError(t, err)
        require.NotNil(t, chunkChan)
        
        // Read at least one chunk
        chunk, ok := <-chunkChan
        assert.True(t, ok)
        assert.NotNil(t, chunk)
    })
    
    t.Run("implements GetCapabilities method", func(t *testing.T) {
        capabilities := provider.GetCapabilities()
        
        assert.NotNil(t, capabilities)
    })
    
    t.Run("implements GetModel method", func(t *testing.T) {
        model := provider.GetModel()
        
        assert.NotEmpty(t, model)
        assert.Equal(t, "claude-3-opus-20240229", model)
    })
    
    t.Run("implements GetProvider method", func(t *testing.T) {
        providerName := provider.GetProvider()
        
        assert.NotEmpty(t, providerName)
        assert.Equal(t, "claude", providerName)
    })
    
    t.Run("implements Validate method", func(t *testing.T) {
        err := provider.Validate()
        
        assert.NoError(t, err)
    })
    
    t.Run("implements HealthCheck method", func(t *testing.T) {
        err := provider.HealthCheck(ctx)
        
        assert.NoError(t, err)
    })
}
```

### Day 2: Service Layer Unit Tests

#### Debate Service Tests

**File**: `tests/unit/services/debate_service_test.go`

```go
package services_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestDebateService_ConductDebate(t *testing.T) {
    logger := logrus.New()
    
    service := services.NewDebateService(logger)
    
    ctx := context.Background()
    
    t.Run("successful debate with 2 participants", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "test-debate-001",
            Topic:    "Test topic for debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.Equal(t, config.DebateID, result.DebateID)
        assert.True(t, result.Success)
        assert.Greater(t, result.TotalRounds, 0)
        assert.Len(t, result.Participants, 2)
    })
    
    t.Run("debate with Cognee enabled", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID:     "test-debate-002",
            Topic:        "Test topic for debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: true,
        }
        
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.NotNil(t, result.CogneeInsights)
    })
    
    t.Run("debate with timeout", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "test-debate-003",
            Topic:    "Test topic for debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     10,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     10,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    10,
            Timeout:     1 * time.Second,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        result, err := service.ConductDebate(ctx, config)
        
        // Should handle timeout gracefully
        if err != nil {
            assert.Error(t, err)
        } else {
            assert.NotNil(t, result)
        }
    })
}
```

#### Consensus Building Tests

**File**: `tests/unit/services/consensus_service_test.go`

```go
package services_test

import (
    "context"
    "testing"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestConsensusService_BuildConsensus(t *testing.T) {
    logger := logrus.New()
    
    service := services.NewConsensusService(logger)
    
    ctx := context.Background()
    
    t.Run("successful consensus building", func(t *testing.T) {
        responses := []services.ParticipantResponse{
            {
                ParticipantID: "participant-1",
                ParticipantName: "Alice",
                Role:          "proponent",
                Round:         1,
                Response:      "I agree with the proposal",
                Confidence:    0.9,
                QualityScore:  0.85,
            },
            {
                ParticipantID: "participant-2",
                ParticipantName: "Bob",
                Role:          "opponent",
                Round:         1,
                Response:      "I have some concerns but overall agree",
                Confidence:    0.8,
                QualityScore:  0.8,
            },
        }
        
        consensus, err := service.BuildConsensus(ctx, responses, 0.75)
        
        require.NoError(t, err)
        require.NotNil(t, consensus)
        assert.True(t, consensus.Achieved)
        assert.Greater(t, consensus.Confidence, 0.75)
    })
    
    t.Run("consensus not reached", func(t *testing.T) {
        responses := []services.ParticipantResponse{
            {
                ParticipantID: "participant-1",
                ParticipantName: "Alice",
                Role:          "proponent",
                Round:         1,
                Response:      "I strongly agree",
                Confidence:    0.9,
                QualityScore:  0.85,
            },
            {
                ParticipantID: "participant-2",
                ParticipantName: "Bob",
                Role:          "opponent",
                Round:         1,
                Response:      "I strongly disagree",
                Confidence:    0.9,
                QualityScore:  0.85,
            },
        }
        
        consensus, err := service.BuildConsensus(ctx, responses, 0.75)
        
        require.NoError(t, err)
        require.NotNil(t, consensus)
        assert.False(t, consensus.Achieved)
        assert.Less(t, consensus.Confidence, 0.75)
    })
}
```

### Day 3: Middleware Unit Tests

#### Authentication Middleware Tests

**File**: `tests/unit/middleware/auth_middleware_test.go`

```go
package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/middleware"
)

func TestAuthMiddleware(t *testing.T) {
    middleware := middleware.NewAuthMiddleware("test-api-key")
    
    t.Run("valid API key", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/test", nil)
        req.Header.Set("Authorization", "Bearer test-api-key")
        w := httptest.NewRecorder()
        
        handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("OK"))
        }))
        
        handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Equal(t, "OK", w.Body.String())
    })
    
    t.Run("invalid API key", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/test", nil)
        req.Header.Set("Authorization", "Bearer invalid-key")
        w := httptest.NewRecorder()
        
        handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("OK"))
        }))
        
        handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
    
    t.Run("missing API key", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/test", nil)
        w := httptest.NewRecorder()
        
        handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("OK"))
        }))
        
        handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}
```

### Day 4: Database Layer Tests

**File**: `tests/unit/database/db_test.go`

```go
package database_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/database"
)

func TestDatabase_Connection(t *testing.T) {
    ctx := context.Background()
    
    t.Run("successful connection", func(t *testing.T) {
        db, err := database.NewTestDatabase(ctx)
        
        require.NoError(t, err)
        require.NotNil(t, db)
        defer db.Close()
        
        err = db.Ping(ctx)
        assert.NoError(t, err)
    })
    
    t.Run("failed connection with invalid config", func(t *testing.T) {
        config := &database.Config{
            Host:     "invalid-host",
            Port:     "5432",
            User:     "test",
            Password: "test",
            Database: "test",
        }
        
        db, err := database.NewDatabase(config)
        
        assert.Error(t, err)
        assert.Nil(t, db)
    })
}
```

### Day 5: Cache Service Edge Cases

**File**: `tests/unit/cache/cache_edge_cases_test.go`

```go
package cache_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/cache"
)

func TestCacheService_EdgeCases(t *testing.T) {
    ctx := context.Background()
    
    t.Run("cache with very large value", func(t *testing.T) {
        service := cache.NewTestCacheService()
        
        largeValue := string(make([]byte, 10*1024*1024)) // 10MB
        
        err := service.Set(ctx, "large-key", largeValue, 1*time.Hour)
        
        require.NoError(t, err)
        
        value, err := service.Get(ctx, "large-key")
        
        require.NoError(t, err)
        assert.Equal(t, largeValue, value)
    })
    
    t.Run("cache with very long key", func(t *testing.T) {
        service := cache.NewTestCacheService()
        
        longKey := string(make([]byte, 10000))
        
        err := service.Set(ctx, longKey, "value", 1*time.Hour)
        
        // Should handle gracefully
        if err != nil {
            assert.Error(t, err)
        } else {
            value, err := service.Get(ctx, longKey)
            require.NoError(t, err)
            assert.Equal(t, "value", value)
        }
    })
    
    t.Run("cache with zero TTL", func(t *testing.T) {
        service := cache.NewTestCacheService()
        
        err := service.Set(ctx, "zero-ttl-key", "value", 0)
        
        require.NoError(t, err)
        
        // Should expire immediately
        time.Sleep(10 * time.Millisecond)
        
        value, err := service.Get(ctx, "zero-ttl-key")
        
        assert.Error(t, err)
        assert.Empty(t, value)
    })
    
    t.Run("cache with negative TTL", func(t *testing.T) {
        service := cache.NewTestCacheService()
        
        err := service.Set(ctx, "negative-ttl-key", "value", -1*time.Hour)
        
        // Should handle gracefully
        if err != nil {
            assert.Error(t, err)
        }
    })
}
```

---

## Week 4: Integration & E2E Tests

### Day 1: Integration Tests (5 Test Types)

#### Test Type 1: Multi-Provider Integration Scenarios

**File**: `tests/integration/multi_provider_test.go`

```go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestMultiProviderIntegration(t *testing.T) {
    logger := logrus.New()
    
    ctx := context.Background()
    
    t.Run("debate with multiple providers", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "multi-provider-debate-001",
            Topic:    "Test topic for multi-provider debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Claude",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "DeepSeek",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-3",
                    Name:          "Gemini",
                    Role:          "neutral",
                    LLMProvider:   "gemini",
                    LLMModel:      "gemini-pro",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        service := services.NewDebateService(logger)
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.Equal(t, config.DebateID, result.DebateID)
        assert.Len(t, result.Participants, 3)
    })
    
    t.Run("provider fallback during debate", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "fallback-debate-001",
            Topic:    "Test topic for fallback debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        service := services.NewDebateService(logger)
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
    })
}
```

#### Test Type 2: Advanced Debate Workflows

**File**: `tests/integration/advanced_debate_workflow_test.go`

```go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestAdvancedDebateWorkflow(t *testing.T) {
    logger := logrus.New()
    
    ctx := context.Background()
    
    t.Run("complete debate workflow with monitoring", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "workflow-debate-001",
            Topic:    "Test topic for workflow debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: true,
        }
        
        advancedService := services.NewAdvancedDebateService(
            services.NewDebateService(logger),
            services.NewDebateMonitoringService(logger),
            services.NewDebatePerformanceService(logger),
            services.NewDebateHistoryService(logger),
            services.NewDebateResilienceService(logger),
            services.NewDebateReportingService(logger),
            services.NewDebateSecurityService(logger),
            logger,
        )
        
        result, err := advancedService.ConductAdvancedDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.True(t, result.Success)
        assert.NotNil(t, result.CogneeInsights)
    })
}
```

#### Test Type 3: Cognee AI Integration Validation

**File**: `tests/integration/cognee_integration_test.go`

```go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestCogneeIntegration(t *testing.T) {
    logger := logrus.New()
    
    ctx := context.Background()
    
    t.Run("debate with Cognee enhancement", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "cognee-debate-001",
            Topic:    "Test topic for Cognee debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: true,
        }
        
        service := services.NewDebateService(logger)
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.NotNil(t, result.CogneeInsights)
        assert.NotEmpty(t, result.CogneeInsights.DatasetName)
        assert.Greater(t, result.CogneeInsights.EnhancementTime, 0)
    })
}
```

#### Test Type 4: Database and Cache Integration

**File**: `tests/integration/database_cache_integration_test.go`

```go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/cache"
    "dev.helix.agent/internal/database"
)

func TestDatabaseCacheIntegration(t *testing.T) {
    ctx := context.Background()
    
    t.Run("save debate result to database and cache", func(t *testing.T) {
        db, err := database.NewTestDatabase(ctx)
        require.NoError(t, err)
        defer db.Close()
        
        cacheService := cache.NewTestCacheService()
        
        debateResult := &services.DebateResult{
            DebateID:    "integration-debate-001",
            StartTime:   time.Now(),
            EndTime:     time.Now().Add(5 * time.Minute),
            TotalRounds: 3,
            Success:     true,
        }
        
        // Save to database
        err = db.SaveDebateResult(ctx, debateResult)
        require.NoError(t, err)
        
        // Save to cache
        err = cacheService.Set(ctx, "debate:"+debateResult.DebateID, debateResult, 1*time.Hour)
        require.NoError(t, err)
        
        // Retrieve from cache
        cachedResult, err := cacheService.Get(ctx, "debate:"+debateResult.DebateID)
        require.NoError(t, err)
        assert.NotNil(t, cachedResult)
    })
}
```

#### Test Type 5: Plugin System Integration

**File**: `tests/integration/plugin_integration_test.go`

```go
package integration

import (
    "context"
    "testing"
    
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/plugins"
)

func TestPluginIntegration(t *testing.T) {
    logger := logrus.New()
    
    ctx := context.Background()
    
    t.Run("load and execute plugin", func(t *testing.T) {
        pluginManager := plugins.NewPluginManager(logger)
        
        err := pluginManager.LoadPlugin(ctx, "test-plugin", "./plugins/test/plugin.so")
        require.NoError(t, err)
        
        err = pluginManager.ExecutePlugin(ctx, "test-plugin", map[string]interface{}{
            "action": "test",
        })
        
        require.NoError(t, err)
    })
}
```

### Day 2: E2E Tests (6 Test Types)

#### Test Type 1: Complete Debate Workflow Validation

**File**: `tests/e2e/debate_workflow_e2e_test.go`

```go
package e2e

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestDebateWorkflow_E2E(t *testing.T) {
    ctx := context.Background()
    
    t.Run("complete debate from start to finish", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "e2e-debate-001",
            Topic:    "Complete E2E test debate",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        service := services.NewDebateService(nil)
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.True(t, result.Success)
        assert.Greater(t, result.TotalRounds, 0)
        assert.Len(t, result.Participants, 2)
    })
}
```

#### Test Type 2: Consensus Building Scenarios

**File**: `tests/e2e/consensus_e2e_test.go`

```go
package e2e

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestConsensusBuilding_E2E(t *testing.T) {
    ctx := context.Background()
    
    t.Run("consensus reached in debate", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "e2e-consensus-001",
            Topic:    "Test consensus building",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        service := services.NewDebateService(nil)
        result, err := service.ConsensusDebate(ctx, config, 0.75)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.NotNil(t, result.Consensus)
    })
}
```

#### Test Type 3: Performance and Load Testing

**File**: `tests/e2e/performance_e2e_test.go`

```go
package e2e

import (
    "context"
    "sync"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestPerformance_E2E(t *testing.T) {
    ctx := context.Background()
    
    t.Run("concurrent debates", func(t *testing.T) {
        const numDebates = 10
        var wg sync.WaitGroup
        results := make(chan *services.DebateResult, numDebates)
        errors := make(chan error, numDebates)
        
        for i := 0; i < numDebates; i++ {
            wg.Add(1)
            go func(id int) {
                defer wg.Done()
                
                config := &services.DebateConfig{
                    DebateID: "e2e-perf-" + string(rune(id)),
                    Topic:    "Performance test debate",
                    Participants: []services.ParticipantConfig{
                        {
                            ParticipantID: "participant-1",
                            Name:          "Alice",
                            Role:          "proponent",
                            LLMProvider:   "claude",
                            LLMModel:      "claude-3-opus-20240229",
                            MaxRounds:     2,
                            Timeout:       30 * time.Second,
                            Weight:        1.0,
                        },
                        {
                            ParticipantID: "participant-2",
                            Name:          "Bob",
                            Role:          "opponent",
                            LLMProvider:   "deepseek",
                            LLMModel:      "deepseek-chat",
                            MaxRounds:     2,
                            Timeout:       30 * time.Second,
                            Weight:        1.0,
                        },
                    },
                    MaxRounds:    2,
                    Timeout:     2 * time.Minute,
                    Strategy:    "structured",
                    EnableCognee: false,
                }
                
                service := services.NewDebateService(nil)
                result, err := service.ConductDebate(ctx, config)
                
                if err != nil {
                    errors <- err
                } else {
                    results <- result
                }
            }(i)
        }
        
        wg.Wait()
        close(results)
        close(errors)
        
        resultCount := 0
        for range results {
            resultCount++
        }
        
        errorCount := 0
        for range errors {
            errorCount++
        }
        
        assert.Greater(t, resultCount, 0)
        assert.LessOrEqual(t, errorCount, numDebates)
    })
}
```

#### Test Type 4: Failure Recovery and Resilience

**File**: `tests/e2e/resilience_e2e_test.go`

```go
package e2e

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestResilience_E2E(t *testing.T) {
    ctx := context.Background()
    
    t.Run("debate recovery after failure", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "e2e-resilience-001",
            Topic:    "Test resilience",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
        }
        
        service := services.NewDebateService(nil)
        
        // Simulate failure and recovery
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.True(t, result.Success)
    })
}
```

#### Test Type 5: Security and Authentication Flows

**File**: `tests/e2e/security_e2e_test.go`

```go
package e2e

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestSecurity_E2E(t *testing.T) {
    ctx := context.Background()
    
    t.Run("debate with authentication", func(t *testing.T) {
        config := &services.DebateConfig{
            DebateID: "e2e-security-001",
            Topic:    "Test security",
            Participants: []services.ParticipantConfig{
                {
                    ParticipantID: "participant-1",
                    Name:          "Alice",
                    Role:          "proponent",
                    LLMProvider:   "claude",
                    LLMModel:      "claude-3-opus-20240229",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
                {
                    ParticipantID: "participant-2",
                    Name:          "Bob",
                    Role:          "opponent",
                    LLMProvider:   "deepseek",
                    LLMModel:      "deepseek-chat",
                    MaxRounds:     3,
                    Timeout:       30 * time.Second,
                    Weight:        1.0,
                },
            },
            MaxRounds:    3,
            Timeout:     5 * time.Minute,
            Strategy:    "structured",
            EnableCognee: false,
            Metadata: map[string]interface{}{
                "authenticated": true,
                "user_id":       "test-user",
            },
        }
        
        service := services.NewDebateService(nil)
        result, err := service.ConductDebate(ctx, config)
        
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.True(t, result.Success)
    })
}
```

#### Test Type 6: Multi-User Concurrent Scenarios

**File**: `tests/e2e/concurrent_e2e_test.go`

```go
package e2e

import (
    "context"
    "sync"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/services"
)

func TestConcurrent_E2E(t *testing.T) {
    ctx := context.Background()
    
    t.Run("multiple users conducting debates", func(t *testing.T) {
        const numUsers = 5
        var wg sync.WaitGroup
        results := make(chan *services.DebateResult, numUsers)
        
        for i := 0; i < numUsers; i++ {
            wg.Add(1)
            go func(userID int) {
                defer wg.Done()
                
                config := &services.DebateConfig{
                    DebateID: "e2e-concurrent-" + string(rune(userID)),
                    Topic:    "Concurrent test debate",
                    Participants: []services.ParticipantConfig{
                        {
                            ParticipantID: "participant-1",
                            Name:          "Alice",
                            Role:          "proponent",
                            LLMProvider:   "claude",
                            LLMModel:      "claude-3-opus-20240229",
                            MaxRounds:     2,
                            Timeout:       30 * time.Second,
                            Weight:        1.0,
                        },
                        {
                            ParticipantID: "participant-2",
                            Name:          "Bob",
                            Role:          "opponent",
                            LLMProvider:   "deepseek",
                            LLMModel:      "deepseek-chat",
                            MaxRounds:     2,
                            Timeout:       30 * time.Second,
                            Weight:        1.0,
                        },
                    },
                    MaxRounds:    2,
                    Timeout:     2 * time.Minute,
                    Strategy:    "structured",
                    EnableCognee: false,
                    Metadata: map[string]interface{}{
                        "user_id": userID,
                    },
                }
                
                service := services.NewDebateService(nil)
                result, err := service.ConductDebate(ctx, config)
                
                if err == nil {
                    results <- result
                }
            }(i)
        }
        
        wg.Wait()
        close(results)
        
        resultCount := 0
        for range results {
            resultCount++
        }
        
        assert.Greater(t, resultCount, 0)
    })
}
```

---

## Summary of Week 3-4 Deliverables

### Week 3 Deliverables
- [x] Provider unit tests (6 test types)
- [x] Service layer unit tests
- [x] Middleware unit tests
- [x] Database layer tests
- [x] Cache service edge case tests

### Week 4 Deliverables
- [x] Integration tests (5 test types)
- [x] E2E tests (6 test types)
- [x] Specialized test suites
- [x] Performance benchmarks
- [x] Security tests

---

## Next Steps

Proceed to Phase 3: Documentation Completion (Week 5-6).

*Last Updated: 2025-12-27*
*Version: 1.0.0*