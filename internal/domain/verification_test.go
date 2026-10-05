package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestValidateVerification_ValidationCases validates all 14 verification test cases in validation-cases.json.
func TestValidateVerification_ValidationCases(t *testing.T) {
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

	count := 0
	for _, tc := range cases {
		switch tc.Schema {
		case "https://rewrotron.dev/specs/verification/acceptance-scope.schema.json":
			count++
			t.Run(tc.Name, func(t *testing.T) {
				err := ValidateAcceptanceScope(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid scope %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid scope %q to fail, but it passed", tc.Name)
				}
			})
		case "https://rewrotron.dev/specs/verification/normalization.schema.json":
			count++
			t.Run(tc.Name, func(t *testing.T) {
				err := ValidateNormalizationRule(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid normalization rule %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid normalization rule %q to fail, but it passed", tc.Name)
				}
			})
		case "https://rewrotron.dev/specs/verification/occurrence-pairing.schema.json":
			count++
			t.Run(tc.Name, func(t *testing.T) {
				err := ValidateOccurrencePairing(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid occurrence pairing %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid occurrence pairing %q to fail, but it passed", tc.Name)
				}
			})
		case "https://rewrotron.dev/specs/verification/identity-map.schema.json":
			count++
			t.Run(tc.Name, func(t *testing.T) {
				err := ValidateIdentityMap(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid identity map %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid identity map %q to fail, but it passed", tc.Name)
				}
			})
		case "https://rewrotron.dev/specs/verification/report.schema.json":
			count++
			t.Run(tc.Name, func(t *testing.T) {
				err := ValidateVerificationReport(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid report %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid report %q to fail, but it passed", tc.Name)
				}
			})
		}
	}

	if count != 14 {
		t.Fatalf("expected exactly 14 verification validation cases, found %d", count)
	}
}

