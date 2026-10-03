# ADR 0003: Executable behavior as canonical obligation

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [BEHAVIOR-MODEL](../../02-domain/BEHAVIOR-MODEL.md). This ADR records the decision, not a duplicate field specification.

## Context

Prose/open JSON cannot establish executable business rules or unseen behavior.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Three evidence-bearing IRs, v2 bounded tagged semantics/scenarios and explicit unsupported gaps; claims are corpus/profile bounded.

## Alternatives considered

Prose/open JSON evaluator; unrestricted DSL; small closed typed vocabulary.

## Why rejected

Prose has no deterministic interpreter and unrestricted evaluation adds unsafe unbounded semantics.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X03/X04/X05/X06/X08. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
