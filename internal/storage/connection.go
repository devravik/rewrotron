// Package storage implements local metadata persistence primitives.
package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLiteVersion is the required embedded engine version in SQLITE-SCHEMA.
const SQLiteVersion = "3.51.3"

// ConnectionError identifies the failed connection-policy check.
type ConnectionError struct {
	Check string
	Cause error
}

// Error describes a policy failure without including the database path.
func (e *ConnectionError) Error() string {
	return "SQLite connection policy: " + e.Check
}

// Unwrap preserves operational errors for cancellation and retry inspection.
func (e *ConnectionError) Unwrap() error { return e.Cause }

// ConfigureConnection establishes and verifies the metadata connection policy.
// The caller must exclusively own a fresh connection outside a transaction,
// opened only after filesystem/path admission. Every new physical connection
// requires this initialization before use; configuring one does not qualify a
// pool. On error the caller must discard the connection, which may be partially
// configured. This primitive grants no workspace, migration or gate authority.
func ConfigureConnection(ctx context.Context, conn *sql.Conn) error {
	if conn == nil {
		return &ConnectionError{Check: "connection required"}
	}
	var version string
	if err := conn.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
		return &ConnectionError{Check: "engine version query", Cause: err}
	}
	if version != SQLiteVersion {
		return &ConnectionError{Check: "unsupported engine version"}
	}

	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		return &ConnectionError{Check: "enable foreign keys", Cause: err}
	}
	if err := requirePragma(ctx, conn, "foreign_keys", 1); err != nil {
		return err
	}

	var mode string
	if err := conn.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&mode); err != nil {
		return &ConnectionError{Check: "enable WAL", Cause: err}
	}
	if mode != "wal" {
		return &ConnectionError{Check: "WAL unavailable"}
	}

	if _, err := conn.ExecContext(ctx, "PRAGMA synchronous=FULL"); err != nil {
		return &ConnectionError{Check: "set synchronous", Cause: err}
	}
	if err := requirePragma(ctx, conn, "synchronous", 2); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA busy_timeout=5000"); err != nil {
		return &ConnectionError{Check: "set busy timeout", Cause: err}
	}
	return requirePragma(ctx, conn, "busy_timeout", 5000)
}

func requirePragma(ctx context.Context, conn *sql.Conn, name string, expected int) error {
	// Names are private fixed literals, never caller-controlled SQL.
	var value int
	if err := conn.QueryRowContext(ctx, "PRAGMA "+name).Scan(&value); err != nil {
		return &ConnectionError{Check: "read " + name, Cause: err}
	}
	if value != expected {
		return &ConnectionError{Check: fmt.Sprintf("%s policy not established", name)}
	}
	return nil
}
