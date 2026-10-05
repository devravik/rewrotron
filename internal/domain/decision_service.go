package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// DecisionAction represents the valid actions available when interacting with recommendations or decisions.
type DecisionAction string

const (
	ActionAccept            DecisionAction = "ACCEPT"
	ActionReject            DecisionAction = "REJECT"
	ActionSelectStrategy    DecisionAction = "SELECT_STRATEGY"
	ActionSupplyAlternative DecisionAction = "SUPPLY_ALTERNATIVE"
	ActionRequestEvidence   DecisionAction = "REQUEST_EVIDENCE"
	ActionApproveException  DecisionAction = "APPROVE_EXCEPTION"
)

// DecisionStrategy represents the supported strategies for resolving dependencies or components.
type DecisionStrategy string

const (
	StrategyDirectUpgrade DecisionStrategy = "DIRECT_UPGRADE"
	StrategyAdapter       DecisionStrategy = "ADAPTER"
	StrategyReplacement   DecisionStrategy = "REPLACEMENT"
	StrategyReimplement   DecisionStrategy = "REIMPLEMENT"
	StrategyRetain        DecisionStrategy = "RETAIN"
)

// MigrationProgressStatus represents the lifecycle state of a migration entity.
type MigrationProgressStatus string

const (
	ProgressPlanned    MigrationProgressStatus = "PLANNED"
	ProgressGenerating MigrationProgressStatus = "GENERATING"
	ProgressBuilding   MigrationProgressStatus = "BUILDING"
	ProgressVerifying  MigrationProgressStatus = "VERIFYING"
	ProgressVerified   MigrationProgressStatus = "VERIFIED"
	ProgressBlocked    MigrationProgressStatus = "BLOCKED"
)

// DecisionSafetyError describes an attempt to use an exception as a verdict waiver or unauthorized activation.
type DecisionSafetyError struct {
	Reason string
	Cause  error
}

// Error identifies the safety violation.
func (e *DecisionSafetyError) Error() string {
	return "decision safety violation: " + e.Reason
}

// Unwrap preserves underlying cause.
func (e *DecisionSafetyError) Unwrap() error { return e.Cause }

// DecisionProposalDocument represents an immutable proposed decision conforming to
// specs/project/decision.schema.json (version 2.0).
type DecisionProposalDocument struct {
	SchemaVersion     string   `json:"schema_version"`
	DecisionID        string   `json:"decision_id"`
	ProjectID         string   `json:"project_id"`
	Action            string   `json:"action"`
	SubjectRef        string   `json:"subject_ref"`
	Strategy          *string  `json:"strategy"`
	ImplementationRef *string  `json:"implementation_ref"`
	ExceptionScopeRef *string  `json:"exception_scope_ref"`
	Rationale         string   `json:"rationale"`
	EvidenceRefs      []string `json:"evidence_refs"`
	Status            string   `json:"status"` // Strictly "PROPOSED"
}

