package eventsadapter

import (
	"testing"
	"time"

	"dev.helix.agent/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetGlobal() {
	// Ensure clean state between tests.
	_ = Shutdown()
}

func TestInitialize_DefaultConfig(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	Initialize(nil)

	assert.True(t, IsInitialized())
}

func TestInitialize_CustomConfig(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	cfg := &events.BusConfig{
		BufferSize:      500,
		PublishTimeout:  50 * time.Millisecond,
		CleanupInterval: 10 * time.Second,
		MaxSubscribers:  50,
	}
	Initialize(cfg)

	bus := GetBus()
	require.NotNil(t, bus)
	assert.True(t, IsInitialized())
}

func TestInitialize_Idempotent(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	Initialize(nil)
	bus1 := GetBus()

	// Second call should be a no-op; same bus returned.
	Initialize(&events.BusConfig{BufferSize: 999})
	bus2 := GetBus()

	assert.Same(t, bus1, bus2)
}

func TestGetBus_LazyInit(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	assert.False(t, IsInitialized())

	bus := GetBus()
	require.NotNil(t, bus)
	assert.True(t, IsInitialized())
}

func TestGetBus_PublishSubscribe(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	bus := GetBus()
	require.NotNil(t, bus)

	ch := bus.Subscribe(events.EventProviderRegistered)
	require.NotNil(t, ch)

	evt := events.NewEvent(events.EventProviderRegistered, "adapter-test", "payload-1")
	bus.Publish(evt)

	select {
	case received := <-ch:
		assert.Equal(t, events.EventProviderRegistered, received.Type)
		assert.Equal(t, "adapter-test", received.Source)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestShutdown_ClosesSubscribers(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	bus := GetBus()
	ch := bus.Subscribe(events.EventSystemStartup)

	err := Shutdown()
	require.NoError(t, err)

	// Channel should be closed after shutdown.
	_, ok := <-ch
	assert.False(t, ok, "subscriber channel should be closed after Shutdown")
	assert.False(t, IsInitialized())
}

func TestShutdown_NilBus(t *testing.T) {
	resetGlobal()

	// Shutdown on nil bus should succeed.
	err := Shutdown()
	assert.NoError(t, err)
}

func TestShutdown_DoubleShutdown(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	Initialize(nil)

	err := Shutdown()
	assert.NoError(t, err)

	err = Shutdown()
	assert.NoError(t, err)
}

func TestReinitializeAfterShutdown(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	Initialize(nil)
	bus1 := GetBus()
	require.NotNil(t, bus1)

	err := Shutdown()
	require.NoError(t, err)

	// Should be able to re-initialize after shutdown.
	Initialize(nil)
	bus2 := GetBus()
	require.NotNil(t, bus2)
	assert.True(t, IsInitialized())

	// New bus should work.
	ch := bus2.Subscribe(events.EventDebateStarted)
	bus2.Publish(events.NewEvent(events.EventDebateStarted, "reinit-test", nil))

	select {
	case received := <-ch:
		assert.Equal(t, events.EventDebateStarted, received.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event after re-initialization")
	}
}

func TestIsInitialized_FalseByDefault(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	assert.False(t, IsInitialized())
}

func TestGetBus_Metrics(t *testing.T) {
	resetGlobal()
	defer resetGlobal()

	bus := GetBus()
	require.NotNil(t, bus)

	ch := bus.Subscribe(events.EventCacheHit)
	bus.Publish(events.NewEvent(events.EventCacheHit, "metrics-test", nil))

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	metrics := bus.Metrics()
	require.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.EventsPublished, int64(1))
	assert.GreaterOrEqual(t, metrics.EventsDelivered, int64(1))
}
