package domain

import (
	"fmt"
	"sort"
)

// GenerationAdmissionChecks is the guarded engine's evidence-resolution port.
// Implementations validate actual artifacts/currentness, not agent assertions.
// ValidateClosure must validate the exact plan envelope, checkpoint/TIM/target,
// Gate 2 receipt and revocation, current source qualification and source-only
// provenance. It must not require evidence outside scope except real prerequisites.
// Remaining methods require sufficient executable contracts, source evidence,
// qualified target lowerings and current completed dependencies respectively.
// Missing, stale, contaminated or UNKNOWN evidence must return an error.
// This port does not replace the publication fence or independent verification.
type GenerationAdmissionChecks interface {
	ValidateClosure(plan *PlanEnvelopeRecord, scope *AcceptanceScope) error
	RequireContract(ref string) error
	RequireSourceCase(caseID string) error
	RequireTargetCapability(capabilityID string) error
	RequireCompletedPrerequisite(unitID string) error
}

// GenerationAdmission is generation readiness, never a verifier PASS or a release
// authorization. ExcludedCaseIDs remain PROJECT obligations. UnresolvedItems are
// retained verbatim, including limitations outside this unit; they are not waived.
// No CLI/storage record is created by this internal evaluation.
type GenerationAdmission struct {
	PlanRef         string
	SnapshotRef     string
	TargetModelRef  string
	Scope           *AcceptanceScope
	ExcludedCaseIDs []string
	ContractRefs    []string
	Capabilities    []string
	PrerequisiteIDs []string
	UnresolvedItems []PlanApprovalUnresolvedItem
	Outcome         string
	Blockers        []string
}

// EvaluateGenerationAdmission derives UNIT obligations from the exact immutable
// plan and checks only those obligations, plus its real dependencies. It never
// accepts an agent-supplied case filter and cannot authorize application execution.
// The guarded caller must revalidate currentness/fences before staging/publication.
func EvaluateGenerationAdmission(plan *PlanEnvelopeRecord, unitID string, checks GenerationAdmissionChecks) (*GenerationAdmission, error) {
	// Round-trip validated canonical bytes so neither checks nor results can mutate
	// the supplied plan, its slices or the source artifacts that established it.
	raw, err := plan.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	sealed, err := ParsePlanEnvelope(raw)
	if err != nil {
		return nil, err
	}
	subject := sealed.Subject
	builder := NewPlanGraphBuilder(subject.ProjectID, subject.CorpusRef, subject.TargetModelRef, nil)
	for _, u := range subject.OrderedUnits {
		builder.AddUnit(SchedulableBehavioralUnit{UnitID: u.UnitID, ContractRefs: u.ContractRefs, CaseIDs: u.CaseIDs, Priority: int64(u.Priority), Capabilities: u.Capabilities})
	}
	for _, e := range subject.DependencyEdges {
		builder.AddDependency(e.PrerequisiteID, e.DependentID)
	}
	graph, err := builder.Build()
	if err != nil {
		return nil, err
	}
	scope, err := graph.DeriveUnitScope(unitID)
	if err != nil {
		return nil, err
	}
	project, err := graph.DeriveProjectScope()
	if err != nil {
		return nil, err
	}
	planRef, err := sealed.Digest()
	if err != nil {
		return nil, err
	}
	u := graph.Units[unitID]
	result := &GenerationAdmission{PlanRef: planRef, SnapshotRef: subject.SnapshotRef, TargetModelRef: subject.TargetModelRef, Scope: scope, ExcludedCaseIDs: []string{}, ContractRefs: append([]string{}, u.ContractRefs...), Capabilities: append([]string{}, u.Capabilities...), PrerequisiteIDs: append([]string{}, u.PrerequisiteIDs...), UnresolvedItems: subject.UnresolvedItems, Outcome: "BLOCKED", Blockers: []string{}}
	included := make(map[string]bool, len(scope.CaseIDs))
	for _, id := range scope.CaseIDs {
		included[id] = true
	}
	for _, id := range project.CaseIDs {
		if !included[id] {
			result.ExcludedCaseIDs = append(result.ExcludedCaseIDs, id)
		}
	}
	if checks == nil {
		result.Blockers = append(result.Blockers, "missing guarded evidence resolver")
		return result, nil
	}
	// Give the guard a separate scope so it cannot weaken subsequent checks by
	// changing the derived obligation set.
	guardScope := *scope
	guardScope.UnitIDs = append([]string{}, scope.UnitIDs...)
	guardScope.CaseIDs = append([]string{}, scope.CaseIDs...)
	guardPlan, err := ParsePlanEnvelope(raw)
	if err != nil {
		return nil, err
	}
	if err := checks.ValidateClosure(guardPlan, &guardScope); err != nil {
		result.Blockers = append(result.Blockers, "current closure: "+err.Error())
		return result, nil
	}
	// An unresolved plan item affecting this unit or one of its required contracts,
	// cases/capabilities/prerequisites cannot disappear behind the scope boundary.
	affected := map[string]bool{unitID: true}
	for _, ids := range [][]string{result.ContractRefs, scope.CaseIDs, result.Capabilities, result.PrerequisiteIDs} {
		for _, id := range ids {
			affected[id] = true
		}
	}
	for _, item := range result.UnresolvedItems {
		if item.Blocking {
			for _, id := range item.AffectedIDs {
				if affected[id] {
					result.Blockers = append(result.Blockers, "unresolved obligation: "+item.Code)
					break
				}
			}
		}
	}
	for _, requirement := range []struct {
		kind  string
		ids   []string
		check func(string) error
	}{
		{"contract", result.ContractRefs, checks.RequireContract},
		{"source case", scope.CaseIDs, checks.RequireSourceCase},
		{"target capability", result.Capabilities, checks.RequireTargetCapability},
		{"prerequisite", result.PrerequisiteIDs, checks.RequireCompletedPrerequisite},
	} {
		for _, id := range requirement.ids {
			if err := requirement.check(id); err != nil {
				result.Blockers = append(result.Blockers, fmt.Sprintf("%s %s: %s", requirement.kind, id, err))
			}
		}
	}
	sort.Strings(result.Blockers)
	if len(result.Blockers) == 0 {
		result.Outcome = "ADMITTED"
	}
	return result, nil
}
