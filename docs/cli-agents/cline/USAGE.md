# Cline - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation & Setup](#installation--setup)
3. [Common Workflows](#common-workflows)
4. [Advanced Features](#advanced-features)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

---

## Getting Started

### What is Cline?

Cline is an autonomous AI coding agent embedded in VS Code. Unlike traditional code completion tools, Cline can:

- **Create and edit files** with diff preview
- **Execute terminal commands** with your approval
- **Browse websites** using headless browser
- **Use MCP tools** to extend capabilities
- **Work step-by-step** with human-in-the-loop approval

### First-Time Setup

1. **Install the Extension**
   ```
   VS Code → Extensions (Ctrl+Shift+X) → Search "Cline" → Install
   ```

2. **Configure API Provider**
   ```
   Click Cline icon → Settings gear → Select Provider → Enter API Key
   ```

3. **Choose Your Model**
   - **Claude 3.7 Sonnet** (recommended) - Best reasoning and coding
   - **GPT-4o** - Good multimodal capabilities
   - **DeepSeek Coder** - Cost-effective option
   - **Local models** - Privacy-focused (via Ollama)

4. **Start Your First Task**
   ```
   Type in chat: "Explain this codebase to me"
   ```

---

## Installation & Setup

### Method 1: VS Code Marketplace (Recommended)

1. Open VS Code
2. Press `Ctrl+Shift+X` (Windows/Linux) or `Cmd+Shift+X` (Mac)
3. Search for "Cline"
4. Click "Install" by Cline Bot Inc.
5. Reload VS Code if prompted

### Method 2: Command Line

```bash
# Using VS Code CLI
code --install-extension saoudrizwan.claude-dev

# Using VSIX file (for beta versions)
code --install-extension cline-x.x.x.vsix
```

### Method 3: OpenVSX (VSCodium, etc.)

```bash
# For VSCodium and compatible editors
codium --install-extension saoudrizwan.claude-dev
```

### Provider Configuration

#### Anthropic (Recommended)

1. Get API key from [console.anthropic.com](https://console.anthropic.com)
2. In Cline settings, select "Anthropic"
3. Paste your API key
4. Select model (Claude 3.7 Sonnet recommended)

#### OpenRouter (Budget-Friendly)

1. Create account at [openrouter.ai](https://openrouter.ai)
2. Add credit ($5 minimum)
3. Generate API key
4. In Cline settings, select "OpenRouter"
5. Paste API key
6. Choose from 100+ models

#### GitHub Copilot (Free Tier)

1. Ensure GitHub Copilot extension is installed
2. Sign into GitHub account in VS Code
3. In Cline settings, select "VS Code LM API"
4. Models are available based on Copilot subscription

#### Ollama (Local/Free)

```bash
# 1. Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 2. Pull a coding model
ollama pull qwen2.5-coder:14b
ollama pull deepseek-coder:6.7b
ollama pull codellama:13b

# 3. Start Ollama server
ollama serve

# 4. In Cline settings, select "Ollama"
# 5. Base URL: http://localhost:11434
# 6. Select your model
```

### Positioning Cline for Best Experience

**Recommended: Right Sidebar**

```
1. Click and drag Cline icon from left sidebar
2. Drop in right sidebar area
3. This keeps file explorer visible while chatting
```

This layout provides:
- Left: File explorer
- Center: Code editor
- Right: Cline chat

---

## Common Workflows

### 1. Code Exploration

**Understanding a New Codebase**

```
> "What does this project do?"

Cline will:
1. Read README.md
2. Analyze package.json / dependencies
3. Explore directory structure
4. Identify main components/modules
5. Summarize architecture
```

**Finding Specific Code**

```
> "Find where authentication is handled"

Cline will:
1. Search for auth-related files
2. Read relevant source files
3. Explain the authentication flow
4. Show key files: @src/auth/login.ts
```

**Understanding Complex Functions**

```
> "Explain the logic in @src/utils/cache.ts"

Cline will:
1. Read the file
2. Break down complex algorithms
3. Explain data structures
4. Suggest improvements
```

### 2. Code Generation

**Creating New Components**

```
> "Create a React component for a user profile card"

Cline will:
1. Check existing component structure
2. Create UserProfileCard.tsx
3. Include TypeScript interfaces
4. Add styling (CSS/styled-components)
5. Show diff for approval
```

**Implementing Features**

```
> "Add JWT authentication middleware to Express"

Cline will:
1. Check existing middleware structure
2. Install required packages (jsonwebtoken)
3. Create auth.middleware.ts
4. Implement JWT verification
5. Add error handling
6. Export for use in routes
```

**API Integration**

```
> "Create a service to fetch data from /api/users"

Cline will:
1. Check existing API service patterns
2. Create users.service.ts
3. Implement fetch with error handling
4. Add TypeScript types
5. Export functions
```

### 3. Code Refactoring

**Modernizing Code**

```
> "Convert these callbacks to async/await"

Cline will:
1. Identify callback patterns
2. Convert to Promise-based
3. Add async/await syntax
4. Handle error cases
5. Maintain functionality
```

**Improving Type Safety**

```
> "Add TypeScript types to this JavaScript file"

Cline will:
1. Analyze function signatures
2. Create interface definitions
3. Add type annotations
4. Handle edge cases
```

**Performance Optimization**

```
> "Optimize this React component rendering"

Cline will:
1. Analyze component structure
2. Add React.memo if beneficial
3. Optimize useEffect dependencies
4. Suggest useMemo/useCallback
```

### 4. Testing & Debugging

**Writing Tests**

```
> "Write unit tests for the auth module"

Cline will:
1. Check testing framework (Jest/Vitest)
2. Create auth.test.ts
3. Write test cases for each function
4. Include edge cases
5. Mock dependencies
```

**Debugging Issues**

```
> "Fix the failing test in users.test.ts"

Cline will:
1. Run the test to see error
2. Analyze the failing code
3. Identify root cause
4. Implement fix
5. Verify test passes
```

**Adding Error Handling**

```
> "Add proper error handling to this API call"

Cline will:
1. Identify unhandled promises
2. Add try/catch blocks
3. Implement error logging
4. Add user feedback
```

### 5. Git Workflows

**Committing Changes**

```
> "Commit these changes with a good message"

Cline will:
1. Run git status
2. Review diffs
3. Stage appropriate files
4. Write descriptive commit message
5. Execute commit
```

**Creating Pull Requests**

```
> "Create a PR for these changes"

Cline will:
1. Create branch if on main
2. Commit uncommitted changes
3. Push to origin
4. Open PR with gh CLI or URL
5. Fill in PR description
```

**Code Review**

```
> "Review the changes in this branch"

Cline will:
1. Show git diff
2. Analyze changes
3. Suggest improvements
4. Check for potential issues
```

### 6. Project Setup

**Initializing New Projects**

```
> "Set up a new Next.js project with TypeScript"

Cline will:
1. Run create-next-app
2. Configure TypeScript
3. Set up ESLint and Prettier
4. Install recommended packages
5. Create folder structure
```

**Adding Dependencies**

```
> "Install and configure Tailwind CSS"

Cline will:
1. Install tailwindcss package
2. Initialize Tailwind config
3. Update CSS imports
4. Configure content paths
```

**Configuration Management**

```
> "Set up environment variables for this project"

Cline will:
1. Create .env.example
2. Add to .gitignore
3. Document required variables
4. Create type definitions
```

### 7. Browser Automation

**Visual Testing**

```
> "Test the login page and take a screenshot"

Cline will:
1. Launch headless browser
2. Navigate to login page
3. Interact with form
4. Capture screenshot
5. Close browser
```

**End-to-End Testing**

```
> "Test the complete checkout flow"

Cline will:
1. Launch browser
2. Add item to cart
3. Proceed to checkout
4. Fill shipping info
5. Verify confirmation page
```

**Debugging UI Issues**

```
> "Check why this button is not clickable"

Cline will:
1. Launch site in browser
2. Inspect element
3. Check CSS properties
4. Identify z-index or overlay issues
```

---

## Advanced Features

### 1. Auto-Approve Settings

Configure automatic approvals for trusted operations:

```
Settings → Cline → Auto-approve:
☑️ Read files (safe)
☑️ Edit files (with restrictions)
☑️ Execute safe commands
☐ Use browser
```

**Recommended Configuration:**
- **Development**: Enable read files, safe commands
- **Production**: Minimal auto-approval
- **New projects**: Conservative settings

### 2. MCP Server Integration

**Installing MCP Servers:**

```
1. Click MCP Server button (server icon)
2. Click "Edit MCP Settings"
3. Add server configuration
4. Save and Cline auto-detects
```

**Example: GitHub MCP**

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": { "GITHUB_TOKEN": "ghp_..." }
    }
  }
}
```

Then use in chat:
```
> "List my open issues"
> "Create an issue for this bug"
> "Review PR #123"
```

### 3. Checkpoints

**Using Checkpoints:**

```
During task execution:
1. Cline creates checkpoint at each step
2. Click "Compare" to see changes
3. Click "Restore" to rollback

Restore options:
- Workspace only: Revert files, keep chat
- Task and workspace: Full rollback
```

### 4. Context7 for Documentation

**Setup:**
```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    }
  }
}
```

**Usage:**
```
> "Using Context7, get the latest React docs for hooks"
> "Check the Next.js 14 documentation for App Router"
```

### 5. Custom Instructions (.clinerules)

Create a `.clinerules` file in project root:

```markdown
# Project Context

