# HelixAgent Video Course - Video Metadata Templates

This document provides metadata templates for all 74 videos in the HelixAgent course. Use these templates when publishing videos to YouTube, learning platforms, or internal systems.

---

## Metadata Template Structure

Each video requires the following metadata:

```yaml
video_id: M{module}_{video}
title: "[Module {N}] {Title}"
description: |
  {Full description with learning objectives}
duration: "MM:SS"
level: "Beginner|Intermediate|Advanced"
tags: [list, of, tags]
timestamps:
  - "00:00 - Introduction"
  - "XX:XX - Section Title"
prerequisites:
  - "Module X required"
related_videos:
  - "M{X}_{Y} - Related Video"
```

---

## Module 1: Introduction to HelixAgent

### Video M01_01: Course Welcome and Learning Path

```yaml
video_id: M01_01
title: "[Module 1] Course Welcome and Learning Path"
description: |
  Welcome to the HelixAgent video course! In this introductory video, you'll learn about the course structure, certification path, and how to get the most out of your learning journey.

  Learning Objectives:
  - Understand the 14-module course structure
  - Learn about the 5-level certification path
  - Know the prerequisites and target audience
  - Navigate course materials effectively

  This course covers HelixAgent, an AI-powered ensemble LLM service that combines responses from multiple language models using intelligent aggregation strategies.

  Prerequisites: None (this is the first video)

  Resources:
  - Course Materials: /docs/courses/
  - HelixAgent Repository: https://github.com/your-org/helix-agent
  - Community Discord: [link]
duration: "08:00"
level: "Beginner"
tags:
  - HelixAgent
  - course introduction
  - AI orchestration
  - LLM
  - multi-provider
  - welcome
timestamps:
  - "00:00 - Welcome and Introduction"
  - "01:30 - Course Structure Overview"
  - "03:00 - Certification Path"
  - "05:00 - Prerequisites Review"
  - "06:30 - How to Get Help"
  - "07:30 - Next Steps"
prerequisites: []
related_videos:
  - "M01_02 - What is HelixAgent?"
```

---

### Video M01_02: What is HelixAgent?

```yaml
video_id: M01_02
title: "[Module 1] What is HelixAgent? - Multi-Provider AI Orchestration"
description: |
  Discover what HelixAgent is and why multi-provider AI orchestration matters. Learn how HelixAgent solves vendor lock-in and improves AI reliability.

  Learning Objectives:
  - Define HelixAgent and its core capabilities
  - Understand multi-provider AI orchestration
  - Know the 10 supported LLM providers
  - Recognize key use cases and benefits

  Covered in this video:
  - The problem with single-provider AI
  - HelixAgent's solution
  - Provider comparison: Claude, DeepSeek, Gemini, Qwen, Ollama, OpenRouter, ZAI, Zen, Mistral, Cerebras
  - Real-world applications

duration: "12:00"
level: "Beginner"
tags:
  - HelixAgent
  - multi-provider
  - AI orchestration
  - LLM
  - Claude
  - DeepSeek
  - Gemini
  - vendor lock-in
timestamps:
  - "00:00 - The Problem with Single-Provider AI"
  - "02:30 - What is HelixAgent?"
  - "05:00 - 10 Supported Providers"
  - "08:00 - Key Benefits"
  - "10:00 - Real-World Use Cases"
  - "11:30 - Summary"
prerequisites:
  - "M01_01 - Course Welcome"
related_videos:
  - "M01_03 - Architecture Overview"
  - "M04_01 - Provider Interface Architecture"
```

---

### Video M01_03: Architecture Overview

```yaml
video_id: M01_03
title: "[Module 1] HelixAgent Architecture Overview"
description: |
  Dive deep into the HelixAgent system architecture. Understand how components work together to provide multi-provider AI orchestration.

  Learning Objectives:
  - Understand the high-level system architecture
  - Know the core components and their responsibilities
  - Learn the LLMProvider interface
  - Trace data flow through the system

  Components Covered:
  - API Gateway
  - Ensemble Engine
  - AI Debate Service
  - Provider Registry
  - Plugin System
  - Cache Layer
  - Monitoring Stack

duration: "15:00"
level: "Beginner"
tags:
  - HelixAgent
  - architecture
  - system design
  - LLMProvider
  - ensemble
  - API gateway
timestamps:
  - "00:00 - High-Level Architecture"
  - "03:00 - Core Components"
  - "06:00 - LLMProvider Interface"
  - "09:00 - Internal Package Structure"
  - "12:00 - Data Flow Walkthrough"
  - "14:00 - Summary"
prerequisites:
  - "M01_02 - What is HelixAgent?"
related_videos:
  - "M01_04 - Use Cases and Applications"
  - "M02_01 - Environment Prerequisites"
```

