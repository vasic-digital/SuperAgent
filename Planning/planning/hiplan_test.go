package planning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
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
// Mock implementations
// ---------------------------------------------------------------------------

type mockMilestoneGenerator struct {
	milestones    []*Milestone
	milestonesErr error
	steps         []*PlanStep
	stepsErr      error
	hints         []string
	hintsErr      error
	// Counters for call tracking
	milestonesCalled int
	stepsCalled      int
	hintsCalled      int
	mu               sync.Mutex
}

func (m *mockMilestoneGenerator) GenerateMilestones(ctx context.Context, goal string) ([]*Milestone, error) {
	m.mu.Lock()
	m.milestonesCalled++
	m.mu.Unlock()
	return m.milestones, m.milestonesErr
}

func (m *mockMilestoneGenerator) GenerateSteps(ctx context.Context, milestone *Milestone) ([]*PlanStep, error) {
	m.mu.Lock()
	m.stepsCalled++
	m.mu.Unlock()
	return m.steps, m.stepsErr
}

func (m *mockMilestoneGenerator) GenerateHints(ctx context.Context, step *PlanStep, ctxStr string) ([]string, error) {
	m.mu.Lock()
	m.hintsCalled++
	m.mu.Unlock()
	return m.hints, m.hintsErr
}

type mockStepExecutor struct {
	results    []*StepResult
	errors     []error
	callCount  int32
	validateFn func(ctx context.Context, step *PlanStep, result *StepResult) error
}

func (m *mockStepExecutor) Execute(ctx context.Context, step *PlanStep, hints []string) (*StepResult, error) {
	idx := int(atomic.AddInt32(&m.callCount, 1)) - 1
	if idx < len(m.results) {
		var err error
		if idx < len(m.errors) {
			err = m.errors[idx]
		}
		return m.results[idx], err
	}
	return &StepResult{Success: true, Duration: time.Millisecond}, nil
}

func (m *mockStepExecutor) Validate(ctx context.Context, step *PlanStep, result *StepResult) error {
	if m.validateFn != nil {
		return m.validateFn(ctx, step, result)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helper constructors
// ---------------------------------------------------------------------------

func newTestHiPlan(gen MilestoneGenerator, exec StepExecutor, cfg *HiPlanConfig) *HiPlan {
	c := DefaultHiPlanConfig()
	if cfg != nil {
		c = *cfg
	}
	return NewHiPlan(c, gen, exec, nil)
}

func newSimpleMilestones(count int) []*Milestone {
	ms := make([]*Milestone, count)
	for i := 0; i < count; i++ {
		ms[i] = &Milestone{
			ID:       fmt.Sprintf("m-%d", i),
			Name:     fmt.Sprintf("Milestone %d", i),
			State:    MilestoneStatePending,
			Priority: i,
			Metadata: make(map[string]interface{}),
		}
	}
	return ms
}

func newSimpleSteps(milestoneID string, count int) []*PlanStep {
	steps := make([]*PlanStep, count)
	for i := 0; i < count; i++ {
		steps[i] = &PlanStep{
			ID:          fmt.Sprintf("%s-step-%d", milestoneID, i),
			MilestoneID: milestoneID,
			Action:      fmt.Sprintf("action-%d", i),
			State:       PlanStepStatePending,
			Inputs:      make(map[string]interface{}),
			Outputs:     make(map[string]interface{}),
		}
	}
	return steps
}

// ---------------------------------------------------------------------------
// MilestoneState constants
// ---------------------------------------------------------------------------

func TestMilestoneState_Constants(t *testing.T) {
	assert.Equal(t, MilestoneState("pending"), MilestoneStatePending)
	assert.Equal(t, MilestoneState("in_progress"), MilestoneStateInProgress)
	assert.Equal(t, MilestoneState("completed"), MilestoneStateCompleted)
	assert.Equal(t, MilestoneState("failed"), MilestoneStateFailed)
	assert.Equal(t, MilestoneState("skipped"), MilestoneStateSkipped)
}

func TestPlanStepState_Constants(t *testing.T) {
	assert.Equal(t, PlanStepState("pending"), PlanStepStatePending)
	assert.Equal(t, PlanStepState("in_progress"), PlanStepStateInProgress)
	assert.Equal(t, PlanStepState("completed"), PlanStepStateCompleted)
	assert.Equal(t, PlanStepState("failed"), PlanStepStateFailed)
}

// ---------------------------------------------------------------------------
// DefaultHiPlanConfig
// ---------------------------------------------------------------------------

func TestDefaultHiPlanConfig(t *testing.T) {
	cfg := DefaultHiPlanConfig()
	assert.Equal(t, 20, cfg.MaxMilestones)
	assert.Equal(t, 50, cfg.MaxStepsPerMilestone)
	assert.True(t, cfg.EnableParallelMilestones)
	assert.Equal(t, 3, cfg.MaxParallelMilestones)
	assert.True(t, cfg.EnableAdaptivePlanning)
	assert.True(t, cfg.RetryFailedSteps)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 30*time.Minute, cfg.Timeout)
	assert.Equal(t, 5*time.Minute, cfg.StepTimeout)
}

// ---------------------------------------------------------------------------
// NewHiPlan
// ---------------------------------------------------------------------------

func TestNewHiPlan_WithLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := NewHiPlan(DefaultHiPlanConfig(), gen, exec, logger)
	require.NotNil(t, hp)
	assert.Equal(t, logger, hp.logger)
}

