// Package containers provides an adapter layer between HelixAgent's
// container management and the extracted digital.vasic.containers
// module.
//
// This adapter centralizes all container operations (runtime
// detection, compose up/down, health checking, remote distribution)
// through the Containers module interfaces. All direct exec.Command
// calls to docker/podman should be replaced with adapter methods.
package containers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.containers/pkg/boot"
	"digital.vasic.containers/pkg/compose"
	"digital.vasic.containers/pkg/distribution"
	"digital.vasic.containers/pkg/endpoint"
	"digital.vasic.containers/pkg/envconfig"
	"digital.vasic.containers/pkg/health"
	"digital.vasic.containers/pkg/logging"
	"digital.vasic.containers/pkg/network"
	"digital.vasic.containers/pkg/orchestrator"
	"digital.vasic.containers/pkg/remote"
	"digital.vasic.containers/pkg/runtime"
	"digital.vasic.containers/pkg/scheduler"
	"digital.vasic.containers/pkg/volume"

	"dev.helix.agent/internal/config"
)

// Adapter bridges HelixAgent's container management with the
// Containers module.
type Adapter struct {
	runtime       runtime.ContainerRuntime
	orchestrator  compose.ComposeOrchestrator
	healthChecker health.HealthChecker
	distributor   distribution.Distributor
	hostManager   remote.HostManager
	executor      remote.RemoteExecutor
	tunnelManager network.TunnelManager
	volumeManager volume.VolumeManager
	logger        logging.Logger
	projectDir    string
	httpClient    *http.Client
}

// Option configures the Adapter.
type Option func(*Adapter)

// WithRuntime sets the container runtime.
func WithRuntime(r runtime.ContainerRuntime) Option {
	return func(a *Adapter) { a.runtime = r }
}

// WithOrchestrator sets the compose orchestrator.
func WithOrchestrator(o compose.ComposeOrchestrator) Option {
	return func(a *Adapter) { a.orchestrator = o }
}

// WithHealthChecker sets the health checker.
func WithHealthChecker(hc health.HealthChecker) Option {
	return func(a *Adapter) { a.healthChecker = hc }
}

// WithDistributor sets the container distributor.
func WithDistributor(d distribution.Distributor) Option {
	return func(a *Adapter) { a.distributor = d }
}

// WithHostManager sets the remote host manager.
func WithHostManager(hm remote.HostManager) Option {
	return func(a *Adapter) { a.hostManager = hm }
}

// WithLogger sets the logger.
func WithLogger(l logging.Logger) Option {
	return func(a *Adapter) { a.logger = l }
}

// WithProjectDir sets the project root directory.
func WithProjectDir(dir string) Option {
	return func(a *Adapter) { a.projectDir = dir }
}

// NewAdapter creates a container Adapter with the given options.
func NewAdapter(opts ...Option) (*Adapter, error) {
	a := &Adapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(a)
	}
	if a.logger == nil {
		a.logger = logging.NopLogger{}
	}
	if a.projectDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("detect project dir: %w", err)
		}
		a.projectDir = dir
	}
	return a, nil
}

