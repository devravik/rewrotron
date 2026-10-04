package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateStepOccurrence_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/observation/step-occurrence.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateStepOccurrence(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid step occurrence %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid step occurrence %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count < 2 {
		t.Fatalf("expected at least 2 step-occurrence validation cases, found %d", count)
	}
}

// TestX65_RepeatedRequestOccurrences validates the X65 invariant:
// Two declared steps sending identical requests produce distinct occurrences with distinct
// visit ordinals (visit=1, visit=2) and distinct physical occurrence IDs.
// Merging them or using array position as identity is forbidden.
func TestX65_RepeatedRequestOccurrences(t *testing.T) {
	makeOccurrence := func(stepID string, visit int, physID string) map[string]any {
		return map[string]any{
			"schema_version":         "1.0",
			"project_id":             "proj-test",
			"side":                   "SOURCE",
			"logical_workflow_id":    "wf-repeat-checkout",
			"step_id":                stepID,
			"visit":                  visit,
			"physical_occurrence_id": physID,
			"execution_witness_ref":  "sha256:" + strings.Repeat("a", 64),
			"pre_context_ref":        "sha256:" + strings.Repeat("b", 64),
			"post_context_ref":       "sha256:" + strings.Repeat("c", 64),
			"guard_results": []map[string]any{
				{
					"guard_id": "auth_ok",
					"outcome":  "TRUE",
					"evidence_refs": []string{
						"sha256:" + strings.Repeat("d", 64),
					},
				},
			},
			"selected_next_step_id": nil,
			"terminal_id":           nil,
			"outcome":               "COMPLETED",
			"evidence_refs": []string{
				"sha256:" + strings.Repeat("e", 64),
			},
		}
	}

	mustMarshal := func(m map[string]any) []byte {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	// Step 1: POST /orders (occurrence #1)
	occ1Raw := mustMarshal(makeOccurrence("create_order", 1, "phys_occ_001"))
	occ1, err := ParseStepOccurrence(occ1Raw)
	if err != nil {
		t.Fatalf("occurrence 1 failed: %v", err)
	}

	// Step 2: POST /orders (occurrence #2, identical request/step, but visit 2 and different physical ID)
	occ2Raw := mustMarshal(makeOccurrence("create_order", 2, "phys_occ_002"))
	occ2, err := ParseStepOccurrence(occ2Raw)
	if err != nil {
		t.Fatalf("occurrence 2 failed: %v", err)
	}

	// 1. Occurrences must have distinct logical keys and physical IDs
	if occ1.LogicalKey() == occ2.LogicalKey() {
		t.Fatalf("expected distinct logical keys, got %s for both", occ1.LogicalKey())
	}
	if occ1.LogicalKey() != "wf-repeat-checkout:create_order:1" {
		t.Fatalf("wanted 'wf-repeat-checkout:create_order:1', got %s", occ1.LogicalKey())
	}
	if occ2.LogicalKey() != "wf-repeat-checkout:create_order:2" {
		t.Fatalf("wanted 'wf-repeat-checkout:create_order:2', got %s", occ2.LogicalKey())
	}
	if occ1.PhysicalOccurrenceID == occ2.PhysicalOccurrenceID {
		t.Fatal("physical occurrence IDs must be distinct")
	}

	// 2. Mutant: non-positive visit (visit=0) must be rejected
	mutantZeroVisit := makeOccurrence("create_order", 0, "phys_occ_bad")
	if err := ValidateStepOccurrence(mustMarshal(mutantZeroVisit)); err == nil {
		t.Fatal("visit=0 must be rejected (visit must be >= 1)")
	}

	// 3. Mutant: negative visit must be rejected
	mutantNegVisit := makeOccurrence("create_order", -1, "phys_occ_bad")
	if err := ValidateStepOccurrence(mustMarshal(mutantNegVisit)); err == nil {
		t.Fatal("negative visit must be rejected")
	}

	// 4. Two distinct execution attempts for same logical step have distinct physical occurrence IDs
	attempt1 := occ1
	attempt2Raw := mustMarshal(makeOccurrence("create_order", 1, "phys_occ_retry_001"))
	attempt2, err := ParseStepOccurrence(attempt2Raw)
	if err != nil {
		t.Fatal(err)
	}
	if attempt1.LogicalKey() != attempt2.LogicalKey() {
		t.Fatal("retried attempt has same logical key")
	}
	if attempt1.PhysicalOccurrenceID == attempt2.PhysicalOccurrenceID {
		t.Fatal("retried attempt must have distinct physical occurrence ID")
	}
}
