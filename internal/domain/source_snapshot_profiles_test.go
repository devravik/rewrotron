package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type snapshotProfilesFixture struct {
	snapshot, profile, admission map[string]any
	capture                      snapshotCapture
}

func newSnapshotProfilesFixture(t *testing.T) snapshotProfilesFixture {
	t.Helper()
	f := snapshotProfilesFixture{
		snapshot: snapshotRecord(t), profile: profileFixture(t),
		admission: newAdmissionFixture(t), capture: snapshotVector(t, "raw-file"),
	}
	bindSnapshotRecord(f.snapshot, f.capture)
	f.profile["content_manifest_ref"] = f.snapshot["manifest_ref"]
	f.admission["original_manifest_ref"] = f.snapshot["manifest_ref"]
	return f
}

func snapshotProfileJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func bindSnapshotProfileBytes(record map[string]any, profile, admission []byte) {
	for field, data := range map[string][]byte{"filesystem_profile_ref": profile, "execution_admission_ref": admission} {
		digest := sha256.Sum256(data)
		record[field] = "sha256:" + hex.EncodeToString(digest[:])
	}
}

func checkSnapshotProfiles(t *testing.T, f snapshotProfilesFixture) error {
	t.Helper()
	profile := snapshotProfileJSON(t, f.profile)
	admission := snapshotProfileJSON(t, f.admission)
	bindSnapshotProfileBytes(f.snapshot, profile, admission)
	return verifySnapshotProfileBytes(t, f, snapshotProfileJSON(t, f.snapshot), profile, admission)
}

func verifySnapshotProfileBytes(t *testing.T, f snapshotProfilesFixture, snapshot, profile, admission []byte) error {
	t.Helper()
	files := make(map[string][]byte)
	for path, content := range f.capture.CapturedFiles {
		files[path] = []byte(content)
	}
	return VerifySourceSnapshotProfiles(snapshot, f.capture.Manifest, profile, admission, files)
}

func TestSourceSnapshotProfilesOutcomes(t *testing.T) {
	t.Parallel()
	for _, outcome := range []string{"ADMITTED_ORIGINAL", "ADMITTED_VARIANT", "REFUSED", "STATIC_ONLY"} {
		t.Run(outcome, func(t *testing.T) {
			t.Parallel()
			f := newSnapshotProfilesFixture(t)
			f.admission["outcome"] = outcome
			switch outcome {
			case "ADMITTED_ORIGINAL":
				f.admission["credential_status"] = "REVOKED"
			case "ADMITTED_VARIANT":
				f.admission["variant_ref"] = "sha256:" + strings.Repeat("b", 64)
				f.admission["variant_credential_status"] = "SYNTHETIC_OR_ABSENT"
			}
			if err := checkSnapshotProfiles(t, f); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSourceSnapshotProfilesCrossBindings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		role, field, reason string
	}{
		{"profile", "project_id", "filesystem_profile_ref project mismatch"},
		{"admission", "project_id", "execution_admission_ref project mismatch"},
		{"profile", "content_manifest_ref", "filesystem_profile_ref manifest mismatch"},
		{"admission", "original_manifest_ref", "execution_admission_ref manifest mismatch"},
	} {
		t.Run(test.reason, func(t *testing.T) {
			t.Parallel()
			f := newSnapshotProfilesFixture(t)
			record := f.profile
			if test.role == "admission" {
				record = f.admission
			}
			value := "sha256:" + strings.Repeat("b", 64)
			if test.field == "project_id" {
				value = "private_canary_other_project"
			}
			record[test.field] = value
			// Rebind digests so the mutant is killed by ownership/manifest
			// correspondence rather than an unrelated stale-byte reference.
			err := checkSnapshotProfiles(t, f)
			assertSnapshotRejected(t, err, test.reason)
			if strings.Contains(err.Error(), value) {
				t.Fatal("private correlation echoed in error")
			}
		})
	}
}

func TestSourceSnapshotProfilesExactBytes(t *testing.T) {
	t.Parallel()
	f := newSnapshotProfilesFixture(t)
	profile := snapshotProfileJSON(t, f.profile)
	admission := snapshotProfileJSON(t, f.admission)
	bindSnapshotProfileBytes(f.snapshot, profile, admission)
	snapshot := snapshotProfileJSON(t, f.snapshot)
	for _, role := range []string{"profile", "admission"} {
		p, a := bytes.Clone(profile), bytes.Clone(admission)
		reason := "filesystem_profile_ref mismatch"
		if role == "profile" {
			p = append(p, '\n')
		} else {
			a = append(a, '\n')
			reason = "execution_admission_ref mismatch"
		}
		assertSnapshotRejected(t, verifySnapshotProfileBytes(t, f, snapshot, p, a), reason)
		// Same JSON meaning can be stored under a different exact-byte ref.
		bindSnapshotProfileBytes(f.snapshot, p, a)
		if err := verifySnapshotProfileBytes(t, f, snapshotProfileJSON(t, f.snapshot), p, a); err != nil {
			t.Fatal(err)
		}
	}
	if err := verifySnapshotProfileBytes(t, f, snapshot, profile, admission); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(profile, snapshotProfileJSON(t, f.profile)) || !bytes.Equal(admission, snapshotProfileJSON(t, f.admission)) {
		t.Fatal("checker changed supplied record bytes")
	}
}

func TestSourceSnapshotProfilesInvalidRecords(t *testing.T) {
	t.Parallel()
	f := newSnapshotProfilesFixture(t)
	properties := f.profile["properties"].([]any)
	properties[1] = properties[0]
	err := checkSnapshotProfiles(t, f)
	assertSnapshotRejected(t, err, "invalid filesystem profile")
	var profileError *FilesystemProfileError
	if !errors.As(err, &profileError) || profileError.Reason != "duplicate property" {
		t.Fatal("filesystem failure cause lost")
	}
	f = newSnapshotProfilesFixture(t)
	f.admission["outcome"] = "ADMITTED_ORIGINAL"
	err = checkSnapshotProfiles(t, f)
	assertSnapshotRejected(t, err, "invalid execution admission")
	var admissionError *ExecutionAdmissionError
	if !errors.As(err, &admissionError) || admissionError.Reason != "unsafe original credential status" {
		t.Fatal("admission failure cause lost")
	}
	f = newSnapshotProfilesFixture(t)
	f.snapshot["content_root_hash"] = f.capture.ManifestHash
	assertSnapshotRejected(t, checkSnapshotProfiles(t, f), "content root mismatch")
	f = newSnapshotProfilesFixture(t)
	profile := snapshotProfileJSON(t, f.profile)
	admission := snapshotProfileJSON(t, f.admission)
	bindSnapshotProfileBytes(f.snapshot, profile, admission)
	snapshot := snapshotProfileJSON(t, f.snapshot)
	for _, inputs := range [][2][]byte{{nil, admission}, {profile, nil}} {
		reason := "invalid filesystem profile"
		if inputs[0] != nil {
			reason = "invalid execution admission"
		}
		assertSnapshotRejected(t, verifySnapshotProfileBytes(t, f, snapshot, inputs[0], inputs[1]), reason)
	}
}
