package domain

import (
	"strings"
	"testing"
)

func TestEvidenceSynthesis_PromotionEvaluation(t *testing.T) {
	// Base valid candidate meeting all criteria
	validCandidate := func() *GlobalPromotionCandidate {
		return &GlobalPromotionCandidate{
			CandidateID:   "cand-001",
			SkillID:       "skill-session-bridge",
			Version:       "1.0.0",
			PackageDigest: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			SynthesisStage: StageSkillCandidate,
			Applicability: map[string]string{"framework": "ci3", "target": "laravel"},
			Fixtures: []FixtureContext{
				{
					FixtureID:        "fix-app1",
					ApplicationID:    "app-ecom",
					Author:           "author-a",
					OracleDerivation: "oracle-diff-1",
					CorrelationGroup: "corp-a",
					Passed:           true,
				},
				{
					FixtureID:        "fix-app2",
					ApplicationID:    "app-crm",
					Author:           "author-b",
					OracleDerivation: "oracle-diff-2",
					CorrelationGroup: "corp-b",
					Passed:           true,
				},
				{
					FixtureID:        "fix-app3",
					ApplicationID:    "app-portal",
					Author:           "author-c",
					OracleDerivation: "oracle-diff-3",
					CorrelationGroup: "corp-c",
					Passed:           true,
				},
			},
			NegativeMutants: []NegativeMutantEvaluation{
				{MutantID: "mut-owner", MutatedDimension: "owner", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "mut-write", MutatedDimension: "write_effect", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "mut-token", MutatedDimension: "token", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "mut-reset", MutatedDimension: "reset", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "mut-capture", MutatedDimension: "capture", DetectedByOracle: true, Verdict: "MISMATCH"},
			},
			Holdout: HoldoutEvaluation{
				HoldoutSuiteID:    "holdout-independent-01",
				TotalCases:        15,
				PassedCases:       15,
				RegressionDetected: false,
			},
			TotalAttempts:      10,
			FailedAttempts:     2,
			HumanReviewedBy:    "human-operator-lead",
			HumanApprovalToken: "tok-approval-gate-review-99",
		}
	}

	t.Run("ValidCandidatePromotes", func(t *testing.T) {
		cand := validCandidate()
		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !verdict.Eligible || verdict.Decision != "PROMOTED" {
			t.Fatalf("expected candidate to be promoted, got decision: %s, reasons: %v", verdict.Decision, verdict.RefusalReasons)
		}
		if verdict.ReceiptHash == "" {
			t.Fatalf("expected promotion receipt hash to be emitted")
		}
		if cand.SynthesisStage != StagePromotedStable {
			t.Fatalf("expected synthesis stage to advance to PROMOTED_STABLE")
		}
	})

	t.Run("X37_CorrelatedFixturesSharedWrongOracleRejected", func(t *testing.T) {
		cand := validCandidate()
		// 3 fixtures but all sharing the same correlation group and same wrong oracle derivation
		cand.Fixtures = []FixtureContext{
			{FixtureID: "fix-1", ApplicationID: "app-1", Author: "author-a", OracleDerivation: "oracle-shared-generator", CorrelationGroup: "grp-shared", Passed: true},
			{FixtureID: "fix-2", ApplicationID: "app-1", Author: "author-a", OracleDerivation: "oracle-shared-generator", CorrelationGroup: "grp-shared", Passed: true},
			{FixtureID: "fix-3", ApplicationID: "app-1", Author: "author-a", OracleDerivation: "oracle-shared-generator", CorrelationGroup: "grp-shared", Passed: true},
		}

		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verdict.Eligible || verdict.Decision != "REJECTED" {
			t.Fatalf("expected correlated fixtures with shared oracle to be rejected")
		}
		foundReason := false
		for _, r := range verdict.RefusalReasons {
			if strings.Contains(r, "INSUFFICIENT_INDEPENDENT_CONTEXTS") {
				foundReason = true
				break
			}
		}
		if !foundReason {
			t.Fatalf("expected INSUFFICIENT_INDEPENDENT_CONTEXTS in refusal reasons: %v", verdict.RefusalReasons)
		}
	})

	t.Run("NegativeMutantFailureBlocksPromotion", func(t *testing.T) {
		cand := validCandidate()
		// Oracle missed detecting the write_effect mutant
		cand.NegativeMutants[1].DetectedByOracle = false

		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verdict.Eligible || verdict.Decision != "REJECTED" {
			t.Fatalf("expected mutant failure to block promotion")
		}
		foundReason := false
		for _, r := range verdict.RefusalReasons {
			if strings.Contains(r, "MUTANT_COVERAGE_FAILURE") {
				foundReason = true
				break
			}
		}
		if !foundReason {
			t.Fatalf("expected MUTANT_COVERAGE_FAILURE in refusal reasons: %v", verdict.RefusalReasons)
		}
	})

	t.Run("HoldoutRegressionBlocksPromotion", func(t *testing.T) {
		cand := validCandidate()
		cand.Holdout.RegressionDetected = true
		cand.Holdout.PassedCases = 14 // 14/15

		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verdict.Eligible || verdict.Decision != "REJECTED" {
			t.Fatalf("expected holdout regression to block promotion")
		}
		foundReason := false
		for _, r := range verdict.RefusalReasons {
			if strings.Contains(r, "HOLDOUT_REGRESSION") {
				foundReason = true
				break
			}
		}
		if !foundReason {
			t.Fatalf("expected HOLDOUT_REGRESSION in refusal reasons: %v", verdict.RefusalReasons)
		}
	})

	t.Run("NewUnknownsBlocksPromotion", func(t *testing.T) {
		cand := validCandidate()
		cand.Fixtures[0].NewlyIntroducedUnknowns = 1

		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verdict.Eligible || verdict.Decision != "REJECTED" {
			t.Fatalf("expected new unknowns to block promotion")
		}
		foundReason := false
		for _, r := range verdict.RefusalReasons {
			if strings.Contains(r, "NEW_UNKNOWNS_INTRODUCED") {
				foundReason = true
				break
			}
		}
		if !foundReason {
			t.Fatalf("expected NEW_UNKNOWNS_INTRODUCED in refusal reasons: %v", verdict.RefusalReasons)
		}
	})

	t.Run("AgentCannotPromoteWithoutHumanApproval", func(t *testing.T) {
		cand := validCandidate()
		cand.HumanReviewedBy = ""
		cand.HumanApprovalToken = ""

		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verdict.Eligible || verdict.Decision != "REJECTED" {
			t.Fatalf("expected missing human approval to block promotion")
		}
		foundReason := false
		for _, r := range verdict.RefusalReasons {
			if strings.Contains(r, "MISSING_HUMAN_APPROVAL") {
				foundReason = true
				break
			}
		}
		if !foundReason {
			t.Fatalf("expected MISSING_HUMAN_APPROVAL in refusal reasons: %v", verdict.RefusalReasons)
		}
	})
}

