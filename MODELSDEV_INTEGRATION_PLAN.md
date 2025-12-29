# Models.dev Integration - Comprehensive Implementation Plan

## Executive Summary
This document outlines the complete plan for integrating Models.dev into the SuperAgent project to dramatically enhance model and provider information capabilities. The integration will provide detailed, up-to-date model metadata, pricing, capabilities, and performance metrics.

## 1. Research & Analysis Phase

### 1.1 Models.dev API Analysis
- [x] Analyze Models.dev repository structure
- [x] Understand available API endpoints
- [x] Document data structures and response formats
- [x] Identify rate limits and usage guidelines
- [x] Document authentication requirements
- [x] Analyze available model metadata fields

### 1.2 OpenCode Integration Analysis
- [x] Study how OpenCode incorporates Models.dev
- [x] Document integration patterns
- [x] Identify reusable components
- [x] Analyze caching strategies
- [x] Document error handling approaches
- [x] Study configuration management

## 2. Architecture Design

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     SuperAgent API Layer                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Model Info   │  │ Provider     │  │ Capabilities │      │
│  │ Endpoints   │  │ Endpoints   │  │ Endpoints   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                 Model Metadata Service                       │
│  ┌──────────────────────────────────────────────────┐      │
│  │  - Fetch model data from Models.dev              │      │
│  │  - Cache model metadata                         │      │
│  │  - Periodic refresh mechanism                   │      │
│  │  - Query optimization                           │      │
│  └──────────────────────────────────────────────────┘      │
│         │                    │                             │
│         ▼                    ▼                             │
│  ┌──────────────┐    ┌──────────────┐                   │
│  │  Models.dev  │    │   PostgreSQL │                   │
│  │    Client    │    │   Database   │                   │
│  └──────────────┘    └──────────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Component Architecture

#### 2.2.1 Models.dev Client (`internal/modelsdev/client.go`)
- HTTP client for Models.dev API
- Request/response parsing
- Error handling and retries
- Rate limit management

#### 2.2.2 Model Metadata Service (`internal/services/model_metadata_service.go`)
- Orchestrate data fetching from Models.dev
- Cache management
- Refresh scheduling
- Query aggregation

#### 2.2.3 Database Layer (`internal/database/model_metadata_repository.go`)
- CRUD operations for model metadata
- Query optimization
- Transaction management

#### 2.2.4 API Handlers (`internal/handlers/model_metadata.go`)
- HTTP endpoints for model information
- Request validation
- Response formatting

### 2.3 Data Flow

**Refresh Flow:**
```
Scheduler → ModelMetadataService → Models.dev Client → Models.dev API
                                              ↓
                                       Parse Response
                                              ↓
                                       Store in Database
                                              ↓
                                       Update Cache
```

**Query Flow:**
```
API Request → Handler → ModelMetadataService → Cache (hit?) → Response
                                              ↓ (miss)
                                       Database Query → Response
```

## 3. Database Schema Design

### 3.1 New Tables

#### models_metadata
```sql
CREATE TABLE models_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id VARCHAR(255) UNIQUE NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,

    -- Model details
    description TEXT,
    context_window INTEGER,
    max_tokens INTEGER,
    pricing_input DECIMAL(10, 6),
    pricing_output DECIMAL(10, 6),
    pricing_currency VARCHAR(10) DEFAULT 'USD',

    -- Capabilities
    supports_vision BOOLEAN DEFAULT FALSE,
    supports_function_calling BOOLEAN DEFAULT FALSE,
    supports_streaming BOOLEAN DEFAULT FALSE,
    supports_json_mode BOOLEAN DEFAULT FALSE,
    supports_image_generation BOOLEAN DEFAULT FALSE,
    supports_audio BOOLEAN DEFAULT FALSE,
    supports_code_generation BOOLEAN DEFAULT FALSE,
    supports_reasoning BOOLEAN DEFAULT FALSE,

    -- Performance metrics
    benchmark_score DECIMAL(5, 2),
    popularity_score INTEGER,
    reliability_score DECIMAL(5, 2),

    -- Categories and tags
    model_type VARCHAR(100),
    model_family VARCHAR(100),
    version VARCHAR(50),
    tags JSONB,

    -- Models.dev specific
    modelsdev_url TEXT,
    modelsdev_id VARCHAR(255),
    modelsdev_api_version VARCHAR(50),

    -- Metadata
    raw_metadata JSONB,
    last_refreshed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_provider FOREIGN KEY (provider_id) REFERENCES llm_providers(id) ON DELETE CASCADE
);

CREATE INDEX idx_models_metadata_provider_id ON models_metadata(provider_id);
CREATE INDEX idx_models_metadata_model_type ON models_metadata(model_type);
CREATE INDEX idx_models_metadata_tags ON models_metadata USING GIN(tags);
CREATE INDEX idx_models_metadata_last_refreshed ON models_metadata(last_refreshed_at);
```

