# HelixAgent Website

Documentation and content for the HelixAgent project website.

## Structure

- `user-manuals/` — Step-by-step user guides (43 manuals)
- `video-courses/` — Video course content and scripts (65+ courses)
- `scripts/` — Build and validation scripts
- `styles/` — Website stylesheet assets
- `public/` — Static public assets
- `build.sh` — Website build script

## User Manuals

Comprehensive guides covering:
- Getting started and API reference
- Provider configuration and deployment
- AI debate system and protocols
- Security, performance, and plugin development
- BigData, gRPC, memory, and code formatters
- Automated security scanning and performance monitoring
- Concurrency patterns and testing strategies
- Challenge development and custom provider guides
- Observability, backup, and disaster recovery
- Enterprise architecture and compliance
- Agentic workflows, LLMOps, planning algorithms
- Module-specific guides (DocProcessor, HelixQA, LLMOrchestrator, VisionEngine, etc.)

Current manuals: 01 through 43 (`user-manuals/01-getting-started.md` through
`user-manuals/43-agentic-ensemble-guide.md`).

## Video Courses

Structured learning paths covering:
- Fundamentals, AI debate, deployment, and protocols
- Advanced providers, plugin development, and MCP mastery
- Security scanning, performance tuning, and stress testing
- Memory management, cloud providers, and enterprise deployment
- HelixMemory, HelixSpecifier, and module deep dives
- Goroutine safety, router completeness, and lazy loading patterns
- Agentic workflows, LLMOps, planning algorithms, and more

Course files use two naming conventions:
- `course-NN-<topic>.md` — Primary course series (01-18, 66-76)
- `video-course-NN-<topic>.md` — Extended video series (53-65)
- `courses-NN-MM/` — Batch course subdirectories

## Contributing

- Follow standard Markdown formatting
- Preserve ALL existing content when updating files
- Add new content at the end of existing files, never remove
- Number new manuals sequentially (next: 44 — `44-<topic>.md`)
- Number new video courses sequentially (next: 77)
- Keep user manuals practical with real examples and curl commands
- Cross-reference related manuals and courses where relevant
