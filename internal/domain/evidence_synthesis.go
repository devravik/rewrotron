package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SynthesisStage represents the progression of a learned technique.
type SynthesisStage string

const (
	StageObservation         SynthesisStage = "OBSERVATION"
	StageHypothesis          SynthesisStage = "HYPOTHESIS"
	StageRepeatedObservation SynthesisStage = "REPEATED_OBSERVATION"
	StageValidatedPattern    SynthesisStage = "VALIDATED_PATTERN"
	StageGeneralizedTechnique SynthesisStage = "GENERALIZED_TECHNIQUE"
	StageSkillCandidate      SynthesisStage = "SKILL_CANDIDATE"
	StagePromotedStable      SynthesisStage = "PROMOTED_STABLE"
)

// FixtureContext represents an execution fixture and its derivation provenance.
type FixtureContext struct {
	FixtureID           string `json:"fixture_id"`
	ApplicationID       string `json:"application_id"`
	Author              string `json:"author"`
	OracleDerivation    string `json:"oracle_derivation"` // derivation method or oracle group
	TemplateDerivation  string `json:"template_derivation"`
	CorrelationGroup    string `json:"correlation_group"` // non-independent fixtures share correlation group
	Passed              bool   `json:"passed"`
	NewlyIntroducedUnknowns int `json:"newly_introduced_unknowns"`
}

// NegativeMutantEvaluation records results against designated negative semantic mutants.
type NegativeMutantEvaluation struct {
	MutantID          string `json:"mutant_id"`
	MutatedDimension  string `json:"mutated_dimension"` // e.g. "owner", "write_effect", "token", "reset", "capture"
	DetectedByOracle  bool   `json:"detected_by_oracle"`
	Verdict           string `json:"verdict"` // MISMATCH or UNKNOWN
}

// HoldoutEvaluation records evaluation results on independent holdout suites.
type HoldoutEvaluation struct {
	HoldoutSuiteID    string `json:"holdout_suite_id"`
	TotalCases        int    `json:"total_cases"`
	PassedCases       int    `json:"passed_cases"`
	RegressionDetected bool  `json:"regression_detected"`
}

// GlobalPromotionCandidate defines a candidate capability revision being evaluated for promotion.
type GlobalPromotionCandidate struct {
	CandidateID       string                     `json:"candidate_id"`
	SkillID           string                     `json:"skill_id"`
	Version           string                     `json:"version"`
	PackageDigest     string                     `json:"package_digest"` // sha256 hex
	SynthesisStage    SynthesisStage             `json:"synthesis_stage"`
	Applicability     map[string]string          `json:"applicability"`
	Fixtures          []FixtureContext           `json:"fixtures"`
	NegativeMutants   []NegativeMutantEvaluation `json:"negative_mutants"`
	Holdout           HoldoutEvaluation          `json:"holdout"`
	TotalAttempts     int                        `json:"total_attempts"`
	FailedAttempts    int                        `json:"failed_attempts"`
	HumanReviewedBy   string                     `json:"human_reviewed_by,omitempty"`
	HumanApprovalToken string                    `json:"human_approval_token,omitempty"`
}

// PromotionVerdict represents the evaluation result.
type PromotionVerdict struct {
	Eligible           bool     `json:"eligible"`
	Decision           string   `json:"decision"` // "PROMOTED", "REJECTED", "BLOCKED"
	RefusalReasons     []string `json:"refusal_reasons,omitempty"`
	IndependentContexts int     `json:"independent_contexts"`
	MutantKillRate     float64  `json:"mutant_kill_rate"`
	ZeroHoldoutRegress bool     `json:"zero_holdout_regress"`
	ZeroNewUnknowns    bool     `json:"zero_new_unknowns"`
	HumanAuthorized    bool     `json:"human_authorized"`
	ReceiptHash        string   `json:"receipt_hash,omitempty"`
}

