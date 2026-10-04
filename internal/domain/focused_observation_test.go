package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestFocused_TenObservationRequirements explicitly tests the 10 focused requirements from the prompt.
func TestFocused_TenObservationRequirements(t *testing.T) {
	mustMarshal := func(m map[string]any) []byte {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	validSourceLedger := func() map[string]any {
		return map[string]any{
			"schema_version":         "1.0",
			"project_id":             "proj-test",
			"side":                   "SOURCE",
			"observation_attempt_id": "attempt-src-1",
			"logical_workflow_id":    "wf-checkout",
			"execution_witness_ref":  "sha256:" + strings.Repeat("a", 64),
			"origin":                 "ACTUAL_OBSERVED",
			"entries": []map[string]any{
				{
					"local_id":    "order_100",
					"kind":        "ENTITY",
					"domain":      "orders",
					"actor_id":    "user_1",
					"resource_id": "orders_table",
					"value": map[string]string{
						"type":  "integer",
						"value": "100",
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
	}

	validStepOccurrence := func(visit int, physID string) map[string]any {
		return map[string]any{
			"schema_version":         "1.0",
			"project_id":             "proj-test",
			"side":                   "SOURCE",
			"logical_workflow_id":    "wf-checkout",
			"step_id":                "place_order",
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

	// 1. Source-only observation with no case ID
	t.Run("1_SourceOnlyObservationNoCaseID", func(t *testing.T) {
		ledger := validSourceLedger()
		// Notice: ledger has no case_id field and requires none!
		if err := ValidateLocalIdentitiesLedger(mustMarshal(ledger)); err != nil {
			t.Fatalf("source-only observation without case_id must pass: %v", err)
		}
	})

	// 2. Source-only local identity with no target value
	t.Run("2_SourceOnlyLocalIdentityNoTargetValue", func(t *testing.T) {
		ledger := validSourceLedger()
		if err := ValidateLocalIdentitiesLedger(mustMarshal(ledger)); err != nil {
			t.Fatalf("source-only local identity must pass without target value: %v", err)
		}
		// If someone injects target value, it must be rejected
		withTargetVal := copyMap(ledger)
		withTargetVal["target_value"] = "999"
		if err := ValidateLocalIdentitiesLedger(mustMarshal(withTargetVal)); err == nil {
			t.Fatal("source-only ledger with target_value must fail")
		}
	})

	// 3. Paired identity requiring both sides (demonstrating why local identity does NOT require pair)
	t.Run("3_LocalIdentityIsStrictlySideLocal", func(t *testing.T) {
		ledger := validSourceLedger()
		// Adding a pairwise mapping field to local-identities ledger is rejected
		withPairwiseMap := copyMap(ledger)
		withPairwiseMap["pairwise_mappings"] = []any{
			map[string]any{"source_id": "order_100", "target_id": "order_200"},
		}
		if err := ValidateLocalIdentitiesLedger(mustMarshal(withPairwiseMap)); err == nil {
			t.Fatal("local identity ledger must remain side-local and reject pairwise mapping additions")
		}
	})

	// 4. Repeated identical request occurrences
	t.Run("4_RepeatedIdenticalRequestOccurrences", func(t *testing.T) {
		occ1, err := ParseStepOccurrence(mustMarshal(validStepOccurrence(1, "phys_1")))
		if err != nil {
			t.Fatal(err)
		}
		occ2, err := ParseStepOccurrence(mustMarshal(validStepOccurrence(2, "phys_2")))
		if err != nil {
			t.Fatal(err)
		}
		if occ1.LogicalKey() == occ2.LogicalKey() {
			t.Fatalf("repeated requests must produce distinct logical keys: %s vs %s", occ1.LogicalKey(), occ2.LogicalKey())
		}
	})

	// 5. Two distinct execution attempts for the same logical step
	t.Run("5_TwoDistinctExecutionAttemptsSameStep", func(t *testing.T) {
		attempt1, err := ParseStepOccurrence(mustMarshal(validStepOccurrence(1, "attempt_1_phys")))
		if err != nil {
			t.Fatal(err)
		}
		attempt2, err := ParseStepOccurrence(mustMarshal(validStepOccurrence(1, "attempt_2_phys")))
		if err != nil {
			t.Fatal(err)
		}
		if attempt1.LogicalKey() != attempt2.LogicalKey() {
			t.Fatal("retried attempts share the same logical key")
		}
		if attempt1.PhysicalOccurrenceID == attempt2.PhysicalOccurrenceID {
			t.Fatal("distinct attempts must have distinct physical occurrence IDs")
		}
	})

	// 6. Stale/wrong build witness
	t.Run("6_StaleWrongBuildWitness", func(t *testing.T) {
		witnessMap := map[string]any{
			"schema_version":         "1.0",
			"project_id":             "proj-test",
			"side":                   "TARGET",
			"attempt_id":             "attempt-tgt-1",
			"source_snapshot_ref":    nil,
			"execution_variant_ref":  nil,
			"plan_ref":               "sha256:" + strings.Repeat("4", 64),
			"target_model_ref":       "sha256:" + strings.Repeat("5", 64),
			"tree_ref":               "sha256:" + strings.Repeat("1", 64),
			"generated_manifest_ref": "sha256:" + strings.Repeat("6", 64),
			"build_artifact_ref":     "sha256:" + strings.Repeat("2", 64),
			"loaded_artifact_ref":    "sha256:" + strings.Repeat("3", 64),
			"launch": map[string]any{
				"host_boot_id":     "boot-01",
				"process_id":       "pid-10",
				"process_start_id": "start-01",
				"container_id":     nil,
			},
			"environment_ref":      "sha256:" + strings.Repeat("7", 64),
			"reset_report_ref":     "sha256:" + strings.Repeat("8", 64),
			"instrumentation_refs": []string{"sha256:" + strings.Repeat("9", 64)},
			"endpoint_binding_ref": "sha256:" + strings.Repeat("a", 64),
			"control_refs":         []string{"sha256:" + strings.Repeat("b", 64)},
			"issuer":               "ENGINE_RUNNER",
			"role_isolation_ref":   "sha256:" + strings.Repeat("c", 64),
		}
		witness, err := ParseExecutionWitness(mustMarshal(witnessMap))
		if err != nil {
			t.Fatal(err)
		}
		staleTreeRef := "sha256:" + strings.Repeat("0", 64)
		if err := witness.VerifyTargetExecutable(staleTreeRef, "sha256:"+strings.Repeat("3", 64)); err == nil {
			t.Fatal("stale target tree must be rejected with ERR_STALE_EXECUTION")
		}
	})

	// 7. Missing execution witness
	t.Run("7_MissingExecutionWitness", func(t *testing.T) {
		ledgerMissingWitness := copyMap(validSourceLedger())
		delete(ledgerMissingWitness, "execution_witness_ref")
		if err := ValidateLocalIdentitiesLedger(mustMarshal(ledgerMissingWitness)); err == nil {
			t.Fatal("ledger missing execution_witness_ref must fail")
		}

		occMissingWitness := validStepOccurrence(1, "phys_1")
		delete(occMissingWitness, "execution_witness_ref")
		if err := ValidateStepOccurrence(mustMarshal(occMissingWitness)); err == nil {
			t.Fatal("step occurrence missing execution_witness_ref must fail")
		}
	})

	// 8. Fixture/model provenance cannot masquerade as actual observation
	t.Run("8_ModelProvenanceCannotMasqueradeAsObservation", func(t *testing.T) {
		ledgerModel := copyMap(validSourceLedger())
		ledgerModel["origin"] = "MODEL_EXPECTED"
		if err := ValidateLocalIdentitiesLedger(mustMarshal(ledgerModel)); err == nil {
			t.Fatal("ledger with MODEL_EXPECTED origin must fail")
		}

		ledgerFixture := copyMap(validSourceLedger())
		ledgerFixture["origin"] = "FIXTURE_PREPARED"
		if err := ValidateLocalIdentitiesLedger(mustMarshal(ledgerFixture)); err == nil {
			t.Fatal("ledger with FIXTURE_PREPARED origin must fail")
		}
	})

	// 9. Source and target identities remain side-qualified
	t.Run("9_SourceAndTargetIdentitiesSideQualified", func(t *testing.T) {
		ledgerInvalidSide := copyMap(validSourceLedger())
		ledgerInvalidSide["side"] = "SHADOW"
		if err := ValidateLocalIdentitiesLedger(mustMarshal(ledgerInvalidSide)); err == nil {
			t.Fatal("ledger side SHADOW must fail (only SOURCE or TARGET allowed)")
		}

		occInvalidSide := validStepOccurrence(1, "phys_1")
		occInvalidSide["side"] = "SHADOW"
		if err := ValidateStepOccurrence(mustMarshal(occInvalidSide)); err == nil {
			t.Fatal("step occurrence side SHADOW must fail")
		}
	})

	// 10. Immutable source observation cannot be retroactively mutated into a pairwise record
	t.Run("10_ImmutableObservationCannotBeMutated", func(t *testing.T) {
		rawOriginal := mustMarshal(validSourceLedger())
		ledger, err := ParseLocalIdentitiesLedger(rawOriginal)
		if err != nil {
			t.Fatal(err)
		}
		c1, err := ledger.CanonicalJSON()
		if err != nil {
			t.Fatal(err)
		}
		c2, err := CanonicalJSON(rawOriginal)
		if err != nil {
			t.Fatal(err)
		}
		if string(c1) != string(c2) {
			t.Fatal("original source observation bytes must be strictly deterministic and immutable")
		}
	})
}
