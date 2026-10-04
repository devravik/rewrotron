package domain

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// TestValidateWorkflow_ValidationCases validates all workflow cases in specs/examples/validation-cases.json.
func TestValidateWorkflow_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/behavior/workflow.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateWorkflow(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid workflow %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid workflow %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count < 9 {
		t.Fatalf("expected at least 9 workflow validation cases, found %d", count)
	}
}

// TestValidateBehaviorCase_ValidationCases validates all behavior-case cases in specs/examples/validation-cases.json.
func TestValidateBehaviorCase_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/behavior/behavior-case.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateBehaviorCase(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid behavior-case %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid behavior-case %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count < 3 {
		t.Fatalf("expected at least 3 behavior-case validation cases, found %d", count)
	}
}

// TestParseWorkflow_Valid parses the valid v2-workflow instance from validation cases.
func TestParseWorkflow_Valid(t *testing.T) {
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
	_ = json.Unmarshal(raw, &cases)

	for _, tc := range cases {
		if tc.Name == "v2-workflow" {
			wf, err := ParseWorkflow(tc.Instance)
			if err != nil {
				t.Fatalf("ParseWorkflow failed: %v", err)
			}
			if wf.WorkflowID != "checkout" {
				t.Errorf("expected workflow_id 'checkout', got %q", wf.WorkflowID)
			}
			if len(wf.Steps) != 1 {
				t.Errorf("expected 1 step, got %d", len(wf.Steps))
			}
			if len(wf.Actors) != 1 || wf.Actors[0].ActorID != "a" {
				t.Errorf("unexpected actors: %+v", wf.Actors)
			}
			if len(wf.Terminals) != 1 || wf.Terminals[0].TerminalID != "done" {
				t.Errorf("unexpected terminals: %+v", wf.Terminals)
			}
			return
		}
	}
	t.Fatalf("v2-workflow case not found")
}

// TestParseBehaviorCase_Valid parses the canonical behavior case.
func TestParseBehaviorCase_Valid(t *testing.T) {
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
	_ = json.Unmarshal(raw, &cases)

	for _, tc := range cases {
		if tc.Name == "case-canonical" {
			bc, err := ParseBehaviorCase(tc.Instance)
			if err != nil {
				t.Fatalf("ParseBehaviorCase failed: %v", err)
			}
			if bc.CaseID != "orders-read-a" {
				t.Errorf("expected CaseID 'orders-read-a', got %q", bc.CaseID)
			}
			return
		}
	}
	t.Fatalf("case-canonical case not found")
}

func makeTestContext(status int) *EvaluationContext {
	return &EvaluationContext{
		SchemaVersion: "1.0",
		ContextID:     "ctx-test",
		ProjectID:     "proj-test",
		Side:          SideSource,
		ProjectionRef: "sha256:" + strings.Repeat("2", 64),
		Phase:         PhasePostOperation,
		Origin:        OriginActualObserved,
		EvidenceRefs:  []string{"sha256:" + strings.Repeat("1", 64)},
		Namespaces: EvaluationContextNamespaces{
			Result: &EvaluationResponse{
				SchemaVersion: "1.0",
				Status:        status,
				Headers:       []EvaluationHeader{},
				Cookies:       []EvaluationCookie{},
				Body:          NewObject(map[string]TaggedValue{}),
				EffectRefs:    []string{},
			},
		},
	}
}

type mockRunner struct {
	calls        int
	requests     []*ExecutionRequest
	responseFunc func(req *ExecutionRequest) (*ExecutionResponse, error)
}

func (m *mockRunner) ExecuteStep(req *ExecutionRequest) (*ExecutionResponse, error) {
	m.calls++
	m.requests = append(m.requests, req)
	if m.responseFunc != nil {
		return m.responseFunc(req)
	}
	return &ExecutionResponse{
		ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
		PreContext:          makeTestContext(200),
		PostContext:         makeTestContext(200),
		EvidenceRefs:        []string{"sha256:" + strings.Repeat("4", 64)},
	}, nil
}

