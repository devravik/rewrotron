package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// AssessmentReportError describes an invalid assessment-report record without echoing private contents.
type AssessmentReportError struct {
	Reason string
	Cause  error
}

// Error identifies the failed assessment-report obligation.
func (e *AssessmentReportError) Error() string {
	return "invalid assessment report: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *AssessmentReportError) Unwrap() error { return e.Cause }

// ValidateAssessmentReport validates an assessment-report 1.0 record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, SHA-256 references, feasibility enums, RFC 3339 timestamps,
// input-ref tuples, structured findings and their category/basis constraints, sequence steps,
// effort estimates, and the invariant that implementation_authorized must be false.
//
// This check does not execute analysis, verify artifact availability on disk,
// prove feasibility, or grant execution authority.
func ValidateAssessmentReport(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &AssessmentReportError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &AssessmentReportError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version", "assessment_id", "project_id", "assessment_status",
		"migration_feasibility", "implementation_authorized", "as_of_sequence",
		"created_at", "configuration_hash", "source_snapshot_ref", "detection_ref",
		"target_intent_ref", "scope_ref", "input_refs", "analysis_record_refs",
		"validity_graph_ref", "system_summary", "evidence_summary", "examined_scope",
		"findings", "missing_evidence", "assumptions", "uncertainty",
		"suggested_sequence", "effort_estimate",
	}

	if len(record) != len(expectedFields) {
		return &AssessmentReportError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &AssessmentReportError{Reason: "missing required field"}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &AssessmentReportError{Reason: "unsupported schema version"}
	}

	for _, field := range []string{"assessment_id", "project_id"} {
		val := profileString(record[field])
		if !validAssessmentReportID(val) {
			return &AssessmentReportError{Reason: field + " identity"}
		}
	}

	if profileString(record["assessment_status"]) != "COMPLETED" {
		return &AssessmentReportError{Reason: "assessment status"}
	}

	feasibility := profileString(record["migration_feasibility"])
	switch feasibility {
	case "VIABLE", "VIABLE_WITH_CONDITIONS", "BLOCKED", "UNKNOWN", "PARTIALLY_UNDERSTOOD":
	default:
		return &AssessmentReportError{Reason: "unsupported migration feasibility"}
	}

	// Invariant: implementation_authorized must be false.
	var implAuth bool
	if string(record["implementation_authorized"]) == "null" || json.Unmarshal(record["implementation_authorized"], &implAuth) != nil || implAuth {
		return &AssessmentReportError{Reason: "implementation cannot be authorized"}
	}

	var asOfSeq int64
	if string(record["as_of_sequence"]) == "null" || json.Unmarshal(record["as_of_sequence"], &asOfSeq) != nil || asOfSeq < 0 {
		return &AssessmentReportError{Reason: "as_of_sequence"}
	}

	var createdAt string
	if string(record["created_at"]) == "null" || json.Unmarshal(record["created_at"], &createdAt) != nil {
		return &AssessmentReportError{Reason: "created_at timestamp"}
	}
	if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
		return &AssessmentReportError{Reason: "created_at timestamp", Cause: err}
	}

	cfgHash := profileString(record["configuration_hash"])
	if !isManifestDigest(cfgHash) {
		return &AssessmentReportError{Reason: "configuration hash"}
	}

	for _, field := range []string{"source_snapshot_ref", "detection_ref", "scope_ref", "validity_graph_ref"} {
		if !validProfileRef(record[field], false) {
			return &AssessmentReportError{Reason: field + " reference"}
		}
	}

	if !validProfileRef(record["target_intent_ref"], true) {
		return &AssessmentReportError{Reason: "target intent reference"}
	}

	if err := validateAssessmentInputRefs(record["input_refs"]); err != nil {
		return err
	}

	if err := validateAssessmentRefsArray(record["analysis_record_refs"], "analysis record reference", 0); err != nil {
		return err
	}

	for _, field := range []string{"system_summary", "evidence_summary", "examined_scope"} {
		var text string
		if string(record[field]) == "null" || json.Unmarshal(record[field], &text) != nil || len(text) < 1 {
			return &AssessmentReportError{Reason: field}
		}
	}

	hasBlocker, err := validateAssessmentFindings(record["findings"])
	if err != nil {
		return err
	}

	if feasibility == "BLOCKED" && !hasBlocker {
		return &AssessmentReportError{Reason: "blocked assessment requires blocker or unsupported finding"}
	}

	if err := validateAssessmentStringArray(record["missing_evidence"], "missing evidence", 0); err != nil {
		return err
	}
	if err := validateAssessmentStringArray(record["assumptions"], "assumptions", 0); err != nil {
		return err
	}
	if err := validateAssessmentStringArray(record["uncertainty"], "uncertainty", 0); err != nil {
		return err
	}

	if err := validateAssessmentSequence(record["suggested_sequence"]); err != nil {
		return err
	}

	if err := validateAssessmentEffort(record["effort_estimate"]); err != nil {
		return err
	}

	return nil
}

