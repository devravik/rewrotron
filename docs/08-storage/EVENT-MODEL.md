# EVENT-MODEL.md — Append-Only Audit Stream & Event Taxonomy

> **Version:** 1.0  
> **Status:** Storage Reference Guide  

---

## 1. Event Invariants

1. **Strictly Append-Only:** Rows in `events` are never updated, deleted, or truncated during normal operations.
2. **Deterministic Sequence:** Events carry monotonic timestamps and ULID/UUID identifiers.
3. **Dual Persistence:** Events are written simultaneously to `project.db` (for relational querying) and `.rewrotron/logs/events.jsonl` (for external log forwarders).

---

## 2. Event Types Catalog

| Event Type | Emitting Subsystem | Payload Contents |
| :--- | :--- | :--- |
| `PROJECT_INITIALIZED` | `internal/project` | Source path, target framework, config flags |
| `ANALYSIS_COMPLETED` | `internal/analyzer` | Symbol counts, files scanned, route count |
| `OBSERVATION_RECORDED` | `internal/observer` | Trace ID, endpoint, duration, query count |
| `CONTRACT_LOCKED` | `internal/behavior` | Contract ID, version, hash |
| `UNIT_PLANNED` | `internal/migration` | Unit ID, dependency list, priority |
| `CODE_GENERATED` | Target Adapter | Target file paths generated, LOC |
| `VERIFICATION_PASSED` | `internal/verification` | Case ID, Unit ID, duration |
| `MISMATCH_DETECTED` | `internal/verification` | Mismatch type, severity, diff summary |
| `REPAIR_ATTEMPTED` | Repair Loop | Mismatch ID, attempt number, outcome |
| `GATE_APPROVED` | Human Review | User identity, rationale, approved action |
