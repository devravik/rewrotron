package domain

import (
	"encoding/json"
	"strings"
)

// WorkScopeError describes an invalid work-scope record without echoing private contents.
type WorkScopeError struct {
	Reason string
	Cause  error
}

// Error identifies the failed work-scope obligation.
func (e *WorkScopeError) Error() string {
	return "invalid work scope: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *WorkScopeError) Unwrap() error { return e.Cause }

// ValidateWorkScope validates a 1.0 work-scope record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, scope and project identity patterns, source snapshot reference, unique
// entrypoints and source paths, safe source path boundaries (forbidding root slashes
// and directory traversal escapes), workflow/constraint/business requirement references,
// optional operational profile reference, and unique limitations.
func ValidateWorkScope(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &WorkScopeError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &WorkScopeError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"scope_id",
		"project_id",
		"source_snapshot_ref",
		"included_entrypoint_ids",
		"included_source_paths",
		"critical_workflow_refs",
		"constraint_refs",
		"business_requirement_refs",
		"operational_profile_ref",
		"excluded_scope_refs",
		"limitations",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &WorkScopeError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &WorkScopeError{Reason: "object fields"}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &WorkScopeError{Reason: "unsupported schema version"}
	}

	if !validAssessmentReportID(profileString(record["scope_id"])) {
		return &WorkScopeError{Reason: "scope identity"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &WorkScopeError{Reason: "project identity"}
	}

	if !validProfileRef(record["source_snapshot_ref"], false) {
		return &WorkScopeError{Reason: "source_snapshot_ref"}
	}

	if err := validateWorkScopeIDs(record["included_entrypoint_ids"], "included_entrypoint_ids"); err != nil {
		return err
	}

	if err := validateWorkScopePaths(record["included_source_paths"]); err != nil {
		return err
	}

	for _, refListField := range []string{"critical_workflow_refs", "constraint_refs", "business_requirement_refs", "excluded_scope_refs"} {
		if err := validateWorkScopeRefList(record[refListField], refListField); err != nil {
			return err
		}
	}

	if !validProfileRef(record["operational_profile_ref"], true) {
		return &WorkScopeError{Reason: "operational_profile_ref"}
	}

	if err := validateWorkScopeLimitations(record["limitations"]); err != nil {
		return err
	}

	return nil
}

func validWorkScopeSourcePath(p string) bool {
	if len(p) == 0 {
		return false
	}
	if strings.HasPrefix(p, "/") {
		return false
	}
	if strings.Contains(p, "\\") || strings.Contains(p, "\x00") {
		return false
	}
	parts := strings.Split(p, "/")
	for _, part := range parts {
		if part == ".." {
			return false
		}
	}
	return true
}

func validateWorkScopeIDs(raw json.RawMessage, fieldName string) error {
	var ids []json.RawMessage
	if err := json.Unmarshal(raw, &ids); err != nil {
		return &WorkScopeError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, item := range ids {
		id := profileString(item)
		if !validAssessmentReportID(id) {
			return &WorkScopeError{Reason: "invalid " + fieldName + " identity"}
		}
		if _, exists := seen[id]; exists {
			return &WorkScopeError{Reason: "duplicate " + fieldName + " identity"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateWorkScopePaths(raw json.RawMessage) error {
	var paths []json.RawMessage
	if err := json.Unmarshal(raw, &paths); err != nil || len(paths) < 1 {
		return &WorkScopeError{Reason: "included_source_paths array", Cause: err}
	}
	seen := make(map[string]struct{}, len(paths))
	for _, item := range paths {
		var p string
		if string(item) == "null" || json.Unmarshal(item, &p) != nil || !validWorkScopeSourcePath(p) {
			return &WorkScopeError{Reason: "invalid source path"}
		}
		if _, exists := seen[p]; exists {
			return &WorkScopeError{Reason: "duplicate source path"}
		}
		seen[p] = struct{}{}
	}
	return nil
}

func validateWorkScopeRefList(raw json.RawMessage, fieldName string) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil {
		return &WorkScopeError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(refs))
	for _, item := range refs {
		if !validProfileRef(item, false) {
			return &WorkScopeError{Reason: "invalid " + fieldName + " reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return &WorkScopeError{Reason: "duplicate " + fieldName + " reference"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}

func validateWorkScopeLimitations(raw json.RawMessage) error {
	var lims []json.RawMessage
	if err := json.Unmarshal(raw, &lims); err != nil {
		return &WorkScopeError{Reason: "limitations array", Cause: err}
	}
	seen := make(map[string]struct{}, len(lims))
	for _, item := range lims {
		var s string
		if string(item) == "null" || json.Unmarshal(item, &s) != nil || strings.TrimSpace(s) == "" {
			return &WorkScopeError{Reason: "invalid limitation entry"}
		}
		if _, exists := seen[s]; exists {
			return &WorkScopeError{Reason: "duplicate limitation entry"}
		}
		seen[s] = struct{}{}
	}
	return nil
}
