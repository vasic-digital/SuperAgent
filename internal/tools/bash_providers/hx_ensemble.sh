#!/usr/bin/env bash
set -e

# @describe Send a completion request to HelixAgent ensemble
# This aggregates responses from multiple LLM providers.

# @option --message! The message to send
# @option --model[=helixagent/ensemble] The model to use
# @option --temperature[=0.7] Sampling temperature
# @option --max-tokens[=4096] Maximum tokens to generate

# @env HELIXAGENT_ENDPOINT=http://localhost:7061 The HelixAgent endpoint
# @env HELIXAGENT_API_KEY API key for authentication
# @env LLM_OUTPUT=/dev/stdout The output path

main() {
    local endpoint="${HELIXAGENT_ENDPOINT:-http://localhost:7061}"
    local model="${argc_model:-helixagent/ensemble}"
    local temp="${argc_temperature:-0.7}"
    local max_tokens="${argc_max_tokens:-4096}"
    
    local headers="-H 'Content-Type: application/json'"
    if [[ -n "$HELIXAGENT_API_KEY" ]]; then
        headers="$headers -H 'Authorization: Bearer $HELIXAGENT_API_KEY'"
    fi
    
    curl -fsS -X POST "${endpoint}/v1/ensemble/completions" \
        $headers \
        -d "{
            \"model\": \"$model\",
            \"messages\": [{\"role\": \"user\", \"content\": \"$argc_message\"}],
            \"temperature\": $temp,
            \"max_tokens\": $max_tokens
        }" 2>/dev/null | jq -r '.choices[0].message.content' >> "$LLM_OUTPUT" || {
        echo "Failed to get ensemble completion" >> "$LLM_OUTPUT"
        exit 1
    }
}

eval "$(argc --argc-eval "$0" "$@")"
