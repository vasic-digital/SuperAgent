// Package plugins provides adapters between HelixAgent's internal/plugins types
// and the extracted digital.vasic.plugins module.
package plugins

import (
	"context"
	"fmt"

	helixplugins "dev.helix.agent/internal/plugins"
	"dev.helix.agent/internal/models"
	modplugin "digital.vasic.plugins/pkg/plugin"
	modregistry "digital.vasic.plugins/pkg/registry"
	modloader "digital.vasic.plugins/pkg/loader"
	modstructured "digital.vasic.plugins/pkg/structured"
)

// RegistryAdapter adapts the module's Registry to HelixAgent's PluginRegistry.
type RegistryAdapter struct {
	registry *modregistry.Registry
}

// NewRegistryAdapter creates a new registry adapter.
func NewRegistryAdapter() *RegistryAdapter {
	return &RegistryAdapter{
		registry: modregistry.New(),
	}
}

// Register registers a plugin wrapped with the adapter.
func (a *RegistryAdapter) Register(plugin helixplugins.LLMPlugin) error {
	wrapped := &pluginWrapper{helix: plugin}
	return a.registry.Register(wrapped)
}

// Unregister removes a plugin by name.
func (a *RegistryAdapter) Unregister(name string) error {
	return a.registry.Remove(name)
}

// Get retrieves a plugin by name.
func (a *RegistryAdapter) Get(name string) (helixplugins.LLMPlugin, bool) {
	p, ok := a.registry.Get(name)
	if !ok {
		return nil, false
	}
	// Unwrap if it's our wrapper
	if wrapper, isWrapper := p.(*pluginWrapper); isWrapper {
		return wrapper.helix, true
	}
	// If it's a native module plugin, wrap it for HelixAgent
	return &helixPluginAdapter{mod: p}, true
}

// List returns all registered plugin names.
func (a *RegistryAdapter) List() []string {
	return a.registry.List()
}

// StartAll starts all plugins in dependency order.
func (a *RegistryAdapter) StartAll(ctx context.Context) error {
	return a.registry.StartAll(ctx)
}

// StopAll stops all plugins in reverse dependency order.
func (a *RegistryAdapter) StopAll(ctx context.Context) error {
	return a.registry.StopAll(ctx)
}

// SetDependencies declares plugin dependencies.
func (a *RegistryAdapter) SetDependencies(pluginName string, dependencies []string) error {
	return a.registry.SetDependencies(pluginName, dependencies)
}

// pluginWrapper wraps a HelixAgent LLMPlugin to implement the module Plugin interface.
type pluginWrapper struct {
	helix helixplugins.LLMPlugin
}

func (w *pluginWrapper) Name() string {
	return w.helix.Name()
}

func (w *pluginWrapper) Version() string {
	return w.helix.Version()
}

func (w *pluginWrapper) Init(ctx context.Context, config modplugin.Config) error {
	// Convert module Config to map[string]interface{}
	helixConfig := make(map[string]interface{})
	for k, v := range config {
		helixConfig[k] = v
	}
	return w.helix.Init(helixConfig)
}

func (w *pluginWrapper) Start(ctx context.Context) error {
	// HelixAgent plugins don't have a separate Start method
	return nil
}

func (w *pluginWrapper) Stop(ctx context.Context) error {
	return w.helix.Shutdown(ctx)
}

func (w *pluginWrapper) HealthCheck(ctx context.Context) error {
	return w.helix.HealthCheck(ctx)
}

// helixPluginAdapter wraps a module Plugin to implement HelixAgent LLMPlugin interface.
type helixPluginAdapter struct {
	mod modplugin.Plugin
}

func (a *helixPluginAdapter) Name() string {
	return a.mod.Name()
}

func (a *helixPluginAdapter) Version() string {
	return a.mod.Version()
}

func (a *helixPluginAdapter) Capabilities() *models.ProviderCapabilities {
	// Return basic capabilities for module plugins
	return &models.ProviderCapabilities{
		SupportedModels: []string{a.mod.Name()},
	}
}

func (a *helixPluginAdapter) Init(config map[string]interface{}) error {
	modConfig := make(modplugin.Config)
	for k, v := range config {
		modConfig[k] = v
	}
	return a.mod.Init(context.Background(), modConfig)
}

func (a *helixPluginAdapter) Shutdown(ctx context.Context) error {
	return a.mod.Stop(ctx)
}

func (a *helixPluginAdapter) HealthCheck(ctx context.Context) error {
	return a.mod.HealthCheck(ctx)
}

func (a *helixPluginAdapter) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	return nil, fmt.Errorf("module plugin does not support LLM completion")
}

func (a *helixPluginAdapter) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	return nil, fmt.Errorf("module plugin does not support LLM streaming")
}

func (a *helixPluginAdapter) SetSecurityContext(context *helixplugins.PluginSecurityContext) error {
	// Module plugins don't support security context
	return nil
}

// LoaderAdapter adapts the module's Loader to HelixAgent's PluginLoader.
type LoaderAdapter struct {
	loader modloader.Loader
}

// NewSharedObjectLoaderAdapter creates a loader for shared object plugins.
func NewSharedObjectLoaderAdapter(config *modloader.Config) *LoaderAdapter {
	if config == nil {
		config = modloader.DefaultConfig()
	}
	return &LoaderAdapter{
		loader: modloader.NewSharedObjectLoader(config),
	}
}