---

### Video M01_04: Use Cases and Applications

```yaml
video_id: M01_04
title: "[Module 1] HelixAgent Use Cases and Applications"
description: |
  Explore real-world applications of HelixAgent across industries. Learn when and why to use multi-provider AI orchestration.

  Learning Objectives:
  - Identify enterprise use cases
  - Understand industry applications
  - Know performance and security considerations
  - Match use cases to HelixAgent features

  Use Cases Covered:
  - Content Generation
  - Code Analysis
  - Research and Analysis
  - Customer Support
  - Decision Support
  - Translation and Localization

duration: "10:00"
level: "Beginner"
tags:
  - HelixAgent
  - use cases
  - enterprise AI
  - content generation
  - code analysis
timestamps:
  - "00:00 - Enterprise Use Cases"
  - "03:00 - Industry Applications"
  - "06:00 - Performance Characteristics"
  - "08:00 - Security Features"
  - "09:30 - Summary"
prerequisites:
  - "M01_03 - Architecture Overview"
related_videos:
  - "M02_01 - Environment Prerequisites"
```

---

## Module 2: Installation and Setup

### Video M02_01: Environment Prerequisites

```yaml
video_id: M02_01
title: "[Module 2] Environment Prerequisites for HelixAgent"
description: |
  Set up your development environment for HelixAgent. Install all required software and verify your system is ready.

  Learning Objectives:
  - Verify system requirements
  - Install Go 1.24+
  - Set up Docker and Docker Compose
  - Configure Git and IDE

  System Requirements:
  - Linux, macOS, or Windows (WSL2)
  - 8GB RAM minimum (16GB recommended)
  - 20GB disk space
  - Go 1.24+, Docker, Git

duration: "10:00"
level: "Beginner"
tags:
  - HelixAgent
  - installation
  - prerequisites
  - Go
  - Docker
  - development environment
timestamps:
  - "00:00 - System Requirements"
  - "02:00 - Installing Go"
  - "04:30 - Docker Setup"
  - "07:00 - Git Configuration"
  - "08:30 - IDE Recommendations"
  - "09:30 - Verification"
prerequisites:
  - "M01_04 - Use Cases"
related_videos:
  - "M02_02 - Docker Quick Start"
```

---

### Video M02_02: Docker Quick Start

```yaml
video_id: M02_02
title: "[Module 2] HelixAgent Docker Quick Start"
description: |
  Get HelixAgent running quickly with Docker. Learn Docker Compose configuration and start all required services.

  Learning Objectives:
  - Understand Docker-based installation
  - Configure Docker Compose profiles
  - Start core and AI services
  - Verify service health

  Commands Demonstrated:
  - make infra-core
  - docker-compose up -d
  - curl http://localhost:7061/health

duration: "15:00"
level: "Beginner"
tags:
  - HelixAgent
  - Docker
  - Docker Compose
  - installation
  - quick start
timestamps:
  - "00:00 - Docker Installation Advantages"
  - "02:30 - Docker Compose Configuration"
  - "05:00 - Starting Core Services"
  - "08:00 - Starting Full Stack"
  - "11:00 - Verifying Services"
  - "13:00 - Viewing Logs"
  - "14:30 - Summary"
prerequisites:
  - "M02_01 - Environment Prerequisites"
related_videos:
  - "M02_03 - Manual Installation"
```

---

### Video M02_03: Manual Installation from Source

```yaml
video_id: M02_03
title: "[Module 2] Manual Installation from Source"
description: |
  Build HelixAgent from source code. Learn the build process, configuration, and how to run locally without Docker.

  Learning Objectives:
  - Clone and set up the repository
  - Build from source
  - Run in development mode
  - Verify the installation

  Commands Demonstrated:
  - make install-deps
  - make build
  - make run-dev
  - curl http://localhost:7061/health

duration: "15:00"
level: "Beginner"
tags:
  - HelixAgent
  - source build
  - Go
  - make
  - development
timestamps:
  - "00:00 - When to Build from Source"
  - "02:00 - Cloning the Repository"
  - "04:00 - Installing Dependencies"
  - "07:00 - Building the Binary"
  - "10:00 - Running Locally"
  - "13:00 - Verifying APIs"
  - "14:30 - Summary"
prerequisites:
  - "M02_02 - Docker Quick Start"
related_videos:
  - "M02_04 - Podman Alternative"
```

---

### Video M02_04: Podman Alternative Setup

