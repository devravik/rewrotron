package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Profiler constants and pattern types
const (
	PatternNPlusOneQuery = "N_PLUS_ONE_QUERY"
	PatternQuerySpike    = "QUERY_SPIKE"
	PatternSlowQuery     = "SLOW_QUERY"
	PatternLargePayload  = "LARGE_RESPONSE_PAYLOAD"
)

// DBQuerySpan represents a single captured database query execution within a request.
type DBQuerySpan struct {
	QueryID      string        `json:"query_id"`
	SQL          string        `json:"sql"`
	Normalized   string        `json:"normalized_sql"`
	Duration     time.Duration `json:"duration"`
	RowsAffected int64         `json:"rows_affected"`
	Error        error         `json:"-"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// RequestProfile represents the collected metrics for a single request execution.
type RequestProfile struct {
	RequestID    string        `json:"request_id"`
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	StatusCode   int           `json:"status_code"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	ResponseSize int64         `json:"response_size_bytes"`
	DBQueries    []DBQuerySpan `json:"db_queries"`
	PeakRSSBytes int64         `json:"peak_rss_bytes"`
	CPUTime      time.Duration `json:"cpu_time"`
}

// TotalDBDuration returns the cumulative duration of all database queries in the request.
func (rp *RequestProfile) TotalDBDuration() time.Duration {
	var total time.Duration
	for _, q := range rp.DBQueries {
		total += q.Duration
	}
	return total
}

// TotalDBQueryCount returns the number of database queries executed in the request.
func (rp *RequestProfile) TotalDBQueryCount() int {
	return len(rp.DBQueries)
}

// QueryPatternAnomaly represents an identified query inefficiency or anti-pattern.
type QueryPatternAnomaly struct {
	PatternType string `json:"pattern_type"`
	Normalized  string `json:"normalized_sql"`
	Occurrences int    `json:"occurrences"`
	Description string `json:"description"`
}

// DetectQueryAnomalies inspects the captured DB queries in a request profile
// and identifies potential N+1 query loops and query spikes.
func (rp *RequestProfile) DetectQueryAnomalies(nPlusOneThreshold int) []QueryPatternAnomaly {
	if nPlusOneThreshold <= 0 {
		nPlusOneThreshold = 3
	}

	counts := make(map[string]int)
	for _, q := range rp.DBQueries {
		norm := q.Normalized
		if norm == "" {
			norm = NormalizeSQL(q.SQL)
		}
		counts[norm]++
	}

	var anomalies []QueryPatternAnomaly
	for norm, count := range counts {
		if count >= nPlusOneThreshold {
			anomalies = append(anomalies, QueryPatternAnomaly{
				PatternType: PatternNPlusOneQuery,
				Normalized:  norm,
				Occurrences: count,
				Description: fmt.Sprintf("detected %d identical normalized queries in single request (exceeds threshold %d): potential N+1 loop", count, nPlusOneThreshold),
			})
		}
	}
	return anomalies
}

// SQL normalization regex patterns
var (
	reNumbers     = regexp.MustCompile(`\b\d+\b`)
	reStrings     = regexp.MustCompile(`'[^']*'`)
	reInList      = regexp.MustCompile(`(?i)\bIN\s*\([^)]+\)`)
	reWhitespace  = regexp.MustCompile(`\s+`)
)

// NormalizeSQL strips literal values, IDs, and in-lists to produce a canonical parameterized query pattern.
func NormalizeSQL(sql string) string {
	s := strings.TrimSpace(sql)
	s = reStrings.ReplaceAllString(s, "?")
	s = reInList.ReplaceAllString(s, "IN (?)")
	s = reNumbers.ReplaceAllString(s, "?")
	s = reWhitespace.ReplaceAllString(s, " ")
	return strings.ToUpper(s)
}

// TrialProfileSummary aggregates request profiles within a single measurement trial.
type TrialProfileSummary struct {
	TrialIndex         int     `json:"trial_index"`
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	MeanDBQueryCount   float64 `json:"mean_db_query_count"`
	MeanDBLatencyMs    float64 `json:"mean_db_latency_ms"`
	MeanResponseBytes  float64 `json:"mean_response_bytes"`
	MeanLatencyMs      float64 `json:"mean_latency_ms"`
	P50LatencyMs       float64 `json:"p50_latency_ms"`
	P95LatencyMs       float64 `json:"p95_latency_ms"`
	P99LatencyMs       float64 `json:"p99_latency_ms"`
	PeakRSSBytes       float64 `json:"peak_rss_bytes"`
	CPUTimeSeconds     float64 `json:"cpu_time_seconds"`
}

// Profiler defines the measurement contract for profiling execution workloads.
type Profiler interface {
	StartRequest(ctx context.Context, endpoint, method string) *RequestContext
	RecordQuery(ctx context.Context, sql string, duration time.Duration, rowsAffected int64, err error)
	FinishRequest(ctx context.Context, rc *RequestContext, statusCode int, responseSize int64, err error) *RequestProfile
	SummarizeTrial(trialIndex int, profiles []*RequestProfile) *TrialProfileSummary
}

