# Amazon Q Developer CLI User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: macOS Installer (Recommended)

```bash
# Download and install
curl -fsSL https://desktop-release.q.us-east-1.amazonaws.com/latest/Amazon%20Q.dmg -o AmazonQ.dmg
open AmazonQ.dmg
# Drag Amazon Q to Applications

# Or use Homebrew
brew install --cask amazon-q
```

### Method 2: Linux Package

```bash
# Download for Linux
curl -fsSL https://desktop-release.q.us-east-1.amazonaws.com/latest/Amazon%20Q.AppImage -o AmazonQ.AppImage
chmod +x AmazonQ.AppImage
./AmazonQ.AppImage
```

### Method 3: Windows Installer

```powershell
# Download from AWS
# https://desktop-release.q.us-east-1.amazonaws.com/latest/Amazon%20Q.exe
# Run the installer

# Or use WinGet
winget install Amazon.Q
```

### Method 4: Zip File Install

```bash
# Download zip
curl -fsSL https://desktop-release.q.us-east-1.amazonaws.com/latest/AmazonQ.zip -o AmazonQ.zip
unzip AmazonQ.zip
sudo mv AmazonQ /usr/local/bin/
```

### Prerequisites

- AWS account
- Amazon Q Developer Pro license (for full features)
- Builder ID or IAM Identity Center
- macOS 11+, Windows 10+, or Linux

## Quick Start

### First-Time Setup

```bash
# Start Amazon Q CLI
q

# Log in with your AWS account
q login --license pro --region us-east-1

# Follow the browser prompts:
# 1. Enter device code
# 2. Sign in with AWS credentials
# 3. Authorize the application
```

### Basic Usage

```bash
# Start interactive session
q

# Start with a specific model
q --model claude-sonnet-4

# Execute a command
q exec "explain this codebase"

# Get help
q --help
```

### Hello World

