package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func x19Ref(path string) *SemanticsExpression {
	parts := strings.SplitN(path[1:], "/", 2)
	pointer := ""
	if len(parts) == 2 {
		pointer = "/" + parts[1]
	}
	valueType := ValueTypeString
	if strings.HasSuffix(path, "accepted") || strings.HasSuffix(path, "usable") {
		valueType = ValueTypeBoolean
	}
	return &SemanticsExpression{Op: "ref", Namespace: parts[0], Pointer: pointer, ValueType: valueType}
}

func x19Literal(value TaggedValue) *SemanticsExpression {
	return &SemanticsExpression{Op: "literal", ValueType: value.Type, Value: value.Value}
}

func x19Eq(path string, expected TaggedValue) *SemanticsExpression {
	if expected.Type == ValueTypeNull {
		ref := x19Ref(path)
		ref.ValueType = ValueTypeNull
		return &SemanticsExpression{Op: "is_null", Arg: ref}
	}
	return &SemanticsExpression{Op: "eq", Left: x19Ref(path), Right: x19Literal(expected)}
}

func x19Post(in *CompareCaseInput, side, step string) *EvaluationContext {
	occ, contexts := in.SourceOccurrences, in.SourceContexts
	if side == SideTarget {
		occ, contexts = in.TargetOccurrences, in.TargetContexts
	}
	for _, o := range occ {
		if o.StepID == step {
			return contexts[*o.PostContextRef]
		}
	}
	panic("missing test occurrence")
}

