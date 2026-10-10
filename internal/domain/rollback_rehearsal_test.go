package domain

import (
	"strings"
	"testing"
)

func makeValidRehearsal() *RollbackRehearsalRecord {
	return &RollbackRehearsalRecord{
		SchemaVersion:  "1.0",
		RehearsalID:    "rehearsal-001",
		ProjectID:      "proj-fixture",
		TargetTreeRef:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EnvironmentRef: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ExecutableCommands: []string{
			"systemctl stop target-app",
			"systemctl start legacy-source-app",
			"/usr/local/bin/reconcile-sessions-and-orders.sh",
		},
		RepresentativeMutations: []TargetMutation{
			{
				MutationID:   "mut-db-1",
				MutationType: MutationTypeDatabaseRow,
				TargetRef:    "orders",
				PayloadHash:  "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				Reversible:   false,
			},
			{
				MutationID:   "mut-sess-1",
				MutationType: MutationTypeSessionState,
				TargetRef:    "user_session_42",
				PayloadHash:  "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				Reversible:   false,
			},
			{
				MutationID:   "mut-pay-1",
				MutationType: MutationTypeExternalReceipt,
				TargetRef:    "payment_receipt_999",
				PayloadHash:  "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
				Reversible:   false,
			},
		},
		ReconciliationContracts: []ReconciliationContract{
			{
				MutationType:   MutationTypeDatabaseRow,
				ReconcileOwner: "db-admin",
				Strategy:       "REPLAY_SOURCE_FORWARD",
				Verified:       true,
			},
			{
				MutationType:   MutationTypeSessionState,
				ReconcileOwner: "auth-service",
				Strategy:       "SESSION_BRIDGE_ROTATE",
				Verified:       true,
			},
			{
				MutationType:   MutationTypeExternalReceipt,
				ReconcileOwner: "payment-gateway-reconciler",
				Strategy:       "IDEMPOTENT_RECHECK",
				Verified:       true,
			},
		},
		Triggers: []RollbackTrigger{
			{
				TriggerID:  "trig-error",
				Metric:     "ERROR_RATE",
				Threshold:  "> 0.01",
				DecisionBy: "release-engineer",
			},
		},
		SchemaOnlyGuaranteed:     false,
		SourceContinuityVerified: true,
		Outcome:                  RollbackRehearsalPass,
		ReasonCodes:              []string{"REHEARSAL_SUCCESS"},
	}
}

// Test X34: Write-aware rollback & target data reconciliation
// "Target changes sessions/orders/payment simulator state before rollback;
//  source continuity/data reconciliation tested or release blocks"
func TestRollbackRehearsal_X34_WriteAwareRollback(t *testing.T) {
	// 1. Positive control: Full write-aware rehearsal with all mutations reconciled passes
	valid := makeValidRehearsal()
	if err := ValidateRollbackRehearsal(valid); err != nil {
		t.Fatalf("expected valid rehearsal to pass, got: %v", err)
	}

	// Verify canonical digest calculation
	digest, err := valid.Digest()
	if err != nil {
		t.Fatalf("failed to calculate rehearsal digest: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("invalid digest format: %s", digest)
	}

	// 2. Mutant: Schema-only rollback claim ("preserved schema guarantees rollback")
	schemaOnly := makeValidRehearsal()
	schemaOnly.SchemaOnlyGuaranteed = true
	err = ValidateRollbackRehearsal(schemaOnly)
	if err == nil || !strings.Contains(err.Error(), "schema-only rollback guarantee rejected") {
		t.Fatalf("expected refusal of schema-only rollback claim, got: %v", err)
	}

	// 3. Mutant: Empty representative target mutations (empty-data traffic switch)
	emptyMutations := makeValidRehearsal()
	emptyMutations.RepresentativeMutations = nil
	err = ValidateRollbackRehearsal(emptyMutations)
	if err == nil || !strings.Contains(err.Error(), "no representative target mutations rehearsed") {
		t.Fatalf("expected refusal of empty mutations rehearsal, got: %v", err)
	}

	// 4. Mutant: Missing executable operator commands
	noCommands := makeValidRehearsal()
	noCommands.ExecutableCommands = nil
	err = ValidateRollbackRehearsal(noCommands)
	if err == nil || !strings.Contains(err.Error(), "missing executable operator rollback commands") {
		t.Fatalf("expected refusal when executable commands are missing, got: %v", err)
	}

	// 5. Mutant: Unreconciled target writes (e.g. payment mutation missing reconciliation contract)
	unreconciledPayment := makeValidRehearsal()
	// Remove ExternalReceipt contract
	unreconciledPayment.ReconciliationContracts = []ReconciliationContract{
		{
			MutationType:   MutationTypeDatabaseRow,
			ReconcileOwner: "db-admin",
			Strategy:       "REPLAY_SOURCE_FORWARD",
			Verified:       true,
		},
		{
			MutationType:   MutationTypeSessionState,
			ReconcileOwner: "auth-service",
			Strategy:       "SESSION_BRIDGE_ROTATE",
			Verified:       true,
		},
	}
	err = ValidateRollbackRehearsal(unreconciledPayment)
	if err == nil || !strings.Contains(err.Error(), "unreconciled target mutation of kind EXTERNAL_RECEIPT") {
		t.Fatalf("expected refusal of unreconciled target mutation, got: %v", err)
	}

	// 6. Mutant: Unverified reconciliation contract
	unverifiedContract := makeValidRehearsal()
	unverifiedContract.ReconciliationContracts[0].Verified = false
	err = ValidateRollbackRehearsal(unverifiedContract)
	if err == nil || !strings.Contains(err.Error(), "unreconciled target mutation") {
		t.Fatalf("expected refusal when reconciliation contract is not verified, got: %v", err)
	}

	// 7. Mutant: Source continuity not verified after rollback
	noSourceContinuity := makeValidRehearsal()
	noSourceContinuity.SourceContinuityVerified = false
	err = ValidateRollbackRehearsal(noSourceContinuity)
	if err == nil || !strings.Contains(err.Error(), "source continuity after rollback is not verified") {
		t.Fatalf("expected refusal when source continuity is unverified, got: %v", err)
	}

	// 8. Mutant: Missing rollback triggers
	noTriggers := makeValidRehearsal()
	noTriggers.Triggers = nil
	err = ValidateRollbackRehearsal(noTriggers)
	if err == nil || !strings.Contains(err.Error(), "no concrete rollback triggers defined") {
		t.Fatalf("expected refusal when triggers are missing, got: %v", err)
	}

	// 9. Mutant: Failed outcome
	failedRehearsal := makeValidRehearsal()
	failedRehearsal.Outcome = RollbackRehearsalFail
	err = ValidateRollbackRehearsal(failedRehearsal)
	if err == nil || !strings.Contains(err.Error(), "rehearsal outcome is FAIL") {
		t.Fatalf("expected refusal on failed outcome, got: %v", err)
	}
}
