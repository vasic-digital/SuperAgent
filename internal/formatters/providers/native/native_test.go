package native

import (
	"context"
	"testing"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNativeFormatter(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	metadata := &formatters.FormatterMetadata{
		Name:            "test-formatter",
		Type:            formatters.FormatterTypeNative,
		Version:         "1.0.0",
		Languages:       []string{"testlang"},
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	formatter := NewNativeFormatter(metadata, "fake-binary", []string{"--quiet"}, true, logger)
	assert.NotNil(t, formatter)
	assert.Equal(t, "test-formatter", formatter.Name())
	assert.Equal(t, "1.0.0", formatter.Version())
	assert.Contains(t, formatter.Languages(), "testlang")
	assert.True(t, formatter.SupportsStdin())
	assert.True(t, formatter.SupportsInPlace())
	assert.True(t, formatter.SupportsCheck())
	assert.True(t, formatter.SupportsConfig())
}

func TestNativeFormatter_buildArgs(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	metadata := &formatters.FormatterMetadata{
		Name:      "test",
		Type:      formatters.FormatterTypeNative,
		Languages: []string{"test"},
	}

	testCases := []struct {
		name          string
		args          []string
		stdinFlag     bool
		checkOnly     bool
		supportsCheck bool
		expected      []string
	}{
		{
			name:      "basic args",
			args:      []string{"--quiet"},
			stdinFlag: false,
			expected:  []string{"--quiet"},
		},
		{
			name:      "with stdin flag",
			args:      []string{"--quiet"},
			stdinFlag: true,
			expected:  []string{"--quiet", "-"},
		},
		{
			name:          "check only with support",
			args:          []string{"--quiet"},
			stdinFlag:     false,
			checkOnly:     true,
			supportsCheck: true,
			expected:      []string{"--quiet", "--check"},
		},
		{
			name:          "check only without support",
			args:          []string{"--quiet"},
			stdinFlag:     false,
			checkOnly:     true,
			supportsCheck: false,
			expected:      []string{"--quiet"},
		},
		{
			name:          "stdin and check",
			args:          []string{"--quiet"},
			stdinFlag:     true,
			checkOnly:     true,
			supportsCheck: true,
			expected:      []string{"--quiet", "-", "--check"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata.SupportsCheck = tc.supportsCheck
			formatter := &NativeFormatter{
				BaseFormatter: formatters.NewBaseFormatter(metadata),
				binaryPath:    "fake",
				args:          tc.args,
				stdinFlag:     tc.stdinFlag,
				logger:        logger,
			}

			req := &formatters.FormatRequest{
				CheckOnly: tc.checkOnly,
			}
			args := formatter.buildArgs(req)
			assert.Equal(t, tc.expected, args)
		})
	}
}

func TestNativeFormatter_HealthCheck_BinaryMissing(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	metadata := &formatters.FormatterMetadata{
		Name:      "test",
		Type:      formatters.FormatterTypeNative,
		Languages: []string{"test"},
	}

	// Use a binary that definitely doesn't exist
	formatter := NewNativeFormatter(metadata, "/nonexistent/binary/that/does/not/exist", nil, false, logger)
	ctx := context.Background()
	err := formatter.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "formatter binary not available")
}

func TestNativeFormatter_FormatBatch(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	metadata := &formatters.FormatterMetadata{
		Name:      "test",
		Type:      formatters.FormatterTypeNative,
		Languages: []string{"test"},
	}

	formatter := NewNativeFormatter(metadata, "fake-binary", []string{"--quiet"}, true, logger)
	ctx := context.Background()
	reqs := []*formatters.FormatRequest{
		{Content: "content1", Language: "test"},
		{Content: "content2", Language: "test"},
	}

	// Since the binary doesn't exist, Format will fail but return a result with error
	results, err := formatter.FormatBatch(ctx, reqs)
	require.NoError(t, err) // FormatBatch returns nil error, errors are in results
	require.Len(t, results, 2)
	for _, result := range results {
		assert.False(t, result.Success)
		assert.Contains(t, result.Error.Error(), "formatter execution failed")
	}
}

func TestNativeFormatter_Format_MissingBinary(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	metadata := &formatters.FormatterMetadata{
		Name:      "test",
		Type:      formatters.FormatterTypeNative,
		Languages: []string{"test"},
	}

	formatter := NewNativeFormatter(metadata, "/nonexistent/binary/that/does/not/exist", []string{"--quiet"}, true, logger)
	ctx := context.Background()
	req := &formatters.FormatRequest{Content: "test", Language: "test"}
	result, err := formatter.Format(ctx, req)
	require.NoError(t, err) // Format returns result with error, not an error
	assert.False(t, result.Success)
	assert.Contains(t, result.Error.Error(), "formatter execution failed")
}