## Tech Stack
- Next.js 14
- TypeScript
- Tailwind CSS
- Prisma ORM

## Commands
- `npm run dev` - Dev server
- `npm run build` - Production build
- `npm run db:migrate` - Database migrations

## Conventions
- Use server components by default
- Client components in components/client/
- Database queries in lib/db/
- API routes in app/api/
```

### 6. Context References

Use special syntax in chat:

| Syntax | Purpose |
|--------|---------|
| `@src/auth.ts` | Include file content |
| `@src/components/` | Include directory structure |
| `@https://example.com/api` | Fetch URL content |
| `@problems` | Include VS Code problems |

---

## Best Practices

### 1. Effective Communication

**Be Specific:**
```
✅ "Add email validation to the signup form in src/auth/signup.tsx"
❌ "Fix the form"
```

**Provide Context:**
```
✅ "This is a React Native app. Add a camera component."
❌ "Add a camera"
```

**Break Down Complex Tasks:**
```
✅ First: "Create the database schema"
✅ Then: "Add the API endpoints"
✅ Finally: "Create the frontend forms"
```

### 2. Security Best Practices

**Review Before Approving:**
- Always check file diffs
- Review terminal commands
- Verify destructive operations

**Use Checkpoints:**
- Create checkpoints before major changes
- Restore if something goes wrong
- Compare to understand changes

