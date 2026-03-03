package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"digital.vasic.planning/planning"
)

// --- mock implementations for integration tests ---

type mockMilestoneGenerator struct {
	milestones []*planning.Milestone
	steps      []*planning.PlanStep
	hints      []string
}

func (m *mockMilestoneGenerator) GenerateMilestones(_ context.Context, goal string) ([]*planning.Milestone, error) {
	if len(m.milestones) > 0 {
		return m.milestones, nil
	}
	return []*planning.Milestone{
		{
			ID:       "m-1",
			Name:     "Milestone 1 for " + goal,
			State:    planning.MilestoneStatePending,
			Priority: 0,
			Metadata: make(map[string]interface{}),
		},
		{
			ID:       "m-2",
			Name:     "Milestone 2 for " + goal,
			State:    planning.MilestoneStatePending,
			Priority: 1,
			Metadata: make(map[string]interface{}),
		},
	}, nil
}

func (m *mockMilestoneGenerator) GenerateSteps(_ context.Context, milestone *planning.Milestone) ([]*planning.PlanStep, error) {
	if len(m.steps) > 0 {
		return m.steps, nil
	}
	return []*planning.PlanStep{
		{
			ID:          milestone.ID + "-s1",
			MilestoneID: milestone.ID,
			Action:      "Execute step 1",
			State:       planning.PlanStepStatePending,
			Inputs:      make(map[string]interface{}),
			Outputs:     make(map[string]interface{}),
		},
		{
			ID:          milestone.ID + "-s2",
			MilestoneID: milestone.ID,
			Action:      "Execute step 2",
			State:       planning.PlanStepStatePending,
			Inputs:      make(map[string]interface{}),
			Outputs:     make(map[string]interface{}),
		},
	}, nil
}

func (m *mockMilestoneGenerator) GenerateHints(_ context.Context, _ *planning.PlanStep, _ string) ([]string, error) {
	if len(m.hints) > 0 {
		return m.hints, nil
	}
	return []string{"hint-1", "hint-2"}, nil
}

type mockStepExecutor struct {
	succeed bool
}

func (m *mockStepExecutor) Execute(_ context.Context, step *planning.PlanStep, _ []string) (*planning.StepResult, error) {
	return &planning.StepResult{
		Success:  m.succeed,
		Outputs:  map[string]interface{}{"result": step.Action + " done"},
		Logs:     []string{"executed " + step.ID},
		Duration: 10 * time.Millisecond,
	}, nil
}

func (m *mockStepExecutor) Validate(_ context.Context, _ *planning.PlanStep, _ *planning.StepResult) error {
	return nil
}

type mockActionGenerator struct{}

func (m *mockActionGenerator) GetActions(_ context.Context, state interface{}) ([]string, error) {
	return []string{"action-a", "action-b"}, nil
}

func (m *mockActionGenerator) ApplyAction(_ context.Context, state interface{}, action string) (interface{}, error) {
	return fmt.Sprintf("%v -> %s", state, action), nil
}

type mockRewardFunction struct {
	terminalDepth int
	callCount     int
	mu            sync.Mutex
}

func (m *mockRewardFunction) Evaluate(_ context.Context, state interface{}) (float64, error) {
	return 0.7, nil
}

func (m *mockRewardFunction) IsTerminal(_ context.Context, state interface{}) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.callCount > m.terminalDepth, nil
}

type mockThoughtGenerator struct{}

func (m *mockThoughtGenerator) GenerateThoughts(_ context.Context, parent *planning.Thought, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("%s-child-%d", parent.ID, i),
			ParentID:  parent.ID,
			Content:   fmt.Sprintf("Thought from %s variant %d", parent.Content, i),
			State:     planning.ThoughtStatePending,
			Depth:     parent.Depth + 1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"generated": true},
		})
	}
	return thoughts, nil
}

func (m *mockThoughtGenerator) GenerateInitialThoughts(_ context.Context, problem string, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("init-%d", i),
			Content:   fmt.Sprintf("Initial approach %d for %s", i, problem),
			State:     planning.ThoughtStatePending,
			Depth:     1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"type": "initial"},
		})
	}
	return thoughts, nil
}

type mockThoughtEvaluator struct {
	baseScore float64
}

