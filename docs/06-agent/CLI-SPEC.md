# CLI contract

**Status:** Initial implementation specification. No CLI binary exists yet. [commands.json](../../specs/cli/commands.json) is the complete command/input/output registry. The command catalog contract is version `2.1`; the envelope remains `1.0`; changes require synchronized documentation, schemas and compatibility fixtures.

## Global behavior

Every command accepts `--json`, `--project <workspace>`, `--quiet`, `--verbose`, `--dry-run`. Workspace defaults to current directory. Flag parsing is strict; reject unknown flags/positionals, duplicate scalar flags, and `--quiet` with `--verbose`. Boolean flags are false unless present. IDs and hashes are validated before operations.

With `--json`, stdout contains exactly one newline-terminated envelope conforming to `cli-output.schema.json`; diagnostics go to stderr. Envelope `command` is the space-separated canonical registry name, `version` is `1.0`, timestamp is UTC, and duration is nonnegative milliseconds. `data` validates against the command's output schema on success. On failure, `data` contains any available sanitized partial result or `{}` and `errors` contains structured catalog entries. `success` is true only for exit 0.

Exit codes: `0` success; `1` invalid input/environment/operational failure; `2` completed verification with MISMATCH or UNKNOWN, or failed performance acceptance; `3` missing human authorization for cutover/sensitive export; `4` missing/stale Gate 1 or Gate 2. The priority is operational failure, then gate refusal before execution, then completed verification outcome. Gate 3 refusal uses 3. No subprocess exit code leaks as the engine exit code.

SIGINT/SIGTERM cancels owned child work and emits `ERR_CANCELLED` when safe to write the envelope. A human status watcher owns only its read loop: interrupting it does not cancel the engine, and its clean exit behavior is specified below. Help/version are successful responses with `data.kind`, `data.text`, `data.engine_version`, `data.contract_version` and no project lookup; they are renderer operations outside the domain command registry. MCP stdio output uses JSON-RPC, not this CLI envelope.

`--dry-run` performs safe reads/validation and returns the command's would-be result with an envelope-level `dry_run: true`; no DB/event/approval/artifact/file mutation, lease claim, dependency acquisition, source execution or gate unlocking occurs. For operations requiring actual evidence, return `data: {}` and `dry_run_summary` instead of inventing a report. A dry-run does not establish completion. Ignored global flags are not silently accepted.

## Public commands and options

Global flags are omitted from this table. Hyphenated CLI flags map to underscore schema fields (`--case` → `case_id`, `--unit` → `unit_id`).