// Helper to construct a baseline PASS test input
func buildBaselinePassInput() *CompareCaseInput {
	defaultHex := strings.Repeat("a", 64)
	defaultEvidence := "sha256:" + defaultHex
	targetBuildHex := strings.Repeat("b", 64)

	srcWitness := &ExecutionWitness{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-test",
		Side:                 SideSource,
		AttemptID:            "attempt-src-1",
		TreeRef:              defaultEvidence,
		BuildArtifactRef:     defaultEvidence,
		LoadedArtifactRef:    defaultEvidence,
		Launch: ExecutionLaunch{
			HostBootID:     "boot-src",
			ProcessID:      "proc-src",
			ProcessStartID: "start-src",
		},
		EnvironmentRef:     defaultEvidence,
		ResetReportRef:     defaultEvidence,
		EndpointBindingRef: defaultEvidence,
		Issuer:             WitnessIssuerEngineRunner,
		RoleIsolationRef:   defaultEvidence,
	}

	tgtWitness := &ExecutionWitness{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-test",
		Side:                 SideTarget,
		AttemptID:            "attempt-tgt-1",
		TreeRef:              "sha256:" + targetBuildHex,
		BuildArtifactRef:     "sha256:" + targetBuildHex,
		LoadedArtifactRef:    "sha256:" + targetBuildHex,
		Launch: ExecutionLaunch{
			HostBootID:     "boot-tgt",
			ProcessID:      "proc-tgt",
			ProcessStartID: "start-tgt",
		},
		EnvironmentRef:     defaultEvidence,
		ResetReportRef:     defaultEvidence,
		EndpointBindingRef: defaultEvidence,
		Issuer:             WitnessIssuerEngineRunner,
		RoleIsolationRef:   defaultEvidence,
	}

	srcPostRef := "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	tgtPostRef := "sha256:2222222222222222222222222222222222222222222222222222222222222222"

	srcOcc := StepOccurrence{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-test",
		Side:                 SideSource,
		LogicalWorkflowID:    "wf-test",
		StepID:               "step_create",
		Visit:                1,
		PhysicalOccurrenceID: "occ-src-1",
		ExecutionWitnessRef:  defaultEvidence,
		PreContextRef:        defaultEvidence,
		PostContextRef:       &srcPostRef,
		SelectedNextStepID:   nil,
		TerminalID:           stringPtr("term_success"),
		Outcome:              StepOutcomeCompleted,
		EvidenceRefs:         []string{defaultEvidence},
	}

	tgtOcc := StepOccurrence{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-test",
		Side:                 SideTarget,
		LogicalWorkflowID:    "wf-test",
		StepID:               "step_create",
		Visit:                1,
		PhysicalOccurrenceID: "occ-tgt-1",
		ExecutionWitnessRef:  "sha256:" + targetBuildHex,
		PreContextRef:        defaultEvidence,
		PostContextRef:       &tgtPostRef,
		SelectedNextStepID:   nil,
		TerminalID:           stringPtr("term_success"),
		Outcome:              StepOutcomeCompleted,
		EvidenceRefs:         []string{defaultEvidence},
	}

	// Database state: 1 inserted user row
	insertedRow := TaggedValue{
		Type: ValueTypeArray,
		Value: []TaggedValue{
			{
				Type: ValueTypeObject,
				Value: map[string]TaggedValue{
					"id":   NewString("user-1"),
					"name": NewString("Alice"),
				},
			},
		},
	}

	srcCtx := &EvaluationContext{
		SchemaVersion: "1.0",
		ContextID:     "ctx-src-1",
		ProjectID:     "proj-test",
		Side:          SideSource,
		Phase:         PhasePostOperation,
		Origin:        OriginActualObserved,
		EvidenceRefs:  []string{defaultEvidence},
		Namespaces: EvaluationContextNamespaces{
			Actor: EvaluationActor{
				FixtureActorRef: defaultEvidence,
				ActualPrincipal: ActualPrincipal{
					Authentication:    "AUTHENTICATED",
					SubjectLocalID:    stringPtr("loc-user-alice"),
					CaptureChannelRef: defaultEvidence,
					EvidenceRefs:      []string{defaultEvidence},
				},
			},
			State: EvaluationState{
				DB: DBTablesOrMissing{
					Tables: []EvaluationDBTable{
						{
							ConnectionID: "conn_main",
							Catalog:      "app_db",
							Table:        "users",
							Rows:         insertedRow,
						},
					},
				},
			},
			Result: &EvaluationResponse{
				SchemaVersion: "1.0",
				Status:        201,
				Headers: []EvaluationHeader{
					{Name: "Content-Type", Value: "application/json"},
				},
				Body: TaggedValue{
					Type:  ValueTypeString,
					Value: `{"status":"created","user_id":"user-1"}`,
				},
				Cookies: []EvaluationCookie{
					{
						Name:     "session_token",
						Value:    NewString("token-abc"),
						HTTPOnly: true,
						Secure:   true,
						SameSite: stringPtr("Strict"),
					},
				},
				EffectRefs: []string{defaultEvidence},
			},
		},
	}

	tgtCtx := &EvaluationContext{
		SchemaVersion: "1.0",
		ContextID:     "ctx-tgt-1",
		ProjectID:     "proj-test",
		Side:          SideTarget,
		Phase:         PhasePostOperation,
		Origin:        OriginActualObserved,
		EvidenceRefs:  []string{defaultEvidence},
		Namespaces: EvaluationContextNamespaces{
			Actor: EvaluationActor{
				FixtureActorRef: defaultEvidence,
				ActualPrincipal: ActualPrincipal{
					Authentication:    "AUTHENTICATED",
					SubjectLocalID:    stringPtr("loc-tgt-user-alice"),
					CaptureChannelRef: defaultEvidence,
					EvidenceRefs:      []string{defaultEvidence},
				},
			},
			State: EvaluationState{
				DB: DBTablesOrMissing{
					Tables: []EvaluationDBTable{
						{
							ConnectionID: "conn_main",
							Catalog:      "app_db",
							Table:        "users",
							Rows:         insertedRow,
						},
					},
				},
			},
			Result: &EvaluationResponse{
				SchemaVersion: "1.0",
				Status:        201,
				Headers: []EvaluationHeader{
					{Name: "Content-Type", Value: "application/json"},
				},
				Body: TaggedValue{
					Type:  ValueTypeString,
					Value: `{"status":"created","user_id":"user-1"}`,
				},
				Cookies: []EvaluationCookie{
					{
						Name:     "session_token",
						Value:    NewString("token-abc"),
						HTTPOnly: true,
						Secure:   true,
						SameSite: stringPtr("Strict"),
					},
				},
				EffectRefs: []string{defaultEvidence},
			},
		},
	}

	srcLedger := &LocalIdentitiesLedger{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-test",
		Side:                 SideSource,
		ObservationAttemptID: "attempt-src-1",
		LogicalWorkflowID:    "wf-test",
		ExecutionWitnessRef:  defaultEvidence,
		Origin:               OriginActualObserved,
		Entries: []LocalIdentityEntry{
			{
				LocalID:           "loc-user-alice",
				Kind:              IdentityKindPrincipal,
				Domain:            "auth",
				ResourceID:        "user-alice",
				Value:             LocalIdentityValue{Type: "string", Value: "alice"},
				CaptureChannelRef: defaultEvidence,
				EvidenceRefs:      []string{defaultEvidence},
			},
		},
		CaptureEvidenceRefs: []string{defaultEvidence},
	}

	tgtLedger := &LocalIdentitiesLedger{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-test",
		Side:                 SideTarget,
		ObservationAttemptID: "attempt-tgt-1",
		LogicalWorkflowID:    "wf-test",
		ExecutionWitnessRef:  "sha256:" + targetBuildHex,
		Origin:               OriginActualObserved,
		Entries: []LocalIdentityEntry{
			{
				LocalID:           "loc-tgt-user-alice",
				Kind:              IdentityKindPrincipal,
				Domain:            "auth",
				ResourceID:        "user-alice",
				Value:             LocalIdentityValue{Type: "string", Value: "alice"},
				CaptureChannelRef: defaultEvidence,
				EvidenceRefs:      []string{defaultEvidence},
			},
		},
		CaptureEvidenceRefs: []string{defaultEvidence},
	}

	scope := AcceptanceScope{
		Kind:      "UNIT",
		UnitIDs:   []string{"unit-create"},
		CaseIDs:   []string{"case-create"},
		CorpusRef: defaultEvidence,
		ScopeHash: defaultHex,
	}

	return &CompareCaseInput{
		RunID:                   "run-test-1",
		ProjectID:               "proj-test",
		CaseID:                  "case-create",
		ScenarioRef:             defaultEvidence,
		WorkflowID:              "wf-test",
		PolicyRef:               defaultEvidence,
		ConfigurationHash:       defaultHex,
		PlanHash:                defaultHex,
		CorpusHash:              defaultHex,
		PolicyHash:              defaultHex,
		TargetTreeHash:          targetBuildHex,
		ExpectedTargetBuildHash: "sha256:" + targetBuildHex,
		AcceptanceScope:         scope,
		SourceWitness:           srcWitness,
		TargetWitness:           tgtWitness,
		SourceOccurrences:       []StepOccurrence{srcOcc},
		TargetOccurrences:       []StepOccurrence{tgtOcc},
		SourceContexts:          map[string]*EvaluationContext{srcPostRef: srcCtx},
		TargetContexts:          map[string]*EvaluationContext{tgtPostRef: tgtCtx},
		SourceIdentityLedger:    srcLedger,
		TargetIdentityLedger:    tgtLedger,
		NormalizationRules:      nil,
		SourceTraceRef:          defaultEvidence,
		TargetTraceRef:          defaultEvidence,
	}
}

