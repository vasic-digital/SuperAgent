<!--
Sync Impact Report:
Version change: null → 1.0.0
Modified principles: N/A (initial creation)
Added sections: Core Principles (5), Technology Stack, Testing Requirements, Development Workflow
Removed sections: N/A
Templates requiring updates: All templates need review for Go/LLM context
Follow-up TODOs: Implement SpecKit integration check
-->

# SuperAgent Constitution

## Core Principles

### I. Model Facade Architecture
SuperAgent MUST expose a unified LLM facade that abstracts multiple underlying models as a single super agent. The system MUST support DeepSeek, Qwen, Z.AI (GLM), Claude, Gemini, and other top LLMs. All configured models MUST work cooperatively to deliver optimal results while maintaining full feature parity including tooling, reasoning, search, and MCP capabilities.

### II. Go-First Implementation (NON-NEGOTIABLE)
All implementation MUST be in Go Lang using Gin Gonic framework. HTTP3/Quic with Toon MUST be the default protocol with fallback to HTTP2/JSON. The codebase MUST be clean, extensible, and follow Go best practices with comprehensive documentation for every component.

### III. Comprehensive Testing Discipline (NON-NEGOTIABLE)
Every feature MUST be covered by: Unit tests, Integration tests, E2E tests, Full automation tests, Stress/Benchmark tests, Security tests, and Challenge tests. Test coverage MUST be 100% with 100% execution success. NO broken, disabled, or incomplete tests are permitted.

### IV. Security-First Development
All code MUST pass SonarQube and Snyk security scans. Postgres or SQLite with SQLCipher encryption MUST be used for data persistence. Sensitive data MUST never be committed. .gitignore files MUST be properly configured to prevent credential exposure. Security tests are mandatory for all releases.

### V. Plugin-Based Extensibility
The architecture MUST support easy addition of new LLM providers through a plugin and strategy system. All interfaces MUST be designed for extensibility. New models, providers, and features MUST be addable without breaking existing functionality.

## Technology Stack Requirements

**Language**: Go 1.21+ (MANDATORY)  
**Framework**: Gin Gonic (MANDATORY)  
**Protocols**: HTTP3/Quic with Toon (default), HTTP2/JSON (fallback)  
**Database**: Postgres (production) or SQLite with SQLCipher (development/lightweight)  
**Security**: SonarQube and Snyk scanning (MANDATORY)  
**Containerization**: Docker for all components (MANDATORY)  
**Compatibility**: OpenAI API standards for all exposed interfaces (MANDATORY)

## Testing Requirements

### Mandatory Test Types
1. **Unit Tests**: Every function and method MUST have unit tests
2. **Integration Tests**: All component interactions MUST be tested
3. **E2E Tests**: Real AI QA execution using system as regular user
4. **Stress/Benchmark Tests**: Performance and load testing with complex data
5. **Security Tests**: Automated SonarQube and Snyk vulnerability scanning
6. **Challenge Tests**: Complete project implementation missions validating production readiness

### Testing Standards
- 100% test coverage REQUIRED
- 100% test execution success REQUIRED
- All tests MUST be automated and repeatable
- Test database MUST track all execution results and discovered issues
- EVERY discovered issue MUST be fixed before release

## Development Workflow

### SpecKit Integration (NON-NEGOTIABLE)
All development MUST follow the SpecKit cycle: Constitution → Specify → Clarify → Plan → Tasks → Analyze → Implement. The system MUST check for SpecKit availability and install/configure locally if not present.

### Quality Gates
- Code MUST pass all linting and formatting checks
- All tests MUST pass (100% success)
- Security scans MUST show zero vulnerabilities
- Documentation MUST be complete for all changes
- Challenge tests MUST validate production readiness

### Documentation Requirements
- Complete technical documentation for every component
- User manuals from beginner to advanced levels
- API documentation with examples
- Installation and configuration guides
- Troubleshooting and maintenance procedures

## Governance

This constitution supersedes all other development practices. Amendments require documentation, approval, and migration plan. All PRs and reviews MUST verify constitutional compliance. Any complexity beyond constitutional requirements MUST be explicitly justified. Use .specify/templates/ for runtime development guidance.

**Version**: 1.0.0 | **Ratified**: 2025-12-08 | **Last Amended**: 2025-12-08
