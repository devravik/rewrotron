# Rewrotron Pre-Implementation Red-Team Review

## 1\. Executive Summary

**The architecture is substantially better than the earlier design, but the verifier’s execution contract is still not precise enough to implement safely.** The repository now correctly distinguishes bounded equivalence, assessment completion, user intent, exact target resolution, immutable decisions, scoped acceptance and historical evidence. Those corrections should be retained.

The highest-risk remaining problems are:

1. Executable behavior actions and verification assertions do not have clearly separated execution roles. A poorly designed interpreter could supply the database write that a defective target omitted.
2. Source-only observation requires an identity-map reference, while the defined identity map requires a verification case and paired source/target values.
3. Typed expression namespaces exist, but their observation-to-value projection is not fully specified.
4. The bounded response action cannot express headers, redirects or cookies directly.
5. Byte-exact source preservation conflicts with the prohibition on production secrets inside execution containers when credentials are hardcoded.
6. Source identity explicitly excludes filesystem metadata that can affect application behavior.
7. Stateful source/target branch divergence lacks an explicit pairing policy.
8. Selective invalidation needs evidence that prerequisite discovery itself was complete.

These are contract problems, not hypothetical bugs in existing Go code. **There is no Go engine, runtime fixture corpus or implemented adapter to inspect.**

Review baseline: commit `0ab4d9bede6f31dcde1cc0105fae6771ee80d39f`. No repository files were changed.

The existing structural validator passed:

- 133 schemas.
- 222 positive/negative fixtures.
- 71 CLI/MCP catalog entries.
- 608 local links.
- 40 SQL tables across two migrations.
- 10 approval vectors.

That establishes structural consistency within the validator’s scope. It does **not** establish runtime safety, observation completeness, correct reconstruction, client compatibility or reliable recovery. Testing contract (docs/10-development/TESTING.md)

### Document authority map

| Level | Current authority | Assessment |
|---|---|---|
| Constitution | AGENTS.md | Correctly states bounded claims, planned adapter maturity and deterministic authority |
| Product | PRODUCT.md | Owns product promises and supported boundaries |
| Architecture | SYSTEM-DESIGN.md and accepted ADRs | Own ports and material architectural decisions |
| Domain | Topic owners listed in docs/README.md | Own lifecycle, behavior, identities and invalidation |
| Machine contracts | specs/, CLI-SPEC.md, MCP-SPEC.md | Own public shapes; services must enforce cross-record meaning |
| Storage/comparison | SQLITE-SCHEMA.md, compatibility owners | Own persistence and comparison obligations |
| Tests | TESTING.md and structural fixtures | Runtime experiments are specified, not executed |
| Implementation | Documentation validator only | No engine implementation establishes additional behavior |

Constitution (AGENTS.md:3), owner index (docs/README.md), machine registry (specs/README.md)

**Duplicates:** architecture summaries, vocabulary indexes and roadmap indexes generally delegate to canonical owners. This is appropriate. Several ADRs, however, contain generic machine-impact text that does not identify their precise contract changes.

**Historical material:** the earlier red-team review and earlier entries in RED-TEAM-RESOLUTION are audit history, not current subsystem authority. Their unresolved problems must be checked against current specifications rather than repeated automatically.

**Contradictions still present:** source secrets versus execution admission; source-only capture versus paired identity mapping; “all mapped consumers” versus implemented-regression scope; overlapping ACTIVE/DORMANT definitions; CLI catalog version wording.

**Unspecified areas:** interpreter execution roles, canonical evaluation context, typed fixture/principal records, branch pairing, effect payload contracts, filesystem behavioral identity and prerequisite-coverage certification.

## 2\. Architecture Strengths

Several difficult decisions are already sound.

- **Evidence-bounded claims.** Observed, contract and semantic equivalence are distinguished. A generated passing corpus does not establish its own completeness.
- **Deterministic acceptance authority.** Agents propose interpretations and patches; the engine owns gates, results and state.
- **Separate intent and resolution.** Wildcards are allowed as user constraints but cannot masquerade as exact environments.
- **Immutable approval closure.** Approved subjects retain executable, schema, adapter, policy and decision identities.
- **Scoped unit acceptance.** Prerequisite units no longer require future dependent behavior to pass.
- **Independent replay environments.** Source and target cannot share writable verification databases, caches, sessions, files or effect endpoints.
- **Mechanism-derived completeness.** Imported `COMPLETE` assertions cannot establish capture sufficiency.
- **Conservative omission.** DEAD is bounded by an approved operational boundary and an evidence bundle.
- **Shared assessment checkpoint.** Assessment can succeed while migration remains blocked, without creating implementation authority.
- **Explicit process ownership.** Foreground and supervised execution have different, documented lifetimes.
- **Honest local trust boundary.** Same-account access is not represented as hostile-agent isolation.
- **Early falsification.** M1 must reject plausible wrong targets before broad adapter development.

These are meaningful architectural strengths. They are still **specified obligations**, not demonstrated product capabilities. Development plan (docs/11-roadmap/DEVELOPMENT-PLAN.md)

## 3\. Critical Risks

### CR-01 — The executable model could become its own oracle

The behavior vocabulary includes `db_write`, `emit`, `set_state`, transaction actions and assertions. Differential testing describes executing steps and then comparing observations, but does not establish exactly where each model action executes.

Concrete failure:

1. The contract declares an order INSERT.
2. The target returns a plausible response but performs no INSERT.
3. The interpreter executes the contract’s INSERT against the target fixture.
4. State comparison finds the expected row.
5. The missing-write mutant passes.

This is an **execution-role ambiguity**, not evidence that an implementation already behaves this way.

Required resolution: separate model simulation, fixture preparation, application execution and verification. Expected-model actions must never modify measured application state or emit measured application effects. A model simulator needs separate state and provenance. Verification assertions must consume independently captured observations.

Behavior actions (docs/02-domain/BEHAVIOR-MODEL.md:27), replay protocol (docs/05-verification/DIFFERENTIAL-TESTING.md:13)

**Priority: P0 before implementing the verifier.**

### CR-02 — Source observation has an unresolved forward dependency

Current runtime observation explicitly permits pre-contract traces with `case_id=null`. However:

- An effects document requires `identity_map_ref`.
- The defined identity map requires a non-null verification `case_id`.
- Each populated mapping requires both source and target values.

Source discovery occurs before the target exists.

Read-only schema probes confirmed that an effects document without the map is rejected and an identity map with `case_id=null` is rejected. The intended type of a source-only map is **UNKNOWN**; no separate source-only identity record is defined.

Required resolution: source capture records side-local identities. Pairwise identity maps are derived later from two observations and referenced by verification results, without modifying original capture bytes.

Pre-contract observation (docs/03-analysis/RUNTIME-OBSERVATION.md:34), effects requirement (specs/observation/effects.schema.json:182), identity-map requirements (specs/verification/identity-map.schema.json:164)

**Priority: P0 before runtime observation publication.**

### CR-03 — Typed expressions lack a complete evaluation-context contract

The language defines namespaces such as `input`, `actor`, `state`, `result`, `binding` and `row`. It defines types and pointer rules. It does not fully define the canonical records exposed under those namespaces.

Questions remain:

- Where are response headers and parsed bodies located?
- How are exact JSON numbers projected into typed values?
- Does `state` mean observed state, simulated state or fixture state?
- What independently establishes the authenticated principal?
- How are effect payloads and session values typed?
- Which values exist before, during and after an operation?

Two interpreters could implement different projections while validating the same contract.

Required resolution: specify a small, versioned evaluation context with phase, origin and type rules. Missing or unobservable values must remain unavailable rather than being supplied by the expected model.

Reference namespaces (docs/02-domain/BEHAVIOR-MODEL.md:23), state semantics (docs/02-domain/BEHAVIOR-MODEL.md:49)

**Priority: P1 before M1.**

### CR-04 — The first stateful slice exceeds the declared response vocabulary

`respond` supports status and body. It does not directly represent ordered headers, `Location`, `Set-Cookie` or cookie attributes.

A structural probe confirmed:

- A response action with status/body validates.
- Adding a `Location` header makes it invalid.

Login, redirects and session continuity are central to M1 and the reference migration. They need an explicit representation or a qualified provider/lowering contract. Undocumented framework defaults are not an acceptable substitute.

Response action (specs/behavior/semantics.schema.json:555), M1 scope (docs/11-roadmap/DEVELOPMENT-PLAN.md:27)

**Priority: P1 before M1.**

### CR-05 — Hardcoded production credentials break execution admission

SECURITY requires byte-exact private source, including hardcoded credentials. SECRET-HANDLING says snapshots are not transformed into sanitized execution source. SANDBOX prohibits production secrets inside the container.

For an application with a hardcoded payment key, these requirements cannot all hold simultaneously.

Required resolution: define explicit admission outcomes. Options include credential revocation, separately evidenced execution overlays or refusal of runtime execution. An overlay must have its own identity and declared behavioral limitations; it cannot be described as the original source.

Private source (docs/09-security/SECURITY.md:15), execution-source preservation (docs/09-security/SECRET-HANDLING.md:25), container prohibition (docs/09-security/SANDBOX.md:7)

**Priority: P1 before customer runtime execution.**

### CR-06 — Coherent source bytes do not establish coherent filesystem behavior

The snapshot records content, executable bits and symlink targets, then states that metadata is not identity.

Legacy applications can depend on:

- `filemtime()` for cache invalidation.
- Permissions for upload acceptance.
- Ownership or writable-directory checks.
- Hardlink relationships.
- Real paths and working directories.
- Case-sensitive filename lookup.

An unchanged byte tree can therefore produce changed behavior without source drift being detected.

Required resolution: capture relevant filesystem properties in an environment/filesystem profile, preserve them where supported and reject unsupported dependencies.

Snapshot identity (docs/03-analysis/ANALYSIS-SPEC.md:9)

**Priority: P1 before reference characterization.**

## 4\. Architecture Contradictions

### Contradictions

| Conflict | Concrete consequence | Precise resolution |
|---|---|---|
| Byte-exact hardcoded secrets versus no production secrets in containers | Exact source cannot be admitted under the stated container policy | Add an explicit runtime-admission decision and separately identified overlay/refusal path |
| Pre-contract source observation versus mandatory paired identity map | Capture requires an artifact whose prerequisites do not yet exist | Separate side-local identities from later pairwise correspondence |
| “Shared component changes reverify all mapped consumers” versus implemented-regression acceptance | A reader can reintroduce future-consumer deadlock | Require current implemented consumers now; mark future consumers pending checks when implemented |
| ACTIVE includes reachability evidence; DORMANT is reachable but unobserved | Classification and progress counts overlap | Define mutually exclusive evidence rules |
| CLI-SPEC header says catalog 2.0; catalogs and later text say 2.1 | Implementers can advertise the wrong contract version | Update the header to the catalog’s authoritative version |
| Generic exact environment versus numeric release-version regex | Exact date-tagged or vendor-specific releases cannot be represented | Separate opaque release identity from adapter-specific constraint interpretation |
| Mutable-looking detection `gate_status` versus immutable detected facts | A detection can continue reporting “awaiting” after confirmation, or require a replacement artifact | Move current gate state into its projection, or explicitly define this field as historical capture state |

Shared ownership wording (docs/04-migration/TARGET-ARCHITECTURE.md:13), acceptance owner (docs/04-migration/MIGRATION-UNITS.md:10), classification (docs/03-analysis/CODE-CLASSIFICATION.md:3), CLI header (docs/06-agent/CLI-SPEC.md:3), release-version shape (specs/project/resolved-environment.schema.json:32)

The following apparent conflicts have already been resolved and should not be reopened without new evidence:

- Modernization versus compatibility: intentional changes require explicit obligations.
- Autonomy versus confirmation: autonomous work operates inside approved subjects.
- Local-first versus external AI: storage defaults do not control an external client’s egress.
- Evolving knowledge versus reproducibility: consumption is revision-pinned.
- SQLite versus large traces: metadata and large artifacts are separated.
- Unit completion versus project completion: acceptance scopes differ.
- Assessment completion versus feasibility: they are independent.

## 5\. Practical Failure Scenarios

**D** denotes a current contract defect or ambiguity. **V** denotes a specified protection whose implementation remains unvalidated. These are concrete failure scenarios, not claims that an existing engine has exhibited them.

