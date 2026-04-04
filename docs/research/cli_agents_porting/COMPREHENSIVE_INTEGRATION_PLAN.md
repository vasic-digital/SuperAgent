# Comprehensive CLI Agents Feature Integration Plan

**Document:** Master Integration Roadmap  
**Scope:** 47 CLI Agents → HelixAgent  
**Date:** 2026-04-04  
**Status:** Planning Phase  
**Estimated Duration:** 52 weeks (12 months)

---

## Executive Summary

### Integration Strategy

This plan integrates the best features from 47 CLI agents into HelixAgent through:
1. **Direct Porting** - Core algorithms and patterns
2. **MCP Wrapping** - Full agent integration via MCP
3. **Native Implementation** - HelixAgent-optimized versions
4. **API Bridging** - Provider and model expansion

### Phase Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        52-WEEK INTEGRATION TIMELINE                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  Foundation      Core Features     Advanced      Ecosystem      Polish          │
│  (Weeks 1-8)     (Weeks 9-20)      (Weeks 21-32) (Weeks 33-44) (Weeks 45-52)   │
│     ┌────────┐      ┌────────┐       ┌────────┐    ┌────────┐    ┌────────┐    │
│     │ PHASE  │      │ PHASE  │       │ PHASE  │    │ PHASE  │    │ PHASE  │    │
│     │   1    │ ───▶ │   2    │ ────▶ │   3    │ ──▶│   4    │───▶│   5    │    │
│     └────────┘      └────────┘       └────────┘    └────────┘    └────────┘    │
│                                                                                 │
│  • Git-Native    • Repo Mapping   • Browser Use  • IDE Ext      • Performance   │
│  • Aider Fusion  • Terminal UI    • Computer     • VS Code      • Hardening     │
│  • Diff Editing  • Claude Tools   • Sandboxing   • JetBrains    • Security      │
│  • Core Types    • Codex Intprtr  • Reasoning    • LSP Bridge   • Docs          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## PHASE 1: Foundation Layer (Weeks 1-8)

### Week 1-2: Aider Integration - Git-Native Workflows

**Source:** Aider (cli_agents/aider)  
**Effort:** 80 hours  
**Owner:** Backend Team

#### Features to Port

| Feature | Source File | Target Location | Complexity |
|---------|-------------|-----------------|------------|
| **Repo Map** | `aider/repo.py` | `internal/clis/aider/repo_map.go` | High |
| **Git Operations** | `aider/repo.py` | `internal/clis/aider/git_ops.go` | Medium |
| **Diff Format** | `aider/coders/editblock_coder.py` | `internal/clis/aider/diff_format.go` | High |
| **Commit Attribution** | `aider/repo.py` | `internal/clis/aider/commit.go` | Low |

#### Implementation Details

```go
// internal/clis/aider/repo_map.go
package aider

type RepoMap struct {
    rootDir    string
    mapTokens  int
    tsParser   *treesitter.Parser
    ctags      *ctags.Universal
}

func (rm *RepoMap) GetRankedTags(ctx context.Context, query string, mentionedFiles []string) (*RepoContext, error) {
    // 1. Find matching files (fuzzy + semantic)
    files, err := rm.findMatchingFiles(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // 2. Extract symbols using tree-sitter
    symbols := make([]*Symbol, 0)
    for _, file := range files {
        fileSymbols, err := rm.extractSymbols(ctx, file)
        if err != nil {
            continue
        }
        symbols = append(symbols, fileSymbols...)
    }
    
    // 3. Rank symbols by relevance
    ranked := rm.rankSymbols(symbols, query, mentionedFiles)
    
    // 4. Format for LLM within token budget
    return rm.formatForLLM(ranked, rm.mapTokens), nil
}
```

#### API Additions

```go
// New endpoints in internal/handlers/
POST /v1/aider/repo-map           // Generate repo context
POST /v1/aider/diff-apply         // Apply SEARCH/REPLACE blocks
POST /v1/aider/git-commit         // Commit with attribution
GET  /v1/aider/repo-status        // Repository status
```

