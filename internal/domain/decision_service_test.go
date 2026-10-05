package domain

import (
	"strings"
	"testing"
)

// TestDecisionService_ActionEnums verifies all valid decision actions and strategy selections
// conforming to specs/project/decision.schema.json.
func TestDecisionService_ActionEnums(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("1", 64)
	altRef := "sha256:" + strings.Repeat("2", 64)
	scopeRef := "sha256:" + strings.Repeat("3", 64)

	// 1. REJECT recommendation
	t.Run("Action_REJECT", func(t *testing.T) {
		req := &ProposeDecisionRequest{
			DecisionID:   "dec_reject_rec",
			ProjectID:    "proj_test",
			Action:       ActionReject,
			SubjectRef:   dummyRef,
			Rationale:    "Legacy caching dependency is obsolete in target environment",
			EvidenceRefs: []string{dummyRef},
		}
		doc, err := ProposeDecision(req)
		if err != nil {
			t.Fatalf("ProposeDecision failed: %v", err)
		}
		if doc.Action != "REJECT" || doc.Status != "PROPOSED" {
			t.Fatalf("unexpected doc: %+v", doc)
		}
	})

	// 2. SELECT_STRATEGY with ADAPTER
	t.Run("Action_SELECT_STRATEGY", func(t *testing.T) {
		strat := StrategyAdapter
		req := &ProposeDecisionRequest{
			DecisionID:   "dec_select_adapter",
			ProjectID:    "proj_test",
			Action:       ActionSelectStrategy,
			SubjectRef:   dummyRef,
			Strategy:     &strat,
			Rationale:    "Encapsulate legacy authentication semantics behind Laravel Auth adapter",
			EvidenceRefs: []string{dummyRef},
		}
		doc, err := ProposeDecision(req)
		if err != nil {
			t.Fatalf("ProposeDecision failed: %v", err)
		}
		if doc.Strategy == nil || *doc.Strategy != "ADAPTER" {
			t.Fatalf("expected ADAPTER strategy, got: %v", doc.Strategy)
		}
	})

	// 3. SUPPLY_ALTERNATIVE
	t.Run("Action_SUPPLY_ALTERNATIVE", func(t *testing.T) {
		req := &ProposeDecisionRequest{
			DecisionID:        "dec_supply_alt",
			ProjectID:         "proj_test",
			Action:            ActionSupplyAlternative,
			SubjectRef:        dummyRef,
			ImplementationRef: &altRef,
			Rationale:         "Supply custom Redis session handler instead of generic array cache",
			EvidenceRefs:      []string{dummyRef},
		}
		doc, err := ProposeDecision(req)
		if err != nil {
			t.Fatalf("ProposeDecision failed: %v", err)
		}
		if doc.ImplementationRef == nil || *doc.ImplementationRef != altRef {
			t.Fatalf("expected implementation ref %s, got: %v", altRef, doc.ImplementationRef)
		}
	})

	// 4. REQUEST_EVIDENCE
	t.Run("Action_REQUEST_EVIDENCE", func(t *testing.T) {
		req := &ProposeDecisionRequest{
			DecisionID:   "dec_req_evidence",
			ProjectID:    "proj_test",
			Action:       ActionRequestEvidence,
			SubjectRef:   dummyRef,
			Rationale:    "Request differential trace evidence for legacy MySQL spatial queries",
			EvidenceRefs: []string{dummyRef},
		}
		doc, err := ProposeDecision(req)
		if err != nil {
			t.Fatalf("ProposeDecision failed: %v", err)
		}
		if doc.Action != "REQUEST_EVIDENCE" {
			t.Fatalf("expected REQUEST_EVIDENCE, got: %s", doc.Action)
		}
	})

	// 5. APPROVE_EXCEPTION
	t.Run("Action_APPROVE_EXCEPTION", func(t *testing.T) {
		req := &ProposeDecisionRequest{
			DecisionID:        "dec_approve_exc",
			ProjectID:         "proj_test",
			Action:            ActionApproveException,
			SubjectRef:        dummyRef,
			ExceptionScopeRef: &scopeRef,
			Rationale:         "Exclude deprecated flashdata cookie behavior from migration scope",
			EvidenceRefs:      []string{dummyRef},
		}
		doc, err := ProposeDecision(req)
		if err != nil {
			t.Fatalf("ProposeDecision failed: %v", err)
		}
		if doc.ExceptionScopeRef == nil || *doc.ExceptionScopeRef != scopeRef {
			t.Fatalf("expected exception scope ref %s, got: %v", scopeRef, doc.ExceptionScopeRef)
		}
	})
}

