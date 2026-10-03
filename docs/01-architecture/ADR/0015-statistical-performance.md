# ADR 0015: Explicit statistical performance claims

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [PERFORMANCE](../../05-verification/PERFORMANCE.md). This ADR records the decision, not a duplicate field specification.

## Context

Strict equality under noise can remain inconclusive; arbitrary allowances weaken claims.

## Problem

An approved plan did not unambiguously freeze performance claim/method/limits. Ratio of means versus mean of ratios and pooled versus trial quantiles can change verdicts on the same data; equality should not be treated as proven regression.

## Decision

Bind an immutable typed performance policy in Gate 2, including evidenced not-applicable scope. Specify paired-trial estimands, bootstrap endpoints/error controls and deterministic outcome mapping in PERFORMANCE. Positive approved non-inferiority margins allow evidence-supported equality; strict improvement stays strict. Bounded measurement budgets prevent endless retries.

## Alternatives considered

Always strict improvement; automatic tolerance; explicit customer-approved statistical claim.

## Why rejected

Equal noisy systems may remain inconclusive under improvement; automatic tolerance silently changes the promised budget.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Performance-policy 1.0 names exact claim/profile/workload/environment/method/metrics/bounds; benchmark 2.1 references it. Native units and sample minimums are typed. Services verify PLAN closure and recompute statistics/outcome from raw samples; a producer CI/object or mutable budget is not authority.

## Migration/compatibility implications

Old benchmark shapes remain history, not current acceptance. Policy changes need successor plan approval and invalidate affected benchmarks/releases; progress does not mutate policy. No new SQL table or reinterpretation of old measurements.

## Acceptance experiment / falsification test

X33/X59: equal noisy values within approved margin can pass budget; strict equality cannot prove improvement. Zero denominator, insufficient p99, missing counters, omitted errors, pooled-quantile substitution and changed post-approval policy cannot pass. Known slowdown must fail against its declared limit.