#### model_benchmarks
```sql
CREATE TABLE model_benchmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id VARCHAR(255) NOT NULL,
    benchmark_name VARCHAR(255) NOT NULL,
    benchmark_type VARCHAR(100),
    score DECIMAL(10, 4),
    rank INTEGER,
    normalized_score DECIMAL(5, 2),
    benchmark_date DATE,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_model FOREIGN KEY (model_id) REFERENCES models_metadata(model_id) ON DELETE CASCADE,
    CONSTRAINT unique_model_benchmark UNIQUE (model_id, benchmark_name)
);

CREATE INDEX idx_benchmarks_model_id ON model_benchmarks(model_id);
CREATE INDEX idx_benchmarks_type ON model_benchmarks(benchmark_type);
```

#### models_refresh_history
```sql
CREATE TABLE models_refresh_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refresh_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    models_refreshed INTEGER,
    models_failed INTEGER,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INTEGER,
    metadata JSONB
);

CREATE INDEX idx_refresh_history_started ON models_refresh_history(started_at);
CREATE INDEX idx_refresh_history_status ON models_refresh_history(status);
```

### 3.2 Enhanced Existing Tables

#### llm_providers (enhanced)
```sql
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS modelsdev_provider_id VARCHAR(255);
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS total_models INTEGER DEFAULT 0;
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS enabled_models INTEGER DEFAULT 0;
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS last_models_sync TIMESTAMP;
```

## 4. Implementation Plan

### 4.1 Phase 1: Core Infrastructure (Days 1-3)

#### 4.1.1 Models.dev Client
- [ ] `internal/modelsdev/client.go` - Main client implementation
- [ ] `internal/modelsdev/types.go` - Data structures
- [ ] `internal/modelsdev/errors.go` - Error types
- [ ] Tests for Models.dev client
  - `internal/modelsdev/client_test.go` - Unit tests
  - `internal/modelsdev/integration_test.go` - Integration tests

#### 4.1.2 Database Layer
- [ ] `internal/database/model_metadata_repository.go` - Repository implementation
- [ ] Migration scripts in `scripts/migrations/`
- [ ] Tests for database layer
  - `internal/database/model_metadata_repository_test.go` - Unit tests
  - `internal/database/migration_test.go` - Migration tests

#### 4.1.3 Model Metadata Service
- [ ] `internal/services/model_metadata_service.go` - Service implementation
- [ ] `internal/services/refresh_scheduler.go` - Background refresh scheduler
- [ ] Tests for service layer
  - `internal/services/model_metadata_service_test.go` - Unit tests
  - `internal/services/refresh_scheduler_test.go` - Unit tests

### 4.2 Phase 2: API Layer (Days 4-5)

#### 4.2.1 Handlers
- [ ] `internal/handlers/model_metadata.go` - API handlers
- [ ] Tests for handlers
  - `internal/handlers/model_metadata_test.go` - Unit tests

#### 4.2.2 Routes
- [ ] Update `internal/router/router.go` with new routes
- [ ] Tests for routes
  - `internal/router/model_metadata_routes_test.go` - Integration tests

### 4.3 Phase 3: Integration (Days 6-7)

#### 4.3.1 Provider Registry Integration
- [ ] Update `internal/services/provider_registry.go`
- [ ] Update `internal/llm/provider.go` interface
- [ ] Tests for integration
  - `internal/services/provider_registry_modeldev_test.go` - Integration tests

#### 4.3.2 Configuration
- [ ] Update `internal/config/config.go`
- [ ] Add Models.dev configuration options
- [ ] Tests for configuration
  - `internal/config/modeldev_config_test.go` - Unit tests

