package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

// ReleaseApprovalError describes an invalid RELEASE approval subject without echoing private contents.
type ReleaseApprovalError struct {
	Reason string
	Cause  error
}

// Error identifies the failed RELEASE approval subject obligation.
func (e *ReleaseApprovalError) Error() string {
	return "invalid release approval subject: " + e.Reason
}

// Unwrap preserves decoding/canonicalization failures for structured inspection.
func (e *ReleaseApprovalError) Unwrap() error { return e.Cause }

// ValidateReleaseApprovalSubjectShape checks the RELEASE branch of subject 2.0 conforming to
// specs/project/approval-subject.schema.json#/$defs/release.
// It verifies canonical RFC 8785 JSON, required subject identity, valid project identity,
// configuration_hash, plan_hash, and all required artifact references.
// Shape validation neither creates an approval nor authorizes execution.
func ValidateReleaseApprovalSubjectShape(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ReleaseApprovalError{Reason: "canonical JSON", Cause: err}
	}

	record, err := parseReleaseSubjectObject(canonical)
	if err != nil {
		return err
	}

	requiredFields := []string{
		"schema_version",
		"canonicalization",
		"project_id",
		"kind",
		"configuration_hash",
		"plan_hash",
		"target_tree_ref",
		"whole_project_report_ref",
		"benchmark_ref",
		"runbook_ref",
		"rollback_rehearsal_ref",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &ReleaseApprovalError{Reason: "missing required field: " + f}
		}
	}

	if len(record) != len(requiredFields) {
		return &ReleaseApprovalError{Reason: "object fields"}
	}

	if profileString(record["schema_version"]) != "2.0" ||
		profileString(record["canonicalization"]) != "rwt-cjson-1" ||
		profileString(record["kind"]) != "RELEASE" {
		return &ReleaseApprovalError{Reason: "unsupported subject identity"}
	}

	if !validApprovalID(profileString(record["project_id"])) {
		return &ReleaseApprovalError{Reason: "project identity"}
	}

	cfgHash := profileString(record["configuration_hash"])
	if !isManifestDigest(cfgHash) {
		return &ReleaseApprovalError{Reason: "configuration_hash"}
	}

	planHash := profileString(record["plan_hash"])
	if !isManifestDigest(planHash) {
		return &ReleaseApprovalError{Reason: "plan_hash"}
	}

	refFields := []string{
		"target_tree_ref",
		"whole_project_report_ref",
		"benchmark_ref",
		"runbook_ref",
		"rollback_rehearsal_ref",
	}
	for _, rf := range refFields {
		if !validProfileRef(record[rf], false) {
			return &ReleaseApprovalError{Reason: "artifact reference: " + rf}
		}
	}

	return nil
}

func parseReleaseSubjectObject(raw []byte) (map[string]json.RawMessage, error) {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil || record == nil {
		return nil, &ReleaseApprovalError{Reason: "object fields", Cause: err}
	}
	return record, nil
}

// MigrationReleaseError describes an invalid migration release envelope without echoing private contents.
type MigrationReleaseError struct {
	Reason string
	Cause  error
}

