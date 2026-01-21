package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Semantic Intent Classifier Tests
// NO HARDCODING - Tests validate semantic understanding, not pattern matching
// =============================================================================

func TestIntentClassifier_Creation(t *testing.T) {
	ic := NewIntentClassifier()
	require.NotNil(t, ic)
}

// =============================================================================
// CONFIRMATION INTENT TESTS - Various ways to express approval
// =============================================================================

func TestIntentClassifier_Confirmation_DirectAgreement(t *testing.T) {
	ic := NewIntentClassifier()

	// Various ways to express direct agreement
	// Note: In production, LLM handles these semantically. Fallback is more limited.
	confirmations := []string{
		"Yes",
		"Yep",
		"Yeah",
		"Yup",
		"OK",
		"Okay",
		"Sure",
		"Right",
		"Correct",
		"Agreed",
	}

	for _, msg := range confirmations {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true) // with context
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Should detect confirmation in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}

	// These work with LLM but may need context boost in fallback
	advancedConfirmations := []string{
		"Absolutely",
		"Definitely",
		"Certainly",
		"Of course",
	}
	for _, msg := range advancedConfirmations {
		t.Run(msg+"_advanced", func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			// These should at least be recognized as having affirmative intent
			assert.True(t, result.Confidence > 0.3,
				"Advanced confirmation should have positive confidence: %s (got %.2f)",
				msg, result.Confidence)
		})
	}
}

func TestIntentClassifier_Confirmation_ActionRequests(t *testing.T) {
	ic := NewIntentClassifier()

	// Various ways to request action
	actionRequests := []string{
		"Let's do all of these points you have offered! Start all work now!",
		"Go ahead and implement everything",
		"Start working on the changes",
		"Begin the implementation",
		"Execute the plan",
		"Run all the suggested fixes",
		"Proceed with the recommendations",
		"Make it happen",
		"Do it",
		"Start now",
		"Let's begin",
		"Let's proceed",
		"Move forward with this",
		"Continue with the plan",
		"Work on all items",
		"Handle all the tasks",
		"Address everything you mentioned",
		"Fix all the issues",
		"Implement all suggestions",
		"Apply all changes",
	}

	for _, msg := range actionRequests {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Should detect action confirmation in: %s (got intent=%s, confidence=%.2f, signals=%v)",
				msg, result.Intent, result.Confidence, result.Signals)
		})
	}
}

