# OpenAI Codex - Architecture Documentation

## System Overview

Codex CLI is a coding agent that combines an AI language model with the ability to execute code, manipulate files, and run shell commands in a sandboxed environment.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CODEX CLI ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐               │
│  │   Terminal   │◄────►│   Codex CLI  │◄────►│  OpenAI API  │               │
│  │   (User)     │      │   (Rust/TS)  │      │  (Responses) │               │
│  └──────────────┘      └──────┬───────┘      └──────────────┘               │
│                               │                                             │
│          ┌────────────────────┼────────────────────┐                        │
│          │                    │                    │                        │
│          ▼                    ▼                    ▼                        │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │   Sandbox    │    │   File Ops   │    │   Tools      │                  │
│  │  (Security)  │    │  (Read/Write)│    │  (Shell/Edit)│                  │
│  └──────────────┘    └──────────────┘    └──────────────┘                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Rust Implementation (codex-rs)

The current implementation uses Rust for performance and safety:

| Crate | Purpose | Location |
|-------|---------|----------|
| `codex-core` | Core library and business logic | `codex-rs/core/` |
| `codex-tui` | Terminal user interface | `codex-rs/tui/` |
| `codex-app-server` | App server protocol | `codex-rs/app-server/` |
| `codex-protocol` | API protocol definitions | `codex-rs/protocol/` |

### 2. Legacy TypeScript (codex-cli)

Original implementation, now deprecated but still functional:

```
codex-cli/
├── src/
│   ├── components/     # UI components
│   ├── utils/         # Utility functions
│   ├── cli.tsx        # Main entry point
│   └── ...
├── tests/             # Test suite
└── package.json       # npm configuration
```

### 3. Sandbox System

Platform-specific sandboxing:

| Platform | Technology | Configuration |
|----------|-----------|---------------|
| macOS | Apple Seatbelt | `sandbox-exec` profiles |
| Linux | Docker + iptables | Container + firewall rules |
| Windows | WSL2 | Linux sandbox via WSL2 |

### 4. Tool System

Built-in tools available to the agent:

| Tool | Purpose | Safety Level |
|------|---------|--------------|
| `shell` | Execute shell commands | Sandboxed, network-disabled |
| `file_read` | Read file contents | No restrictions |
| `file_write` | Write/create files | Within workdir only |
| `patch` | Apply text patches | Within workdir only |

## Data Flow

```
1. User Input
   └─► Parse command and options

2. Context Assembly
   ├─► Load AGENTS.md files
   ├─► Gather project context
   └─► Build conversation history

3. API Request
   └─► Send to OpenAI Responses API

4. Response Processing
   └─► Parse function calls (tools)

5. Tool Execution
   ├─► Apply sandbox restrictions
   ├─► Execute in isolated environment
   └─► Return results

6. Response Streaming
   └─► Display to user in real-time
```

## Security Architecture

### Defense in Depth

```
Layer 1: Approval Modes (User control)
Layer 2: Sandboxing (OS-level isolation)
Layer 3: Network Blocking (No outbound connections)
Layer 4: Directory Confinement (Workdir only)
Layer 5: Git Safety (Revert capability)
```

### Approval Mode Decision Tree

```
User Request
    │
    ▼
┌───────────────┐
│  Suggest Mode │───► All writes/commands require approval
│  (default)    │
└───────────────┘
    │
    ▼
┌───────────────┐
│  Auto Edit    │───► Auto-apply file patches
│               │───► Commands still require approval
└───────────────┘
    │
    ▼
┌───────────────┐
│  Full Auto    │───► Auto-execute everything
│               │───► Sandboxed, network-disabled
└───────────────┘
```

## Multi-Provider Architecture

```
┌────────────────────────────────────────────┐
│           Provider Router                  │
├────────────────────────────────────────────┤
│                                            │
│  OpenAI ◄─── Default                       │
│  Azure  ◄─── Enterprise                    │
│  Gemini ◄─── Google                        │
│  Ollama ◄─── Local                         │
│  ...    ◄─── Extensible                    │
│                                            │
│  All use OpenAI-compatible API format      │
└────────────────────────────────────────────┘
```

## Build System

### Bazel (Primary)

```bash
# Build entire workspace
bazel build //...

# Build specific target
bazel build //codex-rs:codex

# Run tests
bazel test //...
```

### Cargo (Rust Development)

```bash
# Build Rust implementation
cd codex-rs
cargo build --release

# Run tests
cargo test

# With nextest
cargo nextest run
```

### Just (Task Runner)

```bash
# Format code
just fmt

# Run lints
just lint

# Build config schema
just write-config-schema
```

---

*For usage details, see [API Reference](./API.md)*
