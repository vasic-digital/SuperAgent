# Dead Code Analysis Report

**Date:** February 27, 2026  
**Tool:** golang.org/x/tools/cmd/deadcode  
**Scope:** ./internal/...  

---

## Summary

**Total Unreachable Functions:** 50+
**Primary Categories:**
- Adapter unused methods (40+)
- Auth adapters (12)
- Database adapters (15)
- Cache adapters (1)
- Container adapters (2)
- Formatter adapters (3)
- MCP adapters (10)
- Messaging adapters (5)

---

## Critical Findings

### 1. Auth Adapter (internal/adapters/auth/)

**Status:** ⚠️ HIGH - 12 unreachable functions

These adapter methods were created for future use but never integrated:

- `NewFileCredentialReader` - Line 66
- `NewHTTPTokenRefresher` - Line 71
- `NewAutoRefresher` - Line 76
- `BearerTokenMiddleware` - Line 111
- `APIKeyHeaderMiddleware` - Line 116
- `RequireScopesMiddleware` - Line 121
- `NewCredentialReaderAdapter` - Line 171
- `CredentialReaderAdapter.ReadCredentials` - Line 176
- `CredentialReaderAdapter.GetAccessToken` - Line 181
- `CredentialReaderAdapter.HasValidCredentials` - Line 190
- `CredentialReaderAdapter.GetCredentialInfo` - Line 199

**Recommendation:** Either integrate these into production code or remove if truly unused.

### 2. Database Adapter (internal/adapters/database/)

**Status:** ⚠️ MEDIUM - 15 unreachable functions

**adapter.go:**
- `NewClientWithFallback` - Line 57
- `Client.Pool` - Line 116
- `Client.Database` - Line 121
- `Client.Close` - Line 126
- `Client.HealthCheck` - Line 136
- `Client.Exec` - Line 144
- `Client.Query` - Line 151
- `Client.QueryRow` - Line 169
- `Client.Begin` - Line 174
- `Client.Migrate` - Line 179

**compat.go:**
- `Connect` - Line 69
- `PostgresDB.Exec` - Line 84
- `PostgresDB.Query` - Line 89
- `PostgresDB.QueryRow` - Line 94
- `PostgresDB.Close` - Line 99
- `PostgresDB.HealthCheck` - Line 104
- `PostgresDB.GetPool` - Line 110
- `PostgresDB.Database` - Line 115
- `RunMigration` - Line 229

**Recommendation:** These appear to be legacy compatibility functions. Determine if migration is complete, then remove.

### 3. Cache Adapter (internal/adapters/cache/)

**Status:** ℹ️ LOW - 1 function

- `NewRedisClientAdapter` - Line 28

**Recommendation:** Unused factory function, can be removed.

### 4. Container Adapter (internal/adapters/containers/)

**Status:** ⚠️ MEDIUM - 2 functions

- `NewAdapterFromConfig` - Line 117
- `Adapter.setupDistribution` - Line 191

**Recommendation:** Verify if configuration-based adapter creation is needed.

### 5. Formatter Adapter (internal/adapters/formatters/)

**Status:** ℹ️ LOW - 3 functions

- `NativeFormatterFactory.CreateServiceFormatter` - Line 295
- `NewGenericRegistry` - Line 325
- `GetDefaultGenericRegistry` - Line 330

**Recommendation:** Unused registry and factory methods.

### 6. MCP Adapter (internal/adapters/mcp/)

**Status:** ⚠️ MEDIUM - 10 functions

- `NewClientAdapter` - Line 53
- `ClientAdapter.Initialize` - Line 77
- `ClientAdapter.ListTools` - Line 82
- `ClientAdapter.CallTool` - Line 87
- `ClientAdapter.ListResources` - Line 96
- `ClientAdapter.ReadResource` - Line 101
- `ClientAdapter.ListPrompts` - Line 109
- `ClientAdapter.GetPrompt` - Line 114
- `ClientAdapter.Close` - Line 123

**Recommendation:** These adapter methods appear incomplete or replaced by direct implementation.

### 7. Memory Adapter (internal/adapters/memory/)

**Status:** ℹ️ LOW - 1 function

- `NewHelixMemoryProvider` - Line 68

**Recommendation:** Unused factory, verify if needed.

### 8. Messaging Adapter (internal/adapters/messaging/)

**Status:** ⚠️ MEDIUM - 5 functions

- `BrokerAdapter.HealthCheck` - Line 50
- `BrokerAdapter.PublishBatch` - Line 74
- `NewConsumerGroupAdapter` - Line 20
- `ConsumerGroupAdapter.ID` - Line 28

**Recommendation:** These methods may be needed for full messaging implementation.

---

## Action Plan

### Phase 1: Immediate Cleanup (Low Risk)

**Files to Clean:**
1. `internal/adapters/cache/adapter.go` - Remove `NewRedisClientAdapter`
2. `internal/adapters/memory/factory_helixmemory.go` - Remove `NewHelixMemoryProvider`
3. `internal/adapters/formatters/adapter.go` - Remove unused factory methods

### Phase 2: Legacy Code Removal (Medium Risk)

**Files to Review:**
1. `internal/adapters/database/compat.go` - Likely complete migration, remove
2. `internal/adapters/database/adapter.go` - Verify unused methods

### Phase 3: Adapter Integration (High Risk)

**Files to Integrate or Remove:**
1. `internal/adapters/auth/adapter.go` - Integrate auth methods or remove
2. `internal/adapters/mcp/mcp.go` - Complete or remove MCP adapter
3. `internal/adapters/messaging/adapter.go` - Complete messaging implementation

### Phase 4: Verification

1. Run full test suite after each phase
2. Verify no breaking changes
3. Update documentation
4. Add tests for used functions

---

## Commands to Run

```bash
# Full deadcode analysis
deadcode -test ./...

# Generate detailed report
deadcode -test -format=json ./... > deadcode-report.json

# Check specific packages
deadcode -test ./internal/adapters/...
deadcode -test ./internal/llm/...
deadcode -test ./internal/debate/...
```

---

## Next Steps

1. Review each function category with team
2. Prioritize integration vs removal
3. Create migration plan for legacy code
4. Schedule cleanup sprints
5. Add deadcode check to CI pipeline

---

**Report Generated:** February 27, 2026  
**Total Issues:** 50+  
**Priority:** Medium
