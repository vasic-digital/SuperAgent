#!/bin/bash
# SuperAgent Challenges - Report Generator
# Usage: ./scripts/generate_report.sh [options]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
MASTER_RESULTS_DIR="$CHALLENGES_DIR/master_results"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

MASTER_ONLY=false

while [ $# -gt 0 ]; do
    case "$1" in
        --master-only)
            MASTER_ONLY=true
            ;;
        -h|--help)
            echo "Usage: $0 [--master-only]"
            exit 0
            ;;
    esac
    shift
done

mkdir -p "$MASTER_RESULTS_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
MASTER_SUMMARY="$MASTER_RESULTS_DIR/master_summary_${TIMESTAMP}.md"

# Generate master summary
generate_master_summary() {
    print_info "Generating master summary..."

    cat > "$MASTER_SUMMARY" << EOF
# SuperAgent Challenges - Master Summary

Generated: $(date)

## Overview

| Challenge | Status | Duration | Last Run |
|-----------|--------|----------|----------|
EOF

    # All 38 challenges in dependency order
    CHALLENGES=(
        # Infrastructure
        "health_monitoring" "configuration_loading" "caching_layer" "database_operations"
        "authentication" "plugin_system"
        # Security
        "rate_limiting" "input_validation"
        # Providers
        "provider_claude" "provider_deepseek" "provider_gemini" "provider_ollama"
        "provider_openrouter" "provider_qwen" "provider_zai"
        # Core verification
        "provider_verification"
        # Protocols
        "mcp_protocol" "lsp_protocol" "acp_protocol"
        # Cloud integrations
        "cloud_aws_bedrock" "cloud_gcp_vertex" "cloud_azure_openai"
        # Core features
        "ensemble_voting" "embeddings_service" "streaming_responses" "model_metadata"
        # Debate
        "ai_debate_formation" "ai_debate_workflow"
        # API
        "openai_compatibility" "grpc_api" "api_quality_test"
        # Optimization
        "optimization_semantic_cache" "optimization_structured_output"
        # Integration
        "cognee_integration"
        # Resilience
        "circuit_breaker" "error_handling" "concurrent_access" "graceful_shutdown"
        # Session
        "session_management"
    )

    for challenge in "${CHALLENGES[@]}"; do
        local latest_dir=$(find "$CHALLENGES_DIR/results/$challenge" -maxdepth 4 -type d -name "[0-9]*_[0-9]*" 2>/dev/null | sort -r | head -1)

        if [ -n "$latest_dir" ] && [ -d "$latest_dir" ]; then
            local log_file="$latest_dir/logs/challenge.log"
            local status="unknown"
            local duration="N/A"
            local timestamp=$(basename "$latest_dir")

            if [ -f "$log_file" ]; then
                if grep -q '"event":"challenge_completed"' "$log_file"; then
                    local exit_code=$(grep '"event":"challenge_completed"' "$log_file" | tail -1 | grep -o '"exit_code":[0-9]*' | cut -d: -f2)
                    if [ "$exit_code" = "0" ]; then
                        status="PASSED"
                    else
                        status="FAILED"
                    fi
                    duration=$(grep '"event":"challenge_completed"' "$log_file" | tail -1 | grep -o '"duration_seconds":[0-9]*' | cut -d: -f2)
                    duration="${duration}s"
                fi
            fi

            # If status still unknown, check the report file directly
            if [ "$status" = "unknown" ]; then
                local report_file="$latest_dir/results/${challenge}_report.md"
                if [ -f "$report_file" ]; then
                    if grep -q "Status:\*\* PASSED" "$report_file"; then
                        status="PASSED"
                    elif grep -q "Status:\*\* FAILED" "$report_file"; then
                        status="FAILED"
                    fi
                fi
            fi

            echo "| $challenge | $status | $duration | $timestamp |" >> "$MASTER_SUMMARY"
        else
            echo "| $challenge | NOT RUN | - | - |" >> "$MASTER_SUMMARY"
        fi
    done

    cat >> "$MASTER_SUMMARY" << EOF

## Challenge Details

EOF

    # Add details for each challenge
    for challenge in "${CHALLENGES[@]}"; do
        local latest_dir=$(find "$CHALLENGES_DIR/results/$challenge" -maxdepth 4 -type d -name "[0-9]*_[0-9]*" 2>/dev/null | sort -r | head -1)

        echo "### $challenge" >> "$MASTER_SUMMARY"
        echo "" >> "$MASTER_SUMMARY"

        if [ -n "$latest_dir" ] && [ -d "$latest_dir" ]; then
            echo "Results directory: \`$latest_dir\`" >> "$MASTER_SUMMARY"
            echo "" >> "$MASTER_SUMMARY"

            # Include any markdown reports found
            local report_file="$latest_dir/results/${challenge}_report.md"
            if [ -f "$report_file" ]; then
                echo "#### Report" >> "$MASTER_SUMMARY"
                echo "" >> "$MASTER_SUMMARY"
                cat "$report_file" >> "$MASTER_SUMMARY"
                echo "" >> "$MASTER_SUMMARY"
            fi
        else
            echo "No results available." >> "$MASTER_SUMMARY"
            echo "" >> "$MASTER_SUMMARY"
        fi
    done

    cat >> "$MASTER_SUMMARY" << EOF

---

*This summary was automatically generated by the SuperAgent Challenges system.*
EOF

    print_success "Master summary generated: $MASTER_SUMMARY"

    # Create/update symlink to latest summary
    ln -sf "$MASTER_SUMMARY" "$MASTER_RESULTS_DIR/latest_summary.md"
}

generate_master_summary
