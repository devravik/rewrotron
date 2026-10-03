# MCP protocol, tools and resources

**Status:** Initial implementation specification. Protocol revision `2025-11-25`, stdio transport, no vendor SDK required. [tools.json](../../specs/mcp/tools.json) enumerates every public tool with input/output schema, gate and mutation semantics.

## Protocol lifecycle

Implement JSON-RPC 2.0 initialization, version/capability negotiation, `notifications/initialized`, `ping`, `tools/list`, `tools/call`, `resources/list`, `resources/templates/list`, `resources/read`, cancellation and negotiated progress. Do not expose tools before initialization completes. Advertise only implemented capabilities; no sampling, experimental task protocol or remote transport is required.

UTF-8 newline-delimited messages use stdout exclusively; logs go to stderr. Reject malformed JSON (`-32700`), invalid requests (`-32600`), unknown methods (`-32601`) and invalid parameters (`-32602`). Domain operation failures use tool results with `isError: true`, a text explanation and `structuredContent` containing the structured error catalog entry. Successful tool results contain typed `structuredContent` and a matching short text representation; do not wrap them in the CLI envelope.

Limit one message to 16 MiB and allow cancellation through context. Heavy tools execute with persisted operation IDs and enforce the same gates/drift/fences as CLI. SDK dependency choice/pinning is validated in the machine-interface milestone; business services cannot depend on protocol types.

Primary specifications: [lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle) and [transports](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports). Streamable HTTP is deferred and requires its own authentication/origin/localhost policy. Legacy SSE is not the initial transport.

## Public tool catalog

| Group | Tools | Semantics |
| --- | --- | --- |
| Project | `project.init`, `project.detect`, `project.doctor`, `project.get_snapshot`, `project.get_execution`, `project.resume` | Register workspace, record bounded proposals, inspect/reconcile state |
| Human approvals | `project.prepare_source_confirmation`, `project.confirm_source`, `project.confirm_plan`, `project.confirm_cutover` | Prepare subject; consume a human-issued scoped token and persist exact approval |
| Analysis/observation | `analysis.run`, `observation.run`, `behavior.discover` | Gate 1 operations with completeness and evidence |
| Behavior lookup | `behavior.get_case` | Read exact case document; no arbitrary artifact path |
| Assessment checkpoint | `assessment.run`, `assessment.get` | Shared Gate 1 analysis coordination; read report and current validity; no human continuation/approval tool |
| Migration | `migration.plan`, `migration.get_plan`, `migration.generate_code`, `migration.stage_patch`, `migration.apply_patch` | Propose/read plans; gated ready-unit generation |
| Verification | `verification.run_case`, `benchmark.run` | Sandbox replay and current performance evidence |
| Missions | `research.get_next_mission`, `research.claim_mission`, `research.renew_mission`, `research.complete_mission` | Peek is read-only; lease mutations are fenced and evidence-checked |
| Learning/reports | `knowledge.record_lesson`, `skills.search`, `reports.generate` | Project-local evidence; immutable sanitized report |
| Cutover | `cutover.prepare` | Package exactly the approved release; no production mutation |

All tools require registered `project_id`. Workspace registration maps a configured root to that ID; later calls cannot supply arbitrary source/output roots. Mutations require `idempotency_key`; actors/lease owners derive from the authenticated local session and are not agent-supplied identities. Init/detect paths must remain within configured allowed workspace/source roots.

`project.confirm_source` does not accept new overrides. The prepare operation computes the merged reviewed subject; human `approval issue` outside MCP issues a subject-bound token; the confirm call consumes it. Ordinary tools cannot mint human approval. Token/actor data is redacted from audit payloads. Trust limits are explicit in [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md).

`readOnlyHint` describes business mutation, not absence of audit writes. Generation remains target-only, verification has sandbox effects, and mission claiming changes queue state. No raw SQL, arbitrary shell, arbitrary file write, verdict override or unproven source deletion tools exist.

## Resources

Read-only resource templates are:

- `rewrotron://project/{project_id}/snapshot` → snapshot JSON with additive shared execution state.
- `rewrotron://project/{project_id}/execution` → current execution state labeled with state_sequence.
- `rewrotron://project/{project_id}/coverage` → snapshot coverage object and scope hashes.
- `rewrotron://project/{project_id}/case/{case_id}` → canonical case JSON.
- `rewrotron://project/{project_id}/plan/{plan_id}` → canonical plan JSON.
- `rewrotron://project/{project_id}/artifact/{sha256}` → sanitized registered artifact, subject to size/privacy policy.

Resources validate IDs and project ownership, carry JSON MIME types for structured documents, and cannot resolve arbitrary host paths. Pagination uses opaque stable cursors. Clients without MCP can perform every essential migration operation through the documented CLI.

## Compliance fixtures

Initialization/version mismatch, tool/resource listing, each input/output schema, empty mission queue, required approval token, domain-error versus protocol-error representation, idempotency conflict, cancellation, Unicode/newline framing, stderr separation, oversized messages and concurrent leases must be tested. Advertised tools must exactly match the registry and permission matrix.


## Server startup and root registration

