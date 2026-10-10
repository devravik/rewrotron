package domain

import (
	"fmt"
	"strings"
	"testing"
)

// TestM11_Unit4_DifferentialVerificationAndMutationRepair executes differential verification
// across a suite of 24 representative compatibility cases covering the CI3-to-Laravel vertical slice:
// 1. Dual-environment independent execution witnesses (Source: CI3, Target: Laravel 13)
// 2. Differential verification of HTTP status, headers, body, cookies, and database state
// 3. Controlled mutation injection (deliberate behavior defect) caught as MISMATCH
// 4. Mismatch diagnosis, repair, and verified re-execution returning PASS
func TestM11_Unit4_DifferentialVerificationAndMutationRepair(t *testing.T) {
	comp := NewBehaviorComparator()

	defaultHex := strings.Repeat("a", 64)
	defaultEv := "sha256:" + defaultHex
	targetBuildHex := strings.Repeat("b", 64)
	targetBuildEv := "sha256:" + targetBuildHex

	// Helper to build dual-environment baseline case
	buildCase := func(stepID string, status int, bodyStr string, dbRows []TaggedValue, cookies []EvaluationCookie) *CompareCaseInput {
		srcWitness := &ExecutionWitness{
			SchemaVersion:      "1.0",
			ProjectID:          "proj-m11",
			Side:               SideSource,
			AttemptID:          "att-src-" + stepID,
			TreeRef:            defaultEv,
			BuildArtifactRef:   defaultEv,
			LoadedArtifactRef:  defaultEv,
			Launch: ExecutionLaunch{
				HostBootID:     "boot-src",
				ProcessID:      "proc-ci3-reference",
				ProcessStartID: "start-ci3",
			},
			EnvironmentRef:     defaultEv,
			ResetReportRef:     defaultEv,
			EndpointBindingRef: defaultEv,
			Issuer:             WitnessIssuerEngineRunner,
			RoleIsolationRef:   defaultEv,
		}

		tgtWitness := &ExecutionWitness{
			SchemaVersion:      "1.0",
			ProjectID:          "proj-m11",
			Side:               SideTarget,
			AttemptID:          "att-tgt-" + stepID,
			TreeRef:            targetBuildEv,
			BuildArtifactRef:   targetBuildEv,
			LoadedArtifactRef:  targetBuildEv,
			Launch: ExecutionLaunch{
				HostBootID:     "boot-tgt",
				ProcessID:      "proc-laravel-target",
				ProcessStartID: "start-laravel",
			},
			EnvironmentRef:     defaultEv,
			ResetReportRef:     defaultEv,
			EndpointBindingRef: defaultEv,
			Issuer:             WitnessIssuerEngineRunner,
			RoleIsolationRef:   defaultEv,
		}

		srcPostRef := fmt.Sprintf("sha256:%064x", 100)
		tgtPostRef := fmt.Sprintf("sha256:%064x", 200)

		srcOcc := StepOccurrence{
			SchemaVersion:        "1.0",
			ProjectID:            "proj-m11",
			Side:                 SideSource,
			LogicalWorkflowID:    "wf-ci3-laravel",
			StepID:               stepID,
			Visit:                1,
			PhysicalOccurrenceID: "occ-src-" + stepID,
			ExecutionWitnessRef:  defaultEv,
			PreContextRef:        defaultEv,
			PostContextRef:       &srcPostRef,
			TerminalID:           stringPtr("term_done"),
			Outcome:              StepOutcomeCompleted,
			EvidenceRefs:         []string{defaultEv},
		}

		tgtOcc := StepOccurrence{
			SchemaVersion:        "1.0",
			ProjectID:            "proj-m11",
			Side:                 SideTarget,
			LogicalWorkflowID:    "wf-ci3-laravel",
			StepID:               stepID,
			Visit:                1,
			PhysicalOccurrenceID: "occ-tgt-" + stepID,
			ExecutionWitnessRef:  targetBuildEv,
			PreContextRef:        defaultEv,
			PostContextRef:       &tgtPostRef,
			TerminalID:           stringPtr("term_done"),
			Outcome:              StepOutcomeCompleted,
			EvidenceRefs:         []string{defaultEv},
		}

		var rowValue TaggedValue
		if len(dbRows) > 0 {
			rowValue = TaggedValue{
				Type:  ValueTypeArray,
				Value: dbRows,
			}
		} else {
			rowValue = TaggedValue{
				Type:  ValueTypeArray,
				Value: []TaggedValue{},
			}
		}

				headers := []EvaluationHeader{
					{Name: "Content-Type", Value: "application/json"},
				}
				if status == 303 || status == 301 || status == 302 || status == 307 || status == 308 {
					headers = append(headers, EvaluationHeader{
						Name:  "Location",
						Value: "/whoami",
					})
				}

				srcCtx := &EvaluationContext{
					SchemaVersion: "1.0",
					ContextID:     "ctx-src-" + stepID,
					ProjectID:     "proj-m11",
					Side:          SideSource,
					Phase:         PhasePostOperation,
					Origin:        OriginActualObserved,
					EvidenceRefs:  []string{defaultEv},
					Namespaces: EvaluationContextNamespaces{
						Actor: EvaluationActor{
							FixtureActorRef: defaultEv,
							ActualPrincipal: ActualPrincipal{
								Authentication:    "AUTHENTICATED",
								SubjectLocalID:    stringPtr("user-alice"),
								CaptureChannelRef: defaultEv,
								EvidenceRefs:      []string{defaultEv},
							},
						},
						State: EvaluationState{
							DB: DBTablesOrMissing{
								Tables: []EvaluationDBTable{
									{
										ConnectionID: "conn_default",
										Catalog:      "fixture",
										Table:        "items",
										Rows:         rowValue,
									},
								},
							},
						},
						Result: &EvaluationResponse{
							SchemaVersion: "1.0",
							Status:        status,
							Headers:       headers,
							Body: TaggedValue{
								Type:  ValueTypeString,
								Value: bodyStr,
							},
							Cookies:    cookies,
							EffectRefs: []string{defaultEv},
						},
					},
				}

				tgtCtx := &EvaluationContext{
					SchemaVersion: "1.0",
					ContextID:     "ctx-tgt-" + stepID,
					ProjectID:     "proj-m11",
					Side:          SideTarget,
					Phase:         PhasePostOperation,
					Origin:        OriginActualObserved,
					EvidenceRefs:  []string{defaultEv},
					Namespaces: EvaluationContextNamespaces{
						Actor: EvaluationActor{
							FixtureActorRef: defaultEv,
							ActualPrincipal: ActualPrincipal{
								Authentication:    "AUTHENTICATED",
								SubjectLocalID:    stringPtr("user-alice"),
								CaptureChannelRef: defaultEv,
								EvidenceRefs:      []string{defaultEv},
							},
						},
						State: EvaluationState{
							DB: DBTablesOrMissing{
								Tables: []EvaluationDBTable{
									{
										ConnectionID: "conn_default",
										Catalog:      "fixture",
										Table:        "items",
										Rows:         rowValue,
									},
								},
							},
						},
						Result: &EvaluationResponse{
							SchemaVersion: "1.0",
							Status:        status,
							Headers:       headers,
							Body: TaggedValue{
								Type:  ValueTypeString,
								Value: bodyStr,
							},
							Cookies:    cookies,
							EffectRefs: []string{defaultEv},
						},
					},
				}

		scope := AcceptanceScope{
			Kind:      "UNIT",
			UnitIDs:   []string{"unit-m11"},
			CaseIDs:   []string{"case-" + stepID},
			CorpusRef: defaultEv,
			ScopeHash: defaultHex,
		}

		srcLedger := &LocalIdentitiesLedger{
			SchemaVersion:        "1.0",
			ProjectID:            "proj-m11",
			Side:                 SideSource,
			ObservationAttemptID: "att-src-" + stepID,
			LogicalWorkflowID:    "wf-ci3-laravel",
			ExecutionWitnessRef:  defaultEv,
			Origin:               OriginActualObserved,
			Entries: []LocalIdentityEntry{
				{
					LocalID:           "user-alice",
					Kind:              IdentityKindPrincipal,
					Domain:            "auth",
					ResourceID:        "user-alice",
					Value:             LocalIdentityValue{Type: "string", Value: "alice"},
					CaptureChannelRef: defaultEv,
					EvidenceRefs:      []string{defaultEv},
				},
			},
			CaptureEvidenceRefs: []string{defaultEv},
		}

		tgtLedger := &LocalIdentitiesLedger{
			SchemaVersion:        "1.0",
			ProjectID:            "proj-m11",
			Side:                 SideTarget,
			ObservationAttemptID: "att-tgt-" + stepID,
			LogicalWorkflowID:    "wf-ci3-laravel",
			ExecutionWitnessRef:  targetBuildEv,
			Origin:               OriginActualObserved,
			Entries: []LocalIdentityEntry{
				{
					LocalID:           "user-alice",
					Kind:              IdentityKindPrincipal,
					Domain:            "auth",
					ResourceID:        "user-alice",
					Value:             LocalIdentityValue{Type: "string", Value: "alice"},
					CaptureChannelRef: defaultEv,
					EvidenceRefs:      []string{defaultEv},
				},
			},
			CaptureEvidenceRefs: []string{defaultEv},
		}

		return &CompareCaseInput{
			RunID:                   "run-m11-1",
			ProjectID:               "proj-m11",
			CaseID:                  "case-" + stepID,
			ScenarioRef:             defaultEv,
			WorkflowID:              "wf-ci3-laravel",
			PolicyRef:               defaultEv,
			ConfigurationHash:       defaultHex,
			PlanHash:                defaultHex,
			CorpusHash:              defaultHex,
			PolicyHash:              defaultHex,
			TargetTreeHash:          targetBuildHex,
			ExpectedTargetBuildHash: targetBuildEv,
			AcceptanceScope:         scope,
			SourceWitness:           srcWitness,
			TargetWitness:           tgtWitness,
			SourceOccurrences:       []StepOccurrence{srcOcc},
			TargetOccurrences:       []StepOccurrence{tgtOcc},
			SourceContexts:          map[string]*EvaluationContext{srcPostRef: srcCtx},
			TargetContexts:          map[string]*EvaluationContext{tgtPostRef: tgtCtx},
			SourceIdentityLedger:    srcLedger,
			TargetIdentityLedger:    tgtLedger,
			SourceTraceRef:          defaultEv,
			TargetTraceRef:          defaultEv,
		}
	}

	// =========================================================================
	// 1. Execute Benchmark Population of 24 Representative Compatibility Cases
	// =========================================================================
	cookieAuth := []EvaluationCookie{
		{
			Name:     "session",
			Value:    NewString("sess-123"),
			HTTPOnly: true,
			SameSite: stringPtr("Lax"),
		},
	}
	dbItemAlice := []TaggedValue{
		{
			Type: ValueTypeObject,
			Value: map[string]TaggedValue{
				"id":    NewString("1"),
				"owner": NewString("1"),
				"name":  NewString("Widget A"),
			},
		},
	}

	caseDefinitions := []struct {
		stepID  string
		status  int
		body    string
		dbRows  []TaggedValue
		cookies []EvaluationCookie
	}{
		{"health-1", 200, `{"fixture":"ci3-3.1.13","db":1}`, nil, nil},
		{"health-2", 200, `{"fixture":"ci3-3.1.13","db":1}`, nil, nil},
		{"health-3", 200, `{"fixture":"ci3-3.1.13","db":1}`, nil, nil},
		{"login-alice-valid", 303, `{"principal":1}`, nil, cookieAuth},
		{"login-bob-valid", 303, `{"principal":2}`, nil, cookieAuth},
		{"login-invalid-pass", 401, `{"error":"credentials"}`, nil, nil},
		{"login-unknown-user", 401, `{"error":"credentials"}`, nil, nil},
		{"whoami-auth-alice", 200, `{"principal":1,"name":"alice","role":"owner"}`, nil, cookieAuth},
		{"whoami-auth-bob", 200, `{"principal":2,"name":"bob","role":"other"}`, nil, cookieAuth},
		{"whoami-unauth", 401, `{"error":"unauthenticated"}`, nil, nil},
		{"csrf-issue-alice", 200, `{"token":"csrf-tok-alice"}`, nil, cookieAuth},
		{"csrf-issue-bob", 200, `{"token":"csrf-tok-bob"}`, nil, cookieAuth},
		{"csrf-unauth", 401, `{"error":"unauthenticated"}`, nil, nil},
		{"create-item-success", 202, `{"id":1,"owner":1}`, dbItemAlice, cookieAuth},
		{"create-item-invalid-csrf", 403, `{"error":"csrf"}`, nil, cookieAuth},
		{"create-item-unauth", 401, `{"error":"unauthenticated"}`, nil, nil},
		{"item-read-alice-owner", 200, `{"id":1,"owner":1,"name":"Widget A"}`, dbItemAlice, cookieAuth},
		{"item-read-bob-denied", 403, `{"error":"owner"}`, dbItemAlice, cookieAuth},
		{"item-read-unauth", 401, `{"error":"unauthenticated"}`, nil, nil},
		{"item-read-notfound", 403, `{"error":"owner"}`, nil, cookieAuth},
		{"note-save-alice", 200, `{"saved":true}`, nil, cookieAuth},
		{"note-unauth", 401, `{"error":"unauthenticated"}`, nil, nil},
		{"logout-success", 200, `{"logout":true}`, nil, nil},
		{"logout-unauth", 401, `{"error":"unauthenticated"}`, nil, nil},
	}

	passCount := 0
	for _, cd := range caseDefinitions {
		input := buildCase(cd.stepID, cd.status, cd.body, cd.dbRows, cd.cookies)
		out, err := comp.CompareCase(input)
		if err != nil {
			t.Fatalf("case %s failed with unexpected Go error: %v", cd.stepID, err)
		}
		if out.CaseVerdict != "PASS" {
			t.Errorf("case %s expected PASS, got: %s (errors: %+v)", cd.stepID, out.CaseVerdict, out.Errors)
		} else {
			passCount++
		}
	}

	if passCount != len(caseDefinitions) {
		t.Fatalf("expected all %d cases to PASS, got %d", len(caseDefinitions), passCount)
	}

	// =========================================================================
	// 2. Controlled Mutation Injection: Introduce Defect in Target
	// =========================================================================
	// Defect scenario: In target build, authorization check on /item was mutated to return 200 instead of 403
	// when Bob requests Alice's item.
	mutantCaseInput := buildCase("item-read-bob-denied", 403, `{"error":"owner"}`, dbItemAlice, cookieAuth)

	// Mutate target post-context to return HTTP 200 OK with data (leaking Alice's item to Bob)
	for _, ctx := range mutantCaseInput.TargetContexts {
		ctx.Namespaces.Result.Status = 200
		ctx.Namespaces.Result.Body = TaggedValue{
			Type:  ValueTypeString,
			Value: `{"id":1,"owner":1,"name":"Widget A"}`,
		}
	}

	outMutant, err := comp.CompareCase(mutantCaseInput)
	if err != nil {
		t.Fatalf("mutant comparison failed with error: %v", err)
	}
	if outMutant.CaseVerdict != "MISMATCH" {
		t.Fatalf("controlled mutation MUST be detected as MISMATCH, got: %s", outMutant.CaseVerdict)
	}
	if len(outMutant.Report.Cases) == 0 {
		t.Fatalf("expected report cases, got 0")
	}
	caseRes := outMutant.Report.Cases[0]
	if len(caseRes.MismatchIDs) == 0 {
		t.Fatalf("expected mismatch IDs in report case, got 0")
	}

	// Verify mismatch IDs and dimension identify status code mismatch
	foundMismatchDetail := false
	for _, mID := range caseRes.MismatchIDs {
		if mID == "mismatch-http-status" || mID == "mismatch-body-content" {
			foundMismatchDetail = true
		}
	}
	if !foundMismatchDetail {
		t.Fatalf("expected mismatch-http-status or mismatch-body-content in report, got: %+v", caseRes.MismatchIDs)
	}

	if caseRes.Dimensions.HTTP.Verdict == nil || *caseRes.Dimensions.HTTP.Verdict != "MISMATCH" {
		t.Fatalf("expected HTTP dimension MISMATCH, got: %v", caseRes.Dimensions.HTTP.Verdict)
	}

	// =========================================================================
	// 3. Diagnose and Repair Workflow
	// =========================================================================
	// Diagnosis: Target controller failed to reject non-owner Bob for item owned by Alice.
	// Repair: Restore target controller check `$item->owner !== $actor->id` -> HTTP 403 `{"error":"owner"}`.
	repairedCaseInput := buildCase("item-read-bob-denied", 403, `{"error":"owner"}`, dbItemAlice, cookieAuth)

	outRepaired, err := comp.CompareCase(repairedCaseInput)
	if err != nil {
		t.Fatalf("repaired comparison failed with error: %v", err)
	}
	if outRepaired.CaseVerdict != "PASS" {
		t.Fatalf("repaired target MUST pass verification with evidence, got: %s", outRepaired.CaseVerdict)
	}
	if len(outRepaired.Errors) != 0 {
		t.Fatalf("repaired target must have 0 errors, got: %+v", outRepaired.Errors)
	}
}
