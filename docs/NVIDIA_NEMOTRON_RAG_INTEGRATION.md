# NVIDIA Nemotron RAG Integration for HelixAgent

## Overview

HelixAgent integrates with NVIDIA's Nemotron RAG (Retrieval-Augmented Generation) models to provide state-of-the-art document processing, multimodal understanding, and grounded answer generation. This integration enables HelixAgent to parse complex PDFs, extract nested tables, understand charts and diagrams, and provide traceable citations for regulated industries.

**Key Capabilities:**
- Multimodal document processing (text, tables, charts, images)
- GPU-accelerated extraction using NeMo Retriever
- Structured embeddings with Nemotron Embed VL
- Precision reranking with Nemotron Reranker VL
- Citation-backed answer generation with Nemotron Super

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        HelixAgent Nemotron RAG Pipeline                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │   Document   │───▶│  Extraction  │───▶│   Embedding  │───▶│ Reranker │  │
│  │    Input     │    │  (NeMo VLM)  │    │(Nemotron VL) │    │(Cross-  │  │
│  │  (PDF/Image) │    │              │    │              │    │ Encoder) │  │
│  └──────────────┘    └──────────────┘    └──────────────┘    └────┬─────┘  │
│                                                                    │        │
│  ┌─────────────────────────────────────────────────────────────────┘        │
│  │                                                                          │
│  ▼                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                   │
│  │   Vector     │───▶│   Nemotron   │───▶│   Grounded   │                   │
│  │   Database   │    │    Super     │    │   Response   │                   │
│  │  (Milvus)    │    │  (49B LLM)   │    │  + Citations │                   │
│  └──────────────┘    └──────────────┘    └──────────────┘                   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Components

### 1. Document Extraction (NeMo Retriever)

The extraction stage converts complex documents from "pixels and layout" into structured, queryable units using GPU-accelerated microservices.

**Input:** PDF files with mixed content (text, tables, charts)

**Output:** Structured JSON with:
- Text chunks with preserved formatting
- Tables in Markdown format
- Chart/figure crops for multimodal understanding
- Page-level metadata for citations

**Key Features:**
- **Structural preservation:** Tables maintain row/column relationships
- **Multimodal extraction:** Charts and diagrams are preserved as images
- **OCR capabilities:** Handles scanned documents
- **PDFium backend:** High-performance PDF parsing

**HelixAgent Integration:**
```go
// Example: Document extraction through HelixAgent
result, err := helixAgent.ProcessDocument(ctx, &DocumentRequest{
    Source: "financial_report.pdf",
    Extractors: []ExtractorType{
        ExtractorText,
        ExtractorTables,
        ExtractorCharts,
    },
    TableOutputFormat: "markdown",
    Provider: ProviderNVIDIA,
})
```

---

### 2. Multimodal Embedding (Nemotron Embed VL 1B)

Converts extracted content into 2048-dimensional vectors for semantic search.

**Capabilities:**
- **Text-only embedding:** Standard document chunks
- **Image-only embedding:** Chart and diagram understanding
- **Multimodal embedding:** Combined text + image encoding

**Model Details:**
- Model ID: `nvidia/llama-nemotron-embed-vl-1b-v2`
- Vector dimension: 2048
- Supports batch processing
- GPU-accelerated via NVIDIA NIM

**HelixAgent Integration:**
```go
// Example: Generate embeddings
embedding, err := helixAgent.GenerateEmbedding(ctx, &EmbeddingRequest{
    Content: "Quarterly revenue increased by 23%...",
    Image: chartImage, // Optional
    Model: "nvidia/llama-nemotron-embed-vl-1b-v2",
    Modality: ModalityImageText, // or ModalityText, ModalityImage
})
```

---

### 3. Cross-Encoder Reranking (Nemotron Reranker VL 1B)

Precision layer that evaluates query-document relevance using vision-language understanding.

**Why Reranking Matters:**
- Filters out "looks similar but wrong" retrieval results
- Evaluates (query, document, optional image) together
- Critical for enterprise PDFs with charts and tables

**Model Details:**
- Model ID: `nvidia/llama-nemotron-rerank-vl-1b-v2`
- Cross-encoder architecture
- VLM-based relevance scoring

**HelixAgent Integration:**
```go
// Example: Rerank retrieved documents
ranked, err := helixAgent.Rerank(ctx, &RerankRequest{
    Query: "What was the Q3 revenue growth?",
    Candidates: retrievedDocs,
    Model: "nvidia/llama-nemotron-rerank-vl-1b-v2",
    TopK: 5,
})
```

