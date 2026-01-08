# HelixAgent Security Sandboxing

## Overview

HelixAgent implements comprehensive security sandboxing to isolate plugin execution, tool operations, and external integrations. This document covers the security model, sandboxing capabilities, configuration options, and best practices.

## Security Model

### Defense in Depth

HelixAgent employs a multi-layered security approach:

```
+--------------------------------------------------+
|              Application Layer                    |
|  - Input validation                               |
|  - Authentication & Authorization                 |
|  - Rate limiting                                  |
+--------------------------------------------------+
|              Sandbox Layer                        |
|  - Process isolation                              |
|  - Resource limits                                |
|  - Capability restrictions                        |
+--------------------------------------------------+
|              Container Layer                      |
|  - Docker/Podman isolation                        |
|  - Network namespaces                             |
|  - Filesystem restrictions                        |
+--------------------------------------------------+
|              System Layer                         |
|  - SELinux/AppArmor profiles                      |
|  - Kernel security modules                        |
|  - Audit logging                                  |
+--------------------------------------------------+
```

### Security Principles

1. **Least Privilege**: Components receive only the permissions they require
2. **Isolation**: Plugins and tools run in isolated environments
3. **Validation**: All inputs are validated before processing
4. **Audit**: All security-relevant operations are logged
5. **Fail Secure**: Failures default to denying access

## Sandboxing Capabilities

### Process Isolation

HelixAgent can run plugins and tools in isolated processes:

```go
type ProcessSandbox struct {
    // Isolation settings
    Enabled         bool
    UseNamespaces   bool
    UseSeccomp      bool
    UseCgroups      bool

    // Resource limits
    MaxMemoryBytes  int64
    MaxCPUPercent   int
    MaxProcesses    int
    MaxOpenFiles    int

    // Filesystem restrictions
    RootFS          string
    ReadOnlyPaths   []string
    WritablePaths   []string
    MaskedPaths     []string

    // Network restrictions
    NetworkMode     string  // "none", "host", "bridge"
    AllowedHosts    []string
    AllowedPorts    []int
}
```

### Container Isolation

For maximum isolation, HelixAgent supports Docker/Podman containers:

```go
type ContainerSandbox struct {
    // Container runtime
    Runtime         string  // "docker", "podman"
    Image           string

    // Security options
    ReadOnly        bool
    NoNewPrivileges bool
    DropCapabilities []string
    SecurityOpt     []string

    // Resource limits
    MemoryLimit     string  // "512m"
    CPULimit        string  // "1.0"
    PidsLimit       int64

    // Mounts
    Volumes         []VolumeMount
    Tmpfs           []string
}

type VolumeMount struct {
    Source      string
    Target      string
    ReadOnly    bool
    Type        string  // "bind", "volume", "tmpfs"
}
```

### Network Isolation

Control network access for sandboxed components:

```go
type NetworkPolicy struct {
    // Access controls
    AllowOutbound   bool
    AllowInbound    bool
    AllowDNS        bool

    // Allowed destinations
    AllowedHosts    []HostRule
    AllowedCIDRs    []string

    // Port restrictions
    AllowedPorts    []PortRule
    BlockedPorts    []int
}

type HostRule struct {
    Host     string
    Ports    []int
    Protocol string  // "tcp", "udp"
}

type PortRule struct {
    Port     int
    Protocol string
    Allow    bool
}
```

### Filesystem Isolation

Restrict filesystem access:

```go
type FilesystemPolicy struct {
    // Base restrictions
    RootPath        string
    ChrootEnabled   bool

    // Path rules
    AllowedPaths    []PathRule
    BlockedPaths    []string
    TempDirectory   string

    // File operations
    AllowRead       bool
    AllowWrite      bool
    AllowExecute    bool
    AllowSymlinks   bool
}

type PathRule struct {
    Path        string
    Permissions string  // "r", "rw", "rwx"
    Recursive   bool
}
```

## Configuration Options

### Global Security Configuration

```yaml
security:
  enabled: true

  # Sandbox configuration
  sandbox:
    enabled: true
    type: "container"  # "process", "container", "none"

    # Process sandbox settings
    process:
      use_namespaces: true
      use_seccomp: true
      use_cgroups: true

    # Container sandbox settings
    container:
      runtime: "docker"
      default_image: "helixagent/sandbox:latest"
      read_only: true
      no_new_privileges: true
      drop_capabilities:
        - ALL
      add_capabilities:
        - NET_BIND_SERVICE

  # Resource limits
  limits:
    max_memory: "512Mi"
    max_cpu: "1.0"
    max_processes: 100
    max_open_files: 1024
    max_execution_time: 300s

  # Network policy
  network:
    default_policy: "deny"
    allow_dns: true
    allowed_hosts:
      - host: "api.openai.com"
        ports: [443]
      - host: "api.anthropic.com"
        ports: [443]
    blocked_cidrs:
      - "10.0.0.0/8"
      - "172.16.0.0/12"
      - "192.168.0.0/16"

  # Filesystem policy
  filesystem:
    default_policy: "deny"
    temp_directory: "/tmp/sandbox"
    allowed_paths:
      - path: "/etc/ssl/certs"
        permissions: "r"
      - path: "/tmp/sandbox"
        permissions: "rw"
    blocked_paths:
      - "/etc/passwd"
      - "/etc/shadow"
      - "/root"
```

