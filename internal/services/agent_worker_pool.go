package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// AgentWorkerPool manages concurrent AgentWorker instances and orchestrates
// dependency-graph execution by dispatching tasks in layers. All tasks within
// a single layer run in parallel (bounded by the semaphore), while layers
// themselves are executed sequentially so that later layers can depend on the
// results of earlier ones.
type AgentWorkerPool struct {
	sem    chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	logger *logrus.Logger
}

// NewAgentWorkerPool creates a pool that allows at most maxConcurrent agents
// to run simultaneously. maxConcurrent defaults to 5 when <= 0.
func NewAgentWorkerPool(
	maxConcurrent int,
	logger *logrus.Logger,
) *AgentWorkerPool {
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}
	if logger == nil {
		logger = logrus.New()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &AgentWorkerPool{
		sem:    make(chan struct{}, maxConcurrent),
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
}

// DispatchTasks executes task layers sequentially. Within each layer every
// task is dispatched to its own AgentWorker goroutine, bounded by the pool's
// semaphore. Results are streamed to the returned channel as they complete.
// The channel is closed once all layers have finished or the context is
// cancelled.
//
// taskLayers is a slice of layers where each layer is a slice of tasks that
// may run concurrently. Layer N+1 is not started until all tasks in layer N
// have completed.
//
// completeFunc is the LLM completion function forwarded to every AgentWorker.
//
// toolExecutor may be nil when tool execution is not required.
func (p *AgentWorkerPool) DispatchTasks(
	ctx context.Context,
	taskLayers [][]AgenticTask,
	completeFunc CompleteFunc,
	toolExecutor *IterativeToolExecutor,
	maxIterationsPerAgent int,
) (<-chan AgenticResult, error) {
	if len(taskLayers) == 0 {
		ch := make(chan AgenticResult)
		close(ch)
		return ch, nil
	}

	results := make(chan AgenticResult, p.countTasks(taskLayers))

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer close(results)

		mergedCtx, mergedCancel := context.WithCancel(ctx)
		defer mergedCancel()

		// Also respect the pool's own context.
		go func() {
			select {
			case <-p.ctx.Done():
				mergedCancel()
			case <-mergedCtx.Done():
			}
		}()

		for layerIdx, layer := range taskLayers {
			select {
			case <-mergedCtx.Done():
				return
			default:
			}

			p.logger.WithFields(logrus.Fields{
				"layer":      layerIdx,
				"task_count": len(layer),
			}).Debug("Dispatching task layer")

			var layerWg sync.WaitGroup

			for taskIdx, task := range layer {
				select {
				case <-mergedCtx.Done():
					return
				default:
				}

				agentID := fmt.Sprintf(
					"agent-L%d-T%d-%s", layerIdx, taskIdx, task.ID,
				)

				worker := NewAgentWorker(
					agentID,
					task,
					toolExecutor,
					maxIterationsPerAgent,
					p.logger,
				)

				layerWg.Add(1)

				// Acquire semaphore slot before launching goroutine.
				select {
				case p.sem <- struct{}{}:
				case <-mergedCtx.Done():
					layerWg.Done()
					return
				}

				go func(w *AgentWorker) {
					defer layerWg.Done()
					defer func() { <-p.sem }()

					result := w.Execute(mergedCtx, completeFunc)

					select {
					case results <- result:
					case <-mergedCtx.Done():
					}
				}(worker)
			}

			// Wait for all tasks in this layer before moving to the next.
			layerWg.Wait()

			p.logger.WithFields(logrus.Fields{
				"layer": layerIdx,
			}).Debug("Task layer completed")
		}
	}()

	return results, nil
}

// Shutdown cancels the pool's internal context and waits for all dispatched
// goroutines to finish. It is safe to call multiple times.
func (p *AgentWorkerPool) Shutdown() {
	p.cancel()
	p.wg.Wait()
}

// countTasks returns the total number of tasks across all layers.
func (p *AgentWorkerPool) countTasks(layers [][]AgenticTask) int {
	total := 0
	for _, layer := range layers {
		total += len(layer)
	}
	return total
}
