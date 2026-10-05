package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// PlanEnvelopeError describes an invalid plan envelope or subject without echoing private contents.
type PlanEnvelopeError struct {
	Reason string
	Cause  error
}

// Error identifies the failed plan envelope obligation.
func (e *PlanEnvelopeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid plan envelope: %s: %v", e.Reason, e.Cause)
	}
	return "invalid plan envelope: " + e.Reason
}

// Unwrap preserves the underlying error cause.
func (e *PlanEnvelopeError) Unwrap() error { return e.Cause }

// PlanApprovalUnit describes a schedulable reconstruction unit in the Gate 2 plan subject.
type PlanApprovalUnit struct {
	UnitID       string   `json:"unit_id"`
	ContractRefs []string `json:"contract_refs"`
	CaseIDs      []string `json:"case_ids"`
	Priority     int      `json:"priority"`
	Capabilities []string `json:"capabilities"`
}

// PlanApprovalEdge describes a dependency ordering constraint between two plan units.
type PlanApprovalEdge struct {
	PrerequisiteID string `json:"prerequisite_id"`
	DependentID    string `json:"dependent_id"`
}

// PlanApprovalUnresolvedItem describes an explicit unresolved source or migration issue in the plan subject.
type PlanApprovalUnresolvedItem struct {
	Code         string   `json:"code"`
	AffectedIDs  []string `json:"affected_ids"`
	EvidenceRefs []string `json:"evidence_refs"`
	Blocking     bool     `json:"blocking"`
}

// PlanAdapterBinding describes a bound adapter or engine in the plan approval subject.
type PlanAdapterBinding struct {
	Role              string `json:"role"`
	AdapterID         string `json:"adapter_id"`
	Version           string `json:"version"`
	ManifestRef       string `json:"manifest_ref"`
	ImplementationRef string `json:"implementation_ref"`
}

// PlanApprovalSubject represents the immutable Gate 2 PLAN approval subject (schema 2.0).
type PlanApprovalSubject struct {
	SchemaVersion          string                       `json:"schema_version"`
	Canonicalization       string                       `json:"canonicalization"`
	ProjectID              string                       `json:"project_id"`
	Kind                   string                       `json:"kind"`
	ConfigurationHash      string                       `json:"configuration_hash"`
	SnapshotRef            string                       `json:"snapshot_ref"`
	BehaviorIRRef          string                       `json:"behavior_ir_ref"`
	CorpusRef              string                       `json:"corpus_ref"`
	PolicyRef              string                       `json:"policy_ref"`
	TargetIntentRef        string                       `json:"target_intent_ref"`
	ResolvedEnvironmentRef string                       `json:"resolved_environment_ref"`
	TargetModelRef         string                       `json:"target_model_ref"`
	DecisionRefs           []string                     `json:"decision_refs"`
	DependencyGraphRef     string                       `json:"dependency_graph_ref"`
	AdapterBindings        []PlanAdapterBinding         `json:"adapter_bindings"`
	ConsumedRevisionRefs   []string                     `json:"consumed_revision_refs"`
	OrderedUnits           []PlanApprovalUnit           `json:"ordered_units"`
	DependencyEdges        []PlanApprovalEdge           `json:"dependency_edges"`
	OmissionProofRefs      []string                     `json:"omission_proof_refs"`
	ScopeExclusionRefs     []string                     `json:"scope_exclusion_refs"`
	UnresolvedItems        []PlanApprovalUnresolvedItem `json:"unresolved_items"`
}

// CanonicalJSON renders the plan approval subject into canonical RFC 8785 bytes.
func (s *PlanApprovalSubject) CanonicalJSON() ([]byte, error) {
	if s == nil {
		return nil, &PlanEnvelopeError{Reason: "plan approval subject is nil"}
	}
	decRefs := s.DecisionRefs
	if decRefs == nil {
		decRefs = []string{}
	}
	bindings := s.AdapterBindings
	if bindings == nil {
		bindings = []PlanAdapterBinding{}
	}
	consumed := s.ConsumedRevisionRefs
	if consumed == nil {
		consumed = []string{}
	}
	units := s.OrderedUnits
	if units == nil {
		units = []PlanApprovalUnit{}
	}
	edges := s.DependencyEdges
	if edges == nil {
		edges = []PlanApprovalEdge{}
	}
	omissions := s.OmissionProofRefs
	if omissions == nil {
		omissions = []string{}
	}
	scopeExclusions := s.ScopeExclusionRefs
	if scopeExclusions == nil {
		scopeExclusions = []string{}
	}
	unresolved := s.UnresolvedItems
	if unresolved == nil {
		unresolved = []PlanApprovalUnresolvedItem{}
	}

	recordMap := map[string]any{
		"schema_version":           s.SchemaVersion,
		"canonicalization":         s.Canonicalization,
		"project_id":               s.ProjectID,
		"kind":                     s.Kind,
		"configuration_hash":       s.ConfigurationHash,
		"snapshot_ref":             s.SnapshotRef,
		"behavior_ir_ref":          s.BehaviorIRRef,
		"corpus_ref":               s.CorpusRef,
		"policy_ref":               s.PolicyRef,
		"target_intent_ref":        s.TargetIntentRef,
		"resolved_environment_ref": s.ResolvedEnvironmentRef,
		"target_model_ref":         s.TargetModelRef,
		"decision_refs":            decRefs,
		"dependency_graph_ref":     s.DependencyGraphRef,
		"adapter_bindings":         bindings,
		"consumed_revision_refs":   consumed,
		"ordered_units":            units,
		"dependency_edges":         edges,
		"omission_proof_refs":      omissions,
		"scope_exclusion_refs":     scopeExclusions,
		"unresolved_items":         unresolved,
	}

	raw, err := json.Marshal(recordMap)
	if err != nil {
		return nil, &PlanEnvelopeError{Reason: "failed to marshal plan subject", Cause: err}
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &PlanEnvelopeError{Reason: "failed to canonicalize plan subject JSON", Cause: err}
	}
	if err := ValidatePlanApprovalSubjectShape(canon); err != nil {
		return nil, &PlanEnvelopeError{Reason: "subject validation failed", Cause: err}
	}
	return canon, nil
}