| ID | Scenario | Why Current Design Fails | Impact | Detection | Required Design Change | Priority |
|---|---|---|---|---|---|---|
| F01 | Target omits order INSERT | D: interpreter effect execution is not separated from measurement | False PASS | Missing-write mutant with checker writes audited | Separate model and measured effects | P0 |
| F02 | Source-only capture publishes effects | D: mandatory mapping precedes case/target availability | Discovery stalls or invents identity evidence | Publish before any target exists | Side-local identity record | P0 |
| F03 | Contract reads `/result/order_id` | D: canonical result layout is undefined | Different interpreters disagree | Projection compatibility vectors | Versioned evaluation context | P1 |
| F04 | Login returns a session cookie | D: response action lacks header/cookie representation | Hidden framework behavior or blocked slice | Header-action schema probe | Typed response/provider contract | P1 |
| F05 | Source embeds a live payment key | D: exact source and container-secret prohibition conflict | Unsafe admission or altered source claim | Secret-canary admission fixture | Explicit overlay/refusal protocol | P1 |
| F06 | Cache key uses file modification time | D: metadata excluded from identity | Undetected behavioral drift | Equal bytes, different mtime | Filesystem profile identity | P1 |
| F07 | Target follows success branch; source follows failure branch | D: cross-side branch pairing undefined | Invalid comparisons or hidden mismatch | Deliberately divergent branch | Logical step/terminal pairing policy | P1 |
| F08 | Target server still runs an older build | D: actual execution identity needs stronger witnesses | Wrong tree credited with PASS | Serve old build against claimed mutant tree | Bind observations to measured execution identity | P1 |
| F09 | Same-named table exists on two databases | V: qualification and connection checks not executed | Wrong state compared | Multi-DB mutation fixture | Implement qualified model/reset controls | P1 |
| F10 | Direct PDO bypasses framework hooks | V: probe coverage can be incomplete | Missing write accepted | PDO control beside framework driver | Independent DB capture closure | P1 |
| F11 | Cron-only helper has zero HTTP hits | V: operational inventory may be incomplete | Required code omitted | Scheduler-only fixture | Keep UNKNOWN until boundary closure | P1 |
| F12 | Route is loaded from database configuration | V: static inventory cannot enumerate unconstrained dispatch | Missing behavior | Change routing record under profile | Dynamic-path evidence and blockers | P1 |
| F13 | Framework vendor file is patched | V: labels do not establish installed semantics | Incorrect adapter lowering | Modified-vendor fixture | Content-qualified adapter admission | P1 |
| F14 | Production flag enables dormant admin code | V: fixture profile misses configuration | Incomplete assessment | Flag/profile partition tests | Explicit operational profiles | P1 |
| F15 | Replacement SDK retries POST twice | V: API compatibility does not establish effect semantics | Duplicate payment | Retry/timeout simulator | Capability characterization and cardinality | P1 |
| F16 | Proprietary extension cannot be acquired | V: feasible runtime cannot be constructed | Migration blocked | Doctor/acquisition evidence | Useful blocked assessment; no fallback | P1 |
| F17 | INSERT ID differs between source and target | V: mapping quality untested | Broken links normalized away | Wrong foreign-key mutant | Relation-preserving bijection | P1 |
| F18 | Same actor label actually uses another account’s session | D: fixture identity can be confused with observed principal | Authorization defect hidden | Swap session jar while preserving label | Principal provenance checks | P1 |
| F19 | Logout leaves target session valid | V: format compatibility insufficient | Security regression | Replay token after logout | Session lifecycle corpus | P1 |
| F20 | CSRF token is reused after rotation | V: token masking can conceal lifecycle error | False compatibility | Old-token replay mutant | Lifecycle-preserving normalization | P1 |
| F21 | Email arrives after HTTP success | V: response completion can precede effects | Premature PASS | Delayed worker delivery | Causal acknowledgement boundary | P1 |
| F22 | Worker retries after sink acceptance but before ack | V: exactly-once effects are not guaranteed | Duplicate delivery | Crash between acceptance and ack | Receipt reconciliation/idempotency | P1 |
| F23 | Delayed job is scheduled beyond the bounded timeout | D: finite quiescence cannot establish absence of future effects | False absence claim | Inspect pending scheduled work | Bound claim or report unsupported | P2 |
| F24 | Reset restores rows but leaves auto-increment state changed | V: full reset not demonstrated | Later cases diverge | Rollback/sequence fixture | Qualified reset or fresh instance | P1 |
| F25 | Trigger writes through another server connection | V: application probes miss server-side effects | Unsafe or incomplete replay | Trigger/routine escape control | Contain and observe service processes | P1 |
| F26 | Redis lock/TTL survives between scenarios | V: namespace/reset controls untested | Flaky or false comparison | Contamination fixture | Reset all stores and clocks | P1 |
| F27 | Agent omits one dependency edge | D: detecting graph incompleteness is underspecified | Stale implementation reused | Remove known consumption edge | Coverage witness and conservative invalidation | P1 |
| F28 | PHP target changes after most slices complete | V: typed invalidation not implemented | Old results remain current | Version-change acceptance test | Successor resolution/plan and scoped revalidation | P1 |
| F29 | Two agents claim identical work | V: SQL uniqueness exists; service race behavior untested | Duplicate generation | Concurrent claim test | Atomic claim plus ownership | P1 |
| F30 | Expired writer resumes after a new owner publishes | V: filesystem fencing requires OS ownership | Lost edits | Suspend old process across handoff | Ownership across check/write | P1 |
| F31 | Raw HTTP header contains non-UTF-8 bytes | D: string-only representation has no explicit byte policy | Lossless replay claim fails | Byte-preservation vectors | Bytes/encoding contract or refusal | P2 |
| F32 | Two identical order rows have no natural correspondence | V: mapping can be ambiguous | Wrong entity pairing | Duplicate indistinguishable fixtures | UNKNOWN unless evidence establishes mapping | P1 |
| F33 | Normalizer removes a refund deadline | V: variability does not mean irrelevance | Business defect hidden | Boundary-time mutant | Controlled-clock contract | P1 |
| F34 | Included file has top-level side effects | V: declaration-focused extraction can miss initialization | Behavior omitted | Include-time mutation fixture | Preserve file initialization or unsupported node | P1 |
| F35 | Status reader holds a long SQLite transaction | V: checkpoint discipline untested | WAL growth and stalls | Reader/checkpoint soak | Bounded reads and checkpoint controls | P1 |
| F36 | Disk fills while an effect completes | D: emergency persistence capacity is not established | Cleanup/status evidence lost | Exhaust DB/WAL/temp space | Reserved recovery capacity and reconciliation | P1 |
| F37 | Shared route file changes for an unimplemented consumer | D: “all consumers” wording conflicts with scoped acceptance | Scheduling deadlock | Shared-file dependency fixture | Implemented-consumer scope plus pending checks | P1 |
| F38 | Benchmark stops at the first favorable sample count | D: stopping rule not fixed | Inflated performance acceptance | Repeated-look simulation | Fixed sample count or sequential method | P2 |
| F39 | Learned tool drops difficult traces | V: apparent success rises as evidence quality falls | Contaminated knowledge | Capture-degradation holdout | Quality denominators and mutation tests | P2 |
| F40 | Retracted technique was consumed by an offline project | D: global consumer discovery protocol deferred | Stale knowledge remains eligible | Resume disconnected project after retraction | Consumer index and mandatory resume check | P2 |
| F41 | Source hash exposes a low-entropy credential file | V: sanitizing text alone does not sanitize identity metadata | Secret disclosure | Dictionary/canary export test | Private identity and sanitized projection | P1 |
| F42 | Rollback follows target-created orders and sessions | V: preserved schema does not make writes reversible | Lost or inconsistent production state | Write-aware rehearsal | Reconciliation before release claim | P1 |

## 6\. Source Analysis Risks

The current conservative analysis boundary is appropriate. The principal risk is implementing “partial understanding” as apparently complete graphs.

| Feature group | Static limitation | Dynamic limitation |
|---|---|---|
| Computed include/require, custom autoloaders, autoload manipulation | Possible targets can be unbounded | Observing one target does not exclude others |
| Variable classes/methods, magic methods, reflection | Dispatch needs runtime/configuration constraints | Concrete calls establish presence only |
| DI containers, service locators, global/static state | Bootstrap and environment determine bindings | Fresh requests miss persistent-worker state |
| Hooks, middleware, listeners, framework extensions | Registration and order may be conditional | Early exit, cache or shutdown paths bypass probes |
| Eval, generated PHP, runtime patching patterns | General semantics cannot be established safely | Execution needs containment and still leaves alternatives |
| Vendor changes, mixed versions, custom libraries | Package labels are insufficient | Controls must match actual modified bytes |
| Raw SQL, multiple connections, triggers, routines | Language graphs omit server-side logic | Driver-only traces miss indirect effects |
| Cron, CLI, queues, webhooks, admin routes | Deployment configuration establishes roots | Short observation windows miss rare schedules |
| Environment variables, flags, runtime configuration | One profile covers one configuration | Sanitized fixtures can miss production conditions |
| Templates, assets, uploads, image/PDF processing | Files can contain executable or include-time behavior | Binary semantics require supported comparators |

Analysis owner (docs/03-analysis/ANALYSIS-SPEC.md), reference adapter boundary (docs/04-migration/CI3-ADAPTER.md)

**Cannot safely be established statically:** unconstrained dynamic dispatch, eval-generated behavior, runtime container bindings, production-only configuration and external outcomes.

**Cannot safely be observed dynamically:** real irreversible effects without a faithful simulator, inaccessible production-only states, unavailable proprietary/native dependencies, or paths whose execution cannot be contained.

### Dead-code threshold

Retain the existing strong omission bundle:

- Closed operational roots.
- Complete relevant parsing.
- No reachable static path.
- No unresolved candidate path.
- Usable controlled corpus with zero hits.
- Current snapshot/analyzer/graph identities.
- Explicit inclusion in the approved plan.

Customer approval defines a boundary; it does not prove that an undocumented scheduler does not exist. Missing operational evidence must keep the affected code UNKNOWN.

The ACTIVE/DORMANT overlap needs correction, but **it does not justify weakening DEAD**. Omission owner (docs/03-analysis/CODE-CLASSIFICATION.md)

## 7\. Runtime Observation Risks

A reliable observation environment requires more than launching PHP in Docker.

The current contract correctly requires pinned runtimes, images, extensions, stores, sinks and reset controls. Remaining implementation risks include:

- Missing legacy extensions or incompatible native libraries.
- Incorrect document root, working directory or filesystem permissions.
- Environment-dependent TLS certificates, DNS names and redirects.
- Workers and cron inheriting different configuration from HTTP.
- Unowned subprocesses surviving a workflow.
- Server-side database effects escaping application probes.
- Instrumentation perturbing hook order, shutdown behavior or timing.
- Observation output being mistaken for trusted measurement merely because it came from a pinned probe.

**Probe identity is necessary but not sufficient.** Measurements must identify their capture boundary and distinguish application-supplied data from independently observed effects.

The causality envelope is a sound start. Async evidence additionally needs operation occurrences, delivery attempts and retry relationships. A job ID alone does not distinguish its first delivery from a retry.

The five-minute workflow completion bound is defensible for a narrow subset. It cannot establish that a scheduled notification six hours later never occurs. Such behavior needs a separately bounded schedule experiment or explicit unsupported scope.

Source-only identity publication remains the immediate machine-contract blocker described in CR-02. Observation owner (docs/03-analysis/RUNTIME-OBSERVATION.md)

## 8\. Behavior Model Risks

The closed vocabulary is preferable to arbitrary prose or executable source fragments. It still needs clearer roles.

### Required separations

| Concept | Required meaning |
|---|---|
| Fixture state | Prepared initial application state |
| Observed state | State measured from the application environment |
| Expected state | Independently justified contract/model result |
| Simulation state | Separate state used to interpret a model |
| Actual effect | Effect captured from source or target execution |
| Expected effect | Obligation against which actual effects are checked |

These must never be aliases to the same mutable store.

### Remaining semantic gaps

- Canonical evaluation-context shapes and projection rules.
- Phase-specific availability of values.
- Principal identity evidence.
- Header/cookie/redirect response behavior.
- Typed session, cache, queue and filesystem effects.
- Provider boundaries for computations outside the small operator set.
- Cross-side branching and terminal comparison.
- Typed initial-state and principal-fixture contracts.

The language deliberately lacks many real computations: multiplication, division, string transformations, password verification, template rendering and complex date rules. That is acceptable **only if the reference scope declares how each required computation is represented or blocked**.

Do not expand this into a general programming language. Use a small supported subset and explicit, qualified capability boundaries.

Conditional branches and bounded workflows are now representable. What remains missing is how their independently observed execution is related to the model and to the other side’s execution. Behavior owner (docs/02-domain/BEHAVIOR-MODEL.md)

## 9\. Dependency / Target Environment Risks

The dependency model correctly separates strategies from progress:

- DIRECT\_UPGRADE.
- ADAPTER.
- REPLACEMENT.
- REIMPLEMENT.
- RETAIN.
- BLOCKED as a resolution condition.

Partial versus full reimplementation is a scope property, not necessarily another strategy enum. That is a reasonable design.

Practical risks remain:

1. **Declared usage is incomplete.** Dynamic loading, modified vendor code and production-only dependencies require evidence beyond lock files.
2. **Used capability boundaries can be underestimated.** A library’s retry, exception, encoding or configuration behavior may matter even when only one method is called.
3. **Native availability is platform-specific.** Version compatibility does not establish executable availability.
4. **A retained legacy service adds operational obligations.** It needs lifecycle, authentication, session, availability and deployment contracts.
5. **Replacement licensing and acquisition are prerequisites.** They are not post-generation paperwork.
6. **Exact versions need not be numeric release strings.** The current generic component regex rejects exact date-tagged releases. A structural probe confirmed that restriction.
7. **Compiled targets need a clear toolchain/runtime distinction.** Go and Rust resolution should not inherit interpreted-runtime assumptions.

User runtime choices correctly override recommendations. Exact resolution and feasibility are required before execution. **Actual package, extension and target availability remain UNKNOWN until acquisition and qualification evidence exists.**

Target owner (docs/04-migration/TARGET-SPECIFICATION.md), dependency owner (docs/04-migration/DEPENDENCY-RESOLUTION.md)

## 10\. Differential Verification Risks

The strongest remaining risk is false acceptance through contaminated or misattributed evidence.

### Required execution provenance

A report naming `target_tree_hash` is not sufficient by itself. The engine needs witnesses connecting:

- The approved target tree.
- The built image/binary.
- The launched process.
- The environment handle.
- The captured observations.

Otherwise a stale server, stale opcode cache or wrong harness endpoint can produce valid-looking evidence for the wrong implementation.

### Stateful pairing

When source and target choose different branches, the runner must:

1. Record the divergence.
2. Compare the corresponding decision and observed outcome.
3. Avoid pairing unrelated subsequent requests by array position.
4. Preserve safe cleanup and evidence.
5. Distinguish a known wrong branch from an unevaluable guard.

Source and target require separate physical values and cookie jars, but common logical step occurrence identities.

### Identity and normalization

Bijection is necessary but can be ambiguous. Two indistinguishable inserted rows do not acquire correspondence merely because their counts match.

Repeated observations propose dynamic fields; they do not establish harmlessness. A changing refund deadline, nonce lifecycle or randomized ordering can be business-significant.

`EXPECTED_CHANGE` should remain a reported approved difference, not a fourth verdict that bypasses required evidence. UNKNOWN must never be accepted automatically.

