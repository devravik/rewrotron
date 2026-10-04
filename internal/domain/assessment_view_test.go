package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAssessmentViewRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasons := map[string]string{
		"valid-assessment-view-cannot-hide-stale-finding": "valid view contains invalid finding",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/reports/assessment-view.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAssessmentView(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid assessment view for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, exists := reasons[fixture.Name]
			if !exists {
				t.Fatalf("unexpected negative fixture without mapped reason: %s", fixture.Name)
			}
			assertAssessmentViewRejected(t, err, expectedReason)
		})
	}
	if count != 2 {
		t.Fatalf("expected 2 assessment view fixtures, found %d", count)
	}
}

func assessmentViewControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "historical-assessment-and-current-validity-distinct" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing historical-assessment-and-current-validity-distinct fixture")
	return nil
}

func assertAssessmentViewRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var viewErr *AssessmentViewError
	if !errors.As(err, &viewErr) || viewErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid assessment view: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestAssessmentViewPositiveControls(t *testing.T) {
	t.Parallel()

	t.Run("stale-control", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentViewControl(t)
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentView(raw); err != nil {
			t.Fatalf("expected valid view, got: %v", err)
		}
	})

	t.Run("valid-view-all-valid-findings", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentViewControl(t)
		ctrl["validity"] = "VALID"
		ctrl["is_latest_report"] = true
		findings := ctrl["finding_validity"].([]any)
		f0 := findings[0].(map[string]any)
		f0["validity"] = "VALID"
		f0["witness_refs"] = []any{"sha256:" + strings.Repeat("f", 64)}

		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentView(raw); err != nil {
			t.Fatalf("expected valid view, got: %v", err)
		}
	})

	t.Run("unknown-view-nullable-graph-ref", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentViewControl(t)
		ctrl["validity"] = "UNKNOWN"
		ctrl["current_validity_graph_ref"] = nil
		findings := ctrl["finding_validity"].([]any)
		f0 := findings[0].(map[string]any)
		f0["validity"] = "UNKNOWN"
		f0["witness_refs"] = []any{}

		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentView(raw); err != nil {
			t.Fatalf("expected valid unknown view, got: %v", err)
		}
	})

	t.Run("equal-sequence", func(t *testing.T) {
		t.Parallel()
		ctrl := assessmentViewControl(t)
		ctrl["current_state_sequence"] = int64(15) // equal to report as_of_sequence 15
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateAssessmentView(raw); err != nil {
			t.Fatalf("expected valid view at equal sequence, got: %v", err)
		}
	})
}

