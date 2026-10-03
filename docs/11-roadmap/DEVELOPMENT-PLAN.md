# Implementation sequence and milestone contracts

**Maturity: SPECIFIED.** This is the implementation task owner. It supersedes historical phase numbering; no milestone is complete. Roadmap and project plan reference this contract. All acceptance experiments in TESTING are planned unless recorded with actual execution evidence.

Execution order: M0 → M1 → M2 → M3 → M4 → M5 → M6; M7 optional after M6; M8 expansion after M6 (and M7 only where relevant). Do not use historical RWT numbering as a parallel implementation queue.

## M0 — Durable evidence and approval foundation

**Objective:** Durable evidence and approval foundation.

**Prerequisites:** None.

**Inputs:** Synthetic immutable source tree, schemas v2, local filesystem profile.

**Outputs:** Guarded CLI JSON, sealed source artifacts, project DB with migrations 0001/0002, canonical subject service, persisted assessment projection/selection records.

**Owned specs:** CONFIRMATION-PROTOCOL, ANALYSIS-SPEC, SQLITE-SCHEMA, ARTIFACT-STORE, CLI-SPEC.

**Implementation scope:** Go bootstrap/ports; exact driver/toolchain manifest; artifact/privacy/admission; snapshot acquisition; explicit human receipts; current execution/event-window service; assessment report publication over available canonical records and trusted human selection (no second analyzer).

**Tests:** X01 hash/restart/registry; X02 snapshot ABA/symlink; X25 ownership/leases; X28 storage/backup; X30 quota; X41/X43/X45/X46/X51/X53/X55 assessment stop/non-mutation/history/no authority.

**Acceptance criteria:** Reconstruct identical approved subjects from retained bytes after restart; reject stale/forged/missing approval and unsafe snapshot; preserve failure evidence; report completion is separate from migration feasibility and never creates Gate 2/implementation work.

**Evidence required:** Version/checksum manifest, exact command outputs, SQL/Go race results and subject vectors.

**Blockers:** Unsupported local FS; missing pinned SQLite build/version; unavailable canonical/ownership checks.

**Explicit non-goals:** No source runtime/adapters, generation, background survival or universal discovery.

## M1 — Falsifiable deterministic reconstruction experiment

**Objective:** Falsifiable deterministic reconstruction experiment.

**Prerequisites:** M0.

**Inputs:** One synthetic stateful application, independently reviewed two-actor bounded scenario and explicit mutant targets.

**Outputs:** Source/Application/Behavior IR subset, exact resolution, small target model, assessment checkpoint/selected-work planning basis, scoped verifier/recovery evidence.

**Owned specs:** BEHAVIOR-MODEL, RUNTIME-OBSERVATION, TARGET-ARCHITECTURE, MIGRATION-UNITS, DIFFERENTIAL-TESTING, RECOVERY.

**Implementation scope:** Small rule interpreter; minimal fixture source/target adapters; login→create operation with authorization failure; independent DBs and sink; foreground/supervisor ownership; current status; report partial state before planning and require explicit human continuation before fixture target planning.

**Tests:** X03 unsupported operator; X04 branch/state; X05 missing DB write; X06 wrong principal; X07 duplicate effect; X08 incomplete capture; X26 disconnect; X27 event failure; X42/X44/X47/X52/X54 blocked-but-useful assessment, production safety, shared-record reuse and classification.

**Acceptance criteria:** Good target passes UNIT+implemented regression; each wrong mutant fails for intended reason; login completes before create; incomplete capture yields UNKNOWN; interruption recovers without duplicate effects; repeat without LLM; the same fixture can stop successfully at assessment with blocked planning and later reuse valid shared findings after human continuation/Gate 2.

**Evidence required:** 17 premise checks listed below, exact pinned corpus/policy/subject/trace/reset/target identities, mutation kill report, repeated deterministic runs.

**Blockers:** Any false acceptance, unclosed required dimension, nonreconstructable hash, deadlocked units or unowned process.

