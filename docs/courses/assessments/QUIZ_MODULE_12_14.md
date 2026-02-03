# Assessment: Modules 12-14 (Challenge Expert)

## Quiz Instructions

- **Total Questions**: 35
- **Time Limit**: 75 minutes
- **Passing Score**: 80% (28/35)
- **Open book**: No
- **Required for**: Level 5 Certification - Challenge Expert

---

## Section A: Multiple Choice - Challenge System (Questions 1-12)

### 1. What are the three main challenge types in HelixAgent's Challenge System?

A) Unit, Integration, E2E
B) RAGS, MCPS, SKILLS
C) Provider, Ensemble, Debate
D) Security, Performance, Reliability

---

### 2. How many MCP servers are tested in the MCPS challenge?

A) 10
B) 15
C) 22
D) 45

---

### 3. What does "strict real-result validation" ensure in challenge scripts?

A) Tests run in strict mode with no parallelism
B) HTTP 200 responses are accepted as success
C) HTTP 200 alone is not enough - responses must have real content
D) Only real API keys can be used for testing

---

### 4. Which RAG systems are tested in the RAGS challenge?

A) Only Cognee
B) Cognee, Qdrant, RAG Pipeline, Embeddings Service
C) PostgreSQL, Redis, MongoDB
D) ChromaDB, Pinecone, Weaviate

---

### 5. What minimum content length is typically required for strict validation to pass?

A) 10 characters
B) 50 characters
C) 100 characters
D) 500 characters

---

### 6. How many skill categories are tested in the SKILLS challenge?

A) 4
B) 6
C) 8
D) 10

---

### 7. What is a "FALSE SUCCESS" in the context of challenge validation?

A) A test that times out
B) HTTP 200 with no real content or error messages in response
C) A test that fails silently
D) A cached response from a previous test

---

### 8. Which command runs all three challenge scripts?

A) `make test-challenges`
B) `./challenges/scripts/run_all_challenges.sh`
C) `./run_challenges.sh`
D) `make ci-validate-all`

---

### 9. What file format is used for individual test results in challenges?

A) JSON
B) YAML
C) CSV
D) XML

---

### 10. Which protocols are tested in the MCPS challenge besides MCP?

A) HTTP only
B) LSP and ACP
C) gRPC and WebSocket
D) MQTT and AMQP

---

### 11. How many CLI agents are validated across the challenge system?

A) 10+
B) 15+
C) 20+
D) 50+

---

### 12. What is the purpose of the challenge results directory structure?

A) Log rotation
B) Organized storage of test reports by date/time
C) Backup of test configurations
D) Cache for test data

---

## Section B: Multiple Choice - MCP Tool Search (Questions 13-22)

### 13. Which endpoint is used to search for MCP tools by query?

A) `/v1/mcp/search`
B) `/v1/mcp/tools/search`
C) `/v1/tools/find`
D) `/v1/search/mcp`

---

### 14. What HTTP methods are supported for MCP tool search?

A) GET only
B) POST only
C) GET and POST
D) PUT and PATCH

---

### 15. Which endpoint provides AI-powered tool suggestions based on prompts?

A) `/v1/mcp/tools/ai`
B) `/v1/mcp/tools/suggestions`
C) `/v1/mcp/suggest`
D) `/v1/ai/tools`

---

### 16. What query parameter is used for the search query?

A) `query`
B) `search`
C) `q`
D) `term`

---

### 17. What does the tool suggestions response include?

A) Only tool names
B) Tool name, confidence score, and reason
C) Tool documentation
D) Source code

---

### 18. Which endpoint is used to search for MCP adapters?

A) `/v1/mcp/adapters/find`
B) `/v1/mcp/adapters/search`
C) `/v1/adapters/mcp`
D) `/v1/search/adapters`

---

### 19. What parameter limits the number of search results?

A) `max`
B) `count`
C) `limit`
D) `size`

---

### 20. Which endpoint returns MCP tool categories?

A) `/v1/mcp/types`
B) `/v1/mcp/categories`
C) `/v1/mcp/groups`
D) `/v1/mcp/kinds`

---

### 21. What information does `/v1/mcp/stats` provide?

