package topology

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func createTestAgents(count int) []*Agent {
	roles := []AgentRole{
		RoleProposer, RoleCritic, RoleReviewer, RoleOptimizer, RoleModerator,
		RoleArchitect, RoleSecurity, RoleTestAgent, RoleRedTeam, RoleBlueTeam,
		RoleValidator, RoleTeacher,
	}
	specializations := []string{"code", "reasoning", "general", "vision", "search"}

	agents := make([]*Agent, count)
	for i := 0; i < count; i++ {
		agents[i] = &Agent{
			ID:             generateTestID(i),
			Role:           roles[i%len(roles)],
			Provider:       "test_provider",
			Model:          "test_model",
			Score:          7.5 + float64(i)*0.1,
			Confidence:     0.5 + float64(i)*0.05,
			Specialization: specializations[i%len(specializations)],
			Capabilities:   []string{"test_capability"},
			Metadata:       make(map[string]interface{}),
		}
	}
	return agents
}

func generateTestID(i int) string {
	return "agent-" + string(rune('a'+i))
}

// ============================================================================
// Base Topology Tests
// ============================================================================

func TestBaseTopology_Creation(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	assert.NotNil(t, bt)
	assert.Equal(t, TopologyGraphMesh, bt.config.Type)
	assert.NotNil(t, bt.agents)
	assert.NotNil(t, bt.agentsByRole)
	assert.NotNil(t, bt.channels)
}

func TestBaseTopology_AddAndGetAgent(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	agent := &Agent{
		ID:             "test-agent-1",
		Role:           RoleProposer,
		Provider:       "claude",
		Model:          "claude-3-opus",
		Score:          8.5,
		Specialization: "reasoning",
	}

	bt.AddAgent(agent)

	// Test GetAgents
	agents := bt.GetAgents()
	assert.Len(t, agents, 1)
	assert.Equal(t, agent.ID, agents[0].ID)

	// Test GetAgent
	retrieved, ok := bt.GetAgent("test-agent-1")
	assert.True(t, ok)
	assert.Equal(t, agent.ID, retrieved.ID)

	// Test non-existent agent
	_, ok = bt.GetAgent("non-existent")
	assert.False(t, ok)
}

func TestBaseTopology_GetAgentsByRole(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	agents := createTestAgents(6)
	for _, agent := range agents {
		bt.AddAgent(agent)
	}

	// Get proposers
	proposers := bt.GetAgentsByRole(RoleProposer)
	assert.True(t, len(proposers) >= 1)

	// Check all returned agents have the correct role
	for _, agent := range proposers {
		assert.Equal(t, RoleProposer, agent.Role)
	}
}

func TestBaseTopology_AssignRole(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	agent := &Agent{
		ID:   "test-agent-1",
		Role: RoleProposer,
	}
	bt.AddAgent(agent)

	// Assign new role
	err := bt.AssignRole("test-agent-1", RoleCritic)
	assert.NoError(t, err)

	// Verify role changed
	retrieved, _ := bt.GetAgent("test-agent-1")
	assert.Equal(t, RoleCritic, retrieved.Role)

	// Check role lists updated
	proposers := bt.GetAgentsByRole(RoleProposer)
	assert.Len(t, proposers, 0)

	critics := bt.GetAgentsByRole(RoleCritic)
	assert.Len(t, critics, 1)
}

func TestBaseTopology_AssignRole_NotFound(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	err := bt.AssignRole("non-existent", RoleCritic)
	assert.Error(t, err)
	assert.Equal(t, ErrAgentNotFound, err)
}

func TestBaseTopology_GetNextPhase(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	testCases := []struct {
		current  DebatePhase
		expected DebatePhase
	}{
		{PhaseProposal, PhaseCritique},
		{PhaseCritique, PhaseReview},
		{PhaseReview, PhaseOptimization},
		{PhaseOptimization, PhaseConvergence},
		{PhaseConvergence, PhaseProposal}, // Cycles back
	}

	for _, tc := range testCases {
		t.Run(string(tc.current)+"->"+string(tc.expected), func(t *testing.T) {
			next := bt.GetNextPhase(tc.current)
			assert.Equal(t, tc.expected, next)
		})
	}
}

