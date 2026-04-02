# Cline Documentation

Comprehensive documentation for [Cline](https://cline.bot) - the autonomous AI coding agent for VS Code.

## Overview

Cline is a VS Code extension that brings autonomous AI coding capabilities directly into your IDE. Unlike traditional code completion tools, Cline can:

- Create and edit files with diff preview and your approval
- Execute terminal commands in your workspace
- Browse websites using a headless browser
- Integrate with external tools via MCP (Model Context Protocol)
- Work step-by-step with human-in-the-loop approval

## Documentation Structure

| Document | Description | Lines |
|----------|-------------|-------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System design, components, VS Code integration | ~550 |
| [API.md](./API.md) | Commands, settings, configuration reference | ~350 |
| [USAGE.md](./USAGE.md) | Workflows, examples, best practices | ~450 |
| [REFERENCES.md](./REFERENCES.md) | External links, tutorials, community resources | ~330 |
| [DIAGRAMS.md](./DIAGRAMS.md) | Visual documentation and architecture diagrams | ~750 |
| [GAP_ANALYSIS.md](./GAP_ANALYSIS.md) | Improvement opportunities and roadmap | ~320 |

## Quick Start

### Installation

1. Open VS Code
2. Press `Ctrl+Shift+X` (Windows/Linux) or `Cmd+Shift+X` (Mac)
3. Search for "Cline"
4. Click "Install"
5. Configure your preferred AI provider

### First Task

```
Open Cline (click Cline icon) → Type: "Explain this codebase to me"
```

## Key Features

| Feature | Description |
|---------|-------------|
| **Multi-Provider** | Use Claude, GPT, Gemini, or local models via Ollama |
| **MCP Support** | Extend capabilities with Model Context Protocol servers |
| **Browser Automation** | Test web apps visually with headless browser |
| **Human Approval** | Safe workflow with approval for destructive operations |
| **Checkpoints** | Save and restore workspace state at any point |
| **Context Aware** | Understands your codebase structure and patterns |

## Supported Providers

| Provider | Setup Difficulty | Cost |
|----------|-----------------|------|
| **Anthropic Claude** | Easy | Pay per use |
| **OpenAI GPT** | Easy | Pay per use |
| **OpenRouter** | Easy | Pay per use (aggregated) |
| **GitHub Copilot** | Very Easy | Subscription |
| **Ollama (Local)** | Medium | Free |

## Configuration

### .clinerules

Create a `.clinerules` file in your project root:

```markdown
# Project Context

## Tech Stack
- React 18
- TypeScript
- Tailwind CSS

## Common Commands
- `npm run dev` - Start development server
- `npm test` - Run tests

## Style Guidelines
- Use functional components
- Prefer named exports
```

### MCP Servers

Configure in VS Code settings (`cline_mcp_settings.json`):

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

## Resources

- [Official Documentation](https://docs.cline.bot)
- [GitHub Repository](https://github.com/cline/cline)
- [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=saoudrizwan.claude-dev)
- [Discord Community](https://discord.gg/cline)
- [Reddit r/cline](https://www.reddit.com/r/cline/)

## Version Information

- **Current Version**: v3.67.1
- **Last Updated**: February 2025
- **VS Code Engine**: ^1.84.0
- **License**: Apache 2.0

---

*Part of the [HelixAgent CLI Agents Documentation](../README.md)*
