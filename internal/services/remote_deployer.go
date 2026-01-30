package services

import (
	"context"
	"fmt"
	"os"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// RemoteDeployer defines the interface for deploying services to remote hosts.
type RemoteDeployer interface {
	// Deploy deploys a service to its configured remote host.
	Deploy(ctx context.Context, serviceName string, endpoint *config.ServiceEndpoint) error
	// DeployAll deploys all remote-enabled services.
	DeployAll(ctx context.Context) error
	// HealthCheckRemote performs health checks on remote services.
	HealthCheckRemote(ctx context.Context) error
}

// SSHRemoteDeployer implements RemoteDeployer using SSH and Docker commands.
type SSHRemoteDeployer struct {
	config *config.Config
	logger *logrus.Logger
	runner SSHCommandRunner
}

// NewSSHRemoteDeployer creates a new SSHRemoteDeployer.
func NewSSHRemoteDeployer(cfg *config.Config, logger *logrus.Logger) *SSHRemoteDeployer {
	return &SSHRemoteDeployer{
		config: cfg,
		logger: logger,
		runner: NewDefaultSSHCommandRunner(&cfg.RemoteDeployment, logger),
	}
}

// NewSSHRemoteDeployerWithRunner creates a new SSHRemoteDeployer with a custom SSHCommandRunner.
func NewSSHRemoteDeployerWithRunner(cfg *config.Config, logger *logrus.Logger, runner SSHCommandRunner) *SSHRemoteDeployer {
	return &SSHRemoteDeployer{
		config: cfg,
		logger: logger,
		runner: runner,
	}
}

// Deploy deploys a single service to its remote host.
func (d *SSHRemoteDeployer) Deploy(ctx context.Context, serviceName string, endpoint *config.ServiceEndpoint) error {
	if !endpoint.Remote {
		return fmt.Errorf("service %s is not configured as remote", serviceName)
	}

	// Find which host this service belongs to
	host, err := d.findHostForService(serviceName)
	if err != nil {
		return fmt.Errorf("no deployment host found for service %s: %w", serviceName, err)
	}

	d.logger.WithFields(logrus.Fields{
		"service": serviceName,
		"host":    host.SSHHost,
	}).Info("Deploying service to remote host")

	// Step 1: Ensure remote directory exists
	if err := d.ensureRemoteDirectory(host); err != nil {
		return fmt.Errorf("failed to ensure remote directory: %w", err)
	}

	// Step 2: Copy compose files and configurations
	if err := d.copyFilesToRemote(host); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	// Step 3: Deploy the specific service using docker compose
	if err := d.deployService(host, serviceName); err != nil {
		return fmt.Errorf("failed to deploy service: %w", err)
	}

	d.logger.WithField("service", serviceName).Info("Service deployed successfully")
	return nil
}

// DeployAll deploys all remote-enabled services to their respective hosts.
func (d *SSHRemoteDeployer) DeployAll(ctx context.Context) error {
	endpoints := d.config.Services.AllEndpoints()
	for name, ep := range endpoints {
		if !ep.Enabled || !ep.Remote {
			continue
		}
		if err := d.Deploy(ctx, name, &ep); err != nil {
			d.logger.WithFields(logrus.Fields{
				"service": name,
				"error":   err,
			}).Error("Failed to deploy remote service")
			// Continue with other services
		}
	}
	return nil
}

// HealthCheckRemote performs health checks on all remote services.
func (d *SSHRemoteDeployer) HealthCheckRemote(ctx context.Context) error {
	// Use existing health checker
	hc := NewServiceHealthChecker(d.logger)
	endpoints := d.config.Services.AllEndpoints()

	for name, ep := range endpoints {
		if !ep.Enabled || !ep.Remote {
			continue
		}
		d.logger.WithField("service", name).Debug("Health checking remote service")
		if err := hc.Check(name, ep); err != nil {
			d.logger.WithFields(logrus.Fields{
				"service": name,
				"error":   err,
			}).Warn("Remote service health check failed")
		}
	}
	return nil
}

// findHostForService finds the deployment host for a given service.
func (d *SSHRemoteDeployer) findHostForService(serviceName string) (*config.RemoteDeploymentHost, error) {
	rd := d.config.RemoteDeployment
	for _, host := range rd.Hosts {
		for _, svc := range host.Services {
			if svc == serviceName {
				return &host, nil
			}
		}
	}
	// If no host mapping found, fall back to default host if only one host exists
	if len(rd.Hosts) == 1 {
		for _, host := range rd.Hosts {
			return &host, nil
		}
	}
	return nil, fmt.Errorf("no deployment host configured for service %s", serviceName)
}

// ensureRemoteDirectory creates the remote directory if it doesn't exist.
func (d *SSHRemoteDeployer) ensureRemoteDirectory(host *config.RemoteDeploymentHost) error {
	remoteDir := d.config.RemoteDeployment.DefaultRemoteDir
	if remoteDir == "" {
		remoteDir = "/opt/helixagent"
	}
	cmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	_, err := d.runner.RunSSHCommand(host, cmd)
	return err
}

// copyFilesToRemote copies necessary files (compose, configs) to remote host.
func (d *SSHRemoteDeployer) copyFilesToRemote(host *config.RemoteDeploymentHost) error {
	remoteDir := d.config.RemoteDeployment.DefaultRemoteDir
	if remoteDir == "" {
		remoteDir = "/opt/helixagent"
	}

	// Copy docker-compose.yml
	localCompose := "docker-compose.yml"
	if _, err := os.Stat(localCompose); err == nil {
		if err := d.runner.SCPFile(localCompose, host, remoteDir+"/docker-compose.yml"); err != nil {
			return fmt.Errorf("failed to copy compose file: %w", err)
		}
	}

	// Copy configs directory
	if _, err := os.Stat("configs"); err == nil {
		if err := d.runner.SCPDir("configs", host, remoteDir+"/configs"); err != nil {
			d.logger.WithError(err).Warn("Failed to copy configs directory")
		}
	}

	// Copy scripts directory (contains init-db.sql)
	if _, err := os.Stat("scripts"); err == nil {
		if err := d.runner.SCPDir("scripts", host, remoteDir+"/scripts"); err != nil {
			d.logger.WithError(err).Warn("Failed to copy scripts directory")
		}
	}

	return nil
}

// deployService runs docker compose up -d for the specific service on remote host.
func (d *SSHRemoteDeployer) deployService(host *config.RemoteDeploymentHost, serviceName string) error {
	remoteDir := d.config.RemoteDeployment.DefaultRemoteDir
	if remoteDir == "" {
		remoteDir = "/opt/helixagent"
	}
	cmd := fmt.Sprintf("cd %s && docker compose up -d %s", remoteDir, serviceName)
	output, err := d.runner.RunSSHCommand(host, cmd)
	if err != nil {
		d.logger.WithFields(logrus.Fields{
			"service": serviceName,
			"output":  output,
		}).Error("Docker compose command failed")
		return fmt.Errorf("docker compose failed: %w", err)
	}
	return nil
}

