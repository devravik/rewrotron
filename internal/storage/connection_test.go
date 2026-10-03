package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func testConnection(t *testing.T, db *sql.DB) *sql.Conn {
	t.Helper()
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Error(err)
		}
	})
	return conn
}

func TestConfigureConnection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, filepath.Join(t.TempDir(), "metadata.db"))
	first := testConnection(t, db)
	if err := ConfigureConnection(ctx, first); err != nil {
		t.Fatal(err)
	}
	// Holding the first connection forces acquisition of a distinct physical one.
	second := testConnection(t, db)
	for _, conn := range []*sql.Conn{first, second} {
		if err := ConfigureConnection(ctx, conn); err != nil {
			t.Fatal(err)
		}
		var version, mode string
		if err := conn.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
			t.Fatal(err)
		}
		if version != "3.51.3" {
			t.Fatalf("engine = %q", version)
		}
		if err := conn.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
			t.Fatal(err)
		}
		if mode != "wal" {
			t.Fatalf("journal = %q", mode)
		}
		for name, expected := range map[string]int{"foreign_keys": 1, "synchronous": 2, "busy_timeout": 5000} {
			var value int
			if err := conn.QueryRowContext(ctx, "PRAGMA "+name).Scan(&value); err != nil {
				t.Fatal(err)
			}
			if value != expected {
				t.Fatalf("%s = %d, want %d", name, value, expected)
			}
		}
	}
	if _, err := first.ExecContext(ctx, "CREATE TABLE parent(id INTEGER PRIMARY KEY); CREATE TABLE child(parent_id INTEGER REFERENCES parent(id))"); err != nil {
		t.Fatal(err)
	}
	for _, conn := range []*sql.Conn{first, second} {
		if _, err := conn.ExecContext(ctx, "INSERT INTO child VALUES (999)"); err == nil {
			t.Fatal("foreign key violation accepted")
		}
	}
	if _, err := first.ExecContext(ctx, "INSERT INTO parent VALUES (1)"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.ExecContext(ctx, "INSERT INTO child VALUES (1)"); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureConnectionRejectsMemoryJournal(t *testing.T) {
	t.Parallel()
	conn := testConnection(t, openTestDB(t, ":memory:"))
	err := ConfigureConnection(context.Background(), conn)
	var policy *ConnectionError
	if !errors.As(err, &policy) || policy.Check != "WAL unavailable" {
		t.Fatalf("want WAL refusal, got %v", err)
	}
}

func TestConfigureConnectionRejectsForeignKeysInsideTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn := testConnection(t, openTestDB(t, filepath.Join(t.TempDir(), "metadata.db")))
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
	}()
	err = ConfigureConnection(ctx, conn)
	var policy *ConnectionError
	if !errors.As(err, &policy) || policy.Check != "foreign_keys policy not established" {
		t.Fatalf("want foreign key refusal, got %v", err)
	}
}

func TestConfigureConnectionCancellation(t *testing.T) {
	t.Parallel()
	conn := testConnection(t, openTestDB(t, filepath.Join(t.TempDir(), "metadata.db")))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ConfigureConnection(ctx, conn)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want cancellation, got %v", err)
	}
	var policy *ConnectionError
	if !errors.As(err, &policy) || policy.Check != "engine version query" {
		t.Fatalf("want structured version-query error, got %v", err)
	}
}

func TestConfigureConnectionReopen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.db")
	first := openTestDB(t, path)
	conn := testConnection(t, first)
	if err := ConfigureConnection(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, "CREATE TABLE checkpoint(value TEXT NOT NULL); INSERT INTO checkpoint VALUES ('retained')"); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second := testConnection(t, openTestDB(t, path))
	if err := ConfigureConnection(ctx, second); err != nil {
		t.Fatal(err)
	}
	var value string
	if err := second.QueryRowContext(ctx, "SELECT value FROM checkpoint").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "retained" {
		t.Fatalf("reopened value = %q", value)
	}
}

func TestConfigureConnectionRequiresConnection(t *testing.T) {
	err := ConfigureConnection(context.Background(), nil)
	var policy *ConnectionError
	if !errors.As(err, &policy) || policy.Check != "connection required" {
		t.Fatalf("want connection refusal, got %v", err)
	}
}