// TestBaseline_Pass validates that the base input produces a complete PASS report conforming to v2.1.
func TestBaseline_Pass(t *testing.T) {
	input := buildBaselinePassInput()
	comp := NewBehaviorComparator()
	out, err := comp.CompareCase(input)
	if err != nil {
		t.Fatalf("unexpected compare error: %v", err)
	}
	if out.CaseVerdict != "PASS" {
		t.Fatalf("expected case verdict PASS, got: %s (errors: %+v)", out.CaseVerdict, out.Errors)
	}
	if out.TopVerdict != "PASS" {
		t.Fatalf("expected top verdict PASS, got: %s", out.TopVerdict)
	}
	if out.Report == nil {
		t.Fatal("expected non-nil report")
	}
	if out.OccurrencePairing == nil {
		t.Fatal("expected non-nil occurrence pairing")
	}
	if out.IdentityMap == nil {
		t.Fatal("expected non-nil identity map")
	}
}

// TestDecisive1_MissingTargetDBWrite validates Decisive Test 1 and Mutant X05 / X61:
// Target returns 201 Created but omits DB INSERT.
// Verifier returns MISMATCH; verifier does NOT insert row into target DB (read-only invariant).
func TestDecisive1_MissingTargetDBWrite(t *testing.T) {
	input := buildBaselinePassInput()

	// Target returns 201 Created but omits the database INSERT
	tgtCtx := input.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"]
	tgtCtx.Namespaces.State.DB.Tables = []EvaluationDBTable{
		{
			ConnectionID: "conn_main",
			Catalog:      "app_db",
			Table:        "users",
			Rows: TaggedValue{
				Type:  ValueTypeArray,
				Value: []TaggedValue{}, // 0 rows inserted!
			},
		},
	}

	comp := NewBehaviorComparator()
	out, err := comp.CompareCase(input)
	if err != nil {
		t.Fatalf("unexpected compare error: %v", err)
	}

	// 1. Verdict must be MISMATCH
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("expected MISMATCH on missing DB write, got: %s", out.CaseVerdict)
	}
	if *out.Report.Cases[0].Dimensions.Database.Verdict != "MISMATCH" {
		t.Fatalf("expected database dimension MISMATCH, got: %v", *out.Report.Cases[0].Dimensions.Database.Verdict)
	}

	// 2. READ-ONLY INVARIANT: verify that comparator did not insert rows into target state
	postCompareRows := extractRows(tgtCtx.Namespaces.State.DB.Tables[0].Rows)
	if len(postCompareRows) != 0 {
		t.Fatalf("CRITICAL: comparator wrote to target DB state! Row count is %d, expected 0", len(postCompareRows))
	}
}

