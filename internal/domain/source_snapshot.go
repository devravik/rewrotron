package domain

import (
	"encoding/json"
	"strings"
)

// SourceSnapshotError identifies an invalid record or content binding.
type SourceSnapshotError struct {
	Reason string
	Cause  error
}

// Error describes the rejection without echoing private record contents.
func (e *SourceSnapshotError) Error() string { return "invalid source snapshot: " + e.Reason }

// Unwrap preserves manifest and JSON failures for structured inspection.
func (e *SourceSnapshotError) Unwrap() error { return e.Cause }

// VerifySourceSnapshotManifest validates a current snapshot 2.1 record and binds
// its manifest reference and content root to the supplied manifest/file bytes.
// Historical 2.0 records require a separate read-only history path; they are not
// upgraded here. This check does not acquire, register, persist or seal anything.
// SEALED is a required declared field, not evidence of actual publication.
// Artifact ownership, profile/closure evidence, acquisition coherence, private
// storage and approval must be qualified separately by the snapshot service.
// Callers must not mutate the supplied buffers during verification.
func VerifySourceSnapshotManifest(snapshot, manifest []byte, capturedFiles map[string][]byte) error {
	canonical, err := CanonicalJSON(snapshot)
	if err != nil {
		return &SourceSnapshotError{Reason: "canonical JSON", Cause: err}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &SourceSnapshotError{Reason: "object required", Cause: err}
	}
	if profileString(record["schema_version"]) != "2.1" {
		return &SourceSnapshotError{Reason: "current schema version required"}
	}
	fields := []string{"schema_version", "snapshot_id", "project_id", "acquisition", "manifest_ref", "environment_profile_ref", "operational_profile_ref", "closure_evidence_ref", "content_root_hash", "private_location_id", "publication_state", "filesystem_profile_ref", "execution_admission_ref"}
	if len(record) != len(fields) {
		return &SourceSnapshotError{Reason: "object fields"}
	}
	for _, field := range fields {
		if _, present := record[field]; !present {
			return &SourceSnapshotError{Reason: "missing required field"}
		}
	}
	for _, field := range []string{"snapshot_id", "project_id", "private_location_id"} {
		value := profileString(record[field])
		if value == "" || strings.IndexFunc(value, func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
		}) >= 0 {
			return &SourceSnapshotError{Reason: "record identity"}
		}
	}
	for _, field := range []string{"manifest_ref", "environment_profile_ref", "operational_profile_ref", "closure_evidence_ref", "filesystem_profile_ref", "execution_admission_ref"} {
		if !validProfileRef(record[field], false) {
			return &SourceSnapshotError{Reason: "artifact reference"}
		}
	}
	if !isManifestDigest(profileString(record["content_root_hash"])) {
		return &SourceSnapshotError{Reason: "content root syntax"}
	}
	switch profileString(record["acquisition"]) {
	case "IMMUTABLE_VCS_EXPORT", "FILESYSTEM_SNAPSHOT", "QUIESCED_COPY":
	default:
		return &SourceSnapshotError{Reason: "unsupported acquisition"}
	}
	if profileString(record["publication_state"]) != "SEALED" {
		return &SourceSnapshotError{Reason: "declared publication state"}
	}
	root, digest, err := VerifySourceManifest(manifest, capturedFiles)
	if err != nil {
		return &SourceSnapshotError{Reason: "invalid manifest capture", Cause: err}
	}
	if profileString(record["manifest_ref"]) != "sha256:"+digest {
		return &SourceSnapshotError{Reason: "manifest reference mismatch"}
	}
	if profileString(record["content_root_hash"]) != root {
		return &SourceSnapshotError{Reason: "content root mismatch"}
	}
	return nil
}
