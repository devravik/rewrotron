package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestSupervisor_X26_X27_X32 tests Milestone M5 Unit 4:
// - X26: Process lifetime (kill foreground owner, kill MCP client with running supervisor, human pause)
// - X27: Execution truth (read old through_sequence while rows advance; event store failure stops admission)
// - X32: Trust boundary (mediated service revalidates hostile same-account modification)
func TestSupervisor_X26_X27_X32(t *testing.T) {
	now := time.Now()

	// =========================================================================
	// Test X26: Process Lifetime
	// =========================================================================
	t.Run("X26_ProcessLifetime", func(t *testing.T) {
		svc := NewSupervisorService()
		projectID := "proj_x26"

		// 1. Register foreground CLI (PID 100) and its worker runner (PID 101)
		_, err := svc.RegisterProcess(100, ProcessTypeForegroundCLI, projectID, "op_fg_1", now)
		if err != nil {
			t.Fatalf("RegisterProcess foreground failed: %v", err)
		}
		_, err = svc.RegisterProcess(101, ProcessTypeWorkerRunner, projectID, "op_fg_1", now)
		if err != nil {
			t.Fatalf("RegisterProcess worker failed: %v", err)
		}

		// Killing foreground CLI terminates its owned child worker runner
		err = svc.KillProcess(100, 1, "SIGTERM", now)
		if err != nil {
			t.Fatalf("KillProcess failed: %v", err)
		}
		if svc.IsProcessAlive(100) {
			t.Fatalf("expected foreground CLI to be dead")
		}
		if svc.IsProcessAlive(101) {
			t.Fatalf("expected owned worker runner to be killed when foreground CLI terminated")
		}

		// 2. Supervisor survives MCP client disconnect
		// MCP client (PID 200), Supervisor (PID 201), Background Runner (PID 202)
		_, _ = svc.RegisterProcess(200, ProcessTypeForegroundCLI, projectID, "op_mcp", now)
		_, _ = svc.RegisterProcess(201, ProcessTypeSupervisor, projectID, "op_sup", now)
		_, _ = svc.RegisterProcess(202, ProcessTypeWorkerRunner, projectID, "op_sup", now)

		// MCP client disconnects
		svc.HandleClientDisconnect(200, projectID, now)
		if svc.IsProcessAlive(200) {
			t.Fatalf("expected client to be disconnected")
		}
		// Supervisor and background runner remain alive!
		if !svc.IsProcessAlive(201) {
			t.Fatalf("supervisor must survive client disconnect (Test X26)")
		}
		if !svc.IsProcessAlive(202) {
			t.Fatalf("supervised worker must survive client disconnect (Test X26)")
		}

		// 3. Human persistent pause stops new claims/actions
		svc.PauseExecution(projectID, "Operator inspecting logs")
		if !svc.IsPaused(projectID) {
			t.Fatalf("expected project to be paused")
		}
		// Registering a new process while paused is rejected
		_, err = svc.RegisterProcess(300, ProcessTypeWorkerRunner, projectID, "op_blocked", now)
		if err == nil || !strings.Contains(err.Error(), ErrHumanPauseActive) {
			t.Fatalf("expected ErrHumanPauseActive while project is paused, got: %v", err)
		}

		// Continue clears pause
		svc.ContinueExecution(projectID)
		if svc.IsPaused(projectID) {
			t.Fatalf("expected project to be unpaused after continue")
		}
		_, err = svc.RegisterProcess(300, ProcessTypeWorkerRunner, projectID, "op_resumed", now)
		if err != nil {
			t.Fatalf("expected successful registration after continue, got: %v", err)
		}
	})

	// =========================================================================
	// Test X27: Execution Truth
	// =========================================================================
	t.Run("X27_ExecutionTruth", func(t *testing.T) {
		svc := NewSupervisorService()
		projectID := "proj_x27"

		// Advance ledger to sequence 10
		for i := 0; i < 9; i++ {
			_, _ = svc.AdvanceLedgerSequence()
		}
		currentN := svc.GetCurrentSequence()
		if currentN != 10 {
			t.Fatalf("expected current sequence 10, got %d", currentN)
		}

		// Read old through_sequence 5 while rows have advanced to 10
		page, err := svc.ReadExecutionPage(projectID, 0, 5, "client_watcher_1")
		if err != nil {
			t.Fatalf("ReadExecutionPage failed: %v", err)
		}

		// Invariant: current state sequence N (10) is distinct from event bound M (5)
		if page.ThroughSequence != 5 {
			t.Fatalf("expected event window through_sequence 5, got %d", page.ThroughSequence)
		}
		if page.StateSequence != 10 {
			t.Fatalf("expected current state_sequence 10, got %d", page.StateSequence)
		}
		if page.StateSemantics != "CURRENT" {
			t.Fatalf("expected state_semantics CURRENT, got %s", page.StateSemantics)
		}

		// Validate page conforms to specs/execution/page.schema.json v2
		rawPage, err := json.Marshal(page)
		if err != nil {
			t.Fatalf("failed to marshal page: %v", err)
		}
		if err := ValidateExecutionPage(rawPage); err != nil {
			t.Fatalf("execution page failed schema validation: %v", err)
		}

		// Event store failure stops admission
		svc.SetStorageLedgerFailing(true)
		_, err = svc.AdvanceLedgerSequence()
		if err == nil || !strings.Contains(err.Error(), ErrStorageLedgerFailure) {
			t.Fatalf("expected ErrStorageLedgerFailure when event store fails, got: %v", err)
		}
	})

	// =========================================================================
	// Test X32: Security Trust Boundary
	// =========================================================================
	t.Run("X32_TrustBoundary", func(t *testing.T) {
		svc := NewSupervisorService()
		path := "app/Config/app.php"
		trustedHash := strings.Repeat("c", 64)
		svc.RegisterTrustedFile(path, trustedHash)

		// Unmodified file passes mediated check
		ok, err := svc.VerifyFileAccessMediated(path, trustedHash)
		if err != nil || !ok {
			t.Fatalf("expected mediated verification to pass for unmodified file")
		}

		// Hostile file modification by local account with same filesystem privileges
		hostileModifiedHash := strings.Repeat("d", 64)
		ok, err = svc.VerifyFileAccessMediated(path, hostileModifiedHash)
		if ok || err == nil || !strings.Contains(err.Error(), ErrHostileAccessDetected) {
			t.Fatalf("expected ErrHostileAccessDetected when mediated service revalidates modified file, got: %v", err)
		}
	})
}