// CanonicalJSON renders the DecisionProposalDocument into canonical RFC 8785 JSON bytes.
func (d *DecisionProposalDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &DecisionError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical DecisionProposalDocument.
func (d *DecisionProposalDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// MigrationProgressDocument represents an execution progress record conforming to
// specs/migration/progress.schema.json (version 2.0).
type MigrationProgressDocument struct {
	SchemaVersion       string  `json:"schema_version"`
	ProjectID           string  `json:"project_id"`
	PlanHash            string  `json:"plan_hash"`
	DecisionOrModelRef  string  `json:"decision_or_model_ref"`
	EntityID            string  `json:"entity_id"`
	Status              string  `json:"status"`
	ThroughSequence     int     `json:"through_sequence"`
	TargetTreeRef       *string `json:"target_tree_ref"`
	AcceptanceReportRef *string `json:"acceptance_report_ref"`
}

// CanonicalJSON renders the MigrationProgressDocument into canonical RFC 8785 JSON bytes.
func (p *MigrationProgressDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, &MigrationProgressError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical MigrationProgressDocument.
func (p *MigrationProgressDocument) Digest() (string, error) {
	cjson, err := p.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ProposeDecisionRequest defines the parameters needed to create a new decision proposal.
type ProposeDecisionRequest struct {
	DecisionID        string
	ProjectID         string
	Action            DecisionAction
	SubjectRef        string
	Strategy          *DecisionStrategy
	ImplementationRef *string
	ExceptionScopeRef *string
	Rationale         string
	EvidenceRefs      []string
}

// ProposeDecision creates and validates an immutable decision proposal record.
// All proposals have status strictly "PROPOSED" (Test X16: proposals never self-activate;
// only exact Gate 2 confirms and activates decisions).
func ProposeDecision(req *ProposeDecisionRequest) (*DecisionProposalDocument, error) {
	if req == nil {
		return nil, &DecisionError{Reason: "request must not be nil"}
	}
	if !validAssessmentReportID(req.DecisionID) {
		return nil, &DecisionError{Reason: "decision_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !validAssessmentReportID(req.ProjectID) {
		return nil, &DecisionError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !validRefDigest(req.SubjectRef) {
		return nil, &DecisionError{Reason: "subject_ref must be a valid sha256 reference"}
	}
	if strings.TrimSpace(req.Rationale) == "" {
		return nil, &DecisionError{Reason: "rationale is required and must not be empty"}
	}

	actionStr := string(req.Action)
	if _, ok := allowedDecisionActions[actionStr]; !ok {
		return nil, &DecisionError{Reason: fmt.Sprintf("invalid decision action: %q", actionStr)}
	}

	var stratStr *string
	if req.Strategy != nil {
		s := string(*req.Strategy)
		if _, ok := allowedDecisionStrategies[s]; !ok {
			return nil, &DecisionError{Reason: fmt.Sprintf("invalid decision strategy: %q", s)}
		}
		stratStr = &s
	}

	// Action-specific validation
	switch req.Action {
	case ActionSelectStrategy:
		if stratStr == nil {
			return nil, &DecisionError{Reason: "strategy is required when action is SELECT_STRATEGY"}
		}
	case ActionSupplyAlternative:
		if req.ImplementationRef == nil || !validRefDigest(*req.ImplementationRef) {
			return nil, &DecisionError{Reason: "valid implementation_ref is required when action is SUPPLY_ALTERNATIVE"}
		}
	case ActionApproveException:
		if req.ExceptionScopeRef == nil || !validRefDigest(*req.ExceptionScopeRef) {
			return nil, &DecisionError{Reason: "valid exception_scope_ref is required when action is APPROVE_EXCEPTION"}
		}
	}

	evRefs := req.EvidenceRefs
	if evRefs == nil {
		evRefs = []string{}
	}
	for _, ref := range evRefs {
		if !validRefDigest(ref) {
			return nil, &DecisionError{Reason: fmt.Sprintf("evidence ref %q must be a valid sha256 reference", ref)}
		}
	}

	// Sort and deduplicate evidence_refs
	sort.Strings(evRefs)
	for i := 1; i < len(evRefs); i++ {
		if evRefs[i] == evRefs[i-1] {
			return nil, &DecisionError{Reason: fmt.Sprintf("duplicate evidence ref: %s", evRefs[i])}
		}
	}

	doc := &DecisionProposalDocument{
		SchemaVersion:     "2.0",
		DecisionID:        req.DecisionID,
		ProjectID:         req.ProjectID,
		Action:            actionStr,
		SubjectRef:        req.SubjectRef,
		Strategy:          stratStr,
		ImplementationRef: req.ImplementationRef,
		ExceptionScopeRef: req.ExceptionScopeRef,
		Rationale:         req.Rationale,
		EvidenceRefs:      evRefs,
		Status:            "PROPOSED", // Enforced invariant: only Gate 2 activates
	}

	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return nil, &DecisionError{Reason: "failed generating canonical JSON", Cause: err}
	}
	if err := ValidateDecision(cjson); err != nil {
		return nil, &DecisionError{Reason: "decision document failed schema validation", Cause: err}
	}

	return doc, nil
}

// ValidateDecisionExceptionSafety enforces Test X16 invariants:
// 1. A proposed exception cannot grant a PASS verdict or safety waiver.
// 2. UNKNOWN and MISMATCH never become PASS by waiver or proposed exception.
// 3. Exceptions only adjust approved scope in Gate 2; they never alter verifier results.
func ValidateDecisionExceptionSafety(decision *DecisionProposalDocument, attemptedVerdict string) error {
	if decision == nil {
		return &DecisionSafetyError{Reason: "decision must not be nil"}
	}
	if decision.Action == string(ActionApproveException) {
		if attemptedVerdict == "PASS" {
			return &DecisionSafetyError{
				Reason: "proposed exception cannot grant PASS or safety waiver (Test X16: UNKNOWN/MISMATCH cannot become PASS by exception)",
			}
		}
	}
	return nil
}

// RecordProgressRequest defines inputs for recording a migration progress step.
type RecordProgressRequest struct {
	ProjectID           string
	PlanHash            string
	DecisionOrModelRef  string
	EntityID            string
	Status              MigrationProgressStatus
	ThroughSequence     int
	TargetTreeRef       *string
	AcceptanceReportRef *string
}

// RecordMigrationProgress creates and validates an execution progress record conforming to progress.schema.json.
// Invariant Test X17: progress records are separate mutable/reconstructable execution projections.
// Progress updates NEVER mutate the approved plan or alter plan_hash.
func RecordMigrationProgress(req *RecordProgressRequest) (*MigrationProgressDocument, error) {
	if req == nil {
		return nil, &MigrationProgressError{Reason: "request must not be nil"}
	}
	if !validAssessmentReportID(req.ProjectID) {
		return nil, &MigrationProgressError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !isManifestDigest(req.PlanHash) {
		return nil, &MigrationProgressError{Reason: "plan_hash must match pattern ^[a-f0-9]{64}$"}
	}
	if !validRefDigest(req.DecisionOrModelRef) {
		return nil, &MigrationProgressError{Reason: "decision_or_model_ref must be a valid sha256 reference"}
	}
	if !validAssessmentReportID(req.EntityID) {
		return nil, &MigrationProgressError{Reason: "entity_id must match pattern ^[a-z0-9_-]+$"}
	}
	if req.ThroughSequence < 0 {
		return nil, &MigrationProgressError{Reason: "through_sequence must be non-negative"}
	}

	statusStr := string(req.Status)
	if _, ok := allowedMigrationProgressStatuses[statusStr]; !ok {
		return nil, &MigrationProgressError{Reason: fmt.Sprintf("invalid progress status: %q", statusStr)}
	}

	if req.TargetTreeRef != nil && !validRefDigest(*req.TargetTreeRef) {
		return nil, &MigrationProgressError{Reason: "target_tree_ref must be a valid sha256 reference"}
	}
	if req.AcceptanceReportRef != nil && !validRefDigest(*req.AcceptanceReportRef) {
		return nil, &MigrationProgressError{Reason: "acceptance_report_ref must be a valid sha256 reference"}
	}

	doc := &MigrationProgressDocument{
		SchemaVersion:       "2.0",
		ProjectID:           req.ProjectID,
		PlanHash:            req.PlanHash,
		DecisionOrModelRef:  req.DecisionOrModelRef,
		EntityID:            req.EntityID,
		Status:              statusStr,
		ThroughSequence:     req.ThroughSequence,
		TargetTreeRef:       req.TargetTreeRef,
		AcceptanceReportRef: req.AcceptanceReportRef,
	}

	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return nil, &MigrationProgressError{Reason: "failed generating canonical JSON", Cause: err}
	}
	if err := ValidateMigrationProgress(cjson); err != nil {
		return nil, &MigrationProgressError{Reason: "progress document failed schema validation", Cause: err}
	}

	return doc, nil
}

// VerifyPlanHashStability verifies Test X17 invariant:
// Progress transitions (PLANNED -> GENERATING -> BUILDING -> VERIFYING -> VERIFIED)
// leave the original plan hash completely unchanged.
func VerifyPlanHashStability(originalPlanHash string, progressEvents []*MigrationProgressDocument) (bool, error) {
	if !isManifestDigest(originalPlanHash) {
		return false, fmt.Errorf("invalid original plan hash: %s", originalPlanHash)
	}
	for _, prog := range progressEvents {
		if prog.PlanHash != originalPlanHash {
			return false, fmt.Errorf("plan hash mutated during progress transition %s: expected %s, got %s", prog.Status, originalPlanHash, prog.PlanHash)
		}
	}
	return true, nil
}

// CheckApprovalInvalidationOnDecisionChange verifies Test X17 invariant:
// When a decision's strategy or lowering changes, any existing approval whose decision_refs
// contains the old decision ref is invalidated.
func CheckApprovalInvalidationOnDecisionChange(planApprovedDecisionRefs []string, oldDecisionDigest, newDecisionDigest string) (isInvalidated bool, reason string) {
	if oldDecisionDigest == newDecisionDigest {
		return false, "no change in decision digest"
	}
	for _, ref := range planApprovedDecisionRefs {
		if ref == oldDecisionDigest {
			return true, fmt.Sprintf("plan approval invalidated (Test X17): decision %s was modified (new digest: %s); approval requires Gate 2 re-confirmation", oldDecisionDigest, newDecisionDigest)
		}
	}
	return false, "plan decision refs do not contain the modified decision"
}
