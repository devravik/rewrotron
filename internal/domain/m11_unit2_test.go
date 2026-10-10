package domain

import (
	"context"
	"strings"
	"testing"
)

// TestM11_Unit2_LaravelTargetAdapter verifies that the LaravelTargetAdapter conforms to
// TargetAdapter port, registers cleanly in TargetAdapterRegistry, resolves by environment,
// validates capabilities, and produces deterministic standalone Laravel 13 project artifacts.
func TestM11_Unit2_LaravelTargetAdapter(t *testing.T) {
	ctx := context.Background()

	adapter := NewLaravelTargetAdapter()

	// 1. Manifest verification
	m := adapter.Manifest()
	if m.AdapterID != "target-php-laravel13" || m.Language != "php" || m.Framework == nil || *m.Framework != "laravel" {
		t.Fatalf("unexpected manifest for Laravel target adapter: %+v", m)
	}

	// 2. Registry binding & resolution
	reg := NewTargetAdapterRegistry()
	if err := reg.Register(adapter); err != nil {
		t.Fatalf("failed registering Laravel target adapter: %v", err)
	}

	phpEnv := &ResolvedEnvironmentDocument{
		Language: "php",
		Framework: &ResolvedComponent{
			Name: "laravel",
		},
	}
	resolved, err := reg.ResolveForResolution(phpEnv)
	if err != nil {
		t.Fatalf("failed resolving Laravel target adapter: %v", err)
	}
	if resolved.Manifest().AdapterID != "target-php-laravel13" {
		t.Fatalf("expected target-php-laravel13, got: %s", resolved.Manifest().AdapterID)
	}

	// 3. Capability validation
	valRes, err := adapter.ValidateCapabilities(ctx, nil, []string{"routing", "controllers", "eloquent_models", "database"})
	if err != nil || !valRes.Valid || len(valRes.UnsupportedCaps) != 0 {
		t.Fatalf("expected valid capabilities, got: %+v, err: %v", valRes, err)
	}

	badRes, err := adapter.ValidateCapabilities(ctx, nil, []string{"routing", "unsupported_distributed_raft"})
	if err != nil {
		t.Fatalf("unexpected error validating capabilities: %v", err)
	}
	if badRes.Valid || len(badRes.Blockers) != 1 {
		t.Fatalf("expected blocker for unsupported capability, got: %+v", badRes)
	}

	// 4. Harness specification
	harness, err := adapter.PrepareHarness(ctx, nil)
	if err != nil {
		t.Fatalf("prepare harness failed: %v", err)
	}
	if harness.AdapterID != "target-php-laravel13" || harness.Language != "php" || harness.Framework == nil || *harness.Framework != "laravel" {
		t.Fatalf("unexpected harness spec: %+v", harness)
	}
	if len(harness.BuildCommand) == 0 || len(harness.RunCommand) == 0 {
		t.Fatalf("empty build or run commands in harness: %+v", harness)
	}

	// 5. Standalone project generation determinism
	unitID := "auth-and-checkout"
	adm := makeTestAdmission(t, unitID, []string{"routing", "session", "auth", "database"}, []string{"login", "checkout"})
	tm := makeTestTargetModel(t, unitID)
	sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

	input := GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: sc,
	}

	out1, err := adapter.Generate(ctx, input)
	if err != nil {
		t.Fatalf("generate pass 1 failed: %v", err)
	}

	out2, err := adapter.Generate(ctx, input)
	if err != nil {
		t.Fatalf("generate pass 2 failed: %v", err)
	}

	if out1.Manifest.ProvenanceRef != out2.Manifest.ProvenanceRef {
		t.Fatalf("non-deterministic provenance: %s != %s", out1.Manifest.ProvenanceRef, out2.Manifest.ProvenanceRef)
	}

	// Verify required standalone files
	requiredFiles := []string{
		"composer.json",
		"bootstrap/app.php",
		"artisan",
		"public/index.php",
		"config/app.php",
		"config/database.php",
		"config/session.php",
		"routes/web.php",
		"routes/api.php",
		"app/Http/Controllers/ReconstructedController.php",
		"app/Http/Middleware/SessionContinuity.php",
		"app/Models/ReconstructedItem.php",
	}

	for _, rf := range requiredFiles {
		content, ok := out1.Artifacts[rf]
		if !ok || len(content) == 0 {
			t.Errorf("missing or empty required artifact: %s", rf)
		}
	}

	// Verify composer.json specifies Laravel 13 and PHP 8.3+
	composerStr := string(out1.Artifacts["composer.json"])
	if !strings.Contains(composerStr, `"laravel/framework": "^13.0"`) || !strings.Contains(composerStr, `"php": "^8.3"`) {
		t.Errorf("composer.json does not target PHP 8.3+ / Laravel 13: %s", composerStr)
	}

	// Verify no runtime engine dependencies in target artifacts
	for path, content := range out1.Artifacts {
		contentStr := string(content)
		if strings.Contains(contentStr, "github.com/devravik/rewrotron") {
			t.Errorf("artifact %s leaks internal engine import", path)
		}
	}
}
