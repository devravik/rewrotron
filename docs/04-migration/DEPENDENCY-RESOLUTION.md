# Dependency usage, capabilities and resolution decisions

**Status:** Normative design specification; implementation pending. The engine asks what behavior the application requires from a dependency and how to reproduce it safely in the requested target. Package-name similarity is never evidence of equivalence.

## Resolution pipeline and ownership

```text
Exact source inventory → graph and actual usage → required capabilities
  → requested-target compatibility → API/behavior difference analysis
  → proposed resolution → user decision in Gate 2 plan
  → implementation → differential characterization → performance evidence
  → project lesson → optional reviewed global promotion
```

Bounded detection inventories manifest/lock/runtime facts before Gate 1. Deep symbol reachability and runtime observation occur only after Gate 1. `internal/dependency` owns neutral graph, capability, compatibility and resolution rules; ecosystem parsing, registry metadata acquisition and API extraction live behind adapter ports. Deterministic core consumes versioned local evidence and operates offline; unavailable metadata becomes UNKNOWN. Agent proposals cannot supply invented compatibility facts. Package/network tools use controlled runners, never raw MCP shell access or a blind `composer update`.

`internal/recommendations` produces proposals; `internal/target` validates frozen user intent; `internal/migration` plans slices against approved dependency decisions; `internal/knowledge` records verified lessons. Detector, graph, resolver, compatibility, capabilities and mapper can be modules inside `internal/dependency`; compatibility adapters use existing adapter boundaries and learning uses the shared knowledge subsystem. Do not create duplicate approval, planning, artifact or learning engines.

## Immutable source inventory and recursive graph

Record ecosystem/package identity, exact installed version, manifest constraints, lock/content hashes, origin/license, direct/transitive/development/optional role, environment/platform constraints and evidence. Installed versus declared versions remain separate; missing locks or ambiguous identities remain explicit unknowns. Infrastructure such as databases, cache and external services is inventoried with sanitized identity and contracts without treating it as a package upgrade.

The graph includes packages, runtime/framework, extensions/native libraries and retained services. Typed edges express requires/conflicts/provides, platform constraints, optional features and affected units. Preserve source graph history; target resolution creates another graph. Resolve replacements recursively, including their new dependencies. Report cycles, conflicting constraints, unavailable metadata and unresolved native/platform requirements. Shared component cycles may be modeled as strongly connected groups; the migration-unit execution graph must remain schedulable.

`.rewrotron/source-dependencies.yaml` exports source inventory/usage; `.rewrotron/dependencies.yaml` exports decisions and target mappings. Canonical bytes are registered artifacts, not mutable YAML authority. Every record includes input hashes/tool versions and provenance, so resumed agents can reuse findings without rediscovery and can recognize stale evidence.

## Actual usage and safe omission

Record referenced classes/functions/methods, signatures, configuration paths, source spans, reachable entrypoints/units, runtime observations and required capability subset. Declared, statically reachable, observed and required-by-approved-behavior counts have separate denominators; none is substituted for another. Zero runtime hits do not prove a dependency unnecessary. Reflection, dynamic loading, optional jobs and customer operational roots can preserve UNKNOWN or DORMANT usage.

Distinguish ACTIVE, DORMANT, DEAD and UNKNOWN code classifications from development-only/optional dependency roles. “Unused” is a finding, not permission to delete. Automatic omission requires current static unreachable evidence and zero runtime traces under the repository DEAD rule. A development package may still be required to build or verify the target. Replacing a required behavior with native target functionality records its mapping and parity evidence; it is a planned implementation replacement, not a false DEAD classification.

## Strategies and blocked state

| Strategy | Selection evidence and obligations |
| --- | --- |
| `DIRECT_UPGRADE` | Compatible target release; migration mapping for every changed used API; exact version/lock and cases |
| `ADAPTER` | New implementation provides required capabilities behind an explicit compatibility boundary preserving legacy semantics |
| `REPLACEMENT` | Alternative satisfies required capabilities, output/error/edge-case semantics, runtime/platform, license and performance constraints |
| `REIMPLEMENT` | Implement only the evidenced used capability subset, possibly with native target facilities; reviewed contracts and cases required |
| `RETAIN` | Keep a compatible package/service or explicitly isolated legacy component; operational/session/security and availability contracts still required |

`BLOCKED` is a resolution status, not a sixth successful strategy and never a verification verdict. It records missing capabilities/evidence, affected/dependent units, alternatives and the exact human decision needed. A legacy bridge or user adapter is a new proposal requiring support and verification, not an automatic escape hatch. The agent cannot invent an implementation and declare equivalence. Compatible direct retention may also be proposed; a native replacement is represented as REIMPLEMENT rather than an undocumented strategy.

