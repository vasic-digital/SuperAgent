# Warp - Architecture

## System Overview

Warp provides AI-powered coding assistance through a CLI interface.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│              Warp Architecture                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │   Terminal   │◄──►│    Agent     │◄──►│   LLM/API    │   │
│  │    (User)    │    │    (Core)    │    │  (Provider)  │   │
│  └──────────────┘    └──────┬───────┘    └──────────────┘   │
│                             │                                │
│        ┌────────────────────┼────────────────────┐          │
│        │                    │                    │          │
│        ▼                    ▼                    ▼          │
│  ┌──────────┐        ┌──────────┐        ┌──────────┐      │
│  │  Tools   │        │  Files   │        │  Config  │      │
│  └──────────┘        └──────────┘        └──────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Terminal Interface
- User interaction layer
- Command parsing
- Output formatting

### 2. Core Engine
- Request processing
- Context management
- Tool orchestration

### 3. LLM Integration
- API communication
- Model management
- Response handling

## Data Flow

1. User input → Terminal
2. Input processing → Core
3. Context assembly → LLM
4. Response → Tool execution
5. Results → User display

---

*See [README.md](./README.md) for overview*
