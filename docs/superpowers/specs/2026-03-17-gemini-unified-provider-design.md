# Phase 1: Unified Gemini Provider — Design Spec

**Date:** 2026-03-17
**Status:** Approved
**Scope:** Enhance existing Gemini provider with CLI proxy + ACP fallback, extended thinking, Google Search grounding, updated model list

## Problem

The current Gemini provider (`internal/llm/providers/gemini/gemini.go`) is API-only. It lacks CLI headless and ACP fallback modes that other providers (Claude, Junie, Qwen, Zen) already have. The model list is outdated (defaults to `gemini-2.0-flash`), extended thinking is not implemented despite `SupportsReasoning: true`, and Google Search grounding is unused.

## Solution

Refactor into a unified provider with three access methods (API, CLI, ACP) using auto-detect priority, add extended thinking support, enable Google Search grounding on all requests, and update the model inventory to include Gemini 3.x models.

## Architecture

### File Structure

```
internal/llm/providers/gemini/
├── gemini.go          # Unified orchestrator (refactored)
├── gemini_api.go      # API implementation (extracted from current gemini.go)
├── gemini_cli.go      # CLI headless proxy (new)
├── gemini_acp.go      # ACP JSON-RPC 2.0 protocol (new)
├── gemini_test.go     # Unified orchestrator tests (enhanced)
├── gemini_api_test.go # API-specific tests (extracted + new)
├── gemini_cli_test.go # CLI proxy tests (new)
├── gemini_acp_test.go # ACP protocol tests (new)
```

### Access Method Priority (Auto-Detect)

1. `GEMINI_API_KEY` present → use API provider (fastest, most reliable)
2. `gemini` CLI installed + authenticated → use CLI headless proxy
3. `gemini --experimental-acp` available → use ACP protocol
4. Per-request fallback: if primary method fails, cascade to next available

### Components

**GeminiUnifiedProvider (gemini.go)**
- Fields: apiProvider, cliProvider, acpProvider, preferredMethod, sync.Once init
- Auto-detect which methods are available
- Fallback chain on per-request failure
- Merges capabilities from all sub-providers

**GeminiAPIProvider (gemini_api.go)**
- Extracted from current gemini.go (all existing API logic preserved)
- New: thinkingConfig with budget 0-8192 for models that support it (2.5-pro, 3-pro, 3.1-pro)
- New: Google Search grounding always-on (adds googleSearch tool to every request)
- New: Updated model fallback list with Gemini 3.x
- New: Max output tokens cap raised to 65536 for 2.5+ models

**GeminiCLIProvider (gemini_cli.go)**
- Follows Junie CLI pattern
- Headless: `gemini -p "prompt" --output-format json`
- Model selection: `-m <model>`
- Session tracking: parse responseId, pass `--resume <id>`
- Streaming: `--output-format stream-json` (JSONL events)
- Non-interactive: `--approval-mode auto_edit`
- Argument validation via utils.ValidateCommandArg()

**GeminiACPProvider (gemini_acp.go)**
- Follows Qwen ACP pattern
- Launch: `gemini --experimental-acp`
- JSON-RPC 2.0 over stdin/stdout
- Methods: initialize, newSession, prompt
- Channel-based response routing with sync.RWMutex
- Graceful shutdown with process cleanup

### Models (Hardcoded Fallback)

- gemini-3.1-pro-preview (latest)
- gemini-3-pro-preview
- gemini-3-flash-preview
- gemini-2.5-pro (stable)
- gemini-2.5-flash (stable)
- gemini-2.5-flash-lite (stable)
- gemini-2.0-flash (legacy, backward compat)
- gemini-embedding-001

### Extended Thinking

API request adds to GenerationConfig:
```json
{"thinkingConfig": {"thinkingBudget": 8192}}
```
Only for models: gemini-2.5-pro, gemini-3-pro-preview, gemini-3.1-pro-preview.
Thinking content extracted from response parts into metadata["thinking"].

### Google Search Grounding

Always-on for API requests. Adds to tools array:
```json
{"googleSearch": {}}
```

### Provider Registry Integration

`case "gemini"` in provider_registry.go uses `NewGeminiUnifiedProvider(config)` with auto-detect. OAuth detection triggers CLI/ACP fallback paths (same pattern as Claude/Qwen cases).

### Error Handling

- API: existing retry with exponential backoff + jitter (429, 5xx)
- CLI: exit code mapping (0=success, 1=error, 42=invalid input, 53=turn limit)
- ACP: JSON-RPC error codes
- All methods respect context.Context deadlines
- Fallback triggered on any request failure

### Testing

- Unit: table-driven with httptest (API), mocked exec (CLI), mocked stdio (ACP)
- Integration: real API calls (skipped without GEMINI_API_KEY)
- Benchmark: request conversion, confidence calculation, JSON parsing
- Security: command injection validation, API key handling
- Stress: concurrent requests across all methods

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Access priority | Auto-detect (C) | Handles all deployment scenarios |
| Model scope | All models (A) | Curated fallback + dynamic discovery |
| Extended thinking | Full API impl (A) | Competitive in debate rounds |
| Google Search | Always-on (B) | Maximum factual grounding |
| Session mgmt | Full tracking (A) | Context continuity for debate |

## Out of Scope (Later Phases)

- LLMsVerifier integration (Phase 2)
- Debate system wiring (Phase 3)
- Challenge suite (Phase 4)
- Model routing, context compression, subagents, extensions (Phase 5)
- Documentation, guides, diagrams (Phase 6)
