package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type validationFixtureCase struct {
	Name     string          `json:"name"`
	Schema   string          `json:"schema"`
	Instance json.RawMessage `json:"instance"`
	Valid    bool            `json:"valid"`
}

func TestValidateExecutionEvent_Fixtures(t *testing.T) {
	fixtureData, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases fixture: %v", err)
	}

	var cases []validationFixtureCase
	if err := json.Unmarshal(fixtureData, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	testedEvents := 0
	testedCausality := 0
	for _, tc := range cases {
		if tc.Schema == "https://rewrotron.dev/specs/execution/event.schema.json" {
			testedEvents++
			err := ValidateExecutionEvent(tc.Instance)
			if tc.Valid && err != nil {
				t.Errorf("fixture %q expected valid, got error: %v", tc.Name, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("fixture %q expected invalid, got nil error", tc.Name)
			}
		}
		if tc.Schema == "https://rewrotron.dev/specs/observation/causality.schema.json" {
			testedCausality++
			err := ValidateExecutionCausality(tc.Instance)
			if tc.Valid && err != nil {
				t.Errorf("fixture %q expected valid causality, got error: %v", tc.Name, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("fixture %q expected invalid causality, got nil error", tc.Name)
			}
		}
	}

	if testedEvents < 8 {
		t.Fatalf("expected at least 8 execution event fixtures, got %d", testedEvents)
	}
	if testedCausality < 1 {
		t.Fatalf("expected at least 1 causality fixture, got %d", testedCausality)
	}
}

func baseValidEventMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":     "1.0",
		"event_id":           "evt-1",
		"sequence":           1,
		"project_id":         "prj-1",
		"run_id":             "run-1",
		"operation_id":       "op-1",
		"mission_id":         "m-1",
		"task_id":            "t-1",
		"stage":              "OBSERVE",
		"stage_instance_id":  "stage-1",
		"type":               "TASK_STARTED",
		"kind":               "ACTION",
		"severity":           "INFO",
		"producer":           "ENGINE",
		"requires_user":      false,
		"created_at":         "2026-10-04T12:00:00Z",
		"causation_event_id": nil,
		"title":              "Task started",
		"summary":            "Observing behavior",
		"reason":             "Characterizing legacy routes",
		"result":             nil,
		"next_action":        "Run test case",
		"user_request":       nil,
		"entity_type":        "case",
		"entity_id":          "case-1",
		"subjects": map[string]interface{}{
			"configuration_hash": nil,
			"plan_hash":          nil,
			"target_spec_hash":   nil,
			"target_tree_hash":   nil,
			"corpus_hash":        nil,
			"policy_hash":        nil,
		},
		"evidence_refs": []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
		"artifact_refs": []string{},
		"progress":      nil,
		"approval":      nil,
		"details":       map[string]interface{}{},
	}
}

func baseValidCausalityMap() map[string]interface{} {
	return map[string]interface{}{
		"project_id":           "prj-1",
		"run_id":               "run-1",
		"workflow_instance_id": "wf-1",
		"operation_id":         "op-1",
		"request_id":           "req-1",
		"process_id":           "proc-1",
		"trace_id":             "trace-1",
		"span_id":              "span-1",
		"parent_span_id":       nil,
		"cause_ids":            []string{"cause-1", "cause-2"},
		"job_id":               nil,
		"effect_id":            nil,
	}
}

func TestValidateExecutionEvent_Valid10And20(t *testing.T) {
	evt10 := baseValidEventMap()
	raw10, err := json.Marshal(evt10)
	if err != nil {
		t.Fatalf("failed to marshal 1.0 event: %v", err)
	}
	if err := ValidateExecutionEvent(raw10); err != nil {
		t.Errorf("expected valid 1.0 event, got %v", err)
	}

	evt20 := baseValidEventMap()
	evt20["schema_version"] = "2.0"
	evt20["causality"] = baseValidCausalityMap()
	raw20, err := json.Marshal(evt20)
	if err != nil {
		t.Fatalf("failed to marshal 2.0 event: %v", err)
	}
	if err := ValidateExecutionEvent(raw20); err != nil {
		t.Errorf("expected valid 2.0 event, got %v", err)
	}
}

func TestValidateExecutionEvent_SchemaVersionErrors(t *testing.T) {
	evt := baseValidEventMap()
	evt["schema_version"] = "3.0"
	raw, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for unsupported schema_version 3.0")
	}

	// 1.0 with causality should fail because additionalProperties: false
	evt10Extra := baseValidEventMap()
	evt10Extra["causality"] = baseValidCausalityMap()
	rawExtra, _ := json.Marshal(evt10Extra)
	if err := ValidateExecutionEvent(rawExtra); err == nil {
		t.Errorf("expected error for 1.0 event with unexpected causality field")
	}

	// 2.0 without causality should fail
	evt20Missing := baseValidEventMap()
	evt20Missing["schema_version"] = "2.0"
	rawMissing, _ := json.Marshal(evt20Missing)
	if err := ValidateExecutionEvent(rawMissing); err == nil {
		t.Errorf("expected error for 2.0 event missing causality")
	}
}

