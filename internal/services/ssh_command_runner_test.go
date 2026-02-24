package services

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"testing"

	"dev.helix.agent/internal/config"
	"digital.vasic.containers/pkg/remote"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// =====================================================
// MOCK REMOTE EXECUTOR
// =====================================================

// mockRemoteExecutor implements remote.RemoteExecutor for unit testing.
type mockRemoteExecutor struct {
	executeFunc     func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error)
	copyFileFunc    func(ctx context.Context, host remote.RemoteHost, localPath, remotePath string) error
	copyDirFunc     func(ctx context.Context, host remote.RemoteHost, localDir, remoteDir string) error
	isReachableFunc func(ctx context.Context, host remote.RemoteHost) bool
}

func (m *mockRemoteExecutor) Execute(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, host, command)
	}
	return &remote.CommandResult{Stdout: "ok", ExitCode: 0}, nil
}

func (m *mockRemoteExecutor) ExecuteStream(ctx context.Context, host remote.RemoteHost, command string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (m *mockRemoteExecutor) CopyFile(ctx context.Context, host remote.RemoteHost, localPath, remotePath string) error {
	if m.copyFileFunc != nil {
		return m.copyFileFunc(ctx, host, localPath, remotePath)
	}
	return nil
}

func (m *mockRemoteExecutor) CopyDir(ctx context.Context, host remote.RemoteHost, localDir, remoteDir string) error {
	if m.copyDirFunc != nil {
		return m.copyDirFunc(ctx, host, localDir, remoteDir)
	}
	return nil
}

func (m *mockRemoteExecutor) IsReachable(ctx context.Context, host remote.RemoteHost) bool {
	if m.isReachableFunc != nil {
		return m.isReachableFunc(ctx, host)
	}
	return true
}

// =====================================================
// SSHCommandRunner INTERFACE COMPLIANCE
// =====================================================

func TestDefaultSSHCommandRunner_ImplementsInterface(t *testing.T) {
	// Verify compile-time interface compliance
	var _ SSHCommandRunner = (*defaultSSHCommandRunner)(nil)
}

// =====================================================
// NewDefaultSSHCommandRunner TESTS
// =====================================================

func TestNewDefaultSSHCommandRunner_CreatesRunner(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.RemoteDeploymentConfig{
		SSHKey:         "/tmp/test_key",
		DefaultSSHUser: "testuser",
	}

	runner := NewDefaultSSHCommandRunner(cfg, logger)
	require.NotNil(t, runner)
}

// =====================================================
// toRemoteHost TESTS
// =====================================================

func TestDefaultSSHCommandRunner_ToRemoteHost(t *testing.T) {
	tests := []struct {
		name         string
		host         *config.RemoteDeploymentHost
		config       *config.RemoteDeploymentConfig
		expectUser   string
		expectAddr   string
		expectKey    string
	}{
		{
			name: "host with user@address format",
			host: &config.RemoteDeploymentHost{
				SSHHost: "admin@192.168.1.100",
			},
			config: &config.RemoteDeploymentConfig{
				SSHKey: "/home/user/.ssh/id_rsa",
			},
			expectUser: "admin",
			expectAddr: "192.168.1.100",
			expectKey:  "/home/user/.ssh/id_rsa",
		},
		{
			name: "host with override SSH key",
			host: &config.RemoteDeploymentHost{
				SSHHost: "admin@host1",
				SSHKey:  "/custom/key",
			},
			config: &config.RemoteDeploymentConfig{
				SSHKey: "/default/key",
			},
			expectUser: "admin",
			expectAddr: "host1",
			expectKey:  "/custom/key",
		},
		{
			name: "host without user uses default",
			host: &config.RemoteDeploymentHost{
				SSHHost: "192.168.1.50",
			},
			config: &config.RemoteDeploymentConfig{
				SSHKey:         "/home/user/.ssh/id_rsa",
				DefaultSSHUser: "deploy",
			},
			expectUser: "deploy",
			expectAddr: "192.168.1.50",
			expectKey:  "/home/user/.ssh/id_rsa",
		},
		{
			name: "host without user and no default",
			host: &config.RemoteDeploymentHost{
				SSHHost: "somehost.example.com",
			},
			config: &config.RemoteDeploymentConfig{
				SSHKey: "/path/to/key",
			},
			expectUser: "",
			expectAddr: "somehost.example.com",
			expectKey:  "/path/to/key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.PanicLevel)

			runner := &defaultSSHCommandRunner{
				logger:   logger,
				config:   tt.config,
				executor: nil,
			}

			rh := runner.toRemoteHost(tt.host)
			assert.Equal(t, tt.expectUser, rh.User)
			assert.Equal(t, tt.expectAddr, rh.Address)
			assert.Equal(t, tt.expectKey, rh.KeyPath)
			assert.Equal(t, 22, rh.Port)
			assert.Equal(t, "docker", rh.Runtime)
		})
	}
}

