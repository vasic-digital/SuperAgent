# Noi User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [Configuration](#configuration)
5. [Usage Examples](#usage-examples)
6. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: npm
```bash
npm install -g noi
```

### Method 2: yarn
```bash
yarn global add noi
```

### Method 3: Source
```bash
git clone https://github.com/lencx/noi.git
cd noi
npm install
npm run build
```

## Quick Start

```bash
# Start Noi
noi

# Add an AI provider
noi add openai --key your-api-key

# Start chat
noi chat
```

## CLI Commands

### Global Options
| Option | Description | Example |
|--------|-------------|---------|
| --help | Show help | `noi --help` |
| --version | Show version | `noi --version` |

### Command: add
**Description:** Add an AI provider

**Usage:**
```bash
noi add <provider> --key <api-key>
```

### Command: chat
**Description:** Start interactive chat

**Usage:**
```bash
noi chat [--provider <provider>]
```

### Command: ask
**Description:** Single question

**Usage:**
```bash
noi ask "Your question" [--provider <provider>]
```

### Command: list
**Description:** List providers

**Usage:**
```bash
noi list
```

## Configuration

### Configuration File

```json
{
  "providers": {
    "openai": {
      "apiKey": "your-key",
      "model": "gpt-4"
    }
  },
  "defaultProvider": "openai"
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| OPENAI_API_KEY | OpenAI API key |
| NOI_CONFIG | Config path |

## Usage Examples

### Example 1: Quick Question
```bash
noi ask "What is TypeScript?"
```

### Example 2: Interactive Chat
```bash
noi chat --provider openai
```

## Troubleshooting

### Issue: Provider Not Found
**Solution:**
```bash
noi add openai --key sk-...
```

---

**Last Updated:** 2026-04-02
