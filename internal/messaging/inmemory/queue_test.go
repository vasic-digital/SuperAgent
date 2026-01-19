package inmemory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/messaging"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue("test-queue", 100)
	assert.NotNil(t, q)
	assert.Equal(t, "test-queue", q.Name())
	assert.Equal(t, 100, q.Capacity())
	assert.True(t, q.IsEmpty())
	assert.False(t, q.IsFull())
}

func TestQueue_Enqueue(t *testing.T) {
	q := NewQueue("test-queue", 10)

	msg := messaging.NewMessage("test.type", []byte("payload"))
	err := q.Enqueue(msg)
	require.NoError(t, err)
	assert.Equal(t, 1, q.Len())
	assert.False(t, q.IsEmpty())
}

func TestQueue_EnqueueCapacityExceeded(t *testing.T) {
	q := NewQueue("test-queue", 2)

	// Fill the queue
	q.Enqueue(messaging.NewMessage("type", []byte("1")))
	q.Enqueue(messaging.NewMessage("type", []byte("2")))
	assert.True(t, q.IsFull())

	// This should fail
	err := q.Enqueue(messaging.NewMessage("type", []byte("3")))
	assert.Error(t, err)
}

func TestQueue_Dequeue(t *testing.T) {
	q := NewQueue("test-queue", 10)

	// Add messages with different priorities
	msg1 := messaging.NewMessage("type", []byte("low"))
	msg1.Priority = messaging.PriorityLow

	msg2 := messaging.NewMessage("type", []byte("high"))
	msg2.Priority = messaging.PriorityHigh

	msg3 := messaging.NewMessage("type", []byte("normal"))
	msg3.Priority = messaging.PriorityNormal

	q.Enqueue(msg1)
	q.Enqueue(msg2)
	q.Enqueue(msg3)

	// Should dequeue highest priority first
	dequeued, err := q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, []byte("high"), dequeued.Payload)

	// Then normal
	dequeued, err = q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, []byte("normal"), dequeued.Payload)

	// Then low
	dequeued, err = q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, []byte("low"), dequeued.Payload)
}

func TestQueue_DequeueEmpty(t *testing.T) {
	q := NewQueue("test-queue", 10)

	msg, err := q.Dequeue()
	require.NoError(t, err)
	assert.Nil(t, msg)
}

func TestQueue_Peek(t *testing.T) {
	q := NewQueue("test-queue", 10)

	// Peek empty queue
	msg, err := q.Peek()
	require.NoError(t, err)
	assert.Nil(t, msg)

	// Add message
	testMsg := messaging.NewMessage("type", []byte("peek test"))
	q.Enqueue(testMsg)

	// Peek should return without removing
	peeked, err := q.Peek()
	require.NoError(t, err)
	assert.Equal(t, []byte("peek test"), peeked.Payload)
	assert.Equal(t, 1, q.Len()) // Still in queue
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue("test-queue", 10)

	q.Enqueue(messaging.NewMessage("type", []byte("1")))
	q.Enqueue(messaging.NewMessage("type", []byte("2")))
	assert.Equal(t, 2, q.Len())

	q.Clear()
	assert.Equal(t, 0, q.Len())
	assert.True(t, q.IsEmpty())
}

// SimpleQueue tests

func TestNewSimpleQueue(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 100)
	assert.NotNil(t, q)
	assert.Equal(t, "simple-queue", q.Name())
	assert.True(t, q.IsEmpty())
	assert.False(t, q.IsFull())
}

func TestSimpleQueue_Enqueue(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 10)

	msg := messaging.NewMessage("test.type", []byte("payload"))
	err := q.Enqueue(msg)
	require.NoError(t, err)
	assert.Equal(t, 1, q.Len())
}

func TestSimpleQueue_EnqueueCapacityExceeded(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 2)

	q.Enqueue(messaging.NewMessage("type", []byte("1")))
	q.Enqueue(messaging.NewMessage("type", []byte("2")))
	assert.True(t, q.IsFull())

	err := q.Enqueue(messaging.NewMessage("type", []byte("3")))
	assert.Error(t, err)
}

func TestSimpleQueue_Dequeue(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 10)

	// FIFO order
	q.Enqueue(messaging.NewMessage("type", []byte("first")))
	q.Enqueue(messaging.NewMessage("type", []byte("second")))
	q.Enqueue(messaging.NewMessage("type", []byte("third")))

	// Should dequeue in FIFO order
	dequeued, err := q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, []byte("first"), dequeued.Payload)

	dequeued, err = q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), dequeued.Payload)

	dequeued, err = q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, []byte("third"), dequeued.Payload)
}

