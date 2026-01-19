# LSP Servers Research Documentation

## Status: RESEARCHED
**Date**: 2026-01-19

---

## 1. AI-Specific LSP Servers

### LSP-AI

**Repository**: https://github.com/SilasMarvin/lsp-ai
**License**: MIT

**Supported LLM Backends**:
- Local: llama.cpp, Ollama
- Cloud: OpenAI-compatible, Anthropic-compatible, Gemini-compatible
- Specialized: Mistral AI FIM-compatible

**Supported Editors**:
- VS Code
- NeoVim
- Emacs
- Helix
- Sublime Text
- Any LSP-compatible editor

**Core Features**:
1. In-Editor Chatting - Chat with LLMs in codebase context
2. Code Completions - GitHub Copilot alternative
3. Custom Actions - User-defined refactoring and generation

**Roadmap**:
- Semantic search via Tree-sitter
- Additional backends
- Agent-based systems

**Integration Strategy**:
1. Clone as submodule under `third_party/lsp-ai`
2. Build Rust binary during Docker image creation
3. Create LSP-AI adapter in `internal/lsp/servers/lsp_ai.go`
4. Configure backend connections (Ollama, OpenAI, etc.)

### OpenCode LSP

**Features**:
- Automatically loads appropriate LSPs for current LLM
- Supports 75+ LLM providers
- Multi-session support
- Privacy-focused (no code storage)

---

## 2. Language-Specific LSP Servers

### High Priority (Must Implement)

| Language | Server | Repository | Status |
|----------|--------|------------|--------|
| Go | gopls | golang/tools/gopls | EXISTING |
| Rust | rust-analyzer | rust-lang/rust-analyzer | EXISTING |
| Python | pylsp | python-lsp/python-lsp-server | EXISTING |
| TypeScript | typescript-language-server | theia-ide/typescript-language-server | EXISTING |
| C/C++ | clangd | llvm/clangd | NOT_STARTED |
| Java | jdt.ls | eclipse/eclipse.jdt.ls | NOT_STARTED |
| C# | omnisharp-roslyn | OmniSharp/omnisharp-roslyn | NOT_STARTED |

### Medium Priority

| Language | Server | Repository | Status |
|----------|--------|------------|--------|
| Python | pyright | microsoft/pyright | NOT_STARTED |
| PHP | phpactor | phpactor/phpactor | NOT_STARTED |
| Ruby | solargraph | castwide/solargraph | NOT_STARTED |
| Elixir | elixir-ls | elixir-lsp/elixir-ls | NOT_STARTED |
| Haskell | hls | haskell/haskell-language-server | NOT_STARTED |

### Configuration/DevOps

| Language | Server | Repository | Status |
|----------|--------|------------|--------|
| Bash | bash-language-server | bash-lsp/bash-language-server | NOT_STARTED |
| Dockerfile | dockerfile-language-server | rcjsuen/dockerfile-language-server-nodejs | NOT_STARTED |
| YAML | yaml-language-server | redhat-developer/yaml-language-server | NOT_STARTED |
| XML | lemminx | eclipse/lemminx | NOT_STARTED |
| Terraform | terraform-ls | hashicorp/terraform-ls | NOT_STARTED |

### Other Languages

| Language | Server | Repository | Status |
|----------|--------|------------|--------|
| Lua | lua-language-server | sumneko/lua-language-server | NOT_STARTED |
| Scala | metals | scalameta/metals | NOT_STARTED |
| Kotlin | kotlin-language-server | fwcd/kotlin-language-server | NOT_STARTED |
| Swift | sourcekit-lsp | apple/sourcekit-lsp | NOT_STARTED |

---

## 3. MCP-LSP Bridge Servers

### LSP Tools MCP Server

**Features**:
- Regex-based text search
- Directory listing
- Security-focused (explicit directory access)
- Lightweight alternative to full LSP

### mcp-language-server (isaacphi)

**Features**:
- Full LSP protocol bridge
- Find all references
- Go to definition
- Rename symbol
- Semantic code intelligence

### lsp-mcp (Tritlo)

**Features**:
- Deep semantic tools
- Language-aware code intelligence
- Token-level matching

---

## 4. Installation Methods

### Binary Installation (System PATH)

```bash
# Go
go install golang.org/x/tools/gopls@latest

# Rust
rustup component add rust-analyzer

# Python
pip install python-lsp-server

# TypeScript
npm install -g typescript-language-server typescript

# C/C++
apt install clangd  # Debian/Ubuntu
brew install llvm   # macOS

# Java
# Download from Eclipse marketplace

# Bash
npm install -g bash-language-server

# YAML
npm install -g yaml-language-server

# Dockerfile
npm install -g dockerfile-language-server-nodejs
```

### Docker Installation

