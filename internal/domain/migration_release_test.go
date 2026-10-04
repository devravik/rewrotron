package domain

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func baseValidReleaseSubjectMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":           "2.0",
		"canonicalization":         "rwt-cjson-1",
		"project_id":               "fixture",
		"kind":                     "RELEASE",
		"configuration_hash":       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"plan_hash":                "3ccae952a8f049ec692f413163bfb5828b25dcf42a1f0d80eac7cb6a121bcb2e",
		"target_tree_ref":          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"whole_project_report_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"benchmark_ref":            "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"runbook_ref":              "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"rollback_rehearsal_ref":   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
}

func baseValidMigrationReleaseV1Map() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":         "1.0",
		"project_id":             "fixture",
		"release_hash":           "82c45c5409d976c4a942170e430c336a1acd468278a33ecb4afc60ab0612f1db",
		"configuration_hash":     "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"plan_hash":              "3ccae952a8f049ec692f413163bfb5828b25dcf42a1f0d80eac7cb6a121bcb2e",
		"target_tree_hash":       "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"dependency_lock_hashes": []string{"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"},
		"verification_report_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"benchmark_report_ref":    "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"cutover_runbook_ref":     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"rollback_runbook_ref":    "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"session_strategy_ref":    "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"scope_exclusions": []map[string]interface{}{
			{"unit_id": "unit-excluded"},
		},
		"omission_report_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
}

func TestValidateReleaseApprovalSubjectShape_Golden(t *testing.T) {
	var vectors []struct {
		Name      string          `json:"name"`
		Subject   json.RawMessage `json:"subject"`
		Canonical string          `json:"canonical_utf8"`
		Hash      string          `json:"sha256"`
	}
	loadVectors(t, "../../specs/examples/approval-hash-vectors.json", &vectors)

	found := false
	for _, v := range vectors {
		if v.Name != "release" {
			continue
		}
		found = true
		before := bytes.Clone(v.Subject)
		if err := ValidateReleaseApprovalSubjectShape(v.Subject); err != nil {
			t.Fatalf("expected golden release subject to validate, got: %v", err)
		}
		assertGolden(t, v.Subject, v.Canonical, v.Hash)
		if !bytes.Equal(before, v.Subject) {
			t.Fatal("ValidateReleaseApprovalSubjectShape mutated input bytes")
		}
	}
	if !found {
		t.Fatal("missing 'release' vector in approval-hash-vectors.json")
	}
}

func TestValidateReleaseApprovalSubjectShape_Rejections(t *testing.T) {
	required := []string{
		"schema_version",
		"canonicalization",
		"project_id",
		"kind",
		"configuration_hash",
		"plan_hash",
		"target_tree_ref",
		"whole_project_report_ref",
		"benchmark_ref",
		"runbook_ref",
		"rollback_rehearsal_ref",
	}

	for _, f := range required {
		m := baseValidReleaseSubjectMap()
		delete(m, f)
		raw, _ := json.Marshal(m)
		if err := ValidateReleaseApprovalSubjectShape(raw); err == nil {
			t.Errorf("expected error when missing required field %q", f)
		}
	}

	// invalid schema_version
	m := baseValidReleaseSubjectMap()
	m["schema_version"] = "1.0"
	raw, _ := json.Marshal(m)
	if err := ValidateReleaseApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for invalid schema_version")
	}

	// invalid kind
	m = baseValidReleaseSubjectMap()
	m["kind"] = "PLAN"
	raw, _ = json.Marshal(m)
	if err := ValidateReleaseApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for invalid kind")
	}

	// invalid configuration_hash
	m = baseValidReleaseSubjectMap()
	m["configuration_hash"] = "not-a-sha256"
	raw, _ = json.Marshal(m)
	if err := ValidateReleaseApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for invalid configuration_hash")
	}

	// invalid ref
	m = baseValidReleaseSubjectMap()
	m["target_tree_ref"] = "bad-ref"
	raw, _ = json.Marshal(m)
	if err := ValidateReleaseApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for invalid target_tree_ref")
	}

	// unexpected field
	m = baseValidReleaseSubjectMap()
	m["extra"] = "value"
	raw, _ = json.Marshal(m)
	if err := ValidateReleaseApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for unexpected field")
	}
}

