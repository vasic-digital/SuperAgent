# Conduit User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [Configuration](#configuration)
5. [Usage Examples](#usage-examples)
6. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: Go Install
```bash
go install github.com/conduit-ai/conduit@latest
```

### Method 2: Homebrew
```bash
brew tap conduit-ai/tap
brew install conduit
```

### Method 3: Source
```bash
git clone https://github.com/conduit-ai/conduit.git
cd conduit
go build -o conduit
```

## Quick Start

```bash
# Create a pipeline
conduit init my-pipeline

# Add steps
conduit add my-pipeline --step "fetch-data"
conduit add my-pipeline --step "process-data"
conduit add my-pipeline --step "save-results"

# Run pipeline
conduit run my-pipeline
```

## CLI Commands

### Global Options
| Option | Description | Example |
|--------|-------------|---------|
| --help | Show help | `conduit --help` |
| --version | Show version | `conduit --version` |
| --config | Config file | `--config ~/.conduit.yml` |

### Command: init
**Description:** Create a new pipeline

**Usage:**
```bash
conduit init <pipeline-name>
```

### Command: add
**Description:** Add a step to pipeline

**Usage:**
```bash
conduit add <pipeline> --step <step-name> [--type <type>]
```

### Command: run
**Description:** Execute a pipeline

**Usage:**
```bash
conduit run <pipeline-name> [--watch]
```

### Command: list
**Description:** List all pipelines

**Usage:**
```bash
conduit list
```

### Command: status
**Description:** Show pipeline status

**Usage:**
```bash
conduit status <pipeline-name>
```

## Configuration

### Configuration File (YAML)

```yaml
# ~/.conduit/config.yml
pipelines:
  my-pipeline:
    steps:
      - name: fetch-data
        type: http
        config:
          url: https://api.example.com/data
      - name: process-data
        type: transform
        config:
          script: process.js
      - name: save-results
        type: output
        config:
          destination: ./output.json
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| CONDUIT_API_KEY | API key |
| CONDUIT_CONFIG | Config file path |

## Usage Examples

### Example 1: Data Processing Pipeline
```bash
conduit init etl-pipeline
conduit add etl-pipeline --step extract --type http
conduit add etl-pipeline --step transform --type script
conduit add etl-pipeline --step load --type database
conduit run etl-pipeline
```

### Example 2: CI/CD Pipeline
```bash
conduit init deploy-pipeline
conduit add deploy-pipeline --step build --type docker
conduit add deploy-pipeline --step test --type shell
conduit add deploy-pipeline --step deploy --type kubernetes
conduit run deploy-pipeline --watch
```

## Troubleshooting

### Issue: Pipeline Step Fails
**Solution:**
```bash
conduit logs <pipeline-name> --step <step-name>
```

### Issue: Config Not Found
**Solution:**
```bash
export CONDUIT_CONFIG=~/.conduit/config.yml
```

---

**Last Updated:** 2026-04-02
