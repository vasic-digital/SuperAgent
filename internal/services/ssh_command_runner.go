package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"digital.vasic.containers/pkg/logging"
	"digital.vasic.containers/pkg/remote"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// SSHCommandRunner defines the interface for executing SSH and SCP commands.
type SSHCommandRunner interface {
	RunSSHCommand(host *config.RemoteDeploymentHost, command string) (string, error)
	SCPFile(localPath string, host *config.RemoteDeploymentHost, remotePath string) error
	SCPDir(localDir string, host *config.RemoteDeploymentHost, remoteDir string) error
}

// defaultSSHCommandRunner implements SSHCommandRunner by delegating to the
// Containers module's remote.SSHExecutor.
type defaultSSHCommandRunner struct {
	logger   *logrus.Logger
	config   *config.RemoteDeploymentConfig
	executor remote.RemoteExecutor
}

// NewDefaultSSHCommandRunner creates a new default SSH command runner.
// Internally creates a remote.SSHExecutor from the Containers module.
func NewDefaultSSHCommandRunner(
	cfg *config.RemoteDeploymentConfig, logger *logrus.Logger,
) SSHCommandRunner {
	executor, err := remote.NewSSHExecutor(
		&logrusRemoteAdapter{logger: logger},
		remote.WithConnectTimeout(10),
		remote.WithControlMaster(false),
	)
	if err != nil {
		logger.WithError(err).Warn(
			"Failed to create SSHExecutor, SSH operations will fail",
		)
	}
	return &defaultSSHCommandRunner{
		logger:   logger,
		config:   cfg,
		executor: executor,
	}
}

// toRemoteHost converts a config.RemoteDeploymentHost to a remote.RemoteHost.
func (r *defaultSSHCommandRunner) toRemoteHost(
	host *config.RemoteDeploymentHost,
) remote.RemoteHost {
	sshKey := host.SSHKey
	if sshKey == "" {
		sshKey = r.config.SSHKey
	}
	if sshKey == "" {
		sshKey = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}

	user := ""
	address := host.SSHHost
	if strings.Contains(address, "@") {
		parts := strings.SplitN(address, "@", 2)
		user = parts[0]
		address = parts[1]
	} else if r.config.DefaultSSHUser != "" {
		user = r.config.DefaultSSHUser
	}

	return remote.RemoteHost{
		Name:    host.SSHHost,
		Address: address,
		Port:    22,
		User:    user,
		KeyPath: sshKey,
		Runtime: "docker",
	}
}

// RunSSHCommand executes a command on the remote host via the Containers
// module's SSHExecutor.
func (r *defaultSSHCommandRunner) RunSSHCommand(
	host *config.RemoteDeploymentHost, command string,
) (string, error) {
	if r.executor == nil {
		return "", fmt.Errorf("SSH executor not available")
	}

	rh := r.toRemoteHost(host)
	r.logger.WithField("command", command).Debug("Running SSH command")

	ctx := context.Background()
	result, err := r.executor.Execute(ctx, rh, command)
	if err != nil {
		return "", err
	}

	output := result.Stdout
	if result.Stderr != "" {
		output += result.Stderr
	}

	if result.ExitCode != 0 {
		return output, fmt.Errorf(
			"ssh command failed with exit code %d: %s",
			result.ExitCode, result.Stderr,
		)
	}

	return output, nil
}

// SCPFile copies a file to remote host via the Containers module's
// SSHExecutor.CopyFile.
func (r *defaultSSHCommandRunner) SCPFile(
	localPath string,
	host *config.RemoteDeploymentHost,
	remotePath string,
) error {
	if r.executor == nil {
		return fmt.Errorf("SSH executor not available")
	}

	rh := r.toRemoteHost(host)
	r.logger.WithField("file", localPath).Debug("Copying file via SCP")

	ctx := context.Background()
	if err := r.executor.CopyFile(
		ctx, rh, localPath, remotePath,
	); err != nil {
		r.logger.WithFields(logrus.Fields{
			"file":  localPath,
			"error": err,
		}).Error("SCP failed")
		return fmt.Errorf("scp failed: %w", err)
	}
	return nil
}

// SCPDir copies a directory recursively to remote host via the Containers
// module's SSHExecutor.CopyDir.
func (r *defaultSSHCommandRunner) SCPDir(
	localDir string,
	host *config.RemoteDeploymentHost,
	remoteDir string,
) error {
	if r.executor == nil {
		return fmt.Errorf("SSH executor not available")
	}

	rh := r.toRemoteHost(host)
	r.logger.WithField("dir", localDir).Debug(
		"Copying directory via SCP",
	)

	ctx := context.Background()
	if err := r.executor.CopyDir(
		ctx, rh, localDir, remoteDir,
	); err != nil {
		return fmt.Errorf("scp dir failed: %w", err)
	}
	return nil
}

// logrusRemoteAdapter adapts logrus to the logging.Logger interface.
type logrusRemoteAdapter struct {
	logger *logrus.Logger
}

func (l *logrusRemoteAdapter) Debug(msg string, args ...any) {
	l.logger.Debugf(msg, args...)
}

func (l *logrusRemoteAdapter) Info(msg string, args ...any) {
	l.logger.Infof(msg, args...)
}

func (l *logrusRemoteAdapter) Warn(msg string, args ...any) {
	l.logger.Warnf(msg, args...)
}

func (l *logrusRemoteAdapter) Error(msg string, args ...any) {
	l.logger.Errorf(msg, args...)
}

// Compile-time interface assertion.
var _ logging.Logger = (*logrusRemoteAdapter)(nil)
