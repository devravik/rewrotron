package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateLocalIdentitiesLedger_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/observation/local-identities.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateLocalIdentitiesLedger(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid ledger %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid ledger %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count < 5 {
		t.Fatalf("expected at least 5 local-identities validation cases, found %d", count)
	}
}

// TestX62_SourceOnlyIdentity validates the core X62 invariant:
// Observation precedes correspondence. A source-only ledger is valid and publishable
// with no target, no target placeholder, and no verification case.
func TestX62_SourceOnlyIdentity(t *testing.T) {
	validSourceLedger := map[string]any{
		"schema_version":         "1.0",
		"project_id":             "proj-test",
		"side":                   "SOURCE",
		"observation_attempt_id": "attempt-src-1",
		"logical_workflow_id":    "wf-login-create",
		"execution_witness_ref":  "sha256:" + strings.Repeat("a", 64),
		"origin":                 "ACTUAL_OBSERVED",
		"entries": []map[string]any{
			{
				"local_id":    "order_rec_1",
				"kind":        "ENTITY",
				"domain":      "orders",
				"actor_id":    "customer_1",
				"resource_id": "orders_table",
				"value": map[string]string{
					"type":  "integer",
					"value": "42",
				},
				"relationship_refs":   []string{},
				"capture_channel_ref": "sha256:" + strings.Repeat("b", 64),
				"evidence_refs": []string{
					"sha256:" + strings.Repeat("c", 64),
				},
			},
		},
		"capture_evidence_refs": []string{
			"sha256:" + strings.Repeat("d", 64),
		},
	}

	mustMarshal := func(m map[string]any) []byte {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	// 1. Source-only ledger passes without any target or verification case
	rawValid := mustMarshal(validSourceLedger)
	ledger, err := ParseLocalIdentitiesLedger(rawValid)
	if err != nil {
		t.Fatalf("valid source-only ledger failed: %v", err)
	}
	if ledger.Side != SideSource {
		t.Fatalf("wanted side SOURCE, got %s", ledger.Side)
	}
	if entry, ok := ledger.FindEntry("order_rec_1"); !ok || entry.Value.Value != "42" {
		t.Fatalf("failed to find entry in source ledger: %+v", entry)
	}

	// 2. Mutant: demanding target placeholder (e.g. target_value, target_id) must FAIL
	mutantTargetPlaceholder := copyMap(validSourceLedger)
	mutantTargetPlaceholder["target_value"] = 42
	if err := ValidateLocalIdentitiesLedger(mustMarshal(mutantTargetPlaceholder)); err == nil {
		t.Fatal("ledger containing target placeholder must be rejected")
	}

	// 3. Mutant: invented missing value (NO_OBSERVATION) must FAIL
	mutantMissingVal := copyMap(validSourceLedger)
	mutantMissingVal["entries"] = []map[string]any{
		{
			"local_id":    "order_rec_1",
			"kind":        "ENTITY",
			"domain":      "orders",
			"actor_id":    "customer_1",
			"resource_id": "orders_table",
			"value": map[string]string{
				"type":        "missing",
				"reason_code": "NO_OBSERVATION",
			},
			"relationship_refs":   []string{},
			"capture_channel_ref": "sha256:" + strings.Repeat("b", 64),
			"evidence_refs": []string{
				"sha256:" + strings.Repeat("c", 64),
			},
		},
	}
	if err := ValidateLocalIdentitiesLedger(mustMarshal(mutantMissingVal)); err == nil {
		t.Fatal("ledger entry with missing value must be rejected (no invented local identity)")
	}

	// 4. Mutant: MODEL_EXPECTED masquerading as ACTUAL_OBSERVED must FAIL
	mutantModelExpected := copyMap(validSourceLedger)
	mutantModelExpected["origin"] = "MODEL_EXPECTED"
	if err := ValidateLocalIdentitiesLedger(mustMarshal(mutantModelExpected)); err == nil {
		t.Fatal("ledger with MODEL_EXPECTED origin must be rejected")
	}

	// 5. Immutability: canonical serialization reproduces identical bytes
	c1, err := ledger.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := ledger.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(c1) != string(c2) {
		t.Fatal("canonical JSON must be strictly deterministic")
	}
}
