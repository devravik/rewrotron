package domain

import (
	"encoding/json"
	"strings"
)

// ExecutionManifestError describes an invalid execution manifest without exposing private data.
type ExecutionManifestError struct {
	Reason string
	Cause  error
}

// Error identifies the failed execution-manifest obligation.
func (e *ExecutionManifestError) Error() string {
	return "invalid execution manifest: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ExecutionManifestError) Unwrap() error { return e.Cause }

// ValidateExecutionManifest validates an execution-manifest 1.0 or 1.1 record.
// Version 1.0 pins package-free engine, build, dependency, schema, and qualification closure.
// Version 1.1 additionally pins package descriptor and content inventory pairs.
// It checks canonical JSON, field presence, schema versions, identifiers,
// SHA-256 references, canonicalization profile, and unique array members.
// It does not verify the existence of referenced artifacts on disk, qualify engine binaries,
// establish trust, or grant execution authority.
func ValidateExecutionManifest(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ExecutionManifestError{Reason: "canonical JSON", Cause: err}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ExecutionManifestError{Reason: "object fields", Cause: err}
	}

	if len(record) != 11 && len(record) != 12 {
		return &ExecutionManifestError{Reason: "object fields"}
	}

	versionRaw, present := record["schema_version"]
	if !present {
		return &ExecutionManifestError{Reason: "missing required field"}
	}
	version := profileString(versionRaw)
	if version != "1.0" && version != "1.1" {
		return &ExecutionManifestError{Reason: "unsupported schema version"}
	}

	var expectedFields []string
	if version == "1.0" {
		if len(record) != 11 {
			return &ExecutionManifestError{Reason: "object fields"}
		}
		expectedFields = []string{
			"schema_version", "manifest_id", "project_id", "engine_version",
			"engine_binary_ref", "build_environment_ref", "dependency_lock_ref",
			"schema_refs", "canonicalization_profile", "capability_ids", "qualification_refs",
		}
	} else {
		if len(record) != 12 {
			return &ExecutionManifestError{Reason: "object fields"}
		}
		expectedFields = []string{
			"schema_version", "manifest_id", "project_id", "engine_version",
			"engine_binary_ref", "build_environment_ref", "dependency_lock_ref",
			"schema_refs", "canonicalization_profile", "capability_ids", "qualification_refs",
			"package_refs",
		}
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ExecutionManifestError{Reason: "missing required field"}
		}
	}

	for _, field := range []string{"manifest_id", "project_id"} {
		val := profileString(record[field])
		if !validExecutionManifestID(val) {
			return &ExecutionManifestError{Reason: field + " identity"}
		}
	}

	var engineVersion string
	if string(record["engine_version"]) == "null" || json.Unmarshal(record["engine_version"], &engineVersion) != nil || len(engineVersion) < 1 {
		return &ExecutionManifestError{Reason: "engine version"}
	}

	for _, field := range []string{"engine_binary_ref", "build_environment_ref", "dependency_lock_ref"} {
		if !validProfileRef(record[field], false) {
			return &ExecutionManifestError{Reason: field + " reference"}
		}
	}

	if profileString(record["canonicalization_profile"]) != "rwt-cjson-1" {
		return &ExecutionManifestError{Reason: "canonicalization profile"}
	}

	if err := executionManifestRefs(record["schema_refs"], "schema references"); err != nil {
		return err
	}
	if err := executionManifestCapabilityIDs(record["capability_ids"]); err != nil {
		return err
	}
	if err := executionManifestRefs(record["qualification_refs"], "qualification references"); err != nil {
		return err
	}

	if version == "1.1" {
		if err := executionManifestPackageRefs(record["package_refs"]); err != nil {
			return err
		}
	}

	return nil
}

func validExecutionManifestID(value string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) < 0
}

func executionManifestRefs(raw json.RawMessage, label string) error {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) < 1 {
		return &ExecutionManifestError{Reason: label, Cause: err}
	}
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if string(item) == "null" || !validProfileRef(item, false) {
			return &ExecutionManifestError{Reason: label + " item"}
		}
		val := profileString(item)
		if seen[val] {
			return &ExecutionManifestError{Reason: "duplicate " + label}
		}
		seen[val] = true
	}
	return nil
}

func executionManifestCapabilityIDs(raw json.RawMessage) error {
	var ids []json.RawMessage
	if err := json.Unmarshal(raw, &ids); err != nil || len(ids) < 1 {
		return &ExecutionManifestError{Reason: "capability IDs", Cause: err}
	}
	seen := make(map[string]bool, len(ids))
	for _, rawID := range ids {
		var id string
		if string(rawID) == "null" || json.Unmarshal(rawID, &id) != nil || len(id) < 1 {
			return &ExecutionManifestError{Reason: "capability ID item"}
		}
		if seen[id] {
			return &ExecutionManifestError{Reason: "duplicate capability ID"}
		}
		seen[id] = true
	}
	return nil
}

func executionManifestPackageRefs(raw json.RawMessage) error {
	var rawPackages []json.RawMessage
	if err := json.Unmarshal(raw, &rawPackages); err != nil || len(rawPackages) < 1 {
		return &ExecutionManifestError{Reason: "package references", Cause: err}
	}
	seen := make(map[[2]string]bool, len(rawPackages))
	for _, rawPkg := range rawPackages {
		var pkg map[string]json.RawMessage
		if err := json.Unmarshal(rawPkg, &pkg); err != nil || len(pkg) != 2 {
			return &ExecutionManifestError{Reason: "package reference object", Cause: err}
		}
		descRaw, okDesc := pkg["descriptor_ref"]
		contentRaw, okContent := pkg["content_ref"]
		if !okDesc || !okContent {
			return &ExecutionManifestError{Reason: "package reference object fields"}
		}
		if !validProfileRef(descRaw, false) || !validProfileRef(contentRaw, false) {
			return &ExecutionManifestError{Reason: "package reference artifact"}
		}
		key := [2]string{profileString(descRaw), profileString(contentRaw)}
		if seen[key] {
			return &ExecutionManifestError{Reason: "duplicate package reference"}
		}
		seen[key] = true
	}
	return nil
}
