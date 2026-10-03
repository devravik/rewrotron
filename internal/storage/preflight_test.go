package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func migrationSQL(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func revision1Connection(t *testing.T) *sql.Conn {
	t.Helper()
	ctx := context.Background()
	conn := testConnection(t, openTestDB(t, filepath.Join(t.TempDir(), "metadata.db")))
	if err := ConfigureConnection(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, migrationSQL(t, "0001_initial.sql")); err != nil {
		t.Fatal(err)
	}
	// Synthetic metadata only; these rows do not establish runtime approvals.
	execTestSQL(t, conn, `
		INSERT INTO projects VALUES ('p','Project','/p','INITIALIZED',0,NULL,NULL,'t','t');
		INSERT INTO projects VALUES ('q','Other','/q','INITIALIZED',0,NULL,NULL,'t','t');
		INSERT INTO artifacts VALUES ('a','p',printf('%064d',0),'a','application/json',2,'{}','t');
		INSERT INTO artifacts VALUES ('b','q',printf('%064d',1),'b','application/json',2,'{}','t');
		INSERT INTO operations VALUES ('op','p','ASSESS','COMPLETED','{}',NULL,'t','t');
		INSERT INTO operations VALUES ('other-op','q','ASSESS','COMPLETED','{}',NULL,'t','t');`)
	return conn
}

func execTestSQL(t *testing.T, conn *sql.Conn, query string) {
	t.Helper()
	if _, err := conn.ExecContext(context.Background(), query); err != nil {
		t.Fatal(err)
	}
}

func preflightTestTx(t *testing.T, conn *sql.Conn) *sql.Tx {
	t.Helper()
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Error(err)
		}
	})
	return tx
}

func assertPreflight(t *testing.T, conn *sql.Conn, expected string) {
	t.Helper()
	tx := preflightTestTx(t, conn)
	var before, after int64
	if err := tx.QueryRow("SELECT total_changes()").Scan(&before); err != nil {
		t.Fatal(err)
	}
	err := PreflightRevision2(context.Background(), tx)
	if expected == "" {
		if err != nil {
			t.Fatal(err)
		}
	} else {
		var conflict *PreflightError
		if !errors.As(err, &conflict) || conflict.Check != expected || conflict.Cause != nil {
			t.Fatalf("want data conflict %q, got %v", expected, err)
		}
	}
	if err := tx.QueryRow("SELECT total_changes()").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("preflight mutated data")
	}
	var tables int
	if err := tx.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name='approval_subjects'").Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if tables != 0 {
		t.Fatal("preflight applied revision-2 DDL")
	}
}

func TestPreflightRevision2CleanAndNullReferences(t *testing.T) {
	t.Parallel()
	conn := revision1Connection(t)
	execTestSQL(t, conn, `
		INSERT INTO source_detections VALUES ('d','p','hash','a','t');
		INSERT INTO idempotency_requests VALUES ('p','request','hash','op',NULL);
		INSERT INTO artifact_links VALUES ('p','test','entity','a');`)
	assertPreflight(t, conn, "")
}

func TestPreflightRevision2Ownership(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, mutant, check string }{
		{"detection", "INSERT INTO source_detections VALUES ('d','p','hash','b','t')", "source_detections.result_ref ownership"},
		{"artifact link", "INSERT INTO artifact_links VALUES ('p','test','entity','b')", "artifact_links.artifact_id ownership"},
		{"operation", "INSERT INTO idempotency_requests VALUES ('p','request','hash','other-op',NULL)", "idempotency_requests.operation_id ownership"},
		{"nullable result", "INSERT INTO idempotency_requests VALUES ('p','request','hash','op','b')", "idempotency_requests.result_ref ownership"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			conn := revision1Connection(t)
			execTestSQL(t, conn, test.mutant)
			assertPreflight(t, conn, test.check)
		})
	}
}

func TestPreflightRevision2MissionIdentity(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"QUEUED", "LEASED", "RUNNING", "BLOCKED", "COMPLETED", "FAILED", "CANCELLED"} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			conn := revision1Connection(t)
			for _, id := range []string{"m1", "m2"} {
				_, err := conn.ExecContext(context.Background(), `INSERT INTO missions
					(id,project_id,type,priority,status,context_json,fingerprint,lease_owner,lease_expires_at,created_at)
					VALUES (?,'p','TEST',1,?,'{}','work','owner','t','t')`, id, status)
				if err != nil {
					t.Fatal(err)
				}
			}
			expected := ""
			if status == "QUEUED" || status == "LEASED" || status == "RUNNING" || status == "BLOCKED" {
				expected = "duplicate active mission work"
			}
			assertPreflight(t, conn, expected)
		})
	}
}

