# OpenAI Codex - API Reference

## CLI Commands

### Basic Usage

```bash
codex [options] [prompt]
```

### Options

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--model` | `-m` | Model to use | `--model gpt-4.1` |
| `--approval-mode` | `-a` | Permission mode | `-a full-auto` |
| `--provider` | | AI provider | `--provider gemini` |
| `--quiet` | `-q` | Non-interactive mode | `-q` |
| `--json` | | JSON output | `--json` |
| `--notify` | | Desktop notifications | `--notify` |
| `--no-project-doc` | | Skip AGENTS.md | `--no-project-doc` |
| `--help` | `-h` | Show help | `--help` |
| `--version` | `-v` | Show version | `--version` |

### Approval Modes

| Mode | Flag | Auto File Ops | Auto Shell | Description |
|------|------|---------------|------------|-------------|
| Suggest | `suggest` | No | No | Default, ask for everything |
| Auto Edit | `auto-edit` | Yes | No | Auto-apply patches |
| Full Auto | `full-auto` | Yes | Yes* | Fully autonomous |

*Full Auto runs network-disabled and directory-confined

### Examples

```bash
# Interactive mode
codex

# Single prompt
codex "explain utils.py"

# Full auto mode
codex -a full-auto "create a React app"

# Specific model
codex -m gpt-4.1 "review this code"

# CI/CD mode
codex -q -a auto-edit "update deps"

# With notifications
codex --notify "long-running task"
```

## Configuration Schema

### config.yaml / config.json

```yaml
model: o4-mini                    # Default model
approvalMode: suggest             # Default approval mode
fullAutoErrorMode: ask-user       # Error handling in full-auto
notify: true                      # Desktop notifications

providers:                        # Custom providers
  openai:
    name: OpenAI
    baseURL: https://api.openai.com/v1
    envKey: OPENAI_API_KEY
  
  azure:
    name: Azure OpenAI
    baseURL: https://project.openai.azure.com/openai
    envKey: AZURE_OPENAI_API_KEY

history:                          # History settings
  maxSize: 1000
  saveHistory: true
  sensitivePatterns: []           # Regex patterns to filter
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `CODEX_API_KEY` | Alternative API key | - |
| `DEBUG` | Enable debug logging | `false` |
| `CODEX_QUIET_MODE` | Suppress interactive UI | `false` |
| `CODEX_DISABLE_PROJECT_DOC` | Skip AGENTS.md | `false` |
| `CODEX_RUST` | Use Rust implementation | `false` |

### Provider-Specific Keys

| Variable | Provider |
|----------|----------|
| `AZURE_OPENAI_API_KEY` | Azure |
| `OPENROUTER_API_KEY` | OpenRouter |
| `GEMINI_API_KEY` | Gemini |
| `OLLAMA_API_KEY` | Ollama |
| `MISTRAL_API_KEY` | Mistral |
| `DEEPSEEK_API_KEY` | DeepSeek |
| `XAI_API_KEY` | xAI |
| `GROQ_API_KEY` | Groq |
| `ARCEEAI_API_KEY` | ArceeAI |

## Shell Completion

```bash
# Bash
codex completion bash > /etc/bash_completion.d/codex

# Zsh
codex completion zsh > "${fpath[1]}/_codex"

# Fish
codex completion fish > ~/.config/fish/completions/codex.fish
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `130` | Interrupted (Ctrl+C) |

---

*For architecture details, see [ARCHITECTURE.md](./ARCHITECTURE.md)*
