# GPT-Engineer - Architecture Documentation

## System Overview

GPT-Engineer is an open-source AI-powered code generation platform that transforms natural language descriptions into complete, executable codebases. Built in Python, it leverages Large Language Models (LLMs) to automate software development workflows.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         GPT-ENGINEER ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   User Prompt    │─────►│  GPT-Engineer    │─────►│   LLM Provider   │   │
│  │   (Natural Lang) │      │   Core Engine    │      │  (OpenAI/Anthro) │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Preprompts      │    │   Code Gen       │    │   File System    │       │
│  │  (Identity)      │    │   Pipeline       │    │   (Workspace)    │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. CLI Application Layer

**Technology Stack:**
- **Framework**: Typer (Python CLI framework)
- **Entry Points**: `gpte`, `ge`, `gpt-engineer`, `bench`
- **Features**:
  - Natural language prompt processing
  - Multiple generation modes (default, clarify, lite, improve)
  - Self-healing execution
  - Vision support for image inputs

**Key Files:**
```
gpt_engineer/applications/cli/
├── main.py              # Main CLI entry point
├── cli_agent.py         # CLI agent implementation
├── file_selector.py     # Interactive file selection
├── learning.py          # User feedback collection
└── collect.py           # Human review collection
```

### 2. AI Module

The AI module provides a unified interface to various LLM providers.

```
┌─────────────────────────────────────────────────────────────┐
│                      AI MODULE                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   OpenAI    │  │  Anthropic  │  │   Azure     │         │
│  │   (GPT-4)   │  │  (Claude)   │  │  OpenAI     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Local LLM  │  │ Open Router │  │  Clipboard  │         │
│  │ (llama.cpp) │  │  (Proxy)    │  │  (Manual)   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Rate Limit Handling**: Exponential backoff with 7 retries (45s max)
- **Streaming Support**: Real-time response streaming
- **Message Serialization**: JSON-based conversation persistence
- **Token Usage Tracking**: Cost calculation and logging

**Implementation:** `gpt_engineer/core/ai.py`

### 3. Code Generation Pipeline

**Generation Modes:**

| Mode | Flag | Description | Use Case |
|------|------|-------------|----------|
| **Default** | (none) | Full pipeline with specification | New projects |
| **Clarify** | `-c` | Ask clarifying questions before coding | Complex requirements |
| **Lite** | `-l` | Skip preprompts, direct generation | Simple tasks |
| **Improve** | `-i` | Modify existing codebase | Refactoring |
| **Self-Heal** | `-sh` | Auto-fix execution failures | Reliable execution |

**Pipeline Flow:**
```
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌─────────┐
│  Prompt │───►│ Preprom- │───►│  Code    │───►│  Entry   │───►│  Output │
│  Input  │    │ pts/App  │    │  Gen     │    │  Point   │    │  Files  │
└─────────┘    └──────────┘    └──────────┘    └──────────┘    └─────────┘
      │              │               │               │               │
      ▼              ▼               ▼               ▼               ▼
  Natural      AI Identity      LLM Calls      Execute/       Generated
  Language     System Prompt   File Gen       Test           Workspace
```

### 4. Preprompts System

**Purpose**: Define the "identity" of the AI agent and guide code generation.

**Available Preprompts:**
```
gpt_engineer/preprompts/
├── clarify          # Specification clarification questions
├── generate         # Code generation instructions
├── improve          # Code improvement instructions
├── file_format      # Output file format specification
├── file_format_diff # Diff output format
├── file_format_fix  # Fix output format
├── entrypoint       # Entry point execution instructions
├── philosophy       # AI behavior philosophy
└── roadmap          # Development roadmap
```

**Customization:**
```bash
# Use custom preprompts
--use-custom-preprompts
```

Copies default preprompts to project directory for editing.

### 5. File System Management

**Components:**

| Component | Purpose | Location |
|-----------|---------|----------|
| `DiskMemory` | Conversation persistence | `memory/` |
| `FileStore` | Generated code storage | `workspace/` |
| `FilesDict` | In-memory file representation | Runtime |
| `DiskExecutionEnv` | Code execution environment | System |

**Project Structure:**
```
project_folder/
├── prompt              # User requirements (input)
├── preprompts/         # Custom AI identity (optional)
├── memory/             # Conversation history
│   └── logs/           # Debug logs
├── workspace/          # Generated code (output)
└── .git/               # Version control
```

### 6. Diff Processing

**Purpose**: Handle code modifications in improve mode.

**Process:**
```
┌─────────────────────────────────────────────────────────────┐
│                     DIFF PROCESSING                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Parse unified diff format from LLM                     │
│  2. Extract hunks with line numbers                        │
│  3. Apply changes to FilesDict                             │
│  4. Show colored diff to user                              │
│  5. Apply changes on approval                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:** `gpt_engineer/core/diff.py`, `gpt_engineer/core/chat_to_files.py`

### 7. Benchmarking System

**Purpose**: Evaluate agent performance against standard datasets.

**Supported Benchmarks:**
- **APPS** (Automated Programming Progress Standard)
- **MBPP** (Mostly Basic Python Problems)

