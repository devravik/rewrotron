package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Rollback outcome constants
const (
	RollbackRehearsalPass    = "PASS"
	RollbackRehearsalFail    = "FAIL"
	RollbackRehearsalBlocked = "BLOCKED"
)

// MutationType identifies the kind of mutation created by the target before rollback
type MutationType string

const (
	MutationTypeDatabaseRow     MutationType = "DATABASE_ROW"
	MutationTypeSessionState    MutationType = "SESSION_STATE"
	MutationTypeFileWrite       MutationType = "FILE_WRITE"
	MutationTypeExternalReceipt MutationType = "EXTERNAL_RECEIPT"
)

// TargetMutation records a concrete state mutation produced by the target under test.
type TargetMutation struct {
	MutationID   string       `json:"mutation_id"`
	MutationType MutationType `json:"mutation_type"`
	TargetRef    string       `json:"target_ref"` // table name, session key, file path, or external receipt ID
	PayloadHash  string       `json:"payload_hash"`
	Reversible   bool         `json:"reversible"`
}

// ReconciliationContract specifies how target mutations are reconciled upon rollback to source.
type ReconciliationContract struct {
	MutationType   MutationType `json:"mutation_type"`
	ReconcileOwner string       `json:"reconcile_owner"`  // responsible operator or automated procedure
	Strategy       string       `json:"strategy"`         // e.g. "REPLAY_SOURCE", "COMPENSATING_TX", "DISCARD_SESSION", "ARCHIVE"
	Verified       bool         `json:"verified"`
}

// RollbackTrigger specifies a documented failure trigger for initiating rollback.
type RollbackTrigger struct {
	TriggerID   string `json:"trigger_id"`
	Metric      string `json:"metric"`      // e.g. "MISMATCH", "LATENCY_P99", "ERROR_RATE", "SESSION_CORRUPTION"
	Threshold   string `json:"threshold"`   // e.g. "> 0.05", "EXCEEDED_BUDGET"
	DecisionBy  string `json:"decision_by"` // operator role authorized to trigger rollback
}

// RollbackRehearsalRecord models the immutable rehearsal record referenced by rollback_rehearsal_ref.
type RollbackRehearsalRecord struct {
	SchemaVersion            string                   `json:"schema_version"` // "1.0"
	RehearsalID              string                   `json:"rehearsal_id"`
	ProjectID                string                   `json:"project_id"`
	TargetTreeRef            string                   `json:"target_tree_ref"`
	EnvironmentRef           string                   `json:"environment_ref"`
	ExecutableCommands       []string                 `json:"executable_commands"`
	RepresentativeMutations  []TargetMutation         `json:"representative_mutations"`
	ReconciliationContracts  []ReconciliationContract `json:"reconciliation_contracts"`
	Triggers                 []RollbackTrigger        `json:"triggers"`
	SchemaOnlyGuaranteed     bool                     `json:"schema_only_guaranteed"` // if true, indicates dangerous schema-only rollback claim
	SourceContinuityVerified bool                     `json:"source_continuity_verified"`
	Outcome                  string                   `json:"outcome"` // PASS, FAIL, BLOCKED
	ReasonCodes              []string                 `json:"reason_codes"`
}

// RollbackRehearsalError indicates a failure in write-aware rollback validation.
type RollbackRehearsalError struct {
	Reason string
	Cause  error
}

func (e *RollbackRehearsalError) Error() string {
	return "rollback rehearsal validation failed: " + e.Reason
}

func (e *RollbackRehearsalError) Unwrap() error { return e.Cause }

