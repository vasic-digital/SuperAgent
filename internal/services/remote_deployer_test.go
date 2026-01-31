package services

import (
	"context"
	"testing"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHRemoteDeployer_FindHostForService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled: true,
			Hosts: map[string]config.RemoteDeploymentHost{
				"host1": {
					SSHHost:  "user@host1",
					Services: []string{"postgresql", "redis"},
				},
				"host2": {
					SSHHost:  "user@host2",
					Services: []string{"cognee"},
				},
			},
		},
	}

	deployer := NewSSHRemoteDeployer(cfg, logger)

	// Test service mapped to host1
	host, err := deployer.findHostForService("postgresql")
	require.NoError(t, err)
	assert.Equal(t, "user@host1", host.SSHHost)

	host, err = deployer.findHostForService("redis")
	require.NoError(t, err)
	assert.Equal(t, "user@host1", host.SSHHost)

	// Test service mapped to host2
	host, err = deployer.findHostForService("cognee")
	require.NoError(t, err)
	assert.Equal(t, "user@host2", host.SSHHost)

	// Test service not mapped
	host, err = deployer.findHostForService("unknown")
	assert.Error(t, err)
	assert.Nil(t, host)
	assert.Contains(t, err.Error(), "no deployment host configured")
}

func TestSSHRemoteDeployer_FindHostForService_SingleHostFallback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled: true,
			Hosts: map[string]config.RemoteDeploymentHost{
				"single": {
					SSHHost:  "user@single",
					Services: []string{}, // empty list
				},
			},
		},
	}

	deployer := NewSSHRemoteDeployer(cfg, logger)

	// Any service should fall back to the single host
	host, err := deployer.findHostForService("any-service")
	require.NoError(t, err)
	assert.Equal(t, "user@single", host.SSHHost)
}

func TestSSHRemoteDeployer_Deploy_NonRemoteService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled: true,
			Hosts:   map[string]config.RemoteDeploymentHost{},
		},
	}

	deployer := NewSSHRemoteDeployer(cfg, logger)

	endpoint := &config.ServiceEndpoint{
		Remote: false,
	}

	err := deployer.Deploy(context.Background(), "postgresql", endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not configured as remote")
}

// mockSSHCommandRunner implements SSHCommandRunner for testing.
type mockSSHCommandRunner struct {
	commands []string
	errors   map[string]error
}

func (m *mockSSHCommandRunner) RunSSHCommand(host *config.RemoteDeploymentHost, command string) (string, error) {
	m.commands = append(m.commands, command)
	if err, ok := m.errors[command]; ok {
		return "", err
	}
	return "mocked output", nil
}

func (m *mockSSHCommandRunner) SCPFile(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
	m.commands = append(m.commands, "scp:"+localPath+"->"+remotePath)
	return nil
}

func (m *mockSSHCommandRunner) SCPDir(localDir string, host *config.RemoteDeploymentHost, remoteDir string) error {
	m.commands = append(m.commands, "scpdir:"+localDir+"->"+remoteDir)
	return nil
}

func TestSSHRemoteDeployer_EnsureRemoteDirectory(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled:          true,
			DefaultRemoteDir: "/opt/helixagent-test",
			Hosts: map[string]config.RemoteDeploymentHost{
				"host1": {
					SSHHost: "user@host1",
				},
			},
		},
	}

	mockRunner := &mockSSHCommandRunner{
		commands: []string{},
		errors:   map[string]error{},
	}
	deployer := NewSSHRemoteDeployerWithRunner(cfg, logger, mockRunner)
	host := &config.RemoteDeploymentHost{SSHHost: "user@host1"}

	err := deployer.ensureRemoteDirectory(host)
	require.NoError(t, err)

	// Should have called mkdir -p
	assert.Len(t, mockRunner.commands, 1)
	assert.Contains(t, mockRunner.commands[0], "mkdir -p")
	assert.Contains(t, mockRunner.commands[0], "/opt/helixagent-test")
}

func TestSSHRemoteDeployer_DeployAll_NoRemoteServices(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled: true,
			Hosts:   map[string]config.RemoteDeploymentHost{},
		},
	}

	deployer := NewSSHRemoteDeployer(cfg, logger)
	err := deployer.DeployAll(context.Background())
	assert.NoError(t, err) // Should skip non-remote services
}

func TestSSHRemoteDeployer_HealthCheckRemote_NoRemoteServices(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		RemoteDeployment: config.RemoteDeploymentConfig{
			Enabled: true,
			Hosts:   map[string]config.RemoteDeploymentHost{},
		},
	}

	deployer := NewSSHRemoteDeployer(cfg, logger)
	err := deployer.HealthCheckRemote(context.Background())
	assert.NoError(t, err)
}