// These are synthetic independently supplied captures, not a runtime bridge claim.
// Expected obligations remain separate from the actual input under mutation.
func x19Control(t *testing.T) *CompareCaseInput {
	t.Helper()
	in := buildBaselinePassInput()
	baseSrc := x19Post(in, SideSource, "step_create")
	baseTgt := x19Post(in, SideTarget, "step_create")
	for _, ctx := range []*EvaluationContext{baseSrc, baseTgt} {
		ctx.ProjectionRef = "sha256:" + strings.Repeat("a", 64)
		ctx.Namespaces.Clock = NewNull()
		ctx.Namespaces.State.Sessions = NewNull()
		ctx.Namespaces.State.Cache = NewNull()
		ctx.Namespaces.State.Files = NewNull()
		ctx.Namespaces.State.Queues = NewNull()
		ctx.Namespaces.Result.Session = NewNull()
		ctx.Namespaces.Result.Principal = NewNull()
		ctx.Namespaces.Result.ApplicationError = NewNull()
	}
	in.SourceOccurrences, in.TargetOccurrences = nil, nil
	in.SourceContexts, in.TargetContexts = map[string]*EvaluationContext{}, map[string]*EvaluationContext{}
	ref := "sha256:" + strings.Repeat("c", 64)
	decision := "sha256:" + strings.Repeat("d", 64)
	q := &QualificationRecord{SchemaVersion: "1.0", QualificationID: "safe-codec", ProjectID: in.ProjectID,
		SubjectKind: "ADAPTER", ImplementationRef: ref, SchemaRefs: []string{ref}, CapabilityIDs: []string{"safe-session-codec"},
		InputProfileRef: ref, ControlSuiteRef: ref, ControlResultRefs: []string{ref}, Status: "SUPPORTED"}
	qref, err := q.Digest()
	if err != nil {
		t.Fatal(err)
	}
	in.AdmittedSessionQualifications = map[string]*QualificationRecord{qref: q}
	c := &SessionContinuityContract{SchemaVersion: "1.0", ProjectID: in.ProjectID, CaseID: in.CaseID, WorkflowID: in.WorkflowID, DecisionRef: decision, Strategy: "SAFE_BRIDGE", CodecImplementationRef: ref, CodecProfileRef: ref,
		CodecBinding: &LoweringBinding{CapabilityID: "safe-session-codec", ProviderManifestRef: ref, QualificationRef: qref,
			ObligationRefs: []string{ref}, InputShapeRef: ref, OutputShapeRef: ref}}
	in.SessionContinuity = c
	steps := []string{"login", "continuation", "expired-session", "tampered-session", "missing-session", "invalidated-session",
		"csrf-valid", "csrf-wrong", "csrf-replayed", "csrf-expired", "csrf-regenerate", "logout", "logout-source-replay",
		"logout-target-replay", "reset", "remember-me", "throttle", "concurrent-session", "other-actor", "flash-next", "flash-consumed", "rollback"}
	for i, step := range steps {
		for _, side := range []string{SideSource, SideTarget} {
			base, subject, prefix := baseSrc, "loc-user-alice", "src"
			if side == SideTarget {
				base, subject, prefix = baseTgt, "loc-tgt-user-alice", "tgt"
			}
			data, err := json.Marshal(base)
			if err != nil {
				t.Fatal(err)
			}
			var ctx EvaluationContext
			if err := json.Unmarshal(data, &ctx); err != nil {
				t.Fatal(err)
			}
			occID := fmt.Sprintf("occ-%s-%d", prefix, i)
			ctx.ContextID, ctx.OccurrenceID, ctx.Phase = "ctx-"+occID, &occID, PhasePostOperation
			ctx.Namespaces.Actor.ActualPrincipal.SessionLocalID = stringPtr(prefix + "-new")
			ctx.Namespaces.Actor.ActualPrincipal.Roles = []string{"user"}
			if step == "other-actor" {
				subject = "loc-user-bob"
				if side == SideTarget {
					subject = "loc-tgt-user-bob"
				}
				ctx.Namespaces.Actor.ActualPrincipal.SubjectLocalID = stringPtr(subject)
				ctx.Namespaces.Actor.FixtureActorRef = "sha256:" + strings.Repeat("e", 64)
			}
			ctx.Namespaces.State.Sessions = NewObject(map[string]TaggedValue{
				"current_local_id": NewString(prefix + "-new"), "csrf_token": NewString(prefix + "-csrf-new"),
				"csrf_session_local_id": NewString(prefix + "-new"), "principal_local_id": NewString(subject),
				"previous_usable": NewBoolean(false), "csrf_previous_usable": NewBoolean(false),
				"session_accepted": NewBoolean(true), "csrf_accepted": NewBoolean(true),
				"flash": NewString("welcome"), "expires_at": NewString("2030-01-01T00:30:00Z"),
			})
			session := ctx.Namespaces.State.Sessions.Value.(map[string]TaggedValue)
			if step == "concurrent-session" || step == "other-actor" {
				id := prefix + "-" + step
				session["current_local_id"] = NewString(id)
				session["csrf_session_local_id"] = NewString(id)
				ctx.Namespaces.Actor.ActualPrincipal.SessionLocalID = stringPtr(id)
			}
			ctx.Namespaces.Clock = NewString("2030-01-01T00:00:00Z")
			expires, err := NewTimestamp("2030-01-01T00:30:00Z")
			if err != nil {
				t.Fatal(err)
			}
			session["expires_at"] = expires
			clock, err := NewTimestamp("2030-01-01T00:00:00Z")
			if err != nil {
				t.Fatal(err)
			}
			if step == "expired-session" {
				clock, err = NewTimestamp("2030-01-01T00:30:00Z")
				if err != nil {
					t.Fatal(err)
				}
			}
			ctx.Namespaces.Clock = clock
			if strings.Contains(step, "session") && step != "concurrent-session" && step != "continuation" {
				session["session_accepted"] = NewBoolean(false)
			}
			if step == "logout" || strings.Contains(step, "logout-") || step == "reset" || step == "throttle" ||
				step == "expired-session" || step == "tampered-session" || step == "missing-session" || step == "invalidated-session" {
				session["session_accepted"] = NewBoolean(false)
				ctx.Namespaces.Actor.ActualPrincipal.Authentication = "ANONYMOUS"
				ctx.Namespaces.Actor.ActualPrincipal.SubjectLocalID = nil
				ctx.Namespaces.Actor.ActualPrincipal.Roles = []string{}
			}
			if step == "csrf-wrong" || step == "csrf-replayed" || step == "csrf-expired" {
				session["csrf_accepted"] = NewBoolean(false)
			}
			if step == "flash-consumed" {
				session["flash"] = NewNull()
			}
			ctxRef := "sha256:" + hexDigest([]byte(occID))
			preRef := "sha256:" + hexDigest([]byte("pre-"+occID))
			pre := ctx
			pre.Phase = PhasePreOperation
			pre.Namespaces.State.Sessions = NewObject(map[string]TaggedValue{"current_local_id": NewString(prefix + "-old"), "csrf_token": NewString(prefix + "-csrf-old")})
			pre.Namespaces.Result = nil
			o := StepOccurrence{SchemaVersion: "1.0", ProjectID: in.ProjectID, Side: side, LogicalWorkflowID: in.WorkflowID,
				StepID: step, Visit: 1, PhysicalOccurrenceID: occID, PreContextRef: preRef, PostContextRef: &ctxRef,
				ExecutionWitnessRef: ref, Outcome: StepOutcomeCompleted, EvidenceRefs: []string{ref}}
			if i+1 < len(steps) {
				o.SelectedNextStepID = &steps[i+1]
			} else {
				o.TerminalID = stringPtr("done")
			}
			if side == SideSource {
				in.SourceOccurrences = append(in.SourceOccurrences, o)
				in.SourceContexts[ctxRef] = &ctx
				in.SourceContexts[preRef] = &pre
			} else {
				in.TargetOccurrences = append(in.TargetOccurrences, o)
				in.TargetContexts[ctxRef] = &ctx
				in.TargetContexts[preRef] = &pre
			}
			obs := SessionObservation{Side: side, StepID: step, Visit: 1, Phase: PhasePostOperation, ActorRef: ctx.Namespaces.Actor.FixtureActorRef}
			add := func(id string, predicate *SemanticsExpression) {
				c.Checks = append(c.Checks, SessionLifecycleCheck{ID: prefix + "-" + step + "-" + id, Observation: obs, Predicate: predicate})
			}
			add("auth", x19Eq("/actor/actual_principal/authentication", NewString(ctx.Namespaces.Actor.ActualPrincipal.Authentication)))
			add("session-acceptance", x19Eq("/state/sessions/session_accepted", session["session_accepted"]))
			add("csrf-acceptance", x19Eq("/state/sessions/csrf_accepted", session["csrf_accepted"]))
			add("flash", x19Eq("/state/sessions/flash", session["flash"]))
			add("old-session-replay", x19Eq("/state/sessions/previous_usable", NewBoolean(false)))
			add("old-csrf-replay", x19Eq("/state/sessions/csrf_previous_usable", NewBoolean(false)))
			clockExpr := &SemanticsExpression{Op: "ref", Namespace: "clock", Pointer: "", ValueType: ValueTypeTimestamp}
			expiryExpr := x19Ref("/state/sessions/expires_at")
			expiryExpr.ValueType = ValueTypeTimestamp
			if step == "login" || step == "expired-session" {
				before := &SemanticsExpression{Op: "before", Left: clockExpr, Right: expiryExpr}
				if step == "expired-session" {
					before = &SemanticsExpression{Op: "not", Arg: before}
				}
				add("expiry-boundary", before)
			}
			add("csrf-binding", &SemanticsExpression{Op: "eq", Left: x19Ref("/state/sessions/csrf_session_local_id"), Right: x19Ref("/state/sessions/current_local_id")})
			if ctx.Namespaces.Actor.ActualPrincipal.Authentication == "AUTHENTICATED" {
				add("intended-principal", x19Eq("/actor/actual_principal/subject_local_id", NewString(subject)))
				add("principal-binding", &SemanticsExpression{Op: "eq", Left: x19Ref("/state/sessions/principal_local_id"), Right: x19Ref("/actor/actual_principal/subject_local_id")})
				add("session-binding", &SemanticsExpression{Op: "eq", Left: x19Ref("/state/sessions/current_local_id"), Right: x19Ref("/actor/actual_principal/session_local_id")})
			}
			if step == "login" || step == "csrf-regenerate" {
				before := obs
				before.Phase = PhasePreOperation
				path := "/state/sessions/current_local_id"
				if step == "csrf-regenerate" {
					path = "/state/sessions/csrf_token"
				}
				c.Relations = append(c.Relations, SessionLifecycleRelation{ID: prefix + "-" + step + "-regeneration", Different: true,
					Left: SessionLifecycleValue{Observation: before, Path: path}, Right: SessionLifecycleValue{Observation: obs, Path: path}})
			}
			if step == "continuation" || step == "concurrent-session" || step == "rollback" {
				prev := obs
				prev.StepID = "login"
				c.Relations = append(c.Relations, SessionLifecycleRelation{ID: prefix + "-" + step + "-continuity",
					Left:  SessionLifecycleValue{Observation: prev, Path: "/actor/actual_principal/subject_local_id"},
					Right: SessionLifecycleValue{Observation: obs, Path: "/actor/actual_principal/subject_local_id"}})
				if step == "concurrent-session" {
					c.Relations = append(c.Relations, SessionLifecycleRelation{ID: prefix + "-concurrent-distinct", Different: true,
						Left: SessionLifecycleValue{Observation: prev, Path: "/state/sessions/current_local_id"}, Right: SessionLifecycleValue{Observation: obs, Path: "/state/sessions/current_local_id"}})
				}
			}
		}
	}
	// Independent identity ledgers allow opaque values to differ but retain their relations.
	for _, side := range []string{SideSource, SideTarget} {
		ledger, prefix := in.SourceIdentityLedger, "src"
		if side == SideTarget {
			ledger, prefix = in.TargetIdentityLedger, "tgt"
		}
		local := "loc-user-bob"
		if side == SideTarget {
			local = "loc-tgt-user-bob"
		}
		ledger.Entries = append(ledger.Entries, LocalIdentityEntry{LocalID: local, Kind: IdentityKindPrincipal, Domain: "auth", ResourceID: "user-bob", ActorID: stringPtr("actor-bob"), Value: LocalIdentityValue{Type: "string", Value: "bob"}, CaptureChannelRef: ref, EvidenceRefs: []string{ref}})
		for _, name := range []string{"new", "csrf-new"} {
			ledger.Entries = append(ledger.Entries, LocalIdentityEntry{
				LocalID: prefix + "-" + name, Kind: IdentityKindSession, Domain: "session", ResourceID: name, ActorID: stringPtr("actor-default"),
				Value: LocalIdentityValue{Type: "string", Value: prefix + "-" + name}, CaptureChannelRef: ref, EvidenceRefs: []string{ref}})
		}
	}
	for _, field := range []string{"current_local_id", "csrf_token"} {
		path := "/state/sessions/" + field
		left := SessionObservation{Side: SideSource, StepID: "login", Visit: 1, Phase: PhasePostOperation, ActorRef: baseSrc.Namespaces.Actor.FixtureActorRef}
		right := left
		right.Side = SideTarget
		c.Relations = append(c.Relations, SessionLifecycleRelation{ID: field + "-correspondence", Left: SessionLifecycleValue{Observation: left, Path: path},
			Right: SessionLifecycleValue{Observation: right, Path: path}, IdentityRuleID: field, ActorID: "actor-default"})
		in.NormalizationRules = append(in.NormalizationRules, &NormalizationRule{SchemaVersion: "2.0", RuleID: field, Path: path,
			ValueType: "string", Variability: "TOKEN", Relation: "TOKEN_LIFECYCLE", RelationGroup: "relgroup_default", DecisionRef: decision,
			BusinessSignificanceRef: ref, NegativeMutationRefs: []string{ref}})
	}
	for _, step := range steps {
		left := SessionObservation{Side: SideSource, StepID: step, Visit: 1, Phase: PhasePostOperation, ActorRef: x19Post(in, SideSource, step).Namespaces.Actor.FixtureActorRef}
		right := left
		right.Side = SideTarget
		c.Relations = append(c.Relations, SessionLifecycleRelation{ID: step + "-permissions", Left: SessionLifecycleValue{Observation: left, Path: "/actor/actual_principal/roles"}, Right: SessionLifecycleValue{Observation: right, Path: "/actor/actual_principal/roles"}})
	}
	left := SessionObservation{Side: SideSource, StepID: "login", Visit: 1, Phase: PhasePostOperation, ActorRef: baseSrc.Namespaces.Actor.FixtureActorRef}
	right := left
	right.Side = SideTarget
	c.Relations = append(c.Relations, SessionLifecycleRelation{ID: "expiry-boundary", Left: SessionLifecycleValue{Observation: left, Path: "/state/sessions/expires_at"}, Right: SessionLifecycleValue{Observation: right, Path: "/state/sessions/expires_at"}})
	x19Pin(t, in)
	return in
}

