package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDecisionRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/decision.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDecision(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid decision for %s, got: %v", fixture.Name, err)
				}
				return
			}
			t.Fatalf("unexpected invalid fixture in repository: %s", fixture.Name)
		})
	}
	if count != 1 {
		t.Fatalf("expected 1 decision fixture, found %d", count)
	}
}

func decisionControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "v2-decision" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing v2-decision fixture")
	return nil
}

func assertDecisionRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var decErr *DecisionError
	if !errors.As(err, &decErr) || decErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid decision proposal: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestDecisionValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := decisionControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDecision(canonical); err != nil {
		t.Fatalf("expected valid control decision, got: %v", err)
	}
}

func TestDecisionValidWithNullStrategyAndEmptyEvidence(t *testing.T) {
	t.Parallel()
	ctrl := decisionControl(t)
	ctrl["strategy"] = nil
	ctrl["implementation_ref"] = nil
	ctrl["exception_scope_ref"] = "sha256:" + strings.Repeat("e", 64)
	ctrl["evidence_refs"] = []any{}
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDecision(canonical); err != nil {
		t.Fatalf("expected valid modified decision, got: %v", err)
	}
}

func TestDecisionMutations(t *testing.T) {
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
			name:    "non-canonical-float",
			rawJSON: `{"schema_version":"2.0","val":12.34}`,
			reason:  "canonical JSON",
		},
		{
			name: "extra-root-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
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
			name: "unsupported-schema-version",
			mutate: func(m map[string]any) {
				m["schema_version"] = "1.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "missing-decision-id",
			mutate: func(m map[string]any) {
				delete(m, "decision_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-decision-id",
			mutate: func(m map[string]any) {
				m["decision_id"] = "DEC_UPPER"
			},
			reason: "decision identity",
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
				m["project_id"] = "PROJ_UPPER"
			},
			reason: "project identity",
		},
		{
			name: "invalid-action",
			mutate: func(m map[string]any) {
				m["action"] = "EXECUTE_NOW"
			},
			reason: "invalid action",
		},
		{
			name: "missing-subject-ref",
			mutate: func(m map[string]any) {
				delete(m, "subject_ref")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-subject-ref",
			mutate: func(m map[string]any) {
				m["subject_ref"] = "bad-ref"
			},
			reason: "subject_ref",
		},
		{
			name: "invalid-strategy",
			mutate: func(m map[string]any) {
				m["strategy"] = "MAGIC"
			},
			reason: "invalid strategy",
		},
		{
			name: "invalid-implementation-ref",
			mutate: func(m map[string]any) {
				m["implementation_ref"] = "not-sha256"
			},
			reason: "implementation_ref",
		},
		{
			name: "invalid-exception-scope-ref",
			mutate: func(m map[string]any) {
				m["exception_scope_ref"] = "not-sha256"
			},
			reason: "exception_scope_ref",
		},
		{
			name: "empty-rationale",
			mutate: func(m map[string]any) {
				m["rationale"] = "   "
			},
			reason: "rationale",
		},
		{
			name: "evidence-refs-not-array",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = "not-array"
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
				ref := "sha256:" + strings.Repeat("a", 64)
				m["evidence_refs"] = []any{ref, ref}
			},
			reason: "duplicate evidence_refs reference",
		},
		{
			name: "status-not-proposed",
			mutate: func(m map[string]any) {
				m["status"] = "APPROVED"
			},
			reason: "status must be PROPOSED",
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
				ctrl := decisionControl(t)
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

			err := ValidateDecision(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertDecisionRejected(t, err, tc.reason)
		})
	}
}
