package stress

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

// --- mock implementations for stress tests ---

type stressGenerator struct{}

func (g *stressGenerator) GenerateMilestones(_ context.Context, goal string) ([]*planning.Milestone, error) {
	return []*planning.Milestone{
		{
			ID:       "stress-m1",
			Name:     "milestone for " + goal,
			State:    planning.MilestoneStatePending,
			Priority: 0,
			Metadata: make(map[string]interface{}),
		},
	}, nil
}

func (g *stressGenerator) GenerateSteps(_ context.Context, milestone *planning.Milestone) ([]*planning.PlanStep, error) {
	return []*planning.PlanStep{
		{
			ID:          milestone.ID + "-s1",
			MilestoneID: milestone.ID,
			Action:      "action",
			State:       planning.PlanStepStatePending,
			Inputs:      make(map[string]interface{}),
			Outputs:     make(map[string]interface{}),
		},
	}, nil
}

func (g *stressGenerator) GenerateHints(_ context.Context, _ *planning.PlanStep, _ string) ([]string, error) {
	return []string{"hint"}, nil
}

type stressExecutor struct{}

func (e *stressExecutor) Execute(_ context.Context, step *planning.PlanStep, _ []string) (*planning.StepResult, error) {
	return &planning.StepResult{
		Success:  true,
		Outputs:  map[string]interface{}{"ok": true},
		Duration: time.Millisecond,
	}, nil
}

func (e *stressExecutor) Validate(_ context.Context, _ *planning.PlanStep, _ *planning.StepResult) error {
	return nil
}

type stressActionGen struct{}

func (g *stressActionGen) GetActions(_ context.Context, _ interface{}) ([]string, error) {
	return []string{"a", "b"}, nil
}

func (g *stressActionGen) ApplyAction(_ context.Context, state interface{}, action string) (interface{}, error) {
	return fmt.Sprintf("%v->%s", state, action), nil
}

type stressReward struct {
	mu        sync.Mutex
	callCount int
}

func (r *stressReward) Evaluate(_ context.Context, _ interface{}) (float64, error) {
	return 0.6, nil
}

func (r *stressReward) IsTerminal(_ context.Context, _ interface{}) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callCount++
	return r.callCount > 5, nil
}

type stressThoughtGen struct{}

func (g *stressThoughtGen) GenerateThoughts(_ context.Context, parent *planning.Thought, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("%s-c%d", parent.ID, i),
			ParentID:  parent.ID,
			Content:   fmt.Sprintf("thought %d", i),
			State:     planning.ThoughtStatePending,
			Depth:     parent.Depth + 1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{},
		})
	}
	return thoughts, nil
}

func (g *stressThoughtGen) GenerateInitialThoughts(_ context.Context, _ string, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("init-%d", i),
			Content:   fmt.Sprintf("initial %d", i),
			State:     planning.ThoughtStatePending,
			Depth:     1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{},
		})
	}
	return thoughts, nil
}

type stressThoughtEval struct{}

func (e *stressThoughtEval) EvaluateThought(_ context.Context, thought *planning.Thought) (float64, error) {
	if thought.Depth > 2 {
		return 0.95, nil
	}
	return 0.6, nil
}

func (e *stressThoughtEval) EvaluatePath(_ context.Context, path []*planning.Thought) (float64, error) {
	if len(path) == 0 {
		return 0, nil
	}
	total := 0.0
	for _, t := range path {
		total += t.Score
	}
	return total / float64(len(path)), nil
}

func (e *stressThoughtEval) IsTerminal(_ context.Context, thought *planning.Thought) (bool, error) {
	return thought.Score > 0.9, nil
}

// --- Stress Tests ---

func TestHiPlan_ConcurrentPlanCreation_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 75

	gen := &stressGenerator{}
	exec := &stressExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	errors := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := hp.CreatePlan(ctx, fmt.Sprintf("goal-%d", idx))
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}
	assert.Greater(t, successCount, 0, "at least some plans should succeed")
}

func TestHiPlan_ConcurrentPlanExecution_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 50

	gen := &stressGenerator{}
	exec := &stressExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.Timeout = 30 * time.Second
	config.EnableParallelMilestones = false

	ctx := context.Background()

	var wg sync.WaitGroup
	results := make([]*planning.PlanResult, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hp := planning.NewHiPlan(config, gen, exec, nil)
			plan, err := hp.CreatePlan(ctx, fmt.Sprintf("exec-goal-%d", idx))
			if err != nil {
				errs[idx] = err
				return
			}
			result, err := hp.ExecutePlan(ctx, plan)
			results[idx] = result
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	for i := 0; i < goroutines; i++ {
		if errs[i] == nil && results[i] != nil && results[i].Success {
			successCount++
		}
	}
	assert.Equal(t, goroutines, successCount, "all executions should succeed")
}

func TestHiPlan_ConcurrentLibraryAccess_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 100

	config := planning.DefaultHiPlanConfig()
	hp := planning.NewHiPlan(config, &stressGenerator{}, &stressExecutor{}, nil)

	var wg sync.WaitGroup

	// Writers
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			m := &planning.Milestone{
				ID:       fmt.Sprintf("lib-m-%d", idx),
				Name:     fmt.Sprintf("milestone %d", idx),
				State:    planning.MilestoneStateCompleted,
				Metadata: make(map[string]interface{}),
			}
			hp.AddToLibrary(m)
		}(i)
	}

	// Readers
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hp.GetFromLibrary(fmt.Sprintf("lib-m-%d", idx))
		}(i)
	}

	wg.Wait()

	// Verify some entries exist
	found := 0
	for i := 0; i < goroutines/2; i++ {
		if _, ok := hp.GetFromLibrary(fmt.Sprintf("lib-m-%d", i)); ok {
			found++
		}
	}
	assert.Equal(t, goroutines/2, found, "all written milestones should be retrievable")
}

