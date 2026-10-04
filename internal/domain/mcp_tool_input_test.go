package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateMCPToolInput_ValidationCases(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	prefix := "https://rewrotron.dev/specs/mcp/"
	count := 0
	for _, tc := range cases {
		if !strings.HasPrefix(tc.Schema, prefix) {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateMCPToolInput(tc.Schema, tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid MCP fixture %q to pass, got error: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid MCP fixture %q to fail validation, but it passed", tc.Name)
			}
		})
	}

	if count < 13 {
		t.Fatalf("expected at least 13 MCP validation cases, found %d", count)
	}
}

func TestValidateMCPToolInput_AllToolsValid(t *testing.T) {
	const dummyDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const dummyRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	tests := []struct {
		toolName string
		input    string
	}{
		{
			toolName: "analysis.run",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1"}`,
		},
		{
			toolName: "assessment.get",
			input:    `{"project_id":"prj_1","assessment_ref":"` + dummyRef + `"}`,
		},
		{
			toolName: "assessment.run",
			input:    `{"project_id":"prj_1","scope_ref":"` + dummyRef + `","idempotency_key":"key_1","observation":"SAFE_RUNTIME"}`,
		},
		{
			toolName: "behavior.discover",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1"}`,
		},
		{
			toolName: "behavior.get_case",
			input:    `{"project_id":"prj_1","case_id":"case_1"}`,
		},
		{
			toolName: "benchmark.run",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","unit_id":"unit_1"}`,
		},
		{
			toolName: "benchmark.run",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","all":true}`,
		},
		{
			toolName: "cutover.prepare",
			input:    `{"project_id":"prj_1","subject_hash":"` + dummyDigest + `","output":"dist","idempotency_key":"key_1"}`,
		},
		{
			toolName: "decisions.get",
			input:    `{"project_id":"prj_1","decision_id":"dec_1"}`,
		},
		{
			toolName: "decisions.propose",
			input:    `{"project_id":"prj_1","decision_ref":"` + dummyRef + `","idempotency_key":"key_1"}`,
		},
		{
			toolName: "knowledge.record_lesson",
			input:    `{"project_id":"prj_1","experiment_id":"exp_1","finding":"found bug","remediation":"fix it","idempotency_key":"key_1"}`,
		},
		{
			toolName: "migration.apply_patch",
			input:    `{"project_id":"prj_1","patch_ref":"` + dummyRef + `","idempotency_key":"key_1"}`,
		},
		{
			toolName: "migration.generate_code",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","unit_id":"unit_1"}`,
		},
		{
			toolName: "migration.generate_code",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","all":true}`,
		},
		{
			toolName: "migration.get_plan",
			input:    `{"project_id":"prj_1","plan_id":"plan_1"}`,
		},
		{
			toolName: "migration.plan",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","target":"laravel","style":"strict","recalculate":true,"selection_ref":"` + dummyRef + `","performance_policy_ref":"` + dummyRef + `"}`,
		},
		{
			toolName: "migration.stage_patch",
			input:    `{"project_id":"prj_1","patch_id":"patch_1","idempotency_key":"key_1"}`,
		},
		{
			toolName: "observation.run",
			input:    `{"project_id":"prj_1","case_id":"case_1","idempotency_key":"key_1"}`,
		},
		{
			toolName: "project.confirm_cutover",
			input:    `{"project_id":"prj_1","subject_hash":"` + dummyDigest + `","approval_token":"tok","idempotency_key":"key_1"}`,
		},
		{
			toolName: "project.confirm_plan",
			input:    `{"project_id":"prj_1","subject_hash":"` + dummyDigest + `","approval_token":"tok","idempotency_key":"key_1"}`,
		},
		{
			toolName: "project.confirm_source",
			input:    `{"project_id":"prj_1","subject_hash":"` + dummyDigest + `","approval_token":"tok","idempotency_key":"key_1"}`,
		},
		{
			toolName: "project.detect",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","path":"src","fast":true}`,
		},
		{
			toolName: "project.doctor",
			input:    `{"project_id":"prj_1"}`,
		},
		{
			toolName: "project.get_execution",
			input:    `{"project_id":"prj_1","after_sequence":1,"through_sequence":10,"limit":50}`,
		},
		{
			toolName: "project.get_snapshot",
			input:    `{"project_id":"prj_1"}`,
		},
		{
			toolName: "project.init",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1","path":"."}`,
		},
		{
			toolName: "project.prepare_source_confirmation",
			input:    `{"project_id":"prj_1","detection_id":"det_1","manifest_hash":"` + dummyDigest + `","overrides":{"key":"val"},"idempotency_key":"key_1"}`,
		},
		{
			toolName: "project.resume",
			input:    `{"project_id":"prj_1","idempotency_key":"key_1"}`,
		},
		{
			toolName: "reports.generate",
			input:    `{"project_id":"prj_1","format":"json","idempotency_key":"key_1","kind":"assessment","assessment_ref":"` + dummyRef + `"}`,
		},
		{
			toolName: "research.claim_mission",
			input:    `{"project_id":"prj_1","mission_id":"m_1","idempotency_key":"key_1"}`,
		},
		{
			toolName: "research.complete_mission",
			input:    `{"project_id":"prj_1","mission_id":"m_1","fencing_token":10,"evidence_refs":["` + dummyRef + `"],"idempotency_key":"key_1"}`,
		},
		{
			toolName: "research.get_next_mission",
			input:    `{"project_id":"prj_1"}`,
		},
		{
			toolName: "research.renew_mission",
			input:    `{"project_id":"prj_1","mission_id":"m_1","fencing_token":5,"idempotency_key":"key_1"}`,
		},
		{
			toolName: "skills.search",
			input:    `{"project_id":"prj_1","query":"php"}`,
		},
		{
			toolName: "verification.run_case",
			input:    `{"project_id":"prj_1","case_id":"case_1","idempotency_key":"key_1"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.toolName, func(t *testing.T) {
			if err := ValidateMCPToolInputByName(tc.toolName, []byte(tc.input)); err != nil {
				t.Fatalf("expected tool %s to validate input, got: %v", tc.toolName, err)
			}
		})
	}
}

func TestValidateMCPToolInput_Rejections(t *testing.T) {
	// 1. analysis.run missing project_id
	if err := ValidateMCPAnalysisRunInput([]byte(`{"idempotency_key":"key"}`)); err == nil {
		t.Errorf("expected error for missing project_id")
	}

	// 2. assessment.get non-sha256 ref
	if err := ValidateMCPAssessmentGetInput([]byte(`{"project_id":"p","assessment_ref":"invalid"}`)); err == nil {
		t.Errorf("expected error for non-sha256 assessment_ref")
	}

	// 3. assessment.run invalid observation enum
	if err := ValidateMCPAssessmentRunInput([]byte(`{"project_id":"p","idempotency_key":"k","scope_ref":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","observation":"BLIND"}`)); err == nil {
		t.Errorf("expected error for invalid observation enum")
	}

	// 4. benchmark.run both unit_id and all
	if err := ValidateMCPBenchmarkRunInput([]byte(`{"project_id":"p","idempotency_key":"k","unit_id":"u1","all":true}`)); err == nil {
		t.Errorf("expected error when both unit_id and all are present")
	}

	// 5. benchmark.run all=false
	if err := ValidateMCPBenchmarkRunInput([]byte(`{"project_id":"p","idempotency_key":"k","all":false}`)); err == nil {
		t.Errorf("expected error when all is false")
	}

	// 6. benchmark.run neither unit_id nor all
	if err := ValidateMCPBenchmarkRunInput([]byte(`{"project_id":"p","idempotency_key":"k"}`)); err == nil {
		t.Errorf("expected error when neither unit_id nor all is present")
	}

	// 7. cutover.prepare short digest
	if err := ValidateMCPCutoverPrepareInput([]byte(`{"project_id":"p","idempotency_key":"k","subject_hash":"short","output":"out"}`)); err == nil {
		t.Errorf("expected error for short subject_hash")
	}

	// 8. cutover.prepare empty output
	if err := ValidateMCPCutoverPrepareInput([]byte(`{"project_id":"p","idempotency_key":"k","subject_hash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","output":""}`)); err == nil {
		t.Errorf("expected error for empty output")
	}

	// 9. decisions.get missing decision_id
	if err := ValidateMCPDecisionsGetInput([]byte(`{"project_id":"p"}`)); err == nil {
		t.Errorf("expected error for missing decision_id")
	}

	// 10. decisions.propose invalid ref
	if err := ValidateMCPDecisionsProposeInput([]byte(`{"project_id":"p","idempotency_key":"k","decision_ref":"sha256:bad"}`)); err == nil {
		t.Errorf("expected error for invalid decision_ref")
	}

	// 11. knowledge.record_lesson empty finding
	if err := ValidateMCPKnowledgeRecordLessonInput([]byte(`{"project_id":"p","idempotency_key":"k","experiment_id":"exp","finding":"","remediation":"rem"}`)); err == nil {
		t.Errorf("expected error for empty finding")
	}

	// 12. migration.plan invalid target
	if err := ValidateMCPMigrationPlanInput([]byte(`{"project_id":"p","idempotency_key":"k","target":""}`)); err == nil {
		t.Errorf("expected error for empty target")
	}

	// 13. project.get_execution after > through
	if err := ValidateMCPProjectGetExecutionInput([]byte(`{"project_id":"p","after_sequence":10,"through_sequence":5}`)); err == nil {
		t.Errorf("expected error when after_sequence > through_sequence")
	}

	// 14. project.get_execution limit > 500
	if err := ValidateMCPProjectGetExecutionInput([]byte(`{"project_id":"p","limit":501}`)); err == nil {
		t.Errorf("expected error for limit > 500")
	}

	// 15. project.get_execution limit = 0
	if err := ValidateMCPProjectGetExecutionInput([]byte(`{"project_id":"p","limit":0}`)); err == nil {
		t.Errorf("expected error for limit = 0")
	}

	// 16. project.prepare_source_confirmation invalid overrides
	if err := ValidateMCPProjectPrepareSourceConfirmationInput([]byte(`{"project_id":"p","detection_id":"d","manifest_hash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","idempotency_key":"k","overrides":{"key":""}}`)); err == nil {
		t.Errorf("expected error for empty override value")
	}

	// 17. reports.generate kind=migration with assessment_ref
	if err := ValidateMCPReportsGenerateInput([]byte(`{"project_id":"p","format":"json","idempotency_key":"k","kind":"migration","assessment_ref":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`)); err == nil {
		t.Errorf("expected error when kind=migration is combined with assessment_ref")
	}

	// 18. research.complete_mission negative fencing token
	if err := ValidateMCPResearchCompleteMissionInput([]byte(`{"project_id":"p","mission_id":"m","fencing_token":-1,"evidence_refs":[],"idempotency_key":"k"}`)); err == nil {
		t.Errorf("expected error for negative fencing token")
	}

	// 19. skills.search empty query
	if err := ValidateMCPSkillsSearchInput([]byte(`{"project_id":"p","query":""}`)); err == nil {
		t.Errorf("expected error for empty query")
	}

	// 20. unsupported schema
	if err := ValidateMCPToolInput("https://rewrotron.dev/specs/mcp/tools/unknown.input.schema.json", []byte(`{}`)); err == nil {
		t.Errorf("expected error for unsupported MCP schema")
	}
}

func TestValidateMission_AndGetNextMission(t *testing.T) {
	validMission := `{
		"mission_id": "m_1",
		"type": "ANALYZE_SOURCE",
		"priority": 1,
		"status": "QUEUED",
		"unit_id": null,
		"case_id": null,
		"description": "analyze source tree",
		"context": {"key":"value"},
		"attempt_count": 0,
		"fencing_token": 0,
		"lease_expires_at": "2026-10-04T12:00:00Z"
	}`

	if err := ValidateMission([]byte(validMission)); err != nil {
		t.Fatalf("expected valid mission to pass, got: %v", err)
	}

	// Test GetNextMission with valid embedded mission
	getNextMissionValid := `{
		"mission": ` + validMission + `,
		"reason": "READY"
	}`
	if err := ValidateGetNextMission([]byte(getNextMissionValid)); err != nil {
		t.Fatalf("expected valid get-next-mission to pass, got: %v", err)
	}

	// Test GetNextMission with null mission and NO_WORK
	getNextMissionNull := `{
		"mission": null,
		"reason": "NO_WORK"
	}`
	if err := ValidateGetNextMission([]byte(getNextMissionNull)); err != nil {
		t.Fatalf("expected valid get-next-mission with null mission to pass, got: %v", err)
	}

	// Rejection: invalid reason
	badReason := `{
		"mission": null,
		"reason": "INVALID_REASON"
	}`
	if err := ValidateGetNextMission([]byte(badReason)); err == nil {
		t.Errorf("expected error for invalid reason")
	}

	// Rejection: invalid mission status
	badStatusMission := strings.Replace(validMission, `"QUEUED"`, `"NOT_A_STATUS"`, 1)
	if err := ValidateMission([]byte(badStatusMission)); err == nil {
		t.Errorf("expected error for invalid status enum")
	}

	// Rejection: invalid date-time
	badDateMission := strings.Replace(validMission, `"2026-10-04T12:00:00Z"`, `"yesterday"`, 1)
	if err := ValidateMission([]byte(badDateMission)); err == nil {
		t.Errorf("expected error for invalid lease_expires_at date-time")
	}
}

func TestValidateMCPToolInput_PrivacyCanary(t *testing.T) {
	const canary = "CANARY_SECRET_LEAK_DO_NOT_EXPOSE"

	badInputs := []struct {
		fn   func([]byte) error
		json string
	}{
		{
			fn:   ValidateMCPAnalysisRunInput,
			json: `{"project_id":"` + canary + `","idempotency_key":"key"}`,
		},
		{
			fn:   ValidateMCPProjectConfirmPlanInput,
			json: `{"project_id":"prj","subject_hash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","approval_token":"` + canary + `","idempotency_key":"key","extra":1}`,
		},
		{
			fn:   ValidateMCPSkillsSearchInput,
			json: `{"project_id":"prj","query":"","canary":"` + canary + `"}`,
		},
	}

	for _, bi := range badInputs {
		err := bi.fn([]byte(bi.json))
		if err == nil {
			t.Fatalf("expected validation error, got nil")
		}
		if strings.Contains(err.Error(), canary) {
			t.Fatalf("error message leaks secret canary: %s", err.Error())
		}
	}
}
