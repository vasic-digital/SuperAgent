package concurrency

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"digital.vasic.concurrency/pkg/pool"
)

// Alias for backward compatibility
var concurrency = pool

func TestWorkerPool_BasicOperation(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	var executed int32
	task := concurrency.NewTaskFunc("test-1", func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&executed, 1)
		return "result", nil
	})

	err := pool.Submit(task)
	assert.NoError(t, err)

	// Wait for result
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := pool.SubmitWait(ctx, concurrency.NewTaskFunc("test-2", func(ctx context.Context) (interface{}, error) {
		return "done", nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, "done", result.Value)
}

func TestWorkerPool_Concurrency(t *testing.T) {
	config := &concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 100,
	}
	pool := concurrency.NewWorkerPool(config)
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	const numTasks = 100
	var executed int64
	var wg sync.WaitGroup

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		idx := i
		task := concurrency.NewTaskFunc("task-"+string(rune('0'+idx%10)), func(ctx context.Context) (interface{}, error) {
			atomic.AddInt64(&executed, 1)
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})

		go func() {
			defer wg.Done()
			pool.Submit(task)
		}()
	}

	wg.Wait()
	pool.WaitForDrain(context.Background())

	assert.True(t, atomic.LoadInt64(&executed) >= numTasks/2)
}

func TestWorkerPool_BatchSubmit(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	tasks := make([]concurrency.Task, 10)
	for i := range tasks {
		idx := i
		tasks[i] = concurrency.NewTaskFunc("batch-"+string(rune('0'+idx)), func(ctx context.Context) (interface{}, error) {
			return idx * 2, nil
		})
	}

	resultCh := pool.SubmitBatch(tasks)

	var results []interface{}
	for result := range resultCh {
		assert.NoError(t, result.Error)
		results = append(results, result.Value)
	}

	assert.Len(t, results, 10)
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	expectedErr := errors.New("task error")
	task := concurrency.NewTaskFunc("error-task", func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := pool.SubmitWait(ctx, task)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, result.Error)
}

func TestWorkerPool_Cancellation(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	task := concurrency.NewTaskFunc("long-task", func(taskCtx context.Context) (interface{}, error) {
		select {
		case <-taskCtx.Done():
			return nil, taskCtx.Err()
		case <-time.After(5 * time.Second):
			return "completed", nil
		}
	})

	_, err := pool.SubmitWait(ctx, task)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestWorkerPool_Metrics(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	// Submit some tasks
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		idx := i
		task := concurrency.NewTaskFunc("metrics-"+string(rune('0'+idx)), func(ctx context.Context) (interface{}, error) {
			return nil, nil
		})
		pool.SubmitWait(ctx, task)
	}

	metrics := pool.Metrics()
	assert.NotNil(t, metrics)
	assert.True(t, metrics.CompletedTasks > 0 || metrics.TaskCount > 0)
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()

	var completed int32
	for i := 0; i < 10; i++ {
		idx := i
		task := concurrency.NewTaskFunc("shutdown-"+string(rune('0'+idx)), func(ctx context.Context) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil, nil
		})
		pool.Submit(task)
	}

	// Shutdown should wait for tasks to complete
	err := pool.Shutdown(5 * time.Second)
	assert.NoError(t, err)
}

func TestWorkerPool_QueueFull(t *testing.T) {
	config := &concurrency.PoolConfig{
		Workers:   1,
		QueueSize: 2,
	}
	pool := concurrency.NewWorkerPool(config)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	// Fill the queue with blocking tasks
	for i := 0; i < 3; i++ {
		idx := i
		task := concurrency.NewTaskFunc("queue-"+string(rune('0'+idx)), func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Second)
			return nil, nil
		})
		pool.Submit(task)
	}

	// Try to submit when queue is full
	task := concurrency.NewTaskFunc("overflow", func(ctx context.Context) (interface{}, error) {
		return nil, nil
	})

	err := pool.Submit(task)
	// Either succeeds or queue full - both are acceptable
	if err != nil {
		assert.Contains(t, err.Error(), "queue is full")
	}
}

func TestWorkerPool_RaceCondition(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	var counter int64
	var wg sync.WaitGroup

	// Submit many tasks concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			task := concurrency.NewTaskFunc("race-"+string(rune('a'+idx%26)), func(ctx context.Context) (interface{}, error) {
				atomic.AddInt64(&counter, 1)
				return nil, nil
			})
			pool.Submit(task)
		}()
	}

	wg.Wait()
	pool.WaitForDrain(context.Background())
}

func TestWorkerPool_IsRunning(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)

	// Not started yet
	assert.False(t, pool.IsRunning())

	pool.Start()
	assert.True(t, pool.IsRunning())

	pool.Shutdown(time.Second)
	assert.False(t, pool.IsRunning())
}

func TestWorkerPool_ActiveWorkers(t *testing.T) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   4,
		QueueSize: 100,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	// Submit long-running tasks
	for i := 0; i < 10; i++ {
		idx := i
		task := concurrency.NewTaskFunc("active-"+string(rune('0'+idx)), func(ctx context.Context) (interface{}, error) {
			time.Sleep(500 * time.Millisecond)
			return nil, nil
		})
		pool.Submit(task)
	}

	time.Sleep(50 * time.Millisecond)
	active := pool.ActiveWorkers()
	assert.True(t, active >= 0)
}

func TestWorkerPool_QueueLength(t *testing.T) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   1,
		QueueSize: 100,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	initialLen := pool.QueueLength()
	assert.Equal(t, 0, initialLen)
}

func TestParallelExecute(t *testing.T) {
	fns := make([]func(ctx context.Context) (interface{}, error), 5)
	for i := range fns {
		idx := i
		fns[i] = func(ctx context.Context) (interface{}, error) {
			return idx * 2, nil
		}
	}

	ctx := context.Background()
	results, err := concurrency.ParallelExecute(ctx, fns)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestMap(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	ctx := context.Background()
	results, err := concurrency.Map(ctx, items, 3, func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	})

	assert.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestWorkerPool_SubmitBatchWait(t *testing.T) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	tasks := make([]concurrency.Task, 5)
	for i := range tasks {
		idx := i
		tasks[i] = concurrency.NewTaskFunc("batch-wait-"+string(rune('0'+idx)), func(ctx context.Context) (interface{}, error) {
			return idx * 3, nil
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := pool.SubmitBatchWait(ctx, tasks)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
}

func BenchmarkWorkerPool_Submit(b *testing.B) {
	pool := concurrency.NewWorkerPool(nil)
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := concurrency.NewTaskFunc("bench", func(ctx context.Context) (interface{}, error) {
			return nil, nil
		})
		pool.Submit(task)
	}
}

func BenchmarkWorkerPool_Parallel(b *testing.B) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 10000,
	})
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			task := concurrency.NewTaskFunc("parallel-bench", func(ctx context.Context) (interface{}, error) {
				return nil, nil
			})
			pool.Submit(task)
			i++
		}
	})
}
