# System design

**Maturity: SPECIFIED; no engine implementation claimed.** Core owns evidence, immutable subjects and guarded deterministic services. [PRODUCT](../00-vision/PRODUCT.md) owns bounded claims; [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) owns hashes/gates.

## Canonical pipeline

```text
Coherent sealed source + environment/operational profiles + requested intent/scope
 → shared Source IR → Application IR → Behavior IR + evidence/uncertainty
 → Assessment Report (partial evidence permitted; successful stopping point)
 → explicit human work/continuation decision
 → currently valid contracts + bounded scenarios + exact feasibility/resolution
 → immutable proposed Target Implementation Model + executable plan subject
 → explicit Gate 2 approval → staged implementation
 → isolated differential verification → scoped acceptance/repair
 → performance evidence → release package subject → Gate 3
```

Gate 1 precedes deep analysis or runtime execution. Bounded detection and snapshot acquisition do not run source code. The target model is proposed before Gate 2 because its ownership/lowering choices are approved; activation and generation follow confirmation. [TARGET-ARCHITECTURE](../04-migration/TARGET-ARCHITECTURE.md) owns that model.

## Boundaries

CLI/MCP/Skills call guarded service ports. Domain packages define adapter/repository/runner interfaces; adapters and storage implement them. Composition in cmd wires implementations; domain never imports framework packages. Source adapters translate conventions into three neutral IRs without erasing uncertainty. Target adapters consume Behavior IR plus resolved environment and target model, never a CI3-shaped core. [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md) owns neutral record requirements.

Services: snapshot/detection, approval, analysis, observation, behavior, target/dependency resolution, planning, publication, verification, execution and project learning. These are responsibilities, not mandatory one-package-per-noun implementation. SQLite indexes durable immutable artifacts and current projections. Large/private bytes live in filesystem storage. [SERVICE-CONTRACTS](SERVICE-CONTRACTS.md) maps transactions and guarded operations.

## Execution ownership and concurrency

Foreground CLI operations belong to the calling engine process; termination cancels at a safe boundary and a future resume creates/reconciles an attempt. Explicit `--background` submits to an engine-owned local supervisor. Long MCP actions require that supervisor; unavailable supervision returns ERR_UNSUPPORTED_CAPABILITY, never fictitious survival. Watch/client disconnect does not stop independently supervised work. [RECOVERY](../08-storage/RECOVERY.md) owns heartbeats, leases, orphan cleanup and reconnect.

Initial mutation/publication is serialized per project. Atomic mission deduplication and resource claims prevent shared-file races. Read-only inspection can run concurrently. Parallel disjoint mutation is deferred until per-resource revision experiments pass; project CAS invalidation is acceptable under the serial baseline. Events do not advance business revisions merely for telemetry.

## Acceptance and unsupported capabilities

[Migration units](../04-migration/MIGRATION-UNITS.md) distinguishes UNIT, IMPLEMENTED_REGRESSION and PROJECT acceptance; only whole-project acceptance supports release readiness. Verification requires independent source/target state, reset evidence, mechanism-derived completeness and a tested comparator. Unsafe effects, unsupported rule operators and unclosed observation mechanisms block or return UNKNOWN, never empty PASS.

The first supported subset is bounded HTTP/stateful scenarios and explicit synchronous/acknowledged fixture effects. General streams, arbitrary distributed concurrency, hostile local-agent isolation, production deployment, global synthesis and broad adapter portfolios are deferred capabilities, not hidden assumptions. [ROADMAP](../11-roadmap/ROADMAP.md) sequences expansion.

Assessment is a projection/checkpoint of the existing domain services, not an alternative analysis engine. [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md) owns report completion, human selection and optional continuation. Partial/unsupported behavior can be reported before executable semantics or feasibility are ready; downstream execution guards are unchanged.

Implementation admission resolves the exact engine/build/schema closure and adapter qualification owned by CONFIRMATION-PROTOCOL/SERVICE-CONTRACTS. URI/version strings and advertised profile metadata alone cannot confer execution support. DB lowering uses qualified data-model 2.0; current attestation/benchmark versions and their independent derivation requirements are specified by their owners. Performance policy is an immutable plan decision, never mutable execution progress. Assessment keeps using the same IR/evidence/qualification services and can publish gaps without a feasible executable plan.

## Measurement ports and evidence direction

BEHAVIOR-MODEL owns role capabilities. Inject separate FixturePreparer/ScenarioDriver, ShadowSimulator, ApplicationRunner, Observer, AssertionEvaluator, Comparator and Generator ports. A simulator receives only shadow-state handles; evaluator/comparator receive immutable capture refs and read-only projections. They receive no writable application DB, cache, session, filesystem, queue or sink handle. Fixture preparation is an explicit, witnessed pre-measurement phase, not a model action during replay. ApplicationRunner invokes real application bytes; generation does not establish verification truth.

Runtime capture flows execution witness → side-local identities/effects/context → immutable trace. Later verification derives occurrence pairing and pairwise correspondence → assertions → scoped report. Capture never references its future paired report. Expected/model state cannot supply missing actual values. EVALUATION-CONTEXT owns the projection ABI; RUNTIME-OBSERVATION owns witness/publication. These are specified enforcement boundaries awaiting M1 negative controls.
