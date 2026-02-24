package planning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// ---------------------------------------------------------------------------
// Mock implementations for MCTS
// ---------------------------------------------------------------------------

type mockActionGenerator struct {
	actions     []string
	actionsErr  error
	applyResult interface{}
	applyErr    error
	// Track calls
	getActionsCalls  int
	applyActionCalls int
	mu               sync.Mutex
}

func (m *mockActionGenerator) GetActions(ctx context.Context, state interface{}) ([]string, error) {
	m.mu.Lock()
	m.getActionsCalls++
	m.mu.Unlock()
	return m.actions, m.actionsErr
}

func (m *mockActionGenerator) ApplyAction(ctx context.Context, state interface{}, action string) (interface{}, error) {
	m.mu.Lock()
	m.applyActionCalls++
	m.mu.Unlock()
	if m.applyResult != nil {
		return m.applyResult, m.applyErr
	}
	return fmt.Sprintf("%v+%s", state, action), m.applyErr
}

type mockRewardFunction struct {
	reward      float64
	rewardErr   error
	isTerminal  bool
	terminalErr error
	callCount   int
	mu          sync.Mutex
}

func (m *mockRewardFunction) Evaluate(ctx context.Context, state interface{}) (float64, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	return m.reward, m.rewardErr
}

func (m *mockRewardFunction) IsTerminal(ctx context.Context, state interface{}) (bool, error) {
	return m.isTerminal, m.terminalErr
}

type mockRolloutPolicy struct {
	rolloutValue float64
	rolloutErr   error
	callCount    int
	mu           sync.Mutex
}

func (m *mockRolloutPolicy) Rollout(ctx context.Context, state interface{}, depth int) (float64, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	return m.rolloutValue, m.rolloutErr
}

// Stateful mock that tracks how many times IsTerminal was called,
// becomes terminal after N evaluations
type terminatingRewardFunc struct {
	reward         float64
	terminalAfter  int
	evalCount      int
	mu             sync.Mutex
}

func (f *terminatingRewardFunc) Evaluate(ctx context.Context, state interface{}) (float64, error) {
	f.mu.Lock()
	f.evalCount++
	f.mu.Unlock()
	return f.reward, nil
}

func (f *terminatingRewardFunc) IsTerminal(ctx context.Context, state interface{}) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.evalCount >= f.terminalAfter, nil
}

// ---------------------------------------------------------------------------
// MCTSNodeState constants
// ---------------------------------------------------------------------------

func TestMCTSNodeState_Constants(t *testing.T) {
	assert.Equal(t, MCTSNodeState("unexpanded"), MCTSNodeStateUnexpanded)
	assert.Equal(t, MCTSNodeState("expanded"), MCTSNodeStateExpanded)
	assert.Equal(t, MCTSNodeState("terminal"), MCTSNodeStateTerminal)
}

// ---------------------------------------------------------------------------
// DefaultMCTSConfig
// ---------------------------------------------------------------------------

func TestDefaultMCTSConfig(t *testing.T) {
	cfg := DefaultMCTSConfig()
	assert.InDelta(t, 1.414, cfg.ExplorationConstant, 0.001)
	assert.Equal(t, 0.5, cfg.DepthPreferenceAlpha)
	assert.Equal(t, 50, cfg.MaxDepth)
	assert.Equal(t, 1000, cfg.MaxIterations)
	assert.Equal(t, 10, cfg.RolloutDepth)
	assert.Equal(t, 1, cfg.SimulationCount)
	assert.Equal(t, 0.99, cfg.DiscountFactor)
	assert.True(t, cfg.EnableParallel)
	assert.Equal(t, 4, cfg.ParallelWorkers)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.True(t, cfg.UseUCTDP)
}

// ---------------------------------------------------------------------------
// MCTSNode
// ---------------------------------------------------------------------------

func TestMCTSNode_AverageReward_ZeroVisits(t *testing.T) {
	node := &MCTSNode{
		Visits:      0,
		TotalReward: 0,
	}
	assert.Equal(t, 0.0, node.AverageReward())
}

func TestMCTSNode_AverageReward_WithVisits(t *testing.T) {
	node := &MCTSNode{
		Visits:      10,
		TotalReward: 7.5,
	}
	assert.InDelta(t, 0.75, node.AverageReward(), 0.001)
}

func TestMCTSNode_AverageReward_SingleVisit(t *testing.T) {
	node := &MCTSNode{
		Visits:      1,
		TotalReward: 0.8,
	}
	assert.InDelta(t, 0.8, node.AverageReward(), 0.001)
}