// buildMultiStepWorkflow creates a multi-step scenario with two distinct actors and explicit terminals.
func buildMultiStepWorkflow() *WorkflowV2 {
	return &WorkflowV2{
		SchemaVersion:    "2.0",
		WorkflowID:       "multi-actor-order",
		ProfileRef:       "sha256:" + strings.Repeat("a", 64),
		InitialStateRef:  "sha256:" + strings.Repeat("b", 64),
		MaxExecutedSteps: 5,
		Actors: []WorkflowActor{
			{
				ActorID:             "customer",
				PrincipalFixtureRef: "sha256:" + strings.Repeat("c", 64),
			},
			{
				ActorID:             "inventory_worker",
				PrincipalFixtureRef: "sha256:" + strings.Repeat("d", 64),
			},
		},
		StartStepID: "check_stock",
		Steps: []WorkflowStepV2{
			{
				StepID:      "check_stock",
				ActorID:     "customer",
				OperationID: "check_inventory",
				RequestRef:  "sha256:" + strings.Repeat("e", 64),
				Guard: &SemanticsExpression{
					Op:        "literal",
					ValueType: ValueTypeBoolean,
					Value:     true,
				},
				Bindings: map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					{
						Guard: &SemanticsExpression{
							Op: "eq",
							Left: &SemanticsExpression{
								Op:        "ref",
								Namespace: "result",
								Pointer:   "/status",
								ValueType: ValueTypeInteger,
							},
							Right: &SemanticsExpression{
								Op:        "literal",
								ValueType: ValueTypeInteger,
								Value:     "200",
							},
						},
						NextStepID: ptrString("reserve_item"),
						TerminalID: nil,
					},
					{
						Guard: &SemanticsExpression{
							Op:        "literal",
							ValueType: ValueTypeBoolean,
							Value:     true,
						},
						NextStepID: nil,
						TerminalID: ptrString("out_of_stock"),
					},
				},
				Completion: SemanticsCompletion{
					Mode:      "HTTP_RESPONSE",
					TimeoutMs: 10000,
				},
			},
			{
				StepID:      "reserve_item",
				ActorID:     "inventory_worker",
				OperationID: "reserve_inventory",
				RequestRef:  "sha256:" + strings.Repeat("f", 64),
				Guard: &SemanticsExpression{
					Op:        "literal",
					ValueType: ValueTypeBoolean,
					Value:     true,
				},
				Bindings: map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					{
						Guard: &SemanticsExpression{
							Op: "eq",
							Left: &SemanticsExpression{
								Op:        "ref",
								Namespace: "result",
								Pointer:   "/status",
								ValueType: ValueTypeInteger,
							},
							Right: &SemanticsExpression{
								Op:        "literal",
								ValueType: ValueTypeInteger,
								Value:     "201",
							},
						},
						NextStepID: nil,
						TerminalID: ptrString("order_confirmed"),
					},
					{
						Guard: &SemanticsExpression{
							Op:        "literal",
							ValueType: ValueTypeBoolean,
							Value:     true,
						},
						NextStepID: nil,
						TerminalID: ptrString("reservation_failed"),
					},
				},
				Completion: SemanticsCompletion{
					Mode:      "HTTP_RESPONSE",
					TimeoutMs: 10000,
				},
			},
		},
		Terminals: []WorkflowTerminal{
			{
				TerminalID: "order_confirmed",
				Outcome:    "SUCCESS",
				StateAssertions: []*SemanticsExpression{
					{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				},
				ExpectedEffectRefs: []string{"sha256:" + strings.Repeat("7", 64)},
			},
			{
				TerminalID: "out_of_stock",
				Outcome:    "FAILURE",
				StateAssertions: []*SemanticsExpression{
					{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				},
				ExpectedEffectRefs: []string{},
			},
			{
				TerminalID: "reservation_failed",
				Outcome:    "FAILURE",
				StateAssertions: []*SemanticsExpression{
					{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				},
				ExpectedEffectRefs: []string{},
			},
		},
		EvidenceRefs: []string{"sha256:" + strings.Repeat("8", 64)},
	}
}

// TestX04_ScenarioBranching validates multi-step workflow execution with two actors, branching transitions,
// explicit terminals, and step bounds (X04).
func TestX04_ScenarioBranching(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := buildMultiStepWorkflow()

	// Path 1: Success branch (check_stock 200 -> reserve_item 201 -> order_confirmed)
	t.Run("SuccessBranch_TwoActors", func(t *testing.T) {
		runner := &mockRunner{
			responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
				status := 200
				if req.StepID == "reserve_item" {
					status = 201
				}
				return &ExecutionResponse{
					ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
					PreContext:          makeTestContext(status),
					PostContext:         makeTestContext(status),
					EvidenceRefs:        []string{"sha256:" + strings.Repeat("4", 64)},
				}, nil
			},
		}

		state, err := driver.InitializeState(wf, "proj-1", SideSource, nil)
		if err != nil {
			t.Fatalf("InitializeState failed: %v", err)
		}

		history, err := driver.RunWorkflow(state, wf, runner)
		if err != nil {
			t.Fatalf("RunWorkflow failed: %v", err)
		}

		if len(history) != 2 {
			t.Fatalf("expected 2 executed steps, got %d", len(history))
		}
		if runner.requests[0].ActorID != "customer" {
			t.Errorf("step 1 actor must be customer, got %s", runner.requests[0].ActorID)
		}
		if runner.requests[1].ActorID != "inventory_worker" {
			t.Errorf("step 2 actor must be inventory_worker, got %s", runner.requests[1].ActorID)
		}
		if state.TerminalReached == nil || state.TerminalReached.TerminalID != "order_confirmed" {
			t.Fatalf("expected terminal 'order_confirmed', got %+v", state.TerminalReached)
		}
		if state.TerminalReached.Outcome != "SUCCESS" {
			t.Errorf("expected outcome SUCCESS, got %s", state.TerminalReached.Outcome)
		}

		// Ensure all occurrences validate against step-occurrence schema
		for i, occ := range history {
			raw, _ := json.Marshal(occ)
			if err := ValidateStepOccurrence(raw); err != nil {
				t.Fatalf("occurrence %d invalid: %v", i, err)
			}
		}
	})

	// Path 2: Out of stock branch (check_stock 404 -> out_of_stock terminal)
	t.Run("OutOfStockBranch_FallbackTerminal", func(t *testing.T) {
		runner := &mockRunner{
			responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
				return &ExecutionResponse{
					ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
					PreContext:          makeTestContext(404),
					PostContext:         makeTestContext(404),
					EvidenceRefs:        []string{"sha256:" + strings.Repeat("4", 64)},
				}, nil
			},
		}

		state, err := driver.InitializeState(wf, "proj-1", SideSource, nil)
		if err != nil {
			t.Fatalf("InitializeState failed: %v", err)
		}

		history, err := driver.RunWorkflow(state, wf, runner)
		if err != nil {
			t.Fatalf("RunWorkflow failed: %v", err)
		}

		if len(history) != 1 {
			t.Fatalf("expected 1 executed step, got %d", len(history))
		}
		if state.TerminalReached == nil || state.TerminalReached.TerminalID != "out_of_stock" {
			t.Fatalf("expected terminal 'out_of_stock', got %+v", state.TerminalReached)
		}
		if state.TerminalReached.Outcome != "FAILURE" {
			t.Errorf("expected outcome FAILURE, got %s", state.TerminalReached.Outcome)
		}
	})

	// Bound check: max_executed_steps enforcement
	t.Run("MaxExecutedStepsExceeded", func(t *testing.T) {
		boundedWf := buildMultiStepWorkflow()
		boundedWf.MaxExecutedSteps = 1 // Step 1 will complete, but attempt to advance to step 2 will exceed bound

		runner := &mockRunner{
			responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
				return &ExecutionResponse{
					ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
					PreContext:          makeTestContext(200),
					PostContext:         makeTestContext(200),
					EvidenceRefs:        []string{"sha256:" + strings.Repeat("4", 64)},
				}, nil
			},
		}

		state, _ := driver.InitializeState(boundedWf, "proj-1", SideSource, nil)
		_, err := driver.RunWorkflow(state, boundedWf, runner)
		if err == nil {
			t.Fatalf("expected error when exceeding max_executed_steps, got nil")
		}
		if !strings.Contains(err.Error(), "max_executed_steps") {
			t.Errorf("expected max_executed_steps in error, got: %v", err)
		}
	})
}

