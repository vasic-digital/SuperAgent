# HelixAgent Protocols Comprehensive Guide

This guide covers all supported protocols in HelixAgent: MCP, LSP, ACP, Embeddings, Vision, and Cognee.

## Table of Contents

1. [Overview](#overview)
2. [MCP - Model Context Protocol](#mcp---model-context-protocol)
3. [LSP - Language Server Protocol](#lsp---language-server-protocol)
4. [ACP - Agent Communication Protocol](#acp---agent-communication-protocol)
5. [Embeddings API](#embeddings-api)
6. [Vision API](#vision-api)
7. [Cognee - Knowledge Graph & RAG](#cognee---knowledge-graph--rag)
8. [Integration with AI Debate](#integration-with-ai-debate)
9. [Validation & Testing](#validation--testing)
10. [Troubleshooting](#troubleshooting)

---

## Overview

HelixAgent supports multiple protocols to provide comprehensive AI capabilities:

| Protocol | Purpose | Endpoint Base | Port Range |
|----------|---------|---------------|------------|
| **MCP** | Tool integration & context sharing | `/v1/mcp` | 9101-9999 |
| **LSP** | Language server features | `/v1/lsp` | 9501-9599 |
| **ACP** | Agent-to-agent communication | `/v1/acp` | Built-in |
| **Embeddings** | Vector embeddings | `/v1/embeddings` | Built-in |
| **Vision** | Image analysis & OCR | `/v1/vision` | Built-in |
| **Cognee** | Knowledge graph & RAG | `/v1/cognee` | 8000 |

---

## MCP - Model Context Protocol

### What is MCP?

MCP (Model Context Protocol) is an open protocol that enables AI assistants to connect with external data sources and tools. HelixAgent supports 45+ MCP servers.

### Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        HelixAgent MCP System                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐              │
│   │ MCP Server  │   │ MCP Server  │   │ MCP Server  │   ...        │
│   │ (fetch)     │   │ (git)       │   │ (time)      │              │
│   │ Port 9101   │   │ Port 9102   │   │ Port 9103   │              │
│   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘              │
│          │                 │                 │                      │
│          └────────────┬────┴─────────────────┘                      │
│                       │                                             │
│   ┌───────────────────▼─────────────────────┐                       │
│   │            MCP Router                    │                       │
│   │  (Unified Protocol Manager)              │                       │
│   └───────────────────┬─────────────────────┘                       │
│                       │                                             │
│   ┌───────────────────▼─────────────────────┐                       │
│   │          AI Debate System               │                       │
│   │  (Uses MCP tools as context)            │                       │
│   └─────────────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────────┘
```

### MCP Server Tiers

| Tier | Ports | Servers |
|------|-------|---------|
| **Core Tier** | 9101-9110 | fetch, git, time, filesystem, memory, everything, sequentialthinking |
| **Database Tier** | 9111-9120 | sqlite, postgres, mongodb, mysql, elasticsearch, qdrant |
| **DevOps Tier** | 9121-9130 | docker, kubernetes, aws, gcp, vercel, cloudflare |
| **Productivity Tier** | 9131-9150 | slack, discord, telegram, linear, notion, jira, trello |
| **Search Tier** | 9151-9160 | brave-search, google, youtube, twitter |
| **AI Tier** | 9161-9170 | openai, anthropic |
| **HelixAgent Remote** | 9901-9999 | helixagent-mcp, helixagent-acp, helixagent-lsp |

### Starting MCP Servers

```bash
# Start core MCP servers
./scripts/mcp/start-core-mcp.sh

# Or using make
make mcp-start

# Start specific server
./scripts/mcp/start-mcp-server.sh time

# Verify servers are running
./challenges/scripts/mcp_validation_comprehensive.sh
```

### MCP Protocol Reference

MCP uses JSON-RPC 2.0 over TCP/stdio.

**Session Initialization:**

```json
// Request
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{
  "protocolVersion":"2024-11-05",
  "capabilities":{},
  "clientInfo":{"name":"HelixAgent","version":"1.0.0"}
}}

// Response
{"jsonrpc":"2.0","id":1,"result":{
  "protocolVersion":"2024-11-05",
  "capabilities":{"tools":{}},
  "serverInfo":{"name":"mcp-time","version":"1.0.0"}
}}

// Notification (REQUIRED before tools/list)
{"jsonrpc":"2.0","method":"notifications/initialized"}
```

**List Tools:**

```json
// Request
{"jsonrpc":"2.0","id":2,"method":"tools/list"}

// Response
{"jsonrpc":"2.0","id":2,"result":{
  "tools":[
    {"name":"get_current_time","description":"Get current time in timezone",...}
  ]
}}
```

**Call Tool:**

```json
// Request
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{
  "name":"get_current_time",
  "arguments":{"timezone":"UTC"}
}}

// Response
{"jsonrpc":"2.0","id":3,"result":{
  "content":[{"type":"text","text":"2026-01-27T00:00:00Z"}]
}}
```

### Available MCP Tools

| Server | Tools |
|--------|-------|
| **fetch** | `fetch` - Fetch URL contents |
| **git** | `git_status`, `git_log`, `git_diff`, `git_branch_list`, `git_commit`, `git_add` |
| **time** | `get_current_time` |
| **filesystem** | `read_file`, `write_file`, `list_directory`, `create_directory`, `delete_file` |
| **memory** | `create_entities`, `read_graph`, `search_nodes`, `add_observations`, `delete_entities` |
| **everything** | `search`, `everything_search` - Fast file search |
| **sequentialthinking** | `think`, `create_thinking_session`, `continue_thinking` |

---

## LSP - Language Server Protocol

### What is LSP?

LSP (Language Server Protocol) provides IDE-like features for code analysis, completion, and navigation.

### Supported Language Servers

| Server | Language | Port |
|--------|----------|------|
| gopls | Go | 9501 |
| pyright | Python | 9502 |
| typescript-language-server | TypeScript/JavaScript | 9503 |
| rust-analyzer | Rust | 9504 |
| clangd | C/C++ | 9505 |
| jdtls | Java | 9506 |
| omnisharp | C# | 9507 |
| lua-language-server | Lua | 9508 |

### LSP Capabilities

- **Completion** - Code autocompletion
- **Hover** - Type information on hover
- **Definition** - Go to definition
- **References** - Find all references
- **Diagnostics** - Error detection
- **Formatting** - Code formatting
- **Rename** - Symbol renaming

### LSP API Usage

```bash
# Initialize LSP session
curl -X POST http://localhost:8080/v1/lsp/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "language": "go",
    "rootUri": "file:///path/to/project"
  }'

# Get completions
curl -X POST http://localhost:8080/v1/lsp/completion \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "file:///path/to/file.go",
    "position": {"line": 10, "character": 5}
  }'

# Get hover information
curl -X POST http://localhost:8080/v1/lsp/hover \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "file:///path/to/file.go",
    "position": {"line": 10, "character": 5}
  }'
```

---

## ACP - Agent Communication Protocol

### What is ACP?

ACP (Agent Communication Protocol) enables AI agents to communicate and collaborate on tasks.

### Available Agents

| Agent | Description |
|-------|-------------|
| code-reviewer | Reviews code for issues and improvements |
| bug-finder | Identifies potential bugs |
| refactor-assistant | Suggests refactoring improvements |
| documentation-generator | Generates documentation |
| test-generator | Generates test cases |
| security-scanner | Scans for security vulnerabilities |

### ACP API Usage

```bash
# List available agents
curl http://localhost:8080/v1/acp/agents

# Execute agent task
curl -X POST http://localhost:8080/v1/acp/execute \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "code-reviewer",
    "task": "Review this code for issues",
    "context": {
      "code": "func add(a, b int) int { return a + b }",
      "language": "go"
    },
    "timeout": 30
  }'

# Get agent status
curl http://localhost:8080/v1/acp/agents/code-reviewer
```

---

## Embeddings API

### Overview

The Embeddings API provides vector embeddings for semantic search and RAG applications.

### Supported Providers

| Provider | Models | Dimensions |
|----------|--------|------------|
| **OpenAI** | text-embedding-3-small, text-embedding-3-large | 512-3072 |
| **Cohere** | embed-english-v3.0, embed-multilingual-v3.0 | 384-4096 |
| **Voyage** | voyage-3, voyage-3-lite, voyage-code-3 | 512-1536 |
| **Jina** | jina-embeddings-v3, jina-clip-v1 | 128-1024 |
| **Google** | text-embedding-005, textembedding-gecko | 768 |
| **Bedrock** | amazon.titan-embed-text-v2 | 1024-1536 |

### API Usage

```bash
# Generate embeddings
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "text-embedding-3-small",
    "input": ["Hello, world!", "This is a test."]
  }'

# Response
{
  "embeddings": [
    [0.123, -0.456, ...],
    [0.789, -0.012, ...]
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

### Batch Embedding

```bash
# Batch embed multiple texts
curl -X POST http://localhost:8080/v1/embeddings/batch \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "text-embedding-3-small",
    "inputs": [
      {"id": "doc1", "text": "First document"},
      {"id": "doc2", "text": "Second document"}
    ]
  }'
```

---

## Vision API

### Overview

The Vision API provides image analysis, OCR, object detection, and more.

### Capabilities

| Capability | Description |
|------------|-------------|
| **analyze** | General image analysis |
| **ocr** | Optical Character Recognition |
| **detect** | Object detection |
| **caption** | Image captioning |
| **describe** | Detailed description |
| **classify** | Image classification |
| **segment** | Image segmentation |

### API Usage

```bash
# Analyze image (base64)
curl -X POST http://localhost:8080/v1/vision/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "analyze",
    "image": "<base64-encoded-image>",
    "prompt": "Describe what you see in this image"
  }'

# Analyze image (URL)
curl -X POST http://localhost:8080/v1/vision/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "analyze",
    "image_url": "https://example.com/image.png",
    "prompt": "Describe this image"
  }'

# OCR
curl -X POST http://localhost:8080/v1/vision/ocr \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "ocr",
    "image": "<base64-encoded-image>",
    "prompt": "Extract all text from this image"
  }'

# Object detection
curl -X POST http://localhost:8080/v1/vision/detect \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "detect",
    "image": "<base64-encoded-image>",
    "prompt": "Detect all objects"
  }'
```

---

## Cognee - Knowledge Graph & RAG

### Overview

Cognee provides knowledge graph construction and RAG (Retrieval-Augmented Generation) capabilities.

### API Usage

```bash
# Add knowledge
curl -X POST http://localhost:8080/v1/cognee/add \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Your knowledge content here",
    "metadata": {"source": "document.pdf"}
  }'

# Process knowledge into graph
curl -X POST http://localhost:8080/v1/cognee/cognify \
  -H "Content-Type: application/json" \
  -d '{}'

# Search knowledge
curl -X POST http://localhost:8080/v1/cognee/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "search query",
    "top_k": 5
  }'
```

---

## Integration with AI Debate

### How Protocols Integrate with AI Debate

All protocols can provide context to the AI Debate system:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        AI Debate System                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌────────────┐  ┌────────────┐  ┌────────────┐                   │
│   │ MCP Tools  │  │ LSP Data   │  │ Embeddings │                   │
│   │ (context)  │  │ (analysis) │  │ (semantic) │                   │
│   └─────┬──────┘  └─────┬──────┘  └─────┬──────┘                   │
│         │               │               │                           │
│         └───────────────┼───────────────┘                           │
│                         │                                           │
│   ┌─────────────────────▼─────────────────────┐                     │
│   │              Context Builder              │                     │
│   └─────────────────────┬─────────────────────┘                     │
│                         │                                           │
│   ┌─────────────────────▼─────────────────────┐                     │
│   │              AI Debate Engine             │                     │
│   │  • Multi-round discussion                 │                     │
│   │  • Multi-pass validation                  │                     │
│   │  • Consensus building                     │                     │
│   └───────────────────────────────────────────┘                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Debate with MCP Context

```bash
curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "What is the best approach to organize files?",
    "mcp_context": {
      "tool_results": [
        {
          "server": "filesystem",
          "tool": "list_directory",
          "result": {"files": ["doc1.txt", "doc2.txt"]}
        }
      ]
    },
    "enable_multi_pass_validation": true
  }'
```

---

## Validation & Testing

### Running Validations

```bash
# Run all protocol validations
./challenges/scripts/all_protocols_validation.sh

# Run individual validations
./challenges/scripts/mcp_validation_comprehensive.sh
./challenges/scripts/lsp_validation_comprehensive.sh
./challenges/scripts/acp_validation_comprehensive.sh
./challenges/scripts/embeddings_validation_comprehensive.sh
./challenges/scripts/vision_validation_comprehensive.sh
```

### Go Integration Tests

```bash
# Run all integration tests
go test -v ./internal/testing/integration/...

# Run MCP tests
go test -v ./internal/testing/mcp/...

# Run MCP + Debate integration tests
go test -v ./internal/testing/integration/ -run TestMCPDebateIntegration
```

### Validation Summary

| Protocol | Tests | Description |
|----------|-------|-------------|
| MCP | 26 | TCP, Protocol, Tools, Execution |
| LSP | 21 | TCP, Protocol, Capabilities |
| ACP | 14 | Health, Agents, Execution |
| Embeddings | 17 | Providers, Generation, Quality |
| Vision | 15 | Capabilities, Analysis, OCR |

---

## Troubleshooting

### Common Issues

#### MCP Server Not Responding

```bash
# Check if server is running
nc -z localhost 9103 && echo "Time server is running"

# Check server logs
podman logs mcp-time

# Restart server
podman restart mcp-time
```

#### Protocol Compliance Failure

```bash
# Test manual JSON-RPC
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | nc localhost 9103
```

#### LSP Connection Issues

```bash
# Check language server process
ps aux | grep gopls

# Test LSP initialize
curl -X POST http://localhost:8080/v1/lsp/health
```

#### Embeddings API Key Issues

```bash
# Verify API key is set
echo $OPENAI_API_KEY | head -c 10

# Test with curl
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{"provider":"openai","model":"text-embedding-3-small","input":["test"]}'
```

### Getting Help

- **Documentation**: `/docs/` directory
- **Issues**: https://github.com/helixagent/helixagent/issues
- **Logs**: Check server logs at `/tmp/helixagent.log`

---

## Next Steps

1. **Start HelixAgent**: `make run`
2. **Start MCP Servers**: `./scripts/mcp/start-core-mcp.sh`
3. **Run Validations**: `./challenges/scripts/all_protocols_validation.sh`
4. **Explore API**: Use curl or the provided client libraries
5. **Integrate with AI Debate**: Create debates with protocol context

For more detailed information, see the individual protocol documentation in the `/docs/` directory.
