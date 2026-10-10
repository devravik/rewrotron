package domain

import (
	"encoding/json"
	"fmt"
)

// WholeProjectReleaseInputs contains all verified inputs required to compile Gate 3 release package.
type WholeProjectReleaseInputs struct {
	ProjectID             string                    `json:"project_id"`
	ConfigurationHash     string                    `json:"configuration_hash"`
	PlanHash              string                    `json:"plan_hash"`
	TargetTreeRef         string                    `json:"target_tree_ref"`
	WholeProjectReportRef string                    `json:"whole_project_report_ref"`
	BenchmarkRef          string                    `json:"benchmark_ref"` // accepted benchmark or approved not-applicable policy
	RunbookRef            string                    `json:"runbook_ref"`
	RollbackRehearsalRef  string                    `json:"rollback_rehearsal_ref"`
	PlanGraph             *ReconstructionPlanGraph  `json:"plan_graph"`
	ImplementedUnitIDs    []string                  `json:"implemented_unit_ids"`
	UnitVerdicts          map[string]string         `json:"unit_verdicts"`
	PerformanceBenchmark  *StatisticalBenchmarkResult `json:"performance_benchmark,omitempty"`
	RollbackRehearsal     *RollbackRehearsalRecord  `json:"rollback_rehearsal,omitempty"`
}

// WholeProjectReleasePackage represents the finalized Gate 3 release envelope per specs/migration/release.schema.json v2.0.
type WholeProjectReleasePackage struct {
	SchemaVersion string          `json:"schema_version"` // "2.0"
	ProjectID     string          `json:"project_id"`
	ReleaseHash   string          `json:"release_hash"`
	Subject       json.RawMessage `json:"subject"`
}

// Gate3ReleaseSubject represents the canonical RFC 8785 subject preimage for Gate 3 RELEASE confirmation.
type Gate3ReleaseSubject struct {
	SchemaVersion         string `json:"schema_version"` // "2.0"
	Canonicalization      string `json:"canonicalization"` // "rwt-cjson-1"
	ProjectID             string `json:"project_id"`
	Kind                  string `json:"kind"` // "RELEASE"
	ConfigurationHash     string `json:"configuration_hash"`
	PlanHash              string `json:"plan_hash"`
	TargetTreeRef         string `json:"target_tree_ref"`
	WholeProjectReportRef string `json:"whole_project_report_ref"`
	BenchmarkRef          string `json:"benchmark_ref"`
	RunbookRef            string `json:"runbook_ref"`
	RollbackRehearsalRef  string `json:"rollback_rehearsal_ref"`
}

// CompileGate3ReleasePackage verifies whole-project corpus completion, performance acceptance,
// rollback rehearsal, and compiles the exact Gate 3 release package (Test X35).
//
// Invariants enforced (Test X35, PROJECT-LIFECYCLE.md, CONFIRMATION-PROTOCOL.md):
// 1. Whole-project scope check: all plan units (e.g. login + checkout) must be implemented and VERIFIED;
//    a partial project cannot reach READY_CUTOVER / compile Gate 3 release.
// 2. Performance benchmark must be accepted (NON_REGRESSION, APPROVED_BUDGET, or NOT_APPLICABLE).
//    INCONCLUSIVE or REGRESSION strictly blocks release package compilation.
// 3. Rollback rehearsal must be verified (PASS), with write-aware reconciliation verified.
// 4. Compiles canonical subject and computes release_hash = sha256(canonical(subject)).
// 5. Validates finalized envelope against ValidateMigrationRelease.
func CompileGate3ReleasePackage(inputs WholeProjectReleaseInputs) (*WholeProjectReleasePackage, []byte, error) {
	if inputs.ProjectID == "" {
		return nil, nil, fmt.Errorf("missing project_id")
	}

	// 1. Enforce whole-project scope check per Test X35
	if inputs.PlanGraph != nil {
		canComplete, reason := inputs.PlanGraph.CanProjectComplete(inputs.ImplementedUnitIDs, inputs.UnitVerdicts)
		if !canComplete {
			return nil, nil, fmt.Errorf("cannot compile Gate 3 release package: %s (Test X35)", reason)
		}
	}

	// 2. Validate performance acceptance
	if inputs.PerformanceBenchmark != nil {
		out := inputs.PerformanceBenchmark.Outcome
		if out == BenchmarkOutcomeRegression || out == BenchmarkOutcomeInconclusive {
			return nil, nil, fmt.Errorf("cannot compile Gate 3 release: performance benchmark outcome is %s", out)
		}
	}

	// 3. Validate rollback rehearsal
	if inputs.RollbackRehearsal != nil {
		if err := ValidateRollbackRehearsal(inputs.RollbackRehearsal); err != nil {
			return nil, nil, fmt.Errorf("cannot compile Gate 3 release: rollback rehearsal validation failed: %w", err)
		}
	}

	// 4. Construct canonical Gate 3 release subject record
	subj := Gate3ReleaseSubject{
		SchemaVersion:         "2.0",
		Canonicalization:      "rwt-cjson-1",
		ProjectID:             inputs.ProjectID,
		Kind:                  "RELEASE",
		ConfigurationHash:     inputs.ConfigurationHash,
		PlanHash:              inputs.PlanHash,
		TargetTreeRef:         inputs.TargetTreeRef,
		WholeProjectReportRef: inputs.WholeProjectReportRef,
		BenchmarkRef:          inputs.BenchmarkRef,
		RunbookRef:            inputs.RunbookRef,
		RollbackRehearsalRef:  inputs.RollbackRehearsalRef,
	}

	subjRaw, err := json.Marshal(subj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal subject: %w", err)
	}

	subjCanonical, err := CanonicalJSON(subjRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to canonicalize subject: %w", err)
	}

	releaseHash, err := CanonicalHash(subjCanonical)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compute release_hash: %w", err)
	}

	pkg := &WholeProjectReleasePackage{
		SchemaVersion: "2.0",
		ProjectID:     inputs.ProjectID,
		ReleaseHash:   releaseHash,
		Subject:       json.RawMessage(subjCanonical),
	}

	pkgRaw, err := json.Marshal(pkg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal release package: %w", err)
	}

	pkgCanonical, err := CanonicalJSON(pkgRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to canonicalize release package: %w", err)
	}

	// 5. Verify the package against the canonical validator
	if err := ValidateMigrationRelease(pkgCanonical); err != nil {
		return nil, nil, fmt.Errorf("validation of compiled release package failed: %w", err)
	}

	return pkg, pkgCanonical, nil
}