```yaml
video_id: M02_04
title: "[Module 2] Podman Alternative Setup for HelixAgent"
description: |
  Use Podman instead of Docker for running HelixAgent. Learn the differences and Podman-specific configuration.

  Learning Objectives:
  - Understand Podman vs Docker
  - Configure Podman for HelixAgent
  - Use the container-runtime script
  - Run services with Podman

duration: "08:00"
level: "Beginner"
tags:
  - HelixAgent
  - Podman
  - containers
  - rootless
timestamps:
  - "00:00 - Podman vs Docker"
  - "02:00 - When to Use Podman"
  - "04:00 - Configuration Differences"
  - "06:00 - Running with Podman"
  - "07:30 - Summary"
prerequisites:
  - "M02_03 - Manual Installation"
related_videos:
  - "M02_05 - Troubleshooting"
```

---

### Video M02_05: Troubleshooting Installation Issues

```yaml
video_id: M02_05
title: "[Module 2] Troubleshooting Installation Issues"
description: |
  Solve common installation problems with HelixAgent. Learn diagnostic commands and fixes for typical issues.

  Learning Objectives:
  - Diagnose common installation issues
  - Fix port conflicts
  - Resolve Docker problems
  - Debug API key issues

  Issues Covered:
  - Port 7061 already in use
  - Docker daemon not running
  - Go version mismatches
  - Invalid API keys

duration: "12:00"
level: "Beginner"
tags:
  - HelixAgent
  - troubleshooting
  - debugging
  - installation
  - common issues
timestamps:
  - "00:00 - Common Installation Problems"
  - "02:30 - Port Conflicts"
  - "04:30 - Docker Issues"
  - "06:30 - Go Version Problems"
  - "08:30 - API Key Troubleshooting"
  - "10:30 - Log Analysis"
  - "11:30 - Summary"
prerequisites:
  - "M02_04 - Podman Alternative"
related_videos:
  - "M03_01 - Configuration Architecture"
```

---

## Module 3: Configuration

### Video M03_01: Configuration Architecture

```yaml
video_id: M03_01
title: "[Module 3] HelixAgent Configuration Architecture"
description: |
  Understand how HelixAgent configuration works. Learn the hierarchy of config files, environment variables, and defaults.

  Learning Objectives:
  - Understand configuration file hierarchy
  - Know environment variable precedence
  - Learn configuration loading order
  - Apply secrets management best practices

duration: "12:00"
level: "Beginner"
tags:
  - HelixAgent
  - configuration
  - environment variables
  - YAML
  - secrets
timestamps:
  - "00:00 - Configuration Overview"
  - "03:00 - File Hierarchy"
  - "06:00 - Environment Variables"
  - "09:00 - Loading Order"
  - "11:00 - Summary"
prerequisites:
  - "M02_05 - Troubleshooting"
related_videos:
  - "M03_02 - Core Configuration"
```

---

### Video M03_02: Core Configuration Options

```yaml
video_id: M03_02
title: "[Module 3] Core Configuration Options"
description: |
  Configure HelixAgent's core settings including server, database, cache, and logging options.

  Learning Objectives:
  - Configure server settings (PORT, GIN_MODE)
  - Set up database connections
  - Configure Redis cache
  - Manage logging options

duration: "15:00"
level: "Beginner"
tags:
  - HelixAgent
  - configuration
  - server
  - database
  - Redis
  - logging
timestamps:
  - "00:00 - Server Configuration"
  - "04:00 - Database Settings"
  - "08:00 - Redis Configuration"
  - "11:00 - Logging Options"
  - "13:00 - Configuration Files"
  - "14:30 - Summary"
prerequisites:
  - "M03_01 - Configuration Architecture"
related_videos:
  - "M03_03 - Provider Configuration"
```

---

### Video M03_03: Provider Configuration

```yaml
video_id: M03_03
title: "[Module 3] LLM Provider Configuration"
description: |
  Configure API keys and settings for all 10 LLM providers. Learn about OAuth credentials and rate limiting.

  Learning Objectives:
  - Configure API keys for each provider
  - Set up OAuth credentials
  - Configure Ollama for local models
  - Manage rate limiting and timeouts

  Providers Covered:
  - Claude, DeepSeek, Gemini, Qwen
  - ZAI, OpenRouter, Mistral, Cerebras
  - Ollama, Zen

duration: "12:00"
level: "Beginner"
tags:
  - HelixAgent
  - providers
  - API keys
  - OAuth
  - configuration
timestamps:
  - "00:00 - API Key Configuration"
  - "04:00 - OAuth Credentials"
  - "07:00 - Ollama Setup"
  - "09:00 - Rate Limiting"
  - "11:00 - Summary"
prerequisites:
  - "M03_02 - Core Configuration"
related_videos:
  - "M03_04 - Advanced Configuration"
  - "M04_01 - Provider Interface"
```

---

### Video M03_04: Advanced Configuration

