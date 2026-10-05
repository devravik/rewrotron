package domain

import (
	"strings"
	"testing"
)

// TestX20_NormalizationPolicyValidation verifies that normalization rules require
// complete evidence (negative_mutation_refs, decision_ref, lifecycle_contract_ref, business_significance_ref).
func TestX20_NormalizationPolicyValidation(t *testing.T) {
	t.Run("ValidRule_PassesValidation", func(t *testing.T) {
		rules := []*NormalizationRule{
			{
				SchemaVersion:           "2.0",
				RuleID:                  "norm_uuid",
				Path:                    "/id",
				ValueType:               "string",
				Variability:             "GENERATED_ID",
				Relation:                "BIJECTIVE_ID",
				RelationGroup:           "user_identity",
				LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
				BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
				DecisionRef:             "sha256:" + strings.Repeat("c", 64),
				NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
			},
		}
		errs := ValidateNormalizationPolicy(rules)
		if len(errs) != 0 {
			t.Fatalf("expected no policy errors for valid rule, got: %v", errs)
		}
	})

	t.Run("MissingNegativeMutationRefs_PolicyViolation", func(t *testing.T) {
		rules := []*NormalizationRule{
			{
				SchemaVersion:           "2.0",
				RuleID:                  "norm_uuid",
				Path:                    "/id",
				ValueType:               "string",
				Variability:             "GENERATED_ID",
				Relation:                "BIJECTIVE_ID",
				RelationGroup:           "user_identity",
				LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
				BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
				DecisionRef:             "sha256:" + strings.Repeat("c", 64),
				NegativeMutationRefs:    nil, // missing!
			},
		}
		errs := ValidateNormalizationPolicy(rules)
		if len(errs) == 0 {
			t.Fatalf("expected policy error for missing negative_mutation_refs")
		}
	})

	t.Run("MissingDecisionRef_PolicyViolation", func(t *testing.T) {
		rules := []*NormalizationRule{
			{
				SchemaVersion:           "2.0",
				RuleID:                  "norm_clock",
				Path:                    "/created_at",
				ValueType:               "timestamp",
				Variability:             "CLOCK",
				Relation:                "CONTROLLED_CLOCK",
				RelationGroup:           "created_at_group",
				LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
				BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
				DecisionRef:             "", // missing!
				NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
			},
		}
		errs := ValidateNormalizationPolicy(rules)
		if len(errs) == 0 {
			t.Fatalf("expected policy error for missing decision_ref")
		}
	})
}

// TestX20_BrokenForeignKey verifies that a broken foreign key survives normalization as MISMATCH.
func TestX20_BrokenForeignKey(t *testing.T) {
	rules := []*NormalizationRule{
		{
			SchemaVersion:           "2.0",
			RuleID:                  "norm_uuid",
			Path:                    "/id",
			ValueType:               "string",
			Variability:             "GENERATED_ID",
			Relation:                "BIJECTIVE_ID",
			RelationGroup:           "user_identity",
			LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
			BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
			DecisionRef:             "sha256:" + strings.Repeat("c", 64),
			NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
		},
	}

	// Broken FK: source says user_id=42 links to order, target says user_id=99 (different entity)
	mutant := NormalizationMutant{
		MutantID:    "broken-fk-user-order",
		Category:    "BROKEN_FOREIGN_KEY",
		Description: "Source order.user_id=42 links to User[42]; target order.user_id=99 links to User[99]: broken FK relationship",
		SourceValue: TaggedValue{Type: ValueTypeInteger, Value: "42"},
		TargetValue: TaggedValue{Type: ValueTypeInteger, Value: "99"},
		Context: map[string]string{
			"path":           "$.data.orders[0].user_id",
			"source_user_id": "42",
			"target_user_id": "99",
		},
	}

	protector := NewNormalizerMutationProtector(rules)
	result := protector.CheckMutant(mutant)

	if result.Verdict != "MISMATCH" {
		t.Fatalf("X20: broken FK must survive normalization as MISMATCH, got %s; reason: %s",
			result.Verdict, result.ViolationReason)
	}
	if result.WasNormalized {
		t.Fatalf("X20: normalization must NOT absorb broken FK")
	}
}

