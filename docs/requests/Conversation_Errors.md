# Conversation Errors Analysis and Fixes

**Date:** 2026-02-20  
**Status:** Fixed  
**Related Challenge:** `./challenges/scripts/conversation_errors_fix_challenge.sh`

## Summary of Errors Found

Two errors were identified in the conversation output:

### Error 1: NVIDIA Duplicate Prefix Bug
**Location:** Line 38-39  
**Symptom:** Model reference displayed as `nvidia/nvidia/llama-3.1-nemotron-70b-instruct` instead of `nvidia/llama-3.1-nemotron-70b-instruct`

```
   Primary: nvidia/nvidia/llama-3.1-nemotron-70b-instruct (226 ms)
   ❓ Error: nvidia: API error: 404 - {...}
```

### Error 2: Claude CLI Empty stderr Bug  
**Location:** Line 54  
**Symptom:** Error message showed empty stderr, providing no diagnostic information

```
   ❓ Error: claude CLI failed: exit status 1 (stderr: )
```

---

## Root Cause Analysis

### Error 1: NVIDIA Duplicate Prefix

**Root Cause:** The model ID `nvidia/llama-3.1-nemotron-70b-instruct` already contains an `org/model` prefix (NVIDIA API convention). When displaying fallback messages, the format code was concatenating `provider + "/" + model`, resulting in `nvidia/nvidia/llama-3.1-nemotron-70b-instruct`.

**Files Affected:**
- `internal/handlers/debate_format_markdown.go`
- `internal/handlers/formatters.go`

**Technical Details:**
- NVIDIA API uses model IDs like `nvidia/llama-3.1-nemotron-70b-instruct` where `nvidia/` is the org prefix
- HuggingFace API uses `meta-llama/Llama-3.3-70B-Instruct` with org prefix
- Display format code was: `fmt.Sprintf("%s/%s", provider, model)`
- When model already has prefix matching provider, this doubled the prefix

### Error 2: Claude CLI Empty stderr

**Root Cause:** 
1. Claude CLI may output errors to stdout instead of stderr
2. Some failures produce no output at all
3. Running inside a Claude Code session causes silent failure (recursive CLI call protection)

**Files Affected:**
- `internal/llm/providers/claude/claude_cli.go`

**Technical Details:**
- Error handling only captured stderr: `fmt.Errorf("claude CLI failed: %w (stderr: %s)", err, stderr.String())`
- Claude CLI might fail with empty stderr when authentication issues occur
- Claude CLI fails silently when `CLAUDECODE` or `CLAUDE_CODE_ENTRYPOINT` env vars are set (inside session)

---

## Fixes Applied

### Fix 1: Model Reference Formatting Utility

**File:** `internal/handlers/debate_format_markdown.go`

Added a new utility function `formatModelRef()` that:
1. Checks if model ID already starts with provider prefix
2. Returns model ID as-is if prefix matches (case-insensitive)
3. Otherwise, concatenates `provider + "/" + model`

```go
func formatModelRef(provider, model string) string {
    if model == "" {
        return provider
    }
    if provider == "" {
        return model
    }
    // Check if model already starts with provider prefix
    prefix := provider + "/"
    if strings.HasPrefix(model, prefix) {
        return model // Model already has provider prefix, use as-is
    }
    // Check for common variations (e.g., provider name in different case)
    lowerPrefix := strings.ToLower(provider) + "/"
    if strings.HasPrefix(strings.ToLower(model), lowerPrefix) {
        return model
    }
    return provider + "/" + model
}
```

**Updated Functions:**
- `FormatFallbackIndicatorMarkdown()`
- `FormatFallbackTriggeredMarkdown()`
- `FormatFallbackSuccessMarkdown()`
- `FormatFallbackFailedMarkdown()`
- `FormatFallbackChainMarkdown()`
- `formatFallbackWithErrorANSI()`
- `formatFallbackChainPlain()`
- `FormatFallbackIndicatorForAllFormats()`
- All formatters in `internal/handlers/formatters.go`

### Fix 2: Improved Claude CLI Error Handling

**File:** `internal/llm/providers/claude/claude_cli.go`

**Changes:**

1. **Added session detection check:**
```go
func (p *ClaudeCLIProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    // Check if we're inside a Claude Code session (causes recursive CLI issues)
    if IsInsideClaudeCodeSession() {
        return nil, fmt.Errorf("claude CLI cannot run inside another Claude Code session (detected CLAUDECODE or CLAUDE_CODE_ENTRYPOINT env var)")
    }
    // ... rest of function
}
```

2. **Improved error message with both stdout and stderr:**
```go
if err != nil {
    // Build comprehensive error message with both stdout and stderr
    stdoutStr := strings.TrimSpace(stdout.String())
    stderrStr := strings.TrimSpace(stderr.String())
    var errorDetail strings.Builder
    if stderrStr != "" {
        errorDetail.WriteString(stderrStr)
    }
    if stdoutStr != "" {
        if errorDetail.Len() > 0 {
            errorDetail.WriteString(" | stdout: ")
        }
        errorDetail.WriteString(stdoutStr)
    }
    if errorDetail.Len() == 0 {
        errorDetail.WriteString("(no output captured)")
    }
    return nil, fmt.Errorf("claude CLI failed: %w (output: %s)", err, errorDetail.String())
}
```

