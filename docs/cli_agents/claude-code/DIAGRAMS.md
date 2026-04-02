# Claude Code - Diagrams & Visual Documentation

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CLAUDE CODE SYSTEM                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌──────────────┐         ┌──────────────┐         ┌──────────────┐        │
│   │    USER      │◄───────►│  TERMINAL    │◄───────►│ CLAUDE CODE  │        │
│   │   (Developer)│         │     UI       │         │    ENGINE    │        │
│   └──────────────┘         └──────────────┘         └──────┬───────┘        │
│                                                            │                │
│                              ┌─────────────────────────────┼─────────┐      │
│                              │                             │         │      │
│                              ▼                             ▼         ▼      │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ ┌──────────────┐ │
│   │   PLUGINS    │    │    TOOLS     │    │   SESSION    │ │  ANTHROPIC   │ │
│   │  (14 official│    │  (Bash/Edit/ │    │  MANAGER     │ │     API      │ │
│   │   + custom)  │    │  Read/Write) │    │              │ │              │ │
│   └──────────────┘    └──────────────┘    └──────────────┘ └──────────────┘ │
│                              │                                              │
│                              ▼                                              │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│   │     MCP      │    │   FILESYSTEM │    │     GIT      │                  │
│   │   SERVERS    │    │   (Project)  │    │   (Repo)     │                  │
│   └──────────────┘    └──────────────┘    └──────────────┘                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Component Relationships

### Core Engine Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CLAUDE CODE CORE ENGINE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                      API CLIENT                                      │   │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │   │
│   │  │   Request   │─►│   Stream    │─►│   Parse     │─►│   Handle   │  │   │
│   │  │   Builder   │  │   Handler   │  │   Response  │  │   Tools    │  │   │
│   │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│   ┌──────────────────────────────────┼──────────────────────────────────┐   │
│   │                                  ▼                                  │   │
│   │   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐         │   │
│   │   │   CONTEXT    │◄──►│   TOOL       │◄──►│ PERMISSION   │         │   │
│   │   │   MANAGER    │    │   EXECUTOR   │    │   SYSTEM     │         │   │
│   │   └──────────────┘    └──────────────┘    └──────────────┘         │   │
│   │          │                   │                   │                 │   │
│   │          ▼                   ▼                   ▼                 │   │
│   │   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐         │   │
│   │   │   SESSION    │    │    HOOK      │    │   AUTO MODE  │         │   │
│   │   │   STORAGE    │    │   SYSTEM     │    │  CLASSIFIER  │         │   │
│   │   └──────────────┘    └──────────────┘    └──────────────┘         │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Diagrams

