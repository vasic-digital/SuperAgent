# CLI Agents Porting - Implementation Phases

**Document:** Implementation Roadmap  
**Last Updated:** 2026-04-03  
**Status:** Planning Complete  

---

## Quick Start

This document provides the actionable roadmap for porting 47 CLI agent features into HelixAgent. For detailed analysis, see the 5-pass planning documents.

### Key Links
- [Pass 1: Analysis](passes/pass1_analysis.md) - Feature taxonomy and gap analysis
- [Pass 2: Architecture](passes/pass2_architecture.md) - Component organization
- [Pass 3: Integration Strategy](passes/pass3_integration_strategy.md) - Port vs adapt vs implement
- [Pass 4: Optimization](passes/pass4_optimization.md) - Performance and resource planning
- [Pass 5: Final Implementation Plan](passes/pass5_finalization.md) - 24-week roadmap

---

## Implementation Summary

### 5-Phase Approach (24 Weeks)

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              CLI AGENTS PORTING TIMELINE                                │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│  Weeks 1-4          Weeks 5-8          Weeks 9-16         Weeks 17-20      Weeks 21-24 │
│     ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌────────┐ │
│     │  PHASE 1    │   │  PHASE 2    │   │  PHASE 3    │   │  PHASE 4    │   │ PHASE 5│ │
│     │ Foundation  │──▶│ Ensemble    │──▶│ CLI Integr. │──▶│ Output      │──▶│ Testing│ │
│     │ Layer       │   │ Extension   │   │ (Aider,     │   │ System      │   │ &      │ │
│     │             │   │             │   │ Claude, etc)│   │             │   │ Deploy │ │
│     └─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘   └────────┘ │
│          │                  │                 │                 │                │      │
│          ▼                  ▼                 ▼                 ▼                ▼      │
│     • SQL Schema      • Multi-Inst.    • Repo Map         • Formatting     • Tests    │
│     • Core Types      • Coordinator    • Terminal UI      • Pipeline       • Docs     │
│     • Instance Mgr    • Worker Pool    • Sandbox           • Terminal       • Deploy   │
│     • Synchronization • Load Balancer  • Browser           • Streaming                   │
│     • Event Bus       • Health Monitor • Memory            • Templates                   │
│     • Registry        • Auto-scaling   • LSP Client        • Custom                      │
│                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Critical Features (20 Total)

### Priority 1: Essential (Weeks 1-8)

| # | Feature | Source | Target File | Complexity | Effort |
|---|---------|--------|-------------|------------|--------|
| 1 | **Instance Management** | HelixAgent | `internal/clis/instance_manager.go` | High | 120h |
| 2 | **Ensemble Coordination** | HelixAgent | `internal/ensemble/multi_instance/` | High | 140h |
| 3 | **SQL Schema** | HelixAgent | `sql/001_cli_agents_fusion.sql` | Medium | 40h |
| 4 | **Synchronization** | HelixAgent | `internal/ensemble/synchronization/` | High | 100h |
| 5 | **Event Bus** | HelixAgent | `internal/ensemble/event_bus.go` | Medium | 60h |

### Priority 2: Code Understanding (Weeks 9-12)

| # | Feature | Source | Target File | Complexity | Effort |
|---|---------|--------|-------------|------------|--------|
| 6 | **Repo Map** | Aider | `internal/clis/aider/repo_map.go` | High | 120h |
| 7 | **Diff Format** | Aider | `internal/clis/aider/diff_format.go` | Medium | 40h |
| 8 | **Git Integration** | Aider | `internal/clis/aider/git_ops.go` | High | 80h |
| 9 | **Terminal UI** | Claude Code | `internal/clis/claude_code/terminal_ui.go` | High | 100h |
| 10 | **Tool Use** | Claude Code | `internal/clis/claude_code/tool_use.go` | Medium | 60h |

### Priority 3: Execution & Security (Weeks 13-14)

