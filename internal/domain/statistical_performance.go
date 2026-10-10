package domain

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
)

// Metric ID constants per performance-policy.schema.json and benchmark.schema.json
const (
	MetricP50             = "P50"
	MetricP95             = "P95"
	MetricP99             = "P99"
	MetricCPU             = "CPU"
	MetricRSSMemory       = "RSS_MEMORY"
	MetricDBQueryCount    = "DB_QUERY_COUNT"
	MetricDBLatency       = "DB_LATENCY"
	MetricExternalLatency = "EXTERNAL_LATENCY"
	MetricResponseSize    = "RESPONSE_SIZE"
	MetricStartupCold     = "STARTUP_COLD"
	MetricStartupWarm     = "STARTUP_WARM"
)

// Benchmark outcome constants per PERFORMANCE.md
const (
	BenchmarkOutcomeNonRegression  = "NON_REGRESSION"
	BenchmarkOutcomeApprovedBudget = "APPROVED_BUDGET"
	BenchmarkOutcomeRegression     = "REGRESSION"
	BenchmarkOutcomeInconclusive   = "INCONCLUSIVE"
)

// Operational SLO outcome constants
const (
	SLOOutcomePass    = "PASS"
	SLOOutcomeFail    = "FAIL"
	SLOOutcomeUnknown = "UNKNOWN"
)

// PairedTrial represents one raw paired trial measurement across source and target.
type PairedTrial struct {
	TrialIndex int `json:"trial_index"`

	// MetricValues maps metric ID (e.g. "P95", "RSS_MEMORY") to trial values for source and target.
	SourceMetricValues map[string]float64 `json:"source_metric_values"`
	TargetMetricValues map[string]float64 `json:"target_metric_values"`

	SourceSuccesses int `json:"source_successes"`
	TargetSuccesses int `json:"target_successes"`
	SourceErrors    int `json:"source_errors"`
	TargetErrors    int `json:"target_errors"`
}

// WindowSample represents one measured time-window sample for operational SLO evaluation.
type WindowSample struct {
	WindowIndex        int       `json:"window_index"`
	TrialIndex         int       `json:"trial_index"`
	StartTimeMs        int64     `json:"start_time_ms"`
	EndTimeMs          int64     `json:"end_time_ms"`
	SuccessfulRequests []float64 `json:"successful_requests"` // Latency measurements in ms
	TotalAttempts      int       `json:"total_attempts"`
	ErrorCount         int       `json:"error_count"`
}

// OperationalSLOPolicy represents the operational-slo record (schema 1.0).
type OperationalSLOPolicy struct {
	SchemaVersion            string   `json:"schema_version"`
	ProjectID                string   `json:"project_id"`
	WorkloadRef              string   `json:"workload_ref"`
	WindowMs                 int      `json:"window_ms"`
	MinimumRequestsPerWindow int      `json:"minimum_requests_per_window"`
	LatencyPercentile        float64  `json:"latency_percentile"` // 0.5, 0.95, or 0.99
	MaximumLatencyMs         string   `json:"maximum_latency_ms"`
	MaximumErrorFraction     float64  `json:"maximum_error_fraction"`
	Scope                    string   `json:"scope"` // "EVERY_COMPLETE_MEASURED_WINDOW"
	EvidenceRefs             []string `json:"evidence_refs"`
}

// OperationalSLOResult represents the evaluated operational SLO outcome.
type OperationalSLOResult struct {
	PolicyRef           string   `json:"policy_ref"`
	Outcome             string   `json:"outcome"` // PASS, FAIL, UNKNOWN
	RawWindowSamplesRef string   `json:"raw_window_samples_ref"`
	ReasonCodes         []string `json:"reason_codes"`
}

