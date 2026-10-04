package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func filesystemProfileFixtures(t *testing.T) []struct {
	Name     string          `json:"name"`
	Schema   string          `json:"schema"`
	Instance json.RawMessage `json:"instance"`
	Valid    bool            `json:"valid"`
} {
	t.Helper()
	var fixtures []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	loadVectors(t, "../../specs/examples/validation-cases.json", &fixtures)
	return fixtures
}

func TestFilesystemProfileRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	for _, fixture := range filesystemProfileFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/filesystem-profile.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateFilesystemProfile(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			reason := map[string]string{
				"review2-reject-duplicate-property-gap": "duplicate property",
				"review2-required-mtime-needs-capture":  "capture or mechanism reference",
			}[fixture.Name]
			assertProfileRejected(t, err, reason)
		})
	}
	if count != 3 {
		t.Fatal("profile fixture corpus changed; review coverage")
	}
}

func profileFixture(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range filesystemProfileFixtures(t) {
		if fixture.Name == "review2-bounded-filesystem-profile" {
			var profile map[string]any
			if err := json.Unmarshal(fixture.Instance, &profile); err != nil {
				t.Fatal(err)
			}
			return profile
		}
	}
	t.Fatal("missing authoritative positive profile fixture")
	return nil
}

func validateProfileFixture(t *testing.T, profile map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateFilesystemProfile(raw)
}

func assertProfileRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *FilesystemProfileError
	if reason == "" || !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted structured rejection %q; got %v", reason, err)
	}
}

func TestFilesystemProfileTreatmentAndShape(t *testing.T) {
	t.Parallel()
	for _, treatment := range []string{"REQUIRED", "VIRTUALIZED", "IRRELEVANT", "UNSUPPORTED"} {
		t.Run(treatment, func(t *testing.T) {
			t.Parallel()
			profile := profileFixture(t)
			property := profile["properties"].([]any)[0].(map[string]any)
			property["treatment"] = treatment
			if err := validateProfileFixture(t, profile); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"captured_state_ref", "mechanism_ref"} {
				old := property[field]
				property[field] = nil
				err := validateProfileFixture(t, profile)
				if treatment == "REQUIRED" || treatment == "VIRTUALIZED" {
					assertProfileRejected(t, err, "capture or mechanism reference")
				} else if err != nil {
					t.Fatal(err)
				}
				property[field] = old
			}
		})
	}
	// The schema does not sort properties, require controls for record storage,
	// forbid repeated evidence references or forbid empty limitation strings.
	profile := profileFixture(t)
	properties := profile["properties"].([]any)
	properties[0], properties[9] = properties[9], properties[0]
	profile["control_refs"] = []any{}
	profile["limitations"] = []any{""}
	property := properties[0].(map[string]any)
	refs := property["evidence_refs"].([]any)
	property["evidence_refs"] = append(refs, refs[0])
	if err := validateProfileFixture(t, profile); err != nil {
		t.Fatal(err)
	}
}

func TestFilesystemProfileFieldRejections(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		field, reason string
		value         any
	}{
		{"schema_version", "unsupported schema version", "2.0"},
		{"project_id", "project identity", "../private"},
		{"project_id", "project identity", ""},
		{"project_id", "project identity", "UPPER"},
		{"content_manifest_ref", "content manifest reference", nil},
		{"content_manifest_ref", "content manifest reference", "sha256:" + strings.Repeat("A", 64)},
		{"properties", "ten properties required", nil},
		{"properties", "ten properties required", []any{}},
		{"control_refs", "array cardinality or type", nil},
		{"control_refs", "array item type or reference", []any{nil}},
		{"control_refs", "array item type or reference", []any{"unresolved"}},
		{"limitations", "array cardinality or type", nil},
		{"limitations", "array item type or reference", []any{nil}},
		{"limitations", "array item type or reference", []any{42}},
		{"extra", "object fields", true},
	} {
		t.Run(test.field+"/"+test.reason, func(t *testing.T) {
			t.Parallel()
			profile := profileFixture(t)
			profile[test.field] = test.value
			assertProfileRejected(t, validateProfileFixture(t, profile), test.reason)
		})
	}
	for _, field := range []string{"schema_version", "project_id", "content_manifest_ref", "properties", "control_refs", "limitations"} {
		profile := profileFixture(t)
		delete(profile, field)
		assertProfileRejected(t, validateProfileFixture(t, profile), "object fields")
	}
}

func TestFilesystemProfilePropertyRejections(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		field, reason string
		value         any
	}{
		{"property", "unsupported property", "UNKNOWN"},
		{"treatment", "unsupported treatment", nil},
		{"treatment", "unsupported treatment", "ASSUMED"},
		{"captured_state_ref", "capture or mechanism reference", 42},
		{"mechanism_ref", "capture or mechanism reference", "sha256:bad"},
		{"evidence_refs", "array cardinality or type", []any{}},
		{"evidence_refs", "array cardinality or type", nil},
		{"evidence_refs", "array item type or reference", []any{nil}},
		{"extra", "object fields", true},
	} {
		profile := profileFixture(t)
		property := profile["properties"].([]any)[0].(map[string]any)
		property[test.field] = test.value
		assertProfileRejected(t, validateProfileFixture(t, profile), test.reason)
	}
	for _, field := range []string{"property", "treatment", "captured_state_ref", "mechanism_ref", "evidence_refs"} {
		profile := profileFixture(t)
		delete(profile["properties"].([]any)[0].(map[string]any), field)
		assertProfileRejected(t, validateProfileFixture(t, profile), "object fields")
	}
	for _, raw := range []string{`{"schema_version":"1.0","schema_version":"1.0"}`, `{"secret":"\ud800"}`, `{} {}`} {
		assertProfileRejected(t, ValidateFilesystemProfile([]byte(raw)), "canonical JSON")
	}
	assertProfileRejected(t, ValidateFilesystemProfile([]byte(`null`)), "object fields")
}
