package domain

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func baseValidPlanSubjectMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":           "2.0",
		"canonicalization":         "rwt-cjson-1",
		"project_id":               "fixture",
		"kind":                     "PLAN",
		"configuration_hash":       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"snapshot_ref":             "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"behavior_ir_ref":          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"corpus_ref":               "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"policy_ref":               "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"target_intent_ref":        "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"resolved_environment_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"target_model_ref":         "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"decision_refs":            []string{},
		"dependency_graph_ref":     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"adapter_bindings": []map[string]interface{}{
			{
				"role":               "TARGET",
				"adapter_id":         "synthetic",
				"version":            "1.0.0",
				"manifest_ref":       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"implementation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		"consumed_revision_refs": []string{},
		"ordered_units": []map[string]interface{}{
			{
				"unit_id":       "login",
				"contract_refs": []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
				"case_ids":      []string{"login"},
				"priority":      1,
				"capabilities":  []string{"http"},
			},
		},
		"dependency_edges":     []map[string]interface{}{},
		"omission_proof_refs":  []string{},
		"scope_exclusion_refs": []string{},
		"unresolved_items":     []map[string]interface{}{},
	}
}

func TestValidatePlanApprovalSubjectShape_Golden(t *testing.T) {
	var vectors []struct {
		Name      string          `json:"name"`
		Subject   json.RawMessage `json:"subject"`
		Canonical string          `json:"canonical_utf8"`
		Hash      string          `json:"sha256"`
	}
	loadVectors(t, "../../specs/examples/approval-hash-vectors.json", &vectors)

	found := false
	for _, v := range vectors {
		if v.Name != "plan" {
			continue
		}
		found = true
		before := bytes.Clone(v.Subject)
		if err := ValidatePlanApprovalSubjectShape(v.Subject); err != nil {
			t.Fatalf("expected golden plan subject to validate, got: %v", err)
		}
		assertGolden(t, v.Subject, v.Canonical, v.Hash)
		if !bytes.Equal(before, v.Subject) {
			t.Fatal("ValidatePlanApprovalSubjectShape mutated input bytes")
		}
	}
	if !found {
		t.Fatal("missing 'plan' vector in approval-hash-vectors.json")
	}
}

func TestValidatePlanApprovalSubjectShape_Rejections(t *testing.T) {
	required := []string{
		"schema_version",
		"canonicalization",
		"project_id",
		"kind",
		"configuration_hash",
		"snapshot_ref",
		"behavior_ir_ref",
		"corpus_ref",
		"policy_ref",
		"target_intent_ref",
		"resolved_environment_ref",
		"target_model_ref",
		"decision_refs",
		"dependency_graph_ref",
		"adapter_bindings",
		"consumed_revision_refs",
		"ordered_units",
		"dependency_edges",
		"omission_proof_refs",
		"scope_exclusion_refs",
		"unresolved_items",
	}

	for _, f := range required {
		m := baseValidPlanSubjectMap()
		delete(m, f)
		raw, _ := json.Marshal(m)
		if err := ValidatePlanApprovalSubjectShape(raw); err == nil {
			t.Errorf("expected error when missing required field %q", f)
		}
	}

	// invalid schema_version
	m := baseValidPlanSubjectMap()
	m["schema_version"] = "1.0"
	raw, _ := json.Marshal(m)
	if err := ValidatePlanApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for invalid schema_version")
	}

	// invalid kind
	m = baseValidPlanSubjectMap()
	m["kind"] = "SOURCE"
	raw, _ = json.Marshal(m)
	if err := ValidatePlanApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for invalid kind")
	}

	// invalid adapter_bindings ordering
	m = baseValidPlanSubjectMap()
	m["adapter_bindings"] = []map[string]interface{}{
		{
			"role":               "TARGET",
			"adapter_id":         "b",
			"version":            "1.0.0",
			"manifest_ref":       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"implementation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			"role":               "TARGET",
			"adapter_id":         "a",
			"version":            "1.0.0",
			"manifest_ref":       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"implementation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	raw, _ = json.Marshal(m)
	if err := ValidatePlanApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for unsorted adapter_bindings")
	}

	// invalid ordered_units (empty)
	m = baseValidPlanSubjectMap()
	m["ordered_units"] = []map[string]interface{}{}
	raw, _ = json.Marshal(m)
	if err := ValidatePlanApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for empty ordered_units")
	}

	// duplicate dependency edge
	m = baseValidPlanSubjectMap()
	m["dependency_edges"] = []map[string]interface{}{
		{"prerequisite_id": "a", "dependent_id": "b"},
		{"prerequisite_id": "a", "dependent_id": "b"},
	}
	raw, _ = json.Marshal(m)
	if err := ValidatePlanApprovalSubjectShape(raw); err == nil {
		t.Errorf("expected error for duplicate dependency edge")
	}
}

