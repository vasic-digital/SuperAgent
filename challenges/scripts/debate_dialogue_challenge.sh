#!/bin/bash
# ═══════════════════════════════════════════════════════════════
# HelixAgent Debate Dialogue Challenge
# Validates that AI debate ensemble details (team members,
# positions, phases, voting, footer) are displayed to CLI agents
# ═══════════════════════════════════════════════════════════════
set -euo pipefail

PASSED=0
FAILED=0
TOTAL=0

pass() { PASSED=$((PASSED+1)); TOTAL=$((TOTAL+1)); echo "  PASS #$TOTAL: $1"; }
fail() { FAILED=$((FAILED+1)); TOTAL=$((TOTAL+1)); echo "  FAIL #$TOTAL: $1"; }
check() { if eval "$2" >/dev/null 2>&1; then pass "$1"; else fail "$1"; fi; }

echo "╔════════════════════════════════════════════════════════════╗"
echo "║        DEBATE DIALOGUE CHALLENGE (18 tests)               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# ─── SOURCE CODE VALIDATION ───
echo "--- Source Code Validation ---"

check "showDebateDialogue set to true in constructor" \
  "grep -q 'showDebateDialogue: true' internal/handlers/openai_compatible.go"

check "showDebateDialogue field exists in UnifiedHandler" \
  "grep -q 'showDebateDialogue.*bool' internal/handlers/openai_compatible.go"

check "generateDebateDialogueIntroduction function exists" \
  "grep -q 'func.*generateDebateDialogueIntroduction' internal/handlers/openai_compatible.go"

check "generateResponseFooter function exists" \
  "grep -q 'func.*generateResponseFooter' internal/handlers/openai_compatible.go"

check "generateDebateDialogueConclusion function exists" \
  "grep -q 'func.*generateDebateDialogueConclusion' internal/handlers/openai_compatible.go"

check "Footer contains HelixAgent AI Debate Ensemble text" \
  "grep -q 'HelixAgent AI Debate Ensemble' internal/handlers/openai_compatible.go"

check "Footer contains Synthesized text" \
  "grep -q 'Synthesized from' internal/handlers/openai_compatible.go"

check "Dialogue formatter initialized in constructor" \
  "grep -q 'NewDialogueFormatter' internal/handlers/openai_compatible.go"

check "Streaming gating checks showDebateDialogue" \
  "grep -q 'h.showDebateDialogue && h.dialogueFormatter' internal/handlers/openai_compatible.go"

check "Footer gating checks showDebateDialogue" \
  "grep -q 'if h.showDebateDialogue' internal/handlers/openai_compatible.go"

# ─── UNIT TEST VALIDATION ───
echo ""
echo "--- Unit Test Validation ---"

check "TestDebateDialogueEnabledByDefault test exists" \
  "grep -q 'TestDebateDialogueEnabledByDefault' internal/handlers/openai_compatible_test.go"

check "TestDebateDialogueFooterContainsEnsembleInfo test exists" \
  "grep -q 'TestDebateDialogueFooterContainsEnsembleInfo' internal/handlers/openai_compatible_test.go"

check "TestDebateDialogueConclusionContainsConsensus test exists" \
  "grep -q 'TestDebateDialogueConclusionContainsConsensus' internal/handlers/openai_compatible_test.go"

check "TestDebateDialogueGatingConditions test exists" \
  "grep -q 'TestDebateDialogueGatingConditions' internal/handlers/openai_compatible_test.go"

check "TestDebateDialogueFormatterInitialized test exists" \
  "grep -q 'TestDebateDialogueFormatterInitialized' internal/handlers/openai_compatible_test.go"

check "TestSimpleMessageTrimming test exists" \
  "grep -q 'TestSimpleMessageTrimming' internal/handlers/openai_compatible_test.go"

check "Simple message trimming logic in streaming handler" \
  "grep -q 'Trimmed system context for greeting' internal/handlers/openai_compatible.go"

check "Minimal system prompt for simple messages" \
  "grep -q 'Respond naturally and conversationally' internal/handlers/openai_compatible.go"

# ─── SUMMARY ───
echo ""
echo "════════════════════════════════════════════════════════════"
echo "  DEBATE DIALOGUE CHALLENGE: $PASSED/$TOTAL passed ($FAILED failed)"
echo "════════════════════════════════════════════════════════════"

[ "$FAILED" -eq 0 ] && exit 0 || exit 1
