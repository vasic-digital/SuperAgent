# Pass 5: Final Implementation Plan - Roadmap & Execution

**Pass:** 5 of 5  
**Phase:** Finalization  
**Goal:** Create detailed implementation roadmap with exact code references  
**Date:** 2026-04-03  
**Status:** Complete  

---

## Executive Summary

This final pass consolidates all previous analysis into a detailed implementation roadmap. It provides:
- Phase-by-phase execution plan
- Exact code changes required
- Testing strategy
- Deployment checklist
- Risk mitigation

**Implementation Phases:** 5  
**Estimated Duration:** 24 weeks  
**Team Size:** 4-6 engineers  

---

## Implementation Roadmap

### Timeline Overview

```
Week:  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24
       ├──── Phase 1: Foundation ────┤
                                   ├──── Phase 2: Ensemble ────┤
                                                               ├───────── Phase 3: CLI Integration ──────────┤
                                                                                                               ├── Phase 4: Output ──┤
                                                                                                                                       ├─ Phase 5: Test ─┤
```

---

## Phase 1: Foundation Layer (Weeks 1-4)

### Week 1: Database Schema & Core Types

**Day 1-2: SQL Schema Implementation**

Create `sql/001_cli_agents_fusion.sql`:

```sql
-- Agent instances table
CREATE TABLE agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(50) NOT NULL,
    instance_name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'creating',
    config JSONB NOT NULL DEFAULT '{}',
    provider_config JSONB DEFAULT '{}',
    max_memory_mb INTEGER,
    max_cpu_percent INTEGER,
    current_session_id UUID,
    current_task_id UUID,
    last_health_check TIMESTAMP,
    health_status VARCHAR(20) DEFAULT 'unknown',
    requests_processed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    terminated_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_agent_instances_type ON agent_instances(agent_type);
CREATE INDEX idx_agent_instances_status ON agent_instances(status);
CREATE INDEX idx_agent_instances_session ON agent_instances(current_session_id);

-- Ensemble sessions table
CREATE TABLE ensemble_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy VARCHAR(50) NOT NULL,
    participant_types TEXT[] NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    primary_instance_id UUID,
    critique_instance_ids UUID[],
    verification_instance_ids UUID[],
    context JSONB DEFAULT '{}',
    final_result JSONB,
    consensus_reached BOOLEAN,
    confidence_score FLOAT,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Feature registry table
CREATE TABLE feature_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_name VARCHAR(100) NOT NULL UNIQUE,
    feature_category VARCHAR(50),
    source_agent VARCHAR(50),
    source_file VARCHAR(255),
    implementation_type VARCHAR(50),
    internal_path VARCHAR(255),
    dependencies TEXT[],
    external_deps TEXT[],
    status VARCHAR(20) DEFAULT 'planned',
    priority INTEGER DEFAULT 3,
    complexity VARCHAR(20),
    estimated_effort_hours INTEGER,
    porting_notes TEXT,
    test_coverage FLOAT DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Populate feature registry with all 20 critical features
INSERT INTO feature_registry (feature_name, feature_category, source_agent, priority, complexity, estimated_effort_hours) VALUES
('repo_map', 'code_understanding', 'aider', 1, 'high', 120),
('diff_format', 'editing', 'aider', 2, 'medium', 40),
('git_integration', 'vcs', 'aider', 1, 'high', 80),
('terminal_ui', 'output', 'claude_code', 1, 'high', 100),
('tool_use', 'tools', 'claude_code', 2, 'medium', 60),
('code_interpreter', 'execution', 'codex', 3, 'high', 80),
('reasoning_models', 'llm', 'codex', 3, 'low', 20),
('browser_automation', 'automation', 'cline', 3, 'high', 160),
('computer_use', 'automation', 'cline', 4, 'high', 200),
('autonomy', 'execution', 'cline', 3, 'high', 120),
('sandbox', 'security', 'openhands', 1, 'high', 100),
('security_isolation', 'security', 'openhands', 2, 'medium', 60),
('project_memory', 'memory', 'kiro', 2, 'medium', 80),
('cross_session_memory', 'memory', 'kiro', 3, 'medium', 60),
('lsp_client', 'ide', 'continue', 2, 'high', 100),
('universal_ide', 'ide', 'continue', 3, 'high', 120),
('ensemble_coordination', 'ensemble', 'helixagent', 1, 'high', 140),
('instance_management', 'ensemble', 'helixagent', 1, 'high', 120),
('synchronization', 'ensemble', 'helixagent', 2, 'high', 100),
('streaming_pipeline', 'output', 'helixagent', 2, 'medium', 80);
```

**Day 3-5: Core Type Definitions**

Create `internal/clis/types.go`:

```go
package clis

import (
    "context"
    "time"
)

// AgentType represents CLI agent type
type AgentType string

const (
    TypeAider       AgentType = "aider"
    TypeClaudeCode  AgentType = "claude_code"
    TypeCodex       AgentType = "codex"
    TypeCline       AgentType = "cline"
    TypeOpenHands   AgentType = "openhands"
    TypeKiro        AgentType = "kiro"
    TypeContinue    AgentType = "continue"
    // ... etc
)

// InstanceStatus represents agent instance state
type InstanceStatus string

const (
    StatusCreating    InstanceStatus = "creating"
    StatusIdle        InstanceStatus = "idle"
    StatusActive      InstanceStatus = "active"
    StatusBackground  InstanceStatus = "background"
    StatusDegraded    InstanceStatus = "degraded"
    StatusRecovering  InstanceStatus = "recovering"
    StatusTerminating InstanceStatus = "terminating"
    StatusTerminated  InstanceStatus = "terminated"
)

// AgentInstance represents a running CLI agent instance
type AgentInstance struct {
    ID        string
    Type      AgentType
    Name      string
    Status    InstanceStatus
    
    // Configuration
    Config    InstanceConfig
    Provider  ProviderConfig
    
    // Resources
    Resources ResourceLimits
    
    // State
    SessionID string
    TaskID    string
    State     map[string]interface{}
    
    // Lifecycle
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // Runtime
    Client    interface{}  // Type-specific client
    
    // Channels for communication
    RequestCh  chan *Request
    ResponseCh chan *Response
    EventCh    chan *Event
}

type InstanceConfig struct {
    MaxMemoryMB   int
    MaxCPUPercent int
    Timeout       time.Duration
    
    // Type-specific config
    Extra map[string]interface{}
}

type ResourceLimits struct {
    MemoryMB   int64
    CPUPercent float64
    DiskMB     int64
}

type ProviderConfig struct {
    Name       string
    APIKey     string
    BaseURL    string
    Model      string
    Parameters map[string]interface{}
}

// Request/Response types for inter-instance communication
type Request struct {
    ID        string
    Type      RequestType
    Payload   interface{}
    Context   context.Context
    Timeout   time.Duration
}

type Response struct {
    RequestID string
    Success   bool
    Result    interface{}
    Error     error
    Duration  time.Duration
}

type Event struct {
    ID        string
    Type      EventType
    Source    string
    Payload   interface{}
    Timestamp time.Time
}

type RequestType string
const (
    RequestTypeExecute   RequestType = "execute"
    RequestTypeQuery     RequestType = "query"
    RequestTypeToolCall  RequestType = "tool_call"
    RequestTypeStream    RequestType = "stream"
)

type EventType string
const (
    EventTypeStatus    EventType = "status"
    EventTypeProgress  EventType = "progress"
    EventTypeError     EventType = "error"
    EventTypeComplete  EventType = "complete"
)
```

