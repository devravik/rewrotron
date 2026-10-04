package domain

import (
	"encoding/json"
	"testing"
)

func baseValidSnapshotMap() map[string]interface{} {
	return map[string]interface{}{
		"project_id":         "prj-1",
		"state":              "VERIFIED",
		"revision":           1,
		"configuration_hash": "1111111111111111111111111111111111111111111111111111111111111111",
		"plan_hash":          "2222222222222222222222222222222222222222222222222222222222222222",
		"gate_states": map[string]interface{}{
			"GATE_1_SOURCE_DETECTION": "VALID",
			"GATE_2_PLAN_CONFIRMED":   "VALID",
			"GATE_3_CUTOVER":          "PENDING",
		},
		"coverage": map[string]interface{}{
			"entrypoints": map[string]interface{}{
				"numerator":   10,
				"denominator": 10,
				"percent":     100.0,
			},
			"cases": map[string]interface{}{
				"numerator":   100,
				"denominator": 100,
				"percent":     100.0,
			},
			"units": map[string]interface{}{
				"numerator":   20,
				"denominator": 20,
				"percent":     100.0,
			},
			"dimensions": map[string]interface{}{
				"numerator":   4,
				"denominator": 4,
				"percent":     100.0,
			},
		},
		"blockers":            []interface{}{},
		"latest_operation_id": "op-verify",
	}
}

func baseValidMigrationReportMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":     "1.0",
		"project_id":         "prj-1",
		"configuration_hash": "1111111111111111111111111111111111111111111111111111111111111111",
		"plan_hash":          "2222222222222222222222222222222222222222222222222222222222222222",
		"target_tree_hash":   "3333333333333333333333333333333333333333333333333333333333333333",
		"snapshot":           baseValidSnapshotMap(),
		"inventory_dispositions": []map[string]interface{}{
			{
				"entrypoint_id": "ep-login",
				"disposition":   "MIGRATED",
				"unit_id":       "unit-auth",
				"evidence_refs": []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
			},
		},
		"verification_report_refs": []string{"sha256:4444444444444444444444444444444444444444444444444444444444444444"},
		"benchmark_report_refs":    []string{"sha256:5555555555555555555555555555555555555555555555555555555555555555"},
		"omission_report_ref":      "sha256:6666666666666666666666666666666666666666666666666666666666666666",
		"intentional_changes": []map[string]interface{}{
			{
				"change_id": "ic-1",
			},
		},
		"limitations": []string{
			"No live payment gateway characterization",
		},
		"required_decisions": []string{
			"Approve Gate 3 cutover",
		},
	}
}

func TestValidateMigrationReport_Valid(t *testing.T) {
	rep := baseValidMigrationReportMap()
	raw, err := json.Marshal(rep)
	if err != nil {
		t.Fatalf("failed to marshal migration report: %v", err)
	}
	if err := ValidateMigrationReport(raw); err != nil {
		t.Errorf("expected valid migration report, got %v", err)
	}

	// with optional release_ref
	rep["release_ref"] = "sha256:7777777777777777777777777777777777777777777777777777777777777777"
	rawWithRelease, err := json.Marshal(rep)
	if err != nil {
		t.Fatalf("failed to marshal migration report with release_ref: %v", err)
	}
	if err := ValidateMigrationReport(rawWithRelease); err != nil {
		t.Errorf("expected valid migration report with release_ref, got %v", err)
	}
}

func TestValidateMigrationReport_RequiredFields(t *testing.T) {
	required := []string{
		"schema_version",
		"project_id",
		"configuration_hash",
		"plan_hash",
		"target_tree_hash",
		"snapshot",
		"inventory_dispositions",
		"verification_report_refs",
		"benchmark_report_refs",
		"omission_report_ref",
		"intentional_changes",
		"limitations",
		"required_decisions",
	}

	for _, f := range required {
		rep := baseValidMigrationReportMap()
		delete(rep, f)
		raw, _ := json.Marshal(rep)
		if err := ValidateMigrationReport(raw); err == nil {
			t.Errorf("expected error when missing required field %q", f)
		}
	}
}

func TestValidateMigrationReport_EmbeddedSnapshotErrors(t *testing.T) {
	rep := baseValidMigrationReportMap()
	snap := baseValidSnapshotMap()
	snap["state"] = "INVALID_STATE"
	rep["snapshot"] = snap
	raw, _ := json.Marshal(rep)
	if err := ValidateMigrationReport(raw); err == nil {
		t.Errorf("expected error for invalid embedded snapshot")
	}
}

func TestValidateMigrationReport_InventoryDispositionValidation(t *testing.T) {
	// invalid disposition enum
	rep := baseValidMigrationReportMap()
	rep["inventory_dispositions"] = []map[string]interface{}{
		{
			"entrypoint_id": "ep-1",
			"disposition":   "UNKNOWN_DISPOSITION",
			"evidence_refs": []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
		},
	}
	rawBadDisp, _ := json.Marshal(rep)
	if err := ValidateMigrationReport(rawBadDisp); err == nil {
		t.Errorf("expected error for invalid disposition enum")
	}

	// invalid evidence_ref
	rep = baseValidMigrationReportMap()
	rep["inventory_dispositions"] = []map[string]interface{}{
		{
			"entrypoint_id": "ep-1",
			"disposition":   "MIGRATED",
			"evidence_refs": []string{"not-a-sha256"},
		},
	}
	rawBadRef, _ := json.Marshal(rep)
	if err := ValidateMigrationReport(rawBadRef); err == nil {
		t.Errorf("expected error for invalid evidence_refs item")
	}
}

func TestValidateMigrationReport_SyntaxErrors(t *testing.T) {
	if err := ValidateMigrationReport([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	dupJSON := `{"schema_version":"1.0","schema_version":"1.0"}`
	if err := ValidateMigrationReport([]byte(dupJSON)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	rep := baseValidMigrationReportMap()
	raw, _ := json.Marshal(rep)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateMigrationReport(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}
