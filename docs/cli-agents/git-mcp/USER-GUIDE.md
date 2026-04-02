# Git MCP - User Guide

> Model Context Protocol (MCP) servers for Git integration, enabling AI assistants to interact with Git repositories.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [MCP Tools](#mcp-tools)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

Git MCP servers provide AI assistants with comprehensive Git capabilities through the Model Context Protocol (MCP). These servers expose Git operations as tools that AI assistants can invoke, enabling automated version control workflows directly from your AI coding environment.

### Available Git MCP Servers

| Server | Description | Language |
|--------|-------------|----------|
| `@artik0din/mcp-git` | Complete Git CLI wrapper | Node.js |
| `git-mcp-go` | Go-based Git MCP server | Go |
| `git-workflow-mcp-server` | Workflow automation focused | Node.js |
| `git-mob-mcp-server` | Pair/mob programming support | Node.js |

### Key Features

- **Complete Git Operations**: 100% of git CLI functionality
- **Repository Management**: Clone, init, status, and configuration
- **Branch Operations**: Create, switch, merge, and delete branches
- **Commit Workflow**: Stage, commit, and amend changes
- **Remote Operations**: Push, pull, fetch from remotes
- **History & Diff**: View logs, diffs, and commit details
- **Secure**: Uses existing git configuration, no credential storage

---

## Installation

### Prerequisites

- **Git** installed and accessible in PATH
- **Node.js 16+** (for Node.js-based servers)
- **Go 1.18+** (for Go-based servers, optional)

### Installation Methods

#### Option 1: @artik0din/mcp-git (Recommended)

**Using npx (no installation required):**
```bash
npx @artik0din/mcp-git
```

**Global installation:**
```bash
npm install -g @artik0din/mcp-git
mcp-git
```

**Install and configure for Claude Desktop:**
```bash
npx @artik0din/mcp-git --setup claude
```

#### Option 2: git-mcp-go

**Using go install:**
```bash
go install github.com/geropl/git-mcp-go@latest
```

**Download prebuilt binaries:**
1. Visit [GitHub Releases](https://github.com/geropl/git-mcp-go/releases)
2. Download for your platform (Linux, macOS, Windows)
3. Extract and add to PATH

**Build from source:**
```bash
git clone https://github.com/geropl/git-mcp-go.git
cd git-mcp-go
go build -o git-mcp-go .
```

#### Option 3: git-workflow-mcp-server

```bash
git clone https://github.com/Arcia125/git-workflow-mcp-server.git
cd git-workflow-mcp-server
npm install
npm run build
```

#### Option 4: git-mob-mcp-server

```bash
npm install -g git-mob-mcp-server
```

**Prerequisite - Install git-mob:**
```bash
npm install -g git-mob
```

---

## Quick Start

### 1. Configure with Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "git": {
      "command": "npx",
      "args": ["-y", "@artik0din/mcp-git"]
    }
  }
}
```

### 2. Configure with Cursor

Add to Cursor MCP settings (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "git": {
      "command": "npx",
      "args": ["-y", "@artik0din/mcp-git"]
    }
  }
}
```

### 3. Configure with VS Code

Add to VS Code settings:

1. Open Command Palette (`Cmd/Ctrl+Shift+P`)
2. Run "MCP: List Servers"
3. Choose "Add Server" → "stdio"
4. Set command: `npx`
5. Set arguments: `["-y", "@artik0din/mcp-git"]`

### 4. Configure with GitHub Copilot CLI

```bash
# Add MCP server
/mcp add

# Fill in details:
# - Name: git
# - Command: npx
# - Args: -y @artik0din/mcp-git
```

### 5. Test the Connection

Ask your AI assistant:
```
What is the current git status of this repository?
```

The AI should use the git_status tool to check the repository status.

---

## CLI Commands

### @artik0din/mcp-git

#### Direct Execution

```bash
# Run directly
npx @artik0din/mcp-git

# With verbose output
npx @artik0din/mcp-git --verbose

# Show help
npx @artik0din/mcp-git --help
```

#### Global Installation Commands

```bash
# After: npm install -g @artik0din/mcp-git

# Run server
mcp-git

# Run with options
mcp-git --port 3000
mcp-git --transport stdio
mcp-git --transport sse
```

### git-mcp-go

```bash
# Run server
git-mcp-go

# With options
git-mcp-go --write-access  # Enable push operations
git-mcp-go --transport sse --port 8080
```

#### Available Flags

| Flag | Description |
|------|-------------|
| `--write-access` | Enable write operations (push) |
| `--transport` | Transport type: stdio or sse |
| `--port` | Port for SSE transport |
| `--help` | Show help |

### git-workflow-mcp-server

