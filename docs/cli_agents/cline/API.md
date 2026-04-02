# Cline - API Reference

## VS Code Extension Interface

### Installation

```bash
# Via VS Code Marketplace
1. Open VS Code
2. Press Ctrl+Shift+X (Cmd+Shift+X on Mac)
3. Search for "Cline"
4. Click Install

# Via Command Line
code --install-extension saoudrizwan.claude-dev
```

### Extension Activation

**Activation Events:**
- `onLanguage` - When any file is opened
- `onUri` - When URI handler is invoked
- `onStartupFinished` - When VS Code finishes startup
- `workspaceContains:evals.env` - When evals.env file detected

---

## Commands

### Built-in Commands

| Command | Title | Description |
|---------|-------|-------------|
| `cline.plusButtonClicked` | New Task | Start a new task |
| `cline.mcpButtonClicked` | MCP Servers | Open MCP server management |
| `cline.historyButtonClicked` | History | View task history |
| `cline.accountButtonClicked` | Account | Manage account settings |
| `cline.settingsButtonClicked` | Settings | Open Cline settings |
| `cline.addToChat` | Add to Cline | Add selection to chat |
| `cline.addTerminalOutputToChat` | Add to Cline | Add terminal output to chat |
| `cline.focusChatInput` | Jump to Chat Input | Focus chat input field |
| `cline.generateGitCommitMessage` | Generate Commit Message | Generate commit with Cline |
| `cline.explainCode` | Explain with Cline | Explain selected code |
| `cline.improveCode` | Improve with Cline | Improve selected code |
| `cline.jupyterGenerateCell` | Generate Jupyter Cell | Generate notebook cell |
| `cline.jupyterExplainCell` | Explain Jupyter Cell | Explain notebook cell |
| `cline.jupyterImproveCell` | Improve Jupyter Cell | Improve notebook cell |
| `cline.openWalkthrough` | Open Walkthrough | Show getting started guide |

### Keybindings

| Keybinding | Command | When |
|------------|---------|------|
| `Cmd+'` / `Ctrl+'` | `cline.addToChat` | `editorHasSelection` |
| `Cmd+'` / `Ctrl+'` | `cline.focusChatInput` | `!editorHasSelection` |
| `Enter` | `editor.action.submitComment` | Comment editor focused |

---

## Configuration Settings

### VS Code Settings

```json
{
  // API Provider Configuration
  "cline.apiProvider": "anthropic",
  "cline.anthropic.apiKey": "sk-ant-api...",
  "cline.anthropic.model": "claude-3-7-sonnet-20250219",
  
  // OpenAI Configuration
  "cline.openai.apiKey": "sk-...",
  "cline.openai.model": "gpt-4o",
  "cline.openai.baseUrl": "https://api.openai.com/v1",
  
  // OpenRouter Configuration
  "cline.openrouter.apiKey": "sk-or-...",
  "cline.openrouter.model": "anthropic/claude-3.7-sonnet",
  
  // Local Model Configuration (Ollama)
  "cline.ollama.baseUrl": "http://localhost:11434",
  "cline.ollama.model": "codellama",
  
  // Auto-approve Settings
  "cline.autoApprove.readFiles": true,
  "cline.autoApprove.editFiles": false,
  "cline.autoApprove.executeCommands": false,
  "cline.autoApprove.useBrowser": false,
  
  // Browser Settings
  "cline.browser.enabled": true,
  "cline.browser.headless": true,
  
  // UI Settings
  "cline.sidebar.location": "right",
  
  // Custom Instructions
  "cline.customInstructions": "Always use TypeScript strict mode"
}
```

### Setting Descriptions

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `cline.apiProvider` | string | "anthropic" | Default LLM provider |
| `cline.anthropic.apiKey` | string | "" | Anthropic API key |
| `cline.anthropic.model` | string | "claude-3-7-sonnet" | Claude model ID |
| `cline.openai.apiKey` | string | "" | OpenAI API key |
| `cline.openai.model` | string | "gpt-4o" | OpenAI model ID |
| `cline.openrouter.apiKey` | string | "" | OpenRouter API key |
| `cline.ollama.baseUrl` | string | "http://localhost:11434" | Ollama server URL |
| `cline.autoApprove.readFiles` | boolean | false | Auto-approve file reads |
| `cline.autoApprove.editFiles` | boolean | false | Auto-approve file edits |
| `cline.autoApprove.executeCommands` | boolean | false | Auto-approve commands |
| `cline.customInstructions` | string | "" | Custom system prompt |

