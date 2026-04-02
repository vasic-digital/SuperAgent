# Claude Code - External References

## Official Documentation

### Primary Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **Official Docs** | https://code.claude.com/docs/en/overview | Complete documentation |
| **Setup Guide** | https://code.claude.com/docs/en/setup | Installation instructions |
| **Data Usage** | https://code.claude.com/docs/en/data-usage | Privacy and data policies |
| **GitHub Repo** | https://github.com/anthropics/claude-code | Source repository |
| **NPM Package** | https://www.npmjs.com/package/@anthropic-ai/claude-code | Package registry |

### API Documentation

| Resource | URL | Description |
|----------|-----|-------------|
| **Anthropic API** | https://docs.anthropic.com | Claude API reference |
| **Agent SDK** | https://docs.claude.com/en/api/agent-sdk/overview | SDK documentation |
| **MCP Docs** | https://modelcontextprotocol.io/ | Model Context Protocol |

---

## Community Resources

### Forums & Discussion

| Platform | Link | Description |
|----------|------|-------------|
| **Discord** | https://anthropic.com/discord | Official community server |
| **GitHub Issues** | https://github.com/anthropics/claude-code/issues | Bug reports & features |
| **GitHub Discussions** | https://github.com/anthropics/claude-code/discussions | Community discussions |
| **Reddit** | https://reddit.com/r/ClaudeAI | Community forum |
| **Hacker News** | Search "Claude Code" | Tech discussions |

### Tutorials & Guides

| Source | Title | URL |
|--------|-------|-----|
| Anthropic | Getting Started | https://code.claude.com/docs/en/getting-started |
| Anthropic | Best Practices | https://code.claude.com/docs/en/best-practices |
| Anthropic | Advanced Usage | https://code.claude.com/docs/en/advanced |

### Video Content

| Platform | Search Query | Description |
|----------|--------------|-------------|
| YouTube | "Claude Code tutorial" | Video tutorials |
| YouTube | "Anthropic Claude Code" | Official content |
| Twitter/X | @AnthropicAI | Official announcements |

---

## Related Tools & Projects

### Similar Tools

| Tool | Description | Website |
|------|-------------|---------|
| **GitHub Copilot** | AI pair programmer | https://github.com/features/copilot |
| **Cursor** | AI-first code editor | https://cursor.com |
| **Aider** | AI pair programming | https://aider.chat |
| **Continue** | Open-source AI assistant | https://continue.dev |
| **Sourcegraph Cody** | AI coding assistant | https://sourcegraph.com/cody |

### Complementary Tools

| Tool | Purpose | Integration |
|------|---------|-------------|
| **MCP Servers** | External tool integration | Built-in support |
| **GitHub CLI** | GitHub operations | Native integration |
| **tmux** | Terminal multiplexing | Supported |
| **ripgrep** | Fast code search | Used internally |
| **fzf** | Fuzzy finder | Compatible |

---

## MCP Server Registry

### Official MCP Servers

| Server | Description | Install |
|--------|-------------|---------|
| **GitHub** | GitHub integration | `npx @github/mcp-server` |
| **PostgreSQL** | Database access | `npx @modelcontextprotocol/server-postgres` |
| **SQLite** | SQLite integration | `npx @modelcontextprotocol/server-sqlite` |
| **Brave Search** | Web search | `npx @modelcontextprotocol/server-brave-search` |
| **Filesystem** | File operations | `npx @modelcontextprotocol/server-filesystem` |

### Community MCP Servers

| Server | Author | Description |
|--------|--------|-------------|
| **Notion** | Notion | Notion workspace access |
| **Slack** | Community | Slack integration |
| **Discord** | Community | Discord bot |
| **AWS** | Community | AWS operations |
| **Kubernetes** | Community | K8s management |

---

## Technical References

### Protocols & Standards

| Standard | Description | Link |
|----------|-------------|------|
| **Model Context Protocol** | Tool integration standard | https://modelcontextprotocol.io/ |
| **JSON-RPC 2.0** | MCP transport | https://www.jsonrpc.org/specification |
| **Server-Sent Events** | Streaming transport | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events |

### Anthropic Resources

| Resource | URL |
|----------|-----|
| **Anthropic Home** | https://anthropic.com |
| **Claude.ai** | https://claude.ai |
| **API Console** | https://console.anthropic.com |
| **Research** | https://anthropic.com/research |
| **Safety** | https://anthropic.com/safety |

---

## Tutorials & Guides (Web Research)

### Comprehensive Tutorials

| Resource | Author | Description | URL |
|----------|--------|-------------|-----|
| **Complete Guide 2026** | Sid Bharath | Comprehensive tutorial with Opus 4.6 features | https://sidbharath.com/blog/claude-code-the-complete-guide/ |
| **MCP Servers Guide** | Builder.io | How to connect and configure MCP servers | https://www.builder.io/blog/claude-code-mcp-servers |
| **MCP Course** | Anthropic | Official MCP introduction course | https://anthropic.skilljar.com/introduction-to-model-context-protocol |
| **Best MCP Servers** | ComparePriceAcross | Curated list of MCP servers | https://www.comparepriceacross.com/post/best_mcp_servers_for_claude_code/ |

