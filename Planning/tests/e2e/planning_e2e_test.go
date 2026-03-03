package e2e

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

// --- helper implementations ---

type e2eMilestoneGenerator struct {
	stepCount int
}

func (g *e2eMilestoneGenerator) GenerateMilestones(_ context.Context, goal string) ([]*planning.Milestone, error) {
	return []*planning.Milestone{
		{
			ID:          "e2e-m1",
			Name:        "Phase 1: Setup for " + goal,
			Description: "Initial setup phase",
			State:       planning.MilestoneStatePending,
			Priority:    0,
			Metadata:    make(map[string]interface{}),
		},
		{
			ID:           "e2e-m2",
			Name:         "Phase 2: Build for " + goal,
			Description:  "Build phase",
			State:        planning.MilestoneStatePending,
			Priority:     1,
			Dependencies: []string{"e2e-m1"},
			Metadata:     make(map[string]interface{}),
		},
		{
			ID:           "e2e-m3",
			Name:         "Phase 3: Deploy for " + goal,
			Description:  "Deployment phase",
			State:        planning.MilestoneStatePending,
			Priority:     2,
			Dependencies: []string{"e2e-m2"},
			Metadata:     make(map[string]interface{}),
		},
	}, nil
}

func (g *e2eMilestoneGenerator) GenerateSteps(_ context.Context, milestone *planning.Milestone) ([]*planning.PlanStep, error) {
	count := g.stepCount
	if count == 0 {
		count = 3
	}
	steps := make([]*planning.PlanStep, 0, count)
	for i := 0; i < count; i++ {
		steps = append(steps, &planning.PlanStep{
			ID:          fmt.Sprintf("%s-step-%d", milestone.ID, i),
			MilestoneID: milestone.ID,
			Action:      fmt.Sprintf("Execute action %d of %s", i, milestone.Name),
			State:       planning.PlanStepStatePending,
			Inputs:      map[string]interface{}{"index": i},
			Outputs:     make(map[string]interface{}),
		})
	}
	return steps, nil
}

func (g *e2eMilestoneGenerator) GenerateHints(_ context.Context, step *planning.PlanStep, _ string) ([]string, error) {
	return []string{
		"Verify input parameters",
		"Check output consistency",
	}, nil
}

type e2eStepExecutor struct {
	mu            sync.Mutex
	executedSteps []string
}

func (e *e2eStepExecutor) Execute(_ context.Context, step *planning.PlanStep, _ []string) (*planning.StepResult, error) {
	e.mu.Lock()
	e.executedSteps = append(e.executedSteps, step.ID)
	e.mu.Unlock()

	return &planning.StepResult{
		Success:  true,
		Outputs:  map[string]interface{}{"status": "done", "step": step.ID},
		Logs:     []string{"step " + step.ID + " executed"},
		Duration: 5 * time.Millisecond,
	}, nil
}

func (e *e2eStepExecutor) Validate(_ context.Context, _ *planning.PlanStep, _ *planning.StepResult) error {
	return nil
}

type e2eActionGenerator struct{}

func (g *e2eActionGenerator) GetActions(_ context.Context, state interface{}) ([]string, error) {
	return []string{"analyze", "implement", "test"}, nil
}

func (g *e2eActionGenerator) ApplyAction(_ context.Context, state interface{}, action string) (interface{}, error) {
	return fmt.Sprintf("%v/%s", state, action), nil
}

type e2eRewardFunction struct {
	mu    sync.Mutex
	calls int
}

func (f *e2eRewardFunction) Evaluate(_ context.Context, state interface{}) (float64, error) {
	return 0.75, nil
}

func (f *e2eRewardFunction) IsTerminal(_ context.Context, state interface{}) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.calls > 8, nil
}

type e2eThoughtGenerator struct{}

func (g *e2eThoughtGenerator) GenerateThoughts(_ context.Context, parent *planning.Thought, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("%s-t%d", parent.ID, i),
			ParentID:  parent.ID,
			Content:   fmt.Sprintf("solution approach %d from %s", i, parent.Content),
			State:     planning.ThoughtStatePending,
			Depth:     parent.Depth + 1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"variant": i},
		})
	}
	return thoughts, nil
}