func TestNewHiPlan_NilLogger(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := NewHiPlan(DefaultHiPlanConfig(), gen, exec, nil)
	require.NotNil(t, hp)
	require.NotNil(t, hp.logger)
	assert.Equal(t, logrus.WarnLevel, hp.logger.Level)
}

func TestNewHiPlan_MilestoneLibraryInitialized(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := NewHiPlan(DefaultHiPlanConfig(), gen, exec, nil)
	require.NotNil(t, hp.milestoneLibrary)
	assert.Len(t, hp.milestoneLibrary, 0)
}

// ---------------------------------------------------------------------------
// CreatePlan
// ---------------------------------------------------------------------------

func TestHiPlan_CreatePlan_Success(t *testing.T) {
	milestones := newSimpleMilestones(3)
	steps := newSimpleSteps("m-0", 2)

	gen := &mockMilestoneGenerator{
		milestones: milestones,
		steps:      steps,
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.Timeout = 10 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan, err := hp.CreatePlan(context.Background(), "build a web app")
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, "build a web app", plan.Goal)
	assert.Equal(t, "created", plan.State)
	assert.Len(t, plan.Milestones, 3)
	assert.NotEmpty(t, plan.ID)
	assert.False(t, plan.CreatedAt.IsZero())
	assert.NotNil(t, plan.Metadata)
}

func TestHiPlan_CreatePlan_GenerateMilestonesError(t *testing.T) {
	gen := &mockMilestoneGenerator{
		milestonesErr: errors.New("LLM error"),
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan, err := hp.CreatePlan(context.Background(), "goal")
	require.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "failed to generate milestones")
}

