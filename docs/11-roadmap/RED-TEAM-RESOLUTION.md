# Post-red-team revision report and disposition matrix

**Latest checkpoint:** See “Review-2 revision deliverables (current checkpoint)” below for the current A–H manifest, disposition matrix and validation limits. Earlier counts, decisions and manifests are historical audit entries.

Revision date: 2026-10-03. Input: unchanged Rewrotron_Pre-Implementation_Red-Team_Review.md. This report is an audit index, not a competing subsystem specification. RESOLVED means the owning specification/machine representation was revised; it never means implementation/runtime experiments passed.

## Authority and impact map

AGENTS → Product → System/ADRs → domain/behavior → machine interfaces/schemas → storage/compatibility → tests → implementation. The original exploratory project plan is replaced with a current overview pointing to DEVELOPMENT-PLAN, the sole milestone contract. Glossary/overview/audit documents delegate detailed semantics to owners. v1 descriptive artifacts remain readable history; current execution uses explicit v2 records and service invariants. No production Go/adapter/runtime/client implementation was introduced.

Critical owner mapping: CR-01 approval subject to CONFIRMATION-PROTOCOL/ADR0010; CR-02 acceptance to MIGRATION-UNITS/ADR0004; CR-03 executable semantics to BEHAVIOR-MODEL/ADR0003; CR-04 attestation to RUNTIME-OBSERVATION/ADR0005; CR-05 process ownership to RECOVERY/ADR0013; CR-06 validity/reuse to DOMAIN-MODEL/ADR0012; CR-07 independent oracle to TESTING/ADR0005. All are RESOLVED as specifications, pending assigned experiments.

## Red-team resolution matrix

Each original failure scenario has an explicit owning document, companion schema, ADR, priority and disposition. The original review contains its concrete conflict/evidence; the change below is the selected canonical behavior. Engine code for every row remains pending.

| Finding | Decision | Owning document | Resolution | Status |
| --- | --- | --- | --- | --- |
| F01: Database rows select controller classes | Accept defect; revise canonical contract | [ANALYSIS-SPEC](../03-analysis/ANALYSIS-SPEC.md) | Version configuration/data roots and preserve uncertainty Priority: P1. Machine owner: specs/project/source-snapshot.schema.json; ADR 0010. Runtime evidence is still required. | RESOLVED |
| F02: Quarterly cron uses an apparently unused helper | Accept defect; revise canonical contract | [CODE-CLASSIFICATION](../03-analysis/CODE-CLASSIFICATION.md) | Require operational-root closure for DEAD proofs Priority: P0 before omission. Machine owner: specs/analysis/omission-proof.schema.json; ADR 0010. Runtime evidence is still required. | RESOLVED |
| F03: Production generates PHP with `eval` | Accept defect; revise canonical contract | [ANALYSIS-SPEC](../03-analysis/ANALYSIS-SPEC.md) | Capture safe generated artifacts or block scope Priority: P1. Machine owner: specs/observation/attestation.schema.json; ADR 0003. Runtime evidence is still required. | RESOLVED |
| F04: Customer patched the framework/vendor tree | Accept defect; revise canonical contract | [CI3-ADAPTER](../04-migration/CI3-ADAPTER.md) | Bind capabilities to content and modification findings Priority: P1. Machine owner: specs/project/source-snapshot.schema.json; ADR 0009. Runtime evidence is still required. | RESOLVED |
| F05: Production feature flag enables a dormant path | Accept defect; revise canonical contract | [ANALYSIS-SPEC](../03-analysis/ANALYSIS-SPEC.md) | Version approved environment profiles Priority: P1. Machine owner: specs/observation/profile.schema.json; ADR 0010. Runtime evidence is still required. | RESOLVED |
| F06: PHP 7.4 application requires an unavailable extension | Accept defect; revise canonical contract | [RUNTIME-OBSERVATION](../03-analysis/RUNTIME-OBSERVATION.md) | Runtime acquisition/retention decision; no host fallback Priority: P1. Machine owner: specs/project/resolved-environment.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F07: Source changes and changes back during scanning | Accept defect; revise canonical contract | [ANALYSIS-SPEC](../03-analysis/ANALYSIS-SPEC.md) | Execute from an immutable coherent source snapshot Priority: P0 before baseline execution. Machine owner: specs/project/source-snapshot.schema.json; ADR 0010. Runtime evidence is still required. | RESOLVED |
| F08: User override says PHP 8.2, actual runtime is 7.4 | Accept defect; revise canonical contract | [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) | Preserve assertions separately from runtime attestations Priority: P1. Machine owner: specs/project/prepared-confirmation.schema.json; ADR 0012. Runtime evidence is still required. | RESOLVED |
| F09: Checkout branches on inventory and account status | Accept defect; revise canonical contract | [BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md) | Typed branches or explicit separate bounded scenarios Priority: P1. Machine owner: specs/behavior/workflow.schema.json; ADR 0003. Runtime evidence is still required. | RESOLVED |
| F10: HTTP response precedes a queued email | Accept defect; revise canonical contract | [BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md) | Define asynchronous completion boundaries Priority: P0 before effect acceptance. Machine owner: specs/observation/causality.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F11: Application exposes WebSockets | Capability boundary / later validation | [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md) | General stream/WebSocket semantics deferred M8; inventory reports UNSUPPORTED, never HTTP flattening. Priority: P2. Machine owner: specs/ir/application-ir.schema.json; ADR 0009. Runtime evidence is still required. | DEFERRED |
| F12: Helper opens PDO outside CI3’s DB driver | Accept defect; revise canonical contract | [RUNTIME-OBSERVATION](../03-analysis/RUNTIME-OBSERVATION.md) | Completeness attestation covering all connections Priority: P0 before PASS. Machine owner: specs/observation/attestation.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F13: Native extension writes files or contacts a service | Accept defect; revise canonical contract | [RUNTIME-OBSERVATION](../03-analysis/RUNTIME-OBSERVATION.md) | Capability-specific observation or UNKNOWN Priority: P1. Machine owner: specs/observation/attestation.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F14: Failed transaction consumes auto-increment values | Accept defect; revise canonical contract | [DATABASE-COMPARISON](../05-verification/DATABASE-COMPARISON.md) | Restore/check sequence behavior explicitly Priority: P1. Machine owner: specs/observation/reset-report.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F15: Request creates temporary tables or performs DDL | Accept defect; revise canonical contract | [DATABASE-COMPARISON](../05-verification/DATABASE-COMPARISON.md) | Fresh DB instance or supported full reset Priority: P0 before replay. Machine owner: specs/observation/reset-report.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F16: Two connections both contain an `orders` table | Accept defect; revise canonical contract | [DATABASE-COMPARISON](../05-verification/DATABASE-COMPARISON.md) | Add connection/catalog/schema identity and typed value-mapping evidence Priority: P1. Machine owners: specs/observation/effects.schema.json and specs/ir/data-model.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F17: Payment provider offers no safe sandbox | Accept defect; revise canonical contract | [SANDBOX](../09-security/SANDBOX.md) | Verified mock/retained boundary or explicit unsupported scope Priority: P0. Machine owner: specs/observation/effects.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F18: Worker crashes after sink acceptance but before acknowledgement | Accept defect; revise canonical contract | [BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md) | Model acknowledgements, idempotency and delivery semantics Priority: P1. Machine owner: specs/observation/causality.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F19: Redis holds sessions, locks, queues and feature flags | Accept defect; revise canonical contract | [DATABASE-COMPARISON](../05-verification/DATABASE-COMPARISON.md) | Separate stores/namespaces and reset contracts Priority: P1. Machine owner: specs/observation/profile.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F20: CI3 session contains unsafe serialized objects | Accept defect; revise canonical contract | [AUTH-AND-SESSION](../04-migration/AUTH-AND-SESSION.md) | Safe bridge or approved re-login exception Priority: P1. Machine owner: specs/verification/normalization.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F21: Logout on target leaves source session valid | Accept defect; revise canonical contract | [AUTH-AND-SESSION](../04-migration/AUTH-AND-SESSION.md) | Explicit bridge and rollback state ownership Priority: P1. Machine owner: specs/verification/normalization.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F22: Normalizer masks CSRF tokens but misses replay acceptance | Accept defect; revise canonical contract | [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md) | Relation plus lifecycle assertions Priority: P0 before auth acceptance. Machine owner: specs/verification/normalization.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F23: Timestamp determines refund eligibility | Accept defect; revise canonical contract | [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md) | Controlled clock and business-time predicates Priority: P1. Machine owner: specs/behavior/semantics.schema.json; ADR 0003. Runtime evidence is still required. | RESOLVED |
| F24: SQL without ORDER BY returns varying results | Accept defect; revise canonical contract | [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md) | Preserve order unless an approved policy establishes otherwise Priority: P1. Machine owner: specs/verification/normalization.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F25: New PDF library emits different metadata/layout | Capability boundary / later validation | [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md) | General PDF/image semantic comparators deferred M8; approved format comparator required or MISMATCH/UNKNOWN. Priority: P2. Machine owner: specs/verification/normalization.schema.json; ADR 0005. Runtime evidence is still required. | DEFERRED |
| F26: Request uses `?filter=` | Accept defect; revise canonical contract | [DIFFERENTIAL-TESTING](../05-verification/DIFFERENTIAL-TESTING.md) | Lossless request representation allowing empty values Priority: P1. Machine owner: specs/behavior/request.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
| F27: MCP client rounds a large JSON numeric ID | Accept defect; revise canonical contract | [DIFFERENTIAL-TESTING](../05-verification/DIFFERENTIAL-TESTING.md) | Exact typed values/raw payload refs Priority: P1. Machine owner: specs/behavior/request.schema.json; ADR 0007. Runtime evidence is still required. | RESOLVED |
| F28: Target adapter upgrades after plan approval | Accept defect; revise canonical contract | [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) | Immutable adapter identity in hash preimage Priority: P0 before Gate 2. Machine owner: specs/project/approval-subject.schema.json; ADR 0010. Runtime evidence is still required. | RESOLVED |
| F29: “Resolved” target still contains `13.*` and no runtime | Accept defect; revise canonical contract | [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) | Dedicated resolved-environment schema Priority: P1. Machine owner: specs/project/resolved-environment.schema.json; ADR 0012. Runtime evidence is still required. | RESOLVED |
| F30: User rejects one dependency replacement | Accept defect; revise canonical contract | [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) | Typed human decision/import operation Priority: P1. Machine owner: specs/project/decision.schema.json; ADR 0012. Runtime evidence is still required. | RESOLVED |
| F31: Dependency status advances PLANNED → VERIFIED | Accept defect; revise canonical contract | [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) | Separate immutable decision from progress Priority: P1. Machine owner: specs/migration/progress.schema.json; ADR 0012. Runtime evidence is still required. | RESOLVED |
| F32: Login must complete before checkout generation | Accept defect; revise canonical contract | [MIGRATION-UNITS](../04-migration/MIGRATION-UNITS.md) | Scoped unit acceptance Priority: P0. Machine owner: specs/verification/acceptance-scope.schema.json; ADR 0004. Runtime evidence is still required. | RESOLVED |
| F33: Duplicate mission rows represent identical work | Accept defect; revise canonical contract | [MISSION-PROTOCOL](../06-agent/MISSION-PROTOCOL.md) | Deduplication invariant and atomic resource claims Priority: P1. Machine owner: specs/missions/mission.schema.json; ADR 0004. Runtime evidence is still required. | RESOLVED |
| F34: Distinct missions edit the same shared auth file | Accept defect; revise canonical contract | [RECOVERY](../08-storage/RECOVERY.md) | Serialized publication/resource ownership Priority: P1. Machine owner: specs/migration/target-model.schema.json; ADR 0013. Runtime evidence is still required. | RESOLVED |
| F35: Independent unit A increments project revision | Capability boundary / later validation | [SYSTEM-DESIGN](../01-architecture/SYSTEM-DESIGN.md) | Parallel disjoint mutations deferred; serial per-project baseline and explicit claims are canonical, no concurrency guarantee. Priority: P2. Machine owner: specs/missions/mission.schema.json; ADR 0013. Runtime evidence is still required. | DEFERRED |
| F36: PHP target changes 8.3 → 8.4 at 70% completion | Accept defect; revise canonical contract | [DOMAIN-MODEL](../02-domain/DOMAIN-MODEL.md) | Versioned dependency graph and reuse certificates Priority: P1. Machine owner: specs/project/validity-graph.schema.json; ADR 0012. Runtime evidence is still required. | RESOLVED |
| F37: MCP process exits during observation | Accept defect; revise canonical contract | [RECOVERY](../08-storage/RECOVERY.md) | Explicit supervisor/foreground lifecycle Priority: P1. Machine owner: specs/operations/result.schema.json; ADR 0013. Runtime evidence is still required. | RESOLVED |
| F38: Client requests historical state at sequence N | Accept defect; revise canonical contract | [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) | Historical projections or current-state distinction Priority: P1. Machine owner: specs/execution/page.schema.json; ADR 0013. Runtime evidence is still required. | RESOLVED |
| F39: Queue event must be linked to originating workflow | Accept defect; revise canonical contract | [RUNTIME-OBSERVATION](../03-analysis/RUNTIME-OBSERVATION.md) | Canonical correlation envelope Priority: P1. Machine owner: specs/observation/causality.schema.json; ADR 0013. Runtime evidence is still required. | RESOLVED |
| F40: Status client polls continuously | Accept defect; revise canonical contract | [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) | Bounded/coalesced inspection auditing Priority: P2. Machine owner: specs/execution/page.schema.json; ADR 0013. Runtime evidence is still required. | RESOLVED |
| F41: Thousands of cases retain repeated full captures | Accept defect; revise canonical contract | [ARTIFACT-STORE](../08-storage/ARTIFACT-STORE.md) | Retention classes, archives and admission limits Priority: P1. Machine owner: specs/operations/archive.schema.json; ADR 0002. Runtime evidence is still required. | RESOLVED |
| F42: Backup copies DB and artifacts at different times | Accept defect; revise canonical contract | [RECOVERY](../08-storage/RECOVERY.md) | Coordinated backup manifest and retention barrier Priority: P1. Machine owner: specs/operations/archive.schema.json; ADR 0002. Runtime evidence is still required. | RESOLVED |
| F43: Workspace is placed on a network filesystem | Accept defect; revise canonical contract | [SQLITE-SCHEMA](../08-storage/SQLITE-SCHEMA.md) | Supported filesystem policy Priority: P1. Machine owner: specs/project/project.schema.json; ADR 0002. Runtime evidence is still required. | RESOLVED |
| F44: Replacement SDK is API-compatible but changes retry behavior | Accept defect; revise canonical contract | [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) | Effect/error/retry contracts Priority: P1. Machine owner: specs/migration/dependency-resolution.schema.json; ADR 0012. Runtime evidence is still required. | RESOLVED |
| F45: Three generated fixtures share the same incorrect oracle | Accept defect; revise canonical contract | [LEARNING-SYSTEM](../07-learning/LEARNING-SYSTEM.md) | Independence criteria and retraction graph Priority: P1 before promotion. Machine owner: specs/project/validity-graph.schema.json; ADR 0014. Runtime evidence is still required. | RESOLVED |
| F46: Experimental tool omits difficult traces | Accept defect; revise canonical contract | [LEARNING-SYSTEM](../07-learning/LEARNING-SYSTEM.md) | Benchmark evidence quality, not just task success Priority: P1 before activation. Machine owner: specs/project/validity-graph.schema.json; ADR 0014. Runtime evidence is still required. | RESOLVED |
| F47: Rust/Python source needs declarations outside the current symbol vocabulary | Capability boundary / later validation | [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md) | Broader neutral language facets deferred M8; unsupported declarations retain loss/uncertainty and block closure. Priority: P2. Machine owner: specs/ir/source-ir.schema.json; ADR 0009. Runtime evidence is still required. | DEFERRED |
| F48: AST or compiler diagnostic contains a source secret | Accept defect; revise canonical contract | [SECURITY](../09-security/SECURITY.md) | Private-source policy and derived-data redaction Priority: P0 before export. Machine owner: specs/operations/archive.schema.json; ADR 0006. Runtime evidence is still required. | RESOLVED |
| F49: Target performance equals source within noise | Accept defect; revise canonical contract | [PERFORMANCE](../05-verification/PERFORMANCE.md) | Define statistical claim and explicit acceptance budget Priority: P1. Machine owner: specs/verification/benchmark.schema.json; ADR 0015. Runtime evidence is still required. | RESOLVED |
| F50: Target writes new session/payment state during cutover | Accept defect; revise canonical contract | [CUTOVER-RUNBOOK](../10-development/CUTOVER-RUNBOOK.md) | Data/session reconciliation contract Priority: P1 before release. Machine owner: specs/project/approval-subject.schema.json; ADR 0010. Runtime evidence is still required. | RESOLVED |
| F51: Multi-process WAL uses an affected SQLite version | Accept defect; revise canonical contract | [SQLITE-SCHEMA](../08-storage/SQLITE-SCHEMA.md) | Pin a release containing the WAL-reset fix Priority: P0 before multi-writer release. Machine owner: specs/operations/archive.schema.json; ADR 0002. Runtime evidence is still required. | RESOLVED |

