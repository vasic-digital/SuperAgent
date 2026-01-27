# Video Course 11: HelixAgent Protocols Integration

**Duration**: 120 minutes (6 lessons)
**Level**: Intermediate to Advanced
**Prerequisites**: Basic HelixAgent knowledge, Understanding of APIs

---

## Course Overview

This comprehensive course covers all HelixAgent protocols: MCP, LSP, ACP, Embeddings, Vision, and Cognee. You'll learn how to use each protocol independently and how they integrate with the AI Debate system.

---

## Lesson 1: Introduction to HelixAgent Protocols (15 min)

### Video Script

```
[INTRO - 0:00]
Welcome to HelixAgent Protocols Integration! In this course, you'll master all
HelixAgent protocols and learn how to leverage them for powerful AI applications.

[SLIDE - Protocol Overview - 0:30]
HelixAgent supports six major protocols:
- MCP - Model Context Protocol for tool integration
- LSP - Language Server Protocol for code intelligence
- ACP - Agent Communication Protocol for agent collaboration
- Embeddings - Vector embeddings for semantic search
- Vision - Image analysis and OCR
- Cognee - Knowledge graph and RAG

[DEMO - Architecture Diagram - 2:00]
Let me show you how these protocols work together...
[Show architecture diagram]

All protocols feed into the AI Debate system, providing rich context for
intelligent discussions between multiple AI models.

[SLIDE - Port Allocation - 5:00]
Each protocol has its own port range:
- MCP: 9101-9999 (45+ servers)
- LSP: 9501-9599 (8 language servers)
- ACP, Embeddings, Vision, Cognee: Built into HelixAgent core

[DEMO - Starting Services - 8:00]
Let's start HelixAgent and the MCP servers:

$ make run
$ ./scripts/mcp/start-core-mcp.sh

Now let's verify everything is running:
$ curl http://localhost:8080/health
$ ./challenges/scripts/mcp_validation_comprehensive.sh

[SUMMARY - 13:00]
In this lesson, we covered:
- Overview of all six protocols
- Architecture and how protocols integrate
- Port allocation
- Starting services

[NEXT - 14:30]
In the next lesson, we'll dive deep into MCP - the Model Context Protocol.
```

---

## Lesson 2: MCP - Model Context Protocol Deep Dive (25 min)

### Video Script

```
[INTRO - 0:00]
Welcome back! In this lesson, we'll master MCP - the Model Context Protocol.
MCP is the bridge between AI and external tools.

[SLIDE - What is MCP? - 0:30]
MCP is an open protocol that enables AI assistants to:
- Connect to external data sources
- Execute tools
- Share context across sessions

HelixAgent supports 45+ MCP servers!

[DEMO - MCP Server Tiers - 2:00]
Let me show you our MCP server organization:

Core Tier (9101-9110):
- fetch (9101) - HTTP requests
- git (9102) - Git operations
- time (9103) - Time utilities
- filesystem (9104) - File operations
- memory (9105) - Knowledge graph
- everything (9106) - Fast search
- sequentialthinking (9107) - Reasoning

[DEMO - Starting MCP Servers - 5:00]
Let's start the core MCP servers:

$ ./scripts/mcp/start-core-mcp.sh

Watch as each server starts in its container...
[Show podman ps output]

[DEMO - MCP Protocol - 8:00]
MCP uses JSON-RPC 2.0 over TCP. Let me show you the protocol:

Step 1: Initialize session
$ echo '{"jsonrpc":"2.0","id":1,"method":"initialize",...}' | nc localhost 9103

Step 2: Send initialized notification
$ echo '{"jsonrpc":"2.0","method":"notifications/initialized"}' | nc localhost 9103

Step 3: List tools
$ echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | nc localhost 9103

Step 4: Call a tool
$ echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_current_time","arguments":{"timezone":"UTC"}}}' | nc localhost 9103

[DEMO - Using MCP via HelixAgent - 15:00]
You don't have to do this manually! HelixAgent handles it:

$ curl -X POST http://localhost:8080/v1/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "server": "time",
    "tool": "get_current_time",
    "arguments": {"timezone": "UTC"}
  }'

[DEMO - MCP Validation - 20:00]
Let's run the comprehensive MCP validation:

$ ./challenges/scripts/mcp_validation_comprehensive.sh

[Show output with 26/26 tests passing]

[SUMMARY - 23:00]
In this lesson, we covered:
- MCP protocol fundamentals
- Server tiers and organization
- JSON-RPC protocol
- Using MCP via HelixAgent API
- Validation testing

[NEXT - 24:30]
Next, we'll explore LSP for code intelligence!
```

---

## Lesson 3: LSP - Language Server Protocol (20 min)

### Video Script