```bash
# Start Amazon Q
q

# At the prompt, type:
> Create a Python script that prints "Hello, World!"

# Amazon Q will create and execute the code
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `q --version` |
| --help | -h | Show help | `q --help` |
| --model | -m | Select model | `q --model claude-sonnet-4` |
| --license | | License type | `q --license pro` |
| --region | -r | AWS region | `q --region us-east-1` |
| --verbose | | Verbose output | `q --verbose` |
| --quiet | | Suppress output | `q --quiet` |

### Command: login

**Description:** Log in to Amazon Q.

**Usage:**
```bash
q login [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --license | string | No | pro | License type |
| --region | string | No | us-east-1 | AWS region |
| --profile | string | No | default | AWS profile |

**Examples:**
```bash
# Log in with Pro license
q login --license pro --region us-east-1

# Use specific profile
q login --profile development

# Log in to specific region
q login --region eu-west-1
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Authentication failed |
| 2 | Invalid credentials |

### Command: logout

**Description:** Log out from Amazon Q.

**Usage:**
```bash
q logout [OPTIONS]
```

**Examples:**
```bash
# Log out
q logout
```

### Command: status

**Description:** Check authentication and connection status.

**Usage:**
```bash
q status
```

**Examples:**
```bash
# Check status
q status
```

### Command: exec

**Description:** Execute a task non-interactively.

**Usage:**
```bash
q exec [OPTIONS] "PROMPT"
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --model | string | No | claude-sonnet-4 | Model to use |
| --output | string | No | | Output file |

**Examples:**
```bash
# Execute task
q exec "explain the codebase architecture"

# Save output
q exec --output analysis.md "analyze code quality"

# With specific model
q exec --model claude-opus-4 "complex refactoring task"
```

### Command: config

**Description:** Manage Amazon Q configuration.

**Usage:**
```bash
q config [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| get | Get config value |
| set | Set config value |
| list | List all config |

**Examples:**
```bash
# Set default model
q config set model claude-sonnet-4

# Get region
q config get region

# List all settings
q config list
```

### Command: doctor

**Description:** Run diagnostics on the environment.

**Usage:**
```bash
q doctor
```

**Examples:**
```bash
# Run diagnostics
q doctor
```

### Command: update

**Description:** Update Amazon Q CLI to latest version.

**Usage:**
```bash
q update [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --check | boolean | No | false | Check for updates only |

**Examples:**
```bash
# Update
q update

# Check for updates
q update --check
```

## TUI/Interactive Commands

When running in interactive/TUI mode, use these commands:

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| /help | ? | Show help | `/help` |
| /exit | Ctrl+D | Exit | `/exit` |
| /clear | Ctrl+L | Clear screen | `/clear` |
| /model | | Switch model | `/model claude-opus-4` |
| /tools | | Manage tools | `/tools list` |
| /tools trust | | Trust a tool | `/tools trust <tool>` |
| /mcp | | Manage MCP servers | `/mcp list` |
| /history | | Show history | `/history` |
| /context | | Show context | `/context` |

### Tool Management

| Command | Description |
|---------|-------------|
| /tools list | List available tools |
| /tools trust <tool> | Trust a tool |
| /tools untrust <tool> | Untrust a tool |
| /tools status | Show tool status |

### MCP Management

| Command | Description |
|---------|-------------|
| /mcp list | List MCP servers |
| /mcp enable <name> | Enable MCP server |
| /mcp disable <name> | Disable MCP server |
| /mcp status | Show MCP status |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Tab | Accept completion |
| Ctrl+C | Cancel operation |
| Ctrl+D | Exit |
| Ctrl+L | Clear screen |
| Up/Down | Navigate history |

## Configuration

### Configuration File Format

Amazon Q uses `~/.aws/amazonq/config.json`:

```json
{
  "model": "claude-sonnet-4-20250514",
  "license": "pro",
  "region": "us-east-1",
  "profile": "default",
  "autoComplete": true,
  "contextAwareness": true,
  "tools": {
    "trusted": ["git", "npm", "pip"],
    "untrusted": ["rm", "sudo"]
  },
  "mcp": {
    "servers": {
      "awslabs.aws-documentation-mcp-server": {
        "command": "uvx",
        "args": ["awslabs.aws-documentation-mcp-server@latest"],
        "env": {
          "FASTMCP_LOG_LEVEL": "ERROR",
          "AWS_DOCUMENTATION_PARTITION": "aws"
        }
      },
      "context7": {
        "command": "npx",
        "args": ["-y", "@upstash/context7-mcp"],
        "env": {
          "DEFAULT_MINIMUM_TOKENS": "6000"
        }
      }
    }
  },
  "keybindings": {
    "acceptSuggestion": "Tab",
    "showCompletions": "Ctrl+Space"
  }
}
```

### MCP Configuration

Create `~/.aws/amazonq/mcp.json`:

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "env": {
        "DEFAULT_MINIMUM_TOKENS": "6000"
      }
    },
    "awslabs.aws-documentation-mcp-server": {
      "command": "uvx",
      "args": ["awslabs.aws-documentation-mcp-server@latest"],
      "env": {
        "FASTMCP_LOG_LEVEL": "ERROR",
        "AWS_DOCUMENTATION_PARTITION": "aws"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    }
  }
}
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| AWS_PROFILE | AWS profile | No | `default` |
| AWS_REGION | AWS region | No | `us-east-1` |
| AWS_ACCESS_KEY_ID | Access key | Yes* | `AKIA...` |
| AWS_SECRET_ACCESS_KEY | Secret key | Yes* | `...` |
| Q_LICENSE | License type | No | `pro` |
| Q_MODEL | Default model | No | `claude-sonnet-4` |

*Required for programmatic access

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables
3. User config (`~/.aws/amazonq/config.json`)
4. MCP config (`~/.aws/amazonq/mcp.json`)
5. AWS config (`~/.aws/config`)

## API/Protocol Endpoints

### Amazon Q API

Amazon Q CLI primarily uses AWS APIs:

**Base URL:** `https://q.us-east-1.amazonaws.com`