## Missing decision dispositions

| Finding | Decision | Owning document | Resolution | Status |
| --- | --- | --- | --- | --- |
| D01 | Explicit versioned preimage | CONFIRMATION-PROTOCOL | rwt-cjson-1 SOURCE/PLAN/RELEASE and original immutable reference closure | RESOLVED |
| D02 | Three acceptance scopes | MIGRATION-UNITS | Unit + implemented regressions; whole project only at project acceptance | RESOLVED |
| D03 | Bounded typed operators | BEHAVIOR-MODEL | Closed vocabulary, limits, unsupported errors; v1 prose not executable | RESOLVED |
| D04 | Explicit bounded branching | BEHAVIOR-MODEL | Actor/state/guard/binding/terminal/async contracts; no unbounded streams/concurrency | RESOLVED |
| D05 | Mechanism-derived completeness | RUNTIME-OBSERVATION | Required capture/escape/control/closure evidence, not observer assertion | RESOLVED |
| D06 | Foreground and supervised modes | RECOVERY | Explicit owner/start/heartbeat/lease/orphan/reconnect semantics | RESOLVED |
| D07 | Typed proposals plus Gate 2 | CLI-SPEC/MCP-SPEC | ACCEPT/REJECT/strategy/alternative/evidence/exception actions; proposals never self-approve | RESOLVED |
| D08 | Freeze choice, separate progress | DEPENDENCY-RESOLUTION | v2 immutable decision/model with separate progress schema | RESOLVED |
| D09 | Typed graph and certificates | DOMAIN-MODEL | Explicit predicates/reverse invalidation; unknown conservative; verification reruns | RESOLVED |
| D10 | Current state plus event window | EXECUTION-OBSERVABILITY | Historical-state reconstruction recommendation rejected for baseline; N current versus M event cursor explicit | RESOLVED |
| D11 | Coherent sealed acquisition | ANALYSIS-SPEC | Immutable export/snapshot/quiesced copy; live scan/ABA insufficient | RESOLVED |
| D12 | Explicit operational profile | RUNTIME-OBSERVATION | Supported mechanisms/protocols and unclosed roots remain visible | RESOLVED |
| D13 | Private vault/sanitized evidence | SECURITY | Byte-exact private source; mediated export/privacy controls; same-account limit explicit | RESOLVED |
| D14 | Quotas and verified archives | ARTIFACT-STORE/RECOVERY | Protected closure never silently GC; active-history compaction deferred with admission cap | RESOLVED |
| D15 | Explicit statistical claim | PERFORMANCE | Improvement/non-inferiority/budget, bounded paired methodology | RESOLVED |
| D16 | Holdout/quality/retraction; global machine contract deferred | LEARNING-SYSTEM / KNOWLEDGE-MODEL | Baseline project guards resolved; typed global revision/evaluation/consumer/retraction/outbox contracts are explicit prerequisites before M7 activation | DEFERRED |
| D17 | Decision ref before receipt | COMPATIBILITY-SPEC | Normalization proposal ref approved transitively by plan, no circular approval hash | RESOLVED |
| D18 | Serial baseline + resource ownership | MISSION-PROTOCOL/RECOVERY | Atomic active-work uniqueness/claims plus process ownership through writes | RESOLVED |

## Rejected or already-covered recommendations

| Finding | Decision | Owning document | Resolution | Status |
| --- | --- | --- | --- | --- |
| Full historical as-of state as baseline | Reject additional event-sourcing requirement | EXECUTION-OBSERVABILITY / ADR0013 | Current rows at N and event window M are sufficient; historical projection explicitly deferred | REJECTED |
| Same-account agent cannot access SQLite/source | Reject absolute enforcement claim | SECURITY / ADR0006 | Mediated services enforce permissions; hostile-agent OS separation is a different deferred profile | REJECTED |
| New ADR for every finding | Reuse existing owners | ADR0001–0013 | Update existing decisions; only learning independence and statistical performance require new ADR0014/0015 | REJECTED |
| Additional successful verification states | Retain fundamental three outcomes | COMPATIBILITY-SPEC | Intentional change is policy; unsupported/blocking is reason. No UNKNOWN waiver | ALREADY COVERED |
| Core neutrality/local artifacts/independent DB/no blind real-effect replay | Preserve sound invariants | PRODUCT/SYSTEM-DESIGN/SANDBOX | Tightened identities/mechanisms but no replacement with vendor-specific platform | ALREADY COVERED |

## Assumption Elimination Pass

Every unexecuted runtime capability remains ASSUMPTION, not discharged by accepting prose. The following conversion traces review claims to falsifiable prerequisites. UNKNOWN requires these exact evidence classes; obtaining them is assigned to milestones, never guessed.

| ASSUMPTION group | Measurable/configurable prerequisite | Conversion / experiment |
| --- | --- | --- |
| Detection/runtime | Modified vendor content, sandbox measured runtime versus assertions | Detection/doctor controls X10/X13 |
| Immutable baseline/dead code | Coherent sealed bytes and closed operational roots/parsing/graph/corpus | Acquisition/omission bundle X02/X09 |
| Same/equivalent behavior/rules | Typed operators, scope/profile/policy and supported semantics | Executable limit/type tests X03 plus mutants X05–X08 |
| Workflow covered | Actor/branch/state partitions and terminal boundaries | Independent critical scenario inventory X04/X35 |
| No effects/safe replay | Every required mechanism contained/captured and resettable/simulated | Attestation/egress/reset controls X08/X11/X12/X14 |
| Restored fixture/control-neutral instrumentation | Every connection/store/sequence/process state restored; controls match | Reset/control evidence X11/X14 |
| Harmless dynamic values/corresponding IDs | Bijective relationships and token/time lifecycle significance | Normalization decision plus adversarial mutants X19/X20/X21 |
| Authentication preserved | Safe codec/hash/principal/revocation and rollback | X19/X34 or explicit blocker/approved intentional scope |
| Dependency compatible/target pinned | Platform, used API, behavior, license and exact available bytes | Resolution/characterization X13/X15/X18 |
| User approved/unit complete | Current exact receipt and guarded acceptance scope | X01/X16/X17/X35 |
| Unrelated work valid | Typed unchanged-parent witnesses, conservative unknown fallback | Reuse certificate X29; no old PASS relabel |
| Agent cannot bypass/vendor-neutral | Mediated calls and explicit same-account trust; complete CLI/client parity | X25/X31/X32; hostile-agent profile deferred |
| Resume/one owner/background survival | Durable refs/process ownership/resource fence/ack/journal | X22–X26/X28 |
| Current status/meaningful progress | Current N distinct from event M; persisted basis; heartbeats not progress | X27 |
| Privacy/backup/SQLite sufficient | Secret canaries, local filesystem/pinned build, closure and quota | X25/X28/X30/X36; scale UNKNOWN until soak |
| Performance improved | Explicit workload/claim/method and adequate paired samples | X33; equality may inconclusive under strict claim |
| Reusable knowledge/stable tool | Independent holdouts/mutants, evidence quality, pinned consumers | X37–X39 before M7 activation |
| Offline reproducible/open-source distributable | Retained executable/package/image bytes plus SBOM/notices/license | Disconnected replay and native/license audit X13/X36 |
| Safe rollback | Target writes/session/effect state actually reconciled in rehearsal | X34; otherwise block release |

## Remaining blockers and revised order

P0: no unresolved canonical design decision identified after revision; implementation cannot claim validation yet. M0 must verify pinned driver/engine/filesystem and hash/snapshot controls before durable/runtime use.

P1 evidence blockers: M1 wrong-target rejection/capture/isolation/recovery premise remains unexecuted; actual legacy runtime/probe/parser viability, exact target lowering/dependencies and safe session/rollback behavior are UNKNOWN until M2–M6 evidence. Actual client portability, performance, native release and operational security ownership also require named evidence before their claims.

P2 deferred capabilities: streams/general async/distributed concurrency; broader neutral declarations/adapters; hostile-agent OS isolation; online active-history compaction; global synthesis/promotion pending M7 quality/privacy gates; disjoint concurrent mutation pending resource revision experiments. The baseline has explicit refusal/admission bounds.

P3: desktop rendering, vector search and convenience configure UI. None blocks the bounded reference experiment.

Order: M0 durable foundation → M1 falsifiable stateful experiment → M2 real reference characterization → M3 exact resolution/decisions/model → M4 accepted slices → M5 bounded repair/interfaces → M6 project/performance/release; optional M7 learning; M8 capability expansion. TESTING owns the current experiment registry; all are planned, not executed here.

## Contradiction scan

Canonical choices now align: bounded equivalence; SPECIFIED adapters; MCP stdio; BENCHMARKED and whole-project readiness before Gate 3; scoped unit completion; intent separate from exact resolution; mediated same-account trust; current N/event-window M; protected retention plus quotas; pinned revisions; frozen decisions/progress. Known stale-phrase scans, local links, schema references/catalogs and incremental DDL are checked by the validator. The attached historical review is deliberately excluded from normative stale-text scanning.

No unresolved contradiction is intentionally retained among authoritative owners. Explicit deferred capabilities above are limitations, not competing answers. This is a documentation consistency assessment, not a proof of implementation/runtime correctness. Validation results and changed/new file manifest follow below.

## A. Changed files

Every changed pre-existing file is listed; the attached review and migration 0001 are unchanged.

