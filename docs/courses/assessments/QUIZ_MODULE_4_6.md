# Assessment: Modules 4-6 (Provider Expert)

## Quiz Instructions

- Total Questions: 30
- Time Limit: 45 minutes
- Passing Score: 80% (24/30)
- Open book: No
- Required for: Level 2 Certification

---

## Section A: Multiple Choice (18 questions)

### 1. Which interface must all LLM providers implement?

A) `ProviderInterface`
B) `LLMProvider`
C) `AIProvider`
D) `ModelInterface`

---

### 2. What method is used for streaming responses?

A) `Complete()`
B) `Stream()`
C) `CompleteStream()`
D) `GetStream()`

---

### 3. Which Claude model is highest quality?

A) claude-3-haiku
B) claude-3-sonnet
C) claude-3-opus
D) claude-3.5-sonnet

---

### 4. What is DeepSeek-Coder optimized for?

A) Creative writing
B) Code and technical content
C) Image generation
D) Audio processing

---

### 5. Which voting strategy weights votes by confidence scores?

A) `majority`
B) `weighted`
C) `confidence_weighted`
D) `quality_weighted`

---

### 6. What is the default consensus threshold?

A) 0.5
B) 0.6
C) 0.75
D) 0.9

---

### 7. How many fallback LLMs per debate position are recommended?

A) 1
B) 2
C) 3
D) 5

---

### 8. Which debate style shows theatrical dialogue?

A) `minimal`
B) `novel`
C) `theater`
D) `screenplay`

---

### 9. What is the maximum number of debate participants supported?

A) 5
B) 7
C) 10
D) 15

---

### 10. Which provider is typically lowest cost per token?

A) Claude
B) DeepSeek
C) GPT-4
D) Gemini Pro

---

### 11. What does LLMsVerifier do?

A) Encrypts API keys
B) Validates provider accuracy and reliability
C) Creates LLM backups
D) Manages user sessions

---

### 12. How does HelixAgent select the primary provider?

A) Alphabetically
B) Random selection
C) By LLMsVerifier score (highest first)
D) By API key order

---

### 13. Which debate strategy pits opposing viewpoints?

A) `round_robin`
B) `structured`
C) `adversarial`
D) `collaborative`

---

### 14. What is the `weight` parameter in participant config?

A) File size
B) Influence in voting
C) Memory usage
D) Token count

---

### 15. Which endpoint creates a new debate?

A) `GET /v1/debates`
B) `POST /v1/debates`
C) `PUT /v1/debates`
D) `PATCH /v1/debates`

---

### 16. What does `argumentation_style: logical` mean?

A) Uses formal logic and reasoning
B) Appeals to emotions
C) Uses hypothetical scenarios
D) Cites authorities

---

### 17. Which parameter controls response creativity?

A) `max_tokens`
B) `top_p`
C) `temperature`
D) `frequency_penalty`

---

### 18. What is the purpose of `quality_threshold`?

A) Limit response length
B) Minimum acceptable response quality
C) API rate limiting
D) Cache duration

---

## Section B: True/False (6 questions)

### 19. All LLM providers must support streaming.

- [ ] True
- [ ] False

---

### 20. The AI Debate system requires exactly 5 participants.

- [ ] True
- [ ] False

---

### 21. Fallback chains activate when the primary LLM fails.

- [ ] True
- [ ] False

---

### 22. Provider weights must sum to 1.0.

- [ ] True
- [ ] False

---

### 23. Cognee integration enhances debate responses with semantic analysis.

- [ ] True
- [ ] False

---

### 24. The same LLM can be used in multiple debate positions.

- [ ] True
- [ ] False

---

## Section C: Short Answer (6 questions)

### 25. List the five debate participant roles and their primary functions.

| Role | Function |
|------|----------|
| | |
| | |
| | |
| | |
| | |

---

### 26. Write a curl command to create a debate with the topic "AI in healthcare" using 3 rounds.

```bash

```

---

### 27. What are three methods in the LLMProvider interface?

```
1. _______________________
2. _______________________
3. _______________________
```

---

### 28. Explain the difference between `majority` and `confidence_weighted` voting.

```


```

---

### 29. What happens when no consensus is reached after max rounds?

```


```

---

### 30. Configure a participant named "TechExpert" with Claude as primary and DeepSeek as fallback (YAML format).

```yaml


```

---

## Answer Key

### Section A
1. B
2. C
3. C
4. B
5. C
6. C
7. B
8. C
9. C
10. B
11. B
12. C
13. C
14. B
15. B
16. A
17. C
18. B

### Section B
19. False
20. False (2-10 participants supported)
21. True
22. False
23. True
24. True

### Section C
25.
| Role | Function |
|------|----------|
| Analyst | Systematic analysis |
| Proposer | Solution proposals |
| Critic | Challenge assumptions |
| Synthesizer | Combine perspectives |
| Mediator | Reach consensus |

26. `curl -X POST http://localhost:7061/v1/debates -H "Content-Type: application/json" -d '{"topic":"AI in healthcare","rounds":3}'`

27. Any three: Complete(), CompleteStream(), HealthCheck(), GetCapabilities(), ValidateConfig()

28. Majority voting counts each vote equally. Confidence-weighted voting gives more weight to responses with higher confidence scores.

29. The debate returns the best available response with consensus: false, along with partial results and all participant contributions.

30.
```yaml
- name: "TechExpert"
  role: "Technical Expert"
  debate_style: technical
  llms:
    - provider: claude
      model: claude-3-5-sonnet-20241022
    - provider: deepseek
      model: deepseek-chat
```

---

## Grading

| Score | Result |
|-------|--------|
| 27-30 | Excellent (A) |
| 24-26 | Pass (B) |
| 21-23 | Marginal (C) |
| <21 | Fail - Retry |

---

*Assessment Version: 1.0.0*
*Last Updated: January 2026*
