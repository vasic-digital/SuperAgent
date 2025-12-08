# Data Model Specification

## Core Entities

### LLM Provider Configuration
```go
type LLMProvider struct {
    ID              string            `json:"id" db:"id"`
    Name            string            `json:"name" db:"name"`
    Type            string            `json:"type" db:"type"` // deepseek, claude, gemini, qwen, zai
    APIKey          string            `json:"-" db:"api_key"` // encrypted
    BaseURL         string            `json:"base_url" db:"base_url"`
    Model           string            `json:"model" db:"model"`
    Weight          float64           `json:"weight" db:"weight"`
    Enabled         bool              `json:"enabled" db:"enabled"`
    Config          JSON              `json:"config" db:"config"` // provider-specific settings
    HealthStatus    string            `json:"health_status" db:"health_status"`
    ResponseTime    int64             `json:"response_time" db:"response_time"` // milliseconds
    SuccessRate     float64           `json:"success_rate" db:"success_rate"`
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}
```

### LLM Request
```go
type LLMRequest struct {
    ID              string            `json:"id" db:"id"`
    SessionID       string            `json:"session_id" db:"session_id"`
    UserID          string            `json:"user_id" db:"user_id"`
    Prompt          string            `json:"prompt" db:"prompt"`
    Messages        []Message         `json:"messages" db:"messages"`
    ModelParams     ModelParameters   `json:"model_params" db:"model_params"`
    RequestType     string            `json:"request_type" db:"request_type"` // code_generation, reasoning, tool_use
    Priority        int               `json:"priority" db:"priority"`
    EnsembleConfig  *EnsembleConfig   `json:"ensemble_config" db:"ensemble_config"`
    MemoryEnhanced  bool              `json:"memory_enhanced" db:"memory_enhanced"`
    Status          string            `json:"status" db:"status"` // pending, processing, completed, failed
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
    StartedAt       *time.Time        `json:"started_at" db:"started_at"`
    CompletedAt     *time.Time        `json:"completed_at" db:"completed_at"`
}
```

### LLM Response
```go
type LLMResponse struct {
    ID              string            `json:"id" db:"id"`
    RequestID       string            `json:"request_id" db:"request_id"`
    ProviderID      string            `json:"provider_id" db:"provider_id"`
    ProviderName    string            `json:"provider_name" db:"provider_name"`
    Content         string            `json:"content" db:"content"`
    Confidence      float64           `json:"confidence" db:"confidence"`
    TokensUsed      int               `json:"tokens_used" db:"tokens_used"`
    ResponseTime    int64             `json:"response_time" db:"response_time"`
    FinishReason    string            `json:"finish_reason" db:"finish_reason"`
    Metadata        JSON              `json:"metadata" db:"metadata"` // provider-specific data
    Selected        bool              `json:"selected" db:"selected"` // ensemble winner
    SelectionScore  float64           `json:"selection_score" db:"selection_score"`
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}
```

### User Session
```go
type UserSession struct {
    ID              string            `json:"id" db:"id"`
    UserID          string            `json:"user_id" db:"user_id"`
    SessionToken    string            `json:"-" db:"session_token"` // hashed
    Context         JSON              `json:"context" db:"context"` // conversation context
    MemoryID        *string           `json:"memory_id" db:"memory_id"`
    Status          string            `json:"status" db:"status"` // active, expired, terminated
    RequestCount    int               `json:"request_count" db:"request_count"`
    LastActivity    time.Time         `json:"last_activity" db:"last_activity"`
    ExpiresAt       time.Time         `json:"expires_at" db:"expires_at"`
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}
```

### Cognee Memory
```go
type CogneeMemory struct {
    ID              string            `json:"id" db:"id"`
    SessionID       *string           `json:"session_id" db:"session_id"`
    DatasetName     string            `json:"dataset_name" db:"dataset_name"`
    ContentType     string            `json:"content_type" db:"content_type"` // prompt, response, interaction
    Content         string            `json:"content" db:"content"`
    VectorID        string            `json:"vector_id" db:"vector_id"` // Cognee vector reference
    GraphNodes      JSON              `json:"graph_nodes" db:"graph_nodes"` // Cognee graph entities
    SearchKey       string            `json:"search_key" db:"search_key"` // for quick lookup
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}
```

## Supporting Data Structures

### Message
```go
type Message struct {
    Role    string `json:"role" db:"role"` // system, user, assistant, tool
    Content string `json:"content" db:"content"`
    Name    *string `json:"name" db:"name"` // optional for tool messages
}
```

### ModelParameters
```go
type ModelParameters struct {
    Model            string            `json:"model" db:"model"`
    Temperature      float64           `json:"temperature" db:"temperature"`
    MaxTokens        int               `json:"max_tokens" db:"max_tokens"`
    TopP             float64           `json:"top_p" db:"top_p"`
    StopSequences    []string          `json:"stop_sequences" db:"stop_sequences"`
    ProviderSpecific map[string]interface{} `json:"provider_specific" db:"provider_specific"`
}
```

### EnsembleConfig
```go
type EnsembleConfig struct {
    Strategy             string    `json:"strategy" db:"strategy"` // confidence_weighted, majority_vote
    MinProviders         int       `json:"min_providers" db:"min_providers"`
    ConfidenceThreshold  float64   `json:"confidence_threshold" db:"confidence_threshold"`
    FallbackToBest       bool      `json:"fallback_to_best" db:"fallback_to_best"`
    Timeout              int       `json:"timeout" db:"timeout"` // seconds
}
```

