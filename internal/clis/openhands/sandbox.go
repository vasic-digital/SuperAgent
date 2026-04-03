// Package openhands provides OpenHands CLI agent integration for HelixAgent.
package openhands

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Sandbox provides Docker-based secure code execution.
// Ported from OpenHands' sandbox implementation
type Sandbox struct {
	client     *client.Client
	containerID string
	image      string
	
	// Resource limits
	memoryMB     int64
	memorySwapMB int64
	cpuPercent   int64
	cpuCount     float64
	
	// Security
	securityOpt  []string
	capDrop      []string
	capAdd       []string
	
	// Network
	networkMode  string
	dns          []string
	
	// Timeouts
	execTimeout  time.Duration
	
	// State
	isRunning    bool
	workspaceDir string
}

// SandboxConfig contains sandbox configuration.
type SandboxConfig struct {
	Image        string
	MemoryMB     int64
	CPUPercent   int64
	CPUCount     float64
	NetworkMode  string
	Timeout      time.Duration
	WorkspaceDir string
}

// DefaultSandboxConfig returns default configuration.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Image:        "openhands/sandbox:latest",
		MemoryMB:     2048,
		CPUPercent:   100000, // 1 CPU
		CPUCount:     1.0,
		NetworkMode:  "none", // No network by default
		Timeout:      30 * time.Second,
		WorkspaceDir: "/workspace",
	}
}

// NewSandbox creates a new sandbox.
func NewSandbox(cli *client.Client, config SandboxConfig) (*Sandbox, error) {
	if cli == nil {
		var err error
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("create docker client: %w", err)
		}
	}
	
	sb := &Sandbox{
		client:       cli,
		image:        config.Image,
		memoryMB:     config.MemoryMB,
		memorySwapMB: config.MemoryMB, // No swap
		cpuPercent:   config.CPUPercent,
		cpuCount:     config.CPUCount,
		securityOpt:  []string{"no-new-privileges:true"},
		capDrop:      []string{"ALL"},
		capAdd:       []string{"CHOWN", "SETGID", "SETUID"},
		networkMode:  config.NetworkMode,
		execTimeout:  config.Timeout,
		workspaceDir: config.WorkspaceDir,
	}
	
	return sb, nil
}

// Start initializes and starts the sandbox.
func (sb *Sandbox) Start(ctx context.Context, hostWorkspace string) error {
	if sb.isRunning {
		return fmt.Errorf("sandbox already running")
	}
	
	// Pull image if needed
	_, _, err := sb.client.ImageInspectWithRaw(ctx, sb.image)
	if err != nil {
		// Image not found, pull it
		pullReader, err := sb.client.ImagePull(ctx, sb.image, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("pull image: %w", err)
		}
		defer pullReader.Close()
		io.Copy(io.Discard, pullReader)
	}
	
	// Create container
	config := &container.Config{
		Image:        sb.image,
		WorkingDir:   sb.workspaceDir,
		Cmd:          []string{"sleep", "infinity"},
		AttachStdout: true,
		AttachStderr: true,
		Env: []string{
			"SANDBOX=true",
			"WORKSPACE=" + sb.workspaceDir,
		},
	}
	
	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s:rw", hostWorkspace, sb.workspaceDir),
		},
		
		// Resource limits
		Resources: container.Resources{
			Memory:     sb.memoryMB * 1024 * 1024,
			MemorySwap: sb.memorySwapMB * 1024 * 1024,
			CpuQuota:   sb.cpuPercent,
			CpuPeriod:  100000,
			CpuCount:   sb.cpuCount,
		},
		
		// Security
		CapDrop:      sb.capDrop,
		CapAdd:       sb.capAdd,
		SecurityOpt:  sb.securityOpt,
		NetworkMode:  container.NetworkMode(sb.networkMode),
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			"/tmp": "rw,noexec,nosuid,size=100m",
		},
	}
	
	resp, err := sb.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		nil,
		nil,
		"",
	)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	
	sb.containerID = resp.ID
	
	// Start container
	if err := sb.client.ContainerStart(ctx, sb.containerID, types.ContainerStartOptions{}); err != nil {
		// Cleanup on failure
		sb.client.ContainerRemove(ctx, sb.containerID, types.ContainerRemoveOptions{Force: true})
		return fmt.Errorf("start container: %w", err)
	}
	
	sb.isRunning = true
	
	return nil
}

// Execute runs a command in the sandbox.
func (sb *Sandbox) Execute(ctx context.Context, command string, options ...ExecuteOption) (*ExecutionResult, error) {
	if !sb.isRunning {
		return nil, fmt.Errorf("sandbox not running")
	}
	
	// Apply options
	opts := &executeOptions{
		timeout: sb.execTimeout,
		workdir: sb.workspaceDir,
	}
	for _, opt := range options {
		opt(opts)
	}
	
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()
	
	// Create exec
	execConfig := types.ExecConfig{
		Cmd:          []string{"sh", "-c", command},
		WorkingDir:   opts.workdir,
		AttachStdout: true,
		AttachStderr: true,
		Env:          opts.env,
	}
	
	execResp, err := sb.client.ContainerExecCreate(ctx, sb.containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("create exec: %w", err)
	}
	
	// Attach and run
	attachResp, err := sb.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("attach exec: %w", err)
	}
	defer attachResp.Close()
	
	// Read output
	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}
	
	// Get exit code
	inspect, err := sb.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("inspect exec: %w", err)
	}
	
	return &ExecutionResult{
		ExitCode:  inspect.ExitCode,
		Output:    string(output),
		TimedOut:  ctx.Err() == context.DeadlineExceeded,
		Duration:  opts.timeout,
	}, nil
}

