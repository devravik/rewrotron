package domain

import (
	"encoding/json"
	"strings"
)

// ExecutionAdmissionError describes an invalid record without exposing its data.
type ExecutionAdmissionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed admission-record obligation.
func (e *ExecutionAdmissionError) Error() string { return "invalid execution admission: " + e.Reason }

// Unwrap preserves decoding failures for inspection.
func (e *ExecutionAdmissionError) Unwrap() error { return e.Cause }

// ValidateExecutionAdmission checks the execution-admission 1.0 record contract.
// It does not inspect credentials, resolve artifacts, establish revocation or
// validate the referenced variant. Schema-valid outcomes remain untrusted until
// the guarded service qualifies their evidence and approval closure. This
// function neither authorizes execution nor changes the recorded outcome.
func ValidateExecutionAdmission(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ExecutionAdmissionError{Reason: "canonical JSON", Cause: err}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || len(record) != 10 {
		return &ExecutionAdmissionError{Reason: "object fields", Cause: err}
	}
	for _, field := range []string{"schema_version", "project_id", "original_manifest_ref", "outcome", "variant_ref", "credential_status", "policy_ref", "evidence_refs", "limitations", "variant_credential_status"} {
		if _, present := record[field]; !present {
			return &ExecutionAdmissionError{Reason: "missing required field"}
		}
	}
	if profileString(record["schema_version"]) != "1.0" {
		return &ExecutionAdmissionError{Reason: "unsupported schema version"}
	}
	project := profileString(record["project_id"])
	if project == "" || strings.IndexFunc(project, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) >= 0 {
		return &ExecutionAdmissionError{Reason: "project identity"}
	}
	if !validProfileRef(record["original_manifest_ref"], false) || !validProfileRef(record["policy_ref"], false) {
		return &ExecutionAdmissionError{Reason: "manifest or policy reference"}
	}
	status := profileString(record["credential_status"])
	switch status {
	case "SYNTHETIC_OR_ABSENT", "REVOKED", "LIVE_OR_UNKNOWN":
	default:
		return &ExecutionAdmissionError{Reason: "original credential status"}
	}
	switch profileString(record["outcome"]) {
	case "ADMITTED_VARIANT":
		if !validProfileRef(record["variant_ref"], false) {
			return &ExecutionAdmissionError{Reason: "variant reference required"}
		}
		switch profileString(record["variant_credential_status"]) {
		case "SYNTHETIC_OR_ABSENT", "REVOKED":
		default:
			return &ExecutionAdmissionError{Reason: "unsafe variant credential status"}
		}
	case "ADMITTED_ORIGINAL", "REFUSED", "STATIC_ONLY":
		if string(record["variant_ref"]) != "null" || string(record["variant_credential_status"]) != "null" {
			return &ExecutionAdmissionError{Reason: "variant fields must be null"}
		}
		if profileString(record["outcome"]) == "ADMITTED_ORIGINAL" && status == "LIVE_OR_UNKNOWN" {
			return &ExecutionAdmissionError{Reason: "unsafe original credential status"}
		}
	default:
		return &ExecutionAdmissionError{Reason: "unsupported outcome"}
	}
	if err := admissionArray(record["evidence_refs"], true, 1); err != nil {
		return err
	}
	return admissionArray(record["limitations"], false, 0)
}

func admissionArray(raw json.RawMessage, references bool, minimum int) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil || len(values) < minimum {
		return &ExecutionAdmissionError{Reason: "array cardinality or type", Cause: err}
	}
	for _, value := range values {
		var text string
		if string(value) == "null" || json.Unmarshal(value, &text) != nil || references && !validProfileRef(value, false) {
			return &ExecutionAdmissionError{Reason: "array item type or reference"}
		}
	}
	return nil
}