func TestSimpleQueue_DequeueEmpty(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 10)

	msg, err := q.Dequeue()
	require.NoError(t, err)
	assert.Nil(t, msg)
}

func TestSimpleQueue_Peek(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 10)

	// Peek empty
	msg, err := q.Peek()
	require.NoError(t, err)
	assert.Nil(t, msg)

	// Add and peek
	q.Enqueue(messaging.NewMessage("type", []byte("peek")))
	peeked, err := q.Peek()
	require.NoError(t, err)
	assert.Equal(t, []byte("peek"), peeked.Payload)
	assert.Equal(t, 1, q.Len())
}

func TestSimpleQueue_Clear(t *testing.T) {
	q := NewSimpleQueue("simple-queue", 10)

	q.Enqueue(messaging.NewMessage("type", []byte("1")))
	q.Enqueue(messaging.NewMessage("type", []byte("2")))

	q.Clear()
	assert.Equal(t, 0, q.Len())
	assert.True(t, q.IsEmpty())
}

func TestSimpleQueue_CircularBehavior(t *testing.T) {
	q := NewSimpleQueue("circular-queue", 3)

	// Fill and partially empty to test circular behavior
	q.Enqueue(messaging.NewMessage("type", []byte("1")))
	q.Enqueue(messaging.NewMessage("type", []byte("2")))
	q.Dequeue() // Remove "1"

	q.Enqueue(messaging.NewMessage("type", []byte("3")))
	q.Enqueue(messaging.NewMessage("type", []byte("4")))

	// Should have 2, 3, 4
	msg, _ := q.Dequeue()
	assert.Equal(t, []byte("2"), msg.Payload)
	msg, _ = q.Dequeue()
	assert.Equal(t, []byte("3"), msg.Payload)
	msg, _ = q.Dequeue()
	assert.Equal(t, []byte("4"), msg.Payload)
}

// DelayedQueue tests

func TestNewDelayedQueue(t *testing.T) {
	q := NewDelayedQueue("delayed-queue", 100)
	assert.NotNil(t, q)
	assert.Equal(t, "delayed-queue", q.Name())
	assert.Equal(t, 0, q.DelayedCount())
}

func TestDelayedQueue_EnqueueDelayed(t *testing.T) {
	q := NewDelayedQueue("delayed-queue", 100)

	msg := messaging.NewMessage("type", []byte("delayed"))
	deliverAt := time.Now().Add(time.Hour).UnixNano()

	err := q.EnqueueDelayed(msg, deliverAt)
	require.NoError(t, err)
	assert.Equal(t, 1, q.DelayedCount())
	assert.Equal(t, 0, q.Len()) // Not yet in main queue
}

func TestDelayedQueue_ProcessDelayed(t *testing.T) {
	q := NewDelayedQueue("delayed-queue", 100)

	// Add messages with different delivery times
	now := time.Now().UnixNano()

	msg1 := messaging.NewMessage("type", []byte("past"))
	q.EnqueueDelayed(msg1, now-1000) // Past

	msg2 := messaging.NewMessage("type", []byte("future"))
	q.EnqueueDelayed(msg2, now+1000000000) // Future

	msg3 := messaging.NewMessage("type", []byte("now"))
	q.EnqueueDelayed(msg3, now) // Now

	// Process delayed messages
	count := q.ProcessDelayed(now)

	assert.Equal(t, 2, count) // past and now should be moved
	assert.Equal(t, 1, q.DelayedCount()) // future still delayed
	assert.Equal(t, 2, q.Len()) // Two in main queue
}

// Priority queue interface tests

func TestPriorityQueue_LessSwap(t *testing.T) {
	pq := make(priorityQueue, 2)
	pq[0] = &queueItem{priority: 1, index: 0}
	pq[1] = &queueItem{priority: 2, index: 1}

	// Item with higher priority (2) should be "less" (comes first)
	assert.True(t, pq.Less(1, 0))
	assert.False(t, pq.Less(0, 1))

	// Test swap
	pq.Swap(0, 1)
	assert.Equal(t, 2, pq[0].priority)
	assert.Equal(t, 1, pq[1].priority)
	assert.Equal(t, 0, pq[0].index)
	assert.Equal(t, 1, pq[1].index)
}

func TestPriorityQueue_PushPop(t *testing.T) {
	pq := make(priorityQueue, 0)

	// Push items
	item1 := &queueItem{priority: 5}
	item2 := &queueItem{priority: 10}

	pq.Push(item1)
	pq.Push(item2)

	assert.Equal(t, 2, pq.Len())

	// Pop
	popped := pq.Pop().(*queueItem)
	assert.NotNil(t, popped)
	assert.Equal(t, -1, popped.index) // Safety check
}
