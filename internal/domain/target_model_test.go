package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func validTestRef(char string) string {
	return "sha256:" + strings.Repeat(char, 64)
}

func TestTargetModel_Valid21(t *testing.T) {
	builder := NewTargetModelBuilder("model-auth", "proj-test", validTestRef("a"), validTestRef("b"))
	builder.WithRevision(1)

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "auth-service",
		OwnerUnitID:          "unit-auth",
		ConsumerUnitIDs:      []string{"unit-checkout"},
		Paths:                []string{"app/Services/AuthService.php"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "laravel_service",
		RequiredCapabilities: []string{"http", "password_hash"},
		ConstraintRefs:       []string{validTestRef("d")},
		LoweringBindings: []LoweringBinding{
			{
				CapabilityID:        "password_hash",
				ProviderManifestRef: validTestRef("e"),
				QualificationRef:    validTestRef("f"),
				ObligationRefs:      []string{validTestRef("1")},
				InputShapeRef:       validTestRef("2"),
				OutputShapeRef:      validTestRef("3"),
			},
		},
	})

	doc, err := builder.Build()
	if err != nil {
		t.Fatalf("expected valid 2.1 model build, got: %v", err)
	}

	if doc.SchemaVersion != "2.1" {
		t.Errorf("expected 2.1 schema version, got: %s", doc.SchemaVersion)
	}
	if doc.HasBlockingGaps() {
		t.Errorf("expected no blocking gaps")
	}

	digest, err := doc.Digest()
	if err != nil {
		t.Fatalf("expected digest calculation, got: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Errorf("invalid digest format: %s", digest)
	}

	comp, ok := doc.ComponentByOwner("unit-auth")
	if !ok || comp.ComponentID != "auth-service" {
		t.Errorf("expected component by owner unit-auth")
	}

	consumers := doc.ConsumersForUnit("unit-auth")
	if len(consumers) != 1 || consumers[0] != "unit-checkout" {
		t.Errorf("expected unit-checkout as consumer, got: %v", consumers)
	}
}

func TestTargetModel_Valid20(t *testing.T) {
	builder := NewTargetModelBuilder("model-v20", "proj-test", validTestRef("a"), validTestRef("b"))
	builder.WithSchemaVersion("2.0").WithRevision(1)

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "auth-service",
		OwnerUnitID:          "unit-auth",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"auth.go"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
	})

	doc, err := builder.Build()
	if err != nil {
		t.Fatalf("expected valid 2.0 model build, got: %v", err)
	}

	raw, err := doc.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// In 2.0, lowering_bindings must NOT be serialized
	if strings.Contains(string(raw), "lowering_bindings") {
		t.Errorf("version 2.0 must not serialize lowering_bindings: %s", string(raw))
	}

	if err := ValidateTargetModel(raw); err != nil {
		t.Fatalf("expected 2.0 model validation to pass, got: %v", err)
	}
}

func TestTargetModel_ComponentOwnerUnitIDUniqueness(t *testing.T) {
	// Invariant: "Component owner_unit_id is unique" (TARGET-ARCHITECTURE.md:9)
	builder := NewTargetModelBuilder("model-owner-dup", "proj-test", validTestRef("a"), validTestRef("b"))

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-1",
		OwnerUnitID:          "unit-shared",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"comp1.go"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-2",
		OwnerUnitID:          "unit-shared", // Duplicate owner_unit_id!
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"comp2.go"},
		ContractRefs:         []string{validTestRef("d")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error due to duplicate owner_unit_id across components, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate owner_unit_id") {
		t.Errorf("expected duplicate owner_unit_id in error message, got: %v", err)
	}
}

func TestTargetModel_DuplicateComponentID(t *testing.T) {
	builder := NewTargetModelBuilder("model-comp-dup", "proj-test", validTestRef("a"), validTestRef("b"))

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-same",
		OwnerUnitID:          "unit-1",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"comp1.go"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-same", // Duplicate component_id!
		OwnerUnitID:          "unit-2",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"comp2.go"},
		ContractRefs:         []string{validTestRef("d")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error due to duplicate component_id, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate component_id") {
		t.Errorf("expected duplicate component_id in error, got: %v", err)
	}
}

