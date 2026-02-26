package services

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIntentBasedRouter_Greeting(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	tests := []struct {
		name     string
		message  string
		expected RoutingDecision
	}{
		{"hello", "hello", RoutingSingle},
		{"hi there", "hi there", RoutingSingle},
		{"hey", "hey", RoutingSingle},
		{"good morning", "good morning", RoutingSingle},
		{"how are you", "how are you", RoutingSingle},
		{"thanks", "thanks", RoutingSingle},
		{"thank you", "thank you", RoutingSingle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.AnalyzeRequest(tt.message, nil)
			assert.Equal(t, tt.expected, result.Decision)
			assert.False(t, result.ShouldUseEnsemble)
			assert.Contains(t, result.Reason, "greeting")
		})
	}
}

func TestIntentBasedRouter_SimpleConfirmation(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	tests := []struct {
		name     string
		message  string
		expected RoutingDecision
	}{
		{"yes", "yes", RoutingSingle},
		{"no", "no", RoutingSingle},
		{"ok", "ok", RoutingSingle},
		{"okay", "okay", RoutingSingle},
		{"sure", "sure", RoutingSingle},
		{"right", "right", RoutingSingle},
		{"correct", "correct", RoutingSingle},
		{"exactly", "exactly", RoutingSingle},
		{"perfect", "perfect", RoutingSingle},
		{"done", "done", RoutingSingle},
		{"continue", "continue", RoutingSingle},
		{"proceed", "proceed", RoutingSingle},
		{"next", "next", RoutingSingle},
		{"go ahead", "go ahead", RoutingSingle},
		{"sounds good", "sounds good", RoutingSingle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.AnalyzeRequest(tt.message, nil)
			assert.Equal(t, tt.expected, result.Decision)
			assert.False(t, result.ShouldUseEnsemble)
			assert.Contains(t, result.Reason, "confirmation")
		})
	}
}

func TestIntentBasedRouter_SimpleQuery(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	tests := []struct {
		name     string
		message  string
		expected RoutingDecision
	}{
		{"do you see", "do you see my codebase?", RoutingSingle},
		{"are you ready", "are you ready?", RoutingSingle},
		{"what is this", "what is this?", RoutingSingle},
		{"how do I", "how do I run tests?", RoutingSingle},
		{"can you help", "can you help me?", RoutingSingle},
		{"is this correct", "is this correct?", RoutingSingle},
		{"short question", "what?", RoutingSingle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.AnalyzeRequest(tt.message, nil)
			assert.Equal(t, tt.expected, result.Decision)
			assert.False(t, result.ShouldUseEnsemble)
		})
	}
}

func TestIntentBasedRouter_ComplexRequest(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	tests := []struct {
		name     string
		message  string
		expected RoutingDecision
	}{
		{"debug", "debug this issue in my code", RoutingEnsemble},
		{"refactor", "refactor this module to use better patterns", RoutingEnsemble},
		{"implement feature", "implement a new authentication system", RoutingEnsemble},
		{"analyze code", "analyze the code for security issues", RoutingEnsemble},
		{"build system", "build a new microservice architecture", RoutingEnsemble},
		{"create module", "create a new module for user management", RoutingEnsemble},
		{"design pattern", "design a pattern for event handling", RoutingEnsemble},
		{"migrate code", "migrate the codebase to Go 1.24", RoutingEnsemble},
		{"investigate", "investigate why the tests are failing", RoutingEnsemble},
		{"optimize", "optimize the database queries", RoutingEnsemble},
		{"improve performance", "improve performance of the API", RoutingEnsemble},
		{"write tests", "write comprehensive tests for this module", RoutingEnsemble},
		{"explain in detail", "explain in detail how this works", RoutingEnsemble},
		{"discuss options", "let's discuss the best approach", RoutingEnsemble},
		{"debate approach", "debate the pros and cons of this approach", RoutingEnsemble},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.AnalyzeRequest(tt.message, nil)
			assert.Equal(t, tt.expected, result.Decision)
			assert.True(t, result.ShouldUseEnsemble)
		})
	}
}

func TestIntentBasedRouter_CodebaseContext(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	context := map[string]interface{}{
		"files":    []string{"main.go", "handler.go"},
		"codebase": "my-project",
	}

	result := router.AnalyzeRequest("look at this", context)
	assert.Equal(t, RoutingEnsemble, result.Decision)
	assert.True(t, result.ShouldUseEnsemble)
	assert.Contains(t, result.Reason, "codebase_context")
}

func TestIntentBasedRouter_LongMessage(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	longMessage := "I need to implement a comprehensive solution that handles multiple edge cases, " +
		"includes proper error handling, integrates with the existing authentication system, " +
		"and provides a clean API for external consumers. The implementation should be thorough."

	result := router.AnalyzeRequest(longMessage, nil)
	assert.Equal(t, RoutingEnsemble, result.Decision)
	assert.True(t, result.ShouldUseEnsemble)
	assert.True(t, strings.Contains(result.Reason, "complex") || strings.Contains(result.Reason, "long"),
		"reason should contain 'complex' or 'long', got: %s", result.Reason)
}

