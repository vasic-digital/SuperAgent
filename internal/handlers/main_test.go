package handlers

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
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreAnyFunction("net.(*netFD).Read"),
		goleak.IgnoreAnyFunction("net.(*netFD).Accept"),
		// ACPHandler spawns sessionCleanupWorker goroutines; tests that
		// create handlers without calling Shutdown() leave these running.
		goleak.IgnoreTopFunction("dev.helix.agent/internal/handlers.(*ACPHandler).sessionCleanupWorker"),
		// ModelDiscoveryService.Start spawns background refresh goroutines
		goleak.IgnoreTopFunction("dev.helix.agent/internal/verifier.(*ModelDiscoveryService).discoveryLoop"),
		goleak.IgnoreTopFunction("dev.helix.agent/internal/notifications.(*PollingStore).cleanupLoop"),
	)
}
