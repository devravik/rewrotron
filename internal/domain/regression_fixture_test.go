package domain

import (
	"testing"
)

// TestControlledRegression_LatencyDegradation verifies that an injected latency inflation
// is reliably detected by EvaluateStatisticalPerformance as REGRESSION without false pass.
func TestControlledRegression_LatencyDegradation(t *testing.T) {
	fixture := NewControlledRegressionFixture("reg-lat-1", RegressionLatencyInflation, 500, 10)
	trials := fixture.GeneratePairedTrials()

	if len(trials) != 10 {
		t.Fatalf("expected 10 trials, got %d", len(trials))
	}

	policy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
	})

	res, err := EvaluateStatisticalPerformance(policy, trials, 12345, "sha256:samples-reg-lat")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if res.Outcome != BenchmarkOutcomeRegression {
		t.Fatalf("expected REGRESSION outcome for latency inflation, got %s", res.Outcome)
	}

	p95Res := res.Metrics[MetricP95]
	if p95Res.Outcome != BenchmarkOutcomeRegression {
		t.Errorf("expected P95 outcome REGRESSION, got %s", p95Res.Outcome)
	}
	if *p95Res.LowerBound <= 1.0 {
		t.Errorf("expected lower bound strictly > 1.0 for regression, got %f", *p95Res.LowerBound)
	}
}

// TestControlledRegression_QueryInflation verifies that an injected N+1 query regression
// is detected by the DB_QUERY_COUNT and DB_LATENCY metric evaluations.
func TestControlledRegression_QueryInflation(t *testing.T) {
	fixture := NewControlledRegressionFixture("reg-query-1", RegressionQueryInflation, 500, 10)
	trials := fixture.GeneratePairedTrials()

	policy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricDBQueryCount,
			Unit:                     "count",
			LimitKind:                "RELATIVE_NON_INFERIORITY",
			Limit:                    1.05, // 5% margin
			MinimumSuccessesPerTrial: 500,
		},
		{
			MetricID:                 MetricDBLatency,
			Unit:                     "ms",
			LimitKind:                "RELATIVE_NON_INFERIORITY",
			Limit:                    1.10, // 10% margin
			MinimumSuccessesPerTrial: 500,
		},
	})

	res, err := EvaluateStatisticalPerformance(policy, trials, 9999, "sha256:samples-reg-q")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if res.Outcome != BenchmarkOutcomeRegression {
		t.Fatalf("expected REGRESSION outcome for query inflation, got %s", res.Outcome)
	}

	qRes := res.Metrics[MetricDBQueryCount]
	if qRes.Outcome != BenchmarkOutcomeRegression {
		t.Errorf("expected DB_QUERY_COUNT outcome REGRESSION, got %s", qRes.Outcome)
	}
}

// TestControlledRegression_NoFalsePass ensures that when a regression occurs, no early
// stopping or favorable subset manipulation can result in NON_REGRESSION / APPROVED_BUDGET.
func TestControlledRegression_NoFalsePass(t *testing.T) {
	fixture := NewControlledRegressionFixture("reg-composite-1", RegressionCompositeSlowdown, 500, 10)
	trials := fixture.GeneratePairedTrials()

	// 1. Full 10 trials under APPROVED_BUDGET with strict limit 1.0
	policy := makeTestPolicy("APPROVED_BUDGET", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "ABSOLUTE_BUDGET",
			Limit:                    30.0, // baseline is 25ms, mutated is 70ms
			MinimumSuccessesPerTrial: 500,
		},
	})

	res, err := EvaluateStatisticalPerformance(policy, trials, 42, "sha256:samples-budget-fail")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if res.Outcome != BenchmarkOutcomeRegression {
		t.Fatalf("expected REGRESSION, got %s", res.Outcome)
	}

	// 2. Truncated sample size (e.g. attempting to stop early before slowdown manifests)
	truncatedTrials := trials[:6]
	truncRes, err := EvaluateStatisticalPerformance(policy, truncatedTrials, 42, "sha256:samples-trunc")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if truncRes.Outcome != BenchmarkOutcomeInconclusive {
		t.Fatalf("expected INCONCLUSIVE for truncated trials, got %s", truncRes.Outcome)
	}
}
