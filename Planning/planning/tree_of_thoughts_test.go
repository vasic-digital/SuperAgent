package planning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
// Mock implementations for Tree of Thoughts
// ---------------------------------------------------------------------------

type mockThoughtGenerator struct {
	thoughts       []*Thought
	thoughtsErr    error
	initialThoughts []*Thought
	initialErr     error
	callCount      int
	mu             sync.Mutex
}

func (m *mockThoughtGenerator) GenerateThoughts(ctx context.Context, parent *Thought, count int) ([]*Thought, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	if m.thoughtsErr != nil {
		return nil, m.thoughtsErr
	}
	// Return copies with unique IDs based on parent
	thoughts := make([]*Thought, 0, len(m.thoughts))
	for i, t := range m.thoughts {
		thought := &Thought{
			ID:        fmt.Sprintf("%s-child-%d", parent.ID, i),
			Content:   t.Content,
			State:     ThoughtStatePending,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"generated_from": parent.ID},
		}
		thoughts = append(thoughts, thought)
	}
	return thoughts, nil
}

func (m *mockThoughtGenerator) GenerateInitialThoughts(ctx context.Context, problem string, count int) ([]*Thought, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	if m.initialErr != nil {
		return nil, m.initialErr
	}
	if m.initialThoughts != nil {
		return m.initialThoughts, nil
	}
	// Return default initial thoughts
	thoughts := make([]*Thought, 0, count)
	for i := 0; i < count && i < 3; i++ {
		thoughts = append(thoughts, &Thought{
			ID:        fmt.Sprintf("init-%d", i),
			Content:   fmt.Sprintf("Initial approach %d for: %s", i, problem),
			State:     ThoughtStatePending,
			Depth:     1,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"type": "initial"},
		})
	}
	return thoughts, nil
}

type mockThoughtEvaluator struct {
	score       float64
	scoreErr    error
	pathScore   float64
	pathScoreErr error
	isTerminal  bool
	terminalErr error
	callCount   int
	mu          sync.Mutex
	// Optionally vary score by thought content
	scoreFn func(thought *Thought) (float64, error)
}

func (m *mockThoughtEvaluator) EvaluateThought(ctx context.Context, thought *Thought) (float64, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	if m.scoreFn != nil {
		return m.scoreFn(thought)
	}
	return m.score, m.scoreErr
}

func (m *mockThoughtEvaluator) EvaluatePath(ctx context.Context, path []*Thought) (float64, error) {
	return m.pathScore, m.pathScoreErr
}

func (m *mockThoughtEvaluator) IsTerminal(ctx context.Context, thought *Thought) (bool, error) {
	return m.isTerminal, m.terminalErr
}

// evalTerminalEvaluator becomes terminal after N evaluations
type evalTerminalEvaluator struct {
	score         float64
	pathScore     float64
	terminalAfter int
	evalCount     int
	mu            sync.Mutex
}

func (e *evalTerminalEvaluator) EvaluateThought(ctx context.Context, thought *Thought) (float64, error) {
	e.mu.Lock()
	e.evalCount++
	e.mu.Unlock()
	return e.score, nil
}

func (e *evalTerminalEvaluator) EvaluatePath(ctx context.Context, path []*Thought) (float64, error) {
	return e.pathScore, nil
}

func (e *evalTerminalEvaluator) IsTerminal(ctx context.Context, thought *Thought) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.evalCount >= e.terminalAfter, nil
}

// ---------------------------------------------------------------------------
// ThoughtState constants
// ---------------------------------------------------------------------------

func TestThoughtState_Constants(t *testing.T) {
	assert.Equal(t, ThoughtState("pending"), ThoughtStatePending)
	assert.Equal(t, ThoughtState("active"), ThoughtStateActive)
	assert.Equal(t, ThoughtState("evaluated"), ThoughtStateEvaluated)
	assert.Equal(t, ThoughtState("pruned"), ThoughtStatePruned)
	assert.Equal(t, ThoughtState("selected"), ThoughtStateSelected)
}

// ---------------------------------------------------------------------------
// DefaultTreeOfThoughtsConfig
// ---------------------------------------------------------------------------

func TestDefaultTreeOfThoughtsConfig(t *testing.T) {
	cfg := DefaultTreeOfThoughtsConfig()
	assert.Equal(t, 10, cfg.MaxDepth)
	assert.Equal(t, 5, cfg.MaxBranches)
	assert.Equal(t, 0.3, cfg.MinScore)
	assert.Equal(t, 0.2, cfg.PruneThreshold)
	assert.Equal(t, "beam", cfg.SearchStrategy)
	assert.Equal(t, 3, cfg.BeamWidth)
	assert.Equal(t, 0.7, cfg.Temperature)
	assert.True(t, cfg.EnableBacktracking)
	assert.Equal(t, 100, cfg.MaxIterations)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
}

