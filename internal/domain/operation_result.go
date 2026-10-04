package domain

import (
	"encoding/json"
	"strings"
)

var allowedOperationStatuses = map[string]struct{}{
	"COMPLETED": {},
	"FAILED":    {},
	"CANCELLED": {},
	"PREPARED":  {},
	"RUNNING":   {},
}

var allowedOperationLifecycleStates = map[string]struct{}{
	"INITIALIZED":      {},
	"DETECTED":         {},
	"SOURCE_CONFIRMED": {},
	"ANALYZED":         {},
	"MODELED":          {},
	"PLAN_SYNTHESIZED": {},
	"PLAN_CONFIRMED":   {},
	"MIGRATING":        {},
	"VERIFIED":         {},
	"BENCHMARKED":      {},
	"READY_CUTOVER":    {},
}

var allowedBlockerSeverities = map[string]struct{}{
	"CRITICAL": {},
	"HIGH":     {},
	"MEDIUM":   {},
	"LOW":      {},
}

// OperationResultError describes an invalid operation result record without echoing private contents.
type OperationResultError struct {
	Reason string
	Cause  error
}

// Error identifies the failed operation result obligation.
func (e *OperationResultError) Error() string {
	return "invalid operation result: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *OperationResultError) Unwrap() error { return e.Cause }

// ValidateOperationResult validates an operations result record against specs/operations/result.schema.json.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, operation identifier patterns, allowed status and lifecycle state enums,
// valid artifact references and entity identities, structured blocker records with
// severity and evidence associations, and valid summary object.
//
// RUNNING/PREPARED is admission/checkpoint, not completed work. Captured immutable
// evidence references remain project-qualified.
func ValidateOperationResult(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &OperationResultError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &OperationResultError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"operation_id",
		"status",
		"state",
		"artifact_refs",
		"entity_ids",
		"blockers",
		"summary",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &OperationResultError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &OperationResultError{Reason: "object fields"}
	}

	var opID string
	if string(record["operation_id"]) == "null" || json.Unmarshal(record["operation_id"], &opID) != nil || !validAssessmentReportID(opID) {
		return &OperationResultError{Reason: "operation identity"}
	}

	var status string
	if string(record["status"]) == "null" || json.Unmarshal(record["status"], &status) != nil {
		return &OperationResultError{Reason: "invalid status"}
	}
	if _, ok := allowedOperationStatuses[status]; !ok {
		return &OperationResultError{Reason: "invalid status"}
	}

	stateRaw := record["state"]
	if string(stateRaw) != "null" {
		var state string
		if json.Unmarshal(stateRaw, &state) != nil {
			return &OperationResultError{Reason: "invalid state"}
		}
		if _, ok := allowedOperationLifecycleStates[state]; !ok {
			return &OperationResultError{Reason: "invalid state"}
		}
	}

	if err := validateOperationArtifactRefs(record["artifact_refs"]); err != nil {
		return err
	}

	if err := validateOperationEntityIDs(record["entity_ids"]); err != nil {
		return err
	}

	if err := validateOperationBlockers(record["blockers"]); err != nil {
		return err
	}

	if err := validateOperationSummary(record["summary"]); err != nil {
		return err
	}

	return nil
}

