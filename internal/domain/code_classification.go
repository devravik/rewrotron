package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CodeClassificationError describes an invalid code classification operation or invariant failure.
type CodeClassificationError struct {
	Reason string
	Cause  error
}

// Error identifies the failed code classification requirement.
func (e *CodeClassificationError) Error() string {
	return "invalid code classification: " + e.Reason
}

// Unwrap preserves underlying decoding and validation failures.
func (e *CodeClassificationError) Unwrap() error { return e.Cause }

// CodeClassificationState represents the mutually exclusive four-state code classification.
type CodeClassificationState string

const (
	// ClassificationActive indicates current trusted runtime evidence establishes execution.
	ClassificationActive CodeClassificationState = "ACTIVE"
	// ClassificationDormant indicates not observed, resolution is sufficient, and a known reachable path exists.
	ClassificationDormant CodeClassificationState = "DORMANT"
	// ClassificationDead indicates neither observed nor reachable under a closed operational boundary with a verified omission proof.
	ClassificationDead CodeClassificationState = "DEAD"
	// ClassificationUnknown indicates unobserved with remaining parsing/dispatch/environment/root uncertainty, or unreachable without closed omission proof.
	ClassificationUnknown CodeClassificationState = "UNKNOWN"
)

// DirectedEdge describes a dependency or invocation edge between source elements.
type DirectedEdge struct {
	FromID   string `json:"from_id"`
	ToID     string `json:"to_id"`
	Resolved bool   `json:"resolved"`
}

// CandidateClassification records the exact classification outcome and provenance for a single candidate.
type CandidateClassification struct {
	CandidateID      string                  `json:"candidate_id"`
	State            CodeClassificationState `json:"state"`
	EvidenceIDs      []string                `json:"evidence_ids,omitempty"`
	Uncertainty      []string                `json:"uncertainty,omitempty"`
	ReachableFrom    []string                `json:"reachable_from,omitempty"`
	OmissionProofRef *string                 `json:"omission_proof_ref,omitempty"`
	Rationale        string                  `json:"rationale"`
}

// ClassificationRequest defines the inputs required to conservatively classify candidate code elements.
type ClassificationRequest struct {
	ProjectID        string              `json:"project_id"`
	Candidates       []string            `json:"candidates"`
	OperationalRoots []string            `json:"operational_roots"`
	Edges            []DirectedEdge      `json:"edges"`
	UnresolvedNodes  map[string][]string `json:"unresolved_nodes,omitempty"`
	ObservedHits     map[string][]string `json:"observed_hits,omitempty"`
	ClosedBoundary   bool                `json:"closed_boundary"`
	OmissionProof    *OmissionProof      `json:"omission_proof,omitempty"`
	GraphRef         string              `json:"graph_ref"`
	ProfileRef       string              `json:"profile_ref"`
	CorpusRef        string              `json:"corpus_ref"`
	SnapshotRef      string              `json:"snapshot_ref"`
}

// ClassificationReport records the mutually exclusive classifications and invariant counts for all candidates.
type ClassificationReport struct {
	ProjectID        string                            `json:"project_id"`
	GraphRef         string                            `json:"graph_ref"`
	ProfileRef       string                            `json:"profile_ref"`
	CorpusRef        string                            `json:"corpus_ref"`
	SnapshotRef      string                            `json:"snapshot_ref"`
	Classifications  map[string]CandidateClassification `json:"classifications"`
	ActiveCount      int                               `json:"active_count"`
	DormantCount     int                               `json:"dormant_count"`
	DeadCount        int                               `json:"dead_count"`
	UnknownCount     int                               `json:"unknown_count"`
	TotalCount       int                               `json:"total_count"`
	OmissionProofRef *string                           `json:"omission_proof_ref,omitempty"`
}

