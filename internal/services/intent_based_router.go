package services

import (
	"regexp"
	"strings"

	"dev.helix.agent/internal/verifier"
	"github.com/sirupsen/logrus"
)

type RoutingDecision string

const (
	RoutingEnsemble   RoutingDecision = "ensemble"
	RoutingSingle     RoutingDecision = "single"
	RoutingAutoSelect RoutingDecision = "auto"
)

type RoutingResult struct {
	Decision          RoutingDecision             `json:"decision"`
	ShouldUseEnsemble bool                        `json:"should_use_ensemble"`
	Reason            string                      `json:"reason"`
	Confidence        float64                     `json:"confidence"`
	Signals           []string                    `json:"signals"`
	StrongestProvider *verifier.UnifiedProvider   `json:"strongest_provider,omitempty"`
	FallbackProviders []*verifier.UnifiedProvider `json:"fallback_providers,omitempty"`
}

type IntentBasedRouter struct {
	startupVerifier *verifier.StartupVerifier
	logger          *logrus.Logger

	simplePatterns   []*regexp.Regexp
	complexPatterns  []*regexp.Regexp
	greetingPatterns []*regexp.Regexp
}

func NewIntentBasedRouter(startupVerifier *verifier.StartupVerifier, logger *logrus.Logger) *IntentBasedRouter {
	router := &IntentBasedRouter{
		startupVerifier: startupVerifier,
		logger:          logger,
	}

	router.initPatterns()
	return router
}

func (r *IntentBasedRouter) initPatterns() {
	r.greetingPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(hi|hello|hey|greetings|good\s*(morning|afternoon|evening)|howdy|yo|sup)\b`),
		regexp.MustCompile(`(?i)^(what'?s?\s*up|how\s*are\s*you|how\s*is\s*it\s*going)\b`),
		regexp.MustCompile(`(?i)^(thanks|thank\s*you|thx|cheers)\b`),
	}

	r.simplePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(yes|no|ok|okay|sure|right|correct|exactly|perfect)\b\.?$`),
		regexp.MustCompile(`(?i)^(do\s*you\s*(see|understand|know|get|can)\b.*\?$)`),
		regexp.MustCompile(`(?i)^(are\s*you\s*(ready|available|there|online)\b.*\?$)`),
		regexp.MustCompile(`(?i)^(what\s*(is|are|can|do)\s*(this|that|your)\b.*\?$)`),
		regexp.MustCompile(`(?i)^(show\s*me|list|tell\s*me\s*(about|the|what))\b`),
		regexp.MustCompile(`(?i)^(confirm|continue|proceed|next|done|stop)\b\.?$`),
		regexp.MustCompile(`(?i)^(\?\s*)$`),
		regexp.MustCompile(`(?i)^(how\s*do\s*I|where\s*is|when\s*should)\b.*\?$`),
		regexp.MustCompile(`(?i)^(is\s*(this|that|it)\s*(correct|right|working)\b.*\?$)`),
		regexp.MustCompile(`(?i)^(can\s*you\s*(help|explain|clarify)\b.*\?$)`),
	}

	r.complexPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(debug|debugging|fix\s*(this|the|a)|investigate|troubleshoot)\b`),
		regexp.MustCompile(`(?i)\b(refactor|restructure|rewrite|migrate|port)\b`),
		regexp.MustCompile(`(?i)\b(implement|build|create|develop|design)\s+(a|an|the|this|new|feature|system)\b`),
		regexp.MustCompile(`(?i)\b(analyze|analysis|review|audit|assess)\s+(the|this|code|system)\b`),
		regexp.MustCompile(`(?i)\b(improve|optimize|enhance|performance)\b`),
		regexp.MustCompile(`(?i)\b(architect|architecture|design\s*pattern)\b`),
		regexp.MustCompile(`(?i)\b(solve|solution|resolve)\s+(this|the|problem|issue)\b`),
		regexp.MustCompile(`(?i)\b(compare|contrast|evaluate|assess)\s+(.*\bvs\b|options|alternatives)\b`),
		regexp.MustCompile(`(?i)\b(write|generate|produce)\s+.*\b(code|tests?|documentation|implementation)\b`),
		regexp.MustCompile(`(?i)\b(explain|describe)\s+(how|why|what|in\s*detail|thoroughly)\b`),
		regexp.MustCompile(`(?i)\b(find|search|locate|identify)\s+(the|all|any)\s*(bug|issue|error|problem)\b`),
		regexp.MustCompile(`(?i)\b(add|remove|update|modify|change)\s+(feature|functionality|module|component)\b`),
		regexp.MustCompile(`(?i)\b(discuss|debate|argue|deliberate)\b`),
		regexp.MustCompile(`(?i)\b(complex|complicated|intricate|sophisticated)\b`),
		regexp.MustCompile(`(?i)\b(multi|multiple|several|various)\s*(file|module|component|system)\b`),
		regexp.MustCompile(`(?i)\bcomprehensive\s+\w+\s+(tests?|code|review|analysis)\b`),
	}
}