#### Testing

```go
func TestRepoMap_GetRankedTags(t *testing.T) {
    rm := aider.NewRepoMap("testdata/repo", 1024)
    
    ctx := context.Background()
    result, err := rm.GetRankedTags(ctx, "user authentication", nil)
    
    require.NoError(t, err)
    assert.NotEmpty(t, result.Symbols)
    
    // Verify ranking - auth files should be first
    for i := 0; i < min(5, len(result.Symbols)); i++ {
        assert.True(t, strings.Contains(result.Symbols[i].Name, "auth"))
    }
}
```

---

### Week 3-4: Core Types & Instance Management

**Source:** HelixAgent (new) + OpenHands patterns  
**Effort:** 100 hours  
**Owner:** Architecture Team

#### Components

| Component | Purpose | Location |
|-----------|---------|----------|
| **Instance Manager** | Lifecycle management | `internal/clis/instance_manager.go` |
| **Type Registry** | CLI agent type system | `internal/clis/types.go` |
| **Event Bus** | Cross-instance communication | `internal/clis/event_bus.go` |
| **Synchronization** | Distributed coordination | `internal/clis/sync/` |

#### SQL Schema

```sql
-- sql/001_cli_agents_fusion.sql
CREATE TABLE cli_agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- aider, claude_code, codex, etc.
    status VARCHAR(20) NOT NULL DEFAULT 'idle',
    config JSONB NOT NULL,
    session_id UUID REFERENCES sessions(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_heartbeat TIMESTAMP
);

CREATE TABLE cli_agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    input JSONB NOT NULL,
    output JSONB,
    error TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_cli_agent_instances_session ON cli_agent_instances(session_id);
CREATE INDEX idx_cli_agent_tasks_instance ON cli_agent_tasks(instance_id);
CREATE INDEX idx_cli_agent_tasks_status ON cli_agent_tasks(status);
```

---

### Week 5-6: Terminal UI Framework

**Source:** Claude Code terminal patterns  
**Effort:** 80 hours  
**Owner:** Frontend Team

#### Features

| Feature | Source | Target |
|---------|--------|--------|
| **Rich Terminal** | Claude Code | `internal/output/terminal/enhanced.go` |
| **Code Blocks** | Claude Code | `internal/output/formatters/syntax.go` |
| **Progress Bars** | Claude Code | `internal/output/terminal/progress.go` |
| **Diff Rendering** | Aider | `internal/output/formatters/diff.go` |
| **Tool Call UI** | Claude Code | `internal/output/terminal/tool_use.go` |

#### Implementation

```go
// internal/output/terminal/enhanced.go
package terminal

type EnhancedUI struct {
    style     *chroma.Style
    formatter *chroma.Formatter
    width     int
}

func (ui *EnhancedUI) RenderCodeBlock(code, language string, opts RenderOptions) string {
    lexer := lexers.Get(language)
    iterator, _ := lexer.Tokenise(nil, code)
    
    var buf strings.Builder
    
    // Add line numbers if requested
    if opts.LineNumbers {
        ui.writeLineNumbers(&buf, code)
    }
    
    // Syntax highlight
    ui.formatter.Format(&buf, ui.style, iterator)
    
    // Add copy button indicator
    if opts.CopyButton {
        buf.WriteString("\n[Ctrl+C to copy]")
    }
    
    return buf.String()
}

func (ui *EnhancedUI) RenderToolCall(toolName string, args map[string]interface{}, status ToolStatus) string {
    var buf strings.Builder
    
    // Tool header with icon
    icon := ui.getToolIcon(toolName)
    buf.WriteString(fmt.Sprintf("%s %s\n", icon, toolName))
    
    // Arguments
    for key, value := range args {
        buf.WriteString(fmt.Sprintf("  • %s: %v\n", key, value))
    }
    
    // Status indicator
    switch status {
    case StatusRunning:
        buf.WriteString("  ⏳ Running...")
    case StatusComplete:
        buf.WriteString("  ✅ Complete")
    case StatusError:
        buf.WriteString("  ❌ Error")
    }
    
    return buf.String()
}
```