3. **Added empty response handling with stderr check:**
```go
rawOutput := stdout.String()
if rawOutput == "" {
    // Check if stderr has content (might be error message)
    stderrStr := strings.TrimSpace(stderr.String())
    if stderrStr != "" {
        return nil, fmt.Errorf("claude CLI returned empty response with stderr: %s", stderrStr)
    }
    return nil, fmt.Errorf("claude CLI returned empty response (no stdout or stderr)")
}
```

---

## Tests Added

### Unit Tests for formatModelRef()
**File:** `internal/handlers/debate_format_markdown_test.go`

- `TestFormatModelRef` - 9 test cases covering:
  - Standard case (no prefix duplication)
  - NVIDIA model with org prefix
  - Meta model with org prefix
  - HuggingFace model with org prefix
  - Empty model returns provider
  - Empty provider returns model
  - Both empty returns empty
  - Claude model (no prefix duplication)
  - Case insensitive prefix matching

### Unit Tests for Fallback Display
**File:** `internal/handlers/debate_format_markdown_test.go`

- `TestFormatFallbackTriggeredMarkdown_NoDoublePrefix` - Validates:
  - NVIDIA model with org prefix does not double prefix
  - Standard model displays correctly

### Unit Tests for Claude CLI
**File:** `internal/llm/providers/claude/claude_cli_test.go`

- `TestIsInsideClaudeCodeSession` - Session detection function works
- `TestClaudeCLIProvider_Complete_InsideSession` - Complete blocks inside session
- `TestClaudeCLIProvider_CompleteStream_InsideSession` - CompleteStream blocks inside session
- `TestClaudeCLIProvider_Complete_EmptyResponse` - Handling of empty response

---

## Challenge Script

**File:** `challenges/scripts/conversation_errors_fix_challenge.sh`

Validates:
1. Model reference formatting (formatModelRef) - 5 tests
2. Fallback triggered markdown (no double prefix) - 2 tests
3. Claude CLI session detection - 3 tests
4. All fallback formatting tests pass
5. Code compiles
6. No regressions in existing tests

**Run with:**
```bash
./challenges/scripts/conversation_errors_fix_challenge.sh
```

---

## Expected Output After Fixes

### Before (Bug):
```
⚡ Analyst Fallback Triggered
   Primary: nvidia/nvidia/llama-3.1-nemotron-70b-instruct (226 ms)
   ❓ Error: nvidia: API error: 404 - {...}
   → Trying: huggingface/meta-llama/Llama-3.3-70B-Instruct
```

### After (Fixed):
```
⚡ Analyst Fallback Triggered
   Primary: nvidia/llama-3.1-nemotron-70b-instruct (226 ms)
   ❓ Error: nvidia: API error: 404 - {...}
   → Trying: huggingface/meta-llama/Llama-3.3-70B-Instruct
```

### Before (Bug):
```
⚡ Mediator Fallback Triggered
   Primary: claude/claude-sonnet-4-5-20250929 (1.4 s)
   ❓ Error: claude CLI failed: exit status 1 (stderr: )
   → Trying: huggingface/meta-llama/Llama-3.3-70B-Instruct
```

### After (Fixed - with better error details):
```
⚡ Mediator Fallback Triggered
   Primary: claude/claude-sonnet-4-5-20250929 (1.4 s)
   ❓ Error: claude CLI failed: exit status 1 (output: claude CLI cannot run inside another Claude Code session...)
   → Trying: huggingface/meta-llama/Llama-3.3-70B-Instruct
```

---

## Files Modified

1. `internal/handlers/debate_format_markdown.go`
   - Added `formatModelRef()` utility function
   - Updated all formatting functions to use `formatModelRef()`

2. `internal/handlers/formatters.go`
   - Updated HTMLFormatter, RTFFormatter, TerminalFormatter
   - Updated `FormatFallbackIndicatorForAllFormats()`

3. `internal/llm/providers/claude/claude_cli.go`
   - Added session detection in `Complete()` and `CompleteStream()`
   - Improved error message with stdout/stderr capture
   - Added empty response handling with stderr fallback

4. `internal/handlers/debate_format_markdown_test.go`
   - Added `TestFormatModelRef`
   - Added `TestFormatFallbackTriggeredMarkdown_NoDoublePrefix`

5. `internal/llm/providers/claude/claude_cli_test.go`
   - Added session detection tests
   - Added empty response handling tests

6. `challenges/scripts/conversation_errors_fix_challenge.sh` (NEW)
   - Comprehensive validation challenge

---

## Verification

All tests pass:
```bash
# Run all new tests
go test -v -run "TestFormatModelRef|TestFormatFallbackTriggeredMarkdown_NoDoublePrefix|TestIsInsideClaudeCodeSession" ./internal/handlers/... ./internal/llm/providers/claude/...

# Run challenge
./challenges/scripts/conversation_errors_fix_challenge.sh
```

---

## Lessons Learned

1. **Model ID conventions vary by provider:**
   - Some APIs (NVIDIA, HuggingFace) include org prefix in model ID
   - Others (Anthropic, OpenAI) use simple model names
   - Display code must handle both cases gracefully

2. **CLI error handling needs robustness:**
   - Never assume stderr contains error messages
   - Capture both stdout and stderr
   - Detect recursive call scenarios (inside session)
   - Provide meaningful error context even when output is empty

3. **Display formatting must be provider-aware:**
   - Use utility functions to normalize display formats
   - Avoid hardcoded format strings that assume simple provider/model patterns