func TestBaseTopology_Metrics(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	// Initial metrics
	metrics := bt.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalMessages)
	assert.Equal(t, int64(0), metrics.BroadcastCount)

	// Increment message counts
	bt.IncrementMessageCount(true)  // broadcast
	bt.IncrementMessageCount(false) // direct
	bt.IncrementMessageCount(false) // direct

	metrics = bt.GetMetrics()
	assert.Equal(t, int64(3), metrics.TotalMessages)
	assert.Equal(t, int64(1), metrics.BroadcastCount)
	assert.Equal(t, int64(2), metrics.DirectMessageCount)
}

func TestBaseTopology_Close(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	bt := NewBaseTopology(config)

	err := bt.Close()
	assert.NoError(t, err)

	// Second close should be idempotent
	err = bt.Close()
	assert.NoError(t, err)
}

// ============================================================================
// Graph-Mesh Topology Tests
// ============================================================================

func TestGraphMeshTopology_Creation(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	assert.NotNil(t, gm)
	assert.Equal(t, TopologyGraphMesh, gm.GetType())
}

func TestGraphMeshTopology_Initialize(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(5)

	err := gm.Initialize(ctx, agents)
	assert.NoError(t, err)

	// Verify all agents added
	assert.Len(t, gm.GetAgents(), 5)

	// Verify fully connected mesh (n*(n-1) channels for n agents)
	channels := gm.GetChannels()
	expectedChannels := 5 * 4 // Each agent connects to 4 others
	assert.Equal(t, expectedChannels, len(channels))
}

func TestGraphMeshTopology_CanCommunicate(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = gm.Initialize(ctx, agents)

	// All agents can communicate with each other
	assert.True(t, gm.CanCommunicate(agents[0].ID, agents[1].ID))
	assert.True(t, gm.CanCommunicate(agents[1].ID, agents[2].ID))
	assert.True(t, gm.CanCommunicate(agents[2].ID, agents[0].ID))

	// Cannot communicate with self
	assert.False(t, gm.CanCommunicate(agents[0].ID, agents[0].ID))

	// Cannot communicate with non-existent agent
	assert.False(t, gm.CanCommunicate(agents[0].ID, "non-existent"))
}

func TestGraphMeshTopology_GetCommunicationTargets(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)
	_ = gm.Initialize(ctx, agents)

	targets := gm.GetCommunicationTargets(agents[0].ID)

	// Should be able to target all other agents (n-1)
	assert.Len(t, targets, 3)

	// Verify targets don't include self
	for _, target := range targets {
		assert.NotEqual(t, agents[0].ID, target.ID)
	}

	// Verify targets are sorted by score (highest first)
	for i := 0; i < len(targets)-1; i++ {
		assert.GreaterOrEqual(t, targets[i].Score, targets[i+1].Score)
	}
}

func TestGraphMeshTopology_RouteMessage(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)
	_ = gm.Initialize(ctx, agents)

	// Test specific targets
	msg := &Message{
		FromAgent: agents[0].ID,
		ToAgents:  []string{agents[1].ID, agents[2].ID},
	}

	targets, err := gm.RouteMessage(msg)
	assert.NoError(t, err)
	assert.Len(t, targets, 2)

	// Test broadcast (empty ToAgents)
	msg = &Message{
		FromAgent: agents[0].ID,
		ToAgents:  nil,
	}

	targets, err = gm.RouteMessage(msg)
	assert.NoError(t, err)
	assert.Len(t, targets, 3) // All except sender
}

func TestGraphMeshTopology_RouteMessage_InvalidSender(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(2)
	_ = gm.Initialize(ctx, agents)

	msg := &Message{
		FromAgent: "non-existent",
		ToAgents:  []string{agents[0].ID},
	}

	_, err := gm.RouteMessage(msg)
	assert.Error(t, err)
}

