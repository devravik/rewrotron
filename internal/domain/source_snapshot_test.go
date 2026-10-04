package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type snapshotCapture struct {
	Name          string            `json:"name"`
	Manifest      json.RawMessage   `json:"manifest"`
	Admission     string            `json:"expected_admission"`
	Root          string            `json:"sha256"`
	ManifestHash  string            `json:"manifest_sha256"`
	CapturedFiles map[string]string `json:"fixture_bytes_utf8"`
}

func snapshotCaptures(t *testing.T) []snapshotCapture {
	t.Helper()
	var corpus struct {
		Vectors []snapshotCapture `json:"source_root_hash"`
	}
	loadVectors(t, "../../specs/examples/contract-blocker-vectors.json", &corpus)
	return corpus.Vectors
}

func snapshotRecord(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "review2-current-v2-snapshot" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing authoritative current snapshot fixture")
	return nil
}

func snapshotVector(t *testing.T, name string) snapshotCapture {
	t.Helper()
	for _, vector := range snapshotCaptures(t) {
		if vector.Name == name {
			return vector
		}
	}
	t.Fatal("missing manifest control")
	return snapshotCapture{}
}

func bindSnapshotRecord(record map[string]any, vector snapshotCapture) {
	record["manifest_ref"] = "sha256:" + vector.ManifestHash
	record["content_root_hash"] = vector.Root
}

func checkSnapshotRecord(t *testing.T, record map[string]any, vector snapshotCapture) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte)
	for path, text := range vector.CapturedFiles {
		files[path] = []byte(text)
	}
	return VerifySourceSnapshotManifest(raw, vector.Manifest, files)
}

func assertSnapshotRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *SourceSnapshotError
	if !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted snapshot rejection %q; got %v", reason, err)
	}
}

func TestSourceSnapshotManifestGoldenBindings(t *testing.T) {
	t.Parallel()
	count := 0
	for _, vector := range snapshotCaptures(t) {
		if vector.Admission != "ACCEPT" {
			continue
		}
		count++
		t.Run(vector.Name, func(t *testing.T) {
			t.Parallel()
			record := snapshotRecord(t)
			bindSnapshotRecord(record, vector)
			if err := checkSnapshotRecord(t, record, vector); err != nil {
				t.Fatal(err)
			}
		})
	}
	if count != 8 {
		t.Fatal("manifest admission controls changed; review coverage")
	}
}

func TestSourceSnapshotRepositoryRecords(t *testing.T) {
	t.Parallel()
	vector := snapshotVector(t, "empty-tree")
	count := 0
	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/source-snapshot.schema.json" {
			continue
		}
		count++
		var record map[string]any
		if err := json.Unmarshal(fixture.Instance, &record); err != nil {
			t.Fatal(err)
		}
		// These schema fixtures have placeholder identities, not bound captures.
		bindSnapshotRecord(record, vector)
		err := checkSnapshotRecord(t, record, vector)
		switch fixture.Name {
		case "v2-snapshot":
			assertSnapshotRejected(t, err, "current schema version required")
		case "review2-current-v2-snapshot-requires-provenance":
			assertSnapshotRejected(t, err, "object fields")
		case "review2-current-v2-snapshot":
			if err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatal("new fixture needs current/historical admission classification")
		}
	}
	if count != 3 {
		t.Fatal("snapshot fixture corpus changed; review coverage")
	}
}

func TestSourceSnapshotRecordRejections(t *testing.T) {
	t.Parallel()
	vector := snapshotVector(t, "empty-tree")
	for _, test := range []struct {
		field, reason string
		value         any
	}{
		{"schema_version", "current schema version required", "2.0"},
		{"snapshot_id", "record identity", nil},
		{"project_id", "record identity", "Fixture"},
		{"private_location_id", "record identity", "../vault"},
		{"acquisition", "unsupported acquisition", "LIVE_SCAN"},
		{"publication_state", "declared publication state", "PENDING"},
		{"content_root_hash", "content root syntax", "sha256:" + vector.Root},
		{"content_root_hash", "content root mismatch", vector.ManifestHash},
		{"manifest_ref", "manifest reference mismatch", "sha256:" + vector.Root},
		{"extra", "object fields", true},
	} {
		t.Run(test.field+"/"+test.reason, func(t *testing.T) {
			t.Parallel()
			record := snapshotRecord(t)
			bindSnapshotRecord(record, vector)
			record[test.field] = test.value
			assertSnapshotRejected(t, checkSnapshotRecord(t, record, vector), test.reason)
		})
	}
	for _, field := range []string{"manifest_ref", "environment_profile_ref", "operational_profile_ref", "closure_evidence_ref", "filesystem_profile_ref", "execution_admission_ref"} {
		for _, value := range []any{nil, true, "sha256:" + strings.Repeat("A", 64)} {
			record := snapshotRecord(t)
			bindSnapshotRecord(record, vector)
			record[field] = value
			assertSnapshotRejected(t, checkSnapshotRecord(t, record, vector), "artifact reference")
		}
	}
	for field := range snapshotRecord(t) {
		record := snapshotRecord(t)
		delete(record, field)
		reason := "object fields"
		if field == "schema_version" {
			reason = "current schema version required"
		}
		assertSnapshotRejected(t, checkSnapshotRecord(t, record, vector), reason)
	}
	record := snapshotRecord(t)
	delete(record, "filesystem_profile_ref")
	record["extra"] = nil
	assertSnapshotRejected(t, checkSnapshotRecord(t, record, vector), "missing required field")
	for _, acquisition := range []string{"IMMUTABLE_VCS_EXPORT", "FILESYSTEM_SNAPSHOT", "QUIESCED_COPY"} {
		record := snapshotRecord(t)
		bindSnapshotRecord(record, vector)
		record["acquisition"] = acquisition
		if err := checkSnapshotRecord(t, record, vector); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSourceSnapshotStaleCapture(t *testing.T) {
	t.Parallel()
	vector := snapshotVector(t, "raw-file")
	record := snapshotRecord(t)
	bindSnapshotRecord(record, vector)
	// A scanner change retains the content root but changes the full manifest.
	changed := snapshotVector(t, "scanner-change-same-root")
	assertSnapshotRejected(t, checkSnapshotRecord(t, record, changed), "manifest reference mismatch")
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	for _, capture := range []struct {
		files  map[string][]byte
		reason string
	}{
		{nil, "missing file bytes or size mismatch"},
		{map[string][]byte{"main.txt": []byte("jello\n")}, "file digest mismatch"},
	} {
		err := VerifySourceSnapshotManifest(raw, vector.Manifest, capture.files)
		assertSnapshotRejected(t, err, "invalid manifest capture")
		var invalid *SourceManifestError
		if !errors.As(err, &invalid) || invalid.Reason != capture.reason {
			t.Fatalf("wanted manifest cause %q; got %v", capture.reason, err)
		}
	}
	for _, malformed := range []string{`{"schema_version":"2.1","schema_version":"2.1"}`, `{"private":"\ud800"}`, `{} {}`} {
		assertSnapshotRejected(t, VerifySourceSnapshotManifest([]byte(malformed), vector.Manifest, nil), "canonical JSON")
	}
	assertSnapshotRejected(t, VerifySourceSnapshotManifest([]byte(`null`), vector.Manifest, nil), "object required")
}