func x19Pin(t *testing.T, in *CompareCaseInput) {
	t.Helper()
	ref, err := in.SessionContinuity.Digest()
	if err != nil {
		t.Fatal(err)
	}
	in.SessionContinuity.ContractRef, in.RequiredSessionContractRef = ref, ref
	for _, rule := range in.NormalizationRules {
		rule.LifecycleContractRef = ref
	}
	raw, err := json.Marshal(baseValidPlanSubjectMap())
	if err != nil {
		t.Fatal(err)
	}
	var plan PlanApprovalSubject
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatal(err)
	}
	plan.ProjectID = in.ProjectID
	plan.ConfigurationHash = in.ConfigurationHash
	plan.CorpusRef = "sha256:" + in.CorpusHash
	plan.PolicyRef = in.PolicyRef
	plan.DecisionRefs = []string{in.SessionContinuity.DecisionRef}
	plan.OrderedUnits[0].ContractRefs = []string{ref}
	plan.OrderedUnits[0].CaseIDs = []string{in.CaseID}
	hash, err := plan.Digest()
	if err != nil {
		t.Fatal(err)
	}
	in.PlanHash = hash
	in.SessionPlanSubject = &plan
	for _, side := range []string{SideSource, SideTarget} {
		w, occs, ledger := in.SourceWitness, in.SourceOccurrences, in.SourceIdentityLedger
		if side == SideTarget {
			w, occs, ledger = in.TargetWitness, in.TargetOccurrences, in.TargetIdentityLedger
		}
		ref := "sha256:" + strings.Repeat("a", 64)
		w.InstrumentationRefs, w.ControlRefs = []string{ref}, []string{ref}
		if side == SideSource {
			w.SourceSnapshotRef = stringPtr(ref)
		} else {
			w.PlanRef = stringPtr("sha256:" + hash)
			w.TargetModelRef = stringPtr(plan.TargetModelRef)
			w.GeneratedManifestRef = stringPtr(ref)
		}
		raw, err := w.CanonicalJSON()
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateExecutionWitness(raw); err != nil {
			t.Fatal(err)
		}
		wref := "sha256:" + hexDigest(raw)
		for i := range occs {
			occs[i].ExecutionWitnessRef = wref
		}
		if ledger != nil {
			ledger.ExecutionWitnessRef = wref
		}
	}
}

