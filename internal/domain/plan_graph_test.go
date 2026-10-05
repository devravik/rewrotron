package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPlanGraph_ValidTopologicalSort(t *testing.T) {
	// Units:
	// - auth (priority 10, no prerequisites)
	// - session (priority 20, depends on auth)
	// - profile (priority 10, depends on session)
	// - telemetry (priority 5, no prerequisites)
	builder := NewPlanGraphBuilder("proj-test", validTestRef("c"), validTestRef("d"), nil)

	builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "auth",
		ContractRefs: []string{validTestRef("1")},
		CaseIDs:      []string{"case-auth"},
		Priority:     10,
		Capabilities: []string{"http"},
	})
	builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "session",
		ContractRefs: []string{validTestRef("2")},
		CaseIDs:      []string{"case-session"},
		Priority:     20,
		Capabilities: []string{"http"},
	})
	builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "profile",
		ContractRefs: []string{validTestRef("3")},
		CaseIDs:      []string{"case-profile"},
		Priority:     10,
		Capabilities: []string{"http"},
	})
	builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "telemetry",
		ContractRefs: []string{validTestRef("4")},
		CaseIDs:      []string{"case-telemetry"},
		Priority:     5,
		Capabilities: []string{"logging"},
	})

	builder.AddDependency("auth", "session")
	builder.AddDependency("session", "profile")

	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("expected valid plan graph, got: %v", err)
	}

	if len(graph.OrderedUnits) != 4 {
		t.Fatalf("expected 4 ordered units, got: %d", len(graph.OrderedUnits))
	}

	// Verify prerequisite ordering:
	orderMap := make(map[string]int)
	for i, u := range graph.OrderedUnits {
		orderMap[u.UnitID] = i
	}

	if orderMap["auth"] > orderMap["session"] {
		t.Errorf("auth must precede session, got auth at %d, session at %d", orderMap["auth"], orderMap["session"])
	}
	if orderMap["session"] > orderMap["profile"] {
		t.Errorf("session must precede profile, got session at %d, profile at %d", orderMap["session"], orderMap["profile"])
	}

	// Between auth (p=10) and telemetry (p=5), auth has higher priority so auth must precede telemetry
	if orderMap["auth"] > orderMap["telemetry"] {
		t.Errorf("auth (priority 10) must precede telemetry (priority 5)")
	}
}

func TestPlanGraph_CyclicExecutionDAGBlocked(t *testing.T) {
	// Invariant: "A cyclic execution DAG is a plan blocker." (MIGRATION-UNITS.md:3)
	t.Run("two-node cycle", func(t *testing.T) {
		builder := NewPlanGraphBuilder("proj-test", validTestRef("c"), validTestRef("d"), nil)

		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "unit-a",
			ContractRefs: []string{validTestRef("1")},
			CaseIDs:      []string{"case-a"},
			Priority:     1,
		})
		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "unit-b",
			ContractRefs: []string{validTestRef("2")},
			CaseIDs:      []string{"case-b"},
			Priority:     1,
		})

		builder.AddDependency("unit-a", "unit-b")
		builder.AddDependency("unit-b", "unit-a") // Cycle!

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error due to cyclic execution DAG, got nil")
		}
		if !strings.Contains(err.Error(), "cyclic execution DAG is a plan blocker") {
			t.Errorf("expected cyclic DAG error, got: %v", err)
		}
	})

	t.Run("self-loop dependency", func(t *testing.T) {
		builder := NewPlanGraphBuilder("proj-test", validTestRef("c"), validTestRef("d"), nil)

		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "unit-self",
			ContractRefs: []string{validTestRef("1")},
			CaseIDs:      []string{"case-1"},
			Priority:     1,
		})

		builder.AddDependency("unit-self", "unit-self")

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error due to self-loop dependency, got nil")
		}
		if !strings.Contains(err.Error(), "self-loop dependency") {
			t.Errorf("expected self-loop error, got: %v", err)
		}
	})
}

