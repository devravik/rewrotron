# Analysis contract

**Status:** Normative. Analysis requires a valid Gate 1 baseline.

## Outputs

1. Complete source file inventory with bytes/hashes, syntax status and positions.
2. Source IR symbols and typed dependency edges, including unresolved dynamic expressions.
3. Application IR entrypoints, handlers, shared components, data models and workflow candidates resolved by the source adapter.
4. Evidence-backed ACTIVE/DORMANT/DEAD/UNKNOWN classifications.
5. Findings, missing inputs, capability limitations and scope completeness.

Source parsing and framework interpretation are specified in [STATIC-ANALYSIS](STATIC-ANALYSIS.md) and [CI3-ADAPTER](../04-migration/CI3-ADAPTER.md). Generic graph logic must not include CI3 assumptions. Every output binds the confirmed source hash and adapter version.

## Safety and completeness

No application code, Composer scripts, route config, includes or reflection executes during static analysis. Analysis limits, unreadable files and parsing failures are explicit findings. They cannot silently disappear from graph roots or justify omission. Only the algorithm in [CODE-CLASSIFICATION](CODE-CLASSIFICATION.md) grants DEAD eligibility.

Inventory includes templates, assets, hook/bootstrap files, CLI scripts, cron/consumer candidates and supplied operational commands. Detection counts are preliminary; analyzed counts have exact definitions and provenance. Dynamic dependency observations add evidence but do not close all possible dispatch targets.

## Publication and reruns

Analysis operates on a frozen manifest/sandbox snapshot, stages IR artifacts, rechecks drift, then atomically publishes metadata and events. Interrupted or stale results remain historical partial artifacts and do not advance current project state. Identical hash/version inputs can reuse an existing immutable result; changed adapters or inputs create a new revision.

Required fixtures include inheritance/traits, dynamic methods/includes, conditional route config, parser errors, controller subdirectories, shared helpers, dead strongly connected subgraphs, cron-only roots, view/asset dependencies, dirty/untracked files and symlinks. Expected graph/classifications are version-controlled synthetic fixtures.

## Dependency usage evidence

The source inventory preserves exact manifest/installed/locked versions and direct/transitive/development/optional roles. After Gate 1, symbol reachability, operational roots and runtime observations identify required capability subsets and affected behavior units. Declared, reachable, observed and required counts remain distinct; dynamic or unobserved usage cannot justify automatic omission. Target selection never overwrites source dependency facts. [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) owns recursive target constraints, API/capability mapping and user decisions.
