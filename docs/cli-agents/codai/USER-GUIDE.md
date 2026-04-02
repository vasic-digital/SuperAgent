# Codai - User Guide

**Codai** is an AI code assistant that helps developers through a session-based CLI, providing intelligent code suggestions, refactoring, code review, and context-aware completions powered by LLMs.

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

- Go 1.21+ installed
- API keys for LLM providers (OpenAI, etc.)
- Optional: Ollama for local models

### Method 1: Go Install (Recommended)

```bash
go install github.com/meysamhadeli/codai@latest
```

### Method 2: From Source

```bash
# Clone repository
git clone https://github.com/meysamhadeli/codai.git
cd codai

# Build
go build -o codai .

# Move to PATH
mv codai ~/.local/bin/
```

### Method 3: Download Binary

```bash
# Download from releases
curl -L https://github.com/meysamhadeli/codai/releases/latest/download/codai-$(uname -s)-$(uname -m) -o codai

# Make executable
chmod +x codai

# Move to PATH
mv codai ~/.local/bin/
```

### Verify Installation

```bash
codai --version
```

---

## Quick Start

### Set Environment Variables

**macOS/Linux:**
```bash
# Required: Chat API key
export CHAT_API_KEY="your_openai_api_key"

# Optional: Embeddings API key (for RAG)
export EMBEDDINGS_API_KEY="your_embeddings_api_key"

# Add to shell profile for persistence
echo 'export CHAT_API_KEY="your_key"' >> ~/.bashrc
```

**Windows (PowerShell):**
```powershell
$env:CHAT_API_KEY="your_openai_api_key"
$env:EMBEDDINGS_API_KEY="your_embeddings_api_key"
```

### Start Codai

```bash
# Navigate to your project
cd ~/projects/my-app

# Start Codai
codai
```

### Your First Interaction

```bash
# Ask about your codebase
> Explain the structure of this project

# Request code changes
> Add error handling to the main function

# Get suggestions
> How can I improve this code?
```

### Exit

```bash
# Exit session
/exit
# or
quit
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `codai` | Start interactive session |
| `codai --version` | Show version |
| `codai --help` | Show help |

### In-Session Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear conversation history |
| `/context` | Show current context |
| `/exit` or `quit` | End session |

---

## TUI/Interactive Commands

### Interactive Session Controls

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate command history |
| `Tab` | Autocomplete |
| `Ctrl+C` | Cancel current operation |
| `Ctrl+D` | Exit session |

### Context Navigation

Codai maintains context of your codebase:

```bash
# View current context
/context

# Clear context
/clear
```

---

## Configuration

### Configuration File

Codai uses a configuration file for advanced settings. Create `~/.codai/config.yaml`:

```yaml
# LLM Configuration
llm:
  # Chat model provider
  chat_provider: "openai"  # or "ollama"
  
  # Chat model name
  chat_model: "gpt-4o"
  
  # Embeddings provider (for RAG)
  embeddings_provider: "openai"  # or "ollama"
  
  # Embeddings model
  embeddings_model: "text-embedding-3-small"
  
  # API configuration
  api_key: ""  # Uses CHAT_API_KEY env var if empty
  embeddings_api_key: ""  # Uses EMBEDDINGS_API_KEY env var
  base_url: ""  # Custom base URL for API
  
  # Model parameters
  temperature: 0.7
  max_tokens: 2000
  top_p: 1.0

# RAG Configuration
rag:
  enabled: true
  
  # Code chunking settings
  chunk_size: 1000
  chunk_overlap: 200
  
  # Similarity search
  top_k: 5
  
  # File patterns to include
  include_patterns:
    - "*.go"
    - "*.js"
    - "*.ts"
    - "*.py"
    - "*.rs"
  
  # File patterns to exclude
  exclude_patterns:
    - "*.test.go"
    - "*_test.js"
    - "node_modules/**"
    - ".git/**"
    - "vendor/**"

# Session Configuration
session:
  # Save conversation history
  save_history: true
  
  # History file location
  history_file: "~/.codai/history.json"
  
  # Max history entries
  max_history: 100