A) Server metrics only
B) Tool usage statistics
C) Network latency
D) Error rates

---

### 22. Which adapter category includes GitHub and GitLab?

A) Core
B) VCS (Version Control)
C) Cloud
D) Communication

---

## Section C: Multiple Choice - Advanced AI Debate (Questions 23-32)

### 23. How many LLMs are in the full AI Debate Ensemble?

A) 5
B) 10
C) 15
D) 20

---

### 24. How is the 15 LLM debate team structured?

A) 15 independent participants
B) 5 positions with 3 LLMs each (1 primary + 2 fallbacks)
C) 3 teams with 5 LLMs each
D) 1 primary with 14 fallbacks

---

### 25. What are the 4 phases in multi-pass validation?

A) Plan, Execute, Verify, Deploy
B) Request, Process, Validate, Return
C) Initial, Validation, Polish, Final
D) Start, Middle, Review, End

---

### 26. Which scoring component has the highest weight (25%) in LLMsVerifier?

A) Capability alone
B) ResponseSpeed and CostEffectiveness
C) Recency
D) ModelEfficiency

---

### 27. What bonus score do OAuth providers receive?

A) 0.1
B) 0.25
C) 0.5
D) 1.0

---

### 28. What parameter enables multi-pass validation in debate requests?

A) `multi_pass: true`
B) `enable_multi_pass_validation: true`
C) `validation_enabled: true`
D) `multi_validate: true`

---

### 29. What does the `min_confidence_to_skip` parameter control?

A) Minimum response length
B) Skip validation if confidence exceeds this threshold
C) Number of participants required
D) Timeout duration

---

### 30. Which debate topologies are supported in the Debate Orchestrator?

A) Linear only
B) Mesh, Star, Chain
C) Binary Tree, Graph
D) Ring, Broadcast

---

### 31. What is the correct phase sequence in the debate protocol?

A) Critique -> Proposal -> Review -> Synthesis
B) Proposal -> Review -> Critique -> Synthesis
C) Proposal -> Critique -> Review -> Synthesis
D) Review -> Proposal -> Critique -> Synthesis

---

### 32. What HTTP header identifies CLI agents in requests?

A) `User-Agent` only
B) `X-CLI-Agent` header
C) `Authorization`
D) `X-Request-ID`

---

## Section D: Short Answer (Questions 33-35)

### 33. List the five debate participant positions in the AI Debate system.

```
1. _______________________
2. _______________________
3. _______________________
4. _______________________
5. _______________________
```

---

### 34. Write a curl command to search for "file" tools using the MCP tool search endpoint.

```bash

```

---

### 35. Describe the purpose of each multi-pass validation phase in 1-2 sentences each.

```
Phase 1 (INITIAL):


Phase 2 (VALIDATION):


Phase 3 (POLISH):


Phase 4 (FINAL):

```

---

## Answer Key

### Section A: Challenge System (1-12)
| Q | Answer | Explanation |
|---|--------|-------------|
| 1 | B | RAGS, MCPS, and SKILLS are the three challenge types |
| 2 | C | 22 MCP servers are tested including filesystem, git, databases, etc. |
| 3 | C | Strict validation ensures HTTP 200 responses have real content, not just status codes |
| 4 | B | Cognee, Qdrant, RAG Pipeline, and Embeddings Service are all tested |
| 5 | B | 50 characters is the typical minimum for content length validation |
| 6 | C | 8 categories: Code, Debug, Search, Git, Deploy, Docs, Test, Review |
| 7 | B | FALSE SUCCESS = HTTP 200 with empty/error content masquerading as success |
| 8 | B | The run_all_challenges.sh script runs RAGS, MCPS, and SKILLS challenges |
| 9 | C | test_results.csv stores individual test results |
| 10 | B | LSP (Language Server Protocol) and ACP (Agent Client Protocol) are also tested |
| 11 | C | 20+ CLI agents including OpenCode, ClaudeCode, Aider, Cline, etc. |
| 12 | B | Results are organized by challenge type, year, month, day, and timestamp |

