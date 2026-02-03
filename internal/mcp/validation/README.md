# MCP Validation Utilities

This package provides comprehensive validation and verification utilities for MCP (Model Context Protocol) servers. It validates server requirements, checks dependencies, and generates detailed compliance reports.

## Overview

The MCP validator ensures that:
- **Environment Requirements**: All required API keys and credentials are present
- **Local Services**: Required local services (PostgreSQL, Redis, Docker) are running
- **Server Health**: MCP servers are accessible and responding correctly
- **Configuration Validity**: Server configurations are complete and valid

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           MCP Validator                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    Requirements Registry                                 ││
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ ││
│  │  │ Core MCPs│  │ Database │  │  DevOps  │  │  Search  │  │Productivity││
│  │  │          │  │   MCPs   │  │   MCPs   │  │   MCPs   │  │   MCPs    ││
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘ ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                    │                                         │
│                                    ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    Validation Engine                                     ││
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐               ││
│  │  │ Env Var Check │  │ Service Check │  │ Health Check  │               ││
│  │  │ (.env, env)   │  │ (pg, redis)   │  │ (HTTP probe)  │               ││
│  │  └───────────────┘  └───────────────┘  └───────────────┘               ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                    │                                         │
│                                    ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    Validation Report                                     ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  ││
│  │  │  Working     │  │  Disabled    │  │   Failed     │                  ││
│  │  │  MCPs List   │  │  MCPs List   │  │  MCPs List   │                  ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘                  ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
```

## Validation Framework

### MCP Requirements Definition

```go
type MCPRequirement struct {
    Name           string   // MCP server name
    Type           string   // "local" or "remote"
    Package        string   // NPM package name
    Command        []string // Execution command
    URL            string   // For remote MCPs
    RequiredEnvs   []string // Required environment variables
    OptionalEnvs   []string // Optional environment variables
    LocalServices  []string // Required local services
    Description    string   // Human-readable description
    Category       string   // MCP category
    CanWorkLocally bool     // Can work without external API
    Enabled        bool     // Default enabled state
    Priority       int      // Importance (higher = more important)
}
```

### Validation Result

```go
type MCPValidationResult struct {
    Name            string    // MCP name
    Status          string    // "works", "disabled", "failed", "missing_deps"
    CanEnable       bool      // Whether MCP can be enabled
    MissingEnvVars  []string  // Missing environment variables
    MissingServices []string  // Missing local services
    ErrorMessage    string    // Error details
    ResponseTimeMs  int64     // Health check latency
    TestedAt        time.Time // Validation timestamp
    Category        string    // MCP category
    Reason          string    // Human-readable status reason
}
```

## Compliance Checking

### Running Validation

```go
import "dev.helix.agent/internal/mcp/validation"

// Create validator
validator := validation.NewMCPValidator()

// Validate all MCPs
ctx := context.Background()
report := validator.ValidateAll(ctx)

// Print summary
fmt.Printf("Total: %d, Working: %d, Disabled: %d, Failed: %d\n",
    report.TotalMCPs,
    report.WorkingMCPs,
    report.DisabledMCPs,
    report.FailedMCPs)

// List enabled MCPs
for _, name := range report.EnabledMCPList {
    fmt.Printf("  + %s\n", name)
}

// List disabled MCPs with reasons
for _, name := range report.DisabledMCPList {
    result := report.Results[name]
    fmt.Printf("  - %s: %s\n", name, result.Reason)
}
```

### Category-Based Validation

The validator categorizes MCPs by type:

| Category | Description | Examples |
|----------|-------------|----------|
| helixagent | HelixAgent plugin | helixagent |
| core | No API keys needed | filesystem, fetch, memory, time, git |
| database | Local databases | sqlite, postgres, redis, mongodb |
| devops | Infrastructure tools | docker, kubernetes, puppeteer |
| development | Dev platforms | github, gitlab, sentry |
| communication | Chat/messaging | slack, discord |
| search | Search APIs | brave-search, exa |
| productivity | Productivity tools | notion, linear, todoist |
| cloud | Cloud providers | cloudflare, aws-s3 |

### Environment Variable Checks

```go
// Validator checks for required environment variables
validator := validation.NewMCPValidator()

