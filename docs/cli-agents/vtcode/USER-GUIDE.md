# VT Code User Guide

## Overview

VT Code is a Rust-based terminal coding agent with semantic code intelligence. It provides deep code understanding through Tree-sitter parsing, integrates with multiple LLM providers, and supports MCP (Model Context Protocol) servers. VT Code focuses on predictable tool permissions, robust transport controls, and secure execution environments.

**Key Features:**
- Rust-based high-performance CLI
- Tree-sitter semantic code parsing
- Multi-provider LLM support (OpenAI, Anthropic, Gemini, DeepSeek, Ollama)
- MCP server integration
- OAuth 2.0 authentication
- Human-in-the-loop controls
- Workspace trust levels
- Tool policies (allow/deny/prompt)
- Lifecycle hooks
- IDE integration (VS Code extension)

---

## Installation Methods

### Method 1: Cargo (Recommended)

Requirements: Rust 1.75+

```bash
# Install from crates.io
cargo install vtcode

# Or install specific version
cargo install vtcode --version 0.96.12
```

### Method 2: Homebrew (macOS/Linux)

```bash
# Tap and install
brew tap vinhnx/vtcode
brew install vtcode

# Or directly
brew install vtcode
```

### Method 3: NPM

```bash
# Install via npm
npm install -g vtcode-ai

# Or
npm install -g @vinhnx/vtcode
```

### Method 4: Clone and Build

```bash
# Clone repository
git clone https://github.com/vinhnx/vtcode.git
cd vtcode

# Build release
cargo build --release

# Install binary
sudo cp target/release/vtcode /usr/local/bin/
```

### Method 5: Docker

```bash
# Pull image
docker pull vinhnx/vtcode:latest

# Run
docker run -it --rm \
  -v $(pwd):/workspace \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  vinhnx/vtcode:latest
```

### Method 6: VS Code Extension

```bash
# Install from VS Code Marketplace
# Search: "VTCode Companion"

# Or from CLI
code --install-extension NguyenXuanVinh.vtcode-companion
```

### Verify Installation

```bash
# Check version
vtcode --version

# Show help
vtcode --help

# Check health
vtcode doctor
```

---

## Quick Start

### 1. Configure API Keys

```bash
# Required for at least one provider
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GEMINI_API_KEY="..."
```

### 2. Initialize Configuration

```bash
# Create vtcode.toml in project root
vtcode init
```

### 3. First Use

```bash
# Ask a question
vtcode ask "How does the authentication flow work in this codebase?"

# Or execute a task
vtcode exec "Refactor the user service to use dependency injection"
```

---

## CLI Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `vtcode` | Start interactive session |
| `vtcode --version` | Show version |
| `vtcode --help` | Show help |
| `vtcode init` | Initialize configuration |
| `vtcode doctor` | Check installation health |

### Query Commands

| Command | Description |
|---------|-------------|
| `vtcode ask <question>` | Ask a question about code |
| `vtcode search <query>` | Semantic code search |
| `vtcode explain <symbol>` | Explain code symbol |
| `vtcode docs <query>` | Query documentation |

### Action Commands

| Command | Description |
|---------|-------------|
| `vtcode exec <task>` | Execute a coding task |
| `vtcode edit <file>` | Edit file with AI assistance |
| `vtcode refactor <pattern>` | Refactor code pattern |
| `vtcode test <path>` | Generate or run tests |

### Session Commands

| Command | Description |
|---------|-------------|
| `vtcode session new` | Start new session |
| `vtcode session list` | List sessions |
| `vtcode session resume <id>` | Resume session |
| `vtcode session close` | Close current session |

### Configuration Commands

| Command | Description |
|---------|-------------|
| `vtcode config get <key>` | Get config value |
| `vtcode config set <key> <val>` | Set config value |
| `vtcode config list` | List all config |
| `vtcode config edit` | Edit config file |

### MCP Commands