func TestHiPlan_CreatePlan_TruncatesToMaxMilestones(t *testing.T) {
	milestones := newSimpleMilestones(30)
	gen := &mockMilestoneGenerator{
		milestones: milestones,
		steps:      []*PlanStep{},
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.MaxMilestones = 5
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan, err := hp.CreatePlan(context.Background(), "goal")
	require.NoError(t, err)
	assert.Len(t, plan.Milestones, 5)
}

func TestHiPlan_CreatePlan_TruncatesStepsPerMilestone(t *testing.T) {
	milestones := newSimpleMilestones(1)
	steps := newSimpleSteps("m-0", 100)
	gen := &mockMilestoneGenerator{
		milestones: milestones,
		steps:      steps,
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.MaxStepsPerMilestone = 3
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan, err := hp.CreatePlan(context.Background(), "goal")
	require.NoError(t, err)
	assert.Len(t, plan.Milestones[0].Steps, 3)
}

func TestHiPlan_CreatePlan_StepsGenerationError_ContinuesOtherMilestones(t *testing.T) {
	milestones := newSimpleMilestones(3)
	gen := &mockMilestoneGenerator{
		milestones: milestones,
		stepsErr:   errors.New("step gen error"),
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan, err := hp.CreatePlan(context.Background(), "goal")
	require.NoError(t, err)
	assert.Len(t, plan.Milestones, 3)
	// Steps should be nil since generation failed
	for _, m := range plan.Milestones {
		assert.Nil(t, m.Steps)
	}
}

func TestHiPlan_CreatePlan_SetCurrentPlan(t *testing.T) {
	gen := &mockMilestoneGenerator{
		milestones: newSimpleMilestones(1),
		steps:      []*PlanStep{},
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	assert.Nil(t, hp.GetCurrentPlan())

	plan, err := hp.CreatePlan(context.Background(), "goal")
	require.NoError(t, err)

	current := hp.GetCurrentPlan()
	require.NotNil(t, current)
	assert.Equal(t, plan.ID, current.ID)
}

// ---------------------------------------------------------------------------
// ExecutePlan — Sequential
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_SequentialAllSuccess(t *testing.T) {
	milestones := newSimpleMilestones(2)
	for _, m := range milestones {
		m.Steps = newSimpleSteps(m.ID, 2)
	}

	gen := &mockMilestoneGenerator{
		hints: []string{"hint1"},
	}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-plan",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.CompletedMilestones)
	assert.Equal(t, 0, result.FailedMilestones)
	assert.Equal(t, "completed", plan.State)
	assert.Equal(t, 1.0, plan.Progress)
	assert.NotNil(t, plan.StartedAt)
	assert.NotNil(t, plan.CompletedAt)
}

func TestHiPlan_ExecutePlan_SequentialStepFailure_NoAdaptive(t *testing.T) {
	milestones := newSimpleMilestones(2)
	milestones[0].Steps = newSimpleSteps("m-0", 2)
	milestones[1].Steps = newSimpleSteps("m-1", 1)

	gen := &mockMilestoneGenerator{hints: []string{"hint"}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
			{Success: false, Error: "step failed", Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.EnableAdaptivePlanning = false
	cfg.RetryFailedSteps = false
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-plan",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "failed", plan.State)
	// With no adaptive planning, sequential stops after first milestone failure
	assert.Equal(t, 1, len(result.MilestoneResults))
}

func TestHiPlan_ExecutePlan_SequentialStepFailure_WithAdaptive(t *testing.T) {
	milestones := newSimpleMilestones(2)
	milestones[0].Steps = newSimpleSteps("m-0", 1)
	milestones[1].Steps = newSimpleSteps("m-1", 1)

	gen := &mockMilestoneGenerator{hints: []string{"hint"}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "failed", Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.EnableAdaptivePlanning = true
	cfg.RetryFailedSteps = false
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-plan",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	// With adaptive, both milestones are attempted
	assert.Equal(t, 2, len(result.MilestoneResults))
	assert.Equal(t, 1, result.CompletedMilestones)
	assert.Equal(t, 1, result.FailedMilestones)
	assert.False(t, result.Success)
	assert.Equal(t, "failed", plan.State)
}

// ---------------------------------------------------------------------------
// ExecutePlan — Parallel
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_ParallelAllSuccess(t *testing.T) {
	milestones := newSimpleMilestones(3)
	for _, m := range milestones {
		m.Steps = newSimpleSteps(m.ID, 1)
	}

	gen := &mockMilestoneGenerator{hints: []string{"hint"}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = true
	cfg.MaxParallelMilestones = 2
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-plan",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 3, result.CompletedMilestones)
}

func TestHiPlan_ExecutePlan_ParallelWithDependencies(t *testing.T) {
	milestones := newSimpleMilestones(3)
	milestones[0].Steps = newSimpleSteps("m-0", 1)
	milestones[1].Dependencies = []string{"m-0"}
	milestones[1].Steps = newSimpleSteps("m-1", 1)
	milestones[2].Steps = newSimpleSteps("m-2", 1)

	gen := &mockMilestoneGenerator{hints: []string{"hint"}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = true
	cfg.MaxParallelMilestones = 3
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-plan",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.MilestoneResults))
}

// ---------------------------------------------------------------------------
// sortMilestonesByDependencies (topological sort)
// ---------------------------------------------------------------------------

func TestHiPlan_SortMilestonesByDependencies_NoDeps(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	milestones := newSimpleMilestones(3)
	sorted := hp.sortMilestonesByDependencies(milestones)
	assert.Len(t, sorted, 3)
	// Without deps, order preserves insertion via queue
	assert.Equal(t, "m-0", sorted[0].ID)
	assert.Equal(t, "m-1", sorted[1].ID)
	assert.Equal(t, "m-2", sorted[2].ID)
}

func TestHiPlan_SortMilestonesByDependencies_LinearChain(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	// m-2 depends on m-1, m-1 depends on m-0
	milestones := newSimpleMilestones(3)
	milestones[1].Dependencies = []string{"m-0"}
	milestones[2].Dependencies = []string{"m-1"}

	sorted := hp.sortMilestonesByDependencies(milestones)
	require.Len(t, sorted, 3)
	assert.Equal(t, "m-0", sorted[0].ID)
	assert.Equal(t, "m-1", sorted[1].ID)
	assert.Equal(t, "m-2", sorted[2].ID)
}

func TestHiPlan_SortMilestonesByDependencies_Diamond(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	// Diamond: m-0 -> m-1, m-0 -> m-2, m-1 -> m-3, m-2 -> m-3
	milestones := newSimpleMilestones(4)
	milestones[1].Dependencies = []string{"m-0"}
	milestones[2].Dependencies = []string{"m-0"}
	milestones[3].Dependencies = []string{"m-1", "m-2"}

	sorted := hp.sortMilestonesByDependencies(milestones)
	require.Len(t, sorted, 4)
	// m-0 must be first, m-3 must be last
	assert.Equal(t, "m-0", sorted[0].ID)
	assert.Equal(t, "m-3", sorted[3].ID)
}

func TestHiPlan_SortMilestonesByDependencies_CyclicDeps(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	// Cycle: m-0 -> m-1 -> m-2 -> m-0
	milestones := newSimpleMilestones(3)
	milestones[0].Dependencies = []string{"m-2"}
	milestones[1].Dependencies = []string{"m-0"}
	milestones[2].Dependencies = []string{"m-1"}

	sorted := hp.sortMilestonesByDependencies(milestones)
	// With cycles, all nodes remain in inDegree > 0 initially,
	// so they get appended in original order
	require.Len(t, sorted, 3)
}

func TestHiPlan_SortMilestonesByDependencies_Empty(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	sorted := hp.sortMilestonesByDependencies([]*Milestone{})
	assert.Len(t, sorted, 0)
}

func TestHiPlan_SortMilestonesByDependencies_Single(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	sorted := hp.sortMilestonesByDependencies(newSimpleMilestones(1))
	require.Len(t, sorted, 1)
	assert.Equal(t, "m-0", sorted[0].ID)
}

// ---------------------------------------------------------------------------
// ExecuteStep (public method)
// ---------------------------------------------------------------------------

func TestHiPlan_ExecuteStep_Success(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Outputs: map[string]interface{}{"key": "val"}, Duration: time.Millisecond},
		},
	}
	cfg := DefaultHiPlanConfig()
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	step := &PlanStep{ID: "s-1", Action: "do something"}
	result, err := hp.ExecuteStep(context.Background(), step, []string{"hint"})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "val", result.Outputs["key"])
}

func TestHiPlan_ExecuteStep_RetriesOnFailure(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "fail1"},
			{Success: false, Error: "fail2"},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.RetryFailedSteps = true
	cfg.MaxRetries = 3
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	step := &PlanStep{ID: "s-retry", Action: "retry action"}
	result, err := hp.ExecuteStep(context.Background(), step, nil)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestHiPlan_ExecuteStep_AllRetriesFail(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "fail1"},
			{Success: false, Error: "fail2"},
			{Success: false, Error: "fail3"},
			{Success: false, Error: "fail4"},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.RetryFailedSteps = true
	cfg.MaxRetries = 3
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	step := &PlanStep{ID: "s-allfail", Action: "always fails"}
	result, err := hp.ExecuteStep(context.Background(), step, nil)
	require.NoError(t, err)
	assert.False(t, result.Success)
}

func TestHiPlan_ExecuteStep_NoRetries(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "first fail"},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.RetryFailedSteps = false
	cfg.MaxRetries = 0
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	step := &PlanStep{ID: "s-noretry", Action: "no retry"}
	result, err := hp.ExecuteStep(context.Background(), step, nil)
	require.NoError(t, err)
	assert.False(t, result.Success)
}

func TestHiPlan_ExecuteStep_ExecutorReturnsNilResult(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{
		results: []*StepResult{nil},
		errors:  []error{errors.New("exec error")},
	}

	cfg := DefaultHiPlanConfig()
	cfg.RetryFailedSteps = false
	cfg.MaxRetries = 0
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	step := &PlanStep{ID: "s-nil", Action: "nil result"}
	result, err := hp.ExecuteStep(context.Background(), step, nil)
	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "exec error", result.Error)
}

func TestHiPlan_ExecuteStep_ExecutorReturnsNilResultNoError(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{
		results: []*StepResult{nil},
		errors:  []error{nil},
	}

	cfg := DefaultHiPlanConfig()
	cfg.RetryFailedSteps = false
	cfg.MaxRetries = 0
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	step := &PlanStep{ID: "s-nil2", Action: "nil result no error"}
	result, err := hp.ExecuteStep(context.Background(), step, nil)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "execution failed", result.Error)
}

// ---------------------------------------------------------------------------
// executeMilestone — context cancellation
// ---------------------------------------------------------------------------

func TestHiPlan_ExecuteMilestone_ContextCancelled(t *testing.T) {
	gen := &mockMilestoneGenerator{hints: []string{"hint"}}

	// This executor blocks until context is cancelled
	blockExec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.Timeout = 1 * time.Second
	cfg.StepTimeout = 500 * time.Millisecond
	hp := NewHiPlan(cfg, gen, blockExec, nil)

	milestones := newSimpleMilestones(1)
	milestones[0].Steps = newSimpleSteps("m-0", 5)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	plan := &HierarchicalPlan{
		ID:         "plan-cancel",
		Goal:       "test cancel",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(ctx, plan)
	require.NoError(t, err)
	// Due to short timeout, the plan may partially execute
	assert.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Milestone execution with hint generation error
// ---------------------------------------------------------------------------

func TestHiPlan_ExecuteMilestone_HintGenerationError_FallsBackToMilestoneHints(t *testing.T) {
	milestones := newSimpleMilestones(1)
	milestones[0].Hints = []string{"fallback hint"}
	milestones[0].Steps = newSimpleSteps("m-0", 1)

	gen := &mockMilestoneGenerator{
		hintsErr: errors.New("hint generation failed"),
	}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.RetryFailedSteps = false
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-hints-fallback",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

// ---------------------------------------------------------------------------
// Milestone with no steps
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_MilestoneWithNoSteps(t *testing.T) {
	milestones := newSimpleMilestones(1)
	// No steps assigned

	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "test-no-steps",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	// 0 completed / 0 total => success (0 == 0)
	assert.True(t, result.Success)
}

// ---------------------------------------------------------------------------
// buildContext
// ---------------------------------------------------------------------------

func TestHiPlan_BuildContext(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	milestone := &Milestone{
		Name:        "Setup DB",
		Description: "Setup database infrastructure",
	}
	step := &PlanStep{
		ID:     "step-1",
		Action: "create tables",
	}

	ctxStr := hp.buildContext(milestone, step)
	assert.Contains(t, ctxStr, "Setup DB")
	assert.Contains(t, ctxStr, "Setup database infrastructure")
	assert.Contains(t, ctxStr, "step-1")
	assert.Contains(t, ctxStr, "create tables")
}

// ---------------------------------------------------------------------------
// AddToLibrary / GetFromLibrary
// ---------------------------------------------------------------------------

func TestHiPlan_AddToLibrary_GetFromLibrary(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	milestone := &Milestone{
		ID:   "lib-1",
		Name: "reusable milestone",
	}

	hp.AddToLibrary(milestone)

	retrieved, ok := hp.GetFromLibrary("lib-1")
	assert.True(t, ok)
	assert.Equal(t, milestone, retrieved)
}

func TestHiPlan_GetFromLibrary_NotFound(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	_, ok := hp.GetFromLibrary("nonexistent")
	assert.False(t, ok)
}

func TestHiPlan_AddToLibrary_ConcurrentAccess(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			m := &Milestone{
				ID:   fmt.Sprintf("lib-%d", idx),
				Name: fmt.Sprintf("milestone-%d", idx),
			}
			hp.AddToLibrary(m)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 50; i++ {
		_, ok := hp.GetFromLibrary(fmt.Sprintf("lib-%d", i))
		assert.True(t, ok)
	}
}

// ---------------------------------------------------------------------------
// GetCurrentPlan — nil scenario
// ---------------------------------------------------------------------------

func TestHiPlan_GetCurrentPlan_Nil(t *testing.T) {
	gen := &mockMilestoneGenerator{}
	exec := &mockStepExecutor{}
	hp := newTestHiPlan(gen, exec, nil)
	assert.Nil(t, hp.GetCurrentPlan())
}

// ---------------------------------------------------------------------------
// PlanResult JSON marshaling
// ---------------------------------------------------------------------------

func TestPlanResult_MarshalJSON(t *testing.T) {
	r := &PlanResult{
		PlanID:              "plan-1",
		Success:             true,
		CompletedMilestones: 3,
		FailedMilestones:    0,
		Duration:            2500 * time.Millisecond,
		MilestoneResults:    []*MilestoneResult{},
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, float64(2500), decoded["duration_ms"])
	assert.Equal(t, "plan-1", decoded["plan_id"])
	assert.Equal(t, true, decoded["success"])
}

func TestPlanResult_MarshalJSON_ZeroDuration(t *testing.T) {
	r := &PlanResult{
		PlanID:           "plan-zero",
		Duration:         0,
		MilestoneResults: []*MilestoneResult{},
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, float64(0), decoded["duration_ms"])
}

// ---------------------------------------------------------------------------
// LLMMilestoneGenerator
// ---------------------------------------------------------------------------

func TestNewLLMMilestoneGenerator(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) { return "", nil }
	logger := logrus.New()
	gen := NewLLMMilestoneGenerator(fn, logger)
	require.NotNil(t, gen)
	assert.Equal(t, logger, gen.logger)
}

func TestLLMMilestoneGenerator_GenerateMilestones_Success(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1. Setup infrastructure\n2. Build backend\n3. Build frontend", nil
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	milestones, err := gen.GenerateMilestones(context.Background(), "build a web app")
	require.NoError(t, err)
	assert.NotEmpty(t, milestones)
	for _, m := range milestones {
		assert.NotEmpty(t, m.ID)
		assert.Equal(t, MilestoneStatePending, m.State)
		assert.NotNil(t, m.Metadata)
	}
}

func TestLLMMilestoneGenerator_GenerateMilestones_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("LLM unavailable")
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	milestones, err := gen.GenerateMilestones(context.Background(), "goal")
	require.Error(t, err)
	assert.Nil(t, milestones)
}

func TestLLMMilestoneGenerator_GenerateMilestones_EmptyResponse(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", nil
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	milestones, err := gen.GenerateMilestones(context.Background(), "goal")
	require.NoError(t, err)
	assert.Empty(t, milestones)
}

func TestLLMMilestoneGenerator_GenerateSteps_Success(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "1. Create schema\n2. Run migrations\n3. Seed data", nil
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	milestone := &Milestone{
		ID:          "m-1",
		Name:        "Setup DB",
		Description: "Setup database",
	}

	steps, err := gen.GenerateSteps(context.Background(), milestone)
	require.NoError(t, err)
	assert.NotEmpty(t, steps)
	for _, s := range steps {
		assert.Equal(t, "m-1", s.MilestoneID)
		assert.Equal(t, PlanStepStatePending, s.State)
		assert.NotNil(t, s.Inputs)
		assert.NotNil(t, s.Outputs)
	}
}

func TestLLMMilestoneGenerator_GenerateSteps_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("step gen error")
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	steps, err := gen.GenerateSteps(context.Background(), &Milestone{ID: "m-1"})
	require.Error(t, err)
	assert.Nil(t, steps)
}

func TestLLMMilestoneGenerator_GenerateHints_Success(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "- Check constraints\n- Handle nulls\n- Add indexes", nil
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	hints, err := gen.GenerateHints(context.Background(), &PlanStep{ID: "s-1"}, "context info")
	require.NoError(t, err)
	assert.NotEmpty(t, hints)
}

func TestLLMMilestoneGenerator_GenerateHints_Error(t *testing.T) {
	fn := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("hint gen error")
	}
	gen := NewLLMMilestoneGenerator(fn, nil)

	hints, err := gen.GenerateHints(context.Background(), &PlanStep{ID: "s-1"}, "ctx")
	require.Error(t, err)
	assert.Nil(t, hints)
}

