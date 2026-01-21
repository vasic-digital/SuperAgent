# Features Package

The features package provides a feature flag system for controlling HelixAgent functionality at runtime.

## Overview

This package implements a thread-safe feature toggle system that enables gradual rollout, A/B testing, and dynamic configuration of HelixAgent features without code changes or restarts.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Feature Registry                          │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Feature Flags                          ││
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   ││
│  │  │ GraphQL  │ │   TOON   │ │  Compr.  │ │   MCP    │   ││
│  │  │  ✓ ON    │ │  ✓ ON    │ │  ✗ OFF   │ │  ✓ ON    │   ││
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘   ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Config File  │  │ Environment  │  │  Runtime API     │  │
│  │   Source     │  │    Source    │  │    Override      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### FeatureRegistry

The central feature flag manager.

```go
type FeatureRegistry struct {
    features map[string]*Feature
    mu       sync.RWMutex
    onChange []ChangeHandler
}
```

### Feature

Represents a single feature flag.

```go
type Feature struct {
    Name        string            // Feature identifier
    Description string            // Human-readable description
    Enabled     bool              // Current state
    Default     bool              // Default state
    Metadata    map[string]string // Additional metadata
    EnabledAt   time.Time         // When enabled (if enabled)
    ModifiedBy  string            // Who last modified
}
```

### ChangeHandler

Callback for feature state changes.

```go
type ChangeHandler func(feature string, oldValue, newValue bool)
```

## Predefined Features

HelixAgent includes these built-in feature flags:

| Feature | Default | Description |
|---------|---------|-------------|
| `graphql.enabled` | `true` | Enable GraphQL API endpoint |
| `toon.encoding` | `true` | Enable TOON encoding for responses |
| `compression.gzip` | `false` | Enable gzip compression |
| `protocol.mcp` | `true` | Enable MCP protocol support |
| `protocol.lsp` | `true` | Enable LSP protocol support |
| `protocol.acp` | `true` | Enable ACP protocol support |
| `debate.multipass` | `true` | Enable multi-pass validation |
| `debate.learning` | `true` | Enable cross-debate learning |
| `cache.semantic` | `true` | Enable semantic caching |
| `metrics.detailed` | `false` | Enable detailed metrics |
| `security.audit` | `true` | Enable audit logging |

## Configuration

### Via Environment Variables

```bash
# Enable/disable features via environment
export FEATURE_GRAPHQL_ENABLED=true
export FEATURE_TOON_ENCODING=true
export FEATURE_COMPRESSION_GZIP=false
```

### Via Configuration File

```yaml
# config/features.yaml
features:
  graphql.enabled: true
  toon.encoding: true
  compression.gzip: false
  protocol.mcp: true
  debate.multipass: true
  debate.learning: true
```

### Via Code

```go
import "dev.helix.agent/internal/features"

// Get the global registry
registry := features.Default()

// Or create a new one
registry := features.NewRegistry()
```

## Usage Examples

### Check Feature Status

```go
import "dev.helix.agent/internal/features"

// Simple check
if features.IsEnabled("graphql.enabled") {
    setupGraphQLEndpoint()
}

// With default fallback
enabled := features.IsEnabledWithDefault("new.feature", false)
```

### Enable/Disable at Runtime

```go
// Enable a feature
features.Enable("compression.gzip")

// Disable a feature
features.Disable("metrics.detailed")

// Toggle
features.Toggle("debug.mode")
```

### Listen for Changes

```go
features.OnChange(func(feature string, oldVal, newVal bool) {
    log.Printf("Feature %s changed: %v -> %v", feature, oldVal, newVal)

    // React to changes
    if feature == "compression.gzip" && newVal {
        enableGzipMiddleware()
    }
})
```

### Scoped Features

```go
// Check feature for specific context
if features.IsEnabledFor("beta.feature", "user-123") {
    // Enable for specific user
}

// Percentage rollout
features.SetRollout("new.ui", 0.25) // 25% of requests
```

### Feature Groups

```go
// Define feature groups
features.DefineGroup("protocols", []string{
    "protocol.mcp",
    "protocol.lsp",
    "protocol.acp",
})

// Enable entire group
features.EnableGroup("protocols")

// Check if all in group are enabled
if features.GroupEnabled("protocols") {
    // All protocol features are on
}
```

### HTTP Handler

```go
// Expose feature status via HTTP
router.GET("/features", features.HTTPHandler())
router.POST("/features/:name/toggle", features.ToggleHandler())
```

## Integration with HelixAgent

Features are checked throughout the codebase:

```go
// In router setup
if features.IsEnabled("graphql.enabled") {
    router.POST("/graphql", graphqlHandler)
}

// In response encoding
if features.IsEnabled("toon.encoding") {
    return toon.Encode(response)
}

// In debate service
if features.IsEnabled("debate.multipass") {
    return runMultiPassValidation(debate)
}
```

## Feature Lifecycle

```
1. Feature defined with default value
2. Configuration sources checked (env, file)
3. Runtime overrides applied
4. Change handlers notified
5. Feature checked throughout request lifecycle
```

## Testing

### Mocking Features

```go
func TestWithFeature(t *testing.T) {
    // Save current state
    registry := features.NewRegistry()

    // Set specific state for test
    registry.Enable("test.feature")

    // Run test with feature enabled
    result := functionUnderTest(registry)

    // Verify behavior
    assert.True(t, result.UsedFeature)
}
```

### Feature Test Matrix

```go
func TestAllFeatureCombinations(t *testing.T) {
    testCases := []struct {
        features map[string]bool
        expected string
    }{
        {map[string]bool{"a": true, "b": false}, "result1"},
        {map[string]bool{"a": false, "b": true}, "result2"},
    }

    for _, tc := range testCases {
        registry := features.NewRegistry()
        for f, v := range tc.features {
            if v {
                registry.Enable(f)
            }
        }
        // Test with this configuration
    }
}
```

## Testing

```bash
go test -v ./internal/features/...
```

## Best Practices

1. **Use descriptive names**: `debate.multipass.enabled` not `feature1`
2. **Default to safe values**: New features default to `false`
3. **Document features**: Add description for each feature
4. **Clean up old flags**: Remove features after full rollout
5. **Monitor usage**: Track which features are actually used