func validAssessmentReportID(value string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) < 0
}

func validateAssessmentInputRefs(raw json.RawMessage) error {
	var inputs []json.RawMessage
	if err := json.Unmarshal(raw, &inputs); err != nil || len(inputs) < 1 {
		return &AssessmentReportError{Reason: "input references", Cause: err}
	}
	seen := make(map[[2]string]bool, len(inputs))
	for _, rawInput := range inputs {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(rawInput, &obj); err != nil || len(obj) != 2 {
			return &AssessmentReportError{Reason: "input reference object", Cause: err}
		}
		roleRaw, okRole := obj["role"]
		refRaw, okRef := obj["ref"]
		if !okRole || !okRef {
			return &AssessmentReportError{Reason: "input reference object fields"}
		}
		role := profileString(roleRaw)
		switch role {
		case "SOURCE_SNAPSHOT", "DETECTION", "TARGET_INTENT", "OPERATIONAL_PROFILE",
			"CONSTRAINTS", "BUSINESS_REQUIREMENTS", "ANALYSIS", "OBSERVATION",
			"DEPENDENCY", "ADAPTER", "POLICY", "EVIDENCE":
		default:
			return &AssessmentReportError{Reason: "input reference role"}
		}
		if !validProfileRef(refRaw, false) {
			return &AssessmentReportError{Reason: "input reference artifact"}
		}
		key := [2]string{role, profileString(refRaw)}
		if seen[key] {
			return &AssessmentReportError{Reason: "duplicate input reference"}
		}
		seen[key] = true
	}
	return nil
}

func validateAssessmentRefsArray(raw json.RawMessage, label string, minItems int) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil || len(refs) < minItems {
		return &AssessmentReportError{Reason: label + " array", Cause: err}
	}
	seen := make(map[string]bool, len(refs))
	for _, rawRef := range refs {
		if string(rawRef) == "null" || !validProfileRef(rawRef, false) {
			return &AssessmentReportError{Reason: label + " item"}
		}
		val := profileString(rawRef)
		if seen[val] {
			return &AssessmentReportError{Reason: "duplicate " + label}
		}
		seen[val] = true
	}
	return nil
}

func validateAssessmentStringArray(raw json.RawMessage, label string, minItems int) error {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) < minItems {
		return &AssessmentReportError{Reason: label + " array", Cause: err}
	}
	seen := make(map[string]bool, len(items))
	for _, rawItem := range items {
		var text string
		if string(rawItem) == "null" || json.Unmarshal(rawItem, &text) != nil || len(text) < 1 {
			return &AssessmentReportError{Reason: label + " item"}
		}
		if seen[text] {
			return &AssessmentReportError{Reason: "duplicate " + label}
		}
		seen[text] = true
	}
	return nil
}

