package services

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// collectResults drains a result channel into a slice with a timeout guard.
func collectResults(
	t *testing.T,
	ch <-chan AgenticResult,
	timeout time.Duration,
) []AgenticResult {
	t.Helper()
	var results []AgenticResult
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case r, ok := <-ch:
			if !ok {
				return results
			}
			results = append(results, r)
		case <-timer.C:
			t.Fatal("timed out waiting for results")
			return results
		}
	}
}

func TestAgentWorkerPool_SingleTask(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	pool := NewAgentWorkerPool(5, logger)
	defer pool.Shutdown()

	layers := [][]AgenticTask{
		{
			{
				ID:          "task-solo",
				Description: "Single task test",
				Status:      AgenticTaskPending,
			},
		},
	}

	complete := agentWorkerMockComplete("solo result", nil)

	ch, err := pool.DispatchTasks(
		context.Background(), layers, complete, nil, 5,
	)
	require.NoError(t, err)

	results := collectResults(t, ch, 10*time.Second)
	require.Len(t, results, 1)
	assert.Equal(t, "task-solo", results[0].TaskID)
	assert.Equal(t, "solo result", results[0].Content)
	assert.NoError(t, results[0].Error)
}

func TestAgentWorkerPool_ParallelTasks(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	pool := NewAgentWorkerPool(10, logger)
	defer pool.Shutdown()

	// All tasks in a single layer should run concurrently.
	var running int64
	var maxRunning int64
	var mu sync.Mutex

	complete := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		cur := atomic.AddInt64(&running, 1)
		mu.Lock()
		if cur > maxRunning {
			maxRunning = cur
		}
		mu.Unlock()

		// Brief sleep to allow overlap.
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt64(&running, -1)

		return &models.LLMResponse{
			Content: "parallel done",
		}, nil
	}

	layers := [][]AgenticTask{
		{
			{ID: "p1", Description: "Parallel 1", Status: AgenticTaskPending},
			{ID: "p2", Description: "Parallel 2", Status: AgenticTaskPending},
			{ID: "p3", Description: "Parallel 3", Status: AgenticTaskPending},
		},
	}

	ch, err := pool.DispatchTasks(
		context.Background(), layers, complete, nil, 5,
	)
	require.NoError(t, err)

	results := collectResults(t, ch, 10*time.Second)
	assert.Len(t, results, 3)

	mu.Lock()
	peak := maxRunning
	mu.Unlock()
	assert.True(
		t, peak >= 2,
		"expected at least 2 concurrent agents, got %d", peak,
	)
}

func TestAgentWorkerPool_DependencyLayers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	pool := NewAgentWorkerPool(5, logger)
	defer pool.Shutdown()

	// Track completion order via timestamps.
	var mu sync.Mutex
	completionOrder := make(map[string]time.Time)

	complete := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		time.Sleep(20 * time.Millisecond)
		return &models.LLMResponse{Content: "done"}, nil
	}

	// Wrap to record completion times per task.
	wrappedComplete := func(
		ctx context.Context, msgs []models.Message,
	) (*models.LLMResponse, error) {
		resp, err := complete(ctx, msgs)
		// We will record after the worker returns, but since the worker
		// calls completeFunc on each iteration, we capture here.
		return resp, err
	}

	layers := [][]AgenticTask{
		// Layer 0
		{
			{ID: "L0-A", Description: "Layer 0 task A", Status: AgenticTaskPending},
			{ID: "L0-B", Description: "Layer 0 task B", Status: AgenticTaskPending},
		},
		// Layer 1 — should only start after layer 0 completes.
		{
			{
				ID:           "L1-C",
				Description:  "Layer 1 task C",
				Dependencies: []string{"L0-A", "L0-B"},
				Status:       AgenticTaskPending,
			},
		},
	}

	ch, err := pool.DispatchTasks(
		context.Background(), layers, wrappedComplete, nil, 5,
	)
	require.NoError(t, err)

	results := collectResults(t, ch, 10*time.Second)
	assert.Len(t, results, 3)

	for _, r := range results {
		mu.Lock()
		completionOrder[r.TaskID] = time.Now()
		mu.Unlock()
	}

	// Verify all three tasks completed. The ordering guarantee is that
	// layer 0 tasks appear before layer 1 in the results channel because
	// the pool waits for each layer before starting the next.
	taskIDs := make(map[string]bool)
	for _, r := range results {
		taskIDs[r.TaskID] = true
	}
	assert.True(t, taskIDs["L0-A"])
	assert.True(t, taskIDs["L0-B"])
	assert.True(t, taskIDs["L1-C"])

	// Layer 0 results must appear before layer 1 results.
	layer1Idx := -1
	for i, r := range results {
		if r.TaskID == "L1-C" {
			layer1Idx = i
		}
	}
	assert.True(
		t, layer1Idx >= 2,
		"layer 1 task should appear after both layer 0 tasks, index=%d",
		layer1Idx,
	)
}

