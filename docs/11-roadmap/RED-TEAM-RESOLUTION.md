# Post-red-team revision report and disposition matrix

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
| F16: Two connections both contain an `orders` table | Accept defect; revise canonical contract | [DATABASE-COMPARISON](../05-verification/DATABASE-COMPARISON.md) | Add connection/catalog/schema identity Priority: P1. Machine owner: specs/observation/effects.schema.json; ADR 0005. Runtime evidence is still required. | RESOLVED |
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
| D16 | Holdout/quality/retraction | LEARNING-SYSTEM | Pin consumers; negative mutants and independent provenance; global activation M7 | RESOLVED |
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

Order: M0 durable foundation → M1 falsifiable stateful experiment → M2 real reference characterization → M3 exact resolution/decisions/model → M4 accepted slices → M5 bounded repair/interfaces → M6 project/performance/release; optional M7 learning; M8 capability expansion. TESTING owns X01–X40; all are planned, not executed here.

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
