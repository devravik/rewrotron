package domain

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestGoStdlibTargetAdapter_DeterministicGeneration(t *testing.T) {
	ctx := context.Background()
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session", "auth"}, []string{"route", "health"})
	tm := makeTestTargetModel(t, unitID)
	sc := makeTestSessionContract(t, "proj-test", "case-auth", "SAFE_BRIDGE")

	adapter := NewGoStdlibTargetAdapter()
	input := GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: sc,
	}

	out1, err := adapter.Generate(ctx, input)
	if err != nil {
		t.Fatalf("first generate failed: %v", err)
	}

	out2, err := adapter.Generate(ctx, input)
	if err != nil {
		t.Fatalf("second generate failed: %v", err)
	}

	// 1. Provenance digest determinism: identical inputs MUST produce identical hash
	if out1.Manifest.ProvenanceRef != out2.Manifest.ProvenanceRef {
		t.Errorf("non-deterministic provenance ref: %s vs %s", out1.Manifest.ProvenanceRef, out2.Manifest.ProvenanceRef)
	}

	// 2. Expected files created
	expectedPaths := []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/server/router.go",
		"internal/handlers/handler.go",
		"internal/middleware/session.go",
		"generated/provenance.json",
	}

	for _, p := range expectedPaths {
		b1, ok1 := out1.Artifacts[p]
		b2, ok2 := out2.Artifacts[p]
		if !ok1 || !ok2 {
			t.Errorf("missing artifact: %s", p)
			continue
		}
		if string(b1) != string(b2) {
			t.Errorf("artifact %s has non-deterministic bytes", p)
		}
	}

	// 3. Validate generation manifest
	raw1, err := json.Marshal(out1.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenerationManifest(raw1); err != nil {
		t.Fatalf("generated manifest validation failed: %v", err)
	}

	// 4. Harness spec verification
	harness, err := adapter.PrepareHarness(ctx, input.TargetModel)
	if err != nil {
		t.Fatalf("PrepareHarness failed: %v", err)
	}
	if harness.EntrypointPath != "cmd/server/main.go" {
		t.Errorf("expected entrypoint cmd/server/main.go, got %s", harness.EntrypointPath)
	}
	if harness.Language != "go" {
		t.Errorf("expected language go, got %s", harness.Language)
	}
	if harness.Framework != nil {
		t.Errorf("expected framework nil for stdlib, got %v", harness.Framework)
	}
}

func TestGoStdlibTargetAdapter_MissingSessionContract_Fails(t *testing.T) {
	ctx := context.Background()
	unitID := "sync-auth"
	// Admission includes session capability
	adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session"}, []string{"route", "health"})
	tm := makeTestTargetModel(t, unitID)

	adapter := NewGoStdlibTargetAdapter()
	input := GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: nil, // Missing session contract
	}

	_, err := adapter.Generate(ctx, input)
	if err == nil {
		t.Fatalf("expected error when session capability is in scope without session contract, got nil")
	}
}

func TestGoStdlibTargetAdapter_RoutingAndLifecycleLowering(t *testing.T) {
	ctx := context.Background()
	unitID := "sync-checkout"
	adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session", "auth"}, []string{"login", "checkout-pay"})
	tm := makeTestTargetModel(t, unitID)
	sc := makeTestSessionContract(t, "proj-test", "case-auth", "SAFE_BRIDGE")

	adapter := NewGoStdlibTargetAdapter()
	input := GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: sc,
	}

	out, err := adapter.Generate(ctx, input)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	routerSource := string(out.Artifacts["internal/server/router.go"])
	if !strings.Contains(routerSource, `mux.HandleFunc("/login", h.LoginHandler)`) {
		t.Errorf("routerSource missing /login route: %s", routerSource)
	}
	if !strings.Contains(routerSource, `mux.HandleFunc("/checkout-pay", h.CheckoutPayHandler)`) {
		t.Errorf("routerSource missing /checkout-pay route: %s", routerSource)
	}
	if !strings.Contains(routerSource, "middleware.SessionContinuity(mux)") {
		t.Errorf("routerSource missing SessionContinuity wrapping: %s", routerSource)
	}

	handlerSource := string(out.Artifacts["internal/handlers/handler.go"])
	if !strings.Contains(handlerSource, "func (h *Handler) LoginHandler(") {
		t.Errorf("handlerSource missing LoginHandler: %s", handlerSource)
	}
	if !strings.Contains(handlerSource, "func (h *Handler) CheckoutPayHandler(") {
		t.Errorf("handlerSource missing CheckoutPayHandler: %s", handlerSource)
	}

	middlewareSource := string(out.Artifacts["internal/middleware/session.go"])
	if !strings.Contains(middlewareSource, "Session continuity strategy: SAFE_BRIDGE") {
		t.Errorf("middlewareSource missing strategy: %s", middlewareSource)
	}
}
