package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCLIDispatcher_ProtocolAndInvariants(t *testing.T) {
	fencingSvc := NewResourceFencingService()
	incidentSvc := NewIncidentRepairService()
	dispatcher := NewGuardedCLIDispatcher(fencingSvc, incidentSvc)

	// =========================================================================
	// Subtest 1: Strict Flag Parsing
	// =========================================================================
	t.Run("StrictFlagParsing", func(t *testing.T) {
		// Valid invocation
		parsed, err := ParseCLIFlags([]string{"init", "/path/to/legacy", "--project", "test_proj", "--json", "--dry-run"})
		if err != nil {
			t.Fatalf("unexpected flag parsing error: %v", err)
		}
		if parsed.Command != "init" {
			t.Fatalf("expected command init, got %s", parsed.Command)
		}
		if parsed.Project != "test_proj" {
			t.Fatalf("expected project test_proj, got %s", parsed.Project)
		}
		if !parsed.JSON || !parsed.DryRun {
			t.Fatalf("expected JSON and DryRun flags to be true")
		}
		if len(parsed.Positionals) != 1 || parsed.Positionals[0] != "/path/to/legacy" {
			t.Fatalf("unexpected positionals: %v", parsed.Positionals)
		}

		// Rejects --quiet combined with --verbose
		_, err = ParseCLIFlags([]string{"doctor", "--quiet", "--verbose"})
		if err == nil || !strings.Contains(err.Error(), "cannot specify both --quiet and --verbose") {
			t.Fatalf("expected error combining --quiet and --verbose, got %v", err)
		}

		// Rejects duplicate scalar flag
		_, err = ParseCLIFlags([]string{"init", "--project", "p1", "--project", "p2"})
		if err == nil || !strings.Contains(err.Error(), "duplicate scalar flag") {
			t.Fatalf("expected error on duplicate scalar flag, got %v", err)
		}

		// Rejects unknown short flag
		_, err = ParseCLIFlags([]string{"init", "-x"})
		if err == nil || !strings.Contains(err.Error(), "unknown short flag") {
			t.Fatalf("expected error on unknown short flag, got %v", err)
		}
	})

	// =========================================================================
	// Subtest 2: Output Envelope Schema Validity
	// =========================================================================
	t.Run("EnvelopeSchemaValidity", func(t *testing.T) {
		// Test success envelope
		successEnv, err := NewCLISuccessEnvelope("init", map[string]any{"project_id": "test"}, 12, false, nil)
		if err != nil {
			t.Fatalf("NewCLISuccessEnvelope failed: %v", err)
		}
		successBytes, err := successEnv.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}
		if err := ValidateCLIOutput(successBytes); err != nil {
			t.Fatalf("generated success envelope failed schema validation: %v", err)
		}

		// Test failure envelope
		failErrors := []CLIStructuredError{
			{
				Code:              "ERR_INVALID_ARGUMENT",
				Message:           "invalid manifest reference",
				Category:          "USER_ERROR",
				Retryable:         false,
				RecommendedAction: "check manifest hash",
			},
		}
		failEnv, err := NewCLIFailureEnvelope("analyze", ExitCodeMissingGate1OrGate2, failErrors, map[string]any{}, 5, false)
		if err != nil {
			t.Fatalf("NewCLIFailureEnvelope failed: %v", err)
		}
		failBytes, err := failEnv.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}
		if err := ValidateCLIOutput(failBytes); err != nil {
			t.Fatalf("generated failure envelope failed schema validation: %v", err)
		}
	})

	// =========================================================================
	// Subtest 3: Gate Authorization Boundaries
	// =========================================================================
	t.Run("GateAuthorizationBoundaries", func(t *testing.T) {
		projectID := "gate_test_proj"
		dispatcher.SetGateStatus(projectID, false, false, false)

		// 1. Analyze requires Gate 1; without Gate 1 it exits 4
		analyzeFlags := &ParsedCLIFlags{
			Command: "analyze",
			Project: projectID,
			JSON:    true,
		}
		env, err := dispatcher.Dispatch(analyzeFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeMissingGate1OrGate2 {
			t.Fatalf("expected exit code 4 for missing Gate 1, got %d", env.ExitCode)
		}
		if env.Success {
			t.Fatalf("expected success=false for missing Gate 1")
		}

		// Confirm Gate 1
		confirmSourceFlags := &ParsedCLIFlags{
			Command: "confirm source",
			Project: projectID,
			JSON:    true,
		}
		env, err = dispatcher.Dispatch(confirmSourceFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected success confirming Gate 1, got exit code %d", env.ExitCode)
		}

		// Analyze now succeeds under Gate 1
		env, err = dispatcher.Dispatch(analyzeFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected success on analyze under Gate 1, got exit code %d", env.ExitCode)
		}

		// 2. Migrate requires Gate 2; without Gate 2 it exits 4
		migrateFlags := &ParsedCLIFlags{
			Command: "migrate",
			Project: projectID,
			JSON:    true,
		}
		env, err = dispatcher.Dispatch(migrateFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeMissingGate1OrGate2 {
			t.Fatalf("expected exit code 4 for missing Gate 2 on migrate, got %d", env.ExitCode)
		}

		// Confirm Gate 2
		confirmPlanFlags := &ParsedCLIFlags{
			Command: "confirm plan",
			Project: projectID,
			JSON:    true,
		}
		env, err = dispatcher.Dispatch(confirmPlanFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected success confirming Gate 2, got exit code %d", env.ExitCode)
		}

		// Migrate now succeeds under Gate 2
		env, err = dispatcher.Dispatch(migrateFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected success on migrate under Gate 2, got exit code %d", env.ExitCode)
		}

		// 3. Cutover prepare requires Gate 3; without Gate 3 it exits 3
		cutoverFlags := &ParsedCLIFlags{
			Command: "cutover prepare",
			Project: projectID,
			JSON:    true,
		}
		env, err = dispatcher.Dispatch(cutoverFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeMissingCutoverAuth {
			t.Fatalf("expected exit code 3 for missing Gate 3, got %d", env.ExitCode)
		}

		// Confirm Gate 3
		confirmCutoverFlags := &ParsedCLIFlags{
			Command: "confirm cutover",
			Project: projectID,
			JSON:    true,
		}
		env, err = dispatcher.Dispatch(confirmCutoverFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected success confirming Gate 3, got exit code %d", env.ExitCode)
		}

		// Cutover prepare now succeeds
		env, err = dispatcher.Dispatch(cutoverFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected success on cutover prepare, got exit code %d", env.ExitCode)
		}
	})

	// =========================================================================
	// Subtest 4: Dry-Run Safety
	// =========================================================================
	t.Run("DryRunSafety", func(t *testing.T) {
		projectID := "dry_run_proj"
		dispatcher.SetGateStatus(projectID, true, true, false)

		flags := &ParsedCLIFlags{
			Command: "migrate",
			Project: projectID,
			JSON:    true,
			DryRun:  true,
		}
		env, err := dispatcher.Dispatch(flags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if env.ExitCode != ExitCodeSuccess || !env.Success {
			t.Fatalf("expected exit 0 for dry-run")
		}
		if env.DryRun == nil || !*env.DryRun {
			t.Fatalf("expected envelope dry_run to be true")
		}
		if env.DryRunSummary == nil || env.DryRunSummary["simulated_only"] != true {
			t.Fatalf("expected dry_run_summary with simulated_only=true, got %v", env.DryRunSummary)
		}

		raw, err := env.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}
		if err := ValidateCLIOutput(raw); err != nil {
			t.Fatalf("dry-run envelope validation failed: %v", err)
		}

		// Verify dry-run output contains no unexpected fields
		var parsed map[string]any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if parsed["dry_run"] != true {
			t.Fatalf("expected dry_run in JSON")
		}
	})
}