func x19Verdict(t *testing.T, in *CompareCaseInput, want string) {
	t.Helper()
	out, err := NewBehaviorComparator().CompareCase(in)
	if err != nil {
		t.Fatal(err)
	}
	if out.TopVerdict != want {
		t.Fatalf("want %s, got %s; session reason=%v; errors=%+v", want, out.TopVerdict, *out.Report.Cases[0].Dimensions.Session.Reason, out.Errors)
	}
}

func TestX19_PositiveContinuityAndReadOnly(t *testing.T) {
	in := x19Control(t)
	before, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	x19Verdict(t, in, "PASS")
	after, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("verifier mutated measured captures/contracts")
	}
}

func TestX19_LifecycleFalsification(t *testing.T) {
	tests := []struct {
		name, step, field string
		value             TaggedValue
	}{
		{"fixation", "login", "current_local_id", NewString("tgt-old")},
		{"old-session-replay", "login", "previous_usable", NewBoolean(true)},
		{"expired-session", "expired-session", "session_accepted", NewBoolean(true)},
		{"tampered-session", "tampered-session", "session_accepted", NewBoolean(true)},
		{"missing-session", "missing-session", "session_accepted", NewBoolean(true)},
		{"invalidated-session", "invalidated-session", "session_accepted", NewBoolean(true)},
		{"wrong-csrf", "csrf-wrong", "csrf_accepted", NewBoolean(true)},
		{"replayed-csrf", "csrf-replayed", "csrf_accepted", NewBoolean(true)},
		{"expired-csrf", "csrf-expired", "csrf_accepted", NewBoolean(true)},
		{"csrf-session-mismatch", "csrf-valid", "csrf_session_local_id", NewString("other-session")},
		{"token-regeneration", "csrf-regenerate", "csrf_token", NewString("tgt-csrf-old")},
		{"stale-token", "csrf-regenerate", "csrf_previous_usable", NewBoolean(true)},
		{"logout", "logout", "session_accepted", NewBoolean(true)},
		{"logout-source-replay", "logout-source-replay", "session_accepted", NewBoolean(true)},
		{"logout-target-replay", "logout-target-replay", "session_accepted", NewBoolean(true)},
		{"reset", "reset", "session_accepted", NewBoolean(true)},
		{"remember-me", "remember-me", "session_accepted", NewBoolean(false)},
		{"throttle", "throttle", "session_accepted", NewBoolean(true)},
		{"flash-early-loss", "flash-next", "flash", NewString("")},
		{"flash-too-long", "flash-consumed", "flash", NewString("welcome")},
		{"session-principal-binding", "continuation", "principal_local_id", NewString("loc-user-bob")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := x19Control(t)
			ctx := x19Post(in, SideTarget, tc.step)
			ctx.Namespaces.State.Sessions.Value.(map[string]TaggedValue)[tc.field] = tc.value
			want := "MISMATCH"
			if tc.name == "flash-too-long" {
				want = "UNKNOWN"
			} // Declared null versus string is an evaluator type gap.
			x19Verdict(t, in, want)
		})
	}
	for _, step := range []string{"continuation", "concurrent-session", "rollback"} {
		t.Run("wrong-principal-"+step, func(t *testing.T) {
			in := x19Control(t)
			x19Post(in, SideTarget, step).Namespaces.Actor.ActualPrincipal.SubjectLocalID = stringPtr("loc-user-bob")
			x19Verdict(t, in, "MISMATCH")
		})
	}
	t.Run("stale-authentication", func(t *testing.T) {
		in := x19Control(t)
		ap := &x19Post(in, SideTarget, "logout").Namespaces.Actor.ActualPrincipal
		ap.Authentication = "AUTHENTICATED"
		ap.SubjectLocalID = stringPtr("loc-tgt-user-alice")
		x19Verdict(t, in, "MISMATCH")
	})
}

