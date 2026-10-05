package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestX42_BlockedPlanningUsefulAssessment verifies that missing runtime extensions or unsupported behavior
// prevents executable planning while assessment completes with BLOCKED feasibility and implementation_authorized=false per Test X42.
func TestX42_BlockedPlanningUsefulAssessment(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)
	configHash := strings.Repeat("b", 64)

	input := &AssessmentCompilerInput{
		AssessmentID:         "assess_x42_blocked",
		ProjectID:            "proj-x42",
		ConfigurationHash:    configHash,
		SourceSnapshotRef:    dummyRef,
		DetectionRef:         dummyRef,
		ScopeRef:             dummyRef,
		ValidityGraphRef:     dummyRef,
		RuntimeBlockers:      []string{"Missing PHP redis extension", "Dynamic eval() execution unresolvable"},
		MigrationFeasibility: "BLOCKED",
		AsOfSequence:         1,
		CreatedAt:            "2026-10-05T00:00:00Z",
	}

	doc, err := CompileAssessmentReport(input)
	if err != nil {
		t.Fatalf("CompileAssessmentReport failed: %v", err)
	}

	// 1. Verify AssessmentStatus is COMPLETED
	if doc.AssessmentStatus != "COMPLETED" {
		t.Errorf("expected COMPLETED status, got %s", doc.AssessmentStatus)
	}

	// 2. Verify MigrationFeasibility is BLOCKED
	if doc.MigrationFeasibility != "BLOCKED" {
		t.Errorf("expected BLOCKED feasibility, got %s", doc.MigrationFeasibility)
	}

	// 3. Verify ImplementationAuthorized is strictly FALSE
	if doc.ImplementationAuthorized {
		t.Errorf("ARCHITECTURAL INVARIANT VIOLATION: implementation_authorized must be false")
	}

	// 4. Verify blocker finding was generated
	var foundBlocker bool
	for _, f := range doc.Findings {
		if f.Category == "BLOCKER" {
			foundBlocker = true
			if !strings.Contains(f.Reason, "Missing PHP redis extension") {
				t.Errorf("expected blocker reason to contain runtime blocker detail, got %s", f.Reason)
			}
		}
	}
	if !foundBlocker {
		t.Errorf("expected at least one BLOCKER finding in blocked assessment report")
	}

	// 5. Verify document validates against canonical schema
	cjson, err := doc.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if err := ValidateAssessmentReport(cjson); err != nil {
		t.Fatalf("ValidateAssessmentReport failed: %v", err)
	}
}

// TestX44_ProductionNonMutation verifies that assessment execution causes zero writes
// or mutations to declared live production endpoints per Test X44.
func TestX44_ProductionNonMutation(t *testing.T) {
	prodEndpoints := []string{
		"production.mysql.internal:3306",
		"live-api.payments.com/v1",
		"prod-redis.cache.internal:6379",
	}

	// 1. Positive Control: Read-only queries or writes to isolated disposable sandbox
	cleanAttempts := []EffectAttempt{
		{TargetEndpoint: "sandbox.mysql.internal:3306", Operation: "WRITE", IsProduction: false},
		{TargetEndpoint: "production.mysql.internal:3306", Operation: "READ", IsProduction: true},
		{TargetEndpoint: "mock-api.payments.local", Operation: "CONNECT", IsProduction: false},
	}

	cleanRes, err := VerifyProductionNonMutation(cleanAttempts, prodEndpoints)
	if err != nil {
		t.Fatalf("clean VerifyProductionNonMutation failed: %v", err)
	}
	if !cleanRes.Passed || len(cleanRes.ProductionWritesDetected) != 0 {
		t.Errorf("expected clean assessment to pass production non-mutation audit")
	}

	// 2. Mutant: Deliberate attempted write to production endpoint
	mutantAttempts := []EffectAttempt{
		{TargetEndpoint: "production.mysql.internal:3306", Operation: "WRITE", IsProduction: true},
		{TargetEndpoint: "live-api.payments.com/v1", Operation: "DELETE", IsProduction: true},
	}

	mutantRes, err := VerifyProductionNonMutation(mutantAttempts, prodEndpoints)
	if err != nil {
		t.Fatalf("mutant VerifyProductionNonMutation failed: %v", err)
	}
	if mutantRes.Passed {
		t.Errorf("FALSIFICATION FAILED: writes to production endpoints must fail audit")
	}
	if len(mutantRes.ProductionWritesDetected) != 2 {
		t.Errorf("expected 2 detected production write violations, got %d", len(mutantRes.ProductionWritesDetected))
	}
}