// MetricEvaluationResult contains the statistical evaluation for a single required metric.
type MetricEvaluationResult struct {
	MetricID          string   `json:"metric_id"`
	Unit              string   `json:"unit"`
	Required          bool     `json:"required"`
	SourceValue       *float64 `json:"source_value"` // mean of trial values
	TargetValue       *float64 `json:"target_value"` // mean of trial values
	LowerBound        *float64 `json:"lower_bound"`
	UpperBound        *float64 `json:"upper_bound"`
	SourceSampleCount int      `json:"source_sample_count"`
	TargetSampleCount int      `json:"target_sample_count"`
	ApprovedLimit     *float64 `json:"approved_limit"`
	RawSamplesRef     string   `json:"raw_samples_ref"`
	Outcome           string   `json:"outcome"` // NON_REGRESSION, APPROVED_BUDGET, REGRESSION, INCONCLUSIVE
	Reason            string   `json:"reason,omitempty"`
}

// StatisticalBenchmarkResult represents the complete recomputed outcome of benchmark evaluation.
type StatisticalBenchmarkResult struct {
	Outcome              string                            `json:"outcome"`
	Claim                string                            `json:"claim"`
	Metrics              map[string]MetricEvaluationResult `json:"metrics"`
	OperationalSLOResult *OperationalSLOResult             `json:"operational_slo_result"`
	ReasonCodes          []string                          `json:"reason_codes"`
}

// ResampleIndices generates 10,000 resampled draw index arrays of size n using the
// pinned RWT_PAIRED_BOOTSTRAP_1 pseudo-random generation method.
//
// Specification per PERFORMANCE.md:
// "For replicate r and draw d (both zero-based), hash UTF-8 `rwt-paired-bootstrap-1:<seed>:<r>:<d>:<attempt>`
//  with SHA256, starting attempt=0. Interpret the digest as unsigned big-endian x;
//  reject x >= floor(2^256/n)*n and increment attempt, otherwise select index x mod n from the original n paired trials."
func ResampleIndices(seed int, n int, numReplicates int) ([][]int, error) {
	if n < 10 || n > 20 {
		return nil, fmt.Errorf("sample size n must be between 10 and 20; got %d", n)
	}
	if numReplicates <= 0 {
		numReplicates = 10000
	}

	// Calculate twoTo256 = 2^256
	twoTo256 := new(big.Int).Lsh(big.NewInt(1), 256)
	bigN := big.NewInt(int64(n))

	// limit = floor(2^256 / n) * n
	div := new(big.Int).Div(twoTo256, bigN)
	limit := new(big.Int).Mul(div, bigN)

	replicates := make([][]int, numReplicates)
	x := new(big.Int)
	modVal := new(big.Int)

	for r := 0; r < numReplicates; r++ {
		drawIndices := make([]int, n)
		for d := 0; d < n; d++ {
			attempt := 0
			for {
				msg := fmt.Sprintf("rwt-paired-bootstrap-1:%d:%d:%d:%d", seed, r, d, attempt)
				digest := sha256.Sum256([]byte(msg))
				x.SetBytes(digest[:])

				// reject x >= floor(2^256/n)*n
				if x.Cmp(limit) < 0 {
					modVal.Mod(x, bigN)
					drawIndices[d] = int(modVal.Int64())
					break
				}
				attempt++
			}
		}
		replicates[r] = drawIndices
	}

	return replicates, nil
}

// ComputeQuantile calculates the nearest-rank quantile ceil(p*n) on ordered samples.
func ComputeQuantile(ordered []float64, p float64) float64 {
	n := len(ordered)
	if n == 0 {
		return 0
	}
	rank := int(math.Ceil(p * float64(n)))
	if rank < 1 {
		rank = 1
	}
	if rank > n {
		rank = n
	}
	return ordered[rank-1]
}

