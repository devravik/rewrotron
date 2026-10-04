package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// VerifySourceSnapshotProfiles checks current snapshot/manifest binding and the
// supplied filesystem-profile and execution-admission records. Artifact refs
// identify exact supplied bytes, without rewriting their JSON formatting.
// Both records must identify the snapshot's project and original manifest.
// This is a partial reference check, not snapshot publication or runtime
// admission: REFUSED/STATIC_ONLY records can be validly bound. Variant, policy,
// control, environment, operational and other evidence closures are not resolved
// here. Artifact registration/privacy, qualified capture and approval remain
// separate service obligations. Callers must not mutate buffers during checks.
func VerifySourceSnapshotProfiles(snapshot, manifest, filesystemProfile, admission []byte, capturedFiles map[string][]byte) error {
	if err := VerifySourceSnapshotManifest(snapshot, manifest, capturedFiles); err != nil {
		return err
	}
	if err := ValidateFilesystemProfile(filesystemProfile); err != nil {
		return &SourceSnapshotError{Reason: "invalid filesystem profile", Cause: err}
	}
	if err := ValidateExecutionAdmission(admission); err != nil {
		return &SourceSnapshotError{Reason: "invalid execution admission", Cause: err}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(snapshot, &record); err != nil {
		return &SourceSnapshotError{Reason: "snapshot projection", Cause: err}
	}
	if err := verifySnapshotProfileBinding(record, filesystemProfile, "filesystem_profile_ref", "content_manifest_ref"); err != nil {
		return err
	}
	return verifySnapshotProfileBinding(record, admission, "execution_admission_ref", "original_manifest_ref")
}

func verifySnapshotProfileBinding(snapshot map[string]json.RawMessage, data []byte, referenceField, manifestField string) error {
	digest := sha256.Sum256(data)
	if profileString(snapshot[referenceField]) != "sha256:"+hex.EncodeToString(digest[:]) {
		return &SourceSnapshotError{Reason: referenceField + " mismatch"}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(data, &record); err != nil {
		return &SourceSnapshotError{Reason: "bound record projection", Cause: err}
	}
	if profileString(record["project_id"]) != profileString(snapshot["project_id"]) {
		return &SourceSnapshotError{Reason: referenceField + " project mismatch"}
	}
	if profileString(record[manifestField]) != profileString(snapshot["manifest_ref"]) {
		return &SourceSnapshotError{Reason: referenceField + " manifest mismatch"}
	}
	return nil
}
