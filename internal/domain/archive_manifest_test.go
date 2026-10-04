package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestArchiveManifestRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/operations/archive.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateArchiveManifest(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid archive manifest for %s, got: %v", fixture.Name, err)
				}
				return
			}
			t.Fatalf("unexpected invalid fixture in repository: %s", fixture.Name)
		})
	}
	if count != 1 {
		t.Fatalf("expected 1 archive manifest fixture, found %d", count)
	}
}

func archiveManifestControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "v2-archive" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing v2-archive fixture")
	return nil
}

func assertArchiveManifestRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var archErr *ArchiveManifestError
	if !errors.As(err, &archErr) || archErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid archive manifest: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestArchiveManifestPositiveControls(t *testing.T) {
	t.Parallel()

	t.Run("fixture-control", func(t *testing.T) {
		t.Parallel()
		ctrl := archiveManifestControl(t)
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateArchiveManifest(raw); err != nil {
			t.Fatalf("expected valid archive manifest, got: %v", err)
		}
	})

	t.Run("empty-objects", func(t *testing.T) {
		t.Parallel()
		ctrl := archiveManifestControl(t)
		ctrl["objects"] = []any{}
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateArchiveManifest(raw); err != nil {
			t.Fatalf("expected valid archive manifest with empty objects, got: %v", err)
		}
	})

	t.Run("multiple-sorted-objects", func(t *testing.T) {
		t.Parallel()
		ctrl := archiveManifestControl(t)
		ctrl["objects"] = []any{
			map[string]any{
				"ref":             "sha256:" + strings.Repeat("1", 64),
				"size_bytes":      100,
				"retention_class": "APPROVAL",
				"private":         false,
			},
			map[string]any{
				"ref":             "sha256:" + strings.Repeat("2", 64),
				"size_bytes":      200,
				"retention_class": "CONTRACT",
				"private":         true,
			},
			map[string]any{
				"ref":             "sha256:" + strings.Repeat("a", 64),
				"size_bytes":      300,
				"retention_class": "SNAPSHOT",
				"private":         true,
			},
		}
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateArchiveManifest(raw); err != nil {
			t.Fatalf("expected valid archive manifest with multiple sorted objects, got: %v", err)
		}
	})
}

func TestArchiveManifestNegativeMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]any)
		reason string
	}{
		{
			name: "wrong-schema-version",
			mutate: func(m map[string]any) {
				m["schema_version"] = "1.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "missing-project-id",
			mutate: func(m map[string]any) {
				delete(m, "project_id")
			},
			reason: "object fields",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "Invalid_Project!"
			},
			reason: "project identity",
		},
		{
			name: "invalid-db-hash",
			mutate: func(m map[string]any) {
				m["db_hash"] = "not-a-hex-hash"
			},
			reason: "database digest",
		},
		{
			name: "negative-through-sequence",
			mutate: func(m map[string]any) {
				m["through_sequence"] = int64(-1)
			},
			reason: "through_sequence",
		},
		{
			name: "empty-migration-hashes",
			mutate: func(m map[string]any) {
				m["migration_hashes"] = []any{}
			},
			reason: "migration hashes array",
		},
		{
			name: "invalid-migration-hash-format",
			mutate: func(m map[string]any) {
				m["migration_hashes"] = []any{"bad-hash"}
			},
			reason: "migration hash format",
		},
		{
			name: "duplicate-migration-hash",
			mutate: func(m map[string]any) {
				h := strings.Repeat("b", 64)
				m["migration_hashes"] = []any{h, h}
			},
			reason: "duplicate migration hash",
		},
		{
			name: "object-entry-missing-field",
			mutate: func(m map[string]any) {
				m["objects"] = []any{
					map[string]any{
						"ref": "sha256:" + strings.Repeat("a", 64),
					},
				}
			},
			reason: "object entry fields",
		},
		{
			name: "object-invalid-ref-format",
			mutate: func(m map[string]any) {
				objs := m["objects"].([]any)
				o0 := objs[0].(map[string]any)
				o0["ref"] = "bad-ref"
			},
			reason: "object reference",
		},
		{
			name: "objects-unsorted",
			mutate: func(m map[string]any) {
				m["objects"] = []any{
					map[string]any{
						"ref":             "sha256:" + strings.Repeat("b", 64),
						"size_bytes":      10,
						"retention_class": "APPROVAL",
						"private":         false,
					},
					map[string]any{
						"ref":             "sha256:" + strings.Repeat("a", 64),
						"size_bytes":      10,
						"retention_class": "APPROVAL",
						"private":         false,
					},
				}
			},
			reason: "object entries must be strictly sorted by reference",
		},
		{
			name: "objects-duplicate-ref",
			mutate: func(m map[string]any) {
				m["objects"] = []any{
					map[string]any{
						"ref":             "sha256:" + strings.Repeat("a", 64),
						"size_bytes":      10,
						"retention_class": "APPROVAL",
						"private":         false,
					},
					map[string]any{
						"ref":             "sha256:" + strings.Repeat("a", 64),
						"size_bytes":      20,
						"retention_class": "CONTRACT",
						"private":         true,
					},
				}
			},
			reason: "object entries must be strictly sorted by reference",
		},
		{
			name: "object-negative-size",
			mutate: func(m map[string]any) {
				objs := m["objects"].([]any)
				o0 := objs[0].(map[string]any)
				o0["size_bytes"] = int64(-1)
			},
			reason: "object size_bytes",
		},
		{
			name: "object-unsupported-retention-class",
			mutate: func(m map[string]any) {
				objs := m["objects"].([]any)
				o0 := objs[0].(map[string]any)
				o0["retention_class"] = "UNKNOWN_CLASS"
			},
			reason: "unsupported retention class",
		},
		{
			name: "object-private-not-bool",
			mutate: func(m map[string]any) {
				objs := m["objects"].([]any)
				o0 := objs[0].(map[string]any)
				o0["private"] = "false"
			},
			reason: "object private flag",
		},
		{
			name: "invalid-config-ref",
			mutate: func(m map[string]any) {
				m["config_ref"] = "bad-ref"
			},
			reason: "config_ref",
		},
		{
			name: "invalid-private-key-policy-ref",
			mutate: func(m map[string]any) {
				m["private_key_policy_ref"] = "bad-ref"
			},
			reason: "private_key_policy_ref",
		},
		{
			name: "invalid-closure-report-ref",
			mutate: func(m map[string]any) {
				m["closure_report_ref"] = "bad-ref"
			},
			reason: "closure_report_ref",
		},
		{
			name: "extra-top-level-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := archiveManifestControl(t)
			tc.mutate(ctrl)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateArchiveManifest(raw)
			assertArchiveManifestRejected(t, err, tc.reason)
		})
	}
}

func TestArchiveManifestCanonicalJSON(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"float_number", []byte(`{"schema_version":"2.0","num":1.5}`)},
		{"duplicate_keys", []byte(`{"schema_version":"2.0","schema_version":"2.0"}`)},
		{"invalid_utf8", []byte{0xff, 0xfe}},
		{"trailing_data", []byte(`{"schema_version":"2.0"}{}`)},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateArchiveManifest(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var archErr *ArchiveManifestError
			if !errors.As(err, &archErr) || archErr.Reason != "canonical JSON" {
				t.Fatalf("expected canonical JSON error, got %v", err)
			}
		})
	}
}

func TestArchiveManifestErrorSecurity(t *testing.T) {
	t.Parallel()
	ctrl := archiveManifestControl(t)
	ctrl["project_id"] = "CANARY_SECRET_LEAK!"
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateArchiveManifest(raw)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errStr := err.Error()
	if strings.Contains(errStr, "CANARY_SECRET_LEAK") {
		t.Fatalf("error message discloses sensitive input: %s", errStr)
	}
	if errStr != "invalid archive manifest: project identity" {
		t.Fatalf("unexpected error format: %s", errStr)
	}
}