| # | Feature | Source | Target File | Complexity | Effort |
|---|---------|--------|-------------|------------|--------|
| 11 | **Code Interpreter** | Codex | `internal/clis/codex/interpreter.go` | High | 80h |
| 12 | **Reasoning Models** | Codex | `internal/clis/codex/reasoning.go` | Low | 20h |
| 13 | **Browser Automation** | Cline | `internal/clis/cline/browser.go` | High | 160h |
| 14 | **Computer Use** | Cline | `internal/clis/cline/computer_use.go` | High | 200h |
| 15 | **Autonomy** | Cline | `internal/clis/cline/autonomy.go` | High | 120h |

### Priority 4: Memory & IDE (Weeks 15-16)

| # | Feature | Source | Target File | Complexity | Effort |
|---|---------|--------|-------------|------------|--------|
| 16 | **Sandbox** | OpenHands | `internal/clis/openhands/sandbox.go` | High | 100h |
| 17 | **Security Isolation** | OpenHands | `internal/clis/openhands/security.go` | Medium | 60h |
| 18 | **Project Memory** | Kiro | `internal/clis/kiro/memory.go` | Medium | 80h |
| 19 | **LSP Client** | Continue | `internal/clis/continue/lsp.go` | High | 100h |
| 20 | **Streaming Pipeline** | HelixAgent | `internal/output/pipeline.go` | Medium | 80h |

---

## Architecture Overview

### Directory Structure

```
internal/
├── clis/                          # CLI agent integrations
│   ├── types.go                   # Shared types
│   ├── instance_manager.go        # Instance lifecycle
│   ├── aider/                     # Aider components
│   │   ├── repo_map.go            # Repository understanding
│   │   ├── diff_format.go         # SEARCH/REPLACE editing
│   │   └── git_ops.go             # Git operations
│   ├── claude_code/               # Claude Code components
│   │   ├── terminal_ui.go         # Rich terminal UI
│   │   └── tool_use.go            # Tool use framework
│   ├── codex/                     # Codex components
│   │   ├── interpreter.go         # Code execution
│   │   └── reasoning.go           # Reasoning display
│   ├── cline/                     # Cline components
│   │   ├── browser.go             # Browser automation
│   │   ├── computer_use.go        # Computer control
│   │   └── autonomy.go            # Autonomous execution
│   ├── openhands/                 # OpenHands components
│   │   ├── sandbox.go             # Docker sandboxing
│   │   └── security.go            # Security isolation
│   ├── kiro/                      # Kiro components
│   │   └── memory.go              # Project memory
│   └── continue/                  # Continue components
│       └── lsp.go                 # LSP client
├── ensemble/                      # Ensemble system
│   ├── multi_instance/            # Multi-instance coordination
│   │   ├── coordinator.go         # Session coordinator
│   │   ├── pool.go                # Instance pooling
│   │   └── load_balancer.go       # Load distribution
│   ├── background/                # Background workers
│   │   └── worker_pool.go         # Task processing
│   ├── synchronization/           # Distributed sync
│   │   ├── manager.go             # Lock manager
│   │   └── crdt.go                # CRDT primitives
│   └── event_bus.go               # Event bus
└── output/                        # Output system
    ├── pipeline.go                # Processing pipeline
    ├── formatters/                # Content formatters
    │   ├── syntax.go              # Syntax highlighting
    │   ├── diff.go                # Diff formatting
    │   └── table.go               # Table formatting
    ├── renderers/                 # Output renderers
    │   ├── terminal.go            # Terminal output
    │   ├── html.go                # HTML output
    │   └── json.go                # JSON output
    └── terminal/                  # Terminal UI
        └── enhanced.go            # Rich UI components
```

### Component Organization