# UI Configuration
ui:
  # Color scheme
  color_scheme: "auto"  # auto, dark, light, none
  
  # Show token usage
  show_tokens: true
  
  # Code theme
  code_theme: "monokai"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CHAT_API_KEY` | API key for chat model |
| `EMBEDDINGS_API_KEY` | API key for embeddings model |
| `CODAI_CONFIG` | Path to config file |
| `CODAI_DATA` | Path to data directory |

### Model Recommendations

**Best Cloud Models:**
- `gpt-4o` - Best overall performance
- `gpt-4` - High quality, slower
- `claude-3-5-sonnet` - Excellent for code (requires config)

**Recommended Local Models (via Ollama):**
- `phi3:medium` - Good balance of speed and quality
- `mistral-large` - Strong coding performance
- `llama3.1:70b` - High quality, requires more resources

**Embedding Models:**

OpenAI:
- `text-embedding-3-large` - Best quality
- `text-embedding-3-small` - Good balance
- `text-embedding-ada-002` - Legacy, still effective

Ollama:
- `mxbai-embed-large` - High quality
- `nomic-embed-text` - Fast, good quality
- `all-minilm` - Lightweight

### Ollama Configuration

For local models with Ollama:

```yaml
llm:
  chat_provider: "ollama"
  chat_model: "phi3:medium"
  embeddings_provider: "ollama"
  embeddings_model: "nomic-embed-text"
  base_url: "http://localhost:11434"
```

---

## Usage Examples

### Code Exploration

```bash
# Start in project directory
cd ~/projects/my-app
codai

# Ask about project structure
> Explain the architecture of this codebase

# Find specific functionality
> Where is authentication handled?

# Understand complex code
> Explain the logic in src/utils/algorithm.ts
```

### Code Generation

```bash
# Generate new function
> Write a function to validate email addresses

# Generate with context
> Add a method to the User class for password reset

# Generate tests
> Write unit tests for the auth module
```

### Code Refactoring

```bash
# Request refactoring
> Refactor this code to use async/await instead of callbacks

# Improve performance
> Optimize this function for better performance

# Modernize code
> Convert this to use modern JavaScript features
```

### Code Review

```bash
# Review code
> Review this file for potential bugs

# Check for best practices
> Are there any security issues in this code?

# Get improvement suggestions
> How can I make this code more maintainable?
```

### Debugging

```bash
# Debug errors
> I'm getting this error: [paste error]. What's wrong?

# Analyze logs
> Help me understand these error logs

# Fix failing tests
> The tests are failing. Help me debug.
```

### Using RAG

```bash
# With RAG enabled, Codai automatically retrieves relevant context

# Ask about specific patterns
> How do we handle database connections in this project?

# Find usage examples
> Show me examples of API endpoint implementation

# Understand conventions
> What are the coding conventions for error handling?
```

### Context-Aware Assistance

```bash
# Reference specific files
> Look at @src/config/database.ts and explain the connection setup

# Ask about changes
> What changed in the recent commits?

# Compare implementations
> Compare the error handling in @src/api/users.ts vs @src/api/posts.ts
```

### Session Management

```bash
# View conversation history
> /history

# Clear context when switching tasks
> /clear

# Check current context window
> /context
```

---

## Troubleshooting

### Installation Issues

#### "command not found" after go install

```bash
# Check Go bin is in PATH
go env GOPATH
ls $(go env GOPATH)/bin

# Add to PATH
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Build fails from source

```bash
# Check Go version
go version  # Need 1.21+

# Update Go
# Visit https://go.dev/dl/

# Verify module
go mod tidy
go build -v
```

### API Key Issues

#### "API key not set"

```bash
# Set environment variable
export CHAT_API_KEY="sk-..."

# Or add to config
cat >> ~/.codai/config.yaml << 'EOF'
llm:
  api_key: "sk-..."
EOF
```

#### "Authentication failed"

```bash
# Verify key is valid
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $CHAT_API_KEY"

# Check for whitespace in key
echo "$CHAT_API_KEY" | od -c

# Regenerate key if needed
# https://platform.openai.com/api-keys
```

### Model Issues

