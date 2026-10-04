package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestConsumptionCoverageRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasonMap := map[string]string{
		"review2-unknown-coverage-not-sufficient":    "sufficient coverage requires closed or irrelevant category status",
		"review2-unresolved-coverage-not-sufficient": "sufficient coverage requires empty unresolved candidates",
		"review2-empty-graph-not-coverage":           "categories cannot be empty",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/consumption-coverage.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConsumptionCoverage(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid consumption coverage for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, ok := reasonMap[fixture.Name]
			if !ok {
				t.Fatalf("missing expected reason mapping for invalid fixture: %s", fixture.Name)
			}
			assertConsumptionCoverageRejected(t, err, expectedReason)
		})
	}
	if count != 4 {
		t.Fatalf("expected 4 consumption coverage fixtures, found %d", count)
	}
}

func consumptionCoverageControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "review2-positive-consumption-coverage" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing review2-positive-consumption-coverage fixture")
	return nil
}

func assertConsumptionCoverageRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var covErr *ConsumptionCoverageError
	if !errors.As(err, &covErr) || covErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid consumption coverage: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestConsumptionCoverageValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := consumptionCoverageControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateConsumptionCoverage(canonical); err != nil {
		t.Fatalf("expected valid control consumption coverage, got: %v", err)
	}
}

