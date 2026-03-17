# Video Course 64: Fuzz Testing Mastery

## Course Overview

**Duration:** 2 hours
**Level:** Intermediate
**Prerequisites:** Course 01 (Fundamentals), Course 06 (Testing), Go 1.18+ familiarity

Master Go native fuzz testing for API robustness. Learn to write fuzz targets, run fuzzing campaigns, interpret results, fix discovered bugs, and integrate fuzzing into the development workflow.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain how coverage-guided fuzzing discovers bugs
2. Write effective fuzz tests with well-crafted seed corpora
3. Run fuzzing campaigns with appropriate resource limits
4. Interpret fuzzer output and fix discovered crashes
5. Manage fuzz corpora for regression testing
6. Apply property-based assertions beyond "no panics"

---

## Module 1: Fuzzing Theory (20 min)

### Video 1.1: What is Coverage-Guided Fuzzing? (10 min)

**Topics:**
- The limitations of example-based testing
- Coverage-guided mutation: how Go's fuzzer works internally
- Corpus growth: seed inputs mutated to maximize code coverage
- The fuzzer's feedback loop: mutate, execute, observe coverage, select
- Why fuzzing finds bugs that humans miss

**Analogy:**
Traditional tests are like checking specific doors in a building. Fuzz testing is like shaking every door handle, window, and vent -- you discover that the service entrance was left unlocked.

### Video 1.2: What Fuzzing Catches (10 min)

**Topics:**
- Panics from nil pointer dereferences
- Index out of bounds on malformed input
- Integer overflows in size calculations
- Infinite loops from crafted input
- Memory exhaustion from recursive structures
- Unicode/encoding edge cases
- Logic errors where parsing "succeeds" but produces wrong results

**Real Example from HelixAgent:**
```go
// Bug: model ID "provider/" causes index out of bounds
func parseModelID(id string) (string, string) {
    parts := strings.SplitN(id, "/", 2)
    return parts[0], parts[1] // PANIC if parts has length 1
}

// Fixed: bounds check
func parseModelID(id string) (string, string, error) {
    parts := strings.SplitN(id, "/", 2)
    if len(parts) != 2 || parts[1] == "" {
        return "", "", fmt.Errorf("invalid model ID format: %q", id)
    }
    return parts[0], parts[1], nil
}
```

---

## Module 2: Writing Fuzz Tests (40 min)

### Video 2.1: Fuzz Test Anatomy (15 min)

**Topics:**
- The `Fuzz` prefix naming convention
- `*testing.F` parameter
- `f.Add()` for seed corpus entries
- `f.Fuzz()` with the fuzz function
- Supported parameter types: string, []byte, int variants, float variants, bool, rune

**Basic Template:**
```go
func FuzzTargetFunction(f *testing.F) {
    // Step 1: Seed corpus
    f.Add([]byte("valid input"))
    f.Add([]byte(""))
    f.Add([]byte("edge case"))

    // Step 2: Fuzz function
    f.Fuzz(func(t *testing.T, data []byte) {
        // Step 3: Call target -- must not panic
        result, err := TargetFunction(data)
        if err != nil {
            return // Errors are OK; panics are not
        }

        // Step 4: Validate invariants on success
        if result.Field == "" {
            t.Errorf("empty field on successful parse")
        }
    })
}
```

### Video 2.2: Crafting Effective Seed Corpora (15 min)

**Topics:**
- Why seeds matter: they guide the fuzzer's initial mutations
- Categories of seeds: valid, boundary, invalid, adversarial
- Using real-world data from production logs (sanitized)
- Seeding with protocol-specific edge cases

