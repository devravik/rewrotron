# ADR 0009: Three IRs and explicit target model

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owners: [IR-CONTRACTS](../../02-domain/IR-CONTRACTS.md) for neutral records, [TARGET-ARCHITECTURE](../../04-migration/TARGET-ARCHITECTURE.md) for lowering and [SERVICE-CONTRACTS](../SERVICE-CONTRACTS.md) for qualification. This ADR records the decision, not a duplicate field specification.

## Context

Behavior does not specify target component ownership/lowering/build capabilities.

## Problem

Open supported_profiles metadata could silently become admission policy, and unchanged schema URIs could hide changed extraction semantics. A broad adapter can lose declarations not represented by the current neutral facets.

## Decision

Keep three evidence-bearing neutral IRs and target-specific lowering model. Current adapters use exact implementation/schema/profile qualification, not metadata inference. Initial supported subset refuses unrepresented declarations. A generic executable constraint language and wider neutral facets are deferred prerequisites of M8.

## Alternatives considered

Behavior IR directly to files; another universal source IR; target-specific ownership/lowering model. For support admission: trust open manifest metadata, design a universal constraint DSL immediately, or qualify exact pinned profiles first.

## Why rejected

Direct generation hides architectural decisions; duplicating source semantics confuses source obligations and target choices. Open metadata has no deterministic support predicate; a universal profile DSL delays the first falsifiable experiment without evidence that its abstractions preserve real adapter constraints. Exact-profile qualification is the narrow initial boundary.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Adapter manifest 2.1 pins schema_refs and qualification_refs; qualification.schema identifies exact implementation/profile/control/result bytes and status/limitations. SERVICE-CONTRACTS derives admission from trusted current controls. Data-model 2.0 qualifies DB identity without introducing a framework category in core.

## Migration/compatibility implications

Unversioned/2.0 manifests and old IR children remain descriptive/history. Existing approvals retain exact schemas; current admission does not broaden ranges from a newer registry. General profile DSL and language facets need versioned contracts/mutants before expansion.

## Acceptance experiment / falsification test

X40/X58: change schema bytes under a stable URI, advertise an untested framework/runtime profile, and submit a declaration that cannot survive neutral extraction. Refuse current execution/closure without destroying historical hashes or dropping uncertainty.