func TestGraphMeshTopology_BroadcastMessage(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	agents := createTestAgents(3)
	_ = gm.Initialize(ctx, agents)

	msg := &Message{
		FromAgent:   agents[0].ID,
		Content:     "Test broadcast",
		MessageType: MessageTypeProposal,
		Phase:       PhaseProposal,
	}

	err := gm.BroadcastMessage(ctx, msg)
	assert.NoError(t, err)

	// Verify message has ID and timestamp
	assert.NotEmpty(t, msg.ID)
	assert.False(t, msg.Timestamp.IsZero())

	// Verify metrics updated
	metrics := gm.GetMetrics()
	assert.Equal(t, int64(1), metrics.BroadcastCount)
}

func TestGraphMeshTopology_SendMessage(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	agents := createTestAgents(3)
	_ = gm.Initialize(ctx, agents)

	msg := &Message{
		FromAgent:   agents[0].ID,
		ToAgents:    []string{agents[1].ID},
		Content:     "Test direct message",
		MessageType: MessageTypeCritique,
		Phase:       PhaseCritique,
	}

	err := gm.SendMessage(ctx, msg)
	assert.NoError(t, err)

	metrics := gm.GetMetrics()
	assert.Equal(t, int64(1), metrics.DirectMessageCount)
}

func TestGraphMeshTopology_SelectLeader(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(6)
	_ = gm.Initialize(ctx, agents)

	// Test leader selection for different phases
	phases := []DebatePhase{PhaseProposal, PhaseCritique, PhaseReview, PhaseOptimization, PhaseConvergence}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			leader, err := gm.SelectLeader(phase)
			assert.NoError(t, err)
			assert.NotNil(t, leader)

			// Verify leader is tracked
			history := gm.GetLeaderHistory()
			assert.True(t, len(history) > 0)
		})
	}
}

func TestGraphMeshTopology_GetParallelGroups(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(8)
	_ = gm.Initialize(ctx, agents)

	groups := gm.GetParallelGroups(PhaseProposal)
	assert.NotNil(t, groups)

	// Verify all agents appear in at least one group
	agentIDs := make(map[string]bool)
	for _, group := range groups {
		for _, agent := range group {
			agentIDs[agent.ID] = true
		}
	}
	// Some agents should be in groups (based on role)
	assert.True(t, len(agentIDs) > 0)
}

func TestGraphMeshTopology_SetAndGetPhase(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	assert.Equal(t, PhaseProposal, gm.GetCurrentPhase())

	gm.SetPhase(PhaseCritique, "test-agent", "test transition")
	assert.Equal(t, PhaseCritique, gm.GetCurrentPhase())

	history := gm.GetPhaseHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, PhaseProposal, history[0].FromPhase)
	assert.Equal(t, PhaseCritique, history[0].ToPhase)
}

func TestGraphMeshTopology_ExecuteParallelPhase(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)
	_ = gm.Initialize(ctx, agents)

	var executedAgents []string
	var mu sync.Mutex

	taskFn := func(agent *Agent) error {
		mu.Lock()
		executedAgents = append(executedAgents, agent.ID)
		mu.Unlock()
		return nil
	}

	errors := gm.ExecuteParallelPhase(ctx, PhaseProposal, taskFn)
	assert.Empty(t, errors)

	// Some agents should have executed
	assert.True(t, len(executedAgents) > 0)
}

func TestGraphMeshTopology_RegisterHandler(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	handler := func(ctx context.Context, msg *Message) error {
		// Handler logic would go here
		return nil
	}

	gm.RegisterHandler(MessageTypeProposal, handler)

	// Verify handler is registered
	gm.handlerMu.RLock()
	handlers := gm.handlers[MessageTypeProposal]
	gm.handlerMu.RUnlock()

	assert.Len(t, handlers, 1)
}

