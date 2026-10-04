package domain

import (
	"encoding/json"
	"strings"
)

var allowedMigrationProgressStatuses = map[string]struct{}{
	"PLANNED":    {},
	"GENERATING": {},
	"BUILDING":   {},
	"VERIFYING":  {},
	"VERIFIED":   {},
	"BLOCKED":    {},
}

// MigrationProgressError describes an invalid migration progress record without echoing private contents.
type MigrationProgressError struct {
	Reason string
	Cause  error
}

// Error identifies the failed migration progress obligation.
func (e *MigrationProgressError) Error() string {
	return "invalid migration progress: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *MigrationProgressError) Unwrap() error { return e.Cause }

// ValidateMigrationProgress validates a migration progress record conforming to
// specs/migration/progress.schema.json.
// It verifies valid UTF-8 JSON, schema version 2.0, valid project_id, plan_hash,
// decision_or_model_ref, entity_id, status enum, non-negative integer through_sequence,
// nullable sha256 target_tree_ref and acceptance_report_ref, and no unexpected fields.
func ValidateMigrationProgress(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &MigrationProgressError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &MigrationProgressError{Reason: "object required", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"plan_hash",
		"decision_or_model_ref",
		"entity_id",
		"status",
		"through_sequence",
		"target_tree_ref",
		"acceptance_report_ref",
	}

	for _, field := range expectedFields {
		v, present := record[field]
		if !present {
			return &MigrationProgressError{Reason: "missing required field: " + field}
		}
		// target_tree_ref and acceptance_report_ref are nullable; other fields cannot be null
		if string(v) == "null" && field != "target_tree_ref" && field != "acceptance_report_ref" {
			return &MigrationProgressError{Reason: "null not allowed for field: " + field}
		}
	}

	if len(record) != len(expectedFields) {
		return &MigrationProgressError{Reason: "unexpected fields"}
	}

	var version string
	if err := json.Unmarshal(record["schema_version"], &version); err != nil || version != "2.0" {
		return &MigrationProgressError{Reason: "schema_version must be '2.0'"}
	}

	var projectID string
	if err := json.Unmarshal(record["project_id"], &projectID); err != nil || !validAssessmentReportID(projectID) {
		return &MigrationProgressError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}

	var planHash string
	if err := json.Unmarshal(record["plan_hash"], &planHash); err != nil || !isManifestDigest(planHash) {
		return &MigrationProgressError{Reason: "plan_hash must match pattern ^[a-f0-9]{64}$"}
	}

	if !validProfileRef(record["decision_or_model_ref"], false) {
		return &MigrationProgressError{Reason: "decision_or_model_ref must be a valid sha256 reference"}
	}

	var entityID string
	if err := json.Unmarshal(record["entity_id"], &entityID); err != nil || !validAssessmentReportID(entityID) {
		return &MigrationProgressError{Reason: "entity_id must match pattern ^[a-z0-9_-]+$"}
	}

	var status string
	if err := json.Unmarshal(record["status"], &status); err != nil {
		return &MigrationProgressError{Reason: "status must be a string", Cause: err}
	}
	if _, ok := allowedMigrationProgressStatuses[status]; !ok {
		return &MigrationProgressError{Reason: "invalid status enum"}
	}

	var seq int64
	if err := json.Unmarshal(record["through_sequence"], &seq); err != nil || seq < 0 {
		return &MigrationProgressError{Reason: "through_sequence must be a non-negative integer"}
	}
	// Check that through_sequence was not serialized as float (e.g. 1.5)
	if strings.Contains(string(record["through_sequence"]), ".") {
		return &MigrationProgressError{Reason: "through_sequence must be an integer"}
	}

	if !validProfileRef(record["target_tree_ref"], true) {
		return &MigrationProgressError{Reason: "target_tree_ref must be null or a valid sha256 reference"}
	}

	if !validProfileRef(record["acceptance_report_ref"], true) {
		return &MigrationProgressError{Reason: "acceptance_report_ref must be null or a valid sha256 reference"}
	}

	return nil
}