`rewrotron-mcp --project <workspace>` registers the existing workspace and reads its project ID before accepting tools. For a first workspace, startup accepts explicit `--workspace-root`, `--source-root`, `--target-root` roots and `project.init` supplies the schema-valid proposed project ID; the engine validates and adopts it for the new workspace; an existing workspace must match its persisted ID. Roots are startup policy, never expanded by a later agent request. STDIO identity derives from the launching local account. Client-provided names, headers or `project_id` do not establish human approval identity.

## Target and dependency decision boundary

`migration.plan` first requires the supplied or current persisted explicit human CONTINUE_PLANNING selection; it then consumes the validated frozen user target. Existing target/style inputs are optional assertions of configuration; omitting them uses the validated project target. Conflicting assertions fail and cannot silently replace requested versions. The plan's v2 `subject` exposes exact resolution and registered recommendation/dependency/feasibility artifacts through existing resources. `project.confirm_plan` approves the exact target and decisions transitively through `plan_hash`; it retains the existing human-issued token requirements. No target approval tool or unrestricted configuration-write tool is introduced.

Pending target edits invalidate affected approval/results and prevent stale worker publication while preserving unrelated source work. Agents can propose alternate resolutions but cannot make them approved configuration. The typed decision proposal API below is included; an interactive configure renderer remains deferred. See [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) and [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md).

## Live execution state, timeline and return-session changes

`project.get_execution` accepts required project ID plus optional after_sequence, through_sequence and bounded limit (default 100, maximum 500). It returns `specs/execution/page.schema.json`: current execution state at state_sequence N plus ordered events/changes bounded by through_sequence M ≤ N; no historical state reconstruction. Missing through_sequence captures the current committed watermark; supplying it fixes the window across pages. Cursors are exclusive, global sequence gaps are valid, and next_sequence/has_more prevent silent loss. Reject after > through or through beyond the current ledger; validate project ownership. Reading does not claim a mission, approve a gate, clear pause or advance a shared last-seen marker.

Clients poll at bounded intervals (two seconds for an active view) and resume from their last fully rendered sequence. The execution resource provides the same canonical state; durable event pages are independent of optional negotiated progress notifications. No push subscription, hosted service, transport change or stdout framing exception is required. Source messages are untrusted display data; developer details retain all privacy rules.

Initial `project.get_execution` is read-only; no public MCP pause/continue tool is added. Human local CLI `pause`/`continue` control the same projected state, and a future desktop must use authenticated local human services. `project.resume` reconciles without clearing human pause. Approval cards still consume existing human-issued exact-subject tokens through project confirmation tools; events cannot forge authorization. See [EXECUTION-OBSERVABILITY](EXECUTION-OBSERVABILITY.md).

## Decision proposals and supervised operations

`decisions.propose` uses the same registered-ref or contained decision import protocol as CLI and imports a project-owned decision_ref and idempotency key under Gate 1; its typed action supports ACCEPT/REJECT/SELECT_STRATEGY/SUPPLY_ALTERNATIVE/REQUEST_EVIDENCE/APPROVE_EXCEPTION. `decisions.get` reads an exact decision_id under Gate 1. These return decision.schema v2, always PROPOSED. Human local approval issue plus `project.confirm_plan` approves the exact immutable decision refs transitively. No MCP operation can approve its own proposal. Exceptions cannot bypass safety, capability or verdict guards. CLI decision propose/show has complete parity.

Long tools use engine-owned supervision, not the stdio client's process lifetime. Missing supervisor returns ERR_CAPABILITY_UNSUPPORTED. A request can return an admitted operation result while execution continues; catalog output variants include operations/result for verify/benchmark. All results name captured subjects/current versus historical status. Raw payload refs/exact strings protect large numeric identities; end-to-end transport vectors must verify client behavior. Claude/Cursor/Codex/Aider/Hermes/custom client integrations are UNVERIFIED until actual sessions pass. Stdio protocol existence alone is not client certification.

## Assessment checkpoint parity

`assessment.run` accepts project_id, scope_ref, optional observation mode AVAILABLE_ONLY/SAFE_RUNTIME and idempotency_key. It delegates to existing guarded analysis/evidence services under Gate 1, with supervisor ownership for long work. Completion returns operations/result naming the immutable assessment report. Migration feasibility BLOCKED/UNKNOWN is report content, not isError or failure of an otherwise published assessment. Unsafe/unavailable runtime produces explicit findings, never blind replay. No implementation tasks, target publication or automatic planning follow completion.

`assessment.get` accepts project_id and exact assessment_ref, returning assessment-view.schema with report and per-finding current validity. Retrieval works despite blocked/stale migration and cannot rerun evidence acquisition. `reports.generate` accepts kind auto/assessment/migration and optional assessment_ref with the same rendering-only rules as CLI. The artifact resource exposes the same canonical assessed-state record. Sensitive reports still require export controls.

MCP has no tool to mint a human continuation selection. Agents present options and direct the user to local CLI `assessment decide`; the persisted decision is then readable via registered artifacts. No MCP input actor name, recommendation acceptance or report digest counts as that decision or as Gate 2. `migration.plan` optionally names selection_ref and otherwise resolves the current explicit human continuation; missing selection returns ERR_ASSESSMENT_DECISION_REQUIRED. Plan envelope 2.1 exposes assessment/selection/planning-basis refs bound through the existing PLAN subject decision refs. `project.confirm_plan` still consumes an independently human-issued exact plan token. Catalog contract_version 2.1 does not change the pinned MCP transport/version or claim tested client integrations.