| File | Reason | Major changes |
| --- | --- | --- |
| [AGENTS.md](../../AGENTS.md) | Remove unsupported constitutional claims; delegate detailed semantics to owners | Bounded claims/gates/scopes/trust/maturity and pinned dependency floor |
| [README.md](../../README.md) | Reflect revised contract/status | Exact-resolution/model pipeline and current revision report links |
| [REWROTRON_PROJECT_PLAN.md](../../REWROTRON_PROJECT_PLAN.md) | Replace exploratory phase queue | Current evidence-driven M0–M8 sequence and explicit deferred boundaries |
| [docs/00-vision/GLOSSARY.md](../../docs/00-vision/GLOSSARY.md) | Replace broad/competing product vocabulary | Evidence-bounded goals, claim/maturity semantics and owner references |
| [docs/00-vision/GOALS.md](../../docs/00-vision/GOALS.md) | Replace broad/competing product vocabulary | Evidence-bounded goals, claim/maturity semantics and owner references |
| [docs/00-vision/NON-GOALS.md](../../docs/00-vision/NON-GOALS.md) | Replace broad/competing product vocabulary | Evidence-bounded goals, claim/maturity semantics and owner references |
| [docs/00-vision/PRODUCT.md](../../docs/00-vision/PRODUCT.md) | Replace broad/competing product vocabulary | Evidence-bounded goals, claim/maturity semantics and owner references |
| [docs/01-architecture/ADR/0001-go-core.md](../../docs/01-architecture/ADR/0001-go-core.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0002-sqlite-storage.md](../../docs/01-architecture/ADR/0002-sqlite-storage.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0003-behavior-as-canonical-model.md](../../docs/01-architecture/ADR/0003-behavior-as-canonical-model.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0004-vertical-slice-migration.md](../../docs/01-architecture/ADR/0004-vertical-slice-migration.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0005-differential-verification.md](../../docs/01-architecture/ADR/0005-differential-verification.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0006-vendor-neutral-agent-interface.md](../../docs/01-architecture/ADR/0006-vendor-neutral-agent-interface.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0007-cli-as-baseline-interface.md](../../docs/01-architecture/ADR/0007-cli-as-baseline-interface.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0008-mcp-as-agent-interface.md](../../docs/01-architecture/ADR/0008-mcp-as-agent-interface.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0009-three-tier-ir-architecture.md](../../docs/01-architecture/ADR/0009-three-tier-ir-architecture.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0010-mandatory-user-confirmation-gates.md](../../docs/01-architecture/ADR/0010-mandatory-user-confirmation-gates.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0011-development-baseline-and-contract-completion.md](../../docs/01-architecture/ADR/0011-development-baseline-and-contract-completion.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0012-target-specification-and-dependency-resolution.md](../../docs/01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0013-transparent-execution.md](../../docs/01-architecture/ADR/0013-transparent-execution.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ARCHITECTURE.md](../../docs/01-architecture/ARCHITECTURE.md) | Align architecture/ports with approved records | Three IRs, target model, guarded services, process ownership and scoped pipeline |
| [docs/01-architecture/DATA-FLOW.md](../../docs/01-architecture/DATA-FLOW.md) | Align architecture/ports with approved records | Three IRs, target model, guarded services, process ownership and scoped pipeline |
| [docs/01-architecture/SERVICE-CONTRACTS.md](../../docs/01-architecture/SERVICE-CONTRACTS.md) | Align architecture/ports with approved records | Three IRs, target model, guarded services, process ownership and scoped pipeline |
| [docs/01-architecture/SYSTEM-DESIGN.md](../../docs/01-architecture/SYSTEM-DESIGN.md) | Align architecture/ports with approved records | Three IRs, target model, guarded services, process ownership and scoped pipeline |
| [docs/02-domain/BEHAVIOR-MODEL.md](../../docs/02-domain/BEHAVIOR-MODEL.md) | Define executable/reproducible domain contract | Canonical identity, typed semantics/scenarios, separate records and validity/lifecycle rules |
| [docs/02-domain/CONFIRMATION-PROTOCOL.md](../../docs/02-domain/CONFIRMATION-PROTOCOL.md) | Define executable/reproducible domain contract | Canonical identity, typed semantics/scenarios, separate records and validity/lifecycle rules |
| [docs/02-domain/DOMAIN-MODEL.md](../../docs/02-domain/DOMAIN-MODEL.md) | Define executable/reproducible domain contract | Canonical identity, typed semantics/scenarios, separate records and validity/lifecycle rules |
| [docs/02-domain/IR-CONTRACTS.md](../../docs/02-domain/IR-CONTRACTS.md) | Define executable/reproducible domain contract | Canonical identity, typed semantics/scenarios, separate records and validity/lifecycle rules |
| [docs/02-domain/PROJECT-LIFECYCLE.md](../../docs/02-domain/PROJECT-LIFECYCLE.md) | Define executable/reproducible domain contract | Canonical identity, typed semantics/scenarios, separate records and validity/lifecycle rules |
| [docs/02-domain/TERMINOLOGY.md](../../docs/02-domain/TERMINOLOGY.md) | Define executable/reproducible domain contract | Canonical identity, typed semantics/scenarios, separate records and validity/lifecycle rules |
| [docs/03-analysis/ANALYSIS-SPEC.md](../../docs/03-analysis/ANALYSIS-SPEC.md) | Prevent false completeness/omission | Sealed acquisition, operational closure, mechanisms/causality and explicit uncertainty |
| [docs/03-analysis/CODE-CLASSIFICATION.md](../../docs/03-analysis/CODE-CLASSIFICATION.md) | Prevent false completeness/omission | Sealed acquisition, operational closure, mechanisms/causality and explicit uncertainty |
| [docs/03-analysis/RUNTIME-OBSERVATION.md](../../docs/03-analysis/RUNTIME-OBSERVATION.md) | Prevent false completeness/omission | Sealed acquisition, operational closure, mechanisms/causality and explicit uncertainty |
| [docs/03-analysis/WORKFLOW-DISCOVERY.md](../../docs/03-analysis/WORKFLOW-DISCOVERY.md) | Prevent false completeness/omission | Sealed acquisition, operational closure, mechanisms/causality and explicit uncertainty |
| [docs/04-migration/CI3-ADAPTER.md](../../docs/04-migration/CI3-ADAPTER.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/04-migration/DEPENDENCY-RESOLUTION.md](../../docs/04-migration/DEPENDENCY-RESOLUTION.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/04-migration/LARAVEL-ADAPTER.md](../../docs/04-migration/LARAVEL-ADAPTER.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/04-migration/MIGRATION-ENGINE.md](../../docs/04-migration/MIGRATION-ENGINE.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/04-migration/MIGRATION-UNITS.md](../../docs/04-migration/MIGRATION-UNITS.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/04-migration/TARGET-ARCHITECTURE.md](../../docs/04-migration/TARGET-ARCHITECTURE.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/04-migration/TARGET-SPECIFICATION.md](../../docs/04-migration/TARGET-SPECIFICATION.md) | Remove hidden target/acceptance decisions | Exact resolution, immutable choices/model, separate progress, scopes and planned adapter status |
| [docs/05-verification/COMPATIBILITY-SPEC.md](../../docs/05-verification/COMPATIBILITY-SPEC.md) | Make acceptance bounded and falsifiable | Independent replay/reset, relational normalization, scoped reports or explicit statistical claim |
| [docs/05-verification/COVERAGE.md](../../docs/05-verification/COVERAGE.md) | Make acceptance bounded and falsifiable | Independent replay/reset, relational normalization, scoped reports or explicit statistical claim |
| [docs/05-verification/DATABASE-COMPARISON.md](../../docs/05-verification/DATABASE-COMPARISON.md) | Make acceptance bounded and falsifiable | Independent replay/reset, relational normalization, scoped reports or explicit statistical claim |
| [docs/05-verification/DIFFERENTIAL-TESTING.md](../../docs/05-verification/DIFFERENTIAL-TESTING.md) | Make acceptance bounded and falsifiable | Independent replay/reset, relational normalization, scoped reports or explicit statistical claim |
| [docs/05-verification/GOLDEN-MASTER.md](../../docs/05-verification/GOLDEN-MASTER.md) | Make acceptance bounded and falsifiable | Independent replay/reset, relational normalization, scoped reports or explicit statistical claim |
| [docs/05-verification/PERFORMANCE.md](../../docs/05-verification/PERFORMANCE.md) | Make acceptance bounded and falsifiable | Independent replay/reset, relational normalization, scoped reports or explicit statistical claim |
| [docs/06-agent/AGENT-MODEL.md](../../docs/06-agent/AGENT-MODEL.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/AGENT-RULES.md](../../docs/06-agent/AGENT-RULES.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/CLI-SPEC.md](../../docs/06-agent/CLI-SPEC.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/ERROR-CATALOG.md](../../docs/06-agent/ERROR-CATALOG.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/EXECUTION-OBSERVABILITY.md](../../docs/06-agent/EXECUTION-OBSERVABILITY.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/MCP-SPEC.md](../../docs/06-agent/MCP-SPEC.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/MISSION-PROTOCOL.md](../../docs/06-agent/MISSION-PROTOCOL.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/06-agent/SKILL-SPEC.md](../../docs/06-agent/SKILL-SPEC.md) | Align every interface with guarded services | Typed decisions/background admission/current N-event M semantics and explicit trust |
| [docs/07-learning/EXPERIMENT-LEDGER.md](../../docs/07-learning/EXPERIMENT-LEDGER.md) | Prevent self-learning contamination | Independent quality/holdouts, pinned consumers, promotion/retraction and deferred activation |
| [docs/07-learning/LEARNING-SYSTEM.md](../../docs/07-learning/LEARNING-SYSTEM.md) | Prevent self-learning contamination | Independent quality/holdouts, pinned consumers, promotion/retraction and deferred activation |
| [docs/07-learning/SKILL-EVOLUTION.md](../../docs/07-learning/SKILL-EVOLUTION.md) | Prevent self-learning contamination | Independent quality/holdouts, pinned consumers, promotion/retraction and deferred activation |
| [docs/08-storage/ARTIFACT-STORE.md](../../docs/08-storage/ARTIFACT-STORE.md) | Make durable evidence operationally reconstructable | Incremental identity/ownership constraints, quotas/archive/backup and process recovery |
| [docs/08-storage/EVENT-MODEL.md](../../docs/08-storage/EVENT-MODEL.md) | Make durable evidence operationally reconstructable | Incremental identity/ownership constraints, quotas/archive/backup and process recovery |
| [docs/08-storage/RECOVERY.md](../../docs/08-storage/RECOVERY.md) | Make durable evidence operationally reconstructable | Incremental identity/ownership constraints, quotas/archive/backup and process recovery |
| [docs/08-storage/SQLITE-SCHEMA.md](../../docs/08-storage/SQLITE-SCHEMA.md) | Make durable evidence operationally reconstructable | Incremental identity/ownership constraints, quotas/archive/backup and process recovery |
| [docs/09-security/PERMISSION-MODEL.md](../../docs/09-security/PERMISSION-MODEL.md) | Bound safety/privacy guarantees | Private source/sanitized evidence, mediated same-account trust and safe effects |
| [docs/09-security/SANDBOX.md](../../docs/09-security/SANDBOX.md) | Bound safety/privacy guarantees | Private source/sanitized evidence, mediated same-account trust and safe effects |
| [docs/09-security/SECRET-HANDLING.md](../../docs/09-security/SECRET-HANDLING.md) | Bound safety/privacy guarantees | Private source/sanitized evidence, mediated same-account trust and safe effects |
| [docs/09-security/SECURITY.md](../../docs/09-security/SECURITY.md) | Bound safety/privacy guarantees | Private source/sanitized evidence, mediated same-account trust and safe effects |
| [docs/10-development/CUTOVER-RUNBOOK.md](../../docs/10-development/CUTOVER-RUNBOOK.md) | Define falsification evidence and honest validation | X01–X40 or build/release/rollback requirements with no fake results |
| [docs/10-development/DEVELOPMENT.md](../../docs/10-development/DEVELOPMENT.md) | Define falsification evidence and honest validation | X01–X40 or build/release/rollback requirements with no fake results |
| [docs/10-development/FIXTURE-CORPUS.md](../../docs/10-development/FIXTURE-CORPUS.md) | Define falsification evidence and honest validation | X01–X40 or build/release/rollback requirements with no fake results |
| [docs/10-development/RELEASE.md](../../docs/10-development/RELEASE.md) | Define falsification evidence and honest validation | X01–X40 or build/release/rollback requirements with no fake results |
| [docs/10-development/TESTING.md](../../docs/10-development/TESTING.md) | Define falsification evidence and honest validation | X01–X40 or build/release/rollback requirements with no fake results |
| [docs/11-roadmap/BACKLOG.md](../../docs/11-roadmap/BACKLOG.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [docs/11-roadmap/DECISION-REGISTER.md](../../docs/11-roadmap/DECISION-REGISTER.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [docs/11-roadmap/DEVELOPMENT-PLAN.md](../../docs/11-roadmap/DEVELOPMENT-PLAN.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [docs/11-roadmap/DOCUMENTATION-AUDIT.md](../../docs/11-roadmap/DOCUMENTATION-AUDIT.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [docs/11-roadmap/MILESTONES.md](../../docs/11-roadmap/MILESTONES.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [docs/11-roadmap/ROADMAP.md](../../docs/11-roadmap/ROADMAP.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [docs/AGENT-WORKFLOW.md](../../docs/AGENT-WORKFLOW.md) | Eliminate duplicate authority/stale summaries | Canonical owner reading guide and evidence-only contributor workflow |
| [docs/README.md](../../docs/README.md) | Eliminate duplicate authority/stale summaries | Canonical owner reading guide and evidence-only contributor workflow |
| [specs/README.md](../../specs/README.md) | Clarify machine-contract authority/versioning | Current v2 execution versus historical v1, local schemas and structural-check limits |
| [specs/adapters/manifest.schema.json](../../specs/adapters/manifest.schema.json) | Represent required machine concept: manifest | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/behavior/behavior-case.schema.json](../../specs/behavior/behavior-case.schema.json) | Represent required machine concept: behavior-case | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/behavior/behavior-contract.schema.json](../../specs/behavior/behavior-contract.schema.json) | Represent required machine concept: behavior-contract | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/behavior/observation-scenario.schema.json](../../specs/behavior/observation-scenario.schema.json) | Represent required machine concept: observation-scenario | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/behavior/workflow.schema.json](../../specs/behavior/workflow.schema.json) | Represent required machine concept: workflow | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/cli/commands.json](../../specs/cli/commands.json) | Synchronize public machine catalogs | Decision operations, contract version and async admission output variants |
| [specs/cli/commands/analyze.input.schema.json](../../specs/cli/commands/analyze.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/cli/commands/benchmark.input.schema.json](../../specs/cli/commands/benchmark.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/cli/commands/discover.input.schema.json](../../specs/cli/commands/discover.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/cli/commands/migrate.input.schema.json](../../specs/cli/commands/migrate.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/cli/commands/observe.input.schema.json](../../specs/cli/commands/observe.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/cli/commands/verify.input.schema.json](../../specs/cli/commands/verify.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/common.schema.json](../../specs/common.schema.json) | Represent required machine concept: common | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/examples/validation-cases.json](../../specs/examples/validation-cases.json) | Add current positive/negative evidence fixtures | v2 structures and accepted/rejected canonical approval vector bytes/digests |
| [specs/execution/event.schema.json](../../specs/execution/event.schema.json) | Represent required machine concept: event | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/execution/page.schema.json](../../specs/execution/page.schema.json) | Represent required machine concept: page | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/mcp/tools.json](../../specs/mcp/tools.json) | Synchronize public machine catalogs | Decision operations, contract version and async admission output variants |
| [specs/migration/dependency-resolution.schema.json](../../specs/migration/dependency-resolution.schema.json) | Represent required machine concept: dependency-resolution | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/migration/plan.schema.json](../../specs/migration/plan.schema.json) | Represent required machine concept: plan | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/migration/release.schema.json](../../specs/migration/release.schema.json) | Represent required machine concept: release | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/operations/result.schema.json](../../specs/operations/result.schema.json) | Represent required machine concept: result | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/prepared-confirmation.schema.json](../../specs/project/prepared-confirmation.schema.json) | Represent required machine concept: prepared-confirmation | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/target.schema.json](../../specs/project/target.schema.json) | Represent required machine concept: target | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/schemas/trace-schema.json](../../specs/schemas/trace-schema.json) | Represent required machine concept: trace-schema.json | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/verification/benchmark.schema.json](../../specs/verification/benchmark.schema.json) | Represent required machine concept: benchmark | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/verification/report.schema.json](../../specs/verification/report.schema.json) | Represent required machine concept: report | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [tools/docs/validate.py](../../tools/docs/validate.py) | Verify specification artifacts structurally | Both migrations, canonical vectors, new invariants/links and stale-claim checks; no engine implementation |

## B. Necessary new files

New records address concepts absent from existing machine shapes; they do not create competing topic specifications. ADR0014/0015 cover decisions not owned by previous ADRs. No new Skill/tool or production source is created.

| File | Reason | Major changes |
| --- | --- | --- |
| [docs/01-architecture/ADR/0014-learning-independence.md](../../docs/01-architecture/ADR/0014-learning-independence.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/01-architecture/ADR/0015-statistical-performance.md](../../docs/01-architecture/ADR/0015-statistical-performance.md) | Record materially constraining decision and alternatives | Canonical owner, machine/compatibility impact and falsification obligation |
| [docs/11-roadmap/RED-TEAM-RESOLUTION.md](../../docs/11-roadmap/RED-TEAM-RESOLUTION.md) | Make one current implementation queue/audit owner | Milestone prerequisites/scope/evidence/non-goals or explicit dispositions/deferred limits |
| [migrations/0002_revision_contracts.sql](../../migrations/0002_revision_contracts.sql) | Represent new durable records without rewriting 0001 | Five tables, immutable metadata, ownership triggers and uniqueness constraints |
| [specs/analysis/omission-proof.schema.json](../../specs/analysis/omission-proof.schema.json) | Represent required machine concept: omission-proof | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/behavior/request.schema.json](../../specs/behavior/request.schema.json) | Represent required machine concept: request | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/behavior/semantics.schema.json](../../specs/behavior/semantics.schema.json) | Represent required machine concept: semantics | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/cli/commands/decision-propose.input.schema.json](../../specs/cli/commands/decision-propose.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/cli/commands/decision-show.input.schema.json](../../specs/cli/commands/decision-show.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/examples/approval-hash-vectors.json](../../specs/examples/approval-hash-vectors.json) | Add current positive/negative evidence fixtures | v2 structures and accepted/rejected canonical approval vector bytes/digests |
| [specs/mcp/tools/decisions.get.input.schema.json](../../specs/mcp/tools/decisions.get.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/mcp/tools/decisions.propose.input.schema.json](../../specs/mcp/tools/decisions.propose.input.schema.json) | Represent interface arguments | Decision artifact inputs and explicit background admission |
| [specs/migration/progress.schema.json](../../specs/migration/progress.schema.json) | Represent required machine concept: progress | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/migration/target-model.schema.json](../../specs/migration/target-model.schema.json) | Represent required machine concept: target-model | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/observation/attestation.schema.json](../../specs/observation/attestation.schema.json) | Represent required machine concept: attestation | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/observation/causality.schema.json](../../specs/observation/causality.schema.json) | Represent required machine concept: causality | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/observation/effects.schema.json](../../specs/observation/effects.schema.json) | Represent required machine concept: effects | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/observation/profile.schema.json](../../specs/observation/profile.schema.json) | Represent required machine concept: profile | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/observation/reset-report.schema.json](../../specs/observation/reset-report.schema.json) | Represent required machine concept: reset-report | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/operations/archive.schema.json](../../specs/operations/archive.schema.json) | Represent required machine concept: archive | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/approval-subject.schema.json](../../specs/project/approval-subject.schema.json) | Represent required machine concept: approval-subject | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/decision.schema.json](../../specs/project/decision.schema.json) | Represent required machine concept: decision | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/resolved-environment.schema.json](../../specs/project/resolved-environment.schema.json) | Represent required machine concept: resolved-environment | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/reuse-certificate.schema.json](../../specs/project/reuse-certificate.schema.json) | Represent required machine concept: reuse-certificate | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/source-snapshot.schema.json](../../specs/project/source-snapshot.schema.json) | Represent required machine concept: source-snapshot | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/project/validity-graph.schema.json](../../specs/project/validity-graph.schema.json) | Represent required machine concept: validity-graph | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/verification/acceptance-scope.schema.json](../../specs/verification/acceptance-scope.schema.json) | Represent required machine concept: acceptance-scope | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/verification/identity-map.schema.json](../../specs/verification/identity-map.schema.json) | Represent required machine concept: identity-map | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |
| [specs/verification/normalization.schema.json](../../specs/verification/normalization.schema.json) | Represent required machine concept: normalization | Closed v2 fields/refs/lifecycle; historical v1 retained where compatibility requires |

Inventory: 107 changed existing files; 29 new files. The report itself is included.

## C–G. Review disposition, blockers, order, experiments and contradictions

C: The matrices above give every CR-01–07, F01–51 and D01–18 a disposition, including rejected/already-covered recommendations with owners/reasons. RESOLVED is specification status only. F11/F25/F35/F47 remain deferred capability expansion with explicit refusal behavior.

D: Remaining P0/P1/P2/P3 evidence/capability blockers are identified above. No unresolved foundation design choice is concealed as a runtime guarantee. M1 premise evidence remains a mandatory gate before broader work.

E: M0–M8 order and detailed objective/prerequisites/inputs/outputs/specs/scope/tests/acceptance/evidence/blockers/non-goals are in DEVELOPMENT-PLAN. Legacy RWT aliases are not a parallel queue.

F: TESTING specifies X01–X40 and M1's 17 mandatory premise checks. All require actual later execution before their claims. This revision executes only structural validation.

