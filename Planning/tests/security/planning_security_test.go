package security

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"digital.vasic.planning/planning"
)

// --- minimal mocks for security edge-case tests ---

type safeGenerator struct{}

func (g *safeGenerator) GenerateMilestones(_ context.Context, goal string) ([]*planning.Milestone, error) {
	if goal == "" {
		return []*planning.Milestone{}, nil
	}
	return []*planning.Milestone{
		{
			ID:       "sec-m1",
			Name:     goal,
			State:    planning.MilestoneStatePending,
			Priority: 0,
			Metadata: make(map[string]interface{}),
		},
	}, nil
}

func (g *safeGenerator) GenerateSteps(_ context.Context, milestone *planning.Milestone) ([]*planning.PlanStep, error) {
	return []*planning.PlanStep{
		{
			ID:          milestone.ID + "-s1",
			MilestoneID: milestone.ID,
			Action:      "step action",
			State:       planning.PlanStepStatePending,
			Inputs:      make(map[string]interface{}),
			Outputs:     make(map[string]interface{}),
		},
	}, nil
}

func (g *safeGenerator) GenerateHints(_ context.Context, _ *planning.PlanStep, _ string) ([]string, error) {
	return []string{"hint"}, nil
}

type errorGenerator struct{}

func (g *errorGenerator) GenerateMilestones(_ context.Context, _ string) ([]*planning.Milestone, error) {
	return nil, fmt.Errorf("generation failed")
}

func (g *errorGenerator) GenerateSteps(_ context.Context, _ *planning.Milestone) ([]*planning.PlanStep, error) {
	return nil, fmt.Errorf("step generation failed")
}

func (g *errorGenerator) GenerateHints(_ context.Context, _ *planning.PlanStep, _ string) ([]string, error) {
	return nil, fmt.Errorf("hint generation failed")
}

type safeExecutor struct {
	succeed bool
}

func (e *safeExecutor) Execute(_ context.Context, _ *planning.PlanStep, _ []string) (*planning.StepResult, error) {
	return &planning.StepResult{Success: e.succeed, Duration: time.Millisecond}, nil
}

func (e *safeExecutor) Validate(_ context.Context, _ *planning.PlanStep, _ *planning.StepResult) error {
	return nil
}

type nilReturningExecutor struct{}

func (e *nilReturningExecutor) Execute(_ context.Context, _ *planning.PlanStep, _ []string) (*planning.StepResult, error) {
	return nil, fmt.Errorf("executor crashed")
}

func (e *nilReturningExecutor) Validate(_ context.Context, _ *planning.PlanStep, _ *planning.StepResult) error {
	return nil
}

type safeActionGen struct{}

func (g *safeActionGen) GetActions(_ context.Context, _ interface{}) ([]string, error) {
	return []string{"act"}, nil
}

func (g *safeActionGen) ApplyAction(_ context.Context, state interface{}, action string) (interface{}, error) {
	return fmt.Sprintf("%v+%s", state, action), nil
}

type safeReward struct {
	callCount int
}

func (r *safeReward) Evaluate(_ context.Context, _ interface{}) (float64, error) {
	return 0.5, nil
}

func (r *safeReward) IsTerminal(_ context.Context, _ interface{}) (bool, error) {
	r.callCount++
	return r.callCount > 5, nil
}

type safeThoughtGen struct{}

func (g *safeThoughtGen) GenerateThoughts(_ context.Context, parent *planning.Thought, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("%s-c%d", parent.ID, i),
			ParentID:  parent.ID,
			Content:   "thought variant",
			State:     planning.ThoughtStatePending,
			Depth:     parent.Depth + 1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{},
		})
	}
	return thoughts, nil
}

func (g *safeThoughtGen) GenerateInitialThoughts(_ context.Context, _ string, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("init-%d", i),
			Content:   "initial thought",
			State:     planning.ThoughtStatePending,
			Depth:     1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{},
		})
	}
	return thoughts, nil
}

type safeThoughtEval struct {
	score float64
}

func (e *safeThoughtEval) EvaluateThought(_ context.Context, _ *planning.Thought) (float64, error) {
	return e.score, nil
}