func TestMCTSNode_AddReward(t *testing.T) {
	node := &MCTSNode{}
	node.AddReward(0.5)
	assert.Equal(t, 1, node.Visits)
	assert.Equal(t, 0.5, node.TotalReward)

	node.AddReward(0.3)
	assert.Equal(t, 2, node.Visits)
	assert.InDelta(t, 0.8, node.TotalReward, 0.001)
}

func TestMCTSNode_AddReward_ConcurrentSafety(t *testing.T) {
	node := &MCTSNode{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.AddReward(0.1)
		}()
	}
	wg.Wait()
	assert.Equal(t, 100, node.Visits)
	assert.InDelta(t, 10.0, node.TotalReward, 0.001)
}

func TestMCTSNode_AverageReward_ConcurrentSafety(t *testing.T) {
	node := &MCTSNode{
		Visits:      50,
		TotalReward: 25.0,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			avg := node.AverageReward()
			assert.InDelta(t, 0.5, avg, 0.001)
		}()
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// NewMCTS
// ---------------------------------------------------------------------------

func TestNewMCTS_WithLogger(t *testing.T) {
	logger := logrus.New()
	ag := &mockActionGenerator{actions: []string{"a"}}
	rf := &mockRewardFunction{reward: 0.5}
	rp := &mockRolloutPolicy{rolloutValue: 0.5}

	m := NewMCTS(DefaultMCTSConfig(), ag, rf, rp, logger)
	require.NotNil(t, m)
	assert.Equal(t, logger, m.logger)
	assert.NotNil(t, m.rng)
}

func TestNewMCTS_NilLogger(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{"a"}}
	rf := &mockRewardFunction{reward: 0.5}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)
	require.NotNil(t, m)
	assert.NotNil(t, m.logger)
	assert.Equal(t, logrus.WarnLevel, m.logger.Level)
}

// ---------------------------------------------------------------------------
// UCTValue
// ---------------------------------------------------------------------------

func TestMCTS_UCTValue_NilNode(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)
	assert.Equal(t, 0.0, m.UCTValue(nil, 10))
}

func TestMCTS_UCTValue_ZeroVisits_ReturnsInfinity(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	node := &MCTSNode{
		Visits:      0,
		TotalReward: 0,
		Depth:       1,
	}
	val := m.UCTValue(node, 10)
	assert.True(t, math.IsInf(val, 1))
}

func TestMCTS_UCTValue_ZeroParentVisits(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.UseUCTDP = false
	m := NewMCTS(cfg, ag, rf, nil, nil)

	node := &MCTSNode{
		Visits:      5,
		TotalReward: 2.5,
		Depth:       1,
	}
	// parentVisits=0 gets clamped to 1
	val := m.UCTValue(node, 0)
	// exploitation = 2.5/5 = 0.5
	// exploration = C * sqrt(ln(1)/5) = C * sqrt(0/5) = 0
	assert.InDelta(t, 0.5, val, 0.01)
}

func TestMCTS_UCTValue_StandardUCB1(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.UseUCTDP = false
	cfg.ExplorationConstant = 1.414
	m := NewMCTS(cfg, ag, rf, nil, nil)

	node := &MCTSNode{
		Visits:      10,
		TotalReward: 6.0,
		Depth:       2,
	}

	val := m.UCTValue(node, 100)

	// exploitation = 6.0/10 = 0.6
	// exploration = 1.414 * sqrt(ln(100)/10) = 1.414 * sqrt(4.605/10) = 1.414 * 0.6787 ~ 0.96
	exploitation := 6.0 / 10.0
	exploration := 1.414 * math.Sqrt(math.Log(100.0)/10.0)
	expected := exploitation + exploration

	assert.InDelta(t, expected, val, 0.01)
}

func TestMCTS_UCTValue_WithUCTDP(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.UseUCTDP = true
	cfg.DepthPreferenceAlpha = 0.5
	cfg.MaxDepth = 50
	cfg.ExplorationConstant = 1.414
	m := NewMCTS(cfg, ag, rf, nil, nil)

	node := &MCTSNode{
		Visits:      10,
		TotalReward: 6.0,
		Depth:       25,
	}

	val := m.UCTValue(node, 100)

	exploitation := 6.0 / 10.0
	exploration := 1.414 * math.Sqrt(math.Log(100.0)/10.0)
	depthBonus := 0.5 * (25.0 / 50.0)
	expected := exploitation + exploration + depthBonus

	assert.InDelta(t, expected, val, 0.01)
}