func TestDefaultSSHCommandRunner_ToRemoteHost_EmptySSHKey_DefaultsFallback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	runner := &defaultSSHCommandRunner{
		logger: logger,
		config: &config.RemoteDeploymentConfig{
			SSHKey: "", // Empty, triggers $HOME/.ssh/id_rsa fallback
		},
		executor: nil,
	}

	host := &config.RemoteDeploymentHost{
		SSHHost: "user@host",
		SSHKey:  "", // Also empty
	}

	rh := runner.toRemoteHost(host)
	// Should fall back to os.ExpandEnv("$HOME/.ssh/id_rsa")
	assert.Contains(t, rh.KeyPath, ".ssh/id_rsa")
}

// =====================================================
// RunSSHCommand TESTS
// =====================================================

func TestDefaultSSHCommandRunner_RunSSHCommand_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		executeFunc: func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
			return &remote.CommandResult{
				Stdout:   "command output",
				Stderr:   "",
				ExitCode: 0,
			}, nil
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	output, err := runner.RunSSHCommand(host, "echo hello")

	require.NoError(t, err)
	assert.Equal(t, "command output", output)
}

func TestDefaultSSHCommandRunner_RunSSHCommand_WithStderr(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		executeFunc: func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
			return &remote.CommandResult{
				Stdout:   "output",
				Stderr:   "warning: something",
				ExitCode: 0,
			}, nil
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	output, err := runner.RunSSHCommand(host, "test command")

	require.NoError(t, err)
	assert.Contains(t, output, "output")
	assert.Contains(t, output, "warning: something")
}

func TestDefaultSSHCommandRunner_RunSSHCommand_NonZeroExitCode(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		executeFunc: func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
			return &remote.CommandResult{
				Stdout:   "",
				Stderr:   "command not found",
				ExitCode: 127,
			}, nil
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	output, err := runner.RunSSHCommand(host, "nonexistent-command")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit code 127")
	assert.Contains(t, output, "command not found")
}

func TestDefaultSSHCommandRunner_RunSSHCommand_ExecutorError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		executeFunc: func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	output, err := runner.RunSSHCommand(host, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Empty(t, output)
}

func TestDefaultSSHCommandRunner_RunSSHCommand_NilExecutor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: nil,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	output, err := runner.RunSSHCommand(host, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSH executor not available")
	assert.Empty(t, output)
}

// =====================================================
// SCPFile TESTS
// =====================================================

