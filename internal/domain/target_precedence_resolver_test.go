package domain

import (
	"errors"
	"strings"
	"testing"
)

type mockProviderRegistry struct {
	availableRuntimes map[string][]string
	runtimeComps      map[string]*ResolvedComponent
	frameworkComps    map[string]*ResolvedComponent
}

func (m *mockProviderRegistry) ResolveRuntime(language, requestedVersion, os, arch string) (*ResolvedComponent, error) {
	key := language + ":" + requestedVersion
	if comp, ok := m.runtimeComps[key]; ok {
		return comp, nil
	}
	return nil, errors.New("runtime not found")
}

func (m *mockProviderRegistry) ResolveFramework(language, framework, requestedVersion string) (*ResolvedComponent, error) {
	key := language + ":" + framework + ":" + requestedVersion
	if comp, ok := m.frameworkComps[key]; ok {
		return comp, nil
	}
	return nil, errors.New("framework not found")
}

func (m *mockProviderRegistry) ResolveService(serviceKind, name, requestedVersion string) (*ResolvedComponent, error) {
	return nil, errors.New("service not found")
}

func (m *mockProviderRegistry) ResolvePackageManager(name, requestedVersion string) (*ResolvedComponent, error) {
	return nil, errors.New("package manager not found")
}

func (m *mockProviderRegistry) ResolveToolchain(language string) ([]ResolvedToolchainComponent, error) {
	return []ResolvedToolchainComponent{}, nil
}

func (m *mockProviderRegistry) AvailableRuntimes(language string) []string {
	return m.availableRuntimes[language]
}

func standardTestIntent(language, framework string, fwVer *string, rtVer string) *TargetEnvironmentDocument {
	return &TargetEnvironmentDocument{
		SchemaVersion:  "2.0",
		Language:       language,
		Framework:      framework,
		Version:        fwVer,
		Architecture:   TargetArchitecture{Style: "modular_monolith"},
		DatabasePolicy: "preserve_existing",
		Path:           "target",
		Runtime: TargetRuntime{
			Version:      rtVer,
			OS:           "linux",
			Architecture: "amd64",
		},
		Deployment: TargetDeployment{
			Model:      "container",
			ProfileRef: "sha256:" + strings.Repeat("0", 64),
		},
	}
}

// TestTargetPrecedence_RootVsSplitConflict tests that conflicting root and split target configurations
// are rejected with ERR_CONFIG_CONFLICT, requiring explicit user resolution.
func TestTargetPrecedence_RootVsSplitConflict(t *testing.T) {
	fw1 := "11.0.0"
	fw2 := "10.0.0"
	root := standardTestIntent("php", "laravel", &fw1, "8.3.1")
	split := standardTestIntent("php", "laravel", &fw2, "8.3.1")

	inputs := &TargetPrecedenceInputs{
		WorkspaceRootTarget: root,
		SplitTarget:         split,
	}

	_, err := EvaluateTargetPrecedence(inputs)
	if err == nil {
		t.Fatal("expected conflict error between root and split targets, got nil")
	}
	if !strings.Contains(err.Error(), "ERR_CONFIG_CONFLICT") {
		t.Fatalf("expected ERR_CONFIG_CONFLICT in error, got: %v", err)
	}
}

// TestTargetPrecedence_HierarchyOrder tests that user intent overrides project defaults,
// detected suggestions, and recommendations.
func TestTargetPrecedence_HierarchyOrder(t *testing.T) {
	fwUser := "11.0.0"
	userIntent := standardTestIntent("php", "laravel", &fwUser, "8.3.1")

	fwProj := "10.0.0"
	projDefaults := standardTestIntent("php", "laravel", &fwProj, "8.2.0")

	inputs := &TargetPrecedenceInputs{
		ExplicitUserIntent: userIntent,
		ProjectDefaults:    projDefaults,
		Recommendations: []TargetRecommendationInput{
			{
				RecommendationID: "rec_use_symfony",
				Class:            "RECOMMENDED",
				Language:         "php",
				Framework:        "symfony",
				RuntimeVersion:   "8.3.1",
				AcceptedByUser:   false,
			},
		},
	}

	evaluated, err := EvaluateTargetPrecedence(inputs)
	if err != nil {
		t.Fatalf("EvaluateTargetPrecedence failed: %v", err)
	}

	if evaluated.Framework != "laravel" || evaluated.Runtime.Version != "8.3.1" {
		t.Fatalf("expected user intent to win, got framework %s version %s", evaluated.Framework, evaluated.Runtime.Version)
	}
}

// TestTargetPrecedence_HardEngineConstraints tests that hard engine constraints reject
// incompatible choices and NEVER silently replace or downgrade them.
func TestTargetPrecedence_HardEngineConstraints(t *testing.T) {
	fw := "1.0.0"
	userIntent := standardTestIntent("rust", "actix-web", &fw, "1.75.0")

	inputs := &TargetPrecedenceInputs{
		ExplicitUserIntent: userIntent,
		HardConstraints: &HardEngineConstraints{
			SupportedLanguages: map[string]struct{}{
				"php":        {},
				"go":         {},
				"typescript": {},
			},
		},
	}

	_, err := EvaluateTargetPrecedence(inputs)
	if err == nil {
		t.Fatal("expected hard constraint error for unsupported language, got nil")
	}
	if !strings.Contains(err.Error(), "hard constraint violation") {
		t.Fatalf("expected hard constraint violation, got: %v", err)
	}
}

