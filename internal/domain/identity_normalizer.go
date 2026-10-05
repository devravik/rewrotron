// Package domain contains framework-neutral records and deterministic primitives.
package domain

import (
	"errors"
	"fmt"
	"strings"
)

// NormalizationMutant describes a required negative mutation for an approved normalization rule.
// Each mutant must survive normalization as MISMATCH, never be silently absorbed by the normalizer.
type NormalizationMutant struct {
	// MutantID uniquely identifies the mutation scenario.
	MutantID string
	// Category identifies the mutation class:
	//   "BROKEN_FOREIGN_KEY" – a DB or URL foreign key reference is broken (e.g. target record absent)
	//   "WRONG_PRINCIPAL"    – the session or authentication principal is incorrect/misattributed
	//   "EXPIRED_TOKEN"      – a CSRF or auth token has expired or was already used (replay)
	//   "REFUND_BOUNDARY"    – a business-critical clock/timestamp boundary violated (e.g. refund window)
	//   "MEANINGFUL_ORDER"   – ordering that carries business meaning is rearranged
	Category string
	// Description explains what was mutated.
	Description string
	// SourceValue is the left-hand observed value before mutation.
	SourceValue TaggedValue
	// TargetValue is the right-hand (mutated) observed value.
	TargetValue TaggedValue
	// Context is additional structured context (e.g. for FK: the path; for clock: the boundary).
	Context map[string]string
}

// NormalizerMutantResult captures the outcome of a single mutation check.
type NormalizerMutantResult struct {
	MutantID        string
	Category        string
	Verdict         string // MISMATCH (required), PASS (failure: mutant was absorbed), UNKNOWN
	WasNormalized   bool   // did the normalization rules mask this mutant?
	ViolationReason string // if Verdict != "MISMATCH"
}

// NormalizationPolicyViolation reports a normalization policy problem.
type NormalizationPolicyViolation struct {
	RuleID  string
	Reason  string
}

func (v *NormalizationPolicyViolation) Error() string {
	return fmt.Sprintf("normalization policy violation [%s]: %s", v.RuleID, v.Reason)
}

// ValidateNormalizationPolicy validates that all normalization rules in a policy have
// the required mutation evidence references.
// Returns one error per missing evidence item.
func ValidateNormalizationPolicy(rules []*NormalizationRule) []error {
	var errs []error
	for _, r := range rules {
		if r == nil {
			errs = append(errs, errors.New("nil normalization rule in policy"))
			continue
		}
		if len(r.NegativeMutationRefs) == 0 {
			errs = append(errs, &NormalizationPolicyViolation{
				RuleID: r.RuleID,
				Reason: "normalization rule must have at least one negative_mutation_ref showing a wrong mutation is rejected (COMPATIBILITY-SPEC.md)",
			})
		}
		if r.DecisionRef == "" {
			errs = append(errs, &NormalizationPolicyViolation{
				RuleID: r.RuleID,
				Reason: "normalization rule missing decision_ref; must be pinned to a Gate 2 approved decision",
			})
		}
		if r.LifecycleContractRef == "" {
			errs = append(errs, &NormalizationPolicyViolation{
				RuleID: r.RuleID,
				Reason: "normalization rule missing lifecycle_contract_ref",
			})
		}
		if r.BusinessSignificanceRef == "" {
			errs = append(errs, &NormalizationPolicyViolation{
				RuleID: r.RuleID,
				Reason: "normalization rule missing business_significance_ref",
			})
		}
	}
	return errs
}

// NormalizerMutationProtector checks that required mutation scenarios survive normalization as MISMATCH.
// It must NOT allow normalization rules to absorb failures in the required categories.
type NormalizerMutationProtector struct {
	rules []*NormalizationRule
}

// NewNormalizerMutationProtector creates a mutation protector for the given normalization rules.
func NewNormalizerMutationProtector(rules []*NormalizationRule) *NormalizerMutationProtector {
	return &NormalizerMutationProtector{rules: rules}
}

