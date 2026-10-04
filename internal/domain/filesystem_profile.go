package domain

import (
	"encoding/json"
	"strings"
)

// FilesystemProfileError reports an invalid profile record without its contents.
type FilesystemProfileError struct {
	Reason string
	Cause  error
}

// Error describes the rejected record obligation.
func (e *FilesystemProfileError) Error() string { return "invalid filesystem profile: " + e.Reason }

// Unwrap preserves JSON failures for structured inspection.
func (e *FilesystemProfileError) Unwrap() error { return e.Cause }

// ValidateFilesystemProfile checks the filesystem-profile 1.0 record contract.
// It validates reference syntax, not artifact existence, evidence quality or
// coherent capture. A valid record may contain UNSUPPORTED properties and does
// not establish behavioral identity, boundary closure or execution admission.
// The caller must separately resolve evidence and controls before execution.
func ValidateFilesystemProfile(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &FilesystemProfileError{Reason: "canonical JSON", Cause: err}
	}
	profile, err := profileObject(canonical, "schema_version", "project_id", "content_manifest_ref", "properties", "control_refs", "limitations")
	if err != nil {
		return err
	}
	if profileString(profile["schema_version"]) != "1.0" {
		return &FilesystemProfileError{Reason: "unsupported schema version"}
	}
	project := profileString(profile["project_id"])
	if project == "" || strings.IndexFunc(project, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) >= 0 {
		return &FilesystemProfileError{Reason: "project identity"}
	}
	if !validProfileRef(profile["content_manifest_ref"], false) {
		return &FilesystemProfileError{Reason: "content manifest reference"}
	}
	var properties []json.RawMessage
	if err := json.Unmarshal(profile["properties"], &properties); err != nil || len(properties) != 10 {
		return &FilesystemProfileError{Reason: "ten properties required", Cause: err}
	}
	seen := make(map[string]bool, 10)
	for _, rawProperty := range properties {
		property, err := profileObject(rawProperty, "property", "treatment", "captured_state_ref", "mechanism_ref", "evidence_refs")
		if err != nil {
			return err
		}
		name := profileString(property["property"])
		switch name {
		case "PERMISSIONS", "EXECUTABLE", "MTIME", "OWNERSHIP", "HARDLINKS", "SYMLINKS", "CASE_SENSITIVITY", "PATH_RESOLUTION", "WORKING_DIRECTORY", "DIRECTORY_ORDER":
		default:
			return &FilesystemProfileError{Reason: "unsupported property"}
		}
		if seen[name] {
			return &FilesystemProfileError{Reason: "duplicate property"}
		}
		seen[name] = true
		treatment := profileString(property["treatment"])
		var nullable bool
		switch treatment {
		case "REQUIRED", "VIRTUALIZED":
		case "IRRELEVANT", "UNSUPPORTED":
			nullable = true
		default:
			return &FilesystemProfileError{Reason: "unsupported treatment"}
		}
		if !validProfileRef(property["captured_state_ref"], nullable) || !validProfileRef(property["mechanism_ref"], nullable) {
			return &FilesystemProfileError{Reason: "capture or mechanism reference"}
		}
		if err := profileStringArray(property["evidence_refs"], true, 1); err != nil {
			return err
		}
	}
	if err := profileStringArray(profile["control_refs"], true, 0); err != nil {
		return err
	}
	return profileStringArray(profile["limitations"], false, 0)
}

func profileObject(raw []byte, fields ...string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || len(object) != len(fields) {
		return nil, &FilesystemProfileError{Reason: "object fields", Cause: err}
	}
	for _, field := range fields {
		if _, present := object[field]; !present {
			return nil, &FilesystemProfileError{Reason: "missing required field"}
		}
	}
	return object, nil
}

func profileString(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return value
}

func validProfileRef(raw json.RawMessage, nullable bool) bool {
	if string(raw) == "null" {
		return nullable
	}
	value := profileString(raw)
	return strings.HasPrefix(value, "sha256:") && isManifestDigest(strings.TrimPrefix(value, "sha256:"))
}

func profileStringArray(raw json.RawMessage, references bool, minimum int) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil || len(values) < minimum {
		return &FilesystemProfileError{Reason: "array cardinality or type", Cause: err}
	}
	for _, value := range values {
		var text string
		if string(value) == "null" || json.Unmarshal(value, &text) != nil || references && !validProfileRef(value, false) {
			return &FilesystemProfileError{Reason: "array item type or reference"}
		}
	}
	return nil
}