---

## Tool Reference

### File Operations

#### read_file

Read contents of a file.

**Parameters:**
```json
{
  "path": "string"  // Absolute or relative file path
}
```

**Example:**
```json
{
  "path": "src/auth/login.ts"
}
```

#### write_to_file

Create a new file or overwrite existing file.

**Parameters:**
```json
{
  "path": "string",     // File path to create/overwrite
  "content": "string"   // File content
}
```

**Example:**
```json
{
  "path": "src/utils/helper.ts",
  "content": "export function helper() { return true; }"
}
```

#### replace_in_file

Replace text in an existing file.

**Parameters:**
```json
{
  "path": "string",      // File path
  "old_string": "string", // Text to find (exact match)
  "new_string": "string"  // Replacement text
}
```

**Example:**
```json
{
  "path": "config.json",
  "old_string": "\"version\": \"1.0.0\"",
  "new_string": "\"version\": \"1.1.0\""
}
```

### Search Operations

#### search_files

Search file contents using ripgrep.

**Parameters:**
```json
{
  "path": "string",        // Directory to search
  "regex": "string",       // Search pattern
  "file_pattern": "string" // Optional file filter (e.g., "*.ts")
}
```

**Example:**
```json
{
  "path": "src",
  "regex": "function.*auth",
  "file_pattern": "*.ts"
}
```

#### list_files

List directory contents.

**Parameters:**
```json
{
  "path": "string",     // Directory path
  "recursive": boolean  // Include subdirectories
}
```

**Example:**
```json
{
  "path": "src/components",
  "recursive": true
}
```

#### list_code_definition_names

List classes, functions, and variables in a file.

**Parameters:**
```json
{
  "path": "string"  // File or directory path
}
```

**Example:**
```json
{
  "path": "src/auth.ts"
}
```

### Terminal Operations

#### execute_command

Execute a command in the integrated terminal.

**Parameters:**
```json
{
  "command": "string",      // Command to execute
  "description": "string",  // Human-readable description
  "wait_for_completion": boolean  // Wait for command to finish
}
```

**Example:**
```json
{
  "command": "npm test",
  "description": "Run test suite",
  "wait_for_completion": true
}
```

### Browser Operations

#### browser_action

Control a headless browser.

**Parameters:**
```json
{
  "action": "string",     // launch, click, type, scroll, close
  "url": "string",        // URL (for launch)
  "coordinate": "string", // x,y (for click)
  "text": "string",       // Text (for type)
  "scroll_amount": number // Pixels (for scroll)
}
```

**Examples:**
```json
// Launch browser
{
  "action": "launch",
  "url": "http://localhost:3000"
}

// Click element
{
  "action": "click",
  "coordinate": "100,200"
}

// Type text
{
  "action": "type",
  "text": "username"
}

// Close browser
{
  "action": "close"
}
```

### Conversation Operations

#### ask_followup_question

Ask the user for clarification.

**Parameters:**
```json
{
  "question": "string"  // Question to ask
}
```

#### attempt_completion

Mark task as complete with summary.

**Parameters:**
```json
{
  "result": "string",      // Completion summary
  "command": "string"      // Optional: Suggested terminal command
}
```

---

## Provider Configuration

### Anthropic

```json
{
  "cline.apiProvider": "anthropic",
  "cline.anthropic.apiKey": "sk-ant-api03-...",
  "cline.anthropic.model": "claude-3-7-sonnet-20250219"
}
```

**Available Models:**
- `claude-3-7-sonnet-20250219` - Latest Sonnet (recommended)
- `claude-3-5-sonnet-20241022` - Sonnet 3.5
- `claude-3-opus-20240229` - Most capable
- `claude-3-haiku-20240307` - Fastest

### OpenAI

```json
{
  "cline.apiProvider": "openai",
  "cline.openai.apiKey": "sk-...",
  "cline.openai.model": "gpt-4o",
  "cline.openai.baseUrl": "https://api.openai.com/v1"
}
```

**Available Models:**
- `gpt-4o` - Latest multimodal model
- `gpt-4o-mini` - Cost-effective
- `gpt-4-turbo` - Previous generation
- `gpt-3.5-turbo` - Fast, economical

