#!/usr/bin/env bash
set -e

# @describe Check HelixAgent health status
# Use this to verify HelixAgent is running and healthy.

# @env HELIXAGENT_ENDPOINT=http://localhost:7061 The HelixAgent endpoint
# @env LLM_OUTPUT=/dev/stdout The output path

main() {
    local endpoint="${HELIXAGENT_ENDPOINT:-http://localhost:7061}"
    
    curl -fsS "${endpoint}/health" 2>/dev/null | jq -r '.' >> "$LLM_OUTPUT" || {
        echo "HelixAgent is not responding at ${endpoint}" >> "$LLM_OUTPUT"
        exit 1
    }
}

eval "$(argc --argc-eval "$0" "$@")"
