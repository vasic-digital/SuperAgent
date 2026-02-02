package challenges

import (
	"context"
	"fmt"
	"time"
)

// ServiceBooter abstracts the boot manager for starting
// infrastructure services.
type ServiceBooter interface {
	BootAll() []BootResult
	ShutdownAll() error
}

// ServiceChecker abstracts health checking for services.
type ServiceChecker interface {
	CheckWithRetry(ctx context.Context, serviceName string) error
}

// BootResult represents the outcome of booting a service.
type BootResult struct {
	Name     string
	Status   string
	Duration time.Duration
	Error    error
}

// HelixInfraProvider implements infrastructure provisioning
// using HelixAgent's existing boot manager and health checker.
// It bridges HelixAgent's service management with the Challenges
// module's infrastructure requirements.
type HelixInfraProvider struct {
	bootManager   ServiceBooter
	healthChecker ServiceChecker
}

// NewHelixInfraProvider creates a new HelixInfraProvider.
func NewHelixInfraProvider(
	bm ServiceBooter,
	hc ServiceChecker,
) *HelixInfraProvider {
	return &HelixInfraProvider{
		bootManager:   bm,
		healthChecker: hc,
	}
}

// EnsureRunning starts the named service using the boot manager.
func (p *HelixInfraProvider) EnsureRunning(
	ctx context.Context,
	serviceName string,
) error {
	if p.bootManager == nil {
		return fmt.Errorf(
			"boot manager not configured for service %s",
			serviceName,
		)
	}
	// BootManager handles compose-based service startup.
	results := p.bootManager.BootAll()
	for _, r := range results {
		if r.Name == serviceName && r.Error != nil {
			return fmt.Errorf(
				"failed to start %s: %w",
				serviceName, r.Error,
			)
		}
	}
	return nil
}

// Release is a no-op for individual services managed by compose.
func (p *HelixInfraProvider) Release(
	_ context.Context,
	_ string,
) error {
	return nil
}

// HealthCheck checks whether the named service is healthy.
func (p *HelixInfraProvider) HealthCheck(
	ctx context.Context,
	serviceName string,
) error {
	if p.healthChecker == nil {
		return fmt.Errorf(
			"health checker not configured for service %s",
			serviceName,
		)
	}
	return p.healthChecker.CheckWithRetry(ctx, serviceName)
}

// Shutdown stops all services managed by the boot manager.
func (p *HelixInfraProvider) Shutdown(
	ctx context.Context,
) error {
	if p.bootManager == nil {
		return nil
	}
	return p.bootManager.ShutdownAll()
}