### Week 2: Base Instance Management

Create `internal/clis/instance_manager.go`:

```go
package clis

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// InstanceManager manages lifecycle of CLI agent instances
type InstanceManager struct {
    db     *sql.DB
    logger *log.Logger
    
    // Instance pools by type
    pools map[AgentType]*InstancePool
    
    // Active instances
    instances map[string]*AgentInstance
    mu        sync.RWMutex
    
    // Background workers
    workers *WorkerPool
    
    // Event bus
    eventBus *EventBus
}

func NewInstanceManager(db *sql.DB, logger *log.Logger) *InstanceManager {
    return &InstanceManager{
        db:        db,
        logger:    logger,
        pools:     make(map[AgentType]*InstancePool),
        instances: make(map[string]*AgentInstance),
        workers:   NewWorkerPool(100),
        eventBus:  NewEventBus(),
    }
}

func (m *InstanceManager) CreateInstance(
    ctx context.Context,
    agentType AgentType,
    config InstanceConfig,
) (*AgentInstance, error) {
    // Check feature availability
    if !m.isFeatureAvailable(agentType) {
        return nil, fmt.Errorf("agent type %s not available", agentType)
    }
    
    instance := &AgentInstance{
        ID:        generateID(),
        Type:      agentType,
        Name:      fmt.Sprintf("%s-%s", agentType, generateShortID()),
        Status:    StatusCreating,
        Config:    config,
        CreatedAt: time.Now(),
        RequestCh:  make(chan *Request, 10),
        ResponseCh: make(chan *Response, 10),
        EventCh:    make(chan *Event, 10),
    }
    
    // Initialize based on type
    if err := m.initializeInstance(ctx, instance); err != nil {
        return nil, fmt.Errorf("initialize instance: %w", err)
    }
    
    // Store in database
    if err := m.persistInstance(ctx, instance); err != nil {
        return nil, fmt.Errorf("persist instance: %w", err)
    }
    
    // Register in memory
    m.mu.Lock()
    m.instances[instance.ID] = instance
    m.mu.Unlock()
    
    // Start event loop
    go m.instanceEventLoop(instance)
    
    instance.Status = StatusIdle
    instance.UpdatedAt = time.Now()
    
    m.logger.Printf("Created instance %s of type %s", instance.ID, agentType)
    
    return instance, nil
}

func (m *InstanceManager) initializeInstance(ctx context.Context, inst *AgentInstance) error {
    switch inst.Type {
    case TypeAider:
        return m.initAiderInstance(ctx, inst)
    case TypeClaudeCode:
        return m.initClaudeCodeInstance(ctx, inst)
    case TypeCodex:
        return m.initCodexInstance(ctx, inst)
    case TypeCline:
        return m.initClineInstance(ctx, inst)
    case TypeOpenHands:
        return m.initOpenHandsInstance(ctx, inst)
    case TypeKiro:
        return m.initKiroInstance(ctx, inst)
    case TypeContinue:
        return m.initContinueInstance(ctx, inst)
    default:
        return fmt.Errorf("unknown agent type: %s", inst.Type)
    }
}

func (m *InstanceManager) instanceEventLoop(inst *AgentInstance) {
    for {
        select {
        case req := <-inst.RequestCh:
            resp := m.handleRequest(inst, req)
            inst.ResponseCh <- resp
            
        case event := <-inst.EventCh:
            m.eventBus.Publish(event)
            
        case <-time.After(30 * time.Second):
            // Health check
            if err := m.healthCheck(inst); err != nil {
                m.logger.Printf("Health check failed for %s: %v", inst.ID, err)
            }
        }
    }
}

func (m *InstanceManager) TerminateInstance(ctx context.Context, id string) error {
    m.mu.Lock()
    inst, exists := m.instances[id]
    m.mu.Unlock()
    
    if !exists {
        return fmt.Errorf("instance %s not found", id)
    }
    
    inst.Status = StatusTerminating
    
    // Type-specific cleanup
    if err := m.cleanupInstance(ctx, inst); err != nil {
        m.logger.Printf("Cleanup error for %s: %v", id, err)
    }
    
    // Close channels
    close(inst.RequestCh)
    close(inst.ResponseCh)
    close(inst.EventCh)
    
    // Update database
    _, err := m.db.ExecContext(ctx,
        "UPDATE agent_instances SET status = 'terminated', terminated_at = NOW() WHERE id = $1",
        id,
    )
    
    // Remove from memory
    m.mu.Lock()
    delete(m.instances, id)
    m.mu.Unlock()
    
    m.logger.Printf("Terminated instance %s", id)
    
    return err
}
```

### Week 3-4: Synchronization Primitives

Create `internal/ensemble/synchronization/`:

