package domain

import (
	"encoding/json"
)

// OmissionProofError describes an invalid omission proof record without echoing private contents.
type OmissionProofError struct {
	Reason string
	Cause  error
}

// Error identifies the failed omission proof obligation.
func (e *OmissionProofError) Error() string {
	return "invalid omission proof: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *OmissionProofError) Unwrap() error { return e.Cause }

// ValidateOmissionProof validates an omission proof bundle against specs/analysis/omission-proof.schema.json.
// It enforces schema version 2.0, project identity, unique omitted component IDs,
// all required evidence references (snapshot, profile, operational boundary, completeness,
// graph, analyzer, reachability, candidates, corpus, runtime hits, validity graph), and no unexpected fields.
func ValidateOmissionProof(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &OmissionProofError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &OmissionProofError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"omitted_ids",
		"snapshot_ref",
		"profile_ref",
		"closed_operational_boundary_ref",
		"parse_completeness_ref",
		"graph_ref",
		"analyzer_manifest_ref",
		"no_reachable_path_ref",
		"no_unresolved_candidate_ref",
		"usable_corpus_ref",
		"zero_runtime_hits_ref",
		"validity_graph_ref",
	}

	for _, field := range expectedFields {
		val, present := record[field]
		if !present || string(val) == "null" {
			return &OmissionProofError{Reason: "missing required field: " + field}
		}
	}
	if len(record) != len(expectedFields) {
		return &OmissionProofError{Reason: "unexpected fields"}
	}

	var version string
	if err := json.Unmarshal(record["schema_version"], &version); err != nil || version != "2.0" {
		return &OmissionProofError{Reason: "unsupported schema version"}
	}

	var projectID string
	if err := json.Unmarshal(record["project_id"], &projectID); err != nil || !validAssessmentReportID(projectID) {
		return &OmissionProofError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}

	var omittedIDs []string
	if err := json.Unmarshal(record["omitted_ids"], &omittedIDs); err != nil || len(omittedIDs) == 0 {
		return &OmissionProofError{Reason: "omitted_ids must be a non-empty array"}
	}
	seenIDs := make(map[string]struct{}, len(omittedIDs))
	for _, id := range omittedIDs {
		if !validAssessmentReportID(id) {
			return &OmissionProofError{Reason: "omitted_ids item must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenIDs[id]; exists {
			return &OmissionProofError{Reason: "duplicate item in omitted_ids"}
		}
		seenIDs[id] = struct{}{}
	}

	refFields := []string{
		"snapshot_ref",
		"profile_ref",
		"closed_operational_boundary_ref",
		"parse_completeness_ref",
		"graph_ref",
		"analyzer_manifest_ref",
		"no_reachable_path_ref",
		"no_unresolved_candidate_ref",
		"usable_corpus_ref",
		"zero_runtime_hits_ref",
		"validity_graph_ref",
	}
	for _, rf := range refFields {
		if !validProfileRef(record[rf], false) {
			return &OmissionProofError{Reason: rf + " must be a valid sha256 reference"}
		}
	}

	return nil
}
