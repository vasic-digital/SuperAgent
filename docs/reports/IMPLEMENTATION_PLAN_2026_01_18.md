# HelixAgent Implementation Plan
**Date**: 2026-01-18
**Based On**: Comprehensive Audit Report 2026-01-18
**Status**: Ready for Execution

---

## Overview

This plan addresses all issues identified in the comprehensive audit. Work is organized into phases with clear priorities, effort estimates, and verification requirements.

**Total Estimated Effort**: 25-30 developer days
**Timeline**: 3-4 sprints

---

## Phase 1: Critical Fixes (Immediate - 5-7 days)

### Task 1.1: RabbitMQ Test Suite
**Priority**: HIGH | **Effort**: 3-4 days | **Issue**: H1

**Objective**: Increase RabbitMQ coverage from 9.0% to 80%

**Implementation**:
```bash
# File to create/modify
internal/messaging/rabbitmq/broker_test.go
```

**Tests to Add (50+ tests)**:
1. Connection tests (10):
   - TestConnect_Success
   - TestConnect_InvalidURL
   - TestConnect_Timeout
   - TestConnect_Retry
   - TestConnect_TLS
   - TestClose_Success
   - TestClose_AlreadyClosed
   - TestReconnect_Success
   - TestReconnect_MaxAttempts
   - TestHealthCheck_All

2. Publishing tests (15):
   - TestPublish_Success
   - TestPublish_ToNonExistentExchange
   - TestPublish_WithConfirmation
   - TestPublish_Timeout
   - TestPublish_LargeMessage
   - TestPublish_WithHeaders
   - TestPublish_Batch
   - TestPublish_AfterReconnect
   - TestPublish_Concurrent
   - TestPublish_RateLimited
   - TestPublish_ToQueue
   - TestPublish_Persistent
   - TestPublish_Mandatory
   - TestPublish_Immediate
   - TestPublish_Metrics

3. Subscription tests (15):
   - TestSubscribe_Success
   - TestSubscribe_MultipleQueues
   - TestSubscribe_WithAck
   - TestSubscribe_WithNack
   - TestSubscribe_WithRequeue
   - TestSubscribe_Prefetch
   - TestSubscribe_ConsumerTag
   - TestUnsubscribe_Success
   - TestUnsubscribe_NonExistent
   - TestSubscribe_AfterReconnect
   - TestSubscribe_Concurrent
   - TestSubscribe_Handler_Panic
   - TestSubscribe_Handler_Error
   - TestSubscribe_DeadLetter
   - TestSubscribe_Priority

4. Exchange/Queue tests (10):
   - TestDeclareExchange_Direct
   - TestDeclareExchange_Fanout
   - TestDeclareExchange_Topic
   - TestDeclareQueue_Success
   - TestDeclareQueue_WithDLX
   - TestBindQueue_Success
   - TestBindQueue_Pattern
   - TestPurgeQueue_Success
   - TestDeleteQueue_Success
   - TestGetMetrics_All

**Verification**:
```bash
go test -v -cover ./internal/messaging/rabbitmq/...
# Expected: coverage > 80%
```

---

### Task 1.2: Kafka Test Suite
**Priority**: HIGH | **Effort**: 3-4 days | **Issue**: H2

**Objective**: Increase Kafka coverage from 11.6% to 80%

**Implementation**:
```bash
# File to create/modify
internal/messaging/kafka/broker_test.go
```

**Tests to Add (40+ tests)**:
1. Connection tests (8):
   - TestConnect_Success
   - TestConnect_MultipleBootstrap
   - TestConnect_InvalidBroker
   - TestConnect_SASL
   - TestConnect_TLS
   - TestClose_Success
   - TestReconnect_BrokerDown
   - TestHealthCheck_All

2. Producer tests (15):
   - TestPublish_Success
   - TestPublish_ToNonExistentTopic
   - TestPublish_WithKey
   - TestPublish_WithPartition
   - TestPublish_Batch
   - TestPublish_Async
   - TestPublish_WithHeaders
   - TestPublish_LargeMessage
   - TestPublish_Compression
   - TestPublish_Idempotent
   - TestPublish_Transactional
   - TestPublish_Timeout
   - TestPublish_Retry
   - TestPublish_Concurrent
   - TestPublish_Metrics

3. Consumer tests (12):
   - TestSubscribe_Success
   - TestSubscribe_MultipleTopics
   - TestSubscribe_ConsumerGroup
   - TestSubscribe_Rebalance
   - TestSubscribe_Offset_Earliest
   - TestSubscribe_Offset_Latest
   - TestSubscribe_CommitSync
   - TestSubscribe_CommitAsync
   - TestSubscribe_MaxPoll
   - TestSubscribe_Pause_Resume
   - TestUnsubscribe_Success
   - TestSubscribe_Handler_Error