```
┌─────────────────────────────────────────────────────────────────┐
│                      HELIXAGENT CORE                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Router    │  │  Ensemble   │  │   Instance Manager      │ │
│  │             │  │ Coordinator │  │                         │ │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────────┘ │
│         │                │                    │                 │
│         └────────────────┴────────────────────┘                 │
│                          │                                      │
│         ┌────────────────┼────────────────┐                    │
│         ▼                ▼                ▼                    │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐               │
│  │  Aider     │  │Claude Code │  │  Cline     │               │
│  │ Components │  │ Components │  │ Components │  ...           │
│  └────────────┘  └────────────┘  └────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Code Migration Patterns

### Pattern 1: Direct Port (Aider's Repo Map)

```python
# Original: Aider repo_map.py (simplified)
class RepoMap:
    def get_ranked_tags(self, query, mentioned_files):
        files = self.find_matching_files(query)
        symbols = []
        for file in files:
            symbols.extend(self.extract_symbols(file))
        ranked = self.rank_symbols(symbols, query, mentioned_files)
        return self.format_for_llm(ranked)
```

```go
// Migrated: HelixAgent internal/clis/aider/repo_map.go
func (rm *RepoMap) GetRankedTags(ctx context.Context, query string, mentionedFiles []string) (*RepoContext, error) {
    files, err := rm.findMatchingFiles(ctx, query)
    if err != nil {
        return nil, err
    }
    
    symbols := make([]*Symbol, 0)
    for _, file := range files {
        fileSymbols, err := rm.extractSymbols(ctx, file)
        if err != nil {
            continue
        }
        symbols = append(symbols, fileSymbols...)
    }
    
    ranked := rm.rankSymbols(symbols, query, mentionedFiles)
    return rm.formatForLLM(ranked, rm.mapTokens), nil
}
```

### Pattern 2: API Adapter (Claude Code Terminal)

```typescript
// Original: Claude Code uses internal Terminal object
class Terminal {
    renderCode(content: string, language: string): void {
        this.write(syntaxHighlight(content, language));
    }
}
```

```go
// Migrated: HelixAgent internal/clis/claude_code/terminal_ui.go
func (ui *TerminalUI) RenderCodeBlock(code, language string, lineNumbers bool) string {
    lexer := lexers.Get(language)
    iterator, _ := lexer.Tokenise(nil, code)
    
    var buf strings.Builder
    ui.formatter.Format(&buf, ui.style, iterator)
    
    return buf.String()
}

// Usage in HelixAgent handlers
func (h *Handler) handleCompletion(c *gin.Context) {
    // ... get completion ...
    formatted := h.terminalUI.RenderCodeBlock(completion.Code, "go", true)
    c.String(200, formatted)
}
```

### Pattern 3: Native Implementation (Multi-Instance Ensemble)

```go
// New: HelixAgent internal/ensemble/multi_instance/coordinator.go
func (c *Coordinator) ExecuteSession(ctx context.Context, sessionID string, task *Task) (*ExecutionResult, error) {
    session, err := c.getSession(sessionID)
    if err != nil {
        return nil, err
    }
    
    // Parallel execution on all instances
    results := c.executeParallel(ctx, session, task)
    
    // Vote/consensus
    winner, confidence := c.vote(results)
    
    return &ExecutionResult{
        Winner:     winner,
        Confidence: confidence,
    }, nil
}
```

---

## Testing Strategy

### Unit Tests

```go
// Example: RepoMap tests
func TestRepoMap_GetRankedTags(t *testing.T) {
    rm := NewRepoMap("testdata/repo", 1024)
    
    ctx := context.Background()
    result, err := rm.GetRankedTags(ctx, "find user authentication", nil)
    
    require.NoError(t, err)
    assert.NotEmpty(t, result.Symbols)
    
    // Verify ranking - auth files should be first
    for i := 0; i < min(5, len(result.Symbols)); i++ {
        assert.True(t, strings.Contains(result.Symbols[i].Name, "auth"))
    }
}
```

### Integration Tests

```go
// Example: Ensemble coordination tests
func TestEnsemble_VotingStrategy(t *testing.T) {
    coord := NewCoordinator(db, logger)
    
    // Create session with 3 instances
    session, err := coord.CreateSession(ctx, StrategyVoting, ParticipantConfig{
        Primary:   {Type: TypeClaudeCode, Config: defaultConfig},
        Critiques: {{Type: TypeAider, Config: defaultConfig}},
    })
    require.NoError(t, err)
    
    // Execute task
    result, err := coord.ExecuteSession(ctx, session.ID, &Task{
        Type:    TaskTypeCompletion,
        Content: "Write a function to validate email",
    })
    require.NoError(t, err)
    
    // Verify consensus
    assert.True(t, result.Confidence > 0.5)
}
```

### E2E Tests

```bash
# Example: Full ensemble test
make test-ensemble

