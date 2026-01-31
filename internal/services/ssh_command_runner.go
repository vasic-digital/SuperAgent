package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// SSHCommandRunner defines the interface for executing SSH and SCP commands.
type SSHCommandRunner interface {
	RunSSHCommand(host *config.RemoteDeploymentHost, command string) (string, error)
	SCPFile(localPath string, host *config.RemoteDeploymentHost, remotePath string) error
	SCPDir(localDir string, host *config.RemoteDeploymentHost, remoteDir string) error
}

// defaultSSHCommandRunner implements SSHCommandRunner using system ssh and scp commands.
type defaultSSHCommandRunner struct {
	logger *logrus.Logger
	config *config.RemoteDeploymentConfig
}

// NewDefaultSSHCommandRunner creates a new default SSH command runner.
func NewDefaultSSHCommandRunner(cfg *config.RemoteDeploymentConfig, logger *logrus.Logger) SSHCommandRunner {
	return &defaultSSHCommandRunner{
		logger: logger,
		config: cfg,
	}
}

// RunSSHCommand executes a command on the remote host via SSH.
func (r *defaultSSHCommandRunner) RunSSHCommand(host *config.RemoteDeploymentHost, command string) (string, error) {
	sshKey := host.SSHKey
	if sshKey == "" {
		sshKey = r.config.SSHKey
	}
	if sshKey == "" {
		sshKey = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}

	sshHost := host.SSHHost
	if !strings.Contains(sshHost, "@") {
		// Add default SSH user if not present
		defaultUser := r.config.DefaultSSHUser
		if defaultUser != "" {
			sshHost = defaultUser + "@" + sshHost
		}
	}

	args := []string{
		"-i", sshKey,
		"-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=no",
		sshHost,
		command,
	}

	r.logger.WithField("command", command).Debug("Running SSH command")
	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// SCPFile copies a file to remote host via SCP.
func (r *defaultSSHCommandRunner) SCPFile(localPath string, host *config.RemoteDeploymentHost, remotePath string) error {
	sshKey := host.SSHKey
	if sshKey == "" {
		sshKey = r.config.SSHKey
	}
	if sshKey == "" {
		sshKey = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}

	sshHost := host.SSHHost
	if !strings.Contains(sshHost, "@") {
		defaultUser := r.config.DefaultSSHUser
		if defaultUser != "" {
			sshHost = defaultUser + "@" + sshHost
		}
	}

	args := []string{
		"-i", sshKey,
		"-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=no",
		localPath,
		sshHost + ":" + remotePath,
	}

	r.logger.WithField("file", localPath).Debug("Copying file via SCP")
	cmd := exec.Command("scp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"file":   localPath,
			"output": string(output),
		}).Error("SCP failed")
		return fmt.Errorf("scp failed: %w", err)
	}
	return nil
}

// SCPDir copies a directory recursively to remote host via SCP using tar over ssh.
func (r *defaultSSHCommandRunner) SCPDir(localDir string, host *config.RemoteDeploymentHost, remoteDir string) error {
	sshKey := host.SSHKey
	if sshKey == "" {
		sshKey = r.config.SSHKey
	}
	if sshKey == "" {
		sshKey = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}

	sshHost := host.SSHHost
	if !strings.Contains(sshHost, "@") {
		defaultUser := r.config.DefaultSSHUser
		if defaultUser != "" {
			sshHost = defaultUser + "@" + sshHost
		}
	}

	// Use tar over ssh for efficient directory copy
	localAbs, err := filepath.Abs(localDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	tarCmd := fmt.Sprintf("tar czf - -C %s .", localAbs)
	sshCmd := fmt.Sprintf("mkdir -p %s && tar xzf - -C %s", remoteDir, remoteDir)

	// Create tar pipe to ssh
	tar := exec.Command("sh", "-c", tarCmd)
	ssh := exec.Command("ssh", "-i", sshKey,
		"-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=no",
		sshHost, sshCmd)

	ssh.Stdin, _ = tar.StdoutPipe()
	ssh.Stdout = os.Stdout
	ssh.Stderr = os.Stderr

	r.logger.WithField("dir", localDir).Debug("Copying directory via tar+ssh")
	if err := ssh.Start(); err != nil {
		return fmt.Errorf("failed to start ssh: %w", err)
	}
	if err := tar.Run(); err != nil {
		return fmt.Errorf("tar failed: %w", err)
	}
	if err := ssh.Wait(); err != nil {
		return fmt.Errorf("ssh tar extraction failed: %w", err)
	}
	return nil
}