func (r *IntentBasedRouter) AnalyzeRequest(message string, context map[string]interface{}) *RoutingResult {
	result := &RoutingResult{
		Signals:    make([]string, 0),
		Confidence: 0.5,
	}

	trimmedMsg := strings.TrimSpace(message)
	lowerMsg := strings.ToLower(trimmedMsg)

	if r.isGreeting(lowerMsg) {
		result.Decision = RoutingSingle
		result.ShouldUseEnsemble = false
		result.Reason = "greeting_or_social_message"
		result.Confidence = 0.95
		result.Signals = append(result.Signals, "greeting_detected")
		result.StrongestProvider = r.getStrongestProvider()
		return result
	}

	if r.isSimpleConfirmation(lowerMsg) {
		result.Decision = RoutingSingle
		result.ShouldUseEnsemble = false
		result.Reason = "simple_confirmation_response"
		result.Confidence = 0.9
		result.Signals = append(result.Signals, "simple_confirmation")
		result.StrongestProvider = r.getStrongestProvider()
		return result
	}

	if r.isSimpleQuery(lowerMsg, trimmedMsg) {
		result.Decision = RoutingSingle
		result.ShouldUseEnsemble = false
		result.Reason = "simple_query_or_question"
		result.Confidence = 0.85
		result.Signals = append(result.Signals, "simple_query")
		result.StrongestProvider = r.getStrongestProvider()
		return result
	}

	if r.isComplexRequest(lowerMsg) {
		result.Decision = RoutingEnsemble
		result.ShouldUseEnsemble = true
		result.Reason = "complex_request_requiring_ensemble"
		result.Confidence = 0.9
		result.Signals = append(result.Signals, "complex_request_detected")
		return result
	}

	if r.hasCodebaseContext(context) {
		result.Decision = RoutingEnsemble
		result.ShouldUseEnsemble = true
		result.Reason = "has_codebase_context_suggests_complex_work"
		result.Confidence = 0.75
		result.Signals = append(result.Signals, "codebase_context_present")
		return result
	}

	if len(trimmedMsg) > 200 {
		result.Decision = RoutingEnsemble
		result.ShouldUseEnsemble = true
		result.Reason = "long_message_suggests_complex_request"
		result.Confidence = 0.7
		result.Signals = append(result.Signals, "long_message")
		return result
	}

	complexityScore := r.calculateComplexityScore(lowerMsg)
	if complexityScore > 0.6 {
		result.Decision = RoutingEnsemble
		result.ShouldUseEnsemble = true
		result.Reason = "high_complexity_score"
		result.Confidence = complexityScore
		result.Signals = append(result.Signals, "complexity_score_high")
		return result
	}

	result.Decision = RoutingSingle
	result.ShouldUseEnsemble = false
	result.Reason = "default_to_single_provider_for_efficiency"
	result.Confidence = 0.6
	result.Signals = append(result.Signals, "default_single")
	result.StrongestProvider = r.getStrongestProvider()
	result.FallbackProviders = r.getFallbackProviders()

	return result
}

func (r *IntentBasedRouter) isGreeting(msg string) bool {
	for _, pattern := range r.greetingPatterns {
		if pattern.MatchString(msg) {
			return true
		}
	}
	return false
}

func (r *IntentBasedRouter) isSimpleConfirmation(msg string) bool {
	cleanMsg := strings.TrimSuffix(strings.TrimSuffix(msg, "."), "!")
	cleanMsg = strings.TrimSpace(cleanMsg)

	simpleConfirmations := []string{
		"yes", "no", "ok", "okay", "sure", "right", "correct",
		"exactly", "perfect", "yep", "nope", "done", "continue",
		"proceed", "next", "go ahead", "sounds good", "great",
	}

	for _, conf := range simpleConfirmations {
		if cleanMsg == conf {
			return true
		}
	}

	return false
}

