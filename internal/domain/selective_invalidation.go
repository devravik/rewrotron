package domain

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// ArtifactValidityState indicates the derived validity state of an artifact in the project.
type ArtifactValidityState string

const (
	// ValidityStateValid indicates the artifact's prerequisites are unchanged and covered by positive evidence.
	ValidityStateValid ArtifactValidityState = "VALID"

	// ValidityStateStale indicates the artifact is invalidated due to changed prerequisites or unclosed coverage.
	ValidityStateStale ArtifactValidityState = "STALE"

	// ValidityStateUnknown indicates the artifact has unknown impact or missing prerequisite evidence.
	ValidityStateUnknown ArtifactValidityState = "UNKNOWN"
)

// InvalidationDetail describes the state and rationale for an artifact's validity.
type InvalidationDetail struct {
	State  ArtifactValidityState `json:"state"`
	Reason string                `json:"reason"`
}

// SelectiveInvalidationResult contains the computed validity states for all evaluated artifacts.
type SelectiveInvalidationResult struct {
	GraphRef       string                           `json:"graph_ref"`
	ChangedRoots   []string                         `json:"changed_roots"`
	ArtifactStates map[string]ArtifactValidityState `json:"artifact_states"`
	Details        map[string]InvalidationDetail    `json:"details"`
}

// IsArtifactValid returns true if the artifact is in the VALID state.
func (r *SelectiveInvalidationResult) IsArtifactValid(ref string) bool {
	if r == nil || r.ArtifactStates == nil {
		return false
	}
	return r.ArtifactStates[ref] == ValidityStateValid
}

// IsArtifactStale returns true if the artifact is in the STALE state.
func (r *SelectiveInvalidationResult) IsArtifactStale(ref string) bool {
	if r == nil || r.ArtifactStates == nil {
		return false
	}
	return r.ArtifactStates[ref] == ValidityStateStale
}

// IsArtifactUnknown returns true if the artifact is in the UNKNOWN state.
func (r *SelectiveInvalidationResult) IsArtifactUnknown(ref string) bool {
	if r == nil || r.ArtifactStates == nil {
		return false
	}
	return r.ArtifactStates[ref] == ValidityStateUnknown
}

// GetReason returns the invalidation rationale for a given reference.
func (r *SelectiveInvalidationResult) GetReason(ref string) string {
	if r == nil || r.Details == nil {
		return ""
	}
	return r.Details[ref].Reason
}

// StaleRefs returns a sorted list of all stale artifact references.
func (r *SelectiveInvalidationResult) StaleRefs() []string {
	if r == nil || r.ArtifactStates == nil {
		return []string{}
	}
	var stale []string
	for ref, state := range r.ArtifactStates {
		if state == ValidityStateStale {
			stale = append(stale, ref)
		}
	}
	sort.Strings(stale)
	return stale
}

// ValidRefs returns a sorted list of all valid artifact references.
func (r *SelectiveInvalidationResult) ValidRefs() []string {
	if r == nil || r.ArtifactStates == nil {
		return []string{}
	}
	var valid []string
	for ref, state := range r.ArtifactStates {
		if state == ValidityStateValid {
			valid = append(valid, ref)
		}
	}
	sort.Strings(valid)
	return valid
}

// InvalidationError describes a failure during selective invalidation or reuse evaluation.
type InvalidationError struct {
	Reason string
	Cause  error
}

// Error identifies the invalidation failure reason.
func (e *InvalidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalidation error: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("invalidation error: %s", e.Reason)
}

// Unwrap preserves the underlying error cause.
func (e *InvalidationError) Unwrap() error { return e.Cause }

// SelectiveInvalidationService executes typed invalidation over a validity graph,
// enforces positive prerequisite consumption coverage, refuses invalid reuse,
// and protects Gate 2 execution against stale approvals.
type SelectiveInvalidationService struct {
	graph          *ValidityGraphRecord
	graphRef       string
	dependentsOf   map[string][]string // from_ref -> to_refs (dependents of prerequisite)
	prereqsOf      map[string][]string // to_ref -> from_refs (prerequisites of dependent)
	edgesByPair    map[string]ValidityEdge
	allNodes       map[string]struct{}
	coveragesByRef map[string]*ConsumptionCoverageRecord
	examinedInputs map[string]struct{}
}