**Explicit non-goals:** No broad CI3/Laravel completeness, autonomous repair, global synthesis, desktop, production cutover; read-only JSON is smoke only.

## M2 — Characterize real reference source safely

**Objective:** Characterize real reference source safely.

**Prerequisites:** M1.

**Inputs:** Pinned licensed CI3/PHP7.4/MySQL/Redis fixture with custom auth, ≥50 routes, cron/payment sink/uploads/legacy dependencies.

**Outputs:** Modified-vendor-aware Source/Application IR, mechanism attestations, source traces, bounded reviewed corpus and provenance-backed assessment (blocked/unknown feasibility is allowed).

**Owned specs:** ANALYSIS-SPEC, CODE-CLASSIFICATION, RUNTIME-OBSERVATION, WORKFLOW-DISCOVERY, AUTH-AND-SESSION.

**Implementation scope:** CI3/Generic PHP subset parser/dispatch adapter, independent container profile, operational boundary and supported connection/worker probes.

**Tests:** X09 dynamic dispatch/cron; X10 vendor/flags; X11 PDO/native; X12 async email; X13 unavailable extension; X14 DB reset; X42/X44/X48/X51/X54 assessment with runtime blockers, selective source drift and retrievable findings.

**Acceptance criteria:** Unresolved roots remain UNKNOWN; no false DEAD; direct PDO control prevents false COMPLETE; asynchronous receipt correlated; safe reset/egress verified; missing runtime/dependency evidence produces a useful classified report rather than forced plan completion.

**Evidence required:** Licensed fixture/source/image/lock digests; instrumented control comparison; reviewed role/state/critical workflow partitions.

**Blockers:** Unavailable runtime/native driver; unsafe sandbox; missing operational roots.

**Explicit non-goals:** No claiming complete CI3 support from one application; unsupported eval/streams explicit.

## M3 — Freeze feasible target decisions and implementation model

**Objective:** Freeze feasible target decisions and implementation model.

**Prerequisites:** M2.

**Inputs:** Persisted assessment, explicit human selected-work/CONTINUE_PLANNING record, current user target intent, source usage/capability graph, approved behavior candidates.

**Outputs:** Exact resolved PHP/Laravel/service/build artifacts, immutable decisions/model/assessment planning basis/plan envelope 2.1, Gate 2 receipt.

**Owned specs:** TARGET-SPECIFICATION, DEPENDENCY-RESOLUTION, TARGET-ARCHITECTURE, CONFIRMATION-PROTOCOL.

**Implementation scope:** Recursive used-capability feasibility, action-enum decision CLI/MCP-neutral service, target model ownership/lowering, typed invalidation/reuse; assessed-state delta and hashed selection/basis closure.

**Tests:** X15 wildcards; X16 reject/alternative; X17 progress hash stability; X18 dependency graph; X29 selective target drift; X49/X50/X55 target/dependency assessment invalidation and assessment-token rejection.

**Acceptance criteria:** User version retained or blocked; no unresolved in-scope target gaps; exact pinned adapter/model/decision refs reconstruct; changing progress leaves hash unchanged; assessment/continuation alone cannot generate implementation work; plan explains changed/reused findings and matches exact selected scope.

**Evidence required:** Feasibility/API/behavior distinctions, license/acquisition manifests, vectors and human reviewed exact plan.

**Blockers:** Unresolvable runtime/dependency/API/license/capability.

**Explicit non-goals:** No silently choosing older runtime or wholesale composer update; no old-result PASS reuse.

## M4 — Generate and accept reference vertical slices

**Objective:** Generate and accept reference vertical slices.

**Prerequisites:** M3.

**Inputs:** Approved target model, bounded contracts/cases, controlled patches.

**Outputs:** Built reference target slices, ownership mappings and scoped reports.

**Owned specs:** MIGRATION-ENGINE, MIGRATION-UNITS, AUTH-AND-SESSION, COMPATIBILITY-SPEC, DATABASE-COMPARISON.

