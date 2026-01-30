package memory

import (
	"fmt"
	"strings"
	"time"
)

// ConflictStrategy defines how to resolve conflicts between memories
type ConflictStrategy string

const (
	// ConflictStrategyLastWriteWins uses timestamp to determine winner
	ConflictStrategyLastWriteWins ConflictStrategy = "last_write_wins"

	// ConflictStrategyMergeAll merges all fields intelligently
	ConflictStrategyMergeAll ConflictStrategy = "merge_all"

	// ConflictStrategyImportance uses importance score to determine winner
	ConflictStrategyImportance ConflictStrategy = "importance"

	// ConflictStrategyVectorClock uses vector clock for causal ordering
	ConflictStrategyVectorClock ConflictStrategy = "vector_clock"

	// ConflictStrategyCustom uses custom resolution logic
	ConflictStrategyCustom ConflictStrategy = "custom"
)

// CRDTResolver resolves conflicts between memory states using CRDT principles
type CRDTResolver struct {
	strategy       ConflictStrategy
	customResolver func(*Memory, *MemoryEvent) *Memory
}

// NewCRDTResolver creates a new CRDT resolver
func NewCRDTResolver(strategy ConflictStrategy) *CRDTResolver {
	return &CRDTResolver{
		strategy: strategy,
	}
}

// WithCustomResolver sets a custom resolution function
func (cr *CRDTResolver) WithCustomResolver(fn func(*Memory, *MemoryEvent) *Memory) *CRDTResolver {
	cr.customResolver = fn
	return cr
}

// Merge merges a local memory with a remote event
func (cr *CRDTResolver) Merge(local *Memory, remote *MemoryEvent) *Memory {
	switch cr.strategy {
	case ConflictStrategyLastWriteWins:
		return cr.lastWriteWins(local, remote)

	case ConflictStrategyMergeAll:
		return cr.mergeAll(local, remote)

	case ConflictStrategyImportance:
		return cr.mergeByImportance(local, remote)

	case ConflictStrategyVectorClock:
		return cr.mergeByVectorClock(local, remote)

	case ConflictStrategyCustom:
		if cr.customResolver != nil {
			return cr.customResolver(local, remote)
		}
		// Fall back to last write wins
		return cr.lastWriteWins(local, remote)

	default:
		return cr.lastWriteWins(local, remote)
	}
}

// lastWriteWins uses timestamp to determine the winning version
func (cr *CRDTResolver) lastWriteWins(local *Memory, remote *MemoryEvent) *Memory {
	// If remote is newer, use remote data
	if remote.Timestamp.After(local.UpdatedAt) {
		return cr.memoryFromEvent(local.ID, remote)
	}

	// Otherwise keep local
	return local
}

// mergeAll intelligently merges all fields
func (cr *CRDTResolver) mergeAll(local *Memory, remote *MemoryEvent) *Memory {
	merged := &Memory{
		ID:        local.ID,
		UserID:    local.UserID,
		SessionID: local.SessionID,
	}

	// Content: use longer or more recent
	if len(remote.Content) > len(local.Content) {
		merged.Content = remote.Content
	} else if len(remote.Content) == len(local.Content) && remote.Timestamp.After(local.UpdatedAt) {
		merged.Content = remote.Content
	} else {
		merged.Content = local.Content
	}

	// Embedding: use remote if available, otherwise local
	if len(remote.Embedding) > 0 {
		merged.Embedding = remote.Embedding
	} else {
		merged.Embedding = local.Embedding
	}

	// Importance: use maximum
	if remote.Importance > local.Importance {
		merged.Importance = remote.Importance
	} else {
		merged.Importance = local.Importance
	}

	// Note: Tags and Entities are stored in metadata or separately
	// This is simplified - in production you'd retrieve and merge from store

	// Timestamps: use appropriate values
	merged.CreatedAt = local.CreatedAt
	if remote.Timestamp.After(local.UpdatedAt) {
		merged.UpdatedAt = remote.Timestamp
	} else {
		merged.UpdatedAt = local.UpdatedAt
	}
	merged.LastAccess = time.Now()

	// Metadata: merge metadata maps
	if local.Metadata != nil || remote.Metadata != nil {
		merged.Metadata = make(map[string]interface{})
		for k, v := range local.Metadata {
			merged.Metadata[k] = v
		}
		for k, v := range remote.Metadata {
			merged.Metadata[k] = v
		}
	}

	return merged
}

// mergeByImportance uses importance score to determine winner
func (cr *CRDTResolver) mergeByImportance(local *Memory, remote *MemoryEvent) *Memory {
	if remote.Importance > local.Importance {
		return cr.memoryFromEvent(local.ID, remote)
	}
	return local
}

// mergeByVectorClock uses vector clock for causal ordering
func (cr *CRDTResolver) mergeByVectorClock(local *Memory, remote *MemoryEvent) *Memory {
	// Parse remote vector clock
	remoteVC, err := ParseVectorClock(remote.VectorClock)
	if err != nil {
		// Fall back to last write wins
		return cr.lastWriteWins(local, remote)
	}

	// Parse local vector clock (if stored in metadata)
	localVCStr, ok := local.Metadata["vector_clock"].(string)
	if !ok {
		// No local vector clock, use timestamp
		return cr.lastWriteWins(local, remote)
	}

	localVC, err := ParseVectorClock(localVCStr)
	if err != nil {
		return cr.lastWriteWins(local, remote)
	}

	// Check causal ordering
	if remoteVC.HappensBefore(localVC) {
		// Remote happened before local, keep local
		return local
	} else if localVC.HappensBefore(remoteVC) {
		// Local happened before remote, use remote
		return cr.memoryFromEvent(local.ID, remote)
	} else {
		// Concurrent updates, merge all
		return cr.mergeAll(local, remote)
	}
}

