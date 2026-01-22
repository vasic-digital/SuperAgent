#!/bin/bash
# MCP Comprehensive Challenge
# Validates ALL 37 MCP implementations (20 adapters + 17 servers)
# Tests: Implementation, Tests, Interface compliance, Tool definitions

set -e

# Source challenge framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

# Initialize challenge
init_challenge "mcp_comprehensive" "MCP Comprehensive Verification"
load_env

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "test" "$test_name" "true" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "test" "$test_name" "false" "Test command failed"
        return 1
    fi
}

# ============================================================================
# SECTION 1: MCP ADAPTERS (20 adapters)
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: MCP Adapters (20 implementations)"
log_info "=============================================="

# Complete list of MCP adapters
MCP_ADAPTERS=(
    "asana"
    "aws_s3"
    "brave_search"
    "datadog"
    "docker"
    "figma"
    "gitlab"
    "google_drive"
    "jira"
    "kubernetes"
    "linear"
    "miro"
    "mongodb"
    "notion"
    "puppeteer"
    "registry"
    "sentry"
    "slack"
    "stable_diffusion"
    "svgmaker"
)

log_info "Verifying MCP Adapter implementations..."
for adapter in "${MCP_ADAPTERS[@]}"; do
    # Check implementation exists
    run_test "MCP Adapter: $adapter implementation exists" \
        "[[ -f '$PROJECT_ROOT/internal/mcp/adapters/${adapter}.go' ]]"

    # Check tests exist
    run_test "MCP Adapter: $adapter has tests" \
        "[[ -f '$PROJECT_ROOT/internal/mcp/adapters/${adapter}_test.go' ]]"
done

# Verify adapter interface compliance
log_info "Verifying adapter interface compliance..."
run_test "MCPAdapter interface defined" \
    "grep -q 'type MCPAdapter interface' '$PROJECT_ROOT/internal/mcp/adapters/registry.go'"

run_test "GetServerInfo method required" \
    "grep -q 'GetServerInfo()' '$PROJECT_ROOT/internal/mcp/adapters/registry.go'"

run_test "ListTools method required" \
    "grep -q 'ListTools()' '$PROJECT_ROOT/internal/mcp/adapters/registry.go'"

run_test "CallTool method required" \
    "grep -q 'CallTool(' '$PROJECT_ROOT/internal/mcp/adapters/registry.go'"

# ============================================================================
# SECTION 2: MCP SERVERS (17 servers)
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: MCP Servers (17 implementations)"
log_info "=============================================="

# Complete list of MCP servers
MCP_SERVERS=(
    "chroma_adapter"
    "fetch_adapter"
    "figma_adapter"
    "filesystem_adapter"
    "git_adapter"
    "github_adapter"
    "memory_adapter"
    "miro_adapter"
    "postgres_adapter"
    "qdrant_adapter"
    "redis_adapter"
    "replicate_adapter"
    "sqlite_adapter"
    "stablediffusion_adapter"
    "svgmaker_adapter"
    "unified_manager"
    "weaviate_adapter"
)

log_info "Verifying MCP Server implementations..."
for server in "${MCP_SERVERS[@]}"; do
    run_test "MCP Server: $server implementation exists" \
        "[[ -f '$PROJECT_ROOT/internal/mcp/servers/${server}.go' ]]"

    # Check tests exist (skip unified_manager which may have different test structure)
    if [[ "$server" != "unified_manager" ]]; then
        run_test "MCP Server: $server has tests" \
            "[[ -f '$PROJECT_ROOT/internal/mcp/servers/${server}_test.go' ]]"
    fi
done

# ============================================================================
# SECTION 3: MCP TOOL DEFINITIONS
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: MCP Tool Definitions"
log_info "=============================================="

# Verify each adapter defines tools
log_info "Verifying tool definitions in adapters..."

# Asana tools
run_test "Asana: list_tasks tool" \
    "grep -q 'list_tasks' '$PROJECT_ROOT/internal/mcp/adapters/asana.go'"
run_test "Asana: create_task tool" \
    "grep -q 'create_task' '$PROJECT_ROOT/internal/mcp/adapters/asana.go'"

# AWS S3 tools
run_test "AWS S3: list_buckets tool" \
    "grep -q 'list_buckets' '$PROJECT_ROOT/internal/mcp/adapters/aws_s3.go'"
run_test "AWS S3: get_object tool" \
    "grep -q 'get_object' '$PROJECT_ROOT/internal/mcp/adapters/aws_s3.go'"
run_test "AWS S3: put_object tool" \
    "grep -q 'put_object' '$PROJECT_ROOT/internal/mcp/adapters/aws_s3.go'"

