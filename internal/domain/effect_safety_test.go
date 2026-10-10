package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func makeTestEffectAdmission(t *testing.T, unitID string) *GenerationAdmission {
	t.Helper()
	plan, checks := admissionTestPlan(t)
	adm, err := EvaluateGenerationAdmission(plan, "sync-auth", checks)
	if err != nil {
		t.Fatal(err)
	}
	adm.Scope.UnitIDs = []string{unitID}
	adm.Scope.CaseIDs = []string{"order-created", "payment-effect"}
	adm.Capabilities = []string{"http", "routing"}
	return adm
}

func makeTestEffectTargetModel(t *testing.T, unitID string) *TargetModelDocument {
	t.Helper()
	builder := NewTargetModelBuilder("model-effects", "proj-test", validTestRef("a"), validTestRef("b"))
	builder.WithRevision(1)
	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-effects-" + unitID,
		OwnerUnitID:          unitID,
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"app/Services/OrderService.php"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "laravel_events",
		RequiredCapabilities: []string{"http", "routing"},
		ConstraintRefs:       []string{validTestRef("d")},
	})
	doc, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestEffectSafetyPolicy(t *testing.T) {
	prodPolicy := EffectSafetyPolicy{IsProductionEnvironment: true, AllowIrreversible: false}
	sandboxPolicy := EffectSafetyPolicy{IsProductionEnvironment: false, AllowIrreversible: false}
	isolatedPolicy := EffectSafetyPolicy{IsProductionEnvironment: false, AllowIrreversible: true}

	pureEff := BusinessEffectRecord{EffectID: "eff-1", Dimension: EffectDimensionDBRead, EffectClass: EffectClassPure}
	readEff := BusinessEffectRecord{EffectID: "eff-2", Dimension: EffectDimensionDBRead, EffectClass: EffectClassReadOnly}
	writeEff := BusinessEffectRecord{EffectID: "eff-3", Dimension: EffectDimensionDBWrite, EffectClass: EffectClassLocalWrite}
	extEff := BusinessEffectRecord{EffectID: "eff-4", Dimension: EffectDimensionExternal, EffectClass: EffectClassExternalWrite}
	irrevEff := BusinessEffectRecord{EffectID: "eff-5", Dimension: EffectDimensionExternal, EffectClass: EffectClassIrreversible}

	// Production checks
	if err := prodPolicy.CheckEffectSafety(pureEff); err != nil {
		t.Fatalf("pure effect should be allowed in prod: %v", err)
	}
	if err := prodPolicy.CheckEffectSafety(readEff); err != nil {
		t.Fatalf("read-only effect should be allowed in prod: %v", err)
	}
	if err := prodPolicy.CheckEffectSafety(writeEff); err == nil || !strings.Contains(err.Error(), "ERR_PRODUCTION_EFFECT_REFUSED") {
		t.Fatalf("local write should be rejected in prod, got: %v", err)
	}
	if err := prodPolicy.CheckEffectSafety(extEff); err == nil || !strings.Contains(err.Error(), "ERR_PRODUCTION_EFFECT_REFUSED") {
		t.Fatalf("external write should be rejected in prod, got: %v", err)
	}

	// Sandbox checks
	if err := sandboxPolicy.CheckEffectSafety(writeEff); err != nil {
		t.Fatalf("local write should be allowed in sandbox: %v", err)
	}
	if err := sandboxPolicy.CheckEffectSafety(irrevEff); err == nil || !strings.Contains(err.Error(), "ERR_IRREVERSIBLE_EFFECT_REFUSED") {
		t.Fatalf("irreversible effect without flag should be rejected in sandbox: %v", err)
	}

	// Isolated checks
	if err := isolatedPolicy.CheckEffectSafety(irrevEff); err != nil {
		t.Fatalf("irreversible effect should be allowed in isolated sandbox: %v", err)
	}
}