### Key Insights from Research

1. **Context Management Best Practices** (from sidbharath.com):
   - Scope chat to one project/feature
   - Use `/clear` when done with a feature
   - Use `/compact` with specific instructions to preserve important context
   - Newer Sonnet has 1M token context window

2. **Chat Modes** (updated features):
   - **Default Mode**: Interactive with permission prompts
   - **Auto Mode**: Autonomous execution with safety controls
   - **Fast Mode**: 2.5x faster output using Opus 4.6
   - **Plan Mode**: Extended thinking for complex strategies

3. **Advanced Features** (2025-2026 updates):
   - **Automatic Memory**: MEMORY.md files in `~/.claude/projects/`
   - **Agent Teams**: Experimental parallel sub-agents that communicate
   - **Async Sub-agents**: Background task execution with Ctrl+B
   - **Hierarchical CLAUDE.md**: Multiple CLAUDE.md files at different levels
   - **Git Worktrees**: Parallel development with multiple Claude instances

4. **MCP Server Recommendations**:
   - **GitHub**: Repository management and PR workflows
   - **Playwright**: Browser automation and testing
   - **Sentry**: Production error monitoring
   - **PostgreSQL/Supabase**: Database access
   - **Figma**: Design-to-code workflows

5. **Security Considerations** (CVE-2025-58764):
   - RCE vulnerability patched in v1.0.105
   - Always keep Claude Code updated
   - Be cautious with untrusted content sources

### Community Discussions

| Platform | Topics |
|----------|--------|
| **Reddit r/ClaudeAI** | Usage tips, feature discussions |
| **Hacker News** | Technical deep-dives |
| **Twitter/X** | Real-time updates, user experiences |
| **Discord** | Community support, troubleshooting |

## Academic & Research

### Papers

| Paper | Authors | Link |
|-------|---------|------|
| "Constitutional AI" | Anthropic | https://arxiv.org/abs/2212.08073 |
| "Training Helpful AI" | Anthropic | Research blog posts |
| Claude Technical Reports | Anthropic | https://anthropic.com/research |

### Industry Reports

| Report | Source | Description |
|--------|--------|-------------|
| Claude Code Transformation | Apidog | Real-action examples and workflows |
| Engineering Deep Dive | Pragmatic Engineer | How Claude Code is built |
| Enterprise Adoption | CNBC | $350B valuation context |

### Benchmarks

| Benchmark | Description | Relevance |
|-----------|-------------|-----------|
| **SWE-bench** | Software engineering tasks | Claude Code evaluation |
| **HumanEval** | Code generation | Model capability |
| **APPS** | Competitive programming | Reasoning ability |

---

## Development Resources

### Source Code References

This repository contains:

| Directory | Contents |
|-----------|----------|
| `plugins/` | 14 official plugins |
| `.claude/commands/` | Example custom commands |
| `examples/` | Configuration examples |
| `scripts/` | Automation scripts |

### Configuration Examples

| Example | Location | Description |
|---------|----------|-------------|
| Settings - Lax | `examples/settings/settings-lax.json` | Permissive config |
| Settings - Strict | `examples/settings/settings-strict.json` | Restrictive config |
| Bash Sandbox | `examples/settings/settings-bash-sandbox.json` | Isolated bash |
| Hooks | `examples/hooks/` | Hook examples |

---

## Version History

### Recent Releases

| Version | Date | Key Features |
|---------|------|--------------|
| v2.1.90 | Apr 2025 | `/powerup` lessons, bug fixes |
| v2.1.89 | Mar 2025 | PermissionDenied hook, defer decision |
| v2.1.88 | Mar 2025 | Bug fixes |
| v2.1.87 | Mar 2025 | Cowork Dispatch fixes |
| v2.1.86 | Feb 2025 | Session ID headers, improvements |

See full changelog: [CHANGELOG.md](../../../cli-agents/claude-code/CHANGELOG.md)

---

## Support Channels

| Channel | Purpose | Response Time |
|---------|---------|---------------|
| `/bug` command | Bug reports | Immediate logging |
| GitHub Issues | Bugs, features | Varies |
| Discord | Community help | Real-time |
| Email | Enterprise | Business hours |

---

## License & Legal

| Document | Link |
|----------|------|
| License | [LICENSE.md](../../../cli-agents/claude-code/LICENSE.md) |
| Privacy Policy | https://www.anthropic.com/legal/privacy |
| Commercial Terms | https://www.anthropic.com/legal/commercial-terms |
| Security | [SECURITY.md](../../../cli-agents/claude-code/SECURITY.md) |

---

## Contributing

While Claude Code itself is not open-source, community contributions are welcome:

1. **Plugin Development**: Create and share plugins
2. **Documentation**: Improve docs and tutorials
3. **Bug Reports**: Use `/bug` or GitHub Issues
4. **Feedback**: Share experiences and suggestions

---

*Last updated: April 2025*
