# Rewrotron Pre-Implementation Red-Team Review

## 1\. Executive Summary

**Rewrotron has a credible architectural direction, but its specifications are not yet a complete executable contract for safely migrating a real application.** Foundation development can begin. Customer migration, broad adapter support, autonomous repair, and release-readiness claims require several decisions first.

The principal risks are:

- A migration plan’s approved hash cannot be reconstructed unambiguously from its current schema.
- Requiring the entire confirmed corpus before completing any unit can deadlock the vertical-slice dependency graph.
- The Behavior IR describes business rules without defining enough executable semantics to reconstruct ordinary conditional, stateful applications.
- Observation completeness is a declared status without a sufficiently specified evidence mechanism.
- Durable execution, historical execution projections, dependency decision updates, and selective invalidation have unresolved operational contracts.
- A passing corpus can reinforce incorrect contracts and contaminate learning unless the oracle is independently challenged.

This is a **specification-only repository**. There are no Go source files, `go.mod`, working adapters, engine binaries, runtime fixture corpus, or actual `SKILL.md` packages. Accepted ADRs document decisions; they do not demonstrate implementation.

The existing validator passed:

```
96 schemas
89 positive/negative fixtures
62 command/tool entries
251 local links
35 initial SQL tables
```

Its stated scope excludes Go, race, framework, and runtime tests. Additional read-only probes exposed schema and SQL gaps described below. All 195 repository files matched their review-start hashes; this review changed no files.

### Document authority map

The repository’s authority ladder is explicit, but some higher-authority text contradicts the newer detailed specifications.

| Material | Authority and current interpretation |
|---|---|
| `AGENTS.md` | Supreme directive. Its absolute equivalence language, “tested reference” claims, and SSE transport baseline require correction. |
| `PRODUCT.md` | Product scope and success criteria. Correctly limits equivalence claims to evidenced scope. |
| `SYSTEM-DESIGN.md`, accepted ADRs | Architectural decisions and boundaries. Some lifecycle and operational details conflict with owning contracts. |
| Domain and Behavior documents | Canonical meanings, IR records, provenance and lifecycle. Several descriptive requirements lack corresponding machine representation. |
| `specs/`, CLI/MCP specifications | Machine contracts. Structural validation passes, but semantic completeness is not established. |
| Storage and compatibility specifications | Persistence, verdict and comparison contracts. Important requirements depend on future service validation. |
| Executable validator and initial SQL | Evidence of structural consistency and selected SQL constraints only. |
| Existing implementation | No Go implementation exists to assess. |
| `REWROTRON_PROJECT_PLAN.md` | Explicitly exploratory, subordinate to owning specifications. Its historical phase order and examples must not drive implementation. |
| `DOCUMENTATION-AUDIT.md` | Historical review record, not a current inventory or implementation-status report. |
| Summary documents | `GOALS`, `NON-GOALS`, `ARCHITECTURE`, `COMPATIBILITY`, and workflow aliases duplicate summaries. They should continue linking to owners rather than defining alternatives. |

This interpretation follows the repository’s documentation authority map (docs/README.md:7). Contradictions are reported rather than silently resolved.

## 2\. Architecture Strengths

Several difficult decisions are already sound **as specified**:

- **Behavioral slices are the migration unit.** Shared helpers and files support externally verifiable operations rather than defining migration boundaries.
- **Source understanding is separated from target generation.** The three IR tiers avoid a direct translator for every source/target pair.
- **The deterministic engine owns acceptance.** Agents propose; engine services control state, gates, execution and verdicts.
- **Detection, recommendations, intent and results are separate.** Requested runtime versions cannot be silently replaced by easier recommendations.
- **Verification uses independent writable environments.** Sharing a writable customer database is explicitly prohibited.
- **Uncertainty remains visible.** `UNKNOWN` cannot become `PASS`, and missing capture cannot mean “no effects.”
- **Normalization preserves relationships.** Session tokens and generated IDs cannot simply be masked independently.
- **Dead-code omission requires a bounded proof bundle.** The current specification explicitly rejects universal cryptographic proof of dead PHP.
- **Recovery acknowledges filesystem reality.** Multi-file publication uses a journal rather than pretending it is one atomic filesystem transaction.
- **Global learning is optional and reviewed.** A single successful repair does not authorize promotion.
- **Transparent execution is an engine requirement.** Progress has denominators, scope and freshness rather than elapsed-time guesses.

These strengths are documented in compatibility (docs/05-verification/COMPATIBILITY-SPEC.md:5), classification (docs/03-analysis/CODE-CLASSIFICATION.md:1), recovery (docs/08-storage/RECOVERY.md:1), and target authority (docs/04-migration/TARGET-SPECIFICATION.md:5).

The review must not confuse “already specified correctly” with “already implemented or proven.”

## 3\. Critical Risks

### CR-01 — Approval identity is underspecified

The confirmation protocol hashes:

```
adapter_versions
ordered_units
dependency_edges
```

The plan schema contains `units` with nested dependencies, but no `adapter_versions`, `ordered_units`, or `dependency_edges` fields. It rejects additional fields.

A probe confirmed that adding `adapter_versions` to an otherwise valid plan is rejected.

The units and edges could be derived, but the derivation and adapter-version source are not defined. After an engine upgrade, reconstructing the approved subject could depend on the current registry rather than the originally approved adapters.

**Required resolution:** Define a versioned hash-preimage schema, explicit projections from serialized records, and immutable adapter manifest references. Test reconstruction after restart and upgrade.

Evidence: hash construction (docs/02-domain/CONFIRMATION-PROTOCOL.md:27), plan schema (specs/migration/plan.schema.json:41).

### CR-02 — Incremental migration can deadlock

Differential testing requires the **full confirmed corpus** for a unit completion claim. Migration units require prerequisite units to be VERIFIED/COMPLETED before generating dependent units.

For example:

1. Login is a prerequisite of checkout.
2. Login is generated.
3. Its completion requires the checkout case to pass.
4. Checkout cannot generate until login completes.

This is an architectural contradiction, not merely excessive testing.

**Required resolution:** Define unit acceptance over its approved cases and applicable implemented regression scope. Require the entire confirmed corpus for project acceptance. Every result must identify its exact acceptance scope.

Evidence: whole-corpus requirement (docs/05-verification/DIFFERENTIAL-TESTING.md:26), unit dependencies (docs/04-migration/MIGRATION-UNITS.md:9), independent completion ADR (docs/01-architecture/ADR/0004-vertical-slice-migration.md:20).

### CR-03 — The Behavior IR is insufficiently executable

`DECLARATIVE` business rules contain an open `data` object. Authentication, authorization and preconditions are also open objects.

A probe accepted:

```
{
  "representation": "DECLARATIVE",
  "evidence_ids": [],
  "data": {"operator": "invented_operator"}
}
```

Structural openness is legitimate when a semantic validator exists. Here, the operator language, type system, supported predicates, effect semantics and validation rules remain undefined.

The workflow documentation requires bounded branches, initial principal and expected final state. Its schema represents ordered steps and bindings without those explicit fields; a branch field is rejected.

**Required resolution:** Define a bounded executable rule vocabulary and stateful scenario model. Unsupported computations must produce a specific capability blocker. Natural-language descriptions must never substitute for executable semantics.

Evidence: business-rule schema (specs/behavior/behavior-contract.schema.json:103), workflow requirements (docs/03-analysis/WORKFLOW-DISCOVERY.md:7), workflow schema (specs/behavior/workflow.schema.json:1).

### CR-04 — `COMPLETE` needs an attestation mechanism

The observation specification correctly distinguishes complete and incomplete dimensions. It does not sufficiently define what establishes completeness.

A CI3 DB driver probe can miss a direct PDO connection. A filesystem snapshot can miss writes outside declared volumes. A response may finish before a worker sends an email.

The engine cannot accept “the probe reported COMPLETE” as independent evidence that the probe covers all relevant paths.

**Required resolution:** Bind completeness to supported capture mechanisms, declared boundaries, denied escape paths, instrumentation tests and closure conditions. An uncovered mechanism must invalidate the affected dimension.

Evidence: capture dimensions (docs/03-analysis/RUNTIME-OBSERVATION.md:9).

### CR-05 — Durable work has no defined process owner

Closing an MCP connection must not cancel the engine, but the architecture does not identify the process that owns work after the stdio server exits.

A database operation row is not a running worker. Lease renewal, subprocess control, heartbeat generation and cleanup require an owner.

**Required resolution:** Choose foreground-only execution with honest interruption semantics, or an engine-owned local supervisor. Define startup, shutdown, attachment and resource ownership.

Evidence: disconnect guarantee (docs/06-agent/EXECUTION-OBSERVABILITY.md:121), recovery guarantee (docs/08-storage/RECOVERY.md:43).

### CR-06 — Selective invalidation lacks a sufficiently typed dependency contract

The documents require reverse dependencies and preservation of unrelated work. Generic artifact links and embedded JSON references do not yet establish which inputs a result depends on.

A runtime change can affect generated code, native dependencies, shared helpers, serializers and all tests. A controller edit can affect multiple workflows through authentication or transaction behavior.

**Required resolution:** Define dependency-edge types and validity predicates. Unknown impact must conservatively invalidate results. Retained files, completed historical work and current verified work must be separate concepts.

### CR-07 — The oracle can be self-consistent and wrong

The same agent can infer a contract, generate cases, write the replacement and propose normalization. All tests can pass while omitting the same authorization branch.

Human approval and deterministic comparison do not fix an incomplete oracle automatically.

**Required resolution:** Independent business assertions, adversarial negative cases, comparator mutation tests and held-out workflows must challenge inferred contracts before they establish reusable knowledge.

## 4\. Architecture Contradictions

