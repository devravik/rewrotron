# ADR 0015: Fixed-N statistical claims and separate SLO windows

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../05-verification/PERFORMANCE.md) contains normative field/service details.

## Context

Mean trial quantiles and noisy equality need explicit approved claims.

## Problem

Stopping at the first favorable look biases acceptance; mean p99 can hide bad operational windows.

## Decision

Bind fixed N before measurement, retain every attempt/sample and recompute the specified paired bootstrap. Evaluate requested bounded per-window SLO independently. Equality can meet an approved positive non-inferiority margin, never strict improvement.

## Alternatives considered

Unspecified adaptive stopping; general sequential method now; fixed N and separate bounded-window obligation.

## Why rejected

Adaptive favorable looks inflate acceptance; qualified sequential statistics are unnecessary for M6 and remain deferred. A mean quantile is not an every-window SLO.

## Consequences

Missing trials/control evidence is INCONCLUSIVE; known window breach fails SLO even if statistical claim accepts. No universal production SLO guarantee.

## Machine-contract impact

Performance-policy 1.1 pins paired_trials/stopping_rule/operational_slo_ref; operational-slo 1.0; benchmark 2.2 records separate window result and raw refs.

## Migration/compatibility implications

Old benchmark/policy branches remain readable, never current acceptance. Changing claim/method/N/SLO requires successor plan approval; progress never rewrites policy.

## Acceptance experiment / falsification test

X33/X59/X78: equal noisy data under approved margin, early favorable looks, burst-tail windows, missing counters/errors and post-approval N changes are recomputed for the intended result.
