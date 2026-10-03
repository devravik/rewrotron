# Performance measurement and non-regression

The product goal is non-regression in response latency and memory under the agreed workload. A tolerance for measurement noise is distinct from an approved relaxation of that goal. There is no automatic 20% latency or 2x memory allowance.

## Reproducible protocol

Pin environment images, CPU/memory quotas, runtime settings, DB state, dependencies, worker counts, workload mix, concurrency, payload sizes and cache state. Use five paired source/target trials with alternating order, 30 seconds warm-up and 60 seconds measured load per trial. Default concurrency is one for the first fixture; the project plan must include the intended customer load profile. Record successful/failed request counts, latency distributions, throughput, worker/container memory, query counts, and all measurement artifacts.

Verify equivalent behavior for the measured corpus before accepting performance. A target that skips authorization or DB writes cannot qualify as an optimization. Query-count increases trigger a diagnostic finding; they are not alone proof of worse performance.

## Acceptance and uncertainty

Compute each trial's target/source p95-latency and peak-worker-memory ratios. Report the raw ratios and a paired bootstrap 95% confidence interval (10,000 samples with a stored seed). Non-regression is demonstrated when the upper bound is at most 1.0 and errors/throughput satisfy the agreed workload. A lower bound above 1.0 is regression; an interval spanning 1.0 is inconclusive and requires more measurements or an explicit reviewed performance budget. Zero source denominators or missing memory instrumentation are inconclusive.

A customer-approved budget is recorded in the confirmed plan as an intentional deviation, never represented as the universal non-regression guarantee. No benchmark available means Gate 3 readiness is blocked for a project requiring that dimension. Benchmarks are tracked outside unit correctness results, with the same source/plan/target hashes.

Synthetic noisy samples, known regressions, warm/cold-cache effects, insufficient successes, leaked workers, and missing memory counters are required fixtures. Real load measurements remain environment validation work; documentation cannot certify performance in advance.
