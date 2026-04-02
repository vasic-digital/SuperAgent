# Forge - External References

## Official Documentation

### Primary Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **Official Website** | https://forgecode.dev | Main website and documentation |
| **GitHub Repository** | https://github.com/antinomyhq/forge | Source code and issues |
| **Agent Configuration** | https://forgecode.dev/docs/agent-configuration | Agent setup guide |
| **Discord Community** | https://discord.gg/kRZBPpkgwq | Community support |

### NPM Package

| Package | Description | URL |
|---------|-------------|-----|
| `@antinomyhq/forge` | Forge CLI package | https://www.npmjs.com/package/@antinomyhq/forge |

---

## Installation & Setup

### Quick Install

```bash
# Official installer
curl -fsSL https://forgecode.dev/cli | sh

# Alternative URLs
curl -fsSL https://forgecode.dev/install.sh | sh
```

### Nix Installation

```bash
# Run latest dev branch
nix run github:antinomyhq/forge
```

### Building from Source

```bash
git clone https://github.com/antinomyhq/forge.git
cd forge
cargo build --release
```

---

## Tutorials & Guides

### Official Documentation

| Guide | URL | Description |
|-------|-----|-------------|
| Installation Guide | https://forgecode.dev/docs/installation | Step-by-step setup |
| Agent Configuration | https://forgecode.dev/docs/agent-configuration | Creating custom agents |
| Provider Setup | https://forgecode.dev/docs/providers | LLM provider configuration |

### Community Tutorials

| Resource | Author | Description | URL |
|----------|--------|-------------|-----|
| Productivity Guide | Dev.to Community | 10 ways CLI coding boosts productivity | https://dev.to/pankaj_singh_1022ee93e755 |
| Forge on Trendshift | Trendshift | Repository analytics and insights | https://trendshift.io/repository/13878 |

### Video Content

| Platform | Search Query | Description |
|----------|--------------|-------------|
| YouTube | "Forge AI coding agent" | Tutorial videos |
| YouTube | "antinomyhq forge" | Official content |

---

## Community Resources

### Forums & Discussion

| Platform | Link | Description |
|----------|------|-------------|
| **Discord** | https://discord.gg/kRZBPpkgwq | Official community server |
| **GitHub Issues** | https://github.com/antinomyhq/forge/issues | Bug reports & features |
| **GitHub Discussions** | https://github.com/antinomyhq/forge/discussions | Community discussions |
| **Hacker News** | Search "Forge AI coding" | Tech discussions |

### Social Media

| Platform | Handle | Description |
|----------|--------|-------------|
| **Twitter/X** | @antinomyhq | Official announcements |

---

## Related Tools & Projects

### Similar AI Coding Agents

| Tool | Description | Website |
|------|-------------|---------|
| **Claude Code** | Anthropic's official CLI agent | https://code.claude.com |
| **Aider** | AI pair programming in terminal | https://aider.chat |
| **GitHub Copilot CLI** | GitHub's CLI assistant | https://github.com/features/copilot |
| **Cursor** | AI-first code editor | https://cursor.com |
| **Continue.dev** | Open-source AI assistant | https://continue.dev |

### Forge-Based Projects

| Project | Description | URL |
|---------|-------------|-----|
| **ForgeCode** | Forge provider service | https://forgecode.dev |
| **Playbooks** | Forge skill sharing | https://playbooks.com/skills |

### Complementary Tools

| Tool | Purpose | Integration |
|------|---------|-------------|
| **MCP Servers** | External tool integration | Built-in support |
| **GitHub CLI** | GitHub operations | Native integration |
| **ripgrep** | Fast code search | Used internally |
| **fzf** | Fuzzy finder | Compatible |
| **Nerd Fonts** | Terminal icons | Recommended |

---

## MCP Server Registry

### Official MCP Servers

| Server | Description | Install Command |
|--------|-------------|-----------------|
| **GitHub** | Repository management | `npx @github/mcp-server` |
| **PostgreSQL** | Database access | `npx @modelcontextprotocol/server-postgres` |
| **SQLite** | SQLite integration | `npx @modelcontextprotocol/server-sqlite` |
| **Brave Search** | Web search | `npx @modelcontextprotocol/server-brave-search` |
| **Filesystem** | File operations | `npx @modelcontextprotocol/server-filesystem` |
| **Puppeteer** | Browser automation | `npx @modelcontextprotocol/server-puppeteer` |

### Community MCP Servers

| Server | Author | Description |
|--------|--------|-------------|
| **Notion** | Community | Notion workspace access |
| **Slack** | Community | Slack integration |
| **Discord** | Community | Discord bot |
| **AWS** | Community | AWS operations |
| **Kubernetes** | Community | K8s management |
| **Figma** | Community | Design-to-code workflows |

### MCP Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **MCP Documentation** | https://modelcontextprotocol.io | Official protocol docs |
| **MCP Specification** | https://spec.modelcontextprotocol.io | Protocol specification |
| **MCP Servers List** | https://github.com/modelcontextprotocol/servers | Official server repository |

---

## LLM Provider Resources

### Supported Providers

| Provider | Models | Docs URL |
|----------|--------|----------|
| **Anthropic** | Claude 3.5/4 Sonnet, Opus | https://docs.anthropic.com |
| **OpenAI** | GPT-4, GPT-4o, o1, o3 | https://platform.openai.com/docs |
| **OpenRouter** | 300+ models | https://openrouter.ai/docs |
| **Google** | Gemini Pro, Flash | https://ai.google.dev/docs |
| **xAI** | Grok | https://docs.x.ai |
| **Cerebras** | Llama models | https://docs.cerebras.ai |
| **Groq** | Mixtral, Llama | https://console.groq.com/docs |

