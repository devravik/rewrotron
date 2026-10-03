# ADR 0002: SQLite metadata and local artifacts

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [SQLITE-SCHEMA](../../08-storage/SQLITE-SCHEMA.md). This ADR records the decision, not a duplicate field specification.

## Context

Trace volume and durability require bounded metadata and immutable byte closure.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Pinned modernc/sqlite/libc with required embedded engine identity, local WAL single-writer policy, artifact quotas and coordinated backup; no mandatory cloud/vector DB.

## Alternatives considered

All traces in SQLite; mandatory server DB; local metadata plus quota-governed artifacts.

## Why rejected

Trace blobs amplify WAL/backup growth; a mandatory server breaks baseline local operation.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X25/X28/X30. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