// TestDecisive2_BranchDivergenceAuthDenied validates Decisive Test 2 and Mutant X64:
// Source DENIED vs Target ALLOWED. Verifier records branch divergence at step 1;
// target success branch is not followed; downstream steps marked NOT_COMPARED.
func TestDecisive2_BranchDivergenceAuthDenied(t *testing.T) {
	input := buildBaselinePassInput()

	// Source is DENIED at step 1 (reaches term_denied)
	input.SourceOccurrences[0].TerminalID = stringPtr("term_denied")
	input.SourceOccurrences[0].SelectedNextStepID = nil

	// Target is ALLOWED at step 1 (reaches term_success)
	input.TargetOccurrences[0].TerminalID = stringPtr("term_success")
	input.TargetOccurrences[0].SelectedNextStepID = nil

	comp := NewBehaviorComparator()
	out, err := comp.CompareCase(input)
	if err != nil {
		t.Fatalf("unexpected compare error: %v", err)
	}

	// 1. Verdict must be MISMATCH
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("expected MISMATCH on branch divergence, got: %s", out.CaseVerdict)
	}
	if !out.Diverged {
		t.Fatal("expected Diverged=true")
	}
	if out.FirstDivergence == nil {
		t.Fatal("expected FirstDivergence to be populated")
	}
	if out.FirstDivergence.Reason != "BRANCH" {
		t.Fatalf("expected divergence reason BRANCH, got: %s", out.FirstDivergence.Reason)
	}
	if out.OccurrencePairing.Pairs[0].Outcome != "DIVERGED" {
		t.Fatalf("expected first pair outcome DIVERGED, got: %s", out.OccurrencePairing.Pairs[0].Outcome)
	}
}

