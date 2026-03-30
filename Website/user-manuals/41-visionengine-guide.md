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

## Remote Vision Pool

For high-throughput testing scenarios — multiple devices, parallel autonomous sessions, or large screenshot volumes — VisionEngine supports a `VisionPool` that distributes analysis requests across a pool of remote vision backends.

The `VisionPool` assigns each backend a fixed number of concurrent analysis slots (`MaxConcurrentPerSlot`) and routes incoming requests to the least-loaded slot. Backends can be Ollama instances serving multimodal models (e.g., `llava`, `moondream`) or llama.cpp HTTP servers.

Configure the pool in `.env`:

```bash
# Pool backend definitions
VISION_POOL_BACKENDS=http://gpu-host-1:11434,http://gpu-host-2:11434,http://gpu-host-3:8080
VISION_POOL_BACKEND_TYPES=ollama,ollama,llamacpp
VISION_POOL_MAX_CONCURRENT_PER_SLOT=2
VISION_POOL_MODEL=llava:13b
VISION_POOL_TIMEOUT=60s
VISION_POOL_HEALTH_CHECK_INTERVAL=30s
```

Use the pool in code:

```go
import "github.com/digital/vasic/visionengine/pkg/llmvision"

pool, err := llmvision.NewVisionPool(llmvision.VisionPoolConfig{
    Backends: []llmvision.BackendConfig{
        {URL: "http://gpu-host-1:11434", Type: llmvision.BackendOllama, Model: "llava:13b"},
        {URL: "http://gpu-host-2:11434", Type: llmvision.BackendOllama, Model: "llava:13b"},
        {URL: "http://gpu-host-3:8080",  Type: llmvision.BackendLlamaCpp, Model: "llava-v1.6"},
    },
    MaxConcurrentPerSlot: 2,
    Timeout:              60 * time.Second,
})
if err != nil {
    log.Fatal(err)
}

// Pool implements the VisionProvider interface — drop-in replacement
analysis, err := pool.AnalyzeImage(ctx, screenshotBytes, "Describe the UI layout")
```

The pool provides:

- **Automatic health monitoring**: unhealthy backends are removed from rotation and re-checked on the configured interval
- **Least-loaded routing**: requests go to the backend with the most available slots
- **Graceful degradation**: if all backends for a model type are down, falls back to the next `FallbackProvider` in the chain
- **Metrics**: per-backend request counts, error rates, and latency exposed via Prometheus

## LlamaCpp Deployer

The `LlamaCppDeployer` manages llama.cpp server instances on remote GPU hosts. It handles binary distribution, model download, process launch, and health verification — enabling fully automated remote vision backend setup.

```go
import "github.com/digital/vasic/visionengine/pkg/llmvision"

deployer := llmvision.NewLlamaCppDeployer(llmvision.DeployerConfig{
    RemoteHost:   "gpu-host-3",
    RemoteUser:   "ubuntu",
    SSHKeyPath:   "/home/user/.ssh/id_ed25519",
    BinaryPath:   "/usr/local/bin/llama-server",
    ModelPath:    "/models/llava-v1.6-mistral-7b.gguf",
    MmprojPath:   "/models/llava-v1.6-mistral-7b-mmproj.gguf",
    ListenPort:   8080,
    NGL:          35,    // GPU layers offloaded
    CtxSize:      4096,
    Threads:      8,
})

endpoint, err := deployer.Deploy(ctx)
if err != nil {
    log.Fatal("deploy failed:", err)
}
fmt.Printf("llama.cpp server ready at %s\n", endpoint)
// endpoint: "http://gpu-host-3:8080"
```

`LlamaCppDeployer` operations:

| Method | Description |
|--------|-------------|
| `Deploy(ctx)` | SSH to host, start `llama-server`, wait for health check |
| `Undeploy(ctx)` | Stop the remote process cleanly |
| `Status(ctx)` | Check if the remote server is running and healthy |
| `Redeploy(ctx)` | Stop, optionally update model, restart |

All SSH operations use key-based authentication — no password prompts. The deployer is non-interactive and compatible with automated pipelines.

## Semaphore-Based Concurrency

When VisionEngine is used in parallel test execution (multiple `PlatformWorker` instances, multi-device Android testing, or concurrent autonomous sessions), unbounded concurrency against vision backends can exhaust API rate limits or overload GPU hosts.

`MaxConcurrentPerSlot` is enforced via a semaphore on each backend slot within `VisionPool`. Each slot holds a `chan struct{}` of capacity `MaxConcurrentPerSlot`. An analysis request acquires one token before sending to the backend and releases it when the response arrives (or on error/timeout).

```
VisionPool
├── Backend slot 0 (gpu-host-1:11434)  semaphore capacity: 2
│     token ████░░  (1 in-flight, 1 free)
├── Backend slot 1 (gpu-host-2:11434)  semaphore capacity: 2
│     token ████████  (2 in-flight, 0 free → requests queue here)
└── Backend slot 2 (gpu-host-3:8080)   semaphore capacity: 2
      token ░░░░  (0 in-flight, 2 free → next request routed here)
```

The router selects the slot with the most free tokens (least loaded). If all slots are at capacity, the caller blocks until a token is released or the request context is cancelled.

Tuning guidance:

| Scenario | Recommended `MaxConcurrentPerSlot` |
|----------|------------------------------------|
| Cloud LLM API (rate-limited) | 1–2 |
| Ollama on consumer GPU (8 GB VRAM) | 1 |
| Ollama on workstation GPU (24 GB VRAM) | 2–3 |
| llama.cpp with GPU offload (NGL=35) | 1–2 |
| llama.cpp CPU-only | 1 |

Set `VISION_POOL_MAX_CONCURRENT_PER_SLOT=1` when in doubt — it is always safe and prevents GPU OOM errors.

## Related Resources

- [User Manual 39: HelixQA Guide](39-helixqa-guide.md) -- Autonomous QA using VisionEngine for screen analysis
- [User Manual 40: LLMOrchestrator Guide](40-llmorchestrator-guide.md) -- Agent management for vision-guided testing
- [User Manual 44: QA API Guide](44-qa-api-guide.md) -- REST API for programmatic QA control
- Source: `VisionEngine/README.md`, `VisionEngine/CLAUDE.md`