### API Key Management

| Provider | Key Location | Pricing |
|----------|--------------|---------|
| **Anthropic** | https://console.anthropic.com | Usage-based |
| **OpenAI** | https://platform.openai.com/api-keys | Usage-based |
| **OpenRouter** | https://openrouter.ai/keys | Usage-based |
| **Google** | https://aistudio.google.com/app/apikey | Free tier available |

---

## Technical References

### Protocols & Standards

| Standard | Description | Link |
|----------|-------------|------|
| **Model Context Protocol** | Tool integration standard | https://modelcontextprotocol.io |
| **JSON-RPC 2.0** | MCP transport | https://www.jsonrpc.org/specification |
| **Server-Sent Events** | Streaming transport | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events |

### Rust Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **Rust Book** | https://doc.rust-lang.org/book | Official Rust guide |
| **Rust By Example** | https://doc.rust-lang.org/rust-by-example | Learn by example |
| **Cargo Book** | https://doc.rust-lang.org/cargo | Package manager docs |
| **Tokio** | https://tokio.rs | Async runtime used by Forge |

---

## Skills & Playbooks

### Available Skills

Skills are available from the Playbooks registry:

| Skill | Description | URL |
|-------|-------------|-----|
| **pr-description** | Generate PR descriptions | https://playbooks.com/skills/antinomyhq/forge/pr-description |
| **create-agent** | Create custom agents | Built-in |
| **create-command** | Create custom commands | Built-in |
| **debug-cli** | Debug CLI issues | Built-in |
| **resolve-conflicts** | Git conflict resolution | Built-in |

### Creating Custom Skills

Skills are defined in `.forge/skills/`:

```markdown
---
name: my-skill
description: What this skill does
---

Skill instructions here...
```

---

## Version History

### Recent Releases

| Version | Date | Key Features |
|---------|------|--------------|
| Latest | Ongoing | See GitHub releases |

See full changelog: https://github.com/antinomyhq/forge/releases

---

## Support Channels

| Channel | Purpose | Response Time |
|---------|---------|---------------|
| Discord | Community help | Real-time |
| GitHub Issues | Bugs, features | Varies |
| Documentation | Self-service | Immediate |

---

## Contributing

### Ways to Contribute

1. **Report Bugs** - Use GitHub Issues
2. **Suggest Features** - Open a feature request
3. **Share Skills** - Publish to Playbooks
4. **Write Documentation** - Improve docs
5. **Spread the Word** - Star on GitHub

### Development Setup

```bash
# Clone repository
git clone https://github.com/antinomyhq/forge.git
cd forge

# Install dependencies (Rust toolchain)
rustup update

# Build
cargo build

# Run tests
cargo test

# Run with local changes
cargo run
```

---

## License & Legal

| Document | Link |
|----------|------|
| License | [Apache-2.0](https://github.com/antinomyhq/forge/blob/main/LICENSE) |
| Code of Conduct | [GitHub](https://github.com/antinomyhq/forge/blob/main/CODE_OF_CONDUCT.md) |
| Contributing | [GitHub](https://github.com/antinomyhq/forge/blob/main/CONTRIBUTING.md) |

---

## Research Insights

### Key Insights from Community

1. **Terminal-Native Advantages**:
   - No context switching between browser and IDE
   - Integrates with existing shell workflows
   - Works over SSH on remote servers
   - Keyboard-centric efficiency

2. **Multi-Provider Benefits**:
   - Access to 300+ models via OpenRouter
   - Choose models based on task requirements
   - Cost optimization opportunities
   - Fallback provider options

3. **Agent Specialization**:
   - Forge for implementation
   - Sage for research
   - Muse for planning
   - Custom agents for team-specific needs

4. **MCP Ecosystem**:
   - Extensible tool system
   - Growing server library
   - Standardized protocol
   - Community contributions

---

## Comparison with Alternatives

### Feature Comparison

| Feature | Forge | Claude Code | Aider | Copilot CLI |
|---------|-------|-------------|-------|-------------|
| Terminal-based | ✅ | ✅ | ✅ | ✅ |
| Multi-provider | ✅ (300+) | ❌ (Anthropic only) | ✅ | ❌ (GitHub) |
| Custom agents | ✅ | ❌ | ❌ | ❌ |
| MCP support | ✅ | ✅ | ❌ | ❌ |
| Open source | ✅ | ❌ | ✅ | ❌ |
| Zsh plugin | ✅ | ❌ | ❌ | ❌ |
| Semantic search | ✅ | ❌ | ❌ | ❌ |
| Self-hostable | ✅ | ❌ | ✅ | ❌ |

### When to Choose Forge

- **Multiple LLM providers**: Best-in-class multi-provider support
- **Custom workflows**: Agent and skill customization
- **Team collaboration**: Shared agents and commands
- **Terminal-first workflow**: Native shell integration
- **Open source**: Community-driven development

---

## Additional Resources

### Blogs & Articles

| Title | Source | Topics |
|-------|--------|--------|
| "AI-Assisted Rust Development" | Shuttle.dev | Rust + AI workflow |

### Podcasts & Talks

Search for:
- "AI coding agents"
- "Terminal-based AI"
- "Rust development tools"

### Academic Papers

| Paper | Authors | Link |
|-------|---------|------|
| Constitutional AI | Anthropic | https://arxiv.org/abs/2212.08073 |
| Training Helpful AI | Anthropic | https://anthropic.com/research |

---

*Last updated: April 2025*