**Seed Strategy for JSON Parsing:**
```go
func FuzzJSONRequestParsing(f *testing.F) {
    // Valid: typical request
    f.Add([]byte(`{"model":"deepseek-chat","messages":[{"role":"user","content":"hello"}]}`))

    // Valid: minimal
    f.Add([]byte(`{"model":"test","messages":[]}`))

    // Valid: with all optional fields
    f.Add([]byte(`{"model":"test","messages":[{"role":"user","content":"hi"}],"temperature":0.7,"max_tokens":100,"stream":true}`))

    // Boundary: empty object
    f.Add([]byte(`{}`))

    // Boundary: empty string
    f.Add([]byte(``))

    // Invalid: not JSON
    f.Add([]byte(`not json at all`))

    // Invalid: truncated
    f.Add([]byte(`{"model":"test`))

    // Adversarial: deeply nested
    f.Add([]byte(`{"a":{"b":{"c":{"d":{"e":"deep"}}}}}`))

    // Adversarial: very long string
    f.Add([]byte(`{"model":"` + strings.Repeat("x", 10000) + `"}`))

    // Adversarial: null bytes
    f.Add([]byte{0x7b, 0x00, 0x7d}) // {NUL}

    f.Fuzz(func(t *testing.T, data []byte) {
        var req ChatCompletionRequest
        _ = json.Unmarshal(data, &req)
    })
}
```

### Video 2.3: Multi-Parameter Fuzz Tests (10 min)

**Topics:**
- Using multiple parameters for structured input
- Type combinations (string + int, float + float)
- Testing scoring with arbitrary weight combinations
- Testing rate limiters with arbitrary rate/burst values

**Example:**
```go
func FuzzScoringCalculation(f *testing.F) {
    f.Add(0.25, 0.25, 0.20, 0.20, 0.10, 8.0)

    f.Fuzz(func(t *testing.T, speed, cost, efficiency, capability, recency, rawScore float64) {
        weights := Weights{speed, cost, efficiency, capability, recency}
        // Must not panic even with NaN, Inf, negative weights
        _ = CalculateScore(weights, rawScore)
    })
}
```

---

## Module 3: Running and Managing Fuzzing Campaigns (25 min)

### Video 3.1: Running Fuzz Tests (10 min)

**Topics:**
- The `-fuzz` flag to select a target
- The `-fuzztime` flag to set duration
- Resource-limited execution with GOMAXPROCS and nice
- Monitoring fuzzer progress (execs/sec, new interesting)

**Commands:**
```bash
# Quick run (30 seconds)
go test -fuzz=FuzzJSONRequestParsing -fuzztime=30s ./tests/fuzz/

# Resource-limited run
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -fuzz=FuzzJSONRequestParsing -fuzztime=5m -p 1 ./tests/fuzz/

# Extended overnight run
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -fuzz=FuzzJSONRequestParsing -fuzztime=8h -p 1 ./tests/fuzz/ \
  > fuzz-overnight.log 2>&1 &
```

### Video 3.2: Corpus Management (10 min)

**Topics:**
- Where corpus files are stored: `testdata/fuzz/<FuzzName>/`
- Manual corpus entries
- Corpus minimization with `-fuzzminimizetime`
- Committing corpus to version control for regression testing
- Corpus growth over time

**Commands:**
```bash
# View corpus
ls testdata/fuzz/FuzzJSONRequestParsing/

# Read a specific corpus entry
cat testdata/fuzz/FuzzJSONRequestParsing/abc123

# Minimize corpus
go test -fuzz=FuzzJSONRequestParsing -fuzztime=1s -fuzzminimizetime=5m ./tests/fuzz/
```

### Video 3.3: Regression Testing with Corpus (5 min)

**Topics:**
- Running fuzz tests without `-fuzz` flag replays corpus only
- Each corpus entry becomes a deterministic sub-test
- Fast execution (milliseconds per entry)
- Integration with `make test-unit`

---

## Module 4: Fixing Fuzz Failures (20 min)

### Video 4.1: Analyzing a Crash (10 min)

**Topics:**
- Reading the panic stack trace
- Finding the failing input in the corpus
- Reproducing the crash deterministically
- Writing a targeted unit test from the fuzz discovery

**Workflow:**
```
1. Fuzzer finds crash -> saves input to testdata/
2. Read the input: cat testdata/fuzz/FuzzName/hash
3. Write unit test with that exact input
4. Fix the code (bounds check, nil check, etc.)
5. Run go test to verify fix
6. Corpus entry becomes permanent regression test
```

### Video 4.2: Common Fix Patterns (10 min)

**Topics:**
- Bounds checking before array/slice access
- Nil checking before pointer dereference
- Length validation before string operations
- NaN/Inf checking for float operations
- Size limits for recursive structures

**Fix Examples:**
```go
// Fix: bounds check
if len(parts) < 2 {
    return "", fmt.Errorf("invalid format")
}