// NewSelectiveInvalidationService constructs and indexes a SelectiveInvalidationService.
func NewSelectiveInvalidationService(graph *ValidityGraphRecord, coverages []*ConsumptionCoverageRecord) (*SelectiveInvalidationService, error) {
	if graph == nil {
		return nil, &InvalidationError{Reason: "validity graph must not be nil"}
	}
	graphDigest, err := graph.Digest()
	if err != nil {
		return nil, &InvalidationError{Reason: "failed to compute graph digest", Cause: err}
	}

	svc := &SelectiveInvalidationService{
		graph:          graph,
		graphRef:       graphDigest,
		dependentsOf:   make(map[string][]string),
		prereqsOf:      make(map[string][]string),
		edgesByPair:    make(map[string]ValidityEdge),
		allNodes:       make(map[string]struct{}),
		coveragesByRef: make(map[string]*ConsumptionCoverageRecord),
		examinedInputs: make(map[string]struct{}),
	}

	for _, cov := range coverages {
		if cov == nil {
			continue
		}
		digest, err := cov.Digest()
		if err != nil {
			return nil, &InvalidationError{Reason: "failed to compute consumption coverage digest", Cause: err}
		}
		svc.coveragesByRef[digest] = cov
		for _, in := range cov.InputsExamined {
			svc.examinedInputs[in] = struct{}{}
		}
	}

	// In schema 2.1, all declared consumption coverage refs must be resolved
	if graph.SchemaVersion == "2.1" {
		for _, covRef := range graph.ConsumptionCoverageRefs {
			if _, exists := svc.coveragesByRef[covRef]; !exists {
				return nil, &InvalidationError{Reason: fmt.Sprintf("graph declares consumption coverage ref %s which was not provided", covRef)}
			}
		}
	}

	// Index edges and nodes
	for _, edge := range graph.Edges {
		svc.allNodes[edge.FromRef] = struct{}{}
		svc.allNodes[edge.ToRef] = struct{}{}
		svc.dependentsOf[edge.FromRef] = append(svc.dependentsOf[edge.FromRef], edge.ToRef)
		svc.prereqsOf[edge.ToRef] = append(svc.prereqsOf[edge.ToRef], edge.FromRef)
		pairKey := edge.FromRef + "|" + edge.ToRef
		svc.edgesByPair[pairKey] = edge
	}

	for _, ref := range graph.UnknownImpactRefs {
		svc.allNodes[ref] = struct{}{}
	}

	return svc, nil
}

