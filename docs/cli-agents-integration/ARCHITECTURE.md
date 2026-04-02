# CLI Agents Integration Architecture

## Overview

HelixAgent integrates with 47 CLI agents through a unified configuration system. Each agent connects to HelixAgent's OpenAI-compatible API to access the AI Debate Ensemble and 45+ MCP servers.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLI AGENT ECOSYSTEM                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │ Claude Code  │  │    Aider     │  │   Codex      │  │   OpenHands     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └────────┬────────┘  │
│         │                 │                 │                   │          │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐  ┌───────┴────────┐  │
│  │  Cline       │  │  Gemini CLI  │  │  Amazon Q    │  │  GPT-Engineer  │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └────────┬───────┘  │
│         │                 │                 │                   │          │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐  ┌───────┴────────┐  │
│  │    Forge     │  │    GPTMe     │  │  Kilo-Code   │  │     Warp       │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └────────┬───────┘  │
│         │                 │                 │                   │          │
│  ┌──────┴─────────────────┴─────────────────┴───────────────────┴────────┐  │
│  │                         47 CLI AGENTS                                  │  │
│  └─────────────────────────────────┬──────────────────────────────────────┘  │
│                                    │                                         │
│                                    ▼                                         │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                     HELIXAGENT API SERVER                              │  │
│  │                    http://localhost:7061/v1                            │  │
│  │                                                                        │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐  │  │
│  │  │   /chat/    │  │   /mcp/     │  │   /lsp/     │  │   /acp/      │  │  │
│  │  │ completions │  │   tools     │  │  language   │  │ agent comm   │  │  │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬───────┘  │  │
│  │         │                │                │                │         │  │
│  └─────────┼────────────────┼────────────────┼────────────────┼─────────┘  │
│            │                │                │                │            │
└────────────┼────────────────┼────────────────┼────────────────┼────────────┘
             │                │                │                │
             ▼                ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         HELIXAGENT CORE SERVICES                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    AI DEBATE ENSEMBLE                                │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │    │
│  │  │  Claude  │ │ DeepSeek │ │  Gemini  │ │  GPT-4   │ │  Others  │  │    │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘  │    │
│  │       └─────────────┴─────────────┴─────────────┴──────────────────┘    │
│  │                              │                                          │
│  │                         Debate Engine                                   │
│  │                    (Consensus Algorithm)                                │
│  └──────────────────────────────┬──────────────────────────────────────────┘    │
│                                 │                                            │
│  ┌──────────────────────────────┼──────────────────────────────────────────┐ │
│  │                    MCP SERVERS (45+)                                    │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │ │
│  │  │filesystem│ │  browser │ │   git    │ │  memory  │ │  sqlite  │      │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘      │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │ │
│  │  │embedding │ │   rag    │ │  vision  │ │  format  │ │   lsp    │      │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘      │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │              FORMATTERS (32+ Programming Languages)                     │ │
│  │  Python | JavaScript | Go | Rust | Java | C++ | Ruby | etc.           │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Configuration Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Configuration Loading Flow                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Agent reads cli_agents_configs/<agent>.json                 │
│         │                                                        │
│         ▼                                                        │
│  2. Extract provider settings (base_url, api_key_env)           │
│         │                                                        │
│         ▼                                                        │
│  3. Extract MCP server configurations                           │
│         │                                                        │
│         ▼                                                        │
│  4. Extract formatter preferences                               │
│         │                                                        │
│         ▼                                                        │
│  5. Load agent-specific settings (profiles, extensions)         │
│         │                                                        │
│         ▼                                                        │
│  6. Merge with environment variables                            │
│         │                                                        │
│         ▼                                                        │
│  7. Configure HTTP client with HelixAgent endpoint              │
│         │                                                        │
│         ▼                                                        │
│  8. Initialize MCP clients (local + remote)                     │
│         │                                                        │
│         ▼                                                        │
│  9. Agent ready to use HelixAgent capabilities                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Request Flow
```
User Input
    │
    ▼
CLI Agent Interface
    │
    ▼
Format Request (OpenAI-compatible)
    │
    ▼
HTTP/3 + Brotli Compression
    │
    ▼
HelixAgent API Server
    │
    ├──► AI Debate Ensemble
    │       ├──► Multiple LLMs
    │       └──► Consensus Algorithm
    │
    ├──► MCP Tool Calls (if needed)
    │       ├──► Filesystem
    │       ├──► Browser
    │       └──► Other tools
    │
    └──► Formatter Service (if needed)
            └──► 32+ formatters
    │
    ▼
Response Aggregation
    │
    ▼
Return to CLI Agent
    │
    ▼
Display to User
```

## Component Details

### 1. Provider Configuration
All agents use OpenAI-compatible provider:
- **Base URL**: `http://localhost:7061/v1`
- **Model**: `helixagent-debate` (AI Debate Ensemble)
- **Max Tokens**: 128,000
- **Capabilities**: Vision, Streaming, Function Calls, Embeddings, MCP, ACP, LSP

### 2. MCP Servers
Each agent can be configured with multiple MCP servers:

**HelixAgent Remote MCPs:**
- `helixagent-mcp` - Core MCP server
- `helixagent-lsp` - Language Server Protocol
- `helixagent-acp` - Agent Communication Protocol
- `helixagent-embeddings` - Embedding generation
- `helixagent-vision` - Vision/image analysis
- `helixagent-rag` - RAG retrieval
- `helixagent-cognee` - Memory/knowledge graph

**Local MCPs (via npx):**
- `filesystem` - File operations
- `browser` - Browser automation (Puppeteer)
- `git` - Git operations
- `memory` - Persistent memory
- `sqlite` - Database operations
- `sequential-thinking` - Chain-of-thought

### 3. Formatters
32+ code formatters available:
- **Python**: ruff, black, autopep8, yapf
- **JavaScript/TypeScript**: biome, prettier, dprint
- **Go**: gofmt
- **Rust**: rustfmt
- **Java**: google-java-format, spotless
- **C/C++**: clang-format
- **Ruby**: rubocop, standardrb
- **And more...**

### 4. Agent Profiles
Some agents support multiple profiles:
- `default` - General purpose
- `coder` - Code generation
- `architect` - System design
- `reviewer` - Code review

## Security Considerations

1. **API Key Management**: Use environment variables, never commit keys
2. **Local MCP Execution**: Local MCP servers run with user permissions
3. **Remote MCP Communication**: HTTPS/WSS with authentication
4. **Sandboxing**: Consider containerization for untrusted MCP servers

## Performance Optimizations

1. **HTTP/3**: All communication uses HTTP/3 (QUIC) with Brotli compression
2. **Connection Pooling**: Persistent connections to HelixAgent
3. **Streaming**: Real-time response streaming for better UX
4. **Caching**: Response caching for repeated queries

## Extension Points

New agents can be added by:
1. Creating config in `cli_agents_configs/`
2. Following the schema in `docs/cli-agents-integration/CONFIG_SCHEMA.md`
3. Testing integration with HelixAgent

---

**Last Updated:** 2026-04-02