```dockerfile
# Multi-language LSP container
FROM ubuntu:22.04

# Go
RUN wget https://go.dev/dl/go1.21.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
RUN go install golang.org/x/tools/gopls@latest

# Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
RUN rustup component add rust-analyzer

# Python
RUN pip install python-lsp-server pyright

# TypeScript
RUN npm install -g typescript-language-server typescript

# C/C++
RUN apt install -y clangd

# Bash
RUN npm install -g bash-language-server

# YAML
RUN npm install -g yaml-language-server

# Dockerfile
RUN npm install -g dockerfile-language-server-nodejs
```

---

## 5. LSP Server Registry Design

### Registry Structure (`internal/lsp/servers/registry.go`)

```go
type LSPServerDefinition struct {
    ID            string
    Name          string
    Language      string
    FilePatterns  []string  // *.go, *.rs, *.py, etc.
    Command       string
    Args          []string
    InitOptions   map[string]interface{}
    Capabilities  LSPCapabilities
    Priority      int
}

type LSPCapabilities struct {
    Completion     bool
    Hover          bool
    Definition     bool
    References     bool
    Diagnostics    bool
    Rename         bool
    CodeAction     bool
    Formatting     bool
    SignatureHelp  bool
}
```

### Auto-Detection Logic

```go
func DetectLanguage(filePath string) string {
    ext := filepath.Ext(filePath)
    switch ext {
    case ".go":
        return "go"
    case ".rs":
        return "rust"
    case ".py":
        return "python"
    case ".ts", ".tsx":
        return "typescript"
    case ".js", ".jsx":
        return "javascript"
    case ".c", ".h", ".cpp", ".hpp":
        return "c_cpp"
    case ".java":
        return "java"
    case ".cs":
        return "csharp"
    // ... more mappings
    }
    return "unknown"
}
```

---

## 6. LSP-AI Integration

### Configuration

```yaml
lsp_ai:
  enabled: true
  backends:
    - type: ollama
      url: http://localhost:11434
      model: codellama:7b
    - type: openai
      api_key: ${OPENAI_API_KEY}
      model: gpt-4
    - type: anthropic
      api_key: ${ANTHROPIC_API_KEY}
      model: claude-3-opus
  features:
    chat: true
    completions: true
    custom_actions: true
  fallback_to_local: true
```

### Integration with HelixAgent

1. LSP-AI provides code intelligence to AI debate participants
2. Code completions enhance tool call accuracy
3. Semantic understanding improves context retrieval
4. Custom actions enable AI-driven refactoring

---

## 7. Docker Compose Stack

### lsp-servers-stack.yml

```yaml
version: '3.8'

services:
  lsp-ai:
    build:
      context: ./docker/lsp/lsp-ai
      dockerfile: Dockerfile
    ports:
      - "5000:5000"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - workspace:/workspace:ro

  lsp-multi:
    build:
      context: ./docker/lsp/multi-language
      dockerfile: Dockerfile
    volumes:
      - workspace:/workspace:ro
    command: |
      gopls serve -listen=:5001 &
      rust-analyzer &
      pylsp --tcp --port 5002 &
      typescript-language-server --stdio &
      wait

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_models:/root/.ollama
    deploy:
      resources:
        reservations:
          devices:
            - capabilities: [gpu]

volumes:
  workspace:
  ollama_models:
```

---

## 8. Testing Requirements

### Unit Tests
- Server binary detection
- Connection establishment
- JSON-RPC message formatting
- Response parsing

### Integration Tests
- Full LSP lifecycle (initialize → work → shutdown)
- Multi-language workspace
- Concurrent connections
- Error recovery

### Challenge Scripts

```bash
# lsp_core_challenge.sh
#!/bin/bash
set -euo pipefail

echo "Testing LSP Server Infrastructure..."

# Test Go LSP
echo "Testing gopls..."
timeout 10 gopls version || exit 1

# Test Rust LSP
echo "Testing rust-analyzer..."
timeout 10 rust-analyzer --version || exit 1

# Test Python LSP
echo "Testing pylsp..."
timeout 10 pylsp --version || exit 1

# Test TypeScript LSP
echo "Testing typescript-language-server..."
timeout 10 typescript-language-server --version || exit 1

# Test completion endpoint
echo "Testing completion API..."
curl -X POST http://localhost:8080/v1/lsp/execute \
  -H "Content-Type: application/json" \
  -d '{"serverId": "gopls", "toolName": "completion", "arguments": {...}}'

echo "All LSP tests passed!"
```

---

## 9. Performance Considerations

### Connection Pooling
- Reuse LSP connections across requests
- Idle timeout: 5 minutes
- Max connections per server: 3

### Caching
- Cache completion results (TTL: 30s)
- Cache hover info (TTL: 2m)
- Cache diagnostics (TTL: 10s)
- Invalidate on file change

### Resource Limits
- Memory per server: 512MB max
- CPU: 0.5 cores per server
- Startup timeout: 30s
- Request timeout: 10s

---

## 10. Security Considerations

- Run LSP servers in sandboxed containers
- Read-only filesystem access where possible
- No network access except localhost
- Audit all tool calls
- Rate limit requests per client
