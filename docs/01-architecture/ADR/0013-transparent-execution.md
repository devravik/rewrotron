# ADR 0013: Transparent execution and owned processes

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [EXECUTION-OBSERVABILITY](../../06-agent/EXECUTION-OBSERVABILITY.md). This ADR records the decision, not a duplicate field specification.

## Context

Current mutable rows cannot reconstruct old state; durable work needs a process owner.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Current state at N plus event window through M; foreground caller ownership or explicit local supervisor; resource ownership spans writes; heartbeats distinct from progress.

## Alternatives considered

Full event sourcing/historical snapshots; current state plus event window; client-owned detached work.

## Why rejected

Historical projection adds persistence requirements not present; detached client work lacks durable owner/cleanup. Full historical projections remain deferred.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X23/X26/X27. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
