# Product specification

**Status:** Authoritative design specification. **Implementation:** Not started. **Target:** First scoped CI3→Laravel reference release, then broader adapter coverage.

Rewrotron reconstructs externally observable application behavior in a modern target stack. The canonical pipeline is Source IR (syntax) → Application IR (structure) → Behavior IR (rules, inputs, state, effects and outputs) → target plan → target adapter → differential verification and repair. Migration units are behavior slices.

## Two central product requirements

1. Rewrotron preserves what the application does within its declared, evidenced scope.
2. Rewrotron always shows what it is doing and why, including discoveries, changes, blockers, next actions and needed user decisions.

[EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) makes the second promise a deterministic engine contract from the first long-running operation. Shared events/state serve CLI, MCP and a future desktop UI. Factual explanations, honest scoped progress, activity history, return-session changes and safe persistent control are baseline product responsibilities, not a later UI feature.

## Product promise and evidence boundary

Produce a maintainable independent target, evidence for its compatibility with a declared case corpus and scope, and an accountable inventory of omissions, exclusions and unresolved behavior. Finite differential testing does not prove equivalence for every possible execution. Unknown behavior is disclosed and cannot pass. Hashes establish evidence identity/integrity, not that analysis is complete.

Source and target adapters decouple framework extraction from generation, allowing N source implementations and M target implementations to reuse the same model. An additional adapter enables only the capability intersection it supports; every claimed source/target combination requires corpus evidence. Core packages remain vendor/framework-neutral and run without an LLM.

## Initial scope

| Capability | First release commitment | Limitation |
| --- | --- | --- |
| Source | CI3 3.1.x reference adapter | Dynamic extensions and unsupported PHP syntax remain explicit |
| Fallback | Generic PHP inventory/extraction | No automatic general PHP reconstruction guarantee |
| Target | Laravel 13 on pinned PHP 8.3+ | Capability-checked evidenced rules; no universal code synthesis |
| Architecture | Modular monolith | Other styles require adapter support and reviewed target plan |
| Data | Preserve relational schema; independent MySQL sandbox copies | Other engines need their own reset/comparison execution capability |
| Interfaces | CLI JSON and MCP stdio, portable Skill | No provider-specific engine logic |
| Storage | SQLite WAL and local artifacts | Global knowledge optional; no mandatory cloud/vector DB |
| Safety | Isolated execution, target-only generation, no production mutation | Cutover package/runbook after Gate 3; deployment implementation deferred |

## Gates and customer authority

Minimal workspace creation and bounded detection can precede Gate 1 so approval can be persisted. Deep analysis/observation/modeling require source approval. Generation and repair require the exact reviewed Gate 2 plan. The verified release/runbooks require Gate 3 before cutover packaging. Gate validity, user overrides and material drift are specified in [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md).

Customers supply critical workflows, operational roots, sandbox-safe fixtures/integrations and explicit decisions where source facts are ambiguous. The engine reports the missing item and dependent units. It never invents infrastructure, credentials, roles, runtime versions or business logic.

## Adapter maturity

No adapter is currently implemented/tested. A planned adapter becomes `EXPERIMENTAL` when registered with a version and capability manifest. `BETA` requires deterministic reference corpus coverage, safe failure modes, gate integrity and documented limitations. `STABLE` additionally requires the full supported corpus/CLI/MCP/recovery suite, security checks, reproducible target builds, release evidence and a published support matrix. Labels apply to tested capabilities, not every application in a framework family.

Additional PHP/Python/Ruby/Java/.NET sources and Go/Rust/Python/TypeScript targets are roadmap candidates, not available features.

## Successful migration

An in-scope unit succeeds when evidenced contracts/cases exist, target code builds independently, every required dimension passes against current subjects, committed state/effects and auth/session behavior match, regression cases pass, and mappings/artifacts are recorded. The product's performance goal is non-regression under the agreed reproducible workload; inconclusive evidence remains visible.

A project additionally reconciles every entrypoint with a unit, current DEAD proof, or explicit reviewed exclusion; all required cases pass; there are no unresolved required UNKNOWNs or prohibited differences; performance, active sessions, integration behavior and rollback are evidenced; omissions and intentional changes are disclosed; and Gate 3 approves the exact release. Reports use explicit denominators from [COVERAGE](../05-verification/COVERAGE.md).

## Release progression

First deliver safe workspace/detection/approval and negative gate tests, then a machine-verifiable behavior model, then a deterministic one-slice fixture migration, then auth/state/integration/generalization evidence. A first demo is not full production framework support. [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md) defines the dependency order and tasks.

## First-class migration decisions

Source facts, user target intent, the plan and executed results are distinct records. [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) defines editable `rewrotron.yaml`, immutable planning revisions, recommendation types, feasibility and selective target invalidation. A requested runtime is never silently lowered to a recommendation. Gate 2 approves the exact target and plan; YAML approval views are not authorization.

[DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) defines exact inventory, actual usage, recursive constraints and capability-based upgrade/adapt/replace/reimplement/retain decisions. Required unresolved dependencies block affected scope before generation. Behavioral characterization validates proposed implementations; optional global knowledge remains evidence-scoped and opt-in.