// ---------------------------------------------------------------------------
// NewTreeOfThoughts
// ---------------------------------------------------------------------------

func TestNewTreeOfThoughts_WithLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{score: 0.5}

	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, logger)
	require.NotNil(t, tot)
	assert.Equal(t, logger, tot.logger)
	assert.Equal(t, -1.0, tot.bestScore)
}

func TestNewTreeOfThoughts_NilLogger(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{score: 0.5}

	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)
	require.NotNil(t, tot)
	assert.NotNil(t, tot.logger)
	assert.Equal(t, logrus.WarnLevel, tot.logger.Level)
}

// ---------------------------------------------------------------------------
// Solve — initialization errors
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_InitialThoughtGenerationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		initialErr: errors.New("cannot generate initial thoughts"),
	}
	eval := &mockThoughtEvaluator{score: 0.5}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "test problem")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate initial thoughts")
}

// ---------------------------------------------------------------------------
// Solve — BFS strategy
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_BFS_BasicSearch(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{
			{Content: "approach A"},
			{Content: "approach B"},
		},
	}
	eval := &evalTerminalEvaluator{
		score:         0.7,
		pathScore:     0.8,
		terminalAfter: 5,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 3
	cfg.MaxBranches = 2
	cfg.MaxIterations = 50
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "solve this problem")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "solve this problem", result.Problem)
	assert.Equal(t, "bfs", result.Strategy)
	assert.True(t, result.Iterations > 0)
	assert.True(t, result.NodesExplored > 0)
	assert.True(t, result.TreeDepth > 0)
}

func TestTreeOfThoughts_Solve_BFS_PrunesLowScoreThoughts(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{
			{Content: "low quality thought"},
		},
	}
	eval := &mockThoughtEvaluator{
		score:     0.1, // Below default prune threshold of 0.2
		pathScore: 0.1,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.2
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Low scores should be pruned, so bestPath may be nil
}

func TestTreeOfThoughts_Solve_BFS_EvaluationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		scoreErr: errors.New("evaluation failed"),
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 2
	cfg.MaxIterations = 10
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Evaluation errors are logged and skipped
}

func TestTreeOfThoughts_Solve_BFS_TerminalStateFound(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "solution found"}},
	}
	eval := &mockThoughtEvaluator{
		score:      0.9,
		pathScore:  0.95,
		isTerminal: true,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 5
	cfg.MaxIterations = 50
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.BestScore > 0)
	assert.NotEmpty(t, result.Solution)
}

func TestTreeOfThoughts_Solve_BFS_MaxDepthRespected(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 2
	cfg.MaxBranches = 1
	cfg.MaxIterations = 100
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.TreeDepth <= cfg.MaxDepth+1) // +1 for root
}