func TestX19_UnknownEvidenceAndQualification(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CompareCaseInput)
	}{
		{"missing-contract", func(in *CompareCaseInput) { in.SessionContinuity = nil }},
		{"empty-obligations", func(in *CompareCaseInput) { in.SessionContinuity.Checks = nil; x19Pin(t, in) }},
		{"unapproved-edit", func(in *CompareCaseInput) { in.SessionContinuity.Checks[0].Predicate = x19Literal(NewBoolean(true)) }},
		{"outside-plan", func(in *CompareCaseInput) { in.SessionPlanSubject.DecisionRefs = nil }},
		{"missing-codec", func(in *CompareCaseInput) { in.AdmittedSessionQualifications = nil }},
		{"unsafe-serialization", func(in *CompareCaseInput) {
			for _, q := range in.AdmittedSessionQualifications {
				q.Status = "UNSUPPORTED"
				q.ReasonCodes = []string{"UNSAFE_DESERIALIZATION"}
			}
		}},
		{"unknown-encryption", func(in *CompareCaseInput) { in.SessionContinuity.CodecBinding = nil; x19Pin(t, in) }},
		{"swapped-codec", func(in *CompareCaseInput) {
			in.SessionContinuity.CodecImplementationRef = "sha256:" + strings.Repeat("f", 64)
			x19Pin(t, in)
		}},
		{"missing-context", func(in *CompareCaseInput) { delete(in.TargetContexts, *in.TargetOccurrences[0].PostContextRef) }},
		{"missing-projection-pin", func(in *CompareCaseInput) { x19Post(in, SideTarget, "login").ProjectionRef = "" }},
		{"missing-principal", func(in *CompareCaseInput) {
			x19Post(in, SideTarget, "login").Namespaces.Actor.ActualPrincipal.Missing = &MissingValue{Type: "missing", ReasonCode: "NOT_CAPTURED"}
		}},
		{"missing-session", func(in *CompareCaseInput) {
			x19Post(in, SideTarget, "login").Namespaces.State.Sessions = NewMissing("NOT_CAPTURED")
		}},
		{"missing-field", func(in *CompareCaseInput) {
			delete(x19Post(in, SideTarget, "login").Namespaces.State.Sessions.Value.(map[string]TaggedValue), "previous_usable")
		}},
		{"cross-type", func(in *CompareCaseInput) {
			x19Post(in, SideTarget, "csrf-regenerate").Namespaces.State.Sessions.Value.(map[string]TaggedValue)["csrf_token"] = NewBoolean(true)
		}},
		{"misattributed-context", func(in *CompareCaseInput) {
			x19Post(in, SideTarget, "login").OccurrenceID = stringPtr("other-occurrence")
		}},
		{"wrong-actor-capture", func(in *CompareCaseInput) {
			x19Post(in, SideTarget, "login").Namespaces.Actor.FixtureActorRef = "sha256:" + strings.Repeat("f", 64)
		}},
		{"source-only-comparison", func(in *CompareCaseInput) { in.TargetWitness.Side = SideSource }},
		{"missing-identity-correspondence", func(in *CompareCaseInput) {
			var entries []LocalIdentityEntry
			for _, entry := range in.TargetIdentityLedger.Entries {
				if entry.Kind == IdentityKindPrincipal {
					entries = append(entries, entry)
				}
			}
			in.TargetIdentityLedger.Entries = entries
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { in := x19Control(t); tc.mutate(in); x19Verdict(t, in, "UNKNOWN") })
	}
	t.Run("known-violation-dominates-missing", func(t *testing.T) {
		in := x19Control(t)
		in.AdmittedSessionQualifications = nil
		x19Post(in, SideTarget, "csrf-replayed").Namespaces.State.Sessions.Value.(map[string]TaggedValue)["csrf_accepted"] = NewBoolean(true)
		x19Verdict(t, in, "MISMATCH")
	})
	t.Run("fixture-substitution", func(t *testing.T) {
		in := x19Control(t)
		x19Post(in, SideTarget, "login").Origin = OriginFixturePrepared
		x19Verdict(t, in, "MISMATCH")
	})
}