### OpenRouter

```json
{
  "cline.apiProvider": "openrouter",
  "cline.openrouter.apiKey": "sk-or-...",
  "cline.openrouter.model": "anthropic/claude-3.7-sonnet"
}
```

**Benefits:**
- Access to 100+ models
- Unified API
- Cost comparison
- Fallback models

### Local Models (Ollama)

```json
{
  "cline.apiProvider": "ollama",
  "cline.ollama.baseUrl": "http://localhost:11434",
  "cline.ollama.model": "codellama:13b"
}
```

**Setup:**
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull codellama:13b
ollama pull deepseek-coder:6.7b
ollama pull qwen2.5-coder:14b

# Start server
ollama serve
```

### VS Code LM API (GitHub Copilot)

```json
{
  "cline.apiProvider": "vscode-lm"
}
```

**Requirements:**
- GitHub Copilot subscription
- VS Code Copilot extension installed
- Signed into GitHub account

---

## MCP Server Configuration

### Configuration File

Location: `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`

### Schema

```json
{
  "mcpServers": {
    "server-name": {
      "command": "string",       // Executable command
      "args": ["string"],        // Command arguments
      "env": {                   // Environment variables
        "KEY": "value"
      },
      "disabled": boolean,       // Enable/disable server
      "autoApprove": ["string"]  // Tools to auto-approve
    }
  }
}
```

### Example Configurations

#### GitHub MCP

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": {
        "GITHUB_TOKEN": "ghp_..."
      },
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

#### PostgreSQL MCP

```json
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "postgresql://user:pass@localhost/db"
      },
      "disabled": false,
      "autoApprove": ["query"]
    }
  }
}
```

#### Brave Search MCP

```json
{
  "mcpServers": {
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "BS..."
      },
      "disabled": false,
      "autoApprove": ["brave_web_search"]
    }
  }
}
```

#### Filesystem MCP

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/dir"],
      "disabled": false,
      "autoApprove": ["read_file"]
    }
  }
}
```

#### Context7 MCP (Documentation)

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"],
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

---

## .clinerules Format

Project-specific instructions file placed in project root.

### Basic Structure

```markdown
# Project Context

## Overview
Brief description of the project

## Tech Stack
- Framework: React 18
- Language: TypeScript
- Styling: Tailwind CSS
- State: Redux Toolkit

## Common Commands
- `npm run dev` - Start development server
- `npm run build` - Production build
- `npm test` - Run tests
- `npm run lint` - Run ESLint

## Architecture
- All components in src/components/
- Custom hooks in src/hooks/
- API calls in src/services/
- Types in src/types/

## Style Guidelines
- Use functional components
- Prefer named exports
- Use strict TypeScript
- Follow existing file structure
- Use async/await, not callbacks

## Important Files
- src/main.tsx - Entry point
- src/App.tsx - Root component
- src/store/ - Redux store configuration
```

### Advanced Rules

```markdown
## Code Patterns

### Component Structure
```tsx
interface Props {
  // Define props here
}

export function ComponentName({ prop }: Props) {
  // Implementation
}
```

### Error Handling
- Always use try/catch for async operations
- Log errors to console with context
- Show user-friendly error messages

### Testing
- Write tests for all new features
- Use React Testing Library
- Mock external API calls
```

---

## Context Mentions

Special syntax for referencing context in chat:

| Syntax | Description | Example |
|--------|-------------|---------|
| `@filename` | Reference a file | `@src/auth.ts` |
| `@folder/` | Reference a directory | `@src/components/` |
| `@url` | Fetch URL content | `@https://api.example.com/docs` |
| `@problems` | Add VS Code problems panel | `@problems` |

---

## Environment Variables

| Variable | Description | Used By |
|----------|-------------|---------|
| `ANTHROPIC_API_KEY` | Anthropic API key | Anthropic provider |
| `OPENAI_API_KEY` | OpenAI API key | OpenAI provider |
| `OPENROUTER_API_KEY` | OpenRouter API key | OpenRouter provider |
| `GITHUB_TOKEN` | GitHub access token | GitHub MCP |
| `DATABASE_URL` | PostgreSQL connection | Postgres MCP |
| `BRAVE_API_KEY` | Brave Search API key | Brave MCP |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [References](./REFERENCES.md) - External resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