func TestValidateMigrationRelease_V1_Valid(t *testing.T) {
	m := baseValidMigrationReleaseV1Map()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal v1 migration release: %v", err)
	}
	if err := ValidateMigrationRelease(raw); err != nil {
		t.Fatalf("expected valid v1 migration release, got: %v", err)
	}
}

func TestValidateMigrationRelease_V1_Rejections(t *testing.T) {
	required := []string{
		"schema_version",
		"project_id",
		"release_hash",
		"configuration_hash",
		"plan_hash",
		"target_tree_hash",
		"dependency_lock_hashes",
		"verification_report_ref",
		"benchmark_report_ref",
		"cutover_runbook_ref",
		"rollback_runbook_ref",
		"session_strategy_ref",
		"scope_exclusions",
		"omission_report_ref",
	}

	for _, f := range required {
		m := baseValidMigrationReleaseV1Map()
		delete(m, f)
		raw, _ := json.Marshal(m)
		if err := ValidateMigrationRelease(raw); err == nil {
			t.Errorf("expected error when missing required field %q", f)
		}
	}

	// invalid hash
	m := baseValidMigrationReleaseV1Map()
	m["release_hash"] = "invalid"
	raw, _ := json.Marshal(m)
	if err := ValidateMigrationRelease(raw); err == nil {
		t.Errorf("expected error for invalid release_hash")
	}

	// invalid ref
	m = baseValidMigrationReleaseV1Map()
	m["cutover_runbook_ref"] = "invalid"
	raw, _ = json.Marshal(m)
	if err := ValidateMigrationRelease(raw); err == nil {
		t.Errorf("expected error for invalid cutover_runbook_ref")
	}
}

func TestValidateMigrationRelease_V2_Valid(t *testing.T) {
	subjMap := baseValidReleaseSubjectMap()
	subjRaw, err := json.Marshal(subjMap)
	if err != nil {
		t.Fatalf("failed to marshal subject: %v", err)
	}
	expectedHash, err := CanonicalHash(subjRaw)
	if err != nil {
		t.Fatalf("failed to compute canonical hash: %v", err)
	}

	envelope := map[string]interface{}{
		"schema_version": "2.0",
		"project_id":     "fixture",
		"release_hash":   expectedHash,
		"subject":        subjMap,
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("failed to marshal v2 release: %v", err)
	}
	if err := ValidateMigrationRelease(raw); err != nil {
		t.Fatalf("expected valid v2 migration release, got: %v", err)
	}
}

func TestValidateMigrationRelease_V2_Rejections(t *testing.T) {
	subjMap := baseValidReleaseSubjectMap()
	subjRaw, _ := json.Marshal(subjMap)
	expectedHash, _ := CanonicalHash(subjRaw)

	// project_id mismatch
	envelope := map[string]interface{}{
		"schema_version": "2.0",
		"project_id":     "other-project",
		"release_hash":   expectedHash,
		"subject":        subjMap,
	}
	raw, _ := json.Marshal(envelope)
	if err := ValidateMigrationRelease(raw); err == nil || !strings.Contains(err.Error(), "project_id") {
		t.Errorf("expected project_id mismatch error, got: %v", err)
	}

	// release_hash mismatch
	envelope = map[string]interface{}{
		"schema_version": "2.0",
		"project_id":     "fixture",
		"release_hash":   "0000000000000000000000000000000000000000000000000000000000000000",
		"subject":        subjMap,
	}
	raw, _ = json.Marshal(envelope)
	if err := ValidateMigrationRelease(raw); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Errorf("expected release_hash digest mismatch error, got: %v", err)
	}

	// invalid embedded subject
	badSubj := baseValidReleaseSubjectMap()
	delete(badSubj, "plan_hash")
	envelope = map[string]interface{}{
		"schema_version": "2.0",
		"project_id":     "fixture",
		"release_hash":   expectedHash,
		"subject":        badSubj,
	}
	raw, _ = json.Marshal(envelope)
	if err := ValidateMigrationRelease(raw); err == nil {
		t.Errorf("expected error for invalid embedded subject")
	}
}

func TestValidateMigrationRelease_SyntaxErrors(t *testing.T) {
	if err := ValidateMigrationRelease([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	dup := `{"schema_version":"1.0","schema_version":"1.0"}`
	if err := ValidateMigrationRelease([]byte(dup)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	m := baseValidMigrationReleaseV1Map()
	raw, _ := json.Marshal(m)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateMigrationRelease(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}
