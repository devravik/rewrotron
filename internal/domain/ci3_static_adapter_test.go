package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestExtractCI3_CompleteRouteAndControllerDiscovery(t *testing.T) {
	routesContent := []byte(`<?php
defined('BASEPATH') OR exit('No direct script access allowed');

$route['default_controller'] = 'welcome';
$route['404_override'] = '';
$route['translate_uri_dashes'] = FALSE;

// Exact route
$route['login'] = 'auth/login';

// Verb-specific routes
$route['api/orders']['get'] = 'orders/list_orders';
$route['api/orders']['post'] = 'orders/create_order';

// Wildcard / regex routes
$route['users/(:num)'] = 'users/profile/$1';
$route['catalog/(:any)'] = 'catalog/view/$1';
`)

	welcomeCtrlContent := []byte(`<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class Welcome extends CI_Controller {
    public function index() {
        $this->load->view('welcome_message');
    }

    public function about() {
        $this->load->view('about');
    }

    private function _internal() {
        // should not be exposed as entrypoint
    }
}
`)

	remapCtrlContent := []byte(`<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class CustomRemap extends CI_Controller {
    public function _remap($method, $params = array()) {
        if ($method === 'special') {
            $this->special();
        }
    }

    public function special() {
        echo "special";
    }
}
`)

	files := map[string][]byte{
		"application/config/routes.php":           routesContent,
		"application/controllers/Welcome.php":     welcomeCtrlContent,
		"application/controllers/CustomRemap.php": remapCtrlContent,
	}

	input := &CI3AdapterInput{
		ProjectID:         "test-ci3-app",
		ConfigurationHash: strings.Repeat("a", 64),
		DocumentID:        "doc-ci3-extract",
		Files:             files,
		OperationalRoots:  []string{"migrate", "cron:daily_cleanup"},
	}

	res, err := ExtractCI3(input)
	if err != nil {
		t.Fatalf("expected ExtractCI3 to succeed, got: %v", err)
	}

	if res.SourceIR == nil || res.ApplicationIR == nil {
		t.Fatal("expected non-nil SourceIR and ApplicationIR")
	}

	// 1. Verify Entrypoints
	var foundVerbGet, foundVerbPost, foundRemap, foundCLI, foundCron bool
	for _, ep := range res.ApplicationIR.Entrypoints {
		if strings.Contains(ep.EntrypointID, "get_api_orders") {
			foundVerbGet = true
			if ep.CoverageState != CoverageStateResolved {
				t.Errorf("expected verb get route to be RESOLVED, got %s", ep.CoverageState)
			}
		}
		if strings.Contains(ep.EntrypointID, "post_api_orders") {
			foundVerbPost = true
			if ep.CoverageState != CoverageStateResolved {
				t.Errorf("expected verb post route to be RESOLVED, got %s", ep.CoverageState)
			}
		}
		if strings.Contains(ep.EntrypointID, "customremap") {
			foundRemap = true
			if ep.CoverageState != CoverageStatePartial {
				t.Errorf("expected custom _remap route to be PARTIAL, got %s", ep.CoverageState)
			}
		}
		if ep.Type == EntrypointTypeCLI && strings.Contains(ep.EntrypointID, "migrate") {
			foundCLI = true
		}
		if ep.Type == EntrypointTypeCron && strings.Contains(ep.EntrypointID, "daily_cleanup") {
			foundCron = true
		}
	}

	if !foundVerbGet {
		t.Error("expected verb-specific GET route to be discovered")
	}
	if !foundVerbPost {
		t.Error("expected verb-specific POST route to be discovered")
	}
	if !foundRemap {
		t.Error("expected custom _remap controller route to be discovered")
	}
	if !foundCLI {
		t.Error("expected CLI operational root to be discovered")
	}
	if !foundCron {
		t.Error("expected Cron operational root to be discovered")
	}

	// 2. Verify _remap controllers list
	if len(res.RemapControllers) != 1 || res.RemapControllers[0] != "CustomRemap" {
		t.Errorf("expected [CustomRemap], got %v", res.RemapControllers)
	}

	// 3. Verify Source IR and Application IR validate canonically
	sCjson, err := res.SourceIR.CanonicalJSON()
	if err != nil {
		t.Fatalf("SourceIR CanonicalJSON failed: %v", err)
	}
	if err := ValidateSourceIR(sCjson); err != nil {
		t.Fatalf("SourceIR failed schema validation: %v", err)
	}

	aCjson, err := res.ApplicationIR.CanonicalJSON()
	if err != nil {
		t.Fatalf("ApplicationIR CanonicalJSON failed: %v", err)
	}
	if err := ValidateApplicationIR(aCjson); err != nil {
		t.Fatalf("ApplicationIR failed schema validation: %v", err)
	}
}