// TestEighteenMutants_FalsificationSuite executes the 18 mutants falsification suite.
func TestEighteenMutants_FalsificationSuite(t *testing.T) {
	comp := NewBehaviorComparator()

	// Mutant 1: Missing DB write (Target returns 201 Created but omits INSERT) -> MISMATCH
	t.Run("Mutant1_MissingDBWrite", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.State.DB.Tables[0].Rows = TaggedValue{
			Type:  ValueTypeArray,
			Value: []TaggedValue{},
		}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Database.Verdict != "MISMATCH" {
			t.Fatalf("expected database dimension MISMATCH")
		}
	})

	// Mutant 2 (X06 / X77): Wrong principal (Target authenticates User B instead of User A) -> MISMATCH
	t.Run("Mutant2_WrongPrincipal", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Actor.ActualPrincipal.SubjectLocalID = stringPtr("loc-tgt-user-bob")
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Authentication.Verdict != "MISMATCH" {
			t.Fatalf("expected authentication dimension MISMATCH")
		}
	})

	// Mutant 3 (X07 / X72): Duplicate business effect (Target executes business effect twice) -> MISMATCH
	t.Run("Mutant3_DuplicateBusinessEffect", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.EffectRefs = []string{
			"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.External.Verdict != "MISMATCH" {
			t.Fatalf("expected external dimension MISMATCH")
		}
	})

	// Mutant 4: Missing business effect (Target executes zero business effects) -> MISMATCH
	t.Run("Mutant4_MissingBusinessEffect", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.EffectRefs = []string{}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.External.Verdict != "MISMATCH" {
			t.Fatalf("expected external dimension MISMATCH")
		}
	})

	// Mutant 5 (X64): Branch divergence (Source DENIED vs Target ALLOWED) -> MISMATCH
	t.Run("Mutant5_BranchDivergence", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.SourceOccurrences[0].TerminalID = stringPtr("term_denied")
		in.TargetOccurrences[0].TerminalID = stringPtr("term_allowed")
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if !out.Diverged {
			t.Fatal("expected diverged=true")
		}
	})

	// Mutant 6 (X67): Missing Location header on redirect -> MISMATCH
	t.Run("Mutant6_MissingLocationHeader", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.SourceContexts["sha256:1111111111111111111111111111111111111111111111111111111111111111"].Namespaces.Result.Status = 302
		in.SourceContexts["sha256:1111111111111111111111111111111111111111111111111111111111111111"].Namespaces.Result.Headers = []EvaluationHeader{
			{Name: "Location", Value: "/dashboard"},
		}
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Status = 302
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Headers = []EvaluationHeader{}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.HTTP.Verdict != "MISMATCH" {
			t.Fatalf("expected http dimension MISMATCH")
		}
	})

	// Mutant 7 (X67): Dropped cookie attributes (Target sets cookie without HttpOnly/Secure) -> MISMATCH
	t.Run("Mutant7_DroppedCookieAttributes", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Cookies = []EvaluationCookie{
			{
				Name:     "session_token",
				Value:    NewString("token-abc"),
				HTTPOnly: false, // Dropped!
				Secure:   false, // Dropped!
				SameSite: nil,
			},
		}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Session.Verdict != "MISMATCH" {
			t.Fatalf("expected session dimension MISMATCH")
		}
	})

	// Mutant 8 (X08): Incomplete observer (Database capture missing from target) -> UNKNOWN
	t.Run("Mutant8_IncompleteDatabaseObserver", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.State.DB = DBTablesOrMissing{
			Missing: &MissingValue{Type: "missing", ReasonCode: "NOT_CAPTURED"},
		}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "UNKNOWN" {
			t.Fatalf("expected UNKNOWN, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Database.Verdict != "UNKNOWN" {
			t.Fatalf("expected database dimension UNKNOWN")
		}
	})

	// Mutant 9 (X08): Incomplete observer (Principal capture missing from target) -> UNKNOWN
	t.Run("Mutant9_IncompletePrincipalObserver", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Actor.ActualPrincipal = ActualPrincipal{
			Missing: &MissingValue{Type: "missing", ReasonCode: "NOT_CAPTURED"},
		}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "UNKNOWN" {
			t.Fatalf("expected UNKNOWN, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Authentication.Verdict != "UNKNOWN" {
			t.Fatalf("expected authentication dimension UNKNOWN")
		}
	})

	// Mutant 10 (X80): Source-vs-Source comparison (Two SOURCE witnesses) -> UNKNOWN
	t.Run("Mutant10_SourceVsSource", func(t *testing.T) {
		in := buildBaselinePassInput()
		// Make target witness into SOURCE witness
		in.TargetWitness.Side = SideSource
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "UNKNOWN" {
			t.Fatalf("expected UNKNOWN, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if len(out.Errors) == 0 || out.Errors[0].Code != "ERR_INVALID_COMPARISON" {
			t.Fatalf("expected ERR_INVALID_COMPARISON error, got: %+v", out.Errors)
		}
	})

	// Mutant 11 (X66): Stale target build (Target witness build hash mismatch) -> UNKNOWN
	t.Run("Mutant11_StaleTargetBuild", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.ExpectedTargetBuildHash = "sha256:" + strings.Repeat("e", 64)
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "UNKNOWN" {
			t.Fatalf("expected UNKNOWN, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if len(out.Errors) == 0 || out.Errors[0].Code != "ERR_STALE_EXECUTION" {
			t.Fatalf("expected ERR_STALE_EXECUTION error, got: %+v", out.Errors)
		}
	})

	// Mutant 12: Identity bijection violation (Two source IDs map to same target ID) -> MISMATCH
	t.Run("Mutant12_IdentityBijectionViolation", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.SourceIdentityLedger.Entries = append(in.SourceIdentityLedger.Entries, LocalIdentityEntry{
			LocalID:           "loc-user-alice-dup",
			Kind:              IdentityKindPrincipal,
			Domain:            "auth",
			ResourceID:        "user-alice", // Same resource, mapped to same target
			Value:             LocalIdentityValue{Type: "string", Value: "alice"},
			CaptureChannelRef: "sha256:" + strings.Repeat("a", 64),
			EvidenceRefs:      []string{"sha256:" + strings.Repeat("a", 64)},
		})
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
	})

	// Mutant 13 (X63): Model state poisoning (Context origin is MODEL_EXPECTED instead of ACTUAL_OBSERVED) -> MISMATCH
	t.Run("Mutant13_ModelStatePoisoning", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Origin = OriginModelExpected
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Execution.Verdict != "MISMATCH" {
			t.Fatalf("expected execution dimension MISMATCH")
		}
	})

	// Mutant 14: Status code mismatch (Source 200 vs Target 500) -> MISMATCH
	t.Run("Mutant14_StatusCodeMismatch", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Status = 500
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.HTTP.Verdict != "MISMATCH" {
			t.Fatalf("expected http dimension MISMATCH")
		}
	})

	// Mutant 15: Body content mismatch -> MISMATCH
	t.Run("Mutant15_BodyContentMismatch", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Body = TaggedValue{
			Type:  ValueTypeString,
			Value: `{"status":"internal_server_error"}`,
		}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Body.Verdict != "MISMATCH" {
			t.Fatalf("expected body dimension MISMATCH")
		}
	})

	// Mutant 16: Authentication status mismatch (Source AUTHENTICATED vs Target ANONYMOUS) -> MISMATCH
	t.Run("Mutant16_AuthStatusMismatch", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Actor.ActualPrincipal.Authentication = "ANONYMOUS"
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Authentication.Verdict != "MISMATCH" {
			t.Fatalf("expected authentication dimension MISMATCH")
		}
	})

	// Mutant 17: Cookie count mismatch (Target dropped cookies entirely) -> MISMATCH
	t.Run("Mutant17_CookieCountMismatch", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Cookies = []EvaluationCookie{}
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if *out.Report.Cases[0].Dimensions.Session.Verdict != "MISMATCH" {
			t.Fatalf("expected session dimension MISMATCH")
		}
	})

	// Mutant 18 (X65): Target step occurrence missing -> MISMATCH / DIVERGED
	t.Run("Mutant18_MissingTargetOccurrence", func(t *testing.T) {
		in := buildBaselinePassInput()
		in.TargetOccurrences = []StepOccurrence{} // Missing target occurrence!
		out, err := comp.CompareCase(in)
		if err != nil || out.CaseVerdict != "MISMATCH" {
			t.Fatalf("expected MISMATCH, got err=%v, verdict=%s", err, out.CaseVerdict)
		}
		if !out.Diverged {
			t.Fatal("expected diverged=true when target occurrence missing")
		}
	})
}

