# Lab 6: Running Challenge Scripts

## Lab Overview

**Duration**: 90 minutes
**Difficulty**: Intermediate
**Module**: 12 - Challenge System and Validation

## Objectives

By completing this lab, you will:
- Run the RAGS, MCPS, and SKILLS challenges
- Understand strict real-result validation
- Interpret challenge reports and CSV results
- Debug failing challenge tests
- Achieve 100% pass rate on all challenges

## Prerequisites

- Labs 1-5 completed
- HelixAgent running with at least 3 providers configured
- All required API keys set up
- Terminal access with bash

---

## Exercise 1: Understanding the Challenge System (15 minutes)

### Task 1.1: Review Challenge Architecture

The HelixAgent Challenge System validates integration across three dimensions:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    CHALLENGE SYSTEM OVERVIEW                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   RAGS Challenge           MCPS Challenge          SKILLS Challenge │
│   ───────────────          ───────────────         ──────────────── │
│   RAG Integration          MCP Server              Skills           │
│   Validation               Integration             Integration      │
│                            Validation              Validation       │
│                                                                      │
│   - Cognee                 - 22 MCP Servers        - 21 Skills      │
│   - Qdrant                 - Protocol endpoints    - 8 Categories   │
│   - RAG Pipeline           - Tool Search           - CLI Agents     │
│   - Embeddings             - LSP/ACP               - Trigger Tests  │
│                                                                      │
├─────────────────────────────────────────────────────────────────────┤
│                     20+ CLI AGENTS TESTED                            │
│   OpenCode, ClaudeCode, Aider, Cline, HelixCode, Kiro, Forge...     │
└─────────────────────────────────────────────────────────────────────┘
```

### Task 1.2: Explore Challenge Scripts

```bash
# Navigate to challenge scripts
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent

# List available challenge scripts
ls -la challenges/scripts/

# View the RAGS challenge script structure
head -100 challenges/scripts/rags_challenge.sh

# View the MCPS challenge script structure
head -100 challenges/scripts/mcps_challenge.sh

# View the SKILLS challenge script structure
head -100 challenges/scripts/skills_challenge.sh
```

**Record the CLI agents tested**:
1. ____________
2. ____________
3. ____________
4. ____________
5. ____________

---

## Exercise 2: Running the RAGS Challenge (20 minutes)

### Task 2.1: Verify HelixAgent is Running

```bash
# Check HelixAgent health
curl http://localhost:7061/health | jq

# Check provider status
curl http://localhost:7061/v1/providers/status | jq
```

### Task 2.2: Run the RAGS Challenge

```bash
# Set environment variables (optional)
export HELIXAGENT_URL=http://localhost:7061
export TIMEOUT=60
export VERBOSE=true

# Run the RAGS challenge
./challenges/scripts/rags_challenge.sh
```

### Task 2.3: Analyze RAGS Results

```bash
# Find the latest results directory
LATEST_RAGS=$(ls -td challenges/results/rags_challenge/*/*/*/ | head -1)
echo "Latest results: $LATEST_RAGS"

# View the test results CSV
cat "${LATEST_RAGS}/test_results.csv"

# View the markdown report
cat "${LATEST_RAGS}/rags_challenge_report.md"
```

**Document Your Results**:
| Metric | Value |
|--------|-------|
| Total Tests | |
| Passed | |
| Failed | |
| Pass Rate | |

---

## Exercise 3: Running the MCPS Challenge (20 minutes)

### Task 3.1: Run the MCPS Challenge

```bash
# Run the MCPS challenge
./challenges/scripts/mcps_challenge.sh
```

### Task 3.2: Analyze MCPS Results

```bash
# Find the latest results directory
LATEST_MCPS=$(ls -td challenges/results/mcps_challenge/*/*/*/ | head -1)
echo "Latest results: $LATEST_MCPS"

# View test results
cat "${LATEST_MCPS}/test_results.csv" | head -50

# Count results by status
echo "=== Results Summary ==="
grep -c "^PASS" "${LATEST_MCPS}/test_results.csv" || echo "0 PASS"
grep -c "^FAIL" "${LATEST_MCPS}/test_results.csv" || echo "0 FAIL"
```

### Task 3.3: Understanding MCP Tool Search Tests

The MCPS challenge includes Section 9: MCP Tool Search Active Usage Validation. This section uses **strict validation** - it's not enough to get HTTP 200, the response must contain real results.

```bash
# Test MCP tool search manually
curl "http://localhost:7061/v1/mcp/tools/search?q=file" | jq

# Verify the response has real results
curl "http://localhost:7061/v1/mcp/tools/search?q=file" | jq '.count'
```

**Questions**:
1. How many MCP servers are tested? ____________
2. What protocols are tested besides MCP? ____________
3. What does "FALSE SUCCESS" mean? ____________

---

## Exercise 4: Running the SKILLS Challenge (20 minutes)

### Task 4.1: Run the SKILLS Challenge

```bash
# Run the SKILLS challenge
./challenges/scripts/skills_challenge.sh
```

### Task 4.2: Analyze SKILLS Results

```bash
# Find the latest results directory
LATEST_SKILLS=$(ls -td challenges/results/skills_challenge/*/*/*/ | head -1)
echo "Latest results: $LATEST_SKILLS"

