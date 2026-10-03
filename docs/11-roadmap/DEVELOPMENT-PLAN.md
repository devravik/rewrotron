# Rewrotron development implementation plan

**Status:** Dependency-complete engineering specification; implementation unstarted. **Baseline:** 2026-10-03.

**Goal:** Build a safe local engine that characterizes a real CI3 application, reconstructs supported behavior slices in Laravel 13, and reports reproducible compatibility evidence without guessing unresolved facts.

**Architecture:** Three neutral IR tiers, injected source/target adapters, shared gate-checked services for CLI/MCP, SQLite WAL and immutable artifacts. Runtime applications execute in isolated independent environments. Agents provide optional proposals; deterministic verification decides results.

**Technology:** Go 1.22+ constitutional floor with a supported compatible toolchain pinned in RWT-001; database/sql + modernc SQLite; standard-library CLI parsing; VKCOM PHP parser inside source adapter; draft 2020-12 schemas; MCP 2025-11-25 stdio; CI3 3.1.13 reference fixture; Laravel 13/PHP 8.3+ target.

## How to execute

Use one task at a time in dependency order. Do not mark a task complete because code compiles; save the specified evidence. Do not create code until its prerequisites and owning specifications are read. The repository currently has no Go implementation, so the first Go test commands become available in RWT-001.

Each task is a reviewable subsystem deliverable with concrete fixtures and negative cases. Individual implementation steps are checkboxes; commit only the reviewed task files using `feat(<subsystem>): <deliverable>`. Parallel execution is optional only for dependency-independent tasks with separate file ownership; it never bypasses gates/fences. This plan does not require a particular agent runner or unavailable superpowers plugin.

## Global constraints

- Framework-specific logic stays in adapters; core/behavior/verification have no framework imports, provider SDKs or prompts.
- No unrestricted MCP SQL/shell/write tools, source edits, production writes or mandatory cloud/LLM dependency.
- Gate approvals are persisted exact human subjects; source drift invalidates the baseline; plan/release changes invalidate affected approvals.
- UNKNOWN never passes; only current proven DEAD can be automatically omitted from target reconstruction.
- Every CLI/MCP change updates its owning schema/registry; every deployed SQLite change adds an incremental migration and upgrade fixtures.
- Use explicit dependency injection, typed/wrapped errors, exported doc comments, gofmt, strict lint and race checks.
- Record an ADR for changed architecture and maintain spec/code parity. No additional architectural approval is implied beyond the repository owner's authorized documentation work.

## Shared interfaces and records

Operational request/handle fields are fixed in [SERVICE-CONTRACTS](../01-architecture/SERVICE-CONTRACTS.md). `internal/domain/types.go` defines record structs matching the canonical schemas in `specs/`. ProjectID, CaseID, UnitID, MissionID, ArtifactRef and Hash are distinct string types with constructors rejecting invalid schema values. SourceIR, ApplicationIR, BehaviorIR, Project, DetectionResult, MigrationPlan, TargetConfig, GenerationResult, VerificationHarness and RuntimeTraceSet are the records consumed by ports in SYSTEM-DESIGN.

Service request structs use the field names/types from their CLI/MCP input schemas plus trusted constructor/session context (`Actor`, `Channel`, `LeaseOwner`) and captured subject hashes. No later task invents a second payload shape. Shared result records map to operation/result, verification/report, project/snapshot, mission and report schemas. EnvironmentHandle/Workspace/Store/Tx are infrastructure handles with explicit Close/context behavior; no global mutable registry.

Service constructors take repository, artifact, adapter registry, clock, ID generator and runner ports. Clock exposes `Now() time.Time`; ID generator exposes `NewID(prefix string) string`; adapter registry resolves declared IDs/versions/capabilities. A fake clock/ID generator supplies deterministic unit fixtures. Concrete constructors are wired only by `internal/runtime`.

CompareAndPublish runs in a repository transaction: validate current project revision, baseline/plan hashes and live fence, register artifacts, update projections and append events. Long parsing/generation/replay runs outside this transaction. Artifact paths resolve only through registered project references.

Domain structural validation handles cross-references/graph constraints that JSON Schema cannot prove: duplicate IDs, same-project/baseline ownership, current gate subjects, DAG/step bounds, policy rule relations and corpus denominator reconciliation. Every such rule has a negative fixture in the owning task.

## Existing and future files

Documentation/schema/migration files listed in tasks already exist when present in this baseline; Go/runtime/testdata paths are planned creates. Do not assume directories contain implementation. The exact module import path is chosen from repository ownership in RWT-001 and recorded in go.mod; internal package names do not depend on a hosting vendor.

## Dependency checkpoints

| Checkpoint | Tasks | Demonstrable result |
| --- | --- | --- |
| M0 — checked specification | Current documentation/schema validation | Consistent contracts and explicit validation register; no runtime claim |
| M1 — safe engine foundation | 001–006 | Init/detect, human Gate 1, drift/refusal, persisted execution events/state, honest live status/watch and human control |
| M2 — measured source behavior | 007–013 | Framework extraction, safe repeated traces, cases and known comparator counterexamples |
| M3 — first independent slice | 014–016 | Exact Gate 2, staged Laravel slice, PASS/MISMATCH/UNKNOWN and restart recovery |
| M4 — reference application evidence | 017–022 | Auth/session, CRUD/effects/files, full scope report, benchmark and Gate 3 package |
| M5 — publishable reference release | 023–025 | Optional learning, two clients, native platform/security/release evidence |
| M6 — extension validation | 026–027 | New adapter consumes unchanged canonical model with tested capability limits |

## Task protocol