// TestReportDeterminism validates that repeated runs of CompareCase produce byte-identical canonical JSON.
func TestReportDeterminism(t *testing.T) {
	comp := NewBehaviorComparator()
	in1 := buildBaselinePassInput()
	in2 := buildBaselinePassInput()

	out1, err1 := comp.CompareCase(in1)
	if err1 != nil {
		t.Fatalf("compare 1 failed: %v", err1)
	}
	out2, err2 := comp.CompareCase(in2)
	if err2 != nil {
		t.Fatalf("compare 2 failed: %v", err2)
	}

	bytes1, err := out1.Report.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize report 1: %v", err)
	}
	bytes2, err := out2.Report.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize report 2: %v", err)
	}

	if string(bytes1) != string(bytes2) {
		t.Fatalf("non-deterministic report output: run 1 and run 2 differ!\nRun 1: %s\nRun 2: %s", string(bytes1), string(bytes2))
	}

	dig1, _ := out1.Report.Digest()
	dig2, _ := out2.Report.Digest()
	if dig1 != dig2 {
		t.Fatalf("non-deterministic report digest: %s != %s", dig1, dig2)
	}
}

// TestNormalizationRules validates dynamic field normalization (e.g. DYNAMIC_DATE, DYNAMIC_UUID).
func TestNormalizationRules(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Source and target have differing dynamic dates and UUIDs
	in.SourceContexts["sha256:1111111111111111111111111111111111111111111111111111111111111111"].Namespaces.Result.Body = TaggedValue{
		Type:  ValueTypeString,
		Value: `{"created_at":"2026-10-05T12:00:00Z","id":"12345678-1234-1234-1234-123456789abc"}`,
	}
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Body = TaggedValue{
		Type:  ValueTypeString,
		Value: `{"created_at":"2026-10-05T12:05:00Z","id":"87654321-4321-4321-4321-cba987654321"}`,
	}

	// Without normalization: body should MISMATCH
	outWithoutNorm, err := comp.CompareCase(in)
	if err != nil || outWithoutNorm.CaseVerdict != "MISMATCH" {
		t.Fatalf("expected MISMATCH without normalization rules, got: %s", outWithoutNorm.CaseVerdict)
	}

	// With normalization rules: body should PASS
	in.NormalizationRules = []*NormalizationRule{
		{
			SchemaVersion: "2.0",
			RuleID:        "norm_date",
			Path:          "/created_at",
			ValueType:     "timestamp",
			Variability:   "CLOCK",
			Relation:      "CONTROLLED_CLOCK",
		},
		{
			SchemaVersion: "2.0",
			RuleID:        "norm_uuid",
			Path:          "/id",
			ValueType:     "string",
			Variability:   "GENERATED_ID",
			Relation:      "BIJECTIVE_ID",
		},
	}

	outWithNorm, err := comp.CompareCase(in)
	if err != nil || outWithNorm.CaseVerdict != "PASS" {
		t.Fatalf("expected PASS with normalization rules, got err=%v, verdict=%s", err, outWithNorm.CaseVerdict)
	}
}