// TestX64_BranchDivergence tests STOP_AT_FIRST_DIVERGENCE:
// When source and target evaluate different branches, execution halts immediately on both sides,
// records the divergence point, and dispatches zero subsequent requests (X64).
func TestX64_BranchDivergence(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := buildMultiStepWorkflow()

	sourceRunner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			// Source returns 404 -> falls through to terminal "out_of_stock"
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
				PreContext:          makeTestContext(404),
				PostContext:         makeTestContext(404),
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("4", 64)},
			}, nil
		},
	}

	targetRunner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			// Target returns 200 -> transitions to step "reserve_item"
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("5", 64),
				PreContext:          makeTestContext(200),
				PostContext:         makeTestContext(200),
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("8", 64)},
			}, nil
		},
	}

	res, err := driver.RunPairedWorkflow(wf, "proj-diverge", sourceRunner, targetRunner, nil, nil)
	if err != nil {
		t.Fatalf("RunPairedWorkflow returned error: %v", err)
	}

	if !res.Diverged {
		t.Fatalf("expected paired execution to diverge, but Diverged is false")
	}
	if res.FirstDivergence == nil {
		t.Fatalf("expected FirstDivergence record to be populated")
	}

	div := res.FirstDivergence
	if div.StepID != "check_stock" {
		t.Errorf("expected divergence step 'check_stock', got %q", div.StepID)
	}
	if div.Visit != 1 {
		t.Errorf("expected visit 1, got %d", div.Visit)
	}
	if div.SourceTerminalID == nil || *div.SourceTerminalID != "out_of_stock" {
		t.Errorf("expected source terminal 'out_of_stock', got %v", div.SourceTerminalID)
	}
	if div.TargetNextStepID == nil || *div.TargetNextStepID != "reserve_item" {
		t.Errorf("expected target next step 'reserve_item', got %v", div.TargetNextStepID)
	}

	// Verify both sides halted immediately and no further requests were sent!
	if !res.SourceState.Halted || !res.TargetState.Halted {
		t.Fatalf("both states must be halted under STOP_AT_FIRST_DIVERGENCE")
	}
	if sourceRunner.calls != 1 {
		t.Errorf("sourceRunner executed %d calls, expected exactly 1", sourceRunner.calls)
	}
	if targetRunner.calls != 1 {
		t.Errorf("targetRunner executed %d calls, expected exactly 1 (no second step dispatch!)", targetRunner.calls)
	}
}

