package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// PreflightError identifies incompatible retained data or a failed preflight read.
// Cause is nil for a detected data conflict, and non-nil for an operational error.
type PreflightError struct {
	Check string
	Cause error
}

// Error identifies the failed check without exposing retained customer data.
func (e *PreflightError) Error() string { return "SQLite revision-2 preflight: " + e.Check }

// Unwrap preserves cancellation and database errors for caller inspection.
func (e *PreflightError) Unwrap() error { return e.Cause }

// PreflightRevision2 checks existing revision-1 data against revision-2 ownership
// and uniqueness requirements. It executes only SELECTs, never repairs data or
// applies DDL. The caller supplies an admitted/configured connection's transaction
// and retains exclusive migration ownership through preflight and application.
// A successful check grants no approval, schema version or migration authority.
// Operational failures are not evidence of a clean preflight.
func PreflightRevision2(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return &PreflightError{Check: "transaction required"}
	}
	for _, ref := range revision2References() {
		// All identifiers are private literals derived from immutable migration
		// 0002, never caller-controlled SQL or paths.
		query := fmt.Sprintf(`SELECT EXISTS (
            SELECT 1 FROM %s AS child WHERE child.%s IS NOT NULL
            AND NOT EXISTS (SELECT 1 FROM %s AS parent
                WHERE parent.id=child.%s AND parent.project_id=child.project_id))`,
			ref.table, ref.column, ref.parent, ref.column)
		if err := rejectExisting(ctx, tx, ref.table+"."+ref.column+" ownership", query); err != nil {
			return err
		}
	}
	if err := rejectExisting(ctx, tx, "duplicate active mission work", `SELECT EXISTS (
        SELECT 1 FROM missions WHERE status IN ('QUEUED','LEASED','RUNNING','BLOCKED')
        GROUP BY project_id,type,fingerprint HAVING count(*) > 1)`); err != nil {
		return err
	}
	return rejectExisting(ctx, tx, "duplicate case results", `SELECT EXISTS (
        SELECT 1 FROM case_results GROUP BY project_id,run_id,case_id HAVING count(*) > 1)`)
}

func rejectExisting(ctx context.Context, tx *sql.Tx, check, query string) error {
	var found bool
	if err := tx.QueryRowContext(ctx, query).Scan(&found); err != nil {
		return &PreflightError{Check: check, Cause: err}
	}
	if found {
		return &PreflightError{Check: check}
	}
	return nil
}

type ownershipReference struct {
	table  string
	column string
	parent string
}

func revision2References() []ownershipReference {
	return []ownershipReference{
		{table: "source_detections", column: "result_ref", parent: "artifacts"},
		{table: "baselines", column: "manifest_ref", parent: "artifacts"},
		{table: "user_confirmations", column: "reviewed_payload_ref", parent: "artifacts"},
		{table: "ir_documents", column: "document_ref", parent: "artifacts"},
		{table: "entrypoints", column: "evidence_ref", parent: "artifacts"},
		{table: "contracts", column: "document_ref", parent: "artifacts"},
		{table: "behavior_cases", column: "document_ref", parent: "artifacts"},
		{table: "workflows", column: "document_ref", parent: "artifacts"},
		{table: "migration_plans", column: "document_ref", parent: "artifacts"},
		{table: "observations", column: "trace_ref", parent: "artifacts"},
		{table: "verification_runs", column: "report_ref", parent: "artifacts"},
		{table: "mismatches", column: "diff_ref", parent: "artifacts"},
		{table: "experiments", column: "context_ref", parent: "artifacts"},
		{table: "repair_attempts", column: "patch_ref", parent: "artifacts"},
		{table: "operations", column: "checkpoint_ref", parent: "artifacts"},
		{table: "idempotency_requests", column: "result_ref", parent: "artifacts"},
		{table: "idempotency_requests", column: "operation_id", parent: "operations"},
		{table: "findings", column: "evidence_ref", parent: "artifacts"},
		{table: "benchmarks", column: "report_ref", parent: "artifacts"},
		{table: "lessons", column: "evidence_ref", parent: "artifacts"},
		{table: "artifact_links", column: "artifact_id", parent: "artifacts"},
	}
}
