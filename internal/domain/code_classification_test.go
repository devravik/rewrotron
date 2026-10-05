package domain

import (
	"strings"
	"testing"
)

func validOmissionProofForTest(omittedID string, snapRef, profRef, graphRef string) *OmissionProof {
	dummyRef := "sha256:" + strings.Repeat("e", 64)
	return &OmissionProof{
		SchemaVersion:                "2.0",
		ProjectID:                    "test-proj",
		OmittedIDs:                   []string{omittedID},
		SnapshotRef:                  snapRef,
		ProfileRef:                   profRef,
		ClosedOperationalBoundaryRef: dummyRef,
		ParseCompletenessRef:         dummyRef,
		GraphRef:                     graphRef,
		AnalyzerManifestRef:          dummyRef,
		NoReachablePathRef:           dummyRef,
		NoUnresolvedCandidateRef:     dummyRef,
		UsableCorpusRef:              dummyRef,
		ZeroRuntimeHitsRef:           dummyRef,
		ValidityGraphRef:             dummyRef,
	}
}

// TestCodeClassification_X76_Positive tests Test X76 setup and positive controls conforming to
// docs/10-development/TESTING.md and docs/03-analysis/CODE-CLASSIFICATION.md.
func TestCodeClassification_X76_Positive(t *testing.T) {
	snapRef := "sha256:" + strings.Repeat("a", 64)
	profRef := "sha256:" + strings.Repeat("b", 64)
	graphRef := "sha256:" + strings.Repeat("c", 64)
	corpusRef := "sha256:" + strings.Repeat("d", 64)

	omissionProof := validOmissionProofForTest("cand_dead", snapRef, profRef, graphRef)

	req := &ClassificationRequest{
		ProjectID:        "test-proj",
		Candidates:       []string{"cand_observed", "cand_reachable", "cand_unresolved", "cand_dead", "cand_unreachable_no_proof"},
		OperationalRoots: []string{"root_http"},
		Edges: []DirectedEdge{
			{FromID: "root_http", ToID: "cand_observed", Resolved: true},
			{FromID: "root_http", ToID: "cand_reachable", Resolved: true},
			{FromID: "root_http", ToID: "cand_unresolved", Resolved: false}, // unresolved edge
		},
		UnresolvedNodes: map[string][]string{
			"cand_unresolved": {"dynamic method dispatch unresolved"},
		},
		ObservedHits: map[string][]string{
			"cand_observed": {"ev-run-1"},
		},
		ClosedBoundary: true,
		OmissionProof:  omissionProof,
		GraphRef:       graphRef,
		ProfileRef:     profRef,
		CorpusRef:      corpusRef,
		SnapshotRef:    snapRef,
	}

	report, err := ClassifyCode(req)
	if err != nil {
		t.Fatalf("expected successful classification, got: %v", err)
	}

	// Invariant 1: Exactly one classification per candidate
	if len(report.Classifications) != 5 {
		t.Fatalf("expected 5 classifications, got %d", len(report.Classifications))
	}

	// Candidate 1: cand_observed must be ACTIVE
	cObs, ok := report.Classifications["cand_observed"]
	if !ok || cObs.State != ClassificationActive {
		t.Fatalf("expected cand_observed to be ACTIVE, got %v", cObs.State)
	}
	if len(cObs.EvidenceIDs) == 0 {
		t.Fatal("expected evidence_ids on ACTIVE candidate")
	}

	// Candidate 2: cand_reachable must be DORMANT (static reachability alone is never ACTIVE)
	cReach, ok := report.Classifications["cand_reachable"]
	if !ok || cReach.State != ClassificationDormant {
		t.Fatalf("expected cand_reachable to be DORMANT, got %v", cReach.State)
	}
	if len(cReach.ReachableFrom) == 0 {
		t.Fatal("expected reachable_from roots on DORMANT candidate")
	}

	// Candidate 3: cand_unresolved must be UNKNOWN
	cUnres, ok := report.Classifications["cand_unresolved"]
	if !ok || cUnres.State != ClassificationUnknown {
		t.Fatalf("expected cand_unresolved to be UNKNOWN, got %v", cUnres.State)
	}
	if len(cUnres.Uncertainty) == 0 {
		t.Fatal("expected uncertainty details on UNKNOWN candidate")
	}

	// Candidate 4: cand_dead must be DEAD (satisfies entire omission bundle)
	cDead, ok := report.Classifications["cand_dead"]
	if !ok || cDead.State != ClassificationDead {
		t.Fatalf("expected cand_dead to be DEAD, got %v", cDead.State)
	}
	if cDead.OmissionProofRef == nil {
		t.Fatal("expected omission_proof_ref on DEAD candidate")
	}

	// Candidate 5: cand_unreachable_no_proof must be UNKNOWN (never weaker DEAD without omission proof)
	cUnreach, ok := report.Classifications["cand_unreachable_no_proof"]
	if !ok || cUnreach.State != ClassificationUnknown {
		t.Fatalf("expected cand_unreachable_no_proof to be UNKNOWN, got %v", cUnreach.State)
	}

	// Invariant 2: Counts match and sum to total
	if report.ActiveCount != 1 {
		t.Errorf("expected ActiveCount=1, got %d", report.ActiveCount)
	}
	if report.DormantCount != 1 {
		t.Errorf("expected DormantCount=1, got %d", report.DormantCount)
	}
	if report.DeadCount != 1 {
		t.Errorf("expected DeadCount=1, got %d", report.DeadCount)
	}
	if report.UnknownCount != 2 {
		t.Errorf("expected UnknownCount=2, got %d", report.UnknownCount)
	}
	if report.TotalCount != 5 {
		t.Errorf("expected TotalCount=5, got %d", report.TotalCount)
	}
	if report.ActiveCount+report.DormantCount+report.DeadCount+report.UnknownCount != report.TotalCount {
		t.Errorf("counts do not sum to TotalCount")
	}

	// Invariant 3: Provenance refs preserved
	if report.GraphRef != graphRef || report.ProfileRef != profRef || report.CorpusRef != corpusRef || report.SnapshotRef != snapRef {
		t.Errorf("provenance references not properly preserved in report")
	}

	// Invariant 4: Canonical JSON & Digest
	cjson, err := report.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if len(cjson) == 0 {
		t.Fatal("expected non-empty canonical JSON")
	}

	digest, err := report.Digest()
	if err != nil {
		t.Fatalf("Digest failed: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Fatalf("invalid digest: %s", digest)
	}
}

// TestCodeClassification_X76_Falsification_Mutants tests all negative controls and falsification mutants
// specified by Test X76:
// - Count reachable unobserved as ACTIVE and DORMANT (must be rejected)
// - Weaker DEAD (must never classify as DEAD without omission proof and closed boundary)
// - Converting DORMANT to DEAD (must fail)
// - Converting UNKNOWN to DEAD (must fail)
// - Open boundary attempting DEAD omission (must fail to UNKNOWN)
func TestCodeClassification_X76_Falsification_Mutants(t *testing.T) {
	snapRef := "sha256:" + strings.Repeat("a", 64)
	profRef := "sha256:" + strings.Repeat("b", 64)
	graphRef := "sha256:" + strings.Repeat("c", 64)
	corpusRef := "sha256:" + strings.Repeat("d", 64)

	baseReq := func() *ClassificationRequest {
		return &ClassificationRequest{
			ProjectID:        "test-proj",
			Candidates:       []string{"item1"},
			OperationalRoots: []string{"root"},
			Edges: []DirectedEdge{
				{FromID: "root", ToID: "item1", Resolved: true},
			},
			ClosedBoundary: true,
			GraphRef:       graphRef,
			ProfileRef:     profRef,
			CorpusRef:      corpusRef,
			SnapshotRef:    snapRef,
		}
	}

	t.Run("Mutant1_StaticReachabilityAloneNeverActive", func(t *testing.T) {
		// Reachable without runtime hit must be DORMANT, NEVER ACTIVE
		req := baseReq()
		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationActive {
			t.Fatal("FALSIFICATION: static reachability alone was classified as ACTIVE!")
		}
		if report.Classifications["item1"].State != ClassificationDormant {
			t.Fatalf("expected DORMANT, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant2_WeakerDEAD_OpenBoundaryFailsToUnknown", func(t *testing.T) {
		// Unreachable item under open boundary (ClosedBoundary=false) MUST NOT be DEAD
		req := baseReq()
		req.Edges = nil // item1 is unreachable
		req.ClosedBoundary = false
		req.OmissionProof = validOmissionProofForTest("item1", snapRef, profRef, graphRef)

		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationDead {
			t.Fatal("FALSIFICATION: candidate classified as DEAD under an open operational boundary!")
		}
		if report.Classifications["item1"].State != ClassificationUnknown {
			t.Fatalf("expected UNKNOWN, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant3_WeakerDEAD_MissingOmissionProofFailsToUnknown", func(t *testing.T) {
		// Unreachable item under closed boundary but without omission proof MUST NOT be DEAD
		req := baseReq()
		req.Edges = nil // unreachable
		req.ClosedBoundary = true
		req.OmissionProof = nil // no omission proof

		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationDead {
			t.Fatal("FALSIFICATION: candidate classified as DEAD without omission proof!")
		}
		if report.Classifications["item1"].State != ClassificationUnknown {
			t.Fatalf("expected UNKNOWN, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant4_WeakerDEAD_CandidateNotInOmissionProofFailsToUnknown", func(t *testing.T) {
		// Unreachable item not in omission proof omitted_ids MUST NOT be DEAD
		req := baseReq()
		req.Edges = nil // unreachable
		req.ClosedBoundary = true
		req.OmissionProof = validOmissionProofForTest("other_item", snapRef, profRef, graphRef)

		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationDead {
			t.Fatal("FALSIFICATION: candidate classified as DEAD when not included in omission proof!")
		}
		if report.Classifications["item1"].State != ClassificationUnknown {
			t.Fatalf("expected UNKNOWN, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant5_WeakerDEAD_IncomingUnresolvedCandidatePathFailsToUnknown", func(t *testing.T) {
		// Unreachable from operational roots, but has an incoming unresolved edge (e.g. dynamic include)
		req := baseReq()
		req.Edges = []DirectedEdge{
			{FromID: "unresolved_dynamic_caller", ToID: "item1", Resolved: false},
		}
		req.ClosedBoundary = true
		req.OmissionProof = validOmissionProofForTest("item1", snapRef, profRef, graphRef)

		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationDead {
			t.Fatal("FALSIFICATION: candidate with unresolved incoming candidate path was classified as DEAD!")
		}
		if report.Classifications["item1"].State != ClassificationUnknown {
			t.Fatalf("expected UNKNOWN, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant6_DormantCannotBecomeDeadEvenWithZeroRuntimeHits", func(t *testing.T) {
		// Reachable item even with zero hits and omission proof attempting to omit it MUST NOT be DEAD
		req := baseReq()
		req.ClosedBoundary = true
		req.OmissionProof = validOmissionProofForTest("item1", snapRef, profRef, graphRef)

		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationDead {
			t.Fatal("FALSIFICATION: reachable candidate classified as DEAD!")
		}
		if report.Classifications["item1"].State != ClassificationDormant {
			t.Fatalf("expected DORMANT, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant7_UnknownCannotBecomeDead", func(t *testing.T) {
		// Unresolved candidate even with omission proof MUST NOT be DEAD
		req := baseReq()
		req.UnresolvedNodes = map[string][]string{
			"item1": {"syntax parse failed"},
		}
		req.Edges = nil // unreachable
		req.ClosedBoundary = true
		req.OmissionProof = validOmissionProofForTest("item1", snapRef, profRef, graphRef)

		report, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Classifications["item1"].State == ClassificationDead {
			t.Fatal("FALSIFICATION: unresolved candidate classified as DEAD!")
		}
		if report.Classifications["item1"].State != ClassificationUnknown {
			t.Fatalf("expected UNKNOWN, got: %s", report.Classifications["item1"].State)
		}
	})

	t.Run("Mutant8_MismatchedOmissionProofSnapshotRefRejected", func(t *testing.T) {
		req := baseReq()
		mismatchedSnap := "sha256:" + strings.Repeat("f", 64)
		req.OmissionProof = validOmissionProofForTest("item1", mismatchedSnap, profRef, graphRef)

		_, err := ClassifyCode(req)
		if err == nil {
			t.Fatal("expected error for mismatched snapshot_ref in omission proof")
		}
		if !strings.Contains(err.Error(), "snapshot_ref") {
			t.Fatalf("expected snapshot_ref mismatch error, got: %v", err)
		}
	})
}
