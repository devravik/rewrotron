package domain

import (
	"strings"
	"testing"
)

func makeTestPlanGraph() *ReconstructionPlanGraph {
	builder := NewPlanGraphBuilder(
		"proj-fixture",
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		nil,
	)
	_ = builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "unit-login",
		ContractRefs: []string{"sha256:1111111111111111111111111111111111111111111111111111111111111111"},
		CaseIDs:      []string{"case-login-1", "case-login-2"},
		Priority:     10,
	})
	_ = builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "unit-checkout",
		ContractRefs: []string{"sha256:2222222222222222222222222222222222222222222222222222222222222222"},
		CaseIDs:      []string{"case-checkout-1"},
		Priority:     5,
	})
	_ = builder.AddDependency("unit-login", "unit-checkout")
	graph, _ := builder.Build()
	return graph
}

// Test X35: Scope Deadlock & Whole-Project Corpus Acceptance
// "Login prerequisite completes on its UNIT+implemented regressions before checkout;
//  project missing checkout cannot VERIFIED/READY_CUTOVER"
func TestGate3Release_X35_WholeProjectScopeEnforcement(t *testing.T) {
	graph := makeTestPlanGraph()

	baseInputs := WholeProjectReleaseInputs{
		ProjectID:             "proj-fixture",
		ConfigurationHash:     "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		PlanHash:              "3ccae952a8f049ec692f413163bfb5828b25dcf42a1f0d80eac7cb6a121bcb2e",
		TargetTreeRef:         "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		WholeProjectReportRef: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BenchmarkRef:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		RunbookRef:            "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		RollbackRehearsalRef:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PlanGraph:             graph,
		ImplementedUnitIDs:    []string{"unit-login"}, // ONLY login is implemented; checkout is missing
		UnitVerdicts: map[string]string{
			"unit-login": "VERIFIED",
		},
		PerformanceBenchmark: &StatisticalBenchmarkResult{
			Outcome: BenchmarkOutcomeApprovedBudget,
		},
		RollbackRehearsal: makeValidRehearsal(),
	}

	// 1. Falsification (Test X35): Project missing checkout cannot compile Gate 3 release package
	_, _, err := CompileGate3ReleasePackage(baseInputs)
	if err == nil || !strings.Contains(err.Error(), "project missing unit unit-checkout") || !strings.Contains(err.Error(), "Test X35") {
		t.Fatalf("expected Test X35 refusal when checkout unit is missing, got: %v", err)
	}

	// 2. Falsification: All units implemented, but checkout is MISMATCH/not VERIFIED
	partialVerdicts := baseInputs
	partialVerdicts.ImplementedUnitIDs = []string{"unit-login", "unit-checkout"}
	partialVerdicts.UnitVerdicts = map[string]string{
		"unit-login":    "VERIFIED",
		"unit-checkout": "MISMATCH",
	}
	_, _, err = CompileGate3ReleasePackage(partialVerdicts)
	if err == nil || !strings.Contains(err.Error(), "unit unit-checkout is not VERIFIED") {
		t.Fatalf("expected refusal when checkout verdict is MISMATCH, got: %v", err)
	}

	// 3. Falsification: Benchmark is INCONCLUSIVE
	inconclusiveBenchmark := baseInputs
	inconclusiveBenchmark.ImplementedUnitIDs = []string{"unit-login", "unit-checkout"}
	inconclusiveBenchmark.UnitVerdicts = map[string]string{
		"unit-login":    "VERIFIED",
		"unit-checkout": "VERIFIED",
	}
	inconclusiveBenchmark.PerformanceBenchmark = &StatisticalBenchmarkResult{
		Outcome: BenchmarkOutcomeInconclusive,
	}
	_, _, err = CompileGate3ReleasePackage(inconclusiveBenchmark)
	if err == nil || !strings.Contains(err.Error(), "performance benchmark outcome is INCONCLUSIVE") {
		t.Fatalf("expected refusal when performance benchmark is INCONCLUSIVE, got: %v", err)
	}

	// 4. Positive Control: Full project accepted (both login and checkout VERIFIED), performance accepted, rollback verified
	fullProjectInputs := baseInputs
	fullProjectInputs.ImplementedUnitIDs = []string{"unit-login", "unit-checkout"}
	fullProjectInputs.UnitVerdicts = map[string]string{
		"unit-login":    "VERIFIED",
		"unit-checkout": "VERIFIED",
	}

	pkg, pkgBytes, err := CompileGate3ReleasePackage(fullProjectInputs)
	if err != nil {
		t.Fatalf("failed to compile valid Gate 3 release package: %v", err)
	}
	if pkg.SchemaVersion != "2.0" || pkg.ProjectID != "proj-fixture" {
		t.Fatalf("unexpected release package metadata: %+v", pkg)
	}
	if len(pkgBytes) == 0 {
		t.Fatalf("empty compiled release package bytes")
	}

	// 5. Verify the compiled release package against ValidateMigrationRelease
	if err := ValidateMigrationRelease(pkgBytes); err != nil {
		t.Fatalf("ValidateMigrationRelease failed on compiled package: %v", err)
	}
}