func TestConsumptionCoverageValidInsufficientRecord(t *testing.T) {
	t.Parallel()
	ctrl := consumptionCoverageControl(t)
	ctrl["status"] = "INSUFFICIENT"
	ctrl["unresolved_candidates"] = []any{"reflection_loader"}
	ctrl["excluded_mechanisms"] = []any{"eval_evaluator"}
	ctrl["closure_control_refs"] = []any{}
	ctrl["categories"] = []any{
		map[string]any{
			"category": "SOURCE",
			"status":   "UNKNOWN",
			"input_refs": []any{
				"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			"evidence_refs": []any{
				"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		map[string]any{
			"category":      "ENVIRONMENT",
			"status":        "PARTIAL",
			"input_refs":    []any{},
			"evidence_refs": []any{"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		},
	}
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateConsumptionCoverage(canonical); err != nil {
		t.Fatalf("expected valid insufficient coverage record, got: %v", err)
	}
}

func TestConsumptionCoverageMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		reason  string
		rawJSON string
	}{
		{
			name:    "invalid-json",
			rawJSON: `{"schema_version": "1.0",`,
			reason:  "canonical JSON",
		},
		{
			name:    "non-canonical-whitespace",
			rawJSON: `{"analysis_revision_ref":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","categories":[{"category":"SOURCE","evidence_refs":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"input_refs":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"status":"CLOSED"}],"closure_control_refs":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"discovery_method_ref":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"excluded_mechanisms":[],"inputs_examined":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"issuer":"ENGINE","project_id":"fixture","roots_examined":["root"],"schema_version":"1.0","scope_ref":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","status":"SUFFICIENT","unresolved_candidates":[]} `,
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
				m["schema_version"] = "2.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "Project_ID_UpperCase"
			},
			reason: "project identity",
		},
		{
			name: "empty-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = ""
			},
			reason: "project identity",
		},
		{
			name: "invalid-scope-ref",
			mutate: func(m map[string]any) {
				m["scope_ref"] = "invalid-hash"
			},
			reason: "scope_ref",
		},
		{
			name: "null-scope-ref",
			mutate: func(m map[string]any) {
				m["scope_ref"] = nil
			},
			reason: "scope_ref",
		},
		{
			name: "invalid-discovery-method-ref",
			mutate: func(m map[string]any) {
				m["discovery_method_ref"] = "sha256:short"
			},
			reason: "discovery_method_ref",
		},
		{
			name: "invalid-analysis-revision-ref",
			mutate: func(m map[string]any) {
				m["analysis_revision_ref"] = "sha256:ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
			},
			reason: "analysis_revision_ref",
		},
		{
			name: "invalid-issuer",
			mutate: func(m map[string]any) {
				m["issuer"] = "AGENT"
			},
			reason: "issuer must be ENGINE",
		},
		{
			name: "invalid-status",
			mutate: func(m map[string]any) {
				m["status"] = "PASS"
			},
			reason: "invalid status",
		},
		{
			name: "inputs-examined-not-array",
			mutate: func(m map[string]any) {
				m["inputs_examined"] = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			},
			reason: "inputs examined array",
		},
		{
			name: "inputs-examined-invalid-ref",
			mutate: func(m map[string]any) {
				m["inputs_examined"] = []any{"bad-ref"}
			},
			reason: "invalid input examined reference",
		},
		{
			name: "inputs-examined-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				m["inputs_examined"] = []any{ref, ref}
			},
			reason: "duplicate input examined reference",
		},
		{
			name: "roots-examined-not-array",
			mutate: func(m map[string]any) {
				m["roots_examined"] = 123
			},
			reason: "roots examined array",
		},
		{
			name: "roots-examined-invalid-id",
			mutate: func(m map[string]any) {
				m["roots_examined"] = []any{"ROOT_UPPER"}
			},
			reason: "invalid root examined identity",
		},
		{
			name: "roots-examined-duplicate",
			mutate: func(m map[string]any) {
				m["roots_examined"] = []any{"root", "root"}
			},
			reason: "duplicate root examined identity",
		},
		{
			name: "unresolved-candidates-not-array",
			mutate: func(m map[string]any) {
				m["unresolved_candidates"] = "none"
			},
			reason: "unresolved candidates array",
		},
		{
			name: "unresolved-candidates-empty-string",
			mutate: func(m map[string]any) {
				m["unresolved_candidates"] = []any{""}
			},
			reason: "invalid unresolved candidates entry",
		},
		{
			name: "unresolved-candidates-whitespace-entry",
			mutate: func(m map[string]any) {
				m["unresolved_candidates"] = []any{"   "}
			},
			reason: "invalid unresolved candidates entry",
		},
		{
			name: "unresolved-candidates-duplicate",
			mutate: func(m map[string]any) {
				m["status"] = "INSUFFICIENT"
				m["unresolved_candidates"] = []any{"cand1", "cand1"}
			},
			reason: "duplicate unresolved candidates entry",
		},
		{
			name: "unresolved-candidates-not-empty-when-sufficient",
			mutate: func(m map[string]any) {
				m["unresolved_candidates"] = []any{"unexpected-dynamic"}
			},
			reason: "sufficient coverage requires empty unresolved candidates",
		},
		{
			name: "excluded-mechanisms-not-array",
			mutate: func(m map[string]any) {
				m["excluded_mechanisms"] = 42
			},
			reason: "excluded mechanisms array",
		},
		{
			name: "excluded-mechanisms-empty-string",
			mutate: func(m map[string]any) {
				m["excluded_mechanisms"] = []any{""}
			},
			reason: "invalid excluded mechanisms entry",
		},
		{
			name: "excluded-mechanisms-duplicate",
			mutate: func(m map[string]any) {
				m["status"] = "INSUFFICIENT"
				m["excluded_mechanisms"] = []any{"mech1", "mech1"}
			},
			reason: "duplicate excluded mechanisms entry",
		},
		{
			name: "excluded-mechanisms-not-empty-when-sufficient",
			mutate: func(m map[string]any) {
				m["excluded_mechanisms"] = []any{"eval"}
			},
			reason: "sufficient coverage requires empty excluded mechanisms",
		},
		{
			name: "closure-controls-not-array",
			mutate: func(m map[string]any) {
				m["closure_control_refs"] = false
			},
			reason: "closure control references array",
		},
		{
			name: "closure-controls-invalid-ref",
			mutate: func(m map[string]any) {
				m["closure_control_refs"] = []any{"not-a-hash"}
			},
			reason: "invalid closure control reference",
		},
		{
			name: "closure-controls-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				m["closure_control_refs"] = []any{ref, ref}
			},
			reason: "duplicate closure control reference",
		},
		{
			name: "closure-controls-empty-when-sufficient",
			mutate: func(m map[string]any) {
				m["closure_control_refs"] = []any{}
			},
			reason: "sufficient coverage requires closure control references",
		},
		{
			name: "categories-not-array",
			mutate: func(m map[string]any) {
				m["categories"] = "SOURCE"
			},
			reason: "categories array",
		},
		{
			name: "categories-empty",
			mutate: func(m map[string]any) {
				m["categories"] = []any{}
			},
			reason: "categories cannot be empty",
		},
		{
			name: "category-not-object",
			mutate: func(m map[string]any) {
				m["categories"] = []any{"SOURCE"}
			},
			reason: "category object",
		},
		{
			name: "category-extra-field",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["extra"] = "unexpected"
			},
			reason: "category fields",
		},
		{
			name: "category-missing-field",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				delete(cat, "evidence_refs")
			},
			reason: "missing category field",
		},
		{
			name: "category-unknown-name",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["category"] = "MAGIC"
			},
			reason: "unknown category",
		},
		{
			name: "category-duplicate-name",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				m["categories"] = append(cats, cats[0])
			},
			reason: "duplicate category",
		},
		{
			name: "category-unknown-status",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["status"] = "UNVERIFIED"
			},
			reason: "unknown category status",
		},
		{
			name: "category-status-partial-when-sufficient",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["status"] = "PARTIAL"
			},
			reason: "sufficient coverage requires closed or irrelevant category status",
		},
		{
			name: "category-status-unknown-when-sufficient",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["status"] = "UNKNOWN"
			},
			reason: "sufficient coverage requires closed or irrelevant category status",
		},
		{
			name: "category-input-refs-not-array",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["input_refs"] = "not-array"
			},
			reason: "category input references array",
		},
		{
			name: "category-input-refs-invalid-ref",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["input_refs"] = []any{"bad-ref"}
			},
			reason: "invalid category input reference",
		},
		{
			name: "category-input-refs-duplicate",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				cat["input_refs"] = []any{ref, ref}
			},
			reason: "duplicate category input reference",
		},
		{
			name: "category-input-ref-unregistered",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["input_refs"] = []any{"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
			},
			reason: "unregistered category input reference",
		},
		{
			name: "category-evidence-refs-not-array",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["evidence_refs"] = "not-array"
			},
			reason: "category evidence references array",
		},
		{
			name: "category-evidence-refs-empty",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["evidence_refs"] = []any{}
			},
			reason: "category evidence references array",
		},
		{
			name: "category-evidence-refs-invalid-ref",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				cat["evidence_refs"] = []any{"invalid"}
			},
			reason: "invalid category evidence reference",
		},
		{
			name: "category-evidence-refs-duplicate",
			mutate: func(m map[string]any) {
				cats := m["categories"].([]any)
				cat := cats[0].(map[string]any)
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				cat["evidence_refs"] = []any{ref, ref}
			},
			reason: "duplicate category evidence reference",
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
				ctrl := consumptionCoverageControl(t)
				tc.mutate(ctrl)
				data, err := json.Marshal(ctrl)
				if err != nil {
					t.Fatal(err)
				}
				// Use CanonicalJSON to ensure input is canonical unless testing non-canonical
				canon, err := CanonicalJSON(data)
				if err != nil {
					t.Fatal(err)
				}
				raw = canon
			}

			err := ValidateConsumptionCoverage(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertConsumptionCoverageRejected(t, err, tc.reason)
		})
	}
}
