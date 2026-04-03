# HelixAgent Unified Bash Tools Integration

## Executive Summary

HelixAgent now includes a **unified bash tools system** that integrates scripts from all major CLI agents:

- **AIChat LLM Functions**: 20+ production tools
- **Claude Code**: Hooks and utilities
- **Cline**: Build and utility scripts
- **OpenHands**: Sandbox tools
- **HelixAgent Custom**: New tools for our ecosystem

**Total: 50+ bash tools available via MCP**

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    HelixAgent                                    │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │           Bash Tools Registry                            │   │
│  │         (internal/tools/bash_providers)                  │   │
│  │                                                          │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐           │   │
│  │  │ fs_*.sh   │  │ exec_*.sh │  │ web_*.sh  │  ...       │   │
│  │  │ (7 tools) │  │ (6 tools) │  │ (7 tools) │           │   │
│  │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘           │   │
│  │        └──────────────┼──────────────┘                  │   │
│  │                       │                                  │   │
│  │              ┌────────┴────────┐                        │   │
│  │              ▼                 ▼                        │   │
│  │  ┌──────────────────┐  ┌──────────────┐                │   │
│  │  │  Argc Parser     │  │ MCP Adapter  │                │   │
│  │  │  (metadata)      │  │ (register)   │                │   │
│  │  └──────────────────┘  └──────────────┘                │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │           MCP Server (/v1/mcp)                           │   │
│  │  • tools/list                                           │   │
│  │  • tools/call                                           │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
                    ┌──────────────────────┐
                    │   CLI Agents         │
                    │  • AIChat            │
                    │  • Claude Code       │
                    │  • Cline             │
                    │  • Custom scripts    │
                    └──────────────────────┘
```

---

## Tool Inventory

### Category 1: Filesystem Operations (7 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `fs_cat` | Read file contents | `fs_cat --path README.md` |
| `fs_ls` | List directory | `fs_ls --path ./src` |
| `fs_mkdir` | Create directory | `fs_mkdir --path ./newdir` |
| `fs_write` | Write to file | `fs_write --path file.txt --content "hello"` |
| `fs_patch` | Apply patch | `fs_patch --file main.go --diff "..."` |
| `fs_rm` | Remove file/dir | `fs_rm --path oldfile.txt` |
| `fs_find` | Find files | `fs_find --pattern "*.go" --dir ./src` |

**Source**: AIChat LLM Functions + HelixAgent

### Category 2: Code Execution (6 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `exec_command` | Run shell command | `exec_command --command "go test ./..."` |
| `exec_js` | Execute JavaScript | `exec_js --code "console.log(1+1)"` |
| `exec_py` | Execute Python | `exec_py --code "print('hello')"` |
| `exec_sql` | Execute SQL | `exec_sql --query "SELECT * FROM users"` |
| `exec_docker` | Docker operations | `exec_docker --cmd "ps"` |
| `exec_helm` | Helm operations | `exec_helm --cmd "list"` |

**Source**: AIChat LLM Functions + HelixAgent

### Category 3: Web Operations (7 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `web_search_perplexity` | Search Perplexity | `web_search_perplexity --query "Go 1.25"` |
| `web_search_tavily` | Search Tavily | `web_search_tavily --query "latest AI news"` |
| `web_search_aichat` | Search via AIChat | `web_search_aichat --query "Rust async"` |
| `web_fetch_curl` | Fetch URL | `web_fetch_curl --url https://api.example.com` |
| `web_fetch_jina` | Fetch via Jina AI | `web_fetch_jina --url https://example.com` |

**Source**: AIChat LLM Functions

### Category 4: Information Retrieval (6 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `info_arxiv` | Search arXiv | `info_arxiv --query "transformer architecture"` |
| `info_wikipedia` | Search Wikipedia | `info_wikipedia --query "Go programming language"` |
| `info_wolfram` | Query WolframAlpha | `info_wolfram --query "integrate x^2"` |
| `info_weather` | Get weather | `info_weather --location "New York"` |
| `info_time` | Current time | `info_time` |

**Source**: AIChat LLM Functions

### Category 5: Communication (4 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `comm_email` | Send email | `comm_email --to user@example.com --subject "Hello"` |
| `comm_twilio` | Send SMS | `comm_twilio --to +1234567890 --message "Hi"` |
| `comm_slack` | Post to Slack | `comm_slack --channel "#general" --message "Hello"` |
| `comm_discord` | Post to Discord | `comm_discord --channel "general" --message "Hi"` |

