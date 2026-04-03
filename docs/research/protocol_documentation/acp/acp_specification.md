# Agent Communication Protocol (ACP) - Complete Specification

**Protocol:** ACP (Agent Communication Protocol)  
**Version:** 1.0-draft  
**Status:** Draft  
**HelixAgent Implementation:** [internal/acp/](../../../internal/acp/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

ACP is a protocol for coordinating multiple AI agents in a distributed system. It enables agents to discover each other, exchange messages, form consensus, and coordinate actions.

**Key Use Cases:**
- Multi-agent debate orchestration
- Distributed task execution
- Consensus building
- Agent swarm coordination

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     ACP ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│   │   Agent 1    │  │   Agent 2    │  │   Agent N    │         │
│   │  (Claude)    │  │   (GPT-4)    │  │  (DeepSeek)  │         │
│   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│          │                 │                 │                  │
│          └─────────────────┼─────────────────┘                  │
│                            │                                    │
│                    ┌───────┴───────┐                           │
│                    │  ACP Broker   │                           │
│                    │  (HelixAgent) │                           │
│                    ├───────────────┤                           │
│                    │ • Registry    │                           │
│                    │ • Routing     │                           │
│                    │ • Consensus   │                           │
│                    │ • Audit       │                           │
│                    └───────────────┘                           │
│                                                                  │
│   Transport: HTTP, WebSocket, or Message Queue                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Protocol Basics

### Message Format

ACP uses JSON-RPC 2.0 with custom extensions:

```json
{
  "jsonrpc": "2.0",
  "id": "msg-unique-id",
  "method": "agent/message",
  "params": {
    "to": "agent-id",
    "from": "agent-id",
    "message": {
      "type": "argument|response|consensus|action",
      "content": {},
      "metadata": {}
    }
  }
}
```

### Core Types

```typescript
// Agent Identity
interface AgentIdentity {
  id: string;
  name: string;
  provider: string;  // claude, gpt4, etc.
  capabilities: string[];
  metadata: object;
}

// ACP Message
interface ACPMessage {
  id: string;
  to: string;
  from: string;
  type: MessageType;
  content: unknown;
  timestamp: string;
  threadId?: string;
  replyTo?: string;
}

type MessageType = 
  | "argument"      // Debate argument
  | "response"      // Response to message
  | "consensus"     // Consensus proposal
  | "action"        // Action request
  | "status"        // Status update
  | "error";        // Error notification

// Consensus
interface Consensus {
  topic: string;
  proposals: Proposal[];
  votes: Vote[];
  threshold: number;
  deadline: string;
}

interface Proposal {
  id: string;
  agentId: string;
  content: unknown;
  timestamp: string;
}

interface Vote {
  proposalId: string;
  agentId: string;
  vote: "for" | "against" | "abstain";
  confidence: number;
  reasoning: string;
}
```

---

## Core Operations

### 1. Agent Registration

```json
// Register agent with broker
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "agent/register",
  "params": {
    "agent": {
      "id": "claude-debater-1",
      "name": "Claude Debater",
      "provider": "claude-3-5-sonnet",
      "capabilities": ["reasoning", "code_review", "architecture"]
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "registered": true,
    "sessionToken": "eyJhbGc...",
    "expiresAt": "2026-04-03T12:00:00Z"
  }
}
```

### 2. Agent Discovery

```json
// List available agents
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "agent/list",
  "params": {
    "capabilities": ["code_review"],
    "status": "available"
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "agents": [
      {
        "id": "claude-debater-1",
        "name": "Claude Debater",
        "provider": "claude-3-5-sonnet",
        "capabilities": ["reasoning", "code_review"],
        "status": "available"
      },
      {
        "id": "gpt4-reviewer-1",
        "name": "GPT-4 Reviewer",
        "provider": "gpt-4",
        "capabilities": ["code_review", "testing"],
        "status": "available"
      }
    ]
  }
}
```

### 3. Send Message

```json
// Send message to agent
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "agent/send",
  "params": {
    "to": "gpt4-reviewer-1",
    "message": {
      "type": "argument",
      "content": {
        "topic": "API Design",
        "position": "We should use REST over GraphQL",
        "reasoning": [
          "Better caching support",
          "Simpler client implementation",
          "Wider tooling support"
        ]
      },
      "metadata": {
        "round": 1,
        "priority": "high"
      }
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "messageId": "msg-abc123",
    "delivered": true,
    "timestamp": "2026-04-03T10:30:00Z"
  }
}
```

### 4. Broadcast Message

```json
// Broadcast to multiple agents
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "agent/broadcast",
  "params": {
    "to": ["claude-debater-1", "gpt4-reviewer-1", "deepseek-analyst-1"],
    "message": {
      "type": "consensus",
      "content": {
        "topic": "Architecture Decision",
        "proposal": "Adopt microservices architecture",
        "deadline": "2026-04-03T11:00:00Z"
      }
    }
  }
}
```

---

## Debate Coordination

### 1. Create Debate

```json
// Initialize debate session
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "debate/create",
  "params": {
    "topic": "Should we use PostgreSQL or MongoDB?",
    "participants": ["claude-debater-1", "gpt4-reviewer-1"],
    "rules": {
      "maxRounds": 3,
      "timeLimit": 300,
      "consensusThreshold": 0.7
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "debateId": "debate-xyz789",
    "status": "active",
    "round": 0,
    "participants": ["claude-debater-1", "gpt4-reviewer-1"]
  }
}
```

### 2. Submit Argument

```json
// Submit argument in debate
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "debate/argument",
  "params": {
    "debateId": "debate-xyz789",
    "argument": {
      "position": "pro",
      "content": "PostgreSQL provides ACID guarantees essential for financial data...",
      "evidence": [
        {
          "type": "documentation",
          "source": "PostgreSQL docs",
          "url": "https://postgresql.org/docs/..."
        }
      ]
    }
  }
}
```

### 3. Vote on Consensus

```json
// Vote on debate conclusion
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "debate/vote",
  "params": {
    "debateId": "debate-xyz789",
    "vote": {
      "proposal": "Use PostgreSQL with read replicas",
      "decision": "for",
      "confidence": 0.85,
      "reasoning": "ACID compliance outweighs scalability concerns for our use case"
    }
  }
}
```

### 4. Get Debate Results

```json
// Get final debate results
{
  "jsonrpc": "2.0",
  "id": 8,
  "method": "debate/results",
  "params": {
    "debateId": "debate-xyz789"
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 8,
  "result": {
    "debateId": "debate-xyz789",
    "status": "concluded",
    "winner": "Use PostgreSQL with read replicas",
    "consensus": 0.85,
    "votes": [
      {
        "agentId": "claude-debater-1",
        "decision": "for",
        "confidence": 0.90
      },
      {
        "agentId": "gpt4-reviewer-1",
        "decision": "for",
        "confidence": 0.80
      }
    ],
    "arguments": [
      {
        "round": 1,
        "agentId": "claude-debater-1",
        "content": "..."
      }
    ]
  }
}
```

---

## Consensus Building

### Consensus Algorithm

```typescript
interface ConsensusAlgorithm {
  // Weighted voting based on agent expertise
  calculateConsensus(votes: Vote[], weights: Map<string, number>): ConsensusResult;
  
  // Require supermajority for critical decisions
  threshold: number; // e.g., 0.67 for 2/3 majority
  
  // Time-limited consensus window
  deadline: Date;
}

interface ConsensusResult {
  reached: boolean;
  winningProposal: string;
  supportPercentage: number;
  confidence: number;
  dissentingOpinions: string[];
}
```

### Consensus Protocol Flow

```
1. PROPOSE
   Agent 1 → Broker: "Proposal X"
   
2. BROADCAST
   Broker → All Agents: "Vote on Proposal X"
   
3. VOTE
   Agent 1 → Broker: "FOR (confidence: 0.9)"
   Agent 2 → Broker: "FOR (confidence: 0.8)"
   Agent 3 → Broker: "AGAINST (confidence: 0.7)"
   
4. CALCULATE
   Broker: Calculate weighted consensus
   
5. ANNOUNCE
   Broker → All Agents: "Consensus reached: 85% support"
```

---

## HelixAgent ACP Implementation

### Architecture

**Source:** [`internal/acp/`](../../../internal/acp/)

```
internal/acp/
├── acp.go                    # Core ACP protocol
├── broker.go                 # Message broker
├── registry.go               # Agent registry
├── routing.go                # Message routing
├── consensus.go              # Consensus engine
├── debate.go                 # Debate orchestration
├── transport/
│   ├── http.go              # HTTP transport
│   ├── websocket.go         # WebSocket transport
│   └── queue.go             # Message queue
└── storage/
    ├── messages.go          # Message persistence
    └── debates.go           # Debate storage
```

### Broker Implementation

**Source:** [`internal/acp/broker.go`](../../../internal/acp/broker.go)

```go
package acp

// ACPBroker manages agent communication
// Source: internal/acp/broker.go#L1-180

type ACPBroker struct {
    registry    *AgentRegistry
    router      *MessageRouter
    consensus   *ConsensusEngine
    debate      *DebateOrchestrator
    transports  []Transport
    storage     MessageStorage
}

// RegisterAgent adds agent to broker
// Source: internal/acp/broker.go#L45-78
func (b *ACPBroker) RegisterAgent(ctx context.Context, agent *AgentIdentity) (*Session, error) {
    // Validate agent credentials
    if err := b.validateAgent(agent); err != nil {
        return nil, err
    }
    
    // Register in registry
    if err := b.registry.Register(ctx, agent); err != nil {
        return nil, err
    }
    
    // Create session
    session := &Session{
        AgentID:   agent.ID,
        Token:     generateToken(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    return session, nil
}

// RouteMessage delivers message to target agent
// Source: internal/acp/broker.go#L80-125
func (b *ACPBroker) RouteMessage(ctx context.Context, msg *ACPMessage) error {
    // Validate message
    if err := b.validateMessage(msg); err != nil {
        return err
    }
    
    // Get target agent transport
    agent, err := b.registry.Get(ctx, msg.To)
    if err != nil {
        return err
    }
    
    // Route through appropriate transport
    transport := b.selectTransport(agent)
    if err := transport.Send(ctx, msg); err != nil {
        return err
    }
    
    // Persist message
    return b.storage.SaveMessage(ctx, msg)
}

// CoordinateDebate manages multi-agent debate
// Source: internal/acp/broker.go#L127-180
func (b *ACPBroker) CoordinateDebate(ctx context.Context, config *DebateConfig) (*DebateSession, error) {
    // Initialize debate orchestrator
    debate, err := b.debate.Create(ctx, config)
    if err != nil {
        return nil, err
    }
    
    // Notify participants
    for _, agentID := range config.Participants {
        msg := &ACPMessage{
            To:      agentID,
            Type:    "status",
            Content: map[string]string{"event": "debate_invite", "debateId": debate.ID},
        }
        if err := b.RouteMessage(ctx, msg); err != nil {
            log.Printf("Failed to notify %s: %v", agentID, err)
        }
    }
    
    return debate, nil
}
```

### Debate Orchestrator

**Source:** [`internal/acp/debate.go`](../../../internal/acp/debate.go)

```go
package acp

// DebateOrchestrator manages agent debates
// Source: internal/acp/debate.go#L1-200

type DebateOrchestrator struct {
    debates    map[string]*DebateSession
    consensus  *ConsensusEngine
    mu         sync.RWMutex
}

// Create initializes new debate
// Source: internal/acp/debate.go#L25-67
func (d *DebateOrchestrator) Create(ctx context.Context, config *DebateConfig) (*DebateSession, error) {
    debate := &DebateSession{
        ID:           generateID(),
        Topic:        config.Topic,
        Participants: config.Participants,
        Rules:        config.Rules,
        Status:       "active",
        Round:        0,
        Arguments:    []Argument{},
        Votes:        []Vote{},
        StartedAt:    time.Now(),
    }
    
    d.mu.Lock()
    d.debates[debate.ID] = debate
    d.mu.Unlock()
    
    // Start debate timer
    go d.manageDebateLifecycle(debate)
    
    return debate, nil
}

// SubmitArgument records agent argument
// Source: internal/acp/debate.go#L69-110
func (d *DebateOrchestrator) SubmitArgument(ctx context.Context, debateID string, agentID string, argument *Argument) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    debate, ok := d.debates[debateID]
    if !ok {
        return ErrDebateNotFound
    }
    
    // Validate agent is participant
    if !contains(debate.Participants, agentID) {
        return ErrNotParticipant
    }
    
    // Add argument
    argument.AgentID = agentID
    argument.Timestamp = time.Now()
    argument.Round = debate.Round
    debate.Arguments = append(debate.Arguments, *argument)
    
    // Check if round complete
    if d.isRoundComplete(debate) {
        debate.Round++
        d.notifyRoundComplete(debate)
    }
    
    return nil
}

// CalculateConsensus determines debate winner
// Source: internal/acp/debate.go#L145-200
func (d *DebateOrchestrator) CalculateConsensus(debate *DebateSession) (*ConsensusResult, error) {
    if len(debate.Votes) == 0 {
        return nil, ErrNoVotes
    }
    
    // Group votes by proposal
    voteCounts := make(map[string][]Vote)
    for _, vote := range debate.Votes {
        voteCounts[vote.Proposal] = append(voteCounts[vote.Proposal], vote)
    }
    
    // Find proposal with highest support
    var winner string
    var maxSupport float64
    
    for proposal, votes := range voteCounts {
        support := d.calculateSupport(votes)
        if support > maxSupport {
            maxSupport = support
            winner = proposal
        }
    }
    
    // Check threshold
    if maxSupport < debate.Rules.ConsensusThreshold {
        return &ConsensusResult{
            Reached:    false,
            Confidence: maxSupport,
        }, nil
    }
    
    return &ConsensusResult{
        Reached:            true,
        WinningProposal:    winner,
        SupportPercentage:  maxSupport,
        Confidence:         d.calculateConfidence(voteCounts[winner]),
    }, nil
}
```

---

## CLI Agent ACP Integration

### Which Agents Support ACP?

| Agent | ACP Support | Native Protocol | HelixAgent Bridge |
|-------|-------------|-----------------|-------------------|
| Claude Code | ❌ | None | ✅ Via adapter |
| Aider | ❌ | None | ✅ Via adapter |
| Codex | ❌ | None | ✅ Via adapter |
| Cline | ❌ | VS Code API | ⚠️ Partial |
| OpenHands | ⚠️ | Custom | ✅ Adapter |
| Kiro | ✅ | Native | ✅ Full |
| Forge | ✅ | Native | ✅ Full |
| Multiagent Coding | ✅ | Custom | ✅ Adapter |
| Claude Squad | ✅ | Native | ✅ Full |

### ACP vs Other Coordination Methods

| Method | Protocol | Scalability | Latency | Use Case |
|--------|----------|-------------|---------|----------|
| **ACP** | JSON-RPC | High | Medium | General coordination |
| Custom HTTP | REST | Medium | Medium | Simple coordination |
| Message Queue | AMQP | Very High | Higher | Event-driven |
| gRPC | gRPC | High | Low | Service mesh |
| Direct Socket | TCP | Low | Very Low | Real-time games |

---

## Source Code Reference

### ACP Core Files

| Component | Source File | Lines | Description |
|-----------|-------------|-------|-------------|
| Protocol | `internal/acp/acp.go` | 150 | Core types |
| Broker | `internal/acp/broker.go` | 180 | Message broker |
| Registry | `internal/acp/registry.go` | 120 | Agent registry |
| Routing | `internal/acp/routing.go` | 140 | Message routing |
| Consensus | `internal/acp/consensus.go` | 200 | Consensus engine |
| Debate | `internal/acp/debate.go` | 200 | Debate orchestrator |
| HTTP Transport | `internal/acp/transport/http.go` | 110 | HTTP handler |
| WebSocket | `internal/acp/transport/websocket.go` | 130 | WS handler |
| Storage | `internal/acp/storage/messages.go` | 90 | Message storage |
| Tests | `internal/acp/acp_test.go` | 280 | Unit tests |

---

## API Endpoints

### REST Endpoints

```
POST /v1/acp/agent/register      # Register agent
GET  /v1/acp/agent/list          # List agents
POST /v1/acp/agent/send          # Send message
POST /v1/acp/agent/broadcast     # Broadcast message

POST /v1/acp/debate/create       # Create debate
POST /v1/acp/debate/argument     # Submit argument
POST /v1/acp/debate/vote         # Vote on consensus
GET  /v1/acp/debate/{id}         # Get debate status
GET  /v1/acp/debate/{id}/results # Get results
```

### WebSocket Endpoints

```
ws://localhost:7061/v1/acp/stream

// Connect with agent token
{
  "type": "connect",
  "token": "eyJhbGc..."
}

// Receive messages
{
  "type": "message",
  "message": {
    "from": "agent-1",
    "type": "argument",
    "content": {...}
  }
}
```

---

## Conclusion

ACP is the **protocol for multi-agent coordination**. HelixAgent provides:

- ✅ Full ACP 1.0 implementation
- ✅ Debate orchestration with consensus
- ✅ Agent registry and discovery
- ✅ Multiple transport options
- ✅ Integration with all major LLM providers

**Recommendation:** Use ACP when coordinating multiple AI agents for complex decisions.

---

*Specification Version: ACP 1.0-draft*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: aa960946*
