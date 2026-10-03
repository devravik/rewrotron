# Rewrotron project plan

**Maturity: SPECIFIED.** This plan replaces the original exploratory phase list. [PRODUCT](docs/00-vision/PRODUCT.md) owns product claims; [SYSTEM-DESIGN](docs/01-architecture/SYSTEM-DESIGN.md) owns architecture; [DEVELOPMENT-PLAN](docs/11-roadmap/DEVELOPMENT-PLAN.md) owns exact milestone contracts. This overview does not compete with them.

## Objective and constraints

Reconstruct approved observable behavior in a user-owned target using Source/Application/Behavior IR, executable contracts, a Target Implementation Model and independent differential verification. Claims identify bounded evidence/profile/corpus/policy/environment/resolution. Core remains deterministic, language/framework/vendor neutral and local-first. Agents propose; engine owns gates, acceptance and safety. UNKNOWN is never PASS. Human decisions and exact consumed revisions remain reproducible.

## Work order

M0: sealed coherent snapshot, pinned SQLite/artifacts, canonical approvals and honest CLI state. M1: immediately prove a bounded stateful verifier can reject missing-write, wrong-principal, duplicate-effect and incomplete-observation mutants with independent environments, scoped acceptance and interruption recovery. M2: characterize the real CI3 reference safely. M3: freeze exact target/dependency decisions and implementation model. M4: generate/verify reference slices and shared regressions. M5: bounded repair and guarded machine/client integration. M6: whole-project corpus, explicit performance claim, rollback/session/data rehearsal and Gate 3 package. Optional M7 learning follows evidence-quality benchmarks; M8 broader adapters/UI/scale earns separate support.

Each milestone's objective, prerequisites, inputs, outputs, owned specs, scope, falsification tests, acceptance, evidence, blockers and non-goals is defined once in the development plan. Historical RWT IDs are provenance aliases there, not parallel priorities.

## Immediate admission conditions

Documentation/schema checks can validate shapes, hash vectors and incremental DDL. Foundation implementation may start under current owner specs after pin/build verification. Runtime support, observation completeness, semantic acceptance, rollback/performance and client compatibility remain UNKNOWN until measured. Do not start broad reconstruction if M1 accepts a wrong target. Do not start implementation as part of a documentation revision.

## Deferred boundaries

General streaming/concurrency, broad languages/frameworks, hostile same-account isolation, global synthesis, desktop/vector search, production runner and unbounded active-history compaction are deferred. No promise of any-source/any-target, universal proof, zero legacy-runtime setup or irreversible-effect replay. Each capability needs detection/refusal behavior and later falsification evidence before advertisement.
