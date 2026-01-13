# HelixAgent Quick Reference Card

## Essential Commands

### Build & Run
```bash
make build              # Build binary
make run                # Run server
make run-dev            # Run in debug mode
docker-compose up -d    # Start with Docker
```

### Testing
```bash
make test               # All tests
make test-unit          # Unit tests only
make test-integration   # Integration tests
make test-coverage      # With coverage report
```

### Code Quality
```bash
make fmt                # Format code
make lint               # Run linter
make vet                # Static analysis
make security-scan      # Security check
```

---

## API Endpoints

### Health & Status
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/v1/models` | GET | List models |
| `/v1/providers/status` | GET | Provider status |
| `/v1/providers/health` | GET | Provider health |

### Chat Completions
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/chat/completions` | POST | Chat completion |
| `/v1/embeddings` | POST | Generate embeddings |

### AI Debate
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/debates` | POST | Create debate |
| `/v1/debates/:id` | GET | Get debate |
| `/v1/debates/:id/status` | GET | Debate status |
| `/v1/debates/:id` | DELETE | Cancel debate |

### MCP Protocol
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/mcp/tools` | GET | List MCP tools |
| `/v1/mcp/execute` | POST | Execute tool |
| `/v1/mcp/resources` | GET | List resources |

---

## Chat Completion Request

```json
{
  "model": "helixagent-debate",
  "messages": [
    {"role": "system", "content": "System prompt"},
    {"role": "user", "content": "User message"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false
}
```

---

## Environment Variables

### Required (at least one)
```bash
CLAUDE_API_KEY=your-key
DEEPSEEK_API_KEY=your-key
GEMINI_API_KEY=your-key
QWEN_API_KEY=your-key
```

### Server Configuration
```bash
PORT=7061
GIN_MODE=release
JWT_SECRET=your-secret
```

### Database
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=password
DB_NAME=helixagent
```

### Cache
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

---

## Supported Providers

| Provider | Models | Key Variable |
|----------|--------|--------------|
| Claude | claude-3.5-sonnet, claude-3-opus | `CLAUDE_API_KEY` |
| DeepSeek | deepseek-chat, deepseek-coder | `DEEPSEEK_API_KEY` |
| Gemini | gemini-2.0-flash, gemini-pro | `GEMINI_API_KEY` |
| Qwen | qwen-max, qwen-plus | `QWEN_API_KEY` |
| OpenRouter | Multiple | `OPENROUTER_API_KEY` |
| Mistral | mistral-large | `MISTRAL_API_KEY` |
| Cerebras | llama3.1-70b | `CEREBRAS_API_KEY` |
| Groq | llama-3.1-70b-versatile | `GROQ_API_KEY` |

---

## AI Debate Roles

| Role | Character | Purpose |
|------|-----------|---------|
| [A] | THE ANALYST | Systematic analysis |
| [P] | THE PROPOSER | Solution proposals |
| [C] | THE CRITIC | Challenge assumptions |
| [S] | THE SYNTHESIZER | Combine perspectives |
| [M] | THE MEDIATOR | Reach consensus |

---

## Debate Strategies

| Strategy | Description |
|----------|-------------|
| `round_robin` | Fixed turn order |
| `free_form` | Dynamic order |
| `structured` | Organized rounds |
| `adversarial` | Opposing views |
| `collaborative` | Build together |

---

## Voting Strategies

| Strategy | Description |
|----------|-------------|
| `majority` | Simple vote count |
| `weighted` | By participant weight |
| `consensus` | High agreement required |
| `confidence_weighted` | By confidence scores |
| `quality_weighted` | By quality scores |

---

## Debate Styles

| Style | Format |
|-------|--------|
| `theater` | Theatrical dialogue (default) |
| `novel` | Prose narration |
| `screenplay` | Script format |
| `minimal` | Plain text |

---

## Configuration Files

| File | Purpose |
|------|---------|
| `configs/development.yaml` | Dev settings |
| `configs/production.yaml` | Prod settings |
| `configs/multi-provider.yaml` | Multi-LLM setup |
| `configs/providers.yaml` | Provider config |

---

## Docker Profiles

```bash
docker-compose up -d                           # Core only
docker-compose --profile ai up -d              # + AI services
docker-compose --profile monitoring up -d      # + Monitoring
docker-compose --profile optimization up -d    # + LLM optimization
docker-compose --profile full up -d            # Everything
```

---

## Useful Curl Commands

### Health Check
```bash
curl http://localhost:7061/health
```

### Chat Completion
```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"helixagent-debate","messages":[{"role":"user","content":"Hello"}]}'
```

### Streaming
```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -N \
  -d '{"model":"helixagent-debate","messages":[{"role":"user","content":"Hello"}],"stream":true}'
```

### Create Debate
```bash
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{"topic":"Your topic","rounds":3}'
```

---

## Error Codes

| Code | Error | Resolution |
|------|-------|------------|
| 400 | Invalid request | Check JSON format |
| 401 | Auth failed | Verify API key |
| 403 | Access denied | Check permissions |
| 404 | Not found | Verify endpoint |
| 429 | Rate limited | Reduce requests |
| 500 | Server error | Check logs |

---

## Metrics Endpoints

```bash
# Prometheus metrics
curl http://localhost:7061/metrics

# Health checks
curl http://localhost:7061/healthz/live
curl http://localhost:7061/healthz/ready
```

---

*Quick Reference v1.0.0 | January 2026*
