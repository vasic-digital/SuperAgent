# Lab 2: LLM Provider Configuration

## Lab Overview

**Duration**: 60 minutes
**Difficulty**: Intermediate
**Module**: 4 - LLM Provider Integration

## Objectives

By completing this lab, you will:
- Configure multiple LLM providers
- Test provider health checks
- Implement fallback chains
- Compare provider responses

## Prerequisites

- Lab 1 completed
- At least one API key (Claude, DeepSeek, or Gemini)
- HelixAgent running locally

---

## Exercise 1: Provider Configuration (15 minutes)

### Task 1.1: Configure Claude Provider

Edit your `.env` file:

```bash
# Add Claude API key
CLAUDE_API_KEY=your-claude-api-key-here
```

Verify configuration:

```bash
# Restart server to pick up changes
make run-dev

# Check provider status
curl http://localhost:7061/v1/providers/status | jq
```

**Checkpoint**:
- [ ] Claude shows as "verified" or "healthy"

### Task 1.2: Configure DeepSeek Provider

```bash
# Add to .env
DEEPSEEK_API_KEY=your-deepseek-api-key-here
```

### Task 1.3: Configure Gemini Provider

```bash
# Add to .env
GEMINI_API_KEY=your-gemini-api-key-here
```

**Document your configured providers**:
| Provider | Status | Score |
|----------|--------|-------|
| Claude | | |
| DeepSeek | | |
| Gemini | | |

---

## Exercise 2: Provider Health Checks (15 minutes)

### Task 2.1: Check All Provider Health

```bash
curl http://localhost:7061/v1/providers/health | jq
```

**Expected fields per provider**:
- `name`
- `status`
- `latency_ms`
- `last_check`

### Task 2.2: Check Individual Provider

```bash
# Check Claude health
curl http://localhost:7061/v1/providers/claude/health | jq

# Check DeepSeek health
curl http://localhost:7061/v1/providers/deepseek/health | jq
```

### Task 2.3: Verify Provider

```bash
# Manually trigger verification
curl -X POST http://localhost:7061/v1/providers/verify \
  -H "Content-Type: application/json" \
  -d '{"provider": "claude"}' | jq
```

**Record verification results**:
| Provider | Verified | Score | Response Time |
|----------|----------|-------|---------------|
| | | | |
| | | | |

---

## Exercise 3: Making Provider Requests (15 minutes)

### Task 3.1: Request to Specific Provider

```bash
# Request using Claude
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3.5-sonnet",
    "messages": [
      {"role": "user", "content": "What is 2+2? Reply with just the number."}
    ],
    "max_tokens": 10
  }' | jq
```

### Task 3.2: Compare Provider Responses

Create a test script `test_providers.sh`:

```bash
#!/bin/bash

PROMPT="Explain quantum computing in one sentence."

echo "=== Claude Response ==="
curl -s -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"claude-3.5-sonnet\",
    \"messages\": [{\"role\": \"user\", \"content\": \"$PROMPT\"}],
    \"max_tokens\": 100
  }" | jq '.choices[0].message.content'

echo ""
echo "=== DeepSeek Response ==="
curl -s -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"deepseek-chat\",
    \"messages\": [{\"role\": \"user\", \"content\": \"$PROMPT\"}],
    \"max_tokens\": 100
  }" | jq '.choices[0].message.content'

echo ""
echo "=== Gemini Response ==="
curl -s -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"gemini-2.0-flash\",
    \"messages\": [{\"role\": \"user\", \"content\": \"$PROMPT\"}],
    \"max_tokens\": 100
  }" | jq '.choices[0].message.content'
```

Run and compare:
```bash
chmod +x test_providers.sh
./test_providers.sh
```

---

## Exercise 4: Fallback Configuration (15 minutes)

### Task 4.1: Create Provider Config File

Create `configs/my-providers.yaml`:

```yaml
providers:
  claude:
    enabled: true
    api_key: ${CLAUDE_API_KEY}
    models:
      - claude-3-5-sonnet-20241022
      - claude-3-opus-20240229
    weight: 1.0
    timeout: 30s

  deepseek:
    enabled: true
    api_key: ${DEEPSEEK_API_KEY}
    models:
      - deepseek-chat
      - deepseek-coder
    weight: 0.8
    timeout: 30s

  gemini:
    enabled: true
    api_key: ${GEMINI_API_KEY}
    models:
      - gemini-2.0-flash
      - gemini-pro
    weight: 0.9
    timeout: 30s

fallback:
  chain:
    - claude
    - deepseek
    - gemini
  max_retries: 3
  retry_delay: 1s
```

### Task 4.2: Test Fallback Chain

```bash
# Use helixagent-debate which uses fallback chain
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "user", "content": "What is machine learning?"}
    ]
  }' | jq
```

### Task 4.3: Simulate Provider Failure

```bash
# Temporarily disable Claude (simulate by using invalid key)
# Then test if fallback to DeepSeek works

curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "user", "content": "Test fallback"}
    ]
  }' | jq '.model'
```

---

## Lab Completion Checklist

- [ ] At least 2 providers configured
- [ ] Provider health checks working
- [ ] Manual verification completed
- [ ] Made successful requests to each provider
- [ ] Compared responses across providers
- [ ] Fallback configuration created
- [ ] Fallback chain tested

---

## Provider Comparison Matrix

Fill in based on your observations:

| Aspect | Claude | DeepSeek | Gemini |
|--------|--------|----------|--------|
| Response Speed | | | |
| Quality | | | |
| Code Tasks | | | |
| Creative Tasks | | | |
| Cost | | | |

---

## Troubleshooting

### Provider Shows "Unhealthy"
- Verify API key is correct
- Check API key permissions
- Ensure billing is active
- Test with provider's own API directly

### Timeout Errors
```yaml
# Increase timeout in config
timeout: 60s
```

### Rate Limiting
- Implement backoff
- Use multiple providers
- Add rate limit configuration:
```yaml
rate_limit_rps: 5
```

---

## Challenge Exercise (Optional)

Create a benchmark script that:
1. Sends the same prompt to all providers
2. Measures response time
3. Compares response length
4. Reports results in a table

---

## Next Lab

Proceed to **Lab 3: AI Debate Configuration** to learn multi-agent debates.

---

*Lab Version: 1.0.0*
*Last Updated: January 2026*