For each task, first add the listed observable and negative fixture assertions and run the focused command to establish missing behavior. Implement the described contract, run the same checks plus applicable dependencies, review schema/docs parity and save evidence. Until implementation exists, a missing package/binary is an expected unimplemented baseline, not a passing check.

### RWT-001: Build baseline and deterministic domain records

**Dependencies:** None. **Owner role:** Core maintainer. **Files:**

- `go.mod`
- `go.sum`
- `Makefile`
- `.golangci.yml`
- `.github/workflows/ci.yml`
- `internal/domain/types.go`
- `internal/domain/errors.go`
- `internal/domain/types_test.go`
- `tools/architecture/check.go`

**Interface contract:** Domain IDs/hashes, lifecycle/verdict enums and record structs map exactly to specs/common, project, IR, behavior, migration and verification schemas. Internal packages exchange these types; concrete parser/MCP types never cross ports.

**Implementation:** Choose a supported stable Go toolchain satisfying dependency floors; pin it, modernc SQLite, the PHP parser and schema validator with checksums. Record commands/tool versions. Create build/test/lint/schema targets and import-boundary checks. Construct CLI/MCP entrypoints with version metadata and signal cancellation but no fake implemented commands.

**Required failure/counterexample fixtures:** Unknown enum/version, invalid hash, cross-framework imports into neutral packages, provider SDK in deterministic services, missing exported doc comments.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./...; go vet ./...; make lint; make test-schemas; make build`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-002: Transactional SQLite runner and repository primitives

**Dependencies:** 001. **Owner role:** Storage maintainer. **Files:**

- `internal/storage/sqlite/store.go`
- `internal/storage/sqlite/migrations.go`
- `internal/storage/sqlite/store_test.go`
- `internal/storage/sqlite/migrations_test.go`
- `migrations/0001_initial.sql`
- `testdata/storage/`

**Interface contract:** NewStore(ctx context.Context, path string) (*Store, error); Store.WithTx(ctx context.Context, fn func(Tx) error) error; Store.Close() error. Tx provides bound domain repository operations; no public raw SQL.

**Implementation:** Configure every connection for foreign keys, bounded busy handling and full synchronization; coordinate WAL initialization. Create/verify schema_migrations checksums and apply numbered migrations atomically. Implement project revisions, same-project relationship validation and append-only repositories from SQLITE-SCHEMA.

**Required failure/counterexample fixtures:** Malformed JSON, invalid enums/spans, wrong-project references, update/delete immutable records, newer DB, changed checksum, failed migration transaction, busy/cancellation, close error.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/storage/sqlite -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-003: Redaction and durable content-addressed artifacts

**Dependencies:** 001,002. **Owner role:** Storage/security maintainer. **Files:**

- `internal/observer/redaction.go`
- `internal/observer/redaction_test.go`
- `internal/storage/artifacts/store.go`
- `internal/storage/artifacts/store_test.go`
- `testdata/security/`

**Interface contract:** Redactor.Redact(ctx context.Context, input CaptureInput) (RedactedCapture, error); ArtifactStore.Put(ctx context.Context, projectID ProjectID, capture RedactedCapture) (ArtifactRef, error); Get(ctx context.Context, projectID ProjectID, ref ArtifactRef) ([]byte, error). CaptureInput/RedactedCapture carry bytes, format and redaction/withheld-dimension manifest.

**Implementation:** Implement structured redaction, configured secret paths and project-keyed sensitive correlation. Stage only sanitized bytes; hash/fsync/rename and register links transactionally. Verify existing objects and ownership; reject unknown formats when safe persistence cannot be established.

**Required failure/counterexample fixtures:** Token in URL/SQL/stderr/body, malformed payload, raw staging write, low-entropy unkeyed secret hash, path/symlink escape, digest corruption, crash before/after rename, cross-project artifact.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/observer ./internal/storage/artifacts -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-004: Workspace, configuration and baseline manifests

**Dependencies:** 001,002,003. **Owner role:** Core maintainer. **Files:**

- `internal/project/config.go`
- `internal/project/workspace.go`
- `internal/project/manifest.go`
- `internal/project/manifest_test.go`
- `internal/project/config_test.go`
- `testdata/projects/`

**Interface contract:** OpenWorkspace(ctx context.Context, path string) (*Workspace, error); BuildManifest(ctx context.Context, root string, exclusions []string) (SourceManifest, error); CanonicalHash(value any) (Hash, error). SourceManifest records path/content/executable/symlink identity from CONFIRMATION-PROTOCOL.

**Implementation:** Validate strict YAML intent through project schema, reject duplicate keys/approval flags, allow absent target before planning, and create owner-only workspace. Implement canonical JSON golden vectors and complete dirty/untracked inventory. Preserve overrides separately; do not execute source.

**Required failure/counterexample fixtures:** Untracked/dirty file, content change during scan, volatile exclusions hiding code, unreadable file, out-of-root link, path case/Unicode, target overlap, duplicate YAML key, noncanonical numeric hash input.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/project -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-005: Gate-checked services and explicit human confirmations

**Dependencies:** 002,003,004. **Owner role:** Core/security maintainer. **Files:**

- `internal/core/gates.go`
- `internal/core/confirmations.go`
- `internal/core/gates_test.go`
- `internal/core/confirmations_test.go`
- `internal/domain/ports.go`

**Interface contract:** GateService.Require(ctx context.Context, projectID ProjectID, gate GateType, expected Hash) error; PrepareSource(ctx context.Context, req PrepareSourceRequest) (PreparedConfirmation, error); Confirm(ctx context.Context, req ConfirmationRequest) (Confirmation, error). Request fields map to CLI/MCP confirmation inputs; authenticated Actor/Channel are trusted constructor context.

**Implementation:** Compute configuration/plan/release subjects, merge whitelisted overrides, issue local human tokens, and transactionally consume tokens with approval/event/state update. Enforce current parent hashes at every service entry/publication. Keep immutable approvals on drift; expected revisions prevent concurrent advancement.

**Required failure/counterexample fixtures:** Missing gate, silence/YAML spoof, agent actor, stale manifest/plan, changed override, expired/wrong-project token, reused idempotency key with different payload, crash/concurrent approval, dry-run unlocking.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/core -run "Gate|Confirm|Drift" -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-006: CLI foundation, detection and doctor

**Dependencies:** 004,005. **Owner role:** Interface/adapter maintainer. **Files:**

- `internal/cli/parse.go`
- `internal/cli/envelope.go`
- `internal/cli/commands.go`
- `internal/cli/cli_test.go`
- `internal/runtime/container.go`
- `adapters/source/php/codeigniter3/detect.go`
- `adapters/source/php/generic_php/detect.go`
- `testdata/contracts/cli/`

**Interface contract:** CLI.Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int; SourceAdapter.Detect(ctx context.Context, rootPath string) (DetectionResult, error). Registration injects SourceAdapter/TargetAdapter and storage into core; CLI never owns gate logic.

**Implementation:** Implement init/detect/doctor/status/confirmation handlers from registry, standard flags and exact envelopes/errors. Bounded detection reads framework/runtime/dependency metadata without executing PHP. Doctor reports checks/remediation for each requested phase and refuses unsafe weaker runtime fallback. Expose only implemented commands; registry support status is explicit during development.

**Required failure/counterexample fixtures:** Invalid/duplicate flag, stdout contamination, quiet+verbose, no workspace, conflicting detection, unknown runtime, malicious PHP config, source/target overlap, --auto-confirm, mismatched envelope/exit, cancellation.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/cli ./internal/runtime ./adapters/source/php/... -count=1; make build`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-007: Synthetic fixture provenance and PHP Source IR extraction