// TestExtractCI3_X09_DynamicDispatchAndCronRoots tests Test X09 invariants:
// Dynamic includes, dynamic method calls, variable variables, reflection, eval,
// and custom _remap must generate unresolved edges/findings and prevent false DEAD classifications.
func TestExtractCI3_X09_DynamicDispatchAndCronRoots(t *testing.T) {
	dynamicCode := []byte(`<?php
class DynamicCtrl extends CI_Controller {
    public function dynamic_action() {
        $method = $this->input->get('action');
        $this->$method(); // dynamic method dispatch

        $file = $this->input->get('file');
        include($file); // dynamic include

        $ref = new ReflectionClass('SecretService'); // reflection
        eval('$x = 1;'); // eval
    }
}
`)

	files := map[string][]byte{
		"application/controllers/DynamicCtrl.php": dynamicCode,
	}

	input := &CI3AdapterInput{
		ProjectID:         "test-dyn-app",
		ConfigurationHash: strings.Repeat("a", 64),
		DocumentID:        "doc-dyn-extract",
		Files:             files,
	}

	res, err := ExtractCI3(input)
	if err != nil {
		t.Fatalf("ExtractCI3 failed: %v", err)
	}

	if !res.DynamicDispatchDetected {
		t.Fatal("expected DynamicDispatchDetected to be true")
	}

	// Check that dynamic dispatch edges and findings exist
	var hasDynEdge, hasIncEdge bool
	for _, edge := range res.SourceIR.Edges {
		if edge.Kind == "dynamic_dispatch" && !edge.Resolved {
			hasDynEdge = true
		}
		if edge.Kind == "dynamic_include" && !edge.Resolved {
			hasIncEdge = true
		}
	}
	if !hasDynEdge {
		t.Error("expected unresolved dynamic_dispatch edge")
	}
	if !hasIncEdge {
		t.Error("expected unresolved dynamic_include edge")
	}

	var hasDynFinding, hasIncFinding, hasReflectFinding, hasEvalFinding bool
	for _, f := range res.SourceIR.Findings {
		switch f.Code {
		case "DYNAMIC_DISPATCH":
			hasDynFinding = true
		case "DYNAMIC_INCLUDE":
			hasIncFinding = true
		case "REFLECTION_USAGE":
			hasReflectFinding = true
		case "UNSUPPORTED_EVAL":
			hasEvalFinding = true
		}
	}
	if !hasDynFinding || !hasIncFinding || !hasReflectFinding || !hasEvalFinding {
		t.Errorf("missing dynamic findings: dyn=%v, inc=%v, reflect=%v, eval=%v",
			hasDynFinding, hasIncFinding, hasReflectFinding, hasEvalFinding)
	}

	// Feed extracted dynamic candidate into CodeClassification to test X09 invariant:
	// Dynamic candidates must NOT be classified as DEAD even if unobserved!
	graphRef := "sha256:" + strings.Repeat("1", 64)
	profRef := "sha256:" + strings.Repeat("2", 64)
	corpRef := "sha256:" + strings.Repeat("3", 64)
	snapRef := "sha256:" + strings.Repeat("4", 64)

	omissionProof := validOmissionProofForTest("sym_class_dynamicctrl", snapRef, profRef, graphRef)

	unresolvedMap := make(map[string][]string)
	for _, f := range res.SourceIR.Findings {
		for _, affID := range f.AffectedIDs {
			unresolvedMap[affID] = append(unresolvedMap[affID], f.Message)
		}
	}

	classReq := &ClassificationRequest{
		ProjectID:        "test-dyn-app",
		Candidates:       []string{"sym_class_dynamicctrl"},
		OperationalRoots: []string{"root_http"},
		Edges: []DirectedEdge{
			{FromID: "root_http", ToID: "sym_class_dynamicctrl", Resolved: false}, // unresolved edge
		},
		UnresolvedNodes: unresolvedMap,
		ClosedBoundary:  true,
		OmissionProof:   omissionProof,
		GraphRef:        graphRef,
		ProfileRef:      profRef,
		CorpusRef:       corpRef,
		SnapshotRef:     snapRef,
	}

	classReport, err := ClassifyCode(classReq)
	if err != nil {
		t.Fatalf("ClassifyCode failed: %v", err)
	}

	cOutcome := classReport.Classifications["sym_class_dynamicctrl"]
	if cOutcome.State == ClassificationDead {
		t.Fatal("X09 VIOLATION: dynamic candidate was classified as DEAD!")
	}
	if cOutcome.State != ClassificationUnknown {
		t.Fatalf("expected UNKNOWN for dynamic candidate, got %s", cOutcome.State)
	}
}