func TestGraphMeshTopology_GetTopologySnapshot(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = gm.Initialize(ctx, agents)

	snapshot := gm.GetTopologySnapshot()

	assert.Equal(t, TopologyGraphMesh, snapshot.Type)
	assert.Len(t, snapshot.Agents, 3)
	assert.NotEmpty(t, snapshot.Channels)
	assert.False(t, snapshot.Timestamp.IsZero())
}

// ============================================================================
// Star Topology Tests
// ============================================================================

func TestStarTopology_Creation(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	assert.NotNil(t, st)
	assert.Equal(t, TopologyStar, st.GetType())
}

func TestStarTopology_Initialize(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	ctx := context.Background()
	agents := createTestAgents(5)
	agents[2].Role = RoleModerator // Ensure we have a moderator

	err := st.Initialize(ctx, agents)
	assert.NoError(t, err)

	// Verify moderator is set
	moderator := st.GetModerator()
	assert.NotNil(t, moderator)

	// Verify star channels (n-1 channels, all to moderator)
	channels := st.GetChannels()
	assert.Equal(t, 4, len(channels))
}

func TestStarTopology_ModeratorSelection(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	ctx := context.Background()

	// Test with explicit moderator role
	agents := createTestAgents(3)
	agents[1].Role = RoleModerator
	agents[1].Score = 9.0 // Highest score

	_ = st.Initialize(ctx, agents)

	moderator := st.GetModerator()
	assert.Equal(t, RoleModerator, moderator.Role)
}

func TestStarTopology_SetModerator(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = st.Initialize(ctx, agents)

	originalModerator := st.GetModerator()

	// Set new moderator
	newModeratorID := agents[2].ID
	if agents[2].ID == originalModerator.ID {
		newModeratorID = agents[0].ID
	}

	err := st.SetModerator(newModeratorID)
	assert.NoError(t, err)

	newModerator := st.GetModerator()
	assert.Equal(t, newModeratorID, newModerator.ID)
}

func TestStarTopology_CanCommunicate(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = st.Initialize(ctx, agents)

	moderator := st.GetModerator()

	// Non-moderator can communicate with moderator
	for _, agent := range agents {
		if agent.ID != moderator.ID {
			assert.True(t, st.CanCommunicate(agent.ID, moderator.ID))
			assert.True(t, st.CanCommunicate(moderator.ID, agent.ID))
		}
	}

	// Non-moderators cannot communicate directly with each other
	var nonModerators []*Agent
	for _, agent := range agents {
		if agent.ID != moderator.ID {
			nonModerators = append(nonModerators, agent)
		}
	}

	if len(nonModerators) >= 2 {
		assert.False(t, st.CanCommunicate(nonModerators[0].ID, nonModerators[1].ID))
	}
}

func TestStarTopology_GetCommunicationTargets(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)
	_ = st.Initialize(ctx, agents)

	moderator := st.GetModerator()

	// Moderator can target everyone
	moderatorTargets := st.GetCommunicationTargets(moderator.ID)
	assert.Len(t, moderatorTargets, 3)

	// Non-moderator can only target moderator
	for _, agent := range agents {
		if agent.ID != moderator.ID {
			targets := st.GetCommunicationTargets(agent.ID)
			assert.Len(t, targets, 1)
			assert.Equal(t, moderator.ID, targets[0].ID)
		}
	}
}

func TestStarTopology_SelectLeader(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	st := NewStarTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = st.Initialize(ctx, agents)

	// Leader should always be the moderator
	leader, err := st.SelectLeader(PhaseProposal)
	assert.NoError(t, err)

	moderator := st.GetModerator()
	assert.Equal(t, moderator.ID, leader.ID)
}

func TestStarTopology_GetParallelGroups(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	config.MaxParallelism = 2
	st := NewStarTopology(config)

	ctx := context.Background()
	agents := createTestAgents(5)
	_ = st.Initialize(ctx, agents)

	groups := st.GetParallelGroups(PhaseProposal)

	// Should have batched groups
	assert.True(t, len(groups) > 0)

	// Each group should not exceed max parallelism
	for _, group := range groups {
		assert.LessOrEqual(t, len(group), config.MaxParallelism)
	}
}

