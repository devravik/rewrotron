package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
)

func admittedOriginFixture(t *testing.T) (snapshotProfilesFixture, map[string]any) {
	t.Helper()
	f := newSnapshotProfilesFixture(t)
	f.admission["outcome"] = "ADMITTED_VARIANT"
	f.admission["variant_credential_status"] = "SYNTHETIC_OR_ABSENT"
	variant := newVariantFixture(t)
	variant["original_manifest_ref"] = f.snapshot["manifest_ref"]
	variant["original_content_hash"] = f.snapshot["content_root_hash"]
	return f, variant
}

func bindAdmittedOrigin(t *testing.T, f snapshotProfilesFixture, variant []byte) (snapshot, admission []byte) {
	t.Helper()
	if f.admission["outcome"] == "ADMITTED_VARIANT" {
		digest := sha256.Sum256(variant)
		f.admission["variant_ref"] = "sha256:" + hex.EncodeToString(digest[:])
	}
	admission = snapshotProfileJSON(t, f.admission)
	bindSnapshotProfileBytes(f.snapshot, snapshotProfileJSON(t, f.profile), admission)
	return snapshotProfileJSON(t, f.snapshot), admission
}

func verifyOriginBytes(f snapshotProfilesFixture, snapshot, admission, variant []byte) error {
	files := make(map[string][]byte)
	for path, content := range f.capture.CapturedFiles {
		files[path] = []byte(content)
	}
	return VerifyAdmittedVariantOrigin(snapshot, f.capture.Manifest, admission, variant, files)
}

func checkAdmittedOrigin(t *testing.T, f snapshotProfilesFixture, variant map[string]any) error {
	t.Helper()
	raw := snapshotProfileJSON(t, variant)
	snapshot, admission := bindAdmittedOrigin(t, f, raw)
	return verifyOriginBytes(f, snapshot, admission, raw)
}

func TestAdmittedVariantOriginControls(t *testing.T) {
	t.Parallel()
	for _, original := range []string{"SYNTHETIC_OR_ABSENT", "REVOKED", "LIVE_OR_UNKNOWN"} {
		for _, replacement := range []string{"SYNTHETIC_OR_ABSENT", "REVOKED"} {
			f, variant := admittedOriginFixture(t)
			f.admission["credential_status"] = original
			f.admission["variant_credential_status"] = replacement
			if err := checkAdmittedOrigin(t, f, variant); err != nil {
				t.Fatal(err)
			}
			if f.admission["credential_status"] != original {
				t.Fatal("original credential status changed")
			}
		}
	}
}

func TestAdmittedVariantOriginCrossBindings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		field, value, reason string
	}{
		{"project_id", "private_other_project", "variant project mismatch"},
		{"original_manifest_ref", "sha256:" + strings.Repeat("b", 64), "variant original manifest mismatch"},
		{"original_content_hash", strings.Repeat("b", 64), "variant original content mismatch"},
	} {
		f, variant := admittedOriginFixture(t)
		variant[test.field] = test.value
		// Fresh variant and admission refs expose the actual origin mismatch.
		err := checkAdmittedOrigin(t, f, variant)
		assertSnapshotRejected(t, err, test.reason)
		if strings.Contains(err.Error(), test.value) {
			t.Fatal("private origin echoed in error")
		}
	}
	for _, field := range []string{"project_id", "original_manifest_ref"} {
		f, variant := admittedOriginFixture(t)
		reason := "execution_admission_ref project mismatch"
		f.admission[field] = "other_project"
		if field == "original_manifest_ref" {
			f.admission[field] = "sha256:" + strings.Repeat("b", 64)
			reason = "execution_admission_ref manifest mismatch"
		}
		assertSnapshotRejected(t, checkAdmittedOrigin(t, f, variant), reason)
	}
}

func TestAdmittedVariantOriginExactBytes(t *testing.T) {
	t.Parallel()
	f, variant := admittedOriginFixture(t)
	raw := snapshotProfileJSON(t, variant)
	snapshot, admission := bindAdmittedOrigin(t, f, raw)
	changed := append(append([]byte(nil), raw...), '\n')
	assertSnapshotRejected(t, verifyOriginBytes(f, snapshot, admission, changed), "variant reference mismatch")
	snapshot, admission = bindAdmittedOrigin(t, f, changed)
	if err := verifyOriginBytes(f, snapshot, admission, changed); err != nil {
		t.Fatal(err)
	}
	// Refreshing a variant ref alone leaves the snapshot's admission ref stale.
	f.admission["variant_ref"] = "sha256:" + strings.Repeat("c", 64)
	staleAdmission := snapshotProfileJSON(t, f.admission)
	assertSnapshotRejected(t, verifyOriginBytes(f, snapshot, staleAdmission, changed), "execution_admission_ref mismatch")
}

func TestAdmittedVariantOriginPrerequisites(t *testing.T) {
	t.Parallel()
	for _, outcome := range []string{"ADMITTED_ORIGINAL", "REFUSED", "STATIC_ONLY"} {
		f, variant := admittedOriginFixture(t)
		f.admission["outcome"] = outcome
		f.admission["variant_ref"], f.admission["variant_credential_status"] = nil, nil
		f.admission["credential_status"] = "REVOKED"
		assertSnapshotRejected(t, checkAdmittedOrigin(t, f, variant), "variant admission required")
	}
	f, variant := admittedOriginFixture(t)
	variant["declared_changes"] = []any{}
	err := checkAdmittedOrigin(t, f, variant)
	assertSnapshotRejected(t, err, "invalid execution variant")
	var invalidVariant *ExecutionVariantError
	if !errors.As(err, &invalidVariant) || invalidVariant.Reason != "declared changes required" {
		t.Fatal("variant validation cause lost")
	}
	f, variant = admittedOriginFixture(t)
	f.admission["variant_credential_status"] = "LIVE_OR_UNKNOWN"
	err = checkAdmittedOrigin(t, f, variant)
	assertSnapshotRejected(t, err, "invalid execution admission")
	var invalidAdmission *ExecutionAdmissionError
	if !errors.As(err, &invalidAdmission) || invalidAdmission.Reason != "unsafe variant credential status" {
		t.Fatal("unsafe replacement cause lost")
	}
	f, variant = admittedOriginFixture(t)
	f.snapshot["content_root_hash"] = f.capture.ManifestHash
	assertSnapshotRejected(t, checkAdmittedOrigin(t, f, variant), "content root mismatch")
	f, _ = admittedOriginFixture(t)
	snapshot, admission := bindAdmittedOrigin(t, f, nil)
	assertSnapshotRejected(t, verifyOriginBytes(f, snapshot, admission, nil), "invalid execution variant")
}