func TestIntentBasedRouter_DefaultSingle(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	mediumMessage := "Please add a log statement to track this"

	result := router.AnalyzeRequest(mediumMessage, nil)
	assert.Equal(t, RoutingSingle, result.Decision)
	assert.False(t, result.ShouldUseEnsemble)
	assert.Contains(t, result.Reason, "default")
}

func TestIntentBasedRouter_ComplexityScore(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	score := router.calculateComplexityScore("simple message")
	assert.LessOrEqual(t, score, 0.3)

	score = router.calculateComplexityScore("however, furthermore, additionally, this is complex")
	assert.Greater(t, score, 0.3)

	score = router.calculateComplexityScore("func main() { fmt.Println(`code`) }")
	assert.Greater(t, score, 0.1)
}

func TestIntentBasedRouter_ShouldUseEnsemble(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	assert.False(t, router.ShouldUseEnsemble("hello", nil))
	assert.False(t, router.ShouldUseEnsemble("yes", nil))
	assert.False(t, router.ShouldUseEnsemble("do you see my code?", nil))
	assert.True(t, router.ShouldUseEnsemble("debug this issue", nil))
	assert.True(t, router.ShouldUseEnsemble("refactor this module", nil))
}

func TestIntentBasedRouter_GetRoutingDecision(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	assert.Equal(t, RoutingSingle, router.GetRoutingDecision("hello", nil))
	assert.Equal(t, RoutingEnsemble, router.GetRoutingDecision("implement this feature", nil))
}

func TestIntentBasedRouter_GetRoutingResult(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	result := router.GetRoutingResult("debug this", nil)
	assert.NotNil(t, result)
	assert.Equal(t, RoutingEnsemble, result.Decision)
	assert.True(t, result.ShouldUseEnsemble)
	assert.NotEmpty(t, result.Reason)
	assert.NotEmpty(t, result.Signals)
	assert.Greater(t, result.Confidence, 0.0)
}

func TestIntentBasedRouter_EdgeCases(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	t.Run("empty message", func(t *testing.T) {
		result := router.AnalyzeRequest("", nil)
		assert.Equal(t, RoutingSingle, result.Decision)
	})

	t.Run("whitespace only", func(t *testing.T) {
		result := router.AnalyzeRequest("   ", nil)
		assert.Equal(t, RoutingSingle, result.Decision)
	})

	t.Run("mixed case greeting", func(t *testing.T) {
		result := router.AnalyzeRequest("HeLlO", nil)
		assert.Equal(t, RoutingSingle, result.Decision)
		assert.False(t, result.ShouldUseEnsemble)
	})

	t.Run("punctuation in greeting", func(t *testing.T) {
		result := router.AnalyzeRequest("hello!", nil)
		assert.Equal(t, RoutingSingle, result.Decision)
		assert.False(t, result.ShouldUseEnsemble)
	})

	t.Run("confirmation with punctuation", func(t *testing.T) {
		result := router.AnalyzeRequest("yes.", nil)
		assert.Equal(t, RoutingSingle, result.Decision)
		assert.False(t, result.ShouldUseEnsemble)
	})
}

func TestIntentBasedRouter_MultipleComplexIndicators(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	message := "I need a comprehensive analysis of the entire system to discuss the best approach"
	result := router.AnalyzeRequest(message, nil)
	assert.Equal(t, RoutingEnsemble, result.Decision)
	assert.True(t, result.ShouldUseEnsemble)
}

func TestIntentBasedRouter_ContextVariations(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	t.Run("nil context", func(t *testing.T) {
		result := router.AnalyzeRequest("check this", nil)
		assert.Equal(t, RoutingSingle, result.Decision)
	})

	t.Run("empty context", func(t *testing.T) {
		result := router.AnalyzeRequest("check this", map[string]interface{}{})
		assert.Equal(t, RoutingSingle, result.Decision)
	})

	t.Run("context with project", func(t *testing.T) {
		result := router.AnalyzeRequest("check this", map[string]interface{}{
			"project": "my-project",
		})
		assert.Equal(t, RoutingEnsemble, result.Decision)
	})

	t.Run("context with repository", func(t *testing.T) {
		result := router.AnalyzeRequest("check this", map[string]interface{}{
			"repository": "github.com/org/repo",
		})
		assert.Equal(t, RoutingEnsemble, result.Decision)
	})
}

func TestIntentBasedRouter_CodeBlockDetection(t *testing.T) {
	logger := logrus.New()
	router := NewIntentBasedRouter(nil, logger)

	t.Run("go code", func(t *testing.T) {
		message := "Here is some code: func main() {} that needs review"
		result := router.AnalyzeRequest(message, nil)
		assert.Equal(t, RoutingEnsemble, result.Decision)
	})

	t.Run("python code", func(t *testing.T) {
		message := "Check this: def hello(): pass"
		result := router.AnalyzeRequest(message, nil)
		assert.Equal(t, RoutingEnsemble, result.Decision)
	})

	t.Run("code block", func(t *testing.T) {
		message := "```go\nfunc main() {}\n```"
		result := router.AnalyzeRequest(message, nil)
		assert.Equal(t, RoutingEnsemble, result.Decision)
	})
}
