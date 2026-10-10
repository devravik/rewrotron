package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// ArchiveRetentionClass identifies the retention classification of an archived object.
type ArchiveRetentionClass string

const (
	RetentionApproval        ArchiveRetentionClass = "APPROVAL"
	RetentionContract        ArchiveRetentionClass = "CONTRACT"
	RetentionVerification    ArchiveRetentionClass = "VERIFICATION"
	RetentionExperiment      ArchiveRetentionClass = "EXPERIMENT"
	RetentionLesson          ArchiveRetentionClass = "LESSON"
	RetentionTrace           ArchiveRetentionClass = "TRACE"
	RetentionSnapshot        ArchiveRetentionClass = "SNAPSHOT"
	RetentionHeartbeat       ArchiveRetentionClass = "HEARTBEAT"
	RetentionAudit           ArchiveRetentionClass = "AUDIT"
	RetentionTemporary       ArchiveRetentionClass = "TEMPORARY"
	RetentionGlobalKnowledge ArchiveRetentionClass = "GLOBAL_KNOWLEDGE"
)

// ArchiveObjectEntry describes an object stored in the coordinated backup archive.
type ArchiveObjectEntry struct {
	Ref            string                `json:"ref"`
	SizeBytes      int64                 `json:"size_bytes"`
	RetentionClass ArchiveRetentionClass `json:"retention_class"`
	Private        bool                  `json:"private"`
}

// ArchiveManifestRecord represents schema 2.0 of specs/operations/archive.schema.json.
type ArchiveManifestRecord struct {
	SchemaVersion       string               `json:"schema_version"` // "2.0"
	ProjectID           string               `json:"project_id"`
	DBHash              string               `json:"db_hash"`
	ThroughSequence     int64                `json:"through_sequence"`
	MigrationHashes     []string             `json:"migration_hashes"`
	Objects             []ArchiveObjectEntry `json:"objects"`
	ConfigRef           string               `json:"config_ref"`
	PrivateKeyPolicyRef string               `json:"private_key_policy_ref"`
	ClosureReportRef    string               `json:"closure_report_ref"`
}

// BackupError indicates a failure during backup or restore.
type BackupError struct {
	Code    string
	Message string
	Cause   error
}

func (e *BackupError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BackupError) Unwrap() error { return e.Cause }

// CoordinatedBackupManager coordinates atomic online SQLite backup and referenced artifact preservation
// under a publication/GC barrier per docs/08-storage/RECOVERY.md and Test X28.
type CoordinatedBackupManager struct {
	barrierMu sync.RWMutex
}

// NewCoordinatedBackupManager initializes the coordinated backup coordinator.
func NewCoordinatedBackupManager() *CoordinatedBackupManager {
	return &CoordinatedBackupManager{}
}

// WithPublicationBarrier acquires an exclusive lock simulating a publication/GC barrier.
func (m *CoordinatedBackupManager) WithPublicationBarrier(fn func() error) error {
	m.barrierMu.Lock()
	defer m.barrierMu.Unlock()
	return fn()
}