// Fix: nil check
if resp == nil || resp.Body == nil {
    return nil, fmt.Errorf("nil response")
}

// Fix: length limit
if len(input) > maxInputLength {
    return nil, fmt.Errorf("input exceeds %d bytes", maxInputLength)
}

// Fix: NaN/Inf check
if math.IsNaN(score) || math.IsInf(score, 0) {
    return 0, fmt.Errorf("invalid score: %v", score)
}
```

---

## Module 5: HelixAgent's Fuzz Targets (15 min)

### Video 5.1: Walkthrough of All 7 Fuzz Targets (15 min)

**Topics:**
- FuzzJSONRequestParsing: Chat completion request robustness
- FuzzToolSchemaValidation: Tool schema edge cases
- FuzzSSEParsing: Server-Sent Events stream handling
- FuzzModelIDParsing: Provider/model format parsing
- FuzzHTTPHeaderParsing: Authorization header edge cases
- FuzzTokenBucket (Toolkit): Rate limiter algorithm
- FuzzDefaultCategoryInferrer (Toolkit): Model categorization

For each target:
- What it tests and why
- Key seed corpus entries
- Bugs previously discovered
- How to extend with new seeds

---

## Module 6: Hands-On Labs (30 min)

### Lab 1: Write a New Fuzz Target (10 min)

**Objective:** Write a fuzz test for a URL parsing function.

**Steps:**
1. Create `FuzzURLParsing` in a test file
2. Add 5 seed corpus entries (valid URLs, empty, malformed, very long)
3. Implement the fuzz function with invariant checks
4. Run for 30 seconds
5. Examine any discovered inputs

### Lab 2: Fix a Fuzz-Discovered Bug (10 min)

**Objective:** Given a fuzz failure, analyze, fix, and verify.

**Steps:**
1. Read the failing corpus entry
2. Reproduce with a unit test
3. Identify the root cause (missing bounds check)
4. Apply the fix
5. Verify the corpus entry passes

### Lab 3: Extended Fuzzing Campaign (10 min)

**Objective:** Run a 5-minute campaign on all targets and interpret results.

**Steps:**
1. Run all 5 fuzz targets for 1 minute each (resource-limited)
2. Record execs/sec and new interesting counts
3. Check if any crashes were found
4. Commit any new corpus entries

---

## Assessment

### Quiz (10 questions)

1. What does "coverage-guided" mean in the context of fuzzing?
2. What types does Go native fuzzing support as fuzz parameters?
3. Where are fuzz corpus files stored?
4. What is the difference between running with `-fuzz` and without?
5. How does the fuzzer decide which inputs are "interesting"?
6. Why is the seed corpus important for effective fuzzing?
7. What happens if a fuzz function panics?
8. How do you run fuzz tests with HelixAgent's resource limits?
9. What is corpus minimization and when should you use it?
10. How do fuzz discoveries become regression tests?

### Practical Assessment

1. Write a fuzz target for a new function (provided)
2. Seed it with at least 8 corpus entries
3. Run for 2 minutes and analyze results
4. If crashes found, fix them and verify
5. Run the corpus as regression tests
6. Document the findings

---

## Resources

- [User Manual 31: Fuzz Testing Guide](../user-manuals/31-fuzz-testing-guide.md)
- [User Manual 20: Testing Strategies](../user-manuals/20-testing-strategies.md)
- [Go Fuzz Testing Tutorial](https://go.dev/doc/security/fuzz/)
- [Go testing.F Documentation](https://pkg.go.dev/testing#F)
- [HelixAgent Fuzz Tests](../../tests/fuzz/)
- [Toolkit Fuzz Tests](../../Toolkit/pkg/toolkit/common/ratelimit/ratelimit_test.go)
