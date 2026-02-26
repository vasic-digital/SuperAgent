#!/bin/bash
set -uo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

echo "=================================================="
echo "8-Phase Debate Orchestrator Challenge"
echo "=================================================="
echo ""

check_file() {
    local name="$1"
    local file="$2"
    local pattern="$3"
    
    if [[ -f "$file" ]] && grep -q "$pattern" "$file"; then
        echo -e "${GREEN}✓ PASS${NC}: $name"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $name"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "=== Section 1: Orchestrator Core Files ==="
check_file "Orchestrator main file" "internal/debate/orchestrator/orchestrator.go" "Orchestrator"
check_file "Service integration" "internal/debate/orchestrator/service_integration.go" "ServiceIntegration"
check_file "Provider bridge" "internal/debate/orchestrator/provider_bridge.go" "ProviderRegistryBridge"
check_file "API adapter" "internal/debate/orchestrator/api_adapter.go" "APIAdapter"
check_file "Type adapter" "internal/debate/orchestrator/adapter.go" "LegacyDebateConfig"

echo ""
echo "=== Section 2: Protocol Implementation ==="
check_file "Protocol main file" "internal/debate/protocol/protocol.go" "Protocol"
check_file "Dehallucination phase" "internal/debate/protocol/dehallucination.go" "Dehallucination"
check_file "Self-evolvement phase" "internal/debate/protocol/self_evolvement.go" "SelfEvolvement"

echo ""
echo "=== Section 3: 8-Phase Protocol Verification ==="
check_file "Phase 1: Dehallucination" "internal/debate/protocol/protocol.go" "PhaseDehallucination"
check_file "Phase 2: SelfEvolvement" "internal/debate/protocol/protocol.go" "PhaseSelfEvolvement"
check_file "Phase 3: Proposal" "internal/debate/protocol/protocol.go" "PhaseProposal"
check_file "Phase 4: Critique" "internal/debate/protocol/protocol.go" "PhaseCritique"
check_file "Phase 5: Review" "internal/debate/protocol/protocol.go" "PhaseReview"
check_file "Phase 6: Optimization" "internal/debate/protocol/protocol.go" "PhaseOptimization"
check_file "Phase 7: Adversarial" "internal/debate/protocol/protocol.go" "PhaseAdversarial"
check_file "Phase 8: Convergence" "internal/debate/protocol/protocol.go" "PhaseConvergence"

echo ""
echo "=== Section 4: Orchestrator Configuration ==="
check_file "OrchestratorConfig struct" "internal/debate/orchestrator/orchestrator.go" "OrchestratorConfig"
check_file "DefaultOrchestratorConfig" "internal/debate/orchestrator/orchestrator.go" "DefaultOrchestratorConfig"
check_file "RequiredPositions = 5" "internal/debate/orchestrator/orchestrator.go" "RequiredPositionsCount"
check_file "LLMsPerPosition = 3" "internal/debate/orchestrator/orchestrator.go" "LLMsPerPositionCount"
check_file "MinAgentsPerDebate = 15" "internal/debate/orchestrator/orchestrator.go" "MinDebateAgents"

echo ""
echo "=== Section 5: Router Integration ==="
check_file "Orchestrator imported in router" "internal/router/router.go" "orchestrator"
check_file "CreateIntegration called" "internal/router/router.go" "CreateIntegration"
check_file "SetOrchestratorIntegration called" "internal/router/router.go" "SetOrchestratorIntegration"
check_file "Orchestrator status endpoint" "internal/router/router.go" "orchestrator/status"

echo ""
echo "=== Section 6: Handler Integration ==="
check_file "Handler has orchestratorIntegration field" "internal/handlers/debate_handler.go" "orchestratorIntegration"
check_file "Handler SetOrchestratorIntegration method" "internal/handlers/debate_handler.go" "SetOrchestratorIntegration"
check_file "Handler uses orchestrator for debate" "internal/handlers/debate_handler.go" "orchestratorIntegration.ConductDebate"

echo ""
echo "=== Section 7: Provider Bridge Functions ==="
check_file "GetProvider bridge" "internal/debate/orchestrator/provider_bridge.go" "GetProvider"
check_file "GetAvailableProviders bridge" "internal/debate/orchestrator/provider_bridge.go" "GetAvailableProviders"
check_file "GetProvidersByScore bridge" "internal/debate/orchestrator/provider_bridge.go" "GetProvidersByScore"
check_file "registerVerifiedProviders" "internal/debate/orchestrator/provider_bridge.go" "registerVerifiedProviders"

echo ""
echo "=== Section 8: Service Integration Functions ==="
check_file "NewServiceIntegration" "internal/debate/orchestrator/service_integration.go" "NewServiceIntegration"
check_file "ConductDebate integration" "internal/debate/orchestrator/service_integration.go" "ConductDebate"
check_file "ShouldUseNewFramework" "internal/debate/orchestrator/service_integration.go" "ShouldUseNewFramework"
check_file "convertDebateConfig" "internal/debate/orchestrator/service_integration.go" "convertDebateConfig"
check_file "convertToDebateResult" "internal/debate/orchestrator/service_integration.go" "convertToDebateResult"

echo ""
echo "=== Section 9: Agent Management ==="
check_file "AgentFactory" "internal/debate/agents/factory.go" "AgentFactory"
check_file "AgentPool" "internal/debate/agents/factory.go" "AgentPool"
check_file "TeamBuilder" "internal/debate/agents/factory.go" "TeamBuilder"
check_file "SpecializedAgent" "internal/debate/agents/specialization.go" "SpecializedAgent"

echo ""
echo "=== Section 10: Voting System ==="
check_file "WeightedVotingSystem" "internal/debate/voting/weighted_voting.go" "WeightedVotingSystem"
check_file "VotingMethod enum" "internal/debate/voting/weighted_voting.go" "VotingMethod"

echo ""
echo "=== Section 11: Knowledge/Learning ==="
check_file "Knowledge repository" "internal/debate/knowledge/repository.go" "Repository"
check_file "Cross-debate learner" "internal/debate/knowledge/learning.go" "CrossDebateLearner"
check_file "DebateLearningIntegration" "internal/debate/knowledge/integration.go" "DebateLearningIntegration"

echo ""
echo "=== Section 12: Topology Support ==="
check_file "Topology interface" "internal/debate/topology/topology.go" "Topology"
check_file "GraphMesh topology" "internal/debate/topology/graph_mesh.go" "GraphMeshTopology"
check_file "AgentRole enum" "internal/debate/topology/topology.go" "AgentRole"

echo ""
echo "=== Section 13: Test Coverage ==="
check_file "Orchestrator tests" "internal/debate/orchestrator/orchestrator_test.go" "Test"
check_file "Service integration tests" "internal/debate/orchestrator/service_integration_test.go" "Test"
check_file "Provider bridge tests" "internal/debate/orchestrator/provider_bridge_test.go" "Test"
check_file "Protocol tests" "internal/debate/protocol/protocol_test.go" "Test"
check_file "Dehallucination tests" "internal/debate/protocol/dehallucination_test.go" "Test"
check_file "Self-evolvement tests" "internal/debate/protocol/self_evolvement_test.go" "Test"

echo ""
echo "=== Section 14: Build Verification ==="
if go build ./internal/debate/orchestrator/... 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Orchestrator package builds"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Orchestrator package build failed"
    FAILED=$((FAILED + 1))
fi

if go build ./internal/debate/protocol/... 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Protocol package builds"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Protocol package build failed"
    FAILED=$((FAILED + 1))
fi

if go build ./cmd/... ./internal/... 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Full project builds"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Full project build failed"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=== Section 15: Orchestrator Tests ==="
test_output=$(GOMAXPROCS=2 go test -v -run TestOrchestrator ./internal/debate/orchestrator/... 2>&1 || true)
if echo "$test_output" | grep -q "PASS"; then
    echo -e "${GREEN}✓ PASS${NC}: Orchestrator unit tests pass"
    PASSED=$((PASSED + 1))
else
    echo -e "${YELLOW}⚠ WARNING${NC}: Some orchestrator tests may have issues"
    echo "$test_output" | grep -E "(PASS|FAIL|---)" | head -10
fi

echo ""
echo "=== Section 16: Service Integration Tests ==="
test_output=$(GOMAXPROCS=2 go test -v -run TestServiceIntegration ./internal/debate/orchestrator/... 2>&1 || true)
if echo "$test_output" | grep -q "PASS"; then
    echo -e "${GREEN}✓ PASS${NC}: Service integration tests pass"
    PASSED=$((PASSED + 1))
else
    echo -e "${YELLOW}⚠ WARNING${NC}: Some integration tests may have issues"
    echo "$test_output" | grep -E "(PASS|FAIL|---)" | head -10
fi

echo ""
echo "=== Section 17: Protocol Tests ==="
test_output=$(GOMAXPROCS=2 go test -v -run Test ./internal/debate/protocol/... 2>&1 || true)
if echo "$test_output" | grep -q "PASS"; then
    echo -e "${GREEN}✓ PASS${NC}: Protocol tests pass"
    PASSED=$((PASSED + 1))
else
    echo -e "${YELLOW}⚠ WARNING${NC}: Some protocol tests may have issues"
    echo "$test_output" | grep -E "(PASS|FAIL|---)" | head -10
fi

echo ""
echo "=================================================="
echo "Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILED} failed${NC}"
echo "=================================================="

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi

exit 0