// TestX48_SelectiveSourceInvalidation verifies that changing one controller invalidates only
// its transitive reverse dependency closure, leaving unrelated findings valid per Test X48.
func TestX48_SelectiveSourceInvalidation(t *testing.T) {
	// Dependency graph:
	// UserController -> UserService -> UserModel
	// PaymentController -> PaymentService -> PaymentGateway
	// Unrelated: ReportController -> ReportView
	// Unknown impact: PluginLoader (dynamic include)
	graph := &DependencyGraph{
		ReverseEdges: map[string][]string{
			"UserController":    {"UserService"},
			"UserService":       {"UserModel"},
			"PaymentController": {"PaymentService"},
			"PaymentService":    {"PaymentGateway"},
			"ReportController":  {"ReportView"},
		},
		NodeFindingMap: map[string][]string{
			"UserController":    {"find_user_ctrl"},
			"UserService":       {"find_user_svc"},
			"UserModel":         {"find_user_model"},
			"PaymentController": {"find_pay_ctrl"},
			"PaymentService":    {"find_pay_svc"},
			"PaymentGateway":    {"find_pay_gw"},
			"ReportController":  {"find_report_ctrl"},
			"ReportView":        {"find_report_view"},
		},
		UnknownImpactNodes: map[string]bool{
			"PluginLoader": true,
		},
	}

	allNodes := []string{
		"UserController", "UserService", "UserModel",
		"PaymentController", "PaymentService", "PaymentGateway",
		"ReportController", "ReportView",
		"PluginLoader",
	}

	allFindings := []string{
		"find_user_ctrl", "find_user_svc", "find_user_model",
		"find_pay_ctrl", "find_pay_svc", "find_pay_gw",
		"find_report_ctrl", "find_report_view",
	}

	// Change UserController:
	// Reverse closure: UserController, UserService, UserModel, plus unknown impact PluginLoader
	// Unrelated nodes: PaymentController, PaymentService, PaymentGateway, ReportController, ReportView
	inv := InvalidateReverseDependencies(graph, allNodes, allFindings, []string{"UserController"})

	// Verify UserController and dependents are stale
	for _, n := range []string{"UserController", "UserService", "UserModel", "PluginLoader"} {
		found := false
		for _, sn := range inv.StaleNodes {
			if sn == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be in StaleNodes, got %v", n, inv.StaleNodes)
		}
	}

	// Verify PaymentController, ReportController and their findings remain VALID!
	for _, n := range []string{"PaymentController", "PaymentService", "PaymentGateway", "ReportController", "ReportView"} {
		found := false
		for _, vn := range inv.ValidNodes {
			if vn == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to remain in ValidNodes, got %v", n, inv.ValidNodes)
		}
	}

	// Verify findings: user findings are stale, payment/report findings are valid
	for _, f := range []string{"find_user_ctrl", "find_user_svc", "find_user_model"} {
		found := false
		for _, sf := range inv.StaleFindings {
			if sf == f {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be in StaleFindings", f)
		}
	}

	for _, f := range []string{"find_pay_ctrl", "find_pay_svc", "find_pay_gw", "find_report_ctrl", "find_report_view"} {
		found := false
		for _, vf := range inv.ValidFindings {
			if vf == f {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to remain in ValidFindings", f)
		}
	}
}

// TestX51_BlockedReportRetrieval verifies that a report with blocked feasibility
// is retrievable byte-identically without Gate 2 or executable planning per Test X51.
func TestX51_BlockedReportRetrieval(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)
	configHash := strings.Repeat("c", 64)

	input := &AssessmentCompilerInput{
		AssessmentID:         "assess_x51_retrieval",
		ProjectID:            "proj-x51",
		ConfigurationHash:    configHash,
		SourceSnapshotRef:    dummyRef,
		DetectionRef:         dummyRef,
		ScopeRef:             dummyRef,
		ValidityGraphRef:     dummyRef,
		RuntimeBlockers:      []string{"Unavailable native C extension"},
		MigrationFeasibility: "BLOCKED",
		AsOfSequence:         10,
		CreatedAt:            "2026-10-05T00:00:00Z",
	}

	doc, err := CompileAssessmentReport(input)
	if err != nil {
		t.Fatalf("CompileAssessmentReport failed: %v", err)
	}

	cjson, err := doc.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}

	digest, err := doc.Digest()
	if err != nil {
		t.Fatalf("Digest failed: %v", err)
	}

	// Simulate storage round-trip (e.g. CLI/MCP retrieval without Gate 2)
	var retrieved AssessmentReportDocument
	if err := json.Unmarshal(cjson, &retrieved); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	retrievedDigest, err := retrieved.Digest()
	if err != nil {
		t.Fatalf("retrieved Digest failed: %v", err)
	}

	if digest != retrievedDigest {
		t.Errorf("retrieved report digest mismatch: expected %s, got %s", digest, retrievedDigest)
	}

	if retrieved.MigrationFeasibility != "BLOCKED" {
		t.Errorf("expected BLOCKED feasibility, got %s", retrieved.MigrationFeasibility)
	}
	if retrieved.ImplementationAuthorized {
		t.Errorf("implementation_authorized must remain false on retrieval")
	}
}