// CreateBackup performs a coordinated online backup:
// 1. Enforces publication/GC barrier.
// 2. Performs WAL checkpoint / VACUUM INTO / online copy to target backup database file.
// 3. Captures committed through_sequence, db_hash, and applied migration hashes.
// 4. Copies all referenced artifacts in objects, verifying byte counts and SHA-256 digests.
// 5. Emits archive manifest version 2.0.
func (m *CoordinatedBackupManager) CreateBackup(
	ctx context.Context,
	sourceDB *sql.DB,
	sourceArtifactsDir string,
	backupDir string,
	projectID string,
	referencedObjects []ArchiveObjectEntry,
	configRef string,
	privateKeyPolicyRef string,
	closureReportRef string,
) (*ArchiveManifestRecord, error) {
	var manifest *ArchiveManifestRecord
	err := m.WithPublicationBarrier(func() error {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return &BackupError{Code: "ERR_BACKUP_IO", Message: "failed to create backup dir", Cause: err}
		}

		backupDBPath := filepath.Join(backupDir, "project.db")
		// Execute SQLite VACUUM INTO to obtain a consistent atomic backup with flushed WAL
		vacuumQuery := fmt.Sprintf("VACUUM INTO '%s'", backupDBPath)
		if _, err := sourceDB.ExecContext(ctx, vacuumQuery); err != nil {
			return &BackupError{Code: "ERR_BACKUP_SQL", Message: "VACUUM INTO failed", Cause: err}
		}

		// Compute DB file hash
		dbHash, err := hashFileHex(backupDBPath)
		if err != nil {
			return &BackupError{Code: "ERR_BACKUP_HASH", Message: "failed to hash backup db", Cause: err}
		}

		// Read through_sequence from events/operations in backup DB
		var throughSeq int64
		row := sourceDB.QueryRowContext(ctx, "SELECT COALESCE(MAX(sequence), 0) FROM events")
		_ = row.Scan(&throughSeq)

		// Read applied migration hashes
		rows, err := sourceDB.QueryContext(ctx, "SELECT checksum FROM schema_migrations ORDER BY version ASC")
		if err != nil {
			return &BackupError{Code: "ERR_BACKUP_SQL", Message: "failed to query migration checksums", Cause: err}
		}
		defer rows.Close()

		var migrationHashes []string
		for rows.Next() {
			var h string
			if err := rows.Scan(&h); err == nil {
				migrationHashes = append(migrationHashes, h)
			}
		}
		if len(migrationHashes) == 0 {
			migrationHashes = []string{"0000000000000000000000000000000000000000000000000000000000000000"}
		}

		// Copy and verify each referenced artifact into backup artifacts dir
		backupArtifactsDir := filepath.Join(backupDir, "artifacts")
		if err := os.MkdirAll(backupArtifactsDir, 0755); err != nil {
			return &BackupError{Code: "ERR_BACKUP_IO", Message: "failed to create backup artifacts dir", Cause: err}
		}

		// Sort objects by ref for deterministic manifest ordering
		sortedObjects := make([]ArchiveObjectEntry, len(referencedObjects))
		copy(sortedObjects, referencedObjects)
		sort.Slice(sortedObjects, func(i, j int) bool {
			return sortedObjects[i].Ref < sortedObjects[j].Ref
		})

		for _, obj := range sortedObjects {
			// ref is sha256:<digest>
			digest := obj.Ref
			if len(digest) > 7 && digest[:7] == "sha256:" {
				digest = digest[7:]
			}
			srcFile := filepath.Join(sourceArtifactsDir, digest)
			dstFile := filepath.Join(backupArtifactsDir, digest)

			srcInfo, err := os.Stat(srcFile)
			if err != nil {
				return &BackupError{
					Code:    "ERR_BACKUP_MISSING_ARTIFACT",
					Message: fmt.Sprintf("referenced artifact %s missing in source storage", obj.Ref),
					Cause:   err,
				}
			}

			if srcInfo.Size() != obj.SizeBytes {
				return &BackupError{
					Code:    "ERR_BACKUP_CORRUPT_ARTIFACT",
					Message: fmt.Sprintf("artifact %s size mismatch: expected %d, got %d", obj.Ref, obj.SizeBytes, srcInfo.Size()),
				}
			}

			// Copy and hash
			copiedHash, err := copyAndHashFile(srcFile, dstFile)
			if err != nil {
				return &BackupError{Code: "ERR_BACKUP_IO", Message: "failed to copy artifact", Cause: err}
			}
			if copiedHash != digest {
				return &BackupError{
					Code:    "ERR_BACKUP_CORRUPT_ARTIFACT",
					Message: fmt.Sprintf("artifact %s digest mismatch: computed sha256:%s", obj.Ref, copiedHash),
				}
			}
		}

		manifest = &ArchiveManifestRecord{
			SchemaVersion:       "2.0",
			ProjectID:           projectID,
			DBHash:              dbHash,
			ThroughSequence:     throughSeq,
			MigrationHashes:     migrationHashes,
			Objects:             sortedObjects,
			ConfigRef:           configRef,
			PrivateKeyPolicyRef: privateKeyPolicyRef,
			ClosureReportRef:    closureReportRef,
		}

		manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return &BackupError{Code: "ERR_BACKUP_JSON", Message: "failed to marshal manifest", Cause: err}
		}

		if err := os.WriteFile(filepath.Join(backupDir, "manifest.json"), manifestBytes, 0644); err != nil {
			return &BackupError{Code: "ERR_BACKUP_IO", Message: "failed to write manifest.json", Cause: err}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// RestoreBackup restores a project from an archive directory into a destination workspace per RECOVERY.md:
// "Restore into an empty owner-only workspace, verify manifest/DB/object digests, integrity_check/FKs/schema/
//  canonical subject reconstruction and private correlation material... Missing object/archive/key yields
//  BACKUP_INCOMPLETE, not valid gates."
func RestoreBackup(
	ctx context.Context,
	backupDir string,
	restoreDir string,
) (*sql.DB, *ArchiveManifestRecord, error) {
	manifestPath := filepath.Join(backupDir, "manifest.json")
	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: "manifest.json missing", Cause: err}
	}

	var manifest ArchiveManifestRecord
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: "manifest.json corrupt", Cause: err}
	}

	if manifest.SchemaVersion != "2.0" {
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: fmt.Sprintf("unsupported manifest schema_version %s", manifest.SchemaVersion)}
	}

	backupDBPath := filepath.Join(backupDir, "project.db")
	if _, err := os.Stat(backupDBPath); err != nil {
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: "project.db missing from backup", Cause: err}
	}

	// Verify DB file hash against manifest
	dbHash, err := hashFileHex(backupDBPath)
	if err != nil || dbHash != manifest.DBHash {
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: fmt.Sprintf("db_hash mismatch: expected %s, got %s", manifest.DBHash, dbHash)}
	}

	// Verify all referenced objects exist in backup and match digest
	backupArtifactsDir := filepath.Join(backupDir, "artifacts")
	for _, obj := range manifest.Objects {
		digest := obj.Ref
		if len(digest) > 7 && digest[:7] == "sha256:" {
			digest = digest[7:]
		}
		objPath := filepath.Join(backupArtifactsDir, digest)
		info, err := os.Stat(objPath)
		if err != nil {
			return nil, nil, &BackupError{
				Code:    "BACKUP_INCOMPLETE",
				Message: fmt.Sprintf("referenced artifact %s missing from backup", obj.Ref),
			}
		}
		if info.Size() != obj.SizeBytes {
			return nil, nil, &BackupError{
				Code:    "BACKUP_INCOMPLETE",
				Message: fmt.Sprintf("artifact %s size mismatch: expected %d, got %d", obj.Ref, obj.SizeBytes, info.Size()),
			}
		}
		actualHash, err := hashFileHex(objPath)
		if err != nil || actualHash != digest {
			return nil, nil, &BackupError{
				Code:    "BACKUP_INCOMPLETE",
				Message: fmt.Sprintf("artifact %s digest mismatch: expected sha256:%s, got sha256:%s", obj.Ref, digest, actualHash),
			}
		}
	}

	// Prepare restore directory
	if err := os.MkdirAll(restoreDir, 0700); err != nil {
		return nil, nil, &BackupError{Code: "ERR_RESTORE_IO", Message: "failed to create restore dir", Cause: err}
	}

	restoredDBPath := filepath.Join(restoreDir, "project.db")
	if _, err := copyAndHashFile(backupDBPath, restoredDBPath); err != nil {
		return nil, nil, &BackupError{Code: "ERR_RESTORE_IO", Message: "failed to copy restored db", Cause: err}
	}

	restoredArtifactsDir := filepath.Join(restoreDir, "artifacts")
	if err := os.MkdirAll(restoredArtifactsDir, 0700); err != nil {
		return nil, nil, &BackupError{Code: "ERR_RESTORE_IO", Message: "failed to create restored artifacts dir", Cause: err}
	}

	for _, obj := range manifest.Objects {
		digest := obj.Ref
		if len(digest) > 7 && digest[:7] == "sha256:" {
			digest = digest[7:]
		}
		src := filepath.Join(backupArtifactsDir, digest)
		dst := filepath.Join(restoredArtifactsDir, digest)
		if _, err := copyAndHashFile(src, dst); err != nil {
			return nil, nil, &BackupError{Code: "ERR_RESTORE_IO", Message: "failed to copy restored artifact", Cause: err}
		}
	}

	// Open restored DB and perform SQLite integrity_check & foreign_key_check
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", restoredDBPath)
	restoredDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, &BackupError{Code: "ERR_RESTORE_SQL", Message: "failed to open restored db", Cause: err}
	}

	var integrityResult string
	if err := restoredDB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrityResult); err != nil || integrityResult != "ok" {
		_ = restoredDB.Close()
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: fmt.Sprintf("integrity_check failed: %s", integrityResult), Cause: err}
	}

	fkRows, err := restoredDB.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		_ = restoredDB.Close()
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: "foreign_key_check query failed", Cause: err}
	}
	defer fkRows.Close()
	if fkRows.Next() {
		_ = restoredDB.Close()
		return nil, nil, &BackupError{Code: "BACKUP_INCOMPLETE", Message: "foreign_key_check failed; constraint violations present"}
	}

	return restoredDB, &manifest, nil
}

func copyAndHashFile(src string, dst string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	h := sha256.New()
	mw := io.MultiWriter(out, h)
	if _, err := io.Copy(mw, in); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashFileHex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