**Dependencies:** 006. **Owner role:** Source adapter maintainer. **Files:**

- `adapters/source/php/parser/parser.go`
- `adapters/source/php/parser/parser_test.go`
- `adapters/source/php/codeigniter3/analyze.go`
- `tools/fixtures/manifest.json`
- `testdata/ci3-fixtures/routing/`
- `testdata/ci3-fixtures/dynamic/`

**Interface contract:** SourceAdapter.Analyze(ctx context.Context, project Project) (SourceIR, error); PHPParser.Parse(ctx context.Context, path string, source []byte, syntaxVersion string) (ParsedFile, error). ParsedFile is adapter-private; SourceIR is the canonical schema.

**Implementation:** Pin source archive/license/checksum, source runtime image digest and DB settings. Integrate the chosen Go parser behind a neutral record conversion. Emit complete symbols/spans/typed edges and explicit partial/unsupported findings for unsupported syntax and resource limits.

**Required failure/counterexample fixtures:** PHP 5/7/8 syntax fixtures, newer unsupported syntax, duplicate symbol, invalid span, partial parse, dynamic include/method, cancellation/limit, parser AST leaking into domain.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./adapters/source/php/... -count=1; make test-schemas`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-008: Application extraction, entrypoint roots and classifications

**Dependencies:** 007. **Owner role:** Analysis/source adapter maintainer. **Files:**

- `adapters/source/php/codeigniter3/extract.go`
- `adapters/source/php/codeigniter3/routes_test.go`
- `adapters/source/php/generic_php/extract.go`
- `internal/analysis/graph.go`
- `internal/analysis/classification.go`
- `internal/analysis/classification_test.go`

**Interface contract:** SourceAdapter.Extract(ctx context.Context, source SourceIR) (ApplicationIR, error); Classify(ctx context.Context, graph ApplicationGraph, evidence ObservationIndex) (ClassificationSet, error). Records follow IR-CONTRACTS and CODE-CLASSIFICATION.

**Implementation:** Resolve CI3 verb/convention/regex/precedence/default/404/_remap/custom base/loader/hook semantics in the adapter. Merge explicit CLI/cron/event roots. Build neutral reachability and prove omission only for closed graphs plus zero observed hits; retain unresolved candidates and source resources.

**Required failure/counterexample fixtures:** Unreachable cycle with inbound calls, cron-only root, dynamic target missed statically, no runtime corpus, mixed ACTIVE file, view asset, incomplete parser/root inventory, false verb restriction.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/analysis ./adapters/source/php/... -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-009: Isolated container runner and resettable environments

**Dependencies:** 003,006,007. **Owner role:** Runtime/security maintainer. **Files:**

- `internal/sandbox/runner.go`
- `internal/sandbox/environment.go`
- `internal/sandbox/runner_test.go`
- `tools/fixtures/compose.yaml`
- `testdata/security/egress/`

**Interface contract:** SandboxPort.Prepare(ctx context.Context, profile SandboxProfile, fixture ArtifactRef) (EnvironmentHandle, error); Reset(ctx context.Context, handle EnvironmentHandle) error; Close(ctx context.Context, handle EnvironmentHandle) error. Profile records digest-pinned runtimes, quotas, allowed volumes, network and sinks.

**Implementation:** Create independent source/target state and internal mock network, enforce non-root/read-only/drop-capability restrictions, verify reset hashes and label resource ownership. Separate controlled dependency acquisition from runtime. Implement cancellation/process-tree cleanup and restart reconciliation.

**Required failure/counterexample fixtures:** Public DNS/IP/redirect/raw socket egress, Docker socket/host mounts, unsafe local-mode customer run, missing cache, failed restore, timeout cleanup leak, reused mutable DB.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/sandbox -count=1; make test-fixtures`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-010: Proxy capture and CI3 probe completeness

**Dependencies:** 008,009. **Owner role:** Observer/source adapter maintainer. **Files:**