// Digest computes the canonical sha256 hash formatted as 64 hex characters (matching plan_hash and approval hash vectors).
func (s *PlanApprovalSubject) Digest() (string, error) {
	canon, err := s.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return CanonicalHash(canon)
}

// PlanEnvelopeRecord represents an immutable plan envelope record (schema 2.1 in plan.schema.json).
type PlanEnvelopeRecord struct {
	SchemaVersion    string               `json:"schema_version"`
	PlanID           string               `json:"plan_id"`
	ProjectID        string               `json:"project_id"`
	PlanHash         string               `json:"plan_hash"`
	Subject          *PlanApprovalSubject `json:"subject"`
	AssessmentRef    string               `json:"assessment_ref"`
	SelectionRef     string               `json:"selection_ref"`
	PlanningBasisRef string               `json:"planning_basis_ref"`
}

// CanonicalJSON renders the plan envelope into canonical RFC 8785 bytes and validates it.
func (p *PlanEnvelopeRecord) CanonicalJSON() ([]byte, error) {
	if p == nil {
		return nil, &PlanEnvelopeError{Reason: "plan envelope is nil"}
	}
	if p.Subject == nil {
		return nil, &PlanEnvelopeError{Reason: "plan envelope subject is nil"}
	}

	subjectCanon, err := p.Subject.CanonicalJSON()
	if err != nil {
		return nil, &PlanEnvelopeError{Reason: "subject validation failed", Cause: err}
	}
	var subjectRaw json.RawMessage = subjectCanon

	recordMap := map[string]any{
		"schema_version":     p.SchemaVersion,
		"plan_id":            p.PlanID,
		"project_id":         p.ProjectID,
		"plan_hash":          p.PlanHash,
		"subject":            subjectRaw,
		"assessment_ref":     p.AssessmentRef,
		"selection_ref":      p.SelectionRef,
		"planning_basis_ref": p.PlanningBasisRef,
	}

	raw, err := json.Marshal(recordMap)
	if err != nil {
		return nil, &PlanEnvelopeError{Reason: "failed to marshal plan envelope", Cause: err}
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &PlanEnvelopeError{Reason: "failed to canonicalize plan envelope JSON", Cause: err}
	}
	if err := ValidatePlanEnvelope(canon); err != nil {
		return nil, err
	}
	return canon, nil
}

// Digest computes the canonical sha256 reference for the plan envelope.
func (p *PlanEnvelopeRecord) Digest() (string, error) {
	canon, err := p.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canon)
	return fmt.Sprintf("sha256:%x", h), nil
}

// ValidatePlanEnvelope validates a 2.1 plan envelope record against plan.schema.json
// and enforces cross-record consistency with the subject.
func ValidatePlanEnvelope(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil || record == nil {
		return &PlanEnvelopeError{Reason: "malformed JSON", Cause: err}
	}
	if profileString(record["schema_version"]) != "2.1" {
		return &PlanEnvelopeError{Reason: "expected schema_version 2.1"}
	}
	if err := ValidateMigrationPlan(raw); err != nil {
		return &PlanEnvelopeError{Reason: "plan envelope validation failed", Cause: err}
	}
	return nil
}

// ParsePlanEnvelope parses and validates a plan envelope JSON record.
func ParsePlanEnvelope(raw []byte) (*PlanEnvelopeRecord, error) {
	if err := ValidatePlanEnvelope(raw); err != nil {
		return nil, err
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &PlanEnvelopeError{Reason: "canonical JSON", Cause: err}
	}

	var envelope PlanEnvelopeRecord
	if err := json.Unmarshal(canon, &envelope); err != nil {
		return nil, &PlanEnvelopeError{Reason: "unmarshal plan envelope", Cause: err}
	}
	return &envelope, nil
}
