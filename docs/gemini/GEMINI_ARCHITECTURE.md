# Gemini Architecture

## Unified Provider Architecture

The Gemini integration follows a unified provider pattern where a single `GeminiUnifiedProvider` facade orchestrates three independent sub-providers. This is the only HelixAgent provider with three distinct access methods behind a single interface.

```
+================================================================+
|                    GeminiUnifiedProvider                         |
|  gemini.go                                                      |
|                                                                  |
|  Fields:                                                         |
|    model, timeout, maxTokens, apiKey                             |
|    preferredMethod: "auto" | "api" | "cli" | "acp"              |
|    apiProvider  *GeminiAPIProvider                                |
|    cliProvider  *GeminiCLIProvider                                |
|    acpProvider  *GeminiACPProvider                                |
|                                                                  |
|  LLMProvider Interface:                                          |
|    Complete(ctx, req)         -> routes to sub-provider           |
|    CompleteStream(ctx, req)   -> routes to sub-provider           |
|    HealthCheck()              -> checks best available            |
|    GetCapabilities()          -> merged capabilities              |
|    ValidateConfig()           -> checks any method available      |
|    GetName() = "gemini"                                          |
|    GetProviderType() = "gemini"                                  |
+==============+================+=================+================+
               |                |                 |
    +----------v----------+  +-v-----------+  +--v--------------+
    |  GeminiAPIProvider   |  | GeminiCLI   |  | GeminiACP       |
    |  gemini_api.go       |  | Provider    |  | Provider        |
    |                      |  | gemini_cli  |  | gemini_acp.go   |
    |  Transport: HTTP/REST|  | .go         |  |                 |
    |  Auth: x-goog-api-key|  |             |  | Transport:      |
    |  Format: JSON        |  | Transport:  |  |   JSON-RPC 2.0  |
    |                      |  |   exec.Cmd  |  |   stdin/stdout  |
    |  Features:           |  | Auth: API   |  |                 |
    |  - Streaming (SSE)   |  |   key or    |  | Features:       |
    |  - Function calling  |  |   OAuth     |  | - Sessions      |
    |  - Vision            |  | Format:     |  | - Tool support  |
    |  - Extended thinking |  |   JSON/JSONL|  | - Streaming*    |
    |  - Google Search     |  |             |  |   (*via poll)   |
    |  - Retry w/ backoff  |  | Features:   |  |                 |
    |  - Model discovery   |  | - Sessions  |  | Lifecycle:      |
    |                      |  |   (--resume)|  |   Start()       |
    |  Name: "gemini-api"  |  | - Streaming |  |   Stop()        |
    +----------------------+  | - Model     |  |                 |
                              |   discovery |  | Name:           |
                              |             |  |   "gemini-acp"  |
                              | Name:       |  +-----------------+
                              | "gemini-cli"|
                              +-------------+
```

## Request Flow Diagram