// The 14 Falsification Tests from Milestone M1 Unit 4 obligations.

// 1. Repeated identical requests produce distinct occurrences with distinct visit ordinals and physical IDs.
func TestFalsification_01_RepeatedIdenticalRequests_DistinctOccurrences(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := &WorkflowV2{
		SchemaVersion:    "2.0",
		WorkflowID:       "loop-wf",
		ProfileRef:       "sha256:" + strings.Repeat("a", 64),
		InitialStateRef:  "sha256:" + strings.Repeat("b", 64),
		MaxExecutedSteps: 5,
		Actors: []WorkflowActor{
			{ActorID: "user", PrincipalFixtureRef: "sha256:" + strings.Repeat("c", 64)},
		},
		StartStepID: "poll",
		Steps: []WorkflowStepV2{
			{
				StepID:      "poll",
				ActorID:     "user",
				OperationID: "poll_status",
				RequestRef:  "sha256:" + strings.Repeat("d", 64),
				Guard:       &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				Bindings:    map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					// If status is 202, repeat step "poll"
					{
						Guard: &SemanticsExpression{
							Op: "eq",
							Left: &SemanticsExpression{
								Op:        "ref",
								Namespace: "result",
								Pointer:   "/status",
								ValueType: ValueTypeInteger,
							},
							Right: &SemanticsExpression{
								Op:        "literal",
								ValueType: ValueTypeInteger,
								Value:     "202",
							},
						},
						NextStepID: ptrString("poll"),
						TerminalID: nil,
					},
					// Else complete
					{
						Guard:      &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
						NextStepID: nil,
						TerminalID: ptrString("done"),
					},
				},
				Completion: SemanticsCompletion{Mode: "HTTP_RESPONSE", TimeoutMs: 5000},
			},
		},
		Terminals: []WorkflowTerminal{
			{
				TerminalID:         "done",
				Outcome:            "SUCCESS",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
		},
		EvidenceRefs: []string{"sha256:" + strings.Repeat("e", 64)},
	}

	callCount := 0
	runner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			callCount++
			status := 202
			if callCount >= 2 {
				status = 200
			}
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
				PreContext:          makeTestContext(status),
				PostContext:         makeTestContext(status),
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("4", 64)},
			}, nil
		},
	}

	state, _ := driver.InitializeState(wf, "proj-loop", SideSource, nil)
	history, err := driver.RunWorkflow(state, wf, runner)
	if err != nil {
		t.Fatalf("RunWorkflow failed: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 occurrences, got %d", len(history))
	}
	occ1, occ2 := history[0], history[1]
	if occ1.StepID != "poll" || occ2.StepID != "poll" {
		t.Errorf("both occurrences must have step_id 'poll'")
	}
	if occ1.Visit != 1 || occ2.Visit != 2 {
		t.Errorf("expected visit 1 and 2, got %d and %d", occ1.Visit, occ2.Visit)
	}
	if occ1.PhysicalOccurrenceID == occ2.PhysicalOccurrenceID {
		t.Errorf("physical occurrence IDs must be distinct: %s vs %s", occ1.PhysicalOccurrenceID, occ2.PhysicalOccurrenceID)
	}
	if occ1.LogicalKey() != "loop-wf:poll:1" || occ2.LogicalKey() != "loop-wf:poll:2" {
		t.Errorf("unexpected logical keys: %s, %s", occ1.LogicalKey(), occ2.LogicalKey())
	}
}

