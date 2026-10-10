package domain

import (
	"context"
	"fmt"
	"sync"
)

// SourceAdapterCapability represents a capability supported by a source adapter.
type SourceAdapterCapability string

const (
	CapSourceSyntaxParsing          SourceAdapterCapability = "syntax_parsing"
	CapSourceSymbolExtraction        SourceAdapterCapability = "symbol_extraction"
	CapSourceEntrypointDiscovery     SourceAdapterCapability = "entrypoint_discovery"
	CapSourceRouteExtraction         SourceAdapterCapability = "route_extraction"
	CapSourceControllerExtraction    SourceAdapterCapability = "controller_extraction"
	CapSourceDatabaseAccess          SourceAdapterCapability = "database_access"
	CapSourceSessionManagement       SourceAdapterCapability = "session_management"
	CapSourceMiddlewareLifecycle     SourceAdapterCapability = "middleware_lifecycle"
	CapSourceProbeInstrumentation    SourceAdapterCapability = "probe_instrumentation"
)

// SourceAdapterManifest mirrors the manifest 2.1 / 2.2 schema properties for source adapters.
type SourceAdapterManifest struct {
	SchemaVersion        string   `json:"schema_version"`
	AdapterID            string   `json:"adapter_id"`
	Version              string   `json:"version"`
	Direction            string   `json:"direction"` // strictly "SOURCE"
	Language             string   `json:"language"`  // e.g. "php"
	Framework            *string  `json:"framework"` // null for generic language fallback, e.g. "codeigniter3"
	ImplementationStatus string   `json:"implementation_status"` // "IMPLEMENTED", "PLANNED"
	Maturity             string   `json:"maturity"`              // "EXPERIMENTAL", "BETA", "STABLE"
	Capabilities         []string `json:"capabilities"`
	Limitations          []string `json:"limitations"`
	SchemaRefs           []string `json:"schema_refs"`
	EvidenceRefs         []string `json:"evidence_refs"`
}

// SourceCapabilityValidationResult records the outcome of SourceAdapter.ValidateCapabilities.
type SourceCapabilityValidationResult struct {
	Valid           bool     `json:"valid"`
	SupportedCaps   []string `json:"supported_capabilities"`
	UnsupportedCaps []string `json:"unsupported_capabilities"`
	Blockers        []string `json:"blockers"`
}

// SourceDetectionInput provides snapshot context and candidate file trees to Detect.
type SourceDetectionInput struct {
	ProjectID      string
	SnapshotRef    string
	Files          map[string][]byte // relative path -> content
	DeclaredConfig map[string]string
}