// memoryFromEvent creates a Memory from a MemoryEvent
func (cr *CRDTResolver) memoryFromEvent(memoryID string, event *MemoryEvent) *Memory {
	memory := &Memory{
		ID:         memoryID,
		UserID:     event.UserID,
		SessionID:  event.SessionID,
		Content:    event.Content,
		Embedding:  event.Embedding,
		Importance: event.Importance,
		Type:       MemoryTypeSemantic, // Default type
		CreatedAt:  event.Timestamp,
		UpdatedAt:  event.Timestamp,
		LastAccess: time.Now(),
	}

	// Store vector clock and other event data in metadata
	if memory.Metadata == nil {
		memory.Metadata = make(map[string]interface{})
	}
	memory.Metadata["vector_clock"] = event.VectorClock
	if len(event.Tags) > 0 {
		memory.Metadata["tags"] = event.Tags
	}
	if len(event.Entities) > 0 {
		memory.Metadata["entities"] = event.Entities
	}

	return memory
}

// mergeTags merges two tag lists, removing duplicates
func (cr *CRDTResolver) mergeTags(local, remote []string) []string {
	tagSet := make(map[string]bool)

	for _, tag := range local {
		tagSet[tag] = true
	}

	for _, tag := range remote {
		tagSet[tag] = true
	}

	merged := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		merged = append(merged, tag)
	}

	return merged
}

// mergeEntities merges entity lists from metadata
func (cr *CRDTResolver) mergeEntities(local []MemoryEntity, remote []MemoryEntity) []MemoryEntity {
	// Create map of local entities by ID
	localMap := make(map[string]MemoryEntity)
	for _, e := range local {
		localMap[e.ID] = e
	}

	// Merge remote entities
	for _, re := range remote {
		if le, exists := localMap[re.ID]; exists {
			// Entity exists, merge
			if re.Confidence > le.Confidence {
				// Use remote (higher confidence)
				localMap[re.ID] = re
			}
		} else {
			// New entity, add it
			localMap[re.ID] = re
		}
	}

	// Convert back to slice
	merged := make([]MemoryEntity, 0, len(localMap))
	for _, e := range localMap {
		merged = append(merged, e)
	}

	return merged
}

// DetectConflict checks if there's a conflict between local and remote
func (cr *CRDTResolver) DetectConflict(local *Memory, remote *MemoryEvent) (bool, string) {
	conflicts := make([]string, 0)

	// Content conflict
	if local.Content != remote.Content && local.UpdatedAt.After(remote.Timestamp) {
		conflicts = append(conflicts, "content")
	}

	// Importance conflict
	if local.Importance != remote.Importance {
		conflicts = append(conflicts, "importance")
	}

	// Tag conflict (from metadata)
	localTags, _ := local.Metadata["tags"].([]string)
	if !cr.tagsEqual(localTags, remote.Tags) {
		conflicts = append(conflicts, "tags")
	}

	if len(conflicts) > 0 {
		return true, fmt.Sprintf("conflicts: %s", strings.Join(conflicts, ", "))
	}

	return false, ""
}

// tagsEqual checks if two tag lists are equal (order doesn't matter)
func (cr *CRDTResolver) tagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	tagSet := make(map[string]bool)
	for _, tag := range a {
		tagSet[tag] = true
	}

	for _, tag := range b {
		if !tagSet[tag] {
			return false
		}
	}

	return true
}

// ConflictReport represents a detected conflict
type ConflictReport struct {
	MemoryID      string                 `json:"memory_id"`
	LocalVersion  *Memory                `json:"local_version"`
	RemoteVersion *MemoryEvent           `json:"remote_version"`
	ConflictType  string                 `json:"conflict_type"`
	ResolvedWith  ConflictStrategy       `json:"resolved_with"`
	Resolution    *Memory                `json:"resolution"`
	Timestamp     time.Time              `json:"timestamp"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// ResolveWithReport resolves a conflict and returns a detailed report
func (cr *CRDTResolver) ResolveWithReport(local *Memory, remote *MemoryEvent) *ConflictReport {
	hasConflict, conflictType := cr.DetectConflict(local, remote)

	report := &ConflictReport{
		MemoryID:      local.ID,
		LocalVersion:  local,
		RemoteVersion: remote,
		ConflictType:  conflictType,
		ResolvedWith:  cr.strategy,
		Timestamp:     time.Now(),
		Details:       make(map[string]interface{}),
	}

	if !hasConflict {
		report.ConflictType = "none"
		report.Resolution = local
		return report
	}

	// Resolve conflict
	report.Resolution = cr.Merge(local, remote)

	// Add resolution details
	report.Details["content_changed"] = local.Content != report.Resolution.Content
	report.Details["importance_changed"] = local.Importance != report.Resolution.Importance

	// Get tags from metadata for comparison
	localTags, _ := local.Metadata["tags"].([]string)
	resolvedTags, _ := report.Resolution.Metadata["tags"].([]string)
	report.Details["tags_merged"] = len(resolvedTags) > len(localTags)

	return report
}
