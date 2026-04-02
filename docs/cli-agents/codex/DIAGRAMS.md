# OpenAI Codex - Diagrams

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CODEX CLI                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │    User      │◄──►│  Codex CLI   │◄──►│ OpenAI API   │   │
│  │   Terminal   │    │   (Rust)     │    │  Responses   │   │
│  └──────────────┘    └──────┬───────┘    └──────────────┘   │
│                             │                                │
│        ┌────────────────────┼────────────────────┐          │
│        │                    │                    │          │
│        ▼                    ▼                    ▼          │
│  ┌──────────┐        ┌──────────┐        ┌──────────┐      │
│  │ Sandbox  │        │ File Ops │        │  Tools   │      │
│  │(Security)│        │(R/W/Patch│        │(Shell)   │      │
│  └──────────┘        └──────────┘        └──────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Approval Mode Flow

```
User Request
    │
    ├──► Suggest Mode ──► Prompt for writes & commands
    │
    ├──► Auto Edit ─────► Auto-write, prompt for commands
    │
    └──► Full Auto ─────► Auto-execute (sandboxed)
```

## Sandboxing

```
macOS: Seatbelt ──► Read-only jail + Writable roots
Linux: Docker ────► Container + iptables firewall
```

---

*See [ARCHITECTURE.md](./ARCHITECTURE.md) for details*
