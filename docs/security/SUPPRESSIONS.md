# Security Scanner Suppressions

This document catalogs all security scanner suppressions across the HelixAgent project,
including the tool, rule ID, affected files, justification, and review status.

Last updated: 2026-03-08

---

## Overview

HelixAgent uses two security scanning tools:

- **gosec** -- Go source code static analysis for security vulnerabilities
- **Snyk** -- Dependency vulnerability scanning

All suppressions are reviewed periodically to confirm they remain valid.

---

## gosec Suppressions

Configuration file: `/.gosec.yml`

### Global Settings

| Setting | Value | Purpose |
|---------|-------|---------|
| `nosec` | `true` | Enables `// nolint:gosec` inline comments for per-line suppression |

### Excluded Directories

The following directories are excluded from gosec scanning entirely:

| Directory | Reason |
|-----------|--------|
| `vendor/` | Third-party dependency code; not under project control |
| `testdata/` | Test fixture data; not production code |
| `reports/` | Generated report output; not source code |
| `bin/` | Compiled binaries; not source code |
| `docs/` | Documentation files; not source code |

---

### Rule G404 -- Weak Random Number Generator (math/rand)

**Rule description**: G404 flags usage of `math/rand` instead of `crypto/rand`. This is
relevant for cryptographic contexts but is a false positive when randomness is used for
non-security purposes such as jitter in retry backoff timers.

**Justification**: All suppressed instances use `math/rand` for calculating jitter delays
in HTTP retry backoff logic within LLM provider clients. Jitter prevents thundering herd
problems when multiple retries fire simultaneously. Cryptographic randomness is not required
for backoff jitter; `math/rand` provides sufficient distribution and has lower overhead.

**Date added**: 2025 (initial gosec integration)
**Review status**: Confirmed -- all instances are retry jitter only

| # | File | Line | Review Status |
|---|------|------|---------------|
| 1 | `internal/llm/providers/ai21/ai21.go` | 540 | Confirmed |
| 2 | `internal/llm/providers/anthropic/anthropic.go` | 546 | Confirmed |
| 3 | `internal/llm/providers/cerebras/cerebras.go` | 584 | Confirmed |
| 4 | `internal/llm/providers/claude/claude.go` | 675 | Confirmed |
| 5 | `internal/llm/providers/cohere/cohere.go` | 626 | Confirmed |
| 6 | `internal/llm/providers/deepseek/deepseek.go` | 548 | Confirmed |
| 7 | `internal/llm/providers/fireworks/fireworks.go` | 559 | Confirmed |
| 8 | `internal/llm/providers/gemini/gemini.go` | 735 | Confirmed |
| 9 | `internal/llm/providers/groq/groq.go` | 597 | Confirmed |
| 10 | `internal/llm/providers/huggingface/huggingface.go` | 634 | Confirmed |
| 11 | `internal/llm/providers/mistral/mistral.go` | 658 | Confirmed |
| 12 | `internal/llm/providers/ollama/ollama.go` | 430 | Confirmed |
| 13 | `internal/llm/providers/openai/openai.go` | 511 | Confirmed |
| 14 | `internal/llm/providers/openrouter/openrouter.go` | 324 | Confirmed |
| 15 | `internal/llm/providers/perplexity/perplexity.go` | 502 | Confirmed |
| 16 | `internal/llm/providers/qwen/qwen.go` | 892 | Confirmed |
| 17 | `internal/llm/providers/replicate/replicate.go` | 508 | Confirmed |
| 18 | `internal/llm/providers/together/together.go` | 560 | Confirmed |
| 19 | `internal/llm/providers/xai/xai.go` | 571 | Confirmed |
| 20 | `internal/llm/providers/zai/zai.go` | 691 | Confirmed |
| 21 | `internal/llm/providers/zen/zen.go` | 922 | Confirmed |

**Total G404 suppressions**: 21 (all LLM provider retry jitter)

---

### Rule G115 -- Integer Overflow on Conversion

**Rule description**: G115 flags integer type conversions that could overflow when
converting between signed and unsigned or between different bit widths.

**Justification**: The suppressed instances in the formatter executor involve converting
format result sizes and counts between integer types. The values are bounded by the
formatter framework (document sizes are validated before formatting) and cannot reach
overflow thresholds in practice.

**Date added**: 2025
**Review status**: Confirmed -- values are bounded by input validation

| # | File | Line | Context | Review Status |
|---|------|------|---------|---------------|
| 1 | `internal/formatters/executor.go` | 221 | Format result size conversion | Confirmed |
| 2 | `internal/formatters/executor.go` | 228 | Format result size conversion | Confirmed |

**Total G115 suppressions**: 2

---

## Snyk Suppressions

Configuration file: `/.snyk`

### Policy Version

| Setting | Value |
|---------|-------|
| Version | v1.25.0 |
| Language | Go |
| Count Dependencies | true |

### Ignored Vulnerabilities

| # | Vulnerability ID | Package | Reason | Expires | Review Status |
|---|-----------------|---------|--------|---------|---------------|
| 1 | SNYK-GOLANG-GITHUBCOMGOLOGLOG-7261795 | `github.com/go-logr/logr` (transitive) | Logr is used transitively via OpenTelemetry and does not affect application security. The vulnerability applies to scenarios not exercised by HelixAgent. | 2026-06-01 | Confirmed |

**Total Snyk suppressions**: 1

---

## Summary

| Tool | Rule | Count | Category |
|------|------|-------|----------|
| gosec | G404 (weak RNG) | 21 | False positive -- retry jitter |
| gosec | G115 (integer overflow) | 2 | False positive -- bounded values |
| Snyk | SNYK-GOLANG-GITHUBCOMGOLOGLOG-7261795 | 1 | Transitive dependency, no impact |
| **Total** | | **24** | |

---

## Review Schedule

All suppressions are reviewed:

1. **On each security scan** -- verify suppressed rules still apply to the cited lines
2. **Quarterly** -- audit whether upstream fixes have addressed Snyk vulnerabilities
3. **On dependency updates** -- check if transitive dependency vulnerabilities are resolved

## Adding New Suppressions

When adding a new suppression:

1. Add the entry to the appropriate config file (`.gosec.yml` or `.snyk`)
2. Add a row to the corresponding table in this document
3. Include: rule ID, file path, line number, justification, date, and review status
4. Set review status to `needs-review` until confirmed by a second reviewer
5. Set an expiration date for Snyk suppressions (maximum 6 months)
