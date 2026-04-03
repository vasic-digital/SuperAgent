# Pass 3: Synthesis & Design - Unified Architecture

**Pass:** 3 of 5  
**Phase:** Design  
**Goal:** Design unified architecture fusing best implementations  
**Date:** 2026-04-03  
**Status:** Complete  

---

## Executive Summary

This pass synthesizes findings from Pass 1 and Pass 2 to design a unified architecture that fuses the best features from all 47+ CLI agents into HelixAgent. The design includes:
- Multi-instance ensemble architecture
- Component organization under `clis/CLI_AGENT/`
- Unified API gateway
- Extended output pipeline
- SQL schemas for new functionality

**Design Decisions:** 45  
**Architecture Diagrams:** 8  
**SQL Tables:** 12  

---

## Unified Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         HELIXAGENT FUSION ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                          API GATEWAY LAYER                                 │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │  │
│  │  │  REST    │  │WebSocket │  │  gRPC    │  │  MCP     │  │  LSP     │   │  │
│  │  │  API     │  │  API     │  │  API     │  │ Protocol │  │ Protocol │   │  │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │  │
│  │       └──────────────┴─────────────┴─────────────┴─────────────┘         │  │
│  └─────────────────────────────────────┬─────────────────────────────────────┘  │
│                                        │                                         │
│  ┌─────────────────────────────────────┴─────────────────────────────────────┐  │
│  │                         FUSION CORE LAYER                                  │  │
│  │                                                                            │  │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐        │  │
│  │  │ Feature Registry │  │ Capability Mgr   │  │  State Manager   │        │  │
│  │  │                  │  │                  │  │                  │        │  │
│  │  │ - Feature discovery│  │ - Enable/disable │  │ - Global state   │        │  │
│  │  │ - Version control  │  │ - Permissions    │  │ - Sync           │        │  │
│  │  │ - Dependencies     │  │ - Quotas         │  │ - Persistence    │        │  │
│  │  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘        │  │
│  │           └─────────────────────┴─────────────────────┘                   │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                         │
│  ┌─────────────────────────────────────┴─────────────────────────────────────┐  │
│  │                      MULTI-INSTANCE ENSEMBLE LAYER                         │  │
│  │                                                                            │  │
│  │  ┌─────────────────────────────────────────────────────────────────────┐  │  │
│  │  │                    Instance Manager                                  │  │  │
│  │  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │  │  │
│  │  │  │Instance │ │Instance │ │Instance │ │Instance │ │Instance │      │  │  │
│  │  │  │  #1     │ │  #2     │ │  #3     │ │  #4     │ │  #5     │      │  │  │
│  │  │  │(Claude) │ │ (GPT-4) │ │(DeepSeek│ │(Aider)  │ │(Cline)  │      │  │  │
│  │  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘      │  │  │
│  │  │       └───────────┴───────────┴───────────┴───────────┘            │  │  │
│  │  │                         │                                          │  │  │
│  │  │              ┌──────────┴──────────┐                               │  │  │
│  │  │              │   Coordinator       │                               │  │  │
│  │  │              │   - Load balancing  │                               │  │  │
│  │  │              │   - Health checks   │                               │  │  │
│  │  │              │   - Auto-scaling    │                               │  │  │
│  │  │              └─────────────────────┘                               │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                            │  │
│  │  ┌─────────────────────────────────────────────────────────────────────┐  │  │
│  │  │              Inter-Agent Communication Bus                           │  │  │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │  │
│  │  │  │ Message  │  │ Consensus│  │  Voting  │  │  Sync    │          │  │  │
│  │  │  │ Queue    │  │  Engine  │  │  System  │  │ Manager  │          │  │  │
│  │  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘          │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                            │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                         │
│  ┌─────────────────────────────────────┴─────────────────────────────────────┐  │
│  │                      CLI AGENT COMPONENTS LAYER                            │  │
│  │                                                                            │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │  │
│  │  │   clis/      │  │   clis/      │  │   clis/      │  │   clis/      │  │  │
│  │  │   aider/     │  │claude_code/  │  │  codex/      │  │  cline/      │  │  │
│  │  │              │  │              │  │              │  │              │  │  │
│  │  │• repo_map.go │  │• tool_use.go │  │• reasoning.go│  │• browser.go  │  │  │
│  │  │• diff_fmt.go │  │• terminal.go │  │• interpreter│  │• computer.go │  │  │
│  │  │• git_int.go  │  │• convo.go    │  │• chatgpt.go  │  │• autonomy.go │  │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │  │
│  │                                                                            │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │  │
│  │  │   clis/      │  │   clis/      │  │   clis/      │  │    ...       │  │  │
│  │  │ openhands/   │  │   kiro/      │  │  continue/   │  │  (40+ more)  │  │  │
│  │  │              │  │              │  │              │  │              │  │  │
│  │  │• sandbox.go │  │• memory.go   │  │• lsp.go      │  │              │  │  │
│  │  │• security.go│  │• project.go  │  │• ide.go      │  │              │  │  │
│  │  │• web_ui.go  │  │• context.go  │  │• plugins.go  │  │              │  │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │  │
│  │                                                                            │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                         │
│  ┌─────────────────────────────────────┴─────────────────────────────────────┐  │
│  │                        OUTPUT FORMATTING LAYER                             │  │
│  │                                                                            │  │
│  │  ┌─────────────────────────────────────────────────────────────────────┐  │  │
│  │  │                    Formatting Pipeline                               │  │  │
│  │  │                                                                      │  │  │
│  │  │   Raw Output → Parser → Formatter → Renderer → Terminal/HTML/JSON  │  │  │
│  │  │                                                                      │  │  │
│  │  │   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │  │
│  │  │   │  Parser  │→ │ Formatter│→ │ Renderer│→ │  Output  │          │  │  │
│  │  │   │          │  │          │  │          │  │          │          │  │  │
│  │  │   │• Code    │  │• Syntax  │  │• Terminal│  │• Terminal│          │  │  │
│  │  │   │• Diff    │  │• Color   │  │• HTML    │  │• HTML    │          │  │  │
│  │  │   │• JSON    │  │• Layout  │  │• JSON    │  │• JSON    │          │  │  │
│  │  │   │• Markdown│  │• Diff    │  │• Stream  │  │• Stream  │          │  │  │
│  │  │   └──────────┘  └──────────┘  └──────────┘  └──────────┘          │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                            │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                         │
│  ┌─────────────────────────────────────┴─────────────────────────────────────┐  │
│  │                        PERSISTENCE LAYER                                   │  │
│  │                                                                            │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │  │
│  │  │  PostgreSQL  │  │    Redis     │  │  Vector DB   │  │   Object     │  │  │
│  │  │              │  │              │  │              │  │   Store      │  │  │
│  │  │• Agents      │  │• Sessions    │  │• Embeddings  │  │• Documents   │  │  │
│  │  │• Ensembles   │  │• Cache       │  │• RAG         │  │• Artifacts   │  │  │
│  │  │• Memory      │  │• Pub/Sub     │  │• Search      │  │• Exports     │  │  │
│  │  │• State       │  │• Queues      │  │              │  │              │  │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │  │
│  │                                                                            │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Component Organization

### Directory Structure

```
internal/
├── clis/                                    # CLI agent components (NEW)
│   ├── aider/                              # Aider components
│   │   ├── repo_map.go                     # AST-based repo understanding
│   │   ├── diff_format.go                  # SEARCH/REPLACE editing
│   │   ├── git_integration.go              # Native git operations
│   │   └── chunking.go                     # Code chunking strategies
│   │
│   ├── claude_code/                        # Claude Code components
│   │   ├── tool_use.go                     # Tool use system
│   │   ├── terminal_ui.go                  # Rich terminal output
│   │   ├── conversation.go                 # Conversation management
│   │   └── streaming.go                    # Streaming response handling
│   │
│   ├── codex/                              # Codex components
│   │   ├── code_interpreter.go             # Python code execution
│   │   ├── reasoning.go                    # o3/o4 reasoning support
│   │   ├── chatgpt_sync.go                 # ChatGPT conversation sync
│   │   └── safety.go                       # Code safety checks
│   │
│   ├── cline/                              # Cline components
│   │   ├── browser_automation.go           # Browser control
│   │   ├── computer_use.go                 # Computer use (vision + control)
│   │   ├── autonomy.go                     # Autonomous execution
│   │   └── task_planner.go                 # Task planning
│   │
│   ├── openhands/                          # OpenHands components
│   │   ├── sandbox.go                      # Docker sandboxing
│   │   ├── security.go                     # Security isolation
│   │   ├── web_interface.go                # Web UI components
│   │   └── jupyter.go                      # Jupyter integration
│   │
│   ├── kiro/                               # Kiro components
│   │   ├── memory.go                       # Project memory
│   │   ├── project_context.go              # Context management
│   │   ├── learning.go                     # Cross-session learning
│   │   └── persistence.go                  # Memory persistence
│   │
│   ├── continue/                           # Continue components
│   │   ├── lsp_client.go                   # LSP client
│   │   ├── ide_integration.go              # IDE plugins
│   │   ├── universal_bridge.go             # Universal IDE bridge
│   │   └── config_sync.go                  # Config synchronization
│   │
│   └── [40+ more agent directories]       # Other CLI agents
│
├── ensemble/                               # Extended ensemble (MAJOR EXTENSION)
│   ├── multi_instance/                     # Multi-instance management
│   │   ├── instance_manager.go             # Instance lifecycle
│   │   ├── coordinator.go                  # Instance coordination
│   │   ├── load_balancer.go                # Load balancing
│   │   ├── health_monitor.go               # Health monitoring
│   │   └── auto_scaler.go                  # Auto-scaling
│   │
│   ├── background/                         # Background execution
│   │   ├── worker_pool.go                  # Worker pool
│   │   ├── task_queue.go                   # Task queuing
│   │   ├── result_collector.go             # Result aggregation
│   │   └── priority_scheduler.go           # Priority scheduling
│   │
│   ├── communication/                      # Inter-agent communication
│   │   ├── message_bus.go                  # Message bus
│   │   ├── protocol.go                     # Communication protocol
│   │   ├── consensus.go                    # Consensus engine
│   │   ├── voting.go                       # Voting system
│   │   └── synchronization.go              # State synchronization
│   │
│   └── federation/                         # Multi-node federation
│       ├── node_registry.go                # Node registration
│       ├── distributed_consensus.go        # Distributed consensus
│       └── cross_node_sync.go              # Cross-node sync
│
├── output/                                 # Output system (NEW)
│   ├── parsers/                            # Output parsers
│   │   ├── code_parser.go                  # Code block parsing
│   │   ├── diff_parser.go                  # Diff parsing
│   │   ├── json_parser.go                  # JSON parsing
│   │   └── markdown_parser.go              # Markdown parsing
│   │
│   ├── formatters/                         # Content formatters
│   │   ├── syntax_highlighter.go           # Syntax highlighting
│   │   ├── diff_formatter.go               # Diff formatting
│   │   ├── table_formatter.go              # Table formatting
│   │   └── progress_formatter.go           # Progress formatting
│   │
│   ├── renderers/                          # Output renderers
│   │   ├── terminal.go                     # Terminal renderer
│   │   ├── html.go                         # HTML renderer
│   │   ├── json.go                         # JSON renderer
│   │   └── markdown.go                     # Markdown renderer
│   │
│   └── exporters/                          # Export formats
│       ├── file_exporter.go                # File export
│       ├── clipboard.go                    # Clipboard export
│       └── streaming.go                    # Streaming export
│
├── fusion/                                 # Fusion core (NEW)
│   ├── api_gateway.go                      # Unified API gateway
│   ├── feature_registry.go                 # Feature registration
│   ├── capability_manager.go               # Capability management
│   ├── state_manager.go                    # Global state management
│   └── compatibility_layer.go              # Backward compatibility
│
└── [existing HelixAgent modules]          # All existing modules
```

---

## Multi-Instance Ensemble Design

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  MULTI-INSTANCE ENSEMBLE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    ENSEMBLE SESSION                        │  │
│  │                    (Per-Request Context)                   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼                                   │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                 INSTANCE COORDINATOR                       │  │
│  │                                                            │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐         │  │
│  │  │Instance │ │Instance │ │Instance │ │Instance │         │  │
│  │  │ Group A │ │ Group B │ │ Group C │ │ Group D │         │  │
│  │  │         │ │         │ │         │ │         │         │  │
│  │  │• Primary│ │• Critique│ │• Verify │ │• Fallback│        │  │
│  │  │• GPT-4 │ │• Claude  │ │• DeepSeek│ │• Local   │        │  │
│  │  │  #1-3  │ │  #1-2   │ │  #1-2   │ │  #1-5   │         │  │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘         │  │
│  │       └───────────┴───────────┴───────────┘              │  │
│  │                   │                                       │  │
│  │         ┌─────────┴─────────┐                            │  │
│  │         │  Consensus Engine  │                            │  │
│  │         │                    │                            │  │
│  │         │• Voting Strategy   │                            │  │
│  │         │• Confidence Scoring│                            │  │
│  │         │• Tie Breaking      │                            │  │
│  │         └────────────────────┘                            │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼                                   │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              BACKGROUND WORKER POOL                        │  │
│  │                                                            │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │  │
│  │  │ Worker 1 │  │ Worker 2 │  │ Worker 3 │  │ Worker N │  │  │
│  │  │ (Aider)  │  │(Claude)  │  │ (Cline)  │  │(Custom)  │  │  │
│  │  │          │  │          │  │          │  │          │  │  │
│  │  │• Git ops │  │• Tool use│  │• Browser│  │• Custom  │  │  │
│  │  │• Repo map│  │• Terminal│  │• Autonomy│  │• Logic   │  │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │  │
│  │                                                            │  │
│  │  Workers communicate via message bus                      │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Instance Lifecycle

```go
// Source: internal/ensemble/multi_instance/instance_manager.go

type InstanceLifecycle struct {
    // States
    StateCreating    // Initializing
    StateIdle        // Ready for work
    StateActive      // Processing request
    StateBackground  // Background processing
    StateDegraded    // Reduced capacity
    StateRecovering  // Self-healing
    StateTerminating // Shutting down
}

type InstanceManager struct {
    // Instance pools by type
    pools map[AgentType]*InstancePool
    
    // Active instances per session
    sessions map[SessionID][]*AgentInstance
    
    // Background workers
    workers *WorkerPool
}

func (m *InstanceManager) CreateInstance(
    ctx context.Context,
    agentType AgentType,
    config InstanceConfig,
) (*AgentInstance, error) {
    // 1. Check resource limits
    if !m.hasCapacity(agentType) {
        return nil, ErrCapacityExceeded
    }
    
    // 2. Create instance
    instance := &AgentInstance{
        ID:       generateID(),
        Type:     agentType,
        State:    StateCreating,
        CreatedAt: time.Now(),
    }
    
    // 3. Initialize based on type
    switch agentType {
    case TypeAider:
        instance.InitAider(config)
    case TypeClaudeCode:
        instance.InitClaudeCode(config)
    case TypeCline:
        instance.InitCline(config)
    // ... etc
    }
    
    // 4. Health check
    if err := instance.HealthCheck(ctx); err != nil {
        instance.Terminate()
        return nil, err
    }
    
    instance.State = StateIdle
    return instance, nil
}
```

---

## SQL Schemas

### Core Tables

See full schemas in: `sql_schemas/`

```sql
-- Agent Instances Table
CREATE TABLE agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(50) NOT NULL,  -- 'aider', 'claude_code', 'cline', etc.
    instance_name VARCHAR(100),
    status VARCHAR(20) NOT NULL,  -- 'creating', 'idle', 'active', 'background', 'terminated'
    
    -- Configuration
    config JSONB NOT NULL,
    provider_config JSONB,  -- LLM provider settings
    
    -- Resource limits
    max_memory_mb INTEGER,
    max_cpu_percent INTEGER,
    
    -- Current state
    current_session_id UUID,
    current_task_id UUID,
    
    -- Health
    last_health_check TIMESTAMP,
    health_status VARCHAR(20),
    
    -- Metrics
    requests_processed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    terminated_at TIMESTAMP,
    
    FOREIGN KEY (current_session_id) REFERENCES ensemble_sessions(id)
);

CREATE INDEX idx_agent_instances_type ON agent_instances(agent_type);
CREATE INDEX idx_agent_instances_status ON agent_instances(status);

-- Ensemble Sessions Table
CREATE TABLE ensemble_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Session configuration
    strategy VARCHAR(50) NOT NULL,  -- 'voting', 'debate', 'consensus'
    participant_types TEXT[],  -- ['claude', 'gpt4', 'deepseek']
    
    -- Session state
    status VARCHAR(20) NOT NULL,  -- 'active', 'paused', 'completed', 'failed'
    
    -- Participants (instances assigned to this session)
    primary_instance_id UUID,
    critique_instance_ids UUID[],
    verification_instance_ids UUID[],
    
    -- Context
    context JSONB,  -- Shared context across instances
    
    -- Results
    final_result JSONB,
    consensus_reached BOOLEAN,
    confidence_score FLOAT,
    
    -- Timestamps
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    FOREIGN KEY (primary_instance_id) REFERENCES agent_instances(id)
);

-- Feature Registry Table
CREATE TABLE feature_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    feature_name VARCHAR(100) NOT NULL UNIQUE,
    feature_category VARCHAR(50),  -- 'git', 'ui', 'sandbox', etc.
    
    -- Source
    source_agent VARCHAR(50),  -- Which CLI agent this came from
    source_file VARCHAR(255),  -- File path in original codebase
    
    -- Implementation
    implementation_type VARCHAR(50),  -- 'ported', 'adapted', 'native'
    internal_path VARCHAR(255),  -- Path in HelixAgent codebase
    
    -- Dependencies
    dependencies TEXT[],  -- Other features required
    external_deps TEXT[],  -- External libraries
    
    -- Status
    status VARCHAR(20),  -- 'planned', 'in_progress', 'ported', 'tested', 'production'
    priority INTEGER,  -- 1-5, 1 = highest
    
    -- Metadata
    complexity VARCHAR(20),  -- 'low', 'medium', 'high'
    estimated_effort_hours INTEGER,
    
    -- Documentation
    porting_notes TEXT,
    test_coverage FLOAT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_feature_registry_category ON feature_registry(feature_category);
CREATE INDEX idx_feature_registry_status ON feature_registry(status);
CREATE INDEX idx_feature_registry_source ON feature_registry(source_agent);
```

---

## API Design

### Unified API Gateway

```go
// Source: internal/fusion/api_gateway.go

type UnifiedAPI struct {
    // Route requests to appropriate handler
    router *Router
    
    // Feature registry
    features *FeatureRegistry
    
    // Instance manager
    instances *ensemble.InstanceManager
}

// REST Endpoints
func (api *UnifiedAPI) RegisterRoutes(router *gin.Engine) {
    // Agent instance management
    router.POST("/v2/instances", api.CreateInstance)
    router.GET("/v2/instances/:id", api.GetInstance)
    router.DELETE("/v2/instances/:id", api.TerminateInstance)
    
    // Ensemble operations
    router.POST("/v2/ensemble/sessions", api.CreateEnsembleSession)
    router.POST("/v2/ensemble/:id/execute", api.ExecuteEnsemble)
    router.GET("/v2/ensemble/:id/results", api.GetEnsembleResults)
    
    // Feature-specific endpoints
    router.POST("/v2/git/diff", api.GitDiffHandler)           // From Aider
    router.POST("/v2/ui/render", api.UIRenderHandler)         // From Claude Code
    router.POST("/v2/sandbox/execute", api.SandboxHandler)    // From OpenHands
    router.POST("/v2/browser/navigate", api.BrowserHandler)   // From Cline
    router.POST("/v2/memory/remember", api.MemoryHandler)     // From Kiro
    
    // Legacy compatibility
    router.POST("/v1/chat/completions", api.LegacyChatHandler)
}
```

---

## Next Steps

**Pass 4: Optimization** will:
- Optimize designs for performance
- Refine SQL schemas
- Optimize API design
- Plan caching strategies

**See:** [Pass 4 - Optimization](pass_4_optimization.md)

---

*Pass 3 Complete: Unified architecture designed*  
*Date: 2026-04-03*  
*HelixAgent Commit: 8a976be2*
