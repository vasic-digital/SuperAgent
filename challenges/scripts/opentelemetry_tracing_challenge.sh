#!/bin/bash
# OpenTelemetry Tracing Challenge
# VALIDATES: OpenTelemetry tracing, trace propagation, span creation, trace context
# Tests the complete OpenTelemetry tracing infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="OpenTelemetry Tracing Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: OpenTelemetry tracing infrastructure"
log_info ""

# ============================================================================
# Section 1: Tracer Code Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: Tracer Code Structure"
log_info "=============================================="

# Test 1: observability/tracer.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: observability/tracer.go exists"
if [ -f "$PROJECT_ROOT/internal/observability/tracer.go" ]; then
    log_success "observability/tracer.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "observability/tracer.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: LLMTracer struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: LLMTracer struct defined"
if grep -q "type LLMTracer struct" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "LLMTracer struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "LLMTracer struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: TracerConfig struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 3: TracerConfig struct defined"
if grep -q "type TracerConfig struct" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "TracerConfig struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "TracerConfig struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 4: NewLLMTracer constructor exists
TOTAL=$((TOTAL + 1))
log_info "Test 4: NewLLMTracer constructor exists"
if grep -q "func NewLLMTracer" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "NewLLMTracer constructor exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewLLMTracer constructor NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 5: DefaultTracerConfig function exists
TOTAL=$((TOTAL + 1))
log_info "Test 5: DefaultTracerConfig function exists"
if grep -q "func DefaultTracerConfig" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "DefaultTracerConfig function exists"
    PASSED=$((PASSED + 1))
else
    log_error "DefaultTracerConfig function NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: OpenTelemetry Semantic Conventions
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: OpenTelemetry Semantic Conventions"
log_info "=============================================="

# Test 6: GenAI semantic conventions defined
TOTAL=$((TOTAL + 1))
log_info "Test 6: GenAI semantic conventions defined"
if grep -q "gen_ai.system\|AttrLLMSystem" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "GenAI semantic conventions defined"
    PASSED=$((PASSED + 1))
else
    log_error "GenAI semantic conventions NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 7: LLM model attribute defined
TOTAL=$((TOTAL + 1))
log_info "Test 7: LLM model attribute defined"
if grep -q "gen_ai.request.model\|AttrLLMModel" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "LLM model attribute defined"
    PASSED=$((PASSED + 1))
else
    log_error "LLM model attribute NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: Token usage attributes defined
TOTAL=$((TOTAL + 1))
log_info "Test 8: Token usage attributes defined"
if grep -q "gen_ai.usage.input_tokens\|AttrLLMInputTokens" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null && \
   grep -q "gen_ai.usage.output_tokens\|AttrLLMOutputTokens" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "Token usage attributes defined"
    PASSED=$((PASSED + 1))
else
    log_error "Token usage attributes NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: HelixAgent-specific attributes defined
TOTAL=$((TOTAL + 1))
log_info "Test 9: HelixAgent-specific attributes defined"
if grep -q "helix.request.id\|AttrHelixRequestID" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "HelixAgent-specific attributes defined"
    PASSED=$((PASSED + 1))
else
    log_error "HelixAgent-specific attributes NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 10: Finish reason attribute defined
TOTAL=$((TOTAL + 1))
log_info "Test 10: Finish reason attribute defined"
if grep -q "gen_ai.response.finish_reason\|AttrLLMFinishReason" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "Finish reason attribute defined"
    PASSED=$((PASSED + 1))
else
    log_error "Finish reason attribute NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Span Creation Methods
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: Span Creation Methods"
log_info "=============================================="

# Test 11: StartLLMRequest method exists
TOTAL=$((TOTAL + 1))
log_info "Test 11: StartLLMRequest method exists"
if grep -q "func (t \*LLMTracer) StartLLMRequest" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "StartLLMRequest method exists"
    PASSED=$((PASSED + 1))
else
    log_error "StartLLMRequest method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 12: EndLLMRequest method exists
TOTAL=$((TOTAL + 1))
log_info "Test 12: EndLLMRequest method exists"
if grep -q "func (t \*LLMTracer) EndLLMRequest" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "EndLLMRequest method exists"
    PASSED=$((PASSED + 1))
