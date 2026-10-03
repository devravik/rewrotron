# Append-only events and export

Events have global monotonic `sequence`, unique `event_id`, `project_id`, `event_type`, entity identity, request ID, UTC timestamp and validated JSON payload. SQLite sequence defines order; wall clocks can move backward. Event IDs are identities, not chronological proof.

The entity update, expected revision check and event insertion share one transaction. Append-only SQL triggers forbid mutation/deletion. Payloads contain hashes/references and sanitized structured facts, not raw secrets or huge trace bodies. Include schema version in payloads used for projection rebuilding.

| Event | Required facts |
| --- | --- |
| PROJECT_INITIALIZED / DETECTION_RECORDED | Project/source identity, manifest and detection references |
| GATE_CONFIRMED / BASELINE_INVALIDATED | Gate, exact subject, actor/channel/action or drift reason |
| ANALYSIS_COMPLETED / OBSERVATION_RECORDED / CONTRACT_LOCKED | Captured baseline, artifact refs, completeness/counts |
| PLAN_SYNTHESIZED / CODE_GENERATED | Plan/target hashes, units, patch and provenance refs |
| VERIFICATION_COMPLETED / MISMATCH_DETECTED | Run, case, dimensions, verdict, diff refs |
| MISSION_CLAIMED / MISSION_RENEWED / MISSION_COMPLETED | Mission ID, owner, fence, expiry, evidence |
| REPAIR_ATTEMPTED / BENCHMARK_COMPLETED | Fingerprint, outcome and evidence hashes |
| LESSON_RECORDED / EXPORT_APPROVED | Scope, reviewed privacy manifest and evidence |
| OPERATION_FAILED / OPERATION_CANCELLED | Error category/code, retryability, sanitized diagnostics |

A JSONL export is generated after commit using sequence cursors. It can be replayed/deduplicated by event ID and rebuilt from SQLite; it is not a second authoritative ledger. Export failure never rolls back committed business state. Read-only tool calls may append audit events but must not change project revision or lifecycle.

Only documented event-driven projections are promised rebuildable. Artifacts and immutable baseline/contract/plan records are needed too; an event log by itself does not restore a filesystem, DB snapshot or external side effect.

## Canonical execution payloads and views

[EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) owns typed execution payloads in the existing ledger; `specs/execution/event.schema.json` validates the committed public representation. Payloads name schema version, operation/run, stage instance/task/mission, OBSERVED/DECIDED/ACTION/RESULT kind, factual reason, result, refs, progress and user attention. Starts, progress, heartbeat, discovery, decision, action, warnings/errors, approvals, controls and invalidations share ledger ordering and ownership. Existing domain event types remain valid; no second event table or competing log is introduced. Domain events may carry the same execution context; append correlated presentation events in the same transaction where needed and deduplicate their domain fact in projections.

Heartbeats do not claim progress or change lifecycle/business revision. Completion/control/approval events and state changes share a transaction; received approval requires the actual confirmation row. Current projection uses state_sequence N; event window uses through_sequence M ≤ N and exclusive cursors, handle global per-project gaps and delivery duplicates, and expose per-client deltas without server-wide read acknowledgment. Sanitize before persistence; failure of the sink blocks new opaque work at a safe boundary without replaying external effects. JSONL remains a rebuildable export, not live authority.

Current state cannot be reconstructed as-of an earlier event sequence from mutable rows. Historical event pages are supported; historical state projection is deferred. Inspection auditing is coalesced under EXECUTION-OBSERVABILITY; retention/quota admission follows ARTIFACT-STORE.