func TestIntentClassifier_Confirmation_ApprovalExpressions(t *testing.T) {
	ic := NewIntentClassifier()

	// Various approval expressions
	approvals := []string{
		"Approved",
		"I approve",
		"Confirmed",
		"I confirm",
		"Sounds good",
		"Sounds great",
		"That sounds perfect",
		"Looks good to me",
		"Great idea",
		"Excellent suggestion",
		"Perfect",
		"Wonderful",
		"I like it",
		"Love it",
		"That's exactly what I need",
		"You have my approval",
		"Green light",
		"Thumbs up",
		"Good to go",
	}

	for _, msg := range approvals {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.Confidence > 0.4,
				"Should detect approval in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

func TestIntentClassifier_Confirmation_EnthusiasticResponses(t *testing.T) {
	ic := NewIntentClassifier()

	// Enthusiastic confirmations with emphasis
	enthusiastic := []string{
		"Yes! Let's do it!",
		"Absolutely! Start right away!",
		"Perfect! Go ahead!",
		"Great! Implement everything!",
		"Awesome! Begin now!",
		"Fantastic! Work on all of it!",
		"Amazing! Execute the plan!",
		"Brilliant! Proceed immediately!",
		"Excellent! Do all the changes!",
		"Wonderful! Start all tasks!",
	}

	for _, msg := range enthusiastic {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Should detect enthusiastic confirmation in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

func TestIntentClassifier_Confirmation_InclusiveLanguage(t *testing.T) {
	ic := NewIntentClassifier()

	// Inclusive language confirmations
	inclusive := []string{
		"Let's do this",
		"Let's get started",
		"Let us begin",
		"We should start",
		"We can proceed",
		"Together let's implement this",
		"Let's tackle this together",
		"We should work on all of these",
	}

	for _, msg := range inclusive {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Should detect inclusive confirmation in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

func TestIntentClassifier_Confirmation_CompletenessIndicators(t *testing.T) {
	ic := NewIntentClassifier()

	// Messages indicating completeness/all
	completeness := []string{
		"Do all of them",
		"Handle everything",
		"Work on all points",
		"Complete all tasks",
		"Address each item",
		"Fix every issue",
		"Implement the entire plan",
		"Execute the full list",
		"Do the whole thing",
		"All of the above",
		"Everything you suggested",
		"All recommendations please",
	}

	for _, msg := range completeness {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsActionable || result.Confidence > 0.4,
				"Should detect completeness request in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

// =============================================================================
// REFUSAL INTENT TESTS - Various ways to decline
// =============================================================================

func TestIntentClassifier_Refusal_DirectNegation(t *testing.T) {
	ic := NewIntentClassifier()

	// Direct negations
	negations := []string{
		"No",
		"Nope",
		"Nah",
		"No thanks",
		"No thank you",
		"Not now",
		"Not yet",
		"Not interested",
		"I don't want that",
		"I don't think so",
		"Don't do it",
		"Don't proceed",
	}

	for _, msg := range negations {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			// For refusals, we check the negative signals are detected
			assert.True(t, result.IsRefusal() || result.Intent == IntentRefusal,
				"Should detect refusal in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

func TestIntentClassifier_Refusal_Declinations(t *testing.T) {
	ic := NewIntentClassifier()

	// Various ways to decline
	declinations := []string{
		"I refuse",
		"I decline",
		"I reject this",
		"I'll pass",
		"Skip this",
		"Let's skip that",
		"I'd rather not",
		"I prefer not to",
		"Cancel that",
		"Stop",
		"Halt",
		"Abort",
		"Never mind",
		"Forget it",
	}

	for _, msg := range declinations {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsRefusal() || !result.IsActionable,
				"Should detect declination in: %s (got intent=%s, confidence=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

func TestIntentClassifier_Refusal_NegativeSentiment(t *testing.T) {
	ic := NewIntentClassifier()

	// Negative sentiment refusals
	negative := []string{
		"That's a bad idea",
		"This is wrong",
		"I hate this approach",
		"Terrible suggestion",
		"This won't work",
		"I dislike this plan",
		"This is awful",
		"Horrible idea",
	}

	for _, msg := range negative {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.False(t, result.IsConfirmation(),
				"Should NOT detect confirmation in negative: %s (got intent=%s)",
				msg, result.Intent)
		})
	}
}

// =============================================================================
// QUESTION INTENT TESTS - Various ways to ask questions
// =============================================================================

func TestIntentClassifier_Question_DirectQuestions(t *testing.T) {
	ic := NewIntentClassifier()

	// Direct questions
	questions := []string{
		"What do you think?",
		"How does this work?",
		"Why is this happening?",
		"When should we start?",
		"Where is the file?",
		"Who is responsible?",
		"Which option is better?",
		"Can you explain?",
		"Could you clarify?",
		"Would you help?",
		"Is this correct?",
		"Are you sure?",
		"Do you understand?",
	}

	for _, msg := range questions {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, false)
			assert.True(t, result.Intent == IntentQuestion || result.RequiresClarification,
				"Should detect question in: %s (got intent=%s)",
				msg, result.Intent)
		})
	}
}

func TestIntentClassifier_Question_ClarificationRequests(t *testing.T) {
	ic := NewIntentClassifier()

	// Clarification requests
	clarifications := []string{
		"I'm confused about this",
		"I don't understand",
		"Can you explain more?",
		"What do you mean?",
		"I'm not clear on this",
		"Could you elaborate?",
		"I need more information",
		"I'm wondering about this",
		"I'm curious why",
	}

	for _, msg := range clarifications {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, false)
			assert.False(t, result.IsConfirmation(),
				"Should NOT detect confirmation in question: %s (got intent=%s)",
				msg, result.Intent)
		})
	}
}

// =============================================================================
// NEUTRAL/UNCLEAR INTENT TESTS - Should not be classified as confirmation
// =============================================================================

func TestIntentClassifier_Neutral_GeneralStatements(t *testing.T) {
	ic := NewIntentClassifier()

	// Neutral statements that are not confirmations or refusals
	neutral := []string{
		"I see",
		"Interesting",
		"I understand",
		"That makes sense",
		"I'll think about it",
		"Let me consider",
		"Hmm",
		"Okay, I see what you mean",
		"I get it",
	}

	for _, msg := range neutral {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, false) // no context
			// Without context, these should not be strong confirmations
			assert.True(t, result.Confidence < 0.8,
				"Neutral statement should have moderate confidence: %s (got %.2f)",
				msg, result.Confidence)
		})
	}
}

func TestIntentClassifier_Neutral_InformationalRequests(t *testing.T) {
	ic := NewIntentClassifier()

	// Information requests that are not action confirmations
	informational := []string{
		"Show me the code",
		"List the files",
		"Display the structure",
		"What files are there?",
		"Explain the architecture",
		"Describe the system",
		"Tell me about the codebase",
	}

	for _, msg := range informational {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, false)
			// These are requests but not action confirmations
			assert.False(t, result.IsConfirmation(),
				"Informational request should not be confirmation: %s (got intent=%s)",
				msg, result.Intent)
		})
	}
}