// Check specific requirement
req := validator.GetMCPConfig("github")
// req.RequiredEnvs = []string{"GITHUB_TOKEN"}

// Validation result shows missing vars
result := report.Results["github"]
if !result.CanEnable {
    fmt.Printf("Missing: %v\n", result.MissingEnvVars)
    // Output: Missing: [GITHUB_TOKEN]
}
```

### Local Service Checks

The validator checks for running local services:

```go
// Services checked:
// - helixagent: HTTP GET http://localhost:7061/health
// - postgresql: pg_isready -h localhost -p 15432
// - redis: redis-cli -p 16379 ping
// - docker: docker info (or podman info)
// - kubernetes: kubectl cluster-info
```

## Testing Utilities

### Generating Test Reports

```go
// Generate human-readable report
report := validator.GenerateReport()
fmt.Println(report)
```

Output:
```
===============================================================================
MCP VALIDATION REPORT
===============================================================================

Summary: 25 total, 12 working, 10 disabled, 3 failed

WORKING MCPs (Enabled):
-------------------------------------------------------------------------------
  ✓ filesystem               [core] File system access - read, write, list files
  ✓ fetch                    [core] HTTP fetch - make web requests
  ✓ memory                   [core] In-memory key-value storage
  ✓ time                     [core] Time and timezone utilities
  ✓ git                      [core] Git repository operations
  ✓ sqlite                   [database] SQLite database operations
  ✓ docker                   [devops] Docker container management
  ✓ github                   [development] GitHub API operations

DISABLED MCPs (Missing Requirements):
-------------------------------------------------------------------------------
  ✗ gitlab                   [development] Missing required environment variables: GITLAB_TOKEN
  ✗ slack                    [communication] Missing required environment variables: SLACK_BOT_TOKEN, SLACK_TEAM_ID
  ✗ kubernetes               [devops] Missing required environment variables: KUBECONFIG

===============================================================================
```

### JSON Report

```go
jsonData, err := validator.ToJSON()
if err != nil {
    log.Fatal(err)
}
os.WriteFile("mcp_validation_report.json", jsonData, 0644)
```

Output:
```json
{
  "generated_at": "2024-01-15T10:30:00Z",
  "total_mcps": 25,
  "working_mcps": 12,
  "disabled_mcps": 10,
  "failed_mcps": 3,
  "results": {
    "filesystem": {
      "name": "filesystem",
      "status": "works",
      "can_enable": true,
      "category": "core",
      "reason": "All requirements satisfied"
    },
    "gitlab": {
      "name": "gitlab",
      "status": "disabled",
      "can_enable": false,
      "missing_env_vars": ["GITLAB_TOKEN"],
      "category": "development",
      "reason": "Missing required environment variables: GITLAB_TOKEN"
    }
  },
  "enabled_mcp_list": ["filesystem", "fetch", "memory", ...],
  "disabled_mcp_list": ["gitlab", "slack", "kubernetes", ...]
}
```

### Programmatic Testing

```go
// Get enabled MCPs for configuration
enabledMCPs := validator.GetEnabledMCPs()
for _, name := range enabledMCPs {
    config := validator.GetMCPConfig(name)
    fmt.Printf("Configuring %s: %v\n", name, config.Command)
}

