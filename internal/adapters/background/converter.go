// Package background provides adapters bridging HelixAgent's internal background
// types to the generic digital.vasic.background module.
//
// The adapter package maintains backward compatibility with code using
// dev.helix.agent/internal/background while delegating core operations
// to digital.vasic.background.
package background

import (
	"encoding/json"

	internalbackground "dev.helix.agent/internal/background"
	internalmodels "dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
	extractedmodels "digital.vasic.models"
)

// convertToInternalTask converts extracted BackgroundTask to internal BackgroundTask
func convertToInternalTask(extracted *extractedmodels.BackgroundTask) *internalmodels.BackgroundTask {
	if extracted == nil {
		return nil
	}
	// Use JSON marshaling/unmarshaling for conversion since structs are identical
	data, err := json.Marshal(extracted)
	if err != nil {
		return nil
	}
	var internal internalmodels.BackgroundTask
	if err := json.Unmarshal(data, &internal); err != nil {
		return nil
	}
	return &internal
}

// convertToExtractedTask converts internal BackgroundTask to extracted BackgroundTask
func convertToExtractedTask(internal *internalmodels.BackgroundTask) *extractedmodels.BackgroundTask {
	if internal == nil {
		return nil
	}
	data, err := json.Marshal(internal)
	if err != nil {
		return nil
	}
	var extracted extractedmodels.BackgroundTask
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil
	}
	return &extracted
}

// convertToInternalTasks converts slice of extracted BackgroundTask to slice of internal BackgroundTask
func convertToInternalTasks(extracted []*extractedmodels.BackgroundTask) []*internalmodels.BackgroundTask {
	if extracted == nil {
		return nil
	}
	result := make([]*internalmodels.BackgroundTask, len(extracted))
	for i, task := range extracted {
		result[i] = convertToInternalTask(task)
	}
	return result
}

// convertToExtractedTasks converts slice of internal BackgroundTask to slice of extracted BackgroundTask
func convertToExtractedTasks(internal []*internalmodels.BackgroundTask) []*extractedmodels.BackgroundTask {
	if internal == nil {
		return nil
	}
	result := make([]*extractedmodels.BackgroundTask, len(internal))
	for i, task := range internal {
		result[i] = convertToExtractedTask(task)
	}
	return result
}

// convertTaskPriority converts internal TaskPriority to extracted TaskPriority
func convertTaskPriority(priority internalmodels.TaskPriority) extractedmodels.TaskPriority {
	return extractedmodels.TaskPriority(priority)
}

// convertToInternalResourceSnapshot converts extracted ResourceSnapshot to internal ResourceSnapshot
func convertToInternalResourceSnapshot(extracted *extractedmodels.ResourceSnapshot) *internalmodels.ResourceSnapshot {
	if extracted == nil {
		return nil
	}
	data, err := json.Marshal(extracted)
	if err != nil {
		return nil
	}
	var internal internalmodels.ResourceSnapshot
	if err := json.Unmarshal(data, &internal); err != nil {
		return nil
	}
	return &internal
}

// convertToExtractedResourceSnapshot converts internal ResourceSnapshot to extracted ResourceSnapshot
func convertToExtractedResourceSnapshot(internal *internalmodels.ResourceSnapshot) *extractedmodels.ResourceSnapshot {
	if internal == nil {
		return nil
	}
	data, err := json.Marshal(internal)
	if err != nil {
		return nil
	}
	var extracted extractedmodels.ResourceSnapshot
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil
	}
	return &extracted
}

// convertToInternalSystemResources converts extracted SystemResources to internal SystemResources
func convertToInternalSystemResources(extracted *extractedbackground.SystemResources) *internalbackground.SystemResources {
	if extracted == nil {
		return nil
	}
	data, err := json.Marshal(extracted)
	if err != nil {
		return nil
	}
	var internal internalbackground.SystemResources
	if err := json.Unmarshal(data, &internal); err != nil {
		return nil
	}
	return &internal
}

// convertToInternalStuckAnalysis converts extracted StuckAnalysis to internal StuckAnalysis
func convertToInternalStuckAnalysis(extracted *extractedbackground.StuckAnalysis) *internalbackground.StuckAnalysis {
	if extracted == nil {
		return nil
	}
	data, err := json.Marshal(extracted)
	if err != nil {
		return nil
	}
	var internal internalbackground.StuckAnalysis
	if err := json.Unmarshal(data, &internal); err != nil {
		return nil
	}
	return &internal
}

// convertTaskStatus converts internal TaskStatus to extracted TaskStatus
func convertTaskStatus(status internalmodels.TaskStatus) extractedmodels.TaskStatus {
	return extractedmodels.TaskStatus(status)
}

// convertToExtractedSystemResources converts internal SystemResources to extracted SystemResources
func convertToExtractedSystemResources(internal *internalbackground.SystemResources) *extractedbackground.SystemResources {
	if internal == nil {
		return nil
	}
	data, err := json.Marshal(internal)
	if err != nil {
		return nil
	}
	var extracted extractedbackground.SystemResources
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil
	}
	return &extracted
}

// convertToExtractedStuckAnalysis converts internal StuckAnalysis to extracted StuckAnalysis
func convertToExtractedStuckAnalysis(internal *internalbackground.StuckAnalysis) *extractedbackground.StuckAnalysis {
	if internal == nil {
		return nil
	}
	data, err := json.Marshal(internal)
	if err != nil {
		return nil
	}
	var extracted extractedbackground.StuckAnalysis
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil
	}
	return &extracted
}
