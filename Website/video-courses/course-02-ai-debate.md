# AI Debate System Mastery - Complete Video Course Script

**Total Duration: 90 minutes**
**Level: Intermediate**
**Prerequisites: Completion of Course 1 (SuperAgent Fundamentals)**

---

## Module 1: Understanding AI Debate (15 minutes)

### Opening Slide
**Title:** Mastering the AI Debate System
**Duration:** 30 seconds

---

### Section 1.1: What is AI Debate? (4 minutes)

#### Narration Script:

Welcome to the AI Debate System Mastery course. In this course, we'll explore one of SuperAgent's most innovative features - the ability to orchestrate structured debates between multiple AI participants.

AI Debate is a multi-agent collaboration system where different AI models take on distinct roles and engage in thoughtful discussion on complex topics. Unlike simple ensemble approaches that just pick the best response, AI Debate allows models to build on each other's ideas, challenge assumptions, and reach well-reasoned conclusions through iterative refinement.

Think of it like a roundtable discussion where experts from different backgrounds each contribute their unique perspectives. The result is often more nuanced and comprehensive than what any single AI could produce.

#### Key Points to Cover:
- Multi-agent orchestration concept
- Difference from simple ensemble approaches
- Role-based participation
- Iterative refinement through rounds
- Consensus building mechanisms

#### Slide Content:
```
WHAT IS AI DEBATE?

[Traditional Approach]
Query --> Model --> Response

[Ensemble Approach]
Query --> Model A --> Response A
      --> Model B --> Response B --> Pick Best
      --> Model C --> Response C

[AI Debate Approach]
Query --> Round 1: Initial Positions
      --> Round 2: Challenges & Refinement
      --> Round 3: Synthesis & Consensus
      --> Final: Combined Wisdom

Key Benefits:
- Diverse perspectives
- Error checking between models
- Deeper analysis
- More robust conclusions
```

---

### Section 1.2: Multi-Agent Collaboration (4 minutes)

#### Narration Script:

The power of AI Debate comes from structured collaboration. Each participant in a debate has access to what others have said, allowing them to respond, refine, and build upon previous arguments.

In a typical debate, we might have a Proposer who presents the main argument, a Critic who identifies weaknesses, and a Mediator who synthesizes the discussion. Each model brings its own strengths - perhaps Claude excels at nuanced reasoning while DeepSeek provides strong technical analysis.

The debate proceeds through rounds. In round one, participants present their initial positions. In subsequent rounds, they respond to each other, refine their arguments, and work toward consensus.

#### Key Points to Cover:
- Participant roles and responsibilities
- Cross-model communication
- Round-based progression
- Building on previous responses
- Synthesis of perspectives

#### Slide Content:
```
MULTI-AGENT COLLABORATION

[Debate Flow]
Round 1: Opening Arguments
         Each participant presents initial position
              |
              v
Round 2: Cross-Examination
         Respond to others, challenge assumptions
              |
              v
Round 3: Refinement
         Incorporate feedback, strengthen arguments
              |
              v
Final:   Consensus Analysis
         Identify agreements, synthesize conclusions

[Participant Roles]
- Proposer: Advocates for a position
- Critic: Challenges and tests arguments
- Analyst: Deep technical analysis
- Mediator: Synthesizes and finds common ground
```

---

### Section 1.3: Consensus Building (4 minutes)

#### Narration Script:

Consensus building is how AI Debate determines the final outcome. SuperAgent analyzes all responses across all rounds to identify areas of agreement, measure the strength of different arguments, and produce a final position that represents the collective wisdom.

The consensus algorithm considers several factors: text similarity between responses to identify agreement, quality scores for each response, the confidence levels reported by each model, and the coherence of arguments presented.

Early consensus detection is also built in. If participants strongly agree early in the debate, we can conclude early and save time. This is particularly useful when the topic has a clear answer that multiple models recognize.

#### Key Points to Cover:
- Agreement score calculation
- Quality-weighted consensus
- Early consensus detection
- Disagreement identification
- Final position synthesis

#### Code Example - Consensus Result:
```json
{
  "consensus": {
    "reached": true,
    "confidence": 0.87,
    "agreement_level": 0.82,
    "final_position": "Consensus reached: Participants agree that...",
    "key_points": [
      "Point 1 all participants emphasized",
      "Point 2 with strong agreement",
      "Point 3 after refinement"
    ],
    "disagreements": [
      "Minor disagreement on implementation approach"
    ],
    "voting_summary": {
      "strategy": "confidence_weighted",
      "total_votes": 9,
      "vote_distribution": {
        "Analyst": 3,
        "Proposer": 3,
        "Critic": 3
      },
      "winner": "Analyst"
    }
  }
}
```

