package domain

import (
	"fmt"
	"time"
)

// RegressionType specifies the dimension of performance degradation injected.
type RegressionType string

const (
	RegressionLatencyInflation   RegressionType = "LATENCY_INFLATION"
	RegressionQueryInflation     RegressionType = "QUERY_INFLATION"
	RegressionResponseInflation  RegressionType = "RESPONSE_SIZE_INFLATION"
	RegressionCompositeSlowdown  RegressionType = "COMPOSITE_SLOWDOWN"
)

// ControlledRegressionFixture generates controlled benchmark workloads where a known
// implementation mutation introduces measurable performance degradation while strictly
// maintaining 100% behavioral parity (identical responses, side-effects, and status codes).
type ControlledRegressionFixture struct {
	FixtureID           string         `json:"fixture_id"`
	RegressionType      RegressionType `json:"regression_type"`
	BaselineDBQueries   int            `json:"baseline_db_queries"`
	MutatedDBQueries    int            `json:"mutated_db_queries"`
	BaselineLatencyMs   float64        `json:"baseline_latency_ms"`
	MutatedLatencyMs    float64        `json:"mutated_latency_ms"`
	BaselineResponseLen int64          `json:"baseline_response_len"`
	MutatedResponseLen  int64          `json:"mutated_response_len"`
	RequestsPerTrial    int            `json:"requests_per_trial"`
	TrialCount          int            `json:"trial_count"`
}

// NewControlledRegressionFixture creates a controlled regression fixture with explicit degradation parameters.
func NewControlledRegressionFixture(
	fixtureID string,
	regType RegressionType,
	requestsPerTrial int,
	trialCount int,
) *ControlledRegressionFixture {
	if requestsPerTrial < 500 {
		requestsPerTrial = 500
	}
	if trialCount < 10 || trialCount > 20 {
		trialCount = 10
	}

	f := &ControlledRegressionFixture{
		FixtureID:           fixtureID,
		RegressionType:      regType,
		BaselineDBQueries:   2,
		MutatedDBQueries:    2,
		BaselineLatencyMs:   25.0,
		MutatedLatencyMs:    25.0,
		BaselineResponseLen: 1024,
		MutatedResponseLen:  1024,
		RequestsPerTrial:    requestsPerTrial,
		TrialCount:          trialCount,
	}

	switch regType {
	case RegressionLatencyInflation:
		// Baseline 25ms, Mutated 65ms (+160% latency inflation)
		f.MutatedLatencyMs = 65.0
	case RegressionQueryInflation:
		// Baseline 2 queries (e.g. JOIN / eager load), Mutated 8 queries (e.g. injected N+1 query loop)
		f.MutatedDBQueries = 8
		f.MutatedLatencyMs = 55.0
	case RegressionResponseInflation:
		// Baseline 1KB, Mutated 10KB (e.g. redundant serialized payload fields)
		f.MutatedResponseLen = 10240
		f.MutatedLatencyMs = 40.0
	case RegressionCompositeSlowdown:
		f.MutatedDBQueries = 6
		f.MutatedLatencyMs = 70.0
		f.MutatedResponseLen = 4096
	}

	return f
}

// GeneratePairedTrials produces the exact paired trial measurements for source (baseline)
// and target (mutated regression), using the ExecutionProfiler to aggregate request traces.
func (f *ControlledRegressionFixture) GeneratePairedTrials() []PairedTrial {
	profiler := NewExecutionProfiler()
	trials := make([]PairedTrial, f.TrialCount)

	for trialIdx := 0; trialIdx < f.TrialCount; trialIdx++ {
		// 1. Generate Source (Baseline) Request Profiles
		sourceProfiles := make([]*RequestProfile, f.RequestsPerTrial)
		for r := 0; r < f.RequestsPerTrial; r++ {
			queries := make([]DBQuerySpan, f.BaselineDBQueries)
			for q := 0; q < f.BaselineDBQueries; q++ {
				queries[q] = DBQuerySpan{
					QueryID:    fmt.Sprintf("q-%d", q+1),
					SQL:        "SELECT * FROM items WHERE id = ?",
					Normalized: "SELECT * FROM ITEMS WHERE ID = ?",
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

		// 2. Generate Target (Mutated/Regressed) Request Profiles
		targetProfiles := make([]*RequestProfile, f.RequestsPerTrial)
		for r := 0; r < f.RequestsPerTrial; r++ {
			queries := make([]DBQuerySpan, f.MutatedDBQueries)
			for q := 0; q < f.MutatedDBQueries; q++ {
				queries[q] = DBQuerySpan{
					QueryID:    fmt.Sprintf("q-mut-%d", q+1),
					SQL:        "SELECT * FROM items WHERE id = ?",
					Normalized: "SELECT * FROM ITEMS WHERE ID = ?",
					Duration:   time.Duration(f.MutatedLatencyMs/float64(f.MutatedDBQueries)*1000) * time.Microsecond,
				}
			}
			targetProfiles[r] = &RequestProfile{
				RequestID:    fmt.Sprintf("tgt-%d-%d", trialIdx, r),
				StatusCode:   200,
				Success:      true,
				Duration:     time.Duration(f.MutatedLatencyMs*1000) * time.Microsecond,
				ResponseSize: f.MutatedResponseLen,
				DBQueries:    queries,
			}
		}
		targetSummary := profiler.SummarizeTrial(trialIdx, targetProfiles)

		// 3. Map into canonical PairedTrial record
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