// TestX20_WrongPrincipal verifies that a wrong session principal survives normalization as MISMATCH.
func TestX20_WrongPrincipal(t *testing.T) {
	rules := []*NormalizationRule{
		{
			SchemaVersion:           "2.0",
			RuleID:                  "norm_session_token",
			Path:                    "/session_token",
			ValueType:               "string",
			Variability:             "TOKEN",
			Relation:                "TOKEN_LIFECYCLE",
			RelationGroup:           "session_group",
			LifecycleContractRef:    "sha256:" + strings.Repeat("e", 64),
			BusinessSignificanceRef: "sha256:" + strings.Repeat("f", 64),
			DecisionRef:             "sha256:" + strings.Repeat("g", 64),
			NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("h", 64)},
		},
	}

	// Wrong principal: source response session belongs to alice, target response session belongs to bob
	mutant := NormalizationMutant{
		MutantID:    "wrong-principal-alice-bob",
		Category:    "WRONG_PRINCIPAL",
		Description: "Source session principal=alice; target session principal=bob (cross-user contamination)",
		SourceValue: TaggedValue{Type: ValueTypeString, Value: "alice"},
		TargetValue: TaggedValue{Type: ValueTypeString, Value: "bob"},
		Context: map[string]string{
			"path":           "$.session.subject",
			"source_subject": "alice",
			"target_subject": "bob",
		},
	}

	protector := NewNormalizerMutationProtector(rules)
	result := protector.CheckMutant(mutant)

	if result.Verdict != "MISMATCH" {
		t.Fatalf("X20: wrong principal must survive normalization as MISMATCH, got %s", result.Verdict)
	}
	if result.WasNormalized {
		t.Fatalf("X20: TOKEN normalization must NOT absorb wrong principal")
	}
}

// TestX20_ExpiredToken verifies that an expired/replayed token survives normalization.
func TestX20_ExpiredToken(t *testing.T) {
	rules := []*NormalizationRule{
		{
			SchemaVersion:           "2.0",
			RuleID:                  "norm_session_clock",
			Path:                    "/session/expires_at",
			ValueType:               "timestamp",
			Variability:             "CLOCK",
			Relation:                "CONTROLLED_CLOCK",
			RelationGroup:           "session_group",
			LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
			BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
			DecisionRef:             "sha256:" + strings.Repeat("c", 64),
			NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
		},
	}

	// Expired token: source sees valid (active) token, target replays expired/invalid token
	mutant := NormalizationMutant{
		MutantID:    "expired-csrf-replay",
		Category:    "EXPIRED_TOKEN",
		Description: "Source CSRF token is valid; target re-uses expired/replayed CSRF token",
		SourceValue: TaggedValue{Type: ValueTypeString, Value: "CSRF_VALID_ACCEPTED"},
		TargetValue: TaggedValue{Type: ValueTypeString, Value: "CSRF_EXPIRED_REJECTED"},
		Context: map[string]string{
			"is_expiry_boundary": "true", // marks this as a lifecycle boundary, not a generic clock
			"path":               "$.session.csrf_outcome",
		},
	}

	protector := NewNormalizerMutationProtector(rules)
	result := protector.CheckMutant(mutant)

	// Expiry boundary must NOT be absorbed by CLOCK rules
	if result.Verdict != "MISMATCH" {
		t.Fatalf("X20: expired/replayed token must survive normalization as MISMATCH, got %s; reason: %s",
			result.Verdict, result.ViolationReason)
	}
}

