// Package topology provides Chain topology implementation.
// Chain topology provides deterministic sequential communication.
// Slow but predictable - useful for tasks requiring strict ordering.
package topology

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ChainTopology implements a linear chain topology.
// Agents communicate in a sequential order, each passing to the next.
type ChainTopology struct {
	*BaseTopology

	// Ordered chain of agents
	chain   []*Agent
	chainMu sync.RWMutex

	// Current position in chain
	currentPos int
	posMu      sync.RWMutex

	// Message queue
	messageQueue chan *Message
}

// NewChainTopology creates a new Chain topology.
func NewChainTopology(config TopologyConfig) *ChainTopology {
	config.Type = TopologyChain

	ct := &ChainTopology{
		BaseTopology: NewBaseTopology(config),
		chain:        make([]*Agent, 0),
		messageQueue: make(chan *Message, 100),
	}

	return ct
}

// GetType returns the topology type.
func (ct *ChainTopology) GetType() TopologyType {
	return TopologyChain
}

// Initialize sets up the Chain topology with agents.
func (ct *ChainTopology) Initialize(ctx context.Context, agents []*Agent) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if len(agents) == 0 {
		return fmt.Errorf("at least one agent required for Chain topology")
	}

	// Add all agents
	for _, agent := range agents {
		ct.agents[agent.ID] = agent
		ct.agentsByRole[agent.Role] = append(ct.agentsByRole[agent.Role], agent)
	}

	// Build chain order based on roles and scores
	ct.buildChain()

	// Create chain channels
	ct.chainMu.RLock()
	for i := 0; i < len(ct.chain)-1; i++ {
		ct.channels = append(ct.channels, CommunicationChannel{
			FromAgent:     ct.chain[i].ID,
			ToAgent:       ct.chain[i+1].ID,
			Bidirectional: false, // Chain is unidirectional
			Weight:        1.0,
		})
	}
	// Close the loop for continuous processing
	if len(ct.chain) > 1 {
		ct.channels = append(ct.channels, CommunicationChannel{
			FromAgent:     ct.chain[len(ct.chain)-1].ID,
			ToAgent:       ct.chain[0].ID,
			Bidirectional: false,
			Weight:        0.5, // Lower weight for loop-back
		})
	}
	ct.chainMu.RUnlock()

	// Start message processor
	go ct.processMessages(ctx)

	return nil
}

// buildChain creates the ordered chain of agents.
// Order: Proposer -> Architect -> Critic -> Reviewer -> Optimizer -> TestAgent -> Validator -> Moderator
func (ct *ChainTopology) buildChain() {
	ct.chainMu.Lock()
	defer ct.chainMu.Unlock()

	// Define role order
	roleOrder := []AgentRole{
		RoleProposer,
		RoleArchitect,
		RoleCritic,
		RoleRedTeam,
		RoleReviewer,
		RoleBlueTeam,
		RoleOptimizer,
		RoleTestAgent,
		RoleSecurity,
		RoleValidator,
		RoleTeacher,
		RoleModerator,
	}

	ct.chain = make([]*Agent, 0, len(ct.agents))
	addedAgents := make(map[string]bool)

	// Add agents in role order
	for _, role := range roleOrder {
		roleAgents := ct.agentsByRole[role]
		// Sort by score within role
		sort.Slice(roleAgents, func(i, j int) bool {
			return roleAgents[i].Score > roleAgents[j].Score
		})
		for _, agent := range roleAgents {
			if !addedAgents[agent.ID] {
				ct.chain = append(ct.chain, agent)
				addedAgents[agent.ID] = true
			}
		}
	}

	// Add any remaining agents not in roleOrder
	for _, agent := range ct.agents {
		if !addedAgents[agent.ID] {
			ct.chain = append(ct.chain, agent)
		}
	}
}

// GetChain returns the ordered chain of agents.
func (ct *ChainTopology) GetChain() []*Agent {
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	chain := make([]*Agent, len(ct.chain))
	copy(chain, ct.chain)
	return chain
}