// ComputeInvalidation computes the selective invalidation outcome across the graph given changed roots
// and optional consumed inputs.
// Invariants enforced:
//  1. Positive consumption coverage: missing consumption edges or unexamined inputs cause conservative invalidation (Test X73).
//  2. Conservative unknown impact: refs in unknown_impact_refs and their dependents are invalidated (Test X29, X48).
//  3. Transitive reverse closure: only dependents reachable from changed roots become stale; unaffected nodes remain valid (Test X29, X48, X49).
//  4. Old verification PASS results are never relabeled or reused across target/runtime changes (Test X29).
func (s *SelectiveInvalidationService) ComputeInvalidation(
	changedPrerequisiteRefs []string,
	consumedArtifactInputs map[string][]string,
) (*SelectiveInvalidationResult, error) {
	result := &SelectiveInvalidationResult{
		GraphRef:       s.graphRef,
		ChangedRoots:   changedPrerequisiteRefs,
		ArtifactStates: make(map[string]ArtifactValidityState),
		Details:        make(map[string]InvalidationDetail),
	}

	// Initialize all known nodes to VALID
	for node := range s.allNodes {
		result.ArtifactStates[node] = ValidityStateValid
		result.Details[node] = InvalidationDetail{
			State:  ValidityStateValid,
			Reason: "prerequisites unchanged with positive evidence",
		}
	}
	for artRef := range consumedArtifactInputs {
		if _, exists := result.ArtifactStates[artRef]; !exists {
			result.ArtifactStates[artRef] = ValidityStateValid
			result.Details[artRef] = InvalidationDetail{
				State:  ValidityStateValid,
				Reason: "prerequisites unchanged with positive evidence",
			}
		}
	}

	// Step 1: Check positive consumption coverage and missing consumption edges (Test X73)
	// "Missing consumption edge: Known source/shared-component input consumed by completed unit;
	//  Remove edge or entire graph and change consumed input: Reuse refused; descendants stale until coverage closes.
	//  No edge is not proof of independence."
	if consumedArtifactInputs != nil {
		for artRef, inputs := range consumedArtifactInputs {
			for _, inRef := range inputs {
				// Check whether an edge exists in the validity graph from inRef -> artRef
				pairKey := inRef + "|" + artRef
				if _, hasEdge := s.edgesByPair[pairKey]; !hasEdge {
					result.ArtifactStates[artRef] = ValidityStateStale
					result.Details[artRef] = InvalidationDetail{
						State:  ValidityStateStale,
						Reason: fmt.Sprintf("missing consumption edge: input %s consumed by %s has no validity edge; absent edge cannot prove independence (Test X73)", inRef, artRef),
					}
				}
				// Check whether the input was registered in examined inputs
				if _, examined := s.examinedInputs[inRef]; !examined {
					result.ArtifactStates[artRef] = ValidityStateStale
					result.Details[artRef] = InvalidationDetail{
						State:  ValidityStateStale,
						Reason: fmt.Sprintf("unregistered consumption: input %s consumed by %s is not in inputs_examined (Test X73)", inRef, artRef),
					}
				}
			}
		}
	}

	// Check coverage record statuses: if any coverage record is INSUFFICIENT or has unclosed categories,
	// invalidate affected categories conservatively
	for _, cov := range s.coveragesByRef {
		if cov.Status != "SUFFICIENT" {
			for _, cat := range cov.Categories {
				for _, inRef := range cat.InputRefs {
					// Mark dependents of this unclosed input as stale
					for _, depRef := range s.dependentsOf[inRef] {
						result.ArtifactStates[depRef] = ValidityStateStale
						result.Details[depRef] = InvalidationDetail{
							State:  ValidityStateStale,
							Reason: fmt.Sprintf("unclosed consumption coverage: input %s has coverage status %s (Test X73)", inRef, cov.Status),
						}
					}
				}
			}
		} else {
			for _, cat := range cov.Categories {
				if cat.Status != "CLOSED" && cat.Status != "IRRELEVANT" {
					for _, inRef := range cat.InputRefs {
						for _, depRef := range s.dependentsOf[inRef] {
							result.ArtifactStates[depRef] = ValidityStateStale
							result.Details[depRef] = InvalidationDetail{
								State:  ValidityStateStale,
								Reason: fmt.Sprintf("unclosed consumption coverage category %s status %s (Test X73)", cat.Category, cat.Status),
							}
						}
					}
				}
			}
		}
	}

	// Step 2: Conservative unknown impact handling (Test X29, X48)
	// Any ref in unknown_impact_refs is conservatively marked UNKNOWN
	for _, unkRef := range s.graph.UnknownImpactRefs {
		result.ArtifactStates[unkRef] = ValidityStateUnknown
		result.Details[unkRef] = InvalidationDetail{
			State:  ValidityStateUnknown,
			Reason: "conservative invalidation: explicitly listed in unknown_impact_refs (Test X29/X48)",
		}
	}

	// Step 3: Mark changed roots as STALE (Test X29, X48, X49)
	for _, root := range changedPrerequisiteRefs {
		result.ArtifactStates[root] = ValidityStateStale
		result.Details[root] = InvalidationDetail{
			State:  ValidityStateStale,
			Reason: fmt.Sprintf("changed prerequisite root: %s", root),
		}
	}

	// Step 4: Transitive forward propagation along FromRef -> ToRef (reverse dependency closure of dependents)
	// BFS queue initialized with all currently STALE or UNKNOWN nodes
	queue := make([]string, 0, len(result.ArtifactStates))
	for ref, state := range result.ArtifactStates {
		if state == ValidityStateStale || state == ValidityStateUnknown {
			queue = append(queue, ref)
		}
	}

	visited := make(map[string]struct{}, len(queue))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if _, seen := visited[curr]; seen {
			continue
		}
		visited[curr] = struct{}{}

		currDetail := result.Details[curr]

		// Propagate to all dependents of curr
		for _, dep := range s.dependentsOf[curr] {
			existingState := result.ArtifactStates[dep]
			if existingState != ValidityStateStale {
				result.ArtifactStates[dep] = ValidityStateStale
				result.Details[dep] = InvalidationDetail{
					State:  ValidityStateStale,
					Reason: fmt.Sprintf("transitive reverse dependency on changed/stale prerequisite %s: %s", curr, currDetail.Reason),
				}
				queue = append(queue, dep)
			}
		}
	}

	// Step 5: Verification PASS results non-relabel invariant (Test X29)
	// If any target-dependent prerequisite was changed, verify that old EXECUTION_RESULT / PERFORMANCE_RESULT edges
	// are marked STALE and cannot be relabeled.
	for _, edge := range s.graph.Edges {
		if edge.Kind == "EXECUTION_RESULT" || edge.Kind == "PERFORMANCE_RESULT" {
			fromState := result.ArtifactStates[edge.FromRef]
			if fromState == ValidityStateStale || fromState == ValidityStateUnknown {
				result.ArtifactStates[edge.ToRef] = ValidityStateStale
				result.Details[edge.ToRef] = InvalidationDetail{
					State:  ValidityStateStale,
					Reason: "verification result invalidated by prerequisite change; no old PASS relabeling permitted (Test X29)",
				}
			}
		}
	}

	return result, nil
}

