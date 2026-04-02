# Claude Plugins - User Guide

**Claude Plugins** extend Claude Code's capabilities through a plugin marketplace system. Plugins provide skills, agents, commands, MCP servers, and hooks that enhance Claude's functionality for specific workflows, domains, and integrations.

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

---

## Installation

### Prerequisites

- Claude Code CLI installed and configured
- Node.js 18+ (for plugin installers)
- Git (for marketplace plugins)

### Method 1: Plugin Marketplace (Recommended)

Add a marketplace and install plugins:

```bash
# Add marketplace
/plugin marketplace add owner/repo

# Install plugin
/plugin install plugin-name@marketplace-name

# Example: Add official marketplace
/plugin marketplace add anthropics/claude-plugins-official

# Install from official marketplace
/plugin install typescript-lsp@claude-plugins-official
/plugin install frontend-design@claude-plugins-official
```

### Method 2: Skills CLI

Use the universal skills installer:

```bash
# Install skill package
npx skills add owner/repo

# Install specific skill
npx skills add owner/repo --skill skill-name

# Install to specific agent
npx skills add owner/repo -a claude-code
```

### Method 3: Manual Installation

Clone and copy files directly:

```bash
# Clone repository
git clone https://github.com/owner/repo.git

# Copy to Claude skills directory
cp -r repo/skills/skill-name ~/.claude/skills/

# Or for plugins with multiple components
cp -r repo/plugins/plugin-name/* ~/.claude/
```

### Method 4: Direct from GitHub

```bash
# Install directly from GitHub
/plugin install plugin-name@owner/repo
```

### Verify Installation

```bash
# List installed plugins
claude --list-plugins

# Or in-session
/plugins list
```

---

## Quick Start

### Add Your First Marketplace

```bash
# Add a curated marketplace
/plugin marketplace add trancong12102/ccc

# Browse available plugins
/plugin
```

### Install a Plugin

```bash
# Install core plugin
/plugin install ccc-core@ccc

# Install external skills
/plugin install ccc-external@ccc
```

### Use a Skill

```bash
# In Claude session, skills auto-activate based on context
> Help me design a login page

# Or invoke explicitly with $skillname
> $brainstorming Help me design a login page
```

### Update Plugins

```bash
# Update specific plugin
/plugin update plugin-name

# Update all plugins
/plugin update --all
```

---

## CLI Commands

### Marketplace Management

| Command | Description |
|---------|-------------|
| `/plugin marketplace add <repo>` | Add a plugin marketplace |
| `/plugin marketplace list` | List added marketplaces |
| `/plugin marketplace remove <name>` | Remove a marketplace |

### Plugin Management

| Command | Description |
|---------|-------------|
| `/plugin` | Browse and install plugins (interactive) |
| `/plugin install <name>@<marketplace>` | Install a plugin |
| `/plugin install <name> --force` | Force reinstall |
| `/plugin update <name>` | Update a plugin |
| `/plugin update --all` | Update all plugins |
| `/plugin remove <name>` | Remove a plugin |
| `/plugin list` | List installed plugins |

### In-Session Commands

| Command | Description |
|---------|-------------|
| `/plugins` | List and manage plugins |
| `/skills` | List available skills |
| `/skills list` | Show all skills |
| `$skillname` | Invoke specific skill |

---

## TUI/Interactive Commands

### Plugin Browser

Launch interactive plugin browser:

```bash
/plugin
```

Navigation:
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate plugins |
| `Enter` | View details |
| `i` | Install plugin |
| `q` | Quit browser |

### Skills Browser

```bash
/skills
```

Navigation:
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate skills |
| `Enter` | View skill details |
| `d` | View documentation |

---

## Configuration

### Plugin Directory Structure

Plugins are stored in `~/.claude/`:

```
~/.claude/
├── plugins/
│   └── plugin-name/
│       ├── .claude-plugin/
│       │   └── plugin.json       # Plugin manifest
│       ├── skills/
│       │   └── skill-name/
│       │       ├── SKILL.md
│       │       └── reference.md
│       ├── commands/             # Slash commands
│       ├── agents/               # Agent definitions
│       ├── hooks/                # Pre/post hooks
│       ├── .mcp.json            # MCP server config
│       └── README.md
├── skills/                       # Standalone skills
│   └── skill-name/
│       └── SKILL.md
└── commands/                     # Global commands
    └── command-name.md
```

### Plugin Manifest

`plugin.json` structure:

```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Plugin description",
  "author": "Author Name",
  "license": "MIT",
  "skills": ["skill-name"],
  "commands": ["command-name"],
  "agents": ["agent-name"],
  "mcpServers": ["server-name"],
  "hooks": {
    "pre-edit": [],
    "post-edit": []
  }
}
```

### Skill Format

`SKILL.md` structure:

```markdown
---
name: skill-name
description: "When to use this skill"
tags: ["coding", "frontend"]
---

# Skill Name

Detailed instructions for Claude when this skill is active.

## Usage

How to use this skill...

## Examples

Example prompts...
```

### MCP Configuration

`~/.claude/.mcp.json`:

```json
{
  "mcpServers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-name"]
    }
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLAUDE_PLUGINS_DIR` | Custom plugins directory |
| `CLAUDE_SKILLS_DIR` | Custom skills directory |
| `CLAUDE_COMMANDS_DIR` | Custom commands directory |

---

## Usage Examples

### Installing Official Plugins

```bash
# Add official marketplace
/plugin marketplace add anthropics/claude-plugins-official

# Install TypeScript LSP support
/plugin install typescript-lsp@claude-plugins-official

# Install Python LSP support
/plugin install pyright-lsp@claude-plugins-official

# Install Context7 documentation lookup
/plugin install context7@claude-plugins-official

# Install frontend design assistant
/plugin install frontend-design@claude-plugins-official
```

### Using Community Plugins

```bash
# Add community marketplace
/plugin marketplace add trancong12102/ccc

# Install brainstorming skill
/plugin install ccc-core@ccc

# Use brainstorming
> $brainstorming Help me design a new feature
```

### Creating Custom Commands

```bash
# Create command directory
mkdir -p ~/.claude/commands

# Create command file
cat > ~/.claude/commands/deploy.md << 'EOF'
---
description: Deploy to production
---

Run the full deployment pipeline:
1. Run tests
2. Build production bundle
3. Deploy to production
4. Verify deployment

Ask for confirmation before step 3.
EOF
```

Use in session:
```
> /deploy
```

### Managing MCP Servers

```bash
# Add MCP server config
cat > ~/.claude/.mcp.json << 'EOF'
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"]
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
EOF

# Restart Claude to load MCP servers
```

### Creating a Custom Skill

```bash
# Create skill directory
mkdir -p ~/.claude/skills/my-custom-skill

# Create SKILL.md
cat > ~/.claude/skills/my-custom-skill/SKILL.md << 'EOF'
---
name: my-custom-skill
description: "When working with our internal API"
---

# Internal API Guidelines

When working with our API:

1. Base URL: https://api.internal.company.com
2. Authentication: Bearer token in Authorization header
3. Rate limit: 100 requests/minute
4. Always handle 429 responses with exponential backoff

## Common Patterns

- Use the `getClient()` helper for authenticated requests
- Wrap API calls in try/catch with proper error handling
- Log all API errors to Sentry
EOF
```

### Plugin Development Workflow

```bash
# 1. Create plugin directory structure
mkdir -p my-plugin/{.claude-plugin,skills/my-skill}

# 2. Create manifest
cat > my-plugin/.claude-plugin/plugin.json << 'EOF'
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My custom plugin",
  "skills": ["my-skill"]
}
EOF

# 3. Create skill
cat > my-plugin/skills/my-skill/SKILL.md << 'EOF'
---
name: my-skill
description: "When to trigger this skill"
---

# My Skill

Instructions here...
EOF

# 4. Test locally
claude --plugin-dir ./my-plugin

# 5. Publish to GitHub
# Push to GitHub, then others can install with:
# /plugin marketplace add your-username/my-plugin
```

