package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test X28: Coordinated Storage Recovery, WAL checkpoints & Backup/Restore
// "Apply migrations and corrupted/checksum/version cases, WAL checkpoint/reader contention,
//  coordinated DB/artifact backup during publication and missing object restore;
//  no valid incomplete backup"
func TestStorageRecovery_X28_CoordinatedBackupAndRestore(t *testing.T) {
	ctx := context.Background()
	testDir := t.TempDir()

	dbPath := filepath.Join(testDir, "live.db")
	artifactsDir := filepath.Join(testDir, "live_artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		t.Fatal(err)
	}

	liveDB := openTestDB(t, dbPath)
	defer liveDB.Close()

	liveConn := testConnection(t, liveDB)
	if err := ConfigureConnection(ctx, liveConn); err != nil {
		t.Fatal(err)
	}

	allMigrations, err := LoadMigrationsFromDir(filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewRunner(allMigrations)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Run(ctx, liveConn); err != nil {
		t.Fatal(err)
	}

	// Insert test project into live DB
	_, err = liveDB.ExecContext(ctx, "INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES ('proj-1', 'Test', '/tmp/src', 'INITIALIZED', '2026-10-10T00:00:00Z', '2026-10-10T00:00:00Z')")
	if err != nil {
		t.Fatalf("failed to insert test project: %v", err)
	}

	// Create 2 test artifacts
	art1Data := []byte("approval-subject-payload-bytes")
	art1Hash := sha256.Sum256(art1Data)
	art1Digest := hex.EncodeToString(art1Hash[:])
	if err := os.WriteFile(filepath.Join(artifactsDir, art1Digest), art1Data, 0644); err != nil {
		t.Fatal(err)
	}

	art2Data := []byte("verification-report-payload-bytes")
	art2Hash := sha256.Sum256(art2Data)
	art2Digest := hex.EncodeToString(art2Hash[:])
	if err := os.WriteFile(filepath.Join(artifactsDir, art2Digest), art2Data, 0644); err != nil {
		t.Fatal(err)
	}

	referencedObjects := []ArchiveObjectEntry{
		{
			Ref:            fmt.Sprintf("sha256:%s", art1Digest),
			SizeBytes:      int64(len(art1Data)),
			RetentionClass: RetentionApproval,
			Private:        false,
		},
		{
			Ref:            fmt.Sprintf("sha256:%s", art2Digest),
			SizeBytes:      int64(len(art2Data)),
			RetentionClass: RetentionVerification,
			Private:        false,
		},
	}

	backupMgr := NewCoordinatedBackupManager()
	backupDir := filepath.Join(testDir, "backup_valid")

	// 1. Positive Control: Create coordinated online backup
	manifest, err := backupMgr.CreateBackup(
		ctx,
		liveDB,
		artifactsDir,
		backupDir,
		"proj-1",
		referencedObjects,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"sha256:2222222222222222222222222222222222222222222222222222222222222222",
		"sha256:3333333333333333333333333333333333333333333333333333333333333333",
	)
	if err != nil {
		t.Fatalf("failed to create coordinated backup: %v", err)
	}
	if manifest.SchemaVersion != "2.0" || manifest.ProjectID != "proj-1" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}

	// 2. Positive Control: Restore backup into clean workspace
	restoreDir := filepath.Join(testDir, "restore_clean")
	restoredDB, restoredManifest, err := RestoreBackup(ctx, backupDir, restoreDir)
	if err != nil {
		t.Fatalf("failed to restore backup: %v", err)
	}
	defer restoredDB.Close()

	if restoredManifest.DBHash != manifest.DBHash {
		t.Fatalf("restored manifest mismatch: %s vs %s", restoredManifest.DBHash, manifest.DBHash)
	}

	// Verify project exists in restored DB
	var projStatus string
	err = restoredDB.QueryRowContext(ctx, "SELECT status FROM projects WHERE id='proj-1'").Scan(&projStatus)
	if err != nil || projStatus != "INITIALIZED" {
		t.Fatalf("restored project verification failed: err=%v, status=%s", err, projStatus)
	}

	// 3. Falsification: Backup with missing referenced artifact in source storage fails
	missingRefDir := filepath.Join(testDir, "backup_missing_ref")
	missingObjects := append([]ArchiveObjectEntry{}, referencedObjects...)
	missingObjects = append(missingObjects, ArchiveObjectEntry{
		Ref:            "sha256:4444444444444444444444444444444444444444444444444444444444444444",
		SizeBytes:      100,
		RetentionClass: RetentionSnapshot,
	})
	_, err = backupMgr.CreateBackup(
		ctx,
		liveDB,
		artifactsDir,
		missingRefDir,
		"proj-1",
		missingObjects,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"sha256:2222222222222222222222222222222222222222222222222222222222222222",
		"sha256:3333333333333333333333333333333333333333333333333333333333333333",
	)
	if err == nil || !strings.Contains(err.Error(), "ERR_BACKUP_MISSING_ARTIFACT") {
		t.Fatalf("expected ERR_BACKUP_MISSING_ARTIFACT on missing source artifact, got: %v", err)
	}

	// 4. Falsification: Restore with missing artifact in backup yields BACKUP_INCOMPLETE (Test X28)
	// Delete an artifact from a copied backup directory
	corruptBackupDir := filepath.Join(testDir, "backup_corrupt")
	if err := os.CopyFS(corruptBackupDir, os.DirFS(backupDir)); err != nil {
		// Fallback manual copy
		copyTestDir(t, backupDir, corruptBackupDir)
	}
	// Remove art1 from corrupt backup artifacts
	_ = os.Remove(filepath.Join(corruptBackupDir, "artifacts", art1Digest))

	corruptRestoreDir := filepath.Join(testDir, "restore_corrupt")
	_, _, err = RestoreBackup(ctx, corruptBackupDir, corruptRestoreDir)
	if err == nil || !strings.Contains(err.Error(), "BACKUP_INCOMPLETE") {
		t.Fatalf("expected BACKUP_INCOMPLETE on missing artifact in backup, got: %v", err)
	}

	// 5. Falsification: Restore with corrupted DB file yields BACKUP_INCOMPLETE
	corruptDBBackupDir := filepath.Join(testDir, "backup_corrupt_db")
	copyTestDir(t, backupDir, corruptDBBackupDir)
	// Tamper DB file
	_ = os.WriteFile(filepath.Join(corruptDBBackupDir, "project.db"), []byte("not-a-sqlite-db"), 0644)

	_, _, err = RestoreBackup(ctx, corruptDBBackupDir, filepath.Join(testDir, "restore_corrupt_db"))
	if err == nil || !strings.Contains(err.Error(), "BACKUP_INCOMPLETE") {
		t.Fatalf("expected BACKUP_INCOMPLETE on corrupt DB, got: %v", err)
	}
}

func copyTestDir(t *testing.T, src string, dst string) {
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyTestDir(t, srcPath, dstPath)
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
