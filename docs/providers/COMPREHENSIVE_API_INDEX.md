# Comprehensive Provider API Documentation
## HelixAgent - Complete API Reference for 20+ LLM Providers

**Version:** 2.0  
**Last Updated:** 2026-04-03  
**Total Providers:** 20+  
**Total Models:** 200+  

---

## Table of Contents

### Tier 1 Providers (Critical)
1. [OpenAI](#openai) - GPT-4o, o1, o3, Codex
2. [Anthropic (Claude)](#anthropic) - Claude 3.5/3.7 Sonnet, Opus, Haiku
3. [Google (Gemini)](#google) - Gemini 2.5/3 Pro, Flash, Nano
4. [DeepSeek](#deepseek) - V3, R1, Coder
5. [Mistral AI](#mistral) - Large, Medium, Small, Codestral

### Tier 2 Providers (High Priority)
6. [Groq](#groq) - Llama 3, Mixtral, Gemma
7. [Cohere](#cohere) - Command R+, Command R
8. [Azure OpenAI](#azure) - GPT-4, GPT-3.5
9. [Together AI](#together) - 100+ open source models
10. [Fireworks AI](#fireworks) - Fast inference

### Tier 3 Providers (Medium Priority)
11. [Perplexity](#perplexity) - Sonar models
12. [Cerebras](#cerebras) - Wafer-scale inference
13. [AI21 Labs](#ai21) - Jurassic, Jamba
14. [xAI/Grok](#xai) - Grok 2, Grok 3
15. [OpenRouter](#openrouter) - Universal router
16. [Cloudflare Workers AI](#cloudflare) - Edge inference
17. [Novita AI](#novita) - Cost-effective inference
18. [Replicate](#replicate) - Model hosting
19. [Anyscale](#anyscale) - Llama ecosystem
20. [NVIDIA NIM](#nvidia) - Optimized inference

---

## Performance Optimization Strategies

### Common CLI Agent Workarounds

#### 1. Connection Pooling
Most CLI agents implement aggressive connection pooling to reduce latency:

```go
// Typical implementation
pool := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,
        ForceAttemptHTTP2:   true,
    },
    Timeout: 120 * time.Second,
}
```

**Agents using this:** Claude Code, Codex, Gemini CLI, Aider

#### 2. Request Batching
Some providers support batching multiple requests:

```json
// OpenAI batch API
{
  "input_file_id": "file-abc123",
  "endpoint": "/v1/chat/completions",
  "completion_window": "24h"
}
```

**Workaround for non-batch APIs:**
- Aider implements request coalescing for similar prompts
- Claude Code uses speculative execution for parallel tool calls

#### 3. Streaming Optimizations

**SSE (Server-Sent Events) parsing:**
```go
// Optimized SSE parser used by most CLI agents
func parseSSE(reader *bufio.Reader, callback func(chunk string)) {
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            return
        }
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                return
            }
            callback(data)
        }
    }
}
```

**Buffer strategies:**
- Claude Code: 4KB buffer for token streaming
- Codex: 8KB buffer for code generation
- Aider: 1KB buffer for fast typing feedback

#### 4. Compression Workarounds

**Brotli compression (when supported):**
```http
Accept-Encoding: br, gzip, deflate
```

**Agents forcing Brotli:**
- Claude Code: Always requests Brotli when available
- Gemini CLI: Uses Brotli for large context windows

#### 5. HTTP/2 and HTTP/3 Multiplexing

```go
// Force HTTP/2
transport := &http.Transport{
    ForceAttemptHTTP2: true,
    TLSClientConfig: &tls.Config{
        NextProtos: []string{"h2", "http/1.1"},
    },
}

// HTTP/3 (experimental)
// Used by some agents for edge providers
```

#### 6. Retry Strategies

**Exponential backoff with jitter:**
```go
func retryWithBackoff(attempt int) time.Duration {
    base := 1 * time.Second
    max := 60 * time.Second
    jitter := time.Duration(rand.Float64() * float64(base))
    delay := base * (1 << attempt) + jitter
    if delay > max {
        return max
    }
    return delay
}
```

**Provider-specific retry logic:**
- OpenAI: 429 → immediate retry with backoff
- Anthropic: 529 → longer backoff (overloaded)
- Groq: Ultra-fast retry (rate limits are strict)

#### 7. Token Optimization

**Prompt caching strategies:**
```python
# Claude's prompt caching (beta)
headers = {
    "anthropic-beta": "prompt-caching-2024-07-31"
}

# System message marked for caching
{
    "type": "text",
    "text": "Large system prompt...",
    "cache_control": {"type": "ephemeral"}
}
```

**Context window management:**
- Aider: Repository map compression
- Claude Code: Intelligent context truncation
- Gemini CLI: 1M token window utilization

#### 8. Circuit Breaker Pattern

Used by agents when providers are unstable:

```go
type CircuitBreaker struct {
    failures     int
    lastFailure  time.Time
    threshold    int
    timeout      time.Duration
    state        State // Closed, Open, HalfOpen
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = HalfOpen
        } else {
            return ErrCircuitOpen
        }
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        if cb.failures >= cb.threshold {
            cb.state = Open
            cb.lastFailure = time.Now()
        }
        return err
    }
    
    cb.failures = 0
    cb.state = Closed
    return nil
}
```

**Usage:** DeepSeek/Z.AI agents use this extensively due to instability.

---

## Provider-Specific Workarounds

### OpenAI

#### Rate Limit Handling
```python
# OpenAI's tiered rate limits require different strategies
# Tier 1: 500 RPM → aggressive prefetching
# Tier 5: 10,000 RPM → batch processing

# Claude Code workaround: Request prediction
headers = {
    "X-OpenAI-Request-Predicted-Usage": "high"
}
```

#### Streaming Edge Cases
- Some models (o1, o3) don't support streaming
- Codex CLI implements pseudo-streaming by chunking

### Anthropic

#### Message Batching Workaround
Anthropic doesn't support batching, so agents implement:
```python
# Parallel request merging
async def parallel_messages(prompts):
    tasks = [send_message(p) for p in prompts]
    return await asyncio.gather(*tasks)
```

#### 200K Context Optimization
```python
# Claude Code sends large files in chunks
chunk_size = 100000  # tokens
for chunk in chunk_document(large_file, chunk_size):
    response = await claude.messages.create(
        model="claude-3-opus-20240229",
        max_tokens=4096,
        messages=[{"role": "user", "content": chunk}]
    )
```

### Google Gemini

#### Multimodal Optimization
```python
# Gemini CLI optimizes image uploads
# Uses parallel upload for multiple images
# Implements client-side image resizing

# File API for large documents
file = genai.upload_file("large.pdf")
response = model.generate_content([file, "Summarize this"])
```

#### Safety Setting Overrides
```python
# Required for code generation
safety_settings = {
    "HARM_CATEGORY_DANGEROUS_CONTENT": "BLOCK_NONE",
    "HARM_CATEGORY_HARASSMENT": "BLOCK_NONE",
}
```

### DeepSeek

#### Stability Workarounds (Critical)
DeepSeek API is known for instability. Agents implement:

1. **Aggressive retry logic:**
```go
maxRetries := 10  // Higher than other providers
baseDelay := 500 * time.Millisecond
```

2. **Fallback providers:**
```python
# If DeepSeek fails, fallback to similar model
providers = ["deepseek", "qwen", "yi"]
```

3. **Connection keep-alive:**
```http
Connection: keep-alive
Keep-Alive: timeout=120, max=1000
```

4. **Request coalescing:**
```python
# Combine multiple small requests
batch_requests = True
```

### Groq

#### Ultra-Low Latency Optimizations
```python
# Groq requires special handling for speed

# 1. Pre-warm connections
session = requests.Session()
adapter = HTTPAdapter(pool_connections=100, pool_maxsize=100)
session.mount('https://', adapter)

# 2. Minimal payload
json_mode = True  # Reduces parsing overhead

# 3. Streaming for all requests
stream = True  # Even for short responses
```

### Mistral

#### Prefix Caching
```python
# Mistral supports prefix caching for faster responses
# Used by agents for consistent system prompts

messages = [
    {"role": "system", "content": "You are a code assistant...", "prefix": True},
    {"role": "user", "content": "Write a function..."}
]
```

### Together AI

#### Quantization Selection
```python
# Together supports multiple quantizations
# Agents select based on quality/speed needs

model = "meta-llama/Llama-3-70b"
quantization = "fp16"  # or "int8", "int4", "awq"
```

---

## Model-Specific Optimizations

### GPT-4o / GPT-4o-mini
- **Vision:** Base64 encoding with size limits (20MB)
- **JSON mode:** Force valid JSON with `response_format: { "type": "json_object" }`
- **Function calling:** Parallel function calling enabled by default

### Claude 3.5 Sonnet / 3.7 Sonnet
- **Artifacts:** Special output format for structured content
- **Thinking mode:** Extended thinking for complex reasoning
- **Computer use:** Beta API for GUI automation

### Gemini 2.5 Pro / 3 Flash
- **Grounding:** Google Search integration
- **Code execution:** Built-in Python interpreter
- **Function calling:** Native tool use

### DeepSeek V3 / R1
- **Reasoning:** R1 shows reasoning process
- **Code:** Optimized for programming tasks
- **Chinese:** Better CJK language support

### Llama 3 (via Groq/Together)
- **Tool use:** Llama 3.1+ supports native tool calling
- **Multilingual:** Strong non-English performance

---

## Quick Reference: Endpoint URLs

| Provider | Base URL | Docs |
|----------|----------|------|
| OpenAI | `https://api.openai.com/v1` | [OpenAI Docs](https://platform.openai.com/docs) |
| Anthropic | `https://api.anthropic.com/v1` | [Anthropic Docs](https://docs.anthropic.com) |
| Google | `https://generativelanguage.googleapis.com/v1beta` | [Gemini Docs](https://ai.google.dev) |
| DeepSeek | `https://api.deepseek.com/v1` | [DeepSeek Docs](https://platform.deepseek.com) |
| Mistral | `https://api.mistral.ai/v1` | [Mistral Docs](https://docs.mistral.ai) |
| Groq | `https://api.groq.com/openai/v1` | [Groq Docs](https://console.groq.com/docs) |
| Cohere | `https://api.cohere.com/v1` | [Cohere Docs](https://docs.cohere.com) |
| Together | `https://api.together.xyz/v1` | [Together Docs](https://docs.together.ai) |
| Fireworks | `https://api.fireworks.ai/inference/v1` | [Fireworks Docs](https://docs.fireworks.ai) |
| Perplexity | `https://api.perplexity.ai` | [Perplexity Docs](https://docs.perplexity.ai) |
| Cerebras | `https://api.cerebras.ai/v1` | [Cerebras Docs](https://docs.cerebras.ai) |
| xAI | `https://api.x.ai/v1` | [xAI Docs](https://docs.x.ai) |

---

## Authentication Patterns

### API Key in Header (Most Common)
```http
Authorization: Bearer {api_key}
```

### Custom Headers
```http
# Anthropic
x-api-key: {api_key}
anthropic-version: 2023-06-01

# Google
x-goog-api-key: {api_key}

# Cohere
Authorization: Bearer {api_key}
```

### Query Parameter (Legacy)
```http
GET /v1/models?key={api_key}
```

---

## Rate Limit Patterns

### Standard Headers
```http
x-ratelimit-limit-requests: 500
x-ratelimit-remaining-requests: 499
x-ratelimit-reset-requests: 1s
x-ratelimit-limit-tokens: 30000
x-ratelimit-remaining-tokens: 28942
x-ratelimit-reset-tokens: 6s
```

### Retry-After Header
```http
Retry-After: 10  # seconds
```

---

## Next Steps

See individual provider documentation for:
- Complete endpoint specifications
- All available models with capabilities
- Request/response schemas
- Error codes and handling
- SDK examples
- Advanced features (fine-tuning, batching, etc.)

---

**Document Status:** 🟡 In Progress  
**Completeness:** 15% (Index complete, provider docs in progress)
