package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func defaultMigrations(t *testing.T) []Migration {
	t.Helper()
	dir := filepath.Join("..", "..", "migrations")
	migrations, err := LoadMigrationsFromDir(dir)
	if err != nil {
		t.Fatalf("failed to load default migrations: %v", err)
	}
	return migrations
}

func freshTestConnection(t *testing.T) (*sql.DB, *sql.Conn, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "metadata.db")
	db := openTestDB(t, dbPath)
	conn := testConnection(t, db)
	return db, conn, dbPath
}

// 1. First migration
func TestRunner_FirstMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, conn, _ := freshTestConnection(t)

	all := defaultMigrations(t)
	// Run only first migration
	r, err := NewRunner([]Migration{all[0]})
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Run(ctx, conn); err != nil {
		t.Fatalf("failed to run first migration: %v", err)
	}

	applied, err := r.AppliedMigrations(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 1 {
		t.Fatalf("expected 1 applied migration, got %d", len(applied))
	}
	if applied[0].Version != 1 || applied[0].Name != "0001_initial.sql" {
		t.Fatalf("unexpected applied migration: %+v", applied[0])
	}
	if applied[0].Checksum != all[0].Checksum {
		t.Fatalf("checksum mismatch: got %s, want %s", applied[0].Checksum, all[0].Checksum)
	}

	// Verify table from 0001 exists and is usable
	var tableCount int
	if err := conn.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&tableCount); err != nil || tableCount != 1 {
		t.Fatalf("projects table not found after first migration")
	}
}

// 2. Already-applied migration (idempotent no-op)
func TestRunner_AlreadyAppliedMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, conn, _ := freshTestConnection(t)

	r, err := NewRunner(defaultMigrations(t))
	if err != nil {
		t.Fatal(err)
	}

	// First run applies all
	if err := r.Run(ctx, conn); err != nil {
		t.Fatal(err)
	}

	appliedBefore, err := r.AppliedMigrations(ctx, conn)
	if err != nil || len(appliedBefore) != 2 {
		t.Fatalf("expected 2 applied migrations, got %d (err: %v)", len(appliedBefore), err)
	}

	// Second run should be a clean no-op
	if err := r.Run(ctx, conn); err != nil {
		t.Fatalf("expected already-applied run to succeed cleanly: %v", err)
	}

	appliedAfter, err := r.AppliedMigrations(ctx, conn)
	if err != nil || len(appliedAfter) != 2 {
		t.Fatalf("expected still 2 applied migrations, got %d", len(appliedAfter))
	}
	for i := range appliedBefore {
		if appliedBefore[i] != appliedAfter[i] {
			t.Fatalf("applied migration %d mutated on second run", i)
		}
	}
}

// 3. Checksum mismatch
func TestRunner_ChecksumMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, conn, _ := freshTestConnection(t)

	all := defaultMigrations(t)
	r, err := NewRunner([]Migration{all[0]})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Run(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// Tamper with the migration definition (simulate modified script file with different checksum)
	tamperedSQL := all[0].SQL + "\n-- tampered comment"
	tamperedM := Migration{
		Version:  1,
		Name:     all[0].Name,
		Checksum: ComputeChecksum(tamperedSQL),
		SQL:      tamperedSQL,
	}
	rTampered, err := NewRunner([]Migration{tamperedM})
	if err != nil {
		t.Fatal(err)
	}

	err = rTampered.Run(ctx, conn)
	if err == nil {
		t.Fatal("expected error on checksum mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected error message to contain 'checksum mismatch', got: %v", err)
	}
}

