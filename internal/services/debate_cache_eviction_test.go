package services

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDebateService_IntentCache_BoundedSize(t *testing.T) {
	ds := &DebateService{
		logger:      logrus.New(),
		intentCache: make(map[string]*IntentClassificationResult),
	}

	// Fill beyond max
	ds.mu.Lock()
	for i := 0; i < maxIntentCacheSize+500; i++ {
		ds.intentCache[fmt.Sprintf("topic-%d", i)] = &IntentClassificationResult{
			Intent: "test",
		}
	}
	ds.mu.Unlock()

	// Trigger eviction
	ds.evictIntentCacheIfNeeded()

	ds.mu.Lock()
	size := len(ds.intentCache)
	ds.mu.Unlock()

	assert.LessOrEqual(t, size, maxIntentCacheSize,
		"cache should be bounded to maxIntentCacheSize after eviction")
}

func TestDebateService_IntentCache_NoEvictionUnderLimit(t *testing.T) {
	ds := &DebateService{
		logger:      logrus.New(),
		intentCache: make(map[string]*IntentClassificationResult),
	}

	ds.mu.Lock()
	for i := 0; i < 100; i++ {
		ds.intentCache[fmt.Sprintf("topic-%d", i)] = &IntentClassificationResult{
			Intent: "test",
		}
	}
	ds.mu.Unlock()

	ds.evictIntentCacheIfNeeded()

	ds.mu.Lock()
	size := len(ds.intentCache)
	ds.mu.Unlock()

	assert.Equal(t, 100, size, "cache under limit should not be evicted")
}
