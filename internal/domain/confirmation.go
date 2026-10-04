package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// ConfirmationError describes an invalid confirmation receipt without echoing private contents.
type ConfirmationError struct {
	Reason string
	Cause  error
}

// Error identifies the failed confirmation obligation.
func (e *ConfirmationError) Error() string {
	return "invalid confirmation receipt: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ConfirmationError) Unwrap() error { return e.Cause }

// ValidateConfirmation validates an explicit human confirmation receipt.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, subject and configuration hash digests, reviewed payload
// reference, ISO 8601 / RFC 3339 creation timestamp, and strict gate-to-status
// lifecycle correspondence.
func ValidateConfirmation(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ConfirmationError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ConfirmationError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"confirmation_id",
		"project_id",
		"gate",
		"subject_hash",
		"configuration_hash",
		"confirmed_by",
		"channel",
		"action_id",
		"reviewed_payload_ref",
		"created_at",
		"status",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ConfirmationError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ConfirmationError{Reason: "object fields"}
	}

	if !validAssessmentReportID(profileString(record["confirmation_id"])) {
		return &ConfirmationError{Reason: "confirmation identity"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ConfirmationError{Reason: "project identity"}
	}

	if !validAssessmentReportID(profileString(record["action_id"])) {
		return &ConfirmationError{Reason: "action identity"}
	}

	subHash := profileString(record["subject_hash"])
	if !isManifestDigest(subHash) {
		return &ConfirmationError{Reason: "subject_hash"}
	}

	cfgHash := profileString(record["configuration_hash"])
	if !isManifestDigest(cfgHash) {
		return &ConfirmationError{Reason: "configuration_hash"}
	}

	if !validProfileRef(record["reviewed_payload_ref"], false) {
		return &ConfirmationError{Reason: "reviewed_payload_ref"}
	}

	confirmedBy := profileString(record["confirmed_by"])
	if strings.TrimSpace(confirmedBy) == "" {
		return &ConfirmationError{Reason: "confirmed_by"}
	}

	channel := profileString(record["channel"])
	if channel != "CLI" && channel != "MCP" {
		return &ConfirmationError{Reason: "invalid channel"}
	}

	createdAt := profileString(record["created_at"])
	if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
		return &ConfirmationError{Reason: "invalid created_at timestamp", Cause: err}
	}

	gate := profileString(record["gate"])
	status := profileString(record["status"])

	switch gate {
	case "GATE_1_SOURCE_DETECTION":
		if status != "SOURCE_CONFIRMED" {
			return &ConfirmationError{Reason: "gate and status mismatch"}
		}
	case "GATE_2_PLAN_CONFIRMED":
		if status != "PLAN_CONFIRMED" {
			return &ConfirmationError{Reason: "gate and status mismatch"}
		}
	case "GATE_3_CUTOVER":
		if status != "READY_CUTOVER" {
			return &ConfirmationError{Reason: "gate and status mismatch"}
		}
	default:
		return &ConfirmationError{Reason: "invalid gate"}
	}

	return nil
}

// ValidatePreparedConfirmation validates a prepared confirmation record under schema 1.0 (historical)
// or 2.0 (current).
// It verifies canonical RFC 8785 JSON within integer vocabulary, required fields, digest
// formats, payload references, and ensures the canonical subject hash matches the embedded subject.
func ValidatePreparedConfirmation(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ConfirmationError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ConfirmationError{Reason: "object fields", Cause: err}
	}

	versionRaw, hasVersion := record["schema_version"]
	if hasVersion {
		if profileString(versionRaw) != "2.0" {
			return &ConfirmationError{Reason: "unsupported schema version"}
		}
		return validatePreparedConfirmationV2(record)
	}
	return validatePreparedConfirmationV1(record)
}

func validatePreparedConfirmationV1(record map[string]json.RawMessage) error {
	expectedFields := []string{
		"project_id",
		"detection_id",
		"subject_hash",
		"merged_facts",
		"reviewed_payload_ref",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ConfirmationError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ConfirmationError{Reason: "object fields"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ConfirmationError{Reason: "project identity"}
	}

	if !validAssessmentReportID(profileString(record["detection_id"])) {
		return &ConfirmationError{Reason: "detection identity"}
	}

	if !isManifestDigest(profileString(record["subject_hash"])) {
		return &ConfirmationError{Reason: "subject_hash"}
	}

	if !validProfileRef(record["reviewed_payload_ref"], false) {
		return &ConfirmationError{Reason: "reviewed_payload_ref"}
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(record["merged_facts"], &merged); err != nil || merged == nil {
		return &ConfirmationError{Reason: "merged_facts object", Cause: err}
	}

	return nil
}

func validatePreparedConfirmationV2(record map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"project_id",
		"detection_id",
		"subject_hash",
		"measured_detection_ref",
		"assertions_ref",
		"reviewed_payload_ref",
		"subject",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ConfirmationError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ConfirmationError{Reason: "object fields"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ConfirmationError{Reason: "project identity"}
	}

	if !validAssessmentReportID(profileString(record["detection_id"])) {
		return &ConfirmationError{Reason: "detection identity"}
	}

	subHash := profileString(record["subject_hash"])
	if !isManifestDigest(subHash) {
		return &ConfirmationError{Reason: "subject_hash"}
	}

	for _, field := range []string{"measured_detection_ref", "assertions_ref", "reviewed_payload_ref"} {
		if !validProfileRef(record[field], false) {
			return &ConfirmationError{Reason: field}
		}
	}

	subjectRaw := record["subject"]
	if err := ValidateSourceApprovalSubjectShape(subjectRaw); err != nil {
		return &ConfirmationError{Reason: "subject shape", Cause: err}
	}

	calculatedHash, err := CanonicalHash(subjectRaw)
	if err != nil {
		return &ConfirmationError{Reason: "subject canonicalization", Cause: err}
	}
	if calculatedHash != subHash {
		return &ConfirmationError{Reason: "subject hash mismatch"}
	}

	return nil
}

// ValidateApprovalSubject validates an approval subject against specs/project/approval-subject.schema.json.
// It verifies schema_version 2.0, canonicalization rwt-cjson-1, project_id, and dispatches to the
// corresponding kind validator (SOURCE, PLAN, RELEASE) in accordance with CONFIRMATION-PROTOCOL.
func ValidateApprovalSubject(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ConfirmationError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ConfirmationError{Reason: "object fields", Cause: err}
	}

	for _, field := range []string{"schema_version", "canonicalization", "project_id", "kind"} {
		if val, present := record[field]; !present || string(val) == "null" {
			return &ConfirmationError{Reason: "missing required field: " + field}
		}
	}

	var kind string
	if err := json.Unmarshal(record["kind"], &kind); err != nil {
		return &ConfirmationError{Reason: "kind must be a string", Cause: err}
	}

	switch kind {
	case "SOURCE":
		return ValidateSourceApprovalSubjectShape(raw)
	case "PLAN":
		return ValidatePlanApprovalSubjectShape(raw)
	case "RELEASE":
		return ValidateReleaseApprovalSubjectShape(raw)
	default:
		return &ConfirmationError{Reason: "unsupported subject kind"}
	}
}