- `internal/observer/capture.go`
- `internal/observer/capture_test.go`
- `adapters/source/php/codeigniter3/observe.go`
- `tools/php/ci3-probe/`
- `testdata/ci3-fixtures/hooks-effects/`

**Interface contract:** SourceAdapter.Observe(ctx context.Context, project Project, options ObserveOptions) (RuntimeTraceSet, error). ObserveOptions carries scenario/case, fixture, limits and EnvironmentHandle; trace fields match trace-schema.

**Implementation:** Overlay probes only in sandbox copies; preserve hook order and real output capture. Correlate dispatch/session/DB/cache/files/sink receipts. Emit COMPLETE/PARTIAL/UNAVAILABLE/NOT_APPLICABLE per dimension and redaction loss. Compare instrumented and control source runs.

**Required failure/counterexample fixtures:** Original source modification, cached response, constructor failure/exit, unsupported DB/session driver, truncated capture, observer-induced output, blocked unmatched external call, secret persistence.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/observer ./adapters/source/php/... -count=1; make test-fixtures`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-011: Cases, workflows and golden masters

**Dependencies:** 008,010. **Owner role:** Behavior maintainer. **Files:**

- `internal/behavior/contracts.go`
- `internal/behavior/workflows.go`
- `internal/behavior/golden.go`
- `internal/behavior/behavior_test.go`
- `testdata/compatibility-cases/`

**Interface contract:** BehaviorService.Discover(ctx context.Context, projectID ProjectID) (BehaviorIR, error); GoldenMaster.Build(ctx context.Context, scenario ObservationScenario, repeats RuntimeTraceSet) (Corpus, error). Records map to behavior/scenario/IR schemas; unresolved rules remain explicit.

**Implementation:** Implement `case show` and safe registered `artifact read` CLI evidence lookup. Validate customer input-only scenarios, import operational declarations, execute three reset source repeats, synthesize concrete assertions and candidate variance. Bind workflow variables explicitly and preserve per-side continuity. Lock only evidenced contracts/cases and report unresolved computations.

**Required failure/counterexample fixtures:** Invented business rule, candidate automatically normalized, wrong workflow user, source token reused on target, fixture not reset, missing effect dimension, cyclic unbounded scenario, missing negative auth cases.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/behavior -count=1; make test-schemas; make test-fixtures`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-012: Deterministic comparators and logical DB identity

**Dependencies:** 003,011. **Owner role:** Verification maintainer. **Files:**

- `internal/verification/comparator.go`
- `internal/verification/json.go`
- `internal/verification/normalization.go`
- `internal/verification/database.go`
- `internal/verification/comparator_test.go`
- `testdata/compatibility-cases/`

**Interface contract:** Comparator.Compare(ctx context.Context, source ObservationTrace, target ObservationTrace, policy ComparisonPolicy) (CaseResult, error). ComparisonPolicy records approved operators/paths/validators/relations/exceptions; CaseResult maps to verification/report.schema.

**Implementation:** Implement EXACT, semantic JSON without float precision loss, scoped HTML assertions, multivalue headers/cookies, typed committed DB rows and approved bijective key mapping, state/effect comparisons and deterministic verdict precedence. Preserve per-dimension results and approved-change disclosure.

**Required failure/counterexample fixtures:** Amount/type/null/missing/array-order changes, duplicate JSON key, collapsed cookies, wrong principal/token relation, swapped IDs, missing trigger/audit write, ambiguous rows, HTML whitespace-sensitive change, missing capture, out-of-window time.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/verification -run "Compare|Normalize|Database" -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-013: Paired replay and immutable run reporting

**Dependencies:** 009,010,011,012. **Owner role:** Verification maintainer. **Files:**

- `internal/verification/replay.go`
- `internal/verification/replay_test.go`
- `internal/core/verification.go`
- `internal/cli/verification.go`
- `testdata/contracts/cli/verify/`

**Interface contract:** ReplayService.Run(ctx context.Context, request VerificationRequest) (VerificationReport, error). Request fields are project/case-or-unit-or-all, captured subjects and limits; report uses canonical schemas and artifact references.

**Implementation:** Execute cases against independent restored source/target state, capture each workflow side separately, compare all required dimensions, recheck drift/fences and publish a run transactionally. Implement correct operational error versus completed mismatch exits and full-corpus aggregation.

