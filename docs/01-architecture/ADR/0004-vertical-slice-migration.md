# ADR 0004: Vertical slices and scoped acceptance

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [MIGRATION-UNITS](../../04-migration/MIGRATION-UNITS.md). This ADR records the decision, not a duplicate field specification.

## Context

Whole-corpus unit completion deadlocks dependent behaviors.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

UNIT plus applicable IMPLEMENTED_REGRESSION acceptance completes a prerequisite; PROJECT acceptance remains full approved scope. Shared components have explicit ownership.

## Alternatives considered

Whole corpus at every unit; unit-only checks; unit plus implemented regressions.

## Why rejected

Whole corpus deadlocks prerequisites; unit-only checks miss shared-component regressions.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X22/X35. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