| Command | Description |
|---------|-------------|
| `vtcode mcp list` | List MCP servers |
| `vtcode mcp add <name>` | Add MCP server |
| `vtcode mcp remove <name>` | Remove MCP server |
| `vtcode mcp test <name>` | Test MCP connection |

### OAuth Commands

| Command | Description |
|---------|-------------|
| `vtcode auth login <provider>` | Authenticate with provider |
| `vtcode auth logout <provider>` | Logout from provider |
| `vtcode auth status` | Check auth status |
| `vtcode auth token <provider>` | Get token info |

---

## Configuration

### Configuration File (vtcode.toml)

```toml
# Core settings
[core]
version = "0.96.12"
default_provider = "anthropic"
workspace_root = "."
session_timeout = 3600

# LLM Providers
[providers]

[providers.anthropic]
api_key = "${ANTHROPIC_API_KEY}"
model = "claude-3-sonnet-20240229"
temperature = 0.1
max_tokens = 4096
oauth_enabled = true
oauth_provider = "anthropic"

[providers.openai]
api_key = "${OPENAI_API_KEY}"
model = "gpt-4-turbo-preview"
temperature = 0.1
max_tokens = 4096

[providers.gemini]
api_key = "${GEMINI_API_KEY}"
model = "gemini-1.5-pro"
temperature = 0.1
max_tokens = 8192

[providers.ollama]
base_url = "http://localhost:11434"
model = "llama3"
temperature = 0.7

[providers.deepseek]
api_key = "${DEEPSEEK_API_KEY}"
model = "deepseek-coder"
temperature = 0.1

# Security settings
[security]
workspace_trust = "prompt"  # prompt, always, never
tool_policy = "prompt"      # allow, deny, prompt
human_in_the_loop = true
approval_timeout = 300

# Tool policies
[security.tools]
file_read = "allow"
file_write = "prompt"
file_delete = "deny"
terminal_exec = "prompt"
git_commit = "prompt"
web_search = "allow"

# MCP servers
[mcp]
enabled = true
auto_start = true

[[mcp.servers]]
name = "filesystem"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "/workspace"]
transport = "stdio"
timeout = 30

[[mcp.servers]]
name = "postgres"
command = "docker"
args = ["run", "-i", "--rm", "mcp/postgres", "postgresql://..."]
transport = "stdio"
enabled = false

# Lifecycle hooks
[hooks]
pre_session = "echo 'Starting VT Code session'"
post_session = "echo 'Session complete'"
pre_tool = ""
post_tool = ""
pre_edit = "git diff"
post_edit = "git add -A"

# Code intelligence
[intelligence]
enable_tree_sitter = true
enable_ast_grep = true
supported_languages = ["rust", "python", "javascript", "typescript", "go", "java", "swift"]
max_file_size_kb = 1024
index_enabled = true

# Performance
[performance]
context_limit = 128000
concurrent_requests = 4
request_timeout = 120
cache_enabled = true
cache_ttl = 3600

# Logging
[logging]
level = "info"
format = "pretty"
file = "~/.vtcode/vtcode.log"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `XAI_API_KEY` | xAI API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `OLLAMA_HOST` | Ollama server URL |
| `VTCODE_CONFIG` | Custom config path |
| `VTCODE_LOG_LEVEL` | Log level override |

### OAuth Configuration

```toml
[oauth]
enabled = true
default_provider = "anthropic"

[oauth.providers.anthropic]
client_id = "your-client-id"
redirect_uri = "http://localhost:8765/callback"
scopes = ["openid", "profile", "email"]

[oauth.providers.openai]
client_id = "your-client-id"
redirect_uri = "http://localhost:8765/callback"

# Token storage
[oauth.storage]
backend = "keyring"  # keyring, file, memory
```

---

## Usage Examples

### Example 1: Code Analysis

```bash
# Analyze authentication flow
vtcode ask "Explain how JWT authentication works in this codebase"