### AWS Service Integration

Amazon Q can generate AWS CLI commands:

```bash
# In Amazon Q
> How do I list all S3 buckets?

# Amazon Q responds with:
aws s3 ls

# And can execute it with your approval
```

### CDK Integration

```bash
# Generate CDK code
q
> Create a CDK stack with an S3 bucket and Lambda function

# Deploy with CDK
> Deploy this stack to the development environment
```

## Usage Examples

### Example 1: AWS CLI Assistance

```bash
# Start Amazon Q
q

# Get AWS CLI help
> How do I create an EC2 instance?

# Amazon Q will suggest and explain:
aws ec2 run-instances --image-id ami-12345 --instance-type t2.micro

# Execute with approval
```

### Example 2: Code Development

```bash
# Start Amazon Q
q

> Create a Lambda function that processes S3 events
> Use Python with boto3
> Include error handling and logging

# Amazon Q generates the code
> Deploy this to AWS
```

### Example 3: DevOps Tasks

```bash
# Start Amazon Q
q

> Write a CloudFormation template for a VPC with public and private subnets
> Include NAT Gateway and Internet Gateway

# Review and deploy
> Deploy this CloudFormation stack
```

### Example 4: Using MCP

```bash
# Configure MCP first
cat > ~/.aws/amazonq/mcp.json << 'EOF'
{
  "mcpServers": {
    "awslabs.aws-documentation-mcp-server": {
      "command": "uvx",
      "args": ["awslabs.aws-documentation-mcp-server@latest"]
    }
  }
}
EOF

# Use in Amazon Q
q
> What are the best practices for DynamoDB partition keys?
> Use the AWS documentation MCP to find current recommendations
```

### Example 5: Multi-Agent Collaboration

```bash
# Start Amazon Q
q

> Create a multi-agent system for code review
> One agent for security checks
> One agent for performance analysis
> One agent for style review
```

## Troubleshooting

### Issue: Login Fails

**Symptoms:** "Login failed" or "Already logged in" errors

**Solution:**
```bash
# Check current status
q status

# Log out and try again
q logout
q login --license pro --region us-east-1

# For "Already logged in" error
q logout
q login

# Check keychain (macOS)
security unlock-keychain ~/Library/Keychains/login.keychain
```

### Issue: License Not Found

**Symptoms:** "No valid license" error

**Solution:**
- Ensure you have Amazon Q Developer Pro subscription
- Check AWS Console for license status
- Contact AWS support if issues persist

### Issue: MCP Servers Not Loading

**Symptoms:** MCP tools unavailable

**Solution:**
```bash
# Check MCP configuration
cat ~/.aws/amazonq/mcp.json

# Verify tools are trusted
/tools status

# Trust specific tool
/tools trust <tool-name>

# Restart Amazon Q
```

### Issue: Keychain Access Error

**Symptoms:** "User interaction is not allowed"

**Solution:**
```bash
# Unlock keychain (macOS)
security unlock-keychain ~/Library/Keychains/login.keychain

# Or from command line
security unlock-keychain -p <password> ~/Library/Keychains/login.keychain
```

### Issue: Auto-Complete Not Working

**Symptoms:** No IDE-style completions appearing

**Solution:**
```bash
# Check if enabled
q config get autoComplete

# Enable if needed
q config set autoComplete true

# Check shell integration
# For zsh, ensure plugin is loaded
source ~/.amazonq/zsh-plugin.zsh
```

### Issue: Slow Performance

**Symptoms:** Long response times

**Solution:**
```bash
# Check network connection
# Try different region
q --region us-west-2

# Use faster model
q --model claude-sonnet-4 "task"
```

### Issue: Tool Execution Blocked

**Symptoms:** Commands not executing

**Solution:**
```bash
# Check tool trust status
/tools status

# Trust the tool
/tools trust <tool-name>

# Or trust all tools (use with caution)
/tools trust-all
```

---

**Last Updated:** 2026-04-02
**Version:** 1.9.x
