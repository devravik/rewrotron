# Implementation sequence and milestone contracts

**Maturity: SPECIFIED.** This is the implementation task owner. It supersedes historical phase numbering; no milestone is complete. Roadmap and project plan reference this contract. All acceptance experiments in TESTING are planned unless recorded with actual execution evidence.

Execution order: M0 → M1 → M2 → M3 → M4 → M5 → M6; M7 optional after M6; M8 expansion after M6 (and M7 only where relevant). Do not use historical RWT numbering as a parallel implementation queue.

## M0 — Durable evidence and approval foundation

**Objective:** Persist a coherent source snapshot and exact approval closure, then reconstruct the same hash after restart and registry/schema changes; publish an assessment without implementation authority.

**Prerequisites:** Current canonical schemas and owner decisions; qualified filesystem/build required before runtime claims. Documentation resolves architecture, not actual SQLite/WAL/restore qualification.

**Inputs:** Synthetic immutable source tree, current versioned schemas and historical vectors, pinned local filesystem/build profile.

**Outputs:** Guarded CLI JSON, sealed source artifacts, project DB with migrations 0001/0002, canonical subject service, persisted assessment projection/selection records.

**Owned specs:** CONFIRMATION-PROTOCOL, ANALYSIS-SPEC, SQLITE-SCHEMA, ARTIFACT-STORE, CLI-SPEC.

**Implementation scope:** Go bootstrap/ports; exact driver/toolchain manifest; artifact/privacy/admission; snapshot acquisition; explicit human receipts; current execution/event-window service; assessment report publication over available canonical records and trusted human selection (no second analyzer).

**Tests:** X62 source-only record publication; X68/X69 profile/admission records; X73 coverage/reuse; X74 reserved recovery; X75 opaque release structure; X01 hash/restart/registry; X02 snapshot ABA/symlink; X25 ownership/leases; X28 storage/backup; X30 quota; X57 unavailable attestation; X58 pinned execution closure; X41/X43/X45/X46/X51/X53/X55 assessment stop/non-mutation/history/no authority.

**Acceptance criteria:** Reconstruct identical approved subjects from retained bytes after restart; reject stale/forged/missing approval and unsafe snapshot; preserve failure evidence; report completion is separate from migration feasibility and never creates Gate 2/implementation work.

**Evidence required:** Version/checksum manifest, exact command outputs, SQL/Go race results and subject vectors.

**Blockers:** Unsupported local FS; missing pinned SQLite build/version; unavailable canonical/ownership checks.

**Explicit non-goals:** No source runtime/adapters, generation, background survival, physical power-loss hardware qualification (deferred to M6) or universal discovery.

## M1 — Falsifiable deterministic reconstruction experiment

**Objective:** Execute a two-actor stateful scenario against independent environments and reject missing-write, wrong-principal and duplicate-effect mutants while allowing useful assessment-only completion.

**Prerequisites:** M0; explicit model/checker separation and side-local identity contracts (review-2 P0), plus context/principal/branch/response/witness/coverage/admission/filesystem/effect/emergency contracts and structural fixtures. Runtime controls execute inside M1, not as a claim made by this documentation.

**Inputs:** One synthetic stateful application, independently reviewed two-actor bounded scenario and explicit mutant targets.

**Outputs:** Source/Application/Behavior IR subset, exact resolution, small target model, assessment checkpoint/selected-work planning basis, scoped verifier/recovery evidence.

**Owned specs:** BEHAVIOR-MODEL, RUNTIME-OBSERVATION, TARGET-ARCHITECTURE, MIGRATION-UNITS, DIFFERENTIAL-TESTING, RECOVERY.

**Implementation scope:** Small shadow-model interpreter and read-only assertion evaluator with separate capabilities; minimal fixture source/target adapters; login→create operation with authorization failure; independent DBs and sink; foreground/supervisor ownership; current status; report partial state before planning and require explicit human continuation before fixture target planning.