// GetChainPosition returns the position of an agent in the chain.
func (ct *ChainTopology) GetChainPosition(agentID string) int {
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	for i, agent := range ct.chain {
		if agent.ID == agentID {
			return i
		}
	}
	return -1
}

// GetCurrentPosition returns the current position in the chain.
func (ct *ChainTopology) GetCurrentPosition() int {
	ct.posMu.RLock()
	defer ct.posMu.RUnlock()
	return ct.currentPos
}

// AdvancePosition moves to the next position in the chain.
func (ct *ChainTopology) AdvancePosition() *Agent {
	ct.posMu.Lock()
	defer ct.posMu.Unlock()

	ct.chainMu.RLock()
	chainLen := len(ct.chain)
	ct.chainMu.RUnlock()

	if chainLen == 0 {
		return nil
	}

	ct.currentPos = (ct.currentPos + 1) % chainLen

	ct.chainMu.RLock()
	agent := ct.chain[ct.currentPos]
	ct.chainMu.RUnlock()

	return agent
}

// GetCurrentAgent returns the agent at the current chain position.
func (ct *ChainTopology) GetCurrentAgent() *Agent {
	ct.posMu.RLock()
	pos := ct.currentPos
	ct.posMu.RUnlock()

	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	if pos >= 0 && pos < len(ct.chain) {
		return ct.chain[pos]
	}
	return nil
}

// GetNextAgent returns the next agent in the chain.
func (ct *ChainTopology) GetNextAgent(agentID string) *Agent {
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	for i, agent := range ct.chain {
		if agent.ID == agentID {
			nextIdx := (i + 1) % len(ct.chain)
			return ct.chain[nextIdx]
		}
	}
	return nil
}

// GetPreviousAgent returns the previous agent in the chain.
func (ct *ChainTopology) GetPreviousAgent(agentID string) *Agent {
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	for i, agent := range ct.chain {
		if agent.ID == agentID {
			prevIdx := i - 1
			if prevIdx < 0 {
				prevIdx = len(ct.chain) - 1
			}
			return ct.chain[prevIdx]
		}
	}
	return nil
}

// CanCommunicate checks if two agents can communicate.
// In Chain topology, agents can only communicate with adjacent agents.
func (ct *ChainTopology) CanCommunicate(fromID, toID string) bool {
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	fromPos := -1
	toPos := -1

	for i, agent := range ct.chain {
		if agent.ID == fromID {
			fromPos = i
		}
		if agent.ID == toID {
			toPos = i
		}
	}

	if fromPos == -1 || toPos == -1 {
		return false
	}

	// Can communicate with next in chain
	nextPos := (fromPos + 1) % len(ct.chain)
	return toPos == nextPos
}

// GetCommunicationTargets returns agents that can receive messages.
// In Chain topology, only the next agent in the chain.
func (ct *ChainTopology) GetCommunicationTargets(agentID string) []*Agent {
	next := ct.GetNextAgent(agentID)
	if next != nil {
		return []*Agent{next}
	}
	return []*Agent{}
}

// RouteMessage determines routing for a message.
// In Chain topology, messages always go to the next agent.
func (ct *ChainTopology) RouteMessage(msg *Message) ([]*Agent, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	_, senderExists := ct.agents[msg.FromAgent]
	if !senderExists {
		return nil, fmt.Errorf("%w: sender %s", ErrAgentNotFound, msg.FromAgent)
	}

	// Always route to next in chain
	next := ct.GetNextAgent(msg.FromAgent)
	if next == nil {
		return nil, fmt.Errorf("no next agent in chain")
	}

	return []*Agent{next}, nil
}

// BroadcastMessage in Chain topology sends to all agents sequentially.
func (ct *ChainTopology) BroadcastMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	ct.IncrementMessageCount(true)

	// Get chain starting from sender
	ct.chainMu.RLock()
	chain := make([]*Agent, len(ct.chain))
	copy(chain, ct.chain)
	ct.chainMu.RUnlock()

	// Find sender position
	startPos := 0
	for i, agent := range chain {
		if agent.ID == msg.FromAgent {
			startPos = i
			break
		}
	}

	// Send to each agent in order
	for i := 1; i < len(chain); i++ {
		idx := (startPos + i) % len(chain)
		target := chain[idx]

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ct.messageQueue <- msg:
			target.UpdateActivity(0)
		case <-time.After(ct.config.MessageTimeout):
			return fmt.Errorf("message timeout to %s", target.ID)
		}
	}

	return nil
}

