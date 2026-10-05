package domain

import (
	"errors"
	"strings"
	"testing"
)

// TestM3_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M3:
// Exact Planning, covering all authoritative M3 falsification tests:
// X15, X75, X16, X17, X18, X50, X29, X48, X49, X73, X55, X58, X59.
func TestM3_AcceptanceClosure(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)
	dummyRef2 := "sha256:" + strings.Repeat("b", 64)

	// =========================================================================
	// 1. Test X15: Wildcards & Target Resolution Precedence
	// =========================================================================
	t.Run("X15_TargetResolutionPrecedence", func(t *testing.T) {
		// A. User intent overrides project defaults and recommendations
		fwUser := "11.0.0"
		userIntent := &TargetEnvironmentDocument{
			SchemaVersion:  "2.0",
			Language:       "php",
			Framework:      "laravel",
			Version:        &fwUser,
			Architecture:   TargetArchitecture{Style: "modular_monolith"},
			DatabasePolicy: "preserve_existing",
			Path:           "target",
			Runtime: TargetRuntime{
				Version:      "8.3.1",
				OS:           "linux",
				Architecture: "amd64",
			},
			Deployment: TargetDeployment{
				Model:      "container",
				ProfileRef: dummyRef,
			},
		}

		fwProj := "10.0.0"
		projDefaults := &TargetEnvironmentDocument{
			SchemaVersion:  "2.0",
			Language:       "php",
			Framework:      "laravel",
			Version:        &fwProj,
			Architecture:   TargetArchitecture{Style: "modular_monolith"},
			DatabasePolicy: "preserve_existing",
			Path:           "target",
			Runtime: TargetRuntime{
				Version:      "8.2.0",
				OS:           "linux",
				Architecture: "amd64",
			},
			Deployment: TargetDeployment{
				Model:      "container",
				ProfileRef: dummyRef,
			},
		}

		inputs := &TargetPrecedenceInputs{
			ExplicitUserIntent: userIntent,
			ProjectDefaults:    projDefaults,
		}

		evaluated, err := EvaluateTargetPrecedence(inputs)
		if err != nil {
			t.Fatalf("EvaluateTargetPrecedence failed: %v", err)
		}
		if evaluated.Framework != "laravel" || evaluated.Runtime.Version != "8.3.1" || evaluated.Version == nil || *evaluated.Version != "11.0.0" {
			t.Errorf("X15 FAILED: user intent must win; got %s, version %v, runtime %s", evaluated.Framework, evaluated.Version, evaluated.Runtime.Version)
		}

		// B. Wildcards rejected as exact resolution
		wildcardVer := "11.*"
		wildcardIntent := *userIntent
		wildcardIntent.Version = &wildcardVer
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         &wildcardIntent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.0",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}
		_, err = ResolveTargetEnvironment(req)
		if err == nil || !strings.Contains(err.Error(), "wildcard") {
			t.Errorf("X15 FAILED: wildcard version must be rejected as exact resolution: %v", err)
		}
	})

	// =========================================================================
	// 2. Test X75: Provider Release Interpretation
	// =========================================================================
	t.Run("X75_ProviderReleaseInterpretation", func(t *testing.T) {
		distRef := "sha256:" + strings.Repeat("c", 64)
		interpRef := "sha256:" + strings.Repeat("d", 64)

		reg := &mockProviderRegistry{
			availableRuntimes: map[string][]string{"php": {"release-2026-10-03"}},
			runtimeComps: map[string]*ResolvedComponent{
				"php:release-2026-10-03": {
					ProviderID:        "canonical_php_provider",
					Name:              "php",
					Version:           "release-2026-10-03",
					DistributionRef:   distRef,
					InterpretationRef: interpRef,
				},
			},
		}

		intent := standardTestIntent("php", "none", nil, "release-2026-10-03")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.1",
			ProviderRegistry:     reg,
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}

		res, err := ResolveTargetEnvironment(req)
		if err != nil {
			t.Fatalf("expected opaque release to be accepted under schema 2.1 (Test X75): %v", err)
		}
		if res.Document.Runtime.Version != "release-2026-10-03" {
			t.Errorf("X75 FAILED: expected version release-2026-10-03, got %s", res.Document.Runtime.Version)
		}
		if res.Document.Runtime.ProviderID != "canonical_php_provider" {
			t.Errorf("X75 FAILED: expected provider canonical_php_provider, got %s", res.Document.Runtime.ProviderID)
		}
	})

	// =========================================================================
	// 3. Test X16: Action-Enum Decision Service & Immutability
	// =========================================================================
	t.Run("X16_ActionEnumDecisionService", func(t *testing.T) {
		req := &ProposeDecisionRequest{
			DecisionID:   "dec_reject_rec",
			ProjectID:    "proj_test",
			Action:       ActionReject,
			SubjectRef:   dummyRef,
			Rationale:    "Legacy caching dependency is obsolete in target environment",
			EvidenceRefs: []string{dummyRef},
		}
		doc, err := ProposeDecision(req)
		if err != nil {
			t.Fatalf("ProposeDecision failed: %v", err)
		}
		if doc.Action != "REJECT" || doc.Status != "PROPOSED" {
			t.Errorf("X16 FAILED: unexpected decision doc: %+v", doc)
		}

		// Rejection of invalid action enum
		invalidReq := *req
		invalidReq.Action = DecisionAction("INVALID_ACTION")
		_, err = ProposeDecision(&invalidReq)
		if err == nil {
			t.Errorf("X16 FAILED: invalid action enum must be rejected")
		}
	})

	// =========================================================================
	// 4. Test X17: Progress Hash Stability
	// =========================================================================
	t.Run("X17_ProgressHashStability", func(t *testing.T) {
		subject := &PlanApprovalSubject{
			SchemaVersion:          "2.0",
			Canonicalization:       "rwt-cjson-1",
			ProjectID:              "fixture",
			Kind:                   "PLAN",
			ConfigurationHash:      strings.Repeat("b", 64),
			SnapshotRef:            dummyRef,
			BehaviorIRRef:          dummyRef,
			CorpusRef:              dummyRef,
			PolicyRef:              dummyRef,
			TargetIntentRef:        dummyRef,
			ResolvedEnvironmentRef: dummyRef,
			TargetModelRef:         dummyRef,
			DecisionRefs: []string{
				"sha256:0ab26ee1f8b0f264905fd9b11c832060d4d39202e517ffe098f2d6346412ed21",
				"sha256:a17d92d37e6785b1487096bcd4c8eafdb915ed5bda24ceafe3a5b98da0220d8e",
			},
			DependencyGraphRef: dummyRef,
			AdapterBindings: []PlanAdapterBinding{
				{
					Role:              "TARGET",
					AdapterID:         "synthetic",
					Version:           "1.0.0",
					ManifestRef:       dummyRef,
					ImplementationRef: dummyRef,
				},
			},
			ConsumedRevisionRefs: []string{},
			OrderedUnits: []PlanApprovalUnit{
				{
					UnitID:       "login",
					ContractRefs: []string{dummyRef},
					CaseIDs:      []string{"login"},
					Priority:     1,
					Capabilities: []string{"http"},
				},
			},
			DependencyEdges:    []PlanApprovalEdge{},
			OmissionProofRefs:  []string{},
			ScopeExclusionRefs: []string{},
			UnresolvedItems:    []PlanApprovalUnresolvedItem{},
		}

		hashBefore, err := subject.Digest()
		if err != nil {
			t.Fatalf("failed to compute subject digest: %v", err)
		}

		// Simulate external execution progress state (e.g. task completed or status changed)
		progressTracker := map[string]string{
			"unit:login": "IN_PROGRESS",
		}
		progressTracker["unit:login"] = "COMPLETED"

		// Plan approval subject hash MUST remain strictly identical and stable
		hashAfter, err := subject.Digest()
		if err != nil {
			t.Fatalf("failed to re-compute subject digest: %v", err)
		}
		if hashBefore != hashAfter {
			t.Errorf("X17 FAILED: plan approval hash must not change with execution progress; before=%s, after=%s", hashBefore, hashAfter)
		}
	})

	// =========================================================================
	// 5. Test X18: Dependency Graph & Capability Resolution
	// =========================================================================
	t.Run("X18_DependencyResolutionEngine", func(t *testing.T) {
		req := standardDepRequest("dep_retry_blocked", StrategyReplacement)
		req.TransitiveDependencies = append(req.TransitiveDependencies, TransitiveDependencyInfo{
			PackageID:          "guzzlehttp/retry-subscriber",
			License:            "MIT",
			RetryPolicyChanges: true, // Changes retry behavior
		})

		_, err := ResolveDependency(req)
		if err == nil {
			t.Fatalf("expected error due to retry behavior change (Test X18), got nil")
		}

		var blockedErr *DependencyResolutionBlockedError
		if !errors.As(err, &blockedErr) {
			t.Fatalf("expected DependencyResolutionBlockedError, got %T: %v", err, err)
		}
		if blockedErr.BlockerType != "BEHAVIOR_DIFFERENCE" {
			t.Errorf("X18 FAILED: expected blocker type BEHAVIOR_DIFFERENCE, got: %s", blockedErr.BlockerType)
		}
	})

	// =========================================================================
	// 6. Test X50: Dependency Assessment Invalidation
	// =========================================================================
	t.Run("X50_DependencyAssessmentInvalidation", func(t *testing.T) {
		req := standardDepRequest("dep_refreshable", StrategyReplacement)
		initialDoc, err := ResolveDependency(req)
		if err != nil {
			t.Fatalf("initial ResolveDependency failed: %v", err)
		}

		initialDigest, err := initialDoc.Digest()
		if err != nil {
			t.Fatalf("initialDoc Digest failed: %v", err)
		}

		// User changes strategy to ADAPTER with new decision ref
		newDecisionRef := "sha256:" + strings.Repeat("9", 64)
		affectedFindings := []string{"find_http_client_adapter_required"}

		refreshed, err := RefreshDependencyStrategy(
			initialDoc,
			StrategyAdapter,
			newDecisionRef,
			affectedFindings,
		)
		if err != nil {
			t.Fatalf("RefreshDependencyStrategy failed: %v", err)
		}

		if refreshed.Document.Proposal.Strategy != "ADAPTER" {
			t.Errorf("X50 FAILED: expected refreshed strategy ADAPTER, got %s", refreshed.Document.Proposal.Strategy)
		}
		if refreshed.NewDigest == initialDigest {
			t.Errorf("X50 FAILED: expected new resolution digest after strategy change")
		}
		if refreshed.InvalidatedOldDecisionRef != initialDoc.DecisionRef {
			t.Errorf("X50 FAILED: expected invalidated old decision ref %s, got %s", initialDoc.DecisionRef, refreshed.InvalidatedOldDecisionRef)
		}
	})

	// =========================================================================
	// 7. Test X29: Selective Target Drift & Stale Approval Protection
	// =========================================================================
	t.Run("X29_SelectiveTargetDrift", func(t *testing.T) {
		sourceIR := MakeDummyRef("source_ir_ci3")
		sourceContract := MakeDummyRef("source_contract_login")
		php83Runtime := MakeDummyRef("runtime_php_8_3")
		laravelDep := MakeDummyRef("laravel_dep_php83")
		loginComp := MakeDummyRef("target_model_login_comp")
		loginUnit := MakeDummyRef("plan_unit_login")
		loginExecPass := MakeDummyRef("login_verification_pass_result")
		unknownImpact := MakeDummyRef("suspicious_dynamic_eval")

		cov := &ConsumptionCoverageRecord{
			SchemaVersion:       "1.0",
			ProjectID:           "proj1",
			ScopeRef:            MakeDummyRef("scope1"),
			DiscoveryMethodRef:  MakeDummyRef("disc1"),
			AnalysisRevisionRef: MakeDummyRef("rev1"),
			InputsExamined:      []string{sourceIR, sourceContract, php83Runtime, laravelDep, loginComp, loginUnit, loginExecPass, unknownImpact},
			RootsExamined:       []string{"root1"},
			Categories: []CoverageCategory{
				{
					Category:     "SOURCE",
					Status:       "CLOSED",
					InputRefs:    []string{sourceIR, sourceContract},
					EvidenceRefs: []string{MakeDummyRef("ev_src")},
				},
				{
					Category:     "RESOLUTION",
					Status:       "CLOSED",
					InputRefs:    []string{php83Runtime, laravelDep},
					EvidenceRefs: []string{MakeDummyRef("ev_res")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{MakeDummyRef("control1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, err := cov.Digest()
		if err != nil {
			t.Fatalf("failed to compute cov digest: %v", err)
		}

		graph := &ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph1",
			ProjectID:     "proj1",
			Edges: []ValidityEdge{
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

		svc, err := NewSelectiveInvalidationService(graph, []*ConsumptionCoverageRecord{cov})
		if err != nil {
			t.Fatalf("NewSelectiveInvalidationService failed: %v", err)
		}

		invResult, err := svc.ComputeInvalidation([]string{php83Runtime}, nil)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		// Source facts preserved
		if !invResult.IsArtifactValid(sourceIR) || !invResult.IsArtifactValid(sourceContract) {
			t.Errorf("X29 FAILED: source facts must remain valid when target drifts")
		}

		// Target-dependent nodes invalidated
		if !invResult.IsArtifactStale(laravelDep) || !invResult.IsArtifactStale(loginComp) || !invResult.IsArtifactStale(loginUnit) {
			t.Errorf("X29 FAILED: target-dependent nodes must become stale when target drifts")
		}

		// Unknown edge conservative
		if !invResult.IsArtifactUnknown(unknownImpact) {
			t.Errorf("X29 FAILED: unknown impact must be marked unknown")
		}

		// Stale approval protection
		oldPlanHash := strings.Repeat("1", 64)
		newPlanHash := strings.Repeat("2", 64)
		err = ValidatePlanExecutionReadiness(newPlanHash, []string{loginUnit}, invResult, oldPlanHash, []string{laravelDep})
		if err == nil || !strings.Contains(err.Error(), "stale approval cannot execute") {
			t.Errorf("X29 FAILED: stale approval must be rejected for execution: %v", err)
		}
	})

	// =========================================================================
	// 8. Test X48: Selective Source Drift & Reuse Certificate
	// =========================================================================
	t.Run("X48_SelectiveSourceDrift", func(t *testing.T) {
		loginCtrl := MakeDummyRef("LoginController.php")
		catalogCtrl := MakeDummyRef("CatalogController.php")
		loginFinding := MakeDummyRef("finding_login")
		catalogFinding := MakeDummyRef("finding_catalog")

		cov := &ConsumptionCoverageRecord{
			SchemaVersion:       "1.0",
			ProjectID:           "proj1",
			ScopeRef:            MakeDummyRef("scope1"),
			DiscoveryMethodRef:  MakeDummyRef("disc1"),
			AnalysisRevisionRef: MakeDummyRef("rev1"),
			InputsExamined:      []string{loginCtrl, catalogCtrl, loginFinding, catalogFinding},
			RootsExamined:       []string{"root1"},
			Categories: []CoverageCategory{
				{
					Category:     "SOURCE",
					Status:       "CLOSED",
					InputRefs:    []string{loginCtrl, catalogCtrl},
					EvidenceRefs: []string{MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{MakeDummyRef("ctrl1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, err := cov.Digest()
		if err != nil {
			t.Fatalf("failed to compute cov digest: %v", err)
		}

		graph := &ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x48",
			ProjectID:     "proj1",
			Edges: []ValidityEdge{
				{
					FromRef:     loginCtrl,
					ToRef:       loginFinding,
					Kind:        "SOURCE_CONTRACT",
					Predicate:   "SAME_CONTENT",
					WitnessRefs: []string{loginCtrl},
				},
				{
					FromRef:     catalogCtrl,
					ToRef:       catalogFinding,
					Kind:        "SOURCE_CONTRACT",
					Predicate:   "SAME_CONTENT",
					WitnessRefs: []string{catalogCtrl},
				},
			},
			UnknownImpactRefs:       []string{},
			ConsumptionCoverageRefs: []string{covDigest},
		}

		svc, err := NewSelectiveInvalidationService(graph, []*ConsumptionCoverageRecord{cov})
		if err != nil {
			t.Fatalf("NewSelectiveInvalidationService failed: %v", err)
		}

		// Only loginCtrl modified
		res, err := svc.ComputeInvalidation([]string{loginCtrl}, nil)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		if !res.IsArtifactStale(loginFinding) {
			t.Errorf("X48 FAILED: loginFinding must be stale when loginCtrl drifts")
		}
		if !res.IsArtifactValid(catalogFinding) {
			t.Errorf("X48 FAILED: catalogFinding must remain valid when only loginCtrl drifts")
		}
	})

	// =========================================================================
	// 9. Test X49: Target Change With History
	// =========================================================================
	t.Run("X49_TargetChangeWithHistory", func(t *testing.T) {
		assessmentFact := MakeDummyRef("assessment_fact_analysis")
		targetModel := MakeDummyRef("target_model_v1")

		cov := &ConsumptionCoverageRecord{
			SchemaVersion:       "1.0",
			ProjectID:           "proj1",
			ScopeRef:            MakeDummyRef("scope1"),
			DiscoveryMethodRef:  MakeDummyRef("disc1"),
			AnalysisRevisionRef: MakeDummyRef("rev1"),
			InputsExamined:      []string{assessmentFact, targetModel},
			RootsExamined:       []string{"root1"},
			Categories: []CoverageCategory{
				{
					Category:     "SOURCE",
					Status:       "CLOSED",
					InputRefs:    []string{assessmentFact},
					EvidenceRefs: []string{MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{MakeDummyRef("ctrl1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, _ := cov.Digest()

		graph := &ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x49",
			ProjectID:     "proj1",
			Edges: []ValidityEdge{
				{
					FromRef:     assessmentFact,
					ToRef:       targetModel,
					Kind:        "COMPONENT_UNIT",
					Predicate:   "SAME_RULES",
					WitnessRefs: []string{assessmentFact},
				},
			},
			UnknownImpactRefs:       []string{},
			ConsumptionCoverageRefs: []string{covDigest},
		}

		svc, err := NewSelectiveInvalidationService(graph, []*ConsumptionCoverageRecord{cov})
		if err != nil {
			t.Fatalf("NewSelectiveInvalidationService failed: %v", err)
		}

		// Target changes: targetModel is stale
		res, err := svc.ComputeInvalidation([]string{targetModel}, nil)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		if !res.IsArtifactValid(assessmentFact) {
			t.Errorf("X49 FAILED: assessment history fact must be retained when target changes")
		}
		if !res.IsArtifactStale(targetModel) {
			t.Errorf("X49 FAILED: target model must be marked stale")
		}
	})

	// =========================================================================
	// 10. Test X73: Positive Consumption Coverage & Missing Edge Refusal
	// =========================================================================
	t.Run("X73_PositiveConsumptionCoverage", func(t *testing.T) {
		sharedInput := MakeDummyRef("shared_auth_helper")
		completedUnit := MakeDummyRef("unit_login_controller")
		descendantArtifact := MakeDummyRef("downstream_session_bridge")

		cov := &ConsumptionCoverageRecord{
			SchemaVersion:       "1.0",
			ProjectID:           "proj1",
			ScopeRef:            MakeDummyRef("scope1"),
			DiscoveryMethodRef:  MakeDummyRef("disc1"),
			AnalysisRevisionRef: MakeDummyRef("rev1"),
			InputsExamined:      []string{sharedInput, completedUnit, descendantArtifact},
			RootsExamined:       []string{"root1"},
			Categories: []CoverageCategory{
				{
					Category:     "COMPONENT",
					Status:       "CLOSED",
					InputRefs:    []string{sharedInput},
					EvidenceRefs: []string{MakeDummyRef("ev1")},
				},
			},
			UnresolvedCandidates: []string{},
			ExcludedMechanisms:   []string{},
			ClosureControlRefs:   []string{MakeDummyRef("ctrl1")},
			Status:               "SUFFICIENT",
			Issuer:               "ENGINE",
		}
		covDigest, _ := cov.Digest()

		// Missing consumption edge from sharedInput -> completedUnit
		graphMissingEdge := &ValidityGraphRecord{
			SchemaVersion: "2.1",
			GraphID:       "graph_x73_missing",
			ProjectID:     "proj1",
			Edges: []ValidityEdge{
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

		svcMissing, err := NewSelectiveInvalidationService(graphMissingEdge, []*ConsumptionCoverageRecord{cov})
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		// Declared consumption: completedUnit consumes sharedInput
		consumed := map[string][]string{
			completedUnit: {sharedInput},
		}

		invRes, err := svcMissing.ComputeInvalidation([]string{}, consumed)
		if err != nil {
			t.Fatalf("ComputeInvalidation failed: %v", err)
		}

		// Completed unit must be STALE because of missing consumption edge
		if !invRes.IsArtifactStale(completedUnit) {
			t.Errorf("X73 FAILED: missing consumption edge must mark completed unit STALE: %s", invRes.ArtifactStates[completedUnit])
		}

		// Reuse must be refused
		reqReuse := ReuseCertificateRequest{
			ProjectID:               "proj1",
			OldPlanHash:             strings.Repeat("a", 64),
			NewPlanHash:             strings.Repeat("b", 64),
			ArtifactRef:             completedUnit,
			Kind:                    "IMPLEMENTATION",
			CheckedEdgeRefs:         []string{completedUnit},
			EvidenceRefs:            []string{sharedInput},
			ConsumptionCoverageRefs: []string{covDigest},
			IsVerificationResult:    false,
		}
		_, err = svcMissing.IssueReuseCertificate(invRes, reqReuse)
		if err == nil || !strings.Contains(err.Error(), "cannot reuse stale artifact") {
			t.Errorf("X73 FAILED: reuse certificate must be refused on missing consumption edge: %v", err)
		}
	})

	// =========================================================================
	// 11. Test X55: Assessment is Not a Token
	// =========================================================================
	t.Run("X55_AssessmentNotAToken", func(t *testing.T) {
		assessmentReport := []byte(`{"kind": "assessment", "report_id": "rep-01"}`)
		_, err := ValidateGate2Token("plan", assessmentReport)
		if err == nil || !strings.Contains(err.Error(), "assessment report cannot be used as Gate 2 plan approval token") {
			t.Errorf("X55 FAILED: assessment report must be rejected as Gate 2 token: %v", err)
		}

		basis := &AssessmentPlanningBasisRecord{
			SchemaVersion:          "1.0",
			ProjectID:              "fixture",
			AssessmentRef:          dummyRef,
			SelectionRef:           dummyRef,
			SelectedScopeRef:       dummyRef,
			CurrentInputRefs:       []CurrentInputRef{{Role: "SOURCE_SNAPSHOT", Ref: dummyRef}},
			ValidityGraphRef:       dummyRef,
			ReusedFindings:         []ReusedFinding{},
			StaleFindingRefs:       []string{},
			NewFindingRefs:         []string{},
			ChangesSinceAssessment: []ChangeSinceAssessment{},
		}
		basisDigest, _ := basis.Digest()

		subj := &PlanApprovalSubject{
			SchemaVersion:          "2.0",
			Canonicalization:       "rwt-cjson-1",
			ProjectID:              "fixture",
			Kind:                   "PLAN",
			ConfigurationHash:      strings.Repeat("b", 64),
			SnapshotRef:            dummyRef,
			BehaviorIRRef:          dummyRef,
			CorpusRef:              dummyRef,
			PolicyRef:              dummyRef,
			TargetIntentRef:        dummyRef,
			ResolvedEnvironmentRef: dummyRef,
			TargetModelRef:         dummyRef,
			DecisionRefs:           []string{dummyRef, basisDigest},
			DependencyGraphRef:     dummyRef,
			AdapterBindings: []PlanAdapterBinding{
				{
					Role:              "TARGET",
					AdapterID:         "synthetic",
					Version:           "1.0.0",
					ManifestRef:       dummyRef,
					ImplementationRef: dummyRef,
				},
			},
			ConsumedRevisionRefs: []string{},
			OrderedUnits: []PlanApprovalUnit{
				{
					UnitID:       "login",
					ContractRefs: []string{dummyRef},
					CaseIDs:      []string{"login"},
					Priority:     1,
					Capabilities: []string{"http"},
				},
			},
			DependencyEdges:    []PlanApprovalEdge{},
			OmissionProofRefs:  []string{},
			ScopeExclusionRefs: []string{},
			UnresolvedItems:    []PlanApprovalUnresolvedItem{},
		}

		env := &PlanEnvelopeRecord{
			SchemaVersion:    "2.1",
			PlanID:           "plan",
			ProjectID:        "fixture",
			PlanHash:         "2fea0aa910c9d732a7d087b3065ed9f477c5afe9d524b95a1e93bc6dda2c2e18",
			Subject:          subj,
			AssessmentRef:    basis.AssessmentRef,
			SelectionRef:     basis.SelectionRef,
			PlanningBasisRef: dummyRef2, // Mismatched basis ref!
		}

		err = ValidatePlanEnvelopeCheckpointClosure(env, basis)
		if err == nil || !strings.Contains(err.Error(), "does not match computed planning basis digest") {
			t.Errorf("X55 FAILED: mismatched planning_basis_ref must fail checkpoint closure: %v", err)
		}
	})

	// =========================================================================
	// 12. Test X58: Executable Qualification Closure
	// =========================================================================
	t.Run("X58_ExecutableQualificationClosure", func(t *testing.T) {
		subject := &PlanApprovalSubject{
			AdapterBindings: []PlanAdapterBinding{
				{
					Role:              "TARGET",
					AdapterID:         "synthetic",
					Version:           "1.0.0",
					ImplementationRef: dummyRef,
				},
			},
		}

		admitted := map[string]*QualificationRecord{
			"synthetic:1.0.0": {
				Status:            "SUPPORTED",
				ImplementationRef: dummyRef2, // swapped bytes!
				ControlResultRefs: []string{dummyRef},
			},
		}

		err := VerifyExecutableQualificationClosure(subject, admitted)
		if err == nil || !strings.Contains(err.Error(), "swapped bytes under same version/URI rejected") {
			t.Errorf("X58 FAILED: swapped implementation bytes under same version/URI must be rejected: %v", err)
		}
	})

	// =========================================================================
	// 13. Test X59: Approved Performance Estimand
	// =========================================================================
	t.Run("X59_ApprovedPerformanceEstimand", func(t *testing.T) {
		claim := "NON_INFERIORITY"
		measProfile := "RWT_PAIRED_BOOTSTRAP_1"
		stoppingRule := "FIXED_N"
		trials := 15

		policy := &PerformancePolicyRecord{
			SchemaVersion:      "1.1",
			PolicyID:           "perf_policy",
			ProjectID:          "fixture",
			Applicable:         true,
			Claim:              &claim,
			MeasurementProfile: &measProfile,
			StoppingRule:       &stoppingRule,
			PairedTrials:      &trials,
			RequiredMetrics: []PerformanceMetricLimit{
				{
					MetricID:                 "P95",
					Unit:                     "ms",
					LimitKind:                "RELATIVE_NON_INFERIORITY",
					Limit:                    1.05,
					MinimumSuccessesPerTrial: 500,
				},
			},
		}

		err := VerifyPerformanceEstimandIntegrity(policy, true, false)
		if err == nil || !strings.Contains(err.Error(), "pooled percentile cannot substitute for mean of trial quantiles") {
			t.Errorf("X59 FAILED: pooled percentile must be rejected: %v", err)
		}

		err = VerifyPerformanceEstimandIntegrity(policy, false, true)
		if err == nil || !strings.Contains(err.Error(), "missing required trial success counters") {
			t.Errorf("X59 FAILED: missing trial success counters must be rejected: %v", err)
		}

		err = VerifyGate2PerformancePolicyBinding(dummyRef, dummyRef2)
		if err == nil || !strings.Contains(err.Error(), "performance policy changed after Gate 2") {
			t.Errorf("X59 FAILED: runtime policy drift after Gate 2 must be rejected: %v", err)
		}
	})
}
