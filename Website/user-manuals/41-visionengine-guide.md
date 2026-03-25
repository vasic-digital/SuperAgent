# User Manual 41: VisionEngine Guide

## Overview

VisionEngine (`digital.vasic.visionengine`) provides computer vision and LLM Vision capabilities for UI analysis, element detection, and navigation graph building with BFS pathfinding. It supports 4 LLM Vision providers and optional OpenCV integration.

## Prerequisites

- Go 1.24+
- At least one LLM Vision API key (OpenAI GPT-4o, Anthropic Claude, Google Gemini, or Qwen-VL)
- (Optional) GoCV/OpenCV 4.x for local computer vision (enabled via `vision` build tag)

## Step 1: Build VisionEngine

Standard build (LLM Vision only, no OpenCV dependency):

```bash
cd VisionEngine
go build ./...
```

Build with OpenCV support:

```bash
go build -tags vision ./...
```

## Step 2: Configure Vision Providers

Set API keys in your `.env` file:

```bash
# At least one provider is required
VISION_OPENAI_API_KEY=sk-...
VISION_ANTHROPIC_API_KEY=sk-ant-...
VISION_GEMINI_API_KEY=...
VISION_QWEN_API_KEY=sk-...

# Provider selection (comma-separated priority order)
VISION_PROVIDERS=openai,anthropic,gemini,qwen

# Optional: OpenCV toggle
VISION_OPENCV_ENABLED=false
```

Provider fallback: if the primary provider fails, VisionEngine tries the next in the configured order.

## Step 3: Analyze Screenshots

Send a screenshot for UI analysis:

```go
provider := llmvision.NewOpenAIProvider(apiKey)
analysis, err := provider.AnalyzeScreen(ctx, screenshotBytes)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Screen: %s\n", analysis.ScreenName)
for _, elem := range analysis.Elements {
    fmt.Printf("  [%s] %s at (%d,%d)\n", elem.Type, elem.Text, elem.X, elem.Y)
}
```

The `ScreenAnalysis` result contains:

| Field | Description |
|-------|-------------|
| `ScreenName` | Inferred screen identity |
| `Elements` | Detected UI elements with type, text, and coordinates |
| `TextContent` | All detected text on screen |
| `Issues` | Visual or UX issues detected |

## Step 4: Detect UI Elements

Element detection identifies interactive components:

```go
elements, err := provider.DetectElements(ctx, screenshotBytes)
for _, el := range elements {
    fmt.Printf("Type: %s, Label: %s, Bounds: %v, Clickable: %v\n",
        el.Type, el.Text, el.Bounds, el.Clickable)
}
```

Element types: `button`, `text_field`, `label`, `image`, `checkbox`, `dropdown`, `link`, `menu_item`.

## Step 5: Build a NavigationGraph

The NavigationGraph tracks screen transitions and supports BFS shortest-path queries:

```go
g := graph.NewNavigationGraph()

// Add screens discovered during testing
g.AddScreen(analyzer.ScreenIdentity{ID: "home", Name: "Home"})
g.AddScreen(analyzer.ScreenIdentity{ID: "settings", Name: "Settings"})
g.AddScreen(analyzer.ScreenIdentity{ID: "profile", Name: "Profile"})

// Record transitions observed during navigation
g.AddTransition("home", "settings", analyzer.Action{Type: "click", Target: "gear-icon"})
g.AddTransition("home", "profile", analyzer.Action{Type: "click", Target: "avatar"})
g.AddTransition("settings", "home", analyzer.Action{Type: "click", Target: "back-button"})

// Set current position
g.SetCurrent("home")
```

## Step 6: Use BFS Pathfinding

Find the shortest path between any two screens:

```go
path, err := g.PathTo("profile")
if err != nil {
    log.Fatal("no path found:", err)
}

for _, step := range path {
    fmt.Printf("From %s -> %s via %s on %s\n",
        step.From, step.To, step.Action.Type, step.Action.Target)
}
```

BFS guarantees the minimum number of navigation steps.

## Step 7: Export the Navigation Graph

Export for visualization or downstream analysis:

```go
// Mermaid (paste into markdown viewers)
mermaid := graph.ExportMermaid(g)
fmt.Println(mermaid)

// DOT format (for Graphviz rendering)
dot := graph.ExportDOT(g)

// JSON (machine-readable)
jsonBytes := graph.ExportJSON(g)
```

Sample Mermaid output:

```
graph LR
    home[Home] --> |click gear-icon| settings[Settings]
    home[Home] --> |click avatar| profile[Profile]
    settings[Settings] --> |click back-button| home[Home]
```

## Step 8: Run Tests

```bash
go test ./... -race -count=1            # Standard tests
go test -tags vision ./... -race        # With OpenCV tests
```

## Package Reference

| Package | Purpose |
|---------|---------|
| `pkg/analyzer` | Core interfaces: Analyzer, ScreenAnalysis, UIElement, ScreenDiff |
| `pkg/graph` | NavigationGraph with BFS pathfinding and DOT/JSON/Mermaid export |
| `pkg/llmvision` | VisionProvider interface + OpenAI, Anthropic, Gemini, Qwen adapters |
| `pkg/opencv` | OpenCV stubs (real implementation behind `vision` build tag) |
| `pkg/config` | Configuration via environment variables |

## Troubleshooting

- **"no vision provider configured"**: Set at least one `VISION_*_API_KEY` in `.env`
- **OpenCV build fails**: Install GoCV dependencies (`apt install libopencv-dev`) and use `-tags vision`
- **BFS returns no path**: Verify both screens exist in the graph and transitions connect them
- **Low element detection accuracy**: Try a different vision provider or increase image resolution

## Related Resources

- [User Manual 39: HelixQA Guide](39-helixqa-guide.md) -- Autonomous QA using VisionEngine for screen analysis
- [User Manual 40: LLMOrchestrator Guide](40-llmorchestrator-guide.md) -- Agent management for vision-guided testing
- Source: `VisionEngine/README.md`, `VisionEngine/CLAUDE.md`
