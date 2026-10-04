package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestReuseCertificateRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasonMap := map[string]string{
		"review2-current-v2-reuse-requires-provenance": "missing required field",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/reuse-certificate.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateReuseCertificate(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid reuse certificate for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, ok := reasonMap[fixture.Name]
			if !ok {
				t.Fatalf("missing expected reason mapping for invalid fixture: %s", fixture.Name)
			}
			assertReuseCertificateRejected(t, err, expectedReason)
		})
	}
	if count != 3 {
		t.Fatalf("expected 3 reuse certificate fixtures, found %d", count)
	}
}

func reuseCertificateControl(t *testing.T, name string) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == name {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatalf("missing %s fixture", name)
	return nil
}

func assertReuseCertificateRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var reuseErr *ReuseCertificateError
	if !errors.As(err, &reuseErr) || reuseErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid reuse certificate: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestReuseCertificateValidRecords(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"v2-reuse", "review2-current-v2-reuse"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := reuseCertificateControl(t, name)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := CanonicalJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateReuseCertificate(canonical); err != nil {
				t.Fatalf("expected valid %s, got: %v", name, err)
			}
		})
	}
}

func TestReuseCertificateMutationsV20(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		reason  string
		rawJSON string
	}{
		{
			name:    "invalid-json",
			rawJSON: `{"schema_version": "2.0",`,
			reason:  "canonical JSON",
		},
		{
			name:    "non-canonical-whitespace",
			rawJSON: `{"artifact_ref":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","checked_edge_refs":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"evidence_refs":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"graph_ref":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"issuer":"ENGINE","kind":"IMPLEMENTATION","new_plan_hash":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","old_plan_hash":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","project_id":"fixture","schema_version":"2.0","valid":true} `,
			reason:  "canonical JSON",
		},
		{
			name: "extra-field-v20",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
		{
			name: "v20-cannot-have-consumption-coverage-refs",
			mutate: func(m map[string]any) {
				m["consumption_coverage_refs"] = []any{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
			},
			reason: "object fields",
		},
		{
			name: "missing-schema-version",
			mutate: func(m map[string]any) {
				delete(m, "schema_version")
			},
			reason: "missing required field",
		},
		{
			name: "unsupported-schema-version-10",
			mutate: func(m map[string]any) {
				m["schema_version"] = "1.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "unsupported-schema-version-30",
			mutate: func(m map[string]any) {
				m["schema_version"] = "3.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "missing-project-id",
			mutate: func(m map[string]any) {
				delete(m, "project_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "PROJECT_UPPER"
			},
			reason: "project identity",
		},
		{
			name: "missing-old-plan-hash",
			mutate: func(m map[string]any) {
				delete(m, "old_plan_hash")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-old-plan-hash-length",
			mutate: func(m map[string]any) {
				m["old_plan_hash"] = "short"
			},
			reason: "old_plan_hash",
		},
		{
			name: "invalid-old-plan-hash-chars",
			mutate: func(m map[string]any) {
				m["old_plan_hash"] = "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
			},
			reason: "old_plan_hash",
		},
		{
			name: "missing-new-plan-hash",
			mutate: func(m map[string]any) {
				delete(m, "new_plan_hash")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-new-plan-hash",
			mutate: func(m map[string]any) {
				m["new_plan_hash"] = "not-a-hash"
			},
			reason: "new_plan_hash",
		},
		{
			name: "missing-artifact-ref",
			mutate: func(m map[string]any) {
				delete(m, "artifact_ref")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-artifact-ref",
			mutate: func(m map[string]any) {
				m["artifact_ref"] = "no-sha-prefix"
			},
			reason: "artifact_ref",
		},
		{
			name: "missing-graph-ref",
			mutate: func(m map[string]any) {
				delete(m, "graph_ref")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-graph-ref",
			mutate: func(m map[string]any) {
				m["graph_ref"] = "sha256:short"
			},
			reason: "graph_ref",
		},
		{
			name: "missing-kind",
			mutate: func(m map[string]any) {
				delete(m, "kind")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-kind",
			mutate: func(m map[string]any) {
				m["kind"] = "VERIFICATION"
			},
			reason: "invalid kind",
		},
		{
			name: "missing-checked-edge-refs",
			mutate: func(m map[string]any) {
				delete(m, "checked_edge_refs")
			},
			reason: "missing required field",
		},
		{
			name: "checked-edge-refs-not-array",
			mutate: func(m map[string]any) {
				m["checked_edge_refs"] = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			},
			reason: "checked_edge_refs array",
		},
		{
			name: "checked-edge-refs-empty",
			mutate: func(m map[string]any) {
				m["checked_edge_refs"] = []any{}
			},
			reason: "checked_edge_refs array",
		},
		{
			name: "checked-edge-refs-invalid-ref",
			mutate: func(m map[string]any) {
				m["checked_edge_refs"] = []any{"bad-ref"}
			},
			reason: "invalid checked_edge_refs reference",
		},
		{
			name: "checked-edge-refs-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				m["checked_edge_refs"] = []any{ref, ref}
			},
			reason: "duplicate checked_edge_refs reference",
		},
		{
			name: "missing-evidence-refs",
			mutate: func(m map[string]any) {
				delete(m, "evidence_refs")
			},
			reason: "missing required field",
		},
		{
			name: "evidence-refs-not-array",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = 123
			},
			reason: "evidence_refs array",
		},
		{
			name: "evidence-refs-empty",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = []any{}
			},
			reason: "evidence_refs array",
		},
		{
			name: "evidence-refs-invalid-ref",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = []any{"bad-ref"}
			},
			reason: "invalid evidence_refs reference",
		},
		{
			name: "evidence-refs-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				m["evidence_refs"] = []any{ref, ref}
			},
			reason: "duplicate evidence_refs reference",
		},
		{
			name: "missing-issuer",
			mutate: func(m map[string]any) {
				delete(m, "issuer")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-issuer",
			mutate: func(m map[string]any) {
				m["issuer"] = "OPERATOR"
			},
			reason: "issuer must be ENGINE",
		},
		{
			name: "missing-valid",
			mutate: func(m map[string]any) {
				delete(m, "valid")
			},
			reason: "missing required field",
		},
		{
			name: "valid-false",
			mutate: func(m map[string]any) {
				m["valid"] = false
			},
			reason: "valid must be true",
		},
		{
			name: "valid-not-bool",
			mutate: func(m map[string]any) {
				m["valid"] = "true"
			},
			reason: "valid must be true",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var raw []byte
			if tc.rawJSON != "" {
				raw = []byte(tc.rawJSON)
			} else {
				ctrl := reuseCertificateControl(t, "v2-reuse")
				tc.mutate(ctrl)
				data, err := json.Marshal(ctrl)
				if err != nil {
					t.Fatal(err)
				}
				canon, err := CanonicalJSON(data)
				if err != nil {
					t.Fatal(err)
				}
				raw = canon
			}

			err := ValidateReuseCertificate(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertReuseCertificateRejected(t, err, tc.reason)
		})
	}
}

func TestReuseCertificateMutationsV21(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]any)
		reason string
	}{
		{
			name: "valid-kind-characterization",
			mutate: func(m map[string]any) {
				m["kind"] = "CHARACTERIZATION"
			},
			reason: "", // valid case
		},
		{
			name: "v21-missing-consumption-coverage-refs",
			mutate: func(m map[string]any) {
				delete(m, "consumption_coverage_refs")
			},
			reason: "missing required field",
		},
		{
			name: "v21-consumption-coverage-refs-not-array",
			mutate: func(m map[string]any) {
				m["consumption_coverage_refs"] = "bad"
			},
			reason: "consumption_coverage_refs array",
		},
		{
			name: "v21-consumption-coverage-refs-empty",
			mutate: func(m map[string]any) {
				m["consumption_coverage_refs"] = []any{}
			},
			reason: "consumption_coverage_refs array",
		},
		{
			name: "v21-consumption-coverage-refs-invalid-ref",
			mutate: func(m map[string]any) {
				m["consumption_coverage_refs"] = []any{"bad-ref"}
			},
			reason: "invalid consumption_coverage_refs reference",
		},
		{
			name: "v21-consumption-coverage-refs-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				m["consumption_coverage_refs"] = []any{ref, ref}
			},
			reason: "duplicate consumption_coverage_refs reference",
		},
		{
			name: "v21-extra-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := reuseCertificateControl(t, "review2-current-v2-reuse")
			tc.mutate(ctrl)
			data, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canon, err := CanonicalJSON(data)
			if err != nil {
				t.Fatal(err)
			}

			err = ValidateReuseCertificate(canon)
			if tc.reason == "" {
				if err != nil {
					t.Fatalf("expected valid for case %s, got: %v", tc.name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertReuseCertificateRejected(t, err, tc.reason)
		})
	}
}
