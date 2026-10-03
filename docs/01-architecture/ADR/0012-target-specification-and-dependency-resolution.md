# ADR 0012: User intent, exact resolution and frozen decisions

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [TARGET-SPECIFICATION](../../04-migration/TARGET-SPECIFICATION.md). This ADR records the decision, not a duplicate field specification.

## Context

Wildcard intent cannot be execution environment; changing progress cannot rewrite approval.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Separate detected/asserted/requested/recommended/assessed/human-selected/resolved/approved/executed/verified records. Assessment completion and human work selection are not implementation approval; [PROJECT-LIFECYCLE](../../02-domain/PROJECT-LIFECYCLE.md) owns the checkpoint. Decisions immutable, progress separate; typed validity and conservative reuse preserve unrelated source work.

## Alternatives considered

Intent schema reused as resolution; combined changing decision/status; separate immutable choices and progress.

## Why rejected

Wildcards re-resolve over time; status mutation changes approval bytes and causes reapproval loops.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants. New assessment report/view/selection/planning-basis artifacts use existing immutable storage. Plan envelope 2.1 adds checkpoint refs, hashed through unchanged PLAN subject v2 decision_refs; old subjects/vectors remain reconstructable, not current executable authorization.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed. Old plan wrappers remain historical-readable; current execution requires a checkpoint-linked successor plan and its explicit Gate 2 approval. A separate assessment pipeline and implicit continuation were rejected because they duplicate analysis and conflate information with authorization.

## Acceptance experiment / falsification test

X15/X16/X17/X18/X29 and X41–X55 (especially selective assessment reuse and token rejection). Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
