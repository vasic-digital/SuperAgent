# Aider - API Reference

## Command Line Interface

### Global Command

```bash
aider [options] [files...]
```

### Core Options

| Option | Description | Example |
|--------|-------------|---------|
| `--model` | Specify main LLM | `--model sonnet` |
| `--api-key` | Set API key | `--api-key anthropic=sk-...` |
| `--weak-model` | Model for summaries | `--weak-model gpt-4o-mini` |
| `--editor-model` | Model for editor tasks | `--editor-model sonnet` |
| `--edit-format` | Edit format to use | `--edit-format diff` |
| `--architect` | Enable architect mode | `--architect` |

### Model Selection

```bash
# Claude models
aider --model sonnet                    # Claude 3.7 Sonnet
aider --model claude-3-7-sonnet-20250219
aider --model claude-3-5-haiku-20241022

# OpenAI models
aider --model gpt-4o                    # GPT-4o
aider --model o3-mini                   # o3-mini
aider --model o1                        # o1

# DeepSeek models
aider --model deepseek                  # DeepSeek Chat
aider --model deepseek-coder            # DeepSeek Coder

# Gemini models
aider --model gemini-2.5-pro
aider --model gemini-2.0-flash

# Local models
aider --model ollama/llama3.2
aider --model ollama/codellama
```

### File Operations

```bash
# Start with specific files
aider src/main.py src/utils.py

# Add read-only files (reference only)
aider --read config.yaml src/main.py

# Use file watcher (IDE integration)
aider --watch-files src/

# Exclude files
aider --ignore *.min.js src/
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--dark-mode` | Use dark color scheme | auto |
| `--no-pretty` | Disable pretty output | False |
| `--no-stream` | Disable streaming responses | False |
| `--cache-prompts` | Enable prompt caching | False |
| `--map-tokens` | Repo map token limit | 1024 |
| `--no-git` | Disable git integration | False |
| `--auto-commits` | Auto-commit changes | True |
| `--dry-run` | Show changes without applying | False |

### Voice & Input

```bash
# Enable voice input
aider --voice-language en

# Suggest shell commands
aider --suggest-shell-commands

# Enable multiline mode
aider --multiline
```

---

## In-Chat Commands

### File Management

| Command | Description | Usage |
|---------|-------------|-------|
| `/add` | Add files to chat | `/add src/main.py` |
| `/drop` | Remove files from chat | `/drop src/old.py` |
| `/read-only` | Add read-only files | `/read-only README.md` |
| `/ls` | List files in chat | `/ls` |

### Model Control

| Command | Description | Usage |
|---------|-------------|-------|
| `/model` | Switch main model | `/model gpt-4o` |
| `/editor-model` | Switch editor model | `/editor-model sonnet` |
| `/weak-model` | Switch weak model | `/weak-model haiku` |
| `/models` | List available models | `/models` |
| `/chat-mode` | Change chat mode | `/chat-mode architect` |

### Chat Modes

| Command | Description |
|---------|-------------|
| `/ask` | Ask mode (no edits) |
| `/code` | Code mode (default) |
| `/architect` | Architect/editor mode |
| `/context` | Context mode (auto file detection) |

### Git Operations

| Command | Description | Usage |
|---------|-------------|-------|
| `/commit` | Commit external changes | `/commit "message"` |
| `/undo` | Undo last Aider commit | `/undo` |
| `/git` | Run git command | `/git status` |
| `/diff` | Show diff of changes | `/diff` |

### Utility Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/help` | Show help | `/help` |
| `/clear` | Clear chat history | `/clear` |
| `/reset` | Drop all files and clear | `/reset` |
| `/exit` | Exit Aider | `/exit` or `/quit` |
| `/settings` | Show current settings | `/settings` |
| `/tokens` | Show token usage | `/tokens` |

### Advanced Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/map` | Show repository map | `/map` |
| `/map-refresh` | Refresh repo map | `/map-refresh` |
| `/lint` | Run linter | `/lint` |
| `/test` | Run tests | `/test` |
| `/voice` | Voice input | `/voice` |
| `/web` | Add web page | `/web https://example.com` |
| `/paste` | Paste from clipboard | `/paste` |
| `/copy` | Copy last response | `/copy` |
| `/run` | Run shell command | `/run make test` |
| `/load` | Load commands from file | `/load session.aider` |
| `/save` | Save commands to file | `/save session.aider` |
| `/reasoning-effort` | Set reasoning level | `/reasoning-effort high` |
| `/think-tokens` | Set thinking budget | `/think-tokens 8192` |