---

### Section 1.4: Quality Scoring (3 minutes)

#### Narration Script:

Every response in a debate receives a quality score. This score influences both the consensus calculation and helps identify the best contributions. The quality scoring system considers multiple factors.

First, coherence - does the response have good structure and logical flow? We look for transition words, clear organization, and appropriate length.

Second, completeness - did the response fully address the topic? We check the finish reason and whether the response was cut off.

Third, relevance - how well did the response engage with the debate topic and previous participants?

Finally, confidence - the model's own reported confidence in its response.

#### Key Points to Cover:
- Coherence analysis
- Completeness checking
- Relevance scoring
- Confidence weighting
- Combined quality score

#### Slide Content:
```
QUALITY SCORING COMPONENTS

[Score Calculation]
Quality Score =
  (Confidence × 0.30) +
  (Length Factor × 0.20) +
  (Completeness × 0.20) +
  (Token Efficiency × 0.15) +
  (Coherence × 0.15)

[Coherence Indicators]
- Structure words: "first", "however", "therefore"
- Logical transitions
- Sentence variety
- Clear conclusions

[Quality Thresholds]
0.8+ : Excellent response
0.6-0.8 : Good response
0.4-0.6 : Acceptable
<0.4 : May need review
```

---

## Module 2: Configuring Debate Participants (20 minutes)

### Section 2.1: Participant Setup (6 minutes)

#### Narration Script:

Now let's dive into configuring debate participants. Each participant is defined with a name, role, and the LLM provider that will power their responses. The configuration is flexible, allowing you to mix different providers and roles.

Let me show you the basic participant configuration structure.

#### Code Example - Basic Participant Configuration:
```yaml
# ai-debate-config.yaml
participants:
  - name: "Technical Analyst"
    participant_id: "analyst-001"
    role: "analyst"
    llm_provider: "claude"
    llm_model: "claude-3-sonnet-20240229"
    weight: 1.2
    timeout: 30000  # 30 seconds in milliseconds

  - name: "Critical Reviewer"
    participant_id: "critic-001"
    role: "critic"
    llm_provider: "gemini"
    llm_model: "gemini-pro"
    weight: 1.0
    timeout: 30000

  - name: "Solution Proposer"
    participant_id: "proposer-001"
    role: "proposer"
    llm_provider: "deepseek"
    llm_model: "deepseek-coder"
    weight: 1.0
    timeout: 30000
```

#### API Configuration:
```bash
curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "What is the best approach to implement caching in a microservices architecture?",
    "participants": [
      {
        "name": "Technical Analyst",
        "role": "analyst",
        "llm_provider": "claude",
        "llm_model": "claude-3-sonnet-20240229",
        "weight": 1.2
      },
      {
        "name": "Critical Reviewer",
        "role": "critic",
        "llm_provider": "gemini",
        "llm_model": "gemini-pro",
        "weight": 1.0
      }
    ],
    "max_rounds": 3,
    "timeout": 300
  }'
```

#### Key Points to Cover:
- Participant identification (name, ID)
- Role assignment
- Provider and model selection
- Weight for consensus influence
- Timeout configuration

---

### Section 2.2: Role Configuration (6 minutes)

#### Narration Script:

Roles define how each participant approaches the debate. SuperAgent includes several predefined roles, each with a specific system prompt that guides the model's behavior. Understanding these roles helps you design effective debates.

#### Predefined Roles:

**Proposer**
```
System Prompt: "You are presenting and defending the main argument.
Be persuasive and provide evidence."
```
- Advocates for a specific position
- Provides supporting evidence
- Defends against criticism

**Critic / Opponent**
```
System Prompt: "You are challenging the main argument.
Identify weaknesses and present counterarguments."
```
- Questions assumptions
- Identifies logical flaws
- Presents alternatives

**Analyst**
```
System Prompt: "You are analyzing the debate topic deeply.
Provide insights and analysis."
```
- Deep technical analysis
- Data-driven perspectives
- Neutral assessment

**Mediator**
```
System Prompt: "You are facilitating the discussion.
Summarize key points and seek common ground."
```
- Synthesizes viewpoints
- Identifies agreement
- Proposes compromises

