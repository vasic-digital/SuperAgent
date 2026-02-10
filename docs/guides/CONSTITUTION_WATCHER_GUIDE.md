# 📜 Constitution Watcher Operational Guide

## 📋 Overview

The Constitution Watcher is an intelligent background service that automatically monitors project changes and updates the Constitution (CLAUDE.md, AGENTS.md, CONSTITUTION.md) when significant structural or architectural changes occur.

**Key Benefits:**
- ✅ Keeps Constitution synchronized with codebase reality
- ✅ Detects new extracted modules automatically
- ✅ Monitors documentation consistency
- ✅ Tracks test coverage changes
- ✅ Alerts on structural changes requiring rule updates

---

## 🎯 What It Watches

### 1. New Module Extraction
**Detection**: Monitors `go.mod` for new module entries

**Triggers When:**
- New `replace` directive added (e.g., `replace digital.vasic.newmodule => ./NewModule`)
- New module directory created with own `go.mod`
- Module appears in dependency graph

**Action:**
- Updates "Extracted Modules" section in CLAUDE.md
- Adds module to MODULES.md catalog
- Updates AGENTS.md with new module agents
- Regenerates Constitution module count

**Example Detection:**
```bash
# Watcher detects:
replace digital.vasic.authentication => ./Authentication

# Updates Constitution:
"21 extracted modules" → "22 extracted modules"
```

---

### 2. Documentation Changes
**Detection**: Monitors AGENTS.md, CLAUDE.md, README.md for modifications

**Triggers When:**
- CLAUDE.md updated but AGENTS.md not synchronized
- New rules added to Constitution
- Feature documentation added without updating Constitution
- Mandatory principles changed

**Action:**
- Synchronizes CLAUDE.md ↔ AGENTS.md ↔ CONSTITUTION.md
- Updates "Last Updated" timestamps
- Validates Constitution compliance
- Generates synchronization report

**Example Detection:**
```bash
# Watcher detects:
CLAUDE.md modified: 2026-02-10 14:30
AGENTS.md modified: 2026-02-09 10:15

# Action:
Synchronization gap detected (1 day)
Regenerating AGENTS.md from CLAUDE.md...
```

---

### 3. Project Structure Changes
**Detection**: Monitors top-level directories for new additions

**Triggers When:**
- New top-level directory created (e.g., `NewFeature/`)
- Significant file count increase (>50 new files)
- New package added to `internal/`
- Docker compose file changes

**Action:**
- Reviews new structure against Constitution
- Checks for documentation requirements
- Validates testing requirements
- Updates architecture documentation

**Example Detection:**
```bash
# Watcher detects:
New directory: ./Workflow/
Contains: go.mod, README.md, 15 .go files

# Action:
New module detected: "Workflow"
Checking Constitution compliance...
  ✅ README.md present
  ✅ go.mod present
  ❌ CLAUDE.md missing
  ❌ AGENTS.md missing
Alert: Module documentation incomplete
```

---

### 4. Test Coverage Drops
**Detection**: Monitors test coverage reports

**Triggers When:**
- Coverage drops below 65% (current baseline)
- Critical paths lose test coverage
- New files added without tests
- Integration test count decreases

**Action:**
- Alerts development team
- Logs coverage regression
- Blocks deployment if critical
- Updates Constitution testing status

**Example Detection:**
```bash
# Watcher detects:
Coverage change: 65.6% → 58.3% (-7.3%)
Cause: 47 new files in BigData/ without tests

# Action:
Constitution violation: "100% Test Coverage"
Severity: HIGH
Creating GitHub issue: "Test coverage regression in BigData"
```

---

## ⚙️ Configuration

### Environment Variables

```bash
# Enable/disable Constitution Watcher (default: true)
CONSTITUTION_WATCHER_ENABLED=true

# Check interval in minutes (default: 5)
CONSTITUTION_WATCHER_CHECK_INTERVAL=5

# Triggers to monitor (comma-separated)
CONSTITUTION_WATCHER_TRIGGERS=modules,docs,structure,coverage

# Minimum coverage threshold (default: 65.0)
CONSTITUTION_WATCHER_MIN_COVERAGE=65.0

# Auto-sync documentation (default: true)
CONSTITUTION_WATCHER_AUTO_SYNC=true

# Alert on violations (default: true)
CONSTITUTION_WATCHER_ALERT_VIOLATIONS=true

# Notification channels (comma-separated: email,slack,webhook)
CONSTITUTION_WATCHER_NOTIFICATIONS=email,slack

# Log level (debug, info, warn, error)
CONSTITUTION_WATCHER_LOG_LEVEL=info
```

### Configuration File

Create `configs/constitution_watcher.yaml`:

```yaml
constitution_watcher:
  enabled: true
  check_interval_minutes: 5

  triggers:
    new_modules:
      enabled: true
      watch_paths:
        - go.mod
        - "*/go.mod"
      action: update_constitution

    documentation_changes:
      enabled: true
      watch_files:
        - CLAUDE.md
        - AGENTS.md
        - CONSTITUTION.md
        - README.md
      sync_threshold_hours: 24
      auto_sync: true

    structure_changes:
      enabled: true
      watch_paths:
        - "."
      ignore_paths:
        - node_modules
        - .git
        - bin
        - tmp
      min_file_threshold: 50

    test_coverage:
      enabled: true
      watch_files:
        - coverage.out
        - coverage_combined.out
      min_coverage: 65.0
      alert_threshold: 5.0  # Alert if drops by 5%

  documentation_sync:
    enabled: true
    sync_pairs:
      - source: CLAUDE.md
        targets:
          - AGENTS.md
          - CONSTITUTION.md
      - source: CONSTITUTION.md
        targets:
          - CLAUDE.md (Constitution section)
          - AGENTS.md (Constitution references)

  notifications:
    email:
      enabled: true
      smtp_host: smtp.gmail.com
      smtp_port: 587
      from: helixagent@company.com
      to:
        - dev-team@company.com
        - ops-team@company.com

    slack:
      enabled: true
      webhook_url: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
      channel: "#helixagent-alerts"
      mention_on_critical: "@channel"

    webhook:
      enabled: false
      url: https://api.company.com/webhooks/constitution
      headers:
        Authorization: Bearer YOUR_TOKEN

  logging:
    level: info
    file: logs/constitution_watcher.log
    max_size_mb: 100
    max_backups: 10
```

---

## 🚀 Usage

### Starting the Watcher

```bash
# Start with default configuration
helixagent watcher start

# Start with custom interval
helixagent watcher start --interval=10m

# Start with specific triggers only
helixagent watcher start --triggers=modules,coverage

# Start in dry-run mode (no changes, logging only)
helixagent watcher start --dry-run
```

### Checking Status

```bash
# Check watcher status
helixagent watcher status

# Response:
{
  "enabled": true,
  "running": true,
  "check_interval": "5m",
  "last_check": "2026-02-10T14:35:22Z",
  "next_check": "2026-02-10T14:40:22Z",
  "triggers_monitored": ["modules", "docs", "structure", "coverage"],
  "changes_detected_today": 3,
  "last_sync": "2026-02-10T12:15:00Z"
}
```

### Forcing a Check

```bash
# Trigger immediate check
helixagent watcher check

# Force documentation sync
helixagent watcher sync --force

# Check specific trigger
helixagent watcher check --trigger=coverage
```

### Viewing Watcher Logs

```bash
# Tail logs
tail -f logs/constitution_watcher.log

# Filter for alerts
grep "ALERT" logs/constitution_watcher.log

# View today's changes
grep "$(date +%Y-%m-%d)" logs/constitution_watcher.log
```

---

## 📊 Monitoring & Observability

### Health Endpoint

```bash
# Check watcher health
curl http://localhost:7061/v1/constitution/watcher/health

# Response:
{
  "status": "healthy",
  "last_check": "2026-02-10T14:35:22Z",
  "next_check": "2026-02-10T14:40:22Z",
  "uptime": "72h30m",
  "checks_completed": 876,
  "changes_detected": 23,
  "sync_operations": 12,
  "errors": 0
}
```

### Metrics

```bash
# Prometheus metrics
constitution_watcher_checks_total{trigger="modules"} 876
constitution_watcher_changes_detected{trigger="modules"} 3
constitution_watcher_sync_operations_total 12
constitution_watcher_check_duration_seconds 1.23
constitution_watcher_errors_total{type="sync_failure"} 0

# Coverage tracking
constitution_watcher_test_coverage_percent 65.6
constitution_watcher_coverage_change_percent -0.2
```

### Change History

```bash
# View change history
curl http://localhost:7061/v1/constitution/watcher/history

# Response:
{
  "changes": [
    {
      "timestamp": "2026-02-10T12:15:00Z",
      "trigger": "new_modules",
      "description": "New module detected: Authentication",
      "action_taken": "Updated CLAUDE.md, AGENTS.md, MODULES.md",
      "status": "completed"
    },
    {
      "timestamp": "2026-02-09T18:30:00Z",
      "trigger": "documentation_changes",
      "description": "CLAUDE.md updated, AGENTS.md out of sync",
      "action_taken": "Synchronized AGENTS.md with CLAUDE.md",
      "status": "completed"
    },
    {
      "timestamp": "2026-02-08T14:20:00Z",
      "trigger": "test_coverage",
      "description": "Coverage dropped to 58.3% (-7.3%)",
      "action_taken": "Alert sent to dev-team@company.com",
      "status": "alerted"
    }
  ]
}
```

---

## 🔔 Alert Examples

### New Module Alert

