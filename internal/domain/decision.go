package domain

import (
	"encoding/json"
	"strings"
)

var allowedDecisionActions = map[string]struct{}{
	"ACCEPT":             {},
	"REJECT":             {},
	"SELECT_STRATEGY":    {},
	"SUPPLY_ALTERNATIVE": {},
	"REQUEST_EVIDENCE":   {},
	"APPROVE_EXCEPTION":  {},
}

var allowedDecisionStrategies = map[string]struct{}{
	"DIRECT_UPGRADE": {},
	"ADAPTER":        {},
	"REPLACEMENT":    {},
	"REIMPLEMENT":    {},
	"RETAIN":         {},
}

// DecisionError describes an invalid decision proposal record without echoing private contents.
type DecisionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed decision proposal obligation.
func (e *DecisionError) Error() string {
	return "invalid decision proposal: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *DecisionError) Unwrap() error { return e.Cause }

// ValidateDecision validates a 2.0 decision proposal record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, decision and project identities, typed actions, nullable strategy enum,
// nullable implementation and exception scope references, non-empty rationale,
// unique evidence references, and strict PROPOSED status.
func ValidateDecision(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &DecisionError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &DecisionError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"decision_id",
		"project_id",
		"action",
		"subject_ref",
		"strategy",
		"implementation_ref",
		"exception_scope_ref",
		"rationale",
		"evidence_refs",
		"status",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &DecisionError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &DecisionError{Reason: "object fields"}
	}

	if profileString(record["schema_version"]) != "2.0" {
		return &DecisionError{Reason: "unsupported schema version"}
	}

	if !validAssessmentReportID(profileString(record["decision_id"])) {
		return &DecisionError{Reason: "decision identity"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &DecisionError{Reason: "project identity"}
	}

	action := profileString(record["action"])
	if _, ok := allowedDecisionActions[action]; !ok {
		return &DecisionError{Reason: "invalid action"}
	}

	if !validProfileRef(record["subject_ref"], false) {
		return &DecisionError{Reason: "subject_ref"}
	}

	stratRaw := record["strategy"]
	if string(stratRaw) != "null" {
		strat := profileString(stratRaw)
		if _, ok := allowedDecisionStrategies[strat]; !ok {
			return &DecisionError{Reason: "invalid strategy"}
		}
	}

	if !validProfileRef(record["implementation_ref"], true) {
		return &DecisionError{Reason: "implementation_ref"}
	}

	if !validProfileRef(record["exception_scope_ref"], true) {
		return &DecisionError{Reason: "exception_scope_ref"}
	}

	rationale := profileString(record["rationale"])
	if strings.TrimSpace(rationale) == "" {
		return &DecisionError{Reason: "rationale"}
	}

	if err := validateDecisionEvidenceRefs(record["evidence_refs"]); err != nil {
		return err
	}

	if profileString(record["status"]) != "PROPOSED" {
		return &DecisionError{Reason: "status must be PROPOSED"}
	}

	return nil
}

func validateDecisionEvidenceRefs(raw json.RawMessage) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil {
		return &DecisionError{Reason: "evidence_refs array", Cause: err}
	}
	seen := make(map[string]struct{}, len(refs))
	for _, item := range refs {
		if !validProfileRef(item, false) {
			return &DecisionError{Reason: "invalid evidence_refs reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return &DecisionError{Reason: "duplicate evidence_refs reference"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}
