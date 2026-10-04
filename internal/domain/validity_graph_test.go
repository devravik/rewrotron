package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidityGraphRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasonMap := map[string]string{
		"review2-current-v2-validity-graph-requires-provenance": "missing required field",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/validity-graph.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateValidityGraph(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid validity graph for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, ok := reasonMap[fixture.Name]
			if !ok {
				t.Fatalf("missing expected reason mapping for invalid fixture: %s", fixture.Name)
			}
			assertValidityGraphRejected(t, err, expectedReason)
		})
	}
	if count != 6 {
		t.Fatalf("expected 6 validity graph fixtures, found %d", count)
	}
}

func validityGraphControl(t *testing.T, name string) map[string]any {
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

func assertValidityGraphRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var graphErr *ValidityGraphError
	if !errors.As(err, &graphErr) || graphErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid validity graph: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestValidityGraphValidRecords(t *testing.T) {
	t.Parallel()
	fixtures := []string{
		"v2-validity-graph",
		"typed-implementation-qualification-edge",
		"typed-execution-result-edge",
		"typed-performance-result-edge",
		"review2-current-v2-validity-graph",
	}

	for _, name := range fixtures {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := validityGraphControl(t, name)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := CanonicalJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateValidityGraph(canonical); err != nil {
				t.Fatalf("expected valid %s, got: %v", name, err)
			}
		})
	}
}

func TestValidityGraphMutationsV20(t *testing.T) {
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
			name: "missing-graph-id",
			mutate: func(m map[string]any) {
				delete(m, "graph_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-graph-id",
			mutate: func(m map[string]any) {
				m["graph_id"] = "GRAPH_UPPER"
			},
			reason: "graph identity",
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
			name: "edges-not-array",
			mutate: func(m map[string]any) {
				m["edges"] = "not-array"
			},
			reason: "edges array",
		},
		{
			name: "edge-not-object",
			mutate: func(m map[string]any) {
				m["edges"] = []any{"edge"}
			},
			reason: "edge object",
		},
		{
			name: "edge-missing-field",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				delete(edge, "witness_refs")
			},
			reason: "missing edge field",
		},
		{
			name: "edge-extra-field",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["extra"] = "forbidden"
			},
			reason: "edge fields",
		},
		{
			name: "edge-invalid-from-ref",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["from_ref"] = "bad-ref"
			},
			reason: "edge from_ref",
		},
		{
			name: "edge-invalid-to-ref",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["to_ref"] = "bad-ref"
			},
			reason: "edge to_ref",
		},
		{
			name: "edge-cycle-self-reference",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["to_ref"] = edge["from_ref"]
			},
			reason: "edge cycle self-reference",
		},
		{
			name: "edge-invalid-kind",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["kind"] = "UNKNOWN_KIND"
			},
			reason: "invalid edge kind",
		},
		{
			name: "edge-invalid-predicate",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["predicate"] = "UNKNOWN_PREDICATE"
			},
			reason: "invalid edge predicate",
		},
		{
			name: "duplicate-edge",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				m["edges"] = append(edges, edges[0])
			},
			reason: "duplicate edge",
		},
		{
			name: "edge-witness-refs-not-array",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["witness_refs"] = "not-array"
			},
			reason: "witness_refs array",
		},
		{
			name: "edge-witness-refs-empty",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["witness_refs"] = []any{}
			},
			reason: "witness_refs array",
		},
		{
			name: "edge-witness-refs-invalid-ref",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				edge["witness_refs"] = []any{"bad-ref"}
			},
			reason: "invalid witness_refs reference",
		},
		{
			name: "edge-witness-refs-duplicate",
			mutate: func(m map[string]any) {
				edges := m["edges"].([]any)
				edge := edges[0].(map[string]any)
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				edge["witness_refs"] = []any{ref, ref}
			},
			reason: "duplicate witness_refs reference",
		},
		{
			name: "unknown-impact-refs-not-array",
			mutate: func(m map[string]any) {
				m["unknown_impact_refs"] = "not-array"
			},
			reason: "unknown_impact_refs array",
		},
		{
			name: "unknown-impact-refs-invalid-ref",
			mutate: func(m map[string]any) {
				m["unknown_impact_refs"] = []any{"bad-ref"}
			},
			reason: "invalid unknown_impact_refs reference",
		},
		{
			name: "unknown-impact-refs-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				m["unknown_impact_refs"] = []any{ref, ref}
			},
			reason: "duplicate unknown_impact_refs reference",
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
				ctrl := validityGraphControl(t, "v2-validity-graph")
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

			err := ValidateValidityGraph(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertValidityGraphRejected(t, err, tc.reason)
		})
	}
}

func TestValidityGraphMutationsV21(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]any)
		reason string
	}{
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
				m["consumption_coverage_refs"] = "not-array"
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
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := validityGraphControl(t, "review2-current-v2-validity-graph")
			tc.mutate(ctrl)
			data, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canon, err := CanonicalJSON(data)
			if err != nil {
				t.Fatal(err)
			}

			err = ValidateValidityGraph(canon)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertValidityGraphRejected(t, err, tc.reason)
		})
	}
}
