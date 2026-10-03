-- Initial project schema: no deployed predecessor exists.
-- Connection pragmas and schema_migrations are managed by the runner.

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    root_path TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK(status IN ('INITIALIZED','DETECTED','SOURCE_CONFIRMED','ANALYZED','MODELED','PLAN_SYNTHESIZED','PLAN_CONFIRMED','MIGRATING','VERIFIED','BENCHMARKED','READY_CUTOVER')),
    revision INTEGER NOT NULL DEFAULT 0 CHECK(revision >= 0),
    configuration_hash TEXT,
    plan_hash TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE artifacts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    sha256 TEXT NOT NULL CHECK(length(sha256)=64 AND sha256 NOT GLOB '*[^a-f0-9]*'),
    relative_path TEXT NOT NULL,
    media_type TEXT NOT NULL,
    size_bytes INTEGER NOT NULL CHECK(size_bytes >= 0),
    redaction_json TEXT NOT NULL CHECK(json_valid(redaction_json)),
    created_at TEXT NOT NULL,
    UNIQUE(project_id, sha256),
    UNIQUE(project_id, relative_path)
);

CREATE TABLE source_detections (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    source_manifest_hash TEXT NOT NULL,
    result_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    UNIQUE(project_id,id)
);

CREATE TABLE baselines (
    configuration_hash TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    detection_id TEXT NOT NULL,
    manifest_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    facts_json TEXT NOT NULL CHECK(json_valid(facts_json)),
    overrides_json TEXT NOT NULL CHECK(json_valid(overrides_json)),
    environment_hash TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,detection_id) REFERENCES source_detections(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,configuration_hash)
);

CREATE TABLE approval_tokens (
    token_hash TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    gate_type TEXT NOT NULL CHECK(gate_type IN ('GATE_1_SOURCE_DETECTION','GATE_2_PLAN_CONFIRMED','GATE_3_CUTOVER')),
    subject_hash TEXT NOT NULL,
    issued_by TEXT NOT NULL,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    consumed_at TEXT
);

CREATE TABLE user_confirmations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    gate_type TEXT NOT NULL CHECK(gate_type IN ('GATE_1_SOURCE_DETECTION','GATE_2_PLAN_CONFIRMED','GATE_3_CUTOVER')),
    subject_hash TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    confirmed_by TEXT NOT NULL,
    channel TEXT NOT NULL CHECK(channel IN ('CLI','MCP')),
    action_id TEXT NOT NULL,
    reviewed_payload_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    overrides_json TEXT NOT NULL CHECK(json_valid(overrides_json)),
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT,
    UNIQUE(project_id,action_id)
);

CREATE TABLE events (
    sequence INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id TEXT NOT NULL UNIQUE,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    payload_json TEXT NOT NULL CHECK(json_valid(payload_json)),
    request_id TEXT NOT NULL,
    timestamp TEXT NOT NULL
);

CREATE TABLE source_files (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    syntax_status TEXT NOT NULL CHECK(syntax_status IN ('PARSED','PARTIAL','UNSUPPORTED','FAILED')),
    classification TEXT NOT NULL CHECK(classification IN ('ACTIVE','DORMANT','DEAD','UNKNOWN')),
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT,
    UNIQUE(project_id,configuration_hash,relative_path),
    UNIQUE(project_id,id)
);

CREATE TABLE source_symbols (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    file_id TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    start_line INTEGER NOT NULL CHECK(start_line >= 1),
    end_line INTEGER NOT NULL CHECK(end_line >= start_line),
    classification TEXT NOT NULL CHECK(classification IN ('ACTIVE','DORMANT','DEAD','UNKNOWN')),
    FOREIGN KEY(project_id,file_id) REFERENCES source_files(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE ir_documents (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    tier TEXT NOT NULL CHECK(tier IN ('SOURCE','APPLICATION','BEHAVIOR')),
    document_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    content_hash TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE entrypoints (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('http','cli','cron','event')),
    dispatch_json TEXT NOT NULL CHECK(json_valid(dispatch_json)),
    classification TEXT NOT NULL CHECK(classification IN ('ACTIVE','DORMANT','DEAD','UNKNOWN')),
    evidence_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE contracts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    operation_id TEXT NOT NULL,
    version TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision >= 1),
    document_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    content_hash TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('DRAFT','LOCKED','DEPRECATED')),
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT,
    UNIQUE(project_id,operation_id,revision),
    UNIQUE(project_id,id)
);

