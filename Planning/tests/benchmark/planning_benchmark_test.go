package benchmark

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"digital.vasic.planning/planning"
)

// --- mock implementations for benchmarks ---

type benchGenerator struct{}

func (g *benchGenerator) GenerateMilestones(_ context.Context, goal string) ([]*planning.Milestone, error) {
	return []*planning.Milestone{
		{
			ID:       "bench-m1",
			Name:     goal,
			State:    planning.MilestoneStatePending,
			Priority: 0,
			Metadata: make(map[string]interface{}),
		},
	}, nil
}

func (g *benchGenerator) GenerateSteps(_ context.Context, milestone *planning.Milestone) ([]*planning.PlanStep, error) {
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

func (g *benchGenerator) GenerateHints(_ context.Context, _ *planning.PlanStep, _ string) ([]string, error) {
	return []string{"hint"}, nil
}

type benchExecutor struct{}

func (e *benchExecutor) Execute(_ context.Context, _ *planning.PlanStep, _ []string) (*planning.StepResult, error) {
	return &planning.StepResult{
		Success:  true,
		Outputs:  map[string]interface{}{"done": true},
		Duration: time.Microsecond,
	}, nil
}

func (e *benchExecutor) Validate(_ context.Context, _ *planning.PlanStep, _ *planning.StepResult) error {
	return nil
}

type benchActionGen struct{}

func (g *benchActionGen) GetActions(_ context.Context, _ interface{}) ([]string, error) {
	return []string{"act-a", "act-b"}, nil
}

func (g *benchActionGen) ApplyAction(_ context.Context, state interface{}, action string) (interface{}, error) {
	return fmt.Sprintf("%v+%s", state, action), nil
}

type benchReward struct {
	mu    sync.Mutex
	count int
}

func (r *benchReward) Evaluate(_ context.Context, _ interface{}) (float64, error) {
	return 0.7, nil
}

func (r *benchReward) IsTerminal(_ context.Context, _ interface{}) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count++
	return r.count > 5, nil
}

type benchThoughtGen struct{}

func (g *benchThoughtGen) GenerateThoughts(_ context.Context, parent *planning.Thought, count int) ([]*planning.Thought, error) {
	thoughts := make([]*planning.Thought, 0, count)
	for i := 0; i < count; i++ {
		thoughts = append(thoughts, &planning.Thought{
			ID:        fmt.Sprintf("%s-c%d", parent.ID, i),
			ParentID:  parent.ID,
			Content:   fmt.Sprintf("child %d", i),
			State:     planning.ThoughtStatePending,
			Depth:     parent.Depth + 1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{},
		})
	}
	return thoughts, nil
}

