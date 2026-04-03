# Aider: API Documentation & HelixAgent Cross-Reference

**Agent:** Aider  
**Type:** CLI-only (No public API)  
**Primary LLMs:** GPT-4, Claude, DeepSeek (multi-provider)  
**HelixAgent Equivalent:** Ensemble + Git MCP  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

Aider is a **CLI-only tool** with no public API. It communicates directly with LLM providers (OpenAI, Anthropic, etc.) using their native APIs. Aider's uniqueness lies in its git-native workflow and repository mapping, not its API design. HelixAgent can replicate and extend Aider's functionality through ensemble providers and git MCP adapters.

**HelixAgent Alternative:** Use ensemble with Git MCP for equivalent functionality.

---

## Aider Architecture

### Internal Communication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                       AIDER ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Terminal → Aider Core → Multiple LLM APIs                     │
│       │                        │                                │
│       │                        ├── OpenAI API (GPT-4)           │
│       │                        ├── Anthropic API (Claude)       │
│       │                        ├── DeepSeek API                 │
│       │                        └── OpenRouter (100+ models)     │
│       │                                                          │
│       ▼                                                          │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │                    AIDER COMPONENTS                        │  │
│   │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │  │
│   │  │   Repo Map   │  │   Diff       │  │   Git        │   │  │
│   │  │   Generator  │  │   Parser     │  │   Interface  │   │  │
│   │  └──────────────┘  └──────────────┘  └──────────────┘   │  │
│   └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│   No public API endpoint exposed!                                │
│   All intelligence is client-side                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Repository Map System

Aider's unique feature is the **Repository Map** - an AST-based understanding of the codebase:

```python
# Source: aider/repo.py (conceptual)
class RepoMap:
    """
    Generates a map of the repository for context
    """
    
    def __init__(self, root_path):
        self.files = self._scan_repository()
        self.symbols = self._extract_symbols_tree_sitter()
        self.references = self._build_reference_graph()
        self.imports = self._analyze_imports()
    
    def get_ranked_tags(self, query, max_tokens=8000):
        """
        Returns most relevant code symbols for the query
        """
        ranked = self._rank_by_relevance(query)
        return self._format_context(ranked, max_tokens)
```

---

## HelixAgent Equivalent Implementation

### Source Code Reference

**Git MCP Adapter:**
- File: [`internal/mcp/adapters/git.go`](../../../internal/mcp/adapters/git.go)
- Operations: [`internal/mcp/adapters/git_operations.go`](../../../internal/mcp/adapters/git_operations.go)
- Diff Handler: [`internal/handlers/diff.go`](../../../internal/handlers/diff.go)

**Repository Analysis:**
- Repo Parser: [`internal/code/repo_parser.go`](../../../internal/code/repo_parser.go)
- Symbol Extractor: [`internal/code/symbols.go`](../../../internal/code/symbols.go)
- Import Analyzer: [`internal/code/imports.go`](../../../internal/code/imports.go)

**Ensemble Providers:**
- Provider Registry: [`internal/services/provider_registry.go`](../../../internal/services/provider_registry.go)
- Ensemble Engine: [`internal/services/ensemble.go`](../../../internal/services/ensemble.go)

### Repository Map Implementation

**Source:** [`internal/code/repo_parser.go`](../../../internal/code/repo_parser.go)

```go
package code

// RepoMap provides Aider-like repository understanding
type RepoMap struct {
    rootPath  string
    files     []FileInfo
    symbols   map[string][]Symbol
    references map[string][]Reference
}

// GenerateMap creates repository context for LLM
// Source: internal/code/repo_parser.go#L45-89
func (rm *RepoMap) GenerateMap(ctx context.Context, query string, maxTokens int) (*RepoContext, error) {
    // 1. Parse all source files
    // 2. Extract symbols using tree-sitter
    // 3. Build reference graph
    // 4. Rank by relevance to query
    // 5. Format for LLM context window
}
```

---

## API Comparison: Aider vs HelixAgent

Since Aider has no API, we compare internal capabilities:

| Capability | Aider Implementation | HelixAgent Implementation | Advantage |
|------------|---------------------|--------------------------|-----------|
| **Repo Map** | Python + tree-sitter | Go + tree-sitter | Tie |
| **Diff Format** | Unified diffs | Unified + custom | HelixAgent |
| **Multi-Provider** | 15+ via config | 22+ via registry | HelixAgent |
| **Git Integration** | Native GitPython | MCP Git adapter | Aider (native) |
| **Commit Attribution** | Automatic | Via MCP | Aider (native) |
| **Multi-File Edits** | ✅ Native | ✅ Via ensemble | Tie |
| **Ensemble Voting** | ❌ | ✅ | HelixAgent |
| **Debate** | ❌ | ✅ | HelixAgent |
| **API Server** | ❌ | ✅ | HelixAgent |
| **Persistence** | Git only | PostgreSQL | HelixAgent |
| **Caching** | ❌ | Redis | HelixAgent |
| **Rate Limiting** | ❌ | ✅ | HelixAgent |