```yaml
video_id: M03_04
title: "[Module 3] Advanced Configuration Options"
description: |
  Configure advanced HelixAgent features including AI Debate, Cognee integration, BigData, and service overrides.

  Learning Objectives:
  - Configure AI Debate settings
  - Set up Cognee integration
  - Enable BigData components
  - Use service overrides

duration: "12:00"
level: "Intermediate"
tags:
  - HelixAgent
  - advanced configuration
  - AI Debate
  - Cognee
  - BigData
timestamps:
  - "00:00 - AI Debate Configuration"
  - "03:30 - Cognee Integration"
  - "06:30 - BigData Components"
  - "09:00 - Service Overrides"
  - "11:00 - Summary"
prerequisites:
  - "M03_03 - Provider Configuration"
related_videos:
  - "M03_05 - Configuration Best Practices"
```

---

### Video M03_05: Configuration Best Practices

```yaml
video_id: M03_05
title: "[Module 3] Configuration Best Practices"
description: |
  Learn best practices for managing HelixAgent configuration across environments with proper secrets handling.

  Learning Objectives:
  - Create environment-specific configs
  - Manage secrets securely
  - Validate configuration
  - Document configuration

duration: "09:00"
level: "Beginner"
tags:
  - HelixAgent
  - best practices
  - configuration
  - secrets
  - security
timestamps:
  - "00:00 - Environment-Specific Configs"
  - "03:00 - Secrets Management"
  - "05:30 - Configuration Validation"
  - "07:30 - Documentation"
  - "08:30 - Summary"
prerequisites:
  - "M03_04 - Advanced Configuration"
related_videos:
  - "M04_01 - Provider Interface"
```

---

## Module 4-14: Metadata Templates

*For brevity, the remaining modules follow the same template structure. Below is the metadata summary for each video:*

---

### Module 4: LLM Provider Integration (75 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M04_01 | Provider Interface Architecture | 12:00 |
| M04_02 | Claude Integration | 10:00 |
| M04_03 | DeepSeek Integration | 10:00 |
| M04_04 | Gemini Integration | 10:00 |
| M04_05 | Qwen, Ollama, and Other Providers | 18:00 |
| M04_06 | Building Fallback Chains | 15:00 |

**Common Tags**: HelixAgent, LLM providers, API integration, Claude, DeepSeek, Gemini, Qwen, Ollama, fallback

---

### Module 5: Ensemble Strategies (60 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M05_01 | Introduction to Ensemble AI | 12:00 |
| M05_02 | Voting Strategies Explained | 15:00 |
| M05_03 | Implementing Custom Strategies | 15:00 |
| M05_04 | Performance Optimization | 10:00 |
| M05_05 | Ensemble Benchmarking | 08:00 |

**Common Tags**: HelixAgent, ensemble, voting, consensus, multi-model, AI orchestration

---

### Module 6: AI Debate System (90 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M06_01 | AI Debate Concepts | 15:00 |
| M06_02 | Participant Configuration | 18:00 |
| M06_03 | Debate Strategies Deep Dive | 15:00 |
| M06_04 | Cognee AI Integration | 12:00 |
| M06_05 | Programmatic Debate Execution | 15:00 |
| M06_06 | Monitoring and Metrics | 15:00 |

**Common Tags**: HelixAgent, AI Debate, multi-agent, consensus, Cognee, reasoning

---

### Module 7: Plugin Development (75 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M07_01 | Plugin Architecture Overview | 12:00 |
| M07_02 | Plugin Interfaces Deep Dive | 15:00 |
| M07_03 | Developing Your First Plugin | 20:00 |
| M07_04 | Advanced Plugin Topics | 15:00 |
| M07_05 | Plugin Deployment and Testing | 13:00 |

**Common Tags**: HelixAgent, plugins, extensibility, hot-reload, development, Go

---

### Module 8: MCP/LSP Integration (60 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M08_01 | Protocol Support Overview | 10:00 |
| M08_02 | MCP Integration Deep Dive | 15:00 |
| M08_03 | LSP Integration | 12:00 |
| M08_04 | ACP and Embeddings | 13:00 |
| M08_05 | Building Protocol Workflows | 10:00 |

**Common Tags**: HelixAgent, MCP, LSP, ACP, protocols, embeddings, tools

---

### Module 9: Optimization Features (75 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M09_01 | Optimization Framework Overview | 12:00 |
| M09_02 | Semantic Caching with GPTCache | 15:00 |
| M09_03 | Structured Output with Outlines | 12:00 |
| M09_04 | Enhanced Streaming | 12:00 |
| M09_05 | Advanced Optimization (SGLang, LlamaIndex) | 15:00 |
| M09_06 | Measuring Optimization Impact | 09:00 |

**Common Tags**: HelixAgent, optimization, caching, GPTCache, streaming, performance

---