### Request Lifecycle

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│          │     │          │     │          │     │          │     │          │
│   USER   │────►│  PARSE   │────►│  CONTEXT │────►│   API    │────►│  STREAM  │
│  INPUT   │     │  INPUT   │     │  ASSEMBLE│     │  REQUEST │     │ RESPONSE │
│          │     │          │     │          │     │          │     │          │
└──────────┘     └──────────┘     └──────────┘     └──────────┘     └────┬─────┘
                                                                          │
                                                                          ▼
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│          │     │          │     │          │     │          │     │          │
│  DISPLAY │◄────│  RESULT  │◄────│  EXECUTE │◄────│ PERMISSION│◄────│  PARSE   │
│  OUTPUT  │     │  RETURN  │     │   TOOL   │     │   CHECK  │     │  TOOL    │
│          │     │          │     │          │     │          │     │  CALL    │
└──────────┘     └──────────┘     └──────────┘     └──────────┘     └──────────┘
```

### Context Assembly Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CONTEXT ASSEMBLY                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   User Input: "Fix the auth bug"                                             │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 1. IDENTIFY RELEVANT FILES                                          │   │
│   │    ├── Grep for "auth" in codebase                                  │   │
│   │    ├── Find auth.ts, login.ts, middleware.ts                        │   │
│   │    └── Identify related tests                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 2. LOAD PROJECT CONTEXT                                             │   │
│   │    ├── Read CLAUDE.md (if exists)                                   │   │
│   │    ├── Load active skills                                           │   │
│   │    └── Apply project-specific rules                                 │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 3. ASSEMBLE CONVERSATION                                            │   │
│   │    ├── System prompt                                                │   │
│   │    ├── Tool definitions                                             │   │
│   │    ├── Previous messages                                            │   │
│   │    └── Current user message                                         │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 4. OPTIMIZE CONTEXT WINDOW                                          │   │
│   │    ├── Check token count                                            │   │
│   │    ├── Compact if needed                                            │   │
│   │    └── Prioritize recent messages                                   │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   Send to Claude API                                                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Plugin System

### Plugin Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         PLUGIN SYSTEM ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                     PLUGIN MANAGER                                   │   │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │   │
│   │  │   Load      │  │  Register   │  │   Enable    │  │  Execute   │  │   │
│   │  │   Plugins   │─►│  Commands   │─►│   Hooks     │─►│   Hooks    │  │   │
│   │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│   ┌──────────────────────────────────┼──────────────────────────────────┐   │
│   │                                  ▼                                  │   │
│   │   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐         │   │
│   │   │              │    │              │    │              │         │   │
│   │   │   COMMANDS   │    │    AGENTS    │    │    HOOKS     │         │   │
│   │   │              │    │              │    │              │         │   │
│   │   │  /command    │    │  Subagent    │    │  PreToolUse  │         │   │
│   │   │  /feature    │    │  Parallel    │    │  PostToolUse │         │   │
│   │   │  /review     │    │  Analysis    │    │  SessionStart│         │   │
│   │   │              │    │              │    │  Stop        │         │   │
│   │   └──────────────┘    └──────────────┘    └──────────────┘         │   │
│   │                                                                     │   │
│   │   ┌──────────────┐    ┌──────────────┐                             │   │
│   │   │              │    │              │                             │   │
│   │   │    SKILLS    │    │    MCP       │                             │   │
│   │   │              │    │   SERVERS    │                             │   │
│   │   │  Context     │    │              │                             │   │
│   │   │  Injection   │    │  External    │                             │   │
│   │   │  Auto-invoke │    │  Tools       │                             │   │
│   │   │              │    │              │                             │   │
│   │   └──────────────┘    └──────────────┘                             │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Hook Event Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           HOOK EXECUTION FLOW                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Event: PreToolUse                                                          │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 1. RECEIVE EVENT                                                    │   │
│   │    {                                                                │   │
│   │      "tool": "Bash",                                                │   │
│   │      "params": {"command": "npm test"},                             │   │
│   │      "context": {...}                                               │   │
│   │    }                                                                │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 2. EVALUATE CONDITIONS                                              │   │
│   │    if (tool == "Bash" && command.includes("rm"))                    │   │
│   │      → Block                                                        │   │
│   │    else                                                             │   │
│   │      → Allow                                                        │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 3. OUTPUT DECISION                                                  │   │
│   │    Exit 0: {"decision": "allow"}                                    │   │
│   │    Exit 2: {"decision": "block", "reason": "..."}                   │   │
│   │    Exit 0: {"decision": "allow", "modified_params": {...}}          │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│   Tool Execution or Block                                                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Tool System

### Tool Permission Decision Tree

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      TOOL PERMISSION DECISION                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│                          Tool Requested                                      │
│                               │                                              │
│                               ▼                                              │
│                    ┌─────────────────────┐                                   │
│                    │   Read / LS / Grep  │                                   │
│                    │   (Safe Tools)      │                                   │
│                    └──────────┬──────────┘                                   │
│                               │ YES                                          │
│                               ▼                                              │
│                       ┌───────────────┐                                      │
│                       │  AUTO-ALLOW   │                                      │
│                       └───────────────┘                                      │
│                               │                                              │
│                               │ NO                                           │
│                               ▼                                              │
│                    ┌─────────────────────┐                                   │
│                    │  PreToolUse Hooks   │                                   │
│                    └──────────┬──────────┘                                   │
│                               │                                              │
│              ┌────────────────┼────────────────┐                             │
│              │                │                │                             │
│              ▼                ▼                ▼                             │
│         ┌────────┐      ┌────────┐      ┌────────┐                         │
│         │ Block  │      │ Allow  │      │Modify  │                         │
│         │ Exit 2 │      │ Exit 0 │      │ Params │                         │
│         └────────┘      └────┬───┘      └────────┘                         │
│                              │                                               │
│                              ▼                                               │
│                    ┌─────────────────────┐                                   │
│                    │   Auto Mode On?     │                                   │
│                    └──────────┬──────────┘                                   │
│                               │                                              │
│              ┌────────────────┼────────────────┐                             │
│              │ NO             │ YES            │                             │
│              ▼                ▼                ▼                             │
│      ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                   │
│      │ Show Prompt  │  │ Classifier   │  │ Show Prompt  │                   │
│      │ to User      │  │ Confidence   │  │ (Low Conf)   │                   │
│      └──────────────┘  └──────┬───────┘  └──────────────┘                   │
│                               │                                              │
│              ┌────────────────┼────────────────┐                             │
│              │ High           │ Low            │                             │
│              ▼                ▼                ▼                             │
│      ┌──────────────┐  ┌──────────────┐                                     │
│      │ AUTO-EXECUTE │  │ SHOW PROMPT  │                                     │
│      └──────────────┘  └──────────────┘                                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Session Management

### Session Lifecycle

```
Timeline:
────────────────────────────────────────────────────────────────────────────►

    Start          Active           Save            Resume           End
      │              │               │                │               │
      ▼              ▼               ▼                ▼               ▼
