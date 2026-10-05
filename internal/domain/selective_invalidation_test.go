package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/devravik/rewrotron/internal/domain"
)

// Test X29: Typed Invalidation
// "PHP8.3→8.4 at 70%: preserve unchanged source discovery, invalidate relevant dependency/target/verification;
//  unknown edge conservative; no old PASS relabel"
func TestSelectiveInvalidation_X29_TypedInvalidationTargetRuntimeDrift(t *testing.T) {
	// Setup graph simulating 70% completion of migration from CI3 to Laravel:
	// Nodes:
	// - source_ir: CI3 Source IR (independent of target)
	// - source_contract: Contract extracted from CI3 source (independent of target)
	// - php83_runtime: Target PHP 8.3 runtime ref
	// - laravel_dep: Composer dependency resolution finding for PHP 8.3
	// - login_comp: Target model component for Login on PHP 8.3
	// - login_unit: Schedulable unit for Login on PHP 8.3
	// - login_exec_pass: Execution result (PASS) for Login on PHP 8.3
	// - unknown_edge_item: An item declared in unknown_impact_refs
	sourceIR := domain.MakeDummyRef("source_ir_ci3")
	sourceContract := domain.MakeDummyRef("source_contract_login")
	php83Runtime := domain.MakeDummyRef("runtime_php_8_3")
	php84Runtime := domain.MakeDummyRef("runtime_php_8_4")
	laravelDep := domain.MakeDummyRef("laravel_dep_php83")
	loginComp := domain.MakeDummyRef("target_model_login_comp")
	loginUnit := domain.MakeDummyRef("plan_unit_login")
	loginExecPass := domain.MakeDummyRef("login_verification_pass_result")
	unknownImpact := domain.MakeDummyRef("suspicious_dynamic_eval")

	cov := &domain.ConsumptionCoverageRecord{
		SchemaVersion:        "1.0",
		ProjectID:            "proj1",
		ScopeRef:             domain.MakeDummyRef("scope1"),
		DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
		AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
		InputsExamined:       []string{sourceIR, sourceContract, php83Runtime, laravelDep, loginComp, loginUnit, loginExecPass, unknownImpact},
		RootsExamined:        []string{"root1"},
		Categories: []domain.CoverageCategory{
			{
				Category:     "SOURCE",
				Status:       "CLOSED",
				InputRefs:    []string{sourceIR, sourceContract},
				EvidenceRefs: []string{domain.MakeDummyRef("ev_src")},
			},
			{
				Category:     "RESOLUTION",
				Status:       "CLOSED",
				InputRefs:    []string{php83Runtime, laravelDep},
				EvidenceRefs: []string{domain.MakeDummyRef("ev_res")},
			},
			{
				Category:     "COMPONENT",
				Status:       "CLOSED",
				InputRefs:    []string{loginComp},
				EvidenceRefs: []string{domain.MakeDummyRef("ev_comp")},
			},
		},
		UnresolvedCandidates: []string{},
		ExcludedMechanisms:   []string{},
		ClosureControlRefs:   []string{domain.MakeDummyRef("control1")},
		Status:               "SUFFICIENT",
		Issuer:               "ENGINE",
	}
	covDigest, err := cov.Digest()
	if err != nil {
		t.Fatalf("failed to digest coverage: %v", err)
	}

	graph := &domain.ValidityGraphRecord{
		SchemaVersion: "2.1",
		GraphID:       "graph1",
		ProjectID:     "proj1",
		Edges: []domain.ValidityEdge{
			{
				FromRef:     sourceIR,
				ToRef:       sourceContract,
				Kind:        "SOURCE_CONTRACT",
				Predicate:   "SAME_CONTENT",
				WitnessRefs: []string{sourceIR},
			},
			{
				FromRef:     php83Runtime,
				ToRef:       laravelDep,
				Kind:        "DEPENDENCY_IMPLEMENTATION",
				Predicate:   "SAME_PROFILE",
				WitnessRefs: []string{php83Runtime},
			},
			{
				FromRef:     sourceContract,
				ToRef:       loginComp,
				Kind:        "COMPONENT_UNIT",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{sourceContract},
			},
			{
				FromRef:     laravelDep,
				ToRef:       loginComp,
				Kind:        "COMPONENT_UNIT",
				Predicate:   "SAME_CAPABILITY",
				WitnessRefs: []string{laravelDep},
			},
			{
				FromRef:     loginComp,
				ToRef:       loginUnit,
				Kind:        "SELECTION_PLAN",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{loginComp},
			},
			{
				FromRef:     loginUnit,
				ToRef:       loginExecPass,
				Kind:        "EXECUTION_RESULT",
				Predicate:   "SAME_PROFILE",
				WitnessRefs: []string{loginUnit},
			},
		},
		UnknownImpactRefs:       []string{unknownImpact},
		ConsumptionCoverageRefs: []string{covDigest},
	}

	svc, err := domain.NewSelectiveInvalidationService(graph, []*domain.ConsumptionCoverageRecord{cov})
	if err != nil {
		t.Fatalf("failed to create invalidation service: %v", err)
	}

	// Target runtime changes: PHP 8.3 -> PHP 8.4
	// php83Runtime is changed/stale
	invResult, err := svc.ComputeInvalidation([]string{php83Runtime}, nil)
	if err != nil {
		t.Fatalf("ComputeInvalidation failed: %v", err)
	}

	// Invariant 1: Source discovery and contracts MUST remain VALID (target-independent)
	if !invResult.IsArtifactValid(sourceIR) {
		t.Errorf("expected sourceIR to remain VALID, got: %s (reason: %s)", invResult.ArtifactStates[sourceIR], invResult.GetReason(sourceIR))
	}
	if !invResult.IsArtifactValid(sourceContract) {
		t.Errorf("expected sourceContract to remain VALID, got: %s (reason: %s)", invResult.ArtifactStates[sourceContract], invResult.GetReason(sourceContract))
	}

	// Invariant 2: Dependency findings, target model, plan unit, and verification must become STALE
	if !invResult.IsArtifactStale(laravelDep) {
		t.Errorf("expected laravelDep to be STALE, got: %s", invResult.ArtifactStates[laravelDep])
	}
	if !invResult.IsArtifactStale(loginComp) {
		t.Errorf("expected loginComp to be STALE, got: %s", invResult.ArtifactStates[loginComp])
	}
	if !invResult.IsArtifactStale(loginUnit) {
		t.Errorf("expected loginUnit to be STALE, got: %s", invResult.ArtifactStates[loginUnit])
	}
	if !invResult.IsArtifactStale(loginExecPass) {
		t.Errorf("expected loginExecPass to be STALE, got: %s", invResult.ArtifactStates[loginExecPass])
	}

	// Invariant 3: Unknown edge / unknown impact ref must be conservatively UNKNOWN / STALE
	if !invResult.IsArtifactUnknown(unknownImpact) {
		t.Errorf("expected unknownImpact to be UNKNOWN, got: %s", invResult.ArtifactStates[unknownImpact])
	}

	// Invariant 4: No old PASS relabeling permitted
	// Attempt to reuse loginExecPass under the new target
	reqReusePass := domain.ReuseCertificateRequest{
		ProjectID:               "proj1",
		OldPlanHash:             strings.Repeat("a", 64),
		NewPlanHash:             strings.Repeat("b", 64),
		ArtifactRef:             loginExecPass,
		Kind:                    "IMPLEMENTATION",
		CheckedEdgeRefs:         []string{loginExecPass},
		EvidenceRefs:            []string{domain.MakeDummyRef("ev1")},
		ConsumptionCoverageRefs: []string{covDigest},
		IsVerificationResult:    true, // Attempting to reuse verification PASS
	}
	canReuse, err := svc.CanReuseArtifact(invResult, reqReusePass)
	if canReuse || err == nil {
		t.Fatalf("expected reuse of old verification PASS to be rejected (Test X29), got canReuse=%v, err=%v", canReuse, err)
	}
	if !strings.Contains(err.Error(), "verification results cannot be relabeled or reused") || !strings.Contains(err.Error(), "Test X29") {
		t.Errorf("expected Test X29 error message, got: %v", err)
	}

	// Attempting to issue a reuse certificate for the old PASS result must fail
	_, err = svc.IssueReuseCertificate(invResult, reqReusePass)
	if err == nil {
		t.Fatal("expected IssueReuseCertificate for old PASS to fail (Test X29), got nil")
	}

	// Source characterization CAN be reused with valid coverage
	reqReuseSource := domain.ReuseCertificateRequest{
		ProjectID:               "proj1",
		OldPlanHash:             strings.Repeat("a", 64),
		NewPlanHash:             strings.Repeat("b", 64),
		ArtifactRef:             sourceContract,
		Kind:                    "CHARACTERIZATION",
		CheckedEdgeRefs:         []string{sourceContract},
		EvidenceRefs:            []string{sourceIR},
		ConsumptionCoverageRefs: []string{covDigest},
		IsVerificationResult:    false,
	}
	cert, err := svc.IssueReuseCertificate(invResult, reqReuseSource)
	if err != nil {
		t.Fatalf("expected valid characterization reuse certificate for sourceContract, got: %v", err)
	}
	if cert == nil || !cert.Valid {
		t.Fatal("expected valid certificate record, got nil or invalid")
	}
	if cert.SchemaVersion != "2.1" {
		t.Errorf("expected schema_version 2.1, got: %s", cert.SchemaVersion)
	}
	_ = php84Runtime
}

