package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestM4_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M4:
// Migration Reconstruction Engine, integrating Units 1 through 7 and verifying all
// authoritative M4 falsification tests: X19, X20, X21, X22, X23, X67, and scoped acceptance.
func TestM4_AcceptanceClosure(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)
	dummyRef2 := "sha256:" + strings.Repeat("b", 64)

	// =========================================================================
	// 1. Test X22 & X23: Resource Ownership, Fencing & Serialized Publication (Unit 1)
	// =========================================================================
	t.Run("X22_X23_ResourceOwnershipAndFencing", func(t *testing.T) {
		svc := NewResourceFencingService()
		now := time.Now()

		// X22: Duplicate active work rejection
		m1 := &MissionRecord{
			MissionID:    "m_auth_01",
			ProjectID:    "proj_x22",
			Type:         "MIGRATE_UNIT",
			Status:       MissionStatusQueued,
			Fingerprint:  "fp_auth_migration",
			Description:  "Migrate Auth",
		}
		_, err := svc.SubmitMission(m1)
		if err != nil {
			t.Fatalf("first mission submission failed: %v", err)
		}

		// Duplicate active work rejection
		m2 := &MissionRecord{
			MissionID:    "m_auth_02",
			ProjectID:    "proj_x22",
			Type:         "MIGRATE_UNIT",
			Status:       MissionStatusQueued,
			Fingerprint:  "fp_auth_migration",
			Description:  "Duplicate submission",
		}
		_, err = svc.SubmitMission(m2)
		if err == nil || !strings.Contains(err.Error(), ErrDuplicateActiveWork) {
			t.Fatalf("expected ErrDuplicateActiveWork, got: %v", err)
		}

		// Claim mission 1 to get fencing token 1
		claimed, err := svc.ClaimMission("proj_x22", "m_auth_01", "worker_1", now, 5*time.Minute)
		if err != nil {
			t.Fatalf("claim mission failed: %v", err)
		}

		// X22: Shared file conflict rejection
		sharedFile := "app/Services/AuthService.php"
		_, err = svc.AcquireResourceClaim("proj_x22", sharedFile, "m_auth_01", "worker_1", claimed.FencingToken, now, 5*time.Minute)
		if err != nil {
			t.Fatalf("first file claim failed: %v", err)
		}
		// Attempting claim by another worker with lower/equal token fails
		_, err = svc.AcquireResourceClaim("proj_x22", sharedFile, "m_auth_02", "worker_2", 1, now, 5*time.Minute)
		if err == nil || !strings.Contains(err.Error(), ErrResourceLocked) {
			t.Fatalf("expected ErrResourceLocked on shared file claim, got: %v", err)
		}

		// X23: Fenced publication with stale token rejection
		cleanHash := strings.Repeat("a", 64)
		patch := &PatchRecord{
			SchemaVersion: "1.0",
			ProjectID:     "proj_x22",
			UnitID:        "unit_auth",
			FencingToken:  0, // stale fence token (current claimed is 1)
			Files: []PatchFileRecord{
				{
					Path:         sharedFile,
					Action:       PatchActionModify,
					BeforeSHA256: &cleanHash,
					AfterSHA256:  &cleanHash,
				},
			},
		}
		targetFiles := map[string][]byte{sharedFile: []byte("original content")}
		err = svc.ValidatePublication(patch, targetFiles, now)
		if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) {
			t.Fatalf("expected ErrStaleFenceToken on stale token publish, got: %v", err)
		}

		// X23: Before-hash mismatch rejection
		patch.FencingToken = 1
		err = svc.ValidatePublication(patch, targetFiles, now)
		if err == nil || !strings.Contains(err.Error(), ErrContentHashMismatch) {
			t.Fatalf("expected ErrContentHashMismatch, got: %v", err)
		}
	})

	// =========================================================================
	// 2. Test X21: Lossless Inputs & Lexical Representation Fidelity (Unit 2)
	// =========================================================================
	t.Run("X21_LosslessInputsAndLexicalNumbers", func(t *testing.T) {
		// ProjectRawJSONNumber: lexical tag distinction
		intVal := ProjectRawJSONNumber("1")
		if intVal.Type != ValueTypeInteger || intVal.Value != "1" {
			t.Fatalf("expected integer '1', got %+v", intVal)
		}

		decVal := ProjectRawJSONNumber("1.0")
		if decVal.Type != ValueTypeDecimal || decVal.Value != "1" {
			t.Fatalf("expected decimal '1', got %+v", decVal)
		}

		// Strict type isolation: cross-type comparison is UNKNOWN without explicit coercion
		_, verdict, err := CompareTaggedValues(intVal, decVal)
		if verdict != "UNKNOWN" {
			t.Fatalf("cross-type comparison must be UNKNOWN, got verdict=%s err=%v", verdict, err)
		}

		// Explicit coercion allows comparison
		coerced, err := CoerceIntegerToDecimal(intVal)
		if err != nil {
			t.Fatalf("coercion failed: %v", err)
		}
		eq, verdict, _ := CompareTaggedValues(coerced, decVal)
		if !eq || verdict != "PASS" {
			t.Fatalf("coerced comparison should PASS, got eq=%v verdict=%s", eq, verdict)
		}

		// Lossless HTTP request: query absence vs empty string fidelity
		reqAbsent, err := ParseLosslessHTTPRequest("GET", "/api/items", nil, []byte{})
		if err != nil || reqAbsent.HasQuery {
			t.Fatalf("expected HasQuery=false for absent query, got %v", reqAbsent.HasQuery)
		}

		reqEmpty, err := ParseLosslessHTTPRequest("GET", "/api/items?", nil, []byte{})
		if err != nil || !reqEmpty.HasQuery || len(reqEmpty.Query) != 0 {
			t.Fatalf("expected HasQuery=true and empty pairs for '?', got %+v", reqEmpty)
		}

		// Parameter without '=' vs parameter with empty value 'key='
		reqNoVal, err := ParseLosslessHTTPRequest("GET", "/api/items?flag", nil, []byte{})
		if err != nil || len(reqNoVal.Query) != 1 || reqNoVal.Query[0].Value != nil {
			t.Fatalf("expected nil value for param without '=', got %+v", reqNoVal.Query[0])
		}

		reqEmptyVal, err := ParseLosslessHTTPRequest("GET", "/api/items?flag=", nil, []byte{})
		if err != nil || len(reqEmptyVal.Query) != 1 || reqEmptyVal.Query[0].Value == nil || *reqEmptyVal.Query[0].Value != "" {
			t.Fatalf("expected empty string value for param with '=', got %+v", reqEmptyVal.Query[0])
		}

		// CompareLosslessRequests detects query presence divergence
		eq, verdict, mismatchIDs := CompareLosslessRequests(reqAbsent, reqEmpty)
		if eq || verdict != "MISMATCH" || len(mismatchIDs) == 0 {
			t.Fatalf("expected MISMATCH between absent and empty query, got eq=%v verdict=%s ids=%v", eq, verdict, mismatchIDs)
		}
	})

	// =========================================================================
	// 3. Test X20: Relational Identity Normalization & Mutation Protection (Unit 3)
	// =========================================================================
	t.Run("X20_RelationalIdentityNormalizationProtection", func(t *testing.T) {
		rules := []*NormalizationRule{
			{
				SchemaVersion:           "2.0",
				RuleID:                  "norm_uuid",
				Path:                    "/id",
				ValueType:               "string",
				Variability:             "GENERATED_ID",
				Relation:                "BIJECTIVE_ID",
				RelationGroup:           "user_id",
				LifecycleContractRef:    dummyRef,
				BusinessSignificanceRef: dummyRef,
				DecisionRef:             dummyRef,
				NegativeMutationRefs:    []string{dummyRef2},
			},
		}

		// Policy validation enforces negative mutation evidence
		policyErrs := ValidateNormalizationPolicy(rules)
		if len(policyErrs) != 0 {
			t.Fatalf("unexpected policy error for valid rule: %v", policyErrs)
		}

		// Broken foreign key survives normalization as MISMATCH
		protector := NewNormalizerMutationProtector(rules)
		fkMutant := NormalizationMutant{
			MutantID:    "broken-fk-order-user",
			Category:    "BROKEN_FOREIGN_KEY",
			SourceValue: TaggedValue{Type: ValueTypeInteger, Value: "42"},
			TargetValue: TaggedValue{Type: ValueTypeInteger, Value: "99"},
		}
		fkResult := protector.CheckMutant(fkMutant)
		if fkResult.Verdict != "MISMATCH" || fkResult.WasNormalized {
			t.Fatalf("broken foreign key must survive as MISMATCH without normalization: %+v", fkResult)
		}

		// Wrong session principal survives normalization as MISMATCH
		principalMutant := NormalizationMutant{
			MutantID:    "wrong-session-principal",
			Category:    "WRONG_PRINCIPAL",
			SourceValue: TaggedValue{Type: ValueTypeString, Value: "alice"},
			TargetValue: TaggedValue{Type: ValueTypeString, Value: "bob"},
		}
		pResult := protector.CheckMutant(principalMutant)
		if pResult.Verdict != "MISMATCH" || pResult.WasNormalized {
			t.Fatalf("wrong principal must survive as MISMATCH: %+v", pResult)
		}

		// Bijection checker detects two sources mapping to the same target
		checker := &BijectionIntegrityChecker{}
		connID := "conn-1"
		badMap := &IdentityMapV2{
			SchemaVersion:      "2.1",
			ProjectID:          "proj-1",
			CaseID:             "case-1",
			WorkflowInstanceID: "wf-1",
			PolicyRef:          dummyRef,
			Bindings: []IdentityBinding{
				{
					BindingID:     "b1",
					RelationGroup: "grp-1",
					ActorID:       "actor-1",
					ConnectionID:  &connID,
					LogicalKey:    "user_alice",
					SourceLocalID: "src-alice",
					TargetLocalID: "tgt-merged",
					EvidenceRefs:  []string{dummyRef},
				},
				{
					BindingID:     "b2",
					RelationGroup: "grp-1",
					ActorID:       "actor-1",
					ConnectionID:  &connID,
					LogicalKey:    "user_bob",
					SourceLocalID: "src-bob",
					TargetLocalID: "tgt-merged", // duplicate target mapping!
					EvidenceRefs:  []string{dummyRef},
				},
			},
			RelationshipAssertionRefs: []string{dummyRef},
		}
		bijErrs := checker.CheckBijection(badMap)
		if len(bijErrs) == 0 {
			t.Fatal("expected bijection violation for two sources mapped to same target")
		}
	})

	// =========================================================================
	// 4. Test X19 & X67: Session & CSRF Lifecycle Bridge (Unit 4)
	// =========================================================================
	t.Run("X19_X67_SessionAndCSRFLifecycle", func(t *testing.T) {
		sessionContract := &SessionContinuityContract{
			SchemaVersion:          "1.0",
			ProjectID:              "proj-1",
			CaseID:                 "case-login",
			WorkflowID:             "wf-login",
			DecisionRef:            dummyRef,
			Strategy:               "SAFE_BRIDGE",
			CodecImplementationRef: dummyRef,
			CodecProfileRef:        dummyRef,
			CodecBinding: &LoweringBinding{
				CapabilityID:        "safe-session-codec",
				ProviderManifestRef: dummyRef,
				QualificationRef:    dummyRef,
				ObligationRefs:      []string{dummyRef},
				InputShapeRef:       dummyRef,
				OutputShapeRef:      dummyRef,
			},
			Checks: []SessionLifecycleCheck{
				{
					ID: "chk-active",
					Observation: SessionObservation{
						Side:     SideSource,
						StepID:   "login",
						Visit:    1,
						Phase:    PhasePostOperation,
						ActorRef: dummyRef,
					},
					Predicate: &SemanticsExpression{
						Op:        "literal",
						ValueType: ValueTypeBoolean,
						Value:     true,
					},
				},
			},
			Relations: []SessionLifecycleRelation{
				{
					ID: "rel-token",
					Left: SessionLifecycleValue{
						Observation: SessionObservation{
							Side:     SideSource,
							StepID:   "login",
							Visit:    1,
							Phase:    PhasePostOperation,
							ActorRef: dummyRef,
						},
						Path: "/token",
					},
					Right: SessionLifecycleValue{
						Observation: SessionObservation{
							Side:     SideTarget,
							StepID:   "login",
							Visit:    1,
							Phase:    PhasePostOperation,
							ActorRef: dummyRef,
						},
						Path: "/token",
					},
					Different: false,
				},
			},
		}

		rawContract, err := json.Marshal(sessionContract)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateSessionContinuityContract(rawContract); err != nil {
			t.Fatalf("session contract validation failed: %v", err)
		}

		// X67: Response and session lowering verification in comparator
		comp := NewBehaviorComparator()
		in := buildBaselinePassInput()
		// Dropping cookie HttpOnly attribute causes MISMATCH in Session dimension
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Cookies[0].HTTPOnly = false
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("X67: dropped cookie attribute must cause MISMATCH, got verdict=%s err=%v", out.CaseVerdict, err)
		}
	})

	// =========================================================================
	// 5. Test Unit 5: Scoped Admission & Narrow Target Generation
	// =========================================================================
	t.Run("Unit5_NarrowTargetGeneration", func(t *testing.T) {
		unitID := "sync-auth"
		adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session"}, []string{"login", "health"})
		tm := makeTestTargetModel(t, unitID)
		sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

		gen := NewNarrowTargetGenerator()
		out, err := gen.Generate(GenerationInput{
			Admission:       adm,
			SessionContract: sc,
			TargetModel:     tm,
			UnitID:          unitID,
		})
		if err != nil {
			t.Fatalf("narrow generation failed: %v", err)
		}

		// Manifest validation
		rawMan, err := json.Marshal(out.Manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateGenerationManifest(rawMan); err != nil {
			t.Fatalf("generated manifest invalid: %v", err)
		}

		// Staging check
		staged, err := StageGeneration(out, 1, dummyRef)
		if err != nil || staged.FenceToken != 1 {
			t.Fatalf("stage generation failed: %v", err)
		}

		// Artifact verification: routes, controllers, middleware generated
		expectedArtifacts := []string{
			"bootstrap/app.php",
			"routes/web.php",
			"app/Http/Controllers/ReconstructedController.php",
			"app/Http/Middleware/SessionContinuityMiddleware.php",
			"config/session-continuity.php",
			"generated/provenance.json",
		}
		for _, ea := range expectedArtifacts {
			if _, ok := out.Artifacts[ea]; !ok {
				t.Fatalf("missing expected generated artifact: %s", ea)
			}
		}
	})

	// =========================================================================
	// 6. Test Unit 6: Views & File Reconstruction
	// =========================================================================
	t.Run("Unit6_ViewsAndFileReconstruction", func(t *testing.T) {
		unitID := "unit-views"
		allowedPaths := []string{
			"resources/views/auth/login.blade.php",
			"public/css/app.css",
		}
		adm := makeTestViewAdmission(t, unitID)
		tm := makeTestViewTargetModel(t, unitID, allowedPaths)

		reconstructor := NewViewFileReconstructor()
		out, err := reconstructor.Reconstruct(ViewReconstructionInput{
			Admission:   adm,
			TargetModel: tm,
			UnitID:      unitID,
			Obligations: []ViewTemplateObligation{
				{
					ViewID:       "view-login",
					SourcePath:   "views/login.php",
					TargetPath:   "resources/views/auth/login.blade.php",
					Kind:         ViewKindBladeTemplate,
					EscapingMode: EscapingModeAuto,
					RawTemplate:  "<form><input name='email'>{{ $title }}</form>",
					Variables:    []ViewVariable{{Name: "title", Type: "string", RawApproved: false}},
				},
				{
					ViewID:       "asset-style",
					SourcePath:   "css/app.css",
					TargetPath:   "public/css/app.css",
					Kind:         ViewKindStaticAsset,
					EscapingMode: EscapingModeNone,
					RawTemplate:  "body { margin: 0; }",
				},
			},
			FencingToken: 1,
		})
		if err != nil {
			t.Fatalf("view reconstruction failed: %v", err)
		}

		// Manifest and patch validation
		rawMan, _ := json.Marshal(out.Manifest)
		if err := ValidateViewManifest(rawMan); err != nil {
			t.Fatalf("view manifest invalid: %v", err)
		}
		rawPatch, _ := out.Patch.CanonicalJSON()
		if err := ValidatePatch(rawPatch); err != nil {
			t.Fatalf("view patch invalid: %v", err)
		}

		// Path traversal attempt must be rejected
		_, err = reconstructor.Reconstruct(ViewReconstructionInput{
			Admission:   adm,
			TargetModel: tm,
			UnitID:      unitID,
			Obligations: []ViewTemplateObligation{
				{
					ViewID:       "view-traversal",
					SourcePath:   "evil.php",
					TargetPath:   "resources/views/../../etc/passwd",
					Kind:         ViewKindBladeTemplate,
					EscapingMode: EscapingModeAuto,
					RawTemplate:  "evil",
				},
			},
			FencingToken: 1,
		})
		if err == nil || !strings.Contains(err.Error(), "ERR_TARGET_CONFLICT") {
			t.Fatalf("expected ERR_TARGET_CONFLICT on path traversal, got: %v", err)
		}
	})

	// =========================================================================
	// 7. Test Unit 7: Business/External Effect Reconstruction & Safety
	// =========================================================================
	t.Run("Unit7_BusinessEffectReconstructionAndSafety", func(t *testing.T) {
		unitID := "unit-effects"
		adm := makeTestEffectAdmission(t, unitID)
		tm := makeTestEffectTargetModel(t, unitID)

		reconstructor := NewEffectReconstructor()
		parentID := "eff-1"
		out, err := reconstructor.Reconstruct(EffectReconstructionInput{
			Admission:   adm,
			TargetModel: tm,
			UnitID:      unitID,
			Effects: []BusinessEffectRecord{
				{
					EffectID:              "eff-1",
					LogicalEffectID:       "log-1",
					Dimension:             EffectDimensionDBWrite,
					EffectClass:           EffectClassLocalWrite,
					BusinessKey:           "order-1",
					DestinationID:         "db-main",
					BusinessApplicationID: "app-1",
					SequenceIndex:         1,
					PayloadRef:            dummyRef,
					Applied:               true,
				},
				{
					EffectID:              "eff-2",
					LogicalEffectID:       "log-2",
					Dimension:             EffectDimensionQueue,
					EffectClass:           EffectClassLocalWrite,
					BusinessKey:           "order-1",
					DestinationID:         "queue-orders",
					BusinessApplicationID: "app-1-q",
					SequenceIndex:         2,
					CausalParentID:        &parentID,
					PayloadRef:            dummyRef2,
					Applied:               true,
				},
			},
			SafetyPolicy: EffectSafetyPolicy{IsProductionEnvironment: false, AllowIrreversible: false},
		})
		if err != nil {
			t.Fatalf("effect reconstruction failed: %v", err)
		}

		rawMan, _ := json.Marshal(out.Manifest)
		if err := ValidateEffectPublicationManifest(rawMan); err != nil {
			t.Fatalf("effect manifest invalid: %v", err)
		}

		// Cardinality check: missing business effect in target yields MISMATCH
		srcEffects := out.Manifest.Effects
		tgtMissing := []BusinessEffectRecord{} // empty!
		cardResult := VerifyEffectCardinality(srcEffects, tgtMissing)
		if cardResult.Verdict != "MISMATCH" || len(cardResult.MismatchIDs) == 0 || cardResult.MismatchIDs[0] != "mismatch-missing-effect" {
			t.Fatalf("expected mismatch-missing-effect, got: %+v", cardResult)
		}

		// Causal ordering check: non-monotonic index fails
		invertedEffects := []BusinessEffectRecord{
			{EffectID: "eff-a", SequenceIndex: 2},
			{EffectID: "eff-b", SequenceIndex: 1}, // non-monotonic
		}
		err = VerifyEffectCausalityAndOrdering(invertedEffects)
		if err == nil || !strings.Contains(err.Error(), "ERR_ORDERING_VIOLATION") {
			t.Fatalf("expected ERR_ORDERING_VIOLATION, got: %v", err)
		}
	})

	// =========================================================================
	// 8. Test Scoped Admission vs Project Acceptance (Anti-Deadlock Invariant)
	// =========================================================================
	t.Run("ScopedAdmissionVsProjectAcceptance", func(t *testing.T) {
		plan, checks := admissionTestPlan(t)

		// Slice A (synchronous route + auth) is admitted
		admSync, err := EvaluateGenerationAdmission(plan, "sync-auth", checks)
		if err != nil || admSync.Outcome != "ADMITTED" {
			t.Fatalf("sync-auth should be ADMITTED, got outcome=%s err=%v", admSync.Outcome, err)
		}

		// Slice B (async completion with unknown evidence) is blocked
		admAsync, err := EvaluateGenerationAdmission(plan, "async", checks)
		if err != nil || admAsync.Outcome != "BLOCKED" {
			t.Fatalf("async slice must be BLOCKED, got outcome=%s err=%v", admAsync.Outcome, err)
		}

		// Project acceptance: whole project CANNOT complete when async is unknown
		graph, err := NewPlanGraphBuilder(plan.ProjectID, plan.Subject.CorpusRef, plan.Subject.TargetModelRef, nil).
			AddUnit(SchedulableBehavioralUnit{UnitID: "sync-auth", ContractRefs: []string{dummyRef}, CaseIDs: []string{"route", "session"}}).
			AddUnit(SchedulableBehavioralUnit{UnitID: "async", ContractRefs: []string{dummyRef}, CaseIDs: []string{"completion"}}).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		projectScope, err := graph.DeriveProjectScope()
		if err != nil {
			t.Fatal(err)
		}
		if len(projectScope.CaseIDs) != 3 {
			t.Fatalf("project scope lost cases: %v", projectScope.CaseIDs)
		}

		// CanProjectComplete must return false because async is UNKNOWN
		canComplete, blockers := graph.CanProjectComplete(
			[]string{"sync-auth", "async"},
			map[string]string{"sync-auth": "VERIFIED", "async": "UNKNOWN"},
		)
		if canComplete {
			t.Fatal("project must NOT complete while async unit is UNKNOWN")
		}
		if len(blockers) == 0 {
			t.Fatal("expected blockers explaining incomplete project acceptance")
		}
	})
}
