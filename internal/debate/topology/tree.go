// Package topology provides Tree topology implementation.
// Tree topology organizes agents in a hierarchical parent-child structure.
// The root (Architect) delegates to team leads, who delegate to specialists.
// Balanced delegation with natural parallelism across independent subtrees.
package topology

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TreeNode represents a node in the hierarchical tree topology.
type TreeNode struct {
	AgentID  string
	Role     AgentRole
	Parent   *TreeNode
	Children []*TreeNode
	Level    int
	Subtree  string // responsibility area (e.g., "security", "performance")
}

// TreeTopology implements hierarchical parent-child coordination.
// Agents are organized in a tree where the root delegates to team leads,
// who in turn delegate to specialist agents. Communication flows up and
// down the tree, enabling natural task decomposition and parallel execution
// across independent subtrees.
type TreeTopology struct {
	*BaseTopology

	root     *TreeNode
	nodes    map[string]*TreeNode // agentID -> node
	maxDepth int

	// Message queue for tree routing
	messageQueue chan *Message

	nodesMu sync.RWMutex
}

// NewTreeTopology creates a new Tree topology.
func NewTreeTopology(config TopologyConfig) *TreeTopology {
	config.Type = TopologyTree

	tt := &TreeTopology{
		BaseTopology: NewBaseTopology(config),
		nodes:        make(map[string]*TreeNode),
		maxDepth:     3,
		messageQueue: make(chan *Message, 500),
	}

	return tt
}

// GetType returns the topology type.
func (t *TreeTopology) GetType() TopologyType {
	return TopologyTree
}

// Initialize sets up the Tree topology with agents.
func (t *TreeTopology) Initialize(ctx context.Context, agents []*Agent) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(agents) == 0 {
		return fmt.Errorf("at least one agent required for Tree topology")
	}

	// Add all agents to the base topology
	for _, agent := range agents {
		t.agents[agent.ID] = agent
		t.agentsByRole[agent.Role] = append(t.agentsByRole[agent.Role], agent)
	}

	// Build the hierarchical tree structure
	t.buildTree()

	// Establish bidirectional parent-child communication channels
	t.nodesMu.RLock()
	for _, node := range t.nodes {
		for _, child := range node.Children {
			t.channels = append(t.channels, CommunicationChannel{
				FromAgent:     node.AgentID,
				ToAgent:       child.AgentID,
				Bidirectional: true,
				Weight:        1.0,
			})
		}
	}
	t.nodesMu.RUnlock()

	// Start message processor
	go t.processMessages(ctx)

	return nil
}

