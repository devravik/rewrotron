package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration represents a single versioned SQL migration script.
type Migration struct {
	Version  int
	Name     string
	Checksum string
	SQL      string
}

// AppliedMigration represents a migration record stored in schema_migrations.
type AppliedMigration struct {
	Version         int
	Name            string
	Checksum        string
	AppliedAt       string
	ExecutionTimeMs int64
}

// MigrationError identifies a migration execution, admission, or validation failure.
type MigrationError struct {
	Check string
	Cause error
}

// Error describes the migration policy failure without exposing private data.
func (e *MigrationError) Error() string {
	if e.Cause != nil {
		return "migration runner: " + e.Check + ": " + e.Cause.Error()
	}
	return "migration runner: " + e.Check
}

// Unwrap preserves the underlying database or operational error.
func (e *MigrationError) Unwrap() error { return e.Cause }

// PreflightHook defines an optional validation check executed within the migration
// transaction before executing the migration's DDL.
type PreflightHook func(ctx context.Context, tx *sql.Tx) error

// Runner executes versioned migrations transactionally and manages schema_migrations.
type Runner struct {
	migrations []Migration
	preflights map[int]PreflightHook
}

// ComputeChecksum calculates the SHA-256 digest of the raw migration script SQL bytes,
// returning a 64-character lowercase hexadecimal string conforming to SQLITE-SCHEMA.md.
func ComputeChecksum(sqlContent string) string {
	digest := sha256.Sum256([]byte(sqlContent))
	return hex.EncodeToString(digest[:])
}

func isDigest(val string) bool {
	if len(val) != 64 {
		return false
	}
	for i := 0; i < len(val); i++ {
		b := val[i]
		if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')) {
			return false
		}
	}
	return true
}

// NewRunner creates a migration runner for the provided ordered migrations.
// It verifies that migrations are non-empty, start at version 1, are strictly
// contiguous without gaps or duplicate versions, have valid 64-hex checksums
// matching their SQL contents, and have non-empty names.
func NewRunner(migrations []Migration) (*Runner, error) {
	if len(migrations) == 0 {
		return nil, &MigrationError{Check: "no migrations provided"}
	}

	for i, m := range migrations {
		expectedVer := i + 1
		if m.Version != expectedVer {
			return nil, &MigrationError{Check: fmt.Sprintf("invalid migration version ordering: expected %d, got %d", expectedVer, m.Version)}
		}
		if len(m.Name) == 0 {
			return nil, &MigrationError{Check: fmt.Sprintf("migration %d: name cannot be empty", m.Version)}
		}
		if len(m.SQL) == 0 {
			return nil, &MigrationError{Check: fmt.Sprintf("migration %d: SQL content cannot be empty", m.Version)}
		}
		if !isDigest(m.Checksum) {
			return nil, &MigrationError{Check: fmt.Sprintf("migration %d: invalid checksum format", m.Version)}
		}
		computed := ComputeChecksum(m.SQL)
		if m.Checksum != computed {
			return nil, &MigrationError{Check: fmt.Sprintf("migration %d: checksum %s does not match content hash %s", m.Version, m.Checksum, computed)}
		}
	}

	r := &Runner{
		migrations: make([]Migration, len(migrations)),
		preflights: make(map[int]PreflightHook),
	}
	copy(r.migrations, migrations)

	// Revision 2 requires PreflightRevision2 per SQLITE-SCHEMA.md and RECOVERY.md.
	r.preflights[2] = PreflightRevision2

	return r, nil
}

// RegisterPreflight registers a preflight hook for a specific migration version.
func (r *Runner) RegisterPreflight(version int, hook PreflightHook) {
	r.preflights[version] = hook
}

// Migrations returns a copy of the runner's configured migrations.
func (r *Runner) Migrations() []Migration {
	cp := make([]Migration, len(r.migrations))
	copy(cp, r.migrations)
	return cp
}

