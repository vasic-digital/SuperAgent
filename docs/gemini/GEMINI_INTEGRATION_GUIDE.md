# Gemini Integration Guide

## Overview

The Gemini provider in HelixAgent implements a **unified provider architecture** that orchestrates three distinct access methods through a single `GeminiUnifiedProvider` entry point. This design enables automatic fallback between methods, transparent session management, and consistent behavior regardless of which access path is active.

The unified provider is registered in the HelixAgent provider registry as `gemini` and participates in ensemble debates, LLMsVerifier scoring, and dynamic provider selection like any other provider.

### Three Access Methods

```
                     +---------------------------+
                     |  GeminiUnifiedProvider     |
                     |  (gemini.go)               |
                     |  preferredMethod: "auto"   |
                     +--+-------+-------+---------+
                        |       |       |
               +--------+  +---+---+  ++--------+
               |           |       |             |
  +------------v--+  +----v-----+ +---v---------+
  | GeminiAPI     |  | GeminiCLI| | GeminiACP   |
  | Provider      |  | Provider | | Provider    |
  | (gemini_api)  |  | (gemini  | | (gemini_acp)|
  |               |  |  _cli)   | |             |
  | REST API      |  | Headless | | JSON-RPC 2.0|
  | x-goog-api-key|  | -p flag  | | stdin/stdout|
  +---------------+  +----------+ +-------------+
```

1. **API** (primary) -- Direct REST calls to `generativelanguage.googleapis.com/v1beta`. Fastest, supports all features including streaming, function calling, vision, extended thinking, and Google Search grounding. Requires `GEMINI_API_KEY`.

2. **CLI Headless** (fallback 1) -- Invokes the locally installed `gemini` CLI binary with `-p <prompt> --output-format json`. Supports session resumption via `--resume`, streaming via `--output-format stream-json`, and auto-approval via `--approval-mode auto_edit`. Works with either `GEMINI_API_KEY` or OAuth credentials in `~/.gemini/`.

3. **ACP** (fallback 2) -- Starts `gemini --experimental-acp` as a long-running subprocess and communicates via JSON-RPC 2.0 over stdin/stdout. Provides session management, tool support, and IDE-like integration. Protocol version 1.

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMINI_API_KEY` | No* | (none) | Google AI API key from `aistudio.google.com`. Required for API mode. |
| `GEMINI_PREFERRED_METHOD` | No | `auto` | Access method selection: `auto`, `api`, `cli`, `acp`. |

*At least one access method must be available: API key for API mode, or installed+authenticated `gemini` CLI for CLI/ACP modes.

### Method Selection

When `GEMINI_PREFERRED_METHOD` is set to `auto` (the default), the unified provider tries methods in order:

1. **API** -- if `GEMINI_API_KEY` is set
2. **CLI** -- if `gemini` binary is in PATH and authenticated
3. **ACP** -- if `gemini` binary supports `--experimental-acp`

If a specific method is set (`api`, `cli`, `acp`), only that method is used with no fallback.

### Programmatic Configuration

```go
config := gemini.GeminiUnifiedConfig{
    Model:           "gemini-2.5-flash",
    Timeout:         180 * time.Second,
    MaxTokens:       8192,
    APIKey:          os.Getenv("GEMINI_API_KEY"),
    PreferredMethod: "auto",
}
provider := gemini.NewGeminiUnifiedProvider(config)
```

Backward-compatible constructors are also available:

```go
// Simple constructor (API key, base URL, model)
provider := gemini.NewGeminiProvider(apiKey, "", "gemini-2.5-pro")