# Runs:
# 1. Starts multiple instances
# 2. Creates ensemble session
# 3. Sends complex multi-file task
# 4. Verifies all instances participate
# 5. Checks final output
```

---

## Performance Targets

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Instance startup | 5s | 100ms | 98% reduction |
| Request latency (p99) | 2s | 500ms | 75% reduction |
| Memory per instance | 500MB | 200MB | 60% reduction |
| Cache hit rate | 10% | 50% | 5x improvement |
| Throughput | 100 req/s | 1000 req/s | 10x improvement |

---

## Deployment Checklist

### Pre-Deployment
- [ ] All 20 critical features implemented
- [ ] Unit tests >80% coverage
- [ ] Integration tests passing
- [ ] Performance benchmarks met
- [ ] Security audit completed
- [ ] Documentation updated

### Database
- [ ] Run schema migration `sql/001_cli_agents_fusion.sql`
- [ ] Verify tables: `agent_instances`, `ensemble_sessions`, `feature_registry`
- [ ] Verify indexes created
- [ ] Verify feature registry populated

### Configuration
- [ ] Update `config.yaml` with new settings
- [ ] Configure instance pool sizes
- [ ] Set resource limits
- [ ] Configure caching parameters

### Deployment
- [ ] Deploy to staging environment
- [ ] Run smoke tests
- [ ] Deploy to production
- [ ] Monitor metrics dashboard
- [ ] Verify alerts configured

### Post-Deployment
- [ ] Verify all agent types functioning
- [ ] Check ensemble coordination
- [ ] Monitor error rates (<0.1%)
- [ ] Collect performance metrics
- [ ] Gather user feedback

---

## Next Actions

### Immediate (This Week)
1. **Review Pass 1-5 documents** - Ensure team alignment
2. **Set up feature branches** - One branch per phase
3. **Create database migration** - Run `sql/001_cli_agents_fusion.sql`
4. **Implement core types** - Start with `internal/clis/types.go`

### Week 1-2
5. **Build instance manager** - `internal/clis/instance_manager.go`
6. **Create SQL schema** - Full schema with indexes
7. **Add synchronization layer** - Distributed locks and CRDTs
8. **Write unit tests** - Core components

### Week 3-4
9. **Implement event bus** - Pub/sub for agent communication
10. **Build feature registry** - Database-backed registry
11. **Create instance pools** - Pool pattern for efficiency
12. **Integration tests** - End-to-end flow

### Ongoing
- **Daily standups** - Track progress against plan
- **Weekly demos** - Show completed features
- **Performance monitoring** - Track metrics from day 1
- **Documentation** - Keep docs updated with code

---

## Appendix: Full Feature Registry

See [Feature Registry](components/feature_registry.md) for complete list of all 156 features across 47 CLI agents.

### Summary by Category

| Category | Features | Ported | Priority |
|----------|----------|--------|----------|
| Core LLM Integration | 15 | 34 | High |
| Code Understanding | 28 | 8 | High |
| Git Operations | 12 | 4 | High |
| Project Management | 10 | 2 | Medium |
| UI/UX | 18 | 3 | Medium |
| Tool Integration | 20 | 6 | High |
| Security | 8 | 2 | High |
| Extensibility | 14 | 1 | Low |
| Performance | 12 | 3 | Medium |
| Collaboration | 6 | 0 | Low |
| Deployment | 5 | 0 | Low |
| AI Features | 8 | 4 | High |

---

**Document Owner:** HelixAgent Core Team  
**Last Review:** 2026-04-03  
**Next Review:** Upon Phase 1 completion