---

## Troubleshooting

### Installation Issues

#### "Marketplace not found"

```bash
# Verify repository exists
curl -s https://api.github.com/repos/owner/repo | grep "full_name"

# Check format (should be owner/repo)
/plugin marketplace add correct-owner/correct-repo
```

#### "Plugin not found in marketplace"

```bash
# List available plugins in marketplace
/plugin marketplace list

# Check plugin name spelling
/plugin install correct-name@marketplace
```

#### "Permission denied" during install

```bash
# Check directory permissions
ls -la ~/.claude/

# Fix permissions
chmod 755 ~/.claude
chmod -R 644 ~/.claude/skills/*
```

### Loading Issues

#### Plugin not appearing

```bash
# Verify installation
ls ~/.claude/plugins/

# Check manifest is valid
cat ~/.claude/plugins/plugin-name/.claude-plugin/plugin.json

# Restart Claude Code
exit
claude
```

#### Skill not activating

```bash
# Check skill exists
ls ~/.claude/skills/

# Verify SKILL.md format
cat ~/.claude/skills/skill-name/SKILL.md

# Invoke explicitly
> $skill-name Your prompt here
```

#### MCP server not connecting

```bash
# Check MCP config
ls ~/.claude/.mcp.json

# Validate JSON syntax
jq . ~/.claude/.mcp.json

# Check server is installed
which npx
npx -y @modelcontextprotocol/server-name --help
```

### Runtime Issues

#### Command not recognized

```bash
# List available commands
/commands

# Check command exists
ls ~/.claude/commands/

# Verify command format
cat ~/.claude/commands/command-name.md
```

#### Agent not spawning

```bash
# List available agents
/agents

# Check agent definition
ls ~/.claude/agents/
```

#### Hooks not executing

```bash
# Check hook configuration
ls ~/.claude/plugins/plugin-name/hooks/

# Verify hook is executable
chmod +x ~/.claude/plugins/plugin-name/hooks/pre-edit.sh
```

### Common Errors

#### "Invalid plugin.json"

```bash
# Validate JSON
jq . ~/.claude/plugins/plugin-name/.claude-plugin/plugin.json

# Check required fields: name, version, description
```

#### "Skill description missing"

```bash
# Check SKILL.md frontmatter
cat ~/.claude/skills/skill-name/SKILL.md | head -10

# Should have:
# ---
# name: skill-name
# description: "Description here"
# ---
```

### Debug Mode

```bash
# Enable plugin debugging
export CLAUDE_CODE_DEBUG=1
claude

# Check logs
ls ~/.claude/logs/
```

### Getting Help

```bash
# In-session help
/plugins help
/skills help

# Plugin documentation
# Check README in ~/.claude/plugins/plugin-name/

# Community resources
# GitHub: https://github.com/topics/claude-code-plugins
```

---

## Best Practices

1. **Curate Your Plugins**: Only install plugins you actively use
2. **Version Control**: Track custom plugins in a dotfiles repo
3. **Regular Updates**: Keep plugins updated with `/plugin update --all`
4. **Skill Documentation**: Write clear descriptions for custom skills
5. **Test Before Committing**: Verify plugin works in isolated test
6. **MCP Security**: Only install MCP servers from trusted sources
7. **Hook Performance**: Keep hooks fast to avoid slowing down Claude
8. **Namespace Carefully**: Use unique names to avoid conflicts

---

## Popular Plugin Marketplaces

| Marketplace | Description |
|-------------|-------------|
| `anthropics/claude-plugins-official` | Official Anthropic plugins |
| `trancong12102/ccc` | Curated community plugins |
| `cased/claude-code-plugins` | Skills, MCP servers, hooks |
| `microsoft/power-platform-skills` | Power Platform development |
| `datahub-project/datahub-skills` | DataHub connector development |
| `jellydn/my-ai-tools` | Comprehensive AI tools collection |
| `panaversity/agentfactory-business-plugins` | Business workflow plugins |
| `sjnims/plugin-dev` | Plugin development toolkit |

---

*Last Updated: April 2026*