**Required failure/counterexample fixtures:** Shared DB/cookie jar, redirect loop, source outage, target timeout, cancel during capture, stale target/source/policy, missing report artifact, empty corpus falsely 100%, mismatch hidden by aggregate.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/verification ./internal/core ./internal/cli -count=1; make test-fixtures`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-014: Capability registry, unit DAG and exact plan approval

**Dependencies:** 005,008,011. **Owner role:** Migration maintainer. **Files:**

- `internal/migration/planner.go`
- `internal/migration/capabilities.go`
- `internal/migration/planner_test.go`
- `adapters/target/php/laravel13/plan.go`
- `internal/cli/plan.go`

**Interface contract:** TargetAdapter.Plan(ctx context.Context, behavior BehaviorIR, config TargetConfig) (MigrationPlan, error); Planner.Validate(ctx context.Context, plan MigrationPlan) error. Adapter manifest is versioned; plan maps to migration/plan.schema.

**Implementation:** Implement `plan show` for CLI-only evidence access. Partition observable units, condense dependency cycles, stable-sort the DAG, assign shared-file ownership and verify capability coverage. Include omissions, dormant/unknown disposition, scope exclusions, corpus/policy/runtime/version hashes. Gate 2 confirms exact subject; target choices never enter confirmation as unnoticed edits.

**Required failure/counterexample fixtures:** Cycle edge silently removed, unsupported computation/session/DB, overlapping target/source, hidden cron exclusion, stale omission proof, target preference changed at confirmation, missing required cases.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/migration ./internal/core ./adapters/target/php/laravel13 -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-015: Target staging, patch journal and recovery

**Dependencies:** 003,005,014. **Owner role:** Migration/storage maintainer. **Files:**

- `internal/migration/publication.go`
- `internal/migration/publication_test.go`
- `internal/cli/patches.go`
- `internal/core/resume.go`
- `internal/core/resume_test.go`
- `testdata/storage/checkpoints/`

**Interface contract:** Publication.Apply(ctx context.Context, patch PatchManifest, lease MissionLease) (TargetTreeHash, error); Reconcile(ctx context.Context, projectID ProjectID) (RecoveryReport, error). PatchManifest contains safe paths/actions/before-after digests and unit/fence.

**Implementation:** Implement target-only imported patch stage/apply commands and MCP service requests as specified in PROJECT-INPUTS. Stage all generation, journal publication intent, check gates/before hashes/fence, atomically replace individual owned files and verify the tree. Recover every crash point and conflict; persist completion only after all files are consistent. Resume reconciles resources and artifacts before replay.

**Required failure/counterexample fixtures:** Crash after each file, changed human file, symlink path escape, overlapping workers, lease lost mid-write, stale source/plan, rollback clobbering user edits, orphan artifacts, newer DB during resume.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/migration ./internal/core -run "Publish|Recover|Resume|Conflict" -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-016: Laravel bootstrap and deterministic read slice

**Dependencies:** 013,014,015. **Owner role:** Target adapter maintainer. **Files:**

- `adapters/target/php/laravel13/generate.go`
- `adapters/target/php/laravel13/harness.go`
- `adapters/target/php/laravel13/generate_test.go`
- `testdata/ci3-fixtures/read-json/`
- `testdata/laravel-fixtures/read-json/`

**Interface contract:** TargetAdapter.Generate(ctx context.Context, plan MigrationPlan, outputDir string) (GenerationResult, error); SetupVerification(ctx context.Context, project Project) (VerificationHarness, error). OutputDir is staging; GenerationResult includes patch manifest/provenance and build diagnostics.

**Implementation:** Pin Laravel/PHP/Composer tool locks, bootstrap only after Gate 2, and implement a deterministic authenticated/unauthenticated JSON read unit from declarative Behavior IR. Generate standalone target README/env example/compatibility test and verify independently without Rewrotron runtime.

**Required failure/counterexample fixtures:** Gate 2 missing, implicit middleware changes status, invented rule, source framework import in generated target, unsupported cast, dependency script egress, target requires project.db, deliberate JSON/status mutation.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./adapters/target/php/laravel13 -count=1; make test-e2e`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-017: Authentication, cookies and active session continuity

**Dependencies:** 016. **Owner role:** Target/auth maintainer. **Files:**

- `adapters/target/php/laravel13/auth.go`
- `adapters/target/php/laravel13/session.go`
- `adapters/target/php/laravel13/auth_test.go`
- `testdata/ci3-fixtures/auth-session/`
- `testdata/laravel-fixtures/auth-session/`

**Interface contract:** Target capability manifest gains evidenced password/session/CSRF profiles; canonical contracts/cases remain framework-neutral. Unsupported formats produce ACTIVE_SESSION_CONTINUITY_UNPROVEN finding and UNKNOWN required dimension.

**Implementation:** Implement actual detected password verification and safe existing-session continuity strategy in the target fixture. Preserve roles/permissions, flash duration, cookie flags/scope, rotation, expiry, CSRF and logout. Require source-session→target and rollback continuity evidence.

**Required failure/counterexample fixtures:** Wrong principal, expired/tampered/fixation token, missing permissions, CSRF invalid/replay accepted, encrypted-cookie assumption, unsafe deserialization, flash lifetime changed, re-login silently substituted.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `make test-e2e; go test -race ./adapters/target/php/laravel13 ./internal/verification -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-018: Transactional CRUD workflows, views, files and safe effects

**Dependencies:** 017. **Owner role:** Target/verification maintainers. **Files:**

- `adapters/target/php/laravel13/rules.go`
- `adapters/target/php/laravel13/views.go`
- `adapters/target/php/laravel13/effects.go`
- `testdata/ci3-fixtures/crud-workflow/`
- `testdata/ci3-fixtures/views-files/`
- `testdata/laravel-fixtures/`

**Interface contract:** Adapter manifest gains only tested declarative calculations/transactions/view/file/effect capabilities. Generated rules consume contract IDs/evidence and schema-aware models; no source-file transpilation.

**Implementation:** Implement create/read/update failure/success workflow, exact decimal/null/timezone/validation/transaction semantics, source-compatible forms/assets/upload/download and HTTP/mail mock behavior. Verify ordered effect counts and generated-ID relation rules across steps.

**Required failure/counterexample fixtures:** Skipped audit write, wrong FK ID, partial transaction, null/coercion drift, authorization after write, duplicate payment/email sink, removed input, changed binary download, unsupported renderer/file format.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `make test-e2e; make test-fixtures; go test -race ./internal/verification ./adapters/target/php/laravel13 -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-019: Fenced missions, experiment history and repair loop

**Dependencies:** 005,013,015,018. **Owner role:** Core/agent interface maintainer. **Files:**

- `internal/missions/scheduler.go`
- `internal/missions/leases.go`
- `internal/missions/missions_test.go`
- `internal/core/repair.go`
- `internal/core/repair_test.go`
- `internal/cli/missions.go`