// SendMessage sends a message to the next agent in the chain.
func (ct *ChainTopology) SendMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	ct.IncrementMessageCount(false)

	targets, err := ct.RouteMessage(msg)
	if err != nil {
		return err
	}

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ct.messageQueue <- msg:
			target.UpdateActivity(0)
		case <-time.After(ct.config.MessageTimeout):
			return fmt.Errorf("message timeout to %s", target.ID)
		}
	}

	return nil
}

// processMessages processes messages sequentially.
func (ct *ChainTopology) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ct.messageQueue:
			if !ok {
				return
			}
			// Sequential processing - advance chain position
			ct.AdvancePosition()
			_ = msg // Process message (handlers would go here)
		}
	}
}

// SelectLeader selects the leader based on current chain position.
// In Chain topology, the current agent is typically the leader.
func (ct *ChainTopology) SelectLeader(phase DebatePhase) (*Agent, error) {
	// Select based on phase
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	var preferredRoles []AgentRole
	switch phase {
	case PhaseProposal:
		preferredRoles = []AgentRole{RoleProposer, RoleArchitect}
	case PhaseCritique:
		preferredRoles = []AgentRole{RoleCritic, RoleRedTeam}
	case PhaseReview:
		preferredRoles = []AgentRole{RoleReviewer}
	case PhaseOptimization:
		preferredRoles = []AgentRole{RoleOptimizer}
	case PhaseConvergence:
		preferredRoles = []AgentRole{RoleModerator, RoleValidator}
	}

	// Find first agent in chain with preferred role
	for _, agent := range ct.chain {
		for _, role := range preferredRoles {
			if agent.Role == role {
				return agent, nil
			}
		}
	}

	// Fall back to first agent in chain
	if len(ct.chain) > 0 {
		return ct.chain[0], nil
	}

	return nil, fmt.Errorf("no agents in chain")
}

// GetParallelGroups returns groups for parallel execution.
// In Chain topology, there's no parallelism - each agent is its own group.
func (ct *ChainTopology) GetParallelGroups(phase DebatePhase) [][]*Agent {
	ct.chainMu.RLock()
	defer ct.chainMu.RUnlock()

	// Each agent is its own sequential group
	groups := make([][]*Agent, len(ct.chain))
	for i, agent := range ct.chain {
		groups[i] = []*Agent{agent}
	}

	return groups
}

// ReorderChain reorders the chain based on new criteria.
func (ct *ChainTopology) ReorderChain(agentIDs []string) error {
	ct.chainMu.Lock()
	defer ct.chainMu.Unlock()

	if len(agentIDs) != len(ct.chain) {
		return fmt.Errorf("agent count mismatch: expected %d, got %d", len(ct.chain), len(agentIDs))
	}

	newChain := make([]*Agent, len(agentIDs))
	for i, id := range agentIDs {
		agent, ok := ct.agents[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrAgentNotFound, id)
		}
		newChain[i] = agent
	}

	ct.chain = newChain

	// Rebuild channels
	ct.channels = ct.channels[:0]
	for i := 0; i < len(ct.chain)-1; i++ {
		ct.channels = append(ct.channels, CommunicationChannel{
			FromAgent:     ct.chain[i].ID,
			ToAgent:       ct.chain[i+1].ID,
			Bidirectional: false,
			Weight:        1.0,
		})
	}
	// Loop back
	if len(ct.chain) > 1 {
		ct.channels = append(ct.channels, CommunicationChannel{
			FromAgent:     ct.chain[len(ct.chain)-1].ID,
			ToAgent:       ct.chain[0].ID,
			Bidirectional: false,
			Weight:        0.5,
		})
	}

	return nil
}

// Close closes the Chain topology.
func (ct *ChainTopology) Close() error {
	close(ct.messageQueue)
	return ct.BaseTopology.Close()
}
