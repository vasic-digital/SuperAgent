# MCP Adapters Package

This package provides MCP (Model Context Protocol) server adapters for integrating with external services.

## Overview

The MCP adapters package implements standardized interfaces for connecting AI models to external tools and resources through the Model Context Protocol.

## Available Adapters (40+)

### Cloud Storage Adapters

| Adapter | File | Purpose |
|---------|------|---------|
| AWS S3 | `aws_s3.go` | Amazon S3 file operations |
| Google Drive | `google_drive.go` | Google Drive integration |

### Development Tools

| Adapter | File | Purpose |
|---------|------|---------|
| Docker | `docker.go` | Container management |
| Kubernetes | `kubernetes.go` | K8s cluster operations |
| GitLab | `gitlab.go` | GitLab repository operations |
| Puppeteer | `puppeteer.go` | Browser automation |

### Databases

| Adapter | File | Purpose |
|---------|------|---------|
| MongoDB | `mongodb.go` | MongoDB operations |

### Services

| Adapter | File | Purpose |
|---------|------|---------|
| Brave Search | `brave_search.go` | Web search |
| Datadog | `datadog.go` | Monitoring and observability |
| Figma | `figma.go` | Design file access |
| Miro | `miro.go` | Whiteboard collaboration |
| Notion | `notion.go` | Note and wiki management |
| Sentry | `sentry.go` | Error tracking |
| Slack | `slack.go` | Team messaging |
| Stable Diffusion | `stable_diffusion.go` | Image generation |

### Adapter Registry (`registry.go`)

Central registration and discovery of adapters:

```go
registry := adapters.NewRegistry()
registry.Register("aws_s3", adapters.NewAWSS3Adapter(config))
registry.Register("docker", adapters.NewDockerAdapter(config))

// Get adapter
adapter, err := registry.Get("docker")
```

## Architecture

```
┌─────────────────────────────────────────────┐
│              MCP Adapter Layer              │
│  ┌─────────────────────────────────────┐   │
│  │           Adapter Registry          │   │
│  │  ┌───────┐ ┌───────┐ ┌───────┐     │   │
│  │  │AWS S3 │ │Docker │ │GitLab │ ... │   │
│  │  └───────┘ └───────┘ └───────┘     │   │
│  └─────────────────────────────────────┘   │
│                    │                        │
│                    ▼                        │
│  ┌─────────────────────────────────────┐   │
│  │        MCP Protocol Handler         │   │
│  │   Tools │ Resources │ Prompts       │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Implementing a Custom Adapter

```go
package adapters

import (
    "context"
    "dev.helix.agent/internal/mcp"
)

type MyAdapter struct {
    config MyAdapterConfig
    client *MyServiceClient
}

func NewMyAdapter(config MyAdapterConfig) *MyAdapter {
    return &MyAdapter{
        config: config,
        client: NewMyServiceClient(config),
    }
}

// Implement MCPAdapter interface
func (a *MyAdapter) Name() string {
    return "my_adapter"
}

func (a *MyAdapter) Tools() []mcp.Tool {
    return []mcp.Tool{
        {
            Name:        "my_tool",
            Description: "Does something useful",
            Schema:      a.toolSchema(),
        },
    }
}

func (a *MyAdapter) ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
    switch name {
    case "my_tool":
        return a.executMyTool(ctx, params)
    default:
        return nil, mcp.ErrToolNotFound
    }
}

func (a *MyAdapter) Resources() []mcp.Resource {
    return []mcp.Resource{
        {
            URI:         "my-adapter://resource",
            Name:        "My Resource",
            Description: "Provides access to data",
        },
    }
}

func (a *MyAdapter) ReadResource(ctx context.Context, uri string) (interface{}, error) {
    // Implement resource reading
}
```

## Configuration

```yaml
mcp:
  adapters:
    aws_s3:
      enabled: true
      region: "${AWS_REGION}"
      access_key: "${AWS_ACCESS_KEY_ID}"
      secret_key: "${AWS_SECRET_ACCESS_KEY}"

    docker:
      enabled: true
      socket: "unix:///var/run/docker.sock"

    gitlab:
      enabled: true
      url: "https://gitlab.com"
      token: "${GITLAB_TOKEN}"

    slack:
      enabled: true
      token: "${SLACK_BOT_TOKEN}"
```

## Testing

```bash
go test -v ./internal/mcp/adapters/...
```

## Files

- `registry.go` - Adapter registration and discovery
- `aws_s3.go` - AWS S3 adapter
- `brave_search.go` - Brave Search adapter
- `datadog.go` - Datadog adapter
- `docker.go` - Docker adapter
- `docker_test.go` - Docker tests
- `figma.go` - Figma adapter
- `gitlab.go` - GitLab adapter
- `google_drive.go` - Google Drive adapter
- `kubernetes.go` - Kubernetes adapter
- `miro.go` - Miro adapter
- `mongodb.go` - MongoDB adapter
- `notion.go` - Notion adapter
- `puppeteer.go` - Puppeteer adapter
- `sentry.go` - Sentry adapter
- `slack.go` - Slack adapter
- `stable_diffusion.go` - Stable Diffusion adapter
- `stable_diffusion_test.go` - Stable Diffusion tests
