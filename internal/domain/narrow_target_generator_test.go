package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func makeTestAdmission(t *testing.T, unitID string, caps []string, caseIDs []string) *GenerationAdmission {
	t.Helper()
	plan, checks := admissionTestPlan(t)
	adm, err := EvaluateGenerationAdmission(plan, "sync-auth", checks)
	if err != nil {
		t.Fatal(err)
	}
	// Customize unitID, caps and caseIDs for specific tests while preserving valid hashes
	adm.Scope.UnitIDs = []string{unitID}
	adm.Scope.CaseIDs = caseIDs
	adm.Capabilities = caps
	return adm
}

func makeTestTargetModel(t *testing.T, unitID string) *TargetModelDocument {
	t.Helper()
	builder := NewTargetModelBuilder("model-test", "proj-test", validTestRef("a"), validTestRef("b"))
	builder.WithRevision(1)
	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-" + unitID,
		OwnerUnitID:          unitID,
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"bootstrap/app.php", "routes/web.php"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "laravel_skeleton",
		RequiredCapabilities: []string{"http", "routing"},
		ConstraintRefs:       []string{validTestRef("d")},
	})
	doc, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func makeTestSessionContract(t *testing.T, projectID, caseID, strategy string) *SessionContinuityContract {
	t.Helper()
	ref := validTestRef("e")
	qref := validTestRef("f")
	c := &SessionContinuityContract{
		SchemaVersion:          "1.0",
		ProjectID:              projectID,
		CaseID:                 caseID,
		WorkflowID:             "wf-auth",
		DecisionRef:            validTestRef("d"),
		Strategy:               strategy,
		CodecImplementationRef: ref,
		CodecProfileRef:        ref,
		CodecBinding: &LoweringBinding{
			CapabilityID:        "safe-session-codec",
			ProviderManifestRef: ref,
			QualificationRef:    qref,
			ObligationRefs:      []string{ref},
			InputShapeRef:       ref,
			OutputShapeRef:      ref,
		},
		Checks: []SessionLifecycleCheck{
			{
				ID: "chk-login",
				Observation: SessionObservation{
					Side:     SideSource,
					StepID:   "login",
					Visit:    0,
					Phase:    PhasePostOperation,
					ActorRef: "actor-alice",
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
				ID: "rel-continuity",
				Left: SessionLifecycleValue{
					Observation: SessionObservation{
						Side:     SideSource,
						StepID:   "login",
						Visit:    0,
						Phase:    PhasePostOperation,
						ActorRef: "actor-alice",
					},
					Path: "/session_id",
				},
				Right: SessionLifecycleValue{
					Observation: SessionObservation{
						Side:     SideTarget,
						StepID:   "login",
						Visit:    0,
						Phase:    PhasePostOperation,
						ActorRef: "actor-alice",
					},
					Path: "/session_id",
				},
				Different: false,
			},
		},
	}
	return c
}

func TestNarrowTargetGenerator_DeterministicGeneration(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http", "routing"}, []string{"route", "health"})
	tm := makeTestTargetModel(t, unitID)

	gen := NewNarrowTargetGenerator()
	input := GenerationInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
	}

	out1, err := gen.Generate(input)
	if err != nil {
		t.Fatalf("first generate failed: %v", err)
	}

	out2, err := gen.Generate(input)
	if err != nil {
		t.Fatalf("second generate failed: %v", err)
	}

	// Artifact paths and contents must be strictly identical
	if len(out1.Artifacts) != len(out2.Artifacts) {
		t.Fatalf("artifact count mismatch: %d vs %d", len(out1.Artifacts), len(out2.Artifacts))
	}
	for path, bytes1 := range out1.Artifacts {
		bytes2, ok := out2.Artifacts[path]
		if !ok {
			t.Fatalf("path %s missing from second output", path)
		}
		if string(bytes1) != string(bytes2) {
			t.Fatalf("path %s content nondeterministic", path)
		}
	}

	// Manifests must serialize to identical canonical JSON
	raw1, err := json.Marshal(out1.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	raw2, err := json.Marshal(out2.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw1) != string(raw2) {
		t.Fatalf("manifest nondeterministic: %s vs %s", string(raw1), string(raw2))
	}

	// ValidateGenerationManifest must succeed
	if err := ValidateGenerationManifest(raw1); err != nil {
		t.Fatalf("generated manifest validation failed: %v", err)
	}
}

func TestNarrowTargetGenerator_AuthSessionLowering(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session"}, []string{"login", "session"})
	tm := makeTestTargetModel(t, unitID)
	sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

	gen := NewNarrowTargetGenerator()
	input := GenerationInput{
		Admission:       adm,
		SessionContract: sc,
		TargetModel:     tm,
		UnitID:          unitID,
	}

	out, err := gen.Generate(input)
	if err != nil {
		t.Fatalf("generate with session failed: %v", err)
	}

	// Check middleware and config artifacts are present
	mwBytes, ok := out.Artifacts["app/Http/Middleware/SessionContinuityMiddleware.php"]
	if !ok {
		t.Fatal("missing SessionContinuityMiddleware.php")
	}
	if !strings.Contains(string(mwBytes), "SAFE_BRIDGE") {
		t.Fatalf("middleware missing SAFE_BRIDGE strategy: %s", string(mwBytes))
	}

	cfgBytes, ok := out.Artifacts["config/session-continuity.php"]
	if !ok {
		t.Fatal("missing config/session-continuity.php")
	}
	if !strings.Contains(string(cfgBytes), "'strategy' => \"SAFE_BRIDGE\"") {
		t.Fatalf("config missing SAFE_BRIDGE: %s", string(cfgBytes))
	}

	// Manifest session_contract_ref must match sc.Digest()
	expectedDigest, err := sc.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if out.Manifest.SessionContractRef != expectedDigest {
		t.Fatalf("manifest session_contract_ref mismatch: got %s, want %s", out.Manifest.SessionContractRef, expectedDigest)
	}

	raw, err := json.Marshal(out.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenerationManifest(raw); err != nil {
		t.Fatalf("manifest validation failed: %v", err)
	}

	// Also test REVIEWED_RELOGIN strategy
	scRelogin := makeTestSessionContract(t, tm.ProjectID, "login", "REVIEWED_RELOGIN")
	input.SessionContract = scRelogin
	outRelogin, err := gen.Generate(input)
	if err != nil {
		t.Fatalf("generate with relogin failed: %v", err)
	}
	mwRelogin := string(outRelogin.Artifacts["app/Http/Middleware/SessionContinuityMiddleware.php"])
	if !strings.Contains(mwRelogin, "REVIEWED_RELOGIN") {
		t.Fatalf("middleware missing REVIEWED_RELOGIN: %s", mwRelogin)
	}
}

func TestNarrowTargetGenerator_MissingSessionContract_Fails(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http", "routing", "session"}, []string{"login"})
	tm := makeTestTargetModel(t, unitID)

	gen := NewNarrowTargetGenerator()
	input := GenerationInput{
		Admission:       adm,
		SessionContract: nil, // missing despite session capability!
		TargetModel:     tm,
		UnitID:          unitID,
	}

	_, err := gen.Generate(input)
	if err == nil {
		t.Fatal("expected error when session contract is missing for session unit")
	}
	if !strings.Contains(err.Error(), "session continuity contract") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNarrowTargetGenerator_BlockedAdmission_Fails(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http"}, []string{"route"})
	adm.Outcome = "BLOCKED"
	adm.Blockers = []string{"source qualification incomplete"}
	tm := makeTestTargetModel(t, unitID)

	gen := NewNarrowTargetGenerator()
	input := GenerationInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
	}

	_, err := gen.Generate(input)
	if err == nil {
		t.Fatal("expected error on BLOCKED admission")
	}
	if !strings.Contains(err.Error(), "admission outcome is \"BLOCKED\"") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNarrowTargetGenerator_UnitNotInScope_Fails(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http"}, []string{"route"})
	tm := makeTestTargetModel(t, unitID)

	gen := NewNarrowTargetGenerator()
	input := GenerationInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      "other-unplanned-unit", // not in admission scope
	}

	_, err := gen.Generate(input)
	if err == nil {
		t.Fatal("expected error when unit_id is not in admission scope")
	}
	if !strings.Contains(err.Error(), "not in admission scope") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNarrowTargetGenerator_ValidateManifest_Vectors(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http"}, []string{"route"})
	tm := makeTestTargetModel(t, unitID)

	gen := NewNarrowTargetGenerator()
	out, err := gen.Generate(GenerationInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
	})
	if err != nil {
		t.Fatal(err)
	}

	baseRaw, err := json.Marshal(out.Manifest)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ValidManifest_Passes", func(t *testing.T) {
		if err := ValidateGenerationManifest(baseRaw); err != nil {
			t.Fatalf("expected valid manifest to pass: %v", err)
		}
	})

	t.Run("InvalidSchemaVersion_Fails", func(t *testing.T) {
		var m GenerationManifest
		json.Unmarshal(baseRaw, &m)
		m.SchemaVersion = "2.0"
		raw, _ := json.Marshal(m)
		if err := ValidateGenerationManifest(raw); err == nil {
			t.Fatal("expected error for schema_version 2.0")
		}
	})

	t.Run("InvalidProjectID_Fails", func(t *testing.T) {
		var m GenerationManifest
		json.Unmarshal(baseRaw, &m)
		m.ProjectID = "Invalid Project ID!"
		raw, _ := json.Marshal(m)
		if err := ValidateGenerationManifest(raw); err == nil {
			t.Fatal("expected error for invalid project_id")
		}
	})

	t.Run("DuplicateEntryPaths_Fails", func(t *testing.T) {
		var m GenerationManifest
		json.Unmarshal(baseRaw, &m)
		m.Entries = append(m.Entries, m.Entries[0]) // duplicate!
		raw, _ := json.Marshal(m)
		if err := ValidateGenerationManifest(raw); err == nil {
			t.Fatal("expected error for duplicate paths")
		}
	})

	t.Run("CreateWithNonNullBeforeSHA256_Fails", func(t *testing.T) {
		var m GenerationManifest
		json.Unmarshal(baseRaw, &m)
		dummy := strings.Repeat("a", 64)
		m.Entries[0].BeforeSHA256 = &dummy
		raw, _ := json.Marshal(m)
		if err := ValidateGenerationManifest(raw); err == nil {
			t.Fatal("expected error for CREATE with non-null before_sha256")
		}
	})

	t.Run("ModifyWithNullBeforeSHA256_Fails", func(t *testing.T) {
		var m GenerationManifest
		json.Unmarshal(baseRaw, &m)
		m.Entries[0].Action = "MODIFY"
		m.Entries[0].BeforeSHA256 = nil
		raw, _ := json.Marshal(m)
		if err := ValidateGenerationManifest(raw); err == nil {
			t.Fatal("expected error for MODIFY with null before_sha256")
		}
	})
}

