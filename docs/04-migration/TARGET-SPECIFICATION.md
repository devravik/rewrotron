# Target environment specification and configuration authority

**Status:** Normative design specification; engine implementation pending. The target is user-owned data. Recommendations propose choices; they never rewrite those choices. See [ADR 0012](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md).

## Four separate records

| Record | Meaning | Authority |
| --- | --- | --- |
| Source detection | What was observed, with confidence, evidence and unknowns | Immutable detection proposal; Gate 1 confirms a merged source baseline |
| Target specification | What the user wants to run and preserve | Editable intent, captured as an immutable revision for planning; Gate 2 approves the exact revision with the plan |
| Migration plan | What will be changed to satisfy that target and the behavior contracts | Analysis-backed immutable proposal; executable only after Gate 2 |
| Execution result | What was generated and actually verified | Current build, differential and performance evidence; Gate 3 approves release packaging |

The decision flow is **detect → recommend → user configures → freeze target revision → feasibility and migration plan → review/confirm → execute**. Gate 1 still precedes deep usage analysis and behavior discovery. A bounded recommendation based on metadata can precede it, but cannot claim complete dependency usage. Freezing a planning input is not human approval. Gate 2 confirms the frozen target, dependency decisions and plan together; no fourth lifecycle gate is introduced.

## Human-editable files and generated views

`rewrotron.yaml` at the operational workspace root is the primary human-editable project configuration. The workspace may be separate from the legacy source. It validates against [project.schema.json](../../specs/project/project.schema.json). Users may move its `target` object to `.rewrotron/target.yaml`, which validates against [target.schema.json](../../specs/project/target.schema.json); the two files must not both define a target. Duplicate target definitions fail validation rather than merging silently.

```text
workspace/
├── rewrotron.yaml                  # User-owned project intent; target may be inline
└── .rewrotron/
    ├── project.yaml                # Project identity/settings view; legacy input only if no root file
    ├── target.yaml                 # Optional user-owned split target
    ├── detection.yaml              # Generated observations; never overwrites intent
    ├── recommendations.yaml        # Generated evidence-backed proposals/decision views
    ├── source-dependencies.yaml    # Generated source inventory and usage graph view
    ├── dependencies.yaml           # Generated target resolution/capability mappings
    ├── plan.yaml                   # Generated current immutable plan view
    ├── approvals.yaml              # Generated SQLite confirmation view; no gate authority
    ├── project.db                  # Authoritative operational state and confirmations
    ├── mappings/
    ├── contracts/
    ├── traces/
    └── artifacts/sha256/           # Canonical immutable artifact bytes
```

Existing `.rewrotron/project.yaml` inputs remain readable when no root file exists. When `rewrotron.yaml` exists, the project view is never another configuration input; report that provenance explicitly. Moving configuration is an explicit operation, never an automatic overwrite. Generated views are reproducible exports, not independent ledgers. Editing `approvals.yaml`, detection or a plan view grants no authority and does not alter registered evidence. Target edits are imported as new intent revisions only after validation and drift checks. Duplicate YAML keys, unknown fields, unresolved aliases, unsafe paths and raw credentials in secret-bearing fields are rejected before hashing. Secret references use `env:NAME`.

## Canonical target shape

Keep the existing flat language/framework fields and `runtime.version` as the language runtime constraint. Do not maintain competing keys such as `php`, `runtime.php`, `language.version` and `framework_version`. Nested sketches are explanatory, not accepted aliases. Language and framework identifiers are open strings; registered adapter capability checks determine supported combinations.

```yaml
version: 1
project_id: customer-portal
source:
  type: auto
  path: ../legacy-ci3
environment:
  mode: docker
safety:
  production_access: denied
  external_writes: denied
  export_policy: local_only
target:
  language: php
  framework: laravel
  version: "13.*"
  runtime:
    version: "8.3"
    os: linux
    architecture: amd64
  path: ../customer-portal-modern
  web:
    server: nginx
  database:
    engine: mysql
    version: "8.*"
    preserve_schema: true
  database_policy: preserve_existing
  cache:
    engine: redis
  queue:
    engine: redis
  storage:
    local: true
  package_manager:
    name: composer
  architecture:
    style: modular_monolith
  compatibility:
    http: strict
    api: strict
    auth: strict
    database_behavior: strict
migration:
  strategy: vertical_slice
  remove_dead_code: true
  preserve_dormant_code: true
  optimization:
    enabled_after_parity: true
```

Version constraints express user intent, not exact installed versions. The selected ecosystem adapter interprets constraints and records exact framework/runtime/package versions, lock digests, platform and image identity before execution. For example, `8.3` requests the PHP 8.3 runtime line, never permission to use 8.1 or 8.4. Patch selection within an approved constraint is resolved and shown in the plan; a changed resolution/lock changes the plan subject. An incompatible framework/runtime pair creates a blocker; the engine cannot downgrade it. Go, Rust, Python and other targets use the same fields, with their own adapter constraint semantics and explicit support limits.