// TestX20_RefundClockBoundary verifies that business refund timestamp boundaries survive normalization.
func TestX20_RefundClockBoundary(t *testing.T) {
	rules := []*NormalizationRule{
		{
			SchemaVersion:           "2.0",
			RuleID:                  "norm_generic_clock",
			Path:                    "/created_at",
			ValueType:               "timestamp",
			Variability:             "CLOCK",
			Relation:                "CONTROLLED_CLOCK",
			RelationGroup:           "generic_clock",
			LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
			BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
			DecisionRef:             "sha256:" + strings.Repeat("c", 64),
			NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
		},
	}

	// Refund boundary: source correctly rejects refund (outside 30-day window); target incorrectly accepts
	mutant := NormalizationMutant{
		MutantID:    "refund-boundary-violation",
		Category:    "REFUND_BOUNDARY",
		Description: "Source rejects refund (31 days past purchase); target incorrectly accepts refund",
		SourceValue: TaggedValue{Type: ValueTypeString, Value: "REFUND_REJECTED"},
		TargetValue: TaggedValue{Type: ValueTypeString, Value: "REFUND_ACCEPTED"},
		Context: map[string]string{
			"path":            "$.refund.outcome",
			"days_since_purchase": "31",
			"policy_window":   "30",
		},
	}

	protector := NewNormalizerMutationProtector(rules)
	result := protector.CheckMutant(mutant)

	// Refund boundary decisions cannot be masked by CLOCK normalization
	if result.Verdict != "MISMATCH" {
		t.Fatalf("X20: refund clock boundary must survive normalization as MISMATCH, got %s; reason: %s",
			result.Verdict, result.ViolationReason)
	}
	if result.WasNormalized {
		t.Fatalf("X20: CLOCK normalization must NOT absorb refund boundary decisions")
	}
}

// TestX20_MeaningfulOrder verifies that meaningful ordering differences survive normalization.
func TestX20_MeaningfulOrder(t *testing.T) {
	rules := []*NormalizationRule{
		{
			SchemaVersion:           "2.0",
			RuleID:                  "norm_repr",
			Path:                    "/items",
			ValueType:               "array",
			Variability:             "REPRESENTATION",
			Relation:                "CANONICAL_REPRESENTATION",
			RelationGroup:           "items_repr",
			LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
			BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
			DecisionRef:             "sha256:" + strings.Repeat("c", 64),
			NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
		},
	}

	// Meaningful order: source returns items in purchase-date order (oldest first);
	// target returns items in random order (different ordering = business logic failure)
	mutant := NormalizationMutant{
		MutantID:    "meaningful-order-changed",
		Category:    "MEANINGFUL_ORDER",
		Description: "Source returns order history sorted by purchase_date ASC; target returns unsorted/random order",
		SourceValue: TaggedValue{Type: ValueTypeString, Value: `[{"id":1,"date":"2026-01-01"},{"id":2,"date":"2026-06-01"}]`},
		TargetValue: TaggedValue{Type: ValueTypeString, Value: `[{"id":2,"date":"2026-06-01"},{"id":1,"date":"2026-01-01"}]`},
		Context: map[string]string{
			"path":             "$.data.orders",
			"ordering_contract": "purchase_date_asc",
		},
	}

	protector := NewNormalizerMutationProtector(rules)
	result := protector.CheckMutant(mutant)

	// Meaningful ordering must NOT be absorbed by REPRESENTATION normalization
	if result.Verdict != "MISMATCH" {
		t.Fatalf("X20: meaningful order change must survive normalization as MISMATCH, got %s; reason: %s",
			result.Verdict, result.ViolationReason)
	}
	if result.WasNormalized {
		t.Fatalf("X20: REPRESENTATION normalization must NOT absorb meaningful ordering")
	}
}

