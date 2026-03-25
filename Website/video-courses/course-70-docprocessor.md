# Video Course 70: DocProcessor Deep Dive

## Course Overview

**Duration:** 2.5 hours
**Level:** Intermediate to Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 06 (Testing Strategies)

Learn how DocProcessor (`digital.vasic.docprocessor`) loads project documentation, extracts structured feature maps, tracks verification coverage, and builds inter-document link graphs. This course covers the full processing pipeline from filesystem scan to DocGraph visualization, including LLM-powered feature extraction and integration with HelixQA autonomous sessions.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Describe the DocProcessor architecture and its 6-package design
2. Build and query a FeatureMap from project documentation
3. Track verification coverage per platform with the CoverageTracker
4. Visualize inter-document relationships with DocGraph and Mermaid export
5. Integrate an LLM agent for intelligent feature extraction
6. Run the complete pipeline from CLI and validate results in tests

---

## Module 1: Architecture Overview (25 min)

### Video 1.1: The Processing Pipeline (10 min)

**Topics:**
- The 6-stage pipeline: Scan, Load, Extract, Build, Enrich, Track
- Package map: `loader`, `feature`, `coverage`, `docgraph`, `llm`, `config`
- How DocProcessor fits into the HelixQA ecosystem
- Standalone CLI usage via `cmd/docprocessor`

**Pipeline:**
```
Filesystem Scan --> Loader --> Parse Sections --> LLM or Heuristic Extraction
    --> FeatureMapBuilder --> FeatureMap --> CoverageTracker / DocGraph
```

### Video 1.2: Configuration and Document Formats (15 min)

**Topics:**
- `.env` configuration: `HELIX_DOCS_ROOT`, `HELIX_DOCS_FORMATS`
- Supported formats: Markdown, YAML, HTML, AsciiDoc, reStructuredText
- MaxFileSize limit (10 MB) and format-specific parsers
- The `Config` struct and how settings flow into the pipeline

**Config Example:**
```bash
HELIX_DOCS_ROOT=./docs
HELIX_DOCS_FORMATS=md,yaml,html,adoc,rst
HELIX_DOCS_LLM_ENABLED=true
```

---

## Module 2: Feature Map Extraction (30 min)

### Video 2.1: Heuristic vs LLM Extraction (15 min)

**Topics:**
- `FeatureMapBuilder` operates in two modes: heuristic and LLM-powered
- Heuristic mode: keyword scoring on headings and bodies to assign categories
- LLM mode: structured prompts to an injected `llm.LLMAgent` returning `RawFeature` slices
- 8 feature categories: format, ui, network, settings, storage, auth, editor, other
- Deterministic feature IDs via `GenerateID(name)` for deduplication

**Key Types:**
```go
type RawFeature struct {
    Name        string
    Description string
    Category    FeatureCategory
    Platform    string
}

type Feature struct {
    ID          string
    Name        string
    Description string
    Category    FeatureCategory
    Screens     []string
    TestSteps   []TestStep
}
```

### Video 2.2: Building and Querying a FeatureMap (15 min)

**Topics:**
- Assembling extracted features into a queryable `FeatureMap`
- Platform matrix: per-feature availability across android, web, desktop
- Filtering by category, platform, and test coverage status
- The `TestStep` struct: order, action, expected result, screen ID

---

## Module 3: Coverage Tracking (25 min)

### Video 3.1: CoverageTracker Design (10 min)

**Topics:**
- Thread-safe tracking with `sync.RWMutex` (read = RLock, write = Lock)
- Per-platform verification status for each feature
- The `Evidence` type: screenshots, logs, timestamps attached to verifications
- `CoverageReport` snapshots with percentage breakdowns

### Video 3.2: Recording Evidence and Generating Reports (15 min)

**Topics:**
- Recording a verification result with evidence artifacts
- Tracking `Issue` records for failed verifications
- Producing `CoverageReport` snapshots on demand
- Integrating with HelixQA session coordinators for live coverage

**Pattern:**
```go
tracker := coverage.NewCoverageTracker()
tracker.RecordVerification("feature-login", "android", coverage.StatusPass,
    &coverage.Evidence{Screenshot: screenshotBytes, Timestamp: time.Now()})
report := tracker.Report()
fmt.Printf("Android coverage: %.1f%%\n", report.PlatformCoverage["android"]*100)
```