```bash
# Run from build directory
node build/index.js

# Or if installed globally
git-workflow-mcp
```

### git-mob-mcp-server

```bash
# Run server
git-mob-mcp-server

# Setup git-mob
git-mob-mcp-server setup
```

---

## MCP Tools

### @artik0din/mcp-git Tools

#### Repository Information

| Tool | Description |
|------|-------------|
| `git_status` | Show working tree status |
| `git_log` | Show commit logs |
| `git_show` | Show commit details |

#### Diff Operations

| Tool | Description |
|------|-------------|
| `git_diff` | Show changes between commits/branches |
| `git_diff_staged` | Show staged changes |
| `git_diff_unstaged` | Show unstaged changes |

#### Staging & Commits

| Tool | Description |
|------|-------------|
| `git_add` | Add files to staging area |
| `git_commit` | Record changes to repository |
| `git_reset` | Unstage changes |

#### Branch Management

| Tool | Description |
|------|-------------|
| `git_create_branch` | Create a new branch |
| `git_checkout` | Switch branches |
| `git_branch_list` | List branches |
| `git_branch_delete` | Delete branches |

#### Remote Operations

| Tool | Description |
|------|-------------|
| `git_push` | Push to remote (requires --write-access) |
| `git_pull` | Fetch and merge from remote |
| `git_fetch` | Download objects from remote |
| `git_clone` | Clone a repository |

#### Repository Setup

| Tool | Description |
|------|-------------|
| `git_init` | Initialize a new repository |
| `git_config` | Set configuration values |
| `git_list_repositories` | List available repositories |

### git-mcp-go Tools

#### Core Operations

| Tool | Description |
|------|-------------|
| `git_status` | Working tree status |
| `git_diff_unstaged` | Unstaged changes |
| `git_diff_staged` | Staged changes |
| `git_diff` | Diff between refs |
| `git_commit` | Create commit |
| `git_add` | Stage files |
| `git_reset` | Unstage all |
| `git_log` | Commit history |
| `git_create_branch` | Create branch |
| `git_checkout` | Switch branches |
| `git_show` | Show commit |
| `git_init` | Initialize repo |
| `git_push` | Push commits (with --write-access) |
| `git_list_repositories` | List repos |

### git-workflow-mcp-server Tools

| Tool | Description |
|------|-------------|
| `complete_git_workflow` | Full workflow: commit → push → PR |
| `create_branch` | Create feature branch |
| `stage_changes` | Stage all changes |
| `commit_changes` | Commit with message |
| `push_changes` | Push to remote |
| `create_pull_request` | Create GitHub PR |
| `merge_pull_request` | Merge a PR |

#### Workflow Tool Parameters

```json
{
  "commitMessage": "feat: add new feature",
  "prTitle": "Feature: New Feature",
  "prBody": "## Changes\n- Implements feature X\n- Adds tests\n\n## Testing\n- All tests pass",
  "dryRun": false
}
```

### git-mob-mcp-server Tools

| Tool | Description |
|------|-------------|
| `setup_git_mob` | Configure git-mob |
| `add_co_author` | Add team member |
| `remove_co_author` | Remove team member |
| `list_co_authors` | List team members |
| `select_co_authors` | Choose co-authors for session |
| `get_current_mob` | Show current mob |

---

## Configuration

### Claude Desktop Configuration

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Windows:**
```bash
%APPDATA%/Claude/claude_desktop_config.json
```

**Configuration:**
```json
{
  "mcpServers": {
    "git": {
      "command": "npx",
      "args": ["-y", "@artik0din/mcp-git"]
    },
    "git-advanced": {
      "command": "git-mcp-go",
      "args": ["--write-access"]
    }
  }
}
```

### Cursor Configuration

**File:** `~/.cursor/mcp.json`

```json
{
  "mcpServers": {
    "git": {
      "command": "npx",
      "args": ["-y", "@artik0din/mcp-git"]
    }
  }
}
```

### VS Code Configuration

VS Code stores MCP configuration in its settings. Use the command palette to add servers.

### GitHub Copilot CLI Configuration

Add via slash command:
```
/mcp add
```

Or edit directly:
```
~/.copilot/mcp-config.json
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GIT_MCP_VERBOSE` | Enable verbose logging |
| `GIT_MCP_WRITE_ACCESS` | Enable write operations |
| `GIT_MCP_REPO_PATH` | Default repository path |

---

## Usage Examples

### Example 1: Basic Repository Operations

```
AI: What files have been modified in this repository?
→ Uses git_status tool

AI: Show me the diff of the unstaged changes
→ Uses git_diff_unstaged tool

AI: Stage all modified files and commit with message "feat: update config"
→ Uses git_add (all) then git_commit
```

