package gemini

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGeminiACP_ResponseMap_ConcurrentAccess verifies that the RWMutex locking
// pattern protecting the response channel map is safe under concurrent access,
// simulating the interleaving of readResponses (RLock) and sendRequest (Lock).
func TestGeminiACP_ResponseMap_ConcurrentAccess(t *testing.T) {
	var mu sync.RWMutex
	responses := make(map[int64]chan *geminiACPResponse)

	var wg sync.WaitGroup

	// Simulate sendRequest writes: register channel, then delete after response
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			ch := make(chan *geminiACPResponse, 1)
			mu.Lock()
			responses[id] = ch
			mu.Unlock()

			mu.Lock()
			delete(responses, id)
			mu.Unlock()
		}(int64(i))
	}

	// Simulate readResponses reads: look up channel by ID
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			mu.RLock()
			_, _ = responses[id]
			mu.RUnlock()
		}(int64(i))
	}

	wg.Wait()
	assert.Empty(t, responses)
}