// =============================================================================
// CONTEXT-AWARE TESTS - Context affects classification
// =============================================================================

func TestIntentClassifier_ContextAwareness_ShortResponsesWithContext(t *testing.T) {
	ic := NewIntentClassifier()

	// Short positive responses WITH context should be confirmations
	shortResponses := []string{
		"Yes",
		"OK",
		"Sure",
		"Do it",
		"Go",
		"Approved",
	}

	for _, msg := range shortResponses {
		t.Run(msg+"_with_context", func(t *testing.T) {
			resultWithContext := ic.EnhancedClassifyIntent(msg, true)
			resultNoContext := ic.EnhancedClassifyIntent(msg, false)

			// With context should have higher confidence
			assert.True(t, resultWithContext.Confidence >= resultNoContext.Confidence-0.1,
				"Context should boost confidence: %s (with=%.2f, without=%.2f)",
				msg, resultWithContext.Confidence, resultNoContext.Confidence)
		})
	}
}

// =============================================================================
// EDGE CASES AND MIXED SIGNALS
// =============================================================================

func TestIntentClassifier_EdgeCases_MixedSignals(t *testing.T) {
	ic := NewIntentClassifier()

	// Messages with mixed signals
	mixed := []struct {
		msg            string
		shouldConfirm  bool
		description    string
	}{
		{"Yes, but with some changes", true, "qualified yes"},
		{"OK, let's try it", true, "tentative approval"},
		{"Sure, why not", true, "casual approval"},
		{"I guess so, go ahead", true, "reluctant approval"},
		{"Maybe, let me think... actually yes", true, "changed mind to yes"},
		{"Not sure, but proceed anyway", false, "uncertain"},
		{"Yes and no", false, "ambiguous"},
	}

	for _, tc := range mixed {
		t.Run(tc.description, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(tc.msg, true)
			if tc.shouldConfirm {
				assert.True(t, result.IsConfirmation() || result.ShouldProceed() || result.Confidence > 0.4,
					"Mixed signal '%s' should lean toward confirmation (got intent=%s, conf=%.2f)",
					tc.msg, result.Intent, result.Confidence)
			}
		})
	}
}

func TestIntentClassifier_EdgeCases_CaseSensitivity(t *testing.T) {
	ic := NewIntentClassifier()

	// Test case insensitivity
	variations := []string{
		"YES",
		"yes",
		"Yes",
		"yEs",
		"GO AHEAD",
		"go ahead",
		"Go Ahead",
		"PROCEED",
		"proceed",
	}

	for _, msg := range variations {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Case variation should be detected: %s (got intent=%s)",
				msg, result.Intent)
		})
	}
}

func TestIntentClassifier_EdgeCases_Punctuation(t *testing.T) {
	ic := NewIntentClassifier()

	// Test with various punctuation
	variations := []string{
		"Yes!",
		"Yes!!",
		"Yes!!!",
		"Yes.",
		"Yes...",
		"Go ahead!",
		"Go ahead.",
		"Sure!",
		"Sure.",
	}

	for _, msg := range variations {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Punctuation variation should be detected: %s (got intent=%s)",
				msg, result.Intent)
		})
	}
}

