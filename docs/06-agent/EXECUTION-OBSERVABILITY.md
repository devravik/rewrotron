# Transparent execution, live state and user supervision

**Status:** Normative core product requirement; implementation pending. Rewrotron has two central promises: preserve the application's observable behavior, and always show what it is doing and why. Transparency is required from the first long-running operation, not deferred to a desktop UI. [ADR 0013](../01-architecture/ADR/0013-transparent-execution.md) records this decision.

## User questions and canonical answers

Every long-running operation MUST expose its current stage, task, reason, discoveries, changes, result, blockers, next action and requested user input through the canonical execution state. An operation expected to outlive one status refresh, including scans, sandbox startup, dependency resolution, generation, verification, repair and optimization, is long-running. No opaque multi-hour “migration in progress” state is acceptable.

| Question | Required answer |
| --- | --- |
| What is happening? | Active execution runs, stage instances, tasks/actions and last activity |
| Why? | Concise factual rationale, goal, source/contract references and unmet prerequisite |
| What was discovered? | Observed facts/counts, evidence, uncertainty and affected scope |
| What changed? | Committed artifact/target changes, prior/current subjects and invalidations |
| What is blocked? | Concrete issue, affected/dependent units, evidence, alternatives and required decision |
| What happens next? | Next ready mission/action, or the prerequisite preventing it; proposed work is labeled |
| What is needed from me? | Exact decision/approval subject, options, impact, owner and authorized action |

`internal/execution` owns stage projection, structured event validation, progress accounting, explanations and notification views. It uses existing SQLite events, operations, missions, findings, artifacts and confirmations. Suggested stages/events/progress/explanations modules are internal subdivisions, not mandatory duplicate packages. Approval presentation delegates to existing gate services; this subsystem cannot mint approval or make a verification verdict. All producers use injected ports, clock and redactor; the deterministic engine works without an agent or LLM.

## Visible stages versus durable lifecycle

Expose these stages in this display order. They are execution labels, not new project lifecycle states or a promise of eleven sequential passes.

| Stage | Work shown |
| --- | --- |
| `DETECT` | Source/environment metadata, evidence/confidence and unknown facts |
| `CONFIRM` | Review/decision requests, including the three existing gates |
| `ANALYZE` | Syntax/structure, entrypoints, reachability, dependency usage |
| `OBSERVE` | Sandbox preparation, supplied test identities, requests/state/effect capture |
| `MODEL` | Evidenced workflows, contracts, cases and unresolved behavior |
| `PLAN` | Frozen user target, feasibility, recursive resolutions, units and omissions |
| `MIGRATE` | Target staging, builds, ownership and publication |
| `VERIFY` | Differential cases/dimensions and current aggregate evidence |
| `REPAIR` | Recorded hypothesis, bounded attempts and regression results |
| `OPTIMIZE` | Optional post-parity improvements with benchmark/regression evidence |
| `FINALIZE` | Full-scope checks, benchmark, session/rollback evidence, Gate 3 and packaging |

Every visit has a `stage_instance_id` and captured subjects. OBSERVE/MODEL can repeat; independent slices can be MIGRATING and VERIFYING concurrently; CONFIRM may recur at each gate. OPTIMIZE can be `NOT_APPLICABLE` when not selected. A completed stage instance does not mean every case passed or that a whole project is complete. Show latest/current/historical instances and reasons for reopened work; do not turn prior success into a current checkmark after drift.

Stage/task display statuses are `PENDING`, `RUNNING`, `WAITING_USER`, `BLOCKED`, `PAUSED`, `COMPLETED`, `FAILED`, `CANCELLED`, `STALE`, `NOT_APPLICABLE`. They do not replace project, unit, mission, operation or verifier enums. Mission completion still requires engine-validated evidence. A finished verification action may have outcome MISMATCH or UNKNOWN; task outcome and verdict must remain visible even if execution finished successfully.

## One structured event model

An `ExecutionEvent` is a typed payload in the existing append-only event ledger, not another event store. [event.schema.json](../../specs/execution/event.schema.json) owns its canonical structure. The committed public event combines ledger identity/order with payload fields:

- `schema_version`, `event_id`, `sequence`, `project_id`, execution `run_id`, `operation_id`, optional `mission_id` and `task_id`.
- `stage`, `stage_instance_id`, `type`, `kind`, `severity`, `requires_user`, UTC `created_at` and causation event ID.
- `title`, plain-English `summary`, factual `reason`, labeled `result`, next action and requested input.
- Entity identity, captured subject hashes, evidence/artifact refs, structured details and optional measured progress.

An execution run identifies an operation attempt; `run_id` equals its durable operation ID in the first implementation. Verifier case/run IDs are referenced entities, not silently substituted for execution run IDs. Retries receive new operation/run IDs and preserve causation/checkpoint links. Concurrent runs are all shown; a focus run is a display choice, not an assertion that no other work exists.

