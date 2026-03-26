# Security Scan Resolutions

**Scanner:** gosec v2.x
**Date:** 2026-03-26
**Scope:** `./internal/...` (680 files, 321,522 lines)
**Exclusions:** G404 (weak random in retry jitter), G115 (integer conversions in formatters)

## Summary

| Severity | Count | Actionable | False Positive | Acceptable Risk |
|----------|-------|------------|----------------|-----------------|
| HIGH | 282 | 0 | 45 (G101) | 237 (G704 SSRF) |
| MEDIUM | 178 | 0 | 96 (G117) | 82 (G204+G304+G301/2/6+G705) |
| LOW | 4 | 0 | 0 | 4 (G104) |
| **Total** | **464** | **0** | **141** | **323** |

## Finding Details

### G704 — SSRF via taint analysis (237 instances, HIGH)

**Verdict:** Acceptable Risk

All 237 instances are HTTP calls to LLM provider APIs where URLs are constructed from:
- Environment variable configuration (`*_API_KEY`, `*_BASE_URL`)
- Hardcoded provider base URLs in provider implementations
- Health check endpoints defined in service configuration

These are **not user-controllable URLs**. The taint flows from trusted configuration, not request input. Mitigation: provider URLs are validated at startup by LLMsVerifier.

### G101 — Potential hardcoded credentials (45 instances, HIGH)

**Verdict:** False Positive

All instances are:
- Struct field names containing "key", "token", "secret" (data model definitions)
- Test constants with placeholder values like `"test-api-key"`
- Configuration field names, not actual credential values

Real credentials are loaded from environment variables or `.env` files at runtime.

### G117 — Struct fields matching secret patterns (96 instances, MEDIUM)

**Verdict:** False Positive

JSON struct tags like `session_token`, `api_key`, etc. are data model field names. They define the wire format, not credential storage. Actual values come from runtime configuration.

### G204 — Subprocess with variable (45 instances, MEDIUM)

**Verdict:** Acceptable Risk

All instances are CLI provider invocations:
- `claude -p --output-format json`
- `qwen --acp`
- `gemini -p --output-format json`
- `opencode serve`

Binary names are hardcoded constants. Arguments include user prompts but are passed as discrete arguments (not shell-interpreted). No command injection risk.

### G304 — File inclusion via variable (31 instances, MEDIUM)

**Verdict:** Acceptable Risk

File paths come from:
- MCP adapter configurations (admin-defined)
- Filesystem adapter with `utils.ValidatePath` + allowed paths whitelist
- Config file paths from environment variables

All paths are validated or come from trusted admin configuration.

### G705 — XSS via taint analysis (1 instance, MEDIUM)

**File:** `internal/mcp/bridge/sse_bridge.go:827`

**Verdict:** Acceptable Risk

The SSE bridge formats a message endpoint URL using `r.Host`. This is server-side SSE output to a machine client (not browser-rendered HTML). The data is consumed programmatically by MCP clients, not displayed in a browser DOM. No XSS risk.

### G301/G302/G306 — File permissions (5 instances, MEDIUM)

**Verdict:** Acceptable Risk (test code)

All 5 instances are in `internal/testing/helpers.go` — test fixture creation code:
- `os.MkdirAll(path, 0755)` — test directories need executable bit
- `os.WriteFile(path, content, 0755)` — test executable scripts need executable bit
- `os.WriteFile(path, content, 0644)` — test data files

These are test-only utilities, not production code.

### G104 — Unhandled errors (4 instances, LOW)

| File | Line | Code | Assessment |
|------|------|------|------------|
| `internal/testutil/infra.go:82` | `resp.Body.Close()` | Test utility — close error ignorable |
| `internal/testutil/infra.go:71` | `conn.Close()` | Test utility — close error ignorable |
| `internal/services/verification_debate.go:143` | `fmt.Sscanf(...)` | Best-effort confidence parsing, defaults to 0.0 on failure |
| `internal/llmops/experiments.go:399` | `h.Write([]byte(...))` | FNV hash.Write never returns error (per Go docs) |

All 4 are acceptable — test utilities where close errors don't matter, and production code where the error is either impossible (hash.Write) or gracefully degraded (Sscanf defaults).

## Conclusion

**Zero actionable findings.** All 464 findings are either false positives (141) or acceptable risks with documented mitigations (323). The codebase follows security best practices with:
- Input validation at system boundaries
- Parameterized SQL queries (pgx)
- Path validation for file operations
- Trusted configuration for provider URLs
- No hardcoded credentials in production code