func TestVerifyEffectCausalityAndOrdering(t *testing.T) {
	parentID := "eff-parent"

	t.Run("ValidMonotonicAndCausal_Succeeds", func(t *testing.T) {
		effects := []BusinessEffectRecord{
			{EffectID: "eff-parent", SequenceIndex: 1},
			{EffectID: "eff-child", SequenceIndex: 2, CausalParentID: &parentID},
			{EffectID: "eff-sibling", SequenceIndex: 3},
		}
		if err := VerifyEffectCausalityAndOrdering(effects); err != nil {
			t.Fatalf("valid sequence should pass: %v", err)
		}
	})

	t.Run("NonMonotonicSequence_Fails", func(t *testing.T) {
		effects := []BusinessEffectRecord{
			{EffectID: "eff-1", SequenceIndex: 2},
			{EffectID: "eff-2", SequenceIndex: 1}, // non-monotonic!
		}
		err := VerifyEffectCausalityAndOrdering(effects)
		if err == nil || !strings.Contains(err.Error(), "ERR_ORDERING_VIOLATION") {
			t.Fatalf("expected ERR_ORDERING_VIOLATION, got: %v", err)
		}
	})

	t.Run("DuplicateSequenceIndex_Fails", func(t *testing.T) {
		effects := []BusinessEffectRecord{
			{EffectID: "eff-1", SequenceIndex: 1},
			{EffectID: "eff-2", SequenceIndex: 1}, // duplicate!
		}
		err := VerifyEffectCausalityAndOrdering(effects)
		if err == nil || !strings.Contains(err.Error(), "ERR_ORDERING_VIOLATION") {
			t.Fatalf("expected ERR_ORDERING_VIOLATION for duplicate index, got: %v", err)
		}
	})

	t.Run("CausalInversion_Fails", func(t *testing.T) {
		futureParent := "eff-future"
		effects := []BusinessEffectRecord{
			{EffectID: "eff-child", SequenceIndex: 1, CausalParentID: &futureParent}, // parent not seen!
			{EffectID: "eff-future", SequenceIndex: 2},
		}
		err := VerifyEffectCausalityAndOrdering(effects)
		if err == nil || !strings.Contains(err.Error(), "ERR_CAUSAL_INVERSION") {
			t.Fatalf("expected ERR_CAUSAL_INVERSION, got: %v", err)
		}
	})
}

func TestVerifyEffectCardinality(t *testing.T) {
	src := []BusinessEffectRecord{
		{
			EffectID:              "src-1",
			LogicalEffectID:       "log-order-1",
			BusinessKey:           "order-101",
			DestinationID:         "dest-order-queue",
			BusinessApplicationID: "app-order-101",
			Applied:               true,
		},
	}

	t.Run("IdenticalEffects_Passes", func(t *testing.T) {
		tgt := []BusinessEffectRecord{
			{
				EffectID:              "tgt-1",
				LogicalEffectID:       "log-order-1",
				BusinessKey:           "order-101",
				DestinationID:         "dest-order-queue",
				BusinessApplicationID: "app-order-101",
				Applied:               true,
			},
		}
		res := VerifyEffectCardinality(src, tgt)
		if res.Verdict != "PASS" {
			t.Fatalf("expected PASS, got %s: %s", res.Verdict, res.Reason)
		}
	})

	t.Run("MissingEffect_Mismatch", func(t *testing.T) {
		tgt := []BusinessEffectRecord{} // target omitted effect!
		res := VerifyEffectCardinality(src, tgt)
		if res.Verdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH for missing effect, got %s", res.Verdict)
		}
		if len(res.MismatchIDs) == 0 || res.MismatchIDs[0] != "mismatch-missing-effect" {
			t.Fatalf("expected mismatch-missing-effect, got: %v", res.MismatchIDs)
		}
	})

	t.Run("DuplicateAppliedEffect_Mismatch", func(t *testing.T) {
		tgt := []BusinessEffectRecord{
			{
				EffectID:              "tgt-1",
				LogicalEffectID:       "log-order-1",
				BusinessKey:           "order-101",
				DestinationID:         "dest-order-queue",
				BusinessApplicationID: "app-order-101",
				Applied:               true,
			},
			{
				EffectID:              "tgt-2",
				LogicalEffectID:       "log-order-1",
				BusinessKey:           "order-101",
				DestinationID:         "dest-order-queue",
				BusinessApplicationID: "app-order-101",
				Applied:               true, // second applied effect! duplicate cardinality violation
			},
		}
		res := VerifyEffectCardinality(src, tgt)
		if res.Verdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH for duplicate applied effect, got %s", res.Verdict)
		}
		if len(res.MismatchIDs) == 0 || res.MismatchIDs[0] != "mismatch-duplicate-effect" {
			t.Fatalf("expected mismatch-duplicate-effect, got: %v", res.MismatchIDs)
		}
	})

	t.Run("MisboundDestination_Mismatch", func(t *testing.T) {
		tgt := []BusinessEffectRecord{
			{
				EffectID:              "tgt-1",
				LogicalEffectID:       "log-order-1",
				BusinessKey:           "order-101",
				DestinationID:         "dest-wrong-queue", // wrong destination!
				BusinessApplicationID: "app-order-101",
				Applied:               true,
			},
		}
		res := VerifyEffectCardinality(src, tgt)
		if res.Verdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH for misbound destination, got %s", res.Verdict)
		}
		if len(res.MismatchIDs) == 0 || res.MismatchIDs[0] != "mismatch-misbound-effect" {
			t.Fatalf("expected mismatch-misbound-effect, got: %v", res.MismatchIDs)
		}
	})

	t.Run("MisboundBusinessKey_Mismatch", func(t *testing.T) {
		tgt := []BusinessEffectRecord{
			{
				EffectID:              "tgt-1",
				LogicalEffectID:       "log-order-1",
				BusinessKey:           "order-999-wrong", // wrong key!
				DestinationID:         "dest-order-queue",
				BusinessApplicationID: "app-order-101",
				Applied:               true,
			},
		}
		res := VerifyEffectCardinality(src, tgt)
		if res.Verdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH for misbound business key, got %s", res.Verdict)
		}
		if len(res.MismatchIDs) == 0 || res.MismatchIDs[0] != "mismatch-misbound-effect" {
			t.Fatalf("expected mismatch-misbound-effect, got: %v", res.MismatchIDs)
		}
	})
}

