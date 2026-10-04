package domain

import (
	"encoding/json"
	"strings"
)

// ExecutionVariantError identifies an invalid record without exposing its data.
type ExecutionVariantError struct {
	Reason string
	Cause  error
}

// Error describes the rejected execution-variant obligation.
func (e *ExecutionVariantError) Error() string { return "invalid execution variant: " + e.Reason }

// Unwrap preserves JSON failures for structured inspection.
func (e *ExecutionVariantError) Unwrap() error { return e.Cause }

// ValidateExecutionVariant checks the execution-variant 1.0 record contract.
// It does not resolve artifacts, compare captured bytes, validate filesystem
// paths, perform transforms or qualify credentials/review evidence. The schema
// defines path/reason as strings only; this function must not be used as path
// admission. A valid record neither replaces the private original nor grants
// execution authority or establishes equivalence with the original source.
func ValidateExecutionVariant(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ExecutionVariantError{Reason: "canonical JSON", Cause: err}
	}
	record, err := variantObject(canonical, "schema_version", "project_id", "original_manifest_ref", "original_content_hash", "variant_manifest_ref", "variant_content_hash", "transform_ref", "declared_changes", "credential_policy_ref", "limitations", "review_evidence_refs")
	if err != nil {
		return err
	}
	if profileString(record["schema_version"]) != "1.0" {
		return &ExecutionVariantError{Reason: "unsupported schema version"}
	}
	project := profileString(record["project_id"])
	if project == "" || strings.IndexFunc(project, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) >= 0 {
		return &ExecutionVariantError{Reason: "project identity"}
	}
	for _, field := range []string{"original_manifest_ref", "variant_manifest_ref", "transform_ref", "credential_policy_ref"} {
		if !validProfileRef(record[field], false) {
			return &ExecutionVariantError{Reason: "artifact reference"}
		}
	}
	for _, field := range []string{"original_content_hash", "variant_content_hash"} {
		if !isManifestDigest(profileString(record[field])) {
			return &ExecutionVariantError{Reason: "content hash"}
		}
	}
	var changes []json.RawMessage
	if err := json.Unmarshal(record["declared_changes"], &changes); err != nil || len(changes) == 0 {
		return &ExecutionVariantError{Reason: "declared changes required", Cause: err}
	}
	for _, rawChange := range changes {
		change, err := variantObject(rawChange, "path", "before_ref", "after_ref", "reason")
		if err != nil {
			return err
		}
		for _, field := range []string{"path", "reason"} {
			var text string
			if string(change[field]) == "null" || json.Unmarshal(change[field], &text) != nil {
				return &ExecutionVariantError{Reason: "change path or reason type"}
			}
		}
		if !validProfileRef(change["before_ref"], false) || !validProfileRef(change["after_ref"], false) {
			return &ExecutionVariantError{Reason: "change artifact reference"}
		}
	}
	if err := variantStringArray(record["limitations"], false); err != nil {
		return err
	}
	return variantStringArray(record["review_evidence_refs"], true)
}

func variantObject(raw []byte, fields ...string) (map[string]json.RawMessage, error) {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil || len(record) != len(fields) {
		return nil, &ExecutionVariantError{Reason: "object fields", Cause: err}
	}
	for _, field := range fields {
		if _, present := record[field]; !present {
			return nil, &ExecutionVariantError{Reason: "missing required field"}
		}
	}
	return record, nil
}

func variantStringArray(raw json.RawMessage, references bool) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || len(values) == 0 {
		return &ExecutionVariantError{Reason: "nonempty array required", Cause: err}
	}
	for _, value := range values {
		var text string
		if string(value) == "null" || json.Unmarshal(value, &text) != nil || references && !validProfileRef(value, false) {
			return &ExecutionVariantError{Reason: "array item type or reference"}
		}
	}
	return nil
}
