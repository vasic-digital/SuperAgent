# Ollama Code

## Overview

**Ollama Code** is a TypeScript-based CLI for local AI coding using Ollama models. It provides AI-powered code completion and generation running entirely on your local machine.

**Repository:** https://github.com/tcsenpai/ollama-code

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Local AI** | Runs entirely on your machine |
| **Ollama Integration** | Uses Ollama for model management |
| **Privacy** | No data sent to external services |
| **VS Code Extension** | IDE integration available |
| **TypeScript** | Modern TypeScript codebase |

---

## Installation

```bash
# Install Ollama first
# See https://ollama.ai

# Install Ollama Code
npm install -g ollama-code
```

## Quick Start

```bash
# Make sure Ollama is running
ollama serve

# Start Ollama Code
ollama-code
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Ollama Code                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  Terminal/   │    │  Ollama      │    │  Local LLM   │   │
│  │  VS Code     │◄──►│  Code CLI    │◄──►│  (Ollama)    │   │
│  └──────────────┘    └──────────────┘    └──────────────┘   │
│                             │                                │
│                        ┌────┴────┐                          │
│                        │ Models  │                          │
│                        │ (Local) │                          │
│                        └─────────┘                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Requirements

- Ollama installed and running
- Sufficient RAM for local models

---

*Part of the HelixAgent CLI Agent Collection*