// EvaluateStatisticalPerformance independently recomputes the statistical benchmark
// from raw paired trials according to RWT_PAIRED_BOOTSTRAP_1 and the approved PerformancePolicyRecord.
func EvaluateStatisticalPerformance(
	policy *PerformancePolicyRecord,
	trials []PairedTrial,
	seed int,
	rawSamplesRef string,
) (*StatisticalBenchmarkResult, error) {
	if policy == nil {
		return nil, fmt.Errorf("performance policy is nil")
	}

	// 1. Check Not-Applicable policy handling per PERFORMANCE.md:
	// "A not-applicable policy permits an explicitly scoped release without a performance claim,
	//  never a fabricated accepted benchmark."
	if !policy.Applicable {
		return &StatisticalBenchmarkResult{
			Outcome:     BenchmarkOutcomeApprovedBudget,
			Claim:       "",
			Metrics:     map[string]MetricEvaluationResult{},
			ReasonCodes: []string{"POLICY_NOT_APPLICABLE"},
		}, nil
	}

	claim := ""
	if policy.Claim != nil {
		claim = *policy.Claim
	}
	if claim != "IMPROVEMENT" && claim != "NON_INFERIORITY" && claim != "APPROVED_BUDGET" {
		return nil, fmt.Errorf("invalid performance claim: %s", claim)
	}

	result := &StatisticalBenchmarkResult{
		Claim:       claim,
		Metrics:     make(map[string]MetricEvaluationResult),
		ReasonCodes: []string{},
	}

	// 2. Enforce fixed stopping rule (X78) and paired trial bounds:
	// 10 <= N <= 20
	expectedN := 0
	if policy.PairedTrials != nil {
		expectedN = *policy.PairedTrials
	}
	if expectedN < 10 || expectedN > 20 {
		return nil, fmt.Errorf("policy paired_trials must be between 10 and 20; got %d", expectedN)
	}

	if len(trials) != expectedN {
		result.Outcome = BenchmarkOutcomeInconclusive
		result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("ERR_INCOMPLETE_N: expected %d trials, got %d", expectedN, len(trials)))
		return result, nil
	}

	// 3. Error rate validation across trials
	var totalSourceSuccess, totalSourceError int
	var totalTargetSuccess, totalTargetError int
	for _, t := range trials {
		totalSourceSuccess += t.SourceSuccesses
		totalSourceError += t.SourceErrors
		totalTargetSuccess += t.TargetSuccesses
		totalTargetError += t.TargetErrors
	}

	sourceTotal := totalSourceSuccess + totalSourceError
	targetTotal := totalTargetSuccess + totalTargetError
	var sourceErrorRate, targetErrorRate float64
	if sourceTotal > 0 {
		sourceErrorRate = float64(totalSourceError) / float64(sourceTotal)
	}
	if targetTotal > 0 {
		targetErrorRate = float64(totalTargetError) / float64(targetTotal)
	}

	if sourceErrorRate > policy.MaxErrorRate || targetErrorRate > policy.MaxErrorRate {
		result.Outcome = BenchmarkOutcomeInconclusive
		result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("ERR_ERROR_RATE_EXCEEDED: source=%.4f target=%.4f max=%.4f", sourceErrorRate, targetErrorRate, policy.MaxErrorRate))
		return result, nil
	}

	// 4. Precompute 10,000 bootstrap resamples
	replicates, err := ResampleIndices(seed, expectedN, 10000)
	if err != nil {
		return nil, fmt.Errorf("resampling failed: %w", err)
	}

	k := len(policy.RequiredMetrics)
	if k == 0 {
		result.Outcome = BenchmarkOutcomeInconclusive
		result.ReasonCodes = append(result.ReasonCodes, "ERR_NO_REQUIRED_METRICS")
		return result, nil
	}

	// Bonferroni-corrected alpha endpoints: alpha/(2*k) and 1 - alpha/(2*k), alpha = 0.05
	alpha := 0.05
	pLower := alpha / (2.0 * float64(k))
	pUpper := 1.0 - (alpha / (2.0 * float64(k)))

	hasRegression := false
	hasInconclusive := false

	// 5. Evaluate each required metric
	for _, metricLimit := range policy.RequiredMetrics {
		metricID := metricLimit.MetricID
		res := MetricEvaluationResult{
			MetricID:      metricID,
			Unit:          metricLimit.Unit,
			Required:      true,
			RawSamplesRef: rawSamplesRef,
		}
		lim := metricLimit.Limit
		res.ApprovedLimit = &lim

		// Verify sample count per trial requirements
		// Required p50/p95/CPU/memory claims need at least 500 successful requests on each side of each measured trial; p99 needs at least 10,000
		minSamplesPerTrial := metricLimit.MinimumSuccessesPerTrial
		switch metricID {
		case MetricP50, MetricP95, MetricCPU, MetricRSSMemory:
			if minSamplesPerTrial < 500 {
				minSamplesPerTrial = 500
			}
		case MetricP99:
			if minSamplesPerTrial < 10000 {
				minSamplesPerTrial = 10000
			}
		}

		sourceTrialVals := make([]float64, expectedN)
		targetTrialVals := make([]float64, expectedN)
		sufficientSamples := true

		for i, t := range trials {
			if t.SourceSuccesses < minSamplesPerTrial || t.TargetSuccesses < minSamplesPerTrial {
				sufficientSamples = false
				break
			}
			sVal, sOk := t.SourceMetricValues[metricID]
			tVal, tOk := t.TargetMetricValues[metricID]
			if !sOk || !tOk {
				sufficientSamples = false
				break
			}
			sourceTrialVals[i] = sVal
			targetTrialVals[i] = tVal
		}

		if !sufficientSamples {
			res.Outcome = BenchmarkOutcomeInconclusive
			res.Reason = fmt.Sprintf("insufficient successful samples for metric %s (requires >= %d per trial)", metricID, minSamplesPerTrial)
			result.Metrics[metricID] = res
			hasInconclusive = true
			continue
		}

		res.SourceSampleCount = totalSourceSuccess
		res.TargetSampleCount = totalTargetSuccess

		// Calculate mean of trial values
		var sumSource, sumTarget float64
		for i := 0; i < expectedN; i++ {
			sumSource += sourceTrialVals[i]
			sumTarget += targetTrialVals[i]
		}
		meanSource := sumSource / float64(expectedN)
		meanTarget := sumTarget / float64(expectedN)
		res.SourceValue = &meanSource
		res.TargetValue = &meanTarget

		// Check zero denominator for relative estimands
		isRelative := (claim == "IMPROVEMENT" || claim == "NON_INFERIORITY" || metricLimit.LimitKind == "RELATIVE_NON_INFERIORITY" || metricLimit.LimitKind == "STRICT_IMPROVEMENT")
		if isRelative && meanSource == 0 {
			res.Outcome = BenchmarkOutcomeInconclusive
			res.Reason = "zero source mean for relative estimand"
			result.Metrics[metricID] = res
			hasInconclusive = true
			continue
		}

		// Recompute 10,000 replicate estimands
		replicateEstimands := make([]float64, len(replicates))
		zeroDenominatorInBootstrap := false

		for rIdx, draw := range replicates {
			var rSourceSum, rTargetSum float64
			for _, trialIdx := range draw {
				rSourceSum += sourceTrialVals[trialIdx]
				rTargetSum += targetTrialVals[trialIdx]
			}
			rSourceMean := rSourceSum / float64(expectedN)
			rTargetMean := rTargetSum / float64(expectedN)

			if isRelative {
				if rSourceMean == 0 {
					zeroDenominatorInBootstrap = true
					break
				}
				replicateEstimands[rIdx] = rTargetMean / rSourceMean
			} else {
				// Absolute budget
				replicateEstimands[rIdx] = rTargetMean
			}
		}

		if zeroDenominatorInBootstrap {
			res.Outcome = BenchmarkOutcomeInconclusive
			res.Reason = "zero source mean encountered during bootstrap replication"
			result.Metrics[metricID] = res
			hasInconclusive = true
			continue
		}

		sort.Float64s(replicateEstimands)

		lowerBound := ComputeQuantile(replicateEstimands, pLower)
		upperBound := ComputeQuantile(replicateEstimands, pUpper)
		res.LowerBound = &lowerBound
		res.UpperBound = &upperBound

		// Evaluate metric outcome against approved limits per PERFORMANCE.md
		// Claim rules:
		// IMPROVEMENT: all required ratio upper bounds strictly < 1
		// NON_INFERIORITY: all upper bounds <= approved relative limits > 1
		// APPROVED_BUDGET: all upper bounds <= their approved absolute/relative limits
		// Complete evidence: at least one lower bound strictly exceeds its approved upper acceptance limit => REGRESSION
		limit := 1.0
		if res.ApprovedLimit != nil {
			limit = *res.ApprovedLimit
		}

		switch claim {
		case "IMPROVEMENT":
			if lowerBound > 1.0 {
				res.Outcome = BenchmarkOutcomeRegression
				hasRegression = true
			} else if upperBound < 1.0 {
				res.Outcome = BenchmarkOutcomeNonRegression
			} else {
				// Overlapping 1.0 or exact interval [1,1] is INCONCLUSIVE
				res.Outcome = BenchmarkOutcomeInconclusive
				hasInconclusive = true
			}
		case "NON_INFERIORITY":
			if lowerBound > limit {
				res.Outcome = BenchmarkOutcomeRegression
				hasRegression = true
			} else if upperBound <= limit {
				res.Outcome = BenchmarkOutcomeApprovedBudget
			} else {
				res.Outcome = BenchmarkOutcomeInconclusive
				hasInconclusive = true
			}
		case "APPROVED_BUDGET":
			if lowerBound > limit {
				res.Outcome = BenchmarkOutcomeRegression
				hasRegression = true
			} else if upperBound <= limit {
				res.Outcome = BenchmarkOutcomeApprovedBudget
			} else {
				res.Outcome = BenchmarkOutcomeInconclusive
				hasInconclusive = true
			}
		}

		result.Metrics[metricID] = res
	}

	// Determine overall statistical outcome:
	// "Known complete regression evidence may be reported while another metric is inconclusive,
	//  but no accepted outcome may omit an unresolved required metric."
	if hasRegression {
		result.Outcome = BenchmarkOutcomeRegression
	} else if hasInconclusive {
		result.Outcome = BenchmarkOutcomeInconclusive
	} else {
		if claim == "IMPROVEMENT" {
			result.Outcome = BenchmarkOutcomeNonRegression
		} else {
			result.Outcome = BenchmarkOutcomeApprovedBudget
		}
	}

	return result, nil
}