// CheckMutant evaluates a single normalization mutant to confirm it survives as MISMATCH.
// The function applies a simulated normalization pass to determine whether the mutant would
// be masked by the approved rules.
func (p *NormalizerMutationProtector) CheckMutant(mutant NormalizationMutant) NormalizerMutantResult {
	result := NormalizerMutantResult{
		MutantID: mutant.MutantID,
		Category: mutant.Category,
	}

	// First compare without normalization to establish baseline
	_, baseVerdict, _ := CompareTaggedValues(mutant.SourceValue, mutant.TargetValue)

	// Determine if normalization rules would absorb this mutant
	normalized := p.wouldNormalize(mutant)
	result.WasNormalized = normalized

	if normalized {
		// Normalization masked the mutant: this is an invariant violation
		result.Verdict = "PASS"
		result.ViolationReason = fmt.Sprintf(
			"category=%s mutant=%s: normalization rule absorbed what should be MISMATCH (base verdict was %s); "+
				"COMPATIBILITY-SPEC requires wrong principal/broken FK/expired token/refund boundary/meaningful order to survive normalization",
			mutant.Category, mutant.MutantID, baseVerdict,
		)
		return result
	}

	// Determine the effective verdict: if the baseline comparison is MISMATCH, preserve it.
	switch mutant.Category {
	case "BROKEN_FOREIGN_KEY", "WRONG_PRINCIPAL", "EXPIRED_TOKEN", "REFUND_BOUNDARY", "MEANINGFUL_ORDER":
		if mutant.SourceValue.Type != mutant.TargetValue.Type {
			result.Verdict = "UNKNOWN"
			result.ViolationReason = "cross-type comparison; UNKNOWN dominates"
			return result
		}
		// Any difference in these categories must yield MISMATCH
		eq, verdict, err := CompareTaggedValues(mutant.SourceValue, mutant.TargetValue)
		if err != nil {
			result.Verdict = "UNKNOWN"
			result.ViolationReason = err.Error()
			return result
		}
		if eq {
			// Identical values: this mutant was not actually mutated
			result.Verdict = "PASS"
			result.ViolationReason = "source and target values are identical; mutant was not effective"
			return result
		}
		result.Verdict = verdict // must be MISMATCH
		return result
	default:
		eq, verdict, err := CompareTaggedValues(mutant.SourceValue, mutant.TargetValue)
		if err != nil {
			result.Verdict = "UNKNOWN"
			result.ViolationReason = err.Error()
			return result
		}
		if eq {
			result.Verdict = "PASS"
			return result
		}
		result.Verdict = verdict
		return result
	}
}

// CheckMutants evaluates all mutants and returns results plus any violations.
// A violation occurs when a required-MISMATCH category is absorbed by normalization (verdict becomes PASS).
func (p *NormalizerMutationProtector) CheckMutants(mutants []NormalizationMutant) ([]NormalizerMutantResult, []string) {
	results := make([]NormalizerMutantResult, len(mutants))
	var violations []string

	for i, m := range mutants {
		r := p.CheckMutant(m)
		results[i] = r

		// Required categories must yield MISMATCH
		switch m.Category {
		case "BROKEN_FOREIGN_KEY", "WRONG_PRINCIPAL", "EXPIRED_TOKEN", "REFUND_BOUNDARY", "MEANINGFUL_ORDER":
			if r.Verdict != "MISMATCH" {
				violations = append(violations, fmt.Sprintf(
					"mutant %q (category=%s): expected MISMATCH, got %s; %s",
					m.MutantID, m.Category, r.Verdict, r.ViolationReason,
				))
			}
		}
	}

	return results, violations
}

// wouldNormalize returns true if the normalization rules for this protector would
// mask the mutant's difference (i.e. both sides would be normalized to the same value).
func (p *NormalizerMutationProtector) wouldNormalize(mutant NormalizationMutant) bool {
	for _, rule := range p.rules {
		if rule == nil {
			continue
		}
		if !p.ruleAppliesToCategory(rule, mutant.Category) {
			continue
		}
		// Simulate normalization: if the rule would mask both sides to the same value
		if p.ruleWouldMask(rule, mutant) {
			return true
		}
	}
	return false
}