### Example 2: Branch Management

```
AI: Create a new branch called "feature/auth" from main
→ Uses git_create_branch

AI: Switch to the feature/auth branch
→ Uses git_checkout

AI: List all branches
→ Uses git_branch_list
```

### Example 3: Complete Workflow

With git-workflow-mcp-server:

```
AI: Complete the git workflow: commit "feat: add user auth", create PR titled "Add Authentication"
→ Uses complete_git_workflow tool

Parameters:
{
  "commitMessage": "feat: add user auth",
  "prTitle": "Add Authentication",
  "prBody": "## Changes\n- Add login/logout endpoints\n- Add JWT middleware\n- Add auth tests",
  "dryRun": false
}
```

### Example 4: Pair Programming

With git-mob-mcp-server:

```
AI: Setup git-mob for pair programming
→ Uses setup_git_mob

AI: Add Alice as a co-author with email alice@example.com
→ Uses add_co_author

AI: Add Bob as a co-author
→ Uses add_co_author

AI: Select Alice and Bob for this session
→ Uses select_co_authors

AI: Show current mob
→ Uses get_current_mob
```

### Example 5: History Analysis

```
AI: Show the last 10 commits
→ Uses git_log with limit

AI: Show details of commit abc123
→ Uses git_show

AI: What's the diff between main and feature-branch?
→ Uses git_diff
```

### Example 6: Dry Run Testing

```
AI: Test the git workflow without actually executing
→ Uses complete_git_workflow with dryRun: true

Result shows what would happen:
- Files to be staged: modified: src/app.js
- Commit message: feat: add feature
- Branch to push: feature/new-feature
- PR would be created: "Add new feature"
```

---

## Troubleshooting

### Installation Issues

#### "npx command not found"

```bash
# Ensure Node.js is installed
node --version

# Install Node.js from https://nodejs.org/
```

#### "git command not found"

```bash
# Install git
# macOS:
brew install git

# Ubuntu/Debian:
sudo apt-get install git

# Windows:
# Download from https://git-scm.com/
```

#### "Permission denied when installing globally"

```bash
# Use npx without global install
npx -y @artik0din/mcp-git

# Or fix npm permissions
sudo chown -R $(whoami) ~/.npm
```

### Runtime Issues

#### "Not a git repository"

```bash
# Ensure you're in a git repository
git status

# Or initialize one
git init
```

#### "Git configuration missing"

```bash
# Configure git
git config --global user.name "Your Name"
git config --global user.email "your@email.com"
```

#### "Push failed" (with git-mcp-go)

```bash
# Enable write access
git-mcp-go --write-access

# Or in config, add --write-access to args
```

#### "Authentication failed"

```bash
# Ensure git credentials are configured
git config --global credential.helper cache

# Or use SSH keys for authentication
```

### MCP Client Issues

#### "MCP server not connecting"

1. Verify the command in configuration:
   ```bash
   npx -y @artik0din/mcp-git --help
   ```

2. Check for errors in client logs

3. Restart the AI assistant

#### "Tools not appearing"

1. Wait a moment for server initialization
2. Check MCP server is enabled in client
3. Verify configuration syntax

### Debug Mode

```bash
# Enable verbose logging
GIT_MCP_VERBOSE=1 npx @artik0din/mcp-git

# For git-mcp-go
git-mcp-go --verbose
```

### Getting Help

```bash
# Show help
npx @artik0din/mcp-git --help

# Check git installation
git --version
```

### Reporting Issues

1. Check existing issues on the respective GitHub repository
2. Include:
   - MCP server name and version
   - MCP client (Claude, Cursor, etc.)
   - Git version
   - Error messages
   - Configuration (redact sensitive data)

---

## Best Practices

### 1. Use Write Access Carefully

Only enable `--write-access` when necessary. Consider using `dryRun: true` first.

### 2. Atomic Commits

Let the AI make small, focused commits rather than large changes.

### 3. Review Before Push

Always review what will be pushed before enabling write access.

### 4. Secure Credentials

Don't store credentials in MCP configuration. Use git's credential helper.

### 5. Repository Boundaries

Be careful with `git_list_repositories` — it may expose paths you don't intend to share.

---

## Resources

- **@artik0din/mcp-git**: https://github.com/artik0din/mcp-git
- **git-mcp-go**: https://github.com/geropl/git-mcp-go
- **git-workflow-mcp-server**: https://github.com/Arcia125/git-workflow-mcp-server
- **git-mob-mcp-server**: https://github.com/Mubashwer/git-mob-mcp-server
- **MCP Documentation**: https://modelcontextprotocol.io

---

*Last updated: 2026-04-02*