// ============================================================================
// Chain Topology Tests
// ============================================================================

func TestChainTopology_Creation(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	assert.NotNil(t, ct)
	assert.Equal(t, TopologyChain, ct.GetType())
}

func TestChainTopology_Initialize(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)

	err := ct.Initialize(ctx, agents)
	assert.NoError(t, err)

	// Verify chain is built
	chain := ct.GetChain()
	assert.Len(t, chain, 4)

	// Verify chain channels (n-1 forward + 1 loop back)
	channels := ct.GetChannels()
	assert.Equal(t, 4, len(channels)) // 3 forward + 1 loop back
}

func TestChainTopology_ChainOrder(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	agents[0].Role = RoleProposer
	agents[1].Role = RoleCritic
	agents[2].Role = RoleModerator

	_ = ct.Initialize(ctx, agents)

	// Verify chain is built
	chain := ct.GetChain()
	assert.Len(t, chain, 3)

	// Proposer should come before Critic, Critic before Moderator
	proposerPos := ct.GetChainPosition(agents[0].ID)
	criticPos := ct.GetChainPosition(agents[1].ID)
	moderatorPos := ct.GetChainPosition(agents[2].ID)

	assert.Less(t, proposerPos, moderatorPos)
	assert.Less(t, criticPos, moderatorPos)
}

func TestChainTopology_GetChainPosition(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	for i, agent := range agents {
		pos := ct.GetChainPosition(agent.ID)
		assert.GreaterOrEqual(t, pos, 0)
		assert.Less(t, pos, len(agents))
		_ = i // Position may differ due to role ordering
	}

	// Non-existent agent
	pos := ct.GetChainPosition("non-existent")
	assert.Equal(t, -1, pos)
}

func TestChainTopology_GetNextAgent(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	chain := ct.GetChain()

	// Each agent should have a next agent
	for i, agent := range chain {
		next := ct.GetNextAgent(agent.ID)
		assert.NotNil(t, next)

		expectedNextIdx := (i + 1) % len(chain)
		assert.Equal(t, chain[expectedNextIdx].ID, next.ID)
	}
}

func TestChainTopology_GetPreviousAgent(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	chain := ct.GetChain()

	// Each agent should have a previous agent
	for i, agent := range chain {
		prev := ct.GetPreviousAgent(agent.ID)
		assert.NotNil(t, prev)

		expectedPrevIdx := i - 1
		if expectedPrevIdx < 0 {
			expectedPrevIdx = len(chain) - 1
		}
		assert.Equal(t, chain[expectedPrevIdx].ID, prev.ID)
	}
}

func TestChainTopology_CanCommunicate(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	chain := ct.GetChain()

	// Can only communicate with next in chain
	for i, agent := range chain {
		nextIdx := (i + 1) % len(chain)
		assert.True(t, ct.CanCommunicate(agent.ID, chain[nextIdx].ID))

		// Cannot communicate with non-adjacent (unless it's the next)
		for j, other := range chain {
			if j != nextIdx && other.ID != agent.ID {
				assert.False(t, ct.CanCommunicate(agent.ID, other.ID))
			}
		}
	}
}

func TestChainTopology_GetCommunicationTargets(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)
	_ = ct.Initialize(ctx, agents)

	chain := ct.GetChain()

	for i, agent := range chain {
		targets := ct.GetCommunicationTargets(agent.ID)
		assert.Len(t, targets, 1)

		nextIdx := (i + 1) % len(chain)
		assert.Equal(t, chain[nextIdx].ID, targets[0].ID)
	}
}

func TestChainTopology_AdvancePosition(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	initialPos := ct.GetCurrentPosition()
	assert.Equal(t, 0, initialPos)

	// Advance
	next := ct.AdvancePosition()
	assert.NotNil(t, next)
	assert.Equal(t, 1, ct.GetCurrentPosition())

	// Advance again
	ct.AdvancePosition()
	assert.Equal(t, 2, ct.GetCurrentPosition())

	// Wrap around
	ct.AdvancePosition()
	assert.Equal(t, 0, ct.GetCurrentPosition())
}

