package domain

import (
	"encoding/json"
	"strings"
)

// ArchiveManifestError describes an invalid archive-manifest record without echoing private contents.
type ArchiveManifestError struct {
	Reason string
	Cause  error
}

// Error identifies the failed archive-manifest obligation.
func (e *ArchiveManifestError) Error() string {
	return "invalid archive manifest: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ArchiveManifestError) Unwrap() error { return e.Cause }

// ValidateArchiveManifest validates an archive 2.0 manifest record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, database and migration digest formats, monotonic
// through_sequence, sorted and unique object closures with valid retention classes,
// and artifact references.
//
// This check does not execute filesystem extraction, verify external object bytes,
// or qualify encryption keys.
func ValidateArchiveManifest(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ArchiveManifestError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ArchiveManifestError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"db_hash",
		"through_sequence",
		"migration_hashes",
		"objects",
		"config_ref",
		"private_key_policy_ref",
		"closure_report_ref",
	}

	if len(record) != len(expectedFields) {
		return &ArchiveManifestError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ArchiveManifestError{Reason: "missing required field"}
		}
	}

	if profileString(record["schema_version"]) != "2.0" {
		return &ArchiveManifestError{Reason: "unsupported schema version"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ArchiveManifestError{Reason: "project identity"}
	}

	dbHash := profileString(record["db_hash"])
	if !isManifestDigest(dbHash) {
		return &ArchiveManifestError{Reason: "database digest"}
	}

	var throughSeq int64
	if string(record["through_sequence"]) == "null" || json.Unmarshal(record["through_sequence"], &throughSeq) != nil || throughSeq < 0 {
		return &ArchiveManifestError{Reason: "through_sequence"}
	}

	if err := validateArchiveMigrationHashes(record["migration_hashes"]); err != nil {
		return err
	}

	if err := validateArchiveObjects(record["objects"]); err != nil {
		return err
	}

	for _, field := range []string{"config_ref", "private_key_policy_ref", "closure_report_ref"} {
		if !validProfileRef(record[field], false) {
			return &ArchiveManifestError{Reason: field}
		}
	}

	return nil
}

func validateArchiveMigrationHashes(raw json.RawMessage) error {
	var hashes []json.RawMessage
	if err := json.Unmarshal(raw, &hashes); err != nil || len(hashes) < 1 {
		return &ArchiveManifestError{Reason: "migration hashes array", Cause: err}
	}

	seen := make(map[string]struct{}, len(hashes))
	for _, rawHash := range hashes {
		h := profileString(rawHash)
		if !isManifestDigest(h) {
			return &ArchiveManifestError{Reason: "migration hash format"}
		}
		if _, exists := seen[h]; exists {
			return &ArchiveManifestError{Reason: "duplicate migration hash"}
		}
		seen[h] = struct{}{}
	}
	return nil
}

func validateArchiveObjects(raw json.RawMessage) error {
	var objects []json.RawMessage
	if err := json.Unmarshal(raw, &objects); err != nil || objects == nil {
		return &ArchiveManifestError{Reason: "objects array", Cause: err}
	}

	var previousRef string
	expectedFields := []string{"ref", "size_bytes", "retention_class", "private"}

	for i, rawObj := range objects {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(rawObj, &obj); err != nil || len(obj) != len(expectedFields) {
			return &ArchiveManifestError{Reason: "object entry fields"}
		}
		for _, field := range expectedFields {
			if _, present := obj[field]; !present {
				return &ArchiveManifestError{Reason: "missing object entry field"}
			}
		}

		refRaw := obj["ref"]
		if !validProfileRef(refRaw, false) {
			return &ArchiveManifestError{Reason: "object reference"}
		}
		ref := profileString(refRaw)

		if i > 0 && strings.Compare(previousRef, ref) >= 0 {
			return &ArchiveManifestError{Reason: "object entries must be strictly sorted by reference"}
		}
		previousRef = ref

		var sizeBytes int64
		if string(obj["size_bytes"]) == "null" || json.Unmarshal(obj["size_bytes"], &sizeBytes) != nil || sizeBytes < 0 {
			return &ArchiveManifestError{Reason: "object size_bytes"}
		}

		retention := profileString(obj["retention_class"])
		switch retention {
		case "APPROVAL", "CONTRACT", "VERIFICATION", "EXPERIMENT", "LESSON",
			"TRACE", "SNAPSHOT", "HEARTBEAT", "AUDIT", "TEMPORARY", "GLOBAL_KNOWLEDGE":
		default:
			return &ArchiveManifestError{Reason: "unsupported retention class"}
		}

		var isPrivate bool
		if string(obj["private"]) == "null" || json.Unmarshal(obj["private"], &isPrivate) != nil {
			return &ArchiveManifestError{Reason: "object private flag"}
		}
	}

	return nil
}
