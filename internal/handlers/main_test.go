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
	)
}
