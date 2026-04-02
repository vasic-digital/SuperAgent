# Amazon Q Developer CLI User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: AWS CLI (Recommended)
```bash
aws --version  # Ensure AWS CLI v2+
aws qdeveloper install
```

### Method 2: Direct Download
```bash
curl -fsSL https://desktop-release.q.us-east-1.amazonaws.com/latest/amazon-q.zip -o amazon-q.zip
unzip amazon-q.zip
sudo mv amazon-q /usr/local/bin/
```

### Method 3: Homebrew
```bash
brew install --cask amazon-q
```

## Quick Start

```bash
# Authenticate
q login

# Start chat
q chat

# Get help with a command
q explain "git rebase -i HEAD~3"

# Transform natural language to CLI
q translate "find all Python files modified today"
```

## CLI Commands

### Global Options
| Option | Description | Example |
|--------|-------------|---------|
| --help | Show help | `q --help` |
| --version | Show version | `q --version` |
| --profile | AWS profile | `--profile dev` |
| --region | AWS region | `--region us-east-1` |

### Command: login
**Description:** Authenticate with AWS

**Usage:**
```bash
q login [--profile PROFILE]
```

### Command: chat
**Description:** Start interactive chat session

**Usage:**
```bash
q chat [options]
```

**Options:**
| Option | Description |
|--------|-------------|
| --context | Set context directory |
| --no-context | Disable context |

### Command: explain
**Description:** Explain a shell command

**Usage:**
```bash
q explain "command to explain"
```

**Example:**
```bash
q explain "docker compose up -d --build"
```

### Command: translate
**Description:** Natural language to shell command

**Usage:**
```bash
q translate "what you want to do"
```

**Example:**
```bash
q translate "list all processes using port 8080"
# Output: lsof -i :8080
```

### Command: install
**Description:** Install Amazon Q

**Usage:**
```bash
q install [--force]
```

### Command: uninstall
**Description:** Uninstall Amazon Q

**Usage:**
```bash
q uninstall
```

## TUI/Interactive Commands

In interactive chat mode (`q chat`):

| Command | Description |
|---------|-------------|
| /help | Show available commands |
| /exit | Exit chat |
| /clear | Clear conversation |
| /context | Show current context |
| /explain | Explain selected code |
| /transform | Transform code |
| /test | Generate tests |
| /doc | Generate documentation |

## Configuration

### Configuration File Format (JSON)

```json
{
  "profile": "default",
  "region": "us-east-1",
  "chat": {
    "enableContext": true,
    "responseTimeout": 30
  },
  "features": {
    "inlineSuggestions": true,
    "chat": true,
    "transform": true
  }
}
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| AWS_PROFILE | AWS profile | No |
| AWS_REGION | AWS region | No |
| AWS_ACCESS_KEY_ID | Access key | For auth |
| AWS_SECRET_ACCESS_KEY | Secret key | For auth |
| Q_ENABLE_CHAT | Enable chat | No |

### Configuration Locations
1. `~/.aws/q/config.json`
2. `~/.aws/config`
3. Environment variables

## Usage Examples

### Example 1: Explain Command
```bash
q explain "kubectl get pods --all-namespaces"
```
Output explains what each part of the command does.

### Example 2: Generate Command
```bash
q translate "backup all databases to S3"
# Suggests: aws rds describe-db-instances ...
```

### Example 3: Interactive Coding
```bash
q chat
> Create a Python script to parse JSON
> Add error handling
> Save it to parser.py
```

### Example 4: Code Transformation
```bash
q chat
/transform
# Paste code and ask to convert from Java to Python
```

## Troubleshooting

### Issue: Not Authenticated
**Solution:**
```bash
q login
# Or configure AWS credentials:
aws configure
```

### Issue: Chat Not Available
**Solution:**
```bash
export Q_ENABLE_CHAT=true
q chat
```

### Issue: Slow Responses
**Solution:**
```bash
# Check AWS region
q --region us-west-2 chat
```

### Issue: Context Not Working
**Solution:**
```bash
# Ensure you're in a git repository
cd /path/to/git/repo
q chat
```

---

**Last Updated:** 2026-04-02