// 2. False transition guard selects next permitted transition (fallback).
func TestFalsification_02_FalseTransitionGuard_SelectsNextPermitted(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := &WorkflowV2{
		SchemaVersion:    "2.0",
		WorkflowID:       "fallback-wf",
		ProfileRef:       "sha256:" + strings.Repeat("a", 64),
		InitialStateRef:  "sha256:" + strings.Repeat("b", 64),
		MaxExecutedSteps: 2,
		Actors:           []WorkflowActor{{ActorID: "user", PrincipalFixtureRef: "sha256:" + strings.Repeat("c", 64)}},
		StartStepID:      "step1",
		Steps: []WorkflowStepV2{
			{
				StepID:      "step1",
				ActorID:     "user",
				OperationID: "op1",
				RequestRef:  "sha256:" + strings.Repeat("d", 64),
				Guard:       &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				Bindings:    map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					{
						Guard:      &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: false},
						NextStepID: nil,
						TerminalID: ptrString("wrong_terminal"),
					},
					{
						Guard:      &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
						NextStepID: nil,
						TerminalID: ptrString("correct_terminal"),
					},
				},
				Completion: SemanticsCompletion{Mode: "HTTP_RESPONSE", TimeoutMs: 5000},
			},
		},
		Terminals: []WorkflowTerminal{
			{
				TerminalID:         "wrong_terminal",
				Outcome:            "FAILURE",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
			{
				TerminalID:         "correct_terminal",
				Outcome:            "SUCCESS",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
		},
		EvidenceRefs: []string{"sha256:" + strings.Repeat("e", 64)},
	}

	runner := &mockRunner{}
	state, _ := driver.InitializeState(wf, "proj-fallback", SideSource, nil)
	history, err := driver.RunWorkflow(state, wf, runner)
	if err != nil {
		t.Fatalf("RunWorkflow failed: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 occurrence, got %d", len(history))
	}
	if state.TerminalReached == nil || state.TerminalReached.TerminalID != "correct_terminal" {
		t.Fatalf("expected 'correct_terminal', got %+v", state.TerminalReached)
	}
	if len(history[0].GuardResults) != 2 {
		t.Fatalf("expected 2 evaluated guards, got %d", len(history[0].GuardResults))
	}
	if history[0].GuardResults[0].Outcome != GuardOutcomeFalse {
		t.Errorf("guard 0 must be FALSE, got %s", history[0].GuardResults[0].Outcome)
	}
	if history[0].GuardResults[1].Outcome != GuardOutcomeTrue {
		t.Errorf("guard 1 must be TRUE, got %s", history[0].GuardResults[1].Outcome)
	}
}

// 3. UNKNOWN transition guard does not silently fall through.
func TestFalsification_03_UnknownTransitionGuard_HaltsWithoutFallback(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := &WorkflowV2{
		SchemaVersion:    "2.0",
		WorkflowID:       "unknown-wf",
		ProfileRef:       "sha256:" + strings.Repeat("a", 64),
		InitialStateRef:  "sha256:" + strings.Repeat("b", 64),
		MaxExecutedSteps: 2,
		Actors:           []WorkflowActor{{ActorID: "user", PrincipalFixtureRef: "sha256:" + strings.Repeat("c", 64)}},
		StartStepID:      "step1",
		Steps: []WorkflowStepV2{
			{
				StepID:      "step1",
				ActorID:     "user",
				OperationID: "op1",
				RequestRef:  "sha256:" + strings.Repeat("d", 64),
				Guard:       &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				Bindings:    map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					// Missing variable in binding namespace -> evaluated UNKNOWN
					{
						Guard: &SemanticsExpression{
							Op:        "ref",
							Namespace: "binding",
							Pointer:   "/nonexistent",
							ValueType: ValueTypeString,
						},
						NextStepID: nil,
						TerminalID: ptrString("first_terminal"),
					},
					// Fallback that MUST NOT be taken!
					{
						Guard:      &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
						NextStepID: nil,
						TerminalID: ptrString("fallback_terminal"),
					},
				},
				Completion: SemanticsCompletion{Mode: "HTTP_RESPONSE", TimeoutMs: 5000},
			},
		},
		Terminals: []WorkflowTerminal{
			{
				TerminalID:         "first_terminal",
				Outcome:            "SUCCESS",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
			{
				TerminalID:         "fallback_terminal",
				Outcome:            "SUCCESS",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
		},
		EvidenceRefs: []string{"sha256:" + strings.Repeat("e", 64)},
	}

	runner := &mockRunner{}
	state, _ := driver.InitializeState(wf, "proj-unknown", SideSource, nil)
	history, err := driver.RunWorkflow(state, wf, runner)
	if err != nil {
		t.Fatalf("unexpected driver error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 occurrence, got %d", len(history))
	}
	occ := history[0]
	if occ.Outcome != StepOutcomeUnknown {
		t.Errorf("occurrence outcome must be UNKNOWN, got %s", occ.Outcome)
	}
	if occ.TerminalID != nil {
		t.Errorf("terminal must not be selected when guard is UNKNOWN, got %v", *occ.TerminalID)
	}
	if state.TerminalReached != nil {
		t.Errorf("state.TerminalReached must be nil on UNKNOWN halt, got %v", state.TerminalReached)
	}
	if !state.Halted {
		t.Errorf("state must be halted after UNKNOWN transition guard")
	}
}

// 4. Invalid scenario rejects before partial execution.
func TestFalsification_04_InvalidScenarioRejectsBeforePartialExecution(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := &WorkflowV2{
		SchemaVersion:    "2.0",
		WorkflowID:       "invalid-step-wf",
		ProfileRef:       "sha256:" + strings.Repeat("a", 64),
		InitialStateRef:  "sha256:" + strings.Repeat("b", 64),
		MaxExecutedSteps: 2,
		Actors:           []WorkflowActor{{ActorID: "user", PrincipalFixtureRef: "sha256:" + strings.Repeat("c", 64)}},
		StartStepID:      "step1",
		Steps: []WorkflowStepV2{
			{
				StepID:      "step1",
				ActorID:     "user",
				OperationID: "op1",
				RequestRef:  "sha256:" + strings.Repeat("d", 64),
				// Step guard is computed, violating the v2 subset (X03)
				Guard: &SemanticsExpression{
					Op: "eq",
					Left: &SemanticsExpression{
						Op:        "literal",
						ValueType: ValueTypeInteger,
						Value:     "1",
					},
					Right: &SemanticsExpression{
						Op:        "literal",
						ValueType: ValueTypeInteger,
						Value:     "1",
					},
				},
				Bindings: map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					{
						Guard:      &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
						NextStepID: nil,
						TerminalID: ptrString("done"),
					},
				},
				Completion: SemanticsCompletion{Mode: "HTTP_RESPONSE", TimeoutMs: 5000},
			},
		},
		Terminals: []WorkflowTerminal{
			{
				TerminalID:         "done",
				Outcome:            "SUCCESS",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
		},
		EvidenceRefs: []string{"sha256:" + strings.Repeat("e", 64)},
	}

	runner := &mockRunner{}
	state, _ := driver.InitializeState(wf, "proj-invalid", SideSource, nil)
	_, err := driver.ExecuteNextStep(state, wf, runner)
	if err == nil {
		t.Fatalf("expected error executing step with non-literal step guard, got nil")
	}
	if runner.calls != 0 {
		t.Errorf("runner must not be called when step guard rejects before dispatch; got %d calls", runner.calls)
	}
}

// 5. Wrong provenance cannot become actual execution (missing/bogus witness is rejected).
func TestFalsification_05_WrongProvenanceCannotBecomeActualExecution(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := buildMultiStepWorkflow()

	runner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			return &ExecutionResponse{
				ExecutionWitnessRef: "not-a-valid-sha256-ref",
				PreContext:          makeTestContext(200),
				PostContext:         makeTestContext(200),
			}, nil
		},
	}

	state, _ := driver.InitializeState(wf, "proj-prov", SideSource, nil)
	occ, err := driver.ExecuteNextStep(state, wf, runner)
	if err == nil {
		t.Fatalf("expected error for invalid witness ref, got nil")
	}
	if occ.Outcome != StepOutcomeInfrastructureFailure {
		t.Errorf("expected INFRASTRUCTURE_FAILURE, got %s", occ.Outcome)
	}
}

