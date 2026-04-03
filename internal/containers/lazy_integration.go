// Package containers provides lazy container orchestration integration for HelixAgent
package containers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"digital.vasic.containers/pkg/compose"
	"digital.vasic.containers/pkg/health"
	"digital.vasic.containers/pkg/lifecycle"
	"digital.vasic.containers/pkg/logging"
)

// LazyOrchestrator manages lazy container service startup for HelixAgent
type LazyOrchestrator struct {
	services      map[string]*ServiceDefinition
	booters       map[string]*lifecycle.LazyBooter
	orchestrator  compose.ComposeOrchestrator
	healthChecker *health.DefaultChecker
	logger        logging.Logger
	mu            sync.RWMutex
	started       map[string]bool
	failed        map[string]error
	workDir       string
}

// ServiceDefinition defines a lazily-loaded container service
type ServiceDefinition struct {
	Name         string
	ComposeFile  string
	Profile      string
	Required     bool
	Dependencies []string
	HealthTarget *health.HealthTarget
	StartTimeout time.Duration
	StopTimeout  time.Duration
	Description  string
	Category     string
	CostModel    string // "free", "freemium", "paid"
}

// ServiceStatus represents the runtime status of a service
type ServiceStatus struct {
	Name        string
	Category    string
	CostModel   string
	Description string
	Started     bool
	IsStarting  bool
	LastError   string
}

// NewLazyOrchestrator creates a new lazy service orchestrator
func NewLazyOrchestrator(workDir string, logger logging.Logger) (*LazyOrchestrator, error) {
	if logger == nil {
		logger = logging.NopLogger{}
	}

	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// Create default orchestrator
	orch, err := compose.NewDefaultOrchestrator(workDir, logger)
	if err != nil {
		return nil, fmt.Errorf("create default orchestrator: %w", err)
	}

	return &LazyOrchestrator{
		services:      make(map[string]*ServiceDefinition),
		booters:       make(map[string]*lifecycle.LazyBooter),
		started:       make(map[string]bool),
		failed:        make(map[string]error),
		orchestrator:  orch,
		healthChecker: health.NewDefaultChecker(),
		logger:        logger,
		workDir:       workDir,
	}, nil
}

// RegisterNVIDIARAG registers NVIDIA RAG services
func (lo *LazyOrchestrator) RegisterNVIDIARAG() error {
	// Vector Database (Milvus) - Required for RAG
	if err := lo.RegisterService(&ServiceDefinition{
		Name:         "milvus",
		ComposeFile:  filepath.Join(lo.workDir, "docker/nvidia-rag/docker-compose.nvidia-rag.yml"),
		Profile:      "minimal",
		Required:     true,
		Description:  "Milvus vector database for RAG",
		Category:     "rag",
		CostModel:    "free",
		StartTimeout: 5 * time.Minute,
		HealthTarget: &health.HealthTarget{
			Name:     "milvus",
			Host:     "localhost",
			Port:     "19530",
			Type:     health.HealthTCP,
			Timeout:  30 * time.Second,
			Required: true,
		},
	}); err != nil {
		return fmt.Errorf("register milvus: %w", err)
	}

	// Document Extraction (Tika - Open Source)
	if err := lo.RegisterService(&ServiceDefinition{
		Name:         "tika-extraction",
		ComposeFile:  filepath.Join(lo.workDir, "docker/nvidia-rag/docker-compose.nvidia-rag.yml"),
		Profile:      "opensource",
		Required:     false,
		Dependencies: []string{"milvus"},
		Description:  "Apache Tika document extraction (open source)",
		Category:     "rag",
		CostModel:    "free",
		StartTimeout: 2 * time.Minute,
		HealthTarget: &health.HealthTarget{
			Name:     "tika",
			URL:      "http://localhost:8105/version",
			Type:     health.HealthHTTP,
			Timeout:  10 * time.Second,
			Required: false,
		},
	}); err != nil {
		return fmt.Errorf("register tika: %w", err)
	}

	return nil
}

