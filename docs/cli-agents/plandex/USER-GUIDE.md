# Plandex User Guide

## Overview

Plandex is an open-source AI coding agent designed specifically for large projects and real-world tasks. It provides a unique sandboxed approach to AI-assisted development with support for ultra-long contexts (up to 2M tokens) and can index projects of 20M+ tokens using Tree-sitter project maps. Plandex combines models from multiple providers (Anthropic, OpenAI, Google) to achieve better results, cost efficiency, and performance than single-provider solutions.

**Key Features:**
- 2M token effective context window
- Tree-sitter based project indexing (30+ languages)
- Sandboxed change isolation with version control
- Configurable autonomy levels
- Multi-model orchestration
- Automated debugging and testing
- Git integration with automatic commits

---

## Installation Methods

### Method 1: Quick Install (Recommended)

The fastest way to install Plandex using the official install script:

```bash
curl -sL https://plandex.ai/install.sh | bash
```

This installs the `plandex` (or `pdx` alias) command to your system.

**Verify Installation:**
```bash
plandex --version
```

### Method 2: Manual Binary Install

Download the appropriate binary for your platform from the latest release:

```bash
# macOS (Intel)
curl -L -o plandex https://github.com/plandex-ai/plandex/releases/latest/download/plandex-darwin-amd64

# macOS (Apple Silicon)
curl -L -o plandex https://github.com/plandex-ai/plandex/releases/latest/download/plandex-darwin-arm64

# Linux (AMD64)
curl -L -o plandex https://github.com/plandex-ai/plandex/releases/latest/download/plandex-linux-amd64

# Make executable and move to PATH
chmod +x plandex
sudo mv plandex /usr/local/bin/
```

### Method 3: Build from Source

Requirements: Go 1.21+

```bash
git clone https://github.com/plandex-ai/plandex.git
cd plandex/app/cli
go build -ldflags "-X plandex/version.Version=$(cat version.txt)"
sudo mv plandex /usr/local/bin/
```

### Method 4: Docker Self-Hosted Server

For local/self-hosted mode with Docker:

```bash
git clone https://github.com/plandex-ai/plandex.git
cd plandex/app
./start_local.sh
```

This starts the Plandex server locally. The CLI connects to this server.

### Windows Support

Windows is supported via WSL only:

```bash
# In WSL terminal
curl -sL https://plandex.ai/install.sh | bash
```

**Note:** Plandex does not work in Windows CMD or PowerShell directly.

---

## Quick Start

### 1. Initial Setup

Navigate to your project directory:

```bash
cd /path/to/your/project
plandex
```

### 2. Sign In / Create Account

On first run, you'll be prompted to sign in:

```bash
plandex sign-in
```

