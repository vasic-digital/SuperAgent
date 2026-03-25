package cache

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// NOTE: goleak.VerifyTestMain calls m.Run() internally.
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
		// TieredCache background cleanup goroutine — self-terminates when cache is closed.
		goleak.IgnoreTopFunction("dev.helix.agent/internal/cache.(*TieredCache).l1CleanupLoop"),
		// Redis client circuit-breaker maintenance goroutine — self-terminates when client closes.
		goleak.IgnoreTopFunction("github.com/redis/go-redis/v9/maintnotifications.(*CircuitBreakerManager).cleanupLoop"),
	)
}