func TestMCTS_ConcurrentSearches_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 60

	var wg sync.WaitGroup
	results := make([]*planning.MCTSResult, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			config := planning.DefaultMCTSConfig()
			config.MaxIterations = 20
			config.MaxDepth = 3
			config.Timeout = 15 * time.Second

			mcts := planning.NewMCTS(config, &stressActionGen{}, &stressReward{}, nil, nil)
			result, err := mcts.Search(context.Background(), fmt.Sprintf("state-%d", idx))
			results[idx] = result
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	for i := 0; i < goroutines; i++ {
		if errs[i] == nil && results[i] != nil {
			successCount++
		}
	}
	assert.Equal(t, goroutines, successCount, "all MCTS searches should complete")
}

func TestMCTSNode_ConcurrentRewardUpdates_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 100

	node := &planning.MCTSNode{
		ID:       "concurrent-node",
		Metadata: make(map[string]interface{}),
	}

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			node.AddReward(float64(idx) * 0.01)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, goroutines, node.Visits, "all visits should be recorded")
	avgReward := node.AverageReward()
	assert.Greater(t, avgReward, 0.0)
}

func TestTreeOfThoughts_ConcurrentSolves_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 50

	var wg sync.WaitGroup
	results := make([]*planning.ToTResult, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			config := planning.DefaultTreeOfThoughtsConfig()
			config.MaxDepth = 3
			config.MaxBranches = 2
			config.MaxIterations = 15
			config.Timeout = 15 * time.Second

			tot := planning.NewTreeOfThoughts(config, &stressThoughtGen{}, &stressThoughtEval{}, nil)
			result, err := tot.Solve(context.Background(), fmt.Sprintf("problem-%d", idx))
			results[idx] = result
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	for i := 0; i < goroutines; i++ {
		if errs[i] == nil && results[i] != nil {
			successCount++
		}
	}
	assert.Equal(t, goroutines, successCount, "all ToT solves should complete")
}

func TestHiPlan_RapidCreateAndExecute_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 80

	gen := &stressGenerator{}
	exec := &stressExecutor{}
	ctx := context.Background()

	var wg sync.WaitGroup
	var mu sync.Mutex
	totalCompleted := 0

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			config := planning.DefaultHiPlanConfig()
			config.Timeout = 15 * time.Second
			hp := planning.NewHiPlan(config, gen, exec, nil)

			plan, err := hp.CreatePlan(ctx, fmt.Sprintf("rapid-%d", idx))
			if err != nil {
				return
			}

			result, err := hp.ExecutePlan(ctx, plan)
			if err == nil && result.Success {
				mu.Lock()
				totalCompleted++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, goroutines, totalCompleted, "all rapid operations should complete")
}

func TestMixedAlgorithms_ConcurrentExecution_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutinesPerAlgo = 25

	ctx := context.Background()
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := map[string]int{
		"hiplan": 0,
		"mcts":   0,
		"tot":    0,
	}

	// HiPlan goroutines
	for i := 0; i < goroutinesPerAlgo; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			config := planning.DefaultHiPlanConfig()
			config.Timeout = 15 * time.Second
			hp := planning.NewHiPlan(config, &stressGenerator{}, &stressExecutor{}, nil)
			plan, err := hp.CreatePlan(ctx, fmt.Sprintf("hi-%d", idx))
			if err != nil {
				return
			}
			result, err := hp.ExecutePlan(ctx, plan)
			if err == nil && result.Success {
				mu.Lock()
				results["hiplan"]++
				mu.Unlock()
			}
		}(i)
	}

	// MCTS goroutines
	for i := 0; i < goroutinesPerAlgo; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			config := planning.DefaultMCTSConfig()
			config.MaxIterations = 10
			config.MaxDepth = 3
			config.Timeout = 15 * time.Second
			mcts := planning.NewMCTS(config, &stressActionGen{}, &stressReward{}, nil, nil)
			result, err := mcts.Search(ctx, fmt.Sprintf("mcts-state-%d", idx))
			if err == nil && result != nil {
				mu.Lock()
				results["mcts"]++
				mu.Unlock()
			}
		}(i)
	}

	// ToT goroutines
	for i := 0; i < goroutinesPerAlgo; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			config := planning.DefaultTreeOfThoughtsConfig()
			config.MaxDepth = 3
			config.MaxBranches = 2
			config.MaxIterations = 10
			config.Timeout = 15 * time.Second
			tot := planning.NewTreeOfThoughts(config, &stressThoughtGen{}, &stressThoughtEval{}, nil)
			result, err := tot.Solve(ctx, fmt.Sprintf("tot-problem-%d", idx))
			if err == nil && result != nil {
				mu.Lock()
				results["tot"]++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	require.Equal(t, goroutinesPerAlgo, results["hiplan"], "all HiPlan should succeed")
	require.Equal(t, goroutinesPerAlgo, results["mcts"], "all MCTS should succeed")
	require.Equal(t, goroutinesPerAlgo, results["tot"], "all ToT should succeed")
}
