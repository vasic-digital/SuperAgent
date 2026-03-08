# Module Interoperability Matrix

This document maps the dependency relationships between all 27 extracted modules in the
HelixAgent ecosystem. Each module is an independent Go module with its own `go.mod`.

Last updated: 2026-03-08

---

## Dependency Summary

Most modules are fully independent (zero inter-module dependencies). Only two modules
have cross-module dependencies:

- **HelixMemory** depends on **Memory** (`digital.vasic.memory`)
- **Challenges** depends on **Containers** (`digital.vasic.containers`)

All other modules depend only on third-party Go packages and the standard library.

---

## Dependency Matrix

Rows represent the module, columns represent what it depends on. An `X` indicates a
direct `require` dependency declared in the module's `go.mod`.

```
                        Depends On -->
                        EB Co Ob Au St Sr Se Ve Em Da Ca Me Fo MC RA Mm Op Pl Ag LO SI Pn Bm HM HS Cn Ch
Module (rows)           |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
----------------------- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
EventBus         (EB)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Concurrency      (Co)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Observability    (Ob)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Auth             (Au)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Storage          (St)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Streaming        (Sr)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Security         (Se)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
VectorDB         (Ve)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Embeddings       (Em)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Database         (Da)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Cache            (Ca)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Messaging        (Me)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Formatters       (Fo)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
MCP              (MC)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
RAG              (RA)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Memory           (Mm)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Optimization     (Op)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Plugins          (Pl)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Agentic          (Ag)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
LLMOps           (LO)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
SelfImprove      (SI)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Planning         (Pn)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Benchmark        (Bm)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
HelixMemory      (HM)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X  .  .  .  .  .  .  .  .  .  .  .
HelixSpecifier   (HS)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Containers       (Cn)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
Challenges       (Ch)   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X  .
```

**Legend**: `.` = no dependency, `X` = direct dependency

---

## Detailed Dependency Table

| # | Module | Go Module Path | Inter-Module Dependencies | Key Third-Party Dependencies |
|---|--------|----------------|--------------------------|------------------------------|
| 1 | EventBus | `digital.vasic.eventbus` | None | google/uuid |
| 2 | Concurrency | `digital.vasic.concurrency` | None | shirou/gopsutil |
| 3 | Observability | `digital.vasic.observability` | None | opentelemetry, prometheus, clickhouse-go, logrus |
| 4 | Auth | `digital.vasic.auth` | None | golang-jwt/jwt, google/uuid |
| 5 | Storage | `digital.vasic.storage` | None | minio/minio-go, logrus |
| 6 | Streaming | `digital.vasic.streaming` | None | gorilla/websocket, grpc |
| 7 | Security | `digital.vasic.security` | None | (testify only) |
| 8 | VectorDB | `digital.vasic.vectordb` | None | google/uuid |
| 9 | Embeddings | `digital.vasic.embeddings` | None | (testify only) |
| 10 | Database | `digital.vasic.database` | None | jackc/pgx, modernc.org/sqlite |
| 11 | Cache | `digital.vasic.cache` | None | redis/go-redis, alicebob/miniredis |
| 12 | Messaging | `digital.vasic.messaging` | None | google/uuid |
| 13 | Formatters | `digital.vasic.formatters` | None | (testify only) |
| 14 | MCP | `digital.vasic.mcp` | None | google/uuid |
| 15 | RAG | `digital.vasic.rag` | None | (testify only) |
| 16 | Memory | `digital.vasic.memory` | None | google/uuid |
| 17 | Optimization | `digital.vasic.optimization` | None | (testify only) |
| 18 | Plugins | `digital.vasic.plugins` | None | google/uuid, yaml.v3 |
| 19 | Agentic | `digital.vasic.agentic` | None | google/uuid, logrus |
| 20 | LLMOps | `digital.vasic.llmops` | None | google/uuid, logrus |
| 21 | SelfImprove | `digital.vasic.selfimprove` | None | google/uuid, logrus |
| 22 | Planning | `digital.vasic.planning` | None | logrus |
| 23 | Benchmark | `digital.vasic.benchmark` | None | google/uuid, logrus |
| 24 | HelixMemory | `digital.vasic.helixmemory` | **Memory** | prometheus, golang.org/x/sync |
| 25 | HelixSpecifier | `digital.vasic.helixspecifier` | None | google/uuid, logrus |
| 26 | Containers | `digital.vasic.containers` | None | prometheus, golang.org/x/crypto, yaml.v3 |
| 27 | Challenges | `digital.vasic.challenges` | **Containers** | gorilla/websocket, yaml.v3 |