# Search for specific patterns
vtcode search "function.*authenticate"

# Explain a specific function
vtcode explain "AuthService.validateToken"
```

### Example 2: Code Refactoring

```bash
# Refactor with human approval
vtcode exec "Refactor all Promise chains to use async/await"

# Edit specific file
vtcode edit src/auth.ts "Add input validation"

# Rename symbol across codebase
vtcode refactor "rename User to Customer"
```

### Example 3: Test Generation

```bash
# Generate tests for file
vtcode test src/calculator.ts

# Generate with specific framework
vtcode test src/utils.ts --framework jest

# Run existing tests with AI analysis
vtcode test --run --analyze
```

### Example 4: Using MCP Servers

```bash
# List available MCP servers
vtcode mcp list

# Query PostgreSQL via MCP
vtcode ask "What tables exist in the database?" --use-mcp postgres

# Use filesystem MCP
vtcode exec "List all TypeScript files" --use-mcp filesystem
```

### Example 5: Multi-Provider Usage

```bash
# Use specific provider
vtcode ask "Explain this algorithm" --provider openai

# Compare responses
vtcode ask "Review this code" --compare anthropic,openai

# Fallback chain
vtcode exec "Complex refactoring" --fallback anthropic,openai,gemini
```

### Example 6: Workspace Sessions

```bash
# Start named session
vtcode session new --name "auth-refactor"

# Execute tasks in session
vtcode exec "Analyze auth module"
vtcode exec "Refactor login flow"
vtcode exec "Add session management"

# Close session
vtcode session close --name "auth-refactor"

# Resume later
vtcode session resume "auth-refactor"
```

### Example 7: VS Code Integration

```bash
# With VS Code extension installed

# Ask from command palette
# Cmd+Shift+P → "VT Code: Ask Question"

# Quick actions on selected code
# Right-click → "Explain with VT Code"
# Right-click → "Refactor with VT Code"

# Status bar integration
# Click VT Code icon for quick actions
```

### Example 8: Library Usage (Rust)

```rust
// main.rs
use vtcode_core::Agent;
use vtcode_core::config::Config;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Load configuration
    let config = Config::from_file("vtcode.toml")?;
    
    // Create agent
    let agent = Agent::new(config).await?;
    
    // Start session
    let session = agent.new_session().await?;
    
    // Execute task
    let result = session.execute("Refactor error handling").await?;
    println!("{}", result);
    
    // Close session
    session.close().await?;
    
    Ok(())
}
```

### Example 9: Piping and Redirection

```bash
# Pipe output to file
vtcode ask "Generate README" > README.md

# Process output
vtcode search "TODO" | grep -i "fix" | head -10

# Chain with other tools
cat error.log | vtcode ask "Explain these errors" --stdin

# Non-interactive mode
vtcode exec "Fix linting issues" --yes --no-progress
```

### Example 10: CI/CD Integration

```yaml
# .github/workflows/code-review.yml
name: AI Code Review

on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install VT Code
        run: cargo install vtcode
      
      - name: Code Review
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          vtcode ask "Review this PR for security issues" \
            --files $(git diff --name-only HEAD^) \
            --output-format json > review.json
```

---

## TUI / Interactive Features

### Interactive Mode

```bash
# Start interactive session
vtcode

# Commands in interactive mode:
> ask "How does this work?"
> exec "Refactor this"
> search "TODO"
> context add src/auth.ts
> context remove src/test.ts
> diff
> apply
> reject
> history
> quit
```

### Real-time Updates

```bash
# Watch mode for file changes
vtcode watch --ask "Explain changes"

# Auto-refactor on save
vtcode watch --exec "Fix linting issues"
```

### Progress Display

```bash
# Show progress for long tasks
vtcode exec "Large refactoring" --progress

# Quiet mode for scripts
vtcode exec "Task" --quiet
```

---

## Troubleshooting

### Installation Issues

**Problem:** `cargo install vtcode` fails

**Solutions:**
```bash
# Update Rust
rustup update