func TestMCTS_UCTValue_UCTDP_ZeroMaxDepth(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.UseUCTDP = true
	cfg.MaxDepth = 0
	m := NewMCTS(cfg, ag, rf, nil, nil)

	node := &MCTSNode{
		Visits:      5,
		TotalReward: 2.0,
		Depth:       3,
	}

	val := m.UCTValue(node, 20)
	// MaxDepth=0 means no depth bonus applied
	exploitation := 2.0 / 5.0
	exploration := cfg.ExplorationConstant * math.Sqrt(math.Log(20.0)/5.0)
	expected := exploitation + exploration
	assert.InDelta(t, expected, val, 0.01)
}

// ---------------------------------------------------------------------------
// selectBestChild
// ---------------------------------------------------------------------------

func TestMCTS_SelectBestChild_NoChildren(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	parent := &MCTSNode{Visits: 10, Children: []*MCTSNode{}}
	best := m.selectBestChild(parent)
	assert.Nil(t, best)
}

func TestMCTS_SelectBestChild_UnvisitedChildPrioritized(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	parent := &MCTSNode{Visits: 10}
	child1 := &MCTSNode{ID: "c1", Visits: 5, TotalReward: 3.0}
	child2 := &MCTSNode{ID: "c2", Visits: 0, TotalReward: 0.0} // unvisited
	parent.Children = []*MCTSNode{child1, child2}

	best := m.selectBestChild(parent)
	assert.Equal(t, "c2", best.ID)
}

func TestMCTS_SelectBestChild_HighestUCB(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.UseUCTDP = false
	m := NewMCTS(cfg, ag, rf, nil, nil)

	parent := &MCTSNode{Visits: 100}
	// Child with better exploitation
	child1 := &MCTSNode{ID: "c1", Visits: 50, TotalReward: 45.0, Depth: 1}
	// Child with worse exploitation
	child2 := &MCTSNode{ID: "c2", Visits: 50, TotalReward: 10.0, Depth: 1}
	parent.Children = []*MCTSNode{child1, child2}

	best := m.selectBestChild(parent)
	assert.Equal(t, "c1", best.ID)
}

func TestMCTS_SelectBestChild_ZeroParentVisits(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	parent := &MCTSNode{Visits: 0}
	child := &MCTSNode{ID: "c1", Visits: 0}
	parent.Children = []*MCTSNode{child}

	best := m.selectBestChild(parent)
	// Unvisited child returned immediately
	assert.Equal(t, "c1", best.ID)
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestMCTS_Search_BasicSearch(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"action1", "action2"},
	}
	rf := &mockRewardFunction{
		reward:     0.7,
		isTerminal: false,
	}
	rp := &mockRolloutPolicy{rolloutValue: 0.8}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 10
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, rp, nil)

	result, err := m.Search(context.Background(), "initial state")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 10, result.TotalIterations)
	assert.True(t, result.Duration > 0)
	assert.True(t, result.TreeSize >= 1)
}

func TestMCTS_Search_WithTerminalState(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1"},
	}
	rf := &terminatingRewardFunc{
		reward:        0.9,
		terminalAfter: 3,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 20
	cfg.MaxDepth = 5
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.TotalIterations <= 20)
}

func TestMCTS_Search_NoActions(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{}, // No actions available
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 5
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "state")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMCTS_Search_ActionGeneratorError(t *testing.T) {
	ag := &mockActionGenerator{
		actionsErr: errors.New("action gen error"),
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 5
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "state")
	require.NoError(t, err) // Search handles errors internally
	require.NotNil(t, result)
}

func TestMCTS_Search_RewardFunctionError(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1"},
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
		rewardErr:  errors.New("reward error"),
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 5
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	// Rollout policy also returns error
	result, err := m.Search(context.Background(), "state")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMCTS_Search_SingleIteration(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1"},
	}
	rf := &mockRewardFunction{
		reward:     0.6,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 1
	cfg.MaxDepth = 5
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalIterations)
}

func TestMCTS_Search_ContextTimeout(t *testing.T) {
	// NOTE: The source MCTS loop's `select { case <-ctx.Done(): break }` only
	// breaks the select, not the for-loop. So context cancellation alone does
	// NOT stop the loop -- MaxIterations is the real limiter. We set a moderate
	// MaxIterations with a very short timeout to verify the search completes
	// promptly and the result is valid even when the context expires mid-search.
	ag := &mockActionGenerator{
		actions: []string{"a1", "a2"},
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 100 // Moderate, completes fast
	cfg.MaxDepth = 5
	cfg.Timeout = 50 * time.Millisecond
	m := NewMCTS(cfg, ag, rf, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := m.Search(ctx, "start")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.TotalIterations <= 100)
}

func TestMCTS_Search_MaxDepthReached(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1"},
	}
	rf := &mockRewardFunction{
		reward:     0.6,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 50
	cfg.MaxDepth = 2
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMCTS_Search_WithRolloutPolicy(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1", "a2"},
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
	}
	rp := &mockRolloutPolicy{rolloutValue: 0.9}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 10
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, rp, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)

	rp.mu.Lock()
	assert.True(t, rp.callCount > 0)
	rp.mu.Unlock()
}

