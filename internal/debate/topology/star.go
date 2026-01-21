// Package topology provides Star topology implementation.
// Star topology has a central moderator through which all communication flows.
// Simple but creates a bottleneck - use as fallback when Graph-Mesh is too complex.
package topology

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StarTopology implements a star/hub-and-spoke topology.
// All communication goes through a central moderator.
type StarTopology struct {
	*BaseTopology

	// Central moderator
	moderator   *Agent
	moderatorMu sync.RWMutex

	// Message queue through moderator
	incomingQueue chan *Message
	outgoingQueue chan *Message
}

// NewStarTopology creates a new Star topology.
func NewStarTopology(config TopologyConfig) *StarTopology {
	config.Type = TopologyStar

	st := &StarTopology{
		BaseTopology:  NewBaseTopology(config),
		incomingQueue: make(chan *Message, 500),
		outgoingQueue: make(chan *Message, 500),
	}

	return st
}

// GetType returns the topology type.
func (st *StarTopology) GetType() TopologyType {
	return TopologyStar
}

// Initialize sets up the Star topology with agents.
func (st *StarTopology) Initialize(ctx context.Context, agents []*Agent) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if len(agents) == 0 {
		return fmt.Errorf("at least one agent required for Star topology")
	}

	// Add all agents
	for _, agent := range agents {
		st.agents[agent.ID] = agent
		st.agentsByRole[agent.Role] = append(st.agentsByRole[agent.Role], agent)
	}

	// Select moderator (highest scored moderator, or highest scored agent)
	st.selectModerator()

	// Create star channels - all agents connect to moderator
	for _, agent := range agents {
		if agent.ID != st.moderator.ID {
			// Agent -> Moderator
			st.channels = append(st.channels, CommunicationChannel{
				FromAgent:     agent.ID,
				ToAgent:       st.moderator.ID,
				Bidirectional: true,
				Weight:        agent.Score / 10.0,
			})
		}
	}

	// Start message processor
	go st.processMessages(ctx)

	return nil
}

// selectModerator selects the central moderator.
func (st *StarTopology) selectModerator() {
	// Try to find a moderator role agent first
	moderators := st.agentsByRole[RoleModerator]
	if len(moderators) > 0 {
		// Pick highest scored moderator
		sort.Slice(moderators, func(i, j int) bool {
			return moderators[i].Score > moderators[j].Score
		})
		st.moderator = moderators[0]
		return
	}

	// No moderator role, pick highest scored agent
	var bestAgent *Agent
	for _, agent := range st.agents {
		if bestAgent == nil || agent.Score > bestAgent.Score {
			bestAgent = agent
		}
	}
	st.moderator = bestAgent
}

// GetModerator returns the central moderator.
func (st *StarTopology) GetModerator() *Agent {
	st.moderatorMu.RLock()
	defer st.moderatorMu.RUnlock()
	return st.moderator
}

// SetModerator sets a new moderator.
func (st *StarTopology) SetModerator(agentID string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	agent, ok := st.agents[agentID]
	if !ok {
		return ErrAgentNotFound
	}

	st.moderatorMu.Lock()
	oldModerator := st.moderator
	st.moderator = agent
	st.moderatorMu.Unlock()

	// Rebuild channels
	st.channels = st.channels[:0]
	for _, a := range st.agents {
		if a.ID != agent.ID {
			st.channels = append(st.channels, CommunicationChannel{
				FromAgent:     a.ID,
				ToAgent:       agent.ID,
				Bidirectional: true,
				Weight:        a.Score / 10.0,
			})
		}
	}

	// Log the change
	if oldModerator != nil {
		_ = oldModerator // Could log transition
	}

	return nil
}

// CanCommunicate checks if two agents can communicate.
// In Star topology, all communication goes through the moderator.
func (st *StarTopology) CanCommunicate(fromID, toID string) bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	st.moderatorMu.RLock()
	moderatorID := st.moderator.ID
	st.moderatorMu.RUnlock()

	_, fromExists := st.agents[fromID]
	_, toExists := st.agents[toID]

	if !fromExists || !toExists || fromID == toID {
		return false
	}

	// Direct communication only if one is the moderator
	return fromID == moderatorID || toID == moderatorID
}

// GetCommunicationTargets returns agents that can receive messages.
// In Star topology, non-moderators can only message the moderator.
// The moderator can message everyone.
func (st *StarTopology) GetCommunicationTargets(agentID string) []*Agent {
	st.mu.RLock()
	defer st.mu.RUnlock()

	st.moderatorMu.RLock()
	moderator := st.moderator
	st.moderatorMu.RUnlock()

	if agentID == moderator.ID {
		// Moderator can message everyone
		targets := make([]*Agent, 0, len(st.agents)-1)
		for id, agent := range st.agents {
			if id != agentID {
				targets = append(targets, agent)
			}
		}
		return targets
	}

	// Non-moderator can only message moderator
	return []*Agent{moderator}
}

