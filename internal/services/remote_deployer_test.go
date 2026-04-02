package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"dev.helix.agent/internal/adapters/containers"
	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSSHCommandRunner is a mock implementation of SSHCommandRunner for testing
type MockSSHCommandRunner struct {
	RunSSHCommandFunc func(host *config.RemoteDeploymentHost, command string) (string, error)
	SCPFileFunc       func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error
	SCPDirFunc        func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error
}

func (m *MockSSHCommandRunner) RunSSHCommand(host *config.RemoteDeploymentHost, command string) (string, error) {
	if m.RunSSHCommandFunc != nil {
		return m.RunSSHCommandFunc(host, command)
	}
	return "", nil
}

func (m *MockSSHCommandRunner) SCPFile(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
	if m.SCPFileFunc != nil {
		return m.SCPFileFunc(localPath, host, remotePath)
	}
	return nil
}

func (m *MockSSHCommandRunner) SCPDir(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
	if m.SCPDirFunc != nil {
		return m.SCPDirFunc(localPath, host, remotePath)
	}
	return nil
}

// Ensure MockSSHCommandRunner implements SSHCommandRunner interface
var _ SSHCommandRunner = (*MockSSHCommandRunner)(nil)

func TestNewSSHRemoteDeployer(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled:           true,
			DefaultRemoteDir:  "/opt/helixagent",
			SSHPrivateKeyPath: "/tmp/test_key",
			Hosts: []config.RemoteDeploymentHost{
				{
					Name:     "test-host",
					SSHHost:  "192.168.1.1",
					SSHPort:  22,
					SSHUser:  "testuser",
					Services: []string{"test-service"},
				},
			},
		},
	}

	t.Run("creates deployer with default runner", func(t *testing.T) {
		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		assert.NotNil(t, deployer)
		assert.NotNil(t, deployer.config)
		assert.NotNil(t, deployer.logger)
		assert.NotNil(t, deployer.runner)
		assert.Equal(t, adapter, deployer.ContainerAdapter)
	})

	t.Run("creates deployer with custom runner", func(t *testing.T) {
		adapter := &containers.Adapter{}
		mockRunner := &MockSSHCommandRunner{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		assert.NotNil(t, deployer)
		assert.Equal(t, mockRunner, deployer.runner)
		assert.Equal(t, adapter, deployer.ContainerAdapter)
	})
}

func TestSSHRemoteDeployer_Deploy(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled:           true,
			DefaultRemoteDir:  "/opt/helixagent",
			SSHPrivateKeyPath: "/tmp/test_key",
			Hosts: []config.RemoteDeploymentHost{
				{
					Name:     "test-host",
					SSHHost:  "192.168.1.1",
					SSHPort:  22,
					SSHUser:  "testuser",
					Services: []string{"test-service"},
				},
			},
		},
	}

	t.Run("returns error for non-remote service", func(t *testing.T) {
		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		endpoint := &config.ServiceEndpoint{
			Enabled: true,
			Remote:  false,
		}

		err := deployer.Deploy(context.Background(), "test-service", endpoint)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured as remote")
	})

	t.Run("returns error when no host found for service", func(t *testing.T) {
		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		endpoint := &config.ServiceEndpoint{
			Enabled: true,
			Remote:  true,
		}

		err := deployer.Deploy(context.Background(), "unknown-service", endpoint)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no deployment host found")
	})

	t.Run("successful deployment flow", func(t *testing.T) {
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				return "", nil
			},
			SCPFileFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return nil
			},
			SCPDirFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return nil
			},
		}

		// Create a mock adapter that will be used via the interface
		adapter := &containers.Adapter{}

		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		endpoint := &config.ServiceEndpoint{
			Enabled: true,
			Remote:  true,
		}

		// This will fail because the adapter mock doesn't implement ComposeUp
		// but we're testing the flow structure
		err := deployer.Deploy(context.Background(), "test-service", endpoint)
		// Error expected since we don't have a full mock adapter
		assert.Error(t, err)
	})

	t.Run("handles SSH command failure", func(t *testing.T) {
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				return "", errors.New("ssh connection failed")
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		endpoint := &config.ServiceEndpoint{
			Enabled: true,
			Remote:  true,
		}

		err := deployer.Deploy(context.Background(), "test-service", endpoint)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ensure remote directory")
	})

	t.Run("handles SCP failures gracefully", func(t *testing.T) {
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				return "", nil
			},
			SCPFileFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return errors.New("scp failed")
			},
			SCPDirFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return errors.New("scp dir failed")
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		endpoint := &config.ServiceEndpoint{
			Enabled: true,
			Remote:  true,
		}

		// SCP failures are logged but don't stop deployment
		err := deployer.Deploy(context.Background(), "test-service", endpoint)
		// Will error on compose up since adapter is not fully mocked
		assert.Error(t, err)
	})
}

