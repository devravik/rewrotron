package domain

import (
	"context"
	"strings"
	"testing"
)

func TestSkillLifecycle_ContentAddressingAndTransitions(t *testing.T) {
	ctx := context.Background()
	svc := NewSkillRegistryService(nil)

	inv1 := &SkillPackageInventory{
		Identity: "rwt-skill-package-1",
		Entries: []SkillPackageEntry{
			{Path: "SKILL.md", SizeBytes: 120, SHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
			{Path: "references/rules.md", SizeBytes: 340, SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"},
		},
	}

	// 1. Content addressing reproducibility
	d1, _, err1 := inv1.Digest()
	if err1 != nil {
		t.Fatalf("unexpected digest error: %v", err1)
	}

	// Permute entries order; digest must remain byte-identical
	inv2 := &SkillPackageInventory{
		Identity: "rwt-skill-package-1",
		Entries: []SkillPackageEntry{
			{Path: "references/rules.md", SizeBytes: 340, SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"},
			{Path: "SKILL.md", SizeBytes: 120, SHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		},
	}
	d2, _, _ := inv2.Digest()
	if d1 != d2 {
		t.Fatalf("expected identical digest for permuted inventory entries, got %s and %s", d1, d2)
	}

	// 2. Register proposed skill
	cap, err := svc.RegisterProposedSkill(ctx, "skill-laravel-routes", "Laravel Routes Lowering", "Generates Laravel route definitions", "1.0.0", inv1, []string{"routes", "laravel"}, nil)
	if err != nil {
		t.Fatalf("failed to register skill: %v", err)
	}
	if cap.Status != SkillStatusProposed {
		t.Fatalf("expected initial status PROPOSED, got %s", cap.Status)
	}

	// 3. Proposed skills are not automatically discoverable during standard execution
	res := svc.SearchCapabilities("routes", false)
	if len(res) != 0 {
		t.Fatalf("expected 0 discoverable capabilities for unapproved PROPOSED skill, got %d", len(res))
	}

	// Inspectable when non-stable requested
	resAll := svc.SearchCapabilities("routes", true)
	if len(resAll) != 1 {
		t.Fatalf("expected 1 capability when inspecting non-stable, got %d", len(resAll))
	}

	// 4. Transition PROPOSED -> EXPERIMENTAL
	err = svc.TransitionStatus(ctx, "skill-laravel-routes", "1.0.0", SkillStatusExperimental, "")
	if err != nil {
		t.Fatalf("failed transition to EXPERIMENTAL: %v", err)
	}

	// 5. Agent promotion to STABLE without human token is rejected
	err = svc.TransitionStatus(ctx, "skill-laravel-routes", "1.0.0", SkillStatusStable, "")
	if err == nil || !strings.Contains(err.Error(), "AGENT_AUTHORITY_VIOLATION") {
		t.Fatalf("expected AGENT_AUTHORITY_VIOLATION promoting to STABLE without human token, got: %v", err)
	}

	// 6. Valid human promotion to STABLE succeeds
	err = svc.TransitionStatus(ctx, "skill-laravel-routes", "1.0.0", SkillStatusStable, "tok-human-gate-token-42")
	if err != nil {
		t.Fatalf("failed transition to STABLE with human token: %v", err)
	}

	// 7. Now automatically discoverable
	resStable := svc.SearchCapabilities("routes", false)
	if len(resStable) != 1 {
		t.Fatalf("expected 1 discoverable capability for STABLE skill, got %d", len(resStable))
	}

	// 8. Deprecation
	err = svc.TransitionStatus(ctx, "skill-laravel-routes", "1.0.0", SkillStatusDeprecated, "")
	if err != nil {
		t.Fatalf("failed deprecation: %v", err)
	}
	resDeprecated := svc.SearchCapabilities("routes", false)
	if len(resDeprecated) != 0 {
		t.Fatalf("expected 0 discoverable capabilities after deprecation, got %d", len(resDeprecated))
	}
}

func TestSkillLifecycle_X38_DegradedToolRejected(t *testing.T) {
	// Good tool with complete capture
	goodTool := &ToolQualityMetrics{
		ToolName:              "pdo-trace-extractor",
		CaptureCoverage:       1.0,
		TraceDroppedCount:     0,
		UnknownRate:           0.0,
		MutationDetectionRate: 1.0,
		ApparentSuccessRate:   0.95,
	}
	if err := ValidateToolEvidenceQuality(goodTool); err != nil {
		t.Fatalf("expected good tool to pass, got: %v", err)
	}

	// Degraded tool that drops hard PDO traces to improve apparent success rate to 100% (X38)
	degradedTool := &ToolQualityMetrics{
		ToolName:              "pdo-trace-extractor-degraded",
		CaptureCoverage:       0.85,
		TraceDroppedCount:     14,
		UnknownRate:           0.0,
		MutationDetectionRate: 1.0,
		ApparentSuccessRate:   1.0, // Apparent success improved!
	}
	err := ValidateToolEvidenceQuality(degradedTool)
	if err == nil {
		t.Fatalf("expected degraded tool dropping traces to be rejected despite 100%% apparent success")
	}
	if !strings.Contains(err.Error(), "EVIDENCE_QUALITY_REJECTION") || !strings.Contains(err.Error(), "dropped 14 capture/trace records") {
		t.Fatalf("expected EVIDENCE_QUALITY_REJECTION mentioning dropped traces, got: %v", err)
	}
}
