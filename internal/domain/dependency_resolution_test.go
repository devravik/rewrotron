package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateDependencyResolution_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/migration/dependency-resolution.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateDependencyResolution(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid dependency-resolution fixture %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid dependency-resolution fixture %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count != 8 {
		t.Fatalf("expected 8 dependency-resolution cases in validation-cases.json, found %d", count)
	}
}

func TestValidateDependencyResolution_Rejections(t *testing.T) {
	const validHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const validRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	baseV1Valid := func() map[string]any {
		return map[string]any{
			"schema_version":   "1.0",
			"dependency_id":    "dep-search",
			"target_spec_hash": validHash,
			"source": map[string]any{
				"ecosystem":     "composer",
				"package":       "legacy/search-client",
				"version":       "1.8.2",
				"evidence_refs": []string{validRef},
			},
			"usage": map[string]any{
				"symbols":               []string{"Search::find"},
				"source_files":          []string{"lib/Search.php"},
				"affected_unit_ids":     []string{"search_unit"},
				"required_capabilities": []string{"cap_search"},
			},
			"proposal": map[string]any{
				"strategy":         "REPLACEMENT",
				"package":          "target/search",
				"version":          "2.0.0",
				"adapter_required": true,
			},
			"compatibility": map[string]any{
				"metadata":             "COMPATIBLE",
				"used_api":             "PARTIAL",
				"behavior":             "MISMATCH",
				"missing_capabilities": []string{"cap_filter"},
			},
			"capability_mappings": []map[string]any{
				{
					"capability_id": "cap_search",
					"provider":      "target/search",
					"status":        "COVERED",
					"case_ids":      []string{"case_001"},
					"evidence_refs": []string{validRef},
				},
			},
			"transitive_dependency_ids": []string{"dep_log"},
			"decision":                  "ACCEPTED",
			"status":                    "PLANNED",
			"evidence_refs":             []string{validRef},
		}
	}

	baseV2Valid := func() map[string]any {
		return map[string]any{
			"schema_version":   "2.0",
			"dependency_id":    "dep-search",
			"target_spec_hash": validHash,
			"source": map[string]any{
				"ecosystem":     "composer",
				"package":       "legacy/search-client",
				"version":       "1.8.2",
				"evidence_refs": []string{validRef},
			},
			"usage": map[string]any{
				"symbols":               []string{"Search::find"},
				"source_files":          []string{"lib/Search.php"},
				"affected_unit_ids":     []string{"search_unit"},
				"required_capabilities": []string{"cap_search"},
			},
			"proposal": map[string]any{
				"strategy":         "REPLACEMENT",
				"package":          "target/search",
				"version":          "2.0.0",
				"adapter_required": true,
			},
			"compatibility": map[string]any{
				"metadata":             "COMPATIBLE",
				"used_api":             "PARTIAL",
				"behavior":             "MISMATCH",
				"missing_capabilities": []string{"cap_filter"},
			},
			"capability_mappings": []map[string]any{
				{
					"capability_id": "cap_search",
					"provider":      "target/search",
					"status":        "COVERED",
					"case_ids":      []string{"case_001"},
					"evidence_refs": []string{validRef},
				},
			},
			"transitive_dependency_ids": []string{"dep_log"},
			"evidence_refs":             []string{validRef},
			"decision_ref":              validRef,
		}
	}

	// 1. Baselines pass
	rawV1, _ := json.Marshal(baseV1Valid())
	if err := ValidateDependencyResolution(rawV1); err != nil {
		t.Fatalf("expected baseline v1 to pass: %v", err)
	}

	rawV2, _ := json.Marshal(baseV2Valid())
	if err := ValidateDependencyResolution(rawV2); err != nil {
		t.Fatalf("expected baseline v2 to pass: %v", err)
	}

	// 2. Unsupported schema version
	m := baseV1Valid()
	m["schema_version"] = "3.0"
	raw, _ := json.Marshal(m)
	if err := ValidateDependencyResolution(raw); err == nil {
		t.Errorf("expected error for unsupported schema version")
	}

	// 3. Missing required field in v1
	for k := range baseV1Valid() {
		m = baseV1Valid()
		delete(m, k)
		raw, _ = json.Marshal(m)
		if err := ValidateDependencyResolution(raw); err == nil {
			t.Errorf("expected error for missing v1 field %q", k)
		}
	}

	// 4. Missing required field in v2
	for k := range baseV2Valid() {
		m = baseV2Valid()
		delete(m, k)
		raw, _ = json.Marshal(m)
		if err := ValidateDependencyResolution(raw); err == nil {
			t.Errorf("expected error for missing v2 field %q", k)
		}
	}

	// 5. Invalid strategy: BLOCKED is a status, not a strategy
	m = baseV1Valid()
	prop := m["proposal"].(map[string]any)
	prop["strategy"] = "BLOCKED"
	raw, _ = json.Marshal(m)
	if err := ValidateDependencyResolution(raw); err == nil {
		t.Errorf("expected error for BLOCKED strategy")
	}

	// 6. Cross-field: PLANNED status with PENDING decision
	m = baseV1Valid()
	m["status"] = "PLANNED"
	m["decision"] = "PENDING"
	raw, _ = json.Marshal(m)
	if err := ValidateDependencyResolution(raw); err == nil {
		t.Errorf("expected error for PLANNED status with PENDING decision")
	}

	// 7. Cross-field: VERIFIED status requires behavior PASS
	m = baseV1Valid()
	m["status"] = "VERIFIED"
	m["decision"] = "ACCEPTED"
	comp := m["compatibility"].(map[string]any)
	comp["behavior"] = "UNKNOWN"
	comp["metadata"] = "COMPATIBLE"
	comp["used_api"] = "COMPATIBLE"
	comp["missing_capabilities"] = []string{}
	raw, _ = json.Marshal(m)
	if err := ValidateDependencyResolution(raw); err == nil {
		t.Errorf("expected error for VERIFIED status with UNKNOWN behavior")
	}

	// 8. Cross-field: VERIFIED status requires characterization cases
	m = baseV1Valid()
	m["status"] = "VERIFIED"
	m["decision"] = "ACCEPTED"
	comp = m["compatibility"].(map[string]any)
	comp["behavior"] = "PASS"
	comp["metadata"] = "COMPATIBLE"
	comp["used_api"] = "COMPATIBLE"
	comp["missing_capabilities"] = []string{}
	maps := m["capability_mappings"].([]map[string]any)
	maps[0]["case_ids"] = []string{}
	raw, _ = json.Marshal(m)
	if err := ValidateDependencyResolution(raw); err == nil {
		t.Errorf("expected error for VERIFIED status with empty case_ids")
	}

	// 9. Unexpected field
	m = baseV1Valid()
	m["extra_field"] = "not_allowed"
	raw, _ = json.Marshal(m)
	if err := ValidateDependencyResolution(raw); err == nil {
		t.Errorf("expected error for unexpected field in v1")
	}
}

func TestValidateDependencyResolution_PrivacyCanary(t *testing.T) {
	const canary = "CANARY_SECRET_LEAK_DEPENDENCY_RESOLUTION"

	badInput := `{"schema_version":"1.0","dependency_id":"` + canary + `","target_spec_hash":"bad"}`
	err := ValidateDependencyResolution([]byte(badInput))
	if err == nil {
		t.Fatalf("expected error for invalid dependency resolution")
	}
	if strings.Contains(err.Error(), canary) {
		t.Fatalf("error message leaks secret canary: %s", err.Error())
	}
}