**Implementation scope:** Narrow Laravel adapter generation, auth/session bridge or explicit blocked continuity, views/files/effects, serialized journal publication.

**Tests:** X19 CSRF/session; X20 ID mapping; X21 empty query/exact numbers; X22 duplicate mission/shared file; X23 stale fence.

**Acceptance criteria:** Prerequisites complete without dependent corpus; shared consumer regressions pass; every wrong mutation rejected; unchanged comparator cannot hide repair.

**Evidence required:** Clean independent build, UNIT/regression reports, target/patch/journal/reset/causality artifacts.

**Blockers:** Unsupported safe session codec; missing source behavior; wrong mutant accepted.

**Explicit non-goals:** No broad autonomous business-rule inference; no live production data/effects.

## M5 — Bounded repair and vendor-neutral client interfaces

**Objective:** Bounded repair and vendor-neutral client interfaces.

**Prerequisites:** M4.

**Inputs:** Persisted missions/incidents and guarded engine services.

**Outputs:** Bounded repair evidence, full CLI lifecycle and pinned stdio MCP integration results.

**Owned specs:** MISSION-PROTOCOL, AGENT-MODEL, CLI-SPEC, MCP-SPEC, SKILL-SPEC.

**Implementation scope:** Incident budgets, reviewed patches, CLI/MCP adapters; optional orchestration using fixed capabilities, no provider core coupling.

**Tests:** X24 stale/idempotent work; X26 disconnect; X31 numeric transport/client parity; X32 same-account trust; X46/X51/X52/X55 checkpoint/interface parity and no implicit implementation.

**Acceptance criteria:** Client switch resumes same project without conversation; human decisions available via CLI; no approval bypass; supervisor survives client disconnect; at least two actual clients before integration claim; CLI-only assessment/selection/plan approval works, and MCP cannot turn a report or agent proposal into continuation/approval.

**Evidence required:** Machine protocol/recovery/permission logs, named tested client versions, isolated mutation evidence.

**Blockers:** Unowned supervisor, missing client handshake/framing/permission support.

**Explicit non-goals:** No advertised Claude/Cursor/Codex/Aider/Hermes support without test; no raw SQL/shell.

## M6 — Whole-project correctness and performance/release package

**Objective:** Whole-project correctness and performance/release package.

**Prerequisites:** M5.

**Inputs:** All approved reference units, PROJECT corpus, explicit performance claim/budget and rollback plan.

**Outputs:** Current project report, benchmark evidence, rehearsed release subject and Gate 3 package.

**Owned specs:** COVERAGE, PERFORMANCE, CUTOVER-RUNBOOK, RELEASE, PROJECT-LIFECYCLE.

**Implementation scope:** Full corpus replay, paired measurement/statistical decision, session/data rollback rehearsal, package/restore/native scope metadata.

**Tests:** X33 noisy performance; X34 rollback after writes; X35 whole-project scope; X28 restore; X36 native distribution.

**Acceptance criteria:** All required current cases/dimensions accepted; required metrics accepted under declared claim; rollback reconciles target writes or release blocks; Gate 3 exact subject.

**Evidence required:** Raw paired trials, confidence/budget artifacts, staging rehearsal and restoration/release closure.

**Blockers:** INCONCLUSIVE required performance, unproved session/data reconciliation, missing license/signing/security owner.

**Explicit non-goals:** No engine production deployment; no rollback guarantee from preserved schema alone.

## M7 — Evidence-quality learning and capability reuse

**Objective:** Evidence-quality learning and capability reuse.

**Prerequisites:** M6.

**Inputs:** Project lessons, independent holdouts, sanitized promotion candidates.

**Outputs:** Versioned project lessons; opt-in global revisions with consumer/retraction graph and benchmarks.

**Owned specs:** LEARNING-SYSTEM, KNOWLEDGE-MODEL, SKILL-EVOLUTION, EXPERIMENT-LEDGER.

**Implementation scope:** Optional global import/outbox, exact/full-text search, tool/skill candidate evaluation and controlled activation.