// TestExtractCI3_X10_ModifiedEnvironmentAndVendorPatch tests Test X10 invariants:
// Modified vendor files and conflicting environment settings must be detected as separate facts and gaps.
func TestExtractCI3_X10_ModifiedEnvironmentAndVendorPatch(t *testing.T) {
	canonicalVendorBytes := []byte("<?php // Pristine CodeIgniter Core 3.1.13")
	canonicalSum := sha256.Sum256(canonicalVendorBytes)
	canonicalHex := hex.EncodeToString(canonicalSum[:])

	patchedVendorBytes := []byte("<?php // Patched CodeIgniter Core with custom hacks")

	indexPHPContent := []byte(`<?php
define('ENVIRONMENT', 'development');
require_once 'system/core/CodeIgniter.php';
`)

	files := map[string][]byte{
		"system/core/CodeIgniter.php": patchedVendorBytes,
		"index.php":                   indexPHPContent,
	}

	knownHashes := map[string]string{
		"system/core/CodeIgniter.php": canonicalHex,
	}

	input := &CI3AdapterInput{
		ProjectID:           "test-vendor-app",
		ConfigurationHash:   strings.Repeat("a", 64),
		DocumentID:          "doc-vendor-extract",
		Files:               files,
		KnownVendorHashes:   knownHashes,
		DeclaredEnvironment: "production", // contradicts development in index.php
	}

	res, err := ExtractCI3(input)
	if err != nil {
		t.Fatalf("ExtractCI3 failed: %v", err)
	}

	// 1. Verify ModifiedVendorFiles
	if len(res.ModifiedVendorFiles) != 1 || res.ModifiedVendorFiles[0] != "system/core/CodeIgniter.php" {
		t.Errorf("expected [system/core/CodeIgniter.php] in ModifiedVendorFiles, got %v", res.ModifiedVendorFiles)
	}

	// 2. Verify Findings for MODIFIED_VENDOR and ENVIRONMENT_MISMATCH
	var foundModVendor, foundEnvMismatch bool
	for _, f := range res.SourceIR.Findings {
		if f.Code == "MODIFIED_VENDOR" {
			foundModVendor = true
			if f.Severity != FindingSeverityHigh {
				t.Errorf("expected MODIFIED_VENDOR to have HIGH severity, got %s", f.Severity)
			}
		}
		if f.Code == "ENVIRONMENT_MISMATCH" {
			foundEnvMismatch = true
			if f.Severity != FindingSeverityHigh {
				t.Errorf("expected ENVIRONMENT_MISMATCH to have HIGH severity, got %s", f.Severity)
			}
		}
	}

	if !foundModVendor {
		t.Error("expected MODIFIED_VENDOR finding to be recorded")
	}
	if !foundEnvMismatch {
		t.Error("expected ENVIRONMENT_MISMATCH finding to be recorded")
	}

	// 3. Positive control: when vendor hash matches, no MODIFIED_VENDOR finding
	inputPristine := &CI3AdapterInput{
		ProjectID:           "test-vendor-app",
		ConfigurationHash:   strings.Repeat("a", 64),
		DocumentID:          "doc-vendor-pristine",
		Files: map[string][]byte{
			"system/core/CodeIgniter.php": canonicalVendorBytes,
			"index.php":                   []byte("<?php define('ENVIRONMENT', 'production');"),
		},
		KnownVendorHashes:   knownHashes,
		DeclaredEnvironment: "production",
	}

	resPristine, err := ExtractCI3(inputPristine)
	if err != nil {
		t.Fatalf("ExtractCI3 pristine failed: %v", err)
	}
	if len(resPristine.ModifiedVendorFiles) != 0 {
		t.Errorf("expected 0 modified vendor files for pristine codebase, got %d", len(resPristine.ModifiedVendorFiles))
	}
	for _, f := range resPristine.SourceIR.Findings {
		if f.Code == "MODIFIED_VENDOR" || f.Code == "ENVIRONMENT_MISMATCH" {
			t.Errorf("unexpected finding in pristine codebase: %s", f.Code)
		}
	}
}