// NewAdapterFromConfig creates an Adapter auto-configured from
// HelixAgent's config and environment. It detects the local
// container runtime, sets up the compose orchestrator, and
// optionally initializes remote distribution if env vars are set.
func NewAdapterFromConfig(cfg *config.Config) (*Adapter, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("detect project dir: %w", err)
	}

	logger := &logrusAdapter{}

	a := &Adapter{
		logger:     logger,
		projectDir: projectDir,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Auto-detect local runtime.
	rt, err := runtime.AutoDetect(context.Background())
	if err != nil {
		logger.Warn(
			"container runtime not available: %v", err,
		)
	} else {
		a.runtime = rt
		orch, orchErr := compose.NewDefaultOrchestrator(
			projectDir, logger,
		)
		if orchErr != nil {
			logger.Warn(
				"compose orchestrator not available: %v",
				orchErr,
			)
		} else {
			a.orchestrator = orch
		}
	}

	// Set up health checker.
	a.healthChecker = health.NewDefaultChecker()

	// Load Containers/.env as the single source of truth for
	// remote distribution config. Try project-relative path first,
	// then the Containers submodule.
	for _, envPath := range []string{
		filepath.Join(projectDir, "Containers", ".env"),
	} {
		if _, statErr := os.Stat(envPath); statErr == nil {
			if _, loadErr := envconfig.LoadFromFile(
				envPath,
			); loadErr != nil {
				logger.Warn(
					"load %s: %v", envPath, loadErr,
				)
			} else {
				logger.Info(
					"loaded remote config from %s", envPath,
				)
			}
			break
		}
	}

	// Check for remote distribution configuration.
	distCfg := envconfig.LoadFromEnv()
	if distCfg.Enabled {
		if err := a.setupDistribution(distCfg); err != nil {
			logger.Warn(
				"remote distribution setup failed: %v", err,
			)
		}
	}

	return a, nil
}

// setupDistribution configures remote distribution from env config.
func (a *Adapter) setupDistribution(
	cfg *envconfig.DistributionConfig,
) error {
	hosts := cfg.ToRemoteHosts()
	if len(hosts) == 0 {
		return fmt.Errorf("no remote hosts configured")
	}

	// Build SSH executor options from config.
	var sshOpts []remote.Option
	if cfg.ConnectTimeout > 0 {
		sshOpts = append(sshOpts, remote.WithConnectTimeout(
			time.Duration(cfg.ConnectTimeout)*time.Second,
		))
	}
	if cfg.CommandTimeout > 0 {
		sshOpts = append(sshOpts, remote.WithCommandTimeout(
			time.Duration(cfg.CommandTimeout)*time.Second,
		))
	}
	sshOpts = append(sshOpts, remote.WithControlMaster(
		cfg.ControlMasterEnabled,
	))
	if cfg.ControlPersist > 0 {
		sshOpts = append(sshOpts, remote.WithControlPersist(
			time.Duration(cfg.ControlPersist)*time.Second,
		))
	}
	if cfg.MaxConnections > 0 {
		sshOpts = append(sshOpts, remote.WithMaxConnections(
			cfg.MaxConnections,
		))
	}

	executor, execErr := remote.NewSSHExecutor(
		a.logger, sshOpts...,
	)
	if execErr != nil {
		return fmt.Errorf("create SSH executor: %w", execErr)
	}

	// Auto-bootstrap key auth on hosts that need it.
	ctx := context.Background()
	for i := range hosts {
		if executor.NeedsBootstrap(ctx, hosts[i]) {
			a.logger.Info(
				"bootstrapping SSH key auth on %s",
				hosts[i].Name,
			)
			if err := executor.BootstrapKeyAuth(
				ctx, hosts[i],
			); err != nil {
				a.logger.Warn(
					"bootstrap %s failed: %v",
					hosts[i].Name, err,
				)
			}
		}
	}

	hm := remote.NewHostManager(executor, a.logger)

	for _, h := range hosts {
		if err := hm.AddHost(h); err != nil {
			a.logger.Warn("add host %s: %v", h.Name, err)
		}
	}

	strategy := scheduler.StrategyResourceAware
	switch cfg.Scheduler {
	case "round_robin":
		strategy = scheduler.StrategyRoundRobin
	case "affinity":
		strategy = scheduler.StrategyAffinity
	case "spread":
		strategy = scheduler.StrategySpread
	case "bin_pack":
		strategy = scheduler.StrategyBinPack
	}

	sched := scheduler.NewScheduler(hm, a.logger,
		scheduler.WithStrategy(strategy),
	)

	var tm network.TunnelManager
	var vm volume.VolumeManager

	if cfg.PortRangeStart > 0 && cfg.PortRangeEnd > 0 {
		tm = network.NewTunnelManager(hm, a.logger,
			network.WithPortRange(
				cfg.PortRangeStart, cfg.PortRangeEnd,
			),
		)
	} else {
		tm = network.NewTunnelManager(hm, a.logger)
	}

	vm = volume.NewVolumeManager(hm, executor, a.logger)

	a.executor = executor
	a.hostManager = hm
	a.tunnelManager = tm
	a.volumeManager = vm
	a.distributor = distribution.NewDistributor(
		distribution.WithScheduler(sched),
		distribution.WithHostManager(hm),
		distribution.WithExecutor(executor),
		distribution.WithTunnelManager(tm),
		distribution.WithVolumeManager(vm),
		distribution.WithLogger(a.logger),
	)

	a.logger.Info(
		"remote distribution enabled with %d hosts",
		len(hosts),
	)
	return nil
}

