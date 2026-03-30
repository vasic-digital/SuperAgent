# Video Course 73: VisionEngine for UI Analysis

## Course Overview

**Duration:** 2.5 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 70 (DocProcessor), Course 72 (LLMOrchestrator)

Master the VisionEngine module (`digital.vasic.visionengine`), a Go library providing computer vision and LLM-based visual analysis for UI testing and navigation graph building. Learn the Analyzer interface, NavigationGraph with BFS pathfinding, LLM vision providers, optional OpenCV integration, and the FallbackProvider resilience pattern.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Use the Analyzer interface for screen analysis, comparison, and element detection
2. Build and traverse a NavigationGraph with BFS shortest-path finding
3. Configure LLM vision providers for screenshot understanding
4. Understand OpenCV integration via build tags
5. Implement the FallbackProvider pattern for multi-provider resilience
6. Integrate VisionEngine with HelixQA autonomous sessions

---

## Module 1: Analyzer Interface (25 min)

### Video 1.1: Core Analysis Capabilities (15 min)

**Topics:**
- The `Analyzer` interface: 6 methods for comprehensive visual analysis
- `AnalyzeScreen`: full analysis returning layout, elements, and text regions
- `CompareScreens`: diff two screenshots to detect visual regressions
- `DetectElements`: locate UI elements (buttons, inputs, labels) in a screenshot
- `DetectText`: OCR-style text region detection
- `IdentifyScreen`: classify which screen is currently displayed

**Analyzer Interface:**
```go
type Analyzer interface {
    AnalyzeScreen(ctx context.Context, screenshot []byte) (ScreenAnalysis, error)
    CompareScreens(ctx context.Context, before, after []byte) (ScreenDiff, error)
    DetectElements(screenshot []byte) ([]UIElement, error)
    DetectText(screenshot []byte) ([]TextRegion, error)
    IdentifyScreen(ctx context.Context, screenshot []byte) (ScreenIdentity, error)
    DetectIssues(ctx context.Context, screenshot []byte) ([]VisualIssue, error)
}
```

### Video 1.2: VideoProcessor Interface (10 min)

**Topics:**
- `VideoProcessor` for video-based analysis: frame extraction, key frames, scene changes
- `ExtractFrame` at a specific timestamp
- `ExtractKeyFrames` for summarizing video content
- `DetectSceneChanges` for identifying screen transitions in recordings
- `GenerateThumbnail` for report visualizations

---

## Module 2: NavigationGraph and BFS (30 min)

### Video 2.1: Graph Structure (15 min)

**Topics:**
- `NavigationGraph` is a directed graph tracking screen transitions
- `ScreenNode`: ID, ScreenIdentity, visited flag, timestamp
- `Transition`: from/to screen IDs with the action that triggered it
- Thread-safe with `sync.RWMutex` for concurrent graph updates
- Coverage tracking: ratio of visited to total screens

**NavigationGraph Interface:**
```go
type NavigationGraph interface {
    AddScreen(screen analyzer.ScreenIdentity) string
    AddTransition(from, to string, action analyzer.Action)
    CurrentScreen() string
    SetCurrent(screenID string)
    PathTo(targetID string) ([]Transition, error)
    UnvisitedScreens() []string
    Coverage() float64
    Export() GraphSnapshot
}
```

### Video 2.2: BFS Shortest Path (15 min)

**Topics:**
- `PathTo` uses breadth-first search to find the shortest transition sequence
- Error handling: `ErrScreenNotFound`, `ErrNoPath`, `ErrEmptyGraph`
- Self-transition prevention: `ErrSelfTransition`
- The `Transition` result: ordered list of actions to reach the target
- Practical use: HelixQA navigates to unvisited screens via shortest path

**BFS Example:**
```
Graph:  Login --> Dashboard --> Settings --> Profile
                           --> Reports

PathTo("profile") from "login":
  [login->dashboard, dashboard->settings, settings->profile]

PathTo("reports") from "login":
  [login->dashboard, dashboard->reports]
```

---

## Module 3: LLM Vision Providers (25 min)

### Video 3.1: VisionProvider Interface (10 min)