**Tests:** X61–X67/X72/X77/X80 checker isolation, source identity, contexts, branch/occurrence pairing, stale build, cookies, effects, actual principal and source-vs-source; X74 emergency interruption; X03 unsupported operator; X04 branch/state; X05 missing DB write; X06 wrong principal; X07 duplicate effect; X08 incomplete capture; X26 disconnect; X27 event failure; X56 row/binding/terminal semantics; X57 capture derivation; X58 capability qualification; X42/X44/X47/X52/X54 blocked-but-useful assessment, production safety, shared-record reuse and classification.

**Acceptance criteria:** Good target passes UNIT+implemented regression; each wrong mutant fails for intended reason; login completes before create; incomplete capture yields UNKNOWN; interruption recovers without duplicate effects; repeat without LLM; the same fixture can stop successfully at assessment with blocked planning and later reuse valid shared findings after human continuation/Gate 2.

**Evidence required:** 17 premise checks and four decisive boundary controls listed below, exact pinned corpus/policy/subject/trace/reset/target identities, mutation kill report, repeated deterministic runs.

**Blockers:** Any false acceptance, unclosed required dimension, nonreconstructable hash, deadlocked units or unowned process.

**Explicit non-goals:** No broad CI3/Laravel completeness, autonomous repair, global synthesis, desktop, production cutover; read-only JSON is smoke only.

## M2 — Characterize real reference source safely

**Objective:** Boot the pinned legacy fixture in disposable infrastructure and obtain bounded, mechanism-qualified behavior evidence; report unavailable paths rather than declaring false completeness.

**Prerequisites:** M1.

**Inputs:** Pinned licensed CI3/PHP7.4/MySQL/Redis fixture with custom auth, ≥50 routes, cron/payment sink/uploads/legacy dependencies.

**Outputs:** Modified-vendor-aware Source/Application IR, mechanism attestations, source traces, bounded reviewed corpus and provenance-backed assessment (blocked/unknown feasibility is allowed).

**Owned specs:** ANALYSIS-SPEC, CODE-CLASSIFICATION, RUNTIME-OBSERVATION, WORKFLOW-DISCOVERY, AUTH-AND-SESSION.

**Implementation scope:** CI3/Generic PHP subset parser/dispatch adapter, independent container profile, operational boundary and supported connection/worker probes.

**Tests:** X68 filesystem qualification; X69 credential admission; X70 service effects; X76 classification; X09 dynamic dispatch/cron; X10 vendor/flags; X11 PDO/native; X12 async email; X13 unavailable extension; X14 DB reset; X42/X44/X48/X51/X54 assessment with runtime blockers, selective source drift and retrievable findings.

**Acceptance criteria:** Unresolved roots remain UNKNOWN; no false DEAD; direct PDO control prevents false COMPLETE; asynchronous receipt correlated; safe reset/egress verified; missing runtime/dependency evidence produces a useful classified report rather than forced plan completion.

**Evidence required:** Licensed fixture/source/image/lock digests; instrumented control comparison; reviewed role/state/critical workflow partitions.

**Blockers:** Unavailable runtime/native driver; unsafe sandbox; missing operational roots.

**Explicit non-goals:** No claiming complete CI3 support from one application; unsupported eval/streams explicit.

## M3 — Freeze feasible target decisions and implementation model

**Objective:** Derive an exact selected-work plan from current assessment evidence, immutable dependency/performance decisions and owned lowering model, then persist Gate 2 over its reproducible subject.

**Prerequisites:** M2.

**Inputs:** Persisted assessment, explicit human selected-work/CONTINUE_PLANNING record, current user target intent, source usage/capability graph, approved behavior candidates.

**Outputs:** Exact resolved PHP/Laravel/service/build artifacts, immutable dependency/performance decisions/model/assessment planning basis/plan envelope 2.1, Gate 2 receipt.

**Owned specs:** TARGET-SPECIFICATION, DEPENDENCY-RESOLUTION, TARGET-ARCHITECTURE, CONFIRMATION-PROTOCOL.

**Implementation scope:** Recursive used-capability feasibility, action-enum decision CLI/MCP-neutral service, target model ownership/lowering, typed invalidation/reuse; assessed-state delta and hashed selection/basis closure.

