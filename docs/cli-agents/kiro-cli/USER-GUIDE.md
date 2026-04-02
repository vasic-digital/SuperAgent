# Kiro CLI User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Method 1: Install Script (Recommended)

The easiest way to install Kiro CLI on Linux and macOS:

```bash
curl -fsSL https://cli.kiro.dev/install | bash
```

For automated install without prompts:
```bash
curl -fsSL https://cli.kiro.dev/install | bash -s -- --no-confirm
```

### Method 2: Homebrew (macOS)

```bash
brew install kiro-cli
```

### Method 3: Direct Download (Arm Linux)

For Arm Linux distributions:

```bash
# Install prerequisites
sudo apt update
sudo apt install curl unzip -y

# Download ZIP file
curl --proto '=https' --tlsv1.2 -sSf \
  'https://desktop-release.q.us-east-1.amazonaws.com/latest/kirocli-aarch64-linux.zip' \
  -o 'kirocli.zip'

# Extract and install
unzip kirocli.zip
bash ./kirocli/install.sh
```

For musl C library compatibility (Alpine, etc.):
```bash
curl --proto '=https' --tlsv1.2 -sSf \
  'https://desktop-release.q.us-east-1.amazonaws.com/latest/kirocli-aarch64-linux-musl.zip' \
  -o 'kirocli.zip'
```

### Method 4: Build from Source

```bash
git clone https://github.com/kirodotdev/kiro-cli.git
cd kiro-cli
make build
make install
```

---

## Quick Start