| Command | Arguments/options | Prerequisite and result |
| --- | --- | --- |
| `init [source-path]` | Defaults source to `.` | Minimal workspace and detection; no auto-confirm or target generation |
| `detect [source-path]` | `--fast` bounded metadata subset | Detection ID, manifest hash, evidence/confidence and Gate 1 pending |
| `doctor` | No command flags | Structured check/remediation summary; never executes untrusted PHP on host |
| `confirm source` | Required `--detection-id`, `--manifest-hash`; `--php`, `--framework`, `--database`, repeated `--override key=value`; `--prepare`; `--subject-hash` | Display measured facts and separately proposed assertions; explicit local human approval persists Gate 1 |
| `confirm plan` | Required `--subject-hash <plan_hash>` | Exact reviewed plan; changes require replanning |
| `confirm cutover` | Required `--subject-hash <release_hash>` | Current evidence/runbooks; Gate 3 record |
| `approval issue` | Required `--gate source|plan|cutover`, `--subject-hash` | Human terminal approval; emits a single-use token for MCP |
| `analyze` | No command flags | Gate 1; source/application IR and findings |
| `observe` | Required `--case <scenario_id>` | Gate 1; scenario may be a supplied input-only case before a locked master |
| `discover` | No command flags | Gate 1; evidence-backed workflow/contracts/cases |
| `plan` | Optional `--selection-ref`, `--target php/laravel`, `--style modular_monolith` assertions; `--recalculate` | Gate 1 plus current human CONTINUE_PLANNING selection; exact checkpoint-linked executable plan |
| `migrate` | Exactly one of `--unit <id>`, `--all` | Gates 1/2; ready units and target patches |
| `patch stage` | Required `--patch-id` | Gates 1/2; validate/import declared target patch from the project import area |
| `patch apply` | Required `--patch-ref` | Gates 1/2; journaled target publication with current fence/before hashes |
| `verify` | Exactly one of `--case <id>`, `--unit <id>`, `--all` | Gates 1/2; report with deterministic verdicts |
| `benchmark` | Exactly one of `--unit <id>`, `--all` | Current passing verification; reproducible benchmark report |
| `status` | `--watch`, `--verbosity simple|detailed|developer` | Read shared execution snapshot with stage/task/reason, measured progress, freshness, gates and blockers; watch is human-only rendering |
| `execution show` | `--after-sequence <n>`, `--through-sequence <n>`, `--limit <1..500>`, `--verbosity simple|detailed|developer` | Read one consistent state/timeline/change page, no work claim or shared acknowledgment |
| `pause` | No command flags | Explicit local human action; persist request, stop new work and show pending safe-stop state |
| `continue` | No command flags | Explicit local human action; clear control pause after freshness/readiness checks; never confirms gates |
| `case show` | Required `--case <id>` | Read exact registered canonical case and provenance |
| `plan show` | Required `--plan-id` | Read exact registered immutable plan |
| `artifact read` | Required `--ref <sha256:digest>` | Read sanitized registered project evidence; typed UTF-8/base64 content, size/privacy policy |
| `next` | No command flags | Peek highest ready mission; `mission: null` is valid when no work |
| `resume` | No command flags | Reconcile unfinished operations/leases/journals; return snapshot |
| `mission claim` | Required `--mission-id` | Atomic ready-mission lease; owner derived from authenticated session |
| `mission renew` | Required `--mission-id`, `--fencing-token` | Current owner/fence; extend lease |
| `mission complete` | Required `--mission-id`, `--fencing-token`, repeated `--evidence-ref` | Engine validates mission-specific completion evidence |
| `assess` | Required `--scope-ref`; optional `--observation AVAILABLE_ONLY|SAFE_RUNTIME`, `--background` | Gate 1; shared analysis checkpoint, completed report even with blocked feasibility |
| `assessment show` | Required `--assessment-ref` | Historical report plus explicit current per-finding validity |
| `assessment decide` | Required `--assessment-ref`, `--action`, `--rationale`; optional `--selected-scope-ref`, `--target-intent-ref`, repeated `--finding-id`, `--decision-ref`, `--requested-evidence` | Trusted local human selection; never implementation approval |
| `report` | Required `--format json|markdown`; optional `--kind auto|assessment|migration`, `--assessment-ref` | Render assessed-state or migration artifact without acquiring evidence or authorizing continuation |
| `export` | Required `--manifest-ref`, `--output` | Sanitized reviewed export; sensitive data requires human approval |
| `skills search` | Required `--query` | Local skill/capability results; no mandatory global store |
| `learn record` | Required `--experiment-id`, `--finding`, `--remediation` | Project lesson bound to completed evidence |
| `cutover prepare` | Required `--subject-hash`, `--output` | Valid Gate 3; writes approved package/runbook, no production deployment |

The source confirmation `--prepare` path returns a prepared subject (`project/prepared-confirmation.schema.json`) instead of a confirmation record. Registry response variants cover it. Interactive confirmation requires typing the displayed hash; noninteractive local human approval must name the reviewed `--subject-hash`. `--auto-confirm`, `model`, `clean`, `project.pause`, and arbitrary production commands are outside the initial catalog.

MCP mutations require idempotency keys; CLI calls generate a request key, and global `--request-id <id>` allows explicit reuse for scripts/recovery. Reusing a request ID with changed payload fails. Output paths must be contained in an explicit configured export/target root, and source/target overlap is rejected.