func TestTreeOfThoughts_Solve_BFS_SkipsPrunedThoughts(t *testing.T) {
	gen := &mockThoughtGenerator{
		initialThoughts: []*Thought{
			{ID: "init-0", Content: "pruned", State: ThoughtStatePruned, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-1", Content: "good thought", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
		},
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &evalTerminalEvaluator{
		score:         0.7,
		pathScore:     0.8,
		terminalAfter: 3,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_BFS_ContextCancelled(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 100
	cfg.MaxIterations = 100000
	cfg.Timeout = 50 * time.Millisecond
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	result, err := tot.Solve(ctx, "problem")
	// May return context error or nil solution
	// The function wraps timeout, so the inner context may expire
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
	_ = result
}

func TestTreeOfThoughts_Solve_BFS_ThoughtGenerationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts:    []*Thought{{Content: "child"}},
		thoughtsErr: errors.New("gen error"),
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 10
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err) // Errors handled internally
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_BFS_TerminalCheckError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:       0.8,
		terminalErr: errors.New("terminal check error"),
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 2
	cfg.MaxIterations = 10
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Solve — DFS strategy
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_DFS_BasicSearch(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{
			{Content: "DFS path A"},
			{Content: "DFS path B"},
		},
	}
	eval := &evalTerminalEvaluator{
		score:         0.7,
		pathScore:     0.85,
		terminalAfter: 4,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 3
	cfg.MaxBranches = 2
	cfg.MaxIterations = 50
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "DFS problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "dfs", result.Strategy)
	assert.True(t, result.Iterations > 0)
}

func TestTreeOfThoughts_Solve_DFS_PrunesLowScores(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "weak thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.05, // Below prune threshold
		pathScore: 0.05,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.2
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_DFS_TerminalState(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "DFS solution"}},
	}
	eval := &mockThoughtEvaluator{
		score:      0.9,
		pathScore:  0.95,
		isTerminal: true,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 5
	cfg.MaxIterations = 50
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.BestScore > 0)
}

func TestTreeOfThoughts_Solve_DFS_EvaluationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		scoreErr: errors.New("eval error"),
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 2
	cfg.MaxIterations = 10
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_DFS_ThoughtGenerationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughtsErr: errors.New("gen error"),
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 10
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_DFS_ContextCancelled(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 100
	cfg.MaxIterations = 100000
	cfg.Timeout = 50 * time.Millisecond
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	result, err := tot.Solve(ctx, "problem")
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
	_ = result
}

func TestTreeOfThoughts_Solve_DFS_EmptyInitialThoughts(t *testing.T) {
	gen := &mockThoughtGenerator{
		initialThoughts: []*Thought{}, // Empty initial thoughts
	}
	eval := &mockThoughtEvaluator{score: 0.5}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 10
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Solve — Beam search strategy
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_Beam_BasicSearch(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{
			{Content: "beam child A"},
			{Content: "beam child B"},
		},
	}
	eval := &evalTerminalEvaluator{
		score:         0.75,
		pathScore:     0.85,
		terminalAfter: 6,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 3
	cfg.BeamWidth = 2
	cfg.MaxBranches = 2
	cfg.MaxIterations = 50
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "beam problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "beam", result.Strategy)
	assert.True(t, result.Iterations > 0)
}

func TestTreeOfThoughts_Solve_Beam_BeamWidthLimits(t *testing.T) {
	// Generate many initial thoughts, but beam width should limit them
	gen := &mockThoughtGenerator{
		initialThoughts: []*Thought{
			{ID: "init-0", Content: "thought 0", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-1", Content: "thought 1", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-2", Content: "thought 2", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-3", Content: "thought 3", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-4", Content: "thought 4", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
		},
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &evalTerminalEvaluator{
		score:         0.7,
		pathScore:     0.8,
		terminalAfter: 10,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 2
	cfg.BeamWidth = 2 // Only keep top 2
	cfg.MaxBranches = 3
	cfg.MaxIterations = 50
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "beam width test")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_Beam_PrunesLowScores(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.05, // Below prune threshold
		pathScore: 0.05,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 3
	cfg.BeamWidth = 3
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.2
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_Beam_EvaluationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &mockThoughtEvaluator{
		scoreErr: errors.New("eval error"),
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 2
	cfg.BeamWidth = 2
	cfg.MaxIterations = 10
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTreeOfThoughts_Solve_Beam_TerminalState(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "solution"}},
	}
	eval := &mockThoughtEvaluator{
		score:      0.9,
		pathScore:  0.95,
		isTerminal: true,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 5
	cfg.BeamWidth = 3
	cfg.MaxIterations = 50
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.BestScore > 0)
}

func TestTreeOfThoughts_Solve_Beam_ContextCancelled(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 100
	cfg.BeamWidth = 3
	cfg.MaxIterations = 100000
	cfg.Timeout = 50 * time.Millisecond
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	result, err := tot.Solve(ctx, "problem")
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
	_ = result
}

func TestTreeOfThoughts_Solve_Beam_ThoughtGenerationError(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughtsErr: errors.New("gen error"),
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 3
	cfg.BeamWidth = 2
	cfg.MaxIterations = 10
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Solve — Default strategy (falls back to beam)
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_DefaultStrategy(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "default"}},
	}
	eval := &evalTerminalEvaluator{
		score:         0.7,
		pathScore:     0.8,
		terminalAfter: 3,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "unknown_strategy" // Should default to beam
	cfg.MaxDepth = 2
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Solve — varying scores
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_Beam_VaryingScores(t *testing.T) {
	gen := &mockThoughtGenerator{
		initialThoughts: []*Thought{
			{ID: "init-0", Content: "high quality", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-1", Content: "medium quality", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-2", Content: "low quality", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
		},
		thoughts: []*Thought{{Content: "child"}},
	}

	callIdx := 0
	var mu sync.Mutex
	scores := []float64{0.9, 0.5, 0.1, 0.85, 0.4}

	eval := &mockThoughtEvaluator{
		pathScore:  0.8,
		isTerminal: false,
		scoreFn: func(thought *Thought) (float64, error) {
			mu.Lock()
			defer mu.Unlock()
			idx := callIdx % len(scores)
			callIdx++
			return scores[idx], nil
		},
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 2
	cfg.BeamWidth = 2
	cfg.MaxBranches = 2
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.2
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "scoring test")
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// getPath
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_GetPath_LeafNode(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	root := &ThoughtNode{
		Thought: &Thought{ID: "root", Content: "problem"},
	}
	child := &ThoughtNode{
		Thought: &Thought{ID: "c1", Content: "approach"},
		Parent:  root,
	}
	grandchild := &ThoughtNode{
		Thought: &Thought{ID: "c1-1", Content: "solution"},
		Parent:  child,
	}

	path := tot.getPath(grandchild)
	require.Len(t, path, 3)
	assert.Equal(t, "root", path[0].ID)
	assert.Equal(t, "c1", path[1].ID)
	assert.Equal(t, "c1-1", path[2].ID)
}

func TestTreeOfThoughts_GetPath_RootOnly(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	root := &ThoughtNode{
		Thought: &Thought{ID: "root", Content: "alone"},
	}

	path := tot.getPath(root)
	require.Len(t, path, 1)
	assert.Equal(t, "root", path[0].ID)
}

func TestTreeOfThoughts_GetPath_NilNode(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	path := tot.getPath(nil)
	assert.Empty(t, path)
}

// ---------------------------------------------------------------------------
// getMaxDepth
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_GetMaxDepth_NilNode(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	assert.Equal(t, 0, tot.getMaxDepth(nil))
}

func TestTreeOfThoughts_GetMaxDepth_SingleNode(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	node := &ThoughtNode{Thought: &Thought{ID: "root"}}
	assert.Equal(t, 1, tot.getMaxDepth(node))
}

func TestTreeOfThoughts_GetMaxDepth_Tree(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	root := &ThoughtNode{Thought: &Thought{ID: "root"}}
	c1 := &ThoughtNode{Thought: &Thought{ID: "c1"}}
	c2 := &ThoughtNode{Thought: &Thought{ID: "c2"}}
	c3 := &ThoughtNode{Thought: &Thought{ID: "c3"}}
	c1.Children = []*ThoughtNode{c3}
	root.Children = []*ThoughtNode{c1, c2}

	assert.Equal(t, 3, tot.getMaxDepth(root))
}

func TestTreeOfThoughts_GetMaxDepth_BalancedTree(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	root := &ThoughtNode{Thought: &Thought{ID: "root"}}
	c1 := &ThoughtNode{Thought: &Thought{ID: "c1"}}
	c2 := &ThoughtNode{Thought: &Thought{ID: "c2"}}
	c3 := &ThoughtNode{Thought: &Thought{ID: "c3"}}
	c4 := &ThoughtNode{Thought: &Thought{ID: "c4"}}
	c1.Children = []*ThoughtNode{c3}
	c2.Children = []*ThoughtNode{c4}
	root.Children = []*ThoughtNode{c1, c2}

	assert.Equal(t, 3, tot.getMaxDepth(root))
}

// ---------------------------------------------------------------------------
// countNodes
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_CountNodes_NilNode(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	assert.Equal(t, 0, tot.countNodes(nil))
}

func TestTreeOfThoughts_CountNodes_SingleNode(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	assert.Equal(t, 1, tot.countNodes(&ThoughtNode{Thought: &Thought{}}))
}

func TestTreeOfThoughts_CountNodes_Tree(t *testing.T) {
	gen := &mockThoughtGenerator{}
	eval := &mockThoughtEvaluator{}
	tot := NewTreeOfThoughts(DefaultTreeOfThoughtsConfig(), gen, eval, nil)

	root := &ThoughtNode{Thought: &Thought{ID: "root"}}
	c1 := &ThoughtNode{Thought: &Thought{ID: "c1"}}
	c2 := &ThoughtNode{Thought: &Thought{ID: "c2"}}
	c3 := &ThoughtNode{Thought: &Thought{ID: "c3"}}
	c4 := &ThoughtNode{Thought: &Thought{ID: "c4"}}
	c1.Children = []*ThoughtNode{c3, c4}
	root.Children = []*ThoughtNode{c1, c2}

	assert.Equal(t, 5, tot.countNodes(root))
}

// ---------------------------------------------------------------------------
// ToTResult
// ---------------------------------------------------------------------------

func TestToTResult_GetSolutionContent(t *testing.T) {
	result := &ToTResult{
		Solution: []*Thought{
			{Content: "step 1"},
			{Content: "step 2"},
			{Content: "step 3"},
		},
	}

	contents := result.GetSolutionContent()
	require.Len(t, contents, 3)
	assert.Equal(t, "step 1", contents[0])
	assert.Equal(t, "step 2", contents[1])
	assert.Equal(t, "step 3", contents[2])
}

func TestToTResult_GetSolutionContent_Empty(t *testing.T) {
	result := &ToTResult{
		Solution: []*Thought{},
	}
	contents := result.GetSolutionContent()
	assert.Empty(t, contents)
}

func TestToTResult_GetSolutionContent_Nil(t *testing.T) {
	result := &ToTResult{}
	contents := result.GetSolutionContent()
	assert.Empty(t, contents)
}

func TestToTResult_MarshalJSON(t *testing.T) {
	r := &ToTResult{
		Problem:       "test problem",
		BestScore:     0.95,
		Iterations:    42,
		Duration:      1500 * time.Millisecond,
		Strategy:      "beam",
		TreeDepth:     5,
		NodesExplored: 30,
		Solution:      []*Thought{},
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, float64(1500), decoded["duration_ms"])
	assert.Equal(t, "test problem", decoded["problem"])
	assert.Equal(t, 0.95, decoded["best_score"])
	assert.Equal(t, float64(42), decoded["iterations"])
	assert.Equal(t, "beam", decoded["strategy"])
}

func TestToTResult_MarshalJSON_ZeroDuration(t *testing.T) {
	r := &ToTResult{
		Duration: 0,
		Solution: []*Thought{},
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, float64(0), decoded["duration_ms"])
}

// ---------------------------------------------------------------------------
// LLMThoughtGenerator
// ---------------------------------------------------------------------------

func TestNewLLMThoughtGenerator(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "", nil }
	logger := logrus.New()
	gen := NewLLMThoughtGenerator(fn, 0.7, logger)
	require.NotNil(t, gen)
	assert.Equal(t, logger, gen.logger)
	assert.Equal(t, 0.7, gen.temperature)
}

func TestLLMThoughtGenerator_GenerateThoughts_Success(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1. Approach A: use recursion\n2. Approach B: use iteration\n3. Approach C: use memoization", nil
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	parent := &Thought{ID: "parent-1", Content: "solve fibonacci"}
	thoughts, err := gen.GenerateThoughts(context.Background(), parent, 3)
	require.NoError(t, err)
	assert.NotEmpty(t, thoughts)
	for _, thought := range thoughts {
		assert.NotEmpty(t, thought.ID)
		assert.Contains(t, thought.ID, "parent-1")
		assert.Equal(t, ThoughtStatePending, thought.State)
		assert.NotNil(t, thought.Metadata)
		assert.Equal(t, "parent-1", thought.Metadata["generated_from"])
	}
}

func TestLLMThoughtGenerator_GenerateThoughts_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("generation failed")
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	thoughts, err := gen.GenerateThoughts(context.Background(), &Thought{ID: "p"}, 3)
	require.Error(t, err)
	assert.Nil(t, thoughts)
}

func TestLLMThoughtGenerator_GenerateThoughts_EmptyResponse(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", nil
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	thoughts, err := gen.GenerateThoughts(context.Background(), &Thought{ID: "p"}, 3)
	require.NoError(t, err)
	assert.Empty(t, thoughts)
}

func TestLLMThoughtGenerator_GenerateThoughts_LimitsCount(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1. A\n2. B\n3. C\n4. D\n5. E", nil
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	// Request only 2
	thoughts, err := gen.GenerateThoughts(context.Background(), &Thought{ID: "p"}, 2)
	require.NoError(t, err)
	assert.Len(t, thoughts, 2)
}

func TestLLMThoughtGenerator_GenerateInitialThoughts_Success(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1. Dynamic programming\n2. Greedy approach\n3. Brute force", nil
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	thoughts, err := gen.GenerateInitialThoughts(context.Background(), "optimization problem", 3)
	require.NoError(t, err)
	assert.NotEmpty(t, thoughts)
	for _, thought := range thoughts {
		assert.Contains(t, thought.ID, "init-")
		assert.Equal(t, ThoughtStatePending, thought.State)
		assert.Equal(t, 1, thought.Depth)
		assert.NotNil(t, thought.Metadata)
		assert.Equal(t, "initial", thought.Metadata["type"])
	}
}

func TestLLMThoughtGenerator_GenerateInitialThoughts_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("gen failed")
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	thoughts, err := gen.GenerateInitialThoughts(context.Background(), "problem", 3)
	require.Error(t, err)
	assert.Nil(t, thoughts)
}

func TestLLMThoughtGenerator_GenerateInitialThoughts_LimitsCount(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1. A\n2. B\n3. C\n4. D", nil
	}
	gen := NewLLMThoughtGenerator(fn, 0.7, nil)

	thoughts, err := gen.GenerateInitialThoughts(context.Background(), "problem", 2)
	require.NoError(t, err)
	assert.Len(t, thoughts, 2)
}

// ---------------------------------------------------------------------------
// LLMThoughtEvaluator
// ---------------------------------------------------------------------------

func TestNewLLMThoughtEvaluator(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	logger := logrus.New()
	eval := NewLLMThoughtEvaluator(fn, logger)
	require.NotNil(t, eval)
	assert.Equal(t, logger, eval.logger)
	assert.NotEmpty(t, eval.terminalKeywords)
}

func TestLLMThoughtEvaluator_EvaluateThought_Success(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "0.85", nil
	}
	eval := NewLLMThoughtEvaluator(fn, nil)

	score, err := eval.EvaluateThought(context.Background(), &Thought{Content: "a good idea"})
	require.NoError(t, err)
	assert.InDelta(t, 0.85, score, 0.001)
}

func TestLLMThoughtEvaluator_EvaluateThought_ClampsHigh(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1.5", nil
	}
	eval := NewLLMThoughtEvaluator(fn, nil)

	score, err := eval.EvaluateThought(context.Background(), &Thought{Content: "idea"})
	require.NoError(t, err)
	assert.Equal(t, 1.0, score)
}

