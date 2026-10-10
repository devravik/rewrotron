package domain

import (
	"testing"
)

// Helper to construct a standard valid 1.1 PerformancePolicyRecord for tests
func makeTestPolicy(claim string, nTrials int, metricLimits []PerformanceMetricLimit) *PerformancePolicyRecord {
	measProfile := "RWT_PAIRED_BOOTSTRAP_1"
	stoppingRule := "FIXED_N"
	workloadRef := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	envRef := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	methodRef := "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	return &PerformancePolicyRecord{
		SchemaVersion:      "1.1",
		PolicyID:           "test-policy-1",
		ProjectID:          "proj-1",
		Applicable:         true,
		Claim:              &claim,
		WorkloadRef:        &workloadRef,
		EnvironmentRef:     &envRef,
		MethodRef:          &methodRef,
		MeasurementProfile: &measProfile,
		RequiredMetrics:    metricLimits,
		MaxErrorRate:       0.01,
		Rationale:          "statistical performance test suite",
		EvidenceRefs:       []string{"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"},
		PairedTrials:       &nTrials,
		StoppingRule:       &stoppingRule,
	}
}

// Helper to generate N paired trials with controlled values
func makeTestTrials(
	n int,
	sourceP95 float64,
	targetP95 float64,
	sourceSamples int,
	targetSamples int,
	sourceErrors int,
	targetErrors int,
) []PairedTrial {
	trials := make([]PairedTrial, n)
	for i := 0; i < n; i++ {
		trials[i] = PairedTrial{
			TrialIndex: i,
			SourceMetricValues: map[string]float64{
				MetricP95: sourceP95,
			},
			TargetMetricValues: map[string]float64{
				MetricP95: targetP95,
			},
			SourceSuccesses: sourceSamples,
			TargetSuccesses: targetSamples,
			SourceErrors:    sourceErrors,
			TargetErrors:    targetErrors,
		}
	}
	return trials
}

// Test X78: Fixed stopping rule and optional window SLO
// "Policy fixes N and optional per-window SLO; retained raw samples.
//  Stop at favorable early N or hide bad windows in mean p99 =>
//  Incomplete-N INCONCLUSIVE; evaluate only N; violating window SLO FAIL."
func TestStatisticalPerformance_X78_FixedStoppingRule(t *testing.T) {
	policy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
	})

	// Early look at N=5 (e.g. favorable early stopping attempt)
	earlyTrials := makeTestTrials(5, 100.0, 80.0, 500, 500, 0, 0)
	earlyRes, err := EvaluateStatisticalPerformance(policy, earlyTrials, 42, "sha256:samples1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if earlyRes.Outcome != BenchmarkOutcomeInconclusive {
		t.Fatalf("expected INCONCLUSIVE for incomplete N=5, got %s", earlyRes.Outcome)
	}

	// Full N=10 with regression: source=100ms, target=120ms
	fullRegressionTrials := makeTestTrials(10, 100.0, 120.0, 500, 500, 0, 0)
	regRes, err := EvaluateStatisticalPerformance(policy, fullRegressionTrials, 42, "sha256:samples2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if regRes.Outcome != BenchmarkOutcomeRegression {
		t.Fatalf("expected REGRESSION at N=10, got %s", regRes.Outcome)
	}

	// Full N=10 with genuine strict improvement: source=100ms, target=80ms
	fullImprovementTrials := makeTestTrials(10, 100.0, 80.0, 500, 500, 0, 0)
	impRes, err := EvaluateStatisticalPerformance(policy, fullImprovementTrials, 42, "sha256:samples3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if impRes.Outcome != BenchmarkOutcomeNonRegression {
		t.Fatalf("expected NON_REGRESSION at N=10, got %s", impRes.Outcome)
	}
}