The target covers runtime/OS/CPU, web server, database engine/version/schema policy, cache, queue, storage, package manager, architecture and per-dimension compatibility. Missing required fields are requested, not guessed from the source. First-release schema preservation remains a hard constraint. `remove_dead_code` permits only evidenced DEAD omissions; DORMANT and UNKNOWN remain preserved or explicitly scoped out in a reviewed plan. Optimization waits for parity. Target compatibility selections do not themselves authorize behavior exceptions; exceptions still require policy evidence and Gate 2 review.

## Precedence and changing an approved choice

1. Hard engine constraints validate every configuration and may refuse execution.
2. The currently Gate 2-confirmed target is the execution specification.
3. Explicit user edits/overrides form a candidate successor specification.
4. Project intent provides remaining requested settings.
5. Detected facts inform source understanding and proposals only.
6. Recommendations propose unselected target choices.
7. Defaults are visible proposals requiring review.

An explicit edit supersedes a prior choice only through a newly reviewed plan. It never changes ongoing execution under an old approval. Recommendations and defaults cannot fill missing target fields invisibly or override requested versions. Source overrides remain in the Gate 1 baseline without overwriting the original detection. Pending intent that differs from the approved target pauses new target work until replanning/confirmation; in-flight results cannot publish against the superseded subject.

## Recommendations and feasibility

Every recommendation names ID, category, source/current value, requested target, proposed action/version, reasons, risk, affected dependencies/files/units/capabilities, evidence refs, required actions, alternatives and decision state. API changes must state which used calls and behaviors change. The canonical severity vocabulary is separate from generic finding risk levels:

| Type | Engine treatment |
| --- | --- |
| `BLOCKING` | No execution for dependent in-scope units until resolved; approval alone cannot waive technical impossibility |
| `REQUIRED` | User must accept the required action or change target/scope and recompute before affected execution |
| `RECOMMENDED` | Requires an explicit accept/reject/defer decision during plan review; never applied from silence |
| `OPTIONAL` | Requires explicit opt-in; an undecided item remains unapplied |
| `INFORMATIONAL` | Presented without authorizing a change |

Decisions are `PENDING`, `ACCEPTED`, `REJECTED`, `DEFERRED`, or `SUPERSEDED`, bound to recommendation ID/revision and target subject. A human Gate 2 confirmation may accept individually enumerated decisions in the reviewed plan. A blanket unattended “accept all” cannot supply human authorization. Rejecting an action that satisfies a hard requirement preserves its blocker until an evidenced alternative or reviewed scope change resolves it.

Feasibility checks the requested runtime/framework/platform intersection, recursively resolved dependency graph, required capabilities, API/legacy transformations, infrastructure and adapter support. Report compatible/upgradable/replaceable/unknown counts with denominators, unresolved decisions, evidence freshness, risk and `can_proceed` for an explicit scope. Metadata compatibility, capability coverage and behavioral parity are separate results. Unknown required compatibility blocks affected units. A feasible proposal is not a verification PASS.

The human review shows detected, recommended, requested and exact resolved versions side by side; dependency upgrades/replacements; required transformations; risk/evidence; DEAD proofs and unknown/dormant components; decisions and affected scope. A conflicting choice is explained with alternatives, never silently repaired by editing the target. Existing `plan show` and report operations carry this review; a future interactive `configure` screen is an implementation extension, not a command currently in the registry.

## Hashes and selective invalidation

`target_spec_hash` is the canonical JSON hash of the validated effective target intent, explicit decision-bearing migration settings and provenance of overrides. Formatting/comments do not change it. The immutable plan embeds that specification plus exact resolutions and decision artifacts. Gate 2 binds all of them through `plan_hash`. See [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) for canonicalization.

Changing PHP 8.3 to 8.4 invalidates the current Gate 2/3 subject, affected feasibility/resolution records, generated build/runtime evidence and dependent verification/benchmarks. It preserves Gate 1, Source/Application/Behavior IR, route discovery, source usage and characterization evidence when their inputs are unchanged. Unaffected units/artifacts retain history and can be reused after their recorded prerequisites are checked against the new plan. Reuse is evidence-backed, never relabeling stale results as current PASS. Shared/transitive dependencies propagate invalidation; unknown impact conservatively invalidates affected target scope, not unrelated source analysis.

A project already at `MODELED` or later returns to `MODELED` while a replacement plan is prepared, then `PLAN_SYNTHESIZED` pending Gate 2; earlier projects retain their source lifecycle state; retained unit work is tracked separately. Adding/editing the workspace's own `rewrotron.yaml` is target/project intent drift, excluded from the source manifest when co-located. Legacy application configuration is still source material. Actual source/environment drift follows Gate 1 re-detection rules. No revision change deletes completed history or restarts the project wholesale.

## Persistence and verification requirements

Frozen target, recommendations, feasibility and dependency decisions are immutable registered artifacts linked through existing `artifacts`/`artifact_links`, plan documents and events; open blockers use `findings`. No new SQL tables are required by this documentation change. Approval authority remains append-only `user_confirmations`; resume checks current hashes rather than trusting an exported `confirmed` flag.

Required implementation fixtures cover user choice over recommendation, impossible target refusal, duplicate target definitions, exact resolution outside a requested constraint, YAML approval spoofing, changed target during generation, identical semantic YAML, target-only edits retaining Gate 1/source IR, transitive invalidation and reuse without stale PASS. Schemas validate structure; future engine fixtures must prove these semantic rules.
