package domain

import (
	"encoding/json"
)

// AssessmentPlanningBasisError describes an invalid assessment-planning-basis record without echoing private contents.
type AssessmentPlanningBasisError struct {
	Reason string
	Cause  error
}

// Error identifies the failed assessment-planning-basis obligation.
func (e *AssessmentPlanningBasisError) Error() string {
	return "invalid assessment planning basis: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *AssessmentPlanningBasisError) Unwrap() error { return e.Cause }

// ValidateAssessmentPlanningBasis validates an assessment-planning-basis 1.0 record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, SHA-256 references, current input refs with valid roles,
// validity graph ref, reused findings with required witness and prerequisite refs,
// disjointness among reused, stale, and new finding refs, and structured changes since assessment.
//
// This check does not execute planning, verify artifact availability on disk,
// prove feasibility, or grant execution authority.
func ValidateAssessmentPlanningBasis(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &AssessmentPlanningBasisError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &AssessmentPlanningBasisError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"assessment_ref",
		"selection_ref",
		"selected_scope_ref",
		"current_input_refs",
		"validity_graph_ref",
		"reused_findings",
		"stale_finding_refs",
		"new_finding_refs",
		"changes_since_assessment",
	}

	if len(record) != len(expectedFields) {
		return &AssessmentPlanningBasisError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &AssessmentPlanningBasisError{Reason: "missing required field"}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &AssessmentPlanningBasisError{Reason: "unsupported schema version"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &AssessmentPlanningBasisError{Reason: "project identity"}
	}

	for _, field := range []string{"assessment_ref", "selection_ref", "selected_scope_ref", "validity_graph_ref"} {
		if !validProfileRef(record[field], false) {
			return &AssessmentPlanningBasisError{Reason: field}
		}
	}

	if err := validateBasisCurrentInputRefs(record["current_input_refs"]); err != nil {
		return err
	}

	reusedSet, err := validateBasisReusedFindings(record["reused_findings"])
	if err != nil {
		return err
	}

	staleSet, err := validateBasisFindingRefsArray(record["stale_finding_refs"], "stale finding reference")
	if err != nil {
		return err
	}

	newSet, err := validateBasisFindingRefsArray(record["new_finding_refs"], "new finding reference")
	if err != nil {
		return err
	}

	for ref := range reusedSet {
		if _, exists := staleSet[ref]; exists {
			return &AssessmentPlanningBasisError{Reason: "reused finding cannot be stale"}
		}
		if _, exists := newSet[ref]; exists {
			return &AssessmentPlanningBasisError{Reason: "reused finding cannot be new"}
		}
	}
	for ref := range staleSet {
		if _, exists := newSet[ref]; exists {
			return &AssessmentPlanningBasisError{Reason: "stale finding cannot be new"}
		}
	}

	return validateBasisChangesSinceAssessment(record["changes_since_assessment"])
}

func validateBasisCurrentInputRefs(raw json.RawMessage) error {
	var inputRefs []json.RawMessage
	if err := json.Unmarshal(raw, &inputRefs); err != nil || len(inputRefs) < 1 {
		return &AssessmentPlanningBasisError{Reason: "current input references array", Cause: err}
	}

	seenTuples := make(map[string]struct{}, len(inputRefs))
	for _, rawRef := range inputRefs {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(rawRef, &item); err != nil || len(item) != 2 {
			return &AssessmentPlanningBasisError{Reason: "current input reference fields"}
		}
		roleRaw, hasRole := item["role"]
		refRaw, hasRef := item["ref"]
		if !hasRole || !hasRef {
			return &AssessmentPlanningBasisError{Reason: "missing current input reference field"}
		}

		role := profileString(roleRaw)
		switch role {
		case "SOURCE_SNAPSHOT", "DETECTION", "TARGET_INTENT", "OPERATIONAL_PROFILE",
			"CONSTRAINTS", "BUSINESS_REQUIREMENTS", "ANALYSIS", "OBSERVATION",
			"DEPENDENCY", "ADAPTER", "POLICY", "EVIDENCE":
		default:
			return &AssessmentPlanningBasisError{Reason: "unsupported current input reference role"}
		}

		if !validProfileRef(refRaw, false) {
			return &AssessmentPlanningBasisError{Reason: "current input reference"}
		}

		key := role + "|" + profileString(refRaw)
		if _, exists := seenTuples[key]; exists {
			return &AssessmentPlanningBasisError{Reason: "duplicate current input reference"}
		}
		seenTuples[key] = struct{}{}
	}
	return nil
}

func validateBasisReusedFindings(raw json.RawMessage) (map[string]struct{}, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || items == nil {
		return nil, &AssessmentPlanningBasisError{Reason: "reused findings array", Cause: err}
	}

	seenRecords := make(map[string]struct{}, len(items))
	expectedFields := []string{"analysis_record_ref", "checked_prerequisite_refs", "witness_refs"}

	for _, rawItem := range items {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(rawItem, &item); err != nil || len(item) != len(expectedFields) {
			return nil, &AssessmentPlanningBasisError{Reason: "reused finding fields"}
		}
		for _, field := range expectedFields {
			if _, present := item[field]; !present {
				return nil, &AssessmentPlanningBasisError{Reason: "missing reused finding field"}
			}
		}

		recordRefRaw := item["analysis_record_ref"]
		if !validProfileRef(recordRefRaw, false) {
			return nil, &AssessmentPlanningBasisError{Reason: "reused finding analysis record reference"}
		}

		recordRef := profileString(recordRefRaw)
		if _, seen := seenRecords[recordRef]; seen {
			return nil, &AssessmentPlanningBasisError{Reason: "duplicate reused finding"}
		}
		seenRecords[recordRef] = struct{}{}

		if _, err := validateBasisFindingRefsArray(item["checked_prerequisite_refs"], "checked prerequisite reference"); err != nil {
			return nil, err
		}
		if lenPrereqs, err := countBasisRefs(item["checked_prerequisite_refs"]); err != nil || lenPrereqs < 1 {
			return nil, &AssessmentPlanningBasisError{Reason: "reused finding requires checked prerequisite references"}
		}

		witnessRefs, err := validateBasisFindingRefsArray(item["witness_refs"], "witness reference")
		if err != nil {
			return nil, err
		}
		if len(witnessRefs) < 1 {
			return nil, &AssessmentPlanningBasisError{Reason: "reused finding requires witness references"}
		}
	}
	return seenRecords, nil
}

func countBasisRefs(raw json.RawMessage) (int, error) {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil || refs == nil {
		return 0, err
	}
	return len(refs), nil
}

func validateBasisFindingRefsArray(raw json.RawMessage, itemReason string) (map[string]struct{}, error) {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil || refs == nil {
		return nil, &AssessmentPlanningBasisError{Reason: itemReason + " array", Cause: err}
	}

	seen := make(map[string]struct{}, len(refs))
	for _, rawRef := range refs {
		if !validProfileRef(rawRef, false) {
			return nil, &AssessmentPlanningBasisError{Reason: itemReason}
		}
		refStr := profileString(rawRef)
		if _, exists := seen[refStr]; exists {
			return nil, &AssessmentPlanningBasisError{Reason: "duplicate " + itemReason}
		}
		seen[refStr] = struct{}{}
	}
	return seen, nil
}

func validateBasisChangesSinceAssessment(raw json.RawMessage) error {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || items == nil {
		return &AssessmentPlanningBasisError{Reason: "changes since assessment array", Cause: err}
	}

	expectedFields := []string{"input_role", "before_ref", "after_ref", "reason", "provenance_refs"}
	for _, rawItem := range items {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(rawItem, &item); err != nil || len(item) != len(expectedFields) {
			return &AssessmentPlanningBasisError{Reason: "change item fields"}
		}
		for _, field := range expectedFields {
			if _, present := item[field]; !present {
				return &AssessmentPlanningBasisError{Reason: "missing change item field"}
			}
		}

		role := profileString(item["input_role"])
		switch role {
		case "SOURCE_SNAPSHOT", "DETECTION", "TARGET_INTENT", "OPERATIONAL_PROFILE",
			"CONSTRAINTS", "BUSINESS_REQUIREMENTS", "ANALYSIS", "OBSERVATION",
			"DEPENDENCY", "ADAPTER", "POLICY", "EVIDENCE", "DECISION", "SCOPE":
		default:
			return &AssessmentPlanningBasisError{Reason: "unsupported change input role"}
		}

		beforeRaw := item["before_ref"]
		beforeIsNull := string(beforeRaw) == "null"
		if !beforeIsNull && !validProfileRef(beforeRaw, false) {
			return &AssessmentPlanningBasisError{Reason: "change before reference"}
		}

		afterRaw := item["after_ref"]
		afterIsNull := string(afterRaw) == "null"
		if !afterIsNull && !validProfileRef(afterRaw, false) {
			return &AssessmentPlanningBasisError{Reason: "change after reference"}
		}

		if beforeIsNull && afterIsNull {
			return &AssessmentPlanningBasisError{Reason: "change requires before or after reference"}
		}
		if !beforeIsNull && !afterIsNull && profileString(beforeRaw) == profileString(afterRaw) {
			return &AssessmentPlanningBasisError{Reason: "change before and after references cannot be identical"}
		}

		var reason string
		if string(item["reason"]) == "null" || json.Unmarshal(item["reason"], &reason) != nil || len(reason) < 1 {
			return &AssessmentPlanningBasisError{Reason: "change reason"}
		}

		provRefs, err := validateBasisFindingRefsArray(item["provenance_refs"], "provenance reference")
		if err != nil {
			return err
		}
		if len(provRefs) < 1 {
			return &AssessmentPlanningBasisError{Reason: "change requires provenance references"}
		}
	}
	return nil
}