### Plugin Security Configuration

```yaml
plugins:
  security:
    # Default security context for plugins
    default_context:
      sandbox_enabled: true
      max_memory_mb: 256
      max_cpu_percent: 25
      timeout: 60s

      permissions:
        allow_network: true
        allow_filesystem: false
        allow_exec: false
        allow_environment: false

    # Per-plugin overrides
    overrides:
      myplugin:
        max_memory_mb: 512
        permissions:
          allow_filesystem: true
          allowed_paths:
            - "/data/models"
```

### Tool Execution Security

```yaml
tools:
  security:
    # Sandbox all tool executions
    sandbox_enabled: true

    # Command validation
    command_validation:
      enabled: true
      blocked_commands:
        - "rm -rf"
        - "sudo"
        - "chmod"
        - "chown"
      blocked_patterns:
        - ".*\\|.*"  # Pipe commands
        - ".*>.*"    # Redirects
        - ".*`.*"    # Command substitution

    # Environment sanitization
    environment:
      inherit: false
      allowed_vars:
        - "PATH"
        - "HOME"
        - "LANG"
      blocked_vars:
        - "AWS_*"
        - "GOOGLE_*"
        - "*_KEY"
        - "*_SECRET"
```

## Security Implementation

### Sandbox Manager

```go
type SandboxManager struct {
    config     *SandboxConfig
    runtime    ContainerRuntime
    processes  map[string]*SandboxedProcess
    metrics    *SandboxMetrics
    logger     *logrus.Logger
    mu         sync.RWMutex
}

// CreateSandbox creates a new sandbox environment
func (m *SandboxManager) CreateSandbox(ctx context.Context, opts *SandboxOptions) (*Sandbox, error) {
    // Validate options
    if err := m.validateOptions(opts); err != nil {
        return nil, fmt.Errorf("invalid sandbox options: %w", err)
    }

    // Create sandbox based on type
    switch m.config.Type {
    case "container":
        return m.createContainerSandbox(ctx, opts)
    case "process":
        return m.createProcessSandbox(ctx, opts)
    default:
        return nil, fmt.Errorf("unknown sandbox type: %s", m.config.Type)
    }
}

// Execute runs a command in the sandbox
func (s *Sandbox) Execute(ctx context.Context, cmd *Command) (*Result, error) {
    // Validate command
    if err := s.validateCommand(cmd); err != nil {
        return nil, fmt.Errorf("command validation failed: %w", err)
    }

    // Apply resource limits
    ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
    defer cancel()

    // Execute with monitoring
    result, err := s.executeWithMonitoring(ctx, cmd)
    if err != nil {
        s.metrics.RecordFailure(cmd.Name)
        return nil, err
    }

    s.metrics.RecordSuccess(cmd.Name)
    return result, nil
}
```

### Command Validation

```go
type CommandValidator struct {
    blockedCommands  []string
    blockedPatterns  []*regexp.Regexp
    allowedCommands  []string
}

func (v *CommandValidator) Validate(cmd *Command) error {
    // Check blocked commands
    for _, blocked := range v.blockedCommands {
        if strings.HasPrefix(cmd.String(), blocked) {
            return fmt.Errorf("blocked command: %s", blocked)
        }
    }

    // Check blocked patterns
    for _, pattern := range v.blockedPatterns {
        if pattern.MatchString(cmd.String()) {
            return fmt.Errorf("command matches blocked pattern")
        }
    }

    // Validate arguments
    for _, arg := range cmd.Args {
        if err := v.validateArgument(arg); err != nil {
            return fmt.Errorf("invalid argument: %w", err)
        }
    }

    return nil
}

func (v *CommandValidator) validateArgument(arg string) error {
    // Check for shell injection
    dangerousChars := []string{";", "|", "&", "$", "`", "(", ")", "{", "}", "<", ">"}
    for _, char := range dangerousChars {
        if strings.Contains(arg, char) {
            return fmt.Errorf("argument contains dangerous character: %s", char)
        }
    }
    return nil
}
```

### Resource Monitoring

```go
type ResourceMonitor struct {
    pid        int
    limits     *ResourceLimits
    metrics    *ResourceMetrics
    violations chan *Violation
    stopCh     chan struct{}
}