// EvaluateOperationalSLO evaluates fixed measured window latency and error fraction
// per operational-slo.schema.json and PERFORMANCE.md.
//
// Invariants:
// - Scope is EVERY_COMPLETE_MEASURED_WINDOW.
// - Insufficient requests, incomplete windows, missing samples produce UNKNOWN.
// - Any fully observed violating window yields FAIL even if mean trial p99 passes.
// - Zero-request windows are UNKNOWN, never zero latency.
func EvaluateOperationalSLO(
	policy *OperationalSLOPolicy,
	windows []WindowSample,
	policyRef string,
	rawWindowSamplesRef string,
) (*OperationalSLOResult, error) {
	if policy == nil {
		return nil, nil // Null operational SLO requested
	}

	result := &OperationalSLOResult{
		PolicyRef:           policyRef,
		RawWindowSamplesRef: rawWindowSamplesRef,
		ReasonCodes:         []string{},
	}

	if policy.Scope != "EVERY_COMPLETE_MEASURED_WINDOW" {
		result.Outcome = SLOOutcomeFail
		result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("ERR_INVALID_SCOPE: %s", policy.Scope))
		return result, nil
	}

	maxLatency, err := strconv.ParseFloat(policy.MaximumLatencyMs, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid maximum_latency_ms: %s: %w", policy.MaximumLatencyMs, err)
	}

	if len(windows) == 0 {
		result.Outcome = SLOOutcomeUnknown
		result.ReasonCodes = append(result.ReasonCodes, "ERR_NO_MEASURED_WINDOWS")
		return result, nil
	}

	hasUnknown := false

	for _, w := range windows {
		// Zero-request window is UNKNOWN
		if len(w.SuccessfulRequests) == 0 && w.TotalAttempts == 0 {
			hasUnknown = true
			result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("WINDOW_%d_ZERO_REQUESTS", w.WindowIndex))
			continue
		}

		// Insufficient requests per window
		if len(w.SuccessfulRequests) < policy.MinimumRequestsPerWindow {
			hasUnknown = true
			result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("WINDOW_%d_INSUFFICIENT_REQUESTS: got %d < min %d", w.WindowIndex, len(w.SuccessfulRequests), policy.MinimumRequestsPerWindow))
			continue
		}

		// Error fraction over all attempts in window
		if w.TotalAttempts > 0 {
			errFraction := float64(w.ErrorCount) / float64(w.TotalAttempts)
			if errFraction > policy.MaximumErrorFraction {
				result.Outcome = SLOOutcomeFail
				result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("WINDOW_%d_ERROR_FRACTION_EXCEEDED: %.4f > %.4f", w.WindowIndex, errFraction, policy.MaximumErrorFraction))
				return result, nil
			}
		}

		// Latency percentile over successful requests in window
		ordered := make([]float64, len(w.SuccessfulRequests))
		copy(ordered, w.SuccessfulRequests)
		sort.Float64s(ordered)

		latencyQuantile := ComputeQuantile(ordered, policy.LatencyPercentile)
		if latencyQuantile > maxLatency {
			result.Outcome = SLOOutcomeFail
			result.ReasonCodes = append(result.ReasonCodes, fmt.Sprintf("WINDOW_%d_LATENCY_EXCEEDED: p%.2f=%.2f > %.2f ms", w.WindowIndex, policy.LatencyPercentile, latencyQuantile, maxLatency))
			return result, nil
		}
	}

	if hasUnknown {
		result.Outcome = SLOOutcomeUnknown
	} else {
		result.Outcome = SLOOutcomePass
	}

	return result, nil
}

