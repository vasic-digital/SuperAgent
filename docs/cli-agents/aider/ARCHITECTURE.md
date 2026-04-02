# Aider - Architecture Documentation

## System Overview

Aider is a Python-based AI pair programming tool that integrates with git repositories and multiple LLM providers to provide intelligent code editing capabilities.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            AIDER ARCHITECTURE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   Terminal UI    │◄────►│   Aider Core     │◄────►│   LLM Provider   │   │
│  │  (prompt-toolkit)│      │   (Python)       │      │ (Claude/GPT/etc) │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Edit Format     │    │  Repository      │    │  Git Integration │       │
│  │  (Coders)        │    │  Mapper          │    │  (GitPython)     │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Terminal UI Layer

**Technology Stack:**
- **Framework**: prompt-toolkit for interactive terminal handling
- **Features**:
  - Command history and search (Ctrl-R)
  - Emacs/Vi keybindings
  - Syntax highlighting via rich console
  - Multi-line input support
  - Tab completion

**Key UI Elements:**
```
┌──────────────────────────────────────────────────────┐
│ Aider v0.78.0                                [Model] │
├──────────────────────────────────────────────────────┤
│                                                      │
│ > User request here                                  │
│                                                      │
│ The LLM will respond with code edits shown as diffs  │
│                                                      │
│ [Diff output shown...]                               │
│                                                      │
│ Commit message: Updated factorial function           │
│ Committed changes                                    │
│                                                      │
│ > _                                                  │
└──────────────────────────────────────────────────────┘
```

### 2. Core Engine (`main.py`)

The central orchestrator manages:
- **Initialization**: Parse arguments, setup models, initialize repo
- **Event Loop**: Process user input, dispatch to coders
- **Git Integration**: Track changes, create commits
- **Context Management**: File context, conversation history

### 3. Coder System (`coders/`)

**Edit Format Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                     Coder Hierarchy                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  BaseCoder (base_coder.py)                                  │
│       │                                                     │
│       ├── EditBlockCoder (search/replace blocks)            │
│       ├── WholeFileCoder (complete file replacement)        │
│       ├── UdiffCoder (unified diff format)                  │
│       ├── ArchitectCoder (2-model architecture)             │
│       ├── AskCoder (Q&A without edits)                      │
│       └── ContextCoder (auto file identification)           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Edit Formats:**

| Format | Description | Best For |
|--------|-------------|----------|
| `diff` | Search/replace blocks | Most models, default |
| `whole` | Complete file rewrite | Small files, simple changes |
| `udiff` | Unified diff format | Compatibility |
| `architect` | Design + edit models | Complex changes |
| `editor` | Editor-style diff | Specific use cases |
| `ask` | No edits, only Q&A | Code understanding |

### 4. Repository Mapper (`repomap.py`)

**Purpose**: Create a compact representation of the entire codebase for context.

**How it Works:**
1. Uses tree-sitter to parse source files
2. Extracts identifiers (functions, classes, variables)
3. Builds a ranked map of code structure
4. Provides context without sending entire files

```
Repository Map Structure:
├── file1.py
│   ├── ClassName
│   │   ├── method1()
│   │   └── method2()
│   └── function1()
├── file2.py
│   └── function2()
└── [More files...]
```

### 5. Git Integration (`repo.py`)

**Features:**
- Auto-commit with descriptive messages
- Track which changes were made by Aider
- Undo functionality via git
- Branch awareness
- `.gitignore` management for `.aider*` files

**Commit Flow:**
```
User Request → Coder Generates Edits → Apply to Files
                                            │
                                            ▼
                                    Stage Changes
                                            │
                                            ▼
                                    Generate Commit Message
                                            │
                                            ▼
                                    Create Git Commit
```

### 6. Model System (`models.py`)

**Model Configuration:**
- Main Model: Primary LLM for coding tasks
- Weak Model: For commit messages and summaries
- Editor Model: For architect mode editing

**Supported Providers:**
- Anthropic (Claude)
- OpenAI (GPT-4, o1, o3-mini)
- DeepSeek
- Google (Gemini)
- Azure OpenAI
- AWS Bedrock
- OpenRouter
- Ollama (local)
- And more...

### 7. Command System (`commands.py`)

**Slash Commands Architecture:**

```python
class Commands:
    def cmd_add(self, args):      # Add files to chat
    def cmd_drop(self, args):     # Remove files from chat
    def cmd_model(self, args):    # Switch models
    def cmd_ask(self, args):      # Ask mode
    def cmd_undo(self, args):     # Undo last commit
    # ... and more
```