func (g *e2eThoughtGenerator) GenerateInitialThoughts(_ context.Context, problem string, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("init-%d", i),
			Content:   fmt.Sprintf("strategy %d for: %s", i, problem),
			State:     planning.ThoughtStatePending,
			Depth:     1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"type": "initial"},
		})
	}
	return thoughts, nil
}

type e2eThoughtEvaluator struct {
	scoreByDepth map[int]float64
}

func (e *e2eThoughtEvaluator) EvaluateThought(_ context.Context, thought *planning.Thought) (float64, error) {
	if score, ok := e.scoreByDepth[thought.Depth]; ok {
		return score, nil
	}
	return 0.5 + float64(thought.Depth)*0.1, nil
}

func (e *e2eThoughtEvaluator) EvaluatePath(_ context.Context, path []*planning.Thought) (float64, error) {
	if len(path) == 0 {
		return 0, nil
	}
	sum := 0.0
	for _, t := range path {
		sum += t.Score
	}
	return sum / float64(len(path)), nil
}

func (e *e2eThoughtEvaluator) IsTerminal(_ context.Context, thought *planning.Thought) (bool, error) {
	return thought.Score > 0.9, nil
}

// --- E2E Tests ---

func TestFullPlanningWorkflow_HiPlan_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	gen := &e2eMilestoneGenerator{stepCount: 3}
	exec := &e2eStepExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.EnableParallelMilestones = true
	config.MaxParallelMilestones = 2
	config.Timeout = 60 * time.Second
	config.StepTimeout = 10 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)
	ctx := context.Background()

	// Step 1: Create Plan
	plan, err := hp.CreatePlan(ctx, "deploy microservices")
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.Equal(t, "created", plan.State)
	assert.Len(t, plan.Milestones, 3)

	// Verify milestones have steps
	for _, m := range plan.Milestones {
		assert.Len(t, m.Steps, 3, "milestone %s should have 3 steps", m.ID)
	}

	// Step 2: Execute Plan
	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 3, result.CompletedMilestones)
	assert.Equal(t, 0, result.FailedMilestones)
	assert.Equal(t, "completed", plan.State)
	assert.Equal(t, 1.0, plan.Progress)

	// Step 3: Verify execution tracking
	exec.mu.Lock()
	totalSteps := len(exec.executedSteps)
	exec.mu.Unlock()
	assert.Equal(t, 9, totalSteps, "should execute all 9 steps (3 milestones x 3 steps)")

	// Step 4: Verify plan in library
	currentPlan := hp.GetCurrentPlan()
	require.NotNil(t, currentPlan)
	assert.Equal(t, plan.ID, currentPlan.ID)

	// Step 5: Verify result JSON marshaling
	jsonData, err := result.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "duration_ms")
}

func TestFullPlanningWorkflow_MCTS_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.MaxIterations = 100
	config.MaxDepth = 6
	config.RolloutDepth = 3
	config.Timeout = 60 * time.Second
	config.UseUCTDP = true

	actionGen := &e2eActionGenerator{}
	rewardFunc := &e2eRewardFunction{}
	rollout := planning.NewDefaultRolloutPolicy(actionGen, rewardFunc)

	mcts := planning.NewMCTS(config, actionGen, rewardFunc, rollout, nil)

	ctx := context.Background()
	result, err := mcts.Search(ctx, "root-state")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Greater(t, result.TotalIterations, 0)
	assert.Greater(t, result.TreeSize, 1)
	assert.Greater(t, result.RootVisits, 0)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.NotNil(t, result.FinalState)

	// Verify JSON marshaling
	jsonData, err := result.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "duration_ms")
}

func TestFullPlanningWorkflow_TreeOfThoughts_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.SearchStrategy = "beam"
	config.MaxDepth = 5
	config.BeamWidth = 3
	config.MaxBranches = 3
	config.MaxIterations = 100
	config.Timeout = 60 * time.Second

	gen := &e2eThoughtGenerator{}
	eval := &e2eThoughtEvaluator{
		scoreByDepth: map[int]float64{
			1: 0.6,
			2: 0.7,
			3: 0.8,
			4: 0.92,
			5: 0.95,
		},
	}

	tot := planning.NewTreeOfThoughts(config, gen, eval, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "design a scalable architecture")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "design a scalable architecture", result.Problem)
	assert.Equal(t, "beam", result.Strategy)
	assert.Greater(t, result.Iterations, 0)
	assert.Greater(t, result.NodesExplored, 0)
	assert.Greater(t, result.TreeDepth, 0)

	// Verify solution content extraction
	contents := result.GetSolutionContent()
	assert.NotNil(t, contents)

	// Verify JSON marshaling
	jsonData, err := result.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "duration_ms")
}

