// Package topology provides Graph-Mesh topology implementation.
// Graph-Mesh allows all agents to communicate with all others, enabling
// maximum parallelism and flexibility - the best-performing topology per ACL 2025 research.
package topology

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// GraphMeshTopology implements a fully-connected graph topology.
// All agents can communicate directly with all other agents.
// Supports concurrent planning and parallel execution.
type GraphMeshTopology struct {
	*BaseTopology

	// Message handlers
	handlers  map[MessageType][]MessageHandler
	handlerMu sync.RWMutex

	// Phase management
	currentPhase DebatePhase
	phaseHistory []PhaseTransition
	phaseMu      sync.RWMutex

	// Parallel execution groups
	parallelGroups map[DebatePhase][][]*Agent
	groupMu        sync.RWMutex

	// Message routing
	messageQueue chan *Message

	// Leader selection
	currentLeader *Agent
	leaderHistory []LeaderSelection
	leaderMu      sync.RWMutex
}

// MessageHandler handles incoming messages.
type MessageHandler func(ctx context.Context, msg *Message) error

// PhaseTransition records a phase change.
type PhaseTransition struct {
	FromPhase DebatePhase
	ToPhase   DebatePhase
	Timestamp time.Time
	Initiator string // Agent that triggered the transition
	Reason    string
}

// LeaderSelection records a leader selection.
type LeaderSelection struct {
	Phase     DebatePhase
	LeaderID  string
	Score     float64
	Timestamp time.Time
	Method    string // "score", "rotation", "dynamic"
}

// NewGraphMeshTopology creates a new Graph-Mesh topology.
func NewGraphMeshTopology(config TopologyConfig) *GraphMeshTopology {
	config.Type = TopologyGraphMesh

	gm := &GraphMeshTopology{
		BaseTopology:   NewBaseTopology(config),
		handlers:       make(map[MessageType][]MessageHandler),
		currentPhase:   PhaseProposal,
		phaseHistory:   make([]PhaseTransition, 0),
		parallelGroups: make(map[DebatePhase][][]*Agent),
		messageQueue:   make(chan *Message, 1000),
	}

	return gm
}

// GetType returns the topology type.
func (gm *GraphMeshTopology) GetType() TopologyType {
	return TopologyGraphMesh
}

// Initialize sets up the Graph-Mesh topology with agents.
func (gm *GraphMeshTopology) Initialize(ctx context.Context, agents []*Agent) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Add all agents
	for _, agent := range agents {
		gm.agents[agent.ID] = agent
		gm.agentsByRole[agent.Role] = append(gm.agentsByRole[agent.Role], agent)
	}

	// Create fully-connected mesh - every agent can talk to every other agent
	for _, from := range agents {
		for _, to := range agents {
			if from.ID != to.ID {
				channel := CommunicationChannel{
					FromAgent:     from.ID,
					ToAgent:       to.ID,
					Bidirectional: true,
					Weight:        calculateChannelWeight(from, to),
				}
				gm.channels = append(gm.channels, channel)
			}
		}
	}

	// Initialize parallel groups for each phase
	gm.initializeParallelGroups()

	// Start message processor
	go gm.processMessages(ctx)

	return nil
}

// calculateChannelWeight calculates the weight/priority of a channel based on agent scores.
func calculateChannelWeight(from, to *Agent) float64 {
	// Higher scores = higher priority channel
	avgScore := (from.Score + to.Score) / 2

	// Bonus for complementary specializations
	complementaryBonus := 0.0
	if from.Specialization != to.Specialization {
		complementaryBonus = 0.1
	}

	// Bonus for different roles (encourages cross-role communication)
	roleBonus := 0.0
	if from.Role != to.Role {
		roleBonus = 0.15
	}

	return avgScore/10.0 + complementaryBonus + roleBonus
}

