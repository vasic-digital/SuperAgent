# Amazon Q CLI - External References

## Official Documentation

### Primary Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **Official Docs** | https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line.html | Complete AWS documentation |
| **Installation Guide** | https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line-installing.html | Installation instructions |
| **GitHub Repo** | https://github.com/aws/amazon-q-developer-cli | Source repository |
| **AWS Console** | https://console.aws.amazon.com/ | AWS Management Console |

### AWS Service Documentation

| Resource | URL | Description |
|----------|-----|-------------|
| **CodeWhisperer** | https://docs.aws.amazon.com/codewhisperer/ | AI coding companion |
| **AWS CLI** | https://docs.aws.amazon.com/cli/ | AWS Command Line Interface |
| **IAM** | https://docs.aws.amazon.com/iam/ | Identity and Access Management |

---

## Kiro CLI (Successor)

> [!IMPORTANT]
> Amazon Q Developer CLI is now available as Kiro CLI, a closed-source product.

| Resource | URL | Description |
|----------|-----|-------------|
| **Kiro CLI** | https://kiro.dev/cli/ | Successor to Amazon Q CLI |
| **Kiro Documentation** | https://kiro.dev/docs/ | Kiro documentation |

---

## Community Resources

### Forums & Discussion

| Platform | Link | Description |
|----------|------|-------------|
| **GitHub Issues** | https://github.com/aws/amazon-q-developer-cli/issues | Bug reports & features |
| **AWS re:Post** | https://repost.aws/ | AWS community forum |
| **Stack Overflow** | https://stackoverflow.com/questions/tagged/amazon-q | Q&A |
| **Reddit** | https://reddit.com/r/aws | AWS community |

### Tutorials & Guides

| Source | Title | URL |
|--------|-------|-----|
| AWS | Getting Started with Amazon Q | https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/getting-started.html |
| AWS | Command Line Guide | https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line.html |

---

## Related Tools & Projects

### Similar Tools

| Tool | Description | Website |
|------|-------------|---------|
| **Kiro CLI** | AWS's new AI coding assistant | https://kiro.dev/cli/ |
| **Claude Code** | Anthropic's agentic coding tool | https://code.claude.com |
| **GitHub Copilot** | AI pair programmer | https://github.com/features/copilot |
| **Cursor** | AI-first code editor | https://cursor.com |
| **Aider** | AI pair programming | https://aider.chat |
| **Sourcegraph Cody** | AI coding assistant | https://sourcegraph.com/cody |

### Complementary Tools

| Tool | Purpose | Integration |
|------|---------|-------------|
| **MCP Servers** | External tool integration | Built-in support |
| **AWS CLI** | AWS operations | Native integration |
| **GitHub CLI** | GitHub operations | Via MCP |
| **Docker** | Containerization | Supported |

---

## MCP Server Registry

### Official MCP Servers

| Server | Description | Install |
|--------|-------------|---------|
| **GitHub** | GitHub integration | `npx @github/mcp-server` |
| **PostgreSQL** | Database access | `npx @modelcontextprotocol/server-postgres` |
| **SQLite** | SQLite integration | `npx @modelcontextprotocol/server-sqlite` |
| **Filesystem** | File operations | `npx @modelcontextprotocol/server-filesystem` |
| **Brave Search** | Web search | `npx @modelcontextprotocol/server-brave-search` |

### Community MCP Servers

| Server | Author | Description |
|--------|--------|-------------|
| **Kubernetes** | Community | K8s management |
| **Docker** | Community | Container management |
| **Terraform** | Community | Infrastructure as Code |

---

## Technical References

### Protocols & Standards

| Standard | Description | Link |
|----------|-------------|------|
| **Model Context Protocol** | Tool integration standard | https://modelcontextprotocol.io/ |
| **JSON-RPC 2.0** | MCP transport | https://www.jsonrpc.org/specification |
| **Server-Sent Events** | Streaming transport | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events |

### Rust Resources

| Resource | URL | Description |
|----------|-----|-------------|
| **Rust Book** | https://doc.rust-lang.org/book/ | Official Rust documentation |
| **Rustup** | https://rustup.rs | Rust toolchain installer |
| **Cargo** | https://doc.rust-lang.org/cargo/ | Rust package manager |
| **Clap** | https://docs.rs/clap/ | CLI argument parsing |

---

## AWS Resources

### Core Services

| Service | URL | Description |
|---------|-----|-------------|
| **AWS Home** | https://aws.amazon.com | Amazon Web Services |
| **AWS Documentation** | https://docs.aws.amazon.com | All AWS docs |
| **AWS Blog** | https://aws.amazon.com/blogs/aws/ | AWS news and updates |

### Security & Compliance

| Resource | URL |
|----------|-----|
| **AWS Security** | https://aws.amazon.com/security/ |
| **Vulnerability Reporting** | https://aws.amazon.com/security/vulnerability-reporting/ |
| **Privacy Policy** | https://aws.amazon.com/privacy/ |

---

## Development Resources

### Source Code References

This repository contains:

| Directory | Contents |
|-----------|----------|
| `crates/chat-cli/` | Main CLI application |
| `crates/agent/` | Agent runtime and tools |
| `crates/amzn-codewhisperer-client/` | AWS API client |
| `docs/` | Technical documentation |

### Configuration Examples

| Example | Location | Description |
|---------|----------|-------------|
| Agent Config | `docs/agent-format.md` | Agent JSON format |
| Tool Settings | `docs/built-in-tools.md` | Tool configuration |
| Hooks | `docs/hooks.md` | Hook examples |

---

## Version History

### Repository Status

> This open source project is no longer being actively maintained and will only receive critical security fixes.

| Version | Status | Notes |
|---------|--------|-------|
| Open Source | Maintenance only | Critical security fixes only |
| Kiro CLI | Active development | Successor product |

---

## Support Channels

| Channel | Purpose | Response Time |
|---------|---------|---------------|
| GitHub Issues | Bugs, features | Varies |
| AWS Support | Enterprise | Business hours |
| AWS re:Post | Community help | Community-driven |

---

## Contributing

While the open source project is in maintenance mode:

1. **Bug Reports**: Use GitHub Issues for critical bugs
2. **Security Issues**: Report via [AWS Vulnerability Reporting](https://aws.amazon.com/security/vulnerability-reporting/)
3. **Feature Requests**: Consider migrating to Kiro CLI

---

## License & Legal

| Document | Link |
|----------|------|
| MIT License | [LICENSE.MIT](../../../cli_agents/amazon-q/LICENSE.MIT) |
| Apache 2.0 License | [LICENSE.APACHE](../../../cli_agents/amazon-q/LICENSE.APACHE) |
| Contributing Guidelines | [CONTRIBUTING.md](../../../cli_agents/amazon-q/CONTRIBUTING.md) |
| Code of Conduct | [CODE_OF_CONDUCT.md](../../../cli_agents/amazon-q/CODE_OF_CONDUCT.md) |
| Security Policy | [SECURITY.md](../../../cli_agents/amazon-q/SECURITY.md) |

"Amazon Web Services" and all related marks are trademarks of AWS.

---

*Last updated: April 2025*
*Part of the HelixAgent CLI Agent Collection*