func validateAssessmentFindings(raw json.RawMessage) (bool, error) {
	var findings []json.RawMessage
	if err := json.Unmarshal(raw, &findings); err != nil || findings == nil {
		return false, &AssessmentReportError{Reason: "findings array", Cause: err}
	}
	seenIDs := make(map[string]bool, len(findings))
	hasBlocker := false
	expectedFindingFields := []string{
		"finding_id", "analysis_record_ref", "classification", "basis", "category",
		"summary", "reason", "provenance_refs", "prerequisite_refs", "limitations",
		"affected_ids", "recommended_action",
	}

	for _, rawFinding := range findings {
		var f map[string]json.RawMessage
		if err := json.Unmarshal(rawFinding, &f); err != nil || len(f) != len(expectedFindingFields) {
			return false, &AssessmentReportError{Reason: "finding object fields"}
		}
		for _, field := range expectedFindingFields {
			if _, present := f[field]; !present {
				return false, &AssessmentReportError{Reason: "missing finding field"}
			}
		}

		findingID := profileString(f["finding_id"])
		if !validAssessmentReportID(findingID) {
			return false, &AssessmentReportError{Reason: "finding identity"}
		}
		if seenIDs[findingID] {
			return false, &AssessmentReportError{Reason: "duplicate finding ID"}
		}
		seenIDs[findingID] = true

		if !validProfileRef(f["analysis_record_ref"], false) {
			return false, &AssessmentReportError{Reason: "finding analysis record reference"}
		}

		classification := profileString(f["classification"])
		switch classification {
		case "OBSERVED", "INFERRED", "UNKNOWN", "UNSUPPORTED", "RECOMMENDED":
		default:
			return false, &AssessmentReportError{Reason: "finding classification"}
		}

		basis := profileString(f["basis"])
		switch basis {
		case "STATIC_ONLY", "RUNTIME", "USER_ASSERTION", "MIXED", "MISSING_EVIDENCE":
		default:
			return false, &AssessmentReportError{Reason: "finding basis"}
		}

		category := profileString(f["category"])
		switch category {
		case "SYSTEM", "REQUIRED_CHANGE", "DEPENDENCY", "COMPATIBILITY", "RISK",
			"BLOCKER", "ALTERNATIVE", "UNKNOWN", "UNSUPPORTED", "MISSING_EVIDENCE",
			"CHARACTERIZATION", "ASSUMPTION", "RECOMMENDATION":
		default:
			return false, &AssessmentReportError{Reason: "finding category"}
		}

		if category == "BLOCKER" || category == "UNSUPPORTED" {
			hasBlocker = true
		}

		for _, field := range []string{"summary", "reason"} {
			var text string
			if string(f[field]) == "null" || json.Unmarshal(f[field], &text) != nil || len(text) < 1 {
				return false, &AssessmentReportError{Reason: "finding " + field}
			}
		}

		if err := validateAssessmentRefsArray(f["provenance_refs"], "provenance reference", 1); err != nil {
			return false, err
		}
		if err := validateAssessmentRefsArray(f["prerequisite_refs"], "prerequisite reference", 1); err != nil {
			return false, err
		}

		var limitations []json.RawMessage
		if err := json.Unmarshal(f["limitations"], &limitations); err != nil || limitations == nil {
			return false, &AssessmentReportError{Reason: "finding limitations"}
		}
		for _, rawLim := range limitations {
			var lim string
			if string(rawLim) == "null" || json.Unmarshal(rawLim, &lim) != nil || len(lim) < 1 {
				return false, &AssessmentReportError{Reason: "finding limitation item"}
			}
		}

		if basis == "STATIC_ONLY" && len(limitations) < 1 {
			return false, &AssessmentReportError{Reason: "static finding requires limitations"}
		}
		if basis == "MISSING_EVIDENCE" && classification == "OBSERVED" {
			return false, &AssessmentReportError{Reason: "missing evidence finding cannot be observed"}
		}

		var affectedIDs []json.RawMessage
		if err := json.Unmarshal(f["affected_ids"], &affectedIDs); err != nil || affectedIDs == nil {
			return false, &AssessmentReportError{Reason: "finding affected IDs"}
		}
		for _, rawAff := range affectedIDs {
			val := profileString(rawAff)
			if !validAssessmentReportID(val) {
				return false, &AssessmentReportError{Reason: "finding affected ID item"}
			}
		}

		if string(f["recommended_action"]) != "null" {
			var rec string
			if json.Unmarshal(f["recommended_action"], &rec) != nil || len(rec) < 1 {
				return false, &AssessmentReportError{Reason: "finding recommended action"}
			}
		}
	}

	return hasBlocker, nil
}

