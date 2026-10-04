package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAssessmentSelectionRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasons := map[string]string{
		"continuation-needs-selected-work":                         "selected scope reference required for continuation",
		"continuation-is-not-implementation-permission":            "implementation cannot be authorized",
		"agent-cannot-supply-human-mcp-channel":                    "human action channel",
		"assessment-select_work-needs-explicit-subject":            "selected scope reference required",
		"assessment-change_scope-needs-explicit-subject":           "selected scope reference required",
		"assessment-change_target-needs-explicit-subject":          "target intent reference required",
		"assessment-accept_recommendations-needs-explicit-subject": "finding IDs required for recommendations action",
		"assessment-reject_recommendations-needs-explicit-subject": "finding IDs required for recommendations action",
		"assessment-request_evidence-needs-explicit-subject":       "requested evidence required for evidence request",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/assessment-selection.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAssessmentSelection(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid assessment selection for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, exists := reasons[fixture.Name]
			if !exists {
				t.Fatalf("unexpected negative fixture without mapped reason: %s", fixture.Name)
			}
			assertAssessmentSelectionRejected(t, err, expectedReason)
		})
	}
	if count != 17 {
		t.Fatalf("expected 17 assessment selection fixtures, found %d", count)
	}
}

func assessmentSelectionControl() map[string]any {
	return map[string]any{
		"schema_version":            "1.0",
		"selection_id":              "selection_fixture",
		"project_id":                "fixture",
		"assessment_ref":            "sha256:" + strings.Repeat("a", 64),
		"action":                    "CONTINUE_PLANNING",
		"selected_scope_ref":        "sha256:" + strings.Repeat("b", 64),
		"target_intent_ref":         "sha256:" + strings.Repeat("c", 64),
		"finding_ids":               []any{"finding_1"},
		"decision_refs":             []any{"sha256:" + strings.Repeat("d", 64)},
		"requested_evidence":        []any{},
		"rationale":                 "Proceed to planning for authorized scope",
		"implementation_authorized": false,
		"human_action": map[string]any{
			"actor_id":    "fixture-human",
			"channel":     "CLI",
			"action_id":   "action_1",
			"recorded_at": "2026-10-04T00:00:00Z",
		},
	}
}

func validateAssessmentSelectionControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateAssessmentSelection(raw)
}

func assertAssessmentSelectionRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *AssessmentSelectionError
	if reason == "" || !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted rejection %q, got %v", reason, err)
	}
	if err.Error() != "invalid assessment selection: "+reason {
		t.Fatal("error disclosed selection data")
	}
}

func TestAssessmentSelectionControlValid(t *testing.T) {
	t.Parallel()
	if err := validateAssessmentSelectionControl(t, assessmentSelectionControl()); err != nil {
		t.Fatalf("expected control to be valid, got: %v", err)
	}

	stopRecord := assessmentSelectionControl()
	stopRecord["action"] = "STOP"
	stopRecord["selected_scope_ref"] = nil
	stopRecord["target_intent_ref"] = nil
	stopRecord["finding_ids"] = []any{}
	if err := validateAssessmentSelectionControl(t, stopRecord); err != nil {
		t.Fatalf("expected STOP control to be valid, got: %v", err)
	}
}