// TestDecisionService_X16_ExceptionCannotGrantPass tests Test X16 invariant:
// A proposed exception cannot grant a PASS verdict or safety waiver;
// UNKNOWN or MISMATCH cases cannot be turned into PASS by an exception.
func TestDecisionService_X16_ExceptionCannotGrantPass(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("4", 64)
	scopeRef := "sha256:" + strings.Repeat("5", 64)

	doc, err := ProposeDecision(&ProposeDecisionRequest{
		DecisionID:        "dec_exc_test",
		ProjectID:         "proj_test",
		Action:            ActionApproveException,
		SubjectRef:        dummyRef,
		ExceptionScopeRef: &scopeRef,
		Rationale:         "Exclude edge case payment gateway timeout",
		EvidenceRefs:      []string{dummyRef},
	})
	if err != nil {
		t.Fatalf("ProposeDecision failed: %v", err)
	}

	// Attempting to use the exception to assert PASS must fail
	err = ValidateDecisionExceptionSafety(doc, "PASS")
	if err == nil {
		t.Fatal("expected error: proposed exception cannot grant PASS verdict (Test X16), got nil")
	}
	if !strings.Contains(err.Error(), "proposed exception cannot grant PASS") {
		t.Fatalf("expected PASS waiver rejection message, got: %v", err)
	}

	// Exceptions adjusting scope for UNKNOWN/MISMATCH without granting PASS is allowed
	if err := ValidateDecisionExceptionSafety(doc, "UNKNOWN"); err != nil {
		t.Fatalf("expected no error for UNKNOWN, got: %v", err)
	}
	if err := ValidateDecisionExceptionSafety(doc, "MISMATCH"); err != nil {
		t.Fatalf("expected no error for MISMATCH, got: %v", err)
	}
}

// TestDecisionService_X16_OnlyExactGate2Activates tests Test X16 invariant:
// A decision proposal has status strictly PROPOSED; it never self-activates or bypasses Gate 2.
func TestDecisionService_X16_OnlyExactGate2Activates(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("6", 64)
	strat := StrategyDirectUpgrade

	doc, err := ProposeDecision(&ProposeDecisionRequest{
		DecisionID:   "dec_gate2_check",
		ProjectID:    "proj_test",
		Action:       ActionSelectStrategy,
		SubjectRef:   dummyRef,
		Strategy:     &strat,
		Rationale:    "Upgrade Guzzle client",
		EvidenceRefs: []string{dummyRef},
	})
	if err != nil {
		t.Fatalf("ProposeDecision failed: %v", err)
	}

	// Invariant: Status must strictly be PROPOSED
	if doc.Status != "PROPOSED" {
		t.Fatalf("expected status PROPOSED, got: %s", doc.Status)
	}
}