**Interface contract:** MissionService.Peek(ctx context.Context, projectID ProjectID) (NextMissionResult, error); Claim(ctx context.Context, req ClaimRequest) (MissionLease, error); Renew(ctx context.Context, req RenewRequest) (MissionLease, error); Complete(ctx context.Context, req CompleteRequest) (Mission, error). Request fields map to mission schemas and trusted session owner.

**Implementation:** Implement deterministic readiness/ordering, five-minute leases and one-minute renewals, fences and three-attempt unchanged-context limit. Record fingerprinted hypotheses/patches/outcomes, queue repair or observation by verdict, validate completion evidence and never invoke a provider in core.

**Required failure/counterexample fixtures:** Two concurrent claims, expired owner completing, stale fence publishing, free-text false completion, repeated failed fingerprint, missing prerequisite, empty queue, failed patch/customer conflict, UNKNOWN closed as PASS.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/missions ./internal/core ./internal/cli -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-020: MCP stdio server and schema-complete tool/resources

**Dependencies:** 006,019. **Owner role:** Interface maintainer. **Files:**

- `internal/mcp/server.go`
- `internal/mcp/tools.go`
- `internal/mcp/resources.go`
- `internal/mcp/server_test.go`
- `cmd/rewrotron-mcp/main.go`
- `testdata/contracts/mcp/`

**Interface contract:** MCPServer.Serve(ctx context.Context, input io.Reader, output io.Writer, diagnostics io.Writer) error. Handlers use the same services/requests as CLI and input/output refs from specs/mcp/tools.json.

**Implementation:** Pin a compatible Go MCP SDK or implement the explicitly required stdio subset; record the choice/version after protocol fixtures pass. Implement lifecycle, tool/resource listing/calls, structuredContent/domain errors, progress/cancellation and human-token confirmation. Validate every registry schema and permission.

**Required failure/counterexample fixtures:** Before initialized call, malformed/oversized message, newline framing, unsupported negotiated revision, unknown tool, arbitrary artifact path, actor spoof, approval mint through MCP, mutation marked read-only, cancellation/stdout pollution.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/mcp -count=1; make test-schemas; make build`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-021: Reproducible benchmarks and measured acceptance

**Dependencies:** 013,018. **Owner role:** Performance/verification maintainer. **Files:**

- `internal/benchmark/runner.go`
- `internal/benchmark/statistics.go`
- `internal/benchmark/benchmark_test.go`
- `internal/cli/benchmark.go`
- `tools/fixtures/workload.json`

**Interface contract:** BenchmarkService.Run(ctx context.Context, request BenchmarkRequest) (BenchmarkReport, error). Request names project/unit corpus, approved workload/quotas and seed; report records trial ratios, confidence bounds, acceptance and current subjects.

**Implementation:** Implement the paired five-trial protocol, fixed warm-up/measurement, p95/memory/throughput/errors/query diagnostics and deterministic 10,000-sample bootstrap. Require behavioral corpus equivalence under workload. Block inconclusive required metrics instead of inventing a tolerance.

**Required failure/counterexample fixtures:** No memory probe, zero baseline denominator, insufficient successes, noisy inconclusive interval, known regression, wrong quota/cache, skipped behavior yielding fake speedup, stale target hash.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/benchmark -count=1; make benchmark-fixtures`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-022: Coverage/report, release subject and Gate 3 package

**Dependencies:** 005,018,020,021. **Owner role:** Release/core maintainer. **Files:**

- `internal/core/reports.go`
- `internal/core/cutover.go`
- `internal/core/reports_test.go`
- `internal/cli/reports.go`
- `testdata/release/`

**Interface contract:** ReportService.Build(ctx context.Context, projectID ProjectID) (MigrationReport, error); CutoverService.Prepare(ctx context.Context, releaseHash Hash, outputRoot string) (ReleasePackage, error). Report/release schemas bind every required hash/evidence/runbook.

**Implementation:** Reconcile all discovered entrypoints, compute numerator/denominator/null percentages, disclose omissions/exclusions/expected changes, and build release/session/rollback evidence manifest. Require supplied staging-tested operator commands and Gate 3. Package only approved artifacts, independent target and runbook; no live deployment.

**Required failure/counterexample fixtures:** 99% corpus, empty denominator, hidden unsupported cron, stale benchmark/target/runbook, missing rollback/session evidence, Gate 3 bypass, secret export, package path escaping approved root.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/core ./internal/cli ./internal/mcp -count=1; make test-e2e`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-023: Project lessons and privacy-reviewed global patterns

**Dependencies:** 019,022. **Owner role:** Knowledge/security maintainer. **Files:**

- `internal/knowledge/lessons.go`
- `internal/knowledge/promotion.go`
- `internal/knowledge/knowledge_test.go`
- `migrations/global/0001_initial.sql`
- `specs/knowledge/pattern.schema.json`

**Interface contract:** KnowledgeService.RecordLesson(ctx context.Context, request LessonRequest) (Lesson, error); Promote(ctx context.Context, request ReviewedPromotionRequest) (GlobalPattern, error). Lesson request maps to tool input; promotion carries sanitized manifest/hash and human approval, not raw project records.

**Implementation:** Record negative and positive lessons locally. Design/apply separate optional global DB migration with three-independent-context/evaluation evidence, sanitized data, exact opt-in approval and no global dependency for baseline operation. Update storage/global schemas and compatibility fixtures together.

**Required failure/counterexample fixtures:** Customer URL/credential/PII leak, three units falsely counted as three independent projects, no negative case, automatic export, absent global store blocking migration, changed promotion manifest.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./internal/knowledge -count=1; make test-schemas`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-024: Portable Skill and evidence-based capability evolution

**Dependencies:** 020,023. **Owner role:** Agent integration maintainer. **Files:**