# Check dependencies
sudo apt-get install pkg-config libssl-dev  # Ubuntu
brew install openssl pkg-config              # macOS

# Install from git
cargo install --git https://github.com/vinhnx/vtcode
```

**Problem:** NPM install fails

**Solutions:**
```bash
# Check Node.js version
node --version  # Requires 18+

# Use specific version
npm install -g vtcode-ai@latest

# Or use npx
npx vtcode-ai --version
```

### Configuration Issues

**Problem:** Config file not found

**Solutions:**
```bash
# Initialize config
vtcode init

# Or create manually
mkdir -p ~/.vtcode
cat > ~/.vtcode/vtcode.toml << 'EOF'
[core]
default_provider = "openai"
EOF

# Check config location
vtcode config locate
```

### API Key Issues

**Problem:** "API key not configured"

**Solutions:**
```bash
# Set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Or in config
cat >> vtcode.toml << 'EOF'
[providers.anthropic]
api_key = "your-key"
EOF

# Verify
curl -H "x-api-key: $ANTHROPIC_API_KEY" \
  https://api.anthropic.com/v1/models
```

### MCP Server Issues

**Problem:** MCP server not connecting

**Solutions:**
```bash
# Test MCP connection
vtcode mcp test filesystem

# Check server logs
vtcode mcp logs filesystem

# Restart MCP server
vtcode mcp restart filesystem

# Verify in config
cat vtcode.toml | grep -A5 "mcp.servers"
```

### OAuth Issues

**Problem:** OAuth authentication fails

**Solutions:**
```bash
# Check OAuth status
vtcode auth status

# Re-authenticate
vtcode auth logout anthropic
vtcode auth login anthropic

# Check token storage
vtcode auth token anthropic

# Clear and retry
rm -rf ~/.vtcode/tokens
vtcode auth login anthropic
```

### Performance Issues

**Problem:** Slow response times

**Solutions:**
```toml
# vtcode.toml
[performance]
concurrent_requests = 4
request_timeout = 60
cache_enabled = true

[intelligence]
index_enabled = true
max_file_size_kb = 512
```

### Common Error Messages

| Error | Solution |
|-------|----------|
| "No provider configured" | Add API key to config |
| "Workspace not trusted" | Run `vtcode trust .` |
| "Tool execution denied" | Update tool_policy in config |
| "MCP server error" | Check server is running |
| "Rate limit exceeded" | Switch provider or wait |
| "File too large" | Increase max_file_size_kb |

### Getting Help

```bash
# Check health
vtcode doctor

# Verbose logging
vtcode --verbose ask "Question"

# Debug mode
RUST_LOG=debug vtcode ask "Question"

# Documentation
# https://github.com/vinhnx/vtcode

# Issues
# https://github.com/vinhnx/vtcode/issues
```

---

## Best Practices

### 1. Security
```toml
[security]
workspace_trust = "prompt"
tool_policy = "prompt"
human_in_the_loop = true
```

### 2. Project Setup
```bash
# Create .vtcode.toml in project root
vtcode init

# Add to .gitignore
echo ".vtcode/" >> .gitignore
```

### 3. Session Management
- Use named sessions for long tasks
- Close sessions when done
- Review session history

### 4. Tool Policies
- Start with "prompt" mode
- Gradually allow safe tools
- Deny dangerous operations

### 5. Provider Fallback
```toml
[core]
default_provider = "anthropic"
fallback_providers = ["openai", "gemini"]
```

---

## Resources

- **GitHub:** https://github.com/vinhnx/vtcode
- **VS Code Extension:** https://marketplace.visualstudio.com/items?itemName=NguyenXuanVinh.vtcode-companion
- **Crates.io:** https://crates.io/crates/vtcode
- **Documentation:** https://deepwiki.com/vinhnx/vtcode

---

*Last Updated: April 2026*
