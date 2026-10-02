# COMPONENTS.md — Subsystem Catalog & Responsibilities

> **Version:** 1.0  
> **Status:** Architecture Reference Guide  

---

This catalog details the internal packages, their exact Go module paths, responsibilities, and interface boundaries.

## 1. Subsystem Catalog

### 1.1 Command Interfaces
- **`cmd/rewrotron`**: Primary binary entrypoint. CLI parser (Cobra/pflag based), terminal UI formatting, exit code handling, and machine-readable JSON envelope generation.
- **`cmd/rewrotron-mcp`**: Dedicated Model Context Protocol server binary. Exposes JSON-RPC endpoints for MCP clients over stdio or SSE.

### 1.2 Core Domain Subsystems (`internal/`)
- **`internal/project`**: Handles initialization of `.rewrotron`, parsing `project.yaml`, workspace locking (`.rewrotron/lock`), and status aggregation.
- **`internal/analyzer`**:
  - `analyzer/php`: PHP lexing and AST parsing (using robust Go-based tree-sitter or PHP parser bindings).
  - `analyzer/routes`: Extracts static routes, URI patterns, HTTP verbs, and controller mappings.
  - `analyzer/symbols`: Class, interface, method, helper function, and global symbol indexing.
  - `analyzer/dependencies`: Builds call graphs, class inheritance trees, and include/require hierarchies.
  - `analyzer/reachability`: Calculates transitive reachability from defined entrypoints.
- **`internal/observer`**:
  - `observer/http`: Intercepts and logs HTTP request/response pairs with microsecond timing.
  - `observer/ci3`: Framework-specific hooks tapping into CodeIgniter 3's benchmark and hook system (`pre_controller`, `post_controller`).
  - `observer/database`: Captures SQL query logs, execution durations, bound parameters, and affected rows.
  - `observer/session`: Records session read/write mutations and flash messages across lifecycle hooks.
- **`internal/behavior`**:
  - `behavior/graph`: Represents the unified graph linking Entrypoints -> Workflows -> Operations -> StateTransitions -> Responses.
  - `behavior/contracts`: Validates and materializes `BehaviorContract` records.
  - `behavior/normalization`: Contains normalization engines for dynamic tokens (timestamps, UUIDs, CSRF nonces).
- **`internal/migration`**:
  - `migration/planner`: Topologically sorts migration units based on shared models and dependency trees.
  - `migration/generator`: Orchestrates target code emission via target adapters.
  - `migration/mappings`: Tracks legacy symbol/route to modern class/route mappings.
- **`internal/verification`**:
  - `verification/replay`: Orchestrates dual HTTP execution against legacy and target test harnesses.
  - `verification/comparator`: Multi-layered comparator (Byte-Exact, Semantic JSON/HTML, Dynamic Normalized).
  - `verification/database`: Computes diffs between legacy and target DB mutation transaction logs.
  - `verification/diagnostics`: Pinpoints root cause of mismatches for LLM repair loops.
- **`internal/storage`**:
  - `storage/sqlite`: SQLite repository implementations for all persistent entities.
  - `storage/artifacts`: Content-addressable storage for large JSON payloads, traces, and diffs.
- **`internal/learning`**:
  - `learning/lessons`: Captures atomic diagnostic lessons discovered during repair loops.
  - `learning/patterns`: Evaluates lessons against global criteria for promotion into global patterns.

### 1.3 Adapters (`adapters/`)
- **`adapters/source/codeigniter3`**: Parses CI3 configs (`application/config/*.php`), controller inheritance (`CI_Controller`), models (`CI_Model`), and active record queries.
- **`adapters/target/laravel13`**: Emits Laravel 13 code structure (routes in `routes/`, controllers in `app/Http/Controllers/`, models in `app/Models/`, migrations in `database/migrations/`).
