# Mission scheduling, leases and completion

## States and readiness

Mission states are `QUEUED`, `LEASED`, `RUNNING`, `COMPLETED`, `BLOCKED`, `FAILED`, `CANCELLED`. Types are listed in `specs/missions/mission.schema.json`. A mission is ready only when its dependencies are complete, applicable gates are valid, source/context hashes are current and its required capability/environment exists.

`next` / `research.get_next_mission` peeks and returns `{mission: null, reason: NO_WORK|DEPENDENCIES_BLOCKED|CONFIRMATION_REQUIRED}` when appropriate. It never claims. Ready ordering is priority descending, creation time ascending, mission ID ascending. Severity sets deterministic priorities (critical repair 1000, high repair 900, unknown review 800, observation gap 700, generation 500, benchmark 300, optimization 100); dependencies still dominate priority. No history-based opaque scoring is needed initially.

## Lease protocol

Claim under `BEGIN IMMEDIATE`: verify current readiness, select/check the queued or expired mission, increment `fencing_token` and `attempt_count`, set owner and expiry, and append MISSION_CLAIMED. Default lease duration is five minutes; heartbeat renewal every minute. Owner derives from the session. Renew and publish require exact owner/fence and a live lease; expired workers cannot publish even if they finish work later.

Expired leases become eligible for recovery after checkpoints and target journals are reconciled. Automatic retries have at most three attempts per incident subject; trivial patch/context edits cannot reset the budget. Known failed fingerprints require a new hypothesis/context or a concrete review blocker. Execution uses at-least-once scheduling; externally visible artifact/target publication is made idempotent. Do not promise exactly-once external effects.

## Completion evidence

| Type | Required evidence |
| --- | --- |
| ANALYZE_SOURCE / DISCOVER_ROUTE / ANALYZE_DEAD_CODE | Current IR/graph/completeness and classification references |
| OBSERVE_WORKFLOW | Captured trace(s), restored fixture and required dimensions |
| GENERATE_CONTRACT | Valid contract/cases, provenance, explicit uncertainties |
| MIGRATE_UNIT | Owned patch/provenance, target build and current UNIT plus applicable implemented-regression PASS evidence |
| VERIFY_CASE | Current immutable run, case and dimension verdicts; UNKNOWN cannot close as success |
| REPAIR_MISMATCH | Before/after hashes, hypothesis, tests/run evidence and regression success |
| BENCHMARK / OPTIMIZE | Equivalent corpus, reproducible measurement and acceptance report |
| REVIEW_UNKNOWN | Explicit decision/evidence resolving the question or reviewed scope exclusion; no invented PASS |

Completion checks fence, gates, current hashes and evidence in one transaction. A free-text agent summary cannot complete a verification/generation mission. Blocking records include code, evidence, affected units, required input and next action so resume can explain the dependency.

## Execution visibility and pause readiness

Mission claim/start/action/result/completion feed the canonical [execution projection](EXECUTION-OBSERVABILITY.md). Current task reasons, captured subjects, evidence, blockers and next actions must remain inspectable throughout long work; completion is published only with validated evidence. Execution heartbeat every 15 seconds is separate from the one-minute lease renewal. An alive worker has not necessarily made progress.

Human PAUSE_REQUESTED/PAUSED is an additional readiness guard for claims/new actions. Reach safe boundaries, preserve checkpoint/journal evidence and release leases coherently before acknowledging pause. Existing mission status enums remain unchanged; interrupted work records its actual transition/attempt and resume context. Restart/resume preserves pause; explicit continue rechecks gates/dependencies/drift. Mission peek returns the existing `DEPENDENCIES_BLOCKED` reason while human control prevents admission; the canonical execution state supplies the pause reason/detail. Do not add undeclared fields to the peek result or invent a ready lease.

## Work identity and resources

Fingerprint hashes project, mission type, subject, unit/case and requested capability/context; active identical work is unique by project/type/fingerprint in migration 0002. Claim atomically acquires mission lease plus project mutation/resource claim; a unique mission ID alone is insufficient. Target model component owns shared paths; initial project mutation/publication is serial. Disjoint parallel mutations and per-resource CAS are deferred until race/starvation experiments pass. OS/supervisor ownership spans fence-check and filesystem write; lease expiry alone cannot authorize a new writer while the old process remains unreconciled. Completion validates scope identity, current subjects and ownership.

Assessment evidence work uses the existing analysis/observation mission types, scopes and gates. Publishing an assessment-only report never queues MIGRATE_UNIT or patch/generation work. PLAN readiness additionally requires the current explicit human continuation selection; implementation readiness still requires Gate 2. A report with blocked feasibility remains inspectable even when no implementation mission can be admitted.
