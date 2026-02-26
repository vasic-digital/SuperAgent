package inmemory

import (
	"container/heap"
	"sync"

	"dev.helix.agent/internal/messaging"
)

// Queue is an in-memory priority queue implementation.
type Queue struct {
	name     string
	items    priorityQueue
	capacity int
	mu       sync.RWMutex
}

// NewQueue creates a new in-memory queue.
func NewQueue(name string, capacity int) *Queue {
	pq := make(priorityQueue, 0, capacity)
	heap.Init(&pq)
	return &Queue{
		name:     name,
		items:    pq,
		capacity: capacity,
	}
}

// Enqueue adds a message to the queue.
func (q *Queue) Enqueue(msg *messaging.Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.items.Len() >= q.capacity {
		return messaging.NewBrokerError(messaging.ErrCodePublishFailed, "queue capacity exceeded", nil)
	}

	item := &queueItem{
		message:  msg,
		priority: int(msg.Priority),
	}
	heap.Push(&q.items, item)
	return nil
}

// Dequeue removes and returns the highest priority message.
func (q *Queue) Dequeue() (*messaging.Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.items.Len() == 0 {
		return nil, nil
	}

	item := heap.Pop(&q.items).(*queueItem) //nolint:errcheck
	return item.message, nil
}

// Peek returns the highest priority message without removing it.
func (q *Queue) Peek() (*messaging.Message, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.items.Len() == 0 {
		return nil, nil
	}

	return q.items[0].message, nil
}

// Len returns the number of messages in the queue.
func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.items.Len()
}

// Clear removes all messages from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = make(priorityQueue, 0, q.capacity)
	heap.Init(&q.items)
}

// Name returns the queue name.
func (q *Queue) Name() string {
	return q.name
}

// Capacity returns the queue capacity.
func (q *Queue) Capacity() int {
	return q.capacity
}

// IsFull returns true if the queue is at capacity.
func (q *Queue) IsFull() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.items.Len() >= q.capacity
}

// IsEmpty returns true if the queue is empty.
func (q *Queue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.items.Len() == 0
}

// queueItem represents an item in the priority queue.
type queueItem struct {
	message  *messaging.Message
	priority int
	index    int
}

// priorityQueue implements heap.Interface for priority queue.
type priorityQueue []*queueItem

// Len returns the number of items in the queue.
func (pq priorityQueue) Len() int { return len(pq) }

// Less returns true if item i has higher priority than item j.
func (pq priorityQueue) Less(i, j int) bool {
	// Higher priority value = higher priority
	return pq[i].priority > pq[j].priority
}

// Swap swaps items i and j.
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an item to the queue.
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*queueItem)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes and returns the highest priority item.
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// SimpleQueue is a simple FIFO queue without priority.
type SimpleQueue struct {
	name     string
	items    []*messaging.Message
	capacity int
	head     int
	tail     int
	count    int
	mu       sync.RWMutex
}

// NewSimpleQueue creates a new simple FIFO queue.
func NewSimpleQueue(name string, capacity int) *SimpleQueue {
	return &SimpleQueue{
		name:     name,
		items:    make([]*messaging.Message, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		count:    0,
	}
}

// Enqueue adds a message to the queue.
func (q *SimpleQueue) Enqueue(msg *messaging.Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count >= q.capacity {
		return messaging.NewBrokerError(messaging.ErrCodePublishFailed, "queue capacity exceeded", nil)
	}

	q.items[q.tail] = msg
	q.tail = (q.tail + 1) % q.capacity
	q.count++
	return nil
}

// Dequeue removes and returns the oldest message.
func (q *SimpleQueue) Dequeue() (*messaging.Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count == 0 {
		return nil, nil
	}

	msg := q.items[q.head]
	q.items[q.head] = nil // avoid memory leak
	q.head = (q.head + 1) % q.capacity
	q.count--
	return msg, nil
}

// Peek returns the oldest message without removing it.
func (q *SimpleQueue) Peek() (*messaging.Message, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.count == 0 {
		return nil, nil
	}

	return q.items[q.head], nil
}

// Len returns the number of messages in the queue.
func (q *SimpleQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count
}

// Clear removes all messages from the queue.
func (q *SimpleQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = make([]*messaging.Message, q.capacity)
	q.head = 0
	q.tail = 0
	q.count = 0
}

// Name returns the queue name.
func (q *SimpleQueue) Name() string {
	return q.name
}

// IsFull returns true if the queue is at capacity.
func (q *SimpleQueue) IsFull() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count >= q.capacity
}

// IsEmpty returns true if the queue is empty.
func (q *SimpleQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count == 0
}

// DelayedQueue is a queue that delays message delivery.
type DelayedQueue struct {
	*Queue
	delayedItems []*delayedItem
	delayMu      sync.RWMutex
}

type delayedItem struct {
	message   *messaging.Message
	deliverAt int64 // Unix nanoseconds
}

// NewDelayedQueue creates a new delayed queue.
func NewDelayedQueue(name string, capacity int) *DelayedQueue {
	return &DelayedQueue{
		Queue:        NewQueue(name, capacity),
		delayedItems: make([]*delayedItem, 0),
	}
}

// EnqueueDelayed adds a message with a delivery delay.
func (q *DelayedQueue) EnqueueDelayed(msg *messaging.Message, delayNs int64) error {
	q.delayMu.Lock()
	defer q.delayMu.Unlock()

	q.delayedItems = append(q.delayedItems, &delayedItem{
		message:   msg,
		deliverAt: delayNs,
	})
	return nil
}

// ProcessDelayed moves ready delayed messages to the main queue.
func (q *DelayedQueue) ProcessDelayed(nowNs int64) int {
	q.delayMu.Lock()
	defer q.delayMu.Unlock()

	count := 0
	remaining := make([]*delayedItem, 0, len(q.delayedItems))

	for _, item := range q.delayedItems {
		if item.deliverAt <= nowNs {
			if err := q.Enqueue(item.message); err == nil {
				count++
			}
		} else {
			remaining = append(remaining, item)
		}
	}

	q.delayedItems = remaining
	return count
}

// DelayedCount returns the number of delayed messages.
func (q *DelayedQueue) DelayedCount() int {
	q.delayMu.RLock()
	defer q.delayMu.RUnlock()
	return len(q.delayedItems)
}
