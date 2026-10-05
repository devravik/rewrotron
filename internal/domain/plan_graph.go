package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// PlanGraphError describes an invalid plan graph or scheduling violation.
type PlanGraphError struct {
	Reason string
	Cause  error
}

// Error identifies the failed plan graph obligation.
func (e *PlanGraphError) Error() string {
	return "invalid plan graph: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *PlanGraphError) Unwrap() error { return e.Cause }

// SchedulableBehavioralUnit defines an executable behavioral unit in the reconstruction plan.
// A unit is a behavioral vertical slice of operations/workflows; shared code belongs to one
// explicit component owner in the Target Implementation Model (MIGRATION-UNITS.md:3).
type SchedulableBehavioralUnit struct {
	UnitID          string   `json:"unit_id"`
	ContractRefs    []string `json:"contract_refs"`
	CaseIDs         []string `json:"case_ids"`
	Priority        int64    `json:"priority"`
	Capabilities    []string `json:"capabilities"`
	PrerequisiteIDs []string `json:"prerequisite_ids"`
	ConsumerUnitIDs []string `json:"consumer_unit_ids"`
	ComponentID     string   `json:"component_id"`
}

// PlanDependencyEdge defines a directed prerequisite edge between units.
type PlanDependencyEdge struct {
	PrerequisiteID string `json:"prerequisite_id"`
	DependentID    string `json:"dependent_id"`
}

// PlanOrderedUnit matches ordered_units in specs/project/approval-subject.schema.json#/$defs/plan.
type PlanOrderedUnit struct {
	UnitID       string   `json:"unit_id"`
	ContractRefs []string `json:"contract_refs"`
	CaseIDs      []string `json:"case_ids"`
	Priority     int64    `json:"priority"`
	Capabilities []string `json:"capabilities"`
}

// ReconstructionPlanGraph models the executable DAG of schedulable units,
// topological ordering, shared consumer mappings, and acceptance scope derivation.
type ReconstructionPlanGraph struct {
	ProjectID       string                               `json:"project_id"`
	CorpusRef       string                               `json:"corpus_ref"`
	TargetModelRef  string                               `json:"target_model_ref"`
	Units           map[string]*SchedulableBehavioralUnit `json:"units"`
	OrderedUnits    []PlanOrderedUnit                    `json:"ordered_units"`
	DependencyEdges []PlanDependencyEdge                 `json:"dependency_edges"`
}

// CalculateScopeHash computes the rwt-cjson-1 SHA256 hex hash of an acceptance scope
// without the scope_hash field, per MIGRATION-UNITS.md:13.
func CalculateScopeHash(kind string, unitIDs, caseIDs []string, corpusRef string) (string, error) {
	// Ensure slices are sorted and non-nil
	sortedUnits := make([]string, len(unitIDs))
	copy(sortedUnits, unitIDs)
	sort.Strings(sortedUnits)

	sortedCases := make([]string, len(caseIDs))
	copy(sortedCases, caseIDs)
	sort.Strings(sortedCases)

	scopePayload := map[string]interface{}{
		"kind":       kind,
		"unit_ids":   sortedUnits,
		"case_ids":   sortedCases,
		"corpus_ref": corpusRef,
	}

	raw, err := json.Marshal(scopePayload)
	if err != nil {
		return "", fmt.Errorf("marshal scope payload: %w", err)
	}

	canonicalBytes, err := CanonicalJSON(raw)
	if err != nil {
		return "", fmt.Errorf("canonicalize scope payload: %w", err)
	}

	h := sha256.Sum256(canonicalBytes)
	return hex.EncodeToString(h[:]), nil
}

// DeriveUnitScope builds the UNIT acceptance scope for a single unit.
// Unit can complete; future dependent cases need not exist in target yet (MIGRATION-UNITS.md:9, Test X35).
func (g *ReconstructionPlanGraph) DeriveUnitScope(unitID string) (*AcceptanceScope, error) {
	unit, ok := g.Units[unitID]
	if !ok {
		return nil, &PlanGraphError{Reason: fmt.Sprintf("unit %q not found in plan graph", unitID)}
	}

	caseMap := make(map[string]struct{}, len(unit.CaseIDs))
	for _, cid := range unit.CaseIDs {
		caseMap[cid] = struct{}{}
	}

	sortedCases := make([]string, 0, len(caseMap))
	for cid := range caseMap {
		sortedCases = append(sortedCases, cid)
	}
	sort.Strings(sortedCases)

	unitIDs := []string{unitID}
	scopeHash, err := CalculateScopeHash("UNIT", unitIDs, sortedCases, g.CorpusRef)
	if err != nil {
		return nil, &PlanGraphError{Reason: "failed to calculate UNIT scope hash", Cause: err}
	}

	scope := &AcceptanceScope{
		Kind:      "UNIT",
		UnitIDs:   unitIDs,
		CaseIDs:   sortedCases,
		CorpusRef: g.CorpusRef,
		ScopeHash: scopeHash,
	}
	return scope, nil
}

// DeriveImplementedRegressionScope builds the IMPLEMENTED_REGRESSION scope for currently implemented units,
// including all applicable approved cases of currently implemented units and shared component consumers (MIGRATION-UNITS.md:10).
func (g *ReconstructionPlanGraph) DeriveImplementedRegressionScope(implementedUnitIDs []string) (*AcceptanceScope, error) {
	if len(implementedUnitIDs) == 0 {
		return nil, &PlanGraphError{Reason: "implementedUnitIDs cannot be empty for regression scope"}
	}

	seenUnits := make(map[string]struct{}, len(implementedUnitIDs))
	caseMap := make(map[string]struct{})

	for _, uid := range implementedUnitIDs {
		unit, ok := g.Units[uid]
		if !ok {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("implemented unit %q not in plan graph", uid)}
		}
		seenUnits[uid] = struct{}{}
		for _, cid := range unit.CaseIDs {
			caseMap[cid] = struct{}{}
		}
	}

	sortedUnits := make([]string, 0, len(seenUnits))
	for uid := range seenUnits {
		sortedUnits = append(sortedUnits, uid)
	}
	sort.Strings(sortedUnits)

	sortedCases := make([]string, 0, len(caseMap))
	for cid := range caseMap {
		sortedCases = append(sortedCases, cid)
	}
	sort.Strings(sortedCases)

	if len(sortedCases) == 0 {
		return nil, &PlanGraphError{Reason: "regression scope must have at least one case_id"}
	}

	scopeHash, err := CalculateScopeHash("IMPLEMENTED_REGRESSION", sortedUnits, sortedCases, g.CorpusRef)
	if err != nil {
		return nil, &PlanGraphError{Reason: "failed to calculate regression scope hash", Cause: err}
	}

	return &AcceptanceScope{
		Kind:      "IMPLEMENTED_REGRESSION",
		UnitIDs:   sortedUnits,
		CaseIDs:   sortedCases,
		CorpusRef: g.CorpusRef,
		ScopeHash: scopeHash,
	}, nil
}

