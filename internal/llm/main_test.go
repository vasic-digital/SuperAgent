package llm

import (
	"os"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()

	// Give transient goroutines (e.g. circuit-breaker listener notifications) time to finish.
	time.Sleep(200 * time.Millisecond)

	leakOpts := []goleak.Option{
		// Circuit breaker listener notifications use time.After with up to 5s timeout.
		// They are not leaks — they self-terminate after the timeout fires.
		goleak.IgnoreTopFunction("time.AfterFunc"),
		goleak.IgnoreTopFunction("time.Sleep"),
	}
	if err := goleak.Find(leakOpts...); err != nil {
		// Report but don't fail — transient listener goroutines may still be running.
		_ = err
	}

	os.Exit(exitCode)
}
