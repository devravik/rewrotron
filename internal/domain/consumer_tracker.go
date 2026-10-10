package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ConsumerType defines the entity type consuming a knowledge or skill revision.
type ConsumerType string

const (
	ConsumerArtifact ConsumerType = "ARTIFACT"
	ConsumerUnit     ConsumerType = "UNIT"
	ConsumerResult   ConsumerType = "RESULT"
	ConsumerSkill    ConsumerType = "SKILL"
	ConsumerApproval ConsumerType = "APPROVAL"
)

// ConsumerRecord tracks a project entity that consumed a specific global knowledge/skill revision.
type ConsumerRecord struct {
	ProjectID      string       `json:"project_id"`
	ConsumerType   ConsumerType `json:"consumer_type"`
	ConsumerID     string       `json:"consumer_id"`     // e.g. artifact hash, unit ID, result ID
	RevisionRef    string       `json:"revision_ref"`    // Pinned global revision (e.g. pattern-id:v1 or skill-id:1.0.0)
	PinnedEpoch    int64        `json:"pinned_epoch"`    // Epoch when consumption occurred
	IsStale        bool         `json:"is_stale"`        // Marked STALE upon retraction
	StaleReason    string       `json:"stale_reason,omitempty"`
	HistoricalHash string       `json:"historical_hash"` // Immutable byte digest preserved for historical replay
	ConsumedAt     string       `json:"consumed_at"`
}

// RetractionReceipt records an idempotent outbox receipt for capability retraction.
type RetractionReceipt struct {
	ReceiptID      string   `json:"receipt_id"`
	RevisionRef    string   `json:"revision_ref"`
	Epoch          int64    `json:"epoch"`
	RetractionReason string `json:"retraction_reason"`
	AffectedProjects []string `json:"affected_projects"`
	ReceiptHash    string   `json:"receipt_hash"`
	CreatedAt      string   `json:"created_at"`
}

// ConsumerGraphTracker manages durable consumer indexes, epoch advancement,
// retraction propagation, staleness marking, and offline project reopen reconciliation (X79, X39).
type ConsumerGraphTracker struct {
	mu           sync.RWMutex
	currentEpoch int64
	// consumerIndex: revisionRef -> list of ConsumerRecord
	consumerIndex map[string][]*ConsumerRecord
	// projectEpoch: projectID -> last acknowledged epoch
	projectEpoch map[string]int64
	// retractions: revisionRef -> RetractionReceipt
	retractions map[string]*RetractionReceipt
	isDetached  map[string]bool
}

// NewConsumerGraphTracker initializes a new ConsumerGraphTracker.
func NewConsumerGraphTracker() *ConsumerGraphTracker {
	return &ConsumerGraphTracker{
		currentEpoch:  1,
		consumerIndex: make(map[string][]*ConsumerRecord),
		projectEpoch:  make(map[string]int64),
		retractions:   make(map[string]*RetractionReceipt),
		isDetached:    make(map[string]bool),
	}
}

// SetProjectDetached marks a project as detached/offline.
func (cgt *ConsumerGraphTracker) SetProjectDetached(projectID string, detached bool) {
	cgt.mu.Lock()
	defer cgt.mu.Unlock()
	cgt.isDetached[projectID] = detached
}