// DetectRuntime returns the name of the detected container runtime
// (e.g., "docker" or "podman"). If no runtime is available, returns
// an error.
func (a *Adapter) DetectRuntime(
	ctx context.Context,
) (string, error) {
	if a.runtime != nil {
		return a.runtime.Name(), nil
	}
	rt, err := runtime.AutoDetect(ctx)
	if err != nil {
		return "", fmt.Errorf(
			"no container runtime found: %w", err,
		)
	}
	a.runtime = rt
	orch, orchErr := compose.NewDefaultOrchestrator(
		a.projectDir, a.logger,
	)
	if orchErr == nil {
		a.orchestrator = orch
	}
	return rt.Name(), nil
}

// RuntimeAvailable returns true if a container runtime is detected.
func (a *Adapter) RuntimeAvailable(ctx context.Context) bool {
	if a.runtime != nil {
		return a.runtime.IsAvailable(ctx)
	}
	rt, err := runtime.AutoDetect(ctx)
	if err != nil {
		return false
	}
	a.runtime = rt
	return true
}

// ComposeUp starts services from a compose file with the given
// profile.
func (a *Adapter) ComposeUp(
	ctx context.Context, composeFile, profile string,
) error {
	if a.orchestrator == nil {
		return fmt.Errorf("compose orchestrator not available")
	}

	absFile := composeFile
	if !filepath.IsAbs(composeFile) {
		absFile = filepath.Join(a.projectDir, composeFile)
	}

	project := compose.ComposeProject{
		File:    absFile,
		Profile: profile,
	}

	a.logger.Info("compose up: %s (profile: %s)",
		composeFile, profile,
	)
	return a.orchestrator.Up(ctx, project)
}

// ComposeDown stops services from a compose file.
func (a *Adapter) ComposeDown(
	ctx context.Context, composeFile, profile string,
) error {
	if a.orchestrator == nil {
		return fmt.Errorf("compose orchestrator not available")
	}

	absFile := composeFile
	if !filepath.IsAbs(composeFile) {
		absFile = filepath.Join(a.projectDir, composeFile)
	}

	project := compose.ComposeProject{
		File:    absFile,
		Profile: profile,
	}

	a.logger.Info("compose down: %s", composeFile)
	return a.orchestrator.Down(ctx, project)
}

// ComposeStatus returns the status of services from a compose file.
func (a *Adapter) ComposeStatus(
	ctx context.Context, composeFile string,
) ([]compose.ServiceStatus, error) {
	if a.orchestrator == nil {
		return nil, fmt.Errorf(
			"compose orchestrator not available",
		)
	}

	absFile := composeFile
	if !filepath.IsAbs(composeFile) {
		absFile = filepath.Join(a.projectDir, composeFile)
	}

	project := compose.ComposeProject{File: absFile}
	return a.orchestrator.Status(ctx, project)
}

