# CLI Agent Integration Scripts

This directory contains scripts for generating configurations and installing plugins for all 47+ supported CLI agents to integrate with HelixAgent and LLMsVerifier.

## Overview

HelixAgent provides a unified AI Debate ensemble backend that can be used by any CLI-based AI coding assistant. These scripts automate the process of:

1. **Configuration Generation**: Creates proper config files for each CLI agent
2. **Plugin Installation**: Installs HelixAgent integration plugins
3. **Integration Testing**: Verifies all configurations and plugins work correctly

## Supported CLI Agents (47+)

### Tier 1 - Primary Support (10 agents)
| Agent | Language | Config Path | Config Format |
|-------|----------|-------------|---------------|
| Claude Code | TypeScript | `~/.claude` | JSON |
| Aider | Python | `~/.aider.conf.yml` | YAML |
| Cline | TypeScript | `~/.cline` | JSON |
| OpenCode | Go | `~/.config/opencode` | JSON |
| Kilo Code | TypeScript | `~/.kilo-code` | JSON |
| Gemini CLI | Python | `~/.config/gemini-cli` | YAML |
| Qwen Code | Python | `~/.qwen` | JSON |
| DeepSeek CLI | Python | `~/.deepseek` | JSON |
| Forge | TypeScript | `~/.forge` | JSON |
| Codename Goose | Go | `~/.config/goose` | YAML |

### Tier 2 - Secondary Support (15 agents)
| Agent | Language | Config Path | Config Format |
|-------|----------|-------------|---------------|
| Amazon Q | TypeScript | `~/.aws/amazonq` | JSON |
| Kiro | Python | `~/.kiro/steering` | YAML |
| GPT Engineer | Python | `~/.gpt-engineer` | YAML |
| Mistral Code | Python | `~/.mistral` | JSON |
| Ollama Code | Python | `~/.ollama-code` | JSON |
| Plandex | Go | `~/.plandex` | JSON |
| Codex | TypeScript | `~/.codex` | JSON |
| VTCode | TypeScript | `~/.vtcode` | JSON |
| Nanocoder | Python | `~/.nanocoder` | YAML |
| GitMCP | TypeScript | `~/.gitmcp` | JSON |
| TaskWeaver | Python | `~/.taskweaver` | YAML |
| Octogen | Python | `~/.octogen` | YAML |
| FauxPilot | Python | `~/.fauxpilot` | JSON |
| Bridle | Go | `~/.bridle` | YAML |
| Agent Deck | TypeScript | `~/.agent-deck` | JSON |

### Tier 3 - Extended Support (22 agents)
Claude Squad, Codai, Emdash, Get Shit Done, GitHub Copilot CLI, GitHub Spec Kit, GPTme, Mobile Agent, Multiagent Coding, Noi, OpenHands, Postgres MCP, Shai, SnowCLI, Superset, Warp, Cheshire Cat, Conduit, Crush, HelixCode

## Quick Start

```bash
# Generate all configurations
./generate-all-configs.sh

# Generate and install to system locations
./generate-all-configs.sh --install

# Install all plugins
./install-plugins.sh

# Run integration tests
./tests/cli-agent-integration-test.sh

# Generate config for specific agent
./generate-all-configs.sh --agent=claude_code
```

## Scripts

### 1. generate-all-configs.sh

Generates configuration files for all supported CLI agents.

**Usage:**
```bash
./generate-all-configs.sh [OPTIONS]
```

**Options:**
| Option | Description |
|--------|-------------|
| `--install` | Install configs to system locations |
| `--dry-run` | Show what would be done without making changes |
| `--agent=NAME` | Generate config for specific agent only |
| `-h, --help` | Show help message |

**Environment Variables:**
| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_AGENT_URL` | `http://localhost:8080` | HelixAgent server URL |
| `HELIX_AGENT_API_KEY` | (empty) | API key for authentication |
| `LLMS_VERIFIER_URL` | `http://localhost:8081` | LLMsVerifier server URL |

**Output:**
- Generated configs: `scripts/cli-agents/configs/generated/<agent>/`
- Backups (when installing): `scripts/cli-agents/configs/backups/<timestamp>/`

### 2. install-plugins.sh

Installs HelixAgent integration plugins for all CLI agents.

**Usage:**
```bash
./install-plugins.sh [OPTIONS]
```

**Options:**
| Option | Description |
|--------|-------------|
| `--agent=NAME` | Install plugins for specific agent only |
| `--dry-run` | Show what would be done without making changes |
| `--uninstall` | Remove installed plugins |
| `-h, --help` | Show help message |

**Generated Plugins:**
| Plugin | Description | Dependencies |
|--------|-------------|--------------|
| helix-integration | Core HelixAgent API integration | None |
| event-handler | Event subscription and handling | helix-integration |
| verifier-client | LLMsVerifier integration | None |
| debate-ui | AI Debate visualization | helix-integration |
| streaming-adapter | Streaming response adapter | helix-integration |
| mcp-bridge | MCP protocol bridge | helix-integration |

**Output:**
- Generated plugins: `scripts/cli-agents/plugins/generated/<agent>/`
- Installed plugins: `~/.helix-plugins/<agent>/`

### 3. tests/cli-agent-integration-test.sh

Runs comprehensive integration tests for all CLI agent configurations and plugins.

**Usage:**
```bash
./tests/cli-agent-integration-test.sh [OPTIONS]
```