func validateAssessmentSequence(raw json.RawMessage) error {
	var steps []json.RawMessage
	if err := json.Unmarshal(raw, &steps); err != nil || steps == nil {
		return &AssessmentReportError{Reason: "suggested sequence array", Cause: err}
	}
	expectedStepFields := []string{"position", "finding_ids", "action", "prerequisite_finding_ids"}
	for _, rawStep := range steps {
		var step map[string]json.RawMessage
		if err := json.Unmarshal(rawStep, &step); err != nil || len(step) != len(expectedStepFields) {
			return &AssessmentReportError{Reason: "sequence step fields"}
		}
		for _, field := range expectedStepFields {
			if _, present := step[field]; !present {
				return &AssessmentReportError{Reason: "missing sequence step field"}
			}
		}

		var pos int64
		if string(step["position"]) == "null" || json.Unmarshal(step["position"], &pos) != nil || pos < 1 {
			return &AssessmentReportError{Reason: "sequence step position"}
		}

		var findingIDs []json.RawMessage
		if err := json.Unmarshal(step["finding_ids"], &findingIDs); err != nil || len(findingIDs) < 1 {
			return &AssessmentReportError{Reason: "sequence step finding IDs"}
		}
		for _, rawID := range findingIDs {
			val := profileString(rawID)
			if !validAssessmentReportID(val) {
				return &AssessmentReportError{Reason: "sequence step finding ID item"}
			}
		}

		var action string
		if string(step["action"]) == "null" || json.Unmarshal(step["action"], &action) != nil || len(action) < 1 {
			return &AssessmentReportError{Reason: "sequence step action"}
		}

		var prereqIDs []json.RawMessage
		if err := json.Unmarshal(step["prerequisite_finding_ids"], &prereqIDs); err != nil || prereqIDs == nil {
			return &AssessmentReportError{Reason: "sequence step prerequisite finding IDs"}
		}
		for _, rawPrereq := range prereqIDs {
			val := profileString(rawPrereq)
			if !validAssessmentReportID(val) {
				return &AssessmentReportError{Reason: "sequence step prerequisite finding ID item"}
			}
		}
	}
	return nil
}

func validateAssessmentEffort(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var effort map[string]json.RawMessage
	if err := json.Unmarshal(raw, &effort); err != nil || len(effort) != 8 {
		return &AssessmentReportError{Reason: "effort estimate fields", Cause: err}
	}
	for _, field := range []string{"scope_ref", "lower", "upper", "unit", "method", "evidence_refs", "assumptions", "limitations"} {
		if _, present := effort[field]; !present {
			return &AssessmentReportError{Reason: "missing effort estimate field"}
		}
	}
	if !validProfileRef(effort["scope_ref"], false) {
		return &AssessmentReportError{Reason: "effort scope reference"}
	}
	var lower, upper float64
	if string(effort["lower"]) == "null" || json.Unmarshal(effort["lower"], &lower) != nil || lower < 0 {
		return &AssessmentReportError{Reason: "effort lower bound"}
	}
	if string(effort["upper"]) == "null" || json.Unmarshal(effort["upper"], &upper) != nil || upper < 0 {
		return &AssessmentReportError{Reason: "effort upper bound"}
	}
	unit := profileString(effort["unit"])
	if unit != "PERSON_HOURS" && unit != "PERSON_DAYS" {
		return &AssessmentReportError{Reason: "effort unit"}
	}
	var method string
	if string(effort["method"]) == "null" || json.Unmarshal(effort["method"], &method) != nil || len(method) < 1 {
		return &AssessmentReportError{Reason: "effort method"}
	}
	if err := validateAssessmentRefsArray(effort["evidence_refs"], "effort evidence reference", 1); err != nil {
		return err
	}
	if err := validateAssessmentStringArray(effort["assumptions"], "effort assumptions", 1); err != nil {
		return err
	}
	return validateAssessmentStringArray(effort["limitations"], "effort limitations", 1)
}