func TestChainTopology_ReorderChain(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	// Get original chain
	originalChain := ct.GetChain()

	// Create reversed order
	newOrder := make([]string, len(originalChain))
	for i, agent := range originalChain {
		newOrder[len(originalChain)-1-i] = agent.ID
	}

	err := ct.ReorderChain(newOrder)
	assert.NoError(t, err)

	// Verify new order
	newChain := ct.GetChain()
	for i, agent := range newChain {
		assert.Equal(t, newOrder[i], agent.ID)
	}
}

func TestChainTopology_ReorderChain_InvalidCount(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(3)
	_ = ct.Initialize(ctx, agents)

	err := ct.ReorderChain([]string{"agent-a", "agent-b"}) // Wrong count
	assert.Error(t, err)
}

func TestChainTopology_SelectLeader(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(5)
	agents[0].Role = RoleProposer
	agents[1].Role = RoleCritic
	agents[2].Role = RoleReviewer
	agents[3].Role = RoleOptimizer
	agents[4].Role = RoleModerator

	_ = ct.Initialize(ctx, agents)

	// Different phases should select different leaders
	proposalLeader, err := ct.SelectLeader(PhaseProposal)
	assert.NoError(t, err)
	assert.Contains(t, []AgentRole{RoleProposer, RoleArchitect}, proposalLeader.Role)

	critiqueLeader, err := ct.SelectLeader(PhaseCritique)
	assert.NoError(t, err)
	assert.Contains(t, []AgentRole{RoleCritic, RoleRedTeam}, critiqueLeader.Role)
}

func TestChainTopology_GetParallelGroups(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	ct := NewChainTopology(config)

	ctx := context.Background()
	agents := createTestAgents(4)
	_ = ct.Initialize(ctx, agents)

	groups := ct.GetParallelGroups(PhaseProposal)

	// In Chain topology, each agent is its own group (no parallelism)
	assert.Len(t, groups, 4)
	for _, group := range groups {
		assert.Len(t, group, 1)
	}
}

// ============================================================================
// Factory Tests
// ============================================================================

func TestNewTopology_GraphMesh(t *testing.T) {
	config := DefaultTopologyConfig(TopologyGraphMesh)
	topology, err := NewTopology(TopologyGraphMesh, config)

	assert.NoError(t, err)
	assert.NotNil(t, topology)
	assert.Equal(t, TopologyGraphMesh, topology.GetType())
}

func TestNewTopology_Star(t *testing.T) {
	config := DefaultTopologyConfig(TopologyStar)
	topology, err := NewTopology(TopologyStar, config)

	assert.NoError(t, err)
	assert.NotNil(t, topology)
	assert.Equal(t, TopologyStar, topology.GetType())
}

func TestNewTopology_Chain(t *testing.T) {
	config := DefaultTopologyConfig(TopologyChain)
	topology, err := NewTopology(TopologyChain, config)

	assert.NoError(t, err)
	assert.NotNil(t, topology)
	assert.Equal(t, TopologyChain, topology.GetType())
}

func TestNewTopology_Unknown(t *testing.T) {
	config := TopologyConfig{Type: "unknown"}
	_, err := NewTopology("unknown", config)

	assert.Error(t, err)
}

func TestNewDefaultTopology(t *testing.T) {
	topology := NewDefaultTopology()

	assert.NotNil(t, topology)
	assert.Equal(t, TopologyGraphMesh, topology.GetType())
}