// NewProcessLoaderAdapter creates a loader for process-based plugins.
func NewProcessLoaderAdapter(config *modloader.Config) *LoaderAdapter {
	if config == nil {
		config = modloader.DefaultConfig()
	}
	return &LoaderAdapter{
		loader: modloader.NewProcessLoader(config),
	}
}

// Load loads a plugin from a path.
func (a *LoaderAdapter) Load(path string) (helixplugins.LLMPlugin, error) {
	p, err := a.loader.Load(path)
	if err != nil {
		return nil, err
	}
	return &helixPluginAdapter{mod: p}, nil
}

// LoadDir loads all plugins from a directory.
func (a *LoaderAdapter) LoadDir(dir string) ([]helixplugins.LLMPlugin, error) {
	plugins, err := a.loader.LoadDir(dir)
	if err != nil {
		return nil, err
	}
	result := make([]helixplugins.LLMPlugin, len(plugins))
	for i, p := range plugins {
		result[i] = &helixPluginAdapter{mod: p}
	}
	return result, nil
}

// StructuredParserAdapter adapts the module's structured output parsers.
type StructuredParserAdapter struct {
	parser modstructured.Parser
}

// NewJSONParserAdapter creates a JSON parser adapter.
func NewJSONParserAdapter() *StructuredParserAdapter {
	return &StructuredParserAdapter{
		parser: modstructured.NewJSONParser(),
	}
}

// NewYAMLParserAdapter creates a YAML parser adapter.
func NewYAMLParserAdapter() *StructuredParserAdapter {
	return &StructuredParserAdapter{
		parser: modstructured.NewYAMLParser(),
	}
}

// NewMarkdownParserAdapter creates a Markdown parser adapter.
func NewMarkdownParserAdapter() *StructuredParserAdapter {
	return &StructuredParserAdapter{
		parser: modstructured.NewMarkdownParser(),
	}
}

// Parse parses structured output.
func (a *StructuredParserAdapter) Parse(output string, schema *Schema) (interface{}, error) {
	modSchema := ToModuleSchema(schema)
	return a.parser.Parse(output, modSchema)
}

// Schema represents the expected structure of parsed output.
type Schema struct {
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Description string             `json:"description,omitempty"`
}

// ToModuleSchema converts an adapter Schema to module Schema.
func ToModuleSchema(s *Schema) *modstructured.Schema {
	if s == nil {
		return nil
	}
	var props map[string]*modstructured.Schema
	if s.Properties != nil {
		props = make(map[string]*modstructured.Schema)
		for k, v := range s.Properties {
			props[k] = ToModuleSchema(v)
		}
	}
	return &modstructured.Schema{
		Type:        s.Type,
		Properties:  props,
		Required:    s.Required,
		Items:       ToModuleSchema(s.Items),
		Description: s.Description,
	}
}

// ValidatorAdapter adapts the module's Validator.
type ValidatorAdapter struct {
	validator *modstructured.Validator
}

// NewValidatorAdapter creates a new validator adapter.
func NewValidatorAdapter(strictMode bool) *ValidatorAdapter {
	return &ValidatorAdapter{
		validator: modstructured.NewValidator(strictMode),
	}
}

// Validate validates output against a schema.
func (a *ValidatorAdapter) Validate(output string, schema *Schema) (*ValidationResult, error) {
	modSchema := ToModuleSchema(schema)
	result, err := a.validator.Validate(output, modSchema)
	if err != nil {
		return nil, err
	}
	return ToAdapterValidationResult(result), nil
}

// Repair attempts to fix common issues in structured output.
func (a *ValidatorAdapter) Repair(output string, schema *Schema) (string, error) {
	modSchema := ToModuleSchema(schema)
	return a.validator.Repair(output, modSchema)
}

// ValidationResult represents the result of validation.
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
	Data   interface{}
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Path    string
	Message string
	Value   string
}

// ToAdapterValidationResult converts module ValidationResult.
func ToAdapterValidationResult(m *modstructured.ValidationResult) *ValidationResult {
	if m == nil {
		return nil
	}
	errors := make([]ValidationError, len(m.Errors))
	for i, e := range m.Errors {
		errors[i] = ValidationError{
			Path:    e.Path,
			Message: e.Message,
			Value:   e.Value,
		}
	}
	return &ValidationResult{
		Valid:  m.Valid,
		Errors: errors,
		Data:   m.Data,
	}
}

// StateTrackerAdapter adapts the module's StateTracker.
type StateTrackerAdapter struct {
	tracker *modplugin.StateTracker
}

// NewStateTrackerAdapter creates a new state tracker adapter.
func NewStateTrackerAdapter() *StateTrackerAdapter {
	return &StateTrackerAdapter{
		tracker: modplugin.NewStateTracker(),
	}
}

// Get returns the current state.
func (a *StateTrackerAdapter) Get() PluginState {
	return PluginState(a.tracker.Get())
}

// Set sets the state.
func (a *StateTrackerAdapter) Set(state PluginState) {
	a.tracker.Set(modplugin.State(state))
}

// Transition attempts a state transition.
func (a *StateTrackerAdapter) Transition(expected, next PluginState) error {
	return a.tracker.Transition(modplugin.State(expected), modplugin.State(next))
}

// PluginState represents the state of a plugin.
type PluginState int

const (
	StateUninitialized PluginState = iota
	StateInitialized
	StateRunning
	StateStopped
	StateFailed
)

// CheckVersionConstraint checks version compatibility.
func CheckVersionConstraint(version, constraint string) (bool, error) {
	return modregistry.CheckVersionConstraint(version, constraint)
}
