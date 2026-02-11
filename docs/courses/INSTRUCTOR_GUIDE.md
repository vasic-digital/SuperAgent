# HelixAgent Course Instructor Guide

## Overview

This guide provides instructors with everything needed to deliver the HelixAgent training course effectively.

## Course Structure

### Duration
- Total: 14+ hours
- 14 modules (including 3 new advanced modules)
- 8 lab exercises (including 3 new challenge labs)
- 5 certification levels

### Delivery Options
1. **Self-paced online**: Video modules + labs
2. **Instructor-led virtual**: Live sessions + guided labs
3. **In-person workshop**: 3-day intensive (extended for new content)

---

## Module Delivery Guide

### Module 1: Introduction (45 min)

**Preparation**:
- Ensure demo environment is running
- Have architecture diagrams ready
- Prepare real-world examples

**Key Teaching Points**:
1. Explain multi-provider value proposition
2. Show vendor lock-in problems
3. Demonstrate live API call

**Demo Script**:
```bash
# Show health check
curl http://localhost:7061/health | jq

# Show available models
curl http://localhost:7061/v1/models | jq

# Make simple completion
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"helixagent-debate","messages":[{"role":"user","content":"Hello"}]}'
```

**Common Questions**:
- Q: "How is this different from LangChain?"
- A: HelixAgent focuses on provider orchestration and consensus, not chains.

- Q: "Do I need all providers configured?"
- A: No, just one provider minimum.

---

### Module 2: Installation (60 min)

**Preparation**:
- Test Docker setup before class
- Have Go installed for source build
- Prepare troubleshooting scenarios

**Lab Supervision**:
- Walk students through Docker setup
- Monitor for common errors
- Help with API key configuration

**Common Issues**:
1. Port 7061 already in use
2. Docker daemon not running
3. Missing Go installation
4. Invalid API keys

**Solutions**:
```bash
# Port issue
PORT=8080 make run-dev

# Docker check
docker info

# Go check
go version
```

---

### Module 3: Configuration (60 min)

**Preparation**:
- Have example config files ready
- Prepare environment variable cheat sheet
- Set up multiple configuration scenarios

**Teaching Points**:
1. Configuration hierarchy (env vars > config files > defaults)
2. Secrets management best practices
3. Environment-specific configs

**Interactive Exercise**:
Have students create a custom config that:
- Sets a custom port
- Configures 2 providers
- Adjusts rate limiting

---

### Module 4: Providers (75 min)

**Preparation**:
- Obtain API keys for demos (Claude, DeepSeek, Gemini)
- Prepare provider comparison matrix
- Set up fallback scenarios

**Demo Script**:
```bash
# Show provider status
curl http://localhost:7061/v1/providers/status | jq

# Compare responses
for MODEL in "claude-3.5-sonnet" "deepseek-chat" "gemini-2.0-flash"; do
  echo "=== $MODEL ==="
  curl -s -X POST http://localhost:7061/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d "{\"model\":\"$MODEL\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}" | jq '.choices[0].message.content'
done
```

**Discussion Topics**:
- Provider strengths/weaknesses
- Cost optimization strategies
- When to use which provider

---

### Module 5: Ensemble (60 min)

**Preparation**:
- Understand voting algorithms
- Prepare visual diagrams
- Have code examples ready

**Whiteboard Exercises**:
1. Draw voting flow diagram
2. Calculate weighted vote example
3. Compare strategy outcomes

**Code Walkthrough**:
Show `internal/llm/ensemble.go` to explain voting implementation.

---

### Module 6: AI Debate (90 min)

**This is the flagship module - allocate extra time**

**Preparation**:
- Have streaming demo ready
- Prepare different debate topics
- Test all debate styles

**Live Demo**:
```bash
# Create a debate and watch it unfold
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -N \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Should companies require employees to return to office?"}],
    "stream": true
  }'
```

**Discussion Points**:
1. When to use AI Debate vs single provider
2. Consensus building benefits
3. Production considerations

**Group Exercise**:
Divide class into groups. Each group:
1. Picks a debate topic
2. Configures participant roles
3. Runs debate and analyzes results
4. Presents findings to class

---

### Module 7: Plugins (75 min)

**Preparation**:
- Have sample plugin code ready
- Understand plugin lifecycle
- Prepare debugging examples

**Coding Exercise**:
Walk students through creating a simple plugin:
1. Create plugin structure
2. Implement interface
3. Register and test
4. Hot-reload demonstration

---

### Module 8: MCP/LSP (60 min)

**Preparation**:
- Understand protocol specifications
- Have MCP tools configured
- Prepare integration examples

**Demo Script**:
```bash
# List MCP tools
curl http://localhost:7061/v1/mcp/tools | jq

# Execute a tool
curl -X POST http://localhost:7061/v1/mcp/execute \
  -H "Content-Type: application/json" \
  -d '{"tool":"filesystem.read_file","arguments":{"path":"/etc/hostname"}}'
```

---

### Module 9: Optimization (75 min)