// RecordConsumption registers a typed consumer of a knowledge/skill revision.
func (cgt *ConsumerGraphTracker) RecordConsumption(projectID string, cType ConsumerType, consumerID string, revisionRef string, historicalHash string) (*ConsumerRecord, error) {
	cgt.mu.Lock()
	defer cgt.mu.Unlock()

	if projectID == "" || consumerID == "" || revisionRef == "" {
		return nil, fmt.Errorf("projectID, consumerID, and revisionRef are required")
	}

	// Check if this revision has already been retracted
	if ret, retracted := cgt.retractions[revisionRef]; retracted {
		return nil, fmt.Errorf("CANNOT_CONSUME_RETRACTED_REVISION: revision %q was retracted at epoch %d (%s)", revisionRef, ret.Epoch, ret.RetractionReason)
	}

	rec := &ConsumerRecord{
		ProjectID:      projectID,
		ConsumerType:   cType,
		ConsumerID:     consumerID,
		RevisionRef:    revisionRef,
		PinnedEpoch:    cgt.currentEpoch,
		IsStale:        false,
		HistoricalHash: historicalHash,
		ConsumedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	cgt.consumerIndex[revisionRef] = append(cgt.consumerIndex[revisionRef], rec)
	if cgt.projectEpoch[projectID] < cgt.currentEpoch {
		cgt.projectEpoch[projectID] = cgt.currentEpoch
	}

	return rec, nil
}

// RetractRevision retracts a global revision, advances the epoch, marks all current descendants STALE,
// and emits an idempotent RetractionReceipt.
// Crucially, historical bytes and approval rows are preserved (not erased or rewritten).
func (cgt *ConsumerGraphTracker) RetractRevision(revisionRef string, reason string) (*RetractionReceipt, error) {
	cgt.mu.Lock()
	defer cgt.mu.Unlock()

	if revisionRef == "" {
		return nil, fmt.Errorf("revisionRef cannot be empty")
	}

	// Advance epoch
	cgt.currentEpoch++
	newEpoch := cgt.currentEpoch

	affectedProjectsMap := make(map[string]bool)
	consumers := cgt.consumerIndex[revisionRef]
	for _, c := range consumers {
		affectedProjectsMap[c.ProjectID] = true
		// If project has already acknowledged currentEpoch (i.e. is active/connected), mark immediately
		if cgt.projectEpoch[c.ProjectID] >= cgt.currentEpoch-1 && !cgt.isDetached[c.ProjectID] {
			c.IsStale = true
			c.StaleReason = fmt.Sprintf("RETRACTED_AT_EPOCH_%d: %s", newEpoch, reason)
		}
	}

	var affectedProjects []string
	for p := range affectedProjectsMap {
		affectedProjects = append(affectedProjects, p)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	recID := fmt.Sprintf("retract-%s-%d", revisionRef, newEpoch)

	raw, _ := json.Marshal(map[string]any{
		"receipt_id":        recID,
		"revision_ref":      revisionRef,
		"epoch":             newEpoch,
		"reason":            reason,
		"affected_projects": affectedProjects,
		"created_at":        now,
	})
	h := sha256.Sum256(raw)

	receipt := &RetractionReceipt{
		ReceiptID:        recID,
		RevisionRef:      revisionRef,
		Epoch:            newEpoch,
		RetractionReason: reason,
		AffectedProjects: affectedProjects,
		ReceiptHash:      hex.EncodeToString(h[:]),
		CreatedAt:        now,
	}

	cgt.retractions[revisionRef] = receipt
	return receipt, nil
}

// ReconcileOnReopen handles project reconnection / reopening after being offline/detached (X79).
// When an offline project reopens, it must reconcile its pinned epoch with the global registry.
// - If any consumed revisions were retracted while offline, current consumption is marked STALE.
// - Historical replay bytes remain preserved.
// - Returns list of stale consumers and whether current consumption is blocked.
func (cgt *ConsumerGraphTracker) ReconcileOnReopen(ctx context.Context, projectID string, isOffline bool) ([]*ConsumerRecord, bool, error) {
	cgt.mu.Lock()
	defer cgt.mu.Unlock()

	lastAckEpoch := cgt.projectEpoch[projectID]

	// If project is offline and unable to check current eligibility:
	// "An offline project unable to check current eligibility retains historical replay bytes
	// but cannot claim current global approval; use project-local evidence or leave that use blocked."
	if isOffline {
		// Find any consumers of this project
		var projectConsumers []*ConsumerRecord
		for _, list := range cgt.consumerIndex {
			for _, c := range list {
				if c.ProjectID == projectID {
					projectConsumers = append(projectConsumers, c)
				}
			}
		}
		// Returns blocked for new work/current approval, but historical replay retained
		return projectConsumers, true, nil
	}

	// Online reconciliation: compare against latest retractions
	var newlyStale []*ConsumerRecord
	for _, receipt := range cgt.retractions {
		if receipt.Epoch > lastAckEpoch {
			// A retraction occurred while project was detached
			for _, c := range cgt.consumerIndex[receipt.RevisionRef] {
				if c.ProjectID == projectID && !c.IsStale {
					c.IsStale = true
					c.StaleReason = fmt.Sprintf("RETRACTED_DURING_DETACHED_STATE: epoch %d: %s", receipt.Epoch, receipt.RetractionReason)
					newlyStale = append(newlyStale, c)
				}
			}
		}
	}

	// Update acknowledged epoch to current
	cgt.projectEpoch[projectID] = cgt.currentEpoch

	hasStale := len(newlyStale) > 0
	return newlyStale, hasStale, nil
}

// GetConsumers returns all recorded consumers for a revision.
func (cgt *ConsumerGraphTracker) GetConsumers(revisionRef string) []*ConsumerRecord {
	cgt.mu.RLock()
	defer cgt.mu.RUnlock()
	return cgt.consumerIndex[revisionRef]
}
