#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"
CHALLENGE_PORT="${HELIXAGENT_PORT:-7061}"
BASE_URL="http://localhost:$CHALLENGE_PORT"
init_challenge "${challenge_name//_/-}" "$(echo ${challenge_name//_/ } | sed 's/\b\(.\)/\u\1/g') Challenge"
load_env
log_info "Testing ${challenge_name//_/ }..."

test_basic_functionality() {
    local request='{"model":"helixagent-debate","messages":[{"role":"user","content":"Test"}],"max_tokens":10}'
    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)
    local http_code=$(echo "$response" | tail -n1)
    [[ "$http_code" == "200" ]] && record_assertion "basic" "works" "true" "Basic request works" || record_assertion "basic" "works" "false" "Failed: $http_code"
}

main() {
    if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        start_helixagent "$CHALLENGE_PORT" || { finalize_challenge "FAILED"; exit 1; }
    fi
    test_basic_functionality
    local failed_count=$(grep -c "|FAILED|" "$OUTPUT_DIR/logs/assertions.log" 2>/dev/null || echo "0")
    [[ "${failed_count:-0}" -eq 0 ]] && finalize_challenge "PASSED" || finalize_challenge "FAILED"
}
main "$@"