// TestNamedCase_X05_MissingDBWrite verifies X05:
// Good response without required DB INSERT/UPDATE must mismatch despite status/body equality.
func TestNamedCase_X05_MissingDBWrite(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Target returns 201 Created and identical body, but omitted the DB write
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.State.DB.Tables = []EvaluationDBTable{
		{
			ConnectionID: "conn_main",
			Catalog:      "app_db",
			Table:        "users",
			Rows: TaggedValue{
				Type:  ValueTypeArray,
				Value: []TaggedValue{},
			},
		},
	}

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X05 violation: expected MISMATCH on missing DB write, got %s", out.CaseVerdict)
	}
	if *out.Report.Cases[0].Dimensions.Database.Verdict != "MISMATCH" {
		t.Fatalf("X05 violation: expected Database dimension MISMATCH")
	}
}

// TestNamedCase_X06_WrongPrincipal verifies X06:
// Write/respond for another account with plausible ID must mismatch across actor/DB/response/effect relations.
func TestNamedCase_X06_WrongPrincipal(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Target authenticated User Bob instead of mapped User Alice
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Actor.ActualPrincipal.SubjectLocalID = stringPtr("loc-tgt-user-bob")

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X06 violation: expected MISMATCH on wrong principal, got %s", out.CaseVerdict)
	}
	if *out.Report.Cases[0].Dimensions.Authentication.Verdict != "MISMATCH" {
		t.Fatalf("X06 violation: expected Authentication dimension MISMATCH")
	}
}

// TestNamedCase_X07_DuplicateEffect verifies X07:
// Worker accepts effect then crashes before ack; duplicate business application must fail effect cardinality/idempotency.
func TestNamedCase_X07_DuplicateEffect(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Target has 2 business effects emitted when only 1 is expected
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.EffectRefs = []string{
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X07 violation: expected MISMATCH on duplicate effect, got %s", out.CaseVerdict)
	}
	if *out.Report.Cases[0].Dimensions.External.Verdict != "MISMATCH" {
		t.Fatalf("X07 violation: expected External dimension MISMATCH")
	}
}

// TestNamedCase_X72_DeliveryReceiptRetryIdentity verifies X72:
// One message accepted before ack loss, retry same business key; distinct retry receipts alone are not duplicate business effects.
func TestNamedCase_X72_DeliveryReceiptRetryIdentity(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Both source and target have exactly 1 business effect applied (with dedup)
	in.SourceContexts["sha256:1111111111111111111111111111111111111111111111111111111111111111"].Namespaces.Result.EffectRefs = []string{
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.EffectRefs = []string{
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "PASS" {
		t.Fatalf("X72 violation: expected PASS for single business effect with dedup, got %s", out.CaseVerdict)
	}
}

// TestNamedCase_X77_PrincipalProvenance verifies X77:
// Session claims user A but DB or actual principal context shows user B -> MISMATCH.
func TestNamedCase_X77_PrincipalProvenance(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Target session is valid, but actual principal context measures User B
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Actor.ActualPrincipal.SubjectLocalID = stringPtr("loc-tgt-user-b")

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X77 violation: expected MISMATCH on principal provenance, got %s", out.CaseVerdict)
	}
}

// TestNamedCase_X80_SourceVersusSourceComparison verifies X80:
// Supplying two SOURCE witnesses or identical witnesses yields ERR_INVALID_COMPARISON / UNKNOWN.
func TestNamedCase_X80_SourceVersusSourceComparison(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Supply two SOURCE witnesses
	in.TargetWitness.Side = SideSource

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "UNKNOWN" || out.TopVerdict != "UNKNOWN" {
		t.Fatalf("X80 violation: expected UNKNOWN on source-vs-source comparison, got case=%s top=%s", out.CaseVerdict, out.TopVerdict)
	}
	if len(out.Errors) == 0 || out.Errors[0].Code != "ERR_INVALID_COMPARISON" {
		t.Fatalf("X80 violation: expected ERR_INVALID_COMPARISON, got %+v", out.Errors)
	}
}

// TestNamedCase_X61_CheckerWriteContamination verifies X61:
// Target omits INSERT while checker model contains it; comparator must NOT write to DB.
func TestNamedCase_X61_CheckerWriteContamination(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Target DB state has 0 rows
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.State.DB.Tables = []EvaluationDBTable{
		{
			ConnectionID: "conn_main",
			Catalog:      "app_db",
			Table:        "users",
			Rows: TaggedValue{
				Type:  ValueTypeArray,
				Value: []TaggedValue{},
			},
		},
	}

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X61 violation: expected MISMATCH, got %s", out.CaseVerdict)
	}

	// Verify target DB state was not contaminated by comparator
	targetRows := extractRows(in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.State.DB.Tables[0].Rows)
	if len(targetRows) != 0 {
		t.Fatalf("X61 violation: comparator wrote to target DB state!")
	}
}