// SourceDetectionResult records framework detection findings with confidence and evidence.
type SourceDetectionResult struct {
	Language        string                 `json:"language"`
	Framework       *string                `json:"framework"`
	DetectedVersion *string                `json:"detected_version,omitempty"`
	Confidence      float64                `json:"confidence"` // 0.0 to 1.0
	Evidence        []string               `json:"evidence"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

// SourceParseInput provides file contents for AST and symbol parsing into Source IR.
type SourceParseInput struct {
	ProjectID   string
	SnapshotRef string
	Files       map[string][]byte
}

// SourceParseOutput provides the parsed SourceIRDocument and file-level parse statuses.
type SourceParseOutput struct {
	Document *SourceIRDocument
	Statuses map[string]SyntaxStatus
}

// SourceExtractionInput provides Source IR and framework context for Application IR extraction.
type SourceExtractionInput struct {
	ProjectID   string
	SnapshotRef string
	SourceIR    *SourceIRDocument
	Files       map[string][]byte
}

// SourceExtractionOutput provides the extracted ApplicationIRDocument.
type SourceExtractionOutput struct {
	Document *ApplicationIRDocument
	Gaps     []string
}

// SourceInstrumentationSpec defines probe overlay and hook requirements for runtime observation.
type SourceInstrumentationSpec struct {
	AdapterID        string            `json:"adapter_id"`
	Language         string            `json:"language"`
	Framework        *string           `json:"framework"`
	ProbeOverlayPath string            `json:"probe_overlay_path"`
	HookConfig       map[string]string `json:"hook_config"`
	DisposableCopy   bool              `json:"disposable_copy"` // strictly true per CI3-ADAPTER
}

// SourceAdapter defines the core port contract for all source framework/language adapters
// per SERVICE-CONTRACTS, STATIC-ANALYSIS, and SKILL-SPEC.
//
// Invariants:
// 1. Core remains neutral: source adapters own framework-specific conventions and parsing;
//    core consumes only neutral Source IR and Application IR.
// 2. ValidateCapabilities checks requested contract obligations against adapter capabilities,
//    returning explicit typed blockers for unsupported facets rather than guessing.
// 3. Detect identifies language/framework presence with evidence and confidence without execution.
// 4. Parse extracts files, symbols, syntax statuses, and findings into Source IR.
// 5. Extract resolves routing, entrypoints, and components into Application IR, preserving gaps as UNKNOWN.
// 6. PrepareInstrumentation specifies disposable probe overlays without modifying original sources.
type SourceAdapter interface {
	Manifest() SourceAdapterManifest
	ValidateCapabilities(ctx context.Context, requiredCaps []string) (*SourceCapabilityValidationResult, error)
	Detect(ctx context.Context, input SourceDetectionInput) (*SourceDetectionResult, error)
	Parse(ctx context.Context, input SourceParseInput) (*SourceParseOutput, error)
	Extract(ctx context.Context, input SourceExtractionInput) (*SourceExtractionOutput, error)
	PrepareInstrumentation(ctx context.Context, input SourceExtractionInput) (*SourceInstrumentationSpec, error)
}

// SourceAdapterRegistry manages immutable registration and resolution of qualified source adapters.
type SourceAdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]SourceAdapter
}

// NewSourceAdapterRegistry constructs an empty source adapter registry.
func NewSourceAdapterRegistry() *SourceAdapterRegistry {
	return &SourceAdapterRegistry{
		adapters: make(map[string]SourceAdapter),
	}
}

// Register adds a source adapter to the registry. Fails on duplicate registration or invalid manifest.
func (r *SourceAdapterRegistry) Register(adapter SourceAdapter) error {
	if adapter == nil {
		return fmt.Errorf("cannot register nil adapter")
	}
	m := adapter.Manifest()
	if m.AdapterID == "" {
		return fmt.Errorf("adapter manifest missing adapter_id")
	}
	if m.Direction != "SOURCE" {
		return fmt.Errorf("cannot register adapter %s with direction %s in SourceAdapterRegistry", m.AdapterID, m.Direction)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[m.AdapterID]; exists {
		return fmt.Errorf("duplicate source adapter registration: %s", m.AdapterID)
	}
	r.adapters[m.AdapterID] = adapter
	return nil
}

// Resolve retrieves a source adapter by its adapter ID.
func (r *SourceAdapterRegistry) Resolve(adapterID string) (SourceAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[adapterID]
	if !exists {
		return nil, fmt.Errorf("source adapter not found: %s", adapterID)
	}
	return adapter, nil
}

// ResolveForEnvironment locates the qualified source adapter matching language and framework.
// If framework is nil or empty, resolves the generic language fallback adapter (framework=null).
func (r *SourceAdapterRegistry) ResolveForEnvironment(language string, framework *string) (SourceAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, adapter := range r.adapters {
		m := adapter.Manifest()
		if m.Language != language {
			continue
		}

		// Generic fallback: framework == nil && m.Framework == nil
		if (framework == nil || *framework == "") && m.Framework == nil {
			return adapter, nil
		}

		// Framework match: framework != nil && m.Framework != nil && *framework == *m.Framework
		if framework != nil && *framework != "" && m.Framework != nil && *framework == *m.Framework {
			return adapter, nil
		}
	}

	fwName := "generic-language-fallback"
	if framework != nil && *framework != "" {
		fwName = *framework
	}
	return nil, fmt.Errorf("no qualified source adapter registered for language=%s framework=%s", language, fwName)
}

// ListAdapters returns manifests of all registered source adapters.
func (r *SourceAdapterRegistry) ListAdapters() []SourceAdapterManifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]SourceAdapterManifest, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		list = append(list, adapter.Manifest())
	}
	return list
}