**Tests:** X73 positive consumption coverage; X75 provider release interpretation; X15 wildcards; X16 reject/alternative; X17 progress hash stability; X18 dependency graph; X29 selective target drift; X58 profile qualification; X59 performance policy identity; X49/X50/X55 target/dependency assessment invalidation and assessment-token rejection.

**Acceptance criteria:** User version retained or blocked; no unresolved in-scope target gaps; exact pinned adapter/model/decision refs reconstruct; changing progress leaves hash unchanged; assessment/continuation alone cannot generate implementation work; plan explains changed/reused findings and matches exact selected scope.

**Evidence required:** Feasibility/API/behavior distinctions, license/acquisition manifests, vectors and human reviewed exact plan.

**Blockers:** Unresolvable runtime/dependency/API/license/capability.

**Explicit non-goals:** No silently choosing older runtime or wholesale composer update; no old-result PASS reuse.

## M4 — Generate and accept reference vertical slices

**Objective:** Build the reference target from approved mappings and complete prerequisite slices using UNIT plus implemented-regression acceptance without waiting for future dependent cases.

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

### Authorized reference-runtime prerequisite checkpoint

Before Unit 5, the operator explicitly authorized Checkpoint A — Reference Fixture & Runtime Qualification. [ci3-reference](../../fixtures/ci3-reference/README.md) is the small source-only boundary ci3-reference-b1, with pinned dependencies/runtime, sealed source, two independent offline builds and three fresh observed repetitions. Its development evidence proves the listed bounded fixture obligations; its existing 2.1 attestation remains PARTIAL for DB intermediate visibility and current async causal/completion publication. It does not complete M2's full reference breadth or Unit 5, qualify a target/provider, create customer Gate 1/Gate 2 receipts or widen the M4 graph. Unit 5 must preserve these evidence limits. This prerequisite is validated and committed independently; implementation stops at Checkpoint A before Unit 5.

The operator additionally authorized closing pre-Unit-5 admission prerequisites one work item per checkpoint, without changing the eight-unit graph. The first observation sub-item is [SOURCE capture → current context/response projection](../../tools/fixtures/project/README.md), owned by EVALUATION-CONTEXT/RUNTIME-OBSERVATION (M2). The retained three corpora supply typed known values; missing principal, effective request bytes, controlled clock, typed DB visibility and execution/async provenance stay unavailable. Source observation witness/identity/occurrence/trace publication remains a prerequisite before dependent acceptance. This projection does not complete all observation admission, Behavior IR, the fixture session contract or M3 planning closure, and does not authorize Unit 5.

The separately authorized [SOURCE capture role-isolation qualification](../../tools/fixtures/SOURCE-ROLES.md) executes the `rwt-model-roles-1` handle controls against a pinned disposable reference runtime. It qualifies only the successor mediated SOURCE capture/checker implementation and exact profile; it does not upgrade retained captures or publish witnesses/traces. Controlled observation windows and fresh witnessed capture remain the next prerequisite. DB intermediate visibility and async causal/completion ABI remain PARTIAL.

The separately authorized [controlled SOURCE interval checkpoint](../../tools/fixtures/SOURCE-WINDOWS.md) records live supervisor monotonic-to-UTC dispatch anchors and actual end boundaries, qualifies their raw attribution, and collects two fresh reference executions through the unchanged role-isolated read capability. Completion/horizon/queue evidence stays unavailable; DB intermediate visibility and async causal/completion ABI stay PARTIAL. These raw captures do not publish final witnesses, occurrences or traces; source publication closure remains the next prerequisite.

### Explicit M4 implementation unit graph

**Planning decision:** The operator explicitly authorized this engineering decomposition to close the missing M4 unit-graph contract. The original milestone specified the behavior and evidence, but did not uniquely determine these numbered boundaries. This decision preserves that scope, X19–X23 ownership and the canonical specifications; it introduces no new RWT identifiers. “Migration Reconstruction Engine” names the M4 implementation work below, not a separate engine or workflow.