---

### Week 7-8: MCP Integration Layer

**Source:** Various agents via MCP  
**Effort:** 60 hours  
**Owner:** Protocol Team

#### MCP Servers to Create

| Agent | MCP Server | Capabilities |
|-------|------------|--------------|
| **Aider** | `mcp-server-aider` | Repo map, git ops, diff editing |
| **Claude Code** | `mcp-server-claude` | Terminal UI, tool use patterns |
| **Codex** | `mcp-server-codex` | Code interpreter, reasoning |
| **Cline** | `mcp-server-cline` | Browser automation (future) |
| **OpenHands** | `mcp-server-openhands` | Sandboxing (future) |

#### MCP Configuration

```json
{
  "mcpServers": {
    "aider": {
      "command": "helixagent-mcp-aider",
      "args": ["--stdio"],
      "env": {
        "AIDER_REPO_PATH": "${workspace}"
      }
    },
    "claude-code": {
      "command": "helixagent-mcp-claude",
      "args": ["--stdio"]
    }
  }
}
```

---

## PHASE 2: Core Features (Weeks 9-20)

### Week 9-11: Claude Code Integration - Tool Use & UX

**Source:** Claude Code (cli_agents/claude-code)  
**Effort:** 100 hours

#### Features to Port

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **Tool Framework** | `src/lib/tools/` | `internal/clis/claude_code/tool_use.go` | High |
| **Approval System** | `src/lib/approval/` | `internal/clis/claude_code/approval.go` | Medium |
| **Context Management** | `src/lib/context/` | `internal/clis/claude_code/context.go` | High |
| **Auto-Mode** | `src/lib/auto/` | `internal/clis/claude_code/autonomy.go` | Medium |

#### Tool Use Implementation

```go
// internal/clis/claude_code/tool_use.go
package claude_code

type ToolUseFramework struct {
    tools      map[string]Tool
    approvals  ApprovalManager
    executor   ToolExecutor
}

type Tool struct {
    Name        string
    Description string
    Parameters  jsonschema.Schema
    Handler     ToolHandler
    RequiresApproval bool
}

func (tuf *ToolUseFramework) ExecuteTool(ctx context.Context, toolCall ToolCall) (*ToolResult, error) {
    tool, exists := tuf.tools[toolCall.Name]
    if !exists {
        return nil, fmt.Errorf("unknown tool: %s", toolCall.Name)
    }
    
    // Validate parameters
    if err := tuf.validateParams(tool, toolCall.Arguments); err != nil {
        return nil, err
    }
    
    // Check approval
    if tool.RequiresApproval {
        approved, err := tuf.approvals.RequestApproval(ctx, toolCall)
        if err != nil || !approved {
            return nil, fmt.Errorf("tool execution not approved")
        }
    }
    
    // Execute
    result, err := tool.Handler(ctx, toolCall.Arguments)
    if err != nil {
        return nil, err
    }
    
    return &ToolResult{
        Tool:    toolCall.Name,
        Output:  result,
        Timing:  time.Since(start),
    }, nil
}
```

#### Tool Definitions

```go
var DefaultTools = map[string]Tool{
    "Bash": {
        Name:        "Bash",
        Description: "Execute bash commands",
        Parameters: jsonschema.Schema{
            Type: "object",
            Properties: map[string]jsonschema.Schema{
                "command": {Type: "string", Description: "Command to execute"},
                "description": {Type: "string", Description: "What this command does"},
            },
            Required: []string{"command", "description"},
        },
        Handler:      bashHandler,
        RequiresApproval: true,
    },
    "Read": {
        Name:        "Read",
        Description: "Read file contents",
        Parameters: jsonschema.Schema{
            Type: "object",
            Properties: map[string]jsonschema.Schema{
                "file_path": {Type: "string", Description: "Path to file"},
            },
            Required: []string{"file_path"},
        },
        Handler:      readHandler,
        RequiresApproval: false,
    },
    // ... more tools
}
```