#### Custom Role Configuration:
```yaml
participants:
  - name: "Security Expert"
    role: "analyst"
    custom_system_prompt: |
      You are a cybersecurity expert participating in an AI debate.
      Focus on security implications of the discussed topic.
      Consider threat models, vulnerabilities, and best practices.
      Back your arguments with references to security standards and frameworks.
    llm_provider: "claude"
    llm_model: "claude-3-sonnet-20240229"
```

---

### Section 2.3: LLM Selection Strategy (5 minutes)

#### Narration Script:

Choosing which LLM to use for each participant is a strategic decision. Different models have different strengths, and matching the right model to the right role can significantly improve debate quality.

#### Model Strengths Guide:

| Provider | Strengths | Best Used For |
|----------|-----------|---------------|
| Claude | Nuanced reasoning, careful analysis | Mediator, Complex analysis |
| DeepSeek | Code generation, technical depth | Technical analyst, Code review |
| Gemini | Multimodal, broad knowledge | General analyst, Research |
| Qwen | Multilingual, fast | Quick responses, Translation |
| Grok (OpenRouter) | Real-time data, creative | Creative roles, Current events |

#### Example - Strategic LLM Assignment:
```yaml
# For a code architecture debate
participants:
  - name: "System Architect"
    role: "proposer"
    llm_provider: "claude"  # Strong reasoning for architecture
    llm_model: "claude-3-opus-20240229"
    weight: 1.3

  - name: "Implementation Expert"
    role: "analyst"
    llm_provider: "deepseek"  # Best for code analysis
    llm_model: "deepseek-coder"
    weight: 1.2

  - name: "Code Reviewer"
    role: "critic"
    llm_provider: "gemini"  # Different perspective
    llm_model: "gemini-pro"
    weight: 1.0
```

---

### Section 2.4: Weight Assignment (3 minutes)

#### Narration Script:

Participant weights influence how much each participant's responses contribute to the final consensus. Higher weights mean more influence. Use weights strategically to give more authority to domain experts or trusted models.

#### Weight Guidelines:
```yaml
# Weight values and their meaning
weights:
  1.5: Primary expert - maximum influence
  1.3: Subject matter expert - high influence
  1.0: Standard participant - normal influence
  0.8: Supporting voice - reduced influence
  0.5: Minority perspective - limited influence
```

#### Dynamic Weight Adjustment:
```bash
# Update participant weight during debate
curl -X PATCH http://localhost:8080/v1/debates/{debate_id}/participants/{participant_id} \
  -H "Content-Type: application/json" \
  -d '{
    "weight": 1.3
  }'
```

---

## Module 3: Advanced Debate Techniques (25 minutes)

### Section 3.1: Multi-Round Discussions (7 minutes)

#### Narration Script:

The real power of AI Debate emerges in multi-round discussions. In the first round, participants present their initial positions without knowledge of others' responses. In subsequent rounds, each participant receives the full context of previous responses and can engage directly with what others have said.

Let me walk you through how the rounds work and how to configure them effectively.

#### Code Example - Multi-Round Configuration:
```yaml
debate_config:
  topic: "Should microservices always be preferred over monolithic architectures?"
  max_rounds: 4
  timeout: 600000  # 10 minutes total

  round_configs:
    round_1:
      type: "opening"
      description: "Initial position statements"
      time_limit: 60000  # 1 minute per response

    round_2:
      type: "cross_examination"
      description: "Respond to others' arguments"
      time_limit: 90000

    round_3:
      type: "rebuttal"
      description: "Address criticisms and strengthen arguments"
      time_limit: 90000

    round_4:
      type: "closing"
      description: "Final synthesis and conclusions"
      time_limit: 60000
```

#### Demo - Watching a Debate Unfold:

**Round 1 Prompt (Proposer):**
```
DEBATE TOPIC: Should microservices always be preferred over monolithic architectures?

ROUND: 1
YOUR ROLE: Solution Proposer (proposer)

This is the opening round. Present your initial position on the topic.

Your response:
```