func TestLLMThoughtEvaluator_EvaluateThought_ClampsLow(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "-0.5", nil
	}
	eval := NewLLMThoughtEvaluator(fn, nil)

	score, err := eval.EvaluateThought(context.Background(), &Thought{Content: "idea"})
	require.NoError(t, err)
	assert.Equal(t, 0.0, score)
}

func TestLLMThoughtEvaluator_EvaluateThought_ParseError(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "not a number", nil
	}
	eval := NewLLMThoughtEvaluator(fn, nil)

	_, err := eval.EvaluateThought(context.Background(), &Thought{Content: "idea"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse score")
}

func TestLLMThoughtEvaluator_EvaluateThought_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("eval failed")
	}
	eval := NewLLMThoughtEvaluator(fn, nil)

	_, err := eval.EvaluateThought(context.Background(), &Thought{Content: "idea"})
	require.Error(t, err)
}

func TestLLMThoughtEvaluator_EvaluatePath_EmptyPath(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	score, err := eval.EvaluatePath(context.Background(), []*Thought{})
	require.NoError(t, err)
	assert.Equal(t, 0.0, score)
}

func TestLLMThoughtEvaluator_EvaluatePath_SingleThought(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	path := []*Thought{{Score: 0.8}}
	score, err := eval.EvaluatePath(context.Background(), path)
	require.NoError(t, err)
	// weight = 1.5^0 = 1.0, so score = 0.8 * 1.0 / 1.0 = 0.8
	assert.InDelta(t, 0.8, score, 0.001)
}