---

### Week 12-14: Codex Integration - Reasoning & Interpreter

**Source:** OpenAI Codex (cli_agents/codex)  
**Effort:** 80 hours

#### Features

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **Code Interpreter** | Codex | `internal/clis/codex/interpreter.go` | High |
| **Reasoning Display** | Codex | `internal/clis/codex/reasoning.go` | Low |
| **o3/o4 Support** | Codex | `internal/llm/providers/openai/reasoning.go` | Medium |

#### Code Interpreter

```go
// internal/clis/codex/interpreter.go
package codex

type CodeInterpreter struct {
    sandbox    *sandbox.Sandbox
    langRuntimes map[string]Runtime
}

type ExecutionResult struct {
    Output     string
    Error      string
    ExitCode   int
    Artifacts  []Artifact
    Duration   time.Duration
}

func (ci *CodeInterpreter) Execute(ctx context.Context, code, language string) (*ExecutionResult, error) {
    runtime, exists := ci.langRuntimes[language]
    if !exists {
        return nil, fmt.Errorf("unsupported language: %s", language)
    }
    
    // Create isolated execution environment
    env, err := ci.sandbox.CreateEnvironment(ctx, language)
    if err != nil {
        return nil, err
    }
    defer env.Cleanup()
    
    // Execute code
    result, err := runtime.Execute(ctx, env, code)
    if err != nil {
        return &ExecutionResult{
            Error:    err.Error(),
            ExitCode: 1,
        }, nil
    }
    
    return result, nil
}
```

#### Reasoning Model Support

```go
// internal/llm/providers/openai/reasoning.go
package openai

type ReasoningConfig struct {
    Effort          string  // "low", "medium", "high"
    MaxTokens       int
    ShowReasoning   bool    // Display reasoning to user
}

func (p *Provider) CreateReasoningCompletion(ctx context.Context, req *models.LLMRequest, config ReasoningConfig) (*models.LLMResponse, error) {
    // Use o3 or o4 model
    model := "o3-mini"
    if config.Effort == "high" {
        model = "o4-mini"
    }
    
    body := map[string]interface{}{
        "model":       model,
        "messages":    convertMessages(req.Messages),
        "max_tokens":  config.MaxTokens,
        "reasoning": map[string]string{
            "effort": config.Effort,
        },
    }
    
    // Make request
    resp, err := p.client.Do(ctx, "POST", "/v1/chat/completions", body)
    if err != nil {
        return nil, err
    }
    
    // Extract reasoning if present
    var reasoning string
    if config.ShowReasoning && resp.Reasoning != nil {
        reasoning = resp.Reasoning.Content
    }
    
    return &models.LLMResponse{
        Content:   resp.Choices[0].Message.Content,
        Reasoning: reasoning,
        Usage:     convertUsage(resp.Usage),
    }, nil
}
```

---

### Week 15-17: Cline Integration - Browser Automation

**Source:** Cline (cli_agents/cline)  
**Effort:** 160 hours

#### Features

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **Browser Control** | Cline | `internal/browser/automation.go` | High |
| **Computer Use** | Cline | `internal/clis/cline/computer_use.go` | High |
| **Autonomy** | Cline | `internal/clis/cline/autonomy.go` | Medium |

#### Browser Automation