The execution dependency chain is M0 → M1 → M2 → M3 → M4 Unit 1 → Unit 2 → Unit 3 → Unit 4 → Unit 5 → Unit 6 → Unit 7 → Unit 8 → M5. Unit 1 consumes the M0 ownership foundation and still requires M3 under the milestone admission rule. Unit 2 consumes Unit 1 where the existing contracts require publication/ownership. The sequential chain governs development checkpoints even where a semantic dependency is narrower.

| Unit | Scope and implementation artifacts | Prerequisites | Owning requirements and acceptance | Checkpoint status |
| --- | --- | --- | --- | --- |
| 1 — Resource Ownership, Fencing & Serialized Publication | Active work/resource ownership, patch records, fencing and serialized publication primitives | M3; M0 foundation | X22/X23 in [TESTING](../10-development/TESTING.md); [MIGRATION-ENGINE](../04-migration/MIGRATION-ENGINE.md), [TARGET-ARCHITECTURE](../04-migration/TARGET-ARCHITECTURE.md), [RECOVERY](../08-storage/RECOVERY.md). Reject duplicate admission, shared-file conflicts and expired/superseded publishers; validate before hashes | COMPLETE development checkpoint: `4efa1ca` |
| 2 — Lossless Inputs & Lexical Representation Fidelity | Lossless HTTP/JSON projection, exact lexical number tags and explicit coercion boundaries | Unit 1 | X21; [EVALUATION-CONTEXT](../02-domain/EVALUATION-CONTEXT.md), [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md). Preserve absent/empty/repeated/raw inputs, exact numbers and absence/null; refuse unsupported projection and implicit cross-type equality | COMPLETE development checkpoint: `3677ede` |
| 3 — Relational Identity Normalization & Mutation Protection | Typed bijection validation and normalization mutation guards | Unit 2 | X20; [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md), [DATABASE-COMPARISON](../05-verification/DATABASE-COMPARISON.md). Broken relationships, wrong principal, expired/replayed tokens, business clock boundaries and meaningful ordering survive normalization | COMPLETE development checkpoint: `b8dde0e` |
| 4 — Session & CSRF Lifecycle Bridge & Active Continuity Safety | Authoritative semantic contract and read-only acceptance behavior for session identity, actual principal/authentication continuity, login/logout, rotation/fixation, CSRF/session binding, issuance/regeneration/expiry/replay, invalidation, security attributes and cross-request continuity. Later generation consumes this contract | Unit 3 | X19; [AUTH-AND-SESSION](../04-migration/AUTH-AND-SESSION.md), [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md), [BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md), response/evaluation-context/normalization contracts. Required lifecycle mutants fail; missing evidence is UNKNOWN; unsafe/unqualified continuity blocks readiness. No generic target generation | COMPLETE development checkpoint: X19 semantic contracts and witnessed read-only falsification tests; real codec/bridge runtime qualification remains unproven |
| 5 — Narrow Target Generation & Target Authentication/Session Lowering | Deterministic approved target artifacts, routing/handlers/application skeleton, target auth/session glue, generation manifests and safe staging/publication. Consume Unit 4 semantics without reinterpreting them | Unit 4; current M2 evidence and M3 plan/model | Existing M4 generation scope; [LARAVEL-ADAPTER](../04-migration/LARAVEL-ADAPTER.md), [TARGET-ARCHITECTURE](../04-migration/TARGET-ARCHITECTURE.md), [MIGRATION-ENGINE](../04-migration/MIGRATION-ENGINE.md); applicable X67 lowering/provider controls in TESTING. Independent clean build and UNIT/implemented-regression evidence for the narrowly approved surface; exact artifact/provider identities and publication boundaries | INCOMPLETE |
| 6 — Views & File Reconstruction | Approved rendering/output/escaping/encoding, file/path behavior, templates/static assets/resources and safe publication artifacts within the M3 plan | Unit 5 | Existing M4 views/files scope; LARAVEL-ADAPTER, TARGET-ARCHITECTURE, COMPATIBILITY-SPEC and [MIGRATION-UNITS](../04-migration/MIGRATION-UNITS.md). Qualified source obligations and exact planned UNIT/implemented-regression cases preserve observable view/file behavior; no arbitrary project generation | INCOMPLETE |
| 7 — Business/External Effect Reconstruction & Effect Safety | Approved DB/filesystem/external/business/async effects, ordering/causality, duplicate prevention and safe effect publication; recovery/rollback only where required by this scope | Unit 6 | Existing M4 effects scope; MIGRATION-ENGINE, COMPATIBILITY-SPEC, DATABASE-COMPARISON and [DIFFERENTIAL-TESTING](../05-verification/DIFFERENTIAL-TESTING.md). Exact planned effect cases and implemented regressions reject missing/duplicate/misbound effects; checker never mutates measured state. No production effects or widening of deferred concurrency; M6 retains release rollback ownership | INCOMPLETE |
| 8 — M4 Acceptance Closure | Composition evidence for all M4 units: narrow generated reconstruction artifacts, current scoped reports, clean independent build, patch/journal/reset/causality closure and retained M0–M3 invariants | Units 1–7 complete with current valid evidence | All M4 acceptance criteria and X19–X23 plus applicable existing lowering/regression controls. Every required wrong mutation is rejected under the unchanged approved comparator, and all current prerequisite/shared-consumer cases pass. This is the **only M4 completion gate**; M6 still owns whole-project/performance/release acceptance | INCOMPLETE |