func TestSelectTopologyType(t *testing.T) {
	testCases := []struct {
		name         string
		agentCount   int
		requirements TopologyRequirements
		expected     TopologyType
	}{
		{
			name:         "ordering required",
			agentCount:   10,
			requirements: TopologyRequirements{RequireOrdering: true},
			expected:     TopologyChain,
		},
		{
			name:         "deterministic required",
			agentCount:   10,
			requirements: TopologyRequirements{Deterministic: true},
			expected:     TopologyChain,
		},
		{
			name:         "centralized control",
			agentCount:   10,
			requirements: TopologyRequirements{CentralizedControl: true},
			expected:     TopologyStar,
		},
		{
			name:         "small team",
			agentCount:   3,
			requirements: TopologyRequirements{},
			expected:     TopologyChain,
		},
		{
			name:         "medium team low parallelism",
			agentCount:   5,
			requirements: TopologyRequirements{MaxParallelism: 2},
			expected:     TopologyStar,
		},
		{
			name:         "large team",
			agentCount:   12,
			requirements: TopologyRequirements{},
			expected:     TopologyGraphMesh,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SelectTopologyType(tc.agentCount, tc.requirements)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateOptimalTopology(t *testing.T) {
	requirements := TopologyRequirements{
		MaxParallelism:     4,
		EnableDynamicRoles: true,
	}

	topology, err := CreateOptimalTopology(10, requirements)
	assert.NoError(t, err)
	assert.NotNil(t, topology)
}

func TestGetTopologyComparison(t *testing.T) {
	comparison := GetTopologyComparison()

	assert.Equal(t, TopologyGraphMesh, comparison.GraphMesh.Type)
	assert.Equal(t, TopologyStar, comparison.Star.Type)
	assert.Equal(t, TopologyChain, comparison.Chain.Type)

	// Verify research performance rankings
	assert.Contains(t, comparison.GraphMesh.ResearchPerformance, "#1")
	assert.Contains(t, comparison.Star.ResearchPerformance, "#2")
	assert.Contains(t, comparison.Chain.ResearchPerformance, "#3")
}

func TestGetRecommendedRoleAssignments(t *testing.T) {
	assignments := GetRecommendedRoleAssignments()

	// Minimum team should have 5 roles
	minTotal := 0
	for _, count := range assignments.MinimumTeam {
		minTotal += count
	}
	assert.Equal(t, 5, minTotal)

	// Standard team should have 12 roles
	stdTotal := 0
	for _, count := range assignments.StandardTeam {
		stdTotal += count
	}
	assert.Equal(t, 12, stdTotal)
}

func TestCreateAgentFromSpec(t *testing.T) {
	agent := CreateAgentFromSpec("test-1", RoleProposer, "claude", "claude-3-opus", 8.5, "code")

	assert.Equal(t, "test-1", agent.ID)
	assert.Equal(t, RoleProposer, agent.Role)
	assert.Equal(t, "claude", agent.Provider)
	assert.Equal(t, "claude-3-opus", agent.Model)
	assert.Equal(t, 8.5, agent.Score)
	assert.Equal(t, "code", agent.Specialization)
	assert.Contains(t, agent.Capabilities, "code_generation")
}

// ============================================================================
// Agent Tests
// ============================================================================

func TestAgent_UpdateActivity(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Role: RoleProposer,
	}

	// Initial state
	metrics := agent.GetMetrics()
	assert.Equal(t, 0, metrics.MessageCount)
	assert.True(t, metrics.LastActive.IsZero())

	// Update activity
	agent.UpdateActivity(100 * time.Millisecond)

	metrics = agent.GetMetrics()
	assert.Equal(t, 1, metrics.MessageCount)
	assert.Equal(t, 100*time.Millisecond, metrics.AvgResponseTime)
	assert.False(t, metrics.LastActive.IsZero())

	// Update again - should average
	agent.UpdateActivity(200 * time.Millisecond)

	metrics = agent.GetMetrics()
	assert.Equal(t, 2, metrics.MessageCount)
	assert.Equal(t, 150*time.Millisecond, metrics.AvgResponseTime)
}

func TestAgent_ConcurrentUpdates(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Role: RoleProposer,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.UpdateActivity(50 * time.Millisecond)
		}()
	}

	wg.Wait()

	metrics := agent.GetMetrics()
	assert.Equal(t, 100, metrics.MessageCount)
}

