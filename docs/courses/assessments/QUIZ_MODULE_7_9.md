# Quiz: Modules 7-9 (Advanced Features)

## Instructions

- **Total Questions**: 30
- **Time Limit**: 60 minutes
- **Passing Score**: 80% (24/30)
- **Format**: Multiple choice (single answer unless specified)

---

## Section 1: Observability (Questions 1-10)

### Q1. What is OpenTelemetry?

A) A logging framework for Go applications
B) A distributed tracing standard
C) An observability framework combining traces, metrics, and logs
D) A monitoring dashboard tool

---

### Q2. In HelixAgent, which component is responsible for collecting and exporting traces?

A) Prometheus
B) Jaeger directly
C) OpenTelemetry Collector
D) Grafana

---

### Q3. What is the purpose of trace sampling in production?

A) To reduce storage costs only
B) To improve query performance only
C) To reduce overhead while maintaining visibility into system behavior
D) To comply with data retention regulations

---

### Q4. Which trace exporter does HelixAgent support for LLM-specific observability?

A) Jaeger
B) Zipkin
C) Langfuse
D) All of the above

---

### Q5. What is a span in distributed tracing?

A) A single log entry
B) A unit of work or operation within a trace
C) A metric measurement
D) A network packet

---

### Q6. Which HTTP header is used for W3C Trace Context propagation?

A) X-Trace-ID
B) traceparent
C) X-Request-ID
D) Correlation-ID

---

### Q7. What sampling rate means 10% of requests are traced?

A) 1.0
B) 0.1
C) 10
D) 0.01

---

### Q8. In HelixAgent's observability setup, what port is used for Prometheus metrics by default?

A) 8080
B) 9090
C) 4317
D) 6831

---

### Q9. What attribute should be added to spans for LLM requests?

A) http.method only
B) Token count, model name, and provider
C) User agent string
D) Response body

---

### Q10. Which command enables tracing in HelixAgent?

A) `make enable-tracing`
B) Setting `observability.tracing.enabled: true` in config
C) `go run --trace`
D) `OTEL_ENABLED=1`

---

## Section 2: RAG System (Questions 11-20)

### Q11. What does RAG stand for?

A) Retrieval Augmented Generation
B) Random Access Generation
C) Rapid Answer Generator
D) Recursive Answer Graph

---

### Q12. What is hybrid retrieval in HelixAgent's RAG system?

A) Using multiple databases
B) Combining dense (embedding) and sparse (keyword) retrieval methods
C) Switching between cloud and local retrieval
D) Using both SQL and NoSQL databases

---

### Q13. What is the purpose of a reranker in RAG?

A) To sort results alphabetically
B) To reorder retrieved documents by relevance to the query
C) To compress retrieved documents
D) To cache frequently accessed documents

---

### Q14. Which vector database does HelixAgent integrate with for RAG?

A) Pinecone
B) Weaviate
C) Qdrant
D) Milvus

---

### Q15. What is the typical embedding dimension used in HelixAgent?

A) 128
B) 512
C) 1536
D) 4096

---

### Q16. In sparse retrieval, what algorithm is commonly used?

A) Cosine similarity
B) BM25
C) K-means
D) PageRank

---

### Q17. What is chunking in the context of RAG?

A) Data compression
B) Splitting documents into smaller pieces for embedding
C) Network packet segmentation
D) Memory allocation

---

### Q18. What metric measures how well retrieved context supports the generated answer?

A) Precision
B) Recall
C) Faithfulness
D) F1 Score

---

### Q19. In HelixAgent's RAG pipeline, when does reranking occur?

A) Before retrieval
B) After initial retrieval but before LLM generation
C) After LLM generation
D) During embedding

---

### Q20. What is the purpose of the `top_k` parameter in RAG retrieval?

A) Maximum embedding dimensions
B) Number of documents to retrieve
C) Chunk size in tokens
D) Reranker batch size

---

## Section 3: Memory Management (Questions 21-30)

### Q21. What is Mem0-style memory in HelixAgent?

A) In-memory caching for fast access
B) Persistent AI memory with fact extraction and entity graphs
C) Database connection pooling
D) File system storage

---

### Q22. What are the four memory types supported in HelixAgent?

A) Short, Medium, Long, Permanent
B) Episodic, Semantic, Procedural, Working
C) User, System, Application, Shared
D) Local, Remote, Distributed, Cached

---

### Q23. What is episodic memory used for?

A) Storing facts and knowledge
B) Storing conversation and event memories
C) Storing how-to procedures
D) Temporary computation storage

---

### Q24. What is an entity in the memory system?

A) A database table
B) An extracted subject (person, place, thing) from memories
C) A memory storage location
D) A user session

---

### Q25. What does the `importance` field in a Memory struct represent?

A) Memory size in bytes
B) A score indicating how significant the memory is
C) Creation timestamp
D) Access frequency only

---

### Q26. How are relationships between entities represented?

A) Foreign keys in SQL
B) Source ID, Target ID, and relationship type with strength
C) XML tags
D) Nested JSON objects

---

### Q27. What is the purpose of memory decay in the system?

A) To delete old memories automatically
B) To gradually reduce importance of unused memories over time
C) To compress memory storage
D) To encrypt sensitive memories

---

### Q28. Which search method is used for semantic memory retrieval?

A) SQL LIKE queries
B) Vector similarity search with embeddings
C) Full-text search
D) Regular expressions

---

### Q29. What happens when `AccessCount` is incremented for a memory?

A) Memory is marked for deletion
B) Memory importance may increase
C) Memory is compressed
D) Memory is encrypted

---

### Q30. What is the typical use case for working memory?

A) Long-term storage of user preferences
B) Short-term context during a conversation
C) System configuration storage
D) Backup storage

---

## Answer Key

| Q | Answer | Q | Answer | Q | Answer |
|---|--------|---|--------|---|--------|
| 1 | C | 11 | A | 21 | B |
| 2 | C | 12 | B | 22 | B |
| 3 | C | 13 | B | 23 | B |
| 4 | D | 14 | C | 24 | B |
| 5 | B | 15 | C | 25 | B |
| 6 | B | 16 | B | 26 | B |
| 7 | B | 17 | B | 27 | B |
| 8 | B | 18 | C | 28 | B |
| 9 | B | 19 | B | 29 | B |
| 10 | B | 20 | B | 30 | B |

---

## Scoring

- **90-100% (27-30)**: Excellent - Ready for advanced topics
- **80-89% (24-26)**: Good - Review missed topics
- **70-79% (21-23)**: Fair - Additional study recommended
- **Below 70%**: Review modules 7-9 before proceeding

## Next Steps

After passing this quiz:
1. Complete Lab 4: MCP Integration
2. Review any missed topics
3. Proceed to Modules 10-11 (Security & Deployment)