func TestValidateExecutionEvent_SequenceConstraints(t *testing.T) {
	// sequence = 0
	evtZero := baseValidEventMap()
	evtZero["sequence"] = 0
	rawZero, _ := json.Marshal(evtZero)
	if err := ValidateExecutionEvent(rawZero); err == nil {
		t.Errorf("expected error for sequence 0")
	}

	// sequence = -1
	evtNeg := baseValidEventMap()
	evtNeg["sequence"] = -1
	rawNeg, _ := json.Marshal(evtNeg)
	if err := ValidateExecutionEvent(rawNeg); err == nil {
		t.Errorf("expected error for sequence -1")
	}

	// sequence float
	rawFloat := []byte(strings.Replace(string(rawZero), `"sequence":0`, `"sequence":1.5`, 1))
	if err := ValidateExecutionEvent(rawFloat); err == nil {
		t.Errorf("expected error for fractional sequence 1.5")
	}
}

func TestValidateExecutionEvent_RequiresUserInvariants(t *testing.T) {
	evt := baseValidEventMap()
	evt["requires_user"] = true
	evt["user_request"] = nil
	raw, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error when requires_user is true and user_request is null")
	}

	evt["user_request"] = ""
	rawEmpty, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(rawEmpty); err == nil {
		t.Errorf("expected error when requires_user is true and user_request is empty")
	}

	evt["user_request"] = "Please confirm new port mapping"
	rawValid, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(rawValid); err != nil {
		t.Errorf("expected valid event with user_request, got %v", err)
	}
}

func TestValidateExecutionEvent_ApprovalInvariants(t *testing.T) {
	// APPROVAL_REQUIRED without approval
	evtReq := baseValidEventMap()
	evtReq["type"] = "APPROVAL_REQUIRED"
	evtReq["approval"] = nil
	rawReq, _ := json.Marshal(evtReq)
	if err := ValidateExecutionEvent(rawReq); err == nil {
		t.Errorf("expected error for APPROVAL_REQUIRED without approval object")
	}

	// APPROVAL_REQUIRED with approval and null confirmation_id
	evtReq["approval"] = map[string]interface{}{
		"gate_type":       "GATE_1_SOURCE_DETECTION",
		"subject_hash":    "1111111111111111111111111111111111111111111111111111111111111111",
		"confirmation_id": nil,
	}
	rawReqValid, _ := json.Marshal(evtReq)
	if err := ValidateExecutionEvent(rawReqValid); err != nil {
		t.Errorf("expected valid APPROVAL_REQUIRED with null confirmation_id, got %v", err)
	}

	// APPROVAL_RECEIVED without approval
	evtRec := baseValidEventMap()
	evtRec["type"] = "APPROVAL_RECEIVED"
	evtRec["approval"] = nil
	rawRec, _ := json.Marshal(evtRec)
	if err := ValidateExecutionEvent(rawRec); err == nil {
		t.Errorf("expected error for APPROVAL_RECEIVED without approval object")
	}

	// APPROVAL_RECEIVED with null confirmation_id
	evtRec["approval"] = map[string]interface{}{
		"gate_type":       "GATE_2_PLAN_CONFIRMED",
		"subject_hash":    "1111111111111111111111111111111111111111111111111111111111111111",
		"confirmation_id": nil,
	}
	rawRecNullConf, _ := json.Marshal(evtRec)
	if err := ValidateExecutionEvent(rawRecNullConf); err == nil {
		t.Errorf("expected error for APPROVAL_RECEIVED with null confirmation_id")
	}

	// APPROVAL_RECEIVED with confirmation_id but producer AGENT
	evtRec["producer"] = "AGENT"
	evtRec["approval"] = map[string]interface{}{
		"gate_type":       "GATE_2_PLAN_CONFIRMED",
		"subject_hash":    "1111111111111111111111111111111111111111111111111111111111111111",
		"confirmation_id": "conf-123",
	}
	rawRecAgent, _ := json.Marshal(evtRec)
	if err := ValidateExecutionEvent(rawRecAgent); err == nil {
		t.Errorf("expected error for APPROVAL_RECEIVED with producer AGENT")
	}

	// APPROVAL_RECEIVED with confirmation_id and producer ENGINE
	evtRec["producer"] = "ENGINE"
	rawRecValid, _ := json.Marshal(evtRec)
	if err := ValidateExecutionEvent(rawRecValid); err != nil {
		t.Errorf("expected valid APPROVAL_RECEIVED with producer ENGINE, got %v", err)
	}
}