G: No unresolved authoritative contradiction was found in the revised claim/gate/scope/target/process/history/retention/learning/maturity/transport concepts. Historical v1 schemas and the unchanged original review are explicitly historical. Deferred capabilities are stated limitations, not alternate canonical answers.

## Executed validation and limits

Final structural results are recorded after the file manifest is checked. No Go/race/driver/container/adapter/MCP-client/security/runtime/load experiment ran. No engine implementation was started.

Additional schema defect found during revision: the v1 verification report dimension object required a literal `$ref` it did not permit. Current v2 uses typed applicable/verdict/reason/evidence records, rejects PASS containing UNKNOWN/MISMATCH and requires populated completion scope. Legacy v1 is retained solely as non-executable historical shape. This defect is RESOLVED in verification/report.schema and DIFFERENTIAL-TESTING; new structural fixtures cover it.

Final executed commands: `python3 tools/docs/validate.py` and `git diff --check`, both exit 0. Validator: 120 schemas; 130 positive/negative fixtures; 66 CLI/MCP catalog entries; 445 local links; 40 SQL tables across migrations 0001/0002; eight accepted/rejected approval vectors; ownership/immutability/rollback checks. This is structural evidence only.

Structural-check host: Python 3.12.3, Python sqlite3 engine 3.45.1; in-memory DDL checks do not validate the specified future modernc/SQLite build, WAL/filesystem/runtime or supervisor behavior.

Unchanged input review SHA-256: `1b4ca30d5f310213402bdf49dcd48efdde06c864c8347cea50de11b3d152bcc2`. Unchanged migration 0001 SHA-256: `b57b0f81f11a5eaef742ed64adb1b03b255966e23969c60970b14ae3b39d796c`.

## Assessment checkpoint follow-up — 2026-10-03

This follow-up implements requirement 41 in the specifications, not in a running engine. The preceding file counts/validation results describe the earlier red-team revision and remain historical. [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md) is the sole assessment checkpoint owner; [DEVELOPMENT-PLAN](DEVELOPMENT-PLAN.md) remains the sole implementation queue.

### Requirement dispositions

| Requirement group | Decision / resolution | Owner | Status |
| --- | --- | --- | --- |
| One workflow / successful assessment-only stop | Shared analysis services and records; checkpoint can terminate without an executable plan, target publication or implementation missions | PROJECT-LIFECYCLE; SYSTEM-DESIGN | RESOLVED in specification |
| Useful report with blocked/unknown migration | Independent COMPLETED assessment and bounded feasibility; incomplete/static-only evidence is classified with provenance and limits | PROJECT-LIFECYCLE; assessment-report schema | RESOLVED in specification |
| Report contents and current versus history | Immutable assessed-state projection plus a separate current validity view; effort only where supported | PROJECT-LIFECYCLE; report/view schemas | RESOLVED in specification |
| Explicit user decision / independent approval | Local human selection before a separate plan request; no MCP selection/approval minting; Gate 2 still authorizes implementation | CONFIRMATION-PROTOCOL; CLI/MCP specs | RESOLVED in specification |
| Optional continuation and selective reuse | Original analysis IDs plus typed predicates/witnesses; stale only affected findings; preserve historical evidence | DOMAIN-MODEL; validity graph/planning basis | RESOLVED in specification |
| Derived executable plan / changes since assessment | Plan envelope 2.1 binds selection and basis via unchanged v2 approval subject decision_refs; acyclic immutable closure | MIGRATION-ENGINE; plan schema | RESOLVED in specification |
| No executable-corpus prerequisite for scope | Shared work-scope artifact records snapshot/workflow/constraint/business inputs before acceptance cases exist | PROJECT-LIFECYCLE; work-scope schema | RESOLVED in specification |
| Fifteen required acceptance tests | X41–X55 specify non-mutation, non-authorization, shared records, blocked retrieval, classification, drift and reuse | TESTING; M0/M1/M2/M3/M5 | RESOLVED as planned experiments; runtime pending |

No requirement was rejected or assigned to a separate assessment engine. New schema refs use existing immutable artifacts/events/indexes; neither SQL migration was changed. Existing ADRs 0011/0012 were updated instead of creating competing decisions. M1 remains the early wrong-target falsification milestone.

### Changed existing files

Reasons and changes below are scoped to this assessment follow-up, not the preceding red-team revision.

| File | Reason / major changes |
| --- | --- |
| [AGENTS.md](../../AGENTS.md) | Constitution: shared checkpoint, independent completion and explicit continuation without implementation authority. |
| [README.md](../../README.md) | Current overview and workflow/order now include assessment, optional stop and separate plan approval. |
| [REWROTRON_PROJECT_PLAN.md](../../REWROTRON_PROJECT_PLAN.md) | Current overview and workflow/order now include assessment, optional stop and separate plan approval. |
| [docs/00-vision/GLOSSARY.md](../../docs/00-vision/GLOSSARY.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/00-vision/PRODUCT.md](../../docs/00-vision/PRODUCT.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/01-architecture/ADR/0011-development-baseline-and-contract-completion.md](../../docs/01-architecture/ADR/0011-development-baseline-and-contract-completion.md) | Existing accepted decision extended for shared checkpoint sequencing, immutable selections and falsification. |
| [docs/01-architecture/ADR/0012-target-specification-and-dependency-resolution.md](../../docs/01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) | Existing accepted decision extended for shared checkpoint sequencing, immutable selections and falsification. |
| [docs/01-architecture/COMPONENTS.md](../../docs/01-architecture/COMPONENTS.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/01-architecture/SERVICE-CONTRACTS.md](../../docs/01-architecture/SERVICE-CONTRACTS.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/01-architecture/SYSTEM-DESIGN.md](../../docs/01-architecture/SYSTEM-DESIGN.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/02-domain/CONFIRMATION-PROTOCOL.md](../../docs/02-domain/CONFIRMATION-PROTOCOL.md) | Typed report/finding reuse, current versus history and checkpoint refs in immutable plan closure. |
| [docs/02-domain/DOMAIN-MODEL.md](../../docs/02-domain/DOMAIN-MODEL.md) | Typed report/finding reuse, current versus history and checkpoint refs in immutable plan closure. |
| [docs/02-domain/PROJECT-LIFECYCLE.md](../../docs/02-domain/PROJECT-LIFECYCLE.md) | Canonical completion/report/selection/continuation, feasibility, scope, identity, invalidation and cross-record rules. |
| [docs/03-analysis/ANALYSIS-SPEC.md](../../docs/03-analysis/ANALYSIS-SPEC.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/04-migration/DEPENDENCY-RESOLUTION.md](../../docs/04-migration/DEPENDENCY-RESOLUTION.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/04-migration/MIGRATION-ENGINE.md](../../docs/04-migration/MIGRATION-ENGINE.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/04-migration/TARGET-SPECIFICATION.md](../../docs/04-migration/TARGET-SPECIFICATION.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/AGENT-MODEL.md](../../docs/06-agent/AGENT-MODEL.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/CLI-SPEC.md](../../docs/06-agent/CLI-SPEC.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/ERROR-CATALOG.md](../../docs/06-agent/ERROR-CATALOG.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/EXECUTION-OBSERVABILITY.md](../../docs/06-agent/EXECUTION-OBSERVABILITY.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/MCP-SPEC.md](../../docs/06-agent/MCP-SPEC.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/MISSION-PROTOCOL.md](../../docs/06-agent/MISSION-PROTOCOL.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/06-agent/SKILL-SPEC.md](../../docs/06-agent/SKILL-SPEC.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/08-storage/ARTIFACT-STORE.md](../../docs/08-storage/ARTIFACT-STORE.md) | Assessment artifact indexing/atomic publication/retention through existing storage; no DDL change. |
| [docs/08-storage/SQLITE-SCHEMA.md](../../docs/08-storage/SQLITE-SCHEMA.md) | Assessment artifact indexing/atomic publication/retention through existing storage; no DDL change. |
| [docs/09-security/PERMISSION-MODEL.md](../../docs/09-security/PERMISSION-MODEL.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/10-development/FIXTURE-CORPUS.md](../../docs/10-development/FIXTURE-CORPUS.md) | Fifteen X41–X55 falsification experiments and shared reference fixture variants; execution remains pending. |
| [docs/10-development/PROJECT-INPUTS.md](../../docs/10-development/PROJECT-INPUTS.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/10-development/TESTING.md](../../docs/10-development/TESTING.md) | Fifteen X41–X55 falsification experiments and shared reference fixture variants; execution remains pending. |
| [docs/11-roadmap/DECISION-REGISTER.md](../../docs/11-roadmap/DECISION-REGISTER.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/11-roadmap/DEVELOPMENT-PLAN.md](../../docs/11-roadmap/DEVELOPMENT-PLAN.md) | Assessment outputs/prerequisites/tests added to existing milestones without a parallel queue or delaying M1. |
| [docs/11-roadmap/DOCUMENTATION-AUDIT.md](../../docs/11-roadmap/DOCUMENTATION-AUDIT.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [docs/11-roadmap/RED-TEAM-RESOLUTION.md](../../docs/11-roadmap/RED-TEAM-RESOLUTION.md) | Follow-up disposition, file manifest, evidence limits and remaining experiments. |
| [docs/11-roadmap/ROADMAP.md](../../docs/11-roadmap/ROADMAP.md) | Assessment outputs/prerequisites/tests added to existing milestones without a parallel queue or delaying M1. |
| [docs/README.md](../../docs/README.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [specs/README.md](../../specs/README.md) | Align the existing owning or supporting specification with shared assessment, explicit continuation and evidence limits. |
| [specs/cli/commands.json](../../specs/cli/commands.json) | Catalog 2.1 for guarded assessment coordination/inspection and local human selection. |
| [specs/cli/commands/plan.input.schema.json](../../specs/cli/commands/plan.input.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/cli/commands/report.input.schema.json](../../specs/cli/commands/report.input.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/examples/approval-hash-vectors.json](../../specs/examples/approval-hash-vectors.json) | Add one checkpoint-derived v2 subject vector; preserve all eight original vectors. |
| [specs/examples/validation-cases.json](../../specs/examples/validation-cases.json) | Positive/negative synthetic checkpoint, classification, selection, scope, validity and plan fixtures. |
| [specs/mcp/tools.json](../../specs/mcp/tools.json) | Catalog 2.1 for guarded assessment coordination/inspection and local human selection. |
| [specs/mcp/tools/migration.plan.input.schema.json](../../specs/mcp/tools/migration.plan.input.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/mcp/tools/reports.generate.input.schema.json](../../specs/mcp/tools/reports.generate.input.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/migration/plan.schema.json](../../specs/migration/plan.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/project/approval-subject.schema.json](../../specs/project/approval-subject.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/project/snapshot.schema.json](../../specs/project/snapshot.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |
| [specs/project/validity-graph.schema.json](../../specs/project/validity-graph.schema.json) | Align machine contract with checkpoint refs, rendering, validity or versioned current plan; retain historical shapes. |

### Necessary new files

| File | Purpose |
| --- | --- |
| [specs/cli/commands/assess.input.schema.json](../../specs/cli/commands/assess.input.schema.json) | Typed CLI/MCP input for shared checkpoint services; no implementation. |
| [specs/cli/commands/assessment-decide.input.schema.json](../../specs/cli/commands/assessment-decide.input.schema.json) | Typed CLI/MCP input for shared checkpoint services; no implementation. |
| [specs/cli/commands/assessment-show.input.schema.json](../../specs/cli/commands/assessment-show.input.schema.json) | Typed CLI/MCP input for shared checkpoint services; no implementation. |
| [specs/mcp/tools/assessment.get.input.schema.json](../../specs/mcp/tools/assessment.get.input.schema.json) | Typed CLI/MCP input for shared checkpoint services; no implementation. |
| [specs/mcp/tools/assessment.run.input.schema.json](../../specs/mcp/tools/assessment.run.input.schema.json) | Typed CLI/MCP input for shared checkpoint services; no implementation. |
| [specs/migration/assessment-planning-basis.schema.json](../../specs/migration/assessment-planning-basis.schema.json) | Immutable changed/reused finding inputs bound by plan approval. |
| [specs/project/assessment-selection.schema.json](../../specs/project/assessment-selection.schema.json) | Explicit immutable human work decision, never implementation approval. |
| [specs/project/work-scope.schema.json](../../specs/project/work-scope.schema.json) | Shared pre-plan scope without an executable corpus. |
| [specs/reports/assessment-report.schema.json](../../specs/reports/assessment-report.schema.json) | Persisted report or separately evaluated applicability; no plan/approval prerequisite. |
| [specs/reports/assessment-view.schema.json](../../specs/reports/assessment-view.schema.json) | Persisted report or separately evaluated applicability; no plan/approval prerequisite. |

### Remaining blockers, order and validation limits

P0/P1: no newly unresolved assessment design decision identified. Actual human-channel enforcement, source/production non-mutation, atomic completion, typed selective reuse and approval-token rejection are mandatory implementation evidence, not proven by these documents. P2/P3 capability deferrals from the prior review remain unchanged. Missing runtime/dependency capability can block migration while still permitting a useful assessment.

Order remains M0 → M1 → M2 → M3 → M4 → M5 → M6, optional M7, later M8. M0 persists checkpoint records; M1 tests stopping/resuming the same stateful workflow and rejects wrong reconstructions; M2 reports real source evidence/gaps; M3 resolves explicitly selected work into approved exact plans; M5 tests interface parity.

Contradiction scan corrected overview pipeline bypass, generic completion-as-packaging wording, Skill completion requiring verification for assessment, automatic successor-plan wording, plan permission prerequisites and FINALIZE display scope. No conflicting answer remains identified for these revised concepts. Historical red-team results and old schema/plan versions are labeled historical; an assessment applicability claim remains distinct from verification PASS.

Executed structural validation: `python3 tools/docs/validate.py` and `git diff --check`, exit 0 before this audit appendix; rerun after publication is required. At that check: 130 schemas, 178 positive/negative fixtures, 71 CLI/MCP catalog entries, 474 local links, 40 SQL tables across the same two migrations and nine approval vectors. Go/runtime/agent/sandbox/client tests were not run. X41–X55 are acceptance specifications, not passing runtime tests. Original input review and migrations remain byte-identical; no engine, adapter, production Go, new executable tool or Skill was implemented.

Post-publication rerun passed: 130 schemas, 178 positive/negative fixtures, 71 catalog entries, 535 local links, the same 40 SQL tables and nine approval vectors. `git diff --check` passed. All fifteen X41–X55 entries are present, all eight original vectors are unchanged, and the scope check found no Go or SQL migration changes. This follow-up changed 49 existing files and added ten necessary specification files; runtime acceptance remains pending.

## Current re-audit and revision deliverables

This section is the latest revision checkpoint and supersedes earlier validation counts/file manifests. Historical entries above remain an audit of previous passes, not current schema-version requirements. The current delta is measured against repository bytes captured at the start of this revision: **49 changed existing files and three new schemas**. Prior assessment work is preserved. The unchanged review is at docs/Rewrotron_Pre-Implementation_Red-Team_Review.md; SHA256 1b4ca30d5f310213402bdf49dcd48efdde06c864c8347cea50de11b3d152bcc2.

### A. Changed files

| File | Reason | Major changes |
| --- | --- | --- |
| [AGENTS.md](../../AGENTS.md) | Keep constitution aligned | Require engine/build/schema revision pins; retain concise owner references |
| [REWROTRON_PROJECT_PLAN.md](../../REWROTRON_PROJECT_PLAN.md) | Replace vague sequencing with verifiable prerequisites | Execution-bound approvals, early interpreter mutants, approved policies and honest M7/M8 boundaries |
| [docs/00-vision/PRODUCT.md](../../docs/00-vision/PRODUCT.md) | Bound reproducibility claim | Delegate exact engine/schema/control qualification to approval owner |
| [docs/01-architecture/ADR/0003-behavior-as-canonical-model.md](../../docs/01-architecture/ADR/0003-behavior-as-canonical-model.md) | Record implementation-constraining decision | behavior as canonical model: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0005-differential-verification.md](../../docs/01-architecture/ADR/0005-differential-verification.md) | Record implementation-constraining decision | differential verification: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0008-mcp-as-agent-interface.md](../../docs/01-architecture/ADR/0008-mcp-as-agent-interface.md) | Record implementation-constraining decision | mcp as agent interface: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0009-three-tier-ir-architecture.md](../../docs/01-architecture/ADR/0009-three-tier-ir-architecture.md) | Record implementation-constraining decision | three tier ir architecture: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0010-mandatory-user-confirmation-gates.md](../../docs/01-architecture/ADR/0010-mandatory-user-confirmation-gates.md) | Record implementation-constraining decision | mandatory user confirmation gates: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0011-development-baseline-and-contract-completion.md](../../docs/01-architecture/ADR/0011-development-baseline-and-contract-completion.md) | Record implementation-constraining decision | development baseline and contract completion: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0014-learning-independence.md](../../docs/01-architecture/ADR/0014-learning-independence.md) | Record implementation-constraining decision | learning independence: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/ADR/0015-statistical-performance.md](../../docs/01-architecture/ADR/0015-statistical-performance.md) | Record implementation-constraining decision | statistical performance: specific problem/choice/machine impact/compatibility and falsification experiment; all nine ADR sections retained |
| [docs/01-architecture/SERVICE-CONTRACTS.md](../../docs/01-architecture/SERVICE-CONTRACTS.md) | Define actual admission evidence | Exact qualification records, current-version guards, no Gate-1 control cycle, descriptive profiles cannot authorize |
| [docs/01-architecture/SYSTEM-DESIGN.md](../../docs/01-architecture/SYSTEM-DESIGN.md) | Align core pipeline with current services | Execution closure, qualified DB lowering, immutable performance decisions and shared assessment |
| [docs/02-domain/BEHAVIOR-MODEL.md](../../docs/02-domain/BEHAVIOR-MODEL.md) | Eliminate undefined execution semantics | Row context/write output binding, boolean/null/pointer/decimal rules, bounded mutations and explicit transaction terminals |
| [docs/02-domain/CONFIRMATION-PROTOCOL.md](../../docs/02-domain/CONFIRMATION-PROTOCOL.md) | Freeze execution and policy identity | Core RUNNER manifest and performance policy in unchanged v2 preimage; version-aware restart and not-applicable release evidence |
| [docs/02-domain/DOMAIN-MODEL.md](../../docs/02-domain/DOMAIN-MODEL.md) | Represent invalidation of newly pinned inputs | Qualification/execution/performance typed edges and current witnesses |
| [docs/02-domain/IR-CONTRACTS.md](../../docs/02-domain/IR-CONTRACTS.md) | Align neutral child record identities | Qualified data models and pinned schema/qualification bytes; no implicit child upgrades |
| [docs/02-domain/PROJECT-LIFECYCLE.md](../../docs/02-domain/PROJECT-LIFECYCLE.md) | Align Gate 2/3 readiness | Qualified execution/performance inputs; explicit not-applicable readiness without fake measurements |
| [docs/03-analysis/RUNTIME-OBSERVATION.md](../../docs/03-analysis/RUNTIME-OBSERVATION.md) | Make missing observation representable | Current 2.1 attestation branches, independent status derivation and historical reader limits |
| [docs/04-migration/MIGRATION-ENGINE.md](../../docs/04-migration/MIGRATION-ENGINE.md) | Bind all planning choices | Require immutable performance policy before plan approval |
| [docs/04-migration/TARGET-SPECIFICATION.md](../../docs/04-migration/TARGET-SPECIFICATION.md) | Separate target and policy authority | User performance choices use a separate immutable record and guarded plan input |
| [docs/05-verification/DATABASE-COMPARISON.md](../../docs/05-verification/DATABASE-COMPARISON.md) | Prevent same-table identity/type confusion | Qualified model/relationships, native-to-canonical mapping evidence and UNKNOWN columns |
| [docs/05-verification/PERFORMANCE.md](../../docs/05-verification/PERFORMANCE.md) | Make acceptance reproducible | Approved typed policy, paired estimand/seeded bootstrap, exact outcomes/sample/units/error controls |
| [docs/06-agent/CLI-SPEC.md](../../docs/06-agent/CLI-SPEC.md) | Keep complete guarded fallback | Canonical capability error and contained performance-policy proposal/import/reuse protocol |
| [docs/06-agent/ERROR-CATALOG.md](../../docs/06-agent/ERROR-CATALOG.md) | Remove competing public codes | ERR_UNSUPPORTED_CAPABILITY is the sole capability refusal name |
| [docs/06-agent/MCP-SPEC.md](../../docs/06-agent/MCP-SPEC.md) | Remove hidden online schema dependency | Offline recursive schema bundles, uniform errors and performance-policy parity |
| [docs/07-learning/KNOWLEDGE-MODEL.md](../../docs/07-learning/KNOWLEDGE-MODEL.md) | Expose missing future machine contracts | M7 global revision/evaluation/consumer/retraction/outbox prerequisites; project-only baseline |
| [docs/08-storage/RECOVERY.md](../../docs/08-storage/RECOVERY.md) | Keep recovery error parity | Use the single canonical capability error |
| [docs/08-storage/SQLITE-SCHEMA.md](../../docs/08-storage/SQLITE-SCHEMA.md) | Map added immutable records to storage | Existing artifacts/links/approval closure; no applied migration rewrite or fake global tables |
| [docs/10-development/FIXTURE-CORPUS.md](../../docs/10-development/FIXTURE-CORPUS.md) | Align fixture obligations | X01–X60 and synthetic qualification controls before customer execution |
| [docs/10-development/TESTING.md](../../docs/10-development/TESTING.md) | Specify meaningful falsification | X56–X60: row semantics, unavailable capture, execution pins, policy estimand and offline schemas |
| [docs/11-roadmap/DECISION-REGISTER.md](../../docs/11-roadmap/DECISION-REGISTER.md) | Update pending decision status | Current branch map and explicit M7/M8 prerequisites remain separate from evidence |
| [docs/11-roadmap/DEVELOPMENT-PLAN.md](../../docs/11-roadmap/DEVELOPMENT-PLAN.md) | Make milestones implementation-ready | Concrete objectives, assigned new tests, exact decision outputs and global/profile prerequisites |
| [docs/11-roadmap/DOCUMENTATION-AUDIT.md](../../docs/11-roadmap/DOCUMENTATION-AUDIT.md) | Record audit limits | Current contract corrections; runtime evidence remains pending |
| [docs/11-roadmap/RED-TEAM-RESOLUTION.md](../../docs/11-roadmap/RED-TEAM-RESOLUTION.md) | Provide reviewable A–G deliverables | Current delta, all material dispositions, priority boundaries, assumption pass and verification limits |
| [docs/11-roadmap/ROADMAP.md](../../docs/11-roadmap/ROADMAP.md) | Align single queue | Execution-bound M0, performance-bound M3, exact-profile/project-only baseline |
| [docs/README.md](../../docs/README.md) | Fix topic ownership | Separate lifecycle readiness from canonical approval/execution closure |
| [specs/README.md](../../specs/README.md) | Remove contradictory blanket versions | Catalog 2.1 and per-record current-version map; offline schemas and historical read distinction |
| [specs/adapters/manifest.schema.json](../../specs/adapters/manifest.schema.json) | Pin machine support evidence | Current 2.1 schema/qualification refs; preserve prior branches |
| [specs/behavior/semantics.schema.json](../../specs/behavior/semantics.schema.json) | Represent actual interpreter inputs | row namespace, optional write bind and positive asynchronous quiescence |
| [specs/behavior/workflow.schema.json](../../specs/behavior/workflow.schema.json) | Close async boundary loophole | Positive quiescence for current sink-ack completion |
| [specs/cli/commands/plan.input.schema.json](../../specs/cli/commands/plan.input.schema.json) | Expose policy selection in CLI | Typed performance_policy_ref; domain guards distinguish syntax from current planning authority |
| [specs/examples/approval-hash-vectors.json](../../specs/examples/approval-hash-vectors.json) | Demonstrate new pins without breaking history | One additional core/policy-bound vector; previous nine unchanged |
| [specs/examples/validation-cases.json](../../specs/examples/validation-cases.json) | Test current shape rejection | 44 additional synthetic positive/negative cases; no runtime results claimed |
| [specs/ir/data-model.schema.json](../../specs/ir/data-model.schema.json) | Qualify table/type identity | Current 2.0 model, typed destination relationships/value mappings and honest unknown types |
| [specs/mcp/tools/migration.plan.input.schema.json](../../specs/mcp/tools/migration.plan.input.schema.json) | Keep MCP planning parity | Same immutable policy ref without approval/threshold bypass |
| [specs/observation/attestation.schema.json](../../specs/observation/attestation.schema.json) | Distinguish missing capture from closure | Current 2.1 UNAVAILABLE/PARTIAL/COMPLETE/NOT_APPLICABLE shape constraints |
| [specs/project/validity-graph.schema.json](../../specs/project/validity-graph.schema.json) | Type reverse invalidation inputs | Implementation qualification, execution result and performance result edge kinds |
| [specs/verification/benchmark.schema.json](../../specs/verification/benchmark.schema.json) | Bind actual benchmark acceptance | Current 2.1 performance_policy_ref/native units/claim consistency; historical branches retained |

