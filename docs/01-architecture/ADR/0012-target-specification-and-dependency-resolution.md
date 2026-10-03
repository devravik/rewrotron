# ADR 0012: Immutable intent and positive reuse coverage

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../02-domain/DOMAIN-MODEL.md) contains normative field/service details.

## Context

Intent, resolution, immutable decisions and mutable progress already have separate lifecycles.

## Problem

An incomplete graph cannot prove independence; ambiguous shared-consumer wording can deadlock future units.

## Decision

Keep immutable decisions and shared assessment. Require positive consumption coverage for reuse; reverify currently implemented consumers, record future pending obligations. Provider interprets opaque exact release labels; user intent remains sovereign.

## Alternatives considered

Trust absent edges; invalidate every artifact always; positive coverage with conservative fallback.

## Why rejected

Absence is not independence; unconditional restart discards demonstrably unrelated work. Coverage allows bounded selective reuse without weakening acceptance.

## Consequences

Unknown/unexamined prerequisites invalidate relevant categories. No successor-plan reuse of old PASS; verification reruns.

## Machine-contract impact

consumption-coverage 1.0, validity/reuse 2.1, resolution 2.1, detection 2.0 without current gate_status. Existing artifacts/typed links remain storage owners.

## Migration/compatibility implications

Older coverage-free records are historical-only for current selective reuse. Progress updates do not mutate decisions or plan hash; shared assessment stays one workflow.

## Acceptance experiment / falsification test

X17/X29/X35/X73/X75 and X41–X55: drop a consumed edge, change target, preserve unaffected facts only with coverage, and avoid future-consumer deadlock.