| Event type | Required meaning |
| --- | --- |
| `STAGE_STARTED`, `STAGE_COMPLETED` | One stage instance started/finished with scope and outcome |
| `TASK_STARTED`, `TASK_COMPLETED`, `TASK_BLOCKED` | Task reason/readiness, actual result or blocker |
| `DISCOVERY_FOUND` | Evidence-backed observation; candidate/inferred/unknown facts labeled |
| `DECISION_MADE` | Named choice, alternatives, evidence, decision authority and target/plan subject |
| `ACTION_STARTED`, `ACTION_COMPLETED` | Controlled action and observed result, including failures |
| `WARNING`, `ERROR` | Sanitized issue, impact, retry policy and remediation |
| `APPROVAL_REQUIRED`, `APPROVAL_RECEIVED` | Exact gate/decision subject; received only after authoritative confirmation commits |
| `MILESTONE_REACHED` | Criteria/evidence satisfied for the named scope |
| `PROGRESS_UPDATED`, `HEARTBEAT` | Measured work or worker liveness; neither implies success |
| `PAUSE_REQUESTED`, `EXECUTION_PAUSED`, `EXECUTION_CONTINUED` | Durable human control, acknowledged safe boundary and fresh readiness checks |
| `EXECUTION_CANCELLED`, `EXECUTION_RECOVERED`, `ARTIFACT_INVALIDATED` | Termination/reconciliation or affected stale evidence |

`kind` is `OBSERVED`, `DECIDED`, `ACTION`, or `RESULT`. It separates observed facts from intended decisions, attempted actions and actual outcomes. `severity` is `INFO`, `WARNING`, `ERROR`, or `CRITICAL`; recommendation REQUIRED/OPTIONAL/BLOCKING types and finding risk enums remain distinct. A DECIDED event from an agent is a proposal unless an engine policy or persisted human confirmation authorizes it; the event must name that authority. It cannot make a recommendation approved by changing a display label.

## Explanations and drill-down

Every task/action has a concise reason derived from the recorded goal, observed facts, contract gaps or prerequisite. For example: “Test POST /orders with a second supplied identity because it changes a user-owned resource and cross-account authorization is not characterized.” Link the case, ownership evidence and affected unit. Source observations, proposed tests and final assertions remain separate. Never claim an account was created unless that controlled fixture action actually occurred.

Explanations are factual justifications, not private chain-of-thought, prompt logs or speculative narratives. Hypotheses are labeled and link to experiments; “likely cause” never becomes a proven diagnosis without evidence. “Why this?” resolves to the same explanation record in every client. Missing explanation is an instrumentation defect, not permission to fabricate one.

Each stage/entity offers Summary, Evidence, Decisions, Artifacts, Warnings and Open questions. Drill-down uses registered entity/artifact refs and captured hashes, with privacy/ownership checks. A behavioral slice lists its routes/workflows/contracts/cases and dependencies; file names such as OrderController are implementation details within that slice, not file-based migration units. Plain language describes mutually dependent components; developer details may show graph groups and symbol spans.

## Live execution state and project overview

[state.schema.json](../../specs/execution/state.schema.json) defines a reconstructable projection read at a committed sequence watermark. It exposes control mode, lifecycle state, active runs/stage instances/tasks, focus, last actual activity, last heartbeat, progress metrics, discoveries/changes, blockers, required approvals, next action and relevant refs. Null task/reason/result fields explicitly mean no active task or no result yet; every active task requires a recorded rationale. An idle/paused/completed project is explained rather than assigned fictitious work.

`status` includes this projection alongside existing gate validity and coverage. For an older snapshot without the additive execution field, clients report “execution detail unavailable,” not invented progress. New services must populate it. The overview shows source → requested target, exact resolved environment, current subjects, scope, current work and counted routes/workflows/units/cases. “Waiting for nothing” is shown only after checking current blockers and required inputs. Independent work may continue while a dependent scope waits; the overview must show both.

## Progress with a meaningful basis

Every metric has `metric_id`, definition, numerator, denominator, unit, scope hash, basis artifact ref, accounting revision, evidence sequence, freshness and a nullable percentage. The formula is `100 × numerator / denominator`; zero or not-yet-known denominators yield null. Require `0 ≤ numerator ≤ denominator`. Never estimate progress from elapsed time, tokens consumed, log length or an agent's confidence.

