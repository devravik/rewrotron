# ADR 0008: Pinned MCP stdio interface

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [MCP-SPEC](../../06-agent/MCP-SPEC.md). This ADR records the decision, not a duplicate field specification.

## Context

Transport existence does not prove client compatibility or worker survival.

## Problem

Logical schema URLs cannot imply an online schema service or assume every client dereferences external recursive refs. Stdio disconnect also cannot decide whether durable work survives.

## Decision

Keep pinned MCP2025-11-25 stdio and explicit engine supervision for long work. Advertise deterministic self-contained schema bundles from the pinned local registry, preserving recursive local refs. CLI remains the complete guarded fallback; named clients remain UNVERIFIED until actual integration sessions.

## Alternatives considered

Remote schema fetch; naive recursive inlining; offline local bundles. Foreground client-owned long tasks versus explicit supervisor.

## Why rejected

Remote fetch violates offline installation and cannot bind schema bytes; recursive inlining does not terminate for rule expressions. Client-owned tasks lose process ownership on disconnect. Remote transports add authentication/origin work and stay deferred.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

MCP-SPEC owns logical-ref resolution, deterministic $defs naming/rewrites, supported ref subset and message limits. Catalog schema IDs stay stable; advertised schema bytes are locally bundled. Domain failures consistently use ERR_UNSUPPORTED_CAPABILITY. No new business tool/server implementation is added.

## Migration/compatibility implications

Transport revision and CLI envelope are unchanged. Bundle incompatibility blocks advertisement rather than broadening a schema. Existing artifacts remain inspectable; client certification requires actual sessions after M5.

## Acceptance experiment / falsification test

X26/X31/X60: terminate the client while supervisor owns work, then reconnect from another client. With networking disabled validate recursive input/output schemas and the same negative mutants; absent registry refs, unsupported anchors and oversized bundles refuse advertisement.