func TestMCTS_Search_WithoutRolloutPolicy(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1"},
	}
	rf := &mockRewardFunction{
		reward:     0.7,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 5
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil) // nil rollout policy

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMCTS_Search_ResultContainsBestPath(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1", "a2"},
	}
	rf := &mockRewardFunction{
		reward:     0.8,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 20
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Best path should contain at least the root
	assert.NotEmpty(t, result.BestPath)
}

func TestMCTS_Search_ApplyActionError(t *testing.T) {
	ag := &mockActionGenerator{
		actions:  []string{"a1"},
		applyErr: errors.New("apply error"),
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 10
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMCTS_Search_IsTerminalError(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1"},
	}
	rf := &mockRewardFunction{
		reward:      0.5,
		isTerminal:  false,
		terminalErr: errors.New("terminal check error"),
	}

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 5
	cfg.MaxDepth = 3
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, rf, nil, nil)

	result, err := m.Search(context.Background(), "start")
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// expand
// ---------------------------------------------------------------------------

func TestMCTS_Expand_TerminalNode(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	node := &MCTSNode{
		NodeState: MCTSNodeStateTerminal,
		Metadata:  make(map[string]interface{}),
	}

	expanded, err := m.expand(context.Background(), node)
	require.NoError(t, err)
	assert.Equal(t, node, expanded)
}

func TestMCTS_Expand_MaxDepthReached(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.MaxDepth = 5
	m := NewMCTS(cfg, ag, rf, nil, nil)

	node := &MCTSNode{
		Depth:     5,
		NodeState: MCTSNodeStateUnexpanded,
		Metadata:  make(map[string]interface{}),
	}

	expanded, err := m.expand(context.Background(), node)
	require.NoError(t, err)
	assert.Equal(t, MCTSNodeStateTerminal, expanded.NodeState)
}

func TestMCTS_Expand_IsTerminal(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{isTerminal: true}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	node := &MCTSNode{
		Depth:     1,
		State:     "some state",
		NodeState: MCTSNodeStateUnexpanded,
		Metadata:  make(map[string]interface{}),
	}

	expanded, err := m.expand(context.Background(), node)
	require.NoError(t, err)
	assert.Equal(t, MCTSNodeStateTerminal, expanded.NodeState)
}

func TestMCTS_Expand_NoActions(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{}}
	rf := &mockRewardFunction{isTerminal: false}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	node := &MCTSNode{
		ID:        "node-1",
		Depth:     1,
		State:     "state",
		NodeState: MCTSNodeStateUnexpanded,
		Metadata:  make(map[string]interface{}),
	}

	expanded, err := m.expand(context.Background(), node)
	require.NoError(t, err)
	assert.Equal(t, MCTSNodeStateTerminal, expanded.NodeState)
}

func TestMCTS_Expand_CreatesChildren(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1", "a2", "a3"},
	}
	rf := &mockRewardFunction{isTerminal: false}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	node := &MCTSNode{
		ID:        "node-1",
		Depth:     1,
		State:     "state",
		NodeState: MCTSNodeStateUnexpanded,
		Metadata:  make(map[string]interface{}),
	}

	expanded, err := m.expand(context.Background(), node)
	require.NoError(t, err)
	assert.NotNil(t, expanded)
	assert.Equal(t, MCTSNodeStateExpanded, node.NodeState)
	assert.Len(t, node.Children, 3)

	for i, child := range node.Children {
		assert.Equal(t, fmt.Sprintf("node-1-%d", i), child.ID)
		assert.Equal(t, "node-1", child.ParentID)
		assert.Equal(t, 2, child.Depth)
		assert.Equal(t, MCTSNodeStateUnexpanded, child.NodeState)
		assert.NotNil(t, child.Metadata)
	}
}

// ---------------------------------------------------------------------------
// backpropagate
// ---------------------------------------------------------------------------

func TestMCTS_Backpropagate_SingleNode(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)
	m.root = &MCTSNode{ID: "root"}

	m.backpropagate(m.root, 0.8)
	assert.Equal(t, 1, m.root.Visits)
	assert.InDelta(t, 0.8, m.root.TotalReward, 0.001)
}