- `skills/rewrotron/SKILL.md`
- `skills/rewrotron/references/`
- `skills/rewrotron/scripts/`
- `internal/knowledge/skills.go`
- `testdata/contracts/clients/`

**Interface contract:** Skill version/content hashes are experiment metadata. Skills search returns a typed local registry result; generated script activation requires fixed-runner capability metadata and reviewed version evidence.

**Implementation:** Write compact lifecycle/CLI/MCP references and fixed command helpers. Exercise the same migration through two actual clients plus CLI-only path, interrupted/resumed and gate-blocked sessions. Implement proposed/experimental/stable/deprecated tracking and evidence denominators; do not auto-activate generated tools.

**Required failure/counterexample fixtures:** Skill bypasses gate, hidden provider dependence, hallucinated maturity, zero-sample success rate, prompt/script unrestricted execution, deprecated content lost, client completion without engine evidence.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `make test-e2e; make test-schemas; go test -race ./internal/knowledge ./internal/mcp -count=1`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-025: Release compatibility, native platforms and security rehearsal

**Dependencies:** 022,024. **Owner role:** Release/security maintainer. **Files:**

- `.github/workflows/release.yml`
- `tools/release/manifest.json`
- `docs/10-development/RELEASE.md`
- `testdata/release/`
- `docs/11-roadmap/DECISION-REGISTER.md`

**Interface contract:** Published release manifest records binary/contract/adapter/schema/skill versions, checksums, signatures, SBOM/license notices and tested support matrix. Release approval is separate from customer migration Gate 3.

**Implementation:** Run full reference/negative/security/recovery suites, native target-platform smoke checks and independent target builds. Rehearse backup/upgrade/rollback/session behavior. Assign actual signing custody/security contact, document unsupported platforms/capabilities and promote adapters only on corpus evidence.

**Required failure/counterexample fixtures:** Cross-compile mislabeled native support, stale checksums, altered migration, missing license notice/contact/signing owner, skipped required runtime profile, secret artifact, false STABLE claim.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./...; make lint; make test-schemas; make test-fixtures; make test-e2e; make benchmark-fixtures; make build`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-026: Additional target adapter using unchanged canonical behavior

**Dependencies:** 025. **Owner role:** Target adapter maintainer. **Files:**

- `adapters/target/go/std_http/`
- `specs/adapters/`
- `testdata/target-fixtures/go/`
- `docs/01-architecture/ADR/`

**Interface contract:** Implement existing TargetAdapter ports and capability manifest; consume the same BehaviorIR/plan/corpus. A new adapter-specific engine or source-framework translator is prohibited.

**Implementation:** Start with Go std-http for the same read/auth/CRUD supported fixture subset. Record an ADR and runtime/dependency/capability matrix. Reuse comparators and explicitly report unsupported PHP coercion/session/DB semantics. Add new public choices only with schema/docs/compatibility updates.

**Required failure/counterexample fixtures:** Changed canonical IR solely for framework convenience, hidden source import, unsupported behavior generated, same JSON case with wrong types, missing session/transaction parity.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./...; make test-e2e; make test-schemas`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

### RWT-027: Additional source adapter with verified capability intersection

**Dependencies:** 025. **Owner role:** Source adapter maintainer. **Files:**

- `adapters/source/php/legacy_laravel/`
- `testdata/source-fixtures/legacy-laravel/`
- `specs/adapters/`
- `docs/01-architecture/ADR/`

**Interface contract:** Implement existing SourceAdapter Detect/Analyze/Extract/Observe ports and neutral records. The exact source family is selected through an explicit roadmap ADR before adding fixtures.

**Implementation:** Choose a customer-demanded source profile, pin redistribution-safe fixtures, emit the same source/application/behavior records, and verify the intersection with existing targets. Record unsupported syntax/extensions and complete root inventory; no core framework conditionals.

**Required failure/counterexample fixtures:** New source added without scoped demand/profile, core framework checks, false DEAD due to missing operational roots, canonical behavior extension without demonstrated unrepresentable behavior.

- [ ] Read the owning specs and verify prerequisites against their saved hashes/results.
- [ ] Add assertions for the observable result and every listed negative fixture; run the focused checks and record the absent behavior.
- [ ] Implement the specified interface/behavior and structural validation; format code and synchronize schemas/docs if its contract changes.
- [ ] Run `go test -race ./...; make test-fixtures; make test-e2e; make test-schemas`. Expected completion result: exit 0, required assertions executed, no runtime-profile skip represented as a pass.
- [ ] Review dependency/gate/privacy compatibility and save command output, artifact hashes, limitations and task completion record.
- [ ] Commit only the scoped files with an accurate Conventional Commit; mark this task complete after review evidence exists.

## Requirement traceability

| Requirement | Owning specification | Implementation tasks |
| --- | --- | --- |
| Neutral architecture/three IR tiers | SYSTEM-DESIGN, IR-CONTRACTS | 001,007,008,011,014,016,026,027 |
| Gate 1/2/3 and human authority | CONFIRMATION-PROTOCOL, PROJECT-LIFECYCLE | 004,005,006,014,020,022 |
| Drift/canonical hashes/idempotency | CONFIRMATION-PROTOCOL, RECOVERY | 002,004,005,013,015,019 |
| Safe omission and operational roots | ANALYSIS-SPEC, CODE-CLASSIFICATION | 007,008,011,014,022 |
| Runtime isolation/secrets/complete evidence | SECURITY, SANDBOX, SECRET-HANDLING | 003,009,010,013,025 |
| Contract/case/workflow and variability | BEHAVIOR-MODEL, GOLDEN-MASTER | 010,011,012,013 |
| DB/session/effect/HTML/file compatibility | COMPATIBILITY-SPEC, DATABASE-COMPARISON, AUTH-AND-SESSION | 012,013,017,018 |
| Staged generation and owned rollback | MIGRATION-ENGINE, RECOVERY | 014,015,016,019 |
| CLI/MCP machine contracts | CLI-SPEC, MCP-SPEC, ERROR-CATALOG | 001,006,013,019,020,022 |
| Lease/resume/repair evidence | MISSION-PROTOCOL, EXPERIMENT-LEDGER | 015,019,020 |
| Coverage/performance/cutover evidence | COVERAGE, PERFORMANCE, CUTOVER-RUNBOOK | 021,022,025 |
| Project/global privacy and skill evolution | LEARNING-SYSTEM, SKILL-SPEC | 023,024 |
| Migrations/releases/independent target | SQLITE-SCHEMA, RELEASE, TARGET-ARCHITECTURE | 002,015,016,022,025 |