func (r *IntentBasedRouter) isSimpleQuery(msg, originalMsg string) bool {
	for _, pattern := range r.simplePatterns {
		if pattern.MatchString(originalMsg) {
			return true
		}
	}

	if strings.Count(originalMsg, "?") > 0 && len(originalMsg) < 80 {
		hasComplexKeywords := r.isComplexRequest(msg)
		if !hasComplexKeywords {
			return true
		}
	}

	return false
}

func (r *IntentBasedRouter) isComplexRequest(msg string) bool {
	for _, pattern := range r.complexPatterns {
		if pattern.MatchString(msg) {
			return true
		}
	}

	complexIndicators := []string{
		"debate", "discuss", "multi", "several", "various",
		"comprehensive", "detailed", "thorough", "complete",
		"entire", "whole", "all the", "best approach",
		"pros and cons", "alternatives", "options",
	}

	matches := 0
	for _, indicator := range complexIndicators {
		if strings.Contains(msg, indicator) {
			matches++
		}
	}

	return matches >= 2
}

func (r *IntentBasedRouter) hasCodebaseContext(context map[string]interface{}) bool {
	if context == nil {
		return false
	}

	_, hasFiles := context["files"]
	_, hasCodebase := context["codebase"]
	_, hasProject := context["project"]
	_, hasRepo := context["repository"]

	return hasFiles || hasCodebase || hasProject || hasRepo
}

func (r *IntentBasedRouter) calculateComplexityScore(msg string) float64 {
	score := 0.0

	words := strings.Fields(msg)
	if len(words) > 30 {
		score += 0.2
	} else if len(words) > 20 {
		score += 0.1
	}

	complexWords := []string{
		"however", "therefore", "consequently", "furthermore",
		"additionally", "nevertheless", "specifically", "essentially",
	}

	for _, word := range complexWords {
		if strings.Contains(msg, word) {
			score += 0.1
		}
	}

	questions := strings.Count(msg, "?")
	if questions > 2 {
		score += 0.15
	}

	codeBlockIndicators := []string{"```", "func ", "func(", "class ", "def ", "import ", "package ", "type ", "struct ", "interface {"}
	for _, indicator := range codeBlockIndicators {
		if strings.Contains(msg, indicator) {
			score += 0.7
			break
		}
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (r *IntentBasedRouter) getStrongestProvider() *verifier.UnifiedProvider {
	if r.startupVerifier == nil {
		return nil
	}

	ranked := r.startupVerifier.GetRankedProviders()
	if len(ranked) == 0 {
		return nil
	}

	for _, p := range ranked {
		if p.Verified && p.Status == verifier.StatusHealthy {
			return p
		}
	}

	return nil
}

func (r *IntentBasedRouter) getFallbackProviders() []*verifier.UnifiedProvider {
	if r.startupVerifier == nil {
		return nil
	}

	ranked := r.startupVerifier.GetRankedProviders()
	if len(ranked) <= 1 {
		return nil
	}

	var fallbacks []*verifier.UnifiedProvider
	count := 0
	for _, p := range ranked {
		if p.Verified && p.Status == verifier.StatusHealthy && count < 3 {
			fallbacks = append(fallbacks, p)
			count++
		}
	}

	if len(fallbacks) > 3 {
		fallbacks = fallbacks[:3]
	}

	return fallbacks
}

func (r *IntentBasedRouter) ShouldUseEnsemble(message string, context map[string]interface{}) bool {
	result := r.AnalyzeRequest(message, context)

	r.logger.WithFields(logrus.Fields{
		"decision":   result.Decision,
		"reason":     result.Reason,
		"confidence": result.Confidence,
		"signals":    result.Signals,
	}).Debug("IntentBasedRouter decision")

	return result.ShouldUseEnsemble
}

func (r *IntentBasedRouter) GetRoutingDecision(message string, context map[string]interface{}) RoutingDecision {
	result := r.AnalyzeRequest(message, context)
	return result.Decision
}

func (r *IntentBasedRouter) GetRoutingResult(message string, context map[string]interface{}) *RoutingResult {
	return r.AnalyzeRequest(message, context)
}
