package cluster

import (
	"sync"
	"time"
)

// ─── Conflict Resolution (Last-Write-Wins with Timestamps) ───────────

// ConflictResolver handles data conflict resolution using LWW (Last-Write-Wins).
// Each entity ID is tracked with the source node and timestamp of the last write.
// If an incoming write has a newer timestamp, it wins.
// If timestamps are equal, the node with higher priority wins.
type ConflictResolver struct {
	mu sync.RWMutex

	// tracked[entityID] = lastWriteInfo
	tracked map[string]WriteInfo

	// Priority lookup for nodes
	nodePriority map[int64]int
}

// WriteInfo tracks the last write to an entity.
type WriteInfo struct {
	SourceID  int64 `json:"source_id"`
	Timestamp int64 `json:"timestamp"`
}

// NewConflictResolver creates a new conflict resolver.
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		tracked:      make(map[string]WriteInfo),
		nodePriority: make(map[int64]int),
	}
}

// SetNodePriority sets the priority for a cluster node.
func (cr *ConflictResolver) SetNodePriority(nodeID int64, priority int) {
	cr.mu.Lock()
	cr.nodePriority[nodeID] = priority
	cr.mu.Unlock()
}

// Resolve checks if an incoming write should be accepted.
// Returns true if the write should be applied (it wins the conflict).
func (cr *ConflictResolver) Resolve(entityID string, sourceID int64, incomingTimestamp int64) bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	existing, exists := cr.tracked[entityID]
	if !exists {
		// No existing write, accept
		cr.tracked[entityID] = WriteInfo{
			SourceID:  sourceID,
			Timestamp: incomingTimestamp,
		}
		return true
	}

	// LWW: newer timestamp wins
	if incomingTimestamp > existing.Timestamp {
		cr.tracked[entityID] = WriteInfo{
			SourceID:  sourceID,
			Timestamp: incomingTimestamp,
		}
		return true
	}

	// Same timestamp: higher priority node wins
	if incomingTimestamp == existing.Timestamp {
		existingPriority := cr.nodePriority[existing.SourceID]
		incomingPriority := cr.nodePriority[sourceID]
		if incomingPriority > existingPriority {
			cr.tracked[entityID] = WriteInfo{
				SourceID:  sourceID,
				Timestamp: incomingTimestamp,
			}
			return true
		}
	}

	// Existing write is newer or same priority, reject
	return false
}

// GetLastWrite returns the last write info for an entity.
func (cr *ConflictResolver) GetLastWrite(entityID string) (WriteInfo, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	info, ok := cr.tracked[entityID]
	return info, ok
}

// Cleanup removes old entries (> 1 hour) to prevent memory growth.
func (cr *ConflictResolver) Cleanup() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cutoff := time.Now().Add(-1 * time.Hour).UnixMilli()
	for id, info := range cr.tracked {
		if info.Timestamp < cutoff {
			delete(cr.tracked, id)
		}
	}
}

// Reset clears all tracked writes (used after full sync).
func (cr *ConflictResolver) Reset() {
	cr.mu.Lock()
	cr.tracked = make(map[string]WriteInfo)
	cr.mu.Unlock()
}

// Stats returns resolver statistics.
func (cr *ConflictResolver) Stats() map[string]any {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return map[string]any{
		"tracked_entities": len(cr.tracked),
		"tracked_nodes":    len(cr.nodePriority),
	}
}
