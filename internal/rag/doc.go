// Package rag provides Retrieval-Augmented Generation capabilities for HelixAgent.
//
// This package implements a hybrid retrieval system combining dense and sparse
// retrieval methods with reranking for optimal document retrieval.
//
// # RAG Pipeline
//
// The RAG pipeline consists of several stages:
//
//  1. Document ingestion and chunking
//  2. Embedding generation (dense vectors)
//  3. Sparse index creation (BM25)
//  4. Hybrid retrieval (dense + sparse)
//  5. Reranking
//  6. Context assembly for LLM
//
// # Hybrid Retrieval
//
// Combines multiple retrieval strategies:
//
//	pipeline := rag.NewHybridPipeline(config)
//
//	// Add documents
//	if err := pipeline.AddDocuments(ctx, documents); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve relevant documents
//	results, err := pipeline.Retrieve(ctx, query, topK)
//
// Retrieval methods:
//   - Dense: Semantic similarity using embeddings
//   - Sparse: Keyword matching using BM25
//   - Hybrid: Weighted combination of dense and sparse
//
// # Reranking
//
// Improve retrieval quality with reranking:
//
//	reranker := rag.NewCrossEncoderReranker(config)
//	reranked := reranker.Rerank(ctx, query, candidates)
//
// Supported rerankers:
//   - Cross-encoder models
//   - Cohere Rerank
//   - Custom rerankers
//
// # HyDE (Hypothetical Document Embeddings)
//
// Generate hypothetical documents for improved retrieval:
//
//	hyde := rag.NewHyDE(llmProvider)
//	expandedQuery, err := hyde.Expand(ctx, query)
//
// # Vector Store Integration
//
// Supports multiple vector databases:
//
//   - Qdrant: Full-featured vector database
//   - Pinecone: Managed vector service
//   - Milvus: High-performance distributed
//   - pgvector: PostgreSQL extension
//
// Example with Qdrant:
//
//	store := rag.NewQdrantStore(qdrantConfig)
//	pipeline.SetVectorStore(store)
//
// # Document Processing
//
// Chunking strategies:
//
//	chunker := rag.NewSemanticChunker(config)
//	chunks := chunker.Chunk(document)
//
// Chunking methods:
//   - Fixed size: Split by character/token count
//   - Semantic: Split by meaning boundaries
//   - Recursive: Hierarchical splitting
//
// # Embedding Providers
//
// Supported embedding providers:
//
//   - OpenAI (text-embedding-3-small/large)
//   - Cohere (embed-english-v3.0)
//   - Voyage (voyage-3)
//   - Jina (jina-embeddings-v3)
//   - Google (text-embedding-005)
//   - AWS Bedrock (amazon.titan-embed-text-v2)
//
// # Key Files
//
//   - pipeline.go: Main RAG pipeline
//   - retriever.go: Retrieval implementations
//   - reranker.go: Reranking implementations
//   - chunker.go: Document chunking
//   - hyde.go: HyDE implementation
//   - dense.go: Dense retrieval
//   - sparse.go: Sparse (BM25) retrieval
//
// # Configuration
//
//	config := &rag.Config{
//	    ChunkSize:       512,
//	    ChunkOverlap:    50,
//	    TopK:            10,
//	    DenseWeight:     0.7,
//	    SparseWeight:    0.3,
//	    UseReranking:    true,
//	    RerankTopK:      5,
//	}
//
// # Example: Full RAG Pipeline
//
//	// Create pipeline
//	pipeline := rag.NewHybridPipeline(config)
//
//	// Set embedding provider
//	embedder := embedding.NewOpenAIEmbedder(apiKey)
//	pipeline.SetEmbedder(embedder)
//
//	// Set vector store
//	store := rag.NewQdrantStore(qdrantConfig)
//	pipeline.SetVectorStore(store)
//
//	// Ingest documents
//	docs := loadDocuments()
//	if err := pipeline.Ingest(ctx, docs); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Query
//	results, err := pipeline.Query(ctx, "What is HelixAgent?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use results as context for LLM
//	context := pipeline.AssembleContext(results)
package rag