func TestX19_SecurityCookieFalsification(t *testing.T) {
	for _, field := range []string{"domain", "path", "expires", "max-age", "secure", "http-only", "same-site", "extra", "duplicate-scope"} {
		t.Run(field, func(t *testing.T) {
			in := x19Control(t)
			cookie := &x19Post(in, SideTarget, "login").Namespaces.Result.Cookies[0]
			switch field {
			case "domain":
				cookie.Domain = stringPtr("evil.example")
			case "path":
				cookie.Path = stringPtr("/admin")
			case "expires":
				cookie.Expires = stringPtr("2030-01-02T00:00:00Z")
			case "max-age":
				v := 3600
				cookie.MaxAge = &v
			case "secure":
				cookie.Secure = false
			case "http-only":
				cookie.HTTPOnly = false
			case "same-site":
				cookie.SameSite = stringPtr("None")
			case "extra":
				x19Post(in, SideTarget, "login").Namespaces.Result.Cookies = append(x19Post(in, SideTarget, "login").Namespaces.Result.Cookies, *cookie)
			case "duplicate-scope":
				for _, side := range []string{SideSource, SideTarget} {
					resp := x19Post(in, side, "login").Namespaces.Result
					second := resp.Cookies[0]
					second.Path = stringPtr("/other")
					resp.Cookies = append(resp.Cookies, second)
				}
				x19Post(in, SideTarget, "login").Namespaces.Result.Cookies[1].Path = nil
			}
			x19Verdict(t, in, "MISMATCH")
		})
	}
}

