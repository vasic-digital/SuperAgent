#!/usr/bin/env bash
set -e

# @describe List all configured LLM providers in HelixAgent

# @env HELIXAGENT_ENDPOINT=http://localhost:7061 The HelixAgent endpoint
# @env LLM_OUTPUT=/dev/stdout The output path

main() {
    local endpoint="${HELIXAGENT_ENDPOINT:-http://localhost:7061}"
    
    curl -fsS "${endpoint}/v1/providers" 2>/dev/null | jq -r '.providers[] | "\(.type): \(.name) [\(.available ? "available" : "unavailable")]"' >> "$LLM_OUTPUT" || {
        echo "Failed to fetch providers from ${endpoint}" >> "$LLM_OUTPUT"
        exit 1
    }
}

eval "$(argc --argc-eval "$0" "$@")"