// TestX20_BijectionIntegrityCheck verifies bijective identity mapping enforcement.
func TestX20_BijectionIntegrityCheck(t *testing.T) {
	checker := &BijectionIntegrityChecker{}

	t.Run("ValidBijection_NoErrors", func(t *testing.T) {
		conn := "conn-main"
		idMap := &IdentityMapV2{
			SchemaVersion:      "2.1",
			ProjectID:          "proj-test",
			CaseID:             "case-test",
			WorkflowInstanceID: "workflow-1",
			PolicyRef:          "sha256:" + strings.Repeat("a", 64),
			Bindings: []IdentityBinding{
				{
					BindingID:     "bind-1",
					RelationGroup: "user_identity",
					ActorID:       "actor-default",
					ConnectionID:  &conn,
					LogicalKey:    "user_user_alice",
					SourceValue:   TypedValueContainer{ValueType: "string", Value: "alice"},
					TargetValue:   TypedValueContainer{ValueType: "string", Value: "alice-new"},
					SourceLocalID: "src-user-alice",
					TargetLocalID: "tgt-user-alice",
					EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
				},
				{
					BindingID:     "bind-2",
					RelationGroup: "user_identity",
					ActorID:       "actor-default",
					ConnectionID:  &conn,
					LogicalKey:    "user_user_bob",
					SourceValue:   TypedValueContainer{ValueType: "string", Value: "bob"},
					TargetValue:   TypedValueContainer{ValueType: "string", Value: "bob-new"},
					SourceLocalID: "src-user-bob",
					TargetLocalID: "tgt-user-bob",
					EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
				},
			},
			RelationshipAssertionRefs: []string{"sha256:" + strings.Repeat("a", 64)},
		}
		errs := checker.CheckBijection(idMap)
		if len(errs) != 0 {
			t.Fatalf("expected no bijection errors for valid bijection, got: %v", errs)
		}
	})

	t.Run("BrokenBijection_TwoSourcesMappedToSameTarget", func(t *testing.T) {
		conn := "conn-main"
		idMap := &IdentityMapV2{
			SchemaVersion:      "2.1",
			ProjectID:          "proj-test",
			CaseID:             "case-test",
			WorkflowInstanceID: "workflow-1",
			PolicyRef:          "sha256:" + strings.Repeat("a", 64),
			Bindings: []IdentityBinding{
				{
					BindingID:     "bind-1",
					RelationGroup: "user_identity",
					ActorID:       "actor-default",
					ConnectionID:  &conn,
					LogicalKey:    "user_alice",
					SourceValue:   TypedValueContainer{ValueType: "string", Value: "alice"},
					TargetValue:   TypedValueContainer{ValueType: "string", Value: "merged-user"},
					SourceLocalID: "src-alice",
					TargetLocalID: "tgt-merged",
					EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
				},
				{
					BindingID:     "bind-2",
					RelationGroup: "user_identity",
					ActorID:       "actor-default",
					ConnectionID:  &conn,
					LogicalKey:    "user_bob",
					SourceValue:   TypedValueContainer{ValueType: "string", Value: "bob"},
					TargetValue:   TypedValueContainer{ValueType: "string", Value: "merged-user"}, // SAME TARGET as bind-1!
					SourceLocalID: "src-bob",
					TargetLocalID: "tgt-merged", // same target! bijection violation
					EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
				},
			},
			RelationshipAssertionRefs: []string{"sha256:" + strings.Repeat("a", 64)},
		}
		errs := checker.CheckBijection(idMap)
		if len(errs) == 0 {
			t.Fatalf("X20: expected bijection violation for two sources mapping to same target")
		}
	})

	t.Run("BrokenBijection_SameSourceMappedToTwoTargets", func(t *testing.T) {
		conn := "conn-main"
		idMap := &IdentityMapV2{
			SchemaVersion:      "2.1",
			ProjectID:          "proj-test",
			CaseID:             "case-test",
			WorkflowInstanceID: "workflow-1",
			PolicyRef:          "sha256:" + strings.Repeat("a", 64),
			Bindings: []IdentityBinding{
				{
					BindingID:     "bind-1",
					RelationGroup: "user_identity",
					ActorID:       "actor-default",
					ConnectionID:  &conn,
					LogicalKey:    "user_alice_1",
					SourceValue:   TypedValueContainer{ValueType: "string", Value: "alice"},
					TargetValue:   TypedValueContainer{ValueType: "string", Value: "alice-target-1"},
					SourceLocalID: "src-alice",
					TargetLocalID: "tgt-alice-1",
					EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
				},
				{
					BindingID:     "bind-2",
					RelationGroup: "user_identity",
					ActorID:       "actor-default",
					ConnectionID:  &conn,
					LogicalKey:    "user_alice_2",
					SourceValue:   TypedValueContainer{ValueType: "string", Value: "alice"},
					TargetValue:   TypedValueContainer{ValueType: "string", Value: "alice-target-2"},
					SourceLocalID: "src-alice", // SAME SOURCE!
					TargetLocalID: "tgt-alice-2",
					EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
				},
			},
			RelationshipAssertionRefs: []string{"sha256:" + strings.Repeat("a", 64)},
		}
		errs := checker.CheckBijection(idMap)
		if len(errs) == 0 {
			t.Fatalf("X20: expected bijection violation for same source mapped to two targets")
		}
	})
}