**Topics:**
- `VisionProvider` interface: Name, SupportsVision, MaxImageSize, AnalyzeImage
- Image validation: empty image check, size limit enforcement
- Prompt validation: reject empty analysis prompts
- Response parsing: structured analysis results from LLM vision APIs

### Video 3.2: Built-In Providers (15 min)

**Topics:**
- Provider implementations for major LLM vision APIs
- Qwen vision provider: HTTP client with multipart image upload
- Each provider handles API-specific request/response formats
- Configuration via environment variables: API keys, endpoints, timeouts
- Pure Go HTTP clients: no CGo dependencies for LLM providers

**Provider Configuration:**
```bash
VISION_PROVIDER=qwen
QWEN_VISION_API_KEY=sk-xxx
QWEN_VISION_ENDPOINT=https://dashscope.aliyuncs.com
VISION_MAX_IMAGE_SIZE=10485760  # 10 MB
VISION_TIMEOUT=30s
```

---

## Module 4: OpenCV Integration (20 min)

### Video 4.1: Build Tag Architecture (10 min)

**Topics:**
- Default build: OpenCV stubs, all LLM providers work without CGo
- `vision` build tag: full OpenCV/GoCV support for local image processing
- Building with OpenCV: `go build -tags vision ./...`
- Testing with OpenCV: `go test -tags vision ./... -race -count=1`
- Graceful degradation: features fall back to LLM when OpenCV unavailable

### Video 4.2: OpenCV Capabilities (10 min)

**Topics:**
- Template matching for UI element detection
- Edge detection for layout analysis
- Color histogram comparison for screen diffing
- Contour detection for bounding box extraction
- Performance advantage: local processing vs API round-trip

---

## Module 5: FallbackProvider Pattern (20 min)

### Video 5.1: Multi-Provider Resilience (10 min)

**Topics:**
- `FallbackProvider` wraps multiple `VisionProvider` instances in a chain
- If the primary provider fails, subsequent providers are tried in order
- Thread-safe with `sync.RWMutex` for concurrent access
- `SupportsVision` returns true if any provider in the chain supports vision
- `MaxImageSize` returns the minimum across all providers for safety

**FallbackProvider:**
```go
type FallbackProvider struct {
    providers []VisionProvider
    mu        sync.RWMutex
}

func NewFallbackProvider(providers ...VisionProvider) (*FallbackProvider, error) {
    if len(providers) == 0 {
        return nil, fmt.Errorf("at least one provider is required")
    }
    return &FallbackProvider{providers: providers}, nil
}
```

### Video 5.2: Configuring the Fallback Chain (10 min)

**Topics:**
- Ordering providers by reliability and latency (fastest first)
- Error propagation: last error returned when all providers fail
- Logging: record which provider succeeded for monitoring
- Combining OpenCV (local) with LLM (remote) in a single chain
- Dynamic provider management: add/remove providers at runtime

**Configuration Example:**
```go
opencv := opencv.NewProvider()   // Fast, local, limited understanding
qwen := qwen.NewProvider(cfg)    // Slower, remote, deep understanding

fallback, _ := llmvision.NewFallbackProvider(opencv, qwen)
// Tries OpenCV first, falls back to Qwen LLM vision
```

---

## Module 6: Hands-On Lab (10 min)

### Lab 1: Build a Navigation Graph (5 min)

**Objective:** Create a NavigationGraph, add screens, and find paths.

**Steps:**
1. Create a new NavigationGraph
2. Add 5 screens representing an application flow
3. Add transitions between screens with action descriptions
4. Use `PathTo` to find the shortest path between two non-adjacent screens
5. Check `UnvisitedScreens` and `Coverage` values
6. Export the graph snapshot and inspect the JSON

### Lab 2: Configure a FallbackProvider (5 min)

**Objective:** Set up a multi-provider vision chain and test failover.

**Steps:**
1. Create two mock VisionProviders: one that fails and one that succeeds
2. Build a FallbackProvider with the failing provider first
3. Call `AnalyzeImage` and verify the second provider handles the request
4. Verify `SupportsVision` and `MaxImageSize` aggregate correctly

---

## Assessment

### Quiz (10 questions)