// initializeParallelGroups creates parallel execution groups for each phase.
func (gm *GraphMeshTopology) initializeParallelGroups() {
	gm.groupMu.Lock()
	defer gm.groupMu.Unlock()

	// Phase: Proposal - Proposers work in parallel
	proposers := gm.agentsByRole[RoleProposer]
	if len(proposers) > 0 {
		gm.parallelGroups[PhaseProposal] = [][]*Agent{proposers}
	}

	// Phase: Critique - Critics work in parallel, then reviewers
	critics := gm.agentsByRole[RoleCritic]
	reviewers := gm.agentsByRole[RoleReviewer]
	if len(critics) > 0 || len(reviewers) > 0 {
		groups := make([][]*Agent, 0)
		if len(critics) > 0 {
			groups = append(groups, critics)
		}
		if len(reviewers) > 0 {
			groups = append(groups, reviewers)
		}
		gm.parallelGroups[PhaseCritique] = groups
	}

	// Phase: Review - Reviewers and architects work in parallel
	architects := gm.agentsByRole[RoleArchitect]
	if len(reviewers) > 0 || len(architects) > 0 {
		groups := make([][]*Agent, 0)
		if len(reviewers) > 0 {
			groups = append(groups, reviewers)
		}
		if len(architects) > 0 {
			groups = append(groups, architects)
		}
		gm.parallelGroups[PhaseReview] = groups
	}

	// Phase: Optimization - Optimizers and test agents work in parallel
	optimizers := gm.agentsByRole[RoleOptimizer]
	testAgents := gm.agentsByRole[RoleTestAgent]
	if len(optimizers) > 0 || len(testAgents) > 0 {
		groups := make([][]*Agent, 0)
		if len(optimizers) > 0 {
			groups = append(groups, optimizers)
		}
		if len(testAgents) > 0 {
			groups = append(groups, testAgents)
		}
		gm.parallelGroups[PhaseOptimization] = groups
	}

	// Phase: Convergence - Validators and moderators finalize
	validators := gm.agentsByRole[RoleValidator]
	moderators := gm.agentsByRole[RoleModerator]
	if len(validators) > 0 || len(moderators) > 0 {
		groups := make([][]*Agent, 0)
		if len(validators) > 0 {
			groups = append(groups, validators)
		}
		if len(moderators) > 0 {
			groups = append(groups, moderators)
		}
		gm.parallelGroups[PhaseConvergence] = groups
	}

	// Add Red/Blue team for all phases (they work continuously)
	redTeam := gm.agentsByRole[RoleRedTeam]
	blueTeam := gm.agentsByRole[RoleBlueTeam]
	securityAgents := gm.agentsByRole[RoleSecurity]

	adversarialGroup := make([]*Agent, 0)
	adversarialGroup = append(adversarialGroup, redTeam...)
	adversarialGroup = append(adversarialGroup, blueTeam...)
	adversarialGroup = append(adversarialGroup, securityAgents...)

	if len(adversarialGroup) > 0 {
		for phase := range gm.parallelGroups {
			gm.parallelGroups[phase] = append(gm.parallelGroups[phase], adversarialGroup)
		}
	}
}

// CanCommunicate checks if two agents can communicate directly.
// In Graph-Mesh, all agents can communicate with each other.
func (gm *GraphMeshTopology) CanCommunicate(fromID, toID string) bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	_, fromExists := gm.agents[fromID]
	_, toExists := gm.agents[toID]

	return fromExists && toExists && fromID != toID
}

// GetCommunicationTargets returns all agents that can receive messages from the given agent.
// In Graph-Mesh, every agent can communicate with every other agent.
func (gm *GraphMeshTopology) GetCommunicationTargets(agentID string) []*Agent {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	targets := make([]*Agent, 0, len(gm.agents)-1)
	for id, agent := range gm.agents {
		if id != agentID {
			targets = append(targets, agent)
		}
	}

	// Sort by score (highest first) for prioritized communication
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Score > targets[j].Score
	})

	return targets
}

// RouteMessage determines routing for a message.
func (gm *GraphMeshTopology) RouteMessage(msg *Message) ([]*Agent, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Check sender exists
	_, senderExists := gm.agents[msg.FromAgent]
	if !senderExists {
		return nil, fmt.Errorf("%w: sender %s", ErrAgentNotFound, msg.FromAgent)
	}

	// If specific targets, route to them
	if len(msg.ToAgents) > 0 {
		targets := make([]*Agent, 0, len(msg.ToAgents))
		for _, targetID := range msg.ToAgents {
			if agent, ok := gm.agents[targetID]; ok {
				targets = append(targets, agent)
			}
		}
		return targets, nil
	}

	// Broadcast - route to all except sender
	return gm.GetCommunicationTargets(msg.FromAgent), nil
}

// BroadcastMessage sends a message to all agents except the sender.
func (gm *GraphMeshTopology) BroadcastMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.ToAgents = nil // Clear to indicate broadcast
	msg.Timestamp = time.Now()

	targets, err := gm.RouteMessage(msg)
	if err != nil {
		return err
	}

	gm.IncrementMessageCount(true)

	// Send in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(targets))

	for _, target := range targets {
		wg.Add(1)
		go func(t *Agent) {
			defer wg.Done()

			if err := gm.deliverMessage(ctx, msg, t); err != nil {
				errChan <- fmt.Errorf("failed to deliver to %s: %w", t.ID, err)
			}
		}(target)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("broadcast had %d delivery failures", len(errs))
	}

	return nil
}