func TestNarrowTargetGenerator_StageGeneration(t *testing.T) {
	unitID := "sync-auth"
	adm := makeTestAdmission(t, unitID, []string{"http"}, []string{"route"})
	tm := makeTestTargetModel(t, unitID)

	gen := NewNarrowTargetGenerator()
	out, err := gen.Generate(GenerationInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
	})
	if err != nil {
		t.Fatal(err)
	}

	buildWitness := validTestRef("7")

	t.Run("ValidStage_Success", func(t *testing.T) {
		staged, err := StageGeneration(out, 1, buildWitness)
		if err != nil {
			t.Fatalf("valid StageGeneration failed: %v", err)
		}
		if staged.FenceToken != 1 {
			t.Fatalf("unexpected fence token: %d", staged.FenceToken)
		}
		if staged.BuildWitness != buildWitness {
			t.Fatalf("unexpected build witness: %s", staged.BuildWitness)
		}
		if len(staged.ArtifactPaths) != len(out.Artifacts) {
			t.Fatalf("artifact paths count mismatch: %d vs %d", len(staged.ArtifactPaths), len(out.Artifacts))
		}
	})

	t.Run("NegativeFenceToken_Fails", func(t *testing.T) {
		_, err := StageGeneration(out, -1, buildWitness)
		if err == nil {
			t.Fatal("expected error for negative fence token")
		}
	})

	t.Run("InvalidBuildWitness_Fails", func(t *testing.T) {
		_, err := StageGeneration(out, 1, "not-a-sha256")
		if err == nil {
			t.Fatal("expected error for invalid build witness ref")
		}
	})

	t.Run("NilOutput_Fails", func(t *testing.T) {
		_, err := StageGeneration(nil, 1, buildWitness)
		if err == nil {
			t.Fatal("expected error for nil output")
		}
	})
}
