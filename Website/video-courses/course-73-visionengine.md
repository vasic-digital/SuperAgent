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

## Resources

- [VisionEngine CLAUDE.md](../../VisionEngine/CLAUDE.md)
- [VisionEngine Architecture](../../VisionEngine/ARCHITECTURE.md)
- [Analyzer Interface Source](../../VisionEngine/pkg/analyzer/analyzer.go)
- [NavigationGraph Source](../../VisionEngine/pkg/graph/graph.go)
- [FallbackProvider Source](../../VisionEngine/pkg/llmvision/fallback.go)
- [Course 70: DocProcessor Deep Dive](course-70-docprocessor.md)
- [Course 72: LLMOrchestrator Mastery](course-72-llmorchestrator.md)