Development checkpoint completion requires implementation aligned with the owning specifications, meaningful focused falsification tests, applicable contract checks, `go test -race -count=1 ./...`, configured lint checks and structural validation, followed by an atomic commit recording actual evidence and limits. The COMPLETE entries above identify existing engine-development checkpoints, not runtime qualification of a legacy session bridge or generated reference application. Customer migration-unit acceptance still follows MIGRATION-UNITS: exact UNIT cases, applicable IMPLEMENTED_REGRESSION cases and current valid prerequisites, with no blocking UNKNOWN/MISMATCH; future dependent cases cannot block their prerequisite.

Implement exactly one authorized unit per execution checkpoint; validate, commit and stop before the next unit. A separately authorized documentation planning repair is committed independently before implementation. Unit 4 defines and validates auth/session/CSRF semantics; Unit 5 consumes and lowers those semantics into target artifacts. M4 closure occurs only after Units 1–7 are complete and Unit 8 passes. **M5 must not begin before Unit 8 completion.**

## M5 — Bounded repair and vendor-neutral client interfaces

**Objective:** Repair identified mutants within incident budgets and resume one durable project through CLI and actually tested stdio clients without conversation history or approval bypass.

**Prerequisites:** M4.

**Inputs:** Persisted missions/incidents and guarded engine services.

**Outputs:** Bounded repair evidence, full CLI lifecycle and pinned stdio MCP integration results; tested CLI-only first-assessment and named-client setup instructions with complete operator Skill resource/revision paths.

**Owned specs:** MISSION-PROTOCOL, AGENT-MODEL, CLI-SPEC, MCP-SPEC, SKILL-SPEC.

**Implementation scope:** Incident budgets, reviewed patches, CLI/MCP adapters; optional orchestration using fixed capabilities, no provider core coupling; JEV/bounded-decision integration is deferred and is not part of this milestone or required routing.

**Tests:** X24 stale/idempotent work; X26 disconnect; X31 numeric transport/client parity; X60 offline schema bundles; X32 same-account trust; X46/X51/X52/X55 checkpoint/interface parity and no implicit implementation.

**Acceptance criteria:** Client switch resumes same project without conversation; human decisions available via CLI; no approval bypass; supervisor survives client disconnect; at least two actual clients before integration claim; CLI-only assessment/selection/plan approval works, and MCP cannot turn a report or agent proposal into continuation/approval. The first-assessment guide must run through supported scope registration and Gate 1 to a retrievable report, and both CLI-only and tested-client instructions must name actual build/resource/Skill paths without inventing catalog operations.

**Evidence required:** Machine protocol/recovery/permission logs, named tested client versions, isolated mutation evidence; executed [installation guide](../10-development/INSTALLATION.md) onboarding steps against M5 build artifacts, exact guarded work-scope registration/ref, source confirmation and persisted report with no Gate 2/target writes. Client instructions record executable/arguments/discovery paths and fallback; M6 validates final distributed artifacts.

