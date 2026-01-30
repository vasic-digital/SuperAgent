# Cognee Integration Bug - 2026-01-30

## Issue Summary

Cognee container starts successfully and passes health checks, but API endpoints timeout due to an internal bug in the Cognee library.

## Root Cause

**AttributeError in Cognee's extract_subgraph_chunks task:**

```python
AttributeError: 'str' object has no attribute 'nodes'
```

**Location:** `/app/cognee/tasks/memify/extract_subgraph_chunks.py:9`

**What happens:**
1. Cognee receives a list of strings from previous pipeline stage
2. `extract_subgraph_chunks` expects subgraph objects with a `.nodes` attribute
3. Code tries to call `.nodes.values()` on string objects, causing AttributeError
4. API requests hang indefinitely waiting for the crashed task

## Affected Endpoints

- ✅ `GET /` - Health check (works)
- ✅ `POST /api/v1/auth/login` - Authentication (works)
- ❌ `POST /api/v1/search` - Search (times out)
- ❌ `POST /api/v1/add` - Add data (times out)
- ❌ `POST /api/v1/memify` - Memory enrichment (times out)

## Impact

- **API Response Time:** Requests take 30 seconds instead of <5 seconds
- **User Experience:** OpenCode and other CLI agents experience timeouts
- **Functionality:** AI Debate works without Cognee (uses base LLM providers)

## Investigation Timeline

1. **Initial symptom:** API taking 30 seconds per request
2. **First hypothesis:** Cognee authentication failure
   - ✓ Added auth credentials to config
   - ✓ Confirmed authentication succeeds
   - ✗ API still timing out
3. **Second hypothesis:** Missing Cognee configuration
   - ✓ Verified container running
   - ✓ Health endpoint returns 200 OK
   - ✗ API endpoints still timeout
4. **Root cause identified:** Cognee internal bug
   - Examined container logs: `AttributeError: 'str' object has no attribute 'nodes'`
   - Confirmed bug is in Cognee's codebase, not HelixAgent

## Current Workaround

Cognee integration is **disabled** in `configs/development.yaml`:

```yaml
ai_debate:
  cognee:
    enabled: false  # Disabled until Cognee bug is fixed

services:
  cognee:
    enabled: true    # Container still starts (for future use)
    required: false  # But not required for system boot

feature_flags:
  cognee_integration: false  # API calls disabled
```

**Result:** API responds in <5 seconds, AI Debate works normally without Cognee enhancements.

## Why Challenges Passed

Challenge scripts don't test Cognee search functionality:
- Challenges test AI Debate consensus (works without Cognee)
- Challenges test LLM provider verification (independent of Cognee)
- Challenges test service boot (Cognee container starts successfully)
- Challenges don't make `/api/v1/search` or `/api/v1/memify` calls

## Potential Solutions

1. **Report to Cognee developers** (upstream fix required)
2. **Use different Cognee version** (test other releases)
3. **Fork and patch Cognee** (not recommended, high maintenance)
4. **Alternative knowledge graph service** (evaluate LangGraph, Mem0, etc.)

## Verification Tests

Created comprehensive verification tests in:
- `tests/integration/cognee_verification_test.go`
- `challenges/scripts/cognee_verification_challenge.sh`

Tests validate:
- ✓ Cognee container running
- ✓ Health endpoint responds
- ✓ Authentication succeeds
- ✓ Token obtained successfully
- ❌ Search API times out (documented bug)
- ✓ System works without Cognee

## References

- Container logs: `podman logs helixagent-cognee`
- Authentication implementation: `internal/services/cognee_service.go`
- Manual test script: `/tmp/test_cognee_manual.py`

## Status

**BLOCKED** - Waiting for upstream Cognee fix

**Workaround:** Disabled Cognee API integration while keeping container running for future use.