# Brave Search tools
run_test "Brave Search: web_search tool" \
    "grep -q 'web_search' '$PROJECT_ROOT/internal/mcp/adapters/brave_search.go'"

# DataDog tools
run_test "DataDog: query_metrics tool" \
    "grep -q 'query_metrics' '$PROJECT_ROOT/internal/mcp/adapters/datadog.go'"

# Docker tools
run_test "Docker: list_containers tool" \
    "grep -q 'list_containers' '$PROJECT_ROOT/internal/mcp/adapters/docker.go'"
run_test "Docker: start_container tool" \
    "grep -q 'start_container' '$PROJECT_ROOT/internal/mcp/adapters/docker.go'"

# Figma tools
run_test "Figma: get_file tool" \
    "grep -q 'get_file' '$PROJECT_ROOT/internal/mcp/adapters/figma.go'"

# GitLab tools
run_test "GitLab: list_projects tool" \
    "grep -q 'list_projects' '$PROJECT_ROOT/internal/mcp/adapters/gitlab.go'"

# Google Drive tools
run_test "Google Drive: list_files tool" \
    "grep -q 'list_files' '$PROJECT_ROOT/internal/mcp/adapters/google_drive.go'"

# Jira tools
run_test "Jira: get_issue tool" \
    "grep -q 'get_issue' '$PROJECT_ROOT/internal/mcp/adapters/jira.go'"
run_test "Jira: create_issue tool" \
    "grep -q 'create_issue' '$PROJECT_ROOT/internal/mcp/adapters/jira.go'"

# Kubernetes tools
run_test "Kubernetes: list_pods tool" \
    "grep -qE 'list_pods|k8s_list_pods' '$PROJECT_ROOT/internal/mcp/adapters/kubernetes.go'"
run_test "Kubernetes: pod_logs tool" \
    "grep -qE 'get_logs|pod_logs|k8s_pod_logs' '$PROJECT_ROOT/internal/mcp/adapters/kubernetes.go'"

# Linear tools
run_test "Linear: list_issues tool" \
    "grep -q 'list_issues' '$PROJECT_ROOT/internal/mcp/adapters/linear.go'"
run_test "Linear: create_issue tool" \
    "grep -q 'create_issue' '$PROJECT_ROOT/internal/mcp/adapters/linear.go'"

# Miro tools
run_test "Miro: list_boards tool" \
    "grep -q 'list_boards' '$PROJECT_ROOT/internal/mcp/adapters/miro.go'"

# MongoDB tools
run_test "MongoDB: find tool" \
    "grep -q 'find' '$PROJECT_ROOT/internal/mcp/adapters/mongodb.go'"

# Notion tools
run_test "Notion: query_database tool" \
    "grep -q 'query_database' '$PROJECT_ROOT/internal/mcp/adapters/notion.go'"

# Puppeteer tools
run_test "Puppeteer: navigate tool" \
    "grep -q 'navigate' '$PROJECT_ROOT/internal/mcp/adapters/puppeteer.go'"
run_test "Puppeteer: screenshot tool" \
    "grep -q 'screenshot' '$PROJECT_ROOT/internal/mcp/adapters/puppeteer.go'"

# Sentry tools
run_test "Sentry: list_issues tool" \
    "grep -q 'list_issues' '$PROJECT_ROOT/internal/mcp/adapters/sentry.go'"

# Slack tools
run_test "Slack: post_message tool" \
    "grep -qE 'send_message|post_message|slack_post_message' '$PROJECT_ROOT/internal/mcp/adapters/slack.go'"
run_test "Slack: list_channels tool" \
    "grep -qE 'list_channels|slack_list_channels' '$PROJECT_ROOT/internal/mcp/adapters/slack.go'"

# Stable Diffusion tools
run_test "Stable Diffusion: txt2img tool" \
    "grep -qE 'text_to_image|txt2img|sd_txt2img' '$PROJECT_ROOT/internal/mcp/adapters/stable_diffusion.go'"

# SVGMaker tools
run_test "SVGMaker: generate_svg tool" \
    "grep -qE 'create_svg|generate|svg_generate' '$PROJECT_ROOT/internal/mcp/adapters/svgmaker.go'"

# ============================================================================
# SECTION 4: MCP SERVER TOOL DEFINITIONS
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: MCP Server Tool Definitions"
log_info "=============================================="

# Filesystem server
run_test "Filesystem: read_file capability" \
    "grep -q 'read_file\|ReadFile' '$PROJECT_ROOT/internal/mcp/servers/filesystem_adapter.go'"
run_test "Filesystem: write_file capability" \
    "grep -q 'write_file\|WriteFile' '$PROJECT_ROOT/internal/mcp/servers/filesystem_adapter.go'"

