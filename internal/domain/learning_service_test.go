package domain

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devravik/rewrotron/internal/storage"
	_ "modernc.org/sqlite"
)

func setupTestLearningDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "learning_test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	defer conn.Close()

	if err := storage.ConfigureConnection(ctx, conn); err != nil {
		t.Fatalf("failed to configure connection: %v", err)
	}

	migrations, err := storage.LoadMigrationsFromDir(filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("failed to load migrations: %v", err)
	}
	runner, err := storage.NewRunner(migrations)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}
	if err := runner.Run(ctx, conn); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Insert test projects
	_, err = db.ExecContext(ctx, "INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES ('proj-a', 'Project A', '/src/a', 'INITIALIZED', '2026-10-10T00:00:00Z', '2026-10-10T00:00:00Z')")
	if err != nil {
		t.Fatalf("failed to insert proj-a: %v", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES ('proj-b', 'Project B', '/src/b', 'INITIALIZED', '2026-10-10T00:00:00Z', '2026-10-10T00:00:00Z')")
	if err != nil {
		t.Fatalf("failed to insert proj-b: %v", err)
	}

	return db, ctx
}

func TestLearningService_ExperimentFingerprintReproducibility(t *testing.T) {
	expCtx1 := &ExperimentContext{
		SourceConfiguration: "ci3-default",
		PlanHash:            "plan123",
		CorpusHash:          "corpus456",
		PolicyHash:          "policy789",
		CaseHash:            "case01",
		WorkflowHash:        "wf01",
		FixtureHash:         "fix01",
		TargetBeforeTreeHash: "before01",
		AdapterVersions:     map[string]string{"php": "8.3", "ci3": "3.1.13"},
		ToolVersions:        map[string]string{"phpunit": "10.0"},
		DependencyVersions:  map[string]string{"composer": "2.7"},
		HypothesisPatchHash: "patch01",
		SandboxIdentity:     "sb-01",
	}

	// Reorder maps to ensure RFC 8785 canonicalization guarantees identical fingerprint
	expCtx2 := &ExperimentContext{
		SourceConfiguration: "ci3-default",
		PlanHash:            "plan123",
		CorpusHash:          "corpus456",
		PolicyHash:          "policy789",
		CaseHash:            "case01",
		WorkflowHash:        "wf01",
		FixtureHash:         "fix01",
		TargetBeforeTreeHash: "before01",
		AdapterVersions:     map[string]string{"ci3": "3.1.13", "php": "8.3"},
		ToolVersions:        map[string]string{"phpunit": "10.0"},
		DependencyVersions:  map[string]string{"composer": "2.7"},
		HypothesisPatchHash: "patch01",
		SandboxIdentity:     "sb-01",
	}

	fp1, _, err1 := expCtx1.Fingerprint()
	if err1 != nil {
		t.Fatalf("fp1 err: %v", err1)
	}
	fp2, _, err2 := expCtx2.Fingerprint()
	if err2 != nil {
		t.Fatalf("fp2 err: %v", err2)
	}

	if fp1 != fp2 {
		t.Fatalf("expected identical fingerprints, got %s and %s", fp1, fp2)
	}

	// Changed input creates a new fingerprint
	expCtx3 := *expCtx1
	expCtx3.HypothesisPatchHash = "patch02"
	fp3, _, _ := expCtx3.Fingerprint()
	if fp3 == fp1 {
		t.Fatalf("expected different fingerprint when patch changed, got same %s", fp3)
	}
}

func TestLearningService_SanitizationBoundary(t *testing.T) {
	sanitizer := NewSanitizer([]byte("test-hmac-key"))

	rawText := "Error authenticating with api_key: 'sk_live_secret123456789' and password: 'SuperSecretPassword123' for user john.doe@example.com at /home/ravi/project/private/config.php"

	cleanText, modified := sanitizer.SanitizeText(rawText)
	if !modified {
		t.Fatalf("expected text to be sanitized")
	}

	// Verify plaintext secret, email, and private host path are stripped
	if strings.Contains(cleanText, "sk_live_secret123456789") {
		t.Errorf("secret API key was not sanitized: %s", cleanText)
	}
	if strings.Contains(cleanText, "SuperSecretPassword123") {
		t.Errorf("secret password was not sanitized: %s", cleanText)
	}
	if strings.Contains(cleanText, "john.doe@example.com") {
		t.Errorf("PII email was not sanitized: %s", cleanText)
	}
	if strings.Contains(cleanText, "/home/ravi/project/private/config.php") {
		t.Errorf("private filesystem path was not sanitized: %s", cleanText)
	}

	// Verify keyed HMAC token is present for correlation
	if !strings.Contains(cleanText, "[REDACTED:hmac_") {
		t.Errorf("expected HMAC correlation token, got: %s", cleanText)
	}
}

func TestLearningService_ProjectLocalLessonAndCrossProjectIsolation(t *testing.T) {
	db, ctx := setupTestLearningDB(t)
	defer db.Close()

	sanitizer := NewSanitizer([]byte("test-hmac-key"))
	svc := NewLearningService(db, sanitizer)

	// 1. Record experiment in proj-a
	expCtx := &ExperimentContext{
		SourceConfiguration: "ci3-default",
		PlanHash:            "planhash",
		CorpusHash:          "corpus",
		PolicyHash:          "policy",
		CaseHash:            "case",
		WorkflowHash:        "wf",
		FixtureHash:         "fix",
		TargetBeforeTreeHash: "tree",
		AdapterVersions:     map[string]string{"ci3": "3.1.13"},
		HypothesisPatchHash: "patch",
		SandboxIdentity:     "sb1",
	}

	_, err := svc.RecordExperiment(ctx, "proj-a", "exp-1", expCtx, "FAILED")
	if err != nil {
		t.Fatalf("failed to record experiment: %v", err)
	}

	// 2. Record negative lesson in proj-a with embedded secret that must be sanitized
	finding := "Dynamic framework hook in CI3 controller failed with password=Secret999888777"
	remediation := "Use explicit middleware route binding without dynamic hook assumption"

	lesson, err := svc.RecordLesson(ctx, "proj-a", "les-1", "exp-1", finding, remediation, nil)
	if err != nil {
		t.Fatalf("failed to record lesson: %v", err)
	}

	if strings.Contains(lesson.Finding, "Secret999888777") {
		t.Fatalf("lesson finding was not sanitized: %s", lesson.Finding)
	}
	if !strings.Contains(lesson.Finding, "[REDACTED:hmac_") {
		t.Fatalf("expected HMAC token in sanitized finding, got: %s", lesson.Finding)
	}

	// 3. Immutability trigger verification: attempting to UPDATE or DELETE lessons must abort
	_, err = db.ExecContext(ctx, "UPDATE lessons SET finding = 'mutated' WHERE id = 'les-1'")
	if err == nil {
		t.Fatalf("expected error updating append-only lessons table, got nil")
	}
	_, err = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = 'les-1'")
	if err == nil {
		t.Fatalf("expected error deleting from append-only lessons table, got nil")
	}

	// 4. Cross-project isolation: proj-b cannot record lesson against proj-a experiment
	_, err = svc.RecordLesson(ctx, "proj-b", "les-cross", "exp-1", "Finding", "Remediation", nil)
	if err == nil {
		t.Fatalf("expected cross-project violation recording lesson, got nil")
	}

	// 5. Cross-project query isolation: proj-b cannot search proj-a lessons
	_, err = svc.SearchProjectLessons(ctx, "proj-b", "proj-a", "framework hook")
	if err == nil {
		t.Fatalf("expected cross-project search to be denied, got nil")
	}

	// 6. Proj-a can search its own lessons
	matches, err := svc.SearchProjectLessons(ctx, "proj-a", "proj-a", "framework hook")
	if err != nil {
		t.Fatalf("failed to search proj-a lessons: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match for proj-a, got %d", len(matches))
	}
}