func TestEffectReconstructor_DeterministicReconstruction(t *testing.T) {
	unitID := "unit-effects"
	adm := makeTestEffectAdmission(t, unitID)
	tm := makeTestEffectTargetModel(t, unitID)

	parentID := "eff-1"
	effects := []BusinessEffectRecord{
		{
			EffectID:              "eff-1",
			LogicalEffectID:       "log-1",
			Dimension:             EffectDimensionDBWrite,
			EffectClass:           EffectClassLocalWrite,
			BusinessKey:           "order-100",
			DestinationID:         "dest-db",
			BusinessApplicationID: "app-100",
			SequenceIndex:         1,
			PayloadRef:            validTestRef("1"),
			Applied:               true,
		},
		{
			EffectID:              "eff-2",
			LogicalEffectID:       "log-2",
			Dimension:             EffectDimensionQueue,
			EffectClass:           EffectClassLocalWrite,
			BusinessKey:           "order-100",
			DestinationID:         "dest-queue",
			BusinessApplicationID: "app-100-q",
			SequenceIndex:         2,
			CausalParentID:        &parentID,
			PayloadRef:            validTestRef("2"),
			Applied:               true,
		},
	}

	reconstructor := NewEffectReconstructor()
	input := EffectReconstructionInput{
		Admission:    adm,
		TargetModel:  tm,
		UnitID:       unitID,
		Effects:      effects,
		SafetyPolicy: EffectSafetyPolicy{IsProductionEnvironment: false, AllowIrreversible: false},
	}

	out1, err := reconstructor.Reconstruct(input)
	if err != nil {
		t.Fatalf("first reconstruction failed: %v", err)
	}

	out2, err := reconstructor.Reconstruct(input)
	if err != nil {
		t.Fatalf("second reconstruction failed: %v", err)
	}

	// Deterministic serialization
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

	if err := ValidateEffectPublicationManifest(raw1); err != nil {
		t.Fatalf("manifest validation failed: %v", err)
	}
}

func TestValidateEffectPublicationManifest_Vectors(t *testing.T) {
	unitID := "unit-effects"
	adm := makeTestEffectAdmission(t, unitID)
	tm := makeTestEffectTargetModel(t, unitID)

	reconstructor := NewEffectReconstructor()
	out, err := reconstructor.Reconstruct(EffectReconstructionInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
		Effects: []BusinessEffectRecord{
			{
				EffectID:        "eff-1",
				LogicalEffectID: "log-1",
				Dimension:       EffectDimensionDBWrite,
				EffectClass:     EffectClassLocalWrite,
				PayloadRef:      validTestRef("1"),
				SequenceIndex:   1,
				Applied:         true,
			},
		},
		SafetyPolicy: EffectSafetyPolicy{IsProductionEnvironment: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	baseRaw, err := json.Marshal(out.Manifest)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ValidManifest_Passes", func(t *testing.T) {
		if err := ValidateEffectPublicationManifest(baseRaw); err != nil {
			t.Fatalf("expected valid manifest to pass: %v", err)
		}
	})

	t.Run("InvalidSchemaVersion_Fails", func(t *testing.T) {
		var m EffectPublicationManifest
		json.Unmarshal(baseRaw, &m)
		m.SchemaVersion = "2.0"
		raw, _ := json.Marshal(m)
		if err := ValidateEffectPublicationManifest(raw); err == nil {
			t.Fatal("expected error for schema_version 2.0")
		}
	})

	t.Run("DuplicateEffectID_Fails", func(t *testing.T) {
		var m EffectPublicationManifest
		json.Unmarshal(baseRaw, &m)
		m.Effects = append(m.Effects, m.Effects[0]) // duplicate!
		raw, _ := json.Marshal(m)
		if err := ValidateEffectPublicationManifest(raw); err == nil {
			t.Fatal("expected error for duplicate effect_id")
		}
	})

	t.Run("InvalidDimension_Fails", func(t *testing.T) {
		var m EffectPublicationManifest
		json.Unmarshal(baseRaw, &m)
		m.Effects[0].Dimension = "INVALID_DIM"
		raw, _ := json.Marshal(m)
		if err := ValidateEffectPublicationManifest(raw); err == nil {
			t.Fatal("expected error for invalid dimension")
		}
	})
}