func TestDefaultSSHCommandRunner_SCPFile_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	var capturedLocalPath, capturedRemotePath string
	executor := &mockRemoteExecutor{
		copyFileFunc: func(ctx context.Context, host remote.RemoteHost, localPath, remotePath string) error {
			capturedLocalPath = localPath
			capturedRemotePath = remotePath
			return nil
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	err := runner.SCPFile("/local/file.txt", host, "/remote/file.txt")

	require.NoError(t, err)
	assert.Equal(t, "/local/file.txt", capturedLocalPath)
	assert.Equal(t, "/remote/file.txt", capturedRemotePath)
}

func TestDefaultSSHCommandRunner_SCPFile_Error(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		copyFileFunc: func(ctx context.Context, host remote.RemoteHost, localPath, remotePath string) error {
			return fmt.Errorf("permission denied")
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	err := runner.SCPFile("/local/file.txt", host, "/remote/file.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scp failed")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestDefaultSSHCommandRunner_SCPFile_NilExecutor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: nil,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	err := runner.SCPFile("/local/file.txt", host, "/remote/file.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSH executor not available")
}

// =====================================================
// SCPDir TESTS
// =====================================================

func TestDefaultSSHCommandRunner_SCPDir_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	var capturedLocalDir, capturedRemoteDir string
	executor := &mockRemoteExecutor{
		copyDirFunc: func(ctx context.Context, host remote.RemoteHost, localDir, remoteDir string) error {
			capturedLocalDir = localDir
			capturedRemoteDir = remoteDir
			return nil
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	err := runner.SCPDir("/local/dir", host, "/remote/dir")

	require.NoError(t, err)
	assert.Equal(t, "/local/dir", capturedLocalDir)
	assert.Equal(t, "/remote/dir", capturedRemoteDir)
}

func TestDefaultSSHCommandRunner_SCPDir_Error(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		copyDirFunc: func(ctx context.Context, host remote.RemoteHost, localDir, remoteDir string) error {
			return fmt.Errorf("directory not found")
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	err := runner.SCPDir("/local/dir", host, "/remote/dir")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scp dir failed")
	assert.Contains(t, err.Error(), "directory not found")
}

func TestDefaultSSHCommandRunner_SCPDir_NilExecutor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: nil,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
	err := runner.SCPDir("/local/dir", host, "/remote/dir")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSH executor not available")
}

// =====================================================
// logrusRemoteAdapter TESTS
// =====================================================

func TestLogrusRemoteAdapter_Debug(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	adapter := &logrusRemoteAdapter{logger: logger}

	// Should not panic
	adapter.Debug("test debug %s", "message")
}

func TestLogrusRemoteAdapter_Info(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	adapter := &logrusRemoteAdapter{logger: logger}

	adapter.Info("test info %s", "message")
}

func TestLogrusRemoteAdapter_Warn(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	adapter := &logrusRemoteAdapter{logger: logger}

	adapter.Warn("test warn %s", "message")
}

func TestLogrusRemoteAdapter_Error(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	adapter := &logrusRemoteAdapter{logger: logger}

	adapter.Error("test error %s", "message")
}

// =====================================================
// TABLE-DRIVEN COMMAND TESTS
// =====================================================

func TestDefaultSSHCommandRunner_RunSSHCommand_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		stdout      string
		stderr      string
		exitCode    int
		execErr     error
		expectErr   bool
		expectInOut string
	}{
		{
			name:        "successful simple command",
			stdout:      "hello world",
			stderr:      "",
			exitCode:    0,
			expectErr:   false,
			expectInOut: "hello world",
		},
		{
			name:        "command with stderr but success",
			stdout:      "result",
			stderr:      "deprecation warning",
			exitCode:    0,
			expectErr:   false,
			expectInOut: "result",
		},
		{
			name:        "command failure exit code 1",
			stdout:      "",
			stderr:      "error: file not found",
			exitCode:    1,
			expectErr:   true,
			expectInOut: "error: file not found",
		},
		{
			name:      "executor connection error",
			execErr:   fmt.Errorf("connection timeout"),
			expectErr: true,
		},
		{
			name:        "command with exit code 2",
			stdout:      "partial",
			stderr:      "fatal error",
			exitCode:    2,
			expectErr:   true,
			expectInOut: "partial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.PanicLevel)

			executor := &mockRemoteExecutor{
				executeFunc: func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
					if tt.execErr != nil {
						return nil, tt.execErr
					}
					return &remote.CommandResult{
						Stdout:   tt.stdout,
						Stderr:   tt.stderr,
						ExitCode: tt.exitCode,
					}, nil
				},
			}

			runner := &defaultSSHCommandRunner{
				logger:   logger,
				config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
				executor: executor,
			}

			host := &config.RemoteDeploymentHost{SSHHost: "user@host"}
			output, err := runner.RunSSHCommand(host, "test-cmd")

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectInOut != "" {
				assert.Contains(t, output, tt.expectInOut)
			}
		})
	}
}

// =====================================================
// BENCHMARK TESTS
// =====================================================

func BenchmarkDefaultSSHCommandRunner_RunSSHCommand(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	executor := &mockRemoteExecutor{
		executeFunc: func(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
			return &remote.CommandResult{Stdout: "ok", ExitCode: 0}, nil
		},
	}

	runner := &defaultSSHCommandRunner{
		logger:   logger,
		config:   &config.RemoteDeploymentConfig{SSHKey: "/test/key"},
		executor: executor,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "user@host"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.RunSSHCommand(host, "echo hello")
	}
}

func BenchmarkDefaultSSHCommandRunner_ToRemoteHost(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	runner := &defaultSSHCommandRunner{
		logger: logger,
		config: &config.RemoteDeploymentConfig{
			SSHKey:         "/test/key",
			DefaultSSHUser: "deploy",
		},
		executor: nil,
	}

	host := &config.RemoteDeploymentHost{SSHHost: "admin@192.168.1.100"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.toRemoteHost(host)
	}
}
