package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestExecutionManifestRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/execution-manifest.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExecutionManifest(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid manifest for %s, got: %v", fixture.Name, err)
				}
				return
			}
			reasons := map[string]string{
				"reject-unpinned-engine-schemas":       "schema references",
				"binding-execution-incomplete-package": "package reference object",
			}
			assertExecutionManifestRejected(t, err, reasons[fixture.Name])
		})
	}
	if count != 4 {
		t.Fatalf("expected 4 execution manifest fixtures, found %d", count)
	}
}

func executionManifestV10Control() map[string]any {
	return map[string]any{
		"schema_version":           "1.0",
		"manifest_id":              "engine",
		"project_id":               "fixture",
		"engine_version":           "0.0.0-fixture",
		"engine_binary_ref":        "sha256:" + strings.Repeat("a", 64),
		"build_environment_ref":    "sha256:" + strings.Repeat("a", 64),
		"dependency_lock_ref":      "sha256:" + strings.Repeat("a", 64),
		"schema_refs":              []any{"sha256:" + strings.Repeat("a", 64)},
		"canonicalization_profile": "rwt-cjson-1",
		"capability_ids":           []any{"canonical.rwt-cjson-1"},
		"qualification_refs":       []any{"sha256:" + strings.Repeat("a", 64)},
	}
}

func executionManifestV11Control() map[string]any {
	record := executionManifestV10Control()
	record["schema_version"] = "1.1"
	record["package_refs"] = []any{
		map[string]any{
			"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
			"content_ref":    "sha256:" + strings.Repeat("b", 64),
		},
	}
	return record
}

func validateExecutionManifestControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateExecutionManifest(raw)
}

func assertExecutionManifestRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *ExecutionManifestError
	if reason == "" || !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted execution manifest rejection %q, got %v", reason, err)
	}
	if err.Error() != "invalid execution manifest: "+reason {
		t.Fatal("error disclosed manifest content")
	}
}

func TestExecutionManifestControlsValid(t *testing.T) {
	t.Parallel()
	if err := validateExecutionManifestControl(t, executionManifestV10Control()); err != nil {
		t.Fatalf("expected v1.0 control to be valid, got: %v", err)
	}
	if err := validateExecutionManifestControl(t, executionManifestV11Control()); err != nil {
		t.Fatalf("expected v1.1 control to be valid, got: %v", err)
	}
}

func TestExecutionManifestFieldRejectionsV10(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, field, reason string
		value               any
	}{
		{"version_null", "schema_version", "unsupported schema version", nil},
		{"version_unknown", "schema_version", "unsupported schema version", "2.0"},
		{"version_number", "schema_version", "unsupported schema version", 1},
		{"manifest_id_null", "manifest_id", "manifest_id identity", nil},
		{"manifest_id_empty", "manifest_id", "manifest_id identity", ""},
		{"manifest_id_caps", "manifest_id", "manifest_id identity", "Engine"},
		{"manifest_id_space", "manifest_id", "manifest_id identity", "eng ine"},
		{"manifest_id_special", "manifest_id", "manifest_id identity", "eng@ine"},
		{"project_id_null", "project_id", "project_id identity", nil},
		{"project_id_empty", "project_id", "project_id identity", ""},
		{"project_id_caps", "project_id", "project_id identity", "Fixture"},
		{"project_id_special", "project_id", "project_id identity", "proj#1"},
		{"engine_version_null", "engine_version", "engine version", nil},
		{"engine_version_empty", "engine_version", "engine version", ""},
		{"engine_version_number", "engine_version", "engine version", 123},
		{"binary_ref_null", "engine_binary_ref", "engine_binary_ref reference", nil},
		{"binary_ref_bad", "engine_binary_ref", "engine_binary_ref reference", "bad"},
		{"binary_ref_caps", "engine_binary_ref", "engine_binary_ref reference", "sha256:" + strings.Repeat("A", 64)},
		{"binary_ref_short", "engine_binary_ref", "engine_binary_ref reference", "sha256:" + strings.Repeat("a", 63)},
		{"build_env_ref_bad", "build_environment_ref", "build_environment_ref reference", "bad"},
		{"dep_lock_ref_bad", "dependency_lock_ref", "dependency_lock_ref reference", "bad"},
		{"canonicalization_null", "canonicalization_profile", "canonicalization profile", nil},
		{"canonicalization_bad", "canonicalization_profile", "canonicalization profile", "rwt-cjson-2"},
		{"schema_refs_null", "schema_refs", "schema references", nil},
		{"schema_refs_empty", "schema_refs", "schema references", []any{}},
		{"schema_refs_item_null", "schema_refs", "schema references item", []any{nil}},
		{"schema_refs_item_bad", "schema_refs", "schema references item", []any{"bad"}},
		{"schema_refs_duplicate", "schema_refs", "duplicate schema references", []any{
			"sha256:" + strings.Repeat("a", 64),
			"sha256:" + strings.Repeat("a", 64),
		}},
		{"capability_ids_null", "capability_ids", "capability IDs", nil},
		{"capability_ids_empty", "capability_ids", "capability IDs", []any{}},
		{"capability_ids_item_null", "capability_ids", "capability ID item", []any{nil}},
		{"capability_ids_item_empty", "capability_ids", "capability ID item", []any{""}},
		{"capability_ids_duplicate", "capability_ids", "duplicate capability ID", []any{"cap1", "cap1"}},
		{"qualification_refs_null", "qualification_refs", "qualification references", nil},
		{"qualification_refs_empty", "qualification_refs", "qualification references", []any{}},
		{"qualification_refs_item_null", "qualification_refs", "qualification references item", []any{nil}},
		{"qualification_refs_item_bad", "qualification_refs", "qualification references item", []any{"bad"}},
		{"qualification_refs_duplicate", "qualification_refs", "duplicate qualification references", []any{
			"sha256:" + strings.Repeat("a", 64),
			"sha256:" + strings.Repeat("a", 64),
		}},
		{"extra_property_v10", "package_refs", "object fields", []any{}},
		{"unexpected_property", "unknown_field", "object fields", true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			record := executionManifestV10Control()
			record[tc.field] = tc.value
			assertExecutionManifestRejected(t, validateExecutionManifestControl(t, record), tc.reason)
		})
	}

	for field := range executionManifestV10Control() {
		field := field
		t.Run("missing_"+field, func(t *testing.T) {
			t.Parallel()
			record := executionManifestV10Control()
			delete(record, field)
			assertExecutionManifestRejected(t, validateExecutionManifestControl(t, record), "object fields")
		})
	}
}

