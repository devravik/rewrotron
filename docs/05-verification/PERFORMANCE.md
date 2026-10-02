# PERFORMANCE.md — Benchmarking & Regression Guardrails

> **Version:** 1.0  
> **Status:** Verification Reference Guide  

---

## 1. Performance Non-Regression Guarantee

Modernization must never degrade application throughput or introduce latency regressions.

### Acceptance Rule:
A migrated `MigrationUnit` is rejected for production cutover if:
1. **P95 Latency Regression:** Target P95 response time is > 20% higher than legacy CI3 baseline under identical concurrency.
2. **N+1 Query Explosion:** Target executes more database queries than legacy implementation for the same scenario (e.g. accidental lazy loading in Eloquent).
3. **Memory Spike:** Peak PHP worker memory exceeds 2x legacy memory limit.

### Automated Profiling
`rewrotron benchmark --unit <unit_id>` executes a concurrent warm-up and measured load test, emitting side-by-side latency distributions, query counts, and memory allocations.