┌─────────┐    ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ Initialize│    │ Tool    │     │ Persist │     │ Load    │     │ Cleanup │
│ Context   │    │ Calls   │     │ to Disk │     │ from    │     │ Files   │
│ Load      │    │ Message │     │ ~/.claude│    │ ~/.claude    │  Close  │
│ CLAUDE.md │    │ Exchange│     │         │     │         │     │         │
└─────────┘    └─────────┘     └─────────┘     └─────────┘     └─────────┘
      │              │               │                │               │
      │              │               │                │               │
   transcript.jsonl (growing)    saved session    loaded session   archived
                                                                              
────────────────────────────────────────────────────────────────────────────►
```

### Context Window Management

```
Context Window (200K tokens):
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [System: ~2K]                                                              │
│  ├─ Base instructions                                                       │
│  ├─ Tool definitions                                                        │
│  └─ CLAUDE.md content                                                       │
│                                                                             │
│  [History: ~150K]                                                           │
│  ├─ User: "Fix auth bug"                                                    │
│  ├─ Assistant: "I'll help..." + tool calls                                  │
│  ├─ Tool results                                                            │
│  ├─ User: "Also check..."                                                   │
│  ├─ Assistant: ...                                                          │
│  └─ ... (many messages)                                                     │
│                                                                             │
│  [Recent: ~48K]                                                             │
│  ├─ User: "Now add tests"                                                   │
│  ├─ Assistant: "I'll create..."                                             │
│  └─ Tool calls in progress                                                  │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
                    │
                    │ Near limit (90%)
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  COMPACTION TRIGGERED                                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Summarize older messages:                                          │    │
│  │  "Previous work: Fixed auth bug in login.ts, added validation       │    │
│  │   checks, now adding tests..."                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  [Compacted: ~100K]                                                         │
│  ├─ System: ~2K                                                             │
│  ├─ Summary: ~2K                                                            │
│  ├─ Recent history: ~50K                                                    │
│  └─ Current: ~46K                                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## MCP Integration