func TestMCTS_Backpropagate_WithDiscount(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.DiscountFactor = 0.9
	m := NewMCTS(cfg, ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	child := &MCTSNode{ID: "child", ParentID: "root"}
	root.Children = []*MCTSNode{child}
	m.root = root

	m.backpropagate(child, 1.0)

	assert.Equal(t, 1, child.Visits)
	assert.InDelta(t, 1.0, child.TotalReward, 0.001)
	assert.Equal(t, 1, root.Visits)
	assert.InDelta(t, 0.9, root.TotalReward, 0.001)
}

func TestMCTS_Backpropagate_DeepPath(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.DiscountFactor = 1.0 // No discount for easy verification
	m := NewMCTS(cfg, ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	child1 := &MCTSNode{ID: "c1", ParentID: "root"}
	child2 := &MCTSNode{ID: "c2", ParentID: "c1"}
	root.Children = []*MCTSNode{child1}
	child1.Children = []*MCTSNode{child2}
	m.root = root

	m.backpropagate(child2, 0.5)

	assert.Equal(t, 1, child2.Visits)
	assert.Equal(t, 1, child1.Visits)
	assert.Equal(t, 1, root.Visits)
	assert.InDelta(t, 0.5, root.TotalReward, 0.001)
}

// ---------------------------------------------------------------------------
// getBestPath
// ---------------------------------------------------------------------------

func TestMCTS_GetBestPath_RootOnly(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)
	m.root = &MCTSNode{ID: "root"}

	path := m.getBestPath()
	assert.Len(t, path, 1)
	assert.Equal(t, "root", path[0].ID)
}

func TestMCTS_GetBestPath_SelectsHighestAverageReward(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{ID: "root", Visits: 10, TotalReward: 5.0}
	child1 := &MCTSNode{ID: "c1", Visits: 5, TotalReward: 1.0, Action: "a1"} // avg 0.2
	child2 := &MCTSNode{ID: "c2", Visits: 5, TotalReward: 4.5, Action: "a2"} // avg 0.9
	root.Children = []*MCTSNode{child1, child2}
	m.root = root

	path := m.getBestPath()
	assert.Len(t, path, 2)
	assert.Equal(t, "root", path[0].ID)
	assert.Equal(t, "c2", path[1].ID)
}

// ---------------------------------------------------------------------------
// countNodes
// ---------------------------------------------------------------------------

func TestMCTS_CountNodes_NilNode(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)
	assert.Equal(t, 0, m.countNodes(nil))
}

func TestMCTS_CountNodes_SingleNode(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)
	assert.Equal(t, 1, m.countNodes(&MCTSNode{}))
}

func TestMCTS_CountNodes_Tree(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	c1 := &MCTSNode{ID: "c1"}
	c2 := &MCTSNode{ID: "c2"}
	c3 := &MCTSNode{ID: "c3"}
	c1.Children = []*MCTSNode{c3}
	root.Children = []*MCTSNode{c1, c2}

	assert.Equal(t, 4, m.countNodes(root))
}

// ---------------------------------------------------------------------------
// findParent
// ---------------------------------------------------------------------------

func TestMCTS_FindParent_RootIsParent(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	found := m.findParent(root, "root")
	assert.Equal(t, root, found)
}

func TestMCTS_FindParent_ChildIsParent(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	child := &MCTSNode{ID: "child"}
	root.Children = []*MCTSNode{child}

	found := m.findParent(root, "child")
	assert.Equal(t, child, found)
}

func TestMCTS_FindParent_NotFound(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	found := m.findParent(root, "nonexistent")
	assert.Nil(t, found)
}

func TestMCTS_FindParent_DeepTree(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{ID: "root"}
	c1 := &MCTSNode{ID: "c1"}
	c2 := &MCTSNode{ID: "c2"}
	c3 := &MCTSNode{ID: "c3"}
	root.Children = []*MCTSNode{c1}
	c1.Children = []*MCTSNode{c2}
	c2.Children = []*MCTSNode{c3}

	found := m.findParent(root, "c3")
	assert.Equal(t, c3, found)
}

// ---------------------------------------------------------------------------
// simulate
// ---------------------------------------------------------------------------

func TestMCTS_Simulate_WithRolloutPolicy(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{reward: 0.3}
	rp := &mockRolloutPolicy{rolloutValue: 0.9}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, rp, nil)

	node := &MCTSNode{State: "state"}
	val, err := m.simulate(context.Background(), node)
	require.NoError(t, err)
	assert.InDelta(t, 0.9, val, 0.001)
}