**Blockers:** Unowned supervisor, missing client handshake/framing/permission support.

**Explicit non-goals:** No advertised Claude/Cursor/Codex/Aider/Hermes support without test; no raw SQL/shell.

## M6 — Whole-project correctness and performance/release package

**Objective:** Accept the full current corpus, recompute the approved statistical claim and rehearse write-aware rollback before publishing the exact Gate 3 release package.

**Prerequisites:** M5.

**Inputs:** All approved reference units, PROJECT corpus, explicit performance claim/budget and rollback plan.

**Outputs:** Current project report, benchmark evidence, rehearsed release subject and Gate 3 package; separately, authenticated Linux amd64 product archive/installer, complete resource/operator Skill bundle, SBOM/notices, release metadata and installation/upgrade/uninstall instructions under RELEASE.

**Owned specs:** COVERAGE, PERFORMANCE, CUTOVER-RUNBOOK, RELEASE, PROJECT-LIFECYCLE.

**Implementation scope:** Full corpus replay, paired measurement/statistical decision, session/data rollback rehearsal, package/restore/native scope metadata; repeatable product packaging/publication from pinned revision, verified manual/user-local installer routes, safe version activation and preserving uninstall. [RELEASE](../10-development/RELEASE.md) owns installation rules; no new milestone or approval workflow. Product packaging is separate from the customer Gate 3 package.

**Tests:** X78 fixed stopping and operational windows; X33 noisy performance; X59 recomputed approved estimands; X34 rollback after writes; X35 whole-project scope; X28 restore; X36-A–H native installation/authenticity/interruption/PATH/offline/upgrade/uninstall/packaged-onboarding/reproducibility controls, retaining existing native driver/path/process tests.

**Acceptance criteria:** All required current cases/dimensions accepted; required metrics accepted under declared claim; rollback reconciles target writes or release blocks; Gate 3 exact subject. Product publication additionally requires the exact archive/installer to pass X36, independently authenticated release metadata, reproducible unsigned payloads, and the M5 walkthrough to work with installed resources; failed or unsupported platform/client claims cannot be published.

**Evidence required:** Raw paired trials, confidence/budget artifacts, staging rehearsal and restoration/release closure; native fresh-install/removal inventories, upgrade/backup checks, release asset/build digests/signatures/SBOM and the actually tested platform/client matrix.

**Blockers:** INCONCLUSIVE required performance, unproved session/data reconciliation, missing license/signing/security owner; unspecified release hosting/trusted signature-verifier instructions or unavailable guarded scope registration in the published walkthrough; failed X36 installation/retention/reproducibility controls.

**Explicit non-goals:** No engine production deployment; no rollback guarantee from preserved schema alone.

## M7 — Evidence-quality learning and capability reuse

**Objective:** Promote only independently evaluated sanitized revisions and locate/invalidate every consumer on retraction without changing historical approvals.

**Prerequisites:** M6.

**Inputs:** Project lessons, independent holdouts, sanitized promotion candidates.

**Outputs:** Versioned project lessons; opt-in global revisions with consumer/retraction graph and benchmarks.

**Owned specs:** LEARNING-SYSTEM, KNOWLEDGE-MODEL, SKILL-EVOLUTION, EXPERIMENT-LEDGER.

**Implementation scope:** Typed global record/migration prerequisites followed by optional global import/outbox, exact/full-text search, tool/skill candidate evaluation and controlled activation.

**Tests:** X79 detached consumers/reopen; X37 correlated oracle; X38 degraded tool; X39 retraction/privacy.

**Acceptance criteria:** Wrong correlated pattern rejected; capture degradation cannot improve acceptance; every consumer pin/retraction discoverable; opt-in sanitized export only.

**Evidence required:** Holdout/mutation/quality denominators, correlation groups, promotion receipts and retraction replay.

**Blockers:** Nonindependent evidence, failed mutants, privacy leak or no consumer graph.

**Explicit non-goals:** No embeddings/vector service requirement, broad autonomous synthesis or opaque success scores.

