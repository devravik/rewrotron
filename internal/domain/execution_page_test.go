package domain

import (
	"encoding/json"
	"os"
	"testing"
)

func TestValidateExecutionPage_Fixtures(t *testing.T) {
	fixtureData, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases fixture: %v", err)
	}

	var cases []validationFixtureCase
	if err := json.Unmarshal(fixtureData, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	testedPages := 0
	for _, tc := range cases {
		if tc.Schema == "https://rewrotron.dev/specs/execution/page.schema.json" {
			testedPages++
			err := ValidateExecutionPage(tc.Instance)
			if tc.Valid && err != nil {
				t.Errorf("fixture %q expected valid, got error: %v", tc.Name, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("fixture %q expected invalid, got nil error", tc.Name)
			}
		}
	}

	if testedPages < 2 {
		t.Fatalf("expected at least 2 execution page fixtures, got %d", testedPages)
	}
}

func baseValidStateMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":        "1.0",
		"project_id":            "prj-1",
		"lifecycle_state":       "ANALYZED",
		"control_mode":          "RUNNING",
		"control_reason":        "One observation running",
		"latest_event_sequence": 10,
		"observed_at":           "2026-10-04T12:00:01Z",
		"focus_run_id":          "run-1",
		"active_runs": []map[string]interface{}{
			{
				"run_id":            "run-1",
				"operation_id":      "op-1",
				"stage":             "OBSERVE",
				"stage_instance_id": "stage-1",
				"status":            "RUNNING",
				"current_task": map[string]interface{}{
					"task_id":       "t-1",
					"mission_id":    "m-1",
					"title":         "Characterize behavior",
					"reason":        "Characterizing legacy routes",
					"status":        "RUNNING",
					"result":        nil,
					"next_action":   "Run test case",
					"evidence_refs": []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
				},
				"last_activity_at":  "2026-10-04T12:00:00Z",
				"last_heartbeat_at": "2026-10-04T12:00:00Z",
				"subjects": map[string]interface{}{
					"configuration_hash": nil,
					"plan_hash":          nil,
					"target_spec_hash":   nil,
					"target_tree_hash":   nil,
					"corpus_hash":        nil,
					"policy_hash":        nil,
				},
			},
		},
		"stage_instances": []map[string]interface{}{
			{
				"stage_instance_id": "stage-1",
				"run_id":            "run-1",
				"stage":             "OBSERVE",
				"status":            "RUNNING",
				"summary":           "Observing behavior",
				"evidence_refs":     []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
				"artifact_refs":     []string{},
			},
		},
		"last_activity_at":  "2026-10-04T12:00:00Z",
		"last_heartbeat_at": "2026-10-04T12:00:00Z",
		"liveness":          "FRESH",
		"progress_metrics":  []map[string]interface{}{},
		"blockers":          []map[string]interface{}{},
		"pending_approvals": []map[string]interface{}{},
		"next_action":       "Run test case",
		"user_request":      nil,
		"subjects": map[string]interface{}{
			"configuration_hash": nil,
			"plan_hash":          nil,
			"target_spec_hash":   nil,
			"target_tree_hash":   nil,
			"corpus_hash":        nil,
			"policy_hash":        nil,
		},
		"recent_change_refs": []string{},
	}
}

func baseValidPageMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":   "1.0",
		"project_id":       "prj-1",
		"after_sequence":   0,
		"through_sequence": 10,
		"next_sequence":    10,
		"has_more":         false,
		"state":            baseValidStateMap(),
		"events":           []interface{}{baseValidEventMap()},
		"changes": []map[string]interface{}{
			{
				"kind":          "DISCOVERED",
				"entity_type":   "case",
				"entity_id":     "case-1",
				"summary":       "Discovered new route",
				"event_id":      "evt-1",
				"sequence":      1,
				"evidence_refs": []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
			},
		},
	}
}

func TestValidateExecutionPage_Valid10And20(t *testing.T) {
	page10 := baseValidPageMap()
	raw10, err := json.Marshal(page10)
	if err != nil {
		t.Fatalf("failed to marshal 1.0 page: %v", err)
	}
	if err := ValidateExecutionPage(raw10); err != nil {
		t.Errorf("expected valid 1.0 page, got %v", err)
	}

	page20 := baseValidPageMap()
	page20["schema_version"] = "2.0"
	page20["state_semantics"] = "CURRENT"
	page20["state_sequence"] = 10
	raw20, err := json.Marshal(page20)
	if err != nil {
		t.Fatalf("failed to marshal 2.0 page: %v", err)
	}
	if err := ValidateExecutionPage(raw20); err != nil {
		t.Errorf("expected valid 2.0 page, got %v", err)
	}
}

