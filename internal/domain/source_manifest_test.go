package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestSourceManifestGoldenVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Vectors []struct {
			Name          string            `json:"name"`
			Manifest      json.RawMessage   `json:"manifest"`
			Admission     string            `json:"expected_admission"`
			Root          string            `json:"sha256"`
			ManifestHash  string            `json:"manifest_sha256"`
			ClaimedRoot   string            `json:"claimed_root_hash"`
			CapturedFiles map[string]string `json:"fixture_bytes_utf8"`
		} `json:"source_root_hash"`
	}
	loadVectors(t, "../../specs/examples/contract-blocker-vectors.json", &document)
	if len(document.Vectors) != 14 {
		t.Fatal("source manifest vector corpus changed; review coverage")
	}
	for _, vector := range document.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			t.Parallel()
			files := make(map[string][]byte)
			for path, content := range vector.CapturedFiles {
				files[path] = []byte(content)
			}
			root, manifest, err := VerifySourceManifest(vector.Manifest, files)
			if vector.ClaimedRoot != "" {
				if err != nil {
					t.Fatal(err)
				}
				if root == vector.ClaimedRoot || manifest != vector.ClaimedRoot {
					t.Fatal("full-manifest hash was not distinguished from content root")
				}
				return
			}
			if vector.Admission == "REJECT" {
				if err == nil || root != "" || manifest != "" {
					t.Fatal("invalid manifest produced identities")
				}
				expected := map[string]string{
					"same-path-different-content": "entry paths must be strictly ordered",
					"unsorted-paths":              "entry paths must be strictly ordered",
					"absent-parent":               "missing directory parent",
					"file-as-parent":              "missing directory parent",
					"mutated-bytes-old-root":      "file digest mismatch",
				}[vector.Name]
				var invalid *SourceManifestError
				if expected == "" || !errors.As(err, &invalid) || invalid.Reason != expected {
					t.Fatalf("wanted rejection %q, got %v", expected, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if root != vector.Root || manifest != vector.ManifestHash {
				t.Fatalf("identities differ: root %s; manifest %s", root, manifest)
			}
		})
	}
}

func manifestFixture(t *testing.T, entries, exclusions any) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"schema_version": "1.0", "identity_algorithm": "rwt-source-root-1",
		"entries": entries, "exclusions": exclusions, "completeness": "COMPLETE",
		"scanner_ref": "sha256:" + strings.Repeat("a", 64),
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSourceManifestRejectsInvalidPaths(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"", "/absolute", "trailing/", "a//b", ".", "..", "a/../b", "./a", "a/.", "a\\b", "a\x00b"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			raw := manifestFixture(t, []any{map[string]any{"path": path, "kind": "DIRECTORY"}}, []any{})
			assertManifestRejected(t, raw, nil)
		})
	}
}

func TestSourceManifestExclusions(t *testing.T) {
	t.Parallel()
	entry := map[string]any{"path": "data", "kind": "DIRECTORY"}
	for _, test := range []struct {
		name       string
		exclusions []any
	}{
		{"same path", []any{map[string]any{"path": "data", "reason": "omitted"}}},
		{"duplicate", []any{map[string]any{"path": "x", "reason": "one"}, map[string]any{"path": "x", "reason": "two"}}},
		{"unsorted", []any{map[string]any{"path": "z", "reason": "one"}, map[string]any{"path": "x", "reason": "two"}}},
		{"empty reason", []any{map[string]any{"path": "x", "reason": ""}}},
		{"bad path", []any{map[string]any{"path": "../x", "reason": "one"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertManifestRejected(t, manifestFixture(t, []any{entry}, test.exclusions), nil)
		})
	}
	// Excluding data/tmp is valid while retaining its parent directory.
	valid := manifestFixture(t, []any{entry}, []any{map[string]any{"path": "data/tmp", "reason": "reviewed output"}})
	if _, _, err := VerifySourceManifest(valid, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSourceManifestFieldVocabulary(t *testing.T) {
	t.Parallel()
	base := manifestFixture(t, []any{}, []any{})
	for _, test := range []struct {
		name, field string
		value       any
	}{
		{"version", "schema_version", "2.0"},
		{"algorithm", "identity_algorithm", "unknown"},
		{"completeness", "completeness", "PARTIAL"},
		{"scanner", "scanner_ref", "sha256:" + strings.Repeat("A", 64)},
		{"null entries", "entries", nil},
		{"null exclusions", "exclusions", nil},
		{"unexpected field", "extra", true},
		{"wrong scanner type", "scanner_ref", 42},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var object map[string]any
			if err := json.Unmarshal(base, &object); err != nil {
				t.Fatal(err)
			}
			object[test.field] = test.value
			raw, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			assertManifestRejected(t, raw, nil)
		})
	}
	assertManifestRejected(t, []byte(`{"schema_version":"1.0","schema_version":"1.0"}`), nil)
	assertManifestRejected(t, []byte(`null`), nil)
}

func TestSourceManifestEntryVocabularyAndBytes(t *testing.T) {
	t.Parallel()
	for _, entry := range []map[string]any{
		{"path": "node", "kind": "FIFO"},
		{"path": "node", "kind": "DIRECTORY", "size_bytes": 0},
		{"path": "node", "kind": "SYMLINK", "target": ""},
		{"path": "node", "kind": "SYMLINK", "target": "a\x00b"},
		{"path": "node", "kind": "FILE", "sha256": strings.Repeat("0", 64), "size_bytes": -1},
		{"path": "node", "kind": "FILE", "sha256": strings.Repeat("0", 64), "size_bytes": nil},
		{"path": "node", "kind": "FILE", "sha256": strings.Repeat("0", 64), "size_bytes": 0.5},
		{"path": "node", "kind": "FILE", "sha256": strings.Repeat("0", 64), "size_bytes": 0},
		{"path": "node", "kind": "FILE", "sha256": "bad", "size_bytes": 0},
	} {
		assertManifestRejected(t, manifestFixture(t, []any{entry}, []any{}), map[string][]byte{"node": {}})
	}
	// Empty bytes are distinct from an absent FILE capture.
	file := map[string]any{"path": "empty", "kind": "FILE", "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "size_bytes": 0}
	raw := manifestFixture(t, []any{file}, []any{})
	assertManifestRejected(t, raw, nil)
	assertManifestRejected(t, raw, map[string][]byte{"empty": []byte("changed")})
	if _, _, err := VerifySourceManifest(raw, map[string][]byte{"empty": {}}); err != nil {
		t.Fatal(err)
	}
}

func assertManifestRejected(t *testing.T, raw []byte, files map[string][]byte) {
	t.Helper()
	root, manifest, err := VerifySourceManifest(raw, files)
	var invalid *SourceManifestError
	if !errors.As(err, &invalid) || root != "" || manifest != "" {
		t.Fatalf("expected structured rejection with no digests, got %q %q %v", root, manifest, err)
	}
}

func TestSourceManifestBinaryBytes(t *testing.T) {
	t.Parallel()
	file := map[string]any{
		"path": "binary", "kind": "FILE", "size_bytes": 4,
		"sha256": "e9489f37fb3051e9efa1dc916004d7274e7b63975e3209708947267f2393a9be",
	}
	raw := manifestFixture(t, []any{file}, []any{})
	if _, _, err := VerifySourceManifest(raw, map[string][]byte{"binary": {0, 255, 13, 10}}); err != nil {
		t.Fatal(err)
	}
	// A same-length binary mutation must fail independently of text encoding.
	assertManifestRejected(t, raw, map[string][]byte{"binary": {0, 254, 13, 10}})
}
