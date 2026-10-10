package domain

import (
	"fmt"
	"time"
)

// OptimizationCandidate represents a proposed optimization change to a target implementation.
type OptimizationCandidate struct {
	CandidateID           string   `json:"candidate_id"`
	Description           string   `json:"description"`
	TargetComponent       string   `json:"target_component"`
	OptimizedDBQueries    int      `json:"optimized_db_queries"`
	OptimizedLatencyMs    float64  `json:"optimized_latency_ms"`
	OptimizedResponseLen  int64    `json:"optimized_response_len"`
	IntroducesSideEffects bool     `json:"introduces_side_effects"`
	IntroducesMismatch    bool     `json:"introduces_mismatch"`
	MismatchReason        string   `json:"mismatch_reason,omitempty"`
}

// ControlledOptimizationFixture generates paired trial benchmark measurements for both
// baseline and candidate implementations, coupled with behavioral contract execution.
type ControlledOptimizationFixture struct {
	FixtureID           string                 `json:"fixture_id"`
	BaselineDBQueries   int                    `json:"baseline_db_queries"`
	BaselineLatencyMs   float64                `json:"baseline_latency_ms"`
	BaselineResponseLen int64                  `json:"baseline_response_len"`
	Candidate           OptimizationCandidate  `json:"candidate"`
	RequestsPerTrial    int                    `json:"requests_per_trial"`
	TrialCount          int                    `json:"trial_count"`
}

// NewControlledOptimizationFixture creates an optimization fixture with baseline and candidate settings.
func NewControlledOptimizationFixture(
	fixtureID string,
	candidate OptimizationCandidate,
	requestsPerTrial int,
	trialCount int,
) *ControlledOptimizationFixture {
	if requestsPerTrial < 500 {
		requestsPerTrial = 500
	}
	if trialCount < 10 || trialCount > 20 {
		trialCount = 10
	}

	return &ControlledOptimizationFixture{
		FixtureID:           fixtureID,
		BaselineDBQueries:   6, // baseline unoptimized: e.g. 6 queries with N+1 pattern
		BaselineLatencyMs:   50.0,
		BaselineResponseLen: 2048,
		Candidate:           candidate,
		RequestsPerTrial:    requestsPerTrial,
		TrialCount:          trialCount,
	}
}

// GeneratePairedTrials produces paired trial measurements for source (baseline) and target (candidate).
func (f *ControlledOptimizationFixture) GeneratePairedTrials() []PairedTrial {
	profiler := NewExecutionProfiler()
	trials := make([]PairedTrial, f.TrialCount)

	for trialIdx := 0; trialIdx < f.TrialCount; trialIdx++ {
		// 1. Generate Source (Baseline) Profiles
		sourceProfiles := make([]*RequestProfile, f.RequestsPerTrial)
		for r := 0; r < f.RequestsPerTrial; r++ {
			queries := make([]DBQuerySpan, f.BaselineDBQueries)
			for q := 0; q < f.BaselineDBQueries; q++ {
				queries[q] = DBQuerySpan{
					QueryID:    fmt.Sprintf("q-base-%d", q+1),
					SQL:        "SELECT * FROM catalog WHERE id = ?",
					Normalized: "SELECT * FROM CATALOG WHERE ID = ?",
					Duration:   time.Duration(f.BaselineLatencyMs/float64(f.BaselineDBQueries)*1000) * time.Microsecond,
				}
			}
			sourceProfiles[r] = &RequestProfile{
				RequestID:    fmt.Sprintf("src-%d-%d", trialIdx, r),
				StatusCode:   200,
				Success:      true,
				Duration:     time.Duration(f.BaselineLatencyMs*1000) * time.Microsecond,
				ResponseSize: f.BaselineResponseLen,
				DBQueries:    queries,
			}
		}
		sourceSummary := profiler.SummarizeTrial(trialIdx, sourceProfiles)

		// 2. Generate Target (Candidate) Profiles
		targetProfiles := make([]*RequestProfile, f.RequestsPerTrial)
		for r := 0; r < f.RequestsPerTrial; r++ {
			queries := make([]DBQuerySpan, f.Candidate.OptimizedDBQueries)
			for q := 0; q < f.Candidate.OptimizedDBQueries; q++ {
				queries[q] = DBQuerySpan{
					QueryID:    fmt.Sprintf("q-opt-%d", q+1),
					SQL:        "SELECT * FROM catalog WHERE id = ?",
					Normalized: "SELECT * FROM CATALOG WHERE ID = ?",
					Duration:   time.Duration(f.Candidate.OptimizedLatencyMs/float64(f.Candidate.OptimizedDBQueries)*1000) * time.Microsecond,
				}
			}
			targetProfiles[r] = &RequestProfile{
				RequestID:    fmt.Sprintf("tgt-%d-%d", trialIdx, r),
				StatusCode:   200,
				Success:      true,
				Duration:     time.Duration(f.Candidate.OptimizedLatencyMs*1000) * time.Microsecond,
				ResponseSize: f.Candidate.OptimizedResponseLen,
				DBQueries:    queries,
			}
		}
		targetSummary := profiler.SummarizeTrial(trialIdx, targetProfiles)

		trials[trialIdx] = PairedTrial{
			TrialIndex: trialIdx,
			SourceMetricValues: map[string]float64{
				MetricP95:          sourceSummary.P95LatencyMs,
				MetricDBQueryCount: sourceSummary.MeanDBQueryCount,
				MetricDBLatency:    sourceSummary.MeanDBLatencyMs,
				MetricResponseSize: sourceSummary.MeanResponseBytes,
			},
			TargetMetricValues: map[string]float64{
				MetricP95:          targetSummary.P95LatencyMs,
				MetricDBQueryCount: targetSummary.MeanDBQueryCount,
				MetricDBLatency:    targetSummary.MeanDBLatencyMs,
				MetricResponseSize: targetSummary.MeanResponseBytes,
			},
			SourceSuccesses: sourceSummary.SuccessfulRequests,
			TargetSuccesses: targetSummary.SuccessfulRequests,
			SourceErrors:    sourceSummary.FailedRequests,
			TargetErrors:    targetSummary.FailedRequests,
		}
	}

	return trials
}