// Integration test helper
func TestMCPAvailability(t *testing.T) {
    validator := validation.NewMCPValidator()
    report := validator.ValidateAll(context.Background())

    // Ensure core MCPs are working
    coreMCPs := []string{"filesystem", "fetch", "memory", "time", "git"}
    for _, name := range coreMCPs {
        result := report.Results[name]
        if result.Status != "works" {
            t.Errorf("Core MCP %s not working: %s", name, result.Reason)
        }
    }
}
```

## MCP Registry

### Built-in MCP Definitions

```go
// Core MCPs (always available)
requirements["filesystem"] = &MCPRequirement{
    Name:           "filesystem",
    Type:           "local",
    Package:        "@modelcontextprotocol/server-filesystem",
    Command:        []string{"npx", "-y", "@modelcontextprotocol/server-filesystem", "${HOME}"},
    Description:    "File system access - read, write, list files",
    Category:       "core",
    CanWorkLocally: true,
    Enabled:        true,
    Priority:       95,
}

// Database MCPs
requirements["postgres"] = &MCPRequirement{
    Name:           "postgres",
    Type:           "local",
    Package:        "@modelcontextprotocol/server-postgres",
    Command:        []string{"npx", "-y", "@modelcontextprotocol/server-postgres"},
    RequiredEnvs:   []string{"POSTGRES_URL"},
    Description:    "PostgreSQL database operations",
    Category:       "database",
    LocalServices:  []string{"postgresql"},
    CanWorkLocally: true,
    Enabled:        true,
    Priority:       80,
}

// API-dependent MCPs
requirements["github"] = &MCPRequirement{
    Name:           "github",
    Type:           "local",
    Package:        "@modelcontextprotocol/server-github",
    Command:        []string{"npx", "-y", "@modelcontextprotocol/server-github"},
    RequiredEnvs:   []string{"GITHUB_TOKEN"},
    Description:    "GitHub API operations",
    Category:       "development",
    CanWorkLocally: false,
    Enabled:        true,
    Priority:       85,
}
```

### Priority System

Higher priority MCPs are more important:

| Priority | Category | Examples |
|----------|----------|----------|
| 100 | HelixAgent | helixagent |
| 95 | Core | filesystem, fetch |
| 90 | Core | memory, time, git |
| 85 | Development | github, puppeteer, sequential-thinking |
| 80 | Database | postgres, docker |
| 75 | Communication | redis, slack |
| 70 | Other | Various MCPs |
| 50 | Testing | everything |

## Usage Examples

### CLI Integration

```go
func main() {
    validator := validation.NewMCPValidator()
    report := validator.ValidateAll(context.Background())

    if len(os.Args) > 1 && os.Args[1] == "--json" {
        data, _ := validator.ToJSON()
        fmt.Println(string(data))
    } else {
        fmt.Println(validator.GenerateReport())
    }

    // Exit with error if critical MCPs are missing
    criticalMCPs := []string{"filesystem", "fetch", "memory"}
    for _, name := range criticalMCPs {
        if !report.Results[name].CanEnable {
            os.Exit(1)
        }
    }
}
```

### Health Check Endpoint

```go
func handleValidation(w http.ResponseWriter, r *http.Request) {
    validator := validation.NewMCPValidator()
    report := validator.ValidateAll(r.Context())

    w.Header().Set("Content-Type", "application/json")

    if report.WorkingMCPs < 5 {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    json.NewEncoder(w).Encode(report)
}
```

### Pre-flight Check

```go
func PreflightCheck() error {
    validator := validation.NewMCPValidator()
    report := validator.ValidateAll(context.Background())

    // Require minimum MCPs
    if report.WorkingMCPs < 5 {
        return fmt.Errorf("insufficient MCPs: need 5+, have %d", report.WorkingMCPs)
    }

    // Require specific MCPs
    required := []string{"filesystem", "fetch", "git"}
    for _, name := range required {
        if !report.Results[name].CanEnable {
            return fmt.Errorf("required MCP not available: %s", name)
        }
    }

    return nil
}
```

## Testing

```bash
# Unit tests
go test -v ./internal/mcp/validation/...

# Test with environment
GITHUB_TOKEN=test POSTGRES_URL=test \
go test -v ./internal/mcp/validation/...

# Verbose validation test
go test -v ./internal/mcp/validation/... -run TestValidateAll
```

## Related Files

- `validator.go` - Main validator implementation
- `validator_test.go` - Unit tests
