# ADR 0006: Vendor-neutral agent services

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [AGENT-MODEL](../../06-agent/AGENT-MODEL.md). This ADR records the decision, not a duplicate field specification.

## Context

Agent narration and client permissions cannot establish truth or human intent.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

All interfaces call guarded engine services; agents propose. Same-account trust is explicit; no hostile-agent OS isolation claim.

## Alternatives considered

Client prompt enforcement; same-account mediated services; separate hostile-agent OS identities.

## Why rejected

Prompts do not enforce guards; OS identities are a different deferred operating profile, not an implicit guarantee.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X25/X31/X32. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
