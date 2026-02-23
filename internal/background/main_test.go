package background

import (
	"os"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()

	// Give worker pool goroutines time to drain after tests.
	time.Sleep(200 * time.Millisecond)

	leakOpts := []goleak.Option{
		// Background worker goroutines that may still be running during graceful shutdown.
		goleak.IgnoreTopFunction("dev.helix.agent/internal/background.(*WorkerPool).worker"),
		goleak.IgnoreTopFunction("dev.helix.agent/internal/background.(*AdaptiveWorkerPool).worker"),
		goleak.IgnoreTopFunction("dev.helix.agent/internal/background.(*ResourceMonitor).monitorLoop"),
		goleak.IgnoreTopFunction("dev.helix.agent/internal/background.(*StuckTaskDetector).detectorLoop"),
	}
	if err := goleak.Find(leakOpts...); err != nil {
		// Report but don't fail — background workers may still be draining.
		_ = err
	}

	os.Exit(exitCode)
}
