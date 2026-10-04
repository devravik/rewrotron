package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAssessmentPlanningBasisRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasons := map[string]string{
		"reuse-without-current-witness-invalid": "reused finding requires witness references",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/migration/assessment-planning-basis.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAssessmentPlanningBasis(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid assessment planning basis for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, exists := reasons[fixture.Name]
			if !exists {
				t.Fatalf("unexpected negative fixture without mapped reason: %s", fixture.Name)
			}
			assertAssessmentPlanningBasisRejected(t, err, expectedReason)
		})
	}
	if count != 2 {
		t.Fatalf("expected 2 assessment planning basis fixtures, found %d", count)
	}
}

func assessmentPlanningBasisControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "assessment-reuse-before-any-plan" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing assessment-reuse-before-any-plan fixture")
	return nil
}

func assertAssessmentPlanningBasisRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var basisErr *AssessmentPlanningBasisError
	if !errors.As(err, &basisErr) || basisErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid assessment planning basis: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestAssessmentPlanningBasisPositiveControls(t *testing.T) {
	t.Parallel()

	t.Run("fixture-control", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentPlanningBasisControl(t)
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentPlanningBasis(raw); err != nil {
			t.Fatalf("expected valid planning basis, got: %v", err)
		}
	})

	t.Run("empty-optional-arrays", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentPlanningBasisControl(t)
		ctrl["reused_findings"] = []any{}
		ctrl["stale_finding_refs"] = []any{}
		ctrl["new_finding_refs"] = []any{}
		ctrl["changes_since_assessment"] = []any{}
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentPlanningBasis(raw); err != nil {
			t.Fatalf("expected valid planning basis with empty arrays, got: %v", err)
		}
	})

	t.Run("change-variations", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentPlanningBasisControl(t)
		ctrl["changes_since_assessment"] = []any{
			map[string]any{
				"input_role": "SCOPE",
				"before_ref": "sha256:" + strings.Repeat("1", 64),
				"after_ref":  nil,
				"reason":     "Scoped out component",
				"provenance_refs": []any{
					"sha256:" + strings.Repeat("a", 64),
				},
			},
			map[string]any{
				"input_role": "DECISION",
				"before_ref": "sha256:" + strings.Repeat("2", 64),
				"after_ref":  "sha256:" + strings.Repeat("3", 64),
				"reason":     "Updated dependency adapter strategy",
				"provenance_refs": []any{
					"sha256:" + strings.Repeat("b", 64),
				},
			},
		}
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentPlanningBasis(raw); err != nil {
			t.Fatalf("expected valid planning basis with change variations, got: %v", err)
		}
	})
}

func TestAssessmentPlanningBasisNegativeMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]any)
		reason string
	}{
		{
			name: "wrong-schema-version",
			mutate: func(m map[string]any) {
				m["schema_version"] = "2.0"
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
			name: "invalid-assessment-ref",
			mutate: func(m map[string]any) {
				m["assessment_ref"] = "invalid-ref"
			},
			reason: "assessment_ref",
		},
		{
			name: "invalid-selection-ref",
			mutate: func(m map[string]any) {
				m["selection_ref"] = "invalid-ref"
			},
			reason: "selection_ref",
		},
		{
			name: "invalid-selected-scope-ref",
			mutate: func(m map[string]any) {
				m["selected_scope_ref"] = "invalid-ref"
			},
			reason: "selected_scope_ref",
		},
		{
			name: "invalid-validity-graph-ref",
			mutate: func(m map[string]any) {
				m["validity_graph_ref"] = "invalid-ref"
			},
			reason: "validity_graph_ref",
		},
		{
			name: "empty-current-input-refs",
			mutate: func(m map[string]any) {
				m["current_input_refs"] = []any{}
			},
			reason: "current input references array",
		},
		{
			name: "current-input-missing-role",
			mutate: func(m map[string]any) {
				m["current_input_refs"] = []any{
					map[string]any{
						"ref": "sha256:" + strings.Repeat("1", 64),
					},
				}
			},
			reason: "current input reference fields",
		},
		{
			name: "current-input-unsupported-role",
			mutate: func(m map[string]any) {
				m["current_input_refs"] = []any{
					map[string]any{
						"role": "UNSUPPORTED_ROLE",
						"ref":  "sha256:" + strings.Repeat("1", 64),
					},
				}
			},
			reason: "unsupported current input reference role",
		},
		{
			name: "current-input-invalid-ref",
			mutate: func(m map[string]any) {
				m["current_input_refs"] = []any{
					map[string]any{
						"role": "ANALYSIS",
						"ref":  "bad-ref",
					},
				}
			},
			reason: "current input reference",
		},
		{
			name: "current-input-duplicate",
			mutate: func(m map[string]any) {
				inputs := m["current_input_refs"].([]any)
				m["current_input_refs"] = append(inputs, inputs[0])
			},
			reason: "duplicate current input reference",
		},
		{
			name: "reused-finding-missing-field",
			mutate: func(m map[string]any) {
				m["reused_findings"] = []any{
					map[string]any{
						"analysis_record_ref": "sha256:" + strings.Repeat("1", 64),
					},
				}
			},
			reason: "reused finding fields",
		},
		{
			name: "reused-finding-invalid-record-ref",
			mutate: func(m map[string]any) {
				reused := m["reused_findings"].([]any)
				f0 := reused[0].(map[string]any)
				f0["analysis_record_ref"] = "bad-ref"
			},
			reason: "reused finding analysis record reference",
		},
		{
			name: "reused-finding-duplicate",
			mutate: func(m map[string]any) {
				reused := m["reused_findings"].([]any)
				m["reused_findings"] = append(reused, reused[0])
			},
			reason: "duplicate reused finding",
		},
		{
			name: "reused-finding-empty-prereqs",
			mutate: func(m map[string]any) {
				reused := m["reused_findings"].([]any)
				f0 := reused[0].(map[string]any)
				f0["checked_prerequisite_refs"] = []any{}
			},
			reason: "reused finding requires checked prerequisite references",
		},
		{
			name: "reused-finding-empty-witness-refs",
			mutate: func(m map[string]any) {
				reused := m["reused_findings"].([]any)
				f0 := reused[0].(map[string]any)
				f0["witness_refs"] = []any{}
			},
			reason: "reused finding requires witness references",
		},
		{
			name: "reused-finding-duplicate-witness-ref",
			mutate: func(m map[string]any) {
				reused := m["reused_findings"].([]any)
				f0 := reused[0].(map[string]any)
				w := "sha256:" + strings.Repeat("a", 64)
				f0["witness_refs"] = []any{w, w}
			},
			reason: "duplicate witness reference",
		},
		{
			name: "reused-finding-cannot-be-stale",
			mutate: func(m map[string]any) {
				// add reused finding's ref to stale_finding_refs
				reused := m["reused_findings"].([]any)
				f0 := reused[0].(map[string]any)
				ref := f0["analysis_record_ref"].(string)
				stale := m["stale_finding_refs"].([]any)
				m["stale_finding_refs"] = append(stale, ref)
			},
			reason: "reused finding cannot be stale",
		},
		{
			name: "reused-finding-cannot-be-new",
			mutate: func(m map[string]any) {
				reused := m["reused_findings"].([]any)
				f0 := reused[0].(map[string]any)
				ref := f0["analysis_record_ref"].(string)
				newRefs := m["new_finding_refs"].([]any)
				m["new_finding_refs"] = append(newRefs, ref)
			},
			reason: "reused finding cannot be new",
		},
		{
			name: "stale-finding-cannot-be-new",
			mutate: func(m map[string]any) {
				stale := m["stale_finding_refs"].([]any)
				ref := stale[0].(string)
				newRefs := m["new_finding_refs"].([]any)
				m["new_finding_refs"] = append(newRefs, ref)
			},
			reason: "stale finding cannot be new",
		},
		{
			name: "change-unsupported-role",
			mutate: func(m map[string]any) {
				changes := m["changes_since_assessment"].([]any)
				c0 := changes[0].(map[string]any)
				c0["input_role"] = "UNSUPPORTED_ROLE"
			},
			reason: "unsupported change input role",
		},
		{
			name: "change-both-before-and-after-null",
			mutate: func(m map[string]any) {
				changes := m["changes_since_assessment"].([]any)
				c0 := changes[0].(map[string]any)
				c0["before_ref"] = nil
				c0["after_ref"] = nil
			},
			reason: "change requires before or after reference",
		},
		{
			name: "change-before-and-after-identical",
			mutate: func(m map[string]any) {
				changes := m["changes_since_assessment"].([]any)
				c0 := changes[0].(map[string]any)
				same := "sha256:" + strings.Repeat("7", 64)
				c0["before_ref"] = same
				c0["after_ref"] = same
			},
			reason: "change before and after references cannot be identical",
		},
		{
			name: "change-empty-reason",
			mutate: func(m map[string]any) {
				changes := m["changes_since_assessment"].([]any)
				c0 := changes[0].(map[string]any)
				c0["reason"] = ""
			},
			reason: "change reason",
		},
		{
			name: "change-empty-provenance-refs",
			mutate: func(m map[string]any) {
				changes := m["changes_since_assessment"].([]any)
				c0 := changes[0].(map[string]any)
				c0["provenance_refs"] = []any{}
			},
			reason: "change requires provenance references",
		},
		{
			name: "change-duplicate-provenance-ref",
			mutate: func(m map[string]any) {
				changes := m["changes_since_assessment"].([]any)
				c0 := changes[0].(map[string]any)
				p := "sha256:" + strings.Repeat("b", 64)
				c0["provenance_refs"] = []any{p, p}
			},
			reason: "duplicate provenance reference",
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
			ctrl := assessmentPlanningBasisControl(t)
			tc.mutate(ctrl)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateAssessmentPlanningBasis(raw)
			assertAssessmentPlanningBasisRejected(t, err, tc.reason)
		})
	}
}

func TestAssessmentPlanningBasisCanonicalJSON(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"float_number", []byte(`{"schema_version":"1.0","num":1.5}`)},
		{"duplicate_keys", []byte(`{"schema_version":"1.0","schema_version":"1.0"}`)},
		{"invalid_utf8", []byte{0xff, 0xfe}},
		{"trailing_data", []byte(`{"schema_version":"1.0"}{}`)},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAssessmentPlanningBasis(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var basisErr *AssessmentPlanningBasisError
			if !errors.As(err, &basisErr) || basisErr.Reason != "canonical JSON" {
				t.Fatalf("expected canonical JSON error, got %v", err)
			}
		})
	}
}

func TestAssessmentPlanningBasisErrorSecurity(t *testing.T) {
	t.Parallel()
	ctrl := assessmentPlanningBasisControl(t)
	ctrl["project_id"] = "CANARY_SECRET_LEAK!"
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateAssessmentPlanningBasis(raw)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errStr := err.Error()
	if strings.Contains(errStr, "CANARY_SECRET_LEAK") {
		t.Fatalf("error message discloses sensitive input: %s", errStr)
	}
	if errStr != "invalid assessment planning basis: project identity" {
		t.Fatalf("unexpected error format: %s", errStr)
	}
}