// buildTree creates the hierarchical tree from available agents.
func (t *TreeTopology) buildTree() {
	t.nodesMu.Lock()
	defer t.nodesMu.Unlock()

	// Clear any existing tree state
	t.nodes = make(map[string]*TreeNode)
	t.root = nil

	// Collect all agents into a sortable slice
	allAgents := make([]*Agent, 0, len(t.agents))
	for _, agent := range t.agents {
		allAgents = append(allAgents, agent)
	}

	if len(allAgents) == 0 {
		return
	}

	// --- Step 1: Select the root agent ---
	// Prefer RoleArchitect or RoleDesigner; fallback to highest score
	var rootAgent *Agent
	for _, agent := range allAgents {
		if agent.Role == RoleArchitect {
			if rootAgent == nil || agent.Score > rootAgent.Score {
				rootAgent = agent
			}
		}
	}
	if rootAgent == nil {
		for _, agent := range allAgents {
			if agent.Role == RoleDesigner {
				if rootAgent == nil || agent.Score > rootAgent.Score {
					rootAgent = agent
				}
			}
		}
	}
	if rootAgent == nil {
		// Fallback: pick the highest scored agent
		sort.Slice(allAgents, func(i, j int) bool {
			return allAgents[i].Score > allAgents[j].Score
		})
		rootAgent = allAgents[0]
	}

	// Create root node at level 0
	t.root = &TreeNode{
		AgentID:  rootAgent.ID,
		Role:     rootAgent.Role,
		Parent:   nil,
		Children: make([]*TreeNode, 0),
		Level:    0,
		Subtree:  "root",
	}
	t.nodes[rootAgent.ID] = t.root

	// Build remaining agents list (excluding root)
	remaining := make([]*Agent, 0, len(allAgents)-1)
	for _, agent := range allAgents {
		if agent.ID != rootAgent.ID {
			remaining = append(remaining, agent)
		}
	}

	if len(remaining) == 0 {
		return
	}

	// --- Step 2: Find team leads (level 1) ---
	leadRoles := map[AgentRole]string{
		RoleSecurity:            "security",
		RolePerformanceAnalyzer: "performance",
		RoleModerator:           "coordination",
		RoleCritic:              "quality",
		RoleReviewer:            "review",
	}

	leads := make([]*TreeNode, 0)
	leadIDs := make(map[string]bool)

	// Find natural leads by role
	for _, agent := range remaining {
		if subtree, isLeadRole := leadRoles[agent.Role]; isLeadRole {
			// Avoid duplicate subtree responsibilities
			alreadyCovered := false
			for _, existing := range leads {
				if existing.Subtree == subtree {
					alreadyCovered = true
					break
				}
			}
			if !alreadyCovered {
				node := &TreeNode{
					AgentID:  agent.ID,
					Role:     agent.Role,
					Parent:   t.root,
					Children: make([]*TreeNode, 0),
					Level:    1,
					Subtree:  subtree,
				}
				leads = append(leads, node)
				leadIDs[agent.ID] = true
				t.nodes[agent.ID] = node
				t.root.Children = append(t.root.Children, node)
			}
		}
	}

	// If no natural leads found, pick top-scored non-root agents (max 3)
	if len(leads) == 0 {
		sort.Slice(remaining, func(i, j int) bool {
			return remaining[i].Score > remaining[j].Score
		})

		maxLeads := 3
		if len(remaining) < maxLeads {
			maxLeads = len(remaining)
		}

		subtreeNames := []string{"alpha", "beta", "gamma"}
		for i := 0; i < maxLeads; i++ {
			agent := remaining[i]
			node := &TreeNode{
				AgentID:  agent.ID,
				Role:     agent.Role,
				Parent:   t.root,
				Children: make([]*TreeNode, 0),
				Level:    1,
				Subtree:  subtreeNames[i],
			}
			leads = append(leads, node)
			leadIDs[agent.ID] = true
			t.nodes[agent.ID] = node
			t.root.Children = append(t.root.Children, node)
		}
	}

	// --- Step 3: Distribute remaining agents as children (level 2, round-robin) ---
	childAgents := make([]*Agent, 0)
	for _, agent := range remaining {
		if !leadIDs[agent.ID] {
			childAgents = append(childAgents, agent)
		}
	}

	if len(leads) > 0 {
		for i, agent := range childAgents {
			leadIdx := i % len(leads)
			parentLead := leads[leadIdx]

			node := &TreeNode{
				AgentID:  agent.ID,
				Role:     agent.Role,
				Parent:   parentLead,
				Children: make([]*TreeNode, 0),
				Level:    2,
				Subtree:  parentLead.Subtree,
			}
			parentLead.Children = append(parentLead.Children, node)
			t.nodes[agent.ID] = node
		}
	}
}