| Contradiction | Evidence | Precise resolution |
|---|---|---|
| “Proves behavioral equivalence” versus finite-corpus compatibility | AGENTS (AGENTS.md:13) versus compatibility scope (docs/05-verification/COMPATIBILITY-SPEC.md:5) | Reserve proof for an explicitly bounded formal claim. Ordinary reports demonstrate observed or contract equivalence. |
| “Initial tested reference adapter” versus no implementation | AGENTS (AGENTS.md:131) versus repository status (docs/README.md:3) | Say “planned reference adapter”; publish maturity only with executed evidence. |
| stdio/SSE versus pinned stdio and deferred Streamable HTTP | AGENTS (AGENTS.md:150) versus MCP baseline (docs/06-agent/MCP-SPEC.md:13) | Align the constitution with the pinned transport decision. |
| Gate 3 from VERIFIED versus BENCHMARKED | system design (docs/01-architecture/SYSTEM-DESIGN.md:144) versus lifecycle (docs/02-domain/PROJECT-LIFECYCLE.md:19) | Use the canonical lifecycle and explicit benchmark prerequisites everywhere. |
| Independent slice completion versus full-corpus unit acceptance | CR-02 | Separate slice, regression and project acceptance scopes. |
| Branching workflow requirements versus linear workflow schema | CR-03 | Add typed branching/state semantics, or explicitly support only enumerated linear scenarios initially. |
| Exact resolved versions versus reuse of the intent schema | resolved target (specs/migration/plan.schema.json:165) references intent schema (specs/project/target.schema.json:16) | Separate constraint-bearing intent from exact resolved environment records. |
| Agent cannot mutate SQLite versus a shared writable local account | agent model (docs/06-agent/AGENT-MODEL.md:5), acknowledged trust limit (docs/02-domain/CONFIRMATION-PROTOCOL.md:37) | State that enforcement applies to mediated engine calls. Stronger hostile-agent isolation needs OS-level separation. |
| Disconnect-independent work versus undefined worker owner | CR-05 | Specify process ownership and lifecycle. |
| Historical state at a watermark versus mutable current rows | execution projection (docs/06-agent/EXECUTION-OBSERVABILITY.md:153) | Persist historical projections or enough events to reconstruct them; otherwise distinguish current state from historical event windows. |
| Preserve all historical evidence versus bounded long-term local storage | retention (docs/08-storage/ARTIFACT-STORE.md:13) | Establish quotas, archive manifests and explicit retention classes without silently deleting approval evidence. |
| Reproducible migrations versus evolving skills and knowledge | skill evolution (docs/07-learning/SKILL-EVOLUTION.md:20) | Pin every consumed skill/tool/knowledge revision, retain bytes and require explicit upgrades. |
| Frozen dependency decisions versus changing execution statuses | resolution lifecycle (docs/04-migration/DEPENDENCY-RESOLUTION.md:37) | Freeze decision content; track implementation/verification progress separately. |
| Runtime control parity versus instrumentation changing execution | observation control (docs/03-analysis/RUNTIME-OBSERVATION.md:27) | Control fixtures establish bounded evidence, not universal non-interference. Publish the residual limitation. |

Several apparent contradictions already have valid resolutions: local-first permits optional external AI; modernization requires approved intentional changes; autonomy operates within persisted approvals; dead-code omission is conditional on a closed analysis boundary. Those are not unresolved contradictions by themselves.

## 5\. Practical Failure Scenarios

The table includes unsafe outcomes, design deadlocks and safe refusals whose recovery path remains unspecified. **A correctly blocked operation is a capability limitation, not evidence that the implementation is unsafe.**

Priorities apply to the affected capability, not automatically to all repository development.

| ID | Scenario | Why Current Design Fails | Impact | Detection | Required Design Change | Priority |
|---|---|---|---|---|---|---|
| F01 | Database rows select controller classes | Static route inventory cannot close database-driven dispatch | Hidden endpoints; omission risk | Unresolved dispatch plus supplied routing data | Version configuration/data roots and preserve uncertainty | P1 |
| F02 | Quarterly cron uses an apparently unused helper | Observation window misses the job | Required billing behavior omitted | Operational manifest and schedule inventory | Require operational-root closure for DEAD proofs | P0 before omission |
| F03 | Production generates PHP with `eval` | Repository inventory lacks generated semantics | Unknown business logic | Generated-code and eval findings | Capture safe generated artifacts or block scope | P1 |
| F04 | Customer patched the framework/vendor tree | Version name does not describe actual behavior | Wrong adapter assumptions | Content comparison and modified-vendor inventory | Bind capabilities to content and modification findings | P1 |
| F05 | Production feature flag enables a dormant path | Fixture configuration differs | False coverage confidence | Configuration-profile reconciliation | Version approved environment profiles | P1 |
| F06 | PHP 7.4 application requires an unavailable extension | Runtime cannot boot; fallback would change behavior | Observation blocked | Doctor and startup evidence | Runtime acquisition/retention decision; no host fallback | P1 |
| F07 | Source changes and changes back during scanning | Metadata rescans do not establish an atomic snapshot | Mixed baseline | Snapshot identity and consistency checks | Execute from an immutable coherent source snapshot | P0 before baseline execution |
| F08 | User override says PHP 8.2, actual runtime is 7.4 | Asserted intent can be confused with measured fact | Invalid feasibility analysis | Doctor compares measured and asserted facts | Preserve assertions separately from runtime attestations | P1 |
| F09 | Checkout branches on inventory and account status | Workflow schema lacks branch/state representation | Important cases cannot be represented directly | Schema rejection; unresolved predicates | Typed branches or explicit separate bounded scenarios | P1 |
| F10 | HTTP response precedes a queued email | Immediate post-state misses later effects | False parity | Queue/sink correlation and completion barriers | Define asynchronous completion boundaries | P0 before effect acceptance |
| F11 | Application exposes WebSockets | HTTP/CLI/cron/event model lacks connection semantics | Unsupported behavior | Protocol inventory | Declare unsupported; add stream protocol capability later | P2 |
| F12 | Helper opens PDO outside CI3’s DB driver | Driver probe misses writes | Database dimension falsely complete | Driver inventory, snapshots, fault fixtures | Completeness attestation covering all connections | P0 before PASS |
| F13 | Native extension writes files or contacts a service | Language hooks cannot observe it reliably | Missing effects | OS/network containment and volume monitoring | Capability-specific observation or UNKNOWN | P1 |
| F14 | Failed transaction consumes auto-increment values | Row rollback alone does not reset sequence state | Later IDs diverge | Sequence and retry cases | Restore/check sequence behavior explicitly | P1 |
| F15 | Request creates temporary tables or performs DDL | Transaction wrapper cannot undo all effects | Replay contamination | SQL feature inventory and reset tests | Fresh DB instance or supported full reset | P0 before replay |
| F16 | Two connections both contain an `orders` table | Canonical effects lack connection identity | Mutations attributed to wrong database | Connection-qualified schema inspection | Add connection/catalog/schema identity | P1 |
| F17 | Payment provider offers no safe sandbox | Real charge cannot be replayed safely | Verification blocked | Endpoint/effect classification | Verified mock/retained boundary or explicit unsupported scope | P0 |
| F18 | Worker crashes after sink acceptance but before acknowledgement | At-least-once execution duplicates delivery | Wrong counts or false diagnosis | Sink receipts and retry scenarios | Model acknowledgements, idempotency and delivery semantics | P1 |
| F19 | Redis holds sessions, locks, queues and feature flags | Resetting only cache keys leaves hidden state | Non-reproducible workflows | Namespace inventory and restore checks | Separate stores/namespaces and reset contracts | P1 |
| F20 | CI3 session contains unsafe serialized objects | Safe target decoder cannot reproduce arbitrary objects | Continuity blocked | Session-format inspection | Safe bridge or approved re-login exception | P1 |
| F21 | Logout on target leaves source session valid | Independent login tests miss shared revocation | Parallel-rollout security failure | Cross-side continuity/revocation cases | Explicit bridge and rollback state ownership | P1 |
| F22 | Normalizer masks CSRF tokens but misses replay acceptance | Value variation passes while lifecycle is wrong | Security behavior differs | Invalid/replayed/cross-user token tests | Relation plus lifecycle assertions | P0 before auth acceptance |
| F23 | Timestamp determines refund eligibility | Time-window normalization hides a business decision | Wrong refunds accepted | Boundary-time cases | Controlled clock and business-time predicates | P1 |
| F24 | SQL without ORDER BY returns varying results | Repeated runs cannot prove order irrelevant | False normalization or persistent mismatch | Repeated fixtures and contract review | Preserve order unless an approved policy establishes otherwise | P1 |
| F25 | New PDF library emits different metadata/layout | Hash differs; “same-looking” is undefined | Replacement blocked | Format-specific differential cases | Approved comparator with explicit properties | P2 |
| F26 | Request uses `?filter=` | Case schema rejects an empty query value | Valid legacy input cannot be represented | Confirmed schema rejection | Lossless request representation allowing empty values | P1 |
| F27 | MCP client rounds a large JSON numeric ID | Precision can be lost before core parsing | Wrong identity or amount | End-to-end numeric transport vectors | Exact typed values/raw payload refs | P1 |
| F28 | Target adapter upgrades after plan approval | Approved hash input lacks serialized adapter versions | Approval cannot be reproduced | Hash reconstruction test | Immutable adapter identity in hash preimage | P0 before Gate 2 |
| F29 | “Resolved” target still contains `13.*` and no runtime | Intent schema accepts it | Execution may resolve differently later | Confirmed schema probe; service check needed | Dedicated resolved-environment schema | P1 |
| F30 | User rejects one dependency replacement | Generated views are read-only; decision mutation flow is unspecified | User cannot concretely configure alternative | Interface contract review | Typed human decision/import operation | P1 |
| F31 | Dependency status advances PLANNED → VERIFIED | Updating frozen decision artifact can change plan identity | Reapproval loop | Approval/current-subject tests | Separate immutable decision from progress | P1 |
| F32 | Login must complete before checkout generation | Login completion requires checkout corpus | Dependency deadlock | Two-unit scheduling fixture | Scoped unit acceptance | P0 |
| F33 | Duplicate mission rows represent identical work | SQL uniqueness is by mission ID, not work identity | Duplicate expensive work | Confirmed in-memory SQL probe | Deduplication invariant and atomic resource claims | P1 |
| F34 | Distinct missions edit the same shared auth file | Per-mission fences do not establish shared-file ownership | Lost edits or publication races | Overlapping-patch race tests | Serialized publication/resource ownership | P1 |
| F35 | Independent unit A increments project revision | Unit B’s captured revision becomes stale despite unrelated edits | Starvation or effective serialization | Concurrent disjoint-unit fixture | Define revision/conflict granularity | P2 |
| F36 | PHP target changes 8.3 → 8.4 at 70% completion | Reverse impact graph is not sufficiently typed | Excess invalidation or stale acceptance | Transitive-change fixtures | Versioned dependency graph and reuse certificates | P1 |
| F37 | MCP process exits during observation | No durable worker owner is defined | Work dies or resources leak | Kill-client/process test | Explicit supervisor/foreground lifecycle | P1 |
| F38 | Client requests historical state at sequence N | Current mutable rows may contain later state | Inconsistent timeline/overview | Interleaved-write pagination test | Historical projections or current-state distinction | P1 |
| F39 | Queue event must be linked to originating workflow | Trace schema rejects run/workflow/process/span fields | Causal evidence lost | Confirmed schema rejection | Canonical correlation envelope | P1 |
| F40 | Status client polls continuously | Auditing reads can persist millions of rows | DB growth; write contention | Event-volume measurements | Bounded/coalesced inspection auditing | P2 |
| F41 | Thousands of cases retain repeated full captures | Historical reachability preserves all artifacts | Disk exhaustion | Quota forecasts and disk-watermark checks | Retention classes, archives and admission limits | P1 |
| F42 | Backup copies DB and artifacts at different times | DB consistency alone does not guarantee artifact consistency | Unrestorable approved evidence | Restore/reference verification | Coordinated backup manifest and retention barrier | P1 |
| F43 | Workspace is placed on a network filesystem | WAL requires local shared-memory coordination | Storage failure or unsafe assumptions | Doctor filesystem-profile check | Supported filesystem policy | P1 |
| F44 | Replacement SDK is API-compatible but changes retry behavior | Metadata and signatures do not establish behavior | Duplicate payments or calls | Dependency characterization cases | Effect/error/retry contracts | P1 |
| F45 | Three generated fixtures share the same incorrect oracle | Promotion threshold counts correlated evidence | Bad global pattern becomes trusted | Independent holdout and mutation tests | Independence criteria and retraction graph | P1 before promotion |
| F46 | Experimental tool omits difficult traces | Apparent success rises because evidence worsens | False learning and confidence | Capture-denominator and quality checks | Benchmark evidence quality, not just task success | P1 before activation |
| F47 | Rust/Python source needs declarations outside the current symbol vocabulary | Source IR categories reflect initial PHP needs | Lossy extraction or schema change pressure | Neutrality fixtures | Source semantic facets and explicit unsupported records | P2 |
| F48 | AST or compiler diagnostic contains a source secret | Runtime redaction alone misses derived artifacts | Cross-client or global leakage | Secret canaries across every artifact path | Private-source policy and derived-data redaction | P0 before export |
| F49 | Target performance equals source within noise | Upper confidence bound ≤1 is usually inconclusive at equality | Readiness stalls without actual regression | Identical/noisy performance fixtures | Define statistical claim and explicit acceptance budget | P1 |
| F50 | Target writes new session/payment state during cutover | Same schema does not make those writes reversible | Rollback loses or corrupts state | Staging rollback rehearsal | Data/session reconciliation contract | P1 before release |
| F51 | Multi-process WAL uses an affected SQLite version | Dependency selection does not yet establish a fixed engine version | Rare database corruption | Embedded-version/build verification | Pin a release containing the WAL-reset fix | P0 before multi-writer release |

