package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAssessmentReportRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasons := map[string]string{
		"assessment-cannot-be-verification-pass":            "object fields",
		"assessment-never-authorizes-implementation":        "implementation cannot be authorized",
		"static-assessment-must-state-limits":               "static finding requires limitations",
		"assessment-material-finding-needs-provenance":      "provenance reference array",
		"assessment-finding-needs-shared-analysis-identity": "finding object fields",
		"assessment-does-not-borrow-project-status":         "assessment status",
		"blocked-assessment-cannot-hide-blocker":            "blocked assessment requires blocker or unsupported finding",
		"missing-evidence-is-not-observation":               "missing evidence finding cannot be observed",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/reports/assessment-report.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAssessmentReport(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid assessment report for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, exists := reasons[fixture.Name]
			if !exists {
				t.Fatalf("unexpected negative fixture without mapped reason: %s", fixture.Name)
			}
			assertAssessmentReportRejected(t, err, expectedReason)
		})
	}
	if count != 12 {
		t.Fatalf("expected 12 assessment report fixtures, found %d", count)
	}
}

func assessmentReportControl() map[string]any {
	return map[string]any{
		"schema_version":            "1.0",
		"assessment_id":             "assessment_fixture",
		"project_id":                "fixture",
		"assessment_status":         "COMPLETED",
		"migration_feasibility":     "BLOCKED",
		"implementation_authorized": false,
		"as_of_sequence":            15,
		"created_at":                "2026-10-03T00:00:00Z",
		"configuration_hash":        strings.Repeat("0", 64),
		"source_snapshot_ref":       "sha256:" + strings.Repeat("1", 64),
		"detection_ref":             "sha256:" + strings.Repeat("2", 64),
		"target_intent_ref":         "sha256:" + strings.Repeat("3", 64),
		"scope_ref":                 "sha256:" + strings.Repeat("4", 64),
		"input_refs": []any{
			map[string]any{
				"role": "SOURCE_SNAPSHOT",
				"ref":  "sha256:" + strings.Repeat("1", 64),
			},
		},
		"analysis_record_refs": []any{
			"sha256:" + strings.Repeat("7", 64),
		},
		"validity_graph_ref": "sha256:" + strings.Repeat("5", 64),
		"system_summary":     "Synthetic legacy PHP fixture with a payment workflow",
		"evidence_summary":   "Source/dependency/profile findings only; runtime is unavailable",
		"examined_scope":     "The selected fixture payment operation, not the whole application",
		"findings": []any{
			map[string]any{
				"finding_id":          "missing_extension",
				"analysis_record_ref": "sha256:" + strings.Repeat("7", 64),
				"classification":      "UNSUPPORTED",
				"basis":               "STATIC_ONLY",
				"category":            "BLOCKER",
				"summary":             "Required legacy extension is unavailable in the declared sandbox",
				"reason":              "Source dependency inventory requires an extension absent from the available fixture profile",
				"provenance_refs": []any{
					"sha256:" + strings.Repeat("8", 64),
				},
				"prerequisite_refs": []any{
					"sha256:" + strings.Repeat("1", 64),
				},
				"limitations": []any{
					"Static dependency/profile evidence only; no runtime behavior has been established",
				},
				"affected_ids": []any{
					"payment_workflow",
				},
				"recommended_action": "Acquire a pinned supported sandbox extension or narrow the scope explicitly",
			},
		},
		"missing_evidence": []any{
			"Executable sandbox and payment behavioral characterization",
		},
		"assumptions": []any{
			"Customer workflow inventory is available only for the selected operation",
		},
		"uncertainty": []any{
			"Runtime behavior and target feasibility remain uncharacterized",
		},
		"suggested_sequence": []any{
			map[string]any{
				"position": 1,
				"finding_ids": []any{
					"missing_extension",
				},
				"action":                   "Obtain extension/runtime evidence",
				"prerequisite_finding_ids": []any{},
			},
		},
		"effort_estimate": nil,
	}
}

func validateAssessmentReportControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateAssessmentReport(raw)
}

func assertAssessmentReportRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *AssessmentReportError
	if reason == "" || !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted rejection %q, got %v", reason, err)
	}
	if err.Error() != "invalid assessment report: "+reason {
		t.Fatal("error disclosed report content")
	}
}

func TestAssessmentReportControlValid(t *testing.T) {
	t.Parallel()
	if err := validateAssessmentReportControl(t, assessmentReportControl()); err != nil {
		t.Fatalf("expected control to be valid, got: %v", err)
	}
}

func TestAssessmentReportFieldRejections(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, field, reason string
		value               any
	}{
		{"version_null", "schema_version", "unsupported schema version", nil},
		{"version_unknown", "schema_version", "unsupported schema version", "2.0"},
		{"assessment_id_caps", "assessment_id", "assessment_id identity", "Assessment"},
		{"assessment_id_empty", "assessment_id", "assessment_id identity", ""},
		{"project_id_caps", "project_id", "project_id identity", "Fixture"},
		{"project_id_empty", "project_id", "project_id identity", ""},
		{"status_unsupported", "assessment_status", "assessment status", "RUNNING"},
		{"feasibility_bad", "migration_feasibility", "unsupported migration feasibility", "PASSED"},
		{"impl_auth_true", "implementation_authorized", "implementation cannot be authorized", true},
		{"as_of_seq_negative", "as_of_sequence", "as_of_sequence", -1},
		{"created_at_bad", "created_at", "created_at timestamp", "not-a-timestamp"},
		{"config_hash_bad", "configuration_hash", "configuration hash", "bad_hash"},
		{"snapshot_ref_bad", "source_snapshot_ref", "source_snapshot_ref reference", "bad"},
		{"detection_ref_bad", "detection_ref", "detection_ref reference", "bad"},
		{"scope_ref_bad", "scope_ref", "scope_ref reference", "bad"},
		{"validity_graph_ref_bad", "validity_graph_ref", "validity_graph_ref reference", "bad"},
		{"target_intent_ref_bad", "target_intent_ref", "target intent reference", "bad"},
		{"system_summary_empty", "system_summary", "system_summary", ""},
		{"evidence_summary_empty", "evidence_summary", "evidence_summary", ""},
		{"examined_scope_empty", "examined_scope", "examined_scope", ""},
		{"extra_property", "unknown_field", "object fields", true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			record := assessmentReportControl()
			record[tc.field] = tc.value
			assertAssessmentReportRejected(t, validateAssessmentReportControl(t, record), tc.reason)
		})
	}

	for field := range assessmentReportControl() {
		field := field
		t.Run("missing_"+field, func(t *testing.T) {
			t.Parallel()
			record := assessmentReportControl()
			delete(record, field)
			assertAssessmentReportRejected(t, validateAssessmentReportControl(t, record), "object fields")
		})
	}
}

func TestAssessmentReportCanonicalJSON(t *testing.T) {
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
			err := ValidateAssessmentReport(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var repErr *AssessmentReportError
			if !errors.As(err, &repErr) {
				t.Fatalf("expected AssessmentReportError, got %T", err)
			}
		})
	}
}
