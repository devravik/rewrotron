# SQLITE-SCHEMA.md — Complete Relational Data Definition & Storage Architecture

> **Authority Level:** Mandatory Storage Contract (Critical Foundation Document #6)  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** `internal/storage/sqlite`, Migrations, All Entities  

---

## 1. Storage Architecture Overview

All persistent metadata, analysis facts, behavioral contracts, verification runs, confirmation gates, and audit logs are managed in a local SQLite 3 database with WAL mode enabled.
- **Project Database:** Located at `.rewrotron/project.db` inside each customer workspace.
- **Global Database:** Located at `~/.rewrotron/global.db` for multi-project knowledge and skills.

---

## 2. Table Catalog & Complete DDL

### 2.1 `projects`
- **Purpose:** Tracks the primary metadata for the workspace.
```sql
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    root_path TEXT NOT NULL,
    status TEXT NOT NULL,                 -- e.g. 'INITIALIZED', 'DETECTED', 'SOURCE_CONFIRMED', 'ANALYZED', 'PLAN_CONFIRMED', 'MIGRATING', 'VERIFIED'
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 2.2 `source_detections`
- **Purpose:** Records automated repository inspection results, detected frameworks, runtimes, and confidence scores.
```sql
CREATE TABLE source_detections (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    detected_language TEXT NOT NULL,
    detected_framework TEXT NOT NULL,
    detected_version TEXT,
    runtime_compatibility TEXT,
    application_type TEXT,
    routes_count INTEGER,
    controllers_count INTEGER,
    models_count INTEGER,
    views_count INTEGER,
    database_engine TEXT,
    confidence_json JSON NOT NULL,
    concerns_json JSON NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

### 2.3 `user_confirmations`
- **Purpose:** Authoritative record of explicit human confirmations for Gate 1, Gate 2, and Gate 3.
- **Immutability:** Strictly append-only.
```sql
CREATE TABLE user_confirmations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    gate_type TEXT NOT NULL,              -- 'GATE_1_SOURCE_DETECTION', 'GATE_2_PLAN_CONFIRMED', 'GATE_3_CUTOVER'
    confirmed_by TEXT NOT NULL,           -- Username or client identifier
    configuration_hash TEXT NOT NULL,     -- SHA-256 of inspected repository at confirmation time
    overrides_json JSON,                  -- Explicit user corrections (e.g. {"php": "8.2"})
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX idx_confirmations_gate ON user_confirmations(project_id, gate_type);
```

### 2.4 `target_architectures`
- **Purpose:** Confirmed target specification decoupling language, framework, and architectural style.
```sql
CREATE TABLE target_architectures (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    target_language TEXT NOT NULL,        -- 'php', 'go', 'rust', 'python', 'typescript'
    target_framework TEXT NOT NULL,       -- 'laravel', 'std_http', 'fiber', 'axum', 'fastapi'
    architecture_style TEXT NOT NULL,     -- 'modular_monolith', 'clean_architecture', 'hexagonal'
    database_policy TEXT NOT NULL DEFAULT 'preserve_existing',
    api_compatibility TEXT NOT NULL DEFAULT 'strict',
    config_json JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

### 2.5 `events`
- **Purpose:** Append-only event stream recording every pipeline transition, tool call, and verdict.
- **Immutability:** Strictly append-only; NEVER updated or deleted.
```sql
CREATE TABLE events (
    event_id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,            -- e.g. 'GATE_CONFIRMED', 'UNIT_MIGRATED', 'MISMATCH_DETECTED'
    payload_json JSON NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX idx_events_type_time ON events(event_type, timestamp);
```

### 2.6 `source_files`
- **Purpose:** Inventory of all files in the legacy source codebase with content hashes.
```sql
CREATE TABLE source_files (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    file_type TEXT NOT NULL,             -- 'controller', 'model', 'view', 'helper', 'config', 'asset'
    sha256 TEXT NOT NULL,
    loc INTEGER NOT NULL,
    size_bytes INTEGER NOT NULL,
    classification TEXT NOT NULL DEFAULT 'UNKNOWN', -- 'ACTIVE', 'DORMANT', 'DEAD', 'UNKNOWN'
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, relative_path),
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX idx_source_files_class ON source_files(classification);
```

### 2.7 `source_symbols`
- **Purpose:** Indexed classes, methods, functions, and properties discovered by AST parsing (Source IR).
```sql
CREATE TABLE source_symbols (
    id TEXT PRIMARY KEY,
    file_id TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,                  -- 'class', 'method', 'function', 'property', 'constant'
    parent_symbol TEXT,                  -- Enclosing class name
    visibility TEXT,                     -- 'public', 'protected', 'private'
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    classification TEXT NOT NULL DEFAULT 'UNKNOWN',
    FOREIGN KEY(file_id) REFERENCES source_files(id) ON DELETE CASCADE
);
CREATE INDEX idx_symbols_name ON source_symbols(name);
CREATE INDEX idx_symbols_file ON source_symbols(file_id);
```

### 2.8 `routes`
- **Purpose:** Discovered legacy application routes and entrypoints.
```sql
CREATE TABLE routes (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    http_method TEXT NOT NULL,           -- 'GET', 'POST', 'PUT', 'DELETE', 'ANY'
    path_pattern TEXT NOT NULL,          -- e.g. '/orders/(:num)'
    controller TEXT NOT NULL,
    action TEXT NOT NULL,
    source_type TEXT NOT NULL,           -- 'CONFIG', 'CONVENTION', 'OBSERVED'
    classification TEXT NOT NULL DEFAULT 'ACTIVE',
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_routes_unique ON routes(project_id, http_method, path_pattern);
```

### 2.9 `workflows`
- **Purpose:** Multi-step business workflows spanning multiple routes (Application IR).
```sql
CREATE TABLE workflows (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    initial_step TEXT NOT NULL,
    steps_json JSON NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

### 2.10 `contracts`
- **Purpose:** Canonical framework-independent behavioral contracts (Behavior IR).
```sql
CREATE TABLE contracts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    route_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    schema_yaml TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT', -- 'DRAFT', 'LOCKED', 'DEPRECATED'
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(route_id) REFERENCES routes(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_contracts_route_ver ON contracts(route_id, version);
```

### 2.11 `behavior_cases`
- **Purpose:** Concrete executable test scenarios under a contract.
```sql
CREATE TABLE behavior_cases (
    id TEXT PRIMARY KEY,
    contract_id TEXT NOT NULL,
    name TEXT NOT NULL,
    http_method TEXT NOT NULL,
    path TEXT NOT NULL,
    headers_json JSON NOT NULL,
    payload_json JSON,
    preconditions_json JSON NOT NULL,
    expected_status INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(contract_id) REFERENCES contracts(id) ON DELETE CASCADE
);
CREATE INDEX idx_cases_contract ON behavior_cases(contract_id);
```

### 2.12 `observations`
- **Purpose:** Recorded runtime execution traces from either legacy or target environments.
```sql
CREATE TABLE observations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    trace_id TEXT NOT NULL UNIQUE,
    case_id TEXT,
    environment TEXT NOT NULL,           -- 'source_legacy' or 'target_modern'
    status_code INTEGER NOT NULL,
    duration_ms REAL NOT NULL,
    trace_artifact_path TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(case_id) REFERENCES behavior_cases(id) ON DELETE SET NULL
);
CREATE INDEX idx_obs_case_env ON observations(case_id, environment);
```

### 2.13 `migration_units`
- **Purpose:** Dependency-sorted vertical slices of migration.
```sql
CREATE TABLE migration_units (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    dependencies_json JSON NOT NULL,     -- Array of prerequisite unit IDs
    status TEXT NOT NULL DEFAULT 'PLANNED', -- 'PLANNED', 'GENERATED', 'VERIFIED', 'COMPLETED'
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX idx_units_status_priority ON migration_units(status, priority);
```

### 2.14 `mappings`
- **Purpose:** Traceability links between legacy source symbols and generated target files/classes.
```sql
CREATE TABLE mappings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    unit_id TEXT NOT NULL,
    legacy_symbol TEXT NOT NULL,
    target_symbol TEXT NOT NULL,
    target_file TEXT NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(unit_id) REFERENCES migration_units(id) ON DELETE CASCADE
);
CREATE INDEX idx_mappings_unit ON mappings(unit_id);
```

### 2.15 `mismatches`
- **Purpose:** Discrepancies detected during differential verification runs.
```sql
CREATE TABLE mismatches (
    id TEXT PRIMARY KEY,
    case_id TEXT NOT NULL,
    unit_id TEXT NOT NULL,
    mismatch_type TEXT NOT NULL,         -- 'STATUS_CODE', 'HEADER', 'BODY', 'DATABASE', 'SESSION'
    severity TEXT NOT NULL,              -- 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW'
    classification TEXT NOT NULL DEFAULT 'DEFECT', -- 'DEFECT', 'EXPECTED_CHANGE', 'UNKNOWN'
    diff_artifact_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'OPEN', -- 'OPEN', 'RESOLVED', 'ACCEPTED'
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(case_id) REFERENCES behavior_cases(id) ON DELETE CASCADE,
    FOREIGN KEY(unit_id) REFERENCES migration_units(id) ON DELETE CASCADE
);
CREATE INDEX idx_mismatches_status ON mismatches(status);
```

### 2.16 `repair_attempts`
- **Purpose:** Record of iterative repair cycles executed by AI agents or algorithms.
```sql
CREATE TABLE repair_attempts (
    id TEXT PRIMARY KEY,
    mismatch_id TEXT NOT NULL,
    attempt_number INTEGER NOT NULL,
    patch_artifact_path TEXT NOT NULL,
    outcome TEXT NOT NULL,               -- 'FIXED', 'REVERTED', 'FAILED'
    diagnostic_lesson TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(mismatch_id) REFERENCES mismatches(id) ON DELETE CASCADE
);
```

### 2.17 `missions`
- **Purpose:** Autonomous work scheduler queue for agents.
```sql
CREATE TABLE missions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    type TEXT NOT NULL,                  -- 'ANALYZE_SOURCE', 'MIGRATE_UNIT', 'VERIFY_CASE', 'REPAIR_MISMATCH'
    unit_id TEXT,
    case_id TEXT,
    priority INTEGER NOT NULL DEFAULT 50,
    status TEXT NOT NULL DEFAULT 'QUEUED', -- 'QUEUED', 'IN_PROGRESS', 'COMPLETED', 'BLOCKED'
    attempt_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(unit_id) REFERENCES migration_units(id) ON DELETE SET NULL
);
CREATE INDEX idx_missions_sched ON missions(status, priority DESC);
```

### 2.18 `benchmarks`, `lessons`, `skills`, `skill_versions`, `artifacts`
```sql
CREATE TABLE benchmarks (
    id TEXT PRIMARY KEY,
    unit_id TEXT NOT NULL,
    concurrency INTEGER NOT NULL,
    legacy_p95_ms REAL NOT NULL,
    target_p95_ms REAL NOT NULL,
    legacy_mem_mb REAL NOT NULL,
    target_mem_mb REAL NOT NULL,
    passed_rule INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(unit_id) REFERENCES migration_units(id) ON DELETE CASCADE
);

CREATE TABLE lessons (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    category TEXT NOT NULL,
    finding TEXT NOT NULL,
    remediation TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL
);

CREATE TABLE skill_versions (
    id TEXT PRIMARY KEY,
    skill_id TEXT NOT NULL,
    version TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'EXPERIMENTAL',
    FOREIGN KEY(skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

CREATE TABLE artifacts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    relative_path TEXT NOT NULL UNIQUE,
    file_type TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```
