# HelixAgent Constitution

**Version:** 1.0.0
**Created:** 2026-02-10
**Updated:** 2026-02-10

Constitution with 20 rules (20 mandatory) across categories: Quality: 2, Safety: 1, Security: 1, Performance: 2, Containerization: 1, Configuration: 1, Testing: 3, Documentation: 2, Principles: 2, Stability: 1, Observability: 1, GitOps: 1, CI/CD: 1, Architecture: 1

## Performance

### Monitoring and Metrics **[MANDATORY]** (Priority: 2)

**ID:** CONST-009

Create tests that run and perform monitoring and metrics collection. Use collected data for proper optimizations.

### Lazy Loading and Non-Blocking **[MANDATORY]** (Priority: 2)

**ID:** CONST-010

Implement lazy loading and lazy initialization wherever possible. Introduce semaphore mechanisms and non-blocking mechanisms to ensure flawless responsiveness.

## Configuration

### Unified Configuration **[MANDATORY]** (Priority: 2)

**ID:** CONST-016

CLI agent config export uses only HelixAgent + LLMsVerifier's unified generator. No third-party scripts.

## Observability

### Health and Monitoring **[MANDATORY]** (Priority: 2)

**ID:** CONST-017

Every service MUST expose health endpoints. Circuit breakers for all external dependencies. Prometheus/OpenTelemetry integration.

## GitOps

### GitSpec Compliance **[MANDATORY]** (Priority: 2)

**ID:** CONST-018

Follow GitSpec constitution and all constraints from AGENTS.md and CLAUDE.md.

## CI/CD

### Manual CI/CD Only **[MANDATORY]** (Priority: 1)

**ID:** CONST-019

NO GitHub Actions enabled. All CI/CD workflows and pipelines must be executed manually only.

## Documentation

### Complete Documentation **[MANDATORY]** (Priority: 1)

**ID:** CONST-004

Every module and feature MUST have complete documentation: README.md, CLAUDE.md, AGENTS.md, user guides, step-by-step manuals, video courses, diagrams, SQL definitions, and website content. No component can remain undocumented.

### Documentation Synchronization **[MANDATORY]** (Priority: 1)

**ID:** CONST-020

Anything added to Constitution MUST be present in AGENTS.md and CLAUDE.md, and vice versa. Keep all three synchronized.

## Quality

### No Broken Components **[MANDATORY]** (Priority: 1)

**ID:** CONST-005

No module, application, library, or test can remain broken, disabled, or incomplete. Everything must be fully functional and operational.

### No Dead Code **[MANDATORY]** (Priority: 1)

**ID:** CONST-006

Identify and remove all 'dead code' - features or functionalities left unconnected with the system. Perform comprehensive research and cleanup.

## Safety

### Memory Safety **[MANDATORY]** (Priority: 1)

**ID:** CONST-007

Perform comprehensive research for memory leaks, deadlocks, and race conditions. Apply safety fixes and improvements to prevent these issues.

## Security

### Security Scanning **[MANDATORY]** (Priority: 1)

**ID:** CONST-008

Execute Snyk and SonarQube scanning. Analyze findings in depth and resolve everything. Ensure scanning infrastructure is accessible via containerization (Docker/Podman).

## Principles

### Software Principles **[MANDATORY]** (Priority: 2)

**ID:** CONST-011

Apply all software principles: KISS, DRY, SOLID, YAGNI, etc. Ensure code is clean, maintainable, and follows best practices.

### Design Patterns **[MANDATORY]** (Priority: 2)

**ID:** CONST-012

Use appropriate design patterns: Proxy, Facade, Factory, Abstract Factory, Observer, Mediator, Strategy, etc. Apply patterns where they add value.

## Stability

### Rock-Solid Changes **[MANDATORY]** (Priority: 1)

**ID:** CONST-013

All changes must be safe, non-error-prone, and MUST NOT BREAK any existing working functionality. Ensure backward compatibility unless explicitly breaking.

## Containerization

### Full Containerization **[MANDATORY]** (Priority: 2)

**ID:** CONST-015

All services MUST run in containers (Docker/Podman/K8s). Support local default execution AND remote configuration. Services must auto-boot before HelixAgent is ready.

## Architecture

### Comprehensive Decoupling **[MANDATORY]** (Priority: 1)

**ID:** CONST-001

Identify all parts and functionalities that can be extracted as separate modules (libraries) and reused in various projects. Perform additional work to make each module fully decoupled and independent. Each module must be a separate project with its own CLAUDE.md, AGENTS.md, README.md, docs/, tests, and challenges.

## Testing

### 100% Test Coverage **[MANDATORY]** (Priority: 1)

**ID:** CONST-002

Every component MUST have 100% test coverage across ALL test types: unit, integration, E2E, security, stress, chaos, automation, and benchmark tests. No false positives. Use real data and live services (mocks only in unit tests).

### Comprehensive Challenges **[MANDATORY]** (Priority: 1)

**ID:** CONST-003

Every component MUST have Challenge scripts validating real-life use cases. No false success - validate actual behavior, not return codes.

### Stress and Integration Tests **[MANDATORY]** (Priority: 2)

**ID:** CONST-014

Introduce comprehensive stress and integration tests validating that the system is responsive and not possible to overload or break.