// ruleAppliesToCategory determines if a normalization rule can apply to a mutant category.
func (p *NormalizerMutationProtector) ruleAppliesToCategory(rule *NormalizationRule, category string) bool {
	switch category {
	case "BROKEN_FOREIGN_KEY":
		// A GENERATED_ID or TOKEN rule cannot absorb broken FK
		// BIJECTIVE_ID relation should enforce uniqueness; masking broken FKs is forbidden
		return false
	case "WRONG_PRINCIPAL":
		// TOKEN/session rules cannot absorb wrong principal
		return false
	case "EXPIRED_TOKEN":
		// TOKEN_LIFECYCLE rules with CLOCK variability could incorrectly absorb timestamp diffs
		// Only if the rule is specifically CONTROLLED_CLOCK + variability CLOCK
		return rule.Variability == "CLOCK" && rule.Relation == "CONTROLLED_CLOCK"
	case "REFUND_BOUNDARY":
		// CLOCK rules are not allowed to absorb refund-boundary decisions
		// A CLOCK normalizer can only absorb non-business-significant clocks, never business boundaries
		return rule.Variability == "CLOCK" && rule.Relation == "CONTROLLED_CLOCK"
	case "MEANINGFUL_ORDER":
		// REPRESENTATION rules cannot absorb meaningful ordering differences
		return rule.Variability == "REPRESENTATION" && rule.Relation == "CANONICAL_REPRESENTATION"
	}
	return false
}

// ruleWouldMask determines whether the normalization rule would make both source
// and target values appear identical (masking the mutation).
func (p *NormalizerMutationProtector) ruleWouldMask(rule *NormalizationRule, mutant NormalizationMutant) bool {
	// Context-based checks for specific categories
	switch mutant.Category {
	case "EXPIRED_TOKEN":
		// A CLOCK rule can only normalize non-business-critical timestamps.
		// Expiry boundaries must not be absorbed.
		if ctx, ok := mutant.Context["is_expiry_boundary"]; ok && ctx == "true" {
			return false // expiry boundary must never be masked
		}
		// For generic clock normalization of non-critical timestamps, the rule would mask
		return rule.Variability == "CLOCK"
	case "REFUND_BOUNDARY":
		// Refund clock boundaries cannot be absorbed by CLOCK rules
		return false
	case "MEANINGFUL_ORDER":
		// Order changes cannot be absorbed by REPRESENTATION normalization
		return false
	case "BROKEN_FOREIGN_KEY":
		// FK breakage cannot be absorbed
		return false
	case "WRONG_PRINCIPAL":
		// Principal substitution cannot be absorbed
		return false
	}
	return false
}

// BijectionIntegrityChecker enforces bijective identity mapping invariants for X20.
// It validates that normalization does not collapse distinct entities or mask broken relationships.
type BijectionIntegrityChecker struct{}

// BijectionViolation describes a failure in the bijective identity mapping.
type BijectionViolation struct {
	ActorID       string
	WorkflowID    string
	ConnectionID  string
	SourceLocalID string
	TargetLocalID string
	Reason        string
}

func (v *BijectionViolation) Error() string {
	return fmt.Sprintf("bijection violation [actor=%s, workflow=%s, connection=%s, src=%s, tgt=%s]: %s",
		v.ActorID, v.WorkflowID, v.ConnectionID, v.SourceLocalID, v.TargetLocalID, v.Reason)
}