func (m *mockThoughtEvaluator) EvaluateThought(_ context.Context, thought *planning.Thought) (float64, error) {
	score := m.baseScore
	if thought.Depth > 2 {
		score = 0.95
	}
	return score, nil
}

func (m *mockThoughtEvaluator) EvaluatePath(_ context.Context, path []*planning.Thought) (float64, error) {
	if len(path) == 0 {
		return 0, nil
	}
	total := 0.0
	for _, t := range path {
		total += t.Score
	}
	return total / float64(len(path)), nil
}

func (m *mockThoughtEvaluator) IsTerminal(_ context.Context, thought *planning.Thought) (bool, error) {
	return thought.Score > 0.9, nil
}

// --- Integration Tests ---

func TestHiPlan_CreateAndExecutePlan_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{succeed: true}
	config := planning.DefaultHiPlanConfig()
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "build a web service")
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.Equal(t, "build a web service", plan.Goal)
	assert.Len(t, plan.Milestones, 2)
	assert.Equal(t, "created", plan.State)

	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.CompletedMilestones)
	assert.Equal(t, 0, result.FailedMilestones)
	assert.Equal(t, "completed", plan.State)
	assert.Equal(t, 1.0, plan.Progress)
}

func TestHiPlan_SequentialExecution_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{succeed: true}
	config := planning.DefaultHiPlanConfig()
	config.EnableParallelMilestones = false
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "sequential task")
	require.NoError(t, err)

	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.MilestoneResults, 2)

	for _, mr := range result.MilestoneResults {
		assert.True(t, mr.Success)
		assert.Greater(t, mr.Duration, time.Duration(0))
	}
}

func TestHiPlan_FailedSteps_AdaptivePlanning_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{succeed: false}
	config := planning.DefaultHiPlanConfig()
	config.EnableAdaptivePlanning = true
	config.RetryFailedSteps = false
	config.MaxRetries = 0
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "failing task")
	require.NoError(t, err)

	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "failed", plan.State)
}

func TestMCTS_SearchAndExplore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.MaxIterations = 50
	config.MaxDepth = 5
	config.Timeout = 30 * time.Second
	config.EnableParallel = false

	actionGen := &mockActionGenerator{}
	rewardFunc := &mockRewardFunction{terminalDepth: 10}

	mcts := planning.NewMCTS(config, actionGen, rewardFunc, nil, nil)

	ctx := context.Background()
	result, err := mcts.Search(ctx, "initial-state")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Greater(t, result.TotalIterations, 0)
	assert.Greater(t, result.TreeSize, 1)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestMCTS_WithRolloutPolicy_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.MaxIterations = 30
	config.MaxDepth = 4
	config.RolloutDepth = 3
	config.Timeout = 30 * time.Second

	actionGen := &mockActionGenerator{}
	rewardFunc := &mockRewardFunction{terminalDepth: 8}
	rollout := planning.NewDefaultRolloutPolicy(actionGen, rewardFunc)

	mcts := planning.NewMCTS(config, actionGen, rewardFunc, rollout, nil)

	ctx := context.Background()
	result, err := mcts.Search(ctx, "code-start")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, result.RootVisits, 0)
}

func TestTreeOfThoughts_BeamSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.SearchStrategy = "beam"
	config.MaxDepth = 4
	config.BeamWidth = 2
	config.MaxBranches = 3
	config.MaxIterations = 50
	config.Timeout = 30 * time.Second

	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{baseScore: 0.6}

	tot := planning.NewTreeOfThoughts(config, gen, eval, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "design an API")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "design an API", result.Problem)
	assert.Equal(t, "beam", result.Strategy)
	assert.Greater(t, result.Iterations, 0)
	assert.Greater(t, result.NodesExplored, 0)
}

func TestTreeOfThoughts_BFSSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.SearchStrategy = "bfs"
	config.MaxDepth = 3
	config.MaxBranches = 2
	config.MaxIterations = 30
	config.Timeout = 30 * time.Second

	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{baseScore: 0.6}

	tot := planning.NewTreeOfThoughts(config, gen, eval, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "optimize database")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "bfs", result.Strategy)
}