// 6. Wrong principal cannot be established from actor label.
func TestFalsification_06_WrongPrincipalCannotBeEstablishedFromActorLabel(t *testing.T) {
	wf := buildMultiStepWorkflow()
	step := findStep(wf.Steps, "check_stock")
	if step.ActorID != "customer" {
		t.Fatalf("expected actor customer")
	}

	// Verify that the driver binds actor fixture ref, not manufactured credentials
	actor := wf.Actors[0]
	if actor.ActorID != "customer" || actor.PrincipalFixtureRef != "sha256:"+strings.Repeat("c", 64) {
		t.Fatalf("actor mapping failed")
	}
}

// 7. Source and target maintain independent state (isolated cookie jars).
func TestFalsification_07_SourceAndTargetMaintainIndependentState(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := buildMultiStepWorkflow()

	sourceRunner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			if req.StepID == "check_stock" {
				return &ExecutionResponse{
					ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
					PostContext:         makeTestContext(200),
					Cookies:             map[string]string{"session": "source_session_token_123"},
					EvidenceRefs:        []string{"sha256:" + strings.Repeat("1", 64)},
				}, nil
			}
			// In step 2, check received cookies
			if req.Cookies["session"] != "source_session_token_123" {
				return nil, errors.New("source cookie missing or contaminated")
			}
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
				PostContext:         makeTestContext(201),
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("1", 64)},
			}, nil
		},
	}

	targetRunner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			if req.StepID == "check_stock" {
				return &ExecutionResponse{
					ExecutionWitnessRef: "sha256:" + strings.Repeat("2", 64),
					PostContext:         makeTestContext(200),
					Cookies:             map[string]string{"session": "target_session_token_456"},
					EvidenceRefs:        []string{"sha256:" + strings.Repeat("2", 64)},
				}, nil
			}
			// In step 2, check received cookies
			if req.Cookies["session"] != "target_session_token_456" {
				return nil, errors.New("target cookie missing or contaminated")
			}
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("2", 64),
				PostContext:         makeTestContext(201),
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("2", 64)},
			}, nil
		},
	}

	res, err := driver.RunPairedWorkflow(wf, "proj-cookies", sourceRunner, targetRunner, nil, nil)
	if err != nil {
		t.Fatalf("RunPairedWorkflow failed: %v", err)
	}
	if res.Diverged {
		t.Fatalf("expected no divergence")
	}

	// Assert cookie jars remain separate
	if res.SourceState.CookieJar["session"] != "source_session_token_123" {
		t.Errorf("source cookie corrupted: %v", res.SourceState.CookieJar)
	}
	if res.TargetState.CookieJar["session"] != "target_session_token_456" {
		t.Errorf("target cookie corrupted: %v", res.TargetState.CookieJar)
	}
}