**Tests:** X37 correlated oracle; X38 degraded tool; X39 retraction/privacy.

**Acceptance criteria:** Wrong correlated pattern rejected; capture degradation cannot improve acceptance; every consumer pin/retraction discoverable; opt-in sanitized export only.

**Evidence required:** Holdout/mutation/quality denominators, correlation groups, promotion receipts and retraction replay.

**Blockers:** Nonindependent evidence, failed mutants, privacy leak or no consumer graph.

**Explicit non-goals:** No embeddings/vector service requirement, broad autonomous synthesis or opaque success scores.

## M8 — Broader adapters and optional scale/UI

**Objective:** Broader adapters and optional scale/UI.

**Prerequisites:** M6; M7 only for learning-dependent capabilities.

**Inputs:** Neutrality fixtures and capability proposals.

**Outputs:** Versioned language facets/protocols, independently validated adapter/client maturity, optional UI/scale profiles.

**Owned specs:** IR-CONTRACTS, adapter manifests, PRODUCT, SYSTEM-DESIGN.

**Implementation scope:** One additional target then source; capability-gated streams/concurrency; measured archive/compaction/native profiles and desktop rendering later.

**Tests:** X40 Django→Go/Laravel→Rust/custom PHP→Python; X36 platform; X28 archival; ownership race extensions.

**Acceptance criteria:** No PHP/Laravel assumption in core; no lost uncertainty; unsupported semantics remain explicit; maturity reflects executed scope only.

**Evidence required:** Adapter capability manifests, independent build/replay controls, scale/native/client test results.

**Blockers:** Unrepresentable semantics, unproved concurrency/reset/comparator/retention.

**Explicit non-goals:** No any-source/any-target promise, hostile-agent isolation or years-unbounded-storage claim.

## M1 premise checks (all mandatory)

Durable source approval; reproducible approval hash; coherent snapshot; multi-step bounded scenario; state transition; authorization failure; small executable vocabulary; independent source/target environments; exact approved target resolution; scoped unit acceptance; missing DB-write detection; wrong-principal detection; duplicate-effect detection; incomplete-observation detection; process interruption recovery; honest live execution; deterministic repeated runs without an LLM. A good read-only route alone validates none of these hard premises.

## Legacy task references

RWT-001/002/003/004/005/006 map to M0. RWT-007/008/009/010/011 map to M2 after the thin M1 experiment. RWT-014 maps to M3. RWT-012/013/015/016/017/018 map to M4. RWT-019/020 map to M5. RWT-021/022/025 map to M6. RWT-023/024 map to optional M7/M5 client orchestration as appropriate. RWT-026/027 map to M8. These are provenance aliases only; M0–M8 owns prerequisites/acceptance. No obsolete task implies implemented support.

## Claim gates and implementation restraint

STRUCTURALLY_VALIDATED requires offline schema/vector/DDL checks. IMPLEMENTED requires actual source/build evidence; UNIT_TESTED, RUNTIME_VALIDATED, INTEGRATION_VALIDATED and RELEASE_VALIDATED each need named tests/profile/results. No schema-only maturity transition. Production execution, broad autonomy, global synthesis, desktop/vector search and adapter portfolios cannot precede M1 falsification. If M1 accepts a wrong target, stop expansion and fix the contract/verifier before continuing.

## Assessment checkpoint coverage across milestones

Assessment has no parallel engine, phase queue or extra lifecycle enum. M0 specifies/persists report and human decision projections over synthetic available records; M1 proves the same stateful shared workflow can stop without a feasible plan and resume after explicit selection/approval while preserving the early wrong-target falsification. M2 obtains real partial/static/runtime assessment evidence. M3 derives exact executable selected-work plans and deltas. M5 exercises interface parity. TESTING X41–X55 is the canonical fifteen-test checklist; these are planned tests, not executed runtime evidence. Neither runtime availability nor plan completion is a prerequisite to publishing an honest scoped assessment. Supporting these checkpoint tests must not defer M1 behind broad adapters or a separate assessment subsystem.