**Options:**
| Option | Description |
|--------|-------------|
| `--verbose, -v` | Show detailed output |
| `--agent=NAME` | Test specific agent only |
| `-h, --help` | Show help message |

**Test Categories:**
1. Script existence verification
2. Configuration generation
3. Configuration content validation
4. Plugin generation
5. Plugin content verification
6. Go source code validation
7. Integration compliance

## Configuration Structure

### JSON Configuration (Example: Claude Code)
```json
{
  "version": "1.0.0",
  "providers": {
    "helix": {
      "type": "openai-compatible",
      "baseUrl": "http://localhost:8080/v1",
      "apiKey": "",
      "model": "ai-debate-ensemble",
      "features": {
        "streaming": true,
        "tools": true
      }
    }
  },
  "defaultProvider": "helix",
  "features": {
    "aiDebate": true,
    "multiPassValidation": true,
    "verifier": {
      "enabled": true,
      "url": "http://localhost:8081"
    }
  }
}
```

### YAML Configuration (Example: Aider)
```yaml
model: ai-debate-ensemble
openai-api-base: http://localhost:8080/v1
openai-api-key: ""

helix:
  enabled: true
  verifier_url: http://localhost:8081
  ai_debate: true
  multi_pass_validation: true

stream: true
auto-commits: true
auto-lint: true
```

## Plugin Structure

Each plugin follows this structure:

```
<agent>/
├── <plugin-name>/
│   ├── manifest.json       # Plugin metadata and configuration schema
│   ├── <plugin>.go         # Plugin source code
│   └── README.md           # Plugin documentation (optional)
└── index.json              # Agent plugin index
```

### Plugin Manifest Example
```json
{
  "name": "helix-integration",
  "version": "1.0.0",
  "description": "Core HelixAgent API integration",
  "author": "HelixAgent Team",
  "license": "MIT",
  "agent": "claude_code",
  "entry_point": "helix_integration.go",
  "dependencies": [],
  "config_schema": {
    "helix_agent_url": {
      "type": "string",
      "default": "http://localhost:8080",
      "description": "HelixAgent server URL"
    },
    "timeout": {
      "type": "integer",
      "default": 120000,
      "description": "Request timeout in milliseconds"
    }
  }
}
```

## Challenge Verification

Run the CLI agent integration challenge to verify everything works:

```bash
./challenges/scripts/cli_agent_integration_challenge.sh
```

This challenge performs 100+ tests across:
- Script existence (5 tests)
- Configuration generation (50+ tests)
- Plugin generation (30+ tests)
- Content verification (20+ tests)
- Integration compliance (10+ tests)

## Features Enabled by Integration

When a CLI agent is configured with HelixAgent, it gains access to:

| Feature | Description |
|---------|-------------|
| **AI Debate Ensemble** | Responses from 25 LLMs with consensus voting |
| **Multi-Pass Validation** | 4-phase response validation and improvement |
| **Dynamic Provider Selection** | LLMsVerifier-based provider ranking |
| **Streaming Responses** | Real-time streaming with SSE |
| **MCP Protocol Support** | Full Model Context Protocol integration |
| **Event System** | Real-time events for debate progress |
| **Brotli Compression** | Efficient response compression |

## Directory Structure

```
scripts/cli-agents/
├── generate-all-configs.sh     # Configuration generator
├── install-plugins.sh          # Plugin installer
├── README.md                   # This documentation
├── configs/
│   ├── generated/              # Generated configurations
│   │   ├── claude_code/
│   │   ├── aider/
│   │   └── .../
│   └── backups/                # Backup of existing configs
├── plugins/
│   ├── generated/              # Generated plugins
│   │   ├── claude_code/
│   │   │   ├── helix-integration/
│   │   │   ├── event-handler/
│   │   │   └── index.json
│   │   └── .../
│   └── templates/              # Plugin templates
└── tests/
    └── cli-agent-integration-test.sh
```

## Troubleshooting

### Configuration Not Generated
```bash
# Check script permissions
chmod +x generate-all-configs.sh

# Run with verbose output
bash -x generate-all-configs.sh --agent=claude_code
```

### Plugin Not Found
```bash
# Regenerate plugins
./install-plugins.sh --agent=<agent_name>

# Check plugin directory
ls -la ~/.helix-plugins/<agent_name>/
```

### JSON/YAML Validation Errors
```bash
# Validate JSON
python3 -c "import json; json.load(open('config.json'))"

# Validate YAML
python3 -c "import yaml; yaml.safe_load(open('config.yml'))"
```

### Integration Test Failures
```bash
# Run with verbose output
./tests/cli-agent-integration-test.sh --verbose

# Check test logs
cat /tmp/helix-cli-agent-tests/test.log
```

## Contributing

To add support for a new CLI agent:

1. Add agent to `CLI_AGENTS` array in `generate-all-configs.sh`
2. Add agent plugins to `AGENT_PLUGINS` array in `install-plugins.sh`
3. Create agent-specific generator if needed (e.g., `generate_<agent>_config()`)
4. Add tests in `cli-agent-integration-test.sh`
5. Add challenge tests in `cli_agent_integration_challenge.sh`
6. Update this README

## Related Documentation

- [CLI Agent Plugins Development Plan](../../docs/CLI_AGENT_PLUGINS_PLAN.md)
- [HelixAgent Main Documentation](../../CLAUDE.md)
- [Progress Report](../../docs/PROGRESS_REPORT.md)