func validateOperationArtifactRefs(raw json.RawMessage) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil {
		return &OperationResultError{Reason: "artifact_refs array", Cause: err}
	}
	seen := make(map[string]struct{}, len(refs))
	for _, item := range refs {
		if string(item) == "null" || !validProfileRef(item, false) {
			return &OperationResultError{Reason: "invalid artifact_refs reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return &OperationResultError{Reason: "duplicate artifact_refs reference"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}

func validateOperationEntityIDs(raw json.RawMessage) error {
	var ids []json.RawMessage
	if err := json.Unmarshal(raw, &ids); err != nil {
		return &OperationResultError{Reason: "entity_ids array", Cause: err}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, item := range ids {
		var id string
		if string(item) == "null" || json.Unmarshal(item, &id) != nil || !validAssessmentReportID(id) {
			return &OperationResultError{Reason: "invalid entity_ids identity"}
		}
		if _, exists := seen[id]; exists {
			return &OperationResultError{Reason: "duplicate entity_ids identity"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateOperationBlockers(raw json.RawMessage) error {
	var rawBlockers []json.RawMessage
	if err := json.Unmarshal(raw, &rawBlockers); err != nil {
		return &OperationResultError{Reason: "blockers array", Cause: err}
	}

	expectedBlockerFields := []string{
		"finding_id",
		"code",
		"severity",
		"affected_ids",
		"evidence_ids",
		"message",
	}

	seenFindings := make(map[string]struct{}, len(rawBlockers))
	for _, rawBlocker := range rawBlockers {
		var bRecord map[string]json.RawMessage
		if err := json.Unmarshal(rawBlocker, &bRecord); err != nil || bRecord == nil {
			return &OperationResultError{Reason: "blocker object", Cause: err}
		}

		for _, field := range expectedBlockerFields {
			if _, present := bRecord[field]; !present {
				return &OperationResultError{Reason: "missing required blocker field"}
			}
		}
		if len(bRecord) != len(expectedBlockerFields) {
			return &OperationResultError{Reason: "blocker object fields"}
		}

		var findingID string
		if string(bRecord["finding_id"]) == "null" || json.Unmarshal(bRecord["finding_id"], &findingID) != nil || !validAssessmentReportID(findingID) {
			return &OperationResultError{Reason: "invalid blocker finding_id"}
		}
		if _, exists := seenFindings[findingID]; exists {
			return &OperationResultError{Reason: "duplicate blocker finding_id"}
		}
		seenFindings[findingID] = struct{}{}

		var code string
		if string(bRecord["code"]) == "null" || json.Unmarshal(bRecord["code"], &code) != nil || strings.TrimSpace(code) == "" {
			return &OperationResultError{Reason: "invalid blocker code"}
		}

		var severity string
		if string(bRecord["severity"]) == "null" || json.Unmarshal(bRecord["severity"], &severity) != nil {
			return &OperationResultError{Reason: "invalid blocker severity"}
		}
		if _, ok := allowedBlockerSeverities[severity]; !ok {
			return &OperationResultError{Reason: "invalid blocker severity"}
		}

		if err := validateBlockerIDList(bRecord["affected_ids"], "affected_ids"); err != nil {
			return err
		}

		if err := validateBlockerIDList(bRecord["evidence_ids"], "evidence_ids"); err != nil {
			return err
		}

		var message string
		if string(bRecord["message"]) == "null" || json.Unmarshal(bRecord["message"], &message) != nil || strings.TrimSpace(message) == "" {
			return &OperationResultError{Reason: "invalid blocker message"}
		}
	}
	return nil
}

func validateBlockerIDList(raw json.RawMessage, fieldName string) error {
	var ids []json.RawMessage
	if err := json.Unmarshal(raw, &ids); err != nil {
		return &OperationResultError{Reason: "blocker " + fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, item := range ids {
		var id string
		if string(item) == "null" || json.Unmarshal(item, &id) != nil || !validAssessmentReportID(id) {
			return &OperationResultError{Reason: "invalid blocker " + fieldName + " identity"}
		}
		if _, exists := seen[id]; exists {
			return &OperationResultError{Reason: "duplicate blocker " + fieldName + " identity"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateOperationSummary(raw json.RawMessage) error {
	if string(raw) == "null" {
		return &OperationResultError{Reason: "summary must be an object"}
	}
	var summary map[string]json.RawMessage
	if err := json.Unmarshal(raw, &summary); err != nil || summary == nil {
		return &OperationResultError{Reason: "summary must be an object", Cause: err}
	}
	return nil
}
