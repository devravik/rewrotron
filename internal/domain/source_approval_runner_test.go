package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func validExecutableSourceClosure(t *testing.T) ([]byte, []byte) {
	t.Helper()
	manifestMap := map[string]any{
		"schema_version":           "1.0",
		"manifest_id":              "engine",
		"project_id":               "fixture",
		"engine_version":           "1.0.0",
		"engine_binary_ref":        "sha256:" + strings.Repeat("b", 64),
		"build_environment_ref":    "sha256:" + strings.Repeat("c", 64),
		"dependency_lock_ref":      "sha256:" + strings.Repeat("d", 64),
		"schema_refs":              []any{"sha256:" + strings.Repeat("e", 64)},
		"canonicalization_profile": "rwt-cjson-1",
		"capability_ids":           []any{"canonical.rwt-cjson-1"},
		"qualification_refs":       []any{"sha256:" + strings.Repeat("f", 64)},
	}
	manifestBytes, err := json.Marshal(manifestMap)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(manifestBytes)
	manifestRef := "sha256:" + hex.EncodeToString(digest[:])

	// Ordered tuples: (role, adapter_id, version, manifest_ref, implementation_ref)
	// "RUNNER" < "SOURCE", so RUNNER binding comes first.
	subjectMap := map[string]any{
		"schema_version":   "2.0",
		"canonicalization": "rwt-cjson-1",
		"project_id":       "fixture",
		"kind":             "SOURCE",
		"snapshot_ref":     "sha256:" + strings.Repeat("1", 64),
		"detection_ref":    "sha256:" + strings.Repeat("2", 64),
		"assertions_ref":   "sha256:" + strings.Repeat("3", 64),
		"adapter_bindings": []any{
			map[string]any{
				"role":               "RUNNER",
				"adapter_id":         "rewrotron-engine",
				"version":            "1.0.0",
				"manifest_ref":       manifestRef,
				"implementation_ref": "sha256:" + strings.Repeat("b", 64),
			},
			map[string]any{
				"role":               "SOURCE",
				"adapter_id":         "synthetic",
				"version":            "1.0.0",
				"manifest_ref":       "sha256:" + strings.Repeat("4", 64),
				"implementation_ref": "sha256:" + strings.Repeat("5", 64),
			},
		},
	}
	subjectBytes, err := json.Marshal(subjectMap)
	if err != nil {
		t.Fatal(err)
	}

	return subjectBytes, manifestBytes
}

func TestHistoricalSourceGoldenFailsRunnerBinding(t *testing.T) {
	t.Parallel()
	var vectors []struct {
		Name    string          `json:"name"`
		Subject json.RawMessage `json:"subject"`
	}
	loadVectors(t, "../../specs/examples/approval-hash-vectors.json", &vectors)

	var historicalSubject []byte
	for _, v := range vectors {
		if v.Name == "source" {
			historicalSubject = v.Subject
			break
		}
	}
	if historicalSubject == nil {
		t.Fatal("missing historical source golden vector")
	}

	_, manifestBytes := validExecutableSourceClosure(t)
	err := VerifySourceApprovalRunnerBinding(historicalSubject, manifestBytes)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var approvalErr *SourceApprovalError
	if !errors.As(err, &approvalErr) || approvalErr.Reason != "exactly one runner binding required" {
		t.Fatalf("wanted 'exactly one runner binding required', got: %v", err)
	}
}

func TestValidSourceApprovalWithRunnerBinding(t *testing.T) {
	t.Parallel()
	subjectBytes, manifestBytes := validExecutableSourceClosure(t)
	if err := VerifySourceApprovalRunnerBinding(subjectBytes, manifestBytes); err != nil {
		t.Fatalf("expected valid runner binding, got: %v", err)
	}
}

