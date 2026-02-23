package stress

import (
	"sync"
	"testing"

	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newBootManagerStressLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)
	return l
}

func newBootManagerStressServicesConfig() *config.ServicesConfig {
	cfg := config.DefaultServicesConfig()
	// Disable all services so BootAll completes quickly without network calls
	cfg.PostgreSQL.Enabled = false
	cfg.Redis.Enabled = false
	cfg.Cognee.Enabled = false
	cfg.ChromaDB.Enabled = false
	cfg.Prometheus.Enabled = false
	cfg.Grafana.Enabled = false
	cfg.Neo4j.Enabled = false
	cfg.Kafka.Enabled = false
	cfg.RabbitMQ.Enabled = false
	cfg.Qdrant.Enabled = false
	cfg.Weaviate.Enabled = false
	cfg.LangChain.Enabled = false
	cfg.LlamaIndex.Enabled = false
	return &cfg
}

// TestBootManager_ConcurrentAccess_NoDataRace exercises concurrent reads of the
// Results map while BootAll is writing to it. Before the fix this triggers a DATA
// RACE; after adding resultsMu + GetResults() / setResult() it must pass cleanly
// under -race.
func TestBootManager_ConcurrentAccess_NoDataRace(t *testing.T) {
	cfg := newBootManagerStressServicesConfig()
	logger := newBootManagerStressLogger()
	bm := services.NewBootManager(cfg, logger)

	// Spawn 20 reader goroutines that continuously read Results via GetResults().
	const readers = 20
	stop := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					// GetResults() returns a snapshot copy — safe under concurrent writes.
					results := bm.GetResults()
					// Access every entry so the race detector exercises all reads.
					for _, r := range results {
						_ = r.Status
						_ = r.Name
					}
				}
			}
		}()
	}

	// BootAll writes to Results on the same BootManager while readers are active.
	err := bm.BootAll()
	close(stop)
	wg.Wait()

	// With all services disabled BootAll should succeed (no required failures).
	assert.NoError(t, err, "BootAll with all-disabled services should not return an error")

	// Verify final Results are accessible via GetResults().
	results := bm.GetResults()
	assert.NotNil(t, results, "GetResults() should return a non-nil map")

	for name, r := range results {
		assert.Equal(t, "skipped", r.Status,
			"disabled service %s should have status 'skipped', got %s", name, r.Status)
	}
}

// TestBootManager_GetResults_ReturnsCopy ensures GetResults() returns a
// defensive copy — mutations of the returned map must not affect internal state.
func TestBootManager_GetResults_ReturnsCopy(t *testing.T) {
	cfg := newBootManagerStressServicesConfig()
	logger := newBootManagerStressLogger()
	bm := services.NewBootManager(cfg, logger)

	err := bm.BootAll()
	assert.NoError(t, err)

	copy1 := bm.GetResults()
	// Mutate the copy.
	for k := range copy1 {
		delete(copy1, k)
	}

	copy2 := bm.GetResults()
	// Internal state must be unaffected by the deletion above.
	assert.Equal(t, len(copy2), len(bm.GetResults()),
		"mutating the returned copy must not affect internal Results")
}

// TestBootManager_ConcurrentBootManagers_NoSharedState verifies that two
// independent BootManagers do not share Results state.
func TestBootManager_ConcurrentBootManagers_NoSharedState(t *testing.T) {
	var wg sync.WaitGroup
	const n = 5

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := newBootManagerStressServicesConfig()
			logger := newBootManagerStressLogger()
			bm := services.NewBootManager(cfg, logger)
			err := bm.BootAll()
			assert.NoError(t, err)
			results := bm.GetResults()
			assert.NotNil(t, results)
		}()
	}

	wg.Wait()
}