Choose your hosting option:
- **Plandex Cloud**: Managed service (requires subscription)
- **Local Mode Host**: Self-hosted server (default: http://localhost:8099)

### 3. Set API Keys

```bash
# Required for OpenRouter
export OPENROUTER_API_KEY=your_key_here

# Or for direct provider access
export OPENAI_API_KEY=your_key_here
export ANTHROPIC_API_KEY=your_key_here
export GEMINI_API_KEY=your_key_here
```

### 4. Create a Plan

```bash
plandex new
# or
pdx new
```

### 5. Start Working

Enter the REPL and start interacting:

```bash
plandex
```

---

## CLI Commands Reference

### Core Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `plandex` | `pdx` | Start the REPL interface |
| `plandex new` | - | Create a new plan |
| `plandex sign-in` | - | Authenticate with Plandex |
| `plandex sign-out` | - | Sign out from Plandex |
| `plandex version` | - | Show version information |

### Plan Management

| Command | Description |
|---------|-------------|
| `plandex plans` | List all plans |
| `plandex plans --current` | Show current plan details |
| `plandex switch <plan-id>` | Switch to a different plan |
| `plandex delete-plan <plan-id>` | Delete a plan |
| `plandex rename-plan <name>` | Rename current plan |

### Context Management

| Command | Description |
|---------|-------------|
| `plandex load <files...>` | Load files into context |
| `plandex load --all` | Load all project files |
| `plandex unload <files...>` | Remove files from context |
| `plandex context` | Show current context |
| `plandex context --stats` | Show context statistics |
| `plandex update-context` | Refresh loaded files |

### Execution Commands

| Command | Description |
|---------|-------------|
| `plandex tell <prompt>` | Send a prompt to Plandex |
| `plandex tell -f <file>` | Load prompt from file |
| `plandex chat` | Enter chat mode |
| `plandex apply` | Apply pending changes |
| `plandex reject` | Reject pending changes |
| `plandex diff` | Show pending changes diff |

### Version Control

| Command | Description |
|---------|-------------|
| `plandex commits` | List plan commits |
| `plandex commit -m "msg"` | Commit current state |
| `plandex checkout <commit>` | Checkout a commit |
| `plandex branches` | List branches |
| `plandex branch <name>` | Create/switch branch |
| `plandex merge <branch>` | Merge branch into current |

### Model Management

| Command | Description |
|---------|-------------|
| `plandex models` | List available models |
| `plandex models add` | Add a custom model |
| `plandex set-model` | Configure model for roles |
| `plandex model-packs` | List model packs |

### Git Integration

| Command | Description |
|---------|-------------|
| `plandex git-commit` | Create git commit with AI message |
| `plandex git-commit --auto` | Auto-commit changes |
| `plandex git-log` | Show git integration log |

---

## REPL Commands (Interactive Mode)

When in the REPL (started with `plandex` or `pdx`), prefix commands with `\`:

```
\help              Show help
\new               Create new plan
\load <files>      Load files
\unload <files>    Unload files
\context           Show context
\tell <prompt>     Send prompt
\apply             Apply changes
\reject            Reject changes
\diff              Show diff
\commits           Show commits
\branch <name>     Manage branches
\plans             List plans
\quit              Exit REPL
```

### REPL Flags (Startup Options)

```bash
# Mode selection
plandex --chat, -c          # Start in chat mode
plandex --tell, -t          # Start in tell mode

# Autonomy levels
plandex --no-auto           # No automation (step-by-step)
plandex --basic             # Auto-continue plans only
plandex --plus              # Auto-update context, smart context, auto-commit
plandex --semi              # Semi-auto: auto-load context
plandex --full              # Full-auto: auto-apply, auto-exec, auto-debug

# Model packs
plandex --daily             # Daily driver pack (default)
plandex --reasoning         # Reasoning model for planning
plandex --strong            # Stronger models (slower, more capable)
plandex --cheap             # Cheaper models (faster, less capable)
plandex --oss               # Open source models

# Provider-specific planners
plandex --gemini-planner    # Gemini 2.5 Pro for planning
plandex --o3-planner        # OpenAI o3-medium for planning
plandex --r1-planner        # DeepSeek R1 for planning
plandex --opus-planner      # Anthropic Opus 4 for planning
```

---

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENROUTER_API_KEY` | API key for OpenRouter.ai |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `PLANDEX_BASE_DIR` | Base directory for local server files |
| `PLANDEX_HOST` | Custom server host URL |

### Configuration File

Create `~/.plandex/config.json`:

```json
{
  "server": {
    "host": "http://localhost:8099",
    "timeout": 300
  },
  "models": {
    "planner": "anthropic/claude-3.5-sonnet",
    "coder": "openai/gpt-4o",
    "reviewer": "anthropic/claude-3-haiku"
  },
  "autonomy": "semi",
  "git": {
    "auto_commit": true,
    "commit_style": "conventional"
  }
}
```

### Adding Custom Models

```bash
plandex models add
```

Interactive prompts:
1. Select provider (custom for non-built-in)
2. Enter provider name
3. Enter model name (e.g., "Llama-3.3-70B-Instruct")
4. Set model ID (or leave blank)
5. Add description (optional)
6. Enter base URL (e.g., "https://api.regolo.ai/v1")
7. Set API key environment variable name
8. Configure limits (max tokens, output tokens, etc.)

---

## Usage Examples

### Example 1: Basic Feature Implementation

```bash
# Navigate to project
cd my-project

# Start Plandex
plandex

# In REPL:
\new
\tell "Add a user authentication system with login and signup endpoints"
\apply
```

### Example 2: Large Project Navigation

```bash
# Start with full context loading
plandex --semi

# In REPL:
\load --all
\tell "Find and fix all memory leaks in the database connection handling"
```

### Example 3: Multi-Step Implementation

```bash
plandex --full

# In REPL:
\tell "Create a REST API for managing todos"
# Plandex automatically:
# 1. Plans the implementation
# 2. Generates code
# 3. Applies changes
# 4. Runs tests
# 5. Commits to git
```

### Example 4: Debugging Session

```bash
plandex

# In REPL:
\tell "Debug why the build is failing"
# Plandex will:
# 1. Analyze error output
# 2. Investigate the codebase
# 3. Propose fixes
# 4. Run build to verify
```

### Example 5: Scripting Mode

```bash
# One-shot command without REPL
plandex tell "Refactor all console.log statements to use a logger utility" --apply
```

### Example 6: Branch-Based Development

```bash
plandex

# In REPL:
\branch feature-auth
\tell "Implement OAuth2 authentication"
\commit -m "Add OAuth2 support"
\branch feature-oauth-providers
\tell "Add Google and GitHub OAuth providers"
\branch feature-auth
\merge feature-oauth-providers
```

---

## TUI / Interactive Commands

### Fuzzy File Loading

In REPL, use fuzzy matching to load files:

```
> \load user<TAB>
# Shows matching files: user.go, user_test.go, user_service.go
```

### Command Autocompletion

The REPL provides intelligent autocompletion for:
- File paths
- Commands
- Model names
- Branch names

### Diff Viewing

```
> \diff
# Shows color-coded diff of pending changes
# Navigation: j/k or arrow keys
# Quit: q
```

### Commit History Browser

```
> \commits
# Interactive commit browser
# View any commit: Enter
# Checkout: c
# Quit: q
```

---

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to Plandex server

**Solutions:**
```bash
# Check server status
curl http://localhost:8099/health

# Restart local server
cd plandex/app
./start_local.sh

# Verify firewall settings
# Port 8099 must be accessible
```

### API Key Issues

**Problem:** "API key not found" or authentication errors

**Solutions:**
```bash
# Verify environment variables
echo $OPENROUTER_API_KEY
echo $OPENAI_API_KEY

# Set in shell profile
echo 'export OPENROUTER_API_KEY=your_key' >> ~/.bashrc
source ~/.bashrc

# Or use direnv for project-specific keys
echo 'export OPENROUTER_API_KEY=your_key' > .envrc
direnv allow
```

### Context Loading Issues

**Problem:** Files not loading or context too large

**Solutions:**
```bash
# Check context size
plandex context --stats

# Load specific files instead of --all
plandex load src/auth/*.go src/api/*.go

# Use Tree-sitter maps for large projects
plandex load --map-only

# Update context manually
plandex update-context
```

### Model Errors

**Problem:** Model not responding or rate limits

**Solutions:**
```bash
# Switch model pack
plandex --cheap    # Use cheaper models

# Check model status
plandex models

# Add fallback models in config
{
  "models": {
    "planner": "anthropic/claude-3.5-sonnet",
    "planner_fallback": "openai/gpt-4o-mini"
  }
}
```

### Git Integration Issues

**Problem:** Git commits failing

**Solutions:**
```bash
# Check git status
git status

# Ensure git is initialized
git init

# Configure git user
git config user.name "Your Name"
git config user.email "your@email.com"

# Check Plandex git settings
plandex config git
```

### Windows/WSL Issues

**Problem:** Plandex not working on Windows

**Solutions:**
- Ensure you're using WSL2
- Run all commands in WSL terminal
- Install Plandex inside WSL, not Windows
- Check file permissions: `chmod +x /usr/local/bin/plandex`

### Performance Issues

**Problem:** Slow response times

**Solutions:**
```bash
# Use faster model pack
plandex --cheap

# Reduce context size
plandex unload tests/ docs/

# Enable caching
export PLANDEX_CACHE_ENABLED=true

# Use local models for sensitive files
plandex models add  # Add local Ollama model
```

### Common Error Messages

| Error | Solution |
|-------|----------|
| "Plan not found" | Run `plandex new` to create a plan |
| "Context too large" | Remove files or use Tree-sitter maps |
| "Model timeout" | Increase timeout in config or use faster model |
| "Git repository required" | Run `git init` in project root |
| "Permission denied" | Check file permissions and ownership |

### Getting Help

```bash
# Built-in help
plandex help
plandex help <command>

# All commands
plandex help --all

# Community resources
# Discord: https://discord.gg/plandex
# GitHub Issues: https://github.com/plandex-ai/plandex/issues
# Documentation: https://docs.plandex.ai
```

---

## Best Practices

1. **Start Small**: Begin with specific, focused tasks
2. **Use Branches**: Create branches for experimental changes
3. **Review Diffs**: Always review `\diff` before applying
4. **Commit Often**: Use `\commit` to save progress
5. **Load Selectively**: Only load files relevant to your task
6. **Set Autonomy Appropriately**: Use `--no-auto` for sensitive operations
7. **Monitor Costs**: Track API usage with different model packs

---

## Resources

- **GitHub:** https://github.com/plandex-ai/plandex
- **Documentation:** https://docs.plandex.ai
- **Website:** https://plandex.ai
- **Discord:** https://discord.gg/plandex
- **YouTube:** https://youtube.com/@PlandexAI

---

*Last Updated: April 2026*
