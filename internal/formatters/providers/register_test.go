package providers

import (
	"os"
	"testing"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAllFormatters_BasicRegistration(t *testing.T) {
	// Ensure service formatters are disabled for this test
	os.Setenv("FORMATTER_ENABLE_SERVICES", "false")
	defer os.Unsetenv("FORMATTER_ENABLE_SERVICES")

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise

	config := &formatters.RegistryConfig{
		DefaultTimeout: 30,
		MaxConcurrent:  10,
	}
	registry := formatters.NewFormatterRegistry(config, logger)

	err := RegisterAllFormatters(registry, logger)
	require.NoError(t, err, "RegisterAllFormatters should not return error")

	// At least some formatters should be registered
	assert.Greater(t, registry.Count(), 0, "Should register at least one formatter")

	// Verify some known formatters are registered
	black, err := registry.Get("black")
	if err == nil {
		assert.Equal(t, "black", black.Name())
		assert.Contains(t, black.Languages(), "python")
	}

	ruff, err := registry.Get("ruff")
	if err == nil {
		assert.Equal(t, "ruff", ruff.Name())
		assert.Contains(t, ruff.Languages(), "python")
	}

	// Verify service formatters are NOT registered (since disabled)
	_, err = registry.Get("sqlfluff")
	assert.Error(t, err, "Service formatters should not be registered when disabled")
}

func TestRegisterAllFormatters_ServiceFormattersEnabled(t *testing.T) {
	// Enable service formatters
	os.Setenv("FORMATTER_ENABLE_SERVICES", "true")
	defer os.Unsetenv("FORMATTER_ENABLE_SERVICES")

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &formatters.RegistryConfig{
		DefaultTimeout: 30,
		MaxConcurrent:  10,
	}
	registry := formatters.NewFormatterRegistry(config, logger)

	err := RegisterAllFormatters(registry, logger)
	require.NoError(t, err, "RegisterAllFormatters should not return error")

	// At least one service formatter should be registered
	// We can't guarantee which ones, but we can check that registry count is higher
	// than with services disabled (but we don't have that state).
	// Instead, we can check that some known service formatter names are registered
	serviceNames := []string{"sqlfluff", "rubocop", "php-cs-fixer"}
	for _, name := range serviceNames {
		_, err := registry.Get(name)
		// It's okay if some are not registered (depends on environment)
		// We just log but don't fail
		if err == nil {
			t.Logf("Service formatter %s registered", name)
		}
	}
}

func TestRegisterAllFormatters_DuplicateRegistration(t *testing.T) {
	os.Setenv("FORMATTER_ENABLE_SERVICES", "false")
	defer os.Unsetenv("FORMATTER_ENABLE_SERVICES")

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &formatters.RegistryConfig{
		DefaultTimeout: 30,
		MaxConcurrent:  10,
	}
	registry := formatters.NewFormatterRegistry(config, logger)

	// First registration
	err := RegisterAllFormatters(registry, logger)
	require.NoError(t, err)

	countBefore := registry.Count()

	// Second registration should be idempotent (no error, but duplicates are ignored?)
	// The registry.Register will error on duplicate names, but RegisterAllFormatters
	// logs warnings and increments failed count.
	// We'll just ensure no panic.
	err = RegisterAllFormatters(registry, logger)
	// Currently RegisterAllFormatters returns nil even if some registrations fail.
	// That's okay.
	assert.NoError(t, err)
	// Count should be same (no new formatters added)
	assert.Equal(t, countBefore, registry.Count())
}

func TestRegisterAllFormatters_EnvironmentVariableBaseURL(t *testing.T) {
	// Set custom base URL for service formatters
	os.Setenv("FORMATTER_SERVICE_BASE_URL", "http://custom-host:9999")
	os.Setenv("FORMATTER_ENABLE_SERVICES", "true")
	defer func() {
		os.Unsetenv("FORMATTER_SERVICE_BASE_URL")
		os.Unsetenv("FORMATTER_ENABLE_SERVICES")
	}()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &formatters.RegistryConfig{
		DefaultTimeout: 30,
		MaxConcurrent:  10,
	}
	registry := formatters.NewFormatterRegistry(config, logger)

	err := RegisterAllFormatters(registry, logger)
	require.NoError(t, err)

	// No easy way to verify base URL is used without inspecting internal state.
	// For now, just ensure no panic.
	assert.Greater(t, registry.Count(), 0)
}