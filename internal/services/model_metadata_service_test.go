package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultModelMetadataConfig(t *testing.T) {
	config := getDefaultModelMetadataConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 24*time.Hour, config.RefreshInterval)
	assert.Equal(t, 1*time.Hour, config.CacheTTL)
	assert.Equal(t, 100, config.DefaultBatchSize)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5*time.Second, config.RetryDelay)
	assert.True(t, config.EnableAutoRefresh)
}