// SendMessage sends a message to specific agents.
func (gm *GraphMeshTopology) SendMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	if len(msg.ToAgents) == 0 {
		return gm.BroadcastMessage(ctx, msg)
	}

	gm.IncrementMessageCount(false)

	targets, err := gm.RouteMessage(msg)
	if err != nil {
		return err
	}

	// Send in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(targets))

	for _, target := range targets {
		wg.Add(1)
		go func(t *Agent) {
			defer wg.Done()

			if err := gm.deliverMessage(ctx, msg, t); err != nil {
				errChan <- fmt.Errorf("failed to deliver to %s: %w", t.ID, err)
			}
		}(target)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("send had %d delivery failures", len(errs))
	}

	return nil
}

// deliverMessage delivers a message to a specific agent.
func (gm *GraphMeshTopology) deliverMessage(ctx context.Context, msg *Message, target *Agent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case gm.messageQueue <- msg:
		// Message queued
		return nil
	default:
		// Queue full, try with timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gm.messageQueue <- msg:
			return nil
		case <-time.After(gm.config.MessageTimeout):
			return fmt.Errorf("message delivery timeout for %s", target.ID)
		}
	}
}

// processMessages processes messages from the queue.
func (gm *GraphMeshTopology) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-gm.messageQueue:
			if !ok {
				return
			}
			gm.handleMessage(ctx, msg)
		}
	}
}

// handleMessage handles an incoming message.
func (gm *GraphMeshTopology) handleMessage(ctx context.Context, msg *Message) {
	gm.handlerMu.RLock()
	handlers := gm.handlers[msg.MessageType]
	gm.handlerMu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, msg); err != nil {
			// Log error but continue processing
			continue
		}
	}
}

// RegisterHandler registers a handler for a message type.
func (gm *GraphMeshTopology) RegisterHandler(msgType MessageType, handler MessageHandler) {
	gm.handlerMu.Lock()
	defer gm.handlerMu.Unlock()

	gm.handlers[msgType] = append(gm.handlers[msgType], handler)
}

// SelectLeader selects a leader for the given phase based on agent scores and specializations.
func (gm *GraphMeshTopology) SelectLeader(phase DebatePhase) (*Agent, error) {
	gm.mu.RLock()
	candidates := gm.getLeaderCandidates(phase)
	gm.mu.RUnlock()

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates for leader in phase %s", phase)
	}

	// Score each candidate
	type scoredCandidate struct {
		agent *Agent
		score float64
	}

	scored := make([]scoredCandidate, len(candidates))
	for i, candidate := range candidates {
		score := gm.calculateLeaderScore(candidate, phase)
		scored[i] = scoredCandidate{agent: candidate, score: score}
	}

	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	leader := scored[0].agent

	// Record selection
	gm.leaderMu.Lock()
	gm.currentLeader = leader
	gm.leaderHistory = append(gm.leaderHistory, LeaderSelection{
		Phase:     phase,
		LeaderID:  leader.ID,
		Score:     scored[0].score,
		Timestamp: time.Now(),
		Method:    "score",
	})
	gm.leaderMu.Unlock()

	return leader, nil
}

// getLeaderCandidates returns agents suitable for leadership in the given phase.
func (gm *GraphMeshTopology) getLeaderCandidates(phase DebatePhase) []*Agent {
	var preferredRoles []AgentRole

	switch phase {
	case PhaseProposal:
		preferredRoles = []AgentRole{RoleProposer, RoleArchitect, RoleModerator}
	case PhaseCritique:
		preferredRoles = []AgentRole{RoleCritic, RoleReviewer, RoleRedTeam}
	case PhaseReview:
		preferredRoles = []AgentRole{RoleReviewer, RoleModerator, RoleArchitect}
	case PhaseOptimization:
		preferredRoles = []AgentRole{RoleOptimizer, RoleArchitect, RoleTestAgent}
	case PhaseConvergence:
		preferredRoles = []AgentRole{RoleModerator, RoleValidator, RoleTeacher}
	default:
		preferredRoles = []AgentRole{RoleModerator, RoleValidator}
	}

	candidates := make([]*Agent, 0)
	for _, role := range preferredRoles {
		agents := gm.agentsByRole[role]
		candidates = append(candidates, agents...)
	}

	// If no preferred role agents, use any agent
	if len(candidates) == 0 {
		for _, agent := range gm.agents {
			candidates = append(candidates, agent)
		}
	}

	return candidates
}