---

### 4. Answer Generation (Nemotron Super 49B)

Generates grounded, cited answers with strict adherence to source material.

**Key Capabilities:**
- **Source-grounded responses:** Every claim backed by citations
- **Uncertainty admission:** Recognizes when information is insufficient
- **Traceability:** Provides exact page/section references
- **Long context:** Handles extensive retrieved documents

**Model Details:**
- Model ID: `nvidia/llama-3.3-nemotron-super-49b-v1.5`
- 49 billion parameters
- Optimized for RAG with citations
- Available via NVIDIA NIM

**HelixAgent Integration:**
```go
// Example: Generate cited answer
response, err := helixAgent.Generate(ctx, &GenerationRequest{
    Query: "Explain the revenue trends in Q3 2024",
    Context: rankedDocuments,
    Model: "nvidia/llama-3.3-nemotron-super-49b-v1.5",
    RequireCitations: true,
    CitationFormat: CitationFormatPageSection,
})
// Response includes: "According to Section 4.2, Page 47..."
```

---

## Configuration

### Environment Variables

```bash
# NVIDIA API Configuration
NVIDIA_API_KEY=nvapi-xxxxxxxxxxxxxxxx
NVIDIA_BASE_URL=https://integrate.api.nvidia.com/v1

# Model Endpoints
NVIDIA_EMBED_MODEL=nvidia/llama-nemotron-embed-vl-1b-v2
NVIDIA_RERANK_MODEL=nvidia/llama-nemotron-rerank-vl-1b-v2
NVIDIA_GENERATION_MODEL=nvidia/llama-3.3-nemotron-super-49b-v1.5

# Vector Database
MILVUS_URI=milvus.db
MILVUS_COLLECTION=nemotron_rag

# Processing Options
NVIDIA_CHUNK_SIZE=512
NVIDIA_CHUNK_OVERLAP=100
NVIDIA_EXTRACT_TABLES=true
NVIDIA_EXTRACT_CHARTS=true
```

### HelixAgent Configuration

```yaml
# config/nemotron.yaml
rag:
  provider: nvidia
  
  extraction:
    library: nemo_retriever
    mode: library  # or "container" for production
    extractors:
      - text
      - tables
      - charts
    table_format: markdown
    
  embedding:
    model: nvidia/llama-nemotron-embed-vl-1b-v2
    dimension: 2048
    batch_size: 32
    
  reranking:
    model: nvidia/llama-nemotron-rerank-vl-1b-v2
    top_k: 10
    batch_size: 8
    
  generation:
    model: nvidia/llama-3.3-nemotron-super-49b-v1.5
    temperature: 0.3
    max_tokens: 4096
    require_citations: true
    
  vector_store:
    type: milvus
    uri: ${MILVUS_URI}
    collection: ${MILVUS_COLLECTION}
```

---

## Use Cases

### 1. Financial Document Analysis

**Challenge:** Complex financial reports with tables, charts, and conditional statements.

**Solution:**
```go
result := helixAgent.RAGQuery(ctx, &RAGQueryRequest{
    Document: "annual_report.pdf",
    Query: "What are the risk factors if revenue drops below $10M?",
    Provider: ProviderNVIDIA,
    Options: &RAGOptions{
        ExtractTables: true,
        ExtractCharts: true,
        RequireCitations: true,
    },
})
// Returns answer with citations like "Section 3.1 Risk Factors, Page 12"
```

### 2. Regulatory Compliance

**Challenge:** Need precise citations for audit trails in regulated industries.

**Solution:**
```go
result := helixAgent.RAGQuery(ctx, &RAGQueryRequest{
    Document: "compliance_manual.pdf",
    Query: "What are the reporting requirements for data breaches?",
    Provider: ProviderNVIDIA,
    Options: &RAGOptions{
        CitationLevel: CitationLevelParagraph,
        IncludePageNumbers: true,
        IncludeSectionHeaders: true,
    },
})
```

### 3. Technical Documentation

**Challenge:** Understanding diagrams, flowcharts, and conditional logic across pages.

**Solution:**
```go
result := helixAgent.RAGQuery(ctx, &RAGQueryRequest{
    Document: "system_architecture.pdf",
    Query: "Explain the failover process when the primary node fails",
    Provider: ProviderNVIDIA,
    Options: &RAGOptions{
        ExtractCharts: true,     // Process architecture diagrams
        ExtractImages: true,     // Include screenshots
        CrossPageContext: true,  // Connect related sections
    },
})
```

---

## Best Practices

### Chunk Size Tradeoffs