4. Admin tests (5):
   - TestCreateTopic_Success
   - TestDeleteTopic_Success
   - TestListTopics_Success
   - TestDescribeGroup_Success
   - TestGetMetadata_Success

**Verification**:
```bash
go test -v -cover ./internal/messaging/kafka/...
# Expected: coverage > 80%
```

---

### Task 1.3: Move MockResourceMonitor
**Priority**: HIGH | **Effort**: 1 hour | **Issue**: H3

**Objective**: Separate mock from production code

**Current Location**:
```
internal/background/resource_monitor.go (lines 336-415)
```

**Target Location**:
```
internal/background/resource_monitor_test_helpers.go
```

**Implementation**:
1. Create new file `resource_monitor_test_helpers.go`
2. Move `MockResourceMonitor` struct and methods
3. Add build tag `// +build testing` if needed for isolation
4. Update any imports in test files
5. Verify no production code imports the mock

**Verification**:
```bash
# Ensure no production imports
grep -r "MockResourceMonitor" internal/background/*.go --include="*.go" ! --include="*_test*"
go build ./...
go test ./internal/background/...
```

---

### Task 1.4: Create debate_logs Migration
**Priority**: HIGH | **Effort**: 2 hours | **Issue**: H5

**Objective**: Move debate_logs schema from code to migration

**Implementation**:
Create file: `internal/database/migrations/014_debate_logs.sql`

```sql
-- Migration: 014_debate_logs
-- Description: Create debate_logs table for AI debate history
-- Date: 2026-01-18

CREATE TABLE IF NOT EXISTS debate_logs (
    id SERIAL PRIMARY KEY,
    debate_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    participant_id INTEGER,
    participant_identifier VARCHAR(255),
    participant_name VARCHAR(255),
    role VARCHAR(100),
    provider VARCHAR(100),
    model VARCHAR(255),
    round INTEGER,
    action VARCHAR(100),
    response_time_ms BIGINT,
    quality_score DECIMAL(5, 4),
    tokens_used INTEGER,
    content_length INTEGER,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for common queries
CREATE INDEX idx_debate_logs_debate_id ON debate_logs(debate_id);
CREATE INDEX idx_debate_logs_session_id ON debate_logs(session_id);
CREATE INDEX idx_debate_logs_provider ON debate_logs(provider);
CREATE INDEX idx_debate_logs_created_at ON debate_logs(created_at);
CREATE INDEX idx_debate_logs_expires_at ON debate_logs(expires_at);

-- Partial index for active debates
CREATE INDEX idx_debate_logs_active ON debate_logs(debate_id)
    WHERE expires_at IS NULL OR expires_at > NOW();
```

**Verification**:
```bash
# Apply migration
psql -h localhost -U helixagent -d helixagent_db -f internal/database/migrations/014_debate_logs.sql

# Verify table exists
psql -h localhost -U helixagent -d helixagent_db -c "\d debate_logs"
```

---

### Task 1.5: Consolidate Migration Directories
**Priority**: HIGH | **Effort**: 4 hours | **Issue**: H6

**Objective**: All migrations in one directory with proper numbering

**Current State**:
```
scripts/migrations/
├── 002_modelsdev_integration.sql
└── 003_protocol_support.sql

internal/database/migrations/
├── 011_background_tasks.sql
├── 012_performance_indexes.sql
└── 013_materialized_views.sql
```

**Target State**:
```
internal/database/migrations/
├── 001_init_schema.sql          # From init-db.sql
├── 002_modelsdev_integration.sql
├── 003_protocol_support.sql
├── 004_background_tasks.sql     # Renumbered from 011
├── 005_performance_indexes.sql  # Renumbered from 012
├── 006_materialized_views.sql   # Renumbered from 013
└── 007_debate_logs.sql          # New (was 014)
```

**Implementation Steps**:
1. Copy `scripts/migrations/*.sql` to `internal/database/migrations/`
2. Create `001_init_schema.sql` from `scripts/init-db.sql`
3. Renumber existing migrations
4. Update any migration tracking table
5. Update Makefile and scripts to use single directory
6. Remove `scripts/migrations/` directory
7. Update documentation

**Verification**:
```bash
# Test migration sequence
ls -la internal/database/migrations/
# Apply migrations in order
for f in internal/database/migrations/*.sql; do
    echo "Applying $f..."
    psql -h localhost -U helixagent -d helixagent_test < "$f"
done
```

