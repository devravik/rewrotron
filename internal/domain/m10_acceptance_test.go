package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestM10_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M10:
// Additional Source Adapters, verifying:
// 1. Pluggable source adapter contract & registry binding (Universal model boundaries, exact resolution)
// 2. Multi-source detection, capability validation, and deterministic extraction (Generic PHP & CodeIgniter 3)
// 3. Conservative classification and falsification controls (syntax statuses, dynamic findings, unevidenced assumptions rejection)
// 4. Source change selective invalidation (source file mutations change checksum and invalidate prior assessment conclusions)
// 5. Cross-milestone regression protection ensuring M0–M9 contracts remain intact
func TestM10_AcceptanceClosure(t *testing.T) {
	ctx := context.Background()

	// =========================================================================
	// 1. Source Adapter Registry & Exact Resolution
	// =========================================================================
	t.Run("SourceAdapterRegistry_ExactResolution", func(t *testing.T) {
		reg := NewSourceAdapterRegistry()
		phpGeneric := NewGenericPHPSourceAdapter()
		ci3Adapter := NewCodeIgniter3SourceAdapter()

		if err := reg.Register(phpGeneric); err != nil {
			t.Fatalf("register phpGeneric failed: %v", err)
		}
		if err := reg.Register(ci3Adapter); err != nil {
			t.Fatalf("register ci3Adapter failed: %v", err)
		}

		// Duplicate registration rejection
		if err := reg.Register(phpGeneric); err == nil {
			t.Fatalf("expected error on duplicate adapter registration, got nil")
		}

		// Exact resolution for generic PHP (framework = nil)
		resGeneric, err := reg.ResolveForEnvironment("php", nil)
		if err != nil {
			t.Fatalf("expected generic PHP adapter, got err: %v", err)
		}
		if resGeneric.Manifest().AdapterID != "source.php.generic" || resGeneric.Manifest().Framework != nil {
			t.Fatalf("unexpected adapter resolved for generic PHP: %+v", resGeneric.Manifest())
		}

		// Exact resolution for CodeIgniter 3
		ci3Fw := "codeigniter3"
		resCI3, err := reg.ResolveForEnvironment("php", &ci3Fw)
		if err != nil {
			t.Fatalf("expected CI3 adapter, got err: %v", err)
		}
		if resCI3.Manifest().AdapterID != "source.php.ci3" || resCI3.Manifest().Framework == nil || *resCI3.Manifest().Framework != "codeigniter3" {
			t.Fatalf("unexpected adapter resolved for CI3: %+v", resCI3.Manifest())
		}

		// Unregistered framework returns typed blocker error
		unregFw := "symfony"
		_, err = reg.ResolveForEnvironment("php", &unregFw)
		if err == nil || !strings.Contains(err.Error(), "no qualified source adapter registered") {
			t.Fatalf("expected error for unregistered framework, got: %v", err)
		}
	})

	// =========================================================================
	// 2. Capability Validation & Blocker Reporting
	// =========================================================================
	t.Run("SourceAdapter_CapabilityValidation", func(t *testing.T) {
		genericAdapter := NewGenericPHPSourceAdapter()
		ci3Adapter := NewCodeIgniter3SourceAdapter()

		// Generic PHP supports syntax parsing and symbol extraction
		resGen, err := genericAdapter.ValidateCapabilities(ctx, []string{
			string(CapSourceSyntaxParsing),
			string(CapSourceSymbolExtraction),
		})
		if err != nil || !resGen.Valid || len(resGen.UnsupportedCaps) != 0 {
			t.Fatalf("expected valid capabilities for generic PHP, got: %+v, err: %v", resGen, err)
		}

		// Generic PHP does NOT support framework-level route or database extraction
		resGenBad, err := genericAdapter.ValidateCapabilities(ctx, []string{
			string(CapSourceRouteExtraction),
			string(CapSourceDatabaseAccess),
		})
		if err != nil {
			t.Fatalf("unexpected error validating generic PHP caps: %v", err)
		}
		if resGenBad.Valid || len(resGenBad.UnsupportedCaps) != 2 || len(resGenBad.Blockers) != 2 {
			t.Fatalf("expected invalid caps with blockers for generic PHP, got: %+v", resGenBad)
		}

		// CI3 supports route, controller, database, session, and middleware extraction
		resCI3, err := ci3Adapter.ValidateCapabilities(ctx, []string{
			string(CapSourceRouteExtraction),
			string(CapSourceControllerExtraction),
			string(CapSourceDatabaseAccess),
			string(CapSourceSessionManagement),
			string(CapSourceMiddlewareLifecycle),
			string(CapSourceProbeInstrumentation),
		})
		if err != nil || !resCI3.Valid || len(resCI3.UnsupportedCaps) != 0 {
			t.Fatalf("expected valid capabilities for CI3, got: %+v, err: %v", resCI3, err)
		}

		// CI3 does NOT support arbitrary unsupported capability
		resCI3Bad, err := ci3Adapter.ValidateCapabilities(ctx, []string{"unsupported_ai_synthesis"})
		if err != nil {
			t.Fatalf("unexpected error validating CI3 caps: %v", err)
		}
		if resCI3Bad.Valid || len(resCI3Bad.Blockers) != 1 {
			t.Fatalf("expected blocker for unsupported CI3 capability, got: %+v", resCI3Bad)
		}
	})

	// =========================================================================
	// 3. Multi-Source Detection & Deterministic Extraction
	// =========================================================================
	t.Run("MultiSourceDetectionAndExtractionDeterminism", func(t *testing.T) {
		ci3Files := map[string][]byte{
			"index.php": []byte(`<?php
define('ENVIRONMENT', 'production');
$system_path = 'system';
$application_folder = 'application';
require_once BASEPATH.'core/CodeIgniter.php';
`),
			"system/core/CodeIgniter.php": []byte(`<?php
define('CI_VERSION', '3.1.13');
class CI_Controller {}
class CI_Model {}
`),
			"application/config/routes.php": []byte(`<?php
$route['default_controller'] = 'welcome';
$route['404_override'] = 'errors/page_missing';
$route['translate_uri_dashes'] = FALSE;
$route['api/v1/checkout']['post'] = 'checkout/process';
$route['api/v1/catalog'] = 'catalog/index';
`),
			"application/config/database.php": []byte(`<?php
$db['default']['dbdriver'] = 'mysqli';
$db['default']['database'] = 'shop_db';
`),
			"application/config/config.php": []byte(`<?php
$config['sess_driver'] = 'files';
$config['sess_cookie_name'] = 'ci_session';
$config['enable_hooks'] = TRUE;
`),
			"application/config/hooks.php": []byte(`<?php
$hook['pre_controller'] = array(
	'class'    => 'AuthHook',
	'function' => 'checkAuth',
	'filename' => 'AuthHook.php',
	'filepath' => 'hooks'
);
`),
			"application/controllers/Welcome.php": []byte(`<?php
class Welcome extends CI_Controller {
	public function index() {
		echo "Welcome to CI3";
	}
}
`),
			"application/models/OrderModel.php": []byte(`<?php
class OrderModel extends CI_Model {
	public function createOrder() {
		$this->db->trans_start();
		$this->db->insert('orders', ['total' => 100]);
		$this->db->trans_complete();
	}
}
`),
			"application/controllers/LegacyRemap.php": []byte(`<?php
class LegacyRemap extends CI_Controller {
	public function _remap($method, $params = array()) {
		call_user_func_array(array($this, $method), $params);
	}
}
`),
		}

		ci3Adapter := NewCodeIgniter3SourceAdapter()

		// 1. Detection
		detRes, err := ci3Adapter.Detect(ctx, SourceDetectionInput{
			ProjectID:   "proj-ci3-test",
			SnapshotRef: "snap-ci3-001",
			Files:       ci3Files,
		})
		if err != nil {
			t.Fatalf("CI3 detect failed: %v", err)
		}
		if detRes.Language != "php" || detRes.Framework == nil || *detRes.Framework != "codeigniter3" {
			t.Fatalf("unexpected detection: %+v", detRes)
		}
		if detRes.DetectedVersion == nil || *detRes.DetectedVersion != "3.1.13" {
			t.Fatalf("expected detected version 3.1.13, got: %v", detRes.DetectedVersion)
		}
		if detRes.Confidence < 0.95 {
			t.Fatalf("expected high confidence >= 0.95, got %f", detRes.Confidence)
		}

		// 2. Parse Pass 1 & Pass 2 (Determinism Verification)
		parseIn := SourceParseInput{
			ProjectID:   "proj-ci3-test",
			SnapshotRef: "snap-ci3-001",
			Files:       ci3Files,
		}
		pOut1, err := ci3Adapter.Parse(ctx, parseIn)
		if err != nil {
			t.Fatalf("parse pass 1 failed: %v", err)
		}
		pOut2, err := ci3Adapter.Parse(ctx, parseIn)
		if err != nil {
			t.Fatalf("parse pass 2 failed: %v", err)
		}

		// Compare JSON serialization of Source IR for exact determinism
		pBytes1, err := json.Marshal(pOut1.Document)
		if err != nil {
			t.Fatal(err)
		}
		pBytes2, err := json.Marshal(pOut2.Document)
		if err != nil {
			t.Fatal(err)
		}
		if string(pBytes1) != string(pBytes2) {
			t.Fatalf("non-deterministic Source IR document between passes")
		}

		// Verify Source IR structure: must contain files and symbols
		if len(pOut1.Document.Files) == 0 {
			t.Fatalf("expected parsed files in Source IR, got 0")
		}

		// 3. Extract Pass 1 & Pass 2 (Application IR Determinism)
		extractIn := SourceExtractionInput{
			ProjectID:   "proj-ci3-test",
			SnapshotRef: "snap-ci3-001",
			SourceIR:    pOut1.Document,
			Files:       ci3Files,
		}
		eOut1, err := ci3Adapter.Extract(ctx, extractIn)
		if err != nil {
			t.Fatalf("extract pass 1 failed: %v", err)
		}
		eOut2, err := ci3Adapter.Extract(ctx, extractIn)
		if err != nil {
			t.Fatalf("extract pass 2 failed: %v", err)
		}

		eBytes1, err := json.Marshal(eOut1.Document)
		if err != nil {
			t.Fatal(err)
		}
		eBytes2, err := json.Marshal(eOut2.Document)
		if err != nil {
			t.Fatal(err)
		}
		if string(eBytes1) != string(eBytes2) {
			t.Fatalf("non-deterministic Application IR document between passes")
		}

		// Verify non-null array invariant on Application IR
		var rawAppIR map[string]interface{}
		if err := json.Unmarshal(eBytes1, &rawAppIR); err != nil {
			t.Fatalf("failed unmarshaling AppIR JSON: %v", err)
		}
		for _, arrayField := range []string{"entrypoints", "components", "dependencies", "data_models", "workflows", "findings", "evidence"} {
			v, exists := rawAppIR[arrayField]
			if !exists || v == nil {
				t.Fatalf("ApplicationIR invariant violation: field %q must serialize as JSON array, got: %v", arrayField, v)
			}
		}

		// Check entrypoints extracted
		if len(eOut1.Document.Entrypoints) < 4 {
			t.Fatalf("expected at least 4 entrypoints (default, 404, checkout, catalog), got: %d", len(eOut1.Document.Entrypoints))
		}

		// Check findings: LegacyRemap has _remap dynamic dispatch -> CI3_REMAP_DYNAMIC_DISPATCH finding
		foundRemapFinding := false
		for _, f := range eOut1.Document.Findings {
			if f.Code == "CI3_REMAP_DYNAMIC_DISPATCH" {
				foundRemapFinding = true
				if f.Severity != FindingSeverityHigh {
					t.Fatalf("expected HIGH severity for _remap dynamic dispatch, got: %s", f.Severity)
				}
				if !strings.Contains(f.Message, "UNKNOWN") {
					t.Fatalf("expected UNKNOWN coverage reference in _remap message, got: %s", f.Message)
				}
			}
		}
		if !foundRemapFinding {
			t.Fatalf("expected CI3_REMAP_DYNAMIC_DISPATCH finding in extracted AppIR")
		}

		// Check database transaction boundary finding
		foundDBTrans := false
		for _, f := range eOut1.Document.Findings {
			if f.Code == "CI3_DB_TRANSACTION_BOUNDARY" {
				foundDBTrans = true
			}
		}
		if !foundDBTrans {
			t.Fatalf("expected CI3_DB_TRANSACTION_BOUNDARY finding for OrderModel")
		}

		// 4. Instrumentation Specification Verification
		spec, err := ci3Adapter.PrepareInstrumentation(ctx, extractIn)
		if err != nil {
			t.Fatalf("prepare instrumentation failed: %v", err)
		}
		if spec.AdapterID != "source.php.ci3" || !spec.DisposableCopy {
			t.Fatalf("instrumentation spec must enforce disposable copy: %+v", spec)
		}
	})

	// =========================================================================
	// 4. Conservative Classification & Falsification Controls
	// =========================================================================
	t.Run("ConservativeClassificationAndFalsificationControls", func(t *testing.T) {
		phpAdapter := NewGenericPHPSourceAdapter()

		// Case A: Clean PHP file -> PARSED
		cleanFiles := map[string][]byte{
			"UserService.php": []byte(`<?php
class UserService {
	public function findUserById($id) {
		return "user_" . $id;
	}
}
`),
		}
		outClean, err := phpAdapter.Parse(ctx, SourceParseInput{
			ProjectID:   "p-clean",
			SnapshotRef: "s-clean",
			Files:       cleanFiles,
		})
		if err != nil {
			t.Fatalf("parse clean failed: %v", err)
		}
		if outClean.Statuses["UserService.php"] != SyntaxStatusParsed {
			t.Fatalf("expected PARSED for clean file, got: %v", outClean.Statuses["UserService.php"])
		}

		// Case B: Dynamic eval / variable variables -> PARTIAL + findings with UNKNOWN impact
		dynamicFiles := map[string][]byte{
			"DynamicDispatcher.php": []byte(`<?php
class DynamicDispatcher {
	public function dispatch($action, $data) {
		eval("\$result = \$this->" . $action . "(\$data);");
		return $result;
	}
}
`),
		}
		outDyn, err := phpAdapter.Parse(ctx, SourceParseInput{
			ProjectID:   "p-dyn",
			SnapshotRef: "s-dyn",
			Files:       dynamicFiles,
		})
		if err != nil {
			t.Fatalf("parse dynamic failed: %v", err)
		}
		if outDyn.Statuses["DynamicDispatcher.php"] != SyntaxStatusPartial {
			t.Fatalf("expected PARTIAL for dynamic file, got: %v", outDyn.Statuses["DynamicDispatcher.php"])
		}
		if len(outDyn.Document.Findings) == 0 {
			t.Fatalf("expected dynamic dispatch findings, got 0")
		}
		if outDyn.Document.Findings[0].Code != "DYNAMIC_DISPATCH_UNKNOWN" {
			t.Fatalf("expected DYNAMIC_DISPATCH_UNKNOWN code, got: %s", outDyn.Document.Findings[0].Code)
		}

		// Case C: Non-PHP / malformed syntax -> FAILED / UNSUPPORTED
		malformedFiles := map[string][]byte{
			"Broken.php": []byte(`THIS IS NOT PHP AND HAS NO TAGS AT ALL`),
		}
		outBad, err := phpAdapter.Parse(ctx, SourceParseInput{
			ProjectID:   "p-bad",
			SnapshotRef: "s-bad",
			Files:       malformedFiles,
		})
		if err != nil {
			t.Fatalf("parse malformed failed: %v", err)
		}
		if outBad.Statuses["Broken.php"] != SyntaxStatusFailed {
			t.Fatalf("expected FAILED syntax status for malformed file, got: %v", outBad.Statuses["Broken.php"])
		}

		// Case D: CodeIgniter 3 unevidenced assumption rejection
		// If application/config/database.php has an unknown driver (e.g. 'oracle_custom'), it must produce an explicit gap
		ci3UnsupportedDB := map[string][]byte{
			"index.php":                     []byte(`<?php define('ENVIRONMENT', 'test');`),
			"application/config/routes.php": []byte(`<?php $route['default_controller'] = 'test';`),
			"application/config/database.php": []byte(`<?php
$db['default']['dbdriver'] = 'oracle_custom';
`),
		}
		ci3Adapter := NewCodeIgniter3SourceAdapter()
		parseRes, err := ci3Adapter.Parse(ctx, SourceParseInput{
			ProjectID:   "p-gap",
			SnapshotRef: "s-gap",
			Files:       ci3UnsupportedDB,
		})
		if err != nil {
			t.Fatalf("parse ci3 db failed: %v", err)
		}
		extractRes, err := ci3Adapter.Extract(ctx, SourceExtractionInput{
			ProjectID:   "p-gap",
			SnapshotRef: "s-gap",
			SourceIR:    parseRes.Document,
			Files:       ci3UnsupportedDB,
		})
		if err != nil {
			t.Fatalf("extract ci3 db failed: %v", err)
		}
		foundGap := false
		for _, g := range extractRes.Gaps {
			if strings.Contains(g, "unsupported database driver: oracle_custom") {
				foundGap = true
			}
		}
		if !foundGap {
			t.Fatalf("expected explicit gap for unsupported database driver, got gaps: %v", extractRes.Gaps)
		}
	})

	// =========================================================================
	// 5. Selective Invalidation on Source Change
	// =========================================================================
	t.Run("SelectiveInvalidationOnSourceChange", func(t *testing.T) {
		ci3Adapter := NewCodeIgniter3SourceAdapter()

		fileV1 := []byte(`<?php
class ProductModel extends CI_Model {
	public function listItems() {
		return "v1";
	}
}
`)
		fileV2 := []byte(`<?php
class ProductModel extends CI_Model {
	public function listItems() {
		return "v2 modified";
	}
	public function create() {
		$this->db->trans_start();
		$this->db->insert('products', ['name' => 'item']);
		$this->db->trans_complete();
	}
}
`)

		// Snapshot 1 hash
		h1 := sha256.Sum256(fileV1)
		hashV1 := hex.EncodeToString(h1[:])

		// Snapshot 2 hash
		h2 := sha256.Sum256(fileV2)
		hashV2 := hex.EncodeToString(h2[:])

		if hashV1 == hashV2 {
			t.Fatalf("hashes must differ between V1 and V2")
		}

		p1, err := ci3Adapter.Parse(ctx, SourceParseInput{
			ProjectID:   "proj-inv",
			SnapshotRef: hashV1,
			Files: map[string][]byte{
				"application/models/ProductModel.php": fileV1,
			},
		})
		if err != nil {
			t.Fatalf("parse v1 failed: %v", err)
		}

		p2, err := ci3Adapter.Parse(ctx, SourceParseInput{
			ProjectID:   "proj-inv",
			SnapshotRef: hashV2,
			Files: map[string][]byte{
				"application/models/ProductModel.php": fileV2,
			},
		})
		if err != nil {
			t.Fatalf("parse v2 failed: %v", err)
		}

		// Verify SourceIR documents have distinct source file hashes
		f1 := p1.Document.Files[0]
		f2 := p2.Document.Files[0]
		if f1.SHA256 == f2.SHA256 {
			t.Fatalf("expected distinct file hashes in Source IR, got %s == %s", f1.SHA256, f2.SHA256)
		}

		// Application IR extraction demonstrates that V2 introduces transaction findings not present in V1
		e1, err := ci3Adapter.Extract(ctx, SourceExtractionInput{
			ProjectID:   "proj-inv",
			SnapshotRef: hashV1,
			SourceIR:    p1.Document,
			Files: map[string][]byte{
				"application/models/ProductModel.php": fileV1,
			},
		})
		if err != nil {
			t.Fatalf("extract v1 failed: %v", err)
		}

		e2, err := ci3Adapter.Extract(ctx, SourceExtractionInput{
			ProjectID:   "proj-inv",
			SnapshotRef: hashV2,
			SourceIR:    p2.Document,
			Files: map[string][]byte{
				"application/models/ProductModel.php": fileV2,
			},
		})
		if err != nil {
			t.Fatalf("extract v2 failed: %v", err)
		}

		if len(e1.Document.Findings) != 0 {
			t.Fatalf("expected 0 findings in V1, got %d", len(e1.Document.Findings))
		}
		if len(e2.Document.Findings) != 1 || e2.Document.Findings[0].Code != "CI3_DB_TRANSACTION_BOUNDARY" {
			t.Fatalf("expected 1 CI3_DB_TRANSACTION_BOUNDARY finding in V2, got: %+v", e2.Document.Findings)
		}
	})

	// =========================================================================
	// 6. Cross-Milestone Regression Guardrails (M0–M9 Invariant Preservation)
	// =========================================================================
	t.Run("CrossMilestoneRegressionProtection", func(t *testing.T) {
		// M4 Unit 1: Resource Fencing Token Invariant - stale token must be rejected
		svc := NewResourceFencingService()
		now := time.Now()
		m := &MissionRecord{
			MissionID:   "m_m10_01",
			ProjectID:   "proj_m10",
			Type:        "MIGRATE_UNIT",
			Status:      MissionStatusQueued,
			Fingerprint: "fp_m10",
			Description: "M10 test",
		}
		_, err := svc.SubmitMission(m)
		if err != nil {
			t.Fatalf("submit mission failed: %v", err)
		}
		claimed, err := svc.ClaimMission("proj_m10", "m_m10_01", "worker_1", now, 5*time.Minute)
		if err != nil {
			t.Fatalf("claim mission failed: %v", err)
		}
		_, err = svc.AcquireResourceClaim("proj_m10", "app/model.go", "m_m10_01", "worker_1", claimed.FencingToken, now, 5*time.Minute)
		if err != nil {
			t.Fatalf("acquire claim failed: %v", err)
		}

		cleanHash := strings.Repeat("d", 64)
		staleTokenPatch := &PatchRecord{
			SchemaVersion: "1.0",
			ProjectID:     "proj_m10",
			UnitID:        "unit_m10",
			FencingToken:  claimed.FencingToken - 1, // stale token
			Files: []PatchFileRecord{
				{
					Path:         "app/model.go",
					Action:       PatchActionModify,
					BeforeSHA256: &cleanHash,
					AfterSHA256:  &cleanHash,
				},
			},
		}
		err = svc.ValidatePublication(staleTokenPatch, map[string][]byte{"app/model.go": []byte("initial")}, now)
		if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) {
			t.Fatalf("stale token must be rejected: %v", err)
		}

		// M8: Execution Profiler anomaly detection preserved
		profiler := NewExecutionProfiler()
		rc := profiler.StartRequest(ctx, "/api/products", "GET")
		for i := 0; i < 4; i++ {
			rc.RecordQuery("SELECT * FROM products WHERE cat_id = 1", 1*time.Millisecond, 1, nil)
		}
		p := profiler.FinishRequest(ctx, rc, 200, 256, nil)
		anomalies := p.DetectQueryAnomalies(3)
		if len(anomalies) != 1 || anomalies[0].PatternType != PatternNPlusOneQuery {
			t.Fatalf("expected N+1 anomaly preserved from M8, got: %v", anomalies)
		}

		// M9: Target Adapter Exact Resolution preserved
		targetReg := NewTargetAdapterRegistry()
		if err := targetReg.Register(NewGoStdlibTargetAdapter()); err != nil {
			t.Fatalf("register go target adapter failed: %v", err)
		}
		targetAdapter, err := targetReg.ResolveForResolution(&ResolvedEnvironmentDocument{
			Language:  "go",
			Framework: nil,
		})
		if err != nil || targetAdapter.Manifest().AdapterID != "target-go-stdlib" {
			t.Fatalf("target adapter exact resolution failed: %v", err)
		}
	})
}
