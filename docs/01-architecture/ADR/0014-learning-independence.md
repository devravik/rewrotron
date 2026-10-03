# ADR 0014: Independent learning and reversible evidence consumption

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [LEARNING-SYSTEM](../../07-learning/LEARNING-SYSTEM.md). This ADR records the decision, not a duplicate field specification.

## Context

Repeated correlated oracles and degraded capture can promote false techniques.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Independent holdouts/mutants/quality denominators, sanitized opt-in promotion, pinned consumers and transitive retraction; global knowledge advisory.

## Alternatives considered

Promotion by context count/success rate; live registry consumption; independent pinned evidence/holdouts.

## Why rejected

Correlated oracles and degraded tools can inflate success; live revisions defeat reproducibility and transitive retraction.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X37/X38/X39. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
