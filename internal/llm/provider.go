package llm

import "github.com/superagent/superagent/internal/models"

// LLMProvider defines an interface for LLM providers to integrate with the facade.
type LLMProvider interface {
	Complete(req *models.LLMRequest) (*models.LLMResponse, error)
	HealthCheck() error
	GetCapabilities() *ProviderCapabilities
	ValidateConfig(config map[string]interface{}) (bool, []string)
}