// 4. Ordering
func TestRunner_Ordering(t *testing.T) {
	t.Parallel()
	all := defaultMigrations(t)

	// Non-contiguous gap [1, 3]
	m1 := all[0]
	m3 := Migration{
		Version:  3,
		Name:     "0003_future.sql",
		SQL:      "CREATE TABLE future (id INT);",
		Checksum: ComputeChecksum("CREATE TABLE future (id INT);"),
	}
	if _, err := NewRunner([]Migration{m1, m3}); err == nil {
		t.Fatal("expected error for non-contiguous migrations [1, 3]")
	}

	// Out-of-order [2, 1]
	m2 := all[1]
	if _, err := NewRunner([]Migration{m2, m1}); err == nil {
		t.Fatal("expected error for out-of-order migrations [2, 1]")
	}

	// Starting at version 2 instead of 1
	if _, err := NewRunner([]Migration{m2}); err == nil {
		t.Fatal("expected error for migrations starting at version 2")
	}
}

// 5. Partial failure (rollback leaves prior schema intact and usable)
func TestRunner_PartialFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, conn, _ := freshTestConnection(t)

	all := defaultMigrations(t)
	// Apply migration 1
	r1, err := NewRunner([]Migration{all[0]})
	if err != nil {
		t.Fatal(err)
	}
	if err := r1.Run(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// Insert conflicting legacy data that violates PreflightRevision2:
	// a source_detection in project 'p' referencing an artifact belonging to project 'q'.
	// Revision 1 allows this because artifacts.id exists, but PreflightRevision2 rejects
	// cross-project references before migration 2 DDL is executed.
	execTestSQL(t, conn, `
		INSERT INTO projects VALUES ('p','Project P','/p','INITIALIZED',0,NULL,NULL,'t','t');
		INSERT INTO projects VALUES ('q','Project Q','/q','INITIALIZED',0,NULL,NULL,'t','t');
		INSERT INTO artifacts VALUES ('a','p',printf('%064d',0),'a','application/json',2,'{}','t');
		INSERT INTO artifacts VALUES ('b','q',printf('%064d',1),'b','application/json',2,'{}','t');
		INSERT INTO source_detections VALUES ('d','p','hash','b','t');`)

	// Try to run migration 2 with default runner (which registers PreflightRevision2)
	rAll, err := NewRunner(all)
	if err != nil {
		t.Fatal(err)
	}

	err = rAll.Run(ctx, conn)
	if err == nil {
		t.Fatal("expected preflight failure on migration 2, got nil")
	}
	if !strings.Contains(err.Error(), "preflight") {
		t.Fatalf("expected error to identify preflight failure, got: %v", err)
	}

	// Prior schema must remain intact and usable
	applied, err := rAll.AppliedMigrations(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 1 {
		t.Fatalf("expected only 1 applied migration after partial failure, got %d", len(applied))
	}

	// Verify revision 2 tables were NOT created
	var count int
	if err := conn.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='approval_subjects'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("approval_subjects table should not exist after rollback, count = %d", count)
	}

	// Verify revision 1 tables remain usable
	var projCount int
	if err := conn.QueryRowContext(ctx, "SELECT count(*) FROM projects WHERE id='p'").Scan(&projCount); err != nil || projCount != 1 {
		t.Fatalf("revision 1 table was corrupted after rollback: %v", err)
	}
}

// 6. Restart/recovery
func TestRunner_RestartRecovery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, conn, dbPath := freshTestConnection(t)

	all := defaultMigrations(t)

	// Step 1: apply migration 1
	r1, err := NewRunner([]Migration{all[0]})
	if err != nil {
		t.Fatal(err)
	}
	if err := r1.Run(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// Step 2: close connection and database, simulating process termination
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	// Step 3: restart (reopen database and connection)
	dbRestart := openTestDB(t, dbPath)
	connRestart := testConnection(t, dbRestart)

	// Step 4: run full migration set
	rFull, err := NewRunner(all)
	if err != nil {
		t.Fatal(err)
	}
	if err := rFull.Run(ctx, connRestart); err != nil {
		t.Fatalf("restart recovery failed to apply pending migrations: %v", err)
	}

	applied, err := rFull.AppliedMigrations(ctx, connRestart)
	if err != nil || len(applied) != 2 {
		t.Fatalf("expected 2 applied migrations after recovery, got %d", len(applied))
	}

	// Verify revision 2 tables now exist
	var count int
	if err := connRestart.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='approval_subjects'").Scan(&count); err != nil || count != 1 {
		t.Fatalf("approval_subjects table should exist after recovery, count = %d", count)
	}
}