```
Client Request (LLMRequest)
    |
    v
GeminiUnifiedProvider.Complete(ctx, req)
    |
    v
Initialize() [sync.Once]
    |--- Creates CLI provider (GeminiCLIConfig)
    |--- Creates ACP provider (GeminiACPConfig)
    |--- API provider already created in constructor
    |
    v
Switch on preferredMethod
    |
    +-- "api"  --> completeWithAPI(ctx, req)
    |                  |
    |                  +-- apiProvider.Complete(ctx, req)
    |                       |
    |                       +-- convertRequest(req)
    |                       |     |-- Build GeminiContent from messages
    |                       |     |-- Set generation config (temp, topP, maxTokens)
    |                       |     |-- If thinkingModel: add ThinkingConfig
    |                       |     |-- Add function declarations from tools
    |                       |     +-- Always add GoogleSearch grounding
    |                       |
    |                       +-- makeAPICall(ctx, geminiReq)
    |                       |     |-- Marshal JSON body
    |                       |     |-- POST to /v1beta/models/{model}:generateContent
    |                       |     |-- Header: x-goog-api-key
    |                       |     +-- Retry loop (exp backoff + jitter)
    |                       |           |-- 429/5xx: retry after delay
    |                       |           +-- 401: single auth retry (500ms)
    |                       |
    |                       +-- Parse GeminiResponse
    |                       +-- convertResponse(req, resp, startTime)
    |                             |-- Extract text parts (skip thought=true)
    |                             |-- Extract thinking into metadata
    |                             |-- Extract function calls -> ToolCalls
    |                             +-- calculateConfidence(content, finishReason)
    |
    +-- "cli"  --> completeWithCLI(ctx, req)
    |                  |
    |                  +-- cliProvider.Complete(ctx, req)
    |                       |
    |                       +-- buildPromptFromMessages(req)
    |                       +-- exec.CommandContext(gemini, args...)
    |                       |     args: -p, --output-format json,
    |                       |           --approval-mode auto_edit,
    |                       |           -m model, [--resume sessionID]
    |                       +-- parseJSONResponse(stdout)
    |                       +-- Return LLMResponse with session metadata
    |
    +-- "acp"  --> completeWithACP(ctx, req)
    |                  |
    |                  +-- acpProvider.Complete(ctx, req)
    |                       |
    |                       +-- Start() [sync.Once]
    |                       |     |-- exec.Command(gemini, --experimental-acp)
    |                       |     |-- Setup stdin/stdout pipes
    |                       |     |-- Start readResponses goroutine
    |                       |     |-- sendRequest("initialize", ...)
    |                       |     +-- sendRequest("session/new", {cwd})
    |                       |
    |                       +-- Build prompt from messages
    |                       +-- sendRequest("session/prompt", {sessionId, prompt})
    |                       +-- Parse geminiPromptResponse
    |                       +-- Return LLMResponse with session metadata
    |
    +-- "auto" --> Try API -> CLI -> ACP (first success wins)
```

## Fallback Chain Diagram

```
                    preferredMethod = "auto"
                            |
                            v
            +-------------------------------+
            | apiProvider != nil?            |
            +-------+-----------+-----------+
                    |yes        |no
                    v           |
            +---------------+  |
            | API.Complete() |  |
            +-------+-------+  |
                    |           |
              ok? --+-- err     |
              |         |      |
              v         v      v
         [Return]  +-------------------------------+
                   | cliProvider.IsCLIAvailable()?  |
                   +-------+-----------+-----------+
                           |yes        |no
                           v           |
                   +---------------+   |
                   | CLI.Complete() |   |
                   +-------+-------+   |
                           |           |
                     ok? --+-- err     |
                     |         |      |
                     v         v      v
                [Return]  +-------------------------------+
                          | acpProvider.IsAvailable()?     |
                          +-------+-----------+-----------+
                                  |yes        |no
                                  v           |
                          +---------------+   |
                          | ACP.Complete() |   |
                          +-------+-------+   |
                                  |           |
                            ok? --+-- err     |
                            |         |      |
                            v         v      v
                       [Return]  [Error: all   [Error: no
                                  methods       method
                                  failed]       available]
```

## File Structure