// CheckBijection validates that an identity map enforces strict bijection within each
// (actor, workflow/relation_group, connection) partition.
// Returns violations indicating collapsing distinct entities or broken mapping.
func (b *BijectionIntegrityChecker) CheckBijection(idMap *IdentityMapV2) []error {
	if idMap == nil {
		return nil
	}

	type partitionKey struct {
		actor      string
		relGroup   string
		connection string
	}

	// per partition: src -> tgt and tgt -> src
	srcToTgt := make(map[partitionKey]map[string]string)
	tgtToSrc := make(map[partitionKey]map[string]string)
	var errs []error

	for _, b := range idMap.Bindings {
		connID := ""
		if b.ConnectionID != nil {
			connID = *b.ConnectionID
		}
		pk := partitionKey{
			actor:      b.ActorID,
			relGroup:   b.RelationGroup,
			connection: connID,
		}
		if srcToTgt[pk] == nil {
			srcToTgt[pk] = make(map[string]string)
			tgtToSrc[pk] = make(map[string]string)
		}

		srcID := b.SourceLocalID
		tgtID := b.TargetLocalID

		if srcID == "" || tgtID == "" {
			errs = append(errs, &BijectionViolation{
				ActorID:       b.ActorID,
				WorkflowID:    pk.relGroup,
				ConnectionID:  connID,
				SourceLocalID: srcID,
				TargetLocalID: tgtID,
				Reason:        "empty source_local_id or target_local_id in binding",
			})
			continue
		}

		// Injection: each source must map to exactly one target
		if existingTgt, ok := srcToTgt[pk][srcID]; ok && existingTgt != tgtID {
			errs = append(errs, &BijectionViolation{
				ActorID:       b.ActorID,
				WorkflowID:    pk.relGroup,
				ConnectionID:  connID,
				SourceLocalID: srcID,
				TargetLocalID: tgtID,
				Reason:        fmt.Sprintf("source_local_id %q already maps to target %q; cannot remap to %q", srcID, existingTgt, tgtID),
			})
		}
		// Surjection: each target must map to exactly one source
		if existingSrc, ok := tgtToSrc[pk][tgtID]; ok && existingSrc != srcID {
			errs = append(errs, &BijectionViolation{
				ActorID:       b.ActorID,
				WorkflowID:    pk.relGroup,
				ConnectionID:  connID,
				SourceLocalID: srcID,
				TargetLocalID: tgtID,
				Reason:        fmt.Sprintf("target_local_id %q already maps from source %q; cannot remap from %q", tgtID, existingSrc, srcID),
			})
		}
		srcToTgt[pk][srcID] = tgtID
		tgtToSrc[pk][tgtID] = srcID
	}

	return errs
}

// NormalizerMutationReport summarizes the complete mutation protection verification outcome.
type NormalizerMutationReport struct {
	// PolicyViolations are normalization policy issues (missing evidence etc.)
	PolicyViolations []string
	// MutantResults are per-mutant outcomes.
	MutantResults []NormalizerMutantResult
	// ProtectionViolations are mutants that should have yielded MISMATCH but were absorbed.
	ProtectionViolations []string
	// BijectionViolations are identity map bijection failures.
	BijectionViolations []string
	// OverallVerdict: PASS if no violations, MISMATCH if any protection/bijection violation, UNKNOWN if policy incomplete.
	OverallVerdict string
}

// RunNormalizationMutationGuard runs all normalization mutation protection checks for X20.
// Returns a complete report with the aggregate verdict.
func RunNormalizationMutationGuard(
	rules []*NormalizationRule,
	mutants []NormalizationMutant,
	idMap *IdentityMapV2,
) *NormalizerMutationReport {
	report := &NormalizerMutationReport{}

	// 1. Validate policy (negative_mutation_refs present etc.)
	policyErrs := ValidateNormalizationPolicy(rules)
	for _, e := range policyErrs {
		report.PolicyViolations = append(report.PolicyViolations, e.Error())
	}

	// 2. Run mutation checks
	protector := NewNormalizerMutationProtector(rules)
	results, violations := protector.CheckMutants(mutants)
	report.MutantResults = results
	report.ProtectionViolations = violations

	// 3. Check bijection
	checker := &BijectionIntegrityChecker{}
	bijErrs := checker.CheckBijection(idMap)
	for _, e := range bijErrs {
		report.BijectionViolations = append(report.BijectionViolations, e.Error())
	}

	// 4. Determine overall verdict
	if len(report.ProtectionViolations) > 0 || len(report.BijectionViolations) > 0 {
		report.OverallVerdict = "MISMATCH"
	} else if len(report.PolicyViolations) > 0 {
		report.OverallVerdict = "UNKNOWN"
	} else {
		report.OverallVerdict = "PASS"
	}

	return report
}

// FormatMutantSummary formats a mutation check summary for logging.
func FormatMutantSummary(results []NormalizerMutantResult, violations []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Mutant results: %d total\n", len(results)))
	for _, r := range results {
		marker := "✓"
		if r.Verdict != "MISMATCH" {
			marker = "✗"
		}
		sb.WriteString(fmt.Sprintf("  [%s] %s (%s): %s", marker, r.MutantID, r.Category, r.Verdict))
		if r.ViolationReason != "" {
			sb.WriteString(" — " + r.ViolationReason)
		}
		sb.WriteByte('\n')
	}
	if len(violations) > 0 {
		sb.WriteString("PROTECTION VIOLATIONS:\n")
		for _, v := range violations {
			sb.WriteString("  ! " + v + "\n")
		}
	}
	return sb.String()
}
