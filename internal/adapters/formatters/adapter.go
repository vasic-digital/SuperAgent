// Package formatters provides adapters that bridge HelixAgent-specific formatter
// operations with the generic digital.vasic.formatters module.
//
// This adapter layer maintains backward compatibility with existing code while
// allowing gradual migration to the extracted formatters module. The internal
// formatters package has HelixAgent-specific extensions (AgentName, SessionID)
// that are not part of the generic module.
package formatters

import (
	"context"
	"time"

	genericfmt "digital.vasic.formatters/pkg/formatter"
	"digital.vasic.formatters/pkg/native"
	genericreg "digital.vasic.formatters/pkg/registry"

	"dev.helix.agent/internal/formatters"
)

// TypeAliases for re-exporting generic module types that can be used directly.

// GenericFormatter is the generic formatter interface from the extracted module.
type GenericFormatter = genericfmt.Formatter

// GenericFormatRequest is the generic format request from the extracted module.
type GenericFormatRequest = genericfmt.FormatRequest

// GenericFormatResult is the generic format result from the extracted module.
type GenericFormatResult = genericfmt.FormatResult

// GenericFormatStats is the generic format stats from the extracted module.
type GenericFormatStats = genericfmt.FormatStats

// GenericFormatterMetadata is the generic formatter metadata from the extracted module.
type GenericFormatterMetadata = genericfmt.FormatterMetadata

// GenericFormatterType is the generic formatter type from the extracted module.
type GenericFormatterType = genericfmt.FormatterType

// GenericRegistry is the generic formatter registry from the extracted module.
type GenericRegistry = genericreg.Registry

// FormatterType constants re-exported from the generic module.
const (
	FormatterTypeNative  = genericfmt.FormatterTypeNative
	FormatterTypeService = genericfmt.FormatterTypeService
	FormatterTypeBuiltin = genericfmt.FormatterTypeBuiltin
	FormatterTypeUnified = genericfmt.FormatterTypeUnified
)

// ToGenericRequest converts a HelixAgent-specific FormatRequest to the generic module format.
// Note: AgentName and SessionID are HelixAgent-specific and are not carried over.
func ToGenericRequest(req *formatters.FormatRequest) *genericfmt.FormatRequest {
	if req == nil {
		return nil
	}
	return &genericfmt.FormatRequest{
		Content:    req.Content,
		FilePath:   req.FilePath,
		Language:   req.Language,
		Config:     req.Config,
		LineLength: req.LineLength,
		IndentSize: req.IndentSize,
		UseTabs:    req.UseTabs,
		CheckOnly:  req.CheckOnly,
		Timeout:    req.Timeout,
		RequestID:  req.RequestID,
	}
}

// FromGenericResult converts a generic FormatResult to the HelixAgent-specific format.
func FromGenericResult(result *genericfmt.FormatResult) *formatters.FormatResult {
	if result == nil {
		return nil
	}

	var stats *formatters.FormatStats
	if result.Stats != nil {
		stats = &formatters.FormatStats{
			LinesTotal:   result.Stats.LinesTotal,
			LinesChanged: result.Stats.LinesChanged,
			BytesTotal:   result.Stats.BytesTotal,
			BytesChanged: result.Stats.BytesChanged,
			Violations:   result.Stats.Violations,
		}
	}

	return &formatters.FormatResult{
		Content:          result.Content,
		Changed:          result.Changed,
		FormatterName:    result.FormatterName,
		FormatterVersion: result.FormatterVersion,
		Duration:         result.Duration,
		Success:          result.Success,
		Error:            result.Error,
		Warnings:         result.Warnings,
		Stats:            stats,
	}
}

// ToGenericMetadata converts HelixAgent-specific FormatterMetadata to the generic format.
func ToGenericMetadata(metadata *formatters.FormatterMetadata) *genericfmt.FormatterMetadata {
	if metadata == nil {
		return nil
	}
	return &genericfmt.FormatterMetadata{
		Name:            metadata.Name,
		Type:            genericfmt.FormatterType(metadata.Type),
		Architecture:    metadata.Architecture,
		GitHubURL:       metadata.GitHubURL,
		Version:         metadata.Version,
		Languages:       metadata.Languages,
		License:         metadata.License,
		InstallMethod:   metadata.InstallMethod,
		BinaryPath:      metadata.BinaryPath,
		ServiceURL:      metadata.ServiceURL,
		ConfigFormat:    metadata.ConfigFormat,
		DefaultConfig:   metadata.DefaultConfig,
		Performance:     metadata.Performance,
		Complexity:      metadata.Complexity,
		SupportsStdin:   metadata.SupportsStdin,
		SupportsInPlace: metadata.SupportsInPlace,
		SupportsCheck:   metadata.SupportsCheck,
		SupportsConfig:  metadata.SupportsConfig,
	}
}

// FromGenericMetadata converts generic FormatterMetadata to the HelixAgent-specific format.
func FromGenericMetadata(metadata *genericfmt.FormatterMetadata) *formatters.FormatterMetadata {
	if metadata == nil {
		return nil
	}
	return &formatters.FormatterMetadata{
		Name:            metadata.Name,
		Type:            formatters.FormatterType(metadata.Type),
		Architecture:    metadata.Architecture,
		GitHubURL:       metadata.GitHubURL,
		Version:         metadata.Version,
		Languages:       metadata.Languages,
		License:         metadata.License,
		InstallMethod:   metadata.InstallMethod,
		BinaryPath:      metadata.BinaryPath,
		ServiceURL:      metadata.ServiceURL,
		ConfigFormat:    metadata.ConfigFormat,
		DefaultConfig:   metadata.DefaultConfig,
		Performance:     metadata.Performance,
		Complexity:      metadata.Complexity,
		SupportsStdin:   metadata.SupportsStdin,
		SupportsInPlace: metadata.SupportsInPlace,
		SupportsCheck:   metadata.SupportsCheck,
		SupportsConfig:  metadata.SupportsConfig,
	}
}