// RouteMessage determines routing for a message through the tree hierarchy.
// Routing rules:
//   - Child of sender: route down directly
//   - Parent of sender: route up directly
//   - Sibling: route through common parent
//   - Otherwise: route up to root, then down to target's subtree
func (t *TreeTopology) RouteMessage(msg *Message) ([]*Agent, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	_, senderExists := t.agents[msg.FromAgent]
	if !senderExists {
		return nil, fmt.Errorf("%w: sender %s", ErrAgentNotFound, msg.FromAgent)
	}

	// If no specific targets, treat as broadcast
	if len(msg.ToAgents) == 0 {
		targets := make([]*Agent, 0, len(t.agents)-1)
		for id, agent := range t.agents {
			if id != msg.FromAgent {
				targets = append(targets, agent)
			}
		}
		return targets, nil
	}

	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	senderNode, senderInTree := t.nodes[msg.FromAgent]

	targets := make([]*Agent, 0, len(msg.ToAgents))

	for _, targetID := range msg.ToAgents {
		targetAgent, targetExists := t.agents[targetID]
		if !targetExists {
			continue
		}

		targetNode, targetInTree := t.nodes[targetID]

		// If either node isn't in the tree, route via root
		if !senderInTree || !targetInTree {
			targets = append(targets, targetAgent)
			continue
		}

		// Direct child of sender: route down directly
		if t.isChild(senderNode, targetNode) {
			targets = append(targets, targetAgent)
			continue
		}

		// Parent of sender: route up directly
		if t.isChild(targetNode, senderNode) {
			targets = append(targets, targetAgent)
			continue
		}

		// Sibling: route through common parent
		if senderNode.Parent != nil && targetNode.Parent != nil &&
			senderNode.Parent.AgentID == targetNode.Parent.AgentID {
			parentAgent, ok := t.agents[senderNode.Parent.AgentID]
			if ok {
				targets = append(targets, parentAgent)
			}
			continue
		}

		// Otherwise: route up to root, then down to target's subtree
		if t.root != nil {
			rootAgent, ok := t.agents[t.root.AgentID]
			if ok && rootAgent.ID != msg.FromAgent {
				targets = append(targets, rootAgent)
			} else {
				// Sender is root, route directly to target
				targets = append(targets, targetAgent)
			}
		}
	}

	// Deduplicate targets
	targets = t.deduplicateAgents(targets)

	if len(targets) == 0 {
		return nil, ErrRoutingFailed
	}

	return targets, nil
}

// isChild checks if childNode is a direct child of parentNode.
func (t *TreeTopology) isChild(parentNode, childNode *TreeNode) bool {
	for _, c := range parentNode.Children {
		if c.AgentID == childNode.AgentID {
			return true
		}
	}
	return false
}

// deduplicateAgents removes duplicate agents from a slice.
func (t *TreeTopology) deduplicateAgents(agents []*Agent) []*Agent {
	seen := make(map[string]bool, len(agents))
	result := make([]*Agent, 0, len(agents))
	for _, agent := range agents {
		if !seen[agent.ID] {
			seen[agent.ID] = true
			result = append(result, agent)
		}
	}
	return result
}

// BroadcastMessage sends a message from root down through all levels.
func (t *TreeTopology) BroadcastMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	t.IncrementMessageCount(true)

	t.nodesMu.RLock()
	root := t.root
	t.nodesMu.RUnlock()

	if root == nil {
		return fmt.Errorf("tree has no root node")
	}

	// Breadth-first broadcast from root down
	visited := make(map[string]bool)
	visited[msg.FromAgent] = true // Don't send back to sender

	queue := []*TreeNode{root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.AgentID] {
			// Still traverse children
			for _, child := range current.Children {
				queue = append(queue, child)
			}
			continue
		}
		visited[current.AgentID] = true

		agent, ok := t.GetAgent(current.AgentID)
		if !ok {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case t.messageQueue <- msg:
			agent.UpdateActivity(0)
		case <-time.After(t.config.MessageTimeout):
			return fmt.Errorf("broadcast timeout to %s", current.AgentID)
		}

		for _, child := range current.Children {
			queue = append(queue, child)
		}
	}

	return nil
}

// SendMessage sends a message through the tree topology.
func (t *TreeTopology) SendMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	if len(msg.ToAgents) == 0 {
		return t.BroadcastMessage(ctx, msg)
	}

	t.IncrementMessageCount(false)

	targets, err := t.RouteMessage(msg)
	if err != nil {
		return err
	}

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case t.messageQueue <- msg:
			target.UpdateActivity(0)
		case <-time.After(t.config.MessageTimeout):
			return fmt.Errorf("message timeout to %s", target.ID)
		}
	}

	return nil
}

// processMessages processes messages flowing through the tree.
func (t *TreeTopology) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-t.messageQueue:
			if !ok {
				return
			}
			// Message processed by the tree routing layer;
			// actual handler invocation is done by the orchestrator.
		}
	}
}