**Round 2 Prompt (same Proposer):**
```
DEBATE TOPIC: Should microservices always be preferred over monolithic architectures?

ROUND: 2
YOUR ROLE: Solution Proposer (proposer)

PREVIOUS RESPONSES:
-------------------
[Technical Analyst (analyst) - Round 1]:
When evaluating architecture choices, we must consider scale, team size...

[Critical Reviewer (critic) - Round 1]:
The premise of "always preferred" is flawed. Microservices introduce...

[Solution Proposer (proposer) - Round 1]:
Microservices offer significant advantages in modern cloud environments...
-------------------

Based on the previous responses, provide your perspective on the topic.
Address points raised by others and advance the discussion.

Your response:
```

#### Key Points to Cover:
- Round progression mechanics
- Context accumulation
- Response building
- Time management
- Early termination conditions

---

### Section 3.2: Memory Integration (6 minutes)

#### Narration Script:

SuperAgent can integrate persistent memory into debates, allowing participants to recall relevant information from previous conversations or a knowledge base. This is particularly powerful for ongoing projects where context builds over time.

#### Code Example - Memory-Enhanced Debate:
```yaml
debate_config:
  topic: "How should we improve our API rate limiting strategy?"
  enable_memory: true
  memory_config:
    retention_days: 30
    max_context_tokens: 8000
    relevance_threshold: 0.7

  # Memory sources
  memory_sources:
    - type: "conversation_history"
      session_id: "project-api-redesign"

    - type: "document_store"
      collection: "api-documentation"

    - type: "previous_debates"
      tags: ["api", "rate-limiting", "performance"]
```

#### API Call with Memory:
```bash
curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "How should we improve our API rate limiting strategy?",
    "participants": [...],
    "enable_memory": true,
    "memory_context": {
      "session_id": "project-api-redesign",
      "include_previous_debates": true,
      "max_context_length": 8000
    },
    "metadata": {
      "project": "API v2",
      "sprint": "Sprint 14"
    }
  }'
```

---

### Section 3.3: Cognee Enhancement (7 minutes)

#### Narration Script:

Cognee integration takes debates to the next level by providing semantic analysis, entity extraction, knowledge graphs, and enhanced insights. When enabled, Cognee analyzes each response and the overall debate to provide deeper understanding.

#### Code Example - Enabling Cognee:
```yaml
debate_config:
  enable_cognee: true
  cognee_settings:
    enhance_responses: true
    analyze_sentiment: true
    extract_entities: true
    generate_summary: true
    build_knowledge_graph: true
    dataset_name: "project_debates"
```

#### Cognee-Enhanced Response:
```json
{
  "participant_response": {
    "participant_name": "Technical Analyst",
    "content": "When implementing caching...",
    "quality_score": 0.87,

    "cognee_analysis": {
      "enhanced": true,
      "sentiment": "neutral",
      "entities": ["caching", "Redis", "TTL", "invalidation"],
      "key_phrases": [
        "distributed cache",
        "cache coherency",
        "time-to-live policies"
      ],
      "confidence": 0.89
    }
  }
}
```

#### Full Cognee Insights Example:
```json
{
  "cognee_insights": {
    "dataset_name": "debate-insights",
    "enhancement_time_ms": 1234,

    "semantic_analysis": {
      "main_themes": ["caching", "microservices", "performance"],
      "coherence_score": 0.85
    },

    "entity_extraction": [
      {"text": "caching strategies", "type": "CONCEPT", "confidence": 0.95},
      {"text": "Redis", "type": "TECHNOLOGY", "confidence": 0.98},
      {"text": "distributed systems", "type": "CONCEPT", "confidence": 0.92}
    ],

    "sentiment_analysis": {
      "overall_sentiment": "positive",
      "sentiment_score": 0.72
    },

    "knowledge_graph": {
      "nodes": [
        {"id": "topic-1", "label": "Caching Strategy", "type": "topic"},
        {"id": "tech-1", "label": "Redis", "type": "technology"},
        {"id": "tech-2", "label": "Memcached", "type": "technology"}
      ],
      "edges": [
        {"source": "tech-1", "target": "topic-1", "type": "implements"},
        {"source": "tech-2", "target": "topic-1", "type": "implements"}
      ]
    },

    "recommendations": [
      "Consider cache invalidation strategies",
      "Evaluate Redis cluster for high availability",
      "Implement circuit breakers for cache failures"
    ],

    "quality_metrics": {
      "coherence": 0.85,
      "relevance": 0.88,
      "accuracy": 0.82,
      "completeness": 0.79,
      "overall_score": 0.84
    }
  }
}
```

---

### Section 3.4: Custom Strategies (5 minutes)