**Protect Sensitive Data:**
- Never commit API keys
- Use environment variables
- Add .env to .gitignore

### 3. Cost Management

**Monitor Usage:**
```
Check Cline settings for token usage
Use local models for simple tasks
Switch to cheaper models for testing
```

**Optimize Context:**
- Clear chat when switching tasks
- Use specific file references
- Avoid including entire directories

**Provider Selection:**
| Task | Recommended Provider |
|------|---------------------|
| Complex reasoning | Claude 3.7 |
| Simple edits | GPT-4o-mini |
| Testing/experiments | Local Ollama |
| Budget mode | OpenRouter with Qwen |

### 4. Workflow Optimization

**Use Keyboard Shortcuts:**
- `Cmd+'` / `Ctrl+'` - Add selection to chat
- `Cmd+'` / `Ctrl+'` (no selection) - Focus chat input

**Organize Projects:**
- Create .clinerules for each project
- Document common commands
- Set up MCP servers per project

**Batch Operations:**
```
> "Fix all TypeScript errors in src/components"
> "Update all imports to use the new path alias"
```

---

## Troubleshooting

### Common Issues

**1. API Key Not Working**

```
Symptom: "API key invalid" error
Solution:
1. Verify key is copied correctly
2. Check key has necessary permissions
3. Ensure billing is set up (if required)
4. Try regenerating the key
```

**2. Model Not Responding**

```
Symptom: Cline hangs or times out
Solution:
1. Check internet connection
2. Verify API provider status
3. Try a different model
4. Restart VS Code
```

**3. MCP Server Not Connecting**

```
Symptom: "Failed to connect to MCP server"
Solution:
1. Verify MCP settings JSON is valid
2. Check required environment variables
3. Ensure npx command works: npx -y package-name
4. Restart Cline extension
```

**4. Terminal Commands Fail**

```
Symptom: "Command not found" or permission errors
Solution:
1. Check command exists in PATH
2. Verify working directory is correct
3. Try with absolute paths
4. Check file permissions
```

**5. Browser Automation Not Working**

```
Symptom: Browser actions fail
Solution:
1. Ensure browser is launched first
2. Check URL is accessible
3. Verify Chrome/Chromium is installed
4. Try with headless: false for debugging
```

### Getting Help

**Within Cline:**
```
> "I'm getting error X, how do I fix it?"
> "Explain what went wrong here"
```

**Community Resources:**
- [Discord](https://discord.gg/cline)
- [GitHub Issues](https://github.com/cline/cline/issues)
- [Reddit r/cline](https://www.reddit.com/r/cline/)

**Official Documentation:**
- [docs.cline.bot](https://docs.cline.bot)

---

## Quick Reference

### Essential Commands

| Command | Purpose |
|---------|---------|
| Install | `ext install saoudrizwan.claude-dev` |
| Open Cline | Click Cline icon or `Cmd+Shift+P` → "Cline" |
| Add to chat | Select text → `Cmd+'` |
| Focus chat | `Cmd+'` (no selection) |

### Provider Quick Setup

| Provider | Setup Time | Cost |
|----------|-----------|------|
| Anthropic | 2 min | Pay per use |
| OpenRouter | 5 min | Pay per use |
| GitHub Copilot | 1 min | Subscription |
| Ollama | 15 min | Free |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd+'` / `Ctrl+'` | Add selection / Focus chat |
| `Enter` | Submit message |
| `Shift+Enter` | New line in input |
| `Escape` | Cancel operation |

---

*For more details, see the [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md) documentation.*