---

## Dependency Graph

```
Phase 1 (Foundation) -- zero inter-module deps:
  EventBus  Concurrency  Observability  Auth  Storage  Streaming

Phase 2 (Infrastructure) -- zero inter-module deps:
  Security  VectorDB  Embeddings  Database  Cache

Phase 3 (Services) -- zero inter-module deps:
  Messaging  Formatters  MCP

Phase 4 (Integration) -- zero inter-module deps:
  RAG  Memory  Optimization  Plugins

Phase 5 (AI/ML) -- zero inter-module deps:
  Agentic  LLMOps  SelfImprove  Planning  Benchmark

Phase 6 (Cognitive):
  HelixMemory --depends-on--> Memory

Phase 7 (Specification) -- zero inter-module deps:
  HelixSpecifier

Pre-existing:
  Containers (independent)
  Challenges --depends-on--> Containers
```

---

## HelixAgent Root Module Integration

The root `go.mod` (`dev.helix.agent`) uses `replace` directives to point all 27 modules
to their local subdirectories. The root module imports all extracted modules through its
adapter layer (`internal/adapters/`), enabling the main application to compose
functionality from all modules.

```
dev.helix.agent
  |
  +-- internal/adapters/eventbus/      --> digital.vasic.eventbus
  +-- internal/adapters/auth/          --> digital.vasic.auth
  +-- internal/adapters/cache/         --> digital.vasic.cache
  +-- internal/adapters/cloud/         --> digital.vasic.storage
  +-- internal/adapters/containers/    --> digital.vasic.containers
  +-- internal/adapters/database/      --> digital.vasic.database
  +-- internal/adapters/formatters/    --> digital.vasic.formatters
  +-- internal/adapters/mcp/           --> digital.vasic.mcp
  +-- internal/adapters/memory/        --> digital.vasic.memory, digital.vasic.helixmemory
  +-- internal/adapters/messaging/     --> digital.vasic.messaging
  +-- internal/adapters/optimization/  --> digital.vasic.optimization
  +-- internal/adapters/plugins/       --> digital.vasic.plugins
  +-- internal/adapters/rag/           --> digital.vasic.rag
  +-- internal/adapters/security/      --> digital.vasic.security
  +-- internal/adapters/specifier/     --> digital.vasic.helixspecifier
  +-- internal/adapters/storage/       --> digital.vasic.storage
  +-- internal/adapters/streaming/     --> digital.vasic.streaming
  +-- internal/adapters/vectordb/      --> digital.vasic.vectordb
  +-- internal/adapters/agentic/       --> digital.vasic.agentic
  +-- internal/adapters/llmops/        --> digital.vasic.llmops
  +-- internal/adapters/selfimprove/   --> digital.vasic.selfimprove
  +-- internal/adapters/planning/      --> digital.vasic.planning
  +-- internal/adapters/benchmark/     --> digital.vasic.benchmark
  ...
```

---

## Design Principles

1. **Maximum independence**: 25 of 27 modules have zero inter-module dependencies.
   This allows any module to be used standalone in other projects.

2. **Layered phases**: Modules are grouped by phase (Foundation through Specification).
   Higher phases may depend on lower phases but never the reverse.

3. **Adapter isolation**: The root HelixAgent application never imports modules directly
   in business logic. All module access goes through `internal/adapters/`, enabling
   type conversion and backward compatibility.

4. **Replace directives for development**: Local development uses `replace` directives
   in the root `go.mod` so changes to submodules are immediately reflected without
   publishing new versions.

5. **No circular dependencies**: The dependency graph is a strict DAG (directed acyclic
   graph) with only two edges: HelixMemory->Memory and Challenges->Containers.
