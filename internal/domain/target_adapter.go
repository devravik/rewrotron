package domain

import (
	"context"
	"fmt"
	"sync"
)

// TargetAdapterCapability represents a capability supported by a target adapter.
type TargetAdapterCapability string

const (
	CapTargetRouting         TargetAdapterCapability = "routing"
	CapTargetHTTPHandler     TargetAdapterCapability = "http_handler"
	CapTargetAuth            TargetAdapterCapability = "auth"
	CapTargetSession         TargetAdapterCapability = "session"
	CapTargetCSRF            TargetAdapterCapability = "csrf"
	CapTargetDatabase        TargetAdapterCapability = "database"
	CapTargetTransactions    TargetAdapterCapability = "transactions"
	CapTargetViews           TargetAdapterCapability = "views"
	CapTargetJSONAPI         TargetAdapterCapability = "json_api"
	CapTargetAsyncEffects    TargetAdapterCapability = "async_effects"
)

// TargetAdapterManifest mirrors the manifest 2.1 / 2.2 schema properties for target adapters.
type TargetAdapterManifest struct {
	SchemaVersion        string   `json:"schema_version"`
	AdapterID            string   `json:"adapter_id"`
	Version              string   `json:"version"`
	Direction            string   `json:"direction"` // strictly "TARGET"
	Language             string   `json:"language"`  // e.g. "go", "php"
	Framework            *string  `json:"framework"` // null for standard library, e.g. "laravel"
	ImplementationStatus string   `json:"implementation_status"` // "IMPLEMENTED", "PLANNED"
	Maturity             string   `json:"maturity"`              // "EXPERIMENTAL", "BETA", "STABLE"
	Capabilities         []string `json:"capabilities"`
	Limitations          []string `json:"limitations"`
	SchemaRefs           []string `json:"schema_refs"`
	EvidenceRefs         []string `json:"evidence_refs"`
}

// CapabilityValidationResult records the outcome of TargetAdapter.ValidateCapabilities.
type CapabilityValidationResult struct {
	Valid          bool     `json:"valid"`
	SupportedCaps  []string `json:"supported_capabilities"`
	UnsupportedCaps []string `json:"unsupported_capabilities"`
	Blockers       []string `json:"blockers"`
}

// TargetHarnessSpec defines the execution harness requirements for a generated target.
type TargetHarnessSpec struct {
	AdapterID       string   `json:"adapter_id"`
	Language        string   `json:"language"`
	Framework       *string  `json:"framework"`
	EntrypointPath  string   `json:"entrypoint_path"`
	BuildCommand    []string `json:"build_command"`
	RunCommand      []string `json:"run_command"`
	EnvironmentVars []string `json:"environment_vars"`
}

// TargetAdapter defines the core port contract for all target framework/language adapters
// per SERVICE-CONTRACTS, TARGET-ARCHITECTURE, and SKILL-SPEC.
//
// Invariants:
// 1. Core remains neutral: adapters interpret target-specific lowerings; core holds neutral TIM.
// 2. ValidateCapabilities checks requested contract obligations against adapter capabilities,
//    returning explicit typed blockers for unsupported facets rather than inventing code.
// 3. Generate produces deterministic target files and a sealed GenerationManifest.
// 4. PrepareHarness specifies build and execution controls without claiming verifier authority.
type TargetAdapter interface {
	Manifest() TargetAdapterManifest
	ValidateCapabilities(ctx context.Context, targetModel *TargetModelDocument, requiredCaps []string) (*CapabilityValidationResult, error)
	Generate(ctx context.Context, input GenerationInput) (*GenerationOutput, error)
	PrepareHarness(ctx context.Context, targetModel *TargetModelDocument) (*TargetHarnessSpec, error)
}

// TargetAdapterRegistry manages immutable registration and resolution of qualified target adapters.
type TargetAdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]TargetAdapter
}

// NewTargetAdapterRegistry constructs an empty target adapter registry.
func NewTargetAdapterRegistry() *TargetAdapterRegistry {
	return &TargetAdapterRegistry{
		adapters: make(map[string]TargetAdapter),
	}
}

// Register adds a target adapter to the registry. Fails on duplicate registration.
func (r *TargetAdapterRegistry) Register(adapter TargetAdapter) error {
	if adapter == nil {
		return fmt.Errorf("cannot register nil adapter")
	}
	m := adapter.Manifest()
	if m.AdapterID == "" {
		return fmt.Errorf("adapter manifest missing adapter_id")
	}
	if m.Direction != "TARGET" {
		return fmt.Errorf("cannot register adapter %s with direction %s in TargetAdapterRegistry", m.AdapterID, m.Direction)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[m.AdapterID]; exists {
		return fmt.Errorf("duplicate target adapter registration: %s", m.AdapterID)
	}
	r.adapters[m.AdapterID] = adapter
	return nil
}

// Resolve retrieves a target adapter by its adapter ID.
func (r *TargetAdapterRegistry) Resolve(adapterID string) (TargetAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[adapterID]
	if !exists {
		return nil, fmt.Errorf("target adapter not found: %s", adapterID)
	}
	return adapter, nil
}

// ResolveForResolution locates the qualified target adapter matching the exact resolved environment.
func (r *TargetAdapterRegistry) ResolveForResolution(env *ResolvedEnvironmentDocument) (TargetAdapter, error) {
	if env == nil {
		return nil, fmt.Errorf("resolved environment is nil")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, adapter := range r.adapters {
		m := adapter.Manifest()
		if m.Language != env.Language {
			continue
		}

		// Standard library: env.Framework == nil && m.Framework == nil
		if env.Framework == nil && m.Framework == nil {
			return adapter, nil
		}

		// Framework match: env.Framework.Name == *m.Framework
		if env.Framework != nil && m.Framework != nil && env.Framework.Name == *m.Framework {
			return adapter, nil
		}
	}

	fwName := "standard-library"
	if env.Framework != nil {
		fwName = env.Framework.Name
	}
	return nil, fmt.Errorf("no qualified target adapter registered for language=%s framework=%s", env.Language, fwName)
}

// ListAdapters returns manifests of all registered adapters.
func (r *TargetAdapterRegistry) ListAdapters() []TargetAdapterManifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]TargetAdapterManifest, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		list = append(list, adapter.Manifest())
	}
	return list
}
