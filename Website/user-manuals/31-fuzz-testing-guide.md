# User Manual 31: Fuzz Testing Guide

**Version:** 1.0
**Last Updated:** March 17, 2026
**Audience:** Developers, QA Engineers, Security Engineers

---

## Table of Contents

1. [Overview](#overview)
2. [What is Fuzz Testing?](#what-is-fuzz-testing)
3. [Go Native Fuzzing](#go-native-fuzzing)
4. [Running Fuzz Tests](#running-fuzz-tests)
5. [Writing New Fuzz Tests](#writing-new-fuzz-tests)
6. [Interpreting Results](#interpreting-results)
7. [Available Fuzz Targets](#available-fuzz-targets)
8. [Advanced Techniques](#advanced-techniques)
9. [Integration with CI/CD](#integration-with-cicd)
10. [Troubleshooting](#troubleshooting)

---

## Overview

HelixAgent uses Go native fuzz testing (introduced in Go 1.18) to discover edge cases, panics, and unexpected behavior in critical parsing and validation code. Fuzz testing complements unit tests by generating millions of semi-random inputs that exercise code paths developers never think to test manually.

Fuzz testing is particularly valuable for:
- API request parsing (JSON, form data, query parameters)
- Protocol parsing (SSE streams, WebSocket frames)
- Input validation (model IDs, authentication tokens, tool schemas)
- Serialization/deserialization boundaries

---

## What is Fuzz Testing?

Traditional unit tests verify that known inputs produce expected outputs. Fuzz testing inverts this approach: it generates *unknown* inputs and verifies that the code does not crash, panic, or corrupt state.

### How Go's Fuzzer Works

Go's native fuzzer uses **coverage-guided mutation**:

1. **Seed corpus**: You provide known inputs (good and bad) as starting points
2. **Mutation**: The fuzzer mutates these inputs (bit flips, byte insertions, deletions, value substitutions)
3. **Coverage tracking**: The fuzzer monitors which code branches each mutated input exercises
4. **Selection**: Inputs that reach new code paths are saved and further mutated
5. **Crash detection**: If any input causes a panic, the fuzzer saves it and reports the failure

This process is iterative. Over time, the fuzzer builds a corpus of inputs that together exercise deep code paths, often discovering bugs that are invisible to manual testing.

### What Fuzzing Catches

| Bug Type | Example | How Fuzzing Finds It |
|----------|---------|---------------------|
| Nil pointer dereference | JSON with missing required field | Random field omission |
| Index out of bounds | Malformed model ID "a/" | Short/truncated strings |
| Integer overflow | Extremely large max_tokens value | Large number mutation |
| Infinite loop | Deeply nested JSON objects | Recursive structure generation |
| Memory exhaustion | 10MB request body | Large input mutation |
| Unicode handling | Invalid UTF-8 sequences | Byte-level mutation |
| Logic errors | Empty array where non-empty expected | Edge case values |

---

## Go Native Fuzzing

### Prerequisites

- Go 1.18+ (HelixAgent uses Go 1.24+)
- No external tools required -- fuzzing is built into `go test`

### Fuzz Test Structure

A fuzz test has three parts:

```go
func FuzzFunctionName(f *testing.F) {
    // 1. SEED CORPUS: Add known inputs
    f.Add([]byte(`{"valid": "json"}`))
    f.Add([]byte(`invalid`))
    f.Add([]byte{})

    // 2. FUZZ FUNCTION: Called with mutated inputs
    f.Fuzz(func(t *testing.T, data []byte) {
        // 3. INVARIANT CHECK: Must not panic for any input
        result, err := ParseRequest(data)
        if err != nil {
            return // Errors are acceptable; panics are not
        }
        // Additional validation on successful parse
        if result.Model == "" && err == nil {
            t.Error("parsed successfully but model is empty")
        }
    })
}
```

### Seed Corpus

The seed corpus provides starting inputs that the fuzzer mutates. Good seeds include:

- **Valid inputs**: Typical well-formed requests
- **Boundary inputs**: Empty strings, zero-length arrays, max-length values
- **Known bad inputs**: Malformed data that should be rejected gracefully
- **Edge cases**: Unicode, null bytes, very long strings

```go
f.Add([]byte(`{"model":"deepseek-chat","messages":[{"role":"user","content":"hi"}]}`))
f.Add([]byte(`{}`))                           // Minimal valid JSON
f.Add([]byte(`not json`))                     // Invalid format
f.Add([]byte{0x00, 0xff, 0xfe})              // Binary data
f.Add([]byte(`{"model":"` + strings.Repeat("a", 10000) + `"}`))  // Long string
```

---

## Running Fuzz Tests

### Basic Execution

```bash
# Run a specific fuzz target for 30 seconds
go test -fuzz=FuzzJSONRequestParsing -fuzztime=30s ./tests/fuzz/

# Run for 1 minute with verbose output
go test -fuzz=FuzzJSONRequestParsing -fuzztime=1m -v ./tests/fuzz/
```

### Resource-Limited Execution

All fuzz testing on the HelixAgent host must respect the 30-40% resource limit:

```bash
# Recommended: resource-limited fuzzing
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -fuzz=FuzzJSONRequestParsing -fuzztime=60s -p 1 ./tests/fuzz/
```

### Running as Regression Tests

Without the `-fuzz` flag, fuzz tests run against the saved corpus only (fast, deterministic):

```bash
# Run all fuzz corpus entries as regression tests
go test -run=Fuzz ./tests/fuzz/

# Run a specific fuzz target's corpus
go test -run=FuzzJSONRequestParsing ./tests/fuzz/
```

### Extended Fuzzing Sessions

For thorough coverage, run overnight sessions:

```bash
# 8-hour fuzzing session (run in background)
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -fuzz=FuzzJSONRequestParsing -fuzztime=8h -p 1 ./tests/fuzz/ \
  > fuzz-results.log 2>&1 &
```

### Running All Fuzz Targets

```bash
# Run each fuzz target for 30 seconds
for target in FuzzJSONRequestParsing FuzzToolSchemaValidation FuzzSSEParsing \
              FuzzModelIDParsing FuzzHTTPHeaderParsing; do
  echo "Fuzzing $target..."
  GOMAXPROCS=2 nice -n 19 ionice -c 3 \
    go test -fuzz=$target -fuzztime=30s -p 1 ./tests/fuzz/ 2>&1
done
```

---

## Writing New Fuzz Tests

### Step 1: Identify the Target

Good fuzz targets are functions that:
- Accept untrusted input (user requests, external data)
- Parse structured data (JSON, XML, protocol buffers)
- Perform validation logic
- Handle string manipulation

### Step 2: Define the Invariant

Determine what must be true for *all* inputs:
- The function must not panic
- The function must not corrupt shared state
- If parsing succeeds, the result must be internally consistent
- Memory allocation must be bounded

### Step 3: Write the Fuzz Function

```go
func FuzzNewTarget(f *testing.F) {
    // Add seed corpus entries
    f.Add("valid-input-1")
    f.Add("valid-input-2")
    f.Add("")
    f.Add("boundary-case")

    f.Fuzz(func(t *testing.T, input string) {
        // Call the function under test
        result, err := TargetFunction(input)

        // Invariant: must not panic (automatic)
        // Invariant: error or valid result
        if err != nil {
            return
        }

        // Invariant: result consistency
        if result.Field == "" {
            t.Errorf("successful parse but empty field for input: %q", input)
        }
    })
}
```

### Step 4: Run and Iterate

```bash
# Initial run to verify no immediate crashes
go test -fuzz=FuzzNewTarget -fuzztime=10s ./path/to/package/

# Extend duration once initial issues are fixed
go test -fuzz=FuzzNewTarget -fuzztime=5m ./path/to/package/
```

### Multiple Parameter Types

Go fuzzing supports these parameter types: `string`, `[]byte`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `bool`, `rune`.

```go
func FuzzScoringWeights(f *testing.F) {
    f.Add(0.25, 0.25, 0.20, 0.20, 0.10)

    f.Fuzz(func(t *testing.T, speed, cost, efficiency, capability, recency float64) {
        weights := ScoringWeights{
            ResponseSpeed:     speed,
            CostEffectiveness: cost,
            ModelEfficiency:   efficiency,
            Capability:        capability,
            Recency:           recency,
        }
        // Must not panic regardless of weight values
        _ = weights.Validate()
        _ = weights.Normalize()
    })
}
```

---

## Interpreting Results

### Successful Run

```
fuzz: elapsed: 30s, gathering baseline coverage: 0/42 completed
fuzz: elapsed: 30s, gathering baseline coverage: 42/42 completed, now fuzzing with 2 workers
fuzz: elapsed: 33s, execs: 150234 (50102/sec), new interesting: 12 (total: 54)
fuzz: elapsed: 36s, execs: 301456 (50257/sec), new interesting: 15 (total: 57)
...
PASS
```

Key metrics:
- **execs**: Total inputs tested
- **execs/sec**: Throughput (higher is better)
- **new interesting**: Inputs that found new code paths (diminishing returns expected)

### Failure Found

```
--- FAIL: FuzzJSONRequestParsing (2.31s)
    --- FAIL: FuzzJSONRequestParsing/6a3c8b2f (0.00s)
        testing.go:1356: panic: runtime error: index out of range [5] with length 3
            goroutine 35 [running]:
            dev.helix.agent/internal/handlers.parseModelID(...)
                /internal/handlers/model_parser.go:42
```

The failing input is saved to:
```
testdata/fuzz/FuzzJSONRequestParsing/6a3c8b2f
```

### Fixing a Fuzz Failure

1. Read the failing input file:
   ```bash
   cat testdata/fuzz/FuzzJSONRequestParsing/6a3c8b2f
   ```

2. Write a deterministic unit test:
   ```go
   func TestParseModelID_FuzzRegression(t *testing.T) {
       input := "a/"  // From fuzz corpus
       _, err := parseModelID(input)
       assert.Error(t, err, "should reject model ID with empty name")
   }
   ```

3. Fix the bug (add bounds checking, nil checks, etc.)

4. Verify the fix:
   ```bash
   go test -run=FuzzJSONRequestParsing ./tests/fuzz/
   ```

---

## Available Fuzz Targets

### FuzzJSONRequestParsing

**Location:** `tests/fuzz/`
**Input:** `[]byte` (raw HTTP request body)
**Tests:** Chat completion request JSON parsing robustness

Validates that the JSON request parser handles any input without panicking, including:
- Malformed JSON
- Missing required fields
- Unexpected field types
- Deeply nested objects
- Unicode and binary data

### FuzzToolSchemaValidation

**Location:** `tests/fuzz/`
**Input:** `[]byte` (tool schema JSON)
**Tests:** Tool schema parsing and parameter validation

Ensures the tool schema validator handles malformed schemas, missing parameter definitions, and invalid type specifications.

### FuzzSSEParsing

**Location:** `tests/fuzz/`
**Input:** `[]byte` (SSE event stream data)
**Tests:** Server-Sent Events stream parsing

Validates that the SSE parser handles incomplete events, missing newlines, oversized data fields, and interleaved binary data.

### FuzzModelIDParsing

**Location:** `tests/fuzz/`
**Input:** `string` (model identifier)
**Tests:** Model ID format parsing (provider/model)

Ensures model ID parsing handles edge cases: empty strings, missing separators, multiple separators, very long names, special characters.

### FuzzHTTPHeaderParsing

**Location:** `tests/fuzz/`
**Input:** `string` (HTTP header value)
**Tests:** Authorization and Content-Type header parsing

Validates that header parsing handles malformed Bearer tokens, invalid API keys, and unexpected content types without panicking.

### FuzzTokenBucket (Toolkit)

**Location:** `Toolkit/pkg/toolkit/common/ratelimit/`
**Input:** Multiple parameters (rate, burst, requests)
**Tests:** Rate limiter token bucket algorithm correctness

Validates that the token bucket rate limiter handles extreme rate values, zero burst sizes, and negative parameters.

### FuzzDefaultCategoryInferrer (Toolkit)

**Location:** `Toolkit/pkg/toolkit/common/discovery/`
**Input:** `string` (model name)
**Tests:** Model category inference from name patterns

Ensures the category inferrer handles unusual model names, empty strings, and very long inputs.

---

## Advanced Techniques

### Custom Mutators

For domain-specific inputs, guide the fuzzer with structured seeds:

```go
func FuzzDebateConfig(f *testing.F) {
    // Add structurally valid JSON with varying topologies
    topologies := []string{"mesh", "star", "chain", "tree"}
    for _, topo := range topologies {
        cfg := fmt.Sprintf(`{"topology":"%s","rounds":3,"providers":["a","b"]}`, topo)
        f.Add([]byte(cfg))
    }

    f.Fuzz(func(t *testing.T, data []byte) {
        var config DebateConfig
        if err := json.Unmarshal(data, &config); err != nil {
            return
        }
        // Validate does not panic
        _ = config.Validate()
    })
}
```

### Property-Based Assertions

Go beyond "no panics" to check invariants:

```go
f.Fuzz(func(t *testing.T, data []byte) {
    parsed, err := ParseRequest(data)
    if err != nil {
        return
    }

    // Property: re-serialization should produce equivalent result
    serialized, err := json.Marshal(parsed)
    require.NoError(t, err)

    reparsed, err := ParseRequest(serialized)
    require.NoError(t, err)
    assert.Equal(t, parsed.Model, reparsed.Model)
})
```

### Corpus Management

```bash
# View corpus entries
ls testdata/fuzz/FuzzJSONRequestParsing/

# Add a manually crafted corpus entry
echo -n '{"model":"test","messages":[]}' > \
  testdata/fuzz/FuzzJSONRequestParsing/manual_empty_messages

# Clean corpus (remove redundant entries)
go test -fuzz=FuzzJSONRequestParsing -fuzztime=1s -fuzzminimizetime=5m ./tests/fuzz/
```

---

## Integration with CI/CD

### Regression Testing in CI

Run fuzz corpus entries as part of the regular test suite (fast, deterministic):

```bash
# Runs all corpus entries without generating new inputs
make test-unit  # Includes -run=Fuzz by default
```

### Nightly Fuzzing Runs

Schedule extended fuzzing sessions to discover new issues:

```bash
# Run each target for 10 minutes nightly
for target in FuzzJSONRequestParsing FuzzToolSchemaValidation FuzzSSEParsing \
              FuzzModelIDParsing FuzzHTTPHeaderParsing; do
  GOMAXPROCS=2 nice -n 19 ionice -c 3 \
    go test -fuzz=$target -fuzztime=10m -p 1 ./tests/fuzz/ 2>&1 \
    | tee "reports/fuzz/$target-$(date +%Y%m%d).log"
done
```

### Commit Corpus Files

Always commit `testdata/fuzz/` directories to version control. These files serve as regression tests -- if a future code change reintroduces a bug, the corpus entry will catch it in regular `go test` runs.

---

## Troubleshooting

### Fuzzer runs slowly

**Symptom:** Low execs/sec rate.

**Solutions:**
1. Ensure GOMAXPROCS allows at least 2 workers
2. Reduce the fuzz function complexity (avoid network calls)
3. Use `[]byte` instead of `string` for large inputs
4. Profile the fuzz function to find bottlenecks

### Fuzzer finds too many crashes

**Symptom:** Fuzzer stops after the first crash.

**Solutions:**
1. Fix the crash and restart fuzzing
2. The fuzzer stops on first failure by design -- fix and re-run iteratively
3. If the crash is in a known-bad area, add input validation before the fuzz target

### Corpus files grow too large

**Symptom:** `testdata/fuzz/` directory becomes very large.

**Solutions:**
1. Use `-fuzzminimizetime` to reduce corpus entries to minimal reproducing inputs
2. Periodically clean the corpus: `go test -fuzz=Target -fuzztime=1s -fuzzminimizetime=10m`
3. Remove duplicate or redundant entries manually

### Fuzz test passes but unit tests fail

**Symptom:** Fuzz test does not catch a known bug.

**Solutions:**
1. Check the seed corpus -- does it include inputs similar to the failing case?
2. Increase fuzz duration for better coverage
3. Add the specific failing input to the seed corpus
4. Verify the fuzz function checks the right invariant

---

## Related Resources

- [User Manual 20: Testing Strategies](20-testing-strategies.md) -- Overview of all test types
- [User Manual 17: Security Scanning Guide](17-security-scanning-guide.md) -- Complementary security testing
- [Video Course 64: Fuzz Testing Mastery](../video-courses/video-course-64-fuzz-testing-mastery.md) -- Video walkthrough
- [Go Fuzz Testing Documentation](https://go.dev/doc/security/fuzz/)
- [Go Testing Package](https://pkg.go.dev/testing#F)

---

**Next Manual:** User Manual 32 - Automated Security Scanning