func (m *ResourceMonitor) Start() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.checkResources()
        case <-m.stopCh:
            return
        }
    }
}

func (m *ResourceMonitor) checkResources() {
    stats := m.getProcessStats()

    // Check memory
    if stats.MemoryBytes > m.limits.MaxMemory {
        m.violations <- &Violation{
            Type:    "memory",
            Current: stats.MemoryBytes,
            Limit:   m.limits.MaxMemory,
        }
    }

    // Check CPU
    if stats.CPUPercent > float64(m.limits.MaxCPU) {
        m.violations <- &Violation{
            Type:    "cpu",
            Current: int64(stats.CPUPercent),
            Limit:   int64(m.limits.MaxCPU),
        }
    }

    // Record metrics
    m.metrics.RecordMemory(stats.MemoryBytes)
    m.metrics.RecordCPU(stats.CPUPercent)
}
```

## Best Practices

### 1. Enable Sandboxing by Default

Always enable sandboxing for production deployments:

```yaml
security:
  sandbox:
    enabled: true
    type: "container"
```

### 2. Use Minimal Permissions

Grant only the permissions required:

```yaml
plugins:
  security:
    default_context:
      permissions:
        allow_network: true    # Only if needed
        allow_filesystem: false
        allow_exec: false
        allow_environment: false
```

### 3. Whitelist Instead of Blacklist

Prefer allowing specific resources over blocking:

```yaml
# GOOD: Whitelist approach
network:
  default_policy: "deny"
  allowed_hosts:
    - "api.openai.com"

# BAD: Blacklist approach
network:
  default_policy: "allow"
  blocked_hosts:
    - "internal.corp.com"
```

### 4. Set Resource Limits

Always configure resource limits:

```yaml
limits:
  max_memory: "256Mi"      # Prevent memory exhaustion
  max_cpu: "0.5"           # Limit CPU usage
  max_execution_time: 60s  # Prevent runaway processes
```

### 5. Enable Audit Logging

Log all security-relevant events:

```yaml
security:
  audit:
    enabled: true
    log_level: "info"
    events:
      - "sandbox_create"
      - "sandbox_destroy"
      - "command_execute"
      - "resource_violation"
      - "permission_denied"
```

### 6. Validate All Inputs

Implement comprehensive input validation:

```go
func validateInput(input string) error {
    // Check length
    if len(input) > maxInputLength {
        return ErrInputTooLong
    }

    // Check for null bytes
    if strings.ContainsRune(input, 0) {
        return ErrInvalidInput
    }

    // Sanitize for shell execution
    if containsShellMetacharacters(input) {
        return ErrDangerousInput
    }

    return nil
}
```

### 7. Use Container Isolation for Plugins

For maximum security, run plugins in containers:

```yaml
plugins:
  security:
    default_context:
      sandbox_enabled: true
      sandbox_type: "container"
      container:
        image: "helixagent/plugin-sandbox:latest"
        read_only: true
        no_new_privileges: true
```

### 8. Implement Network Segmentation

Isolate sensitive operations:

```yaml
network:
  zones:
    trusted:
      cidrs: ["10.0.0.0/24"]
      allow_all: true
    untrusted:
      default_policy: "deny"
      allowed_hosts:
        - "api.external.com"
```

### 9. Regular Security Updates

Keep sandbox images and dependencies updated:

```bash
# Update sandbox images
docker pull helixagent/sandbox:latest
docker pull helixagent/plugin-sandbox:latest

# Scan for vulnerabilities
trivy image helixagent/sandbox:latest
```

### 10. Monitor and Alert

Set up monitoring for security events:

```yaml
monitoring:
  security:
    alerts:
      - name: "sandbox_violation"
        condition: "sandbox_violations > 0"
        severity: "warning"
      - name: "permission_denied_spike"
        condition: "rate(permission_denied[5m]) > 10"
        severity: "critical"
```

## Troubleshooting

### Common Issues

#### Sandbox Creation Fails

1. Check container runtime is available
2. Verify image exists and is accessible
3. Review resource limits (may be too restrictive)
4. Check filesystem permissions

#### Permission Denied Errors

1. Review allowed paths configuration
2. Check if required permissions are granted
3. Verify network policies allow required connections
4. Enable debug logging for detailed error messages

#### Performance Issues

1. Increase resource limits if appropriate
2. Consider using process isolation instead of containers
3. Review network policies for latency impacts
4. Monitor resource usage metrics

---

For more information, see the [HelixAgent Security Documentation](https://github.com/helixagent/helixagent/tree/main/docs/security).
