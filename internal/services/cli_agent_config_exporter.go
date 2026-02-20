package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"llm-verifier/pkg/cliagents"

	"github.com/sirupsen/logrus"
)

// CLIAgentConfigExporter handles re-exporting CLI agent configurations after verification
type CLIAgentConfigExporter struct {
	generator  *cliagents.UnifiedGenerator
	log        *logrus.Logger
	mu         sync.Mutex
	lastExport time.Time
}

// NewCLIAgentConfigExporter creates a new CLI agent config exporter
func NewCLIAgentConfigExporter(log *logrus.Logger) *CLIAgentConfigExporter {
	if log == nil {
		log = logrus.New()
	}

	config := cliagents.DefaultGeneratorConfig()
	generator := cliagents.NewUnifiedGenerator(config)

	return &CLIAgentConfigExporter{
		generator: generator,
		log:       log,
	}
}

// ExportAllConfigs exports configurations for all supported CLI agents
func (e *CLIAgentConfigExporter) ExportAllConfigs(ctx context.Context) (*ExportResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	start := time.Now()
	result := &ExportResult{
		ExportedAt: start,
		Agents:     make([]AgentExportResult, 0),
	}

	e.log.Info("Starting CLI agent config re-export after verification")

	// Generate all configurations
	results, err := e.generator.GenerateAll(ctx)
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("failed to generate configs: %w", err)
	}

	// Save each configuration
	for _, genResult := range results {
		agentResult := AgentExportResult{
			AgentType: string(genResult.AgentType),
			Success:   genResult.Success,
			Errors:    genResult.Errors,
		}

		if genResult.Success && genResult.Config != nil {
			// Save to file
			if err := e.saveConfig(genResult); err != nil {
				agentResult.Success = false
				agentResult.Errors = append(agentResult.Errors, fmt.Sprintf("save failed: %v", err))
			} else {
				agentResult.Path = genResult.ConfigPath
			}
		}

		result.Agents = append(result.Agents, agentResult)
		if agentResult.Success {
			result.SuccessCount++
		} else {
			result.FailedCount++
		}
	}

	e.lastExport = time.Now()
	result.Duration = time.Since(start)

	e.log.WithFields(logrus.Fields{
		"success":  result.SuccessCount,
		"failed":   result.FailedCount,
		"duration": result.Duration.String(),
	}).Info("CLI agent config re-export completed")

	return result, nil
}

// saveConfig saves a generated configuration to the appropriate location
func (e *CLIAgentConfigExporter) saveConfig(result *cliagents.GenerationResult) error {
	if result.Config == nil {
		return fmt.Errorf("no configuration to save")
	}

	schema, err := e.generator.GetSchema(result.AgentType)
	if err != nil || schema == nil {
		return fmt.Errorf("no schema found for agent %s: %w", result.AgentType, err)
	}

	// Use the generator's SaveConfig method which handles paths
	if err := e.generator.SaveConfig(result); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	e.log.WithFields(logrus.Fields{
		"agent": result.AgentType,
		"path":  result.ConfigPath,
	}).Debug("Saved CLI agent config")

	return nil
}

// ExportResult contains the result of exporting all CLI agent configs
type ExportResult struct {
	ExportedAt   time.Time           `json:"exported_at"`
	Agents       []AgentExportResult `json:"agents"`
	SuccessCount int                 `json:"success_count"`
	FailedCount  int                 `json:"failed_count"`
	Duration     time.Duration       `json:"duration"`
	Error        string              `json:"error,omitempty"`
}

// AgentExportResult contains the result of exporting a single agent config
type AgentExportResult struct {
	AgentType string   `json:"agent_type"`
	Success   bool     `json:"success"`
	Path      string   `json:"path,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

// OnVerificationComplete is the callback function for StartupVerifier
func (e *CLIAgentConfigExporter) OnVerificationComplete(ctx context.Context, result any) error {
	e.log.Info("Verification complete - triggering CLI agent config re-export")
	_, err := e.ExportAllConfigs(ctx)
	return err
}

// GetLastExport returns the time of the last export
func (e *CLIAgentConfigExporter) GetLastExport() time.Time {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastExport
}
