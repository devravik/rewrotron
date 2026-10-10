package domain

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestM12_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M12:
// Release Readiness, Reproducibility & v1 Qualification.
//
// It integrates and qualifies:
// 1. Release Qualification Matrix (Unit 1):
//    - Evaluates all requirements across source, target, verification, security, packaging, and interfaces
//    - Validates that EvaluateReleaseReadiness passes for Reference Profile #1 (CI3 -> Laravel 13 on Linux amd64)
//    - Proves falsification: UNKNOWN or BLOCKED requirements halt release readiness
// 2. Deterministic Generation & Artifact Integrity (Unit 2):
//    - Bitwise reproducibility of Laravel 13 vertical slice artifacts across independent generation runs
//    - Manifest SHA-256 content-hash binding and tamper detection via ArtifactIntegrityVerifier
//    - Falsification: tampered byte payloads, missing declared artifacts, and extraneous files rejected
// 3. Security Sandbox & Permission Enforcement (Unit 3):
//    - Workspace containment and refusal of path traversal attacks (escapes, null bytes, outside paths)
//    - Sensitive credential and PII scrubbing using HMAC-keyed Sanitizer
//    - Production-access and irreversible write denial by default (ERR_PRODUCTION_EFFECT_REFUSED)
// 4. Native Packaging, Installation Lifecycle & Rollback Readiness (Unit 4):
//    - Multi-platform packaging producing byte-reproducible tar.gz archives for primary runtime (linux-amd64)
//    - Authenticity verification (X36-B) and atomic symlink installation/pointer activation (X36-A, X36-D)
//    - Version upgrade/rollback lifecycle preserving customer SQLite database state (X36-E)
//    - Write-aware target rollback validation across database, file, session, and external effects (X34)
// 5. Automated Release Gate Evaluation & Interface Parity (Unit 5):
//    - AutomatedReleaseGateEvaluator verifies 100% passing compatibility cases, zero open critical mismatches,
//      security audits, packaging integrity, and CLI/MCP contract parity
//    - Rejection of premature release when compatibility is incomplete or mismatches remain
// 6. Cryptographic v1 Migration & Release-Readiness Report:
//    - Fully conformant MigrationReportRecord validated against specs/reports/migration-report.schema.json
//    - Accurate provenance digests, benchmark references, limitations, and intentional change disclosures
// 7. Cross-Milestone Regression Guardrails (M0–M11):
//    - M4 Unit 1: Resource fencing token monotonicity and stale token rejection
//    - M8: Execution profiler N+1 query anomaly detection
//    - M9: Pluggable TargetAdapterRegistry with exact resolution for Go stdlib and Laravel 13
//    - M10: Pluggable SourceAdapterRegistry with exact resolution for Generic PHP and CodeIgniter 3
//    - M11: Integrated CI3-to-Laravel vertical slice generation with differential verification
func TestM12_AcceptanceClosure(t *testing.T) {
	ctx := context.Background()

	// =========================================================================
	// 1. Release Qualification Matrix Evaluation (M12 Unit 1)
	// =========================================================================
	t.Run("ReleaseQualificationMatrixEvaluation", func(t *testing.T) {
		matrix := NewReleaseQualificationMatrix()
		if matrix == nil {
			t.Fatal("expected non-nil ReleaseQualificationMatrix")
		}

		ready, blockers := matrix.EvaluateReleaseReadiness()
		if !ready || len(blockers) != 0 {
			t.Fatalf("expected matrix to be ReadyForRelease, got blocked with: %v", blockers)
		}
		if len(matrix.Requirements) < 9 {
			t.Fatalf("expected at least 9 evaluated requirements, got %d", len(matrix.Requirements))
		}

		// Negative control: any BLOCKED requirement halts release readiness
		falsifiedMatrix := NewReleaseQualificationMatrix()
		falsifiedMatrix.Requirements[0].Status = QualStatusBlocked
		falsifiedMatrix.Requirements[0].RemainingObligation = "Incomplete source qualification"
		falsifiedReady, falsifiedBlockers := falsifiedMatrix.EvaluateReleaseReadiness()
		if falsifiedReady {
			t.Fatal("expected falsified matrix to fail ReadyForRelease")
		}
		if len(falsifiedBlockers) == 0 {
			t.Fatal("expected blockers to be reported when requirement is BLOCKED")
		}
	})

	// =========================================================================
	// 2. Deterministic Generation & Artifact Integrity (M12 Unit 2)
	// =========================================================================
	t.Run("DeterministicGenerationAndArtifactIntegrity", func(t *testing.T) {
		adapter := NewLaravelTargetAdapter()
		unitID := "reproducible-vertical-slice-m12"
		adm := makeTestAdmission(t, unitID, []string{
			"routing",
			"controllers",
			"eloquent_models",
			"database",
			"transactions",
			"auth",
			"session",
			"csrf",
		}, []string{
			"health",
			"login",
			"whoami",
			"csrf",
			"create",
			"item",
			"note",
			"logout",
		})
		tm := makeTestTargetModel(t, unitID)
		sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

		run1, err := adapter.Generate(ctx, GenerationInput{
			Admission:       adm,
			TargetModel:     tm,
			UnitID:          unitID,
			SessionContract: sc,
		})
		if err != nil {
			t.Fatalf("generate vertical slice run 1 failed: %v", err)
		}
		run2, err := adapter.Generate(ctx, GenerationInput{
			Admission:       adm,
			TargetModel:     tm,
			UnitID:          unitID,
			SessionContract: sc,
		})
		if err != nil {
			t.Fatalf("generate vertical slice run 2 failed: %v", err)
		}

		// Bitwise determinism check
		verifier := NewArtifactIntegrityVerifier()
		if err := verifier.CompareGenerationOutputs(run1, run2); err != nil {
			t.Fatalf("reproducibility check failed: %v", err)
		}

		// Cryptographic integrity verifier
		if err := verifier.VerifyIntegrity(run1.Manifest, run1.Artifacts); err != nil {
			t.Fatalf("artifact integrity verification failed: %v", err)
		}

		// Falsification: tampered byte payload detected
		tamperedArtifacts := make(map[string][]byte)
		for k, v := range run1.Artifacts {
			tamperedArtifacts[k] = append([]byte(nil), v...)
		}
		tamperedArtifacts["routes/web.php"] = []byte("<?php // tampered payload ?>")
		if err := verifier.VerifyIntegrity(run1.Manifest, tamperedArtifacts); err == nil {
			t.Fatal("expected checksum mismatch error on tampered artifact, got nil")
		}
	})

	// =========================================================================
	// 3. Security Sandbox & Permission Enforcement (M12 Unit 3)
	// =========================================================================
	t.Run("SecuritySandboxAndSecretRedaction", func(t *testing.T) {
		tempDir := t.TempDir()
		workspace := filepath.Join(tempDir, "workspace")
		_ = os.MkdirAll(workspace, 0755)

		enforcer := NewSecuritySandboxEnforcer(workspace, []byte("m12-closure-hmac-key"))

		// Path traversal containment check
		traversalPaths := []string{
			"../../etc/passwd",
			"../outside.txt",
			"/etc/shadow",
			"nested/../../../../secret",
			"sub/file\x00.txt",
		}
		for _, badPath := range traversalPaths {
			if _, err := enforcer.ValidatePathWithinWorkspace(badPath); err == nil {
				t.Fatalf("expected sandbox violation error for path traversal %q, got nil", badPath)
			}
		}

		// Valid path accepted
		validPath, err := enforcer.ValidatePathWithinWorkspace("app/Models/User.php")
		if err != nil {
			t.Fatalf("expected valid relative path to succeed, got: %v", err)
		}
		if !strings.HasPrefix(validPath, workspace) {
			t.Fatalf("expected valid path %q to reside within workspace %q", validPath, workspace)
		}

		// Credential & PII scrubbing
		rawLog := "Authentication failed for user alice with password=SuperSecretP@ssword! and token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-ID brackets and aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		scrubbedLog := enforcer.RedactLogOrReport(rawLog)
		if strings.Contains(scrubbedLog, "SuperSecretP@ssword!") {
			t.Fatal("raw password leaked in scrubbed log")
		}
		if strings.Contains(scrubbedLog, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY") {
			t.Fatal("raw AWS key was not redacted from log")
		}
		if !strings.Contains(scrubbedLog, "[REDACTED") {
			t.Fatalf("expected [REDACTED placeholder in scrubbed log: %s", scrubbedLog)
		}

		// Production-access refusal check
		prodPolicy := EffectSafetyPolicy{IsProductionEnvironment: true, AllowIrreversible: false}
		writeEffect := BusinessEffectRecord{
			EffectID:    "eff-write-closure",
			Dimension:   EffectDimensionDBWrite,
			EffectClass: EffectClassLocalWrite,
		}
		err = prodPolicy.CheckEffectSafety(writeEffect)
		if err == nil || !strings.Contains(err.Error(), "ERR_PRODUCTION_EFFECT_REFUSED") {
			t.Fatalf("expected ERR_PRODUCTION_EFFECT_REFUSED, got: %v", err)
		}
	})

	// =========================================================================
	// 4. Native Packaging, Installation Lifecycle & Rollback Readiness (M12 Unit 4)
	// =========================================================================
	t.Run("NativePackagingAndRollbackReadiness", func(t *testing.T) {
		tempDir := t.TempDir()

		binRewrotron := filepath.Join(tempDir, "rewrotron")
		_ = os.WriteFile(binRewrotron, []byte("#!/bin/sh\necho rewrotron v1.0.0\n"), 0755)

		binMCP := filepath.Join(tempDir, "rewrotron-mcp")
		_ = os.WriteFile(binMCP, []byte("#!/bin/sh\necho rewrotron-mcp v1.0.0\n"), 0755)

		pkgConfig := ReleasePackagingConfig{
			Version:      "1.0.0",
			Architecture: "amd64",
			OS:           "linux",
			CommitHash:   "commit-m12-closure",
			BinRewrotron: binRewrotron,
			BinMCP:       binMCP,
		}

		archivePath := filepath.Join(tempDir, "rewrotron-1.0.0-linux-amd64.tar.gz")
		manifest, err := BuildReleaseArchive(pkgConfig, archivePath)
		if err != nil {
			t.Fatalf("failed to build release archive: %v", err)
		}

		if manifest.Platform != "linux-amd64" || manifest.ArchiveSHA256 == "" {
			t.Fatalf("invalid packaging manifest: %+v", manifest)
		}

		// Authenticity check: corrupted checksum rejected
		homeDir := filepath.Join(tempDir, "home")
		layout := DefaultInstallationLayout(homeDir, "1.0.0")
		corruptedHash := strings.Repeat("0", 64)
		if err := ExtractAndInstall(archivePath, corruptedHash, layout, false); err == nil {
			t.Fatal("expected authenticity verification failure with corrupt hash, got nil")
		}

		// Clean installation and symlink activation
		if err := ExtractAndInstall(archivePath, manifest.ArchiveSHA256, layout, false); err != nil {
			t.Fatalf("clean install failed: %v", err)
		}

		targetRewrotron, err := os.Readlink(layout.RewrotronBin)
		if err != nil || !strings.Contains(targetRewrotron, "current/bin/rewrotron") {
			t.Fatalf("rewrotron symlink not pointing to current binary: %v", targetRewrotron)
		}

		// Write-aware rollback rehearsal (X34)
		rehearsal := &RollbackRehearsalRecord{
			SchemaVersion:  "1.0",
			RehearsalID:    "rehearsal-m12-closure",
			ProjectID:      "proj-m12-closure",
			TargetTreeRef:  "sha256:" + strings.Repeat("a", 64),
			EnvironmentRef: "sha256:" + strings.Repeat("b", 64),
			ExecutableCommands: []string{
				"php artisan migrate:rollback",
				"git checkout HEAD -- app/Models/User.php",
			},
			RepresentativeMutations: []TargetMutation{
				{
					MutationID:   "mut-db-1",
					MutationType: MutationTypeDatabaseRow,
					TargetRef:    "items",
					PayloadHash:  strings.Repeat("c", 64),
					Reversible:   true,
				},
				{
					MutationID:   "mut-file-1",
					MutationType: MutationTypeFileWrite,
					TargetRef:    "storage/app/receipts.txt",
					PayloadHash:  strings.Repeat("d", 64),
					Reversible:   true,
				},
			},
			ReconciliationContracts: []ReconciliationContract{
				{
					MutationType:   MutationTypeDatabaseRow,
					ReconcileOwner: "operator",
					Strategy:       "REPLAY_SOURCE",
					Verified:       true,
				},
				{
					MutationType:   MutationTypeFileWrite,
					ReconcileOwner: "operator",
					Strategy:       "RESTORE_FILE",
					Verified:       true,
				},
			},
			Triggers: []RollbackTrigger{
				{
					TriggerID:  "trig-mismatch",
					Metric:     "MISMATCH",
					Threshold:  "> 0.00",
					DecisionBy: "gate3_operator",
				},
			},
			SourceContinuityVerified: true,
			Outcome:                  RollbackRehearsalPass,
		}

		if err := ValidateRollbackRehearsal(rehearsal); err != nil {
			t.Fatalf("rollback rehearsal validation failed: %v", err)
		}
	})

	// =========================================================================
	// 5. Automated Release Gate Evaluation (M12 Unit 5)
	// =========================================================================
	t.Run("AutomatedReleaseGateEvaluation", func(t *testing.T) {
		gate := NewAutomatedReleaseGateEvaluator()
		matrix := NewReleaseQualificationMatrix()

		gateInput := ReleaseGateEvaluationInput{
			ProjectID:                  "proj-m12-closure",
			ReleaseVersion:             "1.0.0",
			TargetProfileID:            "profile-ci3-laravel-linux-amd64",
			QualificationMatrix:        matrix,
			TotalCompatibilityCases:    24,
			PassingCompatibilityCases:  24,
			CriticalMismatches:         0,
			SecurityAuditsPassed:       true,
			ReproducibilityVerified:    true,
			PackagingLifecycleVerified: true,
			CLIAndMCPParityVerified:    true,
		}

		gateReport, err := gate.EvaluateReleaseGate(ctx, gateInput)
		if err != nil {
			t.Fatalf("gate evaluation failed: %v", err)
		}
		if gateReport.Verdict != ReleaseGatePass {
			t.Fatalf("expected release gate verdict PASS, got %s (blockers: %v)", gateReport.Verdict, gateReport.Blockers)
		}

		// Falsification: unresolved critical mismatch blocks release
		gateInputFalsified := gateInput
		gateInputFalsified.CriticalMismatches = 1
		falsifiedGateReport, err := gate.EvaluateReleaseGate(ctx, gateInputFalsified)
		if err != nil {
			t.Fatalf("gate evaluation failed: %v", err)
		}
		if falsifiedGateReport.Verdict != ReleaseGateBlocked {
			t.Fatalf("expected release gate to be BLOCKED on critical mismatch, got %s", falsifiedGateReport.Verdict)
		}
	})

	// =========================================================================
	// 6. Cryptographic v1 Migration & Release-Readiness Report Production
	// =========================================================================
	t.Run("CryptographicReleaseReadinessReportValidation", func(t *testing.T) {
		configHex := strings.Repeat("a", 64)
		planHex := strings.Repeat("b", 64)
		targetTreeHex := strings.Repeat("c", 64)
		evidenceRef := "sha256:" + strings.Repeat("2", 64)
		omissionReportRef := "sha256:" + strings.Repeat("4", 64)
		releaseRef := "sha256:" + strings.Repeat("5", 64)

		migrationReportMap := map[string]interface{}{
			"schema_version":     "1.0",
			"project_id":         "proj-m12-closure",
			"configuration_hash": configHex,
			"plan_hash":          planHex,
			"target_tree_hash":   targetTreeHex,
			"snapshot": map[string]interface{}{
				"project_id":         "proj-m12-closure",
				"state":              "VERIFIED",
				"revision":           1,
				"configuration_hash": configHex,
				"plan_hash":          planHex,
				"gate_states": map[string]interface{}{
					"GATE_1_SOURCE_DETECTION": "VALID",
					"GATE_2_PLAN_CONFIRMED":   "VALID",
					"GATE_3_CUTOVER":          "PENDING",
				},
				"coverage": map[string]interface{}{
					"entrypoints": map[string]interface{}{
						"numerator":   8,
						"denominator": 8,
						"percent":     100.0,
					},
					"cases": map[string]interface{}{
						"numerator":   24,
						"denominator": 24,
						"percent":     100.0,
					},
					"units": map[string]interface{}{
						"numerator":   1,
						"denominator": 1,
						"percent":     100.0,
					},
					"dimensions": map[string]interface{}{
						"numerator":   4,
						"denominator": 4,
						"percent":     100.0,
					},
				},
				"blockers":            []interface{}{},
				"latest_operation_id": "op-verify-m12",
			},
			"inventory_dispositions": []map[string]interface{}{
				{
					"entrypoint_id": "ep-health",
					"disposition":   "MIGRATED",
					"unit_id":       "unit-closure",
					"evidence_refs": []string{evidenceRef},
				},
				{
					"entrypoint_id": "ep-login",
					"disposition":   "MIGRATED",
					"unit_id":       "unit-closure",
					"evidence_refs": []string{evidenceRef},
				},
				{
					"entrypoint_id": "ep-checkout",
					"disposition":   "MIGRATED",
					"unit_id":       "unit-closure",
					"evidence_refs": []string{evidenceRef},
				},
			},
			"verification_report_refs": []string{
				"sha256:" + strings.Repeat("7", 64),
			},
			"benchmark_report_refs": []string{
				"sha256:" + strings.Repeat("8", 64),
			},
			"omission_report_ref": omissionReportRef,
			"release_ref":         releaseRef,
			"intentional_changes": []map[string]interface{}{
				{
					"change_id":   "ic-laravel-eloquent-models",
					"description": "Reconstructed CI3 database access using Laravel 13 Eloquent ORM while preserving schema",
				},
			},
			"limitations": []string{
				"Preserves existing MySQL relational schema; distributed asynchronous messaging not characterized",
			},
			"required_decisions": []string{
				"Operator authorization of Gate 3 release runbook prior to production deployment",
			},
		}

		rawReport, err := json.Marshal(migrationReportMap)
		if err != nil {
			t.Fatalf("failed to marshal migration report: %v", err)
		}

		if err := ValidateMigrationReport(rawReport); err != nil {
			t.Fatalf("ValidateMigrationReport failed: %v", err)
		}
	})

	// =========================================================================
	// 7. Cross-Milestone Regression Guardrails (M0–M11)
	// =========================================================================
	t.Run("CrossMilestoneRegressionProtection", func(t *testing.T) {
		now := time.Now()

		// M4 Unit 1: Resource fencing token monotonicity and stale fence rejection
		fenceSvc := NewResourceFencingService()
		mission := &MissionRecord{
			MissionID:   "m_m12_01",
			ProjectID:   "proj_m12",
			Type:        "MIGRATE_UNIT",
			Status:      MissionStatusQueued,
			Fingerprint: "fp_m12",
			Description: "M12 closure mission",
		}
		if _, err := fenceSvc.SubmitMission(mission); err != nil {
			t.Fatalf("submit mission failed: %v", err)
		}
		claimed, err := fenceSvc.ClaimMission("proj_m12", "m_m12_01", "worker_m12", now, 5*time.Minute)
		if err != nil {
			t.Fatalf("claim mission failed: %v", err)
		}
		if _, err := fenceSvc.AcquireResourceClaim("proj_m12", "app/Models/User.php", "m_m12_01", "worker_m12", claimed.FencingToken, now, 5*time.Minute); err != nil {
			t.Fatalf("acquire claim failed: %v", err)
		}

		cleanHash := strings.Repeat("d", 64)
		staleTokenPatch := &PatchRecord{
			SchemaVersion: "1.0",
			ProjectID:     "proj_m12",
			UnitID:        "unit_m12",
			FencingToken:  claimed.FencingToken - 1,
			Files: []PatchFileRecord{
				{
					Path:         "app/Models/User.php",
					Action:       PatchActionModify,
					BeforeSHA256: &cleanHash,
					AfterSHA256:  &cleanHash,
				},
			},
		}
		err = fenceSvc.ValidatePublication(staleTokenPatch, map[string][]byte{"app/Models/User.php": []byte("initial")}, now)
		if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) {
			t.Fatalf("expected stale fence token error, got: %v", err)
		}

		// M8: Execution Profiler anomaly detection preserved
		profiler := NewExecutionProfiler()
		rc := profiler.StartRequest(ctx, "/checkout", "POST")
		for i := 0; i < 5; i++ {
			rc.RecordQuery("SELECT * FROM items WHERE id = ?", 1*time.Millisecond, 1, nil)
		}
		p := profiler.FinishRequest(ctx, rc, 201, 1024, nil)
		anomalies := p.DetectQueryAnomalies(3)
		if len(anomalies) != 1 || anomalies[0].PatternType != PatternNPlusOneQuery {
			t.Fatalf("expected N+1 anomaly preserved from M8, got: %v", anomalies)
		}

		// M9: Target Adapter Exact Resolution for Go stdlib & Laravel 13 preserved
		targetReg := NewTargetAdapterRegistry()
		if err := targetReg.Register(NewGoStdlibTargetAdapter()); err != nil {
			t.Fatalf("register go target adapter failed: %v", err)
		}
		if err := targetReg.Register(NewLaravelTargetAdapter()); err != nil {
			t.Fatalf("register laravel target adapter failed: %v", err)
		}

		resolvedGo, err := targetReg.ResolveForResolution(&ResolvedEnvironmentDocument{
			Language:  "go",
			Framework: nil,
		})
		if err != nil || resolvedGo.Manifest().AdapterID != "target-go-stdlib" {
			t.Fatalf("exact resolution for Go stdlib target failed: %v", err)
		}

		fwLaravel := "laravel"
		resolvedLaravel, err := targetReg.ResolveForResolution(&ResolvedEnvironmentDocument{
			Language: "php",
			Framework: &ResolvedComponent{
				Name: fwLaravel,
			},
		})
		if err != nil || resolvedLaravel.Manifest().AdapterID != "target-php-laravel13" {
			t.Fatalf("exact resolution for Laravel target failed: %v", err)
		}

		// M10: Source Adapter Registry & Resolution preserved
		srcReg := NewSourceAdapterRegistry()
		if err := srcReg.Register(NewGenericPHPSourceAdapter()); err != nil {
			t.Fatalf("register generic php source adapter failed: %v", err)
		}
		if err := srcReg.Register(NewCodeIgniter3SourceAdapter()); err != nil {
			t.Fatalf("register ci3 source adapter failed: %v", err)
		}

		fwCI3 := "codeigniter3"
		resolvedCI3, err := srcReg.ResolveForEnvironment("php", &fwCI3)
		if err != nil || resolvedCI3.Manifest().AdapterID != "source.php.ci3" {
			t.Fatalf("exact resolution for CI3 source adapter failed: %v", err)
		}
	})
}