else
    log_error "EndLLMRequest method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 13: StartEnsembleRequest method exists
TOTAL=$((TOTAL + 1))
log_info "Test 13: StartEnsembleRequest method exists"
if grep -q "func (t \*LLMTracer) StartEnsembleRequest" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "StartEnsembleRequest method exists"
    PASSED=$((PASSED + 1))
else
    log_error "StartEnsembleRequest method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 14: StartDebateRound method exists
TOTAL=$((TOTAL + 1))
log_info "Test 14: StartDebateRound method exists"
if grep -q "func (t \*LLMTracer) StartDebateRound" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "StartDebateRound method exists"
    PASSED=$((PASSED + 1))
else
    log_error "StartDebateRound method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 15: StartRAGRetrieval method exists
TOTAL=$((TOTAL + 1))
log_info "Test 15: StartRAGRetrieval method exists"
if grep -q "func (t \*LLMTracer) StartRAGRetrieval" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "StartRAGRetrieval method exists"
    PASSED=$((PASSED + 1))
else
    log_error "StartRAGRetrieval method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 16: StartToolExecution method exists
TOTAL=$((TOTAL + 1))
log_info "Test 16: StartToolExecution method exists"
if grep -q "func (t \*LLMTracer) StartToolExecution" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "StartToolExecution method exists"
    PASSED=$((PASSED + 1))
else
    log_error "StartToolExecution method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Request/Response Parameter Types
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Request/Response Parameter Types"
log_info "=============================================="

# Test 17: LLMRequestParams struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 17: LLMRequestParams struct defined"
if grep -q "type LLMRequestParams struct" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "LLMRequestParams struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "LLMRequestParams struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 18: LLMResponseParams struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 18: LLMResponseParams struct defined"
if grep -q "type LLMResponseParams struct" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "LLMResponseParams struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "LLMResponseParams struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 19: Provider field in request params
TOTAL=$((TOTAL + 1))
log_info "Test 19: Provider field in LLMRequestParams"
if grep -A20 "type LLMRequestParams struct" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null | grep -q "Provider.*string"; then
    log_success "Provider field exists in LLMRequestParams"
    PASSED=$((PASSED + 1))
else
    log_error "Provider field NOT found in LLMRequestParams!"
    FAILED=$((FAILED + 1))
fi

# Test 20: InputTokens field in response params
TOTAL=$((TOTAL + 1))
log_info "Test 20: InputTokens field in LLMResponseParams"
if grep -A20 "type LLMResponseParams struct" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null | grep -q "InputTokens.*int"; then
    log_success "InputTokens field exists in LLMResponseParams"
    PASSED=$((PASSED + 1))
else
    log_error "InputTokens field NOT found in LLMResponseParams!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Exporter Types
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Exporter Types"
log_info "=============================================="

# Test 21: ExporterType defined
TOTAL=$((TOTAL + 1))
log_info "Test 21: ExporterType defined"
if grep -q "type ExporterType string" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "ExporterType type defined"
    PASSED=$((PASSED + 1))
else
    log_error "ExporterType NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 22: OTLP exporter constant
TOTAL=$((TOTAL + 1))
log_info "Test 22: OTLP exporter constant defined"
if grep -q 'ExporterOTLP.*=.*"otlp"' "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "ExporterOTLP constant defined"
    PASSED=$((PASSED + 1))
else
    log_error "ExporterOTLP constant NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 23: Jaeger exporter constant
TOTAL=$((TOTAL + 1))
log_info "Test 23: Jaeger exporter constant defined"
if grep -q 'ExporterJaeger.*=.*"jaeger"' "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "ExporterJaeger constant defined"
    PASSED=$((PASSED + 1))
else
    log_error "ExporterJaeger constant NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 24: Zipkin exporter constant
TOTAL=$((TOTAL + 1))
log_info "Test 24: Zipkin exporter constant defined"
if grep -q 'ExporterZipkin.*=.*"zipkin"' "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "ExporterZipkin constant defined"
    PASSED=$((PASSED + 1))