// RequestContext holds active trace state for an in-flight request being profiled.
type RequestContext struct {
	RequestID string
	Endpoint  string
	Method    string
	StartTime time.Time
	Queries   []DBQuerySpan
	mu        sync.Mutex
}

// RecordQuery appends a captured database query span in a thread-safe manner.
func (rc *RequestContext) RecordQuery(sql string, duration time.Duration, rowsAffected int64, err error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	span := DBQuerySpan{
		QueryID:      fmt.Sprintf("q-%d", len(rc.Queries)+1),
		SQL:          sql,
		Normalized:   NormalizeSQL(sql),
		Duration:     duration,
		RowsAffected: rowsAffected,
		Error:        err,
		ErrorMessage: errMsg,
	}
	rc.Queries = append(rc.Queries, span)
}

// ExecutionProfiler provides an in-memory profiler implementation adhering to PERFORMANCE.md metric rules.
type ExecutionProfiler struct {
	mu           sync.RWMutex
	requestSeq   int64
	activeSpanMu sync.Mutex
}

// NewExecutionProfiler constructs a new execution profiler.
func NewExecutionProfiler() *ExecutionProfiler {
	return &ExecutionProfiler{}
}

// StartRequest begins profiling a single request.
func (p *ExecutionProfiler) StartRequest(ctx context.Context, endpoint, method string) *RequestContext {
	p.mu.Lock()
	p.requestSeq++
	reqID := fmt.Sprintf("req-%d", p.requestSeq)
	p.mu.Unlock()

	return &RequestContext{
		RequestID: reqID,
		Endpoint:  endpoint,
		Method:    method,
		StartTime: time.Now(),
		Queries:   make([]DBQuerySpan, 0),
	}
}

// RecordQuery captures a DB query for the context.
func (p *ExecutionProfiler) RecordQuery(ctx context.Context, sql string, duration time.Duration, rowsAffected int64, err error) {
	// If a RequestContext is passed in ctx or directly handled, record it.
}

// FinishRequest completes request profiling and returns the RequestProfile.
func (p *ExecutionProfiler) FinishRequest(
	ctx context.Context,
	rc *RequestContext,
	statusCode int,
	responseSize int64,
	err error,
) *RequestProfile {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	duration := time.Since(rc.StartTime)
	success := err == nil && statusCode < 400

	queriesCopy := make([]DBQuerySpan, len(rc.Queries))
	copy(queriesCopy, rc.Queries)

	return &RequestProfile{
		RequestID:    rc.RequestID,
		Endpoint:     rc.Endpoint,
		Method:       rc.Method,
		StatusCode:   statusCode,
		Success:      success,
		Duration:     duration,
		ResponseSize: responseSize,
		DBQueries:    queriesCopy,
	}
}

// SummarizeTrial aggregates individual request profiles into a TrialProfileSummary
// following PERFORMANCE.md rules:
// - Latency metrics (p50/p95/p99) use nearest-rank ceil(p*n) on successful samples.
// - Counts, DB latency, and response size are per-success means.
// - Peak RSS is maximum peak observed.
func (p *ExecutionProfiler) SummarizeTrial(trialIndex int, profiles []*RequestProfile) *TrialProfileSummary {
	summary := &TrialProfileSummary{
		TrialIndex:    trialIndex,
		TotalRequests: len(profiles),
	}

	var successfulLatencies []float64
	var sumDBQueries float64
	var sumDBDurationMs float64
	var sumResponseBytes float64
	var sumDurationMs float64
	var maxRSS int64
	var sumCPUTime time.Duration

	for _, prof := range profiles {
		if prof.Success {
			summary.SuccessfulRequests++
			durMs := float64(prof.Duration.Microseconds()) / 1000.0
			successfulLatencies = append(successfulLatencies, durMs)
			sumDurationMs += durMs

			sumDBQueries += float64(prof.TotalDBQueryCount())
			sumDBDurationMs += float64(prof.TotalDBDuration().Microseconds()) / 1000.0
			sumResponseBytes += float64(prof.ResponseSize)
			sumCPUTime += prof.CPUTime
			if prof.PeakRSSBytes > maxRSS {
				maxRSS = prof.PeakRSSBytes
			}
		} else {
			summary.FailedRequests++
		}
	}

	nSuccess := summary.SuccessfulRequests
	if nSuccess > 0 {
		fSuccess := float64(nSuccess)
		summary.MeanDBQueryCount = sumDBQueries / fSuccess
		summary.MeanDBLatencyMs = sumDBDurationMs / fSuccess
		summary.MeanResponseBytes = sumResponseBytes / fSuccess
		summary.MeanLatencyMs = sumDurationMs / fSuccess
		summary.PeakRSSBytes = float64(maxRSS)
		summary.CPUTimeSeconds = sumCPUTime.Seconds() / fSuccess

		// Compute nearest-rank quantiles on sorted successful samples
		sorted := make([]float64, len(successfulLatencies))
		copy(sorted, successfulLatencies)
		// insertion sort / standard sort
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}

		summary.P50LatencyMs = ComputeQuantile(sorted, 0.50)
		summary.P95LatencyMs = ComputeQuantile(sorted, 0.95)
		summary.P99LatencyMs = ComputeQuantile(sorted, 0.99)
	}

	return summary
}