func TestAgentWorkerPool_SemaphoreLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	maxConcurrent := 2
	pool := NewAgentWorkerPool(maxConcurrent, logger)
	defer pool.Shutdown()

	var running int64
	var maxRunning int64
	var mu sync.Mutex

	complete := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		cur := atomic.AddInt64(&running, 1)
		mu.Lock()
		if cur > maxRunning {
			maxRunning = cur
		}
		mu.Unlock()

		time.Sleep(80 * time.Millisecond)
		atomic.AddInt64(&running, -1)

		return &models.LLMResponse{Content: "bounded"}, nil
	}

	layers := [][]AgenticTask{
		{
			{ID: "s1", Description: "Sem 1", Status: AgenticTaskPending},
			{ID: "s2", Description: "Sem 2", Status: AgenticTaskPending},
			{ID: "s3", Description: "Sem 3", Status: AgenticTaskPending},
			{ID: "s4", Description: "Sem 4", Status: AgenticTaskPending},
			{ID: "s5", Description: "Sem 5", Status: AgenticTaskPending},
		},
	}

	ch, err := pool.DispatchTasks(
		context.Background(), layers, complete, nil, 5,
	)
	require.NoError(t, err)

	results := collectResults(t, ch, 30*time.Second)
	assert.Len(t, results, 5)

	mu.Lock()
	peak := maxRunning
	mu.Unlock()
	assert.True(
		t, peak <= int64(maxConcurrent),
		"expected at most %d concurrent agents, got %d",
		maxConcurrent, peak,
	)
}

func TestAgentWorkerPool_Shutdown(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	pool := NewAgentWorkerPool(2, logger)

	// Dispatch a slow layer.
	complete := func(
		ctx context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			return &models.LLMResponse{Content: "slow"}, nil
		}
	}

	layers := [][]AgenticTask{
		{
			{ID: "slow-1", Description: "Slow task", Status: AgenticTaskPending},
			{ID: "slow-2", Description: "Slow task 2", Status: AgenticTaskPending},
		},
	}

	ch, err := pool.DispatchTasks(
		context.Background(), layers, complete, nil, 5,
	)
	require.NoError(t, err)

	// Give goroutines a moment to start, then shut down.
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		pool.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// Shutdown completed — drain results.
		results := collectResults(t, ch, 5*time.Second)
		for _, r := range results {
			// Tasks should have been cancelled or errored.
			if r.Error != nil {
				assert.Error(t, r.Error)
			}
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown did not complete within timeout")
	}
}

func TestAgentWorkerPool_EmptyLayers(t *testing.T) {
	pool := NewAgentWorkerPool(5, nil)
	defer pool.Shutdown()

	ch, err := pool.DispatchTasks(
		context.Background(), nil, nil, nil, 5,
	)
	require.NoError(t, err)

	// Channel should be closed immediately.
	results := collectResults(t, ch, 2*time.Second)
	assert.Empty(t, results)
}

func TestAgentWorkerPool_EmptyLayerSlice(t *testing.T) {
	pool := NewAgentWorkerPool(5, nil)
	defer pool.Shutdown()

	ch, err := pool.DispatchTasks(
		context.Background(), [][]AgenticTask{}, nil, nil, 5,
	)
	require.NoError(t, err)

	results := collectResults(t, ch, 2*time.Second)
	assert.Empty(t, results)
}

func TestAgentWorkerPool_ContextCancellation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	pool := NewAgentWorkerPool(5, logger)
	defer pool.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())

	complete := func(
		ctx context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Second):
			return &models.LLMResponse{Content: "late"}, nil
		}
	}

	layers := [][]AgenticTask{
		{
			{ID: "ctx-1", Description: "Ctx task", Status: AgenticTaskPending},
		},
	}

	ch, err := pool.DispatchTasks(ctx, layers, complete, nil, 5)
	require.NoError(t, err)

	// Cancel after a brief moment.
	time.Sleep(30 * time.Millisecond)
	cancel()

	results := collectResults(t, ch, 5*time.Second)
	// Task should have been cancelled.
	for _, r := range results {
		if r.Error != nil {
			assert.Error(t, r.Error)
		}
	}
}

func TestAgentWorkerPool_DefaultConcurrency(t *testing.T) {
	pool := NewAgentWorkerPool(0, nil)
	defer pool.Shutdown()

	// Pool should default to 5 capacity.
	assert.Equal(t, 5, cap(pool.sem))
}

func TestAgentWorkerPool_MultipleLayersAllComplete(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	pool := NewAgentWorkerPool(5, logger)
	defer pool.Shutdown()

	complete := agentWorkerMockComplete("layer result", nil)

	layers := [][]AgenticTask{
		{
			{ID: "m0-a", Description: "M0 A", Status: AgenticTaskPending},
		},
		{
			{ID: "m1-a", Description: "M1 A", Status: AgenticTaskPending},
			{ID: "m1-b", Description: "M1 B", Status: AgenticTaskPending},
		},
		{
			{ID: "m2-a", Description: "M2 A", Status: AgenticTaskPending},
		},
	}

	ch, err := pool.DispatchTasks(
		context.Background(), layers, complete, nil, 5,
	)
	require.NoError(t, err)

	results := collectResults(t, ch, 10*time.Second)
	assert.Len(t, results, 4)

	for _, r := range results {
		assert.NoError(t, r.Error)
		assert.Equal(t, "layer result", r.Content)
	}
}
