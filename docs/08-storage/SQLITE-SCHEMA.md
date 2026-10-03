# SQLite storage contract

**Status:** Initial implementation specification. The earlier DDL was an unimplemented draft. [0001_initial.sql](../../migrations/0001_initial.sql) is the initial executable project migration; there is no deployed database predecessor to upgrade. Future persistent changes require new numbered migrations and compatibility fixtures.

## Database ownership and connection policy

Project state lives in `.rewrotron/project.db`. The first release uses `database/sql` with `modernc.org/sqlite`. Set `PRAGMA foreign_keys=ON`, `journal_mode=WAL`, `synchronous=FULL`, and `busy_timeout=5000` on every relevant connection (WAL initialization is coordinated once per database). Use bounded busy retries and respect cancellation; never replay an external effect because a transaction was busy.

The runner owns `schema_migrations(version INTEGER PRIMARY KEY, checksum TEXT NOT NULL, applied_at TEXT NOT NULL)`. Apply migrations in numeric order inside transactions, verify previously applied checksums, reject a DB version newer than the binary, and back up before upgrading. WAL mode changes occur outside a migration transaction. File permissions restrict access to the local user.

Timestamp columns store UTC RFC3339 strings supplied by the injected clock. Event order uses `events.sequence`, not wall-clock time or random IDs. Application identifiers are strings; hashes are lowercase SHA-256. JSON columns are TEXT with `json_valid` checks; application schema validation still enforces their domain structure.

## Record catalog

| Group | Tables | Contract |
| --- | --- | --- |
| Workspace and approvals | projects, source_detections, baselines, approval_tokens, user_confirmations | YAML holds intent; immutable reviewed subjects grant authority |
| Source and IR | source_files, source_symbols, ir_documents, entrypoints | Baseline-scoped syntax/structure/evidence; HTTP is one entrypoint type |
| Behavior | contracts, contract_entrypoints, behavior_cases, workflows | Contract revisions and cases are not one-to-one with routes |
| Plan and target | migration_plans, migration_units, unit_cases, mappings | Immutable plans, mutable progress projections, explicit case coverage |
| Verification and repair | observations, verification_runs, case_results, mismatches, experiments, repair_attempts, benchmarks | Immutable evidence/verdicts; intentional changes require new runs |
| Scheduling/recovery | missions, mission_dependencies, operations, idempotency_requests, findings | Fenced leases, checkpoints, captured hashes, concrete blockers |
| Evidence and learning | artifacts, artifact_links, lessons, skills, skill_versions, events | Registered content-addressed bytes, append-only audit, project-local lessons |

All child records preserve project ownership. Composite foreign keys enforce same-project references where available; repository validation also verifies artifact ownership, baseline consistency, ID references and current subject relations. An artifact reference string resolves to a registered artifact ID/digest before inserting metadata. No `ON DELETE CASCADE` can silently erase audit/confirmation history.

## Immutability and revisions

Triggers reject updates/deletes of confirmations, baselines, events, detections, IR documents, plans, cases, observations, verification results, mismatches and completed experiment evidence. Locked contracts cannot change; new revisions have new row IDs. Mismatch resolution/classification is an append-only event and new evidence, not editing the original discrepancy or verdict. Mutable project/unit/mission/finding rows are projections updated transactionally with events.

`approval_tokens` stores only token digests and may update `consumed_at`. A consumed token and confirmation/event commit atomically. Gate validity is derived from exact hash subjects and parents, not existence of any past confirmation. See [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md).

## Transactions and constraints

State changes use expected project revision and short write transactions. Long analysis/generation/replay occurs outside DB locks and publishes only after current-hash/fence checks. Mission claiming uses `BEGIN IMMEDIATE` plus lease/fence compare-and-swap. Repository APIs bind SQL parameters and expose domain methods; no raw SQL is public.

DDL checks enforce nonnegative sizes/attempt counts, valid lifecycle/verdict enums, symbol span order, JSON validity and mission lease field presence. Semantic reference validation, DAG cycle checks, hash recomputation, append-only artifact links and lease expiry are service/repository responsibilities with explicit negative fixtures.

## Global data and compatibility

Global knowledge is optional and deferred to RWT-023. `~/.rewrotron/global.db` must have a separate numbered migration namespace and contain only privacy-reviewed promoted patterns, references and skill versions, never raw customer traces. An absent global database cannot block any baseline operation.

Storage tests must cover fresh creation, schema version 0→1, reopening, WAL/foreign key behavior, invalid references/enums/JSON, immutability, rollback on migration failure, checksum/newer-version rejection, busy contention, backup/restore and actual previous-version upgrades as migrations are added. Executing initial DDL with Python SQLite validates SQL structure; it does not replace Go driver/recovery/race tests.

## Target and dependency decision artifacts

The decision system in [ADR 0012](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) uses existing `artifacts`/`artifact_links` for immutable target revisions, source/target graphs, recommendations, API differences and feasibility. Plans embed subject hashes and references; `events` records revision/decision/invalidation provenance and `findings` projects current blockers. Exact approved target/dependency decisions are bound by the Gate 2 plan and `user_confirmations`; YAML exports cannot update that authority.

Artifact links and service validation preserve project/input subject ownership and transitive impact references. A target change updates the current plan/progress projection without deleting baseline or source evidence. This design adds no tables or columns and changes no initial DDL; future indexing/normalized tables require incremental migrations and compatibility fixtures.

## Execution observability storage

[ADR 0013](../01-architecture/ADR/0013-transparent-execution.md) uses existing event payload JSON for typed starts/results/progress/heartbeat/explanations and persistent human controls. Existing operations/missions/findings/confirmations and registered artifacts provide domain truth; `internal/execution` rebuilds documented stage/control/timeline projections from those records at a sequence watermark. Project lifecycle and operation status columns are not extended with UI stage/pause labels. Client last-seen watermarks remain client-owned; inspection does not mark work or notifications approved.

Execution run IDs identify operation attempts; evidence run IDs stay separate. Same-project ownership, valid fences/subjects, progress arithmetic, human control provenance and projection consistency require service tests. This design changes no tables/columns/initial DDL. Future indexed projection tables or server-side notification receipts require incremental migrations and compatibility fixtures.

## Executable DDL

The complete SQL below is mirrored from the initial migration. Keep it byte-consistent when revising the undeployed baseline; after deployment add incremental migrations and document the resulting schema.

```sql
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

```