**Usage:**
```bash
bench <path_to_agent> [bench_config.toml]
```

**Implementation:** `gpt_engineer/benchmark/`

---

## Data Flow

### New Project Generation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    NEW PROJECT LIFECYCLE                                │
└─────────────────────────────────────────────────────────────────────────┘

1. USER INPUT
   └─► Create project folder with `prompt` file
   └─► Write natural language requirements

2. CLI INITIALIZATION
   └─► Load environment (API keys)
   └─► Parse command-line arguments
   └─► Initialize AI provider

3. PROMPT PROCESSING
   └─► Load preprompts (system identity)
   └─► Combine with user prompt
   └─► Optional: Load images (vision mode)

4. CODE GENERATION
   └─► Send prompt to LLM
   └─► Parse response (chat_to_files_dict)
   └─► Extract file paths and code blocks

5. EXECUTION (Optional)
   └─► Generate entrypoint script
   └─► Execute code
   └─► Self-heal on failure (if enabled)

6. OUTPUT
   └─► Write files to workspace/
   └─► Stage to git
   └─► Show token usage/cost
   └─► Collect user feedback
```

### Improve Mode Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    IMPROVE MODE LIFECYCLE                               │
└─────────────────────────────────────────────────────────────────────────┘

1. FILE SELECTION
   └─► Interactive TOML-based file selection
   └─► Or use existing selection file

2. LINTING (Optional)
   └─► Run code linting
   └─► Identify improvement areas

3. DIFF GENERATION
   └─► Send current code + improvement prompt
   └─► Receive unified diff format
   └─► Parse diffs (parse_diffs)

4. CHANGE REVIEW
   └─► Show colored diff
   └─► User approves/rejects

5. APPLICATION
   └─► Apply approved changes
   └─► Write to workspace/
   └─► Stage to git
```

---

## LLM Provider Integration

### Supported Providers

| Provider | Setup | Notes |
|----------|-------|-------|
| **OpenAI** | `export OPENAI_API_KEY=sk-...` | Default, GPT-4/GPT-3.5 |
| **Anthropic** | `export ANTHROPIC_API_KEY=...` | Claude models |
| **Azure OpenAI** | `--azure https://...` | Enterprise deployment |
| **Local LLM** | `llama-cpp-python` | CodeLlama, Mistral |
| **Open Router** | API proxy | Pay-per-token |

### Local LLM Setup

```bash
# Install llama-cpp-python
CMAKE_ARGS="-DLLAMA_BLAS=ON -DLLAMA_BLAS_VENDOR=OpenBLAS" \
  pip install llama-cpp-python

# Start server
python -m llama_cpp.server \
  --model $model_path \
  --n_batch 256 \
  --n_gpu_layers 30

# Configure environment
export OPENAI_API_BASE="http://localhost:8000/v1"
export OPENAI_API_KEY="sk-xxx"
export MODEL_NAME="CodeLLama"
export LOCAL_MODEL=true
```

---

## Security Architecture

### API Key Management

```
┌─────────────────────────────────────────────────────────────┐
│                   SECURITY LAYERS                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: Environment Variables                             │
│  ├─► OPENAI_API_KEY, ANTHROPIC_API_KEY                     │
│  ├─► .env file support                                     │
│  └─► Never logged or committed                             │
│                                                             │
│  Layer 2: Execution Safety                                  │
│  ├─► User confirmation for entrypoint execution            │
│  ├─━ Diff review before apply (improve mode)               │
│  └─► Git staging for recovery                              │
│                                                             │
│  Layer 3: Rate Limiting                                     │
│  ├─► Exponential backoff                                   │
│  ├─► 7 retries max                                         │
│  └─► 45 second timeout                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Terms of Use

- User must agree to terms by running the tool
- Human review collection (opt-in)
- Open source (MIT License)

---

## Performance Considerations

### Optimization Strategies

1. **LLM Caching**: SQLite cache for repeated prompts (`--use_cache`)
2. **Streaming**: Real-time response display
3. **Message Collapsing**: Combine consecutive messages of same type
4. **Token Tracking**: Monitor usage and costs

### Resource Management

```
Memory Usage:
├─► Conversation history (memory/)
├─► File dictionary (in-memory)
├─► LLM cache (.langchain.db)
└─► Generated workspace files

API Usage:
├─► Token-based pricing
├─► Rate limit handling
└─► Cost display on completion
```

---

## Extension Points

### Custom Steps

```python
# gpt_engineer/tools/custom_steps.py

def clarified_gen(ai, prompt, preprompts_holder):
    """Ask clarifying questions before generation"""
    ...

def lite_gen(ai, prompt, preprompts_holder):
    """Skip preprompts, direct generation"""
    ...

def self_heal(ai, execution_env, files_dict):
    """Auto-fix execution failures"""
    ...
```

### Custom Preprompts

Override default AI behavior by editing files in `preprompts/` folder:
- Change code style
- Add domain knowledge
- Modify output format

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [References](./REFERENCES.md) - External resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities

---

*Last updated: April 2025*
*Version: 0.3.1*
*Python: 3.10-3.12*
