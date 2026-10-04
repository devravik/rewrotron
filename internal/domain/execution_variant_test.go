package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestExecutionVariantRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/execution-variant.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExecutionVariant(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				assertVariantRejected(t, err, "declared changes required")
			}
		})
	}
	if count != 2 {
		t.Fatal("variant fixture corpus changed; review coverage")
	}
}

func newVariantFixture(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "review2-distinct-execution-variant" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing authoritative variant control")
	return nil
}

func validateVariantFixture(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateExecutionVariant(raw)
}

func assertVariantRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *ExecutionVariantError
	if !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted variant rejection %q; got %v", reason, err)
	}
}

func TestExecutionVariantRecordRejections(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		field, reason string
		value         any
	}{
		{"schema_version", "unsupported schema version", "2.0"},
		{"project_id", "project identity", nil},
		{"project_id", "project identity", "Fixture"},
		{"project_id", "project identity", "private/path"},
		{"declared_changes", "declared changes required", nil},
		{"declared_changes", "declared changes required", []any{}},
		{"declared_changes", "declared changes required", "wrong type"},
		{"declared_changes", "object fields", []any{nil}},
		{"limitations", "nonempty array required", []any{}},
		{"limitations", "nonempty array required", nil},
		{"limitations", "array item type or reference", []any{nil}},
		{"limitations", "array item type or reference", []any{42}},
		{"review_evidence_refs", "nonempty array required", []any{}},
		{"review_evidence_refs", "nonempty array required", nil},
		{"review_evidence_refs", "array item type or reference", []any{nil}},
		{"review_evidence_refs", "array item type or reference", []any{"bad"}},
		{"extra", "object fields", true},
	} {
		t.Run(test.field+"/"+test.reason, func(t *testing.T) {
			t.Parallel()
			record := newVariantFixture(t)
			record[test.field] = test.value
			assertVariantRejected(t, validateVariantFixture(t, record), test.reason)
		})
	}
	for field := range newVariantFixture(t) {
		record := newVariantFixture(t)
		delete(record, field)
		assertVariantRejected(t, validateVariantFixture(t, record), "object fields")
	}
	record := newVariantFixture(t)
	delete(record, "variant_manifest_ref")
	record["extra"] = nil
	assertVariantRejected(t, validateVariantFixture(t, record), "missing required field")
}

func TestExecutionVariantIdentitySyntax(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"original_manifest_ref", "variant_manifest_ref", "transform_ref", "credential_policy_ref", "original_content_hash", "variant_content_hash"} {
		for _, value := range []any{nil, true, "sha256:bad", strings.Repeat("A", 64)} {
			record := newVariantFixture(t)
			record[field] = value
			reason := "artifact reference"
			if strings.HasSuffix(field, "_hash") {
				reason = "content hash"
			}
			assertVariantRejected(t, validateVariantFixture(t, record), reason)
		}
	}
}

func TestExecutionVariantChangeRejections(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"path", "before_ref", "after_ref", "reason"} {
		for _, value := range []any{nil, true} {
			record := newVariantFixture(t)
			change := record["declared_changes"].([]any)[0].(map[string]any)
			change[field] = value
			reason := "change artifact reference"
			if field == "path" || field == "reason" {
				reason = "change path or reason type"
			}
			assertVariantRejected(t, validateVariantFixture(t, record), reason)
		}
		record := newVariantFixture(t)
		delete(record["declared_changes"].([]any)[0].(map[string]any), field)
		assertVariantRejected(t, validateVariantFixture(t, record), "object fields")
	}
	record := newVariantFixture(t)
	record["declared_changes"].([]any)[0].(map[string]any)["extra"] = true
	assertVariantRejected(t, validateVariantFixture(t, record), "object fields")
	for _, raw := range []string{`{"project_id":"fixture","project_id":"other"}`, `{"private":"\ud800"}`, `{} {}`} {
		assertVariantRejected(t, ValidateExecutionVariant([]byte(raw)), "canonical JSON")
	}
	assertVariantRejected(t, ValidateExecutionVariant([]byte(`null`)), "object fields")
}

func TestExecutionVariantStructuralLimits(t *testing.T) {
	t.Parallel()
	record := newVariantFixture(t)
	record["project_id"] = "0_fixture-1"
	// Record validation cannot substitute for path admission, useful explanations,
	// independent review or byte-to-manifest verification by the guarded service.
	change := record["declared_changes"].([]any)[0].(map[string]any)
	change["path"], change["reason"] = "", ""
	record["limitations"] = []any{""}
	refs := record["review_evidence_refs"].([]any)
	record["review_evidence_refs"] = append(refs, refs[0])
	if err := validateVariantFixture(t, record); err != nil {
		t.Fatal(err)
	}
}