---

## Phase 2: Medium Priority (This Sprint - 8-10 days)

### Task 2.1: Qdrant Test Suite
**Priority**: MEDIUM | **Effort**: 2 days | **Issue**: M1

**File**: `internal/vectordb/qdrant/client_test.go`

**Tests to Add (25+)**:
- Collection CRUD operations
- Vector upsert/search/delete
- Filter queries
- Batch operations
- Error handling
- Connection resilience

---

### Task 2.2: MinIO Test Suite
**Priority**: MEDIUM | **Effort**: 2 days | **Issue**: M2

**File**: `internal/storage/minio/client_test.go`

**Tests to Add (30+)**:
- Bucket operations
- Object upload/download
- Presigned URLs
- Lifecycle management
- Error handling
- Concurrent operations

---

### Task 2.3: Flink Test Suite
**Priority**: MEDIUM | **Effort**: 2 days | **Issue**: M3

**File**: `internal/streaming/flink/client_test.go`

**Tests to Add (25+)**:
- Job submission
- Job monitoring
- Savepoint operations
- Error handling
- Metrics retrieval

---

### Task 2.4: Create VectorDocumentRepository
**Priority**: MEDIUM | **Effort**: 4 hours | **Issue**: M4

**File**: `internal/database/vector_document_repository.go`

**Interface**:
```go
type VectorDocumentRepository interface {
    Create(ctx context.Context, doc *VectorDocument) error
    GetByID(ctx context.Context, id uuid.UUID) (*VectorDocument, error)
    Search(ctx context.Context, embedding []float64, limit int) ([]VectorDocument, error)
    Update(ctx context.Context, doc *VectorDocument) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListByProvider(ctx context.Context, provider string) ([]VectorDocument, error)
}
```

---

### Task 2.5: Create WebhookDeliveryRepository
**Priority**: MEDIUM | **Effort**: 4 hours | **Issue**: M5

**File**: `internal/database/webhook_delivery_repository.go`

**Interface**:
```go
type WebhookDeliveryRepository interface {
    Create(ctx context.Context, delivery *WebhookDelivery) error
    GetByID(ctx context.Context, id uuid.UUID) (*WebhookDelivery, error)
    GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]WebhookDelivery, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string, attempts int) error
    GetPendingDeliveries(ctx context.Context, limit int) ([]WebhookDelivery, error)
    MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
}
```

---

### Task 2.6: Add ModelBenchmark CRUD
**Priority**: MEDIUM | **Effort**: 2 hours | **Issue**: M6

**File**: `internal/database/model_metadata_repository.go`

**Methods to Add**:
```go
func (r *ModelMetadataRepository) UpdateBenchmark(ctx context.Context, id int, update *BenchmarkUpdate) error
func (r *ModelMetadataRepository) DeleteBenchmark(ctx context.Context, id int) error
func (r *ModelMetadataRepository) DeleteBenchmarksByModel(ctx context.Context, modelID string) error
```

---

### Task 2.7: Document HTTPACPTransport Design
**Priority**: MEDIUM | **Effort**: 1 hour | **Issue**: M8

**File**: `internal/services/protocol_discovery.go`

**Add Documentation**:
```go
// HTTPACPTransport implements ACPTransport using HTTP request-response pattern.
// DESIGN NOTE: HTTP is inherently request-response, so Receive() returns an error.
// For bidirectional communication, use WebSocketACPTransport instead.
// Alternative implementations could use:
// - Long polling (poll for messages periodically)
// - Server-Sent Events (receive-only stream)
// - HTTP/2 server push
type HTTPACPTransport struct {
    // ...
}

// Receive is not supported for HTTP transport as HTTP is request-response only.
// For bidirectional communication, use WebSocketACPTransport.
// Returns: error always, as HTTP cannot receive push messages.
func (t *HTTPACPTransport) Receive(ctx context.Context) (interface{}, error) {
    return nil, fmt.Errorf("HTTP transport does not support receive - use WebSocketACPTransport for bidirectional communication")
}
```

---

## Phase 3: Low Priority (Next Sprint - 3-4 days)

### Task 3.1: Renumber Migrations
**Priority**: LOW | **Effort**: 2 hours | **Issue**: L1

Update migration files to use sequential numbering without gaps.

---

### Task 3.2: Iceberg Test Suite
**Priority**: LOW | **Effort**: 2 days | **Issue**: L4

Add 20+ tests for Apache Iceberg catalog operations.

---

### Task 3.3: Replay Edge Case Tests
**Priority**: LOW | **Effort**: 1 day | **Issue**: L5