## 6\. Source Analysis Risks

The current static-analysis contract is appropriately conservative. Its difficult implementation problem is **closing uncertainty without pretending runtime samples close it**.

### Dynamic constructs

| Legacy feature | Static analysis boundary | Dynamic observation boundary |
|---|---|---|
| Computed `include`/`require`, custom autoloaders | Candidate paths may be unbounded | One loaded file does not exclude others |
| Variable classes/methods, magic methods, reflection | Dispatch requires configuration/type constraints | Observed dispatch establishes only concrete executions |
| Containers, service locators, global/static state | Bindings depend on bootstrap and environment | Fresh requests may miss persistent-worker state |
| Hooks, middleware, listeners, framework extensions | Registration may be conditional or dynamic | Early exits and cached responses can bypass hooks |
| `eval`, generated PHP, runtime patching patterns | Arbitrary generated semantics cannot be safely established generally | Executing them requires containment and may still leave unobserved alternatives |
| Vendor modifications, mixed framework versions | Version labels are insufficient | Probe compatibility must be demonstrated against actual modified content |
| Raw SQL, procedures, triggers, multiple connections | PHP call graphs do not capture all database logic | Driver instrumentation alone misses server-side effects |
| Cron, CLI, queues, callbacks, admin paths | Operational configuration is part of root discovery | Observation windows rarely cover all schedules or identities |
| Flags, environment variables, DB-driven configuration | One configuration is one analysis profile | Safe fixtures may not reproduce production-only conditions |
| Templates, assets, include-time side effects | Files are not merely declaration containers | Executed HTTP routes do not cover every public resource |

The specification already records unresolved expressions and affected scopes. Preserve that rule. Do not make “Generic PHP fallback” mean “generic PHP understanding”; it may provide inventory while semantic extraction remains unsupported. See static analysis (docs/03-analysis/STATIC-ANALYSIS.md:1).

### DEAD must remain a bounded claim

The current DEAD definition is stronger than “no reachable evidence” and should remain stronger.

Automatic omission requires:

1. An explicitly closed operational-root boundary.
2. Complete parsing for the candidate’s relevant region.
3. No reachable static path.
4. No unresolved candidate path.
5. A usable runtime corpus with zero hits.
6. Current baseline, graph and analyzer identities.
7. The proof included in the approved plan.

The remaining missing decision is **what allows the root boundary to be declared closed**. Customer attestation can define the boundary but cannot prove no forgotten production scheduler exists.

Report:

> DEAD under approved analysis boundary B.

Do not report:

> Proven never used.

Also separate omission from replacement. Reimplementing a used formatter with native target functionality does not make its source dependency dead.

## 7\. Runtime Observation Risks

A container profile is necessary but does not establish a working legacy environment.

The actual boot contract must include:

- Exact runtime and extension inventory.
- Startup command and readiness checks.
- DB connections, schemas, routines and settings.
- Redis/session/queue namespaces.
- Writable paths and permissions.
- TLS, DNS and mock endpoint configuration.
- Synthetic identities and permissions.
- Cron and worker launch policy.
- Probe installation and control-run results.

Most of these inputs are acknowledged in PROJECT-INPUTS (docs/10-development/PROJECT-INPUTS.md:3). Their runtime validity remains **UNKNOWN**.

### Correlation is a machine-contract gap

The trace schema has `trace_id`, project, configuration, optional case and timestamp. It does not admit the full run/workflow/request/process/span correlation requested by the product.

A read-only schema probe rejected these added fields. Hiding them in generic nested objects would make correlation adapter-specific.

Introduce a canonical envelope with execution attempt, workflow instance, operation/request, process, trace/span and causal parent references. Timestamps cannot replace causal links.

### Instrumentation cannot promise universal non-interference

An instrumented/uninstrumented control fixture is valuable, but it only detects interference in tested cases.

Instrumentation can change:

- Timing and race outcomes.
- Output buffering.
- Shutdown behavior.
- Connection lifetime.
- SQL performance.
- Memory pressure and timeout branches.

Record probe limitations and require control tests for each claimed capture capability.

### Repetition is insufficient for rare behavior

Three source repetitions identify some variability. They cannot establish that a monthly job, rare authorization branch or occasional external-service failure is absent.

Observation needs scenario diversity and explicit state partitions, not only repeated identical inputs.

## 8\. Behavior Model Risks

The Behavior Model must describe **what is asserted**, **under what conditions**, and **with which evidence**.

A useful contract requires:

| Element | Current position | Required clarification |
|---|---|---|
| Preconditions | Open object | Typed fixture, principal, state and environment predicates |
| Conditional rules | Open declarative data | Versioned predicate/action vocabulary |
| Postconditions | Distributed through response/effects | Explicit cross-dimension assertions |
| Invariants | Descriptive | Scope, evaluation points and executable operators |
| State transitions | Implicit in steps and mutations | Named states, guards, transitions and failure behavior |
| Transactions | Boolean plus mutations | Connection, boundaries, rollback and causal ordering |
| Side effects | Generic objects | Typed destinations, payloads, counts, delivery and reversibility |
| Provenance | Good evidence framework | Evidence eligibility and invalidation rules for each assertion |
| Confidence | Estimate, not proof | No threshold may substitute for missing required evidence |
| Dynamic values | Approved relation operators | Complete typed validator semantics |
| Async completion | Not adequately represented | Bounded waits, acknowledgements, barriers and terminal conditions |
| Unsupported behavior | Blockers/UNKNOWN | Machine-readable capability reasons and affected scope |

For:

```
if A: X
else if B: Y
else: Z
```

Rewrotron needs either an executable branch model or three explicit scenarios with guards and asserted boundaries. A prose rule describing these branches is not sufficient for deterministic generation.

### Terminology

Use precise claims:

- **Observed equivalence:** source and target matched specific captured executions.
- **Contract equivalence:** target satisfies the approved executable contracts and comparison policies.
- **Semantic equivalence:** a defined semantic claim supported by a specified proof method; do not imply this from finite tests.
- **Unverified behavior:** identified behavior without current verification.
- **Unknown behavior:** behavior whose relevant semantics or observation are unresolved.
- **Unsupported behavior:** known required behavior outside declared capabilities.
- **Intentional change:** an approved exception, disclosed separately from strict parity.

The current compatibility document largely follows this distinction. The constitution must align with it.

### Important information must survive every IR transition

Preserve provenance, unresolved alternatives, dispatch precedence, coercion semantics, state ownership, causal order, transaction boundaries, error behavior and observation limitations.

A smaller IR that drops these facts creates cleaner diagrams and incorrect migrations.

## 9\. Dependency / Target Environment Risks

The dependency specification has the right conceptual model: inventory, actual usage, capability coverage and behavioral verification are separate.

The strategies are sensible:

```
RETAIN
DIRECT_UPGRADE
ADAPTER
REPLACEMENT
REIMPLEMENT
```

`BLOCKED` and “user decision required” are states, not successful strategies. Partial versus full reimplementation should be represented as the **required capability scope**, rather than inferring that the entire upstream library must be recreated.

### Remaining dependency decisions

1. **Decision input is incomplete.** Recommendations and dependency YAML are generated views. The CLI/MCP catalogs do not define a sufficiently explicit way to reject a replacement, choose an adapter, or submit an alternative strategy.
2. **Decision and progress are mixed.** Advancing a resolution from PLANNED to VERIFIED must not rewrite the immutable decision approved by Gate 2.
3. **Compatibility phases need precision.** Feasibility can establish constraints and proposed capability coverage before generation. Behavioral parity normally requires the implementation. “Unknown compatibility blocks execution” must identify which compatibility question is unknown, or planning becomes circular.
4. **Graph/API/feasibility artifact types remain insufficiently fixed.** Open artifacts do not define constraint interpretation, transitive resolution or invalidation consistently.
5. **Metadata freshness is not universal.** Offline operation must distinguish unavailable evidence from a cached, dated result.
6. **Modified, proprietary and production-only packages need content-based inventory.** Registry metadata may not describe the installed dependency.
7. **Native dependencies extend beyond language packages.** OS libraries, architecture, extensions and executable tools affect feasibility.
8. **Licensing is a resolution constraint.** A technically equivalent replacement may be unacceptable for redistribution or customer deployment.

### Target configuration

Requested versions correctly override recommendations. The remaining machine-contract gaps are:

- Intent and exact resolution reuse one schema.
- `runtime` is not structurally required in the target schema.
- A “resolved target” can contain wildcard versions and omit runtime.
- Deployment model and operational topology are not explicitly represented.
- Standard-library targets need an explicit convention for framework/version fields.
- Adapter capability profiles remain too open to guarantee consistent constraint intersection.