```go
// internal/browser/automation.go
package browser

type AutomationController struct {
    playwright *playwright.Playwright
    browser    playwright.Browser
    context    playwright.BrowserContext
    page       playwright.Page
}

func (ac *AutomationController) Navigate(ctx context.Context, url string) error {
    _, err := ac.page.Goto(url, playwright.PageGotoOptions{
        WaitUntil: playwright.WaitUntilStateNetworkidle,
    })
    return err
}

func (ac *AutomationController) Click(ctx context.Context, selector string) error {
    return ac.page.Click(selector)
}

func (ac *AutomationController) Type(ctx context.Context, selector, text string) error {
    return ac.page.Fill(selector, text)
}

func (ac *AutomationController) Screenshot(ctx context.Context) ([]byte, error) {
    return ac.page.Screenshot(playwright.PageScreenshotOptions{
        FullPage: playwright.Bool(true),
    })
}

func (ac *AutomationController) GetAccessibleTree(ctx context.Context) (*AccessibleTree, error) {
    // Extract accessibility tree for LLM consumption
    tree, err := ac.page.Accessibility.Snapshot()
    if err != nil {
        return nil, err
    }
    
    return convertToAccessibleTree(tree), nil
}
```

#### Computer Use API

```go
// internal/clis/cline/computer_use.go
package cline

// ComputerUseAction represents actions the LLM can take
type ComputerUseAction struct {
    Action string                 `json:"action"` // click, type, scroll, screenshot
    Params map[string]interface{} `json:"params"`
}

func (cu *ComputerUse) ExecuteAction(ctx context.Context, action ComputerUseAction) (*ActionResult, error) {
    switch action.Action {
    case "click":
        return cu.click(ctx, action.Params)
    case "type":
        return cu.typeText(ctx, action.Params)
    case "scroll":
        return cu.scroll(ctx, action.Params)
    case "screenshot":
        return cu.screenshot(ctx)
    default:
        return nil, fmt.Errorf("unknown action: %s", action.Action)
    }
}
```

---

### Week 18-20: OpenHands Integration - Sandboxing

**Source:** OpenHands (cli_agents/openhands)  
**Effort:** 120 hours

#### Features

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **Docker Sandbox** | OpenHands | `internal/sandbox/docker.go` | High |
| **Security Isolation** | OpenHands | `internal/sandbox/security.go` | High |
| **Jupyter Integration** | OpenHands | `internal/sandbox/jupyter.go` | Medium |

#### Sandboxing Implementation

```go
// internal/sandbox/docker.go
package sandbox

type DockerSandbox struct {
    client    *docker.Client
    container string
    config    SandboxConfig
}

type SandboxConfig struct {
    Image       string
    Memory      int64    // MB
    CPU         float64  // Cores
    Timeout     time.Duration
    Network     bool
    Volumes     map[string]string
}

func (ds *DockerSandbox) Execute(ctx context.Context, command string) (*ExecutionResult, error) {
    ctx, cancel := context.WithTimeout(ctx, ds.config.Timeout)
    defer cancel()
    
    exec, err := ds.client.ContainerExecCreate(ctx, ds.container, types.ExecConfig{
        Cmd:          []string{"sh", "-c", command},
        AttachStdout: true,
        AttachStderr: true,
    })
    if err != nil {
        return nil, err
    }
    
    resp, err := ds.client.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
    if err != nil {
        return nil, err
    }
    defer resp.Close()
    
    // Stream output
    var stdout, stderr bytes.Buffer
    _, err = stdcopy.StdCopy(&stdout, &stderr, resp.Reader)
    if err != nil {
        return nil, err
    }
    
    // Get exit code
    inspect, err := ds.client.ContainerExecInspect(ctx, exec.ID)
    if err != nil {
        return nil, err
    }
    
    return &ExecutionResult{
        Output:   stdout.String(),
        Error:    stderr.String(),
        ExitCode: inspect.ExitCode,
    }, nil
}
```

---

## PHASE 3: Advanced Features (Weeks 21-32)

### Week 21-24: Continue Integration - IDE Support

**Source:** Continue (cli_agents/continue)  
**Effort:** 140 hours

#### Features

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **LSP Client** | Continue | `internal/ide/lsp/client.go` | High |
| **VS Code Extension** | Continue | `ide/vscode/` | High |
| **JetBrains Plugin** | Continue | `ide/jetbrains/` | High |
| **Autocomplete** | Continue | `internal/ide/autocomplete.go` | Medium |

#### LSP Integration

