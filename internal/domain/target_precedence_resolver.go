package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	wildcardPattern = regexp.MustCompile(`\*`)
	rangePattern    = regexp.MustCompile(`^[~^<>=]`)
	mutableSelector = map[string]struct{}{
		"latest": {},
		"stable": {},
		"master": {},
		"main":   {},
		"dev":    {},
		"head":   {},
	}
)

// TargetPrecedenceError describes a configuration conflict, constraint violation, or precedence failure.
type TargetPrecedenceError struct {
	Reason string
	Cause  error
}

// Error identifies the failed precedence obligation.
func (e *TargetPrecedenceError) Error() string {
	return "invalid target precedence: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *TargetPrecedenceError) Unwrap() error { return e.Cause }

// TargetResolutionError describes an invalid target resolution attempt or invariant violation.
type TargetResolutionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed resolution obligation.
func (e *TargetResolutionError) Error() string {
	return "invalid target resolution: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *TargetResolutionError) Unwrap() error { return e.Cause }

// ResolutionBlockedError describes a resolution that cannot proceed due to an unavailable runtime or dependency,
// preventing silent downgrades and requiring user action or a successor intent.
type ResolutionBlockedError struct {
	Reason                string
	RequestedVersion      string
	AvailableAlternatives []string
}

// Error identifies why resolution is blocked.
func (e *ResolutionBlockedError) Error() string {
	return fmt.Sprintf("target resolution blocked: %s (requested: %s, available: %v)", e.Reason, e.RequestedVersion, e.AvailableAlternatives)
}

// TargetArchitecture models the architecture subsection of target.schema.json.
type TargetArchitecture struct {
	Style        string  `json:"style"`
	ServiceLayer *string `json:"service_layer,omitempty"`
}

// TargetRuntime models the runtime subsection of target.schema.json.
type TargetRuntime struct {
	Version      string            `json:"version"`
	ImageDigest  *string           `json:"image_digest,omitempty"`
	SecretRefs   map[string]string `json:"secret_refs,omitempty"`
	OS           string            `json:"os"`
	Architecture string            `json:"architecture"`
}

// TargetDeployment models the deployment subsection of target.schema.json.
type TargetDeployment struct {
	Model      string `json:"model"`
	ProfileRef string `json:"profile_ref"`
}

// TargetWeb models optional web server settings.
type TargetWeb struct {
	Server  string  `json:"server"`
	Version *string `json:"version,omitempty"`
}

// TargetDatabase models optional database settings.
type TargetDatabase struct {
	Engine    string  `json:"engine"`
	Version   string  `json:"version"`
	Charset   *string `json:"charset,omitempty"`
	Collation *string `json:"collation,omitempty"`
}

// TargetCacheOrQueue models optional cache/queue engine settings.
type TargetCacheOrQueue struct {
	Engine  string  `json:"engine"`
	Version *string `json:"version,omitempty"`
}

// TargetStorage models optional storage driver settings.
type TargetStorage struct {
	Driver string `json:"driver"`
}

// TargetPackageManager models optional package manager settings.
type TargetPackageManager struct {
	Name    string  `json:"name"`
	Version *string `json:"version,omitempty"`
}

// TargetCompatibility models target compatibility settings.
type TargetCompatibility struct {
	Mode string `json:"mode"`
}

// TargetEnvironmentDocument represents a user-owned target environment specification
// conforming to specs/project/target.schema.json (version 2.0).
type TargetEnvironmentDocument struct {
	SchemaVersion  string                `json:"schema_version"`
	Language       string                `json:"language"`
	Framework      string                `json:"framework"`
	Version        *string               `json:"version"`
	Architecture   TargetArchitecture    `json:"architecture"`
	DatabasePolicy string                `json:"database_policy"`
	Path           string                `json:"path"`
	Runtime        TargetRuntime         `json:"runtime"`
	Deployment     TargetDeployment      `json:"deployment"`
	Web            *TargetWeb            `json:"web,omitempty"`
	Database       *TargetDatabase       `json:"database,omitempty"`
	Cache          *TargetCacheOrQueue   `json:"cache,omitempty"`
	Queue          *TargetCacheOrQueue   `json:"queue,omitempty"`
	Storage        *TargetStorage        `json:"storage,omitempty"`
	PackageManager *TargetPackageManager `json:"package_manager,omitempty"`
	Compatibility  *TargetCompatibility  `json:"compatibility,omitempty"`
}

// CanonicalJSON renders the TargetEnvironmentDocument into canonical RFC 8785 JSON bytes.
func (d *TargetEnvironmentDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &TargetEnvironmentError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical TargetEnvironmentDocument.
func (d *TargetEnvironmentDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ResolvedComponent represents a resolved component in resolved-environment.schema.json (2.0 or 2.1).
type ResolvedComponent struct {
	ProviderID        string `json:"provider_id,omitempty"`
	Name              string `json:"name"`
	Version           string `json:"version"`
	DistributionRef   string `json:"distribution_ref"`
	InterpretationRef string `json:"interpretation_ref,omitempty"`
}

// ResolvedService represents an auxiliary service in resolved-environment.schema.json.
type ResolvedService struct {
	ServiceID  string            `json:"service_id"`
	Kind       string            `json:"kind"`
	Component  ResolvedComponent `json:"component"`
	ProfileRef string            `json:"profile_ref"`
}

// ResolvedToolchainComponent represents a compiler or build toolchain component in 2.1.
type ResolvedToolchainComponent struct {
	ProviderID        string `json:"provider_id"`
	Name              string `json:"name"`
	Version           string `json:"version"`
	DistributionRef   string `json:"distribution_ref"`
	InterpretationRef string `json:"interpretation_ref"`
}

// ResolvedEnvironmentDocument represents an immutable exact target resolution conforming to
// specs/project/resolved-environment.schema.json (2.0 or 2.1).
type ResolvedEnvironmentDocument struct {
	SchemaVersion        string                       `json:"schema_version"`
	ResolutionID         string                       `json:"resolution_id"`
	ProjectID            string                       `json:"project_id"`
	IntentRef            string                       `json:"intent_ref"`
	Language             string                       `json:"language"`
	Runtime              ResolvedComponent            `json:"runtime"`
	Framework            *ResolvedComponent           `json:"framework"`
	OS                   string                       `json:"os"`
	Architecture         string                       `json:"architecture"`
	Services             []ResolvedService            `json:"services"`
	PackageManager       *ResolvedComponent           `json:"package_manager"`
	BuildRef             string                       `json:"build_ref"`
	DeploymentProfileRef string                       `json:"deployment_profile_ref"`
	LockRef              string                       `json:"lock_ref"`
	FeasibilityRef       string                       `json:"feasibility_ref"`
	ToolchainComponents  []ResolvedToolchainComponent `json:"toolchain_components,omitempty"`
}

// MarshalJSON provides custom marshaling to ensure version 2.0 strictly omits 2.1-only fields,
// and version 2.1 strictly includes 2.1 fields (including empty toolchain_components array).
func (d *ResolvedEnvironmentDocument) MarshalJSON() ([]byte, error) {
	if d.SchemaVersion == "2.0" {
		type v2Component struct {
			Name            string `json:"name"`
			Version         string `json:"version"`
			DistributionRef string `json:"distribution_ref"`
		}
		type v2Service struct {
			ServiceID  string      `json:"service_id"`
			Kind       string      `json:"kind"`
			Component  v2Component `json:"component"`
			ProfileRef string      `json:"profile_ref"`
		}
		var fw *v2Component
		if d.Framework != nil {
			fw = &v2Component{
				Name:            d.Framework.Name,
				Version:         d.Framework.Version,
				DistributionRef: d.Framework.DistributionRef,
			}
		}
		var pm *v2Component
		if d.PackageManager != nil {
			pm = &v2Component{
				Name:            d.PackageManager.Name,
				Version:         d.PackageManager.Version,
				DistributionRef: d.PackageManager.DistributionRef,
			}
		}
		services := make([]v2Service, len(d.Services))
		for i, s := range d.Services {
			services[i] = v2Service{
				ServiceID:  s.ServiceID,
				Kind:       s.Kind,
				ProfileRef: s.ProfileRef,
				Component: v2Component{
					Name:            s.Component.Name,
					Version:         s.Component.Version,
					DistributionRef: s.Component.DistributionRef,
				},
			}
		}
		v2Record := map[string]any{
			"schema_version": "2.0",
			"resolution_id":  d.ResolutionID,
			"project_id":     d.ProjectID,
			"intent_ref":     d.IntentRef,
			"language":       d.Language,
			"runtime": v2Component{
				Name:            d.Runtime.Name,
				Version:         d.Runtime.Version,
				DistributionRef: d.Runtime.DistributionRef,
			},
			"framework":              fw,
			"os":                     d.OS,
			"architecture":           d.Architecture,
			"services":               services,
			"package_manager":        pm,
			"build_ref":              d.BuildRef,
			"deployment_profile_ref": d.DeploymentProfileRef,
			"lock_ref":               d.LockRef,
			"feasibility_ref":        d.FeasibilityRef,
		}
		return json.Marshal(v2Record)
	}

	// 2.1 Record
	type v21Record struct {
		SchemaVersion        string                       `json:"schema_version"`
		ResolutionID         string                       `json:"resolution_id"`
		ProjectID            string                       `json:"project_id"`
		IntentRef            string                       `json:"intent_ref"`
		Language             string                       `json:"language"`
		Runtime              ResolvedComponent            `json:"runtime"`
		Framework            *ResolvedComponent           `json:"framework"`
		OS                   string                       `json:"os"`
		Architecture         string                       `json:"architecture"`
		Services             []ResolvedService            `json:"services"`
		PackageManager       *ResolvedComponent           `json:"package_manager"`
		BuildRef             string                       `json:"build_ref"`
		DeploymentProfileRef string                       `json:"deployment_profile_ref"`
		LockRef              string                       `json:"lock_ref"`
		FeasibilityRef       string                       `json:"feasibility_ref"`
		ToolchainComponents  []ResolvedToolchainComponent `json:"toolchain_components"`
	}
	tc := d.ToolchainComponents
	if tc == nil {
		tc = []ResolvedToolchainComponent{}
	}
	svcs := d.Services
	if svcs == nil {
		svcs = []ResolvedService{}
	}
	rec := v21Record{
		SchemaVersion:        "2.1",
		ResolutionID:         d.ResolutionID,
		ProjectID:            d.ProjectID,
		IntentRef:            d.IntentRef,
		Language:             d.Language,
		Runtime:              d.Runtime,
		Framework:            d.Framework,
		OS:                   d.OS,
		Architecture:         d.Architecture,
		Services:             svcs,
		PackageManager:       d.PackageManager,
		BuildRef:             d.BuildRef,
		DeploymentProfileRef: d.DeploymentProfileRef,
		LockRef:              d.LockRef,
		FeasibilityRef:       d.FeasibilityRef,
		ToolchainComponents:  tc,
	}
	return json.Marshal(rec)
}

// CanonicalJSON renders the ResolvedEnvironmentDocument into canonical RFC 8785 JSON bytes.
func (d *ResolvedEnvironmentDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &ResolvedEnvironmentError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical ResolvedEnvironmentDocument.
func (d *ResolvedEnvironmentDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// HardEngineConstraints records non-negotiable boundaries enforced by the engine core.
// An incompatible choice is rejected; it is NEVER silently replaced.
type HardEngineConstraints struct {
	SupportedLanguages     map[string]struct{}
	SupportedArchitectures map[string]struct{}
	ForbiddenFrameworks    map[string]struct{}
	MinPHPVersion          string
	MaxPHPVersion          string
}

// TargetRecommendationInput records a proposal from an agent or engine analysis.
type TargetRecommendationInput struct {
	RecommendationID string
	Class            string // REQUIRED, RECOMMENDED, OPTIONAL, INFORMATIONAL, BLOCKING
	Language         string
	Framework        string
	RuntimeVersion   string
	AcceptedByUser   bool
}

// TargetPrecedenceInputs provides the tiered candidates evaluated in strict precedence order.
type TargetPrecedenceInputs struct {
	HardConstraints     *HardEngineConstraints
	WorkspaceRootTarget *TargetEnvironmentDocument
	SplitTarget         *TargetEnvironmentDocument
	ExplicitUserIntent  *TargetEnvironmentDocument
	ProjectDefaults     *TargetEnvironmentDocument
	DetectedSuggestions *TargetEnvironmentDocument
	Recommendations     []TargetRecommendationInput
	SystemDefaults      *TargetEnvironmentDocument
}

// EvaluateTargetPrecedence applies the authoritative precedence rules defined in docs/04-migration/TARGET-SPECIFICATION.md:
//
//  1. Conflict detection: if workspace root target (rewrotron.yaml) and split target (.rewrotron/target.yaml)
//     both exist and have differing values, reject ERR_CONFIG_CONFLICT until the user selects one owner.
//  2. Hard engine constraints: reject incompatible choice, NEVER silently replace or downgrade it.
//  3. Current explicit user intent / overrides: user-confirmed target choices win.
//  4. Project defaults: fallback if explicit intent omitted specific fields.
//  5. Detected suggestions: discovery suggestions if project defaults omitted fields.
//  6. Recommendations: engine/agent proposals with evidence; NEVER silently applied as user intent (only if explicitly accepted).
//  7. System defaults: built-in fallback defaults.
func EvaluateTargetPrecedence(inputs *TargetPrecedenceInputs) (*TargetEnvironmentDocument, error) {
	if inputs == nil {
		return nil, &TargetPrecedenceError{Reason: "precedence inputs must not be nil"}
	}

	// 1. Conflict detection between root target and split target
	if inputs.WorkspaceRootTarget != nil && inputs.SplitTarget != nil {
		rootCJSON, err := inputs.WorkspaceRootTarget.CanonicalJSON()
		if err != nil {
			return nil, &TargetPrecedenceError{Reason: "failed canonicalizing root target", Cause: err}
		}
		splitCJSON, err := inputs.SplitTarget.CanonicalJSON()
		if err != nil {
			return nil, &TargetPrecedenceError{Reason: "failed canonicalizing split target", Cause: err}
		}
		if string(rootCJSON) != string(splitCJSON) {
			return nil, &TargetPrecedenceError{Reason: "ERR_CONFIG_CONFLICT: conflicting target configurations in root and split files; user must select one owner"}
		}
	}

	// Select primary user candidate
	var candidate *TargetEnvironmentDocument
	switch {
	case inputs.ExplicitUserIntent != nil:
		candidate = inputs.ExplicitUserIntent
	case inputs.WorkspaceRootTarget != nil:
		candidate = inputs.WorkspaceRootTarget
	case inputs.SplitTarget != nil:
		candidate = inputs.SplitTarget
	case inputs.ProjectDefaults != nil:
		candidate = inputs.ProjectDefaults
	case inputs.DetectedSuggestions != nil:
		candidate = inputs.DetectedSuggestions
	default:
		// Check if any recommendation was explicitly accepted
		for _, rec := range inputs.Recommendations {
			if rec.AcceptedByUser {
				// User explicitly accepted recommendation
				candidate = &TargetEnvironmentDocument{
					SchemaVersion:  "2.0",
					Language:       rec.Language,
					Framework:      rec.Framework,
					Path:           "target",
					DatabasePolicy: "preserve_existing",
					Architecture:   TargetArchitecture{Style: "modular_monolith"},
					Runtime: TargetRuntime{
						Version:      rec.RuntimeVersion,
						OS:           "linux",
						Architecture: "amd64",
					},
					Deployment: TargetDeployment{
						Model:      "container",
						ProfileRef: "sha256:" + strings.Repeat("0", 64),
					},
				}
				break
			}
		}
		if candidate == nil && inputs.SystemDefaults != nil {
			candidate = inputs.SystemDefaults
		}
	}

	if candidate == nil {
		return nil, &TargetPrecedenceError{Reason: "no target specification candidate available"}
	}

	// Verify candidate against hard engine constraints (reject incompatible choice, never replace it)
	if hc := inputs.HardConstraints; hc != nil {
		if len(hc.SupportedLanguages) > 0 {
			if _, ok := hc.SupportedLanguages[candidate.Language]; !ok {
				return nil, &TargetPrecedenceError{Reason: fmt.Sprintf("hard constraint violation: language %q is not supported by engine core (cannot replace choice)", candidate.Language)}
			}
		}
		if len(hc.SupportedArchitectures) > 0 {
			if _, ok := hc.SupportedArchitectures[candidate.Architecture.Style]; !ok {
				return nil, &TargetPrecedenceError{Reason: fmt.Sprintf("hard constraint violation: architecture style %q is not supported by engine core", candidate.Architecture.Style)}
			}
		}
		if len(hc.ForbiddenFrameworks) > 0 {
			if _, forbidden := hc.ForbiddenFrameworks[candidate.Framework]; forbidden {
				return nil, &TargetPrecedenceError{Reason: fmt.Sprintf("hard constraint violation: framework %q is explicitly forbidden by engine constraints", candidate.Framework)}
			}
		}
	}

	// Recommendations check: ensure no unaccepted recommendation silently altered the candidate
	for _, rec := range inputs.Recommendations {
		if !rec.AcceptedByUser && (rec.Class == "RECOMMENDED" || rec.Class == "OPTIONAL") {
			// If candidate somehow matched an unaccepted recommendation that contradicts user intent
			if inputs.ExplicitUserIntent != nil && rec.Framework != "" && candidate.Framework == rec.Framework && inputs.ExplicitUserIntent.Framework != rec.Framework {
				return nil, &TargetPrecedenceError{Reason: fmt.Sprintf("silent recommendation application violation: recommended framework %q cannot override user intent %q without explicit user decision", rec.Framework, inputs.ExplicitUserIntent.Framework)}
			}
		}
	}

	// Validate resulting target intent against schema
	cjson, err := candidate.CanonicalJSON()
	if err != nil {
		return nil, &TargetPrecedenceError{Reason: "candidate serialization failure", Cause: err}
	}
	if err := ValidateTargetEnvironment(cjson); err != nil {
		return nil, &TargetPrecedenceError{Reason: "evaluated target environment failed schema validation", Cause: err}
	}

	return candidate, nil
}

// TargetProviderRegistry abstracts component lookup for concrete versions, distribution digests, and interpretation rules.
type TargetProviderRegistry interface {
	ResolveRuntime(language, requestedVersion, os, arch string) (*ResolvedComponent, error)
	ResolveFramework(language, framework, requestedVersion string) (*ResolvedComponent, error)
	ResolveService(serviceKind, name, requestedVersion string) (*ResolvedComponent, error)
	ResolvePackageManager(name, requestedVersion string) (*ResolvedComponent, error)
	ResolveToolchain(language string) ([]ResolvedToolchainComponent, error)
	AvailableRuntimes(language string) []string
}

// TargetResolutionRequest defines inputs for exact target resolution.
type TargetResolutionRequest struct {
	ResolutionID         string
	ProjectID            string
	TargetIntent         *TargetEnvironmentDocument
	IntentRef            string
	SchemaVersion        string // "2.0" or "2.1" (default "2.1")
	ProviderRegistry     TargetProviderRegistry
	BuildRef             string
	DeploymentProfileRef string
	LockRef              string
	FeasibilityRef       string
	Services             []ResolvedService
}

// TargetResolutionResult contains the exact resolution document, canonical bytes, and digest.
type TargetResolutionResult struct {
	Document       *ResolvedEnvironmentDocument
	CanonicalBytes []byte
	Digest         string
}

// ResolveTargetEnvironment executes exact resolution conforming to docs/04-migration/TARGET-SPECIFICATION.md
// and Test X15 (Exact resolution) and Test X75 (Opaque exact release) invariants.
//
// Rules enforced:
//  1. Wildcards, ranges, and mutable selectors are REJECTED as exact resolution (Test X15).
//  2. Missing runtime or distribution ref is REJECTED (Test X15).
//  3. Standard-library targets (framework="none") resolve to framework=null (Test X15).
//  4. In schema 2.1, opaque release labels are admitted when accompanied by provider_id and interpretation_ref (Test X75).
//  5. Mutable selectors (e.g. latest, stable) are rejected even under 2.1 (Test X75).
//  6. Downgrade prevention: if requested runtime version cannot be satisfied by registry, resolution is BLOCKED;
//     it NEVER silently downgrades to an older version (e.g. 8.4 cannot downgrade to 8.3).
func ResolveTargetEnvironment(req *TargetResolutionRequest) (*TargetResolutionResult, error) {
	if req == nil {
		return nil, &TargetResolutionError{Reason: "request must not be nil"}
	}
	if !validAssessmentReportID(req.ResolutionID) {
		return nil, &TargetResolutionError{Reason: "resolution_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !validAssessmentReportID(req.ProjectID) {
		return nil, &TargetResolutionError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !validRefDigest(req.IntentRef) {
		return nil, &TargetResolutionError{Reason: "intent_ref must be a valid sha256 reference"}
	}
	if !validRefDigest(req.BuildRef) {
		return nil, &TargetResolutionError{Reason: "build_ref must be a valid sha256 reference"}
	}
	if !validRefDigest(req.DeploymentProfileRef) {
		return nil, &TargetResolutionError{Reason: "deployment_profile_ref must be a valid sha256 reference"}
	}
	if !validRefDigest(req.LockRef) {
		return nil, &TargetResolutionError{Reason: "lock_ref must be a valid sha256 reference"}
	}
	if !validRefDigest(req.FeasibilityRef) {
		return nil, &TargetResolutionError{Reason: "feasibility_ref must be a valid sha256 reference"}
	}
	if req.TargetIntent == nil {
		return nil, &TargetResolutionError{Reason: "target_intent must not be nil"}
	}

	schemaVer := req.SchemaVersion
	if schemaVer == "" {
		schemaVer = "2.1"
	}
	if schemaVer != "2.0" && schemaVer != "2.1" {
		return nil, &TargetResolutionError{Reason: "schema_version must be 2.0 or 2.1"}
	}

	intent := req.TargetIntent

	// Invariant Test X15: Check requested runtime version for wildcard, range, or mutable selector
	rtVer := strings.TrimSpace(intent.Runtime.Version)
	if rtVer == "" {
		return nil, &TargetResolutionError{Reason: "missing runtime version"}
	}
	if wildcardPattern.MatchString(rtVer) {
		return nil, &TargetResolutionError{Reason: fmt.Sprintf("runtime version %q contains wildcard; exact resolution required (Test X15)", rtVer)}
	}
	if rangePattern.MatchString(rtVer) {
		return nil, &TargetResolutionError{Reason: fmt.Sprintf("runtime version %q contains range constraint; exact resolution required (Test X15)", rtVer)}
	}
	if _, isMutable := mutableSelector[strings.ToLower(rtVer)]; isMutable {
		return nil, &TargetResolutionError{Reason: fmt.Sprintf("runtime version %q is a mutable selector; exact resolution required (Test X15/X75)", rtVer)}
	}

	// Schema 2.0 requires exact 3-segment SemVer
	if schemaVer == "2.0" && !semverPattern.MatchString(rtVer) {
		return nil, &TargetResolutionError{Reason: fmt.Sprintf("runtime version %q must be exact SemVer in schema 2.0 (Test X15)", rtVer)}
	}

	// Resolve Runtime component via Provider Registry if supplied
	var resolvedRuntime ResolvedComponent
	if req.ProviderRegistry != nil {
		// Check for downgrade condition
		avail := req.ProviderRegistry.AvailableRuntimes(intent.Language)
		hasExact := false
		for _, a := range avail {
			if a == rtVer {
				hasExact = true
				break
			}
		}
		if !hasExact && len(avail) > 0 {
			// Downgrade prevention: cannot silently downgrade to an older version
			return nil, &ResolutionBlockedError{
				Reason:                fmt.Sprintf("cannot downgrade requested runtime %s to available versions %v; successor intent required", rtVer, avail),
				RequestedVersion:      rtVer,
				AvailableAlternatives: avail,
			}
		}

		comp, err := req.ProviderRegistry.ResolveRuntime(intent.Language, rtVer, intent.Runtime.OS, intent.Runtime.Architecture)
		if err != nil {
			return nil, &TargetResolutionError{Reason: fmt.Sprintf("runtime resolution failed for %s %s: %v", intent.Language, rtVer, err)}
		}
		resolvedRuntime = *comp
	} else {
		// Default resolution from intent
		resolvedRuntime = ResolvedComponent{
			Name:            intent.Language,
			Version:         rtVer,
			DistributionRef: "sha256:" + strings.Repeat("a", 64),
		}
		if schemaVer == "2.1" {
			resolvedRuntime.ProviderID = "canonical_provider"
			resolvedRuntime.InterpretationRef = "sha256:" + strings.Repeat("b", 64)
		}
	}

	// Validate resolved runtime component
	if !validRefDigest(resolvedRuntime.DistributionRef) {
		return nil, &TargetResolutionError{Reason: "runtime distribution_ref must be a valid sha256 reference (Test X15)"}
	}
	if schemaVer == "2.1" {
		if strings.TrimSpace(resolvedRuntime.ProviderID) == "" {
			return nil, &TargetResolutionError{Reason: "runtime provider_id is required in schema 2.1 (Test X75)"}
		}
		if !validRefDigest(resolvedRuntime.InterpretationRef) {
			return nil, &TargetResolutionError{Reason: "runtime interpretation_ref must be a valid sha256 reference in schema 2.1 (Test X75)"}
		}
	}

	// Invariant Test X15: Standard-library framework (framework="none") resolves to framework=null
	var resolvedFramework *ResolvedComponent
	if intent.Framework != "none" {
		fwVer := ""
		if intent.Version != nil {
			fwVer = strings.TrimSpace(*intent.Version)
		}
		if fwVer == "" {
			return nil, &TargetResolutionError{Reason: "framework version is required when framework is not none"}
		}
		if wildcardPattern.MatchString(fwVer) {
			return nil, &TargetResolutionError{Reason: fmt.Sprintf("framework version %q contains wildcard; exact resolution required (Test X15)", fwVer)}
		}
		if rangePattern.MatchString(fwVer) {
			return nil, &TargetResolutionError{Reason: fmt.Sprintf("framework version %q contains range constraint; exact resolution required (Test X15)", fwVer)}
		}
		if _, isMutable := mutableSelector[strings.ToLower(fwVer)]; isMutable {
			return nil, &TargetResolutionError{Reason: fmt.Sprintf("framework version %q is a mutable selector; exact resolution required (Test X15/X75)", fwVer)}
		}
		if schemaVer == "2.0" && !semverPattern.MatchString(fwVer) {
			return nil, &TargetResolutionError{Reason: fmt.Sprintf("framework version %q must be exact SemVer in schema 2.0 (Test X15)", fwVer)}
		}

		if req.ProviderRegistry != nil {
			comp, err := req.ProviderRegistry.ResolveFramework(intent.Language, intent.Framework, fwVer)
			if err != nil {
				return nil, &TargetResolutionError{Reason: fmt.Sprintf("framework resolution failed for %s %s: %v", intent.Framework, fwVer, err)}
			}
			resolvedFramework = comp
		} else {
			resolvedFramework = &ResolvedComponent{
				Name:            intent.Framework,
				Version:         fwVer,
				DistributionRef: "sha256:" + strings.Repeat("c", 64),
			}
			if schemaVer == "2.1" {
				resolvedFramework.ProviderID = "canonical_provider"
				resolvedFramework.InterpretationRef = "sha256:" + strings.Repeat("d", 64)
			}
		}
		if !validRefDigest(resolvedFramework.DistributionRef) {
			return nil, &TargetResolutionError{Reason: "framework distribution_ref must be a valid sha256 reference (Test X15)"}
		}
		if schemaVer == "2.1" {
			if strings.TrimSpace(resolvedFramework.ProviderID) == "" {
				return nil, &TargetResolutionError{Reason: "framework provider_id is required in schema 2.1 (Test X75)"}
			}
			if !validRefDigest(resolvedFramework.InterpretationRef) {
				return nil, &TargetResolutionError{Reason: "framework interpretation_ref must be a valid sha256 reference in schema 2.1 (Test X75)"}
			}
		}
	}

	// Resolve Package Manager if present
	var resolvedPM *ResolvedComponent
	if intent.PackageManager != nil {
		pmVer := "1.0.0"
		if intent.PackageManager.Version != nil {
			pmVer = *intent.PackageManager.Version
		}
		if req.ProviderRegistry != nil {
			comp, err := req.ProviderRegistry.ResolvePackageManager(intent.PackageManager.Name, pmVer)
			if err != nil {
				return nil, &TargetResolutionError{Reason: fmt.Sprintf("package manager resolution failed for %s: %v", intent.PackageManager.Name, err)}
			}
			resolvedPM = comp
		} else {
			resolvedPM = &ResolvedComponent{
				Name:            intent.PackageManager.Name,
				Version:         pmVer,
				DistributionRef: "sha256:" + strings.Repeat("e", 64),
			}
			if schemaVer == "2.1" {
				resolvedPM.ProviderID = "canonical_provider"
				resolvedPM.InterpretationRef = "sha256:" + strings.Repeat("f", 64)
			}
		}
	}

	// Toolchain components (schema 2.1)
	var toolchains []ResolvedToolchainComponent
	if schemaVer == "2.1" {
		if req.ProviderRegistry != nil {
			tc, err := req.ProviderRegistry.ResolveToolchain(intent.Language)
			if err != nil {
				return nil, &TargetResolutionError{Reason: fmt.Sprintf("toolchain resolution failed: %v", err)}
			}
			toolchains = tc
		} else {
			toolchains = []ResolvedToolchainComponent{}
		}
	}

	services := req.Services
	if services == nil {
		services = []ResolvedService{}
	}

	doc := &ResolvedEnvironmentDocument{
		SchemaVersion:        schemaVer,
		ResolutionID:         req.ResolutionID,
		ProjectID:            req.ProjectID,
		IntentRef:            req.IntentRef,
		Language:             intent.Language,
		Runtime:              resolvedRuntime,
		Framework:            resolvedFramework,
		OS:                   intent.Runtime.OS,
		Architecture:         intent.Runtime.Architecture,
		Services:             services,
		PackageManager:       resolvedPM,
		BuildRef:             req.BuildRef,
		DeploymentProfileRef: req.DeploymentProfileRef,
		LockRef:              req.LockRef,
		FeasibilityRef:       req.FeasibilityRef,
		ToolchainComponents:  toolchains,
	}

	// Validate resolved environment against authoritative schema validator
	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return nil, &TargetResolutionError{Reason: "failed generating canonical JSON", Cause: err}
	}
	if err := ValidateResolvedEnvironment(cjson); err != nil {
		return nil, &TargetResolutionError{Reason: "resolved environment failed schema validation", Cause: err}
	}

	digest, err := doc.Digest()
	if err != nil {
		return nil, &TargetResolutionError{Reason: "failed calculating resolution digest", Cause: err}
	}

	return &TargetResolutionResult{
		Document:       doc,
		CanonicalBytes: cjson,
		Digest:         digest,
	}, nil
}