#### Narration Script:

SuperAgent supports different debate and voting strategies. The debate strategy determines how participants interact, while the voting strategy determines how the final consensus is calculated.

#### Debate Strategies:

**Round Robin** (default)
```yaml
debate_strategy: "round_robin"
# Each participant speaks once per round
# Responses presented in order
```

**Structured**
```yaml
debate_strategy: "structured"
# Opening -> Cross-examination -> Rebuttal -> Closing
# More formal debate format
```

**Adversarial**
```yaml
debate_strategy: "adversarial"
# Participants assigned opposing positions
# Focus on challenge and defense
```

**Collaborative**
```yaml
debate_strategy: "collaborative"
# Participants work together to build solution
# No opposing positions
```

#### Voting Strategies:

```yaml
voting_strategies:
  majority:
    description: "Simple majority of participants"
    use_case: "Clear yes/no decisions"

  weighted:
    description: "Participant weights influence outcome"
    use_case: "Expert-weighted decisions"

  consensus:
    description: "Require agreement above threshold"
    use_case: "Important decisions requiring buy-in"
    threshold: 0.75

  confidence_weighted:
    description: "Weight by model confidence scores"
    use_case: "Quality-focused outcomes"

  quality_weighted:
    description: "Weight by response quality scores"
    use_case: "Best argument wins"
```

#### Code Example - Custom Strategy Configuration:
```bash
curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Evaluate security implications of using third-party APIs",
    "participants": [...],
    "debate_strategy": "adversarial",
    "voting_strategy": "quality_weighted",
    "consensus_threshold": 0.7,
    "max_rounds": 3
  }'
```

---

## Module 4: Monitoring and Optimization (30 minutes)

### Section 4.1: Performance Metrics (8 minutes)

#### Narration Script:

Monitoring debate performance is essential for optimization. SuperAgent provides comprehensive metrics for each debate, including timing information, quality scores, and provider performance.

#### Key Metrics Dashboard:

```json
{
  "debate_metrics": {
    "debate_id": "debate-abc123",
    "status": "completed",

    "timing": {
      "total_duration_ms": 45230,
      "avg_response_time_ms": 3421,
      "per_round_duration_ms": [12500, 15200, 17530],
      "per_participant_avg_ms": {
        "Technical Analyst": 2890,
        "Critical Reviewer": 3156,
        "Solution Proposer": 4217
      }
    },

    "quality": {
      "overall_score": 0.84,
      "consensus_confidence": 0.87,
      "per_participant_scores": {
        "Technical Analyst": 0.89,
        "Critical Reviewer": 0.82,
        "Solution Proposer": 0.81
      },
      "per_round_scores": [0.78, 0.85, 0.89]
    },

    "engagement": {
      "total_responses": 9,
      "cross_references": 12,
      "argument_refinements": 7
    },

    "tokens": {
      "total_tokens_used": 15432,
      "per_provider": {
        "claude": 5890,
        "gemini": 4321,
        "deepseek": 5221
      }
    }
  }
}
```

#### Prometheus Metrics:
```
# Debate timing metrics
superagent_debate_duration_seconds{debate_id="abc123"} 45.23
superagent_debate_round_duration_seconds{debate_id="abc123",round="1"} 12.5
superagent_debate_round_duration_seconds{debate_id="abc123",round="2"} 15.2

# Quality metrics
superagent_debate_quality_score{debate_id="abc123"} 0.84
superagent_debate_consensus_level{debate_id="abc123"} 0.87

# Provider performance
superagent_provider_response_time_seconds{provider="claude"} 2.89
superagent_provider_success_rate{provider="claude"} 0.98
```

---

### Section 4.2: Quality Analysis (8 minutes)

#### Narration Script:

Understanding why debates succeed or fail is crucial for continuous improvement. SuperAgent provides detailed quality analysis that helps you identify areas for optimization.

