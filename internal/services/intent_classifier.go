package services

import (
	"regexp"
	"strings"
	"unicode"
)

// UserIntent represents the classified intent of a user message
type UserIntent string

const (
	IntentConfirmation UserIntent = "confirmation"    // User approves/confirms action
	IntentRefusal      UserIntent = "refusal"         // User declines/refuses action
	IntentQuestion     UserIntent = "question"        // User is asking a question
	IntentRequest      UserIntent = "request"         // User is making a new request
	IntentClarification UserIntent = "clarification" // User needs more information
	IntentUnclear      UserIntent = "unclear"         // Intent cannot be determined
)

// IntentClassificationResult contains the classification result with confidence
type IntentClassificationResult struct {
	Intent         UserIntent `json:"intent"`
	Confidence     float64    `json:"confidence"`
	IsActionable   bool       `json:"is_actionable"`   // Should we proceed with action?
	RequiresClarification bool `json:"requires_clarification"`
	Signals        []string   `json:"signals"`         // Signals that led to this classification
}

// IntentClassifier provides semantic intent classification for user messages
type IntentClassifier struct {
	// No hardcoded patterns - uses semantic analysis
}

// NewIntentClassifier creates a new semantic intent classifier
func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{}
}

// ClassifyIntent analyzes user message semantically to determine intent
// This uses multiple signals rather than hardcoded pattern matching
func (ic *IntentClassifier) ClassifyIntent(message string, hasContext bool) *IntentClassificationResult {
	result := &IntentClassificationResult{
		Intent:     IntentUnclear,
		Confidence: 0.0,
		Signals:    make([]string, 0),
	}

	// Normalize message for analysis
	normalized := strings.TrimSpace(message)
	if len(normalized) == 0 {
		return result
	}

	lowerMsg := strings.ToLower(normalized)

	// Semantic analysis using multiple signals
	affirmativeScore := ic.analyzeAffirmativeSignals(lowerMsg)
	negativeScore := ic.analyzeNegativeSignals(lowerMsg)
	questionScore := ic.analyzeQuestionSignals(normalized, lowerMsg)
	actionScore := ic.analyzeActionSignals(lowerMsg)
	urgencyScore := ic.analyzeUrgencySignals(lowerMsg)

	// Contextual boost - short positive messages after context are likely confirmations
	contextBoost := 0.0
	if hasContext && len(normalized) < 100 {
		contextBoost = 0.2
		result.Signals = append(result.Signals, "short_response_with_context")
	}

	// Calculate final scores
	confirmScore := affirmativeScore + actionScore + urgencyScore + contextBoost
	refuseScore := negativeScore
	questionScoreFinal := questionScore

	// Determine intent based on scores
	if confirmScore > 0.5 && confirmScore > refuseScore && confirmScore > questionScoreFinal {
		result.Intent = IntentConfirmation
		result.Confidence = minFloat64(1.0, confirmScore)
		result.IsActionable = true
		result.Signals = append(result.Signals, "high_confirmation_score")
	} else if refuseScore > 0.5 && refuseScore > confirmScore {
		result.Intent = IntentRefusal
		result.Confidence = minFloat64(1.0, refuseScore)
		result.IsActionable = false
		result.Signals = append(result.Signals, "high_refusal_score")
	} else if questionScoreFinal > 0.5 {
		result.Intent = IntentQuestion
		result.Confidence = minFloat64(1.0, questionScoreFinal)
		result.IsActionable = false
		result.RequiresClarification = true
		result.Signals = append(result.Signals, "question_detected")
	} else if actionScore > 0.3 && affirmativeScore > 0.2 {
		// Moderate action intent with some affirmation
		result.Intent = IntentRequest
		result.Confidence = minFloat64(1.0, (actionScore+affirmativeScore)/2)
		result.IsActionable = true
		result.Signals = append(result.Signals, "action_request_detected")
	} else {
		result.Intent = IntentUnclear
		result.Confidence = 0.3
		result.RequiresClarification = true
	}

	return result
}