// ReuseCertificateRequest specifies the arguments for evaluating and issuing a reuse certificate.
type ReuseCertificateRequest struct {
	ProjectID               string
	OldPlanHash             string
	NewPlanHash             string
	ArtifactRef             string
	Kind                    string // "CHARACTERIZATION" or "IMPLEMENTATION"
	CheckedEdgeRefs         []string
	EvidenceRefs            []string
	ConsumptionCoverageRefs []string
	IsVerificationResult    bool
}

// CanReuseArtifact checks whether an artifact may be selectively reused.
// Enforces Test X29 ("no old PASS relabel") and Test X73 ("positive coverage required").
func (s *SelectiveInvalidationService) CanReuseArtifact(
	invResult *SelectiveInvalidationResult,
	req ReuseCertificateRequest,
) (bool, error) {
	// Invariant Test X29: Old verification PASS results can NEVER be reused across target changes
	if req.IsVerificationResult || req.Kind == "VERIFICATION" {
		return false, &InvalidationError{
			Reason: "verification results cannot be relabeled or reused across target/environment changes; verification must rerun against successor target (Test X29)",
		}
	}

	// Kind must be CHARACTERIZATION or IMPLEMENTATION
	if req.Kind != "CHARACTERIZATION" && req.Kind != "IMPLEMENTATION" {
		return false, &InvalidationError{
			Reason: fmt.Sprintf("reuse certificate only permits CHARACTERIZATION or IMPLEMENTATION kinds, got %q (Test X29)", req.Kind),
		}
	}

	// Artifact must be currently VALID in invalidation result
	if invResult.IsArtifactStale(req.ArtifactRef) {
		return false, &InvalidationError{
			Reason: fmt.Sprintf("cannot reuse stale artifact %s: %s", req.ArtifactRef, invResult.GetReason(req.ArtifactRef)),
		}
	}
	if invResult.IsArtifactUnknown(req.ArtifactRef) {
		return false, &InvalidationError{
			Reason: fmt.Sprintf("cannot reuse artifact %s with unknown impact: %s", req.ArtifactRef, invResult.GetReason(req.ArtifactRef)),
		}
	}

	// Invariant Test X73: Positive consumption coverage must be present and SUFFICIENT
	if len(req.ConsumptionCoverageRefs) == 0 {
		return false, &InvalidationError{
			Reason: "positive consumption coverage references required for reuse certificate (Test X73)",
		}
	}
	for _, covRef := range req.ConsumptionCoverageRefs {
		cov, exists := s.coveragesByRef[covRef]
		if !exists {
			return false, &InvalidationError{
				Reason: fmt.Sprintf("referenced consumption coverage %s does not exist in service (Test X73)", covRef),
			}
		}
		if cov.Status != "SUFFICIENT" {
			return false, &InvalidationError{
				Reason: fmt.Sprintf("referenced consumption coverage %s has status %s; SUFFICIENT required (Test X73)", covRef, cov.Status),
			}
		}
	}

	// Checked edges must exist in the validity graph
	for _, edgeRef := range req.CheckedEdgeRefs {
		found := false
		for _, edge := range s.graph.Edges {
			witnessFound := false
			for _, w := range edge.WitnessRefs {
				if w == edgeRef {
					witnessFound = true
					break
				}
			}
			if edge.FromRef == edgeRef || edge.ToRef == edgeRef || witnessFound {
				found = true
				break
			}
		}
		if !found {
			return false, &InvalidationError{
				Reason: fmt.Sprintf("checked edge ref %s not found in validity graph (Test X73)", edgeRef),
			}
		}
	}

	return true, nil
}