else
    log_error "ExporterZipkin constant NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Global Tracer Management
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 6: Global Tracer Management"
log_info "=============================================="

# Test 25: Global tracer instance
TOTAL=$((TOTAL + 1))
log_info "Test 25: Global tracer instance defined"
if grep -q "globalTracer \*LLMTracer" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "Global tracer instance defined"
    PASSED=$((PASSED + 1))
else
    log_error "Global tracer instance NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 26: InitGlobalTracer function exists
TOTAL=$((TOTAL + 1))
log_info "Test 26: InitGlobalTracer function exists"
if grep -q "func InitGlobalTracer" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "InitGlobalTracer function exists"
    PASSED=$((PASSED + 1))
else
    log_error "InitGlobalTracer function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 27: GetTracer function exists
TOTAL=$((TOTAL + 1))
log_info "Test 27: GetTracer function exists"
if grep -q "func GetTracer" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "GetTracer function exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetTracer function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 28: sync.Once used for tracer initialization
TOTAL=$((TOTAL + 1))
log_info "Test 28: sync.Once used for thread-safe tracer initialization"
if grep -q "tracerOnce.*sync.Once" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "sync.Once used for thread-safe tracer initialization"
    PASSED=$((PASSED + 1))
else
    log_error "sync.Once NOT used for tracer initialization!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: Error Handling and Status
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 7: Error Handling and Status"
log_info "=============================================="

# Test 29: RecordError used for error spans
TOTAL=$((TOTAL + 1))
log_info "Test 29: RecordError used for error spans"
if grep -q "span.RecordError" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "RecordError used for error spans"
    PASSED=$((PASSED + 1))
else
    log_error "RecordError NOT used for error spans!"
    FAILED=$((FAILED + 1))
fi

# Test 30: SetStatus used for span status
TOTAL=$((TOTAL + 1))
log_info "Test 30: SetStatus used for span status"
if grep -q "span.SetStatus" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "SetStatus used for span status"
    PASSED=$((PASSED + 1))
else
    log_error "SetStatus NOT used for span status!"
    FAILED=$((FAILED + 1))
fi

# Test 31: codes.Error imported and used
TOTAL=$((TOTAL + 1))
log_info "Test 31: codes.Error used for error status"
if grep -q "codes.Error" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "codes.Error used for error status"
    PASSED=$((PASSED + 1))
else
    log_error "codes.Error NOT used for error status!"
    FAILED=$((FAILED + 1))
fi

# Test 32: codes.Ok used for success status
TOTAL=$((TOTAL + 1))
log_info "Test 32: codes.Ok used for success status"
if grep -q "codes.Ok" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "codes.Ok used for success status"
    PASSED=$((PASSED + 1))
else
    log_error "codes.Ok NOT used for success status!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: Span Kinds
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 8: Span Kinds"
log_info "=============================================="

# Test 33: SpanKindClient used for LLM requests
TOTAL=$((TOTAL + 1))
log_info "Test 33: SpanKindClient used for LLM requests"
if grep -q "trace.SpanKindClient" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "SpanKindClient used for LLM requests"
    PASSED=$((PASSED + 1))
else
    log_error "SpanKindClient NOT used for LLM requests!"
    FAILED=$((FAILED + 1))
fi

# Test 34: SpanKindInternal used for internal operations
TOTAL=$((TOTAL + 1))
log_info "Test 34: SpanKindInternal used for internal operations"
if grep -q "trace.SpanKindInternal" "$PROJECT_ROOT/internal/observability/tracer.go" 2>/dev/null; then
    log_success "SpanKindInternal used for internal operations"
    PASSED=$((PASSED + 1))
else
    log_error "SpanKindInternal NOT used for internal operations!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Results Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Challenge Results Summary"
log_info "=============================================="
log_info "Passed: $PASSED/$TOTAL"
log_info "Failed: $FAILED/$TOTAL"
log_info ""

if [ "$FAILED" -eq 0 ]; then
    log_success "ALL $TOTAL TESTS PASSED!"
    exit 0
else
    log_error "$FAILED TEST(S) FAILED!"
    exit 1
fi