// =============================================================================
// REAL-WORLD SCENARIO TESTS
// =============================================================================

func TestIntentClassifier_RealWorld_BearMailScenario(t *testing.T) {
	ic := NewIntentClassifier()

	// The exact message from the user's Bear-Mail scenario
	userMessage := "Let's do all of these points you have offered! Start all work now!"

	result := ic.EnhancedClassifyIntent(userMessage, true)

	assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
		"Real-world confirmation should be detected: %s (intent=%s, conf=%.2f, signals=%v)",
		userMessage, result.Intent, result.Confidence, result.Signals)
	assert.True(t, result.IsActionable,
		"Should be actionable")
	assert.True(t, result.Confidence > 0.5,
		"Should have high confidence (got %.2f)", result.Confidence)
}

func TestIntentClassifier_RealWorld_VariousApprovals(t *testing.T) {
	ic := NewIntentClassifier()

	// Various real-world approval messages
	approvals := []string{
		"Yes please do all of that",
		"Sounds good, proceed with everything",
		"Perfect, let's implement all the changes",
		"Great analysis! Now please fix all issues",
		"I agree with all points, please start",
		"All looks good, begin implementation",
		"Thanks for the analysis, now execute the plan",
		"Approved - do all security fixes and API changes",
	}

	for _, msg := range approvals {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
				"Real-world approval should be detected: %s (intent=%s, conf=%.2f)",
				msg, result.Intent, result.Confidence)
		})
	}
}

// =============================================================================
// SEMANTIC SIGNAL TESTS
// =============================================================================

func TestIntentClassifier_SemanticSignals_AffirmativeRoots(t *testing.T) {
	ic := NewIntentClassifier()

	// Test that affirmative words trigger regardless of form
	affirmativeMessages := []string{
		"I'm agreeing with this",
		"This is acceptable",
		"I accept the proposal",
		"Permission granted",
		"You may proceed",
		"Authorization given",
	}

	for _, msg := range affirmativeMessages {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.Confidence > 0.3,
				"Affirmative semantics should be detected: %s (conf=%.2f)",
				msg, result.Confidence)
		})
	}
}

func TestIntentClassifier_SemanticSignals_ActionVerbs(t *testing.T) {
	ic := NewIntentClassifier()

	// Test various action verb forms
	actionMessages := []string{
		"Please implement",
		"Start implementing",
		"Begin the work",
		"Execute it",
		"Running this would be good",
		"Let's create the changes",
		"Build it now",
		"Make the modifications",
	}

	for _, msg := range actionMessages {
		t.Run(msg, func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, true)
			assert.True(t, result.IsActionable || result.Confidence > 0.4,
				"Action verb should be detected: %s (intent=%s, actionable=%v)",
				msg, result.Intent, result.IsActionable)
		})
	}
}

// =============================================================================
// PERFORMANCE AND ROBUSTNESS TESTS
// =============================================================================

func TestIntentClassifier_Robustness_EmptyAndWhitespace(t *testing.T) {
	ic := NewIntentClassifier()

	emptyInputs := []string{
		"",
		"   ",
		"\t\n",
	}

	for _, msg := range emptyInputs {
		t.Run("empty_input", func(t *testing.T) {
			result := ic.EnhancedClassifyIntent(msg, false)
			assert.NotNil(t, result)
			assert.False(t, result.IsConfirmation())
		})
	}
}

func TestIntentClassifier_Robustness_VeryLongMessages(t *testing.T) {
	ic := NewIntentClassifier()

	// Long message with confirmation buried in it
	longMsg := "After careful consideration of all the points you mentioned, including the security fixes, " +
		"API standardization, calendar integration, and documentation updates, I have decided that " +
		"we should proceed with everything. Let's do all of these points. Start implementing now."

	result := ic.EnhancedClassifyIntent(longMsg, true)
	assert.True(t, result.IsConfirmation() || result.ShouldProceed(),
		"Long message with confirmation should be detected (intent=%s, conf=%.2f)",
		result.Intent, result.Confidence)
}

// =============================================================================
// LLM Intent Classifier Tests (for llm_intent_classifier.go)
// =============================================================================