```go
// internal/ide/lsp/client.go
package lsp

type Client struct {
    conn   jsonrpc2.Conn
    server ServerInfo
    caps   ClientCapabilities
}

type ServerInfo struct {
    Name    string
    Version string
    RootURI string
}

func (c *Client) Initialize(ctx context.Context, rootURI string) (*InitializeResult, error) {
    params := &InitializeParams{
        RootURI: rootURI,
        Capabilities: ClientCapabilities{
            TextDocument: TextDocumentClientCapabilities{
                Completion: &CompletionClientCapabilities{
                    DynamicRegistration: true,
                    CompletionItem: &CompletionItemCapabilities{
                        SnippetSupport: true,
                    },
                },
            },
        },
    }
    
    var result InitializeResult
    err := c.conn.Call(ctx, "initialize", params, &result)
    return &result, err
}

func (c *Client) GetCompletions(ctx context.Context, uri string, line, char int) ([]CompletionItem, error) {
    params := &CompletionParams{
        TextDocumentPositionParams: TextDocumentPositionParams{
            TextDocument: TextDocumentIdentifier{URI: uri},
            Position: Position{Line: line, Character: char},
        },
    }
    
    var result CompletionList
    err := c.conn.Call(ctx, "textDocument/completion", params, &result)
    return result.Items, err
}
```

---

### Week 25-28: Kiro Integration - Project Memory

**Source:** Kiro (cli_agents/kiro-cli)  
**Effort:** 100 hours

#### Features

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **Project Memory** | Kiro | `internal/memory/project.go` | High |
| **Spec-Driven Dev** | Kiro | `internal/spec/engine.go` | Medium |
| **Context Learning** | Kiro | `internal/memory/learning.go` | Medium |

#### Project Memory

```go
// internal/memory/project.go
package memory

type ProjectMemory struct {
    db         *sql.DB
    embeddings EmbeddingClient
    projectID  string
}

type MemoryEntry struct {
    ID          string
    Type        string // code, conversation, decision, error
    Content     string
    Embedding   []float32
    Metadata    map[string]interface{}
    Importance  float64
    CreatedAt   time.Time
    AccessedAt  time.Time
    AccessCount int
}

func (pm *ProjectMemory) Store(ctx context.Context, entry *MemoryEntry) error {
    // Generate embedding
    embedding, err := pm.embeddings.Embed(ctx, entry.Content)
    if err != nil {
        return err
    }
    entry.Embedding = embedding
    
    // Store in database
    _, err = pm.db.ExecContext(ctx, `
        INSERT INTO project_memory 
        (id, project_id, type, content, embedding, metadata, importance)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, entry.ID, pm.projectID, entry.Type, entry.Content, 
        pgvector.NewVector(embedding), entry.Metadata, entry.Importance)
    
    return err
}