```
[INTRO - 0:00]
Welcome to Lesson 3! Today we're covering LSP - the Language Server Protocol.
LSP brings IDE-like code intelligence to HelixAgent.

[SLIDE - What is LSP? - 0:30]
LSP provides:
- Code completion
- Hover information
- Go to definition
- Find references
- Error detection
- Code formatting

Originally created by Microsoft for VS Code, now an industry standard.

[SLIDE - Supported Languages - 2:00]
HelixAgent supports 8 language servers:
- Go (gopls) - Port 9501
- Python (pyright) - Port 9502
- TypeScript/JavaScript - Port 9503
- Rust (rust-analyzer) - Port 9504
- C/C++ (clangd) - Port 9505
- Java (jdtls) - Port 9506
- C# (omnisharp) - Port 9507
- Lua - Port 9508

[DEMO - Starting LSP Servers - 4:00]
Let's start the Go language server:

$ podman run -d --name lsp-gopls -p 9501:9501 helixagent/lsp-gopls

[DEMO - LSP Protocol - 6:00]
LSP uses JSON-RPC 2.0 with Content-Length headers:

Content-Length: 217

{"jsonrpc":"2.0","id":1,"method":"initialize","params":{
  "processId":null,
  "rootUri":"file:///path/to/project",
  "capabilities":{"textDocument":{"completion":{}}}
}}

[DEMO - Using LSP via HelixAgent - 10:00]
HelixAgent provides a unified LSP interface:

# Get completions
$ curl -X POST http://localhost:8080/v1/lsp/completion \
  -H "Content-Type: application/json" \
  -d '{
    "language": "go",
    "uri": "file:///path/to/main.go",
    "position": {"line": 10, "character": 5}
  }'

# Get hover info
$ curl -X POST http://localhost:8080/v1/lsp/hover \
  -H "Content-Type: application/json" \
  -d '{
    "language": "go",
    "uri": "file:///path/to/main.go",
    "position": {"line": 10, "character": 5}
  }'

[DEMO - LSP Validation - 15:00]
Run the LSP validation:

$ ./challenges/scripts/lsp_validation_comprehensive.sh

[SUMMARY - 18:00]
In this lesson, we covered:
- LSP fundamentals
- Supported language servers
- Protocol details
- Using LSP via HelixAgent

[NEXT - 19:30]
Next up: ACP for agent collaboration!
```

---

## Lesson 4: Embeddings, Vision & Cognee (25 min)

### Video Script

```
[INTRO - 0:00]
Welcome to Lesson 4! Today we cover three powerful protocols:
Embeddings, Vision, and Cognee.

[SLIDE - Embeddings Overview - 0:30]
Embeddings convert text to vectors for semantic search.

Supported providers:
- OpenAI (text-embedding-3-small/large)
- Cohere (embed-english-v3.0)
- Voyage (voyage-3)
- Jina (jina-embeddings-v3)
- Google (text-embedding-005)
- AWS Bedrock (titan-embed-text-v2)

[DEMO - Generating Embeddings - 3:00]
Let's generate some embeddings:

$ curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "text-embedding-3-small",
    "input": ["Hello world", "Semantic search is powerful"]
  }'

[Show response with embedding vectors]

[SLIDE - Vision Overview - 8:00]
Vision API provides image analysis capabilities:
- analyze - General analysis
- ocr - Text extraction
- detect - Object detection
- caption - Image captioning
- classify - Classification
- segment - Segmentation

[DEMO - Using Vision API - 10:00]
Let's analyze an image:

$ curl -X POST http://localhost:8080/v1/vision/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "analyze",
    "image_url": "https://example.com/image.png",
    "prompt": "Describe what you see"
  }'

# OCR example
$ curl -X POST http://localhost:8080/v1/vision/ocr \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "ocr",
    "image": "<base64-encoded-image>"
  }'

[SLIDE - Cognee Overview - 15:00]
Cognee provides:
- Knowledge graph construction
- RAG (Retrieval-Augmented Generation)
- Entity extraction
- Semantic search over knowledge

[DEMO - Using Cognee - 17:00]
# Add knowledge
$ curl -X POST http://localhost:8080/v1/cognee/add \
  -d '{"content": "HelixAgent is an AI orchestration system."}'

# Process into graph
$ curl -X POST http://localhost:8080/v1/cognee/cognify

# Search
$ curl -X POST http://localhost:8080/v1/cognee/search \
  -d '{"query": "What is HelixAgent?"}'

[DEMO - Validation - 22:00]
$ ./challenges/scripts/embeddings_validation_comprehensive.sh
$ ./challenges/scripts/vision_validation_comprehensive.sh

[SUMMARY - 24:00]
We covered:
- Embeddings for semantic search
- Vision for image analysis
- Cognee for knowledge graphs

[NEXT - 24:30]
Next: Integrating all protocols with AI Debate!
```

---