// EvaluatePromotion evaluates a candidate against the mandatory LEARNING-SYSTEM and SKILL-EVOLUTION criteria:
// 1. Minimum 3 independent fixture/application contexts (independent author/oracle/template/correlation group).
// 2. 100% detection of designated negative semantic mutants.
// 3. Zero holdout regressions.
// 4. Zero newly introduced UNKNOWNs.
// 5. Explicit human promotion review (agents cannot promote to STABLE).
// 6. Denominators cannot omit failed attempts.
func EvaluatePromotion(cand *GlobalPromotionCandidate) (*PromotionVerdict, error) {
	if cand == nil {
		return nil, fmt.Errorf("candidate cannot be nil")
	}

	var refusalReasons []string

	// 1. Check independence across fixtures
	// Distinct correlation groups and oracle derivations
	uniqueCorrelationGroups := make(map[string]bool)
	uniqueOracles := make(map[string]bool)
	passedIndependentCount := 0
	totalNewUnknowns := 0

	for _, fix := range cand.Fixtures {
		totalNewUnknowns += fix.NewlyIntroducedUnknowns
		if !fix.Passed {
			continue
		}
		// Group key combining application, correlation group, and oracle derivation
		grpKey := fmt.Sprintf("%s:%s", fix.CorrelationGroup, fix.OracleDerivation)
		if !uniqueCorrelationGroups[grpKey] {
			uniqueCorrelationGroups[grpKey] = true
			uniqueOracles[fix.OracleDerivation] = true
			passedIndependentCount++
		}
	}

	if passedIndependentCount < 3 {
		refusalReasons = append(refusalReasons, fmt.Sprintf(
			"INSUFFICIENT_INDEPENDENT_CONTEXTS: required at least 3 independent fixture/application contexts, found %d (correlation groups: %d, oracles: %d)",
			passedIndependentCount, len(uniqueCorrelationGroups), len(uniqueOracles),
		))
	}

	// 2. Negative semantic mutant kill rate (must be 100%)
	totalMutants := len(cand.NegativeMutants)
	detectedMutants := 0
	for _, m := range cand.NegativeMutants {
		if m.DetectedByOracle {
			detectedMutants++
		}
	}

	mutantKillRate := 0.0
	if totalMutants > 0 {
		mutantKillRate = float64(detectedMutants) / float64(totalMutants)
	}

	if totalMutants == 0 || detectedMutants < totalMutants {
		refusalReasons = append(refusalReasons, fmt.Sprintf(
			"MUTANT_COVERAGE_FAILURE: required 100%% negative semantic mutant detection, detected %d/%d (%.1f%%)",
			detectedMutants, totalMutants, mutantKillRate*100,
		))
	}

	// 3. Holdout evaluation (zero regression allowed)
	if cand.Holdout.RegressionDetected || (cand.Holdout.TotalCases > 0 && cand.Holdout.PassedCases < cand.Holdout.TotalCases) {
		refusalReasons = append(refusalReasons, fmt.Sprintf(
			"HOLDOUT_REGRESSION: holdout suite %s detected regressions (%d/%d passed)",
			cand.Holdout.HoldoutSuiteID, cand.Holdout.PassedCases, cand.Holdout.TotalCases,
		))
	}

	// 4. Zero newly introduced UNKNOWNs
	if totalNewUnknowns > 0 {
		refusalReasons = append(refusalReasons, fmt.Sprintf(
			"NEW_UNKNOWNS_INTRODUCED: candidate introduced %d new UNKNOWN verdicts across evaluation fixtures",
			totalNewUnknowns,
		))
	}

	// 5. Denominators cannot hide failures (consistency check)
	if cand.TotalAttempts < cand.FailedAttempts || cand.TotalAttempts < len(cand.Fixtures) {
		refusalReasons = append(refusalReasons, fmt.Sprintf(
			"INVALID_DENOMINATOR: total_attempts (%d) cannot be less than failed_attempts (%d) or evaluated fixtures (%d)",
			cand.TotalAttempts, cand.FailedAttempts, len(cand.Fixtures),
		))
	}

	// 6. Human authorization required for STABLE promotion
	humanAuth := strings.TrimSpace(cand.HumanReviewedBy) != "" && strings.TrimSpace(cand.HumanApprovalToken) != ""
	if !humanAuth {
		refusalReasons = append(refusalReasons, "MISSING_HUMAN_APPROVAL: promotion to STABLE requires explicit human reviewer and approval token")
	}

	eligible := len(refusalReasons) == 0
	decision := "PROMOTED"
	if !eligible {
		decision = "REJECTED"
	}

	verdict := &PromotionVerdict{
		Eligible:           eligible,
		Decision:           decision,
		RefusalReasons:     refusalReasons,
		IndependentContexts: passedIndependentCount,
		MutantKillRate:     mutantKillRate,
		ZeroHoldoutRegress: !cand.Holdout.RegressionDetected && cand.Holdout.PassedCases == cand.Holdout.TotalCases,
		ZeroNewUnknowns:    totalNewUnknowns == 0,
		HumanAuthorized:    humanAuth,
	}

	if eligible {
		// Emit reproducible receipt hash
		recJSON, _ := json.Marshal(map[string]any{
			"candidate_id": cand.CandidateID,
			"package_digest": cand.PackageDigest,
			"independent_contexts": passedIndependentCount,
			"human_reviewed_by": cand.HumanReviewedBy,
			"evaluated_at": time.Now().UTC().Format(time.RFC3339),
		})
		h := sha256.Sum256(recJSON)
		verdict.ReceiptHash = hex.EncodeToString(h[:])
		cand.SynthesisStage = StagePromotedStable
	}

	return verdict, nil
}

// GlobalKnowledgePattern represents an approved, sanitized global pattern ready for publication.
type GlobalKnowledgePattern struct {
	PatternID          string            `json:"pattern_id"`
	Version            string            `json:"version"`
	PackageDigest      string            `json:"package_digest"`
	CapabilityTag      string            `json:"capability_tag"`
	SanitizedExamples  []string          `json:"sanitized_examples"`
	IndependentEvidenceCount int         `json:"independent_evidence_count"`
	ReceiptHash        string            `json:"receipt_hash"`
	ApprovedBy         string            `json:"approved_by"`
	CreatedAt          string            `json:"created_at"`
}