// calculateLeaderScore calculates a leadership score for an agent in a phase.
func (gm *GraphMeshTopology) calculateLeaderScore(agent *Agent, phase DebatePhase) float64 {
	score := agent.Score // Base score from LLMsVerifier

	// Role bonus
	roleBonus := 0.0
	switch phase {
	case PhaseProposal:
		if agent.Role == RoleProposer || agent.Role == RoleArchitect {
			roleBonus = 0.2
		}
	case PhaseCritique:
		if agent.Role == RoleCritic || agent.Role == RoleRedTeam {
			roleBonus = 0.2
		}
	case PhaseReview:
		if agent.Role == RoleReviewer || agent.Role == RoleModerator {
			roleBonus = 0.2
		}
	case PhaseOptimization:
		if agent.Role == RoleOptimizer {
			roleBonus = 0.2
		}
	case PhaseConvergence:
		if agent.Role == RoleModerator || agent.Role == RoleValidator {
			roleBonus = 0.2
		}
	}

	// Specialization bonus
	specBonus := 0.0
	switch agent.Specialization {
	case "reasoning":
		if phase == PhaseCritique || phase == PhaseReview {
			specBonus = 0.15
		}
	case "code":
		if phase == PhaseOptimization {
			specBonus = 0.15
		}
	}

	// Activity bonus (more active = better leader)
	metrics := agent.GetMetrics()
	activityBonus := float64(metrics.MessageCount) * 0.01
	if activityBonus > 0.1 {
		activityBonus = 0.1
	}

	// Confidence bonus
	confidenceBonus := agent.Confidence * 0.1

	return score + roleBonus + specBonus + activityBonus + confidenceBonus
}

// GetParallelGroups returns groups of agents that can work in parallel for a phase.
func (gm *GraphMeshTopology) GetParallelGroups(phase DebatePhase) [][]*Agent {
	gm.groupMu.RLock()
	defer gm.groupMu.RUnlock()

	groups := gm.parallelGroups[phase]
	if groups == nil {
		// Return all agents as a single group if no specific groups defined
		return [][]*Agent{gm.GetAgents()}
	}

	// Deep copy to avoid race conditions
	result := make([][]*Agent, len(groups))
	for i, group := range groups {
		result[i] = make([]*Agent, len(group))
		copy(result[i], group)
	}

	return result
}

// SetPhase sets the current debate phase.
func (gm *GraphMeshTopology) SetPhase(phase DebatePhase, initiator string, reason string) {
	gm.phaseMu.Lock()
	defer gm.phaseMu.Unlock()

	if gm.currentPhase != phase {
		gm.phaseHistory = append(gm.phaseHistory, PhaseTransition{
			FromPhase: gm.currentPhase,
			ToPhase:   phase,
			Timestamp: time.Now(),
			Initiator: initiator,
			Reason:    reason,
		})
		gm.currentPhase = phase
		gm.IncrementPhaseTransitions()
	}
}

// GetCurrentPhase returns the current debate phase.
func (gm *GraphMeshTopology) GetCurrentPhase() DebatePhase {
	gm.phaseMu.RLock()
	defer gm.phaseMu.RUnlock()

	return gm.currentPhase
}

// GetPhaseHistory returns the history of phase transitions.
func (gm *GraphMeshTopology) GetPhaseHistory() []PhaseTransition {
	gm.phaseMu.RLock()
	defer gm.phaseMu.RUnlock()

	history := make([]PhaseTransition, len(gm.phaseHistory))
	copy(history, gm.phaseHistory)
	return history
}

// GetLeaderHistory returns the history of leader selections.
func (gm *GraphMeshTopology) GetLeaderHistory() []LeaderSelection {
	gm.leaderMu.RLock()
	defer gm.leaderMu.RUnlock()

	history := make([]LeaderSelection, len(gm.leaderHistory))
	copy(history, gm.leaderHistory)
	return history
}

// GetCurrentLeader returns the current leader.
func (gm *GraphMeshTopology) GetCurrentLeader() *Agent {
	gm.leaderMu.RLock()
	defer gm.leaderMu.RUnlock()

	return gm.currentLeader
}

// ExecuteParallelPhase executes a phase with parallel agent groups.
func (gm *GraphMeshTopology) ExecuteParallelPhase(ctx context.Context, phase DebatePhase, taskFn func(agent *Agent) error) []error {
	groups := gm.GetParallelGroups(phase)
	var allErrors []error
	var errorMu sync.Mutex

	// Execute groups sequentially, agents within groups in parallel
	for _, group := range groups {
		var wg sync.WaitGroup
		groupErrors := make(chan error, len(group))

		for _, agent := range group {
			wg.Add(1)
			go func(a *Agent) {
				defer wg.Done()
				start := time.Now()

				if err := taskFn(a); err != nil {
					groupErrors <- fmt.Errorf("agent %s error: %w", a.ID, err)
				}

				a.UpdateActivity(time.Since(start))
			}(agent)
		}

		wg.Wait()
		close(groupErrors)

		for err := range groupErrors {
			errorMu.Lock()
			allErrors = append(allErrors, err)
			errorMu.Unlock()
		}
	}

	return allErrors
}