func TestX19_NormalizationCannotHideSecurityRelations(t *testing.T) {
	for _, mutation := range []string{"expiry", "permissions", "cross-actor", "concurrent-collapse", "source-logout"} {
		t.Run(mutation, func(t *testing.T) {
			in := x19Control(t)
			in.NormalizationRules = append(in.NormalizationRules, &NormalizationRule{SchemaVersion: "2.0", RuleID: "clock", Path: "/state/sessions/expires_at", ValueType: "timestamp", Variability: "CLOCK", Relation: "CONTROLLED_CLOCK", RelationGroup: "clock", LifecycleContractRef: in.RequiredSessionContractRef, DecisionRef: in.SessionContinuity.DecisionRef, BusinessSignificanceRef: in.RequiredSessionContractRef, NegativeMutationRefs: []string{in.RequiredSessionContractRef}})
			switch mutation {
			case "expiry":
				v, err := NewTimestamp("2030-01-01T00:31:00Z")
				if err != nil {
					t.Fatal(err)
				}
				x19Post(in, SideTarget, "login").Namespaces.State.Sessions.Value.(map[string]TaggedValue)["expires_at"] = v
			case "permissions":
				x19Post(in, SideTarget, "continuation").Namespaces.Actor.ActualPrincipal.Roles = []string{"admin"}
			case "cross-actor":
				ctx := x19Post(in, SideTarget, "other-actor")
				ctx.Namespaces.Actor.ActualPrincipal.SubjectLocalID = stringPtr("loc-tgt-user-alice")
				ctx.Namespaces.State.Sessions.Value.(map[string]TaggedValue)["principal_local_id"] = NewString("loc-tgt-user-alice")
			case "concurrent-collapse":
				ctx := x19Post(in, SideTarget, "concurrent-session")
				ctx.Namespaces.Actor.ActualPrincipal.SessionLocalID = stringPtr("tgt-new")
				values := ctx.Namespaces.State.Sessions.Value.(map[string]TaggedValue)
				values["current_local_id"], values["csrf_session_local_id"] = NewString("tgt-new"), NewString("tgt-new")
			case "source-logout":
				x19Post(in, SideSource, "logout-source-replay").Namespaces.State.Sessions.Value.(map[string]TaggedValue)["session_accepted"] = NewBoolean(true)
			}
			x19Verdict(t, in, "MISMATCH")
		})
	}
}

