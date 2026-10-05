package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

// TestM2_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M2:
// Source Qualification, covering all 15 M2 tests and demonstrating deterministic characterization
// under an explicit boundary while preserving uncertainty.
func TestM2_AcceptanceClosure(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)
	configHash := strings.Repeat("0", 64)

	// =========================================================================
	// 1. Test X68: Filesystem Semantic Qualification
	// =========================================================================
	t.Run("X68_FilesystemSemantics", func(t *testing.T) {
		profile := &FilesystemProfileRecord{
			SchemaVersion:      "1.0",
			ProjectID:          "proj-m2",
			ContentManifestRef: dummyRef,
			Properties: []FilesystemPropertyRecord{
				{Property: "PERMISSIONS", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
				{Property: "EXECUTABLE", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
				{Property: "MTIME", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
				{Property: "OWNERSHIP", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
				{Property: "HARDLINKS", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
				{Property: "SYMLINKS", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
				{Property: "CASE_SENSITIVITY", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
				{Property: "PATH_RESOLUTION", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
				{Property: "WORKING_DIRECTORY", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
				{Property: "DIRECTORY_ORDER", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
			},
			ControlRefs: []string{dummyRef},
			Limitations: []string{},
		}

		cHash := strings.Repeat("e", 64)
		base := map[string]FileMetadataState{
			"cli/run.php": {Path: "cli/run.php", ContentHash: cHash, Mode: 0755, IsExecutable: true, MtimeNs: 100, Owner: "root"},
		}
		drift := map[string]FileMetadataState{
			"cli/run.php": {Path: "cli/run.php", ContentHash: cHash, Mode: 0644, IsExecutable: false, MtimeNs: 100, Owner: "root"},
		}

		res, err := QualifyFilesystemProfile(profile, base, drift)
		if err != nil {
			t.Fatalf("QualifyFilesystemProfile failed: %v", err)
		}
		if res.Valid || !res.Stale || res.Status != AttestationStatusUnavailable || !res.SuccessorProfileRequired {
			t.Errorf("X68 FAILED: executable bit drift must invalidate qualification to STALE/UNAVAILABLE")
		}
	})

	// =========================================================================
	// 2. Test X69: Embedded Credential Admission
	// =========================================================================
	t.Run("X69_EmbeddedCredentialAdmission", func(t *testing.T) {
		files := map[string][]byte{
			"application/config/database.php": []byte(`<?php $db['default']['password'] = 'canary_prod_live_key_7712';`),
		}
		scan := ScanSourceCredentials(files)
		if scan.CredentialStatus != "LIVE_OR_UNKNOWN" {
			t.Fatalf("X69 FAILED: expected LIVE_OR_UNKNOWN, got %s", scan.CredentialStatus)
		}

		// Attempting ADMITTED_ORIGINAL must be refused
		_, err := EvaluateExecutionAdmission("proj-m2", dummyRef, configHash, scan, dummyRef, "ADMITTED_ORIGINAL", nil)
		if err == nil || !strings.Contains(err.Error(), "ERR_UNSAFE_ORIGINAL_CREDENTIALS") {
			t.Errorf("X69 FAILED: live canary must refuse ADMITTED_ORIGINAL: %v", err)
		}

		// STATIC_ONLY must succeed
		staticAdm, err := EvaluateExecutionAdmission("proj-m2", dummyRef, configHash, scan, dummyRef, "STATIC_ONLY", nil)
		if err != nil || staticAdm.Outcome != "STATIC_ONLY" {
			t.Errorf("X69 FAILED: STATIC_ONLY admission must succeed: %v", err)
		}
	})

	// =========================================================================
	// 3. Test X70: Server-Side Database Effects
	// =========================================================================
	t.Run("X70_ServerSideDatabaseEffects", func(t *testing.T) {
		driver := &DatabaseDriverCapture{
			ExecutedQueries: []string{"INSERT INTO users (id) VALUES (1)"},
			TargetTables:    []string{"users"},
		}
		serverMissing := &ServerDatabaseCapture{MechanismInstalled: false}
		resMissing, _ := VerifyDatabaseEffectsCompleteness(driver, serverMissing)
		if resMissing.CompletenessStatus != AttestationStatusUnavailable {
			t.Errorf("X70 FAILED: missing server capture must yield UNAVAILABLE status")
		}

		serverWithTrigger := &ServerDatabaseCapture{
			MechanismInstalled: true,
			ModifiedTables:     []string{"users", "user_audit_logs"},
		}
		resOmission, _ := VerifyDatabaseEffectsCompleteness(driver, serverWithTrigger)
		if resOmission.CompletenessStatus != AttestationStatusPartial || len(resOmission.OmittedWrites) != 1 {
			t.Errorf("X70 FAILED: omitted server trigger write must yield PARTIAL status with omitted table")
		}
	})

	// =========================================================================
	// 4. Test X76: Conservative Code Classification Precedence
	// =========================================================================
	t.Run("X76_ClassificationOverlap", func(t *testing.T) {
		// Node with dynamic unclosed edge must remain UNKNOWN, never becoming DEAD
		req := &ClassificationRequest{
			ProjectID:        "proj-m2",
			Candidates:       []string{"root_http", "node_dyn"},
			OperationalRoots: []string{"root_http"},
			Edges: []DirectedEdge{
				{FromID: "root_http", ToID: "node_dyn", Resolved: false},
			},
			UnresolvedNodes: map[string][]string{
				"node_dyn": {"dynamic dispatch"},
			},
			ObservedHits: map[string][]string{
				"root_http": {"ev-1"},
			},
			ClosedBoundary: false,
			GraphRef:       dummyRef,
			ProfileRef:     dummyRef,
			CorpusRef:      dummyRef,
			SnapshotRef:    dummyRef,
		}

		res, err := ClassifyCode(req)
		if err != nil {
			t.Fatalf("ClassifyCode failed: %v", err)
		}
		if res.Classifications["node_dyn"].State != ClassificationUnknown {
			t.Errorf("X76 FAILED: dynamic unclosed candidate must be UNKNOWN, got %s", res.Classifications["node_dyn"].State)
		}
	})

	// =========================================================================
	// 5. Test X09: Dynamic Dispatch and Cron Roots Discovery
	// =========================================================================
	t.Run("X09_DynamicDispatchAndCronRoots", func(t *testing.T) {
		ci3Files := map[string][]byte{
			"application/controllers/Tasks.php": []byte(`<?php
class Tasks extends CI_Controller {
    public function run_cron() {
        if ($this->input->is_cli_request()) {
            $dyn = "action";
            $this->$dyn();
        }
    }
}
`),
		}
		input := &CI3AdapterInput{
			ProjectID:         "proj-m2",
			ConfigurationHash: configHash,
			DocumentID:        "doc-x09",
			Files:             ci3Files,
		}
		res, err := ExtractCI3(input)
		if err != nil {
			t.Fatalf("ExtractCI3 failed: %v", err)
		}
		if !res.DynamicDispatchDetected {
			t.Errorf("X09 FAILED: dynamic dispatch must be detected")
		}

		// Verify that dynamic dispatch edge prevents candidate from becoming DEAD
		hasDynFinding := false
		for _, f := range res.SourceIR.Findings {
			if f.Code == "DYNAMIC_DISPATCH" {
				hasDynFinding = true
				break
			}
		}
		if !hasDynFinding {
			t.Errorf("X09 FAILED: expected DYNAMIC_DISPATCH finding in Source IR")
		}
	})

	// =========================================================================
	// 6. Test X10: Vendor / Framework Patches & Environment Contradiction
	// =========================================================================
	t.Run("X10_ModifiedEnvironmentAndVendorPatch", func(t *testing.T) {
		canonicalVendorBytes := []byte("<?php // Canonical CodeIgniter 3.1.13 core ?>")
		h := sha256.Sum256(canonicalVendorBytes)
		canonicalHex := hex.EncodeToString(h[:])

		patchedVendorBytes := []byte("<?php // Patched CodeIgniter 3.1.13 core with inline fix ?>")

		ci3Files := map[string][]byte{
			"system/core/CodeIgniter.php": patchedVendorBytes,
			"index.php":                   []byte("<?php define('ENVIRONMENT', 'development');"),
		}

		knownHashes := map[string]string{
			"system/core/CodeIgniter.php": canonicalHex,
		}

		input := &CI3AdapterInput{
			ProjectID:           "proj-m2",
			ConfigurationHash:   configHash,
			DocumentID:          "doc-x10",
			Files:               ci3Files,
			KnownVendorHashes:   knownHashes,
			DeclaredEnvironment: "production", // contradicts development in index.php
		}

		res, err := ExtractCI3(input)
		if err != nil {
			t.Fatalf("ExtractCI3 failed: %v", err)
		}
		if len(res.ModifiedVendorFiles) != 1 || res.ModifiedVendorFiles[0] != "system/core/CodeIgniter.php" {
			t.Errorf("X10 FAILED: modified vendor file not identified: %v", res.ModifiedVendorFiles)
		}

		var foundModVendor, foundEnvMismatch bool
		for _, f := range res.SourceIR.Findings {
			if f.Code == "MODIFIED_VENDOR" {
				foundModVendor = true
			}
			if f.Code == "ENVIRONMENT_MISMATCH" {
				foundEnvMismatch = true
			}
		}
		if !foundModVendor || !foundEnvMismatch {
			t.Errorf("X10 FAILED: expected MODIFIED_VENDOR and ENVIRONMENT_MISMATCH findings")
		}
	})

	// =========================================================================
	// 7. Test X11: Probe Bypass Detection
	// =========================================================================
	t.Run("X11_ProbeBypass", func(t *testing.T) {
		mechs := []ObservationMechanism{
			{
				MechanismID:        "db_framework",
				Dimension:          DimensionDB,
				BoundaryRef:        dummyRef,
				InstrumentationRef: &dummyRef,
				SupportedPaths:     []string{"app/models"},
				ControlTestRefs:    []string{dummyRef},
				ClosureTestRefs:    []string{dummyRef},
				Coverage:           MechanismCoverageClosed,
			},
		}

		bypassFiles := map[string][]byte{
			"app/models/Raw.php": []byte(`<?php $pdo = new PDO("mysql:host=localhost;dbname=test", "root", "");`),
		}

		bypassRes, err := DetectProbeBypass(bypassFiles, mechs)
		if err != nil {
			t.Fatalf("DetectProbeBypass failed: %v", err)
		}
		if !bypassRes.CompletenessInvalidated || len(bypassRes.EscapePathsDetected) != 1 {
			t.Errorf("X11 FAILED: direct PDO call must be detected as escape path")
		}

		att, err := DeriveObservationAttestation(&DeriveAttestationRequest{
			AttestationID: "att-x11",
			ProjectID:     "proj-m2",
			ProfileRef:    dummyRef,
			Mechanisms:    bypassRes.MechanismsWithEscapes,
		})
		if err != nil {
			t.Fatalf("DeriveObservationAttestation failed: %v", err)
		}
		if att.DerivedStatus == AttestationStatusComplete {
			t.Errorf("X11 FAILED: unmonitored PDO escape path must never derive COMPLETE status")
		}
	})

	// =========================================================================
	// 8. Test X12: Async Causality & Delivery Acknowledgment
	// =========================================================================
	t.Run("X12_AsyncCausality", func(t *testing.T) {
		traceMissingAck := &AsyncObservationTrace{
			HTTPSuccess:         true,
			DispatchSpanID:      "span-http",
			InitiatingCauseID:   "cause-1",
			DeliverySpanID:      "span-email",
			DeliveryAckRecorded: false, // missing ack!
		}
		resA, _ := VerifyAsyncCausality(traceMissingAck)
		if resA.Outcome == AsyncCausalityComplete {
			t.Errorf("X12 FAILED: missing delivery ack must not produce COMPLETE")
		}

		traceLate := &AsyncObservationTrace{
			HTTPSuccess:         true,
			DispatchSpanID:      "span-http",
			InitiatingCauseID:   "cause-1",
			ScheduleHorizonMs:   3000,
			DeliverySpanID:      "span-email",
			DeliveryElapsedMs:   8000, // exceeds horizon!
			DeliveryAckRecorded: true,
			DeliveryCauseIDs:    []string{"cause-1"},
		}
		resB, _ := VerifyAsyncCausality(traceLate)
		if resB.Outcome == AsyncCausalityComplete {
			t.Errorf("X12 FAILED: late delivery beyond horizon must not produce COMPLETE")
		}
	})

	// =========================================================================
	// 9. Test X13: Runtime Acquisition Prerequisites
	// =========================================================================
	t.Run("X13_RuntimeAcquisition", func(t *testing.T) {
		req := &RuntimePrerequisitesRequest{
			RequiredExtensions:  []string{"redis", "pdo_mysql"},
			AvailableExtensions: []string{"pdo_mysql"}, // redis missing!
			LockFileHash:        configHash,
			ExpectedLockHash:    configHash,
		}
		res, _ := VerifyRuntimePrerequisites(req)
		if !res.BootBlocked || res.Coverage != MechanismCoverageUnavailable {
			t.Errorf("X13 FAILED: missing required extension must block boot and set UNAVAILABLE coverage")
		}
	})

	// =========================================================================
	// 10. Test X14: Reset Dimensions & Contamination
	// =========================================================================
	t.Run("X14_ResetDimensions", func(t *testing.T) {
		contaminatedChecks := &ResetContaminationChecks{
			DDLImplicitCommitContamination: true,
		}
		res, _ := VerifyResetContamination(contaminatedChecks)
		if res.Outcome == ResetOutcomeRestored || !res.ContaminationPresent {
			t.Errorf("X14 FAILED: DDL implicit commit contamination must fail reset")
		}
	})

	// =========================================================================
	// 11. Test X42: Blocked Planning, Useful Assessment
	// =========================================================================
	t.Run("X42_BlockedPlanningUsefulAssessment", func(t *testing.T) {
		input := &AssessmentCompilerInput{
			AssessmentID:         "assess_m2_closure",
			ProjectID:            "proj-m2",
			ConfigurationHash:    configHash,
			SourceSnapshotRef:    dummyRef,
			DetectionRef:         dummyRef,
			ScopeRef:             dummyRef,
			ValidityGraphRef:     dummyRef,
			RuntimeBlockers:      []string{"Missing extension pdo_mysql"},
			MigrationFeasibility: "BLOCKED",
			AsOfSequence:         1,
		}
		doc, err := CompileAssessmentReport(input)
		if err != nil {
			t.Fatalf("CompileAssessmentReport failed: %v", err)
		}
		if doc.AssessmentStatus != "COMPLETED" || doc.MigrationFeasibility != "BLOCKED" {
			t.Errorf("X42 FAILED: expected COMPLETED assessment with BLOCKED feasibility")
		}
		if doc.ImplementationAuthorized {
			t.Errorf("X42 FAILED: implementation_authorized must be false")
		}
	})

	// =========================================================================
	// 12. Test X44: Production Non-Mutation
	// =========================================================================
	t.Run("X44_ProductionNonMutation", func(t *testing.T) {
		prodEndpoints := []string{"prod-db.internal:3306"}
		attempts := []EffectAttempt{
			{TargetEndpoint: "prod-db.internal:3306", Operation: "UPDATE", IsProduction: true},
		}
		audit, _ := VerifyProductionNonMutation(attempts, prodEndpoints)
		if audit.Passed {
			t.Errorf("X44 FAILED: write to production endpoint must be rejected in audit")
		}
	})

	// =========================================================================
	// 13. Test X48: Selective Reverse Dependency Invalidation
	// =========================================================================
	t.Run("X48_SelectiveSourceInvalidation", func(t *testing.T) {
		graph := &DependencyGraph{
			ReverseEdges: map[string][]string{
				"CartCtrl": {"CartService"},
			},
			NodeFindingMap: map[string][]string{
				"CartCtrl":    {"find_cart"},
				"CartService": {"find_cart_svc"},
				"CatalogCtrl": {"find_catalog"},
			},
			UnknownImpactNodes: map[string]bool{},
		}
		allNodes := []string{"CartCtrl", "CartService", "CatalogCtrl"}
		allFindings := []string{"find_cart", "find_cart_svc", "find_catalog"}

		inv := InvalidateReverseDependencies(graph, allNodes, allFindings, []string{"CartCtrl"})
		if len(inv.StaleNodes) != 2 || len(inv.ValidNodes) != 1 || inv.ValidNodes[0] != "CatalogCtrl" {
			t.Errorf("X48 FAILED: selective invalidation failed: Stale=%v, Valid=%v", inv.StaleNodes, inv.ValidNodes)
		}
	})

	// =========================================================================
	// 14. Test X51: Blocked Report Retrieval
	// =========================================================================
	t.Run("X51_BlockedReportRetrieval", func(t *testing.T) {
		input := &AssessmentCompilerInput{
			AssessmentID:         "assess_m2_retrieval",
			ProjectID:            "proj-m2",
			ConfigurationHash:    configHash,
			SourceSnapshotRef:    dummyRef,
			DetectionRef:         dummyRef,
			ScopeRef:             dummyRef,
			ValidityGraphRef:     dummyRef,
			RuntimeBlockers:      []string{"Blocked native runtime"},
			MigrationFeasibility: "BLOCKED",
			AsOfSequence:         1,
		}
		doc, err := CompileAssessmentReport(input)
		if err != nil {
			t.Fatalf("CompileAssessmentReport failed: %v", err)
		}

		cjson, err := doc.CanonicalJSON()
		if err != nil {
			t.Fatalf("CanonicalJSON failed: %v", err)
		}

		var retrieved AssessmentReportDocument
		if err := json.Unmarshal(cjson, &retrieved); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if retrieved.MigrationFeasibility != "BLOCKED" || retrieved.ImplementationAuthorized {
			t.Errorf("X51 FAILED: retrieved report must retain BLOCKED and implementation_authorized=false")
		}
	})

	// =========================================================================
	// 15. Test X54: Finding Classification Verification
	// =========================================================================
	t.Run("X54_FindingClassification", func(t *testing.T) {
		plantedFinding := AssessmentFindingRecord{
			FindingID:         "find_fake_parity",
			AnalysisRecordRef: dummyRef,
			Classification:    "INFERRED",
			Basis:             "STATIC_ONLY",
			Category:          "CHARACTERIZATION",
			Summary:           "Static AST traversal proves full runtime parity with target",
			Reason:            "Planted claim",
			ProvenanceRefs:    []string{dummyRef},
			PrerequisiteRefs:  []string{dummyRef},
			Limitations:       []string{"none"},
			AffectedIDs:       []string{"proj"},
		}
		err := ValidateFindingClassification(&plantedFinding)
		if err == nil || !strings.Contains(err.Error(), "ERR_PLANTED_RUNTIME_CLAIM") {
			t.Errorf("X54 FAILED: planted static claim of runtime parity must be rejected: %v", err)
		}
	})
}

// TestM2_EndToEndReferenceCharacterization executes an integrated pipeline on a legacy CI3 application fixture,
// demonstrating source qualification across static IR extraction, reachability, probe attestation, filesystem semantics,
// credential boundary checks, and honest assessment compilation.
func TestM2_EndToEndReferenceCharacterization(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("c", 64)
	configHash := strings.Repeat("d", 64)

	fixtureFiles := map[string][]byte{
		"application/config/routes.php": []byte(`<?php
$route['default_controller'] = 'welcome';
$route['api/users']['get'] = 'api/get_users';
$route['api/users']['post'] = 'api/create_user';
`),
		"application/controllers/Welcome.php": []byte(`<?php
class Welcome extends CI_Controller {
    public function index() {
        $this->load->view('welcome_message');
    }
}
`),
		"application/controllers/Api.php": []byte(`<?php
class Api extends CI_Controller {
    public function get_users() {
        $this->load->model('User_model');
        $users = $this->User_model->get_all();
    }
    public function create_user() {
        $this->load->model('User_model');
        $this->User_model->insert(['name' => 'test']);
    }
}
`),
		"application/models/User_model.php": []byte(`<?php
class User_model extends CI_Model {
    public function get_all() {
        return $this->db->get('users')->result_array();
    }
    public function insert($data) {
        return $this->db->insert('users', $data);
    }
}
`),
		"application/config/database.php": []byte(`<?php
$db['default'] = array(
	'hostname' => 'localhost',
	'username' => 'synthetic_user',
	'password' => 'synthetic_password',
	'database' => 'synthetic_db',
);
`),
	}

	// 1. Static Parser & Route Extraction
	input := &CI3AdapterInput{
		ProjectID:         "proj-ci3-ref",
		ConfigurationHash: configHash,
		DocumentID:        "doc-ci3-ref",
		Files:             fixtureFiles,
	}
	extracted, err := ExtractCI3(input)
	if err != nil {
		t.Fatalf("ExtractCI3 failed: %v", err)
	}
	if len(extracted.SourceIR.Symbols) == 0 {
		t.Errorf("expected symbols in Source IR")
	}
	if len(extracted.ApplicationIR.Entrypoints) != 7 {
		t.Errorf("expected 7 HTTP entrypoints, got %d", len(extracted.ApplicationIR.Entrypoints))
	}

	// 2. Credential Scan
	credScan := ScanSourceCredentials(fixtureFiles)
	if credScan.CredentialStatus != "SYNTHETIC_OR_ABSENT" {
		t.Errorf("expected SYNTHETIC_OR_ABSENT, got %s", credScan.CredentialStatus)
	}

	// 3. Probe Bypass Check
	mechs := []ObservationMechanism{
		{
			MechanismID:        "mech_http",
			Dimension:          DimensionHTTP,
			BoundaryRef:        dummyRef,
			InstrumentationRef: &dummyRef,
			SupportedPaths:     []string{"application/controllers"},
			ControlTestRefs:    []string{dummyRef},
			ClosureTestRefs:    []string{dummyRef},
			Coverage:           MechanismCoverageClosed,
		},
		{
			MechanismID:        "mech_db",
			Dimension:          DimensionDB,
			BoundaryRef:        dummyRef,
			InstrumentationRef: &dummyRef,
			SupportedPaths:     []string{"application/models"},
			ControlTestRefs:    []string{dummyRef},
			ClosureTestRefs:    []string{dummyRef},
			Coverage:           MechanismCoverageClosed,
		},
	}
	bypassRes, err := DetectProbeBypass(fixtureFiles, mechs)
	if err != nil {
		t.Fatalf("DetectProbeBypass failed: %v", err)
	}
	if bypassRes.CompletenessInvalidated {
		t.Errorf("expected no probe bypass in clean reference fixture")
	}

	// 4. Observation Attestation Derivation
	att, err := DeriveObservationAttestation(&DeriveAttestationRequest{
		AttestationID: "att-ref-fixture",
		ProjectID:     "proj-ci3-ref",
		ProfileRef:    dummyRef,
		Mechanisms:    bypassRes.MechanismsWithEscapes,
	})
	if err != nil {
		t.Fatalf("DeriveObservationAttestation failed: %v", err)
	}
	if att.DerivedStatus != AttestationStatusComplete {
		t.Errorf("expected COMPLETE attestation status, got %s", att.DerivedStatus)
	}

	// 5. Code Classification Engine
	var candidates []string
	for _, sym := range extracted.SourceIR.Symbols {
		candidates = append(candidates, sym.SymbolID)
	}
	classReq := &ClassificationRequest{
		ProjectID:        "proj-ci3-ref",
		Candidates:       candidates,
		OperationalRoots: []string{"sym_welcome_welcome", "sym_api_api"},
		GraphRef:         dummyRef,
		ProfileRef:       dummyRef,
		CorpusRef:        dummyRef,
		SnapshotRef:      dummyRef,
		ClosedBoundary:   true,
		OmissionProof:    nil,
	}
	classRes, err := ClassifyCode(classReq)
	if err != nil {
		t.Fatalf("ClassifyCode failed: %v", err)
	}
	if classRes.DeadCount != 0 {
		t.Errorf("no false DEAD expected in active reference code")
	}

	// 6. Compile Assessment Report
	assessDoc, err := CompileAssessmentReport(&AssessmentCompilerInput{
		AssessmentID:         "assess_ci3_ref",
		ProjectID:            "proj-ci3-ref",
		ConfigurationHash:    configHash,
		SourceSnapshotRef:    dummyRef,
		DetectionRef:         dummyRef,
		ScopeRef:             dummyRef,
		ValidityGraphRef:     dummyRef,
		MigrationFeasibility: "VIABLE_WITH_CONDITIONS",
		Findings: []AssessmentFindingRecord{
			{
				FindingID:         "find_ci3_extracted",
				AnalysisRecordRef: dummyRef,
				Classification:    "INFERRED",
				Basis:             "STATIC_ONLY",
				Category:          "CHARACTERIZATION",
				Summary:           "CI3 static characterization completed successfully",
				Reason:            "2 entrypoints and 3 controllers extracted",
				ProvenanceRefs:    []string{dummyRef},
				PrerequisiteRefs:  []string{dummyRef},
				Limitations:       []string{"Dynamic routes require runtime verification"},
				AffectedIDs:       []string{"proj-ci3-ref"},
			},
		},
		AsOfSequence: 1,
	})
	if err != nil {
		t.Fatalf("CompileAssessmentReport failed: %v", err)
	}
	if assessDoc.AssessmentStatus != "COMPLETED" {
		t.Errorf("expected COMPLETED assessment, got %s", assessDoc.AssessmentStatus)
	}
	if assessDoc.ImplementationAuthorized {
		t.Errorf("ARCHITECTURAL INVARIANT: implementation_authorized must remain false")
	}
}
