# Config Package

The config package provides configuration management for HelixAgent.

## Overview

Implements:
- YAML/JSON configuration loading
- Environment variable overrides
- Configuration validation
- Hot-reload support

## Key Components

```go
cfg, err := config.Load("config.yaml")

// Access values
port := cfg.GetInt("server.port")
debug := cfg.GetBool("debug")
```

## Configuration Files

- `configs/development.yaml`
- `configs/production.yaml`
- `configs/multi-provider.yaml`
