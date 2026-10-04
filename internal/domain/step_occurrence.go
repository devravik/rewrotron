package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

// Allowed step occurrence outcomes per specs/observation/step-occurrence.schema.json
const (
	StepOutcomeCompleted             = "COMPLETED"
	StepOutcomeUnknown               = "UNKNOWN"
	StepOutcomeInfrastructureFailure = "INFRASTRUCTURE_FAILURE"
)

// Allowed step guard outcomes
const (
	GuardOutcomeTrue    = "TRUE"
	GuardOutcomeFalse   = "FALSE"
	GuardOutcomeUnknown = "UNKNOWN"
)

var allowedStepOutcomes = map[string]struct{}{
	StepOutcomeCompleted:             {},
	StepOutcomeUnknown:               {},
	StepOutcomeInfrastructureFailure: {},
}

var allowedGuardOutcomes = map[string]struct{}{
	GuardOutcomeTrue:    {},
	GuardOutcomeFalse:   {},
	GuardOutcomeUnknown: {},
}

var (
	stepIDPattern  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	guardIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

// StepOccurrenceError describes an invalid step occurrence record.
type StepOccurrenceError struct {
	Reason string
	Cause  error
}

func (e *StepOccurrenceError) Error() string {
	if e.Cause != nil {
		return "invalid step occurrence: " + e.Reason + ": " + e.Cause.Error()
	}
	return "invalid step occurrence: " + e.Reason
}

func (e *StepOccurrenceError) Unwrap() error { return e.Cause }

// StepOccurrence represents an observed execution occurrence of a workflow step.
// Logical occurrence identity is defined by (logical_workflow_id, step_id, visit) (X65).
// Repeated requests produce distinct occurrences with increasing visit ordinals,
// not merged records or array position identities.
type StepOccurrence struct {
	SchemaVersion        string            `json:"schema_version"`
	ProjectID            string            `json:"project_id"`
	Side                 string            `json:"side"`
	LogicalWorkflowID    string            `json:"logical_workflow_id"`
	StepID               string            `json:"step_id"`
	Visit                int               `json:"visit"`
	PhysicalOccurrenceID string            `json:"physical_occurrence_id"`
	ExecutionWitnessRef  string            `json:"execution_witness_ref"`
	PreContextRef        string            `json:"pre_context_ref"`
	PostContextRef       *string           `json:"post_context_ref"`
	GuardResults         []StepGuardResult `json:"guard_results"`
	SelectedNextStepID   *string           `json:"selected_next_step_id"`
	TerminalID           *string           `json:"terminal_id"`
	Outcome              string            `json:"outcome"`
	EvidenceRefs         []string          `json:"evidence_refs"`
}

// StepGuardResult represents an evaluated guard outcome during step dispatch.
type StepGuardResult struct {
	GuardID      string   `json:"guard_id"`
	Outcome      string   `json:"outcome"`
	EvidenceRefs []string `json:"evidence_refs"`
}

// ValidateStepOccurrence validates raw JSON bytes against specs/observation/step-occurrence.schema.json.
func ValidateStepOccurrence(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return &StepOccurrenceError{Reason: "not a valid JSON object", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"side",
		"logical_workflow_id",
		"step_id",
		"visit",
		"physical_occurrence_id",
		"execution_witness_ref",
		"pre_context_ref",
		"post_context_ref",
		"guard_results",
		"selected_next_step_id",
		"terminal_id",
		"outcome",
		"evidence_refs",
	}

	for _, f := range expectedFields {
		if _, ok := record[f]; !ok {
			return &StepOccurrenceError{Reason: "missing required field " + f}
		}
	}
	if len(record) != len(expectedFields) {
		return &StepOccurrenceError{Reason: "step occurrence contains unrecognized fields"}
	}

	// schema_version
	var schemaVer string
	if err := json.Unmarshal(record["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &StepOccurrenceError{Reason: "schema_version must be '1.0'"}
	}

	// project_id, logical_workflow_id, step_id, physical_occurrence_id
	var projectID, workflowID, stepID, physicalOccID string
	if err := json.Unmarshal(record["project_id"], &projectID); err != nil || !projectIDPattern.MatchString(projectID) {
		return &StepOccurrenceError{Reason: "invalid project_id pattern"}
	}
	if err := json.Unmarshal(record["logical_workflow_id"], &workflowID); err != nil || !workflowIDPattern.MatchString(workflowID) {
		return &StepOccurrenceError{Reason: "invalid logical_workflow_id pattern"}
	}
	if err := json.Unmarshal(record["step_id"], &stepID); err != nil || !stepIDPattern.MatchString(stepID) {
		return &StepOccurrenceError{Reason: "invalid step_id pattern"}
	}
	if err := json.Unmarshal(record["physical_occurrence_id"], &physicalOccID); err != nil || !occurrencePattern.MatchString(physicalOccID) {
		return &StepOccurrenceError{Reason: "invalid physical_occurrence_id pattern"}
	}

	// side: SOURCE or TARGET
	var side string
	if err := json.Unmarshal(record["side"], &side); err != nil || (side != SideSource && side != SideTarget) {
		return &StepOccurrenceError{Reason: "side must be SOURCE or TARGET"}
	}

	// visit: integer minimum 1
	var visit int
	if err := json.Unmarshal(record["visit"], &visit); err != nil || visit < 1 {
		return &StepOccurrenceError{Reason: "visit must be an integer >= 1"}
	}

	// execution_witness_ref, pre_context_ref
	var witnessRef, preContextRef string
	if err := json.Unmarshal(record["execution_witness_ref"], &witnessRef); err != nil || !digestPattern.MatchString(witnessRef) {
		return &StepOccurrenceError{Reason: "execution_witness_ref must be sha256 digest"}
	}
	if err := json.Unmarshal(record["pre_context_ref"], &preContextRef); err != nil || !digestPattern.MatchString(preContextRef) {
		return &StepOccurrenceError{Reason: "pre_context_ref must be sha256 digest"}
	}

	// post_context_ref: string sha256 or null
	if string(bytes.TrimSpace(record["post_context_ref"])) != "null" {
		var postContextRef string
		if err := json.Unmarshal(record["post_context_ref"], &postContextRef); err != nil || !digestPattern.MatchString(postContextRef) {
			return &StepOccurrenceError{Reason: "post_context_ref must be null or sha256 digest"}
		}
	}

	// selected_next_step_id: string pattern or null
	if string(bytes.TrimSpace(record["selected_next_step_id"])) != "null" {
		var nextStepID string
		if err := json.Unmarshal(record["selected_next_step_id"], &nextStepID); err != nil || !stepIDPattern.MatchString(nextStepID) {
			return &StepOccurrenceError{Reason: "selected_next_step_id must be null or match step_id pattern"}
		}
	}

	// terminal_id: string pattern or null
	if string(bytes.TrimSpace(record["terminal_id"])) != "null" {
		var terminalID string
		if err := json.Unmarshal(record["terminal_id"], &terminalID); err != nil || !stepIDPattern.MatchString(terminalID) {
			return &StepOccurrenceError{Reason: "terminal_id must be null or match identifier pattern"}
		}
	}

	// outcome: COMPLETED, UNKNOWN, INFRASTRUCTURE_FAILURE
	var outcome string
	if err := json.Unmarshal(record["outcome"], &outcome); err != nil {
		return &StepOccurrenceError{Reason: "outcome must be string"}
	}
	if _, ok := allowedStepOutcomes[outcome]; !ok {
		return &StepOccurrenceError{Reason: fmt.Sprintf("unsupported step outcome %q", outcome)}
	}

	// evidence_refs: minItems 1
	var evRefs []string
	if err := json.Unmarshal(record["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
		return &StepOccurrenceError{Reason: "evidence_refs must be a non-empty array of sha256 digests"}
	}
	for _, ref := range evRefs {
		if !digestPattern.MatchString(ref) {
			return &StepOccurrenceError{Reason: "evidence_refs item must match sha256 digest pattern"}
		}
	}

	// guard_results
	var guardResults []map[string]json.RawMessage
	if err := json.Unmarshal(record["guard_results"], &guardResults); err != nil {
		return &StepOccurrenceError{Reason: "guard_results must be an array"}
	}
	for _, gr := range guardResults {
		if len(gr) != 3 {
			return &StepOccurrenceError{Reason: "guard_result contains unrecognized fields"}
		}
		rawGID, hasGID := gr["guard_id"]
		rawOut, hasOut := gr["outcome"]
		rawEv, hasEv := gr["evidence_refs"]
		if !hasGID || !hasOut || !hasEv {
			return &StepOccurrenceError{Reason: "guard_result missing required field"}
		}
		var gid, gout string
		if err := json.Unmarshal(rawGID, &gid); err != nil || !guardIDPattern.MatchString(gid) {
			return &StepOccurrenceError{Reason: "guard_id must match identifier pattern"}
		}
		if err := json.Unmarshal(rawOut, &gout); err != nil {
			return &StepOccurrenceError{Reason: "guard outcome must be string"}
		}
		if _, ok := allowedGuardOutcomes[gout]; !ok {
			return &StepOccurrenceError{Reason: fmt.Sprintf("unsupported guard outcome %q", gout)}
		}
		var gEvRefs []string
		if err := json.Unmarshal(rawEv, &gEvRefs); err != nil || len(gEvRefs) < 1 {
			return &StepOccurrenceError{Reason: "guard evidence_refs must be non-empty array of sha256 digests"}
		}
		for _, ref := range gEvRefs {
			if !digestPattern.MatchString(ref) {
				return &StepOccurrenceError{Reason: "guard evidence_refs item must match sha256 digest pattern"}
			}
		}
	}

	return nil
}

// ParseStepOccurrence validates and parses a StepOccurrence from raw JSON.
func ParseStepOccurrence(raw []byte) (*StepOccurrence, error) {
	if err := ValidateStepOccurrence(raw); err != nil {
		return nil, err
	}
	var occ StepOccurrence
	if err := json.Unmarshal(raw, &occ); err != nil {
		return nil, &StepOccurrenceError{Reason: "unmarshal failure", Cause: err}
	}
	return &occ, nil
}

// CanonicalJSON serializes the StepOccurrence to rwt-cjson-1 canonical JSON bytes.
func (s *StepOccurrence) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

// LogicalKey returns the logical occurrence key (workflow:step:visit).
// This uniquely identifies the logical step occurrence in scenario branching (X65).
func (s *StepOccurrence) LogicalKey() string {
	return fmt.Sprintf("%s:%s:%d", s.LogicalWorkflowID, s.StepID, s.Visit)
}
