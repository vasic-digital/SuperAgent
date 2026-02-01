package formatters

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// FormatterFactory creates formatters based on metadata
type FormatterFactory struct {
	config *Config
	logger *logrus.Logger
}

// NewFormatterFactory creates a new formatter factory
func NewFormatterFactory(config *Config, logger *logrus.Logger) *FormatterFactory {
	return &FormatterFactory{
		config: config,
		logger: logger,
	}
}

// Create creates a formatter based on metadata
func (f *FormatterFactory) Create(metadata *FormatterMetadata) (Formatter, error) {
	switch metadata.Type {
	case FormatterTypeNative:
		return f.createNativeFormatter(metadata)
	case FormatterTypeService:
		return f.createServiceFormatter(metadata)
	case FormatterTypeBuiltin:
		return f.createBuiltinFormatter(metadata)
	case FormatterTypeUnified:
		return f.createUnifiedFormatter(metadata)
	default:
		return nil, fmt.Errorf("unknown formatter type: %s", metadata.Type)
	}
}

// createNativeFormatter creates a native binary formatter
func (f *FormatterFactory) createNativeFormatter(metadata *FormatterMetadata) (Formatter, error) {
	// Native formatters are registered via providers/register.go
	// Use RegisterAllFormatters() instead of dynamic creation
	f.logger.Warnf("Native formatter creation not supported; use register.go: %s", metadata.Name)
	return nil, fmt.Errorf("native formatter creation not supported; register formatters via providers/register.go: %s", metadata.Name)
}

// createServiceFormatter creates a service-based formatter
func (f *FormatterFactory) createServiceFormatter(metadata *FormatterMetadata) (Formatter, error) {
	// Service formatters are registered via providers/register.go
	// Use RegisterAllFormatters() instead of dynamic creation
	f.logger.Warnf("Service formatter creation not supported; use register.go: %s", metadata.Name)
	return nil, fmt.Errorf("service formatter creation not supported; register formatters via providers/register.go: %s", metadata.Name)
}

// createBuiltinFormatter creates a built-in formatter
func (f *FormatterFactory) createBuiltinFormatter(metadata *FormatterMetadata) (Formatter, error) {
	// Builtin formatters are registered via providers/register.go
	// Use RegisterAllFormatters() instead of dynamic creation
	f.logger.Warnf("Builtin formatter creation not supported; use register.go: %s", metadata.Name)
	return nil, fmt.Errorf("builtin formatter creation not supported; register formatters via providers/register.go: %s", metadata.Name)
}

// createUnifiedFormatter creates a unified multi-language formatter
func (f *FormatterFactory) createUnifiedFormatter(metadata *FormatterMetadata) (Formatter, error) {
	// Unified formatters are registered via providers/register.go
	// Use RegisterAllFormatters() instead of dynamic creation
	f.logger.Warnf("Unified formatter creation not supported; use register.go: %s", metadata.Name)
	return nil, fmt.Errorf("unified formatter creation not supported; register formatters via providers/register.go: %s", metadata.Name)
}

// CreateAll creates all formatters from a list of metadata
func (f *FormatterFactory) CreateAll(metadataList []*FormatterMetadata) ([]Formatter, []error) {
	formatters := make([]Formatter, 0)
	errors := make([]error, 0)

	for _, metadata := range metadataList {
		formatter, err := f.Create(metadata)
		if err != nil {
			f.logger.Warnf("Failed to create formatter %s: %v", metadata.Name, err)
			errors = append(errors, err)
			continue
		}

		formatters = append(formatters, formatter)
	}

	return formatters, errors
}