func (e *safeThoughtEval) EvaluatePath(_ context.Context, path []*planning.Thought) (float64, error) {
	if len(path) == 0 {
		return 0, nil
	}
	return e.score, nil
}

func (e *safeThoughtEval) IsTerminal(_ context.Context, thought *planning.Thought) (bool, error) {
	return thought.Score > 0.9, nil
}

// --- Security / Edge-Case Tests ---

func TestHiPlan_NilLogger_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	hp := planning.NewHiPlan(config, &safeGenerator{}, &safeExecutor{succeed: true}, nil)
	require.NotNil(t, hp)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "test goal")
	require.NoError(t, err)
	assert.NotNil(t, plan)
}

func TestHiPlan_EmptyGoal_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	config.Timeout = 10 * time.Second
	hp := planning.NewHiPlan(config, &safeGenerator{}, &safeExecutor{succeed: true}, nil)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "", plan.Goal)
}

func TestHiPlan_GeneratorError_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	config.Timeout = 10 * time.Second
	hp := planning.NewHiPlan(config, &errorGenerator{}, &safeExecutor{succeed: true}, nil)

	ctx := context.Background()
	_, err := hp.CreatePlan(ctx, "should fail")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate milestones")
}

func TestHiPlan_VeryLargeGoalString_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	config.Timeout = 10 * time.Second
	hp := planning.NewHiPlan(config, &safeGenerator{}, &safeExecutor{succeed: true}, nil)

	largeGoal := strings.Repeat("x", 100000)
	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, largeGoal)
	require.NoError(t, err)
	assert.Equal(t, largeGoal, plan.Goal)
}

func TestHiPlan_NilExecutorResult_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	config.RetryFailedSteps = false
	config.MaxRetries = 0
	config.EnableAdaptivePlanning = false
	config.Timeout = 10 * time.Second

	hp := planning.NewHiPlan(config, &safeGenerator{}, &nilReturningExecutor{}, nil)

	ctx := context.Background()
	plan, err := hp.CreatePlan(ctx, "crash test")
	require.NoError(t, err)

	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	assert.False(t, result.Success)
}

func TestHiPlan_ZeroConfig_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.HiPlanConfig{
		MaxMilestones:        1,
		MaxStepsPerMilestone: 1,
		MaxRetries:           0,
		Timeout:              10 * time.Second,
		StepTimeout:          5 * time.Second,
	}

	hp := planning.NewHiPlan(config, &safeGenerator{}, &safeExecutor{succeed: true}, nil)
	ctx := context.Background()

	plan, err := hp.CreatePlan(ctx, "minimal")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(plan.Milestones), 1)
}

func TestHiPlan_CancelledContext_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	config.Timeout = 10 * time.Second
	hp := planning.NewHiPlan(config, &safeGenerator{}, &safeExecutor{succeed: true}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := hp.CreatePlan(ctx, "cancelled")
	// Either an error or the plan creation may still succeed depending on timing
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
}

func TestMCTS_NilLogger_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.MaxIterations = 5
	config.Timeout = 10 * time.Second

	mcts := planning.NewMCTS(config, &safeActionGen{}, &safeReward{}, nil, nil)
	require.NotNil(t, mcts)
}

func TestMCTS_ZeroIterations_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.MaxIterations = 0
	config.Timeout = 10 * time.Second

	mcts := planning.NewMCTS(config, &safeActionGen{}, &safeReward{}, nil, nil)

	ctx := context.Background()
	result, err := mcts.Search(ctx, "state")
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalIterations)
}

func TestMCTS_ZeroDepth_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	config.MaxDepth = 0
	config.MaxIterations = 10
	config.Timeout = 10 * time.Second

	mcts := planning.NewMCTS(config, &safeActionGen{}, &safeReward{}, nil, nil)

	ctx := context.Background()
	result, err := mcts.Search(ctx, "state")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMCTS_UCTValue_UnvisitedNode_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	mcts := planning.NewMCTS(config, &safeActionGen{}, &safeReward{}, nil, nil)

	node := &planning.MCTSNode{ID: "unvisited", Depth: 1}
	val := mcts.UCTValue(node, 100)
	assert.True(t, val > 1e10, "unvisited nodes should have very high UCT value")
}