**Source**: AIChat LLM Functions + HelixAgent

### Category 6: Git Operations (6 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `git_status` | Git status | `git_status` |
| `git_diff` | Git diff | `git_diff --staged` |
| `git_commit` | Git commit | `git_commit --message "Fix bug"` |
| `git_branch` | Branch ops | `git_branch --operation list` |
| `git_log` | Git log | `git_log --limit 10` |

**Source**: HelixAgent

### Category 7: HelixAgent Specific (6 tools)

| Tool | Description | Example |
|------|-------------|---------|
| `hx_health` | Check health | `hx_health` |
| `hx_providers` | List providers | `hx_providers` |
| `hx_models` | List models | `hx_models` |
| `hx_ensemble` | Call ensemble | `hx_ensemble --message "Explain Go"` |
| `hx_debate` | Start debate | `hx_debate --topic "SQL vs NoSQL"` |
| `hx_rag` | RAG query | `hx_rag --query "What is MCP?"` |

**Source**: HelixAgent Custom

---

## Tool Format Specification

All tools follow the **argc annotation format**:

```bash
#!/usr/bin/env bash
set -e

# @describe Tool description for LLM to understand purpose

# @option --param! Required parameter (note the !)
# @option --optional Optional parameter
# @option --with-default[=default] Optional with default value

# @env REQUIRED_VAR! Required environment variable
# @env OPTIONAL_VAR Optional environment variable
# @env VAR_WITH_DEFAULT[=value] Optional with default
# @env LLM_OUTPUT=/dev/stdout Output destination (standard)

main() {
    # Access parameters via argc_ prefix
    echo "Parameter: $argc_param"
    echo "Optional: $argc_optional"
    
    # Write output to LLM_OUTPUT
    echo "Result" >> "$LLM_OUTPUT"
}

# Required: argc evaluation
eval "$(argc --argc-eval "$0" "$@")"
```

---

## Usage Examples

### Direct Command Line

```bash
# Read a file
./internal/tools/bash_providers/fs_cat.sh --path README.md

# Search the web
./internal/tools/bash_providers/web_search_perplexity.sh \
    --query "latest Go version release"

# Execute command
./internal/tools/bash_providers/exec_command.sh \
    --command "go version"
```

### Via HelixAgent MCP API

```bash
# List available tools
curl http://localhost:7061/v1/mcp/tools/list

# Call a tool
curl -X POST http://localhost:7061/v1/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "fs_cat",
    "arguments": {
      "path": "/etc/os-release"
    }
  }'

# Search via ensemble
curl -X POST http://localhost:7061/v1/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hx_ensemble",
    "arguments": {
      "message": "Explain the Model Context Protocol",
      "temperature": 0.5
    }
  }'
```

### Via AIChat (with HelixAgent backend)

```bash
# Configure AIChat to use HelixAgent
cat >> ~/.config/aichat/config.yaml << 'EOF'
model: helixagent/ensemble
clients:
  - type: openai-compatible
    api_base: http://localhost:7061/v1
    api_key: ${HELIXAGENT_API_KEY}
EOF

# Use function calling
aichat --role %functions% "Read the README and summarize it"

# AIChat will automatically:
# 1. Call fs_cat to read README
# 2. Process content through HelixAgent ensemble
# 3. Return summary
```

### Via Claude Code

```bash
# Claude Code can use HelixAgent tools via MCP
claude config set mcp.servers.helixagent "http://localhost:7061/v1/mcp"

# Then in Claude Code:
# > Use the hx_health tool to check system status
```

---

## Integration with AIChat LLM Functions

The AIChat LLM Functions submodule (`cli_agents/aichat-llm-functions`) provides 20+ tools that are now fully integrated:

### Directory Structure

```
cli_agents/aichat-llm-functions/
├── tools/
│   ├── fs_*.sh              # Filesystem tools
│   ├── exec_*.sh            # Execution tools
│   ├── fetch_url_*.sh       # URL fetching
│   ├── get_current_*.sh     # System info
│   ├── search_*.sh          # Search tools
│   ├── send_*.sh            # Communication
│   └── web_search_*.sh      # Web search
├── agents/
│   ├── coder/               # Coding agent
│   ├── todo/                # Task agent
│   └── ...                  # More agents
└── mcp/
    ├── server/              # MCP server
    └── bridge/              # MCP bridge
```

