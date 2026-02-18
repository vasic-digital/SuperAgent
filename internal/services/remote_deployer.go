package services

import (
	"context"
	"fmt"

	"dev.helix.agent/internal/adapters/containers"
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

// SSHRemoteDeployer implements RemoteDeployer using the Containers module's
// distribution system. It delegates SSH operations to the adapter's
// remote.SSHExecutor and scheduling to the adapter's distribution.Distributor.
type SSHRemoteDeployer struct {
	config           *config.Config
	logger           *logrus.Logger
	runner           SSHCommandRunner
	ContainerAdapter *containers.Adapter
}

// NewSSHRemoteDeployer creates a new SSHRemoteDeployer.
func NewSSHRemoteDeployer(
	cfg *config.Config, logger *logrus.Logger,
) *SSHRemoteDeployer {
	return &SSHRemoteDeployer{
		config: cfg,
		logger: logger,
		runner: NewDefaultSSHCommandRunner(
			&cfg.RemoteDeployment, logger,
		),
	}
}

// NewSSHRemoteDeployerWithRunner creates a new SSHRemoteDeployer with
// a custom SSHCommandRunner.
func NewSSHRemoteDeployerWithRunner(
	cfg *config.Config,
	logger *logrus.Logger,
	runner SSHCommandRunner,
) *SSHRemoteDeployer {
	return &SSHRemoteDeployer{
		config: cfg,
		logger: logger,
		runner: runner,
	}
}

// Deploy deploys a single service to its remote host using the
// Containers module's distribution system when available, with
// fallback to the SSHCommandRunner.
func (d *SSHRemoteDeployer) Deploy(
	ctx context.Context,
	serviceName string,
	endpoint *config.ServiceEndpoint,
) error {
	if !endpoint.Remote {
		return fmt.Errorf(
			"service %s is not configured as remote", serviceName,
		)
	}

	// Find which host this service belongs to
	host, err := d.findHostForService(serviceName)
	if err != nil {
		return fmt.Errorf(
			"no deployment host found for service %s: %w",
			serviceName, err,
		)
	}

	d.logger.WithFields(logrus.Fields{
		"service": serviceName,
		"host":    host.SSHHost,
	}).Info("Deploying service to remote host")

	// Step 1: Ensure remote directory exists
	if err := d.ensureRemoteDirectory(host); err != nil {
		return fmt.Errorf(
			"failed to ensure remote directory: %w", err,
		)
	}

	// Step 2: Copy compose files and configurations
	if err := d.copyFilesToRemote(host); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	// Step 3: Deploy the specific service. Use the adapter when
	// available for compose operations, fall back to the runner.
	if d.ContainerAdapter != nil {
		// Use adapter's compose up on remote host via distribution.
		composeFile := "docker-compose.yml"
		if err := d.ContainerAdapter.ComposeUp(
			ctx, composeFile, "",
		); err != nil {
			d.logger.WithError(err).Warn(
				"Adapter compose up failed, falling back to runner",
			)
		} else {
			d.logger.WithField("service", serviceName).Info(
				"Service deployed via Containers module",
			)
			return nil
		}
	}

	// Fallback: use SSHCommandRunner
	if err := d.deployService(host, serviceName); err != nil {
		return fmt.Errorf("failed to deploy service: %w", err)
	}

	d.logger.WithField("service", serviceName).Info(
		"Service deployed successfully",
	)
	return nil
}

// DeployAll deploys all remote-enabled services to their respective hosts.
func (d *SSHRemoteDeployer) DeployAll(ctx context.Context) error {
	// When the adapter has distribution configured, use it.
	if d.ContainerAdapter != nil && d.ContainerAdapter.RemoteEnabled() {
		var remoteNames []string
		endpoints := d.config.Services.AllEndpoints()
		for name, ep := range endpoints {
			if ep.Enabled && ep.Remote {
				remoteNames = append(remoteNames, name)
			}
		}
		if len(remoteNames) > 0 {
			d.logger.WithField("services", len(remoteNames)).Info(
				"Distributing remote services via Containers module",
			)
			// Use adapter's distribution.
			_, err := d.ContainerAdapter.Distribute(ctx, nil)
			if err != nil {
				d.logger.WithError(err).Warn(
					"Distribution failed, falling back to manual deploy",
				)
			} else {
				return nil
			}
		}
	}

	// Fallback: deploy each service individually.
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
		}
	}
	return nil
}

// HealthCheckRemote performs health checks on all remote services.
func (d *SSHRemoteDeployer) HealthCheckRemote(
	ctx context.Context,
) error {
	hc := NewServiceHealthChecker(d.logger)
	// Wire adapter into the health checker if available.
	if d.ContainerAdapter != nil {
		hc.ContainerAdapter = d.ContainerAdapter
	}
	endpoints := d.config.Services.AllEndpoints()

	for name, ep := range endpoints {
		if !ep.Enabled || !ep.Remote {
			continue
		}
		d.logger.WithField("service", name).Debug(
			"Health checking remote service",
		)
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
func (d *SSHRemoteDeployer) findHostForService(
	serviceName string,
) (*config.RemoteDeploymentHost, error) {
	rd := d.config.RemoteDeployment
	for _, host := range rd.Hosts {
		for _, svc := range host.Services {
			if svc == serviceName {
				return &host, nil
			}
		}
	}
	if len(rd.Hosts) == 1 {
		for _, host := range rd.Hosts {
			return &host, nil
		}
	}
	return nil, fmt.Errorf(
		"no deployment host configured for service %s", serviceName,
	)
}

// ensureRemoteDirectory creates the remote directory if needed.
func (d *SSHRemoteDeployer) ensureRemoteDirectory(
	host *config.RemoteDeploymentHost,
) error {
	remoteDir := d.config.RemoteDeployment.DefaultRemoteDir
	if remoteDir == "" {
		remoteDir = "/opt/helixagent"
	}
	cmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	_, err := d.runner.RunSSHCommand(host, cmd)
	return err
}

// copyFilesToRemote copies necessary files to remote host.
func (d *SSHRemoteDeployer) copyFilesToRemote(
	host *config.RemoteDeploymentHost,
) error {
	remoteDir := d.config.RemoteDeployment.DefaultRemoteDir
	if remoteDir == "" {
		remoteDir = "/opt/helixagent"
	}

	localCompose := "docker-compose.yml"
	if err := d.runner.SCPFile(
		localCompose, host, remoteDir+"/docker-compose.yml",
	); err != nil {
		d.logger.WithError(err).Warn(
			"Failed to copy compose file",
		)
	}

	if err := d.runner.SCPDir(
		"configs", host, remoteDir+"/configs",
	); err != nil {
		d.logger.WithError(err).Warn(
			"Failed to copy configs directory",
		)
	}

	if err := d.runner.SCPDir(
		"scripts", host, remoteDir+"/scripts",
	); err != nil {
		d.logger.WithError(err).Warn(
			"Failed to copy scripts directory",
		)
	}

	return nil
}

// deployService runs docker compose up for the specific service on remote.
func (d *SSHRemoteDeployer) deployService(
	host *config.RemoteDeploymentHost, serviceName string,
) error {
	remoteDir := d.config.RemoteDeployment.DefaultRemoteDir
	if remoteDir == "" {
		remoteDir = "/opt/helixagent"
	}
	cmd := fmt.Sprintf(
		"cd %s && docker compose up -d %s", remoteDir, serviceName,
	)
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
