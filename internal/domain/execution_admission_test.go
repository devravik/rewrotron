package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type admissionFixture struct {
	Name     string          `json:"name"`
	Schema   string          `json:"schema"`
	Instance json.RawMessage `json:"instance"`
	Valid    bool            `json:"valid"`
}

func admissionFixtures(t *testing.T) []admissionFixture {
	t.Helper()
	var fixtures []admissionFixture
	loadVectors(t, "../../specs/examples/validation-cases.json", &fixtures)
	return fixtures
}

func TestExecutionAdmissionRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/execution-admission.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExecutionAdmission(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			reason := map[string]string{
				"review2-live-credential-original-refused": "unsafe original credential status",
				"review2-unsafe-variant-refused":           "unsafe variant credential status",
			}[fixture.Name]
			assertAdmissionRejected(t, err, reason)
		})
	}
	if count != 4 {
		t.Fatal("admission fixture corpus changed; review coverage")
	}
}

func newAdmissionFixture(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "review2-static-assessment-credential-refusal" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing authoritative admission control")
	return nil
}

func validateAdmissionFixture(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateExecutionAdmission(raw)
}

func assertAdmissionRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *ExecutionAdmissionError
	if reason == "" || !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted admission rejection %q; got %v", reason, err)
	}
}

func TestExecutionAdmissionCredentialMatrix(t *testing.T) {
	t.Parallel()
	for _, outcome := range []string{"ADMITTED_ORIGINAL", "ADMITTED_VARIANT", "REFUSED", "STATIC_ONLY"} {
		for _, status := range []string{"SYNTHETIC_OR_ABSENT", "REVOKED", "LIVE_OR_UNKNOWN"} {
			for _, variantStatus := range []any{nil, "SYNTHETIC_OR_ABSENT", "REVOKED", "LIVE_OR_UNKNOWN"} {
				record := newAdmissionFixture(t)
				record["outcome"], record["credential_status"] = outcome, status
				record["variant_credential_status"] = variantStatus
				if outcome == "ADMITTED_VARIANT" {
					record["variant_ref"] = "sha256:" + strings.Repeat("b", 64)
				}
				reason := ""
				switch {
				case outcome == "ADMITTED_VARIANT" && (variantStatus == nil || variantStatus == "LIVE_OR_UNKNOWN"):
					reason = "unsafe variant credential status"
				case outcome != "ADMITTED_VARIANT" && variantStatus != nil:
					reason = "variant fields must be null"
				case outcome == "ADMITTED_ORIGINAL" && status == "LIVE_OR_UNKNOWN":
					reason = "unsafe original credential status"
				}
				err := validateAdmissionFixture(t, record)
				if reason != "" {
					assertAdmissionRejected(t, err, reason)
				} else if err != nil {
					t.Fatalf("outcome %s original %s variant %v: %v", outcome, status, variantStatus, err)
				}
			}
		}
	}
}

func TestExecutionAdmissionRecordRejections(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		field, reason string
		value         any
	}{
		{"schema_version", "unsupported schema version", "2.0"},
		{"project_id", "project identity", nil},
		{"project_id", "project identity", "private/path"},
		{"project_id", "project identity", "Fixture"},
		{"original_manifest_ref", "manifest or policy reference", nil},
		{"policy_ref", "manifest or policy reference", "sha256:" + strings.Repeat("A", 64)},
		{"credential_status", "original credential status", nil},
		{"credential_status", "original credential status", "APPROVED"},
		{"outcome", "unsupported outcome", true},
		{"outcome", "unsupported outcome", "PASS"},
		{"variant_ref", "variant fields must be null", "sha256:" + strings.Repeat("b", 64)},
		{"variant_ref", "variant fields must be null", ""},
		{"evidence_refs", "array cardinality or type", []any{}},
		{"evidence_refs", "array cardinality or type", nil},
		{"evidence_refs", "array item type or reference", []any{nil}},
		{"evidence_refs", "array item type or reference", []any{"bad"}},
		{"limitations", "array cardinality or type", nil},
		{"limitations", "array item type or reference", []any{nil}},
		{"limitations", "array item type or reference", []any{42}},
		{"extra", "object fields", true},
	} {
		t.Run(test.field+"/"+test.reason, func(t *testing.T) {
			t.Parallel()
			record := newAdmissionFixture(t)
			record[test.field] = test.value
			assertAdmissionRejected(t, validateAdmissionFixture(t, record), test.reason)
		})
	}
	for field := range newAdmissionFixture(t) {
		record := newAdmissionFixture(t)
		delete(record, field)
		assertAdmissionRejected(t, validateAdmissionFixture(t, record), "object fields")
	}
	// Same field count cannot conceal an unknown field replacing a required one.
	record := newAdmissionFixture(t)
	delete(record, "variant_ref")
	record["extra"] = nil
	assertAdmissionRejected(t, validateAdmissionFixture(t, record), "missing required field")
	for _, raw := range []string{`{"outcome":"REFUSED","outcome":"STATIC_ONLY"}`, `{"private":"\ud800"}`, `{} {}`} {
		assertAdmissionRejected(t, ValidateExecutionAdmission([]byte(raw)), "canonical JSON")
	}
	assertAdmissionRejected(t, ValidateExecutionAdmission([]byte(`null`)), "object fields")
}

func TestExecutionAdmissionVariantReference(t *testing.T) {
	t.Parallel()
	for _, ref := range []any{nil, "", "sha256:bad", true} {
		record := newAdmissionFixture(t)
		record["outcome"] = "ADMITTED_VARIANT"
		record["variant_credential_status"] = "REVOKED"
		record["variant_ref"] = ref
		assertAdmissionRejected(t, validateAdmissionFixture(t, record), "variant reference required")
	}
	// Schema-valid evidence references are not verified artifacts or authority.
	record := newAdmissionFixture(t)
	record["project_id"] = "0_fixture-1"
	record["limitations"] = []any{""}
	refs := record["evidence_refs"].([]any)
	record["evidence_refs"] = append(refs, refs[0])
	if err := validateAdmissionFixture(t, record); err != nil {
		t.Fatal(err)
	}
}