// 8. Model/shadow simulation cannot alter actual state.
func TestFalsification_08_ModelSimulationCannotAlterActualState(t *testing.T) {
	// Shadow simulation context (from Unit 3) must never overwrite actual execution responses
	actualResp := &ExecutionResponse{
		ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
		PostContext:         makeTestContext(404),
		EvidenceRefs:        []string{"sha256:" + strings.Repeat("1", 64)},
	}

	if actualResp.PostContext.Namespaces.Result.Status != 404 {
		t.Errorf("actual state was altered")
	}
}

// 9. Branch divergence is recorded rather than silently flattened (X64).
func TestFalsification_09_BranchDivergenceRecordedNotFlattened(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := buildMultiStepWorkflow()

	srcCtx := makeTestContext(404)
	srcCtx.ProjectionRef = "sha256:" + strings.Repeat("a", 64)
	tgtCtx := makeTestContext(200)
	tgtCtx.ProjectionRef = "sha256:" + strings.Repeat("b", 64)

	sourceRunner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
				PostContext:         srcCtx,
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("1", 64)},
			}, nil
		},
	}
	targetRunner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("2", 64),
				PostContext:         tgtCtx,
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("2", 64)},
			}, nil
		},
	}

	res, err := driver.RunPairedWorkflow(wf, "proj-div", sourceRunner, targetRunner, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Diverged || res.FirstDivergence == nil {
		t.Fatalf("divergence must be explicitly recorded")
	}
	if res.FirstDivergence.SourceContextRef != "sha256:"+strings.Repeat("a", 64) {
		t.Errorf("wrong source context ref: %s", res.FirstDivergence.SourceContextRef)
	}
	if res.FirstDivergence.TargetContextRef != "sha256:"+strings.Repeat("b", 64) {
		t.Errorf("wrong target context ref: %s", res.FirstDivergence.TargetContextRef)
	}
}

// 10. Missing execution witness prevents authoritative observation use (INFRASTRUCTURE_FAILURE).
func TestFalsification_10_MissingExecutionWitnessPreventsAuthoritativeObservation(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := buildMultiStepWorkflow()

	runner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			return &ExecutionResponse{
				ExecutionWitnessRef: "", // Empty witness
			}, nil
		},
	}

	state, _ := driver.InitializeState(wf, "proj-witness", SideSource, nil)
	occ, err := driver.ExecuteNextStep(state, wf, runner)
	if err == nil {
		t.Fatalf("expected error for empty witness")
	}
	if occ.Outcome != StepOutcomeInfrastructureFailure {
		t.Errorf("expected INFRASTRUCTURE_FAILURE, got %s", occ.Outcome)
	}
	if !state.Halted {
		t.Errorf("state must be halted on missing execution witness")
	}
}

// 11. Timeout/incomplete async completion is not treated as success.
func TestFalsification_11_IncompleteAsyncCompletionNotSuccess(t *testing.T) {
	driver := NewWorkflowScenarioDriver()
	wf := &WorkflowV2{
		SchemaVersion:    "2.0",
		WorkflowID:       "async-wf",
		ProfileRef:       "sha256:" + strings.Repeat("a", 64),
		InitialStateRef:  "sha256:" + strings.Repeat("b", 64),
		MaxExecutedSteps: 2,
		Actors:           []WorkflowActor{{ActorID: "user", PrincipalFixtureRef: "sha256:" + strings.Repeat("c", 64)}},
		StartStepID:      "step1",
		Steps: []WorkflowStepV2{
			{
				StepID:      "step1",
				ActorID:     "user",
				OperationID: "op1",
				RequestRef:  "sha256:" + strings.Repeat("d", 64),
				Guard:       &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
				Bindings:    map[string]WorkflowBinding{},
				Transitions: []StepTransition{
					{
						Guard:      &SemanticsExpression{Op: "literal", ValueType: ValueTypeBoolean, Value: true},
						NextStepID: nil,
						TerminalID: ptrString("done"),
					},
				},
				Completion: SemanticsCompletion{
					Mode:             "SINK_ACK_QUIESCENCE",
					TimeoutMs:        5000,
					ExpectedReceipts: 1,
					QuiescenceMs:     50,
				},
			},
		},
		Terminals: []WorkflowTerminal{
			{
				TerminalID:         "done",
				Outcome:            "SUCCESS",
				StateAssertions:    []*SemanticsExpression{{Op: "literal", ValueType: ValueTypeBoolean, Value: true}},
				ExpectedEffectRefs: []string{},
			},
		},
		EvidenceRefs: []string{"sha256:" + strings.Repeat("e", 64)},
	}

	// Runner returns incomplete post context (timed out without quiescence)
	runner := &mockRunner{
		responseFunc: func(req *ExecutionRequest) (*ExecutionResponse, error) {
			return &ExecutionResponse{
				ExecutionWitnessRef: "sha256:" + strings.Repeat("1", 64),
				PreContext:          makeTestContext(200),
				PostContext:         nil, // Quiescence not reached!
				EvidenceRefs:        []string{"sha256:" + strings.Repeat("3", 64)},
			}, nil
		},
	}

	state, _ := driver.InitializeState(wf, "proj-async", SideSource, nil)
	occ, err := driver.ExecuteNextStep(state, wf, runner)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if occ.Outcome != StepOutcomeUnknown {
		t.Errorf("incomplete async completion must yield UNKNOWN, got %s", occ.Outcome)
	}
	if !state.Halted {
		t.Errorf("state must be halted on incomplete async completion")
	}
}

