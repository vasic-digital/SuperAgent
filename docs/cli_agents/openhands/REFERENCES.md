# OpenHands - External References

## Official Documentation

### Primary Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **Official Docs** | https://docs.openhands.dev | Complete documentation |
| **GitHub Repository** | https://github.com/OpenHands/OpenHands | Source code |
| **SDK Documentation** | https://docs.openhands.dev/sdk | Software Agent SDK |
| **CLI Documentation** | https://docs.openhands.dev/openhands/usage/run-openhands/cli-mode | CLI usage guide |
| **Local Setup** | https://docs.openhands.dev/openhands/usage/run-openhands/local-setup | Local installation |

### API Documentation

| Resource | URL | Description |
|----------|-----|-------------|
| **LiteLLM** | https://docs.litellm.ai | LLM provider abstraction |
| **FastAPI** | https://fastapi.tiangolo.com | Web framework |
| **Anthropic API** | https://docs.anthropic.com | Claude API reference |
| **OpenAI API** | https://platform.openai.com/docs | GPT API reference |

---

## Community Resources

### Forums & Discussion

| Platform | Link | Description |
|----------|------|-------------|
| **Slack** | https://dub.sh/openhands | Official community |
| **GitHub Issues** | https://github.com/OpenHands/OpenHands/issues | Bug reports |
| **GitHub Discussions** | https://github.com/OpenHands/OpenHands/discussions | Community discussions |
| **Discord** | Check Slack for invite | Alternative community |

### Social Media

| Platform | Handle | Description |
|----------|--------|-------------|
| **Twitter/X** | @AllHandsAI | Official account |
| **LinkedIn** | All-Hands AI | Company updates |
| **YouTube** | Search "OpenHands AI" | Video tutorials |

---

## Related Tools & Projects

### Similar AI Coding Agents

| Tool | Description | Website |
|------|-------------|---------|
| **Claude Code** | Anthropic's agentic coding tool | https://code.claude.com |
| **GitHub Copilot** | AI pair programmer | https://github.com/features/copilot |
| **Cursor** | AI-first code editor | https://cursor.com |
| **Aider** | AI pair programming | https://aider.chat |
| **Devin** | Autonomous AI engineer | https://devin.ai |
| **SWE-agent** | Stanford's agent for software engineering | https://swe-agent.com |

### Complementary Tools

| Tool | Purpose | Integration |
|------|---------|-------------|
| **Docker** | Container runtime | Primary runtime |
| **Kubernetes** | Container orchestration | Supported runtime |
| **Poetry** | Python dependency management | Required for dev |
| **Node.js** | Frontend runtime | Required for UI |
| **LiteLLM Proxy** | LLM routing | Compatible |

---

## MCP Server Registry

### Official MCP Servers

| Server | Description | Install |
|--------|-------------|---------|
| **Filesystem** | File operations | `npx @modelcontextprotocol/server-filesystem` |
| **GitHub** | GitHub integration | `npx @github/mcp-server` |
| **PostgreSQL** | Database access | `npx @modelcontextprotocol/server-postgres` |
| **SQLite** | SQLite integration | `npx @modelcontextprotocol/server-sqlite` |
| **Brave Search** | Web search | `npx @modelcontextprotocol/server-brave-search` |
| **Fetch** | Web content fetching | `uvx mcp-server-fetch` |

### Community MCP Servers

| Server | Author | Description |
|--------|--------|-------------|
| **Kubernetes** | Community | K8s cluster management |
| **AWS** | Community | AWS operations |
| **Slack** | Community | Slack integration |
| **Notion** | Community | Notion workspace access |

**Full Registry:** https://modelcontextprotocol.io/servers

---

## Technical References

### Protocols & Standards

| Standard | Description | Link |
|----------|-------------|------|
| **Model Context Protocol** | Tool integration standard | https://modelcontextprotocol.io/ |
| **JSON-RPC 2.0** | MCP transport | https://www.jsonrpc.org/specification |
| **Server-Sent Events** | Streaming transport | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events |
| **WebSocket** | Real-time communication | https://developer.mozilla.org/en-US/docs/Web/API/WebSocket |

### Research Papers

| Paper | Authors | Link | Relevance |
|-------|---------|------|-----------|
| **CodeAct** | Xingyao Wang et al. | https://arxiv.org/abs/2402.01030 | Core OpenHands framework |
| **SWE-bench** | Jimenez et al. | https://arxiv.org/abs/2310.06770 | Evaluation benchmark |
| **SWE-agent** | Princeton NLP | https://arxiv.org/abs/2405.15793 | Similar agent research |

---

## Benchmarks

### SWE-bench Results

OpenHands achieves competitive results on SWE-bench:

| Benchmark | OpenHands Score | Leaderboard |
|-----------|-----------------|-------------|
| **SWE-bench Verified** | 77.6% | https://www.swebench.com/ |
| **SWE-bench Lite** | 64.0% | https://www.swebench.com/ |

### Evaluation Infrastructure

| Resource | URL | Description |
|----------|-----|-------------|
| **OpenHands Benchmarks** | https://github.com/OpenHands/benchmarks | Evaluation framework |
| **SWE-bench** | https://www.swebench.com/ | Software engineering benchmark |

---

## Docker Resources

### Official Images

| Image | Purpose | Pull |
|-------|---------|------|
| `openhands:latest` | Main application | `docker pull openhands:latest` |
| `ghcr.io/openhands/agent-server` | Agent server | `docker pull ghcr.io/openhands/agent-server` |

### Base Runtime Images

| Image | Description |
|-------|-------------|
| `nikolaik/python-nodejs:python3.12-nodejs22` | Default base image |
| `openhands/runtime:*` | Pre-built runtime images |

---

## Development Resources

### Source Code Structure

| Directory | Contents |
|-----------|----------|
| `openhands/agenthub/` | Agent implementations |
| `openhands/core/` | Core configuration and loop |
| `openhands/events/` | Event system |
| `openhands/runtime/` | Runtime implementations |
| `openhands/server/` | WebSocket API server |
| `frontend/` | React frontend |
| `enterprise/` | Enterprise features |

### Contributing

| Resource | Link |
|----------|------|
| **Contributing Guide** | https://github.com/OpenHands/OpenHands/blob/main/CONTRIBUTING.md |
| **Code of Conduct** | https://github.com/OpenHands/OpenHands/blob/main/CODE_OF_CONDUCT.md |
| **Development Guide** | https://github.com/OpenHands/OpenHands/blob/main/Development.md |

---

## Tutorials & Guides

### Official Tutorials

| Resource | Description | Link |
|----------|-------------|------|
| **Getting Started** | First steps with OpenHands | https://docs.openhands.dev/openhands/getting-started |
| **Architecture** | System architecture docs | https://docs.openhands.dev/usage/architecture/backend |
| **LLM Configuration** | Setting up different LLMs | https://docs.openhands.dev/usage/llms |

### Community Tutorials

| Resource | Author | Description |
|----------|--------|-------------|
| **YouTube Tutorials** | Various | Search "OpenHands AI tutorial" |
| **Blog Posts** | All-Hands | https://www.all-hands.dev/blog |

---

## Enterprise Resources

### OpenHands Enterprise

| Resource | Link |
|----------|------|
| **Enterprise Page** | https://openhands.dev/enterprise |
| **Pricing** | Contact for details |
| **Features** | SSO, RBAC, integrations |

### Integrations

| Integration | Type |
|-------------|------|
| **GitHub** | Git provider |
| **GitLab** | Git provider |
| **Jira** | Issue tracking |
| **Linear** | Issue tracking |
| **Slack** | Communication |

---

## Version History

### Recent Releases

| Version | Date | Key Features |
|---------|------|--------------|
| v1.4.0 | Recent | Latest stable release |
| v1.3.x | 2025 | V1 architecture, SDK |
| v1.0.x | 2024 | Initial stable release |

**Full Changelog:** https://github.com/OpenHands/OpenHands/releases

---

## Support Channels

| Channel | Purpose | Response Time |
|---------|---------|---------------|
| **GitHub Issues** | Bugs, features | Community-driven |
| **Slack** | Community help | Real-time |
| **Email** | Enterprise support | Business hours |

---

## License & Legal

| Document | Link |
|----------|------|
| **MIT License** | https://github.com/OpenHands/OpenHands/blob/main/LICENSE |
| **Enterprise License** | https://github.com/OpenHands/OpenHands/blob/main/enterprise/LICENSE |
| **Security Policy** | https://github.com/OpenHands/OpenHands/blob/main/SECURITY.md |

---

## Quick Links

### For New Users
- [Getting Started](https://docs.openhands.dev/openhands/getting-started)
- [Installation Guide](https://docs.openhands.dev/openhands/usage/installation)
- [LLM Setup](https://docs.openhands.dev/usage/llms)

### For Developers
- [Development Guide](https://github.com/OpenHands/OpenHands/blob/main/Development.md)
- [Architecture Docs](https://docs.openhands.dev/usage/architecture/backend)
- [Contributing](https://github.com/OpenHands/OpenHands/blob/main/CONTRIBUTING.md)

### For Enterprise
- [Enterprise Page](https://openhands.dev/enterprise)
- [Contact Sales](https://www.all-hands.dev/contact)

---

*Last updated: April 2025*