// TestX20_RunNormalizationMutationGuard verifies the full X20 mutation guard integration.
func TestX20_RunNormalizationMutationGuard(t *testing.T) {
	rules := []*NormalizationRule{
		{
			SchemaVersion:           "2.0",
			RuleID:                  "norm_uuid",
			Path:                    "/id",
			ValueType:               "string",
			Variability:             "GENERATED_ID",
			Relation:                "BIJECTIVE_ID",
			RelationGroup:           "user_identity",
			LifecycleContractRef:    "sha256:" + strings.Repeat("a", 64),
			BusinessSignificanceRef: "sha256:" + strings.Repeat("b", 64),
			DecisionRef:             "sha256:" + strings.Repeat("c", 64),
			NegativeMutationRefs:    []string{"sha256:" + strings.Repeat("d", 64)},
		},
	}

	conn := "conn-main"
	idMap := &IdentityMapV2{
		SchemaVersion:      "2.1",
		ProjectID:          "proj-test",
		CaseID:             "case-x20",
		WorkflowInstanceID: "workflow-1",
		PolicyRef:          "sha256:" + strings.Repeat("a", 64),
		Bindings: []IdentityBinding{
			{
				BindingID:     "bind-1",
				RelationGroup: "user_identity",
				ActorID:       "actor-default",
				ConnectionID:  &conn,
				LogicalKey:    "user_alice",
				SourceValue:   TypedValueContainer{ValueType: "string", Value: "src-alice-id-12345"},
				TargetValue:   TypedValueContainer{ValueType: "string", Value: "tgt-alice-uuid-6789"},
				SourceLocalID: "src-user-alice",
				TargetLocalID: "tgt-user-alice",
				EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
			},
		},
		RelationshipAssertionRefs: []string{"sha256:" + strings.Repeat("a", 64)},
	}

	mutants := []NormalizationMutant{
		{
			MutantID:    "broken-fk",
			Category:    "BROKEN_FOREIGN_KEY",
			SourceValue: TaggedValue{Type: ValueTypeInteger, Value: "42"},
			TargetValue: TaggedValue{Type: ValueTypeInteger, Value: "99"},
		},
		{
			MutantID:    "wrong-principal",
			Category:    "WRONG_PRINCIPAL",
			SourceValue: TaggedValue{Type: ValueTypeString, Value: "alice"},
			TargetValue: TaggedValue{Type: ValueTypeString, Value: "bob"},
		},
		{
			MutantID: "expired-token",
			Category: "EXPIRED_TOKEN",
			SourceValue: TaggedValue{Type: ValueTypeString, Value: "ACCEPTED"},
			TargetValue: TaggedValue{Type: ValueTypeString, Value: "REJECTED"},
			Context:  map[string]string{"is_expiry_boundary": "true"},
		},
		{
			MutantID:    "refund-boundary",
			Category:    "REFUND_BOUNDARY",
			SourceValue: TaggedValue{Type: ValueTypeString, Value: "REFUND_REJECTED"},
			TargetValue: TaggedValue{Type: ValueTypeString, Value: "REFUND_ACCEPTED"},
		},
		{
			MutantID:    "meaningful-order",
			Category:    "MEANINGFUL_ORDER",
			SourceValue: TaggedValue{Type: ValueTypeString, Value: `[1,2,3]`},
			TargetValue: TaggedValue{Type: ValueTypeString, Value: `[3,1,2]`},
		},
	}

	report := RunNormalizationMutationGuard(rules, mutants, idMap)

	if len(report.PolicyViolations) != 0 {
		t.Fatalf("unexpected policy violations: %v", report.PolicyViolations)
	}
	if len(report.ProtectionViolations) != 0 {
		t.Fatalf("X20: normalization absorbed required mutants: %v", report.ProtectionViolations)
	}
	if len(report.BijectionViolations) != 0 {
		t.Fatalf("X20: bijection violations: %v", report.BijectionViolations)
	}
	if report.OverallVerdict != "PASS" {
		t.Fatalf("X20: expected PASS, got %s", report.OverallVerdict)
	}

	// All required mutants must be MISMATCH
	for _, r := range report.MutantResults {
		switch r.Category {
		case "BROKEN_FOREIGN_KEY", "WRONG_PRINCIPAL", "EXPIRED_TOKEN", "REFUND_BOUNDARY", "MEANINGFUL_ORDER":
			if r.Verdict != "MISMATCH" {
				t.Errorf("X20: mutant %q (%s) must be MISMATCH, got %s", r.MutantID, r.Category, r.Verdict)
			}
		}
	}
}