## Example lifecycle (planned commands)

```bash
rewrotron init /path/to/legacy --project /path/to/workspace --json
rewrotron confirm source --detection-id det_001 --manifest-hash <manifest_hash> --prepare --project /path/to/workspace --json
rewrotron confirm source --detection-id det_001 --manifest-hash <manifest_hash> --subject-hash <configuration_hash> --project /path/to/workspace
rewrotron analyze --project /path/to/workspace --json
rewrotron observe --case login-input --project /path/to/workspace --json
rewrotron discover --project /path/to/workspace --json
rewrotron assess --scope-ref <scope_ref> --project /path/to/workspace --json
rewrotron assessment show --assessment-ref <assessment_ref> --project /path/to/workspace --json
# Assessment-only use stops here. Continuation is an explicit local human action:
rewrotron assessment decide --assessment-ref <assessment_ref> --action CONTINUE_PLANNING --selected-scope-ref <scope_ref> --target-intent-ref <intent_ref> --rationale "Proceed with selected work" --project /path/to/workspace
rewrotron plan --selection-ref <selection_ref> --project /path/to/workspace --json
rewrotron confirm plan --subject-hash <plan_hash> --project /path/to/workspace
rewrotron migrate --unit auth-login --project /path/to/workspace --json
rewrotron verify --all --project /path/to/workspace --json
rewrotron benchmark --all --project /path/to/workspace --json
rewrotron report --format markdown --project /path/to/workspace --json
rewrotron confirm cutover --subject-hash <release_hash> --project /path/to/workspace
rewrotron cutover prepare --subject-hash <release_hash> --output /path/to/package --project /path/to/workspace --json
```

Angle-bracket hashes are values returned by the prior reviewed operation, not commands to paste without substitution. Gate 3 is reached only after the entire confirmed project scope passes; the one-unit migration line illustrates a slice.

## Required contract checks

Every command has success, invalid-input, domain-error, missing/stale-gate, cancellation and dry-run fixtures where applicable. Verify stdout contains exactly one envelope, exit code matches it, output schema resolves locally, and no secret appears in stdout/stderr. Registry completeness is a CI check; adding flags/results requires schema and examples in the same change.


## Result summaries and no-workspace responses

Shared operation results name artifact refs/entity IDs/blockers; `summary` contains bounded descriptive counts/checks, not gate authority or verdicts. Doctor may return `state: null` when no workspace exists. Canonical artifacts carry full IR/patch/run/report documents. Plan and verifier/benchmark responses use dedicated typed schemas. A prepared approval command returns a reviewed subject; the token issuing result places the one-time token only in its local human response, never events/log exports.


CLI-only agents can inspect mission context, `case show`, `plan show` and `artifact read`, then stage/apply a target patch and invoke verification. Read operations check project ownership, reference hashes and privacy/capture limits. They cannot open arbitrary filesystem paths or export protected secret material. Case/plan absence is a structured lookup error, never an invented document.

Migration reports validate against `specs/reports/migration-report.schema.json`; release_ref requires complete current release prerequisites. Assessment reports validate against `specs/reports/assessment-report.schema.json` and require no executable plan. `report --kind auto` renders assessed-state before a plan (detection summary only before source confirmation); after a plan it renders migration state. Explicit `--kind assessment` remains available with a blocked plan or later migration. A supplied assessment_ref renders exactly that report, labeled historical/current applicability; it cannot request kind=migration. Rendering does not perform analysis/observation, plan, publish target code or approve continuation.

## Target and dependency review through existing commands

`plan` first checks the supplied or current explicit human CONTINUE_PLANNING selection and currently valid assessed work, then loads validated user intent from `rewrotron.yaml` (or the documented legacy/split layout), freezes its revision and assesses recursive dependency feasibility. With no command-specific flags it uses the configured target. Existing `--target` and `--style` flags are optional assertions of that intent; conflicting values fail before planning and never rewrite it. Explicit runtime edits, such as 8.3 → 8.4, refresh affected assessment/feasibility findings; renewed implementation requires plan review and Gate 2 confirmation, while assessment-only use may stay stopped. No fallback framework/runtime is silently chosen.