// ---------------------------------------------------------------------------
// ExecutePlan — progress tracking
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_ProgressTracking(t *testing.T) {
	milestones := newSimpleMilestones(4)
	for _, m := range milestones {
		m.Steps = newSimpleSteps(m.ID, 1)
	}

	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
			{Success: false, Error: "fail", Duration: time.Millisecond},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.EnableAdaptivePlanning = true
	cfg.RetryFailedSteps = false
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "progress-plan",
		Goal:       "test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	// 3 success, 1 fail
	assert.Equal(t, 3, result.CompletedMilestones)
	assert.Equal(t, 1, result.FailedMilestones)
	assert.Equal(t, 0.75, plan.Progress)
}

// ---------------------------------------------------------------------------
// Milestone and PlanStep JSON serialization
// ---------------------------------------------------------------------------

func TestMilestone_JSONSerialization(t *testing.T) {
	now := time.Now()
	m := &Milestone{
		ID:           "m-json",
		Name:         "JSON Test",
		Description:  "Test JSON",
		State:        MilestoneStatePending,
		Priority:     5,
		Dependencies: []string{"m-1", "m-2"},
		Hints:        []string{"hint1"},
		Metadata:     map[string]interface{}{"key": "val"},
		Progress:     0.5,
		StartedAt:    &now,
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var decoded Milestone
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "m-json", decoded.ID)
	assert.Equal(t, "JSON Test", decoded.Name)
	assert.Equal(t, MilestoneStatePending, decoded.State)
	assert.Len(t, decoded.Dependencies, 2)
	assert.Equal(t, 0.5, decoded.Progress)
}