## Blockers and validation sequencing

The [decision register](DECISION-REGISTER.md) names the remaining external validation inputs and their owners. RWT-001 resolves actual Go/dependency compatibility; RWT-007/009 resolve fixture/image/runtime availability; RWT-017 establishes actual session interoperability; RWT-020 tests actual client protocol support; RWT-021 measures actual performance; RWT-025 assigns publication credentials/contact. These are explicit task deliverables, not hidden assumptions and not evidence already achieved.

A customer migration additionally needs [PROJECT-INPUTS](../10-development/PROJECT-INPUTS.md). If a required input is unavailable, record the affected mission/unit and requested decision, continue unrelated safe work, and retain UNKNOWN/unverified coverage. Never claim no possible blocker: the plan eliminates undocumented engineering prerequisites while making external and application-specific uncertainty visible.

## First demonstration and completion evidence

The first demo detects the synthetic source, records Gate 1, analyzes/observes login and CRUD with independent fixtures, locks the corpus, proposes and records Gate 2, generates target slices, demonstrates deliberate mismatch and incomplete-capture UNKNOWN, repairs through fenced evidence, resumes after a crash, reruns the full corpus, benchmarks, assembles rollback/session evidence, records Gate 3 and writes a release package. Production is not modified.

A task completion record includes task ID, commit, owning specification, tests/profiles executed, exit status, current subjects, report/artifact references, unresolved limitations and reviewer. A milestone additionally reconciles all prerequisite records. A generated target must start from its locked dependencies/configuration without Rewrotron services or source-framework code.

## Target and dependency decision requirements (ADR 0012)

[TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) and [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) extend the existing prerequisite tasks rather than creating a competing queue:

| Task | Additional deliverable and required counterexamples |
| --- | --- |
| RWT-001 | Neutral target/recommendation/resolution types matching shared schemas; open language identifiers remain capability-checked |
| RWT-003/004 | Root intent/legacy/split loading and immutable decision artifacts; reject duplicate keys/target definitions; generated YAML approval spoof never grants authority |
| RWT-005/006 | Preserve exact user target; separate workspace intent drift from source drift; keep three gates and reviewed subject/token rules |
| RWT-007/008/010 | Exact source dependency inventory, usage/recursive graph and runtime capability evidence; dynamic usage and zero observations never imply DEAD |
| RWT-014 | `internal/target`, `internal/recommendations`, `internal/dependency` validation/feasibility/resolution; freeze target before planning; new plans include `decision_system`; missing metadata, removed APIs, transitive conflicts and missing capabilities block dependent scope |
| RWT-013/016/019 | Dependency characterization plus end-to-end regression; compatible metadata with changed behavior fails; user target edits during generation refuse publication; selective invalidation retains unrelated source IR and history without stale PASS |
| RWT-020/022 | Existing plan/artifact/report review exposes detected/recommended/requested/resolved values and decisions; no undocumented configure command/tool |
| RWT-023 | Capability-scoped resolution lessons with success/failure evidence; promotion never overrides configuration or replaces verification |

These neutral subsystem files are planned implementations, not created by this documentation update. Use existing migration/core/knowledge ownership instead of parallel planning/approval/learning engines. Run appropriate Go race, schema, gate, fixture and recovery checks in each implementation task; current documentation validation establishes only contract structure and examples.

## Transparent execution baseline (ADR 0013)

[EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) is a core deliverable in existing task order, not a later desktop workstream. Every long-running service must be inspectable before its task is complete.

| Task | Required observability work and evidence |
| --- | --- |
| RWT-001/002/003 | Canonical execution types, transactional event sink/reader, privacy and same-project validation; no extra ledger |
| RWT-004/005 | Stage/context/rationale and exact approval cards; approval-received events cannot forge authority |
| RWT-006 | `internal/execution/{events,state,progress,explanations,control}.go` and tests; status execution field, human watch, execution show/pause/continue; slow-action heartbeat and watch disconnect do not stop engine |
| RWT-007–014 | Measured discovery counts and evidence-backed explanations, concurrent/repeated stages, unknown denominators and scoped decision blockers |
| RWT-013/015/019 | Correct failed-action/verdict distinction; safe pause during replay/publication; restart while paused; no stale completion or duplicate effects after telemetry failure |
| RWT-020 | `project.get_execution`, execution resource and cursor reconnection; global sequence gaps, duplicates, watermarks and privacy; no required push transport |
| RWT-021/022 | Persisted progress basis/weights, current coverage and final readiness; 100% work never replaces Gate 3 |

RWT-006 exposes the foundation from the first long operation; later tasks instrument their real work and extend tested projections. Files are planned Go implementations. Required acceptance includes unknown/zero denominators, changed accounting scope, retry deduplication, per-client return-session deltas, missing-heartbeat freshness, paused readiness, stale approvals, forged events and meaningful error remediation. Run focused Go race/contract/recovery tests as those tasks are implemented. A future desktop only renders the same engine contract.