// TestTargetPrecedence_RecommendationsNotSilentlyApplied verifies that unaccepted recommendations
// cannot silently override user choices.
func TestTargetPrecedence_RecommendationsNotSilentlyApplied(t *testing.T) {
	fwUser := "none"
	userIntent := standardTestIntent("php", fwUser, nil, "8.3.1")

	inputs := &TargetPrecedenceInputs{
		ExplicitUserIntent: userIntent,
		Recommendations: []TargetRecommendationInput{
			{
				RecommendationID: "rec_use_laravel",
				Class:            "RECOMMENDED",
				Language:         "php",
				Framework:        "laravel",
				RuntimeVersion:   "8.3.1",
				AcceptedByUser:   false,
			},
		},
	}

	evaluated, err := EvaluateTargetPrecedence(inputs)
	if err != nil {
		t.Fatalf("EvaluateTargetPrecedence failed: %v", err)
	}

	if evaluated.Framework != "none" {
		t.Fatalf("recommendation was silently applied: expected framework none, got %s", evaluated.Framework)
	}
}

// TestTargetResolution_X15_ExactResolutionInvariants tests Test X15 invariants:
// - 13.* wildcard rejected as resolution
// - latest mutable selector rejected as resolution
// - ^8.3 range rejected as resolution
// - 8.3 two-digit version rejected in schema 2.0
// - standard-library framework=null supported when framework="none"
// - missing runtime/distribution rejected
func TestTargetResolution_X15_ExactResolutionInvariants(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)

	// Subtest 1: Wildcard version 13.* rejected
	t.Run("X15_RejectWildcard13", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "13.*")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.0",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}
		_, err := ResolveTargetEnvironment(req)
		if err == nil {
			t.Fatal("expected wildcard 13.* to be rejected as exact resolution (Test X15), got nil")
		}
		if !strings.Contains(err.Error(), "wildcard") {
			t.Fatalf("expected wildcard rejection message, got: %v", err)
		}
	})

	// Subtest 2: Mutable selector "latest" rejected
	t.Run("X15_RejectMutableLatest", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "latest")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.0",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}
		_, err := ResolveTargetEnvironment(req)
		if err == nil {
			t.Fatal("expected mutable selector 'latest' to be rejected as exact resolution (Test X15), got nil")
		}
		if !strings.Contains(err.Error(), "mutable selector") {
			t.Fatalf("expected mutable selector rejection message, got: %v", err)
		}
	})

	// Subtest 3: Range constraint "^8.3" rejected
	t.Run("X15_RejectRangeConstraint", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "^8.3")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.0",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}
		_, err := ResolveTargetEnvironment(req)
		if err == nil {
			t.Fatal("expected range constraint '^8.3' to be rejected as exact resolution (Test X15), got nil")
		}
		if !strings.Contains(err.Error(), "range") {
			t.Fatalf("expected range constraint rejection message, got: %v", err)
		}
	})

	// Subtest 4: Non-SemVer version "8.3" rejected in schema 2.0
	t.Run("X15_RejectNonSemVerIn20", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "8.3")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.0",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}
		_, err := ResolveTargetEnvironment(req)
		if err == nil {
			t.Fatal("expected 2-digit version '8.3' to be rejected in schema 2.0 (Test X15), got nil")
		}
		if !strings.Contains(err.Error(), "exact SemVer") {
			t.Fatalf("expected exact SemVer message, got: %v", err)
		}
	})

	// Subtest 5: Standard-library framework="none" resolves to framework=null
	t.Run("X15_StandardLibraryFrameworkNull", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "8.3.1")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.0",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}
		res, err := ResolveTargetEnvironment(req)
		if err != nil {
			t.Fatalf("ResolveTargetEnvironment failed: %v", err)
		}
		if res.Document.Framework != nil {
			t.Fatalf("expected framework to be null for standard-library framework=none, got: %+v", res.Document.Framework)
		}
		// Confirm canonical JSON contains framework: null
		if !strings.Contains(string(res.CanonicalBytes), `"framework":null`) {
			t.Fatalf("expected canonical JSON to contain '\"framework\":null', got: %s", string(res.CanonicalBytes))
		}
	})
}

