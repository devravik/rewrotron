package domain

import (
	"errors"
	"strings"
	"testing"
)

func standardDepRequest(depID string, strat DecisionStrategy) *DependencyResolutionRequest {
	dummyRef := "sha256:" + strings.Repeat("a", 64)
	candPkg := "guzzlehttp/guzzle"
	candVer := "7.8.0"

	return &DependencyResolutionRequest{
		DependencyID:   depID,
		TargetSpecHash: strings.Repeat("b", 64),
		Source: SourceDependencyInfo{
			Ecosystem:    "composer",
			Package:      "legacy/http-client",
			Version:      "1.2.0",
			EvidenceRefs: []string{dummyRef},
		},
		Usage: DependencyUsage{
			Symbols:              []string{"LegacyHttpClient::get"},
			SourceFiles:          []string{"application/controllers/Api.php"},
			AffectedUnitIDs:      []string{"unit_api"},
			RequiredCapabilities: []string{"cap_http_request"},
		},
		CandidateStrategy: strat,
		CandidatePackage:  &candPkg,
		CandidateVersion:  &candVer,
		DecisionRef:       dummyRef,
		TargetSupportedExts: map[string]struct{}{
			"curl":     {},
			"json":     {},
			"mbstring": {},
		},
		TargetPermittedLicenses: map[string]struct{}{
			"MIT":          {},
			"Apache-2.0":   {},
			"BSD-3-Clause": {},
		},
		TransitiveDependencies: []TransitiveDependencyInfo{
			{
				PackageID:          "dep_psr_http_client",
				RequiredExtensions: []string{},
				License:            "MIT",
			},
		},
		KnownBehaviors: map[string]string{
			"cap_http_request": "PASS",
		},
	}
}

// TestResolveDependency_HappyPath tests basic successful dependency resolution conforming to schema 2.0.
func TestResolveDependency_HappyPath(t *testing.T) {
	req := standardDepRequest("dep_http", StrategyReplacement)
	doc, err := ResolveDependency(req)
	if err != nil {
		t.Fatalf("ResolveDependency failed: %v", err)
	}

	if doc.Proposal.Strategy != "REPLACEMENT" {
		t.Fatalf("expected strategy REPLACEMENT, got %s", doc.Proposal.Strategy)
	}
	if doc.Compatibility.Behavior != "PASS" {
		t.Fatalf("expected PASS behavior, got %s", doc.Compatibility.Behavior)
	}
	if len(doc.CapabilityMappings) != 1 || doc.CapabilityMappings[0].Status != "COVERED" {
		t.Fatalf("expected 1 COVERED mapping, got: %+v", doc.CapabilityMappings)
	}
}

// TestResolveDependency_X18_IncompatibleTransitiveExtensionBlocked tests Test X18 invariant:
// A replacement that introduces a transitive dependency requiring an unavailable native extension
// is BLOCKED, not upgraded blindly.
func TestResolveDependency_X18_IncompatibleTransitiveExtensionBlocked(t *testing.T) {
	req := standardDepRequest("dep_http_native_blocked", StrategyReplacement)
	// Add transitive dependency requiring ext-sodium which is not in TargetSupportedExts
	req.TransitiveDependencies = append(req.TransitiveDependencies, TransitiveDependencyInfo{
		PackageID:          "paragonie/sodium_compat",
		RequiredExtensions: []string{"sodium"}, // "sodium" is unavailable
		License:            "MIT",
	})

	_, err := ResolveDependency(req)
	if err == nil {
		t.Fatal("expected error due to missing native extension in transitive dependency (Test X18), got nil")
	}

	var blockedErr *DependencyResolutionBlockedError
	if !errors.As(err, &blockedErr) {
		t.Fatalf("expected DependencyResolutionBlockedError, got %T: %v", err, err)
	}

	if blockedErr.BlockerType != "NATIVE_EXTENSION_MISSING" {
		t.Fatalf("expected blocker type NATIVE_EXTENSION_MISSING, got: %s", blockedErr.BlockerType)
	}
	if !strings.Contains(blockedErr.Reason, "unavailable native extension") {
		t.Fatalf("unexpected reason: %s", blockedErr.Reason)
	}
}

