# Optimization API Reference

Complete API reference for the LLM Optimization Framework.

## OptimizationService

The central service that coordinates all optimization capabilities.

### NewService

```go
func NewService(config *Config) (*Service, error)
```

Creates a new optimization service.

**Parameters:**
- `config`: Configuration for the service. Pass `nil` for defaults.

**Returns:**
- `*Service`: The optimization service
- `error`: Error if configuration is invalid

**Example:**
```go
config := optimization.DefaultConfig()
svc, err := optimization.NewService(config)
```

### OptimizeRequest

```go
func (s *Service) OptimizeRequest(ctx context.Context, prompt string, embedding []float64) (*OptimizedRequest, error)
```

Optimizes an LLM request by checking cache, retrieving context, and decomposing tasks.

**Parameters:**
- `ctx`: Context for cancellation
- `prompt`: The user's prompt
- `embedding`: Vector embedding of the prompt (optional)

**Returns:**
- `*OptimizedRequest`: Optimization results
- `error`: Error if optimization fails

**OptimizedRequest Fields:**
```go
type OptimizedRequest struct {
    OriginalPrompt    string     // The original prompt
    OptimizedPrompt   string     // Enhanced prompt with context
    CacheHit          bool       // True if response was cached
    CachedResponse    string     // Cached response (if hit)
    RetrievedContext  []string   // Context from document retrieval
    DecomposedTasks   []string   // Subtasks (if complex task)
    WarmPrefix        bool       // True if prefix was warmed
}
```

### OptimizeResponse

```go
func (s *Service) OptimizeResponse(ctx context.Context, response string, embedding []float64, query string, schema *outlines.JSONSchema) (*OptimizedResponse, error)
```

Processes and caches an LLM response.

**Parameters:**
- `ctx`: Context for cancellation
- `response`: The LLM response
- `embedding`: Vector embedding for caching
- `query`: Original query (for cache key)
- `schema`: JSON schema for validation (optional)

**Returns:**
- `*OptimizedResponse`: Optimization results
- `error`: Error if optimization fails

**OptimizedResponse Fields:**
```go
type OptimizedResponse struct {
    Content           string                    // The response content
    Cached            bool                      // True if response was cached
    StructuredOutput  interface{}               // Parsed structured data
    ValidationResult  *outlines.ValidationResult // Schema validation result
    StreamingMetrics  *streaming.AggregatedStream // Streaming stats
}
```

### StreamEnhanced

```go
func (s *Service) StreamEnhanced(ctx context.Context, stream <-chan *streaming.StreamChunk, progress streaming.ProgressCallback) (<-chan *streaming.StreamChunk, func() *streaming.AggregatedStream)
```

Enhances a streaming response with buffering and progress tracking.

**Parameters:**
- `ctx`: Context for cancellation
- `stream`: Input stream channel
- `progress`: Progress callback function (optional)

**Returns:**
- `<-chan *streaming.StreamChunk`: Enhanced output stream
- `func() *streaming.AggregatedStream`: Function to get final stats

### GenerateStructured

```go
func (s *Service) GenerateStructured(ctx context.Context, prompt string, schema *outlines.JSONSchema, generator func(string) (string, error)) (*outlines.StructuredResponse, error)
```

Generates and validates structured output.

**Parameters:**
- `ctx`: Context for cancellation
- `prompt`: The generation prompt
- `schema`: JSON schema for validation
- `generator`: Function that generates the response

**Returns:**
- `*outlines.StructuredResponse`: Generation result
- `error`: Error if generation fails

### DecomposeTask

```go
func (s *Service) DecomposeTask(ctx context.Context, task string) (*langchain.DecomposeResponse, error)
```

Decomposes a complex task into subtasks.

**Parameters:**
- `ctx`: Context for cancellation
- `task`: The task to decompose

**Returns:**
- `*langchain.DecomposeResponse`: Decomposition result
- `error`: Error if service unavailable

### RunReActAgent

```go
func (s *Service) RunReActAgent(ctx context.Context, goal string, tools []string) (*langchain.ReActResponse, error)
```

Runs a ReAct reasoning agent.

**Parameters:**
- `ctx`: Context for cancellation
- `goal`: The agent's goal
- `tools`: Available tools for the agent

**Returns:**
- `*langchain.ReActResponse`: Agent execution result
- `error`: Error if service unavailable

### QueryDocuments

```go
func (s *Service) QueryDocuments(ctx context.Context, query string, options *llamaindex.QueryRequest) (*llamaindex.QueryResponse, error)
```

Queries documents with advanced retrieval.

**Parameters:**
- `ctx`: Context for cancellation
- `query`: Search query
- `options`: Query options (optional)

**Returns:**
- `*llamaindex.QueryResponse`: Query results
- `error`: Error if service unavailable

### GenerateConstrained

```go
func (s *Service) GenerateConstrained(ctx context.Context, prompt string, constraints []lmql.Constraint) (*lmql.ConstrainedResponse, error)
```

Generates text with constraints.

**Parameters:**
- `ctx`: Context for cancellation
- `prompt`: Generation prompt
- `constraints`: List of constraints

**Returns:**
- `*lmql.ConstrainedResponse`: Generation result
- `error`: Error if service unavailable

### SelectFromOptions

```go
func (s *Service) SelectFromOptions(ctx context.Context, prompt string, options []string) (string, error)
```