// TestNamedCase_X62_SourceOnlyIdentityLedger verifies X62:
// Source ledger is valid without target or pairing; subsequent paired map leaves source ledger hashes unchanged.
func TestNamedCase_X62_SourceOnlyIdentityLedger(t *testing.T) {
	in := buildBaselinePassInput()

	srcBytesBefore, _ := toCanonicalJSON(in.SourceIdentityLedger)
	hashBefore := hexDigest(srcBytesBefore)

	comp := NewBehaviorComparator()
	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "PASS" {
		t.Fatalf("expected PASS, got %s", out.CaseVerdict)
	}

	srcBytesAfter, _ := toCanonicalJSON(in.SourceIdentityLedger)
	hashAfter := hexDigest(srcBytesAfter)

	if hashBefore != hashAfter {
		t.Fatalf("X62 violation: source identity ledger hash changed after pairing! %s != %s", hashBefore, hashAfter)
	}
}

// TestNamedCase_X63_ProjectionPoisoning verifies X63:
// Context origin MUST be ACTUAL_OBSERVED. Model expected state passed as actual yields MISMATCH.
func TestNamedCase_X63_ProjectionPoisoning(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Inject MODEL_EXPECTED into source context
	in.SourceContexts["sha256:1111111111111111111111111111111111111111111111111111111111111111"].Origin = OriginModelExpected

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X63 violation: expected MISMATCH when Origin is not ACTUAL_OBSERVED, got %s", out.CaseVerdict)
	}
}

// TestNamedCase_X66_StaleExecutedBuild verifies X66:
// Stale executed build hash yields ERR_STALE_EXECUTION / UNKNOWN.
func TestNamedCase_X66_StaleExecutedBuild(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	in.ExpectedTargetBuildHash = "sha256:" + strings.Repeat("f", 64)

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "UNKNOWN" || out.TopVerdict != "UNKNOWN" {
		t.Fatalf("X66 violation: expected UNKNOWN on stale build, got %s", out.CaseVerdict)
	}
	if len(out.Errors) == 0 || out.Errors[0].Code != "ERR_STALE_EXECUTION" {
		t.Fatalf("X66 violation: expected ERR_STALE_EXECUTION error")
	}
}

// TestNamedCase_X67_ResponseAndSessionLowering verifies X67:
// Missing Location header or dropped cookie attributes produce MISMATCH in HTTP and Session dimensions.
func TestNamedCase_X67_ResponseAndSessionLowering(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	// Target drops HttpOnly attribute
	in.TargetContexts["sha256:2222222222222222222222222222222222222222222222222222222222222222"].Namespaces.Result.Cookies[0].HTTPOnly = false

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "MISMATCH" {
		t.Fatalf("X67 violation: expected MISMATCH on dropped cookie attribute, got %s", out.CaseVerdict)
	}
	if *out.Report.Cases[0].Dimensions.Session.Verdict != "MISMATCH" {
		t.Fatalf("X67 violation: expected Session dimension MISMATCH")
	}
}

// TestNamedCase_X74_EmergencyInterruption verifies X74:
// Incomplete capture / interruption yields UNKNOWN, not PASS.
func TestNamedCase_X74_EmergencyInterruption(t *testing.T) {
	comp := NewBehaviorComparator()
	in := buildBaselinePassInput()

	in.IncompleteCaptureReason = "EMERGENCY_INTERRUPTED"

	out, err := comp.CompareCase(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CaseVerdict != "UNKNOWN" || out.TopVerdict != "UNKNOWN" {
		t.Fatalf("X74 violation: expected UNKNOWN on interrupted run, got %s", out.CaseVerdict)
	}
}
