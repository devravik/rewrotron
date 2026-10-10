package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestOptimizationReport_GenerateMission verifies that an identified N+1 anomaly generates
// a well-formed OPTIMIZE mission adhering to specs/missions/mission.schema.json rules.
func TestOptimizationReport_GenerateMission(t *testing.T) {
	anomaly := QueryPatternAnomaly{
		PatternType: PatternNPlusOneQuery,
		Normalized:  "SELECT * FROM USERS WHERE ID = ?",
		Occurrences: 12,
		Description: "detected 12 identical normalized queries in single request (exceeds threshold 3): potential N+1 loop",
	}

	mission, err := GenerateOptimizationMission(
		"proj-123",
		"unit-catalog",
		"CatalogController",
		anomaly,
		65.5,
		14,
		"sha256:evidence111",
	)
	if err != nil {
		t.Fatalf("GenerateOptimizationMission failed: %v", err)
	}

	if mission.Type != "OPTIMIZE" {
		t.Errorf("expected mission type OPTIMIZE, got %s", mission.Type)
	}
	if mission.Status != "QUEUED" {
		t.Errorf("expected status QUEUED, got %s", mission.Status)
	}
	if *mission.UnitID != "unit-catalog" {
		t.Errorf("expected unit-catalog, got %s", *mission.UnitID)
	}
	if !strings.Contains(mission.Description, "CatalogController") {
		t.Errorf("description should mention CatalogController: %s", mission.Description)
	}

	// Verify context flags: auto_deploy must be false, requires_gate2 must be true
	if mission.Context["auto_deploy"] != false {
		t.Errorf("expected auto_deploy=false")
	}
	if mission.Context["requires_gate2"] != true {
		t.Errorf("expected requires_gate2=true")
	}

	// Verify JSON marshal against schema requirements
	bytes, err := json.Marshal(mission)
	if err != nil {
		t.Fatalf("failed to marshal mission: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		t.Fatalf("unmarshal raw failed: %v", err)
	}

	for _, requiredField := range []string{"mission_id", "type", "priority", "status", "unit_id", "case_id", "description", "context", "attempt_count", "fencing_token", "lease_expires_at"} {
		if _, ok := raw[requiredField]; !ok {
			t.Errorf("missing required mission schema field: %s", requiredField)
		}
	}
}

// TestOptimizationReport_CompileReport verifies that a compiled optimization report
// records baseline and candidate metrics, behavioral outcomes, and report digest.
func TestOptimizationReport_CompileReport(t *testing.T) {
	decision := &OptimizationDecision{
		CandidateID:      "cand-1",
		Accepted:         true,
		BehaviorVerdict:  "PASS",
		BenchmarkOutcome: BenchmarkOutcomeNonRegression,
		BenchmarkResult: &StatisticalBenchmarkResult{
			Outcome: BenchmarkOutcomeNonRegression,
			Claim:   "IMPROVEMENT",
			Metrics: map[string]MetricEvaluationResult{
				MetricP95: {
					MetricID: "P95",
					Unit:     "ms",
					Required: true,
					Outcome:  BenchmarkOutcomeNonRegression,
				},
			},
		},
	}

	baselineMetrics := map[string]float64{
		MetricP95:          50.0,
		MetricDBQueryCount: 6.0,
	}
	candidateMetrics := map[string]float64{
		MetricP95:          18.0,
		MetricDBQueryCount: 2.0,
	}

	report, err := CompileOptimizationReport(
		"proj-123",
		"CatalogController",
		"build-src-v1",
		"build-tgt-opt-v2",
		"sha256:workload1",
		"sha256:env1",
		"sha256:policy1",
		decision,
		baselineMetrics,
		candidateMetrics,
	)
	if err != nil {
		t.Fatalf("CompileOptimizationReport failed: %v", err)
	}

	if !report.Accepted {
		t.Errorf("expected report.Accepted=true")
	}
	if report.BehaviorVerdict != "PASS" {
		t.Errorf("expected PASS, got %s", report.BehaviorVerdict)
	}
	if !strings.HasPrefix(report.ReportDigest, "sha256:") {
		t.Errorf("expected sha256 prefix for report digest: %s", report.ReportDigest)
	}
	if report.BaselineMetrics[MetricDBQueryCount] != 6.0 {
		t.Errorf("expected baseline query count 6.0, got %f", report.BaselineMetrics[MetricDBQueryCount])
	}
	if report.CandidateMetrics[MetricDBQueryCount] != 2.0 {
		t.Errorf("expected candidate query count 2.0, got %f", report.CandidateMetrics[MetricDBQueryCount])
	}
}
