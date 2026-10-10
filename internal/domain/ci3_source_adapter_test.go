package domain

import (
	"context"
	"strings"
	"testing"
)

func TestCodeIgniter3SourceAdapter_Detect(t *testing.T) {
	ctx := context.Background()
	adapter := NewCodeIgniter3SourceAdapter()

	// CI3 file structure
	files := map[string][]byte{
		"application/config/config.php": []byte("<?php $config['base_url'] = '';"),
		"application/config/routes.php": []byte("<?php $route['default_controller'] = 'welcome';"),
		"system/core/CodeIgniter.php":   []byte("<?php define('CI_VERSION', '3.1.13');"),
	}

	input := SourceDetectionInput{
		ProjectID:   "proj-ci3-detect",
		SnapshotRef: "sha256:" + strings.Repeat("a", 64),
		Files:       files,
	}

	res, err := adapter.Detect(ctx, input)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if res.Language != "php" || res.Framework == nil || *res.Framework != "codeigniter3" {
		t.Fatalf("expected php/codeigniter3, got: %+v", res)
	}
	if res.DetectedVersion == nil || *res.DetectedVersion != "3.1.13" {
		t.Errorf("expected version 3.1.13, got %v", res.DetectedVersion)
	}
	if res.Confidence < 0.95 {
		t.Errorf("expected high confidence >= 0.95, got %f", res.Confidence)
	}
}

func TestCodeIgniter3SourceAdapter_Extract_RoutesAndRemap(t *testing.T) {
	ctx := context.Background()
	adapter := NewCodeIgniter3SourceAdapter()

	files := map[string][]byte{
		"application/config/routes.php": []byte(`<?php
$route['default_controller'] = 'welcome';
$route['404_override'] = '';
$route['translate_uri_dashes'] = FALSE;
$route['api/users']['get'] = 'users/list';
$route['api/users']['post'] = 'users/create';
$route['catalog/(:num)'] = 'catalog/product/$1';
`),
		"application/controllers/Welcome.php": []byte(`<?php
class Welcome extends CI_Controller {
    public function index() {
        echo "welcome";
    }
}
`),
		"application/controllers/RemapCtrl.php": []byte(`<?php
class RemapCtrl extends CI_Controller {
    public function _remap($method, $params = array()) {
        // dynamic dispatch
        return $this->$method();
    }
}
`),
		"application/models/User_model.php": []byte(`<?php
class User_model extends CI_Model {
    public function get_user() {}
}
`),
	}

	input := SourceExtractionInput{
		ProjectID:   "proj-ci3-extract",
		SnapshotRef: "sha256:" + strings.Repeat("b", 64),
		Files:       files,
	}

	out, err := adapter.Extract(ctx, input)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	doc := out.Document
	if len(doc.Entrypoints) < 4 {
		t.Errorf("expected at least 4 entrypoints, got %d", len(doc.Entrypoints))
	}

	// Verify controllers and models extracted
	foundWelcome := false
	foundRemap := false
	foundModel := false
	for _, c := range doc.Components {
		if c.Role == ComponentRoleHandler && len(c.SymbolIDs) > 0 {
			if c.SymbolIDs[0] == "welcome" {
				foundWelcome = true
			}
			if c.SymbolIDs[0] == "remapctrl" {
				foundRemap = true
			}
		}
		if c.Role == ComponentRoleDataModel && len(c.SymbolIDs) > 0 && c.SymbolIDs[0] == "user_model" {
			foundModel = true
		}
	}

	if !foundWelcome || !foundRemap {
		t.Errorf("expected welcome and remapctrl controllers, got: %+v", doc.Components)
	}
	if !foundModel {
		t.Errorf("expected user_model data_model, got: %+v", doc.Components)
	}

	// Verify _remap dynamic dispatch produces finding and explicit gap
	foundRemapFinding := false
	for _, f := range doc.Findings {
		if f.Code == "CI3_REMAP_DYNAMIC_DISPATCH" {
			foundRemapFinding = true
			break
		}
	}
	if !foundRemapFinding {
		t.Errorf("expected CI3_REMAP_DYNAMIC_DISPATCH finding")
	}
	if len(out.Gaps) == 0 {
		t.Errorf("expected explicit gaps reporting from CI3 adapter")
	}

	// Validate against canonical Application IR rules
	cjson, err := doc.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if err := ValidateApplicationIR(cjson); err != nil {
		t.Fatalf("generated ApplicationIRDocument failed ValidateApplicationIR: %v", err)
	}
}

func TestCodeIgniter3SourceAdapter_InstrumentationSpec(t *testing.T) {
	ctx := context.Background()
	adapter := NewCodeIgniter3SourceAdapter()

	spec, err := adapter.PrepareInstrumentation(ctx, SourceExtractionInput{})
	if err != nil {
		t.Fatalf("PrepareInstrumentation failed: %v", err)
	}
	if !spec.DisposableCopy {
		t.Errorf("CI3 probe must require disposable copy")
	}
	if spec.HookConfig["post_controller_constructor"] == "" {
		t.Errorf("expected post_controller_constructor hook")
	}
}