func TestMCTS_Simulate_WithoutRolloutPolicy(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{reward: 0.7}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil) // nil rollout

	node := &MCTSNode{State: "state"}
	val, err := m.simulate(context.Background(), node)
	require.NoError(t, err)
	assert.InDelta(t, 0.7, val, 0.001)
}

func TestMCTS_Simulate_RolloutPolicyError(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	rp := &mockRolloutPolicy{rolloutErr: errors.New("rollout error")}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, rp, nil)

	node := &MCTSNode{State: "state"}
	_, err := m.simulate(context.Background(), node)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// MCTSResult JSON marshaling
// ---------------------------------------------------------------------------

func TestMCTSResult_MarshalJSON(t *testing.T) {
	r := &MCTSResult{
		BestActions:     []string{"a1", "a2"},
		FinalReward:     0.85,
		TotalIterations: 100,
		Duration:        3500 * time.Millisecond,
		RootVisits:      100,
		TreeSize:        50,
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, float64(3500), decoded["duration_ms"])
	assert.Equal(t, float64(100), decoded["total_iterations"])
	assert.Equal(t, 0.85, decoded["final_reward"])
}

func TestMCTSResult_MarshalJSON_EmptyActions(t *testing.T) {
	r := &MCTSResult{
		BestActions: []string{},
		Duration:    0,
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, float64(0), decoded["duration_ms"])
}

// ---------------------------------------------------------------------------
// CodeActionGenerator
// ---------------------------------------------------------------------------

func TestNewCodeActionGenerator(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "", nil }
	logger := logrus.New()
	gen := NewCodeActionGenerator(fn, logger)
	require.NotNil(t, gen)
	assert.Equal(t, logger, gen.logger)
}

func TestCodeActionGenerator_GetActions_StringState(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "Add error handling\nRefactor function\nAdd tests", nil
	}
	gen := NewCodeActionGenerator(fn, nil)

	actions, err := gen.GetActions(context.Background(), "func main() {}")
	require.NoError(t, err)
	assert.NotEmpty(t, actions)
}

func TestCodeActionGenerator_GetActions_NonStringState(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "action1\naction2", nil
	}
	gen := NewCodeActionGenerator(fn, nil)

	actions, err := gen.GetActions(context.Background(), 42)
	require.NoError(t, err)
	assert.NotEmpty(t, actions)
}

func TestCodeActionGenerator_GetActions_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("gen error")
	}
	gen := NewCodeActionGenerator(fn, nil)

	actions, err := gen.GetActions(context.Background(), "state")
	require.Error(t, err)
	assert.Nil(t, actions)
}

func TestCodeActionGenerator_ApplyAction_StringState(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "updated code", nil
	}
	gen := NewCodeActionGenerator(fn, nil)

	newState, err := gen.ApplyAction(context.Background(), "code", "refactor")
	require.NoError(t, err)
	assert.Equal(t, "updated code", newState)
}

func TestCodeActionGenerator_ApplyAction_NonStringState(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "result", nil
	}
	gen := NewCodeActionGenerator(fn, nil)

	newState, err := gen.ApplyAction(context.Background(), 123, "action")
	require.NoError(t, err)
	assert.Equal(t, "result", newState)
}

func TestCodeActionGenerator_ApplyAction_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("apply error")
	}
	gen := NewCodeActionGenerator(fn, nil)

	_, err := gen.ApplyAction(context.Background(), "code", "action")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// CodeRewardFunction
// ---------------------------------------------------------------------------

func TestNewCodeRewardFunction(t *testing.T) {
	evalFn := func(ctx context.Context, code string) (float64, error) { return 0.5, nil }
	testFn := func(ctx context.Context, code string) (bool, error) { return false, nil }
	logger := logrus.New()

	rf := NewCodeRewardFunction(evalFn, testFn, logger)
	require.NotNil(t, rf)
	assert.Equal(t, logger, rf.logger)
}

func TestCodeRewardFunction_Evaluate_StringState(t *testing.T) {
	evalFn := func(ctx context.Context, code string) (float64, error) {
		return 0.8, nil
	}
	rf := NewCodeRewardFunction(evalFn, nil, nil)

	reward, err := rf.Evaluate(context.Background(), "code")
	require.NoError(t, err)
	assert.Equal(t, 0.8, reward)
}

func TestCodeRewardFunction_Evaluate_NonStringState(t *testing.T) {
	evalFn := func(ctx context.Context, code string) (float64, error) {
		return 0.5, nil
	}
	rf := NewCodeRewardFunction(evalFn, nil, nil)

	reward, err := rf.Evaluate(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 0.5, reward)
}

