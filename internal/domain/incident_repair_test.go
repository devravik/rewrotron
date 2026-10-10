package domain

import (
	"strings"
	"testing"
	"time"
)

// TestIncidentRepair_X24_IdempotencyAndBoundedRepair tests Test X24:
// - Crash after commit before response returns existing result
// - Stale project/incident/lease cannot publish
// - Trivial context edit cannot reset three-attempt budget
// - Exhausting three attempts transitions incident to blocked state
func TestIncidentRepair_X24_IdempotencyAndBoundedRepair(t *testing.T) {
	t0 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	leaseDur := 5 * time.Minute

	// =========================================================================
	// Subtest 1: Crash after commit before response returns existing result
	// =========================================================================
	t.Run("X24_CrashAfterCommitReturnsExistingResult", func(t *testing.T) {
		svc := NewIncidentRepairService()
		projectID := "proj_crash_test"
		subjectID := "subject_session_fix"
		idempotencyKey := "idem_key_uuid_1234"

		// Begin attempt 1
		attempt, receipt, err := svc.BeginAttempt(
			projectID,
			subjectID,
			idempotencyKey,
			"fp_fix_1",
			"Fix session expiration parameter",
			map[string]any{"timeout": 3600},
			1,
			"worker_alpha",
			t0,
			leaseDur,
		)
		if err != nil {
			t.Fatalf("BeginAttempt failed: %v", err)
		}
		if receipt != nil {
			t.Fatalf("expected nil receipt on fresh attempt, got %v", receipt)
		}
		if attempt.AttemptNumber != 1 {
			t.Fatalf("expected attempt 1, got %d", attempt.AttemptNumber)
		}

		// Commit result and store receipt (simulating commit phase)
		resultPayload := map[string]any{
			"status":   "COMMITTED",
			"patch_id": "patch_sess_01",
			"files":    []string{"app/Session/Manager.php"},
		}
		committedReceipt, err := svc.CommitResultAndStoreReceipt(
			attempt,
			idempotencyKey,
			svc.GetProjectVersion(projectID),
			1,
			resultPayload,
			t0.Add(1*time.Minute),
		)
		if err != nil {
			t.Fatalf("CommitResultAndStoreReceipt failed: %v", err)
		}
		if committedReceipt.IdempotencyKey != idempotencyKey {
			t.Fatalf("receipt key mismatch: %s vs %s", committedReceipt.IdempotencyKey, idempotencyKey)
		}

		// Simulate client or process crash before response was acknowledged.
		// Client retries or recovers with the exact same idempotency key:
		retryAttempt, existingReceipt, err := svc.BeginAttempt(
			projectID,
			subjectID,
			idempotencyKey,
			"fp_fix_1",
			"Fix session expiration parameter",
			map[string]any{"timeout": 3600},
			1,
			"worker_alpha",
			t0.Add(2*time.Minute),
			leaseDur,
		)
		if err != nil {
			t.Fatalf("retried BeginAttempt failed: %v", err)
		}
		if retryAttempt != nil {
			t.Fatalf("expected nil attempt when idempotency receipt matches, got %v", retryAttempt)
		}
		if existingReceipt == nil {
			t.Fatalf("expected existing receipt to be returned on crash recovery")
		}
		if existingReceipt.ResponseResult["patch_id"] != "patch_sess_01" {
			t.Fatalf("expected existing committed result payload, got %v", existingReceipt.ResponseResult)
		}
	})

	// =========================================================================
	// Subtest 2: Three-attempt budget enforcement & trivial context edit rejection
	// =========================================================================
	t.Run("X24_ThreeAttemptBudgetAndTrivialContextRejection", func(t *testing.T) {
		svc := NewIncidentRepairService()
		projectID := "proj_budget_test"
		subjectID := "subject_csrf_mismatch"

		// Attempt 1: First try fails
		att1, _, err := svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_att1",
			"fp_csrf_v1",
			"Hypothesis 1: CSRF header case sensitivity",
			map[string]any{"header": "X-CSRF-TOKEN"},
			1,
			"worker_1",
			t0,
			leaseDur,
		)
		if err != nil {
			t.Fatalf("attempt 1 failed: %v", err)
		}
		if att1.AttemptNumber != 1 {
			t.Fatalf("expected attempt number 1, got %d", att1.AttemptNumber)
		}

		// Record failure of attempt 1
		err = svc.RecordFailure(projectID, subjectID, att1.AttemptID, "fp_csrf_v1", "Test MISMATCH", t0.Add(1*time.Minute))
		if err != nil {
			t.Fatalf("record failure 1 failed: %v", err)
		}

		// Attempt with identical fingerprint: rejected as trivial fingerprint reset
		_, _, err = svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_att1_dup",
			"fp_csrf_v1", // same fingerprint
			"Hypothesis 1 retry",
			map[string]any{"header": "X-CSRF-TOKEN"},
			1,
			"worker_1",
			t0.Add(2*time.Minute),
			leaseDur,
		)
		if err == nil || !strings.Contains(err.Error(), ErrTrivialFingerprintReset) {
			t.Fatalf("expected ErrTrivialFingerprintReset for identical fingerprint, got: %v", err)
		}

		// Attempt with trivial context edit (only comments/whitespace added): rejected
		_, _, err = svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_att1_trivial_ctx",
			"fp_csrf_v1_trivial",
			"", // no new hypothesis
			map[string]any{"header": "X-CSRF-TOKEN", "comment": "trivial comment edit"}, // non-semantic change
			1,
			"worker_1",
			t0.Add(2*time.Minute),
			leaseDur,
		)
		if err == nil || !strings.Contains(err.Error(), ErrTrivialFingerprintReset) {
			t.Fatalf("expected ErrTrivialFingerprintReset for trivial context edit, got: %v", err)
		}

		// Attempt 2: Genuine new hypothesis and distinct fingerprint
		att2, _, err := svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_att2",
			"fp_csrf_v2",
			"Hypothesis 2: Cookie same-site attribute mismatch",
			map[string]any{"header": "X-CSRF-TOKEN", "same_site": "Lax"},
			1,
			"worker_1",
			t0.Add(3*time.Minute),
			leaseDur,
		)
		if err != nil {
			t.Fatalf("attempt 2 failed: %v", err)
		}
		if att2.AttemptNumber != 2 {
			t.Fatalf("expected attempt number 2, got %d", att2.AttemptNumber)
		}
		_ = svc.RecordFailure(projectID, subjectID, att2.AttemptID, "fp_csrf_v2", "Test MISMATCH", t0.Add(4*time.Minute))

		// Attempt 3: Genuine third hypothesis
		att3, _, err := svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_att3",
			"fp_csrf_v3",
			"Hypothesis 3: Session encryption key rotation",
			map[string]any{"header": "X-CSRF-TOKEN", "encryption": "AES-256"},
			1,
			"worker_1",
			t0.Add(5*time.Minute),
			leaseDur,
		)
		if err != nil {
			t.Fatalf("attempt 3 failed: %v", err)
		}
		if att3.AttemptNumber != 3 {
			t.Fatalf("expected attempt number 3, got %d", att3.AttemptNumber)
		}
		_ = svc.RecordFailure(projectID, subjectID, att3.AttemptID, "fp_csrf_v3", "Test MISMATCH", t0.Add(6*time.Minute))

		// Attempt 4: Should be blocked by 3-attempt budget
		_, _, err = svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_att4",
			"fp_csrf_v4",
			"Hypothesis 4: Another try",
			map[string]any{"header": "X-CSRF-TOKEN", "extra": "data"},
			1,
			"worker_1",
			t0.Add(7*time.Minute),
			leaseDur,
		)
		if err == nil || !strings.Contains(err.Error(), ErrIncidentBudgetExceeded) {
			t.Fatalf("expected ErrIncidentBudgetExceeded on attempt 4, got: %v", err)
		}

		// Verify incident subject is blocked with review requirement
		inc, err := svc.RegisterIncident(projectID, subjectID, t0.Add(8*time.Minute))
		if err != nil {
			t.Fatalf("RegisterIncident failed: %v", err)
		}
		if !inc.IsBlocked {
			t.Fatalf("expected incident to be blocked after 3 failed attempts")
		}
		if inc.ReviewBlockerReason != "INCIDENT_BUDGET_EXHAUSTED" {
			t.Fatalf("expected review blocker reason, got: %s", inc.ReviewBlockerReason)
		}

		// Reset budget requires explicit reviewed incident cause
		err = svc.ResetBudgetWithReviewedCause(projectID, subjectID, "", t0.Add(9*time.Minute))
		if err == nil {
			t.Fatalf("expected error resetting budget with empty review comment")
		}

		err = svc.ResetBudgetWithReviewedCause(projectID, subjectID, "Human verified root cause: upstream gateway strips cookie header", t0.Add(9*time.Minute))
		if err != nil {
			t.Fatalf("ResetBudgetWithReviewedCause failed: %v", err)
		}

		// Now attempt can be admitted under fresh budget
		attPostReset, _, err := svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_csrf_post_reset",
			"fp_csrf_v5",
			"Reviewed fix: gateway header propagation",
			map[string]any{"gateway_header": true},
			1,
			"worker_1",
			t0.Add(10*time.Minute),
			leaseDur,
		)
		if err != nil {
			t.Fatalf("attempt post reset failed: %v", err)
		}
		if attPostReset.AttemptNumber != 1 {
			t.Fatalf("expected attempt number 1 after reset, got %d", attPostReset.AttemptNumber)
		}
	})

	// =========================================================================
	// Subtest 3: Stale project/incident/lease publication rejection
	// =========================================================================
	t.Run("X24_StaleProjectIncidentLeasePublicationRejection", func(t *testing.T) {
		svc := NewIncidentRepairService()
		projectID := "proj_stale_test"
		subjectID := "subject_stale_check"

		svc.SetProjectVersion(projectID, 1)

		att, _, err := svc.BeginAttempt(
			projectID,
			subjectID,
			"idem_stale_1",
			"fp_stale_1",
			"Fixing database column",
			map[string]any{"col": "active"},
			1,
			"worker_1",
			t0,
			leaseDur,
		)
		if err != nil {
			t.Fatalf("BeginAttempt failed: %v", err)
		}

		// Case A: Stale project publication check
		// Project version advances externally (e.g., someone committed another migration or rolled back)
		svc.SetProjectVersion(projectID, 2)
		_, err = svc.CommitResultAndStoreReceipt(
			att,
			"idem_stale_1",
			1, // expecting version 1, but project is at 2
			1,
			map[string]any{"status": "OK"},
			t0.Add(1*time.Minute),
		)
		if err == nil || !strings.Contains(err.Error(), ErrStaleProjectPublication) {
			t.Fatalf("expected ErrStaleProjectPublication, got: %v", err)
		}

		// Restore project version for next sub-check
		svc.SetProjectVersion(projectID, 1)

		// Case B: Stale lease publication check
		// Publication happens after lease expiration time
		pastExpiry := t0.Add(6 * time.Minute) // lease duration was 5m
		_, err = svc.CommitResultAndStoreReceipt(
			att,
			"idem_stale_1",
			1,
			1,
			map[string]any{"status": "OK"},
			pastExpiry,
		)
		if err == nil || !strings.Contains(err.Error(), ErrStaleLeasePublication) {
			t.Fatalf("expected ErrStaleLeasePublication for expired lease, got: %v", err)
		}

		// Case C: Stale fencing token check
		// Fencing token in project was superseded
		_, err = svc.CommitResultAndStoreReceipt(
			att,
			"idem_stale_1",
			1,
			2, // current fencing token is 2, att has token 1
			map[string]any{"status": "OK"},
			t0.Add(2*time.Minute),
		)
		if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) {
			t.Fatalf("expected ErrStaleFenceToken for superseded fencing token, got: %v", err)
		}
	})
}
