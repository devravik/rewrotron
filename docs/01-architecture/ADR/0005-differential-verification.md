# ADR 0005: Independent differential verification

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [DIFFERENTIAL-TESTING](../../05-verification/DIFFERENTIAL-TESTING.md). This ADR records the decision, not a duplicate field specification.

## Context

Finite samples, self-generated oracles and probe assertions can falsely accept wrong targets.

## Problem

The prior attestation shape required installed instrumentation and tests even for UNAVAILABLE, while COMPLETE could contain PARTIAL mechanisms. Unqualified data models could select the wrong orders table across databases. Either permits fabricated evidence or blocks honest assessment.

## Decision

Use independent owned/reset environments and derive observation completeness from actual mechanism/control closure. Current attestations distinguish unavailable evidence from closed coverage; current data models are connection/catalog/schema-qualified. Relation-preserving comparisons, independent oracle review and mutants determine scoped acceptance. Assessment may complete with unavailable capture; verification cannot pass it.

## Alternatives considered

Shared writable replay state; probe-declared COMPLETE; isolated state plus mechanism controls.

## Why rejected

Shared state contaminates replay and a probe assertion cannot establish uncaptured PDO/native paths.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Attestation 2.1 supports absent instrumentation/control refs for UNAVAILABLE but requires pinned controls for CLOSED and all CLOSED for COMPLETE. Data-model 2.0 adds qualified identity/server evidence and typed destination relationships. Service validates causal/profile/control currentness and column arity; historical branches cannot satisfy current acceptance.

## Migration/compatibility implications

Old 2.0 attestations and unversioned models remain history. Re-extract/qualify rather than silently upgrade their meaning. No shared writable replay, no fabricated no-effect observation, no SQL migration rewrite.

## Acceptance experiment / falsification test

X05–X08/X11/X14/X20/X57: remove a write, use another database with the same table, bypass the framework with PDO and submit COMPLETE without control closure. Known wrong effects must mismatch; missing capture must UNKNOWN. Unavailable observation must still be representable in a completed assessment.
