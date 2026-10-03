# ADR 0010: Durable exact-subject human gates

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [CONFIRMATION-PROTOCOL](../../02-domain/CONFIRMATION-PROTOCOL.md). This ADR records the decision, not a duplicate field specification.

## Context

Implicit hash construction/registry lookups and incoherent scans break approval identity.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Versioned canonical SOURCE/PLAN/RELEASE subjects bind sealed snapshot/exact resolution/adapter content/choices; reconstruct original bytes; explicit human receipts/token consumption.

## Alternatives considered

Registry-derived hash; mutable YAML approval; retained versioned canonical subject.

## Why rejected

Registry changes make old preimages irreconstructable and edited YAML cannot establish human intent.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

X01/X02/X16. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