// Test X48: Selective Source Change
// "Change one controller and an unrelated source fact; only actual reverse dependency closure becomes stale,
//  unrelated finding stays valid; missing edge/unknown impact invalidates conservatively"
func TestSelectiveInvalidation_X48_SelectiveSourceChange(t *testing.T) {
	// Source facts:
	// - login_controller (changed)
	// - catalog_controller (unrelated source fact)
	loginCtrl := domain.MakeDummyRef("LoginController.php")
	catalogCtrl := domain.MakeDummyRef("CatalogController.php")

	loginFinding := domain.MakeDummyRef("finding_login")
	catalogFinding := domain.MakeDummyRef("finding_catalog")

	loginComp := domain.MakeDummyRef("comp_login")
	catalogComp := domain.MakeDummyRef("comp_catalog")

	loginUnit := domain.MakeDummyRef("unit_login")
	catalogUnit := domain.MakeDummyRef("unit_catalog")

	unknownImpact := domain.MakeDummyRef("unexamined_eval")

	cov := &domain.ConsumptionCoverageRecord{
		SchemaVersion:        "1.0",
		ProjectID:            "proj1",
		ScopeRef:             domain.MakeDummyRef("scope1"),
		DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
		AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
		InputsExamined:       []string{loginCtrl, catalogCtrl, loginFinding, catalogFinding, loginComp, catalogComp, loginUnit, catalogUnit, unknownImpact},
		RootsExamined:        []string{"root1"},
		Categories: []domain.CoverageCategory{
			{
				Category:     "SOURCE",
				Status:       "CLOSED",
				InputRefs:    []string{loginCtrl, catalogCtrl},
				EvidenceRefs: []string{domain.MakeDummyRef("ev1")},
			},
		},
		UnresolvedCandidates: []string{},
		ExcludedMechanisms:   []string{},
		ClosureControlRefs:   []string{domain.MakeDummyRef("ctrl1")},
		Status:               "SUFFICIENT",
		Issuer:               "ENGINE",
	}
	covDigest, err := cov.Digest()
	if err != nil {
		t.Fatalf("failed to digest coverage: %v", err)
	}

	graph := &domain.ValidityGraphRecord{
		SchemaVersion: "2.1",
		GraphID:       "graph_x48",
		ProjectID:     "proj1",
		Edges: []domain.ValidityEdge{
			// Login chain
			{
				FromRef:     loginCtrl,
				ToRef:       loginFinding,
				Kind:        "INPUT_ASSESSMENT_FINDING",
				Predicate:   "SAME_CONTENT",
				WitnessRefs: []string{loginCtrl},
			},
			{
				FromRef:     loginFinding,
				ToRef:       loginComp,
				Kind:        "COMPONENT_UNIT",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{loginFinding},
			},
			{
				FromRef:     loginComp,
				ToRef:       loginUnit,
				Kind:        "SELECTION_PLAN",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{loginComp},
			},
			// Catalog chain (independent)
			{
				FromRef:     catalogCtrl,
				ToRef:       catalogFinding,
				Kind:        "INPUT_ASSESSMENT_FINDING",
				Predicate:   "SAME_CONTENT",
				WitnessRefs: []string{catalogCtrl},
			},
			{
				FromRef:     catalogFinding,
				ToRef:       catalogComp,
				Kind:        "COMPONENT_UNIT",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{catalogFinding},
			},
			{
				FromRef:     catalogComp,
				ToRef:       catalogUnit,
				Kind:        "SELECTION_PLAN",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{catalogComp},
			},
		},
		UnknownImpactRefs:       []string{unknownImpact},
		ConsumptionCoverageRefs: []string{covDigest},
	}

	svc, err := domain.NewSelectiveInvalidationService(graph, []*domain.ConsumptionCoverageRecord{cov})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Mutate ONLY loginCtrl
	invResult, err := svc.ComputeInvalidation([]string{loginCtrl}, nil)
	if err != nil {
		t.Fatalf("ComputeInvalidation failed: %v", err)
	}

	// Invariant 1: Only actual reverse dependency closure of loginCtrl becomes STALE
	if !invResult.IsArtifactStale(loginCtrl) {
		t.Errorf("expected loginCtrl to be STALE, got: %s", invResult.ArtifactStates[loginCtrl])
	}
	if !invResult.IsArtifactStale(loginFinding) {
		t.Errorf("expected loginFinding to be STALE, got: %s", invResult.ArtifactStates[loginFinding])
	}
	if !invResult.IsArtifactStale(loginComp) {
		t.Errorf("expected loginComp to be STALE, got: %s", invResult.ArtifactStates[loginComp])
	}
	if !invResult.IsArtifactStale(loginUnit) {
		t.Errorf("expected loginUnit to be STALE, got: %s", invResult.ArtifactStates[loginUnit])
	}

	// Invariant 2: Unrelated findings and components stay VALID
	if !invResult.IsArtifactValid(catalogCtrl) {
		t.Errorf("expected catalogCtrl to stay VALID, got: %s (reason: %s)", invResult.ArtifactStates[catalogCtrl], invResult.GetReason(catalogCtrl))
	}
	if !invResult.IsArtifactValid(catalogFinding) {
		t.Errorf("expected catalogFinding to stay VALID, got: %s (reason: %s)", invResult.ArtifactStates[catalogFinding], invResult.GetReason(catalogFinding))
	}
	if !invResult.IsArtifactValid(catalogComp) {
		t.Errorf("expected catalogComp to stay VALID, got: %s (reason: %s)", invResult.ArtifactStates[catalogComp], invResult.GetReason(catalogComp))
	}
	if !invResult.IsArtifactValid(catalogUnit) {
		t.Errorf("expected catalogUnit to stay VALID, got: %s (reason: %s)", invResult.ArtifactStates[catalogUnit], invResult.GetReason(catalogUnit))
	}

	// Invariant 3: Unknown impact ref invalidates conservatively
	if !invResult.IsArtifactUnknown(unknownImpact) {
		t.Errorf("expected unknownImpact to be UNKNOWN, got: %s", invResult.ArtifactStates[unknownImpact])
	}
}