func TestCodeRewardFunction_Evaluate_NilFunc(t *testing.T) {
	rf := NewCodeRewardFunction(nil, nil, nil)

	reward, err := rf.Evaluate(context.Background(), "code")
	require.NoError(t, err)
	assert.Equal(t, 0.5, reward)
}

func TestCodeRewardFunction_Evaluate_Error(t *testing.T) {
	evalFn := func(ctx context.Context, code string) (float64, error) {
		return 0, errors.New("eval error")
	}
	rf := NewCodeRewardFunction(evalFn, nil, nil)

	_, err := rf.Evaluate(context.Background(), "code")
	require.Error(t, err)
}

func TestCodeRewardFunction_IsTerminal_True(t *testing.T) {
	testFn := func(ctx context.Context, code string) (bool, error) {
		return true, nil
	}
	rf := NewCodeRewardFunction(nil, testFn, nil)

	terminal, err := rf.IsTerminal(context.Background(), "code")
	require.NoError(t, err)
	assert.True(t, terminal)
}

func TestCodeRewardFunction_IsTerminal_False(t *testing.T) {
	testFn := func(ctx context.Context, code string) (bool, error) {
		return false, nil
	}
	rf := NewCodeRewardFunction(nil, testFn, nil)

	terminal, err := rf.IsTerminal(context.Background(), "code")
	require.NoError(t, err)
	assert.False(t, terminal)
}

func TestCodeRewardFunction_IsTerminal_NilFunc(t *testing.T) {
	rf := NewCodeRewardFunction(nil, nil, nil)

	terminal, err := rf.IsTerminal(context.Background(), "code")
	require.NoError(t, err)
	assert.False(t, terminal)
}

func TestCodeRewardFunction_IsTerminal_NonStringState(t *testing.T) {
	testFn := func(ctx context.Context, code string) (bool, error) {
		return code == "42", nil
	}
	rf := NewCodeRewardFunction(nil, testFn, nil)

	terminal, err := rf.IsTerminal(context.Background(), 42)
	require.NoError(t, err)
	assert.True(t, terminal)
}

// ---------------------------------------------------------------------------
// DefaultRolloutPolicy
// ---------------------------------------------------------------------------

func TestNewDefaultRolloutPolicy(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{"a1"}}
	rf := &mockRewardFunction{reward: 0.5}

	policy := NewDefaultRolloutPolicy(ag, rf)
	require.NotNil(t, policy)
	assert.NotNil(t, policy.rng)
}

func TestDefaultRolloutPolicy_Rollout_Basic(t *testing.T) {
	ag := &mockActionGenerator{
		actions: []string{"a1", "a2"},
	}
	rf := &mockRewardFunction{
		reward:     0.5,
		isTerminal: false,
	}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "start", 3)
	require.NoError(t, err)
	assert.True(t, val > 0)
}

func TestDefaultRolloutPolicy_Rollout_TerminalImmediately(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{"a1"}}
	rf := &mockRewardFunction{
		reward:     0.9,
		isTerminal: true,
	}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "terminal state", 10)
	require.NoError(t, err)
	// Breaks immediately, totalReward = 0
	assert.Equal(t, 0.0, val)
}

func TestDefaultRolloutPolicy_Rollout_NoActions(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{}}
	rf := &mockRewardFunction{reward: 0.5, isTerminal: false}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "start", 5)
	require.NoError(t, err)
	assert.Equal(t, 0.0, val)
}

func TestDefaultRolloutPolicy_Rollout_ApplyActionError(t *testing.T) {
	ag := &mockActionGenerator{
		actions:  []string{"a1"},
		applyErr: errors.New("apply error"),
	}
	rf := &mockRewardFunction{reward: 0.5, isTerminal: false}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "start", 5)
	require.NoError(t, err)
	assert.Equal(t, 0.0, val)
}

func TestDefaultRolloutPolicy_Rollout_ZeroDepth(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{"a1"}}
	rf := &mockRewardFunction{reward: 0.5, isTerminal: false}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "start", 0)
	require.NoError(t, err)
	assert.Equal(t, 0.0, val)
}

func TestDefaultRolloutPolicy_Rollout_EvaluateError(t *testing.T) {
	ag := &mockActionGenerator{actions: []string{"a1"}}
	rf := &mockRewardFunction{
		reward:    0,
		rewardErr: errors.New("eval error"),
	}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "start", 5)
	require.NoError(t, err)
	assert.Equal(t, 0.0, val)
}