// Test X33: Noisy performance, insufficient samples, zero denominator, known slowdown
// "Identical noisy trials, insufficient p99 samples, zero denominator, known slowdown
//  and warm/cold order effects yield correct explicit claim/inconclusive outcome"
func TestStatisticalPerformance_X33_FalsificationScenarios(t *testing.T) {
	// Scenario 1: Insufficient samples for P99 (requires at least 10,000 successful requests per trial)
	p99Policy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP99,
			Unit:                     "ms",
			LimitKind:                "RELATIVE_NON_INFERIORITY",
			Limit:                    1.10,
			MinimumSuccessesPerTrial: 5000, // Policy specifies 5000, but P99 has strict mandatory lower bound of 10,000
		},
	})
	// Provide only 5000 samples per trial
	trialsInsufficientP99 := make([]PairedTrial, 10)
	for i := 0; i < 10; i++ {
		trialsInsufficientP99[i] = PairedTrial{
			TrialIndex: i,
			SourceMetricValues: map[string]float64{
				MetricP99: 100.0,
			},
			TargetMetricValues: map[string]float64{
				MetricP99: 102.0,
			},
			SourceSuccesses: 5000,
			TargetSuccesses: 5000,
		}
	}
	resP99, err := EvaluateStatisticalPerformance(p99Policy, trialsInsufficientP99, 42, "sha256:s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resP99.Outcome != BenchmarkOutcomeInconclusive {
		t.Fatalf("expected INCONCLUSIVE for insufficient P99 samples, got %s", resP99.Outcome)
	}

	// Scenario 2: Zero denominator in relative claim
	zeroPolicy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "RELATIVE_NON_INFERIORITY",
			Limit:                    1.10,
			MinimumSuccessesPerTrial: 500,
		},
	})
	zeroTrials := makeTestTrials(10, 0.0, 50.0, 500, 500, 0, 0)
	resZero, err := EvaluateStatisticalPerformance(zeroPolicy, zeroTrials, 42, "sha256:s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resZero.Outcome != BenchmarkOutcomeInconclusive {
		t.Fatalf("expected INCONCLUSIVE for zero source denominator, got %s", resZero.Outcome)
	}

	// Scenario 3: Identical noisy trials under STRICT IMPROVEMENT
	// Equal values yield interval centered around 1.0; under strict improvement, [1,1] or crossing 1.0 is INCONCLUSIVE
	strictPolicy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
	})
	identicalTrials := makeTestTrials(10, 100.0, 100.0, 500, 500, 0, 0)
	resIdentical, err := EvaluateStatisticalPerformance(strictPolicy, identicalTrials, 42, "sha256:s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resIdentical.Outcome != BenchmarkOutcomeInconclusive {
		t.Fatalf("expected INCONCLUSIVE for identical trials under IMPROVEMENT, got %s", resIdentical.Outcome)
	}

	// Scenario 4: Identical noisy trials under NON_INFERIORITY with approved margin (e.g. limit 1.05)
	// Equal values (ratio 1.0) with margin 1.05: upper bound <= 1.05 => APPROVED_BUDGET
	nonInfPolicy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "RELATIVE_NON_INFERIORITY",
			Limit:                    1.05,
			MinimumSuccessesPerTrial: 500,
		},
	})
	resNonInf, err := EvaluateStatisticalPerformance(nonInfPolicy, identicalTrials, 42, "sha256:s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resNonInf.Outcome != BenchmarkOutcomeApprovedBudget {
		t.Fatalf("expected APPROVED_BUDGET for identical trials under NON_INFERIORITY margin 1.05, got %s", resNonInf.Outcome)
	}

	// Scenario 5: Exceeded max error rate
	highErrorTrials := makeTestTrials(10, 100.0, 90.0, 500, 500, 20, 20) // 20 errors / 520 = ~3.8% > 1%
	resHighError, err := EvaluateStatisticalPerformance(nonInfPolicy, highErrorTrials, 42, "sha256:s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resHighError.Outcome != BenchmarkOutcomeInconclusive {
		t.Fatalf("expected INCONCLUSIVE for exceeded error rate, got %s", resHighError.Outcome)
	}
}