// With custom retry configuration
provider := gemini.NewGeminiProviderWithRetry(apiKey, "", model, retryConfig)
```

## Model Inventory

| Model | Extended Thinking | Max Output Tokens | Context Window | Best For |
|-------|:-:|:-:|:-:|----------|
| `gemini-3.1-pro-preview` | Yes | 65,536 | 1M | Latest preview, highest capability |
| `gemini-3-pro-preview` | Yes | 65,536 | 1M | Next-gen pro model preview |
| `gemini-3-flash-preview` | No | 65,536 | 1M | Next-gen fast model preview |
| `gemini-2.5-pro` | Yes | 65,536 | 1M | Production reasoning tasks |
| `gemini-2.5-flash` | No | 65,536 | 1M | Default model, balanced speed/quality |
| `gemini-2.5-flash-lite` | No | 65,536 | 1M | Lightweight, cost-effective |
| `gemini-2.0-flash` | No | 8,192 | 1M | Legacy fast model |
| `gemini-embedding-001` | N/A | N/A | N/A | Text embedding only |

The default model is `gemini-2.5-flash` (`GeminiDefaultModel` constant).

## Extended Thinking

Extended thinking is automatically enabled for models that support it: `gemini-2.5-pro`, `gemini-3-pro-preview`, and `gemini-3.1-pro-preview`. These are tracked in the `thinkingModels` map in `gemini_api.go`.

When enabled, the API request includes a `thinkingConfig` in the generation config:

```json
{
  "generationConfig": {
    "thinkingConfig": {
      "thinkingBudget": 8192
    }
  }
}
```

The default thinking budget is 8,192 tokens (`thinkingBudgetDefault`). The model uses this budget for internal chain-of-thought reasoning before producing its final answer.

### Thinking Content in Responses

The Gemini API returns thinking content as response parts with `"thought": true`. The provider separates these from the main content:

- **Main content** goes into `LLMResponse.Content`
- **Thinking content** goes into `LLMResponse.Metadata["thinking"]`

This allows debate participants and downstream consumers to inspect the model's reasoning process without it contaminating the final output.

## Google Search Grounding

Google Search grounding is **always enabled** for API requests. Every request includes a `googleSearch` tool definition alongside any user-specified function declarations:

```json
{
  "tools": [
    { "functionDeclarations": [...] },
    { "googleSearch": {} }
  ]
}
```

This allows the model to ground its responses in real-time web search results when relevant. There is no configuration to disable it -- it is an always-on feature that improves factual accuracy.

## Session Management

### CLI Sessions

The CLI provider maintains session state via the `--resume` flag. After the first request, the provider stores the returned `sessionID` and passes it on subsequent requests, enabling multi-turn conversations:

```
gemini -p "first message" --output-format json
# returns sessionID: "abc123"

gemini -p "follow up" --output-format json --resume abc123
# continues the conversation
```

### ACP Sessions

The ACP provider creates a persistent session on startup via the `session/new` JSON-RPC method. All subsequent prompts reference the same `sessionId`, providing true stateful conversation management. The ACP process runs as a long-lived subprocess until explicitly stopped.

## Fallback Chain Behavior

The unified provider implements a cascading fallback chain in `auto` mode:

```
Request arrives
    |
    v
[API available?] --yes--> Try API
    |                        |
    no                    success? --yes--> Return response
    |                        |
    v                       no
[CLI available?] --yes--> Try CLI
    |                        |
    no                    success? --yes--> Return response
    |                        |
    v                       no
[ACP available?] --yes--> Try ACP
    |                        |
    no                    success? --yes--> Return response
    |                        |
    v                       no
Return error:              Return error:
"no access method          "all methods failed:
 available"                 <last error>"