### B. Necessary new files

| File | Need / owning contract |
| --- | --- |
| [specs/project/execution-manifest.schema.json](../../specs/project/execution-manifest.schema.json) | Core/build/dependency/schema/canonicalizer byte identity was not representable as an adapter version; CONFIRMATION-PROTOCOL owns it |
| [specs/adapters/qualification.schema.json](../../specs/adapters/qualification.schema.json) | Exact implementation/schema/profile/control/result support must be a typed record rather than advertised open metadata; SERVICE-CONTRACTS owns it |
| [specs/verification/performance-policy.schema.json](../../specs/verification/performance-policy.schema.json) | Claim/method/workload/metric limits and explicit not-applicability must be immutable reviewed decisions; PERFORMANCE owns it |

No new ADR, Go file, executable tool, Skill or SQL migration was created. Existing artifact/reference storage represents these records; migrations 0001/0002 remain byte-identical.

### C. Red-team disposition and impact map

The original CR-01–CR-07, F01–F51 and D01–D18 dispositions above remain the full review matrix, with the clarified global-contract deferrals below. No finding is silently dropped. Current re-audit findings trace the remaining concrete mismatches to those original IDs:

| Finding | Decision | Owning document | Resolution | Status |
| --- | --- | --- | --- | --- |
| R2-01 (CR-04, F06, F12, D05) | Accept concrete contract defect | [docs/03-analysis/RUNTIME-OBSERVATION.md](../../docs/03-analysis/RUNTIME-OBSERVATION.md) | Versioned current attestation distinguishes missing evidence from closed coverage; current COMPLETE conditions and independent derivation | RESOLVED |
| R2-02 (F16, D03, D12) | Accept concrete contract defect | [docs/05-verification/DATABASE-COMPARISON.md](../../docs/05-verification/DATABASE-COMPARISON.md) | Current v2 qualified model and typed cross-connection relationships; legacy descriptive shape retained | RESOLVED |
| R2-03 (CR-03, F09, D03, D04) | Accept concrete contract defect | [docs/02-domain/BEHAVIOR-MODEL.md](../../docs/02-domain/BEHAVIOR-MODEL.md) | Bounded operator/context/action semantics, optional additive row/bind fields with capability guards and unsupported failure paths | RESOLVED |
| R2-04 (F29, F47, D12) | Narrow current support; defer broader machine capability | [docs/01-architecture/SERVICE-CONTRACTS.md](../../docs/01-architecture/SERVICE-CONTRACTS.md) | Initial support established only by exact-profile qualification controls/typed ports; generic constraint profile language deferred before M8 | DEFERRED |
| R2-05 (F49, D15) | Accept concrete contract defect | [docs/05-verification/PERFORMANCE.md](../../docs/05-verification/PERFORMANCE.md) | Immutable typed performance policy in PLAN decision closure; v2.1 benchmark subject; precise paired estimand and outcomes | RESOLVED |
| R2-06 (D01, D05) | Accept concrete contract defect | [docs/06-agent/ERROR-CATALOG.md](../../docs/06-agent/ERROR-CATALOG.md) | One canonical domain code across all interfaces and owners | RESOLVED |
| R2-07 (F45, F46, D16) | Narrow current support; defer broader machine capability | [docs/07-learning/KNOWLEDGE-MODEL.md](../../docs/07-learning/KNOWLEDGE-MODEL.md) | M7 owns typed global schema/migrations before any promotion API; project-only baseline remains | DEFERRED |
| R2-08 (F27, D07) | Accept concrete contract defect | [docs/06-agent/MCP-SPEC.md](../../docs/06-agent/MCP-SPEC.md) | Deterministic self-contained local schema bundles and recursive refs; integration tests remain pending | RESOLVED |
| R2-09 (CR-01, F28, D01, D09) | Accept concrete contract defect | [docs/02-domain/CONFIRMATION-PROTOCOL.md](../../docs/02-domain/CONFIRMATION-PROTOCOL.md) | Exact core RUNNER binding plus trusted qualification and typed reverse invalidation; missing bytes block current execution, historical hashes unchanged | RESOLVED |
| Existing bounded claims/intent-resolution/scoped acceptance/snapshot/side-effect/assessment mechanisms | Retain already revised owner contract | PRODUCT, PROJECT-LIFECYCLE, MIGRATION-UNITS, ANALYSIS-SPEC, SANDBOX | No competing pipeline, extra successful verdict or duplicate specification; original experiments remain required | ALREADY COVERED |
| Historical-state reconstruction and hostile same-account prevention as baseline promises | Retain rejection with precise boundary | EXECUTION-OBSERVABILITY, SECURITY | Current state plus event window; mediated safety and trusted local account; expanded profiles remain deferred | REJECTED |

| Finding | Severity / resolution gate | Machine owners | Related decision owners | Actual conflict before revision |
| --- | --- | --- | --- | --- |
| R2-01 | P0 before capture acceptance | specs/observation/attestation.schema.json | ADR 0005; related owners referenced by the canonical document | UNAVAILABLE requires supported paths/instrumentation/control/closure evidence; COMPLETE accepts PARTIAL mechanisms structurally |
| R2-02 | P1 before DB extraction | specs/ir/data-model.schema.json | ADR 0005; related owners referenced by the canonical document | Unqualified data model cannot distinguish duplicate table names while owner requires qualified identity |
| R2-03 | P1 before M1 | specs/behavior/semantics.schema.json | ADR 0003; related owners referenced by the canonical document | No row context or write output binding; truth evaluation/termination/rollback/decimal/pointer behavior insufficiently deterministic |
| R2-04 | P1 reference / P2 broader profiles | specs/adapters/manifest.schema.json | ADR 0009; related owners referenced by the canonical document | Open advertised supported_profiles could be treated as compatibility authority |
| R2-05 | P1 before benchmark acceptance | specs/verification/benchmark.schema.json; specs/verification/performance-policy.schema.json | ADR 0015; related owners referenced by the canonical document | Chosen policy/claim/method not explicitly bound via Gate2; estimand/outcome mapping and required measurement completeness ambiguous |
| R2-06 | P1 protocol parity | Interface/domain owner only | ADR 0011; related owners referenced by the canonical document | ERR_CAPABILITY_UNSUPPORTED versus ERR_UNSUPPORTED_CAPABILITY competing canonical names |
| R2-07 | P2 before M7 | Interface/domain owner only | ADR 0014; related owners referenced by the canonical document | Old RWT-023 sequencing and undeveloped global record schema |
| R2-08 | P1 before M5 interfaces | Interface/domain owner only | ADR 0008; related owners referenced by the canonical document | Logical remote schema refs must not imply client network/schema hosting dependency |
| R2-09 | P1 before M0 approval execution | specs/project/execution-manifest.schema.json; specs/adapters/qualification.schema.json; specs/project/validity-graph.schema.json | ADR 0010; related owners referenced by the canonical document | Adapter pins alone do not freeze core evaluator/build/schema bytes; no current typed exact-profile qualification or performance/execution validity edges |

R2-04 resolves baseline admission by exact qualified profiles; only generic executable constraints remain DEFERRED before M8. R2-07 resolves incorrect sequencing language; typed global records/migrations remain DEFERRED before optional M7. Core promotion/acceptance cannot use either deferred capability. Every RESOLVED status denotes specification resolution, not executed behavior.