Add tests for edge cases in message replay functionality.

---

## Phase 4: Security & Monitoring (Ongoing)

### Task 4.1: Security Dependency Monitoring

**Weekly Checklist**:
- [ ] Check `golang.org/x/crypto` for new CVE patches
- [ ] Run `govulncheck ./...` for vulnerability scanning
- [ ] Review security advisories for direct dependencies

**Bi-Weekly Checklist**:
- [ ] Run `go mod tidy`
- [ ] Check for dependency updates with `go list -m -u all`
- [ ] Review and update indirect dependencies

**Monthly Checklist**:
- [ ] Full dependency audit
- [ ] Review and update outdated packages
- [ ] Security scan with gosec

### Task 4.2: Set Up Automated Security Scanning

**GitHub Actions Workflow**:
```yaml
name: Security Scan
on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday
  push:
    branches: [main]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
      - name: Run gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec ./...
```

---

## Verification Procedures

### For Each Task Completion

1. **Code Review Checklist**:
   - [ ] Tests added with meaningful assertions
   - [ ] Edge cases covered
   - [ ] Error paths tested
   - [ ] No hardcoded values
   - [ ] Proper error messages
   - [ ] Documentation updated

2. **CI/CD Verification**:
   ```bash
   # Run full test suite
   make test

   # Check coverage
   make test-coverage

   # Run linting
   make lint

   # Run security scan
   make security-scan
   ```

3. **Manual Verification**:
   - Start test infrastructure: `make test-infra-start`
   - Run integration tests: `make test-integration`
   - Check coverage report in browser

### Coverage Verification Script

Create `scripts/verify_coverage.sh`:
```bash
#!/bin/bash

# Target coverage thresholds
declare -A THRESHOLDS=(
    ["internal/messaging/rabbitmq"]=80
    ["internal/messaging/kafka"]=80
    ["internal/messaging/inmemory"]=80
    ["internal/vectordb/qdrant"]=80
    ["internal/storage/minio"]=80
    ["internal/streaming/flink"]=80
    ["internal/messaging/dlq"]=80
)

# Run coverage
go test -coverprofile=coverage.out ./...

# Check each package
for pkg in "${!THRESHOLDS[@]}"; do
    target=${THRESHOLDS[$pkg]}
    actual=$(go tool cover -func=coverage.out | grep "$pkg" | tail -1 | awk '{print $NF}' | tr -d '%')

    if (( $(echo "$actual < $target" | bc -l) )); then
        echo "FAIL: $pkg coverage $actual% < $target%"
        exit 1
    else
        echo "PASS: $pkg coverage $actual% >= $target%"
    fi
done

echo "All coverage thresholds met!"
```

---

## Progress Tracking

### Sprint 1 (Current)
| Task | Status | Assignee | Notes |
|------|--------|----------|-------|
| 1.1 RabbitMQ Tests | Not Started | | |
| 1.2 Kafka Tests | Not Started | | |
| 1.3 Move Mock | Not Started | | |
| 1.4 debate_logs Migration | Not Started | | |
| 1.5 Consolidate Migrations | Not Started | | |

### Sprint 2
| Task | Status | Assignee | Notes |
|------|--------|----------|-------|
| 2.1 Qdrant Tests | Not Started | | |
| 2.2 MinIO Tests | Not Started | | |
| 2.3 Flink Tests | Not Started | | |
| 2.4 VectorDocumentRepo | Not Started | | |
| 2.5 WebhookDeliveryRepo | Not Started | | |
| 2.6 Benchmark CRUD | Not Started | | |

### Sprint 3
| Task | Status | Assignee | Notes |
|------|--------|----------|-------|
| 3.1 Renumber Migrations | Not Started | | |
| 3.2 Iceberg Tests | Not Started | | |
| 3.3 Replay Tests | Not Started | | |

---

## Success Metrics

### Coverage Targets
| Package | Current | Target | Status |
|---------|---------|--------|--------|
| Overall | 59.4% | 80% | Pending |
| messaging/rabbitmq | 9.0% | 80% | Critical |
| messaging/kafka | 11.6% | 80% | Critical |
| vectordb/qdrant | 35.3% | 80% | Pending |
| storage/minio | 45.2% | 80% | Pending |

### Quality Metrics
- All tests passing: Required
- No security vulnerabilities: Required
- No TODO/FIXME in new code: Required
- Documentation updated: Required

---

*Plan generated by Claude Code (Opus 4.5) on 2026-01-18*
*Based on Comprehensive Audit Report 2026-01-18*
