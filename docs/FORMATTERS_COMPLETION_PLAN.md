# Code Formatters Integration - Completion Plan

**Project**: Complete Code Formatters Integration
**Status**: Phase 3 In Progress (API Endpoints)
**Date**: 2026-01-29

---

## Current Status

### ✅ Completed (Phases 1-2)

**Phase 1: Core Infrastructure** (100% Complete)
- ✅ `internal/formatters/` package (8 files, 2,100+ lines)
- ✅ FormatterRegistry with thread-safe operations
- ✅ FormatterExecutor with 6 middleware types
- ✅ FormatterCache with LRU and TTL
- ✅ Complete configuration system
- ✅ FormatterFactory for creating instances
- ✅ HealthChecker for all formatters
- ✅ VersionsManifest for Git submodules
- ✅ Unit tests (8 tests, 100% pass rate)

**Phase 2: Git Submodules Infrastructure** (100% Complete)
- ✅ `formatters/` directory structure
- ✅ `formatters/VERSIONS.yaml` manifest (118 formatters)
- ✅ `formatters/README.md` (comprehensive documentation)
- ✅ `formatters/scripts/init-submodules.sh` (initialization script)
- ✅ `formatters/scripts/build-all.sh` (build script for native binaries)
- ✅ `formatters/scripts/health-check-all.sh` (health check script)

**Phase 3: API Endpoints** (90% Complete)
- ✅ `internal/handlers/formatters_handler.go` (850+ lines)
- ✅ 8 REST endpoints implemented:
  - `POST /v1/format` - Format code
  - `POST /v1/format/batch` - Batch formatting
  - `POST /v1/format/check` - Check if formatted
  - `GET /v1/formatters` - List formatters
  - `GET /v1/formatters/detect` - Detect formatter
  - `GET /v1/formatters/:name` - Get metadata
  - `GET /v1/formatters/:name/health` - Health check
  - `POST /v1/formatters/:name/validate-config` - Validate config
- ⏳ Router integration (pending)
- ⏳ Integration tests (pending)

---

## Remaining Work

### Phase 4: Native Formatter Providers (Highest Priority)

**Scope**: Implement 60+ native binary formatter providers

**Implementation Strategy**:
1. Create base native formatter implementation
2. Implement top 20 most popular formatters first
3. Test each formatter with real binaries
4. Create mock implementations for remaining formatters

**Top 20 Formatters to Implement**:
1. `black` (Python)
2. `ruff` (Python - 30x faster)
3. `prettier` (JS/TS/HTML/CSS/etc.)
4. `biome` (JS/TS - 35x faster)
5. `rustfmt` (Rust)
6. `gofmt` (Go)
7. `clang-format` (C/C++)
8. `google-java-format` (Java)
9. `ktlint` (Kotlin)
10. `scalafmt` (Scala)
11. `swift-format` (Swift)
12. `shfmt` (Bash)
13. `yamlfmt` (YAML)
14. `taplo` (TOML)
15. `buf` (Protobuf)
16. `stylua` (Lua)
17. `ormolu` (Haskell)
18. `ocamlformat` (OCaml)
19. `dprint` (Pluggable)
20. `terraform fmt` (Terraform)

**Files to Create**:
- `internal/formatters/providers/native/base.go` - Base implementation
- `internal/formatters/providers/native/black.go`
- `internal/formatters/providers/native/ruff.go`
- `internal/formatters/providers/native/prettier.go`
- ... (20 files for top formatters)
- `internal/formatters/providers/native/*_test.go` - Unit tests

**Estimated Time**: 2 days

---

### Phase 5: Service Formatter Providers

**Scope**: Implement 20+ service-based formatter providers

**Implementation Strategy**:
1. Create base HTTP service formatter
2. Implement top 10 service formatters
3. Create Docker containers for each
4. Test service communication

**Top 10 Service Formatters**:
1. `sqlfluff` (SQL multi-dialect) - port 9201
2. `rubocop` (Ruby) - port 9202
3. `spotless` (JVM multi-language) - port 9203
4. `php-cs-fixer` (PHP) - port 9204
5. `npm-groovy-lint` (Groovy) - port 9205
6. `scalafmt` (Scala service mode) - port 9206
7. `styler` (R) - port 9207
8. `psscriptanalyzer` (PowerShell) - port 9208
9. `autopep8` (Python) - port 9209
10. `yapf` (Python) - port 9210

**Files to Create**:
- `internal/formatters/providers/service/base.go` - Base HTTP client
- `internal/formatters/providers/service/sqlfluff.go`
- `internal/formatters/providers/service/rubocop.go`
- ... (10 files)
- `docker/formatters/Dockerfile.sqlfluff`
- `docker/formatters/Dockerfile.rubocop`
- ... (10 Dockerfiles)
- `docker/formatters/services/*.py` - Service wrappers
- `docker/formatters/docker-compose.formatters.yml`

**Estimated Time**: 3 days

---

### Phase 6: Router Integration & Testing

**Scope**: Wire formatters handler into router, create integration tests

**Tasks**:
1. Add formatters handler to `internal/router/router.go`
2. Initialize formatters registry in router setup
3. Register top 20-30 formatters on startup
4. Create integration tests
5. Test all 8 API endpoints

**Files to Modify**:
- `internal/router/router.go` - Add formatters handler
- `cmd/helixagent/main.go` - Initialize formatters system

**Files to Create**:
- `tests/integration/formatters_api_test.go` - API integration tests
- `tests/integration/formatters_e2e_test.go` - End-to-end tests

**Estimated Time**: 1 day

---

### Phase 7: CLI Agent Integration

**Scope**: Integrate formatters with all 48 CLI agents