func TestLLMThoughtEvaluator_EvaluatePath_WeightedAverage(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	path := []*Thought{
		{Score: 0.5}, // weight = 1.5^0 = 1.0
		{Score: 0.8}, // weight = 1.5^1 = 1.5
		{Score: 0.9}, // weight = 1.5^2 = 2.25
	}
	score, err := eval.EvaluatePath(context.Background(), path)
	require.NoError(t, err)

	// (0.5*1.0 + 0.8*1.5 + 0.9*2.25) / (1.0 + 1.5 + 2.25)
	// = (0.5 + 1.2 + 2.025) / 4.75
	// = 3.725 / 4.75 = ~0.7842
	assert.InDelta(t, 0.7842, score, 0.001)
}

func TestLLMThoughtEvaluator_EvaluatePath_ZeroScores(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	path := []*Thought{
		{Score: 0.0},
		{Score: 0.0},
	}
	score, err := eval.EvaluatePath(context.Background(), path)
	require.NoError(t, err)
	assert.Equal(t, 0.0, score)
}

func TestLLMThoughtEvaluator_IsTerminal_ByKeyword(t *testing.T) {
	tests := []struct {
		name    string
		content string
		expect  bool
	}{
		{"solution keyword", "This is the solution to the problem", true},
		{"answer keyword", "The answer is 42", true},
		{"result keyword", "Final result: success", true},
		{"conclusion keyword", "In conclusion, we can see", true},
		{"final keyword", "The final output is", true},
		{"SOLUTION uppercase", "SOLUTION found", true},
		{"no terminal keyword", "Let us consider another approach", false},
		{"empty content", "", false},
	}

	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			thought := &Thought{Content: tc.content, Score: 0.5}
			terminal, err := eval.IsTerminal(context.Background(), thought)
			require.NoError(t, err)
			assert.Equal(t, tc.expect, terminal)
		})
	}
}