```
internal/llm/providers/gemini/
    |
    +-- gemini.go           Unified provider (GeminiUnifiedProvider)
    |                         - Shared types (GeminiRequest, GeminiResponse, etc.)
    |                         - Retry configuration
    |                         - Prompt building helpers
    |                         - Model list (getAllGeminiModels)
    |                         - Availability checks
    |                         - Provider info for registry
    |
    +-- gemini_api.go       API sub-provider (GeminiAPIProvider)
    |                         - REST client with retry + jitter
    |                         - Extended request types (thinking, search)
    |                         - Streaming SSE parser
    |                         - Confidence calculation
    |                         - Model discovery integration
    |                         - Health check (list models endpoint)
    |
    +-- gemini_cli.go       CLI sub-provider (GeminiCLIProvider)
    |                         - CLI binary detection + auth check
    |                         - JSON output parsing
    |                         - JSONL stream parsing
    |                         - Session management (--resume)
    |                         - 3-tier model discovery (CLI help,
    |                           models.dev, hardcoded)
    |
    +-- gemini_acp.go       ACP sub-provider (GeminiACPProvider)
    |                         - JSON-RPC 2.0 protocol
    |                         - Process lifecycle (Start/Stop)
    |                         - Session management (session/new,
    |                           session/prompt)
    |                         - Response multiplexing by request ID
    |                         - Availability probing
    |
    +-- gemini_test.go      Comprehensive test suite
    |                         - Unified provider tests
    |                         - API provider tests (mock HTTP server)
    |                         - CLI provider basic tests
    |                         - ACP provider basic tests
    |                         - Benchmarks
    |
    +-- README.md           Package-level documentation
```

## Integration Points with HelixAgent Subsystems

```
+------------------+     +------------------+     +------------------+
|  Provider        |     |  LLMsVerifier    |     |  Model Discovery |
|  Registry        |     |  (startup)       |     |  (3-tier)        |
|                  |     |                  |     |                  |
|  Registers       |     |  8-test pipeline |     |  Tier 1: Gemini  |
|  "gemini" with   +---->+  scores Gemini   +---->+  /v1beta/models  |
|  unified provider|     |  models for      |     |  API endpoint    |
|                  |     |  debate team     |     |                  |
+--------+---------+     +------------------+     |  Tier 2:         |
         |                                        |  models.dev API  |
         v                                        |                  |
+------------------+     +------------------+     |  Tier 3:         |
|  Ensemble        |     |  Debate Service  |     |  Hardcoded list  |
|  Orchestration   |     |                  |     +------------------+
|                  |     |  Multi-round     |
|  Confidence-     |     |  debate with     |     +------------------+
|  weighted voting |     |  Gemini as       |     |  Circuit Breaker |
|  using Gemini    |     |  participant     |     |                  |
|  scores          |     |                  |     |  Fault tolerance  |
+------------------+     +--------+---------+     |  for API/CLI/ACP |
                                  |               |  failures         |
                                  v               +------------------+
                         +------------------+
                         |  Formatter       |     +------------------+
                         |  Integration     |     |  Background Task |
                         |                  |     |  Queue           |
                         |  Code formatting |     |                  |
                         |  of debate       |     |  Async Gemini    |
                         |  responses       |     |  requests via    |
                         +------------------+     |  worker pool     |
                                                  +------------------+
```

### Sub-Provider Capability Matrix

| Capability | API | CLI | ACP |
|------------|:---:|:---:|:---:|
| Non-streaming completion | Yes | Yes | Yes |
| Streaming completion | SSE | JSONL | Poll* |
| Function calling | Yes | No | No |
| Vision (multimodal) | Yes | No | No |
| Extended thinking | Yes | No | No |
| Google Search grounding | Yes | No | No |
| Session management | No | Yes | Yes |
| Model discovery | API endpoint | CLI help | N/A |
| Concurrent requests | 10 | 1 | 1 |
| Health check method | List models API | CLI available | Process running |

*ACP streaming delegates to `Complete()` and returns result on a channel.

### Initialization Sequence

```
NewGeminiUnifiedProvider(config)
    |
    +-- If apiKey != "": create GeminiAPIProvider (eager)
    |
    v
First Complete/CompleteStream/HealthCheck call
    |
    +-- Initialize() [sync.Once]
         |
         +-- Create GeminiCLIProvider
         |     +-- If model == "": GetBestAvailableModel()
         |           (triggers 3-tier model discovery)
         |
         +-- Create GeminiACPProvider
         |
         +-- Set initialized = true
```

The API provider is created eagerly (in the constructor) because it has no side effects. CLI and ACP providers are created lazily on first use to avoid unnecessary process spawning.
