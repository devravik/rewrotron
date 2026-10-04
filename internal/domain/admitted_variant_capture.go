package domain

import "encoding/json"

// VerifyAdmittedVariantCapture binds an admitted variant to its verified original
// and verifies the variant's separate manifest reference, content root and FILE
// bytes. Captures are supplied independently; missing variant bytes are never
// filled from the original. This is a read-only identity check, not proof of
// transform applicability, declared-change completeness, safe credentials,
// coherent acquisition, profile/control qualification or execution approval.
// Callers must not mutate any supplied buffers during verification.
func VerifyAdmittedVariantCapture(snapshot, originalManifest, admission, variant, variantManifest []byte, originalFiles, variantFiles map[string][]byte) error {
	if err := VerifyAdmittedVariantOrigin(snapshot, originalManifest, admission, variant, originalFiles); err != nil {
		return err
	}
	root, digest, err := VerifySourceManifest(variantManifest, variantFiles)
	if err != nil {
		return &SourceSnapshotError{Reason: "invalid variant manifest capture", Cause: err}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(variant, &record); err != nil {
		return &SourceSnapshotError{Reason: "variant projection", Cause: err}
	}
	if profileString(record["variant_manifest_ref"]) != "sha256:"+digest {
		return &SourceSnapshotError{Reason: "variant manifest reference mismatch"}
	}
	if profileString(record["variant_content_hash"]) != root {
		return &SourceSnapshotError{Reason: "variant content root mismatch"}
	}
	return nil
}