| Metric | Numerator / denominator |
| --- | --- |
| Scan work | Processed files / files in the frozen manifest, labeled as work rather than completeness of understanding |
| Observation queue | Completed scheduled scenarios / scenarios in the current queue revision; exploration can add work |
| Workflow characterization | In-scope workflows with current evidenced contracts/cases / inventoried in-scope workflows at that revision |
| Migration | Current completed units / units in the confirmed plan |
| Compatibility | Current PASS cases / required cases in the frozen corpus; MISMATCH/UNKNOWN/unevaluated remain in denominator |
| Overall work, optional | Sum of satisfied weights / sum of persisted work-item weights in a versioned accounting manifest |

Overall work is null until a persisted manifest defines scope, all included work items, weights and completion criteria. Show its revision and why it changed; revised scope can lower progress. Do not average stage percentages or treat eleven stage labels as equal effort. Retries cannot double-count completion; stale results no longer satisfy the current numerator. A 100% overall work number cannot authorize Gate 3 or imply full application parity. [COVERAGE](../05-verification/COVERAGE.md) governs compatibility and release acceptance.

Counts such as requests, queries, session mutations and external calls cite captures and distinguish discovered/observed/candidate/unknown items. Query count is diagnostic, not a parity requirement. Failed/completed actions and mismatches have separate totals. No fabricated ETA: show an estimate only with a documented measured basis and uncertainty; otherwise say unavailable.

## Liveness, timeline and return-session changes

Record a start before long work and a committed event at each meaningful boundary. While running, emit progress or heartbeat at least every 15 seconds, including during slow subprocesses; publish the event before presenting it. After 45 seconds without a heartbeat, render liveness `STALE` with last known work, not “still running” or an automatic failed verdict. A heartbeat proves worker liveness, not forward progress; show last activity separately and explain expected external waits. Timeouts are runner/operation policy, not a cosmetic UI timer.

Timeline order is SQLite `sequence`, not timestamp. Sequence is global and may have gaps within one project. Use exclusive `after_sequence`, bounded pages (default 100, maximum 500), a fixed `through_sequence` watermark per read and `next_sequence`; deduplicate event ID/sequence across delivery retries. Filter project ownership in every query. Never require a cursor to be contiguous or let another project's events enter a page. An expired/future/invalid cursor produces a structured error or explicit resynchronization need, never silent truncation.

“Since your last session” compares an explicit client-held prior watermark with the current watermark. Reading does not advance that client marker or mark alerts seen for everyone. Return typed additions/completions/resolutions/new blockers/lessons and invalidations with entity IDs/evidence refs; count distinct committed transitions, deduplicate retries, and keep rediscoveries, stale work and reopened mismatches visible. The client advances its own marker only after rendering the entire page. Large deltas are paginated; do not give complete totals for a partial page. Historical stage events remain inspectable across target changes and restarts.

## Notifications and human decisions

Warnings, blockers and approval cards name impact, evidence, options, current/requested/recommended values, affected units and required input. Notification views deduplicate by finding/decision subject revision, remain pending until resolved, and do not repeatedly prompt on every heartbeat. Desktop alerts are local, user-configured presentation of committed records; no email/chat/cloud delivery is implied. Verbosity never hides blocking or security-relevant decisions.

Gate 1/2/3 cards invoke the existing exact-subject confirmation protocol. `requires_user: true` only requests attention. Approval received events are committed with `user_confirmations`; approving a stale card fails freshness checks. Recommendation decisions follow the frozen target/plan rules. “Keep PHP 8.3” preserves requested intent and records the decision; changing it creates a successor target and revalidation, never a silent downgrade. Receipt/dismissal of an alert is not approval. A user cannot waive a technical incompatibility through an approval button.

## Pause, continue, cancellation and recovery

Execution control modes are `RUNNING`, `PAUSE_REQUESTED`, `PAUSED`, `IDLE`, `COMPLETED`, `FAILED`. They are projections of durable control events and operation records, not additional lifecycle/gate states. Explicit human `pause` records PAUSE_REQUESTED, stops new claims/actions for that project, and asks active workers to reach safe checkpoints. Continue lease renewals only until a safe boundary; do not let a lease expire while publishing. A multi-file patch must finish or reconcile its journal before pause is acknowledged. Fixture replay may be cancelled and restored rather than serialized mid-request.

Report PAUSED only when active work is safely stopped and checkpoints/resource cleanup are recorded. Show any task still stopping and why. A checkpointed operation may retain its existing RUNNING storage status while execution is paused; that status alone does not indicate liveness. Release worker leases coherently and preserve attempts/history; interrupted unsafe-to-resume operations are cancelled and replaced by fresh attempts. No new side effects are admitted after acknowledged pause. Read-only inspection and explicit approval review remain available.

`continue` is a human command clearing the control pause after checking drift, gates, leases and readiness; a missing approval leaves affected work WAITING_USER/BLOCKED. `resume` remains recovery/reconciliation and must preserve a human pause. Restart does not clear it. Closing status watch or an MCP connection does not pause/cancel the engine. Cancelling an operation is distinct from pausing the project; neither deletes evidence nor makes a failed/unknown result pass. UI “Skip” can only propose a visible scope/plan change requiring the applicable gate; it cannot silently waive required cases.

