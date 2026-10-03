# Performance claims and statistical acceptance

**Owner:** Workload, statistical claim and benchmark outcome. Applicable scope must have current parity before measurement. Performance cannot compensate for skipped behavior.

Gate 2 chooses IMPROVEMENT, NON_INFERIORITY or APPROVED_BUDGET with workload identity and per-metric thresholds. Strict improvement requires confidence interval upper bound below 1 for selected lower-is-better ratios. Non-inferiority requires upper bound ≤1+an explicitly approved positive margin. Approved budgets specify absolute/relative limits and failure rates. No automatic allowance exists; equal/noisy measurements under a strict zero margin may remain inconclusive. Stop after a bounded measurement budget and request a reviewed claim/budget change rather than retrying forever or claiming regression from uncertainty.

## Initial measurement method

Pin workload/fixtures/profile, environment bytes, CPU/memory quotas, runtime/workers, concurrency, cache/clock, target tree and comparator. At least 10 paired trials alternating side order; 30-second warm-up, at least 60-second measured block, max 20 paired trials per request. Resample paired trial-level metrics, not autocorrelated individual requests: 10,000 bootstrap replicates with stored seed and family-wise alpha 0.05 split across k required metrics. Publish raw samples and method ref. Unknown dependence/insufficient blocks makes confidence inconclusive. These are initial method requirements, not a claim all real workloads are statistically modeled.

Measure p50/p95/p99, CPU, peak/RSS memory, query count/DB latency, external latency, response size, startup and warm/cold behavior where required. Report missing dimensions explicitly, never zero. Initial p95/CPU/memory claim needs ≥500 successful requests per trial; required p99 needs ≥10,000 per trial or an approved alternative estimand/method. Query count is a diagnostic unless a budget explicitly requires it. Zero denominators use reviewed absolute budgets; missing instrumentation is INCONCLUSIVE.

Outcome vocabulary remains NON_REGRESSION (strict upper ≤1), REGRESSION (lower exceeds approved bound), INCONCLUSIVE, or APPROVED_BUDGET (chosen explicit non-inferiority/absolute budget satisfied). The report also names the claim, margin/budget decision, workload and method, so NON_REGRESSION is not confused with a relaxed allowance. Required rejected/inconclusive metrics block release until adequate evidence or successor approved claim; no conversion of correctness UNKNOWN.

Falsify with identical noisy samples, known slowdown, small samples, zero denominators, cache/order effects, failed requests and skipped writes. Real benchmarks are pending runtime experiments, not documentation validation.

Benchmark v2 metrics are typed by metric_id/unit/required status, source/target values, sample counts, confidence bounds, approved limit and raw sample refs. Missing values are null; incomplete measurements may return INCONCLUSIVE with fewer than ten trials. Accepted outcomes require at least ten paired trials and non-null required metric evidence; semantic checks validate metric IDs, workload-required set, method/counts and budget identity. Missing counters are never fabricated as zero.