## M8 — Qualify pluggable source/target framework skills

**Objective:** Demonstrate a substantially different source/target technology through the defined package, adapter binding, TIM, capability and qualification contracts without modifying unrelated neutral migration behavior; preserve the shared workflow and independent verification. Adding another framework alone is insufficient.

**Prerequisites:** M6; M7 only for learning-dependent capabilities. Specify/qualify any wider neutral facets and bounded executable profile constraints before widening exact-profile admission. No JEV dependency. For external package execution, first implement/qualify the existing port bridge specified by SERVICE-CONTRACTS; built-in composition remains valid. Installation alone is not execution, and new neutral semantics require intentional versioned contract extensions.

**Inputs:** Reference Profile #1 retained controls; selected extension proposal with exact source/target/package/provider identities, named supported subset, semantic gaps and independent fixtures. Django→Go, Laravel→Rust, custom PHP→Python and React Native/Expo→native Android are candidates, not committed implementations or automatic support claims.

**Outputs:** Independently packaged source and target skills under existing categories/index; exact adapter/schema/profile qualification records; TIM obligation/provider mappings; compatible and incompatible composition evidence; retained reference regression results. Any new facet is versioned with loss-preservation controls before use.

**Owned specs:** SKILL-SPEC, SERVICE-CONTRACTS, IR-CONTRACTS, TARGET-ARCHITECTURE, adapter manifest/qualification contracts, PRODUCT, SYSTEM-DESIGN; TESTING X40 owns falsification.

**Implementation scope:** Add at least one substantially different source or target technology via package assets/adapters through existing ports. Demonstrate independently selected source/target composition and reuse of the other side's package, without required pair-specific core branches. If required semantics cannot be represented, add the smallest reviewed neutral facet with qualification first; do not redesign the IR or introduce a parallel workflow. Mobile/native is one optional qualification candidate; any selected lifecycle/navigation/platform effects need explicit facets, facilities, harnesses and independent comparators.

**Tests:** Extend X40 with package/adapter identity changes, compatible neutral lowering, missing target facility, unsupported operator/protocol, misleading matching tags, unqualified exact profile, source-target pair restriction and self-certifying harness mutants. Repeat Reference Profile #1 regressions. X58/X61/X67 cover pinned qualification, checker separation and lowering; X36/X28 apply only to selected platform/storage expansions.

**Acceptance criteria:** A new technology works through existing skill/adapter composition and TIM with no reference-framework assumption added to deterministic core; one compatible composition passes its approved scope, and one incompatible/incomplete composition is rejected before generation or reports the required UNKNOWN capture dimension. Both sides preserve provenance/uncertainty. A target cannot certify its own output. Reference profile scope/controls do not regress. Every support claim names exact qualified bytes/profile/corpus, not all frameworks or all combinations.

**Evidence required:** Package version/content manifests, exact installed adapter/schema/profile/control closure, composition/TIM mapping and gap records, independent build/capture/replay/mutant evidence, core dependency/diff audit and retained reference regression reports. Distribution labels alone are insufficient.

**Blockers:** Unrepresentable selected semantics, missing required facilities/lowerings, unqualified capture/comparator/reset, hidden pair coupling or absent independent negative controls.

**Explicit non-goals:** No any-source/any-target promise; no mandatory React Native/Android/mobile profile; no JEV; no broad distributed concurrency, hosted integrations or production execution. Desktop/UI, vector search, scale/compaction and wider platform releases remain separately evidenced later work, not substitutes for the composition experiment.

## M1 premise checks (all mandatory)

Durable source approval; reproducible approval hash; coherent snapshot; multi-step bounded scenario; state transition; authorization failure; small executable vocabulary; independent source/target environments; exact approved target resolution; scoped unit acceptance; missing DB-write detection; wrong-principal detection; duplicate-effect detection; incomplete-observation detection; process interruption recovery; honest live execution; deterministic repeated runs without an LLM. A good read-only route alone validates none of these hard premises.

## Legacy task references

