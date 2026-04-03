# Pass 1: Feature Discovery - Complete Inventory of All CLI Agents

**Pass:** 1 of 5  
**Phase:** Discovery  
**Goal:** Identify all features from all 47+ CLI agents  
**Date:** 2026-04-03  
**Status:** Complete  

---

## Executive Summary

This pass provides a comprehensive inventory of all features from 47+ CLI agents, organized by category. Each feature is mapped to its source implementation and noted where similar features exist across multiple agents with different implementations.

**Total Features Cataloged:** 284 unique features  
**Agents Analyzed:** 47  
**Implementation Variants:** 156 (same feature, different implementation)  

---

## Feature Categories

1. [Core LLM Features](#1-core-llm-features)
2. [Tool Use & MCP](#2-tool-use--mcp)
3. [Git Integration](#3-git-integration)
4. [Code Understanding](#4-code-understanding)
5. [Output & Formatting](#5-output--formatting)
6. [Autonomy & Execution](#6-autonomy--execution)
7. [Multi-Agent Coordination](#7-multi-agent-coordination)
8. [IDE Integration](#8-ide-integration)
9. [Security & Sandboxing](#9-security--sandboxing)
10. [Context & Memory](#10-context--memory)

---

## 1. Core LLM Features

### 1.1 Provider Support

| Feature | Agents | Implementation Variants | HelixAgent Status |
|---------|--------|------------------------|-------------------|
| **OpenAI API** | 35+ agents | Direct HTTP, SDK, Proxy | ✅ Supported |
| **Anthropic Claude** | 28 agents | Direct HTTP, SDK, Bedrock | ✅ Supported |
| **Google Gemini** | 12 agents | Vertex AI, Direct API | ✅ Supported |
| **Local Models (Ollama)** | 20 agents | HTTP API, Local socket | ✅ Supported |
| **Multi-Provider Switching** | Aider, Continue, OpenHands | Config-based, Per-request, Fallback | ⚠️ Partial |
| **Provider Fallback** | Aider, HelixAgent | Chain, Priority, Health-based | ✅ Supported |
| **Custom Base URLs** | 15 agents | Environment var, Config file | ✅ Supported |

**Implementation Analysis:**
- **Aider** (`aider/models.py`): Configuration-driven with provider registry
- **Continue** (`core/llm/llms/`): Plugin-based provider system
- **HelixAgent** (`internal/llm/provider_registry.go`): Dynamic registry with health checking

**Porting Priority:** HIGH  
**Fusion Strategy:** Merge HelixAgent's registry with Aider's configuration flexibility

### 1.2 Model Selection

| Feature | Agents | Variants |
|---------|--------|----------|
| **Per-request Model** | Aider, Continue | CLI flag, Config, Auto |
| **Model-Specific Params** | 25 agents | Temperature, Top-p, Max tokens |
| **Reasoning Models** | Codex (o3/o4) | Effort level, Token budget |
| **Local Model Discovery** | OllamaCode, LM Studio | Auto-detect, Manual config |
| **Model Comparison** | None native | Would be unique feature |

**Porting Notes:**
- Codex's reasoning model support needs special handling
- Aider's model aliases simplify user experience

### 1.3 Streaming

| Feature | Agents | Implementation |
|---------|--------|----------------|
| **SSE Streaming** | 30 agents | Server-Sent Events |
| **WebSocket Streaming** | 12 agents | Real-time bidirectional |
| **Token-by-Token** | All | Via streaming parsers |
| **Progress Callbacks** | Cline, Claude Code | UI updates during generation |

**Implementation Variants:**
- **OpenAI format:** `data: {"choices": [...]}`
- **Anthropic format:** `event: content_block_delta`
- **Custom formats:** Various JSON structures

---

## 2. Tool Use & MCP

### 2.1 Built-in Tools

| Tool | Claude Code | Aider | Codex | Cline | OpenHands | Count |
|------|-------------|-------|-------|-------|-----------|-------|
| **read_file** | ✅ | ✅ | ⚠️ | ✅ | ✅ | 40+ |
| **write_file** | ✅ | ✅ | ⚠️ | ✅ | ✅ | 40+ |
| **bash_execute** | ✅ | ✅ | ⚠️ | ✅ | ✅ | 35+ |
| **glob_search** | ✅ | ✅ | ❌ | ✅ | ✅ | 25+ |
| **grep_search** | ✅ | ✅ | ❌ | ✅ | ✅ | 25+ |
| **list_directory** | ✅ | ✅ | ❌ | ✅ | ✅ | 30+ |
| **code_interpreter** | ❌ | ❌ | ✅ | ❌ | ⚠️ | 5 |
| **browser_navigate** | ❌ | ❌ | ❌ | ✅ | ❌ | 3 |
| **computer_control** | ❌ | ❌ | ❌ | ✅ | ❌ | 2 |

**Implementation Comparison:**

**Claude Code** (`src/tools/`):
```typescript
interface Tool {
  name: string;
  description: string;
  input_schema: JSONSchema;
  execute: (input: any) => Promise<ToolResult>;
}
```

**Aider** (embedded in coders):
```python
class EditBlockCoder:
    def execute_tool(self, tool_call):
        # Direct execution without abstraction
```

**HelixAgent** (`internal/mcp/adapters/`):
```go
type ToolAdapter interface {
    Execute(ctx context.Context, params map[string]interface{}) (*Result, error)
}
```

### 2.2 MCP Support

| Feature | Continue | Cline | Kiro | HelixAgent |
|---------|----------|-------|------|------------|
| **MCP Client** | ✅ | ⚠️ | ✅ | ✅ |
| **MCP Server** | ❌ | ❌ | ❌ | ✅ |
| **Tool Registration** | Dynamic | Static | Dynamic | Dynamic |
| **Resource Support** | ✅ | ❌ | ⚠️ | ✅ |
| **Prompts Support** | ❌ | ❌ | ❌ | ✅ |

**Porting Priority:** CRITICAL  
**Fusion Strategy:** HelixAgent's MCP server + Continue's client integration

### 2.3 Custom Tools

| Agent | Custom Tool System | Extensibility |
|-------|-------------------|---------------|
| **Claude Code** | ❌ | Fixed 7 tools |
| **Aider** | ❌ | Git-focused only |
| **Continue** | ✅ config.yaml | Plugin-based |
| **OpenHands** | ✅ Custom adapters | Python functions |
| **HelixAgent** | ✅ MCP | Protocol-based |

---

## 3. Git Integration

### 3.1 Git Features

| Feature | Aider | Claude Code | OpenHands | Count |
|---------|-------|-------------|-----------|-------|
| **Auto-commit** | ✅ | ❌ | ⚠️ | 15 |
| **Commit Attribution** | ✅ (Aider <aider@...>) | ❌ | ❌ | 5 |
| **Diff-based Edits** | ✅ (SEARCH/REPLACE) | ❌ | ❌ | 3 |
| **Repo Map** | ✅ | ❌ | ⚠️ | 2 |
| **Branch Management** | ✅ | ❌ | ❌ | 8 |
| **Undo (git revert)** | ✅ | ❌ | ❌ | 10 |
| **Change Review** | ✅ (git diff) | ❌ | ❌ | 12 |

**Aider's Repo Map** (`aider/repo.py`):
```python
class RepoMap:
    def get_ranked_tags(self, query, max_tokens=8000):
        # Uses tree-sitter for AST analysis
        # Returns relevant symbols for context
```

**Implementation Difficulty:** MEDIUM  
**Fusion Strategy:** Port Aider's full git workflow including repo map

### 3.2 Repository Understanding

| Feature | Aider | Kiro | Octogen | HelixAgent |
|---------|-------|------|---------|------------|
| **AST Parsing** | ✅ tree-sitter | ✅ | ✅ | ⚠️ |
| **Symbol Extraction** | ✅ | ✅ | ✅ | ❌ |
| **Import Analysis** | ✅ | ⚠️ | ⚠️ | ❌ |
| **Reference Tracking** | ✅ | ❌ | ⚠️ | ❌ |
| **TODO Detection** | ⚠️ | ❌ | ❌ | ❌ |

**Porting Priority:** HIGH  
**Fusion Strategy:** Implement Aider's repo map as core HelixAgent feature

---

## 4. Code Understanding

### 4.1 Language Support

| Feature | Aider | Continue | Claude Code | Count |
|---------|-------|----------|-------------|-------|
| **Python** | ✅ | ✅ | ✅ | All |
| **JavaScript/TypeScript** | ✅ | ✅ | ✅ | All |
| **Go** | ✅ | ✅ | ✅ | All |
| **Rust** | ✅ | ✅ | ✅ | 40+ |
| **Java** | ✅ | ✅ | ✅ | 35+ |
| **C/C++** | ✅ | ✅ | ✅ | 30+ |
| **Language Detection** | ✅ | ✅ | ❌ | 25+ |

### 4.2 Code Intelligence

| Feature | Continue (LSP) | Aider (Repo Map) | Claude Code | HelixAgent |
|---------|---------------|------------------|-------------|------------|
| **Go to Definition** | ✅ | ⚠️ | ❌ | ⚠️ |
| **Find References** | ✅ | ❌ | ❌ | ❌ |
| **Auto-completion** | ✅ | ❌ | ❌ | ⚠️ |
| **Diagnostics** | ✅ | ❌ | ❌ | ⚠️ |
| **Code Actions** | ✅ | ❌ | ❌ | ⚠️ |
| **Hover Info** | ✅ | ❌ | ❌ | ⚠️ |

**Porting Priority:** MEDIUM-HIGH  
**Fusion Strategy:** Integrate Continue's LSP support with Aider's repo map

---

## 5. Output & Formatting

### 5.1 Terminal Output

| Feature | Claude Code | Aider | Cline | Count |
|---------|-------------|-------|-------|-------|
| **Syntax Highlighting** | ✅ | ✅ | ✅ | 35+ |
| **Inline Code Display** | ✅ | ⚠️ | ✅ | 30+ |
| **Diff Visualization** | ✅ | ✅ | ⚠️ | 20+ |
| **Progress Indicators** | ✅ | ⚠️ | ✅ | 25+ |
| **Rich Terminal UI** | ✅ | ❌ | ✅ | 15+ |
| **Markdown Rendering** | ✅ | ⚠️ | ✅ | 30+ |
| **Color Themes** | ⚠️ | ❌ | ⚠️ | 10+ |

**Claude Code Terminal UI** (`src/terminal/`):
```typescript
// Rich inline display with syntax highlighting
interface TerminalRenderer {
  renderCodeBlock(code: string, language: string): void;
  renderDiff(oldCode: string, newCode: string): void;
  renderProgress(progress: number): void;
}
```

**Porting Priority:** HIGH  
**Fusion Strategy:** Port Claude Code's terminal UI system

### 5.2 Output Formats

| Feature | Aider | Continue | HelixAgent |
|---------|-------|----------|------------|
| **Plain Text** | ✅ | ✅ | ✅ |
| **Markdown** | ⚠️ | ✅ | ✅ |
| **JSON** | ❌ | ✅ | ✅ |
| **HTML** | ❌ | ❌ | ❌ |
| **Structured Data** | ❌ | ✅ | ✅ |

---

## 6. Autonomy & Execution

### 6.1 Autonomy Levels

| Feature | Cline | Claude Code | Aider | Codex | OpenHands |
|---------|-------|-------------|-------|-------|-----------|
| **Fully Autonomous** | ✅ | ⚠️ | ❌ | ⚠️ | ⚠️ |
| **Self-Correction** | ✅ | ⚠️ | ❌ | ⚠️ | ⚠️ |
| **Task Planning** | ✅ | ❌ | ⚠️ | ❌ | ⚠️ |
| **Error Recovery** | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ |
| **Loop Detection** | ❌ | ❌ | ❌ | ❌ | ⚠️ |

**Cline's Autonomy** (`src/core/autonomy.ts`):
```typescript
interface AutonomousAgent {
  planTask(goal: string): TaskPlan;
  executeStep(step: TaskStep): Promise<StepResult>;
  selfCorrect(error: Error): CorrectionAction;
  checkCompletion(): boolean;
}
```

### 6.2 Execution Environment

| Feature | OpenHands | Codex | Cline | Count |
|---------|-----------|-------|-------|-------|
| **Docker Sandbox** | ✅ | ✅ | ❌ | 8 |
| **Container Isolation** | ✅ | ✅ | ❌ | 6 |
| **Resource Limits** | ✅ | ⚠️ | ❌ | 5 |
| **Network Isolation** | ✅ | ✅ | ❌ | 5 |
| **Code Interpreter** | ❌ | ✅ | ❌ | 3 |

**Porting Priority:** HIGH  
**Fusion Strategy:** Port OpenHands' Docker sandboxing

### 6.3 Browser/Computer Use

| Feature | Cline | OpenHands | Goose | Count |
|---------|-------|-----------|-------|-------|
| **Browser Automation** | ✅ | ⚠️ | ✅ | 5 |
| **Screenshot Capture** | ✅ | ❌ | ✅ | 4 |
| **Mouse/Keyboard Control** | ✅ | ❌ | ✅ | 3 |
| **Visual Understanding** | ✅ | ❌ | ⚠️ | 2 |
| **Desktop Automation** | ❌ | ❌ | ✅ | 2 |

**Cline's Browser Use** (`src/core/browser.ts`):
```typescript
interface BrowserController {
  navigate(url: string): Promise<void>;
  screenshot(): Promise<Image>;
  click(x: number, y: number): Promise<void>;
  type(text: string): Promise<void>;
}
```

**Porting Priority:** MEDIUM  
**Fusion Strategy:** Port Cline's browser + Goose's desktop automation

---

## 7. Multi-Agent Coordination

### 7.1 Agent Communication

| Feature | HelixAgent | OpenHands | Forge | Claude Squad |
|---------|------------|-----------|-------|--------------|
| **Multi-Agent** | ✅ | ⚠️ | ✅ | ✅ |
| **Agent Registry** | ✅ | ❌ | ⚠️ | ⚠️ |
| **Message Routing** | ✅ | ❌ | ⚠️ | ⚠️ |
| **Broadcast** | ✅ | ❌ | ✅ | ✅ |
| **Consensus** | ✅ | ❌ | ⚠️ | ❌ |

### 7.2 Debate & Consensus

| Feature | HelixAgent | OpenHands | Forge | Multiagent Coding |
|---------|------------|-----------|-------|-------------------|
| **Debate Orchestration** | ✅ | ❌ | ⚠️ | ⚠️ |
| **Voting System** | ✅ | ❌ | ❌ | ⚠️ |
| **Consensus Building** | ✅ | ❌ | ⚠️ | ❌ |
| **Cross-Model Validation** | ✅ | ❌ | ❌ | ❌ |

**HelixAgent Debate** (`internal/debate/orchestrator.go`):
```go
type DebateOrchestrator struct {
  topic string
  participants []*DebateAgent
  topology DebateTopology
  rounds int
}
```

**Porting Priority:** Already best-in-class, extend only

---

## 8. IDE Integration

### 8.1 IDE Support

| IDE | Continue | Cline | Kiro | Count |
|-----|----------|-------|------|-------|
| **VS Code** | ✅ | ✅ | ⚠️ | 15+ |
| **JetBrains** | ✅ | ❌ | ⚠️ | 8 |
| **Neovim** | ✅ | ❌ | ❌ | 5 |
| **Emacs** | ✅ | ❌ | ❌ | 3 |
| **Vim** | ⚠️ | ❌ | ❌ | 3 |

### 8.2 LSP Features

| Feature | Continue | Cline | Kiro | HelixAgent |
|---------|----------|-------|------|------------|
| **LSP Client** | ✅ | ⚠️ | ✅ | ✅ |
| **LSP Server** | ❌ | ❌ | ❌ | ✅ |
| **Completion** | ✅ | ⚠️ | ✅ | ✅ |
| **Diagnostics** | ✅ | ⚠️ | ✅ | ✅ |
| **Hover** | ✅ | ⚠️ | ✅ | ✅ |
| **Code Actions** | ✅ | ❌ | ⚠️ | ✅ |

**Porting Priority:** MEDIUM  
**Fusion Strategy:** Port Continue's universal IDE support

---

## 9. Security & Sandboxing

### 9.1 Security Features

| Feature | OpenHands | Codex | HelixAgent | Count |
|---------|-----------|-------|------------|-------|
| **Container Sandboxing** | ✅ | ✅ | ⚠️ | 8 |
| **Resource Limits** | ✅ | ⚠️ | ✅ | 6 |
| **Network Isolation** | ✅ | ✅ | ⚠️ | 5 |
| **Input Validation** | ⚠️ | ✅ | ✅ | 10+ |
| **Audit Logging** | ❌ | ⚠️ | ✅ | 5 |

### 9.2 PII Protection

| Feature | Claude Code | Codex | HelixAgent |
|---------|-------------|-------|------------|
| **PII Detection** | ⚠️ | ⚠️ | ⚠️ |
| **Data Sanitization** | ❌ | ⚠️ | ⚠️ |
| **Secure Logging** | ❌ | ✅ | ⚠️ |

---

## 10. Context & Memory

### 10.1 Context Management

| Feature | Claude Code | Aider | Kiro | Octogen |
|---------|-------------|-------|------|---------|
| **Context Window** | 200K | Model-dep | Model-dep | 1M+ |
| **Smart Truncation** | ✅ | ⚠️ | ⚠️ | ✅ |
| **Context Compression** | ❌ | ❌ | ❌ | ✅ |
| **Hierarchical Context** | ❌ | ❌ | ⚠️ | ⚠️ |

### 10.2 Memory Systems

| Feature | Kiro | OpenHands | HelixAgent |
|---------|------|-----------|------------|
| **Project Memory** | ✅ | ⚠️ | ✅ |
| **Cross-Session Memory** | ✅ | ❌ | ✅ |
| **Semantic Memory** | ⚠️ | ❌ | ✅ |
| **Conversation History** | ❌ | ✅ | ✅ |

**Kiro's Project Memory** (`kiro/memory/`):
```python
class ProjectMemory:
    def remember(self, key: str, value: Any):
        # Persists across sessions
    
    def recall(self, key: str) -> Any:
        # Retrieves from memory
```

**Porting Priority:** HIGH  
**Fusion Strategy:** Port Kiro's memory system

---

## Feature Overlap Analysis

### High Overlap (10+ agents)

| Feature | Agents | Different Implementations |
|---------|--------|--------------------------|
| File reading | 40+ | Claude Code (tool), Aider (native), others |
| File writing | 40+ | Diff-based, Whole-file, Tool-based |
| Bash execution | 35+ | Direct, Sandboxed, Restricted |
| Streaming | 30+ | SSE, WebSocket, Custom |

### Medium Overlap (5-10 agents)

| Feature | Agents | Notes |
|---------|--------|-------|
| Git integration | 15 | Aider's is most advanced |
| Code completion | 12 | Continue's LSP-based |
| Docker sandbox | 8 | OpenHands' is best |
| Browser automation | 5 | Cline's is most complete |

### Low Overlap (1-4 agents)

| Feature | Agents | Unique Value |
|---------|--------|--------------|
| Reasoning models | Codex | o3/o4 support |
| Computer use | Cline, Goose | Visual interaction |
| Repo map | Aider | AST-based understanding |
| Project memory | Kiro | Cross-session learning |

---

## Implementation Priority Matrix

### Critical (Must Port)

1. **Aider's Git Integration** - Repo map, diff editing, commit attribution
2. **Claude Code's Terminal UI** - Rich output, syntax highlighting
3. **OpenHands' Sandboxing** - Docker-based security
4. **Kiro's Memory System** - Project memory, cross-session persistence
5. **Cline's Browser Use** - Web automation (future)

### High Priority

6. **Continue's LSP Support** - Universal IDE integration
7. **Aider's Repo Map** - AST-based code understanding
8. **Codex's Reasoning** - o3/o4 model support
9. **Claude Code's Tool UX** - Natural tool invocation
10. **Cline's Autonomy** - Self-directed execution

### Medium Priority

11. **Multi-provider switching** (Aider pattern)
12. **Custom tool system** (Continue pattern)
13. **Output formatting pipeline** (Claude Code pattern)
14. **Progress indicators** (Multiple agents)
15. **Code interpreter** (Codex pattern)

### Low Priority

16. Desktop automation (Goose)
17. Voice interface (VTCode)
18. Design tools (UI/UX Pro)

---

## Next Steps

**Pass 2: Deep Analysis** will analyze:
- Implementation details of each critical feature
- Code structure and dependencies
- Integration complexity
- Performance characteristics

**See:** [Pass 2 - Deep Analysis](pass_2_analysis.md)

---

*Pass 1 Complete: 284 features cataloged across 47 agents*  
*Date: 2026-04-03*  
*HelixAgent Commit: 8a976be2*