### D. Remaining architecture blockers

- **P0:** No unresolved canonical design decision identified that prevents the bounded M0 foundation from starting. Hash/snapshot/driver/filesystem/privacy controls must execute before their dependent claims; an existing schema-only PASS is not admission evidence.
- **P1 (before reference/runtime claims):** Actual core/adapter qualification, independent stateful mutant rejection, mechanism/control closure, runtime/image/extension acquisition, reset, target/dependency feasibility and session/rollback continuity remain UNKNOWN until M0–M6 experiments. Signing/security custody remains required before release publication. These are missing evidence, not permission to guess support.
- **P2 (explicit capability prerequisites):** Typed global learning/promotion/retraction machine records and migrations before M7; generic executable profile constraints and wider neutral facets/protocols before M8; general distributed schedules/streams, hostile-agent OS isolation, disjoint parallel mutation and active-history compaction remain outside the initial capability boundary.
- **P3:** Desktop, vector search and convenience configuration rendering remain optional later work.

Missing runtime evidence is recorded as ASSUMPTION/UNKNOWN with its acquisition experiment and blocking claim. The documentation pass does not discharge it.

### E. Revised implementation order

1. M0: coherent snapshot, pinned build/schema/storage closure, durable human subjects and useful assessment projection without implementation authority.
2. M1: bounded two-actor stateful reconstruction, typed row/write binding, independent environments, scoped acceptance, wrong-target mutants, capture incompleteness and interruption without an LLM.
3. M2: characterize licensed CI3/PHP7.4/MySQL/Redis reference safely; preserve dynamic/cron/vendor/native gaps.
4. M3: explicit selected-work continuation; exact target, immutable dependency/performance decisions, lowering ownership model and Gate 2.
5. M4: generated reference slices, session/identity/effect fidelity and implemented regressions.
6. M5: bounded repair and complete CLI/offline-schema stdio interfaces; actual client switch/resume tests.
7. M6: full PROJECT corpus, recomputed approved performance claim, write-aware rollback and exact Gate 3 package.
8. Optional M7: global typed-contract prerequisites, independent holdouts/quality/retraction before promotion.
9. M8: broader profile/neutral facets/adapters; optional UI/scale only after their capability experiments.

[DEVELOPMENT-PLAN](DEVELOPMENT-PLAN.md) retains all eleven required fields for every milestone and owns the queue. Assessment has no second engine, queue or implementation authorization.

### F. Acceptance experiments and Assumption Elimination Pass

[TESTING](../10-development/TESTING.md) owns X01–X60. Existing X01–X55 and all seventeen M1 premise checks remain required. New obligations:

| Undefined/unverified claim | Deterministic requirement / detection / decision | Falsification gate |
| --- | --- | --- |
| ASSUMPTION: a typed rule selects/mutates the correct row and binds its ID | Explicit row context, old-row simultaneous values, real write result, strict types/pointers/terminals and relation assertions | X03–X06/X56 before M1 |
| ASSUMPTION: absent observer means no effects, or producer COMPLETE is true | Current 2.1 unavailable shape plus current trusted mechanism/control derivation | X08/X57 before capture acceptance; useful assessment may still complete |
| ASSUMPTION: stable version/URI means identical execution/support | Exact core/build/schema/adapter bytes, typed qualification and reverse invalidation witnesses | X01/X58 before M0/M1 admission; old vectors remain reconstructable |
| ASSUMPTION: no profile metadata constraints means any profile works | Admit only exact qualified profile; generic constraint language is explicit limitation | X58 baseline; M8 prerequisite for widening |
| ASSUMPTION: benchmark verdict/method is reproducible and approved | Immutable policy in PLAN; paired ratio-of-means/mean-of-trial-quantiles, deterministic bootstrap and complete samples/errors | X33/X59 before accepted performance/Gate 3 |
| ASSUMPTION: a client can resolve schema URLs offline | Self-contained recursive local bundles with preserved constraints, no network fallback | X31/X60 before M5 client claim |
| ASSUMPTION: three successful projects can promote global truth | Typed revisions/consumers/holdouts/mutants/privacy/retraction required before activation | X37–X39 and M7 contracts; explicitly deferred |

New/modified structural examples are illustrative exact strings/refs and rejection cases, not acquired legacy/runtime/control evidence. Qualification status is derived from trusted actual results; an imported agent assertion cannot establish support.

### G. Contradiction scan and verification

Reviewed canonical owners and searched the repository for equivalence/proof/completeness/maturity, approvals/hashes/resolution, corpus/UNKNOWN/DEAD, observation, background/disconnect, history/watermark, rollback/performance, vendor-neutral/SQLite/learning/skill/adapter/transport. Corrected the competing capability-error spelling, blanket registry version, old fixture/test range and obsolete global-task sequencing. Historical review/audit wording is not normative.

Current answers align across AGENTS, PRODUCT, SYSTEM-DESIGN, domain owners, schemas, ADRs and roadmap: bounded claims; current executable branch versus historical reader; qualified core/schema/profile bytes; independent capture and replay; UNIT/regression/PROJECT scopes; immutable decisions versus progress; protected retention; current N/event-window M; shared assessment checkpoint with explicit continuation and separate approval. No unresolved normative contradiction was identified. Intentional M7/M8 deferrals are named prerequisites/refusal boundaries, not competing specifications.

Structural validation: `python3 tools/docs/validate.py` verifies 133 schemas, 222 positive/negative fixtures, 71 CLI/MCP entries, 608 local links, 40 tables across unchanged migrations and ten approval vectors. `git diff --check` checks whitespace. All fifteen ADRs have the nine required sections; all nine milestones have eleven required fields. Original nine approval vectors, the input review and both migrations were independently compared byte/value-wise with the pre-revision repository and preserved.

**No Go/race/framework/runtime/client/security/performance experiment was executed or passed in this task.** Foundation implementation may begin under the narrow current contracts; broader claims remain blocked by their named evidence. Do not begin implementation as part of this revision.

## Review-2 revision deliverables (current checkpoint)

Input: [review-2.md](review-2.md), preserved unchanged. The input contains an initial pass ending partway through section 22 and a second complete pass; their CR/F IDs differ. Prefix A denotes the first pass and B the second. All 42 A failure scenarios and 44 B scenarios are mapped below, as are both critical-risk lists and unnumbered material decisions. Earlier sections of this audit are historical checkpoints, not current machine-version authority.

Authority/impact method: constitution → product → architecture/accepted ADRs → canonical topic owner → machine contracts → storage/comparison → falsification registry. The owner column is the canonical decision authority; changed-schema/ADR links are dependent contracts. Each decision was checked against current owner semantics, not adopted solely from the review. No finding requires reversing the shared assessment architecture. RESOLVED means specified and represented; IMPLEMENTATION VALIDATION means existing or revised safeguards still need execution; DEFERRED means a bounded unsupported capability. No material finding was silently rejected.

### C. Review-2 disposition matrix