func TestFullWorkflow_AllAlgorithms_SameProblem_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ctx := context.Background()
	goal := "implement a REST API"

	// Run HiPlan
	hiConfig := planning.DefaultHiPlanConfig()
	hiConfig.Timeout = 30 * time.Second
	hp := planning.NewHiPlan(hiConfig, &e2eMilestoneGenerator{stepCount: 2}, &e2eStepExecutor{}, nil)
	plan, err := hp.CreatePlan(ctx, goal)
	require.NoError(t, err)
	planResult, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.True(t, planResult.Success, "HiPlan should succeed")

	// Run MCTS
	mctsConfig := planning.DefaultMCTSConfig()
	mctsConfig.MaxIterations = 40
	mctsConfig.MaxDepth = 4
	mctsConfig.Timeout = 30 * time.Second
	mcts := planning.NewMCTS(mctsConfig, &e2eActionGenerator{}, &e2eRewardFunction{}, nil, nil)
	mctsResult, err := mcts.Search(ctx, goal)
	require.NoError(t, err)
	assert.Greater(t, mctsResult.TreeSize, 0, "MCTS should explore nodes")

	// Run ToT with all 3 strategies
	for _, strategy := range []string{"beam", "bfs", "dfs"} {
		totConfig := planning.DefaultTreeOfThoughtsConfig()
		totConfig.SearchStrategy = strategy
		totConfig.MaxDepth = 3
		totConfig.MaxBranches = 2
		totConfig.MaxIterations = 30
		totConfig.Timeout = 30 * time.Second

		tot := planning.NewTreeOfThoughts(totConfig, &e2eThoughtGenerator{},
			&e2eThoughtEvaluator{scoreByDepth: map[int]float64{1: 0.5, 2: 0.7, 3: 0.95}}, nil)
		totResult, err := tot.Solve(ctx, goal)
		require.NoError(t, err, "ToT %s should not error", strategy)
		assert.Greater(t, totResult.NodesExplored, 0, "ToT %s should explore nodes", strategy)
	}
}

func TestHiPlan_ExecuteStep_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	gen := &e2eMilestoneGenerator{}
	exec := &e2eStepExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.RetryFailedSteps = true
	config.MaxRetries = 2
	config.Timeout = 30 * time.Second
	config.StepTimeout = 10 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	step := &planning.PlanStep{
		ID:     "standalone-step",
		Action: "run analysis",
		State:  planning.PlanStepStatePending,
		Inputs: map[string]interface{}{"data": "sample"},
	}

	ctx := context.Background()
	result, err := hp.ExecuteStep(ctx, step, []string{"use caution"})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Outputs, "status")
}

func TestMCTSNode_RewardAccumulation_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	node := &planning.MCTSNode{
		ID:       "test-node",
		Depth:    3,
		Metadata: make(map[string]interface{}),
	}

	rewards := []float64{0.5, 0.7, 0.9, 0.3, 0.8}
	for _, r := range rewards {
		node.AddReward(r)
	}

	assert.Equal(t, 5, node.Visits)
	expectedAvg := (0.5 + 0.7 + 0.9 + 0.3 + 0.8) / 5.0
	assert.InDelta(t, expectedAvg, node.AverageReward(), 0.001)
}

func TestHiPlan_ContextCancellation_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	gen := &e2eMilestoneGenerator{stepCount: 100}
	exec := &e2eStepExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.EnableParallelMilestones = false
	config.Timeout = 2 * time.Second
	config.StepTimeout = 1 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	plan, err := hp.CreatePlan(ctx, "long task")
	if err != nil {
		// Context may expire during creation
		return
	}
	require.NotNil(t, plan)
}
