package domain

import (
	"context"
	"strings"
	"testing"
)

// MockTargetAdapter implements TargetAdapter for testing registry and capability validation.
type MockTargetAdapter struct {
	manifest TargetAdapterManifest
}

func (m *MockTargetAdapter) Manifest() TargetAdapterManifest {
	return m.manifest
}

func (m *MockTargetAdapter) ValidateCapabilities(ctx context.Context, targetModel *TargetModelDocument, requiredCaps []string) (*CapabilityValidationResult, error) {
	supportedMap := make(map[string]bool)
	for _, c := range m.manifest.Capabilities {
		supportedMap[c] = true
	}

	res := &CapabilityValidationResult{
		Valid: true,
	}

	for _, req := range requiredCaps {
		if supportedMap[req] {
			res.SupportedCaps = append(res.SupportedCaps, req)
		} else {
			res.UnsupportedCaps = append(res.UnsupportedCaps, req)
			res.Blockers = append(res.Blockers, "unsupported capability: "+req)
			res.Valid = false
		}
	}
	return res, nil
}

func (m *MockTargetAdapter) Generate(ctx context.Context, input GenerationInput) (*GenerationOutput, error) {
	return nil, nil
}

func (m *MockTargetAdapter) PrepareHarness(ctx context.Context, targetModel *TargetModelDocument) (*TargetHarnessSpec, error) {
	return &TargetHarnessSpec{
		AdapterID:      m.manifest.AdapterID,
		Language:       m.manifest.Language,
		Framework:      m.manifest.Framework,
		EntrypointPath: "main.go",
	}, nil
}

func TestTargetAdapterRegistry_RegistrationAndResolution(t *testing.T) {
	ctx := context.Background()
	registry := NewTargetAdapterRegistry()

	// 1. Register Go Standard Library Target Adapter (framework=null)
	goAdapter := &MockTargetAdapter{
		manifest: TargetAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "target-go-stdlib",
			Version:              "1.0.0",
			Direction:            "TARGET",
			Language:             "go",
			Framework:            nil, // standard library
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities:         []string{"routing", "http_handler", "database", "session"},
			EvidenceRefs:         []string{"sha256:" + strings.Repeat("a", 64)},
		},
	}
	if err := registry.Register(goAdapter); err != nil {
		t.Fatalf("failed to register goAdapter: %v", err)
	}

	// 2. Register Laravel 13 Target Adapter (framework="laravel")
	fwLaravel := "laravel"
	laravelAdapter := &MockTargetAdapter{
		manifest: TargetAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "target-php-laravel13",
			Version:              "1.0.0",
			Direction:            "TARGET",
			Language:             "php",
			Framework:            &fwLaravel,
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities:         []string{"routing", "views", "auth", "session", "csrf", "database"},
			EvidenceRefs:         []string{"sha256:" + strings.Repeat("b", 64)},
		},
	}
	if err := registry.Register(laravelAdapter); err != nil {
		t.Fatalf("failed to register laravelAdapter: %v", err)
	}

	// Duplicate registration must fail
	if err := registry.Register(goAdapter); err == nil {
		t.Fatalf("expected error on duplicate registration, got nil")
	}

	// 3. Resolve by ID
	resolved, err := registry.Resolve("target-go-stdlib")
	if err != nil {
		t.Fatalf("failed to resolve target-go-stdlib: %v", err)
	}
	if resolved.Manifest().AdapterID != "target-go-stdlib" {
		t.Errorf("expected target-go-stdlib, got %s", resolved.Manifest().AdapterID)
	}

	// 4. Resolve by ResolvedEnvironmentDocument (standard library: framework=nil)
	goEnv := &ResolvedEnvironmentDocument{
		Language:  "go",
		Framework: nil, // standard library
	}
	resolvedEnvAdapter, err := registry.ResolveForResolution(goEnv)
	if err != nil {
		t.Fatalf("failed to resolve for goEnv: %v", err)
	}
	if resolvedEnvAdapter.Manifest().AdapterID != "target-go-stdlib" {
		t.Errorf("expected target-go-stdlib, got %s", resolvedEnvAdapter.Manifest().AdapterID)
	}

	// 5. Resolve by ResolvedEnvironmentDocument (framework=laravel)
	phpEnv := &ResolvedEnvironmentDocument{
		Language: "php",
		Framework: &ResolvedComponent{
			Name: "laravel",
		},
	}
	resolvedLaravelAdapter, err := registry.ResolveForResolution(phpEnv)
	if err != nil {
		t.Fatalf("failed to resolve for phpEnv: %v", err)
	}
	if resolvedLaravelAdapter.Manifest().AdapterID != "target-php-laravel13" {
		t.Errorf("expected target-php-laravel13, got %s", resolvedLaravelAdapter.Manifest().AdapterID)
	}

	// 6. Capability validation: supported vs unsupported
	valRes, err := goAdapter.ValidateCapabilities(ctx, nil, []string{"routing", "database", "unsupported_blockchain_operator"})
	if err != nil {
		t.Fatalf("ValidateCapabilities failed: %v", err)
	}
	if valRes.Valid {
		t.Errorf("expected validation to be invalid due to unsupported capability")
	}
	if len(valRes.UnsupportedCaps) != 1 || valRes.UnsupportedCaps[0] != "unsupported_blockchain_operator" {
		t.Errorf("expected unsupported_blockchain_operator in unsupported caps, got %v", valRes.UnsupportedCaps)
	}
	if len(valRes.Blockers) != 1 {
		t.Errorf("expected 1 blocker, got %d", len(valRes.Blockers))
	}
}