func TestPreflightRevision2DistinctMissionWork(t *testing.T) {
	t.Parallel()
	conn := revision1Connection(t)
	execTestSQL(t, conn, `
		INSERT INTO missions(id,project_id,type,priority,status,context_json,fingerprint,created_at) VALUES ('m1','p','TEST',1,'QUEUED','{}','work','t');
		INSERT INTO missions(id,project_id,type,priority,status,context_json,fingerprint,created_at) VALUES ('m2','q','TEST',1,'QUEUED','{}','work','t');
		INSERT INTO missions(id,project_id,type,priority,status,context_json,fingerprint,created_at) VALUES ('m3','p','OTHER',1,'QUEUED','{}','work','t');
		INSERT INTO missions(id,project_id,type,priority,status,context_json,fingerprint,created_at) VALUES ('m4','p','TEST',1,'QUEUED','{}','other-work','t');`)
	assertPreflight(t, conn, "")
}

func TestPreflightRevision2DuplicateCaseResults(t *testing.T) {
	t.Parallel()
	conn := revision1Connection(t)
	execTestSQL(t, conn, `
		INSERT INTO source_detections VALUES ('d','p','hash','a','t');
		INSERT INTO baselines VALUES ('baseline','p','d','a','{}','{}','environment','t');
		INSERT INTO contracts VALUES ('contract','p','baseline','operation','2.0',1,'a','hash','LOCKED','t');
		INSERT INTO behavior_cases VALUES ('case','p','contract','a','hash','fixture','policy','t');
		INSERT INTO migration_plans VALUES ('plan','p','baseline','plan-hash','corpus','policy','a','t');
		INSERT INTO observations VALUES ('source','p','case','a','source_legacy','{}','t');
		INSERT INTO observations VALUES ('target','p','case','a','target_modern','{}','t');
		INSERT INTO verification_runs VALUES ('run','p','plan','target','corpus','policy','COMPLETED','UNKNOWN','a','t');
		INSERT INTO case_results VALUES ('r1','p','run','case','source','target','UNKNOWN','{}');`)
	// A single retained result is admissible, independent of its verdict.
	tx := preflightTestTx(t, conn)
	if err := PreflightRevision2(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	execTestSQL(t, conn, "INSERT INTO case_results VALUES ('r2','p','run','case','source','target','UNKNOWN','{}')")
	assertPreflight(t, conn, "duplicate case results")
}

func TestPreflightRevision2OperationalErrors(t *testing.T) {
	t.Parallel()
	if err := PreflightRevision2(context.Background(), nil); err == nil {
		t.Fatal("nil transaction accepted")
	}
	conn := revision1Connection(t)
	tx := preflightTestTx(t, conn)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := PreflightRevision2(ctx, tx); !errors.Is(err, context.Canceled) {
		t.Fatalf("want cancellation, got %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := PreflightRevision2(context.Background(), tx); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("want closed transaction failure, got %v", err)
	}
}

func TestPreflightRevision2ReferencesMatchMigration(t *testing.T) {
	// Check coverage against all actual revision-2 owner-insert predicates.
	pattern := regexp.MustCompile(`CREATE TRIGGER (\w+)_owner_insert BEFORE INSERT ON (\w+)\nWHEN NEW\.(\w+) IS NOT NULL AND NOT EXISTS \(SELECT 1 FROM (\w+) WHERE id=NEW\.(\w+) AND project_id=NEW\.project_id\)`)
	matches := pattern.FindAllStringSubmatch(migrationSQL(t, "0002_revision_contracts.sql"), -1)
	refs := revision2References()
	if len(matches) != len(refs) || len(matches) != 21 {
		t.Fatal("preflight ownership coverage does not match revision 2")
	}
	for i, match := range matches {
		if refs[i].table != match[2] || refs[i].column != match[3] || refs[i].parent != match[4] || match[3] != match[5] {
			t.Fatalf("ownership check %d differs from migration", i)
		}
	}
}