| Chunk Size | Precision | Context | Best For |
|------------|-----------|---------|----------|
| 256-512 tokens | High | Limited | Fact retrieval, specific data points |
| 512-1024 tokens | Balanced | Balanced | **Recommended for enterprise docs** |
| 1024-2048 tokens | Lower | High | Narrative understanding, summaries |

**Recommendation:** Use 512-1024 tokens with 100-200 token overlap for enterprise documents.

### Extraction Depth

- **Page-level splitting:** Enables precise citations and verification
- **Document-level splitting:** Maintains narrative flow and broader context

**Recommendation:** Use page-level for compliance/audit scenarios, document-level for research/analysis.

### Table Output Format

- **Markdown:** Preserves row/column relationships, LLM-native
- **JSON:** Structured data for programmatic access
- **CSV:** Compatibility with data tools

**Recommendation:** Use Markdown for RAG pipelines (reduces numeric hallucinations).

### Deployment Modes

| Mode | Use Case | Scaling |
|------|----------|---------|
| Library Mode | Development, <100 docs | Single node |
| Container Mode | Production, 1000s of docs | Horizontal scaling with Redis/Kafka |

---

## API Reference

### Document Processing

```go
// ProcessDocument extracts and structures document content
func (c *Client) ProcessDocument(ctx context.Context, req *DocumentRequest) (*DocumentResult, error)

// IndexDocument adds document to vector store
func (c *Client) IndexDocument(ctx context.Context, req *IndexRequest) (*IndexResult, error)
```

### RAG Query

```go
// RAGQuery performs full retrieval-augmented generation
func (c *Client) RAGQuery(ctx context.Context, req *RAGQueryRequest) (*RAGResult, error)

// RAGResult contains the response with citations
type RAGResult struct {
    Answer      string
    Citations   []Citation
    Sources     []SourceDocument
    Confidence  float64
}
```

### Citation Format

```go
type Citation struct {
    Text        string   // Quoted text from source
    Page        int      // Page number
    Section     string   // Section header
    Paragraph   int      // Paragraph number
    SourceDoc   string   // Document name
}
```

---

## Troubleshooting

### Common Issues

**Issue:** Poor table extraction quality
- **Solution:** Ensure `table_output_format="markdown"` and use PDFium extraction method

**Issue:** Missing chart/diagram understanding
- **Solution:** Enable `extract_charts=true` and verify GPU memory (minimum 24GB)

**Issue:** Citations missing or incorrect
- **Solution:** Use page-level extraction and enable `require_citations=true`

**Issue:** Slow processing on large documents
- **Solution:** Switch to Container Mode with Redis/Kafka for distributed processing

### Performance Optimization

1. **GPU Memory:** Ensure at least 24GB VRAM for local model deployment
2. **Batch Size:** Adjust embedding/reranking batch sizes based on GPU memory
3. **Chunk Overlap:** Use 100-200 token overlap to maintain context across chunks
4. **Caching:** Enable result caching for repeated queries

---

## Resources

### Models on Hugging Face
- [nvidia/llama-nemotron-embed-vl-1b-v2](https://huggingface.co/nvidia/llama-nemotron-embed-vl-1b-v2) - Multimodal embedding
- [nvidia/llama-nemotron-rerank-vl-1b-v2](https://huggingface.co/nvidia/llama-nemotron-rerank-vl-1b-v2) - Cross-encoder reranker
- [Nemotron RAG Collection](https://huggingface.co/collections/nvidia/nemotron-rag) - Extraction models

### Cloud Endpoints
- [Nemotron OCR Document Extraction](https://build.nvidia.com/nvidia/nemotron-ocr)
- [Llama 3.3 Nemotron Super 49B](https://build.nvidia.com/nvidia/llama-3_3-nemotron-super-49b-v1)
- [NVIDIA NIM Models](https://build.nvidia.com/explore/discover)

### Code and Documentation
- [NeMo Retriever Library](https://github.com/NVIDIA/NeMo-Retriever)
- [Tutorial Notebook](https://github.com/NVIDIA/NeMo-Retriever/tree/main/tutorials)
- [NVIDIA Blueprint for Enterprise RAG](https://build.nvidia.com/nvidia/nemotron-rag)

---

## References

This documentation incorporates content from:
- [NVIDIA Developer Blog: How to Build a Document Processing Pipeline for RAG with Nemotron](https://developer.nvidia.com/blog/how-to-build-a-document-processing-pipeline-for-rag-with-nemotron/)

---

## License

NVIDIA models and software are subject to their respective licenses. HelixAgent integration is provided under the project's main license.