## Lesson 5: Protocol Integration with AI Debate (20 min)

### Video Script

```
[INTRO - 0:00]
Welcome to Lesson 5! This is where it all comes together.
We'll integrate all protocols with the AI Debate system.

[SLIDE - Integration Architecture - 0:30]
[Show architecture diagram]

The AI Debate system can receive context from ALL protocols:
- MCP tool results
- LSP code analysis
- Embeddings for semantic relevance
- Vision for image understanding
- Cognee for knowledge retrieval

[DEMO - Debate with MCP Context - 3:00]
Let's create a debate with MCP tool context:

# First, get some MCP data
$ curl -X POST http://localhost:8080/v1/mcp/tools/call \
  -d '{"server":"time","tool":"get_current_time","arguments":{"timezone":"UTC"}}'

# Now create a debate with this context
$ curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Why is time management important?",
    "mcp_context": {
      "tool_results": [{
        "server": "time",
        "tool": "get_current_time",
        "result": {"time": "2026-01-27T00:00:00Z"}
      }]
    },
    "enable_multi_pass_validation": true
  }'

[Show debate response with consensus]

[DEMO - Debate with Embeddings - 8:00]
Let's use embeddings for semantic context:

$ curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Compare functional and object-oriented programming",
    "embedding_context": {
      "provider": "openai",
      "queries": ["functional programming benefits", "OOP advantages"]
    }
  }'

[DEMO - Full Integration - 12:00]
Now let's combine everything:

$ curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "How should we organize this codebase?",
    "mcp_context": {
      "tool_results": [{
        "server": "filesystem",
        "tool": "list_directory",
        "result": {"files": ["main.go", "handlers/", "services/"]}
      }]
    },
    "lsp_context": {
      "language": "go",
      "analysis": ["code_structure", "dependencies"]
    },
    "cognee_context": {
      "query": "best practices for Go project structure"
    },
    "enable_multi_pass_validation": true
  }'

[DEMO - Run Integration Tests - 16:00]
$ go test -v ./internal/testing/integration/ -run TestMCPDebateIntegration

[SUMMARY - 18:00]
We covered:
- Integration architecture
- Debates with MCP context
- Debates with embedding context
- Full multi-protocol integration

[NEXT - 19:30]
Final lesson: Production deployment and best practices!
```

---

## Lesson 6: Production Deployment & Best Practices (15 min)

### Video Script

```
[INTRO - 0:00]
Welcome to the final lesson! We'll cover production deployment
and best practices for HelixAgent protocols.

[SLIDE - Production Checklist - 0:30]
Before going to production:
✓ All validations passing (./challenges/scripts/all_protocols_validation.sh)
✓ API keys configured for all providers
✓ Container health checks enabled
✓ Monitoring and logging configured
✓ Rate limiting enabled
✓ Authentication configured

[DEMO - Running All Validations - 2:00]
$ ./challenges/scripts/all_protocols_validation.sh

[Show complete validation output]

[SLIDE - Best Practices - 5:00]
1. Use environment variables for all API keys
2. Enable container health checks
3. Set appropriate timeouts
4. Implement retry logic
5. Monitor resource usage
6. Log all protocol operations

[DEMO - Docker Compose Production - 8:00]
# Use the production compose file
$ docker-compose -f docker-compose.production.yml up -d

# Verify all services
$ docker-compose ps
$ curl http://localhost:8080/health

[SLIDE - Monitoring - 11:00]
Monitor these metrics:
- MCP tool call latency
- LSP response times
- Embedding generation throughput
- Vision API usage
- Debate completion rates

[SUMMARY - 13:00]
Congratulations! You've completed the Protocols Integration course!

You now know:
- All six HelixAgent protocols
- How to use each protocol
- Integration with AI Debate
- Production deployment

[OUTRO - 14:00]
Thank you for taking this course!
Check out the documentation at /docs/user-guides/PROTOCOLS_COMPREHENSIVE_GUIDE.md
for more details.

Happy building with HelixAgent!
```

---

## Supplementary Materials

### Hands-On Exercises

1. **MCP Exercise**: Create a custom MCP tool that returns system information
2. **LSP Exercise**: Implement code completion for a custom language
3. **Integration Exercise**: Build a debate that uses all protocols

### Quizzes

1. What port range is used for MCP servers?
2. Which LSP method is used for code completion?
3. How do you pass MCP context to a debate?

### Resources

- Documentation: `/docs/user-guides/PROTOCOLS_COMPREHENSIVE_GUIDE.md`
- Challenge Scripts: `/challenges/scripts/`
- Integration Tests: `/internal/testing/integration/`

---

## Course Assets

- Slide deck: `11-protocols-slides.pptx`
- Demo scripts: `11-protocols-demo.sh`
- Exercise solutions: `11-protocols-exercises-solutions.zip`