// BehavioralEvaluationResult records whether behavioral parity was preserved.
type BehavioralEvaluationResult struct {
	OverallVerdict string            `json:"overall_verdict"` // PASS, MISMATCH, UNKNOWN
	DimensionVerdicts map[string]string `json:"dimension_verdicts"`
	MismatchReason    string            `json:"mismatch_reason,omitempty"`
}

// EvaluateBehavioralParity evaluates whether the optimization preserved 100% behavioral compatibility.
// Invariant per PERFORMANCE.md and AGENTS.md:
// "An optimization is acceptable only when:
//  - behavioral compatibility passes for the required acceptance scope; and
//  - performance improves under the specified method.
//  If compatibility fails, reject the optimization even if the benchmark shows a substantial speedup."
func (f *ControlledOptimizationFixture) EvaluateBehavioralParity() BehavioralEvaluationResult {
	if f.Candidate.IntroducesMismatch {
		return BehavioralEvaluationResult{
			OverallVerdict: "MISMATCH",
			DimensionVerdicts: map[string]string{
				"http":     "PASS",
				"database": "MISMATCH",
				"body":     "MISMATCH",
			},
			MismatchReason: f.Candidate.MismatchReason,
		}
	}

	return BehavioralEvaluationResult{
		OverallVerdict: "PASS",
		DimensionVerdicts: map[string]string{
			"http":           "PASS",
			"body":           "PASS",
			"authentication": "PASS",
			"authorization":  "PASS",
			"session":        "PASS",
			"database":       "PASS",
			"external":       "PASS",
			"execution":      "PASS",
		},
	}
}

// OptimizationDecision captures the combined outcome of performance benchmarking and behavioral verification.
type OptimizationDecision struct {
	CandidateID        string                      `json:"candidate_id"`
	Accepted           bool                        `json:"accepted"`
	BehaviorVerdict    string                      `json:"behavior_verdict"`
	BenchmarkOutcome   string                      `json:"benchmark_outcome"`
	RejectionReason    string                      `json:"rejection_reason,omitempty"`
	BenchmarkResult    *StatisticalBenchmarkResult `json:"benchmark_result"`
	BehavioralResult   BehavioralEvaluationResult  `json:"behavioral_result"`
}

// EvaluateOptimization evaluates both the statistical performance improvement and the behavioral parity.
func EvaluateOptimization(
	fixture *ControlledOptimizationFixture,
	policy *PerformancePolicyRecord,
	seed int,
	rawSamplesRef string,
) (*OptimizationDecision, error) {
	if fixture == nil {
		return nil, fmt.Errorf("fixture is nil")
	}

	// 1. Evaluate behavioral parity FIRST (correctness precedes performance)
	behaviorRes := fixture.EvaluateBehavioralParity()

	// 2. Evaluate statistical benchmark
	trials := fixture.GeneratePairedTrials()
	benchRes, err := EvaluateStatisticalPerformance(policy, trials, seed, rawSamplesRef)
	if err != nil {
		return nil, fmt.Errorf("benchmark evaluation failed: %w", err)
	}

	decision := &OptimizationDecision{
		CandidateID:      fixture.Candidate.CandidateID,
		BehaviorVerdict:  behaviorRes.OverallVerdict,
		BenchmarkOutcome: benchRes.Outcome,
		BenchmarkResult:  benchRes,
		BehavioralResult: behaviorRes,
	}

	// Invariant: Behavioral parity MUST be PASS. Any MISMATCH or UNKNOWN forbids acceptance.
	if behaviorRes.OverallVerdict != "PASS" {
		decision.Accepted = false
		decision.RejectionReason = fmt.Sprintf("REJECTED_BEHAVIORAL_MISMATCH: candidate failed behavioral contract (%s: %s) despite benchmark outcome %s",
			behaviorRes.OverallVerdict, behaviorRes.MismatchReason, benchRes.Outcome)
		return decision, nil
	}

	// Invariant: Benchmark outcome must demonstrate improvement or non-regression.
	// If REGRESSION or INCONCLUSIVE, reject.
	if benchRes.Outcome != BenchmarkOutcomeNonRegression && benchRes.Outcome != BenchmarkOutcomeApprovedBudget {
		decision.Accepted = false
		decision.RejectionReason = fmt.Sprintf("REJECTED_PERFORMANCE_CRITERIA: benchmark outcome is %s", benchRes.Outcome)
		return decision, nil
	}

	decision.Accepted = true
	return decision, nil
}