`plan show` exposes the requested specification, exact resolutions and v2 `subject` artifact refs. `artifact read` reads the registered recommendation, dependency and feasibility documents. Human report rendering shows detected/recommended/requested/resolved versions, used capabilities, upgrades/replacements, required transformations, evidence/risk, decisions and blocked scope. `status`/`resume` retain source progress while reporting affected stale target work and required review. CLI envelopes and existing flag-based calls remain compatible; target/style assertions are now optional in synchronized CLI/MCP input schemas; shared artifacts gain explicitly versioned v2 records; v1 is historical read-only.

A future `rewrotron configure` interactive review is proposed, not part of the current registry. Add its input/output schemas, catalog, dry-run/approval behavior and compatibility fixtures before implementing or advertising it. See [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md).

## Transparent execution and live inspection

[EXECUTION-OBSERVABILITY](EXECUTION-OBSERVABILITY.md) defines the shared events/state, reasons, work accounting, return-session changes, approval cards and persistent control. New status/snapshot services populate the additive `execution` field. Visible stage/control labels do not change lifecycle/verdict enums. `simple`, `detailed` and `developer` affect human presentation, never redaction, work or approval.

`status --watch` refreshes committed state every two seconds until Ctrl-C; leaving watch affects only the observer. It is an explicit human-rendering mode: `--watch --json`, watch with quiet/dry-run, and unknown verbosity fail `ERR_INVALID_ARGUMENT` before starting the watcher. Non-watch JSON still emits exactly one normal envelope. A successful watch interrupt exits 0 with a final human snapshot; errors retain their documented operational exit codes. Required user decisions/blockers remain visible in every verbosity. Machine clients use one-envelope `execution show --json` requests with cursors instead of an undocumented streaming format.

`execution show` defaults to `after_sequence: 0` and `limit: 100`. If omitted, through_sequence is captured at read start; pass the returned watermark on subsequent pages for a fixed window. Cursor is exclusive and may contain global per-project gaps. `next_sequence` advances to the last inspected relevant ledger position (through_sequence when exhausted); return `has_more` so a client knows whether a page covers the whole window. Validate nonnegative cursors, after ≤ through ≤ current watermark and same-project access. The page's state is CURRENT at state_sequence N; through_sequence M bounds only events/changes. Changes cover only its delivered window; clients deduplicate entity/event IDs and cannot label partial totals complete.

`pause`/`continue` are local human-control commands, using trusted session context and the existing request-ID/idempotency rules; no caller actor field, approval token or fourth lifecycle gate is introduced. Pause returns current execution state, possibly PAUSE_REQUESTED while workers reach safe boundaries. It never falsely reports PAUSED while publication/replay is active. Continue leaves missing/stale gates or dependency blockers unresolved. `resume` remains reconciliation and never clears pause. Dry-run validates/reports would-be control without mutation. The local-human trust limit follows the existing approval protocol.

Planned supervision calls:

```bash
rewrotron status --watch --verbosity detailed
rewrotron execution show --after-sequence 120 --limit 100 --json
rewrotron pause --json
rewrotron resume --json
rewrotron continue --json
```

These specifications do not claim a working CLI binary or executed live-watch/control tests.

## Decision operations and ownership (contract v2)

`decision propose --decision-ref <sha256:digest>` validates/imports an immutable decision proposal under Gate 1. `decision show --decision-id <id>` returns its exact registered record and requires Gate 1. Actions in decision.schema are ACCEPT, REJECT, SELECT_STRATEGY, SUPPLY_ALTERNATIVE, REQUEST_EVIDENCE and APPROVE_EXCEPTION; each is a typed operation selected by the input artifact, with strategy/provider/scope/evidence fields. For an unregistered decision_ref, the service reads only `.rewrotron/imports/decisions/<digest>.json` under its configured owner-only import root, verifies the digest/schema/ownership and sanitization, registers immutable bytes, then records the proposal. A registered ref is read through the artifact store. No caller path is opened directly. Neither command approves the decision: `plan` binds selected refs and `confirm plan` is the human approval. Alternatives/rejection create successor proposals; successor planning requires the explicit continuation decision; generated views are never edited as authority. An exception changes visible scope/policy only and cannot waive unsafe execution or make UNKNOWN PASS. CLI users can complete every decision without MCP.

