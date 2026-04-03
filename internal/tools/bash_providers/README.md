# HelixAgent Unified Bash Tools System

## Overview

This directory contains a curated collection of bash tools extracted from various CLI agents, unified and integrated into HelixAgent's MCP (Model Context Protocol) system.

**Sources**:
- AIChat LLM Functions (sigoden/llm-functions)
- Claude Code hooks and scripts
- Cline utility scripts
- OpenHands sandbox tools
- Aider helper scripts
- Custom HelixAgent tools

## Tool Categories

### 1. Filesystem Operations (`fs_*.sh`)

| Tool | Description | Source |
|------|-------------|--------|
| `fs_cat.sh` | Read file contents | AIChat LLM Functions |
| `fs_ls.sh` | List directory contents | AIChat LLM Functions |
| `fs_mkdir.sh` | Create directories | AIChat LLM Functions |
| `fs_write.sh` | Write to files | AIChat LLM Functions |
| `fs_patch.sh` | Apply patches | AIChat LLM Functions |
| `fs_rm.sh` | Remove files/directories | AIChat LLM Functions |
| `fs_find.sh` | Find files (HelixAgent) | Custom |
| `fs_stat.sh` | File statistics (HelixAgent) | Custom |

### 2. Execution (`exec_*.sh`)

| Tool | Description | Source |
|------|-------------|--------|
| `exec_command.sh` | Execute shell commands | AIChat LLM Functions |
| `exec_js.sh` | Execute JavaScript | AIChat LLM Functions |
| `exec_py.sh` | Execute Python | AIChat LLM Functions |
| `exec_sql.sh` | Execute SQL | AIChat LLM Functions |
| `exec_docker.sh` | Docker operations (HelixAgent) | Custom |
| `exec_helm.sh` | Helm operations (HelixAgent) | Custom |

### 3. Web Operations (`web_*.sh`)

| Tool | Description | Source |
|------|-------------|--------|
| `web_search_perplexity.sh` | Search via Perplexity | AIChat LLM Functions |
| `web_search_tavily.sh` | Search via Tavily | AIChat LLM Functions |
| `web_search_aichat.sh` | Search via AIChat | AIChat LLM Functions |
| `web_fetch_curl.sh` | Fetch URL via curl | AIChat LLM Functions |
| `web_fetch_jina.sh` | Fetch URL via jina.ai | AIChat LLM Functions |

### 4. Information Retrieval (`info_*.sh`)

| Tool | Description | Source |
|------|-------------|--------|
| `info_arxiv.sh` | Search arXiv papers | AIChat LLM Functions |
| `info_wikipedia.sh` | Search Wikipedia | AIChat LLM Functions |
| `info_wolfram.sh` | Query WolframAlpha | AIChat LLM Functions |
| `info_weather.sh` | Get weather | AIChat LLM Functions |
| `info_time.sh` | Get current time | AIChat LLM Functions |

### 5. Communication (`comm_*.sh`)

| Tool | Description | Source |
|------|-------------|--------|
| `comm_email.sh` | Send emails | AIChat LLM Functions |
| `comm_twilio.sh` | Send SMS via Twilio | AIChat LLM Functions |
| `comm_slack.sh` | Post to Slack (HelixAgent) | Custom |
| `comm_discord.sh` | Post to Discord (HelixAgent) | Custom |

### 6. Git Operations (`git_*.sh`)

| Tool | Description | Source |
|------|-------------|--------|
| `git_status.sh` | Git status | HelixAgent |
| `git_diff.sh` | Git diff | HelixAgent |
| `git_commit.sh` | Git commit | HelixAgent |
| `git_branch.sh` | Git branch operations | HelixAgent |
| `git_log.sh` | Git log | HelixAgent |

### 7. HelixAgent Specific (`hx_*.sh`)

| Tool | Description |
|------|-------------|
| `hx_health.sh` | Check HelixAgent health |
| `hx_providers.sh` | List LLM providers |
| `hx_models.sh` | List available models |
| `hx_ensemble.sh` | Call ensemble endpoint |
| `hx_debate.sh` | Start debate session |
| `hx_rag.sh` | RAG operations |

## Tool Format

All tools follow the argc annotation format:

```bash
#!/usr/bin/env bash
set -e

# @describe Tool description for LLM
# @option --param! Required parameter
# @option --optional Optional parameter
# @env ENV_VAR! Required environment variable
# @env OPTIONAL_VAR[=default] Optional with default

tool_name() {
    # Implementation
}

eval "$(argc --argc-eval "$0" "$@")"
```

## Usage

### Direct Execution

```bash
./fs_cat.sh --path /etc/os-release
./web_search_perplexity.sh --query "latest Go version"
```

### Via HelixAgent MCP

```bash
curl http://localhost:7061/v1/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "fs_cat",
    "arguments": {"path": "/etc/os-release"}
  }'
```

### Via AIChat

```bash
aichat --role %functions% "Read the README file"
```

## Security

- All tools use `set -e` for error handling
- Dangerous operations require confirmation via `guard_operation.sh`
- Environment variables for API keys (never hardcoded)
- Sandboxed execution where possible

## Integration with HelixAgent

The tools are automatically:
1. Discovered at startup
2. Converted to MCP Tool format
3. Registered with the MCP server
4. Made available via `/v1/mcp/tools/list`

## Adding New Tools

1. Create script following argc format
2. Make executable: `chmod +x tool_name.sh`
3. Add to appropriate category
4. Register in `tools.yaml`
5. Restart HelixAgent

## Maintenance

Tools are synchronized with upstream sources:
- AIChat LLM Functions: Monthly sync
- Claude Code: Quarterly review
- Cline: Quarterly review

## License

Tools retain their original licenses (MIT, Apache-2.0).
HelixAgent-specific tools: Apache-2.0