```

Availability checks:
- **API**: `apiProvider != nil` (created when API key is present)
- **CLI**: `IsCLIAvailable()` checks PATH lookup, version command, and authentication
- **ACP**: `IsAvailable()` checks if `gemini` binary is in PATH

For streaming (`CompleteStream`), the same fallback order applies.

## Debate System Integration

The Gemini provider integrates with the HelixAgent debate system as a standard `LLMProvider`. Key integration points:

- **Provider Registry**: Registered as `"gemini"` in `internal/services/provider_registry.go`
- **Ensemble Voting**: Participates in confidence-weighted voting with scores derived from finish reason and content length
- **Debate Positions**: Can serve as any debate position (5 positions x N LLMs)
- **Multi-Round Debates**: Session management (CLI/ACP) preserves context across debate rounds
- **Debate Formatter**: Responses pass through `internal/services/debate_formatter_integration.go`

### Confidence Scoring

The API provider calculates confidence scores based on:

| Factor | Effect |
|--------|--------|
| Base confidence | 0.85 |
| Finish reason: STOP | +0.10 |
| Finish reason: MAX_TOKENS | -0.10 |
| Finish reason: SAFETY | -0.30 |
| Finish reason: RECITATION | -0.20 |
| Content length > 100 chars | +0.03 |
| Content length > 500 chars | +0.03 |

Final confidence is clamped to [0.0, 1.0].

## LLMsVerifier Scoring

During startup verification, Gemini models are evaluated through the 8-test pipeline with 5 weighted scoring components:

| Component | Weight |
|-----------|--------|
| ResponseSpeed | 25% |
| CostEffectiveness | 25% |
| ModelEfficiency | 20% |
| Capability | 20% |
| Recency | 10% |

Gemini is classified as an **API Key** provider type. Minimum score for inclusion in the debate team: 5.0.

## Retry Configuration

The API provider implements exponential backoff with jitter:

| Parameter | Default |
|-----------|---------|
| MaxRetries | 3 |
| InitialDelay | 1 second |
| MaxDelay | 30 seconds |
| Multiplier | 2.0x |
| Jitter | +10% random |

Retryable HTTP status codes: 429 (rate limit), 500, 502, 503, 504 (server errors).

Auth errors (401) get a single retry after 500ms to handle transient authentication issues.

## Troubleshooting

### API Mode Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `Gemini API error: 401` | Invalid or expired API key | Verify `GEMINI_API_KEY` at `aistudio.google.com` |
| `Gemini API error: 429` | Rate limit exceeded | Reduce request frequency or upgrade plan |
| `all retry attempts failed` | Persistent server errors | Check Google AI status page; increase MaxRetries |
| Empty response (no candidates) | Content blocked by safety | Prompt may trigger safety filters; rephrase |
| `context deadline exceeded` | Timeout too short for model | Increase timeout (default 180s); use flash model |

### CLI Mode Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `gemini command not found` | CLI not installed | Install from `cloud.google.com/sdk` |
| `CLI not authenticated` | No credentials in `~/.gemini/` | Run `gemini auth login` or set `GEMINI_API_KEY` |
| `exit code 1` with stderr | CLI execution failure | Check stderr output for details |
| Empty response | CLI returned no output | Verify prompt is non-empty; check CLI version |

### ACP Mode Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `--experimental-acp` not supported | CLI version too old | Update `gemini` CLI to latest version |
| `ACP process not running` | Process crashed | Check for initialization errors; restart |
| `session creation failed` | ACP handshake issue | Verify credentials; check process stderr |
| Timeout on response | Process hung | Increase timeout; check for deadlocks |

### General Debugging

1. Check available access methods: call `provider.GetAvailableAccessMethods()`
2. Force a specific method: set `GEMINI_PREFERRED_METHOD=api` (or `cli`/`acp`)
3. Run health check: `provider.HealthCheck()` tests the best available method
4. Check provider info: `gemini.GetGeminiProviderInfo()` returns full capability map

## Key Files

| File | Purpose |
|------|---------|
| `internal/llm/providers/gemini/gemini.go` | Unified provider, shared types, orchestration |
| `internal/llm/providers/gemini/gemini_api.go` | Direct REST API provider with retry, thinking, search |
| `internal/llm/providers/gemini/gemini_cli.go` | CLI headless provider with session management |
| `internal/llm/providers/gemini/gemini_acp.go` | ACP JSON-RPC 2.0 provider |
| `internal/llm/providers/gemini/gemini_test.go` | Comprehensive test suite |
| `internal/services/provider_registry.go` | Provider registration |
| `internal/llm/discovery/` | 3-tier model discovery |
