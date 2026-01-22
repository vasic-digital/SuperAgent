# CLI Notifications Package

This package provides CLI-specific rendering and output formatting for background task notifications.

## Overview

The CLI package handles rendering of task progress, status updates, and notifications for various CLI clients (OpenCode, Crush, HelixCode, KiloCode) with support for different output styles and color schemes.

## Features

- **Multiple Render Styles**: Theater, novel, screenplay, minimal, plain
- **Progress Bar Styles**: ASCII, Unicode, block, dots
- **Color Schemes**: None, 8-color, 256-color, true color
- **CLI Detection**: Auto-detect CLI client capabilities
- **Animated Output**: Support for real-time progress updates

## Components

### Types (`types.go`)

Core type definitions:

```go
type RenderStyle string

const (
    RenderStyleTheater    RenderStyle = "theater"    // Theatrical with boxes
    RenderStyleNovel      RenderStyle = "novel"      // Prose narration
    RenderStyleScreenplay RenderStyle = "screenplay" // Script format
    RenderStyleMinimal    RenderStyle = "minimal"    // Minimal formatting
    RenderStylePlain      RenderStyle = "plain"      // No formatting
)

type CLIClient string

const (
    CLIClientOpenCode  CLIClient = "opencode"
    CLIClientCrush     CLIClient = "crush"
    CLIClientHelixCode CLIClient = "helixcode"
    CLIClientKiloCode  CLIClient = "kilocode"
    CLIClientUnknown   CLIClient = "unknown"
)
```

### Configuration

```go
type RenderConfig struct {
    Style         RenderStyle      // Output style
    ProgressStyle ProgressBarStyle // Progress bar appearance
    ColorScheme   ColorScheme      // Color support level
    ShowResources bool             // Display resource usage
    ShowLogs      bool             // Display log lines
    LogLines      int              // Number of log lines
    Width         int              // Terminal width
    Animate       bool             // Enable animations
    RefreshRate   time.Duration    // Update frequency
}
```

### Progress Bar Styles

```go
type ProgressBarStyle string

const (
    ProgressBarStyleASCII   // [=====     ]
    ProgressBarStyleUnicode // [█████░░░░░]
    ProgressBarStyleBlock   // ▓▓▓▓▓░░░░░
    ProgressBarStyleDots    // ●●●●●○○○○○
)
```

### Renderer (`renderer.go`)

Handles output formatting and display.

### Detection (`detection.go`)

Auto-detects CLI client and capabilities.

## Data Types

### ProgressBarContent

```go
type ProgressBarContent struct {
    TaskID      string         // Task identifier
    TaskName    string         // Display name
    TaskType    string         // Task type
    Progress    float64        // 0-100 percentage
    Message     string         // Status message
    Status      string         // Current status
    StartedAt   time.Time      // Start time
    ETA         *time.Duration // Estimated time remaining
    CurrentStep int            // Current step number
    TotalSteps  int            // Total steps
}
```

### StatusTableContent

```go
type StatusTableContent struct {
    Tasks      []TaskStatusRow // Task rows
    TotalCount int             // Total task count
    Timestamp  time.Time       // Update timestamp
}
```

## Usage

### Basic Configuration

```go
import "dev.helix.agent/internal/notifications/cli"

config := cli.DefaultRenderConfig()
config.Style = cli.RenderStyleTheater
config.ColorScheme = cli.ColorScheme256
```

### Progress Bar Rendering

```go
progress := &cli.ProgressBarContent{
    TaskID:   "task-123",
    TaskName: "Processing files",
    Progress: 45.5,
    Message:  "Processing file 23/50",
    Status:   "running",
}

// Render with configured style
output := renderer.RenderProgress(progress)
```

### CLI Client Detection

```go
import "dev.helix.agent/internal/notifications/cli"

client := cli.DetectCLIClient()
switch client {
case cli.CLIClientOpenCode:
    // OpenCode-specific formatting
case cli.CLIClientCrush:
    // Crush-specific formatting
default:
    // Default formatting
}
```

### Custom Render Configuration

```go
config := &cli.RenderConfig{
    Style:         cli.RenderStyleMinimal,
    ProgressStyle: cli.ProgressBarStyleASCII,
    ColorScheme:   cli.ColorSchemeNone,
    Width:         120,
    Animate:       false,
}
```

## Render Style Examples

### Theater Style
```
╔═══════════════════════════════════════════╗
║ Task: Processing files (45%)              ║
║ [█████████████████░░░░░░░░░░░░░░░░] 45%  ║
║ Status: Processing file 23/50             ║
╚═══════════════════════════════════════════╝
```

### Minimal Style
```
Processing files: [=====     ] 45% - Processing file 23/50
```

### Plain Style
```
Task: Processing files (45%) - Processing file 23/50
```

## Testing

```bash
go test -v ./internal/notifications/cli/...
```

## Files

- `types.go` - Type definitions and constants
- `renderer.go` - Output rendering logic
- `detection.go` - CLI client detection
