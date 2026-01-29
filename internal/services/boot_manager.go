package services

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// BootResult records the outcome of booting a single service.
type BootResult struct {
	Name     string
	Status   string // "started", "already_running", "remote", "failed", "skipped"
	Duration time.Duration
	Error    error
}

// BootManager handles starting, health-checking, and stopping all configured services.
type BootManager struct {
	Config        *config.ServicesConfig
	Logger        *logrus.Logger
	Results       map[string]*BootResult
	HealthChecker *ServiceHealthChecker
	ProjectDir    string
}

// NewBootManager creates a new BootManager.
func NewBootManager(cfg *config.ServicesConfig, logger *logrus.Logger) *BootManager {
	return &BootManager{
		Config:        cfg,
		Logger:        logger,
		Results:       make(map[string]*BootResult),
		HealthChecker: NewServiceHealthChecker(logger),
	}
}

// BootAll starts all enabled local services and health-checks all enabled services.
// Remote services (Remote: true) are only health-checked, not started via compose.
// Required services that fail health check will cause an error return.
func (bm *BootManager) BootAll() error {
	bm.Logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	bm.Logger.Info("║              UNIFIED SERVICE BOOT MANAGER                        ║")
	bm.Logger.Info("╚══════════════════════════════════════════════════════════════════╝")

	endpoints := bm.Config.AllEndpoints()

	// Group local services by compose file for batch startup
	composeGroups := make(map[string][]string) // compose_file -> []service_name
	for name, ep := range endpoints {
		if !ep.Enabled {
			bm.Results[name] = &BootResult{Name: name, Status: "skipped"}
			bm.Logger.WithField("service", name).Debug("Service disabled, skipping")
			continue
		}
		if ep.Remote {
			bm.Results[name] = &BootResult{Name: name, Status: "remote"}
			bm.Logger.WithField("service", name).Info("Service configured as remote, skipping compose start")
			continue
		}
		if ep.ComposeFile != "" && ep.ServiceName != "" {
			key := ep.ComposeFile
			if ep.Profile != "" {
				key = ep.ComposeFile + "|" + ep.Profile
			}
			composeGroups[key] = append(composeGroups[key], ep.ServiceName)
		}
	}

	// Start local services grouped by compose file
	for key, services := range composeGroups {
		parts := strings.SplitN(key, "|", 2)
		composeFile := parts[0]
		profile := ""
		if len(parts) > 1 {
			profile = parts[1]
		}

		start := time.Now()
		err := bm.startComposeServices(composeFile, profile, services)
		duration := time.Since(start)

		if err != nil {
			bm.Logger.WithFields(logrus.Fields{
				"compose_file": composeFile,
				"services":     strings.Join(services, ", "),
				"error":        err,
			}).Warn("Failed to start compose services")
			for _, svc := range services {
				bm.Results[svc] = &BootResult{Name: svc, Status: "failed", Duration: duration, Error: err}
			}
		} else {
			for _, svc := range services {
				bm.Results[svc] = &BootResult{Name: svc, Status: "started", Duration: duration}
			}
		}
	}

	// Health check all enabled services
	bm.Logger.Info("Running health checks on all enabled services...")
	var requiredFailures []string

	for name, ep := range endpoints {
		if !ep.Enabled {
			continue
		}

		start := time.Now()
		err := bm.HealthChecker.CheckWithRetry(name, ep)
		duration := time.Since(start)

		if err != nil {
			if ep.Required {
				requiredFailures = append(requiredFailures, fmt.Sprintf("%s: %v", name, err))
				bm.Logger.WithFields(logrus.Fields{
					"service":  name,
					"duration": duration,
					"error":    err,
				}).Error("REQUIRED service health check FAILED")
			} else {
				bm.Logger.WithFields(logrus.Fields{
					"service":  name,
					"duration": duration,
					"error":    err,
				}).Warn("Optional service health check failed")
			}
			if result, ok := bm.Results[name]; ok {
				result.Error = err
				if result.Status != "remote" {
					result.Status = "failed"
				}
			} else {
				bm.Results[name] = &BootResult{Name: name, Status: "failed", Duration: duration, Error: err}
			}
		} else {
			bm.Logger.WithFields(logrus.Fields{
				"service":  name,
				"duration": duration,
			}).Info("Service health check passed")
			if result, ok := bm.Results[name]; ok {
				result.Duration = duration
				if result.Status == "" {
					result.Status = "already_running"
				}
			}
		}
	}

	// Log summary
	bm.logSummary()

	if len(requiredFailures) > 0 {
		errMsg := fmt.Sprintf("BOOT BLOCKED: %d required service(s) failed health check:\n", len(requiredFailures))
		for i, f := range requiredFailures {
			errMsg += fmt.Sprintf("  %d. %s\n", i+1, f)
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// HealthCheckAll checks all enabled services and returns a map of errors.
func (bm *BootManager) HealthCheckAll() map[string]error {
	results := make(map[string]error)
	for name, ep := range bm.Config.AllEndpoints() {
		if !ep.Enabled {
			continue
		}
		results[name] = bm.HealthChecker.Check(name, ep)
	}
	return results
}

// ShutdownAll stops all local services that were started by this boot manager.
func (bm *BootManager) ShutdownAll() error {
	bm.Logger.Info("Shutting down all managed services...")

	endpoints := bm.Config.AllEndpoints()

	// Group by compose file
	composeGroups := make(map[string][]string)
	for name, ep := range endpoints {
		if !ep.Enabled || ep.Remote {
			continue
		}
		result, ok := bm.Results[name]
		if !ok || (result.Status != "started" && result.Status != "already_running") {
			continue
		}
		if ep.ComposeFile != "" && ep.ServiceName != "" {
			key := ep.ComposeFile
			if ep.Profile != "" {
				key = ep.ComposeFile + "|" + ep.Profile
			}
			composeGroups[key] = append(composeGroups[key], ep.ServiceName)
		}
	}

	var errors []string
	for key, services := range composeGroups {
		parts := strings.SplitN(key, "|", 2)
		composeFile := parts[0]
		profile := ""
		if len(parts) > 1 {
			profile = parts[1]
		}

		if err := bm.stopComposeServices(composeFile, profile, services); err != nil {
			errors = append(errors, fmt.Sprintf("compose %s: %v", composeFile, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %s", strings.Join(errors, "; "))
	}

	bm.Logger.Info("All managed services stopped")
	return nil
}

func (bm *BootManager) startComposeServices(composeFile, profile string, services []string) error {
	composeCmd, composeArgs := detectComposeCmd()

	var cmdArgs []string
	cmdArgs = append(cmdArgs, composeArgs...)
	if composeFile != "" {
		cmdArgs = append(cmdArgs, "-f", composeFile)
	}
	if profile != "" {
		cmdArgs = append(cmdArgs, "--profile", profile)
	}
	cmdArgs = append(cmdArgs, "up", "-d")
	cmdArgs = append(cmdArgs, services...)

	bm.Logger.WithFields(logrus.Fields{
		"command":  composeCmd,
		"services": strings.Join(services, ", "),
	}).Info("Starting compose services")

	cmd := exec.Command(composeCmd, cmdArgs...)
	if bm.ProjectDir != "" {
		cmd.Dir = bm.ProjectDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose up failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (bm *BootManager) stopComposeServices(composeFile, profile string, services []string) error {
	composeCmd, composeArgs := detectComposeCmd()

	var cmdArgs []string
	cmdArgs = append(cmdArgs, composeArgs...)
	if composeFile != "" {
		cmdArgs = append(cmdArgs, "-f", composeFile)
	}
	if profile != "" {
		cmdArgs = append(cmdArgs, "--profile", profile)
	}
	cmdArgs = append(cmdArgs, "stop")
	cmdArgs = append(cmdArgs, services...)

	bm.Logger.WithFields(logrus.Fields{
		"command":  composeCmd,
		"services": strings.Join(services, ", "),
	}).Info("Stopping compose services")

	cmd := exec.Command(composeCmd, cmdArgs...)
	if bm.ProjectDir != "" {
		cmd.Dir = bm.ProjectDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose stop failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (bm *BootManager) logSummary() {
	started, remote, failed, skipped := 0, 0, 0, 0
	for _, r := range bm.Results {
		switch r.Status {
		case "started", "already_running":
			started++
		case "remote":
			remote++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}

	bm.Logger.WithFields(logrus.Fields{
		"started": started,
		"remote":  remote,
		"failed":  failed,
		"skipped": skipped,
		"total":   len(bm.Results),
	}).Info("Service boot summary")
}

// detectComposeCmd returns the compose command and any prefix args.
func detectComposeCmd() (string, []string) {
	// Try "docker compose" (v2) first
	if path, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command(path, "compose", "version")
		if err := cmd.Run(); err == nil {
			return path, []string{"compose"}
		}
	}
	// Fallback to docker-compose (v1)
	if path, err := exec.LookPath("docker-compose"); err == nil {
		return path, nil
	}
	// Try podman-compose
	if path, err := exec.LookPath("podman-compose"); err == nil {
		return path, nil
	}
	// Try podman compose
	if path, err := exec.LookPath("podman"); err == nil {
		return path, []string{"compose"}
	}
	return "docker", []string{"compose"}
}
