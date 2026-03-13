package background

import (
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
	extractedmodels "digital.vasic.models"
)

// ProcessResourceMonitor implements ResourceMonitor using extracted module
type ProcessResourceMonitor struct {
	// extractedMonitor is the actual implementation from the extracted module
	extractedMonitor extractedbackground.ResourceMonitor
	// repository is kept for backward compatibility but unused
	repository TaskRepository
	logger     *logrus.Logger
}

// NewProcessResourceMonitor creates a new process resource monitor
func NewProcessResourceMonitor(repository TaskRepository, logger *logrus.Logger) *ProcessResourceMonitor {
	// Adapt internal repository to extracted repository interface
	adaptedRepo := NewTaskRepositoryAdapter(repository)
	// Create extracted ProcessResourceMonitor
	extractedMonitor := extractedbackground.NewProcessResourceMonitor(adaptedRepo, logger)

	return &ProcessResourceMonitor{
		extractedMonitor: extractedMonitor,
		repository:       repository,
		logger:           logger,
	}
}

// GetSystemResources returns current system resource usage
func (m *ProcessResourceMonitor) GetSystemResources() (*SystemResources, error) {
	extractedResources, err := m.extractedMonitor.GetSystemResources()
	if err != nil {
		return nil, err
	}
	return convertToInternalSystemResources(extractedResources), nil
}

// GetProcessResources returns resource usage for a specific process
func (m *ProcessResourceMonitor) GetProcessResources(pid int) (*models.ResourceSnapshot, error) {
	extractedSnapshot, err := m.extractedMonitor.GetProcessResources(pid)
	if err != nil {
		return nil, err
	}
	return convertToInternalResourceSnapshot(extractedSnapshot), nil
}

// StartMonitoring begins periodic monitoring of a process
func (m *ProcessResourceMonitor) StartMonitoring(taskID string, pid int, interval time.Duration) error {
	return m.extractedMonitor.StartMonitoring(taskID, pid, interval)
}

// StopMonitoring stops monitoring a process
func (m *ProcessResourceMonitor) StopMonitoring(taskID string) error {
	return m.extractedMonitor.StopMonitoring(taskID)
}

// GetLatestSnapshot returns the most recent snapshot for a task
func (m *ProcessResourceMonitor) GetLatestSnapshot(taskID string) (*models.ResourceSnapshot, error) {
	extractedSnapshot, err := m.extractedMonitor.GetLatestSnapshot(taskID)
	if err != nil {
		return nil, err
	}
	return convertToInternalResourceSnapshot(extractedSnapshot), nil
}

// IsResourceAvailable checks if system has enough resources
func (m *ProcessResourceMonitor) IsResourceAvailable(requirements ResourceRequirements) bool {
	extractedRequirements := extractedbackground.ResourceRequirements{
		CPUCores: requirements.CPUCores,
		MemoryMB: requirements.MemoryMB,
		DiskMB:   requirements.DiskMB,
		GPUCount: requirements.GPUCount,
		Priority: extractedmodels.TaskPriority(requirements.Priority),
	}
	return m.extractedMonitor.IsResourceAvailable(extractedRequirements)
}