// TestDecisionService_X17_ProgressLeavesPlanHashUnchanged tests Test X17 invariant:
// PLANNED -> GENERATING -> BUILDING -> VERIFYING -> VERIFIED progress updates
// leave the original plan hash completely unchanged.
func TestDecisionService_X17_ProgressLeavesPlanHashUnchanged(t *testing.T) {
	planHash := strings.Repeat("a", 64)
	decisionRef := "sha256:" + strings.Repeat("7", 64)
	treeRef := "sha256:" + strings.Repeat("8", 64)
	repRef := "sha256:" + strings.Repeat("9", 64)

	lifecycle := []MigrationProgressStatus{
		ProgressPlanned,
		ProgressGenerating,
		ProgressBuilding,
		ProgressVerifying,
		ProgressVerified,
	}

	var progressDocs []*MigrationProgressDocument
	for seq, st := range lifecycle {
		var tt *string
		var ar *string
		if st == ProgressBuilding || st == ProgressVerifying || st == ProgressVerified {
			tt = &treeRef
		}
		if st == ProgressVerified {
			ar = &repRef
		}

		doc, err := RecordMigrationProgress(&RecordProgressRequest{
			ProjectID:           "proj_test",
			PlanHash:            planHash,
			DecisionOrModelRef:  decisionRef,
			EntityID:            "unit_auth",
			Status:              st,
			ThroughSequence:     seq + 1,
			TargetTreeRef:       tt,
			AcceptanceReportRef: ar,
		})
		if err != nil {
			t.Fatalf("RecordMigrationProgress failed at step %s: %v", st, err)
		}
		progressDocs = append(progressDocs, doc)
	}

	// Verify plan hash stability across all progress transitions
	stable, err := VerifyPlanHashStability(planHash, progressDocs)
	if err != nil || !stable {
		t.Fatalf("VerifyPlanHashStability failed (Test X17): %v", err)
	}
}

// TestDecisionService_X17_StrategyChangeInvalidatesApproval tests Test X17 invariant:
// When a decision's strategy or lowering changes, any existing approval whose decision_refs
// contains the old decision ref is invalidated.
func TestDecisionService_X17_StrategyChangeInvalidatesApproval(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)

	// Decision version 1: DIRECT_UPGRADE
	strat1 := StrategyDirectUpgrade
	dec1, err := ProposeDecision(&ProposeDecisionRequest{
		DecisionID:   "dec_monitored",
		ProjectID:    "proj_test",
		Action:       ActionSelectStrategy,
		SubjectRef:   dummyRef,
		Strategy:     &strat1,
		Rationale:    "Direct upgrade",
		EvidenceRefs: []string{dummyRef},
	})
	if err != nil {
		t.Fatalf("ProposeDecision dec1 failed: %v", err)
	}
	digest1, err := dec1.Digest()
	if err != nil {
		t.Fatalf("dec1 Digest failed: %v", err)
	}

	// Plan approved with digest1 in decision_refs
	planApprovedDecisionRefs := []string{digest1}

	// Decision version 2: strategy changed to ADAPTER
	strat2 := StrategyAdapter
	dec2, err := ProposeDecision(&ProposeDecisionRequest{
		DecisionID:   "dec_monitored",
		ProjectID:    "proj_test",
		Action:       ActionSelectStrategy,
		SubjectRef:   dummyRef,
		Strategy:     &strat2,
		Rationale:    "Changed strategy to adapter due to breaking changes in target",
		EvidenceRefs: []string{dummyRef},
	})
	if err != nil {
		t.Fatalf("ProposeDecision dec2 failed: %v", err)
	}
	digest2, err := dec2.Digest()
	if err != nil {
		t.Fatalf("dec2 Digest failed: %v", err)
	}

	if digest1 == digest2 {
		t.Fatal("expected digest change when strategy changes")
	}

	// Check invalidation
	isInvalidated, reason := CheckApprovalInvalidationOnDecisionChange(planApprovedDecisionRefs, digest1, digest2)
	if !isInvalidated {
		t.Fatalf("expected plan approval to be invalidated when decision strategy changes (Test X17), reason: %s", reason)
	}
	if !strings.Contains(reason, "plan approval invalidated (Test X17)") {
		t.Fatalf("expected Test X17 invalidation reason, got: %s", reason)
	}
}