// Test X49: Target Change with History
// "PHP intent 8.3→8.4 makes affected dependency/target findings and downstream plan stale
//  while source-only facts and historical report bytes persist; stale approval cannot execute"
func TestSelectiveInvalidation_X49_TargetChangeWithHistoryAndStaleApproval(t *testing.T) {
	sourceFact := domain.MakeDummyRef("source_fact_auth")
	historicalReport := domain.MakeDummyRef("assessment_report_v1")
	targetIntentPHP83 := domain.MakeDummyRef("target_intent_php83")
	depFinding := domain.MakeDummyRef("dependency_finding_php83")
	targetModelComp := domain.MakeDummyRef("tim_component_login")
	planUnit := domain.MakeDummyRef("plan_unit_login")

	cov := &domain.ConsumptionCoverageRecord{
		SchemaVersion:        "1.0",
		ProjectID:            "proj1",
		ScopeRef:             domain.MakeDummyRef("scope1"),
		DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
		AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
		InputsExamined:       []string{sourceFact, historicalReport, targetIntentPHP83, depFinding, targetModelComp, planUnit},
		RootsExamined:        []string{"root1"},
		Categories: []domain.CoverageCategory{
			{
				Category:     "SOURCE",
				Status:       "CLOSED",
				InputRefs:    []string{sourceFact},
				EvidenceRefs: []string{domain.MakeDummyRef("ev_src")},
			},
		},
		UnresolvedCandidates: []string{},
		ExcludedMechanisms:   []string{},
		ClosureControlRefs:   []string{domain.MakeDummyRef("ctrl1")},
		Status:               "SUFFICIENT",
		Issuer:               "ENGINE",
	}
	covDigest, err := cov.Digest()
	if err != nil {
		t.Fatalf("failed to digest coverage: %v", err)
	}

	graph := &domain.ValidityGraphRecord{
		SchemaVersion: "2.1",
		GraphID:       "graph_x49",
		ProjectID:     "proj1",
		Edges: []domain.ValidityEdge{
			// Historical report indexes source fact
			{
				FromRef:     sourceFact,
				ToRef:       historicalReport,
				Kind:        "FINDING_ASSESSMENT_REPORT",
				Predicate:   "SAME_CONTENT",
				WitnessRefs: []string{sourceFact},
			},
			// Target intent connects to dependency finding
			{
				FromRef:     targetIntentPHP83,
				ToRef:       depFinding,
				Kind:        "INPUT_ASSESSMENT_FINDING",
				Predicate:   "SAME_PROFILE",
				WitnessRefs: []string{targetIntentPHP83},
			},
			// Target model component connects to dependency finding
			{
				FromRef:     depFinding,
				ToRef:       targetModelComp,
				Kind:        "COMPONENT_UNIT",
				Predicate:   "SAME_CAPABILITY",
				WitnessRefs: []string{depFinding},
			},
			// Downstream plan unit connects to target model component
			{
				FromRef:     targetModelComp,
				ToRef:       planUnit,
				Kind:        "SELECTION_PLAN",
				Predicate:   "SAME_RULES",
				WitnessRefs: []string{targetModelComp},
			},
		},
		UnknownImpactRefs:       []string{},
		ConsumptionCoverageRefs: []string{covDigest},
	}

	svc, err := domain.NewSelectiveInvalidationService(graph, []*domain.ConsumptionCoverageRecord{cov})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Change target intent: PHP 8.3 -> 8.4
	invResult, err := svc.ComputeInvalidation([]string{targetIntentPHP83}, nil)
	if err != nil {
		t.Fatalf("ComputeInvalidation failed: %v", err)
	}

	// Invariant 1: Source-only facts and historical report remain VALID
	if !invResult.IsArtifactValid(sourceFact) {
		t.Errorf("expected sourceFact to remain VALID, got: %s", invResult.ArtifactStates[sourceFact])
	}
	if !invResult.IsArtifactValid(historicalReport) {
		t.Errorf("expected historicalReport to remain VALID, got: %s", invResult.ArtifactStates[historicalReport])
	}

	// Invariant 2: Affected dependency finding, target model component, and downstream plan unit are STALE
	if !invResult.IsArtifactStale(depFinding) {
		t.Errorf("expected depFinding to be STALE, got: %s", invResult.ArtifactStates[depFinding])
	}
	if !invResult.IsArtifactStale(targetModelComp) {
		t.Errorf("expected targetModelComp to be STALE, got: %s", invResult.ArtifactStates[targetModelComp])
	}
	if !invResult.IsArtifactStale(planUnit) {
		t.Errorf("expected planUnit to be STALE, got: %s", invResult.ArtifactStates[planUnit])
	}

	// Invariant 3: Stale Gate 2 approval CANNOT execute
	oldPlanHash := strings.Repeat("1", 64)
	newPlanHash := strings.Repeat("2", 64)

	// Case A: Plan hash mismatch (receipt was for old plan)
	err = domain.ValidatePlanExecutionReadiness(
		newPlanHash,
		[]string{planUnit},
		invResult,
		oldPlanHash, // receipt has old plan hash
		[]string{depFinding},
	)
	if err == nil {
		t.Fatal("expected ValidatePlanExecutionReadiness to fail on plan hash mismatch (Test X49), got nil")
	}
	if !strings.Contains(err.Error(), "stale approval cannot execute") || !strings.Contains(err.Error(), "Test X49") {
		t.Errorf("expected Test X49 stale approval error, got: %v", err)
	}

	// Case B: Same plan hash but constituent artifact is stale
	err = domain.ValidatePlanExecutionReadiness(
		oldPlanHash,
		[]string{planUnit}, // planUnit is STALE
		invResult,
		oldPlanHash, // matches receipt
		[]string{},
	)
	if err == nil {
		t.Fatal("expected ValidatePlanExecutionReadiness to fail on stale plan artifact (Test X49), got nil")
	}
	if !strings.Contains(err.Error(), "stale plan cannot execute") || !strings.Contains(err.Error(), "Test X49") {
		t.Errorf("expected Test X49 stale plan error, got: %v", err)
	}

	// Case C: Stale decision ref in approval receipt
	err = domain.ValidatePlanExecutionReadiness(
		oldPlanHash,
		[]string{sourceFact}, // sourceFact is valid
		invResult,
		oldPlanHash,
		[]string{depFinding}, // depFinding is STALE
	)
	if err == nil {
		t.Fatal("expected ValidatePlanExecutionReadiness to fail on stale decision ref (Test X49), got nil")
	}
	if !strings.Contains(err.Error(), "stale approval cannot execute") || !strings.Contains(err.Error(), "Gate 2 re-confirmation required") {
		t.Errorf("expected Gate 2 re-confirmation required error, got: %v", err)
	}
}