```
📦 New Module Detected

Module: Authentication
Location: ./Authentication
Detected: 2026-02-10 12:15:00

Constitution Update Required:
  ✅ README.md present
  ✅ go.mod present
  ❌ CLAUDE.md missing
  ❌ AGENTS.md missing
  ❌ tests/ directory missing

Action Required:
  1. Create CLAUDE.md in Authentication/
  2. Create AGENTS.md in Authentication/
  3. Add unit tests
  4. Update main CLAUDE.md with new module

Severity: MEDIUM
Auto-sync: PARTIAL (updated MODULES.md)
```

---

### Documentation Sync Alert

```
📄 Documentation Synchronization Issue

Files Out of Sync:
  - CLAUDE.md (modified: 2026-02-10 14:30)
  - AGENTS.md (modified: 2026-02-09 10:15)
  - Gap: 1 day 4 hours

Changes in CLAUDE.md:
  - New section: "SpecKit Auto-Activation"
  - Updated provider count: 10 → 21
  - Added Constitution Watcher documentation

Action Taken:
  ✅ Regenerated AGENTS.md from CLAUDE.md
  ✅ Updated timestamps
  ✅ Validated Constitution compliance

Severity: LOW
Auto-sync: COMPLETED
```

---

### Coverage Drop Alert

```
⚠️ Test Coverage Regression

Current Coverage: 58.3%
Previous Coverage: 65.6%
Change: -7.3% (Below threshold)

Cause Analysis:
  - 47 new files in BigData/ module
  - 0 test files added
  - Untested functions: 156

Constitution Violation:
  Rule: "100% Test Coverage"
  Severity: HIGH

Action Required:
  1. Add tests for BigData/ module
  2. Restore coverage to ≥65%
  3. Update Constitution with coverage status

Severity: HIGH
Auto-sync: N/A (requires manual fix)
```

---

## ❓ Troubleshooting

### Watcher Not Starting

**Problem**: Watcher fails to start

**Solutions**:
```bash
# Check configuration
helixagent watcher validate-config

# Enable debug logging
export CONSTITUTION_WATCHER_LOG_LEVEL=debug
helixagent watcher start

# Check for port conflicts
lsof -i :7061
```

---

### Sync Failures

**Problem**: Documentation sync fails

**Solutions**:
```bash
# Check file permissions
ls -la CLAUDE.md AGENTS.md

# Validate file syntax
helixagent watcher validate-files

# Manual sync with verbose output
helixagent watcher sync --verbose
```

---

### False Positives

**Problem**: Watcher detects changes that aren't real

**Solutions**:
```bash
# Increase threshold
export CONSTITUTION_WATCHER_MIN_FILE_THRESHOLD=100

# Ignore specific paths
# Add to configs/constitution_watcher.yaml:
ignore_paths:
  - tmp
  - vendor
  - .cache

# Clear cache
helixagent watcher clear-cache
```

---

## 🎓 Best Practices

### 1. Regular Monitoring
- Check watcher status daily
- Review change history weekly
- Investigate all alerts promptly
- Maintain alert configuration

### 2. Timely Synchronization
- Don't delay documentation sync >24 hours
- Address Constitution violations immediately
- Keep CLAUDE.md, AGENTS.md, CONSTITUTION.md in sync
- Review auto-sync changes for accuracy

### 3. Coverage Maintenance
- Never let coverage drop below baseline (65%)
- Add tests for all new modules immediately
- Run coverage checks before committing
- Monitor coverage trends

### 4. Module Documentation
- Document new modules within 24 hours
- Follow Constitution documentation requirements
- Use existing module docs as templates
- Validate against checklist

### 5. Alert Handling
- Treat HIGH severity as urgent
- Investigate MEDIUM within 1 day
- Log LOW for review
- Never disable alerts globally

---

## 📚 Related Documentation

- **[CLAUDE.md](../../CLAUDE.md)** - Development standards and Constitution
- **[AGENTS.md](../../AGENTS.md)** - Agent registry and configurations
- **[CONSTITUTION.md](../../CONSTITUTION.md)** - Project rules and mandatory principles
- **[MODULES.md](../MODULES.md)** - Extracted modules catalog

---

## 🔗 Quick Reference

### Commands

```bash
# Start watcher
helixagent watcher start

# Check status
helixagent watcher status

# Force check
helixagent watcher check

# Force sync
helixagent watcher sync --force

# View history
helixagent watcher history

# Validate config
helixagent watcher validate-config
```

### Key Files

- Configuration: `configs/constitution_watcher.yaml`
- Logs: `logs/constitution_watcher.log`
- Cache: `.watcher/cache/`
- History: `.watcher/history/`

### API Endpoints

- Health: `GET /v1/constitution/watcher/health`
- Status: `GET /v1/constitution/watcher/status`
- History: `GET /v1/constitution/watcher/history`
- Force Check: `POST /v1/constitution/watcher/check`
- Force Sync: `POST /v1/constitution/watcher/sync`

---

**Last Updated**: February 10, 2026
**Version**: 1.0.0
**Status**: ✅ Production Ready