Selects from constrained options.

**Parameters:**
- `ctx`: Context for cancellation
- `prompt`: Selection prompt
- `options`: Available options

**Returns:**
- `string`: Selected option
- `error`: Error if service unavailable

### CreateSession

```go
func (s *Service) CreateSession(ctx context.Context, sessionID, systemPrompt string) error
```

Creates a new conversation session with prefix caching.

**Parameters:**
- `ctx`: Context for cancellation
- `sessionID`: Unique session identifier
- `systemPrompt`: System prompt to cache

**Returns:**
- `error`: Error if service unavailable

### ContinueSession

```go
func (s *Service) ContinueSession(ctx context.Context, sessionID, message string) (string, error)
```

Continues a conversation with prefix caching.

**Parameters:**
- `ctx`: Context for cancellation
- `sessionID`: Existing session ID
- `message`: User message

**Returns:**
- `string`: Assistant response
- `error`: Error if service unavailable

### GetCacheStats

```go
func (s *Service) GetCacheStats() map[string]interface{}
```

Returns semantic cache statistics.

**Returns:**
- `map[string]interface{}`: Cache statistics

**Response Fields:**
```go
{
    "enabled":  bool,    // Cache enabled
    "hits":     int64,   // Total cache hits
    "misses":   int64,   // Total cache misses
    "entries":  int,     // Current entry count
    "hit_rate": float64, // Hit rate (0.0-1.0)
}
```

### GetServiceStatus

```go
func (s *Service) GetServiceStatus(ctx context.Context) map[string]bool
```

Returns the status of all external services.

**Returns:**
- `map[string]bool`: Service availability map

**Example Response:**
```go
{
    "sglang":     true,
    "llamaindex": false,
    "langchain":  true,
    "guidance":   true,
    "lmql":       false,
}
```

---

## Configuration

### Config Structure

```go
type Config struct {
    Enabled          bool
    SemanticCache    SemanticCacheConfig
    StructuredOutput StructuredOutputConfig
    Streaming        StreamingConfig
    SGLang           SGLangConfig
    LlamaIndex       LlamaIndexConfig
    LangChain        LangChainConfig
    Guidance         GuidanceConfig
    LMQL             LMQLConfig
    Fallback         FallbackConfig
}
```

### SemanticCacheConfig

```go
type SemanticCacheConfig struct {
    Enabled             bool          // Enable semantic cache
    SimilarityThreshold float64       // Similarity threshold (0.0-1.0)
    MaxEntries          int           // Maximum cache entries
    TTL                 time.Duration // Entry time-to-live
    EmbeddingModel      string        // Embedding model name
    EvictionPolicy      string        // lru, ttl, lru_with_relevance
}
```

### StreamingConfig

```go
type StreamingConfig struct {
    Enabled          bool          // Enable streaming
    BufferType       string        // character, word, sentence, etc.
    ProgressInterval time.Duration // Progress update interval
    RateLimit        float64       // Tokens per second (0 = unlimited)
}
```

### External Service Configs

```go
type SGLangConfig struct {
    Enabled               bool
    Endpoint              string
    Timeout               time.Duration
    FallbackOnUnavailable bool
}

type LlamaIndexConfig struct {
    Enabled        bool
    Endpoint       string
    Timeout        time.Duration
    UseCogneeIndex bool
}

type LangChainConfig struct {
    Enabled      bool
    Endpoint     string
    Timeout      time.Duration
    DefaultChain string
}

type GuidanceConfig struct {
    Enabled       bool
    Endpoint      string
    Timeout       time.Duration
    CachePrograms bool
}

type LMQLConfig struct {
    Enabled      bool
    Endpoint     string
    Timeout      time.Duration
    CacheQueries bool
}
```

### FallbackConfig

```go
type FallbackConfig struct {
    OnServiceUnavailable  string        // skip, error, cache_only
    HealthCheckInterval   time.Duration // Health check frequency
    RetryUnavailableAfter time.Duration // Retry delay after failure
}
```

### DefaultConfig

```go
func DefaultConfig() *Config
```

Returns default configuration with sensible production defaults.

---

## Error Handling

### Common Errors

| Error | Cause | Resolution |
|-------|-------|------------|
| "service not available" | External service down | Check service status, enable fallback |
| "invalid config" | Bad configuration | Validate config values |
| "structured output not enabled" | Feature disabled | Enable in config |
| "cache miss" | Not found in cache | Normal operation, proceed with LLM |

### Error Checking

```go
result, err := svc.OptimizeRequest(ctx, prompt, embedding)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "not available"):
        // Service unavailable, use fallback
    case strings.Contains(err.Error(), "timeout"):
        // Increase timeout or retry
    default:
        // Handle other errors
    }
}
```

---

## HTTP Endpoints Summary

### Native Services (Embedded)

No HTTP endpoints - direct Go API calls.

### External Services

| Service | Port | Base Path |
|---------|------|-----------|
| LangChain | 8011 | /decompose, /chain, /react |
| LlamaIndex | 8012 | /query, /hyde, /rerank |
| Guidance | 8013 | /grammar, /template, /select |
| LMQL | 8014 | /query, /constrained, /decode |
| SGLang | 30000 | /v1/chat/completions |

All services expose `/health` for health checks.
