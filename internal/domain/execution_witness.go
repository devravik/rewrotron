package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

// Allowed witness issuer
const (
	WitnessIssuerEngineRunner = "ENGINE_RUNNER"
)

var (
	bootIDPattern    = regexp.MustCompile(`^[a-z0-9_-]+$`)
	processIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
	startIDPattern   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	containerPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

// ExecutionWitnessError describes an invalid execution witness record.
type ExecutionWitnessError struct {
	Reason string
	Cause  error
}

func (e *ExecutionWitnessError) Error() string {
	if e.Cause != nil {
		return "invalid execution witness: " + e.Reason + ": " + e.Cause.Error()
	}
	return "invalid execution witness: " + e.Reason
}

func (e *ExecutionWitnessError) Unwrap() error { return e.Cause }

// ExecutionWitness represents a trusted execution witness conforming to specs/observation/execution-witness.schema.json.
// It establishes the chain of provenance from snapshot/plan to launched process/container and loaded executable.
// A source witness cannot serve as target (and vice versa). Stale/wrong build or endpoint witness yields
// ERR_STALE_EXECUTION (X66).
type ExecutionWitness struct {
	SchemaVersion        string          `json:"schema_version"`
	ProjectID            string          `json:"project_id"`
	Side                 string          `json:"side"`
	AttemptID            string          `json:"attempt_id"`
	SourceSnapshotRef    *string         `json:"source_snapshot_ref"`
	ExecutionVariantRef  *string         `json:"execution_variant_ref"`
	PlanRef              *string         `json:"plan_ref"`
	TargetModelRef       *string         `json:"target_model_ref"`
	TreeRef              string          `json:"tree_ref"`
	GeneratedManifestRef *string         `json:"generated_manifest_ref"`
	BuildArtifactRef     string          `json:"build_artifact_ref"`
	LoadedArtifactRef    string          `json:"loaded_artifact_ref"`
	Launch               ExecutionLaunch `json:"launch"`
	EnvironmentRef       string          `json:"environment_ref"`
	ResetReportRef       string          `json:"reset_report_ref"`
	InstrumentationRefs  []string        `json:"instrumentation_refs"`
	EndpointBindingRef   string          `json:"endpoint_binding_ref"`
	ControlRefs          []string        `json:"control_refs"`
	Issuer               string          `json:"issuer"`
	RoleIsolationRef     string          `json:"role_isolation_ref"`
}

// ExecutionLaunch records the host boot, process, start ID, and optional container identity.
type ExecutionLaunch struct {
	HostBootID     string  `json:"host_boot_id"`
	ProcessID      string  `json:"process_id"`
	ProcessStartID string  `json:"process_start_id"`
	ContainerID    *string `json:"container_id"`
}

// ValidateExecutionWitness validates raw JSON bytes against specs/observation/execution-witness.schema.json.
func ValidateExecutionWitness(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return &ExecutionWitnessError{Reason: "not a valid JSON object", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"side",
		"attempt_id",
		"source_snapshot_ref",
		"execution_variant_ref",
		"plan_ref",
		"target_model_ref",
		"tree_ref",
		"generated_manifest_ref",
		"build_artifact_ref",
		"loaded_artifact_ref",
		"launch",
		"environment_ref",
		"reset_report_ref",
		"instrumentation_refs",
		"endpoint_binding_ref",
		"control_refs",
		"issuer",
		"role_isolation_ref",
	}

	for _, f := range expectedFields {
		if _, ok := record[f]; !ok {
			return &ExecutionWitnessError{Reason: "missing required field " + f}
		}
	}
	if len(record) != len(expectedFields) {
		return &ExecutionWitnessError{Reason: "execution witness contains unrecognized fields"}
	}

	// schema_version
	var schemaVer string
	if err := json.Unmarshal(record["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &ExecutionWitnessError{Reason: "schema_version must be '1.0'"}
	}

	// project_id, attempt_id
	var projectID, attemptID string
	if err := json.Unmarshal(record["project_id"], &projectID); err != nil || !projectIDPattern.MatchString(projectID) {
		return &ExecutionWitnessError{Reason: "invalid project_id pattern"}
	}
	if err := json.Unmarshal(record["attempt_id"], &attemptID); err != nil || !occurrencePattern.MatchString(attemptID) {
		return &ExecutionWitnessError{Reason: "invalid attempt_id pattern"}
	}

	// side: SOURCE or TARGET
	var side string
	if err := json.Unmarshal(record["side"], &side); err != nil || (side != SideSource && side != SideTarget) {
		return &ExecutionWitnessError{Reason: "side must be SOURCE or TARGET"}
	}

	// issuer: ENGINE_RUNNER
	var issuer string
	if err := json.Unmarshal(record["issuer"], &issuer); err != nil || issuer != WitnessIssuerEngineRunner {
		return &ExecutionWitnessError{Reason: "issuer must be ENGINE_RUNNER"}
	}

	// tree_ref, build_artifact_ref, loaded_artifact_ref, environment_ref, reset_report_ref, endpoint_binding_ref, role_isolation_ref
	singleDigests := map[string]string{
		"tree_ref":             "tree_ref",
		"build_artifact_ref":   "build_artifact_ref",
		"loaded_artifact_ref":  "loaded_artifact_ref",
		"environment_ref":      "environment_ref",
		"reset_report_ref":     "reset_report_ref",
		"endpoint_binding_ref": "endpoint_binding_ref",
		"role_isolation_ref":   "role_isolation_ref",
	}
	for fieldName := range singleDigests {
		var d string
		if err := json.Unmarshal(record[fieldName], &d); err != nil || !digestPattern.MatchString(d) {
			return &ExecutionWitnessError{Reason: fieldName + " must be sha256 digest"}
		}
	}

	// instrumentation_refs, control_refs: minItems 1
	for _, arrField := range []string{"instrumentation_refs", "control_refs"} {
		var refs []string
		if err := json.Unmarshal(record[arrField], &refs); err != nil || len(refs) < 1 {
			return &ExecutionWitnessError{Reason: arrField + " must be a non-empty array of sha256 digests"}
		}
		for _, ref := range refs {
			if !digestPattern.MatchString(ref) {
				return &ExecutionWitnessError{Reason: arrField + " item must match sha256 digest pattern"}
			}
		}
	}

	// launch object
	var launch map[string]json.RawMessage
	if err := json.Unmarshal(record["launch"], &launch); err != nil {
		return &ExecutionWitnessError{Reason: "launch must be an object", Cause: err}
	}
	launchReq := []string{"host_boot_id", "process_id", "process_start_id", "container_id"}
	for _, f := range launchReq {
		if _, ok := launch[f]; !ok {
			return &ExecutionWitnessError{Reason: "launch missing required field " + f}
		}
	}
	if len(launch) != len(launchReq) {
		return &ExecutionWitnessError{Reason: "launch contains unrecognized fields"}
	}
	var bootID, procID, startID string
	if err := json.Unmarshal(launch["host_boot_id"], &bootID); err != nil || !bootIDPattern.MatchString(bootID) {
		return &ExecutionWitnessError{Reason: "launch.host_boot_id must match identifier pattern"}
	}
	if err := json.Unmarshal(launch["process_id"], &procID); err != nil || !processIDPattern.MatchString(procID) {
		return &ExecutionWitnessError{Reason: "launch.process_id must match identifier pattern"}
	}
	if err := json.Unmarshal(launch["process_start_id"], &startID); err != nil || !startIDPattern.MatchString(startID) {
		return &ExecutionWitnessError{Reason: "launch.process_start_id must match identifier pattern"}
	}
	if string(bytes.TrimSpace(launch["container_id"])) != "null" {
		var containerID string
		if err := json.Unmarshal(launch["container_id"], &containerID); err != nil || !containerPattern.MatchString(containerID) {
			return &ExecutionWitnessError{Reason: "launch.container_id must be null or match identifier pattern"}
		}
	}

	// allOf conditional rules for SOURCE vs TARGET
	if side == SideSource {
		// source_snapshot_ref must be string sha256
		if string(bytes.TrimSpace(record["source_snapshot_ref"])) == "null" {
			return &ExecutionWitnessError{Reason: "SOURCE witness requires non-null source_snapshot_ref"}
		}
		var snapRef string
		if err := json.Unmarshal(record["source_snapshot_ref"], &snapRef); err != nil || !digestPattern.MatchString(snapRef) {
			return &ExecutionWitnessError{Reason: "source_snapshot_ref must be sha256 digest"}
		}
		// plan_ref and target_model_ref must be null
		if string(bytes.TrimSpace(record["plan_ref"])) != "null" {
			return &ExecutionWitnessError{Reason: "SOURCE witness requires plan_ref to be null"}
		}
		if string(bytes.TrimSpace(record["target_model_ref"])) != "null" {
			return &ExecutionWitnessError{Reason: "SOURCE witness requires target_model_ref to be null"}
		}
		// execution_variant_ref can be sha256 or null
		if string(bytes.TrimSpace(record["execution_variant_ref"])) != "null" {
			var evRef string
			if err := json.Unmarshal(record["execution_variant_ref"], &evRef); err != nil || !digestPattern.MatchString(evRef) {
				return &ExecutionWitnessError{Reason: "execution_variant_ref must be null or sha256 digest"}
			}
		}
	} else {
		// TARGET witness
		// source_snapshot_ref and execution_variant_ref must be null
		if string(bytes.TrimSpace(record["source_snapshot_ref"])) != "null" {
			return &ExecutionWitnessError{Reason: "TARGET witness requires source_snapshot_ref to be null"}
		}
		if string(bytes.TrimSpace(record["execution_variant_ref"])) != "null" {
			return &ExecutionWitnessError{Reason: "TARGET witness requires execution_variant_ref to be null"}
		}
		// plan_ref, target_model_ref, generated_manifest_ref must be string sha256
		for _, reqDigest := range []string{"plan_ref", "target_model_ref", "generated_manifest_ref"} {
			if string(bytes.TrimSpace(record[reqDigest])) == "null" {
				return &ExecutionWitnessError{Reason: fmt.Sprintf("TARGET witness requires non-null %s", reqDigest)}
			}
			var d string
			if err := json.Unmarshal(record[reqDigest], &d); err != nil || !digestPattern.MatchString(d) {
				return &ExecutionWitnessError{Reason: reqDigest + " must be sha256 digest"}
			}
		}
	}

	return nil
}

// ParseExecutionWitness validates and parses an ExecutionWitness from raw JSON.
func ParseExecutionWitness(raw []byte) (*ExecutionWitness, error) {
	if err := ValidateExecutionWitness(raw); err != nil {
		return nil, err
	}
	var witness ExecutionWitness
	if err := json.Unmarshal(raw, &witness); err != nil {
		return nil, &ExecutionWitnessError{Reason: "unmarshal failure", Cause: err}
	}
	return &witness, nil
}

// CanonicalJSON serializes the ExecutionWitness to rwt-cjson-1 canonical JSON bytes.
func (w *ExecutionWitness) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

// VerifyTargetExecutable checks that the execution witness corresponds to the expected target tree and loaded artifact.
// Under X66, serving an old/stale build or opcode cache while claiming current target tree is rejected with ERR_STALE_EXECUTION.
func (w *ExecutionWitness) VerifyTargetExecutable(expectedTreeRef, expectedLoadedArtifactRef string) error {
	if w.Side != SideTarget {
		return &ExecutionWitnessError{Reason: "cannot verify target executable on SOURCE witness"}
	}
	if w.TreeRef != expectedTreeRef {
		return &ExecutionWitnessError{Reason: "stale executed build: tree_ref does not match expected target tree (ERR_STALE_EXECUTION)"}
	}
	if expectedLoadedArtifactRef != "" && w.LoadedArtifactRef != expectedLoadedArtifactRef {
		return &ExecutionWitnessError{Reason: "stale executed build: loaded_artifact_ref does not match expected executable (ERR_STALE_EXECUTION)"}
	}
	return nil
}
