# Service records and infrastructure ports

[COMPONENTS](COMPONENTS.md) owns package dependencies; [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md) owns canonical data; [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md) names the planned Go methods. This document fixes the shared operational records so adjacent tasks do not invent incompatible handles.

## Common request context

Every mutation receives project ID, operation/request ID, idempotency key, expected project revision, captured configuration/plan/target/corpus/policy hashes where applicable, and cancellation context. Trusted actor/channel/lease owner derive from constructor/session state. Client parameters cannot overwrite them. Idempotency hashes exclude transport request IDs but include all semantic input/subject fields.

## Operational record fields

| Record | Required fields |
| --- | --- |
| Project | ID, workspace root, source root, optional target root, parsed intent config, current state/revision, current configuration/plan hashes, adapter IDs/versions |
| Workspace | Project record, SQLite path, artifact/import/output roots, private correlation-key reference and close/lock ownership |
| SourceManifest | Sorted path/content/executable/symlink records, exclusions, completeness, content hash and scan tool/version |
| CaptureInput | Project ID, format/media type, bounded bytes/stream, configured sensitive field/value refs, capture dimension and limits |
| RedactedCapture | Sanitized bytes/stream, media type, redaction manifest, withheld dimensions, keyed correlation outcomes; no raw persisted payload |
| SandboxProfile | Source/target runtime/image digests, quotas, approved volumes, internal destinations/sinks, DB engine/schema/settings, reset strategy and profile hash |
| EnvironmentHandle | Project/operation ID, environment identity, container/network/volume IDs, safe local proxy endpoint, DB/session/cache/sink handles, reset/checkpoint refs, cleanup ownership |
| ObserveOptions | Scenario/case ref, EnvironmentHandle, fixture refs, step/body/redirect limits, required dimensions, probe version/hash |
| RuntimeTraceSet | Ordered observation trace refs, corresponding case/step IDs, completeness, captured subjects and control-run outcome |
| VerificationHarness | Adapter ID/version, safe startup/execution specification, supported dimensions, environment profile and compatibility-test refs |
| GenerationResult | Patch artifact ref, target provenance, generated file ownership, required build-check outcomes and blockers; no verified verdict |
| MissionLease | Mission record, trusted owner, fencing token, expiry, captured context hashes and checkpoint refs |
| RecoveryReport | Reconciled operation IDs, current hashes, repaired journals/resources, unrecoverable conflicts and required next actions |

Fields map to canonical schemas where a schema exists. Infrastructure handles are opaque within the process; they are never serialized as arbitrary commands or credentials. Planned service request names alias catalog inputs plus this context. A transient handle cannot count as durable approval/evidence.

## Ports and deterministic injection

- Repository provides transaction-scoped project/detection/baseline/approval/IR/case/plan/run/mission/event operations, expected-revision compare-and-swap, and same-project/reference validation. Only the SQLite implementation accepts raw database handles internally.
- ArtifactStore provides sanitized Put, registered Get, entity Link, Verify and reconciliation. Byte writes precede SQLite metadata; immutable evidence stays reachable.
- AdapterRegistry resolves a unique ID/version and manifest, returns supported source/target ports, and rejects missing/duplicate registrations. It is constructor-owned immutable configuration.
- Runner accepts a fixed ToolSpec (`executable_id`, argument array, safe working root, explicit env secret refs, timeout/output limits, sandbox handle), returns exit/capture/cleanup evidence, and never accepts a shell string from an agent.
- Clock supplies UTC `Now() time.Time`; IDGenerator supplies `NewID(prefix string) string`. Tests inject fixed/advanced clocks and deterministic unique IDs. Wall time never substitutes for event sequence or a fencing token.

Source/target ports are defined in SYSTEM-DESIGN. Their operations are invoked only through gate-checked services. Construction resolves credentials at execution into memory and redacts them before persistence; no repository API writes plaintext secrets.

## Result publication

A producer stages immutable result artifacts, captures all parent hashes and its live fence, then CompareAndPublish checks the current revision/subjects/ownership and commits projection plus event links. Failed publication returns a stale/conflict error with preserved diagnostic refs. A previous idempotent result is historical unless its subjects remain current.