func (g *benchThoughtGen) GenerateInitialThoughts(_ context.Context, _ string, count int) ([]*planning.Thought, error) {
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

type benchThoughtEval struct{}

func (e *benchThoughtEval) EvaluateThought(_ context.Context, thought *planning.Thought) (float64, error) {
	if thought.Depth > 2 {
		return 0.95, nil
	}
	return 0.6, nil
}

func (e *benchThoughtEval) EvaluatePath(_ context.Context, path []*planning.Thought) (float64, error) {
	if len(path) == 0 {
		return 0, nil
	}
	s := 0.0
	for _, t := range path {
		s += t.Score
	}
	return s / float64(len(path)), nil
}

func (e *benchThoughtEval) IsTerminal(_ context.Context, thought *planning.Thought) (bool, error) {
	return thought.Score > 0.9, nil
}

// --- Benchmarks ---

func BenchmarkHiPlan_CreatePlan(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	gen := &benchGenerator{}
	exec := &benchExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.Timeout = 30 * time.Second

	hp := planning.NewHiPlan(config, gen, exec, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hp.CreatePlan(ctx, fmt.Sprintf("goal-%d", i))
	}
}

func BenchmarkHiPlan_ExecutePlan(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	gen := &benchGenerator{}
	exec := &benchExecutor{}
	config := planning.DefaultHiPlanConfig()
	config.Timeout = 30 * time.Second
	config.EnableParallelMilestones = false

	hp := planning.NewHiPlan(config, gen, exec, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plan, err := hp.CreatePlan(ctx, fmt.Sprintf("goal-%d", i))
		if err != nil {
			b.Fatal(err)
		}
		_, err = hp.ExecutePlan(ctx, plan)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHiPlan_AddToLibrary(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	hp := planning.NewHiPlan(config, &benchGenerator{}, &benchExecutor{}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := &planning.Milestone{
			ID:       fmt.Sprintf("lib-%d", i),
			Name:     "milestone",
			State:    planning.MilestoneStateCompleted,
			Metadata: make(map[string]interface{}),
		}
		hp.AddToLibrary(m)
	}
}

func BenchmarkHiPlan_GetFromLibrary(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	config := planning.DefaultHiPlanConfig()
	hp := planning.NewHiPlan(config, &benchGenerator{}, &benchExecutor{}, nil)

	for i := 0; i < 1000; i++ {
		hp.AddToLibrary(&planning.Milestone{
			ID:       fmt.Sprintf("lib-%d", i),
			Name:     "milestone",
			State:    planning.MilestoneStateCompleted,
			Metadata: make(map[string]interface{}),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hp.GetFromLibrary(fmt.Sprintf("lib-%d", i%1000))
	}
}

func BenchmarkMCTS_Search(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := planning.DefaultMCTSConfig()
		config.MaxIterations = 20
		config.MaxDepth = 3
		config.Timeout = 30 * time.Second

		mcts := planning.NewMCTS(config, &benchActionGen{}, &benchReward{}, nil, nil)
		_, _ = mcts.Search(context.Background(), "start")
	}
}

func BenchmarkMCTS_UCTValue(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	config := planning.DefaultMCTSConfig()
	mcts := planning.NewMCTS(config, &benchActionGen{}, &benchReward{}, nil, nil)

	node := &planning.MCTSNode{ID: "node", Depth: 5}
	node.AddReward(0.7)
	node.AddReward(0.8)
	node.AddReward(0.6)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mcts.UCTValue(node, 100)
	}
}

func BenchmarkMCTSNode_AddReward(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	node := &planning.MCTSNode{ID: "bench-node", Metadata: make(map[string]interface{})}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.AddReward(0.5)
	}
}

func BenchmarkMCTSNode_AverageReward(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	node := &planning.MCTSNode{ID: "bench-node", Metadata: make(map[string]interface{})}
	for i := 0; i < 1000; i++ {
		node.AddReward(float64(i) * 0.001)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.AverageReward()
	}
}

func BenchmarkTreeOfThoughts_Solve_Beam(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := planning.DefaultTreeOfThoughtsConfig()
		config.SearchStrategy = "beam"
		config.MaxDepth = 3
		config.BeamWidth = 2
		config.MaxBranches = 2
		config.MaxIterations = 20
		config.Timeout = 30 * time.Second

		tot := planning.NewTreeOfThoughts(config, &benchThoughtGen{}, &benchThoughtEval{}, nil)
		_, _ = tot.Solve(context.Background(), "problem")
	}
}

func BenchmarkTreeOfThoughts_Solve_BFS(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := planning.DefaultTreeOfThoughtsConfig()
		config.SearchStrategy = "bfs"
		config.MaxDepth = 3
		config.MaxBranches = 2
		config.MaxIterations = 20
		config.Timeout = 30 * time.Second

		tot := planning.NewTreeOfThoughts(config, &benchThoughtGen{}, &benchThoughtEval{}, nil)
		_, _ = tot.Solve(context.Background(), "problem")
	}
}

func BenchmarkTreeOfThoughts_Solve_DFS(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := planning.DefaultTreeOfThoughtsConfig()
		config.SearchStrategy = "dfs"
		config.MaxDepth = 3
		config.MaxBranches = 2
		config.MaxIterations = 20
		config.Timeout = 30 * time.Second

		tot := planning.NewTreeOfThoughts(config, &benchThoughtGen{}, &benchThoughtEval{}, nil)
		_, _ = tot.Solve(context.Background(), "problem")
	}
}

func BenchmarkDefaultConfigs(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		planning.DefaultHiPlanConfig()
		planning.DefaultMCTSConfig()
		planning.DefaultTreeOfThoughtsConfig()
	}
}