func (pm *ProjectMemory) Retrieve(ctx context.Context, query string, limit int) ([]*MemoryEntry, error) {
    // Generate query embedding
    queryEmbedding, err := pm.embeddings.Embed(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Search by similarity
    rows, err := pm.db.QueryContext(ctx, `
        SELECT id, type, content, metadata, importance,
               embedding <=> $1 as distance
        FROM project_memory
        WHERE project_id = $2
        ORDER BY embedding <=> $1
        LIMIT $3
    `, pgvector.NewVector(queryEmbedding), pm.projectID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var entries []*MemoryEntry
    for rows.Next() {
        var entry MemoryEntry
        var distance float64
        err := rows.Scan(&entry.ID, &entry.Type, &entry.Content, 
                        &entry.Metadata, &entry.Importance, &distance)
        if err != nil {
            continue
        }
        entries = append(entries, &entry)
    }
    
    return entries, nil
}
```

---

### Week 29-32: Plandex Integration - Task Planning

**Source:** Plandex (cli_agents/plandex)  
**Effort:** 80 hours

#### Features

| Feature | Source | Target | Complexity |
|---------|--------|--------|------------|
| **Task Planning** | Plandex | `internal/planning/task_planner.go` | High |
| **Plan Execution** | Plandex | `internal/planning/executor.go` | Medium |
| **Progress Tracking** | Plandex | `internal/planning/progress.go` | Low |

#### Task Planner

```go
// internal/planning/task_planner.go
package planning

type TaskPlanner struct {
    llm      LLMClient
    executor TaskExecutor
}

type Plan struct {
    ID          string
    Objective   string
    Tasks       []Task
    Dependencies map[string][]string // task ID -> dependency IDs
    Status      PlanStatus
}

type Task struct {
    ID          string
    Description string
    Type        string // research, implement, test, review
    Status      TaskStatus
    EstDuration time.Duration
    Result      *TaskResult
}

func (tp *TaskPlanner) CreatePlan(ctx context.Context, objective string, context []string) (*Plan, error) {
    // Use LLM to break down objective into tasks
    prompt := fmt.Sprintf(`
        Create a detailed plan for: %s
        
        Context:
        %s
        
        Break this down into specific, actionable tasks.
        For each task, specify:
        - Description
        - Type (research, implement, test, review)
        - Dependencies on other tasks
        - Estimated duration
        
        Return as JSON.
    `, objective, strings.Join(context, "\n"))
    
    response, err := tp.llm.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse plan from JSON
    var plan Plan
    if err := json.Unmarshal([]byte(response), &plan); err != nil {
        return nil, err
    }
    
    plan.ID = uuid.New().String()
    plan.Status = PlanStatusPending
    
    return &plan, nil
}

func (tp *TaskPlanner) ExecutePlan(ctx context.Context, plan *Plan) error {
    plan.Status = PlanStatusRunning
    
    // Topological sort for dependency order
    sorted, err := tp.topologicalSort(plan)
    if err != nil {
        return err
    }
    
    // Execute tasks in order
    for _, taskID := range sorted {
        task := plan.GetTask(taskID)
        
        // Check dependencies
        if !tp.dependenciesMet(task, plan) {
            continue
        }
        
        // Execute task
        result, err := tp.executor.Execute(ctx, task)
        if err != nil {
            task.Status = TaskStatusFailed
            return err
        }
        
        task.Result = result
        task.Status = TaskStatusComplete
    }
    
    plan.Status = PlanStatusComplete
    return nil
}
```

---

## PHASE 4: Ecosystem Expansion (Weeks 33-44)

### Week 33-36: Provider Expansion

**Goal:** Add support for all providers used by CLI agents

| Provider | Status | Priority | CLI Agents Using |
|----------|--------|----------|------------------|
| **Amazon Bedrock** | ❌ Missing | High | Amazon Q |
| **Azure OpenAI** | ⚠️ Partial | High | TaskWeaver |
| **Cohere** | ❌ Missing | Medium | - |
| **AI21 Labs** | ❌ Missing | Low | - |
| **Aleph Alpha** | ❌ Missing | Low | - |

#### Amazon Bedrock Provider

```go
// internal/llm/providers/bedrock/bedrock.go
package bedrock

type Provider struct {
    client    *bedrockruntime.Client
    region    string
}

func (p *Provider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    input := &bedrockruntime.InvokeModelInput{
        ModelId:     aws.String(req.Model),
        ContentType: aws.String("application/json"),
        Body:        p.buildRequestBody(req),
    }
    
    output, err := p.client.InvokeModel(ctx, input)
    if err != nil {
        return nil, err
    }
    
    return p.parseResponse(output.Body)
}
```

---

### Week 37-40: Specialized Integrations

#### Goose - Desktop Automation

```go
// internal/desktop/automation.go
package desktop

type DesktopAutomation struct {
    screen    ScreenCapture
    input     InputController
    vision    VisionClient
}

func (da *DesktopAutomation) Click(x, y int) error
func (da *DesktopAutomation) Type(text string) error
func (da *DesktopAutomation) Screenshot() ([]byte, error)
func (da *DesktopAutomation) FindElement(description string) (*Element, error)
```

#### TaskWeaver - Microsoft Integration

```go
// internal/integrations/microsoft/client.go
package microsoft

type Client struct {
    graph *msgraphsdk.GraphServiceClient
}

func (c *Client) GetOutlookEmails(ctx context.Context, filter string) ([]Email, error)
func (c *Client) GetTeamsMessages(ctx context.Context, teamID string) ([]Message, error)
func (c *Client) CreateCalendarEvent(ctx context.Context, event *Event) error
```

---

### Week 41-44: VS Code Extension Development

**Extension Features:**
- Chat panel with HelixAgent integration
- Inline completions
- Diff view for edits
- Tool use visualization
- Ensemble voting display

```typescript
// ide/vscode/src/extension.ts
import * as vscode from 'vscode';
import { HelixAgentClient } from './client';

export function activate(context: vscode.ExtensionContext) {
    const client = new HelixAgentClient();
    
    // Register chat provider
    vscode.chat.registerChatParticipant('helixagent', {
        async request(request, context, response, token) {
            const result = await client.complete(request.prompt);
            response.markdown(result.content);
        }
    });
    
    // Register completion provider
    vscode.languages.registerInlineCompletionItemProvider(
        { pattern: '**/*' },
        new HelixCompletionProvider(client)
    );
}
```

---

## PHASE 5: Polish & Hardening (Weeks 45-52)

### Week 45-48: Performance Optimization

- **Caching Layer:** Semantic caching for all providers
- **Connection Pooling:** Reuse LLM connections
- **Async Processing:** Background task execution
- **Streaming Optimization:** Reduce latency

### Week 49-50: Security Hardening

- **Sandbox Security:** Audit sandbox escapes
- **Input Validation:** Sanitize all inputs
- **Secrets Management:** Secure API key storage
- **Audit Logging:** Complete audit trail

### Week 51-52: Documentation & Release

- **API Documentation:** Complete OpenAPI specs
- **Integration Guides:** Step-by-step tutorials
- **Architecture Docs:** System design documentation
- **Release Notes:** Feature summary

---

## Resource Requirements

### Team Composition

| Role | Count | Duration |
|------|-------|----------|
| Backend Engineers | 3 | Full 12 months |
| Frontend/IDE Engineers | 2 | Months 3-12 |
| DevOps/SRE | 1 | Months 6-12 |
| Technical Writer | 1 | Months 9-12 |
| QA Engineer | 1 | Months 6-12 |

### Infrastructure

| Resource | Purpose | Cost/Month |
|----------|---------|------------|
| GPU Servers | Model hosting | $2,000 |
| Test Environments | CI/CD | $500 |
| Sandbox Infrastructure | Secure execution | $1,000 |
| **Total** | | **$3,500** |

---

## Success Metrics

### Feature Completeness

| Phase | Target | Measurement |
|-------|--------|-------------|
| Phase 1 | 100% | All foundation features working |
| Phase 2 | 90% | Core CLI agent features ported |
| Phase 3 | 80% | Advanced features operational |
| Phase 4 | 70% | Ecosystem integrations complete |
| Phase 5 | 100% | Production ready |

### Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| Latency (p95) | < 2s | - |
| Throughput | 1000 req/s | - |
| Availability | 99.9% | - |
| Error Rate | < 0.1% | - |

---

## Conclusion

This 52-week plan integrates the best features from 47 CLI agents into HelixAgent, transforming it into a universal AI orchestration platform. By the end of this plan, HelixAgent will:

1. **Support all major CLI agent workflows** via MCP
2. **Provide IDE-native experience** through VS Code/JetBrains extensions
3. **Offer enterprise-grade sandboxing** via OpenHands integration
4. **Include git-native workflows** via Aider integration
5. **Support browser/computer use** via Cline integration
6. **Provide comprehensive project memory** via Kiro integration
7. **Enable task planning** via Plandex integration

**HelixAgent will become the single platform that unifies the fragmented CLI agent ecosystem.**