// Test X59: Estimand integrity and recomputed outcomes from raw samples
// "Change policy/workload/method after Gate 2, omit required counter, conflate pooled percentile
//  with mean of trial quantiles, fabricate CI or error omission; service must reject/inconclusive,
//  equal data within approved positive margin can accept"
func TestStatisticalPerformance_X59_RecomputedEstimands(t *testing.T) {
	policy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "RELATIVE_NON_INFERIORITY",
			Limit:                    1.10,
			MinimumSuccessesPerTrial: 500,
		},
	})

	// Trials with slight noise but target safely faster: source around 100, target around 90
	trials := make([]PairedTrial, 10)
	for i := 0; i < 10; i++ {
		trials[i] = PairedTrial{
			TrialIndex: i,
			SourceMetricValues: map[string]float64{
				MetricP95: 100.0 + float64(i%3),
			},
			TargetMetricValues: map[string]float64{
				MetricP95: 90.0 + float64(i%2),
			},
			SourceSuccesses: 500,
			TargetSuccesses: 500,
			SourceErrors:    0,
			TargetErrors:    0,
		}
	}

	// Recompute independently using seed 123
	res, err := EvaluateStatisticalPerformance(policy, trials, 123, "sha256:raw1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Outcome != BenchmarkOutcomeApprovedBudget {
		t.Fatalf("expected APPROVED_BUDGET, got %s", res.Outcome)
	}
	p95Res, ok := res.Metrics[MetricP95]
	if !ok {
		t.Fatalf("missing P95 metric in evaluation result")
	}
	if p95Res.LowerBound == nil || p95Res.UpperBound == nil {
		t.Fatalf("missing computed confidence bounds")
	}
	if *p95Res.UpperBound > 1.10 {
		t.Fatalf("expected upper bound <= 1.10, got %f", *p95Res.UpperBound)
	}

	// Verify deterministic reproducibility: same seed gives identical bounds
	res2, err := EvaluateStatisticalPerformance(policy, trials, 123, "sha256:raw1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p95Res2 := res2.Metrics[MetricP95]
	if *p95Res.LowerBound != *p95Res2.LowerBound || *p95Res.UpperBound != *p95Res2.UpperBound {
		t.Fatalf("deterministic PRNG failed reproducibility: %f vs %f", *p95Res.LowerBound, *p95Res2.LowerBound)
	}
}

