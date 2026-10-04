package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

var allowedInventoryDispositions = map[string]struct{}{
	"MIGRATED":           {},
	"DEAD_OMISSION":      {},
	"REVIEWED_EXCLUSION": {},
	"UNRESOLVED":         {},
}

// MigrationReportError describes an invalid migration report record without echoing private contents.
type MigrationReportError struct {
	Reason string
	Cause  error
}

// Error identifies the failed migration report obligation.
func (e *MigrationReportError) Error() string {
	return "invalid migration report: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *MigrationReportError) Unwrap() error { return e.Cause }

// ValidateMigrationReport validates a migration report record conforming to
// specs/reports/migration-report.schema.json.
// It verifies schema version 1.0, required hashes (configuration, plan, target_tree),
// embedded project snapshot conforming to ValidateProjectSnapshot, inventory dispositions
// with allowed disposition enums and evidence references, report reference arrays, and
// optional release reference.
func ValidateMigrationReport(raw []byte) error {
	record, err := parseMigrationReportJSON(raw)
	if err != nil {
		return err
	}

	requiredFields := []string{
		"schema_version",
		"project_id",
		"configuration_hash",
		"plan_hash",
		"target_tree_hash",
		"snapshot",
		"inventory_dispositions",
		"verification_report_refs",
		"benchmark_report_refs",
		"omission_report_ref",
		"intentional_changes",
		"limitations",
		"required_decisions",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &MigrationReportError{Reason: "missing required field: " + f}
		}
	}

	allowedOptionalFields := map[string]struct{}{
		"release_ref": {},
	}

	reqMap := make(map[string]struct{}, len(requiredFields))
	for _, f := range requiredFields {
		reqMap[f] = struct{}{}
	}

	for k := range record {
		if _, req := reqMap[k]; req {
			continue
		}
		if _, opt := allowedOptionalFields[k]; opt {
			continue
		}
		return &MigrationReportError{Reason: "unexpected field: " + k}
	}

	var version string
	if string(record["schema_version"]) == "null" || json.Unmarshal(record["schema_version"], &version) != nil || version != "1.0" {
		return &MigrationReportError{Reason: "schema_version must be 1.0"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &MigrationReportError{Reason: "invalid project_id"}
	}

	var cfgHash string
	if string(record["configuration_hash"]) == "null" || json.Unmarshal(record["configuration_hash"], &cfgHash) != nil || !isManifestDigest(cfgHash) {
		return &MigrationReportError{Reason: "invalid configuration_hash"}
	}

	var planHash string
	if string(record["plan_hash"]) == "null" || json.Unmarshal(record["plan_hash"], &planHash) != nil || !isManifestDigest(planHash) {
		return &MigrationReportError{Reason: "invalid plan_hash"}
	}

	var targetTreeHash string
	if string(record["target_tree_hash"]) == "null" || json.Unmarshal(record["target_tree_hash"], &targetTreeHash) != nil || !isManifestDigest(targetTreeHash) {
		return &MigrationReportError{Reason: "invalid target_tree_hash"}
	}

	if err := ValidateProjectSnapshot(record["snapshot"]); err != nil {
		return &MigrationReportError{Reason: "invalid embedded snapshot", Cause: err}
	}

	if err := validateInventoryDispositions(record["inventory_dispositions"]); err != nil {
		return err
	}

	if err := validateReportRefArray(record["verification_report_refs"], "verification_report_refs"); err != nil {
		return err
	}

	if err := validateReportRefArray(record["benchmark_report_refs"], "benchmark_report_refs"); err != nil {
		return err
	}

	if string(record["omission_report_ref"]) == "null" || !validProfileRef(record["omission_report_ref"], false) {
		return &MigrationReportError{Reason: "invalid omission_report_ref"}
	}

	var intChanges []json.RawMessage
	if string(record["intentional_changes"]) == "null" || json.Unmarshal(record["intentional_changes"], &intChanges) != nil {
		return &MigrationReportError{Reason: "intentional_changes must be an array"}
	}
	for _, ic := range intChanges {
		var icObj map[string]json.RawMessage
		if json.Unmarshal(ic, &icObj) != nil || icObj == nil {
			return &MigrationReportError{Reason: "intentional_changes item must be an object"}
		}
	}

	var limitations []string
	if string(record["limitations"]) == "null" || json.Unmarshal(record["limitations"], &limitations) != nil {
		return &MigrationReportError{Reason: "limitations must be an array"}
	}

	var reqDecisions []string
	if string(record["required_decisions"]) == "null" || json.Unmarshal(record["required_decisions"], &reqDecisions) != nil {
		return &MigrationReportError{Reason: "required_decisions must be an array"}
	}

	if val, present := record["release_ref"]; present && string(val) != "null" {
		if !validProfileRef(val, false) {
			return &MigrationReportError{Reason: "invalid release_ref"}
		}
	}

	return nil
}

func validateInventoryDispositions(raw json.RawMessage) error {
	var list []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &list) != nil {
		return &MigrationReportError{Reason: "inventory_dispositions must be an array"}
	}

	requiredFields := []string{"entrypoint_id", "disposition", "evidence_refs"}

	for _, item := range list {
		var obj map[string]json.RawMessage
		if json.Unmarshal(item, &obj) != nil || obj == nil {
			return &MigrationReportError{Reason: "inventory_dispositions item must be an object"}
		}

		for _, req := range requiredFields {
			if _, present := obj[req]; !present || string(obj[req]) == "null" {
				return &MigrationReportError{Reason: "missing required inventory_dispositions field: " + req}
			}
		}

		for k := range obj {
			if k != "entrypoint_id" && k != "disposition" && k != "evidence_refs" && k != "unit_id" {
				return &MigrationReportError{Reason: "unexpected field in inventory_dispositions: " + k}
			}
		}

		var epID string
		if json.Unmarshal(obj["entrypoint_id"], &epID) != nil || !validAssessmentReportID(epID) {
			return &MigrationReportError{Reason: "invalid inventory disposition entrypoint_id"}
		}

		var disp string
		if json.Unmarshal(obj["disposition"], &disp) != nil {
			return &MigrationReportError{Reason: "invalid inventory disposition"}
		}
		if _, ok := allowedInventoryDispositions[disp]; !ok {
			return &MigrationReportError{Reason: "invalid inventory disposition enum: " + disp}
		}

		if uid, present := obj["unit_id"]; present && string(uid) != "null" {
			var u string
			if json.Unmarshal(uid, &u) != nil || !validAssessmentReportID(u) {
				return &MigrationReportError{Reason: "invalid inventory disposition unit_id"}
			}
		}

		var evRefs []json.RawMessage
		if json.Unmarshal(obj["evidence_refs"], &evRefs) != nil {
			return &MigrationReportError{Reason: "inventory disposition evidence_refs must be an array"}
		}
		for _, ref := range evRefs {
			if !validProfileRef(ref, false) {
				return &MigrationReportError{Reason: "invalid inventory disposition evidence_refs item"}
			}
		}
	}
	return nil
}

func validateReportRefArray(raw json.RawMessage, fieldName string) error {
	var list []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &list) != nil {
		return &MigrationReportError{Reason: fieldName + " must be an array"}
	}
	for _, ref := range list {
		if !validProfileRef(ref, false) {
			return &MigrationReportError{Reason: "invalid " + fieldName + " item"}
		}
	}
	return nil
}

func parseMigrationReportJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &MigrationReportError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &MigrationReportError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &MigrationReportError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &MigrationReportError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &MigrationReportError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &MigrationReportError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &MigrationReportError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &MigrationReportError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &MigrationReportError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &MigrationReportError{Reason: "trailing input"}
	}
	return record, nil
}