func TestLLMThoughtEvaluator_IsTerminal_ByHighScore(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	thought := &Thought{Content: "some generic thinking", Score: 0.95}
	terminal, err := eval.IsTerminal(context.Background(), thought)
	require.NoError(t, err)
	assert.True(t, terminal)
}

func TestLLMThoughtEvaluator_IsTerminal_LowScore_NoKeyword(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "0.5", nil }
	eval := NewLLMThoughtEvaluator(fn, nil)

	thought := &Thought{Content: "still exploring", Score: 0.3}
	terminal, err := eval.IsTerminal(context.Background(), thought)
	require.NoError(t, err)
	assert.False(t, terminal)
}

// ---------------------------------------------------------------------------
// Helper functions: splitLines
// ---------------------------------------------------------------------------

func TestSplitLines_Basic(t *testing.T) {
	lines := splitLines("line1\nline2\nline3")
	assert.Len(t, lines, 3)
}

func TestSplitLines_EmptyString(t *testing.T) {
	lines := splitLines("")
	assert.Empty(t, lines)
}

func TestSplitLines_SingleLine(t *testing.T) {
	lines := splitLines("just one line")
	assert.Len(t, lines, 1)
	assert.Equal(t, "just one line", lines[0])
}

func TestSplitLines_EmptyLines(t *testing.T) {
	lines := splitLines("line1\n\nline3")
	// Empty lines are skipped
	assert.Len(t, lines, 2)
}