// Test Operational SLO evaluation per operational-slo.schema.json & PERFORMANCE.md
func TestOperationalSLO_Evaluation(t *testing.T) {
	sloPolicy := &OperationalSLOPolicy{
		SchemaVersion:            "1.0",
		ProjectID:                "proj-1",
		WorkloadRef:              "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		WindowMs:                 10000,
		MinimumRequestsPerWindow: 100,
		LatencyPercentile:        0.99,
		MaximumLatencyMs:         "250.0",
		MaximumErrorFraction:     0.01,
		Scope:                    "EVERY_COMPLETE_MEASURED_WINDOW",
		EvidenceRefs:             []string{"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	}

	// Case 1: All windows pass
	goodWindows := make([]WindowSample, 5)
	for i := 0; i < 5; i++ {
		reqs := make([]float64, 100)
		for j := 0; j < 100; j++ {
			reqs[j] = 150.0 + float64(j%20) // max ~170 ms < 250 ms
		}
		goodWindows[i] = WindowSample{
			WindowIndex:        i,
			TrialIndex:         0,
			StartTimeMs:        int64(i * 10000),
			EndTimeMs:          int64((i + 1) * 10000),
			SuccessfulRequests: reqs,
			TotalAttempts:      100,
			ErrorCount:         0,
		}
	}
	resGood, err := EvaluateOperationalSLO(sloPolicy, goodWindows, "sha256:pol", "sha256:w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resGood.Outcome != SLOOutcomePass {
		t.Fatalf("expected SLO PASS, got %s", resGood.Outcome)
	}

	// Case 2: One window violates latency bound even if others pass => FAIL
	violatingLatencyWindows := make([]WindowSample, 5)
	copy(violatingLatencyWindows, goodWindows)
	badReqs := make([]float64, 100)
	for j := 0; j < 100; j++ {
		badReqs[j] = 200.0
	}
	badReqs[98] = 300.0 // p99 rank 99 exceeds 250 ms
	badReqs[99] = 300.0
	violatingLatencyWindows[2] = WindowSample{
		WindowIndex:        2,
		TrialIndex:         0,
		StartTimeMs:        20000,
		EndTimeMs:          30000,
		SuccessfulRequests: badReqs,
		TotalAttempts:      100,
		ErrorCount:         0,
	}
	resBadLatency, err := EvaluateOperationalSLO(sloPolicy, violatingLatencyWindows, "sha256:pol", "sha256:w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resBadLatency.Outcome != SLOOutcomeFail {
		t.Fatalf("expected SLO FAIL on latency breach, got %s (reasons: %v)", resBadLatency.Outcome, resBadLatency.ReasonCodes)
	}

	// Case 3: Error fraction exceeded in one window => FAIL
	violatingErrorWindows := make([]WindowSample, 5)
	copy(violatingErrorWindows, goodWindows)
	violatingErrorWindows[1] = WindowSample{
		WindowIndex:        1,
		TrialIndex:         0,
		StartTimeMs:        10000,
		EndTimeMs:          20000,
		SuccessfulRequests: goodWindows[1].SuccessfulRequests,
		TotalAttempts:      100,
		ErrorCount:         5, // 5/100 = 5% > 1%
	}
	resBadError, err := EvaluateOperationalSLO(sloPolicy, violatingErrorWindows, "sha256:pol", "sha256:w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resBadError.Outcome != SLOOutcomeFail {
		t.Fatalf("expected SLO FAIL on error fraction breach, got %s", resBadError.Outcome)
	}

	// Case 4: Zero requests in window => UNKNOWN
	zeroReqWindows := make([]WindowSample, 2)
	zeroReqWindows[0] = goodWindows[0]
	zeroReqWindows[1] = WindowSample{
		WindowIndex:        1,
		TrialIndex:         0,
		StartTimeMs:        10000,
		EndTimeMs:          20000,
		SuccessfulRequests: []float64{},
		TotalAttempts:      0,
		ErrorCount:         0,
	}
	resZeroReq, err := EvaluateOperationalSLO(sloPolicy, zeroReqWindows, "sha256:pol", "sha256:w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resZeroReq.Outcome != SLOOutcomeUnknown {
		t.Fatalf("expected SLO UNKNOWN on zero-request window, got %s", resZeroReq.Outcome)
	}
}

// Test Not-Applicable policy handling per PERFORMANCE.md
func TestStatisticalPerformance_NotApplicablePolicy(t *testing.T) {
	naPolicy := &PerformancePolicyRecord{
		SchemaVersion: "1.1",
		PolicyID:      "na-policy",
		ProjectID:     "proj-1",
		Applicable:    false,
		Rationale:     "static-only scope without workload requirements",
		EvidenceRefs:  []string{"sha256:evidence1"},
	}

	res, err := EvaluateStatisticalPerformance(naPolicy, nil, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Outcome != BenchmarkOutcomeApprovedBudget {
		t.Fatalf("expected APPROVED_BUDGET for not-applicable policy release scope, got %s", res.Outcome)
	}
	if len(res.ReasonCodes) == 0 || res.ReasonCodes[0] != "POLICY_NOT_APPLICABLE" {
		t.Fatalf("expected POLICY_NOT_APPLICABLE reason code, got %v", res.ReasonCodes)
	}
}
