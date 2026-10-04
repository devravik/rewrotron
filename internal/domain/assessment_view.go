package domain

import (
	"encoding/json"
)

// AssessmentViewError describes an invalid assessment-view record without echoing private contents.
type AssessmentViewError struct {
	Reason string
	Cause  error
}

// Error identifies the failed assessment-view obligation.
func (e *AssessmentViewError) Error() string {
	return "invalid assessment view: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *AssessmentViewError) Unwrap() error { return e.Cause }

// ValidateAssessmentView validates an assessment-view 1.0 projection record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, SHA-256 references, embedded report integrity,
// current state sequence consistency, validity enums, and finding-level coverage
// and validity constraints.
//
// This check does not execute analysis, verify artifact availability on disk,
// prove feasibility, or grant execution authority.
func ValidateAssessmentView(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &AssessmentViewError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &AssessmentViewError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"report_ref",
		"report",
		"is_latest_report",
		"current_state_sequence",
		"validity",
		"current_validity_graph_ref",
		"finding_validity",
	}

	if len(record) != len(expectedFields) {
		return &AssessmentViewError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &AssessmentViewError{Reason: "missing required field"}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &AssessmentViewError{Reason: "unsupported schema version"}
	}

	projectID := profileString(record["project_id"])
	if !validAssessmentReportID(projectID) {
		return &AssessmentViewError{Reason: "project identity"}
	}

	if !validProfileRef(record["report_ref"], false) {
		return &AssessmentViewError{Reason: "report reference"}
	}

	var isLatest bool
	if string(record["is_latest_report"]) == "null" || json.Unmarshal(record["is_latest_report"], &isLatest) != nil {
		return &AssessmentViewError{Reason: "is_latest_report"}
	}

	if err := ValidateAssessmentReport(record["report"]); err != nil {
		return &AssessmentViewError{Reason: "report", Cause: err}
	}

	var repSummary struct {
		ProjectID    string `json:"project_id"`
		AsOfSequence int64  `json:"as_of_sequence"`
		Findings     []struct {
			FindingID         string `json:"finding_id"`
			AnalysisRecordRef string `json:"analysis_record_ref"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(record["report"], &repSummary); err != nil {
		return &AssessmentViewError{Reason: "report summary", Cause: err}
	}

	if repSummary.ProjectID != projectID {
		return &AssessmentViewError{Reason: "project ID mismatch"}
	}

	var currentSeq int64
	if string(record["current_state_sequence"]) == "null" || json.Unmarshal(record["current_state_sequence"], &currentSeq) != nil || currentSeq < 0 {
		return &AssessmentViewError{Reason: "current_state_sequence"}
	}
	if currentSeq < repSummary.AsOfSequence {
		return &AssessmentViewError{Reason: "sequence regression"}
	}

	validity := profileString(record["validity"])
	switch validity {
	case "VALID", "STALE", "UNKNOWN":
	default:
		return &AssessmentViewError{Reason: "unsupported validity"}
	}

	if validity == "VALID" {
		if string(record["current_validity_graph_ref"]) == "null" || !validProfileRef(record["current_validity_graph_ref"], false) {
			return &AssessmentViewError{Reason: "valid view requires validity graph reference"}
		}
	} else {
		if !validProfileRef(record["current_validity_graph_ref"], true) {
			return &AssessmentViewError{Reason: "validity graph reference"}
		}
	}

	var rawFindings []json.RawMessage
	if err := json.Unmarshal(record["finding_validity"], &rawFindings); err != nil || rawFindings == nil {
		return &AssessmentViewError{Reason: "finding validity array", Cause: err}
	}

	reportFindings := make(map[string]string, len(repSummary.Findings))
	for _, f := range repSummary.Findings {
		reportFindings[f.FindingID] = f.AnalysisRecordRef
	}

	if len(rawFindings) != len(reportFindings) {
		return &AssessmentViewError{Reason: "finding validity count mismatch"}
	}

	seenFindings := make(map[string]struct{}, len(rawFindings))
	hasStaleFinding := false
	hasUnknownFinding := false

	expectedItemFields := []string{"finding_id", "analysis_record_ref", "validity", "reason", "witness_refs"}
	for _, rawItem := range rawFindings {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(rawItem, &item); err != nil || len(item) != len(expectedItemFields) {
			return &AssessmentViewError{Reason: "finding validity fields"}
		}
		for _, f := range expectedItemFields {
			if _, present := item[f]; !present {
				return &AssessmentViewError{Reason: "missing finding validity field"}
			}
		}

		findingID := profileString(item["finding_id"])
		if !validAssessmentReportID(findingID) {
			return &AssessmentViewError{Reason: "finding ID"}
		}
		if _, seen := seenFindings[findingID]; seen {
			return &AssessmentViewError{Reason: "duplicate finding validity"}
		}
		seenFindings[findingID] = struct{}{}

		expectedRecordRef, found := reportFindings[findingID]
		if !found {
			return &AssessmentViewError{Reason: "unknown finding ID"}
		}

		if !validProfileRef(item["analysis_record_ref"], false) {
			return &AssessmentViewError{Reason: "finding analysis record reference"}
		}
		if profileString(item["analysis_record_ref"]) != expectedRecordRef {
			return &AssessmentViewError{Reason: "finding analysis record reference mismatch"}
		}

		itemValidity := profileString(item["validity"])
		switch itemValidity {
		case "VALID":
		case "STALE":
			hasStaleFinding = true
		case "UNKNOWN":
			hasUnknownFinding = true
		default:
			return &AssessmentViewError{Reason: "unsupported finding validity"}
		}

		if validity == "VALID" && itemValidity != "VALID" {
			return &AssessmentViewError{Reason: "valid view contains invalid finding"}
		}

		var reason string
		if string(item["reason"]) == "null" || json.Unmarshal(item["reason"], &reason) != nil || len(reason) < 1 {
			return &AssessmentViewError{Reason: "finding reason"}
		}

		var witnessRefs []json.RawMessage
		if err := json.Unmarshal(item["witness_refs"], &witnessRefs); err != nil || witnessRefs == nil {
			return &AssessmentViewError{Reason: "witness references array", Cause: err}
		}
		if itemValidity == "VALID" && len(witnessRefs) < 1 {
			return &AssessmentViewError{Reason: "valid finding requires witness references"}
		}
		seenWitnesses := make(map[string]struct{}, len(witnessRefs))
		for _, rawWitness := range witnessRefs {
			if !validProfileRef(rawWitness, false) {
				return &AssessmentViewError{Reason: "witness reference"}
			}
			wStr := profileString(rawWitness)
			if _, seen := seenWitnesses[wStr]; seen {
				return &AssessmentViewError{Reason: "duplicate witness reference"}
			}
			seenWitnesses[wStr] = struct{}{}
		}
	}

	if hasStaleFinding && validity != "STALE" {
		return &AssessmentViewError{Reason: "stale finding requires stale view"}
	}
	if hasUnknownFinding && !hasStaleFinding && validity == "VALID" {
		return &AssessmentViewError{Reason: "unknown finding prevents valid view"}
	}

	return nil
}