func TestMCTS_UCTValue_ZeroParentVisits_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultMCTSConfig()
	mcts := planning.NewMCTS(config, &safeActionGen{}, &safeReward{}, nil, nil)

	node := &planning.MCTSNode{ID: "node", Depth: 1}
	node.AddReward(0.5)
	val := mcts.UCTValue(node, 0)
	assert.False(t, val != val, "should not be NaN") // NaN check
}

func TestMCTSNode_ZeroVisits_AverageReward_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	node := &planning.MCTSNode{ID: "zero"}
	assert.Equal(t, 0.0, node.AverageReward())
}

func TestTreeOfThoughts_NilLogger_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.MaxIterations = 5
	config.Timeout = 10 * time.Second

	tot := planning.NewTreeOfThoughts(config, &safeThoughtGen{}, &safeThoughtEval{score: 0.5}, nil)
	require.NotNil(t, tot)
}

func TestTreeOfThoughts_EmptyProblem_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.MaxIterations = 10
	config.MaxDepth = 2
	config.Timeout = 10 * time.Second

	tot := planning.NewTreeOfThoughts(config, &safeThoughtGen{}, &safeThoughtEval{score: 0.5}, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "", result.Problem)
}

func TestTreeOfThoughts_ZeroBranches_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.MaxBranches = 0
	config.MaxIterations = 5
	config.Timeout = 10 * time.Second

	tot := planning.NewTreeOfThoughts(config, &safeThoughtGen{}, &safeThoughtEval{score: 0.5}, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "zero branch test")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestTreeOfThoughts_UnknownStrategy_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.SearchStrategy = "unknown_strategy"
	config.MaxIterations = 10
	config.MaxDepth = 2
	config.MaxBranches = 2
	config.Timeout = 10 * time.Second

	tot := planning.NewTreeOfThoughts(config, &safeThoughtGen{}, &safeThoughtEval{score: 0.6}, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "unknown strategy test")
	require.NoError(t, err)
	// Should fallback to beam search
	assert.NotNil(t, result)
}

func TestTreeOfThoughts_LowPruneThreshold_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	config := planning.DefaultTreeOfThoughtsConfig()
	config.PruneThreshold = 0.99
	config.MaxIterations = 20
	config.MaxDepth = 3
	config.MaxBranches = 2
	config.Timeout = 10 * time.Second

	tot := planning.NewTreeOfThoughts(config, &safeThoughtGen{}, &safeThoughtEval{score: 0.5}, nil)

	ctx := context.Background()
	result, err := tot.Solve(ctx, "aggressive pruning")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestToTResult_GetSolutionContent_Empty_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	result := &planning.ToTResult{
		Solution: nil,
	}
	contents := result.GetSolutionContent()
	assert.Empty(t, contents)
}

func TestDefaultConfigs_Sanity_Security(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")
	}

	hiConfig := planning.DefaultHiPlanConfig()
	assert.Greater(t, hiConfig.MaxMilestones, 0)
	assert.Greater(t, hiConfig.MaxStepsPerMilestone, 0)
	assert.Greater(t, hiConfig.Timeout, time.Duration(0))
	assert.Greater(t, hiConfig.StepTimeout, time.Duration(0))

	mctsConfig := planning.DefaultMCTSConfig()
	assert.Greater(t, mctsConfig.ExplorationConstant, 0.0)
	assert.Greater(t, mctsConfig.MaxDepth, 0)
	assert.Greater(t, mctsConfig.MaxIterations, 0)
	assert.Greater(t, mctsConfig.Timeout, time.Duration(0))

	totConfig := planning.DefaultTreeOfThoughtsConfig()
	assert.Greater(t, totConfig.MaxDepth, 0)
	assert.Greater(t, totConfig.MaxBranches, 0)
	assert.Greater(t, totConfig.MaxIterations, 0)
	assert.Greater(t, totConfig.Timeout, time.Duration(0))
	assert.Greater(t, totConfig.BeamWidth, 0)
}
