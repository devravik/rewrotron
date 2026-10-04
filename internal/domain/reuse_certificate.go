package domain

import (
	"encoding/json"
)

// ReuseCertificateError describes an invalid reuse-certificate record without echoing private contents.
type ReuseCertificateError struct {
	Reason string
	Cause  error
}

// Error identifies the failed reuse-certificate obligation.
func (e *ReuseCertificateError) Error() string {
	return "invalid reuse certificate: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ReuseCertificateError) Unwrap() error { return e.Cause }

// ValidateReuseCertificate validates a reuse-certificate record under schema 2.0 (historical)
// or 2.1 (current).
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level fields,
// identifier patterns, plan digest formats, artifact and graph references, kind enumeration,
// unique checked edges and evidence references, and required consumption coverage references
// for schema 2.1.
func ValidateReuseCertificate(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ReuseCertificateError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ReuseCertificateError{Reason: "object fields", Cause: err}
	}

	versionRaw, present := record["schema_version"]
	if !present {
		return &ReuseCertificateError{Reason: "missing required field"}
	}
	version := profileString(versionRaw)
	if version != "2.0" && version != "2.1" {
		return &ReuseCertificateError{Reason: "unsupported schema version"}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"old_plan_hash",
		"new_plan_hash",
		"artifact_ref",
		"graph_ref",
		"kind",
		"checked_edge_refs",
		"evidence_refs",
		"issuer",
		"valid",
	}
	if version == "2.1" {
		expectedFields = append(expectedFields, "consumption_coverage_refs")
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ReuseCertificateError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ReuseCertificateError{Reason: "object fields"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ReuseCertificateError{Reason: "project identity"}
	}

	oldHash := profileString(record["old_plan_hash"])
	if !isManifestDigest(oldHash) {
		return &ReuseCertificateError{Reason: "old_plan_hash"}
	}

	newHash := profileString(record["new_plan_hash"])
	if !isManifestDigest(newHash) {
		return &ReuseCertificateError{Reason: "new_plan_hash"}
	}

	if !validProfileRef(record["artifact_ref"], false) {
		return &ReuseCertificateError{Reason: "artifact_ref"}
	}

	if !validProfileRef(record["graph_ref"], false) {
		return &ReuseCertificateError{Reason: "graph_ref"}
	}

	kind := profileString(record["kind"])
	if kind != "CHARACTERIZATION" && kind != "IMPLEMENTATION" {
		return &ReuseCertificateError{Reason: "invalid kind"}
	}

	if err := validateReuseRefList(record["checked_edge_refs"], "checked_edge_refs"); err != nil {
		return err
	}

	if err := validateReuseRefList(record["evidence_refs"], "evidence_refs"); err != nil {
		return err
	}

	if profileString(record["issuer"]) != "ENGINE" {
		return &ReuseCertificateError{Reason: "issuer must be ENGINE"}
	}

	var validVal bool
	if string(record["valid"]) == "null" || json.Unmarshal(record["valid"], &validVal) != nil || !validVal {
		return &ReuseCertificateError{Reason: "valid must be true"}
	}

	if version == "2.1" {
		if err := validateReuseRefList(record["consumption_coverage_refs"], "consumption_coverage_refs"); err != nil {
			return err
		}
	}

	return nil
}

func validateReuseRefList(raw json.RawMessage, fieldName string) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil || len(refs) < 1 {
		return &ReuseCertificateError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(refs))
	for _, item := range refs {
		if !validProfileRef(item, false) {
			return &ReuseCertificateError{Reason: "invalid " + fieldName + " reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return &ReuseCertificateError{Reason: "duplicate " + fieldName + " reference"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}