**Preparation**:
- Start optimization Docker services
- Understand caching mechanisms
- Prepare benchmark examples

**Demo Topics**:
1. Semantic caching hit rates
2. Structured output validation
3. Streaming performance

---

### Module 10: Security (60 min)

**Preparation**:
- Review security best practices
- Prepare vulnerability examples
- Have security scan ready

**Security Demo**:
```bash
# Run security scan
make security-scan

# Show rate limiting
for i in {1..20}; do
  curl -s http://localhost:7061/health > /dev/null
  echo "Request $i completed"
done
```

---

### Module 11: Testing & CI/CD (75 min)

**Preparation**:
- Have test infrastructure running
- Prepare CI/CD pipeline examples
- Set up coverage reports

**Demo Script**:
```bash
# Run all tests
make test

# Show coverage
make test-coverage
open coverage.html

# Run specific test types
make test-unit
make test-integration
```

---

### Module 12: Challenge System and Validation (90 min)

**This is a NEW flagship module - demonstrates 100% test pass rate methodology**

**Preparation**:
- Have HelixAgent running and healthy
- Ensure all API keys are configured
- Prepare challenge scripts directory
- Have sample challenge results ready

**Key Teaching Points**:
1. Challenge System Architecture
   - RAGS (RAG Integration Validation)
   - MCPS (MCP Server Integration Validation)
   - SKILLS (Skills Integration Validation)
2. Strict Real-Result Validation (no FALSE SUCCESSES)
3. 20+ CLI Agent testing across all endpoints

**Demo Script**:
```bash
# Run the RAGS challenge
./challenges/scripts/rags_challenge.sh

# Run the MCPS challenge
./challenges/scripts/mcps_challenge.sh

# Run the SKILLS challenge
./challenges/scripts/skills_challenge.sh

# Run all challenges at once
./challenges/scripts/run_all_challenges.sh
```

**Key Concepts to Emphasize**:
1. **Strict Validation**: HTTP 200 is NOT enough - must verify actual content
2. **FALSE SUCCESS Detection**: Check for empty responses, error messages
3. **Real Content Verification**: Content length > 50 chars, valid choices array
4. **CLI Agent Headers**: X-CLI-Agent header for agent identification

**Discussion Topics**:
- Why strict validation matters for production systems
- How to debug failing challenges
- Interpreting challenge reports and CSV results

---

### Module 13: MCP Tool Search and Discovery (60 min)

**Preparation**:
- Have MCP servers configured
- Prepare search query examples
- Understand tool registry structure

**Demo Script**:
```bash
# Search for file-related tools
curl http://localhost:7061/v1/mcp/tools/search?q=file | jq

# Search for git tools
curl http://localhost:7061/v1/mcp/tools/search?q=git | jq

# Get tool suggestions for a prompt
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=list%20files" | jq

# Search for adapters
curl http://localhost:7061/v1/mcp/adapters/search?q=github | jq

# Get tool categories
curl http://localhost:7061/v1/mcp/categories | jq

# Get MCP statistics
curl http://localhost:7061/v1/mcp/stats | jq
```

**Key Teaching Points**:
1. Tool search vs adapter search
2. AI-powered tool suggestions
3. Category-based filtering
4. Real-time tool discovery during chat

---

### Module 14: AI Debate System Advanced (90 min)

**This is an ADVANCED module - extends Module 6**

**Preparation**:
- Have all 10 LLM providers configured
- Understand LLMsVerifier scoring
- Prepare multi-pass validation examples

**Demo Script**:
```bash
# Create a debate with multi-pass validation
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Should AI development be open source?",
    "rounds": 3,
    "style": "theater",
    "enable_multi_pass_validation": true,
    "validation_config": {
      "enable_validation": true,
      "enable_polish": true,
      "validation_timeout": 120,
      "polish_timeout": 60,
      "min_confidence_to_skip": 0.9,
      "max_validation_rounds": 3,
      "show_phase_indicators": true
    }
  }' | jq
```

**Key Teaching Points**:
1. **25 LLM Team**: 5 positions x 5 LLMs (primary + 4 fallbacks)
2. **Multi-Pass Validation**: 4 phases for quality improvement
3. **LLMsVerifier Scoring**: 5-component weighted algorithm
4. **OAuth Priority**: Claude and Qwen get selection priority

**Whiteboard Exercises**:
1. Draw the 25 LLM team configuration
2. Map the 4 validation phases
3. Calculate weighted scores for providers

---

## Lab Supervision Guide

### Lab 1: Getting Started
- **Duration**: 45 min
- **Key Checkpoint**: Health endpoint responding
- **Common Issue**: Port conflicts

### Lab 2: Provider Setup
- **Duration**: 60 min
- **Key Checkpoint**: Multiple providers healthy
- **Common Issue**: Invalid API keys

### Lab 3: AI Debate
- **Duration**: 75 min
- **Key Checkpoint**: Successful debate completion
- **Common Issue**: Timeout errors