// DeriveProjectScope builds the PROJECT acceptance scope containing all in-scope units and cases (MIGRATION-UNITS.md:11).
func (g *ReconstructionPlanGraph) DeriveProjectScope() (*AcceptanceScope, error) {
	seenUnits := make(map[string]struct{}, len(g.Units))
	caseMap := make(map[string]struct{})

	for uid, u := range g.Units {
		seenUnits[uid] = struct{}{}
		for _, cid := range u.CaseIDs {
			caseMap[cid] = struct{}{}
		}
	}

	sortedUnits := make([]string, 0, len(seenUnits))
	for uid := range seenUnits {
		sortedUnits = append(sortedUnits, uid)
	}
	sort.Strings(sortedUnits)

	sortedCases := make([]string, 0, len(caseMap))
	for cid := range caseMap {
		sortedCases = append(sortedCases, cid)
	}
	sort.Strings(sortedCases)

	if len(sortedCases) == 0 {
		return nil, &PlanGraphError{Reason: "project scope must have at least one case_id"}
	}

	scopeHash, err := CalculateScopeHash("PROJECT", sortedUnits, sortedCases, g.CorpusRef)
	if err != nil {
		return nil, &PlanGraphError{Reason: "failed to calculate project scope hash", Cause: err}
	}

	return &AcceptanceScope{
		Kind:      "PROJECT",
		UnitIDs:   sortedUnits,
		CaseIDs:   sortedCases,
		CorpusRef: g.CorpusRef,
		ScopeHash: scopeHash,
	}, nil
}