// ExecuteScript runs a script file in the sandbox.
func (sb *Sandbox) ExecuteScript(ctx context.Context, scriptPath string, interpreter string) (*ExecutionResult, error) {
	if interpreter == "" {
		interpreter = "sh"
	}
	
	command := fmt.Sprintf("%s %s", interpreter, scriptPath)
	return sb.Execute(ctx, command)
}

// CopyTo copies a file into the sandbox.
func (sb *Sandbox) CopyTo(ctx context.Context, srcPath, dstPath string) error {
	if !sb.isRunning {
		return fmt.Errorf("sandbox not running")
	}
	
	// Use tar archive for copying
	// Implementation would use client.CopyToContainer
	return fmt.Errorf("not implemented")
}

// CopyFrom copies a file from the sandbox.
func (sb *Sandbox) CopyFrom(ctx context.Context, srcPath, dstPath string) error {
	if !sb.isRunning {
		return fmt.Errorf("sandbox not running")
	}
	
	// Use tar archive for copying
	// Implementation would use client.CopyFromContainer
	return fmt.Errorf("not implemented")
}

// Stop stops the sandbox.
func (sb *Sandbox) Stop(ctx context.Context) error {
	if !sb.isRunning {
		return nil
	}
	
	// Stop container
	timeout := 10
	if err := sb.client.ContainerStop(ctx, sb.containerID, container.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}
	
	// Remove container
	if err := sb.client.ContainerRemove(ctx, sb.containerID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	
	sb.isRunning = false
	sb.containerID = ""
	
	return nil
}

// IsRunning returns whether the sandbox is running.
func (sb *Sandbox) IsRunning() bool {
	return sb.isRunning
}

// GetContainerID returns the container ID.
func (sb *Sandbox) GetContainerID() string {
	return sb.containerID
}

// ExecutionResult contains command execution results.
type ExecutionResult struct {
	ExitCode   int
	Output     string
	ErrorOutput string
	TimedOut   bool
	Duration   time.Duration
}

// Success returns whether the execution succeeded.
func (r *ExecutionResult) Success() bool {
	return r.ExitCode == 0 && !r.TimedOut
}

// executeOptions contains execution options.
type executeOptions struct {
	timeout time.Duration
	workdir string
	env     []string
}

// ExecuteOption is a functional option for Execute.
type ExecuteOption func(*executeOptions)

// WithTimeout sets the execution timeout.
func WithTimeout(timeout time.Duration) ExecuteOption {
	return func(o *executeOptions) {
		o.timeout = timeout
	}
}

// WithWorkDir sets the working directory.
func WithWorkDir(workdir string) ExecuteOption {
	return func(o *executeOptions) {
		o.workdir = workdir
	}
}

// WithEnv sets environment variables.
func WithEnv(env []string) ExecuteOption {
	return func(o *executeOptions) {
		o.env = env
	}
}

// SandboxManager manages multiple sandboxes.
type SandboxManager struct {
	sandboxes map[string]*Sandbox
	mu        sync.RWMutex
	client    *client.Client
}

// NewSandboxManager creates a new sandbox manager.
func NewSandboxManager() (*SandboxManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	
	return &SandboxManager{
		sandboxes: make(map[string]*Sandbox),
		client:    cli,
	}, nil
}

// Create creates a new sandbox.
func (sm *SandboxManager) Create(id string, config SandboxConfig) (*Sandbox, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, exists := sm.sandboxes[id]; exists {
		return nil, fmt.Errorf("sandbox %s already exists", id)
	}
	
	sb, err := NewSandbox(sm.client, config)
	if err != nil {
		return nil, err
	}
	
	sm.sandboxes[id] = sb
	return sb, nil
}

// Get retrieves a sandbox by ID.
func (sm *SandboxManager) Get(id string) (*Sandbox, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sb, ok := sm.sandboxes[id]
	if !ok {
		return nil, fmt.Errorf("sandbox %s not found", id)
	}
	
	return sb, nil
}

// Remove removes a sandbox.
func (sm *SandboxManager) Remove(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sb, ok := sm.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox %s not found", id)
	}
	
	// Stop if running
	if sb.IsRunning() {
		sb.Stop(context.Background())
	}
	
	delete(sm.sandboxes, id)
	return nil
}

// List returns all sandbox IDs.
func (sm *SandboxManager) List() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	ids := make([]string, 0, len(sm.sandboxes))
	for id := range sm.sandboxes {
		ids = append(ids, id)
	}
	
	return ids
}

// Cleanup stops and removes all sandboxes.
func (sm *SandboxManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	ctx := context.Background()
	for _, sb := range sm.sandboxes {
		if sb.IsRunning() {
			sb.Stop(ctx)
		}
	}
	
	sm.sandboxes = make(map[string]*Sandbox)
}
