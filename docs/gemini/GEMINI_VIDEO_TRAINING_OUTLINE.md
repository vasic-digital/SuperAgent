# Gemini Video Training Outline

Six-episode training series covering the complete Gemini integration in HelixAgent. Total runtime: approximately 75 minutes.

---

## Episode 1: Setup and Configuration (10 min)

### Objectives
- Obtain and configure a Gemini API key
- Install the Gemini CLI for headless and ACP modes
- Verify all three access methods are operational

### Key Concepts
- Unified provider architecture: one registration, three access methods
- `GEMINI_API_KEY` environment variable
- `GEMINI_PREFERRED_METHOD` selection (`auto`, `api`, `cli`, `acp`)
- CLI authentication via `~/.gemini/` credential files

### Demo Steps
1. Navigate to `aistudio.google.com` and generate an API key
2. Add `GEMINI_API_KEY` to the project `.env` file
3. Install the `gemini` CLI binary and run `gemini auth login`
4. Verify API mode: `curl` the Gemini models endpoint with `x-goog-api-key` header
5. Verify CLI mode: `gemini -p "Hello" --output-format json`
6. Verify ACP mode: inspect `gemini --experimental-acp` handshake
7. Start HelixAgent and observe Gemini appearing in startup verification logs

### Key Takeaways
- API mode is the primary and fastest method
- CLI and ACP serve as fallbacks when no API key is available
- The unified provider tries methods in order: API, CLI, ACP
- Authentication credentials in `~/.gemini/` enable CLI/ACP without an API key

---

## Episode 2: API Integration Deep Dive (15 min)

### Objectives
- Understand the direct REST API implementation
- Trace a request through conversion, execution, and response parsing
- Configure retry behavior and understand error handling

### Key Concepts
- `GeminiAPIProvider` structure and lifecycle
- Request conversion: `LLMRequest` to `GeminiAPIRequest`
- Safety settings (all categories set to `BLOCK_NONE`)
- Retry with exponential backoff and jitter
- Auth-retry for transient 401 errors
- Confidence scoring algorithm
- 3-tier model discovery (API endpoint, models.dev, hardcoded fallback)

### Demo Steps
1. Open `gemini_api.go` and walk through `NewGeminiAPIProvider`
2. Trace `convertRequest()`: message mapping, generation config, tools, Google Search
3. Show `makeAPICall()` retry loop with status code handling
4. Demonstrate a successful API call using the test HTTP server pattern from `gemini_test.go`
5. Trigger a 429 rate limit and observe retry behavior
6. Show confidence calculation: base 0.85, adjustments for finish reason and length
7. Demonstrate model discovery via the `/v1beta/models` endpoint

### Key Takeaways
- All requests automatically include Google Search grounding
- Extended thinking is auto-enabled for supported models (2.5-pro, 3-pro, 3.1-pro)
- Retry handles 429, 5xx with exponential backoff; 401 with a single fast retry
- Model discovery falls back gracefully through three tiers

---

## Episode 3: CLI and ACP Modes (15 min)

### Objectives
- Understand when and why CLI/ACP modes are used
- Implement session continuity across multiple requests
- Stream responses via CLI JSONL output

### Key Concepts
- `GeminiCLIProvider`: exec-based, headless mode
- CLI flags: `-p`, `--output-format`, `--approval-mode`, `-m`, `--resume`
- JSON vs stream-json output formats
- Session ID tracking and resumption
- `GeminiACPProvider`: JSON-RPC 2.0 over stdin/stdout
- ACP lifecycle: `Start()`, `initialize`, `session/new`, `session/prompt`, `Stop()`
- Response multiplexing by request ID

### Demo Steps
1. Open `gemini_cli.go` and trace the `Complete()` flow
2. Show `IsCLIAvailable()` check sequence: PATH lookup, version, auth
3. Demonstrate session resumption: first call gets sessionID, second call uses `--resume`
4. Open `gemini_acp.go` and trace the `startProcess()` initialization
5. Walk through `sendRequest()` and the `readResponses()` goroutine
6. Show the JSON-RPC handshake: `initialize` then `session/new`
7. Demonstrate ACP availability probing in `testGeminiACPAvailability()`

