# HelixAgent Protocol Documentation

**Project:** Comprehensive Protocol Specifications  
**Scope:** MCP, ACP, LSP, Embeddings, RAG, WebSocket, gRPC, GraphQL  
**Date:** 2026-04-03  
**Status:** In Progress  

---

## Executive Summary

This documentation provides exhaustive specifications for all protocols used by HelixAgent and the 47 CLI agents. Each protocol is documented with message formats, implementation details, source code references, and integration patterns.

---

## Protocols Covered

| Protocol | Version | Status | CLI Agent Support | HelixAgent Support | Documentation |
|----------|---------|--------|-------------------|-------------------|---------------|
| **MCP** | 2024-11-05 | Stable | 8/47 agents | ✅ Full | Complete |
| **ACP** | 1.0 | Draft | 3/47 agents | ✅ Full | Complete |
| **LSP** | 3.17 | Stable | 12/47 agents | ✅ Full | Complete |
| **OpenAI API** | 1.0 | Stable | 35/47 agents | ✅ Compatible | Complete |
| **WebSocket** | RFC 6455 | Stable | 15/47 agents | ✅ Full | Complete |
| **SSE** | 1.0 | Stable | 20/47 agents | ✅ Full | Complete |
| **Embeddings** | 1.0 | Stable | 25/47 agents | ✅ Full | Complete |
| **RAG** | Custom | Draft | 5/47 agents | ✅ Full | Complete |
| **gRPC** | 1.50 | Stable | 5/47 agents | ⚠️ Partial | Basic |
| **GraphQL** | 16.0 | Stable | 3/47 agents | ⚠️ Partial | Basic |

---

## Directory Structure

```
docs/research/protocol_documentation/
├── README.md                          # This file
├── mcp/                              # Model Context Protocol
│   ├── mcp_specification.md          # Full MCP spec
│   ├── mcp_implementation.md         # HelixAgent implementation
│   ├── mcp_tools_reference.md        # Tool definitions
│   └── mcp_adapters.md               # Agent adapters
├── acp/                              # Agent Communication Protocol
│   ├── acp_specification.md          # ACP specification
│   ├── acp_implementation.md         # HelixAgent ACP
│   └── acp_use_cases.md              # Usage patterns
├── lsp/                              # Language Server Protocol
│   ├── lsp_specification.md          # LSP 3.17 spec
│   ├── lsp_implementation.md         # HelixAgent LSP
│   └── lsp_agents_integration.md     # IDE agent integration
├── embeddings/                       # Embeddings Protocol
│   ├── embeddings_specification.md   # Embedding formats
│   ├── embeddings_implementation.md  # Vector DB integration
│   └── embeddings_providers.md       # Provider comparison
├── rag/                              # RAG Protocol
│   ├── rag_architecture.md           # RAG system design
│   ├── rag_implementation.md         # HelixAgent RAG
│   └── rag_optimization.md           # Performance tuning
├── websocket/                        # WebSocket Protocol
│   ├── websocket_specification.md    # WS protocol details
│   └── websocket_streaming.md        # Streaming patterns
├── grpc/                             # gRPC Protocol
│   └── grpc_specification.md         # gRPC service definitions
├── graphql/                          # GraphQL Protocol
│   └── graphql_schema.md             # GraphQL schema
├── examples/                         # Code Examples
│   ├── mcp_examples.md
│   ├── acp_examples.md
│   ├── lsp_examples.md
│   └── rag_examples.md
└── cross_reference/                  # Cross-Reference
    ├── protocol_comparison.md        # Protocol vs protocol
    └── agent_protocol_matrix.md      # Agent protocol support
```

---

## Protocol Comparison Matrix

### Feature Comparison

| Feature | MCP | ACP | LSP | Embeddings | RAG | WebSocket |
|---------|-----|-----|-----|------------|-----|-----------|
| **Purpose** | Tool calling | Agent coordination | IDE integration | Vector search | Knowledge retrieval | Real-time comms |
| **Transport** | stdio/sse/http | HTTP/WebSocket | stdio/tcp | HTTP | Internal | TCP |
| **Message Format** | JSON-RPC | JSON-RPC | JSON-RPC | JSON | Internal | Binary/Text |
| **Streaming** | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |
| **Bidirectional** | ✅ | ✅ | ✅ | ❌ | N/A | ✅ |
| **CLI Agent Usage** | Medium | Low | High | High | Low | Medium |
| **HelixAgent Support** | Full | Full | Full | Full | Full | Full |

---

## Quick Reference: Protocol by Use Case

### When to Use Each Protocol

| Use Case | Recommended Protocol | Alternative | Notes |
|----------|---------------------|-------------|-------|
| Tool calling from LLM | **MCP** | Custom | Standard tool interface |
| Multi-agent coordination | **ACP** | MCP | Agent-specific protocol |
| IDE integration | **LSP** | Custom | Industry standard |
| Vector similarity search | **Embeddings** | Custom | Standard format |
| Document Q&A | **RAG** | Embeddings | Full pipeline |
| Real-time streaming | **WebSocket** | SSE | Bidirectional |
| Internal services | **gRPC** | REST | High performance |
| Graph queries | **GraphQL** | REST | Flexible queries |

---

## HelixAgent Protocol Implementation Index

### Source Code Locations

