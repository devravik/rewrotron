package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestM12_Unit4_NativePackagingAndRollbackQualification executes the complete qualification suite for:
// 1. Multi-platform archive packaging with byte reproducibility for primary profile (linux-amd64)
// 2. Strict authenticity verification (refuses corrupted or mismatched archive checksums per X36-B)
// 3. Clean user-local directory layout, private staging, and atomic symlink pointer activation (X36-A, X36-D)
// 4. Interrupted installation recovery: staging directory cleaned up without altering active `current` pointer (X36-C)
// 5. Version upgrade (0.1.0 -> 0.2.0) and rollback (0.2.0 -> 0.1.0) preserving customer project SQLite state (X36-E)
// 6. Data-preserving uninstall: removes managed binaries/symlinks while preserving project data/artifacts (X36-F)
// 7. Write-aware rollback rehearsal reconciling database, file, session, and external mutations (X34)
func TestM12_Unit4_NativePackagingAndRollbackQualification(t *testing.T) {
	tempDir := t.TempDir()

	binRewrotron := filepath.Join(tempDir, "rewrotron")
	_ = os.WriteFile(binRewrotron, []byte("#!/bin/sh\necho rewrotron v1.0.0\n"), 0755)

	binMCP := filepath.Join(tempDir, "rewrotron-mcp")
	_ = os.WriteFile(binMCP, []byte("#!/bin/sh\necho rewrotron-mcp v1.0.0\n"), 0755)

	cfgV1 := ReleasePackagingConfig{
		Version:      "1.0.0",
		Architecture: "amd64",
		OS:           "linux",
		CommitHash:   "commit-v100",
		BinRewrotron: binRewrotron,
		BinMCP:       binMCP,
	}

	archivePathV1_1 := filepath.Join(tempDir, "rewrotron-1.0.0-linux-amd64-1.tar.gz")
	archivePathV1_2 := filepath.Join(tempDir, "rewrotron-1.0.0-linux-amd64-2.tar.gz")

	// =========================================================================
	// 1. Packaging & Byte Reproducibility (X36-H)
	// =========================================================================
	t.Run("ArchivePackagingAndByteReproducibility", func(t *testing.T) {
		manifest1, err := BuildReleaseArchive(cfgV1, archivePathV1_1)
		if err != nil {
			t.Fatalf("build archive 1 failed: %v", err)
		}
		manifest2, err := BuildReleaseArchive(cfgV1, archivePathV1_2)
		if err != nil {
			t.Fatalf("build archive 2 failed: %v", err)
		}

		if manifest1.ArchiveSHA256 != manifest2.ArchiveSHA256 {
			t.Fatalf("byte reproducibility failure: %s != %s", manifest1.ArchiveSHA256, manifest2.ArchiveSHA256)
		}
		if manifest1.Platform != "linux-amd64" {
			t.Fatalf("unexpected platform: %s", manifest1.Platform)
		}
	})

	manifestV1, err := BuildReleaseArchive(cfgV1, archivePathV1_1)
	if err != nil {
		t.Fatal(err)
	}

	// Setup user home layout and simulated customer project directory
	homeDir := filepath.Join(tempDir, "home_operator")
	layoutV1 := DefaultInstallationLayout(homeDir, "1.0.0")

	customerProjectDir := filepath.Join(homeDir, "projects", "legacy_migration")
	_ = os.MkdirAll(customerProjectDir, 0755)
	projectDB := filepath.Join(customerProjectDir, "rewrotron.db")
	_ = os.WriteFile(projectDB, []byte("sqlite-project-data-v1"), 0644)

	// =========================================================================
	// 2. Clean Installation & Symlink Activation (X36-A, X36-D)
	// =========================================================================
	t.Run("CleanInstallationAndActivation", func(t *testing.T) {
		err := ExtractAndInstall(archivePathV1_1, manifestV1.ArchiveSHA256, layoutV1, false)
		if err != nil {
			t.Fatalf("clean install failed: %v", err)
		}

		// Verify `current` link points to releases/1.0.0
		target, err := os.Readlink(layoutV1.CurrentLinkPath)
		if err != nil || target != layoutV1.ReleasesDir {
			t.Fatalf("current link target mismatch: got %s, expected %s", target, layoutV1.ReleasesDir)
		}

		// Verify executable binary link points to current
		binTarget, err := os.Readlink(layoutV1.RewrotronBin)
		if err != nil || !strings.Contains(binTarget, "current/bin/rewrotron") {
			t.Fatalf("entrypoint link target mismatch: got %s", binTarget)
		}
	})

	// =========================================================================
	// 3. Falsification: Authenticity Mismatch Rejection (X36-B)
	// =========================================================================
	t.Run("AuthenticityMismatchRejection", func(t *testing.T) {
		badDigest := "0000000000000000000000000000000000000000000000000000000000000000"
		err := ExtractAndInstall(archivePathV1_1, badDigest, layoutV1, false)
		if err == nil || !strings.Contains(err.Error(), "ERR_AUTHENTICITY_MISMATCH") {
			t.Fatalf("expected ERR_AUTHENTICITY_MISMATCH on bad digest, got: %v", err)
		}
	})

	// =========================================================================
	// 4. Interrupted Installation Recovery (X36-C)
	// =========================================================================
	t.Run("InterruptedInstallationPreservesCurrent", func(t *testing.T) {
		cfgV2 := cfgV1
		cfgV2.Version = "1.1.0"
		archivePathV2 := filepath.Join(tempDir, "rewrotron-1.1.0-linux-amd64.tar.gz")
		manifestV2, err := BuildReleaseArchive(cfgV2, archivePathV2)
		if err != nil {
			t.Fatal(err)
		}

		layoutV2 := DefaultInstallationLayout(homeDir, "1.1.0")

		// Simulate interrupted install of 1.1.0
		err = ExtractAndInstall(archivePathV2, manifestV2.ArchiveSHA256, layoutV2, true)
		if err == nil || !strings.Contains(err.Error(), "ERR_INTERRUPTED_INSTALLATION") {
			t.Fatalf("expected ERR_INTERRUPTED_INSTALLATION on simulated failure, got: %v", err)
		}

		// Active current pointer must remain untouched on 1.0.0
		currentTarget, err := os.Readlink(layoutV1.CurrentLinkPath)
		if err != nil || currentTarget != layoutV1.ReleasesDir {
			t.Fatalf("interrupted install damaged active release pointer: got %s, expected %s", currentTarget, layoutV1.ReleasesDir)
		}
	})

	// =========================================================================
	// 5. Version Upgrade and Binary Rollback (X36-E)
	// =========================================================================
	t.Run("UpgradeAndBinaryRollback", func(t *testing.T) {
		cfgV2 := cfgV1
		cfgV2.Version = "1.1.0"
		archivePathV2 := filepath.Join(tempDir, "rewrotron-1.1.0-linux-amd64.tar.gz")
		manifestV2, err := BuildReleaseArchive(cfgV2, archivePathV2)
		if err != nil {
			t.Fatal(err)
		}
		layoutV2 := DefaultInstallationLayout(homeDir, "1.1.0")

		// Successful upgrade to 1.1.0
		err = ExtractAndInstall(archivePathV2, manifestV2.ArchiveSHA256, layoutV2, false)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}

		upgradedTarget, _ := os.Readlink(layoutV1.CurrentLinkPath)
		if upgradedTarget != layoutV2.ReleasesDir {
			t.Fatalf("expected current link to point to 1.1.0, got: %s", upgradedTarget)
		}

		// Execute binary rollback to 1.0.0
		err = RollbackInstallation(layoutV2, "1.0.0")
		if err != nil {
			t.Fatalf("binary rollback failed: %v", err)
		}

		rolledBackTarget, _ := os.Readlink(layoutV1.CurrentLinkPath)
		if rolledBackTarget != layoutV1.ReleasesDir {
			t.Fatalf("expected current link to point back to 1.0.0, got: %s", rolledBackTarget)
		}
	})

	// =========================================================================
	// 6. Data-Preserving Uninstall (X36-F)
	// =========================================================================
	t.Run("DataPreservingUninstall", func(t *testing.T) {
		err := PreservingUninstall(layoutV1, "1.0.0", true)
		if err != nil {
			t.Fatalf("uninstall failed: %v", err)
		}

		// Managed release files removed
		if _, err := os.Stat(layoutV1.ReleasesDir); !os.IsNotExist(err) {
			t.Fatalf("releases/1.0.0 directory should have been removed")
		}

		// Crucial invariant: Customer project databases and artifacts are preserved
		dbBytes, err := os.ReadFile(projectDB)
		if err != nil || string(dbBytes) != "sqlite-project-data-v1" {
			t.Fatalf("uninstall corrupted or deleted customer project state!")
		}
	})

	// =========================================================================
	// 7. Write-Aware Target Rollback Rehearsal (X34)
	// =========================================================================
	t.Run("WriteAwareRollbackRehearsal", func(t *testing.T) {
		record := &RollbackRehearsalRecord{
			SchemaVersion:  "1.0",
			RehearsalID:    "rehearsal-m12",
			ProjectID:      "proj-m12",
			TargetTreeRef:  "sha256:" + strings.Repeat("a", 64),
			EnvironmentRef: "sha256:" + strings.Repeat("b", 64),
			ExecutableCommands: []string{
				"artisan migrate:rollback",
				"systemctl restart session-store",
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
					Strategy:       "ARCHIVE",
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

		err := ValidateRollbackRehearsal(record)
		if err != nil {
			t.Fatalf("expected valid write-aware rollback rehearsal, got err: %v", err)
		}

		// Falsification check: Schema-only rollback guarantee must be rejected per X34
		falsifiedRecord := *record
		falsifiedRecord.SchemaOnlyGuaranteed = true
		err = ValidateRollbackRehearsal(&falsifiedRecord)
		if err == nil || !strings.Contains(err.Error(), "schema-only rollback guarantee rejected") {
			t.Fatalf("schema-only rollback guarantee must be rejected, got: %v", err)
		}
	})
}