func TestTreeOfThoughts_DFSSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.SearchStrategy = "dfs"
	config.MaxDepth = 3
	config.MaxBranches = 2
	config.MaxIterations = 30
	config.Timeout = 30 * time.Second

	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{baseScore: 0.6}

	tot := planning.NewTreeOfThoughts(config, gen, eval, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "refactor code")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "dfs", result.Strategy)
}

func TestHiPlan_LibraryOperations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{succeed: true}
	config := planning.DefaultHiPlanConfig()
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	milestone := &planning.Milestone{
		ID:       "lib-m1",
		Name:     "Library milestone",
		State:    planning.MilestoneStateCompleted,
		Metadata: map[string]interface{}{"reusable": true},
	}

	hp.AddToLibrary(milestone)
	retrieved, ok := hp.GetFromLibrary("lib-m1")
	require.True(t, ok)
	assert.Equal(t, "Library milestone", retrieved.Name)

	_, ok = hp.GetFromLibrary("nonexistent")
	assert.False(t, ok)
}

func TestHiPlan_WithDependencies_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gen := &mockMilestoneGenerator{
		milestones: []*planning.Milestone{
			{
				ID:       "m-a",
				Name:     "First milestone",
				State:    planning.MilestoneStatePending,
				Priority: 0,
				Metadata: make(map[string]interface{}),
			},
			{
				ID:           "m-b",
				Name:         "Second milestone",
				State:        planning.MilestoneStatePending,
				Priority:     1,
				Dependencies: []string{"m-a"},
				Metadata:     make(map[string]interface{}),
			},
		},
	}
	exec := &mockStepExecutor{succeed: true}
	config := planning.DefaultHiPlanConfig()
	config.EnableParallelMilestones = true
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "ordered task")
	require.NoError(t, err)
	assert.Len(t, plan.Milestones, 2)

	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.Equal(t, 2, result.CompletedMilestones)
}

func TestMCTS_UCTValue_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.UseUCTDP = true
	config.ExplorationConstant = 1.414
	config.DepthPreferenceAlpha = 0.5
	config.MaxDepth = 50

	actionGen := &mockActionGenerator{}
	rewardFunc := &mockRewardFunction{terminalDepth: 100}

	mcts := planning.NewMCTS(config, actionGen, rewardFunc, nil, nil)

	node := &planning.MCTSNode{
		ID:    "test-node",
		Depth: 10,
	}
	node.AddReward(0.8)

	uctVal := mcts.UCTValue(node, 100)
	assert.Greater(t, uctVal, 0.0)

	nilVal := mcts.UCTValue(nil, 100)
	assert.Equal(t, 0.0, nilVal)
}

func TestAllThreeAlgorithms_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// HiPlan
	hiConfig := planning.DefaultHiPlanConfig()
	hiConfig.Timeout = 15 * time.Second
	hp := planning.NewHiPlan(hiConfig, &mockMilestoneGenerator{}, &mockStepExecutor{succeed: true}, nil)
	plan, err := hp.CreatePlan(ctx, "integration test goal")
	require.NoError(t, err)
	planResult, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.True(t, planResult.Success)

	// MCTS
	mctsConfig := planning.DefaultMCTSConfig()
	mctsConfig.MaxIterations = 20
	mctsConfig.MaxDepth = 3
	mctsConfig.Timeout = 15 * time.Second
	mcts := planning.NewMCTS(mctsConfig, &mockActionGenerator{}, &mockRewardFunction{terminalDepth: 5}, nil, nil)
	mctsResult, err := mcts.Search(ctx, "start")
	require.NoError(t, err)
	assert.Greater(t, mctsResult.TreeSize, 0)

	// ToT
	totConfig := planning.DefaultTreeOfThoughtsConfig()
	totConfig.MaxDepth = 3
	totConfig.MaxBranches = 2
	totConfig.MaxIterations = 20
	totConfig.Timeout = 15 * time.Second
	tot := planning.NewTreeOfThoughts(totConfig, &mockThoughtGenerator{}, &mockThoughtEvaluator{baseScore: 0.6}, nil)
	totResult, err := tot.Solve(ctx, "solve a problem")
	require.NoError(t, err)
	assert.Greater(t, totResult.NodesExplored, 0)
}
