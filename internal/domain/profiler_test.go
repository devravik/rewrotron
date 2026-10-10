package domain

import (
	"context"
	"testing"
	"time"
)

func TestProfiler_NormalizeSQL(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "SELECT * FROM users WHERE id = 42",
			expected: "SELECT * FROM USERS WHERE ID = ?",
		},
		{
			input:    "SELECT name, email FROM users WHERE id IN (1, 2, 3, 4)",
			expected: "SELECT NAME, EMAIL FROM USERS WHERE ID IN (?)",
		},
		{
			input:    "UPDATE accounts SET status = 'active' WHERE id = 100",
			expected: "UPDATE ACCOUNTS SET STATUS = ? WHERE ID = ?",
		},
		{
			input:    "SELECT *   FROM  orders \n WHERE customer_id = 99",
			expected: "SELECT * FROM ORDERS WHERE CUSTOMER_ID = ?",
		},
	}

	for _, tc := range cases {
		got := NormalizeSQL(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeSQL(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestProfiler_NPlusOneDetection(t *testing.T) {
	ctx := context.Background()
	profiler := NewExecutionProfiler()

	rc := profiler.StartRequest(ctx, "/api/orders", "GET")

	// 1 main query
	rc.RecordQuery("SELECT * FROM orders WHERE status = 'shipped'", 5*time.Millisecond, 10, nil)

	// 5 repeated item queries (N+1 anti-pattern)
	for i := 1; i <= 5; i++ {
		rc.RecordQuery("SELECT * FROM order_items WHERE order_id = "+string(rune('0'+i)), 2*time.Millisecond, 2, nil)
	}

	profile := profiler.FinishRequest(ctx, rc, 200, 4096, nil)

	if profile.TotalDBQueryCount() != 6 {
		t.Fatalf("expected 6 DB queries, got %d", profile.TotalDBQueryCount())
	}

	anomalies := profile.DetectQueryAnomalies(3)
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly detected, got %d", len(anomalies))
	}

	if anomalies[0].PatternType != PatternNPlusOneQuery {
		t.Errorf("expected %s, got %s", PatternNPlusOneQuery, anomalies[0].PatternType)
	}
	if anomalies[0].Occurrences != 5 {
		t.Errorf("expected 5 occurrences, got %d", anomalies[0].Occurrences)
	}
}

func TestProfiler_SummarizeTrial(t *testing.T) {
	profiler := NewExecutionProfiler()

	profiles := make([]*RequestProfile, 100)
	for i := 0; i < 100; i++ {
		dbQueries := []DBQuerySpan{
			{
				QueryID:    "q1",
				SQL:        "SELECT * FROM items WHERE id = ?",
				Normalized: "SELECT * FROM ITEMS WHERE ID = ?",
				Duration:   time.Duration(i%5+1) * time.Millisecond,
			},
			{
				QueryID:    "q2",
				SQL:        "SELECT * FROM meta WHERE item_id = ?",
				Normalized: "SELECT * FROM META WHERE ITEM_ID = ?",
				Duration:   time.Duration(i%3+1) * time.Millisecond,
			},
		}

		profiles[i] = &RequestProfile{
			RequestID:    "req",
			StatusCode:   200,
			Success:      true,
			Duration:     time.Duration(10+i) * time.Millisecond,
			ResponseSize: 1024,
			DBQueries:    dbQueries,
			PeakRSSBytes: 1024 * 1024 * 50,
			CPUTime:      5 * time.Millisecond,
		}
	}

	// Add 5 failed requests
	for i := 0; i < 5; i++ {
		profiles = append(profiles, &RequestProfile{
			RequestID:  "req-fail",
			StatusCode: 500,
			Success:    false,
			Duration:   100 * time.Millisecond,
		})
	}

	summary := profiler.SummarizeTrial(1, profiles)

	if summary.TotalRequests != 105 {
		t.Errorf("expected 105 total requests, got %d", summary.TotalRequests)
	}
	if summary.SuccessfulRequests != 100 {
		t.Errorf("expected 100 successes, got %d", summary.SuccessfulRequests)
	}
	if summary.FailedRequests != 5 {
		t.Errorf("expected 5 failures, got %d", summary.FailedRequests)
	}

	// Each successful request has 2 queries
	if summary.MeanDBQueryCount != 2.0 {
		t.Errorf("expected mean query count 2.0, got %f", summary.MeanDBQueryCount)
	}
	if summary.MeanResponseBytes != 1024.0 {
		t.Errorf("expected mean response bytes 1024.0, got %f", summary.MeanResponseBytes)
	}
	if summary.P50LatencyMs == 0 || summary.P95LatencyMs == 0 || summary.P99LatencyMs == 0 {
		t.Errorf("expected non-zero latency quantiles")
	}
	if summary.P50LatencyMs > summary.P95LatencyMs || summary.P95LatencyMs > summary.P99LatencyMs {
		t.Errorf("invalid quantile ordering: p50=%f, p95=%f, p99=%f",
			summary.P50LatencyMs, summary.P95LatencyMs, summary.P99LatencyMs)
	}
}

func TestProfiler_IntegrationWithPairedTrial(t *testing.T) {
	profiler := NewExecutionProfiler()

	// Simulate source trial (e.g. 5 DB queries per request, 20ms DB latency, 500 requests)
	sourceProfiles := make([]*RequestProfile, 500)
	for i := 0; i < 500; i++ {
		sourceProfiles[i] = &RequestProfile{
			Success:      true,
			Duration:     30 * time.Millisecond,
			ResponseSize: 2048,
			DBQueries: []DBQuerySpan{
				{Duration: 4 * time.Millisecond},
				{Duration: 4 * time.Millisecond},
				{Duration: 4 * time.Millisecond},
				{Duration: 4 * time.Millisecond},
				{Duration: 4 * time.Millisecond},
			},
		}
	}
	sourceSumm := profiler.SummarizeTrial(0, sourceProfiles)

	// Simulate target trial (e.g. optimized: 2 DB queries per request, 8ms DB latency)
	targetProfiles := make([]*RequestProfile, 500)
	for i := 0; i < 500; i++ {
		targetProfiles[i] = &RequestProfile{
			Success:      true,
			Duration:     15 * time.Millisecond,
			ResponseSize: 2048,
			DBQueries: []DBQuerySpan{
				{Duration: 4 * time.Millisecond},
				{Duration: 4 * time.Millisecond},
			},
		}
	}
	targetSumm := profiler.SummarizeTrial(0, targetProfiles)

	// Convert summaries to PairedTrial metric values
	pairedTrial := PairedTrial{
		TrialIndex: 0,
		SourceMetricValues: map[string]float64{
			MetricDBQueryCount: sourceSumm.MeanDBQueryCount,
			MetricDBLatency:    sourceSumm.MeanDBLatencyMs,
			MetricResponseSize: sourceSumm.MeanResponseBytes,
			MetricP95:          sourceSumm.P95LatencyMs,
		},
		TargetMetricValues: map[string]float64{
			MetricDBQueryCount: targetSumm.MeanDBQueryCount,
			MetricDBLatency:    targetSumm.MeanDBLatencyMs,
			MetricResponseSize: targetSumm.MeanResponseBytes,
			MetricP95:          targetSumm.P95LatencyMs,
		},
		SourceSuccesses: sourceSumm.SuccessfulRequests,
		TargetSuccesses: targetSumm.SuccessfulRequests,
		SourceErrors:    sourceSumm.FailedRequests,
		TargetErrors:    targetSumm.FailedRequests,
	}

	if pairedTrial.SourceMetricValues[MetricDBQueryCount] != 5.0 {
		t.Errorf("expected source query count 5.0, got %f", pairedTrial.SourceMetricValues[MetricDBQueryCount])
	}
	if pairedTrial.TargetMetricValues[MetricDBQueryCount] != 2.0 {
		t.Errorf("expected target query count 2.0, got %f", pairedTrial.TargetMetricValues[MetricDBQueryCount])
	}
	if pairedTrial.SourceMetricValues[MetricDBLatency] != 20.0 {
		t.Errorf("expected source DB latency 20.0ms, got %f", pairedTrial.SourceMetricValues[MetricDBLatency])
	}
	if pairedTrial.TargetMetricValues[MetricDBLatency] != 8.0 {
		t.Errorf("expected target DB latency 8.0ms, got %f", pairedTrial.TargetMetricValues[MetricDBLatency])
	}
}