1. What are the 6 methods on the Analyzer interface?
2. How does NavigationGraph achieve thread safety?
3. What algorithm does `PathTo` use for shortest path finding?
4. What build tag enables full OpenCV support?
5. How does FallbackProvider determine `MaxImageSize`?
6. What happens when all providers in a FallbackProvider fail?
7. What is the difference between `DetectElements` and `DetectText`?
8. How does `Coverage()` calculate the visited ratio?
9. What error is returned for a self-transition attempt?
10. Why does VisionEngine use stubs instead of conditional imports for OpenCV?

### Practical Assessment

Build a complete VisionEngine integration:
1. Configure a FallbackProvider with 2 providers
2. Build a NavigationGraph from analyzed screenshots
3. Use BFS to plan navigation to unvisited screens
4. Export the graph as JSON and visualize the transitions
5. Write tests covering fallback behavior and graph pathfinding

Deliverables:
1. FallbackProvider configuration with mock providers
2. NavigationGraph with at least 8 screens and 12 transitions
3. BFS path results for 3 different source-target pairs
4. Test suite with fallback, pathfinding, and thread safety tests

---

## Module 7: Remote Vision Pool Architecture (30 min)

### Video 7.1: VisionPool Design and Routing (15 min)

**Topics:**
- Why a pool: rate limits, GPU memory constraints, parallel test execution at scale
- `VisionPool` struct: backend slots, semaphore channels, health monitor goroutine
- Backend types: `BackendOllama` (Ollama HTTP API) and `BackendLlamaCpp` (llama.cpp server)
- Least-loaded routing: slot selection based on available semaphore tokens
- Health monitoring: per-backend liveness checks on a configurable interval; unhealthy backends removed from rotation and retried
- Prometheus metrics: per-backend request count, error rate, and latency histograms

**VisionPool Configuration:**
```bash
VISION_POOL_BACKENDS=http://gpu-host-1:11434,http://gpu-host-2:11434,http://gpu-host-3:8080
VISION_POOL_BACKEND_TYPES=ollama,ollama,llamacpp
VISION_POOL_MAX_CONCURRENT_PER_SLOT=2
VISION_POOL_MODEL=llava:13b
VISION_POOL_TIMEOUT=60s
VISION_POOL_HEALTH_CHECK_INTERVAL=30s
```

**VisionPool Code:**
```go
pool, err := llmvision.NewVisionPool(llmvision.VisionPoolConfig{
    Backends: []llmvision.BackendConfig{
        {URL: "http://gpu-host-1:11434", Type: llmvision.BackendOllama,   Model: "llava:13b"},
        {URL: "http://gpu-host-2:11434", Type: llmvision.BackendOllama,   Model: "llava:13b"},
        {URL: "http://gpu-host-3:8080",  Type: llmvision.BackendLlamaCpp, Model: "llava-v1.6"},
    },
    MaxConcurrentPerSlot: 2,
    Timeout:              60 * time.Second,
})

// VisionPool implements VisionProvider — use anywhere a provider is accepted
analysis, err := pool.AnalyzeImage(ctx, screenshotBytes, "Describe UI layout")
```

### Video 7.2: Integrating VisionPool with FallbackProvider (15 min)

**Topics:**
- Composing a pool inside a `FallbackProvider`: local backends first, cloud LLM fallback
- Graceful degradation: when all pool backends are down, FallbackProvider tries the next provider
- Thread safety: `VisionPool` is safe for concurrent use across multiple `PlatformWorker` goroutines
- Monitoring pool health at runtime: `/v1/vision/pool/status` endpoint
- Capacity planning: how to size `MaxConcurrentPerSlot` for different GPU configurations

**Pool + Fallback Composition:**
```go
// Fast local pool — tries Ollama GPU hosts first
pool, _ := llmvision.NewVisionPool(cfg)

// Cloud LLM fallback — used when pool is fully loaded or unhealthy
gemini := llmvision.NewGeminiProvider(geminCfg)

// FallbackProvider: pool → Gemini → OpenAI
fallback, _ := llmvision.NewFallbackProvider(pool, gemini, openai)
```

---

## Module 8: LlamaCpp Multi-Instance Setup (25 min)

