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

// TestM7_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M7:
// Evidence-quality learning, cross-project privacy boundary, multi-context synthesis progression,
// independent promotion criteria, content-addressed skill lifecycle, quality accounting (X38),
// consumer graph tracking, epoch retraction, and offline project reopen reconciliation (X79).
//
// Authoritative tests integrated:
// - X39: Retraction, privacy boundary, cross-project data isolation, sanitization of secrets/PII/paths.
// - X37: Contaminated learning rejection: 3 correlated fixtures with shared wrong oracle fail promotion.
// - X38: Evidence quality accounting: tool/skill dropping hard traces rejected despite 100% apparent success.
// - X79: Offline consumer retraction: detached projects reconcile latest retraction epoch on reopen.
func TestM7_AcceptanceClosure(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup SQLite database for M7
	dbPath := filepath.Join(tempDir, "m7_closure.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	defer db.Close()

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to obtain conn: %v", err)
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
		t.Fatalf("failed to create migration runner: %v", err)
	}
	if err := runner.Run(ctx, conn); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Insert test projects
	_, err = db.ExecContext(ctx, "INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES ('proj-alpha', 'Project Alpha', '/src/alpha', 'INITIALIZED', '2026-10-10T00:00:00Z', '2026-10-10T00:00:00Z')")
	if err != nil {
		t.Fatalf("failed to insert proj-alpha: %v", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES ('proj-beta', 'Project Beta', '/src/beta', 'INITIALIZED', '2026-10-10T00:00:00Z', '2026-10-10T00:00:00Z')")
	if err != nil {
		t.Fatalf("failed to insert proj-beta: %v", err)
	}

	// =========================================================================
	// 1. Unit 1: Project-Local Learning, Sanitization & Privacy Boundary (X39)
	// =========================================================================
	t.Run("Unit1_ProjectLocalLearningAndPrivacyBoundary", func(t *testing.T) {
		sanitizer := NewSanitizer([]byte("closure-hmac-key-01"))
		svc := NewLearningService(db, sanitizer)

		expCtx := &ExperimentContext{
			SourceConfiguration: "ci3-legacy-profile",
			PlanHash:            "plan-hash-alpha",
			CorpusHash:          "corpus-hash-alpha",
			PolicyHash:          "policy-hash-alpha",
			CaseHash:            "case-hash-01",
			WorkflowHash:        "wf-hash-01",
			FixtureHash:         "fix-hash-01",
			TargetBeforeTreeHash: "before-hash-01",
			AdapterVersions:     map[string]string{"ci3": "3.1.13", "php": "8.3"},
			ToolVersions:        map[string]string{"phpunit": "10.0"},
			DependencyVersions:  map[string]string{"composer": "2.7"},
			HypothesisPatchHash: "patch-hash-01",
			SandboxIdentity:     "sb-alpha-01",
		}

		// Experiment fingerprinting reproducibility
		fp1, _, err := expCtx.Fingerprint()
		if err != nil {
			t.Fatalf("failed to compute fingerprint: %v", err)
		}
		fpRec, err := svc.RecordExperiment(ctx, "proj-alpha", "exp-alpha-1", expCtx, "FAILED")
		if err != nil {
			t.Fatalf("failed to record experiment: %v", err)
		}
		if fp1 != fpRec {
			t.Fatalf("fingerprint mismatch: %s != %s", fp1, fpRec)
		}

		// Negative lesson with embedded credentials, host path, and PII
		rawFinding := "Failed login simulation with password='SuperSecretPassword999' at /home/user/workspace/private/auth.php for user admin@corporate.com"
		rawRemediation := "Strip dynamic hook and bind API key api_key='sk_live_verysecrettoken123'"

		lesson, err := svc.RecordLesson(ctx, "proj-alpha", "les-alpha-1", "exp-alpha-1", rawFinding, rawRemediation, nil)
		if err != nil {
			t.Fatalf("failed to record lesson: %v", err)
		}

		// Verify sanitization occurred
		if strings.Contains(lesson.Finding, "SuperSecretPassword999") || strings.Contains(lesson.Finding, "admin@corporate.com") || strings.Contains(lesson.Finding, "/home/user/workspace/private/auth.php") {
			t.Fatalf("sensitive data leaked into lesson finding: %s", lesson.Finding)
		}
		if strings.Contains(lesson.Remediation, "sk_live_verysecrettoken123") {
			t.Fatalf("sensitive api key leaked into lesson remediation: %s", lesson.Remediation)
		}
		if !strings.Contains(lesson.Finding, "[REDACTED:hmac_") || !strings.Contains(lesson.Finding, "[REDACTED_PII]") {
			t.Fatalf("expected HMAC and PII tokens in sanitized finding: %s", lesson.Finding)
		}

		// Verify cross-project isolation: proj-beta cannot query or record against proj-alpha
		_, err = svc.RecordLesson(ctx, "proj-beta", "les-beta-x", "exp-alpha-1", "Finding", "Remediation", nil)
		if err == nil {
			t.Fatalf("expected cross-project violation recording lesson across projects, got nil")
		}
		_, err = svc.SearchProjectLessons(ctx, "proj-beta", "proj-alpha", "login")
		if err == nil {
			t.Fatalf("expected cross-project lesson search to be refused, got nil")
		}
	})

	// =========================================================================
	// 2. Unit 2: Multi-Context Evidence Synthesis & Independent Promotion (X37)
	// =========================================================================
	t.Run("Unit2_MultiContextSynthesisAndX37PromotionEvaluation", func(t *testing.T) {
		// Valid promotion candidate
		cand := &GlobalPromotionCandidate{
			CandidateID:   "cand-session-bridge",
			SkillID:       "skill-session-bridge",
			Version:       "1.0.0",
			PackageDigest: "9999999999999999999999999999999999999999999999999999999999999999",
			SynthesisStage: StageSkillCandidate,
			Applicability: map[string]string{"framework": "ci3", "target": "laravel"},
			Fixtures: []FixtureContext{
				{FixtureID: "f1", ApplicationID: "app1", Author: "auth1", OracleDerivation: "diff-oracle-1", CorrelationGroup: "grp1", Passed: true},
				{FixtureID: "f2", ApplicationID: "app2", Author: "auth2", OracleDerivation: "diff-oracle-2", CorrelationGroup: "grp2", Passed: true},
				{FixtureID: "f3", ApplicationID: "app3", Author: "auth3", OracleDerivation: "diff-oracle-3", CorrelationGroup: "grp3", Passed: true},
			},
			NegativeMutants: []NegativeMutantEvaluation{
				{MutantID: "m1", MutatedDimension: "owner", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "m2", MutatedDimension: "write_effect", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "m3", MutatedDimension: "token", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "m4", MutatedDimension: "reset", DetectedByOracle: true, Verdict: "MISMATCH"},
				{MutantID: "m5", MutatedDimension: "capture", DetectedByOracle: true, Verdict: "MISMATCH"},
			},
			Holdout: HoldoutEvaluation{
				HoldoutSuiteID:    "holdout-independent-alpha",
				TotalCases:        10,
				PassedCases:       10,
				RegressionDetected: false,
			},
			TotalAttempts:      8,
			FailedAttempts:     1,
			HumanReviewedBy:    "operator-lead",
			HumanApprovalToken: "tok-gate-human-approval-777",
		}

		verdict, err := EvaluatePromotion(cand)
		if err != nil {
			t.Fatalf("unexpected promotion evaluation error: %v", err)
		}
		if !verdict.Eligible || verdict.Decision != "PROMOTED" || verdict.ReceiptHash == "" {
			t.Fatalf("expected promotion success, got decision: %s, reasons: %v", verdict.Decision, verdict.RefusalReasons)
		}

		// X37 Falsification: 3 correlated fixtures with shared wrong oracle MUST fail promotion
		correlatedCand := *cand
		correlatedCand.Fixtures = []FixtureContext{
			{FixtureID: "cf1", ApplicationID: "app1", Author: "auth1", OracleDerivation: "shared-wrong-oracle", CorrelationGroup: "shared-cluster", Passed: true},
			{FixtureID: "cf2", ApplicationID: "app1", Author: "auth1", OracleDerivation: "shared-wrong-oracle", CorrelationGroup: "shared-cluster", Passed: true},
			{FixtureID: "cf3", ApplicationID: "app1", Author: "auth1", OracleDerivation: "shared-wrong-oracle", CorrelationGroup: "shared-cluster", Passed: true},
		}

		corrVerdict, err := EvaluatePromotion(&correlatedCand)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if corrVerdict.Eligible || corrVerdict.Decision != "REJECTED" {
			t.Fatalf("expected X37 correlated fixtures with shared wrong oracle to be REJECTED, got: %s", corrVerdict.Decision)
		}
	})

	// =========================================================================
	// 3. Unit 3: Content-Addressed Skill Lifecycle, X38 Quality & Routing
	// =========================================================================
	t.Run("Unit3_SkillLifecycleX38QualityAndRouting", func(t *testing.T) {
		regSvc := NewSkillRegistryService(nil)

		inv := &SkillPackageInventory{
			Identity: "rwt-skill-package-1",
			Entries: []SkillPackageEntry{
				{Path: "SKILL.md", SizeBytes: 250, SHA256: "1111222233334444555566667777888899990000111122223333444455556666"},
				{Path: "references/templates.md", SizeBytes: 500, SHA256: "2222333344445555666677778888999900001111222233334444555566667777"},
			},
		}

		// Register PROPOSED
		cap, err := regSvc.RegisterProposedSkill(ctx, "skill-auth-bridge", "Auth Bridge Skill", "Lowers auth mechanisms", "1.0.0", inv, []string{"auth", "session"}, nil)
		if err != nil {
			t.Fatalf("failed to register skill: %v", err)
		}
		if cap.Status != SkillStatusProposed {
			t.Fatalf("expected PROPOSED status, got %s", cap.Status)
		}

		// Standard routing excludes unapproved PROPOSED skills
		if len(regSvc.SearchCapabilities("auth", false)) != 0 {
			t.Fatalf("unapproved skill was returned during standard execution routing")
		}

		// Advance PROPOSED -> EXPERIMENTAL
		if err := regSvc.TransitionStatus(ctx, "skill-auth-bridge", "1.0.0", SkillStatusExperimental, ""); err != nil {
			t.Fatalf("failed to transition to EXPERIMENTAL: %v", err)
		}

		// Agent cannot promote to STABLE without human token
		if err := regSvc.TransitionStatus(ctx, "skill-auth-bridge", "1.0.0", SkillStatusStable, ""); err == nil {
			t.Fatalf("expected error promoting to STABLE without human approval token, got nil")
		}

		// Human promotion to STABLE succeeds
		if err := regSvc.TransitionStatus(ctx, "skill-auth-bridge", "1.0.0", SkillStatusStable, "tok-human-operator-gate-33"); err != nil {
			t.Fatalf("failed human promotion to STABLE: %v", err)
		}

		// Now routed during standard execution
		if len(regSvc.SearchCapabilities("auth", false)) != 1 {
			t.Fatalf("expected 1 routed capability for STABLE skill")
		}

		// X38 Falsification: Degraded tool dropping capture/traces rejected despite 100% apparent success
		degradedTool := &ToolQualityMetrics{
			ToolName:              "pdo-trace-filter",
			CaptureCoverage:       0.80,
			TraceDroppedCount:     12,
			UnknownRate:           0.0,
			MutationDetectionRate: 1.0,
			ApparentSuccessRate:   1.0,
		}
		if err := ValidateToolEvidenceQuality(degradedTool); err == nil {
			t.Fatalf("expected X38 degraded tool dropping traces to be rejected")
		}
	})

	// =========================================================================
	// 4. Unit 4: Consumer Graph Tracking, Retraction & Reopen Reconciliation (X79, X39)
	// =========================================================================
	t.Run("Unit4_ConsumerGraphRetractionAndOfflineReopenX79", func(t *testing.T) {
		tracker := NewConsumerGraphTracker()

		revA := "pattern-auth-migration:v1"
		histBytes := "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

		// Record consumption in online project
		_, err := tracker.RecordConsumption("proj-alpha", ConsumerArtifact, "art-alpha-output", revA, histBytes)
		if err != nil {
			t.Fatalf("failed to record consumption: %v", err)
		}

		// Record consumption in detached/offline project
		revB := "pattern-cookie-cipher:v1"
		_, err = tracker.RecordConsumption("proj-beta", ConsumerResult, "res-beta-output", revB, histBytes)
		if err != nil {
			t.Fatalf("failed to record consumption for proj-beta: %v", err)
		}

		// Detach proj-beta
		tracker.SetProjectDetached("proj-beta", true)

		// Retract revB globally while proj-beta is detached
		receipt, err := tracker.RetractRevision(revB, "Cryptographic flaw in legacy cookie cipher")
		if err != nil {
			t.Fatalf("failed to retract revision: %v", err)
		}
		if receipt.ReceiptHash == "" {
			t.Fatalf("expected receipt hash for retraction")
		}

		// Offline reopen: proj-beta cannot claim current global approval
		_, isBlocked, err := tracker.ReconcileOnReopen(ctx, "proj-beta", true)
		if err != nil {
			t.Fatalf("unexpected offline reopen error: %v", err)
		}
		if !isBlocked {
			t.Fatalf("expected offline project to be blocked from current global approval")
		}

		// Online reconnect: reconciles against latest retraction epoch, discovers staleness
		newlyStale, hasStale, err := tracker.ReconcileOnReopen(ctx, "proj-beta", false)
		if err != nil {
			t.Fatalf("unexpected online reconciliation error: %v", err)
		}
		if !hasStale || len(newlyStale) != 1 {
			t.Fatalf("expected 1 stale consumer detected upon online reconnect, got %d", len(newlyStale))
		}
		if !strings.Contains(newlyStale[0].StaleReason, "RETRACTED_DURING_DETACHED_STATE") {
			t.Fatalf("expected RETRACTED_DURING_DETACHED_STATE in reason, got: %s", newlyStale[0].StaleReason)
		}

		// Invariant: historical hash is retained for replay
		if newlyStale[0].HistoricalHash != histBytes {
			t.Fatalf("historical hash was altered or lost upon retraction")
		}
	})
}
