// Package handlers implements HTTP request handlers for HelixAgent's REST API.
//
// This package provides the HTTP layer for HelixAgent, implementing OpenAI-compatible
// endpoints and HelixAgent-specific functionality.
//
// # OpenAI-Compatible Endpoints
//
// The following endpoints are OpenAI API compatible:
//
//	POST /v1/completions       - Text completions
//	POST /v1/chat/completions  - Chat completions
//	POST /v1/embeddings        - Generate embeddings
//	GET  /v1/models            - List available models
//
// # HelixAgent-Specific Endpoints
//
// Additional endpoints for HelixAgent features:
//
//	POST /v1/debates           - Start AI debate session
//	GET  /v1/debates/:id       - Get debate status/result
//	POST /v1/tasks             - Create background task
//	GET  /v1/tasks/:id/status  - Get task status
//	GET  /v1/tasks/:id/events  - SSE stream for task events
//	GET  /v1/ws/tasks/:id      - WebSocket for task updates
//
// # Protocol Endpoints
//
// Protocol-specific handlers:
//
//	POST /v1/mcp              - Model Context Protocol
//	POST /v1/acp              - Agent Communication Protocol
//	POST /v1/lsp              - Language Server Protocol
//	POST /v1/cognee           - Cognee knowledge graph integration
//	POST /v1/vision           - Vision/image analysis
//
// # Handler Structure
//
// Each handler follows the pattern:
//
//	type CompletionHandler struct {
//	    service *services.EnsembleService
//	    logger  *logrus.Logger
//	}
//
//	func (h *CompletionHandler) HandleCompletion(c *gin.Context) {
//	    // Parse request
//	    // Validate input
//	    // Call service
//	    // Format response
//	}
//
// # Request/Response Models
//
// Models follow OpenAI conventions with extensions:
//
//	type CompletionRequest struct {
//	    Model       string   `json:"model"`
//	    Prompt      string   `json:"prompt"`
//	    MaxTokens   int      `json:"max_tokens"`
//	    Temperature float64  `json:"temperature"`
//	    Stream      bool     `json:"stream"`
//	    // HelixAgent extensions
//	    UseDebate   bool     `json:"use_debate,omitempty"`
//	}
//
// # Streaming Support
//
// Streaming responses use Server-Sent Events (SSE):
//
//	c.Writer.Header().Set("Content-Type", "text/event-stream")
//	c.Writer.Header().Set("Cache-Control", "no-cache")
//
//	for chunk := range responseStream {
//	    c.SSEvent("message", chunk)
//	    c.Writer.Flush()
//	}
//
// # Error Handling
//
// Errors follow OpenAI error format:
//
//	{
//	    "error": {
//	        "message": "Invalid API key",
//	        "type": "authentication_error",
//	        "code": "invalid_api_key"
//	    }
//	}
//
// # Key Files
//
//   - completion.go: Text completion endpoints
//   - chat.go: Chat completion endpoints
//   - openai_compatible.go: OpenAI API compatibility layer
//   - debate_handler.go: AI debate endpoints
//   - background_task_handler.go: Background task management
//   - protocol_sse.go: Server-Sent Events handling
//   - agent_handler.go: CLI agent endpoints
//   - cognee_handler.go: Cognee integration
//   - monitoring_handler.go: Health and metrics
//
// # Middleware Integration
//
// Handlers work with middleware from internal/middleware:
//
//   - Authentication (JWT, API key)
//   - Rate limiting
//   - CORS
//   - Request validation
//   - Logging
package handlers