// ============================================================================
// Message Tests
// ============================================================================

func TestMessage_Types(t *testing.T) {
	messageTypes := []MessageType{
		MessageTypeProposal,
		MessageTypeCritique,
		MessageTypeReview,
		MessageTypeOptimization,
		MessageTypeConvergence,
		MessageTypeQuestion,
		MessageTypeAnswer,
		MessageTypeAcknowledge,
		MessageTypeValidation,
		MessageTypeRefinement,
	}

	for _, mt := range messageTypes {
		assert.NotEmpty(t, string(mt))
	}
}

func TestDebatePhase_Order(t *testing.T) {
	phases := []DebatePhase{
		PhaseProposal,
		PhaseCritique,
		PhaseReview,
		PhaseOptimization,
		PhaseConvergence,
	}

	for _, p := range phases {
		assert.NotEmpty(t, string(p))
	}
}

// ============================================================================
// Error Tests
// ============================================================================

func TestTopologyErrors(t *testing.T) {
	assert.Equal(t, "agent not found in topology", ErrAgentNotFound.Error())
	assert.Equal(t, "communication channel not found", ErrChannelNotFound.Error())
	assert.Equal(t, "topology has been closed", ErrTopologyClosed.Error())
	assert.Equal(t, "invalid message format", ErrInvalidMessage.Error())
	assert.Equal(t, "message routing failed", ErrRoutingFailed.Error())
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestTopology_FullDebateFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create Graph-Mesh topology (best performer)
	config := DefaultTopologyConfig(TopologyGraphMesh)
	gm := NewGraphMeshTopology(config)

	// Initialize with diverse agents
	agents := []*Agent{
		CreateAgentFromSpec("proposer-1", RoleProposer, "claude", "claude-3-opus", 8.5, "code"),
		CreateAgentFromSpec("critic-1", RoleCritic, "deepseek", "deepseek-coder", 8.0, "reasoning"),
		CreateAgentFromSpec("reviewer-1", RoleReviewer, "gemini", "gemini-pro", 7.8, "general"),
		CreateAgentFromSpec("optimizer-1", RoleOptimizer, "mistral", "mistral-large", 7.5, "code"),
		CreateAgentFromSpec("moderator-1", RoleModerator, "openai", "gpt-4", 8.2, "reasoning"),
	}

	err := gm.Initialize(ctx, agents)
	require.NoError(t, err)

	// Phase 1: Proposal
	gm.SetPhase(PhaseProposal, "test", "starting debate")
	leader, err := gm.SelectLeader(PhaseProposal)
	require.NoError(t, err)
	assert.NotNil(t, leader)

	// Send proposal
	proposalMsg := &Message{
		FromAgent:   leader.ID,
		Content:     "Initial proposal for the debate topic",
		MessageType: MessageTypeProposal,
		Phase:       PhaseProposal,
		Round:       1,
		Confidence:  0.8,
	}
	err = gm.BroadcastMessage(ctx, proposalMsg)
	require.NoError(t, err)

	// Phase 2: Critique
	nextPhase := gm.GetNextPhase(PhaseProposal)
	assert.Equal(t, PhaseCritique, nextPhase)
	gm.SetPhase(PhaseCritique, "test", "critique phase")

	// Phase 3: Review
	nextPhase = gm.GetNextPhase(PhaseCritique)
	assert.Equal(t, PhaseReview, nextPhase)

	// Phase 4: Optimization
	nextPhase = gm.GetNextPhase(PhaseReview)
	assert.Equal(t, PhaseOptimization, nextPhase)

	// Phase 5: Convergence
	nextPhase = gm.GetNextPhase(PhaseOptimization)
	assert.Equal(t, PhaseConvergence, nextPhase)

	// Verify metrics
	metrics := gm.GetMetrics()
	assert.Greater(t, metrics.TotalMessages, int64(0))
	assert.Greater(t, metrics.PhaseTransitions, 0)

	// Clean up
	err = gm.Close()
	assert.NoError(t, err)
}
