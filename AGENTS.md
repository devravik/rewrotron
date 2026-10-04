# AGENTS.md — Rewrotron constitution

## Authority and scope

Applies to all source, documentation, schemas and tools. Authority: this file → [Product](docs/00-vision/PRODUCT.md) → [System design](docs/01-architecture/SYSTEM-DESIGN.md) and accepted ADRs → domain/behavior specifications → machine interfaces/schemas → storage/compatibility contracts → tests → implementation. Lower documents cannot redefine higher invariants. [Documentation index](docs/README.md) assigns topic owners; derivative documents reference those owners. Resolve an actual conflict explicitly before dependent work.

## Product invariants

Rewrotron reconstructs legacy application behavior in a user-selected target. Core is language, framework and vendor neutral. Source IR → Application IR → Behavior IR preserve provenance and uncertainty; the [Target Implementation Model](docs/04-migration/TARGET-ARCHITECTURE.md) records target choices. Behavioral vertical slices, not files, are migration units.

Claims are bounded by the approved source/environment/operational profile, corpus, executable contracts, comparison policies and exact target resolution. Finite tests do not prove universal equivalence or corpus completeness. [Behavior terminology](docs/02-domain/BEHAVIOR-MODEL.md) owns these meanings. PASS / MISMATCH / UNKNOWN are verifier results; UNKNOWN never becomes PASS by agent assertion, waiver or missing capture. Intentional changes remain explicit approved policy decisions.

Source framework skills package source adapters and source knowledge; target framework skills package target adapters and TIM lowerings under [SKILL-SPEC](docs/06-agent/SKILL-SPEC.md). Compose them through neutral IR/TIM and exact qualification, never a required framework pair. CI3 → Laravel is Reference Profile #1 intended for first qualification; it is currently SPECIFIED. Source adapters own source conventions; target adapters own target conventions. No framework assumptions or provider SDKs/prompts belong in deterministic domain/analysis/verification. Core uses injected adapter/storage ports; implementations depend on these interfaces, avoiding circular imports. CLI, MCP, Skills and runners call the same guarded services. No unrestricted SQL/shell in agent interfaces. CLI JSON is the complete vendor-neutral fallback. [MCP](docs/06-agent/MCP-SPEC.md) owns the pinned stdio baseline and deferred transports.

## Authority over execution

The engine owns state, acceptance, permissions and the three persistent human gates. Agents propose, never establish truth. [Confirmation protocol](docs/02-domain/CONFIRMATION-PROTOCOL.md) owns canonical subjects, explicit local human approval/token handling and invalidation. Silence, a display event, edited YAML and agent-declared actor identity do not approve anything. Gate 1 authorizes analysis/observation of a coherent snapshot; Gate 2 authorizes execution of an exact plan; Gate 3 approves a release package/runbook after whole-project acceptance and required performance/rollback evidence. Initial release does not deploy production.

Assessment is a checkpoint in the shared analysis/migration workflow, owned by [Project lifecycle](docs/02-domain/PROJECT-LIFECYCLE.md). It may complete with blocked or unknown migration feasibility. Assessment, work selection and continuation never authorize implementation; executable planning requires an explicit human continuation decision, and implementation still requires Gate 2. No parallel assessment engine or automatic continuation.

Detected facts, human assertions, requested target, recommendations, exact resolution, approved decisions, progress and verification are separate records. User intent overrides recommendations; infeasible choices block, never silently downgrade. Target/decision changes preserve history and invalidate affected work through [typed validity rules](docs/02-domain/DOMAIN-MODEL.md). Unit acceptance is scoped; future dependent cases are not prerequisites to completing their prerequisite unit.

DEAD means DEAD under an approved closed boundary, not proof of never being used. Only [the omission evidence bundle](docs/03-analysis/CODE-CLASSIFICATION.md) included in Gate 2 permits automatic omission. DORMANT, UNKNOWN and unresolved dynamic paths are never automatically omitted.

All long operations expose committed stage/task/reason/outcome/blockers/next action. [Execution observability](docs/06-agent/EXECUTION-OBSERVABILITY.md) owns truthful progress and current-state/event-window semantics; [Recovery](docs/08-storage/RECOVERY.md) owns foreground/supervised process lifetimes. Do not claim background survival without a supervisor.

SQLite plus local artifacts is authoritative; cloud/vector services are optional. Large evidence is stored outside SQLite. Approval/release evidence cannot be silently collected. [Storage](docs/08-storage/SQLITE-SCHEMA.md) owns driver/version/filesystem policy. Verification uses independent environments and proven reset profiles. Irreversible external effects are never blindly replayed. [Security](docs/09-security/SECURITY.md) distinguishes mediated enforcement from same-account trust; it does not promise hostile-agent OS isolation.

Learning is evidence-based, scoped, versioned, retractable and advisory. Pin engine/build/schema bytes and all consumed tool/skill/knowledge revisions through the approval execution closure. Global knowledge cannot approve targets or establish compatibility. [Learning](docs/07-learning/LEARNING-SYSTEM.md) owns promotion.

## Status and development rules

The repository is specification-first. Planned CI3/Generic PHP source and Laravel target adapters are SPECIFIED, not runtime tested. Validation maturity: DESIGNED → SPECIFIED → STRUCTURALLY_VALIDATED → IMPLEMENTED → UNIT_TESTED → RUNTIME_VALIDATED → INTEGRATION_VALIDATED → RELEASE_VALIDATED. Each transition requires its named evidence; schema validation is not execution. Adapter distribution tiers EXPERIMENTAL/BETA/STABLE are separate and cannot imply maturity evidence. All client integrations are UNVERIFIED until exercised.

Before changes read this file, owning specifications, relevant ADRs, schemas and tests; inspect existing implementation without treating it as authority. Follow [Development](docs/10-development/DEVELOPMENT.md). Keep changes atomic; update public CLI/MCP contracts with signatures, and storage changes with incremental checksummed migrations and compatibility fixtures. Do not rewrite applied migrations. Go implementation must use the pinned toolchain, explicit constructors, structured errors, gofmt/goimports, documented exports and no global mutable state. Implementation work requires meaningful tests, `go test -race ./...` and configured lint checks; documentation-only work runs structural validation and states its limits.

Completion means alignment with owning specs, preserved neutrality/gates, machine-contract parity and evidence appropriate to the claimed maturity. The [current development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md) sequences falsification before broad capabilities. Do not implement deferred capabilities by pretending they are supported.

Model actions execute only in isolated shadow simulation or generated application code. Fixture preparation ends before measurement. Drivers/checkers/comparators cannot manufacture application effects or fill missing observations from expectations. [Behavior roles](docs/02-domain/BEHAVIOR-MODEL.md) and [evaluation context](docs/02-domain/EVALUATION-CONTEXT.md) own this invariant. Source-only identities precede paired correspondence; acceptance requires trusted execution witnesses for distinct SOURCE and TARGET sides.