// ValidateRollbackRehearsal evaluates a rollback rehearsal record per Test X34 and CUTOVER-RUNBOOK.md.
//
// Invariants enforced (Test X34):
// 1. Missing executable operator commands => BLOCKED / error.
// 2. Empty representative target mutations => BLOCKED ("an empty-data traffic switch is insufficient").
// 3. Schema-only rollback claim ("Preserving the schema reduces risk but does not make all new target writes reversible;
//    no rollback guarantee from preserved schema alone") => REFUSED / FAIL.
// 4. Any unreconciled target mutation (sessions, orders, payment/external receipts) => BLOCKED / FAIL.
// 5. Source continuity must be verified in staging.
func ValidateRollbackRehearsal(r *RollbackRehearsalRecord) error {
	if r == nil {
		return &RollbackRehearsalError{Reason: "rehearsal record is nil"}
	}

	if r.SchemaVersion != "1.0" {
		return &RollbackRehearsalError{Reason: "unsupported schema_version; expected 1.0"}
	}

	if r.ProjectID == "" {
		return &RollbackRehearsalError{Reason: "missing project_id"}
	}

	// Invariant 1: Missing executable operator commands
	if len(r.ExecutableCommands) == 0 {
		return &RollbackRehearsalError{Reason: "missing executable operator rollback commands (Test X34)"}
	}

	// Invariant 2: Empty mutations rehearsal rejected
	// CUTOVER-RUNBOOK: "Rehearsal must occur after representative target writes; an empty-data traffic switch is insufficient."
	if len(r.RepresentativeMutations) == 0 {
		return &RollbackRehearsalError{Reason: "no representative target mutations rehearsed; empty-data switch is insufficient (Test X34)"}
	}

	// Invariant 3: Refuse schema-only rollback guarantees
	// DEVELOPMENT-PLAN: "no rollback guarantee from preserved schema alone"
	// CUTOVER-RUNBOOK: "Preserving the schema reduces risk but does not make all new target writes reversible."
	if r.SchemaOnlyGuaranteed {
		return &RollbackRehearsalError{
			Reason: "schema-only rollback guarantee rejected; preserved schema does not make new target writes reversible (Test X34)",
		}
	}

	// Invariant 4: Every target mutation kind must have a verified reconciliation contract
	// Build map of verified reconciliation strategies
	reconciledKinds := make(map[MutationType]bool)
	for _, rc := range r.ReconciliationContracts {
		if rc.Verified && rc.ReconcileOwner != "" && rc.Strategy != "" {
			reconciledKinds[rc.MutationType] = true
		}
	}

	for _, m := range r.RepresentativeMutations {
		// If mutation is marked irreversible and has no verified reconciliation contract, rollback cannot proceed safely
		if !reconciledKinds[m.MutationType] {
			return &RollbackRehearsalError{
				Reason: fmt.Sprintf("unreconciled target mutation of kind %s for ref %s blocks release (Test X34)", m.MutationType, m.TargetRef),
			}
		}
	}

	// Invariant 5: Source continuity must be verified
	if !r.SourceContinuityVerified {
		return &RollbackRehearsalError{
			Reason: "source continuity after rollback is not verified in staging (Test X34)",
		}
	}

	// Invariant 6: Concrete triggers must be defined
	if len(r.Triggers) == 0 {
		return &RollbackRehearsalError{
			Reason: "no concrete rollback triggers defined (CUTOVER-RUNBOOK.md)",
		}
	}

	if r.Outcome != RollbackRehearsalPass {
		return &RollbackRehearsalError{
			Reason: fmt.Sprintf("rehearsal outcome is %s; release blocked", r.Outcome),
		}
	}

	return nil
}

// CanonicalJSON computes RFC 8785 canonical bytes for a RollbackRehearsalRecord.
func (r *RollbackRehearsalRecord) CanonicalJSON() ([]byte, error) {
	if r == nil {
		return nil, &RollbackRehearsalError{Reason: "rehearsal record is nil"}
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, &RollbackRehearsalError{Reason: "failed to marshal rollback rehearsal", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the canonical sha256 reference for the rollback rehearsal record.
func (r *RollbackRehearsalRecord) Digest() (string, error) {
	canon, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canon)
	return fmt.Sprintf("sha256:%x", h), nil
}