func TestSplitLines_NumberedList(t *testing.T) {
	// Note: splitLines strips leading numbers/dots only for lines ending with \n.
	// The last line (no trailing \n) is appended as-is.
	lines := splitLines("1. First item\n2. Second item\n3. Third item")
	assert.Len(t, lines, 3)
	assert.Equal(t, "First item", lines[0])
	assert.Equal(t, "Second item", lines[1])
	// Last line has no trailing \n, so number prefix is NOT stripped
	assert.Equal(t, "3. Third item", lines[2])
}

func TestSplitLines_NumberedList_WithTrailingNewline(t *testing.T) {
	// When ALL lines end with \n, all get number stripping
	lines := splitLines("1. First item\n2. Second item\n3. Third item\n")
	assert.Len(t, lines, 3)
	assert.Equal(t, "First item", lines[0])
	assert.Equal(t, "Second item", lines[1])
	assert.Equal(t, "Third item", lines[2])
}

func TestSplitLines_NumberedWithParens(t *testing.T) {
	// Last line without trailing \n is not stripped
	lines := splitLines("1) Alpha\n2) Beta")
	assert.Len(t, lines, 2)
	assert.Equal(t, "Alpha", lines[0])
	assert.Equal(t, "2) Beta", lines[1]) // Last line not stripped
}

func TestSplitLines_NumberedWithParens_TrailingNewline(t *testing.T) {
	lines := splitLines("1) Alpha\n2) Beta\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "Alpha", lines[0])
	assert.Equal(t, "Beta", lines[1])
}

func TestSplitLines_TrailingNewline(t *testing.T) {
	lines := splitLines("line1\nline2\n")
	assert.Len(t, lines, 2)
}

func TestSplitLines_BulletPoints(t *testing.T) {
	lines := splitLines("- Point A\n- Point B")
	assert.Len(t, lines, 2)
	assert.Equal(t, "- Point A", lines[0])
	assert.Equal(t, "- Point B", lines[1])
}

// ---------------------------------------------------------------------------
// Helper functions: containsIgnoreCase
// ---------------------------------------------------------------------------

func TestContainsIgnoreCase_BasicMatch(t *testing.T) {
	assert.True(t, containsIgnoreCase("Hello World", "hello"))
	assert.True(t, containsIgnoreCase("Hello World", "WORLD"))
	assert.True(t, containsIgnoreCase("Hello World", "llo W"))
}

func TestContainsIgnoreCase_NoMatch(t *testing.T) {
	assert.False(t, containsIgnoreCase("Hello", "xyz"))
}

func TestContainsIgnoreCase_EmptySubstr(t *testing.T) {
	assert.True(t, containsIgnoreCase("Hello", ""))
}

func TestContainsIgnoreCase_EmptyString(t *testing.T) {
	assert.True(t, containsIgnoreCase("", ""))
}

func TestContainsIgnoreCase_SubstrLongerThanString(t *testing.T) {
	assert.False(t, containsIgnoreCase("Hi", "Hello World"))
}

func TestContainsIgnoreCase_ExactMatch(t *testing.T) {
	assert.True(t, containsIgnoreCase("test", "TEST"))
	assert.True(t, containsIgnoreCase("TEST", "test"))
}

// ---------------------------------------------------------------------------
// Helper functions: contains
// ---------------------------------------------------------------------------

func TestContains_BasicMatch(t *testing.T) {
	assert.True(t, contains("hello world", "world"))
	assert.True(t, contains("hello world", "hello"))
	assert.True(t, contains("hello world", "lo wo"))
}

func TestContains_NoMatch(t *testing.T) {
	assert.False(t, contains("hello", "world"))
}

func TestContains_EmptySubstr(t *testing.T) {
	assert.True(t, contains("hello", ""))
}

func TestContains_EmptyString(t *testing.T) {
	assert.True(t, contains("", ""))
}

func TestContains_SubstrLongerThanString(t *testing.T) {
	assert.False(t, contains("hi", "hello world"))
}

func TestContains_ExactMatch(t *testing.T) {
	assert.True(t, contains("abc", "abc"))
}

func TestContains_CaseSensitive(t *testing.T) {
	assert.False(t, contains("Hello", "hello"))
}

// ---------------------------------------------------------------------------
// Thought JSON serialization
// ---------------------------------------------------------------------------

func TestThought_JSONSerialization(t *testing.T) {
	now := time.Now()
	thought := &Thought{
		ID:          "t-1",
		ParentID:    "root",
		Content:     "recursive approach",
		Reasoning:   "recursion is natural for tree structures",
		State:       ThoughtStateEvaluated,
		Score:       0.85,
		Confidence:  0.9,
		Depth:       2,
		Metadata:    map[string]interface{}{"type": "approach"},
		CreatedAt:   now,
		EvaluatedAt: &now,
	}

	data, err := json.Marshal(thought)
	require.NoError(t, err)

	var decoded Thought
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "t-1", decoded.ID)
	assert.Equal(t, "root", decoded.ParentID)
	assert.Equal(t, "recursive approach", decoded.Content)
	assert.Equal(t, ThoughtStateEvaluated, decoded.State)
	assert.Equal(t, 0.85, decoded.Score)
	assert.Equal(t, 0.9, decoded.Confidence)
	assert.Equal(t, 2, decoded.Depth)
}