| Finding / priority | Status | Decision | Owner | Changed contracts / related ADR | Test/fixture | Remaining limitation |
| --- | --- | --- | --- | --- | --- | --- |
| A.CR-01 B.CR-01 A.F01 B.F01 (P0) | RESOLVED | Separate fixture, shadow, app, observer, evaluator, comparator and generator handles; actual-only checking | [BEHAVIOR-MODEL.md](../02-domain/BEHAVIOR-MODEL.md) | [BEHAVIOR-MODEL.md](../02-domain/BEHAVIOR-MODEL.md)<br>[behavior/evaluation-context.schema.json](../../specs/behavior/evaluation-context.schema.json)<br>[observation/execution-witness.schema.json](../../specs/observation/execution-witness.schema.json)<br>[ADR 0003](../01-architecture/ADR/0003-behavior-as-canonical-model.md) | X61/X63 | Write-denial controls still must execute |
| A.CR-02 B.CR-02 A.F02 B.F02 (P0) | RESOLVED | Side-local ledger before effects/trace; pairwise map derived later without rewriting capture | [RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md) | [RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md)<br>[observation/local-identities.schema.json](../../specs/observation/local-identities.schema.json)<br>[observation/effects.schema.json](../../specs/observation/effects.schema.json)<br>[verification/identity-map.schema.json](../../specs/verification/identity-map.schema.json)<br>[ADR 0005](../01-architecture/ADR/0005-differential-verification.md) | X62 | Trusted capture/publication remains unimplemented |
| A.CR-03 B.CR-03 A.F03 B.F04 (P1) | RESOLVED | Versioned closed projection, tagged exact values, phase/origin, missing-value rules | [EVALUATION-CONTEXT.md](../02-domain/EVALUATION-CONTEXT.md) | [EVALUATION-CONTEXT.md](../02-domain/EVALUATION-CONTEXT.md)<br>[behavior/evaluation-context.schema.json](../../specs/behavior/evaluation-context.schema.json)<br>[behavior/initial-state.schema.json](../../specs/behavior/initial-state.schema.json)<br>[ADR 0003](../01-architecture/ADR/0003-behavior-as-canonical-model.md) | X63/X56 | Bounded typed projection only |
| A.CR-04 B.CR-07 A.F04 B.F05 (P1) | RESOLVED | Typed headers/cookies/session actions and exact qualified provider bindings | [TARGET-ARCHITECTURE.md](../04-migration/TARGET-ARCHITECTURE.md) | [TARGET-ARCHITECTURE.md](../04-migration/TARGET-ARCHITECTURE.md)<br>[behavior/response.schema.json](../../specs/behavior/response.schema.json)<br>[behavior/semantics.schema.json](../../specs/behavior/semantics.schema.json)<br>[migration/target-model.schema.json](../../specs/migration/target-model.schema.json)<br>[ADR 0009](../01-architecture/ADR/0009-three-tier-ir-architecture.md) | X67 | Only qualified narrow HTTP/session providers |
| A.CR-05 B.CR-05 A.F05 B.F06 (P1) | RESOLVED | Refuse live/unknown embedded credentials; separately approved synthetic/revoked variant or static-only assessment | [SECURITY.md](../09-security/SECURITY.md) | [SECURITY.md](../09-security/SECURITY.md)<br>[project/execution-admission.schema.json](../../specs/project/execution-admission.schema.json)<br>[project/execution-variant.schema.json](../../specs/project/execution-variant.schema.json)<br>[ADR 0010](../01-architecture/ADR/0010-mandatory-user-confirmation-gates.md) | X69 | Variant evidence does not prove original behavior |
| A.CR-06 B.CR-04 A.F06 B.F07 (P1) | RESOLVED | Separate content and filesystem semantic profile, coherent capture and relevant drift invalidation | [ANALYSIS-SPEC.md](../03-analysis/ANALYSIS-SPEC.md) | [ANALYSIS-SPEC.md](../03-analysis/ANALYSIS-SPEC.md)<br>[project/filesystem-profile.schema.json](../../specs/project/filesystem-profile.schema.json)<br>[project/source-snapshot.schema.json](../../specs/project/source-snapshot.schema.json)<br>[ADR 0010](../01-architecture/ADR/0010-mandatory-user-confirmation-gates.md) | X68/X02 | Unsupported relevant metadata blocks runtime claim |
| B.CR-06 A.F07 B.F03 (P1) | RESOLVED | Logical occurrences and STOP_AT_FIRST_DIVERGENCE; preserve known mismatch and unmatched evidence | [DIFFERENTIAL-TESTING.md](../05-verification/DIFFERENTIAL-TESTING.md) | [DIFFERENTIAL-TESTING.md](../05-verification/DIFFERENTIAL-TESTING.md)<br>[observation/step-occurrence.schema.json](../../specs/observation/step-occurrence.schema.json)<br>[verification/occurrence-pairing.schema.json](../../specs/verification/occurrence-pairing.schema.json)<br>[ADR 0005](../01-architecture/ADR/0005-differential-verification.md) | X64/X65 | Independent diagnostic continuation deferred |
| A.F08 B.F08 (P1) | RESOLVED | Witness actual loaded artifact, process, environment, endpoint and role controls | [RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md) | [RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md)<br>[observation/execution-witness.schema.json](../../specs/observation/execution-witness.schema.json)<br>[schemas/trace-schema.json](../../specs/schemas/trace-schema.json)<br>[ADR 0005](../01-architecture/ADR/0005-differential-verification.md) | X66 | A claimed tree/process version alone is insufficient |
| A.F18 B.F09 (P1) | RESOLVED | Typed intended principal separate from captured auth/session/DB principal | [AUTH-AND-SESSION.md](../04-migration/AUTH-AND-SESSION.md) | [AUTH-AND-SESSION.md](../04-migration/AUTH-AND-SESSION.md)<br>[behavior/principal-fixture.schema.json](../../specs/behavior/principal-fixture.schema.json)<br>[behavior/evaluation-context.schema.json](../../specs/behavior/evaluation-context.schema.json)<br>[ADR 0003](../01-architecture/ADR/0003-behavior-as-canonical-model.md) | X77/X06 | Actual authentication channel must be qualified |
| A.F27 A.F28 B.F36 (P1) | RESOLVED | Positive prerequisite coverage; missing edges/unexamined mechanisms invalidate; no old PASS reuse | [DOMAIN-MODEL.md](../02-domain/DOMAIN-MODEL.md) | [DOMAIN-MODEL.md](../02-domain/DOMAIN-MODEL.md)<br>[project/consumption-coverage.schema.json](../../specs/project/consumption-coverage.schema.json)<br>[project/validity-graph.schema.json](../../specs/project/validity-graph.schema.json)<br>[project/reuse-certificate.schema.json](../../specs/project/reuse-certificate.schema.json)<br>[ADR 0012](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) | X73/X29 | Coverage quality still needs runtime controls |
| A.F37 B.F10 (P1) | RESOLVED | Reverify currently implemented consumers only; future consumers retain pending obligations | [MIGRATION-UNITS.md](../04-migration/MIGRATION-UNITS.md) | [MIGRATION-UNITS.md](../04-migration/MIGRATION-UNITS.md)<br>[ADR 0012](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) | X35/X22 | Unknown consumer coverage conservatively invalidates |
| A.F15 A.F22 B.F20 B.F23 (P1) | RESOLVED | Typed business application versus attempt/retry/receipt; reconcile accepted-before-ack | [COMPATIBILITY-SPEC.md](../05-verification/COMPATIBILITY-SPEC.md) | [COMPATIBILITY-SPEC.md](../05-verification/COMPATIBILITY-SPEC.md)<br>[observation/effect-payload.schema.json](../../specs/observation/effect-payload.schema.json)<br>[observation/effect-occurrence.schema.json](../../specs/observation/effect-occurrence.schema.json)<br>[ADR 0005](../01-architecture/ADR/0005-differential-verification.md) | X72/X07 | Unavailable application/deduplication evidence is UNKNOWN |
| A.F09 A.F10 A.F24 A.F25 A.F26 B.F14 B.F15 B.F16 B.F17 B.F18 (P1) | RESOLVED | Extend existing reset/capture boundary explicitly to service processes, routines and qualified connections | [DATABASE-COMPARISON.md](../05-verification/DATABASE-COMPARISON.md) | [DATABASE-COMPARISON.md](../05-verification/DATABASE-COMPARISON.md)<br>[observation/effect-payload.schema.json](../../specs/observation/effect-payload.schema.json)<br>[ADR 0005](../01-architecture/ADR/0005-differential-verification.md) | X70/X11/X14 | Real DB/worker/reset qualification not demonstrated |
| A.F21 A.F23 B.F19 B.F21 (P2) | RESOLVED | Bound finite windows/horizons and inspect pending queues; long-delay obligations remain unsupported | [RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md) | [RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md)<br>[schemas/trace-schema.json](../../specs/schemas/trace-schema.json) | X71/X12 | General long-delay scheduling deferred to M8 |
| Classification contradiction, both passes (P1) | RESOLVED | Exclusive precedence: observed ACTIVE; unresolved UNKNOWN; resolved reachable DORMANT; complete omission bundle DEAD | [CODE-CLASSIFICATION.md](../03-analysis/CODE-CLASSIFICATION.md) | [CODE-CLASSIFICATION.md](../03-analysis/CODE-CLASSIFICATION.md) | X76/X09 | Never lowers the omission threshold |
| B.F26 (P2) | RESOLVED | Opaque provider-qualified release labels plus immutable digest and separate toolchain components | [TARGET-SPECIFICATION.md](../04-migration/TARGET-SPECIFICATION.md) | [TARGET-SPECIFICATION.md](../04-migration/TARGET-SPECIFICATION.md)<br>[project/resolved-environment.schema.json](../../specs/project/resolved-environment.schema.json)<br>[ADR 0009](../01-architecture/ADR/0009-three-tier-ir-architecture.md) | X75/X15 | Provider constraint/availability interpretation must execute |
| Detection gate-state contradiction (P1) | RESOLVED | Current detection 2.0 removes gate_status; old value is historical only | [TARGET-SPECIFICATION.md](../04-migration/TARGET-SPECIFICATION.md) | [TARGET-SPECIFICATION.md](../04-migration/TARGET-SPECIFICATION.md)<br>[schemas/detection-schema.json](../../specs/schemas/detection-schema.json)<br>[ADR 0012](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) | Current detection structural negatives | Projection readiness remains separately enforced |
| CLI version and scenario-import contradictions (P1) | RESOLVED | Catalog 2.1 header and scenario 2.0 reference closure; CLI/MCP share guarded records | [CLI-SPEC.md](../06-agent/CLI-SPEC.md) | [CLI-SPEC.md](../06-agent/CLI-SPEC.md) | X31/X60 | Actual named clients remain UNVERIFIED |
| Source-vs-source comparison, section 10 (P1) | RESOLVED | Distinct witnessed SOURCE/TARGET prerequisites; nullable refs only for genuine UNKNOWN | [DIFFERENTIAL-TESTING.md](../05-verification/DIFFERENTIAL-TESTING.md) | [DIFFERENTIAL-TESTING.md](../05-verification/DIFFERENTIAL-TESTING.md)<br>[verification/report.schema.json](../../specs/verification/report.schema.json)<br>[observation/execution-witness.schema.json](../../specs/observation/execution-witness.schema.json)<br>[ADR 0005](../01-architecture/ADR/0005-differential-verification.md) | X80 | Refusal is infrastructure, not a semantic kill |
| A.F36 B.F37 (P1) | RESOLVED | Reserve bounded bytes and emergency slots; durable receipts before ACK; explicit uncommitted diagnostics/reconciliation | [RECOVERY.md](../08-storage/RECOVERY.md) | [RECOVERY.md](../08-storage/RECOVERY.md)<br>[operations/recovery-journal.schema.json](../../specs/operations/recovery-journal.schema.json)<br>[ADR 0013](../01-architecture/ADR/0013-transparent-execution.md) | X74/X27 | Device loss can still prevent last-detail durability |
| A.F38 B.F39 B.F40 (P2) | RESOLVED | Predeclare fixed N; retain all attempts; separate every-window operational SLO from mean quantile statistics | [PERFORMANCE.md](../05-verification/PERFORMANCE.md) | [PERFORMANCE.md](../05-verification/PERFORMANCE.md)<br>[verification/performance-policy.schema.json](../../specs/verification/performance-policy.schema.json)<br>[verification/operational-slo.schema.json](../../specs/verification/operational-slo.schema.json)<br>[verification/benchmark.schema.json](../../specs/verification/benchmark.schema.json)<br>[ADR 0015](../01-architecture/ADR/0015-statistical-performance.md) | X78/X33/X59 | M6 must qualify method/workload; sequential methods deferred |
| A.F39 A.F40 B.F41 B.F42 B.F43 (P2) | DEFERRED | Retain advisory pins; require M7 consumer index, epochs, independent holdouts and offline reopen reconciliation | [LEARNING-SYSTEM.md](../07-learning/LEARNING-SYSTEM.md) | [LEARNING-SYSTEM.md](../07-learning/LEARNING-SYSTEM.md) | X79/X37/X38/X39 | No global activation before typed M7 contracts and experiments |
| A.F41; parser/capture/vulnerability gaps in section 16 (P1) | RESOLVED | Inherited secret sensitivity, bounded parser worker, trusted channels, explicit vulnerability disposition | [SECURITY.md](../09-security/SECURITY.md) | [SECURITY.md](../09-security/SECURITY.md) | X30/X32/X69 plus parser limit controls | Same-account hostile-agent isolation remains deferred |
| A.F11 A.F12 A.F13 A.F14 A.F34 B.F11 B.F12 B.F13 B.F34 (P1) | IMPLEMENTATION VALIDATION | Preserve dynamic/initialization uncertainty, modified bytes and closed operational roots; never infer DEAD from no hits | [ANALYSIS-SPEC.md](../03-analysis/ANALYSIS-SPEC.md) | Existing owning contract retained; see manifest for dependent edits | X09/X10/X76 | Actual ugly-source parser/adapter controls remain M2 |
| A.F16 B.F25 (P1) | IMPLEMENTATION VALIDATION | Existing acquisition/license/native/platform blockers stay explicit; useful assessment can complete | [DEPENDENCY-RESOLUTION.md](../04-migration/DEPENDENCY-RESOLUTION.md) | Existing owning contract retained; see manifest for dependent edits | X13/X18/X42 | Actual licensed artifacts/runtime availability unknown |
| A.F17 A.F32 A.F33 B.F27 B.F28 B.F30 (P1) | IMPLEMENTATION VALIDATION | Preserve exact inputs, bijection/relations and business-time normalizer controls; ambiguity stays UNKNOWN | [COMPATIBILITY-SPEC.md](../05-verification/COMPATIBILITY-SPEC.md) | Existing owning contract retained; see manifest for dependent edits<br>[verification/identity-map.schema.json](../../specs/verification/identity-map.schema.json)<br>[behavior/evaluation-context.schema.json](../../specs/behavior/evaluation-context.schema.json) | X20/X21/X63 | Initial paired scalar identity subset; no inferred arbitrary mapping |
| A.F19 A.F20 B.F31 (P1) | IMPLEMENTATION VALIDATION | Lifecycle controls supplement typed session/response obligations; old tokens and cross-account jars are tested | [AUTH-AND-SESSION.md](../04-migration/AUTH-AND-SESSION.md) | Existing owning contract retained; see manifest for dependent edits<br>[behavior/response.schema.json](../../specs/behavior/response.schema.json) | X19/X67/X77 | Real legacy session bridge remains unqualified |
| A.F29 A.F30 B.F32 B.F33 (P1) | IMPLEMENTATION VALIDATION | Retain atomic claims and OS ownership across publication; no SQL-only fencing claim | [RECOVERY.md](../08-storage/RECOVERY.md) | Existing owning contract retained; see manifest for dependent edits | X22/X23/X25 | No multi-process runtime tests yet |
| A.F31 B.F29 (P2) | DEFERRED | Reject unsupported non-UTF8 HTTP header encoding; preserve raw bytes privately, never lossy replay | [EVALUATION-CONTEXT.md](../02-domain/EVALUATION-CONTEXT.md) | [EVALUATION-CONTEXT.md](../02-domain/EVALUATION-CONTEXT.md) | X21/X60 | Wider invocation/encoding support must be qualified separately |
| A.F35 (P1) | IMPLEMENTATION VALIDATION | Keep short reads/checkpoint controls and actual pinned SQLite build/restore qualification | [SQLITE-SCHEMA.md](../08-storage/SQLITE-SCHEMA.md) | Existing owning contract retained; see manifest for dependent edits | X28 | Scale/WAL behavior not established by Python validation |
| A.F42 B.F44 (P1) | IMPLEMENTATION VALIDATION | Retain target-write/session reconciliation before release; unchanged schema is insufficient | [CUTOVER-RUNBOOK.md](../10-development/CUTOVER-RUNBOOK.md) | Existing owning contract retained; see manifest for dependent edits | X34 | No production runner or proven rollback yet |
| B.F35 (P1) | IMPLEMENTATION VALIDATION | Retain coherent snapshot acquisition and ABA/symlink controls; add filesystem profile identity | [ANALYSIS-SPEC.md](../03-analysis/ANALYSIS-SPEC.md) | Existing owning contract retained; see manifest for dependent edits<br>[project/source-snapshot.schema.json](../../specs/project/source-snapshot.schema.json) | X02/X68 | Actual acquisition/platform controls pending |
| B.F38; visibility/retry findings in section 13 (P1) | RESOLVED | Preserve state N/window M; expose retry/no-progress and uncommitted emergency diagnostics | [EXECUTION-OBSERVABILITY.md](../06-agent/EXECUTION-OBSERVABILITY.md) | [EXECUTION-OBSERVABILITY.md](../06-agent/EXECUTION-OBSERVABILITY.md)<br>[ADR 0013](../01-architecture/ADR/0013-transparent-execution.md) | X27/X74 | No historical mutable-state projection promised |
| B.F22 (P1) | IMPLEMENTATION VALIDATION | Retain closed egress and no live irreversible replay; extend matrix to all service processes | [SANDBOX.md](../09-security/SANDBOX.md) | Existing owning contract retained; see manifest for dependent edits | X44/X69/X70 | No faithful simulator means UNKNOWN/unsupported |
| B.F24 (P1) | ALREADY RESOLVED | User target remains sovereign; infeasible requested versions block instead of silent substitution | [TARGET-SPECIFICATION.md](../04-migration/TARGET-SPECIFICATION.md) | Existing owning contract retained; see manifest for dependent edits | X16/X49 | Actual engine enforcement pending |
| Shared assessment / scoped acceptance / authority strengths (P1) | ALREADY RESOLVED | Preserve one checkpoint and separate explicit selection, plan approval and acceptance scopes | [PROJECT-LIFECYCLE.md](../02-domain/PROJECT-LIFECYCLE.md) | Existing owning contract retained; see manifest for dependent edits | X35/X41–X55 | Report completion never establishes feasibility or approval |
| Distribution/vendor/scale unknowns, sections 11/17/18 (P2) | IMPLEMENTATION VALIDATION | Retain narrow platform, licensing/acquisition checks and UNVERIFIED clients; no universal neutrality/zero-setup claim | [RELEASE.md](../10-development/RELEASE.md) | Existing owning contract retained; see manifest for dependent edits | X28/X31/X36/X40 | Runtime/client/distribution/scale evidence still absent |
| First milestone and assumption elimination, sections 19/25/26 (P1) | RESOLVED | M0 foundation then narrow M1 with four decisive controls; retain all seventeen checks and deferred boundaries | [DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md) | [DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md) | X61–X80 plus existing registry | Specified contracts are not demonstrated capabilities |

### D. Remaining blockers and limits

- **P0:** No unresolved specification decision from the two P0 findings. Actual checker isolation and source-only publication are mandatory M1 qualification, not demonstrated properties.
- **P1:** Runtime validation remains outstanding: exact Go/SQLite/filesystem build, role isolation, projection/authentication/branch/response semantics, witness/containment/capture/reset, positive coverage, safe credential admission and emergency persistence. Missing customer/license/native/session/rollback evidence blocks those scopes/claims. These do not block bounded M0 implementation.
- **P2:** Explicitly deferred non-HTTP/wider language facets, general concurrency/long-delay schedules, wider encodings/platforms, active-history compaction, M7 global consumer/retraction machine lifecycle and sequential statistical methods. M6 fixed-N/SLO qualification is still required before performance claims.
- **P3:** Desktop UI, embeddings/vector search, hosted integrations and richer advisory renderers remain later. Production execution is outside initial release scope.

### E. Revised dependency/order graph

Canonical record/role contracts → M0 persistence/snapshot/approval/assessment/state → M1 narrow wrong-target falsification → M2 real source qualification → M3 exact decisions/model/coverage → M4 reference slices → M5 actual CLI/client recovery → M6 complete approved corpus/performance/release package → optional M7 learning and M8 expanded capability qualification. DEVELOPMENT-PLAN owns prerequisites and non-goals; there is no second assessment engine or parallel implementation queue.

### F. M0/M1 acceptance and experiments

M0 proves durable records, coherent bounded snapshot/profile identity, exact approval reconstruction, source-only record publication, assessment without authorization, current state and reserved recovery/backup mechanics. It does not prove a verifier, adapter, authentication bridge or customer runtime.

M1 retains the seventeen premise checks and the small two-actor login/create/denial/DB/acknowledged-sink fixture. It additionally requires checker non-mutation, source-only evidence before a target, stale/wrong build refusal and semantic branch-divergence detection. X61–X80 extend the existing TESTING registry with setup, mutant, expected result/reason, required evidence, failure class and milestone. A good control must execute successfully first; unrelated startup failure is never a semantic mutant kill. No arbitrary PHP/browser/distributed/payment/production expansion.

### G. Contradiction report

Current owners agree on: shadow-only model effects; actual-only evaluator inputs; local identity before pairing; target model 2.1 qualified lowerings; source snapshot 2.1 filesystem/admission closure; current report 2.1 nullable missing evidence only for UNKNOWN; current detection without gate state; opaque provider-qualified versions; currently implemented consumer regressions; CLI/MCP catalog 2.1; fixed-N policy 1.1/benchmark 2.2 and separate SLO; emergency diagnostics distinct from committed state. Current branch versions are listed in specs/README. Historical schema branches and earlier review/audit checkpoints intentionally preserve old wording/bytes and cannot authorize current execution. Global lifecycle/wider capabilities remain explicitly deferred, not conflicting current guarantees.

### H. Claim audit

All runtime/adapter/client/security/performance/scale claims remain SPECIFIED or UNVERIFIED. Structural schema/fixture/hash/DDL validation does not prove enforcement. Content identity does not establish filesystem behavior; approved credential variants do not prove original-source behavior; finite quiescence does not establish unbounded absence; repeated observations do not prove harmless normalization; absent dependency edges do not prove independence; actor labels do not prove authentication; a digest reported by an app does not prove loaded execution; mean trial p99 does not certify an every-window or production SLO; restored schema alone does not prove write-aware rollback. Device loss can defeat emergency persistence. Same-account hostile access is outside the mediated boundary. Each remaining assumption is assigned to TESTING/DEVELOPMENT-PLAN or an explicit unsupported boundary.

### A. Changed files

