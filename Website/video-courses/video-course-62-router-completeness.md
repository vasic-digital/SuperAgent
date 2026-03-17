# Video Course 62: Router Completeness & Handler Architecture

## Course Overview

**Duration:** 2.5 hours
**Level:** Intermediate
**Prerequisites:** Course 01 (Fundamentals), Course 04 (API Reference), Course 10 (Handler Development)

Learn how HelixAgent's Gin router connects HTTP handlers to API endpoints, understand the 5 handler groups added in WS1 (BackgroundTask, Discovery, Scoring, Verification, Health), and master the patterns for registering new endpoints.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain HelixAgent's router architecture and handler registration pattern
2. Describe the 5 new endpoint groups and their purpose
3. Identify and remove dead handlers (unused code) from the codebase
4. Register new handler groups following established conventions
5. Write tests that validate route registration completeness
6. Understand the relationship between handlers, services, and middleware

---

## Module 1: Router Architecture Overview (30 min)

### Video 1.1: How Gin Routes Map to Handlers (15 min)

**Topics:**
- Gin's `RouterGroup` and path-based routing
- The `setupRouter()` function in `cmd/helixagent/main.go`
- Route groups: `/v1/`, `/v1/auth/`, `/v1/chat/`, etc.
- Middleware chain: auth, rate limiting, CORS, compression
- HTTP/3 (QUIC) transport layer with Brotli compression

**Key Code Pattern:**
```go
v1 := router.Group("/v1")
v1.Use(middleware.Auth(), middleware.RateLimit())
{
    v1.POST("/chat/completions", handlers.ChatCompletions)
    v1.GET("/models", handlers.ListModels)
    v1.GET("/models/:id", handlers.GetModel)
}
```

### Video 1.2: Handler Structure and Service Injection (15 min)

**Topics:**
- Handler struct pattern with service dependencies
- Constructor injection: `NewHandler(service, logger)`
- Request validation and error response formatting
- OpenAI-compatible response structure
- Handler lifecycle: registration, request handling, shutdown

**Pattern:**
```go
type BackgroundTaskHandler struct {
    taskService *services.BackgroundTaskService
    logger      *slog.Logger
}

func NewBackgroundTaskHandler(svc *services.BackgroundTaskService, l *slog.Logger) *BackgroundTaskHandler {
    return &BackgroundTaskHandler{taskService: svc, logger: l}
}

func (h *BackgroundTaskHandler) CreateTask(c *gin.Context) {
    var req CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    task, err := h.taskService.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, task)
}
```

---

## Module 2: The 5 New Endpoint Groups (45 min)

### Video 2.1: Background Tasks API /v1/tasks (15 min)

**Topics:**
- Task lifecycle: create, queue, execute, complete/fail
- Worker pool integration for async processing
- WebSocket support for real-time task updates
- Task analytics and queue statistics
- Webhook registration for external notifications

**Endpoints covered:**
- `POST /v1/tasks` -- Create a new task
- `GET /v1/tasks` -- List tasks with filtering
- `GET /v1/tasks/:id` -- Task details
- `POST /v1/tasks/:id/pause|resume|cancel` -- Task control
- `GET /v1/tasks/queue/stats` -- Queue analytics
- `GET /v1/tasks/:id/ws` -- WebSocket updates

### Video 2.2: Model Discovery API /v1/discovery (10 min)

**Topics:**
- 3-tier discovery: Provider API, models.dev, hardcoded fallback
- Discovery trigger and refresh cycle
- Cache management with 1-hour TTL
- Ensemble configuration endpoint

**Endpoints covered:**
- `GET /v1/discovery/models` -- All discovered models
- `GET /v1/discovery/models/selected` -- Debate-selected models
- `POST /v1/discovery/trigger` -- Manual discovery trigger
- `GET /v1/discovery/stats` -- Discovery statistics

### Video 2.3: Model Scoring API /v1/scoring (10 min)

**Topics:**
- 5-component weighted scoring (speed, cost, efficiency, capability, recency)
- Score calculation pipeline
- Weight configuration and cache invalidation
- Model comparison endpoint

**Endpoints covered:**
- `GET /v1/scoring/model/:name` -- Model score breakdown
- `POST /v1/scoring/batch` -- Batch scoring
- `GET /v1/scoring/top` -- Top-scored models
- `PUT /v1/scoring/weights` -- Update weights

### Video 2.4: Provider Verification API /v1/verification (5 min)

**Topics:**
- 8-test verification pipeline
- Single and batch verification
- Re-verification after outage recovery