## Relationships

### Request-Response Relationship
- One `LLMRequest` can have multiple `LLMResponse` entries (one per provider)
- Ensemble selection marks one response as `Selected: true`

### Session-Request Relationship
- One `UserSession` can have multiple `LLMRequest` entries
- Session context propagates through requests

### Memory-Request Relationship
- `CogneeMemory` can be linked to `UserSession` for context
- Memory enhancement uses semantic search against memory entries

### Provider-Response Relationship
- One `LLMProvider` generates multiple `LLMResponse` entries
- Provider metrics updated from response data

## Database Schema

### Tables

#### llm_providers
```sql
CREATE TABLE llm_providers (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(100) NOT NULL,
    api_key TEXT NOT NULL, -- encrypted
    base_url VARCHAR(500) NOT NULL,
    model VARCHAR(255) NOT NULL,
    weight DECIMAL(3,2) DEFAULT 1.0,
    enabled BOOLEAN DEFAULT true,
    config JSONB,
    health_status VARCHAR(50) DEFAULT 'unknown',
    response_time BIGINT DEFAULT 0,
    success_rate DECIMAL(5,4) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_llm_providers_enabled ON llm_providers(enabled);
CREATE INDEX idx_llm_providers_type ON llm_providers(type);
```

#### llm_requests
```sql
CREATE TABLE llm_requests (
    id VARCHAR(255) PRIMARY KEY,
    session_id VARCHAR(255) REFERENCES user_sessions(id),
    user_id VARCHAR(255) NOT NULL,
    prompt TEXT NOT NULL,
    messages JSONB NOT NULL,
    model_params JSONB NOT NULL,
    request_type VARCHAR(50) NOT NULL,
    priority INTEGER DEFAULT 0,
    ensemble_config JSONB,
    memory_enhanced BOOLEAN DEFAULT false,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_llm_requests_session_id ON llm_requests(session_id);
CREATE INDEX idx_llm_requests_user_id ON llm_requests(user_id);
CREATE INDEX idx_llm_requests_status ON llm_requests(status);
CREATE INDEX idx_llm_requests_created_at ON llm_requests(created_at);
```

#### llm_responses
```sql
CREATE TABLE llm_responses (
    id VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL REFERENCES llm_requests(id),
    provider_id VARCHAR(255) NOT NULL REFERENCES llm_providers(id),
    provider_name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    confidence DECIMAL(5,4),
    tokens_used INTEGER DEFAULT 0,
    response_time BIGINT NOT NULL,
    finish_reason VARCHAR(100),
    metadata JSONB,
    selected BOOLEAN DEFAULT false,
    selection_score DECIMAL(5,4),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_llm_responses_request_id ON llm_responses(request_id);
CREATE INDEX idx_llm_responses_provider_id ON llm_responses(provider_id);
CREATE INDEX idx_llm_responses_selected ON llm_responses(selected);
```

#### user_sessions
```sql
CREATE TABLE user_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    session_token VARCHAR(500) NOT NULL UNIQUE,
    context JSONB,
    memory_id VARCHAR(255) REFERENCES cognee_memory(id),
    status VARCHAR(50) DEFAULT 'active',
    request_count INTEGER DEFAULT 0,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_status ON user_sessions(status);
```

#### cognee_memory
```sql
CREATE TABLE cognee_memory (
    id VARCHAR(255) PRIMARY KEY,
    session_id VARCHAR(255) REFERENCES user_sessions(id),
    dataset_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    vector_id VARCHAR(255),
    graph_nodes JSONB,
    search_key VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_cognee_memory_session_id ON cognee_memory(session_id);
CREATE INDEX idx_cognee_memory_dataset ON cognee_memory(dataset_name);
CREATE INDEX idx_cognee_memory_search_key ON cognee_memory(search_key);
```

## Validation Rules

### LLMProvider
- Weight must be between 0.1 and 5.0
- BaseURL must be valid HTTPS URL
- API key must be encrypted before storage
- Type must be one of: deepseek, claude, gemini, qwen, zai

### LLMRequest
- Prompt length <= 32,000 characters
- Temperature between 0.0 and 2.0
- MaxTokens between 1 and 32,000
- RequestType must be: code_generation, reasoning, tool_use

### LLMResponse
- Confidence score between 0.0 and 1.0
- Response time must be positive
- Only one response per request can be marked as selected
- Tokens used must be non-negative

### UserSession
- Session token must be hashed using SHA-256
- ExpiresAt must be in the future
- Request count must be non-negative

## State Transitions

### Request Status
```
pending → processing → completed
    ↓         ↓
  failed    failed
```

### Provider Health Status
```
unknown → healthy
    ↓
  degraded → failed
    ↓
  recovering → healthy
```

### Session Status
```
active → expired
    ↓
terminated
```

## Performance Considerations

### Indexing Strategy
- Composite indexes on frequently queried columns
- Partial indexes for active providers/sessions
- Time-based indexes for cleanup operations

### Partitioning
- Partition llm_requests by month for large datasets
- Partition llm_responses by request_id for efficient joins

### Caching
- Provider configuration cached in memory
- Active sessions cached with TTL
- Response results cached based on prompt hash