### Module 10: Security Best Practices (60 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M10_01 | Security Architecture | 12:00 |
| M10_02 | API Security Configuration | 12:00 |
| M10_03 | Secrets Management | 12:00 |
| M10_04 | Production Security Hardening | 15:00 |
| M10_05 | Security Testing | 09:00 |

**Common Tags**: HelixAgent, security, authentication, JWT, secrets, hardening

---

### Module 11: Testing and CI/CD (75 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M11_01 | Testing Strategy Overview | 12:00 |
| M11_02 | Running All Test Types | 18:00 |
| M11_03 | Writing Effective Tests | 18:00 |
| M11_04 | CI/CD Pipeline Setup | 18:00 |
| M11_05 | Quality Gates and Automation | 09:00 |

**Common Tags**: HelixAgent, testing, CI/CD, GitHub Actions, quality, automation

---

### Module 12: Challenge System and Validation (90 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M12_01 | Challenge System Architecture | 15:00 |
| M12_02 | RAGS Challenge - RAG Integration | 18:00 |
| M12_03 | MCPS Challenge - MCP Server Integration | 18:00 |
| M12_04 | SKILLS Challenge - Skills Integration | 12:00 |
| M12_05 | Strict Real-Result Validation | 15:00 |
| M12_06 | Debugging Challenge Failures | 12:00 |

**Common Tags**: HelixAgent, challenges, RAGS, MCPS, SKILLS, validation, testing

---

### Module 13: MCP Tool Search and Discovery (60 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M13_01 | MCP Tool Search Overview | 12:00 |
| M13_02 | Tool Search Implementation | 15:00 |
| M13_03 | AI-Powered Tool Suggestions | 12:00 |
| M13_04 | Adapter Search and Discovery | 12:00 |
| M13_05 | Building Discovery Workflows | 09:00 |

**Common Tags**: HelixAgent, MCP, tool search, discovery, AI suggestions, adapters

---

### Module 14: AI Debate System Advanced (90 min)

| Video ID | Title | Duration |
|----------|-------|----------|
| M14_01 | 15-LLM Debate Team Configuration | 18:00 |
| M14_02 | Multi-Pass Validation System | 18:00 |
| M14_03 | Debate Orchestrator Framework | 18:00 |
| M14_04 | LLMsVerifier Integration | 15:00 |
| M14_05 | CLI Agent Integration | 12:00 |
| M14_06 | Production Debate Deployment | 09:00 |

**Common Tags**: HelixAgent, AI Debate, multi-pass validation, LLMsVerifier, CLI agents, production

---

## Thumbnail Guidelines

### Template Structure

```
+----------------------------------------+
|                                        |
|   [MODULE NUMBER]                      |
|                                        |
|   VIDEO TITLE                          |
|   (Large, Bold Text)                   |
|                                        |
|   [Visual Icon/Screenshot]             |
|                                        |
|   HelixAgent Logo                      |
+----------------------------------------+
```

### Specifications

- **Size**: 1280x720 pixels
- **Format**: PNG or JPEG
- **Text**: Max 6 words visible
- **Colors**: Use brand palette
- **Font**: Inter Bold for titles

### Color Coding by Level

| Level | Color | Hex |
|-------|-------|-----|
| Beginner | Green | #10B981 |
| Intermediate | Blue | #2563EB |
| Advanced | Purple | #7C3AED |

---

## SEO Recommendations

### Title Optimization

- Include "HelixAgent" in all titles
- Keep titles under 60 characters
- Use action words: "Master", "Build", "Configure"
- Include video number for series

### Description Best Practices

- First 150 characters are most important
- Include primary keywords
- List learning objectives
- Add timestamps
- Include links to resources

### Tag Strategy

**Always Include**:
- HelixAgent
- AI orchestration
- LLM
- multi-provider

**Module-Specific Tags**:
- Add 10-15 relevant tags per video
- Include technology names
- Add problem/solution keywords

---

## Playlist Organization

### Main Playlists

1. **Complete Course** - All 74 videos in order
2. **Module Playlists** - One per module (14 total)
3. **Skill Level Playlists**:
   - Beginner (Modules 1-3)
   - Intermediate (Modules 4-9)
   - Advanced (Modules 10-14)

### Quick Start Playlist

- M01_01 - Course Welcome
- M02_02 - Docker Quick Start
- M03_03 - Provider Configuration
- M04_06 - Building Fallback Chains
- M06_05 - Programmatic Debate Execution

---

## Publishing Checklist

Before publishing each video:

- [ ] Title follows naming convention
- [ ] Description complete with timestamps
- [ ] All tags added (15-20 per video)
- [ ] Thumbnail uploaded
- [ ] Captions added/verified
- [ ] End screens configured
- [ ] Cards added at relevant points
- [ ] Added to appropriate playlists
- [ ] Linked to related videos
- [ ] Preview tested on mobile

