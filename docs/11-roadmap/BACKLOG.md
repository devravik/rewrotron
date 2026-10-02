# BACKLOG.md — Sprint Backlog & Priority Epics

> **Version:** 1.0  
> **Status:** Implementation Backlog  

---

## Sprint 1 Backlog: Foundation & Core Scaffolding

- [ ] **TASK-001:** Setup Go monorepo structure (`cmd/rewrotron`, `cmd/rewrotron-mcp`, `internal/`).
- [ ] **TASK-002:** Implement SQLite storage layer (`internal/storage/sqlite`) with schema migration runner for `SQLITE-SCHEMA.md`.
- [ ] **TASK-003:** Implement `rewrotron init` command creating `.rewrotron/`, `project.db`, and `project.yaml`.
- [ ] **TASK-004:** Implement `rewrotron doctor` checking local Go, PHP, Composer, and Docker runtimes.
- [ ] **TASK-005:** Implement Go-based PHP AST parser and CodeIgniter 3 route configuration extractor in `internal/analyzer`.
- [ ] **TASK-006:** Implement `internal/observer` HTTP interception proxy and CI3 hook generator.
- [ ] **TASK-007:** Implement Canonical Behavior Model data structures and JSON Schema validator in `internal/behavior`.
- [ ] **TASK-008:** Implement initial Differential Verification runner in `internal/verification`.
- [ ] **TASK-009:** Implement MCP server skeleton in `cmd/rewrotron-mcp` exposing `project.get_snapshot`.