// CanCommunicate checks if two agents can communicate directly.
// In Tree topology, direct communication is allowed between parent and child.
func (t *TreeTopology) CanCommunicate(fromID, toID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	_, fromExists := t.agents[fromID]
	_, toExists := t.agents[toID]
	if !fromExists || !toExists || fromID == toID {
		return false
	}

	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	fromNode, fromInTree := t.nodes[fromID]
	toNode, toInTree := t.nodes[toID]

	if !fromInTree || !toInTree {
		return false
	}

	// Parent-child relationship (either direction)
	return t.isChild(fromNode, toNode) || t.isChild(toNode, fromNode)
}

// GetCommunicationTargets returns the parent and children of the node.
// In Tree topology, a node can directly communicate with its parent
// and its children (no siblings directly).
func (t *TreeTopology) GetCommunicationTargets(agentID string) []*Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	node, ok := t.nodes[agentID]
	if !ok {
		return []*Agent{}
	}

	targets := make([]*Agent, 0)

	// Add parent
	if node.Parent != nil {
		if parent, exists := t.agents[node.Parent.AgentID]; exists {
			targets = append(targets, parent)
		}
	}

	// Add children
	for _, child := range node.Children {
		if childAgent, exists := t.agents[child.AgentID]; exists {
			targets = append(targets, childAgent)
		}
	}

	return targets
}

// SelectLeader dynamically selects a leader for the current phase.
// The tree topology leverages the hierarchy for leader selection:
//   - Dehallucination/Convergence/Proposal/Review: root (Architect)
//   - Critique/Adversarial: security lead if exists
//   - Optimization: performance lead if exists
func (t *TreeTopology) SelectLeader(phase DebatePhase) (*Agent, error) {
	t.nodesMu.RLock()
	root := t.root
	t.nodesMu.RUnlock()

	if root == nil {
		return nil, fmt.Errorf("tree has no root node")
	}

	switch phase {
	case PhaseCritique, PhaseAdversarial:
		// Security lead if exists
		lead := t.findLeadBySubtree("security")
		if lead != nil {
			if agent, ok := t.GetAgent(lead.AgentID); ok {
				return agent, nil
			}
		}

	case PhaseOptimization:
		// Performance lead if exists
		lead := t.findLeadBySubtree("performance")
		if lead != nil {
			if agent, ok := t.GetAgent(lead.AgentID); ok {
				return agent, nil
			}
		}
	}

	// Default: root (Architect) for Dehallucination, Convergence,
	// Proposal, Review, SelfEvolvement, and fallback
	rootAgent, ok := t.GetAgent(root.AgentID)
	if !ok {
		return nil, fmt.Errorf("root agent %s not found", root.AgentID)
	}
	return rootAgent, nil
}

// findLeadBySubtree finds a level-1 lead node by subtree name.
func (t *TreeTopology) findLeadBySubtree(subtree string) *TreeNode {
	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	if t.root == nil {
		return nil
	}

	for _, child := range t.root.Children {
		if child.Subtree == subtree {
			return child
		}
	}
	return nil
}

// GetParallelGroups returns each lead's subtree (lead + children) as
// independent groups that can execute concurrently since they reside
// in different subtrees of the hierarchy.
func (t *TreeTopology) GetParallelGroups(phase DebatePhase) [][]*Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	if t.root == nil {
		return nil
	}

	groups := make([][]*Agent, 0, len(t.root.Children))

	for _, lead := range t.root.Children {
		group := make([]*Agent, 0, 1+len(lead.Children))

		// Add the lead itself
		if leadAgent, ok := t.agents[lead.AgentID]; ok {
			group = append(group, leadAgent)
		}

		// Add all children of this lead
		for _, child := range lead.Children {
			if childAgent, ok := t.agents[child.AgentID]; ok {
				group = append(group, childAgent)
			}
		}

		if len(group) > 0 {
			groups = append(groups, group)
		}
	}

	// If no subtree groups exist (root is the only node), return root as a group
	if len(groups) == 0 {
		if rootAgent, ok := t.agents[t.root.AgentID]; ok {
			groups = append(groups, []*Agent{rootAgent})
		}
	}

	return groups
}