// CanUnitComplete checks if unitID is eligible to complete per MIGRATION-UNITS.md:17 and Test X35:
// 1. Its UNIT cases pass.
// 2. All applicable implemented-regression cases pass.
// 3. Required prerequisites are current valid completed units.
// 4. Prerequisite acceptance NEVER depends on future dependent-unit cases (e.g. Login can complete before checkout).
func (g *ReconstructionPlanGraph) CanUnitComplete(unitID string, implementedUnitIDs []string, unitVerdicts map[string]string) (bool, string) {
	unit, ok := g.Units[unitID]
	if !ok {
		return false, fmt.Sprintf("unit %s not found in plan graph", unitID)
	}

	// 1. Check prerequisites: all prerequisites must be implemented and VERIFIED
	implSet := make(map[string]struct{}, len(implementedUnitIDs))
	for _, id := range implementedUnitIDs {
		implSet[id] = struct{}{}
	}

	for _, preID := range unit.PrerequisiteIDs {
		if _, implemented := implSet[preID]; !implemented {
			return false, fmt.Sprintf("prerequisite unit %s is not yet implemented", preID)
		}
		if verdict := unitVerdicts[preID]; verdict != "VERIFIED" {
			return false, fmt.Sprintf("prerequisite unit %s is not VERIFIED (status: %s)", preID, verdict)
		}
	}

	// 2. Unit's own verdict must be VERIFIED
	if verdict := unitVerdicts[unitID]; verdict != "VERIFIED" {
		return false, fmt.Sprintf("unit %s is not VERIFIED (status: %s)", unitID, verdict)
	}

	// Note: Future dependent units (e.g. checkout depending on login) do NOT block this unit (Test X35).
	return true, ""
}

// CanProjectComplete checks whole-project completion per MIGRATION-UNITS.md:17 and Test X35:
// Once all units are implemented, PROJECT acceptance still executes the complete approved corpus;
// a project missing any in-scope unit cannot achieve VERIFIED/READY_CUTOVER.
func (g *ReconstructionPlanGraph) CanProjectComplete(implementedUnitIDs []string, unitVerdicts map[string]string) (bool, string) {
	implSet := make(map[string]struct{}, len(implementedUnitIDs))
	for _, id := range implementedUnitIDs {
		implSet[id] = struct{}{}
	}

	for _, ou := range g.OrderedUnits {
		if _, implemented := implSet[ou.UnitID]; !implemented {
			return false, fmt.Sprintf("project missing unit %s (Test X35: cannot VERIFIED/READY_CUTOVER without all units)", ou.UnitID)
		}
		if verdict := unitVerdicts[ou.UnitID]; verdict != "VERIFIED" {
			return false, fmt.Sprintf("unit %s is not VERIFIED (status: %s)", ou.UnitID, verdict)
		}
	}
	return true, ""
}

// PlanGraphBuilder builds and validates ReconstructionPlanGraphs.
type PlanGraphBuilder struct {
	projectID      string
	corpusRef      string
	targetModelRef string
	targetModel    *TargetModelDocument
	units          map[string]*SchedulableBehavioralUnit
	edges          []PlanDependencyEdge
}

// NewPlanGraphBuilder creates a new builder for a reconstruction plan graph.
func NewPlanGraphBuilder(projectID, corpusRef, targetModelRef string, targetModel *TargetModelDocument) *PlanGraphBuilder {
	return &PlanGraphBuilder{
		projectID:      projectID,
		corpusRef:      corpusRef,
		targetModelRef: targetModelRef,
		targetModel:    targetModel,
		units:          make(map[string]*SchedulableBehavioralUnit),
		edges:          make([]PlanDependencyEdge, 0),
	}
}

// AddUnit registers a schedulable behavioral unit.
func (b *PlanGraphBuilder) AddUnit(u SchedulableBehavioralUnit) *PlanGraphBuilder {
	if u.ContractRefs == nil {
		u.ContractRefs = []string{}
	}
	if u.CaseIDs == nil {
		u.CaseIDs = []string{}
	}
	if u.Capabilities == nil {
		u.Capabilities = []string{}
	}
	if u.PrerequisiteIDs == nil {
		u.PrerequisiteIDs = []string{}
	}
	if u.ConsumerUnitIDs == nil {
		u.ConsumerUnitIDs = []string{}
	}
	unitCopy := u
	b.units[u.UnitID] = &unitCopy
	return b
}