---

## Module 4: DocGraph Visualization (25 min)

### Video 4.1: Building the Document Link Graph (10 min)

**Topics:**
- `DocGraph` is a directed graph of inter-document references
- Nodes represent documents; edges represent cross-references
- Thread-safe with `sync.RWMutex` for concurrent access
- Automatic node creation when adding edges between unknown documents

**Core Types:**
```go
type Node struct {
    ID    string `json:"id"`
    Title string `json:"title"`
}

type Edge struct {
    From string `json:"from"`
    To   string `json:"to"`
}
```

### Video 4.2: JSON and Mermaid Export (15 min)

**Topics:**
- `ExportJSON()` for machine-readable graph snapshots
- `ExportMermaid()` for human-readable diagrams in documentation
- Integrating DocGraph output into CI reports and dashboards
- Detecting orphaned documents (nodes with no incoming edges)

---

## Module 5: LLM Agent Integration (20 min)

### Video 5.1: The LLMAgent Interface (10 min)

**Topics:**
- `llm.LLMAgent` is an injected interface with no module-level dependencies
- Prompt templates for structured feature extraction
- Response parsing: JSON deserialization of `[]RawFeature`
- Graceful degradation to heuristic mode when no agent is available

### Video 5.2: Connecting to LLMOrchestrator (10 min)

**Topics:**
- Wrapping an LLMOrchestrator agent as a `llm.LLMAgent`
- Using the adapter pattern to bridge module boundaries
- Configuring extraction prompts for domain-specific documentation
- Performance considerations: batch processing vs per-document calls

---

## Module 6: Hands-On Lab (15 min)

### Lab 1: Extract Features from a Sample Project (5 min)

**Objective:** Run DocProcessor on a sample documentation tree and inspect the FeatureMap.

**Steps:**
1. Create a `docs/` directory with 5 Markdown files covering different features
2. Configure `.env` with `HELIX_DOCS_ROOT=./docs` and `HELIX_DOCS_FORMATS=md`
3. Run `go run ./cmd/docprocessor` and inspect the output FeatureMap
4. Verify feature IDs are deterministic by running twice

### Lab 2: Track Coverage and Generate a Report (5 min)

**Objective:** Record verification evidence and produce a CoverageReport.

**Steps:**
1. Load the FeatureMap from Lab 1
2. Record 3 pass and 2 fail verifications with evidence
3. Generate a CoverageReport and verify platform percentages
4. Attach screenshot evidence to one verification

### Lab 3: Build and Export a DocGraph (5 min)

**Objective:** Build a DocGraph from cross-references and export as Mermaid.

**Steps:**
1. Parse cross-references from the sample documentation
2. Build a DocGraph with nodes and edges
3. Export as JSON and verify structure
4. Export as Mermaid and render in a Markdown preview

---

## Assessment

### Quiz (10 questions)

1. What are the 6 packages in DocProcessor and their roles?
2. How does `GenerateID` ensure deterministic feature IDs?
3. What is the MaxFileSize limit and why does it exist?
4. How does CoverageTracker achieve thread safety?
5. What are the two extraction modes in FeatureMapBuilder?
6. What formats does DocProcessor support for document loading?
7. How does DocGraph handle edges between nodes that do not exist yet?
8. What is the purpose of the `Evidence` type in coverage tracking?
9. How does the `llm.LLMAgent` interface avoid module-level dependencies?
10. What export formats does DocGraph support?

### Practical Assessment

Build a complete DocProcessor pipeline:
1. Load documentation from a multi-format project (Markdown + YAML)
2. Extract features using heuristic mode
3. Build a DocGraph from cross-references
4. Track coverage for 3 platforms
5. Export a CoverageReport and a Mermaid diagram

Deliverables:
1. Go code for the complete pipeline
2. Test suite validating extraction, coverage, and graph export
3. Mermaid diagram of the document link graph
4. CoverageReport showing per-platform percentages

---

## Resources

- [DocProcessor Architecture](../../DocProcessor/docs/architecture.md)
- [DocProcessor CLAUDE.md](../../DocProcessor/CLAUDE.md)
- [Feature Package Source](../../DocProcessor/pkg/feature/feature.go)
- [DocGraph Package Source](../../DocProcessor/pkg/docgraph/graph.go)
- [Coverage Package Source](../../DocProcessor/pkg/coverage/)