### Video 8.1: LlamaCppDeployer Internals (15 min)

**Topics:**
- `LlamaCppDeployer`: SSH-based remote management of llama.cpp server processes
- Non-interactive operation: key-based SSH authentication, no password prompts
- Deployment steps: connect → verify binary → launch `llama-server` → poll health endpoint
- Key configuration fields: `RemoteHost`, `RemoteUser`, `SSHKeyPath`, `BinaryPath`, `ModelPath`, `MmprojPath`, `ListenPort`, `NGL`, `CtxSize`, `Threads`
- `NGL` (GPU layers): tuning GPU offload for VRAM capacity — example values for 8 GB, 16 GB, 24 GB GPUs

**Deployer Usage:**
```go
deployer := llmvision.NewLlamaCppDeployer(llmvision.DeployerConfig{
    RemoteHost: "gpu-host-3",
    RemoteUser: "ubuntu",
    SSHKeyPath: "/home/user/.ssh/id_ed25519",
    BinaryPath: "/usr/local/bin/llama-server",
    ModelPath:  "/models/llava-v1.6-mistral-7b.gguf",
    MmprojPath: "/models/llava-v1.6-mistral-7b-mmproj.gguf",
    ListenPort: 8080,
    NGL:        35,   // GPU layers; set 0 for CPU-only
    CtxSize:    4096,
    Threads:    8,
})

endpoint, err := deployer.Deploy(ctx)
// endpoint == "http://gpu-host-3:8080"
```

**Deployer methods:**

| Method | Description |
|--------|-------------|
| `Deploy(ctx)` | SSH to host, start `llama-server`, await health check |
| `Undeploy(ctx)` | Gracefully stop the remote process |
| `Status(ctx)` | Check if the remote server is running and healthy |
| `Redeploy(ctx)` | Stop, optionally update the model path, restart |

### Video 8.2: Multi-Instance Deployment Patterns (10 min)

**Topics:**
- Deploying a fleet of llama.cpp servers for a `VisionPool` using multiple `LlamaCppDeployer` instances
- Coordinating deployment in parallel: `errgroup` for concurrent SSH operations
- Model sharding: different model variants per host based on available VRAM
- Automated teardown: `defer deployer.Undeploy(ctx)` for test session lifecycle management
- Health check integration: deployer registers the endpoint with `VisionPool` only after health passes

**Fleet Deployment Pattern:**
```go
hosts := []llmvision.DeployerConfig{
    {RemoteHost: "gpu-host-1", NGL: 35, ListenPort: 8080, /* ... */},
    {RemoteHost: "gpu-host-2", NGL: 28, ListenPort: 8080, /* ... */},
    {RemoteHost: "gpu-host-3", NGL: 0,  ListenPort: 8080, /* ... */}, // CPU fallback
}

var endpoints []string
g, gctx := errgroup.WithContext(ctx)
var mu sync.Mutex

for _, cfg := range hosts {
    cfg := cfg
    g.Go(func() error {
        d := llmvision.NewLlamaCppDeployer(cfg)
        ep, err := d.Deploy(gctx)
        if err != nil { return err }
        mu.Lock()
        endpoints = append(endpoints, ep)
        mu.Unlock()
        return nil
    })
}
if err := g.Wait(); err != nil {
    log.Fatal("fleet deployment failed:", err)
}
// All endpoints are healthy — build VisionPool
```

---

## Resources

- [VisionEngine CLAUDE.md](../../VisionEngine/CLAUDE.md)
- [VisionEngine Architecture](../../VisionEngine/ARCHITECTURE.md)
- [Analyzer Interface Source](../../VisionEngine/pkg/analyzer/analyzer.go)
- [NavigationGraph Source](../../VisionEngine/pkg/graph/graph.go)
- [FallbackProvider Source](../../VisionEngine/pkg/llmvision/fallback.go)
- [User Manual 41: VisionEngine Guide](../user-manuals/41-visionengine-guide.md)
- [Course 70: DocProcessor Deep Dive](course-70-docprocessor.md)
- [Course 72: LLMOrchestrator Mastery](course-72-llmorchestrator.md)