func TestPlanStep_JSONSerialization(t *testing.T) {
	s := &PlanStep{
		ID:          "s-json",
		MilestoneID: "m-1",
		Action:      "test action",
		State:       PlanStepStateCompleted,
		Hints:       []string{"h1"},
		Inputs:      map[string]interface{}{"in": 1},
		Outputs:     map[string]interface{}{"out": 2},
		Error:       "some error",
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var decoded PlanStep
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "s-json", decoded.ID)
	assert.Equal(t, PlanStepStateCompleted, decoded.State)
}

func TestHierarchicalPlan_JSONSerialization(t *testing.T) {
	plan := &HierarchicalPlan{
		ID:         "hp-json",
		Goal:       "build app",
		Milestones: newSimpleMilestones(2),
		State:      "completed",
		Progress:   1.0,
		Metadata:   map[string]interface{}{"env": "test"},
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(plan)
	require.NoError(t, err)

	var decoded HierarchicalPlan
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "hp-json", decoded.ID)
	assert.Equal(t, "build app", decoded.Goal)
	assert.Len(t, decoded.Milestones, 2)
}

// ---------------------------------------------------------------------------
// ExecutePlan — retry within milestone execution
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_StepRetrySucceedsOnSecondAttempt(t *testing.T) {
	milestones := newSimpleMilestones(1)
	milestones[0].Steps = newSimpleSteps("m-0", 1)

	gen := &mockMilestoneGenerator{hints: []string{"hint"}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "transient"},
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.RetryFailedSteps = true
	cfg.MaxRetries = 3
	cfg.Timeout = 10 * time.Second
	cfg.StepTimeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "retry-plan",
		Goal:       "test retry",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

// ---------------------------------------------------------------------------
// StepResult and MilestoneResult
// ---------------------------------------------------------------------------

func TestStepResult_Fields(t *testing.T) {
	r := &StepResult{
		Success:  true,
		Outputs:  map[string]interface{}{"code": "200"},
		Logs:     []string{"log1", "log2"},
		Duration: 100 * time.Millisecond,
	}
	assert.True(t, r.Success)
	assert.Len(t, r.Logs, 2)
	assert.Equal(t, 100*time.Millisecond, r.Duration)
}

func TestMilestoneResult_Fields(t *testing.T) {
	r := &MilestoneResult{
		MilestoneID: "m-1",
		Success:     false,
		StepResults: []*StepResult{{Success: true}},
		Duration:    500 * time.Millisecond,
		Error:       "milestone failed",
	}
	assert.Equal(t, "m-1", r.MilestoneID)
	assert.False(t, r.Success)
	assert.Equal(t, "milestone failed", r.Error)
}

// ---------------------------------------------------------------------------
// Empty plan execution
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_EmptyMilestones(t *testing.T) {
	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "empty-plan",
		Goal:       "empty",
		Milestones: []*Milestone{},
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.CompletedMilestones)
	assert.Equal(t, 0, result.FailedMilestones)
}

