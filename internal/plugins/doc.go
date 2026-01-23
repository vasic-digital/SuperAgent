// Package plugins provides a hot-reloadable plugin system for HelixAgent.
//
// This package enables dynamic extension of HelixAgent functionality through
// plugins that can be loaded, unloaded, and updated at runtime without
// server restarts.
//
// # Plugin Architecture
//
// Plugins follow a defined interface:
//
//	type Plugin interface {
//	    Name() string
//	    Version() string
//	    Init(ctx context.Context, config map[string]interface{}) error
//	    Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)
//	    Shutdown(ctx context.Context) error
//	}
//
// # Plugin Registry
//
// The registry manages plugin lifecycle:
//
//	registry := plugins.NewRegistry(config)
//
//	// Register a plugin
//	if err := registry.Register(myPlugin); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get plugin by name
//	plugin, err := registry.Get("my-plugin")
//
//	// List all plugins
//	allPlugins := registry.List()
//
// # Plugin Loader
//
// The loader handles plugin discovery and loading:
//
//	loader := plugins.NewLoader(pluginDir, registry)
//
//	// Load all plugins from directory
//	if err := loader.LoadAll(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load specific plugin
//	if err := loader.Load(ctx, "plugin-name"); err != nil {
//	    log.Fatal(err)
//	}
//
// # Hot Reload
//
// Plugins support hot reload for development:
//
//	// Enable hot reload watching
//	loader.EnableHotReload(true)
//
//	// Plugins are automatically reloaded when files change
//
// Note: Hot reload may be disabled in production for stability.
//
// # Plugin Configuration
//
// Plugins are configured via YAML files:
//
//	plugins:
//	  my-plugin:
//	    enabled: true
//	    config:
//	      option1: value1
//	      option2: value2
//
// # Dependency Resolution
//
// The plugin system supports dependencies:
//
//	type PluginManifest struct {
//	    Name         string   `yaml:"name"`
//	    Version      string   `yaml:"version"`
//	    Dependencies []string `yaml:"dependencies"`
//	}
//
// Dependencies are resolved and loaded in order.
//
// # Plugin Types
//
// Supported plugin types:
//
//   - Tool plugins: Add new tools to the tool registry
//   - Provider plugins: Add new LLM providers
//   - Middleware plugins: Add request/response processing
//   - Handler plugins: Add new API endpoints
//
// # Key Files
//
//   - registry.go: Plugin registry
//   - loader.go: Plugin loading and discovery
//   - lifecycle.go: Plugin lifecycle management
//   - hot_reload.go: Hot reload functionality
//   - types.go: Plugin type definitions
//
// # Example: Creating a Plugin
//
//	type MyPlugin struct {
//	    name    string
//	    version string
//	}
//
//	func (p *MyPlugin) Name() string { return p.name }
//	func (p *MyPlugin) Version() string { return p.version }
//
//	func (p *MyPlugin) Init(ctx context.Context, config map[string]interface{}) error {
//	    // Initialize plugin
//	    return nil
//	}
//
//	func (p *MyPlugin) Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error) {
//	    // Process input and return output
//	    return &PluginOutput{Result: "processed"}, nil
//	}
//
//	func (p *MyPlugin) Shutdown(ctx context.Context) error {
//	    // Cleanup resources
//	    return nil
//	}
package plugins