#### "Model not found"

```bash
# Check available models
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $CHAT_API_KEY" | jq '.data[].id'

# Update config with valid model
cat > ~/.codai/config.yaml << 'EOF'
llm:
  chat_model: "gpt-4o"
EOF
```

#### "Rate limit exceeded"

```bash
# Check usage
# https://platform.openai.com/usage

# Implement retry logic in config
# Or wait and retry

# Consider upgrading tier
```

### RAG Issues

#### "Embeddings not working"

```bash
# Check embeddings API key
export EMBEDDINGS_API_KEY="sk-..."

# Or disable RAG temporarily
cat > ~/.codai/config.yaml << 'EOF'
rag:
  enabled: false
EOF
```

#### "Index build fails"

```bash
# Check file patterns
# Ensure include_patterns match your files

# Check for large files
find . -type f -size +1M

# Exclude large binary files
# Update exclude_patterns in config
```

### Context Issues

#### "Context too long"

```bash
# Clear context
/clear

# Reduce context in config
session:
  max_history: 50

# Compact conversation periodically
```

#### "File not found in context"

```bash
# Check file exists
ls -la path/to/file

# Verify RAG includes the file extension
# Check exclude_patterns aren't too broad
```

### Performance Issues

#### Slow responses

```bash
# Check network
ping api.openai.com

# Use faster model
cat > ~/.codai/config.yaml << 'EOF'
llm:
  chat_model: "gpt-4o-mini"  # Faster, cheaper
EOF

# Enable local models with Ollama
```

#### High token usage

```bash
# Monitor usage in UI
# Check show_tokens: true in config

# Reduce max_tokens
cat > ~/.codai/config.yaml << 'EOF'
llm:
  max_tokens: 1000
EOF

# Clear context regularly
/clear
```

### Ollama Issues

#### "Cannot connect to Ollama"

```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama
ollama serve

# Check model is pulled
ollama list
ollama pull phi3:medium
```

#### "Model not found in Ollama"

```bash
# List available models
ollama list

# Pull model
ollama pull llama3.1:70b

# Verify in config
cat > ~/.codai/config.yaml << 'EOF'
llm:
  chat_provider: "ollama"
  chat_model: "llama3.1:70b"
  base_url: "http://localhost:11434"
EOF
```

### Common Errors

#### "Config file not found"

```bash
# Create default config
mkdir -p ~/.codai
cat > ~/.codai/config.yaml << 'EOF'
llm:
  chat_provider: "openai"
  chat_model: "gpt-4o"
  api_key: ""
rag:
  enabled: true
EOF
```

#### "Permission denied"

```bash
# Check file permissions
ls -la ~/.codai/

# Fix permissions
chmod 755 ~/.codai
chmod 644 ~/.codai/config.yaml
```

### Debug Mode

```bash
# Enable debug logging
export CODAI_DEBUG=1
codai

# Check logs
# Logs stored in ~/.codai/logs/
```

### Getting Help

```bash
# In-session help
/help

# CLI help
codai --help

# GitHub Issues
# https://github.com/meysamhadeli/codai/issues
```

---

## Best Practices

1. **Set API Keys Early**: Configure before first use
2. **Use RAG**: Enable for better context awareness
3. **Clear Context Regularly**: Prevent token bloat
4. **Be Specific**: Clear prompts yield better results
5. **Review Changes**: Always review AI-generated code
6. **Version Control**: Commit before major changes
7. **Local Models**: Use Ollama for sensitive code
8. **Monitor Usage**: Keep track of API costs
9. **Customize Config**: Tune for your workflow
10. **Update Regularly**: Keep Codai updated

---

## Comparison with Alternatives

| Feature | Codai | Claude Code | Aider |
|---------|-------|-------------|-------|
| Open Source | Yes | Partial | Yes |
| Local Models | Yes | No | Yes |
| RAG Support | Built-in | Via MCP | Via Aider's repo map |
| Multi-provider | Yes | Anthropic only | Yes |
| Git Integration | Basic | Built-in | Advanced |
| Language | Go | TypeScript | Python |

---

*Last Updated: April 2026*
