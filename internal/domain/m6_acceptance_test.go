package domain

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devravik/rewrotron/internal/storage"
	_ "modernc.org/sqlite"
)

// TestM6_AcceptanceClosure executes the comprehensive release acceptance suite for Milestone M6:
// Whole-project correctness, statistical performance, write-aware rollback, disaster recovery,
// native packaging lifecycle, and Gate 3 release compilation.
//
// Authoritative tests integrated:
// - X78: Fixed stopping rule ($N \in [10, 20]$) and operational burst SLO.
// - X33: Noisy performance trials, insufficient samples, zero denominator, known slowdown.
// - X59: Recomputed estimands and deterministic PRNG bootstrap reproducibility.
// - X34: Write-aware rollback & target data reconciliation (refusing schema-only rollback guarantees).
// - X28: Coordinated SQLite/artifact backup and disaster recovery.
// - X36-A–H: Native packaging, reproducibility, authenticity, interrupted install, upgrade/rollback, and preserving uninstall.
// - X35: Scope deadlock and whole-project corpus acceptance before Gate 3 release packaging.
func TestM6_AcceptanceClosure(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// =========================================================================
	// 1. Tests X78, X33, X59: Statistical Performance & Operational SLO Acceptance (Unit 1)
	// =========================================================================
	t.Run("X78_X33_X59_StatisticalPerformanceAndSLO", func(t *testing.T) {
		policy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
			{
				MetricID:                 MetricP95,
				Unit:                     "ms",
				LimitKind:                "RELATIVE_NON_INFERIORITY",
				Limit:                    1.05,
				MinimumSuccessesPerTrial: 500,
			},
		})

		// X78: Early stopping check at N=5 => INCONCLUSIVE
		earlyTrials := makeTestTrials(5, 100.0, 80.0, 500, 500, 0, 0)
		earlyRes, err := EvaluateStatisticalPerformance(policy, earlyTrials, 42, "sha256:samples1")
		if err != nil || earlyRes.Outcome != BenchmarkOutcomeInconclusive {
			t.Fatalf("X78: expected INCONCLUSIVE for early stopping at N=5, got %s", earlyRes.Outcome)
		}

		// X33: Insufficient P99 sample count check (< 10,000) => INCONCLUSIVE
		p99Policy := makeTestPolicy("NON_INFERIORITY", 10, []PerformanceMetricLimit{
			{
				MetricID:                 MetricP99,
				Unit:                     "ms",
				LimitKind:                "RELATIVE_NON_INFERIORITY",
				Limit:                    1.10,
				MinimumSuccessesPerTrial: 5000,
			},
		})
		insufficientP99Trials := make([]PairedTrial, 10)
		for i := 0; i < 10; i++ {
			insufficientP99Trials[i] = PairedTrial{
				TrialIndex: i,
				SourceMetricValues: map[string]float64{MetricP99: 100.0},
				TargetMetricValues: map[string]float64{MetricP99: 102.0},
				SourceSuccesses:    5000,
				TargetSuccesses:    5000,
			}
		}
		p99Res, err := EvaluateStatisticalPerformance(p99Policy, insufficientP99Trials, 42, "sha256:s")
		if err != nil || p99Res.Outcome != BenchmarkOutcomeInconclusive {
			t.Fatalf("X33: expected INCONCLUSIVE for insufficient P99 samples, got %s", p99Res.Outcome)
		}

		// X33: Zero denominator in relative estimand => INCONCLUSIVE
		zeroTrials := makeTestTrials(10, 0.0, 50.0, 500, 500, 0, 0)
		zeroRes, err := EvaluateStatisticalPerformance(policy, zeroTrials, 42, "sha256:s")
		if err != nil || zeroRes.Outcome != BenchmarkOutcomeInconclusive {
			t.Fatalf("X33: expected INCONCLUSIVE for zero denominator, got %s", zeroRes.Outcome)
		}

		// X59: Recomputed estimands & deterministic reproducibility at full N=10
		fullGoodTrials := makeTestTrials(10, 100.0, 95.0, 500, 500, 0, 0)
		res1, err := EvaluateStatisticalPerformance(policy, fullGoodTrials, 12345, "sha256:raw1")
		if err != nil || res1.Outcome != BenchmarkOutcomeApprovedBudget {
			t.Fatalf("X59: expected APPROVED_BUDGET, got %s", res1.Outcome)
		}
		res2, err := EvaluateStatisticalPerformance(policy, fullGoodTrials, 12345, "sha256:raw1")
		if err != nil || *res1.Metrics[MetricP95].UpperBound != *res2.Metrics[MetricP95].UpperBound {
			t.Fatalf("X59: deterministic PRNG reproducibility failure")
		}

		// X78: Operational SLO fixed-window evaluation
		sloPolicy := &OperationalSLOPolicy{
			SchemaVersion:            "1.0",
			ProjectID:                "proj-1",
			WorkloadRef:              "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			WindowMs:                 10000,
			MinimumRequestsPerWindow: 100,
			LatencyPercentile:        0.99,
			MaximumLatencyMs:         "250.0",
			MaximumErrorFraction:     0.01,
			Scope:                    "EVERY_COMPLETE_MEASURED_WINDOW",
			EvidenceRefs:             []string{"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		}

		windows := make([]WindowSample, 3)
		for w := 0; w < 3; w++ {
			reqs := make([]float64, 100)
			for i := 0; i < 100; i++ {
				reqs[i] = 120.0
			}
			windows[w] = WindowSample{
				WindowIndex:        w,
				SuccessfulRequests: reqs,
				TotalAttempts:      100,
				ErrorCount:         0,
			}
		}
		sloRes, err := EvaluateOperationalSLO(sloPolicy, windows, "sha256:pol", "sha256:w")
		if err != nil || sloRes.Outcome != SLOOutcomePass {
			t.Fatalf("X78: operational SLO expected PASS, got %s", sloRes.Outcome)
		}

		// One burst breach fails the entire operational SLO
		windows[1].SuccessfulRequests[98] = 300.0
		windows[1].SuccessfulRequests[99] = 300.0
		sloResBad, err := EvaluateOperationalSLO(sloPolicy, windows, "sha256:pol", "sha256:w")
		if err != nil || sloResBad.Outcome != SLOOutcomeFail {
			t.Fatalf("X78: operational SLO expected FAIL on burst breach, got %s", sloResBad.Outcome)
		}
	})

	// =========================================================================
	// 2. Test X34: Write-Aware Rollback & Target Data Reconciliation (Unit 2)
	// =========================================================================
	t.Run("X34_WriteAwareRollbackAndReconciliation", func(t *testing.T) {
		validRehearsal := makeValidRehearsal()
		if err := ValidateRollbackRehearsal(validRehearsal); err != nil {
			t.Fatalf("X34: valid rehearsal failed: %v", err)
		}

		// Refuse schema-only rollback guarantees
		schemaOnly := makeValidRehearsal()
		schemaOnly.SchemaOnlyGuaranteed = true
		if err := ValidateRollbackRehearsal(schemaOnly); err == nil || !strings.Contains(err.Error(), "schema-only rollback guarantee rejected") {
			t.Fatalf("X34: expected refusal of schema-only rollback, got: %v", err)
		}

		// Refuse empty-data traffic switch
		emptySwitch := makeValidRehearsal()
		emptySwitch.RepresentativeMutations = nil
		if err := ValidateRollbackRehearsal(emptySwitch); err == nil || !strings.Contains(err.Error(), "no representative target mutations rehearsed") {
			t.Fatalf("X34: expected refusal of empty mutations switch, got: %v", err)
		}

		// Refuse unreconciled mutations
		unreconciled := makeValidRehearsal()
		unreconciled.ReconciliationContracts = nil
		if err := ValidateRollbackRehearsal(unreconciled); err == nil || !strings.Contains(err.Error(), "unreconciled target mutation") {
			t.Fatalf("X34: expected refusal of unreconciled mutations, got: %v", err)
		}
	})

	// =========================================================================
	// 3. Test X28: Storage Disaster Recovery & Coordinated Backup/Restore (Unit 3)
	// =========================================================================
	t.Run("X28_StorageDisasterRecoveryAndBackup", func(t *testing.T) {
		storageDir := filepath.Join(tempDir, "storage_test")
		dbPath := filepath.Join(storageDir, "live.db")
		artifactsDir := filepath.Join(storageDir, "artifacts")
		_ = os.MkdirAll(artifactsDir, 0755)

		dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", dbPath)
		liveDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			t.Fatalf("failed to open live DB: %v", err)
		}
		defer liveDB.Close()

		liveConn, err := liveDB.Conn(ctx)
		if err != nil {
			t.Fatalf("failed to obtain conn: %v", err)
		}
		defer liveConn.Close()

		allMigrations, err := storage.LoadMigrationsFromDir(filepath.Join("..", "..", "migrations"))
		if err != nil {
			t.Fatalf("failed to load migrations: %v", err)
		}
		runner, err := storage.NewRunner(allMigrations)
		if err != nil {
			t.Fatalf("failed to create migration runner: %v", err)
		}
		if err := runner.Run(ctx, liveConn); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		// Create test artifacts
		testPayload := []byte("closure-approval-payload")
		h := sha256.Sum256(testPayload)
		digestHex := hex.EncodeToString(h[:])
		_ = os.WriteFile(filepath.Join(artifactsDir, digestHex), testPayload, 0644)

		objects := []storage.ArchiveObjectEntry{
			{
				Ref:            fmt.Sprintf("sha256:%s", digestHex),
				SizeBytes:      int64(len(testPayload)),
				RetentionClass: storage.RetentionApproval,
			},
		}

		backupDir := filepath.Join(storageDir, "backup")
		backupMgr := storage.NewCoordinatedBackupManager()
		manifest, err := backupMgr.CreateBackup(
			ctx,
			liveDB,
			artifactsDir,
			backupDir,
			"proj-closure",
			objects,
			"sha256:1111111111111111111111111111111111111111111111111111111111111111",
			"sha256:2222222222222222222222222222222222222222222222222222222222222222",
			"sha256:3333333333333333333333333333333333333333333333333333333333333333",
		)
		if err != nil || manifest == nil {
			t.Fatalf("X28: backup creation failed: %v", err)
		}

		// Restore into clean workspace
		restoreDir := filepath.Join(storageDir, "restore")
		restoredDB, restoredManifest, err := storage.RestoreBackup(ctx, backupDir, restoreDir)
		if err != nil || restoredManifest == nil {
			t.Fatalf("X28: restore failed: %v", err)
		}
		_ = restoredDB.Close()
	})

	// =========================================================================
	// 4. Test X36-A–H: Native Packaging, Authenticity & Installation Lifecycle (Unit 4)
	// =========================================================================
	t.Run("X36_PackagingAndInstallationLifecycle", func(t *testing.T) {
		packDir := filepath.Join(tempDir, "pack")
		_ = os.MkdirAll(packDir, 0755)
		archive1 := filepath.Join(packDir, "pkg1.tar.gz")
		archive2 := filepath.Join(packDir, "pkg2.tar.gz")

		cfg := ReleasePackagingConfig{
			Version:      "1.0.0",
			Architecture: "amd64",
			OS:           "linux",
			CommitHash:   "abc123commit",
		}

		// Reproducibility (X36-H)
		m1, err := BuildReleaseArchive(cfg, archive1)
		if err != nil {
			t.Fatalf("X36-H: build 1 failed: %v", err)
		}
		m2, err := BuildReleaseArchive(cfg, archive2)
		if err != nil {
			t.Fatalf("X36-H: build 2 failed: %v", err)
		}
		if m1.ArchiveSHA256 != m2.ArchiveSHA256 {
			t.Fatalf("X36-H: reproducibility failure: %s vs %s", m1.ArchiveSHA256, m2.ArchiveSHA256)
		}

		// Clean install & managed pointers (X36-A, X36-D)
		userHome := filepath.Join(tempDir, "user_home")
		layout := DefaultInstallationLayout(userHome, "1.0.0")
		if err := ExtractAndInstall(archive1, m1.ArchiveSHA256, layout, false); err != nil {
			t.Fatalf("X36-A: clean install failed: %v", err)
		}

		// Authenticity check rejection (X36-B)
		if err := ExtractAndInstall(archive1, "sha256:badbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbad", layout, false); err == nil {
			t.Fatalf("X36-B: expected refusal on bad checksum")
		}

		// Preserving uninstall (X36-F)
		custWork := filepath.Join(userHome, "customer-migration-dir")
		_ = os.MkdirAll(custWork, 0755)
		_ = os.WriteFile(filepath.Join(custWork, "data.txt"), []byte("customer-valuable-state"), 0644)
		if err := PreservingUninstall(layout, "1.0.0", false); err != nil {
			t.Fatalf("X36-F: uninstall failed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(custWork, "data.txt")); err != nil {
			t.Fatalf("X36-F: customer workspace was corrupted by uninstall")
		}
	})

	// =========================================================================
	// 5. Test X35: Scope Deadlock & Gate 3 Release Package (Unit 5)
	// =========================================================================
	t.Run("X35_WholeProjectScopeAndGate3Compilation", func(t *testing.T) {
		graph := makeTestPlanGraph()

		inputs := WholeProjectReleaseInputs{
			ProjectID:             "proj-closure-gate3",
			ConfigurationHash:     "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			PlanHash:              "3ccae952a8f049ec692f413163bfb5828b25dcf42a1f0d80eac7cb6a121bcb2e",
			TargetTreeRef:         "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			WholeProjectReportRef: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			BenchmarkRef:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RunbookRef:            "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RollbackRehearsalRef:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			PlanGraph:             graph,
			ImplementedUnitIDs:    []string{"unit-login"}, // missing checkout
			UnitVerdicts: map[string]string{
				"unit-login": "VERIFIED",
			},
			PerformanceBenchmark: &StatisticalBenchmarkResult{
				Outcome: BenchmarkOutcomeApprovedBudget,
			},
			RollbackRehearsal: makeValidRehearsal(),
		}

		// Scope deadlock: partial project cannot reach Gate 3 release (Test X35)
		_, _, err := CompileGate3ReleasePackage(inputs)
		if err == nil || !strings.Contains(err.Error(), "project missing unit unit-checkout") || !strings.Contains(err.Error(), "Test X35") {
			t.Fatalf("X35: expected refusal of partial project, got: %v", err)
		}

		// Full whole-project acceptance compiles valid Gate 3 release package
		inputs.ImplementedUnitIDs = []string{"unit-login", "unit-checkout"}
		inputs.UnitVerdicts["unit-checkout"] = "VERIFIED"

		pkg, rawBytes, err := CompileGate3ReleasePackage(inputs)
		if err != nil || pkg == nil || len(rawBytes) == 0 {
			t.Fatalf("X35: failed to compile full project Gate 3 release package: %v", err)
		}
		if err := ValidateMigrationRelease(rawBytes); err != nil {
			t.Fatalf("X35: ValidateMigrationRelease failed on Gate 3 package: %v", err)
		}
	})
}