### Video 2.5: Provider Health API /v1/health (5 min)

**Topics:**
- Real-time health monitoring
- Circuit breaker state exposure
- Latency percentile tracking
- Fastest provider selection

---

## Module 3: Dead Handler Removal (20 min)

### Video 3.1: Identifying Dead Handlers (10 min)

**Topics:**
- What constitutes a dead handler: registered but never called, or code exists but not registered
- Three dead handlers removed in WS1: CogneeHandler, GraphQLHandler, OpenRouterModelsHandler
- Why dead code violates the "No Dead Code" constitution rule
- Using grep and IDE tools to find unregistered handlers

**Detection Pattern:**
```bash
# Find handler structs
grep -r "type.*Handler struct" internal/handlers/

# Find route registrations
grep -r "v1\.\(GET\|POST\|PUT\|DELETE\)" cmd/helixagent/main.go

# Compare: any handler without a route is dead code
```

### Video 3.2: Safe Removal Process (10 min)

**Topics:**
- Verify no references exist (grep for all usages)
- Check tests -- dead handler tests should also be removed
- Commit with clear message: `refactor: remove dead CogneeHandler`
- Run full test suite to verify no breakage

---

## Module 4: Route Registration Testing (25 min)

### Video 4.1: The Router Completeness Challenge (15 min)

**Topics:**
- Why route completeness matters (handler exists but is not wired = dead code)
- The challenge script pattern: enumerate handlers, verify routes exist
- Testing route parameters, methods, and middleware
- The challenge validates all 5 new endpoint groups are registered

**Challenge Pattern:**
```bash
# Verify a handler is registered
check_route() {
    local method=$1
    local path=$2
    grep -q "${method}.*${path}" cmd/helixagent/main.go
    if [ $? -eq 0 ]; then
        echo "PASS: ${method} ${path} is registered"
    else
        echo "FAIL: ${method} ${path} is NOT registered"
    fi
}

check_route "POST" "/v1/tasks"
check_route "GET" "/v1/discovery/models"
check_route "GET" "/v1/scoring/model"
check_route "POST" "/v1/verification/model"
check_route "GET" "/v1/health/providers"
```

### Video 4.2: Integration Testing New Endpoints (10 min)

**Topics:**
- Writing integration tests for new routes
- Using `httptest.NewServer` with the full router
- Verifying response formats match OpenAI compatibility
- Testing error responses (400, 401, 404, 429, 500)

---

## Module 5: Hands-On Lab (30 min)

### Lab 1: Register a New Handler Group (15 min)

**Objective:** Create and register a new `/v1/custom` handler group with GET, POST, and DELETE endpoints.

**Steps:**
1. Create `internal/handlers/custom_handler.go` with handler struct
2. Implement 3 handler methods (List, Create, Delete)
3. Register routes in `setupRouter()`
4. Write unit tests for each handler
5. Verify registration with the router completeness pattern

### Lab 2: Detect and Remove a Dead Handler (15 min)

**Objective:** Given a codebase with an intentionally dead handler, identify and safely remove it.

**Steps:**
1. Grep for all handler structs
2. Cross-reference with route registrations
3. Verify no runtime references exist
4. Remove the handler and its tests
5. Run the full test suite

---

## Assessment

### Quiz (10 questions)

1. What function registers all API routes in HelixAgent?
2. Name the 5 handler groups added in WS1.
3. What are the 3 dead handlers that were removed?
4. Why should dead handlers be removed rather than left in place?
5. What is the middleware chain order for `/v1/` routes?
6. How does the Background Tasks handler support real-time updates?
7. What are the 3 tiers of model discovery?
8. What 5 components make up the scoring system?
9. How many tests does the verification pipeline run per model?
10. What circuit breaker states does the health API expose?

### Practical Assessment

Build a new `/v1/reports` handler group that:
- Lists reports (`GET /v1/reports`)
- Creates a report (`POST /v1/reports`)
- Gets a specific report (`GET /v1/reports/:id`)
- Deletes a report (`DELETE /v1/reports/:id`)

Requirements:
- Follow the handler struct pattern
- Use service injection
- Write table-driven tests
- Register in setupRouter()
- Verify with the completeness check

---

## Resources

- [API Reference Manual](../user-manuals/04-api-reference.md)
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [HelixAgent Handler Source](../../internal/handlers/)
- [Router Setup](../../cmd/helixagent/main.go)
- [Course 04: API Reference Deep Dive](video-course-04-api-reference.md)