### Section B: MCP Tool Search (13-22)
| Q | Answer | Explanation |
|---|--------|-------------|
| 13 | B | `/v1/mcp/tools/search` is the tool search endpoint |
| 14 | C | Both GET (with query params) and POST (with JSON body) are supported |
| 15 | B | `/v1/mcp/tools/suggestions` provides AI-powered recommendations |
| 16 | C | The `q` parameter is used for search queries |
| 17 | B | Suggestions include tool name, confidence score, and reason for recommendation |
| 18 | B | `/v1/mcp/adapters/search` is the adapter search endpoint |
| 19 | C | The `limit` parameter controls maximum results |
| 20 | B | `/v1/mcp/categories` lists tool categories |
| 21 | B | Stats endpoint provides tool usage statistics |
| 22 | B | VCS (Version Control System) category includes GitHub and GitLab |

### Section C: Advanced AI Debate (23-32)
| Q | Answer | Explanation |
|---|--------|-------------|
| 23 | C | 15 LLMs: 5 positions x 3 LLMs each |
| 24 | B | 5 positions with 1 primary + 2 fallback LLMs each |
| 25 | C | Initial -> Validation -> Polish -> Final |
| 26 | B | ResponseSpeed (25%) and CostEffectiveness (25%) share highest weight |
| 27 | C | OAuth providers receive +0.5 bonus score |
| 28 | B | `enable_multi_pass_validation: true` enables multi-pass |
| 29 | B | Skip validation phases if confidence already exceeds threshold |
| 30 | B | Mesh (parallel), Star (hub-spoke), and Chain (sequential) |
| 31 | C | Proposal -> Critique -> Review -> Synthesis |
| 32 | B | X-CLI-Agent header identifies CLI agents |

### Section D: Short Answer (33-35)

**33. Five Debate Positions**:
1. Analyst
2. Proposer
3. Critic
4. Synthesizer
5. Mediator

**34. MCP Tool Search curl command**:
```bash
curl "http://localhost:7061/v1/mcp/tools/search?q=file"
```
OR
```bash
curl -X POST http://localhost:7061/v1/mcp/tools/search \
  -H "Content-Type: application/json" \
  -d '{"query": "file"}'
```

**35. Multi-Pass Validation Phases**:

Phase 1 (INITIAL): Each AI participant provides their initial perspective and analysis of the topic. This is the first round of responses from all debate positions.

Phase 2 (VALIDATION): Cross-validation of responses for accuracy and completeness. Participants verify factual claims and check for inconsistencies.

Phase 3 (POLISH): Refinement and improvement based on validation feedback. Responses are enhanced for clarity, accuracy, and structure.

Phase 4 (FINAL): Synthesized consensus with confidence scores. The mediator combines all validated and polished perspectives into a final unified response.

---

## Grading

| Score | Result |
|-------|--------|
| 32-35 | Excellent (A) - Challenge Expert |
| 28-31 | Pass (B) - Certified |
| 24-27 | Marginal (C) - Additional study required |
| <24 | Fail - Retry after review |

---

## Certification Requirements

To achieve Level 5: Challenge Expert certification:

1. **Pass this quiz with 80%+** (28/35 correct)
2. **Run all challenge scripts successfully**:
   - RAGS challenge: 100% pass rate
   - MCPS challenge: 100% pass rate
   - SKILLS challenge: 100% pass rate
3. **Configure 15 LLM debate team** with multi-pass validation
4. **Demonstrate MCP Tool Search** integration
5. **Document strict validation methodology**

---

## Next Steps

After passing this assessment:

1. Run all three challenges and achieve 100% pass rates
2. Configure and test the 15 LLM debate team
3. Build a custom MCP tool discovery workflow
4. Review advanced debugging techniques
5. Prepare for production deployment

---

## Additional Resources

- [Challenge Scripts Documentation](../../../challenges/README.md)
- [MCP Tool Search Guide](../../guides/mcp-tool-search.md)
- [AI Debate Configuration](../../guides/ai-debate.md)
- [LLMsVerifier Documentation](../../../LLMsVerifier/README.md)

---

*Assessment Version: 1.0.0*
*Last Updated: January 2026*
*Level: 5 - Challenge Expert*
