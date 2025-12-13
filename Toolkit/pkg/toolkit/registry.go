// Package toolkit provides a registry for managing AI providers.
package toolkit

import (
	"fmt"
	"sync"
)

// ProviderRegistry manages the registration and retrieval of AI providers.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewProviderRegistry creates a new ProviderRegistry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider with the given name.
func (r *ProviderRegistry) Register(name string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Unregister removes a provider from the registry.
func (r *ProviderRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}

// Get retrieves a provider by name.
func (r *ProviderRegistry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}

// List returns a list of all registered provider names.
func (r *ProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// GetAll returns all registered providers.
func (r *ProviderRegistry) GetAll() map[string]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make(map[string]Provider)
	for name, provider := range r.providers {
		providers[name] = provider
	}
	return providers
}

// ProviderFactory defines a function type for creating providers.
type ProviderFactory func(config map[string]interface{}) (Provider, error)

// ProviderFactoryRegistry manages provider factories.
type ProviderFactoryRegistry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

// NewProviderFactoryRegistry creates a new ProviderFactoryRegistry.
func NewProviderFactoryRegistry() *ProviderFactoryRegistry {
	return &ProviderFactoryRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register registers a provider factory.
func (r *ProviderFactoryRegistry) Register(name string, factory ProviderFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("factory for provider %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Create creates a provider using the registered factory.
func (r *ProviderFactoryRegistry) Create(name string, config map[string]interface{}) (Provider, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no factory registered for provider %s", name)
	}

	return factory(config)
}

// List returns a list of all registered factory names.
func (r *ProviderFactoryRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}