#### Quality Report Structure:
```json
{
  "quality_report": {
    "debate_id": "debate-abc123",
    "overall_assessment": "good",
    "score": 0.84,

    "strengths": [
      "High coherence across responses",
      "Good engagement between participants",
      "Clear consensus reached"
    ],

    "areas_for_improvement": [
      "Round 1 responses could be more focused",
      "Consider adding more diverse perspectives",
      "Some arguments lacked supporting evidence"
    ],

    "participant_analysis": [
      {
        "name": "Technical Analyst",
        "scores": {
          "clarity": 0.92,
          "relevance": 0.88,
          "evidence": 0.85,
          "engagement": 0.79
        },
        "feedback": "Strong technical analysis but could engage more with other participants"
      },
      {
        "name": "Critical Reviewer",
        "scores": {
          "clarity": 0.85,
          "relevance": 0.90,
          "evidence": 0.78,
          "engagement": 0.88
        },
        "feedback": "Good critical analysis, consider providing more constructive alternatives"
      }
    ],

    "round_analysis": [
      {
        "round": 1,
        "quality": 0.78,
        "notes": "Opening arguments established good foundation"
      },
      {
        "round": 2,
        "quality": 0.85,
        "notes": "Strong engagement and refinement"
      },
      {
        "round": 3,
        "quality": 0.89,
        "notes": "Effective synthesis and conclusion"
      }
    ],

    "recommendations": [
      "Consider using 'structured' debate strategy for complex topics",
      "Add a mediator participant for better synthesis",
      "Increase timeout for round 1 to allow deeper initial analysis"
    ]
  }
}
```

#### Quality Monitoring API:
```bash
# Get quality analysis for a debate
curl http://localhost:8080/v1/debates/{debate_id}/quality

# Get historical quality trends
curl http://localhost:8080/v1/debates/analytics?period=30d

# Response includes trend data
{
  "analytics": {
    "period": "30d",
    "total_debates": 156,
    "avg_quality_score": 0.82,
    "avg_consensus_rate": 0.78,
    "quality_trend": "improving",
    "most_effective_strategies": ["structured", "collaborative"],
    "top_performing_combinations": [
      {
        "participants": ["claude", "gemini", "deepseek"],
        "avg_score": 0.88
      }
    ]
  }
}
```

---

### Section 4.3: Troubleshooting (8 minutes)

#### Narration Script:

When debates don't go as expected, systematic troubleshooting helps identify and resolve issues quickly. Let me walk you through common problems and their solutions.

#### Common Issues and Solutions:

**Issue 1: Low Consensus Score**
```
Symptom: Consensus confidence below 0.5
Possible Causes:
- Topic is genuinely controversial
- Participants have mismatched capabilities
- Poor role assignment

Solutions:
1. Increase max_rounds to allow more refinement
2. Add a mediator participant
3. Use collaborative strategy
4. Review and adjust participant weights

Debug Query:
curl http://localhost:8080/v1/debates/{id}/diagnostics
```

**Issue 2: Timeout Errors**
```
Symptom: Debate fails with timeout
Possible Causes:
- Provider response times too slow
- Timeout set too low
- Network issues

Solutions:
1. Increase timeout configuration
2. Check provider health: curl http://localhost:8080/v1/providers/health
3. Use faster models for less critical participants
4. Enable provider fallback

Configuration Fix:
debate_config:
  timeout: 600000  # Increase to 10 minutes
  participant_timeout: 60000  # Per-participant timeout
```

**Issue 3: Poor Quality Responses**
```
Symptom: Quality scores consistently below 0.6
Possible Causes:
- Vague or broad topic
- Inappropriate model selection
- System prompt issues

Solutions:
1. Make topic more specific
2. Match model strengths to participant roles
3. Customize system prompts
4. Increase max_tokens

Example Topic Refinement:
Before: "Discuss database options"
After: "Compare PostgreSQL vs MongoDB for a high-write e-commerce
       application with 10M daily transactions"
```

**Issue 4: Provider Failures**
```
Symptom: One or more providers failing consistently
Debug:
curl http://localhost:8080/v1/providers/health

Check circuit breaker status:
{
  "claude": {
    "status": "healthy",
    "circuit_breaker": "closed"
  },
  "gemini": {
    "status": "unhealthy",
    "circuit_breaker": "open",
    "last_error": "rate_limit_exceeded"
  }
}

Solutions:
1. Configure fallback providers
2. Adjust rate limits
3. Implement retry logic
4. Use alternative models
```

#### Debug Mode Configuration:
```yaml
# Enable comprehensive debugging
debug:
  enabled: true
  log_prompts: true
  log_responses: true
  log_scores: true
  trace_requests: true
```

---

### Section 4.4: Best Practices (6 minutes)

#### Narration Script:

Let me share the best practices we've learned from running thousands of debates. These guidelines will help you get the most out of the AI Debate system.

#### Best Practice Guidelines:

**1. Topic Design**
```
DO:
- Be specific and focused
- Include context and constraints
- Define success criteria
- Example: "Design a caching strategy for a read-heavy API serving
            1M requests/minute with 99.9% cache hit rate target"

DON'T:
- Be too vague
- Ask multiple unrelated questions
- Leave constraints ambiguous
- Example: "Talk about caching" (too vague)
```

**2. Participant Selection**
```
Optimal Configuration:
- 3-5 participants (more can slow debate)
- Mix of complementary roles
- Match provider strengths to roles

Recommended Combinations:
[Technical Topic]
  - Claude (Analyst) - deep reasoning
  - DeepSeek (Proposer) - technical implementation
  - Gemini (Critic) - broad perspective

[Creative Topic]
  - Claude (Mediator) - synthesis
  - Grok (Proposer) - creative ideas
  - Gemini (Critic) - grounding
```

**3. Round Configuration**
```
Recommended Settings:
- max_rounds: 3 (for most topics)
- max_rounds: 4-5 (for complex topics)
- timeout per round: 60-90 seconds
- total timeout: 5-10 minutes

Early Termination:
- consensus_threshold: 0.85 (for early exit)
- This saves time when agreement is reached
```

**4. Quality Optimization**
```
Improve Quality By:
- Setting minimum response length
- Using structured debate strategy
- Enabling Cognee for complex topics
- Reviewing and iterating on prompts

Configuration:
quality_config:
  min_response_length: 100
  max_response_length: 2000
  quality_threshold: 0.7
  require_evidence: true
```

**5. Cost Optimization**
```
Reduce Costs By:
- Using local models (Ollama) for simpler roles
- Setting appropriate max_tokens
- Using faster models for critics
- Caching common debates

Cost-Effective Setup:
participants:
  - name: "Primary"
    llm_provider: "claude"  # Premium for main role
  - name: "Secondary"
    llm_provider: "ollama"  # Free for support role
```

#### Slide Content:
```
BEST PRACTICES SUMMARY

[Topic Design]
Specific > Vague
Constrained > Open-ended
Measurable > Abstract

[Participants]
3-5 optimal
Mix complementary roles
Match strengths to roles

[Configuration]
3 rounds default
5-10 min timeout
0.7 quality threshold

[Optimization]
Monitor metrics
Iterate on prompts
Balance cost vs quality
```

---

## Course Wrap-up (2 minutes)

#### Narration Script:

Congratulations on completing the AI Debate System Mastery course! You now have a comprehensive understanding of how to configure, run, and optimize AI debates using SuperAgent.

You've learned how multi-agent collaboration works, how to configure participants with different roles and providers, advanced techniques like memory integration and Cognee enhancement, and how to monitor and troubleshoot your debates.

In Course 3, we'll move to production deployment, covering architecture, scaling, and operational excellence.

#### Slide Content:
```
COURSE COMPLETE!

What You Mastered:
- AI Debate fundamentals and consensus building
- Participant configuration and role assignment
- Multi-round debate orchestration
- Memory and Cognee integration
- Performance monitoring and troubleshooting
- Best practices for quality debates

Next Steps:
- Course 3: Production Deployment
- Practice with complex debate scenarios
- Experiment with different strategies

Hands-on Exercises:
1. Configure a 4-participant technical debate
2. Implement Cognee-enhanced analysis
3. Compare voting strategies
```

---

## Supplementary Materials

### Exercise 1: Technical Architecture Debate
Configure a debate on "Microservices vs Monolith" with 4 participants using different providers.

### Exercise 2: Cognee Integration
Enable Cognee and analyze the knowledge graph output from a debate.

### Exercise 3: Strategy Comparison
Run the same topic with different voting strategies and compare results.

### Quick Reference

```yaml
# Minimal Debate Configuration
minimal_debate:
  topic: "Your topic here"
  participants:
    - name: "Participant 1"
      role: "proposer"
      llm_provider: "claude"
    - name: "Participant 2"
      role: "critic"
      llm_provider: "gemini"
  max_rounds: 3
  timeout: 300000

# Full-Featured Configuration
full_debate:
  topic: "Your topic here"
  participants: [...]
  max_rounds: 4
  timeout: 600000
  debate_strategy: "structured"
  voting_strategy: "confidence_weighted"
  consensus_threshold: 0.75
  enable_cognee: true
  enable_memory: true
  quality_threshold: 0.7
```