### Comparison boundaries

DB reads and locks should be compared where contract-relevant. Query count and SQL spelling should remain diagnostic so optimization is possible. The database owner already makes this distinction; implementers must not turn it into universal query-sequence equality.

Comparison owner (docs/05-verification/COMPATIBILITY-SPEC.md), database relevance (docs/05-verification/DATABASE-COMPARISON.md:5)

## 11\. Database / SQLite Risks

### Application databases

The retained-schema strategy is feasible, but schema preservation does not preserve:

- SQL mode and implicit casts.
- Collation and ordering.
- Time zones.
- Isolation and locking.
- Generated values and sequence consumption.
- Temporary tables.
- Connection-local state.
- Triggers, routines and views.
- Nontransactional behavior.
- Multi-connection consistency.
- Legacy corrupted data.

Sharing the writable database between source and target is correctly prohibited. Keep that prohibition.

Reset must cover application processes and service processes as well as rows. Database routines, native extensions and server-side integrations can create effects outside an application proxy.

### SQLite

The specified local-filesystem, WAL, single-writer, bounded-transaction and backup policy is sound. SQLite’s official documentation confirms that WAL requires same-host operation, has one writer and can grow under checkpoint starvation. [SQLite WAL documentation](<https://www.sqlite.org/wal.html>)

The current embedded-engine requirement includes the documented WAL-reset fix in SQLite 3.51.3. **The actual Go driver build remains unverified**; the Python validator cannot establish it. [SQLite fix details](<https://www.sqlite.org/wal.html>), repository pin (docs/08-storage/SQLITE-SCHEMA.md:3)

### Illustrative growth model

These are sizing assumptions, not measured limits.

| Data | Example workload | Approximate growth before indexes/overhead |
|---|---|---|
| Symbols | 100,000 × 1 KiB metadata | About 98 MiB |
| Cases/contracts | 10,000 × 20 KiB | About 195 MiB of artifacts |
| Observation traces | 10,000 cases × 100 KiB | About 0.95 GiB per full run |
| Repeated verification | 100 such runs | About 95 GiB before deduplication |
| Heartbeats | One persisted event every 15 seconds | 5,760 events/day |
| Heartbeat metadata | Above × 2 KiB | About 4 GiB/year |
| Detailed traffic | 100,000 requests × 100 KiB | About 9.5 GiB/day |
| Learning/tool history | Variable experiments and retained closures | UNKNOWN until measured |

At the heartbeat rate above, 100,000 events represents roughly 17 days of continuous operation. This does not prove failure—the contract permits coalescing—but it shows why the coalescing policy must be operationally precise.

The default 50 GiB artifact quota can be exhausted by repeated protected evidence. Active-project archival and metadata compaction remain deferred. The product must stop honestly rather than silently remove approval evidence.

One project DB plus one optional global DB is sufficient for the narrow baseline. Years-long active history needs measured retention and compaction work before being advertised. Retention owner (docs/08-storage/ARTIFACT-STORE.md:20)

## 12\. Agent / MCP / Skill Risks

The engine boundary is correctly designed. Its enforcement has not been demonstrated.

Remaining risks:

- Untrusted source text entering agent explanations as instructions.
- Agent proposals being mislabeled as observations.
- Imported artifacts being schema-valid but semantically unrelated.
- External patches changing behavior outside their declared ownership.
- Same-account agents invoking human CLI commands outside mediated interfaces.
- Client numeric conversion changing business identities.
- Clients mishandling recursive schema bundles.
- Cancellation being confused with supervisor termination.
- Tool output variants being interpreted as completed work rather than admission.

MCP stdio is a valid baseline, but the protocol requires negotiated lifecycle and capabilities. It does not certify any named client. [Official MCP transports](<https://modelcontextprotocol.io/specification/2025-11-25/basic/transports>)

No actual Rewrotron SKILL.md exists in this repository. SKILL-SPEC describes planned packaging. Generated skills therefore have no implemented promotion, isolation or compatibility evidence.

Without an LLM, deterministic replay and acceptance should remain usable. General interpretation or generation may remain blocked. The current agent specification correctly avoids promising that arbitrary reconstruction will proceed unaided.

Agent owner (docs/06-agent/AGENT-MODEL.md), MCP owner (docs/06-agent/MCP-SPEC.md), Skill specification (docs/06-agent/SKILL-SPEC.md)

## 13\. User Experience / Transparency Risks

The current event-window design is internally coherent: current state at N plus historical events through M. It does not promise reconstructing mutable state at M.

Remaining product risks:

- A live heartbeat can conceal hours of repeated unsuccessful work.
- “Completed assessment” can be mistaken for “migration viable.”
- Historical checkmarks can be mistaken for current acceptance.
- Source actor labels can be mistaken for measured authentication.
- A stale report can appear current merely because it is the latest report.
- Technical capture gaps can disappear behind high-level percentages.
- Event-store failure can prevent persistence of the very blocker being displayed.

### Visibility policy

| Always visible | Expandable | Developer detail |
|---|---|---|
| Current task and reason | Evidence and decisions | Trace/causal identifiers |
| Waiting/blocker state | Scope changes | Capture mechanism controls |
| Required user action | Historical results | Graph and predicate witnesses |
| Currentness/uncertainty | Failed attempts | Qualified environment identities |
| Actual result/verdict | Normalization choices | Sanitized diagnostics |

Private source, credentials, prompts and sensitive raw captures must remain restricted even in developer verbosity.

Add explicit activity categories for **retrying without progress**, **waiting on a process**, **waiting on evidence**, and **cleanup incomplete**. These can use the existing projection; they do not require another event store.

Execution owner (docs/06-agent/EXECUTION-OBSERVABILITY.md)

## 14\. Learning / Self-Improvement Risks

The learning policy now correctly rejects promotion by success count alone.

The remaining hard problem is implementing independence:

- Three projects generated from the same flawed fixture are not independent.
- Different agents using the same assertion template are not independent.
- A holdout derived from the same source interpretation can share the same error.
- A tool that drops traces can improve apparent success.
- A comparator upgrade can make historical patterns appear more reliable without improving reconstruction.

Promotion must retain oracle, fixture, probe, comparator and author correlation groups. Evidence-quality denominators must include unsupported and missing cases.

Retraction needs a discoverable consumer index across paused, moved and temporarily unavailable projects. On resume, a project must check current eligibility of pinned knowledge while preserving historical reproducibility.

The global schemas and outbox lifecycle are explicitly deferred before M7. This is an honest capability boundary, not a foundation blocker.

**Global knowledge must never establish compatibility, pick target versions or waive verification.** Preserve that invariant.

Learning owner (docs/07-learning/LEARNING-SYSTEM.md)

## 15\. Multi-Project / Resume Risks

SQL uniqueness, project ownership triggers, resource claims and process identities are useful foundations. They do not establish filesystem publication safety until race tests execute.

### Crash boundaries

| Interrupted operation | Required recovery |
|---|---|
| Analysis | Preserve partial evidence; publish no complete graph |
| Migration | Reconcile every journaled file; preserve human edits |
| Verification | Terminate owned processes; reset all dimensions; start a fresh attempt |
| External simulator effect | Reconcile receipt before retry |
| Learning promotion | Reconcile project outbox and global receipt |
| Skill activation | Preserve prior active revision and consumer pins |

The current recovery specification appropriately abandons unsafe mid-request replay rather than trying to serialize process memory. Recovery owner (docs/08-storage/RECOVERY.md:17)

### Resume months later

Resume needs available:

- Source snapshot and private fixture bytes.
- Exact executable/build/schema closure.
- Correlation material where required.
- Environment and dependency archives.
- Current validity witnesses.
- Process/resource reconciliation.
- Human pause and approvals.
- Knowledge retraction eligibility.

Missing old binaries must block execution without changing what was historically approved.

Selective reuse has another prerequisite: the engine must know that all relevant consumption edges were recorded. An empty or incomplete graph cannot prove independence. Add a coverage witness for the dependency manifest; otherwise conservatively invalidate the affected category.

Validity owner (docs/02-domain/DOMAIN-MODEL.md)

## 16\. Security / Sandbox Risks

### Side-effect matrix

| Class | Permitted verification | When unavailable |
|---|---|---|
| PURE | Controlled inputs and clock | Missing effect-boundary evidence remains UNKNOWN |
| READ\_ONLY | Owned snapshots or explicitly safe read endpoints | Unproved read-only behavior blocks |
| LOCAL\_WRITE | Independent owned state with qualified reset | Missing reset blocks |
| EXTERNAL\_WRITE | Closed egress and validated sink/simulator | Unsupported/UNKNOWN |
| IRREVERSIBLE | Faithful simulator or explicitly retained boundary | Never blind replay |

The existing matrix is appropriate. It must apply to **all service processes**, including database servers, workers and native helpers, not just PHP HTTP traffic.

Additional decisions required:

- Admission of hardcoded production credentials.
- Evidence sensitivity inherited through hashes and derived artifacts.
- Trusted capture channels versus application-controlled output.
- Parser memory/time limits before oversized AST construction.
- Emergency recovery behavior under disk exhaustion.
- Known source vulnerabilities versus approved intentional security changes.

If the legacy application permits cross-account access, strict reconstruction can reproduce that vulnerability. The assessment must identify it, and the plan must explicitly preserve or intentionally change the behavior. Silent “security improvements” and silent vulnerability inheritance are both unacceptable.

Same-account hostile agents remain outside the claimed isolation boundary. Keep that limitation explicit.

Security owner (docs/09-security/SECURITY.md), sandbox owner (docs/09-security/SANDBOX.md)

## 17\. Vendor-Neutrality Risks

The package direction is neutral. Some machine shapes remain reference-oriented:

- Detection includes `php_compatibility`.
- Inventory counters include controllers, models and views.
- Executable requests are HTTP-shaped.
- Response actions use HTTP status.
- Exact component versions assume numeric release spelling.

These do not make the entire design a CI3→Laravel engine. They do require explicit capability boundaries.

Before broader adapters:

1. Separate common detection facts from language-specific facets.
2. Define executable operation input/output variants for CLI, jobs and events.
3. Distinguish compiler/toolchain identity from deployed runtime identity.
4. Preserve unsupported language constructs rather than mapping them into PHP-shaped categories.
5. Use opaque exact release identity plus adapter-specific version semantics.

Claude Code, Cursor, Codex CLI, Aider, Hermes and headless clients remain **UNVERIFIED**. CLI JSON must be tested as the complete fallback, including human decisions, not just inspection.

Neutral IR boundary (docs/02-domain/IR-CONTRACTS.md), detection schema (specs/schemas/detection-schema.json)

## 18\. Open-Source / Distribution Risks

The repository has an Apache-2.0 license. That alone does not establish distribution rights for future bundled materials.

**UNKNOWN pending evidence:**

- Parser and runtime dependency inventory.
- Licensed fixture sources and data.
- Proprietary SDK redistribution.
- Container-image and native-library notices.
- Generated-code provenance and attribution policy.
- Signing custody and security contact.
- Platform-specific runtime behavior.

The first planned runtime profile is Linux amd64. macOS, Windows and other architectures are targets, not validated distributions.

“Zero setup” must remain limited to interface installation. Legacy PHP, extensions, database fixtures, container availability and safe operational inputs still require setup.

Offline replay is feasible only after exact packages, images, executables and schemas have been acquired and retained. Reproducibility cannot rely on a registry keeping abandoned packages forever.

Distribution owner (docs/10-development/RELEASE.md), fixture acquisition (docs/10-development/FIXTURE-CORPUS.md)

## 19\. Missing Specifications

### Missing Decisions

| Decision | Why it matters | Options | Recommended option | Consequence | Owner |
|---|---|---|---|---|---|
| Interpreter execution roles | Prevents model effects from supplying missing target effects | One interpreter/store; separate simulation and verification | Separate roles/stores with explicit provenance | Additional ports, strong false-PASS protection | BEHAVIOR-MODEL, DIFFERENTIAL-TESTING |
| Source-only identities | Discovery precedes paired verification | Placeholder map; nullable pair; side-local catalog | Side-local identities, later correspondence | Removes forward dependency | RUNTIME-OBSERVATION |
| Evaluation-context layout | Makes pointers and types interoperable | Open JSON; versioned projection | Small typed projection | Unsupported values fail honestly | BEHAVIOR-MODEL |
| Headers/cookies/redirects | Required for login and reference behavior | Framework defaults; typed action/provider | Explicit typed contract | M1 gains a real admission boundary | BEHAVIOR-MODEL, TARGET-ARCHITECTURE |
| Typed fixture/principal records | Prevents actor labels from becoming identity evidence | Opaque refs; shared schemas | Minimal versioned schemas | Validatable initial state and principal provenance | DOMAIN-MODEL |
| Branch pairing | Different branches produce different subsequent work | Independent index pairing; logical occurrence pairing | Logical steps plus divergence outcome | Known wrong branch remains MISMATCH | DIFFERENTIAL-TESTING |
| Hardcoded-secret admission | Exact source may contain prohibited credentials | Refuse; revoke; evidenced overlay | Refuse by default; explicit reviewed alternatives | Some runtime assessments remain blocked | SECURITY, SANDBOX |
| Filesystem behavioral identity | Bytes alone miss metadata-dependent behavior | Ignore; capture all; declared relevant profile | Explicit supported profile and refusal | More snapshot/environment witnesses | ANALYSIS-SPEC |
| Actual execution identity | Prevents stale servers receiving current credit | Expected tree only; build/process witness chain | Witness chain plus stale-build mutant | More run evidence | RUNTIME-OBSERVATION |
| Dependency-graph coverage | Missing edges can cause invalid reuse | Trust graph; mandatory consumption coverage | Engine-derived coverage witness | Conservative invalidation when incomplete | DOMAIN-MODEL |
| Effect payload contracts | Generic refs cannot establish comparison semantics | Opaque blobs; typed dimensions | Narrow typed payloads | Supported comparison becomes executable | COMPATIBILITY-SPEC |
| Performance stopping rule | Repeated looks can inflate acceptance | Fixed N; sequential method | Fixed predeclared N initially | Simpler defensible M6 experiment | PERFORMANCE |
| Generic version identity | Exact releases are not always numeric | Force SemVer; opaque versions plus provider rules | Opaque identity and digest | Broader services remain representable | TARGET-SPECIFICATION |
| ACTIVE versus DORMANT | Current definitions overlap | Reachability-only; observation plus conditional reachability | Mutually exclusive explicit rules | Honest counts without weakening DEAD | CODE-CLASSIFICATION |
| Recovery under full disk | Normal event persistence may be impossible | Best effort; reserved recovery capacity | Reserved capacity and documented fallback | Recoverable failure rather than opaque continuation | RECOVERY, ARTIFACT-STORE |

### UNKNOWN register

| Unknown | Why it matters | Required evidence |
|---|---|---|
| Actual driver/build compatibility | Structural DDL is not Go runtime behavior | Pinned build/version/compile-option tests |
| Reference fixture availability | M2 depends on licensed executable inputs | Source/image/extension/lock manifests |
| Capture fidelity | Correctness depends on complete required dimensions | Escape controls and independent observations |
| Auth/session bridge feasibility | Framework formats can differ materially | Lifecycle and serialization characterization |
| Client compatibility | MCP existence is insufficient | Actual named-client sessions |
| Supported scale | Quotas are not performance evidence | Soak, checkpoint and restore measurements |
| Simulator fidelity | A mock can make both sides agree incorrectly | Protocol/error/effect characterization |
| Global independence/retraction | Generic prose cannot enforce promotion | M7 typed records and holdout experiments |
| Release distribution rights | Engine license does not cover all inputs | SBOM, notices and fixture audit |

### Assumption Elimination Pass

Every unexecuted capability below remains **ASSUMPTION**, even where its specification is sound.

| Claim | What must be true | Conversion into evidence or limitation |
|---|---|---|
| “Detected framework” | Actual bytes support the identification | Detection controls; unresolved mixed installations stay UNKNOWN |
| “Coherent snapshot” | Acquisition is consistent and preserves relevant behavior inputs | ABA/symlink and metadata tests |
| “Dead code” | Operational roots and relevant graph are closed | Omission bundle; no auto-omission on missing closure |
| “Executable contract” | Vocabulary, context and phase are supported | Type/operator/projection qualification |
| “Same behavior” | Scope, environment, assertions and capture are fixed | Scoped differential results |
| “Workflow covered” | Required actor/state/branch partitions exist | Reviewed scenario inventory and mutants |
| “No effects” | Capture or containment covers relevant mechanisms | Mechanism controls; otherwise UNKNOWN |
| “Safe replay” | Every effect is owned, resettable or simulated | Isolation/egress/receipt controls |
| “Restored fixture” | All required stores and processes match | Reset witnesses and contamination tests |
| “Authenticated actor” | Actual session/principal matches the fixture label | Principal provenance test |
| “Harmless dynamic value” | Relations and lifecycle remain meaningful | Approved normalizer plus negative mutations |
| “Compatible dependency” | Availability, API and used behavior all fit | Acquisition, graph and characterization evidence |
| “Approved target” | Exact subject has a current human receipt | Gate/token/currentness tests |
| “Unit complete” | UNIT and implemented regressions are current | Scoped reports; no future-consumer requirement |
| “Reusable artifact” | All prerequisites are known and unchanged | Coverage witness and reuse certificate |
| “Resume works” | Bytes, keys, versions and ownership are recoverable | Crash/restore/upgrade experiments |
| “Progress” | Counted work has a persisted basis | Accounting tests; heartbeat is only liveness |
| “Private evidence” | Every export path applies inherited sensitivity | Canaries, metadata checks and disclosure manifest |
| “Knowledge is useful” | Independent holdouts and mutants support applicability | Promotion/retraction experiments |
| “Performance accepted” | Workload, estimand and stopping rule are fixed | Raw paired evidence and recomputation |
| “Vendor-neutral” | Complete lifecycle works through fallback/interfaces | CLI and actual client tests |
| “Release-ready” | Correctness, data/session continuity and rollback are evidenced | Write-aware staging rehearsal |

No confidence score, narrative explanation or agent assertion discharges these prerequisites.

## 20\. True Implementation Blockers

### P0 — Blocks implementation of the affected core behavior

1. **Interpreter/oracle execution-role contract.** Resolve before implementing a verifier that evaluates effectful contracts.
2. **Source-only identity publication.** Resolve before implementing current effects publication during discovery.

These do **not** prevent isolated storage, canonicalization or CLI-envelope work.

### P1 — Required before the relevant milestone

- Canonical evaluation context and typed fixture/principal inputs.
- Explicit response/header/session capability for M1.
- Cross-side branch pairing.
- Customer-source credential admission.
- Filesystem behavior profile.
- Executed-code provenance controls.
- Dependency-graph coverage witnesses.
- Typed required effect payloads.
- Reference runtime, reset and capture qualification.
- Security disposition of known legacy vulnerabilities.

### P2 — Deferrable behind explicit boundaries

- Generic non-HTTP execution.
- Broader language facets and release-version semantics.
- Long-delay scheduling and distributed concurrency.
- Global consumer/retraction contracts.
- Active-history compaction.
- Statistical method expansion.

### P3 — Later improvements

- Embeddings and vector search.
- Desktop presentation.
- Broader adapter portfolios.
- More sophisticated effort estimation.
- Hosted integrations.

Ordinary implementation details are not architecture blockers. Missing runtime evidence blocks claims and admission, not all foundational development.

## 21\. Recommended Architecture Changes

### KEEP

Keep the three neutral IRs, Target Implementation Model, behavioral slices, shared assessment checkpoint, deterministic gates, SQLite/artifact split and independent replay.

Their ownership is sound; replacing them would discard useful corrections.

### CHANGE

- Define separate semantics for modeling, simulation, application execution and checking.
- Define source-local identity before pairwise identity.
- Make evaluation projections and required payload types explicit.
- Bind runtime observations to actual execution witnesses.
- Extend source/environment identity to relevant filesystem behavior.
- Require prerequisite-coverage evidence for selective reuse.
- Clarify shared-consumer regression scope.

### REMOVE

Remove:

- Mandatory pairwise mapping from source-only capture.
- Current gate status from immutable detection facts, unless explicitly historical.
- Numeric release spelling as a universal exact-version requirement.
- Ambiguous “all consumers” language.
- Any interpretation of model actions as measured application effects.

### ADD

Add narrowly scoped contracts for:

- Evaluation context.
- Side-local identities.
- Principal and initial-state fixtures.
- Response headers/cookies or a qualified provider boundary.
- Execution-origin witnesses.
- Graph prerequisite coverage.
- Secret-bearing source admission.
- Relevant filesystem properties.

These should extend existing owners rather than create another migration pipeline.

### DEFER

Defer general languages, unrestricted temporal logic, distributed schedules, hostile-agent OS isolation, global synthesis and production execution.

The revised narrow flow should be:

```
sealed source + operational/environment profiles
  → source/application evidence
  → side-local observations and identities
  → bounded behavior obligations
  → assessment checkpoint
  → explicit selected work
  → exact target/model/plan
  → approval
  → independently executed target
  → pairwise correspondence + read-only checks
  → scoped acceptance
  → performance/release evidence
```

## 22\. Required Documentation Changes

No changes were made during this review.

| Document | Section | Current problem | Required change | Reason |
|---|---|---|---|---|
| AGENTS.md | Product invariants | Oracle separation is implicit | Reference the explicit execution-role owner | Prevent contaminated acceptance |
| PRODUCT.md | Defensible claims | Correct bounds but no execution-origin condition | Include measured implementation identity | Prevent stale-build credit |
| SYSTEM-DESIGN.md | Pipeline/ports | Model execution roles unclear | Separate simulator, runner and checker ports | Architectural enforcement |
| DOMAIN-MODEL.md | Validity | Graph completeness is assumed | Add prerequisite-coverage witness rules | Safe selective reuse |
| BEHAVIOR-MODEL.md | Vocabulary/context | Namespace layouts and response effects incomplete | Define typed context, phase and supported response behavior | Deterministic semantics |
| IR-CONTRACTS.md | Provenance | Source-local and paired identities conflated downstream | Define identity lifecycle and loss-preservation obligations | Avoid forward dependencies |
| ANALYSIS-SPEC.md | Snapshot | Metadata excluded from identity | Define relevant filesystem profile | Preserve observable behavior |
| CODE-CLASSIFICATION.md | Classification | ACTIVE/DORMANT overlap | Mutually exclusive rules | Honest inventory and progress |
| RUNTIME-OBSERVATION.md | Trace/publication | Source-only identity and executed-code witnesses unclear | Separate capture identities from correspondence | Valid discovery evidence |
| WORKFLOW-DISCOVERY.md | Scenario execution | Divergent paths lack pairing policy | Reference logical occurrence/terminal rules | Stateful comparison |
| TARGET-SPECIFICATION.md | Resolution | Universal numeric version shape | Provider-specific version interpretation plus exact digest | Framework neutrality |
| TARGET-ARCHITECTURE.md | Shared components | “All mapped consumers” includes future units | Current implemented consumers; pending future checks | Avoid renewed deadlock |
| MIGRATION-ENGINE.md | Verification/repair | Effect-role distinction implicit | Explicitly prohibit checker-supplied application effects | False-PASS prevention |
| COMPATIBILITY-SPEC.md | Dimensions | Payload/projection ownership incomplete | Define required typed comparator inputs | Executable comparison |
| DIFFERENTIAL-TESTING.md | Protocol | Branch pairing and oracle store unspecified | Specify independent stores and divergence behavior | Correct replay |
| DATABASE-COMPARISON.md | Reset/capture | Application-level view can miss server processes | Include service-process\# Rewrotron Pre-Implementation Red-Team Review |  |

## 1\. Executive Summary

**The current architecture supports starting a narrow persistence and approval foundation. It does not yet provide a sufficiently precise contract for implementing the reconstruction verifier.**

The documentation has improved substantially since the previous review. Several earlier blockers now have explicit answers: bounded equivalence claims, immutable approval subjects, separate target intent and resolution, scoped unit acceptance, mechanism-based observation completeness, process ownership, and assessment as a checkpoint in one shared workflow.

The remaining highest-impact problems are less obvious:

- The executable behavior language contains writes and effects, but does not clearly distinguish model execution from application execution and verification.
- Source-only observations require an identity-map reference, while the defined identity map belongs to paired verification.
- The evaluator’s `input`, `state`, `actor` and `result` namespaces lack a complete shared projection contract.
- Source snapshot identity excludes filesystem metadata that can affect application behavior.
- Byte-exact source containing production credentials conflicts with the prohibition on placing those credentials in execution containers.
- Source and target workflow branches can diverge without a sufficiently explicit pairing protocol.
- Important reference behavior—particularly response headers and session actions—has no clearly defined lowering in the bounded action vocabulary.

These are specification problems, not demonstrated runtime defects: **there is no Go engine to exercise yet.**

Review evidence:

| Check | Result |
|---|---|
| Repository revision | `0ab4d9bede6f31dcde1cc0105fae6771ee80d39f` |
| Existing Go source | None found |
| Existing executable Agent Skill files | None found |
| Documentation validator | Passed |
| Schemas | 133 |
| Positive/negative structural fixtures | 222 |
| CLI/MCP catalog entries | 71 |
| SQL tables across two migrations | 40 |
| Approval vectors | 10 |
| Runtime, adapter, client or security validation | Not established |
| Repository modifications during this review | None; all 238 inventoried files unchanged |

The structural checks validate shapes, references, selected SQL constraints and hash vectors. They do not validate reconstruction, containment, observer completeness or recovery.

### Document authority map

| Material | Authority and role | Review conclusion |
|---|---|---|
| AGENTS.md:3 | Repository constitution | Current version bounds claims and distinguishes specified from tested |
| docs/00-vision/PRODUCT.md:1 | Product scope and promises | Correctly rejects universal certainty from finite tests |
| SYSTEM-DESIGN and accepted ADRs | Architecture decisions | Generally coherent; remaining execution semantics need decisions |
| Domain and behavior specifications | Lifecycle, identity, semantics and acceptance | Primary owners of the outstanding verifier gaps |
| `specs/`, CLI-SPEC and MCP-SPEC | Machine contracts | Structurally valid; several important referenced records remain underspecified |
| Storage and compatibility specifications | Persistence, replay and comparison | Strong safety direction; operational evidence remains absent |
| DEVELOPMENT-PLAN | Sole implementation queue | M0–M8 replaces historical phase numbering |
| README, project plan, roadmap and architecture summaries | Derived views | Should reference owners rather than introduce independent rules |
| Previous red-team review and resolution report | Historical, non-normative audit evidence | “RESOLVED” means specification disposition, not runtime validation |
| Tests and implementation | Evidence below specifications | Only structural documentation tooling currently exists |

**Duplicates:** architecture overviews, roadmap indexes and glossary documents mostly delegate appropriately.

**Contradictions:** identified in section 4.

**Obsolete material:** historical phase identifiers, old machine-contract branches and the earlier constitution quoted in the conversation must not be mistaken for current executable requirements.

**Unspecified areas:** evaluator projection, model-effect ownership, source-only identity representation, metadata-sensitive snapshots, branch pairing and several referenced infrastructure record formats.

***

## 2\. Architecture Strengths

The following decisions should remain:

1. **Evidence-bounded claims.** Observed, contract and semantic equivalence have distinct meanings. A passing corpus does not establish its completeness.
2. **Three neutral IRs plus a target implementation model.** Source syntax, application structure, behavior obligations and target choices have separate owners.
3. **Engine-owned authority.** Agents propose; deterministic services own gates, publication and verdicts.
4. **Separate intent and resolution.** User constraints cannot silently become different runtime versions.
5. **Immutable approval closure.** Approved subjects retain executable, schema, adapter and decision identities.
6. **Independent verification environments.** Source and target cannot share writable databases, sessions, Redis namespaces or effect endpoints.
7. **Scoped acceptance.** A prerequisite unit can complete before future dependent behavior exists.
8. **Conservative omission.** DORMANT, UNKNOWN and unresolved dynamic paths cannot become automatic omissions.
9. **Assessment completion independent of feasibility.** A useful assessment can finish while migration remains blocked.
10. **Early falsification.** M1 must reject missing writes, wrong principals, duplicate effects and incomplete capture.
11. **Explicit local trust boundary.** Same-account mediation is not advertised as hostile-agent isolation.
12. **Conservative learning.** Global knowledge is advisory, pinned, independently evaluated and retractable.

These are architectural strengths, not implemented capabilities.

***

## 3\. Critical Risks

### CR-01 — Model execution could manufacture verification evidence

The behavior interpreter supports `db_write`, transaction operations, `set_state` and `emit`. Differential testing says to execute bounded steps and compare observations, but does not explicitly assign those model actions to generation, simulation or verification phases.

Concrete failure:

1. The target returns a correct-looking order response but omits the INSERT.
2. The verifier evaluates the contract’s `db_write`.
3. That action writes into the target fixture.
4. The final database comparison passes.

The verifier has repaired the defect it was supposed to detect.

The eventual behavior is **UNKNOWN**, but the specification permits incompatible interpretations. This must be resolved before implementing the interpreter/verifier seam.

**Required decision:** expected-state simulation uses separate shadow state; observation and assertion evaluation cannot create application effects. Any fixture preparation must finish before the measured boundary and carry a distinct provenance class.

Evidence: BEHAVIOR-MODEL (docs/02-domain/BEHAVIOR-MODEL.md:27), DIFFERENTIAL-TESTING (docs/05-verification/DIFFERENTIAL-TESTING.md:13).

### CR-02 — Source-only capture has an unresolved identity dependency

Pre-contract traces explicitly permit `case_id=null`. Their effects document nevertheless requires `identity_map_ref`. The defined verification identity map requires a case ID and paired source/target values.

A source-only observation cannot legitimately supply a target value that does not exist yet.

Read-only schema probes confirmed:

- An effects document without `identity_map_ref` is rejected.
- A verification identity map with `case_id=null` is rejected.

**UNKNOWN:** whether `identity_map_ref` was intended to reference another, currently undefined record.

**Required decision:** distinguish source-local identity observations from a later comparison mapping. Paired identity maps should be derived from observations, without requiring observations to depend on them.

Evidence: runtime observation (docs/03-analysis/RUNTIME-OBSERVATION.md:34), effects schema (specs/observation/effects.schema.json:182), identity-map schema (specs/verification/identity-map.schema.json:13).

### CR-03 — Typed expressions lack a complete evaluation projection

The language names namespaces and JSON Pointers, but does not completely define:

- The structure of `result`.
- How raw HTTP JSON becomes typed values.
- Where observed authentication identity lives.
- Which state is simulated versus measured.
- Which observation a workflow guard reads.
- How DB, cookie, session and asynchronous effects enter assertions.

Two implementations can interpret `/result/body/order_id` differently and both satisfy the present prose.

**Required decision:** a versioned observation-to-evaluation projection, including phase, provenance, exact numeric conversion and namespace ownership.

Evidence: BEHAVIOR-MODEL (docs/02-domain/BEHAVIOR-MODEL.md:23).

### CR-04 — Snapshot identity excludes behaviorally relevant filesystem state

The snapshot protocol records content, executable bits and symlink targets, then says metadata is not identity.

A legacy application can use:

- `filemtime()` for cache invalidation.
- Permissions for upload acceptance.
- Hard links for deduplication.
- Absolute paths for configuration lookup.
- Directory order or case sensitivity.
- Ownership for administrative checks.

Identical bytes can produce different behavior.

**Required decision:** preserve relevant metadata in an execution profile, virtualize it explicitly, or mark affected behavior unsupported. Content identity alone is insufficient.

Evidence: ANALYSIS-SPEC (docs/03-analysis/ANALYSIS-SPEC.md:9).

### CR-05 — Exact source and credential-free execution conflict

Security requires byte-exact private snapshots, including hardcoded credentials. Sandbox rules prohibit production secrets inside containers.

A hardcoded payment key necessarily enters a container if that exact file is executed.

Denied egress reduces the danger but does not remove the secret from process memory, logs or generated output.

**Required decision:** an execution-admission policy for embedded credentials. Options include revoked credentials, an explicitly characterized execution variant, or refusing runtime observation. A transformed execution copy must never be represented as byte-identical source behavior without its own evidence.

Evidence: SECURITY (docs/09-security/SECURITY.md:15), SANDBOX (docs/09-security/SANDBOX.md:7).

### CR-06 — Branch divergence lacks an explicit comparison protocol

If source rejects checkout and target accepts it, subsequent guards may choose different requests.

The specification does not fully determine whether to:

- Execute both paths independently.
- Force the source path onto the target.
- Stop at the first divergence.
- Compare different terminal traces.

Incorrect pairing can turn a known wrong branch into an unrelated setup failure or UNKNOWN.

**Required decision:** logical step-occurrence identities and a first-divergence rule, with controlled continuation where useful.

Evidence: scenario semantics (docs/02-domain/BEHAVIOR-MODEL.md:33), replay protocol (docs/05-verification/DIFFERENTIAL-TESTING.md:13).

### CR-07 — Reference generation exceeds the clearly defined action subset

The `respond` action supports status and body, but not headers. There are no explicit session or cookie actions.

A structural probe accepted a response without headers and rejected the same response with a `Location` header.

An adapter may implement headers through a lowering constraint, but that mechanism and its executable obligations are not defined precisely enough. Otherwise important behavior moves back into prose.

**Required decision:** specify a narrow response/session capability mechanism before claiming the login/create or Laravel reference path can be generated.

Evidence: response action (specs/behavior/semantics.schema.json:555), Laravel adapter (docs/04-migration/LARAVEL-ADAPTER.md:7).

***

## 4\. Architecture Contradictions

### Contradictions

| Conflict | Concrete consequence | Precise resolution |
|---|---|---|
| Byte-exact source versus no production secrets in containers | Hardcoded credentials violate one rule whichever path is chosen | Define execution admission and separately identified execution variants |
| Source-only observations versus required paired identity map | Discovery must invent a target mapping or use an undefined record | Source-local identity catalog first; paired mapping after capture |
| Independent verification versus executable model writes | Checker can alter the measured application state | Separate driver, model simulator, observer and comparator capabilities |
| All mapped consumers must be reverified versus implemented-regression scope | Shared-component changes can be interpreted as requiring future consumers | Say “all currently implemented consumers”; record pending obligations for future consumers |
| CLI catalog version `2.0` versus current `2.1` catalogs | Clients negotiate or document the wrong contract | Correct CLI-SPEC’s opening version declaration |
| ACTIVE has reachability evidence; DORMANT has reachable but unexecuted paths | The classification rules overlap | Define precedence and distinguish observed activity from static reachability |
| Immutable detected facts versus required `gate_status` in detection | A historical detection document can appear to contradict current approval state | Remove gate state from facts or label it explicitly as captured historical state |
| Framework-neutral exact resolution versus numeric release-version regex | A digest-pinned date-named release is rejected | Component-specific exact-version validation; digest remains the byte identity |
| Current scenario v2 versus customer-input description listing v1 fields | Users construct a schema-valid historical scenario that current execution refuses | Document the current workflow/profile/input-ref import closure |

Evidence: unit scopes (docs/04-migration/MIGRATION-UNITS.md:9), shared consumers (docs/04-migration/TARGET-ARCHITECTURE.md:13), CLI version (docs/06-agent/CLI-SPEC.md:3), classification (docs/03-analysis/CODE-CLASSIFICATION.md:3), detection schema (specs/schemas/detection-schema.json:73), resolved versions (specs/project/resolved-environment.schema.json:32), customer inputs (docs/10-development/PROJECT-INPUTS.md:22).

Several apparent contradictions are already resolved and should not be reopened:

- Modernization versus compatibility: intentional changes require explicit approval.
- Autonomy versus confirmation: autonomy operates inside current approved subjects.
- Local-first versus external AI: external clients remain outside core egress control.
- Learning versus reproducibility: consumed revisions are pinned.
- SQLite versus large traces: payloads remain artifacts with quotas.
- Incremental units versus project acceptance: separate scopes exist.

***

## 5\. Practical Failure Scenarios

**D** identifies a specification gap or conflict. **V** identifies an existing defense whose implementation must still be falsified. These are concrete predicted scenarios, not observed engine failures.

| ID | Scenario | Why Current Design Fails | Impact | Detection | Required Design Change | Priority |
|---|---|---|---|---|---|---|
| F01 | Contract interpreter inserts the order missing from target code | D: model writes and verification reads have no explicit phase boundary | False PASS | Missing-write mutant with checker write audit | Shadow model state; read-only checking | P0 |
| F02 | Observe source before target or case exists | D: effects require identity mapping belonging to paired verification | Discovery blockage or fabricated mapping | Source-only capture fixture | Separate local identity observations | P0 |
| F03 | Source takes failure branch; target takes success branch | D: branch pairing is unspecified | Wrong comparison or UNKNOWN instead of defect | Divergent checkout fixture | Logical occurrences and first-divergence verdict | P1 |
| F04 | Assertion reads modeled owner instead of observed owner | D: namespace projection is incomplete | Wrong-principal PASS | Poison expected state while actual write is wrong | Provenance-qualified evaluation contexts | P1 |
| F05 | Order redirect needs `Location` | D: response action has no header field or defined equivalent | Generation blocks or invents behavior | Header-action schema probe | Narrow typed response capability | P1 |
| F06 | Exact source contains a live payment key | D: exactness conflicts with credential prohibition | Secret exposure or forced undocumented rewrite | Embedded-key fixture | Explicit execution admission | P1 |
| F07 | Cache depends on file modification time | D: snapshot identity excludes metadata | Different cache/output behavior | Same bytes, different mtime | Filesystem semantic profile | P1 |
| F08 | Web server executes an old opcode cache after a patch | D: required evidence of executed code identity is insufficiently explicit | Report names new tree but tests old code | Serve stale good build while declaring mutant | Runtime build/process identity witness | P1 |
| F09 | Legacy authenticates account B into account A | D: fixture actor can be confused with actual authenticated principal | Authorization defect hidden | Principal-label mutation | Bind actor to observed auth/session/DB identity | P1 |
| F10 | Shared router is changed while future units are unimplemented | D: “all consumers” wording can recreate acceptance deadlock | Migration stalls | Shared-component dependency fixture | Implemented-only regression rule | P1 |
| F11 | DB-configured route loads an unindexed controller | V: dynamic-root closure is untested | False omission | Database-dispatch fixture | Require root evidence; otherwise UNKNOWN | P1 |
| F12 | Annual cron helper receives zero HTTP hits | V: operational inventory may be incomplete | Lost administrative behavior | Scheduler inventory versus graph | Cron closure controls | P1 |
| F13 | Framework version matches but vendor files are patched | V: actual-byte detection is untested | Wrong instrumentation or lowering | Modified-vendor fixture | Qualification against modified bytes | P1 |
| F14 | CI3 library opens PDO outside framework hooks | V: control closure must detect bypass | Missing writes appear absent | PDO bypass control | Driver/server capture or UNKNOWN | P1 |
| F15 | Procedure writes another catalog or server | D/V: app-level capture may miss server-side effects | Hidden mutation or unsafe replay | Routine/secondary-store fixture | Whole-service containment and qualified effects | P1 |
| F16 | DDL commits despite application rollback | V: reset protocol is unimplemented | Contaminated next case | DDL/reset mutant | Fresh instance or complete reset proof | P1 |
| F17 | Two databases both contain `orders` | V: qualification checks could be omitted by implementation | Wrong table passes | Multi-DB identity mutant | Enforce qualified model/effect equality | P1 |
| F18 | Redis lock survives between scenarios | V: namespaces and TTL reset are untested | Order-dependent results | Contamination fixture | Restore all store dimensions | P1 |
| F19 | Worker delivers email after HTTP success | V: asynchronous causality and receipt capture are untested | False no-effect PASS | Delayed delivery fixture | Await causal completion boundary | P1 |
| F20 | Duplicate queue delivery appears twice under one logical ID | D/V: delivery attempt versus logical effect identity needs precision | Duplicate effect hidden | Acceptance-before-ack crash | Separate attempt, receipt and business identity | P1 |
| F21 | Job is delayed longer than permitted completion window | D: finite quiescence cannot establish future absence | False completeness if scope is misstated | Long-delay job fixture | Scheduled-effect obligation or unsupported scope | P2 |
| F22 | Legacy payment provider has no safe sandbox | V: safe refusal is specified but untested | Real charge or fabricated parity | Endpoint without simulator | Refuse replay; retain UNKNOWN | P1 |
| F23 | Replacement SDK retries non-idempotent operations | V: capability mapping may miss retry semantics | Duplicate business transaction | Timeout/retry characterization | Protocol-level dependency contract | P1 |
| F24 | User requests PHP 8.4; agent proposes 8.3 | V: sovereignty is specified but unenforced | Wrong target | Conflicting resolution proposal | Reject substitution; successor intent only | P1 |
| F25 | Native extension exists only in production | V: acquisition/profile evidence is missing | Boot failure or unsupported behavior | Production-profile inventory | Explicit unavailable capability | P1 |
| F26 | Runtime uses a date-named exact release | D: universal numeric version regex rejects it | False infeasibility | Exact-version schema fixture | Component-specific version grammar | P2 |
| F27 | Large JSON ID becomes a rounded client number | V: end-to-end exactness is untested | Wrong identity | `9007199254740993` transport fixture | Raw bytes/typed values throughout | P1 |
| F28 | Missing query parameter becomes empty string | V: lossless input handling is untested | Wrong validation branch | Absent/empty/repeated input fixture | Preserve distinctions in projection | P1 |
| F29 | Header contains bytes outside supported text encoding | D: string-only request representation needs a boundary | Lossy replay | Byte-level header fixture | Exact supported encoding or explicit refusal | P2 |
| F30 | Normalizer maps both users to the same identity | V: bijection checks are untested | Cross-account PASS | Relation-collapse mutant | Enforce cross-dimension bijection | P1 |
| F31 | Session bridge restores expired CSRF token | V: lifecycle compatibility is untested | Security regression | Rotation/expiry/replay fixture | Safe codec and lifecycle contract | P1 |
| F32 | Two agents claim identical migration work | V: uniqueness/fencing is untested | Duplicate publication | Concurrent claims | Atomic claim and active-work uniqueness | P1 |
| F33 | Old worker resumes after lease transfer | V: SQL fencing alone cannot stop filesystem writes | Target corruption | Pause after check, expire, resume | Ownership across check/write | P1 |
| F34 | Include-time code changes globals without a class call | D/V: file initialization must survive extraction | Missing behavior | Top-level side-effect fixture | Explicit initialization representation | P1 |
| F35 | Source changes A→B→A while scanning | V: coherent acquisition is untested | Mixed baseline falsely accepted | ABA scan fixture | Snapshot/quiescence protocol | P1 |
| F36 | Target changes at 70% and graph omitted one dependency | D: absence of an edge is not proof of independence | Stale result reused | Missing-edge mutation | Consumption coverage witness | P1 |
| F37 | Disk fills after admission but before receipt/event commit | D/V: emergency persistence capacity is undefined | Lost status or unsafe retry | Fill disk during publication | Reservation and recovery policy | P1 |
| F38 | Historical event watermark is displayed as historical state | V: current/window semantics are untested | Misleading supervision | Advance rows while reading old window | Enforce current N versus event M | P1 |
| F39 | Benchmark accepts at first favorable trial count | D: stopping rule is not fixed precisely | Optimistic statistical acceptance | Repeated-look simulation | Fixed sample count or qualified sequential method | P2 |
| F40 | Mean trial p99 hides unacceptable bad periods | D: estimand may not match customer SLO | Misleading performance claim | Burst-tail workload | Separate estimand from operational SLO | P2 |
| F41 | Three generated projects repeat the same wrong oracle | V: independence tests are deferred | Global contamination | Correlated holdout failure | Reject promotion | P2 |
| F42 | “Improved” tool drops hard traces | V: quality benchmarks are deferred | More false PASS results | Capture-dropping tool mutant | Quality and mutation gates | P2 |
| F43 | Retracted knowledge was consumed by an offline project | D: cross-project discovery/reconciliation is deferred | Stale technique resumes later | Detached workspace retraction | Pending retraction checked on reopen | P2 |
| F44 | Target writes cannot be read correctly by rollback source | V: preserved schema is insufficient | Failed cutback late in migration | Write-aware rollback rehearsal | Reconciliation contract or block release | P1 |

***

## 6\. Source Analysis Risks

Static analysis cannot generally establish all dynamic legacy behavior.

| Feature group | Required handling | Remaining practical risk |
|---|---|---|
| Computed includes, variable classes/methods, reflection, magic methods | Preserve unresolved candidate edges | One successful execution does not close other dispatch possibilities |
| Custom autoloaders, containers, service locators | Capture registration and configuration provenance | Bindings can differ by environment or persisted data |
| Hooks, middleware, listeners, framework extensions | Preserve execution order and conditional registration | Early exits and cached output can bypass probes |
| Globals, statics and persistent workers | Model initialization and lifetime | Request-local tests miss process-lifetime state |
| `eval`, generated PHP and runtime patching | Explicit unsupported semantics unless qualified | Safe execution still does not establish unseen alternatives |
| Vendor modifications and mixed framework versions | Compare actual bytes and supported extension behavior | Framework version labels are insufficient |
| SQL, routines, triggers and multiple connections | Include server behavior and qualified identities | PHP call graphs do not describe all database logic |
| Cron, CLI, queues, callbacks and admin routes | Operational-root inventory | Repository scanning misses deployment-owned roots |
| Feature flags, environment variables and DB configuration | Versioned profiles and state partitions | Sanitized fixtures may exclude production-only conditions |
| Uploads, images, PDFs, files and templates | Preserve format and side-effect obligations | A HTTP-only inventory misses processing jobs and assets |

A parser also needs resource budgets. A malformed or adversarial source file must not exhaust the engine before it can persist an honest failure.

### Dead-code conclusion

The current omission bundle is appropriately conservative. Keep it.

However:

- A customer declaration establishes a boundary, not omniscience.
- Zero runtime hits are only corroborating evidence.
- Parsing success does not establish resolved dispatch.
- Unknown data-driven loaders must prevent affected omission.
- File-level initialization must participate in reachability.

Automatic omission should remain **DEAD under approved boundary B**, with the complete current proof bundle in Gate 2. Absolute deadness is not promised.

Evidence: CODE-CLASSIFICATION (docs/03-analysis/CODE-CLASSIFICATION.md:5).

***

## 7\. Runtime Observation Risks

The current mechanism-based attestation is a sound requirement. Its hard part is demonstrating closure against actual escape mechanisms.

Necessary operational checks include:

- Pinned old runtime and extensions.
- Startup, permissions, DNS and TLS configuration.
- Independent DB/cache/session/filesystem state.
- Owned cron, workers and subprocesses.
- Authentication fixtures.
- Actual effect sinks.
- Instrumented versus uninstrumented controls.
- Unknown connections and protocols.
- Capture truncation and disk quotas.
- Observer provenance independent of application narration.

### Observation cannot safely establish everything

Examples:

- Real payment settlement without a faithful simulator.
- Production-only credentials or data-dependent routing.
- Hardware/native behavior without suitable capture.
- Unbounded worker lifetimes.
- External callbacks whose provider cannot be controlled.
- Arbitrary browser or WebSocket execution outside declared capabilities.

These must remain UNKNOWN/UNSUPPORTED or visibly excluded.

### Missing runtime evidence contract

Every measured trace should identify the execution witness connecting it to:

- Actual source snapshot or target build.
- Running process/container identity.
- Loaded executable/application bytes.
- Instrumentation revision.
- Environment/reset profile.

Expected configuration alone is insufficient. Testing an old process while reporting a new tree is a realistic false-acceptance failure.

### Trace volume

Capture needs deterministic limits and admission behavior. Sampling can support diagnostics, but cannot support a required absence claim.

Missing causal propagation must make affected asynchronous evidence incomplete. Timestamps cannot reconstruct causation.

***

## 8\. Behavior Model Risks

The bounded vocabulary is preferable to unrestricted JSON or executable prose. It is still incomplete as a shared implementation contract.

### Required separation

| Layer | May do | Must not do |
|---|---|---|
| Scenario driver | Supply approved inputs and manage fixture bindings | Manufacture expected application effects |
| Model simulator | Compute expected state in isolated model storage | Mutate measured source/target stores |
| Observer | Capture actual behavior | Declare correctness |
| Assertion evaluator | Inspect typed observations and relations | Repair application state |
| Comparator | Derive verdict under approved policy | Change corpus or normalization to obtain PASS |
| Generator | Lower approved obligations | Establish its own verification truth |

### Missing executable details

- Namespace layouts and phase visibility.
- Authentication principal projection.
- Observation-to-typed-value conversion.
- State schema ownership.
- Header, redirect, cookie and session lowering.
- Error-response versus engine-error representation.
- Branch-divergence handling.
- Explicit semantics for qualified provider capabilities beyond primitive arithmetic.
- Concrete payload contracts for effects.

The language cannot currently express many ordinary computations directly: multiplication, division, formatting, hashing and rendering. That does not require a general-purpose language. It requires a precise rule for supported provider capabilities or honest blockers.

### Stateful workflows

Keep bounded graphs and explicit failure terminals. Add logical occurrence identity so repeated identical requests remain distinct executions.

Unsupported loops, concurrency and distributed temporal behavior must remain explicit. Flattening them into a linear success path is unacceptable.

***

## 9\. Dependency / Target Environment Risks

The current strategy separation is correct:

- DIRECT\_UPGRADE
- ADAPTER
- REPLACEMENT
- REIMPLEMENT
- RETAIN
- BLOCKED as a resolution status

Immutable decisions and mutable progress are also correctly separated.

Remaining problems are practical:

1. **Package identity includes modifications.** A patched vendor package is not the upstream release.
2. **Usage includes initialization and side effects.** Three explicit method calls do not prove the rest of the package irrelevant.
3. **Compatibility includes protocol behavior.** Retries, timeout handling, signatures and error classification can change without method-name changes.
4. **Native compatibility includes platform and ABI.** A PHP version constraint does not establish extension availability.
5. **Proprietary availability is separate from metadata.** A replacement cannot be planned as acquired when access is unknown.
6. **License facts must be preserved.** Reimplementation scope does not automatically remove obligations attached to copied material.
7. **Provider capabilities need executable contracts.** Capability names alone do not support generation.

The target configuration separates facts, intent, recommendations and exact resolution appropriately. User overrides already win over recommendations.

The universal version regex remains unnecessarily restrictive. Exactness should mean a component-qualified release identity plus immutable bytes, not mandatory SemVer-like spelling.

**UNKNOWN:** actual reference package/image/extension availability and licensed fixture bytes. Obtain these during acquisition and doctor qualification, before claiming feasibility.

Evidence: DEPENDENCY-RESOLUTION (docs/04-migration/DEPENDENCY-RESOLUTION.md:1), TARGET-SPECIFICATION (docs/04-migration/TARGET-SPECIFICATION.md:1).

***

## 10\. Differential Verification Risks

Comparison is practical only for a declared supported subset with sufficient capture.

The verifier must compare more than final responses:

- Authentication and authorization relations.
- State transitions.
- Committed mutations.
- Relevant reads, transactions and errors.
- Cookies and sessions.
- Files, cache and queues.
- External effect requests, receipts and delivery.
- Redirect and token lifecycle.
- Logical identity across all applicable dimensions.

### Principal false-acceptance hazards

- Comparing source to source accidentally.
- Reading expected state instead of observed state.
- Executing model writes into fixtures.
- Pairing different workflow occurrences.
- Trusting a fixture actor label as authenticated identity.
- Matching duplicate effects as one logical entity.
- Broadening normalization until a defect disappears.
- Declaring a required dimension not applicable.

### Golden masters

Three repetitions identify variability candidates; they do not identify harmless variability.

A varying value may control:

- Refund eligibility.
- Token expiry.
- Account ownership.
- Inventory allocation.
- Retry deduplication.
- Ordering.
- Billing precision.

Normalization must preserve those relationships and kill deliberately wrong mutations.

### Verdicts

- **PASS:** all required scoped obligations have current sufficient evidence.
- **MISMATCH:** an established violation exists.
- **UNKNOWN:** required evidence or interpretation is insufficient.
- **Intentional change:** approved comparison obligations, separately reported.

UNKNOWN cannot be accepted as PASS. An excluded unknown remains excluded—not verified.

Evidence: COMPATIBILITY-SPEC (docs/05-verification/COMPATIBILITY-SPEC.md:5).

***

## 11\. Database / SQLite Risks

### Application databases

The retained-database contract correctly recognizes:

- Connection identity.
- SQL modes and implicit conversion.
- Collation, timezone and precision.
- Transactions, isolation and locking.
- Triggers, routines and generated values.
- Temporary tables and nontransactional behavior.
- Connection-local state.

The remaining implementation challenge is collecting these without relying solely on framework hooks.

Fresh owned instances are preferable whenever reset cannot demonstrate all required dimensions. Sharing the writable application database between source and target remains prohibited.

Database-side external effects also require containment. Isolating only the PHP container does not establish that a routine, extension or secondary database cannot produce effects elsewhere.

Evidence: DATABASE-COMPARISON (docs/05-verification/DATABASE-COMPARISON.md:5).

### SQLite

The current local-filesystem, single-writer, short-transaction and coordinated-backup direction is appropriate. SQLite itself documents that WAL requires same-host coordination, allows one writer and can suffer checkpoint starvation under sustained readers. [SQLite WAL documentation](<https://www.sqlite.org/wal.html>)

The specification’s required SQLite 3.51.3 engine also includes the documented WAL-reset fix. That is a sound requirement, but the repository has not built the selected driver or demonstrated its embedded runtime identity. [SQLite WAL-reset guidance](<https://www.sqlite.org/wal.html>)

**UNKNOWN:** actual driver build, compile options, native filesystem behavior, contention and power-loss recovery.

### Illustrative growth model

These are planning assumptions, not measurements:

| Data | Assumed workload | Approximate growth |
|---|---|---|
| Symbols | 100,000 × 1 KiB | 98 MiB before indexes and evidence |
| Cases | 10,000 × 10 KiB | 98 MiB per retained revision |
| Raw observations | 10,000 × 100 KiB | 0\.95 GiB per corpus run |
| Dense traces | 1 million operations × 100 KiB | About 95 GiB |
| Heartbeat events | One every 15 seconds | 5,760 per day per continuously active producer |
| Heartbeat payloads | 5,760 × 2 KiB | About 11 MiB/day before indexes |

At one persisted heartbeat every 15 seconds, 100,000 events represent approximately 17 days. That is the current soak target, not a storage guarantee or hard capacity limit.

The 50 GiB artifact and 2 GiB metadata defaults therefore require explicit long-running admission and archival behavior. Retained approval closure can legitimately exhaust them.

One project DB plus an optional global DB remains sufficient for the narrow reference path. Years of active history remain a deferred capacity problem.

***

## 12\. Agent / MCP / Skill Risks

The authority boundary is correctly assigned to the engine. Three implementation traps remain:

1. **Ref ownership is not evidence validity.** A same-project artifact can still contain an incorrect claim.
2. **Tool output is not automatically observation truth.** Generated or application-produced reports need trusted capture provenance.
3. **Same-account restrictions are mediation, not isolation.** An external agent with shell access can bypass engine interfaces.

The deterministic engine should remain usable without an LLM for supported execution, replay and acceptance. General interpretation or generation can remain blocked when no supported method exists.

MCP’s initial stdio choice is reasonable and matches the pinned transport specification. It does not establish compatibility with any named client. [MCP transport specification](<https://modelcontextprotocol.io/specification/2025-11-25/basic/transports>)

No actual Skill implementation exists. Its future scripts must invoke guarded services and cannot become another state authority.

A resumed client must receive exact pending decisions, current subjects and unresolved uncertainty. A short conversational summary is insufficient.

***

## 13\. User Experience / Transparency Risks

The current execution-state design correctly distinguishes current state at N from an event window through M. Keep that distinction.

Remaining requirements needing falsification:

- Event persistence failure must stop admission.
- A live worker must not imply forward progress.
- Pause acknowledgement must mean work actually reached a safe boundary.
- Blocked scope and independent runnable scope must remain separately visible.
- Historical completion must not look current after invalidation.
- Partial pages must not display complete unread totals.
- Explanations must be factual and attributed.

### Presentation boundary

| Visibility | Content |
|---|---|
| Always visible | Stage, task, actual activity, blockers, requested decision, outcome, next action |
| Expandable | Evidence, scope, alternatives, invalidations, attempts and approval subjects |
| Developer detail | Qualified resources, spans, hashes, process identities and comparison diagnostics |
| Restricted | Raw secrets, private source, sensitive captures and approval tokens |

There is also a failure-mode UX problem: disk exhaustion can prevent persistence of the event that explains disk exhaustion. The specification needs an emergency reporting/recovery policy rather than promising a durable blocked state when persistence is unavailable.

Evidence: execution failure behavior (docs/06-agent/EXECUTION-OBSERVABILITY.md:150).

***

## 14\. Learning / Self-Improvement Risks

The current learning policy is materially better than promotion by success counts.

Remaining risks:

- Different projects can share one flawed generated oracle.
- Holdouts can leak through tool metadata or earlier experiments.
- Dropped traces can increase apparent success.
- A replacement pattern can be valid only for a narrow capability subset.
- Revocation must reach paused and detached projects.
- Global applicability can drift as dependency releases change.
- A pinned revision can remain reproducible while becoming ineligible for current use.

Therefore separate:

- Historical reproducibility.
- Current applicability.
- Retraction.
- Security eligibility.
- Promotion status.

No global machine lifecycle is currently implemented. Its records and reconciliation are explicitly prerequisites of M7; that is an acceptable deferral.

A single LLM assertion must remain a hypothesis. Neither repeated narration nor repeated generated fixtures makes it independent evidence.

Evidence: LEARNING-SYSTEM (docs/07-learning/LEARNING-SYSTEM.md:1).

***

## 15\. Multi-Project / Resume Risks

The current design includes atomic claims, active-work uniqueness, fencing, serial project publication and process ownership. These are appropriate.

The difficult remaining cases are:

- Source drift while a worker is active.
- A missing dependency edge during selective invalidation.
- Human edits during multi-file publication.
- A supervisor crash after sink acceptance.
- Restoring DB metadata without its artifact closure.
- Missing private correlation keys.
- A client resuming after schema or engine upgrades.
- Global retraction while a project is detached.

### Selective reuse needs positive coverage

A graph with no edge from a changed source input does not establish independence.

Reuse should require an input-consumption manifest or equivalent coverage witness showing that all relevant prerequisite categories were accounted for. Missing coverage must invalidate conservatively.

### Recovery granularity

Interrupted stateful scenarios should restart from a proven reset boundary. They must not resume halfway through with lost in-memory tokens or bindings.

The current recovery specification already directs interrupted replay toward a fresh attempt. Preserve that rule.

Evidence: DOMAIN-MODEL (docs/02-domain/DOMAIN-MODEL.md:11), RECOVERY (docs/08-storage/RECOVERY.md:17).

***

## 16\. Security / Sandbox Risks

### Effect matrix

| Class | Permitted verification | Required refusal |
|---|---|---|
| PURE | Controlled inputs and clock | Unknown effect surface |
| READ\_ONLY | Owned snapshot or proven safe endpoint | Unverified read-only behavior |
| LOCAL\_WRITE | Owned stores with complete reset | Missing isolation/reset |
| EXTERNAL\_WRITE | Faithful sink/simulator with receipts | Real endpoint or unmatched call |
| IRREVERSIBLE | Qualified simulation only | Blind real replay |

Payments, SMS, email, cloud uploads and third-party transactions must never fall through from a mock to a live provider.

### Remaining security gaps

- Embedded production credentials need an execution-admission decision.
- Private source identity must not leak through ostensibly sanitized IR hashes or diagnostics.
- Parser/decompressor resource limits must precede expensive processing.
- Database and worker containers need the same egress scrutiny as application containers.
- Probe output needs a defined trust boundary.
- Source-controlled strings must remain data in explanations and tool selection.
- Package scripts, plugins and native builds require separately controlled acquisition.

### Legacy vulnerabilities

Strict compatibility may reproduce insecure legacy behavior. The product needs an explicit disposition for observed authorization bypasses, injection behavior and unsafe session handling:

- Preserve under visible reviewed risk.
- Intentionally change with new obligations.
- Exclude the affected scope.
- Block release.

The agent must not silently fix a vulnerability and call it equivalent, or silently preserve it and call the target secure.

***

## 17\. Vendor-Neutrality Risks

The core boundaries are mostly neutral. The risk lies in the machine vocabulary and proving fixtures.

Examples:

- `php_compatibility` remains in the general detection schema.
- Controller/model/view counts reflect the reference architecture.
- Executable requests are HTTP-shaped.
- Component versions assume numeric release spelling.
- Current computation and effect capabilities are much narrower than ordinary Django, Go, Rust or Python behavior.

These do not prove that the core is irreparably coupled. They establish that neutrality is **specified, not demonstrated**.

Before broader adapters:

1. Preserve source-language facets without lossy coercion.
2. Define non-HTTP invocation envelopes.
3. Separate compiler/build identities from deployed runtime identities where relevant.
4. Qualify component-specific release formats.
5. Test uncertainty preservation across IR transitions.

Claude Code, Cursor, Codex CLI, Aider, Hermes and custom agents remain UNVERIFIED. CLI JSON must be the complete fallback, including human decisions and recovery.

***

## 18\. Open-Source / Distribution Risks

The repository contains an Apache-2.0 license, but this does not establish redistribution rights for all future bundled material.

**UNKNOWN:**

- Runtime image and fixture licensing.
- Parser dependency selection and license.
- Proprietary SDK acquisition rights.
- Generated code provenance.
- Copied templates, adapters and dependency source.
- Release signing custody and security contact.
- Actual native platform support.
- Offline cache completeness.

Before release, produce a component and fixture provenance inventory with notices and an explicit distribution decision.

“Zero setup” should mean portable installation of the interface. It cannot mean that arbitrary old PHP extensions, databases, proprietary SDKs and safe customer fixtures appear automatically.

The first supported runtime platform should remain narrow. Cross-compilation must not confer runtime maturity.

Evidence: RELEASE (docs/10-development/RELEASE.md:1).

***

## 19\. Missing Specifications

### Missing Decisions

| Decision | Why it matters | Options | Recommended option | Consequence | Owner |
|---|---|---|---|---|---|
| Model-action execution phase | Prevents checker-created parity | Execute in fixtures; simulate separately; generation-only | Separate shadow simulation and read-only checking | Explicit capability boundary | BEHAVIOR-MODEL, DIFFERENTIAL-TESTING |
| Source-only identity representation | Observation precedes target | Placeholder paired map; local catalog; optional untyped ref | Local catalog plus derived paired map | Breaks forward dependency | RUNTIME-OBSERVATION, effects schemas |
| Evaluation projection ABI | Expressions need stable meanings | Adapter-defined; open JSON; shared typed projection | Versioned shared projection | More precise narrow interface | BEHAVIOR-MODEL |
| Branch pairing | Divergent paths need a verdict | Independent paths; forced path; first divergence | First-divergence protocol with logical occurrences | Known branch defects remain MISMATCH | DIFFERENTIAL-TESTING |
| Header/session lowering | Reference login requires it | Prose constraints; new primitives; qualified capability | Small typed capability | Honest generation admission | TARGET-ARCHITECTURE |
| Embedded-secret admission | Exact source may contain credentials | Execute unchanged; transform silently; explicit refusal/variant | Refusal or characterized variant | Some assessments remain static-only | SECURITY, SANDBOX |
| Filesystem semantics | Bytes alone do not fix behavior | Ignore metadata; preserve all; declared relevant profile | Declared profile with unsupported remainder | Reproducible bounded execution | ANALYSIS-SPEC |
| Executed-build witness | Disk tree may differ from running code | Trust requested hash; process attestation; fresh qualified image | Qualified actual-build/process witness | Prevent stale-code PASS | RUNTIME-OBSERVATION |
| Consumption coverage | Missing graph edges are invisible | Trust producer; invalidate all; coverage witness | Witness plus conservative fallback | Reliable selective reuse | DOMAIN-MODEL |
| Effect payload and delivery identity | Duplicate attempts must remain visible | Opaque payload; logical ID only; typed phases/attempts | Typed resource payload and occurrence records | Better async comparison | COMPATIBILITY-SPEC |
| Statistical stopping rule | Repeated looks can bias acceptance | Adaptive unspecified; fixed N; sequential method | Fixed predeclared N initially | Simpler defensible claim | PERFORMANCE |
| Emergency disk policy | Failure can prevent its own event | Best effort; reserved capacity; alternate journal | Reserved recovery capacity and explicit limits | Honest failure reporting | ARTIFACT-STORE, RECOVERY |
| Vulnerability disposition | Compatibility and security can conflict | Silent repair; silent preservation; explicit decision | Explicit intentional-change/risk decision | Transparent release limitations | SECURITY, COMPATIBILITY-SPEC |
| Neutral invocation/version formats | Wider adapters need honest representation | HTTP/SemVer everywhere; open strings; typed facets | Small versioned facets before M8 | Avoid reference assumptions becoming core law | IR-CONTRACTS |

### UNKNOWN register

| Unknown | Why it matters | Evidence needed | Acquisition |
|---|---|---|---|
| Real observer closure | Required absence claims | Executed bypass and containment controls | M1/M2 fixtures |
| Actual engine/driver identity | Durability and reproducibility | Build manifest and runtime checks | M0 |
| Feasible reference target | Package/runtime compatibility | Acquired exact artifacts and recursive feasibility | M3 |
| Session continuity | Login/logout/cutback behavior | Lifecycle and bridge tests | M4/M6 |
| Licensed fixtures | Open distribution | Provenance and terms inventory | M2 acquisition |
| Client compatibility | Portable operation | Named real sessions | M5 |
| Supported scale | Long-running projects | Soak and quota/recovery results | Scale qualification |
| Rollback safety | Target writes can change semantics | Representative write-aware rehearsal | M6 |

### Assumption Elimination Pass

All unexecuted capabilities remain **ASSUMPTION**. The following conversions make claims testable:

| Claim | Prerequisite | Required establishment |
|---|---|---|
| “Detected correctly” | Evidence distinguishes modified/mixed systems | Detection controls and separate assertions |
| “Immutable source” | Coherent content and relevant execution state | Snapshot/drift/metadata tests |
| “Dead” | Closed roots and resolved relevant graph | Approved omission bundle |
| “Same behavior” | Defined scope, projection and comparator | Current scoped contracts and independent observations |
| “Safe replay” | All effect surfaces contained | Egress/reset/sink controls |
| “Complete capture” | Mechanism inventory is closed | Derived attestation with bypass tests |
| “Actor A” | Actual authentication agrees with fixture identity | Session/DB/principal relation checks |
| “IDs correspond” | Mapping preserves distinct entities | Cross-dimension bijection tests |
| “Workflow covered” | Branch/state/actor partitions are examined | Reviewed scenario inventory |
| “Dependency compatible” | Runtime, API and used behavior agree | Acquisition plus characterization |
| “User approved” | Current exact subject has a valid receipt | Gate/token/restart tests |
| “Unit complete” | UNIT and implemented regressions pass | Scoped acceptance evidence |
| “Reuse valid” | All relevant inputs are accounted for | Coverage witness and validity predicates |
| “Resume works” | Bytes, keys, processes and journals reconcile | Crash/restore/upgrade experiments |
| “Progress truthful” | Counts have a persisted basis | Accounting and watermark tests |
| “Private” | Every disclosure path is controlled | Secret-canary and export tests |
| “SQLite sufficient” | Filesystem, engine and volume fit | Runtime checks and soak |
| “Knowledge useful” | Evidence is independent and applicable | Holdout/mutation/retraction tests |
| “Tool stable” | It preserves evidence quality | Capture degradation and security benchmarks |
| “Performance acceptable” | Claim, estimand and stopping rule are fixed | Paired raw measurements and recomputation |
| “Vendor-neutral” | Complete fallback and actual clients work | CLI-only and named-client sessions |
| “Offline reproducible” | Required bytes remain locally available | Disconnected replay |
| “Rollback safe” | New writes and sessions reconcile | Write-aware rehearsal |

No phrase such as “the agent can determine” substitutes for one of these mechanisms.

***

## 20\. True Implementation Blockers

### P0 — Resolve before the affected verifier implementation

1. **Model/checker effect separation.** Otherwise the verifier can manufacture parity.
2. **Source-only identity dependency.** Observation publication must not require nonexistent paired verification evidence.

These do not block unrelated storage foundation work.

### P1 — Resolve before dependent reference work

- Evaluation projection and principal provenance.
- Workflow pairing and first-divergence behavior.
- Narrow header/session/provider lowering.
- Embedded-secret execution admission.
- Metadata-sensitive source execution.
- Actual executed-build witness.
- Input-consumption coverage for selective reuse.
- Typed effect/attempt comparison.
- Shared-consumer acceptance wording.
- Emergency persistence behavior.

Actual runtime acquisition, containment, reset and session evidence also block their corresponding claims. They are missing evidence, not necessarily missing architectural decisions.

### P2 — Can defer behind explicit capability boundaries

- Non-HTTP executable invocation formats.
- Wider computation and language facets.
- General concurrency and long-delay temporal behavior.
- Active-history compaction.
- Global consumer/retraction records.
- Broader platform support.
- Qualified sequential performance methods.

### P3 — Later improvements

- Desktop presentation.
- Embeddings and vector search.
- Richer exploratory recommendations.
- Additional summary and export renderers.

***

## 21\. Recommended Architecture Changes

### KEEP

Keep neutral IRs, vertical slices, guarded services, local SQLite/artifacts, exact approvals, scoped acceptance, conservative omission, independent environments and evidence-based learning.

They address real failure classes and should not be replaced merely because implementation is difficult.

### CHANGE

- Specify model execution separately from measured execution.
- Make evaluator projections canonical.
- Move paired identity mapping downstream of observations.
- Add logical workflow occurrence pairing.
- Include filesystem semantics in the execution boundary.
- Require positive consumption coverage for reuse.
- Bind observations to actual executed builds/processes.
- Clarify implemented-only shared regressions.

### REMOVE

Remove reliance on:

- Placeholder target values during source observation.
- Prose as the sole owner of executable lowering.
- Static reachability as interchangeable with observed activity.
- Universal release-version grammar.
- Mutable gate state embedded ambiguously in immutable detection facts.

### ADD

Add narrow, versioned contracts for:

- Evaluation contexts.
- Source-local identity observations.
- Effect payloads and delivery occurrences.
- Filesystem execution profiles.
- Executed-build witnesses.
- Consumption coverage.
- Embedded-credential admission.

The resulting flow should be:

```
Coherent source + execution profiles
    → Source IR
    → Application IR
    → source observations and local identities
    → reviewed Behavior IR/contracts/scenarios
    → assessment checkpoint
    → explicit selected work
    → exact resolution and target model
    → approved plan
    → implementation
    → independent observations
    → derived paired identity mapping
    → read-only comparison and scoped acceptance
```

Model simulation occupies separate shadow state and never supplies measured application effects.

### DEFER

Keep broad adapters, global synthesis, desktop UI, vector search, hostile-agent isolation and production execution deferred until the narrow verifier survives falsification.

***

## 22\. Required Documentation Changes

These are proposed changes only; no files were edited.

| Document | Section | Current problem | Required change | Reason |
|---|---|---|---|---|
| AGENTS.md | Verification invariants | Model/checker boundary unstated | Reference strict effect/provenance separation | Prevent false authority |
| PRODUCT.md | Claims and support | Supported execution subset needs sharper boundary | Name unsupported projection/lowering cases | Defensible product claims |
| SYSTEM-DESIGN.md | Ports/pipeline | Simulator and checker roles overlap | Separate driver/model/observer/comparator ports | Prevent oracle contamination |
| DOMAIN-MODEL.md | Validity | Missing-edge detection lacks positive coverage | Define input-consumption witness | Reliable selective reuse |
| BEHAVIOR-MODEL.md | Interpreter/scenarios | Namespace, phase and response capabilities incomplete | Define projection, effects and branch pairing | Executable semantics |
| ANALYSIS-SPEC.md | Snapshot identity | Metadata excluded categorically | Add relevant filesystem profile | Coherent behavior baseline |
| RUNTIME-OBSERVATION.md | Trace/publication | Local identity and actual build witness missing | Define source-only identity and execution evidence | Honest capture |
| MIGRATION-ENGINE.md | Generation/repair | Provider lowering and shared regressions ambiguous | Capability admission; implemented consumers | Avoid hidden logic/deadlock |
| MIGRATION-UNITS.md | Shared ownership | Shared-file maintenance needs explicit scope | Define owner maintenance and regression obligations | Schedulable reference work |
| TARGET-SPECIFICATION.md | Resolution | Exact versions use universal numeric grammar | Component-specific release identity | Framework neutrality |
| COMPATIBILITY-SPEC.md | Effects/identity | Payload, occurrence and principal contracts incomplete | Typed comparison records | Detect real defects |
| DIFFERENTIAL-TESTING.md | Replay protocol | Effect roles and divergence pairing unspecified | Read-only checks and occurrence protocol | Avoid self-fulfilling PASS |
| DATABASE-COMPARISON.md | Capture/reset | Server-side effects and application caches need explicit controls | Include all service/process state | Full reset and containment |
| AGENT-MODEL.md | Evidence authority | Tool/probe output trust needs precision | Require trusted capture provenance | Agents cannot establish truth |
| CLI-SPEC.md | Version/imports | Opening version is stale | Align catalog 2.1 and current scenario imports | Deterministic client contract |
| MCP-SPEC.md | Observation/results | Same missing domain records | Expose guarded canonical records | Interface parity |
| EXECUTION-OBSERVABILITY.md | Storage failure | Durable failure event may be impossible | Define emergency/unavailable reporting | Truthful supervision |
| LEARNING-SYSTEM.md | Retraction | Detached consumers unresolved | Reopen reconciliation and applicability policy | Prevent stale learning |
| SQLITE-SCHEMA.md | Operational evidence | Structural checks can be overread | Require actual build/WAL/restore evidence | Durable local operation |
| ARTIFACT-STORE.md | Quota | Admission check does not reserve capacity | Reservation and recovery budget | Survive exhaustion |
| SECURITY.md / SANDBOX.md | Embedded credentials | Exactness conflicts with execution rule | Explicit admission/variant/refusal | Safe legacy execution |
| PERFORMANCE.md | Method | Trial stopping rule not fixed | Predeclared N or qualified sequential method | Defensible acceptance |
| TESTING.md | Falsification | Some oracle defects untested | Add checker-write, stale-build and branch-pair mutants | Catch plausible wrong targets |
| ROADMAP.md / DEVELOPMENT-PLAN.md | Prerequisites | M1 assumes these seams are settled | Resolve narrowly before M1 | Preserve early falsification |
| REWROTRON\_PROJECT\_PLAN.md | Immediate admission | Foundation permission too broad if read as verifier readiness | Distinguish safe foundation from blocked seams | Honest sequencing |

Affected machine owners include behavior, observation, identity, resolution and validity schemas. They must change with their owning specification, rather than acquire undocumented service conventions.

***

## 23\. Recommended New ADRs

Prefer updating existing ADRs:

| ADR | Proposed decision | Falsification |
|---|---|---|
| 0003 — Executable behavior | Model phases, typed projection and supported provider lowering | Checker cannot supply missing target write |
| 0005 — Differential verification | Source-local identity, branch pairing and executed-build provenance | Source-only capture succeeds; stale/wrong target fails |
| 0009 — IR/target model | Neutral invocation and component-specific identity boundaries | Non-reference fixture preserves unsupported semantics |
| 0010 — Approval | Relevant filesystem and execution-variant identity | Same bytes with changed execution semantics cannot reuse approval |
| 0012 — Intent/decisions | Consumption coverage and shared-component scope | Missing edge prevents reuse |
| 0013 — Execution | Emergency storage failure behavior | Disk full does not create success or unsafe retry |
| 0015 — Performance | Fixed stopping rule and claim/estimand distinction | Repeated favorable looks cannot inflate acceptance |

A separate **legacy execution admission** ADR is justified only if credential-bearing execution variants materially constrain the security architecture.

Each decision must record context, problem, alternatives, rejection reasons, consequences, machine impact, compatibility implications and a falsification experiment. Repeated generic ADR boilerplate is insufficient.

***

## 24\. Recommended New Tests / Fixtures

Extend the existing X01–X60 registry rather than create a competing suite.

Mandatory additions:

1. **Checker-write contamination:** omit target INSERT; forbid verifier writes during measurement.
2. **Source-only identity:** publish observations before a case or target exists.
3. **Projection poisoning:** modeled state says owner A; actual DB says B; must MISMATCH.
4. **Branch divergence:** source rejects; target accepts; identify first wrong transition.
5. **Repeated request occurrence:** identical requests twice remain distinct.
6. **Stale build:** report mutant tree while serving old good process; refuse evidence.
7. **Response lowering:** redirect headers, repeated cookies and session rotation.
8. **Metadata-sensitive source:** identical bytes with changed mtime/permissions.
9. **Embedded production key:** refuse unsafe execution or require a separately characterized variant.
10. **Server-side effect:** routine writes another resource outside app-driver visibility.
11. **Late scheduled effect:** finite quiescence cannot imply unlimited absence.
12. **Delivery identity:** one logical message, two deliveries, one retry acknowledgement.
13. **Missing invalidation edge:** changed prerequisite cannot retain current acceptance.
14. **Disk exhaustion:** fail during effect receipt, journal and event persistence.
15. **Exact nonnumeric release:** retain digest-pinned component identity.
16. **Classification overlap:** reachable unexecuted code gets one deterministic classification.
17. **Principal provenance:** fixture label disagrees with actual session identity.
18. **Performance repeated looks:** enforce the approved stopping rule.
19. **Detached consumer retraction:** reopen a paused project after global revocation.
20. **Source-versus-source comparison:** reject two observations from the same side/build.

A mutant rejected because the sandbox failed to start is not a semantic kill. Every test needs the expected failure reason and actual evidence.

***

## 25\. First-Milestone Reassessment

Keep M0 followed immediately by M1. Do not replace early falsification with broad adapter work.

### What M0 should prove

- Coherent bounded snapshot.
- Relevant execution identity.
- Durable exact source approval.
- Reconstructable canonical hash.
- Immutable artifact closure.
- Coordinated backup/restore.
- Honest current execution state.
- Assessment completion without implementation authority.

### What M1 should prove

Use one small two-actor application with:

- Login and state continuity.
- An authorized create operation.
- A forbidden actor branch.
- A DB mutation.
- One acknowledged sink effect.
- Independent source/target environments.
- Exact target resolution and plan approval.
- UNIT and implemented-regression acceptance.
- Missing-write, wrong-principal and duplicate-effect mutants.
- Incomplete capture yielding UNKNOWN.
- Interruption recovery.
- Deterministic repetition without an LLM.

Add two decisive controls:

- The checker cannot modify measured application state.
- Source-only evidence can exist before target comparison.

All seventeen existing premise checks remain mandatory. The fixture need not solve arbitrary PHP, password schemes, browser execution or production cutover.

### Simulation: first realistic CI3 migration

| Stage | What can go wrong | Required outcome |
|---|---|---|
| Detect | Patched CI3, mixed roots, unknown production runtime | Separate measured facts, assertions and gaps |
| Confirm | Snapshot excludes generated files or embeds live secrets | Exact reviewed boundary; runtime admission may remain blocked |
| Analyze | `_remap`, autoload, top-level effects and DB dispatch evade graph | Preserve candidates and UNKNOWN |
| Observe | Missing extension, PDO bypass, unsafe payment endpoint | Safe refusal or incomplete capture |
| Discover | Cron/admin/role branches absent from traffic | Review operational inventory; no false DEAD |
| Model | Session, headers, payment retries or computations exceed vocabulary | Explicit capability gaps |
| Plan | Old packages cannot satisfy PHP 8.3; shared components create cycles | Exact feasible selected scope or blocker |
| Migrate | Laravel middleware changes validation, CSRF or redirects | Deliberate lowering and owned patches |
| Verify | IDs differ, actor labels lie, source/target branch differently | Relational and occurrence-aware comparison |
| Repair | Agent changes normalizer instead of target | Reject policy drift; preserve mismatch |
| Optimize | Fewer queries change locking, precision or authorization | Reverify relevant scope before accepting performance |
| Finalize | Target writes/session formats prevent rollback | Rehearsal or blocked release package |

Assessment may terminate successfully at any point where a supported report can be persisted. A blocked migration is not a failed assessment.

### Framework-neutrality simulations

| Path | Architecture stress | Required boundary |
|---|---|---|
| Django → Go | ORM transactions, timezone behavior, middleware, tasks and model signals | Neutral state/effects; explicit Python source facets; Go build identity |
| Laravel → Rust | Dynamic container bindings, queues, PHP coercion, sessions and error formats | Supported capability lowerings; no silent Rust type coercion |
| Custom PHP → Python | Includes, top-level initialization, globals, raw SQL and ad hoc routing | Generic inventory with honest unknowns; no invented framework structure |

These are future neutrality experiments. None establishes current support.

***

## 26\. Final Go / No-Go Assessment

**What can safely start now**

- Narrow Go bootstrap and injected ports.
- Pinned SQLite build qualification.
- Immutable artifact registration.
- Canonical approval serialization.
- Coherent snapshot acquisition.
- Assessment persistence.
- Current-state/event-window projection.
- Recovery and structural contract enforcement.

**What must be resolved first**

Before implementing the verifier/interpreter seam:

- Separate model effects from measured application effects.
- Remove the source-only identity forward dependency.
- Define typed evaluation projection and branch pairing.

Before customer runtime/reference generation:

- Resolve embedded credentials and filesystem semantics.
- Define required response/session lowering.
- Establish actual executed-build provenance.
- Execute containment, capture and reset controls.

**What can be deferred**

Broad adapters, general concurrency, long-delay distributed behavior, global synthesis, desktop UI, vector search, active-history compaction and production execution can remain behind explicit capability boundaries.

**What remains assumed**

Every runtime capability, native platform, client integration, scalability claim, session bridge, dependency resolution and rollback strategy remains unvalidated until its named experiment executes.
**Assessment:** proceed with the bounded foundation; hold verifier expansion until the identified contracts are explicit and the wrong-target experiment rejects defects for their intended reasons. The repository is structurally consistent enough to begin that foundation, but it has not yet demonstrated that its verifier can distinguish a correct reconstruction from a convincing wrong one.