# View the report
cat "${LATEST_SKILLS}/skills_challenge_report.md"
```

### Task 4.3: Understanding Skill Categories

The SKILLS challenge tests 21 skills across 8 categories:

| Category | Skills | Description |
|----------|--------|-------------|
| Code | generate, refactor, optimize | Code manipulation |
| Debug | trace, profile, analyze | Debugging and analysis |
| Search | find, grep, semantic-search | Code and file search |
| Git | commit, branch, merge | Version control |
| Deploy | build, deploy | Build and deployment |
| Docs | document, explain, readme | Documentation |
| Test | unit-test, integration-test | Test generation |
| Review | lint, security-scan | Code review |

**Record Skills Results by Category**:
| Category | Passed | Failed |
|----------|--------|--------|
| Code | | |
| Debug | | |
| Search | | |
| Git | | |
| Deploy | | |
| Docs | | |
| Test | | |
| Review | | |

---

## Exercise 5: Understanding Strict Validation (10 minutes)

### Task 5.1: What is Strict Real-Result Validation?

Strict validation ensures tests don't report false successes. Key checks:

1. **HTTP 200 is not enough**: Response must have real content
2. **Choices array verification**: Must have non-empty choices
3. **Content length check**: Response content > 50 characters
4. **Error detection**: Content must not start with "Error", "Failed", etc.

### Task 5.2: Examine Validation Code

```bash
# Look at the validation function in rags_challenge.sh
grep -A 30 "STRICT REAL-RESULT VALIDATION" challenges/scripts/rags_challenge.sh
```

**Key validation checks**:
```bash
# 1. Check response has choices array
has_choices=$(echo "$response_body" | grep -q '"choices"' && echo "yes" || echo "no")

# 2. Check response has actual content
content=$(echo "$response_body" | jq -r '.choices[0].message.content // ""')
content_length=${#content}

# 3. Verify content is not error message
is_real_content="no"
if [[ "$content_length" -gt 50 ]] && [[ ! "$content" =~ ^(Error|error:|Failed|null|undefined) ]]; then
    is_real_content="yes"
fi
```

### Task 5.3: Test Strict Validation Manually

```bash
# Make a test request
response=$(curl -s -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Write a function to calculate factorial"}],
    "max_tokens": 500
  }')

# Check for real content
echo "$response" | jq '.choices[0].message.content' | head -c 200

# Get content length
echo "$response" | jq -r '.choices[0].message.content' | wc -c
```

---

## Exercise 6: Debugging Failed Challenges (5 minutes)

### Task 6.1: Common Failure Causes

| Issue | Symptom | Solution |
|-------|---------|----------|
| HelixAgent not running | HTTP 000 | Start HelixAgent |
| Missing API keys | HTTP 401/403 | Configure keys in .env |
| Timeout | HTTP 000 | Increase TIMEOUT value |
| Provider unavailable | HTTP 503 | Check provider status |
| Empty responses | FALSE SUCCESS | Check LLM configuration |

### Task 6.2: Debugging Commands

```bash
# Check HelixAgent logs
docker-compose logs helixagent | tail -50

# Check specific provider
curl http://localhost:7061/v1/providers/deepseek/health | jq

# Run single challenge with verbose output
VERBOSE=true ./challenges/scripts/rags_challenge.sh 2>&1 | tee debug.log
```

---

## Lab Completion Checklist

- [ ] RAGS challenge completed
- [ ] MCPS challenge completed
- [ ] SKILLS challenge completed
- [ ] Analyzed all three reports
- [ ] Understood strict validation
- [ ] Documented pass rates

**Final Pass Rate Summary**:
| Challenge | Pass Rate | Status |
|-----------|-----------|--------|
| RAGS | ___% | PASS/FAIL |
| MCPS | ___% | PASS/FAIL |
| SKILLS | ___% | PASS/FAIL |

---

## Challenge Script Reference

| Challenge | Script | Description |
|-----------|--------|-------------|
| RAGS | `rags_challenge.sh` | RAG integration validation |
| MCPS | `mcps_challenge.sh` | MCP server integration |
| SKILLS | `skills_challenge.sh` | Skills integration |
| All | `run_all_challenges.sh` | Run all challenges |

---

## Troubleshooting

### "HelixAgent is not running"
```bash
# Start HelixAgent
make run-dev
# Or with Docker
docker-compose up -d
```

### "Timeout errors"
```bash
# Increase timeout
TIMEOUT=120 ./challenges/scripts/rags_challenge.sh
```

### "FALSE SUCCESS detected"
- Check LLM provider configuration
- Verify API keys are valid
- Check provider rate limits

---

## Challenge Exercise (Optional)

Create a script that:
1. Runs all three challenges
2. Aggregates results into a single report
3. Calculates overall pass rate
4. Fails if any challenge is below 95%

---

## Next Lab

Proceed to **Lab 7: MCP Tool Search** to learn tool discovery and integration.

---

*Lab Version: 1.0.0*
*Last Updated: January 2026*
