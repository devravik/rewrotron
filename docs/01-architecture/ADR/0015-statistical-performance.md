# ADR 0015: Explicit statistical performance claims

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [PERFORMANCE](../../05-verification/PERFORMANCE.md). This ADR records the decision, not a duplicate field specification.

## Context

Strict equality under noise can remain inconclusive; arbitrary allowances weaken claims.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Choose improvement/non-inferiority/approved budget explicitly with paired block methodology, sample requirements, bounded retries and required evidence.

## Alternatives considered

Always strict improvement; automatic tolerance; explicit customer-approved statistical claim.

## Why rejected

Equal noisy systems may remain inconclusive under improvement; automatic tolerance silently changes the promised budget.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X33. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