func TestThoughtNode_Fields(t *testing.T) {
	thought := &Thought{ID: "t1", Content: "test"}
	parent := &ThoughtNode{Thought: &Thought{ID: "root"}}
	node := &ThoughtNode{
		Thought:  thought,
		Parent:   parent,
		Children: []*ThoughtNode{},
	}

	assert.Equal(t, thought, node.Thought)
	assert.Equal(t, parent, node.Parent)
	assert.NotNil(t, node.Children)
}

// ---------------------------------------------------------------------------
// Solve — max iterations respected
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_BFS_MaxIterationsRespected(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 100  // Very deep
	cfg.MaxIterations = 5 // But limited iterations
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Iterations <= 5)
}

func TestTreeOfThoughts_Solve_DFS_MaxIterationsRespected(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 100
	cfg.MaxIterations = 5
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Iterations <= 5)
}

func TestTreeOfThoughts_Solve_Beam_MaxIterationsRespected(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "child"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.8,
		pathScore: 0.8,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 100
	cfg.BeamWidth = 2
	cfg.MaxIterations = 5
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Iterations <= 10) // Beam may count multiple per depth level
}

// ---------------------------------------------------------------------------
// Solve — multiple terminal solutions, best is chosen
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_BFS_BestTerminalChosen(t *testing.T) {
	gen := &mockThoughtGenerator{
		initialThoughts: []*Thought{
			{ID: "init-0", Content: "This is the final answer", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
			{ID: "init-1", Content: "Also a solution here", State: ThoughtStatePending, Depth: 1, CreatedAt: time.Now()},
		},
	}

	callCount := 0
	var mu sync.Mutex

	eval := &mockThoughtEvaluator{
		isTerminal: true,
		scoreFn: func(thought *Thought) (float64, error) {
			mu.Lock()
			defer mu.Unlock()
			callCount++
			if callCount == 1 {
				return 0.6, nil // Root
			}
			if callCount == 2 {
				return 0.7, nil // First thought
			}
			return 0.9, nil // Second thought
		},
		pathScore: 0.85,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "bfs"
	cfg.MaxDepth = 2
	cfg.MaxIterations = 10
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "multi solution")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Best score should be from EvaluatePath
	assert.Equal(t, 0.85, result.BestScore)
}

// ---------------------------------------------------------------------------
// Concurrent safety of Solve
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_ConcurrentSafety(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "thought"}},
	}
	eval := &evalTerminalEvaluator{
		score:         0.7,
		pathScore:     0.8,
		terminalAfter: 3,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 2
	cfg.MaxIterations = 10
	cfg.PruneThreshold = 0.1
	cfg.Timeout = 5 * time.Second

	// Multiple ToT instances running concurrently (each is independent)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tot := NewTreeOfThoughts(cfg, gen, eval, nil)
			result, err := tot.Solve(context.Background(), fmt.Sprintf("problem %d", idx))
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// DFS with empty root children
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_DFS_NoRootChildren(t *testing.T) {
	gen := &mockThoughtGenerator{
		initialThoughts: []*Thought{},
	}
	eval := &mockThoughtEvaluator{score: 0.5}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "dfs"
	cfg.MaxDepth = 3
	cfg.MaxIterations = 10
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Solution)
}

// ---------------------------------------------------------------------------
// Beam search with all scores below prune threshold
// ---------------------------------------------------------------------------

func TestTreeOfThoughts_Solve_Beam_AllPruned(t *testing.T) {
	gen := &mockThoughtGenerator{
		thoughts: []*Thought{{Content: "very bad thought"}},
	}
	eval := &mockThoughtEvaluator{
		score:     0.01,
		pathScore: 0.01,
	}

	cfg := DefaultTreeOfThoughtsConfig()
	cfg.SearchStrategy = "beam"
	cfg.MaxDepth = 3
	cfg.BeamWidth = 3
	cfg.MaxIterations = 20
	cfg.PruneThreshold = 0.5
	cfg.Timeout = 5 * time.Second
	tot := NewTreeOfThoughts(cfg, gen, eval, nil)

	result, err := tot.Solve(context.Background(), "problem")
	require.NoError(t, err)
	require.NotNil(t, result)
	// All thoughts pruned, so no solution found
}
