package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateOmissionProof_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/analysis/omission-proof.schema.json"
	found := false
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		found = true
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateOmissionProof(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid omission proof fixture %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid omission proof fixture %q to fail validation, but it passed", tc.Name)
			}
		})
	}

	if !found {
		t.Fatalf("missing omission proof fixture in validation cases")
	}
}

func TestValidateOmissionProof_Rejections(t *testing.T) {
	const dummyRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	baseValid := func() map[string]any {
		return map[string]any{
			"schema_version":                  "2.0",
			"project_id":                      "fixture",
			"omitted_ids":                     []string{"unused_mod"},
			"snapshot_ref":                    dummyRef,
			"profile_ref":                     dummyRef,
			"closed_operational_boundary_ref": dummyRef,
			"parse_completeness_ref":          dummyRef,
			"graph_ref":                       dummyRef,
			"analyzer_manifest_ref":           dummyRef,
			"no_reachable_path_ref":           dummyRef,
			"no_unresolved_candidate_ref":     dummyRef,
			"usable_corpus_ref":               dummyRef,
			"zero_runtime_hits_ref":           dummyRef,
			"validity_graph_ref":              dummyRef,
		}
	}

	// 1. Valid baseline passes
	baseBytes, _ := json.Marshal(baseValid())
	if err := ValidateOmissionProof(baseBytes); err != nil {
		t.Fatalf("expected baseline omission proof to pass, got: %v", err)
	}

	// 2. Missing required field
	for k := range baseValid() {
		m := baseValid()
		delete(m, k)
		raw, _ := json.Marshal(m)
		if err := ValidateOmissionProof(raw); err == nil {
			t.Errorf("expected error for missing field %q", k)
		}
	}

	// 3. Null required field
	for k := range baseValid() {
		m := baseValid()
		m[k] = nil
		raw, _ := json.Marshal(m)
		if err := ValidateOmissionProof(raw); err == nil {
			t.Errorf("expected error for null field %q", k)
		}
	}

	// 4. Invalid schema_version
	m := baseValid()
	m["schema_version"] = "1.0"
	raw, _ := json.Marshal(m)
	if err := ValidateOmissionProof(raw); err == nil {
		t.Errorf("expected error for invalid schema_version")
	}

	// 5. Invalid project_id
	m = baseValid()
	m["project_id"] = "PRJ UPPER"
	raw, _ = json.Marshal(m)
	if err := ValidateOmissionProof(raw); err == nil {
		t.Errorf("expected error for invalid project_id")
	}

	// 6. Empty omitted_ids
	m = baseValid()
	m["omitted_ids"] = []string{}
	raw, _ = json.Marshal(m)
	if err := ValidateOmissionProof(raw); err == nil {
		t.Errorf("expected error for empty omitted_ids")
	}

	// 7. Duplicate omitted_ids
	m = baseValid()
	m["omitted_ids"] = []string{"mod_1", "mod_1"}
	raw, _ = json.Marshal(m)
	if err := ValidateOmissionProof(raw); err == nil {
		t.Errorf("expected error for duplicate omitted_ids")
	}

	// 8. Invalid ID format in omitted_ids
	m = baseValid()
	m["omitted_ids"] = []string{"bad id with spaces"}
	raw, _ = json.Marshal(m)
	if err := ValidateOmissionProof(raw); err == nil {
		t.Errorf("expected error for invalid ID in omitted_ids")
	}

	// 9. Invalid ref field
	refFields := []string{
		"snapshot_ref",
		"profile_ref",
		"closed_operational_boundary_ref",
		"parse_completeness_ref",
		"graph_ref",
		"analyzer_manifest_ref",
		"no_reachable_path_ref",
		"no_unresolved_candidate_ref",
		"usable_corpus_ref",
		"zero_runtime_hits_ref",
		"validity_graph_ref",
	}
	for _, rf := range refFields {
		m := baseValid()
		m[rf] = "invalid-ref-format"
		raw, _ := json.Marshal(m)
		if err := ValidateOmissionProof(raw); err == nil {
			t.Errorf("expected error for invalid ref in %q", rf)
		}
	}

	// 10. Unexpected field
	m = baseValid()
	m["extra_field"] = "not_allowed"
	raw, _ = json.Marshal(m)
	if err := ValidateOmissionProof(raw); err == nil {
		t.Errorf("expected error for unexpected field")
	}
}

func TestValidateOmissionProof_PrivacyCanary(t *testing.T) {
	const canary = "CANARY_SECRET_LEAK_OMISSION"

	badInput := `{"schema_version":"2.0","project_id":"` + canary + `","omitted_ids":["valid_id"],"extra":"secret"}`
	err := ValidateOmissionProof([]byte(badInput))
	if err == nil {
		t.Fatalf("expected error for bad input")
	}
	if strings.Contains(err.Error(), canary) {
		t.Fatalf("error message leaks secret canary: %s", err.Error())
	}
}