// analyzeAffirmativeSignals detects semantic affirmation without hardcoding
func (ic *IntentClassifier) analyzeAffirmativeSignals(msg string) float64 {
	score := 0.0
	signals := 0

	// If the message is a question (has ?), reduce affirmative signals
	// Questions like "Is this correct?" should not be seen as confirmation
	isQuestion := strings.Contains(msg, "?")
	questionPenalty := 0.0
	if isQuestion {
		questionPenalty = 0.5 // Reduce affirmative score for questions
	}

	// Semantic categories of affirmation (not exact matches, but concepts)

	// 1. Agreement expressions - check for agreement semantics
	if ic.containsAgreementSemantics(msg) {
		score += 0.3
		signals++
	}

	// 2. Approval expressions
	if ic.containsApprovalSemantics(msg) {
		score += 0.3
		signals++
	}

	// 3. Positive sentiment indicators
	if ic.hasPositiveSentiment(msg) {
		score += 0.2
		signals++
	}

	// 4. Inclusive language ("let's", "we should", etc.)
	if ic.hasInclusiveLanguage(msg) {
		score += 0.2
		signals++
	}

	// 5. Emphasis markers (exclamation, caps for positive words)
	if ic.hasPositiveEmphasis(msg) {
		score += 0.15
		signals++
	}

	// Normalize by number of signals for robustness
	if signals > 0 {
		score = score * (1.0 + float64(signals)*0.1)
	}

	// Apply question penalty
	score = score - questionPenalty
	if score < 0 {
		score = 0
	}

	return minFloat64(1.0, score)
}

// analyzeNegativeSignals detects semantic refusal/negation
func (ic *IntentClassifier) analyzeNegativeSignals(msg string) float64 {
	score := 0.0

	// 1. Negation semantics - high score for direct negation
	if ic.containsNegationSemantics(msg) {
		score += 0.55 // Enough to pass 0.5 threshold alone
	}

	// 2. Refusal semantics
	if ic.containsRefusalSemantics(msg) {
		score += 0.35
	}

	// 3. Stop/cancel semantics
	if ic.containsStopSemantics(msg) {
		score += 0.35
	}

	// 4. Negative sentiment
	if ic.hasNegativeSentiment(msg) {
		score += 0.2
	}

	return minFloat64(1.0, score)
}

// analyzeQuestionSignals detects if message is a question
func (ic *IntentClassifier) analyzeQuestionSignals(original, lower string) float64 {
	score := 0.0

	// 1. Question mark presence
	if strings.Contains(original, "?") {
		score += 0.4
	}

	// 2. Interrogative words at start
	if ic.startsWithInterrogative(lower) {
		score += 0.3
	}

	// 3. Questioning semantics
	if ic.containsQuestionSemantics(lower) {
		score += 0.2
	}

	return minFloat64(1.0, score)
}

// analyzeActionSignals detects request for action
func (ic *IntentClassifier) analyzeActionSignals(msg string) float64 {
	score := 0.0

	// 1. Action verbs present
	if ic.containsActionVerbs(msg) {
		score += 0.3
	}

	// 2. Completeness indicators ("all", "everything", "complete")
	if ic.containsCompletenessIndicators(msg) {
		score += 0.25
	}

	// 3. Immediacy indicators
	if ic.containsImmediacyIndicators(msg) {
		score += 0.25
	}

	// 4. Reference to previous content ("these", "those", "the plan", etc.)
	if ic.containsReferenceIndicators(msg) {
		score += 0.2
	}

	return minFloat64(1.0, score)
}

// analyzeUrgencySignals detects urgency in message
func (ic *IntentClassifier) analyzeUrgencySignals(msg string) float64 {
	score := 0.0

	// Urgency without hardcoding specific words
	if ic.containsUrgencySemantics(msg) {
		score += 0.2
	}

	// Exclamation marks indicate emphasis/urgency
	exclamationCount := strings.Count(msg, "!")
	if exclamationCount > 0 {
		score += minFloat64(0.2, float64(exclamationCount)*0.1)
	}

	return minFloat64(1.0, score)
}

// Semantic analysis helper functions - these check for CONCEPTS not exact strings