func TestExecutionManifestFieldRejectionsV11(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, reason string
		modify       func(map[string]any)
	}{
		{"missing_package_refs", "object fields", func(m map[string]any) {
			delete(m, "package_refs")
		}},
		{"package_refs_null", "package references", func(m map[string]any) {
			m["package_refs"] = nil
		}},
		{"package_refs_empty", "package references", func(m map[string]any) {
			m["package_refs"] = []any{}
		}},
		{"package_refs_item_not_object", "package reference object", func(m map[string]any) {
			m["package_refs"] = []any{"bad"}
		}},
		{"package_refs_item_extra_fields", "package reference object", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
					"content_ref":    "sha256:" + strings.Repeat("b", 64),
					"extra":          true,
				},
			}
		}},
		{"package_refs_item_missing_descriptor", "package reference object", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"content_ref": "sha256:" + strings.Repeat("b", 64),
				},
			}
		}},
		{"package_refs_item_missing_content", "package reference object", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
				},
			}
		}},
		{"package_refs_item_replaced_field", "package reference object fields", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
					"other_ref":      "sha256:" + strings.Repeat("b", 64),
				},
			}
		}},
		{"package_refs_item_bad_descriptor_ref", "package reference artifact", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"descriptor_ref": "bad",
					"content_ref":    "sha256:" + strings.Repeat("b", 64),
				},
			}
		}},
		{"package_refs_item_bad_content_ref", "package reference artifact", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
					"content_ref":    "bad",
				},
			}
		}},
		{"package_refs_duplicate_items", "duplicate package reference", func(m map[string]any) {
			m["package_refs"] = []any{
				map[string]any{
					"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
					"content_ref":    "sha256:" + strings.Repeat("b", 64),
				},
				map[string]any{
					"descriptor_ref": "sha256:" + strings.Repeat("a", 64),
					"content_ref":    "sha256:" + strings.Repeat("b", 64),
				},
			}
		}},
		{"unexpected_property_v11", "object fields", func(m map[string]any) {
			m["extra_property"] = true
		}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			record := executionManifestV11Control()
			tc.modify(record)
			assertExecutionManifestRejected(t, validateExecutionManifestControl(t, record), tc.reason)
		})
	}
}

func TestExecutionManifestCanonicalJSONRejections(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"float_number", []byte(`{"schema_version":"1.0","val":1.5}`)},
		{"duplicate_keys", []byte(`{"schema_version":"1.0","schema_version":"1.0"}`)},
		{"invalid_utf8", []byte{0xff, 0xfe}},
		{"trailing_data", []byte(`{"schema_version":"1.0"}{}`)},
		{"empty_input", []byte(``)},
		{"null_input", []byte(`null`)},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExecutionManifest(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var manifestErr *ExecutionManifestError
			if !errors.As(err, &manifestErr) {
				t.Fatalf("expected ExecutionManifestError, got %T", err)
			}
		})
	}
}