Cross-record semantic checks cover duplicate IDs, local-ID/baseline resolution, ownership, gate parents, rule/case coverage, DAG order, fixture reset identity, target path/before hashes and lease expiry. JSON Schema validates structure; these service rules validate meaning.

## Execution records and injected ports

`ExecutionEvent`, `ExecutionState`, `ProgressMetric` and `ExecutionPage` map to `specs/execution/`. Events identify execution operation/run, stage instance, task/mission/entity, reason/result, captured subjects, refs and sequence. `ExecutionSink` validates/redacts append requests and participates in the domain transaction for state-changing completion/control/approval events. `ExecutionReader` reads current state at N plus event page through M ≤ N; it cannot claim historical projection; it cannot mutate gates, start work or acknowledge alerts globally.

`internal/execution` derives control mode from persisted control events, checks active run/fence/captured subjects, validates numerator/denominator/basis/accounting revisions and emits bounded liveness. Domain service constructors inject these ports; runners expose enough lifecycle/control information to heartbeat during slow work and checkpoint safely. Explanations are deterministic factual templates with attributed evidence-backed proposals, not prompt or chain-of-thought storage. Existing operation IDs serve as execution run IDs; verifier runs remain separate entities.

Current durable service records map to v2 approval/resolution/model/decision/profile/causality/attestation/scope/validity schemas. Process-private handles do not replace these records. Supervisor/resource ownership and fenced writes follow RECOVERY; serial project mutation is the initial revision model.

## Required adapter ports (conceptual contracts, not implementation)

| Port operation | Inputs | Output / refusal |
| --- | --- | --- |
| SourceAdapter.Detect | Sealed snapshot + bounded detection profile | Evidence-backed proposed facts/assertions separated; no runtime execution |
| SourceAdapter.Parse | Snapshot manifest + pinned parser capability | Neutral Source IR with spans/status/unresolved nodes |
| SourceAdapter.Extract | Source IR + operational profile | Application IR with resolved/unknown entrypoints/dependencies |
| SourceAdapter.PrepareInstrumentation | Disposable source + approved capture profile | Pinned probe/closure capability manifest; no source-tree mutation |
| TargetAdapter.ValidateCapabilities | v2 contracts + exact resolution + proposed target model | Supported operator/lowering checks or typed gaps |
| TargetAdapter.Generate | Current approved unit/model + owned staging root + pinned inputs | Patch/build requirement artifacts, never acceptance verdict |
| TargetAdapter.PrepareHarness | Exact build/environment + supported profile | Fixed runner specification and observable dimension capabilities |

Observer/runner owns startup/capture/cleanup and verifier owns comparison; adapter generation cannot set PASS. Registry lookups resolve approved IDs/manifest/content, not latest versions. Schema v2 adapter manifests distinguish distribution tier from validation_maturity and exact implementation_ref; PLANNED cannot be RUNTIME_VALIDATED. Unsupported port capability returns a named blocker.

## Shared assessment checkpoint services

`assess`/`assessment.run` coordinate existing guarded analysis/observation/behavior/feasibility services with a selected scope; no assessment analyzer, new adapter port or duplicate IR is introduced. A failed/unsupported evidence-acquisition dimension creates a diagnostic finding and may still publish a useful COMPLETED assessment operation/report. Storage/event failure that prevents durable publication cannot report completion. `assessment show`/`assessment.get` read immutable assessed-state projection plus separately evaluated current validity. They do not synthesize a plan or execute source code.

`assessment decide` derives trusted human action metadata and persists immutable selection, index and event atomically with project revision checks; it never writes user_confirmations or implementation missions. Omitted nullable/list decision input fields become explicit null/empty values in the engine-produced selection; CONTINUE_PLANNING requires exact selected scope and intent. Imports, assessment refs and linked analysis/dependency/evidence resolve under the same project/secret controls. A current explicit CONTINUE_PLANNING selection is an additional prerequisite of `plan`/`migration.plan`, not Gate 2 itself. Planning publishes its immutable assessment delta/reuse witnesses and hashes selection+basis refs in the exact PLAN subject. Semantic checks enforce same project, report/selection/basis equality, latest effective human continuation, current source/target/scope and compatible chosen units. An assessment result or selection alone never satisfies unit admission.
