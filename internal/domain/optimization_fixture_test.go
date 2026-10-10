package domain

import (
	"testing"
)

// TestControlledOptimization_ValidOptimizationAccepted verifies that a valid optimization
// (e.g. batching queries from 6 -> 2, reducing latency from 50ms -> 18ms) that preserves
// 100% behavioral parity is accepted with NON_REGRESSION (strict improvement).
func TestControlledOptimization_ValidOptimizationAccepted(t *testing.T) {
	candidate := OptimizationCandidate{
		CandidateID:          "opt-eager-load-1",
		Description:          "Eliminate N+1 query loop with eager loading",
		TargetComponent:      "catalog_controller",
		OptimizedDBQueries:   2,
		OptimizedLatencyMs:   18.0,
		OptimizedResponseLen: 2048,
		IntroducesMismatch:   false,
	}

	fixture := NewControlledOptimizationFixture("opt-fix-1", candidate, 500, 10)

	policy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
		{
			MetricID:                 MetricDBQueryCount,
			Unit:                     "count",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
	})

	decision, err := EvaluateOptimization(fixture, policy, 42, "sha256:samples-opt-valid")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if !decision.Accepted {
		t.Fatalf("expected optimization to be accepted, but got rejected: %s", decision.RejectionReason)
	}
	if decision.BehaviorVerdict != "PASS" {
		t.Errorf("expected behavior verdict PASS, got %s", decision.BehaviorVerdict)
	}
	if decision.BenchmarkOutcome != BenchmarkOutcomeNonRegression {
		t.Errorf("expected benchmark outcome NON_REGRESSION, got %s", decision.BenchmarkOutcome)
	}
}

// TestControlledOptimization_SpeedupWithBehavioralMismatchRejected verifies the core invariant:
// An optimization that achieves massive speedup (e.g. 50ms -> 5ms, 6 queries -> 0) by dropping
// a required database write or altering response content MUST BE REJECTED.
func TestControlledOptimization_SpeedupWithBehavioralMismatchRejected(t *testing.T) {
	candidate := OptimizationCandidate{
		CandidateID:          "opt-unsafe-cache-1",
		Description:          "Unsafe caching that omits audit log write and returns stale body",
		TargetComponent:      "order_service",
		OptimizedDBQueries:   0,
		OptimizedLatencyMs:   4.0, // 12x faster
		OptimizedResponseLen: 1024,
		IntroducesMismatch:   true,
		MismatchReason:       "omitted required DB audit log write and altered response headers",
	}

	fixture := NewControlledOptimizationFixture("opt-fix-unsafe", candidate, 500, 10)

	policy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
	})

	decision, err := EvaluateOptimization(fixture, policy, 42, "sha256:samples-opt-unsafe")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if decision.Accepted {
		t.Fatalf("CRITICAL SAFETY VIOLATION: optimization with behavioral mismatch was accepted!")
	}
	if decision.BehaviorVerdict != "MISMATCH" {
		t.Errorf("expected behavior verdict MISMATCH, got %s", decision.BehaviorVerdict)
	}
	// Even though benchmark alone would have been improvement, decision is rejected
	if decision.RejectionReason == "" {
		t.Errorf("expected non-empty rejection reason")
	}
}

// TestControlledOptimization_ParityPassButPerformanceRegressionRejected ensures that an
// optimization candidate that preserves behavior but accidentally regresses performance
// is rejected by the performance criteria.
func TestControlledOptimization_ParityPassButPerformanceRegressionRejected(t *testing.T) {
	candidate := OptimizationCandidate{
		CandidateID:          "opt-failed-index-1",
		Description:          "Faulty query rewrite that causes table scan",
		TargetComponent:      "search_service",
		OptimizedDBQueries:   6,
		OptimizedLatencyMs:   95.0, // Slowdown from 50ms -> 95ms
		OptimizedResponseLen: 2048,
		IntroducesMismatch:   false, // Behavioral parity preserved
	}

	fixture := NewControlledOptimizationFixture("opt-fix-slow", candidate, 500, 10)

	policy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
		{
			MetricID:                 MetricP95,
			Unit:                     "ms",
			LimitKind:                "STRICT_IMPROVEMENT",
			Limit:                    1.0,
			MinimumSuccessesPerTrial: 500,
		},
	})

	decision, err := EvaluateOptimization(fixture, policy, 42, "sha256:samples-opt-slow")
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if decision.Accepted {
		t.Fatalf("expected candidate with performance regression to be rejected")
	}
	if decision.BehaviorVerdict != "PASS" {
		t.Errorf("expected behavior verdict PASS, got %s", decision.BehaviorVerdict)
	}
	if decision.BenchmarkOutcome != BenchmarkOutcomeRegression {
		t.Errorf("expected benchmark outcome REGRESSION, got %s", decision.BenchmarkOutcome)
	}
}
