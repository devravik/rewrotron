package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// VerifySourceApprovalRunnerBinding checks that a SOURCE approval subject shape is valid,
// that the supplied execution manifest record is valid, and that the subject contains exactly one
// role=RUNNER adapter binding with adapter_id=rewrotron-engine binding to the exact execution
// manifest bytes (manifest_ref equals sha256 of manifest), engine version, and engine binary ref
// per CONFIRMATION-PROTOCOL line 47. Both documents must identify the same project_id.
//
// Historical subjects lacking a RUNNER binding (such as baseline hash vectors) fail this check
// because they cannot authorize current execution.
//
// This check does not verify filesystem artifact existence, qualify running engine bytes,
// establish trust, or grant Gate 1 approval. Callers must not mutate supplied buffers.
func VerifySourceApprovalRunnerBinding(subject, executionManifest []byte) error {
	if err := ValidateSourceApprovalSubjectShape(subject); err != nil {
		return err
	}
	if err := ValidateExecutionManifest(executionManifest); err != nil {
		return &SourceApprovalError{Reason: "invalid execution manifest", Cause: err}
	}

	digest := sha256.Sum256(executionManifest)
	expectedManifestRef := "sha256:" + hex.EncodeToString(digest[:])

	var manifestRecord map[string]json.RawMessage
	if err := json.Unmarshal(executionManifest, &manifestRecord); err != nil {
		return &SourceApprovalError{Reason: "manifest projection", Cause: err}
	}
	manifestProjectID := profileString(manifestRecord["project_id"])
	manifestEngineVersion := profileString(manifestRecord["engine_version"])
	manifestEngineBinaryRef := profileString(manifestRecord["engine_binary_ref"])

	var subjectRecord map[string]json.RawMessage
	if err := json.Unmarshal(subject, &subjectRecord); err != nil {
		return &SourceApprovalError{Reason: "subject projection", Cause: err}
	}
	subjectProjectID := profileString(subjectRecord["project_id"])

	if subjectProjectID != manifestProjectID {
		return &SourceApprovalError{Reason: "project ID mismatch"}
	}

	var bindings []json.RawMessage
	if err := json.Unmarshal(subjectRecord["adapter_bindings"], &bindings); err != nil {
		return &SourceApprovalError{Reason: "binding array", Cause: err}
	}

	var runnerBindings []map[string]json.RawMessage
	for _, rawBinding := range bindings {
		var binding map[string]json.RawMessage
		if err := json.Unmarshal(rawBinding, &binding); err != nil {
			return &SourceApprovalError{Reason: "binding projection", Cause: err}
		}
		if profileString(binding["role"]) == "RUNNER" {
			runnerBindings = append(runnerBindings, binding)
		}
	}

	if len(runnerBindings) != 1 {
		return &SourceApprovalError{Reason: "exactly one runner binding required"}
	}

	runner := runnerBindings[0]
	if profileString(runner["adapter_id"]) != "rewrotron-engine" {
		return &SourceApprovalError{Reason: "runner adapter identity"}
	}
	if profileString(runner["manifest_ref"]) != expectedManifestRef {
		return &SourceApprovalError{Reason: "runner manifest reference mismatch"}
	}
	if profileString(runner["version"]) != manifestEngineVersion {
		return &SourceApprovalError{Reason: "runner engine version mismatch"}
	}
	if profileString(runner["implementation_ref"]) != manifestEngineBinaryRef {
		return &SourceApprovalError{Reason: "runner implementation reference mismatch"}
	}

	return nil
}