func TestTargetModel_PathsOverlapAmbiguously(t *testing.T) {
	// Invariant: "Paths cannot overlap owners ambiguously." (TARGET-ARCHITECTURE.md:9)

	t.Run("exact duplicate path between owners", func(t *testing.T) {
		builder := NewTargetModelBuilder("model-path-dup", "proj-test", validTestRef("a"), validTestRef("b"))

		builder.AddComponent(TargetModelComponent{
			ComponentID:          "comp-1",
			OwnerUnitID:          "unit-1",
			ConsumerUnitIDs:      []string{},
			Paths:                []string{"routes/api.php"},
			ContractRefs:         []string{validTestRef("c")},
			LoweringChoice:       "synthetic",
			RequiredCapabilities: []string{"http"},
			ConstraintRefs:       []string{},
			LoweringBindings:     []LoweringBinding{},
		})

		builder.AddComponent(TargetModelComponent{
			ComponentID:          "comp-2",
			OwnerUnitID:          "unit-2",
			ConsumerUnitIDs:      []string{},
			Paths:                []string{"routes/api.php"}, // Same path!
			ContractRefs:         []string{validTestRef("d")},
			LoweringChoice:       "synthetic",
			RequiredCapabilities: []string{"http"},
			ConstraintRefs:       []string{},
			LoweringBindings:     []LoweringBinding{},
		})

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error due to path conflict, got nil")
		}
		if !strings.Contains(err.Error(), "path conflict") {
			t.Errorf("expected path conflict error, got: %v", err)
		}
	})

	t.Run("directory prefix overlap between owners", func(t *testing.T) {
		builder := NewTargetModelBuilder("model-path-prefix", "proj-test", validTestRef("a"), validTestRef("b"))

		builder.AddComponent(TargetModelComponent{
			ComponentID:          "comp-1",
			OwnerUnitID:          "unit-1",
			ConsumerUnitIDs:      []string{},
			Paths:                []string{"app/Services"},
			ContractRefs:         []string{validTestRef("c")},
			LoweringChoice:       "synthetic",
			RequiredCapabilities: []string{"http"},
			ConstraintRefs:       []string{},
			LoweringBindings:     []LoweringBinding{},
		})

		builder.AddComponent(TargetModelComponent{
			ComponentID:          "comp-2",
			OwnerUnitID:          "unit-2",
			ConsumerUnitIDs:      []string{},
			Paths:                []string{"app/Services/SubService.php"}, // Inside comp-1 path!
			ContractRefs:         []string{validTestRef("d")},
			LoweringChoice:       "synthetic",
			RequiredCapabilities: []string{"http"},
			ConstraintRefs:       []string{},
			LoweringBindings:     []LoweringBinding{},
		})

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error due to ambiguous path overlap, got nil")
		}
		if !strings.Contains(err.Error(), "ambiguous path overlap") {
			t.Errorf("expected ambiguous path overlap error, got: %v", err)
		}
	})
}

func TestTargetModel_OwnerCannotConsumeItself(t *testing.T) {
	builder := NewTargetModelBuilder("model-self-consume", "proj-test", validTestRef("a"), validTestRef("b"))

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-1",
		OwnerUnitID:          "unit-login",
		ConsumerUnitIDs:      []string{"unit-login"}, // Owner listed as consumer!
		Paths:                []string{"auth.go"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error when owner consumes itself, got nil")
	}
	if !strings.Contains(err.Error(), "cannot contain its own owner_unit_id") {
		t.Errorf("expected cannot contain its own owner_unit_id error, got: %v", err)
	}
}

func TestTargetModel_NonprimitiveCapabilityCoverage(t *testing.T) {
	// Invariant: "Primitive core capabilities need no provider binding; each nonprimitive requirement
	// (password verification, session codec, template rendering) must resolve exactly one qualified
	// binding or remain a blocking gap." (TARGET-ARCHITECTURE.md:23)

	t.Run("unbound nonprimitive capability without declared gap rejected", func(t *testing.T) {
		builder := NewTargetModelBuilder("model-unbound-cap", "proj-test", validTestRef("a"), validTestRef("b"))

		builder.AddComponent(TargetModelComponent{
			ComponentID:          "comp-auth",
			OwnerUnitID:          "unit-auth",
			ConsumerUnitIDs:      []string{},
			Paths:                []string{"auth.go"},
			ContractRefs:         []string{validTestRef("c")},
			LoweringChoice:       "synthetic",
			RequiredCapabilities: []string{"password_hash"}, // Non-primitive!
			ConstraintRefs:       []string{},
			LoweringBindings:     []LoweringBinding{}, // Empty bindings!
		})

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error due to unbound nonprimitive capability, got nil")
		}
		if !strings.Contains(err.Error(), "without a qualified lowering binding or declared capability gap") {
			t.Errorf("expected unbound capability error, got: %v", err)
		}
	})

	t.Run("unbound nonprimitive capability with declared gap accepted", func(t *testing.T) {
		builder := NewTargetModelBuilder("model-gap-cap", "proj-test", validTestRef("a"), validTestRef("b"))

		builder.AddComponent(TargetModelComponent{
			ComponentID:          "comp-auth",
			OwnerUnitID:          "unit-auth",
			ConsumerUnitIDs:      []string{},
			Paths:                []string{"auth.go"},
			ContractRefs:         []string{validTestRef("c")},
			LoweringChoice:       "synthetic",
			RequiredCapabilities: []string{"password_hash"},
			ConstraintRefs:       []string{},
			LoweringBindings:     []LoweringBinding{},
		})

		builder.AddCapabilityGap(CapabilityGap{
			Code:         "MISSING_PASSWORD_HASH_PROVIDER",
			ComponentIDs: []string{"comp-auth"},
			EvidenceRefs: []string{validTestRef("e")},
			Blocking:     true,
		})

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected declared gap to permit model compilation, got: %v", err)
		}
		if !doc.HasBlockingGaps() {
			t.Errorf("expected model to have blocking gaps")
		}
		if len(doc.BlockingGaps()) != 1 {
			t.Errorf("expected 1 blocking gap, got: %d", len(doc.BlockingGaps()))
		}
	})
}

