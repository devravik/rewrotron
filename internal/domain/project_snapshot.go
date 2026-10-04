package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

var allowedSnapshotGateStatuses = map[string]struct{}{
	"PENDING": {},
	"VALID":   {},
	"STALE":   {},
}

// ProjectSnapshotError describes an invalid project snapshot record without echoing private contents.
type ProjectSnapshotError struct {
	Reason string
	Cause  error
}

// Error identifies the failed project snapshot obligation.
func (e *ProjectSnapshotError) Error() string {
	return "invalid project snapshot: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ProjectSnapshotError) Unwrap() error { return e.Cause }

// ValidateProjectSnapshot validates a project snapshot record conforming to
// specs/project/snapshot.schema.json.
// It verifies required snapshot fields, project identity, lifecycle state, revision non-negativity,
// nullable configuration and plan digests, three gate states, four coverage dimensions with
// empty-denominator invariants, structured blockers, nullable latest_operation_id, and optional
// execution and assessment sub-projections.
func ValidateProjectSnapshot(raw []byte) error {
	record, err := parseProjectSnapshotJSON(raw)
	if err != nil {
		return err
	}

	requiredFields := []string{
		"project_id",
		"state",
		"revision",
		"configuration_hash",
		"plan_hash",
		"gate_states",
		"coverage",
		"blockers",
		"latest_operation_id",
	}

	allowedOptionalFields := map[string]struct{}{
		"execution":  {},
		"assessment": {},
	}

	for _, field := range requiredFields {
		if _, present := record[field]; !present {
			return &ProjectSnapshotError{Reason: "missing required field"}
		}
	}

	for field := range record {
		isReq := false
		for _, req := range requiredFields {
			if field == req {
				isReq = true
				break
			}
		}
		if isReq {
			continue
		}
		if _, allowed := allowedOptionalFields[field]; !allowed {
			return &ProjectSnapshotError{Reason: "object fields"}
		}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &ProjectSnapshotError{Reason: "project identity"}
	}

	var state string
	if string(record["state"]) == "null" || json.Unmarshal(record["state"], &state) != nil {
		return &ProjectSnapshotError{Reason: "invalid state"}
	}
	if _, ok := allowedOperationLifecycleStates[state]; !ok {
		return &ProjectSnapshotError{Reason: "invalid state"}
	}

	var revision int64
	if string(record["revision"]) == "null" || json.Unmarshal(record["revision"], &revision) != nil || revision < 0 {
		return &ProjectSnapshotError{Reason: "revision"}
	}

	if string(record["configuration_hash"]) != "null" {
		var cfgHash string
		if json.Unmarshal(record["configuration_hash"], &cfgHash) != nil || !isManifestDigest(cfgHash) {
			return &ProjectSnapshotError{Reason: "configuration_hash"}
		}
	}

	if string(record["plan_hash"]) != "null" {
		var planHash string
		if json.Unmarshal(record["plan_hash"], &planHash) != nil || !isManifestDigest(planHash) {
			return &ProjectSnapshotError{Reason: "plan_hash"}
		}
	}

	if err := validateSnapshotGateStates(record["gate_states"]); err != nil {
		return err
	}

	if err := validateSnapshotCoverage(record["coverage"]); err != nil {
		return err
	}

	if err := validateSnapshotBlockers(record["blockers"]); err != nil {
		return err
	}

	if string(record["latest_operation_id"]) != "null" {
		var opID string
		if json.Unmarshal(record["latest_operation_id"], &opID) != nil || !validAssessmentReportID(opID) {
			return &ProjectSnapshotError{Reason: "latest_operation_id"}
		}
	}

	if execRaw, present := record["execution"]; present && string(execRaw) != "null" {
		if err := ValidateExecutionState(execRaw); err != nil {
			return &ProjectSnapshotError{Reason: "invalid execution projection", Cause: err}
		}
	}

	if assessRaw, present := record["assessment"]; present && string(assessRaw) != "null" {
		if err := ValidateAssessmentView(assessRaw); err != nil {
			return &ProjectSnapshotError{Reason: "invalid assessment projection", Cause: err}
		}
	}

	return nil
}

func validateSnapshotGateStates(raw json.RawMessage) error {
	if string(raw) == "null" {
		return &ProjectSnapshotError{Reason: "gate_states must be an object"}
	}
	var gates map[string]json.RawMessage
	if err := json.Unmarshal(raw, &gates); err != nil || gates == nil {
		return &ProjectSnapshotError{Reason: "gate_states object", Cause: err}
	}

	expectedGates := []string{
		"GATE_1_SOURCE_DETECTION",
		"GATE_2_PLAN_CONFIRMED",
		"GATE_3_CUTOVER",
	}

	for _, req := range expectedGates {
		if _, present := gates[req]; !present {
			return &ProjectSnapshotError{Reason: "missing required gate"}
		}
	}
	if len(gates) != len(expectedGates) {
		return &ProjectSnapshotError{Reason: "gate_states object fields"}
	}

	for _, req := range expectedGates {
		var status string
		if string(gates[req]) == "null" || json.Unmarshal(gates[req], &status) != nil {
			return &ProjectSnapshotError{Reason: "gate status"}
		}
		if _, ok := allowedSnapshotGateStatuses[status]; !ok {
			return &ProjectSnapshotError{Reason: "gate status"}
		}
	}
	return nil
}

func validateSnapshotCoverage(raw json.RawMessage) error {
	if string(raw) == "null" {
		return &ProjectSnapshotError{Reason: "coverage must be an object"}
	}
	var cov map[string]json.RawMessage
	if err := json.Unmarshal(raw, &cov); err != nil || cov == nil {
		return &ProjectSnapshotError{Reason: "coverage object", Cause: err}
	}

	expectedDimensions := []string{
		"entrypoints",
		"cases",
		"units",
		"dimensions",
	}

	for _, req := range expectedDimensions {
		if _, present := cov[req]; !present {
			return &ProjectSnapshotError{Reason: "missing required coverage dimension"}
		}
	}
	if len(cov) != len(expectedDimensions) {
		return &ProjectSnapshotError{Reason: "coverage object fields"}
	}

	for _, dim := range expectedDimensions {
		if err := validateCoverageDimension(cov[dim], dim); err != nil {
			return err
		}
	}
	return nil
}

func validateCoverageDimension(raw json.RawMessage, dimName string) error {
	if string(raw) == "null" {
		return &ProjectSnapshotError{Reason: dimName + " coverage must be an object"}
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return &ProjectSnapshotError{Reason: dimName + " coverage object", Cause: err}
	}

	expectedFields := []string{"numerator", "denominator", "percent"}
	for _, req := range expectedFields {
		if _, present := obj[req]; !present {
			return &ProjectSnapshotError{Reason: "missing " + dimName + " coverage field"}
		}
	}
	if len(obj) != len(expectedFields) {
		return &ProjectSnapshotError{Reason: dimName + " coverage object fields"}
	}

	var num int64
	if string(obj["numerator"]) == "null" || json.Unmarshal(obj["numerator"], &num) != nil || num < 0 {
		return &ProjectSnapshotError{Reason: dimName + " numerator"}
	}

	var den int64
	if string(obj["denominator"]) == "null" || json.Unmarshal(obj["denominator"], &den) != nil || den < 0 {
		return &ProjectSnapshotError{Reason: dimName + " denominator"}
	}

	var percent *float64
	if string(obj["percent"]) != "null" {
		var p float64
		if json.Unmarshal(obj["percent"], &p) != nil || p < 0 || p > 100 {
			return &ProjectSnapshotError{Reason: dimName + " percent"}
		}
		percent = &p
	}

	// Schema allOf invariant: if denominator == 0, numerator must be 0 and percent must be null
	if den == 0 {
		if num != 0 {
			return &ProjectSnapshotError{Reason: dimName + " zero denominator requires zero numerator"}
		}
		if percent != nil {
			return &ProjectSnapshotError{Reason: dimName + " zero denominator requires null percent"}
		}
	}

	return nil
}

func validateSnapshotBlockers(raw json.RawMessage) error {
	var rawBlockers []json.RawMessage
	if err := json.Unmarshal(raw, &rawBlockers); err != nil {
		return &ProjectSnapshotError{Reason: "blockers array", Cause: err}
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
			return &ProjectSnapshotError{Reason: "blocker object", Cause: err}
		}

		for _, field := range expectedBlockerFields {
			if _, present := bRecord[field]; !present {
				return &ProjectSnapshotError{Reason: "missing required blocker field"}
			}
		}
		if len(bRecord) != len(expectedBlockerFields) {
			return &ProjectSnapshotError{Reason: "blocker object fields"}
		}

		var findingID string
		if string(bRecord["finding_id"]) == "null" || json.Unmarshal(bRecord["finding_id"], &findingID) != nil || !validAssessmentReportID(findingID) {
			return &ProjectSnapshotError{Reason: "invalid blocker finding_id"}
		}
		if _, exists := seenFindings[findingID]; exists {
			return &ProjectSnapshotError{Reason: "duplicate blocker finding_id"}
		}
		seenFindings[findingID] = struct{}{}

		var code string
		if string(bRecord["code"]) == "null" || json.Unmarshal(bRecord["code"], &code) != nil || strings.TrimSpace(code) == "" {
			return &ProjectSnapshotError{Reason: "invalid blocker code"}
		}

		var severity string
		if string(bRecord["severity"]) == "null" || json.Unmarshal(bRecord["severity"], &severity) != nil {
			return &ProjectSnapshotError{Reason: "invalid blocker severity"}
		}
		if _, ok := allowedBlockerSeverities[severity]; !ok {
			return &ProjectSnapshotError{Reason: "invalid blocker severity"}
		}

		if err := validateSnapshotBlockerIDList(bRecord["affected_ids"], "affected_ids"); err != nil {
			return err
		}

		if err := validateSnapshotBlockerIDList(bRecord["evidence_ids"], "evidence_ids"); err != nil {
			return err
		}

		var message string
		if string(bRecord["message"]) == "null" || json.Unmarshal(bRecord["message"], &message) != nil || strings.TrimSpace(message) == "" {
			return &ProjectSnapshotError{Reason: "invalid blocker message"}
		}
	}
	return nil
}

func validateSnapshotBlockerIDList(raw json.RawMessage, fieldName string) error {
	var ids []json.RawMessage
	if err := json.Unmarshal(raw, &ids); err != nil {
		return &ProjectSnapshotError{Reason: "blocker " + fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, item := range ids {
		var id string
		if string(item) == "null" || json.Unmarshal(item, &id) != nil || !validAssessmentReportID(id) {
			return &ProjectSnapshotError{Reason: "invalid blocker " + fieldName + " identity"}
		}
		if _, exists := seen[id]; exists {
			return &ProjectSnapshotError{Reason: "duplicate blocker " + fieldName + " identity"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func parseProjectSnapshotJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ProjectSnapshotError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ProjectSnapshotError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ProjectSnapshotError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ProjectSnapshotError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ProjectSnapshotError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ProjectSnapshotError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ProjectSnapshotError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ProjectSnapshotError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ProjectSnapshotError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ProjectSnapshotError{Reason: "trailing input"}
	}
	return record, nil
}
