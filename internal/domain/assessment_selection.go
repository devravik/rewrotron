package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// AssessmentSelectionError describes an invalid assessment-selection record without echoing private contents.
type AssessmentSelectionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed assessment-selection obligation.
func (e *AssessmentSelectionError) Error() string {
	return "invalid assessment selection: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *AssessmentSelectionError) Unwrap() error { return e.Cause }

// ValidateAssessmentSelection validates an assessment-selection 1.0 record.
// It verifies canonical JSON, required properties, identifier grammar,
// SHA-256 references, action enums, conditional action requirements,
// human CLI action metadata with RFC 3339 timestamps, and the strict invariant
// that implementation_authorized must be false.
//
// This check does not verify artifact existence, resolve assessment reports,
// authorize planning, or grant execution authority.
func ValidateAssessmentSelection(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &AssessmentSelectionError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &AssessmentSelectionError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version", "selection_id", "project_id", "assessment_ref",
		"action", "selected_scope_ref", "target_intent_ref", "finding_ids",
		"decision_refs", "requested_evidence", "rationale",
		"implementation_authorized", "human_action",
	}

	if len(record) != len(expectedFields) {
		return &AssessmentSelectionError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &AssessmentSelectionError{Reason: "missing required field"}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &AssessmentSelectionError{Reason: "unsupported schema version"}
	}

	for _, field := range []string{"selection_id", "project_id"} {
		val := profileString(record[field])
		if !validAssessmentSelectionID(val) {
			return &AssessmentSelectionError{Reason: field + " identity"}
		}
	}

	if !validProfileRef(record["assessment_ref"], false) {
		return &AssessmentSelectionError{Reason: "assessment reference"}
	}

	action := profileString(record["action"])
	switch action {
	case "STOP", "REQUEST_EVIDENCE", "CHANGE_TARGET", "CHANGE_SCOPE",
		"SELECT_WORK", "ACCEPT_RECOMMENDATIONS", "REJECT_RECOMMENDATIONS", "CONTINUE_PLANNING":
	default:
		return &AssessmentSelectionError{Reason: "unsupported action"}
	}

	if !validProfileRef(record["selected_scope_ref"], true) {
		return &AssessmentSelectionError{Reason: "selected scope reference"}
	}
	if !validProfileRef(record["target_intent_ref"], true) {
		return &AssessmentSelectionError{Reason: "target intent reference"}
	}

	var findingIDs []json.RawMessage
	if err := json.Unmarshal(record["finding_ids"], &findingIDs); err != nil || findingIDs == nil {
		return &AssessmentSelectionError{Reason: "finding IDs", Cause: err}
	}
	seenFindings := make(map[string]bool, len(findingIDs))
	for _, rawFinding := range findingIDs {
		val := profileString(rawFinding)
		if !validAssessmentSelectionID(val) {
			return &AssessmentSelectionError{Reason: "finding ID item"}
		}
		if seenFindings[val] {
			return &AssessmentSelectionError{Reason: "duplicate finding ID"}
		}
		seenFindings[val] = true
	}

	var decisionRefs []json.RawMessage
	if err := json.Unmarshal(record["decision_refs"], &decisionRefs); err != nil || decisionRefs == nil {
		return &AssessmentSelectionError{Reason: "decision references", Cause: err}
	}
	seenDecisions := make(map[string]bool, len(decisionRefs))
	for _, rawRef := range decisionRefs {
		if string(rawRef) == "null" || !validProfileRef(rawRef, false) {
			return &AssessmentSelectionError{Reason: "decision reference item"}
		}
		val := profileString(rawRef)
		if seenDecisions[val] {
			return &AssessmentSelectionError{Reason: "duplicate decision reference"}
		}
		seenDecisions[val] = true
	}

	var requestedEvidence []json.RawMessage
	if err := json.Unmarshal(record["requested_evidence"], &requestedEvidence); err != nil || requestedEvidence == nil {
		return &AssessmentSelectionError{Reason: "requested evidence", Cause: err}
	}
	seenEvidence := make(map[string]bool, len(requestedEvidence))
	for _, rawEvidence := range requestedEvidence {
		var item string
		if string(rawEvidence) == "null" || json.Unmarshal(rawEvidence, &item) != nil || len(item) < 1 {
			return &AssessmentSelectionError{Reason: "requested evidence item"}
		}
		if seenEvidence[item] {
			return &AssessmentSelectionError{Reason: "duplicate requested evidence"}
		}
		seenEvidence[item] = true
	}

	var rationale string
	if string(record["rationale"]) == "null" || json.Unmarshal(record["rationale"], &rationale) != nil || len(rationale) < 1 {
		return &AssessmentSelectionError{Reason: "rationale"}
	}

	// Invariant: implementation_authorized must be false.
	var implAuth bool
	if string(record["implementation_authorized"]) == "null" || json.Unmarshal(record["implementation_authorized"], &implAuth) != nil || implAuth {
		return &AssessmentSelectionError{Reason: "implementation cannot be authorized"}
	}

	// Human action block.
	var humanAction map[string]json.RawMessage
	if err := json.Unmarshal(record["human_action"], &humanAction); err != nil || len(humanAction) != 4 {
		return &AssessmentSelectionError{Reason: "human action fields", Cause: err}
	}
	for _, hField := range []string{"actor_id", "channel", "action_id", "recorded_at"} {
		if _, present := humanAction[hField]; !present {
			return &AssessmentSelectionError{Reason: "missing human action field"}
		}
	}

	var actorID string
	if string(humanAction["actor_id"]) == "null" || json.Unmarshal(humanAction["actor_id"], &actorID) != nil || len(actorID) < 1 {
		return &AssessmentSelectionError{Reason: "human action actor identity"}
	}

	if profileString(humanAction["channel"]) != "CLI" {
		return &AssessmentSelectionError{Reason: "human action channel"}
	}

	if !validAssessmentSelectionID(profileString(humanAction["action_id"])) {
		return &AssessmentSelectionError{Reason: "human action action identity"}
	}

	var recordedAt string
	if string(humanAction["recorded_at"]) == "null" || json.Unmarshal(humanAction["recorded_at"], &recordedAt) != nil {
		return &AssessmentSelectionError{Reason: "human action recorded_at timestamp"}
	}
	if _, err := time.Parse(time.RFC3339, recordedAt); err != nil {
		return &AssessmentSelectionError{Reason: "human action recorded_at timestamp", Cause: err}
	}

	// Conditional constraints per action.
	switch action {
	case "CONTINUE_PLANNING":
		if string(record["selected_scope_ref"]) == "null" {
			return &AssessmentSelectionError{Reason: "selected scope reference required for continuation"}
		}
		if string(record["target_intent_ref"]) == "null" {
			return &AssessmentSelectionError{Reason: "target intent reference required for continuation"}
		}
	case "SELECT_WORK", "CHANGE_SCOPE":
		if string(record["selected_scope_ref"]) == "null" {
			return &AssessmentSelectionError{Reason: "selected scope reference required"}
		}
	case "CHANGE_TARGET":
		if string(record["target_intent_ref"]) == "null" {
			return &AssessmentSelectionError{Reason: "target intent reference required"}
		}
	case "ACCEPT_RECOMMENDATIONS", "REJECT_RECOMMENDATIONS":
		if len(findingIDs) < 1 {
			return &AssessmentSelectionError{Reason: "finding IDs required for recommendations action"}
		}
	case "REQUEST_EVIDENCE":
		if len(requestedEvidence) < 1 {
			return &AssessmentSelectionError{Reason: "requested evidence required for evidence request"}
		}
	}

	return nil
}

func validAssessmentSelectionID(value string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) < 0
}