---

---

## Module S7.1: Advanced AI/ML Modules — Part 1

### Video MS71_01: Agentic Module — Graph-Based Workflow Orchestration

```yaml
video_id: MS71_01
title: "[Module S7.1] Agentic Module — Graph-Based Workflow Orchestration"
description: |
  Master the Agentic module (digital.vasic.agentic) for building autonomous AI workflows using
  directed graphs. Learn how to define nodes, edges, and mutable WorkflowState to create agents
  that can plan, execute, and self-correct multi-step tasks.

  Learning Objectives:
  - Understand graph-based workflow architecture (Workflow, WorkflowGraph, Node, Edge)
  - Implement NodeHandler functions for each step
  - Use WorkflowState to thread mutable context through all nodes
  - Enable dynamic routing via NextNode output
  - Integrate with HelixAgent via the agentic adapter

  Topics Covered:
  - Workflow and WorkflowGraph types
  - NodeHandler signature and NodeOutput routing
  - WorkflowState get/set API
  - Building a three-node code-review agent (live demo)
  - HelixAgent adapter usage at internal/adapters/agentic/

  Prerequisites: Modules 1-3 (HelixAgent basics), familiarity with Go interfaces

  Resources:
  - Module source: Agentic/
  - Adapter: internal/adapters/agentic/adapter.go
  - Challenge: challenges/scripts/agentic_challenge.sh

duration: "30:00"
level: "Advanced"
tags:
  - HelixAgent
  - agentic
  - workflow orchestration
  - autonomous agents
  - graph-based
  - Go
  - digital.vasic.agentic
  - NodeHandler
  - WorkflowState
  - planning
  - self-correction
timestamps:
  - "00:00 - Introduction: What the Agentic Module Solves"
  - "05:00 - Core Concepts: Workflow, Node, Edge, WorkflowState"
  - "10:00 - NodeHandler Signature and NodeOutput Routing"
  - "15:00 - Live Demo: Three-Node Code-Review Agent"
  - "25:00 - Integration Patterns with HelixAgent Adapters"
  - "29:00 - Summary and Next Steps"
prerequisites:
  - "Modules 1-3 completed"
  - "Go 1.24+ installed"
  - "HelixAgent running locally"
related_videos:
  - "MS71_02 - LLMOps Module"
  - "MS72_01 - Planning Module"
  - "M07_01 - Plugin Architecture"
```

---

### Video MS71_02: LLMOps Module — Evaluation, Experiments, and Prompt Versioning

```yaml
video_id: MS71_02
title: "[Module S7.1] LLMOps Module — Continuous Evaluation and A/B Experiments"
description: |
  Learn how to operate LLMs in production with the LLMOps module (digital.vasic.llmops).
  Implement continuous evaluation pipelines, A/B experiment management, and prompt versioning
  to detect regressions and improve model quality over time.

  Learning Objectives:
  - Run continuous evaluation against golden, synthetic, and production datasets
  - Create and manage A/B experiments between two model configurations
  - Version prompt templates and measure their impact on quality
  - Interpret EvaluationRun metrics (MeanScore, per-example results)
  - Integrate with HelixAgent's LLMsVerifier pipeline

  Topics Covered:
  - InMemoryContinuousEvaluator and Dataset types
  - InMemoryExperimentManager: creating experiments, recording results, finding the winner
  - EvaluationRun structure and AggregateMetrics
  - Prompt versioning and rollback patterns
  - Live demo: A/B experiment between DeepSeek and Claude for code generation
  - HelixAgent adapter at internal/adapters/llmops/

  Prerequisites: Module S7.1.1 (Agentic module), LLMsVerifier basics (Module 14)

  Resources:
  - Module source: LLMOps/
  - Adapter: internal/adapters/llmops/adapter.go
  - Challenge: challenges/scripts/llmops_challenge.sh

duration: "30:00"
level: "Advanced"
tags:
  - HelixAgent
  - LLMOps
  - evaluation
  - A/B testing
  - experiment management
  - prompt versioning
  - digital.vasic.llmops
  - InMemoryContinuousEvaluator
  - dataset management
  - model quality
  - regression detection
timestamps:
  - "00:00 - Introduction: Why LLMOps Matters"
  - "05:00 - Core Concepts: Evaluator, Experiment, Dataset"
  - "10:00 - EvaluationRun and Aggregate Metrics"
  - "15:00 - Live Demo: A/B Experiment Between Providers"
  - "25:00 - Integration Patterns: LLMsVerifier and Provider Promotions"
  - "29:00 - Summary"
prerequisites:
  - "MS71_01 - Agentic Module"
  - "Module 14 - LLMsVerifier Integration"
related_videos:
  - "MS71_01 - Agentic Module"
  - "MS71_03 - SelfImprove Module"
  - "M14_04 - LLMsVerifier Integration"
```

