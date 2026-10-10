package domain

import (
	"context"
	"strings"
	"testing"
)

// TestM11_Unit3_VerticalSliceGeneration verifies that the end-to-end vertical slice generated
// by LaravelTargetAdapter accurately reconstructs:
// 1. Read-only health and whoami inspection
// 2. Authentication and password verification against User model
// 3. Stateful checkout mutation (Create) with database transaction wrapping Item and Job insertions
// 4. Session continuity, flash messages, and CSRF protection token validation
// 5. Relational database mapping preserving foreign keys and auto-increment identity
// 6. Sealed GenerationManifest provenance and entries
func TestM11_Unit3_VerticalSliceGeneration(t *testing.T) {
	ctx := context.Background()

	unitID := "vertical-slice-auth-checkout"
	adm := makeTestAdmission(t, unitID, []string{
		"routing",
		"controllers",
		"eloquent_models",
		"database",
		"transactions",
		"auth",
		"session",
		"csrf",
	}, []string{
		"health",
		"login",
		"whoami",
		"csrf",
		"create",
		"item",
		"note",
		"logout",
	})
	tm := makeTestTargetModel(t, unitID)
	sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

	adapter := NewLaravelTargetAdapter()

	out, err := adapter.Generate(ctx, GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: sc,
	})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// 1. Verify routes/web.php contains all 8 vertical slice endpoints
	routesBytes, ok := out.Artifacts["routes/web.php"]
	if !ok {
		t.Fatalf("missing routes/web.php")
	}
	rStr := string(routesBytes)
	for _, ep := range []string{"health", "login", "whoami", "csrf", "create", "item", "note", "logout"} {
		if !strings.Contains(rStr, "'/"+ep+"'") {
			t.Errorf("routes/web.php missing endpoint /%s", ep)
		}
	}

	// 2. Verify ReconstructedController contains all handlers and transactional boundaries
	ctrlBytes, ok := out.Artifacts["app/Http/Controllers/ReconstructedController.php"]
	if !ok {
		t.Fatalf("missing ReconstructedController.php")
	}
	cStr := string(ctrlBytes)

	// Check transactional database mutation in Create
	if !strings.Contains(cStr, "DB::transaction(function ()") {
		t.Errorf("ReconstructedController missing DB::transaction in Create method")
	}
	if !strings.Contains(cStr, "Item::create(") || !strings.Contains(cStr, "Job::create(") {
		t.Errorf("ReconstructedController missing Item and Job model creations in Create method")
	}

	// Check password verification in Login
	if !strings.Contains(cStr, "password_verify($password, $user->password)") {
		t.Errorf("ReconstructedController missing password_verify in Login method")
	}

	// Check CSRF error response
	if !strings.Contains(cStr, "response()->json(['error' => 'csrf'], 403)") {
		t.Errorf("ReconstructedController missing CSRF 403 error response")
	}

	// Check unauthenticated 401 error response
	if !strings.Contains(cStr, "response()->json(['error' => 'unauthenticated'], 401)") {
		t.Errorf("ReconstructedController missing unauthenticated 401 response")
	}

	// Check item ownership verification (403 for mismatch)
	if !strings.Contains(cStr, "response()->json(['error' => 'owner'], 403)") {
		t.Errorf("ReconstructedController missing item owner 403 check")
	}

	// 3. Verify Eloquent models
	userBytes, ok := out.Artifacts["app/Models/User.php"]
	if !ok || !strings.Contains(string(userBytes), "protected $table = 'users';") {
		t.Errorf("missing or invalid User Eloquent model")
	}

	itemBytes, ok := out.Artifacts["app/Models/Item.php"]
	if !ok || !strings.Contains(string(itemBytes), "protected $table = 'items';") || !strings.Contains(string(itemBytes), "belongsTo(User::class") {
		t.Errorf("missing or invalid Item Eloquent model")
	}

	jobBytes, ok := out.Artifacts["app/Models/Job.php"]
	if !ok || !strings.Contains(string(jobBytes), "protected $table = 'jobs';") {
		t.Errorf("missing or invalid Job Eloquent model")
	}

	// 4. Verify Session Continuity Middleware
	mwBytes, ok := out.Artifacts["app/Http/Middleware/SessionContinuity.php"]
	if !ok || !strings.Contains(string(mwBytes), "X-Session-Continuity") || !strings.Contains(string(mwBytes), "SAFE_BRIDGE") {
		t.Errorf("missing or invalid SessionContinuity middleware")
	}

	// 5. Verify GenerationManifest entries and provenance
	if out.Manifest.UnitID != unitID {
		t.Errorf("manifest unit ID mismatch: expected %s, got %s", unitID, out.Manifest.UnitID)
	}
	if len(out.Manifest.Entries) != len(out.Artifacts) {
		t.Errorf("manifest entries count mismatch: %d != %d", len(out.Manifest.Entries), len(out.Artifacts))
	}
	if out.Manifest.ProvenanceRef == "" || !strings.HasPrefix(out.Manifest.ProvenanceRef, "sha256:") {
		t.Errorf("invalid manifest provenance ref: %s", out.Manifest.ProvenanceRef)
	}
}
