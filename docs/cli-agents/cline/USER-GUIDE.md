# Cline User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [VS Code Commands](#vs-code-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: VS Code Marketplace (Recommended)
```bash
code --install-extension saoudrizwan.claude-dev
```

### Method 2: Within VS Code
1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X)
3. Search for "Cline"
4. Click Install

### Method 3: OpenVSX
```bash
code --install-extension claude-dev --open-vsx
```

## Quick Start

```bash
# After installation, open VS Code
# Access Cline via:
# - Sidebar icon
# - Command Palette: Ctrl+Shift+P → "Cline: Open"
# - Keyboard shortcut: Ctrl+Shift+L
```

In VS Code:
1. Open Command Palette (Ctrl+Shift+P)
2. Type "Cline: Open"
3. Start chatting with Claude

## VS Code Commands

Access via Command Palette (Ctrl+Shift+P):

| Command | Description |
|---------|-------------|
| Cline: Open | Open Cline panel |
| Cline: New Task | Start new task |
| Cline: Settings | Open settings |
| Cline: Add to Context | Add file to context |
| Cline: Export Chat | Export conversation |

### Command: Open
**Description:** Open Cline sidebar panel

**Usage:**
```
Ctrl+Shift+P → "Cline: Open"
# or
Ctrl+Shift+L (customizable)
```

### Command: New Task
**Description:** Start a new conversation

**Usage:**
```
Ctrl+Shift+P → "Cline: New Task"
```

### Command: Settings
**Description:** Configure Cline

**Usage:**
```
Ctrl+Shift+P → "Cline: Settings"
```

## TUI/Interactive Commands

In the Cline chat panel, use these commands:

| Command | Description |
|---------|-------------|
| @file | Reference a file |
| @folder | Reference a folder |
| @code | Reference specific code |
| @terminal | Reference terminal output |
| @problems | Reference problems panel |
| @web | Search the web |

### Command: @file
**Description:** Include file in context

**Usage:**
```
@file src/main.js
Explain what this file does
```

### Command: @folder
**Description:** Include entire folder

**Usage:**
```
@folder src/components
Review all components in this folder
```

### Command: @code
**Description:** Reference specific code block

**Usage:**
```
Select code in editor, then:
@code
Refactor this to use async/await
```

### Command: @web
**Description:** Search web for information

**Usage:**
```
@web React 19 release date
```

## Configuration

### Settings (VS Code settings.json)

```json
{
  "cline.apiKey": "your-api-key",
  "cline.model": "claude-3-5-sonnet-20241022",
  "cline.maxRequestsPerTask": 25,
  "cline.browserViewport": {
    "width": 900,
    "height": 600
  },
  "cline.terminalOutputLineLimit": 500,
  "cline.fuzzyMatchThreshold": 1.0,
  "cline.preferredLanguage": "English",
  "cline.customInstructions": "",
  "cline.autoApprove": false,
  "cline.readOnlyMode": false
}
```

### Configuration File (.clinerules)

Create `.clinerules` in project root:

```markdown
# Project Rules for Cline

## Code Style
- Use TypeScript for all new files
- Follow ESLint configuration
- Write tests for all functions

## Architecture
- Use functional components
- Prefer hooks over HOCs
- Keep components under 200 lines
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| ANTHROPIC_API_KEY | Claude API key |
| OPENAI_API_KEY | OpenAI API key (alternative) |
| CLINE_API_KEY | Cline-specific key |

## Usage Examples

### Example 1: Code Explanation
```
@file src/utils/api.ts
Explain the error handling in this file
```

### Example 2: Refactoring
```
@file src/components/Button.tsx
@file src/components/Input.tsx
Refactor these to use a shared base component
```

### Example 3: Bug Fix
```
@file src/hooks/useData.ts
@terminal
The fetch is failing. Fix the error handling.
```

### Example 4: Feature Implementation
```
Create a new user settings page with:
- Profile section
- Notification preferences
- Theme selector
Use the existing design system.
```

### Example 5: Testing
```
@folder src/utils
Generate unit tests for all utility functions
```

## Troubleshooting

### Issue: API Key Not Set
**Solution:**
1. Open Cline settings
2. Add your Anthropic API key
3. Or set `ANTHROPIC_API_KEY` environment variable

### Issue: Cline Not Responding
**Solution:**
```bash
# Reload VS Code window
Ctrl+Shift+P → "Developer: Reload Window"
```

### Issue: Files Not Found
**Solution:**
- Ensure file paths are correct
- Use @file with relative paths from workspace root
- Check that file is in workspace

### Issue: Browser Actions Failing
**Solution:**
- Check browser viewport settings
- Ensure URL is accessible
- Try increasing timeout in settings

---

**Last Updated:** 2026-04-02
