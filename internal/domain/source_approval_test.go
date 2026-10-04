package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func sourceApprovalControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "source-subject-historical-shape" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing SOURCE control")
	return nil
}

func validateSourceApprovalControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateSourceApprovalSubjectShape(raw)
}

func assertSourceApprovalRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *SourceApprovalError
	if !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted %q, got %v", reason, err)
	}
	if err.Error() != "invalid source approval subject: "+reason {
		t.Fatal("error disclosed subject data")
	}
}

func TestSourceApprovalGoldenAndFixtures(t *testing.T) {
	var vectors []struct {
		Name      string          `json:"name"`
		Subject   json.RawMessage `json:"subject"`
		Canonical string          `json:"canonical_utf8"`
		Hash      string          `json:"sha256"`
	}
	loadVectors(t, "../../specs/examples/approval-hash-vectors.json", &vectors)
	count := 0
	for _, vector := range vectors {
		if vector.Name != "source" {
			continue
		}
		count++
		before := bytes.Clone(vector.Subject)
		if err := ValidateSourceApprovalSubjectShape(vector.Subject); err != nil {
			t.Fatal(err)
		}
		assertGolden(t, vector.Subject, vector.Canonical, vector.Hash)
		if !bytes.Equal(before, vector.Subject) {
			t.Fatal("shape validation modified historical subject")
		}
	}
	if count != 1 {
		t.Fatal("SOURCE golden corpus changed")
	}
	count = 0
	for _, fixture := range admissionFixtures(t) {
		if !strings.HasPrefix(fixture.Name, "source-subject-") {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			err := ValidateSourceApprovalSubjectShape(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				reason := "artifact reference"
				if fixture.Name == "source-subject-duplicate-binding" {
					reason = "binding order or duplicate"
				}
				assertSourceApprovalRejected(t, err, reason)
			}
		})
	}
	if count != 3 {
		t.Fatal("SOURCE shape fixture corpus changed")
	}
}

func TestSourceApprovalShapeRejections(t *testing.T) {
	for _, tc := range []struct {
		field, reason string
		value         any
	}{
		{"schema_version", "unsupported subject identity", "1.0"},
		{"canonicalization", "unsupported subject identity", "other"},
		{"kind", "unsupported subject identity", "PLAN"},
		{"kind", "unsupported subject identity", "RELEASE"},
		{"project_id", "project identity", "private/path"},
		{"project_id", "project identity", nil},
		{"snapshot_ref", "artifact reference", true},
		{"detection_ref", "artifact reference", "sha256:" + strings.Repeat("A", 64)},
		{"assertions_ref", "artifact reference", "bad"},
		{"adapter_bindings", "binding array", nil},
		{"adapter_bindings", "binding array", []any{}},
		{"extra", "object fields", true},
	} {
		t.Run(tc.field+"/"+tc.reason, func(t *testing.T) {
			record := sourceApprovalControl(t)
			record[tc.field] = tc.value
			assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), tc.reason)
		})
	}
	for field := range sourceApprovalControl(t) {
		record := sourceApprovalControl(t)
		delete(record, field)
		assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), "object fields")
	}
	record := sourceApprovalControl(t)
	delete(record, "kind")
	record["unexpected"] = "SOURCE"
	assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), "missing required field")
	for _, raw := range []string{`null`, `[]`, `"private"`} {
		assertSourceApprovalRejected(t, ValidateSourceApprovalSubjectShape([]byte(raw)), "object fields")
	}
	for _, raw := range []string{`{"kind":"SOURCE","kind":"SOURCE"}`, `{"private":"\ud800"}`, `{} {}`} {
		err := ValidateSourceApprovalSubjectShape([]byte(raw))
		assertSourceApprovalRejected(t, err, "canonical JSON")
		if errors.Unwrap(err) == nil {
			t.Fatal("lost canonicalization cause")
		}
	}
}

func TestSourceApprovalBindingContract(t *testing.T) {
	for _, tc := range []struct {
		field, reason string
		value         any
	}{
		{"role", "binding role", "AGENT"},
		{"role", "binding role", nil},
		{"adapter_id", "adapter identity", "Synthetic"},
		{"adapter_id", "adapter identity", nil},
		{"version", "adapter version", "latest"},
		{"version", "adapter version", "1.0.0+build"},
		{"version", "adapter version", nil},
		{"manifest_ref", "binding artifact reference", nil},
		{"implementation_ref", "binding artifact reference", "bad"},
		{"extra", "object fields", true},
	} {
		record := sourceApprovalControl(t)
		binding := record["adapter_bindings"].([]any)[0].(map[string]any)
		binding[tc.field] = tc.value
		assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), tc.reason)
	}
	for _, version := range []string{"1.0.0", "0.0.0-fixture", "01.2.3-a.."} {
		record := sourceApprovalControl(t)
		record["adapter_bindings"].([]any)[0].(map[string]any)["version"] = version
		if err := validateSourceApprovalControl(t, record); err != nil {
			t.Fatal("schema version grammar tightened by inference", err)
		}
	}
	for _, role := range []string{"SOURCE", "TARGET", "PROBE", "COMPARATOR", "RUNNER"} {
		record := sourceApprovalControl(t)
		record["adapter_bindings"].([]any)[0].(map[string]any)["role"] = role
		if err := validateSourceApprovalControl(t, record); err != nil {
			t.Fatal("shape validator invented role eligibility", err)
		}
	}
	for _, item := range []any{nil, "private", map[string]any{}} {
		record := sourceApprovalControl(t)
		record["adapter_bindings"] = []any{item}
		assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), "object fields")
	}
	for _, field := range []string{"role", "adapter_id", "version", "manifest_ref", "implementation_ref"} {
		record := sourceApprovalControl(t)
		binding := record["adapter_bindings"].([]any)[0].(map[string]any)
		delete(binding, field)
		assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), "object fields")
	}
}

func TestSourceApprovalBindingTupleOrder(t *testing.T) {
	for _, tc := range []struct{ field, first, second string }{
		{"role", "PROBE", "SOURCE"},
		{"adapter_id", "a", "b"},
		{"version", "1.10.0", "1.2.0"}, // Byte order, not numeric release order.
		{"manifest_ref", "sha256:" + strings.Repeat("a", 64), "sha256:" + strings.Repeat("b", 64)},
		{"implementation_ref", "sha256:" + strings.Repeat("a", 64), "sha256:" + strings.Repeat("b", 64)},
	} {
		t.Run(tc.field, func(t *testing.T) {
			record := sourceApprovalControl(t)
			first := record["adapter_bindings"].([]any)[0].(map[string]any)
			second := sourceApprovalControl(t)["adapter_bindings"].([]any)[0].(map[string]any)
			first[tc.field], second[tc.field] = tc.first, tc.second
			record["adapter_bindings"] = []any{first, second}
			if err := validateSourceApprovalControl(t, record); err != nil {
				t.Fatal(err)
			}
			record["adapter_bindings"] = []any{second, first}
			assertSourceApprovalRejected(t, validateSourceApprovalControl(t, record), "binding order or duplicate")
		})
	}
}