Initial CLI pause/continue commands use the local human-control channel. MCP exposes read-only control state; no agent tool grants or clears human control. Future desktop controls must invoke authenticated local human services, not impersonate an agent actor. The local-host trust limit is the same as the existing confirmation protocol; a boolean client claim is not proof of human intent.

## Presentation and interfaces

Verbosity is `simple`, `detailed`, or `developer`: plain-English summary; summary plus counts/decisions/gaps; then sanitized technical entity/trace/graph details. All use the same facts and cursor; verbosity does not change gates, work or privacy. Developer view never reveals credentials, raw sensitive request data, prompt contents or hidden reasoning. Core event strings use deterministic templates; optional agent explanations are attributed proposals grounded in recorded refs.

| Interface | Contract |
| --- | --- |
| `status` | One snapshot; additive execution state, honest coverage and freshness |
| `status --watch` | Human live rendering; poll committed state every two seconds; Ctrl-C stops only watching |
| `execution show --after-sequence N --limit N` | One state/timeline/change page; machine-readable with `--json` |
| `pause`, `continue` | Persistent human execution control; explicit idempotent local command, no gate approval |
| MCP `project.get_execution` | Same read-only state/timeline/change page; bounded polling/reconnect |
| MCP execution resource | Same canonical state at a watermark; no required subscription or transport change |
| Future desktop UI | Same state/events/explanations and existing approval subjects; no new ledger |

`status --watch --json`, watch with quiet/dry-run, and unrecognized verbosity are invalid arguments. Non-watch JSON remains exactly one existing CLI envelope. Machine clients poll `execution show --json` or MCP with cursors; no undocumented JSONL streaming exception. Approval/cursor/control details are specified in synchronized CLI/MCP schemas. Negotiated MCP progress notifications can mirror current call activity but are not the durable timeline or an authority source. No new LLM, hosted websocket service or desktop implementation is required for initial engine transparency.

## Clear failures and correctness boundaries

An error view shows task/stage/scope, concrete difference or missing evidence, sanitized diagnostics, recorded attempts, labeled hypothesis, affected work and next action. For example, “legacy returns 302 /orders/918; target returns 200 /orders” links the exact case/diff; a suspected validation branch is a hypothesis, not an asserted cause. Repeated repairs respect the existing attempt limit. ERROR severity, task execution status and behavioral verdict remain separate.

Validate/redact events before persistence; source-controlled text is untrusted display data, never commands or instructions. Domain updates, completion/control/approval events and related metadata commit atomically. Progress/heartbeat transactions do not change project lifecycle, approve subjects or invalidate a worker's business revision solely because telemetry advanced. Successful effects are not reported before publication. If the durable event sink fails, stop admitting further work at a safe boundary and preserve recoverable publication evidence; do not continue an opaque run or repeat an external effect because telemetry failed. Existing transactional recovery rules still apply.

## Required acceptance evidence

The first engine milestone includes instrumentation and live inspection. Implementation tests must cover a slow action with heartbeats, missing-heartbeat freshness, concurrent stage instances, failed action versus UNKNOWN/MISMATCH, duplicate/out-of-order delivery, global sequence gaps, page watermarks, per-client unread markers, target invalidation, pause during publication/replay, restart while paused, continue with stale gates, rejected skip, disconnected observers, storage failure, privacy redaction and a forged approval event.

Progress tests cover unknown/zero denominators, numerator bounds, accounting-manifest revisions, retries and stale PASS removal. Contract examples validate shapes only; race/recovery/client/sandbox tests must establish actual behavior. A task is not complete merely because it emits reassuring text. Completion evidence includes the events/state/explanations corresponding to its real work.

## Audit and pagination clarification

Inspection/audit events do not count as task activity, discoveries, heartbeat, progress or user unread work. Cursor pages filter typed execution events and advance across inspected non-execution ledger positions so audit-only traffic cannot create phantom progress or a polling feedback loop. A fixed through_sequence spans all pages in one window; current state is projected at that watermark, with as-of identity visible. after_sequence ≤ through_sequence is required, and historical reads never make old results current. When the window is exhausted, next_sequence equals through_sequence even if no execution events matched.

## Terminal execution scope

Control-mode COMPLETED means the current scheduled execution scope has no unfinished required work with current evidence; it is not deployment/cutover approval or a full-application equivalence claim. IDLE means no work is actively admitted, including waiting for input; FAILED names an actual unrecovered execution failure. Always display scope, outstanding approvals and last outcome. Preserve the separate required corpus/release guards even when all scheduled tasks have finished.