// CanonicalJSON renders the ClassificationReport into canonical JSON bytes.
func (r *ClassificationReport) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, &CodeClassificationError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical ClassificationReport.
func (r *ClassificationReport) Digest() (string, error) {
	cjson, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ClassifyCode executes conservative code classification conforming to docs/03-analysis/CODE-CLASSIFICATION.md
// and Test X76 invariants.
//
// Rules enforced:
// 1. Classification is strictly mutually exclusive across ACTIVE / DORMANT / DEAD / UNKNOWN.
// 2. Exact evaluation precedence:
//    - ACTIVE if current trusted runtime evidence establishes execution.
//    - UNKNOWN if not observed and relevant parsing/dispatch/environment/root uncertainty remains.
//    - DORMANT if not observed, resolution is sufficient, and a known reachable path exists from an operational root.
//    - DEAD only if neither observed nor reachable AND the entire omission bundle is satisfied (closed operational boundary, verified omission proof).
//    - Otherwise UNKNOWN.
// 3. Static reachability alone is NEVER ACTIVE.
// 4. Counts use exactly one state per candidate and sum to total candidates.
// 5. DEAD means no reachable evidence under approved analysis boundary B, never "proven never used".
func ClassifyCode(req *ClassificationRequest) (*ClassificationReport, error) {
	if req == nil {
		return nil, &CodeClassificationError{Reason: "request must not be nil"}
	}
	if !validAssessmentReportID(req.ProjectID) {
		return nil, &CodeClassificationError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !validRefDigest(req.GraphRef) {
		return nil, &CodeClassificationError{Reason: "graph_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}
	if !validRefDigest(req.ProfileRef) {
		return nil, &CodeClassificationError{Reason: "profile_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}
	if !validRefDigest(req.CorpusRef) {
		return nil, &CodeClassificationError{Reason: "corpus_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}
	if !validRefDigest(req.SnapshotRef) {
		return nil, &CodeClassificationError{Reason: "snapshot_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}
	if len(req.Candidates) == 0 {
		return nil, &CodeClassificationError{Reason: "candidates list must not be empty"}
	}

	seenCandidates := make(map[string]struct{}, len(req.Candidates))
	for _, c := range req.Candidates {
		if !validAssessmentReportID(c) {
			return nil, &CodeClassificationError{Reason: fmt.Sprintf("candidate %q must match pattern ^[a-z0-9_-]+$", c)}
		}
		if _, dup := seenCandidates[c]; dup {
			return nil, &CodeClassificationError{Reason: fmt.Sprintf("duplicate candidate %q", c)}
		}
		seenCandidates[c] = struct{}{}
	}

	// Validate omission proof if supplied
	var omissionDigest *string
	omittedIDSet := make(map[string]struct{})
	if req.OmissionProof != nil {
		rawOP, err := json.Marshal(req.OmissionProof)
		if err != nil {
			return nil, &CodeClassificationError{Reason: "omission proof serialization failure", Cause: err}
		}
		if err := ValidateOmissionProof(rawOP); err != nil {
			return nil, &CodeClassificationError{Reason: "invalid omission proof bundle", Cause: err}
		}
		if req.OmissionProof.SnapshotRef != req.SnapshotRef {
			return nil, &CodeClassificationError{Reason: "omission proof snapshot_ref does not match request snapshot_ref"}
		}
		if req.OmissionProof.ProfileRef != req.ProfileRef {
			return nil, &CodeClassificationError{Reason: "omission proof profile_ref does not match request profile_ref"}
		}
		if req.OmissionProof.GraphRef != req.GraphRef {
			return nil, &CodeClassificationError{Reason: "omission proof graph_ref does not match request graph_ref"}
		}
		d, err := req.OmissionProof.Digest()
		if err != nil {
			return nil, &CodeClassificationError{Reason: "omission proof digest computation failed", Cause: err}
		}
		omissionDigest = &d
		for _, oid := range req.OmissionProof.OmittedIDs {
			omittedIDSet[oid] = struct{}{}
		}
	}

	// Build graph representation: forward adjacency list and incoming edges
	forwardAdj := make(map[string][]DirectedEdge)
	incomingEdges := make(map[string][]DirectedEdge)
	for _, edge := range req.Edges {
		forwardAdj[edge.FromID] = append(forwardAdj[edge.FromID], edge)
		incomingEdges[edge.ToID] = append(incomingEdges[edge.ToID], edge)
	}

	// Precompute reachability from operational roots using BFS
	// rootReachable[c] = list of root IDs that can reach c via at least one path
	// fullyResolvedPathExists[c] = true if at least one path from a root to c has all edges resolved and no intermediate unresolved nodes
	// hasUnresolvedEdgeOnPath[c] = true if reachability paths encounter unresolved edges or unresolved nodes
	rootReachable := make(map[string][]string)
	fullyResolvedPathExists := make(map[string]bool)

	for _, rootID := range req.OperationalRoots {
		// BFS from rootID
		type queueItem struct {
			nodeID   string
			resolved bool
		}

		visited := make(map[string]bool)
		queue := []queueItem{{nodeID: rootID, resolved: len(req.UnresolvedNodes[rootID]) == 0}}
		visited[rootID] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if curr.nodeID != rootID {
				rootReachable[curr.nodeID] = append(rootReachable[curr.nodeID], rootID)
				if curr.resolved && len(req.UnresolvedNodes[curr.nodeID]) == 0 {
					fullyResolvedPathExists[curr.nodeID] = true
				}
			}

			for _, edge := range forwardAdj[curr.nodeID] {
				nextResolved := curr.resolved && edge.Resolved && len(req.UnresolvedNodes[curr.nodeID]) == 0
				if !visited[edge.ToID] {
					visited[edge.ToID] = true
					queue = append(queue, queueItem{nodeID: edge.ToID, resolved: nextResolved})
				} else if nextResolved && !fullyResolvedPathExists[edge.ToID] {
					// Found an alternate resolved path
					fullyResolvedPathExists[edge.ToID] = true
				}
			}
		}
	}

	// Classify each candidate in strict precedence order
	classifications := make(map[string]CandidateClassification, len(req.Candidates))
	activeCount := 0
	dormantCount := 0
	deadCount := 0
	unknownCount := 0

	for _, c := range req.Candidates {
		hits := req.ObservedHits[c]
		roots := rootReachable[c]
		nodeUncertainty := req.UnresolvedNodes[c]

		var classification CandidateClassification

		// 1. ACTIVE: Current trusted runtime evidence establishes execution
		if len(hits) > 0 {
			classification = CandidateClassification{
				CandidateID: c,
				State:       ClassificationActive,
				EvidenceIDs: hits,
				Uncertainty: nodeUncertainty,
				Rationale:   "Observed in trusted runtime observation evidence",
			}
			activeCount++
		} else if len(nodeUncertainty) > 0 || (len(roots) > 0 && !fullyResolvedPathExists[c]) {
			// 2. UNKNOWN: Not observed and relevant parsing/dispatch/environment/root uncertainty remains
			var reasons []string
			if len(nodeUncertainty) > 0 {
				reasons = append(reasons, nodeUncertainty...)
			}
			if len(roots) > 0 && !fullyResolvedPathExists[c] {
				reasons = append(reasons, "reachable path traverses unresolved edge or dynamic dispatch")
			}
			classification = CandidateClassification{
				CandidateID: c,
				State:       ClassificationUnknown,
				Uncertainty: reasons,
				Rationale:   "Not observed at runtime and relevant parsing/dispatch/environment uncertainty remains",
			}
			unknownCount++
		} else if len(roots) > 0 && fullyResolvedPathExists[c] {
			// 3. DORMANT: Not observed, resolution is sufficient, and a known reachable path exists
			sort.Strings(roots)
			classification = CandidateClassification{
				CandidateID:   c,
				State:         ClassificationDormant,
				ReachableFrom: roots,
				Rationale:     "Reachable from operational roots with sufficient resolution but unobserved at runtime",
			}
			dormantCount++
		} else {
			// 4. DEAD: Neither observed nor reachable AND the entire omission bundle is satisfied
			// Condition check:
			// - unobserved: checked (hits == 0)
			// - unreachable: checked (len(roots) == 0)
			// - closed operational boundary: req.ClosedBoundary == true
			// - omission proof provided: req.OmissionProof != nil
			// - candidate in omitted IDs: _, ok := omittedIDSet[c]
			// - no incoming unresolved candidate edges targeting c
			hasUnresolvedIncoming := false
			for _, inEdge := range incomingEdges[c] {
				if !inEdge.Resolved || len(req.UnresolvedNodes[inEdge.FromID]) > 0 {
					hasUnresolvedIncoming = true
					break
				}
			}

			_, inOmissionProof := omittedIDSet[c]

			if req.ClosedBoundary && req.OmissionProof != nil && inOmissionProof && !hasUnresolvedIncoming {
				classification = CandidateClassification{
					CandidateID:      c,
					State:            ClassificationDead,
					OmissionProofRef: omissionDigest,
					Rationale:        "Neither observed nor reachable under approved closed operational boundary with verified omission proof",
				}
				deadCount++
			} else {
				// 5. Otherwise UNKNOWN (e.g. boundary open, missing omission proof, or unresolved candidate paths)
				var reasons []string
				if !req.ClosedBoundary {
					reasons = append(reasons, "operational root boundary is open (unclosed inventory)")
				}
				if req.OmissionProof == nil {
					reasons = append(reasons, "omission proof bundle is absent")
				} else if !inOmissionProof {
					reasons = append(reasons, "candidate not included in omission proof omitted_ids")
				}
				if hasUnresolvedIncoming {
					reasons = append(reasons, "unresolved candidate path targeting candidate")
				}

				classification = CandidateClassification{
					CandidateID: c,
					State:       ClassificationUnknown,
					Uncertainty: reasons,
					Rationale:   "Unobserved and unreachable, but omission bundle is incomplete or boundary is unclosed",
				}
				unknownCount++
			}
		}

		classifications[c] = classification
	}

	total := len(req.Candidates)
	if activeCount+dormantCount+deadCount+unknownCount != total {
		return nil, &CodeClassificationError{
			Reason: fmt.Sprintf("classification counts do not sum to total candidates: %d != %d",
				activeCount+dormantCount+deadCount+unknownCount, total),
		}
	}

	report := &ClassificationReport{
		ProjectID:        req.ProjectID,
		GraphRef:         req.GraphRef,
		ProfileRef:       req.ProfileRef,
		CorpusRef:        req.CorpusRef,
		SnapshotRef:      req.SnapshotRef,
		Classifications:  classifications,
		ActiveCount:      activeCount,
		DormantCount:     dormantCount,
		DeadCount:        deadCount,
		UnknownCount:     unknownCount,
		TotalCount:       total,
		OmissionProofRef: omissionDigest,
	}

	return report, nil
}

func validRefDigest(ref string) bool {
	if !strings.HasPrefix(ref, "sha256:") {
		return false
	}
	return isManifestDigest(strings.TrimPrefix(ref, "sha256:"))
}