```bash
# Verify installation
kiro-cli --version

# Authenticate (supports GitHub, Google, AWS Builder ID, or IAM Identity Center)
kiro-cli login

# Start interactive chat session
kiro-cli chat

# Navigate to your project and start
cd my-project
kiro-cli
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `kiro-cli --help` |
| `--version` | `-v` | Show version | `kiro-cli --version` |
| `--silent` | `-s` | Suppress status messages | `kiro-cli --silent` |

### Command: kiro-cli

**Description:** Start the interactive TUI chat session.

**Usage:**
```bash
kiro-cli [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--workspace` | string | No | Current directory | Project workspace path |

**Examples:**
```bash
# Start in current directory
kiro-cli

# Start with specific workspace
kiro-cli --workspace /path/to/project
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: kiro-cli chat

**Description:** Start an interactive chat session with Kiro.

**Usage:**
```bash
kiro-cli chat [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--resume` | flag | No | false | Resume last session |

**Examples:**
```bash
kiro-cli chat
kiro-cli chat --resume
```

### Command: kiro-cli login

**Description:** Authenticate with your AWS Builder ID or other identity provider.

**Usage:**
```bash
kiro-cli login [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--provider` | string | No | auto | Identity provider (github, google, aws) |

**Examples:**
```bash
kiro-cli login
kiro-cli login --provider github
```

### Command: kiro-cli translate

**Description:** Translate natural language instructions to executable shell commands using AI.

**Usage:**
```bash
kiro-cli translate [OPTIONS] [INPUT...]
```

**Options:**
| Option | Short | Type | Required | Default | Description |
|--------|-------|------|----------|---------|-------------|
| `--n` | `-n` | number | No | 1 | Number of completions (max 5) |
| `INPUT` | | string | Yes | - | Natural language description |

**Examples:**
```bash
kiro-cli translate "list all files in the current directory"
kiro-cli translate "find all Python files modified in the last week"
kiro-cli translate "compress all log files older than 30 days"
kiro-cli translate -n 3 "search for text in files"
```

### Command: kiro-cli doctor

**Description:** Diagnose and fix common installation and configuration issues.

**Usage:**
```bash
kiro-cli doctor [OPTIONS]
```

**Options:**
| Option | Short | Type | Required | Default | Description |
|--------|-------|------|----------|---------|-------------|
| `--all` | `-a` | flag | No | false | Run all diagnostic tests without fixes |
| `--strict` | `-s` | flag | No | false | Error on warnings |
| `--format` | `-f` | string | No | plain | Output format (plain, json, json-pretty) |

**Examples:**
```bash
kiro-cli doctor
kiro-cli doctor --all
kiro-cli doctor --strict
kiro-cli doctor --format json
```

### Command: kiro-cli update

**Description:** Update Kiro CLI to the latest version.

**Usage:**
```bash
kiro-cli update [OPTIONS]
```

**Options:**
| Option | Short | Type | Required | Default | Description |
|--------|-------|------|----------|---------|-------------|
| `--non-interactive` | `-y` | flag | No | false | Don't prompt for confirmation |
| `--relaunch-dashboard` | | flag | No | true | Relaunch dashboard after update |

**Examples:**
```bash
kiro-cli update
kiro-cli update --non-interactive
```

### Command: kiro-cli integrations

**Description:** Manage system integrations for Kiro.

**Usage:**
```bash
kiro-cli integrations [SUBCOMMAND] [OPTIONS]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| `install` | Install an integration |
| `uninstall` | Uninstall an integration |
| `reinstall` | Reinstall an integration |
| `status` | Check integration status |

**Options:**
| Option | Short | Description |
|--------|-------|-------------|
| `--silent` | `-s` | Suppress status messages |
| `--format` | `-f` | Output format (for status) |

**Examples:**
```bash
# Install kiro command router
kiro-cli integrations install kiro-command-router

# Check integration status
kiro-cli integrations status

# Uninstall silently
kiro-cli integrations uninstall --silent
```

#### Kiro Command Router (v1.26.0+)

The kiro command router routes the `kiro` command between CLI and IDE:

```bash
# Install the router
kiro-cli integrations install kiro-command-router

# Set CLI as default
kiro set-default cli

# Or set IDE as default
kiro set-default ide
```

After installation:
- `kiro` - Launches your default (CLI or IDE)
- `kiro-cli` - Always launches CLI
- `kiro ide` - Always launches IDE

### Command: kiro-cli diagnostic

**Description:** Run system diagnostics for troubleshooting.

**Usage:**
```bash
kiro-cli diagnostic [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--force` | flag | No | false | Run without launching app |

**Examples:**
```bash
kiro-cli diagnostic
kiro-cli diagnostic --force
```

---

## TUI/Interactive Commands

Once inside the Kiro CLI chat interface, use these slash commands:

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| `/help` | | Show help message | `/help` |
| `/tools` | | List available tools | `/tools` |
| `/account` | | Manage credentials and API keys | `/account` |
| `/clear` | | Clear conversation history | `/clear` |
| `/exit` | | Exit chat session | `/exit` |

### Spec-Driven Development Commands

Kiro supports Kiro-style SDD (Spec-Driven Development) specs:

| Command | Description |
|---------|-------------|
| `/kiro:spec-init` | Initialize a new spec |
| `/kiro:steering` | View project steering information |
| `/kiro:spec-requirements` | Work with requirements |
| `/kiro:validate-gap` | Validate specification gaps |

---

## Configuration

### Configuration File Format

Kiro CLI uses JSON configuration files:

**File Location:** `~/.kiro/settings/config.json`

```json
{
  "mcpServers": {
    "arm_mcp_server": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v", "/path/to/codebase:/workspace",
        "--name", "arm-mcp",
        "armlimited/arm-mcp:latest"
      ],
      "env": {},
      "timeout": 60000
    }
  },
  "preferences": {
    "theme": "dark",
    "autoComplete": true
  }
}
```

### MCP Server Configuration

To configure the Arm MCP server for Arm architecture development:

1. Pull the MCP server image:
```bash
docker pull armlimited/arm-mcp:latest
```

2. Modify `~/.kiro/settings/mcp.json`:
```json
{
  "mcpServers": {
    "arm_mcp_server": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v", "/Users/yourname/codebase:/workspace",
        "--name", "arm-mcp",
        "armlimited/arm-mcp:latest"
      ],
      "env": {},
      "timeout": 60000
    }
  }
}
```

3. Verify the MCP server:
```bash
kiro-cli chat
/tools
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `KIRO_HOME` | Kiro configuration directory |
| `KIRO_API_KEY` | API key for authentication |
| `KIRO_LOG_LEVEL` | Log level (debug, info, warn, error) |

### Configuration Locations