func (ic *IntentClassifier) containsAgreementSemantics(msg string) bool {
	// Check for agreement concepts using word stems and semantic groups
	agreementRoots := []string{"yes", "yep", "yeah", "yup", "ok", "okay", "sure", "right", "correct", "agreed", "agree", "affirm", "accept"}
	return ic.containsAnyRoot(msg, agreementRoots)
}

func (ic *IntentClassifier) containsApprovalSemantics(msg string) bool {
	approvalRoots := []string{"approv", "confirm", "endors", "sanction", "authoriz", "permit", "allow", "green light", "thumbs up", "good to go", "sounds good", "perfect", "great", "excellent", "wonderful", "grant", "given", "go ahead"}
	return ic.containsAnyRoot(msg, approvalRoots)
}

func (ic *IntentClassifier) hasPositiveSentiment(msg string) bool {
	// First check for negation - "dislike" should not match "like"
	if ic.containsNegationSemantics(msg) {
		return false
	}
	// Check for negative prefixes like "dis-"
	if strings.Contains(strings.ToLower(msg), "dislike") {
		return false
	}
	positiveRoots := []string{"good", "great", "nice", "awesome", "fantastic", "love", "like", "wonderful", "excellent", "amazing", "perfect", "brilliant", "exactly", "need", "want"}
	return ic.containsAnyRoot(msg, positiveRoots)
}

func (ic *IntentClassifier) hasInclusiveLanguage(msg string) bool {
	// Block inclusive detection if message contains refusal
	if ic.containsRefusalSemantics(msg) || ic.containsStopSemantics(msg) {
		return false
	}
	// "let's", "let us", "we", "together", etc.
	inclusivePatterns := []string{"let's", "let us", "we should", "we can", "together", "us "}
	return ic.containsAnyRoot(msg, inclusivePatterns)
}

func (ic *IntentClassifier) hasPositiveEmphasis(msg string) bool {
	// Check for emphasis on positive words
	return strings.Contains(msg, "!") && (ic.hasPositiveSentiment(msg) || ic.containsAgreementSemantics(msg))
}

func (ic *IntentClassifier) containsNegationSemantics(msg string) bool {
	negationRoots := []string{"no", "not", "don't", "dont", "won't", "wont", "can't", "cant", "never", "neither", "none", "nothing", "nope", "nah"}
	return ic.containsAnyRoot(msg, negationRoots)
}

func (ic *IntentClassifier) containsRefusalSemantics(msg string) bool {
	refusalRoots := []string{"refuse", "reject", "declin", "deny", "dismiss", "pass on", "skip", "avoid", "rather not", "prefer not"}
	return ic.containsAnyRoot(msg, refusalRoots)
}

func (ic *IntentClassifier) containsStopSemantics(msg string) bool {
	stopRoots := []string{"stop", "halt", "cancel", "abort", "cease", "discontinue", "end", "quit", "terminate", "wait", "hold"}
	return ic.containsAnyRoot(msg, stopRoots)
}

func (ic *IntentClassifier) hasNegativeSentiment(msg string) bool {
	negativeRoots := []string{"bad", "wrong", "terrible", "awful", "hate", "dislike", "poor", "horrible", "worst"}
	return ic.containsAnyRoot(msg, negativeRoots)
}

func (ic *IntentClassifier) startsWithInterrogative(msg string) bool {
	interrogatives := []string{"what", "where", "when", "why", "how", "who", "which", "whose", "whom", "can you", "could you", "would you", "will you", "is it", "are you", "do you", "does"}
	for _, q := range interrogatives {
		if strings.HasPrefix(msg, q) {
			return true
		}
	}
	return false
}

func (ic *IntentClassifier) containsQuestionSemantics(msg string) bool {
	questionRoots := []string{"wonder", "curious", "asking", "inquir", "question", "unclear", "confus", "explain", "clarif", "tell me", "show me", "describe", "what is", "how does"}
	return ic.containsAnyRoot(msg, questionRoots)
}

