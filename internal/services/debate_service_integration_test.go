package services

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestDebateServiceIntegration_Initialization tests integrated feature initialization
func TestDebateServiceIntegration_Initialization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Quiet logs during tests

	t.Run("InitializesAllFeatures", func(t *testing.T) {
		registry := NewProviderRegistry(nil, nil)
		service := NewDebateServiceWithDeps(logger, registry, nil)

		assert.NotNil(t, service.testGenerator, "Test generator should be initialized")
		assert.NotNil(t, service.testExecutor, "Test executor should be initialized")
		assert.NotNil(t, service.contrastiveAnalyzer, "Contrastive analyzer should be initialized")
		assert.NotNil(t, service.validationPipeline, "Validation pipeline should be initialized")
		assert.NotNil(t, service.toolIntegration, "Tool integration should be initialized")
		assert.NotNil(t, service.serviceBridge, "Service bridge should be initialized")
	})
}

// TestDebateServiceIntegration_CodeDetection tests code generation intent detection
func TestDebateServiceIntegration_CodeDetection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	registry := NewProviderRegistry(nil, nil)
	service := NewDebateServiceWithDeps(logger, registry, nil)

	tests := []struct {
		name     string
		topic    string
		expected bool
	}{
		{"Code_Write", "Write a function to calculate fibonacci", true},
		{"Code_Implement", "Implement a binary search algorithm", true},
		{"Code_Create", "Create a REST API endpoint", true},
		{"Code_Develop", "Develop a sorting function", true},
		{"Code_Build", "Build a web scraper", true},
		{"NonCode_Explain", "Explain how recursion works", false},
		{"NonCode_Discuss", "Discuss the benefits of functional programming", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectCodeGenerationIntent(tt.topic, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDebateServiceIntegration_LanguageDetection tests programming language detection
func TestDebateServiceIntegration_LanguageDetection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	registry := NewProviderRegistry(nil, nil)
	service := NewDebateServiceWithDeps(logger, registry, nil)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"Python_Def", "def hello():\\n    print('Hello')", "python"},
		{"JavaScript_Function", "function test() { return 42; }", "javascript"},
		{"Go_Package", "package main\\nfunc main() {}", "go"},
		{"Java_Class", "public class Test { }", "java"},
		{"Rust_Fn", "fn main() { println!(\"Hello\"); }", "rust"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectLanguage(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDebateServiceIntegration_RoleSelection tests specialized role selection
func TestDebateServiceIntegration_RoleSelection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	registry := NewProviderRegistry(nil, nil)
	service := NewDebateServiceWithDeps(logger, registry, nil)

	tests := []struct {
		name     string
		topic    string
		expected string
	}{
		{"Generator_Write", "Write a new authentication module", "generator"},
		{"Refactorer_Improve", "Refactor the database layer", "refactorer"},
		{"PerformanceAnalyzer_Optimize", "Optimize the API response time", "performance_analyzer"},
		{"SecurityAnalyst_Audit", "Security audit of the authentication flow", "security_analyst"},
		{"Debugger_Fix", "Fix the login bug", "debugger"},
		{"Architect_Design", "Design a microservices architecture", "architect"},
		{"Reviewer_Review", "Review the pull request", "reviewer"},
		{"Tester_Test", "Write unit tests for the service", "tester"},
		{"Default_Generic", "Explain how the system works", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.selectSpecializedRole(context.Background(), tt.topic, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDebateServiceIntegration_Components tests that all components are initialized
func TestDebateServiceIntegration_Components(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	registry := NewProviderRegistry(nil, nil)
	service := NewDebateServiceWithDeps(logger, registry, nil)

	t.Run("TestGenerator", func(t *testing.T) {
		assert.NotNil(t, service.testGenerator)
	})

	t.Run("TestExecutor", func(t *testing.T) {
		assert.NotNil(t, service.testExecutor)
	})

	t.Run("ContrastiveAnalyzer", func(t *testing.T) {
		assert.NotNil(t, service.contrastiveAnalyzer)
	})

	t.Run("ValidationPipeline", func(t *testing.T) {
		assert.NotNil(t, service.validationPipeline)
	})

	t.Run("ToolIntegration", func(t *testing.T) {
		assert.NotNil(t, service.toolIntegration)
	})

	t.Run("ServiceBridge", func(t *testing.T) {
		assert.NotNil(t, service.serviceBridge)
	})
}