func TestValidateMigrationPlan_ValidationCases(t *testing.T) {
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

	count := 0
	for _, tc := range cases {
		if tc.Schema != "https://rewrotron.dev/specs/migration/plan.schema.json" {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateMigrationPlan(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid plan fixture %q, got error: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid plan fixture %q to fail validation, but it passed", tc.Name)
			}
		})
	}

	if count != 7 {
		t.Fatalf("expected 7 validation cases for plan.schema.json, found %d", count)
	}
}

func TestValidateMigrationPlan_V21_CrossFieldRules(t *testing.T) {
	subjMap := baseValidPlanSubjectMap()
	subjMap["decision_refs"] = []string{
		"sha256:0ab26ee1f8b0f264905fd9b11c832060d4d39202e517ffe098f2d6346412ed21",
		"sha256:a17d92d37e6785b1487096bcd4c8eafdb915ed5bda24ceafe3a5b98da0220d8e",
	}
	subjRaw, err := json.Marshal(subjMap)
	if err != nil {
		t.Fatalf("failed to marshal subject: %v", err)
	}
	expectedHash, err := CanonicalHash(subjRaw)
	if err != nil {
		t.Fatalf("failed to compute canonical hash: %v", err)
	}

	validEnvelope := map[string]interface{}{
		"schema_version":     "2.1",
		"plan_id":            "plan-1",
		"project_id":         "fixture",
		"plan_hash":          expectedHash,
		"subject":            subjMap,
		"assessment_ref":     "sha256:0f20635e1861db9ebbf7f466e50511be0c230ec561cf53d021cf581cf55a07e2",
		"selection_ref":      "sha256:0ab26ee1f8b0f264905fd9b11c832060d4d39202e517ffe098f2d6346412ed21",
		"planning_basis_ref": "sha256:a17d92d37e6785b1487096bcd4c8eafdb915ed5bda24ceafe3a5b98da0220d8e",
	}

	raw, err := json.Marshal(validEnvelope)
	if err != nil {
		t.Fatalf("failed to marshal valid envelope: %v", err)
	}
	if err := ValidateMigrationPlan(raw); err != nil {
		t.Fatalf("expected valid v2.1 plan envelope, got: %v", err)
	}

	// project_id mismatch
	badProj := map[string]interface{}{}
	for k, v := range validEnvelope {
		badProj[k] = v
	}
	badProj["project_id"] = "different-proj"
	rawBadProj, _ := json.Marshal(badProj)
	if err := ValidateMigrationPlan(rawBadProj); err == nil || !strings.Contains(err.Error(), "project_id") {
		t.Errorf("expected project_id mismatch error, got: %v", err)
	}

	// selection_ref missing from decision_refs
	badDecision := map[string]interface{}{}
	for k, v := range validEnvelope {
		badDecision[k] = v
	}
	badDecision["selection_ref"] = "sha256:9999999999999999999999999999999999999999999999999999999999999999"
	rawBadDec, _ := json.Marshal(badDecision)
	if err := ValidateMigrationPlan(rawBadDec); err == nil || !strings.Contains(err.Error(), "selection_ref") {
		t.Errorf("expected error when selection_ref not in subject decision_refs, got: %v", err)
	}
}

func TestValidateMigrationPlan_SyntaxErrors(t *testing.T) {
	if err := ValidateMigrationPlan([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	dup := `{"schema_version":"2.0","schema_version":"2.0"}`
	if err := ValidateMigrationPlan([]byte(dup)); err == nil {
		t.Errorf("expected error for duplicate key")
	}
}