---

## Configuration File Reference

### .aider.conf.yml

```yaml
# Model settings
model: claude-3-7-sonnet-20250219
weak-model: claude-3-5-haiku-20241022
editor-model: claude-3-7-sonnet-20250219

# API keys (can also use .env file)
anthropic-api-key: sk-...  # Or set ANTHROPIC_API_KEY
openai-api-key: sk-...     # Or set OPENAI_API_KEY

# Edit format
edit-format: diff

# Repository map
map-tokens: 1024
map-refresh: auto  # auto, always, files, manual

# Git settings
auto-commits: true
dirty-commits: true
commit-prompt: "Please write a commit message..."

# UI settings
dark-mode: true
pretty: true
stream: true

# Voice
voice-language: en
voice-format: wav

# Linting and testing
lint-cmd: python -m flake8
test-cmd: pytest
auto-lint: true
auto-test: false

# History
input-history-file: .aider.input.history
chat-history-file: .aider.chat.history.md
restore-chat-history: false

# Caching
cache-prompts: false
cache-keepalive-pings: 0

# Other
suggest-shell-commands: true
show-model-warnings: true
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `AIDER_MODEL` | Default model |
| `AIDER_WEAK_MODEL` | Weak model |
| `AIDER_EDITOR_MODEL` | Editor model |
| `AIDER_DARK_MODE` | Enable dark mode |
| `AIDER_CACHE_PROMPTS` | Enable prompt caching |
| `AIDER_AUTO_COMMITS` | Auto-commit changes |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `GEMINI_API_KEY` | Gemini API key |

### .env File Example

```bash
# API Keys
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
DEEPSEEK_API_KEY=sk-...
GEMINI_API_KEY=AI...

# Aider Settings
AIDER_MODEL=sonnet
AIDER_WEAK_MODEL=claude-3-5-haiku-20241022
AIDER_DARK_MODE=true
AIDER_CACHE_PROMPTS=true
```

---

## Model Settings File

### .aider.model.settings.yml

Define custom models or override defaults:

```yaml
- name: custom-model
  edit_format: diff
  weak_model_name: gpt-4o-mini
  use_repo_map: true
  send_undo_reply: true
  accepts_temperature: true
  accepts_top_p: true

- name: local-model
  edit_format: whole
  weak_model_name: local-model
  use_repo_map: false
  accepts_temperature: false

- name: reasoning-model
  edit_format: diff
  weak_model_name: gpt-4o-mini
  use_repo_map: true
  accepts_reasoning_effort: true
  accepts_thinking_tokens: true
```

### .aider.model.metadata.json

```json
{
  "custom-model": {
    "max_tokens": 128000,
    "max_input_tokens": 128000,
    "max_output_tokens": 4096,
    "input_cost_per_token": 0.000003,
    "output_cost_per_token": 0.000015
  }
}
```

---

## Chat Mode Reference

### Available Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `code` | Standard coding mode | Most tasks |
| `ask` | Questions only, no edits | Understanding code |
| `architect` | Design + edit models | Complex changes |
| `context` | Auto file detection | Large projects |
| `help` | Aider documentation | Learning aider |

### Mode Switching

```bash
# From command line
aider --edit-format ask
aider --architect

# In chat
> /ask How does this function work?
> /code Please refactor this
> /architect Design a new API endpoint
> /context Implement feature X
```

---

## Edit Format Reference

### Format Comparison

| Format | Description | Best For |
|--------|-------------|----------|
| `diff` | Search/replace blocks | Most models |
| `whole` | Complete file replacement | Small files |
| `udiff` | Unified diff format | Compatibility |
| `editor` | Editor-style changes | Specific tools |
| `editor-diff` | Editor with diff | Large changes |
| `single Wholefile` | One file at a time | Limited context |

### Format Selection

Aider automatically selects the best format for each model:

```python
# Default mappings
"claude-3-7-sonnet": "diff"
"gpt-4o": "diff"
"deepseek": "diff"
"o1": "diff"
"haiku": "whole"
```

Override with `--edit-format` flag.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `130` | Interrupted (Ctrl+C) |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
