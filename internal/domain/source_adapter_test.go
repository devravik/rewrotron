package domain

import (
	"context"
	"strings"
	"testing"
)

// mockSourceAdapter provides an implementation of SourceAdapter for unit testing the port & registry.
type mockSourceAdapter struct {
	manifest SourceAdapterManifest
}

func (m *mockSourceAdapter) Manifest() SourceAdapterManifest {
	return m.manifest
}

func (m *mockSourceAdapter) ValidateCapabilities(ctx context.Context, requiredCaps []string) (*SourceCapabilityValidationResult, error) {
	supportedMap := make(map[string]bool)
	for _, c := range m.manifest.Capabilities {
		supportedMap[c] = true
	}

	res := &SourceCapabilityValidationResult{
		Valid: true,
	}
	for _, req := range requiredCaps {
		if supportedMap[req] {
			res.SupportedCaps = append(res.SupportedCaps, req)
		} else {
			res.Valid = false
			res.UnsupportedCaps = append(res.UnsupportedCaps, req)
			res.Blockers = append(res.Blockers, "unsupported source capability: "+req)
		}
	}
	return res, nil
}

func (m *mockSourceAdapter) Detect(ctx context.Context, input SourceDetectionInput) (*SourceDetectionResult, error) {
	return &SourceDetectionResult{
		Language:   m.manifest.Language,
		Framework:  m.manifest.Framework,
		Confidence: 1.0,
		Evidence:   []string{"mock-file"},
	}, nil
}

func (m *mockSourceAdapter) Parse(ctx context.Context, input SourceParseInput) (*SourceParseOutput, error) {
	return &SourceParseOutput{
		Document: &SourceIRDocument{
			SchemaVersion: "1.0",
			ProjectID:     input.ProjectID,
		},
		Statuses: make(map[string]SyntaxStatus),
	}, nil
}

func (m *mockSourceAdapter) Extract(ctx context.Context, input SourceExtractionInput) (*SourceExtractionOutput, error) {
	return &SourceExtractionOutput{
		Document: &ApplicationIRDocument{
			SchemaVersion: "1.0",
			ProjectID:     input.ProjectID,
		},
	}, nil
}

func (m *mockSourceAdapter) PrepareInstrumentation(ctx context.Context, input SourceExtractionInput) (*SourceInstrumentationSpec, error) {
	return &SourceInstrumentationSpec{
		AdapterID:      m.manifest.AdapterID,
		Language:       m.manifest.Language,
		Framework:      m.manifest.Framework,
		DisposableCopy: true,
	}, nil
}

func TestSourceAdapterRegistry_RegisterAndResolve(t *testing.T) {
	reg := NewSourceAdapterRegistry()

	genericPHP := &mockSourceAdapter{
		manifest: SourceAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "source.php.generic",
			Version:              "1.0.0",
			Direction:            "SOURCE",
			Language:             "php",
			Framework:            nil,
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities:         []string{"syntax_parsing", "symbol_extraction"},
			EvidenceRefs:         []string{"sha256:" + strings.Repeat("a", 64)},
		},
	}

	ci3FW := "codeigniter3"
	ci3Adapter := &mockSourceAdapter{
		manifest: SourceAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "source.php.ci3",
			Version:              "1.0.0",
			Direction:            "SOURCE",
			Language:             "php",
			Framework:            &ci3FW,
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities:         []string{"syntax_parsing", "symbol_extraction", "route_extraction", "controller_extraction"},
			EvidenceRefs:         []string{"sha256:" + strings.Repeat("b", 64)},
		},
	}

	// Register generic PHP
	if err := reg.Register(genericPHP); err != nil {
		t.Fatalf("register genericPHP failed: %v", err)
	}

	// Register CI3
	if err := reg.Register(ci3Adapter); err != nil {
		t.Fatalf("register ci3Adapter failed: %v", err)
	}

	// Duplicate registration must fail
	if err := reg.Register(genericPHP); err == nil {
		t.Fatalf("expected error on duplicate registration, got nil")
	}

	// Resolve by ID
	resolved, err := reg.Resolve("source.php.generic")
	if err != nil || resolved.Manifest().AdapterID != "source.php.generic" {
		t.Fatalf("resolve by ID failed: %v", err)
	}

	// Resolve for environment: generic PHP (framework = nil)
	resolvedGeneric, err := reg.ResolveForEnvironment("php", nil)
	if err != nil || resolvedGeneric.Manifest().Framework != nil {
		t.Fatalf("resolve for generic environment failed: %v", err)
	}

	// Resolve for environment: CI3 (framework = "codeigniter3")
	fw := "codeigniter3"
	resolvedCI3, err := reg.ResolveForEnvironment("php", &fw)
	if err != nil || resolvedCI3.Manifest().Framework == nil || *resolvedCI3.Manifest().Framework != "codeigniter3" {
		t.Fatalf("resolve for CI3 environment failed: %v", err)
	}

	// Unregistered environment returns typed error
	unknownFW := "laravel"
	_, err = reg.ResolveForEnvironment("php", &unknownFW)
	if err == nil || !strings.Contains(err.Error(), "no qualified source adapter registered") {
		t.Fatalf("expected error for unregistered framework, got: %v", err)
	}

	// List adapters
	list := reg.ListAdapters()
	if len(list) != 2 {
		t.Fatalf("expected 2 adapters in list, got %d", len(list))
	}
}

func TestSourceAdapter_ValidateCapabilities(t *testing.T) {
	ctx := context.Background()
	adapter := &mockSourceAdapter{
		manifest: SourceAdapterManifest{
			SchemaVersion: "2.1",
			AdapterID:     "source.php.test",
			Direction:     "SOURCE",
			Language:      "php",
			Capabilities:  []string{"syntax_parsing", "symbol_extraction"},
		},
	}

	// Valid capabilities
	res, err := adapter.ValidateCapabilities(ctx, []string{"syntax_parsing"})
	if err != nil || !res.Valid || len(res.UnsupportedCaps) != 0 {
		t.Fatalf("expected valid capabilities, got: %+v, err: %v", res, err)
	}

	// Unsupported capabilities report typed blockers
	resBad, err := adapter.ValidateCapabilities(ctx, []string{"syntax_parsing", "database_access"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resBad.Valid {
		t.Fatalf("expected Valid=false for missing capability, got true")
	}
	if len(resBad.UnsupportedCaps) != 1 || resBad.UnsupportedCaps[0] != "database_access" {
		t.Fatalf("expected database_access in unsupported caps, got: %v", resBad.UnsupportedCaps)
	}
	if len(resBad.Blockers) != 1 {
		t.Fatalf("expected 1 blocker, got: %v", resBad.Blockers)
	}
}

func TestSourceAdapterRegistry_RejectsInvalidManifest(t *testing.T) {
	reg := NewSourceAdapterRegistry()

	// Nil adapter
	if err := reg.Register(nil); err == nil {
		t.Fatalf("expected error for nil adapter, got nil")
	}

	// Missing adapter_id
	noID := &mockSourceAdapter{
		manifest: SourceAdapterManifest{
			Direction: "SOURCE",
		},
	}
	if err := reg.Register(noID); err == nil {
		t.Fatalf("expected error for missing adapter_id, got nil")
	}

	// Wrong direction (TARGET instead of SOURCE)
	targetAdapter := &mockSourceAdapter{
		manifest: SourceAdapterManifest{
			AdapterID: "test.target",
			Direction: "TARGET",
		},
	}
	if err := reg.Register(targetAdapter); err == nil || !strings.Contains(err.Error(), "cannot register adapter") {
		t.Fatalf("expected error for wrong direction, got: %v", err)
	}
}