Laravel 13’s PHP 8.3 minimum agrees with the reference direction. A recommendation of PHP 8.1 cannot satisfy that selected framework. [Laravel’s release requirements](<https://laravel.com/framework/docs/13.x/releases>) support this constraint; actual project resolutions still need pinned evidence.

## 10\. Differential Verification Risks

Comparison is practical for a bounded supported profile. It is not automatically practical for every application effect.

### Identity mapping

The existing requirement for bijective generated-ID mapping is sound. Implementation must detect:

- Swapped entities.
- Reused mappings.
- Missing references.
- Ambiguous duplicate rows.
- IDs appearing in filenames, emails, callbacks or session state.
- IDs intentionally exposed as business identifiers.

An ID policy cannot normalize each field independently.

### Nondeterminism

Repeated output changes can indicate randomness, unstable ordering, environmental variation **or a defect**.

Rules must identify:

- Exact paths.
- Allowed format and range.
- Equality/inequality relations.
- Cross-step lifecycle.
- Business significance.
- Evidence and approval.
- Negative mutations the rule must still reject.

Time windows must not hide wrong expiry, retention, billing or refund logic.

### Stateful and concurrent verification

Sequential request tests do not prove:

- Lost-update prevention.
- Lock ordering.
- Idempotency.
- Duplicate-message handling.
- Deadlock recovery.
- Concurrent logout or session revocation.
- Read-after-write behavior.

Concurrency scenarios need bounded schedules or explicit coordination points. Merely running cases concurrently in independent environments tests isolation, not application concurrency semantics.

### Safe side-effect matrix

| Class | Example | Permitted compatibility execution |
|---|---|---|
| PURE | Formatting or calculation | Controlled inputs; compare outputs and errors |
| READ\_ONLY | Database/API lookup | Isolated fixtures or approved recorded responses; verify that “read” causes no relevant mutation |
| LOCAL\_WRITE | Upload, cache, session, DB mutation | Independent resettable stores; compare committed effects and restoration |
| EXTERNAL\_WRITE | Email, SMS, webhook, object upload | Closed-network sinks with destination/payload/count assertions |
| IRREVERSIBLE | Payment, external transaction, non-reversible production action | Verified simulator or safe provider sandbox; otherwise do not execute and report unsupported/UNKNOWN |

An HTTP GET is not proof of READ\_ONLY. Legacy GET routes can send mail or mutate records.

### Verdicts

Keep the current three verdicts:

- `PASS`: all required dimensions complete and compliant.
- `MISMATCH`: at least one prohibited difference established.
- `UNKNOWN`: required evidence cannot be safely or completely evaluated.

`EXPECTED_CHANGE` remains a diagnostic classification backed by an approved exception. It must not become a convenient fourth pass verdict.

### Performance acceptance needs a defined statistical claim

The prescribed upper confidence bound of at most 1.0 is stringent. A target whose performance is equal to the source can remain inconclusive because measurement noise spans 1.0.

That is not a justification for an automatic tolerance. The specification must choose between demonstrating improvement, bounded non-inferiority, or an explicitly approved budget.

Five paired trials also require adequate successful samples and a justified treatment of correlated measurements. The benchmark schema exposes p95 and peak memory, while p50, p99, CPU, DB/external latency and response size depend on external artifacts rather than a complete typed measurement contract.

Evidence: performance protocol (docs/05-verification/PERFORMANCE.md:5).

## 11\. Database / SQLite Risks

### Application database behavior

Preserving the schema does not preserve connection behavior.

Contracts must account for:

- Connection/catalog identity.
- SQL mode- &#32;and implicit casts.
- Transaction isolation and locking.
- Collations and column-specific comparison.
- Timezone and server-generated time.
- Generated columns, views, triggers and routines.
- Numeric precision and binary values.
- Connection-local variables and temporary tables.
- Corrupt or historically tolerated data.
- Nontransactional tables and DDL.
- Persistent connection state.

The data-model schema has precision and scale, which is useful. It lacks sufficient connection identity and server-side behavioral representation.

MySQL explicitly documents statements that commit implicitly and temporary-table operations that are not undone by rollback. Fresh/resettable environments are therefore essential; wrapping a request in a transaction is insufficient. [MySQL implicit-commit rules](<https://dev.mysql.com/doc/refman/8.0/en/implicit-commit.html>).

### SQL constraints versus service guarantees

An in-memory probe with foreign keys enabled showed:

- A Project A detection can reference a Project B artifact.
- Two leased missions can share the same work fingerprint.
- SQLite reports no foreign-key violation.

The specification delegates semantic ownership to repository services. That is legitimate, but it makes **centralized, mandatory validation** essential. Do not rely on every service author remembering the same checks.

The second result also shows that atomic claiming of one mission row does not prevent duplicate mission rows representing the same work.

Evidence: missions DDL (migrations/0001\_initial.sql:306), service validation (docs/01-architecture/SERVICE-CONTRACTS.md:37).

### Growth estimates

These are illustrative **ASSUMPTIONS**, not measured Rewrotron results:

| Assumed workload | Approximate annual/storage consequence |
|---|---|
| One continuously active worker, heartbeat every 15 seconds | 2\.1 million heartbeat events/year |
| Heartbeat event averages 2 KiB | About 4 GiB/year before indexes and other events |
| One continuously polling observer every two seconds, each read audited | 15\.8 million audit records/year |
| Audit record averages 1 KiB | About 15 GiB/year before indexes |
| 10,000 cases, three source captures plus one target capture, 1 MiB each | About 39 GiB for one corpus pass |
| A 5 GiB database restored by full copying for 10,000 cases | About 49 TiB of copy traffic, before retries |

Content addressing helps only when bytes repeat. Timestamps, tokens, logs and DB dumps can prevent useful deduplication.

Retention should distinguish:

| Data | Recommended lifecycle |
|---|---|
| Approvals, hashes, decision provenance, release manifests | Durable, small, retained |
| Current contracts/cases and required evidence | Retain while active and for agreed audit period |
| Failed experiments and reusable lessons | Retain summaries plus selected evidentiary artifacts |
| Raw traces and full snapshots | Quota-controlled; archive with verified manifests |
| Heartbeats and inspection audits | Bounded/coalesced or summarized under an explicit policy |
| Temporary/unreferenced artifacts | Reconcile, then collect after safe grace period |
| Global knowledge | Separate privacy-reviewed store and migration namespace |

### WAL and version selection

WAL does not remove the single-writer constraint and does not work across network filesystems. Long readers can also prevent useful checkpoint progress. [SQLite WAL documentation](<https://sqlite.org/wal.html>).

A particularly relevant current issue is SQLite’s documented **WAL-reset corruption bug**: the fix is in 3.51.3 and later, with identified backports. It concerns concurrent writing/checkpointing across connections—the proposed access pattern. Pin an embedded SQLite version containing the fix. [SQLite’s WAL-reset advisory](<https://sqlite.org/wal.html#walreset>).

The documentation validator used Python SQLite 3.45.1 for in-memory DDL checks. That neither demonstrates an affected engine deployment nor validates the future Go driver’s WAL behavior.

One project DB plus one optional global DB is sufficient for metadata initially. It is not a complete retention, archive, backup or multi-process ownership strategy.

## 12\. Agent / MCP / Skill Risks

The deterministic boundary is well stated. Its implementation must ensure every CLI, MCP and worker path reaches the same guarded services.

### Authority must remain narrow

Agents may propose:

- Interpretations.
- Plans.
- Patches.
- Diagnostics.
- Knowledge.

They must not establish:

- Runtime observations.
- Completeness.
- Gate approval.
- DEAD proof.
- Comparator success.
- Current result validity.

An agent-produced JSON object claiming `COMPLETE` is not observation evidence.

### Local trust limits

The confirmation protocol correctly acknowledges that unrestricted terminal access under the same account is not cryptographic proof of human intent.

Likewise, an agent with direct write access to the DB, source and verification binaries can bypass mediated policies outside Rewrotron.

The product must distinguish:

1. Engine enforcement for supported interfaces.
2. Trusted local-account operation.
3. Hostile-agent isolation, which requires stronger OS/process permissions.

Do not advertise the third while implementing only the first two.

### MCP portability

The tool catalog is not a tested server. Required protocol tests include:

- Initialization/version negotiation.
- Cancellation.
- Structured outputs and errors.
- Tool schema delivery.
- Offline resolution of schema references.
- Large results and artifact pagination.
- Reconnection after committed mutations.
- Approval-token handling.
- Client disconnect during work.

The MCP standard’s human-in-the-loop recommendations do not themselves provide Rewrotron’s durable approval mechanism. [MCP tool specification](<https://modelcontextprotocol.io/specification/2025-11-25/server/tools>).

### Skills and generated tools

No actual Skill package exists here. Its planned scripts must invoke bounded domain operations; a helper script must not become an unrestricted execution bypass.

The tool registry needs exact content hashes, runtime requirements, capability permissions, evaluation results and revocation status. Selection should begin with exact metadata/capability filters and full-text search. Embeddings are optional discovery aids, never authorization or compatibility evidence.

## 13\. User Experience / Transparency Risks

The execution design is substantially stronger than a generic activity log.

The remaining risks are accuracy and recoverability.

### Historical state is not automatically reconstructable

The API promises state at a fixed `through_sequence`. Current mutable mission/project rows may already include later updates.

A consistent SQLite read transaction gives a current database snapshot. It does not necessarily give the historical state represented by an arbitrary older event watermark.

Choose one:

- Event-complete historical projections.
- Versioned projection snapshots.
- A current-state snapshot plus explicitly historical event window.

Do not label the third as historical state at the watermark.

### Heartbeats are not progress

The specification already distinguishes activity and liveness. Also show:

- Time since meaningful activity.
- Expected wait condition.
- Configured timeout.
- Whether a subprocess is responsive.
- Whether the result is stale or merely waiting.

A live worker repeatedly failing the same experiment should not look productive.

### Visibility layers

| Layer | Information |
|---|---|
| Always visible | Scope, current task/reason, liveness, blockers, required decisions, current target and next action |
| Expandable | Discoveries, changes, attempts, decisions, evidence summaries |
| Technical detail | Trace IDs, spans, graph edges, capture mechanisms, normalization predicates |
| Restricted | Secrets, private raw captures, prompts, correlation keys and unreviewed exports |

Plain-language explanations must be grounded in persisted facts. Hypotheses must remain labeled as hypotheses.

If event persistence fails, the system should stop admission at safe boundaries and display degraded observability. “Always shows what it is doing” cannot mean guaranteed live display during disk failure or total process failure.

## 14\. Learning / Self-Improvement Risks

The current promotion policy is conservative, but its evidence can still be contaminated.

### Independence is undefined

Three fixture contexts are not independent if they were produced by the same generator, share the same mistaken comparator, or omit the same branch.

Promotion needs provenance-based independence and held-out evaluation, not just a count of context IDs.

### Success metrics can reward degraded research

A tool that captures fewer dimensions can reduce apparent mismatches. A skill that proposes broader normalization can increase pass rates.

Evaluate:

- Required evidence retained.
- New UNKNOWNs.
- Denominator changes.
- Negative mutations detected.
- Security/effect coverage.
- Human correction frequency.
- Rollbacks and manual reversions.
- Applicability boundaries.

### Retraction must be transitive

A bad pattern must identify:

- Tools/skills derived from it.
- Projects and experiments that consumed it.
- Results whose validity depended on it.
- Successor revisions and replacement guidance.

“Deprecated” without consumer impact analysis leaves old projects trusting invalid knowledge.

### Knowledge is advisory

Global knowledge may recommend a transformation. It must not establish current compatibility, choose a target version or bypass verification.

Pin consumed revisions. Preserve reproducibility when registry defaults evolve.

Evidence: promotion thresholds (docs/07-learning/LEARNING-SYSTEM.md:45), knowledge scope (docs/07-learning/KNOWLEDGE-MODEL.md:1).

## 15\. Multi-Project / Resume Risks

### Mission uniqueness and resource ownership

Per-mission fencing is necessary but insufficient.

Distinct missions can represent identical work or touch one shared component. Define:

- Work identity/deduplication.
- Resource ownership.
- Shared-file publication ordering.
- Lease transfer and checkpoint recovery.
- Conflict granularity.
- Completion cardinality.

Default serial execution reduces exposure but must be enforced, not assumed.

### Resume months later

A resumable project needs a durable dependency manifest containing:

- Source snapshot.
- Engine and schema versions.
- Adapters and probes.
- Runtime/image and package bytes.
- Fixture and policy identities.
- Tools and skills.
- Knowledge revisions.
- Secret-reference availability.
- Approved scope and current subjects.

A content hash proves identity, not continued availability.

### Cross-project isolation

Artifact ownership checks must apply to reads, writes, nested references, exports, global promotion and tool outputs.

Global knowledge must not include customer identifiers or source fragments merely because those fragments look reusable.

### Source and target changes

Maintain separate concepts:

- Historical work performed.
- Retained implementation.
- Reusable source characterization.
- Current verified work.

After PHP 8.3 → 8.4, retaining generated files does not mean their previous verification remains current. Selective reuse requires evidence of unchanged dependencies and revalidation under the successor plan.

## 16\. Security / Sandbox Risks

The container profile rejects several common mistakes: privileged mode, Docker socket mounts, proxy-only containment, production secrets and unsafe host fallback.

Remaining issues require explicit threat-model decisions:

1. **Container isolation is not a guarantee against hostile code exploiting the host kernel.** Define whether malicious customer applications are in scope or only untrusted application behavior.
2. **Source artifacts can already contain secrets.** ASTs, source snippets, diagnostics, screenshots and dependency archives may preserve them.
3. **Byte-preserving source identity and sanitized evidence have different purposes.** Define private source storage/access separately from agent-readable artifacts.
4. **Package acquisition is an attack surface.** Archive traversal, install scripts, plugins and native builds require isolated acquisition and verified outputs.
5. **Egress denial must be tested.** Include IP literals, IPv6, redirects, alternate protocols, DNS and programs ignoring proxy settings.
6. **Path checks must survive races and platform differences.** Symlink swaps, hardlinks, Windows aliases, case sensitivity and source/target overlap require native tests.
7. **HMAC correlation depends on key custody.** Key loss can prevent comparison; export must not leak keys or provide a low-entropy secret oracle.
8. **Source text is untrusted input.** Terminal escapes, HTML and agent prompt injection must remain data throughout presentation and orchestration.

The system must refuse irreversible replay when no safe simulator exists. User approval cannot make a real payment replay reversible.

Evidence: sandbox (docs/09-security/SANDBOX.md:3), derived-data redaction (docs/09-security/SECRET-HANDLING.md:13).

## 17\. Vendor-Neutrality Risks

The proposed core is vendor-neutral. Actual integrations are **UNKNOWN**.

| Client | Evidence available here | Required verification |
|---|---|---|
| Claude Code | No executed integration | CLI/MCP lifecycle, approvals, interruption and resume |
| Cursor | No executed integration | Same, including schema/output handling |
| OpenAI Codex CLI | No executed integration | Same, including local process and permission behavior |
| Aider | No executed integration | Verify the chosen CLI orchestration path; do not assume MCP/Skill behavior |
| Hermes | No executed integration | Verify the chosen integration and restart semantics |
| Custom headless agent | No executed integration | Complete machine-readable CLI flow and deterministic error handling |

This does not claim the clients lack those capabilities. It means Rewrotron interoperability has not been demonstrated.

The CLI fallback must cover the complete workflow, including user decisions and review responses. Today, “complete fallback” is weakened by the missing explicit dependency/recommendation decision flow.

The planned Skill’s description is CI3→Laravel-specific. Keep that as a reference capability description, while the product-level operating contract remains source/target-neutral.

## 18\. Open-Source / Distribution Risks

The repository has an Apache-2.0 license. That does not settle:

- Parser and runtime dependency notices.
- Redis/container image distribution terms for selected versions.
- Proprietary SDK redistribution.
- Fixture licenses.
- Adapter/template provenance.
- Generated-code attribution.
- Customer-source rights.
- Global knowledge derived from licensed code.
- Release signing and security contact ownership.

No dependency lock or actual distributed bundle exists yet, so release compliance remains **UNKNOWN** pending the planned SBOM/license review.

“Local-first” must remain distinct from “data never leaves the machine.” An external agent can transmit source context independently of the engine. Document the supported data boundary and optional provider behavior without implying engine control over every client.

“Zero setup” is not realistic for arbitrary legacy execution. A portable Go binary does not supply PHP 7.4 extensions, container support, databases, fixtures, credentials and proprietary libraries.

Use precise distribution claims:

- Single-binary metadata operations where supported.
- Doctor-checked runtime prerequisites.
- Cached/offline replay after acquisition.
- Published tested OS/architecture profiles.

Native Windows/macOS process, filesystem and container tests are required; cross-compilation is insufficient. The release specification already acknowledges this. Distribution requirements (docs/10-development/RELEASE.md:9).

## 19\. Missing Specifications

### Missing Decisions

| Decision | Why it matters | Possible options | Recommended option | Consequences | Document |
|---|---|---|---|---|---|
| D01: Approval hash preimage | Approval must survive restart/upgrade | Implicit reconstruction; explicit preimage | Versioned explicit preimage | More fixtures; reproducible approval subjects | CONFIRMATION-PROTOCOL, plan schema |
| D02: Unit acceptance scope | Prevents dependency deadlock | Whole corpus; unit-only; unit plus implemented regressions | Unit plus applicable implemented regressions; whole corpus at project completion | Explicit scope identities and invalidation | DIFFERENTIAL-TESTING, MIGRATION-UNITS |
| D03: Executable business-rule language | Generation cannot interpret arbitrary prose deterministically | General DSL; bounded operators; external patches | Small typed operator set plus reviewed patches for unsupported logic | Honest capability limits | BEHAVIOR-MODEL, IR-CONTRACTS |
| D04: Workflow state semantics | Branches, actors and async completion need representation | Linear enumerated cases; state machine | Bounded linear scenarios initially, explicit state/actor/effect boundaries | Some applications remain unsupported | Workflow/scenario schemas |
| D05: Completeness attestation | Prevents false absence-of-effects claims | Probe assertions; mechanism-based evidence | Mechanism/profile attestations plus escape-path tests | More UNKNOWNs where closure is absent | RUNTIME-OBSERVATION |
| D06: Process ownership | Durable work requires an owner | Foreground-only; detached supervisor; daemon | Explicit engine-owned supervisor for background claims | Lifecycle, cleanup and attachment design | SYSTEM-DESIGN, RECOVERY |
| D07: Human decision input | Users must choose concrete alternatives | Edit generated YAML; import decisions; CLI/MCP decision API | Typed decision proposals/import plus human approval | New reviewed machine contracts | CLI-SPEC, MCP-SPEC |
| D08: Frozen decision versus progress | Avoids approval self-invalidation | One changing record; separate records | Immutable decisions, separate execution projection | Clear status semantics | DEPENDENCY-RESOLUTION, DOMAIN-MODEL |
| D09: Dependency/invalidation graph | Selective reuse must be justified | Generic links; typed graph; full invalidation | Typed edges with conservative fallback | Storage/query contract and reuse certificates | DOMAIN-MODEL, SQLITE-SCHEMA |
| D10: Historical projection | Watermark promises require history | Event sourcing; snapshots; current state only | Explicit historical projection strategy | More storage or reduced historical API promise | EXECUTION-OBSERVABILITY |
| D11: Source snapshot | Scanning live files can produce mixed input | Quiescence; filesystem snapshot; immutable copy | Coherent immutable source snapshot with drift checks | Acquisition and update/rebase protocol | ANALYSIS-SPEC, CONFIRMATION-PROTOCOL |
| D12: Observation boundary | Browser, workers and services change equivalence scope | Server-only; browser-inclusive; capability profiles | Explicit observation profiles | Unsupported client/stream behavior remains visible | PRODUCT, BEHAVIOR-MODEL |
| D13: Private source versus sanitized evidence | Source and ASTs can contain secrets | Refuse; private vault; transformed copies | Explicit private-source policy and sanitized derived artifacts | Storage/access and replay constraints | SECURITY, ARTIFACT-STORE |
| D14: Storage lifecycle | Years of trace retention can exhaust disk | Retain all; quotas; archives | Quotas plus verified archives and retention classes | Pauses/archives instead of silent deletion | SQLITE-SCHEMA, ARTIFACT-STORE |
| D15: Performance claim | Equal performance may stay inconclusive | Improvement; non-inferiority; approved budget | Explicit claim and customer workload/budget | No automatic relaxed acceptance | PERFORMANCE |
| D16: Learning independence/retraction | Counts alone can promote correlated mistakes | Context count; independent holdouts; causal provenance | Holdouts, mutation tests and consumer graph | More expensive promotion; safer reuse | LEARNING-SYSTEM |
| D17: Normalization approval identity | `approval_ref` lifecycle is unclear before Gate 2 | Existing confirmation; proposed decision; preallocated ID | Immutable decision identity approved transitively | Avoid circular/ambiguous references | COMPATIBILITY-SPEC, confirmation schema |
| D18: Resource claims | Mission IDs do not uniquely own work/files | Serial-only; resource locks; optimistic conflicts | Enforced serial baseline plus explicit resource ownership | Parallelism remains gated | MISSION-PROTOCOL, RECOVERY |

### UNKNOWN register

| Unknown | Why it matters | Evidence needed | How to obtain it |
|---|---|---|---|
| Real parser coverage | Partial parsing affects reachability | Pinned backend and adversarial source corpus | Execute parser/IR fixtures |
| Probe completeness | Missing effects can invalidate parity | Mechanism tests and control traces | Runtime mutation fixtures |
| Legacy environment viability | Source may not boot safely | Image, extension, startup and fixture evidence | Doctor plus isolated boot |
| Target lowering capability | Ordinary rules may be unsupported | Typed operator coverage and generated builds | Representative generation fixtures |
| Session interoperability | Default continuity may be impossible | Safe format/key/driver evidence | Tampered/expired/cross-side session tests |
| Client portability | Catalogs are not integrations | Pinned sessions and transcripts | Execute client matrix |
| Embedded SQLite version | WAL safety depends on actual build | Version and fix provenance | Pin dependency; inspect embedded engine |
| Customer business boundary | Repository cannot reveal every operation | Operational manifest, critical workflows and invariants | Customer review and safe characterization |

### Assumption Elimination Pass

**Every unexecuted runtime capability remains ASSUMPTION.** The following groups trace the principal requirements and invariants to measurable prerequisites. Specification acceptance alone does not discharge them.

| ASSUMPTION / claim | What must be true | Required conversion |
|---|---|---|
| “Detected framework X” | Evidence distinguishes X from modified/mixed installations | Detection tests, evidence and explicit correction |
| “Runtime version is X” | Measured sandbox runtime agrees with assertion | Doctor attestation; retain provenance |
| “Baseline is immutable” | All execution reads a coherent captured source | Snapshot mechanism and mutation tests |
| “Dead code” | Roots and relevant graph are closed | Machine proof bundle plus boundary declaration |
| “Same behavior” | Scope, states, dimensions and policies are defined | Approved contracts and differential tests |
| “Equivalent rule” | Operator semantics and types are supported | Typed capability validation |
| “Workflow covered” | Branch/state/actor partitions are represented | Scenario inventory and boundary cases |
| “No side effects” | Capture or containment covers all effect mechanisms | Completeness attestation |
| “Safe replay” | Every effect is isolated or simulated | Egress/reset/sink tests; otherwise limitation |
| “Fixture restored” | DB, sequences, files, sessions, caches and queues match | Restoration hashes and injected-failure tests |
| “Instrumentation does not change behavior” | Tested controls match within defined scope | Control fixtures plus residual limitation |
| “Dynamic value is harmless” | Format, relations and lifecycle are preserved | Approved rule plus negative mutations |
| “IDs correspond” | Mapping is bijective and consistent across sinks | Cross-dimension mapping tests |
| “Authentication preserved” | Hash verification, sessions, roles and revocation match | Auth/session adversarial corpus |
| “Dependency compatible” | Platform, used API and behavior are separately supported | Graph checks and characterization |
| “Target is pinned” | Exact versions and bytes satisfy requested constraints | Resolved schema, locks and acquisition records |
| “User approved” | Current exact subject has valid persisted authorization | Gate tests and explicit local trust limitation |
| “Unit complete” | Its acceptance scope passes with current evidence | Scoped completion rule |
| “Unrelated work remains valid” | Dependency closure proves unchanged prerequisites | Reuse certificate; conservative invalidation otherwise |
| “Agent cannot bypass” | Calls are mediated; account permissions match threat model | Interface tests and OS-boundary decision |
| “Resume works” | State, bytes, versions, keys and resources are recoverable | Crash/restore/upgrade fixtures |
| “One worker owns work” | Deduplication, claims and resource fences hold atomically | Race tests and ownership invariant |
| “Background work survives disconnect” | A durable owner remains alive | Process-kill/reconnect test |
| “Status is current” | Projection and watermark are consistent | Interleaved-write projection tests |
| “Progress is meaningful” | Basis and completion criteria exist; retries deduplicate | Accounting invariant tests |
| “Privacy preserved” | All persisted/exported paths sanitize or restrict sensitive data | Secret canaries and export review |
| “SQLite is sufficient” | Filesystem, version, contention and volume fit profile | Version checks and soak tests |
| “Backup is usable” | DB and referenced bytes form one recoverable set | Restore-and-reference verification |
| “Performance improved” | Workload is equivalent and measurement supports the claim | Paired tests and explicit statistical policy |
| “Knowledge is reusable” | Evidence is independent and applicability bounded | Holdout evaluation and human promotion |
| “Stable tool/skill” | Content, permissions, quality and rollback metrics are verified | Sandboxed benchmark and versioned activation |
| “Vendor-neutral” | Complete fallback works across actual client paths | End-to-end CLI and client tests |
| “Offline reproducible” | Required executable/package/image bytes remain available | Disconnected replay |
| “Open-source distributable” | All bundled material has acceptable provenance/terms | SBOM, notices and fixture audit |
| “Rollback is safe” | Traffic, data, sessions and external effects have a rehearsed strategy | Staging rehearsal or explicit release blocker |

## 20\. True Implementation Blockers

### P0 — Blocks safe operation of the affected subsystem

- **Gate 2 approval:** unresolved canonical plan hash construction.
- **Incremental migration:** whole-corpus unit completion deadlock.
- **Customer parity acceptance:** undefined evidence standard for declaring required capture dimensions complete.
- **Irreversible replay:** absence of a verified safe sink/simulator.
- **Sensitive export:** no demonstrated protection across source-derived artifacts.
- **Multi-writer release:** no pinned SQLite build containing the relevant WAL fix.

These do **not** prohibit writing safe foundational code or synthetic tests.

### P1 — Major decisions before dependent implementation

- Executable rule and workflow semantics.
- Durable process ownership.
- Human recommendation/dependency decision flow.
- Frozen decision versus operational status.
- Typed invalidation and reuse.
- Historical execution projection strategy.
- Source snapshot acquisition.
- Multi-connection and asynchronous behavior boundaries.
- Session continuity strategy.
- Mission deduplication/resource ownership.
- Storage and backup lifecycle.
- Performance acceptance claim.

### P2 — Can defer with explicit capability boundaries

- WebSockets and streaming.
- General concurrent workflow verification.
- Additional databases.
- Broader source/target languages.
- Advanced binary comparators.
- Global knowledge and tool synthesis.
- High-volume archive/index optimizations.

### P3 — Later improvements

- Desktop presentation.
- Embedding-based tool discovery.
- Vector search.
- Automated infrastructure cutover.
- Broad automatic architecture optimization.

Go package names, formatter setup and ordinary constructor details are not architecture blockers.

## 21\. Recommended Architecture Changes

### KEEP

- Three IR tiers.
- Behavioral migration units.
- Deterministic service authority.
- CLI baseline and optional MCP/Skills.
- SQLite metadata plus filesystem artifacts.
- Exact-subject approvals.
- Independent replay environments.
- PASS/MISMATCH/UNKNOWN.
- Conservative omission and normalization.
- Optional reviewed learning.

### CHANGE

- Replace implicit approval hashing with a versioned canonical subject contract.
- Separate requested intent, exact resolved environment and execution results.
- Separate unit acceptance from whole-project acceptance.
- Define bounded executable rules and scenario state semantics.
- Replace generic invalidation assumptions with typed dependency predicates.
- Align constitution, lifecycle and protocol baseline.
- Make completeness mechanism-backed.

### REMOVE

- Unsupported “tested” claims.
- Absolute equivalence wording without a bounded proof method.
- Whole-corpus completion requirements for prerequisites of ungenerated units.
- Any implication that a same-account agent is physically unable to edit local files.
- Any inference that a passing generated corpus validates its own completeness.

### ADD

- A lightweight **Target Implementation Model** between Behavior IR and generation: resolved environment, component ownership, lowering choices, behavior-to-implementation mappings and build requirements.
- A canonical trace/causality envelope.
- A process-owner/supervisor contract.
- Human decision submission.
- Resource ownership and deduplication.
- Historical projection strategy.
- Evidence quality/mutation evaluation.
- Storage quotas and coordinated backup manifests.
- Knowledge/tool consumer dependency records.

Recommended flow:

```
Immutable source + environment/operational profiles
    → Source IR
    → Application IR
    → Behavior IR + evidence/uncertainty
    → executable contracts and bounded scenarios
    → requested target + exact feasibility/resolution
    → approved plan subject
    → Target Implementation Model
    → staged implementation
    → isolated differential verification
    → scoped acceptance and repair
    → performance evidence
    → release package
```

The additional target model is not another universal source translator. It makes target choices and shared implementation ownership inspectable.

### DEFER

Broad language support, autonomous global synthesis and desktop work until the reference path demonstrates that the oracle detects deliberately wrong implementations.

## 22\. Required Documentation Changes

These are proposed changes only; no documents were edited.

| Document | Section | Current problem | Required change | Reason |
|---|---|---|---|---|
| `AGENTS.md` | Identity, adapters, tooling | Absolute proof, tested-adapter and transport conflicts | Bounded claims, planned maturity, pinned transport; scoped unit acceptance | Supreme authority must be truthful and consistent |
| `PRODUCT.md` | Promise and scope | Observation boundary insufficiently operational | Define supported observer profiles and claim terminology | Prevent server-only evidence becoming whole-app claims |
| `SYSTEM-DESIGN.md` | Ports, lifecycle, execution | Gate 3 conflict; process owner and evidence resolution unclear | Canonical lifecycle; supervisor and typed input/lookup contracts | Adjacent implementations need compatible ownership |
| `DOMAIN-MODEL.md` | Entities/cardinality | Decisions, progress and resource ownership mixed | Separate immutable decisions, execution states and resource claims | Avoid approval loops and duplicate work |
| `BEHAVIOR-MODEL.md` | Rules/workflows | Descriptive semantics too open | Typed predicates, effects, state boundaries and unsupported rules | Deterministic lowering requires executable meaning |
| `IR-CONTRACTS.md` | Revisions and references | Unsupported source semantics and lowering requirements unclear | Preservation obligations, semantic facets and capability negotiation | Prevent lossy normalization |
| `ANALYSIS-SPEC.md` | Inputs/publication | Live scan consistency and root closure underspecified | Coherent snapshot protocol and operational-boundary record | Protect baseline and omission validity |
| `RUNTIME-OBSERVATION.md` | Trace/completeness | Correlation and COMPLETE evidence gaps | Canonical causal envelope and mechanism attestations | Prevent missed effects becoming PASS |
| `MIGRATION-ENGINE.md` | Planning/completion | Scope deadlock and lowering gap | Scoped acceptance and target implementation model | Permit verifiable incremental progress |
| `MIGRATION-UNITS.md` | Dependencies | Acceptance scope not explicit | Case/regression scope per unit and shared ownership | Make DAG executable |
| `TARGET-SPECIFICATION.md` | Resolution/decisions | Exact resolution shares intent schema | Dedicated resolved record and decision-input contract | Preserve user choice and reproducibility |
| `DEPENDENCY-RESOLUTION.md` | Lifecycle | Decision changes can mix with progress | Frozen strategy records, separate progress and typed graph/API evidence | Avoid reapproval loops |
| `COMPATIBILITY-SPEC.md` | Policies/normalization | Approval-ref lifecycle and predicate semantics unclear | Typed normalization validators and decision identity | Keep exceptions reproducible |
| `DIFFERENTIAL-TESTING.md` | Completion/concurrency | Full-corpus unit deadlock; async closure unclear | Unit/project scopes and supported causal schedules | Avoid impossible acceptance |
| `DATABASE-COMPARISON.md` | Capture/reset | Connection-local and multi-DB state incomplete | Qualified identities, server/session effects and reset profiles | Same schema does not ensure same behavior |
| `AGENT-MODEL.md` | Authority | Mediated policy can be read as OS containment | Explicit local trust tiers and permitted proposal evidence | Avoid false safety guarantees |
| `CLI-SPEC.md` | Decisions/background work | Complete decision fallback absent | Human decision/import and process attachment semantics | Make CLI a complete operational fallback |
| `MCP-SPEC.md` | Schema/results/lifecycle | Client portability and disconnect behavior unproven | Self-contained schema delivery and process lifecycle tests | Avoid client-specific dependencies |
| `EXECUTION-OBSERVABILITY.md` | Watermarks/control | Historical state reconstruction undefined | Projection strategy and failure/ownership semantics | Keep status truthful |
| `LEARNING-SYSTEM.md` | Promotion/retraction | Evidence independence and invalidation unclear | Holdouts, mutation tests, consumer graph and retraction | Prevent contaminated global knowledge |
| `SQLITE-SCHEMA.md` | Ownership/concurrency | Some ownership and uniqueness rely on future services | Central semantic invariants, work dedup and fixed SQLite requirement | Make safety auditable |
| `ARTIFACT-STORE.md` | Retention/backup | Unbounded historical retention and private-source ambiguity | Retention classes, archive manifests, quotas and private-source policy | Prevent storage/privacy failure |
| `SECURITY.md` | Threat model | Hostile code/agent scope insufficiently concrete | Trust profiles and capability-specific containment claims | Match guarantees to actual permissions |
| `TESTING.md` | Acceptance | Shape checks cannot validate oracle quality | Mutation, semantic, process, concurrency and archive tests | Prove failures are detected |
| `PERFORMANCE.md` | Acceptance | Statistical non-regression claim ambiguous at equality | Defined claim, sample adequacy and budget semantics | Prevent indefinite inconclusive readiness |
| `ROADMAP.md` | Order | Semantic/oracle risk discovered relatively late | Bring falsifiable behavior experiment forward | Test the hardest premise early |

Schema changes must follow the repository’s versioning/migration rules when later authorized.

## 23\. Recommended New ADRs

1. **Canonical approval subjects and hash preimages** — serialized projections, adapter identities and upgrade reconstruction.
2. **Unit versus project acceptance scope** — dependency readiness and current regression evidence.
3. **Bounded executable behavior semantics** — operator vocabulary, unsupported logic and external patch path.
4. **Observation completeness attestation** — mechanisms, boundaries and residual uncertainty.
5. **Engine process ownership** — foreground/background execution, reconnect and cleanup.
6. **Immutable decisions versus execution progress** — dependency/recommendation lifecycle.
7. **Typed invalidation and evidence reuse** — conservative fallback and successor-plan certificates.
8. **Historical execution projections** — watermark consistency and pagination.
9. **Source snapshot and operational boundary** — acquisition, drift and root completeness.
10. **Resource ownership and mission deduplication** — shared files and parallel work.
11. **Private source, sanitized evidence and export boundaries**.
12. **Retention, archive and consistent backup lifecycle**.
13. **Knowledge independence, retraction and reproducibility**.
14. **Statistical performance acceptance**.

Each ADR should contain an executable acceptance experiment and a rejected alternative, not merely another architectural diagram.

## 24\. Recommended New Tests / Fixtures

Prioritize falsification.

| Fixture/test | Required failure detection |
|---|---|
| Login → checkout dependency DAG | Full-corpus unit deadlock cannot occur |
| Hash reconstruction after registry upgrade | Old approval reproduces only with original pinned subjects |
| Unknown declarative operator | Generation refuses explicitly |
| Branching checkout with three outcomes | Every branch represented or visibly unsupported |
| PDO bypassing framework driver | DB completeness becomes incomplete |
| Background email after HTTP completion | Effect waits/correlation or UNKNOWN |
| Dynamic DB-selected controller | No false DEAD |
| Forgotten cron-only helper | Operational-boundary gap prevents omission |
| Modified vendor framework | Version-only support claim rejected |
| Source mutation/ABA/symlink swap | No mixed approved snapshot |
| Empty/repeated/raw query parameters | Lossless request reconstruction |
| Large integer/decimal through CLI and MCP | No precision loss |
| Two DBs with identical table names | Effects remain connection-qualified |
| DDL, temporary table and sequence rollback | Reset verified or execution refused |
| Redis session/queue/lock contamination | Restoration failure detected |
| Source session tamper/expiry/logout/rollback | Continuity cannot pass on response status alone |
| CSRF equality-preserving wrong implementation | Replay and cross-user defects detected |
| Swapped generated IDs across response/email/DB | Mapping mismatch detected |
| Duplicate queue/payment delivery | Effect count and idempotency defect detected |
| PDF/image replacement | Comparator accepts only approved properties |
| Two duplicate missions | Only one logical work owner admitted |
| Two patches touching shared file | Ownership/conflict enforcement |
| Expired lease publishing after reassignment | Stale fence rejected |
| Disjoint concurrent publications | Defined revision semantics, no unexplained starvation |
| MCP disconnect/process kill | Work ownership and cleanup match documented policy |
| Old watermark during new writes | Historical/current state distinction remains correct |
| Event-store failure after sink effect | No duplicate side effect during recovery |
| Disk quota exhausted mid-capture | Safe stop with preserved actionable evidence |
| DB/artifact backup during publication | Restored references all resolve |
| Network filesystem workspace | Unsupported WAL profile rejected |
| Embedded SQLite version test | Build contains required fix |
| False global pattern from shared oracle | Promotion fails independent evaluation |
| Experimental tool dropping dimensions | Quality regression blocks activation |
| Secret canaries in source/AST/stderr/DOM/events | No unauthorized persistence/export |
| Django/Go, Laravel/Rust, PHP/Python IR fixtures | No silent framework-specific canonical semantics |
| Identical noisy performance distributions | INCONCLUSIVE rather than false regression/non-regression |
| Rollback after new target writes | Explicit reconciliation or readiness blocker |

No runtime fixture currently demonstrates these properties.

## 25\. First-Milestone Reassessment

The current M1 safe-foundation milestone is reasonable, but it does not test the product’s hardest premise. Bring a small **falsifiable behavior reconstruction experiment** immediately behind it.

The first meaningful demonstration should prove:

1. Durable source approval and reproducible hash subjects.
2. A coherent source snapshot.
3. A bounded multistep scenario with state and an authorization failure.
4. Independent resettable source/target environments.
5. A small supported executable rule vocabulary.
6. Approved target resolution and scoped unit acceptance.
7. Detection of deliberately missing DB writes, wrong principals, duplicated effects and incomplete capture.
8. Recovery after publication and process interruption.
9. Honest event/state reporting.
10. Repeatability without an LLM.

A read-only JSON route remains a useful smoke test. It is insufficient evidence for the reconstruction architecture.

### Simulation: first realistic CI3 migration

Source: PHP 7.4, CI3, MySQL 8, Redis, custom auth, 50+ routes, HTML/API, cron, payments, uploads, custom helpers and old packages. Target: PHP 8.3/Laravel, preserved MySQL/Redis semantics.

| Stage | What goes wrong | Required handling |
|---|---|---|
| Detect | Missing extension, patched CI3, ambiguous runtime, production-only configuration | Separate measured facts, assertions and unknowns; inventory modifications |
| Confirm | User correction disagrees with executable environment; scan is stale | Review exact merged subject; doctor mismatch remains visible |
| Analyze | `_remap`, dynamic loaders, cron and custom auth evade conventional extraction | Preserve unresolved candidates and operational roots |
| Observe | App cannot boot; source reaches production endpoints; Redis reset incomplete | Doctor refusal, closed sinks and verified reset |
| Discover | Crawler sees links but not admin, failed-payment or monthly workflows | Customer critical scenarios and role/state partitions |
| Model | Payment/retry branches and post-response effects do not fit linear contracts | Explicit supported state/effect semantics or blockers |
| Plan | Old SDK has no PHP 8.3 release; target resolution remains wildcard; login dependency deadlocks | Reviewed strategy, exact resolution and scoped acceptance |
| Migrate | Slices share auth/helper files; target defaults alter CSRF/session behavior | Component ownership and explicit compatibility lowering |
| Verify | IDs, sequences, time, multipart uploads and async emails diverge | Relational mapping, raw inputs, clocks and completion barriers |
| Repair | Agent changes comparator or endlessly resets context with trivial patches | Policy changes require approval; incident-level escalation budget |
| Optimize | Caching changes authorization freshness or transaction behavior | Reverify state/security under measured workload |
| Finalize | Session bridge unproved; rollback cannot undo target writes; benchmarks inconclusive | Block readiness; disclose exclusions and require rehearsed strategy |

This application cannot honestly be declared whole-app equivalent if cron, payment verification or active-session continuity remains unsupported. A scoped migration is valid only when its exclusions remain explicit.

### Framework-neutrality simulations

These migrations are not currently supported. The exercise tests whether the architecture represents their problems without silently converting them into CI3/Laravel concepts.

| Stage | Django → Go | Laravel → Rust | Custom PHP → Python |
|---|---|---|---|
| Detect | Multiple settings profiles, Python packages and ASGI/WSGI | Providers, packages, runtime and queue drivers | No framework identity; includes, globals and custom dispatch |
| Confirm | Distinguish measured settings from supplied deployment facts | Confirm actual runtime/worker topology | Confirm operational roots without inventing conventions |
| Analyze | Signals, middleware, ORM hooks and management commands | Container bindings, model events, facades and scheduled jobs | Dynamic includes, autoloaders, eval and raw SQL |
| Observe | Transactions, signals and workers need correlated capture | Jobs/listeners may outlive requests | No standard hooks; black-box capture has explicit gaps |
| Discover | Admin workflows and permission branches | Artisan, queues and webhook flows | Customer-supplied workflows dominate |
| Model | Query laziness, Decimal, signals and auth semantics | Serialization, policies, events and transaction behavior | PHP coercion, global state and custom sessions |
| Plan | Go provider for each required capability; no package-name equivalence | Rust provider/bridge decisions; async semantics explicit | Python providers preserving used behavior |
| Migrate | HTTP handlers/services plus declared transaction ownership | Async handlers/services with explicit ownership | Custom handlers/services and compatibility helpers |
| Verify | ORM query semantics, timezone and permission differences | Queue ordering, numeric values and serializer differences | Missing/null, numeric strings, encodings and ordering |
|  |  |  |  |
|---|---|---|---|
| Repair | Fix semantic lowering rather than mimic Django classes | Diagnose effects/lifetimes rather than recreate facades | Fix PHP/Python semantic differences with evidence |
| Optimize | Pooling/concurrency must not change transactions | Concurrency must not change delivery/order | Caching/vectorization must not alter coercion or effects |
| Finalize | Auth/session and worker rollback evidence | Queue/session/state rollback evidence | Custom-session and operational-command evidence |

The same Behavior IR can support these only if it preserves the relevant semantics and can declare unsupported behavior. Reusing the same schema name is not evidence of neutrality.

## 26\. Final Go / No-Go Assessment

**Safe to start now:**

- Go/module/bootstrap work.
- Storage and artifact foundations.
- Canonical hash vectors.
- Bounded detection.
- Gate services.
- Machine-readable CLI contracts.
- Execution events.
- Synthetic refusal, recovery and semantic fixtures.

**Resolve before the corresponding operational work:**

- Approval hash preimages.
- Unit acceptance scope.
- Executable rule/scenario semantics.
- Completeness attestation.
- Process ownership.
- Human dependency/recommendation decisions.
- Snapshot consistency and sensitive-evidence boundaries.
- Resource ownership and selective invalidation.

**Defer with explicit limits:**

- Broad source/target support.
- WebSockets/general asynchronous reconstruction.
- Global skill synthesis.
- Vector search.
- Desktop UI.
- Automated production cutover.

**Assumptions remaining:**

Parser coverage, probe completeness, runtime viability, generator capability, session continuity, performance, client interoperability and reproducibility are all unproven until their named fixtures execute.

The architecture is suitable for a constrained, evidence-driven migration system. It is **not yet specified tightly enough to support confident customer migration claims**. The next milestone should demonstrate that Rewrotron reliably rejects a plausible but wrong reconstruction—not merely that it can generate a plausible target application.