---

### Video MS71_03: SelfImprove Module — RLHF, Reward Modeling, and Preference Optimization

```yaml
video_id: MS71_03
title: "[Module S7.1] SelfImprove Module — RLHF and Self-Refinement Loops"
description: |
  Implement AI self-improvement infrastructure with the SelfImprove module
  (digital.vasic.selfimprove). Learn how to collect explicit and implicit feedback, train reward
  models, run Direct Preference Optimization, and build self-refinement loops that make every
  response better than the last.

  Learning Objectives:
  - Collect explicit (thumbs up/down) and implicit (behavioral) feedback
  - Build and train a RewardModel from PreferencePair data
  - Implement a SelfRefinementLoop for iterative response improvement
  - Integrate feedback collection into HelixAgent streaming responses
  - Apply preference optimization at inference time

  Topics Covered:
  - ExplicitFeedback, ImplicitFeedback, and PreferencePair types
  - RewardModel interface: Score, Train, Evaluate
  - FeedbackCollector buffering and batch export
  - SelfRefinementLoop: initial → critique → refine → score → done
  - Live demo: Self-refinement improving a code explanation through 3 iterations
  - HelixAgent adapter at internal/adapters/selfimprove/
  - PII considerations for feedback data

  Prerequisites: Module S7.1.2 (LLMOps), Module 12 (Challenge System)

  Resources:
  - Module source: SelfImprove/
  - Adapter: internal/adapters/selfimprove/adapter.go
  - Challenge: challenges/scripts/selfimprove_challenge.sh

duration: "30:00"
level: "Advanced"
tags:
  - HelixAgent
  - RLHF
  - reward modeling
  - preference optimization
  - self-improvement
  - digital.vasic.selfimprove
  - FeedbackCollector
  - SelfRefinementLoop
  - DPO
  - AI quality
  - continuous improvement
timestamps:
  - "00:00 - Introduction: RLHF in Production Systems"
  - "05:00 - Core Concepts: Feedback Types and RewardModel"
  - "10:00 - SelfRefinementLoop Architecture"
  - "15:00 - Live Demo: Three-Iteration Self-Refinement"
  - "25:00 - Integration Patterns: Feedback in Streaming Responses"
  - "29:00 - Summary"
prerequisites:
  - "MS71_02 - LLMOps Module"
  - "Module 12 - Challenge System"
related_videos:
  - "MS71_02 - LLMOps Module"
  - "MS72_01 - Planning Module"
  - "M14_02 - Multi-Pass Validation"
```

---

## Module S7.2: Advanced AI/ML Modules — Part 2

### Video MS72_01: Planning Module — HiPlan, MCTS, and Tree of Thoughts

```yaml
video_id: MS72_01
title: "[Module S7.2] Planning Module — HiPlan, MCTS, and Tree of Thoughts"
description: |
  Give your AI agents the ability to plan ahead using the Planning module
  (digital.vasic.planning). Master three complementary planning algorithms: HiPlan for
  hierarchical decomposition, Monte Carlo Tree Search (MCTS) for exploratory planning, and
  Tree of Thoughts for open-ended reasoning.

  Learning Objectives:
  - Choose the right planning algorithm for each task type
  - Implement HiPlan with custom MilestoneGenerator and StepExecutor
  - Configure MCTS with UCB exploration, action generators, and reward functions
  - Build Tree of Thoughts with LLM-backed ThoughtGenerator and ThoughtEvaluator
  - Integrate planning algorithms into HelixAgent's SpecKit orchestrator

  Topics Covered:
  - HiPlan: HierarchicalPlan, Milestone, PlanStep, LLMMilestoneGenerator
  - MCTS: MCTSConfig, MCTSNode, CodeActionGenerator, CodeRewardFunction
  - Tree of Thoughts: TreeOfThoughtsConfig, Thought, ThoughtNode, LLMThoughtEvaluator
  - Algorithm selection guide: when to use HiPlan vs MCTS vs ToT
  - Live demo: HiPlan decomposing a software feature; MCTS optimizing code
  - Cost control: MaxDepth, MaxIterations, BranchingFactor tradeoffs
  - HelixAgent adapter at internal/adapters/planning/

  Prerequisites: Module S7.1.1 (Agentic), Module 9 (Optimization)

  Resources:
  - Module source: Planning/
  - Adapter: internal/adapters/planning/adapter.go
  - Challenge: challenges/scripts/planning_challenge.sh

duration: "30:00"
level: "Advanced"
tags:
  - HelixAgent
  - planning
  - HiPlan
  - MCTS
  - Monte Carlo Tree Search
  - Tree of Thoughts
  - digital.vasic.planning
  - hierarchical planning
  - autonomous agents
  - AI reasoning
  - lookahead
timestamps:
  - "00:00 - Introduction: Why Agents Need Planning"
  - "05:00 - Core Concepts: Three Planning Algorithms Compared"
  - "10:00 - HiPlan, MCTS, and ToT Data Types"
  - "15:00 - Live Demo: HiPlan for Software Features; MCTS for Code"
  - "25:00 - Integration Patterns: SpecKit and Debate System"
  - "29:00 - Summary"
prerequisites:
  - "MS71_01 - Agentic Module"
  - "Module 9 - Optimization Features"
related_videos:
  - "MS71_01 - Agentic Module"
  - "MS72_02 - Benchmark Module"
  - "M14_03 - Debate Orchestrator Framework"
```