func TestRunnerBindingRejections(t *testing.T) {
	t.Parallel()

	t.Run("invalid_subject_shape", func(t *testing.T) {
		t.Parallel()
		_, manifestBytes := validExecutableSourceClosure(t)
		err := VerifySourceApprovalRunnerBinding([]byte(`{"invalid":true}`), manifestBytes)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) {
			t.Fatalf("expected SourceApprovalError, got %T", err)
		}
	})

	t.Run("invalid_execution_manifest", func(t *testing.T) {
		t.Parallel()
		subjectBytes, _ := validExecutableSourceClosure(t)
		err := VerifySourceApprovalRunnerBinding(subjectBytes, []byte(`{"invalid":true}`))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "invalid execution manifest" {
			t.Fatalf("wanted 'invalid execution manifest', got %v", err)
		}
	})

	t.Run("project_id_mismatch", func(t *testing.T) {
		t.Parallel()
		subjectBytes, _ := validExecutableSourceClosure(t)
		manifestMap := map[string]any{
			"schema_version":           "1.0",
			"manifest_id":              "engine",
			"project_id":               "other-project",
			"engine_version":           "1.0.0",
			"engine_binary_ref":        "sha256:" + strings.Repeat("b", 64),
			"build_environment_ref":    "sha256:" + strings.Repeat("c", 64),
			"dependency_lock_ref":      "sha256:" + strings.Repeat("d", 64),
			"schema_refs":              []any{"sha256:" + strings.Repeat("e", 64)},
			"canonicalization_profile": "rwt-cjson-1",
			"capability_ids":           []any{"canonical.rwt-cjson-1"},
			"qualification_refs":       []any{"sha256:" + strings.Repeat("f", 64)},
		}
		manifestBytes, err := json.Marshal(manifestMap)
		if err != nil {
			t.Fatal(err)
		}
		err = VerifySourceApprovalRunnerBinding(subjectBytes, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "project ID mismatch" {
			t.Fatalf("wanted 'project ID mismatch', got %v", err)
		}
	})

	t.Run("runner_adapter_identity_mismatch", func(t *testing.T) {
		t.Parallel()
		manifestMap := map[string]any{
			"schema_version":           "1.0",
			"manifest_id":              "engine",
			"project_id":               "fixture",
			"engine_version":           "1.0.0",
			"engine_binary_ref":        "sha256:" + strings.Repeat("b", 64),
			"build_environment_ref":    "sha256:" + strings.Repeat("c", 64),
			"dependency_lock_ref":      "sha256:" + strings.Repeat("d", 64),
			"schema_refs":              []any{"sha256:" + strings.Repeat("e", 64)},
			"canonicalization_profile": "rwt-cjson-1",
			"capability_ids":           []any{"canonical.rwt-cjson-1"},
			"qualification_refs":       []any{"sha256:" + strings.Repeat("f", 64)},
		}
		manifestBytes, _ := json.Marshal(manifestMap)
		digest := sha256.Sum256(manifestBytes)
		manifestRef := "sha256:" + hex.EncodeToString(digest[:])

		subjectMap := map[string]any{
			"schema_version":   "2.0",
			"canonicalization": "rwt-cjson-1",
			"project_id":       "fixture",
			"kind":             "SOURCE",
			"snapshot_ref":     "sha256:" + strings.Repeat("1", 64),
			"detection_ref":    "sha256:" + strings.Repeat("2", 64),
			"assertions_ref":   "sha256:" + strings.Repeat("3", 64),
			"adapter_bindings": []any{
				map[string]any{
					"role":               "RUNNER",
					"adapter_id":         "foreign-engine",
					"version":            "1.0.0",
					"manifest_ref":       manifestRef,
					"implementation_ref": "sha256:" + strings.Repeat("b", 64),
				},
			},
		}
		subjectBytes, _ := json.Marshal(subjectMap)
		err := VerifySourceApprovalRunnerBinding(subjectBytes, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "runner adapter identity" {
			t.Fatalf("wanted 'runner adapter identity', got %v", err)
		}
	})

	t.Run("runner_manifest_ref_mismatch", func(t *testing.T) {
		t.Parallel()
		subjectBytes, manifestBytes := validExecutableSourceClosure(t)
		var subjectMap map[string]any
		_ = json.Unmarshal(subjectBytes, &subjectMap)
		bindings := subjectMap["adapter_bindings"].([]any)
		runnerBinding := bindings[0].(map[string]any)
		runnerBinding["manifest_ref"] = "sha256:" + strings.Repeat("0", 64)
		tamperedSubject, _ := json.Marshal(subjectMap)

		err := VerifySourceApprovalRunnerBinding(tamperedSubject, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "runner manifest reference mismatch" {
			t.Fatalf("wanted 'runner manifest reference mismatch', got %v", err)
		}
	})

	t.Run("runner_engine_version_mismatch", func(t *testing.T) {
		t.Parallel()
		subjectBytes, manifestBytes := validExecutableSourceClosure(t)
		var subjectMap map[string]any
		_ = json.Unmarshal(subjectBytes, &subjectMap)
		bindings := subjectMap["adapter_bindings"].([]any)
		runnerBinding := bindings[0].(map[string]any)
		runnerBinding["version"] = "2.0.0"
		tamperedSubject, _ := json.Marshal(subjectMap)

		err := VerifySourceApprovalRunnerBinding(tamperedSubject, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "runner engine version mismatch" {
			t.Fatalf("wanted 'runner engine version mismatch', got %v", err)
		}
	})

	t.Run("runner_implementation_ref_mismatch", func(t *testing.T) {
		t.Parallel()
		subjectBytes, manifestBytes := validExecutableSourceClosure(t)
		var subjectMap map[string]any
		_ = json.Unmarshal(subjectBytes, &subjectMap)
		bindings := subjectMap["adapter_bindings"].([]any)
		runnerBinding := bindings[0].(map[string]any)
		runnerBinding["implementation_ref"] = "sha256:" + strings.Repeat("9", 64)
		tamperedSubject, _ := json.Marshal(subjectMap)

		err := VerifySourceApprovalRunnerBinding(tamperedSubject, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "runner implementation reference mismatch" {
			t.Fatalf("wanted 'runner implementation reference mismatch', got %v", err)
		}
	})

	t.Run("multiple_runner_bindings", func(t *testing.T) {
		t.Parallel()
		subjectBytes, manifestBytes := validExecutableSourceClosure(t)
		var subjectMap map[string]any
		_ = json.Unmarshal(subjectBytes, &subjectMap)
		bindings := subjectMap["adapter_bindings"].([]any)
		runnerBinding := bindings[0].(map[string]any)
		// Add another runner binding with different version to keep order valid:
		secondRunner := map[string]any{
			"role":               "RUNNER",
			"adapter_id":         "rewrotron-engine",
			"version":            "1.0.1",
			"manifest_ref":       runnerBinding["manifest_ref"],
			"implementation_ref": runnerBinding["implementation_ref"],
		}
		subjectMap["adapter_bindings"] = []any{runnerBinding, secondRunner}
		tamperedSubject, _ := json.Marshal(subjectMap)

		err := VerifySourceApprovalRunnerBinding(tamperedSubject, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "exactly one runner binding required" {
			t.Fatalf("wanted 'exactly one runner binding required', got %v", err)
		}
	})

	t.Run("inverted_binding_order", func(t *testing.T) {
		t.Parallel()
		subjectBytes, manifestBytes := validExecutableSourceClosure(t)
		var subjectMap map[string]any
		_ = json.Unmarshal(subjectBytes, &subjectMap)
		bindings := subjectMap["adapter_bindings"].([]any)
		// Reverse bindings: SOURCE first, then RUNNER (invalid order!)
		subjectMap["adapter_bindings"] = []any{bindings[1], bindings[0]}
		invertedSubject, _ := json.Marshal(subjectMap)

		err := VerifySourceApprovalRunnerBinding(invertedSubject, manifestBytes)
		var approvalErr *SourceApprovalError
		if !errors.As(err, &approvalErr) || approvalErr.Reason != "binding order or duplicate" {
			t.Fatalf("wanted 'binding order or duplicate', got %v", err)
		}
	})
}