### 4.4 Phase 4: Comprehensive Testing (Days 8-10)

#### 4.4.1 Test Coverage
- [ ] Unit tests - Target 100% coverage
- [ ] Integration tests - Target 100% coverage
- [ ] E2E tests - Target 100% coverage
- [ ] Security tests
- [ ] Stress tests
- [ ] Chaos tests

#### 4.4.2 Test Execution
- [ ] Run all test types sequentially
- [ ] Verify 100% pass rate
- [ ] Generate coverage reports
- [ ] Fix any failing tests

### 4.5 Phase 5: Documentation (Days 11-12)

#### 4.5.1 Technical Documentation
- [ ] Architecture documentation
- [ ] API documentation
- [ ] Database schema documentation
- [ ] Configuration guide

#### 4.5.2 User Documentation
- [ ] Setup guide
- [ ] Usage examples
- [ ] Troubleshooting guide

#### 4.5.3 Developer Documentation
- [ ] Update AGENTS.md
- [ ] Code examples
- [ ] Contributing guidelines

## 5. Data Structures

### 5.1 Models.dev Client Types

```go
// ModelInfo represents comprehensive model information from Models.dev
type ModelInfo struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Provider     string                 `json:"provider"`
    DisplayName  string                 `json:"display_name"`
    Description  string                 `json:"description"`
    ContextWindow int                   `json:"context_window"`
    MaxTokens    int                    `json:"max_tokens"`
    Pricing      *ModelPricing          `json:"pricing"`
    Capabilities ModelCapabilities       `json:"capabilities"`
    Performance  *ModelPerformance      `json:"performance"`
    Tags        []string               `json:"tags"`
    Categories   []string               `json:"categories"`
    Family       string                 `json:"family"`
    Version      string                 `json:"version"`
    Metadata     map[string]interface{} `json:"metadata"`
}

type ModelPricing struct {
    InputPrice  float64 `json:"input_price"`
    OutputPrice float64 `json:"output_price"`
    Currency    string  `json:"currency"`
    Unit        string  `json:"unit"` // "tokens", "characters", etc.
}

type ModelCapabilities struct {
    Vision           bool `json:"vision"`
    FunctionCalling  bool `json:"function_calling"`
    Streaming        bool `json:"streaming"`
    JSONMode         bool `json:"json_mode"`
    ImageGeneration  bool `json:"image_generation"`
    Audio            bool `json:"audio"`
    CodeGeneration   bool `json:"code_generation"`
    Reasoning        bool `json:"reasoning"`
    ToolUse          bool `json:"tool_use"`
}

type ModelPerformance struct {
    BenchmarkScore  float64            `json:"benchmark_score"`
    PopularityScore int                `json:"popularity_score"`
    ReliabilityScore float64           `json:"reliability_score"`
    Benchmarks       map[string]float64 `json:"benchmarks"`
}

// ModelsListResponse represents API response for listing models
type ModelsListResponse struct {
    Models []ModelInfo `json:"models"`
    Total  int          `json:"total"`
    Page   int          `json:"page"`
    Limit  int          `json:"limit"`
}

// ProviderInfo represents provider information
type ProviderInfo struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    DisplayName string   `json:"display_name"`
    Description string   `json:"description"`
    ModelsCount int      `json:"models_count"`
    Website     string   `json:"website"`
    APIDocsURL string   `json:"api_docs_url"`
    Features   []string `json:"features"`
}
```

### 5.2 Internal Types