---

### Video MS72_02: Benchmark Module — Standardized LLM Evaluation

```yaml
video_id: MS72_02
title: "[Module S7.2] Benchmark Module — Standardized LLM Provider Evaluation"
description: |
  Objectively compare LLM providers using the Benchmark module (digital.vasic.benchmark).
  Run standardized benchmarks (MMLU, HumanEval, GSM8K, SWE-Bench, and more) as well as custom
  domain benchmarks to make data-driven provider selection decisions.

  Learning Objectives:
  - Run MMLU, HumanEval, and GSM8K benchmarks against multiple providers
  - Create custom benchmark datasets for domain-specific evaluation
  - Interpret BenchmarkResult metrics: score, sub-scores, latency, cost
  - Generate ComparisonReports with statistical significance testing
  - Integrate benchmark results into HelixAgent's provider scoring and LLMsVerifier pipeline

  Topics Covered:
  - Supported benchmarks: MMLU, HumanEval, GSM8K, SWE-Bench, MBPP, LMSYS, HellaSwag, MATH
  - BenchmarkRunner, RunConfig, BenchmarkResult, ComparisonReport types
  - Custom benchmark workflow: define dataset, register, run, compare
  - Statistical significance: bootstrapped confidence intervals and p-values
  - Live demo: HumanEval comparison between DeepSeek-Coder and Claude-3.5-Sonnet
  - Resource management: GOMAXPROCS limits, MaxExamples for CI vs nightly runs
  - Storing results in PostgreSQL for trend analysis
  - HelixAgent adapter at internal/adapters/benchmark/

  Prerequisites: Module S7.1.2 (LLMOps), Module 14 (LLMsVerifier Integration)

  Resources:
  - Module source: Benchmark/
  - Adapter: internal/adapters/benchmark/adapter.go
  - Challenge: challenges/scripts/benchmark_challenge.sh

duration: "30:00"
level: "Advanced"
tags:
  - HelixAgent
  - benchmark
  - MMLU
  - HumanEval
  - GSM8K
  - digital.vasic.benchmark
  - LLM evaluation
  - provider comparison
  - BenchmarkRunner
  - ComparisonReport
  - A/B testing
  - model quality
timestamps:
  - "00:00 - Introduction: Why Standardized Benchmarks Matter"
  - "05:00 - Core Concepts: Supported Benchmarks and Key Types"
  - "10:00 - BenchmarkRunner, RunConfig, BenchmarkResult"
  - "15:00 - Live Demo: HumanEval on DeepSeek vs Claude"
  - "25:00 - Integration Patterns: LLMsVerifier and Provider Promotions"
  - "29:00 - Summary"
prerequisites:
  - "MS71_02 - LLMOps Module"
  - "Module 14 - LLMsVerifier Integration"
related_videos:
  - "MS71_02 - LLMOps Module"
  - "MS71_03 - SelfImprove Module"
  - "M14_04 - LLMsVerifier Integration"
```

---

### Module S7.1-S7.2: Summary Table

| Video ID | Title | Duration | Level |
|----------|-------|----------|-------|
| MS71_01 | Agentic — Graph-Based Workflow Orchestration | 30:00 | Advanced |
| MS71_02 | LLMOps — Evaluation, Experiments, Prompt Versioning | 30:00 | Advanced |
| MS71_03 | SelfImprove — RLHF, Reward Modeling, Preference Optimization | 30:00 | Advanced |
| MS72_01 | Planning — HiPlan, MCTS, Tree of Thoughts | 30:00 | Advanced |
| MS72_02 | Benchmark — Standardized LLM Evaluation | 30:00 | Advanced |

**Common Tags**: HelixAgent, advanced AI/ML, Go modules, extracted modules, digital.vasic,
autonomous agents, LLM operations, self-improvement, planning algorithms, benchmarking

---

*Metadata Templates Version: 1.1.0*
*Last Updated: February 2026*
*Total Videos: 79 (74 original + 5 new AI/ML module videos)*