// FormatterAdapter wraps a generic formatter to implement the HelixAgent-specific
// Formatter interface.
type FormatterAdapter struct {
	generic genericfmt.Formatter
}

// NewFormatterAdapter creates a new adapter wrapping a generic formatter.
func NewFormatterAdapter(generic genericfmt.Formatter) *FormatterAdapter {
	return &FormatterAdapter{generic: generic}
}

// Name returns the formatter name.
func (a *FormatterAdapter) Name() string {
	return a.generic.Name()
}

// Version returns the formatter version.
func (a *FormatterAdapter) Version() string {
	return a.generic.Version()
}

// Languages returns supported languages.
func (a *FormatterAdapter) Languages() []string {
	return a.generic.Languages()
}

// SupportsStdin returns whether the formatter supports stdin.
func (a *FormatterAdapter) SupportsStdin() bool {
	return a.generic.SupportsStdin()
}

// SupportsInPlace returns whether the formatter supports in-place formatting.
func (a *FormatterAdapter) SupportsInPlace() bool {
	return a.generic.SupportsInPlace()
}

// SupportsCheck returns whether the formatter supports check mode.
func (a *FormatterAdapter) SupportsCheck() bool {
	return a.generic.SupportsCheck()
}

// SupportsConfig returns whether the formatter supports configuration.
func (a *FormatterAdapter) SupportsConfig() bool {
	return a.generic.SupportsConfig()
}

// Format formats code using the generic formatter.
func (a *FormatterAdapter) Format(ctx context.Context, req *formatters.FormatRequest) (*formatters.FormatResult, error) {
	genericReq := ToGenericRequest(req)
	result, err := a.generic.Format(ctx, genericReq)
	if err != nil {
		return nil, err
	}
	return FromGenericResult(result), nil
}

// FormatBatch formats multiple requests.
func (a *FormatterAdapter) FormatBatch(ctx context.Context, reqs []*formatters.FormatRequest) ([]*formatters.FormatResult, error) {
	genericReqs := make([]*genericfmt.FormatRequest, len(reqs))
	for i, req := range reqs {
		genericReqs[i] = ToGenericRequest(req)
	}

	results, err := a.generic.FormatBatch(ctx, genericReqs)
	if err != nil {
		return nil, err
	}

	helixResults := make([]*formatters.FormatResult, len(results))
	for i, result := range results {
		helixResults[i] = FromGenericResult(result)
	}
	return helixResults, nil
}

// HealthCheck performs a health check.
func (a *FormatterAdapter) HealthCheck(ctx context.Context) error {
	return a.generic.HealthCheck(ctx)
}

// ValidateConfig validates formatter configuration.
func (a *FormatterAdapter) ValidateConfig(config map[string]interface{}) error {
	return a.generic.ValidateConfig(config)
}

// DefaultConfig returns the default configuration.
func (a *FormatterAdapter) DefaultConfig() map[string]interface{} {
	return a.generic.DefaultConfig()
}

// Generic returns the underlying generic formatter.
func (a *FormatterAdapter) Generic() genericfmt.Formatter {
	return a.generic
}

// NativeFormatterFactory provides factory methods for creating native formatters
// using the generic module's implementations.
type NativeFormatterFactory struct{}

// NewNativeFormatterFactory creates a new native formatter factory.
func NewNativeFormatterFactory() *NativeFormatterFactory {
	return &NativeFormatterFactory{}
}

// CreateGoFormatter creates a gofmt formatter using the generic module.
func (f *NativeFormatterFactory) CreateGoFormatter() formatters.Formatter {
	return NewFormatterAdapter(native.NewGoFormatter())
}

// CreatePythonFormatter creates a Black Python formatter using the generic module.
func (f *NativeFormatterFactory) CreatePythonFormatter() formatters.Formatter {
	return NewFormatterAdapter(native.NewPythonFormatter())
}

// CreateJSFormatter creates a Prettier JS/TS formatter using the generic module.
func (f *NativeFormatterFactory) CreateJSFormatter() formatters.Formatter {
	return NewFormatterAdapter(native.NewJSFormatter())
}

// CreateRustFormatter creates a rustfmt Rust formatter using the generic module.
func (f *NativeFormatterFactory) CreateRustFormatter() formatters.Formatter {
	return NewFormatterAdapter(native.NewRustFormatter())
}

// CreateSQLFormatter creates a SQL formatter using the generic module.
func (f *NativeFormatterFactory) CreateSQLFormatter() formatters.Formatter {
	return NewFormatterAdapter(native.NewSQLFormatter())
}

// ServiceFormatterConfig holds configuration for creating a service formatter.
type ServiceFormatterConfig struct {
	Endpoint   string
	Timeout    time.Duration
	HealthPath string
	FormatPath string
}

// CreateServiceFormatter creates a service-based formatter using the generic module.

// DetectLanguageFromPath detects language from file extension using the generic module.
func DetectLanguageFromPath(filePath string) string {
	return genericreg.DetectLanguageFromPath(filePath)
}

// NewGenericRegistry creates a new generic registry from the extracted module.

// GetDefaultGenericRegistry returns the default registry singleton from the extracted module.
