package domain

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestM9_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M9:
// Additional Target Adapters, verifying:
// 1. Target adapter contract & registry binding (Universal model boundaries, exact resolution)
// 2. Multi-target generation determinism (Go standard library and Reference Profile #1 Laravel)
// 3. Rejection of unsupported target capabilities and unresolvable targets
// 4. Request lifecycle, routing, and session continuity lowering across targets
// 5. Schema-aware database query and transaction lowering with structured error handling
// 6. Cross-milestone regression protection ensuring M0–M8 contracts remain intact
func TestM9_AcceptanceClosure(t *testing.T) {
	ctx := context.Background()

	// =========================================================================
	// 1. Target Adapter Registry & Exact Resolution
	// =========================================================================
	t.Run("TargetAdapterRegistry_ExactResolution", func(t *testing.T) {
		reg := NewTargetAdapterRegistry()
		goAdapter := NewGoStdlibTargetAdapter()
		if err := reg.Register(goAdapter); err != nil {
			t.Fatalf("register goAdapter failed: %v", err)
		}

		// Exact resolution for Go standard library (framework=null)
		resolvedDocGo := &ResolvedEnvironmentDocument{
			Language:  "go",
			Framework: nil,
		}
		adapter, err := reg.ResolveForResolution(resolvedDocGo)
		if err != nil {
			t.Fatalf("expected successful resolution for Go stdlib, got: %v", err)
		}
		if adapter.Manifest().Language != "go" || adapter.Manifest().Framework != nil {
			t.Fatalf("unexpected adapter resolved: %+v", adapter.Manifest())
		}

		// Unregistered target returns typed error
		unregisteredDoc := &ResolvedEnvironmentDocument{
			Language:  "rust",
			Framework: nil,
		}
		_, err = reg.ResolveForResolution(unregisteredDoc)
		if err == nil || !strings.Contains(err.Error(), "no qualified target adapter registered") {
			t.Fatalf("expected no target adapter registered error, got: %v", err)
		}
	})

	// =========================================================================
	// 2. Capability Validation & Rejection of Unsupported Behaviors
	// =========================================================================
	t.Run("TargetAdapter_CapabilityValidation", func(t *testing.T) {
		adapter := NewGoStdlibTargetAdapter()
		unitID := "unit-caps"
		tm := makeTestTargetModel(t, unitID)

		// Supported capabilities
		res, err := adapter.ValidateCapabilities(ctx, tm, []string{"routing", "database", "transactions", "json_api"})
		if err != nil || !res.Valid || len(res.UnsupportedCaps) != 0 {
			t.Fatalf("expected valid capabilities, got res=%+v err=%v", res, err)
		}

		// Unsupported capabilities must be reported as typed blockers
		resBad, err := adapter.ValidateCapabilities(ctx, tm, []string{"routing", "blade_templates", "active_record_reflection"})
		if err != nil {
			t.Fatalf("unexpected error during capability validation: %v", err)
		}
		if resBad.Valid {
			t.Fatalf("expected Valid=false for unsupported capabilities, got true")
		}
		if len(resBad.UnsupportedCaps) != 2 {
			t.Fatalf("expected 2 unsupported capabilities, got: %v", resBad.UnsupportedCaps)
		}
	})

	// =========================================================================
	// 3. Multi-Target Generation Determinism & Artifact Reproducibility
	// =========================================================================
	t.Run("MultiTargetGeneration_DeterminismAndParity", func(t *testing.T) {
		unitID := "sync-checkout"
		adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session", "auth", "database", "transactions"}, []string{"login", "checkout-pay"})
		tm := makeTestTargetModel(t, unitID)
		sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

		goAdapter := NewGoStdlibTargetAdapter()
		input := GenerationInput{
			Admission:       adm,
			TargetModel:     tm,
			UnitID:          unitID,
			SessionContract: sc,
		}

		// Pass 1: Go stdlib generation
		out1, err := goAdapter.Generate(ctx, input)
		if err != nil {
			t.Fatalf("Go stdlib generation pass 1 failed: %v", err)
		}

		// Pass 2: Go stdlib generation
		out2, err := goAdapter.Generate(ctx, input)
		if err != nil {
			t.Fatalf("Go stdlib generation pass 2 failed: %v", err)
		}

		// Determinism check: identical manifest digest
		if out1.Manifest.ProvenanceRef != out2.Manifest.ProvenanceRef {
			t.Fatalf("non-deterministic Go target manifest provenance: %s != %s", out1.Manifest.ProvenanceRef, out2.Manifest.ProvenanceRef)
		}

		// Manifest validation
		rawMan, err := json.Marshal(out1.Manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateGenerationManifest(rawMan); err != nil {
			t.Fatalf("generated Go manifest invalid: %v", err)
		}

		// Verify Laravel narrow target generator determinism for the same input
		laravelGen := NewNarrowTargetGenerator()
		lOut1, err := laravelGen.Generate(input)
		if err != nil {
			t.Fatalf("Laravel target generation pass 1 failed: %v", err)
		}
		lOut2, err := laravelGen.Generate(input)
		if err != nil {
			t.Fatalf("Laravel target generation pass 2 failed: %v", err)
		}
		if lOut1.Manifest.ProvenanceRef != lOut2.Manifest.ProvenanceRef {
			t.Fatalf("non-deterministic Laravel manifest provenance: %s != %s", lOut1.Manifest.ProvenanceRef, lOut2.Manifest.ProvenanceRef)
		}

		// Cross-target closure separation: distinct generators produce distinct artifact sets
		if len(out1.Artifacts) == 0 || len(lOut1.Artifacts) == 0 {
			t.Fatalf("empty artifacts generated")
		}
		if out1.Manifest.ProvenanceRef == lOut1.Manifest.ProvenanceRef {
			t.Fatalf("expected different provenance between Go and Laravel target artifacts")
		}
	})

	// =========================================================================
	// 4. Request Lifecycle, Routing, and Session Continuity Lowering
	// =========================================================================
	t.Run("RoutingAndSessionContinuityLowering", func(t *testing.T) {
		unitID := "auth-flow"
		adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session", "auth"}, []string{"login", "verify-2fa"})
		tm := makeTestTargetModel(t, unitID)
		sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

		goAdapter := NewGoStdlibTargetAdapter()
		out, err := goAdapter.Generate(ctx, GenerationInput{
			Admission:       adm,
			TargetModel:     tm,
			UnitID:          unitID,
			SessionContract: sc,
		})
		if err != nil {
			t.Fatalf("generate failed: %v", err)
		}

		// Routing table check
		routerBytes, ok := out.Artifacts["internal/server/router.go"]
		if !ok {
			t.Fatalf("missing internal/server/router.go")
		}
		rStr := string(routerBytes)
		if !strings.Contains(rStr, `mux.HandleFunc("/login", h.LoginHandler)`) {
			t.Errorf("missing login route: %s", rStr)
		}
		if !strings.Contains(rStr, `mux.HandleFunc("/verify-2fa", h.Verify2faHandler)`) {
			t.Errorf("missing verify-2fa route: %s", rStr)
		}
		if !strings.Contains(rStr, "middleware.SessionContinuity(mux)") {
			t.Errorf("missing session continuity middleware wrapper: %s", rStr)
		}

		// Session middleware check
		mwBytes, ok := out.Artifacts["internal/middleware/session.go"]
		if !ok {
			t.Fatalf("missing internal/middleware/session.go")
		}
		if !strings.Contains(string(mwBytes), "Session continuity strategy: SAFE_BRIDGE") {
			t.Errorf("session middleware missing SAFE_BRIDGE strategy: %s", string(mwBytes))
		}
	})

	// =========================================================================
	// 5. Database Access, Transaction Semantics & Error Handling Lowering
	// =========================================================================
	t.Run("DatabaseAndTransactionLowering", func(t *testing.T) {
		unitID := "repo-mutation"
		adm := makeTestAdmission(t, unitID, []string{"http", "routing", "database", "transactions"}, []string{"update-profile"})
		tm := makeTestTargetModel(t, unitID)

		goAdapter := NewGoStdlibTargetAdapter()
		out, err := goAdapter.Generate(ctx, GenerationInput{
			Admission:   adm,
			TargetModel: tm,
			UnitID:      unitID,
		})
		if err != nil {
			t.Fatalf("generate failed: %v", err)
		}

		// DB repository check
		dbBytes, ok := out.Artifacts["internal/repository/db.go"]
		if !ok {
			t.Fatalf("missing internal/repository/db.go")
		}
		dbStr := string(dbBytes)
		if !strings.Contains(dbStr, "WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error") {
			t.Errorf("missing transaction boundary method: %s", dbStr)
		}
		if !strings.Contains(dbStr, "MutateUpdateProfile(ctx context.Context, query string, args ...any) (int64, error)") {
			t.Errorf("missing MutateUpdateProfile: %s", dbStr)
		}
		if !strings.Contains(dbStr, "res.RowsAffected()") {
			t.Errorf("missing rows-affected semantics: %s", dbStr)
		}

		// Error mapping check
		errBytes, ok := out.Artifacts["internal/repository/errors.go"]
		if !ok {
			t.Fatalf("missing internal/repository/errors.go")
		}
		errStr := string(errBytes)
		if !strings.Contains(errStr, "ErrRecordNotFound") || !strings.Contains(errStr, "MapDBError(err error) error") {
			t.Errorf("missing structured error mapping: %s", errStr)
		}
	})

	// =========================================================================
	// 6. Cross-Milestone Regression Guardrails (M0–M8 Invariant Preservation)
	// =========================================================================
	t.Run("CrossMilestoneRegressionProtection", func(t *testing.T) {
		// Fencing Token Invariant (M4 Unit 1): Stale token rejection
		svc := NewResourceFencingService()
		now := time.Now()
		m := &MissionRecord{
			MissionID:   "m_m9_01",
			ProjectID:   "proj_m9",
			Type:        "MIGRATE_UNIT",
			Status:      MissionStatusQueued,
			Fingerprint: "fp_m9",
			Description: "M9 test",
		}
		_, err := svc.SubmitMission(m)
		if err != nil {
			t.Fatalf("submit mission failed: %v", err)
		}
		claimed, err := svc.ClaimMission("proj_m9", "m_m9_01", "worker_1", now, 5*time.Minute)
		if err != nil {
			t.Fatalf("claim mission failed: %v", err)
		}
		_, err = svc.AcquireResourceClaim("proj_m9", "app/model.go", "m_m9_01", "worker_1", claimed.FencingToken, now, 5*time.Minute)
		if err != nil {
			t.Fatalf("acquire claim failed: %v", err)
		}

		cleanHash := strings.Repeat("c", 64)
		staleTokenPatch := &PatchRecord{
			SchemaVersion: "1.0",
			ProjectID:     "proj_m9",
			UnitID:        "unit_m9",
			FencingToken:  claimed.FencingToken - 1, // stale token
			Files: []PatchFileRecord{
				{
					Path:         "app/model.go",
					Action:       PatchActionModify,
					BeforeSHA256: &cleanHash,
					AfterSHA256:  &cleanHash,
				},
			},
		}
		err = svc.ValidatePublication(staleTokenPatch, map[string][]byte{"app/model.go": []byte("initial")}, now)
		if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) {
			t.Fatalf("stale token must be rejected: %v", err)
		}

		// Profiler Anomaly Invariant (M8): Profiler retains query anomaly detection
		profiler := NewExecutionProfiler()
		rc := profiler.StartRequest(ctx, "/api/data", "GET")
		for i := 0; i < 5; i++ {
			rc.RecordQuery("SELECT * FROM items WHERE user_id = 1", 1*time.Millisecond, 1, nil)
		}
		p := profiler.FinishRequest(ctx, rc, 200, 512, nil)
		anomalies := p.DetectQueryAnomalies(3)
		if len(anomalies) != 1 || anomalies[0].PatternType != PatternNPlusOneQuery {
			t.Fatalf("expected N+1 anomaly preserved from M8, got: %v", anomalies)
		}
	})
}