// DynamicRoleReassignment reassigns roles based on current performance.
func (gm *GraphMeshTopology) DynamicRoleReassignment(ctx context.Context) error {
	if !gm.config.EnableDynamicRoles {
		return nil
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Analyze agent performance
	type agentPerformance struct {
		agent        *Agent
		score        float64
		responseTime time.Duration
		messageCount int
	}

	performances := make([]agentPerformance, 0, len(gm.agents))
	for _, agent := range gm.agents {
		metrics := agent.GetMetrics()
		performances = append(performances, agentPerformance{
			agent:        agent,
			score:        agent.Score,
			responseTime: metrics.AvgResponseTime,
			messageCount: metrics.MessageCount,
		})
	}

	// Sort by combined performance score
	sort.Slice(performances, func(i, j int) bool {
		// Higher score, faster response, more messages = better
		scoreI := performances[i].score - float64(performances[i].responseTime.Milliseconds())/1000 + float64(performances[i].messageCount)*0.1
		scoreJ := performances[j].score - float64(performances[j].responseTime.Milliseconds())/1000 + float64(performances[j].messageCount)*0.1
		return scoreI > scoreJ
	})

	// Top performers can become reviewers/validators
	// Bottom performers become proposers/critics (lower responsibility)
	n := len(performances)
	if n >= 4 {
		// Top 25% -> validators/moderators
		topQuartile := n / 4
		for i := 0; i < topQuartile; i++ {
			agent := performances[i].agent
			if agent.Role != RoleValidator && agent.Role != RoleModerator {
				// Only reassign if not already in a leader role
				gm.assignRoleInternal(agent.ID, RoleValidator)
			}
		}
	}

	return nil
}

// assignRoleInternal assigns a role without locking (caller must hold lock).
func (gm *GraphMeshTopology) assignRoleInternal(agentID string, role AgentRole) {
	agent, ok := gm.agents[agentID]
	if !ok {
		return
	}

	// Remove from old role list
	oldRoleAgents := gm.agentsByRole[agent.Role]
	for i, a := range oldRoleAgents {
		if a.ID == agentID {
			gm.agentsByRole[agent.Role] = append(oldRoleAgents[:i], oldRoleAgents[i+1:]...)
			break
		}
	}

	// Assign new role
	agent.Role = role
	gm.agentsByRole[role] = append(gm.agentsByRole[role], agent)
}

// GetTopologySnapshot returns a snapshot of the current topology state.
func (gm *GraphMeshTopology) GetTopologySnapshot() *TopologySnapshot {
	gm.mu.RLock()
	gm.phaseMu.RLock()
	gm.leaderMu.RLock()
	defer gm.mu.RUnlock()
	defer gm.phaseMu.RUnlock()
	defer gm.leaderMu.RUnlock()

	agents := make([]*Agent, 0, len(gm.agents))
	for _, agent := range gm.agents {
		agents = append(agents, agent)
	}

	channels := make([]CommunicationChannel, len(gm.channels))
	copy(channels, gm.channels)

	var leaderID string
	if gm.currentLeader != nil {
		leaderID = gm.currentLeader.ID
	}

	return &TopologySnapshot{
		Type:          TopologyGraphMesh,
		Agents:        agents,
		Channels:      channels,
		CurrentPhase:  gm.currentPhase,
		CurrentLeader: leaderID,
		Metrics:       gm.GetMetrics(),
		Timestamp:     time.Now(),
	}
}

// TopologySnapshot represents a point-in-time snapshot of the topology.
type TopologySnapshot struct {
	Type          TopologyType           `json:"type"`
	Agents        []*Agent               `json:"agents"`
	Channels      []CommunicationChannel `json:"channels"`
	CurrentPhase  DebatePhase            `json:"current_phase"`
	CurrentLeader string                 `json:"current_leader"`
	Metrics       TopologyMetrics        `json:"metrics"`
	Timestamp     time.Time              `json:"timestamp"`
}

// Close closes the Graph-Mesh topology.
func (gm *GraphMeshTopology) Close() error {
	close(gm.messageQueue)
	return gm.BaseTopology.Close()
}
