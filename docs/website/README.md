# HelixAgent Website Documentation

This directory contains content for the HelixAgent public website. Each file corresponds to a section or page on the website and is written in Markdown format for easy conversion to HTML.

## Directory Structure

| File | Description | Target Audience |
|------|-------------|-----------------|
| `LANDING_PAGE.md` | Hero section, key value propositions, call-to-action | All visitors |
| `FEATURES.md` | Comprehensive feature list with descriptions | Technical evaluators |
| `ARCHITECTURE.md` | System architecture and design patterns | Architects, senior engineers |
| `INTEGRATIONS.md` | Integration guides for providers and services | Developers implementing |
| `GETTING_STARTED.md` | Quick start guide for new users | New users, developers |
| `SECURITY.md` | Security features and compliance information | Security teams, enterprise |
| `BIGDATA.md` | Big data integration capabilities | Data engineers, architects |
| `GRPC_API.md` | gRPC API documentation | Backend developers |
| `MEMORY_SYSTEM.md` | Memory and knowledge graph documentation | AI engineers |

## Content Guidelines

### Writing Style

- **Professional but approachable** - Avoid excessive jargon while maintaining technical accuracy
- **Benefit-focused** - Lead with outcomes, follow with technical details
- **Scannable** - Use headers, bullet points, and code blocks liberally
- **Consistent terminology** - Use terms as defined in the glossary below

### Glossary

| Term | Definition |
|------|------------|
| **HelixAgent** | The AI-powered ensemble LLM service |
| **Ensemble** | Multiple LLMs working together via AI debate |
| **AI Debate** | The consensus mechanism combining multiple model responses |
| **Provider** | An LLM service (Claude, DeepSeek, Gemini, etc.) |
| **Virtual Model** | The unified `helixagent-debate` model exposed to clients |

### Code Examples

All code examples should:
- Be complete and runnable
- Include necessary imports
- Use realistic variable names
- Include error handling where appropriate

### Formatting Standards

- Use ATX-style headers (`#`, `##`, `###`)
- Use fenced code blocks with language identifiers
- Tables should use GitHub-flavored Markdown
- Links should be relative where possible

## Conversion Notes

When converting to website:

1. **Images** - Replace placeholder references with actual image paths
2. **Navigation** - Add appropriate navigation links between pages
3. **SEO** - Extract meta descriptions from opening paragraphs
4. **Analytics** - Add tracking codes as per marketing requirements

## Related Documentation

- `/docs/api/` - Detailed API reference
- `/docs/guides/` - Step-by-step implementation guides
- `/docs/architecture/` - In-depth architecture documents
- `/CLAUDE.md` - Technical reference for developers

## Maintenance

- Review quarterly for accuracy
- Update version numbers with each release
- Add new features as they ship
- Remove deprecated features after deprecation period

---

**Last Updated**: February 2026
**Maintainer**: HelixAgent Documentation Team
