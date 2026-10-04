package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func admittedCaptureFixture(t *testing.T) (snapshotProfilesFixture, map[string]any, snapshotCapture) {
	t.Helper()
	f, variant := admittedOriginFixture(t)
	capture := snapshotVector(t, "content-change-changes-root")
	variant["variant_manifest_ref"] = "sha256:" + capture.ManifestHash
	variant["variant_content_hash"] = capture.Root
	before := sha256.Sum256([]byte(f.capture.CapturedFiles["main.txt"]))
	after := sha256.Sum256([]byte(capture.CapturedFiles["main.txt"]))
	variant["declared_changes"] = []any{map[string]any{
		"path": "main.txt", "before_ref": "sha256:" + hex.EncodeToString(before[:]),
		"after_ref": "sha256:" + hex.EncodeToString(after[:]), "reason": "synthetic fixture change",
	}}
	return f, variant, capture
}

func captureFiles(capture snapshotCapture) map[string][]byte {
	files := make(map[string][]byte)
	for path, content := range capture.CapturedFiles {
		files[path] = []byte(content)
	}
	return files
}

func checkAdmittedCapture(t *testing.T, f snapshotProfilesFixture, variant map[string]any, capture snapshotCapture) error {
	t.Helper()
	raw := snapshotProfileJSON(t, variant)
	snapshot, admission := bindAdmittedOrigin(t, f, raw)
	return VerifyAdmittedVariantCapture(snapshot, f.capture.Manifest, admission, raw, capture.Manifest, captureFiles(f.capture), captureFiles(capture))
}

func TestAdmittedVariantCaptureControl(t *testing.T) {
	t.Parallel()
	f, variant, capture := admittedCaptureFixture(t)
	if err := checkAdmittedCapture(t, f, variant, capture); err != nil {
		t.Fatal(err)
	}
	// Both captures remain distinct and unmodified, with no filesystem writes.
	raw := snapshotProfileJSON(t, variant)
	snapshot, admission := bindAdmittedOrigin(t, f, raw)
	originalFiles, variantFiles := captureFiles(f.capture), captureFiles(capture)
	beforeOriginal := bytes.Clone(originalFiles["main.txt"])
	beforeVariant := bytes.Clone(variantFiles["main.txt"])
	if err := VerifyAdmittedVariantCapture(snapshot, f.capture.Manifest, admission, raw, capture.Manifest, originalFiles, variantFiles); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(originalFiles["main.txt"], beforeOriginal) || !bytes.Equal(variantFiles["main.txt"], beforeVariant) || bytes.Equal(beforeOriginal, beforeVariant) {
		t.Fatal("captures were changed or collapsed")
	}
}

func TestAdmittedVariantCaptureIdentityMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		field, reason string
		value         func(snapshotProfilesFixture, snapshotCapture) string
	}{
		{"variant_manifest_ref", "variant manifest reference mismatch", func(f snapshotProfilesFixture, _ snapshotCapture) string {
			return "sha256:" + f.capture.ManifestHash
		}},
		{"variant_manifest_ref", "variant manifest reference mismatch", func(_ snapshotProfilesFixture, c snapshotCapture) string {
			return "sha256:" + c.Root
		}},
		{"variant_content_hash", "variant content root mismatch", func(f snapshotProfilesFixture, _ snapshotCapture) string {
			return f.capture.Root
		}},
		{"variant_content_hash", "variant content root mismatch", func(_ snapshotProfilesFixture, c snapshotCapture) string {
			return c.ManifestHash
		}},
	} {
		f, variant, capture := admittedCaptureFixture(t)
		variant[test.field] = test.value(f, capture)
		// Renew variant/admission refs so capture identity is the actual failure.
		assertSnapshotRejected(t, checkAdmittedCapture(t, f, variant, capture), test.reason)
	}
	f, variant, _ := admittedCaptureFixture(t)
	assertSnapshotRejected(t, checkAdmittedCapture(t, f, variant, f.capture), "variant manifest reference mismatch")
}

func TestAdmittedVariantCaptureMissingOrChangedBytes(t *testing.T) {
	t.Parallel()
	f, variant, capture := admittedCaptureFixture(t)
	raw := snapshotProfileJSON(t, variant)
	snapshot, admission := bindAdmittedOrigin(t, f, raw)
	for _, test := range []struct {
		files  map[string][]byte
		reason string
	}{
		{nil, "missing file bytes or size mismatch"},
		{map[string][]byte{"main.txt": []byte("short")}, "missing file bytes or size mismatch"},
		{map[string][]byte{"main.txt": []byte("JELLO\n")}, "file digest mismatch"},
		{captureFiles(f.capture), "file digest mismatch"},
	} {
		err := VerifyAdmittedVariantCapture(snapshot, f.capture.Manifest, admission, raw, capture.Manifest, captureFiles(f.capture), test.files)
		assertSnapshotRejected(t, err, "invalid variant manifest capture")
		var cause *SourceManifestError
		if !errors.As(err, &cause) || cause.Reason != test.reason {
			t.Fatalf("wanted variant manifest cause %q; got %v", test.reason, err)
		}
	}
	// Missing original bytes are likewise not supplied from the variant capture.
	err := VerifyAdmittedVariantCapture(snapshot, f.capture.Manifest, admission, raw, capture.Manifest, nil, captureFiles(capture))
	assertSnapshotRejected(t, err, "invalid manifest capture")
	err = VerifyAdmittedVariantCapture(snapshot, f.capture.Manifest, admission, raw, nil, captureFiles(f.capture), captureFiles(capture))
	assertSnapshotRejected(t, err, "invalid variant manifest capture")
}

func TestAdmittedVariantCaptureOriginPrerequisite(t *testing.T) {
	t.Parallel()
	f, variant, capture := admittedCaptureFixture(t)
	variant["original_content_hash"] = capture.Root
	assertSnapshotRejected(t, checkAdmittedCapture(t, f, variant, capture), "variant original content mismatch")
}