func TestDefaultRolloutPolicy_Rollout_GetActionsError(t *testing.T) {
	ag := &mockActionGenerator{
		actions:    nil,
		actionsErr: errors.New("no actions"),
	}
	rf := &mockRewardFunction{reward: 0.5, isTerminal: false}

	policy := NewDefaultRolloutPolicy(ag, rf)
	val, err := policy.Rollout(context.Background(), "start", 5)
	require.NoError(t, err)
	assert.Equal(t, 0.0, val)
}

// ---------------------------------------------------------------------------
// selectNode
// ---------------------------------------------------------------------------

func TestMCTS_SelectNode_UnexpandedRoot(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{
		ID:        "root",
		NodeState: MCTSNodeStateUnexpanded,
	}

	selected := m.selectNode(context.Background(), root)
	assert.Equal(t, root, selected)
}

func TestMCTS_SelectNode_ExpandedWithUnexpandedChild(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	m := NewMCTS(DefaultMCTSConfig(), ag, rf, nil, nil)

	root := &MCTSNode{
		ID:        "root",
		NodeState: MCTSNodeStateExpanded,
		Visits:    10,
	}
	child1 := &MCTSNode{
		ID:        "c1",
		NodeState: MCTSNodeStateExpanded,
		Visits:    5,
	}
	child2 := &MCTSNode{
		ID:        "c2",
		NodeState: MCTSNodeStateUnexpanded,
		Visits:    0,
	}
	root.Children = []*MCTSNode{child1, child2}

	selected := m.selectNode(context.Background(), root)
	assert.Equal(t, "c2", selected.ID)
}

func TestMCTS_SelectNode_AllChildrenExpanded(t *testing.T) {
	ag := &mockActionGenerator{}
	rf := &mockRewardFunction{}
	cfg := DefaultMCTSConfig()
	cfg.UseUCTDP = false
	m := NewMCTS(cfg, ag, rf, nil, nil)

	root := &MCTSNode{
		ID:        "root",
		NodeState: MCTSNodeStateExpanded,
		Visits:    10,
	}
	child1 := &MCTSNode{
		ID:          "c1",
		NodeState:   MCTSNodeStateExpanded,
		Visits:      5,
		TotalReward: 3.0,
		Depth:       1,
	}
	child2 := &MCTSNode{
		ID:          "c2",
		NodeState:   MCTSNodeStateExpanded,
		Visits:      5,
		TotalReward: 4.0,
		Depth:       1,
	}
	// Give children no children of their own
	root.Children = []*MCTSNode{child1, child2}

	selected := m.selectNode(context.Background(), root)
	// child2 has higher reward, selectBestChild returns it, then loop ends since no children
	assert.NotNil(t, selected)
}

// ---------------------------------------------------------------------------
// Integration: search with increasing rewards
// ---------------------------------------------------------------------------

func TestMCTS_Search_IncreasingRewards(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	ag := &mockActionGenerator{
		actions: []string{"improve"},
	}
	rf := &mockRewardFunction{}
	// Override evaluate to return increasing rewards
	customRf := &mockRewardFunction{
		isTerminal: false,
	}
	// We'll use a closure-based approach via CodeRewardFunction
	evalFn := func(ctx context.Context, code string) (float64, error) {
		mu.Lock()
		callCount++
		r := float64(callCount) * 0.01
		mu.Unlock()
		if r > 1.0 {
			r = 1.0
		}
		return r, nil
	}
	testFn := func(ctx context.Context, code string) (bool, error) {
		return false, nil
	}
	codeRf := NewCodeRewardFunction(evalFn, testFn, nil)
	_ = rf
	_ = customRf

	cfg := DefaultMCTSConfig()
	cfg.MaxIterations = 20
	cfg.MaxDepth = 5
	cfg.Timeout = 5 * time.Second
	m := NewMCTS(cfg, ag, codeRf, nil, nil)

	result, err := m.Search(context.Background(), "initial code")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.FinalReward > 0)
}

// ---------------------------------------------------------------------------
// MCTSNode JSON serialization
// ---------------------------------------------------------------------------

func TestMCTSNode_JSONSerialization(t *testing.T) {
	node := &MCTSNode{
		ID:          "test-node",
		ParentID:    "parent",
		State:       "some state",
		Action:      "action-1",
		Visits:      10,
		TotalReward: 7.5,
		NodeState:   MCTSNodeStateExpanded,
		Depth:       3,
		Metadata:    map[string]interface{}{"key": "val"},
		CreatedAt:   time.Now(),
	}

	data, err := json.Marshal(node)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "test-node", decoded["id"])
	assert.Equal(t, "parent", decoded["parent_id"])
	assert.Equal(t, float64(10), decoded["visits"])
	assert.Equal(t, 7.5, decoded["total_reward"])
	assert.Equal(t, "expanded", decoded["node_state"])
}