// Test X73: Missing Consumption Edge & Positive Consumption Coverage
// "Missing consumption edge: Known source/shared-component input consumed by completed unit;
//  Remove edge or entire graph and change consumed input: Reuse refused; descendants stale until coverage closes.
//  No edge is not proof of independence. Examined inputs/roots, consumption manifest, coverage controls and successor graph."
func TestSelectiveInvalidation_X73_MissingConsumptionEdgeRefusesReuse(t *testing.T) {
	sharedInput := domain.MakeDummyRef("shared_auth_helper")
	completedUnit := domain.MakeDummyRef("unit_login_controller")
	descendantArtifact := domain.MakeDummyRef("downstream_session_bridge")

	// Subtest 1: Missing consumption edge in graph refuses reuse and marks artifact STALE
	t.Run("missing_consumption_edge_marks_stale_and_refuses_reuse", func(t *testing.T) {
		cov := &domain.ConsumptionCoverageRecord{
			SchemaVersion:        "1.0",
			ProjectID:            "proj1",
			ScopeRef:             domain.MakeDummyRef("scope1"),
			DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
			AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
			InputsExamined:       []string{sharedInput, completedUnit, descendantArtifact},
			RootsExamined:        []string{"root1"},
			Categories: []domain.CoverageCategory{
				{
					Category:     "COMPONENT",
					Status:       "CLOSED",
					InputRefs:    []string{sharedInput},
					EvidenceRefs: []string{domain.MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{domain.MakeDummyRef("ctrl1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, err := cov.Digest()
		if err != nil {
			t.Fatalf("failed to digest coverage: %v", err)
		}

		// Graph has an edge from completedUnit -> descendantArtifact,
		// BUT MISSES the edge from sharedInput -> completedUnit!
		graphMissingEdge := &domain.ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x73_missing",
			ProjectID:     "proj1",
			Edges: []domain.ValidityEdge{
				{
					FromRef:     completedUnit,
					ToRef:       descendantArtifact,
					Kind:        "COMPONENT_UNIT",
					Predicate:   "SAME_RULES",
					WitnessRefs: []string{completedUnit},
				},
			},
			UnknownImpactRefs:       []string{},
			ConsumptionCoverageRefs: []string{covDigest},
		}

		svc, err := domain.NewSelectiveInvalidationService(graphMissingEdge, []*domain.ConsumptionCoverageRecord{cov})
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		// Declared consumption: completedUnit consumes sharedInput
		consumed := map[string][]string{
			completedUnit: {sharedInput},
		}

		invResult, err := svc.ComputeInvalidation([]string{}, consumed)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		// Invariant: completedUnit MUST be marked STALE due to missing consumption edge
		if !invResult.IsArtifactStale(completedUnit) {
			t.Fatalf("expected completedUnit to be STALE due to missing consumption edge (Test X73), got: %s", invResult.ArtifactStates[completedUnit])
		}
		if !strings.Contains(invResult.GetReason(completedUnit), "missing consumption edge") || !strings.Contains(invResult.GetReason(completedUnit), "Test X73") {
			t.Errorf("expected Test X73 missing consumption edge reason, got: %s", invResult.GetReason(completedUnit))
		}

		// Invariant: descendantArtifact MUST also become STALE
		if !invResult.IsArtifactStale(descendantArtifact) {
			t.Fatalf("expected descendantArtifact to be STALE, got: %s", invResult.ArtifactStates[descendantArtifact])
		}

		// Invariant: Reuse certificate MUST be refused
		reqReuse := domain.ReuseCertificateRequest{
			ProjectID:               "proj1",
			OldPlanHash:             strings.Repeat("a", 64),
			NewPlanHash:             strings.Repeat("b", 64),
			ArtifactRef:             completedUnit,
			Kind:                    "IMPLEMENTATION",
			CheckedEdgeRefs:         []string{completedUnit},
			EvidenceRefs:            []string{sharedInput},
			ConsumptionCoverageRefs: []string{covDigest},
		}
		_, err = svc.IssueReuseCertificate(invResult, reqReuse)
		if err == nil {
			t.Fatal("expected reuse certificate to be refused for unit with missing consumption edge (Test X73), got nil")
		}
		if !strings.Contains(err.Error(), "cannot reuse stale artifact") {
			t.Errorf("expected cannot reuse stale artifact error, got: %v", err)
		}
	})

	// Subtest 2: Unexamined input in consumption coverage refuses reuse
	t.Run("unregistered_consumption_input_refuses_reuse", func(t *testing.T) {
		covWithoutSharedInput := &domain.ConsumptionCoverageRecord{
			SchemaVersion:        "1.0",
			ProjectID:            "proj1",
			ScopeRef:             domain.MakeDummyRef("scope1"),
			DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
			AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
			InputsExamined:       []string{completedUnit}, // sharedInput NOT in inputs_examined!
			RootsExamined:        []string{"root1"},
			Categories: []domain.CoverageCategory{
				{
					Category:     "COMPONENT",
					Status:       "CLOSED",
					InputRefs:    []string{completedUnit},
					EvidenceRefs: []string{domain.MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{domain.MakeDummyRef("ctrl1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, err := covWithoutSharedInput.Digest()
		if err != nil {
			t.Fatalf("failed to digest coverage: %v", err)
		}

		graph := &domain.ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x73_unregistered",
			ProjectID:     "proj1",
			Edges: []domain.ValidityEdge{
				{
					FromRef:     sharedInput,
					ToRef:       completedUnit,
					Kind:        "COMPONENT_UNIT",
					Predicate:   "SAME_RULES",
					WitnessRefs: []string{sharedInput},
				},
			},
			UnknownImpactRefs:       []string{},
			ConsumptionCoverageRefs: []string{covDigest},
		}

		svc, err := domain.NewSelectiveInvalidationService(graph, []*domain.ConsumptionCoverageRecord{covWithoutSharedInput})
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		consumed := map[string][]string{
			completedUnit: {sharedInput},
		}
		invResult, err := svc.ComputeInvalidation([]string{}, consumed)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		if !invResult.IsArtifactStale(completedUnit) {
			t.Fatalf("expected completedUnit to be STALE due to unregistered consumption, got: %s", invResult.ArtifactStates[completedUnit])
		}
		if !strings.Contains(invResult.GetReason(completedUnit), "unregistered consumption") || !strings.Contains(invResult.GetReason(completedUnit), "Test X73") {
			t.Errorf("expected Test X73 unregistered consumption error, got: %s", invResult.GetReason(completedUnit))
		}
	})

	// Subtest 3: INSUFFICIENT consumption coverage refuses reuse
	t.Run("insufficient_coverage_refuses_reuse", func(t *testing.T) {
		insufficientCov := &domain.ConsumptionCoverageRecord{
			SchemaVersion:        "1.0",
			ProjectID:            "proj1",
			ScopeRef:             domain.MakeDummyRef("scope1"),
			DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
			AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
			InputsExamined:       []string{sharedInput, completedUnit},
			RootsExamined:        []string{"root1"},
			Categories: []domain.CoverageCategory{
				{
					Category:     "COMPONENT",
					Status:       "UNKNOWN", // UNKNOWN category status makes coverage INSUFFICIENT
					InputRefs:    []string{sharedInput},
					EvidenceRefs: []string{domain.MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{"dynamic dispatch unresolved"},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{},
			Status:               "INSUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, err := insufficientCov.Digest()
		if err != nil {
			t.Fatalf("failed to digest coverage: %v", err)
		}

		graph := &domain.ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x73_insufficient",
			ProjectID:     "proj1",
			Edges: []domain.ValidityEdge{
				{
					FromRef:     sharedInput,
					ToRef:       completedUnit,
					Kind:        "COMPONENT_UNIT",
					Predicate:   "SAME_RULES",
					WitnessRefs: []string{sharedInput},
				},
			},
			UnknownImpactRefs:       []string{},
			ConsumptionCoverageRefs: []string{covDigest},
		}

		svc, err := domain.NewSelectiveInvalidationService(graph, []*domain.ConsumptionCoverageRecord{insufficientCov})
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		invResult, err := svc.ComputeInvalidation([]string{}, nil)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		// Descendants of the unclosed input must be marked STALE
		if !invResult.IsArtifactStale(completedUnit) {
			t.Fatalf("expected completedUnit to be STALE with insufficient coverage, got: %s", invResult.ArtifactStates[completedUnit])
		}

		// Attempting reuse certificate fails
		req := domain.ReuseCertificateRequest{
			ProjectID:               "proj1",
			OldPlanHash:             strings.Repeat("a", 64),
			NewPlanHash:             strings.Repeat("b", 64),
			ArtifactRef:             completedUnit,
			Kind:                    "IMPLEMENTATION",
			CheckedEdgeRefs:         []string{completedUnit},
			EvidenceRefs:            []string{sharedInput},
			ConsumptionCoverageRefs: []string{covDigest},
		}
		_, err = svc.IssueReuseCertificate(invResult, req)
		if err == nil {
			t.Fatal("expected reuse certificate to be refused for insufficient coverage, got nil")
		}
	})

	// Subtest 4: Closing consumption coverage and adding consumption edge permits reuse
	t.Run("closed_coverage_and_edge_permits_reuse", func(t *testing.T) {
		sufficientCov := &domain.ConsumptionCoverageRecord{
			SchemaVersion:        "1.0",
			ProjectID:            "proj1",
			ScopeRef:             domain.MakeDummyRef("scope1"),
			DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
			AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
			InputsExamined:       []string{sharedInput, completedUnit, descendantArtifact},
			RootsExamined:        []string{"root1"},
			Categories: []domain.CoverageCategory{
				{
					Category:     "COMPONENT",
					Status:       "CLOSED",
					InputRefs:    []string{sharedInput, completedUnit},
					EvidenceRefs: []string{domain.MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{domain.MakeDummyRef("ctrl1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, err := sufficientCov.Digest()
		if err != nil {
			t.Fatalf("failed to digest coverage: %v", err)
		}

		graph := &domain.ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x73_closed",
			ProjectID:     "proj1",
			Edges: []domain.ValidityEdge{
				{
					FromRef:     sharedInput,
					ToRef:       completedUnit,
					Kind:        "COMPONENT_UNIT",
					Predicate:   "SAME_RULES",
					WitnessRefs: []string{sharedInput},
				},
				{
					FromRef:     completedUnit,
					ToRef:       descendantArtifact,
					Kind:        "COMPONENT_UNIT",
					Predicate:   "SAME_RULES",
					WitnessRefs: []string{completedUnit},
				},
			},
			UnknownImpactRefs:       []string{},
			ConsumptionCoverageRefs: []string{covDigest},
		}

		svc, err := domain.NewSelectiveInvalidationService(graph, []*domain.ConsumptionCoverageRecord{sufficientCov})
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		consumed := map[string][]string{
			completedUnit: {sharedInput},
		}

		invResult, err := svc.ComputeInvalidation([]string{}, consumed)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		if !invResult.IsArtifactValid(completedUnit) {
			t.Fatalf("expected completedUnit to be VALID with closed coverage, got: %s", invResult.ArtifactStates[completedUnit])
		}
		if !invResult.IsArtifactValid(descendantArtifact) {
			t.Fatalf("expected descendantArtifact to be VALID with closed coverage, got: %s", invResult.ArtifactStates[descendantArtifact])
		}

		req := domain.ReuseCertificateRequest{
			ProjectID:               "proj1",
			OldPlanHash:             strings.Repeat("a", 64),
			NewPlanHash:             strings.Repeat("b", 64),
			ArtifactRef:             completedUnit,
			Kind:                    "IMPLEMENTATION",
			CheckedEdgeRefs:         []string{completedUnit},
			EvidenceRefs:            []string{sharedInput},
			ConsumptionCoverageRefs: []string{covDigest},
		}
		cert, err := svc.IssueReuseCertificate(invResult, req)
		if err != nil {
			t.Fatalf("expected reuse certificate issuance to succeed (Test X73), got: %v", err)
		}
		if cert == nil || !cert.Valid {
			t.Fatal("expected valid certificate, got nil or invalid")
		}
	})
}

func TestValidityGraphRecord_Methods(t *testing.T) {
	record := &domain.ValidityGraphRecord{
		SchemaVersion: "2.1",
		GraphID:       "graph1",
		ProjectID:     "proj1",
		Edges: []domain.ValidityEdge{
			{
				FromRef:     domain.MakeDummyRef("from1"),
				ToRef:       domain.MakeDummyRef("to1"),
				Kind:        "SOURCE_IR",
				Predicate:   "SAME_CONTENT",
				WitnessRefs: []string{domain.MakeDummyRef("w1")},
			},
		},
		UnknownImpactRefs:       []string{},
		ConsumptionCoverageRefs: []string{domain.MakeDummyRef("cov1")},
	}

	canon, err := record.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize validity graph: %v", err)
	}

	parsed, err := domain.ParseValidityGraph(canon)
	if err != nil {
		t.Fatalf("failed to parse validity graph: %v", err)
	}
	if parsed.GraphID != "graph1" || len(parsed.Edges) != 1 {
		t.Errorf("unexpected parsed record: %+v", parsed)
	}

	digest, err := record.Digest()
	if err != nil {
		t.Fatalf("failed to compute digest: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Errorf("unexpected digest format: %s", digest)
	}

	// Nil validity graph record
	var nilRecord *domain.ValidityGraphRecord
	if _, err := nilRecord.CanonicalJSON(); err == nil {
		t.Fatal("expected error on nil record, got nil")
	}

	// Schema 2.0 validity graph
	record20 := &domain.ValidityGraphRecord{
		SchemaVersion: "2.0",
		GraphID:       "graph20",
		ProjectID:     "proj1",
		Edges: []domain.ValidityEdge{
			{
				FromRef:     domain.MakeDummyRef("from1"),
				ToRef:       domain.MakeDummyRef("to1"),
				Kind:        "SOURCE_IR",
				Predicate:   "SAME_CONTENT",
				WitnessRefs: []string{domain.MakeDummyRef("w1")},
			},
		},
		UnknownImpactRefs: []string{},
	}
	canon20, err := record20.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize 2.0 graph: %v", err)
	}
	if strings.Contains(string(canon20), "consumption_coverage_refs") {
		t.Errorf("expected 2.0 graph to NOT contain consumption_coverage_refs, got: %s", string(canon20))
	}
}

func TestConsumptionCoverageRecord_Methods(t *testing.T) {
	inRef := domain.MakeDummyRef("in1")
	cov := &domain.ConsumptionCoverageRecord{
		SchemaVersion:        "1.0",
		ProjectID:            "proj1",
		ScopeRef:             domain.MakeDummyRef("scope1"),
		DiscoveryMethodRef:   domain.MakeDummyRef("disc1"),
		AnalysisRevisionRef:  domain.MakeDummyRef("rev1"),
		InputsExamined:       []string{inRef},
		RootsExamined:        []string{"root1"},
		Categories: []domain.CoverageCategory{
			{
				Category:     "SOURCE",
				Status:       "CLOSED",
				InputRefs:    []string{inRef},
				EvidenceRefs: []string{domain.MakeDummyRef("ev1")},
			},
		},
		UnresolvedCandidates: []string{},
		ExcludedMechanisms:   []string{},
		ClosureControlRefs:   []string{domain.MakeDummyRef("ctrl1")},
		Status:               "SUFFICIENT",
		Issuer:               "ENGINE",
	}

	if !cov.IsSufficient() {
		t.Error("expected IsSufficient to return true")
	}

	canon, err := cov.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize consumption coverage: %v", err)
	}

	parsed, err := domain.ParseConsumptionCoverage(canon)
	if err != nil {
		t.Fatalf("failed to parse consumption coverage: %v", err)
	}
	if parsed.ProjectID != "proj1" || len(parsed.Categories) != 1 {
		t.Errorf("unexpected parsed record: %+v", parsed)
	}

	digest, err := cov.Digest()
	if err != nil {
		t.Fatalf("failed to compute digest: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Errorf("unexpected digest format: %s", digest)
	}

	// Nil coverage
	var nilCov *domain.ConsumptionCoverageRecord
	if nilCov.IsSufficient() {
		t.Error("expected nil coverage IsSufficient to be false")
	}
	if _, err := nilCov.CanonicalJSON(); err == nil {
		t.Fatal("expected error on nil CanonicalJSON, got nil")
	}
}

func TestReuseCertificateRecord_Methods(t *testing.T) {
	cert := &domain.ReuseCertificateRecord{
		SchemaVersion:           "2.1",
		ProjectID:               "proj1",
		OldPlanHash:             strings.Repeat("a", 64),
		NewPlanHash:             strings.Repeat("b", 64),
		ArtifactRef:             domain.MakeDummyRef("art1"),
		GraphRef:                domain.MakeDummyRef("graph1"),
		Kind:                    "IMPLEMENTATION",
		CheckedEdgeRefs:         []string{domain.MakeDummyRef("edge1")},
		EvidenceRefs:            []string{domain.MakeDummyRef("ev1")},
		Issuer:                  "ENGINE",
		Valid:                   true,
		ConsumptionCoverageRefs: []string{domain.MakeDummyRef("cov1")},
	}

	canon, err := cert.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize reuse certificate: %v", err)
	}

	parsed, err := domain.ParseReuseCertificate(canon)
	if err != nil {
		t.Fatalf("failed to parse reuse certificate: %v", err)
	}
	if parsed.ProjectID != "proj1" || !parsed.Valid {
		t.Errorf("unexpected parsed certificate: %+v", parsed)
	}

	digest, err := cert.Digest()
	if err != nil {
		t.Fatalf("failed to compute digest: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Errorf("unexpected digest format: %s", digest)
	}

	// Nil reuse certificate
	var nilCert *domain.ReuseCertificateRecord
	if _, err := nilCert.CanonicalJSON(); err == nil {
		t.Fatal("expected error on nil CanonicalJSON, got nil")
	}

	// Schema 2.0 reuse certificate
	cert20 := &domain.ReuseCertificateRecord{
		SchemaVersion:   "2.0",
		ProjectID:       "proj1",
		OldPlanHash:     strings.Repeat("a", 64),
		NewPlanHash:     strings.Repeat("b", 64),
		ArtifactRef:     domain.MakeDummyRef("art1"),
		GraphRef:        domain.MakeDummyRef("graph1"),
		Kind:            "CHARACTERIZATION",
		CheckedEdgeRefs: []string{domain.MakeDummyRef("edge1")},
		EvidenceRefs:    []string{domain.MakeDummyRef("ev1")},
		Issuer:          "ENGINE",
		Valid:           true,
	}
	canon20, err := cert20.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize 2.0 cert: %v", err)
	}
	if strings.Contains(string(canon20), "consumption_coverage_refs") {
		t.Errorf("expected 2.0 cert to NOT contain consumption_coverage_refs, got: %s", string(canon20))
	}
}

func TestSelectiveInvalidation_EdgeCases(t *testing.T) {
	// Nil graph in constructor
	_, err := domain.NewSelectiveInvalidationService(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil graph, got nil")
	}

	// Undeclared coverage in schema 2.1
	graph21 := &domain.ValidityGraphRecord{
		SchemaVersion:           "2.1",
		GraphID:                 "graph1",
		ProjectID:               "proj1",
		Edges:                   []domain.ValidityEdge{},
		UnknownImpactRefs:       []string{},
		ConsumptionCoverageRefs: []string{domain.MakeDummyRef("missing_cov")},
	}
	_, err = domain.NewSelectiveInvalidationService(graph21, nil)
	if err == nil {
		t.Fatal("expected error for missing declared coverage, got nil")
	}

	// InvalidationError methods
	underlyingErr := errors.New("underlying failure")
	baseErr := &domain.InvalidationError{
		Reason: "test reason",
		Cause:  underlyingErr,
	}
	if !strings.Contains(baseErr.Error(), "test reason") {
		t.Errorf("unexpected Error() string: %s", baseErr.Error())
	}
	if baseErr.Unwrap() != underlyingErr {
		t.Errorf("expected unwrapped underlyingErr, got: %v", baseErr.Unwrap())
	}

	// Nil receiver checks on SelectiveInvalidationResult
	var nilResult *domain.SelectiveInvalidationResult
	if nilResult.IsArtifactValid("foo") {
		t.Error("expected false on nil result")
	}
	if nilResult.IsArtifactStale("foo") {
		t.Error("expected false on nil result")
	}
	if nilResult.IsArtifactUnknown("foo") {
		t.Error("expected false on nil result")
	}
	if nilResult.GetReason("foo") != "" {
		t.Error("expected empty string on nil result")
	}
	if len(nilResult.StaleRefs()) != 0 {
		t.Error("expected empty slice on nil result")
	}
	if len(nilResult.ValidRefs()) != 0 {
		t.Error("expected empty slice on nil result")
	}
}