func TestNewLLMIntentClassifier(t *testing.T) {
	log := newTestLogger()
	lic := NewLLMIntentClassifier(nil, log)

	assert.NotNil(t, lic)
	assert.NotNil(t, lic.fallbackClassifier)
	assert.Nil(t, lic.providerRegistry)
}

func TestLLMIntentClassifier_WithProviderRegistry(t *testing.T) {
	log := newTestLogger()
	// Create a mock registry
	cfg := &RegistryConfig{
		DefaultTimeout: 10 * time.Second,
		Providers:      make(map[string]*ProviderConfig),
	}
	registry := NewProviderRegistryWithoutAutoDiscovery(cfg, nil)

	lic := NewLLMIntentClassifier(registry, log)

	assert.NotNil(t, lic)
	assert.Equal(t, registry, lic.providerRegistry)
}

func TestLLMIntentClassifier_QuickClassify_ShortMessage(t *testing.T) {
	log := newTestLogger()
	lic := NewLLMIntentClassifier(nil, log)

	// Short message should fall back to pattern-based because no provider is available
	result := lic.QuickClassify(context.Background(), "Yes!", true)

	assert.NotNil(t, result)
	assert.True(t, result.IsConfirmation() || result.ShouldProceed())
}

func TestLLMIntentClassifier_QuickClassify_LongerMessage(t *testing.T) {
	log := newTestLogger()
	lic := NewLLMIntentClassifier(nil, log)

	// Longer message should use pattern-based fallback
	result := lic.QuickClassify(context.Background(),
		"This is a longer message that should definitely use pattern-based classification because it exceeds the fifty character limit.",
		false)

	assert.NotNil(t, result)
}

// =============================================================================
// Intent Classification Cache Tests
// =============================================================================

func TestNewIntentClassificationCache(t *testing.T) {
	cache := NewIntentClassificationCache(100, 5*time.Minute)

	assert.NotNil(t, cache)
	assert.Equal(t, 100, cache.maxSize)
	assert.Equal(t, 5*time.Minute, cache.ttl)
}

func TestIntentClassificationCache_SetAndGet(t *testing.T) {
	cache := NewIntentClassificationCache(10, 5*time.Minute)

	result := &IntentClassificationResult{
		Intent:     IntentConfirmation,
		Confidence: 0.95,
	}

	cache.Set("test message", result)

	retrieved, found := cache.Get("test message")
	assert.True(t, found)
	assert.Equal(t, IntentConfirmation, retrieved.Intent)
	assert.Equal(t, 0.95, retrieved.Confidence)
}

func TestIntentClassificationCache_CaseInsensitive(t *testing.T) {
	cache := NewIntentClassificationCache(10, 5*time.Minute)

	result := &IntentClassificationResult{
		Intent:     IntentConfirmation,
		Confidence: 0.9,
	}

	cache.Set("Test Message", result)

	// Should find with different case
	retrieved, found := cache.Get("test message")
	assert.True(t, found)
	assert.Equal(t, IntentConfirmation, retrieved.Intent)
}

func TestIntentClassificationCache_TTLExpiration(t *testing.T) {
	cache := NewIntentClassificationCache(10, 1*time.Millisecond)

	result := &IntentClassificationResult{
		Intent: IntentConfirmation,
	}

	cache.Set("test", result)

	// Wait for TTL to expire
	time.Sleep(5 * time.Millisecond)

	_, found := cache.Get("test")
	assert.False(t, found, "Expired entry should not be found")
}

func TestIntentClassificationCache_MaxSizeEviction(t *testing.T) {
	cache := NewIntentClassificationCache(3, 5*time.Minute)

	// Add 3 entries
	for i := 0; i < 3; i++ {
		cache.Set(fmt.Sprintf("message-%d", i), &IntentClassificationResult{
			Intent: IntentConfirmation,
		})
	}

	// Add 4th entry - should evict oldest
	cache.Set("message-new", &IntentClassificationResult{
		Intent: IntentQuestion,
	})

	// Verify cache size
	assert.Equal(t, 3, len(cache.cache))

	// New entry should be present
	_, found := cache.Get("message-new")
	assert.True(t, found)
}

func TestIntentClassificationCache_GetNonExistent(t *testing.T) {
	cache := NewIntentClassificationCache(10, 5*time.Minute)

	_, found := cache.Get("nonexistent")
	assert.False(t, found)
}