CREATE TABLE contract_entrypoints (
    project_id TEXT NOT NULL,
    contract_id TEXT NOT NULL,
    entrypoint_id TEXT NOT NULL,
    PRIMARY KEY(project_id,contract_id,entrypoint_id),
    FOREIGN KEY(project_id,contract_id) REFERENCES contracts(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,entrypoint_id) REFERENCES entrypoints(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE behavior_cases (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    contract_id TEXT NOT NULL,
    document_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    content_hash TEXT NOT NULL,
    fixture_hash TEXT NOT NULL,
    policy_hash TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,contract_id) REFERENCES contracts(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE workflows (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    document_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT
);

CREATE TABLE migration_plans (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    configuration_hash TEXT NOT NULL,
    plan_hash TEXT NOT NULL UNIQUE,
    corpus_hash TEXT NOT NULL,
    policy_hash TEXT NOT NULL,
    document_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,configuration_hash) REFERENCES baselines(project_id,configuration_hash) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE migration_units (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    unit_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('DISCOVERED','PLANNED','GENERATING','GENERATED','VERIFYING','VERIFIED','COMPLETED','BLOCKED')),
    priority INTEGER NOT NULL CHECK(priority >= 0),
    record_json TEXT NOT NULL CHECK(json_valid(record_json)),
    target_tree_hash TEXT,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(project_id,plan_id) REFERENCES migration_plans(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,plan_id,unit_id),
    UNIQUE(project_id,id)
);

CREATE TABLE unit_cases (
    project_id TEXT NOT NULL,
    unit_id TEXT NOT NULL,
    case_id TEXT NOT NULL,
    PRIMARY KEY(project_id,unit_id,case_id),
    FOREIGN KEY(project_id,unit_id) REFERENCES migration_units(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,case_id) REFERENCES behavior_cases(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE observations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    case_id TEXT,
    trace_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    environment TEXT NOT NULL CHECK(environment IN ('source_legacy','target_modern')),
    completeness_json TEXT NOT NULL CHECK(json_valid(completeness_json)),
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,case_id) REFERENCES behavior_cases(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE verification_runs (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    target_tree_hash TEXT NOT NULL,
    corpus_hash TEXT NOT NULL,
    policy_hash TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('COMPLETED','FAILED','CANCELLED')),
    verdict TEXT NOT NULL CHECK(verdict IN ('PASS','MISMATCH','UNKNOWN')),
    report_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,plan_id) REFERENCES migration_plans(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE case_results (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    case_id TEXT NOT NULL,
    source_observation_id TEXT NOT NULL,
    target_observation_id TEXT NOT NULL,
    verdict TEXT NOT NULL CHECK(verdict IN ('PASS','MISMATCH','UNKNOWN')),
    dimensions_json TEXT NOT NULL CHECK(json_valid(dimensions_json)),
    FOREIGN KEY(project_id,run_id) REFERENCES verification_runs(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,case_id) REFERENCES behavior_cases(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,source_observation_id) REFERENCES observations(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,target_observation_id) REFERENCES observations(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE mismatches (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    case_result_id TEXT NOT NULL,
    dimension TEXT NOT NULL,
    severity TEXT NOT NULL CHECK(severity IN ('CRITICAL','HIGH','MEDIUM','LOW')),
    diff_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,case_result_id) REFERENCES case_results(project_id,id) ON DELETE RESTRICT,
    UNIQUE(project_id,id)
);

CREATE TABLE experiments (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    fingerprint TEXT NOT NULL,
    context_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    outcome TEXT NOT NULL CHECK(outcome IN ('PASSED','FAILED','CANCELLED','UNKNOWN')),
    created_at TEXT NOT NULL,
    UNIQUE(project_id,id)
);

CREATE TABLE repair_attempts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    mismatch_id TEXT NOT NULL,
    experiment_id TEXT NOT NULL,
    attempt_number INTEGER NOT NULL CHECK(attempt_number >= 1),
    patch_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    before_hash TEXT NOT NULL,
    after_hash TEXT,
    outcome TEXT NOT NULL CHECK(outcome IN ('FIXED','REVERTED','FAILED','CONFLICT')),
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,mismatch_id) REFERENCES mismatches(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,experiment_id) REFERENCES experiments(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE missions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    unit_id TEXT,
    case_id TEXT,
    priority INTEGER NOT NULL CHECK(priority >= 0),
    status TEXT NOT NULL CHECK(status IN ('QUEUED','LEASED','RUNNING','COMPLETED','BLOCKED','FAILED','CANCELLED')),
    context_json TEXT NOT NULL CHECK(json_valid(context_json)),
    fingerprint TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK(attempt_count >= 0),
    fencing_token INTEGER NOT NULL DEFAULT 0 CHECK(fencing_token >= 0),
    lease_owner TEXT,
    lease_expires_at TEXT,
    created_at TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY(project_id,unit_id) REFERENCES migration_units(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,case_id) REFERENCES behavior_cases(project_id,id) ON DELETE RESTRICT,
    CHECK((status IN ('LEASED','RUNNING') AND lease_owner IS NOT NULL AND lease_expires_at IS NOT NULL) OR (status NOT IN ('LEASED','RUNNING'))),
    UNIQUE(project_id,id)
);

CREATE TABLE mission_dependencies (
    project_id TEXT NOT NULL,
    mission_id TEXT NOT NULL,
    depends_on TEXT NOT NULL,
    PRIMARY KEY(project_id,mission_id,depends_on),
    CHECK(mission_id <> depends_on),
    FOREIGN KEY(project_id,mission_id) REFERENCES missions(project_id,id) ON DELETE RESTRICT,
    FOREIGN KEY(project_id,depends_on) REFERENCES missions(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE operations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('PREPARED','RUNNING','COMPLETED','FAILED','CANCELLED')),
    captured_hashes_json TEXT NOT NULL CHECK(json_valid(captured_hashes_json)),
    checkpoint_ref TEXT REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE idempotency_requests (
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    request_key TEXT NOT NULL,
    payload_hash TEXT NOT NULL,
    operation_id TEXT NOT NULL REFERENCES operations(id) ON DELETE RESTRICT,
    result_ref TEXT REFERENCES artifacts(id) ON DELETE RESTRICT,
    PRIMARY KEY(project_id,request_key)
);

CREATE TABLE findings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    code TEXT NOT NULL,
    severity TEXT NOT NULL CHECK(severity IN ('CRITICAL','HIGH','MEDIUM','LOW')),
    state TEXT NOT NULL CHECK(state IN ('OPEN','RESOLVED','EXCLUDED')),
    evidence_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    required_action TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE benchmarks (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    target_tree_hash TEXT NOT NULL,
    outcome TEXT NOT NULL CHECK(outcome IN ('NON_REGRESSION','REGRESSION','INCONCLUSIVE','APPROVED_BUDGET')),
    report_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,run_id) REFERENCES verification_runs(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE mappings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    unit_id TEXT NOT NULL,
    source_ref TEXT NOT NULL,
    target_path TEXT NOT NULL,
    target_symbol TEXT,
    target_hash TEXT NOT NULL,
    FOREIGN KEY(project_id,unit_id) REFERENCES migration_units(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE lessons (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    experiment_id TEXT NOT NULL,
    finding TEXT NOT NULL,
    remediation TEXT NOT NULL,
    evidence_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(project_id,experiment_id) REFERENCES experiments(project_id,id) ON DELETE RESTRICT
);

CREATE TABLE skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL
);

CREATE TABLE skill_versions (
    id TEXT PRIMARY KEY,
    skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    version TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('PROPOSED','EXPERIMENTAL','STABLE','DEPRECATED')),
    document_ref TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    UNIQUE(skill_id,version)
);

CREATE TABLE artifact_links (
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    artifact_id TEXT NOT NULL REFERENCES artifacts(id) ON DELETE RESTRICT,
    PRIMARY KEY(project_id,entity_type,entity_id,artifact_id)
);

CREATE INDEX idx_confirmations_subject ON user_confirmations(project_id,gate_type,subject_hash);
CREATE INDEX idx_events_project_sequence ON events(project_id,sequence);
CREATE INDEX idx_missions_schedule ON missions(project_id,status,priority DESC,created_at,id);
CREATE INDEX idx_experiments_fingerprint ON experiments(project_id,fingerprint);
CREATE INDEX idx_findings_open ON findings(project_id,state,severity);
CREATE INDEX idx_units_plan_status ON migration_units(project_id,plan_id,status);

CREATE TRIGGER events_no_update BEFORE UPDATE ON events
BEGIN SELECT RAISE(ABORT, 'events is append-only'); END;
CREATE TRIGGER events_no_delete BEFORE DELETE ON events
BEGIN SELECT RAISE(ABORT, 'events is append-only'); END;
CREATE TRIGGER baselines_no_update BEFORE UPDATE ON baselines
BEGIN SELECT RAISE(ABORT, 'baselines is append-only'); END;
CREATE TRIGGER baselines_no_delete BEFORE DELETE ON baselines
BEGIN SELECT RAISE(ABORT, 'baselines is append-only'); END;
CREATE TRIGGER source_detections_no_update BEFORE UPDATE ON source_detections
BEGIN SELECT RAISE(ABORT, 'source_detections is append-only'); END;
CREATE TRIGGER source_detections_no_delete BEFORE DELETE ON source_detections
BEGIN SELECT RAISE(ABORT, 'source_detections is append-only'); END;
CREATE TRIGGER user_confirmations_no_update BEFORE UPDATE ON user_confirmations
BEGIN SELECT RAISE(ABORT, 'user_confirmations is append-only'); END;
CREATE TRIGGER user_confirmations_no_delete BEFORE DELETE ON user_confirmations
BEGIN SELECT RAISE(ABORT, 'user_confirmations is append-only'); END;
CREATE TRIGGER ir_documents_no_update BEFORE UPDATE ON ir_documents
BEGIN SELECT RAISE(ABORT, 'ir_documents is append-only'); END;
CREATE TRIGGER ir_documents_no_delete BEFORE DELETE ON ir_documents
BEGIN SELECT RAISE(ABORT, 'ir_documents is append-only'); END;
CREATE TRIGGER behavior_cases_no_update BEFORE UPDATE ON behavior_cases
BEGIN SELECT RAISE(ABORT, 'behavior_cases is append-only'); END;
CREATE TRIGGER behavior_cases_no_delete BEFORE DELETE ON behavior_cases
BEGIN SELECT RAISE(ABORT, 'behavior_cases is append-only'); END;
CREATE TRIGGER workflows_no_update BEFORE UPDATE ON workflows
BEGIN SELECT RAISE(ABORT, 'workflows is append-only'); END;
CREATE TRIGGER workflows_no_delete BEFORE DELETE ON workflows
BEGIN SELECT RAISE(ABORT, 'workflows is append-only'); END;
CREATE TRIGGER migration_plans_no_update BEFORE UPDATE ON migration_plans
BEGIN SELECT RAISE(ABORT, 'migration_plans is append-only'); END;
CREATE TRIGGER migration_plans_no_delete BEFORE DELETE ON migration_plans
BEGIN SELECT RAISE(ABORT, 'migration_plans is append-only'); END;
CREATE TRIGGER observations_no_update BEFORE UPDATE ON observations
BEGIN SELECT RAISE(ABORT, 'observations is append-only'); END;
CREATE TRIGGER observations_no_delete BEFORE DELETE ON observations
BEGIN SELECT RAISE(ABORT, 'observations is append-only'); END;
CREATE TRIGGER verification_runs_no_update BEFORE UPDATE ON verification_runs
BEGIN SELECT RAISE(ABORT, 'verification_runs is append-only'); END;
CREATE TRIGGER verification_runs_no_delete BEFORE DELETE ON verification_runs
BEGIN SELECT RAISE(ABORT, 'verification_runs is append-only'); END;
CREATE TRIGGER case_results_no_update BEFORE UPDATE ON case_results
BEGIN SELECT RAISE(ABORT, 'case_results is append-only'); END;
CREATE TRIGGER case_results_no_delete BEFORE DELETE ON case_results
BEGIN SELECT RAISE(ABORT, 'case_results is append-only'); END;
CREATE TRIGGER mismatches_no_update BEFORE UPDATE ON mismatches
BEGIN SELECT RAISE(ABORT, 'mismatches is append-only'); END;
CREATE TRIGGER mismatches_no_delete BEFORE DELETE ON mismatches
BEGIN SELECT RAISE(ABORT, 'mismatches is append-only'); END;
CREATE TRIGGER experiments_no_update BEFORE UPDATE ON experiments
BEGIN SELECT RAISE(ABORT, 'experiments is append-only'); END;
CREATE TRIGGER experiments_no_delete BEFORE DELETE ON experiments
BEGIN SELECT RAISE(ABORT, 'experiments is append-only'); END;
CREATE TRIGGER repair_attempts_no_update BEFORE UPDATE ON repair_attempts
BEGIN SELECT RAISE(ABORT, 'repair_attempts is append-only'); END;
CREATE TRIGGER repair_attempts_no_delete BEFORE DELETE ON repair_attempts
BEGIN SELECT RAISE(ABORT, 'repair_attempts is append-only'); END;
CREATE TRIGGER benchmarks_no_update BEFORE UPDATE ON benchmarks
BEGIN SELECT RAISE(ABORT, 'benchmarks is append-only'); END;
CREATE TRIGGER benchmarks_no_delete BEFORE DELETE ON benchmarks
BEGIN SELECT RAISE(ABORT, 'benchmarks is append-only'); END;
CREATE TRIGGER lessons_no_update BEFORE UPDATE ON lessons
BEGIN SELECT RAISE(ABORT, 'lessons is append-only'); END;
CREATE TRIGGER lessons_no_delete BEFORE DELETE ON lessons
BEGIN SELECT RAISE(ABORT, 'lessons is append-only'); END;

CREATE TRIGGER contracts_locked_no_update BEFORE UPDATE ON contracts
WHEN OLD.status IN ('LOCKED','DEPRECATED')
BEGIN SELECT RAISE(ABORT, 'locked contract is immutable'); END;
CREATE TRIGGER contracts_no_delete BEFORE DELETE ON contracts
BEGIN SELECT RAISE(ABORT, 'contracts cannot be deleted'); END;