// CanonicalBenchmark evaluates both statistical performance and operational SLO,
// validating complete acceptance criteria per PERFORMANCE.md.
func EvaluateBenchmark(
	policy *PerformancePolicyRecord,
	trials []PairedTrial,
	seed int,
	rawSamplesRef string,
	sloPolicy *OperationalSLOPolicy,
	windows []WindowSample,
	sloPolicyRef string,
	rawWindowSamplesRef string,
) (*StatisticalBenchmarkResult, error) {
	statResult, err := EvaluateStatisticalPerformance(policy, trials, seed, rawSamplesRef)
	if err != nil {
		return nil, err
	}

	if sloPolicy != nil {
		sloResult, err := EvaluateOperationalSLO(sloPolicy, windows, sloPolicyRef, rawWindowSamplesRef)
		if err != nil {
			return nil, err
		}
		statResult.OperationalSLOResult = sloResult

		// If operational SLO fails or is unknown, it blocks acceptance:
		// "A requested SLO requires the exact policy ref and independently recomputed PASS
		//  in addition to statistical acceptance; FAIL/UNKNOWN blocks its release obligation."
		if sloResult.Outcome != SLOOutcomePass {
			if sloResult.Outcome == SLOOutcomeFail && statResult.Outcome != BenchmarkOutcomeRegression {
				statResult.Outcome = BenchmarkOutcomeInconclusive
			} else if sloResult.Outcome == SLOOutcomeUnknown && statResult.Outcome != BenchmarkOutcomeRegression {
				statResult.Outcome = BenchmarkOutcomeInconclusive
			}
			statResult.ReasonCodes = append(statResult.ReasonCodes, fmt.Sprintf("SLO_%s", sloResult.Outcome))
		}
	}

	return statResult, nil
}