// LoadMigrationsFromDir loads all .sql migration files from a directory, sorts them
// by numeric version prefix, verifies contiguous numbering starting at 1, and computes checksums.
func LoadMigrationsFromDir(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, &MigrationError{Check: "read migrations directory", Cause: err}
	}

	type fileItem struct {
		version  int
		filename string
		path     string
	}
	var items []fileItem

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		name := entry.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		ver, err := strconv.Atoi(parts[0])
		if err != nil || ver < 1 {
			continue
		}
		items = append(items, fileItem{
			version:  ver,
			filename: name,
			path:     filepath.Join(dir, name),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].version < items[j].version
	})

	if len(items) == 0 {
		return nil, &MigrationError{Check: "no .sql migrations found in directory"}
	}

	var migrations []Migration
	for i, it := range items {
		expectedVer := i + 1
		if it.version != expectedVer {
			return nil, &MigrationError{Check: fmt.Sprintf("non-contiguous migration version: expected %d, found %d (%s)", expectedVer, it.version, it.filename)}
		}
		data, err := os.ReadFile(it.path)
		if err != nil {
			return nil, &MigrationError{Check: fmt.Sprintf("read migration file %s", it.filename), Cause: err}
		}
		sqlContent := string(data)
		migrations = append(migrations, Migration{
			Version:  it.version,
			Name:     it.filename,
			Checksum: ComputeChecksum(sqlContent),
			SQL:      sqlContent,
		})
	}

	return migrations, nil
}