func TestValidateExecutionPage_SequenceConstraints(t *testing.T) {
	// after_sequence > through_sequence
	page := baseValidPageMap()
	page["after_sequence"] = 15
	page["through_sequence"] = 10
	raw, _ := json.Marshal(page)
	if err := ValidateExecutionPage(raw); err == nil {
		t.Errorf("expected error when after_sequence > through_sequence")
	}

	// negative after_sequence
	page["after_sequence"] = -1
	rawNeg, _ := json.Marshal(page)
	if err := ValidateExecutionPage(rawNeg); err == nil {
		t.Errorf("expected error for negative after_sequence")
	}

	// negative through_sequence
	page = baseValidPageMap()
	page["through_sequence"] = -1
	rawNegThrough, _ := json.Marshal(page)
	if err := ValidateExecutionPage(rawNegThrough); err == nil {
		t.Errorf("expected error for negative through_sequence")
	}

	// negative next_sequence
	page = baseValidPageMap()
	page["next_sequence"] = -1
	rawNegNext, _ := json.Marshal(page)
	if err := ValidateExecutionPage(rawNegNext); err == nil {
		t.Errorf("expected error for negative next_sequence")
	}
}

func TestValidateExecutionPage_Version20Constraints(t *testing.T) {
	page20 := baseValidPageMap()
	page20["schema_version"] = "2.0"
	page20["state_semantics"] = "HISTORICAL"
	page20["state_sequence"] = 10
	raw, _ := json.Marshal(page20)
	if err := ValidateExecutionPage(raw); err == nil {
		t.Errorf("expected error when state_semantics != CURRENT")
	}

	// negative state_sequence
	page20["state_semantics"] = "CURRENT"
	page20["state_sequence"] = -1
	rawNegSeq, _ := json.Marshal(page20)
	if err := ValidateExecutionPage(rawNegSeq); err == nil {
		t.Errorf("expected error for negative state_sequence")
	}

	// 1.0 with state_semantics should fail (unexpected properties)
	page10Extra := baseValidPageMap()
	page10Extra["state_semantics"] = "CURRENT"
	page10Extra["state_sequence"] = 10
	raw10Extra, _ := json.Marshal(page10Extra)
	if err := ValidateExecutionPage(raw10Extra); err == nil {
		t.Errorf("expected error for 1.0 page with unexpected 2.0 fields")
	}
}

func TestValidateExecutionPage_EmbeddedErrors(t *testing.T) {
	// invalid state
	page := baseValidPageMap()
	state := baseValidStateMap()
	state["lifecycle_state"] = "INVALID_STATE"
	page["state"] = state
	rawBadState, _ := json.Marshal(page)
	if err := ValidateExecutionPage(rawBadState); err == nil {
		t.Errorf("expected error for invalid embedded state")
	}

	// invalid event
	page = baseValidPageMap()
	badEvt := baseValidEventMap()
	badEvt["stage"] = "INVALID_STAGE"
	page["events"] = []interface{}{badEvt}
	rawBadEvt, _ := json.Marshal(page)
	if err := ValidateExecutionPage(rawBadEvt); err == nil {
		t.Errorf("expected error for invalid embedded event")
	}

	// invalid change kind
	page = baseValidPageMap()
	page["changes"] = []map[string]interface{}{
		{
			"kind":          "INVALID_KIND",
			"entity_type":   "case",
			"entity_id":     "case-1",
			"summary":       "Discovered new route",
			"event_id":      "evt-1",
			"sequence":      1,
			"evidence_refs": []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
		},
	}
	rawBadChange, _ := json.Marshal(page)
	if err := ValidateExecutionPage(rawBadChange); err == nil {
		t.Errorf("expected error for invalid change kind")
	}
}

func TestValidateExecutionPage_SyntaxErrors(t *testing.T) {
	// non-UTF8
	if err := ValidateExecutionPage([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	// duplicate key
	dupJSON := `{"schema_version":"1.0","schema_version":"1.0"}`
	if err := ValidateExecutionPage([]byte(dupJSON)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	// trailing input
	validPage := baseValidPageMap()
	raw, _ := json.Marshal(validPage)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateExecutionPage(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}
