# internal/adapters

## Overview

The adapters package is the bridge layer that connects HelixAgent's internal
types to the extracted submodule interfaces. Each adapter translates between
internal domain models and the module-specific types, enabling clean
decoupling while maintaining full integration.

## Architecture

```
HelixAgent internal types  <-->  Adapter  <-->  Extracted module types
```

Adapters follow a consistent pattern: they wrap the extracted module's API,
translate types bidirectionally, and expose a simplified interface that
HelixAgent's service layer consumes.

## Adapter Catalog

| Subdirectory | Bridges To | Key File |
|-------------|------------|----------|
| `agentic/` | `digital.vasic.agentic` | `adapter.go` |
| `auth/` | `digital.vasic.auth` | `adapter.go`, `integration.go` |
| `benchmark/` | `digital.vasic.benchmark` | `adapter.go` |
| `cache/` | `digital.vasic.cache` | `adapter.go` |
| `cloud/` | Cloud storage providers | `adapter.go`, `client.go` |
| `containers/` | `digital.vasic.containers` | `adapter.go` |
| `database/` | `digital.vasic.database` | `adapter.go`, `compat.go` |
| `formatters/` | `digital.vasic.formatters` | `adapter.go` |
| `llmops/` | `digital.vasic.llmops` | `adapter.go` |
| `mcp/` | `digital.vasic.mcp` | `mcp.go` |
| `memory/` | `digital.vasic.memory` / `digital.vasic.helixmemory` | `adapter.go`, `factory_*.go` |
| `messaging/` | `digital.vasic.messaging` | `adapter.go`, `kafka_adapter.go`, `rabbitmq_adapter.go` |
| `optimization/` | `digital.vasic.optimization` | `adapter.go` |
| `planning/` | `digital.vasic.planning` | `adapter.go` |
| `plugins/` | `digital.vasic.plugins` | `adapter.go` |
| `rag/` | `digital.vasic.rag` | `adapter.go` |
| `security/` | `digital.vasic.security` | `security.go` |
| `selfimprove/` | `digital.vasic.selfimprove` | `adapter.go` |
| `specifier/` | `digital.vasic.helixspecifier` | `adapter.go`, `factory_*.go` |
| `storage/minio/` | `digital.vasic.storage` (MinIO) | `adapter.go` |
| `streaming/` | `digital.vasic.streaming` | `adapter.go` |
| `vectordb/qdrant/` | `digital.vasic.vectordb` (Qdrant) | `adapter.go` |

The root-level `eventbus.go` bridges to `digital.vasic.eventbus`.

## Factory Pattern

Several adapters use build-tag-selected factories for conditional compilation:

- `memory/factory_helixmemory.go` -- Active when HelixMemory submodule is present
- `memory/factory_standard.go` -- Fallback to the standard Memory module
- `specifier/factory_helixspecifier.go` -- Active when HelixSpecifier is present
- `specifier/factory_standard.go` -- Fallback (no-op)

## Performance Optimizations

Several adapters implement performance optimizations to reduce startup time
and prevent resource exhaustion:

### Lazy Loading
- **Database adapter**: Uses `sync.Once` for deferred PostgreSQL connection establishment
- **Containers adapter**: Runtime detection deferred until first use with `sync.Once`
- Connection errors are handled at operation time, not startup

### Concurrency Control  
- **Containers adapter**: Weighted semaphore limits concurrent container operations
- Semaphore weight: `2 * CPU cores` (capped between 2 and 10)
- Prevents system overload from too many simultaneous `docker compose` commands

### Thread Safety
- All lazy initialization is thread-safe via `sync.Once`
- Concurrent access to adapters is safe
- Race conditions prevented with proper synchronization

## Testing

Each adapter subdirectory contains `*_test.go` files. Total: 30+ test files
covering adapter initialization, type conversion, and integration behavior.

Run all adapter tests:

```bash
go test ./internal/adapters/...
```

Run security-focused adapter tests:

```bash
go test -tags security ./internal/adapters/...
```
