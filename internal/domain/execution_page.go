package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

var allowedExecutionChangeKinds = map[string]struct{}{
	"DISCOVERED":  {},
	"COMPLETED":   {},
	"RESOLVED":    {},
	"BLOCKED":     {},
	"LESSON":      {},
	"INVALIDATED": {},
	"REOPENED":    {},
}

// ExecutionPageError describes an invalid execution timeline page record without echoing private contents.
type ExecutionPageError struct {
	Reason string
	Cause  error
}

// Error identifies the failed execution page obligation.
func (e *ExecutionPageError) Error() string {
	return "invalid execution page: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ExecutionPageError) Unwrap() error { return e.Cause }

// ValidateExecutionPage validates a bounded execution timeline and changes page record conforming
// to specs/execution/page.schema.json (schema_version 1.0 or 2.0).
// It verifies required fields, project identity, non-negative sequence bounds (after_sequence <= through_sequence),
// embedded current execution state conforming to ValidateExecutionState, bounded events array (at most 500 items)
// where each item conforms to ValidateExecutionEvent, structured changes array with allowed change kinds,
// and in schema version 2.0, state_semantics ("CURRENT") and non-negative state_sequence watermark.
func ValidateExecutionPage(raw []byte) error {
	record, err := parseExecutionPageJSON(raw)
	if err != nil {
		return err
	}

	var version string
	if string(record["schema_version"]) == "null" || json.Unmarshal(record["schema_version"], &version) != nil {
		return &ExecutionPageError{Reason: "schema_version"}
	}
	if version != "1.0" && version != "2.0" {
		return &ExecutionPageError{Reason: "unsupported schema_version"}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"after_sequence",
		"through_sequence",
		"next_sequence",
		"has_more",
		"state",
		"events",
		"changes",
	}
	if version == "2.0" {
		expectedFields = append(expectedFields, "state_semantics", "state_sequence")
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ExecutionPageError{Reason: "missing required field: " + field}
		}
	}
	if len(record) != len(expectedFields) {
		return &ExecutionPageError{Reason: "unexpected additional properties"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &ExecutionPageError{Reason: "invalid project_id"}
	}

	var afterSeq int64
	if string(record["after_sequence"]) == "null" || json.Unmarshal(record["after_sequence"], &afterSeq) != nil || afterSeq < 0 {
		return &ExecutionPageError{Reason: "after_sequence must be integer >= 0"}
	}
	if strings.ContainsAny(string(record["after_sequence"]), ".eE") {
		return &ExecutionPageError{Reason: "after_sequence must be integer"}
	}

	var throughSeq int64
	if string(record["through_sequence"]) == "null" || json.Unmarshal(record["through_sequence"], &throughSeq) != nil || throughSeq < 0 {
		return &ExecutionPageError{Reason: "through_sequence must be integer >= 0"}
	}
	if strings.ContainsAny(string(record["through_sequence"]), ".eE") {
		return &ExecutionPageError{Reason: "through_sequence must be integer"}
	}

	if afterSeq > throughSeq {
		return &ExecutionPageError{Reason: "after_sequence must be <= through_sequence"}
	}

	var nextSeq int64
	if string(record["next_sequence"]) == "null" || json.Unmarshal(record["next_sequence"], &nextSeq) != nil || nextSeq < 0 {
		return &ExecutionPageError{Reason: "next_sequence must be integer >= 0"}
	}
	if strings.ContainsAny(string(record["next_sequence"]), ".eE") {
		return &ExecutionPageError{Reason: "next_sequence must be integer"}
	}

	var hasMore bool
	if string(record["has_more"]) == "null" || json.Unmarshal(record["has_more"], &hasMore) != nil {
		return &ExecutionPageError{Reason: "has_more must be a boolean"}
	}

	if err := ValidateExecutionState(record["state"]); err != nil {
		return &ExecutionPageError{Reason: "invalid embedded execution state", Cause: err}
	}

	var rawEvents []json.RawMessage
	if string(record["events"]) == "null" || json.Unmarshal(record["events"], &rawEvents) != nil {
		return &ExecutionPageError{Reason: "events must be an array"}
	}
	if len(rawEvents) > 500 {
		return &ExecutionPageError{Reason: "events exceeds maxItems 500"}
	}
	for _, rawEvent := range rawEvents {
		if err := ValidateExecutionEvent(rawEvent); err != nil {
			return &ExecutionPageError{Reason: "invalid event in page", Cause: err}
		}
	}

	if err := validateExecutionPageChanges(record["changes"]); err != nil {
		return err
	}

	if version == "2.0" {
		var semantics string
		if string(record["state_semantics"]) == "null" || json.Unmarshal(record["state_semantics"], &semantics) != nil {
			return &ExecutionPageError{Reason: "state_semantics"}
		}
		if semantics != "CURRENT" {
			return &ExecutionPageError{Reason: "state_semantics must be CURRENT"}
		}

		var stateSeq int64
		if string(record["state_sequence"]) == "null" || json.Unmarshal(record["state_sequence"], &stateSeq) != nil || stateSeq < 0 {
			return &ExecutionPageError{Reason: "state_sequence must be integer >= 0"}
		}
		if strings.ContainsAny(string(record["state_sequence"]), ".eE") {
			return &ExecutionPageError{Reason: "state_sequence must be integer"}
		}
	}

	return nil
}

func validateExecutionPageChanges(raw json.RawMessage) error {
	var changes []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &changes) != nil {
		return &ExecutionPageError{Reason: "changes must be an array"}
	}

	expectedFields := []string{
		"kind",
		"entity_type",
		"entity_id",
		"summary",
		"event_id",
		"sequence",
		"evidence_refs",
	}

	for _, rawChange := range changes {
		var chObj map[string]json.RawMessage
		if json.Unmarshal(rawChange, &chObj) != nil || chObj == nil {
			return &ExecutionPageError{Reason: "change item must be an object"}
		}
		for _, req := range expectedFields {
			if _, present := chObj[req]; !present {
				return &ExecutionPageError{Reason: "missing required change field: " + req}
			}
		}
		if len(chObj) != len(expectedFields) {
			return &ExecutionPageError{Reason: "unexpected change fields"}
		}

		var kind string
		if string(chObj["kind"]) == "null" || json.Unmarshal(chObj["kind"], &kind) != nil {
			return &ExecutionPageError{Reason: "invalid change kind"}
		}
		if _, ok := allowedExecutionChangeKinds[kind]; !ok {
			return &ExecutionPageError{Reason: "invalid change kind enum: " + kind}
		}

		var entityType string
		if string(chObj["entity_type"]) == "null" || json.Unmarshal(chObj["entity_type"], &entityType) != nil || len(entityType) < 1 {
			return &ExecutionPageError{Reason: "invalid change entity_type"}
		}

		var entityID string
		if string(chObj["entity_id"]) == "null" || json.Unmarshal(chObj["entity_id"], &entityID) != nil || !validAssessmentReportID(entityID) {
			return &ExecutionPageError{Reason: "invalid change entity_id"}
		}

		var summary string
		if string(chObj["summary"]) == "null" || json.Unmarshal(chObj["summary"], &summary) != nil || len(summary) < 1 {
			return &ExecutionPageError{Reason: "invalid change summary"}
		}

		var eventID string
		if string(chObj["event_id"]) == "null" || json.Unmarshal(chObj["event_id"], &eventID) != nil || !validAssessmentReportID(eventID) {
			return &ExecutionPageError{Reason: "invalid change event_id"}
		}

		var sequence int64
		if string(chObj["sequence"]) == "null" || json.Unmarshal(chObj["sequence"], &sequence) != nil || sequence < 1 {
			return &ExecutionPageError{Reason: "change sequence must be integer >= 1"}
		}
		if strings.ContainsAny(string(chObj["sequence"]), ".eE") {
			return &ExecutionPageError{Reason: "change sequence must be integer"}
		}

		var evRefs []json.RawMessage
		if string(chObj["evidence_refs"]) == "null" || json.Unmarshal(chObj["evidence_refs"], &evRefs) != nil {
			return &ExecutionPageError{Reason: "change evidence_refs must be an array"}
		}
		for _, ref := range evRefs {
			if !validProfileRef(ref, false) {
				return &ExecutionPageError{Reason: "invalid change evidence_refs item"}
			}
		}
	}
	return nil
}

func parseExecutionPageJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ExecutionPageError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ExecutionPageError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ExecutionPageError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ExecutionPageError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ExecutionPageError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ExecutionPageError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ExecutionPageError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ExecutionPageError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ExecutionPageError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ExecutionPageError{Reason: "trailing input"}
	}
	return record, nil
}
