package domain

import (
	"context"
	"strings"
	"testing"
)

func TestConsumerTracker_RetractionAndOfflineReconciliation(t *testing.T) {
	ctx := context.Background()
	tracker := NewConsumerGraphTracker()

	// 1. Record consumption of global revision in Project A and Project B
	rev1 := "pattern-ci3-session:v1"
	histHashA := "sha256:aaaa1111222233334444555566667777888899990000aaaa1111222233334444"
	histHashB := "sha256:bbbb1111222233334444555566667777888899990000bbbb1111222233334444"

	c1, err := tracker.RecordConsumption("proj-a", ConsumerArtifact, "art-a1", rev1, histHashA)
	if err != nil {
		t.Fatalf("failed to record consumption: %v", err)
	}
	if c1.IsStale {
		t.Fatalf("new consumption should not be stale")
	}

	_, err = tracker.RecordConsumption("proj-b", ConsumerUnit, "unit-b1", rev1, histHashB)
	if err != nil {
		t.Fatalf("failed to record consumption for proj-b: %v", err)
	}

	// Record another unaffected revision in Project A
	rev2 := "pattern-csrf-token:v1"
	c3, _ := tracker.RecordConsumption("proj-a", ConsumerArtifact, "art-a2", rev2, "sha256:cccc...")

	// 2. Retract rev1 globally (X39)
	receipt, err := tracker.RetractRevision(rev1, "Security flaw detected in legacy cipher handling")
	if err != nil {
		t.Fatalf("failed to retract revision: %v", err)
	}
	if receipt.ReceiptHash == "" {
		t.Fatalf("expected non-empty receipt hash")
	}
	if len(receipt.AffectedProjects) != 2 {
		t.Fatalf("expected 2 affected projects, got %v", receipt.AffectedProjects)
	}

	// Verify all current consumers of rev1 are marked STALE
	consumers := tracker.GetConsumers(rev1)
	for _, c := range consumers {
		if !c.IsStale {
			t.Fatalf("expected consumer %s to be marked STALE", c.ConsumerID)
		}
		if !strings.Contains(c.StaleReason, "RETRACTED_AT_EPOCH") {
			t.Fatalf("expected StaleReason with epoch, got: %s", c.StaleReason)
		}
		// Invariant: Historical hash bytes must remain preserved for reproducible replay
		if c.HistoricalHash == "" {
			t.Fatalf("historical hash was erased!")
		}
	}

	// Verify unaffected rev2 consumer remains valid (not stale)
	if c3.IsStale {
		t.Fatalf("unaffected revision consumer was erroneously marked stale")
	}

	// Attempting to consume the retracted revision for new work is refused
	_, err = tracker.RecordConsumption("proj-c", ConsumerArtifact, "art-c1", rev1, "sha256:dddd...")
	if err == nil || !strings.Contains(err.Error(), "CANNOT_CONSUME_RETRACTED_REVISION") {
		t.Fatalf("expected CANNOT_CONSUME_RETRACTED_REVISION, got: %v", err)
	}

	// 3. Detached / Offline project reconciliation (X79)
	// Project D consumes rev3 at epoch 2
	rev3 := "skill-form-helper:v1"
	_, _ = tracker.RecordConsumption("proj-d", ConsumerResult, "res-d1", rev3, "sha256:eeee...")

	// Project D is detached/offline. While detached, rev3 is retracted globally.
	tracker.SetProjectDetached("proj-d", true)
	_, _ = tracker.RetractRevision(rev3, "Upstream syntax deprecation")

	// Proj-D attempts offline reopen check: cannot claim current global approval
	offlineConsumers, isBlocked, err := tracker.ReconcileOnReopen(ctx, "proj-d", true)
	if err != nil {
		t.Fatalf("unexpected reopen error: %v", err)
	}
	if !isBlocked {
		t.Fatalf("expected offline project to be blocked from claiming current global approval")
	}
	if len(offlineConsumers) == 0 {
		t.Fatalf("expected historical consumers to be returned for replay")
	}

	// Proj-D reconnects online: reconciles against latest retraction epoch
	newlyStale, hasStale, err := tracker.ReconcileOnReopen(ctx, "proj-d", false)
	if err != nil {
		t.Fatalf("unexpected online reconciliation error: %v", err)
	}
	if !hasStale || len(newlyStale) != 1 {
		t.Fatalf("expected 1 newly stale consumer discovered upon reopen, got %d", len(newlyStale))
	}
	if !strings.Contains(newlyStale[0].StaleReason, "RETRACTED_DURING_DETACHED_STATE") {
		t.Fatalf("expected RETRACTED_DURING_DETACHED_STATE reason, got: %s", newlyStale[0].StaleReason)
	}
}