// RouteMessage determines routing for a message.
// In Star topology, all messages route through the moderator.
func (st *StarTopology) RouteMessage(msg *Message) ([]*Agent, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	st.moderatorMu.RLock()
	moderator := st.moderator
	st.moderatorMu.RUnlock()

	_, senderExists := st.agents[msg.FromAgent]
	if !senderExists {
		return nil, fmt.Errorf("%w: sender %s", ErrAgentNotFound, msg.FromAgent)
	}

	// If sender is moderator, can route to anyone
	if msg.FromAgent == moderator.ID {
		if len(msg.ToAgents) > 0 {
			targets := make([]*Agent, 0, len(msg.ToAgents))
			for _, targetID := range msg.ToAgents {
				if agent, ok := st.agents[targetID]; ok {
					targets = append(targets, agent)
				}
			}
			return targets, nil
		}
		// Broadcast from moderator
		return st.GetCommunicationTargets(moderator.ID), nil
	}

	// Non-moderator messages go to moderator first
	return []*Agent{moderator}, nil
}

// BroadcastMessage sends a message to all agents through the moderator.
func (st *StarTopology) BroadcastMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	st.moderatorMu.RLock()
	moderator := st.moderator
	st.moderatorMu.RUnlock()

	// Non-moderator broadcasts must go through moderator
	if msg.FromAgent != moderator.ID {
		msg.ToAgents = []string{moderator.ID}
		msg.Metadata = map[string]interface{}{
			"broadcast_request": true,
			"original_sender":   msg.FromAgent,
		}
	}

	st.IncrementMessageCount(true)

	targets, err := st.RouteMessage(msg)
	if err != nil {
		return err
	}

	// Send sequentially through moderator
	for _, target := range targets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case st.outgoingQueue <- msg:
			target.UpdateActivity(0) // Will update with actual time on response
		}
	}

	return nil
}

// SendMessage sends a message through the star topology.
func (st *StarTopology) SendMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	if len(msg.ToAgents) == 0 {
		return st.BroadcastMessage(ctx, msg)
	}

	st.IncrementMessageCount(false)

	targets, err := st.RouteMessage(msg)
	if err != nil {
		return err
	}

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case st.outgoingQueue <- msg:
			target.UpdateActivity(0)
		}
	}

	return nil
}

// processMessages processes messages through the moderator.
func (st *StarTopology) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-st.outgoingQueue:
			if !ok {
				return
			}
			// Moderator processes/relays the message
			st.handleModeratorRelay(ctx, msg)
		}
	}
}

// handleModeratorRelay handles message relay through moderator.
func (st *StarTopology) handleModeratorRelay(ctx context.Context, msg *Message) {
	// Check if this is a broadcast request from a non-moderator
	if broadcastReq, ok := msg.Metadata["broadcast_request"].(bool); ok && broadcastReq {
		// Relay to all other agents
		st.mu.RLock()
		for _, agent := range st.agents {
			if agent.ID != msg.FromAgent && agent.ID != st.moderator.ID {
				// Create relay message
				relayMsg := &Message{
					ID:          uuid.New().String(),
					FromAgent:   st.moderator.ID,
					ToAgents:    []string{agent.ID},
					Content:     msg.Content,
					MessageType: msg.MessageType,
					Phase:       msg.Phase,
					Round:       msg.Round,
					Timestamp:   time.Now(),
					ReplyTo:     msg.ID,
					Confidence:  msg.Confidence,
					Metadata: map[string]interface{}{
						"relayed_from": msg.FromAgent,
						"original_id":  msg.ID,
					},
				}
				select {
				case st.incomingQueue <- relayMsg:
				case <-ctx.Done():
					st.mu.RUnlock()
					return
				}
			}
		}
		st.mu.RUnlock()
	}
}

// SelectLeader returns the moderator as the leader (in Star topology).
func (st *StarTopology) SelectLeader(phase DebatePhase) (*Agent, error) {
	st.moderatorMu.RLock()
	defer st.moderatorMu.RUnlock()

	if st.moderator == nil {
		return nil, fmt.Errorf("no moderator set")
	}
	return st.moderator, nil
}

// GetParallelGroups returns groups for parallel execution.
// In Star topology, parallelism is limited - agents work in batches.
func (st *StarTopology) GetParallelGroups(phase DebatePhase) [][]*Agent {
	st.mu.RLock()
	defer st.mu.RUnlock()

	st.moderatorMu.RLock()
	moderator := st.moderator
	st.moderatorMu.RUnlock()

	// In Star topology, all non-moderator agents can work in parallel
	// but results must flow through moderator
	nonModerators := make([]*Agent, 0, len(st.agents)-1)
	for _, agent := range st.agents {
		if agent.ID != moderator.ID {
			nonModerators = append(nonModerators, agent)
		}
	}

	// Batch them based on max parallelism
	batchSize := st.config.MaxParallelism
	if batchSize <= 0 {
		batchSize = 4
	}

	var groups [][]*Agent
	for i := 0; i < len(nonModerators); i += batchSize {
		end := i + batchSize
		if end > len(nonModerators) {
			end = len(nonModerators)
		}
		groups = append(groups, nonModerators[i:end])
	}

	return groups
}

// Close closes the Star topology.
func (st *StarTopology) Close() error {
	close(st.incomingQueue)
	close(st.outgoingQueue)
	return st.BaseTopology.Close()
}
