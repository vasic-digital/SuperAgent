# Assessment: Modules 1-3 (HelixAgent Fundamentals)

## Quiz Instructions

- Total Questions: 25
- Time Limit: 30 minutes
- Passing Score: 80% (20/25)
- Open book: No
- Required for: Level 1 Certification

---

## Section A: Multiple Choice (15 questions)

### 1. What is HelixAgent primarily designed for?

A) Single LLM API gateway
B) Multi-provider AI orchestration with intelligent routing
C) Database management
D) Container orchestration

---

### 2. How many LLM providers does HelixAgent support?

A) 5
B) 10
C) 18+
D) 3

---

### 3. What is the default port for HelixAgent?

A) 8080
B) 3000
C) 7061
D) 9000

---

### 4. Which command builds HelixAgent from source?

A) `go run .`
B) `make build`
C) `npm build`
D) `docker build`

---

### 5. What is the AI Debate Ensemble?

A) A single LLM responding to queries
B) Multiple AI participants with distinct roles reaching consensus
C) A database for storing conversations
D) A caching system

---

### 6. How many participant roles are in the AI Debate system?

A) 3
B) 4
C) 5
D) 7

---

### 7. Which file contains development configuration?

A) `config.json`
B) `configs/development.yaml`
C) `.env.production`
D) `settings.xml`

---

### 8. What is the minimum Go version required?

A) 1.20
B) 1.21
C) 1.23
D) 1.24+

---

### 9. Which Docker profile includes all services?

A) `--profile core`
B) `--profile ai`
C) `--profile full`
D) `--profile all`

---

### 10. What endpoint returns health status?

A) `/status`
B) `/health`
C) `/ping`
D) `/check`

---

### 11. Which environment variable configures the Claude API key?

A) `ANTHROPIC_KEY`
B) `CLAUDE_KEY`
C) `CLAUDE_API_KEY`
D) `ANTHROPIC_API_KEY`

---

### 12. What is the OpenAI-compatible endpoint for chat completions?

A) `/api/chat`
B) `/v1/chat/completions`
C) `/completions`
D) `/v1/complete`

---

### 13. What database does HelixAgent use?

A) MySQL
B) MongoDB
C) PostgreSQL
D) SQLite

---

### 14. What is used for caching in HelixAgent?

A) Memcached
B) Redis
C) Local files
D) Browser cache

---

### 15. Which command runs all tests?

A) `npm test`
B) `go test`
C) `make test`
D) `./test.sh`

---

## Section B: True/False (5 questions)

### 16. HelixAgent can only use one LLM provider at a time.

- [ ] True
- [ ] False

---

### 17. The AI Debate system uses fallback chains per participant.

- [ ] True
- [ ] False

---

### 18. HelixAgent provides OpenAI-compatible APIs.

- [ ] True
- [ ] False

---

### 19. You must configure all 18+ providers to run HelixAgent.

- [ ] True
- [ ] False

---

### 20. Docker is required to run HelixAgent.

- [ ] True
- [ ] False

---

## Section C: Short Answer (5 questions)

### 21. Name three core components of HelixAgent architecture.

```
1. _______________________
2. _______________________
3. _______________________
```

---

### 22. What are the five AI Debate participant roles?

```
1. _______________________
2. _______________________
3. _______________________
4. _______________________
5. _______________________
```

---

### 23. Write the curl command to check HelixAgent health.

```bash

```

---

### 24. List two benefits of multi-provider AI orchestration.

```
1. _______________________
2. _______________________
```

---

### 25. What is the purpose of the fallback chain in HelixAgent?

```


```

---

## Answer Key

### Section A
1. B
2. C
3. C
4. B
5. B
6. C
7. B
8. D
9. C
10. B
11. C
12. B
13. C
14. B
15. C

### Section B
16. False
17. True
18. True
19. False
20. False

### Section C
21. Any three of: LLM Provider Layer, Ensemble Orchestration, AI Debate Service, Plugin System, Protocol Managers, Caching Layer, Monitoring
22. Analyst, Proposer, Critic, Synthesizer, Mediator
23. `curl http://localhost:7061/health`
24. Any two of: No vendor lock-in, Better reliability, Multiple perspectives, Cost optimization, Performance optimization
25. Provides automatic failover when primary LLM fails, ensuring high availability

---

## Grading

| Score | Result |
|-------|--------|
| 23-25 | Excellent (A) |
| 20-22 | Pass (B) |
| 17-19 | Marginal (C) |
| <17 | Fail - Retry |

---

*Assessment Version: 1.0.0*
*Last Updated: January 2026*