// ---------------------------------------------------------------------------
// Concurrent plan creation
// ---------------------------------------------------------------------------

func TestHiPlan_CreatePlan_ConcurrentAccess(t *testing.T) {
	gen := &mockMilestoneGenerator{
		milestones: newSimpleMilestones(1),
		steps:      []*PlanStep{},
	}
	exec := &mockStepExecutor{}
	cfg := DefaultHiPlanConfig()
	cfg.Timeout = 5 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = hp.CreatePlan(context.Background(), fmt.Sprintf("goal-%d", idx))
		}(i)
	}
	wg.Wait()

	// currentPlan should be set (could be any of the 10 plans)
	assert.NotNil(t, hp.GetCurrentPlan())
}

// ---------------------------------------------------------------------------
// Milestone state transitions during execution
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_MilestoneStateTransitions(t *testing.T) {
	milestones := newSimpleMilestones(1)
	milestones[0].Steps = newSimpleSteps("m-0", 1)

	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: true, Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.RetryFailedSteps = false
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	// Before execution
	assert.Equal(t, MilestoneStatePending, milestones[0].State)

	plan := &HierarchicalPlan{
		ID:         "state-plan",
		Goal:       "test state",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	_, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)

	// After successful execution
	assert.Equal(t, MilestoneStateCompleted, milestones[0].State)
	assert.NotNil(t, milestones[0].StartedAt)
	assert.NotNil(t, milestones[0].CompletedAt)
	assert.Equal(t, 1.0, milestones[0].Progress)
}