Long assess/analyze/observe/discover/migrate/verify/benchmark calls accept `--background` (schema background=true) for explicit supervisor admission; default foreground lifetime follows RECOVERY. Unsupported supervisor returns ERR_UNSUPPORTED_CAPABILITY. Existing result/envelope typing remains unchanged: background returns the admitted operation/checkpoint result, not invented completion; polling status/execution show supplies current outcome. verify/benchmark dedicated reports are returned only on completion; background admission uses operations/result as an explicit catalog output variant. Machine input payloads use registered raw refs/exact numeric strings. through_sequence is an event window cursor, never historical state.

## Assessment-only and optional continuation

[PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md) owns the checkpoint. `assess` returns operations/result with the report artifact ref on completion; migration_feasibility BLOCKED/UNKNOWN does not make that operation failed or assign CLI error exit. Background admission remains RUNNING/PREPARED until the report is durable. `assessment show` returns assessment-view.schema, with immutable report and separately evaluated current validity. `status` has an additive optional assessment view (null means none); assessment completion does not advance projects.status to VERIFIED or PLAN_SYNTHESIZED.

Only human local `assessment decide` creates a selection. CONTINUE_PLANNING requires both selected_scope_ref and target_intent_ref; other actions never create implicit continuation. No actor/channel/approval token fields are caller inputs. `--action` values are those in assessment-selection.schema; selection never grants Gate 2. STOP supersedes an unexecuted continuation; `continue` only clears pause. `plan` without a current valid continuation returns ERR_ASSESSMENT_DECISION_REQUIRED. Changed selected inputs require renewed human selection or an explicitly reviewed scope/intent decision; unchanged source findings remain reusable with witnesses. Current plan documents use envelope 2.1 with hashed selection/basis closure; old envelopes remain read-only history. CLI/MCP catalogs are contract_version 2.1; existing envelope versions remain unchanged. Dry-run never persists a report/selection or authorizes planning. Every shared lifecycle operation remains possible through CLI alone.

`plan --performance-policy-ref <sha256:digest>` supplies the immutable proposed policy to include in PLAN decision_refs. A registered ref resolves through the project artifact store; a new policy is imported only from the owned `.rewrotron/imports/performance-policies/<digest>.json`, with digest/schema/ownership/sanitization checks matching decision imports. No arbitrary path is accepted. Omitting the flag can reuse only the exact previously selected policy with current validated inputs and a visible unchanged-policy planning delta; otherwise ERR_INVALID_ARGUMENT identifies the missing policy. An explicit not-applicable policy is required where no performance claim is requested. This is a plan proposal, not approval. Gate 2 shows the claim/workload/metrics/limits and approves the exact closure. `benchmark` consumes that policy and cannot supply alternate thresholds at execution time.

## Review-2 domain records

Catalog contract remains 2.1; JSON envelope/protocol versions are unchanged. Existing observe/assessment/verify/artifact/report services expose the same current typed records: source-local identities, evaluation contexts, principal/initial fixtures, responses, effect payload/occurrences, execution witnesses and derived pairwise maps. CLI imports register current scenario 2.0 closure as PROJECT-INPUTS specifies. MCP does not implement separate evaluation/pairing logic. Inspection returns only authorized sanitized projections; private contexts/credentials/journals remain restricted. Current version admission follows specs/README and SERVICE-CONTRACTS; historical schema validity is never execution permission. Storage-failure outputs distinguish the last committed sequence from emergency diagnostics under RECOVERY.