// TestTargetResolution_X75_OpaqueExactReleaseInvariants tests Test X75 invariants:
// - Digest-pinned date/vendor/git component release accepted under schema 2.1
// - Mutable latest selector rejected
// - Missing provider_id or interpretation_ref in 2.1 opaque release rejected
// - Digest identifies bytes; provider interprets release
func TestTargetResolution_X75_OpaqueExactReleaseInvariants(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("b", 64)
	distRef := "sha256:" + strings.Repeat("c", 64)
	interpRef := "sha256:" + strings.Repeat("d", 64)

	// Subtest 1: Opaque date release (release-2026-10-03) accepted in 2.1 with provider
	t.Run("X75_OpaqueDateReleaseAcceptedIn21", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "release-2026-10-03")
		reg := &mockProviderRegistry{
			availableRuntimes: map[string][]string{"php": {"release-2026-10-03"}},
			runtimeComps: map[string]*ResolvedComponent{
				"php:release-2026-10-03": {
					ProviderID:        "canonical_php_provider",
					Name:              "php",
					Version:           "release-2026-10-03",
					DistributionRef:   distRef,
					InterpretationRef: interpRef,
				},
			},
		}

		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.1",
			ProviderRegistry:     reg,
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}

		res, err := ResolveTargetEnvironment(req)
		if err != nil {
			t.Fatalf("expected opaque release to be accepted under schema 2.1 (Test X75), got: %v", err)
		}
		if res.Document.Runtime.Version != "release-2026-10-03" {
			t.Fatalf("expected version release-2026-10-03, got %s", res.Document.Runtime.Version)
		}
		if res.Document.Runtime.ProviderID != "canonical_php_provider" {
			t.Fatalf("expected provider canonical_php_provider, got %s", res.Document.Runtime.ProviderID)
		}
	})

	// Subtest 2: Opaque release missing provider_id or interpretation_ref in 2.1 rejected
	t.Run("X75_OpaqueReleaseMissingProviderRejected", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "release-2026-10-03")
		reg := &mockProviderRegistry{
			availableRuntimes: map[string][]string{"php": {"release-2026-10-03"}},
			runtimeComps: map[string]*ResolvedComponent{
				"php:release-2026-10-03": {
					ProviderID:      "", // Missing provider_id
					Name:            "php",
					Version:         "release-2026-10-03",
					DistributionRef: distRef,
				},
			},
		}

		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.1",
			ProviderRegistry:     reg,
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}

		_, err := ResolveTargetEnvironment(req)
		if err == nil {
			t.Fatal("expected opaque release without provider_id to be rejected (Test X75), got nil")
		}
		if !strings.Contains(err.Error(), "provider_id") {
			t.Fatalf("expected provider_id error, got: %v", err)
		}
	})

	// Subtest 3: Mutable selector "latest" rejected even in 2.1
	t.Run("X75_MutableSelectorRejectedIn21", func(t *testing.T) {
		intent := standardTestIntent("php", "none", nil, "latest")
		req := &TargetResolutionRequest{
			ResolutionID:         "res_test",
			ProjectID:            "proj_test",
			TargetIntent:         intent,
			IntentRef:            dummyRef,
			SchemaVersion:        "2.1",
			BuildRef:             dummyRef,
			DeploymentProfileRef: dummyRef,
			LockRef:              dummyRef,
			FeasibilityRef:       dummyRef,
		}

		_, err := ResolveTargetEnvironment(req)
		if err == nil {
			t.Fatal("expected mutable selector 'latest' to be rejected in 2.1 (Test X75), got nil")
		}
	})
}

// TestTargetResolution_DowngradePrevention verifies that when a user requests PHP 8.4 and only 8.3 is available,
// the resolver blocks with ResolutionBlockedError and NEVER silently downgrades to 8.3.
func TestTargetResolution_DowngradePrevention(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("e", 64)
	intent := standardTestIntent("php", "none", nil, "8.4.0")

	// Registry only has PHP 8.3.1
	reg := &mockProviderRegistry{
		availableRuntimes: map[string][]string{
			"php": {"8.3.1"},
		},
		runtimeComps: map[string]*ResolvedComponent{
			"php:8.3.1": {
				Name:            "php",
				Version:         "8.3.1",
				DistributionRef: dummyRef,
			},
		},
	}

	req := &TargetResolutionRequest{
		ResolutionID:         "res_test",
		ProjectID:            "proj_test",
		TargetIntent:         intent,
		IntentRef:            dummyRef,
		SchemaVersion:        "2.0",
		ProviderRegistry:     reg,
		BuildRef:             dummyRef,
		DeploymentProfileRef: dummyRef,
		LockRef:              dummyRef,
		FeasibilityRef:       dummyRef,
	}

	_, err := ResolveTargetEnvironment(req)
	if err == nil {
		t.Fatal("expected downgrade condition to block resolution, got nil")
	}

	var blockedErr *ResolutionBlockedError
	if !errors.As(err, &blockedErr) {
		t.Fatalf("expected ResolutionBlockedError, got %T: %v", err, err)
	}

	if blockedErr.RequestedVersion != "8.4.0" {
		t.Fatalf("expected requested version 8.4.0, got: %s", blockedErr.RequestedVersion)
	}

	if len(blockedErr.AvailableAlternatives) != 1 || blockedErr.AvailableAlternatives[0] != "8.3.1" {
		t.Fatalf("expected alternatives [8.3.1], got: %v", blockedErr.AvailableAlternatives)
	}
}