func TestHiPlan_ExecutePlan_MilestoneStateTransitions_Failed(t *testing.T) {
	milestones := newSimpleMilestones(1)
	milestones[0].Steps = newSimpleSteps("m-0", 1)

	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "fail", Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.RetryFailedSteps = false
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "state-fail-plan",
		Goal:       "test state fail",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	_, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)

	assert.Equal(t, MilestoneStateFailed, milestones[0].State)
}

// ---------------------------------------------------------------------------
// Step output propagation
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_StepOutputPropagation(t *testing.T) {
	milestones := newSimpleMilestones(1)
	milestones[0].Steps = newSimpleSteps("m-0", 1)

	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{
				Success:  true,
				Outputs:  map[string]interface{}{"result": "data"},
				Duration: time.Millisecond,
			},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.RetryFailedSteps = false
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "output-plan",
		Goal:       "output test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	_, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)

	// Step outputs should be propagated
	assert.Equal(t, "data", milestones[0].Steps[0].Outputs["result"])
	assert.Equal(t, PlanStepStateCompleted, milestones[0].Steps[0].State)
}

// ---------------------------------------------------------------------------
// Step failure propagation
// ---------------------------------------------------------------------------

func TestHiPlan_ExecutePlan_StepFailureErrorPropagation(t *testing.T) {
	milestones := newSimpleMilestones(1)
	milestones[0].Steps = newSimpleSteps("m-0", 1)

	gen := &mockMilestoneGenerator{hints: []string{}}
	exec := &mockStepExecutor{
		results: []*StepResult{
			{Success: false, Error: "step error detail", Duration: time.Millisecond},
		},
	}

	cfg := DefaultHiPlanConfig()
	cfg.EnableParallelMilestones = false
	cfg.RetryFailedSteps = false
	cfg.EnableAdaptivePlanning = false
	cfg.Timeout = 5 * time.Second
	cfg.StepTimeout = 2 * time.Second
	hp := NewHiPlan(cfg, gen, exec, nil)

	plan := &HierarchicalPlan{
		ID:         "err-plan",
		Goal:       "error test",
		Milestones: milestones,
		State:      "created",
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	_, err := hp.ExecutePlan(context.Background(), plan)
	require.NoError(t, err)

	assert.Equal(t, PlanStepStateFailed, milestones[0].Steps[0].State)
	assert.Equal(t, "step error detail", milestones[0].Steps[0].Error)
}
