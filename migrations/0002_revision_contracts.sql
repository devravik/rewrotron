-- Specification revision 2; preserve 0001 unchanged. Apply transactionally after ownership preflight.
CREATE UNIQUE INDEX artifacts_project_id ON artifacts(project_id,id);
CREATE UNIQUE INDEX operations_project_id ON operations(project_id,id);
CREATE UNIQUE INDEX missions_active_work ON missions(project_id,type,fingerprint)
 WHERE status IN ('QUEUED','LEASED','RUNNING','BLOCKED');
CREATE UNIQUE INDEX case_results_once ON case_results(project_id,run_id,case_id);

CREATE TABLE approval_subjects (
 id TEXT PRIMARY KEY,
 project_id TEXT NOT NULL REFERENCES projects(id),
 kind TEXT NOT NULL CHECK(kind IN ('SOURCE','PLAN','RELEASE')),
 subject_hash TEXT NOT NULL CHECK(length(subject_hash)=64 AND subject_hash NOT GLOB '*[^a-f0-9]*'),
 document_ref TEXT NOT NULL,
 schema_version TEXT NOT NULL CHECK(schema_version='2.0'),
 canonicalization TEXT NOT NULL CHECK(canonicalization='rwt-cjson-1'),
 FOREIGN KEY(project_id,document_ref) REFERENCES artifacts(project_id,id),
 UNIQUE(project_id,subject_hash)
);
CREATE TABLE source_snapshots (
 id TEXT PRIMARY KEY,
 project_id TEXT NOT NULL REFERENCES projects(id),
 content_root_hash TEXT NOT NULL CHECK(length(content_root_hash)=64 AND content_root_hash NOT GLOB '*[^a-f0-9]*'),
 document_ref TEXT NOT NULL,
 FOREIGN KEY(project_id,document_ref) REFERENCES artifacts(project_id,id),
 UNIQUE(project_id,id)
);
CREATE TABLE decision_records (
 id TEXT PRIMARY KEY,
 project_id TEXT NOT NULL REFERENCES projects(id),
 document_ref TEXT NOT NULL,
 FOREIGN KEY(project_id,document_ref) REFERENCES artifacts(project_id,id),
 UNIQUE(project_id,id)
);
CREATE TABLE resource_claims (
 project_id TEXT NOT NULL REFERENCES projects(id),
 resource_key TEXT NOT NULL,
 mission_id TEXT NOT NULL,
 lease_owner TEXT NOT NULL,
 fencing_token INTEGER NOT NULL CHECK(fencing_token>=1),
 expires_at TEXT NOT NULL,
 PRIMARY KEY(project_id,resource_key),
 FOREIGN KEY(project_id,mission_id) REFERENCES missions(project_id,id)
);
CREATE TABLE operation_owners (
 project_id TEXT NOT NULL REFERENCES projects(id),
 operation_id TEXT NOT NULL,
 mode TEXT NOT NULL CHECK(mode IN ('FOREGROUND','SUPERVISED')),
 owner_instance_id TEXT NOT NULL,
 host_boot_id TEXT NOT NULL,
 process_id TEXT NOT NULL,
 process_start_identity TEXT NOT NULL,
 heartbeat_at TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('RUNNING','STOPPING','STOPPED','STALE')),
 PRIMARY KEY(project_id,operation_id),
 FOREIGN KEY(project_id,operation_id) REFERENCES operations(project_id,id)
);
CREATE TRIGGER approval_subjects_no_update BEFORE UPDATE ON approval_subjects
BEGIN SELECT RAISE(ABORT, 'approval_subjects is append-only'); END;
CREATE TRIGGER approval_subjects_no_delete BEFORE DELETE ON approval_subjects
BEGIN SELECT RAISE(ABORT, 'approval_subjects is append-only'); END;
CREATE TRIGGER source_snapshots_no_update BEFORE UPDATE ON source_snapshots
BEGIN SELECT RAISE(ABORT, 'source_snapshots is append-only'); END;
CREATE TRIGGER source_snapshots_no_delete BEFORE DELETE ON source_snapshots
BEGIN SELECT RAISE(ABORT, 'source_snapshots is append-only'); END;
CREATE TRIGGER decision_records_no_update BEFORE UPDATE ON decision_records
BEGIN SELECT RAISE(ABORT, 'decision_records is append-only'); END;
CREATE TRIGGER decision_records_no_delete BEFORE DELETE ON decision_records
BEGIN SELECT RAISE(ABORT, 'decision_records is append-only'); END;
CREATE TRIGGER source_detections_result_ref_owner_insert BEFORE INSERT ON source_detections
WHEN NEW.result_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.result_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project result_ref'); END;
CREATE TRIGGER source_detections_result_ref_owner_update BEFORE UPDATE ON source_detections
WHEN NEW.result_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.result_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project result_ref'); END;
CREATE TRIGGER baselines_manifest_ref_owner_insert BEFORE INSERT ON baselines
WHEN NEW.manifest_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.manifest_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project manifest_ref'); END;
CREATE TRIGGER baselines_manifest_ref_owner_update BEFORE UPDATE ON baselines
WHEN NEW.manifest_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.manifest_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project manifest_ref'); END;
CREATE TRIGGER user_confirmations_reviewed_payload_ref_owner_insert BEFORE INSERT ON user_confirmations
WHEN NEW.reviewed_payload_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.reviewed_payload_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project reviewed_payload_ref'); END;
CREATE TRIGGER user_confirmations_reviewed_payload_ref_owner_update BEFORE UPDATE ON user_confirmations
WHEN NEW.reviewed_payload_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.reviewed_payload_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project reviewed_payload_ref'); END;
CREATE TRIGGER ir_documents_document_ref_owner_insert BEFORE INSERT ON ir_documents
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER ir_documents_document_ref_owner_update BEFORE UPDATE ON ir_documents
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER entrypoints_evidence_ref_owner_insert BEFORE INSERT ON entrypoints
WHEN NEW.evidence_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.evidence_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project evidence_ref'); END;
CREATE TRIGGER entrypoints_evidence_ref_owner_update BEFORE UPDATE ON entrypoints
WHEN NEW.evidence_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.evidence_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project evidence_ref'); END;
CREATE TRIGGER contracts_document_ref_owner_insert BEFORE INSERT ON contracts
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER contracts_document_ref_owner_update BEFORE UPDATE ON contracts
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER behavior_cases_document_ref_owner_insert BEFORE INSERT ON behavior_cases
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER behavior_cases_document_ref_owner_update BEFORE UPDATE ON behavior_cases
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER workflows_document_ref_owner_insert BEFORE INSERT ON workflows
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER workflows_document_ref_owner_update BEFORE UPDATE ON workflows
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER migration_plans_document_ref_owner_insert BEFORE INSERT ON migration_plans
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER migration_plans_document_ref_owner_update BEFORE UPDATE ON migration_plans
WHEN NEW.document_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.document_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project document_ref'); END;
CREATE TRIGGER observations_trace_ref_owner_insert BEFORE INSERT ON observations
WHEN NEW.trace_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.trace_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project trace_ref'); END;
CREATE TRIGGER observations_trace_ref_owner_update BEFORE UPDATE ON observations
WHEN NEW.trace_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.trace_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project trace_ref'); END;
CREATE TRIGGER verification_runs_report_ref_owner_insert BEFORE INSERT ON verification_runs
WHEN NEW.report_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.report_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project report_ref'); END;
CREATE TRIGGER verification_runs_report_ref_owner_update BEFORE UPDATE ON verification_runs
WHEN NEW.report_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.report_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project report_ref'); END;
CREATE TRIGGER mismatches_diff_ref_owner_insert BEFORE INSERT ON mismatches
WHEN NEW.diff_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.diff_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project diff_ref'); END;
CREATE TRIGGER mismatches_diff_ref_owner_update BEFORE UPDATE ON mismatches
WHEN NEW.diff_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.diff_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project diff_ref'); END;
CREATE TRIGGER experiments_context_ref_owner_insert BEFORE INSERT ON experiments
WHEN NEW.context_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.context_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project context_ref'); END;
CREATE TRIGGER experiments_context_ref_owner_update BEFORE UPDATE ON experiments
WHEN NEW.context_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.context_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project context_ref'); END;
CREATE TRIGGER repair_attempts_patch_ref_owner_insert BEFORE INSERT ON repair_attempts
WHEN NEW.patch_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.patch_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project patch_ref'); END;
CREATE TRIGGER repair_attempts_patch_ref_owner_update BEFORE UPDATE ON repair_attempts
WHEN NEW.patch_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.patch_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project patch_ref'); END;
CREATE TRIGGER operations_checkpoint_ref_owner_insert BEFORE INSERT ON operations
WHEN NEW.checkpoint_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.checkpoint_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project checkpoint_ref'); END;
CREATE TRIGGER operations_checkpoint_ref_owner_update BEFORE UPDATE ON operations
WHEN NEW.checkpoint_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.checkpoint_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project checkpoint_ref'); END;
CREATE TRIGGER idempotency_requests_result_ref_owner_insert BEFORE INSERT ON idempotency_requests
WHEN NEW.result_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.result_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project result_ref'); END;
CREATE TRIGGER idempotency_requests_result_ref_owner_update BEFORE UPDATE ON idempotency_requests
WHEN NEW.result_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.result_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project result_ref'); END;
CREATE TRIGGER idempotency_requests_operation_id_owner_insert BEFORE INSERT ON idempotency_requests
WHEN NEW.operation_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM operations WHERE id=NEW.operation_id AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project operation_id'); END;
CREATE TRIGGER idempotency_requests_operation_id_owner_update BEFORE UPDATE ON idempotency_requests
WHEN NEW.operation_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM operations WHERE id=NEW.operation_id AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project operation_id'); END;
CREATE TRIGGER findings_evidence_ref_owner_insert BEFORE INSERT ON findings
WHEN NEW.evidence_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.evidence_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project evidence_ref'); END;
CREATE TRIGGER findings_evidence_ref_owner_update BEFORE UPDATE ON findings
WHEN NEW.evidence_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.evidence_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project evidence_ref'); END;
CREATE TRIGGER benchmarks_report_ref_owner_insert BEFORE INSERT ON benchmarks
WHEN NEW.report_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.report_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project report_ref'); END;
CREATE TRIGGER benchmarks_report_ref_owner_update BEFORE UPDATE ON benchmarks
WHEN NEW.report_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.report_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project report_ref'); END;
CREATE TRIGGER lessons_evidence_ref_owner_insert BEFORE INSERT ON lessons
WHEN NEW.evidence_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.evidence_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project evidence_ref'); END;
CREATE TRIGGER lessons_evidence_ref_owner_update BEFORE UPDATE ON lessons
WHEN NEW.evidence_ref IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.evidence_ref AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project evidence_ref'); END;
CREATE TRIGGER artifact_links_artifact_id_owner_insert BEFORE INSERT ON artifact_links
WHEN NEW.artifact_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.artifact_id AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project artifact_id'); END;
CREATE TRIGGER artifact_links_artifact_id_owner_update BEFORE UPDATE ON artifact_links
WHEN NEW.artifact_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM artifacts WHERE id=NEW.artifact_id AND project_id=NEW.project_id)
BEGIN SELECT RAISE(ABORT, 'cross-project artifact_id'); END;
CREATE TRIGGER artifacts_no_update BEFORE UPDATE ON artifacts
BEGIN SELECT RAISE(ABORT, 'artifact identity is immutable'); END;
CREATE TRIGGER artifacts_no_delete BEFORE DELETE ON artifacts
BEGIN SELECT RAISE(ABORT, 'artifact metadata is retained'); END;
