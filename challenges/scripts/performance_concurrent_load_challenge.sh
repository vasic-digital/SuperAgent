#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"
CHALLENGE_PORT="${HELIXAGENT_PORT:-7061}"
BASE_URL="http://localhost:$CHALLENGE_PORT"
init_challenge "${challenge//_/-}" "$(echo ${challenge//_/ } | sed 's/\b\(.\)/\u\1/g') Challenge"
load_env
log_info "Testing $(echo ${challenge//_/ })..."
test_performance() {
    local start=$(date +%s%N)
    local req='{"model":"helixagent-debate","messages":[{"role":"user","content":"Test"}],"max_tokens":10}'
    local resp=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" -H "Content-Type: application/json" -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" -d "$req" --max-time 30 2>/dev/null || true)
    local latency=$(( ($(date +%s%N) - start) / 1000000 ))
    record_metric "latency_ms" "$latency"
    [[ "$(echo "$resp" | tail -n1)" == "200" ]] && [[ $latency -lt 30000 ]] && record_assertion "perf" "acceptable" "true" "${latency}ms" || record_assertion "perf" "acceptable" "false" "${latency}ms"
}
main() {
    curl -s "$BASE_URL/health" > /dev/null 2>&1 || start_helixagent "$CHALLENGE_PORT" || { finalize_challenge "FAILED"; exit 1; }
    test_performance
    [[ "$(grep -c "|FAILED|" "$OUTPUT_DIR/logs/assertions.log" 2>/dev/null || echo 0)" -eq 0 ]] && finalize_challenge "PASSED" || finalize_challenge "FAILED"
}
main "$@"