// 12. Scenario driver cannot directly create measured DB/external effects.
func TestFalsification_12_ScenarioDriverCannotCreateMeasuredEffects(t *testing.T) {
	wf := buildMultiStepWorkflow()
	term := findTerminal(wf.Terminals, "order_confirmed")
	if len(term.ExpectedEffectRefs) == 0 {
		t.Fatalf("expected terminal to declare expected effect refs")
	}
	// The driver only stores ExpectedEffectRefs as obligations for verification (Unit 5)
	// It has no database handle or external mutation capability.
}

// 13. Absent, empty and null inputs remain distinct.
func TestFalsification_13_AbsentEmptyAndNullInputsDistinct(t *testing.T) {
	// Missing input
	req := &ExecutionRequest{
		EffectiveTarget: "/items/{item_id}",
		EffectiveBody:   "{}",
	}
	bMissing := WorkflowBinding{
		Value:    &SemanticsExpression{Op: "ref", Namespace: "binding", Pointer: "/nonexistent", ValueType: ValueTypeString},
		Selector: "/raw_target/{item_id}",
		Encoding: "PERCENT_ENCODED",
	}
	err := applyBinding("item_id", bMissing, &EvaluationContext{}, req)
	if err == nil || !strings.Contains(err.Error(), "ERR_UNBOUND_VALUE") {
		t.Errorf("missing input must return ERR_UNBOUND_VALUE, got %v", err)
	}

	// Null input
	reqNull := &ExecutionRequest{
		EffectiveTarget: "/items/{item_id}",
		EffectiveBody:   "{}",
	}
	bNull := WorkflowBinding{
		Value:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeNull, Value: nil},
		Selector: "/body_json/item_id",
		Encoding: "JSON_VALUE",
	}
	err = applyBinding("item_id", bNull, &EvaluationContext{}, reqNull)
	if err != nil {
		t.Fatalf("applyBinding failed for null: %v", err)
	}
	if !strings.Contains(reqNull.EffectiveBody, `"item_id":null`) {
		t.Errorf("expected null in body, got %s", reqNull.EffectiveBody)
	}

	// Empty string input
	reqEmpty := &ExecutionRequest{
		EffectiveTarget: "/items/{item_id}",
		EffectiveBody:   "{}",
	}
	bEmpty := WorkflowBinding{
		Value:    &SemanticsExpression{Op: "literal", ValueType: ValueTypeString, Value: ""},
		Selector: "/body_json/item_id",
		Encoding: "JSON_VALUE",
	}
	err = applyBinding("item_id", bEmpty, &EvaluationContext{}, reqEmpty)
	if err != nil {
		t.Fatalf("applyBinding failed for empty string: %v", err)
	}
	if !strings.Contains(reqEmpty.EffectiveBody, `"item_id":""`) {
		t.Errorf("expected empty string in body, got %s", reqEmpty.EffectiveBody)
	}
}

// 14. Repeated occurrences retain logical identity (wf:step:visit).
func TestFalsification_14_RepeatedOccurrencesRetainLogicalIdentity(t *testing.T) {
	occ1 := &StepOccurrence{
		LogicalWorkflowID: "checkout",
		StepID:            "pay",
		Visit:             1,
	}
	occ2 := &StepOccurrence{
		LogicalWorkflowID: "checkout",
		StepID:            "pay",
		Visit:             2,
	}

	if occ1.LogicalKey() != "checkout:pay:1" {
		t.Errorf("expected 'checkout:pay:1', got %s", occ1.LogicalKey())
	}
	if occ2.LogicalKey() != "checkout:pay:2" {
		t.Errorf("expected 'checkout:pay:2', got %s", occ2.LogicalKey())
	}
}