func TestAssessmentSelectionFieldRejections(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, field, reason string
		value               any
	}{
		{"version_null", "schema_version", "unsupported schema version", nil},
		{"version_unknown", "schema_version", "unsupported schema version", "2.0"},
		{"selection_id_empty", "selection_id", "selection_id identity", ""},
		{"selection_id_caps", "selection_id", "selection_id identity", "Selection"},
		{"selection_id_null", "selection_id", "selection_id identity", nil},
		{"project_id_empty", "project_id", "project_id identity", ""},
		{"project_id_caps", "project_id", "project_id identity", "Project"},
		{"project_id_null", "project_id", "project_id identity", nil},
		{"assessment_ref_bad", "assessment_ref", "assessment reference", "bad"},
		{"assessment_ref_null", "assessment_ref", "assessment reference", nil},
		{"action_unsupported", "action", "unsupported action", "EXECUTE_NOW"},
		{"scope_ref_bad", "selected_scope_ref", "selected scope reference", "bad"},
		{"target_ref_bad", "target_intent_ref", "target intent reference", "bad"},
		{"finding_ids_bad_item", "finding_ids", "finding ID item", []any{"Bad_ID!"}},
		{"finding_ids_duplicate", "finding_ids", "duplicate finding ID", []any{"id1", "id1"}},
		{"decision_refs_bad", "decision_refs", "decision reference item", []any{"bad"}},
		{"decision_refs_duplicate", "decision_refs", "duplicate decision reference", []any{
			"sha256:" + strings.Repeat("a", 64),
			"sha256:" + strings.Repeat("a", 64),
		}},
		{"requested_evidence_bad", "requested_evidence", "requested evidence item", []any{""}},
		{"requested_evidence_duplicate", "requested_evidence", "duplicate requested evidence", []any{"item1", "item1"}},
		{"rationale_empty", "rationale", "rationale", ""},
		{"rationale_null", "rationale", "rationale", nil},
		{"impl_auth_true", "implementation_authorized", "implementation cannot be authorized", true},
		{"impl_auth_null", "implementation_authorized", "implementation cannot be authorized", nil},
		{"extra_property", "unknown_field", "object fields", true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			record := assessmentSelectionControl()
			record[tc.field] = tc.value
			assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), tc.reason)
		})
	}

	for field := range assessmentSelectionControl() {
		field := field
		t.Run("missing_"+field, func(t *testing.T) {
			t.Parallel()
			record := assessmentSelectionControl()
			delete(record, field)
			assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "object fields")
		})
	}
}

func TestAssessmentSelectionHumanActionRejections(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, reason string
		modify       func(map[string]any)
	}{
		{"actor_id_empty", "human action actor identity", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["actor_id"] = ""
		}},
		{"actor_id_null", "human action actor identity", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["actor_id"] = nil
		}},
		{"channel_mcp", "human action channel", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["channel"] = "MCP"
		}},
		{"action_id_caps", "human action action identity", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["action_id"] = "Action"
		}},
		{"action_id_empty", "human action action identity", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["action_id"] = ""
		}},
		{"recorded_at_bad", "human action recorded_at timestamp", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["recorded_at"] = "not-a-timestamp"
		}},
		{"missing_actor_id", "human action fields", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			delete(h, "actor_id")
		}},
		{"extra_human_field", "human action fields", func(m map[string]any) {
			h := m["human_action"].(map[string]any)
			h["extra"] = true
		}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			record := assessmentSelectionControl()
			tc.modify(record)
			assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), tc.reason)
		})
	}
}

func TestAssessmentSelectionConditionalActionRules(t *testing.T) {
	t.Parallel()

	t.Run("continue_planning_missing_target", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["target_intent_ref"] = nil
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "target intent reference required for continuation")
	})

	t.Run("select_work_missing_scope", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["action"] = "SELECT_WORK"
		record["selected_scope_ref"] = nil
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "selected scope reference required")
	})

	t.Run("change_scope_missing_scope", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["action"] = "CHANGE_SCOPE"
		record["selected_scope_ref"] = nil
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "selected scope reference required")
	})

	t.Run("change_target_missing_target", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["action"] = "CHANGE_TARGET"
		record["target_intent_ref"] = nil
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "target intent reference required")
	})

	t.Run("accept_recommendations_empty_findings", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["action"] = "ACCEPT_RECOMMENDATIONS"
		record["finding_ids"] = []any{}
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "finding IDs required for recommendations action")
	})

	t.Run("reject_recommendations_empty_findings", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["action"] = "REJECT_RECOMMENDATIONS"
		record["finding_ids"] = []any{}
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "finding IDs required for recommendations action")
	})

	t.Run("request_evidence_empty_evidence", func(t *testing.T) {
		t.Parallel()
		record := assessmentSelectionControl()
		record["action"] = "REQUEST_EVIDENCE"
		record["requested_evidence"] = []any{}
		assertAssessmentSelectionRejected(t, validateAssessmentSelectionControl(t, record), "requested evidence required for evidence request")
	})
}

func TestAssessmentSelectionCanonicalJSON(t *testing.T) {
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
			err := ValidateAssessmentSelection(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var selErr *AssessmentSelectionError
			if !errors.As(err, &selErr) {
				t.Fatalf("expected AssessmentSelectionError, got %T", err)
			}
		})
	}
}