// Run executes pending migrations against the provided database connection.
// It verifies connection policy, checks for unmanaged pre-existing tables,
// validates already-applied migrations for contiguity and checksum integrity,
// executes preflight checks inside migration transactions, and updates schema_migrations.
func (r *Runner) Run(ctx context.Context, conn *sql.Conn) error {
	if conn == nil {
		return &MigrationError{Check: "connection required"}
	}

	// 1. Verify or establish connection pragmas (foreign_keys=ON, WAL, synchronous=FULL, busy_timeout=5000).
	if err := ConfigureConnection(ctx, conn); err != nil {
		return &MigrationError{Check: "configure connection", Cause: err}
	}

	// 2. Inspect existing tables in database.
	rows, err := conn.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return &MigrationError{Check: "inspect existing tables", Cause: err}
	}
	defer rows.Close()

	hasSchemaMigrations := false
	hasUserTables := false
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return &MigrationError{Check: "scan table name", Cause: err}
		}
		if name == "schema_migrations" {
			hasSchemaMigrations = true
		} else {
			hasUserTables = true
		}
	}
	if err := rows.Err(); err != nil {
		return &MigrationError{Check: "iterate table names", Cause: err}
	}

	// Admission check: unmanaged pre-existing database.
	if hasUserTables && !hasSchemaMigrations {
		return &MigrationError{Check: "unmanaged database: existing user tables detected without schema_migrations table"}
	}

	// 3. Ensure schema_migrations table and its append-only triggers exist.
	if !hasSchemaMigrations {
		const initDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY CHECK(version >= 1),
    name TEXT NOT NULL,
    checksum TEXT NOT NULL CHECK(length(checksum)=64 AND checksum NOT GLOB '*[^a-f0-9]*'),
    applied_at TEXT NOT NULL,
    execution_time_ms INTEGER NOT NULL CHECK(execution_time_ms >= 0)
);
CREATE TRIGGER IF NOT EXISTS schema_migrations_no_update BEFORE UPDATE ON schema_migrations
BEGIN SELECT RAISE(ABORT, 'schema_migrations is append-only'); END;
CREATE TRIGGER IF NOT EXISTS schema_migrations_no_delete BEFORE DELETE ON schema_migrations
BEGIN SELECT RAISE(ABORT, 'schema_migrations is append-only'); END;`
		if _, err := conn.ExecContext(ctx, initDDL); err != nil {
			return &MigrationError{Check: "create schema_migrations table", Cause: err}
		}
	}

	// 4. Query and validate already-applied migrations.
	applied, err := r.AppliedMigrations(ctx, conn)
	if err != nil {
		return err
	}

	for i, a := range applied {
		expectedVer := i + 1
		if a.Version != expectedVer {
			return &MigrationError{Check: fmt.Sprintf("schema_migrations corruption: expected version %d at index %d, found %d", expectedVer, i, a.Version)}
		}
		if a.Version > len(r.migrations) {
			return &MigrationError{Check: fmt.Sprintf("unsupported future schema version: database version %d exceeds runner known migrations (%d)", a.Version, len(r.migrations))}
		}
		exp := r.migrations[i]
		if a.Checksum != exp.Checksum {
			return &MigrationError{Check: fmt.Sprintf("checksum mismatch for migration %04d (%s): recorded %s != expected %s", a.Version, a.Name, a.Checksum, exp.Checksum)}
		}
		if a.Name != exp.Name {
			return &MigrationError{Check: fmt.Sprintf("name mismatch for migration %04d: recorded %s != expected %s", a.Version, a.Name, exp.Name)}
		}
	}

	// 5. Apply pending migrations incrementally.
	appliedCount := len(applied)
	if appliedCount == len(r.migrations) {
		return nil // Up-to-date
	}

	for idx := appliedCount; idx < len(r.migrations); idx++ {
		m := r.migrations[idx]
		if err := r.applyOne(ctx, conn, m); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) applyOne(ctx context.Context, conn *sql.Conn, m Migration) error {
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return &MigrationError{Check: fmt.Sprintf("begin transaction for migration %04d", m.Version), Cause: err}
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Check if a concurrent runner already applied this migration.
	var recordedChecksum string
	err = tx.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE version = ?", m.Version).Scan(&recordedChecksum)
	if err == nil {
		if recordedChecksum != m.Checksum {
			return &MigrationError{Check: fmt.Sprintf("checksum mismatch for migration %04d: recorded %s != expected %s", m.Version, recordedChecksum, m.Checksum)}
		}
		// Already applied by concurrent runner; nothing to do.
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return &MigrationError{Check: fmt.Sprintf("check applied status for migration %04d", m.Version), Cause: err}
	}

	// Execute preflight if registered for this version.
	if hook, exists := r.preflights[m.Version]; exists && hook != nil {
		if err := hook(ctx, tx); err != nil {
			return &MigrationError{Check: fmt.Sprintf("preflight migration %04d (%s)", m.Version, m.Name), Cause: err}
		}
	}

	start := time.Now()
	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return &MigrationError{Check: fmt.Sprintf("execute migration %04d (%s)", m.Version, m.Name), Cause: err}
	}
	durationMs := time.Since(start).Milliseconds()

	now := time.Now().UTC().Format(time.RFC3339)
	const recordSQL = `INSERT INTO schema_migrations (version, name, checksum, applied_at, execution_time_ms) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, recordSQL, m.Version, m.Name, m.Checksum, now, durationMs); err != nil {
		return &MigrationError{Check: fmt.Sprintf("record migration %04d in schema_migrations", m.Version), Cause: err}
	}

	if err := tx.Commit(); err != nil {
		return &MigrationError{Check: fmt.Sprintf("commit migration %04d", m.Version), Cause: err}
	}

	return nil
}

// AppliedMigrations returns all recorded migrations from schema_migrations in version order.
// If schema_migrations does not exist, an empty slice is returned.
func (r *Runner) AppliedMigrations(ctx context.Context, conn *sql.Conn) ([]AppliedMigration, error) {
	if conn == nil {
		return nil, &MigrationError{Check: "connection required"}
	}

	var count int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&count); err != nil {
		return nil, &MigrationError{Check: "check schema_migrations existence", Cause: err}
	}
	if count == 0 {
		return nil, nil
	}

	rows, err := conn.QueryContext(ctx, `SELECT version, name, checksum, applied_at, execution_time_ms FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, &MigrationError{Check: "query schema_migrations", Cause: err}
	}
	defer rows.Close()

	var applied []AppliedMigration
	for rows.Next() {
		var a AppliedMigration
		if err := rows.Scan(&a.Version, &a.Name, &a.Checksum, &a.AppliedAt, &a.ExecutionTimeMs); err != nil {
			return nil, &MigrationError{Check: "scan schema_migrations row", Cause: err}
		}
		applied = append(applied, a)
	}
	if err := rows.Err(); err != nil {
		return nil, &MigrationError{Check: "iterate schema_migrations rows", Cause: err}
	}

	return applied, nil
}