// Rebalance handles the failure of an agent by redistributing its
// children to sibling nodes and removing the failed node from the tree.
func (t *TreeTopology) Rebalance(failedAgentID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.nodesMu.Lock()
	defer t.nodesMu.Unlock()

	failedNode, ok := t.nodes[failedAgentID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, failedAgentID)
	}

	// Cannot remove root without promoting a replacement
	if failedNode.Parent == nil {
		// Promote first child as new root if available
		if len(failedNode.Children) > 0 {
			newRoot := failedNode.Children[0]
			newRoot.Parent = nil
			newRoot.Level = 0
			newRoot.Subtree = "root"

			// Remaining children of failed root become children of new root
			for _, child := range failedNode.Children[1:] {
				child.Parent = newRoot
				newRoot.Children = append(newRoot.Children, child)
			}

			t.root = newRoot
		} else {
			t.root = nil
		}

		delete(t.nodes, failedAgentID)
		delete(t.agents, failedAgentID)
		t.removeFromRoleMap(failedAgentID)
		t.rebuildChannels()
		return nil
	}

	parent := failedNode.Parent

	// Remove failed node from parent's children
	for i, child := range parent.Children {
		if child.AgentID == failedAgentID {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			break
		}
	}

	// Redistribute orphaned children to sibling nodes (round-robin)
	siblings := parent.Children
	if len(siblings) > 0 && len(failedNode.Children) > 0 {
		for i, orphan := range failedNode.Children {
			siblingIdx := i % len(siblings)
			targetSibling := siblings[siblingIdx]

			orphan.Parent = targetSibling
			orphan.Subtree = targetSibling.Subtree
			targetSibling.Children = append(targetSibling.Children, orphan)
		}
	} else if len(failedNode.Children) > 0 {
		// No siblings: attach orphans directly to the parent
		for _, orphan := range failedNode.Children {
			orphan.Parent = parent
			orphan.Level = failedNode.Level
			parent.Children = append(parent.Children, orphan)
		}
	}

	// Remove the failed node
	delete(t.nodes, failedAgentID)
	delete(t.agents, failedAgentID)
	t.removeFromRoleMap(failedAgentID)

	// Rebuild communication channels to reflect new structure
	t.rebuildChannels()

	return nil
}

// removeFromRoleMap removes an agent from the agentsByRole map.
// Caller must hold t.mu.
func (t *TreeTopology) removeFromRoleMap(agentID string) {
	for role, agents := range t.agentsByRole {
		for i, a := range agents {
			if a.ID == agentID {
				t.agentsByRole[role] = append(agents[:i], agents[i+1:]...)
				break
			}
		}
	}
}

// rebuildChannels rebuilds all communication channels from the current tree structure.
// Caller must hold both t.mu and t.nodesMu.
func (t *TreeTopology) rebuildChannels() {
	t.channels = t.channels[:0]

	for _, node := range t.nodes {
		for _, child := range node.Children {
			t.channels = append(t.channels, CommunicationChannel{
				FromAgent:     node.AgentID,
				ToAgent:       child.AgentID,
				Bidirectional: true,
				Weight:        1.0,
			})
		}
	}
}

// GetRoot returns the root node of the tree.
func (t *TreeTopology) GetRoot() *TreeNode {
	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()
	return t.root
}

// GetNode returns a tree node by agent ID.
func (t *TreeTopology) GetNode(agentID string) (*TreeNode, bool) {
	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	node, ok := t.nodes[agentID]
	return node, ok
}

// GetDepth returns the maximum depth of the tree.
func (t *TreeTopology) GetDepth() int {
	t.nodesMu.RLock()
	defer t.nodesMu.RUnlock()

	maxLevel := 0
	for _, node := range t.nodes {
		if node.Level > maxLevel {
			maxLevel = node.Level
		}
	}
	return maxLevel
}

// Close cleans up tree topology resources.
func (t *TreeTopology) Close() error {
	close(t.messageQueue)
	return t.BaseTopology.Close()
}