func TestX19_QualificationAndMissingCannotBeSelfApproved(t *testing.T) {
	for _, status := range []string{"UNSUPPORTED", "INCONCLUSIVE"} {
		t.Run(status, func(t *testing.T) {
			in := x19Control(t)
			for _, q := range in.AdmittedSessionQualifications {
				q.Status = status
				q.ReasonCodes = []string{"UNSAFE_OR_UNKNOWN_CODEC"}
				ref, err := q.Digest()
				if err != nil {
					t.Fatal(err)
				}
				in.AdmittedSessionQualifications = map[string]*QualificationRecord{ref: q}
				in.SessionContinuity.CodecBinding.QualificationRef = ref
				break
			}
			x19Pin(t, in)
			x19Verdict(t, in, "UNKNOWN")
		})
	}
	t.Run("is-missing-is-not-observation", func(t *testing.T) {
		in := x19Control(t)
		in.SessionContinuity.Checks[0].Predicate = &SemanticsExpression{Op: "is_missing", Arg: x19Ref("/state/sessions/not_captured")}
		x19Pin(t, in)
		x19Verdict(t, in, "UNKNOWN")
	})
	t.Run("wrong-declared-type-is-not-missing", func(t *testing.T) {
		in := x19Control(t)
		ref := x19Ref("/state/sessions/current_local_id")
		ref.ValueType = ValueTypeBoolean
		in.SessionContinuity.Checks[0].Predicate = &SemanticsExpression{Op: "is_missing", Arg: ref}
		x19Pin(t, in)
		x19Verdict(t, in, "UNKNOWN")
	})
	t.Run("reviewed-relogin-is-explicit", func(t *testing.T) {
		in := x19Control(t)
		in.SessionContinuity.Strategy = "REVIEWED_RELOGIN"
		in.SessionContinuity.CodecBinding = nil
		in.SessionContinuity.CodecImplementationRef = ""
		in.SessionContinuity.CodecProfileRef = ""
		in.AdmittedSessionQualifications = nil
		x19Pin(t, in)
		out, err := NewBehaviorComparator().CompareCase(in)
		if err != nil {
			t.Fatal(err)
		}
		if out.TopVerdict != "PASS" || out.Report.Cases[0].Dimensions.Session.Reason == nil || !strings.Contains(*out.Report.Cases[0].Dimensions.Session.Reason, "active continuity excluded") {
			t.Fatal("explicit exception lost its scope limitation")
		}
	})
	t.Run("untrusted-issuer", func(t *testing.T) {
		in := x19Control(t)
		in.TargetWitness.Issuer = "AGENT"
		x19Verdict(t, in, "UNKNOWN")
	})
	t.Run("wrong-witness-ref", func(t *testing.T) {
		in := x19Control(t)
		in.TargetOccurrences[0].ExecutionWitnessRef = in.SourceOccurrences[0].ExecutionWitnessRef
		x19Verdict(t, in, "UNKNOWN")
	})
}

func TestX19_ContractWireShape(t *testing.T) {
	in := x19Control(t)
	raw, err := json.Marshal(in.SessionContinuity)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SessionContinuityContract
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	ref, err := decoded.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if ref != in.RequiredSessionContractRef {
		t.Fatal("contract wire round trip changed consumed bytes")
	}
	if err := ValidateSessionContinuityContract(raw); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExpression(mustX19ExpressionJSON(t, &SemanticsExpression{Op: "eq", Left: &SemanticsExpression{Op: "ref", Namespace: "actor", Pointer: "/actual_principal/subject_local_id", ValueType: ValueTypeNull}, Right: x19Literal(NewNull())})); err != nil {
		t.Fatal(err)
	}
}

func TestX19_ContractSchemaFixtures(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name     string
		Schema   string
		Instance json.RawMessage
		Valid    bool
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, fixture := range fixtures {
		if fixture.Schema != "https://rewrotron.dev/specs/verification/session-continuity.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			err := ValidateSessionContinuityContract(fixture.Instance)
			if (err == nil) != fixture.Valid {
				t.Fatalf("valid=%v; error=%v", fixture.Valid, err)
			}
		})
	}
	if count != 4 {
		t.Fatalf("expected four session contract fixtures, got %d", count)
	}
}

func mustX19ExpressionJSON(t *testing.T, expression *SemanticsExpression) []byte {
	t.Helper()
	raw, err := sessionExpressionJSON(expression)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
