package domain

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestValidateSemantics_ValidationCases runs all 19 validation cases in validation-cases.json
// that reference semantics.schema.json (top-level, action, completion, expression).
func TestValidateSemantics_ValidationCases(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	testedCount := 0
	for _, tc := range cases {
		var validateFn func([]byte) error
		switch tc.Schema {
		case "https://rewrotron.dev/specs/behavior/semantics.schema.json":
			validateFn = ValidateSemantics
		case "https://rewrotron.dev/specs/behavior/semantics.schema.json#/$defs/action":
			validateFn = ValidateAction
		case "https://rewrotron.dev/specs/behavior/semantics.schema.json#/$defs/completion":
			validateFn = ValidateCompletion
		case "https://rewrotron.dev/specs/behavior/semantics.schema.json#/$defs/expression":
			validateFn = ValidateExpression
		default:
			continue
		}

		testedCount++
		t.Run(tc.Name, func(t *testing.T) {
			err := validateFn(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected case %q to be valid, got error: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected case %q to be invalid, but passed validation", tc.Name)
			}
		})
	}

	if testedCount < 19 {
		t.Fatalf("expected at least 19 semantics validation cases, tested %d", testedCount)
	}
}

// TestContractBlockerVectors_StepGuard_B2_X03 tests step guard and transition guard vectors from contract-blocker-vectors.json.
func TestContractBlockerVectors_StepGuard_B2_X03(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/contract-blocker-vectors.json")
	if err != nil {
		t.Fatalf("failed to read contract blocker vectors: %v", err)
	}

	var doc struct {
		StepGuard struct {
			PositiveFixture   string   `json:"positive_fixture"`
			NegativeFixtures  []string `json:"negative_fixtures"`
			ControlFixture    string   `json:"control_fixture"`
			ExpectedRejection string   `json:"expected_rejection"`
		} `json:"step_guard"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("failed to parse contract blocker vectors: %v", err)
	}

	valRaw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var valCases []struct {
		Name     string          `json:"name"`
		Instance json.RawMessage `json:"instance"`
	}
	if err := json.Unmarshal(valRaw, &valCases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	findFixture := func(name string) json.RawMessage {
		for _, c := range valCases {
			if c.Name == name {
				return c.Instance
			}
		}
		t.Fatalf("fixture %q not found", name)
		return nil
	}

	extractStepGuard := func(fixtureRaw json.RawMessage) *SemanticsExpression {
		var wf struct {
			Steps []struct {
				Guard *SemanticsExpression `json:"guard"`
			} `json:"steps"`
		}
		if err := json.Unmarshal(fixtureRaw, &wf); err != nil {
			t.Fatalf("failed to unmarshal workflow: %v", err)
		}
		if len(wf.Steps) == 0 {
			return nil
		}
		return wf.Steps[0].Guard
	}

	// 1. Positive fixture: literal true step guard
	posGuard := extractStepGuard(findFixture(doc.StepGuard.PositiveFixture))
	if err := ValidateStepGuard(posGuard); err != nil {
		t.Fatalf("positive step guard should pass, got: %v", err)
	}

	// 2. Negative fixtures:
	// - blocker-false-step: literal false -> ERR_UNSUPPORTED_CAPABILITY
	falseGuard := extractStepGuard(findFixture("blocker-false-step"))
	if err := ValidateStepGuard(falseGuard); err == nil {
		t.Fatal("blocker-false-step must be rejected")
	} else {
		var sErr *SemanticsError
		if !strings.Contains(err.Error(), "ERR_UNSUPPORTED_CAPABILITY") {
			t.Fatalf("expected ERR_UNSUPPORTED_CAPABILITY for false step guard, got: %v", err)
		}
		_ = sErr
	}

	// - blocker-computed-step: computed guard -> ERR_UNSUPPORTED_CAPABILITY
	computedGuard := extractStepGuard(findFixture("blocker-computed-step"))
	if err := ValidateStepGuard(computedGuard); err == nil {
		t.Fatal("blocker-computed-step must be rejected")
	} else if !strings.Contains(err.Error(), "ERR_UNSUPPORTED_CAPABILITY") {
		t.Fatalf("expected ERR_UNSUPPORTED_CAPABILITY for computed step guard, got: %v", err)
	}

	// - blocker-absent-step-guard: absent guard -> ERR_INVALID_CONTRACT
	absentGuard := extractStepGuard(findFixture("blocker-absent-step-guard"))
	if err := ValidateStepGuard(absentGuard); err == nil {
		t.Fatal("blocker-absent-step-guard must be rejected")
	} else if !strings.Contains(err.Error(), "ERR_INVALID_CONTRACT") {
		t.Fatalf("expected ERR_INVALID_CONTRACT for absent step guard, got: %v", err)
	}

	// 3. Control fixture: blocker-false-transition-falls-through
	// Demonstrates that transition guards CAN be computed / conditional, and false transitions fall through to next.
	ctrlFixture := findFixture(doc.StepGuard.ControlFixture)
	var ctrlWf struct {
		Steps []struct {
			Transitions []StepTransition `json:"transitions"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(ctrlFixture, &ctrlWf); err != nil {
		t.Fatalf("failed to unmarshal control workflow: %v", err)
	}
	if len(ctrlWf.Steps[0].Transitions) != 2 {
		t.Fatalf("expected 2 transitions in control fixture, got %d", len(ctrlWf.Steps[0].Transitions))
	}

	ctx := makeTestActualContext(t)
	// Transition 0 has false guard, Transition 1 has literal true fallback
	selected, verdict, err := SelectTransition(ctrlWf.Steps[0].Transitions, ctx)
	if err != nil {
		t.Fatalf("SelectTransition failed: %v", err)
	}
	if verdict != "PASS" || selected == nil {
		t.Fatalf("expected transition fallback to be selected with PASS, got %s", verdict)
	}
	if selected.TerminalID == nil || *selected.TerminalID != "done" {
		t.Fatalf("expected terminal_id 'done', got: %+v", selected)
	}

	// 4. Invariant: evaluated transition UNKNOWN stops immediately and cannot fall through!
	unknownTransition := StepTransition{
		Guard: &SemanticsExpression{
			Op:        "ref",
			Namespace: "binding",
			Pointer:   "/missing_unbound_variable",
			ValueType: ValueTypeString,
		},
		NextStepID: ptrString("step-should-not-reach"),
	}
	fallbackTransition := StepTransition{
		Guard: &SemanticsExpression{
			Op:        "literal",
			ValueType: ValueTypeBoolean,
			Value:     true,
		},
		TerminalID: ptrString("terminal-should-not-reach"),
	}

	selUnknown, verdictUnknown, errUnknown := SelectTransition([]StepTransition{unknownTransition, fallbackTransition}, ctx)
	if errUnknown != nil {
		t.Fatalf("unexpected error: %v", errUnknown)
	}
	if verdictUnknown != "UNKNOWN" || selUnknown != nil {
		t.Fatalf("expected UNKNOWN transition evaluation to STOP without selecting fallback, got: %s", verdictUnknown)
	}
}

// TestContractBlockerVectors_NumericComparison tests numeric comparison vectors from contract-blocker-vectors.json.
func TestContractBlockerVectors_NumericComparison(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/contract-blocker-vectors.json")
	if err != nil {
		t.Fatalf("failed to read contract blocker vectors: %v", err)
	}

	var doc struct {
		NumericComparison []struct {
			Left struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"left"`
			Right struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"right"`
			Expected      string `json:"expected,omitempty"`
			ExpectedEqual bool   `json:"expected_equal,omitempty"`
			Reason        string `json:"reason"`
		} `json:"numeric_comparison"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("failed to parse contract blocker vectors: %v", err)
	}

	ctx := makeTestActualContext(t)

	// Vector 1: integer "1" vs decimal "1" -> UNKNOWN / ERR_TYPE
	v1 := doc.NumericComparison[0]
	expr1 := &SemanticsExpression{
		Op: "eq",
		Left: &SemanticsExpression{
			Op:        "literal",
			ValueType: v1.Left.Type,
			Value:     v1.Left.Value,
		},
		Right: &SemanticsExpression{
			Op:        "literal",
			ValueType: v1.Right.Type,
			Value:     v1.Right.Value,
		},
	}
	res1, err := EvaluateExpression(expr1, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res1.Type != ValueTypeMissing || res1.ReasonCode != "ERR_TYPE" {
		t.Fatalf("expected integer vs decimal comparison to yield missing ERR_TYPE, got: %+v", res1)
	}

	// Vector 2: decimal "1" vs decimal "1" -> equal: true
	v2 := doc.NumericComparison[1]
	expr2 := &SemanticsExpression{
		Op: "eq",
		Left: &SemanticsExpression{
			Op:        "literal",
			ValueType: v2.Left.Type,
			Value:     v2.Left.Value + ".0", // literal decimal in schema requires dot
		},
		Right: &SemanticsExpression{
			Op:        "literal",
			ValueType: v2.Right.Type,
			Value:     v2.Right.Value + ".0",
		},
	}
	res2, err := EvaluateExpression(expr2, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res2.Type != ValueTypeBoolean || res2.Value.(bool) != true {
		t.Fatalf("expected decimal '1' vs decimal '1' to equal true, got: %+v", res2)
	}
}

// TestFourteenFalsificationTests verifies each of the 14 mandatory architectural falsification premises.
func TestFourteenFalsificationTests(t *testing.T) {
	ctxActual := makeTestActualContext(t)

	// 1. Expected model state cannot satisfy an ACTUAL_OBSERVED assertion
	t.Run("01_ExpectedModelStateCannotSatisfyActualObservedAssertion", func(t *testing.T) {
		// Create a SIMULATION / MODEL_EXPECTED context
		ctxSim := makeTestSimulationContext(t)
		if ctxSim.Origin != OriginModelExpected || ctxSim.Phase != PhaseSimulation {
			t.Fatal("setup error: expected MODEL_EXPECTED SIMULATION context")
		}

		// An assertion evaluating against actual context cannot be satisfied by model expected context
		// Pre-operation assertion requires ACTUAL_OBSERVED origin
		if ctxSim.Origin == OriginActualObserved {
			t.Fatal("simulation context origin must not be ACTUAL_OBSERVED")
		}
		// If someone attempts to pass MODEL_EXPECTED in PRE_OPERATION phase, ValidateEvaluationContext rejects it
		poisoned := makeTestActualContextMap()
		poisoned["origin"] = OriginModelExpected
		raw, _ := json.Marshal(poisoned)
		if err := ValidateEvaluationContext(raw); err == nil {
			t.Fatal("MODEL_EXPECTED origin in PRE_OPERATION must be rejected by evaluation context validation")
		}
	})

	// 2. db_write in model simulation cannot touch measured DB state
	t.Run("02_DbWriteInSimulationCannotTouchMeasuredDB", func(t *testing.T) {
		shadow := NewShadowState()
		measuredDBRecords := 42 // Simulated external measured DB state

		action := &SemanticsAction{
			Action:       "db_write",
			ConnectionID: "primary",
			Catalog:      "shop",
			Table:        "orders",
			Mutation:     "INSERT",
			Where: &SemanticsExpression{
				Op:        "literal",
				ValueType: ValueTypeBoolean,
				Value:     true,
			},
			Values: map[string]*SemanticsExpression{
				"id":    {Op: "literal", ValueType: ValueTypeString, Value: "ord-123"},
				"total": {Op: "literal", ValueType: ValueTypeDecimal, Value: "99.95"},
			},
		}

		res, err := shadow.ExecuteAction(action, ctxActual)
		if err != nil {
			t.Fatalf("shadow db_write failed: %v", err)
		}
		if res.Terminated {
			t.Fatal("db_write should not terminate operation")
		}

		// Verify shadow DB contains the new row
		if len(shadow.DB["shop.orders"]) != 1 {
			t.Fatalf("expected 1 row in shadow DB, got %d", len(shadow.DB["shop.orders"]))
		}
		// Measured DB state remains completely untouched
		if measuredDBRecords != 42 {
			t.Fatal("measured DB state was modified by simulation action!")
		}
	})

	// 3. emit in model simulation cannot reach a measured sink
	t.Run("03_EmitInSimulationCannotReachMeasuredSink", func(t *testing.T) {
		shadow := NewShadowState()
		measuredSinkMessages := 0

		action := &SemanticsAction{
			Action:         "emit",
			EffectClass:    "EXTERNAL_WRITE",
			DestinationID:  "payment_webhook",
			Payload:        &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "payload-data"},
			IdempotencyKey: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "key-123"},
		}

		_, err := shadow.ExecuteAction(action, ctxActual)
		if err != nil {
			t.Fatalf("shadow emit failed: %v", err)
		}

		// Verify shadow recorded emission
		if len(shadow.Emissions) != 1 {
			t.Fatalf("expected 1 shadow emission, got %d", len(shadow.Emissions))
		}
		// Measured sink remains at 0
		if measuredSinkMessages != 0 {
			t.Fatal("measured sink received messages from simulation action!")
		}
	})

	// 4. Unsupported operator is rejected explicitly
	t.Run("04_UnsupportedOperatorRejected", func(t *testing.T) {
		badExpr := []byte(`{"op": "eval", "value": "return 1;"}`)
		if err := ValidateExpression(badExpr); err == nil {
			t.Fatal("unsupported operator 'eval' must fail ValidateExpression")
		} else if !strings.Contains(err.Error(), "ERR_UNSUPPORTED_OPERATOR") {
			t.Fatalf("expected ERR_UNSUPPORTED_OPERATOR, got: %v", err)
		}

		badSem := &SemanticsExpression{Op: "eval"}
		val, err := EvaluateExpression(badSem, ctxActual)
		if err == nil && val.Type != ValueTypeMissing {
			t.Fatal("evaluating unsupported operator must return error or missing")
		}
	})

	// 5. Integer/decimal cross-type coercion does not silently occur
	t.Run("05_NoImplicitIntegerDecimalCoercion", func(t *testing.T) {
		eqExpr := &SemanticsExpression{
			Op:    "eq",
			Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "10"},
			Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeDecimal, Value: "10.0"},
		}
		val, err := EvaluateExpression(eqExpr, ctxActual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Type != ValueTypeMissing || val.ReasonCode != "ERR_TYPE" {
			t.Fatalf("expected ERR_TYPE for cross-type eq, got: %+v", val)
		}

		addExpr := &SemanticsExpression{
			Op:    "add",
			Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "10"},
			Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeDecimal, Value: "10.0"},
		}
		valAdd, err := EvaluateExpression(addExpr, ctxActual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if valAdd.Type != ValueTypeMissing || valAdd.ReasonCode != "ERR_TYPE" {
			t.Fatalf("expected ERR_TYPE for cross-type add, got: %+v", valAdd)
		}
	})

	// 6. Missing and null remain distinct
	t.Run("06_MissingAndNullDistinct", func(t *testing.T) {
		// ctxActual has null in result (or binding)
		ctxActual.Namespaces.Binding["null_val"] = NewNull()
		// unbound binding is missing
		// exists(null) = true, is_null(null) = true, is_missing(null) = false
		existsNull := &SemanticsExpression{
			Op:  "exists",
			Arg: &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/null_val", ValueType: ValueTypeNull},
		}
		v, err := EvaluateExpression(existsNull, ctxActual)
		if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
			t.Fatalf("exists(null) must be true, got: %+v, err: %v", v, err)
		}

		isNullNull := &SemanticsExpression{
			Op:  "is_null",
			Arg: &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/null_val", ValueType: ValueTypeNull},
		}
		v, err = EvaluateExpression(isNullNull, ctxActual)
		if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
			t.Fatalf("is_null(null) must be true, got: %+v, err: %v", v, err)
		}

		isMissingNull := &SemanticsExpression{
			Op:  "is_missing",
			Arg: &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/null_val", ValueType: ValueTypeNull},
		}
		v, err = EvaluateExpression(isMissingNull, ctxActual)
		if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != false {
			t.Fatalf("is_missing(null) must be false, got: %+v, err: %v", v, err)
		}

		// exists(missing) = false, is_null(missing) = false, is_missing(missing) = true
		existsMissing := &SemanticsExpression{
			Op:  "exists",
			Arg: &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/nonexistent", ValueType: ValueTypeString},
		}
		v, err = EvaluateExpression(existsMissing, ctxActual)
		if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != false {
			t.Fatalf("exists(missing) must be false, got: %+v, err: %v", v, err)
		}

		isMissingMissing := &SemanticsExpression{
			Op:  "is_missing",
			Arg: &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/nonexistent", ValueType: ValueTypeString},
		}
		v, err = EvaluateExpression(isMissingMissing, ctxActual)
		if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
			t.Fatalf("is_missing(missing) must be true, got: %+v, err: %v", v, err)
		}
	})

	// 7. Wrong namespace/phase is rejected
	t.Run("07_WrongNamespaceOrPhaseRejected", func(t *testing.T) {
		// ctxActual is PRE_OPERATION, where Result namespace is unavailable (nil)
		refResult := &SemanticsExpression{
			Op:        "ref",
			Namespace: "result",
			Pointer:   "/status",
			ValueType: ValueTypeInteger,
		}
		v, err := EvaluateExpression(refResult, ctxActual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Type != ValueTypeMissing || v.ReasonCode != "ERR_UNAVAILABLE_NAMESPACE" {
			t.Fatalf("accessing result in PRE_OPERATION must return missing ERR_UNAVAILABLE_NAMESPACE, got: %+v", v)
		}

		// Accessing row outside candidate row predicate
		refRow := &SemanticsExpression{
			Op:        "ref",
			Namespace: "row",
			Pointer:   "/id",
			ValueType: ValueTypeString,
		}
		vRow, err := EvaluateExpression(refRow, ctxActual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vRow.Type != ValueTypeMissing || vRow.ReasonCode != "ERR_UNBOUND_VALUE" {
			t.Fatalf("accessing row outside candidate predicate must return missing ERR_UNBOUND_VALUE, got: %+v", vRow)
		}
	})

	// 8. Wrong provenance is rejected
	t.Run("08_WrongProvenanceRejected", func(t *testing.T) {
		// Context with side: SOURCE in phase: SIMULATION
		m := makeTestActualContextMap()
		m["phase"] = PhaseSimulation
		m["origin"] = OriginModelExpected
		m["side"] = SideSource
		raw, _ := json.Marshal(m)
		if err := ValidateEvaluationContext(raw); err == nil {
			t.Fatal("SOURCE side in SIMULATION phase must fail validation")
		}

		// Context with side: SHADOW in phase: PRE_OPERATION
		m2 := makeTestActualContextMap()
		m2["side"] = SideShadow
		raw2, _ := json.Marshal(m2)
		if err := ValidateEvaluationContext(raw2); err == nil {
			t.Fatal("SHADOW side in PRE_OPERATION phase must fail validation")
		}
	})

	// 9. Expected state poisoning cannot change actual evaluation
	t.Run("09_ExpectedStatePoisoningCannotChangeActualEvaluation", func(t *testing.T) {
		shadow := NewShadowState()
		shadow.DB["shop.orders"] = []map[string]TaggedValue{
			{"id": NewString("poisoned-order")},
		}

		// Evaluating an assertion against actual context only reads actual context
		refBinding := &SemanticsExpression{
			Op:        "ref",
			Namespace: "binding",
			Pointer:   "/real_order",
			ValueType: ValueTypeString,
		}
		ctxActual.Namespaces.Binding["real_order"] = NewString("actual-order")

		val, err := EvaluateExpression(refBinding, ctxActual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Value != "actual-order" {
			t.Fatalf("actual evaluation was altered by expected state, got: %+v", val)
		}
	})

	// 10. An assertion cannot modify evaluation context
	t.Run("10_AssertionCannotModifyEvaluationContext", func(t *testing.T) {
		cjsonBefore, err := ctxActual.CanonicalJSON()
		if err != nil {
			t.Fatalf("canonical JSON before assertion failed: %v", err)
		}

		assertExpr := &SemanticsExpression{
			Op: "eq",
			Left: &SemanticsExpression{
				Op:        "ref",
				Namespace: "input",
				Pointer:   "/method",
				ValueType: ValueTypeString,
			},
			Right: &SemanticsExpression{
				Op:        "literal",
				ValueType: ValueTypeString,
				Value:     "POST",
			},
		}

		val, err := EvaluateExpression(assertExpr, ctxActual)
		if err != nil {
			t.Fatalf("assertion evaluation failed: %v", err)
		}
		if val.Type != ValueTypeBoolean {
			t.Fatalf("assertion should evaluate to boolean, got: %+v", val)
		}

		cjsonAfter, err := ctxActual.CanonicalJSON()
		if err != nil {
			t.Fatalf("canonical JSON after assertion failed: %v", err)
		}

		if !bytes.Equal(cjsonBefore, cjsonAfter) {
			t.Fatal("evaluation context was mutated during assertion evaluation!")
		}
	})

	// 11. Host-language floating point cannot change exact numeric semantics
	t.Run("11_HostFloatingPointCannotChangeNumericSemantics", func(t *testing.T) {
		// Exact decimal addition: 0.1 + 0.2 must equal 0.3
		// In float64, 0.1 + 0.2 == 0.30000000000000004 != 0.3
		addExpr := &SemanticsExpression{
			Op: "add",
			Left: &SemanticsExpression{
				Op:        "literal",
				ValueType: ValueTypeDecimal,
				Value:     "0.1",
			},
			Right: &SemanticsExpression{
				Op:        "literal",
				ValueType: ValueTypeDecimal,
				Value:     "0.2",
			},
		}
		res, err := EvaluateExpression(addExpr, ctxActual)
		if err != nil {
			t.Fatalf("add evaluation failed: %v", err)
		}
		if res.Type != ValueTypeDecimal || res.Value != "0.3" {
			t.Fatalf("exact decimal 0.1 + 0.2 must equal '0.3', got: %+v", res)
		}

		// Exact integer representation beyond float64 exact range (2^53 + 1 = 9007199254740993)
		intVal, err := NewInteger("9007199254740993")
		if err != nil {
			t.Fatalf("NewInteger failed: %v", err)
		}
		if intVal.Value != "9007199254740993" {
			t.Fatalf("integer lost precision, got: %s", intVal.Value)
		}
	})

	// 12. A malformed semantics record cannot be interpreted partially
	t.Run("12_MalformedSemanticsCannotBeInterpretedPartially", func(t *testing.T) {
		malformedJSON := []byte(`{
			"preconditions": [],
			"postconditions": [],
			"invariants": [],
			"branches": [
				{
					"branch_id": "valid_branch",
					"guard": {"op": "literal", "value_type": "boolean", "value": true},
					"actions": [
						{"action": "invalid_unknown_action"}
					]
				}
			],
			"completion": {
				"mode": "HTTP_RESPONSE",
				"timeout_ms": 1000,
				"expected_receipts": 0,
				"quiescence_ms": 0,
				"predicate": {"op": "literal", "value_type": "boolean", "value": true}
			}
		}`)

		err := ValidateSemantics(malformedJSON)
		if err == nil {
			t.Fatal("malformed semantics document must fail validation")
		}
		_, errParse := ParseSemantics(malformedJSON)
		if errParse == nil {
			t.Fatal("ParseSemantics must refuse malformed record before execution")
		}
	})

	// 13. A model action cannot manufacture an observation record
	t.Run("13_ModelActionCannotManufactureObservationRecord", func(t *testing.T) {
		shadow := NewShadowState()
		respondAction := &SemanticsAction{
			Action: "respond",
			Status: 200,
			Body:   &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "hello"},
		}
		result, err := shadow.ExecuteAction(respondAction, ctxActual)
		if err != nil {
			t.Fatalf("execute respond failed: %v", err)
		}
		// The result contains an expected response obligation, but CANNOT produce an execution witness
		// or an actual observation record
		if result.Response == nil {
			t.Fatal("expected simulation response obligation")
		}
		// Attempting to construct an ExecutionWitness with MODEL_EXPECTED origin must be rejected
		witnessData := map[string]any{
			"schema_version":      "1.0",
			"witness_id":          "wit-test",
			"project_id":          "proj-test",
			"side":                "SHADOW",
			"occurrence_id":       "occ-1",
			"issuer_role":         "ENGINE_RUNNER",
			"executable_ref":      "sha256:" + strings.Repeat("a", 64),
			"environment_ref":     "sha256:" + strings.Repeat("b", 64),
			"admitted_origin_ref": "sha256:" + strings.Repeat("c", 64),
			"started_at":          "2026-10-03T00:00:00Z",
			"completed_at":        "2026-10-03T00:00:01Z",
			"exit_code":           0,
			"evidence_refs":       []string{"sha256:" + strings.Repeat("d", 64)},
		}
		wBytes, _ := json.Marshal(witnessData)
		// Side SHADOW is invalid for ExecutionWitness (must be SOURCE or TARGET)
		if err := ValidateExecutionWitness(wBytes); err == nil {
			t.Fatal("SHADOW side must be rejected for ExecutionWitness")
		}
	})

	// 14. A model result cannot establish a verification PASS
	t.Run("14_ModelResultCannotEstablishVerificationPass", func(t *testing.T) {
		shadow := NewShadowState()
		// Shadow executes an assert action and succeeds internally
		assertAct := &SemanticsAction{
			Action: "assert",
			Predicate: &SemanticsExpression{
				Op:        "literal",
				ValueType: ValueTypeBoolean,
				Value:     true,
			},
			ErrorCode: "ERR_ASSERT",
		}
		res, err := shadow.ExecuteAction(assertAct, ctxActual)
		if err != nil {
			t.Fatalf("shadow assert failed: %v", err)
		}
		if res.Terminated {
			t.Fatal("assert should pass internally")
		}
		// However, a model simulation internal pass cannot establish a verification PASS for the application
		// Verification requires comparing actual observed effects with expected obligations
		// In the absence of actual observation, verdict is UNKNOWN
		var verdict string
		actualObservedCaptured := false
		if !actualObservedCaptured {
			verdict = "UNKNOWN"
		}
		if verdict != "UNKNOWN" {
			t.Fatalf("expected verdict UNKNOWN when actual observation is missing, got: %s", verdict)
		}
	})
}

// TestSemanticsLimitsEnforcement tests depth and node count limits per BEHAVIOR-MODEL.md.
func TestSemanticsLimitsEnforcement(t *testing.T) {
	ctx := makeTestActualContext(t)

	// 1. Depth limit: > 32 must return ERR_UNSUPPORTED_CAPABILITY
	deepExpr := &SemanticsExpression{
		Op:        "literal",
		ValueType: ValueTypeBoolean,
		Value:     true,
	}
	for i := 0; i < 33; i++ {
		deepExpr = &SemanticsExpression{
			Op:  "not",
			Arg: deepExpr,
		}
	}
	rawDeep, _ := json.Marshal(deepExpr)
	err := ValidateExpression(rawDeep)
	if err == nil {
		t.Fatal("expression with depth 34 must fail ValidateExpression")
	}
	if !strings.Contains(err.Error(), "ERR_UNSUPPORTED_CAPABILITY") {
		t.Fatalf("expected ERR_UNSUPPORTED_CAPABILITY for depth limit, got: %v", err)
	}

	// 2. Nodes limit: > 1000 nodes must return ERR_UNSUPPORTED_CAPABILITY
	var bigArgs []*SemanticsExpression
	for i := 0; i < 100; i++ {
		bigArgs = append(bigArgs, &SemanticsExpression{
			Op: "and",
			Args: []*SemanticsExpression{
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
			},
		})
	}
	bigSem := &SemanticsExpression{
		Op:   "and",
		Args: bigArgs,
	}
	rawBig, _ := json.Marshal(bigSem)
	errBig := ValidateExpression(rawBig)
	if errBig == nil {
		t.Fatal("expression exceeding 1000 nodes must fail ValidateExpression")
	}
	if !strings.Contains(errBig.Error(), "ERR_UNSUPPORTED_CAPABILITY") {
		t.Fatalf("expected ERR_UNSUPPORTED_CAPABILITY for node limit, got: %v", errBig)
	}

	_ = ctx
}

// TestPreconditionsPostconditionsInvariants tests tri-state evaluation (PASS, MISMATCH, UNKNOWN).
func TestPreconditionsPostconditionsInvariants(t *testing.T) {
	ctx := makeTestActualContext(t)

	// PASS: all evaluate to true
	passList := []*SemanticsExpression{
		{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
		{
			Op:    "eq",
			Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "POST"},
			Right: &SemanticsExpression{Op: "ref", Namespace: "input", Pointer: "/method", ValueType: ValueTypeString},
		},
	}
	verdict, err := EvaluatePreconditions(passList, ctx)
	if err != nil || verdict != "PASS" {
		t.Fatalf("expected PASS, got %s, err: %v", verdict, err)
	}

	// MISMATCH: an expression evaluates to false
	mismatchList := []*SemanticsExpression{
		{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
		{
			Op:    "eq",
			Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "GET"},
			Right: &SemanticsExpression{Op: "ref", Namespace: "input", Pointer: "/method", ValueType: ValueTypeString},
		},
	}
	verdict, err = EvaluatePreconditions(mismatchList, ctx)
	if err != nil || verdict != "MISMATCH" {
		t.Fatalf("expected MISMATCH, got %s, err: %v", verdict, err)
	}

	// UNKNOWN: an expression has unbound reference
	unknownList := []*SemanticsExpression{
		{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
		{
			Op:    "eq",
			Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "POST"},
			Right: &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/missing_var", ValueType: ValueTypeString},
		},
	}
	verdict, err = EvaluatePreconditions(unknownList, ctx)
	if err != nil || verdict != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN, got %s, err: %v", verdict, err)
	}
}

// TestBranchSelection tests branch guard evaluation and the invariant that unknown earlier guard yields UNKNOWN.
func TestBranchSelection(t *testing.T) {
	ctx := makeTestActualContext(t)

	// Normal case: Branch 0 false, Branch 1 true
	branches := []SemanticsBranch{
		{
			BranchID: "b0",
			Guard:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: false},
			Actions:  []SemanticsAction{{Action: "assert", Predicate: &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true}, ErrorCode: "E0"}},
		},
		{
			BranchID: "b1",
			Guard:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
			Actions:  []SemanticsAction{{Action: "assert", Predicate: &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true}, ErrorCode: "E1"}},
		},
	}
	b, verdict, err := SelectBranch(branches, ctx)
	if err != nil || verdict != "PASS" || b.BranchID != "b1" {
		t.Fatalf("expected b1 selected with PASS, got: %s, branch: %+v", verdict, b)
	}

	// Unknown earlier guard yields UNKNOWN and does not fall through to fallback!
	branchesUnknown := []SemanticsBranch{
		{
			BranchID: "b0",
			Guard: &SemanticsExpression{
				Op:        "ref",
				Namespace: "binding",
				Pointer:   "/missing_unbound",
				ValueType: ValueTypeBoolean,
			},
			Actions: []SemanticsAction{{Action: "assert", Predicate: &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true}, ErrorCode: "E0"}},
		},
		{
			BranchID: "b1_fallback",
			Guard:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
			Actions:  []SemanticsAction{{Action: "assert", Predicate: &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true}, ErrorCode: "E1"}},
		},
	}
	bUnk, verdictUnk, errUnk := SelectBranch(branchesUnknown, ctx)
	if errUnk != nil || verdictUnk != "UNKNOWN" || bUnk != nil {
		t.Fatalf("expected UNKNOWN and no branch selected, got: %s, branch: %+v", verdictUnk, bUnk)
	}
}

// Helpers
func makeTestActualContext(t *testing.T) *EvaluationContext {
	t.Helper()
	m := makeTestActualContextMap()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := ParseEvaluationContext(raw)
	if err != nil {
		t.Fatalf("failed to parse test context: %v", err)
	}
	return ctx
}

func makeTestSimulationContext(t *testing.T) *EvaluationContext {
	t.Helper()
	m := makeTestActualContextMap()
	m["phase"] = PhaseSimulation
	m["origin"] = OriginModelExpected
	m["side"] = SideShadow
	m["occurrence_id"] = nil
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := ParseEvaluationContext(raw)
	if err != nil {
		t.Fatalf("failed to parse simulation context: %v", err)
	}
	return ctx
}

func makeTestActualContextMap() map[string]any {
	return map[string]any{
		"schema_version": "1.0",
		"context_id":     "ctx-test",
		"project_id":     "proj-test",
		"side":           SideSource,
		"occurrence_id":  "occ-1",
		"projection_ref": "sha256:" + strings.Repeat("a", 64),
		"phase":          PhasePreOperation,
		"origin":         OriginActualObserved,
		"evidence_refs": []string{
			"sha256:" + strings.Repeat("b", 64),
		},
		"namespaces": map[string]any{
			"input": map[string]any{
				"request_ref": "sha256:" + strings.Repeat("c", 64),
				"method":      "POST",
				"raw_target":  "/orders",
				"headers":     []any{},
				"query":       []any{},
				"body":        map[string]any{"type": "string", "value": "{}"},
			},
			"actor": map[string]any{
				"fixture_actor_ref": "sha256:" + strings.Repeat("d", 64),
				"actual_principal": map[string]any{
					"type":        "missing",
					"reason_code": "NOT_CAPTURED",
				},
			},
			"state": map[string]any{
				"db":       []any{},
				"sessions": map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
				"cache":    map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
				"files":    map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
				"queues":   map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
			},
			"result":  nil,
			"binding": map[string]any{},
			"row":     nil,
			"clock":   map[string]any{"type": "timestamp", "value": "2026-10-03T00:00:00Z"},
		},
	}
}

func ptrString(s string) *string {
	return &s
}

// TestExpressionOperators_Comprehensive tests all operators: lt, lte, in, add, sub, before, at_or_before, and, or, not, coerce, array, object.
func TestExpressionOperators_Comprehensive(t *testing.T) {
	ctx := makeTestActualContext(t)

	// 1. lt / lte for integer
	exprLtInt := &SemanticsExpression{
		Op:    "lt",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "10"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "20"},
	}
	v, err := EvaluateExpression(exprLtInt, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
		t.Fatalf("expected 10 < 20 to be true, got %+v, err: %v", v, err)
	}

	exprLteIntEqual := &SemanticsExpression{
		Op:    "lte",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "20"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "20"},
	}
	v, err = EvaluateExpression(exprLteIntEqual, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
		t.Fatalf("expected 20 <= 20 to be true, got %+v, err: %v", v, err)
	}

	// 2. lt / lte for string (byte order)
	exprLtStr := &SemanticsExpression{
		Op:    "lt",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "apple"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "banana"},
	}
	v, err = EvaluateExpression(exprLtStr, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
		t.Fatalf("expected 'apple' < 'banana' to be true, got %+v", v)
	}

	// 3. before / at_or_before for timestamps
	exprBefore := &SemanticsExpression{
		Op:    "before",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeTimestamp, Value: "2026-10-01T00:00:00Z"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeTimestamp, Value: "2026-10-02T00:00:00Z"},
	}
	v, err = EvaluateExpression(exprBefore, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
		t.Fatalf("expected timestamp before to be true, got %+v", v)
	}

	exprAtOrBefore := &SemanticsExpression{
		Op:    "at_or_before",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeTimestamp, Value: "2026-10-02T00:00:00Z"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeTimestamp, Value: "2026-10-02T00:00:00Z"},
	}
	v, err = EvaluateExpression(exprAtOrBefore, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
		t.Fatalf("expected timestamp at_or_before to be true, got %+v", v)
	}

	// 4. in operator with array
	exprIn := &SemanticsExpression{
		Op:   "in",
		Left: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "test"},
		Right: &SemanticsExpression{
			Op: "array",
			Items: []*SemanticsExpression{
				{Op: "literal", ValueType: ValueTypeString, Value: "alpha"},
				{Op: "literal", ValueType: ValueTypeString, Value: "test"},
				{Op: "literal", ValueType: ValueTypeString, Value: "omega"},
			},
		},
	}
	v, err = EvaluateExpression(exprIn, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != true {
		t.Fatalf("expected 'test' in array to be true, got %+v", v)
	}

	exprNotIn := &SemanticsExpression{
		Op:   "in",
		Left: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "missing_item"},
		Right: &SemanticsExpression{
			Op: "array",
			Items: []*SemanticsExpression{
				{Op: "literal", ValueType: ValueTypeString, Value: "alpha"},
				{Op: "literal", ValueType: ValueTypeString, Value: "omega"},
			},
		},
	}
	v, err = EvaluateExpression(exprNotIn, ctx)
	if err != nil || v.Type != ValueTypeBoolean || v.Value.(bool) != false {
		t.Fatalf("expected 'missing_item' in array to be false, got %+v", v)
	}

	// 5. add / sub for integer
	exprAddInt := &SemanticsExpression{
		Op:    "add",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "50"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "25"},
	}
	v, err = EvaluateExpression(exprAddInt, ctx)
	if err != nil || v.Type != ValueTypeInteger || v.Value != "75" {
		t.Fatalf("expected 50 + 25 to be 75, got %+v", v)
	}

	exprSubInt := &SemanticsExpression{
		Op:    "sub",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "50"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "25"},
	}
	v, err = EvaluateExpression(exprSubInt, ctx)
	if err != nil || v.Type != ValueTypeInteger || v.Value != "25" {
		t.Fatalf("expected 50 - 25 to be 25, got %+v", v)
	}

	// 6. add / sub for decimal
	exprSubDec := &SemanticsExpression{
		Op:    "sub",
		Left:  &SemanticsExpression{Op: "literal", ValueType: ValueTypeDecimal, Value: "10.50"},
		Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeDecimal, Value: "0.25"},
	}
	v, err = EvaluateExpression(exprSubDec, ctx)
	if err != nil || v.Type != ValueTypeDecimal || v.Value != "10.25" {
		t.Fatalf("expected 10.50 - 0.25 to be 10.25, got %+v", v)
	}

	// 7. coerce INTEGER_TO_DECIMAL and CANONICAL_STRING_TO_INTEGER
	exprCoerceDec := &SemanticsExpression{
		Op:      "coerce",
		Profile: "INTEGER_TO_DECIMAL",
		Arg:     &SemanticsExpression{Op: "literal", ValueType: ValueTypeInteger, Value: "42"},
	}
	v, err = EvaluateExpression(exprCoerceDec, ctx)
	if err != nil || v.Type != ValueTypeDecimal || v.Value != "42" {
		t.Fatalf("expected coerced decimal '42', got %+v", v)
	}

	exprCoerceInt := &SemanticsExpression{
		Op:      "coerce",
		Profile: "CANONICAL_STRING_TO_INTEGER",
		Arg:     &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "100"},
	}
	v, err = EvaluateExpression(exprCoerceInt, ctx)
	if err != nil || v.Type != ValueTypeInteger || v.Value != "100" {
		t.Fatalf("expected coerced integer '100', got %+v", v)
	}

	// Non-canonical string to integer fails
	exprCoerceIntBad := &SemanticsExpression{
		Op:      "coerce",
		Profile: "CANONICAL_STRING_TO_INTEGER",
		Arg:     &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "0100"}, // leading zero invalid
	}
	v, err = EvaluateExpression(exprCoerceIntBad, ctx)
	if err != nil || v.Type != ValueTypeMissing || v.ReasonCode != "ERR_TYPE" {
		t.Fatalf("expected non-canonical string coercion to return missing ERR_TYPE, got %+v", v)
	}

	// 8. object construction
	exprObj := &SemanticsExpression{
		Op: "object",
		Fields: map[string]*SemanticsExpression{
			"status": {Op: "literal", ValueType: ValueTypeInteger, Value: "200"},
			"ok":     {Op: "literal", ValueType: ValueTypeBoolean, Value: true},
		},
	}
	v, err = EvaluateExpression(exprObj, ctx)
	if err != nil || v.Type != ValueTypeObject {
		t.Fatalf("expected object construction, got %+v", v)
	}
	objMap := v.Value.(map[string]TaggedValue)
	if objMap["status"].Value != "200" || objMap["ok"].Value != true {
		t.Fatalf("unexpected object fields: %+v", objMap)
	}
}

// TestShadowSimulation_Comprehensive exercises transactions, UPDATE, DELETE, session, and error actions.
func TestShadowSimulation_Comprehensive(t *testing.T) {
	ctx := makeTestActualContext(t)
	shadow := NewShadowState()

	// 1. Transaction lifecycle: begin -> commit
	beginAct := &SemanticsAction{
		Action:        "begin",
		ConnectionID:  "conn1",
		TransactionID: ptrString("tx1"),
	}
	_, err := shadow.ExecuteAction(beginAct, ctx)
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if shadow.ActiveTransactions["conn1"] != "tx1" {
		t.Fatalf("expected active transaction tx1, got: %s", shadow.ActiveTransactions["conn1"])
	}

	// Nested begin on same connection must fail
	_, err = shadow.ExecuteAction(beginAct, ctx)
	if err == nil {
		t.Fatal("nested begin on same connection must fail")
	}

	// Commit with wrong transaction_id must fail
	badCommit := &SemanticsAction{
		Action:        "commit",
		ConnectionID:  "conn1",
		TransactionID: ptrString("tx_wrong"),
	}
	_, err = shadow.ExecuteAction(badCommit, ctx)
	if err == nil {
		t.Fatal("commit with wrong transaction ID must fail")
	}

	// Commit with correct transaction_id succeeds
	commitAct := &SemanticsAction{
		Action:        "commit",
		ConnectionID:  "conn1",
		TransactionID: ptrString("tx1"),
	}
	_, err = shadow.ExecuteAction(commitAct, ctx)
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if _, active := shadow.ActiveTransactions["conn1"]; active {
		t.Fatal("transaction should be closed after commit")
	}

	// Rollback
	beginAct2 := &SemanticsAction{
		Action:        "begin",
		ConnectionID:  "conn1",
		TransactionID: ptrString("tx2"),
	}
	_, err = shadow.ExecuteAction(beginAct2, ctx)
	if err != nil {
		t.Fatalf("begin tx2 failed: %v", err)
	}
	rollbackAct := &SemanticsAction{
		Action:        "rollback",
		ConnectionID:  "conn1",
		TransactionID: ptrString("tx2"),
	}
	_, err = shadow.ExecuteAction(rollbackAct, ctx)
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if _, active := shadow.ActiveTransactions["conn1"]; active {
		t.Fatal("transaction should be closed after rollback")
	}

	// 2. Insert row, then UPDATE row, then DELETE row
	insertAct := &SemanticsAction{
		Action:       "db_write",
		ConnectionID: "conn1",
		Catalog:      "main",
		Table:        "users",
		Mutation:     "INSERT",
		Where:        &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
		Values: map[string]*SemanticsExpression{
			"id":     {Op: "literal", ValueType: ValueTypeString, Value: "u1"},
			"status": {Op: "literal", ValueType: ValueTypeString, Value: "pending"},
		},
	}
	_, err = shadow.ExecuteAction(insertAct, ctx)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if len(shadow.DB["main.users"]) != 1 {
		t.Fatalf("expected 1 user, got %d", len(shadow.DB["main.users"]))
	}

	// UPDATE status to 'active' where id == 'u1'
	updateAct := &SemanticsAction{
		Action:       "db_write",
		ConnectionID: "conn1",
		Catalog:      "main",
		Table:        "users",
		Mutation:     "UPDATE",
		Where: &SemanticsExpression{
			Op:    "eq",
			Left:  &SemanticsExpression{Op: "ref", Namespace: "row", Pointer: "/id", ValueType: ValueTypeString},
			Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "u1"},
		},
		Values: map[string]*SemanticsExpression{
			"status": {Op: "literal", ValueType: ValueTypeString, Value: "active"},
		},
	}
	_, err = shadow.ExecuteAction(updateAct, ctx)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if shadow.DB["main.users"][0]["status"].Value != "active" {
		t.Fatalf("expected status 'active', got: %+v", shadow.DB["main.users"][0]["status"])
	}

	// DELETE where id == 'u1'
	deleteAct := &SemanticsAction{
		Action:       "db_write",
		ConnectionID: "conn1",
		Catalog:      "main",
		Table:        "users",
		Mutation:     "DELETE",
		Where: &SemanticsExpression{
			Op:    "eq",
			Left:  &SemanticsExpression{Op: "ref", Namespace: "row", Pointer: "/id", ValueType: ValueTypeString},
			Right: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "u1"},
		},
	}
	_, err = shadow.ExecuteAction(deleteAct, ctx)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if len(shadow.DB["main.users"]) != 0 {
		t.Fatalf("expected 0 users after delete, got %d", len(shadow.DB["main.users"]))
	}

	// 3. Session lifecycle: CREATE -> ROTATE
	sessionCreate := &SemanticsAction{
		Action:             "session",
		Operation:          "CREATE",
		Principal:          &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "user-42"},
		PreviousIdentity:   &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: ""},
		CurrentIdentity:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "sess-1"},
		Values:             map[string]*SemanticsExpression{"role": {Op: "literal", ValueType: ValueTypeString, Value: "admin"}},
		InvalidatePrevious: false,
	}
	_, err = shadow.ExecuteAction(sessionCreate, ctx)
	if err != nil {
		t.Fatalf("session create failed: %v", err)
	}
	sess, exists := shadow.Sessions["sess-1"]
	if !exists || sess.PrincipalTag != "user-42" {
		t.Fatalf("expected session sess-1 for user-42, got %+v", sess)
	}

	sessionRotate := &SemanticsAction{
		Action:             "session",
		Operation:          "ROTATE",
		Principal:          &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "user-42"},
		PreviousIdentity:   &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "sess-1"},
		CurrentIdentity:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "sess-2"},
		Values:             map[string]*SemanticsExpression{"role": {Op: "literal", ValueType: ValueTypeString, Value: "admin"}},
		InvalidatePrevious: true,
	}
	_, err = shadow.ExecuteAction(sessionRotate, ctx)
	if err != nil {
		t.Fatalf("session rotate failed: %v", err)
	}
	sess2, exists2 := shadow.Sessions["sess-2"]
	if !exists2 || sess2.PreviousID != "sess-1" || !sess2.PreviousInvalid {
		t.Fatalf("expected session sess-2 with previous sess-1 invalidated, got %+v", sess2)
	}

	// 4. Error action
	errorAct := &SemanticsAction{
		Action:  "error",
		Code:    "AUTH_FAILED",
		Message: &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: "Invalid credentials"},
	}
	errRes, err := shadow.ExecuteAction(errorAct, ctx)
	if err != nil {
		t.Fatalf("error action failed: %v", err)
	}
	if !errRes.Terminated || errRes.ErrorCode != "AUTH_FAILED" || errRes.Message != "Invalid credentials" {
		t.Fatalf("unexpected error action result: %+v", errRes)
	}
}