func TestPlanGraph_TestX73_MissingConsumptionEdge(t *testing.T) {
	// Test X73: "Missing consumption edge: Known source/shared-component input consumed by completed unit;
	// remove edge or entire graph and change consumed input -> Reuse refused" (TESTING.md, TARGET-ARCHITECTURE.md:9)
	timBuilder := NewTargetModelBuilder("model-test", "proj-test", validTestRef("a"), validTestRef("b"))
	timBuilder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-shared-auth",
		OwnerUnitID:          "auth",
		ConsumerUnitIDs:      []string{"checkout"}, // TIM lists checkout as consumer of auth!
		Paths:                []string{"app/Services/AuthService.php"},
		ContractRefs:         []string{validTestRef("1")},
		LoweringChoice:       "synthetic",
		RequiredCapabilities: []string{"http"},
		ConstraintRefs:       []string{},
		LoweringBindings:     []LoweringBinding{},
	})
	tim, err := timBuilder.Build()
	if err != nil {
		t.Fatalf("build TIM: %v", err)
	}

	t.Run("missing dependency edge from auth to consumer checkout rejected (Test X73)", func(t *testing.T) {
		builder := NewPlanGraphBuilder("proj-test", validTestRef("c"), validTestRef("d"), tim)
		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "auth",
			ContractRefs: []string{validTestRef("1")},
			CaseIDs:      []string{"case-auth"},
			Priority:     10,
		})
		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "checkout",
			ContractRefs: []string{validTestRef("2")},
			CaseIDs:      []string{"case-checkout"},
			Priority:     5,
		})
		// Edge auth -> checkout omitted!

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error due to missing consumption edge (Test X73), got nil")
		}
		if !strings.Contains(err.Error(), "missing consumption edge") || !strings.Contains(err.Error(), "Test X73") {
			t.Errorf("expected Test X73 missing consumption edge error, got: %v", err)
		}
	})

	t.Run("declared consumption edge matches TIM accepted", func(t *testing.T) {
		builder := NewPlanGraphBuilder("proj-test", validTestRef("c"), validTestRef("d"), tim)
		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "auth",
			ContractRefs: []string{validTestRef("1")},
			CaseIDs:      []string{"case-auth"},
			Priority:     10,
		})
		builder.AddUnit(SchedulableBehavioralUnit{
			UnitID:       "checkout",
			ContractRefs: []string{validTestRef("2")},
			CaseIDs:      []string{"case-checkout"},
			Priority:     5,
		})
		builder.AddDependency("auth", "checkout") // Declared!

		graph, err := builder.Build()
		if err != nil {
			t.Fatalf("expected successful build with matching consumption edge, got: %v", err)
		}
		if len(graph.OrderedUnits) != 2 {
			t.Errorf("expected 2 ordered units, got: %d", len(graph.OrderedUnits))
		}
	})
}

func TestPlanGraph_TestX35_ScopeDeadlockAndAcceptanceScopes(t *testing.T) {
	// Test X35: "Scope deadlock: Login prerequisite completes on its UNIT+implemented regressions
	// before checkout; project missing checkout cannot VERIFIED/READY_CUTOVER | M1/M6"
	builder := NewPlanGraphBuilder("proj-test", validTestRef("c"), validTestRef("d"), nil)

	builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "login",
		ContractRefs: []string{validTestRef("1")},
		CaseIDs:      []string{"case-login"},
		Priority:     10,
	})
	builder.AddUnit(SchedulableBehavioralUnit{
		UnitID:       "checkout",
		ContractRefs: []string{validTestRef("2")},
		CaseIDs:      []string{"case-checkout"},
		Priority:     5,
	})

	builder.AddDependency("login", "checkout")

	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("build graph: %v", err)
	}

	// 1. Verify UNIT scope derivation and schema validity
	unitScope, err := graph.DeriveUnitScope("login")
	if err != nil {
		t.Fatalf("derive unit scope: %v", err)
	}
	if unitScope.Kind != "UNIT" || len(unitScope.UnitIDs) != 1 || unitScope.UnitIDs[0] != "login" {
		t.Errorf("invalid unit scope content: %v", unitScope)
	}

	rawScope, _ := json.Marshal(unitScope)
	if err := ValidateAcceptanceScope(rawScope); err != nil {
		t.Errorf("derived unit scope failed schema validation: %v", err)
	}

	// 2. Test X35: Login can complete on UNIT scope + implemented regressions without checkout implemented
	loginVerdicts := map[string]string{
		"login": "VERIFIED",
	}
	canCompleteLogin, reason := graph.CanUnitComplete("login", []string{"login"}, loginVerdicts)
	if !canCompleteLogin {
		t.Errorf("login must be able to complete without checkout implemented (Test X35), rejected with: %s", reason)
	}

	// 3. Test X35: Project missing checkout cannot complete / achieve READY_CUTOVER
	canCompleteProject, reason := graph.CanProjectComplete([]string{"login"}, loginVerdicts)
	if canCompleteProject {
		t.Fatal("project missing checkout must NOT complete (Test X35)")
	}
	if !strings.Contains(reason, "project missing unit checkout") || !strings.Contains(reason, "Test X35") {
		t.Errorf("expected Test X35 missing unit checkout error, got: %s", reason)
	}

	// 4. Test X35: Checkout subsequently checks login regressions
	regScope, err := graph.DeriveImplementedRegressionScope([]string{"login", "checkout"})
	if err != nil {
		t.Fatalf("derive regression scope: %v", err)
	}
	rawRegScope, _ := json.Marshal(regScope)
	if err := ValidateAcceptanceScope(rawRegScope); err != nil {
		t.Errorf("derived regression scope failed schema validation: %v", err)
	}
	if len(regScope.CaseIDs) != 2 {
		t.Errorf("expected 2 case IDs in regression scope, got: %v", regScope.CaseIDs)
	}

	// 5. When checkout is also implemented and VERIFIED, project can complete
	allVerdicts := map[string]string{
		"login":    "VERIFIED",
		"checkout": "VERIFIED",
	}
	canCompleteProjectAll, reason := graph.CanProjectComplete([]string{"login", "checkout"}, allVerdicts)
	if !canCompleteProjectAll {
		t.Errorf("expected project completion when all units VERIFIED, got: %s", reason)
	}

	// 6. Verify PROJECT scope derivation and schema validity
	projScope, err := graph.DeriveProjectScope()
	if err != nil {
		t.Fatalf("derive project scope: %v", err)
	}
	rawProjScope, _ := json.Marshal(projScope)
	if err := ValidateAcceptanceScope(rawProjScope); err != nil {
		t.Errorf("derived project scope failed schema validation: %v", err)
	}
}