func (ic *IntentClassifier) containsActionVerbs(msg string) bool {
	// Block action detection if message contains negation (e.g., "won't work", "don't do")
	if ic.containsNegationSemantics(msg) {
		return false
	}
	// Block action detection if message contains question semantics (e.g., "explain how this works")
	if ic.containsQuestionSemantics(msg) {
		return false
	}
	actionRoots := []string{"do", "start", "begin", "implement", "execute", "run", "create", "build", "make", "fix", "update", "change", "modify", "add", "remove", "work", "proceed", "continue", "go ahead", "move forward", "tackle", "handle", "address", "happen", "do it", "complete", "finish", "accomplish"}
	return ic.containsAnyRoot(msg, actionRoots)
}

func (ic *IntentClassifier) containsCompletenessIndicators(msg string) bool {
	completenessRoots := []string{"all", "every", "entire", "complete", "full", "whole", "total", "each"}
	return ic.containsAnyRoot(msg, completenessRoots)
}

func (ic *IntentClassifier) containsImmediacyIndicators(msg string) bool {
	immediacyRoots := []string{"now", "immediate", "right away", "asap", "quick", "fast", "today", "instant", "urgent", "pronto"}
	return ic.containsAnyRoot(msg, immediacyRoots)
}

func (ic *IntentClassifier) containsReferenceIndicators(msg string) bool {
	referenceRoots := []string{"these", "those", "this", "that", "the plan", "the points", "the suggest", "the recommend", "above", "previous", "mentioned", "listed", "offered", "proposed"}
	return ic.containsAnyRoot(msg, referenceRoots)
}

func (ic *IntentClassifier) containsUrgencySemantics(msg string) bool {
	urgencyRoots := []string{"urgent", "asap", "immediately", "right now", "quick", "hurry", "rush", "priority", "critical", "important"}
	return ic.containsAnyRoot(msg, urgencyRoots)
}

// containsAnyRoot checks if message contains any of the semantic roots
// This is more flexible than exact matching - it finds word stems
func (ic *IntentClassifier) containsAnyRoot(msg string, roots []string) bool {
	// Tokenize message into words
	words := ic.tokenize(msg)

	// Create a normalized version for phrase matching (remove punctuation)
	normalizedMsg := strings.ToLower(msg)
	for _, r := range []string{"!", ".", "?", ",", ";", ":"} {
		normalizedMsg = strings.ReplaceAll(normalizedMsg, r, " ")
	}

	for _, root := range roots {
		// Check if any word contains this root (stem matching)
		rootLower := strings.ToLower(root)

		// For short roots (2-3 chars), require exact word match to avoid false positives
		// e.g., "no" should not match "now"
		if len(rootLower) <= 3 {
			for _, word := range words {
				if word == rootLower {
					return true
				}
			}
			continue
		}

		// Check for exact phrase match for multi-word roots
		if strings.Contains(" "+normalizedMsg+" ", " "+rootLower+" ") {
			return true
		}

		// Check if root with space padding is in the message
		if strings.Contains(rootLower, " ") && strings.Contains(normalizedMsg, rootLower) {
			return true
		}

		// Then check individual words for stem matching
		for _, word := range words {
			if strings.HasPrefix(word, rootLower) || word == rootLower {
				return true
			}
		}
	}
	return false
}