```go
// ModelMetadata represents cached model information
type ModelMetadata struct {
    ID                string                 `json:"id"`
    ModelID           string                 `json:"model_id"`
    ModelName         string                 `json:"model_name"`
    ProviderID        string                 `json:"provider_id"`
    ProviderName      string                 `json:"provider_name"`
    Description       string                 `json:"description"`
    ContextWindow     int                    `json:"context_window"`
    MaxTokens         int                    `json:"max_tokens"`
    PricingInput      float64                `json:"pricing_input"`
    PricingOutput     float64                `json:"pricing_output"`
    PricingCurrency   string                 `json:"pricing_currency"`
    SupportsVision    bool                   `json:"supports_vision"`
    SupportsFunctionCalling bool             `json:"supports_function_calling"`
    SupportsStreaming bool                   `json:"supports_streaming"`
    SupportsJSONMode  bool                   `json:"supports_json_mode"`
    BenchmarkScore    float64                `json:"benchmark_score"`
    PopularityScore   int                    `json:"popularity_score"`
    ReliabilityScore  float64                `json:"reliability_score"`
    ModelType         string                 `json:"model_type"`
    ModelFamily       string                 `json:"model_family"`
    Version           string                 `json:"version"`
    Tags              []string               `json:"tags"`
    ModelsDevURL      string                 `json:"modelsdev_url"`
    ModelsDevID       string                 `json:"modelsdev_id"`
    RawMetadata       map[string]interface{} `json:"raw_metadata"`
    LastRefreshedAt   time.Time              `json:"last_refreshed_at"`
    CreatedAt         time.Time              `json:"created_at"`
    UpdatedAt         time.Time              `json:"updated_at"`
}

// ModelBenchmark represents benchmark results
type ModelBenchmark struct {
    ID              string                 `json:"id"`
    ModelID         string                 `json:"model_id"`
    BenchmarkName   string                 `json:"benchmark_name"`
    BenchmarkType   string                 `json:"benchmark_type"`
    Score           float64                `json:"score"`
    Rank            int                    `json:"rank"`
    NormalizedScore float64                `json:"normalized_score"`
    BenchmarkDate   time.Time              `json:"benchmark_date"`
    Metadata        map[string]interface{} `json:"metadata"`
    CreatedAt       time.Time              `json:"created_at"`
}
```

## 6. API Endpoints

### 6.1 Model Information Endpoints

```
GET    /api/v1/models                    List all models
GET    /api/v1/models/:id               Get model by ID
GET    /api/v1/providers/:id/models     List models by provider
GET    /api/v1/models/search            Search models
GET    /api/v1/models/:id/benchmarks    Get model benchmarks
GET    /api/v1/models/compare           Compare multiple models
GET    /api/v1/providers                List all providers
POST   /api/v1/models/refresh          Trigger manual refresh
GET    /api/v1/models/refresh/status    Get refresh status
```

### 6.2 Request/Response Examples

#### List Models
```
GET /api/v1/models?page=1&limit=20&provider=anthropic

Response:
{
  "models": [...],
  "total": 150,
  "page": 1,
  "limit": 20
}
```

#### Get Model Details
```
GET /api/v1/models/claude-3-sonnet-20240229

Response:
{
  "id": "claude-3-sonnet-20240229",
  "name": "Claude 3 Sonnet",
  "provider": "anthropic",
  "description": "...",
  "context_window": 200000,
  "max_tokens": 4096,
  "pricing": {
    "input_price": 0.000003,
    "output_price": 0.000015,
    "currency": "USD",
    "unit": "tokens"
  },
  "capabilities": {
    "vision": true,
    "function_calling": true,
    ...
  },
  "performance": {
    "benchmark_score": 95.5,
    ...
  }
}
```

## 7. Caching Strategy

### 7.1 Multi-Layer Cache

```
Level 1: In-Memory Cache (Redis)
  - TTL: 1 hour for hot data
  - Size: 1000 models

Level 2: Database Cache (PostgreSQL)
  - Persistent storage
  - Query optimized
  - Indexed properly

Level 3: Models.dev API
  - Source of truth
  - Refreshed periodically
  - Rate limited
```

### 7.2 Cache Invalidation

- Time-based: Every 24 hours
- Manual: Admin-triggered
- Event-based: On provider configuration changes

## 8. Refresh Mechanism

### 8.1 Scheduled Refresh

- Frequency: Daily at 2 AM UTC
- Strategy: Incremental updates
- Fallback: Full refresh on failure
- Retries: 3 attempts with exponential backoff

### 8.2 Manual Refresh

- Admin endpoint: `POST /api/v1/models/refresh`
- Options: Full refresh or specific provider
- Async execution with status tracking

### 8.3 Refresh History

- Track all refresh attempts
- Store success/failure metrics
- Enable audit and debugging

## 9. Error Handling

### 9.1 Error Types

```go
type ModelDevError struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    Code    int    `json:"code"`
    Details string `json:"details,omitempty"`
}
```

### 9.2 Error Categories

