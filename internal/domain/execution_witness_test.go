package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateExecutionWitness_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/observation/execution-witness.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateExecutionWitness(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid witness %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid witness %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count < 4 {
		t.Fatalf("expected at least 4 execution-witness validation cases, found %d", count)
	}
}

// TestX66_StaleExecutedBuild validates the X66 invariant:
// An observation must carry an execution witness proving the actual launched/loaded bytes
// correspond to the current approved target tree. Serving an old build or opcode cache
// blocks contribution to PASS with ERR_STALE_EXECUTION.
func TestX66_StaleExecutedBuild(t *testing.T) {
	targetTreeRef := "sha256:" + strings.Repeat("1", 64)
	buildArtifactRef := "sha256:" + strings.Repeat("2", 64)
	loadedArtifactRef := "sha256:" + strings.Repeat("3", 64)

	validTargetWitness := map[string]any{
		"schema_version":         "1.0",
		"project_id":             "proj-test",
		"side":                   "TARGET",
		"attempt_id":             "attempt-tgt-1",
		"source_snapshot_ref":    nil,
		"execution_variant_ref":  nil,
		"plan_ref":               "sha256:" + strings.Repeat("4", 64),
		"target_model_ref":       "sha256:" + strings.Repeat("5", 64),
		"tree_ref":               targetTreeRef,
		"generated_manifest_ref": "sha256:" + strings.Repeat("6", 64),
		"build_artifact_ref":     buildArtifactRef,
		"loaded_artifact_ref":    loadedArtifactRef,
		"launch": map[string]any{
			"host_boot_id":     "boot-uuid-01",
			"process_id":       "pid-1234",
			"process_start_id": "start-monotonic-01",
			"container_id":     "container-sha-01",
		},
		"environment_ref":      "sha256:" + strings.Repeat("7", 64),
		"reset_report_ref":     "sha256:" + strings.Repeat("8", 64),
		"instrumentation_refs": []string{"sha256:" + strings.Repeat("9", 64)},
		"endpoint_binding_ref": "sha256:" + strings.Repeat("a", 64),
		"control_refs":         []string{"sha256:" + strings.Repeat("b", 64)},
		"issuer":               "ENGINE_RUNNER",
		"role_isolation_ref":   "sha256:" + strings.Repeat("c", 64),
	}

	mustMarshal := func(m map[string]any) []byte {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	// 1. Control passes
	rawWitness := mustMarshal(validTargetWitness)
	witness, err := ParseExecutionWitness(rawWitness)
	if err != nil {
		t.Fatalf("valid target witness failed: %v", err)
	}

	// 2. VerifyTargetExecutable succeeds when tree_ref and loaded_artifact_ref match
	if err := witness.VerifyTargetExecutable(targetTreeRef, loadedArtifactRef); err != nil {
		t.Fatalf("expected executable verification to pass, got: %v", err)
	}

	// 3. X66 Stale Build Mutant: tree_ref does not match current target tree
	mutantWrongTree := "sha256:" + strings.Repeat("f", 64)
	err = witness.VerifyTargetExecutable(mutantWrongTree, loadedArtifactRef)
	if err == nil || !strings.Contains(err.Error(), "ERR_STALE_EXECUTION") {
		t.Fatalf("expected ERR_STALE_EXECUTION when tree_ref mismatches, got: %v", err)
	}

	// 4. X66 Stale Cache Mutant: loaded_artifact_ref does not match current built executable
	mutantWrongLoaded := "sha256:" + strings.Repeat("e", 64)
	err = witness.VerifyTargetExecutable(targetTreeRef, mutantWrongLoaded)
	if err == nil || !strings.Contains(err.Error(), "ERR_STALE_EXECUTION") {
		t.Fatalf("expected ERR_STALE_EXECUTION when loaded_artifact_ref mismatches, got: %v", err)
	}

	// 5. Source witness cannot masquerade as target witness (Review-2 invariant)
	sourceWitnessMap := copyMap(validTargetWitness)
	sourceWitnessMap["side"] = "SOURCE"
	sourceWitnessMap["source_snapshot_ref"] = "sha256:" + strings.Repeat("d", 64)
	sourceWitnessMap["plan_ref"] = nil
	sourceWitnessMap["target_model_ref"] = nil
	sourceWitnessMap["generated_manifest_ref"] = nil
	sourceWitness, err := ParseExecutionWitness(mustMarshal(sourceWitnessMap))
	if err != nil {
		t.Fatalf("valid source witness failed: %v", err)
	}
	if err := sourceWitness.VerifyTargetExecutable(targetTreeRef, loadedArtifactRef); err == nil {
		t.Fatal("source witness must not verify as target executable")
	}

	// 6. Source witness with plan_ref is rejected
	sourceWithPlan := copyMap(sourceWitnessMap)
	sourceWithPlan["plan_ref"] = "sha256:" + strings.Repeat("4", 64)
	if err := ValidateExecutionWitness(mustMarshal(sourceWithPlan)); err == nil {
		t.Fatal("SOURCE witness with non-null plan_ref must be rejected")
	}

	// 7. Target witness without plan_ref is rejected
	targetWithoutPlan := copyMap(validTargetWitness)
	targetWithoutPlan["plan_ref"] = nil
	if err := ValidateExecutionWitness(mustMarshal(targetWithoutPlan)); err == nil {
		t.Fatal("TARGET witness with null plan_ref must be rejected")
	}

	// 8. Issuer must be ENGINE_RUNNER
	wrongIssuer := copyMap(validTargetWitness)
	wrongIssuer["issuer"] = "AGENT_PROPOSAL"
	if err := ValidateExecutionWitness(mustMarshal(wrongIssuer)); err == nil {
		t.Fatal("witness with issuer != ENGINE_RUNNER must be rejected")
	}
}