func TestAssessmentViewNegativeMutations(t *testing.T) {
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
			name: "invalid-report-ref",
			mutate: func(m map[string]any) {
				m["report_ref"] = "md5:abcdef"
			},
			reason: "report reference",
		},
		{
			name: "is-latest-not-bool",
			mutate: func(m map[string]any) {
				m["is_latest_report"] = "true"
			},
			reason: "is_latest_report",
		},
		{
			name: "is-latest-null",
			mutate: func(m map[string]any) {
				m["is_latest_report"] = nil
			},
			reason: "is_latest_report",
		},
		{
			name: "project-id-mismatch",
			mutate: func(m map[string]any) {
				m["project_id"] = "other-project"
			},
			reason: "project ID mismatch",
		},
		{
			name: "sequence-regression",
			mutate: func(m map[string]any) {
				m["current_state_sequence"] = int64(10) // report as_of_sequence is 15
			},
			reason: "sequence regression",
		},
		{
			name: "negative-current-sequence",
			mutate: func(m map[string]any) {
				m["current_state_sequence"] = int64(-1)
			},
			reason: "current_state_sequence",
		},
		{
			name: "unsupported-view-validity",
			mutate: func(m map[string]any) {
				m["validity"] = "PENDING"
			},
			reason: "unsupported validity",
		},
		{
			name: "valid-view-missing-graph-ref",
			mutate: func(m map[string]any) {
				m["validity"] = "VALID"
				m["current_validity_graph_ref"] = nil
				findings := m["finding_validity"].([]any)
				f0 := findings[0].(map[string]any)
				f0["validity"] = "VALID"
				f0["witness_refs"] = []any{"sha256:" + strings.Repeat("f", 64)}
			},
			reason: "valid view requires validity graph reference",
		},
		{
			name: "stale-view-invalid-graph-ref-format",
			mutate: func(m map[string]any) {
				m["current_validity_graph_ref"] = "sha256:not-hex"
			},
			reason: "validity graph reference",
		},
		{
			name: "finding-count-mismatch-missing-finding",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				m["finding_validity"] = findings[:1]
			},
			reason: "finding validity count mismatch",
		},
		{
			name: "finding-count-mismatch-extra-finding",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				extra := map[string]any{
					"finding_id":          "extra_finding",
					"analysis_record_ref": "sha256:" + strings.Repeat("9", 64),
					"validity":            "STALE",
					"reason":              "extra",
					"witness_refs":        []any{},
				}
				m["finding_validity"] = append(findings, extra)
			},
			reason: "finding validity count mismatch",
		},
		{
			name: "unknown-finding-id",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f0 := findings[0].(map[string]any)
				f0["finding_id"] = "unknown_finding"
			},
			reason: "unknown finding ID",
		},
		{
			name: "duplicate-finding-id",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f1 := findings[1].(map[string]any)
				f1["finding_id"] = "missing_extension"
				f1["analysis_record_ref"] = findings[0].(map[string]any)["analysis_record_ref"]
			},
			reason: "duplicate finding validity",
		},
		{
			name: "finding-analysis-record-ref-mismatch",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f0 := findings[0].(map[string]any)
				f0["analysis_record_ref"] = "sha256:" + strings.Repeat("9", 64)
			},
			reason: "finding analysis record reference mismatch",
		},
		{
			name: "unsupported-finding-validity",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f0 := findings[0].(map[string]any)
				f0["validity"] = "EXPIRED"
			},
			reason: "unsupported finding validity",
		},
		{
			name: "valid-view-contains-stale-finding",
			mutate: func(m map[string]any) {
				m["validity"] = "VALID"
				// f0 is STALE in control
			},
			reason: "valid view contains invalid finding",
		},
		{
			name: "stale-finding-requires-stale-view",
			mutate: func(m map[string]any) {
				m["validity"] = "UNKNOWN"
				// f0 is STALE, view is UNKNOWN
			},
			reason: "stale finding requires stale view",
		},
		{
			name: "finding-empty-reason",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f0 := findings[0].(map[string]any)
				f0["reason"] = ""
			},
			reason: "finding reason",
		},
		{
			name: "valid-finding-missing-witness-refs",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f1 := findings[1].(map[string]any) // validity is VALID
				f1["witness_refs"] = []any{}
			},
			reason: "valid finding requires witness references",
		},
		{
			name: "duplicate-witness-ref",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f1 := findings[1].(map[string]any)
				w := "sha256:" + strings.Repeat("d", 64)
				f1["witness_refs"] = []any{w, w}
			},
			reason: "duplicate witness reference",
		},
		{
			name: "invalid-witness-ref-format",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f1 := findings[1].(map[string]any)
				f1["witness_refs"] = []any{"bad-ref"}
			},
			reason: "witness reference",
		},
		{
			name: "extra-top-level-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
		{
			name: "finding-extra-field",
			mutate: func(m map[string]any) {
				findings := m["finding_validity"].([]any)
				f0 := findings[0].(map[string]any)
				f0["extra"] = "forbidden"
			},
			reason: "finding validity fields",
		},
		{
			name: "invalid-report",
			mutate: func(m map[string]any) {
				rep := m["report"].(map[string]any)
				rep["implementation_authorized"] = true
			},
			reason: "report",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := assessmentViewControl(t)
			tc.mutate(ctrl)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateAssessmentView(raw)
			assertAssessmentViewRejected(t, err, tc.reason)
		})
	}
}

func TestAssessmentViewCanonicalJSON(t *testing.T) {
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
			err := ValidateAssessmentView(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var viewErr *AssessmentViewError
			if !errors.As(err, &viewErr) || viewErr.Reason != "canonical JSON" {
				t.Fatalf("expected canonical JSON error, got %v", err)
			}
		})
	}
}

func TestAssessmentViewErrorSecurity(t *testing.T) {
	t.Parallel()
	ctrl := assessmentViewControl(t)
	ctrl["project_id"] = "CANARY_SECRET_LEAK!"
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateAssessmentView(raw)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errStr := err.Error()
	if strings.Contains(errStr, "CANARY_SECRET_LEAK") {
		t.Fatalf("error message discloses sensitive input: %s", errStr)
	}
	if errStr != "invalid assessment view: project identity" {
		t.Fatalf("unexpected error format: %s", errStr)
	}
}
