# User Manual 38: DocProcessor Guide

## Overview

DocProcessor (`digital.vasic.docprocessor`) loads project documentation, extracts structured feature maps, tracks verification coverage, and builds inter-document link graphs. It powers the documentation-driven QA pipeline inside HelixQA.

## Prerequisites

- Go 1.24+
- DocProcessor source cloned: `git clone git@github.com:vasic-digital/DocProcessor.git`
- (Optional) An LLM API key for intelligent feature extraction

## Step 1: Install and Build

```bash
cd DocProcessor
go build ./...
```

Verify the build:

```bash
go run ./cmd/docprocessor --help
```

## Step 2: Configure the Environment

Copy the example env file and edit it:

```bash
cp .env.example .env
```

Key variables in `.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_DOCS_ROOT` | `./docs` | Root directory to scan for documents |
| `HELIX_DOCS_AUTO_DISCOVER` | `true` | Recursively discover documentation files |
| `HELIX_DOCS_FORMATS` | `md,yaml,html,adoc,rst` | File extensions to process |

## Step 3: Extract Features from Documentation

Run the processor against a project directory:

```bash
go run ./cmd/docprocessor /path/to/project/docs
```

DocProcessor follows a 5-stage pipeline:

1. **Load and Parse** -- Scan the tree for documentation files matching configured formats
2. **Extract Features** -- Heuristic extraction identifies features, headings, and keywords
3. **Build FeatureMap** -- Structured, queryable map with categories and platform matrix
4. **Enrich** -- If an LLM agent is configured, infer screens and generate test steps
5. **Track Coverage** -- Thread-safe per-platform verification tracking via `CoverageTracker`

## Step 4: Use LLM-Powered Extraction (Optional)

DocProcessor accepts an `LLMAgent` interface for richer extraction. When integrated with LLMOrchestrator, the agent enriches features with inferred UI screens and suggested test steps. Without an LLM, the heuristic extractor still produces valid feature maps.

## Step 5: Track Coverage

The `CoverageTracker` is thread-safe (uses `sync.RWMutex`) and supports concurrent updates from multiple platform workers:

```go
tracker := coverage.NewCoverageTracker()
tracker.MarkVerified("feature-login", "android", evidence)
report := tracker.GenerateReport()
fmt.Printf("Coverage: %.1f%%\n", report.CoveragePercent)
```

Coverage states per feature: `unverified`, `verified`, `failed`, `skipped`.

## Step 6: Export the Document Graph

DocGraph tracks inter-document links and exports in multiple formats:

```bash
# JSON export
go run ./cmd/docprocessor --export-graph=json /path/to/docs > docgraph.json

# Mermaid export (paste into markdown for visual diagrams)
go run ./cmd/docprocessor --export-graph=mermaid /path/to/docs > docgraph.mmd
```

The graph identifies orphan documents (no inbound links) and circular references.

## Step 7: Run Tests

```bash
make test          # 190+ tests across 6 types
make test-race     # Race detection for thread-safe components
make test-cover    # Coverage report
```

## Package Reference

| Package | Purpose |
|---------|---------|
| `pkg/loader` | Document loading and parsing (Markdown, YAML, HTML, AsciiDoc, RST) |
| `pkg/feature` | Feature extraction, FeatureMap, FeatureMapBuilder |
| `pkg/coverage` | Thread-safe coverage tracking with RWMutex |
| `pkg/docgraph` | Inter-document link graph with JSON/Mermaid export |
| `pkg/llm` | LLMAgent interface and prompt templates |
| `pkg/config` | Configuration from .env files |

## Troubleshooting

- **No features extracted**: Verify `HELIX_DOCS_FORMATS` includes the file extensions in your project
- **Empty graph**: Ensure documents contain cross-references (links to other docs)
- **Large file skipped**: Files exceeding 10 MB are rejected by default

## Related Resources

- [User Manual 39: HelixQA Guide](39-helixqa-guide.md) -- QA orchestration using DocProcessor feature maps
- [User Manual 40: LLMOrchestrator Guide](40-llmorchestrator-guide.md) -- Agent management for LLM-powered extraction
- Source: `DocProcessor/README.md`, `DocProcessor/CLAUDE.md`