// HealthCheck checks the health of a service target.
func (a *Adapter) HealthCheck(
	ctx context.Context,
	name, host, port, healthPath, healthType string,
	timeout time.Duration,
) (*health.HealthResult, error) {
	if a.healthChecker == nil {
		return nil, fmt.Errorf("health checker not available")
	}

	target := health.HealthTarget{
		Name:    name,
		Host:    host,
		Port:    port,
		Path:    healthPath,
		Type:    health.HealthType(healthType),
		Timeout: timeout,
	}

	result := a.healthChecker.Check(ctx, target)
	return result, nil
}

// HealthCheckHTTP performs an HTTP health check on the given URL.
func (a *Adapter) HealthCheckHTTP(url string) error {
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("cannot connect: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"health check failed with status: %d",
			resp.StatusCode,
		)
	}
	return nil
}

// HealthCheckTCP performs a TCP health check.
func (a *Adapter) HealthCheckTCP(
	host string, port int,
) bool {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// ToEndpoint converts HelixAgent service config fields to a
// Containers module ServiceEndpoint.
func (a *Adapter) ToEndpoint(
	name, host, port, healthPath, healthType,
	composeFile, serviceName, profile string,
	enabled, required, isRemote bool,
) endpoint.ServiceEndpoint {
	return endpoint.ServiceEndpoint{
		Host:        host,
		Port:        port,
		HealthPath:  healthPath,
		HealthType:  healthType,
		ComposeFile: composeFile,
		ServiceName: serviceName,
		Profile:     profile,
		Enabled:     enabled,
		Required:    required,
		Remote:      isRemote,
	}
}

// BootAll boots all provided endpoints using the Containers module's
// BootManager. It creates a BootManager, registers endpoints, starts
// compose groups, and runs health checks.
func (a *Adapter) BootAll(
	ctx context.Context,
	endpoints map[string]endpoint.ServiceEndpoint,
) (*boot.BootSummary, error) {
	opts := []boot.BootManagerOption{
		boot.WithLogger(a.logger),
	}
	if a.runtime != nil {
		opts = append(opts, boot.WithRuntime(a.runtime))
	}
	if a.orchestrator != nil {
		opts = append(opts, boot.WithOrchestrator(a.orchestrator))
	}
	if a.healthChecker != nil {
		opts = append(opts, boot.WithHealthChecker(a.healthChecker))
	}
	if a.projectDir != "" {
		opts = append(opts, boot.WithProjectDir(a.projectDir))
	}
	if a.distributor != nil {
		if d, ok := a.distributor.(*distribution.DefaultDistributor); ok {
			opts = append(opts, boot.WithDistributor(d))
		}
	}
	if a.hostManager != nil {
		opts = append(opts, boot.WithHostManager(a.hostManager))
	}

	bm := boot.NewBootManager(endpoints, opts...)
	return bm.BootAll(ctx)
}

// ToHealthTarget converts service configuration fields into a
// Containers module HealthTarget.
func (a *Adapter) ToHealthTarget(
	name, host, port, healthPath, healthType string,
	timeout time.Duration, required bool,
) health.HealthTarget {
	return health.HealthTarget{
		Name:     name,
		Host:     host,
		Port:     port,
		Type:     health.HealthType(healthType),
		Path:     healthPath,
		Timeout:  timeout,
		Required: required,
	}
}

// ListContainers returns the status of services from the compose
// file at the given path. Uses the adapter's compose orchestrator.
func (a *Adapter) ListContainers(
	ctx context.Context, composeFile string,
) ([]compose.ServiceStatus, error) {
	return a.ComposeStatus(ctx, composeFile)
}

// HealthCheckAll performs health checks on a list of targets and
// returns errors keyed by target name.
func (a *Adapter) HealthCheckAll(
	ctx context.Context, targets []health.HealthTarget,
) map[string]error {
	errors := make(map[string]error)
	if a.healthChecker == nil {
		return errors
	}
	results := a.healthChecker.CheckAll(ctx, targets)
	for i, result := range results {
		if !result.Healthy {
			errors[targets[i].Name] = fmt.Errorf(
				"health check failed: %s", result.Error,
			)
		}
	}
	return errors
}

// StatusAll returns the status of all running containers.
func (a *Adapter) StatusAll(
	ctx context.Context,
) (map[string]string, error) {
	status := make(map[string]string)
	if a.runtime == nil {
		return status, fmt.Errorf("no container runtime available")
	}
	containers, err := a.runtime.List(
		ctx, runtime.ListFilter{},
	)
	if err != nil {
		return status, fmt.Errorf("list containers: %w", err)
	}
	for _, c := range containers {
		status[c.Name] = string(c.State)
	}
	return status, nil
}

// Distribute distributes containers across local and remote hosts
// using the configured scheduler and remote executor.
func (a *Adapter) Distribute(
	ctx context.Context,
	reqs []scheduler.ContainerRequirements,
) (*distribution.DistributionSummary, error) {
	if a.distributor == nil {
		return nil, fmt.Errorf("distributor not configured")
	}
	return a.distributor.Distribute(ctx, reqs)
}

// Undistribute stops all distributed containers.
func (a *Adapter) Undistribute(ctx context.Context) error {
	if a.distributor == nil {
		return nil
	}
	return a.distributor.Undistribute(ctx)
}

// DistributionStatus returns the current state of all distributed
// containers.
func (a *Adapter) DistributionStatus(
	ctx context.Context,
) []distribution.DistributedContainer {
	if a.distributor == nil {
		return nil
	}
	return a.distributor.Status(ctx)
}

// RemoteEnabled returns true if remote distribution is configured.
func (a *Adapter) RemoteEnabled() bool {
	return a.distributor != nil && a.hostManager != nil
}

// RemoteComposeUp deploys a compose file to the first available
// remote host and starts its services. It copies the compose file
// and any supporting files in the same directory, then runs
// `docker compose up -d` on the remote host.
func (a *Adapter) RemoteComposeUp(
	ctx context.Context, composeFile, profile string,
) error {
	if a.hostManager == nil || a.executor == nil {
		return fmt.Errorf(
			"remote distribution not configured",
		)
	}

	hosts := a.hostManager.ListHosts()
	if len(hosts) == 0 {
		return fmt.Errorf("no remote hosts available")
	}

	host := hosts[0]

	absFile := composeFile
	if !filepath.IsAbs(composeFile) {
		absFile = filepath.Join(a.projectDir, composeFile)
	}

	if _, err := os.Stat(absFile); err != nil {
		return fmt.Errorf(
			"compose file not found: %s", absFile,
		)
	}

	// Use user's home directory for HelixAgent deployments (no sudo required)
	// This avoids permission issues with /opt/helixagent
	remoteDir := fmt.Sprintf("/home/%s/helixagent", host.User)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if _, err := a.executor.Execute(
		ctx, host, mkdirCmd,
	); err != nil {
		return fmt.Errorf(
			"create remote dir on %s: %w",
			host.Name, err,
		)
	}

	// Copy the compose file's directory to the remote host.
	localDir := filepath.Dir(absFile)
	remoteDest := remoteDir + "/" + filepath.Base(localDir)
	if err := a.executor.CopyDir(
		ctx, host, localDir, remoteDest,
	); err != nil {
		return fmt.Errorf(
			"copy compose dir to %s: %w", host.Name, err,
		)
	}

	// Use RemoteComposeOrchestrator to start services.
	remoteOrch := remote.NewRemoteComposeOrchestrator(
		host, a.executor, a.logger,
	)
	remoteFile := remoteDest + "/" + filepath.Base(absFile)
	project := compose.ComposeProject{
		File:    remoteFile,
		Profile: profile,
	}

	a.logger.Info(
		"remote compose up on %s: %s (profile: %s)",
		host.Name, remoteFile, profile,
	)
	return remoteOrch.Up(ctx, project)
}

// ListHosts returns all registered remote hosts.
func (a *Adapter) ListHosts() []remote.RemoteHost {
	if a.hostManager == nil {
		return nil
	}
	return a.hostManager.ListHosts()
}

// ProbeHost returns resource info for a specific remote host.
func (a *Adapter) ProbeHost(
	ctx context.Context, name string,
) (*remote.HostResources, error) {
	if a.hostManager == nil {
		return nil, fmt.Errorf("host manager not configured")
	}
	return a.hostManager.ProbeHost(ctx, name)
}

// Shutdown gracefully shuts down all container operations:
// closes tunnels, unmounts volumes, stops distributed containers.
func (a *Adapter) Shutdown(ctx context.Context) error {
	var errs []string

	if a.distributor != nil {
		if err := a.distributor.Undistribute(ctx); err != nil {
			errs = append(errs, fmt.Sprintf(
				"undistribute: %v", err,
			))
		}
	}

	if a.tunnelManager != nil {
		if err := a.tunnelManager.CloseAll(); err != nil {
			errs = append(errs, fmt.Sprintf(
				"close tunnels: %v", err,
			))
		}
	}

	if a.volumeManager != nil {
		if err := a.volumeManager.UnmountAll(ctx); err != nil {
			errs = append(errs, fmt.Sprintf(
				"unmount volumes: %v", err,
			))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"shutdown errors: %s",
			strings.Join(errs, "; "),
		)
	}
	return nil
}

// Runtime returns the underlying container runtime. May be nil.
func (a *Adapter) Runtime() runtime.ContainerRuntime {
	return a.runtime
}

// Orchestrator returns the underlying compose orchestrator. May
// be nil.
func (a *Adapter) Orchestrator() compose.ComposeOrchestrator {
	return a.orchestrator
}

type serviceOrchAdapter struct {
	orch compose.ComposeOrchestrator
}

func (a *serviceOrchAdapter) Up(ctx context.Context, project compose.ComposeProject) error {
	return a.orch.Up(ctx, project)
}

func (a *serviceOrchAdapter) Down(ctx context.Context, project compose.ComposeProject) error {
	return a.orch.Down(ctx, project)
}

type remoteExecAdapter struct {
	exec remote.RemoteExecutor
}

func (a *remoteExecAdapter) Execute(ctx context.Context, host remote.RemoteHost, cmd string) (*remote.CommandResult, error) {
	return a.exec.Execute(ctx, host, cmd)
}

func (a *remoteExecAdapter) CopyDir(ctx context.Context, host remote.RemoteHost, src, dst string) error {
	return a.exec.CopyDir(ctx, host, src, dst)
}

type hostMgrAdapter struct {
	mgr remote.HostManager
}

func (a *hostMgrAdapter) ListHosts() []remote.RemoteHost {
	return a.mgr.ListHosts()
}

func (a *Adapter) NewServiceOrchestrator() *orchestrator.DefaultOrchestrator {
	opts := []orchestrator.Option{
		orchestrator.WithLogger(a.logger),
		orchestrator.WithProjectDir(a.projectDir),
	}
	if a.orchestrator != nil {
		opts = append(opts, orchestrator.WithLocalOrchestrator(&serviceOrchAdapter{orch: a.orchestrator}))
	}
	if a.executor != nil {
		opts = append(opts, orchestrator.WithRemoteExecutor(&remoteExecAdapter{exec: a.executor}))
	}
	if a.hostManager != nil {
		opts = append(opts, orchestrator.WithHostManager(&hostMgrAdapter{mgr: a.hostManager}))
	}
	if a.healthChecker != nil {
		opts = append(opts, orchestrator.WithHealthChecker(a.healthChecker))
	}
	return orchestrator.New(opts...)
}

// logrusAdapter bridges the logging.Logger interface with
// HelixAgent's typical logrus usage.
type logrusAdapter struct{}

func (l *logrusAdapter) Debug(msg string, args ...any) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

func (l *logrusAdapter) Info(msg string, args ...any) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

func (l *logrusAdapter) Warn(msg string, args ...any) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

func (l *logrusAdapter) Error(msg string, args ...any) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}