| Path | Purpose | Major change |
| --- | --- | --- |
| [AGENTS.md](../../AGENTS.md) | Constitution | Makes measured-state non-mutation, actual-only evidence and source-before-pair identity constitutional. |
| [REWROTRON_PROJECT_PLAN.md](../../REWROTRON_PROJECT_PLAN.md) | Implementation overview | Names resolved verifier contract prerequisites and keeps M0/M1 narrow with deferred later scope. |
| [docs/00-vision/GLOSSARY.md](../00-vision/GLOSSARY.md) | Domain vocabulary | Indexes fixture/model/actual state, identity, occurrence, witness, coverage and admission owners. |
| [docs/00-vision/PRODUCT.md](../00-vision/PRODUCT.md) | Bounded product claims | Requires actual execution identity and explicit limits for variants, projection and long-delay behavior. |
| [docs/01-architecture/ADR/0003-behavior-as-canonical-model.md](../01-architecture/ADR/0003-behavior-as-canonical-model.md) | Architectural decision | Executable behavior with isolated model execution; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/ADR/0005-differential-verification.md](../01-architecture/ADR/0005-differential-verification.md) | Architectural decision | Independent capture and downstream correspondence; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/ADR/0009-three-tier-ir-architecture.md](../01-architecture/ADR/0009-three-tier-ir-architecture.md) | Architectural decision | Neutral IRs and qualified target lowering; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/ADR/0010-mandatory-user-confirmation-gates.md](../01-architecture/ADR/0010-mandatory-user-confirmation-gates.md) | Architectural decision | Durable exact-subject gates and execution boundary; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/ADR/0012-target-specification-and-dependency-resolution.md](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md) | Architectural decision | Immutable intent and positive reuse coverage; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/ADR/0013-transparent-execution.md](../01-architecture/ADR/0013-transparent-execution.md) | Architectural decision | Owned execution and emergency persistence; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/ADR/0015-statistical-performance.md](../01-architecture/ADR/0015-statistical-performance.md) | Architectural decision | Fixed-N statistical claims and separate SLO windows; records tradeoffs, machine impact, compatibility and falsification. |
| [docs/01-architecture/COMPONENTS.md](../01-architecture/COMPONENTS.md) | Package responsibilities | Maps role-specific ports into existing neutral services without a package or pipeline per role. |
| [docs/01-architecture/DATA-FLOW.md](../01-architecture/DATA-FLOW.md) | Publication order | Documents acyclic capture/context/occurrence/trace/pair/report dependencies. |
| [docs/01-architecture/SERVICE-CONTRACTS.md](../01-architecture/SERVICE-CONTRACTS.md) | Guarded ports/publication | Separates role handles and records current schema, provenance and qualification prerequisites. |
| [docs/01-architecture/SYSTEM-DESIGN.md](../01-architecture/SYSTEM-DESIGN.md) | Architecture boundaries | Adds explicit simulator/runner/observer/evaluator/comparator roles and one-way evidence flow. |
| [docs/02-domain/BEHAVIOR-MODEL.md](../02-domain/BEHAVIOR-MODEL.md) | Executable behavior owner | Defines effect isolation, response/session lowering, supported provider and finite-window limits. |
| [docs/02-domain/CONFIRMATION-PROTOCOL.md](../02-domain/CONFIRMATION-PROTOCOL.md) | Approval identity | Binds filesystem/admission/variant and current schema closure without changing historical hash preimages. |
| [docs/02-domain/DOMAIN-MODEL.md](../02-domain/DOMAIN-MODEL.md) | Validity and reuse | Requires positive consumption coverage and conservative invalidation when discovery is incomplete. |
| [docs/02-domain/IR-CONTRACTS.md](../02-domain/IR-CONTRACTS.md) | Neutral provenance | Distinguishes local identities from paired evidence and preserves initialization uncertainty. |
| [docs/02-domain/PROJECT-LIFECYCLE.md](../02-domain/PROJECT-LIFECYCLE.md) | Readiness reference | Updates current benchmark version; preserves shared assessment completion and separate approval. |
| [docs/02-domain/TERMINOLOGY.md](../02-domain/TERMINOLOGY.md) | Term consistency | Separates actor label/principal, local identity/correspondence and model/observed state. |
| [docs/03-analysis/ANALYSIS-SPEC.md](../03-analysis/ANALYSIS-SPEC.md) | Source boundary | Adds coherent filesystem semantic profiles and distinct admission/variant identity. |
| [docs/03-analysis/CODE-CLASSIFICATION.md](../03-analysis/CODE-CLASSIFICATION.md) | Exclusive classification | Defines ACTIVE/UNKNOWN/DORMANT/DEAD precedence without weakening omission evidence. |
| [docs/03-analysis/RUNTIME-OBSERVATION.md](../03-analysis/RUNTIME-OBSERVATION.md) | Actual capture owner | Defines source-only identity, execution witnesses, service boundaries and bounded observation windows. |
| [docs/03-analysis/WORKFLOW-DISCOVERY.md](../03-analysis/WORKFLOW-DISCOVERY.md) | Discovery integration | Uses side-local outcomes and logical occurrence/first-divergence semantics. |
| [docs/04-migration/AUTH-AND-SESSION.md](../04-migration/AUTH-AND-SESSION.md) | Authentication obligations | Separates fixture intention from captured session/DB principal and typed lifecycle lowering. |
| [docs/04-migration/MIGRATION-ENGINE.md](../04-migration/MIGRATION-ENGINE.md) | Execution orchestration | Limits shared regression to currently implemented consumers; future obligations stay pending. |
| [docs/04-migration/MIGRATION-UNITS.md](../04-migration/MIGRATION-UNITS.md) | Schedulable acceptance | Makes shared-consumer obligations current-scope-only and coverage-backed. |
| [docs/04-migration/TARGET-ARCHITECTURE.md](../04-migration/TARGET-ARCHITECTURE.md) | Implementation model | Adds qualified provider bindings and implemented-consumer scope. |
| [docs/04-migration/TARGET-SPECIFICATION.md](../04-migration/TARGET-SPECIFICATION.md) | Exact environment and facts | Uses provider-qualified opaque releases/toolchains and separates detection from gate projection. |
| [docs/05-verification/COMPATIBILITY-SPEC.md](../05-verification/COMPATIBILITY-SPEC.md) | Comparison semantics | Adds typed effects, delivery/application cardinality and actual principal/identity requirements. |
| [docs/05-verification/DATABASE-COMPARISON.md](../05-verification/DATABASE-COMPARISON.md) | Capture/reset closure | Includes DB servers/routines/triggers and all owned service-process/reset state. |
| [docs/05-verification/DIFFERENTIAL-TESTING.md](../05-verification/DIFFERENTIAL-TESTING.md) | Read-only verdict protocol | Defines measured boundary, logical pairing, divergence and nullable UNKNOWN without fake evidence. |
| [docs/05-verification/PERFORMANCE.md](../05-verification/PERFORMANCE.md) | Statistical acceptance | Fixes N before sampling and separates statistical estimand from observed per-window SLO. |
| [docs/06-agent/AGENT-MODEL.md](../06-agent/AGENT-MODEL.md) | Evidence authority | Rejects agent/application assertions as trusted capture and forbids silent vulnerability choices. |
| [docs/06-agent/CLI-SPEC.md](../06-agent/CLI-SPEC.md) | CLI parity | Corrects catalog 2.1 and exposes current guarded records/emergency diagnostics. |
| [docs/06-agent/ERROR-CATALOG.md](../06-agent/ERROR-CATALOG.md) | Machine failure reasons | Defines stale execution, invalid comparison, type and persistence errors with safe retry rules. |
| [docs/06-agent/EXECUTION-OBSERVABILITY.md](../06-agent/EXECUTION-OBSERVABILITY.md) | Truthful status | Distinguishes activity, retry/no-progress and emergency/uncommitted facts from durable N. |
| [docs/06-agent/MCP-SPEC.md](../06-agent/MCP-SPEC.md) | MCP parity | Consumes the same guarded records and current imports without duplicating domain semantics. |
| [docs/07-learning/LEARNING-SYSTEM.md](../07-learning/LEARNING-SYSTEM.md) | Deferred retraction | Specifies M7 consumer epochs/reopen checks without pulling global lifecycle into foundation. |
| [docs/08-storage/ARTIFACT-STORE.md](../08-storage/ARTIFACT-STORE.md) | Capacity admission | Reserves bounded artifact/receipt capacity and fixed emergency slots before effects. |
| [docs/08-storage/RECOVERY.md](../08-storage/RECOVERY.md) | Crash/effect reconciliation | Defines emergency journal durability, status, receipts, cleanup and no blind retry. |
| [docs/08-storage/SQLITE-SCHEMA.md](../08-storage/SQLITE-SCHEMA.md) | Storage mapping | Maps new artifacts to existing tables and distinguishes structural DDL from runtime qualification. |
| [docs/09-security/SANDBOX.md](../09-security/SANDBOX.md) | Whole-service containment | Applies admission/effect restrictions to DB/workers/helpers/simulator and denies measured checker writes. |
| [docs/09-security/SECRET-HANDLING.md](../09-security/SECRET-HANDLING.md) | Source versus safe variant | Preserves private original while allowing only separately identified reviewed execution variants. |
| [docs/09-security/SECURITY.md](../09-security/SECURITY.md) | Execution/privacy policy | Defines credential refusal/variant admission, parser limits, trusted capture and explicit vulnerability decisions. |
| [docs/10-development/FIXTURE-CORPUS.md](../10-development/FIXTURE-CORPUS.md) | Fixture registry integration | References X61–X80 and keeps structural examples distinct from runtime experiments. |
| [docs/10-development/PROJECT-INPUTS.md](../10-development/PROJECT-INPUTS.md) | Current import shape | Replaces stale flat scenario fields with current workflow/profile/input-ref closure. |
| [docs/10-development/TESTING.md](../10-development/TESTING.md) | Falsification registry | Adds X61–X80 with setup, mutant, result/reason, evidence, failure class and milestone. |
| [docs/11-roadmap/BACKLOG.md](BACKLOG.md) | Deferred sequencing | Assigns fixed-N/SLO and global retraction to later milestones. |
| [docs/11-roadmap/DECISION-REGISTER.md](DECISION-REGISTER.md) | Decision/validation index | Records specified boundary decisions and outstanding runtime evidence. |
| [docs/11-roadmap/DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md) | Sole implementation queue | Adds exact M0/M1 prerequisites and four decisive controls; assigns later tests to M2–M8. |
| [docs/11-roadmap/DOCUMENTATION-AUDIT.md](DOCUMENTATION-AUDIT.md) | Audit index | Points to current review-2 dispositions and separates historical evidence from current owners. |
| [docs/11-roadmap/MILESTONES.md](MILESTONES.md) | Milestone index | Points to revised canonical prerequisites and avoids a parallel queue. |
| [docs/11-roadmap/RED-TEAM-RESOLUTION.md](RED-TEAM-RESOLUTION.md) | Revision audit deliverable | Maps both critical-risk lists/all 86 failures; records changed/new files, limits, ordering and claim audit. |
| [docs/11-roadmap/ROADMAP.md](ROADMAP.md) | Sequencing overview | Preserves foundation then falsification and explicit later/deferred boundaries. |
| [docs/README.md](../README.md) | Canonical registry/index | Updates current owner/version map and provenance/admission rules. |
| [specs/README.md](../../specs/README.md) | Canonical registry/index | Updates current owner/version map and provenance/admission rules. |
| [specs/behavior/semantics.schema.json](../../specs/behavior/semantics.schema.json) | Bounded actions | Adds typed response/header/cookie/session actions under pinned model-role capability. |
| [specs/examples/validation-cases.json](../../specs/examples/validation-cases.json) | Structural fixture registry | Adds 79 synthetic positive/negative cases; old vectors/examples remain historical, no runtime results claimed. |
| [specs/migration/target-model.schema.json](../../specs/migration/target-model.schema.json) | Qualified lowering | 2.1 binds provider manifest/qualification and input/output shapes to obligations. |
| [specs/observation/effects.schema.json](../../specs/observation/effects.schema.json) | Actual effect publication | 2.1 removes paired-map dependency and requires local identity, execution witness and provenance. |
| [specs/project/resolved-environment.schema.json](../../specs/project/resolved-environment.schema.json) | Exact components | 2.1 provider-qualified opaque releases, immutable distributions and toolchain components. |
| [specs/project/reuse-certificate.schema.json](../../specs/project/reuse-certificate.schema.json) | Reuse authorization | 2.1 requires nonempty positive coverage references. |
| [specs/project/source-snapshot.schema.json](../../specs/project/source-snapshot.schema.json) | Approved source closure | 2.1 adds filesystem/admission refs; historical content/approval bytes retained. |
| [specs/project/validity-graph.schema.json](../../specs/project/validity-graph.schema.json) | Dependency closure | 2.1 records consumption-coverage refs rather than trusting absent edges. |
| [specs/schemas/detection-schema.json](../../specs/schemas/detection-schema.json) | Immutable facts | 2.0 removes current gate state; 1.0 remains historical-readable. |
| [specs/schemas/trace-schema.json](../../specs/schemas/trace-schema.json) | Current actual traces | 2.1 adds context/local identity/occurrence/witness and bounded window/horizon. |
| [specs/verification/benchmark.schema.json](../../specs/verification/benchmark.schema.json) | Current performance result | 2.2 separates operational window outcome from statistical acceptance. |
| [specs/verification/identity-map.schema.json](../../specs/verification/identity-map.schema.json) | Derived correspondence | 2.1 binds both local ledgers/IDs and occurrence pairing downstream of capture. |
| [specs/verification/performance-policy.schema.json](../../specs/verification/performance-policy.schema.json) | Approved sampling | 1.1 fixes N/stopping rule and explicit optional operational SLO. |
| [specs/verification/report.schema.json](../../specs/verification/report.schema.json) | Scoped results | 2.1 requires pairing for PASS and permits explicit missing evidence for infrastructure UNKNOWN. |

### B. New files

Only one new canonical document (evaluation context) and sixteen schemas representing required records. No new ADR, tool, Skill, Go source or SQL migration. The supplied untracked review-2.md existed at the task baseline and is not a created deliverable.

| Path | Purpose | Major change |
| --- | --- | --- |
| [docs/02-domain/EVALUATION-CONTEXT.md](../02-domain/EVALUATION-CONTEXT.md) | Canonical projection ABI | New single owner for namespace shapes, phase/provenance, exact values, missing/null and principal projection. |
| [specs/behavior/evaluation-context.schema.json](../../specs/behavior/evaluation-context.schema.json) | Typed projection | Closed namespaces, exact tagged values, phase/origin constraints and actual principal. |
| [specs/behavior/initial-state.schema.json](../../specs/behavior/initial-state.schema.json) | Fixture preparation | Typed state, profile/restore recipe and separate FIXTURE_PREPARED origin. |
| [specs/behavior/principal-fixture.schema.json](../../specs/behavior/principal-fixture.schema.json) | Intended identity | Credential/session/DB intention without an authentication verdict. |
| [specs/behavior/response.schema.json](../../specs/behavior/response.schema.json) | HTTP/session projection | Ordered headers/cookies, attributes, body, principal, session and effect references. |
| [specs/observation/effect-occurrence.schema.json](../../specs/observation/effect-occurrence.schema.json) | Delivery identity | Business application separated from logical effect, attempt, retry, receipt and acknowledgement. |
| [specs/observation/effect-payload.schema.json](../../specs/observation/effect-payload.schema.json) | Required effect shapes | Closed DB/transaction/session/cache/file/delivery/error payload variants. |
| [specs/observation/execution-witness.schema.json](../../specs/observation/execution-witness.schema.json) | Actual executable provenance | Source/target-specific build/load/process/environment/endpoint/role qualification chain. |
| [specs/observation/local-identities.schema.json](../../specs/observation/local-identities.schema.json) | Pre-target identities | Side-local scalar identities and capture provenance with no case/pair dependency. |
| [specs/observation/step-occurrence.schema.json](../../specs/observation/step-occurrence.schema.json) | Workflow execution | Logical step/visit and physical occurrence, guards, terminal and actual contexts. |
| [specs/operations/recovery-journal.schema.json](../../specs/operations/recovery-journal.schema.json) | Emergency reconciliation | Bounded phase/sequence/receipt/reservation records without acceptance authority. |
| [specs/project/consumption-coverage.schema.json](../../specs/project/consumption-coverage.schema.json) | Positive reuse evidence | Examined inputs/roots/categories, unresolved/excluded mechanisms and closure controls. |
| [specs/project/execution-admission.schema.json](../../specs/project/execution-admission.schema.json) | Credential admission | Explicit original/variant/refused/static outcomes and original versus variant credential status. |
| [specs/project/execution-variant.schema.json](../../specs/project/execution-variant.schema.json) | Transformed execution identity | Distinct original/variant manifests/hashes, exact edits, transform, review and limitations. |
| [specs/project/filesystem-profile.schema.json](../../specs/project/filesystem-profile.schema.json) | Relevant filesystem identity | Exactly one treatment for each supported property; required values need capture/mechanism refs. |
| [specs/verification/occurrence-pairing.schema.json](../../specs/verification/occurrence-pairing.schema.json) | Cross-side alignment | Pairing, first divergence, unmatched occurrences and cleanup under a fixed policy. |
| [specs/verification/operational-slo.schema.json](../../specs/verification/operational-slo.schema.json) | Window obligation | Fixed tested windows, sample/latency/error limits distinct from mean trial quantiles. |

### Validation and preservation

Executed `python3 tools/docs/validate.py`: PASS — 149 schemas, 301 positive/negative fixtures (79 added), 71 command/tool entries, 837 local links, 40 SQL tables across two unchanged migrations, 10 unchanged approval vectors, and ownership/immutability/rollback structural checks. `git diff --check` passed. The targeted stale-contract scan across 92 current prose owners/indexes found no unexplained matches; historical review/schema branches remain deliberately labeled history. Both review inputs, both migrations, approval vectors and the validator are byte-identical to the task baseline. No Go files, runtime tests, adapters, verifier/interpreter or tools were implemented. Runtime mutation, security, recovery, client and statistical experiments remain unexecuted.

Scope: 72 existing files changed and 17 necessary new files. The supplied review-2.md was already present and is preserved. The three-way distinction remains: resolved specification decisions, passing structural validation, and pending implementation/runtime qualification.