| Platform | Path |
|----------|------|
| Linux/macOS | `~/.kiro/` |
| Windows | `%USERPROFILE%\.kiro\` |

### Settings Files

| File | Purpose |
|------|---------|
| `~/.kiro/settings/config.json` | Main configuration |
| `~/.kiro/settings/mcp.json` | MCP server configuration |
| `~/.kiro/settings/agents/` | Custom agent definitions |

---

## API/Protocol Endpoints

### MCP (Model Context Protocol)

Kiro CLI supports MCP for extending functionality:

#### Endpoint: MCP Server Tools

**Description:** List available MCP tools.

**Command:**
```bash
/tools
```

**Response:**
```
Available tools:
- arm_mcp_server:analyze - Analyze code for Arm compatibility
- arm_mcp_server:migrate - Migrate x86 code to Arm
```

### Steering Files

Kiro uses steering files for project context:

**File:** `.kiro/steering.md`

```markdown
# Project Steering

## Architecture
- Microservices pattern
- Kubernetes deployment
- PostgreSQL for data storage

## Coding Standards
- Use TypeScript
- Follow ESLint rules
- Write unit tests for all functions
```

---

## Usage Examples

### Example 1: Basic Chat Session

```bash
# Navigate to project
cd my-project

# Start Kiro
kiro-cli

# In the chat:
What does this project do?
```

### Example 2: Translate Natural Language to Commands

```bash
# Get shell command suggestions
kiro-cli translate "find all JavaScript files modified today"
# Output: find . -name "*.js" -mtime -1

kiro-cli translate -n 3 "show git log with graph"
# Output:
# 1. git log --graph --oneline
# 2. git log --graph --decorate --oneline
# 3. git log --graph --all --oneline
```

### Example 3: Arm Development with MCP

```bash
# Configure Arm MCP server
cat > ~/.kiro/settings/mcp.json << 'EOF'
{
  "mcpServers": {
    "arm_mcp_server": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-v", "/home/user/mycode:/workspace",
        "armlimited/arm-mcp:latest"
      ],
      "timeout": 60000
    }
  }
}
EOF

# Start Kiro and use Arm tools
kiro-cli chat
/tools
# Then ask:
Analyze this codebase for Arm compatibility
```

### Example 4: Spec-Driven Development

```bash
# Initialize a spec
kiro-cli chat
/kiro:spec-init "Build a user authentication system"

# Follow the SDD workflow:
# 1. Requirements gathering
# 2. Design documentation
# 3. Task breakdown
# 4. Implementation
```

### Example 5: Command Router Setup

```bash
# Install command router
kiro-cli integrations install kiro-command-router

# Set default to CLI
kiro set-default cli

# Now 'kiro' launches CLI
kiro

# And 'kiro ide' launches IDE
kiro ide
```

---

## Troubleshooting

### Issue: "Kiro CLI app is not running" error

**Solution:** 
```bash
# Launch the app
kiro-cli launch

# Or use --force for standalone diagnostics
kiro-cli diagnostic --force
```

### Issue: Diagnostic hangs

**Solution:** Use `--force` for faster limited output:
```bash
kiro-cli diagnostic --force
```

### Issue: Permission errors

**Solution:** Run with appropriate permissions or ignore errors with `--force`.

### Issue: iTerm shell integration conflicts

**Cause:** Settings in `.bash_profile` or `.bashrc` interfere with Kiro.

**Solution:** 
1. Disable iTerm shell integration temporarily
2. Or create a separate shell profile for Kiro sessions
3. Comment out pre/post sections for kiro-cli in shell configs

### Issue: MCP server not working

**Solution:**
```bash
# Check Docker is running
docker ps

# Verify MCP configuration
kiro-cli chat
/tools

# If still loading, wait and try again
```

### Issue: Shell config modification prompt

**Solution:** The installer asks about modifying shell config. To automate:
```bash
curl -fsSL https://cli.kiro.dev/install | bash -s -- --no-confirm
```

### Issue: Outdated version

**Solution:**
```bash
kiro-cli update
# Or
kiro-cli update --non-interactive
```

---

## Additional Resources

- **Website:** https://kiro.dev
- **Documentation:** https://kiro.dev/docs/cli
- **MCP Documentation:** https://kiro.dev/docs/cli/mcp
- **AWS Builder ID:** https://docs.aws.amazon.com/signin/latest/userguide/create-aws-builder-id.html