```go
// internal/ensemble/synchronization/manager.go

package synchronization

import (
    "context"
    "sync"
    "time"
)

// SyncManager manages distributed state synchronization
type SyncManager struct {
    db     *sql.DB
    
    // Distributed locks
    locks map[string]*DistributedLock
    mu    sync.RWMutex
    
    // CRDTs for conflict-free updates
    crdts map[string]CRDT
}

// DistributedLock for distributed locking
type DistributedLock struct {
    name    string
    owner   string
    expires time.Time
    
    // Local tracking
    held    bool
    mu      sync.Mutex
}

func (sm *SyncManager) AcquireLock(ctx context.Context, name, owner string, ttl time.Duration) (*DistributedLock, error) {
    expires := time.Now().Add(ttl)
    
    // Try to acquire in database
    result, err := sm.db.ExecContext(ctx, `
        INSERT INTO distributed_locks (name, owner, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (name) DO UPDATE
        SET owner = $2, expires_at = $3, acquired_at = NOW()
        WHERE distributed_locks.expires_at < NOW()
    `, name, owner, expires)
    
    if err != nil {
        return nil, err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return nil, ErrLockNotAvailable
    }
    
    lock := &DistributedLock{
        name:    name,
        owner:   owner,
        expires: expires,
        held:    true,
    }
    
    sm.mu.Lock()
    sm.locks[name] = lock
    sm.mu.Unlock()
    
    // Auto-renewal goroutine
    go sm.renewLock(lock, ttl/2)
    
    return lock, nil
}

func (sm *SyncManager) renewLock(lock *DistributedLock, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        lock.mu.Lock()
        if !lock.held {
            lock.mu.Unlock()
            return
        }
        lock.mu.Unlock()
        
        // Renew in database
        newExpires := time.Now().Add(interval * 2)
        _, err := sm.db.Exec("UPDATE distributed_locks SET expires_at = $1 WHERE name = $2",
            newExpires, lock.name)
        if err != nil {
            // Lock lost
            lock.mu.Lock()
            lock.held = false
            lock.mu.Unlock()
            return
        }
    }
}
```

---

## Phase 2: Ensemble Extension (Weeks 5-8)

### Week 5-6: Multi-Instance Coordinator

Create `internal/ensemble/multi_instance/coordinator.go`:

```go
package multi_instance

import (
    "context"
    "sync"
    "time"
)

// Coordinator manages multiple agent instances for ensemble execution
type Coordinator struct {
    instanceMgr *clis.InstanceManager
    syncMgr     *synchronization.SyncManager
    
    // Active sessions
    sessions map[string]*EnsembleSession
    mu       sync.RWMutex
    
    // Load balancer
    loadBalancer LoadBalancer
    
    // Health monitor
    healthMonitor *HealthMonitor
}

type EnsembleSession struct {
    ID          string
    Strategy    EnsembleStrategy
    Status      SessionStatus
    
    // Instances
    Primary     *clis.AgentInstance
    Critiques   []*clis.AgentInstance
    Verifiers   []*clis.AgentInstance
    Fallbacks   []*clis.AgentInstance
    
    // Context
    Context     *SessionContext
    
    // Communication
    MessageBus  *MessageBus
    
    // Results
    Results     chan *AgentResult
    Consensus   *ConsensusResult
    
    // Lifecycle
    CreatedAt   time.Time
    StartedAt   *time.Time
    CompletedAt *time.Time
}

func (c *Coordinator) CreateSession(
    ctx context.Context,
    strategy EnsembleStrategy,
    participants ParticipantConfig,
) (*EnsembleSession, error) {
    session := &EnsembleSession{
        ID:        generateID(),
        Strategy:  strategy,
        Status:    SessionStatusCreating,
        Context:   NewSessionContext(),
        MessageBus: NewMessageBus(),
        Results:   make(chan *AgentResult, 100),
        CreatedAt: time.Now(),
    }
    
    // Create primary instance
    primary, err := c.instanceMgr.CreateInstance(ctx, participants.Primary.Type, participants.Primary.Config)
    if err != nil {
        return nil, fmt.Errorf("create primary: %w", err)
    }
    session.Primary = primary
    
    // Create critique instances
    for _, cfg := range participants.Critiques {
        inst, err := c.instanceMgr.CreateInstance(ctx, cfg.Type, cfg.Config)
        if err != nil {
            return nil, fmt.Errorf("create critique: %w", err)
        }
        session.Critiques = append(session.Critiques, inst)
    }
    
    // Create verifier instances
    for _, cfg := range participants.Verifiers {
        inst, err := c.instanceMgr.CreateInstance(ctx, cfg.Type, cfg.Config)
        if err != nil {
            return nil, fmt.Errorf("create verifier: %w", err)
        }
        session.Verifiers = append(session.Verifiers, inst)
    }
    
    // Register session
    c.mu.Lock()
    c.sessions[session.ID] = session
    c.mu.Unlock()
    
    // Persist to database
    if err := c.persistSession(ctx, session); err != nil {
        return nil, err
    }
    
    return session, nil
}

func (c *Coordinator) ExecuteSession(ctx context.Context, sessionID string, task *Task) (*ExecutionResult, error) {
    c.mu.RLock()
    session, exists := c.sessions[sessionID]
    c.mu.RUnlock()
    
    if !exists {
        return nil, ErrSessionNotFound
    }
    
    now := time.Now()
    session.StartedAt = &now
    session.Status = SessionStatusActive
    
    // Execute based on strategy
    switch session.Strategy {
    case StrategyVoting:
        return c.executeVotingStrategy(ctx, session, task)
    case StrategyDebate:
        return c.executeDebateStrategy(ctx, session, task)
    case StrategyConsensus:
        return c.executeConsensusStrategy(ctx, session, task)
    case StrategyPipeline:
        return c.executePipelineStrategy(ctx, session, task)
    default:
        return nil, fmt.Errorf("unknown strategy: %s", session.Strategy)
    }
}

func (c *Coordinator) executeVotingStrategy(
    ctx context.Context,
    session *EnsembleSession,
    task *Task,
) (*ExecutionResult, error) {
    // Send task to all participants concurrently
    var wg sync.WaitGroup
    results := make(map[string]*AgentResult)
    var mu sync.Mutex
    
    // Primary
    wg.Add(1)
    go func() {
        defer wg.Done()
        result := c.executeOnInstance(ctx, session.Primary, task)
        mu.Lock()
        results[session.Primary.ID] = result
        mu.Unlock()
    }()
    
    // Critiques
    for _, inst := range session.Critiques {
        wg.Add(1)
        go func(i *clis.AgentInstance) {
            defer wg.Done()
            result := c.executeOnInstance(ctx, i, task)
            mu.Lock()
            results[i.ID] = result
            mu.Unlock()
        }(inst)
    }
    
    // Wait with timeout
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All completed
    case <-time.After(task.Timeout):
        // Timeout - use partial results
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    // Count votes
    voteCounts := make(map[string]int)
    for _, result := range results {
        if result.Success {
            key := result.Hash() // Hash of response content
            voteCounts[key]++
        }
    }
    
    // Find winner
    var winner string
    maxVotes := 0
    for key, count := range voteCounts {
        if count > maxVotes {
            maxVotes = count
            winner = key
        }
    }
    
    // Calculate confidence
    total := len(results)
    confidence := float64(maxVotes) / float64(total)
    
    return &ExecutionResult{
        Winner:      winner,
        Confidence:  confidence,
        AllResults:  results,
        Consensus:   maxVotes > total/2,
    }, nil
}
```

### Week 7-8: Background Worker Pool

Create `internal/ensemble/background/worker_pool.go`:

```go
package background

import (
    "context"
    "sync"
    "sync/atomic"
)

// WorkerPool manages background agent workers
type WorkerPool struct {
    size    int
    workers []*Worker
    
    // Task queue
    taskQueue chan *BackgroundTask
    
    // Result collection
    resultQueue chan *TaskResult
    
    // Metrics
    tasksProcessed uint64
    tasksFailed    uint64
    
    // Control
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

type Worker struct {
    id       int
    instance *clis.AgentInstance
    pool     *WorkerPool
}

type BackgroundTask struct {
    ID        string
    Type      TaskType
    Payload   interface{}
    Priority  int
    Callback  func(*TaskResult)
}

type TaskResult struct {
    TaskID   string
    Success  bool
    Result   interface{}
    Error    error
    Duration time.Duration
}

func NewWorkerPool(size int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool{
        size:        size,
        taskQueue:   make(chan *BackgroundTask, size*10),
        resultQueue: make(chan *TaskResult, size*10),
        ctx:         ctx,
        cancel:      cancel,
    }
}

func (wp *WorkerPool) Start() error {
    for i := 0; i < wp.size; i++ {
        worker := &Worker{
            id:   i,
            pool: wp,
        }
        wp.workers = append(wp.workers, worker)
        
        wp.wg.Add(1)
        go worker.run()
    }
    
    // Start result collector
    go wp.collectResults()
    
    return nil
}

func (w *Worker) run() {
    defer w.pool.wg.Done()
    
    for {
        select {
        case task := <-w.pool.taskQueue:
            start := time.Now()
            
            // Execute task
            result := w.execute(task)
            result.Duration = time.Since(start)
            
            // Send to result queue
            select {
            case w.pool.resultQueue <- result:
            case <-w.pool.ctx.Done():
                return
            }
            
            // Update metrics
            if result.Success {
                atomic.AddUint64(&w.pool.tasksProcessed, 1)
            } else {
                atomic.AddUint64(&w.pool.tasksFailed, 1)
            }
            
        case <-w.pool.ctx.Done():
            return
        }
    }
}

func (w *Worker) execute(task *BackgroundTask) *TaskResult {
    // Route to appropriate handler based on task type
    switch task.Type {
    case TaskTypeGitOperation:
        return w.executeGitOperation(task)
    case TaskTypeCodeAnalysis:
        return w.executeCodeAnalysis(task)
    case TaskTypeDocumentation:
        return w.executeDocumentation(task)
    case TaskTypeTesting:
        return w.executeTesting(task)
    default:
        return &TaskResult{
            TaskID:  task.ID,
            Success: false,
            Error:   fmt.Errorf("unknown task type: %s", task.Type),
        }
    }
}

func (w *Worker) executeGitOperation(task *BackgroundTask) *TaskResult {
    // Use Aider instance for git operations
    // Implementation...
}

func (w *Worker) executeCodeAnalysis(task *BackgroundTask) *TaskResult {
    // Use multiple instances for parallel analysis
    // Implementation...
}

func (wp *WorkerPool) Submit(ctx context.Context, task *BackgroundTask) error {
    select {
    case wp.taskQueue <- task:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (wp *WorkerPool) collectResults() {
    for result := range wp.resultQueue {
        // Find task callback
        if task := wp.findTask(result.TaskID); task != nil {
            if task.Callback != nil {
                go task.Callback(result)
            }
        }
    }
}
```

---

## Phase 3: CLI Agent Integration (Weeks 9-16)

### Week 9-10: Aider Components

Create `internal/clis/aider/`:

```go
// internal/clis/aider/repo_map.go

package aider

import (
    "context"
    "path/filepath"
    "strings"
)

// RepoMap provides AST-based repository understanding
type RepoMap struct {
    rootDir    string
    mapTokens  int
    
    // Parsers
    tsParser   *TreeSitterParser
    ctags      *CTagsParser
    
    // Cache
    symbolCache *SymbolCache
    fileCache   *FileCache
}

func NewRepoMap(rootDir string, mapTokens int) *RepoMap {
    return &RepoMap{
        rootDir:     rootDir,
        mapTokens:   mapTokens,
        tsParser:    NewTreeSitterParser(),
        ctags:       NewCTagsParser(),
        symbolCache: NewSymbolCache(),
        fileCache:   NewFileCache(),
    }
}

func (rm *RepoMap) GetRankedTags(ctx context.Context, query string, mentionedFiles []string) (*RepoContext, error) {
    // 1. Find matching files
    files, err := rm.findMatchingFiles(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // 2. Extract symbols
    symbols := make([]*Symbol, 0)
    for _, file := range files {
        fileSymbols, err := rm.extractSymbols(ctx, file)
        if err != nil {
            continue
        }
        symbols = append(symbols, fileSymbols...)
    }
    
    // 3. Rank symbols
    ranked := rm.rankSymbols(symbols, query, mentionedFiles)
    
    // 4. Format for LLM within token budget
    return rm.formatForLLM(ranked, rm.mapTokens), nil
}

func (rm *RepoMap) findMatchingFiles(ctx context.Context, query string) ([]string, error) {
    var matches []string
    
    // Strategy 1: Direct file mention
    for _, file := range rm.listAllFiles() {
        if strings.Contains(query, filepath.Base(file)) {
            matches = append(matches, file)
        }
    }
    
    // Strategy 2: Fuzzy filename matching
    // Implementation using fuzzy matching library
    
    // Strategy 3: Ctags lookup
    ctagsMatches := rm.ctags.Find(query)
    matches = append(matches, ctagsMatches...)
    
    return unique(matches), nil
}

func (rm *RepoMap) extractSymbols(ctx context.Context, file string) ([]*Symbol, error) {
    // Check cache
    if cached := rm.symbolCache.Get(file); cached != nil {
        return cached, nil
    }
    
    // Parse with tree-sitter
    lang := detectLanguage(file)
    tree := rm.tsParser.Parse(file, lang)
    
    // Query for definitions
    query := rm.tsParser.Query(lang, `
        (function_definition name: (identifier) @func)
        (class_definition name: (identifier) @class)
        (method_definition name: (identifier) @method)
    `)
    
    captures := query.Captures(tree.Root())
    
    symbols := make([]*Symbol, 0, len(captures))
    for _, capture := range captures {
        symbol := &Symbol{
            Name:     capture.Node.Text(),
            Type:     capture.Name, // 'func', 'class', 'method'
            File:     file,
            Line:     capture.Node.StartLine(),
            Language: lang,
        }
        symbols = append(symbols, symbol)
    }
    
    // Cache results
    rm.symbolCache.Set(file, symbols)
    
    return symbols, nil
}

func (rm *RepoMap) rankSymbols(symbols []*Symbol, query string, mentionedFiles []string) []*RankedSymbol {
    ranked := make([]*RankedSymbol, 0, len(symbols))
    
    // Build reference graph
    graph := rm.buildReferenceGraph()
    
    for _, sym := range symbols {
        score := 0.0
        
        // Factor 1: Distance from mentioned files
        if len(mentionedFiles) > 0 {
            minDist := graph.MinDistance(sym.File, mentionedFiles)
            score += 1.0 / (1.0 + float64(minDist))
        }
        
        // Factor 2: Symbol type weight
        weights := map[string]float64{
            "class":  1.0,
            "method": 0.9,
            "func":   0.9,
            "var":    0.5,
        }
        score += weights[sym.Type]
        
        // Factor 3: Name similarity to query
        score += fuzzyScore(sym.Name, query) / 2.0
        
        // Factor 4: Reference count
        refCount := graph.ReferenceCount(sym)
        score += min(float64(refCount)/10.0, 0.5)
        
        ranked = append(ranked, &RankedSymbol{
            Symbol: sym,
            Score:  score,
        })
    }
    
    // Sort by score
    sort.Slice(ranked, func(i, j int) bool {
        return ranked[i].Score > ranked[j].Score
    })
    
    return ranked
}
```

Create `internal/clis/aider/diff_format.go`:

```go
package aider

import (
    "fmt"
    "os"
    "strings"
)

// DiffFormat implements SEARCH/REPLACE block editing
type DiffFormat struct{}

type EditBlock struct {
    Search  string
    Replace string
    File    string
}

// ParseEditBlocks parses SEARCH/REPLACE blocks from text
func (df *DiffFormat) ParseEditBlocks(text string) ([]*EditBlock, error) {
    pattern := regexp.MustCompile(`(?m)^<<<<<<< SEARCH\n(.*?)\n=======\n(.*?)\n>>>>>>> REPLACE$`)
    
    matches := pattern.FindAllStringSubmatch(text, -1)
    blocks := make([]*EditBlock, 0, len(matches))
    
    for _, match := range matches {
        blocks = append(blocks, &EditBlock{
            Search:  match[1],
            Replace: match[2],
        })
    }
    
    return blocks, nil
}

// ApplyEditBlocks applies edit blocks to file
func (df *DiffFormat) ApplyEditBlocks(filePath string, blocks []*EditBlock) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("read file: %w", err)
    }
    
    contentStr := string(content)
    
    for i, block := range blocks {
        // Validate search text exists
        if !strings.Contains(contentStr, block.Search) {
            return fmt.Errorf("block %d: search text not found", i)
        }
        
        // Apply replacement (only first occurrence)
        contentStr = strings.Replace(contentStr, block.Search, block.Replace, 1)
    }
    
    // Write back
    if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
        return fmt.Errorf("write file: %w", err)
    }
    
    return nil
}

// ValidateEditBlocks validates blocks without applying
func (df *DiffFormat) ValidateEditBlocks(filePath string, blocks []*EditBlock) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }
    
    contentStr := string(content)
    
    for i, block := range blocks {
        if !strings.Contains(contentStr, block.Search) {
            return fmt.Errorf("block %d: search text not found in %s", i, filePath)
        }
    }
    
    return nil
}
```

### Week 11-12: Claude Code Components

Create `internal/clis/claude_code/terminal_ui.go`:

```go
package claude_code

import (
    "fmt"
    "strings"
    
    "github.com/alecthomas/chroma"
    "github.com/alecthomas/chroma/formatters"
    "github.com/alecthomas/chroma/lexers"
    "github.com/alecthomas/chroma/styles"
    "github.com/fatih/color"
)

// TerminalUI provides rich terminal output
type TerminalUI struct {
    formatter Formatter
    style     *chroma.Style
}

func NewTerminalUI() *TerminalUI {
    return &TerminalUI{
        formatter: formatters.Get("terminal256"),
        style:     styles.Get("dracula"),
    }
}

func (ui *TerminalUI) RenderCodeBlock(code, language string, lineNumbers bool) string {
    // Get lexer for language
    lexer := lexers.Get(language)
    if lexer == nil {
        lexer = lexers.Fallback
    }
    
    // Tokenize
    iterator, err := lexer.Tokenise(nil, code)
    if err != nil {
        return code
    }
    
    // Format with syntax highlighting
    var buf strings.Builder
    err = ui.formatter.Format(&buf, ui.style, iterator)
    if err != nil {
        return code
    }
    
    highlighted := buf.String()
    
    // Add line numbers if requested
    if lineNumbers {
        lines := strings.Split(highlighted, "\n")
        var numbered strings.Builder
        
        for i, line := range lines {
            lineNum := color.HiBlackString(fmt.Sprintf("%4d │ ", i+1))
            numbered.WriteString(lineNum + line + "\n")
        }
        
        highlighted = numbered.String()
    }
    
    // Add border
    width := 80
    top := color.HiBlackString("┌" + strings.Repeat("─", width-2) + "┐")
    bottom := color.HiBlackString("└" + strings.Repeat("─", width-2) + "┘")
    
    var result strings.Builder
    result.WriteString(top + "\n")
    for _, line := range strings.Split(highlighted, "\n") {
        padded := fmt.Sprintf("%-78s", line)
        result.WriteString(color.HiBlackString("│ ") + padded + color.HiBlackString(" │\n"))
    }
    result.WriteString(bottom)
    
    return result.String()
}

func (ui *TerminalUI) RenderDiff(oldCode, newCode string) string {
    // Generate unified diff
    diff := generateUnifiedDiff(oldCode, newCode)
    
    // Colorize
    var result strings.Builder
    for _, line := range strings.Split(diff, "\n") {
        switch {
        case strings.HasPrefix(line, "+"):
            result.WriteString(color.GreenString(line) + "\n")
        case strings.HasPrefix(line, "-"):
            result.WriteString(color.RedString(line) + "\n")
        case strings.HasPrefix(line, "@@"):
            result.WriteString(color.CyanString(line) + "\n")
        default:
            result.WriteString(color.HiBlackString(line) + "\n")
        }
    }
    
    return result.String()
}

func (ui *TerminalUI) RenderProgress(percent int, message string) string {
    width := 40
    filled := int(float64(width) * float64(percent) / 100.0)
    
    bar := color.GreenString(strings.Repeat("█", filled)) + 
           color.HiBlackString(strings.Repeat("░", width-filled))
    
    return fmt.Sprintf("\r[%s] %3d%% %s", bar, percent, message)
}
```

### Week 13-14: OpenHands & Cline Components

Create `internal/clis/openhands/sandbox.go`:

```go
package openhands

import (
    "context"
    "fmt"
    "time"
    
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

// Sandbox provides Docker-based secure execution
type Sandbox struct {
    client    *client.Client
    container string
    image     string
    
    // Resource limits
    memoryMB    int64
    cpuPercent  int64
    timeout     time.Duration
}

func NewSandbox(cli *client.Client, image string) *Sandbox {
    return &Sandbox{
        client:     cli,
        image:      image,
        memoryMB:   2048,  // 2GB default
        cpuPercent: 100000, // 1 core
        timeout:    30 * time.Second,
    }
}

func (s *Sandbox) Start(ctx context.Context, workspaceDir string) error {
    // Create container with security restrictions
    resp, err := s.client.ContainerCreate(ctx, 
        &container.Config{
            Image:        s.image,
            WorkingDir:   "/workspace",
            Cmd:          []string{"sleep", "infinity"},
            AttachStdout: true,
            AttachStderr: true,
        },
        &container.HostConfig{
            Binds: []string{fmt.Sprintf("%s:/workspace:rw", workspaceDir)},
            
            // Resource limits
            Memory:     s.memoryMB * 1024 * 1024,
            MemorySwap: s.memoryMB * 1024 * 1024, // No swap
            CPUQuota:   s.cpuPercent,
            
            // Security
            CapDrop:      []string{"ALL"},
            CapAdd:       []string{"CHOWN", "SETGID", "SETUID"},
            SecurityOpt:  []string{"no-new-privileges:true"},
            NetworkMode:  "none", // No network by default
            ReadonlyRootfs: true,
        },
        nil, nil, "",
    )
    if err != nil {
        return fmt.Errorf("create container: %w", err)
    }
    
    s.container = resp.ID
    
    // Start container
    if err := s.client.ContainerStart(ctx, s.container, types.ContainerStartOptions{}); err != nil {
        return fmt.Errorf("start container: %w", err)
    }
    
    return nil
}

func (s *Sandbox) Execute(ctx context.Context, command string) (*ExecutionResult, error) {
    ctx, cancel := context.WithTimeout(ctx, s.timeout)
    defer cancel()
    
    // Create exec
    execConfig := types.ExecConfig{
        Cmd:          []string{"sh", "-c", command},
        WorkingDir:   "/workspace",
        AttachStdout: true,
        AttachStderr: true,
    }
    
    execResp, err := s.client.ContainerExecCreate(ctx, s.container, execConfig)
    if err != nil {
        return nil, fmt.Errorf("create exec: %w", err)
    }
    
    // Attach and run
    resp, err := s.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
    if err != nil {
        return nil, fmt.Errorf("attach exec: %w", err)
    }
    defer resp.Close()
    
    // Read output
    output, err := io.ReadAll(resp.Reader)
    if err != nil {
        return nil, fmt.Errorf("read output: %w", err)
    }
    
    // Get exit code
    inspect, err := s.client.ContainerExecInspect(ctx, execResp.ID)
    if err != nil {
        return nil, fmt.Errorf("inspect exec: %w", err)
    }
    
    return &ExecutionResult{
        ExitCode: inspect.ExitCode,
        Output:   string(output),
        TimedOut: ctx.Err() == context.DeadlineExceeded,
    }, nil
}

func (s *Sandbox) Stop(ctx context.Context) error {
    if s.container == "" {
        return nil
    }
    
    // Stop container
    timeout := 10
    if err := s.client.ContainerStop(ctx, s.container, container.StopOptions{Timeout: &timeout}); err != nil {
        return err
    }
    
    // Remove container
    return s.client.ContainerRemove(ctx, s.container, types.ContainerRemoveOptions{
        Force: true,
    })
}
```

Create `internal/clis/cline/browser.go`:

```go
package cline

import (
    "context"
    "fmt"
    
    "github.com/playwright-community/playwright-go"
)

// BrowserController provides browser automation
type BrowserController struct {
    pw      *playwright.Playwright
    browser playwright.Browser
    page    playwright.Page
}

func NewBrowserController() (*BrowserController, error) {
    pw, err := playwright.Run()
    if err != nil {
        return nil, fmt.Errorf("start playwright: %w", err)
    }
    
    browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
        Headless: playwright.Bool(true),
    })
    if err != nil {
        return nil, fmt.Errorf("launch browser: %w", err)
    }
    
    page, err := browser.NewPage()
    if err != nil {
        return nil, fmt.Errorf("new page: %w", err)
    }
    
    return &BrowserController{
        pw:      pw,
        browser: browser,
        page:    page,
    }, nil
}

func (bc *BrowserController) Navigate(ctx context.Context, url string) error {
    _, err := bc.page.Goto(url, playwright.PageGotoOptions{
        WaitUntil: playwright.WaitUntilStateNetworkidle,
    })
    return err
}

func (bc *BrowserController) Screenshot(ctx context.Context) ([]byte, error) {
    return bc.page.Screenshot(playwright.PageScreenshotOptions{
        FullPage: playwright.Bool(true),
    })
}

func (bc *BrowserController) Click(ctx context.Context, selector string) error {
    return bc.page.Click(selector)
}

func (bc *BrowserController) Type(ctx context.Context, selector, text string) error {
    return bc.page.Fill(selector, text)
}

func (bc *BrowserController) GetContent(ctx context.Context) (string, error) {
    return bc.page.Content()
}

func (bc *BrowserController) Close() error {
    if err := bc.browser.Close(); err != nil {
        return err
    }
    return bc.pw.Stop()
}
```

### Week 15-16: Kiro & Continue Components

Create `internal/clis/kiro/memory.go`:

```go
package kiro

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/jackc/pgx/v5"
)

// ProjectMemory provides persistent project memory
type ProjectMemory struct {
    db        *pgx.Conn
    projectID string
    
    // Embedding generator
    embedder EmbeddingGenerator
    
    // Cache
    shortTerm map[string]*MemoryEntry
}

type MemoryEntry struct {
    Key        string
    Value      interface{}
    Importance float64     // 0.0-1.0
    Embedding  []float32   // For semantic search
    Timestamp  time.Time
}

func NewProjectMemory(db *pgx.Conn, projectID string, embedder EmbeddingGenerator) *ProjectMemory {
    return &ProjectMemory{
        db:        db,
        projectID: projectID,
        embedder:  embedder,
        shortTerm: make(map[string]*MemoryEntry),
    }
}

func (pm *ProjectMemory) Remember(ctx context.Context, key string, value interface{}, importance float64) error {
    entry := &MemoryEntry{
        Key:        key,
        Value:      value,
        Importance: importance,
        Timestamp:  time.Now(),
    }
    
    // Store in short-term memory
    pm.shortTerm[key] = entry
    
    // For important memories, persist to database
    if importance >= 0.3 {
        // Generate embedding for semantic search
        valueJSON, _ := json.Marshal(value)
        embedding, err := pm.embedder.Embed(ctx, string(valueJSON))
        if err == nil {
            entry.Embedding = embedding
        }
        
        // Persist to database
        _, err = pm.db.Exec(ctx, `
            INSERT INTO project_memory (project_id, key, value, importance, embedding)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (project_id, key) DO UPDATE
            SET value = $3, importance = $4, embedding = $5, updated_at = NOW()
        `, pm.projectID, key, valueJSON, importance, pgx.Array(embedding))
        
        if err != nil {
            return fmt.Errorf("persist memory: %w", err)
        }
    }
    
    return nil
}

func (pm *ProjectMemory) Recall(ctx context.Context, key string) (interface{}, error) {
    // Check short-term first
    if entry, ok := pm.shortTerm[key]; ok {
        return entry.Value, nil
    }
    
    // Query database
    var valueJSON []byte
    err := pm.db.QueryRow(ctx, `
        SELECT value FROM project_memory
        WHERE project_id = $1 AND key = $2
    `, pm.projectID, key).Scan(&valueJSON)
    
    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    var value interface{}
    if err := json.Unmarshal(valueJSON, &value); err != nil {
        return nil, err
    }
    
    return value, nil
}

func (pm *ProjectMemory) Search(ctx context.Context, query string, topK int) ([]*MemoryEntry, error) {
    // Generate query embedding
    queryEmb, err := pm.embedder.Embed(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Semantic search in database
    rows, err := pm.db.Query(ctx, `
        SELECT key, value, importance, 1 - (embedding <=> $3) as similarity
        FROM project_memory
        WHERE project_id = $1
        ORDER BY embedding <=> $3
        LIMIT $2
    `, pm.projectID, topK, pgx.Array(queryEmb))
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []*MemoryEntry
    for rows.Next() {
        var entry MemoryEntry
        var valueJSON []byte
        var similarity float64
        
        if err := rows.Scan(&entry.Key, &valueJSON, &entry.Importance, &similarity); err != nil {
            continue
        }
        
        if err := json.Unmarshal(valueJSON, &entry.Value); err != nil {
            continue
        }
        
        results = append(results, &entry)
    }
    
    return results, nil
}
```

---

## Phase 4: Output System (Weeks 17-20)

### Week 17-18: Formatting Pipeline

Create `internal/output/pipeline.go`:

```go
package output

import (
    "context"
    "io"
)

// Pipeline manages output formatting
type Pipeline struct {
    // Parsers
    parsers map[string]Parser
    
    // Formatters
    formatters map[string]Formatter
    
    // Renderers
    renderers map[string]Renderer
}

func NewPipeline() *Pipeline {
    p := &Pipeline{
        parsers:    make(map[string]Parser),
        formatters: make(map[string]Formatter),
        renderers:  make(map[string]Renderer),
    }
    
    // Register default parsers
    p.RegisterParser("code", &CodeParser{})
    p.RegisterParser("diff", &DiffParser{})
    p.RegisterParser("json", &JSONParser{})
    p.RegisterParser("markdown", &MarkdownParser{})
    
    // Register default formatters
    p.RegisterFormatter("syntax", &SyntaxHighlighter{})
    p.RegisterFormatter("diff", &DiffFormatter{})
    p.RegisterFormatter("table", &TableFormatter{})
    
    // Register default renderers
    p.RegisterRenderer("terminal", &TerminalRenderer{})
    p.RegisterRenderer("html", &HTMLRenderer{})
    p.RegisterRenderer("json", &JSONRenderer{})
    
    return p
}

func (p *Pipeline) Process(ctx context.Context, input *Input, opts *Options) error {
    // 1. Parse input
    parsed, err := p.parse(ctx, input)
    if err != nil {
        return err
    }
    
    // 2. Format content
    formatted, err := p.format(ctx, parsed, opts)
    if err != nil {
        return err
    }
    
    // 3. Render output
    return p.render(ctx, formatted, opts.Output, opts.Writer)
}

func (p *Pipeline) parse(ctx context.Context, input *Input) (*ParsedContent, error) {
    parser, ok := p.parsers[input.Type]
    if !ok {
        return nil, fmt.Errorf("unknown parser: %s", input.Type)
    }
    return parser.Parse(ctx, input.Data)
}

func (p *Pipeline) format(ctx context.Context, parsed *ParsedContent, opts *Options) (*FormattedContent, error) {
    formatter, ok := p.formatters[opts.Format]
    if !ok {
        return nil, fmt.Errorf("unknown formatter: %s", opts.Format)
    }
    return formatter.Format(ctx, parsed, opts.FormatOptions)
}

func (p *Pipeline) render(ctx context.Context, formatted *FormattedContent, outputType string, w io.Writer) error {
    renderer, ok := p.renderers[outputType]
    if !ok {
        return fmt.Errorf("unknown renderer: %s", outputType)
    }
    return renderer.Render(ctx, formatted, w)
}
```

### Week 19-20: Terminal Enhancements

Create `internal/output/terminal/enhanced.go`:

```go
package terminal

import (
    "fmt"
    "strings"
    
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// EnhancedTerminal provides rich interactive terminal UI
type EnhancedTerminal struct {
    program *tea.Program
    model   *TerminalModel
}

type TerminalModel struct {
    // Content
    messages []Message
    
    // Progress
    progress   progress.Model
    spinner    spinner.Model
    showProgress bool
    
    // Styles
    styles Styles
}

type Styles struct {
    User      lipgloss.Style
    Assistant lipgloss.Style
    System    lipgloss.Style
    Code      lipgloss.Style
    DiffAdd   lipgloss.Style
    DiffDel   lipgloss.Style
}

func NewEnhancedTerminal() *EnhancedTerminal {
    model := &TerminalModel{
        messages: make([]Message, 0),
        progress: progress.New(progress.WithDefaultGradient()),
        spinner:  spinner.New(spinner.WithSpinner(spinner.Dot)),
        styles: Styles{
            User:      lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")),
            Assistant: lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
            System:    lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
            Code:      lipgloss.NewStyle().Background(lipgloss.Color("#1a1a1a")),
            DiffAdd:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
            DiffDel:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
        },
    }
    
    return &EnhancedTerminal{
        program: tea.NewProgram(model),
        model:   model,
    }
}

func (et *EnhancedTerminal) Run() error {
    _, err := et.program.Run()
    return err
}

func (et *EnhancedTerminal) AddMessage(role, content string) {
    et.model.messages = append(et.model.messages, Message{
        Role:    role,
        Content: content,
    })
}

func (et *EnhancedTerminal) SetProgress(percent float64) {
    et.model.progress.SetPercent(percent)
    et.model.showProgress = true
}

func (m *TerminalModel) View() string {
    var b strings.Builder
    
    // Render messages
    for _, msg := range m.messages {
        switch msg.Role {
        case "user":
            b.WriteString(m.styles.User.Render("You: "))
            b.WriteString(msg.Content)
        case "assistant":
            b.WriteString(m.styles.Assistant.Render("Assistant: "))
            b.WriteString(m.renderContent(msg.Content))
        case "system":
            b.WriteString(m.styles.System.Render(msg.Content))
        }
        b.WriteString("\n\n")
    }
    
    // Render progress if active
    if m.showProgress {
        b.WriteString(m.progress.View())
        b.WriteString("\n")
    }
    
    return b.String()
}

func (m *TerminalModel) renderContent(content string) string {
    // Parse and render markdown, code blocks, etc.
    // Implementation...
    return content
}
```

---

## Phase 5: Testing & Deployment (Weeks 21-24)

### Week 21-22: Comprehensive Testing

Create test plans for all ported components:

```go
// internal/clis/aider/repo_map_test.go

func TestRepoMap_GetRankedTags(t *testing.T) {
    rm := NewRepoMap("testdata/repo", 1024)
    
    ctx := context.Background()
    result, err := rm.GetRankedTags(ctx, "find user authentication", nil)
    
    require.NoError(t, err)
    assert.NotEmpty(t, result.Symbols)
    assert.True(t, len(result.Symbols) > 0)
    
    // Check that auth-related symbols are ranked higher
    for i, sym := range result.Symbols {
        if i < 5 {
            assert.True(t, strings.Contains(strings.ToLower(sym.Name), "auth") ||
                strings.Contains(strings.ToLower(sym.File), "auth"))
        }
    }
}

func TestRepoMap_ExtractSymbols(t *testing.T) {
    rm := NewRepoMap(".", 1024)
    
    symbols, err := rm.extractSymbols(context.Background(), "test.go")
    require.NoError(t, err)
    
    // Should find Go functions and types
    var foundFunc, foundType bool
    for _, sym := range symbols {
        if sym.Type == "function" {
            foundFunc = true
        }
        if sym.Type == "type" {
            foundType = true
        }
    }
    
    assert.True(t, foundFunc, "should find functions")
    assert.True(t, foundType, "should find types")
}
```

### Week 23-24: Deployment & Documentation

Create deployment checklist:

```markdown
# Deployment Checklist

## Pre-Deployment
- [ ] All 20 critical features implemented
- [ ] Unit tests passing (>80% coverage)
- [ ] Integration tests passing
- [ ] Performance benchmarks met
- [ ] Security audit completed
- [ ] Documentation updated

## Database Migration
- [ ] Run `sql/001_cli_agents_fusion.sql`
- [ ] Verify tables created
- [ ] Verify indexes created
- [ ] Verify feature registry populated

## Configuration
- [ ] Update `config.yaml` with new settings
- [ ] Configure instance pools
- [ ] Set resource limits
- [ ] Configure caching

## Deployment
- [ ] Deploy to staging
- [ ] Run smoke tests
- [ ] Deploy to production
- [ ] Monitor metrics
- [ ] Verify alerts

## Post-Deployment
- [ ] Verify all agents functioning
- [ ] Check ensemble coordination
- [ ] Monitor error rates
- [ ] Collect performance metrics
- [ ] Gather user feedback
```

---

## Risk Mitigation

### High Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Tree-sitter integration complexity | High | Start with simpler AST parsers, iterate |
| Docker sandboxing performance | Medium | Use pool pattern, pre-warm containers |
| Browser automation dependencies | High | Make optional, lazy load |

### Medium Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Memory usage with multiple instances | Medium | Implement strict limits, auto-scaling |
| Database migration complexity | Medium | Use migrations framework, rollback plan |
| Backward compatibility | Medium | Maintain legacy API endpoints |

---

## Success Metrics

### Technical Metrics
- **Feature Coverage:** 100% of 20 critical features ported
- **Test Coverage:** >80% for new code
- **Performance:** <500ms p99 latency
- **Reliability:** >99.9% uptime

### User Metrics
- **Adoption:** 80% of users migrate to new features
- **Satisfaction:** >4.5/5 user rating
- **Error Rate:** <0.1% error rate

---

## Next Steps

1. Begin Phase 1: Foundation Layer
2. Set up development environment
3. Create feature branches
4. Implement in order of priority
5. Test continuously
6. Deploy incrementally

---

*Final Implementation Plan Complete*  
*Total Duration: 24 weeks*  
*Date: 2026-04-03*  
*HelixAgent Commit: 8a976be2*
