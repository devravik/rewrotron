# ADR 0009: Three IRs and explicit target model

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [TARGET-ARCHITECTURE](../../04-migration/TARGET-ARCHITECTURE.md). This ADR records the decision, not a duplicate field specification.

## Context

Behavior does not specify target component ownership/lowering/build capabilities.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Source/Application/Behavior IR plus lightweight target-specific implementation model proposed before Gate 2 and activated after; progress separate.

## Alternatives considered

Behavior IR directly to files; another universal source IR; target-specific ownership/lowering model.

## Why rejected

Direct generation hides architectural decisions; duplicating source semantics confuses source obligations and target choices.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X17/X22/X40. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