### MCP Server Communication

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MCP SERVER INTEGRATION                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────────┐                                                      │
│   │   CLAUDE CODE     │                                                      │
│   │   MCP CLIENT      │                                                      │
│   └─────────┬─────────┘                                                      │
│             │                                                                │
│   ┌─────────┴─────────┬─────────┬─────────┐                                  │
│   │                   │         │         │                                  │
│   ▼                   ▼         ▼         ▼                                  │
│  stdio              HTTP       SSE     WebSocket                           │
│   │                   │         │         │                                  │
│   │                   │         │         │                                  │
│   ▼                   ▼         ▼         ▼                                  │
│ ┌─────────┐      ┌─────────┐ ┌─────────┐ ┌─────────┐                       │
│ │ GitHub  │      │ Postgres│ │Brave Srch│ │ Custom  │                       │
│ │  MCP    │      │   MCP   │ │   MCP    │ │  MCP    │                       │
│ └─────────┘      └─────────┘ └─────────┘ └─────────┘                       │
│                                                                              │
│ Transport Examples:                                                          │
│ ┌────────────────────────────────────────────────────────────────────────┐   │
│ │ stdio: ./github-mcp-server                                             │   │
│ │ HTTP:  http://localhost:3000/sse                                       │   │
│ │ SSE:   http://api.example.com/events                                   │   │
│ └────────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Security Model

### Security Layers

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY ARCHITECTURE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Layer 5: User Control                                                      │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                │   │
│   │  │ Approve │  │  Deny   │  │  Undo   │  │  /exit  │                │   │
│   │  └─────────┘  └─────────┘  └─────────┘  └─────────┘                │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Layer 4: Permission System                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Auto Mode Classifier ──► Confidence Score ──► Decision             │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Layer 3: Hook System                                                       │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  PreToolUse Hooks ──► Pattern Matching ──► Block/Allow/Modify       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Layer 2: Protected Directories                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  ~/.ssh  ~/.aws  ~/.gnupg  ~/.claude  /etc  /System                 │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Layer 1: Tool Restrictions                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Read: Auto-allow    Bash: Prompt    Write: Prompt    Edit: Prompt  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## File Organization

### Project Structure Diagram

```
claude-code/
│
├── .claude/                          # Project-specific config
│   ├── commands/                     # Custom slash commands
│   │   ├── commit-push-pr.md
│   │   ├── dedupe.md
│   │   └── triage-issue.md
│   └── settings.json                 # Project settings
│
├── .claude-plugin/                   # Plugin metadata
│   └── marketplace.json
│
├── .github/                          # GitHub integration
│   ├── ISSUE_TEMPLATE/
│   └── workflows/                    # Automation workflows
│       ├── auto-close-duplicates.yml
│       ├── backfill-duplicates.yml
│       └── claude-dedupe-issues.yml
│
├── .devcontainer/                    # Dev container config
│   └── devcontainer.json
│
├── examples/                         # Example configurations
│   ├── hooks/
│   │   └── bash_command_validator_example.py
│   └── settings/
│       ├── settings-bash-sandbox.json
│       ├── settings-lax.json
│       └── settings-strict.json
│
├── plugins/                          # 14 official plugins
│   ├── agent-sdk-dev/
│   ├── claude-opus-4-5-migration/
│   ├── code-review/
│   ├── commit-commands/
│   ├── explanatory-output-style/
│   ├── feature-dev/
│   ├── frontend-design/
│   ├── hookify/
│   ├── learning-output-style/
│   ├── plugin-dev/
│   ├── pr-review-toolkit/
│   ├── ralph-wiggum/
│   └── security-guidance/
│
├── scripts/                          # Automation scripts
│   ├── auto-close-duplicates.ts
│   ├── backfill-duplicate-comments.ts
│   ├── comment-on-duplicates.sh
│   ├── edit-issue-labels.sh
│   ├── gh.sh
│   ├── issue-lifecycle.ts
│   ├── lifecycle-comment.ts
│   └── sweep.ts
│
├── CHANGELOG.md                      # Version history
├── LICENSE.md                        # License
├── README.md                         # Main documentation
└── SECURITY.md                       # Security policy
```

---

*For more information, see the [Architecture Documentation](./ARCHITECTURE.md).*