# Git server
run_test "Git: clone capability" \
    "grep -q 'clone\|Clone' '$PROJECT_ROOT/internal/mcp/servers/git_adapter.go'"
run_test "Git: commit capability" \
    "grep -q 'commit\|Commit' '$PROJECT_ROOT/internal/mcp/servers/git_adapter.go'"

# GitHub server
run_test "GitHub: list_repos capability" \
    "grep -q 'list_repos\|ListRepos\|repos' '$PROJECT_ROOT/internal/mcp/servers/github_adapter.go'"

# Fetch server
run_test "Fetch: http_get capability" \
    "grep -qE 'http_get|GET|fetch' '$PROJECT_ROOT/internal/mcp/servers/fetch_adapter.go'"

# Memory server
run_test "Memory: store capability" \
    "grep -qE 'store|Store|Set' '$PROJECT_ROOT/internal/mcp/servers/memory_adapter.go'"

# Postgres server
run_test "Postgres: query capability" \
    "grep -qE 'query|Query|Execute' '$PROJECT_ROOT/internal/mcp/servers/postgres_adapter.go'"

# SQLite server
run_test "SQLite: query capability" \
    "grep -qE 'query|Query|Execute' '$PROJECT_ROOT/internal/mcp/servers/sqlite_adapter.go'"

# Redis server
run_test "Redis: get/set capability" \
    "grep -qE 'Get|Set|get|set' '$PROJECT_ROOT/internal/mcp/servers/redis_adapter.go'"

# Qdrant server
run_test "Qdrant: search capability" \
    "grep -qE 'search|Search|Query' '$PROJECT_ROOT/internal/mcp/servers/qdrant_adapter.go'"

# Chroma server
run_test "Chroma: query capability" \
    "grep -qE 'query|Query|Search' '$PROJECT_ROOT/internal/mcp/servers/chroma_adapter.go'"

# Weaviate server
run_test "Weaviate: search capability" \
    "grep -qE 'search|Search|Query' '$PROJECT_ROOT/internal/mcp/servers/weaviate_adapter.go'"

# ============================================================================
# SECTION 5: MCP CATEGORIES
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: MCP Category Coverage"
log_info "=============================================="

run_test "Category: Database (mongodb, postgres, sqlite, redis)" \
    "grep -l 'CategoryDatabase\|database' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Storage (aws_s3, google_drive)" \
    "grep -l 'CategoryStorage\|storage' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Version Control (gitlab, github)" \
    "grep -l 'CategoryVersionControl\|git' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Productivity (jira, linear, asana, notion)" \
    "grep -l 'CategoryProductivity\|task\|issue' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Communication (slack)" \
    "grep -l 'CategoryCommunication\|message' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Infrastructure (docker, kubernetes)" \
    "grep -l 'CategoryInfrastructure\|container\|pod' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Analytics (datadog, sentry)" \
    "grep -l 'CategoryAnalytics\|metrics\|logs' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: AI (stable_diffusion)" \
    "grep -l 'CategoryAI\|image\|generation' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Design (figma, miro)" \
    "grep -l 'CategoryDesign\|design\|board' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

run_test "Category: Automation (puppeteer)" \
    "grep -l 'CategoryAutomation\|browser\|navigate' '$PROJECT_ROOT/internal/mcp/adapters/'*.go | wc -l | grep -qE '[1-9]'"

# ============================================================================
# SECTION 6: GO TESTS VALIDATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: Go Tests Validation"
log_info "=============================================="

log_info "Running MCP adapter tests..."
run_test "All MCP adapter tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/mcp/adapters/... -timeout 120s"

log_info "Running MCP server tests..."
run_test "All MCP server tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/mcp/servers/... -timeout 120s"

# ============================================================================
# SECTION 7: UNIFIED MANAGER
# ============================================================================
log_info "=============================================="
log_info "SECTION 7: Unified Manager"
log_info "=============================================="

run_test "UnifiedManager struct exists" \
    "grep -qE 'type Unified.*Manager struct' '$PROJECT_ROOT/internal/mcp/servers/unified_manager.go'"

run_test "UnifiedManager has RegisterAdapter" \
    "grep -q 'RegisterAdapter\|Register' '$PROJECT_ROOT/internal/mcp/servers/unified_manager.go'"

run_test "UnifiedManager has ExecuteRequest" \
    "grep -q 'ExecuteRequest\|Execute\|CallTool' '$PROJECT_ROOT/internal/mcp/servers/unified_manager.go'"

# ============================================================================
# SUMMARY
# ============================================================================
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "=============================================="
    log_success "ALL MCP TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge 0
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge 1
fi