**Tasks**:
1. Update CLI agent config generators
2. Add formatter preferences to each agent
3. Implement `format` command for agents
4. Test with OpenCode and Crush first
5. Roll out to remaining 46 agents

**Files to Modify**:
- `internal/cliagents/config_generators/*.go` - Add formatter config
- CLI agent templates

**Files to Create**:
- CLI command implementations (if needed)

**Estimated Time**: 2 days

---

### Phase 8: AI Debate Integration

**Scope**: Auto-format code in AI Debate responses

**Tasks**:
1. Create `internal/services/debate_formatter_integration.go`
2. Implement code block extraction
3. Implement auto-formatting logic
4. Add event streaming for format progress
5. Integrate with existing debate service

**Files to Create**:
- `internal/services/debate_formatter_integration.go` (400+ lines)
- `internal/services/debate_formatter_integration_test.go`

**Files to Modify**:
- `internal/services/debate_service.go` - Integrate formatter
- `internal/handlers/debate_handler.go` - Add format events

**Estimated Time**: 2 days

---

### Phase 9: Comprehensive Testing

**Scope**: 100% test coverage with all test types

**Test Categories**:
1. Unit tests (200+ tests)
   - Provider implementations
   - Registry operations
   - Executor middleware
   - Cache behavior
   - Config validation

2. Integration tests (50+ tests)
   - API endpoints
   - Service communication
   - CLI agent integration
   - AI Debate integration

3. Challenge scripts (6 scripts)
   - `formatters_native_challenge.sh` (60 tests)
   - `formatters_service_challenge.sh` (20 tests)
   - `formatters_builtin_challenge.sh` (15 tests)
   - `formatters_unified_challenge.sh` (10 tests)
   - `formatters_performance_challenge.sh` (20 tests)
   - `formatters_integration_challenge.sh` (30 tests)

4. Final challenge
   - `cli_agents_formatters_challenge.sh` (5,782 tests)
     - 118 formatters × 48 agents = 5,664 tests
     - 118 formatters × AI Debate = 118 tests

**Estimated Time**: 3 days

---

### Phase 10: Documentation

**Scope**: Complete documentation update

**Documents to Create**:
1. `docs/api/FORMATTERS_API.md` - API reference
2. `docs/user/FORMATTERS_GUIDE.md` - User guide
3. `docs/architecture/FORMATTERS_INTEGRATION.md` - Integration guide
4. `docs/deployment/FORMATTERS_DEPLOYMENT.md` - Deployment guide

**Documents to Update**:
1. `CLAUDE.md` - Add formatters section
2. `docs/architecture/ARCHITECTURE.md` - Add formatters
3. `docs/user/QUICKSTART.md` - Add formatting examples
4. `docs/deployment/DEPLOYMENT_GUIDE.md` - Add formatter deployment
5. SQL schemas (if applicable)
6. System diagrams (Mermaid/PlantUML)

**Estimated Time**: 2 days

---

## Timeline Summary

| Phase | Status | Estimated Time | Actual Time |
|-------|--------|----------------|-------------|
| 1. Core Infrastructure | ✅ Complete | 1 day | 4 hours |
| 2. Git Submodules | ✅ Complete | 1 day | 2 hours |
| 3. API Endpoints | 90% Complete | 1 day | 2 hours |
| 4. Native Providers | ⏳ Pending | 2 days | - |
| 5. Service Providers | ⏳ Pending | 3 days | - |
| 6. Router & Tests | ⏳ Pending | 1 day | - |
| 7. CLI Agent Integration | ⏳ Pending | 2 days | - |
| 8. AI Debate Integration | ⏳ Pending | 2 days | - |
| 9. Comprehensive Testing | ⏳ Pending | 3 days | - |
| 10. Documentation | ⏳ Pending | 2 days | - |
| **Total** | **30% Complete** | **18 days** | **8 hours** |

**Remaining Time**: 17 days
**Estimated Completion**: 2026-02-15

---

## Immediate Next Steps (Priority Order)

1. **Implement Native Formatter Providers** (Highest Priority)
   - Create `internal/formatters/providers/native/base.go`
   - Implement top 20 formatters
   - Create unit tests

2. **Complete Router Integration**
   - Wire formatters handler into router
   - Initialize registry on startup
   - Register formatters

3. **Implement Service Formatter Providers**
   - Create base HTTP service formatter
   - Implement top 10 service formatters
   - Create Docker containers

4. **CLI Agent Integration**
   - Update config generators
   - Add formatter commands

5. **AI Debate Integration**
   - Auto-format code blocks

6. **Testing & Validation**
   - Write all tests
   - Run challenge scripts
   - Achieve 100% pass rate

7. **Documentation**
   - Complete all docs
   - Update diagrams

---

## Success Criteria

- [ ] All 118 formatters integrated (native/service/builtin)
- [ ] 8 REST API endpoints working
- [ ] All 48 CLI agents have formatter support
- [ ] AI Debate system auto-formats code
- [ ] 100% test coverage
- [ ] All challenge scripts pass (5,782 tests)
- [ ] Zero false positives
- [ ] Complete documentation

---

## Risk Mitigation

**Risk**: Not all formatters available on all systems
**Mitigation**: Graceful degradation - formatters that aren't installed return clear errors

**Risk**: Service containers may fail to start
**Mitigation**: Health checks, automatic restart policies, clear error messages

**Risk**: Performance issues with 118 formatters
**Mitigation**: Lazy loading, caching, parallel execution

**Risk**: Integration complexity with 48 CLI agents
**Mitigation**: Standardized config format, automated testing

---

**Last Updated**: 2026-01-29 14:00 EET
**Next Review**: 2026-01-30
**Completion Target**: 2026-02-15