// 7. Concurrent runner attempts
func TestRunner_ConcurrentAttempts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "metadata.db")
	db := openTestDB(t, dbPath)

	// Configure initial connection to put the DB into WAL mode and establish pragma invariants
	initConn := testConnection(t, db)
	if err := ConfigureConnection(ctx, initConn); err != nil {
		t.Fatal(err)
	}
	if err := initConn.Close(); err != nil {
		t.Fatal(err)
	}

	all := defaultMigrations(t)

	const goroutines = 4
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn := testConnection(t, db)
			defer conn.Close()
			r, err := NewRunner(all)
			if err != nil {
				errs[idx] = err
				return
			}
			errs[idx] = r.Run(ctx, conn)
		}(i)
	}
	wg.Wait()

	// Every attempt should either succeed (clean migration or idempotent no-op)
	// or return a bounded busy/timeout error, without metadata corruption.
	for i, err := range errs {
		if err != nil {
			isBusy := false
			for cur := err; cur != nil; cur = errors.Unwrap(cur) {
				msg := strings.ToLower(cur.Error())
				if strings.Contains(msg, "locked") || strings.Contains(msg, "busy") {
					isBusy = true
					break
				}
			}
			if !isBusy {
				var rootErr error
				for cur := err; cur != nil; cur = errors.Unwrap(cur) {
					rootErr = cur
				}
				t.Fatalf("goroutine %d failed with unexpected error: %v (root cause: %v)", i, err, rootErr)
			}
		}
	}

	// Verify final database state reaches exactly 2 applied migrations
	verifyConn := testConnection(t, db)
	defer verifyConn.Close()
	r, _ := NewRunner(all)
	if err := r.Run(ctx, verifyConn); err != nil {
		t.Fatalf("subsequent migration failed after concurrent attempts: %v", err)
	}
	applied, err := r.AppliedMigrations(ctx, verifyConn)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 2 {
		t.Fatalf("expected 2 applied migrations after concurrent attempts, got %d", len(applied))
	}
}

// 8. Pre-existing database (user tables exist without schema_migrations)
func TestRunner_PreExistingDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, conn, _ := freshTestConnection(t)

	// Create an unmanaged user table without schema_migrations
	if err := ConfigureConnection(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, "CREATE TABLE custom_user_data (id INT PRIMARY KEY, name TEXT);"); err != nil {
		t.Fatal(err)
	}

	r, err := NewRunner(defaultMigrations(t))
	if err != nil {
		t.Fatal(err)
	}

	err = r.Run(ctx, conn)
	if err == nil {
		t.Fatal("expected error on pre-existing unmanaged database, got nil")
	}
	if !strings.Contains(err.Error(), "unmanaged database") {
		t.Fatalf("expected 'unmanaged database' error, got: %v", err)
	}
}