---

## Diff Format Deep Dive

### Aider's SEARCH/REPLACE Format

Aider uses a unique diff format optimized for LLMs:

```
<<<<<<< SEARCH
old code to find
=======
new code to replace
>>>>>>> REPLACE
```

**Source:** Aider's `coders/editblock_coder.py` (conceptual)

### HelixAgent Diff Implementation

**Source:** [`internal/handlers/diff.go`](../../../internal/handlers/diff.go)

```go
package handlers

// DiffRequest represents a multi-file edit request
type DiffRequest struct {
    Files []FileEdit `json:"files"`
}

type FileEdit struct {
    Path    string `json:"path"`
    Search  string `json:"search"`   // Aider-style SEARCH
    Replace string `json:"replace"`  // Aider-style REPLACE
}

// ApplyDiff applies Aider-style diffs
// Source: internal/handlers/diff.go#L45-120
func ApplyDiff(ctx context.Context, req *DiffRequest) (*DiffResult, error) {
    results := make([]FileResult, 0, len(req.Files))
    
    for _, edit := range req.Files {
        // Read original file
        content, err := os.ReadFile(edit.Path)
        if err != nil {
            return nil, err
        }
        
        // Apply SEARCH/REPLACE
        newContent := strings.Replace(string(content), edit.Search, edit.Replace, 1)
        
        // Write back
        err = os.WriteFile(edit.Path, []byte(newContent), 0644)
        
        results = append(results, FileResult{
            Path:   edit.Path,
            Status: "modified",
        })
    }
    
    return &DiffResult{Files: results}, nil
}
```

### API Endpoint for Diff Operations

