package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// PerformanceOptimizationMission represents a generated candidate optimization mission.
// It maps directly to specs/missions/mission.schema.json with type "OPTIMIZE".
type PerformanceOptimizationMission struct {
	MissionID       string                 `json:"mission_id"`
	Type            string                 `json:"type"` // "OPTIMIZE"
	Priority        int                    `json:"priority"`
	Status          string                 `json:"status"` // "QUEUED"
	UnitID          *string                `json:"unit_id"`
	CaseID          *string                `json:"case_id"`
	Description     string                 `json:"description"`
	Context         map[string]interface{} `json:"context"`
	AttemptCount    int                    `json:"attempt_count"`
	FencingToken    int                    `json:"fencing_token"`
	LeaseExpiresAt  *string                `json:"lease_expires_at"`
}

// GenerateOptimizationMission creates an auditable candidate optimization mission
// from an identified anomaly (e.g. N+1 query loop) or measured bottleneck.
// Invariant per MISSION-PROTOCOL and M8:
// "A performance finding may produce a candidate investigation or optimization mission.
//  It must not automatically authorize destructive changes, deployment, or an unapproved migration."
func GenerateOptimizationMission(
	projectID string,
	unitID string,
	targetComponent string,
	anomaly QueryPatternAnomaly,
	baselineLatencyMs float64,
	baselineQueries int,
	evidenceRef string,
) (*PerformanceOptimizationMission, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if unitID == "" {
		return nil, fmt.Errorf("unitID is required")
	}

	missionID := fmt.Sprintf("mission-opt-%s-%d", unitID, time.Now().UnixNano())
	desc := fmt.Sprintf("Optimize component %s: %s (baseline: %d queries, %.2f ms)",
		targetComponent, anomaly.Description, baselineQueries, baselineLatencyMs)

	ctxData := map[string]interface{}{
		"project_id":          projectID,
		"target_component":    targetComponent,
		"pattern_type":        anomaly.PatternType,
		"normalized_sql":      anomaly.Normalized,
		"occurrences":         anomaly.Occurrences,
		"baseline_queries":    baselineQueries,
		"baseline_latency_ms": baselineLatencyMs,
		"evidence_ref":        evidenceRef,
		"requires_gate2":      true,
		"auto_deploy":         false,
	}

	uID := unitID
	return &PerformanceOptimizationMission{
		MissionID:      missionID,
		Type:           "OPTIMIZE",
		Priority:       10,
		Status:         "QUEUED",
		UnitID:         &uID,
		CaseID:         nil,
		Description:    desc,
		Context:        ctxData,
		AttemptCount:   0,
		FencingToken:   0,
		LeaseExpiresAt: nil,
	}, nil
}

// PerformanceOptimizationReport compiles an auditable optimization report comparing
// baseline measurements with the candidate optimization.
type PerformanceOptimizationReport struct {
	ReportID             string                            `json:"report_id"`
	ProjectID            string                            `json:"project_id"`
	TargetComponent      string                            `json:"target_component"`
	BaselineBuildID      string                            `json:"baseline_build_id"`
	CandidateBuildID     string                            `json:"candidate_build_id"`
	WorkloadRef          string                            `json:"workload_ref"`
	EnvironmentRef       string                            `json:"environment_ref"`
	PerformancePolicyRef string                            `json:"performance_policy_ref"`
	BehaviorVerdict      string                            `json:"behavior_verdict"` // PASS, MISMATCH, UNKNOWN
	BenchmarkOutcome     string                            `json:"benchmark_outcome"` // NON_REGRESSION, APPROVED_BUDGET, REGRESSION, INCONCLUSIVE
	Accepted             bool                              `json:"accepted"`
	RejectionReason      string                            `json:"rejection_reason,omitempty"`
	BaselineMetrics      map[string]float64                `json:"baseline_metrics"`
	CandidateMetrics     map[string]float64                `json:"candidate_metrics"`
	MetricEvaluations    map[string]MetricEvaluationResult `json:"metric_evaluations"`
	CreatedAt            string                            `json:"created_at"`
	ReportDigest         string                            `json:"report_digest"`
}

// CompileOptimizationReport creates a cryptographic, auditable optimization report
// documenting baseline vs candidate metrics, behavioral parity, and final acceptance.
func CompileOptimizationReport(
	projectID string,
	targetComponent string,
	baselineBuildID string,
	candidateBuildID string,
	workloadRef string,
	envRef string,
	policyRef string,
	decision *OptimizationDecision,
	baselineMetrics map[string]float64,
	candidateMetrics map[string]float64,
) (*PerformanceOptimizationReport, error) {
	if decision == nil {
		return nil, fmt.Errorf("decision is nil")
	}

	reportID := fmt.Sprintf("opt-rep-%s-%d", projectID, time.Now().UnixNano())
	createdAt := time.Now().UTC().Format(time.RFC3339)

	metricEvals := make(map[string]MetricEvaluationResult)
	if decision.BenchmarkResult != nil {
		metricEvals = decision.BenchmarkResult.Metrics
	}

	rep := &PerformanceOptimizationReport{
		ReportID:             reportID,
		ProjectID:            projectID,
		TargetComponent:      targetComponent,
		BaselineBuildID:      baselineBuildID,
		CandidateBuildID:     candidateBuildID,
		WorkloadRef:          workloadRef,
		EnvironmentRef:       envRef,
		PerformancePolicyRef: policyRef,
		BehaviorVerdict:      decision.BehaviorVerdict,
		BenchmarkOutcome:     decision.BenchmarkOutcome,
		Accepted:             decision.Accepted,
		RejectionReason:      decision.RejectionReason,
		BaselineMetrics:      baselineMetrics,
		CandidateMetrics:     candidateMetrics,
		MetricEvaluations:    metricEvals,
		CreatedAt:            createdAt,
	}

	// Compute canonical digest
	bytes, err := json.Marshal(rep)
	if err != nil {
		return nil, fmt.Errorf("marshal report failed: %w", err)
	}
	hash := sha256.Sum256(bytes)
	rep.ReportDigest = fmt.Sprintf("sha256:%x", hash)

	return rep, nil
}