// 9. schema_migrations corruption/inconsistency
func TestRunner_SchemaMigrationsCorruption(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	all := defaultMigrations(t)

	// Scenario A: First migration is version 2 (missing version 1)
	t.Run("missing_version_1", func(t *testing.T) {
		t.Parallel()
		_, conn, _ := freshTestConnection(t)
		if err := ConfigureConnection(ctx, conn); err != nil {
			t.Fatal(err)
		}
		execTestSQL(t, conn, `
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY CHECK(version >= 1),
				name TEXT NOT NULL,
				checksum TEXT NOT NULL,
				applied_at TEXT NOT NULL,
				execution_time_ms INTEGER NOT NULL
			);
			INSERT INTO schema_migrations VALUES (2, '0002_revision_contracts.sql', '`+all[1].Checksum+`', '2026-10-05T00:00:00Z', 10);`)

		r, _ := NewRunner(all)
		err := r.Run(ctx, conn)
		if err == nil {
			t.Fatal("expected error when schema_migrations starts at version 2")
		}
		if !strings.Contains(err.Error(), "corruption") {
			t.Fatalf("expected corruption error, got: %v", err)
		}
	})

	// Scenario B: Non-contiguous versions (e.g. 1 and 3)
	t.Run("non_contiguous_versions", func(t *testing.T) {
		t.Parallel()
		_, conn, _ := freshTestConnection(t)
		if err := ConfigureConnection(ctx, conn); err != nil {
			t.Fatal(err)
		}
		execTestSQL(t, conn, `
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY CHECK(version >= 1),
				name TEXT NOT NULL,
				checksum TEXT NOT NULL,
				applied_at TEXT NOT NULL,
				execution_time_ms INTEGER NOT NULL
			);
			INSERT INTO schema_migrations VALUES (1, '0001_initial.sql', '`+all[0].Checksum+`', '2026-10-05T00:00:00Z', 10);
			INSERT INTO schema_migrations VALUES (3, '0003_future.sql', '`+strings.Repeat("a", 64)+`', '2026-10-05T00:00:00Z', 10);`)

		r, _ := NewRunner(all)
		err := r.Run(ctx, conn)
		if err == nil {
			t.Fatal("expected error for non-contiguous schema_migrations")
		}
		if !strings.Contains(err.Error(), "corruption") {
			t.Fatalf("expected corruption error, got: %v", err)
		}
	})

	// Scenario C: Unsupported future version (database version exceeds runner known migrations)
	t.Run("future_version_downgrade_rejected", func(t *testing.T) {
		t.Parallel()
		_, conn, _ := freshTestConnection(t)
		if err := ConfigureConnection(ctx, conn); err != nil {
			t.Fatal(err)
		}
		execTestSQL(t, conn, `
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY CHECK(version >= 1),
				name TEXT NOT NULL,
				checksum TEXT NOT NULL,
				applied_at TEXT NOT NULL,
				execution_time_ms INTEGER NOT NULL
			);
			INSERT INTO schema_migrations VALUES (1, '0001_initial.sql', '`+all[0].Checksum+`', '2026-10-05T00:00:00Z', 10);
			INSERT INTO schema_migrations VALUES (2, '0002_revision_contracts.sql', '`+all[1].Checksum+`', '2026-10-05T00:00:00Z', 10);
			INSERT INTO schema_migrations VALUES (3, '0003_future.sql', '`+strings.Repeat("a", 64)+`', '2026-10-05T00:00:00Z', 10);`)

		// Runner only knows up to version 2
		r, _ := NewRunner(all)
		err := r.Run(ctx, conn)
		if err == nil {
			t.Fatal("expected error when database has newer version than runner")
		}
		if !strings.Contains(err.Error(), "unsupported future schema version") {
			t.Fatalf("expected future schema version error, got: %v", err)
		}
	})
}

// 10. Immutability trigger enforcement on schema_migrations
func TestRunner_SchemaMigrationsAppendOnlyTriggers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, conn, _ := freshTestConnection(t)

	r, err := NewRunner(defaultMigrations(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Run(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// UPDATE must be rejected by trigger
	if _, err := conn.ExecContext(ctx, "UPDATE schema_migrations SET name='hacked.sql' WHERE version=1"); err == nil {
		t.Fatal("expected UPDATE on schema_migrations to be rejected by append-only trigger")
	}

	// DELETE must be rejected by trigger
	if _, err := conn.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version=1"); err == nil {
		t.Fatal("expected DELETE on schema_migrations to be rejected by append-only trigger")
	}
}