- **Network Errors**: Connection timeouts, DNS failures
- **API Errors**: Rate limits, invalid responses
- **Data Errors**: Parse errors, invalid data
- **Cache Errors**: Cache misses, write failures
- **Database Errors**: Query failures, connection issues

### 9.3 Error Handling Strategy

- Log all errors with context
- Retry transient failures
- Fall back to cached data
- Alert on critical failures
- Provide user-friendly messages

## 10. Security Considerations

### 10.1 API Key Management
- Store in environment variables
- Never log or expose in responses
- Rotate regularly

### 10.2 Rate Limiting
- Implement client-side rate limiting
- Respect Models.dev rate limits
- Cache responses to reduce API calls

### 10.3 Input Validation
- Validate all user inputs
- Sanitize query parameters
- Prevent SQL injection

### 10.4 Access Control
- Admin-only refresh endpoints
- Rate limit public endpoints
- Audit all refresh operations

## 11. Performance Optimization

### 11.1 Database Optimization
- Proper indexing
- Query optimization
- Connection pooling
- Read replicas

### 11.2 Caching Optimization
- Redis for hot data
- Cache warming
- Efficient cache keys
- Cache size management

### 11.3 API Optimization
- Pagination
- Field selection
- Batch operations
- Response compression

## 12. Monitoring & Observability

### 12.1 Metrics
- Cache hit/miss ratio
- API response times
- Refresh success/failure rates
- Database query performance

### 12.2 Logging
- Structured logging
- Request/response logging (debug mode)
- Error logging with stack traces
- Refresh operation logs

### 12.3 Alerts
- High error rates
- Cache failures
- Refresh failures
- API rate limit breaches

## 13. Testing Strategy

### 13.1 Unit Tests (100% coverage)
- Test individual components in isolation
- Mock external dependencies
- Cover all code paths
- Use table-driven tests

### 13.2 Integration Tests (100% coverage)
- Test component interactions
- Use test database
- Mock Models.dev API
- Test database operations

### 13.3 E2E Tests (100% coverage)
- Test complete user flows
- Deploy test environment
- Use real database
- Mock or use staging Models.dev API

### 13.4 Security Tests
- SQL injection attempts
- XSS attacks
- Rate limit enforcement
- Authentication bypasses

### 13.5 Stress Tests
- High concurrent requests
- Large dataset handling
- Memory leak detection
- Performance under load

### 13.6 Chaos Tests
- Network failures
- Database failures
- Cache failures
- API failures

## 14. Rollout Plan

### 14.1 Phase 1: Development
- Implement all components
- Write comprehensive tests
- Code review and testing

### 14.2 Phase 2: Staging
- Deploy to staging
- Run E2E tests
- Performance testing
- Security audit

### 14.3 Phase 3: Production
- Feature flag deployment
- Gradual rollout (10% → 50% → 100%)
- Monitor metrics closely
- Be ready to rollback

### 14.4 Phase 4: Post-Deployment
- Monitor for 48 hours
- Collect feedback
- Optimize based on usage
- Update documentation

## 15. Success Criteria

### 15.1 Functional Criteria
- [ ] Successfully fetch and store model metadata
- [ ] API endpoints return correct data
- [ ] Refresh mechanism works reliably
- [ ] Cache performance meets SLA
- [ ] All tests pass with 100% coverage

### 15.2 Performance Criteria
- [ ] API response time < 100ms (p95)
- [ ] Cache hit ratio > 80%
- [ ] Database query time < 50ms (p95)
- [ ] Support 1000 concurrent requests

### 15.3 Reliability Criteria
- [ ] 99.9% uptime for API endpoints
- [ ] 99% successful refresh rate
- [ ] Zero data loss
- [ ] Graceful degradation on failures

## 16. Future Enhancements

### 16.1 Advanced Features
- Model recommendation engine
- Cost optimization suggestions
- Performance analytics
- Custom benchmarking

### 16.2 Integrations
- Provider-specific optimizations
- Model fine-tuning tracking
- A/B testing framework
- Model version comparison

### 16.3 User Features
- User preference tracking
- Usage analytics
- Custom model tags
- Saved model comparisons

---

**Document Version:** 1.0
**Last Updated:** 2025-12-29
**Status:** Draft - Ready for Review