// RegisterService registers a service for lazy loading
func (lo *LazyOrchestrator) RegisterService(svc *ServiceDefinition) error {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if svc.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if svc.ComposeFile == "" {
		return fmt.Errorf("service %s: compose file is required", svc.Name)
	}

	// Set defaults
	if svc.StartTimeout == 0 {
		svc.StartTimeout = 5 * time.Minute
	}
	if svc.StopTimeout == 0 {
		svc.StopTimeout = 30 * time.Second
	}
	if svc.CostModel == "" {
		svc.CostModel = "free"
	}

	lo.services[svc.Name] = svc

	// Create lazy booter for this service
	startFn := func() error {
		return lo.startServiceInternal(svc)
	}
	lo.booters[svc.Name] = lifecycle.NewLazyBooter(startFn)

	lo.logger.Info("registered lazy service: %s (category=%s, cost=%s)",
		svc.Name, svc.Category, svc.CostModel)

	return nil
}

// StartService starts a service and its dependencies on-demand
func (lo *LazyOrchestrator) StartService(ctx context.Context, name string) error {
	lo.mu.RLock()
	svc, exists := lo.services[name]
	booter, hasBooter := lo.booters[name]
	lo.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}
	if !hasBooter {
		return fmt.Errorf("service %s has no booter", name)
	}

	// Start dependencies first
	for _, depName := range svc.Dependencies {
		if err := lo.StartService(ctx, depName); err != nil {
			return fmt.Errorf("dependency %s failed: %w", depName, err)
		}
	}

	// Start this service via lazy booter
	if err := booter.EnsureStarted(); err != nil {
		return fmt.Errorf("start service %s: %w", name, err)
	}

	return nil
}

// GetServiceStatus returns the current status of a service
func (lo *LazyOrchestrator) GetServiceStatus(name string) (*ServiceStatus, error) {
	lo.mu.RLock()
	defer lo.mu.RUnlock()

	svc, exists := lo.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	booter, hasBooter := lo.booters[name]
	status := &ServiceStatus{
		Name:        svc.Name,
		Category:    svc.Category,
		CostModel:   svc.CostModel,
		Description: svc.Description,
	}

	if hasBooter {
		status.Started = booter.Started()
		status.IsStarting = booter.IsStarting()
		if err := booter.GetError(); err != nil {
			status.LastError = err.Error()
		}
	}

	return status, nil
}

// ListServices returns all registered services
func (lo *LazyOrchestrator) ListServices() []*ServiceDefinition {
	lo.mu.RLock()
	defer lo.mu.RUnlock()

	result := make([]*ServiceDefinition, 0, len(lo.services))
	for _, svc := range lo.services {
		result = append(result, svc)
	}
	return result
}

// startServiceInternal performs the actual service startup
func (lo *LazyOrchestrator) startServiceInternal(svc *ServiceDefinition) error {
	project := compose.ComposeProject{
		File:    svc.ComposeFile,
		Profile: svc.Profile,
	}

	lo.logger.Info("starting lazy service: %s (file=%s)", svc.Name, svc.ComposeFile)

	ctx, cancel := context.WithTimeout(context.Background(), svc.StartTimeout)
	defer cancel()

	// Check if compose file exists
	if _, err := os.Stat(svc.ComposeFile); os.IsNotExist(err) {
		lo.mu.Lock()
		lo.failed[svc.Name] = fmt.Errorf("compose file not found: %s", svc.ComposeFile)
		lo.mu.Unlock()
		return lo.failed[svc.Name]
	}

	// Start the service
	if err := lo.orchestrator.Up(ctx, project, compose.WithUpDetach(true), compose.WithWait(true)); err != nil {
		lo.mu.Lock()
		lo.failed[svc.Name] = err
		lo.mu.Unlock()
		return fmt.Errorf("compose up failed: %w", err)
	}

	// Wait for health check if defined
	if svc.HealthTarget != nil {
		if err := lo.waitForHealth(ctx, svc); err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}
	}

	lo.mu.Lock()
	lo.started[svc.Name] = true
	lo.mu.Unlock()

	lo.logger.Info("lazy service started successfully: %s", svc.Name)
	return nil
}

// waitForHealth waits for a service to become healthy
func (lo *LazyOrchestrator) waitForHealth(ctx context.Context, svc *ServiceDefinition) error {
	if svc.HealthTarget == nil {
		return nil
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(2 * time.Minute)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("health check timeout")
			}

			result := lo.healthChecker.Check(ctx, *svc.HealthTarget)
			if result.Healthy {
				return nil
			}
		}
	}
}

// InitializeDefaultServices registers default services for HelixAgent
func InitializeDefaultServices(orch *LazyOrchestrator) error {
	// Register NVIDIA RAG services
	if err := orch.RegisterNVIDIARAG(); err != nil {
		return fmt.Errorf("register NVIDIA RAG: %w", err)
	}

	return nil
}
