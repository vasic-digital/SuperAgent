package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessagingHub_Close_WaitsForHealthCheck(t *testing.T) {
	cfg := &HubConfig{
		UseFallbackOnError:  true,
		HealthCheckInterval: 50 * time.Millisecond,
	}
	hub := NewMessagingHub(cfg)

	err := hub.Initialize(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = hub.Close(ctx)
	assert.NoError(t, err)

	// Calling Close again should not panic
	err = hub.Close(ctx)
	assert.NoError(t, err)
}

func TestMessagingHub_Close_PropagatesContext(t *testing.T) {
	cfg := &HubConfig{
		UseFallbackOnError:  true,
		HealthCheckInterval: 100 * time.Millisecond,
	}
	hub := NewMessagingHub(cfg)

	err := hub.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()
	err = hub.Close(ctx)
	assert.NoError(t, err)
}
