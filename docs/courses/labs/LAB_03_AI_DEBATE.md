# Lab 3: AI Debate Configuration

## Lab Overview

**Duration**: 75 minutes
**Difficulty**: Intermediate
**Module**: 6 - AI Debate System

## Objectives

By completing this lab, you will:
- Configure an AI debate with multiple participants
- Set up LLM fallback chains per participant
- Test different debate strategies
- Analyze consensus results

## Prerequisites

- Lab 2 completed
- At least 2 providers configured
- HelixAgent running with API keys

---

## Exercise 1: Understanding AI Debate (10 minutes)

### Task 1.1: Review Debate Architecture

The AI Debate system consists of:
- **5 Participant Roles**: Analyst, Proposer, Critic, Synthesizer, Mediator
- **Multiple Rounds**: Iterative discussion until consensus
- **Voting Strategies**: Various methods to reach agreement
- **LLM Fallbacks**: Each position has fallback LLMs

Draw the debate flow:

```
Topic → [Analyst] → [Proposer] → [Critic] → [Synthesizer] → [Mediator] → Consensus
            ↑__________________________________________|
                        (Multi-round if needed)
```

### Task 1.2: View Default Configuration

```bash
# Check if debate endpoint exists
curl http://localhost:7061/v1/debates | jq
```

---

## Exercise 2: Create Basic Debate (20 minutes)

### Task 2.1: Start a Simple Debate

```bash
# Create a debate
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Should remote work become the default for tech companies?",
    "rounds": 3,
    "style": "theater"
  }' | jq
```

**Record the response**:
- Debate ID: ____________
- Status: ____________
- Created At: ____________

### Task 2.2: Check Debate Status

```bash
# Replace with your debate ID
DEBATE_ID="debate-xxx"

curl http://localhost:7061/v1/debates/$DEBATE_ID/status | jq
```

### Task 2.3: Get Debate Results

```bash
# Wait for completion, then get results
curl http://localhost:7061/v1/debates/$DEBATE_ID | jq
```

**Document the results**:
| Metric | Value |
|--------|-------|
| Consensus Reached | |
| Rounds Completed | |
| Duration (ms) | |
| Confidence Score | |

---

## Exercise 3: Debate with Streaming (15 minutes)

### Task 3.1: Stream Debate Response

```bash
# Use streaming for real-time output
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -N \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "user", "content": "What is the best programming language for AI development?"}
    ],
    "stream": true
  }'
```

Watch the theatrical dialogue unfold:

```
╔══════════════════════════════════════════════════════════════════╗
║           HELIXAGENT AI DEBATE ENSEMBLE                          ║
╠══════════════════════════════════════════════════════════════════╣
║  Five AI minds deliberate to synthesize the optimal response.    ║
╚══════════════════════════════════════════════════════════════════╝

[A] THE ANALYST: "Let me analyze this systematically..."
[P] THE PROPOSER: "I propose we consider..."
...
```

### Task 3.2: Try Different Styles

```bash
# Novel style
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Is AI art real art?",
    "rounds": 2,
    "style": "novel"
  }'

# Screenplay style
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Will autonomous vehicles replace human drivers?",
    "rounds": 2,
    "style": "screenplay"
  }'

# Minimal style
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Cryptocurrency: revolution or speculation?",
    "rounds": 2,
    "style": "minimal"
  }'
```

**Compare styles**:
| Style | Best For |
|-------|----------|
| theater | |
| novel | |
| screenplay | |
| minimal | |

---

## Exercise 4: Custom Debate Configuration (20 minutes)

### Task 4.1: Create Debate Configuration File

Create `configs/my-debate.yaml`:

```yaml
ai_debate:
  enabled: true
  maximal_repeat_rounds: 5
  debate_timeout: 600000
  consensus_threshold: 0.75

  debate_strategy: structured
  voting_strategy: confidence_weighted
  response_format: detailed

  enable_memory: true
  memory_retention: 2592000000

  # Cognee enhancement (optional)
  cognee_config:
    enabled: false
    enhance_responses: true
    analyze_consensus: true

  # Participant configuration
  participants:
    - name: "PrimaryAnalyst"
      role: "Lead Analyst"
      enabled: true
      weight: 1.5
      priority: 1
      debate_style: analytical
      argumentation_style: logical
      persuasion_level: 0.7
      openness_to_change: 0.4
      response_timeout: 45000
      quality_threshold: 0.8
      min_response_length: 100
      max_response_length: 2000

      llms:
        - name: "Claude Primary"
          provider: claude
          model: claude-3-5-sonnet-20241022
          enabled: true
          temperature: 0.3
          max_tokens: 2000
          timeout: 45000
          max_retries: 2

        - name: "DeepSeek Fallback"
          provider: deepseek
          model: deepseek-chat
          enabled: true
          temperature: 0.3
          max_tokens: 2000
          timeout: 35000

    - name: "DevilsAdvocate"
      role: "Critical Analyst"
      enabled: true
      weight: 1.2
      priority: 2
      debate_style: aggressive
      argumentation_style: evidence_based
      persuasion_level: 0.9
      openness_to_change: 0.2

      llms:
        - name: "Gemini Primary"
          provider: gemini
          model: gemini-2.0-flash
          enabled: true
          temperature: 0.4
          max_tokens: 2000

    - name: "CreativeThinker"
      role: "Innovation Specialist"
      enabled: true
      weight: 1.0
      priority: 3
      debate_style: creative
      argumentation_style: hypothetical

      llms:
        - name: "DeepSeek Creative"
          provider: deepseek
          model: deepseek-chat
          enabled: true
          temperature: 0.7
          max_tokens: 2000
```

### Task 4.2: Test Custom Configuration

```bash
# Load and test the configuration
# (Implementation depends on how HelixAgent loads config)
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "How should companies balance AI innovation with ethical concerns?",
    "rounds": 4,
    "config_file": "my-debate.yaml"
  }' | jq
```

---

## Exercise 5: Analyze Debate Results (10 minutes)

### Task 5.1: Review Participant Contributions

```bash
# Get detailed debate results
curl http://localhost:7061/v1/debates/$DEBATE_ID | jq '.participants_summary'
```

**Fill in the analysis**:
| Participant | Responses | Avg Quality | Key Arguments |
|-------------|-----------|-------------|---------------|
| Analyst | | | |
| Critic | | | |
| Creative | | | |

### Task 5.2: Consensus Analysis

```bash
# Check consensus details
curl http://localhost:7061/v1/debates/$DEBATE_ID | jq '.consensus'
```

**Record**:
- Final Consensus: ____________
- Confidence Level: ____________
- Agreement Points: ____________
- Disagreement Points: ____________

### Task 5.3: View Metrics

```bash
# Get debate metrics
curl http://localhost:7061/v1/debate/metrics | jq
```

---

## Lab Completion Checklist

- [ ] Created first debate
- [ ] Monitored debate status
- [ ] Retrieved debate results
- [ ] Tested streaming output
- [ ] Compared different styles
- [ ] Created custom configuration
- [ ] Analyzed participant contributions
- [ ] Reviewed consensus results

---

## Debate Strategy Reference

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `round_robin` | Fixed turn order | Balanced discussions |
| `free_form` | Any participant anytime | Dynamic debates |
| `structured` | Organized rounds | Complex topics |
| `adversarial` | Opposing viewpoints | Devil's advocate |
| `collaborative` | Consensus building | Team decisions |

---

## Troubleshooting

### Debate Timeout
- Increase `debate_timeout`
- Reduce number of rounds
- Use faster providers

### No Consensus Reached
- Lower `consensus_threshold`
- Increase `maximal_repeat_rounds`
- Review participant weights

### Low Quality Responses
- Adjust `quality_threshold`
- Increase `min_response_length`
- Use higher-quality models

---

## Challenge Exercise (Optional)

Create a debate comparison script that:
1. Runs the same topic with different strategies
2. Compares consensus quality
3. Measures time to consensus
4. Reports which strategy worked best

---

## Next Lab

Proceed to **Lab 4: MCP Integration** to learn protocol integration.

---

*Lab Version: 1.0.0*
*Last Updated: January 2026*
