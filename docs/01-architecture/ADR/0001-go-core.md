# ADR 0001: Go deterministic core

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [SYSTEM-DESIGN](../SYSTEM-DESIGN.md). This ADR records the decision, not a duplicate field specification.

## Context

Core orchestration needs provider-independent execution and portable explicit ports.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

Go core without LLM clients; pinned storage dependencies require Go floor 1.25.0. CLI/MCP compose domain ports, adapters and repositories.

## Alternatives considered

Provider-owned orchestration; deterministic Go services with injected ports.

## Why rejected

Provider core would make state/acceptance depend on vendor availability and conversation.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed.

## Acceptance experiment / falsification test

M0/M1 deterministic LLM-free runs. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.