### Key Takeaways
- CLI mode works with either API key or OAuth credentials
- Session management enables multi-turn conversations without API state
- ACP provides richer integration but requires `--experimental-acp` support
- Both CLI and ACP are limited to single concurrent request (vs 10 for API)

---

## Episode 4: Extended Thinking and Search Grounding (10 min)

### Objectives
- Understand which models support extended thinking
- Configure and interpret thinking budget
- Observe Google Search grounding in action

### Key Concepts
- `thinkingModels` map: `gemini-2.5-pro`, `gemini-3-pro-preview`, `gemini-3.1-pro-preview`
- `thinkingBudgetDefault` = 8,192 tokens
- `GeminiThinkingConfig` in generation config
- Thinking parts: `GeminiPart.Thought = true`
- Thinking content stored in `LLMResponse.Metadata["thinking"]`
- `GeminiGoogleSearch` tool definition (always-on)
- Extended token models: `gemini-2.5*` and `gemini-3*` get 65,536 max output

### Demo Steps
1. Show the `thinkingModels` map and `extendedTokenPrefixes` list in `gemini_api.go`
2. Walk through `convertRequest()` thinking config injection
3. Show `convertResponse()` separating thought parts from content parts
4. Send a complex reasoning prompt to `gemini-2.5-pro` and inspect the `thinking` metadata
5. Demonstrate Google Search grounding: send a factual query and observe grounded response
6. Show the `resolveMaxTokens()` function and model-based token cap selection

### Key Takeaways
- Extended thinking is transparent: enabled automatically, thinking content separated into metadata
- The thinking budget (8,192 tokens) is the model's internal reasoning allowance
- Google Search is always on -- no configuration needed
- Token limits are model-aware: 65,536 for 2.5+ models, 8,192 for legacy

---

## Episode 5: Debate System Integration (15 min)

### Objectives
- Trace how Gemini participates in multi-LLM ensemble debates
- Understand scoring, selection, and fallback within debates
- Configure Gemini as a debate team member

### Key Concepts
- Provider registry registration (`"gemini"`)
- LLMsVerifier 8-test pipeline and 5-component scoring
- Debate team dynamic selection based on verification scores
- Confidence-weighted voting with Gemini's scoring algorithm
- Session continuity benefits for multi-round debates
- Fallback chain within debate context

### Demo Steps
1. Show Gemini registration in `internal/services/provider_registry.go`
2. Walk through HelixAgent startup: provider discovery, verification, scoring
3. Observe Gemini's verification score in startup logs
4. Trigger a multi-LLM debate with Gemini as a participant
5. Show how Gemini's confidence score (from `calculateConfidence`) feeds into voting
6. Demonstrate fallback: disable API key, observe CLI fallback during debate
7. Show debate performance optimizer handling Gemini alongside other providers

### Key Takeaways
- Gemini is scored like any other provider through LLMsVerifier
- The unified provider's fallback chain works transparently within debates
- Extended thinking metadata can inform debate evaluation
- Session-based CLI/ACP modes maintain context across debate rounds

---

## Episode 6: Monitoring and Troubleshooting (10 min)

### Objectives
- Monitor Gemini provider health and performance
- Diagnose common failure modes
- Use debugging tools to resolve issues

### Key Concepts
- Health check hierarchy: API list-models, CLI availability, ACP process status
- Circuit breaker integration for Gemini failures
- Available access methods inspection
- Error categories: auth, rate limit, safety, timeout, network
- Retry behavior and when retries are exhausted

### Demo Steps
1. Call `provider.HealthCheck()` and trace the check hierarchy
2. Call `provider.GetAvailableAccessMethods()` to see active methods
3. Simulate an API key revocation and observe health check failure + CLI fallback
4. Show circuit breaker state after repeated Gemini failures
5. Force `GEMINI_PREFERRED_METHOD=api` and observe no-fallback behavior
6. Demonstrate `GetGeminiProviderInfo()` output for diagnostics
7. Walk through the troubleshooting tables in the integration guide

### Key Takeaways
- Health checks use the lightest available method (API models list preferred)
- `GetAvailableAccessMethods()` is the primary diagnostic tool
- Forcing a specific method disables fallback -- useful for debugging
- Safety-filtered responses reduce confidence but do not cause errors
