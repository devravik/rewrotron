package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateCLICommandInput_ValidationCases(t *testing.T) {
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

	prefix := "https://rewrotron.dev/specs/cli/commands/"
	count := 0
	for _, tc := range cases {
		if !strings.HasPrefix(tc.Schema, prefix) {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateCLICommandInput(tc.Schema, tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid CLI command fixture %q to pass, got error: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid CLI command fixture %q to fail validation, but it passed", tc.Name)
			}
		})
	}

	if count < 15 {
		t.Fatalf("expected at least 15 CLI command validation cases, found %d", count)
	}
}

func TestValidateCLICommandInput_Rejections(t *testing.T) {
	// Report command: missing format
	if err := ValidateCLIReportInput([]byte(`{}`)); err == nil {
		t.Errorf("expected error for missing format in report input")
	}

	// Report command: invalid format
	if err := ValidateCLIReportInput([]byte(`{"format":"pdf"}`)); err == nil {
		t.Errorf("expected error for invalid format enum in report input")
	}

	// Assess command: missing scope_ref
	if err := ValidateCLIAssessInput([]byte(`{}`)); err == nil {
		t.Errorf("expected error for missing scope_ref in assess input")
	}

	// Assessment show command: missing assessment_ref
	if err := ValidateCLIAssessmentShowInput([]byte(`{}`)); err == nil {
		t.Errorf("expected error for missing assessment_ref")
	}

	// Assessment decide: missing rationale
	badDecide := `{"assessment_ref":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","action":"STOP"}`
	if err := ValidateCLIAssessmentDecideInput([]byte(badDecide)); err == nil {
		t.Errorf("expected error for missing rationale in assessment decide")
	}

	// Status command: invalid verbosity
	if err := ValidateCLIStatusInput([]byte(`{"verbosity":"ultra"}`)); err == nil {
		t.Errorf("expected error for invalid verbosity")
	}

	// Execution show: limit > 500
	if err := ValidateCLIExecutionShowInput([]byte(`{"limit":501}`)); err == nil {
		t.Errorf("expected error for limit > 500")
	}

	// Pause command: unexpected field
	if err := ValidateCLIPauseInput([]byte(`{"now":true}`)); err == nil {
		t.Errorf("expected error for unexpected field in pause input")
	}

	// Continue command: unexpected field
	if err := ValidateCLIContinueInput([]byte(`{"actor":"human"}`)); err == nil {
		t.Errorf("expected error for unexpected field in continue input")
	}

	// Migrate command: both unit_id and all
	if err := ValidateCLIMigrateInput([]byte(`{"unit_id":"auth","all":true}`)); err == nil {
		t.Errorf("expected error when both unit_id and all are set in migrate input")
	}

	// Migrate command: neither unit_id nor all
	if err := ValidateCLIMigrateInput([]byte(`{}`)); err == nil {
		t.Errorf("expected error when neither unit_id nor all is set in migrate input")
	}

	// Approval issue: invalid gate enum
	if err := ValidateCLIApprovalIssueInput([]byte(`{"gate":"invalid","subject_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)); err == nil {
		t.Errorf("expected error for invalid gate in approval issue")
	}

	// Confirm source: missing detection_id
	if err := ValidateCLIConfirmSourceInput([]byte(`{"manifest_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)); err == nil {
		t.Errorf("expected error for missing detection_id in confirm source")
	}

	// Unsupported schema
	if err := ValidateCLICommandInput("https://rewrotron.dev/specs/cli/commands/unknown.schema.json", []byte(`{}`)); err == nil {
		t.Errorf("expected error for unsupported command schema")
	}
}

func TestValidateCLICommandInput_SyntaxErrors(t *testing.T) {
	if err := ValidateCLIReportInput([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	dup := `{"format":"json","format":"json"}`
	if err := ValidateCLIReportInput([]byte(dup)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	trailing := `{"format":"json"} "extra"`
	if err := ValidateCLIReportInput([]byte(trailing)); err == nil {
		t.Errorf("expected error for trailing input")
	}
}