// Error identifies the failed migration release envelope obligation.
func (e *MigrationReleaseError) Error() string {
	return "invalid migration release: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *MigrationReleaseError) Unwrap() error { return e.Cause }

// ValidateMigrationRelease validates a migration release record conforming to
// specs/migration/release.schema.json.
// It verifies version 1.0 (historical read-only) or version 2.0 (current output),
// all required hashes and artifact references, and in version 2.0 verifies that the
// embedded subject conforms to ValidateReleaseApprovalSubjectShape, matches the envelope project_id,
// and matches the release_hash digest.
func ValidateMigrationRelease(raw []byte) error {
	record, err := parseMigrationReleaseJSON(raw)
	if err != nil {
		return err
	}

	versionRaw, present := record["schema_version"]
	if !present || string(versionRaw) == "null" {
		return &MigrationReleaseError{Reason: "missing schema_version"}
	}
	var version string
	if err := json.Unmarshal(versionRaw, &version); err != nil {
		return &MigrationReleaseError{Reason: "schema_version must be a string", Cause: err}
	}

	switch version {
	case "1.0":
		return validateMigrationReleaseV1(record)
	case "2.0":
		return validateMigrationReleaseV2(record)
	default:
		return &MigrationReleaseError{Reason: "unsupported schema_version: " + version}
	}
}

func validateMigrationReleaseV1(record map[string]json.RawMessage) error {
	requiredFields := []string{
		"schema_version",
		"project_id",
		"release_hash",
		"configuration_hash",
		"plan_hash",
		"target_tree_hash",
		"dependency_lock_hashes",
		"verification_report_ref",
		"benchmark_report_ref",
		"cutover_runbook_ref",
		"rollback_runbook_ref",
		"session_strategy_ref",
		"scope_exclusions",
		"omission_report_ref",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &MigrationReleaseError{Reason: "missing required field: " + f}
		}
	}

	if len(record) != len(requiredFields) {
		return &MigrationReleaseError{Reason: "object fields"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &MigrationReleaseError{Reason: "invalid project_id"}
	}

	for _, hashField := range []string{"release_hash", "configuration_hash", "plan_hash", "target_tree_hash"} {
		var h string
		if string(record[hashField]) == "null" || json.Unmarshal(record[hashField], &h) != nil || !isManifestDigest(h) {
			return &MigrationReleaseError{Reason: "invalid " + hashField}
		}
	}

	var depLockHashes []json.RawMessage
	if string(record["dependency_lock_hashes"]) == "null" || json.Unmarshal(record["dependency_lock_hashes"], &depLockHashes) != nil {
		return &MigrationReleaseError{Reason: "dependency_lock_hashes must be an array"}
	}
	for _, item := range depLockHashes {
		var h string
		if string(item) == "null" || json.Unmarshal(item, &h) != nil || !isManifestDigest(h) {
			return &MigrationReleaseError{Reason: "invalid dependency_lock_hashes item"}
		}
	}

	for _, refField := range []string{
		"verification_report_ref",
		"benchmark_report_ref",
		"cutover_runbook_ref",
		"rollback_runbook_ref",
		"session_strategy_ref",
		"omission_report_ref",
	} {
		if !validProfileRef(record[refField], false) {
			return &MigrationReleaseError{Reason: "invalid " + refField}
		}
	}

	var scopeExclusions []json.RawMessage
	if string(record["scope_exclusions"]) == "null" || json.Unmarshal(record["scope_exclusions"], &scopeExclusions) != nil {
		return &MigrationReleaseError{Reason: "scope_exclusions must be an array"}
	}
	for _, item := range scopeExclusions {
		var obj map[string]json.RawMessage
		if json.Unmarshal(item, &obj) != nil || obj == nil {
			return &MigrationReleaseError{Reason: "scope_exclusions item must be an object"}
		}
	}

	return nil
}

func validateMigrationReleaseV2(record map[string]json.RawMessage) error {
	requiredFields := []string{
		"schema_version",
		"project_id",
		"release_hash",
		"subject",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &MigrationReleaseError{Reason: "missing required field: " + f}
		}
	}

	if len(record) != len(requiredFields) {
		return &MigrationReleaseError{Reason: "object fields"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &MigrationReleaseError{Reason: "invalid project_id"}
	}

	var releaseHash string
	if string(record["release_hash"]) == "null" || json.Unmarshal(record["release_hash"], &releaseHash) != nil || !isManifestDigest(releaseHash) {
		return &MigrationReleaseError{Reason: "invalid release_hash"}
	}

	subjectRaw := record["subject"]
	if string(subjectRaw) == "null" {
		return &MigrationReleaseError{Reason: "subject cannot be null"}
	}

	if err := ValidateReleaseApprovalSubjectShape(subjectRaw); err != nil {
		return &MigrationReleaseError{Reason: "invalid subject", Cause: err}
	}

	var subjectObj map[string]json.RawMessage
	if err := json.Unmarshal(subjectRaw, &subjectObj); err == nil && subjectObj != nil {
		subjectProjectID := profileString(subjectObj["project_id"])
		if subjectProjectID != projectID {
			return &MigrationReleaseError{Reason: "subject project_id does not match envelope project_id"}
		}
	}

	expectedHash, err := CanonicalHash(subjectRaw)
	if err != nil {
		return &MigrationReleaseError{Reason: "failed to compute canonical subject hash", Cause: err}
	}
	if expectedHash != releaseHash {
		return &MigrationReleaseError{Reason: "release_hash does not match canonical subject digest"}
	}

	return nil
}

func parseMigrationReleaseJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &MigrationReleaseError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &MigrationReleaseError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &MigrationReleaseError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &MigrationReleaseError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &MigrationReleaseError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &MigrationReleaseError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &MigrationReleaseError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &MigrationReleaseError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &MigrationReleaseError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &MigrationReleaseError{Reason: "trailing input"}
	}
	return record, nil
}