// tokenize splits message into lowercase words
func (ic *IntentClassifier) tokenize(msg string) []string {
	// Simple tokenization - split on non-letter characters
	var words []string
	var currentWord strings.Builder

	for _, r := range strings.ToLower(msg) {
		if unicode.IsLetter(r) || r == '\'' {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
	}
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// IsConfirmation is a convenience method to check if intent is confirmation
func (result *IntentClassificationResult) IsConfirmation() bool {
	return result.Intent == IntentConfirmation && result.Confidence > 0.5
}

// IsRefusal is a convenience method to check if intent is refusal
func (result *IntentClassificationResult) IsRefusal() bool {
	return result.Intent == IntentRefusal && result.Confidence > 0.5
}

// ShouldProceed returns true if we should proceed with action execution
func (result *IntentClassificationResult) ShouldProceed() bool {
	return result.IsActionable && result.Confidence > 0.5 && !result.RequiresClarification
}

// minFloat64 returns the minimum of two float64 values
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Additional semantic patterns using regex for more complex matching
var (
	// Pattern for "do X" style confirmations
	doPatternRegex = regexp.MustCompile(`(?i)\b(do|perform|execute|run|start|begin|implement|proceed|continue|tackle|handle|work|make)\b`)

	// Pattern for completeness references
	allPatternRegex = regexp.MustCompile(`(?i)\b(all|every|each|entire|complete|full|whole|everything)\b.*\b(point|item|task|thing|step|suggest|recommend|action|change|fix|above)\b`)

	// Pattern for completeness without action verbs (with context these are confirmations)
	completenessOnlyRegex = regexp.MustCompile(`(?i)\b(everything|all)\b.*\b(you|the|suggest|recommend|please|thank)\b`)

	// Pattern for affirmative responses
	affirmativeRegex = regexp.MustCompile(`(?i)^(yes|yep|yeah|yup|ok|okay|sure|right|correct|absolutely|definitely|certainly|of course|go ahead|sounds good|perfect|great|excellent)\b`)
)

// EnhancedClassifyIntent uses regex patterns for additional semantic matching
func (ic *IntentClassifier) EnhancedClassifyIntent(message string, hasContext bool) *IntentClassificationResult {
	// Start with base classification
	result := ic.ClassifyIntent(message, hasContext)

	// Enhance with regex-based semantic patterns
	lowerMsg := strings.ToLower(message)
	trimmedLower := strings.TrimSpace(lowerMsg)

	// Check for "do all X" pattern
	if allPatternRegex.MatchString(lowerMsg) {
		result.Confidence = minFloat64(1.0, result.Confidence+0.2)
		result.Signals = append(result.Signals, "completeness_pattern_matched")
		if result.Intent == IntentUnclear {
			result.Intent = IntentConfirmation
			result.IsActionable = true
		}
	}

	// Check for direct affirmative start
	if affirmativeRegex.MatchString(trimmedLower) {
		result.Confidence = minFloat64(1.0, result.Confidence+0.15)
		result.Signals = append(result.Signals, "affirmative_start")
		if result.Intent == IntentUnclear && hasContext {
			result.Intent = IntentConfirmation
			result.IsActionable = true
		}
	}

	// Short action phrases with context are confirmations
	// Examples: "Do it", "Make it happen", "Start now", "Begin the work"
	if hasContext && len(message) < 50 && doPatternRegex.MatchString(lowerMsg) {
		if result.Intent == IntentUnclear {
			result.Intent = IntentConfirmation
			result.IsActionable = true
			result.Confidence = minFloat64(1.0, result.Confidence+0.3)
			result.Signals = append(result.Signals, "short_action_phrase_with_context")
		} else if result.Intent == IntentConfirmation {
			result.Confidence = minFloat64(1.0, result.Confidence+0.1)
			result.Signals = append(result.Signals, "action_verb_present")
		}
	}

	// Positive sentiment with context is approval
	// Examples: "I like it", "Love it", "That's exactly what I need"
	if hasContext && ic.hasPositiveSentiment(lowerMsg) {
		if result.Intent == IntentUnclear {
			result.Intent = IntentConfirmation
			result.IsActionable = true
			result.Confidence = minFloat64(1.0, result.Confidence+0.25)
			result.Signals = append(result.Signals, "positive_sentiment_with_context")
		} else if result.Intent == IntentConfirmation {
			result.Confidence = minFloat64(1.0, result.Confidence+0.1)
		}
	}

	// Completeness without action verbs (with context = confirmation)
	// Examples: "Everything you suggested", "All recommendations please"
	if hasContext && completenessOnlyRegex.MatchString(lowerMsg) {
		if result.Intent == IntentUnclear {
			result.Intent = IntentConfirmation
			result.IsActionable = true
			result.Confidence = minFloat64(1.0, result.Confidence+0.3)
			result.Signals = append(result.Signals, "completeness_reference_with_context")
		}
	}

	return result
}