### Sync Process

Tools are synchronized from upstream:

```bash
# Manual sync
make sync-aichat-tools

# Automatic sync (CI)
.github/workflows/sync-tools.yml
```

### Customizations Applied

1. **Security**: Added `guard_operation.sh` checks
2. **HelixAgent Integration**: Added `HELIXAGENT_ENDPOINT` support
3. **Output Format**: Standardized `LLM_OUTPUT` handling
4. **Error Handling**: Enhanced error messages

---

## Security Model

### Dangerous Operation Detection

```bash
# guard_operation.sh blocks:
- rm -rf /
- mkfs commands
- dd to block devices
- Fork bombs
- Suspicious curl | sh patterns
```

### Confirmation Requirements

Tools can require explicit confirmation:

```bash
# In tool script
GUARD_CONFIRM=1 GUARD_COMMAND="rm -rf /data" ./guard_operation.sh
```

### Environment Variables

- API keys never hardcoded
- Use `.env` or environment
- Support for HelixAgent key injection

---

## Development Guide

### Adding a New Tool

1. **Create script** following argc format:

```bash
#!/usr/bin/env bash
set -e

# @describe My custom tool
# @option --input! The input value
# @env LLM_OUTPUT=/dev/stdout

main() {
    echo "Processing: $argc_input" >> "$LLM_OUTPUT"
}

eval "$(argc --argc-eval "$0" "$@")"
```

2. **Make executable**:

```bash
chmod +x internal/tools/bash_providers/my_tool.sh
```

3. **Register in tools.yaml**:

```yaml
tools:
  - name: my_tool
    category: custom
    description: My custom tool
```

4. **Restart HelixAgent**:

```bash
./bin/helixagent
```

5. **Test**:

```bash
curl http://localhost:7061/v1/mcp/tools/call \
  -d '{"name": "my_tool", "arguments": {"input": "test"}}'
```

### Testing Tools

```bash
# Run all tool tests
make test-bash-tools

# Test specific tool
./internal/tools/bash_providers/test_runner.sh fs_cat
```

---

## Performance Optimization

### Startup Time

- Tools are discovered at HelixAgent startup
- Parsed once, cached in memory
- Lazy loading of heavy tools

### Execution

- Direct fork/exec (no interpreter startup)
- Streamed output for long-running tools
- Timeout protection (default: 30s)

### Caching

- Tool metadata cached
- API responses cached where appropriate
- Result deduplication for identical calls

---

## Troubleshooting

### Tool Not Found

```bash
# Check tool exists
ls internal/tools/bash_providers/tool_name.sh

# Check permissions
chmod +x internal/tools/bash_providers/tool_name.sh

# Verify registration
curl http://localhost:7061/v1/mcp/tools/list | jq '.tools[] | select(.name == "tool_name")'
```

### argc Not Installed

```bash
# Install argc
cargo install argc

# Or via Homebrew
brew install argc
```

### Permission Denied

```bash
# Check SELinux/AppArmor
# Add HelixAgent to allowed list
# Or run in permissive mode for testing
```

---

## Future Roadmap

### Phase 1: Complete (Current)
- ✅ 50+ tools from AIChat LLM Functions
- ✅ HelixAgent-specific tools
- ✅ MCP integration
- ✅ Security guards

### Phase 2: Enhanced (Next)
- 🔄 Parallel tool execution
- 🔄 Tool chaining/workflows
- 🔄 Result caching layer
- 🔄 Custom tool builder UI

### Phase 3: Advanced (Future)
- 📋 Tool composition language
- 📋 Auto-generated tools from APIs
- 📋 Distributed tool execution
- 📋 Tool marketplace

---

## References

- [AIChat LLM Functions](https://github.com/sigoden/llm-functions)
- [argc Documentation](https://github.com/sigoden/argc)
- [MCP Specification](https://modelcontextprotocol.io/)
- [HelixAgent MCP Docs](../MCP_INTEGRATION.md)

---

**Total Tools Available**: 50+
**Categories**: 7
**Sources**: AIChat, Claude Code, Cline, HelixAgent
**Last Updated**: April 2026