// TestX54_FindingClassification verifies that static findings require limitations,
// missing evidence cannot be observed, and planted static claims of runtime parity fail per Test X54.
func TestX54_FindingClassification(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("d", 64)

	// 1. Positive Control: Static finding with explicit limitations
	validStatic := AssessmentFindingRecord{
		FindingID:         "find_ok_static",
		AnalysisRecordRef: dummyRef,
		Classification:    "INFERRED",
		Basis:             "STATIC_ONLY",
		Category:          "CHARACTERIZATION",
		Summary:           "Static route map extracted from routes.php",
		Reason:            "AST traversal identified 15 route patterns",
		ProvenanceRefs:    []string{dummyRef},
		PrerequisiteRefs:  []string{dummyRef},
		Limitations:       []string{"Dynamic routes with _remap cannot be statically resolved"},
		AffectedIDs:       []string{"proj"},
		RecommendedAction: nil,
	}

	if err := ValidateFindingClassification(&validStatic); err != nil {
		t.Errorf("expected valid static finding to pass, got: %v", err)
	}

	// 2. Mutant A: Static finding with empty limitations
	badStatic := validStatic
	badStatic.Limitations = []string{}
	if err := ValidateFindingClassification(&badStatic); err == nil {
		t.Errorf("FALSIFICATION FAILED: static finding with empty limitations must fail")
	}

	// 3. Mutant B: Missing evidence marked as OBSERVED
	badMissing := validStatic
	badMissing.Basis = "MISSING_EVIDENCE"
	badMissing.Classification = "OBSERVED"
	if err := ValidateFindingClassification(&badMissing); err == nil {
		t.Errorf("FALSIFICATION FAILED: missing evidence classified as OBSERVED must fail")
	}

	// 4. Mutant C: Planted static claim of runtime parity
	plantedParity := validStatic
	plantedParity.Summary = "Static analysis verifies complete runtime parity with target"
	if err := ValidateFindingClassification(&plantedParity); err == nil {
		t.Errorf("FALSIFICATION FAILED: planted static claim of runtime parity must fail")
	} else if !strings.Contains(err.Error(), "ERR_PLANTED_RUNTIME_CLAIM") {
		t.Errorf("expected ERR_PLANTED_RUNTIME_CLAIM error, got: %v", err)
	}
}
