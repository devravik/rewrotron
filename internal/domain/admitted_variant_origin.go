package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// VerifyAdmittedVariantOrigin binds a supplied execution variant to a current
// snapshot/manifest capture and its ADMITTED_VARIANT record. Artifact references
// hash exact supplied bytes. This checks the original identity only: variant
// manifest/file bytes, transform applicability, policies, review/control evidence
// and filesystem-profile closure remain unqualified. It grants no execution
// authority and never replaces the private original. Callers must not mutate
// supplied buffers during verification.
func VerifyAdmittedVariantOrigin(snapshot, manifest, admission, variant []byte, capturedFiles map[string][]byte) error {
	if err := VerifySourceSnapshotManifest(snapshot, manifest, capturedFiles); err != nil {
		return err
	}
	if err := ValidateExecutionAdmission(admission); err != nil {
		return &SourceSnapshotError{Reason: "invalid execution admission", Cause: err}
	}
	var source map[string]json.RawMessage
	if err := json.Unmarshal(snapshot, &source); err != nil {
		return &SourceSnapshotError{Reason: "snapshot projection", Cause: err}
	}
	if err := verifySnapshotProfileBinding(source, admission, "execution_admission_ref", "original_manifest_ref"); err != nil {
		return err
	}
	var admitted map[string]json.RawMessage
	if err := json.Unmarshal(admission, &admitted); err != nil {
		return &SourceSnapshotError{Reason: "admission projection", Cause: err}
	}
	if profileString(admitted["outcome"]) != "ADMITTED_VARIANT" {
		return &SourceSnapshotError{Reason: "variant admission required"}
	}
	if err := ValidateExecutionVariant(variant); err != nil {
		return &SourceSnapshotError{Reason: "invalid execution variant", Cause: err}
	}
	digest := sha256.Sum256(variant)
	if profileString(admitted["variant_ref"]) != "sha256:"+hex.EncodeToString(digest[:]) {
		return &SourceSnapshotError{Reason: "variant reference mismatch"}
	}
	var origin map[string]json.RawMessage
	if err := json.Unmarshal(variant, &origin); err != nil {
		return &SourceSnapshotError{Reason: "variant projection", Cause: err}
	}
	if profileString(origin["project_id"]) != profileString(source["project_id"]) {
		return &SourceSnapshotError{Reason: "variant project mismatch"}
	}
	if profileString(origin["original_manifest_ref"]) != profileString(source["manifest_ref"]) {
		return &SourceSnapshotError{Reason: "variant original manifest mismatch"}
	}
	if profileString(origin["original_content_hash"]) != profileString(source["content_root_hash"]) {
		return &SourceSnapshotError{Reason: "variant original content mismatch"}
	}
	return nil
}