// IssueReuseCertificate issues a valid 2.1 reuse certificate after verifying all invariants.
func (s *SelectiveInvalidationService) IssueReuseCertificate(
	invResult *SelectiveInvalidationResult,
	req ReuseCertificateRequest,
) (*ReuseCertificateRecord, error) {
	canReuse, err := s.CanReuseArtifact(invResult, req)
	if err != nil {
		return nil, err
	}
	if !canReuse {
		return nil, &InvalidationError{Reason: "artifact cannot be reused"}
	}

	cert := &ReuseCertificateRecord{
		SchemaVersion:           "2.1",
		ProjectID:               req.ProjectID,
		OldPlanHash:             req.OldPlanHash,
		NewPlanHash:             req.NewPlanHash,
		ArtifactRef:             req.ArtifactRef,
		GraphRef:                s.graphRef,
		Kind:                    req.Kind,
		CheckedEdgeRefs:         req.CheckedEdgeRefs,
		EvidenceRefs:            req.EvidenceRefs,
		Issuer:                  "ENGINE",
		Valid:                   true,
		ConsumptionCoverageRefs: req.ConsumptionCoverageRefs,
	}

	// Ensure canonical validation passes
	if _, err := cert.CanonicalJSON(); err != nil {
		return nil, &InvalidationError{Reason: "generated reuse certificate fails canonical validation", Cause: err}
	}

	return cert, nil
}

// ValidatePlanExecutionReadiness verifies whether a plan can be executed under Gate 2.
// Enforces Test X49:
// "Target change with history: PHP intent 8.3→8.4 makes affected dependency/target findings and
//  downstream plan stale while source-only facts and historical report bytes persist; stale approval cannot execute"
func ValidatePlanExecutionReadiness(
	currentPlanHash string,
	planArtifactRefs []string,
	invResult *SelectiveInvalidationResult,
	approvalReceiptPlanHash string,
	approvalReceiptDecisionRefs []string,
) error {
	// Plan hash must match the approved plan hash
	if currentPlanHash != approvalReceiptPlanHash {
		return &InvalidationError{
			Reason: fmt.Sprintf("stale approval cannot execute (Test X49): receipt plan hash %s does not match current plan hash %s", approvalReceiptPlanHash, currentPlanHash),
		}
	}

	// Check if any constituent plan artifact is stale
	if invResult != nil {
		for _, ref := range planArtifactRefs {
			if invResult.IsArtifactStale(ref) {
				return &InvalidationError{
					Reason: fmt.Sprintf("stale plan cannot execute (Test X49): plan artifact %s is stale (%s); approval is inoperative", ref, invResult.GetReason(ref)),
				}
			}
			if invResult.IsArtifactUnknown(ref) {
				return &InvalidationError{
					Reason: fmt.Sprintf("stale plan cannot execute (Test X49): plan artifact %s has unknown impact (%s); approval is inoperative", ref, invResult.GetReason(ref)),
				}
			}
		}
		for _, decRef := range approvalReceiptDecisionRefs {
			if invResult.IsArtifactStale(decRef) {
				return &InvalidationError{
					Reason: fmt.Sprintf("stale approval cannot execute (Test X49): approved decision ref %s is stale (%s); Gate 2 re-confirmation required", decRef, invResult.GetReason(decRef)),
				}
			}
		}
	}

	return nil
}

// MakeDummyRef generates a deterministic sha256 reference for testing.
func MakeDummyRef(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("sha256:%x", h)
}