func TestValidateExecutionEvent_ProgressInvariants(t *testing.T) {
	evt := baseValidEventMap()
	evt["type"] = "PROGRESS_UPDATED"
	evt["progress"] = nil
	rawNull, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(rawNull); err == nil {
		t.Errorf("expected error for PROGRESS_UPDATED with null progress")
	}

	// valid progress
	evt["progress"] = map[string]interface{}{
		"metric_id":           "metric-test",
		"definition":          "Testing progress update",
		"numerator":           10,
		"denominator":         20,
		"unit":                "units",
		"percent":             50.0,
		"scope_hash":          "1111111111111111111111111111111111111111111111111111111111111111",
		"basis_ref":           "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"accounting_revision": 1,
		"evidence_sequence":   5,
		"freshness":           "CURRENT",
	}
	rawValid, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(rawValid); err != nil {
		t.Errorf("expected valid PROGRESS_UPDATED with progress, got %v", err)
	}
}

func TestValidateExecutionEvent_ResultInvariants(t *testing.T) {
	for _, typ := range []string{"ACTION_COMPLETED", "TASK_COMPLETED", "STAGE_COMPLETED"} {
		evt := baseValidEventMap()
		evt["type"] = typ
		evt["result"] = nil
		rawNull, _ := json.Marshal(evt)
		if err := ValidateExecutionEvent(rawNull); err == nil {
			t.Errorf("expected error for %s with null result", typ)
		}

		evt["result"] = ""
		rawEmpty, _ := json.Marshal(evt)
		if err := ValidateExecutionEvent(rawEmpty); err == nil {
			t.Errorf("expected error for %s with empty string result", typ)
		}

		evt["result"] = "Operation completed successfully"
		rawValid, _ := json.Marshal(evt)
		if err := ValidateExecutionEvent(rawValid); err != nil {
			t.Errorf("expected valid for %s with non-empty result, got %v", typ, err)
		}
	}
}

func TestValidateExecutionEvent_EnumAndFieldValidation(t *testing.T) {
	// invalid stage
	evt := baseValidEventMap()
	evt["stage"] = "INVALID_STAGE"
	raw, _ := json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid stage")
	}

	// invalid type
	evt = baseValidEventMap()
	evt["type"] = "INVALID_TYPE"
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid type")
	}

	// invalid kind
	evt = baseValidEventMap()
	evt["kind"] = "INVALID_KIND"
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid kind")
	}

	// invalid severity
	evt = baseValidEventMap()
	evt["severity"] = "INVALID_SEVERITY"
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid severity")
	}

	// invalid producer
	evt = baseValidEventMap()
	evt["producer"] = "INVALID_PRODUCER"
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid producer")
	}

	// invalid created_at
	evt = baseValidEventMap()
	evt["created_at"] = "not-a-timestamp"
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid created_at")
	}

	// invalid details (not object)
	evt = baseValidEventMap()
	evt["details"] = []string{"not an object"}
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for non-object details")
	}

	// invalid evidence_ref item
	evt = baseValidEventMap()
	evt["evidence_refs"] = []string{"not-a-sha256"}
	raw, _ = json.Marshal(evt)
	if err := ValidateExecutionEvent(raw); err == nil {
		t.Errorf("expected error for invalid evidence_refs")
	}
}

func TestValidateExecutionEvent_JSONSyntaxErrors(t *testing.T) {
	// invalid UTF-8
	if err := ValidateExecutionEvent([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	// duplicate key
	dupJSON := `{"schema_version":"1.0","schema_version":"1.0"}`
	if err := ValidateExecutionEvent([]byte(dupJSON)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	// trailing input
	validEvt := baseValidEventMap()
	raw, _ := json.Marshal(validEvt)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateExecutionEvent(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}

func TestValidateExecutionCausality_Direct(t *testing.T) {
	cMap := baseValidCausalityMap()
	raw, err := json.Marshal(cMap)
	if err != nil {
		t.Fatalf("failed to marshal causality: %v", err)
	}
	if err := ValidateExecutionCausality(raw); err != nil {
		t.Errorf("expected valid causality, got %v", err)
	}

	// duplicate cause_id
	cMap["cause_ids"] = []string{"id-1", "id-1"}
	rawDup, _ := json.Marshal(cMap)
	if err := ValidateExecutionCausality(rawDup); err == nil {
		t.Errorf("expected error for duplicate cause_ids")
	}

	// invalid ID
	cMap["cause_ids"] = []string{"valid-id"}
	cMap["project_id"] = "Invalid Project ID!"
	rawBadID, _ := json.Marshal(cMap)
	if err := ValidateExecutionCausality(rawBadID); err == nil {
		t.Errorf("expected error for invalid project_id in causality")
	}

	// missing field
	delete(cMap, "project_id")
	rawMissing, _ := json.Marshal(cMap)
	if err := ValidateExecutionCausality(rawMissing); err == nil {
		t.Errorf("expected error for missing required causality field")
	}
}