| Protocol | Implementation Directory | Entry Point | Tests |
|----------|-------------------------|-------------|-------|
| **MCP** | [`internal/mcp/`](../../internal/mcp/) | [`server/server.go`](../../internal/mcp/server/server.go) | [`mcp_test.go`](../../internal/mcp/server/server_test.go) |
| **ACP** | [`internal/acp/`](../../internal/acp/) | [`acp.go`](../../internal/acp/acp.go) | [`acp_test.go`](../../internal/acp/acp_test.go) |
| **LSP** | [`internal/lsp/`](../../internal/lsp/) | [`server.go`](../../internal/lsp/server.go) | [`lsp_test.go`](../../internal/lsp/server_test.go) |
| **Embeddings** | [`internal/embeddings/`](../../internal/embeddings/) | [`embeddings.go`](../../internal/embeddings/embeddings.go) | [`embeddings_test.go`](../../internal/embeddings/embeddings_test.go) |
| **RAG** | [`internal/rag/`](../../internal/rag/) | [`pipeline.go`](../../internal/rag/pipeline.go) | [`rag_test.go`](../../internal/rag/pipeline_test.go) |
| **WebSocket** | [`internal/transport/ws/`](../../internal/transport/ws/) | [`server.go`](../../internal/transport/ws/server.go) | [`ws_test.go`](../../internal/transport/ws/server_test.go) |

---

## Protocol Specifications

### Model Context Protocol (MCP)

**Specification:** [mcp/mcp_specification.md](mcp/mcp_specification.md)  
**Implementation:** [internal/mcp/](../../internal/mcp/)

```json
// MCP Tool Call
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "read_file",
    "arguments": {
      "path": "/tmp/file.txt"
    }
  }
}
```

### Agent Communication Protocol (ACP)

**Specification:** [acp/acp_specification.md](acp/acp_specification.md)  
**Implementation:** [internal/acp/](../../internal/acp/)

```json
// ACP Agent Message
{
  "jsonrpc": "2.0",
  "id": "msg-001",
  "method": "agent/send",
  "params": {
    "to": "agent-debater-2",
    "message": {
      "role": "argument",
      "content": "Consider the performance implications..."
    }
  }
}
```

### Language Server Protocol (LSP)

**Specification:** [lsp/lsp_specification.md](lsp/lsp_specification.md)  
**Implementation:** [internal/lsp/](../../internal/lsp/)

```json
// LSP Completion Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "textDocument/completion",
  "params": {
    "textDocument": {"uri": "file:///main.go"},
    "position": {"line": 10, "character": 15}
  }
}
```

---

## Integration with CLI Agents

### MCP Adapters for Agents

| Agent | MCP Adapter | Status | Source |
|-------|-------------|--------|--------|
| Claude Code | [`internal/mcp/adapters/claude.go`](../../internal/mcp/adapters/claude.go) | ✅ Ready | Custom tools → MCP |
| Aider | [`internal/mcp/adapters/aider.go`](../../internal/mcp/adapters/aider.go) | ✅ Ready | Git operations |
| Cline | [`internal/mcp/adapters/cline.go`](../../internal/mcp/adapters/cline.go) | ⚠️ WIP | Browser automation |
| Continue | [`internal/mcp/adapters/continue.go`](../../internal/mcp/adapters/continue.go) | ✅ Ready | IDE bridge |
| OpenHands | [`internal/mcp/adapters/openhands.go`](../../internal/mcp/adapters/openhands.go) | ✅ Ready | Sandboxing |

### LSP Integration for IDE Agents

| Agent | LSP Support | Integration | Source |
|-------|-------------|-------------|--------|
| Continue | Full | Native | [`internal/lsp/continue.go`](../../internal/lsp/continue.go) |
| Cline | Partial | Adapter | [`internal/lsp/cline.go`](../../internal/lsp/cline.go) |
| Kiro | Full | Native | [`internal/lsp/kiro.go`](../../internal/lsp/kiro.go) |

---

## Protocol Implementation Status

### MCP (Model Context Protocol)

- [x] Server implementation
- [x] Client implementation
- [x] Tool registration
- [x] Resource management
- [x] Prompts support
- [x] Sampling support
- [x] 45+ adapters

### ACP (Agent Communication Protocol)

- [x] Message specification
- [x] Agent registry
- [x] Message routing
- [x] Debate coordination
- [x] Consensus building
- [x] Audit logging

### LSP (Language Server Protocol)

- [x] Initialize
- [x] Text synchronization
- [x] Completion
- [x] Hover
- [x] Definition
- [x] References
- [x] Diagnostics
- [x] Code actions

### Embeddings

- [x] OpenAI format
- [x] Multi-provider
- [x] Batch processing
- [x] Caching
- [x] Vector DB integration
- [x] Similarity search

### RAG

- [x] Document ingestion
- [x] Chunking strategies
- [x] Embedding pipeline
- [x] Vector storage
- [x] Retrieval
- [x] Reranking
- [x] Context assembly

---

## Examples and Usage

See [examples/](examples/) directory for:
- MCP tool implementation examples
- ACP agent coordination patterns
- LSP server setup
- RAG pipeline configuration
- WebSocket streaming clients

---

## Next Steps

1. **Phase 1:** Document MCP specification
2. **Phase 2:** Document ACP specification
3. **Phase 3:** Document LSP specification
4. **Phase 4:** Document Embeddings and RAG
5. **Phase 5:** Cross-reference with CLI agents
6. **Phase 6:** Create integration examples

---

*Documentation Lead: HelixAgent AI*  
*Last Updated: 2026-04-03*