Immutable v2 dependency resolution contains selected strategy/provider/capability evidence and decision_ref. It has no mutable execution status or approval flag. The decision is PROPOSED until its ref is approved transitively by Gate 2; approved bytes never change. Separate progress.schema records PLANNED, GENERATING, BUILDING, VERIFYING, VERIFIED or BLOCKED keyed by decision/ref/plan. VERIFIED requires current characterization and affected application scope PASS. BLOCKED is a progress/feasibility reason, never a successful strategy or verdict. Independent scope may proceed only under a current plan exposing exclusions/blockers; human approval cannot make a technical incompatibility feasible.

## Capability coverage, compatibility and API differences

A capability record names its input/output shapes, preconditions, state/effects, error behavior, configuration and evidence/case refs. Map each required capability to a target provider (library, native facility, adapter or retained service). Missing streaming support, timeout semantics or encryption/error behavior is explicit. Capability coverage reports `covered required capabilities / total required capabilities`; unknowns remain in the denominator. Even 100% structural coverage does not establish behavioral parity.

Answer separately: does the package/release exist; does it support the requested runtime/platform; does the application's used API and behavior work? Freshness and source evidence are attached to each answer. Abandonment/deprecation is a risk fact with dated evidence, not automatic incompatibility or permission to replace.

API differences compare the application's used symbols, signatures, types, exceptions, configuration, deprecations/removals and observable semantics between exact releases. List compatible, renamed, removed and unknown calls with their affected source spans, capability mappings and required adapters. Across languages, compare required capabilities and contracts rather than expecting symbol-to-symbol translation. Metadata alone cannot certify unchanged behavior.

## Machine decision records

v1 combined decision/status examples are historical only. Current dependency-resolution.schema v2 has immutable strategy/capability/decision refs; project/decision.schema records typed action/alternative/evidence, and migration/progress.schema tracks implementation separately. REIMPLEMENT states required used capability subset explicitly; partial versus full scope is evidenced, not inferred from package size. Proprietary/native/license/platform/production-only dependencies require acquisition/availability and used-behavior evidence or remain BLOCKED.

CLI decision propose/show and MCP decisions.propose/get accept or reject recommendations, select a strategy, supply an alternative, request evidence or propose a scope exception using action enum. These only propose. Gate 2 confirms frozen selected decisions; no progress change requests reapproval.

## User review and characterization

Show keep, upgrade, adapt, replace, reimplement and blocked groups, with exact source/requested/resolved versions, used capabilities, affected units, migration effort/risk, evidence and missing behavior. Every material strategy needs an explicit decision; OPTIONAL proposals remain unapplied without opt-in. Accepting a recommendation does not mutate the target. Changing the target triggers feasibility/replanning first.

Characterize dependency boundaries using source inputs and fixture state, including edge/error cases and side effects; then compare target outputs under the applicable [compatibility contract](../05-verification/COMPATIBILITY-SPEC.md). These use the existing case/trace/verdict machinery, not another verifier. Binary image/PDF changes need approved format comparators or remain MISMATCH/UNKNOWN. Also rerun affected end-to-end workflows: matching a method signature or one isolated output is insufficient. Build/runtime checks and performance evidence are separate from parity.

Required tests include a compatible package with changed behavior, dynamic-only usage, a removed used API, native-subset replacement, partial capability coverage, incompatible transitive replacement, cycles/conflicting constraints, unknown metadata, stale locks, missing characterization, and unresolved dependencies blocking only their dependent scope.

## Invalidation and cumulative knowledge

A target/runtime/package/strategy change recomputes affected graph constraints, API differences, feasibility, generated implementations and dependent verification/benchmarks. Source inventory, usage and characterization remain reusable when their source/fixture subjects are unchanged. Record reverse dependency edges; conservative invalidation is required when impact is unknown. Old decisions and approvals remain historical, never silently rewritten.

Project lessons capture source/target versions, capability predicates, adapter pattern, attempted alternatives, successes/failures, corpus/policy/environment hashes and evidence. Reusable patterns require the existing [learning promotion policy](../07-learning/LEARNING-SYSTEM.md), privacy review and opt-in human approval. Show observed project/case counts and failures; do not invent a confidence such as 0.94 from seven successes. A known transformation is a recommendation scoped to its evidenced capabilities, never authority to change a target or skip user approval/verification.

## Assess without resolving every dependency

Inventory/usage/capability findings are the same canonical records for assessment and migration. A dependency may remain unresolved/unsupported while the assessment report successfully documents current version/usage evidence, candidate retain/upgrade/adapter/replacement/reimplementation choices, risks, alternatives and missing data. Such recommendations are not approved resolutions. Assessment human recommendation selection records intended work; current exact decision refs and recursive feasibility remain prerequisites to Gate 2. Changes to a dependency decision refresh only affected assessed findings and downstream planning via typed edges; old report/decision evidence remains retrievable. See [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md).