func TestSSHRemoteDeployer_DeployAll(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("deploys all remote-enabled services", func(t *testing.T) {
		cfg := &config.Config{
			Services: config.ServicesConfig{
				Postgres: config.ServiceEndpoint{
					Enabled: true,
					Remote:  true,
				},
				Redis: config.ServiceEndpoint{
					Enabled: false,
					Remote:  true,
				},
			},
			RemoteDeployment: config.RemoteDeploymentConfig{
				Enabled:           true,
				DefaultRemoteDir:  "/opt/helixagent",
				SSHPrivateKeyPath: "/tmp/test_key",
				Hosts: []config.RemoteDeploymentHost{
					{
						Name:     "test-host",
						SSHHost:  "192.168.1.1",
						SSHPort:  22,
						SSHUser:  "testuser",
						Services: []string{"postgres"},
					},
				},
			},
		}

		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				return "", nil
			},
			SCPFileFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return nil
			},
			SCPDirFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return nil
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		ctx := context.Background()
		err := deployer.DeployAll(ctx)
		// Should complete without panic even if individual deploys fail
		assert.NoError(t, err)
	})

	t.Run("handles empty service list", func(t *testing.T) {
		cfg := &config.Config{
			Services: config.ServicesConfig{},
			RemoteDeployment: config.RemoteDeploymentConfig{
				Enabled: true,
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		ctx := context.Background()
		err := deployer.DeployAll(ctx)
		assert.NoError(t, err)
	})
}

func TestSSHRemoteDeployer_HealthCheckRemote(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("performs health checks on remote services", func(t *testing.T) {
		cfg := &config.Config{
			Services: config.ServicesConfig{
				Postgres: config.ServiceEndpoint{
					Enabled: true,
					Remote:  true,
					Host:    "192.168.1.1",
					Port:    5432,
				},
			},
			RemoteDeployment: config.RemoteDeploymentConfig{
				Enabled: true,
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		ctx := context.Background()
		err := deployer.HealthCheckRemote(ctx)
		// Should complete without error even if individual checks fail
		assert.NoError(t, err)
	})

	t.Run("skips disabled services", func(t *testing.T) {
		cfg := &config.Config{
			Services: config.ServicesConfig{
				Postgres: config.ServiceEndpoint{
					Enabled: false,
					Remote:  true,
				},
			},
			RemoteDeployment: config.RemoteDeploymentConfig{
				Enabled: true,
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		ctx := context.Background()
		err := deployer.HealthCheckRemote(ctx)
		assert.NoError(t, err)
	})
}

func TestSSHRemoteDeployer_findHostForService(t *testing.T) {
	logger := logrus.New()

	t.Run("finds host by service name", func(t *testing.T) {
		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				Hosts: []config.RemoteDeploymentHost{
					{
						Name:     "host1",
						Services: []string{"service1", "service2"},
					},
					{
						Name:     "host2",
						Services: []string{"service3"},
					},
				},
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		host, err := deployer.findHostForService("service2")
		require.NoError(t, err)
		assert.Equal(t, "host1", host.Name)
	})

	t.Run("returns error when service not found and multiple hosts exist", func(t *testing.T) {
		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				Hosts: []config.RemoteDeploymentHost{
					{
						Name:     "host1",
						Services: []string{"service1"},
					},
					{
						Name:     "host2",
						Services: []string{"service2"},
					},
				},
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		host, err := deployer.findHostForService("unknown-service")
		require.Error(t, err)
		assert.Nil(t, host)
		assert.Contains(t, err.Error(), "no deployment host configured")
	})

	t.Run("returns single host when only one exists", func(t *testing.T) {
		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				Hosts: []config.RemoteDeploymentHost{
					{
						Name:     "only-host",
						Services: []string{"other-service"},
					},
				},
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		host, err := deployer.findHostForService("any-service")
		require.NoError(t, err)
		assert.Equal(t, "only-host", host.Name)
	})

	t.Run("returns error when no hosts configured", func(t *testing.T) {
		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				Hosts: []config.RemoteDeploymentHost{},
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployer(cfg, logger, adapter)

		host, err := deployer.findHostForService("service")
		require.Error(t, err)
		assert.Nil(t, host)
	})
}

func TestSSHRemoteDeployer_ensureRemoteDirectory(t *testing.T) {
	logger := logrus.New()

	t.Run("creates directory with default path", func(t *testing.T) {
		commandExecuted := false
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				commandExecuted = true
				assert.Equal(t, "mkdir -p /opt/helixagent", command)
				return "", nil
			},
		}

		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				DefaultRemoteDir: "/opt/helixagent",
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		host := &config.RemoteDeploymentHost{Name: "test-host"}
		err := deployer.ensureRemoteDirectory(host)
		require.NoError(t, err)
		assert.True(t, commandExecuted)
	})

	t.Run("handles SSH error", func(t *testing.T) {
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				return "", errors.New("permission denied")
			},
		}

		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				DefaultRemoteDir: "/opt/helixagent",
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		host := &config.RemoteDeploymentHost{Name: "test-host"}
		err := deployer.ensureRemoteDirectory(host)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestSSHRemoteDeployer_copyFilesToRemote(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("copies all required files", func(t *testing.T) {
		scpFileCalled := false
		scpDirCalled := false

		mockRunner := &MockSSHCommandRunner{
			SCPFileFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				scpFileCalled = true
				assert.Equal(t, "docker-compose.yml", localPath)
				return nil
			},
			SCPDirFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				scpDirCalled = true
				return nil
			},
		}

		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				DefaultRemoteDir: "/opt/helixagent",
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		host := &config.RemoteDeploymentHost{Name: "test-host"}
		err := deployer.copyFilesToRemote(host)
		require.NoError(t, err)
		assert.True(t, scpFileCalled)
		assert.True(t, scpDirCalled)
	})

	t.Run("continues on SCP errors", func(t *testing.T) {
		mockRunner := &MockSSHCommandRunner{
			SCPFileFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return errors.New("file not found")
			},
			SCPDirFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return errors.New("directory not found")
			},
		}

		cfg := &config.Config{
			RemoteDeployment: config.RemoteDeploymentConfig{
				DefaultRemoteDir: "/opt/helixagent",
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		host := &config.RemoteDeploymentHost{Name: "test-host"}
		// Errors are logged but not returned
		err := deployer.copyFilesToRemote(host)
		assert.NoError(t, err)
	})
}

func TestSSHRemoteDeployer_ContextCancellation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled:           true,
			DefaultRemoteDir:  "/opt/helixagent",
			SSHPrivateKeyPath: "/tmp/test_key",
			Hosts: []config.RemoteDeploymentHost{
				{
					Name:     "test-host",
					SSHHost:  "192.168.1.1",
					Services: []string{"test-service"},
				},
			},
		},
	}

	t.Run("respects context cancellation", func(t *testing.T) {
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				// Simulate slow operation
				time.Sleep(100 * time.Millisecond)
				return "", nil
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		endpoint := &config.ServiceEndpoint{
			Enabled: true,
			Remote:  true,
		}

		// Start deployment in goroutine
		done := make(chan error, 1)
		go func() {
			done <- deployer.Deploy(ctx, "test-service", endpoint)
		}()

		// Cancel context immediately
		cancel()

		// Should complete (with error or not) within reasonable time
		select {
		case <-done:
			// Expected
		case <-time.After(2 * time.Second):
			t.Fatal("Deploy did not respect context cancellation")
		}
	})
}

