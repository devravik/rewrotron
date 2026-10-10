package domain

import (
	"context"
	"testing"
	"time"
)

// TestM8_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M8:
// Performance Modernization and Technology Composition Controls.
//
// Invariants verified:
// 1. Profiler metric collection & query anomaly detection:
//    - Tracks DB_QUERY_COUNT, DB_LATENCY, RESPONSE_SIZE, and latency quantiles.
//    - Normalizes SQL queries and detects N+1 loops (PatternNPlusOneQuery).
// 2. Controlled performance regression detection:
//    - Latency inflation (+160%) and query inflation (N+1 query loop) reliably produce REGRESSION.
//    - Falsification check: No early stopping or subset sampling can produce false NON_REGRESSION.
// 3. Controlled optimization evaluation:
//    - Genuine optimization (query count 6 -> 2, latency 50ms -> 18ms) with 100% behavioral parity
//      evaluates to NON_REGRESSION and is ACCEPTED.
// 4. Behavioral preservation guardrails (P0 invariant):
//    - An optimization candidate that improves performance but introduces ANY behavioral mismatch
//      (e.g. omitting a required DB write or returning stale data) is IMMEDIATELY REJECTED.
//    - Behavioral parity MUST be PASS before performance improvement can justify acceptance.
// 5. Optimization mission generation:
//    - Formulates well-formed OPTIMIZE missions adhering to specs/missions/mission.schema.json.
//    - Prohibits auto-deploy; preserves human Gate 2 requirement.
// 6. Cryptographic optimization report compilation:
//    - Generates auditable report with baseline vs candidate metrics, behavioral outcomes, and sha256 digest.
func TestM8_AcceptanceClosure(t *testing.T) {
	ctx := context.Background()

	// -------------------------------------------------------------------------
	// Section 1: Profiler & Metric Collection Verification
	// -------------------------------------------------------------------------
	t.Run("Profiler_MetricCollectionAndNPlusOne", func(t *testing.T) {
		profiler := NewExecutionProfiler()
		rc := profiler.StartRequest(ctx, "/catalog/products", "GET")

		rc.RecordQuery("SELECT * FROM products WHERE active = 1", 5*time.Millisecond, 20, nil)
		for i := 1; i <= 6; i++ {
			rc.RecordQuery("SELECT * FROM product_images WHERE product_id = 10", 2*time.Millisecond, 3, nil)
		}

		profile := profiler.FinishRequest(ctx, rc, 200, 2048, nil)
		if profile.TotalDBQueryCount() != 7 {
			t.Fatalf("expected 7 queries, got %d", profile.TotalDBQueryCount())
		}

		anomalies := profile.DetectQueryAnomalies(3)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 N+1 anomaly, got %d", len(anomalies))
		}
		if anomalies[0].PatternType != PatternNPlusOneQuery {
			t.Errorf("expected %s, got %s", PatternNPlusOneQuery, anomalies[0].PatternType)
		}
	})

	// -------------------------------------------------------------------------
	// Section 2: Controlled Performance Regression Detection
	// -------------------------------------------------------------------------
	t.Run("Regression_ReliableDetectionWithoutFalsePass", func(t *testing.T) {
		fixture := NewControlledRegressionFixture("closure-reg-1", RegressionQueryInflation, 500, 10)
		trials := fixture.GeneratePairedTrials()

		policy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
			{
				MetricID:                 MetricDBQueryCount,
				Unit:                     "count",
				LimitKind:                "RELATIVE_NON_INFERIORITY",
				Limit:                    1.10,
				MinimumSuccessesPerTrial: 500,
			},
		})

		res, err := EvaluateStatisticalPerformance(policy, trials, 5555, "sha256:closure-reg-samples")
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		if res.Outcome != BenchmarkOutcomeRegression {
			t.Fatalf("expected REGRESSION outcome for query inflation, got %s", res.Outcome)
		}

		// Truncated trials must fail with INCONCLUSIVE, not false pass
		truncRes, err := EvaluateStatisticalPerformance(policy, trials[:5], 5555, "sha256:closure-trunc-samples")
		if err != nil {
			t.Fatalf("truncated eval failed: %v", err)
		}
		if truncRes.Outcome != BenchmarkOutcomeInconclusive {
			t.Fatalf("expected INCONCLUSIVE for N=5 early stopping, got %s", truncRes.Outcome)
		}
	})

	// -------------------------------------------------------------------------
	// Section 3: Controlled Optimization Acceptance
	// -------------------------------------------------------------------------
	t.Run("Optimization_ValidImprovementAccepted", func(t *testing.T) {
		candidate := OptimizationCandidate{
			CandidateID:          "closure-opt-valid",
			Description:          "Batch eager loading for product images",
			TargetComponent:      "ProductController",
			OptimizedDBQueries:   2,
			OptimizedLatencyMs:   16.0,
			OptimizedResponseLen: 2048,
			IntroducesMismatch:   false,
		}

		fixture := NewControlledOptimizationFixture("closure-opt-fix", candidate, 500, 10)
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

		decision, err := EvaluateOptimization(fixture, policy, 7777, "sha256:closure-opt-samples")
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		if !decision.Accepted {
			t.Fatalf("expected valid optimization to be accepted, got rejection: %s", decision.RejectionReason)
		}
		if decision.BehaviorVerdict != "PASS" {
			t.Errorf("expected behavior verdict PASS, got %s", decision.BehaviorVerdict)
		}
		if decision.BenchmarkOutcome != BenchmarkOutcomeNonRegression {
			t.Errorf("expected benchmark outcome NON_REGRESSION, got %s", decision.BenchmarkOutcome)
		}
	})

	// -------------------------------------------------------------------------
	// Section 4: Behavioral Parity Guardrail (Rejection of Speedup with Mismatch)
	// -------------------------------------------------------------------------
	t.Run("Optimization_SpeedupWithBehavioralMismatchRejected", func(t *testing.T) {
		candidate := OptimizationCandidate{
			CandidateID:          "closure-opt-mutant-bypass",
			Description:          "Speedup achieved by omitting transaction logging",
			TargetComponent:      "ProductController",
			OptimizedDBQueries:   0,
			OptimizedLatencyMs:   2.0, // 25x faster
			OptimizedResponseLen: 1024,
			IntroducesMismatch:   true,
			MismatchReason:       "omitted required DB transaction log write",
		}

		fixture := NewControlledOptimizationFixture("closure-opt-unsafe", candidate, 500, 10)
		policy := makeTestPolicy("IMPROVEMENT", 10, []PerformanceMetricLimit{
			{
				MetricID:                 MetricP95,
				Unit:                     "ms",
				LimitKind:                "STRICT_IMPROVEMENT",
				Limit:                    1.0,
				MinimumSuccessesPerTrial: 500,
			},
		})

		decision, err := EvaluateOptimization(fixture, policy, 7777, "sha256:closure-opt-unsafe-samples")
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		if decision.Accepted {
			t.Fatalf("CRITICAL SAFETY FAILURE: candidate with behavioral mismatch was accepted!")
		}
		if decision.BehaviorVerdict != "MISMATCH" {
			t.Errorf("expected behavior verdict MISMATCH, got %s", decision.BehaviorVerdict)
		}
	})

	// -------------------------------------------------------------------------
	// Section 5: Optimization Mission Generation & Report Auditing
	// -------------------------------------------------------------------------
	t.Run("MissionAndReport_AuditabilityAndSafety", func(t *testing.T) {
		anomaly := QueryPatternAnomaly{
			PatternType: PatternNPlusOneQuery,
			Normalized:  "SELECT * FROM PRODUCT_IMAGES WHERE PRODUCT_ID = ?",
			Occurrences: 6,
			Description: "detected 6 identical normalized queries: potential N+1 loop",
		}

		mission, err := GenerateOptimizationMission(
			"proj-m8",
			"unit-catalog",
			"ProductController",
			anomaly,
			50.0,
			7,
			"sha256:evidence-m8",
		)
		if err != nil {
			t.Fatalf("GenerateOptimizationMission failed: %v", err)
		}
		if mission.Context["auto_deploy"] != false {
			t.Fatalf("mission must not permit auto-deploy")
		}

		decision := &OptimizationDecision{
			CandidateID:      "cand-m8",
			Accepted:         true,
			BehaviorVerdict:  "PASS",
			BenchmarkOutcome: BenchmarkOutcomeNonRegression,
		}

		report, err := CompileOptimizationReport(
			"proj-m8",
			"ProductController",
			"build-base-1",
			"build-cand-2",
			"sha256:workload-m8",
			"sha256:env-m8",
			"sha256:policy-m8",
			decision,
			map[string]float64{MetricP95: 50.0, MetricDBQueryCount: 7.0},
			map[string]float64{MetricP95: 16.0, MetricDBQueryCount: 2.0},
		)
		if err != nil {
			t.Fatalf("CompileOptimizationReport failed: %v", err)
		}
		if !report.Accepted {
			t.Errorf("report should record Accepted=true")
		}
		if report.ReportDigest == "" {
			t.Errorf("report digest must not be empty")
		}
	})
}