---

## Data Flow

### Request-Response Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         REQUEST LIFECYCLE                               │
└─────────────────────────────────────────────────────────────────────────┘

1. USER INPUT
   └─► Natural language command received at prompt

2. CONTEXT ASSEMBLY
   ├─► Files added to chat (explicit or auto-detected)
   ├─► Repository map (if enabled)
   ├─► Chat history (summarized if needed)
   └─► System prompts and instructions

3. LLM REQUEST
   └─► Streaming request via litellm library
       ├─► Model selection
       ├─► Temperature/top-p settings
       └─► Reasoning parameters (if supported)

4. RESPONSE PROCESSING
   └─► Parse edit format from response
       ├─► Extract code blocks
       ├─► Apply edits to files
       └─► Handle errors/refusals

5. EDIT APPLICATION
   └─► Apply changes to files
       ├─► Verify edits applied correctly
       ├─► Run linters (if configured)
       └─► Run tests (if configured)

6. GIT COMMIT
   └─► Stage modified files
       ├─► Generate commit message (weak model)
       └─► Create commit

7. FOLLOW-UP
   └─► Model may ask for clarification
   └─► Or confirm completion
```

### Context Management

**Context Window Strategy:**
```
┌────────────────────────────────────────────────────────────┐
│                    CONTEXT WINDOW                          │
├────────────────────────────────────────────────────────────┤
│ System Prompt                                              │
│ ├─► Base instructions                                      │
│ ├─► Edit format examples                                   │
│ └─► Tool definitions                                       │
├────────────────────────────────────────────────────────────┤
│ Repository Map (if enabled)                                │
│ └─► ~1000 tokens of codebase structure                     │
├────────────────────────────────────────────────────────────┤
│ Chat Files (added to session)                              │
│ └─► Full content of relevant files                         │
├────────────────────────────────────────────────────────────┤
│ Conversation History                                       │
│ ├─► Recent messages (full)                                 │
│ └─► Older messages (summarized by weak model)              │
└────────────────────────────────────────────────────────────┘
```

---

## Configuration Architecture

### Configuration Hierarchy

```
Configuration Precedence (lowest to highest):

1. Default Settings
   └─► Built into aider source

2. Global Config
   └─► ~/.aider.conf.yml

3. Project Config
   └─► .aider.conf.yml (in git root)

4. Environment Variables
   └─► AIDER_xxx variables

5. .env File
   └─► Variables in .env

6. Command Line Arguments
   └─► --flags
```

### Key Configuration Files

| File | Purpose | Location |
|------|---------|----------|
| `.aider.conf.yml` | Main configuration | Project root or home |
| `.aider.model.settings.yml` | Model-specific settings | Project root |
| `.aider.model.metadata.json` | Model metadata (context, cost) | Project root |
| `.env` | API keys and secrets | Project root |
| `.aider.chat.history.md` | Chat history | Project root |

---

## Extension Points

### 1. Custom Model Settings

Define unknown models in `.aider.model.settings.yml`:
```yaml
- name: custom-model
  edit_format: diff
  weak_model_name: gpt-4o-mini
  use_repo_map: true
  send_undo_reply: true
```

### 2. Linting Integration

Configure linters to run on AI-generated code:
```bash
aider --lint-cmd "python -m flake8"
aider --lint-cmd "eslint"
```

### 3. Testing Integration

Auto-run tests on changes:
```bash
aider --test-cmd "pytest"
aider --auto-test
```

---

## Performance Considerations

### Optimization Strategies

1. **Prompt Caching**: Reduces API costs for repeated context
2. **Repository Mapping**: Provides context without full file content
3. **Chat History Summarization**: Keeps context window manageable
4. **Streaming Responses**: Real-time display of LLM output

### Resource Management

```
Memory Usage:
├─► Repository map (cached tree-sitter data)
├─► Chat history (summarized periodically)
├─► File contents (only added files)
└─► LLM response buffers

API Usage:
├─► Main model calls (code generation)
├─► Weak model calls (summaries, commit messages)
└─► Token caching (where supported)
```

---

## Security Architecture

### API Key Handling

- Keys stored in `.env` file
- `.env` automatically added to `.gitignore`
- Keys never logged or transmitted beyond LLM APIs
- Support for multiple provider keys

### Safe Defaults

- Confirmation before destructive operations
- Git integration for change tracking
- Undo capability via `/undo`
- Protected file patterns

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