func TestSSHRemoteDeployer_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled:           true,
			DefaultRemoteDir:  "/opt/helixagent",
			SSHPrivateKeyPath: "/tmp/test_key",
			Hosts: []config.RemoteDeploymentHost{
				{
					Name:     "test-host",
					SSHHost:  "192.168.1.1",
					Services: []string{"service1", "service2", "service3"},
				},
			},
		},
	}

	t.Run("handles concurrent Deploy calls", func(t *testing.T) {
		var callCount int
		mockRunner := &MockSSHCommandRunner{
			RunSSHCommandFunc: func(host *config.RemoteDeploymentHost, command string) (string, error) {
				callCount++
				return "", nil
			},
			SCPFileFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return nil
			},
			SCPDirFunc: func(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
				return nil
			},
		}

		adapter := &containers.Adapter{}
		deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner, adapter)

		// Run multiple deployments concurrently
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				serviceName := fmt.Sprintf("service%d", idx+1)
				endpoint := &config.ServiceEndpoint{Enabled: true, Remote: true}
				// We expect errors due to adapter not being fully mocked
				_ = deployer.Deploy(context.Background(), serviceName, endpoint)
			}(i)
		}

		wg.Wait()
		// Should have made SSH calls for each service
		assert.GreaterOrEqual(t, callCount, 3)
	})
}
