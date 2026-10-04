package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateMigrationProgress_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/migration/progress.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateMigrationProgress(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid migration-progress fixture %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid migration-progress fixture %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count != 1 {
		t.Fatalf("expected 1 migration-progress case in validation-cases.json, found %d", count)
	}
}

func TestValidateMigrationProgress_Rejections(t *testing.T) {
	const validHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const validRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	baseValid := func() map[string]any {
		return map[string]any{
			"schema_version":        "2.0",
			"project_id":            "fixture",
			"plan_hash":             validHash,
			"decision_or_model_ref": validRef,
			"entity_id":             "login",
			"status":                "VERIFYING",
			"through_sequence":      1,
			"target_tree_ref":       validRef,
			"acceptance_report_ref": nil,
		}
	}

	// 1. Valid baseline passes
	baseBytes, _ := json.Marshal(baseValid())
	if err := ValidateMigrationProgress(baseBytes); err != nil {
		t.Fatalf("expected baseline migration progress to pass, got: %v", err)
	}

	// Also test with target_tree_ref null and acceptance_report_ref non-null
	mOther := baseValid()
	mOther["target_tree_ref"] = nil
	mOther["acceptance_report_ref"] = validRef
	rawOther, _ := json.Marshal(mOther)
	if err := ValidateMigrationProgress(rawOther); err != nil {
		t.Fatalf("expected migration progress with null target_tree_ref to pass, got: %v", err)
	}

	// 2. Missing required fields
	for k := range baseValid() {
		m := baseValid()
		delete(m, k)
		raw, _ := json.Marshal(m)
		if err := ValidateMigrationProgress(raw); err == nil {
			t.Errorf("expected error for missing field %q", k)
		}
	}

	// 3. Null on non-nullable fields
	for _, field := range []string{"schema_version", "project_id", "plan_hash", "decision_or_model_ref", "entity_id", "status", "through_sequence"} {
		m := baseValid()
		m[field] = nil
		raw, _ := json.Marshal(m)
		if err := ValidateMigrationProgress(raw); err == nil {
			t.Errorf("expected error for null field %q", field)
		}
	}

	// 4. Invalid schema_version
	m := baseValid()
	m["schema_version"] = "1.0"
	raw, _ := json.Marshal(m)
	if err := ValidateMigrationProgress(raw); err == nil {
		t.Errorf("expected error for invalid schema_version")
	}

	// 5. Invalid project_id
	m = baseValid()
	m["project_id"] = "PRJ UPPER"
	raw, _ = json.Marshal(m)
	if err := ValidateMigrationProgress(raw); err == nil {
		t.Errorf("expected error for invalid project_id")
	}

	// 6. Invalid status enum
	m = baseValid()
	m["status"] = "INVALID_STATUS"
	raw, _ = json.Marshal(m)
	if err := ValidateMigrationProgress(raw); err == nil {
		t.Errorf("expected error for invalid status")
	}

	// 7. Negative through_sequence
	m = baseValid()
	m["through_sequence"] = -1
	raw, _ = json.Marshal(m)
	if err := ValidateMigrationProgress(raw); err == nil {
		t.Errorf("expected error for negative through_sequence")
	}

	// 8. Float through_sequence
	floatInput := `{"schema_version":"2.0","project_id":"fixture","plan_hash":"` + validHash + `","decision_or_model_ref":"` + validRef + `","entity_id":"login","status":"VERIFYING","through_sequence":1.5,"target_tree_ref":null,"acceptance_report_ref":null}`
	if err := ValidateMigrationProgress([]byte(floatInput)); err == nil {
		t.Errorf("expected error for float through_sequence")
	}

	// 9. Unexpected fields
	m = baseValid()
	m["unexpected"] = "field"
	raw, _ = json.Marshal(m)
	if err := ValidateMigrationProgress(raw); err == nil {
		t.Errorf("expected error for unexpected field")
	}
}

func TestValidateMigrationProgress_PrivacyCanary(t *testing.T) {
	const canary = "CANARY_SECRET_LEAK_MIGRATION_PROGRESS"

	badInput := `{"schema_version":"2.0","project_id":"` + canary + `","plan_hash":"invalid"}`
	err := ValidateMigrationProgress([]byte(badInput))
	if err == nil {
		t.Fatalf("expected error for invalid migration progress")
	}
	if strings.Contains(err.Error(), canary) {
		t.Fatalf("error message leaks secret canary: %s", err.Error())
	}
}