**Source:** [`internal/handlers/diff.go:45`](../../../internal/handlers/diff.go#L45)

```
POST /v1/diff/apply
```

**Request:**
```json
{
  "files": [
    {
      "path": "src/main.py",
      "search": "def old_function():\n    pass",
      "replace": "def new_function():\n    return True"
    }
  ],
  "commit": true,
  "commit_message": "refactor: Update function"
}
```

**Response:**
```json
{
  "success": true,
  "files_modified": 1,
  "files": [
    {
      "path": "src/main.py",
      "status": "modified",
      "lines_changed": 2
    }
  ],
  "commit": {
    "hash": "a1b2c3d",
    "message": "refactor: Update function"
  }
}
```

---

## Git Operations API

### Aider Git Integration

Aider uses GitPython for deep git integration:

```python
# Conceptual: aider/git.py
class GitIntegration:
    def commit_changes(self, message, files=None):
        """Commit changes with proper attribution"""
        self.repo.index.add(files or [])
        commit = self.repo.index.commit(message)
        return commit.hexsha
    
    def get_repo_map(self):
        """Generate repository structure"""
        return self._analyze_tree()
```

### HelixAgent Git MCP

**Source:** [`internal/mcp/adapters/git.go`](../../../internal/mcp/adapters/git.go)

```go
package adapters

// GitAdapter implements git operations via MCP
type GitAdapter struct {
    repo *git.Repository
}

// MCP Tools exposed:
// - git/status     : Get working tree status
// - git/diff       : Get diffs
// - git/commit     : Create commits
// - git/log        : Get commit history
// - git/branch     : Branch operations
// - git/checkout   : Checkout files/branches

// Execute implements MCP tool execution
// Source: internal/mcp/adapters/git.go#L78-156
func (g *GitAdapter) Execute(ctx context.Context, tool string, params map[string]interface{}) (*Result, error) {
    switch tool {
    case "git/commit":
        return g.commit(params)
    case "git/diff":
        return g.diff(params)
    case "git/status":
        return g.status(params)
    // ...
    }
}
```

### Git API Endpoints

**Source:** [`internal/handlers/git.go`](../../../internal/handlers/git.go)

```
GET  /v1/git/status              # Working tree status
POST /v1/git/commit              # Create commit
GET  /v1/git/diff                # Get diff
GET  /v1/git/log                 # Commit history
POST /v1/git/branch              # Create branch
POST /v1/git/checkout            # Checkout
```

**Example:**
```bash
# Get status
curl http://localhost:7061/v1/git/status

# Create commit
curl -X POST http://localhost:7061/v1/git/commit \
  -d '{
    "message": "feat: Add new feature",
    "files": ["src/main.py"],
    "author": "Aider <aider@example.com>"
  }'
```

---

## Multi-Provider Support Comparison

### Aider Provider Configuration

**File:** `.aider.conf.yml`

```yaml
# Aider supports 15+ providers via config
openai:
  api_key: ${OPENAI_API_KEY}
  model: gpt-4

anthropic:
  api_key: ${ANTHROPIC_API_KEY}
  model: claude-3-5-sonnet

deepseek:
  api_key: ${DEEPSEEK_API_KEY}
  model: deepseek-chat

openrouter:
  api_key: ${OPENROUTER_API_KEY}
  model: openai/gpt-4
```

### HelixAgent Provider Registry

**Source:** [`internal/services/provider_registry.go`](../../../internal/services/provider_registry.go)

```go
package services

// ProviderRegistry manages all LLM providers
type ProviderRegistry struct {
    providers map[string]llm.Provider
    ensemble  *Ensemble
}

// Register adds a provider
// Source: internal/services/provider_registry.go#L45-67
func (r *ProviderRegistry) Register(name string, provider llm.Provider) error {
    r.providers[name] = provider
    return nil
}

// Get returns provider by name
func (r *ProviderRegistry) Get(name string) (llm.Provider, error) {
    if p, ok := r.providers[name]; ok {
        return p, nil
    }
    return nil, ErrProviderNotFound
}

// GetAll returns all registered providers
func (r *ProviderRegistry) GetAll() []llm.Provider {
    // Used for ensemble voting
}
```

### Provider API

**Source:** [`internal/handlers/providers.go`](../../../internal/handlers/providers.go)

```
GET  /v1/providers              # List all providers
GET  /v1/providers/{name}       # Get provider info
POST /v1/providers/{name}/chat  # Chat with specific provider
```

---

## Ensemble vs Single Provider

### Aider: Single Provider

```python
# Aider uses one provider per session
class AiderCoder:
    def __init__(self, model):
        self.model = model  # Single model
    
    def generate(self, prompt):
        return self.model.complete(prompt)  # One response
```

### HelixAgent: Ensemble

**Source:** [`internal/services/ensemble.go`](../../../internal/services/ensemble.go)

```go
package services

// Ensemble coordinates multiple providers
type Ensemble struct {
    providers []llm.Provider
    strategy  VotingStrategy
}

// Execute queries all providers and votes
// Source: internal/services/ensemble.go#L56-120
func (e *Ensemble) Execute(ctx context.Context, req *CompletionRequest) (*EnsembleResult, error) {
    // 1. Query all providers concurrently
    responses := make(chan ProviderResponse, len(e.providers))
    
    for _, provider := range e.providers {
        go func(p llm.Provider) {
            resp, err := p.Generate(ctx, req)
            responses <- ProviderResponse{Provider: p.Name(), Response: resp, Error: err}
        }(provider)
    }
    
    // 2. Collect all responses
    var results []ProviderResponse
    for i := 0; i < len(e.providers); i++ {
        results = append(results, <-responses)
    }
    
    // 3. Vote on best response
    winner := e.strategy.Vote(results)
    
    return &EnsembleResult{
        Winner:    winner,
        AllVotes:  results,
        Confidence: e.calculateConfidence(results),
    }, nil
}
```

### Ensemble API

```
POST /v1/ensemble/completions
```

**Request:**
```json
{
  "providers": ["claude", "gpt4", "deepseek"],
  "messages": [{"role": "user", "content": "Refactor this function"}],
  "voting_strategy": "best_of_n",
  "min_confidence": 0.8
}
```

**Response:**
```json
{
  "winner": {
    "provider": "claude",
    "content": "// Refactored code...",
    "confidence": 0.92
  },
  "all_votes": [
    {"provider": "claude", "confidence": 0.92},
    {"provider": "gpt4", "confidence": 0.85},
    {"provider": "deepseek", "confidence": 0.78}
  ],
  "consensus": true
}
```

---

## WebSocket Support

### Aider: No WebSocket

Aider is CLI-only with no real-time streaming API.

### HelixAgent WebSocket

**Source:** [`internal/handlers/websocket.go`](../../../internal/handlers/websocket.go)

```go
// WebSocketHandler manages real-time connections
// Source: internal/handlers/websocket.go#L30-85

ws://localhost:7061/v1/stream
```

**Protocol for Git Operations:**
```json
// Client → Server: Start git operation
{
  "type": "git_operation",
  "id": "op-123",
  "operation": "commit",
  "params": {
    "message": "feat: Add feature",
    "files": ["src/main.py"]
  }
}

// Server → Client: Progress updates
{
  "type": "progress",
  "id": "op-123",
  "status": "staging_files",
  "progress": 50
}

// Server → Client: Completion
{
  "type": "complete",
  "id": "op-123",
  "result": {
    "commit_hash": "a1b2c3d",
    "files_changed": 1
  }
}
```

---

## Source Code Reference Index

### Aider-Related HelixAgent Files

| Aider Feature | HelixAgent Implementation | File | Lines |
|---------------|--------------------------|------|-------|
| Repo Map | Repo Parser | `internal/code/repo_parser.go` | 145 |
| Symbol Extraction | Symbol Analyzer | `internal/code/symbols.go` | 98 |
| Diff Format | Diff Handler | `internal/handlers/diff.go` | 120 |
| Git Integration | Git MCP Adapter | `internal/mcp/adapters/git.go` | 156 |
| Git Operations | Git Operations | `internal/mcp/adapters/git_operations.go` | 134 |
| Multi-Provider | Provider Registry | `internal/services/provider_registry.go` | 189 |
| Ensemble | Ensemble Engine | `internal/services/ensemble.go` | 167 |
| Provider Config | Provider Config | `internal/config/providers.go` | 234 |
| Commit Attribution | Git Handler | `internal/handlers/git.go` | 87 |
| File Watching | File Watcher | `internal/utils/file_watcher.go` | 76 |

---

## Integration Guide: Aider → HelixAgent Migration

### Configuration Mapping

**Aider `.aider.conf.yml` → HelixAgent `config.yaml`:**

```yaml
# Aider config
openai:
  api_key: ${OPENAI_API_KEY}
  model: gpt-4

# HelixAgent equivalent
providers:
  openai:
    type: openai
    api_key: ${OPENAI_API_KEY}
    model: gpt-4
    priority: 1
```

### Command Mapping

| Aider Command | HelixAgent Equivalent | API Endpoint |
|---------------|----------------------|--------------|
| `aider` | Start server | `docker-compose up` |
| `/add file.py` | MCP add | `POST /v1/mcp/filesystem/add` |
| `/commit` | Git commit | `POST /v1/git/commit` |
| `/diff` | Get diff | `GET /v1/git/diff` |
| `/undo` | Git revert | `POST /v1/git/revert` |
| `/model gpt-4` | Switch provider | `POST /v1/providers/switch` |

### Migration Script

```bash
#!/bin/bash
# migrate-aider-to-helixagent.sh

# 1. Export Aider config
export OPENAI_API_KEY=$(grep api_key .aider.conf.yml | cut -d: -f2)

# 2. Start HelixAgent
docker-compose up -d helixagent

# 3. Configure providers
curl -X POST http://localhost:7061/v1/admin/providers \
  -d "{
    \"openai\": {\"api_key\": \"$OPENAI_API_KEY\"},
    \"anthropic\": {\"api_key\": \"$ANTHROPIC_API_KEY\"}
  }"

# 4. Enable ensemble for better results
curl -X POST http://localhost:7061/v1/admin/ensemble \
  -d '{"enabled": true, "providers": ["openai", "anthropic"]}'
```

---

## Strengths & Weaknesses

### Aider Strengths

1. **Git-Native Excellence**
   - Automatic commit attribution
   - Diff-based editing
   - Repository map
   - Best-in-class git integration

2. **Multi-Provider CLI**
   - 15+ providers supported
   - Easy switching
   - No server needed

3. **Repository Understanding**
   - AST-based mapping
   - Symbol extraction
   - Import analysis

### HelixAgent Strengths

1. **Ensemble Intelligence**
   - Multi-model voting
   - Better accuracy
   - Consensus building

2. **API-First Design**
   - REST API
   - WebSocket streaming
   - CI/CD integration

3. **Enterprise Features**
   - Authentication
   - Rate limiting
   - Audit trails

### When to Use Which

| Use Case | Winner | Reason |
|----------|--------|--------|
| Git-native workflow | Aider | Automatic commits |
| Repository mapping | Aider | Built-in AST parsing |
| Multi-file edits | Tie | Both support |
| Ensemble decisions | HelixAgent | Voting system |
| API integration | HelixAgent | REST/WebSocket |
| Team deployment | HelixAgent | Multi-user |
| CI/CD automation | HelixAgent | API access |

---

## Conclusion

Aider is a **CLI-only tool** focused on git-native workflows. Its strengths are:
- Automatic commit attribution
- Repository map (AST-based)
- Diff-based editing

**HelixAgent provides equivalent functionality through:**
- Git MCP adapter
- Repository parser
- Diff handler

**Plus additional capabilities:**
- Ensemble voting
- 22+ providers
- API server
- Enterprise features

**Recommendation:** Use HelixAgent for team/enterprise scenarios; Aider remains excellent for individual git-centric workflows.

---

*Documentation: API Specification & Cross-Reference*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: 7ec2da53*