### Lab 4: MCP Integration
- **Duration**: 60 min
- **Key Checkpoint**: Tool execution working
- **Common Issue**: Missing MCP servers

### Lab 5: Production Deployment
- **Duration**: 120 min
- **Key Checkpoint**: Docker stack running
- **Common Issue**: Resource constraints

### Lab 6: Running Challenge Scripts (NEW)
- **Duration**: 90 min
- **Key Checkpoint**: All three challenges pass (RAGS, MCPS, SKILLS)
- **Common Issues**:
  - HelixAgent not running
  - Missing API keys
  - Timeout issues with large test suites
- **Success Criteria**: 100% pass rate on all challenges

**Troubleshooting Guide**:
```bash
# If challenges fail to start
curl http://localhost:7061/health

# If timeout errors occur
TIMEOUT=120 ./challenges/scripts/rags_challenge.sh

# Check challenge results
ls -la challenges/results/*/
cat challenges/results/*/test_results.csv
```

### Lab 7: MCP Tool Search Integration (NEW)
- **Duration**: 60 min
- **Key Checkpoint**: Tool search returns real results
- **Common Issues**:
  - Empty search results (check tool registry)
  - Adapter search not finding expected adapters
- **Success Criteria**: Search queries return valid tool matches

### Lab 8: Multi-Pass Validation Debate (NEW)
- **Duration**: 75 min
- **Key Checkpoint**: 4-phase validation completes successfully
- **Common Issues**:
  - Validation timeout
  - Low confidence scores
  - Provider fallback issues
- **Success Criteria**: Overall confidence > 0.8

---

## Assessment Administration

### Quiz Guidelines
1. No open book during quizzes
2. 80% passing score required
3. One retry allowed per module
4. Proctor must verify identity

### Practical Assessment
1. Students complete in controlled environment
2. Instructor reviews submitted configuration
3. Working demo required for pass

---

## Troubleshooting Common Issues

### "Server won't start"
```bash
# Check port
lsof -i :7061

# Check dependencies
docker-compose ps

# Check logs
docker-compose logs helixagent
```

### "API key not working"
- Verify key is active in provider dashboard
- Check environment variable is set
- Restart server after config change

### "Debate times out"
- Increase timeout in config
- Reduce number of rounds
- Check provider health

### "Low quality responses"
- Check temperature setting
- Verify model selection
- Review system prompts

---

## Slide Timing Guide

| Module | Slides | Time/Slide | Total |
|--------|--------|------------|-------|
| 1 | 28 | ~1.5 min | 45 min |
| 2 | 28 | ~2 min | 60 min |
| 3 | 28 | ~2 min | 60 min |
| 4 | 28 | ~2.5 min | 75 min |
| 5 | 28 | ~2 min | 60 min |
| 6 | 28 | ~3 min | 90 min |
| 7 | 28 | ~2.5 min | 75 min |
| 8 | 28 | ~2 min | 60 min |
| 9 | 28 | ~2.5 min | 75 min |
| 10 | 28 | ~2 min | 60 min |
| 11 | 28 | ~2.5 min | 75 min |
| 12 | 35 | ~2.5 min | 90 min |
| 13 | 24 | ~2.5 min | 60 min |
| 14 | 35 | ~2.5 min | 90 min |

---

## Post-Course Resources

### For Students
- Access to course materials for 1 year
- Community Discord channel
- GitHub repository access
- Office hours schedule

### For Instructors
- Monthly instructor meetup
- Course update notifications
- Feedback submission process
- Teaching tip newsletter

---

## Certification Process

### Level 1: Fundamentals
- Quiz Modules 1-3 (80%+)
- Lab 1 completion

### Level 2: Provider Expert
- Quiz Modules 4-6 (80%+)
- Labs 2-3 completion
- Implementation project

### Level 3: Advanced
- Quiz Modules 7-9 (80%+)
- Custom plugin submission
- Labs 4 completion

### Level 4: Master
- Quiz Modules 10-11 (80%+)
- Lab 5 completion
- Production deployment review

### Level 5: Challenge Expert (NEW)
- Quiz Modules 12-14 (80%+)
- Labs 6-8 completion
- **Special Requirements**:
  - 100% pass rate on RAGS challenge
  - 100% pass rate on MCPS challenge
  - 100% pass rate on SKILLS challenge
  - Demonstration of MCP Tool Search integration
  - Multi-pass validation debate with >0.8 confidence

---

## Feedback Collection

After each module, collect:
1. Content clarity rating (1-5)
2. Lab difficulty rating (1-5)
3. Instructor effectiveness (1-5)
4. Suggestions for improvement

Use Google Forms or similar for collection.

---

## Contact & Support

- **Technical Issues**: support@helixagent.ai
- **Course Feedback**: training@helixagent.ai
- **Certification**: certs@helixagent.ai
- **GitHub Issues**: https://github.com/your-org/helix-agent/issues

---

*Instructor Guide Version: 2.0.0*
*Last Updated: January 2026*
*New Content: Modules 12-14 (Challenge System, MCP Tool Search, Advanced AI Debate)*