// TestResolveDependency_X18_IncompatibleTransitiveLicenseBlocked tests Test X18 invariant:
// A replacement with an incompatible transitive license is BLOCKED.
func TestResolveDependency_X18_IncompatibleTransitiveLicenseBlocked(t *testing.T) {
	req := standardDepRequest("dep_license_blocked", StrategyReplacement)
	// Add transitive dependency with GPL-3.0 which is not permitted by project policy
	req.TransitiveDependencies = append(req.TransitiveDependencies, TransitiveDependencyInfo{
		PackageID: "copyleft/library",
		License:   "GPL-3.0",
	})

	_, err := ResolveDependency(req)
	if err == nil {
		t.Fatal("expected error due to license conflict (Test X18), got nil")
	}

	var blockedErr *DependencyResolutionBlockedError
	if !errors.As(err, &blockedErr) {
		t.Fatalf("expected DependencyResolutionBlockedError, got %T: %v", err, err)
	}

	if blockedErr.BlockerType != "LICENSE_CONFLICT" {
		t.Fatalf("expected blocker type LICENSE_CONFLICT, got: %s", blockedErr.BlockerType)
	}
}

// TestResolveDependency_X18_RetryPolicyBehaviorDifferenceBlocked tests Test X18 invariant:
// Apparent API compatibility that changes retries or timeout semantics must block/characterize
// rather than upgrade blindly.
func TestResolveDependency_X18_RetryPolicyBehaviorDifferenceBlocked(t *testing.T) {
	req := standardDepRequest("dep_retry_blocked", StrategyReplacement)
	req.TransitiveDependencies = append(req.TransitiveDependencies, TransitiveDependencyInfo{
		PackageID:          "guzzlehttp/retry-subscriber",
		License:            "MIT",
		RetryPolicyChanges: true, // Changes retry behavior
	})

	_, err := ResolveDependency(req)
	if err == nil {
		t.Fatal("expected error due to retry behavior change (Test X18), got nil")
	}

	var blockedErr *DependencyResolutionBlockedError
	if !errors.As(err, &blockedErr) {
		t.Fatalf("expected DependencyResolutionBlockedError, got %T: %v", err, err)
	}

	if blockedErr.BlockerType != "BEHAVIOR_DIFFERENCE" {
		t.Fatalf("expected blocker type BEHAVIOR_DIFFERENCE, got: %s", blockedErr.BlockerType)
	}
	if !strings.Contains(blockedErr.Reason, "apparent API compatibility changes retries") {
		t.Fatalf("expected Test X18 invariant in error, got: %s", blockedErr.Reason)
	}
}

// TestResolveDependency_X50_RefreshDependencyStrategy tests Test X50 invariant:
// Changing dependency strategy from REPLACEMENT to ADAPTER refreshes affected findings and planning inputs,
// updates the resolution digest, and marks the old decision ref as invalidated.
func TestResolveDependency_X50_RefreshDependencyStrategy(t *testing.T) {
	req := standardDepRequest("dep_refreshable", StrategyReplacement)
	initialDoc, err := ResolveDependency(req)
	if err != nil {
		t.Fatalf("initial ResolveDependency failed: %v", err)
	}

	initialDigest, err := initialDoc.Digest()
	if err != nil {
		t.Fatalf("initialDoc Digest failed: %v", err)
	}

	// User changes strategy to ADAPTER with new decision ref
	newDecisionRef := "sha256:" + strings.Repeat("9", 64)
	affectedFindings := []string{"find_http_client_adapter_required"}

	refreshed, err := RefreshDependencyStrategy(
		initialDoc,
		StrategyAdapter,
		newDecisionRef,
		affectedFindings,
	)
	if err != nil {
		t.Fatalf("RefreshDependencyStrategy failed: %v", err)
	}

	// Verify strategy updated to ADAPTER
	if refreshed.Document.Proposal.Strategy != "ADAPTER" {
		t.Fatalf("expected refreshed strategy ADAPTER, got %s", refreshed.Document.Proposal.Strategy)
	}
	if !refreshed.Document.Proposal.AdapterRequired {
		t.Fatal("expected AdapterRequired to be true")
	}

	// Verify digests differ
	if refreshed.NewDigest == initialDigest {
		t.Fatal("expected new resolution digest after strategy change")
	}
	if refreshed.PreviousDigest != initialDigest {
		t.Fatalf("expected previous digest %s, got %s", initialDigest, refreshed.PreviousDigest)
	}

	// Verify old decision ref is identified as invalidated
	if refreshed.InvalidatedOldDecisionRef != initialDoc.DecisionRef {
		t.Fatalf("expected invalidated old decision ref %s, got %s", initialDoc.DecisionRef, refreshed.InvalidatedOldDecisionRef)
	}
	if len(refreshed.RefreshedFindings) != 1 || refreshed.RefreshedFindings[0] != "find_http_client_adapter_required" {
		t.Fatalf("expected refreshed findings, got: %v", refreshed.RefreshedFindings)
	}
}