func TestTargetModel_CapabilityGapUnknownComponent(t *testing.T) {
	builder := NewTargetModelBuilder("model-bad-gap", "proj-test", validTestRef("a"), validTestRef("b"))

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-1",
		OwnerUnitID:          "unit-1",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"comp.go"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	builder.AddCapabilityGap(CapabilityGap{
		Code:         "UNSUPPORTED_FACET",
		ComponentIDs: []string{"comp-nonexistent"}, // Not in components!
		EvidenceRefs: []string{validTestRef("e")},
		Blocking:     true,
	})

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error due to unknown component_id in gap, got nil")
	}
	if !strings.Contains(err.Error(), "references unknown component_id") {
		t.Errorf("expected unknown component_id in error, got: %v", err)
	}
}

func TestTargetModel_SuccessorModelCreation(t *testing.T) {
	// Invariant: "Changing ownership/lowering/resolution creates a successor model and Gate 2." (TARGET-ARCHITECTURE.md:13)
	builder := NewTargetModelBuilder("model-auth", "proj-test", validTestRef("a"), validTestRef("b"))
	builder.WithRevision(1)

	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-auth",
		OwnerUnitID:          "unit-auth",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"auth.go"},
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	doc1, err := builder.Build()
	if err != nil {
		t.Fatalf("build doc1: %v", err)
	}

	successorBuilder := doc1.CreateSuccessor(2)
	successorBuilder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-session",
		OwnerUnitID:          "unit-session",
		ConsumerUnitIDs:      []string{},
		Paths:                []string{"session.go"},
		ContractRefs:         []string{validTestRef("d")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})

	doc2, err := successorBuilder.Build()
	if err != nil {
		t.Fatalf("build doc2: %v", err)
	}

	if doc2.Revision != 2 {
		t.Errorf("expected successor revision 2, got: %d", doc2.Revision)
	}
	if len(doc2.Components) != 2 {
		t.Errorf("expected 2 components in successor, got: %d", len(doc2.Components))
	}
	if len(doc1.Components) != 1 {
		t.Errorf("original doc1 components must not be mutated, got: %d", len(doc1.Components))
	}

	digest1, _ := doc1.Digest()
	digest2, _ := doc2.Digest()
	if digest1 == digest2 {
		t.Errorf("successor model digest must differ from original model digest")
	}
}

func TestTargetModel_ValidationCasesGoldenParity(t *testing.T) {
	rawGolden, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		// Try alternate path relative to repo root
		rawGolden, err = os.ReadFile("specs/examples/validation-cases.json")
		if err != nil {
			t.Skipf("cannot locate validation-cases.json: %v", err)
		}
	}

	type validationCase struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}

	var cases []validationCase
	if err := json.Unmarshal(rawGolden, &cases); err != nil {
		t.Fatalf("unmarshal golden cases: %v", err)
	}

	tested := 0
	for _, c := range cases {
		if c.Schema != "https://rewrotron.dev/specs/migration/target-model.schema.json" {
			continue
		}
		tested++
		err := ValidateTargetModel(c.Instance)
		if c.Valid && err != nil {
			t.Errorf("case %q expected valid, got error: %v", c.Name, err)
		}
		if !c.Valid && err == nil {
			t.Errorf("case %q expected invalid, got nil error", c.Name)
		}
	}

	if tested < 3 {
		t.Errorf("expected at least 3 target-model golden cases tested, got: %d", tested)
	}
}