// AddDependency adds a directed prerequisite edge: prerequisiteID must complete before dependentID.
func (b *PlanGraphBuilder) AddDependency(prerequisiteID, dependentID string) *PlanGraphBuilder {
	b.edges = append(b.edges, PlanDependencyEdge{
		PrerequisiteID: prerequisiteID,
		DependentID:    dependentID,
	})
	return b
}

// Build compiles and validates the ReconstructionPlanGraph, detecting cycles (MIGRATION-UNITS.md:3),
// enforcing topological ordering with priority and deterministic tie-breaking, and cross-checking
// shared consumers with the Target Implementation Model (Test X73, TARGET-ARCHITECTURE.md:9).
func (b *PlanGraphBuilder) Build() (*ReconstructionPlanGraph, error) {
	if !validAssessmentReportID(b.projectID) {
		return nil, &PlanGraphError{Reason: "invalid project_id format"}
	}
	if !isSHA256Ref(b.corpusRef) {
		return nil, &PlanGraphError{Reason: "invalid corpus_ref format"}
	}
	if !isSHA256Ref(b.targetModelRef) {
		return nil, &PlanGraphError{Reason: "invalid target_model_ref format"}
	}
	if len(b.units) == 0 {
		return nil, &PlanGraphError{Reason: "plan graph must contain at least one unit"}
	}

	// Validate units
	for uid, u := range b.units {
		if !validAssessmentReportID(uid) {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("invalid unit_id format %q", uid)}
		}
		if len(u.ContractRefs) == 0 {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("unit %s contract_refs must be non-empty", uid)}
		}
		for _, cr := range u.ContractRefs {
			if !isSHA256Ref(cr) {
				return nil, &PlanGraphError{Reason: fmt.Sprintf("unit %s invalid contract_ref %q", uid, cr)}
			}
		}
		if len(u.CaseIDs) == 0 {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("unit %s case_ids must be non-empty", uid)}
		}
		for _, cid := range u.CaseIDs {
			if !validAssessmentReportID(cid) {
				return nil, &PlanGraphError{Reason: fmt.Sprintf("unit %s invalid case_id %q", uid, cid)}
			}
		}
		if u.Priority < 0 {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("unit %s priority must be >= 0", uid)}
		}
	}

	// Validate and deduplicate dependency edges
	dedupEdges := make([]PlanDependencyEdge, 0, len(b.edges))
	seenEdgePairs := make(map[[2]string]struct{}, len(b.edges))

	for _, e := range b.edges {
		if !validAssessmentReportID(e.PrerequisiteID) {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("invalid prerequisite_id format %q", e.PrerequisiteID)}
		}
		if !validAssessmentReportID(e.DependentID) {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("invalid dependent_id format %q", e.DependentID)}
		}
		if e.PrerequisiteID == e.DependentID {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("self-loop dependency: unit %s cannot depend on itself (MIGRATION-UNITS.md:3)", e.PrerequisiteID)}
		}
		if _, ok := b.units[e.PrerequisiteID]; !ok {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("prerequisite_id %s not in plan graph units", e.PrerequisiteID)}
		}
		if _, ok := b.units[e.DependentID]; !ok {
			return nil, &PlanGraphError{Reason: fmt.Sprintf("dependent_id %s not in plan graph units", e.DependentID)}
		}

		pair := [2]string{e.PrerequisiteID, e.DependentID}
		if _, exists := seenEdgePairs[pair]; !exists {
			seenEdgePairs[pair] = struct{}{}
			dedupEdges = append(dedupEdges, e)
		}
	}

	// Sort deduplicated edges deterministically
	sort.Slice(dedupEdges, func(i, j int) bool {
		if dedupEdges[i].PrerequisiteID != dedupEdges[j].PrerequisiteID {
			return dedupEdges[i].PrerequisiteID < dedupEdges[j].PrerequisiteID
		}
		return dedupEdges[i].DependentID < dedupEdges[j].DependentID
	})

	// Build adjacency and in-degrees
	inDegree := make(map[string]int, len(b.units))
	dependents := make(map[string][]string, len(b.units))
	prerequisites := make(map[string][]string, len(b.units))

	for uid := range b.units {
		inDegree[uid] = 0
		dependents[uid] = make([]string, 0)
		prerequisites[uid] = make([]string, 0)
	}

	for _, e := range dedupEdges {
		inDegree[e.DependentID]++
		dependents[e.PrerequisiteID] = append(dependents[e.PrerequisiteID], e.DependentID)
		prerequisites[e.DependentID] = append(prerequisites[e.DependentID], e.PrerequisiteID)
	}

	// Populate PrerequisiteIDs and ConsumerUnitIDs on unit structs
	for uid, u := range b.units {
		pres := prerequisites[uid]
		sort.Strings(pres)
		u.PrerequisiteIDs = pres

		cons := dependents[uid]
		sort.Strings(cons)
		u.ConsumerUnitIDs = cons
	}

	// Topological sort with priority and deterministic tie-breaking (MIGRATION-UNITS.md:3)
	// At each step, available units (in-degree == 0) are ordered by:
	// 1. Higher priority first (descending)
	// 2. Lexicographical unit_id ascending
	var orderedUnits []PlanOrderedUnit

	for len(orderedUnits) < len(b.units) {
		var available []string
		for uid, deg := range inDegree {
			if deg == 0 {
				available = append(available, uid)
			}
		}

		if len(available) == 0 {
			// Cycle detected!
			return nil, &PlanGraphError{
				Reason: "cyclic execution DAG is a plan blocker: cycle prevents scheduling remaining units (MIGRATION-UNITS.md:3)",
			}
		}

		// Sort available units by priority descending, then unit_id ascending
		sort.Slice(available, func(i, j int) bool {
			pi := b.units[available[i]].Priority
			pj := b.units[available[j]].Priority
			if pi != pj {
				return pi > pj
			}
			return available[i] < available[j]
		})

		// Pick the highest priority / lowest ID unit
		selected := available[0]
		u := b.units[selected]

		sortedContracts := make([]string, len(u.ContractRefs))
		copy(sortedContracts, u.ContractRefs)
		sort.Strings(sortedContracts)

		sortedCases := make([]string, len(u.CaseIDs))
		copy(sortedCases, u.CaseIDs)
		sort.Strings(sortedCases)

		sortedCaps := make([]string, len(u.Capabilities))
		copy(sortedCaps, u.Capabilities)
		sort.Strings(sortedCaps)

		orderedUnits = append(orderedUnits, PlanOrderedUnit{
			UnitID:       u.UnitID,
			ContractRefs: sortedContracts,
			CaseIDs:      sortedCases,
			Priority:     u.Priority,
			Capabilities: sortedCaps,
		})

		// Remove selected from inDegree map (mark as scheduled)
		delete(inDegree, selected)

		// Decrement in-degree of its dependents
		for _, dep := range dependents[selected] {
			if _, exists := inDegree[dep]; exists {
				inDegree[dep]--
			}
		}
	}

	// Cross-validate with Target Implementation Model if provided
	if b.targetModel != nil {
		for _, ou := range orderedUnits {
			comp, ok := b.targetModel.ComponentByOwner(ou.UnitID)
			if !ok {
				continue
			}
			// Invariant Test X73 & TARGET-ARCHITECTURE.md:9:
			// "Component owner_unit_id is unique; consumer units and prerequisites must agree with the plan DAG."
			// Every consumer unit declared in TIM must depend on this unit in the plan graph
			graphConsumers := make(map[string]struct{}, len(dependents[ou.UnitID]))
			for _, dep := range dependents[ou.UnitID] {
				graphConsumers[dep] = struct{}{}
			}

			for _, timConsumer := range comp.ConsumerUnitIDs {
				if _, ok := graphConsumers[timConsumer]; !ok {
					return nil, &PlanGraphError{
						Reason: fmt.Sprintf("missing consumption edge: component %q lists consumer %q but no dependency edge exists in plan graph (Test X73)", comp.ComponentID, timConsumer),
					}
				}
			}
		}
	}

	graph := &ReconstructionPlanGraph{
		ProjectID:       b.projectID,
		CorpusRef:       b.corpusRef,
		TargetModelRef:  b.targetModelRef,
		Units:           b.units,
		OrderedUnits:    orderedUnits,
		DependencyEdges: dedupEdges,
	}

	return graph, nil
}