RWT-001/002/003/004/005/006 map to M0. RWT-007/008/009/010/011 map to M2 after the thin M1 experiment. RWT-014 maps to M3. RWT-012/013/015/016/017/018 map to M4. RWT-019/020 map to M5. RWT-021/022/025 map to M6. RWT-023/024 map to optional M7/M5 client orchestration as appropriate. RWT-026/027 map to M8. These are provenance aliases only; M0–M8 owns prerequisites/acceptance. No obsolete task implies implemented support.

## Claim gates and implementation restraint

STRUCTURALLY_VALIDATED requires offline schema/vector/DDL checks. IMPLEMENTED requires actual source/build evidence; UNIT_TESTED, RUNTIME_VALIDATED, INTEGRATION_VALIDATED and RELEASE_VALIDATED each need named tests/profile/results. No schema-only maturity transition. Production execution, broad autonomy, global synthesis, desktop/vector search and adapter portfolios cannot precede M1 falsification. If M1 accepts a wrong target, stop expansion and fix the contract/verifier before continuing.

## Assessment checkpoint coverage across milestones

Assessment has no parallel engine, phase queue or extra lifecycle enum. M0 specifies/persists report and human decision projections over synthetic available records; M1 proves the same stateful shared workflow can stop without a feasible plan and resume after explicit selection/approval while preserving the early wrong-target falsification. M2 obtains real partial/static/runtime assessment evidence. M3 derives exact executable selected-work plans and deltas. M5 exercises interface parity. TESTING X41–X55 is the canonical fifteen-test checklist; these are planned tests, not executed runtime evidence. Neither runtime availability nor plan completion is a prerequisite to publishing an honest scoped assessment. Supporting these checkpoint tests must not defer M1 behind broad adapters or a separate assessment subsystem.

## Review-2 dependency and admission contract

```text
Canonical contracts + structural negative fixtures
  → M0: persistence/approval/snapshot/assessment/current-state foundation
  → M1: isolated measured execution + read-only verifier falsification
  → M2: real source/profile/service capture qualification
  → M3: exact selected target/decisions/model + coverage-based reuse
  → M4: reference generation and currently implemented regressions
  → M5: CLI and actual client sessions
  → M6: whole-project corpus + fixed-N statistics/SLO + release package
  → optional M7 consumer/retraction learning; M8 pluggable profile qualification
```

Foundation may start with Go bootstrap, pinned SQLite, immutable artifact registration, canonical serialization/hash, coherent snapshot acquisition, assessment persistence, event/current-state projection and structural contracts. Verifier implementation requires the two P0 contracts now owned by BEHAVIOR-MODEL and RUNTIME-OBSERVATION: no model/checker effects in measured state; source-only identity requires no pair/case. Neither is a claim of executed controls.

Before M1/reference generation, consume EVALUATION-CONTEXT, principal fixtures, branch pairing, typed response/session, execution witnesses, filesystem/admission, coverage and typed effects, plus ARTIFACT-STORE/RECOVERY reservations. M0 checks their storage/shape; M1 proves only the synthetic admitted subset; M2 qualifies customer/runtime filesystem and credential behavior. Do not move arbitrary legacy admission support into M1.

Four mandatory M1 controls supplement the existing seventeen: checker cannot mutate measured stores; source evidence publishes before a target exists; stale/wrong executable provenance is rejected; established branch divergence is a meaningful MISMATCH. Good control must PASS first; wrong-target semantic failures require valid startup/capture. No infrastructure failure counts as a semantic kill. M0 proves durable bounded records, not a working verifier. M1 proves the named narrow fixture's falsification behavior, not broad application reconstruction, corpus completeness, arbitrary PHP, browser support, concurrency, live payments or production cutover.

General concurrency/long-delay behavior (including X71 beyond the refusal control), broader adapters, global synthesis, desktop/vector search, hosted integrations and production execution stay deferred. M7 must design consumer/retraction schemas before activation; it is not a foundation prerequisite.

Target generation additionally requires current Target Implementation Model 2.1 provider bindings per TARGET-ARCHITECTURE. Current report/trace/context ref validation follows the machine branch registry; historical variants are not fallback execution contracts.
